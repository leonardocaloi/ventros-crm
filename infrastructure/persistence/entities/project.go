package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectEntity represents a Project entity in the database
type ProjectEntity struct {
	ID                    uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID                uuid.UUID              `gorm:"type:uuid;not null;index:idx_projects_user"`
	BillingAccountID      uuid.UUID              `gorm:"type:uuid;not null;index:idx_projects_billing"`
	TenantID              string                 `gorm:"uniqueIndex:idx_projects_tenant_unique;not null;index:idx_projects_tenant;index:idx_projects_tenant_active,priority:1"`
	Name                  string                 `gorm:"not null;index:idx_projects_name"`
	Description           string                 `gorm:"type:text"`
	Configuration         map[string]interface{} `gorm:"type:jsonb;index:idx_projects_config,type:gin"`
	Active                bool                   `gorm:"default:true;index:idx_projects_active;index:idx_projects_tenant_active,priority:2"`
	SessionTimeoutMinutes int                    `gorm:"default:30;not null;index:idx_projects_timeout"` // Timeout padrão para todas as sessões do projeto
	CreatedAt             time.Time              `gorm:"autoCreateTime;index:idx_projects_created"`
	UpdatedAt             time.Time              `gorm:"autoUpdateTime;index:idx_projects_updated"`
	DeletedAt             gorm.DeletedAt         `gorm:"index:idx_projects_deleted"`

	// Relacionamentos
	User           UserEntity           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	BillingAccount BillingAccountEntity `gorm:"foreignKey:BillingAccountID;constraint:OnDelete:RESTRICT"`
	Contacts       []ContactEntity      `gorm:"foreignKey:ProjectID"`
	Messages       []MessageEntity      `gorm:"foreignKey:ProjectID"`
	Pipelines      []PipelineEntity     `gorm:"foreignKey:ProjectID"`
}

func (ProjectEntity) TableName() string {
	return "projects"
}
