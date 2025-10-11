package project

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/core/project"
	"github.com/google/uuid"
)

// GetProjectUseCase handles project retrieval
type GetProjectUseCase struct {
	projectRepo project.Repository
}

// NewGetProjectUseCase creates a new instance
func NewGetProjectUseCase(projectRepo project.Repository) *GetProjectUseCase {
	return &GetProjectUseCase{
		projectRepo: projectRepo,
	}
}

// GetProjectRequest represents the request to get a project
type GetProjectRequest struct {
	ProjectID uuid.UUID `json:"project_id" validate:"required"`
}

// GetProjectByTenantRequest represents the request to get a project by tenant ID
type GetProjectByTenantRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
}

// GetProjectResponse represents the response with project details
type GetProjectResponse struct {
	ProjectID        uuid.UUID              `json:"project_id"`
	CustomerID       uuid.UUID              `json:"customer_id"`
	BillingAccountID uuid.UUID              `json:"billing_account_id"`
	TenantID         string                 `json:"tenant_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Configuration    map[string]interface{} `json:"configuration"`
	IsActive         bool                   `json:"is_active"`
	SessionTimeout   int                    `json:"session_timeout_minutes"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

// Execute retrieves a project by ID
func (uc *GetProjectUseCase) Execute(ctx context.Context, req GetProjectRequest) (*GetProjectResponse, error) {
	// Find project
	foundProject, err := uc.projectRepo.FindByID(ctx, req.ProjectID)
	if err != nil {
		if err == project.ErrProjectNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// Return response
	return &GetProjectResponse{
		ProjectID:        foundProject.ID(),
		CustomerID:       foundProject.CustomerID(),
		BillingAccountID: foundProject.BillingAccountID(),
		TenantID:         foundProject.TenantID(),
		Name:             foundProject.Name(),
		Description:      foundProject.Description(),
		Configuration:    foundProject.Configuration(),
		IsActive:         foundProject.IsActive(),
		SessionTimeout:   foundProject.GetSessionTimeout(),
		CreatedAt:        foundProject.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        foundProject.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetProjectByTenantUseCase handles project retrieval by tenant ID
type GetProjectByTenantUseCase struct {
	projectRepo project.Repository
}

// NewGetProjectByTenantUseCase creates a new instance
func NewGetProjectByTenantUseCase(projectRepo project.Repository) *GetProjectByTenantUseCase {
	return &GetProjectByTenantUseCase{
		projectRepo: projectRepo,
	}
}

// Execute retrieves a project by tenant ID
func (uc *GetProjectByTenantUseCase) Execute(ctx context.Context, req GetProjectByTenantRequest) (*GetProjectResponse, error) {
	// Find project
	foundProject, err := uc.projectRepo.FindByTenantID(ctx, req.TenantID)
	if err != nil {
		if err == project.ErrProjectNotFound {
			return nil, fmt.Errorf("project not found for tenant %s", req.TenantID)
		}
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	// Return response
	return &GetProjectResponse{
		ProjectID:        foundProject.ID(),
		CustomerID:       foundProject.CustomerID(),
		BillingAccountID: foundProject.BillingAccountID(),
		TenantID:         foundProject.TenantID(),
		Name:             foundProject.Name(),
		Description:      foundProject.Description(),
		Configuration:    foundProject.Configuration(),
		IsActive:         foundProject.IsActive(),
		SessionTimeout:   foundProject.GetSessionTimeout(),
		CreatedAt:        foundProject.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        foundProject.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ListProjectsUseCase handles listing projects
type ListProjectsUseCase struct {
	projectRepo project.Repository
}

// NewListProjectsUseCase creates a new instance
func NewListProjectsUseCase(projectRepo project.Repository) *ListProjectsUseCase {
	return &ListProjectsUseCase{
		projectRepo: projectRepo,
	}
}

// ListProjectsRequest represents the request to list projects
type ListProjectsRequest struct {
	CustomerID uuid.UUID `json:"customer_id,omitempty"`
	ActiveOnly bool      `json:"active_only"`
	Limit      int       `json:"limit" validate:"min=1,max=100"`
	Offset     int       `json:"offset" validate:"min=0"`
}

// ListProjectsResponse represents the response with list of projects
type ListProjectsResponse struct {
	Projects []GetProjectResponse `json:"projects"`
	Total    int                  `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

// Execute lists projects
func (uc *ListProjectsUseCase) Execute(ctx context.Context, req ListProjectsRequest) (*ListProjectsResponse, error) {
	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 20
	}

	// TODO: Implement FindByCustomerID and FindActiveProjects methods in repository
	return nil, fmt.Errorf("listing projects not implemented yet - repository methods missing")
}
