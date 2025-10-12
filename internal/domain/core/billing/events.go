package billing

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

// Type alias for backwards compatibility with invoice, subscription, usage_meter
type DomainEvent = shared.DomainEvent

type BillingAccountCreatedEvent struct {
	shared.BaseEvent
	AccountID    uuid.UUID
	UserID       uuid.UUID
	Name         string
	BillingEmail string
}

func NewBillingAccountCreatedEvent(accountID, userID uuid.UUID, name, billingEmail string) BillingAccountCreatedEvent {
	return BillingAccountCreatedEvent{
		BaseEvent:    shared.NewBaseEvent("billing.account.created", time.Now()),
		AccountID:    accountID,
		UserID:       userID,
		Name:         name,
		BillingEmail: billingEmail,
	}
}

type PaymentMethodActivatedEvent struct {
	shared.BaseEvent
	AccountID     uuid.UUID
	PaymentMethod PaymentMethod
}

func NewPaymentMethodActivatedEvent(accountID uuid.UUID, paymentMethod PaymentMethod) PaymentMethodActivatedEvent {
	return PaymentMethodActivatedEvent{
		BaseEvent:     shared.NewBaseEvent("billing.payment.activated", time.Now()),
		AccountID:     accountID,
		PaymentMethod: paymentMethod,
	}
}

type BillingAccountSuspendedEvent struct {
	shared.BaseEvent
	AccountID uuid.UUID
	Reason    string
}

func NewBillingAccountSuspendedEvent(accountID uuid.UUID, reason string) BillingAccountSuspendedEvent {
	return BillingAccountSuspendedEvent{
		BaseEvent: shared.NewBaseEvent("billing.account.suspended", time.Now()),
		AccountID: accountID,
		Reason:    reason,
	}
}

type BillingAccountReactivatedEvent struct {
	shared.BaseEvent
	AccountID uuid.UUID
}

func NewBillingAccountReactivatedEvent(accountID uuid.UUID) BillingAccountReactivatedEvent {
	return BillingAccountReactivatedEvent{
		BaseEvent: shared.NewBaseEvent("billing.account.reactivated", time.Now()),
		AccountID: accountID,
	}
}

type BillingAccountCanceledEvent struct {
	shared.BaseEvent
	AccountID uuid.UUID
}

func NewBillingAccountCanceledEvent(accountID uuid.UUID) BillingAccountCanceledEvent {
	return BillingAccountCanceledEvent{
		BaseEvent: shared.NewBaseEvent("billing.account.canceled", time.Now()),
		AccountID: accountID,
	}
}

type StripeCustomerLinkedEvent struct {
	shared.BaseEvent
	AccountID        uuid.UUID
	StripeCustomerID string
}

func NewStripeCustomerLinkedEvent(accountID uuid.UUID, stripeCustomerID string) StripeCustomerLinkedEvent {
	return StripeCustomerLinkedEvent{
		BaseEvent:        shared.NewBaseEvent("billing.stripe.customer_linked", time.Now()),
		AccountID:        accountID,
		StripeCustomerID: stripeCustomerID,
	}
}

// Subscription Events

type SubscriptionCreatedEvent struct {
	shared.BaseEvent
	SubscriptionID       uuid.UUID
	BillingAccountID     uuid.UUID
	StripeSubscriptionID string
	StripePriceID        string
	Status               string
}

func NewSubscriptionCreatedEvent(subscriptionID, billingAccountID uuid.UUID, stripeSubscriptionID, stripePriceID, status string) SubscriptionCreatedEvent {
	return SubscriptionCreatedEvent{
		BaseEvent:            shared.NewBaseEvent("billing.subscription.created", time.Now()),
		SubscriptionID:       subscriptionID,
		BillingAccountID:     billingAccountID,
		StripeSubscriptionID: stripeSubscriptionID,
		StripePriceID:        stripePriceID,
		Status:               status,
	}
}

type SubscriptionStatusChangedEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
	OldStatus      string
	NewStatus      string
}

func NewSubscriptionStatusChangedEvent(subscriptionID uuid.UUID, oldStatus, newStatus string) SubscriptionStatusChangedEvent {
	return SubscriptionStatusChangedEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.status_changed", time.Now()),
		SubscriptionID: subscriptionID,
		OldStatus:      oldStatus,
		NewStatus:      newStatus,
	}
}

type SubscriptionPeriodUpdatedEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
}

func NewSubscriptionPeriodUpdatedEvent(subscriptionID uuid.UUID, periodStart, periodEnd time.Time) SubscriptionPeriodUpdatedEvent {
	return SubscriptionPeriodUpdatedEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.period_updated", time.Now()),
		SubscriptionID: subscriptionID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
	}
}

type SubscriptionPriceChangedEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
	OldPriceID     string
	NewPriceID     string
}

func NewSubscriptionPriceChangedEvent(subscriptionID uuid.UUID, oldPriceID, newPriceID string) SubscriptionPriceChangedEvent {
	return SubscriptionPriceChangedEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.price_changed", time.Now()),
		SubscriptionID: subscriptionID,
		OldPriceID:     oldPriceID,
		NewPriceID:     newPriceID,
	}
}

type SubscriptionTrialStartedEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
	TrialEnd       time.Time
}

func NewSubscriptionTrialStartedEvent(subscriptionID uuid.UUID, trialEnd time.Time) SubscriptionTrialStartedEvent {
	return SubscriptionTrialStartedEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.trial_started", time.Now()),
		SubscriptionID: subscriptionID,
		TrialEnd:       trialEnd,
	}
}

type SubscriptionCancelScheduledEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
	CancelAt       time.Time
}

func NewSubscriptionCancelScheduledEvent(subscriptionID uuid.UUID, cancelAt time.Time) SubscriptionCancelScheduledEvent {
	return SubscriptionCancelScheduledEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.cancel_scheduled", time.Now()),
		SubscriptionID: subscriptionID,
		CancelAt:       cancelAt,
	}
}

type SubscriptionCanceledEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
}

func NewSubscriptionCanceledEvent(subscriptionID uuid.UUID) SubscriptionCanceledEvent {
	return SubscriptionCanceledEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.canceled", time.Now()),
		SubscriptionID: subscriptionID,
	}
}

type SubscriptionReactivatedEvent struct {
	shared.BaseEvent
	SubscriptionID uuid.UUID
}

func NewSubscriptionReactivatedEvent(subscriptionID uuid.UUID) SubscriptionReactivatedEvent {
	return SubscriptionReactivatedEvent{
		BaseEvent:      shared.NewBaseEvent("billing.subscription.reactivated", time.Now()),
		SubscriptionID: subscriptionID,
	}
}

// Invoice Events

type InvoiceCreatedEvent struct {
	shared.BaseEvent
	InvoiceID        uuid.UUID
	BillingAccountID uuid.UUID
	StripeInvoiceID  string
	AmountDue        int64
	Currency         string
	Status           string
}

func NewInvoiceCreatedEvent(invoiceID, billingAccountID uuid.UUID, stripeInvoiceID string, amountDue int64, currency, status string) InvoiceCreatedEvent {
	return InvoiceCreatedEvent{
		BaseEvent:        shared.NewBaseEvent("billing.invoice.created", time.Now()),
		InvoiceID:        invoiceID,
		BillingAccountID: billingAccountID,
		StripeInvoiceID:  stripeInvoiceID,
		AmountDue:        amountDue,
		Currency:         currency,
		Status:           status,
	}
}

type InvoicePaidEvent struct {
	shared.BaseEvent
	InvoiceID  uuid.UUID
	AmountPaid int64
}

func NewInvoicePaidEvent(invoiceID uuid.UUID, amountPaid int64) InvoicePaidEvent {
	return InvoicePaidEvent{
		BaseEvent:  shared.NewBaseEvent("billing.invoice.paid", time.Now()),
		InvoiceID:  invoiceID,
		AmountPaid: amountPaid,
	}
}

type InvoicePaymentFailedEvent struct {
	shared.BaseEvent
	InvoiceID uuid.UUID
}

func NewInvoicePaymentFailedEvent(invoiceID uuid.UUID) InvoicePaymentFailedEvent {
	return InvoicePaymentFailedEvent{
		BaseEvent: shared.NewBaseEvent("billing.invoice.payment_failed", time.Now()),
		InvoiceID: invoiceID,
	}
}

type InvoiceMarkedUncollectibleEvent struct {
	shared.BaseEvent
	InvoiceID uuid.UUID
}

func NewInvoiceMarkedUncollectibleEvent(invoiceID uuid.UUID) InvoiceMarkedUncollectibleEvent {
	return InvoiceMarkedUncollectibleEvent{
		BaseEvent: shared.NewBaseEvent("billing.invoice.uncollectible", time.Now()),
		InvoiceID: invoiceID,
	}
}

type InvoiceVoidedEvent struct {
	shared.BaseEvent
	InvoiceID uuid.UUID
}

func NewInvoiceVoidedEvent(invoiceID uuid.UUID) InvoiceVoidedEvent {
	return InvoiceVoidedEvent{
		BaseEvent: shared.NewBaseEvent("billing.invoice.voided", time.Now()),
		InvoiceID: invoiceID,
	}
}

// UsageMeter Events

type UsageMeterCreatedEvent struct {
	shared.BaseEvent
	MeterID          uuid.UUID
	BillingAccountID uuid.UUID
	MetricName       string
	MeterEventName   string // Renamed to avoid conflict with BaseEvent.EventName()
}

func NewUsageMeterCreatedEvent(meterID, billingAccountID uuid.UUID, metricName, meterEventName string) UsageMeterCreatedEvent {
	return UsageMeterCreatedEvent{
		BaseEvent:        shared.NewBaseEvent("billing.usage_meter.created", time.Now()),
		MeterID:          meterID,
		BillingAccountID: billingAccountID,
		MetricName:       metricName,
		MeterEventName:   meterEventName,
	}
}

type UsageIncrementedEvent struct {
	shared.BaseEvent
	MeterID       uuid.UUID
	MetricName    string
	Quantity      int64
	TotalQuantity int64
}

func NewUsageIncrementedEvent(meterID uuid.UUID, metricName string, quantity, totalQuantity int64) UsageIncrementedEvent {
	return UsageIncrementedEvent{
		BaseEvent:     shared.NewBaseEvent("billing.usage.incremented", time.Now()),
		MeterID:       meterID,
		MetricName:    metricName,
		Quantity:      quantity,
		TotalQuantity: totalQuantity,
	}
}

type UsageReportedEvent struct {
	shared.BaseEvent
	MeterID    uuid.UUID
	MetricName string
	Quantity   int64
}

func NewUsageReportedEvent(meterID uuid.UUID, metricName string, quantity int64) UsageReportedEvent {
	return UsageReportedEvent{
		BaseEvent:  shared.NewBaseEvent("billing.usage.reported", time.Now()),
		MeterID:    meterID,
		MetricName: metricName,
		Quantity:   quantity,
	}
}

type UsagePeriodResetEvent struct {
	shared.BaseEvent
	MeterID        uuid.UUID
	MetricName     string
	OldQuantity    int64
	NewPeriodStart time.Time
	NewPeriodEnd   time.Time
}

func NewUsagePeriodResetEvent(meterID uuid.UUID, metricName string, oldQuantity int64, newPeriodStart, newPeriodEnd time.Time) UsagePeriodResetEvent {
	return UsagePeriodResetEvent{
		BaseEvent:      shared.NewBaseEvent("billing.usage.period_reset", time.Now()),
		MeterID:        meterID,
		MetricName:     metricName,
		OldQuantity:    oldQuantity,
		NewPeriodStart: newPeriodStart,
		NewPeriodEnd:   newPeriodEnd,
	}
}
