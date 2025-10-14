package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/ventros/crm/internal/domain/core/outbox"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PostgresNotifyOutboxProcessor processa eventos do outbox usando LISTEN/NOTIFY
// SEM POLLING! Push-based, latência < 100ms
type PostgresNotifyOutboxProcessor struct {
	db             *gorm.DB
	outboxRepo     outbox.Repository
	eventPublisher EventPublisher
	logger         *zap.Logger
	listener       *pq.Listener
	stopChan       chan struct{}
	connStr        string // Connection string for pq.Listener
}

// NewPostgresNotifyOutboxProcessor cria um novo processor com LISTEN/NOTIFY
func NewPostgresNotifyOutboxProcessor(
	db *gorm.DB,
	outboxRepo outbox.Repository,
	eventPublisher EventPublisher,
	logger *zap.Logger,
	connStr string, // Connection string fornecida externamente
) *PostgresNotifyOutboxProcessor {
	return &PostgresNotifyOutboxProcessor{
		db:             db,
		outboxRepo:     outboxRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
		connStr:        connStr,
		stopChan:       make(chan struct{}),
	}
}

// Start inicia o listener do PostgreSQL
func (p *PostgresNotifyOutboxProcessor) Start(ctx context.Context) error {
	// Usar connection string fornecida
	connStr := p.connStr

	// Criar listener do PostgreSQL
	p.listener = pq.NewListener(
		connStr,
		10*time.Second, // minReconnectInterval
		time.Minute,    // maxReconnectInterval
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				p.logger.Error("PostgreSQL listener event", zap.Error(err))
			}
		},
	)

	// Escutar canal "outbox_events"
	if err := p.listener.Listen("outbox_events"); err != nil {
		return fmt.Errorf("failed to listen on outbox_events channel: %w", err)
	}

	p.logger.Info("✅ PostgreSQL LISTEN/NOTIFY started (push-based, no polling!)",
		zap.String("channel", "outbox_events"))

	// Processar eventos pendentes que já existem (startup)
	go p.processExistingEvents(ctx)

	// Escutar notificações em tempo real (PUSH!)
	go p.listenForNotifications(ctx)

	return nil
}

// Stop para o listener
func (p *PostgresNotifyOutboxProcessor) Stop() {
	close(p.stopChan)
	if p.listener != nil {
		p.listener.Close()
	}
	p.logger.Info("PostgreSQL LISTEN/NOTIFY stopped")
}

// listenForNotifications escuta notificações do PostgreSQL (PUSH-BASED!)
func (p *PostgresNotifyOutboxProcessor) listenForNotifications(ctx context.Context) {
	for {
		select {
		case <-p.stopChan:
			return
		case <-ctx.Done():
			return
		case notification := <-p.listener.Notify:
			if notification == nil {
				continue
			}

			p.logger.Debug("Received notification from PostgreSQL",
				zap.String("channel", notification.Channel),
				zap.String("payload", notification.Extra))

			// Processar evento imediatamente (< 100ms latência!)
			go p.processNotification(ctx, notification.Extra)

		case <-time.After(90 * time.Second):
			// Ping para manter conexão viva
			go func() {
				if err := p.listener.Ping(); err != nil {
					p.logger.Warn("Failed to ping listener", zap.Error(err))
				}
			}()
		}
	}
}

// processNotification processa um evento notificado
func (p *PostgresNotifyOutboxProcessor) processNotification(ctx context.Context, eventIDStr string) {
	// Parse UUID from string
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		p.logger.Error("Invalid event ID format",
			zap.String("event_id_str", eventIDStr),
			zap.Error(err))
		return
	}

	// Use outboxRepo to fetch the event - this handles entity conversion correctly
	event, err := p.outboxRepo.GetByID(ctx, eventID)
	if err != nil {
		// Event might have been already processed by startup scan
		p.logger.Debug("Event not found or already processed",
			zap.String("event_id", eventIDStr),
			zap.Error(err))
		return
	}

	// Only process if still pending
	if event.Status != outbox.StatusPending {
		p.logger.Debug("Event is not pending, skipping",
			zap.String("event_id", eventIDStr),
			zap.String("status", string(event.Status)))
		return
	}

	// Processar evento
	if err := p.processEvent(ctx, event); err != nil {
		p.logger.Error("Failed to process event",
			zap.String("event_id", eventIDStr),
			zap.Error(err))
	}
}

// processExistingEvents processa eventos que já existiam (startup)
func (p *PostgresNotifyOutboxProcessor) processExistingEvents(ctx context.Context) {
	p.logger.Info("Processing existing pending events...")

	events, err := p.outboxRepo.GetPendingEvents(ctx, 1000)
	if err != nil {
		p.logger.Error("Failed to get pending events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		p.logger.Info("No pending events to process")
		return
	}

	p.logger.Info("Found pending events", zap.Int("count", len(events)))

	for _, event := range events {
		if err := p.processEvent(ctx, event); err != nil {
			p.logger.Error("Failed to process existing event",
				zap.String("event_id", event.EventID.String()),
				zap.Error(err))
		}
	}

	p.logger.Info("Finished processing existing events")
}

// processEvent processa um único evento
func (p *PostgresNotifyOutboxProcessor) processEvent(ctx context.Context, event *outbox.OutboxEvent) error {
	// Marca como "processing" (lock otimista)
	if err := p.outboxRepo.MarkAsProcessing(ctx, event.EventID); err != nil {
		return fmt.Errorf("failed to mark as processing: %w", err)
	}

	// Publica no RabbitMQ
	queue := fmt.Sprintf("domain.events.%s", event.EventType)
	if err := p.eventPublisher.PublishRaw(ctx, queue, event.EventData); err != nil {
		// Marca como falho
		p.outboxRepo.MarkAsFailed(ctx, event.EventID, err.Error())
		return fmt.Errorf("failed to publish: %w", err)
	}

	// Marca como processado
	if err := p.outboxRepo.MarkAsProcessed(ctx, event.EventID); err != nil {
		p.logger.Error("Failed to mark as processed", zap.Error(err))
	}

	p.logger.Debug("Event published successfully",
		zap.String("event_id", event.EventID.String()),
		zap.String("event_type", event.EventType))

	return nil
}

// EventPublisher interface para publicar eventos no RabbitMQ
type EventPublisher interface {
	PublishRaw(ctx context.Context, queue string, eventData []byte) error
}
