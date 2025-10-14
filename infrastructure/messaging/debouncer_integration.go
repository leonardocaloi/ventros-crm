package messaging

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/ventros/crm/infrastructure/channels/waha"
)

// DebouncerIntegration integra debouncer com sistema existente
type DebouncerIntegration struct {
	debouncer *MessageDebouncerV2
	processor *MessageBatchProcessor
}

// NewDebouncerIntegration cria integração completa
func NewDebouncerIntegration(
	redisClient *redis.Client,
	processor *MessageBatchProcessor,
) *DebouncerIntegration {
	// Cria debouncer com processor como callback
	var processorFunc ProcessorFunc
	if processor != nil {
		processorFunc = func(ctx context.Context, sessionKey string, messages []BufferedMessage) error {
			return processor.Process(ctx, sessionKey, messages)
		}
	}

	debouncer := NewMessageDebouncerV2(redisClient, 0, processorFunc)

	return &DebouncerIntegration{
		debouncer: debouncer,
		processor: processor,
	}
}

// ProcessWAHAMessage processa mensagem WAHA com debouncing
func (d *DebouncerIntegration) ProcessWAHAMessage(
	ctx context.Context,
	wahaEvent waha.WAHAMessageEvent,
) error {
	// 1. Extrai dados da mensagem WAHA
	msg := d.wahaEventToBufferedMessage(wahaEvent)

	// 2. Monta session key
	sessionKey := BuildSessionKey(
		msg.ContactID,
		"waha", // inbox type
		wahaEvent.Session,
	)

	// 3. Push + Check automático
	return d.debouncer.PushAndCheck(ctx, sessionKey, msg)
}

// ProcessMessage versão genérica (não acoplada ao WAHA)
func (d *DebouncerIntegration) ProcessMessage(
	ctx context.Context,
	contactID string,
	channelType string,
	channelID string,
	messageID string,
	text string,
	messageType string,
	timestamp int64,
	fromContact bool,
	metadata map[string]interface{},
) error {
	sessionKey := BuildSessionKey(contactID, channelType, channelID)

	msg := BufferedMessage{
		MessageID:   messageID,
		Text:        text,
		Type:        messageType,
		Timestamp:   timestamp,
		FromContact: fromContact,
		ContactID:   contactID,
		SessionID:   channelID,
		Metadata:    metadata,
	}

	return d.debouncer.PushAndCheck(ctx, sessionKey, msg)
}

// wahaEventToBufferedMessage converte evento WAHA para BufferedMessage
func (d *DebouncerIntegration) wahaEventToBufferedMessage(event waha.WAHAMessageEvent) BufferedMessage {
	// Extrai texto baseado no tipo
	text := extractTextFromWAHA(event)

	// Extrai contact ID
	contactID := event.Payload.From
	if event.Payload.Participant != nil && *event.Payload.Participant != "" {
		contactID = *event.Payload.Participant
	}

	return BufferedMessage{
		MessageID:   event.Payload.ID,
		Text:        text,
		Type:        getWAHAMessageType(event),
		Timestamp:   event.Payload.Timestamp * 1000, // converte para millis
		FromContact: !event.Payload.FromMe,
		ContactID:   contactID,
		SessionID:   event.Session,
		Metadata: map[string]interface{}{
			"waha_event_id": event.ID,
			"has_media":     event.Payload.HasMedia,
			"source":        event.Payload.Source,
		},
	}
}

// Helper functions

func extractTextFromWAHA(event waha.WAHAMessageEvent) string {
	if event.Payload.Body != nil && *event.Payload.Body != "" {
		return *event.Payload.Body
	}

	// Se tem mídia mas sem texto, retorna indicador
	if event.Payload.HasMedia {
		return "[MEDIA]"
	}

	return ""
}

func getWAHAMessageType(event waha.WAHAMessageEvent) string {
	// Inferir tipo pela mídia
	if event.Payload.HasMedia && event.Payload.Media != nil {
		if event.Payload.Media.Mimetype != "" {
			if contains(event.Payload.Media.Mimetype, "image") {
				return "image"
			}
			if contains(event.Payload.Media.Mimetype, "video") {
				return "video"
			}
			if contains(event.Payload.Media.Mimetype, "audio") {
				return "audio"
			}
			return "document"
		}
	}

	return "text"
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && s[:len(substr)] == substr)
}

// GetDebouncer expõe debouncer para uso direto (se necessário)
func (d *DebouncerIntegration) GetDebouncer() *MessageDebouncerV2 {
	return d.debouncer
}

// GetProcessor expõe processor para uso direto (se necessário)
func (d *DebouncerIntegration) GetProcessor() *MessageBatchProcessor {
	return d.processor
}
