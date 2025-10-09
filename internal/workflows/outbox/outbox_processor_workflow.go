package outbox

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// OutboxProcessorWorkflowInput define os parâmetros do workflow.
type OutboxProcessorWorkflowInput struct {
	BatchSize    int           `json:"batch_size"`    // Quantos eventos processar por vez
	PollInterval time.Duration `json:"poll_interval"` // Intervalo entre verificações
	MaxRetries   int           `json:"max_retries"`   // Máximo de retries por evento
	RetryBackoff time.Duration `json:"retry_backoff"` // Tempo mínimo entre retries
}

// OutboxProcessorWorkflow processa eventos do outbox continuamente.
// Este workflow roda indefinidamente processando eventos pendentes.
//
// **Vantagens do Temporal**:
// - Visibilidade: UI mostra status de processamento
// - Retry automático: Se falhar, Temporal tenta novamente
// - Scheduling: Não precisa de cron job externo
// - Métricas: Temporal expõe métricas nativas
func OutboxProcessorWorkflow(ctx workflow.Context, input OutboxProcessorWorkflowInput) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Outbox Processor Workflow started",
		"batch_size", input.BatchSize,
		"poll_interval", input.PollInterval)

	// Configurações das activities
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    1 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    30 * time.Second,
		MaximumAttempts:    3,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Loop infinito processando eventos
	for {
		// 1. Processar eventos pendentes
		var pendingResult ProcessPendingEventsResult
		err := workflow.ExecuteActivity(
			ctx,
			"ProcessPendingEventsActivity",
			ProcessPendingEventsInput{
				BatchSize: input.BatchSize,
			},
		).Get(ctx, &pendingResult)

		if err != nil {
			logger.Error("Failed to process pending events", "error", err)
		} else if pendingResult.EventsProcessed > 0 {
			logger.Info("Processed pending events",
				"count", pendingResult.EventsProcessed,
				"failed", pendingResult.EventsFailed)
		}

		// 2. Processar eventos falhados (retry)
		var retryResult ProcessFailedEventsResult
		err = workflow.ExecuteActivity(
			ctx,
			"ProcessFailedEventsActivity",
			ProcessFailedEventsInput{
				BatchSize:    input.BatchSize,
				MaxRetries:   input.MaxRetries,
				RetryBackoff: input.RetryBackoff,
			},
		).Get(ctx, &retryResult)

		if err != nil {
			logger.Error("Failed to process failed events for retry", "error", err)
		} else if retryResult.EventsRetried > 0 {
			logger.Info("Retried failed events",
				"count", retryResult.EventsRetried,
				"succeeded", retryResult.EventsSucceeded)
		}

		// 3. Aguardar antes da próxima iteração
		err = workflow.Sleep(ctx, input.PollInterval)
		if err != nil {
			logger.Error("Sleep interrupted", "error", err)
			return err
		}
	}
}

// StartOutboxProcessorWorkflow inicia o workflow de processamento do outbox.
// Retorna o workflowID para que possa ser monitorado/cancelado.
func StartOutboxProcessorWorkflow(
	ctx workflow.Context,
	batchSize int,
	pollInterval time.Duration,
	maxRetries int,
	retryBackoff time.Duration,
) (string, error) {
	workflowID := "outbox-processor-workflow"

	input := OutboxProcessorWorkflowInput{
		BatchSize:    batchSize,
		PollInterval: pollInterval,
		MaxRetries:   maxRetries,
		RetryBackoff: retryBackoff,
	}

	workflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID: workflowID,
		TaskQueue:  "outbox-processor",
	}
	ctx = workflow.WithChildOptions(ctx, workflowOptions)

	future := workflow.ExecuteChildWorkflow(ctx, OutboxProcessorWorkflow, input)

	var result interface{}
	err := future.Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return workflowID, nil
}
