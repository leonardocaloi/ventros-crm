package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactPipelineStatusEntity representa o status de um contato em um pipeline
type ContactPipelineStatusEntity struct {
	ID         uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID  uuid.UUID              `gorm:"type:uuid;not null;index"`
	PipelineID uuid.UUID              `gorm:"type:uuid;not null;index"`
	StatusID   uuid.UUID              `gorm:"type:uuid;not null;index"`
	TenantID   string                 `gorm:"not null;index"`
	EnteredAt  time.Time              `gorm:"not null"`
	ExitedAt   *time.Time             `gorm:""`
	Duration   *int64                 `gorm:""` // Duration in seconds
	Notes      string                 `gorm:"type:text"`
	Metadata   map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt  time.Time              `gorm:"autoCreateTime"`
	UpdatedAt  time.Time              `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt         `gorm:"index"`

	// Relacionamentos
	Contact  ContactEntity        `gorm:"foreignKey:ContactID"`
	Pipeline PipelineEntity       `gorm:"foreignKey:PipelineID"`
	Status   PipelineStatusEntity `gorm:"foreignKey:StatusID"`
}

func (ContactPipelineStatusEntity) TableName() string {
	return "contact_pipeline_statuses"
}
