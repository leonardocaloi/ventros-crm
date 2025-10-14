package message

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/application/message"
	"github.com/ventros/crm/internal/domain/crm/contact"
	domainMessage "github.com/ventros/crm/internal/domain/crm/message"
	"github.com/ventros/crm/internal/domain/crm/session"
)

// SendMessageCommand representa o comando para enviar uma mensagem
type SendMessageCommand struct {
	ContactID   uuid.UUID                 `json:"contact_id" validate:"required"`
	ChannelID   uuid.UUID                 `json:"channel_id" validate:"required"`
	ContentType domainMessage.ContentType `json:"content_type" validate:"required"`
	Text        *string                   `json:"text,omitempty"`
	MediaURL    *string                   `json:"media_url,omitempty"`
	ReplyToID   *uuid.UUID                `json:"reply_to_id,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
	Priority    message.MessagePriority   `json:"priority,omitempty"`
	ScheduledAt *time.Time                `json:"scheduled_at,omitempty"`

	// Contexto de autenticação
	TenantID   string    `json:"-"` // Preenchido pelo middleware
	ProjectID  uuid.UUID `json:"-"` // Preenchido pelo middleware
	CustomerID uuid.UUID `json:"-"` // Preenchido pelo middleware
	AgentID    uuid.UUID `json:"-"` // OBRIGATÓRIO: Preenchido pelo middleware (agente autenticado)
}

// SendMessageResult representa o resultado do envio
type SendMessageResult struct {
	MessageID  uuid.UUID `json:"message_id"`
	ExternalID *string   `json:"external_id,omitempty"`
	Status     string    `json:"status"`
	SentAt     time.Time `json:"sent_at"`
	Error      *string   `json:"error,omitempty"`
}

// Validate valida o comando
func (cmd *SendMessageCommand) Validate() error {
	if cmd.ContactID == uuid.Nil {
		return errors.New("contact_id is required")
	}
	if cmd.ChannelID == uuid.Nil {
		return errors.New("channel_id is required")
	}
	if !cmd.ContentType.IsValid() {
		return errors.New("invalid content_type")
	}

	// Validar que texto existe para mensagens de texto
	if cmd.ContentType.IsText() && (cmd.Text == nil || *cmd.Text == "") {
		return errors.New("text is required for text messages")
	}

	// Validar que URL existe para mensagens de mídia
	if cmd.ContentType.IsMedia() && (cmd.MediaURL == nil || *cmd.MediaURL == "") {
		return errors.New("media_url is required for media messages")
	}

	if cmd.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if cmd.ProjectID == uuid.Nil {
		return errors.New("project_id is required")
	}
	if cmd.CustomerID == uuid.Nil {
		return errors.New("customer_id is required")
	}
	if cmd.AgentID == uuid.Nil {
		return errors.New("agent_id is required - must be authenticated agent")
	}

	return nil
}

// Repositories - usar as interfaces do domínio diretamente
type ContactRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error)
}

type SessionRepository interface {
	GetActiveSessionByContact(ctx context.Context, contactID uuid.UUID) (*session.Session, error)
	Save(ctx context.Context, sess *session.Session) error
}

type MessageRepository interface {
	Save(ctx context.Context, msg *domainMessage.Message) error
}

// TransactionManager gerencia transações de banco de dados.
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// SendMessageHandler processa o comando de envio de mensagem
type SendMessageHandler struct {
	contactRepo   ContactRepository
	sessionRepo   SessionRepository
	messageRepo   MessageRepository
	messageSender message.MessageSender
	txManager     TransactionManager
}

// NewSendMessageHandler cria um novo handler
func NewSendMessageHandler(
	contactRepo ContactRepository,
	sessionRepo SessionRepository,
	messageRepo MessageRepository,
	messageSender message.MessageSender,
	txManager TransactionManager,
) *SendMessageHandler {
	return &SendMessageHandler{
		contactRepo:   contactRepo,
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		messageSender: messageSender,
		txManager:     txManager,
	}
}

// Handle executa o comando
func (h *SendMessageHandler) Handle(ctx context.Context, cmd *SendMessageCommand) (*SendMessageResult, error) {
	// 1. Validar comando
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// 2. Verificar se o contato existe
	existingContact, err := h.contactRepo.FindByID(ctx, cmd.ContactID)
	if err != nil {
		return nil, errors.New("contact not found")
	}

	// 3-6. ✅ TRANSAÇÃO 1: Criar sessão (se necessário) + criar mensagem atomicamente
	var msg *domainMessage.Message
	var activeSession *session.Session

	err = h.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 3. Obter ou criar sessão ativa
		activeSession, err = h.sessionRepo.GetActiveSessionByContact(txCtx, cmd.ContactID)
		if err != nil || activeSession == nil {
			// Criar nova sessão se não existir
			// TODO: Get channelTypeID from channel repository
			var channelTypeID *int = nil
			timeoutDuration := 30 * time.Minute // Default timeout

			newSession, err := session.NewSession(
				cmd.ContactID,
				cmd.TenantID,
				channelTypeID,
				timeoutDuration,
			)
			if err != nil {
				return err
			}

			if err := h.sessionRepo.Save(txCtx, newSession); err != nil {
				return err
			}

			activeSession = newSession
		}

		// 4. Criar mensagem de domínio (fromMe = true para mensagens enviadas)
		msg, err = domainMessage.NewMessage(
			cmd.ContactID,
			cmd.ProjectID,
			cmd.CustomerID,
			cmd.ContentType,
			true, // fromMe = true
		)
		if err != nil {
			return err
		}

		// 5. Configurar conteúdo da mensagem
		msg.AssignToChannel(cmd.ChannelID, nil)
		msg.AssignToSession(activeSession.ID())

		if cmd.Text != nil && cmd.ContentType.IsText() {
			if err := msg.SetText(*cmd.Text); err != nil {
				return err
			}
		}

		if cmd.MediaURL != nil && cmd.ContentType.IsMedia() {
			// Media mimetype será determinado pelo adapter
			if err := msg.SetMediaContent(*cmd.MediaURL, ""); err != nil {
				return err
			}
		}

		// 6. Persistir mensagem (com status pending)
		if err := h.messageRepo.Save(txCtx, msg); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 7. Enviar mensagem via adapter (FORA da transação - pode demorar)
	outboundMsg := h.convertToOutboundMessage(cmd, msg, existingContact)
	sendResult, err := h.messageSender.SendMessage(ctx, outboundMsg)

	// 8-9. ✅ TRANSAÇÃO 2: Atualizar status da mensagem + contato atomicamente
	err2 := h.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if err != nil {
			// Marcar mensagem como falha
			msg.MarkAsFailed()
			if err := h.messageRepo.Save(txCtx, msg); err != nil {
				return err
			}
			return nil // Não propaga erro do SendMessage aqui
		}

		// Atualizar mensagem com ID externo
		if sendResult.ExternalID != nil {
			msg.SetChannelMessageID(*sendResult.ExternalID)
		}
		msg.MarkAsDelivered()
		if err := h.messageRepo.Save(txCtx, msg); err != nil {
			return err
		}

		// Registrar interação no contato
		existingContact.RecordInteraction()
		// Note: ContactRepository.Save não está sendo chamado porque Contact
		// não tem método Save exposto. RecordInteraction apenas atualiza timestamp interno.
		// TODO: Adicionar Save do contato se necessário.

		return nil
	})

	if err2 != nil {
		// Log erro mas retorna resultado da mensagem
		// Em produção: usar logging adequado
	}

	if err != nil {
		errMsg := err.Error()
		return &SendMessageResult{
			MessageID: msg.ID(),
			Status:    string(domainMessage.StatusFailed),
			SentAt:    time.Now(),
			Error:     &errMsg,
		}, err
	}

	return &SendMessageResult{
		MessageID:  msg.ID(),
		ExternalID: sendResult.ExternalID,
		Status:     sendResult.Status,
		SentAt:     time.Now(),
		Error:      nil,
	}, nil
}

// convertToOutboundMessage converte o comando em uma mensagem outbound
func (h *SendMessageHandler) convertToOutboundMessage(
	cmd *SendMessageCommand,
	msg *domainMessage.Message,
	contact *contact.Contact,
) *message.OutboundMessage {
	priority := cmd.Priority
	if priority == "" {
		priority = message.PriorityNormal
	}

	var msgType message.MessageType
	switch cmd.ContentType {
	case domainMessage.ContentTypeText:
		msgType = message.MessageTypeText
	case domainMessage.ContentTypeImage:
		msgType = message.MessageTypeImage
	case domainMessage.ContentTypeAudio:
		msgType = message.MessageTypeAudio
	case domainMessage.ContentTypeVideo:
		msgType = message.MessageTypeVideo
	case domainMessage.ContentTypeDocument:
		msgType = message.MessageTypeDocument
	case domainMessage.ContentTypeLocation:
		msgType = message.MessageTypeLocation
	case domainMessage.ContentTypeContact:
		msgType = message.MessageTypeContact
	default:
		msgType = message.MessageTypeText
	}

	content := ""
	if cmd.Text != nil {
		content = *cmd.Text
	}

	sessionID := msg.SessionID()

	return &message.OutboundMessage{
		ID:          msg.ID(),
		ChannelID:   cmd.ChannelID,
		ContactID:   cmd.ContactID,
		SessionID:   sessionID,
		AgentID:     cmd.AgentID,
		Source:      domainMessage.SourceManual, // Mensagem manual enviada por agente autenticado
		Type:        msgType,
		Content:     content,
		MediaURL:    cmd.MediaURL,
		Metadata:    cmd.Metadata,
		Priority:    priority,
		ScheduledAt: cmd.ScheduledAt,
		ReplyToID:   cmd.ReplyToID,
		CreatedAt:   time.Now(),
	}
}
