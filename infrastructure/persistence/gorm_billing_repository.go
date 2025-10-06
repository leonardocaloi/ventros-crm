package persistence

import (
	"context"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/google/uuid"
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
	return r.db.WithContext(ctx).Save(account).Error
}

// Delete deleta uma billing account (soft delete)
func (r *GormBillingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.BillingAccountEntity{}, "id = ?", id).Error
}
