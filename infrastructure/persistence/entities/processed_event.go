package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProcessedEventEntity registra que um evento foi processado por um consumer específico.
// Implementa Idempotency Pattern para prevenir processamento duplicado.
type ProcessedEventEntity struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement"`
	EventID              uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uq_processed_event_consumer"`
	ConsumerName         string    `gorm:"type:varchar(100);not null;uniqueIndex:uq_processed_event_consumer;index:idx_processed_events_lookup"`
	ProcessedAt          time.Time `gorm:"not null;default:NOW();index:idx_processed_events_cleanup"`
	ProcessingDurationMs *int      `gorm:"type:int"` // Duração em milissegundos para métricas
}

// TableName especifica o nome da tabela no banco.
func (ProcessedEventEntity) TableName() string {
	return "processed_events"
}
