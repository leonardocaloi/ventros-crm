package messaging

import (
	"context"
	"fmt"
	"time"

	messageports "github.com/caloi/ventros-crm/internal/application/message"
	"github.com/google/uuid"
)

// MessageSenderService implementa o serviço principal de envio de mensagens
// Seguindo Single Responsibility Principle (SRP) e Dependency Inversion Principle (DIP)
type MessageSenderService struct {
	senderFactory messageports.MessageSenderFactory
	messageRepo   messageports.MessageRepository
	messageQueue  messageports.MessageQueue
	validator     messageports.MessageValidator
	scheduler     messageports.MessageScheduler
	metrics       messageports.MessageMetrics
}

// NewMessageSenderService cria uma nova instância do serviço
func NewMessageSenderService(
	senderFactory messageports.MessageSenderFactory,
	messageRepo messageports.MessageRepository,
	messageQueue messageports.MessageQueue,
	validator messageports.MessageValidator,
	scheduler messageports.MessageScheduler,
	metrics messageports.MessageMetrics,
) *MessageSenderService {
	return &MessageSenderService{
		senderFactory: senderFactory,
		messageRepo:   messageRepo,
		messageQueue:  messageQueue,
		validator:     validator,
		scheduler:     scheduler,
		metrics:       metrics,
	}
}

// SendMessageRequest representa uma requisição de envio de mensagem
type SendMessageRequest struct {
	ChannelID    uuid.UUID                    `json:"channel_id" validate:"required"`
	ContactID    uuid.UUID                    `json:"contact_id" validate:"required"`
	SessionID    *uuid.UUID                   `json:"session_id,omitempty"`
	AgentID      *uuid.UUID                   `json:"agent_id,omitempty"`
	Type         messageports.MessageType     `json:"type" validate:"required"`
	Content      string                       `json:"content" validate:"required"`
	MediaURL     *string                      `json:"media_url,omitempty"`
	MediaType    *string                      `json:"media_type,omitempty"`
	Metadata     map[string]interface{}       `json:"metadata,omitempty"`
	Priority     messageports.MessagePriority `json:"priority"`
	ScheduledAt  *time.Time                   `json:"scheduled_at,omitempty"`
	ExpiresAt    *time.Time                   `json:"expires_at,omitempty"`
	ReplyToID    *uuid.UUID                   `json:"reply_to_id,omitempty"`
	TemplateData map[string]interface{}       `json:"template_data,omitempty"`
}

// SendMessageResponse representa a resposta do envio de mensagem
type SendMessageResponse struct {
	MessageID   uuid.UUID  `json:"message_id"`
	Status      string     `json:"status"`
	ExternalID  *string    `json:"external_id,omitempty"`
	QueuedAt    *time.Time `json:"queued_at,omitempty"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Error       *string    `json:"error,omitempty"`
}

// SendMessage envia uma mensagem através do canal apropriado
func (s *MessageSenderService) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	// Criar mensagem outbound
	message := &messageports.OutboundMessage{
		ID:           uuid.New(),
		ChannelID:    req.ChannelID,
		ContactID:    req.ContactID,
		SessionID:    req.SessionID,
		AgentID:      req.AgentID,
		Type:         req.Type,
		Content:      req.Content,
		MediaURL:     req.MediaURL,
		MediaType:    req.MediaType,
		Metadata:     req.Metadata,
		Priority:     req.Priority,
		ScheduledAt:  req.ScheduledAt,
		ExpiresAt:    req.ExpiresAt,
		ReplyToID:    req.ReplyToID,
		TemplateData: req.TemplateData,
		CreatedAt:    time.Now(),
	}

	// Validar mensagem
	if err := s.validator.ValidateContent(message.Type, message.Content); err != nil {
		return nil, fmt.Errorf("invalid message content: %w", err)
	}

	if message.MediaURL != nil && message.MediaType != nil {
		if err := s.validator.ValidateMedia(*message.MediaURL, *message.MediaType); err != nil {
			return nil, fmt.Errorf("invalid media: %w", err)
		}
	}

	if message.TemplateData != nil {
		if err := s.validator.ValidateTemplate(message.TemplateData); err != nil {
			return nil, fmt.Errorf("invalid template data: %w", err)
		}
	}

	// Salvar mensagem no repositório
	if err := s.messageRepo.SaveOutboundMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	response := &SendMessageResponse{
		MessageID: message.ID,
		Status:    "created",
	}

	// Se mensagem é agendada, usar scheduler
	if message.ScheduledAt != nil && message.ScheduledAt.After(time.Now()) {
		if err := s.scheduler.ScheduleMessage(ctx, message); err != nil {
			return nil, fmt.Errorf("failed to schedule message: %w", err)
		}
		response.Status = "scheduled"
		response.ScheduledAt = message.ScheduledAt
		return response, nil
	}

	// Enviar mensagem imediatamente ou enfileirar
	if message.Priority == messageports.PriorityUrgent {
		// Enviar imediatamente para mensagens urgentes
		result, err := s.sendMessageNow(ctx, message)
		if err != nil {
			// Se falhar, enfileirar para retry
			if queueErr := s.messageQueue.Enqueue(ctx, message); queueErr != nil {
				return nil, fmt.Errorf("failed to send message and enqueue: %v, queue error: %w", err, queueErr)
			}
			response.Status = "queued"
			now := time.Now()
			response.QueuedAt = &now
		} else {
			response.Status = result.Status
			response.ExternalID = result.ExternalID
		}
	} else {
		// Enfileirar mensagem para processamento assíncrono
		if err := s.messageQueue.Enqueue(ctx, message); err != nil {
			return nil, fmt.Errorf("failed to enqueue message: %w", err)
		}
		response.Status = "queued"
		now := time.Now()
		response.QueuedAt = &now
	}

	return response, nil
}

// SendBulkMessages envia múltiplas mensagens
func (s *MessageSenderService) SendBulkMessages(ctx context.Context, requests []*SendMessageRequest) ([]*SendMessageResponse, error) {
	responses := make([]*SendMessageResponse, len(requests))

	for i, req := range requests {
		response, err := s.SendMessage(ctx, req)
		if err != nil {
			responses[i] = &SendMessageResponse{
				MessageID: uuid.New(),
				Status:    "failed",
				Error:     stringPtr(err.Error()),
			}
		} else {
			responses[i] = response
		}
	}

	return responses, nil
}

// ProcessQueuedMessages processa mensagens na fila
func (s *MessageSenderService) ProcessQueuedMessages(ctx context.Context, batchSize int) error {
	messages, err := s.messageQueue.GetPendingMessages(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending messages: %w", err)
	}

	for _, message := range messages {
		if err := s.processMessage(ctx, message); err != nil {
			// Log error but continue processing other messages
			continue
		}
	}

	return nil
}

// sendMessageNow envia uma mensagem imediatamente
func (s *MessageSenderService) sendMessageNow(ctx context.Context, message *messageports.OutboundMessage) (*messageports.SendResult, error) {
	// Determinar o tipo de canal (isso deveria vir do channel repository)
	channelType := "waha" // TODO: Get from channel repository

	sender, err := s.senderFactory.CreateSender(channelType)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender for channel type %s: %w", channelType, err)
	}

	// Validar se o canal suporta a mensagem
	if !sender.IsChannelSupported(message.ChannelID) {
		return nil, fmt.Errorf("channel %s not supported by sender", message.ChannelID)
	}

	// Enviar mensagem
	result, err := sender.SendMessage(ctx, message)
	if err != nil {
		s.metrics.RecordMessageFailed(channelType, message.Type, err.Error())
		return nil, err
	}

	// Atualizar status da mensagem
	if err := s.messageRepo.UpdateMessageStatus(ctx, message.ID, result); err != nil {
		// Log error but don't fail the send operation
	}

	s.metrics.RecordMessageSent(channelType, message.Type)
	return result, nil
}

// processMessage processa uma mensagem individual da fila
func (s *MessageSenderService) processMessage(ctx context.Context, message *messageports.OutboundMessage) error {
	// Verificar se mensagem expirou
	if message.ExpiresAt != nil && message.ExpiresAt.Before(time.Now()) {
		return s.messageRepo.UpdateMessageStatus(ctx, message.ID, &messageports.SendResult{
			MessageID: message.ID,
			Status:    "expired",
			Error:     stringPtr("message expired"),
		})
	}

	// Enviar mensagem
	result, err := s.sendMessageNow(ctx, message)
	if err != nil {
		return s.messageRepo.UpdateMessageStatus(ctx, message.ID, &messageports.SendResult{
			MessageID: message.ID,
			Status:    "failed",
			Error:     stringPtr(err.Error()),
		})
	}

	return s.messageRepo.UpdateMessageStatus(ctx, message.ID, result)
}

// GetMessageStatus retorna o status de uma mensagem
func (s *MessageSenderService) GetMessageStatus(ctx context.Context, messageID uuid.UUID) (*messageports.OutboundMessage, error) {
	return s.messageRepo.GetMessageByID(ctx, messageID)
}

// GetSessionMessages retorna mensagens de uma sessão
func (s *MessageSenderService) GetSessionMessages(ctx context.Context, sessionID uuid.UUID) ([]*messageports.OutboundMessage, error) {
	return s.messageRepo.GetMessagesBySession(ctx, sessionID)
}

// CancelScheduledMessage cancela uma mensagem agendada
func (s *MessageSenderService) CancelScheduledMessage(ctx context.Context, messageID uuid.UUID) error {
	return s.scheduler.CancelScheduledMessage(ctx, messageID)
}

// GetQueueStats retorna estatísticas da fila
func (s *MessageSenderService) GetQueueStats(ctx context.Context) (map[string]interface{}, error) {
	queueSize, err := s.messageQueue.GetQueueSize(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"queue_size": queueSize,
		"timestamp":  time.Now(),
	}, nil
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
