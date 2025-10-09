package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AutomationEntity representa uma automação genérica no banco
type AutomationEntity struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AutomationType string         `gorm:"type:varchar(100);not null;index:idx_automations_type;default:'pipeline_automation'"`
	PipelineID     *uuid.UUID     `gorm:"type:uuid;index:idx_automations_pipeline"` // nullable - apenas para automações de pipeline
	TenantID       string         `gorm:"type:varchar(255);not null;index:idx_automations_tenant"`
	Name           string         `gorm:"type:varchar(255);not null"`
	Description    string         `gorm:"type:text"`
	Trigger        string         `gorm:"type:varchar(100);not null;index:idx_automations_trigger"`
	Conditions     datatypes.JSON `gorm:"type:jsonb;default:'[]'"` // []RuleCondition
	Actions        datatypes.JSON `gorm:"type:jsonb;default:'[]'"` // []RuleAction
	Priority       int            `gorm:"default:0;not null;index:idx_automations_priority"`
	Enabled        bool           `gorm:"default:true;not null;index:idx_automations_enabled"`
	CreatedAt      time.Time      `gorm:"not null"`
	UpdatedAt      time.Time      `gorm:"not null"`

	// Campos para regras agendadas (scheduled)
	Schedule      datatypes.JSON `gorm:"type:jsonb"` // ScheduledRuleConfig (nullable)
	LastExecuted  *time.Time     `gorm:"index:idx_automations_last_executed"`
	NextExecution *time.Time     `gorm:"index:idx_automations_next_execution"`

	// Relacionamentos
	Pipeline *PipelineEntity `gorm:"foreignKey:PipelineID;constraint:OnDelete:SET NULL"` // SET NULL pois agora é opcional
}

// TableName especifica o nome da tabela no banco
func (AutomationEntity) TableName() string {
	return "automations"
}

// FromDomain converte do domain model para entity
func (e *AutomationEntity) FromDomain(automation interface{}) error {
	// Implementado pelo repository usando reflection ou type assertion
	return nil
}

// ToDomain converte da entity para domain model
func (e *AutomationEntity) ToDomain() (interface{}, error) {
	// Implementado pelo repository usando reflection ou type assertion
	return nil, nil
}
