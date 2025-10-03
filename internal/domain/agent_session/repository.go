package agent_session

import (
	"context"

	"github.com/google/uuid"
)

// Repository define o contrato para persistência de AgentSession.
type Repository interface {
	// Create persiste um novo AgentSession.
	Create(ctx context.Context, agentSession *AgentSession) error

	// Update atualiza um AgentSession existente.
	Update(ctx context.Context, agentSession *AgentSession) error

	// FindByID busca por ID.
	FindByID(ctx context.Context, id uuid.UUID) (*AgentSession, error)

	// FindActiveBySessionID busca agentes ativos em uma sessão.
	FindActiveBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*AgentSession, error)

	// FindByAgentID busca todas as participações de um agente.
	FindByAgentID(ctx context.Context, agentID uuid.UUID) ([]*AgentSession, error)

	// FindByAgentAndSession busca uma participação específica.
	FindByAgentAndSession(ctx context.Context, agentID, sessionID uuid.UUID) (*AgentSession, error)

	// Delete remove um AgentSession.
	Delete(ctx context.Context, id uuid.UUID) error
}
