package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/application/message"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"go.uber.org/zap"
)

// WAHAMessageSenderAdapter adapta o WAHA client para implementar MessageSender
type WAHAMessageSenderAdapter struct {
	client           *waha.WAHAClient
	channelRepo      channel.Repository
	contactRepo      contact.Repository
	conversionPolicy waha.ConversionPolicy
	logger           *zap.Logger
}

// NewWAHAMessageSenderAdapter cria um novo adapter
func NewWAHAMessageSenderAdapter(client *waha.WAHAClient, channelRepo channel.Repository, contactRepo contact.Repository, logger *zap.Logger) *WAHAMessageSenderAdapter {
	return &WAHAMessageSenderAdapter{
		client:           client,
		channelRepo:      channelRepo,
		contactRepo:      contactRepo,
		conversionPolicy: waha.NewDefaultConversionPolicy(), // Política padrão: sempre converte
		logger:           logger,
	}
}

// WithConversionPolicy define uma política de conversão customizada
func (a *WAHAMessageSenderAdapter) WithConversionPolicy(policy waha.ConversionPolicy) *WAHAMessageSenderAdapter {
	a.conversionPolicy = policy
	return a
}

// SendMessage envia uma mensagem via WAHA
func (a *WAHAMessageSenderAdapter) SendMessage(ctx context.Context, msg *message.OutboundMessage) (*message.SendResult, error) {
	a.logger.Info("Sending message via WAHA",
		zap.String("channel_id", msg.ChannelID.String()),
		zap.String("contact_id", msg.ContactID.String()),
		zap.String("type", string(msg.Type)))

	startTime := time.Now()

	// Validar mensagem
	if err := a.ValidateMessage(msg); err != nil {
		return a.failResult(msg.ID, err), fmt.Errorf("message validation failed: %w", err)
	}

	// 1. Buscar canal para pegar sessionName (WAHA session ID)
	ch, err := a.channelRepo.GetByID(msg.ChannelID)
	if err != nil {
		a.logger.Error("Failed to find channel", zap.Error(err))
		return a.failResult(msg.ID, err), fmt.Errorf("failed to find channel: %w", err)
	}

	if ch.Type != channel.TypeWAHA && ch.Type != channel.TypeWhatsAppBusiness {
		err := fmt.Errorf("channel type %s is not supported for WAHA sending", ch.Type)
		return a.failResult(msg.ID, err), err
	}

	// Define política de conversão baseada no tipo de canal
	a.conversionPolicy = waha.NewChannelTypeConversionPolicy(string(ch.Type))
	a.logger.Debug("Using channel-type-based conversion policy",
		zap.String("channel_id", ch.ID.String()),
		zap.String("channel_type", string(ch.Type)))

	// Extract session name from WAHA config
	wahaConfig, err := ch.GetWAHAConfig()
	if err != nil {
		a.logger.Error("Failed to get WAHA config", zap.Error(err))
		return a.failResult(msg.ID, err), fmt.Errorf("failed to get WAHA config: %w", err)
	}
	sessionName := wahaConfig.SessionID

	// 2. Buscar contato para pegar waha_wid (WhatsApp ID)
	cont, err := a.contactRepo.FindByID(ctx, msg.ContactID)
	if err != nil {
		a.logger.Error("Failed to find contact", zap.Error(err))
		return a.failResult(msg.ID, err), fmt.Errorf("failed to find contact: %w", err)
	}

	// 3. Buscar custom fields do contato para pegar waha_wid
	customFields, err := a.contactRepo.GetCustomFields(ctx, msg.ContactID)
	if err != nil {
		a.logger.Error("Failed to get contact custom fields", zap.Error(err))
		return a.failResult(msg.ID, err), fmt.Errorf("failed to get contact custom fields: %w", err)
	}

	// Extract waha_wid from custom fields
	wahaWID, ok := customFields["waha_wid"]
	if !ok || wahaWID == "" {
		// Fallback: try to construct from phone if available
		if cont.Phone() != nil {
			// Remove non-digits and add @c.us suffix
			phoneStr := cont.Phone().String()
			wahaWID = phoneStr + "@c.us"
			a.logger.Warn("Contact missing waha_wid, using phone fallback",
				zap.String("contact_id", msg.ContactID.String()),
				zap.String("phone", phoneStr),
				zap.String("waha_wid", wahaWID))
		} else {
			err := fmt.Errorf("contact %s has no waha_wid and no phone for fallback", msg.ContactID)
			return a.failResult(msg.ID, err), err
		}
	}

	chatID := wahaWID

	// 4. Send based on message type
	var response *waha.SendMessageResponse
	switch msg.Type {
	case message.MessageTypeText:
		response, err = a.sendTextMessage(ctx, sessionName, chatID, msg)
	case message.MessageTypeImage:
		response, err = a.sendImageMessage(ctx, sessionName, chatID, msg)
	case message.MessageTypeVideo:
		response, err = a.sendVideoMessage(ctx, sessionName, chatID, msg)
	case message.MessageTypeAudio:
		response, err = a.sendVoiceMessage(ctx, sessionName, chatID, msg)
	case message.MessageTypeDocument:
		response, err = a.sendDocumentMessage(ctx, sessionName, chatID, msg)
	case message.MessageTypeLocation:
		response, err = a.sendLocationMessage(ctx, sessionName, chatID, msg)
	case message.MessageTypeContact:
		response, err = a.sendContactMessage(ctx, sessionName, chatID, msg)
	default:
		err := fmt.Errorf("unsupported message type: %s", msg.Type)
		return a.failResult(msg.ID, err), err
	}

	if err != nil {
		a.logger.Error("Failed to send message via WAHA", zap.Error(err))
		return a.failResult(msg.ID, err), err
	}

	// Build success result
	result := &message.SendResult{
		MessageID:   msg.ID,
		ExternalID:  strPtr(response.ID),
		Status:      "sent",
		DeliveredAt: timePtr(time.Now()),
		RetryCount:  0,
		Metadata: map[string]interface{}{
			"waha_response": response,
			"send_duration": time.Since(startTime).Milliseconds(),
			"session_name":  sessionName,
			"waha_wid":      wahaWID,
		},
	}

	a.logger.Info("Message sent successfully via WAHA",
		zap.String("message_id", msg.ID.String()),
		zap.String("external_id", response.ID),
		zap.Int64("duration_ms", time.Since(startTime).Milliseconds()))

	return result, nil
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

// sendTextMessage envia mensagem de texto
func (a *WAHAMessageSenderAdapter) sendTextMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	req := waha.SendTextRequest{
		ChatID: chatID,
		Text:   msg.Content,
	}
	return a.client.SendText(ctx, sessionName, req)
}

// sendImageMessage envia mensagem de imagem
func (a *WAHAMessageSenderAdapter) sendImageMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	if msg.MediaURL == nil {
		return nil, fmt.Errorf("media_url is required for image messages")
	}

	caption := &msg.Content
	if msg.Content == "" {
		caption = nil
	}

	// Extract mimetype from metadata or use default
	mimetype := "image/jpeg"
	if msg.Metadata != nil {
		if mt, ok := msg.Metadata["mimetype"].(string); ok && mt != "" {
			mimetype = mt
		}
	}

	req := waha.SendFileRequest{
		ChatID: chatID,
		File: waha.FilePayload{
			URL:      *msg.MediaURL,
			Mimetype: mimetype,
		},
		Caption: caption,
	}
	return a.client.SendImage(ctx, sessionName, req)
}

// sendVideoMessage envia mensagem de vídeo
func (a *WAHAMessageSenderAdapter) sendVideoMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	if msg.MediaURL == nil {
		return nil, fmt.Errorf("media_url is required for video messages")
	}

	caption := &msg.Content
	if msg.Content == "" {
		caption = nil
	}

	// Extract mimetype from metadata or use default
	mimetype := "video/mp4"
	if msg.Metadata != nil {
		if mt, ok := msg.Metadata["mimetype"].(string); ok && mt != "" {
			mimetype = mt
		}
	}

	// Usa política de conversão para decidir se deve converter
	var convertPtr *bool
	if a.conversionPolicy.ShouldConvertVideo(mimetype) {
		shouldConvert := true
		convertPtr = &shouldConvert
	}

	req := waha.SendVideoRequest{
		ChatID: chatID,
		File: waha.FilePayload{
			URL:      *msg.MediaURL,
			Mimetype: mimetype,
		},
		Caption: caption,
		Convert: convertPtr, // Conversão baseada na política injetada
	}
	return a.client.SendVideo(ctx, sessionName, req)
}

// sendVoiceMessage envia mensagem de voz/áudio
func (a *WAHAMessageSenderAdapter) sendVoiceMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	if msg.MediaURL == nil {
		return nil, fmt.Errorf("media_url is required for voice messages")
	}

	// Extract mimetype from metadata or use default
	mimetype := "audio/ogg; codecs=opus"
	if msg.Metadata != nil {
		if mt, ok := msg.Metadata["mimetype"].(string); ok && mt != "" {
			mimetype = mt
		}
	}

	// Usa política de conversão para decidir se deve converter
	var convertPtr *bool
	shouldConvertAudio := a.conversionPolicy.ShouldConvertAudio(mimetype)
	if shouldConvertAudio {
		shouldConvert := true
		convertPtr = &shouldConvert
		a.logger.Info("Audio will be converted by WAHA",
			zap.String("mimetype", mimetype),
			zap.String("url", *msg.MediaURL))
	} else {
		a.logger.Info("Audio will NOT be converted (native format)",
			zap.String("mimetype", mimetype),
			zap.String("url", *msg.MediaURL))
	}

	req := waha.SendVoiceRequest{
		ChatID: chatID,
		File: waha.FilePayload{
			URL:      *msg.MediaURL,
			Mimetype: mimetype,
		},
		Convert: convertPtr, // Conversão baseada na política injetada
	}
	return a.client.SendVoice(ctx, sessionName, req)
}

// sendDocumentMessage envia mensagem de documento
func (a *WAHAMessageSenderAdapter) sendDocumentMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	if msg.MediaURL == nil {
		return nil, fmt.Errorf("media_url is required for document messages")
	}

	filename := "document"
	mimetype := "application/pdf"

	if msg.Metadata != nil {
		if fn, ok := msg.Metadata["filename"].(string); ok && fn != "" {
			filename = fn
		}
		if mt, ok := msg.Metadata["mimetype"].(string); ok && mt != "" {
			mimetype = mt
		}
	}

	caption := &msg.Content
	if msg.Content == "" {
		caption = nil
	}

	req := waha.SendFileRequest{
		ChatID: chatID,
		File: waha.FilePayload{
			URL:      *msg.MediaURL,
			Mimetype: mimetype,
			Filename: filename,
		},
		Caption: caption,
	}
	return a.client.SendFile(ctx, sessionName, req)
}

// sendLocationMessage envia mensagem de localização
func (a *WAHAMessageSenderAdapter) sendLocationMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	if msg.Metadata == nil {
		return nil, fmt.Errorf("metadata is required for location messages")
	}

	lat, ok := msg.Metadata["latitude"].(float64)
	if !ok {
		return nil, fmt.Errorf("latitude is required in metadata for location messages")
	}

	lng, ok := msg.Metadata["longitude"].(float64)
	if !ok {
		return nil, fmt.Errorf("longitude is required in metadata for location messages")
	}

	var title *string
	if msg.Content != "" {
		title = &msg.Content
	}

	req := waha.SendLocationRequest{
		ChatID:    chatID,
		Latitude:  lat,
		Longitude: lng,
		Title:     title,
	}
	return a.client.SendLocation(ctx, sessionName, req)
}

// sendContactMessage envia mensagem de contato
func (a *WAHAMessageSenderAdapter) sendContactMessage(ctx context.Context, sessionName, chatID string, msg *message.OutboundMessage) (*waha.SendMessageResponse, error) {
	if msg.Metadata == nil {
		return nil, fmt.Errorf("metadata is required for contact messages")
	}

	vcard, ok := msg.Metadata["vcard"].(string)
	if !ok || vcard == "" {
		return nil, fmt.Errorf("vcard is required in metadata for contact messages")
	}

	req := waha.SendContactRequest{
		ChatID: chatID,
		Contacts: []waha.ContactPayload{
			{VCard: vcard},
		},
	}
	return a.client.SendContact(ctx, sessionName, req)
}

// failResult creates a failed send result
func (a *WAHAMessageSenderAdapter) failResult(messageID uuid.UUID, err error) *message.SendResult {
	errStr := err.Error()
	return &message.SendResult{
		MessageID:  messageID,
		Status:     "failed",
		Error:      &errStr,
		RetryCount: 0,
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}
