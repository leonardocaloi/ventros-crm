package billing

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusActive    PaymentStatus = "active"
	PaymentStatusSuspended PaymentStatus = "suspended"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

type PaymentMethod struct {
	Type       string
	LastDigits string
	ExpiresAt  *time.Time
	IsDefault  bool
}

type BillingAccount struct {
	id               uuid.UUID
	version          int // Optimistic locking - prevents lost updates
	userID           uuid.UUID
	name             string
	stripeCustomerID string // Stripe Customer ID (cus_xxx)
	paymentStatus    PaymentStatus
	paymentMethods   []PaymentMethod
	billingEmail     string
	suspended        bool
	suspendedAt      *time.Time
	suspensionReason string
	createdAt        time.Time
	updatedAt        time.Time

	events []shared.DomainEvent
}

var (
	ErrInvalidUserID    = errors.New("user ID cannot be nil")
	ErrInvalidName      = errors.New("name cannot be empty")
	ErrInvalidEmail     = errors.New("billing email cannot be empty")
	ErrAccountSuspended = errors.New("billing account is suspended")
	ErrAccountCanceled  = errors.New("billing account is canceled")
	ErrPaymentNotActive = errors.New("payment method not active")
)

func NewBillingAccount(userID uuid.UUID, name, billingEmail string) (*BillingAccount, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if billingEmail == "" {
		return nil, ErrInvalidEmail
	}

	now := time.Now()
	account := &BillingAccount{
		id:             uuid.New(),
		version:        1, // Start with version 1 for new aggregates
		userID:         userID,
		name:           name,
		paymentStatus:  PaymentStatusPending,
		paymentMethods: []PaymentMethod{},
		billingEmail:   billingEmail,
		suspended:      false,
		createdAt:      now,
		updatedAt:      now,
		events:         []shared.DomainEvent{},
	}

	account.addEvent(NewBillingAccountCreatedEvent(account.id, userID, name, billingEmail))

	return account, nil
}

func ReconstructBillingAccount(
	id uuid.UUID,
	version int, // Optimistic locking version
	userID uuid.UUID,
	name string,
	stripeCustomerID string,
	paymentStatus PaymentStatus,
	paymentMethods []PaymentMethod,
	billingEmail string,
	suspended bool,
	suspendedAt *time.Time,
	suspensionReason string,
	createdAt time.Time,
	updatedAt time.Time,
) *BillingAccount {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}
	if paymentMethods == nil {
		paymentMethods = []PaymentMethod{}
	}

	return &BillingAccount{
		id:               id,
		version:          version,
		userID:           userID,
		name:             name,
		stripeCustomerID: stripeCustomerID,
		paymentStatus:    paymentStatus,
		paymentMethods:   paymentMethods,
		billingEmail:     billingEmail,
		suspended:        suspended,
		suspendedAt:      suspendedAt,
		suspensionReason: suspensionReason,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		events:           []shared.DomainEvent{},
	}
}

func (b *BillingAccount) ActivatePayment(method PaymentMethod) error {

	if b.paymentStatus == PaymentStatusCanceled {
		return ErrAccountCanceled
	}
	if b.suspended {
		return ErrAccountSuspended
	}

	for i := range b.paymentMethods {
		b.paymentMethods[i].IsDefault = false
	}

	method.IsDefault = true
	b.paymentMethods = append(b.paymentMethods, method)
	b.paymentStatus = PaymentStatusActive
	b.updatedAt = time.Now()

	b.addEvent(NewPaymentMethodActivatedEvent(b.id, method))

	return nil
}

func (b *BillingAccount) Suspend(reason string) {
	if !b.suspended {
		now := time.Now()
		b.suspended = true
		b.suspendedAt = &now
		b.suspensionReason = reason
		b.paymentStatus = PaymentStatusSuspended
		b.updatedAt = now

		b.addEvent(NewBillingAccountSuspendedEvent(b.id, reason))
	}
}

func (b *BillingAccount) Reactivate() error {
	if !b.suspended {
		return nil
	}

	if len(b.paymentMethods) == 0 {
		return errors.New("cannot reactivate without payment method")
	}

	b.suspended = false
	b.suspendedAt = nil
	b.suspensionReason = ""
	b.paymentStatus = PaymentStatusActive
	b.updatedAt = time.Now()

	b.addEvent(NewBillingAccountReactivatedEvent(b.id))

	return nil
}

func (b *BillingAccount) Cancel() {
	now := time.Now()
	b.paymentStatus = PaymentStatusCanceled
	b.suspended = true
	b.suspendedAt = &now
	b.suspensionReason = "Canceled by user"
	b.updatedAt = now

	b.addEvent(NewBillingAccountCanceledEvent(b.id))
}

func (b *BillingAccount) UpdateBillingEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	b.billingEmail = email
	b.updatedAt = time.Now()
	return nil
}

func (b *BillingAccount) CanCreateProject() bool {
	return b.paymentStatus == PaymentStatusActive && !b.suspended
}

func (b *BillingAccount) IsActive() bool {
	return b.paymentStatus == PaymentStatusActive && !b.suspended
}

// SetStripeCustomerID associa um Stripe Customer ao billing account
func (b *BillingAccount) SetStripeCustomerID(customerID string) error {
	if customerID == "" {
		return errors.New("stripe customer ID cannot be empty")
	}

	b.stripeCustomerID = customerID
	b.updatedAt = time.Now()

	b.addEvent(NewStripeCustomerLinkedEvent(b.id, customerID))

	return nil
}

func (b *BillingAccount) ID() uuid.UUID                { return b.id }
func (b *BillingAccount) Version() int                 { return b.version }
func (b *BillingAccount) UserID() uuid.UUID            { return b.userID }
func (b *BillingAccount) Name() string                 { return b.name }
func (b *BillingAccount) StripeCustomerID() string     { return b.stripeCustomerID }
func (b *BillingAccount) PaymentStatus() PaymentStatus { return b.paymentStatus }
func (b *BillingAccount) PaymentMethods() []PaymentMethod {
	return append([]PaymentMethod{}, b.paymentMethods...)
}
func (b *BillingAccount) BillingEmail() string     { return b.billingEmail }
func (b *BillingAccount) IsSuspended() bool        { return b.suspended }
func (b *BillingAccount) SuspendedAt() *time.Time  { return b.suspendedAt }
func (b *BillingAccount) SuspensionReason() string { return b.suspensionReason }
func (b *BillingAccount) CreatedAt() time.Time     { return b.createdAt }
func (b *BillingAccount) UpdatedAt() time.Time     { return b.updatedAt }

func (b *BillingAccount) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, b.events...)
}

func (b *BillingAccount) ClearEvents() {
	b.events = []shared.DomainEvent{}
}

func (b *BillingAccount) addEvent(event shared.DomainEvent) {
	b.events = append(b.events, event)
}

// Compile-time check that BillingAccount implements AggregateRoot interface
var _ shared.AggregateRoot = (*BillingAccount)(nil)
