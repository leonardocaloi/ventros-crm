package tracking

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, tracking *Tracking) error

	FindByID(ctx context.Context, id uuid.UUID) (*Tracking, error)

	FindByContactID(ctx context.Context, contactID uuid.UUID) ([]*Tracking, error)

	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*Tracking, error)

	FindByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*Tracking, error)

	FindBySource(ctx context.Context, projectID uuid.UUID, source Source, limit, offset int) ([]*Tracking, error)

	FindByCampaign(ctx context.Context, projectID uuid.UUID, campaign string, limit, offset int) ([]*Tracking, error)

	FindByClickID(ctx context.Context, clickID string) (*Tracking, error)

	Update(ctx context.Context, tracking *Tracking) error

	Delete(ctx context.Context, id uuid.UUID) error
}
