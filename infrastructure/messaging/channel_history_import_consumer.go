package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	importpkg "github.com/ventros/crm/internal/application/channel/import"
	"github.com/ventros/crm/internal/domain/crm/channel"
)

// ChannelHistoryImportConsumer consome eventos de importação de histórico
// Pattern: Event-Driven + Strategy Pattern
// Quando recebe channel.history_import.requested:
// 1. Busca o canal no repository
// 2. Usa StrategyFactory para escolher strategy do tipo correto
// 3. Executa CanImport() + Import() (inicia Temporal Workflow)
// 4. O workflow atualiza status via events (completed/failed)
type ChannelHistoryImportConsumer struct {
	conn            *RabbitMQConnection
	channelRepo     channel.Repository
	strategyFactory *importpkg.StrategyFactory
	eventBus        ChannelEventBus
	txManager       ChannelTransactionManager
	logger          *zap.Logger
}

func NewChannelHistoryImportConsumer(
	conn *RabbitMQConnection,
	channelRepo channel.Repository,
	strategyFactory *importpkg.StrategyFactory,
	eventBus ChannelEventBus,
	txManager ChannelTransactionManager,
	logger *zap.Logger,
) *ChannelHistoryImportConsumer {
	return &ChannelHistoryImportConsumer{
		conn:            conn,
		channelRepo:     channelRepo,
		strategyFactory: strategyFactory,
		eventBus:        eventBus,
		txManager:       txManager,
		logger:          logger,
	}
}

// Start inicia o consumo de eventos de importação
func (c *ChannelHistoryImportConsumer) Start(ctx context.Context) error {
	queueName := "domain.events.channel.history_import.requested"
	consumerTag := fmt.Sprintf("channel-import-consumer-%s", uuid.New().String()[:8])

	consumer := &channelHistoryImportRequestedConsumer{parent: c}

	if err := c.conn.StartConsumer(ctx, queueName, consumerTag, consumer, 5); err != nil {
		c.logger.Error("Failed to start channel history import consumer",
			zap.String("queue", queueName),
			zap.Error(err))
		return err
	}

	c.logger.Info("Channel history import consumer started",
		zap.String("queue", queueName),
		zap.String("consumer_tag", consumerTag))

	return nil
}

// channelHistoryImportRequestedConsumer processa eventos channel.history_import.requested
type channelHistoryImportRequestedConsumer struct {
	parent *ChannelHistoryImportConsumer
}

func (c *channelHistoryImportRequestedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event channel.ChannelHistoryImportRequestedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ChannelHistoryImportRequestedEvent", zap.Error(err))
		return err
	}

	c.parent.logger.Info("Processing channel history import request",
		zap.String("channel_id", event.ChannelID.String()),
		zap.String("channel_type", string(event.ChannelType)),
		zap.String("correlation_id", event.CorrelationID),
		zap.String("strategy", event.Strategy),
		zap.Int("time_range_days", event.TimeRangeDays),
		zap.Int("limit", event.Limit))

	// 1. Buscar canal no repository
	ch, err := c.parent.channelRepo.GetByID(event.ChannelID)
	if err != nil {
		c.parent.logger.Error("Failed to get channel",
			zap.Error(err),
			zap.String("channel_id", event.ChannelID.String()),
			zap.String("correlation_id", event.CorrelationID))
		return err
	}

	// 2. Obter strategy apropriada para o tipo de canal
	strategy, err := c.parent.strategyFactory.GetStrategy(string(ch.Type))
	if err != nil {
		c.parent.logger.Error("Failed to get import strategy",
			zap.Error(err),
			zap.String("channel_type", string(ch.Type)),
			zap.String("correlation_id", event.CorrelationID))

		// Strategy não existe → marca como failed
		ch.FailHistoryImport(fmt.Sprintf("No import strategy for channel type: %s", ch.Type))
		return c.failChannelImport(ctx, ch, event.CorrelationID)
	}

	// 3. Verificar pré-condições (CanImport)
	if err := strategy.CanImport(ctx, ch, event.Strategy); err != nil {
		c.parent.logger.Warn("Channel cannot start import - pre-conditions not met",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("correlation_id", event.CorrelationID))

		ch.FailHistoryImport(fmt.Sprintf("Pre-import check failed: %v", err))
		return c.failChannelImport(ctx, ch, event.CorrelationID)
	}

	// 4. Iniciar importação (strategy.Import inicia Temporal Workflow - não bloqueia!)
	params := importpkg.ImportParams{
		Strategy:      event.Strategy,
		TimeRangeDays: event.TimeRangeDays,
		Limit:         event.Limit,
		CorrelationID: event.CorrelationID,
		UserID:        ch.UserID.String(),
	}

	workflowID, err := strategy.Import(ctx, ch, params)
	if err != nil {
		c.parent.logger.Error("Failed to start import workflow",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("correlation_id", event.CorrelationID))

		ch.FailHistoryImport(fmt.Sprintf("Failed to start import workflow: %v", err))
		return c.failChannelImport(ctx, ch, event.CorrelationID)
	}

	// 5. Sucesso! Workflow iniciado - publicar evento started
	if err := ch.StartHistoryImport(); err != nil {
		c.parent.logger.Error("Failed to mark import as started",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("correlation_id", event.CorrelationID))
		// Não é crítico - workflow já está rodando
	}

	// 6. Salvar canal + publicar eventos em transação (Outbox Pattern)
	err = c.parent.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if err := c.parent.channelRepo.Update(ch); err != nil {
			return fmt.Errorf("failed to save channel: %w", err)
		}

		// Publicar eventos de domínio (channel.history_import.started)
		events := ch.DomainEvents()
		for _, domainEvent := range events {
			if err := c.parent.eventBus.Publish(txCtx, domainEvent); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		c.parent.logger.Error("Failed to save channel import start",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("correlation_id", event.CorrelationID))
		return err
	}

	ch.ClearEvents()

	c.parent.logger.Info("Channel history import started successfully (async)",
		zap.String("channel_id", ch.ID.String()),
		zap.String("type", string(ch.Type)),
		zap.String("workflow_id", workflowID),
		zap.String("correlation_id", event.CorrelationID),
		zap.String("strategy", event.Strategy))

	return nil
}

// failChannelImport salva canal com status failed e publica evento de falha
func (c *channelHistoryImportRequestedConsumer) failChannelImport(ctx context.Context, ch *channel.Channel, correlationID string) error {
	// Salvar + publicar em transação
	err := c.parent.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if err := c.parent.channelRepo.Update(ch); err != nil {
			return fmt.Errorf("failed to save channel: %w", err)
		}

		// Publicar evento channel.history_import.failed
		events := ch.DomainEvents()
		for _, domainEvent := range events {
			if err := c.parent.eventBus.Publish(txCtx, domainEvent); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		c.parent.logger.Error("Failed to mark channel import as failed",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("correlation_id", correlationID))
		return err
	}

	ch.ClearEvents()

	c.parent.logger.Warn("Channel history import marked as failed",
		zap.String("channel_id", ch.ID.String()),
		zap.String("correlation_id", correlationID),
		zap.String("reason", ch.LastError))

	return nil // Retorna nil para ACK a mensagem (já processamos)
}
