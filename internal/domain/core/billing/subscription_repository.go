package billing

import (
	"context"

	"github.com/google/uuid"
)

// SubscriptionRepository define operações de persistência para Subscriptions
type SubscriptionRepository interface {
	// Create cria uma nova subscription
	Create(ctx context.Context, subscription *Subscription) error

	// FindByID busca uma subscription por ID
	FindByID(ctx context.Context, id uuid.UUID) (*Subscription, error)

	// FindByBillingAccount busca todas as subscriptions de uma billing account
	FindByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*Subscription, error)

	// FindActiveByBillingAccount busca a subscription ativa de uma billing account
	FindActiveByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) (*Subscription, error)

	// FindByStripeSubscriptionID busca subscription pelo ID do Stripe
	FindByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error)

	// Update atualiza uma subscription existente
	Update(ctx context.Context, subscription *Subscription) error

	// Delete remove uma subscription
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByStatus busca subscriptions por status
	FindByStatus(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error)

	// FindExpiringTrials busca subscriptions com trial expirando nos próximos N dias
	FindExpiringTrials(ctx context.Context, daysUntilExpiration int) ([]*Subscription, error)
}
