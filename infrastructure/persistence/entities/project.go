package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectEntity represents a Project entity in the database
type ProjectEntity struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	BillingAccountID uuid.UUID      `gorm:"type:uuid;not null;index"`
	TenantID         string         `gorm:"uniqueIndex;not null;index"`
	Name             string         `gorm:"not null"`
	Description      string         `gorm:"type:text"`
	Configuration    map[string]interface{} `gorm:"type:jsonb"`
	Active           bool           `gorm:"default:true"`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

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
