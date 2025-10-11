package workflow

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	sessionworkflow "github.com/caloi/ventros-crm/internal/workflows/session"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

// SessionWorker gerencia o worker Temporal para workflows de sessão
type SessionWorker struct {
	client      client.Client
	worker      worker.Worker
	sessionRepo session.Repository
	messageRepo sessionworkflow.MessageRepository
	eventBus    EventBus
	logger      *zap.Logger
}

// EventBus interface para publicar eventos
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
	PublishBatch(ctx context.Context, events []shared.DomainEvent) error
}

// NewSessionWorker cria um novo worker para sessões
func NewSessionWorker(
	temporalClient client.Client,
	sessionRepo session.Repository,
	messageRepo sessionworkflow.MessageRepository,
	eventBus EventBus,
	logger *zap.Logger,
) *SessionWorker {
	// Cria worker para task queue de sessões
	w := worker.New(temporalClient, "session-management", worker.Options{})

	return &SessionWorker{
		client:      temporalClient,
		worker:      w,
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		eventBus:    eventBus,
		logger:      logger,
	}
}

// Start inicia o worker Temporal
func (sw *SessionWorker) Start(ctx context.Context) error {
	// Registra workflows
	sw.worker.RegisterWorkflow(sessionworkflow.SessionLifecycleWorkflow)
	sw.worker.RegisterWorkflow(sessionworkflow.SessionCleanupWorkflow)

	// Cria e registra activities com nomes explícitos
	activities := sessionworkflow.NewSessionActivities(sw.sessionRepo, sw.messageRepo, sw.eventBus)
	sw.worker.RegisterActivityWithOptions(activities.EndSessionActivity, activity.RegisterOptions{Name: "EndSessionActivity"})
	sw.worker.RegisterActivityWithOptions(activities.CleanupSessionsActivity, activity.RegisterOptions{Name: "CleanupSessionsActivity"})

	sw.logger.Info("Starting session worker",
		zap.String("task_queue", "session-management"))

	// Inicia o worker
	err := sw.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start session worker: %w", err)
	}

	sw.logger.Info("Session worker started successfully")
	return nil
}

// Stop para o worker
func (sw *SessionWorker) Stop() {
	sw.logger.Info("Stopping session worker")
	sw.worker.Stop()
}

// ScheduleCleanup agenda a limpeza periódica de sessões
func (sw *SessionWorker) ScheduleCleanup(ctx context.Context) error {
	sessionManager := sessionworkflow.NewSessionManager(sw.client)

	err := sessionManager.ScheduleSessionCleanup(ctx)
	if err != nil {
		sw.logger.Error("Failed to schedule session cleanup", zap.Error(err))
		return err
	}

	sw.logger.Info("Session cleanup scheduled successfully")
	return nil
}

// GetWorkerStatus retorna informações sobre o worker
func (sw *SessionWorker) GetWorkerStatus() WorkerStatus {
	return WorkerStatus{
		TaskQueue: "session-management",
		IsRunning: sw.worker != nil,
	}
}

// WorkerStatus representa o status do worker
type WorkerStatus struct {
	TaskQueue string `json:"task_queue"`
	IsRunning bool   `json:"is_running"`
}
