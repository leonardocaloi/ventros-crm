package billing

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

// UsageMeter representa um medidor de uso (Stripe Billing Meter)
type UsageMeter struct {
	id               uuid.UUID
	version          int // Optimistic locking - prevents lost updates
	billingAccountID uuid.UUID
	stripeCustomerID string // Stripe customer ID
	stripeMeterID    string // Stripe meter ID (mtr_xxx)
	metricName       string // Ex: "messages_sent", "ai_tokens", "contacts"
	eventName        string // Ex: "message.sent", "ai.token.used"
	quantity         int64  // Current usage (running total)
	periodStart      time.Time
	periodEnd        time.Time
	lastReportedAt   *time.Time
	metadata         map[string]string
	createdAt        time.Time
	updatedAt        time.Time

	events []shared.DomainEvent
}

var (
	ErrInvalidStripeCustomerID = errors.New("stripe customer ID cannot be empty")
	ErrInvalidStripeMeterID    = errors.New("stripe meter ID cannot be empty")
	ErrInvalidMetricName       = errors.New("metric name cannot be empty")
	ErrInvalidEventName        = errors.New("event name cannot be empty")
	ErrNegativeQuantity        = errors.New("quantity cannot be negative")
)

// NewUsageMeter cria um novo medidor de uso
func NewUsageMeter(
	billingAccountID uuid.UUID,
	stripeCustomerID string,
	stripeMeterID string,
	metricName string,
	eventName string,
	periodStart time.Time,
	periodEnd time.Time,
) (*UsageMeter, error) {
	if billingAccountID == uuid.Nil {
		return nil, ErrInvalidBillingAccountID
	}
	if stripeCustomerID == "" {
		return nil, ErrInvalidStripeCustomerID
	}
	if stripeMeterID == "" {
		return nil, ErrInvalidStripeMeterID
	}
	if metricName == "" {
		return nil, ErrInvalidMetricName
	}
	if eventName == "" {
		return nil, ErrInvalidEventName
	}

	now := time.Now()
	meter := &UsageMeter{
		id:               uuid.New(),
		version:          1, // Start with version 1 for new aggregates
		billingAccountID: billingAccountID,
		stripeCustomerID: stripeCustomerID,
		stripeMeterID:    stripeMeterID,
		metricName:       metricName,
		eventName:        eventName,
		quantity:         0,
		periodStart:      periodStart,
		periodEnd:        periodEnd,
		metadata:         make(map[string]string),
		createdAt:        now,
		updatedAt:        now,
		events:           []shared.DomainEvent{},
	}

	meter.addEvent(NewUsageMeterCreatedEvent(
		meter.id,
		billingAccountID,
		metricName,
		eventName,
	))

	return meter, nil
}

// ReconstructUsageMeter reconstrói um meter do banco
func ReconstructUsageMeter(
	id uuid.UUID,
	version int, // Optimistic locking version
	billingAccountID uuid.UUID,
	stripeCustomerID string,
	stripeMeterID string,
	metricName string,
	eventName string,
	quantity int64,
	periodStart time.Time,
	periodEnd time.Time,
	lastReportedAt *time.Time,
	metadata map[string]string,
	createdAt time.Time,
	updatedAt time.Time,
) *UsageMeter {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &UsageMeter{
		id:               id,
		version:          version,
		billingAccountID: billingAccountID,
		stripeCustomerID: stripeCustomerID,
		stripeMeterID:    stripeMeterID,
		metricName:       metricName,
		eventName:        eventName,
		quantity:         quantity,
		periodStart:      periodStart,
		periodEnd:        periodEnd,
		lastReportedAt:   lastReportedAt,
		metadata:         metadata,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		events:           []shared.DomainEvent{},
	}
}

// IncrementUsage incrementa o uso
func (m *UsageMeter) IncrementUsage(quantity int64) error {
	if quantity < 0 {
		return ErrNegativeQuantity
	}

	m.quantity += quantity
	m.updatedAt = time.Now()

	m.addEvent(NewUsageIncrementedEvent(m.id, m.metricName, quantity, m.quantity))

	return nil
}

// ReportToStripe marca como reportado ao Stripe
func (m *UsageMeter) ReportToStripe() {
	now := time.Now()
	m.lastReportedAt = &now
	m.updatedAt = now

	m.addEvent(NewUsageReportedEvent(m.id, m.metricName, m.quantity))
}

// ResetForNewPeriod reseta o medidor para novo período
func (m *UsageMeter) ResetForNewPeriod(newPeriodStart, newPeriodEnd time.Time) {
	oldQuantity := m.quantity

	m.quantity = 0
	m.periodStart = newPeriodStart
	m.periodEnd = newPeriodEnd
	m.lastReportedAt = nil
	m.updatedAt = time.Now()

	m.addEvent(NewUsagePeriodResetEvent(m.id, m.metricName, oldQuantity, newPeriodStart, newPeriodEnd))
}

// IsInPeriod verifica se está dentro do período
func (m *UsageMeter) IsInPeriod() bool {
	now := time.Now()
	return now.After(m.periodStart) && now.Before(m.periodEnd)
}

// DaysUntilPeriodEnd retorna dias até fim do período
func (m *UsageMeter) DaysUntilPeriodEnd() int {
	duration := time.Until(m.periodEnd)
	return int(duration.Hours() / 24)
}

// SetMetadata define metadados customizados
func (m *UsageMeter) SetMetadata(key, value string) {
	if m.metadata == nil {
		m.metadata = make(map[string]string)
	}
	m.metadata[key] = value
	m.updatedAt = time.Now()
}

// Getters
func (m *UsageMeter) ID() uuid.UUID               { return m.id }
func (m *UsageMeter) Version() int                { return m.version }
func (m *UsageMeter) BillingAccountID() uuid.UUID { return m.billingAccountID }
func (m *UsageMeter) StripeCustomerID() string    { return m.stripeCustomerID }
func (m *UsageMeter) StripeMeterID() string       { return m.stripeMeterID }
func (m *UsageMeter) MetricName() string          { return m.metricName }
func (m *UsageMeter) EventName() string           { return m.eventName }
func (m *UsageMeter) Quantity() int64             { return m.quantity }
func (m *UsageMeter) PeriodStart() time.Time      { return m.periodStart }
func (m *UsageMeter) PeriodEnd() time.Time        { return m.periodEnd }
func (m *UsageMeter) LastReportedAt() *time.Time  { return m.lastReportedAt }
func (m *UsageMeter) Metadata() map[string]string { return m.metadata }
func (m *UsageMeter) CreatedAt() time.Time        { return m.createdAt }
func (m *UsageMeter) UpdatedAt() time.Time        { return m.updatedAt }

func (m *UsageMeter) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, m.events...)
}

func (m *UsageMeter) ClearEvents() {
	m.events = []shared.DomainEvent{}
}

func (m *UsageMeter) addEvent(event shared.DomainEvent) {
	m.events = append(m.events, event)
}

// Compile-time check that UsageMeter implements AggregateRoot interface
var _ shared.AggregateRoot = (*UsageMeter)(nil)
