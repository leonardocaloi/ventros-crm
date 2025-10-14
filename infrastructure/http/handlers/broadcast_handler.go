package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/http/errors"
	"github.com/ventros/crm/infrastructure/persistence"
	"github.com/ventros/crm/internal/domain/automation/broadcast"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BroadcastHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewBroadcastHandler(logger *zap.Logger, db *gorm.DB) *BroadcastHandler {
	return &BroadcastHandler{
		logger: logger,
		db:     db,
	}
}

// ListBroadcasts lists all broadcasts for the authenticated tenant
//
//	@Summary		List broadcasts
//	@Description	Get a paginated list of broadcasts with optional filters
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			status	query		string					false	"Filter by status (draft, scheduled, running, completed, failed, cancelled)"
//	@Param			page	query		int						false	"Page number (default: 1)"
//	@Param			limit	query		int						false	"Items per page (default: 20, max: 100)"
//	@Success		200		{object}	map[string]interface{}	"broadcasts: array of broadcasts, total: total count, page: current page, limit: items per page"
//	@Failure		400		{object}	map[string]interface{}	"error: validation error message"
//	@Failure		500		{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts [get]
func (h *BroadcastHandler) ListBroadcasts(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	// Get filter parameters
	statusFilter := c.Query("status")

	repo := persistence.NewGormBroadcastRepository(h.db)

	// Get broadcasts
	var allBroadcasts []*broadcast.Broadcast
	var err error

	if statusFilter != "" {
		// Filter by status
		status := broadcast.BroadcastStatus(statusFilter)
		allBroadcasts, err = repo.FindByStatus(status)
	} else {
		// Get all for tenant
		allBroadcasts, err = repo.FindByTenantID(tenantID)
	}

	if err != nil {
		h.logger.Error("Failed to list broadcasts", zap.Error(err))
		errors.InternalError(c, "Failed to list broadcasts", err)
		return
	}

	// Pagination (in-memory for simplicity, can be optimized with DB pagination)
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

	total := len(allBroadcasts)
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	paginatedBroadcasts := allBroadcasts[offset:end]

	// Convert to response
	broadcasts := make([]map[string]interface{}, len(paginatedBroadcasts))
	for i, bc := range paginatedBroadcasts {
		broadcasts[i] = h.broadcastToResponse(bc)
	}

	c.JSON(http.StatusOK, gin.H{
		"broadcasts": broadcasts,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// CreateBroadcast creates a new broadcast
//
//	@Summary		Create broadcast
//	@Description	Create a new broadcast for mass messaging
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			broadcast	body		CreateBroadcastRequest	true	"Broadcast details"
//	@Success		201			{object}	map[string]interface{}	"broadcast: created broadcast object"
//	@Failure		400			{object}	map[string]interface{}	"error: validation error message"
//	@Failure		500			{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts [post]
func (h *BroadcastHandler) CreateBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	var req CreateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		errors.BadRequest(c, "name is required")
		return
	}
	if req.ListID == "" {
		errors.BadRequest(c, "list_id is required")
		return
	}
	if req.MessageTemplate.Type == "" {
		errors.BadRequest(c, "message_template.type is required")
		return
	}
	if req.MessageTemplate.Content == "" && req.MessageTemplate.TemplateID == nil {
		errors.BadRequest(c, "message_template must have either content or template_id")
		return
	}

	listID, err := uuid.Parse(req.ListID)
	if err != nil {
		errors.BadRequest(c, "invalid list_id format")
		return
	}

	// Create message template
	template := broadcast.MessageTemplate{
		Type:       req.MessageTemplate.Type,
		Content:    req.MessageTemplate.Content,
		TemplateID: req.MessageTemplate.TemplateID,
		Variables:  req.MessageTemplate.Variables,
		MediaURL:   req.MessageTemplate.MediaURL,
	}

	// Create broadcast
	bc, err := broadcast.NewBroadcast(tenantID, req.Name, listID, template, req.RateLimit)
	if err != nil {
		errors.BadRequest(c, "Failed to create broadcast: "+err.Error())
		return
	}

	// Save to database
	repo := persistence.NewGormBroadcastRepository(h.db)
	if err := repo.Save(bc); err != nil {
		h.logger.Error("Failed to save broadcast", zap.Error(err))
		errors.InternalError(c, "Failed to create broadcast", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"broadcast": h.broadcastToResponse(bc),
	})
}

// GetBroadcast gets a broadcast by ID
//
//	@Summary		Get broadcast
//	@Description	Get a broadcast by ID
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Broadcast ID"
//	@Success		200	{object}	map[string]interface{}	"broadcast: broadcast object"
//	@Failure		404	{object}	map[string]interface{}	"error: broadcast not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id} [get]
func (h *BroadcastHandler) GetBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		h.logger.Error("Failed to find broadcast", zap.Error(err))
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"broadcast": h.broadcastToResponse(bc),
	})
}

// UpdateBroadcast updates a broadcast
//
//	@Summary		Update broadcast
//	@Description	Update a broadcast (only allowed in draft status)
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Broadcast ID"
//	@Param			broadcast	body		UpdateBroadcastRequest	true	"Updated broadcast details"
//	@Success		200			{object}	map[string]interface{}	"broadcast: updated broadcast object"
//	@Failure		400			{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404			{object}	map[string]interface{}	"error: broadcast not found"
//	@Failure		500			{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id} [put]
func (h *BroadcastHandler) UpdateBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	var req UpdateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Update name
	if req.Name != nil {
		if err := bc.UpdateName(*req.Name); err != nil {
			errors.BadRequest(c, err.Error())
			return
		}
	}

	// Update message template
	if req.MessageTemplate != nil {
		template := broadcast.MessageTemplate{
			Type:       req.MessageTemplate.Type,
			Content:    req.MessageTemplate.Content,
			TemplateID: req.MessageTemplate.TemplateID,
			Variables:  req.MessageTemplate.Variables,
			MediaURL:   req.MessageTemplate.MediaURL,
		}
		if err := bc.UpdateMessageTemplate(template); err != nil {
			errors.BadRequest(c, err.Error())
			return
		}
	}

	// Update rate limit
	if req.RateLimit != nil {
		if err := bc.UpdateRateLimit(*req.RateLimit); err != nil {
			errors.BadRequest(c, err.Error())
			return
		}
	}

	// Save
	if err := repo.Save(bc); err != nil {
		h.logger.Error("Failed to update broadcast", zap.Error(err))
		errors.InternalError(c, "Failed to update broadcast", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"broadcast": h.broadcastToResponse(bc),
	})
}

// ScheduleBroadcast schedules a broadcast for execution
//
//	@Summary		Schedule broadcast
//	@Description	Schedule a broadcast for a specific date/time
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string						true	"Broadcast ID"
//	@Param			schedule	body		ScheduleBroadcastRequest	true	"Schedule details"
//	@Success		200			{object}	map[string]interface{}		"broadcast: scheduled broadcast object"
//	@Failure		400			{object}	map[string]interface{}		"error: validation error message"
//	@Failure		404			{object}	map[string]interface{}		"error: broadcast not found"
//	@Failure		500			{object}	map[string]interface{}		"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id}/schedule [post]
func (h *BroadcastHandler) ScheduleBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	var req ScheduleBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Schedule
	if err := bc.Schedule(req.ScheduledFor); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(bc); err != nil {
		h.logger.Error("Failed to schedule broadcast", zap.Error(err))
		errors.InternalError(c, "Failed to schedule broadcast", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"broadcast": h.broadcastToResponse(bc),
	})
}

// ExecuteBroadcast starts executing a broadcast immediately
//
//	@Summary		Execute broadcast
//	@Description	Start executing a broadcast immediately
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Broadcast ID"
//	@Success		200	{object}	map[string]interface{}	"broadcast: executing broadcast object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: broadcast not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id}/execute [post]
func (h *BroadcastHandler) ExecuteBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Start execution
	if err := bc.Start(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(bc); err != nil {
		h.logger.Error("Failed to execute broadcast", zap.Error(err))
		errors.InternalError(c, "Failed to execute broadcast", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"broadcast": h.broadcastToResponse(bc),
		"message":   "Broadcast execution started",
	})
}

// CancelBroadcast cancels a broadcast
//
//	@Summary		Cancel broadcast
//	@Description	Cancel a broadcast (if not completed)
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Broadcast ID"
//	@Success		200	{object}	map[string]interface{}	"broadcast: cancelled broadcast object"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: broadcast not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id}/cancel [post]
func (h *BroadcastHandler) CancelBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Cancel
	if err := bc.Cancel(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(bc); err != nil {
		h.logger.Error("Failed to cancel broadcast", zap.Error(err))
		errors.InternalError(c, "Failed to cancel broadcast", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"broadcast": h.broadcastToResponse(bc),
		"message":   "Broadcast cancelled",
	})
}

// GetBroadcastStats gets broadcast statistics
//
//	@Summary		Get broadcast stats
//	@Description	Get statistics for a broadcast
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Broadcast ID"
//	@Success		200	{object}	map[string]interface{}	"stats: broadcast statistics object"
//	@Failure		404	{object}	map[string]interface{}	"error: broadcast not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id}/stats [get]
func (h *BroadcastHandler) GetBroadcastStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	stats := bc.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"stats": map[string]interface{}{
			"total_contacts": stats.TotalContacts,
			"sent_count":     stats.SentCount,
			"failed_count":   stats.FailedCount,
			"pending_count":  stats.PendingCount,
			"progress":       stats.Progress,
		},
	})
}

// DeleteBroadcast deletes a broadcast
//
//	@Summary		Delete broadcast
//	@Description	Delete a broadcast (only in draft or cancelled status)
//	@Tags			AUTOMATION - Broadcasts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string					true	"Broadcast ID"
//	@Success		200	{object}	map[string]interface{}	"message: success message"
//	@Failure		400	{object}	map[string]interface{}	"error: validation error message"
//	@Failure		404	{object}	map[string]interface{}	"error: broadcast not found"
//	@Failure		500	{object}	map[string]interface{}	"error: internal server error message"
//	@Router			/api/v1/automation/broadcasts/{id} [delete]
func (h *BroadcastHandler) DeleteBroadcast(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	broadcastID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid broadcast id")
		return
	}

	repo := persistence.NewGormBroadcastRepository(h.db)
	bc, err := repo.FindByID(broadcastID)
	if err != nil {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Check tenant ownership
	if bc.TenantID() != tenantID {
		errors.NotFound(c, "broadcast", broadcastID.String())
		return
	}

	// Only allow deletion of draft or cancelled broadcasts
	if bc.Status() != broadcast.BroadcastStatusDraft && bc.Status() != broadcast.BroadcastStatusCancelled {
		errors.BadRequest(c, "Can only delete broadcasts in draft or cancelled status")
		return
	}

	// Delete
	if err := repo.Delete(broadcastID); err != nil {
		h.logger.Error("Failed to delete broadcast", zap.Error(err))
		errors.InternalError(c, "Failed to delete broadcast", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Broadcast deleted successfully",
	})
}

// Helper methods
func (h *BroadcastHandler) broadcastToResponse(bc *broadcast.Broadcast) map[string]interface{} {
	return map[string]interface{}{
		"id":               bc.ID(),
		"tenant_id":        bc.TenantID(),
		"name":             bc.Name(),
		"list_id":          bc.ListID(),
		"message_template": bc.MessageTemplate(),
		"status":           bc.Status(),
		"scheduled_for":    bc.ScheduledFor(),
		"started_at":       bc.StartedAt(),
		"completed_at":     bc.CompletedAt(),
		"total_contacts":   bc.TotalContacts(),
		"sent_count":       bc.SentCount(),
		"failed_count":     bc.FailedCount(),
		"pending_count":    bc.PendingCount(),
		"rate_limit":       bc.RateLimit(),
		"created_at":       bc.CreatedAt(),
		"updated_at":       bc.UpdatedAt(),
	}
}

// Request/Response types
type CreateBroadcastRequest struct {
	Name            string                 `json:"name" binding:"required"`
	ListID          string                 `json:"list_id" binding:"required"`
	MessageTemplate MessageTemplateRequest `json:"message_template" binding:"required"`
	RateLimit       int                    `json:"rate_limit"`
}

type UpdateBroadcastRequest struct {
	Name            *string                 `json:"name"`
	MessageTemplate *MessageTemplateRequest `json:"message_template"`
	RateLimit       *int                    `json:"rate_limit"`
}

type MessageTemplateRequest struct {
	Type       string            `json:"type" binding:"required"`
	Content    string            `json:"content"`
	TemplateID *string           `json:"template_id"`
	Variables  map[string]string `json:"variables"`
	MediaURL   *string           `json:"media_url"`
}

type ScheduleBroadcastRequest struct {
	ScheduledFor time.Time `json:"scheduled_for" binding:"required"`
}
