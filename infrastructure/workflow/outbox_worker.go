package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/outbox"
	outboxworkflow "github.com/caloi/ventros-crm/internal/workflows/outbox"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

// OutboxWorker gerencia o worker Temporal para processamento do outbox
type OutboxWorker struct {
	client          client.Client
	worker          worker.Worker
	outboxRepo      outbox.Repository
	eventPublisher  outboxworkflow.EventPublisher
	webhookNotifier outboxworkflow.WebhookNotifier
	logger          *zap.Logger
}

// NewOutboxWorker cria um novo worker para processamento do outbox
func NewOutboxWorker(
	temporalClient client.Client,
	outboxRepo outbox.Repository,
	eventPublisher outboxworkflow.EventPublisher,
	webhookNotifier outboxworkflow.WebhookNotifier,
	logger *zap.Logger,
) *OutboxWorker {
	// Cria worker para task queue do outbox
	w := worker.New(temporalClient, "outbox-processor", worker.Options{})

	return &OutboxWorker{
		client:          temporalClient,
		worker:          w,
		outboxRepo:      outboxRepo,
		eventPublisher:  eventPublisher,
		webhookNotifier: webhookNotifier,
		logger:          logger,
	}
}

// Start inicia o worker Temporal
func (ow *OutboxWorker) Start(ctx context.Context) error {
	// Registra workflows
	ow.worker.RegisterWorkflow(outboxworkflow.OutboxProcessorWorkflow)

	// Cria e registra activities
	activities := outboxworkflow.NewOutboxActivities(ow.outboxRepo, ow.eventPublisher, ow.webhookNotifier)
	for _, activity := range activities.RegisterActivities() {
		ow.worker.RegisterActivity(activity)
	}

	ow.logger.Info("Starting outbox worker",
		zap.String("task_queue", "outbox-processor"))

	// Inicia o worker
	err := ow.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start outbox worker: %w", err)
	}

	ow.logger.Info("✅ Outbox worker started successfully (Transactional Outbox Pattern enabled)")
	return nil
}

// Stop para o worker
func (ow *OutboxWorker) Stop() {
	ow.logger.Info("Stopping outbox worker")
	ow.worker.Stop()
}

// StartProcessorWorkflow inicia o workflow de processamento contínuo do outbox
func (ow *OutboxWorker) StartProcessorWorkflow(ctx context.Context) error {
	workflowOptions := client.StartWorkflowOptions{
		ID:        "outbox-processor-workflow",
		TaskQueue: "outbox-processor",
		// Workflow roda indefinidamente, sem timeout
		WorkflowExecutionTimeout: 0,
	}

	input := outboxworkflow.OutboxProcessorWorkflowInput{
		BatchSize:    100,              // Processar 100 eventos por vez
		PollInterval: 30 * time.Second, // Verificar a cada 30 segundos (reduz ruído nos logs)
		MaxRetries:   5,                // Máximo 5 retries por evento
		RetryBackoff: 30 * time.Second, // Aguardar 30s antes de tentar novamente
	}

	workflowRun, err := ow.client.ExecuteWorkflow(ctx, workflowOptions, outboxworkflow.OutboxProcessorWorkflow, input)
	if err != nil {
		return fmt.Errorf("failed to start outbox processor workflow: %w", err)
	}

	ow.logger.Info("✅ Outbox Processor Workflow started successfully",
		zap.String("workflow_id", workflowRun.GetID()),
		zap.String("run_id", workflowRun.GetRunID()),
		zap.Int("batch_size", input.BatchSize),
		zap.Duration("poll_interval", input.PollInterval),
		zap.Int("max_retries", input.MaxRetries))

	return nil
}

// GetWorkerStatus retorna informações sobre o worker
func (ow *OutboxWorker) GetWorkerStatus() WorkerStatus {
	return WorkerStatus{
		TaskQueue: "outbox-processor",
		IsRunning: ow.worker != nil,
	}
}
