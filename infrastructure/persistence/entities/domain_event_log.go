package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DomainEventLogEntity representa o log de eventos de domínio disparados
type DomainEventLogEntity struct {
	ID            uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EventType     string                 `gorm:"not null;index"`        // contact.created, session.started, etc
	AggregateID   uuid.UUID              `gorm:"type:uuid;not null;index"` // ID da entidade que gerou o evento
	AggregateType string                 `gorm:"not null;index"`        // contact, session, message, etc
	TenantID      string                 `gorm:"not null;index"`
	ProjectID     *uuid.UUID             `gorm:"type:uuid;index"`
	UserID        *uuid.UUID             `gorm:"type:uuid;index"`
	Payload       map[string]interface{} `gorm:"type:jsonb"`            // Dados completos do evento
	OccurredAt    time.Time              `gorm:"not null;index"`        // Quando o evento ocorreu
	PublishedAt   time.Time              `gorm:"not null;index"`        // Quando foi publicado
	CreatedAt     time.Time              `gorm:"autoCreateTime"`
	DeletedAt     gorm.DeletedAt         `gorm:"index"`
}

func (DomainEventLogEntity) TableName() string {
	return "domain_event_logs"
}
