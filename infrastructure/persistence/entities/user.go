package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserEntity representa a entidade User no banco de dados (usuário autenticado)
type UserEntity struct {
	ID           uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name         string                 `gorm:"not null"`
	Email        string                 `gorm:"uniqueIndex;not null"`
	PasswordHash string                 `gorm:"not null"`
	Status       string                 `gorm:"default:'active'"`
	Role         string                 `gorm:"default:'user';check:role IN ('admin','user','manager','readonly')"`
	Settings     map[string]interface{} `gorm:"type:jsonb"`
	CreatedAt    time.Time              `gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt         `gorm:"index"`

	// Relacionamentos
	Projects []ProjectEntity    `gorm:"foreignKey:UserID"`
	APIKeys  []UserAPIKeyEntity `gorm:"foreignKey:UserID"`
}

func (UserEntity) TableName() string {
	return "users"
}
