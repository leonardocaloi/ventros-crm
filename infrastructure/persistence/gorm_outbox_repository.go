package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/application/shared"
	"github.com/ventros/crm/internal/domain/core/outbox"
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
// Usa a transação do contexto se existir (para atomicidade com Save do agregado).
func (r *GormOutboxRepository) Save(ctx context.Context, event *outbox.OutboxEvent) error {
	// Marshal metadata to JSON
	var metadataJSON []byte
	if event.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	entity := &entities.OutboxEventEntity{
		EventID:       event.EventID,
		AggregateID:   event.AggregateID,
		AggregateType: event.AggregateType,
		EventType:     event.EventType,
		EventVersion:  event.EventVersion,
		EventData:     event.EventData,
		Metadata:      metadataJSON,
		TenantID:      event.TenantID,
		ProjectID:     event.ProjectID,
		CreatedAt:     event.CreatedAt,
		Status:        string(event.Status),
		RetryCount:    event.RetryCount,
	}

	// Usa a transação do contexto se existir, senão usa a conexão padrão
	db := r.getDB(ctx)
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	event.ID = entity.ID
	return nil
}

// getDB retorna a transação do contexto se existir, senão retorna a conexão padrão.
func (r *GormOutboxRepository) getDB(ctx context.Context) *gorm.DB {
	// Tenta extrair transação do contexto (usa shared.TransactionFromContext)
	if tx := shared.TransactionFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	// Se não houver transação, usa conexão padrão
	return r.db.WithContext(ctx)
}

// GetByID retorna um evento específico por ID.
func (r *GormOutboxRepository) GetByID(ctx context.Context, id uuid.UUID) (*outbox.OutboxEvent, error) {
	var entity entities.OutboxEventEntity

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&entity).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get event by ID: %w", err)
	}

	events := r.entitiesToDomain([]entities.OutboxEventEntity{entity})
	if len(events) == 0 {
		return nil, fmt.Errorf("failed to convert event entity to domain")
	}

	return events[0], nil
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

// GetSagaEvents retorna todos os eventos de uma Saga específica pelo correlation_id.
func (r *GormOutboxRepository) GetSagaEvents(ctx context.Context, correlationID string) ([]*outbox.OutboxEvent, error) {
	var entities []entities.OutboxEventEntity

	err := r.db.WithContext(ctx).
		Where("metadata->>'correlation_id' = ?", correlationID).
		Order("created_at ASC").
		Find(&entities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get saga events: %w", err)
	}

	return r.entitiesToDomain(entities), nil
}

// GetSagaStatus retorna o status agregado de uma Saga (status, total, completed, failed).
func (r *GormOutboxRepository) GetSagaStatus(ctx context.Context, correlationID string) (string, int, int, error) {
	type SagaStats struct {
		Total     int
		Processed int
		Failed    int
	}

	var stats SagaStats

	err := r.db.WithContext(ctx).
		Model(&entities.OutboxEventEntity{}).
		Select(`
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'processed') as processed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		`).
		Where("metadata->>'correlation_id' = ?", correlationID).
		Scan(&stats).Error

	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to get saga status: %w", err)
	}

	// Determinar status geral da Saga
	status := "in_progress"
	if stats.Failed > 0 {
		status = "failed"
	} else if stats.Processed == stats.Total && stats.Total > 0 {
		status = "completed"
	}

	return status, stats.Total, stats.Processed, nil
}

// entitiesToDomain converte entities para domain objects.
func (r *GormOutboxRepository) entitiesToDomain(entities []entities.OutboxEventEntity) []*outbox.OutboxEvent {
	events := make([]*outbox.OutboxEvent, len(entities))
	for i, e := range entities {
		// Unmarshal metadata from JSON
		var metadata map[string]interface{}
		if len(e.Metadata) > 0 {
			if err := json.Unmarshal(e.Metadata, &metadata); err != nil {
				// Log error but don't fail - metadata is optional
				metadata = nil
			}
		}

		events[i] = &outbox.OutboxEvent{
			ID:            e.ID,
			EventID:       e.EventID,
			AggregateID:   e.AggregateID,
			AggregateType: e.AggregateType,
			EventType:     e.EventType,
			EventVersion:  e.EventVersion,
			EventData:     e.EventData,
			Metadata:      metadata,
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
