package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactListEntity representa a entidade ContactList no banco de dados
type ContactListEntity struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Version          int            `gorm:"default:1;not null"` // Optimistic locking
	ProjectID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	TenantID         string         `gorm:"not null;index"`
	Name             string         `gorm:"not null"`
	Description      string         `gorm:"type:text"`
	LogicalOperator  string         `gorm:"not null;default:'AND'"` // AND ou OR
	IsStatic         bool           `gorm:"not null;default:false"`
	ContactCount     int            `gorm:"not null;default:0"`
	LastCalculatedAt *time.Time     `gorm:""`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	FilterRules []ContactListFilterRuleEntity `gorm:"foreignKey:ContactListID"`
}

func (ContactListEntity) TableName() string {
	return "contact_lists"
}

// ContactListFilterRuleEntity represents a filter rule in the database
// ContactListFilterRuleEntity representa uma regra de filtro no banco de dados
type ContactListFilterRuleEntity struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactListID uuid.UUID  `gorm:"type:uuid;not null;index"`
	FilterType    string     `gorm:"not null"`        // custom_field, pipeline_status, tag, event, interaction, attribute
	Operator      string     `gorm:"not null"`        // eq, ne, gt, lt, contains, etc.
	FieldKey      string     `gorm:"not null"`        // Nome do campo a ser filtrado
	FieldType     string     `gorm:""`                // Tipo do campo (apenas para custom_field)
	Value         string     `gorm:"type:text"`       // Valor serializado como JSON
	PipelineID    *uuid.UUID `gorm:"type:uuid;index"` // Apenas para pipeline_status
	CreatedAt     time.Time  `gorm:"autoCreateTime"`

	// Relacionamento
	ContactList ContactListEntity `gorm:"foreignKey:ContactListID"`
}

func (ContactListFilterRuleEntity) TableName() string {
	return "contact_list_filters"
}

// ContactListMemberEntity representa um contato em uma lista estática
type ContactListMemberEntity struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactListID uuid.UUID `gorm:"type:uuid;not null;index:idx_contact_list_member"`
	ContactID     uuid.UUID `gorm:"type:uuid;not null;index:idx_contact_list_member"`
	AddedAt       time.Time `gorm:"autoCreateTime"`
}

func (ContactListMemberEntity) TableName() string {
	return "contact_list_members"
}
