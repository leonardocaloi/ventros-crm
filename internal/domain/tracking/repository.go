package tracking

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para persistência de Tracking.
type Repository interface {
	// Create cria um novo tracking
	Create(ctx context.Context, tracking *Tracking) error

	// FindByID busca um tracking por ID
	FindByID(ctx context.Context, id uuid.UUID) (*Tracking, error)

	// FindByContactID busca todos os trackings de um contato
	FindByContactID(ctx context.Context, contactID uuid.UUID) ([]*Tracking, error)

	// FindBySessionID busca todos os trackings de uma sessão
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*Tracking, error)

	// FindByProjectID busca todos os trackings de um projeto
	FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*Tracking, error)

	// FindBySource busca trackings por fonte
	FindBySource(ctx context.Context, projectID uuid.UUID, source Source, limit, offset int) ([]*Tracking, error)

	// FindByCampaign busca trackings por campanha
	FindByCampaign(ctx context.Context, projectID uuid.UUID, campaign string, limit, offset int) ([]*Tracking, error)

	// FindByClickID busca tracking por click ID (único)
	FindByClickID(ctx context.Context, clickID string) (*Tracking, error)

	// Update atualiza um tracking
	Update(ctx context.Context, tracking *Tracking) error

	// Delete remove um tracking (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error
}
