package channel

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/channel"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}

// TransactionManager gerencia transações de banco de dados
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// ActivateChannelHandler handler para o comando ActivateChannel
// Implementa ativação assíncrona seguindo Event-Driven Architecture
type ActivateChannelHandler struct {
	repository channel.Repository
	eventBus   EventBus
	txManager  TransactionManager
	logger     *logrus.Logger
}

// NewActivateChannelHandler cria uma nova instância do handler
func NewActivateChannelHandler(
	repository channel.Repository,
	eventBus EventBus,
	txManager TransactionManager,
	logger *logrus.Logger,
) *ActivateChannelHandler {
	return &ActivateChannelHandler{
		repository: repository,
		eventBus:   eventBus,
		txManager:  txManager,
		logger:     logger,
	}
}

// Handle executa o comando de ativação de canal
// Retorna imediatamente após mudar status para "activating" e publicar evento
// A validação real (health check, etc) acontece de forma assíncrona via worker
func (h *ActivateChannelHandler) Handle(ctx context.Context, cmd ActivateChannelCommand) error {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid ActivateChannel command")
		return err
	}

	// Get channel from repository
	ch, err := h.repository.GetByID(cmd.ChannelID)
	if err != nil {
		h.logger.WithError(err).WithField("channel_id", cmd.ChannelID).Error("Channel not found")
		return fmt.Errorf("%w: %v", ErrChannelNotFound, err)
	}

	// Verify tenant ownership
	if ch.TenantID != cmd.TenantID {
		h.logger.WithFields(logrus.Fields{
			"channel_id":      cmd.ChannelID,
			"expected_tenant": cmd.TenantID,
			"actual_tenant":   ch.TenantID,
		}).Error("Tenant mismatch")
		return fmt.Errorf("channel does not belong to tenant %s", cmd.TenantID)
	}

	// Business rules: check if channel can be activated
	if ch.Status == channel.StatusActive {
		h.logger.WithField("channel_id", cmd.ChannelID).Warn("Channel is already active")
		return ErrChannelAlreadyActive
	}

	if ch.Status == channel.StatusActivating {
		h.logger.WithField("channel_id", cmd.ChannelID).Warn("Channel is already being activated")
		return ErrChannelAlreadyActivating
	}

	// Request activation (changes status to "activating" and publishes event)
	ch.RequestActivation()

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos (Outbox Pattern)
	err = h.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Save to repository (usa transação do contexto)
		if err := h.repository.Update(ch); err != nil {
			return fmt.Errorf("failed to save channel: %w", err)
		}

		// 2. Publish domain events (usa mesma transação → Outbox)
		events := ch.DomainEvents()
		for _, event := range events {
			if err := h.eventBus.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		h.logger.WithError(err).WithField("channel_id", cmd.ChannelID).Error("Failed to activate channel")
		return fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	// Clear domain events from aggregate
	ch.ClearEvents()

	h.logger.WithFields(logrus.Fields{
		"channel_id": ch.ID,
		"tenant_id":  ch.TenantID,
		"type":       ch.Type,
		"status":     ch.Status,
	}).Info("Channel activation requested successfully (async processing)")

	return nil
}

// GetChannelStatus retorna o status atual do canal (para polling)
func (h *ActivateChannelHandler) GetChannelStatus(ctx context.Context, channelID uuid.UUID) (channel.ChannelStatus, error) {
	ch, err := h.repository.GetByID(channelID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrChannelNotFound, err)
	}
	return ch.Status, nil
}
