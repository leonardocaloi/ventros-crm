package agent

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/agent"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// CreateAgentUseCase handles agent creation
type CreateAgentUseCase struct {
	agentRepo agent.Repository
	eventBus  EventBus
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}

// NewCreateAgentUseCase creates a new instance
func NewCreateAgentUseCase(agentRepo agent.Repository, eventBus EventBus) *CreateAgentUseCase {
	return &CreateAgentUseCase{
		agentRepo: agentRepo,
		eventBus:  eventBus,
	}
}

// CreateAgentRequest represents the request to create an agent
type CreateAgentRequest struct {
	ProjectID uuid.UUID       `json:"project_id" validate:"required"`
	TenantID  string          `json:"tenant_id" validate:"required"`
	Name      string          `json:"name" validate:"required,min=2,max=100"`
	Email     string          `json:"email" validate:"required,email"`
	AgentType agent.AgentType `json:"agent_type" validate:"required"`
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
	IsActive  bool            `json:"is_active"`
}

// CreateAgentResponse represents the response after creating an agent
type CreateAgentResponse struct {
	AgentID   uuid.UUID `json:"agent_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsActive  bool      `json:"is_active"`
	CreatedAt string    `json:"created_at"`
}

// Execute creates a new agent
func (uc *CreateAgentUseCase) Execute(ctx context.Context, req CreateAgentRequest) (*CreateAgentResponse, error) {
	// Check if agent with email already exists
	existingAgent, err := uc.agentRepo.FindByEmail(ctx, req.TenantID, req.Email)
	if err == nil && existingAgent != nil {
		return nil, fmt.Errorf("agent with email %s already exists", req.Email)
	}
	if err != nil && err != agent.ErrAgentNotFound {
		return nil, fmt.Errorf("failed to check existing agent: %w", err)
	}

	// Create new agent
	newAgent, err := agent.NewAgent(
		req.ProjectID,
		req.TenantID,
		req.Name,
		req.AgentType,
		req.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// TODO: Add SetEmail method to Agent domain or modify NewAgent to accept email

	// Set active status
	if req.IsActive {
		if err := newAgent.Activate(); err != nil {
			return nil, fmt.Errorf("failed to activate agent: %w", err)
		}
	}

	// Save to repository
	if err := uc.agentRepo.Save(ctx, newAgent); err != nil {
		return nil, fmt.Errorf("failed to save agent: %w", err)
	}

	// Publish domain events
	events := newAgent.DomainEvents()
	for _, event := range events {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the use case
			// Events are not critical for agent creation
		}
	}
	newAgent.ClearEvents()

	// Return response
	return &CreateAgentResponse{
		AgentID:   newAgent.ID(),
		Name:      newAgent.Name(),
		Email:     newAgent.Email(),
		IsActive:  newAgent.IsActive(),
		CreatedAt: newAgent.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
