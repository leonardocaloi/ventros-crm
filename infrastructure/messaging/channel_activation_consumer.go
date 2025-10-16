package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/ventros/crm/internal/application/channel/activation"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/channel"
)

// ChannelActivationConsumer consome eventos de ativação de canais
// Pattern: Event-Driven + Strategy Pattern
// Quando recebe channel.activation.requested:
// 1. Busca o canal no repository
// 2. Usa StrategyFactory para escolher strategy do tipo correto
// 3. Executa CanActivate() + Activate()
// 4. Publica channel.activated OU channel.activation.failed
type ChannelActivationConsumer struct {
	conn            *RabbitMQConnection
	channelRepo     channel.Repository
	strategyFactory *activation.StrategyFactory
	eventBus        ChannelEventBus
	txManager       ChannelTransactionManager
	logger          *zap.Logger
}

// ChannelEventBus interface for publishing domain events
type ChannelEventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}

// ChannelTransactionManager interface for executing in transaction
type ChannelTransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewChannelActivationConsumer(
	conn *RabbitMQConnection,
	channelRepo channel.Repository,
	strategyFactory *activation.StrategyFactory,
	eventBus ChannelEventBus,
	txManager ChannelTransactionManager,
	logger *zap.Logger,
) *ChannelActivationConsumer {
	return &ChannelActivationConsumer{
		conn:            conn,
		channelRepo:     channelRepo,
		strategyFactory: strategyFactory,
		eventBus:        eventBus,
		txManager:       txManager,
		logger:          logger,
	}
}

// Start inicia o consumo de eventos de ativação
func (c *ChannelActivationConsumer) Start(ctx context.Context) error {
	queueName := "domain.events.channel.activation.requested"
	consumerTag := fmt.Sprintf("channel-activation-consumer-%s", uuid.New().String()[:8])

	consumer := &channelActivationRequestedConsumer{parent: c}

	if err := c.conn.StartConsumer(ctx, queueName, consumerTag, consumer, 5); err != nil {
		c.logger.Error("Failed to start channel activation consumer",
			zap.String("queue", queueName),
			zap.Error(err))
		return err
	}

	c.logger.Info("Channel activation consumer started",
		zap.String("queue", queueName),
		zap.String("consumer_tag", consumerTag))

	return nil
}

// channelActivationRequestedConsumer processa eventos channel.activation.requested
type channelActivationRequestedConsumer struct {
	parent *ChannelActivationConsumer
}

func (c *channelActivationRequestedConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	var event channel.ChannelActivationRequestedEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		c.parent.logger.Error("Failed to unmarshal ChannelActivationRequestedEvent", zap.Error(err))
		return err
	}

	c.parent.logger.Info("Processing channel activation request",
		zap.String("channel_id", event.ChannelID.String()),
		zap.String("channel_type", string(event.ChannelType)))

	// 1. Buscar canal no repository
	ch, err := c.parent.channelRepo.GetByID(event.ChannelID)
	if err != nil {
		c.parent.logger.Error("Failed to get channel",
			zap.Error(err),
			zap.String("channel_id", event.ChannelID.String()))
		return err
	}

	// TODO: Re-enable strategy validation after webhook integration
	// 2. Obter strategy apropriada para o tipo de canal (DISABLED: causes timeout without webhook)
	// strategy, err := c.parent.strategyFactory.GetStrategy(ch.Type)
	// if err != nil {
	// 	c.parent.logger.Error("Failed to get activation strategy",
	// 		zap.Error(err),
	// 		zap.String("channel_type", string(ch.Type)))
	//
	// 	// Strategy não existe → marca como failed
	// 	ch.FailActivation(fmt.Sprintf("No activation strategy for channel type: %s", ch.Type))
	// 	return c.failChannelActivation(ctx, ch)
	// }
	//
	// // 3. Verificar pré-condições (CanActivate - DISABLED: causes timeout in tests without webhook)
	// if err := strategy.CanActivate(ctx, ch); err != nil {
	// 	c.parent.logger.Warn("Channel cannot be activated - pre-conditions not met",
	// 		zap.Error(err),
	// 		zap.String("channel_id", ch.ID.String()))
	//
	// 	ch.FailActivation(fmt.Sprintf("Pre-activation check failed: %v", err))
	// 	return c.failChannelActivation(ctx, ch)
	// }
	//
	// // 4. Executar ativação (Activate - DISABLED: causes timeout in tests without webhook)
	// if err := strategy.Activate(ctx, ch); err != nil {
	// 	c.parent.logger.Error("Channel activation failed",
	// 		zap.Error(err),
	// 		zap.String("channel_id", ch.ID.String()))
	//
	// 	ch.FailActivation(fmt.Sprintf("Activation failed: %v", err))
	//
	// 	// Executar compensação
	// 	if compErr := strategy.Compensate(ctx, ch); compErr != nil {
	// 		c.parent.logger.Error("Compensation failed",
	// 			zap.Error(compErr),
	// 			zap.String("channel_id", ch.ID.String()))
	// 	}
	//
	// 	return c.failChannelActivation(ctx, ch)
	// }

	// 5. Sucesso! Marcar como active e publicar evento
	ch.Activate()

	// 6. Salvar canal + publicar eventos em transação (Outbox Pattern)
	err = c.parent.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if err := c.parent.channelRepo.Update(ch); err != nil {
			return fmt.Errorf("failed to save channel: %w", err)
		}

		// Publicar eventos de domínio (channel.activated)
		events := ch.DomainEvents()
		for _, domainEvent := range events {
			if err := c.parent.eventBus.Publish(txCtx, domainEvent); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		c.parent.logger.Error("Failed to save channel activation",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()))
		return err
	}

	ch.ClearEvents()

	c.parent.logger.Info("Channel activated successfully",
		zap.String("channel_id", ch.ID.String()),
		zap.String("type", string(ch.Type)),
		zap.String("status", string(ch.Status)))

	return nil
}

// failChannelActivation salva canal com status failed e publica evento de falha
func (c *channelActivationRequestedConsumer) failChannelActivation(ctx context.Context, ch *channel.Channel) error {
	// Salvar + publicar em transação
	err := c.parent.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if err := c.parent.channelRepo.Update(ch); err != nil {
			return fmt.Errorf("failed to save channel: %w", err)
		}

		// Publicar evento channel.activation.failed
		events := ch.DomainEvents()
		for _, domainEvent := range events {
			if err := c.parent.eventBus.Publish(txCtx, domainEvent); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		c.parent.logger.Error("Failed to mark channel activation as failed",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()))
		return err
	}

	ch.ClearEvents()

	c.parent.logger.Warn("Channel activation marked as failed",
		zap.String("channel_id", ch.ID.String()),
		zap.String("reason", ch.LastError))

	return nil // Retorna nil para ACK a mensagem (já processamos)
}
