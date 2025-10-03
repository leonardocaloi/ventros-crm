package entities

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactEntity representa a entidade Contact no banco de dados
type ContactEntity struct {
	ID                   uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID            uuid.UUID   `gorm:"type:uuid;not null;index"`
	TenantID             string      `gorm:"not null;index"`
	Name                 string      `gorm:""`
	Email                string      `gorm:"index"`
	Phone                string      `gorm:"index"`
	ExternalID           string      `gorm:"index"`
	SourceChannel        string      `gorm:""`
	Language             string      `gorm:"default:'en'"`
	Timezone             string      `gorm:""`
	Tags                 StringArray `gorm:"type:jsonb"`
	FirstInteractionAt   *time.Time  `gorm:""`
	LastInteractionAt    *time.Time  `gorm:""`
	CreatedAt            time.Time   `gorm:"autoCreateTime"`
	UpdatedAt            time.Time   `gorm:"autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	Project      ProjectEntity              `gorm:"foreignKey:ProjectID"`
	Sessions     []SessionEntity            `gorm:"foreignKey:ContactID"`
	Messages     []MessageEntity            `gorm:"foreignKey:ContactID"`
	CustomFields []ContactCustomFieldEntity `gorm:"foreignKey:ContactID"`
}

// StringArray é um tipo customizado para serializar []string como JSON
type StringArray []string

// Value implementa driver.Valuer para GORM
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implementa sql.Scanner para GORM
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StringArray", value)
	}
	
	return json.Unmarshal(bytes, s)
}

func (ContactEntity) TableName() string {
	return "contacts"
}
