package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillingAccountEntity representa a entidade de conta de faturamento no banco de dados
type BillingAccountEntity struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name             string         `gorm:"not null"`
	PaymentStatus    string         `gorm:"not null;default:'pending';index"` // pending, active, suspended, canceled
	PaymentMethods   []byte         `gorm:"type:jsonb"` // JSON array of payment methods
	BillingEmail     string         `gorm:"not null"`
	Suspended        bool           `gorm:"default:false;index"`
	SuspendedAt      *time.Time     `gorm:""`
	SuspensionReason string         `gorm:""`
	CreatedAt        time.Time      `gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	User     UserEntity      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Projects []ProjectEntity `gorm:"foreignKey:BillingAccountID"`
}

func (BillingAccountEntity) TableName() string {
	return "billing_accounts"
}
