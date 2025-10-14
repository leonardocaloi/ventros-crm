package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ventros/crm/infrastructure/channels/waha"
	messageapp "github.com/ventros/crm/internal/application/message"
	domainchannel "github.com/ventros/crm/internal/domain/crm/channel"
	domainchat "github.com/ventros/crm/internal/domain/crm/chat"
	domaincontact "github.com/ventros/crm/internal/domain/crm/contact"
	domainmessage "github.com/ventros/crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

// WAHARawEventProcessor processa eventos raw da fila waha.events.raw
// Implementa a coreografia de alto nível com tratamento robusto de erros
type WAHARawEventProcessor struct {
	logger             *zap.Logger
	eventBus           *WAHARawEventBus
	wahaMessageService *messageapp.WAHAMessageService
	messageRepo        domainmessage.Repository
	channelRepo        domainchannel.Repository
	contactRepo        domaincontact.Repository
	chatRepo           domainchat.Repository
}

// NewWAHARawEventProcessor cria um novo processor de eventos raw
func NewWAHARawEventProcessor(
	logger *zap.Logger,
	eventBus *WAHARawEventBus,
	wahaMessageService *messageapp.WAHAMessageService,
	messageRepo domainmessage.Repository,
	channelRepo domainchannel.Repository,
	contactRepo domaincontact.Repository,
	chatRepo domainchat.Repository,
) *WAHARawEventProcessor {
	return &WAHARawEventProcessor{
		logger:             logger,
		eventBus:           eventBus,
		wahaMessageService: wahaMessageService,
		messageRepo:        messageRepo,
		channelRepo:        channelRepo,
		contactRepo:        contactRepo,
		chatRepo:           chatRepo,
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

	case "session.status":
		return p.processSessionStatusEvent(ctx, rawEvent, wahaEvent)

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

	// ✅ EXTRAÇÃO DE IDENTIFICADORES DO WHATSAPP
	// Após processar a mensagem com sucesso, extrai e salva os identificadores normalizados
	p.extractAndSaveWhatsAppIdentifiers(ctx, rawEvent, messageEvent)

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

// processMessageAckEvent processa ACKs de mensagem (atualizações de status).
//
// Fluxo de camadas:
// 1. Infrastructure (WAHA): Recebe valores -1 a 4 do WhatsApp
// 2. Application (este processor): Adapta usando waha.WhatsAppAck
// 3. Domain (message.Status): Usa valores agnósticos (queued, sent, delivered, read, failed)
//
// O WhatsApp envia eventos message.ack quando o status de entrega/leitura muda:
// - ACK 1 (SERVER)  → Mensagem enviada ao servidor WhatsApp
// - ACK 2 (DEVICE)  → Mensagem entregue ao dispositivo (✓✓)
// - ACK 3 (READ)    → Mensagem lida pelo destinatário (✓✓ azul)
// - ACK 4 (PLAYED)  → Mensagem de voz/áudio reproduzida (SOMENTE voice messages)
func (p *WAHARawEventProcessor) processMessageAckEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload para extrair informações do ACK
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal ack payload: %w", err)
	}

	var ackPayload struct {
		ID      string `json:"id"`      // ID da mensagem no WhatsApp
		Ack     int    `json:"ack"`     // Status: -1 a 4 (ver WhatsAppAck)
		AckName string `json:"ackName"` // Nome do ACK: "ERROR", "PENDING", "SERVER", "DEVICE", "READ", "PLAYED"
		From    string `json:"from"`    // Remetente
	}

	if err := json.Unmarshal(payloadBytes, &ackPayload); err != nil {
		p.logger.Warn("Failed to unmarshal ACK payload, skipping",
			zap.Error(err),
			zap.String("raw_event_id", rawEvent.ID))
		return nil // Não falha, apenas ignora ACK malformado
	}

	// Cria WhatsAppAck value object (adapter infrastructure → domain)
	whatsappAck, err := waha.NewWhatsAppAck(ackPayload.Ack)
	if err != nil {
		p.logger.Warn("Invalid WhatsApp ACK value, skipping",
			zap.Error(err),
			zap.Int("ack", ackPayload.Ack),
			zap.String("ack_name", ackPayload.AckName),
			zap.String("message_id", ackPayload.ID))
		return nil // Não falha, apenas ignora ACK inválido
	}

	// ACKs de erro e pending não atualizam mensagens existentes
	if !whatsappAck.ShouldUpdateStatus() {
		p.logger.Debug("ACK should not update status, skipping",
			zap.String("ack", whatsappAck.String()),
			zap.Int("ack_value", whatsappAck.Value()),
			zap.String("message_id", ackPayload.ID))
		return nil
	}

	// Converte ACK para Status do domínio
	newStatus, err := whatsappAck.ToStatus()
	if err != nil {
		p.logger.Error("Failed to convert ACK to status",
			zap.Error(err),
			zap.String("ack", whatsappAck.String()),
			zap.String("message_id", ackPayload.ID))
		return nil
	}

	// Busca mensagem pelo channel_message_id
	msg, err := p.messageRepo.FindByChannelMessageID(ctx, ackPayload.ID)
	if err != nil {
		// Mensagem não encontrada - pode ser que ainda não foi processada
		p.logger.Debug("Message not found for ACK, might be processed later",
			zap.String("channel_message_id", ackPayload.ID),
			zap.String("ack", whatsappAck.String()),
			zap.String("status", string(newStatus)))
		return nil // Não falha, ACK é best-effort
	}

	// Atualiza status da mensagem usando métodos do domínio
	switch newStatus {
	case domainmessage.StatusDelivered:
		msg.MarkAsDelivered()
	case domainmessage.StatusRead:
		msg.MarkAsRead()
	case domainmessage.StatusPlayed:
		msg.MarkAsPlayed()
	case domainmessage.StatusSent:
		// Já está como sent por padrão, não precisa atualizar
	case domainmessage.StatusFailed:
		msg.MarkAsFailed()
	}

	// Salva mensagem atualizada
	if err := p.messageRepo.Save(ctx, msg); err != nil {
		p.logger.Warn("Failed to save message status from ACK",
			zap.Error(err),
			zap.String("channel_message_id", ackPayload.ID),
			zap.String("new_status", string(newStatus)),
			zap.String("ack", whatsappAck.String()))
		return nil // Não falha, ACK é best-effort
	}

	p.logger.Info("Message status updated from ACK",
		zap.String("message_id", msg.ID().String()),
		zap.String("channel_message_id", ackPayload.ID),
		zap.String("new_status", string(newStatus)),
		zap.String("ack", whatsappAck.String()),
		zap.Int("ack_value", whatsappAck.Value()),
		zap.String("ack_name", ackPayload.AckName))

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
// Eventos suportados:
// - label.upsert: cria/atualiza label no canal
// - label.deleted: remove label do canal
// - label.chat.added: adiciona label a um chat
// - label.chat.deleted: remove label de um chat
func (p *WAHARawEventProcessor) processLabelEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	p.logger.Info("Label event received",
		zap.String("raw_event_id", rawEvent.ID),
		zap.String("event", wahaEvent.Event),
		zap.String("session", wahaEvent.Session))

	// Busca canal pelo session
	ch, err := p.channelRepo.GetByExternalID(wahaEvent.Session)
	if err != nil {
		p.logger.Warn("Channel not found for label event",
			zap.String("session", wahaEvent.Session),
			zap.String("raw_event_id", rawEvent.ID),
			zap.Error(err))
		return nil // Não falha se canal não encontrado
	}

	// Processa baseado no tipo de evento
	switch wahaEvent.Event {
	case "label.upsert":
		return p.processLabelUpsert(ctx, ch, wahaEvent)
	case "label.deleted":
		return p.processLabelDeleted(ctx, ch, wahaEvent)
	case "label.chat.added":
		return p.processLabelChatAdded(ctx, ch, wahaEvent)
	case "label.chat.deleted":
		return p.processLabelChatDeleted(ctx, ch, wahaEvent)
	default:
		p.logger.Warn("Unknown label event type",
			zap.String("event", wahaEvent.Event),
			zap.String("raw_event_id", rawEvent.ID))
		return nil
	}
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

// processSessionStatusEvent processa eventos de status da sessão
//
// Eventos session.status são usados para atualizar automaticamente o status do canal:
// - STARTING → "connecting"
// - WORKING → "active"
// - STOPPED → "inactive"
// - FAILED → "error"
func (p *WAHARawEventProcessor) processSessionStatusEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload para extrair informações do status da sessão
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal session status payload: %w", err)
	}

	var statusPayload struct {
		Name   string `json:"name"`   // Nome da sessão
		Status string `json:"status"` // STARTING, WORKING, STOPPED, FAILED
	}

	if err := json.Unmarshal(payloadBytes, &statusPayload); err != nil {
		p.logger.Warn("Failed to unmarshal session status payload, skipping",
			zap.Error(err),
			zap.String("raw_event_id", rawEvent.ID))
		return nil // Não falha, apenas ignora status malformado
	}

	// Busca canal pelo ExternalID (que é o nome da sessão WAHA)
	ch, err := p.channelRepo.GetByExternalID(wahaEvent.Session)
	if err != nil {
		p.logger.Warn("Channel not found for session",
			zap.String("session", wahaEvent.Session),
			zap.String("raw_event_id", rawEvent.ID),
			zap.Error(err))
		return nil // Não falha se canal não encontrado (pode ser sessão de outro sistema)
	}

	// Mapeia status do WAHA para status do canal
	var newStatus domainchannel.ChannelStatus
	switch statusPayload.Status {
	case "STARTING":
		newStatus = domainchannel.StatusConnecting
	case "WORKING":
		newStatus = domainchannel.StatusActive
		ch.Activate() // Usa método do domínio
	case "STOPPED":
		newStatus = domainchannel.StatusInactive
		ch.Deactivate() // Usa método do domínio
	case "FAILED":
		newStatus = domainchannel.StatusError
		ch.SetError("WAHA session failed")
	default:
		p.logger.Debug("Unknown WAHA session status, skipping",
			zap.String("status", statusPayload.Status),
			zap.String("session", wahaEvent.Session))
		return nil
	}

	// Atualiza status da sessão WAHA no config
	if ch.Config == nil {
		ch.Config = make(map[string]interface{})
	}
	ch.Config["waha_session_status"] = statusPayload.Status

	// Salva canal atualizado
	if err := p.channelRepo.Update(ch); err != nil {
		p.logger.Error("Failed to update channel status from session event",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("session", wahaEvent.Session),
			zap.String("waha_status", statusPayload.Status),
			zap.String("new_status", string(newStatus)))
		return fmt.Errorf("failed to update channel: %w", err)
	}

	p.logger.Info("Channel status updated from session.status event",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session", wahaEvent.Session),
		zap.String("waha_status", statusPayload.Status),
		zap.String("channel_status", string(newStatus)))

	return nil
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

// extractAndSaveWhatsAppIdentifiers extrai e salva identificadores do WhatsApp como custom fields
func (p *WAHARawEventProcessor) extractAndSaveWhatsAppIdentifiers(ctx context.Context, rawEvent waha.WAHARawEvent, messageEvent waha.WAHAMessageEvent) {
	// 1. Extrai identificadores usando o IdentifierExtractor
	extractor := waha.NewIdentifierExtractor(p.logger)
	identifiers, err := extractor.ExtractFromMessageEvent(messageEvent)
	if err != nil {
		p.logger.Warn("Failed to extract WhatsApp identifiers",
			zap.Error(err),
			zap.String("raw_event_id", rawEvent.ID),
			zap.String("from", messageEvent.Payload.From))
		return // Não falha o processamento da mensagem, apenas não salva identifiers
	}

	// 2. Converte para custom fields
	customFields := identifiers.ToCustomFields()
	if len(customFields) == 0 {
		p.logger.Debug("No custom fields to save",
			zap.String("raw_event_id", rawEvent.ID))
		return
	}

	// 3. Extrai telefone do contato para buscar
	// Remove sufixos do WhatsApp para obter apenas o número
	contactPhone := messageEvent.Payload.From
	if normalizedPhone, err := domaincontact.NormalizeWhatsAppID(contactPhone); err == nil {
		contactPhone = normalizedPhone
	}

	// 4. Busca canal para obter ProjectID
	channel, err := p.channelRepo.GetByExternalID(messageEvent.Session)
	if err != nil {
		p.logger.Warn("Channel not found for WhatsApp identifiers",
			zap.Error(err),
			zap.String("session", messageEvent.Session))
		return
	}

	// 5. Busca contato pelo telefone e project
	contact, err := p.contactRepo.FindByPhone(ctx, channel.ProjectID, contactPhone)
	if err != nil {
		p.logger.Warn("Contact not found for WhatsApp identifiers",
			zap.Error(err),
			zap.String("phone", contactPhone),
			zap.String("project_id", channel.ProjectID.String()))
		return
	}

	// 6. Salva custom fields
	if err := p.contactRepo.SaveCustomFields(ctx, contact.ID(), customFields); err != nil {
		p.logger.Error("Failed to save WhatsApp identifiers as custom fields",
			zap.Error(err),
			zap.String("contact_id", contact.ID().String()),
			zap.Any("custom_fields", customFields))
		return
	}

	p.logger.Info("WhatsApp identifiers saved as custom fields",
		zap.String("contact_id", contact.ID().String()),
		zap.String("wid", identifiers.WID()),
		zap.Bool("has_lid", identifiers.HasLID()),
		zap.Bool("has_jid", identifiers.HasJID()),
		zap.Int("fields_saved", len(customFields)))
}

// processLabelUpsert processa evento de criação/atualização de label
func (p *WAHARawEventProcessor) processLabelUpsert(ctx context.Context, ch *domainchannel.Channel, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload para label data
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		p.logger.Warn("Failed to marshal label payload", zap.Error(err))
		return nil
	}

	var labelPayload waha.WAHALabelPayload
	if err := json.Unmarshal(payloadBytes, &labelPayload); err != nil {
		p.logger.Warn("Failed to unmarshal label payload", zap.Error(err))
		return nil
	}

	// Cria label do domínio
	label, err := domainchannel.NewLabel(labelPayload.ID, labelPayload.Name, labelPayload.Color, labelPayload.ColorHex)
	if err != nil {
		p.logger.Warn("Invalid label data from WAHA",
			zap.Error(err),
			zap.String("label_id", labelPayload.ID),
			zap.String("label_name", labelPayload.Name))
		return nil
	}

	// Adiciona/atualiza label no canal
	if err := ch.AddLabel(label); err != nil {
		p.logger.Error("Failed to add label to channel",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("label_id", label.ID))
		return nil
	}

	// Salva channel atualizado
	if err := p.channelRepo.Update(ch); err != nil {
		p.logger.Error("Failed to update channel with new label",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("label_id", label.ID))
		return fmt.Errorf("failed to update channel: %w", err)
	}

	p.logger.Info("Label upserted successfully",
		zap.String("channel_id", ch.ID.String()),
		zap.String("label_id", label.ID),
		zap.String("label_name", label.Name))

	return nil
}

// processLabelDeleted processa evento de remoção de label
func (p *WAHARawEventProcessor) processLabelDeleted(ctx context.Context, ch *domainchannel.Channel, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload para label data
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		p.logger.Warn("Failed to marshal label payload", zap.Error(err))
		return nil
	}

	var labelPayload waha.WAHALabelPayload
	if err := json.Unmarshal(payloadBytes, &labelPayload); err != nil {
		p.logger.Warn("Failed to unmarshal label payload", zap.Error(err))
		return nil
	}

	// Remove label do canal
	if err := ch.RemoveLabel(labelPayload.ID); err != nil {
		p.logger.Warn("Failed to remove label from channel",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("label_id", labelPayload.ID))
		return nil // Não falha se label não existir
	}

	// Salva channel atualizado
	if err := p.channelRepo.Update(ch); err != nil {
		p.logger.Error("Failed to update channel after deleting label",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("label_id", labelPayload.ID))
		return fmt.Errorf("failed to update channel: %w", err)
	}

	p.logger.Info("Label deleted successfully",
		zap.String("channel_id", ch.ID.String()),
		zap.String("label_id", labelPayload.ID))

	return nil
}

// processLabelChatAdded processa evento de label adicionada a um chat
func (p *WAHARawEventProcessor) processLabelChatAdded(ctx context.Context, ch *domainchannel.Channel, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		p.logger.Warn("Failed to marshal label chat payload", zap.Error(err))
		return nil
	}

	var chatPayload waha.WAHALabelChatPayload
	if err := json.Unmarshal(payloadBytes, &chatPayload); err != nil {
		p.logger.Warn("Failed to unmarshal label chat payload", zap.Error(err))
		return nil
	}

	// Busca chat pelo external ID (WhatsApp chat ID)
	chat, err := p.chatRepo.FindByExternalID(ctx, chatPayload.ChatID)
	if err != nil {
		p.logger.Debug("Chat not found for label association",
			zap.String("chat_id", chatPayload.ChatID),
			zap.String("label_id", chatPayload.LabelID),
			zap.Error(err))
		return nil // Chat pode não existir ainda
	}

	// Adiciona label ao chat
	if err := chat.AddLabel(chatPayload.LabelID); err != nil {
		p.logger.Warn("Failed to add label to chat",
			zap.Error(err),
			zap.String("chat_id", chat.ID().String()),
			zap.String("label_id", chatPayload.LabelID))
		return nil
	}

	// Salva chat atualizado (Update para chat existente)
	if err := p.chatRepo.Update(ctx, chat); err != nil {
		p.logger.Error("Failed to update chat with label",
			zap.Error(err),
			zap.String("chat_id", chat.ID().String()),
			zap.String("label_id", chatPayload.LabelID))
		return fmt.Errorf("failed to update chat: %w", err)
	}

	p.logger.Info("Label added to chat successfully",
		zap.String("chat_id", chat.ID().String()),
		zap.String("external_chat_id", chatPayload.ChatID),
		zap.String("label_id", chatPayload.LabelID))

	return nil
}

// processLabelChatDeleted processa evento de label removida de um chat
func (p *WAHARawEventProcessor) processLabelChatDeleted(ctx context.Context, ch *domainchannel.Channel, wahaEvent *waha.WAHAWebhookEvent) error {
	// Converte payload
	payloadBytes, err := json.Marshal(wahaEvent.Payload)
	if err != nil {
		p.logger.Warn("Failed to marshal label chat payload", zap.Error(err))
		return nil
	}

	var chatPayload waha.WAHALabelChatPayload
	if err := json.Unmarshal(payloadBytes, &chatPayload); err != nil {
		p.logger.Warn("Failed to unmarshal label chat payload", zap.Error(err))
		return nil
	}

	// Busca chat pelo external ID
	chat, err := p.chatRepo.FindByExternalID(ctx, chatPayload.ChatID)
	if err != nil {
		p.logger.Debug("Chat not found for label removal",
			zap.String("chat_id", chatPayload.ChatID),
			zap.String("label_id", chatPayload.LabelID),
			zap.Error(err))
		return nil // Chat pode não existir
	}

	// Remove label do chat
	if err := chat.RemoveLabel(chatPayload.LabelID); err != nil {
		p.logger.Warn("Failed to remove label from chat",
			zap.Error(err),
			zap.String("chat_id", chat.ID().String()),
			zap.String("label_id", chatPayload.LabelID))
		return nil // Não falha se label não existir
	}

	// Salva chat atualizado (Update para chat existente)
	if err := p.chatRepo.Update(ctx, chat); err != nil {
		p.logger.Error("Failed to update chat after removing label",
			zap.Error(err),
			zap.String("chat_id", chat.ID().String()),
			zap.String("label_id", chatPayload.LabelID))
		return fmt.Errorf("failed to update chat: %w", err)
	}

	p.logger.Info("Label removed from chat successfully",
		zap.String("chat_id", chat.ID().String()),
		zap.String("external_chat_id", chatPayload.ChatID),
		zap.String("label_id", chatPayload.LabelID))

	return nil
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
