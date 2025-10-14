package saga

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/outbox"
)

// SagaCoordinator coordena execuções de Saga usando Choreography Pattern.
//
// **Design Philosophy:**
// - Lightweight: Sem DB extra, usa Outbox Events como Event Store
// - Choreography-first: Eventos coordenam o fluxo (RabbitMQ)
// - Zero overhead: Ideal para fast paths (<100ms como WAHA webhooks)
// - Compensation support: Dispara eventos de compensação em caso de falha
//
// **Como funciona:**
// 1. StartSaga() cria um correlation_id e injeta no contexto
// 2. Cada evento publicado automaticamente inclui metadata via DomainEventBus
// 3. SagaTracker reconstrói estado da Saga consultando Outbox por correlation_id
// 4. CompensateSaga() dispara eventos de compensação na ordem reversa
//
// **Exemplo de uso (Choreography - Fast Path):**
//
//	ctx := coordinator.StartSaga(ctx, ProcessInboundMessageSaga, tenantID)
//
//	// Passo 1: Criar contato
//	ctx = saga.NextStep(ctx, StepContactCreated)
//	contact := contactAggregate.Create(...)
//	eventBus.Publish(ctx, contact.DomainEvents()...) // metadata incluído automaticamente
//
//	// Passo 2: Iniciar sessão
//	ctx = saga.NextStep(ctx, StepSessionStarted)
//	session := sessionAggregate.Start(...)
//	eventBus.Publish(ctx, session.DomainEvents()...) // metadata incluído automaticamente
//
//	// Passo 3: Criar mensagem
//	ctx = saga.NextStep(ctx, StepMessageCreated)
//	message := messageAggregate.Create(...)
//	eventBus.Publish(ctx, message.DomainEvents()...) // metadata incluído automaticamente
//
// Se algum passo falhar, o coordinator pode disparar compensação:
//
//	if err := step3(); err != nil {
//	    coordinator.CompensateSaga(ctx, correlationID)
//	}
type SagaCoordinator struct {
	tracker    *SagaTracker
	outboxRepo outbox.Repository
	// Compensation handlers registrados por tipo de Saga
	compensationHandlers map[SagaType]CompensationHandler
}

// CompensationHandler define como compensar uma Saga específica.
type CompensationHandler func(ctx context.Context, execution *SagaExecution) error

// NewSagaCoordinator cria um novo coordinator.
func NewSagaCoordinator(
	tracker *SagaTracker,
	outboxRepo outbox.Repository,
) *SagaCoordinator {
	return &SagaCoordinator{
		tracker:              tracker,
		outboxRepo:           outboxRepo,
		compensationHandlers: make(map[SagaType]CompensationHandler),
	}
}

// StartSaga inicia uma nova execução de Saga e retorna contexto com metadata.
//
// **Uso:**
//
//	ctx := coordinator.StartSaga(ctx, ProcessInboundMessageSaga, tenantID)
//
// Após chamar StartSaga(), todos os eventos publicados via DomainEventBus
// automaticamente incluirão o correlation_id para rastreamento.
func (c *SagaCoordinator) StartSaga(
	ctx context.Context,
	sagaType SagaType,
	tenantID string,
) context.Context {
	// Cria novo contexto com Saga metadata
	ctx = WithSaga(ctx, string(sagaType))

	// Adiciona tenant ID ao contexto
	if tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}

	correlationID, _ := GetCorrelationID(ctx)
	fmt.Printf("🎬 Saga started: %s (correlation_id: %s)\n", sagaType, correlationID)

	return ctx
}

// GetCorrelationIDFromContext extrai o correlation_id do contexto.
// Útil para logging e debugging.
func (c *SagaCoordinator) GetCorrelationIDFromContext(ctx context.Context) string {
	correlationID, _ := GetCorrelationID(ctx)
	return correlationID
}

// TrackSaga retorna o status completo de uma Saga execution.
// Reconstrói o estado consultando Outbox Events por correlation_id.
func (c *SagaCoordinator) TrackSaga(ctx context.Context, correlationID string) (*SagaExecution, error) {
	return c.tracker.TrackSaga(ctx, correlationID)
}

// IsCompleted verifica se uma Saga foi completada com sucesso.
func (c *SagaCoordinator) IsCompleted(ctx context.Context, correlationID string) (bool, error) {
	return c.tracker.IsCompleted(ctx, correlationID)
}

// IsFailed verifica se uma Saga falhou.
func (c *SagaCoordinator) IsFailed(ctx context.Context, correlationID string) (bool, error) {
	return c.tracker.IsFailed(ctx, correlationID)
}

// CompensateSaga dispara o fluxo de compensação para uma Saga falhada.
//
// **Como funciona:**
// 1. Busca todos os eventos da Saga via correlation_id
// 2. Identifica eventos processados com sucesso
// 3. Dispara eventos de compensação na ordem reversa (LIFO)
// 4. Cada agregado deve ter handlers para seus eventos de compensação
//
// **Exemplo:**
// - Saga: CreateContactWithSession
// - Steps executados: [contact.created, session.started]
// - Step falhado: message.created
// - Compensação: [session.ended, contact.deleted] (ordem reversa!)
func (c *SagaCoordinator) CompensateSaga(ctx context.Context, correlationID string) error {
	fmt.Printf("⚠️  Starting compensation for Saga: %s\n", correlationID)

	// 1. Busca execução da Saga
	execution, err := c.tracker.TrackSaga(ctx, correlationID)
	if err != nil {
		return fmt.Errorf("failed to track saga for compensation: %w", err)
	}

	// 2. Verifica se há compensation handler registrado
	handler, exists := c.compensationHandlers[SagaType(execution.SagaType)]
	if !exists {
		return fmt.Errorf("no compensation handler registered for saga type: %s", execution.SagaType)
	}

	// 3. Executa compensation handler
	if err := handler(ctx, execution); err != nil {
		return fmt.Errorf("compensation failed: %w", err)
	}

	fmt.Printf("✅ Compensation completed for Saga: %s\n", correlationID)
	return nil
}

// RegisterCompensationHandler registra um handler de compensação para um tipo de Saga.
//
// **Exemplo:**
//
//	coordinator.RegisterCompensationHandler(
//	    ProcessInboundMessageSaga,
//	    func(ctx context.Context, execution *SagaExecution) error {
//	        // Compensar na ordem reversa
//	        for i := len(execution.Events) - 1; i >= 0; i-- {
//	            event := execution.Events[i]
//	            if event.Status != outbox.StatusProcessed {
//	                continue // Só compensa eventos processados
//	            }
//
//	            switch event.EventType {
//	            case "contact.created":
//	                // Disparar contact.deleted
//	            case "session.started":
//	                // Disparar session.ended
//	            case "message.created":
//	                // Disparar message.deleted
//	            }
//	        }
//	        return nil
//	    },
//	)
func (c *SagaCoordinator) RegisterCompensationHandler(
	sagaType SagaType,
	handler CompensationHandler,
) {
	c.compensationHandlers[sagaType] = handler
	fmt.Printf("✅ Compensation handler registered for Saga: %s\n", sagaType)
}

// GetSagaEvents retorna todos os eventos de uma Saga.
func (c *SagaCoordinator) GetSagaEvents(ctx context.Context, correlationID string) ([]*outbox.OutboxEvent, error) {
	return c.outboxRepo.GetSagaEvents(ctx, correlationID)
}

// GetFailedSteps retorna os passos que falharam em uma Saga.
func (c *SagaCoordinator) GetFailedSteps(ctx context.Context, correlationID string) ([]*outbox.OutboxEvent, error) {
	return c.tracker.GetFailedSteps(ctx, correlationID)
}

// GetExecutionTimeline retorna a timeline de execução de uma Saga.
func (c *SagaCoordinator) GetExecutionTimeline(ctx context.Context, correlationID string) ([]TimelineEntry, error) {
	return c.tracker.GetExecutionTimeline(ctx, correlationID)
}

// GenerateCorrelationID gera um novo correlation ID para uso manual.
// Normalmente não é necessário, pois StartSaga() já gera automaticamente.
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// BuildCompensationContext cria um contexto para execução de compensação.
// Útil quando é necessário disparar compensação manualmente.
func BuildCompensationContext(correlationID string, sagaType SagaType, tenantID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, correlationIDKey, correlationID)
	ctx = context.WithValue(ctx, sagaTypeKey, string(sagaType))
	if tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}
	return ctx
}
