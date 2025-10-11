package message

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

// MessageFilters represents filtering options for message queries
type MessageFilters struct {
	TenantID       string
	ContactID      *uuid.UUID
	SessionID      *uuid.UUID
	ChannelID      *uuid.UUID
	ProjectID      *uuid.UUID
	ChannelTypeID  *int
	FromMe         *bool
	ContentType    *string
	Status         *string
	AgentID        *uuid.UUID
	TimestampAfter *time.Time
	TimestampBefore *time.Time
	HasMedia       *bool
	Limit          int
	Offset         int
	SortBy         string // timestamp, created_at
	SortOrder      string // asc, desc
}

type Repository interface {
	Save(ctx context.Context, message *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	FindBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*Message, error)
	FindByContact(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*Message, error)
	FindByChannelMessageID(ctx context.Context, channelMessageID string) (*Message, error)
	CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error)

	// Advanced query methods
	FindByTenantWithFilters(ctx context.Context, filters MessageFilters) ([]*Message, int64, error)

	SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Message, int64, error)
}
