package agent_session

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, agentSession *AgentSession) error

	Update(ctx context.Context, agentSession *AgentSession) error

	FindByID(ctx context.Context, id uuid.UUID) (*AgentSession, error)

	FindActiveBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*AgentSession, error)

	FindByAgentID(ctx context.Context, agentID uuid.UUID) ([]*AgentSession, error)

	FindByAgentAndSession(ctx context.Context, agentID, sessionID uuid.UUID) (*AgentSession, error)

	Delete(ctx context.Context, id uuid.UUID) error
}
