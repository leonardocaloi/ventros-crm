package message

import (
	"context"
	"errors"
	"time"

	domainMessage "github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/google/uuid"
)

// ConfirmMessageDeliveryCommand representa o comando para confirmar entrega de mensagem
type ConfirmMessageDeliveryCommand struct {
	MessageID     uuid.UUID  `json:"message_id" validate:"required"`
	ExternalID    string     `json:"external_id" validate:"required"`
	Status        string     `json:"status" validate:"required"` // delivered, read, failed
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	FailureReason *string    `json:"failure_reason,omitempty"`
}

// Validate valida o comando
func (cmd *ConfirmMessageDeliveryCommand) Validate() error {
	if cmd.MessageID == uuid.Nil {
		return errors.New("message_id is required")
	}
	if cmd.ExternalID == "" {
		return errors.New("external_id is required")
	}
	if cmd.Status == "" {
		return errors.New("status is required")
	}

	// Validar status válido
	validStatuses := []string{"delivered", "read", "failed"}
	valid := false
	for _, s := range validStatuses {
		if cmd.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("invalid status: must be delivered, read, or failed")
	}

	return nil
}

// MessageFinder interface para buscar e atualizar mensagens
type MessageFinder interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domainMessage.Message, error)
	FindByChannelMessageID(ctx context.Context, channelMessageID string) (*domainMessage.Message, error)
	Save(ctx context.Context, msg *domainMessage.Message) error
}

// ConfirmMessageDeliveryHandler processa confirmações de entrega
type ConfirmMessageDeliveryHandler struct {
	messageRepo MessageFinder
}

// NewConfirmMessageDeliveryHandler cria um novo handler
func NewConfirmMessageDeliveryHandler(messageRepo MessageFinder) *ConfirmMessageDeliveryHandler {
	return &ConfirmMessageDeliveryHandler{
		messageRepo: messageRepo,
	}
}

// Handle executa o comando
func (h *ConfirmMessageDeliveryHandler) Handle(ctx context.Context, cmd *ConfirmMessageDeliveryCommand) error {
	// 1. Validar comando
	if err := cmd.Validate(); err != nil {
		return err
	}

	// 2. Buscar mensagem (tentar por ID primeiro, depois por ExternalID)
	var msg *domainMessage.Message
	var err error

	if cmd.MessageID != uuid.Nil {
		msg, err = h.messageRepo.FindByID(ctx, cmd.MessageID)
	}

	// Se não encontrou por ID, tentar por ExternalID (ChannelMessageID)
	if (err != nil || msg == nil) && cmd.ExternalID != "" {
		msg, err = h.messageRepo.FindByChannelMessageID(ctx, cmd.ExternalID)
	}

	if err != nil || msg == nil {
		return errors.New("message not found")
	}

	// 3. Atualizar status baseado no comando
	switch cmd.Status {
	case "delivered":
		msg.MarkAsDelivered()
	case "read":
		msg.MarkAsRead()
	case "failed":
		msg.MarkAsFailed()
	default:
		return errors.New("invalid status")
	}

	// 4. Persistir mensagem atualizada
	if err := h.messageRepo.Save(ctx, msg); err != nil {
		return err
	}

	return nil
}
