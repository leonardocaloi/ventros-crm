package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/outbox"
	"go.temporal.io/sdk/activity"
)

// ProcessPendingEventsInput são os parâmetros para processar eventos pendentes.
type ProcessPendingEventsInput struct {
	BatchSize int `json:"batch_size"`
}

// ProcessPendingEventsResult é o resultado do processamento.
type ProcessPendingEventsResult struct {
	EventsProcessed int `json:"events_processed"`
	EventsFailed    int `json:"events_failed"`
}

// ProcessFailedEventsInput são os parâmetros para processar eventos falhados.
type ProcessFailedEventsInput struct {
	BatchSize    int           `json:"batch_size"`
	MaxRetries   int           `json:"max_retries"`
	RetryBackoff time.Duration `json:"retry_backoff"`
}

// ProcessFailedEventsResult é o resultado do retry.
type ProcessFailedEventsResult struct {
	EventsRetried   int `json:"events_retried"`
	EventsSucceeded int `json:"events_succeeded"`
}

// OutboxActivities contém as dependências para as activities.
type OutboxActivities struct {
	outboxRepo      outbox.Repository
	eventPublisher  EventPublisher
	webhookNotifier WebhookNotifier
}

// EventPublisher é a interface para publicar eventos no RabbitMQ.
type EventPublisher interface {
	PublishRaw(ctx context.Context, queue string, eventData []byte) error
}

// WebhookNotifier é a interface para notificar webhooks HTTP.
type WebhookNotifier interface {
	NotifyWebhooks(ctx context.Context, eventType string, eventData interface{})
}

// NewOutboxActivities cria uma nova instância das activities.
func NewOutboxActivities(outboxRepo outbox.Repository, eventPublisher EventPublisher, webhookNotifier WebhookNotifier) *OutboxActivities {
	return &OutboxActivities{
		outboxRepo:      outboxRepo,
		eventPublisher:  eventPublisher,
		webhookNotifier: webhookNotifier,
	}
}

// ProcessPendingEventsActivity processa eventos pendentes do outbox.
func (a *OutboxActivities) ProcessPendingEventsActivity(
	ctx context.Context,
	input ProcessPendingEventsInput,
) (ProcessPendingEventsResult, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Processing pending events from outbox", "batch_size", input.BatchSize)

	// 1. Buscar eventos pendentes
	events, err := a.outboxRepo.GetPendingEvents(ctx, input.BatchSize)
	if err != nil {
		return ProcessPendingEventsResult{}, fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(events) == 0 {
		return ProcessPendingEventsResult{}, nil
	}

	logger.Info("Found pending events", "count", len(events))

	// 2. Processar cada evento
	var processed, failed int
	for _, event := range events {
		// Marca como "processing" (lock otimista)
		if err := a.outboxRepo.MarkAsProcessing(ctx, event.EventID); err != nil {
			logger.Warn("Failed to mark event as processing (may already be processed)",
				"event_id", event.EventID,
				"error", err)
			continue
		}

		// Publica no RabbitMQ
		if err := a.publishEvent(ctx, event); err != nil {
			logger.Error("Failed to publish event",
				"event_id", event.EventID,
				"event_type", event.EventType,
				"error", err)

			// Marca como falho
			if markErr := a.outboxRepo.MarkAsFailed(ctx, event.EventID, err.Error()); markErr != nil {
				logger.Error("Failed to mark event as failed", "error", markErr)
			}
			failed++
			continue
		}

		// Marca como processado
		if err := a.outboxRepo.MarkAsProcessed(ctx, event.EventID); err != nil {
			logger.Error("Failed to mark event as processed", "event_id", event.EventID, "error", err)
		}
		processed++

		logger.Debug("Event published successfully",
			"event_id", event.EventID,
			"event_type", event.EventType)
	}

	logger.Info("Finished processing pending events",
		"processed", processed,
		"failed", failed)

	return ProcessPendingEventsResult{
		EventsProcessed: processed,
		EventsFailed:    failed,
	}, nil
}

// ProcessFailedEventsActivity processa eventos falhados para retry.
func (a *OutboxActivities) ProcessFailedEventsActivity(
	ctx context.Context,
	input ProcessFailedEventsInput,
) (ProcessFailedEventsResult, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Processing failed events for retry",
		"batch_size", input.BatchSize,
		"max_retries", input.MaxRetries,
		"retry_backoff", input.RetryBackoff)

	// 1. Buscar eventos falhados que podem ser tentados novamente
	events, err := a.outboxRepo.GetFailedEventsForRetry(
		ctx,
		input.MaxRetries,
		input.RetryBackoff,
		input.BatchSize,
	)
	if err != nil {
		return ProcessFailedEventsResult{}, fmt.Errorf("failed to get failed events: %w", err)
	}

	if len(events) == 0 {
		return ProcessFailedEventsResult{}, nil
	}

	logger.Info("Found failed events to retry", "count", len(events))

	// 2. Tentar novamente cada evento
	var retried, succeeded int
	for _, event := range events {
		retried++

		// Tenta publicar novamente
		if err := a.publishEvent(ctx, event); err != nil {
			logger.Error("Retry failed",
				"event_id", event.EventID,
				"retry_count", event.RetryCount,
				"error", err)

			// Marca como falho novamente (incrementa retry count)
			if markErr := a.outboxRepo.MarkAsFailed(ctx, event.EventID, err.Error()); markErr != nil {
				logger.Error("Failed to update retry count", "error", markErr)
			}
			continue
		}

		// Marca como processado
		if err := a.outboxRepo.MarkAsProcessed(ctx, event.EventID); err != nil {
			logger.Error("Failed to mark event as processed", "event_id", event.EventID, "error", err)
		}
		succeeded++

		logger.Info("Event retry succeeded",
			"event_id", event.EventID,
			"retry_count", event.RetryCount)
	}

	logger.Info("Finished retrying failed events",
		"retried", retried,
		"succeeded", succeeded)

	return ProcessFailedEventsResult{
		EventsRetried:   retried,
		EventsSucceeded: succeeded,
	}, nil
}

// publishEvent publica um evento no RabbitMQ e notifica webhooks.
func (a *OutboxActivities) publishEvent(ctx context.Context, event *outbox.OutboxEvent) error {
	logger := activity.GetLogger(ctx)

	// 1. Publica no RabbitMQ
	queue := fmt.Sprintf("domain.events.%s", event.EventType)
	if err := a.eventPublisher.PublishRaw(ctx, queue, event.EventData); err != nil {
		return fmt.Errorf("failed to publish to RabbitMQ: %w", err)
	}

	// 2. Notifica webhooks HTTP
	if a.webhookNotifier != nil {
		// Deserializa o evento como JSON genérico para passar para os webhooks
		var rawData interface{}
		if err := json.Unmarshal(event.EventData, &rawData); err != nil {
			logger.Error("Failed to unmarshal event data for webhook notification",
				"event_id", event.EventID,
				"event_type", event.EventType,
				"error", err)
			// Continua mesmo com erro de unmarshal - não bloqueia o processamento
		} else {
			// Mapeia evento de domínio para eventos de negócio (webhooks)
			businessEvents := mapDomainToBusinessEvents(event.EventType)
			for _, businessEvent := range businessEvents {
				if businessEvent != "" {
					logger.Info("Notifying webhooks for event",
						"business_event", businessEvent,
						"domain_event", event.EventType)
					a.webhookNotifier.NotifyWebhooks(ctx, businessEvent, rawData)
				}
			}
		}
	}

	return nil
}

// RegisterActivities retorna as activities para registro no worker Temporal.
func (a *OutboxActivities) RegisterActivities() []interface{} {
	return []interface{}{
		a.ProcessPendingEventsActivity,
		a.ProcessFailedEventsActivity,
	}
}

// CleanupOldEventsActivity limpa eventos antigos já processados (manutenção).
type CleanupOldEventsInput struct {
	OlderThan time.Duration `json:"older_than"` // Remove eventos processados há mais de X tempo
}

type CleanupOldEventsResult struct {
	EventsCleaned int64 `json:"events_cleaned"`
}

func (a *OutboxActivities) CleanupOldEventsActivity(
	ctx context.Context,
	input CleanupOldEventsInput,
) (CleanupOldEventsResult, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Cleaning up old processed events", "older_than", input.OlderThan)

	// Não implementado no repositório atual, mas seria algo assim:
	// count, err := a.outboxRepo.DeleteProcessedOlderThan(ctx, input.OlderThan)

	// Por enquanto, apenas retorna
	logger.Info("Cleanup not yet implemented")

	return CleanupOldEventsResult{
		EventsCleaned: 0,
	}, nil
}

// GetOutboxMetricsActivity retorna métricas do outbox para monitoramento.
type OutboxMetrics struct {
	PendingCount int64 `json:"pending_count"`
	FailedCount  int64 `json:"failed_count"`
}

func (a *OutboxActivities) GetOutboxMetricsActivity(ctx context.Context) (OutboxMetrics, error) {
	pending, err := a.outboxRepo.CountPending(ctx)
	if err != nil {
		return OutboxMetrics{}, fmt.Errorf("failed to count pending: %w", err)
	}

	failed, err := a.outboxRepo.CountFailed(ctx)
	if err != nil {
		return OutboxMetrics{}, fmt.Errorf("failed to count failed: %w", err)
	}

	return OutboxMetrics{
		PendingCount: pending,
		FailedCount:  failed,
	}, nil
}

// mapDomainToBusinessEvents mapeia eventos de domínio para eventos de negócio (webhooks).
// Um evento de domínio pode gerar múltiplos eventos de negócio.
func mapDomainToBusinessEvents(domainEvent string) []string {
	switch domainEvent {
	// Eventos de contato
	case "contact.created":
		return []string{"contact.created"}
	case "contact.updated":
		return []string{"contact.updated"}
	case "contact.merged":
		return []string{"contact.merged"}
	case "contact.enriched":
		return []string{"contact.enriched"}
	case "contact.deleted":
		return []string{"contact.deleted"}
	case "contact.profile_picture_updated":
		return []string{"contact.profile_picture_updated"}

	// Eventos de pipeline e status
	case "contact.status_changed":
		return []string{"contact.status_changed"}
	case "contact.entered_pipeline":
		return []string{"contact.entered_pipeline"}
	case "contact.exited_pipeline":
		return []string{"contact.exited_pipeline"}
	case "contact.pipeline_status_changed":
		return []string{"pipeline.status.changed"}

	// Eventos de sessão
	case "session.started":
		return []string{"session.created"}
	case "session.ended":
		return []string{"session.closed"}
	case "session.agent_assigned":
		return []string{"session.agent_assigned"}
	case "session.resolved":
		return []string{"session.resolved"}
	case "session.escalated":
		return []string{"session.escalated"}
	case "session.summarized":
		return []string{"session.summarized"}
	case "session.abandoned":
		return []string{"session.abandoned"}

	// Eventos de mensagem
	case "message.created":
		return []string{"message.received"} // Mensagens criadas são sempre recebidas inicialmente
	case "message.sent":
		return []string{"message.sent"}
	case "message.delivered":
		return []string{"message.delivered"}
	case "message.read":
		return []string{"message.read"}
	case "message.failed":
		return []string{"message.failed"}

	// Eventos de tracking
	case "tracking.message.meta_ads":
		return []string{"tracking.message.meta_ads"}
	case "tracking.created":
		return []string{"tracking.created"}
	case "tracking.enriched":
		return []string{"tracking.enriched"}

	// Eventos de agente
	case "agent.created":
		return []string{"agent.created"}
	case "agent.updated":
		return []string{"agent.updated"}
	case "agent.activated":
		return []string{"agent.activated"}
	case "agent.deactivated":
		return []string{"agent.deactivated"}
	case "agent.logged_in":
		return []string{"agent.logged_in"}
	case "agent.permission_granted":
		return []string{"agent.permission_granted"}
	case "agent.permission_revoked":
		return []string{"agent.permission_revoked"}

	// Eventos de canal
	case "channel.created":
		return []string{"channel.created"}
	case "channel.activated":
		return []string{"channel.activated"}
	case "channel.deactivated":
		return []string{"channel.deactivated"}
	case "channel.deleted":
		return []string{"channel.deleted"}
	case "channel.pipeline.associated":
		return []string{"channel.pipeline_associated"}
	case "channel.pipeline.disassociated":
		return []string{"channel.pipeline_disassociated"}

	// Eventos de nota
	case "note.added":
		return []string{"note.added"}
	case "note.updated":
		return []string{"note.updated"}
	case "note.deleted":
		return []string{"note.deleted"}
	case "note.pinned":
		return []string{"note.pinned"}

	// Eventos de pipeline
	case "pipeline.created":
		return []string{"pipeline.created"}
	case "pipeline.updated":
		return []string{"pipeline.updated"}
	case "pipeline.activated":
		return []string{"pipeline.activated"}
	case "pipeline.deactivated":
		return []string{"pipeline.deactivated"}
	case "status.created":
		return []string{"pipeline.status.created"}
	case "status.updated":
		return []string{"pipeline.status.updated"}
	case "status.activated":
		return []string{"pipeline.status.activated"}
	case "status.deactivated":
		return []string{"pipeline.status.deactivated"}
	case "pipeline.status_added":
		return []string{"pipeline.status.added"}
	case "pipeline.status_removed":
		return []string{"pipeline.status.removed"}

	// Eventos de billing (cobrança)
	case "billing.account_created":
		return []string{"billing.account_created"}
	case "billing.payment_method_activated":
		return []string{"billing.payment_method_activated"}
	case "billing.account_suspended":
		return []string{"billing.account_suspended"}
	case "billing.account_reactivated":
		return []string{"billing.account_reactivated"}
	case "billing.account_canceled":
		return []string{"billing.account_canceled"}

	// Eventos de credential (CRÍTICO para OAuth Meta)
	case "credential.created":
		return []string{"credential.created"}
	case "credential.updated":
		return []string{"credential.updated"}
	case "credential.oauth_refreshed":
		return []string{"credential.oauth_token_refreshed"}
	case "credential.activated":
		return []string{"credential.activated"}
	case "credential.deactivated":
		return []string{"credential.deactivated"}
	case "credential.used":
		return []string{"credential.used"}
	case "credential.expired":
		return []string{"credential.expired"}

	// Eventos de agent session
	case "agent_session.joined":
		return []string{"agent_session.joined"}
	case "agent_session.left":
		return []string{"agent_session.left"}
	case "agent_session.role_changed":
		return []string{"agent_session.role_changed"}

	// Eventos de contact list
	case "contact_list.created":
		return []string{"contact_list.created"}
	case "contact_list.updated":
		return []string{"contact_list.updated"}
	case "contact_list.deleted":
		return []string{"contact_list.deleted"}
	case "contact_list.filter_rule_added":
		return []string{"contact_list.filter_rule_added"}
	case "contact_list.filter_rule_removed":
		return []string{"contact_list.filter_rule_removed"}
	case "contact_list.filter_rules_cleared":
		return []string{"contact_list.filter_rules_cleared"}
	case "contact_list.recalculated":
		return []string{"contact_list.recalculated"}
	case "contact_list.contact_added":
		return []string{"contact_list.contact_added"}
	case "contact_list.contact_removed":
		return []string{"contact_list.contact_removed"}

	// Eventos de automation
	case "automation.created":
		return []string{"automation.created"}
	case "automation.enabled":
		return []string{"automation.enabled"}
	case "automation.disabled":
		return []string{"automation.disabled"}
	case "automation_rule.triggered":
		return []string{"automation.rule_triggered"}
	case "automation_rule.executed":
		return []string{"automation.rule_executed"}
	case "automation_rule.failed":
		return []string{"automation.rule_failed"}

	// Eventos de channel type
	case "channel_type.created":
		return []string{"channel_type.created"}
	case "channel_type.activated":
		return []string{"channel_type.activated"}
	case "channel_type.deactivated":
		return []string{"channel_type.deactivated"}

	// Eventos de project
	case "project.created":
		return []string{"project.created"}

	// Eventos de customer (tenant)
	case "customer.created":
		return []string{"customer.created"}
	case "customer.activated":
		return []string{"customer.activated"}
	case "customer.suspended":
		return []string{"customer.suspended"}

	// Eventos de AI processing
	case "message.ai.process_image_requested":
		return []string{"ai.process_image_requested"}
	case "message.ai.process_video_requested":
		return []string{"ai.process_video_requested"}
	case "message.ai.process_audio_requested":
		return []string{"ai.process_audio_requested"}
	case "message.ai.process_voice_requested":
		return []string{"ai.process_voice_requested"}

	// Eventos internos que não notificam webhooks
	case "session.message_recorded":
		return []string{}

	default:
		// Eventos não mapeados: usa o nome original
		return []string{domainEvent}
	}
}
