package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"gorm.io/gorm"
)

// GormWebhookRepository implementa o repositório de webhook subscriptions usando GORM
type GormWebhookRepository struct {
	db *gorm.DB
}

// NewGormWebhookRepository cria uma nova instância do repositório
func NewGormWebhookRepository(db *gorm.DB) *GormWebhookRepository {
	return &GormWebhookRepository{db: db}
}

// Create cria uma nova webhook subscription
func (r *GormWebhookRepository) Create(ctx context.Context, webhook *entities.WebhookSubscriptionEntity) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

// FindByID busca uma webhook subscription por ID
func (r *GormWebhookRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.WebhookSubscriptionEntity, error) {
	var webhook entities.WebhookSubscriptionEntity
	err := r.db.WithContext(ctx).First(&webhook, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// FindAll lista todas as webhook subscriptions
func (r *GormWebhookRepository) FindAll(ctx context.Context) ([]*entities.WebhookSubscriptionEntity, error) {
	var webhooks []*entities.WebhookSubscriptionEntity
	err := r.db.WithContext(ctx).Find(&webhooks).Error
	return webhooks, err
}

// FindActive lista apenas webhook subscriptions ativas
func (r *GormWebhookRepository) FindActive(ctx context.Context) ([]*entities.WebhookSubscriptionEntity, error) {
	var webhooks []*entities.WebhookSubscriptionEntity
	err := r.db.WithContext(ctx).Where("active = ?", true).Find(&webhooks).Error
	return webhooks, err
}

// Update atualiza uma webhook subscription
func (r *GormWebhookRepository) Update(ctx context.Context, webhook *entities.WebhookSubscriptionEntity) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

// Delete deleta uma webhook subscription (soft delete)
func (r *GormWebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.WebhookSubscriptionEntity{}, "id = ?", id).Error
}
