package project

import (
	"context"

	"github.com/google/uuid"
)

// ProjectFilters represents filtering options for project queries
type ProjectFilters struct {
	TenantID   string
	CustomerID *uuid.UUID
	Active     *bool
	Limit      int
	Offset     int
	SortBy     string // name, created_at
	SortOrder  string // asc, desc
}

type Repository interface {
	Save(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*Project, error)
	FindByTenantID(ctx context.Context, tenantID string) (*Project, error)
	FindByCustomer(ctx context.Context, customerID uuid.UUID) ([]*Project, error)

	// Advanced query methods
	FindByTenantWithFilters(ctx context.Context, filters ProjectFilters) ([]*Project, int64, error)

	SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Project, int64, error)
}
