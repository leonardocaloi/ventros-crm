package project

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/core/project"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}

// TransactionManager gerencia transações de banco de dados.
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CreateProjectUseCase handles project creation
type CreateProjectUseCase struct {
	projectRepo project.Repository
	eventBus    EventBus
	txManager   TransactionManager
}

// NewCreateProjectUseCase creates a new instance
func NewCreateProjectUseCase(projectRepo project.Repository, eventBus EventBus, txManager TransactionManager) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		projectRepo: projectRepo,
		eventBus:    eventBus,
		txManager:   txManager,
	}
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	CustomerID       uuid.UUID              `json:"customer_id" validate:"required"`
	BillingAccountID uuid.UUID              `json:"billing_account_id" validate:"required"`
	TenantID         string                 `json:"tenant_id" validate:"required,min=3,max=50"`
	Name             string                 `json:"name" validate:"required,min=2,max=100"`
	Description      string                 `json:"description,omitempty"`
	Configuration    map[string]interface{} `json:"configuration,omitempty"`
	SessionTimeout   int                    `json:"session_timeout_minutes,omitempty" validate:"omitempty,min=5,max=480"`
}

// CreateProjectResponse represents the response after creating a project
type CreateProjectResponse struct {
	ProjectID        uuid.UUID              `json:"project_id"`
	CustomerID       uuid.UUID              `json:"customer_id"`
	BillingAccountID uuid.UUID              `json:"billing_account_id"`
	TenantID         string                 `json:"tenant_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Configuration    map[string]interface{} `json:"configuration"`
	IsActive         bool                   `json:"is_active"`
	CreatedAt        string                 `json:"created_at"`
}

// Execute creates a new project
func (uc *CreateProjectUseCase) Execute(ctx context.Context, req CreateProjectRequest) (*CreateProjectResponse, error) {
	// Check if project with tenant ID already exists
	existingProject, err := uc.projectRepo.FindByTenantID(ctx, req.TenantID)
	if err == nil && existingProject != nil {
		return nil, fmt.Errorf("project with tenant ID %s already exists", req.TenantID)
	}
	if err != nil && err != project.ErrProjectNotFound {
		return nil, fmt.Errorf("failed to check existing project: %w", err)
	}

	// Create new project
	newProject, err := project.NewProject(
		req.CustomerID,
		req.BillingAccountID,
		req.TenantID,
		req.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Set optional fields
	if req.Description != "" {
		newProject.UpdateDescription(req.Description)
	}

	if req.Configuration != nil {
		newProject.UpdateConfiguration(req.Configuration)
	}

	// Set session timeout if provided
	if req.SessionTimeout > 0 {
		newProject.SetSessionTimeout(req.SessionTimeout)
	}

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Save to repository (usa transação do contexto)
		if err := uc.projectRepo.Save(txCtx, newProject); err != nil {
			return fmt.Errorf("failed to save project: %w", err)
		}

		// Publish domain events (usa mesma transação)
		events := newProject.DomainEvents()
		for _, event := range events {
			if err := uc.eventBus.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	newProject.ClearEvents()

	// Return response
	return &CreateProjectResponse{
		ProjectID:        newProject.ID(),
		CustomerID:       newProject.CustomerID(),
		BillingAccountID: newProject.BillingAccountID(),
		TenantID:         newProject.TenantID(),
		Name:             newProject.Name(),
		Description:      newProject.Description(),
		Configuration:    newProject.Configuration(),
		IsActive:         newProject.IsActive(),
		CreatedAt:        newProject.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
