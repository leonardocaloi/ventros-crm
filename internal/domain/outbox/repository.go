package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// OutboxEvent representa um evento armazenado no outbox aguardando publicação.
type OutboxEvent struct {
	ID            uuid.UUID
	EventID       uuid.UUID
	AggregateID   uuid.UUID
	AggregateType string
	EventType     string
	EventVersion  string
	EventData     []byte // JSON serializado
	TenantID      *string
	ProjectID     *uuid.UUID
	CreatedAt     time.Time
	ProcessedAt   *time.Time
	Status        OutboxStatus
	RetryCount    int
	LastError     *string
	LastRetryAt   *time.Time
}

// OutboxStatus representa o status de processamento do evento.
type OutboxStatus string

const (
	StatusPending    OutboxStatus = "pending"
	StatusProcessing OutboxStatus = "processing"
	StatusProcessed  OutboxStatus = "processed"
	StatusFailed     OutboxStatus = "failed"
)

// Repository define as operações de persistência para eventos do outbox.
type Repository interface {
	// Save persiste um novo evento no outbox (DEVE ser chamado dentro de uma transação).
	Save(ctx context.Context, event *OutboxEvent) error

	// GetPendingEvents retorna eventos aguardando processamento.
	// limit: número máximo de eventos a retornar.
	GetPendingEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// MarkAsProcessing marca um evento como sendo processado (lock otimista).
	MarkAsProcessing(ctx context.Context, eventID uuid.UUID) error

	// MarkAsProcessed marca um evento como processado com sucesso.
	MarkAsProcessed(ctx context.Context, eventID uuid.UUID) error

	// MarkAsFailed marca um evento como falho e incrementa retry count.
	MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error

	// GetFailedEventsForRetry retorna eventos falhados que podem ser tentados novamente.
	// maxRetries: número máximo de retries permitidos.
	// retryBackoff: tempo mínimo desde a última tentativa.
	GetFailedEventsForRetry(ctx context.Context, maxRetries int, retryBackoff time.Duration, limit int) ([]*OutboxEvent, error)

	// CountPending retorna o número de eventos aguardando processamento.
	CountPending(ctx context.Context) (int64, error)

	// CountFailed retorna o número de eventos falhados.
	CountFailed(ctx context.Context) (int64, error)
}
