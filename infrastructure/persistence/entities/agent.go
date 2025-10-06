package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentEntity representa a entidade Agent no banco de dados
type AgentEntity struct {
	ID          uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    string                 `gorm:"not null;index"`
	Name        string                 `gorm:"not null"`
	Email       string                 `gorm:"not null;uniqueIndex:idx_agents_tenant_email"`
	Role        string                 `gorm:"not null;index"`
	Active      bool                   `gorm:"default:true;index"`
	Permissions map[string]interface{} `gorm:"type:jsonb"`
	Settings    map[string]interface{} `gorm:"type:jsonb"`
	LastLoginAt *time.Time             `gorm:""`
	CreatedAt   time.Time              `gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt         `gorm:"index"`
}

func (AgentEntity) TableName() string {
	return "agents"
}
