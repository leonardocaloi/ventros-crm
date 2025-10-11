package billing

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type BillingAccountCreatedEvent struct {
	AccountID    uuid.UUID
	UserID       uuid.UUID
	Name         string
	BillingEmail string
	CreatedAt    time.Time
}

func (e BillingAccountCreatedEvent) EventType() string     { return "billing.account.created" }
func (e BillingAccountCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type PaymentMethodActivatedEvent struct {
	AccountID     uuid.UUID
	PaymentMethod PaymentMethod
	ActivatedAt   time.Time
}

func (e PaymentMethodActivatedEvent) EventType() string     { return "billing.payment.activated" }
func (e PaymentMethodActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

type BillingAccountSuspendedEvent struct {
	AccountID   uuid.UUID
	Reason      string
	SuspendedAt time.Time
}

func (e BillingAccountSuspendedEvent) EventType() string     { return "billing.account.suspended" }
func (e BillingAccountSuspendedEvent) OccurredAt() time.Time { return e.SuspendedAt }

type BillingAccountReactivatedEvent struct {
	AccountID     uuid.UUID
	ReactivatedAt time.Time
}

func (e BillingAccountReactivatedEvent) EventType() string     { return "billing.account.reactivated" }
func (e BillingAccountReactivatedEvent) OccurredAt() time.Time { return e.ReactivatedAt }

type BillingAccountCanceledEvent struct {
	AccountID  uuid.UUID
	CanceledAt time.Time
}

func (e BillingAccountCanceledEvent) EventType() string     { return "billing.account.canceled" }
func (e BillingAccountCanceledEvent) OccurredAt() time.Time { return e.CanceledAt }
