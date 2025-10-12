package billing

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

// InvoiceStatus representa o status de uma invoice no Stripe
type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "draft"         // Draft invoice
	InvoiceStatusOpen          InvoiceStatus = "open"          // Open (unpaid)
	InvoiceStatusPaid          InvoiceStatus = "paid"          // Paid successfully
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible" // Uncollectible (after retries)
	InvoiceStatusVoid          InvoiceStatus = "void"          // Void (canceled)
)

// Invoice representa uma invoice do Stripe (domínio)
type Invoice struct {
	id                   uuid.UUID
	version              int // Optimistic locking - prevents lost updates
	billingAccountID     uuid.UUID
	stripeInvoiceID      string // Stripe invoice ID (in_xxx)
	subscriptionID       *uuid.UUID
	stripeSubscriptionID *string
	amountDue            int64  // Amount in cents
	amountPaid           int64  // Amount paid in cents
	amountRemaining      int64  // Remaining to pay in cents
	currency             string // USD, BRL, EUR
	status               InvoiceStatus
	hostedInvoiceURL     string // Stripe hosted invoice URL
	invoicePDF           string // PDF URL
	dueDate              *time.Time
	paidAt               *time.Time
	metadata             map[string]string
	createdAt            time.Time
	updatedAt            time.Time

	events []shared.DomainEvent
}

var (
	ErrInvalidStripeInvoiceID = errors.New("stripe invoice ID cannot be empty")
	ErrInvalidAmount          = errors.New("amount must be non-negative")
	ErrInvalidCurrency        = errors.New("currency cannot be empty")
	ErrInvoicePaid            = errors.New("invoice is already paid")
	ErrInvoiceVoid            = errors.New("invoice is void")
)

// NewInvoice cria uma nova invoice
func NewInvoice(
	billingAccountID uuid.UUID,
	stripeInvoiceID string,
	subscriptionID *uuid.UUID,
	amountDue int64,
	currency string,
	status InvoiceStatus,
) (*Invoice, error) {
	if billingAccountID == uuid.Nil {
		return nil, ErrInvalidBillingAccountID
	}
	if stripeInvoiceID == "" {
		return nil, ErrInvalidStripeInvoiceID
	}
	if amountDue < 0 {
		return nil, ErrInvalidAmount
	}
	if currency == "" {
		return nil, ErrInvalidCurrency
	}

	now := time.Now()
	invoice := &Invoice{
		id:               uuid.New(),
		version:          1, // Start with version 1 for new aggregates
		billingAccountID: billingAccountID,
		stripeInvoiceID:  stripeInvoiceID,
		subscriptionID:   subscriptionID,
		amountDue:        amountDue,
		amountPaid:       0,
		amountRemaining:  amountDue,
		currency:         currency,
		status:           status,
		metadata:         make(map[string]string),
		createdAt:        now,
		updatedAt:        now,
		events:           []shared.DomainEvent{},
	}

	invoice.addEvent(NewInvoiceCreatedEvent(
		invoice.id,
		billingAccountID,
		stripeInvoiceID,
		amountDue,
		currency,
		string(status),
	))

	return invoice, nil
}

// ReconstructInvoice reconstrói uma invoice do banco
func ReconstructInvoice(
	id uuid.UUID,
	version int, // Optimistic locking version
	billingAccountID uuid.UUID,
	stripeInvoiceID string,
	subscriptionID *uuid.UUID,
	stripeSubscriptionID *string,
	amountDue int64,
	amountPaid int64,
	amountRemaining int64,
	currency string,
	status InvoiceStatus,
	hostedInvoiceURL string,
	invoicePDF string,
	dueDate *time.Time,
	paidAt *time.Time,
	metadata map[string]string,
	createdAt time.Time,
	updatedAt time.Time,
) *Invoice {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}

	return &Invoice{
		id:                   id,
		version:              version,
		billingAccountID:     billingAccountID,
		stripeInvoiceID:      stripeInvoiceID,
		subscriptionID:       subscriptionID,
		stripeSubscriptionID: stripeSubscriptionID,
		amountDue:            amountDue,
		amountPaid:           amountPaid,
		amountRemaining:      amountRemaining,
		currency:             currency,
		status:               status,
		hostedInvoiceURL:     hostedInvoiceURL,
		invoicePDF:           invoicePDF,
		dueDate:              dueDate,
		paidAt:               paidAt,
		metadata:             metadata,
		createdAt:            createdAt,
		updatedAt:            updatedAt,
		events:               []shared.DomainEvent{},
	}
}

// MarkAsPaid marca a invoice como paga
func (i *Invoice) MarkAsPaid(amountPaid int64) error {
	if i.status == InvoiceStatusPaid {
		return ErrInvoicePaid
	}
	if i.status == InvoiceStatusVoid {
		return ErrInvoiceVoid
	}

	now := time.Now()
	i.status = InvoiceStatusPaid
	i.amountPaid = amountPaid
	i.amountRemaining = i.amountDue - amountPaid
	i.paidAt = &now
	i.updatedAt = now

	i.addEvent(NewInvoicePaidEvent(i.id, amountPaid))

	return nil
}

// MarkAsFailedPayment marca a invoice com falha no pagamento
func (i *Invoice) MarkAsFailedPayment() {
	if i.status == InvoiceStatusOpen {
		// Mantém como open mas registra tentativa
		now := time.Now()
		i.updatedAt = now

		i.addEvent(NewInvoicePaymentFailedEvent(i.id))
	}
}

// MarkAsUncollectible marca como não cobrável (após múltiplas falhas)
func (i *Invoice) MarkAsUncollectible() {
	now := time.Now()
	i.status = InvoiceStatusUncollectible
	i.updatedAt = now

	i.addEvent(NewInvoiceMarkedUncollectibleEvent(i.id))
}

// Void cancela a invoice
func (i *Invoice) Void() error {
	if i.status == InvoiceStatusPaid {
		return ErrInvoicePaid
	}

	now := time.Now()
	i.status = InvoiceStatusVoid
	i.updatedAt = now

	i.addEvent(NewInvoiceVoidedEvent(i.id))

	return nil
}

// UpdatePaymentDetails atualiza detalhes de pagamento parcial
func (i *Invoice) UpdatePaymentDetails(amountPaid int64) {
	i.amountPaid = amountPaid
	i.amountRemaining = i.amountDue - amountPaid
	i.updatedAt = time.Now()
}

// SetInvoiceURLs define URLs da invoice
func (i *Invoice) SetInvoiceURLs(hostedURL, pdfURL string) {
	i.hostedInvoiceURL = hostedURL
	i.invoicePDF = pdfURL
	i.updatedAt = time.Now()
}

// SetDueDate define data de vencimento
func (i *Invoice) SetDueDate(dueDate time.Time) {
	i.dueDate = &dueDate
	i.updatedAt = time.Now()
}

// IsOverdue verifica se a invoice está vencida
func (i *Invoice) IsOverdue() bool {
	if i.status == InvoiceStatusPaid || i.status == InvoiceStatusVoid {
		return false
	}
	if i.dueDate == nil {
		return false
	}
	return time.Now().After(*i.dueDate)
}

// IsPaid verifica se está paga
func (i *Invoice) IsPaid() bool {
	return i.status == InvoiceStatusPaid
}

// IsOpen verifica se está aberta (pendente pagamento)
func (i *Invoice) IsOpen() bool {
	return i.status == InvoiceStatusOpen
}

// DaysOverdue retorna dias de atraso
func (i *Invoice) DaysOverdue() int {
	if !i.IsOverdue() {
		return 0
	}
	duration := time.Since(*i.dueDate)
	return int(duration.Hours() / 24)
}

// SetMetadata define metadados customizados
func (i *Invoice) SetMetadata(key, value string) {
	if i.metadata == nil {
		i.metadata = make(map[string]string)
	}
	i.metadata[key] = value
	i.updatedAt = time.Now()
}

// Getters
func (i *Invoice) ID() uuid.UUID                 { return i.id }
func (i *Invoice) Version() int                  { return i.version }
func (i *Invoice) BillingAccountID() uuid.UUID   { return i.billingAccountID }
func (i *Invoice) StripeInvoiceID() string       { return i.stripeInvoiceID }
func (i *Invoice) SubscriptionID() *uuid.UUID    { return i.subscriptionID }
func (i *Invoice) StripeSubscriptionID() *string { return i.stripeSubscriptionID }
func (i *Invoice) AmountDue() int64              { return i.amountDue }
func (i *Invoice) AmountPaid() int64             { return i.amountPaid }
func (i *Invoice) AmountRemaining() int64        { return i.amountRemaining }
func (i *Invoice) Currency() string              { return i.currency }
func (i *Invoice) Status() InvoiceStatus         { return i.status }
func (i *Invoice) HostedInvoiceURL() string      { return i.hostedInvoiceURL }
func (i *Invoice) InvoicePDF() string            { return i.invoicePDF }
func (i *Invoice) DueDate() *time.Time           { return i.dueDate }
func (i *Invoice) PaidAt() *time.Time            { return i.paidAt }
func (i *Invoice) Metadata() map[string]string   { return i.metadata }
func (i *Invoice) CreatedAt() time.Time          { return i.createdAt }
func (i *Invoice) UpdatedAt() time.Time          { return i.updatedAt }

func (i *Invoice) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, i.events...)
}

func (i *Invoice) ClearEvents() {
	i.events = []shared.DomainEvent{}
}

func (i *Invoice) addEvent(event shared.DomainEvent) {
	i.events = append(i.events, event)
}

// Compile-time check that Invoice implements AggregateRoot interface
var _ shared.AggregateRoot = (*Invoice)(nil)
