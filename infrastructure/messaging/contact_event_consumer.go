package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	contacteventapp "github.com/caloi/ventros-crm/internal/application/contact_event"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/caloi/ventros-crm/internal/domain/contact_event"
	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ContactEventConsumer consome Domain Events e cria Contact Events para a timeline
// Seguindo DDD: Application Service que coordena entre Domain Events e Contact Events
type ContactEventConsumer struct {
	conn                      *RabbitMQConnection
	createContactEventUseCase *contacteventapp.CreateContactEventUseCase
	idempotencyChecker        IdempotencyChecker
	logger                    *zap.Logger
}

// IdempotencyChecker interface for checking event idempotency
type IdempotencyChecker interface {
	IsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error)
	MarkAsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string, processingDurationMs *int) error
}

func NewContactEventConsumer(
	conn *RabbitMQConnection,
	createContactEventUseCase *contacteventapp.CreateContactEventUseCase,
	idempotencyChecker IdempotencyChecker,
	logger *zap.Logger,
) *ContactEventConsumer {
	return &ContactEventConsumer{
		conn:                      conn,
		createContactEventUseCase: createContactEventUseCase,
		idempotencyChecker:        idempotencyChecker,
		logger:                    logger,
	}
}

// Start inicia o consumo de eventos de domínio
func (c *ContactEventConsumer) Start(ctx context.Context) error {
	// Iniciar consumers para cada tipo de evento
	consumers := []struct {
		queueName string
		consumer  Consumer
	}{
		{"domain.events.contact.created", &contactCreatedConsumer{c}},
		{"domain.events.contact.updated", &contactUpdatedConsumer{c}},
		{"domain.events.contact.profile_picture_updated", &contactProfilePictureUpdatedConsumer{c}},
		{"domain.events.contact.status_changed", &contactStatusChangedConsumer{c}},
		{"domain.events.contact.pipeline_status_changed", &contactPipelineStatusChangedConsumer{c}},
		{"domain.events.contact.entered_pipeline", &contactEnteredPipelineConsumer{c}},
		{"domain.events.contact.exited_pipeline", &contactExitedPipelineConsumer{c}},
		{"domain.events.session.started", &sessionStartedConsumer{c}},
		{"domain.events.session.ended", &sessionEndedConsumer{c}},
		{"domain.events.tracking.message.meta_ads", &trackingAdConversionConsumer{c}},
		{"domain.events.note.added", &noteAddedConsumer{c}},
	}

	for _, cfg := range consumers {
		consumerTag := fmt.Sprintf("contact-event-consumer-%s-%s", cfg.queueName, uuid.New().String()[:8])

		if err := c.conn.StartConsumer(ctx, cfg.queueName, consumerTag, cfg.consumer, 10); err != nil {
			c.logger.Error("Failed to start consumer",
				zap.String("queue", cfg.queueName),
				zap.Error(err))
			return err
		}

		c.logger.Info("Consumer started",
			zap.String("queue", cfg.queueName),
			zap.String("consumer_tag", consumerTag))
	}

	return nil
}

// contactCreatedConsumer processa eventos de contato criado
type contactCreatedConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactCreatedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event contact.ContactCreatedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactCreatedEvent", zap.Error(err))
		return err
	}

	title := fmt.Sprintf("Contato %s criado", event.Name)
	description := "Novo contato adicionado ao sistema"

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:   event.ContactID,
		TenantID:    event.TenantID,
		EventType:   contact_event.EventTypeContactCreated,
		Category:    contact_event.CategorySystem,
		Priority:    contact_event.PriorityNormal,
		Source:      contact_event.SourceSystem,
		Title:       &title,
		Description: &description,
		Payload: map[string]interface{}{
			"contact_name": event.Name,
			"project_id":   event.ProjectID.String(),
		},
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	c.parent.logger.Debug("Contact event created for ContactCreated",
		zap.String("contact_id", event.ContactID.String()))
	return nil
}

// contactUpdatedConsumer processa eventos de contato atualizado
type contactUpdatedConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactUpdatedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event contact.ContactUpdatedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactUpdatedEvent", zap.Error(err))
		return err
	}

	title := "Contato atualizado"
	description := "Informações do contato foram modificadas"

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:       event.ContactID,
		TenantID:        "", // TODO: Adicionar TenantID no evento
		EventType:       contact_event.EventTypeContactUpdated,
		Category:        contact_event.CategorySystem,
		Priority:        contact_event.PriorityLow,
		Source:          contact_event.SourceSystem,
		Title:           &title,
		Description:     &description,
		IsRealtime:      false,
		VisibleToClient: false,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	return nil
}

// contactStatusChangedConsumer processa eventos de mudança de status
type contactStatusChangedConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactStatusChangedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event pipeline.ContactStatusChangedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactStatusChangedEvent", zap.Error(err))
		return err
	}

	title := fmt.Sprintf("Status alterado para %s", event.NewStatusName)

	var description string
	if event.OldStatusName != nil {
		description = fmt.Sprintf("Status mudou de %s para %s", *event.OldStatusName, event.NewStatusName)
	} else {
		description = fmt.Sprintf("Status definido como %s", event.NewStatusName)
	}

	if event.Reason != "" {
		description += fmt.Sprintf(". Motivo: %s", event.Reason)
	}

	payload := map[string]interface{}{
		"new_status_id":   event.NewStatusID.String(),
		"new_status_name": event.NewStatusName,
		"pipeline_id":     event.PipelineID.String(),
	}

	if event.OldStatusID != nil {
		payload["old_status_id"] = event.OldStatusID.String()
	}
	if event.OldStatusName != nil {
		payload["old_status_name"] = *event.OldStatusName
	}
	if event.Reason != "" {
		payload["reason"] = event.Reason
	}

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:       event.ContactID,
		TenantID:        "", // TODO: Adicionar TenantID no evento
		EventType:       contact_event.EventTypeStatusChanged,
		Category:        contact_event.CategoryStatus,
		Priority:        contact_event.PriorityHigh,
		Source:          contact_event.SourceSystem,
		Title:           &title,
		Description:     &description,
		Payload:         payload,
		TriggeredBy:     event.ChangedBy,
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	c.parent.logger.Debug("Contact event created for ContactStatusChanged",
		zap.String("contact_id", event.ContactID.String()),
		zap.String("new_status", event.NewStatusName))
	return nil
}

// contactEnteredPipelineConsumer processa eventos de entrada no pipeline
type contactEnteredPipelineConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactEnteredPipelineConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event pipeline.ContactEnteredPipelineEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactEnteredPipelineEvent", zap.Error(err))
		return err
	}

	title := "Contato entrou no pipeline"
	description := fmt.Sprintf("Contato adicionado ao pipeline no estágio %s", event.StatusName)

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:   event.ContactID,
		TenantID:    "", // TODO: Adicionar TenantID no evento
		EventType:   contact_event.EventTypeEnteredPipeline,
		Category:    contact_event.CategoryPipeline,
		Priority:    contact_event.PriorityHigh,
		Source:      contact_event.SourceSystem,
		Title:       &title,
		Description: &description,
		Payload: map[string]interface{}{
			"pipeline_id": event.PipelineID.String(),
			"status_id":   event.StatusID.String(),
			"status_name": event.StatusName,
		},
		TriggeredBy:     event.EnteredBy,
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	return nil
}

// contactExitedPipelineConsumer processa eventos de saída do pipeline
type contactExitedPipelineConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactExitedPipelineConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event pipeline.ContactExitedPipelineEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactExitedPipelineEvent", zap.Error(err))
		return err
	}

	title := "Contato saiu do pipeline"
	description := fmt.Sprintf("Contato removido do pipeline (último estágio: %s)", event.LastStatusName)

	if event.Reason != "" {
		description += fmt.Sprintf(". Motivo: %s", event.Reason)
	}

	payload := map[string]interface{}{
		"pipeline_id":      event.PipelineID.String(),
		"last_status_id":   event.LastStatusID.String(),
		"last_status_name": event.LastStatusName,
	}

	if event.Reason != "" {
		payload["reason"] = event.Reason
	}

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:       event.ContactID,
		TenantID:        "", // TODO: Adicionar TenantID no evento
		EventType:       contact_event.EventTypeExitedPipeline,
		Category:        contact_event.CategoryPipeline,
		Priority:        contact_event.PriorityNormal,
		Source:          contact_event.SourceSystem,
		Title:           &title,
		Description:     &description,
		Payload:         payload,
		TriggeredBy:     event.ExitedBy,
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	return nil
}

// sessionStartedConsumer processa eventos de sessão iniciada
type sessionStartedConsumer struct {
	parent *ContactEventConsumer
}

func (c *sessionStartedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event session.SessionStartedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal SessionStartedEvent", zap.Error(err))
		return err
	}

	title := "Nova sessão iniciada"
	description := "Sessão de atendimento iniciada"

	sessionID := event.SessionID

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:   event.ContactID,
		SessionID:   &sessionID,
		TenantID:    event.TenantID,
		EventType:   contact_event.EventTypeSessionStarted,
		Category:    contact_event.CategorySession,
		Priority:    contact_event.PriorityNormal,
		Source:      contact_event.SourceSystem,
		Title:       &title,
		Description: &description,
		Payload: map[string]interface{}{
			"session_id": event.SessionID.String(),
		},
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()),
			zap.String("session_id", event.SessionID.String()))
		return err
	}

	return nil
}

// sessionEndedConsumer processa eventos de sessão encerrada
type sessionEndedConsumer struct {
	parent *ContactEventConsumer
}

func (c *sessionEndedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event session.SessionEndedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal SessionEndedEvent", zap.Error(err))
		return err
	}

	title := "Sessão encerrada"
	description := fmt.Sprintf("Sessão de atendimento finalizada. Motivo: %s", event.Reason)

	sessionID := event.SessionID

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:   uuid.Nil, // TODO: SessionEndedEvent precisa ter ContactID
		SessionID:   &sessionID,
		TenantID:    "", // TODO: SessionEndedEvent precisa ter TenantID
		EventType:   contact_event.EventTypeSessionEnded,
		Category:    contact_event.CategorySession,
		Priority:    contact_event.PriorityLow,
		Source:      contact_event.SourceSystem,
		Title:       &title,
		Description: &description,
		Payload: map[string]interface{}{
			"session_id":       event.SessionID.String(),
			"reason":           string(event.Reason),
			"duration_seconds": event.Duration,
		},
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	// TODO: Buscar ContactID da sessão antes de criar o evento
	// Por enquanto, vamos logar e pular
	if cmd.ContactID == uuid.Nil {
		c.parent.logger.Warn("SessionEndedEvent without ContactID, skipping contact event creation",
			zap.String("session_id", event.SessionID.String()))
		return nil
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("session_id", event.SessionID.String()))
		return err
	}

	return nil
}

// contactPipelineStatusChangedConsumer processa eventos de mudança de status no pipeline
type contactPipelineStatusChangedConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactPipelineStatusChangedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event contact.ContactPipelineStatusChangedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactPipelineStatusChangedEvent", zap.Error(err))
		return err
	}

	var title string
	var description string

	if event.IsFirstStatus() {
		title = fmt.Sprintf("Entrou no pipeline: %s", event.NewStatusName)
		description = fmt.Sprintf("Contato adicionado ao pipeline com status: %s", event.NewStatusName)
	} else {
		title = fmt.Sprintf("Status mudou: %s → %s", event.PreviousStatusName, event.NewStatusName)
		description = fmt.Sprintf("Status do pipeline alterado de %s para %s", event.PreviousStatusName, event.NewStatusName)
	}

	if event.Reason != "" {
		description += fmt.Sprintf(". Motivo: %s", event.Reason)
	}

	payload := map[string]interface{}{
		"pipeline_id":     event.PipelineID.String(),
		"new_status_id":   event.NewStatusID.String(),
		"new_status_name": event.NewStatusName,
	}

	if event.PreviousStatusID != nil {
		payload["previous_status_id"] = event.PreviousStatusID.String()
		payload["previous_status_name"] = event.PreviousStatusName
	}
	if event.Reason != "" {
		payload["reason"] = event.Reason
	}

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:       event.ContactID,
		TenantID:        event.TenantID,
		EventType:       contact_event.EventTypePipelineStageChanged,
		Category:        contact_event.CategoryPipeline,
		Priority:        contact_event.PriorityHigh,
		Source:          contact_event.SourceSystem,
		Title:           &title,
		Description:     &description,
		Payload:         payload,
		TriggeredBy:     event.ChangedBy,
		IsRealtime:      true,
		VisibleToClient: true,
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	c.parent.logger.Debug("Contact event created for PipelineStatusChanged",
		zap.String("contact_id", event.ContactID.String()),
		zap.String("new_status", event.NewStatusName))
	return nil
}

// trackingAdConversionConsumer processa eventos de conversão de anúncios
type trackingAdConversionConsumer struct {
	parent *ContactEventConsumer
}

func (c *trackingAdConversionConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event contact.AdConversionTrackedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal AdConversionTrackedEvent", zap.Error(err))
		return err
	}

	title := event.GetTitle()
	description := event.GetDescription()

	sessionID := event.SessionID

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:         event.ContactID,
		SessionID:         &sessionID,
		TenantID:          event.TenantID,
		EventType:         "ad_conversion_tracked",
		Category:          contact_event.CategorySystem,
		Priority:          contact_event.PriorityHigh,
		Source:            contact_event.SourceIntegration,
		Title:             &title,
		Description:       &description,
		Payload:           event.ToContactEventPayload(),
		IntegrationSource: &event.ConversionSource,
		IsRealtime:        true,
		VisibleToClient:   false, // Tracking é interno
		VisibleToAgent:    true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	c.parent.logger.Debug("Contact event created for AdConversionTracked",
		zap.String("contact_id", event.ContactID.String()),
		zap.String("conversion_source", event.ConversionSource))
	return nil
}

// contactProfilePictureUpdatedConsumer processa eventos de foto de perfil atualizada
type contactProfilePictureUpdatedConsumer struct {
	parent *ContactEventConsumer
}

func (c *contactProfilePictureUpdatedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event contact.ContactProfilePictureUpdatedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ContactProfilePictureUpdatedEvent", zap.Error(err))
		return err
	}

	title := "Foto de perfil atualizada"
	description := "Foto de perfil do WhatsApp foi sincronizada"

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:   event.ContactID,
		TenantID:    event.TenantID,
		EventType:   "profile_picture_updated",
		Category:    contact_event.CategorySystem,
		Priority:    contact_event.PriorityLow,
		Source:      contact_event.SourceIntegration,
		Title:       &title,
		Description: &description,
		Payload: map[string]interface{}{
			"profile_picture_url": event.ProfilePictureURL,
		},
		IntegrationSource: stringPtr("whatsapp"),
		IsRealtime:        false,
		VisibleToClient:   false, // Interno
		VisibleToAgent:    true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	c.parent.logger.Debug("Contact event created for ProfilePictureUpdated",
		zap.String("contact_id", event.ContactID.String()))
	return nil
}

// noteAddedConsumer processa eventos de nota adicionada
type noteAddedConsumer struct {
	parent *ContactEventConsumer
}

func (c *noteAddedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event struct {
		NoteID     uuid.UUID  `json:"note_id"`
		ContactID  uuid.UUID  `json:"contact_id"`
		SessionID  *uuid.UUID `json:"session_id"`
		TenantID   string     `json:"tenant_id"`
		AuthorID   uuid.UUID  `json:"author_id"`
		AuthorType string     `json:"author_type"`
		AuthorName string     `json:"author_name"`
		Content    string     `json:"content"`
		NoteType   string     `json:"note_type"`
		Priority   string     `json:"priority"`
		AddedAt    time.Time  `json:"added_at"`
	}

	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal NoteAddedEvent", zap.Error(err))
		return err
	}

	title := "Nota adicionada"
	description := fmt.Sprintf("%s adicionou uma nota", event.AuthorName)

	cmd := contacteventapp.CreateContactEventCommand{
		ContactID:   event.ContactID,
		SessionID:   event.SessionID,
		TenantID:    event.TenantID,
		EventType:   "note_added",
		Category:    contact_event.CategoryNote,
		Priority:    contact_event.Priority(event.Priority),
		Source:      contact_event.SourceSystem,
		Title:       &title,
		Description: &description,
		Payload: map[string]interface{}{
			"note_id":         event.NoteID.String(),
			"author_id":       event.AuthorID.String(),
			"author_type":     event.AuthorType,
			"author_name":     event.AuthorName,
			"note_type":       event.NoteType,
			"content_preview": truncateString(event.Content, 100),
		},
		IsRealtime:      false,
		VisibleToClient: false, // Notas são internas por padrão
		VisibleToAgent:  true,
	}

	_, err := c.parent.createContactEventUseCase.Execute(ctx, cmd)
	if err != nil {
		c.parent.logger.Error("Failed to create contact event",
			zap.Error(err),
			zap.String("contact_id", event.ContactID.String()))
		return err
	}

	c.parent.logger.Debug("Contact event created for NoteAdded",
		zap.String("contact_id", event.ContactID.String()),
		zap.String("note_id", event.NoteID.String()))
	return nil
}

// idempotentConsumerWrapper wraps a consumer with idempotency checking
type idempotentConsumerWrapper struct {
	consumer           Consumer
	idempotencyChecker IdempotencyChecker
	consumerName       string
	extractEventID     func([]byte) (uuid.UUID, error)
}

func (w *idempotentConsumerWrapper) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	startTime := time.Now()

	// Extract event ID
	eventID, err := w.extractEventID(delivery.Body)
	if err != nil {
		// Se não conseguir extrair ID, processa sem idempotência
		return w.consumer.ProcessMessage(ctx, delivery)
	}

	// Check if already processed
	if w.idempotencyChecker != nil {
		processed, err := w.idempotencyChecker.IsProcessed(ctx, eventID, w.consumerName)
		if err != nil {
			// Log error but continue processing (fail-open)
			fmt.Printf("⚠️  Failed to check idempotency: %v\n", err)
		} else if processed {
			fmt.Printf("⏭️  Event already processed, skipping: consumer=%s, event_id=%s\n", w.consumerName, eventID)
			return nil
		}
	}

	// Process message
	if err := w.consumer.ProcessMessage(ctx, delivery); err != nil {
		return err
	}

	// Mark as processed
	if w.idempotencyChecker != nil {
		duration := int(time.Since(startTime).Milliseconds())
		if err := w.idempotencyChecker.MarkAsProcessed(ctx, eventID, w.consumerName, &duration); err != nil {
			fmt.Printf("⚠️  Failed to mark as processed: %v\n", err)
		}
	}

	return nil
}

// extractDomainEventID extracts event ID from domain event JSON
func extractDomainEventID(data []byte) (uuid.UUID, error) {
	var event struct {
		EventID uuid.UUID `json:"event_id"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return uuid.Nil, err
	}
	if event.EventID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("event_id is nil")
	}
	return event.EventID, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
