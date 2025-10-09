package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OutboxEventEntity representa um evento de domínio armazenado no outbox
// antes de ser publicado no RabbitMQ (Transactional Outbox Pattern).
type OutboxEventEntity struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EventID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex"` // ID único do evento de domínio
	AggregateID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_outbox_aggregate"`
	AggregateType string         `gorm:"type:varchar(100);not null;index:idx_outbox_aggregate"`
	EventType     string         `gorm:"type:varchar(100);not null;index:idx_outbox_event_type"`
	EventVersion  string         `gorm:"type:varchar(20);not null;default:'v1'"`
	EventData     []byte         `gorm:"type:jsonb;not null"` // JSON serializado do evento
	TenantID      *string        `gorm:"type:varchar(100);index:idx_outbox_tenant"`
	ProjectID     *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt     time.Time      `gorm:"not null;default:NOW()"`
	ProcessedAt   *time.Time     `gorm:"index:idx_outbox_status_created"`
	Status        string         `gorm:"type:varchar(20);not null;default:'pending';index:idx_outbox_status_created"`
	RetryCount    int            `gorm:"not null;default:0;index:idx_outbox_retry"`
	LastError     *string        `gorm:"type:text"`
	LastRetryAt   *time.Time     `gorm:"index:idx_outbox_retry"`
	DeletedAt     gorm.DeletedAt `gorm:"index"` // Soft delete para histórico
}

// TableName especifica o nome da tabela no banco.
func (OutboxEventEntity) TableName() string {
	return "outbox_events"
}

// Constantes de status
const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusProcessed  = "processed"
	OutboxStatusFailed     = "failed"
)

// IsPending verifica se o evento está aguardando processamento.
func (e *OutboxEventEntity) IsPending() bool {
	return e.Status == OutboxStatusPending
}

// IsProcessing verifica se o evento está sendo processado.
func (e *OutboxEventEntity) IsProcessing() bool {
	return e.Status == OutboxStatusProcessing
}

// IsProcessed verifica se o evento foi processado com sucesso.
func (e *OutboxEventEntity) IsProcessed() bool {
	return e.Status == OutboxStatusProcessed
}

// IsFailed verifica se o evento falhou após retries.
func (e *OutboxEventEntity) IsFailed() bool {
	return e.Status == OutboxStatusFailed
}

// CanRetry verifica se o evento pode ser tentado novamente.
func (e *OutboxEventEntity) CanRetry(maxRetries int) bool {
	return e.IsFailed() && e.RetryCount < maxRetries
}
