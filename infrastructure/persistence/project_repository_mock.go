package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/project"
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

func (r *MockProjectRepository) FindByTenantWithFilters(ctx context.Context, filters project.ProjectFilters) ([]*project.Project, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *MockProjectRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*project.Project, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}
