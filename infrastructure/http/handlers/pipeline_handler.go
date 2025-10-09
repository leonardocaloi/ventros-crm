package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/helpers"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PipelineHandler struct {
	logger       *zap.Logger
	pipelineRepo pipeline.Repository
}

func NewPipelineHandler(logger *zap.Logger, pipelineRepo pipeline.Repository) *PipelineHandler {
	return &PipelineHandler{
		logger:       logger,
		pipelineRepo: pipelineRepo,
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
//	@Tags			pipelines
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
	// Verificar autenticação
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	ownershipHelper := helpers.NewOwnershipHelper()
	projectID, ok := ownershipHelper.ParseUUID(c, c.Query("project_id"), "project_id")
	if !ok {
		return
	}

	// Verificar se o projeto pertence ao usuário
	if !ownershipHelper.CheckProjectOwnership(c, projectID, authCtx.UserID) {
		ownershipHelper.DenyAccess(c, "Project")
		return
	}

	activeOnly := c.Query("active") == "true"

	var pipelines []*pipeline.Pipeline
	var err error
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
//	@Tags			pipelines
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
//	@Tags			pipelines
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
//	@Tags			pipelines
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
//	@Tags			pipelines
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
//	@Tags			pipelines
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
