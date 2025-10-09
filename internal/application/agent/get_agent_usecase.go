package agent

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/agent"
	"github.com/google/uuid"
)

// GetAgentUseCase handles agent retrieval
type GetAgentUseCase struct {
	agentRepo agent.Repository
}

// NewGetAgentUseCase creates a new instance
func NewGetAgentUseCase(agentRepo agent.Repository) *GetAgentUseCase {
	return &GetAgentUseCase{
		agentRepo: agentRepo,
	}
}

// GetAgentRequest represents the request to get an agent
type GetAgentRequest struct {
	AgentID uuid.UUID `json:"agent_id" validate:"required"`
}

// GetAgentResponse represents the response with agent details
type GetAgentResponse struct {
	AgentID   uuid.UUID         `json:"agent_id"`
	TenantID  string            `json:"tenant_id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	AgentType agent.AgentType   `json:"agent_type"`
	Status    agent.AgentStatus `json:"status"`
	IsActive  bool              `json:"is_active"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// Execute retrieves an agent by ID
func (uc *GetAgentUseCase) Execute(ctx context.Context, req GetAgentRequest) (*GetAgentResponse, error) {
	// Find agent
	foundAgent, err := uc.agentRepo.FindByID(ctx, req.AgentID)
	if err != nil {
		if err == agent.ErrAgentNotFound {
			return nil, fmt.Errorf("agent not found")
		}
		return nil, fmt.Errorf("failed to find agent: %w", err)
	}

	// Return response
	return &GetAgentResponse{
		AgentID:   foundAgent.ID(),
		TenantID:  foundAgent.TenantID(),
		Name:      foundAgent.Name(),
		Email:     foundAgent.Email(),
		AgentType: foundAgent.Type(),
		Status:    foundAgent.Status(),
		IsActive:  foundAgent.IsActive(),
		CreatedAt: foundAgent.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: foundAgent.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ListAgentsUseCase handles listing agents
type ListAgentsUseCase struct {
	agentRepo agent.Repository
}

// NewListAgentsUseCase creates a new instance
func NewListAgentsUseCase(agentRepo agent.Repository) *ListAgentsUseCase {
	return &ListAgentsUseCase{
		agentRepo: agentRepo,
	}
}

// ListAgentsRequest represents the request to list agents
type ListAgentsRequest struct {
	TenantID   string `json:"tenant_id" validate:"required"`
	ActiveOnly bool   `json:"active_only"`
	Limit      int    `json:"limit" validate:"min=1,max=100"`
	Offset     int    `json:"offset" validate:"min=0"`
}

// ListAgentsResponse represents the response with list of agents
type ListAgentsResponse struct {
	Agents []GetAgentResponse `json:"agents"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

// Execute lists agents for a tenant
func (uc *ListAgentsUseCase) Execute(ctx context.Context, req ListAgentsRequest) (*ListAgentsResponse, error) {
	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 20
	}

	// Find agents
	var agents []*agent.Agent
	var err error

	if req.ActiveOnly {
		agents, err = uc.agentRepo.FindActiveByTenant(ctx, req.TenantID)
	} else {
		agents, err = uc.agentRepo.FindByTenant(ctx, req.TenantID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find agents: %w", err)
	}

	// Convert to response format
	agentResponses := make([]GetAgentResponse, len(agents))
	for i, ag := range agents {
		agentResponses[i] = GetAgentResponse{
			AgentID:   ag.ID(),
			TenantID:  ag.TenantID(),
			Name:      ag.Name(),
			Email:     ag.Email(),
			AgentType: ag.Type(),
			Status:    ag.Status(),
			IsActive:  ag.IsActive(),
			CreatedAt: ag.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: ag.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &ListAgentsResponse{
		Agents: agentResponses,
		Total:  len(agentResponses), // TODO: Implement proper count query
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}
