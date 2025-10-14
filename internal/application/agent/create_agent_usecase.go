package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/agent"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}

// TransactionManager gerencia transações de banco de dados.
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CreateAgentUseCase handles agent creation
type CreateAgentUseCase struct {
	agentRepo agent.Repository
	eventBus  EventBus
	txManager TransactionManager
}

// NewCreateAgentUseCase creates a new instance
func NewCreateAgentUseCase(agentRepo agent.Repository, eventBus EventBus, txManager TransactionManager) *CreateAgentUseCase {
	return &CreateAgentUseCase{
		agentRepo: agentRepo,
		eventBus:  eventBus,
		txManager: txManager,
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

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Save to repository (usa transação do contexto)
		if err := uc.agentRepo.Save(txCtx, newAgent); err != nil {
			return fmt.Errorf("failed to save agent: %w", err)
		}

		// 2. Publish domain events (usa mesma transação)
		events := newAgent.DomainEvents()
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
