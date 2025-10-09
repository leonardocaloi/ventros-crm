package billing

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus representa o status de pagamento da conta
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // Aguardando configuração de pagamento
	PaymentStatusActive    PaymentStatus = "active"    // Pagamento configurado e verificado
	PaymentStatusSuspended PaymentStatus = "suspended" // Suspensa por falta de pagamento
	PaymentStatusCanceled  PaymentStatus = "canceled"  // Cancelada pelo usuário
)

// PaymentMethod representa o método de pagamento (fake por enquanto)
type PaymentMethod struct {
	Type       string // "credit_card", "boleto", "pix", etc
	LastDigits string // Últimos 4 dígitos do cartão, etc
	ExpiresAt  *time.Time
	IsDefault  bool
}

// BillingAccount é o Aggregate Root para contas de faturamento
type BillingAccount struct {
	id               uuid.UUID
	userID           uuid.UUID
	name             string
	paymentStatus    PaymentStatus
	paymentMethods   []PaymentMethod
	billingEmail     string
	suspended        bool
	suspendedAt      *time.Time
	suspensionReason string
	createdAt        time.Time
	updatedAt        time.Time

	events []DomainEvent
}

var (
	ErrInvalidUserID    = errors.New("user ID cannot be nil")
	ErrInvalidName      = errors.New("name cannot be empty")
	ErrInvalidEmail     = errors.New("billing email cannot be empty")
	ErrAccountSuspended = errors.New("billing account is suspended")
	ErrAccountCanceled  = errors.New("billing account is canceled")
	ErrPaymentNotActive = errors.New("payment method not active")
)

// NewBillingAccount cria uma nova conta de faturamento
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
		userID:         userID,
		name:           name,
		paymentStatus:  PaymentStatusPending,
		paymentMethods: []PaymentMethod{},
		billingEmail:   billingEmail,
		suspended:      false,
		createdAt:      now,
		updatedAt:      now,
		events:         []DomainEvent{},
	}

	account.addEvent(BillingAccountCreatedEvent{
		AccountID:    account.id,
		UserID:       userID,
		Name:         name,
		BillingEmail: billingEmail,
		CreatedAt:    now,
	})

	return account, nil
}

// ReconstructBillingAccount reconstrói uma conta de faturamento a partir de dados persistidos
func ReconstructBillingAccount(
	id uuid.UUID,
	userID uuid.UUID,
	name string,
	paymentStatus PaymentStatus,
	paymentMethods []PaymentMethod,
	billingEmail string,
	suspended bool,
	suspendedAt *time.Time,
	suspensionReason string,
	createdAt time.Time,
	updatedAt time.Time,
) *BillingAccount {
	if paymentMethods == nil {
		paymentMethods = []PaymentMethod{}
	}

	return &BillingAccount{
		id:               id,
		userID:           userID,
		name:             name,
		paymentStatus:    paymentStatus,
		paymentMethods:   paymentMethods,
		billingEmail:     billingEmail,
		suspended:        suspended,
		suspendedAt:      suspendedAt,
		suspensionReason: suspensionReason,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		events:           []DomainEvent{},
	}
}

// ActivatePayment ativa o método de pagamento (fake por enquanto)
func (b *BillingAccount) ActivatePayment(method PaymentMethod) error {
	// Check canceled first (more specific than suspended)
	if b.paymentStatus == PaymentStatusCanceled {
		return ErrAccountCanceled
	}
	if b.suspended {
		return ErrAccountSuspended
	}

	// Marca outros métodos como não-default
	for i := range b.paymentMethods {
		b.paymentMethods[i].IsDefault = false
	}

	// Adiciona novo método como default
	method.IsDefault = true
	b.paymentMethods = append(b.paymentMethods, method)
	b.paymentStatus = PaymentStatusActive
	b.updatedAt = time.Now()

	b.addEvent(PaymentMethodActivatedEvent{
		AccountID:     b.id,
		PaymentMethod: method,
		ActivatedAt:   time.Now(),
	})

	return nil
}

// Suspend suspende a conta de faturamento
func (b *BillingAccount) Suspend(reason string) {
	if !b.suspended {
		now := time.Now()
		b.suspended = true
		b.suspendedAt = &now
		b.suspensionReason = reason
		b.paymentStatus = PaymentStatusSuspended
		b.updatedAt = now

		b.addEvent(BillingAccountSuspendedEvent{
			AccountID:   b.id,
			Reason:      reason,
			SuspendedAt: now,
		})
	}
}

// Reactivate reativa a conta suspensa
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

	b.addEvent(BillingAccountReactivatedEvent{
		AccountID:     b.id,
		ReactivatedAt: time.Now(),
	})

	return nil
}

// Cancel cancela permanentemente a conta
func (b *BillingAccount) Cancel() {
	now := time.Now()
	b.paymentStatus = PaymentStatusCanceled
	b.suspended = true
	b.suspendedAt = &now
	b.suspensionReason = "Canceled by user"
	b.updatedAt = now

	b.addEvent(BillingAccountCanceledEvent{
		AccountID:  b.id,
		CanceledAt: now,
	})
}

// UpdateBillingEmail atualiza o email de faturamento
func (b *BillingAccount) UpdateBillingEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	b.billingEmail = email
	b.updatedAt = time.Now()
	return nil
}

// CanCreateProject verifica se a conta pode criar projetos
func (b *BillingAccount) CanCreateProject() bool {
	return b.paymentStatus == PaymentStatusActive && !b.suspended
}

// IsActive verifica se a conta está ativa
func (b *BillingAccount) IsActive() bool {
	return b.paymentStatus == PaymentStatusActive && !b.suspended
}

// Getters
func (b *BillingAccount) ID() uuid.UUID                { return b.id }
func (b *BillingAccount) UserID() uuid.UUID            { return b.userID }
func (b *BillingAccount) Name() string                 { return b.name }
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

// Domain Events
func (b *BillingAccount) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, b.events...)
}

func (b *BillingAccount) ClearEvents() {
	b.events = []DomainEvent{}
}

func (b *BillingAccount) addEvent(event DomainEvent) {
	b.events = append(b.events, event)
}
