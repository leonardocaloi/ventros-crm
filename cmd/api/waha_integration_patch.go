package main

import (
	"context"

	"github.com/caloi/ventros-crm/infrastructure/messaging"
	messageapp "github.com/caloi/ventros-crm/internal/application/message"
	domainchannel "github.com/caloi/ventros-crm/internal/domain/channel"
	domainchat "github.com/caloi/ventros-crm/internal/domain/chat"
	domaincontact "github.com/caloi/ventros-crm/internal/domain/contact"
	domainmessage "github.com/caloi/ventros-crm/internal/domain/message"
	"go.uber.org/zap"
)

// setupWAHAIntegration configura a nova arquitetura de eventos WAHA
// Esta função deve ser chamada no main.go após a inicialização do wahaMessageService
func setupWAHAIntegration(
	ctx context.Context,
	rabbitConn *messaging.RabbitMQConnection,
	wahaMessageService *messageapp.WAHAMessageService,
	messageRepo domainmessage.Repository,
	channelRepo domainchannel.Repository,
	contactRepo domaincontact.Repository,
	chatRepo domainchat.Repository,
	logger *zap.Logger,
) (*messaging.WAHAIntegration, error) {

	// Cria a integração WAHA completa
	wahaIntegration := messaging.NewWAHAIntegration(
		rabbitConn,
		wahaMessageService,
		messageRepo,
		channelRepo,
		contactRepo,
		chatRepo,
		logger,
	)

	// Configura as filas
	if err := wahaIntegration.SetupQueues(); err != nil {
		return nil, err
	}

	// Inicia os processors
	if err := wahaIntegration.StartProcessors(ctx, rabbitConn); err != nil {
		return nil, err
	}

	logger.Info("WAHA integration setup completed successfully")
	return wahaIntegration, nil
}

// setupMessageGroupWorker configura e inicia o worker de processamento de grupos de mensagens
// Esta função deve ser chamada no main.go após a inicialização dos serviços necessários
//
// NOTA: Esta função é um exemplo de integração e não está sendo usada atualmente.
// O MessageGroupWorker será inicializado de forma diferente quando as dependências
// (MessageEnrichmentService, AIAgentService) estiverem implementadas.
/*
func setupMessageGroupWorker(
	ctx context.Context,
	debouncerService *messageapp.MessageDebouncerService,
	enrichmentService *messageapp.MessageEnrichmentService,
	aiAgentService *messageapp.AIAgentService,
	messageGroupRepo message_group.Repository,
	logger *zap.Logger,
) (*messageapp.MessageGroupWorker, error) {

	// Cria o worker
	worker := messageapp.NewMessageGroupWorker(
		logger,
		debouncerService,
		enrichmentService,
		aiAgentService,
		messageGroupRepo,
	)

	// Inicia o worker em goroutine separada
	go func() {
		if err := worker.Start(ctx); err != nil {
			logger.Error("Message group worker stopped with error", zap.Error(err))
		}
	}()

	logger.Info("Message group worker started successfully")
	return worker, nil
}
*/

// Exemplo de como integrar no main.go:
/*
// Após a linha 244 (wahaMessageService := messageapp.NewWAHAMessageService(...))
// Adicionar:

// Setup nova arquitetura WAHA (eventos raw)
wahaIntegration, err := setupWAHAIntegration(ctx, rabbitConn, wahaMessageService, logger)
if err != nil {
	logger.Fatal("Failed to setup WAHA integration", zap.Error(err))
}

// Modificar a linha 274 para usar o novo handler:
wahaHandler := handlers.NewWAHAWebhookHandler(logger, wahaIntegration.RawEventBus)
*/
