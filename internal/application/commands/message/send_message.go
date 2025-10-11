package message

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/internal/application/message"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	domainMessage "github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/google/uuid"
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
	TenantID   string     `json:"-"` // Preenchido pelo middleware
	ProjectID  uuid.UUID  `json:"-"` // Preenchido pelo middleware
	CustomerID uuid.UUID  `json:"-"` // Preenchido pelo middleware
	AgentID    *uuid.UUID `json:"-"` // Opcional, preenchido pelo middleware se for um agente
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

// SendMessageHandler processa o comando de envio de mensagem
type SendMessageHandler struct {
	contactRepo   ContactRepository
	sessionRepo   SessionRepository
	messageRepo   MessageRepository
	messageSender message.MessageSender
}

// NewSendMessageHandler cria um novo handler
func NewSendMessageHandler(
	contactRepo ContactRepository,
	sessionRepo SessionRepository,
	messageRepo MessageRepository,
	messageSender message.MessageSender,
) *SendMessageHandler {
	return &SendMessageHandler{
		contactRepo:   contactRepo,
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		messageSender: messageSender,
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

	// 3. Obter ou criar sessão ativa
	activeSession, err := h.sessionRepo.GetActiveSessionByContact(ctx, cmd.ContactID)
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
			return nil, err
		}

		if err := h.sessionRepo.Save(ctx, newSession); err != nil {
			return nil, err
		}

		activeSession = newSession
	}

	// 4. Criar mensagem de domínio (fromMe = true para mensagens enviadas)
	msg, err := domainMessage.NewMessage(
		cmd.ContactID,
		cmd.ProjectID,
		cmd.CustomerID,
		cmd.ContentType,
		true, // fromMe = true
	)
	if err != nil {
		return nil, err
	}

	// 5. Configurar conteúdo da mensagem
	msg.AssignToChannel(cmd.ChannelID, nil)
	msg.AssignToSession(activeSession.ID())

	if cmd.Text != nil && cmd.ContentType.IsText() {
		if err := msg.SetText(*cmd.Text); err != nil {
			return nil, err
		}
	}

	if cmd.MediaURL != nil && cmd.ContentType.IsMedia() {
		// Media mimetype será determinado pelo adapter
		if err := msg.SetMediaContent(*cmd.MediaURL, ""); err != nil {
			return nil, err
		}
	}

	// 6. Persistir mensagem
	if err := h.messageRepo.Save(ctx, msg); err != nil {
		return nil, err
	}

	// 7. Enviar mensagem via adapter
	outboundMsg := h.convertToOutboundMessage(cmd, msg, existingContact)
	sendResult, err := h.messageSender.SendMessage(ctx, outboundMsg)
	if err != nil {
		// Marcar mensagem como falha
		msg.MarkAsFailed()
		_ = h.messageRepo.Save(ctx, msg) // Best effort update

		errMsg := err.Error()
		return &SendMessageResult{
			MessageID: msg.ID(),
			Status:    string(domainMessage.StatusFailed),
			SentAt:    time.Now(),
			Error:     &errMsg,
		}, err
	}

	// 8. Atualizar mensagem com ID externo
	if sendResult.ExternalID != nil {
		msg.SetChannelMessageID(*sendResult.ExternalID)
	}
	msg.MarkAsDelivered()
	_ = h.messageRepo.Save(ctx, msg) // Best effort update

	// 9. Registrar interação no contato
	existingContact.RecordInteraction()

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
