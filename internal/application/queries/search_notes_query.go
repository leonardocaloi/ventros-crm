package queries

import (
	"context"
	"strings"

	"github.com/caloi/ventros-crm/internal/domain/note"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"go.uber.org/zap"
)

// SearchNotesQuery query to search notes by text
type SearchNotesQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchNotesResponse response for search notes query
type SearchNotesResponse struct {
	Notes []NoteSearchResultDTO `json:"notes"`
	Count int                   `json:"count"`
}

// NoteSearchResultDTO search result for note
type NoteSearchResultDTO struct {
	ID         string  `json:"id"`
	ContactID  string  `json:"contact_id"`
	AuthorName string  `json:"author_name"`
	Content    string  `json:"content"`
	NoteType   string  `json:"note_type"`
	Priority   string  `json:"priority"`
	Pinned     bool    `json:"pinned"`
	CreatedAt  string  `json:"created_at"`
	MatchScore float64 `json:"match_score"`
	MatchField string  `json:"match_field"`
}

// SearchNotesQueryHandler handles SearchNotesQuery
type SearchNotesQueryHandler struct {
	noteRepo note.Repository
	logger   *zap.Logger
}

// NewSearchNotesQueryHandler creates a new SearchNotesQueryHandler
func NewSearchNotesQueryHandler(noteRepo note.Repository, logger *zap.Logger) *SearchNotesQueryHandler {
	return &SearchNotesQueryHandler{
		noteRepo: noteRepo,
		logger:   logger,
	}
}

// Handle executes the SearchNotesQuery
func (h *SearchNotesQueryHandler) Handle(ctx context.Context, query SearchNotesQuery) (*SearchNotesResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchNotesResponse{
			Notes: []NoteSearchResultDTO{},
			Count: 0,
		}, nil
	}

	h.logger.Info("Searching notes",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search notes using repository
	notes, _, err := h.noteRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
		0,
	)
	if err != nil {
		h.logger.Error("Failed to search notes", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs with match scoring
	results := make([]NoteSearchResultDTO, len(notes))
	for i, n := range notes {
		dto := NoteSearchResultDTO{
			ID:         n.ID().String(),
			ContactID:  n.ContactID().String(),
			AuthorName: n.AuthorName(),
			Content:    n.Content(),
			NoteType:   string(n.NoteType()),
			Priority:   string(n.Priority()),
			Pinned:     n.Pinned(),
			CreatedAt:  n.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			MatchScore: 1.0,
			MatchField: "content",
		}

		// Determine match field
		if strings.Contains(strings.ToLower(n.Content()), searchText) {
			dto.MatchScore = 1.5
			dto.MatchField = "content"
		} else if strings.Contains(strings.ToLower(n.AuthorName()), searchText) {
			dto.MatchScore = 1.2
			dto.MatchField = "author_name"
		}

		results[i] = dto
	}

	return &SearchNotesResponse{
		Notes: results,
		Count: len(results),
	}, nil
}
