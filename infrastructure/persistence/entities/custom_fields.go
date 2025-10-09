package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactCustomFieldEntity representa campos customizados de contatos
type ContactCustomFieldEntity struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	TenantID   string         `gorm:"not null;index"`
	FieldKey   string         `gorm:"not null;index"`
	FieldType  string         `gorm:"not null"`
	FieldValue interface{}    `gorm:"type:jsonb"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Contact ContactEntity `gorm:"foreignKey:ContactID"`
}

func (ContactCustomFieldEntity) TableName() string {
	return "contact_custom_fields"
}

// SessionCustomFieldEntity representa campos customizados de sessões
type SessionCustomFieldEntity struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	TenantID   string         `gorm:"not null;index"`
	FieldKey   string         `gorm:"not null;index"`
	FieldType  string         `gorm:"not null"`
	FieldValue interface{}    `gorm:"type:jsonb"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Session SessionEntity `gorm:"foreignKey:SessionID"`
}

func (SessionCustomFieldEntity) TableName() string {
	return "session_custom_fields"
}
