package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/helpers"
	"github.com/caloi/ventros-crm/internal/domain/project"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProjectHandler struct {
	logger      *zap.Logger
	projectRepo project.Repository
}

func NewProjectHandler(logger *zap.Logger, projectRepo project.Repository) *ProjectHandler {
	return &ProjectHandler{
		logger:      logger,
		projectRepo: projectRepo,
	}
}

var ownershipHelper = helpers.NewOwnershipHelper()

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
//	@Tags			projects
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
	// Verificar autenticação
	authCtx, ok := ownershipHelper.RequireAuth(c)
	if !ok {
		return
	}

	// TODO: Implementar listagem real filtrada por user_id
	projects := ownershipHelper.GetUserProjects(authCtx.UserID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Projects retrieved successfully",
		"user_id":  authCtx.UserID,
		"projects": projects,
		"count":    len(projects),
	})
}

// CreateProject creates a new project
//
//	@Summary		Create project
//	@Description	Cria um novo projeto
//	@Tags			projects
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
//	@Tags			projects
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
//	@Tags			projects
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
//	@Tags			projects
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
