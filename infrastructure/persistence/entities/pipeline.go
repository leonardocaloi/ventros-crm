package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PipelineEntity representa a entidade Pipeline no banco de dados
type PipelineEntity struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index:idx_pipelines_project"`
	TenantID    string    `gorm:"not null;index:idx_pipelines_tenant;index:idx_pipelines_tenant_active,priority:1;index:idx_pipelines_tenant_name,priority:1"`
	Name        string    `gorm:"not null;index:idx_pipelines_name;index:idx_pipelines_tenant_name,priority:2"`
	Description string    `gorm:"type:text"`
	Color       string    `gorm:"index:idx_pipelines_color"`
	Position    int       `gorm:"default:0;index:idx_pipelines_position"`
	Active      bool      `gorm:"default:true;index:idx_pipelines_active;index:idx_pipelines_tenant_active,priority:2"`

	// Session Timeout Override (NULL = inherit from channel or project)
	SessionTimeoutMinutes *int `gorm:"index:idx_pipelines_timeout"` // Override final do timeout (NULL = herda de channel/project)

	// AI Features
	EnableAISummary bool    `gorm:"default:false;index:idx_pipelines_ai_summary"` // Ativar resumo inteligente de sessão ao final
	AIProvider      *string `gorm:"index:idx_pipelines_ai_provider"`              // openai, anthropic, etc
	AIModel         *string `gorm:""`                                             // gpt-4, claude-3, etc

	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_pipelines_created"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_pipelines_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_pipelines_deleted"`

	// Relacionamentos
	Project ProjectEntity `gorm:"foreignKey:ProjectID"`
}

func (PipelineEntity) TableName() string {
	return "pipelines"
}
