package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	messageapp "github.com/caloi/ventros-crm/internal/application/message"
	domainmessage "github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// WAHARawEventProcessor processa eventos raw da fila waha.events.raw
// Implementa a coreografia de alto nível com tratamento robusto de erros
type WAHARawEventProcessor struct {
	logger             *zap.Logger
	eventBus           *WAHARawEventBus
	wahaMessageService *messageapp.WAHAMessageService
	messageRepo        domainmessage.Repository
}

// NewWAHARawEventProcessor cria um novo processor de eventos raw
func NewWAHARawEventProcessor(
	logger *zap.Logger,
	eventBus *WAHARawEventBus,
	wahaMessageService *messageapp.WAHAMessageService,
	messageRepo domainmessage.Repository,
) *WAHARawEventProcessor {
	return &WAHARawEventProcessor{
		logger:             logger,
		eventBus:           eventBus,
		wahaMessageService: wahaMessageService,
		messageRepo:        messageRepo,
	}
}

// ProcessMessage implementa a interface Consumer para RabbitMQ
func (p *WAHARawEventProcessor) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	// Deserializa evento raw
	var rawEvent waha.WAHARawEvent
	if err := json.Unmarshal(delivery.Body, &rawEvent); err != nil {
		p.logger.Error("Failed to unmarshal raw event", zap.Error(err))
		return fmt.Errorf("invalid raw event JSON: %w", err)
	}

	p.logger.Info("Processing raw WAHA event",
		zap.String("event_id", rawEvent.ID),
		zap.String("session", rawEvent.Session),
		zap.Int("body_size", rawEvent.GetBodySize()),
		zap.String("source", rawEvent.Source))

	// Processa com recovery de panic
	return p.processRawEventWithRecovery(ctx, rawEvent)
}

// processRawEventWithRecovery processa evento com recovery de panic
func (p *WAHARawEventProcessor) processRawEventWithRecovery(ctx context.Context, rawEvent waha.WAHARawEvent) (err error) {
	// Recovery de panic
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("Panic in raw event processing",
				zap.Any("panic", r),
				zap.String("event_id", rawEvent.ID))

			// Cria erro de panic para DLQ
			parseError := waha.WAHAParseError{
				RawEventID: rawEvent.ID,
				Error:      fmt.Sprintf("panic: %v", r),
				ErrorType:  "panic",
				OccurredAt: time.Now(),
				RawBody:    rawEvent.Body,
			}

			// Envia para DLQ (não falha se der erro)
			if dlqErr := p.eventBus.PublishParseError(ctx, parseError); dlqErr != nil {
				p.logger.Error("Failed to publish panic error to DLQ", zap.Error(dlqErr))
			}

			err = fmt.Errorf("panic in processing: %v", r)
		}
	}()

	// Processa o evento
	return p.processRawEvent(ctx, rawEvent)
}

// processRawEvent processa um evento raw específico
func (p *WAHARawEventProcessor) processRawEvent(ctx context.Context, rawEvent waha.WAHARawEvent) error {
	// 1. Tenta fazer parse do evento WAHA
	wahaEvent, err := waha.ParseWebhookEvent(rawEvent.Body)
	if err != nil {
		return p.handleParseError(ctx, rawEvent, err, "webhook_parse")
	}

	// 2. Valida estrutura básica
	if wahaEvent.Event == "" {
		return p.handleParseError(ctx, rawEvent,
			fmt.Errorf("missing event type"), "missing_event_type")
	}

	if wahaEvent.Session == "" {
		// Usa session do raw event se não estiver no payload
		wahaEvent.Session = rawEvent.Session
	}

	// 3. Roteamento baseado no tipo de evento
	return p.routeEvent(ctx, rawEvent, wahaEvent)
}

// routeEvent roteia o evento para o processador apropriado
func (p *WAHARawEventProcessor) routeEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	switch wahaEvent.Event {
	case "message.any":
		// ✅ Processa TODAS as mensagens (fromMe: true/false)
		// Deduplicação já implementada para fromMe: true
		return p.processMessageEvent(ctx, rawEvent, wahaEvent)

	case "message":
		// ⚠️ Ignora "message" - usar apenas "message.any" para capturar tudo
		p.logger.Debug("Ignoring 'message' event, using 'message.any' instead",
			zap.String("raw_event_id", rawEvent.ID))
		return nil

	case "message.ack":
		return p.processMessageAckEvent(ctx, rawEvent, wahaEvent)

	case "call.received", "call.accepted", "call.rejected":
		return p.processCallEvent(ctx, rawEvent, wahaEvent)

	case "label.upsert", "label.deleted", "label.chat.added", "label.chat.deleted":
		return p.processLabelEvent(ctx, rawEvent, wahaEvent)

	case "group.v2.join", "group.v2.leave", "group.v2.update", "group.v2.participants":
		return p.processGroupEvent(ctx, rawEvent, wahaEvent)

	default:
		return p.processUnknownEvent(ctx, rawEvent, wahaEvent)
	}
}

// processMessageEvent processa eventos de mensagem
func (p *WAHARawEventProcessor) processMessageEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte para WAHAMessageEvent
	messageEvent, err := p.convertToMessageEvent(rawEvent, wahaEvent)
	if err != nil {
		return p.handleParseError(ctx, rawEvent, err, "message_conversion")
	}

	// 🎯 DEDUPLICAÇÃO: Se fromMe=true, verifica se já existe no banco
	// Cenário 1: Enviada pela API futura → já foi salva → descarta
	// Cenário 2: Enviada por outro device/WhatsApp Web → não existe → salva normalmente
	if messageEvent.Payload.FromMe {
		existingMsg, err := p.messageRepo.FindByChannelMessageID(ctx, messageEvent.Payload.ID)
		if err == nil && existingMsg != nil {
			// ✅ Mensagem já existe no banco (enviada pela API futura)
			p.logger.Debug("Message fromMe already exists (sent via API), skipping",
				zap.String("channel_message_id", messageEvent.Payload.ID),
				zap.String("message_id", existingMsg.ID().String()))
			return nil // Sucesso, mas não processa novamente
		}
		// ⚠️ Não existe no banco → foi enviada por outro device/WhatsApp Web
		// Continua processamento para salvar (será atribuída ao agente que enviou)
		p.logger.Info("Message fromMe not found in DB, processing as sent from another device",
			zap.String("channel_message_id", messageEvent.Payload.ID))
	}

	// Processa usando o service existente
	if err := p.wahaMessageService.ProcessWAHAMessage(ctx, messageEvent); err != nil {
		// Se falhar no processamento, envia para fila específica para retry
		return p.publishToProcessedQueue(ctx, rawEvent, wahaEvent, "waha.events.message.parsed")
	}

	p.logger.Info("Message event processed successfully",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("message_id", messageEvent.Payload.ID),
		zap.Bool("from_me", messageEvent.Payload.FromMe))

	return nil
}

// convertToMessageEvent converte evento WAHA para WAHAMessageEvent
func (p *WAHARawEventProcessor) convertToMessageEvent(rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) (waha.WAHAMessageEvent, error) {
	// Serializa e deserializa payload para converter interface{} para WAHAPayload
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		return waha.WAHAMessageEvent{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload waha.WAHAPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return waha.WAHAMessageEvent{}, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return waha.WAHAMessageEvent{
		ID:        rawEvent.ID, // Usa ID do raw event
		Timestamp: rawEvent.Timestamp.Unix(),
		Event:     wahaEvent.Event,
		Session:   wahaEvent.Session,
		Payload:   payload,
	}, nil
}

// processMessageAckEvent processa ACKs de mensagem (atualizações de status)
func (p *WAHARawEventProcessor) processMessageAckEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload para extrair informações do ACK
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal ack payload: %w", err)
	}

	var ackPayload struct {
		ID   string `json:"id"`   // ID da mensagem no WhatsApp
		Ack  int    `json:"ack"`  // Status: 1=sent, 2=delivered, 3=read, 4=played
		From string `json:"from"` // Remetente
	}

	if err := json.Unmarshal(payloadBytes, &ackPayload); err != nil {
		p.logger.Warn("Failed to unmarshal ACK payload, skipping",
			zap.Error(err),
			zap.String("raw_event_id", rawEvent.ID))
		return nil // Não falha, apenas ignora ACK malformado
	}

	// Mapeia ACK do WhatsApp para status do sistema
	var newStatus string
	switch ackPayload.Ack {
	case 1:
		newStatus = "sent"
	case 2:
		newStatus = "delivered"
	case 3:
		newStatus = "read"
	case 4:
		newStatus = "played" // Para áudios
	default:
		p.logger.Debug("Unknown ACK status, skipping",
			zap.Int("ack", ackPayload.Ack),
			zap.String("message_id", ackPayload.ID))
		return nil
	}

	// Busca mensagem pelo channel_message_id
	msg, err := p.messageRepo.FindByChannelMessageID(ctx, ackPayload.ID)
	if err != nil {
		// Mensagem não encontrada - pode ser que ainda não foi processada
		p.logger.Debug("Message not found for ACK, might be processed later",
			zap.String("channel_message_id", ackPayload.ID),
			zap.String("status", newStatus))
		return nil // Não falha, ACK é best-effort
	}

	// Atualiza status da mensagem usando métodos do domínio
	switch newStatus {
	case "delivered":
		msg.MarkAsDelivered()
	case "read", "played":
		msg.MarkAsRead()
	case "sent":
		// Já está como sent por padrão, não precisa fazer nada
	}

	// Salva mensagem atualizada
	if err := p.messageRepo.Save(ctx, msg); err != nil {
		p.logger.Warn("Failed to save message status from ACK",
			zap.Error(err),
			zap.String("channel_message_id", ackPayload.ID),
			zap.String("new_status", newStatus))
		return nil // Não falha, ACK é best-effort
	}

	p.logger.Info("Message status updated from ACK",
		zap.String("message_id", msg.ID().String()),
		zap.String("channel_message_id", ackPayload.ID),
		zap.String("new_status", newStatus),
		zap.Int("ack", ackPayload.Ack))

	return nil
}

// processCallEvent processa eventos de chamada
func (p *WAHARawEventProcessor) processCallEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// TODO: Implementar processamento de chamadas
	p.logger.Info("Call event received",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("event", wahaEvent.Event),
		zap.String("session", wahaEvent.Session))

	return p.publishToProcessedQueue(ctx, rawEvent, wahaEvent, "waha.events.call.parsed")
}

// processLabelEvent processa eventos de label/tag
func (p *WAHARawEventProcessor) processLabelEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// TODO: Implementar processamento de labels
	p.logger.Info("Label event received",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("event", wahaEvent.Event),
		zap.String("session", wahaEvent.Session))

	return p.publishToProcessedQueue(ctx, rawEvent, wahaEvent, "waha.events.label.parsed")
}

// processGroupEvent processa eventos de grupo
func (p *WAHARawEventProcessor) processGroupEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// TODO: Implementar processamento de grupos
	p.logger.Info("Group event received",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("event", wahaEvent.Event),
		zap.String("session", wahaEvent.Session))

	return p.publishToProcessedQueue(ctx, rawEvent, wahaEvent, "waha.events.group.parsed")
}

// processUnknownEvent processa eventos desconhecidos
func (p *WAHARawEventProcessor) processUnknownEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	p.logger.Warn("Unknown WAHA event type",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("event", wahaEvent.Event),
		zap.String("session", wahaEvent.Session))

	return p.publishToProcessedQueue(ctx, rawEvent, wahaEvent, "waha.events.unknown.parsed")
}

// publishToProcessedQueue publica evento para fila processada
func (p *WAHARawEventProcessor) publishToProcessedQueue(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent, queueName string) error {
	processedEvent := waha.WAHAProcessedEvent{
		RawEventID: rawEvent.ID,
		EventType:  wahaEvent.Event,
		Session:    wahaEvent.Session,
		ParsedAt:   time.Now(),
		Payload:    wahaEvent.Payload.(map[string]interface{}),
		Metadata: map[string]interface{}{
			"original_timestamp": rawEvent.Timestamp,
			"source":             rawEvent.Source,
			"body_size":          rawEvent.GetBodySize(),
		},
	}

	return p.eventBus.PublishProcessedEvent(ctx, queueName, processedEvent)
}

// handleParseError trata erros de parsing
func (p *WAHARawEventProcessor) handleParseError(ctx context.Context, rawEvent waha.WAHARawEvent, err error, errorType string) error {
	p.logger.Error("Parse error in raw event",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("error_type", errorType),
		zap.Error(err))

	// Cria erro estruturado
	parseError := waha.WAHAParseError{
		RawEventID: rawEvent.ID,
		Error:      err.Error(),
		ErrorType:  errorType,
		OccurredAt: time.Now(),
		RawBody:    rawEvent.Body,
	}

	// Envia para DLQ
	if dlqErr := p.eventBus.PublishParseError(ctx, parseError); dlqErr != nil {
		p.logger.Error("Failed to publish parse error to DLQ", zap.Error(dlqErr))
	}

	// Retorna erro original para que seja tratado pelo RabbitMQ
	return fmt.Errorf("parse error (%s): %w", errorType, err)
}

// Start inicia o consumer de eventos raw
func (p *WAHARawEventProcessor) Start(ctx context.Context, rabbitConn *RabbitMQConnection) error {
	queueName := "waha.events.raw"
	consumerTag := fmt.Sprintf("ventros-crm-raw-processor-%s", uuid.New().String()[:8])

	p.logger.Info("Starting WAHA raw event processor",
		zap.String("queue", queueName),
		zap.String("consumer_tag", consumerTag))

	return rabbitConn.StartConsumer(ctx, queueName, consumerTag, p, 5) // Processa 5 eventos simultaneamente
}
