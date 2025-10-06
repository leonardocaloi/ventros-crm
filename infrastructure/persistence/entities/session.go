package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SessionEntity representa a entidade Session no banco de dados
type SessionEntity struct {
	ID                   uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID            uuid.UUID   `gorm:"type:uuid;not null;index"`
	TenantID             string      `gorm:"not null;index"`
	ChannelTypeID        *int        `gorm:"index"`
	StartedAt            time.Time   `gorm:"not null;index"`
	EndedAt              *time.Time  `gorm:"index"`
	Status               string      `gorm:"default:'active';index"`
	EndReason            *string     `gorm:""`
	TimeoutDuration      int64       `gorm:"default:1800000000000"` // nanoseconds
	LastActivityAt       time.Time   `gorm:"not null;index"`
	
	// Metrics
	MessageCount         int         `gorm:"default:0"`
	MessagesFromContact  int         `gorm:"default:0"`
	MessagesFromAgent    int         `gorm:"default:0"`
	DurationSeconds      int         `gorm:"default:0"`
	
	// Agents
	AgentIDs             []uuid.UUID `gorm:"type:jsonb"`
	AgentTransfers       int         `gorm:"default:0"`
	
	// AI/Analytics
	Summary              *string     `gorm:"type:text"`
	Sentiment            *string     `gorm:""`
	SentimentScore       *float64    `gorm:""`
	Topics               []string    `gorm:"type:jsonb"`
	NextSteps            []string    `gorm:"type:jsonb"`
	KeyEntities          datatypes.JSON `gorm:"type:jsonb"`
	
	// Business flags
	Resolved             bool        `gorm:"default:false"`
	Escalated            bool        `gorm:"default:false"`
	Converted            bool        `gorm:"default:false"`
	OutcomeTags          []string    `gorm:"type:jsonb"`
	
	CreatedAt            time.Time   `gorm:"autoCreateTime"`
	UpdatedAt            time.Time   `gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Contact      ContactEntity              `gorm:"foreignKey:ContactID"`
	Messages     []MessageEntity            `gorm:"foreignKey:SessionID"`
	CustomFields []SessionCustomFieldEntity `gorm:"foreignKey:SessionID"`
}

func (SessionEntity) TableName() string {
	return "sessions"
}
