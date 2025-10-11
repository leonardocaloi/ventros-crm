package queries

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ListPipelinesQuery query to list pipelines with filters, pagination, and sorting
type ListPipelinesQuery struct {
	TenantID  shared.TenantID
	ProjectID *uuid.UUID
	Active    *bool
	Color     *string
	Page      int
	Limit     int
	SortBy    string
	SortDir   string
}

// ListPipelinesResponse response for list pipelines query
type ListPipelinesResponse struct {
	Pipelines  []PipelineDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// PipelineDTO data transfer object for pipeline
type PipelineDTO struct {
	ID                    string  `json:"id"`
	ProjectID             string  `json:"project_id"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	Color                 string  `json:"color"`
	Position              int     `json:"position"`
	Active                bool    `json:"active"`
	SessionTimeoutMinutes *int    `json:"session_timeout_minutes,omitempty"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

// ListPipelinesQueryHandler handles ListPipelinesQuery
type ListPipelinesQueryHandler struct {
	pipelineRepo pipeline.Repository
	logger       *zap.Logger
}

// NewListPipelinesQueryHandler creates a new ListPipelinesQueryHandler
func NewListPipelinesQueryHandler(pipelineRepo pipeline.Repository, logger *zap.Logger) *ListPipelinesQueryHandler {
	return &ListPipelinesQueryHandler{
		pipelineRepo: pipelineRepo,
		logger:       logger,
	}
}

// Handle executes the ListPipelinesQuery
func (h *ListPipelinesQueryHandler) Handle(ctx context.Context, query ListPipelinesQuery) (*ListPipelinesResponse, error) {
	h.logger.Info("Listing pipelines",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Build filters
	filters := pipeline.PipelineFilters{
		TenantID:  query.TenantID.String(),
		ProjectID: query.ProjectID,
		Active:    query.Active,
		Color:     query.Color,
		Limit:     query.Limit,
		Offset:    (query.Page - 1) * query.Limit,
		SortBy:    query.SortBy,
		SortOrder: query.SortDir,
	}

	// Fetch pipelines from repository
	pipelines, totalCount, err := h.pipelineRepo.FindByTenantWithFilters(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list pipelines", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	pipelineDTOs := make([]PipelineDTO, len(pipelines))
	for i, p := range pipelines {
		dto := PipelineDTO{
			ID:                    p.ID().String(),
			ProjectID:             p.ProjectID().String(),
			Name:                  p.Name(),
			Description:           p.Description(),
			Color:                 p.Color(),
			Position:              p.Position(),
			Active:                p.IsActive(),
			SessionTimeoutMinutes: p.SessionTimeoutMinutes(),
			CreatedAt:             p.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:             p.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		pipelineDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListPipelinesResponse{
		Pipelines:  pipelineDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
