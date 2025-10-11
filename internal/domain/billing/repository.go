package billing

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, account *BillingAccount) error

	FindByID(ctx context.Context, id uuid.UUID) (*BillingAccount, error)

	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*BillingAccount, error)

	FindActiveByUserID(ctx context.Context, userID uuid.UUID) (*BillingAccount, error)

	Update(ctx context.Context, account *BillingAccount) error

	Delete(ctx context.Context, id uuid.UUID) error
}
