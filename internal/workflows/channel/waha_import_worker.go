package channel

import (
	"context"

	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"github.com/ventros/crm/internal/domain/crm/contact"
	"github.com/ventros/crm/internal/domain/crm/message"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

// WAHAImportWorker gerencia o worker Temporal para importação WAHA
type WAHAImportWorker struct {
	client     client.Client
	worker     worker.Worker
	logger     *zap.Logger
	activities *WAHAHistoryImportActivities
}

// NewWAHAImportWorker cria um novo worker para importação WAHA
func NewWAHAImportWorker(
	temporalClient client.Client,
	wahaClient *waha.WAHAClient,
	channelRepo channel.Repository,
	contactRepo contact.Repository,
	sessionRepo session.Repository,
	messageRepo message.Repository,
	logger *zap.Logger,
) *WAHAImportWorker {
	// Criar activities
	activities := NewWAHAHistoryImportActivities(
		logger,
		wahaClient,
		channelRepo,
		contactRepo,
		sessionRepo,
		messageRepo,
	)

	// Criar worker
	w := worker.New(temporalClient, "waha-imports", worker.Options{
		MaxConcurrentActivityExecutionSize: 10,
	})

	// Registrar workflow
	w.RegisterWorkflow(WAHAHistoryImportWorkflow)

	// Registrar activities com nomes explícitos (para corresponder aos nomes no workflow)
	w.RegisterActivityWithOptions(activities.FetchWAHAChatsActivity, activity.RegisterOptions{Name: "FetchWAHAChatsActivity"})
	w.RegisterActivityWithOptions(activities.ImportChatHistoryActivity, activity.RegisterOptions{Name: "ImportChatHistoryActivity"})
	w.RegisterActivityWithOptions(activities.MarkImportCompletedActivity, activity.RegisterOptions{Name: "MarkImportCompletedActivity"})

	return &WAHAImportWorker{
		client:     temporalClient,
		worker:     w,
		logger:     logger,
		activities: activities,
	}
}

// Start inicia o worker
func (w *WAHAImportWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting WAHA import worker")

	// Start worker in background
	go func() {
		if err := w.worker.Run(worker.InterruptCh()); err != nil {
			w.logger.Error("WAHA import worker failed", zap.Error(err))
		}
	}()

	w.logger.Info("WAHA import worker started successfully")
	return nil
}

// Stop para o worker
func (w *WAHAImportWorker) Stop() {
	w.logger.Info("Stopping WAHA import worker")
	w.worker.Stop()
	w.logger.Info("WAHA import worker stopped")
}
