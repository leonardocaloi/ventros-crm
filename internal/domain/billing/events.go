package billing

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface para todos os eventos de domínio
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

// BillingAccountCreatedEvent é disparado quando uma conta de faturamento é criada
type BillingAccountCreatedEvent struct {
	AccountID    uuid.UUID
	UserID       uuid.UUID
	Name         string
	BillingEmail string
	CreatedAt    time.Time
}

func (e BillingAccountCreatedEvent) EventType() string  { return "billing.account.created" }
func (e BillingAccountCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// PaymentMethodActivatedEvent é disparado quando um método de pagamento é ativado
type PaymentMethodActivatedEvent struct {
	AccountID     uuid.UUID
	PaymentMethod PaymentMethod
	ActivatedAt   time.Time
}

func (e PaymentMethodActivatedEvent) EventType() string    { return "billing.payment.activated" }
func (e PaymentMethodActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

// BillingAccountSuspendedEvent é disparado quando uma conta é suspensa
type BillingAccountSuspendedEvent struct {
	AccountID   uuid.UUID
	Reason      string
	SuspendedAt time.Time
}

func (e BillingAccountSuspendedEvent) EventType() string    { return "billing.account.suspended" }
func (e BillingAccountSuspendedEvent) OccurredAt() time.Time { return e.SuspendedAt }

// BillingAccountReactivatedEvent é disparado quando uma conta é reativada
type BillingAccountReactivatedEvent struct {
	AccountID     uuid.UUID
	ReactivatedAt time.Time
}

func (e BillingAccountReactivatedEvent) EventType() string    { return "billing.account.reactivated" }
func (e BillingAccountReactivatedEvent) OccurredAt() time.Time { return e.ReactivatedAt }

// BillingAccountCanceledEvent é disparado quando uma conta é cancelada
type BillingAccountCanceledEvent struct {
	AccountID  uuid.UUID
	CanceledAt time.Time
}

func (e BillingAccountCanceledEvent) EventType() string    { return "billing.account.canceled" }
func (e BillingAccountCanceledEvent) OccurredAt() time.Time { return e.CanceledAt }
