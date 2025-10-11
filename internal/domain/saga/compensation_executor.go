package saga

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/outbox"
)

// CompensationExecutor executa compensações para Sagas falhadas.
type CompensationExecutor struct {
	outboxRepo outbox.Repository
}

// NewCompensationExecutor cria um novo executor de compensação.
func NewCompensationExecutor(outboxRepo outbox.Repository) *CompensationExecutor {
	return &CompensationExecutor{
		outboxRepo: outboxRepo,
	}
}

// Execute executa a compensação para uma Saga falhada.
//
// **Como funciona:**
// 1. Busca todos os eventos processados com sucesso da Saga
// 2. Determina eventos de compensação necessários (via CompensationMapping)
// 3. Dispara eventos de compensação na estratégia configurada (default: ordem reversa)
// 4. Cada agregado deve ter handlers para processar eventos de compensação
//
// **Exemplo:**
// - Saga: CreateContactWithSession
// - Steps executados: [contact.created ✅, session.started ✅, message.created ❌]
// - Compensação: [compensate.session.started, compensate.contact.created] (ordem reversa!)
func (e *CompensationExecutor) Execute(
	ctx context.Context,
	execution *SagaExecution,
	config CompensationConfig,
) error {
	fmt.Printf("🔄 Starting compensation for Saga: %s (correlation_id: %s)\n",
		execution.SagaType, execution.CorrelationID)

	// Coleta eventos que precisam ser compensados
	compensationActions := e.buildCompensationActions(execution, config)

	if len(compensationActions) == 0 {
		fmt.Printf("✅ No compensation needed for Saga: %s\n", execution.CorrelationID)
		return nil
	}

	fmt.Printf("📋 Compensation plan: %d actions to execute\n", len(compensationActions))

	// Executa compensação baseado na estratégia
	switch config.Strategy {
	case ReverseOrder:
		return e.executeInReverseOrder(ctx, compensationActions, execution)
	case AllAtOnce:
		return e.executeAllAtOnce(ctx, compensationActions, execution)
	case Selective:
		return e.executeSelective(ctx, compensationActions, execution, config)
	default:
		return fmt.Errorf("unknown compensation strategy: %s", config.Strategy)
	}
}

// buildCompensationActions constrói lista de ações de compensação.
func (e *CompensationExecutor) buildCompensationActions(
	execution *SagaExecution,
	config CompensationConfig,
) []CompensationAction {
	actions := []CompensationAction{}

	// Itera sobre eventos processados com sucesso (ordem cronológica)
	for _, event := range execution.Events {
		// Só compensa eventos processados com sucesso
		if event.Status != outbox.StatusProcessed {
			continue
		}

		// Verifica se evento precisa compensação
		compensationType := GetCompensationEvent(event.EventType)
		if compensationType == "" {
			continue // Sem compensação necessária
		}

		// Se há selector, aplica filtro
		if config.EventSelector != nil && !config.EventSelector(event.EventType) {
			continue
		}

		action := CompensationAction{
			EventType:     compensationType,
			AggregateID:   event.AggregateID.String(),
			AggregateType: event.AggregateType,
			Reason:        fmt.Sprintf("Saga %s failed", execution.SagaType),
			OriginalEvent: event.EventType,
			Metadata: map[string]interface{}{
				"correlation_id":    execution.CorrelationID,
				"saga_type":         execution.SagaType,
				"original_event_id": event.EventID.String(),
			},
		}

		actions = append(actions, action)
	}

	return actions
}

// executeInReverseOrder executa compensação na ordem reversa (LIFO).
// Esta é a estratégia mais segura e comumente usada.
func (e *CompensationExecutor) executeInReverseOrder(
	ctx context.Context,
	actions []CompensationAction,
	execution *SagaExecution,
) error {
	// Itera na ordem reversa (LIFO)
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		fmt.Printf("⏪ Compensating step %d/%d: %s (aggregate: %s)\n",
			len(actions)-i, len(actions), action.EventType, action.AggregateID)

		if err := e.executeCompensationAction(ctx, action); err != nil {
			return fmt.Errorf("compensation failed at step %d: %w", i, err)
		}
	}

	fmt.Printf("✅ Compensation completed for Saga: %s\n", execution.CorrelationID)
	return nil
}

// executeAllAtOnce executa todas as compensações em paralelo.
// Usar com cuidado - apenas quando as compensações são independentes.
func (e *CompensationExecutor) executeAllAtOnce(
	ctx context.Context,
	actions []CompensationAction,
	execution *SagaExecution,
) error {
	errChan := make(chan error, len(actions))

	// Dispara todas as compensações em goroutines
	for _, action := range actions {
		go func(act CompensationAction) {
			err := e.executeCompensationAction(ctx, act)
			errChan <- err
		}(action)
	}

	// Aguarda todas completarem
	var firstError error
	for i := 0; i < len(actions); i++ {
		if err := <-errChan; err != nil && firstError == nil {
			firstError = err
		}
	}

	if firstError != nil {
		return fmt.Errorf("compensation failed: %w", firstError)
	}

	fmt.Printf("✅ Compensation completed for Saga: %s\n", execution.CorrelationID)
	return nil
}

// executeSelective executa apenas compensações selecionadas.
func (e *CompensationExecutor) executeSelective(
	ctx context.Context,
	actions []CompensationAction,
	execution *SagaExecution,
	config CompensationConfig,
) error {
	// Filtra ações baseado no selector (já filtrado em buildCompensationActions)
	// Executa na ordem reversa
	return e.executeInReverseOrder(ctx, actions, execution)
}

// executeCompensationAction executa uma ação de compensação individual.
//
// **IMPORTANTE:** Esta implementação cria um evento de compensação no Outbox.
// Os handlers reais de compensação devem estar implementados nos agregados
// e serão invocados quando o evento for processado pelo RabbitMQ.
//
// TODO: Implementar disparo de evento de compensação via DomainEventBus
// Por enquanto, este é um placeholder que documenta o contrato.
func (e *CompensationExecutor) executeCompensationAction(
	ctx context.Context,
	action CompensationAction,
) error {
	fmt.Printf("   🔧 Executing compensation: %s for %s %s\n",
		action.EventType, action.AggregateType, action.AggregateID)

	// TODO: Implementar disparo real de evento de compensação
	// Opções:
	// 1. Criar evento de compensação no Outbox diretamente
	// 2. Chamar método de compensação no agregado (ex: contact.Compensate())
	// 3. Publicar evento de compensação via DomainEventBus
	//
	// Por enquanto, logamos a intenção:
	fmt.Printf("   ⚠️  TODO: Implement actual compensation dispatch for %s\n", action.EventType)

	return nil
}

// BuildDefaultCompensationHandler cria um handler padrão de compensação para uma Saga.
// Usa a estratégia ReverseOrder e tenta compensar todos os eventos processados.
func BuildDefaultCompensationHandler(executor *CompensationExecutor) CompensationHandler {
	return func(ctx context.Context, execution *SagaExecution) error {
		config := DefaultCompensationConfig()
		return executor.Execute(ctx, execution, config)
	}
}

// BuildSelectiveCompensationHandler cria um handler que compensa apenas eventos específicos.
func BuildSelectiveCompensationHandler(
	executor *CompensationExecutor,
	eventSelector func(eventType string) bool,
) CompensationHandler {
	return func(ctx context.Context, execution *SagaExecution) error {
		config := CompensationConfig{
			Strategy:      Selective,
			MaxRetries:    3,
			EventSelector: eventSelector,
		}
		return executor.Execute(ctx, execution, config)
	}
}
