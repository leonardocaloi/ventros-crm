package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/project"
	"github.com/ventros/crm/internal/domain/core/shared"
	"go.uber.org/zap"
)

// ListProjectsQuery query to list projects with filters, pagination, and sorting
type ListProjectsQuery struct {
	TenantID   shared.TenantID
	CustomerID *uuid.UUID
	Active     *bool
	Page       int
	Limit      int
	SortBy     string
	SortDir    string
}

// ListProjectsResponse response for list projects query
type ListProjectsResponse struct {
	Projects   []ProjectDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// ProjectDTO data transfer object for project
type ProjectDTO struct {
	ID                    string                 `json:"id"`
	CustomerID            string                 `json:"customer_id"`
	BillingAccountID      string                 `json:"billing_account_id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	Configuration         map[string]interface{} `json:"configuration,omitempty"`
	Active                bool                   `json:"active"`
	SessionTimeoutMinutes int                    `json:"session_timeout_minutes"`
	CreatedAt             string                 `json:"created_at"`
	UpdatedAt             string                 `json:"updated_at"`
}

// ListProjectsQueryHandler handles ListProjectsQuery
type ListProjectsQueryHandler struct {
	projectRepo project.Repository
	logger      *zap.Logger
}

// NewListProjectsQueryHandler creates a new ListProjectsQueryHandler
func NewListProjectsQueryHandler(projectRepo project.Repository, logger *zap.Logger) *ListProjectsQueryHandler {
	return &ListProjectsQueryHandler{
		projectRepo: projectRepo,
		logger:      logger,
	}
}

// Handle executes the ListProjectsQuery
func (h *ListProjectsQueryHandler) Handle(ctx context.Context, query ListProjectsQuery) (*ListProjectsResponse, error) {
	h.logger.Info("Listing projects",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Build filters
	filters := project.ProjectFilters{
		TenantID:   query.TenantID.String(),
		CustomerID: query.CustomerID,
		Active:     query.Active,
		Limit:      query.Limit,
		Offset:     (query.Page - 1) * query.Limit,
		SortBy:     query.SortBy,
		SortOrder:  query.SortDir,
	}

	// Fetch projects from repository
	projects, totalCount, err := h.projectRepo.FindByTenantWithFilters(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list projects", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	projectDTOs := make([]ProjectDTO, len(projects))
	for i, p := range projects {
		dto := ProjectDTO{
			ID:                    p.ID().String(),
			CustomerID:            p.CustomerID().String(),
			BillingAccountID:      p.BillingAccountID().String(),
			Name:                  p.Name(),
			Description:           p.Description(),
			Configuration:         p.Configuration(),
			Active:                p.IsActive(),
			SessionTimeoutMinutes: p.SessionTimeoutMinutes(),
			CreatedAt:             p.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:             p.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		projectDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListProjectsResponse{
		Projects:   projectDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
