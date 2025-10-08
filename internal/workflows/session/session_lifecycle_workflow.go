package session

import (
	"time"

	"go.temporal.io/sdk/workflow"
	"github.com/google/uuid"
)

// SessionLifecycleWorkflowInput define os parâmetros do workflow
type SessionLifecycleWorkflowInput struct {
	SessionID       uuid.UUID     `json:"session_id"`
	ContactID       uuid.UUID     `json:"contact_id"`
	TenantID        string        `json:"tenant_id"`
	ChannelTypeID   *int          `json:"channel_type_id"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
}

// SessionLifecycleWorkflow gerencia o ciclo de vida completo de uma sessão
// Agenda automaticamente o encerramento por timeout e permite cancelamento
func SessionLifecycleWorkflow(ctx workflow.Context, input SessionLifecycleWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	
	logger.Info("Session lifecycle workflow started", 
		"session_id", input.SessionID.String(),
		"timeout_duration", input.TimeoutDuration.String())
	
	// Configurações do workflow
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Loop para permitir reset do timer quando há atividade
	currentTimeout := input.TimeoutDuration
	
	for {
		selector := workflow.NewSelector(ctx)
		
		// Timer para timeout automático
		timeoutTimer := workflow.NewTimer(ctx, currentTimeout)
		selector.AddFuture(timeoutTimer, func(f workflow.Future) {
			logger.Info("Session timeout reached, ending session",
				"session_id", input.SessionID.String())
			
			// Executa activity para encerrar sessão por timeout
			var result EndSessionActivityResult
			err := workflow.ExecuteActivity(ctx, "EndSessionActivity", EndSessionActivityInput{
				SessionID: input.SessionID,
				Reason:    "inactivity_timeout",
			}).Get(ctx, &result)
			
			if err != nil {
				logger.Error("Failed to end session by timeout", "error", err)
				return
			}
			
			logger.Info("Session ended successfully by timeout",
				"session_id", input.SessionID.String(),
				"events_published", result.EventsPublished)
		})
		
		// Canal para receber sinais de atividade (nova mensagem na sessão)
		activityChannel := workflow.GetSignalChannel(ctx, "session-activity")
		
		timedOut := false
		selector.AddReceive(activityChannel, func(c workflow.ReceiveChannel, more bool) {
			var signal SessionActivitySignal
			c.Receive(ctx, &signal)
			
			logger.Info("Session activity detected, resetting timeout",
				"session_id", input.SessionID.String(),
				"new_timeout", signal.NewTimeoutDuration.String())
			
			// Atualiza o timeout para o próximo loop
			currentTimeout = signal.NewTimeoutDuration
		})
		
		// Aguarda timeout ou nova atividade
		selector.Select(ctx)
		
		// Se o timer expirou, encerra o workflow
		err := timeoutTimer.Get(ctx, nil)
		if err == nil {
			timedOut = true
		}
		
		if timedOut {
			break
		}
		
		// Caso contrário, continua o loop com novo timeout
	}
	
	return nil
}

// SessionActivitySignal é enviado quando há nova atividade na sessão
type SessionActivitySignal struct {
	NewTimeoutDuration time.Duration `json:"new_timeout_duration"`
	LastActivityAt     time.Time     `json:"last_activity_at"`
}

// SessionCleanupWorkflow executa limpeza periódica de sessões órfãs
func SessionCleanupWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	
	logger.Info("Session cleanup workflow started")
	
	// Configurações da activity
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	
	// Executa limpeza de sessões órfãs
	var result CleanupSessionsActivityResult
	err := workflow.ExecuteActivity(ctx, "CleanupSessionsActivity", CleanupSessionsActivityInput{
		MaxInactivityDuration: 45 * time.Minute, // Margem de segurança
	}).Get(ctx, &result)
	
	if err != nil {
		logger.Error("Failed to cleanup sessions", "error", err)
		return err
	}
	
	logger.Info("Session cleanup completed",
		"sessions_cleaned", result.SessionsCleaned,
		"events_published", result.EventsPublished)
	
	return nil
}
