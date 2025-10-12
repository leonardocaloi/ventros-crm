package entities

import (
	"time"

	"github.com/google/uuid"
)

// ProjectMemberEntity representa um membro de projeto no banco de dados
type ProjectMemberEntity struct {
	// Primary fields
	ID      uuid.UUID `gorm:"type:uuid;primary_key"`
	Version int       `gorm:"default:1;not null"` // Optimistic locking

	// Core fields
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index:idx_project_members_project"`
	AgentID   string    `gorm:"type:varchar(255);not null;index:idx_project_members_agent"`
	Role      string    `gorm:"type:varchar(50);not null"` // admin, supervisor, agent, viewer

	// Audit fields
	InvitedBy string    `gorm:"type:varchar(255);not null"`
	InvitedAt time.Time `gorm:"not null"`

	// Timestamps
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"` // Soft delete

	// Índices compostos
	// UNIQUE (project_id, agent_id) - Um agent só pode ter um role por projeto
}

// TableName especifica o nome da tabela
func (ProjectMemberEntity) TableName() string {
	return "project_members"
}
