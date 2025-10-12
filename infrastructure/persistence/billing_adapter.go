package persistence

import (
	"context"
	"encoding/json"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/billing"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillingRepositoryAdapter adapta o GormBillingRepository para implementar billing.Repository
type BillingRepositoryAdapter struct {
	gormRepo *GormBillingRepository
}

// NewBillingRepositoryAdapter cria um novo adapter
func NewBillingRepositoryAdapter(db *gorm.DB) billing.Repository {
	return &BillingRepositoryAdapter{
		gormRepo: NewGormBillingRepository(db),
	}
}

// Create cria uma nova conta de faturamento
func (a *BillingRepositoryAdapter) Create(ctx context.Context, account *billing.BillingAccount) error {
	entity, err := a.domainToEntity(account)
	if err != nil {
		return err
	}
	return a.gormRepo.Create(ctx, entity)
}

// FindByID busca uma conta por ID
func (a *BillingRepositoryAdapter) FindByID(ctx context.Context, id uuid.UUID) (*billing.BillingAccount, error) {
	entity, err := a.gormRepo.FindByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// FindByUserID busca contas de um usuário
func (a *BillingRepositoryAdapter) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*billing.BillingAccount, error) {
	entities, err := a.gormRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accounts := make([]*billing.BillingAccount, len(entities))
	for i, entity := range entities {
		account, err := a.entityToDomain(entity)
		if err != nil {
			return nil, err
		}
		accounts[i] = account
	}
	return accounts, nil
}

// FindActiveByUserID busca a primeira conta ativa de um usuário
func (a *BillingRepositoryAdapter) FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*billing.BillingAccount, error) {
	entity, err := a.gormRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return a.entityToDomain(entity)
}

// Update atualiza uma conta existente
func (a *BillingRepositoryAdapter) Update(ctx context.Context, account *billing.BillingAccount) error {
	entity, err := a.domainToEntity(account)
	if err != nil {
		return err
	}
	return a.gormRepo.Update(ctx, entity)
}

// Delete remove uma conta
func (a *BillingRepositoryAdapter) Delete(ctx context.Context, id uuid.UUID) error {
	return a.gormRepo.Delete(ctx, id)
}

// domainToEntity converte billing.BillingAccount para entities.BillingAccountEntity
func (a *BillingRepositoryAdapter) domainToEntity(account *billing.BillingAccount) (*entities.BillingAccountEntity, error) {
	// Converter payment methods para JSON bytes
	var paymentMethodsBytes []byte
	if len(account.PaymentMethods()) > 0 {
		var err error
		paymentMethodsBytes, err = json.Marshal(account.PaymentMethods())
		if err != nil {
			return nil, err
		}
	}

	return &entities.BillingAccountEntity{
		ID:               account.ID(),
		Version:          account.Version(),
		UserID:           account.UserID(),
		Name:             account.Name(),
		StripeCustomerID: account.StripeCustomerID(),
		PaymentStatus:    string(account.PaymentStatus()),
		PaymentMethods:   paymentMethodsBytes,
		BillingEmail:     account.BillingEmail(),
		Suspended:        account.IsSuspended(),
		SuspendedAt:      account.SuspendedAt(),
		SuspensionReason: account.SuspensionReason(),
		CreatedAt:        account.CreatedAt(),
		UpdatedAt:        account.UpdatedAt(),
	}, nil
}

// entityToDomain converte entities.BillingAccountEntity para billing.BillingAccount
func (a *BillingRepositoryAdapter) entityToDomain(entity *entities.BillingAccountEntity) (*billing.BillingAccount, error) {
	// Converter JSON bytes para payment methods
	var paymentMethods []billing.PaymentMethod
	if len(entity.PaymentMethods) > 0 {
		err := json.Unmarshal(entity.PaymentMethods, &paymentMethods)
		if err != nil {
			return nil, err
		}
	}

	return billing.ReconstructBillingAccount(
		entity.ID,
		entity.Version,
		entity.UserID,
		entity.Name,
		entity.StripeCustomerID,
		billing.PaymentStatus(entity.PaymentStatus),
		paymentMethods,
		entity.BillingEmail,
		entity.Suspended,
		entity.SuspendedAt,
		entity.SuspensionReason,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}
