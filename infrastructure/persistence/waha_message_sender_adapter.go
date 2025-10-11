package persistence

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/application/message"
	"go.uber.org/zap"
)

// WAHAMessageSenderAdapter adapta o WAHA client para implementar MessageSender
type WAHAMessageSenderAdapter struct {
	client *waha.WAHAClient
	logger *zap.Logger
}

// NewWAHAMessageSenderAdapter cria um novo adapter
func NewWAHAMessageSenderAdapter(client *waha.WAHAClient, logger *zap.Logger) *WAHAMessageSenderAdapter {
	return &WAHAMessageSenderAdapter{
		client: client,
		logger: logger,
	}
}

// SendMessage envia uma mensagem via WAHA
func (a *WAHAMessageSenderAdapter) SendMessage(ctx context.Context, msg *message.OutboundMessage) (*message.SendResult, error) {
	a.logger.Info("Sending message via WAHA",
		zap.String("channel_id", msg.ChannelID.String()),
		zap.String("contact_id", msg.ContactID.String()),
		zap.String("type", string(msg.Type)))

	// TODO: Buscar configuração do canal (sessionName) usando channelID
	// Por agora, retornar erro informativo
	return nil, fmt.Errorf("WAHA message sending not fully implemented yet - need channel configuration")
}

// SendBulkMessages envia múltiplas mensagens via WAHA
func (a *WAHAMessageSenderAdapter) SendBulkMessages(ctx context.Context, messages []*message.OutboundMessage) ([]*message.SendResult, error) {
	results := make([]*message.SendResult, 0, len(messages))
	for _, msg := range messages {
		result, err := a.SendMessage(ctx, msg)
		if err != nil {
			a.logger.Error("Failed to send bulk message", zap.Error(err))
		}
		results = append(results, result)
	}
	return results, nil
}

// GetSupportedTypes retorna os tipos de mensagem suportados pelo WAHA
func (a *WAHAMessageSenderAdapter) GetSupportedTypes() []message.MessageType {
	return []message.MessageType{
		message.MessageTypeText,
		message.MessageTypeImage,
		message.MessageTypeAudio,
		message.MessageTypeVideo,
		message.MessageTypeDocument,
		message.MessageTypeLocation,
		message.MessageTypeContact,
	}
}

// ValidateMessage valida uma mensagem antes do envio
func (a *WAHAMessageSenderAdapter) ValidateMessage(msg *message.OutboundMessage) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}

	if msg.ChannelID.String() == "" || msg.ChannelID.String() == "00000000-0000-0000-0000-000000000000" {
		return fmt.Errorf("channel_id is required")
	}
	if msg.ContactID.String() == "" || msg.ContactID.String() == "00000000-0000-0000-0000-000000000000" {
		return fmt.Errorf("contact_id is required")
	}

	// Validar tipo suportado
	supported := false
	for _, t := range a.GetSupportedTypes() {
		if msg.Type == t {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("message type %s not supported", msg.Type)
	}

	// Validar conteúdo para mensagens de texto
	if msg.Type == message.MessageTypeText && msg.Content == "" {
		return fmt.Errorf("content is required for text messages")
	}

	// Validar URL para mensagens de mídia
	if (msg.Type == message.MessageTypeImage ||
		msg.Type == message.MessageTypeAudio ||
		msg.Type == message.MessageTypeVideo ||
		msg.Type == message.MessageTypeDocument) &&
		(msg.MediaURL == nil || *msg.MediaURL == "") {
		return fmt.Errorf("media_url is required for media messages")
	}

	return nil
}
