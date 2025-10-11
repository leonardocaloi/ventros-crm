package note

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NoteFilters represents filtering options for note queries
type NoteFilters struct {
	TenantID        string
	ContactID       *uuid.UUID
	SessionID       *uuid.UUID
	AuthorID        *uuid.UUID
	AuthorType      *string
	NoteType        *string
	Priority        *string
	VisibleToClient *bool
	Pinned          *bool
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
	Limit           int
	Offset          int
	SortBy          string // created_at, priority
	SortOrder       string // asc, desc
}

type Repository interface {
	Save(ctx context.Context, note *Note) error

	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)

	FindByContactID(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*Note, error)

	FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*Note, error)

	FindPinned(ctx context.Context, contactID uuid.UUID) ([]*Note, error)

	Delete(ctx context.Context, id uuid.UUID) error

	CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)

	// Advanced query methods
	FindByTenantWithFilters(ctx context.Context, filters NoteFilters) ([]*Note, int64, error)

	SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Note, int64, error)
}
