package handlers

import (
	"net/http"
	"strconv"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/internal/application/queries"
	"github.com/caloi/ventros-crm/internal/domain/core/project"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProjectHandler struct {
	logger                     *zap.Logger
	projectRepo                project.Repository
	listProjectsQueryHandler   *queries.ListProjectsQueryHandler
	searchProjectsQueryHandler *queries.SearchProjectsQueryHandler
}

func NewProjectHandler(logger *zap.Logger, projectRepo project.Repository) *ProjectHandler {
	return &ProjectHandler{
		logger:                     logger,
		projectRepo:                projectRepo,
		listProjectsQueryHandler:   queries.NewListProjectsQueryHandler(projectRepo, logger),
		searchProjectsQueryHandler: queries.NewSearchProjectsQueryHandler(projectRepo, logger),
	}
}

// Removed generic helpers - using direct uuid.Parse instead

// CreateProjectRequest representa o payload para criar um projeto
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required" example:"Projeto Vendas"`
	Description string `json:"description" example:"Projeto principal de vendas"`
	TenantID    string `json:"tenant_id" binding:"required" example:"tenant_123"`
}

// UpdateProjectRequest representa o payload para atualizar um projeto
type UpdateProjectRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Active      *bool   `json:"active,omitempty"`
}

// ListProjects lists all projects with optional filters
//
//	@Summary		List projects
//	@Description	Lista todos os projetos do usuário autenticado
//	@Tags			CRM - Projects
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			tenant_id	query		string					false	"Filter by tenant ID"
//	@Param			active		query		bool					false	"Filter by active status"
//	@Param			limit		query		int						false	"Limit results"			default(50)
//	@Param			offset		query		int						false	"Offset for pagination"	default(0)
//	@Success		200			{array}		map[string]interface{}	"List of projects"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401			{object}	map[string]interface{}	"Not authenticated"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	// TODO: Implementar listagem real filtrada por user_id via RLS
	// RLS middleware já filtra por user_id automaticamente
	projects := []map[string]interface{}{}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Projects retrieved successfully",
		"projects": projects,
		"count":    len(projects),
	})
}

// CreateProject creates a new project
//
//	@Summary		Create project
//	@Description	Cria um novo projeto
//	@Tags			CRM - Projects
//	@Accept			json
//	@Produce		json
//	@Param			project	body		CreateProjectRequest	true	"Project data"
//	@Success		201		{object}	map[string]interface{}	"Project created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse project request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement project creation
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Project creation not yet implemented",
		"name":        req.Name,
		"description": req.Description,
		"tenant_id":   req.TenantID,
	})
}

// GetProject gets a project by ID
//
//	@Summary		Get project by ID
//	@Description	Obtém detalhes de um projeto específico
//	@Tags			CRM - Projects
//	@Produce		json
//	@Param			id	path		string					true	"Project ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Project details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid project ID"
//	@Failure		404	{object}	map[string]interface{}	"Project not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid project ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	// TODO: Implement project retrieval
	c.JSON(http.StatusOK, gin.H{
		"message":    "Project retrieval not yet implemented",
		"project_id": projectID,
	})
}

// UpdateProject updates a project
//
//	@Summary		Update project
//	@Description	Atualiza um projeto existente
//	@Tags			CRM - Projects
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Project ID (UUID)"
//	@Param			project	body		UpdateProjectRequest	true	"Project update data"
//	@Success		200		{object}	map[string]interface{}	"Project updated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		404		{object}	map[string]interface{}	"Project not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement project update
	c.JSON(http.StatusOK, gin.H{
		"message":    "Project update not yet implemented",
		"project_id": projectID,
	})
}

// DeleteProject deletes a project
//
//	@Summary		Delete project
//	@Description	Remove um projeto (soft delete)
//	@Tags			CRM - Projects
//	@Produce		json
//	@Param			id	path	string	true	"Project ID (UUID)"
//	@Success		204	"Project deleted successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid project ID"
//	@Failure		404	{object}	map[string]interface{}	"Project not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID format"})
		return
	}

	// TODO: Implement project deletion
	c.JSON(http.StatusOK, gin.H{
		"message":    "Project deletion not yet implemented",
		"project_id": projectID,
	})
}

// ListProjectsAdvanced lists projects with advanced filters, pagination, and sorting
//
//	@Summary		List projects with advanced filters and pagination
//	@Description	Retrieve all projects with filtering by customer and active status. Projects serve as organizational containers that group channels, pipelines, and contacts into business units or departments. Essential for multi-brand organizations and enterprise segmentation.
//	@Description
//	@Description	**Filtering Capabilities:**
//	@Description	- Filter by customer_id to view all projects for a specific customer account
//	@Description	- Filter by active status to show/hide archived projects
//	@Description
//	@Description	**Common Use Cases:**
//	@Description	- Load all active projects for the main dashboard
//	@Description	- Build project selector dropdowns for channel/pipeline assignment
//	@Description	- View complete project portfolio for a specific customer
//	@Description	- Audit project configurations across the organization
//	@Description	- Identify inactive projects for cleanup and archival
//	@Description	- Generate project reports and analytics by customer
//	@Description
//	@Description	**Sorting Options:**
//	@Description	- Sort by name (alphabetical organization)
//	@Description	- Sort by created_at (chronological order)
//	@Description	- Ascending or descending order
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized GORM indexes on tenant+customer for fast customer project queries
//	@Description	- Composite indexes on tenant+active for quick active project retrieval
//	@Description	- Small result sets (typically < 100 projects per tenant) for instant responses
//	@Tags			CRM - Projects
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			customer_id	query		string							false	"Filter by customer UUID - Example: 550e8400-e29b-41d4-a716-446655440000"
//	@Param			active		query		bool							false	"Filter by active status - true: only active, false: only inactive"	example(true)
//	@Param			page		query		int								false	"Page number for pagination (starts at 1)"							default(1)				minimum(1)			example(1)
//	@Param			limit		query		int								false	"Results per page (max 100)"										default(20)				minimum(1)			maximum(100)	example(20)
//	@Param			sort_by		query		string							false	"Field to sort by"													Enums(name, created_at)	default(created_at)	example(name)
//	@Param			sort_dir	query		string							false	"Sort direction"													Enums(asc, desc)		default(desc)		example(asc)
//	@Success		200			{object}	queries.ListProjectsResponse	"Successfully retrieved projects with full details"
//	@Failure		400			{object}	map[string]interface{}			"Bad Request - Invalid UUID or parameter format"
//	@Failure		401			{object}	map[string]interface{}			"Unauthorized - Authentication required"
//	@Failure		403			{object}	map[string]interface{}			"Forbidden - No access to this tenant's projects"
//	@Failure		500			{object}	map[string]interface{}			"Internal Server Error"
//	@Router			/api/v1/crm/projects/advanced [get]
func (h *ProjectHandler) ListProjectsAdvanced(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		apierrors.InternalError(c, "Invalid tenant configuration", err)
		return
	}

	// Parse pagination
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse sorting
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	// Parse customer_id filter
	var customerID *uuid.UUID
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if cid, err := uuid.Parse(customerIDStr); err == nil {
			customerID = &cid
		} else {
			apierrors.ValidationError(c, "customer_id", "Invalid UUID format")
			return
		}
	}

	// Parse active filter
	var active *bool
	if activeStr := c.Query("active"); activeStr != "" {
		if a, err := strconv.ParseBool(activeStr); err == nil {
			active = &a
		}
	}

	query := queries.ListProjectsQuery{
		TenantID:   tenantID,
		CustomerID: customerID,
		Active:     active,
		Page:       page,
		Limit:      limit,
		SortBy:     sortBy,
		SortDir:    sortDir,
	}

	response, err := h.listProjectsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list projects", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve projects", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchProjects performs full-text search on projects
//
//	@Summary		Search projects by name and description
//	@Description	Full-text search across project names and descriptions. Perfect for finding specific business units or departments in organizations with many projects.
//	@Description
//	@Description	**Search Capabilities:**
//	@Description	- Searches project names (primary field)
//	@Description	- Searches project descriptions (secondary field)
//	@Description	- Case-insensitive ILIKE matching
//	@Description
//	@Description	**Match Scoring:**
//	@Description	- Name matches: 1.5 score (higher priority)
//	@Description	- Description matches: 1.2 score (lower priority)
//	@Description
//	@Description	**Search Examples:**
//	@Description	- "sales" - Find all sales-related projects
//	@Description	- "support" - Find customer support departments
//	@Description	- "Q1 2024" - Locate quarterly projects
//	@Description	- "EMEA" - Find regional projects
//	@Tags			CRM - Projects
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string							true	"Search query - name or description"	minlength(1)	example(sales project)
//	@Param			limit	query		int								false	"Maximum results (max 100)"				default(20)		minimum(1)	maximum(100)	example(10)
//	@Success		200		{object}	queries.SearchProjectsResponse	"Found projects with match scores"
//	@Failure		400		{object}	map[string]interface{}			"Bad Request - Empty search query"
//	@Failure		401		{object}	map[string]interface{}			"Unauthorized"
//	@Failure		403		{object}	map[string]interface{}			"Forbidden"
//	@Failure		500		{object}	map[string]interface{}			"Internal Server Error"
//	@Router			/api/v1/crm/projects/search [get]
func (h *ProjectHandler) SearchProjects(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		h.logger.Error("Invalid tenant ID", zap.Error(err))
		apierrors.InternalError(c, "Invalid tenant configuration", err)
		return
	}

	searchText := c.Query("q")
	if searchText == "" {
		apierrors.ValidationError(c, "q", "Search query 'q' is required")
		return
	}

	// Parse limit
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	query := queries.SearchProjectsQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchProjectsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search projects", zap.Error(err))
		apierrors.InternalError(c, "Failed to search projects", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
