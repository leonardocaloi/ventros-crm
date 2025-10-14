package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"gorm.io/gorm"
)

// IdempotencyChecker verifica e registra eventos processados para prevenir duplicação.
// Implementa o Idempotency Pattern para sistemas event-driven.
type IdempotencyChecker struct {
	db *gorm.DB
}

// NewIdempotencyChecker cria uma nova instância do checker.
func NewIdempotencyChecker(db *gorm.DB) *IdempotencyChecker {
	return &IdempotencyChecker{db: db}
}

// IsProcessed verifica se um evento já foi processado por um consumer específico.
// Retorna true se o evento já foi processado, false caso contrário.
func (ic *IdempotencyChecker) IsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error) {
	var count int64
	err := ic.db.WithContext(ctx).
		Model(&entities.ProcessedEventEntity{}).
		Where("event_id = ? AND consumer_name = ?", eventID, consumerName).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if event is processed: %w", err)
	}

	return count > 0, nil
}

// MarkAsProcessed marca um evento como processado por um consumer.
// Usa INSERT ... ON CONFLICT DO NOTHING para evitar duplicatas (idempotente).
func (ic *IdempotencyChecker) MarkAsProcessed(
	ctx context.Context,
	eventID uuid.UUID,
	consumerName string,
	processingDurationMs *int,
) error {
	entity := &entities.ProcessedEventEntity{
		EventID:              eventID,
		ConsumerName:         consumerName,
		ProcessedAt:          time.Now(),
		ProcessingDurationMs: processingDurationMs,
	}

	// Usa ON CONFLICT DO NOTHING para garantir idempotência
	// Se já existe um registro, não faz nada (não retorna erro)
	err := ic.db.WithContext(ctx).
		Clauses(OnConflictDoNothing("uq_processed_event_consumer")).
		Create(entity).Error

	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	return nil
}

// CleanupOldRecords remove registros de eventos processados há mais de X dias.
// Útil para manutenção e evitar crescimento infinito da tabela.
func (ic *IdempotencyChecker) CleanupOldRecords(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result := ic.db.WithContext(ctx).
		Where("processed_at < ?", cutoffTime).
		Delete(&entities.ProcessedEventEntity{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old records: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetProcessingStats retorna estatísticas de processamento por consumer.
type ConsumerStats struct {
	ConsumerName        string
	TotalProcessed      int64
	AvgProcessingTimeMs float64
	LastProcessedAt     time.Time
}

// GetStats retorna estatísticas de processamento.
func (ic *IdempotencyChecker) GetStats(ctx context.Context, consumerName string, since time.Time) (*ConsumerStats, error) {
	var stats struct {
		Count         int64
		AvgDuration   float64
		LastProcessed time.Time
	}

	err := ic.db.WithContext(ctx).
		Model(&entities.ProcessedEventEntity{}).
		Select(`
			COUNT(*) as count,
			AVG(processing_duration_ms) as avg_duration,
			MAX(processed_at) as last_processed
		`).
		Where("consumer_name = ? AND processed_at >= ?", consumerName, since).
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &ConsumerStats{
		ConsumerName:        consumerName,
		TotalProcessed:      stats.Count,
		AvgProcessingTimeMs: stats.AvgDuration,
		LastProcessedAt:     stats.LastProcessed,
	}, nil
}
