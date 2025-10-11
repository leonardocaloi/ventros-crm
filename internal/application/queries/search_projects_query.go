package queries

import (
	"context"
	"strings"

	"github.com/caloi/ventros-crm/internal/domain/core/project"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"go.uber.org/zap"
)

// SearchProjectsQuery query to search projects by text
type SearchProjectsQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchProjectsResponse response for search projects query
type SearchProjectsResponse struct {
	Projects []ProjectSearchResultDTO `json:"projects"`
	Count    int                      `json:"count"`
}

// ProjectSearchResultDTO search result for project
type ProjectSearchResultDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Active      bool    `json:"active"`
	MatchScore  float64 `json:"match_score"`
	MatchField  string  `json:"match_field"`
}

// SearchProjectsQueryHandler handles SearchProjectsQuery
type SearchProjectsQueryHandler struct {
	projectRepo project.Repository
	logger      *zap.Logger
}

// NewSearchProjectsQueryHandler creates a new SearchProjectsQueryHandler
func NewSearchProjectsQueryHandler(projectRepo project.Repository, logger *zap.Logger) *SearchProjectsQueryHandler {
	return &SearchProjectsQueryHandler{
		projectRepo: projectRepo,
		logger:      logger,
	}
}

// Handle executes the SearchProjectsQuery
func (h *SearchProjectsQueryHandler) Handle(ctx context.Context, query SearchProjectsQuery) (*SearchProjectsResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchProjectsResponse{
			Projects: []ProjectSearchResultDTO{},
			Count:    0,
		}, nil
	}

	h.logger.Info("Searching projects",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search projects using repository
	projects, _, err := h.projectRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
		0,
	)
	if err != nil {
		h.logger.Error("Failed to search projects", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs with match scoring
	results := make([]ProjectSearchResultDTO, len(projects))
	for i, p := range projects {
		dto := ProjectSearchResultDTO{
			ID:          p.ID().String(),
			Name:        p.Name(),
			Description: p.Description(),
			Active:      p.IsActive(),
			MatchScore:  1.0,
			MatchField:  "name",
		}

		// Determine match field
		if strings.Contains(strings.ToLower(p.Name()), searchText) {
			dto.MatchScore = 1.5
			dto.MatchField = "name"
		} else if strings.Contains(strings.ToLower(p.Description()), searchText) {
			dto.MatchScore = 1.2
			dto.MatchField = "description"
		}

		results[i] = dto
	}

	return &SearchProjectsResponse{
		Projects: results,
		Count:    len(results),
	}, nil
}
