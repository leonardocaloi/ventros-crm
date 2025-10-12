package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/internal/domain/automation/campaign"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CampaignHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewCampaignHandler(logger *zap.Logger, db *gorm.DB) *CampaignHandler {
	return &CampaignHandler{
		logger: logger,
		db:     db,
	}
}

// ListCampaigns lists all campaigns for the authenticated tenant
//
//	@Summary		List campaigns
//	@Description	Get a paginated list of campaigns with optional filters
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			status	query		string					false	"Filter by status (draft, scheduled, active, paused, completed, archived)"
//	@Param			page	query		int						false	"Page number (default: 1)"
//	@Param			limit	query		int						false	"Items per page (default: 20, max: 100)"
//	@Success		200		{object}	map[string]interface{}	"campaigns: array of campaigns, total: total count, page: current page, limit: items per page"
//	@Failure		400		{object}	map[string]interface{}	"error: validation error message"
//	@Failure		500		{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns [get]
func (h *CampaignHandler) ListCampaigns(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	// Get filter parameters
	statusFilter := c.Query("status")

	repo := persistence.NewGormCampaignRepository(h.db)

	// Get campaigns
	var allCampaigns []*campaign.Campaign
	var err error

	if statusFilter != "" {
		// Filter by status
		status := campaign.CampaignStatus(statusFilter)
		allCampaigns, err = repo.FindActiveByStatus(status)
	} else {
		// Get all for tenant
		allCampaigns, err = repo.FindByTenantID(tenantID)
	}

	if err != nil {
		h.logger.Error("Failed to list campaigns", zap.Error(err))
		errors.InternalError(c, "Failed to list campaigns", err)
		return
	}

	// Pagination (in-memory for simplicity)
	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil || limit < 1 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
	}

	total := len(allCampaigns)
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	paginatedCampaigns := allCampaigns[offset:end]

	// Convert to response
	campaigns := make([]map[string]interface{}, len(paginatedCampaigns))
	for i, camp := range paginatedCampaigns {
		campaigns[i] = h.campaignToResponse(camp)
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// CreateCampaign creates a new campaign
//
//	@Summary		Create campaign
//	@Description	Create a new marketing campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			campaign	body		CreateCampaignRequest	true	"Campaign details"
//	@Success		201			{object}	map[string]interface{}	"campaign: created campaign object"
//	@Failure		400			{object}	map[string]interface{}	"error: validation error message"
//	@Failure		500			{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns [post]
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	var req CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		errors.BadRequest(c, "name is required")
		return
	}
	if req.GoalType == "" {
		errors.BadRequest(c, "goal_type is required")
		return
	}

	// Create campaign
	camp, err := campaign.NewCampaign(
		tenantID,
		req.Name,
		req.Description,
		campaign.GoalType(req.GoalType),
		req.GoalValue,
	)
	if err != nil {
		errors.BadRequest(c, "Failed to create campaign: "+err.Error())
		return
	}

	// Add steps if provided
	for _, stepReq := range req.Steps {
		config := campaign.StepConfig{
			BroadcastID:   stepReq.Config.BroadcastID,
			SequenceID:    stepReq.Config.SequenceID,
			DelayAmount:   stepReq.Config.DelayAmount,
			DelayUnit:     stepReq.Config.DelayUnit,
			ConditionType: stepReq.Config.ConditionType,
			ConditionData: stepReq.Config.ConditionData,
			WaitFor:       stepReq.Config.WaitFor,
			WaitTimeout:   stepReq.Config.WaitTimeout,
			TimeoutStep:   stepReq.Config.TimeoutStep,
		}

		step := campaign.NewCampaignStep(
			stepReq.Order,
			stepReq.Name,
			campaign.StepType(stepReq.Type),
			config,
		)

		// Add conditions
		for _, condReq := range stepReq.Conditions {
			step.AddCondition(campaign.StepCondition{
				Type:     condReq.Type,
				Field:    condReq.Field,
				Operator: condReq.Operator,
				Value:    condReq.Value,
				Metadata: condReq.Metadata,
			})
		}

		if err := camp.AddStep(step); err != nil {
			errors.BadRequest(c, "Failed to add step: "+err.Error())
			return
		}
	}

	// Save to database
	repo := persistence.NewGormCampaignRepository(h.db)
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to save campaign", zap.Error(err))
		errors.InternalError(c, "Failed to create campaign", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"campaign": h.campaignToResponse(camp),
	})
}

// GetCampaign gets a campaign by ID
//
//	@Summary		Get campaign
//	@Description	Get a campaign by ID
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"campaign: campaign object"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id} [get]
func (h *CampaignHandler) GetCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		h.logger.Error("Failed to find campaign", zap.Error(err))
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
	})
}

// UpdateCampaign updates a campaign
//
//	@Summary		Update campaign
//	@Description	Update a campaign (only allowed in draft status)
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Campaign ID"
//	@Param			campaign	body		UpdateCampaignRequest	true	"Updated campaign details"
//	@Success		200			{object}	map[string]interface{}	"campaign: updated campaign object"
//	@Failure		400			{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404			{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500			{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id} [put]
func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	var req UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Update name
	if req.Name != nil {
		if err := camp.UpdateName(*req.Name); err != nil {
			errors.BadRequest(c, err.Error())
			return
		}
	}

	// Update description
	if req.Description != nil {
		camp.UpdateDescription(*req.Description)
	}

	// Update goal
	if req.GoalType != nil && req.GoalValue != nil {
		if err := camp.UpdateGoal(campaign.GoalType(*req.GoalType), *req.GoalValue); err != nil {
			errors.BadRequest(c, err.Error())
			return
		}
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to update campaign", zap.Error(err))
		errors.InternalError(c, "Failed to update campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
	})
}

// ActivateCampaign activates a campaign
//
//	@Summary		Activate campaign
//	@Description	Activate a campaign to start execution
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"campaign: activated campaign object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/activate [post]
func (h *CampaignHandler) ActivateCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Activate
	if err := camp.Activate(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to activate campaign", zap.Error(err))
		errors.InternalError(c, "Failed to activate campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
		"message":  "Campaign activated successfully",
	})
}

// ScheduleCampaign schedules a campaign
//
//	@Summary		Schedule campaign
//	@Description	Schedule a campaign to start at a specific time
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Campaign ID"
//	@Param			schedule	body		ScheduleCampaignRequest	true	"Schedule details"
//	@Success		200			{object}	map[string]interface{}	"campaign: scheduled campaign object"
//	@Failure		400			{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404			{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500			{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/schedule [post]
func (h *CampaignHandler) ScheduleCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	var req ScheduleCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Parse start date
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		errors.BadRequest(c, "invalid start_date format, use RFC3339")
		return
	}

	// Schedule
	if err := camp.Schedule(startDate); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to schedule campaign", zap.Error(err))
		errors.InternalError(c, "Failed to schedule campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
		"message":  "Campaign scheduled successfully",
	})
}

// PauseCampaign pauses a campaign
//
//	@Summary		Pause campaign
//	@Description	Pause an active campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"campaign: paused campaign object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/pause [post]
func (h *CampaignHandler) PauseCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Pause
	if err := camp.Pause(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to pause campaign", zap.Error(err))
		errors.InternalError(c, "Failed to pause campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
		"message":  "Campaign paused successfully",
	})
}

// ResumeCampaign resumes a paused campaign
//
//	@Summary		Resume campaign
//	@Description	Resume a paused campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"campaign: resumed campaign object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/resume [post]
func (h *CampaignHandler) ResumeCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Resume
	if err := camp.Resume(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to resume campaign", zap.Error(err))
		errors.InternalError(c, "Failed to resume campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
		"message":  "Campaign resumed successfully",
	})
}

// CompleteCampaign marks a campaign as completed
//
//	@Summary		Complete campaign
//	@Description	Mark a campaign as completed
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"campaign: completed campaign object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/complete [post]
func (h *CampaignHandler) CompleteCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Complete
	if err := camp.Complete(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to complete campaign", zap.Error(err))
		errors.InternalError(c, "Failed to complete campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
		"message":  "Campaign completed successfully",
	})
}

// ArchiveCampaign archives a campaign
//
//	@Summary		Archive campaign
//	@Description	Archive a campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"campaign: archived campaign object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/archive [post]
func (h *CampaignHandler) ArchiveCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Archive
	if err := camp.Archive(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(camp); err != nil {
		h.logger.Error("Failed to archive campaign", zap.Error(err))
		errors.InternalError(c, "Failed to archive campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaign": h.campaignToResponse(camp),
		"message":  "Campaign archived successfully",
	})
}

// DeleteCampaign deletes a campaign
//
//	@Summary		Delete campaign
//	@Description	Delete a campaign (only in draft status)
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"message: success message"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id} [delete]
func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Only allow deletion of draft campaigns
	if camp.Status() != campaign.CampaignStatusDraft {
		errors.BadRequest(c, "Can only delete campaigns in draft status")
		return
	}

	// Delete
	if err := repo.Delete(campaignID); err != nil {
		h.logger.Error("Failed to delete campaign", zap.Error(err))
		errors.InternalError(c, "Failed to delete campaign", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Campaign deleted successfully",
	})
}

// GetCampaignStats gets campaign statistics
//
//	@Summary		Get campaign stats
//	@Description	Get statistics for a campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"stats: campaign statistics object"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/stats [get]
func (h *CampaignHandler) GetCampaignStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	repo := persistence.NewGormCampaignRepository(h.db)
	camp, err := repo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	stats := camp.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"stats": map[string]interface{}{
			"contacts_reached":  stats.ContactsReached,
			"conversions_count": stats.ConversionsCount,
			"conversion_rate":   stats.ConversionRate,
			"progress_rate":     stats.ProgressRate,
		},
	})
}

// EnrollContact enrolls a contact in a campaign
//
//	@Summary		Enroll contact
//	@Description	Enroll a contact in a campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Campaign ID"
//	@Param			enrollment	body		EnrollContactRequest	true	"Enrollment details"
//	@Success		201			{object}	map[string]interface{}	"enrollment: created enrollment object"
//	@Failure		400			{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404			{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500			{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/enroll [post]
func (h *CampaignHandler) EnrollContact(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	var req EnrollContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if req.ContactID == "" {
		errors.BadRequest(c, "contact_id is required")
		return
	}

	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		errors.BadRequest(c, "invalid contact_id format")
		return
	}

	// Check if campaign exists
	campRepo := persistence.NewGormCampaignRepository(h.db)
	camp, err := campRepo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check tenant ownership
	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Check if campaign is active
	if camp.Status() != campaign.CampaignStatusActive {
		errors.BadRequest(c, "Campaign must be active to enroll contacts")
		return
	}

	// Check if already enrolled
	enrollmentRepo := persistence.NewGormCampaignEnrollmentRepository(h.db)
	existing, err := enrollmentRepo.FindActiveByCampaignAndContact(campaignID, contactID)
	if err != nil {
		h.logger.Error("Failed to check existing enrollment", zap.Error(err))
		errors.InternalError(c, "Failed to check enrollment", err)
		return
	}
	if existing != nil {
		errors.BadRequest(c, "Contact is already enrolled in this campaign")
		return
	}

	// Get first step delay
	firstStep, err := camp.GetStepByOrder(0)
	if err != nil || firstStep == nil {
		errors.BadRequest(c, "Campaign has no steps")
		return
	}

	// Create enrollment
	enrollment, err := campaign.NewCampaignEnrollment(
		campaignID,
		contactID,
		firstStep.GetDelayDuration(),
	)
	if err != nil {
		errors.BadRequest(c, "Failed to create enrollment: "+err.Error())
		return
	}

	// Save enrollment
	if err := enrollmentRepo.Save(enrollment); err != nil {
		h.logger.Error("Failed to save enrollment", zap.Error(err))
		errors.InternalError(c, "Failed to enroll contact", err)
		return
	}

	// Update campaign stats
	camp.IncrementContactsReached()
	if err := campRepo.Save(camp); err != nil {
		h.logger.Error("Failed to update campaign stats", zap.Error(err))
	}

	c.JSON(http.StatusCreated, gin.H{
		"enrollment": h.enrollmentToResponse(enrollment),
		"message":    "Contact enrolled successfully",
	})
}

// ListEnrollments lists all enrollments for a campaign
//
//	@Summary		List enrollments
//	@Description	Get all enrollments for a campaign
//	@Tags			AUTOMATION - Campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Campaign ID"
//	@Success		200	{object}	map[string]interface{}	"enrollments: array of enrollments"
//	@Failure		404	{object}	map[string]interface{}	"error: campaign not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/campaigns/{id}/enrollments [get]
func (h *CampaignHandler) ListEnrollments(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	campaignID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid campaign id")
		return
	}

	// Check campaign exists and ownership
	campRepo := persistence.NewGormCampaignRepository(h.db)
	camp, err := campRepo.FindByID(campaignID)
	if err != nil {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	if camp.TenantID() != tenantID {
		errors.NotFound(c, "campaign", campaignID.String())
		return
	}

	// Get enrollments
	enrollmentRepo := persistence.NewGormCampaignEnrollmentRepository(h.db)
	enrollments, err := enrollmentRepo.FindByCampaignID(campaignID)
	if err != nil {
		h.logger.Error("Failed to list enrollments", zap.Error(err))
		errors.InternalError(c, "Failed to list enrollments", err)
		return
	}

	// Convert to response
	enrollmentResponses := make([]map[string]interface{}, len(enrollments))
	for i, enrollment := range enrollments {
		enrollmentResponses[i] = h.enrollmentToResponse(enrollment)
	}

	c.JSON(http.StatusOK, gin.H{
		"enrollments": enrollmentResponses,
		"total":       len(enrollments),
	})
}

// Helper methods
func (h *CampaignHandler) campaignToResponse(camp *campaign.Campaign) map[string]interface{} {
	// Convert steps to response format
	steps := camp.Steps()
	stepResponses := make([]map[string]interface{}, len(steps))
	for i, step := range steps {
		stepResponses[i] = map[string]interface{}{
			"id":         step.ID,
			"order":      step.Order,
			"name":       step.Name,
			"type":       step.Type,
			"config":     step.Config,
			"conditions": step.Conditions,
			"created_at": step.CreatedAt,
		}
	}

	return map[string]interface{}{
		"id":                camp.ID(),
		"tenant_id":         camp.TenantID(),
		"name":              camp.Name(),
		"description":       camp.Description(),
		"status":            camp.Status(),
		"goal_type":         camp.GoalType(),
		"goal_value":        camp.GoalValue(),
		"contacts_reached":  camp.ContactsReached(),
		"conversions_count": camp.ConversionsCount(),
		"start_date":        camp.StartDate(),
		"end_date":          camp.EndDate(),
		"steps":             stepResponses,
		"created_at":        camp.CreatedAt(),
		"updated_at":        camp.UpdatedAt(),
	}
}

func (h *CampaignHandler) enrollmentToResponse(enrollment *campaign.CampaignEnrollment) map[string]interface{} {
	return map[string]interface{}{
		"id":                 enrollment.ID(),
		"campaign_id":        enrollment.CampaignID(),
		"contact_id":         enrollment.ContactID(),
		"status":             enrollment.Status(),
		"current_step_order": enrollment.CurrentStepOrder(),
		"next_scheduled_at":  enrollment.NextScheduledAt(),
		"exited_at":          enrollment.ExitedAt(),
		"exit_reason":        enrollment.ExitReason(),
		"completed_at":       enrollment.CompletedAt(),
		"enrolled_at":        enrollment.EnrolledAt(),
		"updated_at":         enrollment.UpdatedAt(),
	}
}

// Request/Response types
type CreateCampaignRequest struct {
	Name        string                      `json:"name" binding:"required"`
	Description string                      `json:"description"`
	GoalType    string                      `json:"goal_type" binding:"required"`
	GoalValue   int                         `json:"goal_value"`
	Steps       []CreateCampaignStepRequest `json:"steps"`
}

type CreateCampaignStepRequest struct {
	Order      int                    `json:"order" binding:"required"`
	Name       string                 `json:"name" binding:"required"`
	Type       string                 `json:"type" binding:"required"`
	Config     StepConfigRequest      `json:"config" binding:"required"`
	Conditions []StepConditionRequest `json:"conditions"`
}

type StepConfigRequest struct {
	BroadcastID   *uuid.UUID             `json:"broadcast_id,omitempty"`
	SequenceID    *uuid.UUID             `json:"sequence_id,omitempty"`
	DelayAmount   *int                   `json:"delay_amount,omitempty"`
	DelayUnit     *string                `json:"delay_unit,omitempty"`
	ConditionType *string                `json:"condition_type,omitempty"`
	ConditionData map[string]interface{} `json:"condition_data,omitempty"`
	WaitFor       *string                `json:"wait_for,omitempty"`
	WaitTimeout   *int                   `json:"wait_timeout,omitempty"`
	TimeoutStep   *int                   `json:"timeout_step,omitempty"`
}

type StepConditionRequest struct {
	Type     string                 `json:"type"`
	Field    string                 `json:"field"`
	Operator string                 `json:"operator"`
	Value    interface{}            `json:"value"`
	Metadata map[string]interface{} `json:"metadata"`
}

type UpdateCampaignRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	GoalType    *string `json:"goal_type"`
	GoalValue   *int    `json:"goal_value"`
}

type ScheduleCampaignRequest struct {
	StartDate string `json:"start_date" binding:"required"`
}
