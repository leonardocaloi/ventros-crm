package saga

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ProcessInboundMessageSaga é o workflow Temporal que orquestra o processamento de mensagens inbound.
//
// Este workflow implementa Saga Orchestration Pattern com compensação automática.
//
// **Fluxo:**
// 1. FindOrCreateContact (compensate: DeleteContact se criado)
// 2. FindOrCreateSession (compensate: CloseSession se criado)
// 3. CreateMessage (compensate: DeleteMessage)
// 4. PublishDomainEvents (sem compensação - idempotência resolve)
//
// **Compensação**: Em caso de falha, executa compensação em ordem REVERSA (LIFO).
//
// **Retry**: Temporal gerencia retry automático (3x, backoff exponencial).
//
// **Visibilidade**: Temporal UI mostra status em tempo real.
func ProcessInboundMessageSaga(ctx workflow.Context, input ProcessInboundMessageInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProcessInboundMessageSaga started", "channel_message_id", input.ChannelMessageID)

	// Configuração de activities
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Estado da Saga
	sagaState := SagaState{
		CorrelationID: workflow.GetInfo(ctx).WorkflowExecution.ID,
		StartedAt:     workflow.Now(ctx),
	}

	// Defer para compensação em caso de erro
	var workflowErr error
	defer func() {
		if workflowErr != nil {
			logger.Error("Saga failed, executing compensation", "error", workflowErr.Error())
			compensateErr := compensateProcessInboundMessage(ctx, sagaState)
			if compensateErr != nil {
				logger.Error("Compensation failed", "error", compensateErr.Error())
			}
		}
	}()

	// ===== Step 1: FindOrCreate Contact =====
	logger.Info("Step 1: FindOrCreate Contact")
	var contactResult ContactCreatedResult
	err := workflow.ExecuteActivity(ctx, "FindOrCreateContactActivity", input).Get(ctx, &contactResult)
	if err != nil {
		workflowErr = fmt.Errorf("step 1 failed [contact]: %w", err)
		return workflowErr
	}
	sagaState.ContactID = contactResult.ContactID
	sagaState.ContactCreated = contactResult.WasCreated
	logger.Info("Step 1 completed", "contact_id", contactResult.ContactID, "was_created", contactResult.WasCreated)

	// ===== Step 2: FindOrCreate Session =====
	logger.Info("Step 2: FindOrCreate Session")
	var sessionResult SessionCreatedResult
	err = workflow.ExecuteActivity(ctx, "FindOrCreateSessionActivity", contactResult.ContactID, input).Get(ctx, &sessionResult)
	if err != nil {
		workflowErr = fmt.Errorf("step 2 failed [session]: %w", err)
		return workflowErr
	}
	sagaState.SessionID = sessionResult.SessionID
	sagaState.SessionCreated = sessionResult.WasCreated
	logger.Info("Step 2 completed", "session_id", sessionResult.SessionID, "was_created", sessionResult.WasCreated)

	// ===== Step 3: Create Message =====
	logger.Info("Step 3: Create Message")
	var messageResult MessageCreatedResult
	err = workflow.ExecuteActivity(ctx, "CreateMessageActivity", contactResult.ContactID, sessionResult.SessionID, input).Get(ctx, &messageResult)
	if err != nil {
		workflowErr = fmt.Errorf("step 3 failed [message]: %w", err)
		return workflowErr
	}
	sagaState.MessageID = messageResult.MessageID
	logger.Info("Step 3 completed", "message_id", messageResult.MessageID)

	// ===== Step 4: Publish Domain Events =====
	logger.Info("Step 4: Publish Domain Events")
	err = workflow.ExecuteActivity(ctx, "PublishDomainEventsActivity", sagaState).Get(ctx, nil)
	if err != nil {
		workflowErr = fmt.Errorf("step 4 failed [events]: %w", err)
		return workflowErr
	}
	logger.Info("Step 4 completed")

	// ===== Optional: Process Debouncer =====
	if !input.FromMe {
		logger.Info("Step 5: Process Message Debouncer (optional)")
		// Best effort - não falha a saga
		_ = workflow.ExecuteActivity(ctx, "ProcessMessageDebouncerActivity", messageResult.MessageID, input.ChannelID, sessionResult.SessionID).Get(ctx, nil)
	}

	// ===== Optional: Track Ad Conversion =====
	if isFromAd, ok := input.Metadata["is_from_ad"].(bool); ok && isFromAd {
		logger.Info("Step 6: Track Ad Conversion (optional)")
		// Best effort - não falha a saga
		_ = workflow.ExecuteActivity(ctx, "TrackAdConversionActivity", sagaState, input.TrackingData).Get(ctx, nil)
	}

	// Saga concluída com sucesso
	completedAt := workflow.Now(ctx)
	sagaState.CompletedAt = &completedAt

	logger.Info("ProcessInboundMessageSaga completed successfully",
		"contact_id", sagaState.ContactID,
		"session_id", sagaState.SessionID,
		"message_id", sagaState.MessageID,
		"duration_ms", completedAt.Sub(sagaState.StartedAt).Milliseconds())

	return nil
}

// compensateProcessInboundMessage executa compensação em ordem REVERSA (LIFO).
func compensateProcessInboundMessage(ctx workflow.Context, state SagaState) error {
	logger := workflow.GetLogger(ctx)
	logger.Warn("Starting compensation", "correlation_id", state.CorrelationID)

	// Configuração de activities de compensação (sem retry - best effort)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // Best effort, não insiste
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Compensação em ordem REVERSA (LIFO)

	// 4. Events já publicados → Não tem compensação (idempotência resolve duplicatas)

	// 3. Delete Message (se criado)
	if state.MessageID != uuid.Nil {
		logger.Info("Compensating: Delete Message", "message_id", state.MessageID)
		err := workflow.ExecuteActivity(ctx, "DeleteMessageActivity", state.MessageID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to compensate message", "error", err.Error())
		}
	}

	// 2. Close Session (se foi criado nesta saga)
	if state.SessionCreated && state.SessionID != uuid.Nil {
		logger.Info("Compensating: Close Session", "session_id", state.SessionID)
		err := workflow.ExecuteActivity(ctx, "CloseSessionActivity", state.SessionID, "saga_rollback").Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to compensate session", "error", err.Error())
		}
	}

	// 1. Delete Contact (se foi criado nesta saga)
	if state.ContactCreated && state.ContactID != uuid.Nil {
		logger.Info("Compensating: Delete Contact", "contact_id", state.ContactID)
		err := workflow.ExecuteActivity(ctx, "DeleteContactActivity", state.ContactID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to compensate contact", "error", err.Error())
		}
	}

	logger.Warn("Compensation completed")
	return nil
}
