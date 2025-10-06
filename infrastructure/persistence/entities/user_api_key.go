package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserAPIKeyEntity representa uma API key de usuário no banco de dados
type UserAPIKeyEntity struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name      string         `gorm:"not null"`
	KeyHash   string         `gorm:"uniqueIndex;not null"` // Hash da API key
	Active    bool           `gorm:"default:true;index"`
	LastUsed  *time.Time     `gorm:""`
	ExpiresAt *time.Time     `gorm:""`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	User UserEntity `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (UserAPIKeyEntity) TableName() string {
	return "user_api_keys"
}
