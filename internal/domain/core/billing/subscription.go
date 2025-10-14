package billing

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// SubscriptionStatus representa o status de uma subscription no Stripe
type SubscriptionStatus string

const (
	SubscriptionStatusTrialing          SubscriptionStatus = "trialing"           // Trial period
	SubscriptionStatusActive            SubscriptionStatus = "active"             // Active subscription
	SubscriptionStatusIncomplete        SubscriptionStatus = "incomplete"         // Payment pending
	SubscriptionStatusIncompleteExpired SubscriptionStatus = "incomplete_expired" // Payment failed
	SubscriptionStatusPastDue           SubscriptionStatus = "past_due"           // Payment failed, retrying
	SubscriptionStatusCanceled          SubscriptionStatus = "canceled"           // Canceled
	SubscriptionStatusUnpaid            SubscriptionStatus = "unpaid"             // Multiple payment failures
	SubscriptionStatusPaused            SubscriptionStatus = "paused"             // Paused (new in 2025)
)

// Subscription representa uma subscription do Stripe (domínio)
type Subscription struct {
	id                   uuid.UUID
	version              int // Optimistic locking - prevents lost updates
	billingAccountID     uuid.UUID
	stripeSubscriptionID string // Stripe subscription ID (sub_xxx)
	stripePriceID        string // Stripe price ID (price_xxx)
	status               SubscriptionStatus
	currentPeriodStart   time.Time
	currentPeriodEnd     time.Time
	trialStart           *time.Time
	trialEnd             *time.Time
	cancelAt             *time.Time
	canceledAt           *time.Time
	cancelAtPeriodEnd    bool
	metadata             map[string]string
	createdAt            time.Time
	updatedAt            time.Time

	events []shared.DomainEvent
}

var (
	ErrInvalidBillingAccountID     = errors.New("billing account ID cannot be nil")
	ErrInvalidStripeSubscriptionID = errors.New("stripe subscription ID cannot be empty")
	ErrInvalidStripePriceID        = errors.New("stripe price ID cannot be empty")
	ErrSubscriptionCanceled        = errors.New("subscription is canceled")
	ErrSubscriptionIncomplete      = errors.New("subscription payment is incomplete")
)

// NewSubscription cria uma nova subscription
func NewSubscription(
	billingAccountID uuid.UUID,
	stripeSubscriptionID string,
	stripePriceID string,
	status SubscriptionStatus,
	currentPeriodStart time.Time,
	currentPeriodEnd time.Time,
) (*Subscription, error) {
	if billingAccountID == uuid.Nil {
		return nil, ErrInvalidBillingAccountID
	}
	if stripeSubscriptionID == "" {
		return nil, ErrInvalidStripeSubscriptionID
	}
	if stripePriceID == "" {
		return nil, ErrInvalidStripePriceID
	}

	now := time.Now()
	subscription := &Subscription{
		id:                   uuid.New(),
		version:              1, // Start with version 1 for new aggregates
		billingAccountID:     billingAccountID,
		stripeSubscriptionID: stripeSubscriptionID,
		stripePriceID:        stripePriceID,
		status:               status,
		currentPeriodStart:   currentPeriodStart,
		currentPeriodEnd:     currentPeriodEnd,
		cancelAtPeriodEnd:    false,
		metadata:             make(map[string]string),
		createdAt:            now,
		updatedAt:            now,
		events:               []shared.DomainEvent{},
	}

	subscription.addEvent(NewSubscriptionCreatedEvent(
		subscription.id,
		billingAccountID,
		stripeSubscriptionID,
		stripePriceID,
		string(status),
	))

	return subscription, nil
}

// ReconstructSubscription reconstrói uma subscription do banco
func ReconstructSubscription(
	id uuid.UUID,
	version int, // Optimistic locking version
	billingAccountID uuid.UUID,
	stripeSubscriptionID string,
	stripePriceID string,
	status SubscriptionStatus,
	currentPeriodStart time.Time,
	currentPeriodEnd time.Time,
	trialStart *time.Time,
	trialEnd *time.Time,
	cancelAt *time.Time,
	canceledAt *time.Time,
	cancelAtPeriodEnd bool,
	metadata map[string]string,
	createdAt time.Time,
	updatedAt time.Time,
) *Subscription {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &Subscription{
		id:                   id,
		version:              version,
		billingAccountID:     billingAccountID,
		stripeSubscriptionID: stripeSubscriptionID,
		stripePriceID:        stripePriceID,
		status:               status,
		currentPeriodStart:   currentPeriodStart,
		currentPeriodEnd:     currentPeriodEnd,
		trialStart:           trialStart,
		trialEnd:             trialEnd,
		cancelAt:             cancelAt,
		canceledAt:           canceledAt,
		cancelAtPeriodEnd:    cancelAtPeriodEnd,
		metadata:             metadata,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
		events:               []shared.DomainEvent{},
	}
}

// UpdateStatus atualiza o status da subscription
func (s *Subscription) UpdateStatus(newStatus SubscriptionStatus) {
	if s.status != newStatus {
		oldStatus := s.status
		s.status = newStatus
		s.updatedAt = time.Now()

		s.addEvent(NewSubscriptionStatusChangedEvent(s.id, string(oldStatus), string(newStatus)))
	}
}

// UpdatePeriod atualiza o período de cobrança
func (s *Subscription) UpdatePeriod(start, end time.Time) {
	s.currentPeriodStart = start
	s.currentPeriodEnd = end
	s.updatedAt = time.Now()

	s.addEvent(NewSubscriptionPeriodUpdatedEvent(s.id, start, end))
}

// UpdatePrice atualiza o preço da subscription (upgrade/downgrade)
func (s *Subscription) UpdatePrice(newPriceID string) error {
	if newPriceID == "" {
		return ErrInvalidStripePriceID
	}

	oldPriceID := s.stripePriceID
	s.stripePriceID = newPriceID
	s.updatedAt = time.Now()

	s.addEvent(NewSubscriptionPriceChangedEvent(s.id, oldPriceID, newPriceID))

	return nil
}

// StartTrial inicia um trial
func (s *Subscription) StartTrial(trialEnd time.Time) {
	now := time.Now()
	s.trialStart = &now
	s.trialEnd = &trialEnd
	s.status = SubscriptionStatusTrialing
	s.updatedAt = now

	s.addEvent(NewSubscriptionTrialStartedEvent(s.id, trialEnd))
}

// ScheduleCancelAtPeriodEnd agenda cancelamento no final do período
func (s *Subscription) ScheduleCancelAtPeriodEnd() {
	s.cancelAtPeriodEnd = true
	s.cancelAt = &s.currentPeriodEnd
	s.updatedAt = time.Now()

	s.addEvent(NewSubscriptionCancelScheduledEvent(s.id, s.currentPeriodEnd))
}

// CancelImmediately cancela a subscription imediatamente
func (s *Subscription) CancelImmediately() {
	now := time.Now()
	s.status = SubscriptionStatusCanceled
	s.canceledAt = &now
	s.cancelAt = &now
	s.updatedAt = now

	s.addEvent(NewSubscriptionCanceledEvent(s.id))
}

// Reactivate reativa uma subscription cancelada
func (s *Subscription) Reactivate() error {
	if s.status != SubscriptionStatusCanceled {
		return nil // Já está ativa
	}

	s.status = SubscriptionStatusActive
	s.cancelAt = nil
	s.canceledAt = nil
	s.cancelAtPeriodEnd = false
	s.updatedAt = time.Now()

	s.addEvent(NewSubscriptionReactivatedEvent(s.id))

	return nil
}

// IsActive verifica se a subscription está ativa
func (s *Subscription) IsActive() bool {
	return s.status == SubscriptionStatusActive || s.status == SubscriptionStatusTrialing
}

// IsCanceled verifica se a subscription está cancelada
func (s *Subscription) IsCanceled() bool {
	return s.status == SubscriptionStatusCanceled
}

// IsTrial verifica se está em trial
func (s *Subscription) IsTrial() bool {
	return s.status == SubscriptionStatusTrialing
}

// DaysUntilRenewal retorna dias até renovação
func (s *Subscription) DaysUntilRenewal() int {
	duration := time.Until(s.currentPeriodEnd)
	return int(duration.Hours() / 24)
}

// SetMetadata define metadados customizados
func (s *Subscription) SetMetadata(key, value string) {
	if s.metadata == nil {
		s.metadata = make(map[string]string)
	}
	s.metadata[key] = value
	s.updatedAt = time.Now()
}

// Getters
func (s *Subscription) ID() uuid.UUID                 { return s.id }
func (s *Subscription) Version() int                  { return s.version }
func (s *Subscription) BillingAccountID() uuid.UUID   { return s.billingAccountID }
func (s *Subscription) StripeSubscriptionID() string  { return s.stripeSubscriptionID }
func (s *Subscription) StripePriceID() string         { return s.stripePriceID }
func (s *Subscription) Status() SubscriptionStatus    { return s.status }
func (s *Subscription) CurrentPeriodStart() time.Time { return s.currentPeriodStart }
func (s *Subscription) CurrentPeriodEnd() time.Time   { return s.currentPeriodEnd }
func (s *Subscription) TrialStart() *time.Time        { return s.trialStart }
func (s *Subscription) TrialEnd() *time.Time          { return s.trialEnd }
func (s *Subscription) CancelAt() *time.Time          { return s.cancelAt }
func (s *Subscription) CanceledAt() *time.Time        { return s.canceledAt }
func (s *Subscription) CancelAtPeriodEnd() bool       { return s.cancelAtPeriodEnd }
func (s *Subscription) Metadata() map[string]string   { return s.metadata }
func (s *Subscription) CreatedAt() time.Time          { return s.createdAt }
func (s *Subscription) UpdatedAt() time.Time          { return s.updatedAt }

func (s *Subscription) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, s.events...)
}

func (s *Subscription) ClearEvents() {
	s.events = []shared.DomainEvent{}
}

func (s *Subscription) addEvent(event shared.DomainEvent) {
	s.events = append(s.events, event)
}

// Compile-time check that Subscription implements AggregateRoot interface
var _ shared.AggregateRoot = (*Subscription)(nil)
