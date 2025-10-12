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
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Version       int         `gorm:"default:1;not null"` // Optimistic locking
	ProjectID     uuid.UUID   `gorm:"type:uuid;not null;index"`
	TenantID      string      `gorm:"not null;index:idx_contacts_tenant_deleted,priority:1;index:idx_contacts_tenant_name,priority:1;index:idx_contacts_tenant_created,priority:1"`
	Name          string      `gorm:"index:idx_contacts_name;index:idx_contacts_tenant_name,priority:2"`
	Email         string      `gorm:"index:idx_contacts_email"`
	Phone         string      `gorm:"index:idx_contacts_phone"`
	ExternalID    string      `gorm:"index:idx_contacts_external_id"`
	SourceChannel string      `gorm:""`
	Language      string      `gorm:"default:'en'"`
	Timezone      string      `gorm:""`
	Tags          StringArray `gorm:"type:jsonb;index:idx_contacts_tags,type:gin"`

	// WhatsApp Profile
	ProfilePictureURL       *string    `gorm:""` // URL da foto de perfil do WhatsApp
	ProfilePictureFetchedAt *time.Time `gorm:""` // Última vez que a foto foi buscada

	FirstInteractionAt *time.Time     `gorm:""`
	LastInteractionAt  *time.Time     `gorm:""`
	CreatedAt          time.Time      `gorm:"autoCreateTime;index:idx_contacts_tenant_created,priority:2;index:idx_contacts_created"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime;index:idx_contacts_updated"`
	DeletedAt          gorm.DeletedAt `gorm:"index:idx_contacts_deleted;index:idx_contacts_tenant_deleted,priority:2"`

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
