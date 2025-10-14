package queries

import (
	"context"
	"strings"

	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/message"
	"go.uber.org/zap"
)

// SearchMessagesQuery query to search messages by text
type SearchMessagesQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchMessagesResponse response for search messages query
type SearchMessagesResponse struct {
	Messages []MessageSearchResultDTO `json:"messages"`
	Count    int                      `json:"count"`
}

// MessageSearchResultDTO search result for message
type MessageSearchResultDTO struct {
	ID          string  `json:"id"`
	Timestamp   string  `json:"timestamp"`
	ContactID   string  `json:"contact_id"`
	FromMe      bool    `json:"from_me"`
	ContentType string  `json:"content_type"`
	Text        *string `json:"text,omitempty"`
	MatchScore  float64 `json:"match_score"`
}

// SearchMessagesQueryHandler handles SearchMessagesQuery
type SearchMessagesQueryHandler struct {
	messageRepo message.Repository
	logger      *zap.Logger
}

// NewSearchMessagesQueryHandler creates a new SearchMessagesQueryHandler
func NewSearchMessagesQueryHandler(messageRepo message.Repository, logger *zap.Logger) *SearchMessagesQueryHandler {
	return &SearchMessagesQueryHandler{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

// Handle executes the SearchMessagesQuery
func (h *SearchMessagesQueryHandler) Handle(ctx context.Context, query SearchMessagesQuery) (*SearchMessagesResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchMessagesResponse{
			Messages: []MessageSearchResultDTO{},
			Count:    0,
		}, nil
	}

	h.logger.Info("Searching messages",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search messages using repository
	messages, _, err := h.messageRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
		0,
	)
	if err != nil {
		h.logger.Error("Failed to search messages", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	results := make([]MessageSearchResultDTO, len(messages))
	for i, m := range messages {
		dto := MessageSearchResultDTO{
			ID:          m.ID().String(),
			Timestamp:   m.Timestamp().Format("2006-01-02T15:04:05Z07:00"),
			ContactID:   m.ContactID().String(),
			FromMe:      m.FromMe(),
			ContentType: m.ContentType().String(),
			Text:        m.Text(),
			MatchScore:  1.0,
		}

		results[i] = dto
	}

	return &SearchMessagesResponse{
		Messages: results,
		Count:    len(results),
	}, nil
}
