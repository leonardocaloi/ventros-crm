package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SessionHistoryQuery query to get session history for a contact
type SessionHistoryQuery struct {
	ContactID uuid.UUID
	TenantID  shared.TenantID
	Page      int
	Limit     int
}

// SessionHistoryResponse response for session history
type SessionHistoryResponse struct {
	Sessions   []SessionHistoryDTO `json:"sessions"`
	TotalCount int64               `json:"total_count"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
}

// SessionHistoryDTO data transfer object for session history
type SessionHistoryDTO struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	ChannelID       string                 `json:"channel_id"`
	ChannelName     string                 `json:"channel_name,omitempty"`
	AgentID         *string                `json:"agent_id,omitempty"`
	AgentName       *string                `json:"agent_name,omitempty"`
	StartedAt       string                 `json:"started_at"`
	ClosedAt        *string                `json:"closed_at,omitempty"`
	Duration        string                 `json:"duration"`
	MessageCount    int                    `json:"message_count"`
	UnreadCount     int                    `json:"unread_count"`
	LastMessageText *string                `json:"last_message_text,omitempty"`
	LastMessageAt   *string                `json:"last_message_at,omitempty"`
	CustomFields    map[string]interface{} `json:"custom_fields,omitempty"`
	Notes           []string               `json:"notes,omitempty"`
}

// SessionHistoryQueryHandler handles SessionHistoryQuery
type SessionHistoryQueryHandler struct {
	sessionRepo session.Repository
	logger      *zap.Logger
}

// NewSessionHistoryQueryHandler creates a new SessionHistoryQueryHandler
func NewSessionHistoryQueryHandler(sessionRepo session.Repository, logger *zap.Logger) *SessionHistoryQueryHandler {
	return &SessionHistoryQueryHandler{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// Handle executes the SessionHistoryQuery
func (h *SessionHistoryQueryHandler) Handle(ctx context.Context, query SessionHistoryQuery) (*SessionHistoryResponse, error) {
	// TODO: Implement repository method FindByContactWithPagination
	// For now, using basic FindByContact if it exists

	h.logger.Info("Getting session history",
		zap.String("contact_id", query.ContactID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Placeholder - needs proper repository implementation
	sessions := []SessionHistoryDTO{}

	return &SessionHistoryResponse{
		Sessions:   sessions,
		TotalCount: 0,
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}
