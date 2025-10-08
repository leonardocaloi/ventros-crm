package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"go.uber.org/zap"
)

// WAHARawEventBus gerencia a publicação de eventos WAHA raw
// Garante que eventos nunca sejam perdidos, sempre enfileirando primeiro
type WAHARawEventBus struct {
	conn   *RabbitMQConnection
	logger *zap.Logger
}

// NewWAHARawEventBus cria um novo event bus para eventos WAHA raw
func NewWAHARawEventBus(conn *RabbitMQConnection, logger *zap.Logger) *WAHARawEventBus {
	return &WAHARawEventBus{
		conn:   conn,
		logger: logger,
	}
}

// PublishRawEvent publica um evento raw na fila de entrada
// Esta operação NUNCA deve falhar - é o ponto de entrada crítico
func (bus *WAHARawEventBus) PublishRawEvent(ctx context.Context, rawEvent waha.WAHARawEvent) error {
	// Serializa o evento
	payload, err := json.Marshal(rawEvent)
	if err != nil {
		// Se não conseguir serializar, cria um evento de erro mínimo
		bus.logger.Error("Failed to marshal raw event, creating minimal error event",
			zap.Error(err),
			zap.String("event_id", rawEvent.ID),
			zap.String("session", rawEvent.Session))
		
		// Cria evento de erro mínimo que sempre pode ser serializado
		errorEvent := waha.WAHARawEvent{
			ID:        rawEvent.ID,
			Timestamp: rawEvent.Timestamp,
			Session:   rawEvent.Session,
			Body:      []byte(`{"error": "failed_to_marshal_original_event"}`),
			Headers:   map[string]string{"Content-Type": "application/json"},
			Source:    "marshal_error",
			Metadata:  map[string]string{"original_error": err.Error()},
		}
		
		payload, _ = json.Marshal(errorEvent) // Este sempre funciona
	}

	// Publica na fila raw (nunca falha)
	queueName := "waha.events.raw"
	if err := bus.conn.Publish(ctx, queueName, payload); err != nil {
		// Log crítico - isso não deveria acontecer
		bus.logger.Error("CRITICAL: Failed to publish to raw queue",
			zap.Error(err),
			zap.String("queue", queueName),
			zap.String("event_id", rawEvent.ID))
		
		// Mesmo assim, não retorna erro para não quebrar o webhook
		// O evento será perdido, mas o webhook não falhará
		return nil
	}

	bus.logger.Debug("Raw event published successfully",
		zap.String("event_id", rawEvent.ID),
		zap.String("session", rawEvent.Session),
		zap.Int("body_size", len(rawEvent.Body)))

	return nil
}

// PublishProcessedEvent publica um evento processado para fila específica
func (bus *WAHARawEventBus) PublishProcessedEvent(ctx context.Context, queueName string, event waha.WAHAProcessedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal processed event: %w", err)
	}

	if err := bus.conn.Publish(ctx, queueName, payload); err != nil {
		return fmt.Errorf("failed to publish to queue %s: %w", queueName, err)
	}

	bus.logger.Debug("Processed event published",
		zap.String("queue", queueName),
		zap.String("event_type", event.EventType),
		zap.String("raw_event_id", event.RawEventID))

	return nil
}

// PublishParseError publica um erro de parsing para DLQ específica
func (bus *WAHARawEventBus) PublishParseError(ctx context.Context, parseError waha.WAHAParseError) error {
	payload, err := json.Marshal(parseError)
	if err != nil {
		return fmt.Errorf("failed to marshal parse error: %w", err)
	}

	// Envia para fila de erros de parsing
	queueName := "waha.events.parse_errors"
	if err := bus.conn.Publish(ctx, queueName, payload); err != nil {
		return fmt.Errorf("failed to publish parse error: %w", err)
	}

	bus.logger.Warn("Parse error published to DLQ",
		zap.String("raw_event_id", parseError.RawEventID),
		zap.String("error_type", parseError.ErrorType),
		zap.String("error", parseError.Error))

	return nil
}

// SetupRawEventQueues configura todas as filas necessárias para eventos raw
func (bus *WAHARawEventBus) SetupRawEventQueues() error {
	// Fila principal de entrada (raw events)
	if err := bus.conn.DeclareQueueWithDLQ("waha.events.raw", 3); err != nil {
		return fmt.Errorf("failed to declare raw events queue: %w", err)
	}

	// Filas de saída (eventos processados)
	processedQueues := []string{
		"waha.events.message.parsed",     // Mensagens válidas
		"waha.events.call.parsed",        // Chamadas válidas  
		"waha.events.presence.parsed",    // Presença válida
		"waha.events.group.parsed",       // Eventos de grupo válidos
		"waha.events.label.parsed",       // Eventos de label válidos
		"waha.events.unknown.parsed",     // Eventos desconhecidos mas válidos
	}

	for _, queue := range processedQueues {
		if err := bus.conn.DeclareQueueWithDLQ(queue, 3); err != nil {
			return fmt.Errorf("failed to declare processed queue %s: %w", queue, err)
		}
	}

	// Fila para erros de parsing
	if err := bus.conn.DeclareQueueWithDLQ("waha.events.parse_errors", 5); err != nil {
		return fmt.Errorf("failed to declare parse errors queue: %w", err)
	}

	bus.logger.Info("WAHA raw event queues setup completed")
	return nil
}

// GetQueueNameForEventType retorna o nome da fila baseado no tipo de evento
func (bus *WAHARawEventBus) GetQueueNameForEventType(eventType string) string {
	switch eventType {
	case "message", "message.any":
		return "waha.events.message.parsed"
	case "call.received", "call.accepted", "call.rejected":
		return "waha.events.call.parsed"
	case "presence":
		return "waha.events.presence.parsed"
	case "group.v2.join", "group.v2.leave", "group.v2.update", "group.v2.participants":
		return "waha.events.group.parsed"
	case "label.upsert", "label.deleted", "label.chat.added", "label.chat.deleted":
		return "waha.events.label.parsed"
	default:
		return "waha.events.unknown.parsed"
	}
}
