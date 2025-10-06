package customer

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, customer *Customer) error
	FindByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	FindByEmail(ctx context.Context, email string) (*Customer, error)
	FindAll(ctx context.Context, limit, offset int) ([]*Customer, error)
}
