package queries

import (
	"context"

	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.uber.org/zap"
)

// GetActiveSessionsQuery query to get active sessions
type GetActiveSessionsQuery struct {
	TenantID  shared.TenantID
	ChannelID string
	AgentID   string
	Limit     int
}

// GetActiveSessionsResponse response for active sessions query
type GetActiveSessionsResponse struct {
	Sessions []ActiveSessionDTO `json:"sessions"`
	Count    int                `json:"count"`
}

// ActiveSessionDTO data transfer object for active session
type ActiveSessionDTO struct {
	ID              string  `json:"id"`
	ContactID       string  `json:"contact_id"`
	ContactName     string  `json:"contact_name"`
	ContactPhone    string  `json:"contact_phone"`
	ChannelID       string  `json:"channel_id"`
	AgentID         *string `json:"agent_id,omitempty"`
	AgentName       *string `json:"agent_name,omitempty"`
	Status          string  `json:"status"`
	StartedAt       string  `json:"started_at"`
	LastMessageAt   *string `json:"last_message_at,omitempty"`
	MessageCount    int     `json:"message_count"`
	UnreadCount     int     `json:"unread_count"`
	WaitingTime     string  `json:"waiting_time"`
	SessionDuration string  `json:"session_duration"`
}

// GetActiveSessionsQueryHandler handles GetActiveSessionsQuery
type GetActiveSessionsQueryHandler struct {
	sessionRepo session.Repository
	logger      *zap.Logger
}

// NewGetActiveSessionsQueryHandler creates a new GetActiveSessionsQueryHandler
func NewGetActiveSessionsQueryHandler(sessionRepo session.Repository, logger *zap.Logger) *GetActiveSessionsQueryHandler {
	return &GetActiveSessionsQueryHandler{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// Handle executes the GetActiveSessionsQuery
func (h *GetActiveSessionsQueryHandler) Handle(ctx context.Context, query GetActiveSessionsQuery) (*GetActiveSessionsResponse, error) {
	// TODO: Implement repository method FindActiveByTenant with filters
	// For now, using basic FindAll (needs implementation)

	h.logger.Info("Getting active sessions",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("channel_id", query.ChannelID),
		zap.String("agent_id", query.AgentID))

	// Placeholder - needs proper repository implementation
	sessions := []ActiveSessionDTO{}

	return &GetActiveSessionsResponse{
		Sessions: sessions,
		Count:    len(sessions),
	}, nil
}
