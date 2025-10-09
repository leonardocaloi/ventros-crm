package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

// SessionManager gerencia sessões via Temporal workflows
type SessionManager struct {
	temporalClient client.Client
}

// NewSessionManager cria um novo gerenciador de sessões
func NewSessionManager(temporalClient client.Client) *SessionManager {
	return &SessionManager{
		temporalClient: temporalClient,
	}
}

// StartSessionLifecycle inicia o workflow de ciclo de vida para uma sessão
func (sm *SessionManager) StartSessionLifecycle(ctx context.Context, sessionID uuid.UUID, contactID uuid.UUID, tenantID string, channelTypeID *int, timeoutDuration time.Duration) error {
	workflowID := fmt.Sprintf("session-lifecycle-%s", sessionID.String())

	input := SessionLifecycleWorkflowInput{
		SessionID:       sessionID,
		ContactID:       contactID,
		TenantID:        tenantID,
		ChannelTypeID:   channelTypeID,
		TimeoutDuration: timeoutDuration,
	}

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "session-management",
		// Workflow pode rodar por até 24 horas (sessões muito longas)
		WorkflowExecutionTimeout: 24 * time.Hour,
	}

	_, err := sm.temporalClient.ExecuteWorkflow(ctx, options, SessionLifecycleWorkflow, input)
	if err != nil {
		return fmt.Errorf("failed to start session lifecycle workflow: %w", err)
	}

	return nil
}

// ExtendSessionTimeout sinaliza nova atividade na sessão, estendendo o timeout
func (sm *SessionManager) ExtendSessionTimeout(ctx context.Context, sessionID uuid.UUID, newTimeoutDuration time.Duration) error {
	workflowID := fmt.Sprintf("session-lifecycle-%s", sessionID.String())

	signal := SessionActivitySignal{
		NewTimeoutDuration: newTimeoutDuration,
		LastActivityAt:     time.Now(),
	}

	err := sm.temporalClient.SignalWorkflow(ctx, workflowID, "", "session-activity", signal)
	if err != nil {
		return fmt.Errorf("failed to signal session activity: %w", err)
	}

	return nil
}

// EndSessionManually encerra uma sessão manualmente (cancela o workflow)
func (sm *SessionManager) EndSessionManually(ctx context.Context, sessionID uuid.UUID, reason string) error {
	workflowID := fmt.Sprintf("session-lifecycle-%s", sessionID.String())

	// Cancela o workflow de lifecycle
	err := sm.temporalClient.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		return fmt.Errorf("failed to cancel session workflow: %w", err)
	}

	// Executa activity para encerrar a sessão
	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("end-session-%s-%d", sessionID.String(), time.Now().Unix()),
		TaskQueue: "session-management",
	}

	input := EndSessionActivityInput{
		SessionID: sessionID,
		Reason:    reason,
	}

	_, err = sm.temporalClient.ExecuteWorkflow(ctx, options, "EndSessionActivity", input)
	if err != nil {
		return fmt.Errorf("failed to end session manually: %w", err)
	}

	return nil
}

// ScheduleSessionCleanup agenda limpeza periódica de sessões
func (sm *SessionManager) ScheduleSessionCleanup(ctx context.Context) error {
	scheduleID := "session-cleanup-schedule"

	// Cria schedule para rodar a cada 15 minutos
	schedule := client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{
					Every: 15 * time.Minute,
				},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "session-cleanup",
			Workflow:  SessionCleanupWorkflow,
			TaskQueue: "session-management",
		},
	}

	_, err := sm.temporalClient.ScheduleClient().Create(ctx, schedule)
	if err != nil {
		// Se já existe, não é erro (ignora por enquanto)
		return nil
	}

	return nil
}

// GetSessionStatus retorna o status do workflow de uma sessão
func (sm *SessionManager) GetSessionStatus(ctx context.Context, sessionID uuid.UUID) (*SessionWorkflowStatus, error) {
	workflowID := fmt.Sprintf("session-lifecycle-%s", sessionID.String())

	workflowRun := sm.temporalClient.GetWorkflow(ctx, workflowID, "")

	// Verifica se o workflow está rodando
	err := workflowRun.Get(ctx, nil)
	if err != nil {
		return &SessionWorkflowStatus{
			SessionID: sessionID,
			IsActive:  false,
			Error:     err.Error(),
		}, nil
	}

	return &SessionWorkflowStatus{
		SessionID: sessionID,
		IsActive:  true,
	}, nil
}

// SessionWorkflowStatus representa o status do workflow de uma sessão
type SessionWorkflowStatus struct {
	SessionID uuid.UUID `json:"session_id"`
	IsActive  bool      `json:"is_active"`
	Error     string    `json:"error,omitempty"`
}
