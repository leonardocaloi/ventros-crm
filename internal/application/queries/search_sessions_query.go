package queries

import (
	"context"
	"strings"

	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.uber.org/zap"
)

// SearchSessionsQuery query to search sessions by text
type SearchSessionsQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchSessionsResponse response for search sessions query
type SearchSessionsResponse struct {
	Sessions []SessionSearchResultDTO `json:"sessions"`
	Count    int                      `json:"count"`
}

// SessionSearchResultDTO search result for session
type SessionSearchResultDTO struct {
	ID           string   `json:"id"`
	ContactID    string   `json:"contact_id"`
	Status       string   `json:"status"`
	StartedAt    string   `json:"started_at"`
	MessageCount int      `json:"message_count"`
	Summary      *string  `json:"summary,omitempty"`
	Topics       []string `json:"topics,omitempty"`
	MatchScore   float64  `json:"match_score"`
	MatchField   string   `json:"match_field"`
}

// SearchSessionsQueryHandler handles SearchSessionsQuery
type SearchSessionsQueryHandler struct {
	sessionRepo session.Repository
	logger      *zap.Logger
}

// NewSearchSessionsQueryHandler creates a new SearchSessionsQueryHandler
func NewSearchSessionsQueryHandler(sessionRepo session.Repository, logger *zap.Logger) *SearchSessionsQueryHandler {
	return &SearchSessionsQueryHandler{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// Handle executes the SearchSessionsQuery
func (h *SearchSessionsQueryHandler) Handle(ctx context.Context, query SearchSessionsQuery) (*SearchSessionsResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchSessionsResponse{
			Sessions: []SessionSearchResultDTO{},
			Count:    0,
		}, nil
	}

	h.logger.Info("Searching sessions",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search sessions using repository
	sessions, _, err := h.sessionRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
		0,
	)
	if err != nil {
		h.logger.Error("Failed to search sessions", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs with match scoring
	results := make([]SessionSearchResultDTO, len(sessions))
	for i, s := range sessions {
		dto := SessionSearchResultDTO{
			ID:           s.ID().String(),
			ContactID:    s.ContactID().String(),
			Status:       s.Status().String(),
			StartedAt:    s.StartedAt().Format("2006-01-02T15:04:05Z07:00"),
			MessageCount: s.MessageCount(),
			Summary:      s.Summary(),
			Topics:       s.Topics(),
			MatchScore:   1.0,
			MatchField:   "summary",
		}

		// Determine match field based on what was matched
		if s.Summary() != nil && strings.Contains(strings.ToLower(*s.Summary()), searchText) {
			dto.MatchScore = 1.5
			dto.MatchField = "summary"
		} else if len(s.Topics()) > 0 {
			for _, topic := range s.Topics() {
				if strings.Contains(strings.ToLower(topic), searchText) {
					dto.MatchScore = 1.3
					dto.MatchField = "topics"
					break
				}
			}
		}

		results[i] = dto
	}

	return &SearchSessionsResponse{
		Sessions: results,
		Count:    len(results),
	}, nil
}
