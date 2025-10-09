package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PipelineEntity representa a entidade Pipeline no banco de dados
type PipelineEntity struct {
	ID                    uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID             uuid.UUID `gorm:"type:uuid;not null;index"`
	TenantID              string    `gorm:"not null;index"`
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	Color       string `gorm:""`
	Position    int    `gorm:"default:0;index"`
	Active      bool   `gorm:"default:true;index"`

	// Session Timeout Override (NULL = inherit from channel or project)
	SessionTimeoutMinutes *int `gorm:"index"` // Override final do timeout (NULL = herda de channel/project)

	// AI Features
	EnableAISummary bool    `gorm:"default:false"` // Ativar resumo inteligente de sessão ao final
	AIProvider      *string `gorm:""`              // openai, anthropic, etc
	AIModel         *string `gorm:""`              // gpt-4, claude-3, etc

	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Project ProjectEntity `gorm:"foreignKey:ProjectID"`
}

func (PipelineEntity) TableName() string {
	return "pipelines"
}
