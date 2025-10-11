package handlers

import (
	"context"
	"net/http"
	"strconv"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AutomationHandler handles cross-product automation endpoints
type AutomationHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewAutomationHandler(logger *zap.Logger, db *gorm.DB) *AutomationHandler {
	return &AutomationHandler{
		logger: logger,
		db:     db,
	}
}

// CreateAutomationRequest represents the payload to create an automation
type CreateAutomationRequest struct {
	AutomationType string                   `json:"automation_type" binding:"required" example:"event"`
	PipelineID     *uuid.UUID               `json:"pipeline_id,omitempty"`
	Name           string                   `json:"name" binding:"required" example:"Send welcome message"`
	Description    string                   `json:"description" example:"Automatically send welcome message to new contacts"`
	Trigger        string                   `json:"trigger" binding:"required" example:"contact.created"`
	Conditions     []pipeline.RuleCondition `json:"conditions,omitempty"`
	Actions        []pipeline.RuleAction    `json:"actions" binding:"required"`
	Priority       int                      `json:"priority" example:"10"`
	Enabled        bool                     `json:"enabled" example:"true"`
}

// UpdateAutomationRequest represents the payload to update an automation
type UpdateAutomationRequest struct {
	Name        *string                   `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	Conditions  *[]pipeline.RuleCondition `json:"conditions,omitempty"`
	Actions     *[]pipeline.RuleAction    `json:"actions,omitempty"`
	Priority    *int                      `json:"priority,omitempty"`
	Enabled     *bool                     `json:"enabled,omitempty"`
}

// ListAutomations lists all automations for the authenticated tenant
//
//	@Summary		List all automations
//	@Description	Lista todas as automações (cross-product) do tenant autenticado com filtros avançados
//	@Tags			AUTOMATION - Automations
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			automation_type	query		string					false	"Filter by automation type"
//	@Param			enabled			query		bool					false	"Filter by enabled status"
//	@Param			pipeline_id		query		string					false	"Filter by pipeline ID (UUID)"
//	@Param			page			query		int						false	"Page number"		default(1)
//	@Param			limit			query		int						false	"Items per page"	default(20)
//	@Success		200				{object}	map[string]interface{}	"List of automations"
//	@Failure		401				{object}	map[string]interface{}	"Not authenticated"
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/automation [get]
func (h *AutomationHandler) ListAutomations(c *gin.Context) {
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

	// Parse filters
	automationType := c.Query("automation_type")
	enabledStr := c.Query("enabled")
	pipelineIDStr := c.Query("pipeline_id")

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

	offset := (page - 1) * limit

	// Build query
	query := h.db.Table("automations").Where("tenant_id = ?", tenantID.String())

	if automationType != "" {
		query = query.Where("automation_type = ?", automationType)
	}

	if enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			query = query.Where("enabled = ?", enabled)
		}
	}

	if pipelineIDStr != "" {
		if pipelineID, err := uuid.Parse(pipelineIDStr); err == nil {
			query = query.Where("pipeline_id = ?", pipelineID)
		}
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.logger.Error("Failed to count automations", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve automations", err)
		return
	}

	// Get automations
	type AutomationRow struct {
		ID             uuid.UUID  `json:"id"`
		AutomationType string     `json:"automation_type"`
		PipelineID     *uuid.UUID `json:"pipeline_id"`
		TenantID       string     `json:"tenant_id"`
		Name           string     `json:"name"`
		Description    string     `json:"description"`
		Trigger        string     `json:"trigger"`
		Priority       int        `json:"priority"`
		Enabled        bool       `json:"enabled"`
		CreatedAt      string     `json:"created_at"`
		UpdatedAt      string     `json:"updated_at"`
	}

	var automations []AutomationRow
	if err := query.
		Order("priority ASC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&automations).Error; err != nil {
		h.logger.Error("Failed to list automations", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve automations", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": automations,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetAutomation gets an automation by ID
//
//	@Summary		Get automation by ID
//	@Description	Obtém detalhes de uma automação específica
//	@Tags			AUTOMATION - Automations
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Automation ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Automation details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid automation ID"
//	@Failure		404	{object}	map[string]interface{}	"Automation not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/automation/{id} [get]
func (h *AutomationHandler) GetAutomation(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	automationID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid automation ID", zap.String("id", idStr), zap.Error(err))
		apierrors.ValidationError(c, "id", "Invalid automation ID format")
		return
	}

	type AutomationDetail struct {
		ID             uuid.UUID                `json:"id"`
		AutomationType string                   `json:"automation_type"`
		PipelineID     *uuid.UUID               `json:"pipeline_id"`
		TenantID       string                   `json:"tenant_id"`
		Name           string                   `json:"name"`
		Description    string                   `json:"description"`
		Trigger        string                   `json:"trigger"`
		Conditions     []pipeline.RuleCondition `json:"conditions"`
		Actions        []pipeline.RuleAction    `json:"actions"`
		Priority       int                      `json:"priority"`
		Enabled        bool                     `json:"enabled"`
		CreatedAt      string                   `json:"created_at"`
		UpdatedAt      string                   `json:"updated_at"`
	}

	var automation AutomationDetail
	if err := h.db.Table("automations").
		Where("id = ? AND tenant_id = ?", automationID, authCtx.TenantID).
		First(&automation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierrors.NotFound(c, "automation", automationID.String())
			return
		}
		h.logger.Error("Failed to get automation", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve automation", err)
		return
	}

	c.JSON(http.StatusOK, automation)
}

// CreateAutomation creates a new automation
//
//	@Summary		Create automation
//	@Description	Cria uma nova automação (cross-product ou pipeline-specific)
//	@Tags			AUTOMATION - Automations
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			automation	body		CreateAutomationRequest	true	"Automation data"
//	@Success		201			{object}	map[string]interface{}	"Automation created successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/automation [post]
func (h *AutomationHandler) CreateAutomation(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	var req CreateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse automation request", zap.Error(err))
		apierrors.ValidationError(c, "body", "Invalid request: "+err.Error())
		return
	}

	// Create domain automation
	automation, err := pipeline.NewAutomation(
		pipeline.AutomationType(req.AutomationType),
		authCtx.TenantID,
		req.Name,
		pipeline.AutomationTrigger(req.Trigger),
		req.PipelineID,
	)
	if err != nil {
		h.logger.Error("Failed to create domain automation", zap.Error(err))
		apierrors.ValidationError(c, "automation", err.Error())
		return
	}

	// Set optional fields
	if req.Description != "" {
		automation.UpdateDescription(req.Description)
	}

	if len(req.Conditions) > 0 {
		automation.SetConditions(req.Conditions)
	}

	if len(req.Actions) > 0 {
		automation.SetActions(req.Actions)
	}

	if req.Priority > 0 {
		if err := automation.SetPriority(req.Priority); err != nil {
			apierrors.ValidationError(c, "priority", err.Error())
			return
		}
	}

	if !req.Enabled {
		automation.Disable()
	}

	// Save to database using repository
	repo := &GormAutomationRepository{db: h.db}
	if err := repo.Save(context.Background(), automation); err != nil {
		h.logger.Error("Failed to save automation", zap.Error(err))
		apierrors.InternalError(c, "Failed to create automation", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Automation created successfully",
		"id":      automation.ID(),
		"name":    automation.Name(),
		"type":    automation.Type(),
		"enabled": automation.IsEnabled(),
	})
}

// UpdateAutomation updates an automation
//
//	@Summary		Update automation
//	@Description	Atualiza uma automação existente
//	@Tags			AUTOMATION - Automations
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id			path		string					true	"Automation ID (UUID)"
//	@Param			automation	body		UpdateAutomationRequest	true	"Automation data"
//	@Success		200			{object}	map[string]interface{}	"Automation updated successfully"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		404			{object}	map[string]interface{}	"Automation not found"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/automation/{id} [put]
func (h *AutomationHandler) UpdateAutomation(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	automationID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid automation ID format")
		return
	}

	var req UpdateAutomationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		apierrors.ValidationError(c, "body", "Invalid request: "+err.Error())
		return
	}

	// Load automation from repository
	repo := &GormAutomationRepository{db: h.db}
	automation, err := repo.FindByID(context.Background(), automationID)
	if err != nil {
		h.logger.Error("Failed to find automation", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve automation", err)
		return
	}

	if automation == nil || automation.TenantID() != authCtx.TenantID {
		apierrors.NotFound(c, "automation", automationID.String())
		return
	}

	// Apply updates
	if req.Description != nil {
		automation.UpdateDescription(*req.Description)
	}

	if req.Conditions != nil {
		automation.SetConditions(*req.Conditions)
	}

	if req.Actions != nil {
		automation.SetActions(*req.Actions)
	}

	if req.Priority != nil {
		if err := automation.SetPriority(*req.Priority); err != nil {
			apierrors.ValidationError(c, "priority", err.Error())
			return
		}
	}

	if req.Enabled != nil {
		if *req.Enabled {
			automation.Enable()
		} else {
			automation.Disable()
		}
	}

	// Save updates
	if err := repo.Save(context.Background(), automation); err != nil {
		h.logger.Error("Failed to update automation", zap.Error(err))
		apierrors.InternalError(c, "Failed to update automation", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Automation updated successfully",
		"id":      automation.ID(),
	})
}

// DeleteAutomation deletes an automation
//
//	@Summary		Delete automation
//	@Description	Deleta uma automação
//	@Tags			AUTOMATION - Automations
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Automation ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Automation deleted successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid automation ID"
//	@Failure		404	{object}	map[string]interface{}	"Automation not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/automation/{id} [delete]
func (h *AutomationHandler) DeleteAutomation(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	idStr := c.Param("id")
	automationID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid automation ID format")
		return
	}

	// Verify ownership before deleting
	repo := &GormAutomationRepository{db: h.db}
	automation, err := repo.FindByID(context.Background(), automationID)
	if err != nil {
		h.logger.Error("Failed to find automation", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve automation", err)
		return
	}

	if automation == nil || automation.TenantID() != authCtx.TenantID {
		apierrors.NotFound(c, "automation", automationID.String())
		return
	}

	// Delete
	if err := repo.Delete(context.Background(), automationID); err != nil {
		h.logger.Error("Failed to delete automation", zap.Error(err))
		apierrors.InternalError(c, "Failed to delete automation", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Automation deleted successfully",
		"id":      automationID,
	})
}

// GetAutomationTypes returns all available automation types with descriptions
//
//	@Summary		Get automation types
//	@Description	Lista todos os tipos de automação disponíveis
//	@Tags			AUTOMATION - Automations
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Automation types"
//	@Router			/api/v1/automation/types [get]
func (h *AutomationHandler) GetAutomationTypes(c *gin.Context) {
	types := []map[string]interface{}{
		{
			"code":              string(pipeline.AutomationTypePipelineBased),
			"name":              "Pipeline-Based",
			"description":       "Automação vinculada a pipeline específico",
			"requires_pipeline": true,
			"category":          "crm",
		},
		{
			"code":              string(pipeline.AutomationTypeFollowUp),
			"name":              "Follow-Up",
			"description":       "Automação de follow-up em pipeline",
			"requires_pipeline": true,
			"category":          "crm",
		},
		{
			"code":              string(pipeline.AutomationTypeReengagement),
			"name":              "Reengagement",
			"description":       "Reativar contatos inativos",
			"requires_pipeline": true,
			"category":          "crm",
		},
		{
			"code":              string(pipeline.AutomationTypeOnboarding),
			"name":              "Onboarding",
			"description":       "Processo de onboarding automatizado",
			"requires_pipeline": true,
			"category":          "crm",
		},
		{
			"code":              string(pipeline.AutomationTypeEvent),
			"name":              "Event-Driven",
			"description":       "Disparada por eventos do sistema",
			"requires_pipeline": false,
			"category":          "automation",
		},
		{
			"code":              string(pipeline.AutomationTypeScheduled),
			"name":              "Scheduled",
			"description":       "Executada em horários específicos",
			"requires_pipeline": false,
			"category":          "automation",
		},
		{
			"code":              string(pipeline.AutomationTypeScheduledReport),
			"name":              "Scheduled Report",
			"description":       "Relatórios periódicos automáticos",
			"requires_pipeline": false,
			"category":          "automation",
		},
		{
			"code":              string(pipeline.AutomationTypeTimeNotification),
			"name":              "Time Notification",
			"description":       "Notificações baseadas em tempo",
			"requires_pipeline": false,
			"category":          "automation",
		},
		{
			"code":              string(pipeline.AutomationTypeWebhook),
			"name":              "Webhook",
			"description":       "Integração via webhook",
			"requires_pipeline": false,
			"category":          "automation",
		},
		{
			"code":              string(pipeline.AutomationTypeCustom),
			"name":              "Custom",
			"description":       "Automação customizada",
			"requires_pipeline": false,
			"category":          "automation",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"types": types,
	})
}

// GetAvailableActions returns all available automation actions
//
//	@Summary		Get available actions
//	@Description	Lista todas as ações disponíveis para automações
//	@Tags			AUTOMATION - Automations
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Available actions"
//	@Router			/api/v1/automation/actions [get]
func (h *AutomationHandler) GetAvailableActions(c *gin.Context) {
	actions := pipeline.GetAvailableActions()
	c.JSON(http.StatusOK, gin.H{
		"actions": actions,
	})
}

// GetAvailableOperators returns all available condition operators
//
//	@Summary		Get available operators
//	@Description	Lista todos os operadores disponíveis para condições
//	@Tags			AUTOMATION - Automations
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Available operators"
//	@Router			/api/v1/automation/operators [get]
func (h *AutomationHandler) GetAvailableOperators(c *gin.Context) {
	operators := pipeline.GetAvailableOperators()
	c.JSON(http.StatusOK, gin.H{
		"operators": operators,
	})
}

// GormAutomationRepository is a minimal repository adapter for the handler
type GormAutomationRepository struct {
	db *gorm.DB
}

func (r *GormAutomationRepository) Save(ctx context.Context, automation *pipeline.Automation) error {
	// Use the existing repository implementation
	repo := persistence.NewGormAutomationRuleRepository(r.db)
	return repo.Save(automation)
}

func (r *GormAutomationRepository) FindByID(ctx context.Context, id uuid.UUID) (*pipeline.Automation, error) {
	repo := persistence.NewGormAutomationRuleRepository(r.db)
	return repo.FindByID(id)
}

func (r *GormAutomationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	repo := persistence.NewGormAutomationRuleRepository(r.db)
	return repo.Delete(id)
}
