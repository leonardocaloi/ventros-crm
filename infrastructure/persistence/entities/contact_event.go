package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactEventEntity representa a entidade ContactEvent no banco de dados
type ContactEventEntity struct {
	ID                uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID         uuid.UUID              `gorm:"type:uuid;not null;index"`
	SessionID         *uuid.UUID             `gorm:"type:uuid;index"`
	TenantID          string                 `gorm:"not null;index"`
	EventType         string                 `gorm:"not null;index"`
	Category          string                 `gorm:"not null;index"` // interaction, status_change, note, etc
	Priority          string                 `gorm:"not null;index"` // low, normal, high, urgent
	Title             *string                `gorm:""`
	Description       *string                `gorm:"type:text"`
	Payload           map[string]interface{} `gorm:"type:jsonb"`
	Metadata          map[string]interface{} `gorm:"type:jsonb"`
	Source            string                 `gorm:"not null;index"` // system, agent, integration, automation
	TriggeredBy       *uuid.UUID             `gorm:"type:uuid;index"` // Agent ID
	IntegrationSource *string                `gorm:""`
	IsRealtime        bool                   `gorm:"default:true;index"`
	Delivered         bool                   `gorm:"default:false;index"`
	DeliveredAt       *time.Time             `gorm:""`
	Read              bool                   `gorm:"default:false;index"`
	ReadAt            *time.Time             `gorm:""`
	VisibleToClient   bool                   `gorm:"default:true;index"`
	VisibleToAgent    bool                   `gorm:"default:true;index"`
	ExpiresAt         *time.Time             `gorm:"index"`
	OccurredAt        time.Time              `gorm:"not null;index"`
	CreatedAt         time.Time              `gorm:"autoCreateTime"`
	DeletedAt         gorm.DeletedAt         `gorm:"index"`

	// Relacionamentos
	Contact ContactEntity  `gorm:"foreignKey:ContactID"`
	Session *SessionEntity `gorm:"foreignKey:SessionID"`
}

func (ContactEventEntity) TableName() string {
	return "contact_events"
}
