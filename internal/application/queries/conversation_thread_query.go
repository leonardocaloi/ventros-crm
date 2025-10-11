package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConversationThreadQuery query to get full conversation thread for a contact
type ConversationThreadQuery struct {
	ContactID uuid.UUID
	TenantID  shared.TenantID
	ChannelID string // optional filter by channel
	Limit     int    // limit messages per session
}

// ConversationThreadResponse response for conversation thread
type ConversationThreadResponse struct {
	Sessions []ConversationSessionDTO `json:"sessions"`
	Count    int                      `json:"count"`
}

// ConversationSessionDTO session with messages in conversation thread
type ConversationSessionDTO struct {
	SessionID   string                 `json:"session_id"`
	ChannelID   string                 `json:"channel_id"`
	ChannelName string                 `json:"channel_name,omitempty"`
	AgentID     *string                `json:"agent_id,omitempty"`
	AgentName   *string                `json:"agent_name,omitempty"`
	Status      string                 `json:"status"`
	StartedAt   string                 `json:"started_at"`
	ClosedAt    *string                `json:"closed_at,omitempty"`
	Messages    []ConversationMessage  `json:"messages"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationMessage message in conversation thread
type ConversationMessage struct {
	ID          string                 `json:"id"`
	Direction   string                 `json:"direction"`
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	MediaURL    *string                `json:"media_url,omitempty"`
	Status      string                 `json:"status"`
	SenderName  string                 `json:"sender_name,omitempty"`
	SentAt      string                 `json:"sent_at"`
	ReadAt      *string                `json:"read_at,omitempty"`
	TrackingID  *string                `json:"tracking_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationThreadQueryHandler handles ConversationThreadQuery
type ConversationThreadQueryHandler struct {
	messageRepo message.Repository
	logger      *zap.Logger
}

// NewConversationThreadQueryHandler creates a new ConversationThreadQueryHandler
func NewConversationThreadQueryHandler(messageRepo message.Repository, logger *zap.Logger) *ConversationThreadQueryHandler {
	return &ConversationThreadQueryHandler{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// Handle executes the ConversationThreadQuery
func (h *ConversationThreadQueryHandler) Handle(ctx context.Context, query ConversationThreadQuery) (*ConversationThreadResponse, error) {
	// TODO: Implement repository method to get conversation thread
	// This should:
	// 1. Get all sessions for contact (optionally filtered by channel)
	// 2. Get messages for each session (limited)
	// 3. Join with agent/channel information
	// 4. Order by session start time DESC

	h.logger.Info("Getting conversation thread",
		zap.String("contact_id", query.ContactID.String()),
		zap.String("channel_id", query.ChannelID),
		zap.Int("limit", query.Limit))

	// Placeholder - needs proper repository implementation with joins
	sessions := []ConversationSessionDTO{}

	return &ConversationThreadResponse{
		Sessions: sessions,
		Count:    len(sessions),
	}, nil
}
