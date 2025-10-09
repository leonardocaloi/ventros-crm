package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/outbox"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormOutboxRepository implementa outbox.Repository usando GORM.
type GormOutboxRepository struct {
	db *gorm.DB
}

// NewGormOutboxRepository cria uma nova instância do repositório.
func NewGormOutboxRepository(db *gorm.DB) *GormOutboxRepository {
	return &GormOutboxRepository{db: db}
}

// Save persiste um evento no outbox.
func (r *GormOutboxRepository) Save(ctx context.Context, event *outbox.OutboxEvent) error {
	entity := &entities.OutboxEventEntity{
		EventID:       event.EventID,
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		EventType:     event.EventType,
		EventVersion:  event.EventVersion,
		EventData:     event.EventData,
		TenantID:      event.TenantID,
		ProjectID:     event.ProjectID,
		CreatedAt:     event.CreatedAt,
		Status:        string(event.Status),
		RetryCount:    event.RetryCount,
	}

	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	event.ID = entity.ID
	return nil
}

// GetPendingEvents retorna eventos aguardando processamento.
func (r *GormOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
	var entities []entities.OutboxEventEntity

	err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}

	return r.entitiesToDomain(entities), nil
}

// MarkAsProcessing marca um evento como sendo processado (lock otimista).
func (r *GormOutboxRepository) MarkAsProcessing(ctx context.Context, eventID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&entities.OutboxEventEntity{}).
		Where("event_id = ? AND status = ?", eventID, "pending").
		Update("status", "processing")

	if result.Error != nil {
		return fmt.Errorf("failed to mark event as processing: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("event %s is not pending or does not exist", eventID)
	}

	return nil
}

// MarkAsProcessed marca um evento como processado com sucesso.
func (r *GormOutboxRepository) MarkAsProcessed(ctx context.Context, eventID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entities.OutboxEventEntity{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":       "processed",
			"processed_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark event as processed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("event %s does not exist", eventID)
	}

	return nil
}

// MarkAsFailed marca um evento como falho e incrementa retry count.
func (r *GormOutboxRepository) MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entities.OutboxEventEntity{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":        "failed",
			"retry_count":   gorm.Expr("retry_count + 1"),
			"last_error":    errorMsg,
			"last_retry_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark event as failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("event %s does not exist", eventID)
	}

	return nil
}

// GetFailedEventsForRetry retorna eventos falhados que podem ser tentados novamente.
func (r *GormOutboxRepository) GetFailedEventsForRetry(
	ctx context.Context,
	maxRetries int,
	retryBackoff time.Duration,
	limit int,
) ([]*outbox.OutboxEvent, error) {
	var entities []entities.OutboxEventEntity

	cutoffTime := time.Now().Add(-retryBackoff)

	err := r.db.WithContext(ctx).
		Where("status = ?", "failed").
		Where("retry_count < ?", maxRetries).
		Where("last_retry_at IS NULL OR last_retry_at < ?", cutoffTime).
		Order("created_at ASC").
		Limit(limit).
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get failed events for retry: %w", err)
	}

	return r.entitiesToDomain(entities), nil
}

// CountPending retorna o número de eventos aguardando processamento.
func (r *GormOutboxRepository) CountPending(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.OutboxEventEntity{}).
		Where("status = ?", "pending").
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count pending events: %w", err)
	}

	return count, nil
}

// CountFailed retorna o número de eventos falhados.
func (r *GormOutboxRepository) CountFailed(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.OutboxEventEntity{}).
		Where("status = ?", "failed").
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count failed events: %w", err)
	}

	return count, nil
}

// entitiesToDomain converte entities para domain objects.
func (r *GormOutboxRepository) entitiesToDomain(entities []entities.OutboxEventEntity) []*outbox.OutboxEvent {
	events := make([]*outbox.OutboxEvent, len(entities))
	for i, e := range entities {
		events[i] = &outbox.OutboxEvent{
			ID:            e.ID,
			EventID:       e.EventID,
			AggregateID:   e.AggregateID,
			AggregateType: e.AggregateType,
			EventType:     e.EventType,
			EventVersion:  e.EventVersion,
			EventData:     e.EventData,
			TenantID:      e.TenantID,
			ProjectID:     e.ProjectID,
			CreatedAt:     e.CreatedAt,
			ProcessedAt:   e.ProcessedAt,
			Status:        outbox.OutboxStatus(e.Status),
			RetryCount:    e.RetryCount,
			LastError:     e.LastError,
			LastRetryAt:   e.LastRetryAt,
		}
	}
	return events
}
