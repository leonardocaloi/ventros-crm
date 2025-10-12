package persistence

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormSubscriptionRepository implementa o repositório de subscriptions usando GORM
type GormSubscriptionRepository struct {
	db *gorm.DB
}

// NewGormSubscriptionRepository cria uma nova instância do repositório
func NewGormSubscriptionRepository(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

// Create cria uma nova subscription
func (r *GormSubscriptionRepository) Create(ctx context.Context, subscription *entities.SubscriptionEntity) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

// FindByID busca uma subscription por ID
func (r *GormSubscriptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.SubscriptionEntity, error) {
	var subscription entities.SubscriptionEntity
	err := r.db.WithContext(ctx).First(&subscription, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// FindByStripeSubscriptionID busca uma subscription pelo Stripe Subscription ID
func (r *GormSubscriptionRepository) FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*entities.SubscriptionEntity, error) {
	var subscription entities.SubscriptionEntity
	err := r.db.WithContext(ctx).Where("stripe_subscription_id = ?", stripeSubID).First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// FindByBillingAccountID busca todas as subscriptions de uma billing account
func (r *GormSubscriptionRepository) FindByBillingAccountID(ctx context.Context, billingAccountID uuid.UUID) ([]*entities.SubscriptionEntity, error) {
	var subscriptions []*entities.SubscriptionEntity
	err := r.db.WithContext(ctx).
		Where("billing_account_id = ?", billingAccountID).
		Order("created_at DESC").
		Find(&subscriptions).Error
	return subscriptions, err
}

// FindActiveByBillingAccountID busca subscriptions ativas de uma billing account
func (r *GormSubscriptionRepository) FindActiveByBillingAccountID(ctx context.Context, billingAccountID uuid.UUID) ([]*entities.SubscriptionEntity, error) {
	var subscriptions []*entities.SubscriptionEntity
	err := r.db.WithContext(ctx).
		Where("billing_account_id = ? AND status IN (?)", billingAccountID, []string{"active", "trialing"}).
		Order("created_at DESC").
		Find(&subscriptions).Error
	return subscriptions, err
}

// FindExpiringSubscriptions busca subscriptions que expiram em breve
func (r *GormSubscriptionRepository) FindExpiringSubscriptions(ctx context.Context, daysUntilExpiration int) ([]*entities.SubscriptionEntity, error) {
	var subscriptions []*entities.SubscriptionEntity
	err := r.db.WithContext(ctx).
		Where("status IN (?) AND current_period_end <= NOW() + INTERVAL '? days'",
			[]string{"active", "trialing"}, daysUntilExpiration).
		Order("current_period_end ASC").
		Find(&subscriptions).Error
	return subscriptions, err
}

// Update atualiza uma subscription
func (r *GormSubscriptionRepository) Update(ctx context.Context, subscription *entities.SubscriptionEntity) error {
	// Check if exists
	var existing entities.SubscriptionEntity
	err := r.db.WithContext(ctx).Where("id = ?", subscription.ID).First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Insert if not found
			return r.db.WithContext(ctx).Create(subscription).Error
		}
		return err
	}

	// Update with optimistic locking
	result := r.db.WithContext(ctx).Model(&entities.SubscriptionEntity{}).
		Where("id = ? AND version = ?", subscription.ID, existing.Version).
		Updates(map[string]interface{}{
			"version":                 existing.Version + 1, // Increment version
			"billing_account_id":      subscription.BillingAccountID,
			"stripe_subscription_id":  subscription.StripeSubscriptionID,
			"stripe_price_id":         subscription.StripePriceID,
			"status":                  subscription.Status,
			"current_period_start":    subscription.CurrentPeriodStart,
			"current_period_end":      subscription.CurrentPeriodEnd,
			"trial_start":             subscription.TrialStart,
			"trial_end":               subscription.TrialEnd,
			"cancel_at":               subscription.CancelAt,
			"canceled_at":             subscription.CanceledAt,
			"cancel_at_period_end":    subscription.CancelAtPeriodEnd,
			"metadata":                subscription.Metadata,
			"updated_at":              subscription.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
	if result.RowsAffected == 0 {
		return shared.NewOptimisticLockError(
			"Subscription",
			subscription.ID.String(),
			existing.Version,
			subscription.Version,
		)
	}

	return nil
}

// Delete deleta uma subscription (soft delete)
func (r *GormSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.SubscriptionEntity{}, "id = ?", id).Error
}
