package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// SessionFilters represents filtering options for session queries
type SessionFilters struct {
	TenantID      string
	ContactID     *uuid.UUID
	PipelineID    *uuid.UUID
	ChannelTypeID *int
	Status        *string
	Resolved      *bool
	Escalated     *bool
	Converted     *bool
	Sentiment     *string
	StartedAfter  *time.Time
	StartedBefore *time.Time
	EndedAfter    *time.Time
	EndedBefore   *time.Time
	MinMessages   *int
	MaxMessages   *int
	Limit         int
	Offset        int
	SortBy        string // started_at, ended_at, message_count, duration_seconds
	SortOrder     string // asc, desc
}

type Repository interface {
	Save(ctx context.Context, session *Session) error

	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)

	FindActiveByContact(ctx context.Context, contactID uuid.UUID, channelTypeID *int) (*Session, error)

	FindInactiveSessions(ctx context.Context, tenantID string) ([]*Session, error)

	FindSessionsRequiringSummary(ctx context.Context, tenantID string, limit int) ([]*Session, error)

	CountActiveByTenant(ctx context.Context, tenantID string) (int, error)

	FindActiveBeforeTime(ctx context.Context, cutoffTime time.Time) ([]*Session, error)

	// Advanced query methods
	FindByTenantWithFilters(ctx context.Context, filters SessionFilters) ([]*Session, int64, error)

	SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Session, int64, error)

	// Consolidation methods (for history import post-processing)
	FindByChannelPaginated(ctx context.Context, channelID uuid.UUID, limit int, offset int) ([]*Session, error)
	CountByChannel(ctx context.Context, channelID uuid.UUID) (int64, error)
	DeleteBatch(ctx context.Context, sessionIDs []uuid.UUID) error

	// 🔥 FIX Bug 1: Contact-based batching for consolidation
	GetContactIDsByChannel(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error)
	FindByChannelAndContacts(ctx context.Context, channelID uuid.UUID, contactIDs []uuid.UUID) ([]*Session, error)
}
