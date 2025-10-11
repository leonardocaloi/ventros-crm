package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, event *Event) error

	FindByID(ctx context.Context, id uuid.UUID) (*Event, error)

	FindByContact(ctx context.Context, contactID uuid.UUID, limit int) ([]*Event, error)

	FindBySession(ctx context.Context, sessionID uuid.UUID) ([]*Event, error)

	FindByTenantAndType(ctx context.Context, tenantID, eventType string, limit int) ([]*Event, error)

	FindByTimeRange(ctx context.Context, tenantID string, start, end time.Time) ([]*Event, error)

	CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)

	CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error)
}
