package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentSessionEntity representa a participação de um agente em uma sessão (Many-to-Many)
type AgentSessionEntity struct {
	ID            uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AgentID       uuid.UUID              `gorm:"type:uuid;not null;index"`
	SessionID     uuid.UUID              `gorm:"type:uuid;not null;index"`
	RoleInSession *string                `gorm:""` // primary, support, observer, etc
	JoinedAt      time.Time              `gorm:"not null;index"`
	LeftAt        *time.Time             `gorm:"index"`
	IsActive      bool                   `gorm:"default:true;index"`
	Metadata      map[string]interface{} `gorm:"type:jsonb"` // Para integração ADK
	CreatedAt     time.Time              `gorm:"autoCreateTime"`
	UpdatedAt     time.Time              `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt         `gorm:"index"`

	// Relacionamentos
	Agent   AgentEntity   `gorm:"foreignKey:AgentID"`
	Session SessionEntity `gorm:"foreignKey:SessionID"`
}

func (AgentSessionEntity) TableName() string {
	return "agent_sessions"
}
