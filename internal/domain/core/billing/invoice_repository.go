package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// InvoiceRepository define operações de persistência para Invoices
type InvoiceRepository interface {
	// Create cria uma nova invoice
	Create(ctx context.Context, invoice *Invoice) error

	// FindByID busca uma invoice por ID
	FindByID(ctx context.Context, id uuid.UUID) (*Invoice, error)

	// FindByBillingAccount busca todas as invoices de uma billing account
	FindByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*Invoice, error)

	// FindByStripeInvoiceID busca invoice pelo ID do Stripe
	FindByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*Invoice, error)

	// FindBySubscription busca invoices de uma subscription
	FindBySubscription(ctx context.Context, subscriptionID uuid.UUID) ([]*Invoice, error)

	// Update atualiza uma invoice existente
	Update(ctx context.Context, invoice *Invoice) error

	// Delete remove uma invoice
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByStatus busca invoices por status
	FindByStatus(ctx context.Context, status InvoiceStatus) ([]*Invoice, error)

	// FindOverdue busca invoices vencidas
	FindOverdue(ctx context.Context) ([]*Invoice, error)

	// FindUnpaidByBillingAccount busca invoices não pagas de uma billing account
	FindUnpaidByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*Invoice, error)

	// FindDueInPeriod busca invoices com vencimento em um período
	FindDueInPeriod(ctx context.Context, start, end time.Time) ([]*Invoice, error)
}
