package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionEntity representa uma subscription do Stripe no banco de dados
type SubscriptionEntity struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key"`
	Version              int            `gorm:"default:1;not null"` // Optimistic locking
	BillingAccountID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	StripeSubscriptionID string         `gorm:"not null;uniqueIndex"`
	StripePriceID        string         `gorm:"not null"`
	Status               string         `gorm:"not null;index"`
	CurrentPeriodStart   time.Time      `gorm:"not null"`
	CurrentPeriodEnd     time.Time      `gorm:"not null"`
	TrialStart           *time.Time     `gorm:""`
	TrialEnd             *time.Time     `gorm:""`
	CancelAt             *time.Time     `gorm:""`
	CanceledAt           *time.Time     `gorm:""`
	CancelAtPeriodEnd    bool           `gorm:"not null;default:false"`
	Metadata             []byte         `gorm:"type:jsonb"`
	CreatedAt            time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	BillingAccount BillingAccountEntity `gorm:"foreignKey:BillingAccountID;constraint:OnDelete:CASCADE"`
}

func (SubscriptionEntity) TableName() string {
	return "subscriptions"
}
