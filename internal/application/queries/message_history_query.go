package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MessageHistoryQuery query to get message history for a session
type MessageHistoryQuery struct {
	SessionID uuid.UUID
	TenantID  shared.TenantID
	Page      int
	Limit     int
	Direction string // "asc" or "desc"
}

// MessageHistoryResponse response for message history
type MessageHistoryResponse struct {
	Messages   []MessageDTO `json:"messages"`
	TotalCount int64        `json:"total_count"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
}

// MessageDTO data transfer object for message
type MessageDTO struct {
	ID           string                 `json:"id"`
	SessionID    string                 `json:"session_id"`
	ContactID    string                 `json:"contact_id"`
	Direction    string                 `json:"direction"` // "inbound" or "outbound"
	Content      string                 `json:"content"`
	ContentType  string                 `json:"content_type"`
	MediaURL     *string                `json:"media_url,omitempty"`
	Status       string                 `json:"status"`
	SentAt       *string                `json:"sent_at,omitempty"`
	DeliveredAt  *string                `json:"delivered_at,omitempty"`
	ReadAt       *string                `json:"read_at,omitempty"`
	FailedAt     *string                `json:"failed_at,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	TrackingID   *string                `json:"tracking_id,omitempty"`
	CreatedAt    string                 `json:"created_at"`
}

// MessageHistoryQueryHandler handles MessageHistoryQuery
type MessageHistoryQueryHandler struct {
	messageRepo message.Repository
	logger      *zap.Logger
}

// NewMessageHistoryQueryHandler creates a new MessageHistoryQueryHandler
func NewMessageHistoryQueryHandler(messageRepo message.Repository, logger *zap.Logger) *MessageHistoryQueryHandler {
	return &MessageHistoryQueryHandler{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// Handle executes the MessageHistoryQuery
func (h *MessageHistoryQueryHandler) Handle(ctx context.Context, query MessageHistoryQuery) (*MessageHistoryResponse, error) {
	// TODO: Implement repository method FindBySessionWithPagination
	// For now, using basic FindBySession if it exists

	h.logger.Info("Getting message history",
		zap.String("session_id", query.SessionID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Placeholder - needs proper repository implementation
	messages := []MessageDTO{}

	return &MessageHistoryResponse{
		Messages:   messages,
		TotalCount: 0,
		Page:       query.Page,
		Limit:      query.Limit,
	}, nil
}
