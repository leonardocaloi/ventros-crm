package agent

import (
	"context"

	"github.com/google/uuid"
)

// AgentFilters represents filtering options for agent queries
type AgentFilters struct {
	TenantID  string
	ProjectID *uuid.UUID
	Type      *AgentType
	Status    *AgentStatus
	Active    *bool
	Limit     int
	Offset    int
	SortBy    string // name, created_at, last_activity_at
	SortOrder string // asc, desc
}

type Repository interface {
	Save(ctx context.Context, agent *Agent) error

	FindByID(ctx context.Context, id uuid.UUID) (*Agent, error)

	FindByEmail(ctx context.Context, tenantID, email string) (*Agent, error)

	FindByTenant(ctx context.Context, tenantID string) ([]*Agent, error)

	FindActiveByTenant(ctx context.Context, tenantID string) ([]*Agent, error)

	Delete(ctx context.Context, id uuid.UUID) error

	// Advanced query methods
	FindByTenantWithFilters(ctx context.Context, filters AgentFilters) ([]*Agent, int64, error)

	SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Agent, int64, error)
}
