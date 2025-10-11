package queries

import (
	"context"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ListMessagesQuery query to list messages with filters, pagination, and sorting
type ListMessagesQuery struct {
	TenantID       shared.TenantID
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
	Page           int
	Limit          int
	SortBy         string
	SortDir        string
}

// ListMessagesResponse response for list messages query
type ListMessagesResponse struct {
	Messages   []ListMessageDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// ListMessageDTO data transfer object for message in list
type ListMessageDTO struct {
	ID               string                 `json:"id"`
	Timestamp        string                 `json:"timestamp"`
	ContactID        string                 `json:"contact_id"`
	SessionID        *string                `json:"session_id,omitempty"`
	ChannelID        string                 `json:"channel_id"`
	FromMe           bool                   `json:"from_me"`
	ContentType      string                 `json:"content_type"`
	Text             *string                `json:"text,omitempty"`
	MediaURL         *string                `json:"media_url,omitempty"`
	Status           string                 `json:"status"`
	AgentID          *string                `json:"agent_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ListMessagesQueryHandler handles ListMessagesQuery
type ListMessagesQueryHandler struct {
	messageRepo message.Repository
	logger      *zap.Logger
}

// NewListMessagesQueryHandler creates a new ListMessagesQueryHandler
func NewListMessagesQueryHandler(messageRepo message.Repository, logger *zap.Logger) *ListMessagesQueryHandler {
	return &ListMessagesQueryHandler{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// Handle executes the ListMessagesQuery
func (h *ListMessagesQueryHandler) Handle(ctx context.Context, query ListMessagesQuery) (*ListMessagesResponse, error) {
	h.logger.Info("Listing messages",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Build filters
	filters := message.MessageFilters{
		TenantID:        query.TenantID.String(),
		ContactID:       query.ContactID,
		SessionID:       query.SessionID,
		ChannelID:       query.ChannelID,
		ProjectID:       query.ProjectID,
		ChannelTypeID:   query.ChannelTypeID,
		FromMe:          query.FromMe,
		ContentType:     query.ContentType,
		Status:          query.Status,
		AgentID:         query.AgentID,
		TimestampAfter:  query.TimestampAfter,
		TimestampBefore: query.TimestampBefore,
		HasMedia:        query.HasMedia,
		Limit:           query.Limit,
		Offset:          (query.Page - 1) * query.Limit,
		SortBy:          query.SortBy,
		SortOrder:       query.SortDir,
	}

	// Fetch messages from repository
	messages, totalCount, err := h.messageRepo.FindByTenantWithFilters(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list messages", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	messageDTOs := make([]ListMessageDTO, len(messages))
	for i, m := range messages {
		dto := ListMessageDTO{
			ID:          m.ID().String(),
			Timestamp:   m.Timestamp().Format("2006-01-02T15:04:05Z07:00"),
			ContactID:   m.ContactID().String(),
			ChannelID:   m.ChannelID().String(),
			FromMe:      m.FromMe(),
			ContentType: m.ContentType().String(),
			Text:        m.Text(),
			MediaURL:    m.MediaURL(),
			Status:      m.Status().String(),
			Metadata:    m.Metadata(),
		}

		if sessionID := m.SessionID(); sessionID != nil {
			sessionStr := sessionID.String()
			dto.SessionID = &sessionStr
		}

		if agentID := m.AgentID(); agentID != nil {
			agentStr := agentID.String()
			dto.AgentID = &agentStr
		}

		messageDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListMessagesResponse{
		Messages:   messageDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
