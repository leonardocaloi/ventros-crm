package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/agent"
	"go.uber.org/zap"
)

// ListAgentsQuery query to list agents with filters, pagination, and sorting
type ListAgentsQuery struct {
	TenantID  shared.TenantID
	ProjectID *uuid.UUID
	Type      *agent.AgentType
	Status    *agent.AgentStatus
	Active    *bool
	Page      int
	Limit     int
	SortBy    string
	SortDir   string
}

// ListAgentsResponse response for list agents query
type ListAgentsResponse struct {
	Agents     []AgentDTO
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// AgentDTO data transfer object for agent
type AgentDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListAgentsQueryHandler handles ListAgentsQuery
type ListAgentsQueryHandler struct {
	agentRepo agent.Repository
	logger    *zap.Logger
}

// NewListAgentsQueryHandler creates a new ListAgentsQueryHandler
func NewListAgentsQueryHandler(agentRepo agent.Repository, logger *zap.Logger) *ListAgentsQueryHandler {
	return &ListAgentsQueryHandler{
		agentRepo: agentRepo,
		logger:    logger,
	}
}

// Handle executes the ListAgentsQuery
func (h *ListAgentsQueryHandler) Handle(ctx context.Context, query ListAgentsQuery) (*ListAgentsResponse, error) {
	h.logger.Info("Listing agents",
		zap.String("tenant_id", query.TenantID.String()),
		zap.Int("page", query.Page),
		zap.Int("limit", query.Limit))

	// Build filters
	filters := agent.AgentFilters{
		TenantID:  query.TenantID.String(),
		ProjectID: query.ProjectID,
		Type:      query.Type,
		Status:    query.Status,
		Active:    query.Active,
		Limit:     query.Limit,
		Offset:    (query.Page - 1) * query.Limit,
		SortBy:    query.SortBy,
		SortOrder: query.SortDir,
	}

	// Fetch agents from repository
	agents, totalCount, err := h.agentRepo.FindByTenantWithFilters(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to list agents", zap.Error(err))
		return nil, err
	}

	// Convert domain entities to DTOs
	agentDTOs := make([]AgentDTO, len(agents))
	for i, a := range agents {
		dto := AgentDTO{
			ID:        a.ID().String(),
			Name:      a.Name(),
			Email:     a.Email(),
			Active:    a.IsActive(),
			CreatedAt: a.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: a.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}

		agentDTOs[i] = dto
	}

	// Calculate pagination
	totalPages := int(totalCount) / query.Limit
	if int(totalCount)%query.Limit > 0 {
		totalPages++
	}

	return &ListAgentsResponse{
		Agents:     agentDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}, nil
}
