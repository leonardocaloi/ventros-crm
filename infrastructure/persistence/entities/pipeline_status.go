package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PipelineStatusEntity representa a entidade Status de um Pipeline no banco de dados
type PipelineStatusEntity struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PipelineID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name        string         `gorm:"not null"`
	Description string         `gorm:"type:text"`
	Color       string         `gorm:""`
	StatusType  string         `gorm:"not null;index"` // open, active, closed
	Position    int            `gorm:"default:0;index"`
	Active      bool           `gorm:"default:true;index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Pipeline PipelineEntity `gorm:"foreignKey:PipelineID"`
}

func (PipelineStatusEntity) TableName() string {
	return "pipeline_statuses"
}
