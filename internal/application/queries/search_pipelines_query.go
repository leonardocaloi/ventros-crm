package queries

import (
	"context"
	"strings"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"go.uber.org/zap"
)

// SearchPipelinesQuery query to search pipelines by text
type SearchPipelinesQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchPipelinesResponse response for search pipelines query
type SearchPipelinesResponse struct {
	Pipelines []PipelineSearchResultDTO `json:"pipelines"`
	Count     int                       `json:"count"`
}

// PipelineSearchResultDTO search result for pipeline
type PipelineSearchResultDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Active      bool    `json:"active"`
	MatchScore  float64 `json:"match_score"`
	MatchField  string  `json:"match_field"`
}

// SearchPipelinesQueryHandler handles SearchPipelinesQuery
type SearchPipelinesQueryHandler struct {
	pipelineRepo pipeline.Repository
	logger       *zap.Logger
}

// NewSearchPipelinesQueryHandler creates a new SearchPipelinesQueryHandler
func NewSearchPipelinesQueryHandler(pipelineRepo pipeline.Repository, logger *zap.Logger) *SearchPipelinesQueryHandler {
	return &SearchPipelinesQueryHandler{
		pipelineRepo: pipelineRepo,
		logger:       logger,
	}
}

// Handle executes the SearchPipelinesQuery
func (h *SearchPipelinesQueryHandler) Handle(ctx context.Context, query SearchPipelinesQuery) (*SearchPipelinesResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchPipelinesResponse{
			Pipelines: []PipelineSearchResultDTO{},
			Count:     0,
		}, nil
	}

	h.logger.Info("Searching pipelines",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search pipelines using repository
	pipelines, _, err := h.pipelineRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
		0,
	)
	if err != nil {
		h.logger.Error("Failed to search pipelines", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs with match scoring
	results := make([]PipelineSearchResultDTO, len(pipelines))
	for i, p := range pipelines {
		dto := PipelineSearchResultDTO{
			ID:          p.ID().String(),
			Name:        p.Name(),
			Description: p.Description(),
			Color:       p.Color(),
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

	return &SearchPipelinesResponse{
		Pipelines: results,
		Count:     len(results),
	}, nil
}
