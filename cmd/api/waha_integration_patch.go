package main

import (
	"context"

	"github.com/caloi/ventros-crm/infrastructure/messaging"
	messageapp "github.com/caloi/ventros-crm/internal/application/message"
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
	logger *zap.Logger,
) (*messaging.WAHAIntegration, error) {
	
	// Cria a integração WAHA completa
	wahaIntegration := messaging.NewWAHAIntegration(
		rabbitConn,
		wahaMessageService,
		messageRepo,
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
