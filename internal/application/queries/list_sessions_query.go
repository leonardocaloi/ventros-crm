package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ListSessionsQuery query to list sessions with filters, pagination, and sorting
type ListSessionsQuery struct {
	TenantID      shared.TenantID
	ContactID     *uuid.UUID
	PipelineID    *uuid.UUID
	ChannelTypeID *int
	Status        *string
	Resolved      *bool
	Escalated     *bool
	Converted     *bool
	Sentiment     *string
	Page          int
	Limit         int
	SortBy        string
	SortDir       string
}

// ListSessionsResponse response for list sessions query
type ListSessionsResponse struct {
	Sessions   []SessionDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// SessionDTO data transfer object for session
type SessionDTO struct {
	ID                  string   `json:"id"`
	ContactID           string   `json:"contact_id"`
	PipelineID          *string  `json:"pipeline_id,omitempty"`
	Status              string   `json:"status"`
	StartedAt           string   `json:"started_at"`
	EndedAt             *string  `json:"ended_at,omitempty"`
	MessageCount        int      `json:"message_count"`
	MessagesFromContact int      `json:"messages_from_contact"`
	MessagesFromAgent   int      `json:"messages_from_agent"`
	DurationSeconds     int      `json:"duration_seconds"`
	Summary             *string  `json:"summary,omitempty"`
	Sentiment           *string  `json:"sentiment,omitempty"`
	Topics              []string `json:"topics,omitempty"`
	Resolved            bool     `json:"resolved"`
	Escalated           bool     `json:"escalated"`
	Converted           bool     `json:"converted"`
	OutcomeTags         []string `json:"outcome_tags,omitempty"`
}

// ListSessionsQueryHandler handles ListSessionsQuery
type ListSessionsQueryHandler struct {
	sessionRepo session.Repository
	logger      *zap.Logger
}

// NewListSessionsQueryHandler creates a new ListSessionsQueryHandler
func NewListSessionsQueryHandler(sessionRepo session.Repository, logger *zap.Logger) *ListSessionsQueryHandler {
	return &ListSessionsQueryHandler{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

// Handle executes the ListSessionsQuery
func (h *ListSessionsQueryHandler) Handle(ctx context.Context, query ListSessionsQuery) (*ListSessionsResponse, error) {
	h.logger.Info("Listing sessions",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Build filters
	filters := session.SessionFilters{
		TenantID:      query.TenantID.String(),
		ContactID:     query.ContactID,
		PipelineID:    query.PipelineID,
		ChannelTypeID: query.ChannelTypeID,
		Status:        query.Status,
		Resolved:      query.Resolved,
		Escalated:     query.Escalated,
		Converted:     query.Converted,
		Sentiment:     query.Sentiment,
		Limit:         query.Limit,
		Offset:        (query.Page - 1) * query.Limit,
		SortBy:        query.SortBy,
		SortOrder:     query.SortDir,
	}

	// Fetch sessions from repository
	sessions, totalCount, err := h.sessionRepo.FindByTenantWithFilters(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list sessions", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	sessionDTOs := make([]SessionDTO, len(sessions))
	for i, s := range sessions {
		dto := SessionDTO{
			ID:                  s.ID().String(),
			ContactID:           s.ContactID().String(),
			Status:              s.Status().String(),
			StartedAt:           s.StartedAt().Format("2006-01-02T15:04:05Z07:00"),
			MessageCount:        s.MessageCount(),
			MessagesFromContact: s.MessagesFromContact(),
			MessagesFromAgent:   s.MessagesFromAgent(),
			DurationSeconds:     s.DurationSeconds(),
			Summary:             s.Summary(),
			Topics:              s.Topics(),
			Resolved:            s.IsResolved(),
			Escalated:           s.IsEscalated(),
			Converted:           s.IsConverted(),
			OutcomeTags:         s.OutcomeTags(),
		}

		if pipelineID := s.PipelineID(); pipelineID != nil {
			pipelineStr := pipelineID.String()
			dto.PipelineID = &pipelineStr
		}

		if endedAt := s.EndedAt(); endedAt != nil {
			endedAtStr := endedAt.Format("2006-01-02T15:04:05Z07:00")
			dto.EndedAt = &endedAtStr
		}

		if sentiment := s.Sentiment(); sentiment != nil {
			sentimentStr := sentiment.String()
			dto.Sentiment = &sentimentStr
		}

		sessionDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListSessionsResponse{
		Sessions:   sessionDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
