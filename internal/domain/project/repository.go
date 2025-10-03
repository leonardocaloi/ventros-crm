package project

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*Project, error)
	FindByTenantID(ctx context.Context, tenantID string) (*Project, error)
	FindByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Project, error)
}
