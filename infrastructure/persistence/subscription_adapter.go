package persistence

import (
	"context"
	"encoding/json"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/billing"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionRepositoryAdapter adapta o GormSubscriptionRepository para implementar billing.SubscriptionRepository
type SubscriptionRepositoryAdapter struct {
	gormRepo *GormSubscriptionRepository
}

// NewSubscriptionRepositoryAdapter cria um novo adapter
func NewSubscriptionRepositoryAdapter(db *gorm.DB) billing.SubscriptionRepository {
	return &SubscriptionRepositoryAdapter{
		gormRepo: NewGormSubscriptionRepository(db),
	}
}

// Create cria uma nova subscription
func (a *SubscriptionRepositoryAdapter) Create(ctx context.Context, subscription *billing.Subscription) error {
	entity, err := a.domainToEntity(subscription)
	if err != nil {
		return err
	}
	return a.gormRepo.Create(ctx, entity)
}

// FindByID busca uma subscription por ID
func (a *SubscriptionRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*billing.Subscription, error) {
	entity, err := a.gormRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByStripeSubscriptionID busca uma subscription pelo Stripe Subscription ID
func (a *SubscriptionRepositoryAdapter) FindByStripeSubscriptionID(ctx context.Context, stripeSubID string) (*billing.Subscription, error) {
	entity, err := a.gormRepo.FindByStripeSubscriptionID(ctx, stripeSubID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByBillingAccount busca todas as subscriptions de uma billing account
func (a *SubscriptionRepositoryAdapter) FindByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) ([]*billing.Subscription, error) {
	entities, err := a.gormRepo.FindByBillingAccountID(ctx, billingAccountID)
	if err != nil {
		return nil, err
	}

	subscriptions := make([]*billing.Subscription, len(entities))
	for i, entity := range entities {
		subscription, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		subscriptions[i] = subscription
	}
	return subscriptions, nil
}

// FindActiveByBillingAccount busca a subscription ativa de uma billing account
func (a *SubscriptionRepositoryAdapter) FindActiveByBillingAccount(ctx context.Context, billingAccountID uuid.UUID) (*billing.Subscription, error) {
	entities, err := a.gormRepo.FindActiveByBillingAccountID(ctx, billingAccountID)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return nil, billing.ErrNotFound
	}
	return a.entityToDomain(entities[0])
}

// FindByStatus busca subscriptions por status
func (a *SubscriptionRepositoryAdapter) FindByStatus(ctx context.Context, status billing.SubscriptionStatus) ([]*billing.Subscription, error) {
	var entities []*entities.SubscriptionEntity
	err := a.gormRepo.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("created_at DESC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	subscriptions := make([]*billing.Subscription, len(entities))
	for i, entity := range entities {
		subscription, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		subscriptions[i] = subscription
	}
	return subscriptions, nil
}

// FindExpiringTrials busca subscriptions com trial expirando nos próximos N dias
func (a *SubscriptionRepositoryAdapter) FindExpiringTrials(ctx context.Context, daysUntilExpiration int) ([]*billing.Subscription, error) {
	entities, err := a.gormRepo.FindExpiringSubscriptions(ctx, daysUntilExpiration)
	if err != nil {
		return nil, err
	}

	subscriptions := make([]*billing.Subscription, len(entities))
	for i, entity := range entities {
		subscription, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		subscriptions[i] = subscription
	}
	return subscriptions, nil
}

// Update atualiza uma subscription existente
func (a *SubscriptionRepositoryAdapter) Update(ctx context.Context, subscription *billing.Subscription) error {
	entity, err := a.domainToEntity(subscription)
	if err != nil {
		return err
	}
	return a.gormRepo.Update(ctx, entity)
}

// Delete remove uma subscription
func (a *SubscriptionRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.gormRepo.Delete(ctx, id)
}

// domainToEntity converte billing.Subscription para entities.SubscriptionEntity
func (a *SubscriptionRepositoryAdapter) domainToEntity(subscription *billing.Subscription) (*entities.SubscriptionEntity, error) {
	// Converter metadata para JSON bytes
	var metadataBytes []byte
	if len(subscription.Metadata()) > 0 {
		var err error
		metadataBytes, err = json.Marshal(subscription.Metadata())
		if err != nil {
			return nil, err
		}
	}

	return &entities.SubscriptionEntity{
		ID:                   subscription.ID(),
		Version:              subscription.Version(),
		BillingAccountID:     subscription.BillingAccountID(),
		StripeSubscriptionID: subscription.StripeSubscriptionID(),
		StripePriceID:        subscription.StripePriceID(),
		Status:               string(subscription.Status()),
		CurrentPeriodStart:   subscription.CurrentPeriodStart(),
		CurrentPeriodEnd:     subscription.CurrentPeriodEnd(),
		TrialStart:           subscription.TrialStart(),
		TrialEnd:             subscription.TrialEnd(),
		CancelAt:             subscription.CancelAt(),
		CanceledAt:           subscription.CanceledAt(),
		CancelAtPeriodEnd:    subscription.CancelAtPeriodEnd(),
		Metadata:             metadataBytes,
		CreatedAt:            subscription.CreatedAt(),
		UpdatedAt:            subscription.UpdatedAt(),
	}, nil
}

// entityToDomain converte entities.SubscriptionEntity para billing.Subscription
func (a *SubscriptionRepositoryAdapter) entityToDomain(entity *entities.SubscriptionEntity) (*billing.Subscription, error) {
	// Converter JSON bytes para metadata
	var metadata map[string]string
	if len(entity.Metadata) > 0 {
		err := json.Unmarshal(entity.Metadata, &metadata)
		if err != nil {
			return nil, err
		}
	}

	return billing.ReconstructSubscription(
		entity.ID,
		entity.Version,
		entity.BillingAccountID,
		entity.StripeSubscriptionID,
		entity.StripePriceID,
		billing.SubscriptionStatus(entity.Status),
		entity.CurrentPeriodStart,
		entity.CurrentPeriodEnd,
		entity.TrialStart,
		entity.TrialEnd,
		entity.CancelAt,
		entity.CanceledAt,
		entity.CancelAtPeriodEnd,
		metadata,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}
