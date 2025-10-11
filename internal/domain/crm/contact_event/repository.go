package contact_event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, event *ContactEvent) error

	Update(ctx context.Context, event *ContactEvent) error

	FindByID(ctx context.Context, id uuid.UUID) (*ContactEvent, error)

	FindByContactID(ctx context.Context, contactID uuid.UUID, limit int, offset int) ([]*ContactEvent, error)

	FindByContactIDVisible(ctx context.Context, contactID uuid.UUID, visibleToClient bool, limit int, offset int) ([]*ContactEvent, error)

	FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, offset int) ([]*ContactEvent, error)

	FindUndeliveredRealtime(ctx context.Context, limit int) ([]*ContactEvent, error)

	FindUndeliveredForContact(ctx context.Context, contactID uuid.UUID) ([]*ContactEvent, error)

	FindByTenantAndType(ctx context.Context, tenantID string, eventType string, since time.Time, limit int) ([]*ContactEvent, error)

	FindByCategory(ctx context.Context, tenantID string, category Category, since time.Time, limit int) ([]*ContactEvent, error)

	FindExpired(ctx context.Context, before time.Time, limit int) ([]*ContactEvent, error)

	Delete(ctx context.Context, id uuid.UUID) error

	DeleteExpired(ctx context.Context, before time.Time) (int, error)

	CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)
}
