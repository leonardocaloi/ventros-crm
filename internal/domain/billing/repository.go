package billing

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para persistência de contas de faturamento
type Repository interface {
	// Create cria uma nova conta de faturamento
	Create(ctx context.Context, account *BillingAccount) error

	// FindByID busca uma conta por ID
	FindByID(ctx context.Context, id uuid.UUID) (*BillingAccount, error)

	// FindByUserID busca contas de um usuário
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*BillingAccount, error)

	// FindActiveByUserID busca a primeira conta ativa de um usuário
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*BillingAccount, error)

	// Update atualiza uma conta existente
	Update(ctx context.Context, account *BillingAccount) error

	// Delete remove uma conta (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error
}
