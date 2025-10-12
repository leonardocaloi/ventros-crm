package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UsageMeterEntity representa um medidor de uso do Stripe no banco de dados
type UsageMeterEntity struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key"`
	Version          int            `gorm:"default:1;not null"` // Optimistic locking
	BillingAccountID uuid.UUID      `gorm:"type:uuid;not null;index"`
	StripeCustomerID string         `gorm:"not null;index"`
	StripeMeterID    string         `gorm:"not null;index"`
	MetricName       string         `gorm:"not null;size:100"`
	EventName        string         `gorm:"not null;size:100"`
	Quantity         int64          `gorm:"not null;default:0"`
	PeriodStart      time.Time      `gorm:"not null"`
	PeriodEnd        time.Time      `gorm:"not null"`
	LastReportedAt   *time.Time     `gorm:""`
	Metadata         []byte         `gorm:"type:jsonb"`
	CreatedAt        time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	BillingAccount BillingAccountEntity `gorm:"foreignKey:BillingAccountID;constraint:OnDelete:CASCADE"`
}

func (UsageMeterEntity) TableName() string {
	return "usage_meters"
}
