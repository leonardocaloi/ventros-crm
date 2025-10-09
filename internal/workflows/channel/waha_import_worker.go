package channel

import (
	"context"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/caloi/ventros-crm/internal/domain/session"
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

	// Registrar activities
	w.RegisterActivity(activities.FetchWAHAChatsActivity)
	w.RegisterActivity(activities.ImportChatHistoryActivity)
	w.RegisterActivity(activities.MarkImportCompletedActivity)

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
