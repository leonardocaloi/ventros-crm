package persistence

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/project"
	"github.com/google/uuid"
)

// MockProjectRepository is a temporary mock implementation
type MockProjectRepository struct{}

func NewMockProjectRepository() project.Repository {
	return &MockProjectRepository{}
}

func (r *MockProjectRepository) Save(ctx context.Context, proj *project.Project) error {
	return fmt.Errorf("not implemented")
}

func (r *MockProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *MockProjectRepository) FindByTenantID(ctx context.Context, tenantID string) (*project.Project, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *MockProjectRepository) FindByCustomer(ctx context.Context, customerID uuid.UUID) ([]*project.Project, error) {
	return nil, fmt.Errorf("not implemented")
}
