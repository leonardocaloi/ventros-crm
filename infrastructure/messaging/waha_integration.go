package messaging

import (
	"context"

	messageapp "github.com/caloi/ventros-crm/internal/application/message"
	domainchannel "github.com/caloi/ventros-crm/internal/domain/crm/channel"
	domainchat "github.com/caloi/ventros-crm/internal/domain/crm/chat"
	domaincontact "github.com/caloi/ventros-crm/internal/domain/crm/contact"
	domainmessage "github.com/caloi/ventros-crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

// WAHAIntegration encapsula toda a nova arquitetura de eventos WAHA
type WAHAIntegration struct {
	RawEventBus       *WAHARawEventBus
	RawEventProcessor *WAHARawEventProcessor
	logger            *zap.Logger
}

// NewWAHAIntegration cria uma nova integração WAHA completa
func NewWAHAIntegration(
	rabbitConn *RabbitMQConnection,
	wahaMessageService *messageapp.WAHAMessageService,
	messageRepo domainmessage.Repository,
	channelRepo domainchannel.Repository,
	contactRepo domaincontact.Repository,
	chatRepo domainchat.Repository,
	logger *zap.Logger,
) *WAHAIntegration {
	// Cria o event bus para eventos raw
	rawEventBus := NewWAHARawEventBus(rabbitConn, logger)

	// Cria o processor para eventos raw
	rawEventProcessor := NewWAHARawEventProcessor(
		logger,
		rawEventBus,
		wahaMessageService,
		messageRepo,
		channelRepo,
		contactRepo,
		chatRepo,
	)

	return &WAHAIntegration{
		RawEventBus:       rawEventBus,
		RawEventProcessor: rawEventProcessor,
		logger:            logger,
	}
}

// SetupQueues configura todas as filas necessárias
func (w *WAHAIntegration) SetupQueues() error {
	w.logger.Info("Setting up WAHA raw event queues...")
	return w.RawEventBus.SetupRawEventQueues()
}

// StartProcessors inicia todos os consumers/processors
func (w *WAHAIntegration) StartProcessors(ctx context.Context, rabbitConn *RabbitMQConnection) error {
	w.logger.Info("Starting WAHA raw event processors...")

	// Inicia o processor de eventos raw
	if err := w.RawEventProcessor.Start(ctx, rabbitConn); err != nil {
		return err
	}

	w.logger.Info("WAHA integration started successfully")
	return nil
}
