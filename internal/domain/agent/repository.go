package agent

import (
	"context"

	"github.com/google/uuid"
)

// Repository define as operações de persistência para Agent.
type Repository interface {
	// Save persiste um agente (create ou update).
	Save(ctx context.Context, agent *Agent) error

	// FindByID busca um agente por ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Agent, error)

	// FindByEmail busca um agente por email dentro de um tenant.
	FindByEmail(ctx context.Context, tenantID, email string) (*Agent, error)

	// FindByTenant lista todos os agentes de um tenant.
	FindByTenant(ctx context.Context, tenantID string) ([]*Agent, error)

	// FindActiveByTenant lista agentes ativos de um tenant.
	FindActiveByTenant(ctx context.Context, tenantID string) ([]*Agent, error)

	// Delete remove um agente (soft delete).
	Delete(ctx context.Context, id uuid.UUID) error
}
