package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	apierrors "github.com/ventros/crm/infrastructure/http/errors"
	"github.com/ventros/crm/infrastructure/http/middleware"
	pipelineapp "github.com/ventros/crm/internal/application/pipeline"
	"github.com/ventros/crm/internal/application/queries"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
	"go.uber.org/zap"
)

type PipelineHandler struct {
	logger                      *zap.Logger
	pipelineRepo                pipeline.Repository
	listPipelinesQueryHandler   *queries.ListPipelinesQueryHandler
	searchPipelinesQueryHandler *queries.SearchPipelinesQueryHandler
	logrusLogger                *logrus.Logger
}

func NewPipelineHandler(logger *zap.Logger, pipelineRepo pipeline.Repository, logrusLogger *logrus.Logger) *PipelineHandler {
	return &PipelineHandler{
		logger:                      logger,
		pipelineRepo:                pipelineRepo,
		listPipelinesQueryHandler:   queries.NewListPipelinesQueryHandler(pipelineRepo, logger),
		searchPipelinesQueryHandler: queries.NewSearchPipelinesQueryHandler(pipelineRepo, logger),
		logrusLogger:                logrusLogger,
	}
}

// CreatePipelineRequest representa o payload para criar um pipeline
type CreatePipelineRequest struct {
	Name        string `json:"name" binding:"required" example:"Vendas"`
	Description string `json:"description" example:"Pipeline de vendas principal"`
	Color       string `json:"color" example:"#3B82F6"`
	Position    int    `json:"position" example:"0"`
}

// UpdatePipelineRequest representa o payload para atualizar um pipeline
type UpdatePipelineRequest struct {
	Name                  *string `json:"name,omitempty"`
	Description           *string `json:"description,omitempty"`
	Color                 *string `json:"color,omitempty"`
	Position              *int    `json:"position,omitempty"`
	Active                *bool   `json:"active,omitempty"`
	SessionTimeoutMinutes *int    `json:"session_timeout_minutes,omitempty" example:"1"`
}

// CreateStatusRequest representa o payload para criar um status
type CreateStatusRequest struct {
	Name        string              `json:"name" binding:"required" example:"Novo Lead"`
	Description string              `json:"description" example:"Lead recém chegado"`
	Color       string              `json:"color" example:"#10B981"`
	StatusType  pipeline.StatusType `json:"status_type" binding:"required" example:"open"`
	Position    int                 `json:"position" example:"0"`
}

// UpdateStatusRequest representa o payload para atualizar um status
type UpdateStatusRequest struct {
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Color       *string              `json:"color,omitempty"`
	StatusType  *pipeline.StatusType `json:"status_type,omitempty"`
	Position    *int                 `json:"position,omitempty"`
	Active      *bool                `json:"active,omitempty"`
}

// ChangeContactStatusRequest representa o payload para mudar status de contato
type ChangeContactStatusRequest struct {
	StatusID uuid.UUID `json:"status_id" binding:"required"`
	Reason   string    `json:"reason" example:"Contato respondeu"`
	Notes    string    `json:"notes" example:"Cliente interessado no produto X"`
}

// ListPipelines lists all pipelines for a project
//
//	@Summary		List pipelines
//	@Description	Lista todos os pipelines de um projeto (apenas do usuário autenticado)
//	@Tags			CRM - Pipelines
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			project_id	query		string					true	"Project ID (UUID)"
//	@Param			active		query		bool					false	"Filter by active status"
//	@Success		200			{array}		map[string]interface{}	"List of pipelines"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401			{object}	map[string]interface{}	"Not authenticated"
//	@Failure		403			{object}	map[string]interface{}	"Project not owned by user"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/pipelines [get]
func (h *PipelineHandler) ListPipelines(c *gin.Context) {
	// RLS middleware já garante autenticação e acesso aos recursos do usuário

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id format"})
		return
	}

	// RLS middleware já garante acesso apenas aos recursos do usuário
	// Removido ownership check - RLS cuida disso automaticamente

	activeOnly := c.Query("active") == "true"

	var pipelines []*pipeline.Pipeline
	if activeOnly {
		pipelines, err = h.pipelineRepo.FindActivePipelinesByProject(c.Request.Context(), projectID)
		if err != nil {
			h.logger.Error("Failed to find pipelines", zap.String("project_id", projectID.String()), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pipelines"})
			return
		}

		// Convert to response
		response := make([]map[string]interface{}, len(pipelines))
		for i, p := range pipelines {
			response[i] = h.pipelineToResponse(p)
		}
		c.JSON(http.StatusOK, response)
	} else {
		pipelines, err := h.pipelineRepo.FindPipelinesByProject(c.Request.Context(), projectID)
		if err != nil {
			h.logger.Error("Failed to find pipelines", zap.String("project_id", projectID.String()), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pipelines"})
			return
		}

		// Convert to response
		response := make([]map[string]interface{}, len(pipelines))
		for i, p := range pipelines {
			response[i] = h.pipelineToResponse(p)
		}
		c.JSON(http.StatusOK, response)
	}
}

// CreatePipeline creates a new pipeline
//
//	@Summary		Create pipeline
//	@Description	Cria um novo pipeline
//	@Tags			CRM - Pipelines
//	@Accept			json
//	@Produce		json
//	@Param			project_id	query		string					true	"Project ID (UUID)"
//	@Param			pipeline	body		CreatePipelineRequest	true	"Pipeline data"
//	@Success		201			{object}	map[string]interface{}	"Pipeline created successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/pipelines [post]
func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id format"})
		return
	}

	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse pipeline request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Get tenant_id from context/auth
	tenantID := "default" // Placeholder

	// Create domain pipeline
	domainPipeline, err := pipeline.NewPipeline(projectID, tenantID, req.Name)
	if err != nil {
		h.logger.Error("Failed to create domain pipeline", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set optional fields
	if req.Description != "" {
		domainPipeline.UpdateDescription(req.Description)
	}

	if req.Color != "" {
		domainPipeline.UpdateColor(req.Color)
	}

	domainPipeline.UpdatePosition(req.Position)

	// Save pipeline
	if err := h.pipelineRepo.SavePipeline(c.Request.Context(), domainPipeline); err != nil {
		h.logger.Error("Failed to save pipeline", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pipeline"})
		return
	}

	// Convert to response
	response := h.pipelineToResponse(domainPipeline)
	c.JSON(http.StatusCreated, response)
}

// GetPipeline gets a pipeline by ID with its statuses
//
//	@Summary		Get pipeline by ID
//	@Description	Obtém detalhes de um pipeline específico com seus status
//	@Tags			CRM - Pipelines
//	@Produce		json
//	@Param			id	path		string					true	"Pipeline ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Pipeline details with statuses"
//	@Failure		400	{object}	map[string]interface{}	"Invalid pipeline ID"
//	@Failure		404	{object}	map[string]interface{}	"Pipeline not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/pipelines/{id} [get]
func (h *PipelineHandler) GetPipeline(c *gin.Context) {
	idStr := c.Param("id")
	pipelineID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid pipeline ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID format"})
		return
	}

	domainPipeline, statuses, err := h.pipelineRepo.GetPipelineWithStatuses(c.Request.Context(), pipelineID)
	if err != nil {
		h.logger.Error("Failed to find pipeline", zap.String("pipeline_id", pipelineID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pipeline"})
		return
	}

	if domainPipeline == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}

	// Convert to response
	response := h.pipelineToResponse(domainPipeline)

	// Add statuses
	statusesResponse := make([]map[string]interface{}, len(statuses))
	for i, s := range statuses {
		statusesResponse[i] = h.statusToResponse(s)
	}
	response["statuses"] = statusesResponse

	c.JSON(http.StatusOK, response)
}

// CreateStatus creates a new status in a pipeline
//
//	@Summary		Create status
//	@Description	Cria um novo status em um pipeline
//	@Tags			CRM - Pipelines
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Pipeline ID (UUID)"
//	@Param			status	body		CreateStatusRequest		true	"Status data"
//	@Success		201		{object}	map[string]interface{}	"Status created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		404		{object}	map[string]interface{}	"Pipeline not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/pipelines/{id}/statuses [post]
func (h *PipelineHandler) CreateStatus(c *gin.Context) {
	pipelineIDStr := c.Param("id")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID format"})
		return
	}

	var req CreateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse status request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Verify pipeline exists
	domainPipeline, err := h.pipelineRepo.FindPipelineByID(c.Request.Context(), pipelineID)
	if err != nil {
		h.logger.Error("Failed to find pipeline", zap.String("pipeline_id", pipelineID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve pipeline"})
		return
	}

	if domainPipeline == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}

	// Create domain status
	domainStatus, err := pipeline.NewStatus(pipelineID, req.Name, req.StatusType)
	if err != nil {
		h.logger.Error("Failed to create domain status", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set optional fields
	if req.Description != "" {
		domainStatus.UpdateDescription(req.Description)
	}

	if req.Color != "" {
		domainStatus.UpdateColor(req.Color)
	}

	domainStatus.UpdatePosition(req.Position)

	// Save status
	if err := h.pipelineRepo.SaveStatus(c.Request.Context(), domainStatus); err != nil {
		h.logger.Error("Failed to save status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create status"})
		return
	}

	// Add status to pipeline
	if err := h.pipelineRepo.AddStatusToPipeline(c.Request.Context(), pipelineID, domainStatus.ID()); err != nil {
		h.logger.Error("Failed to add status to pipeline", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add status to pipeline"})
		return
	}

	// Convert to response
	response := h.statusToResponse(domainStatus)
	c.JSON(http.StatusCreated, response)
}

// ChangeContactStatus changes the status of a contact in a pipeline
//
//	@Summary		Change contact status
//	@Description	Altera o status de um contato em um pipeline
//	@Tags			CRM - Pipelines
//	@Accept			json
//	@Produce		json
//	@Param			pipeline_id	path		string						true	"Pipeline ID (UUID)"
//	@Param			contact_id	path		string						true	"Contact ID (UUID)"
//	@Param			request		body		ChangeContactStatusRequest	true	"Status change data"
//	@Success		200			{object}	map[string]interface{}		"Status changed successfully"
//	@Failure		400			{object}	map[string]interface{}		"Invalid request"
//	@Failure		404			{object}	map[string]interface{}		"Pipeline or contact not found"
//	@Failure		500			{object}	map[string]interface{}		"Internal server error"
//	@Router			/api/v1/pipelines/{pipeline_id}/contacts/{contact_id}/status [put]
func (h *PipelineHandler) ChangeContactStatus(c *gin.Context) {
	pipelineIDStr := c.Param("pipeline_id")
	contactIDStr := c.Param("contact_id")

	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID format"})
		return
	}

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
		return
	}

	var req ChangeContactStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse status change request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Verify status exists in pipeline
	status, err := h.pipelineRepo.FindStatusByID(c.Request.Context(), req.StatusID)
	if err != nil {
		h.logger.Error("Failed to find status", zap.String("status_id", req.StatusID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve status"})
		return
	}

	if status == nil || status.PipelineID() != pipelineID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status not found in this pipeline"})
		return
	}

	// Set contact status
	if err := h.pipelineRepo.SetContactStatus(c.Request.Context(), contactID, pipelineID, req.StatusID); err != nil {
		h.logger.Error("Failed to set contact status",
			zap.String("contact_id", contactID.String()),
			zap.String("pipeline_id", pipelineID.String()),
			zap.String("status_id", req.StatusID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change contact status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Contact status changed successfully",
		"contact_id":  contactID,
		"pipeline_id": pipelineID,
		"status_id":   req.StatusID,
		"status_name": status.Name(),
		"reason":      req.Reason,
	})
}

// GetContactStatus gets the current status of a contact in a pipeline
//
//	@Summary		Get contact status
//	@Description	Obtém o status atual de um contato em um pipeline
//	@Tags			CRM - Pipelines
//	@Produce		json
//	@Param			pipeline_id	path		string					true	"Pipeline ID (UUID)"
//	@Param			contact_id	path		string					true	"Contact ID (UUID)"
//	@Success		200			{object}	map[string]interface{}	"Contact status"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		404			{object}	map[string]interface{}	"Status not found"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/pipelines/{pipeline_id}/contacts/{contact_id}/status [get]
func (h *PipelineHandler) GetContactStatus(c *gin.Context) {
	pipelineIDStr := c.Param("pipeline_id")
	contactIDStr := c.Param("contact_id")

	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID format"})
		return
	}

	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID format"})
		return
	}

	status, err := h.pipelineRepo.GetContactStatus(c.Request.Context(), contactID, pipelineID)
	if err != nil {
		h.logger.Error("Failed to get contact status",
			zap.String("contact_id", contactID.String()),
			zap.String("pipeline_id", pipelineID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve contact status"})
		return
	}

	if status == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found in this pipeline"})
		return
	}

	response := h.statusToResponse(status)
	response["contact_id"] = contactID
	response["pipeline_id"] = pipelineID

	c.JSON(http.StatusOK, response)
}

// pipelineToResponse converts domain pipeline to API response
func (h *PipelineHandler) pipelineToResponse(p *pipeline.Pipeline) map[string]interface{} {
	return map[string]interface{}{
		"id":          p.ID(),
		"project_id":  p.ProjectID(),
		"tenant_id":   p.TenantID(),
		"name":        p.Name(),
		"description": p.Description(),
		"color":       p.Color(),
		"position":    p.Position(),
		"active":      p.IsActive(),
		"created_at":  p.CreatedAt(),
		"updated_at":  p.UpdatedAt(),
	}
}

// statusToResponse converts domain status to API response
func (h *PipelineHandler) statusToResponse(s *pipeline.Status) map[string]interface{} {
	return map[string]interface{}{
		"id":          s.ID(),
		"pipeline_id": s.PipelineID(),
		"name":        s.Name(),
		"description": s.Description(),
		"color":       s.Color(),
		"status_type": s.StatusType(),
		"position":    s.Position(),
		"active":      s.IsActiveStatus(),
		"created_at":  s.CreatedAt(),
		"updated_at":  s.UpdatedAt(),
	}
}

// ListPipelinesAdvanced lists pipelines with advanced filters, pagination, and sorting
//
//	@Summary		List pipelines with advanced filters
//	@Description	Retrieve all pipelines with filtering by project, active status, and color. Pipelines organize contacts into workflow stages with customizable statuses. Essential for sales processes, support tickets, and multi-stage customer journeys.
//	@Description
//	@Description	**Filtering Capabilities:**
//	@Description	- Filter by project_id to get pipelines for a specific business unit
//	@Description	- Filter by active status to show/hide archived pipelines
//	@Description	- Filter by color for UI organization and visual pipeline management
//	@Description
//	@Description	**Common Use Cases:**
//	@Description	- Load all active pipelines for a project's dashboard
//	@Description	- Build pipeline selector dropdowns for contact assignment
//	@Description	- Audit pipeline configuration across projects
//	@Description	- Identify inactive pipelines for cleanup
//	@Description	- Generate pipeline reports by color-coded categories
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized GORM indexes on tenant+active for fast active pipeline queries
//	@Description	- Small result sets (typically < 50 pipelines per tenant) for instant responses
//	@Tags			CRM - Pipelines
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	query		string							false	"Filter by project UUID"											example(550e8400-e29b-41d4-a716-446655440000)
//	@Param			active		query		bool							false	"Filter by active status - true: only active, false: only inactive"	example(true)
//	@Param			color		query		string							false	"Filter by hex color code - Example: #FF5733, #3B82F6"				example(#3B82F6)
//	@Param			page		query		int								false	"Page number (starts at 1)"											default(1)							minimum(1)			example(1)
//	@Param			limit		query		int								false	"Results per page (max 100)"										default(20)							minimum(1)			maximum(100)	example(20)
//	@Param			sort_by		query		string							false	"Sort field"														Enums(name, position, created_at)	default(created_at)	example(position)
//	@Param			sort_dir	query		string							false	"Sort direction"													Enums(asc, desc)					default(desc)		example(asc)
//	@Success		200			{object}	queries.ListPipelinesResponse	"Successfully retrieved pipelines with full configuration details"
//	@Failure		400			{object}	map[string]interface{}			"Bad Request - Invalid UUID or parameter format"
//	@Failure		401			{object}	map[string]interface{}			"Unauthorized - Authentication required"
//	@Failure		403			{object}	map[string]interface{}			"Forbidden - No access to this tenant's pipelines"
//	@Failure		500			{object}	map[string]interface{}			"Internal Server Error"
//	@Router			/api/v1/crm/pipelines/advanced [get]
func (h *PipelineHandler) ListPipelinesAdvanced(c *gin.Context) {
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

	// Parse project_id filter
	var projectID *uuid.UUID
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if pid, err := uuid.Parse(projectIDStr); err == nil {
			projectID = &pid
		} else {
			apierrors.ValidationError(c, "project_id", "Invalid UUID format")
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

	// Parse color filter
	var color *string
	if colorStr := c.Query("color"); colorStr != "" {
		color = &colorStr
	}

	query := queries.ListPipelinesQuery{
		TenantID:  tenantID,
		ProjectID: projectID,
		Active:    active,
		Color:     color,
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortDir:   sortDir,
	}

	response, err := h.listPipelinesQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list pipelines", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve pipelines", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchPipelines performs full-text search on pipelines
//
//	@Summary		Search pipelines by name and description
//	@Description	Full-text search across pipeline names and descriptions. Perfect for finding specific pipelines in large organizations with many workflow configurations.
//	@Description
//	@Description	**Search Capabilities:**
//	@Description	- Searches pipeline names (primary field)
//	@Description	- Searches pipeline descriptions (secondary field)
//	@Description	- Case-insensitive ILIKE matching
//	@Description
//	@Description	**Match Scoring:**
//	@Description	- Name matches: 1.5 score (higher priority)
//	@Description	- Description matches: 1.2 score (lower priority)
//	@Description
//	@Description	**Search Examples:**
//	@Description	- "sales" - Find all sales-related pipelines
//	@Description	- "support" - Find customer support workflows
//	@Description	- "onboarding" - Locate onboarding process pipelines
//	@Tags			CRM - Pipelines
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string							true	"Search query - name or description"	minlength(1)	example(sales pipeline)
//	@Param			limit	query		int								false	"Maximum results (max 100)"				default(20)		minimum(1)	maximum(100)	example(10)
//	@Success		200		{object}	queries.SearchPipelinesResponse	"Found pipelines with match scores"
//	@Failure		400		{object}	map[string]interface{}			"Bad Request - Empty search query"
//	@Failure		401		{object}	map[string]interface{}			"Unauthorized"
//	@Failure		403		{object}	map[string]interface{}			"Forbidden"
//	@Failure		500		{object}	map[string]interface{}			"Internal Server Error"
//	@Router			/api/v1/crm/pipelines/search [get]
func (h *PipelineHandler) SearchPipelines(c *gin.Context) {
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

	query := queries.SearchPipelinesQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchPipelinesQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search pipelines", zap.Error(err))
		apierrors.InternalError(c, "Failed to search pipelines", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SetCustomFieldRequest represents the request to set a custom field
type SetCustomFieldRequest struct {
	Key   string      `json:"key" binding:"required" example:"budget"`
	Type  string      `json:"type" binding:"required" example:"number"`
	Value interface{} `json:"value" binding:"required" example:"50000"`
}

// SetCustomField sets or updates a custom field for a pipeline
//
//	@Summary		Set pipeline custom field
//	@Description	Set or update a custom field value for a pipeline
//	@Tags			CRM - Pipelines
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Pipeline ID (UUID)"
//	@Param			request	body		SetCustomFieldRequest	true	"Custom field data"
//	@Success		200		{object}	map[string]interface{}	"Custom field set successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		401		{object}	map[string]interface{}	"Unauthorized"
//	@Failure		404		{object}	map[string]interface{}	"Pipeline not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/crm/pipelines/{id}/custom-fields [post]
func (h *PipelineHandler) SetCustomField(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	pipelineIDStr := c.Param("id")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		apierrors.BadRequest(c, "Invalid pipeline ID format")
		return
	}

	var req SetCustomFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierrors.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// Create use case
	useCase := pipelineapp.NewSetCustomFieldUseCase(h.pipelineRepo, h.logrusLogger)

	// Execute
	cmd := pipelineapp.SetCustomFieldCommand{
		PipelineID: pipelineID,
		TenantID:   authCtx.TenantID,
		Key:        req.Key,
		Type:       shared.FieldType(req.Type),
		Value:      req.Value,
	}

	customField, err := useCase.Execute(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to set custom field", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom field set successfully",
		"custom_field": gin.H{
			"id":         customField.ID(),
			"key":        customField.FieldKey(),
			"type":       customField.FieldType(),
			"value":      customField.FieldValue(),
			"created_at": customField.CreatedAt(),
			"updated_at": customField.UpdatedAt(),
		},
	})
}

// GetCustomFields retrieves all custom fields for a pipeline
//
//	@Summary		Get pipeline custom fields
//	@Description	Retrieve all custom fields for a pipeline
//	@Tags			CRM - Pipelines
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"Pipeline ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"List of custom fields"
//	@Failure		400	{object}	map[string]interface{}	"Invalid pipeline ID"
//	@Failure		401	{object}	map[string]interface{}	"Unauthorized"
//	@Failure		404	{object}	map[string]interface{}	"Pipeline not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/crm/pipelines/{id}/custom-fields [get]
func (h *PipelineHandler) GetCustomFields(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	pipelineIDStr := c.Param("id")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		apierrors.BadRequest(c, "Invalid pipeline ID format")
		return
	}

	// Create use case
	useCase := pipelineapp.NewGetCustomFieldsUseCase(h.pipelineRepo, h.logrusLogger)

	// Execute
	query := pipelineapp.GetCustomFieldsQuery{
		PipelineID: pipelineID,
		TenantID:   authCtx.TenantID,
	}

	fields, err := useCase.Execute(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get custom fields", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	// Convert to response
	response := make([]map[string]interface{}, len(fields))
	for i, field := range fields {
		response[i] = map[string]interface{}{
			"id":         field.ID(),
			"key":        field.FieldKey(),
			"type":       field.FieldType(),
			"value":      field.FieldValue(),
			"created_at": field.CreatedAt(),
			"updated_at": field.UpdatedAt(),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"custom_fields": response,
		"total":         len(response),
	})
}

// RemoveCustomField removes a custom field from a pipeline
//
//	@Summary		Remove pipeline custom field
//	@Description	Remove a custom field from a pipeline by key
//	@Tags			CRM - Pipelines
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string					true	"Pipeline ID (UUID)"
//	@Param			key	path		string					true	"Custom field key"
//	@Success		200	{object}	map[string]interface{}	"Custom field removed successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		401	{object}	map[string]interface{}	"Unauthorized"
//	@Failure		404	{object}	map[string]interface{}	"Pipeline or custom field not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/crm/pipelines/{id}/custom-fields/{key} [delete]
func (h *PipelineHandler) RemoveCustomField(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	pipelineIDStr := c.Param("id")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		apierrors.BadRequest(c, "Invalid pipeline ID format")
		return
	}

	key := c.Param("key")
	if key == "" {
		apierrors.BadRequest(c, "Custom field key is required")
		return
	}

	// Create use case
	useCase := pipelineapp.NewRemoveCustomFieldUseCase(h.pipelineRepo, h.logrusLogger)

	// Execute
	cmd := pipelineapp.RemoveCustomFieldCommand{
		PipelineID: pipelineID,
		TenantID:   authCtx.TenantID,
		Key:        key,
	}

	if err := useCase.Execute(c.Request.Context(), cmd); err != nil {
		h.logger.Error("Failed to remove custom field", zap.Error(err))
		apierrors.RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Custom field removed successfully",
		"key":     key,
	})
}
