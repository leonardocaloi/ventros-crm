package queries

import (
	"context"
	"strings"

	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/agent"
	"go.uber.org/zap"
)

// SearchAgentsQuery query to search agents by text
type SearchAgentsQuery struct {
	TenantID   shared.TenantID
	SearchText string
	Limit      int
}

// SearchAgentsResponse response for search agents query
type SearchAgentsResponse struct {
	Agents []AgentSearchResultDTO `json:"agents"`
	Count  int                    `json:"count"`
}

// AgentSearchResultDTO search result for agent
type AgentSearchResultDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Active     bool    `json:"active"`
	MatchScore float64 `json:"match_score"`
	MatchField string  `json:"match_field"`
}

// SearchAgentsQueryHandler handles SearchAgentsQuery
type SearchAgentsQueryHandler struct {
	agentRepo agent.Repository
	logger    *zap.Logger
}

// NewSearchAgentsQueryHandler creates a new SearchAgentsQueryHandler
func NewSearchAgentsQueryHandler(agentRepo agent.Repository, logger *zap.Logger) *SearchAgentsQueryHandler {
	return &SearchAgentsQueryHandler{
		agentRepo: agentRepo,
		logger:    logger,
	}
}

// Handle executes the SearchAgentsQuery
func (h *SearchAgentsQueryHandler) Handle(ctx context.Context, query SearchAgentsQuery) (*SearchAgentsResponse, error) {
	// Normalize search text
	searchText := strings.ToLower(strings.TrimSpace(query.SearchText))
	if searchText == "" {
		return &SearchAgentsResponse{
			Agents: []AgentSearchResultDTO{},
			Count:  0,
		}, nil
	}

	h.logger.Info("Searching agents",
		zap.String("tenant_id", query.TenantID.String()),
		zap.String("search_text", searchText),
		zap.Int("limit", query.Limit))

	// Search agents using repository
	agents, _, err := h.agentRepo.SearchByText(
		ctx,
		query.TenantID.String(),
		searchText,
		query.Limit,
		0,
	)
	if err != nil {
		h.logger.Error("Failed to search agents", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs with match scoring
	results := make([]AgentSearchResultDTO, len(agents))
	for i, a := range agents {
		dto := AgentSearchResultDTO{
			ID:         a.ID().String(),
			Name:       a.Name(),
			Email:      a.Email(),
			Active:     a.IsActive(),
			MatchScore: 1.0,
			MatchField: "name",
		}

		// Determine match field
		if strings.Contains(strings.ToLower(a.Name()), searchText) {
			dto.MatchScore = 1.5
			dto.MatchField = "name"
		} else if strings.Contains(strings.ToLower(a.Email()), searchText) {
			dto.MatchScore = 1.3
			dto.MatchField = "email"
		}

		results[i] = dto
	}

	return &SearchAgentsResponse{
		Agents: results,
		Count:  len(results),
	}, nil
}
