package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID            uuid.UUID
	EventID       uuid.UUID
	AggregateID   uuid.UUID
	AggregateType string
	EventType     string
	EventVersion  string
	EventData     []byte
	Metadata      map[string]interface{} // Saga metadata (correlation_id, saga_type, saga_step)
	TenantID      *string
	ProjectID     *uuid.UUID
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	Status        OutboxStatus
	RetryCount    int
	LastError     *string
	LastRetryAt   *time.Time
}

type OutboxStatus string

const (
	StatusPending    OutboxStatus = "pending"
	StatusProcessing OutboxStatus = "processing"
	StatusProcessed  OutboxStatus = "processed"
	StatusFailed     OutboxStatus = "failed"
)

type Repository interface {
	Save(ctx context.Context, event *OutboxEvent) error

	GetByID(ctx context.Context, id uuid.UUID) (*OutboxEvent, error)

	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	MarkAsProcessing(ctx context.Context, eventID uuid.UUID) error

	MarkAsProcessed(ctx context.Context, eventID uuid.UUID) error

	MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error

	GetFailedEventsForRetry(ctx context.Context, maxRetries int, retryBackoff time.Duration, limit int) ([]*OutboxEvent, error)

	CountPending(ctx context.Context) (int64, error)

	CountFailed(ctx context.Context) (int64, error)

	// Saga tracking methods
	GetSagaEvents(ctx context.Context, correlationID string) ([]*OutboxEvent, error)

	GetSagaStatus(ctx context.Context, correlationID string) (string, int, int, error) // status, total, completed, failed
}
