package queries

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/note"
	"go.uber.org/zap"
)

// ListNotesQuery query to list notes with filters, pagination, and sorting
type ListNotesQuery struct {
	TenantID        shared.TenantID
	ContactID       *uuid.UUID
	SessionID       *uuid.UUID
	AuthorID        *uuid.UUID
	AuthorType      *string
	NoteType        *string
	Priority        *string
	VisibleToClient *bool
	Pinned          *bool
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
	Page            int
	Limit           int
	SortBy          string
	SortDir         string
}

// ListNotesResponse response for list notes query
type ListNotesResponse struct {
	Notes      []NoteDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// NoteDTO data transfer object for note
type NoteDTO struct {
	ID              string   `json:"id"`
	ContactID       string   `json:"contact_id"`
	SessionID       *string  `json:"session_id,omitempty"`
	AuthorID        string   `json:"author_id"`
	AuthorType      string   `json:"author_type"`
	AuthorName      string   `json:"author_name"`
	Content         string   `json:"content"`
	NoteType        string   `json:"note_type"`
	Priority        string   `json:"priority"`
	VisibleToClient bool     `json:"visible_to_client"`
	Pinned          bool     `json:"pinned"`
	Tags            []string `json:"tags,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// ListNotesQueryHandler handles ListNotesQuery
type ListNotesQueryHandler struct {
	noteRepo note.Repository
	logger   *zap.Logger
}

// NewListNotesQueryHandler creates a new ListNotesQueryHandler
func NewListNotesQueryHandler(noteRepo note.Repository, logger *zap.Logger) *ListNotesQueryHandler {
	return &ListNotesQueryHandler{
		noteRepo: noteRepo,
		logger:   logger,
	}
}

// Handle executes the ListNotesQuery
func (h *ListNotesQueryHandler) Handle(ctx context.Context, query ListNotesQuery) (*ListNotesResponse, error) {
	h.logger.Info("Listing notes",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Build filters
	filters := note.NoteFilters{
		TenantID:        query.TenantID.String(),
		ContactID:       query.ContactID,
		SessionID:       query.SessionID,
		AuthorID:        query.AuthorID,
		AuthorType:      query.AuthorType,
		NoteType:        query.NoteType,
		Priority:        query.Priority,
		VisibleToClient: query.VisibleToClient,
		Pinned:          query.Pinned,
		CreatedAfter:    query.CreatedAfter,
		CreatedBefore:   query.CreatedBefore,
		Limit:           query.Limit,
		Offset:          (query.Page - 1) * query.Limit,
		SortBy:          query.SortBy,
		SortOrder:       query.SortDir,
	}

	// Fetch notes from repository
	notes, totalCount, err := h.noteRepo.FindByTenantWithFilters(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list notes", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	noteDTOs := make([]NoteDTO, len(notes))
	for i, n := range notes {
		dto := NoteDTO{
			ID:              n.ID().String(),
			ContactID:       n.ContactID().String(),
			AuthorID:        n.AuthorID().String(),
			AuthorType:      string(n.AuthorType()),
			AuthorName:      n.AuthorName(),
			Content:         n.Content(),
			NoteType:        string(n.NoteType()),
			Priority:        string(n.Priority()),
			VisibleToClient: n.VisibleToClient(),
			Pinned:          n.Pinned(),
			Tags:            n.Tags(),
			CreatedAt:       n.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       n.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		if sessionID := n.SessionID(); sessionID != nil {
			sessionStr := sessionID.String()
			dto.SessionID = &sessionStr
		}

		noteDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListNotesResponse{
		Notes:      noteDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
