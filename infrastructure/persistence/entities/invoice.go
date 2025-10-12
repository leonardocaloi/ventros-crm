package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvoiceEntity representa uma invoice do Stripe no banco de dados
type InvoiceEntity struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key"`
	Version              int            `gorm:"default:1;not null"` // Optimistic locking
	BillingAccountID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	StripeInvoiceID      string         `gorm:"not null;uniqueIndex"`
	SubscriptionID       *uuid.UUID     `gorm:"type:uuid;index"`
	StripeSubscriptionID *string        `gorm:""`
	AmountDue            int64          `gorm:"not null"`
	AmountPaid           int64          `gorm:"not null;default:0"`
	AmountRemaining      int64          `gorm:"not null"`
	Currency             string         `gorm:"not null;size:3"`
	Status               string         `gorm:"not null;index"`
	HostedInvoiceURL     string         `gorm:"type:text"`
	InvoicePDF           string         `gorm:"type:text"`
	DueDate              *time.Time     `gorm:""`
	PaidAt               *time.Time     `gorm:""`
	Metadata             []byte         `gorm:"type:jsonb"`
	CreatedAt            time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt            time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt            gorm.DeletedAt `gorm:"index"`

	// Relacionamentos
	BillingAccount BillingAccountEntity `gorm:"foreignKey:BillingAccountID;constraint:OnDelete:CASCADE"`
	Subscription   *SubscriptionEntity  `gorm:"foreignKey:SubscriptionID;constraint:OnDelete:SET NULL"`
}

func (InvoiceEntity) TableName() string {
	return "invoices"
}
