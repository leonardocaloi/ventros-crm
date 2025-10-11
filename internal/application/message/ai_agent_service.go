package message

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/caloi/ventros-crm/internal/domain/crm/message_enrichment"
	"github.com/caloi/ventros-crm/internal/domain/crm/message_group"
)

// AIAgentService gerencia envio de mensagens concatenadas para AI Agent
type AIAgentService struct {
	logger                   *zap.Logger
	db                       *gorm.DB
	messageRepo              message.Repository
	messageGroupRepo         message_group.Repository
	messageEnrichmentService *MessageEnrichmentService
}

// NewAIAgentService cria um novo serviço de AI Agent
func NewAIAgentService(
	logger *zap.Logger,
	db *gorm.DB,
	messageRepo message.Repository,
	messageGroupRepo message_group.Repository,
	messageEnrichmentService *MessageEnrichmentService,
) *AIAgentService {
	return &AIAgentService{
		logger:                   logger,
		db:                       db,
		messageRepo:              messageRepo,
		messageGroupRepo:         messageGroupRepo,
		messageEnrichmentService: messageEnrichmentService,
	}
}

// ProcessCompletedGroup processa grupo concluído e envia para AI Agent
func (s *AIAgentService) ProcessCompletedGroup(
	ctx context.Context,
	group *message_group.MessageGroup,
) error {
	s.logger.Info("Processing completed group for AI Agent",
		zap.String("group_id", group.ID().String()),
		zap.Int("message_count", group.MessageCount()))

	// 1. Buscar todas as mensagens do grupo
	messages, err := s.getGroupMessages(ctx, group.MessageIDs())
	if err != nil {
		return fmt.Errorf("failed to get group messages: %w", err)
	}

	// 2. Buscar enriquecimentos do grupo
	enrichments, err := s.messageEnrichmentService.GetGroupEnrichments(ctx, group.ID())
	if err != nil {
		return fmt.Errorf("failed to get enrichments: %w", err)
	}

	// 3. Concatenar mensagens + enriquecimentos
	concatenatedContent := s.concatenateContent(messages, enrichments)

	// 4. Criar registro de histórico
	history := &entities.AIAgentHistoryEntity{
		ID:                  uuid.New(),
		GroupID:             group.ID(),
		SessionID:           group.SessionID(),
		ContactID:           group.ContactID(),
		ChannelID:           group.ChannelID(),
		TenantID:            group.TenantID(),
		ConcatenatedContent: concatenatedContent,
		MessageCount:        group.MessageCount(),
		EnrichmentCount:     len(enrichments),
		SentToAI:            false,
		CreatedAt:           time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to create AI agent history: %w", err)
	}

	s.logger.Info("Created AI agent history record",
		zap.String("history_id", history.ID.String()),
		zap.Int("message_count", history.MessageCount),
		zap.Int("enrichment_count", history.EnrichmentCount))

	// 5. Enviar para AI Agent (assíncrono)
	if err := s.sendToAIAgent(ctx, history); err != nil {
		s.logger.Error("Failed to send to AI agent",
			zap.Error(err),
			zap.String("history_id", history.ID.String()))
		return err
	}

	// 6. Marcar grupo como concluído
	if err := group.MarkAsCompleted(); err != nil {
		return fmt.Errorf("failed to mark group as completed: %w", err)
	}

	if err := s.messageGroupRepo.Save(ctx, group); err != nil {
		return fmt.Errorf("failed to save group: %w", err)
	}

	return nil
}

// getGroupMessages busca todas as mensagens do grupo
func (s *AIAgentService) getGroupMessages(
	ctx context.Context,
	messageIDs []uuid.UUID,
) ([]*message.Message, error) {
	messages := make([]*message.Message, 0, len(messageIDs))

	for _, id := range messageIDs {
		msg, err := s.messageRepo.FindByID(ctx, id)
		if err != nil {
			s.logger.Warn("Failed to get message",
				zap.Error(err),
				zap.String("message_id", id.String()))
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// concatenateContent concatena mensagens originais com enriquecimentos
func (s *AIAgentService) concatenateContent(
	messages []*message.Message,
	enrichments []*message_enrichment.MessageEnrichment,
) string {
	var builder strings.Builder

	// Criar mapa de enriquecimentos por message_id para lookup rápido
	enrichmentMap := make(map[uuid.UUID]*message_enrichment.MessageEnrichment)
	for _, e := range enrichments {
		enrichmentMap[e.MessageID()] = e
	}

	// Concatenar mensagens na ordem
	for i, msg := range messages {
		// Adicionar separador entre mensagens
		if i > 0 {
			builder.WriteString("\n\n---\n\n")
		}

		// Adicionar timestamp e identificação
		builder.WriteString(fmt.Sprintf("[Mensagem %d - %s]\n",
			i+1,
			msg.Timestamp().Format("15:04:05")))

		// Se tem texto, adicionar
		if msg.Text() != nil && *msg.Text() != "" {
			builder.WriteString(*msg.Text())
			builder.WriteString("\n")
		}

		// Se tem enriquecimento, adicionar
		if enrichment, exists := enrichmentMap[msg.ID()]; exists {
			if enrichment.ExtractedText() != nil && *enrichment.ExtractedText() != "" {
				builder.WriteString("\n[Conteúdo Enriquecido - ")
				builder.WriteString(string(enrichment.ContentType()))
				builder.WriteString("]\n")
				builder.WriteString(*enrichment.ExtractedText())
				builder.WriteString("\n")
			}
		}

		// Se tem mídia sem enriquecimento, mencionar
		if msg.MediaURL() != nil && enrichmentMap[msg.ID()] == nil {
			builder.WriteString(fmt.Sprintf("\n[Mídia: %s - %s]\n",
				msg.ContentType(),
				*msg.MediaURL()))
		}
	}

	return builder.String()
}

// sendToAIAgent envia conteúdo concatenado para AI Agent
func (s *AIAgentService) sendToAIAgent(
	ctx context.Context,
	history *entities.AIAgentHistoryEntity,
) error {
	startTime := time.Now()

	// TODO: Integrar com AI Agent real (OpenAI, Anthropic, etc)
	// Por enquanto, apenas simular
	s.logger.Info("Sending to AI Agent",
		zap.String("history_id", history.ID.String()),
		zap.Int("content_length", len(history.ConcatenatedContent)))

	// Simular resposta do AI
	time.Sleep(100 * time.Millisecond) // Simular latência
	aiResponse := "Entendi as mensagens e vou processar conforme solicitado."

	// Atualizar histórico
	now := time.Now()
	processingTimeMs := int(time.Since(startTime).Milliseconds())

	provider := "openai"
	model := "gpt-4"

	return s.db.WithContext(ctx).
		Model(&entities.AIAgentHistoryEntity{}).
		Where("id = ?", history.ID).
		Updates(map[string]interface{}{
			"sent_to_ai":           true,
			"ai_response":          aiResponse,
			"ai_provider":          provider,
			"ai_model":             model,
			"processing_time_ms":   processingTimeMs,
			"sent_at":              now,
			"response_received_at": now,
		}).Error
}

// GetAIAgentHistory retorna histórico de envio para AI Agent
func (s *AIAgentService) GetAIAgentHistory(
	ctx context.Context,
	sessionID uuid.UUID,
	limit int,
) ([]*entities.AIAgentHistoryEntity, error) {
	var history []*entities.AIAgentHistoryEntity
	err := s.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error

	if err != nil {
		return nil, err
	}

	return history, nil
}
