package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence"
	"github.com/ventros/crm/infrastructure/webhooks"
	"github.com/ventros/crm/internal/domain/core/outbox"
	"github.com/ventros/crm/internal/domain/core/saga"
	"github.com/ventros/crm/internal/domain/core/shared"
	"gorm.io/gorm"
)

// DomainEventBus publica eventos de domínio usando Transactional Outbox Pattern.
//
// **Como funciona (PostgreSQL LISTEN/NOTIFY + Outbox Pattern)**:
// 1. Salva evento no outbox (mesma transação do agregado)
// 2. Database trigger envia NOTIFY 'outbox_events' após commit
// 3. PostgresNotifyOutboxProcessor recebe notificação < 100ms (push-based!)
// 4. Processor publica no RabbitMQ + notifica webhooks
// 5. Fallback: Temporal worker processa eventos pendentes a cada 30s
//
// **Benefícios**:
// - Atomicidade: Estado + evento salvos juntos (ou ambos, ou nenhum)
// - Zero perda: Se crash após commit, evento está no banco
// - Latência baixa: < 100ms via LISTEN/NOTIFY (push, não polling!)
// - Retry automático: Temporal tenta eventos pendentes/falhados
// - Visibilidade: Temporal UI + database query mostram status
type DomainEventBus struct {
	db              *gorm.DB // Para transações
	outboxRepo      outbox.Repository
	webhookNotifier *webhooks.WebhookNotifier
	eventLogRepo    *persistence.DomainEventLogRepository
	rabbitMQ        *RabbitMQConnection // Para publicar eventos
}

// NewDomainEventBus cria um novo event bus com suporte a Transactional Outbox.
func NewDomainEventBus(
	db *gorm.DB,
	outboxRepo outbox.Repository,
	webhookNotifier *webhooks.WebhookNotifier,
	eventLogRepo *persistence.DomainEventLogRepository,
	rabbitMQ *RabbitMQConnection,
) *DomainEventBus {
	return &DomainEventBus{
		db:              db,
		outboxRepo:      outboxRepo,
		webhookNotifier: webhookNotifier,
		eventLogRepo:    eventLogRepo,
		rabbitMQ:        rabbitMQ,
	}
}

// Publish salva um evento de domínio no outbox (Transactional Outbox Pattern).
//
// **IMPORTANTE**: Este método DEVE ser chamado dentro de ExecuteInTransaction().
// Exemplo correto:
//
//	txManager.ExecuteInTransaction(ctx, func(ctx context.Context) error {
//	    if err := contactRepo.Save(ctx, contact); err != nil {
//	        return err
//	    }
//	    return eventBus.Publish(ctx, contact.DomainEvents()...)
//	})
//
// O método automaticamente detecta se há uma transação ativa no contexto e a usa.
// Se não houver transação, usa a conexão padrão (não recomendado para consistência).
//
// **Fluxo após commit**:
// 1. Serializa o evento como JSON
// 2. Salva no outbox_events table (status: pending)
// 3. Database trigger envia NOTIFY 'outbox_events' (após commit bem-sucedido)
// 4. PostgresNotifyOutboxProcessor recebe push notification < 100ms
// 5. Processor publica no RabbitMQ + notifica webhooks
// 6. Marca evento como processado
//
// **NÃO** publica diretamente no RabbitMQ para garantir atomicidade.
//
// **Saga Support**: Se o contexto contiver Saga metadata (correlation_id, saga_type, saga_step),
// esses metadados serão automaticamente anexados ao evento no outbox para rastreamento de Saga.
func (bus *DomainEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
	// Serializa evento como JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Extrai metadados do evento (tenantID, projectID)
	// TODO: Implementar extração de tenant/project do contexto ou evento
	var tenantID *string
	var projectID *uuid.UUID

	// ✅ Saga Support: Extrai Saga metadata do contexto
	var sagaMetadata map[string]interface{}
	if sagaMeta := saga.GetMetadata(ctx); sagaMeta != nil {
		sagaMetadata = map[string]interface{}{
			"correlation_id": sagaMeta.CorrelationID,
			"saga_type":      sagaMeta.SagaType,
			"saga_step":      sagaMeta.SagaStep,
			"step_number":    sagaMeta.StepNumber,
		}

		// Se há TenantID no Saga metadata, usa ele
		if sagaMeta.TenantID != "" {
			tenantID = &sagaMeta.TenantID
		}
	}

	// Cria evento de outbox
	outboxEvent := &outbox.OutboxEvent{
		ID:            uuid.New(),
		EventID:       event.EventID(), // ID único do evento de domínio
		AggregateID:   uuid.New(),      // TODO: Extrair do evento (depende de cada agregado)
		AggregateType: extractAggregateType(event.EventName()),
		EventType:     event.EventName(),
		EventVersion:  event.EventVersion(),
		EventData:     payload,
		Metadata:      sagaMetadata, // ✅ Saga metadata para correlação
		TenantID:      tenantID,
		ProjectID:     projectID,
		CreatedAt:     time.Now(),
		Status:        outbox.StatusPending,
		RetryCount:    0,
	}

	// Salva no outbox (dentro da transação atual)
	// A publicação será feita APÓS o commit via PostgreSQL LISTEN/NOTIFY
	if err := bus.outboxRepo.Save(ctx, outboxEvent); err != nil {
		return fmt.Errorf("failed to save event to outbox: %w", err)
	}

	// ✅ Transactional Outbox Pattern (SEM POLLING!):
	// 1. Evento salvo no outbox dentro da mesma transação do agregado
	// 2. Trigger NOTIFY 'outbox_events' envia notificação push ao PostgresNotifyOutboxProcessor
	// 3. Processor recebe notificação < 100ms (push-based!)
	// 4. Processor publica no RabbitMQ + notifica webhooks
	// 5. Fallback: Temporal worker processa eventos pendentes a cada 30s (polling)
	//
	// Benefícios:
	// - Atomicidade: Estado + evento salvos juntos (ou ambos, ou nenhum)
	// - Zero perda: Se crash após commit, evento está no banco
	// - Latência baixa: < 100ms (push via LISTEN/NOTIFY)
	// - Retry automático: Temporal tenta novamente se RabbitMQ falhar

	// Log de evento (não bloqueia)
	if bus.eventLogRepo != nil {
		go func() {
			if err := bus.eventLogRepo.LogEvent(context.Background(), event, "default", nil, nil); err != nil {
				fmt.Printf("⚠️  Failed to log domain event: %v\n", err)
			}
		}()
	}

	fmt.Printf("✅ Event %s saved to outbox (event_id: %s)\n", event.EventName(), event.EventID())
	return nil
}

// extractAggregateType extrai o tipo do agregado a partir do nome do evento.
// Exemplo: "contact.created" → "Contact"
func extractAggregateType(eventName string) string {
	// Formato esperado: "resource.action" (ex: "contact.created", "session.started")
	parts := splitEventName(eventName)
	if len(parts) > 0 {
		return capitalize(parts[0])
	}
	return "Unknown"
}

func splitEventName(eventName string) []string {
	result := []string{}
	current := ""
	for _, char := range eventName {
		if char == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Converte primeira letra para maiúscula
	first := s[0:1]
	rest := s[1:]
	return string(first[0]-32) + rest // ASCII: 'a' = 97, 'A' = 65, diff = 32
}

// PublishBatch publica múltiplos eventos em batch.
func (bus *DomainEventBus) PublishBatch(ctx context.Context, events []shared.DomainEvent) error {
	for _, event := range events {
		if err := bus.Publish(ctx, event); err != nil {
			// Log error but continue with other events
			fmt.Printf("Failed to publish event %s: %v\n", event.EventName(), err)
		}

		// Small delay to avoid overwhelming RabbitMQ
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// mapDomainToBusinessEvents está agora sendo usado pelo Outbox Processor
// para mapear eventos de domínio para webhooks após publicar no RabbitMQ

// mapDomainToBusinessEvents mapeia eventos de domínio para eventos de negócio (webhooks).
// Um evento de domínio pode gerar múltiplos eventos de negócio.
func (bus *DomainEventBus) mapDomainToBusinessEvents(domainEvent string) []string {
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
	case "message.played":
		return []string{"message.played"} // Voz/áudio reproduzido (SOMENTE voice messages)
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

	// Eventos de canal
	case "channel.created":
		return []string{"channel.created"}
	case "channel.activated":
		return []string{"channel.activated"}
	case "channel.deactivated":
		return []string{"channel.deactivated"}
	case "channel.deleted":
		return []string{"channel.deleted"}

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

	// Eventos de contact adicionais
	case "contact.deleted":
		return []string{"contact.deleted"}
	case "contact.profile_picture_updated":
		return []string{"contact.profile_picture_updated"}

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

	// Eventos de agent adicionais
	case "agent.logged_in":
		return []string{"agent.logged_in"}
	case "agent.permission_granted":
		return []string{"agent.permission_granted"}
	case "agent.permission_revoked":
		return []string{"agent.permission_revoked"}

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

	// Eventos de channel adicionais
	case "channel.pipeline.associated":
		return []string{"channel.pipeline_associated"}
	case "channel.pipeline.disassociated":
		return []string{"channel.pipeline_disassociated"}

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
