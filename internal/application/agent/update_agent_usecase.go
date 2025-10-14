package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/agent"
)

// UpdateAgentUseCase handles agent updates
type UpdateAgentUseCase struct {
	agentRepo agent.Repository
	eventBus  EventBus
}

// NewUpdateAgentUseCase creates a new instance
func NewUpdateAgentUseCase(agentRepo agent.Repository, eventBus EventBus) *UpdateAgentUseCase {
	return &UpdateAgentUseCase{
		agentRepo: agentRepo,
		eventBus:  eventBus,
	}
}

// UpdateAgentRequest represents the request to update an agent
type UpdateAgentRequest struct {
	AgentID     uuid.UUID `json:"agent_id" validate:"required"`
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone       *string   `json:"phone,omitempty"`
	Department  *string   `json:"department,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
	MaxSessions *int      `json:"max_sessions,omitempty" validate:"omitempty,min=1,max=50"`
}

// UpdateAgentResponse represents the response after updating an agent
type UpdateAgentResponse struct {
	AgentID     uuid.UUID `json:"agent_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Department  string    `json:"department"`
	IsActive    bool      `json:"is_active"`
	MaxSessions int       `json:"max_sessions"`
	UpdatedAt   string    `json:"updated_at"`
}

// Execute updates an existing agent
func (uc *UpdateAgentUseCase) Execute(ctx context.Context, req UpdateAgentRequest) (*UpdateAgentResponse, error) {
	return nil, fmt.Errorf("agent update not implemented yet - domain methods missing")
}
