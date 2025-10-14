package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/core/shared"
	"gorm.io/gorm"
)

// GormBillingRepository implementa o repositório de billing accounts usando GORM
type GormBillingRepository struct {
	db *gorm.DB
}

// NewGormBillingRepository cria uma nova instância do repositório
func NewGormBillingRepository(db *gorm.DB) *GormBillingRepository {
	return &GormBillingRepository{db: db}
}

// Create cria uma nova billing account
func (r *GormBillingRepository) Create(ctx context.Context, account *entities.BillingAccountEntity) error {
	return r.db.WithContext(ctx).Create(account).Error
}

// FindByID busca uma billing account por ID
func (r *GormBillingRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.BillingAccountEntity, error) {
	var account entities.BillingAccountEntity
	err := r.db.WithContext(ctx).First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// FindByUserID busca todas as billing accounts de um usuário
func (r *GormBillingRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.BillingAccountEntity, error) {
	var accounts []*entities.BillingAccountEntity
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

// FindActiveByUserID busca a primeira conta ativa de um usuário
func (r *GormBillingRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*entities.BillingAccountEntity, error) {
	var account entities.BillingAccountEntity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND payment_status = ? AND suspended = ?", userID, "active", false).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Update atualiza uma billing account
func (r *GormBillingRepository) Update(ctx context.Context, account *entities.BillingAccountEntity) error {
	// Check if exists
	var existing entities.BillingAccountEntity
	err := r.db.WithContext(ctx).Where("id = ?", account.ID).First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Insert if not found
			return r.db.WithContext(ctx).Create(account).Error
		}
		return err
	}

	// Update with optimistic locking
	result := r.db.WithContext(ctx).Model(&entities.BillingAccountEntity{}).
		Where("id = ? AND version = ?", account.ID, existing.Version).
		Updates(map[string]interface{}{
			"version":            existing.Version + 1, // Increment version
			"user_id":            account.UserID,
			"name":               account.Name,
			"stripe_customer_id": account.StripeCustomerID,
			"payment_status":     account.PaymentStatus,
			"payment_methods":    account.PaymentMethods,
			"billing_email":      account.BillingEmail,
			"suspended":          account.Suspended,
			"suspended_at":       account.SuspendedAt,
			"suspension_reason":  account.SuspensionReason,
			"updated_at":         account.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
	if result.RowsAffected == 0 {
		return shared.NewOptimisticLockError(
			"BillingAccount",
			account.ID.String(),
			existing.Version,
			account.Version,
		)
	}

	return nil
}

// Delete deleta uma billing account (soft delete)
func (r *GormBillingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.BillingAccountEntity{}, "id = ?", id).Error
}
