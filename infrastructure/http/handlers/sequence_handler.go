package handlers

import (
	"fmt"
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/caloi/ventros-crm/internal/domain/automation/sequence"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SequenceHandler struct {
	logger *zap.Logger
	db     *gorm.DB
}

func NewSequenceHandler(logger *zap.Logger, db *gorm.DB) *SequenceHandler {
	return &SequenceHandler{
		logger: logger,
		db:     db,
	}
}

// ListSequences lists all sequences for the authenticated tenant
// @Summary List sequences
// @Description Get a paginated list of sequences with optional filters
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (draft, active, paused, archived)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} map[string]interface{} "sequences: array of sequences, total: total count, page: current page, limit: items per page"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences [get]
func (h *SequenceHandler) ListSequences(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	// Get filter parameters
	statusFilter := c.Query("status")

	repo := persistence.NewGormSequenceRepository(h.db)

	// Get sequences
	var allSequences []*sequence.Sequence
	var err error

	if statusFilter != "" {
		// Filter by status
		status := sequence.SequenceStatus(statusFilter)
		allSequences, err = repo.FindByStatus(status)
	} else {
		// Get all for tenant
		allSequences, err = repo.FindByTenantID(tenantID)
	}

	if err != nil {
		h.logger.Error("Failed to list sequences", zap.Error(err))
		errors.InternalError(c, "Failed to list sequences", err)
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

	total := len(allSequences)
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	paginatedSequences := allSequences[offset:end]

	// Convert to response
	sequences := make([]map[string]interface{}, len(paginatedSequences))
	for i, seq := range paginatedSequences {
		sequences[i] = h.sequenceToResponse(seq)
	}

	c.JSON(http.StatusOK, gin.H{
		"sequences": sequences,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// CreateSequence creates a new sequence
// @Summary Create sequence
// @Description Create a new automated sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param sequence body CreateSequenceRequest true "Sequence details"
// @Success 201 {object} map[string]interface{} "sequence: created sequence object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences [post]
func (h *SequenceHandler) CreateSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	var req CreateSequenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		errors.BadRequest(c, "name is required")
		return
	}
	if req.TriggerType == "" {
		errors.BadRequest(c, "trigger_type is required")
		return
	}

	// Create sequence
	seq, err := sequence.NewSequence(
		tenantID,
		req.Name,
		req.Description,
		sequence.TriggerType(req.TriggerType),
	)
	if err != nil {
		errors.BadRequest(c, "Failed to create sequence: "+err.Error())
		return
	}

	// Add steps if provided
	for _, stepReq := range req.Steps {
		template := sequence.MessageTemplate{
			Type:       stepReq.MessageTemplate.Type,
			Content:    stepReq.MessageTemplate.Content,
			TemplateID: stepReq.MessageTemplate.TemplateID,
			Variables:  stepReq.MessageTemplate.Variables,
			MediaURL:   stepReq.MessageTemplate.MediaURL,
		}

		step := sequence.NewSequenceStep(
			stepReq.Order,
			stepReq.Name,
			stepReq.DelayAmount,
			sequence.DelayUnit(stepReq.DelayUnit),
			template,
		)

		if err := seq.AddStep(step); err != nil {
			errors.BadRequest(c, "Failed to add step: "+err.Error())
			return
		}
	}

	// Save to database
	repo := persistence.NewGormSequenceRepository(h.db)
	if err := repo.Save(seq); err != nil {
		h.logger.Error("Failed to save sequence", zap.Error(err))
		errors.InternalError(c, "Failed to create sequence", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"sequence": h.sequenceToResponse(seq),
	})
}

// GetSequence gets a sequence by ID
// @Summary Get sequence
// @Description Get a sequence by ID
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "sequence: sequence object"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id} [get]
func (h *SequenceHandler) GetSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		h.logger.Error("Failed to find sequence", zap.Error(err))
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence": h.sequenceToResponse(seq),
	})
}

// UpdateSequence updates a sequence
// @Summary Update sequence
// @Description Update a sequence (only allowed in draft status)
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Param sequence body UpdateSequenceRequest true "Updated sequence details"
// @Success 200 {object} map[string]interface{} "sequence: updated sequence object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id} [put]
func (h *SequenceHandler) UpdateSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	var req UpdateSequenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Update name
	if req.Name != nil {
		if err := seq.UpdateName(*req.Name); err != nil {
			errors.BadRequest(c, err.Error())
			return
		}
	}

	// Update description
	if req.Description != nil {
		seq.UpdateDescription(*req.Description)
	}

	// Update exit_on_reply
	if req.ExitOnReply != nil {
		seq.UpdateExitOnReply(*req.ExitOnReply)
	}

	// Save
	if err := repo.Save(seq); err != nil {
		h.logger.Error("Failed to update sequence", zap.Error(err))
		errors.InternalError(c, "Failed to update sequence", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence": h.sequenceToResponse(seq),
	})
}

// ActivateSequence activates a sequence
// @Summary Activate sequence
// @Description Activate a sequence to start accepting enrollments
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "sequence: activated sequence object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/activate [post]
func (h *SequenceHandler) ActivateSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Activate
	if err := seq.Activate(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(seq); err != nil {
		h.logger.Error("Failed to activate sequence", zap.Error(err))
		errors.InternalError(c, "Failed to activate sequence", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence": h.sequenceToResponse(seq),
		"message":  "Sequence activated successfully",
	})
}

// PauseSequence pauses a sequence
// @Summary Pause sequence
// @Description Pause an active sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "sequence: paused sequence object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/pause [post]
func (h *SequenceHandler) PauseSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Pause
	if err := seq.Pause(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(seq); err != nil {
		h.logger.Error("Failed to pause sequence", zap.Error(err))
		errors.InternalError(c, "Failed to pause sequence", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence": h.sequenceToResponse(seq),
		"message":  "Sequence paused successfully",
	})
}

// ResumeSequence resumes a paused sequence
// @Summary Resume sequence
// @Description Resume a paused sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "sequence: resumed sequence object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/resume [post]
func (h *SequenceHandler) ResumeSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Resume
	if err := seq.Resume(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(seq); err != nil {
		h.logger.Error("Failed to resume sequence", zap.Error(err))
		errors.InternalError(c, "Failed to resume sequence", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence": h.sequenceToResponse(seq),
		"message":  "Sequence resumed successfully",
	})
}

// ArchiveSequence archives a sequence
// @Summary Archive sequence
// @Description Archive a sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "sequence: archived sequence object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/archive [post]
func (h *SequenceHandler) ArchiveSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Archive
	if err := seq.Archive(); err != nil {
		errors.BadRequest(c, err.Error())
		return
	}

	// Save
	if err := repo.Save(seq); err != nil {
		h.logger.Error("Failed to archive sequence", zap.Error(err))
		errors.InternalError(c, "Failed to archive sequence", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence": h.sequenceToResponse(seq),
		"message":  "Sequence archived successfully",
	})
}

// DeleteSequence deletes a sequence
// @Summary Delete sequence
// @Description Delete a sequence (only in draft status)
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "message: success message"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id} [delete]
func (h *SequenceHandler) DeleteSequence(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Only allow deletion of draft sequences
	if seq.Status() != sequence.SequenceStatusDraft {
		errors.BadRequest(c, "Can only delete sequences in draft status")
		return
	}

	// Delete
	if err := repo.Delete(sequenceID); err != nil {
		h.logger.Error("Failed to delete sequence", zap.Error(err))
		errors.InternalError(c, "Failed to delete sequence", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sequence deleted successfully",
	})
}

// GetSequenceStats gets sequence statistics
// @Summary Get sequence stats
// @Description Get statistics for a sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "stats: sequence statistics object"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/stats [get]
func (h *SequenceHandler) GetSequenceStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	repo := persistence.NewGormSequenceRepository(h.db)
	seq, err := repo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	stats := seq.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"stats": map[string]interface{}{
			"total_enrolled":  stats.TotalEnrolled,
			"active_count":    stats.ActiveCount,
			"completed_count": stats.CompletedCount,
			"exited_count":    stats.ExitedCount,
			"completion_rate": stats.CompletionRate,
		},
	})
}

// EnrollContact enrolls a contact in a sequence
// @Summary Enroll contact
// @Description Enroll a contact in a sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Param enrollment body EnrollContactRequest true "Enrollment details"
// @Success 201 {object} map[string]interface{} "enrollment: created enrollment object"
// @Failure 400 {object} map[string]interface{} "error: validation error message"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/enroll [post]
func (h *SequenceHandler) EnrollContact(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
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

	// Check if sequence exists
	seqRepo := persistence.NewGormSequenceRepository(h.db)
	seq, err := seqRepo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check tenant ownership
	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Check if sequence is active
	if seq.Status() != sequence.SequenceStatusActive {
		errors.BadRequest(c, "Sequence must be active to enroll contacts")
		return
	}

	// Check if already enrolled
	enrollmentRepo := persistence.NewGormSequenceEnrollmentRepository(h.db)
	existing, err := enrollmentRepo.FindActiveBySequenceAndContact(sequenceID, contactID)
	if err != nil {
		h.logger.Error("Failed to check existing enrollment", zap.Error(err))
		errors.InternalError(c, "Failed to check enrollment", err)
		return
	}
	if existing != nil {
		errors.BadRequest(c, "Contact is already enrolled in this sequence")
		return
	}

	// Get first step delay
	firstStep, err := seq.GetStepByOrder(0)
	if err != nil || firstStep == nil {
		errors.BadRequest(c, "Sequence has no steps")
		return
	}

	// Create enrollment
	enrollment, err := sequence.NewSequenceEnrollment(
		sequenceID,
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

	// Update sequence stats
	seq.IncrementEnrolled()
	if err := seqRepo.Save(seq); err != nil {
		h.logger.Error("Failed to update sequence stats", zap.Error(err))
	}

	c.JSON(http.StatusCreated, gin.H{
		"enrollment": h.enrollmentToResponse(enrollment),
		"message":    "Contact enrolled successfully",
	})
}

// ListEnrollments lists all enrollments for a sequence
// @Summary List enrollments
// @Description Get all enrollments for a sequence
// @Tags AUTOMATION - Sequences
// @Accept json
// @Produce json
// @Param id path string true "Sequence ID"
// @Success 200 {object} map[string]interface{} "enrollments: array of enrollments"
// @Failure 404 {object} map[string]interface{} "error: sequence not found"
// @Failure 500 {object} map[string]interface{} "error: internal server error message"
// @Router /api/v1/automation/sequences/{id}/enrollments [get]
func (h *SequenceHandler) ListEnrollments(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	sequenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errors.BadRequest(c, "invalid sequence id")
		return
	}

	// Check sequence exists and ownership
	seqRepo := persistence.NewGormSequenceRepository(h.db)
	seq, err := seqRepo.FindByID(sequenceID)
	if err != nil {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	if seq.TenantID() != tenantID {
		errors.NotFound(c, "sequence", sequenceID.String())
		return
	}

	// Get enrollments
	enrollmentRepo := persistence.NewGormSequenceEnrollmentRepository(h.db)
	enrollments, err := enrollmentRepo.FindBySequenceID(sequenceID)
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
func (h *SequenceHandler) sequenceToResponse(seq *sequence.Sequence) map[string]interface{} {
	// Convert steps to response format
	steps := seq.Steps()
	stepResponses := make([]map[string]interface{}, len(steps))
	for i, step := range steps {
		stepResponses[i] = map[string]interface{}{
			"id":               step.ID,
			"order":            step.Order,
			"name":             step.Name,
			"delay_amount":     step.DelayAmount,
			"delay_unit":       step.DelayUnit,
			"message_template": step.MessageTemplate,
			"conditions":       step.Conditions,
			"created_at":       step.CreatedAt,
		}
	}

	return map[string]interface{}{
		"id":              seq.ID(),
		"tenant_id":       seq.TenantID(),
		"name":            seq.Name(),
		"description":     seq.Description(),
		"status":          seq.Status(),
		"trigger_type":    seq.TriggerType(),
		"trigger_data":    seq.TriggerData(),
		"exit_on_reply":   seq.ExitOnReply(),
		"steps":           stepResponses,
		"total_enrolled":  seq.TotalEnrolled(),
		"active_count":    seq.ActiveCount(),
		"completed_count": seq.CompletedCount(),
		"exited_count":    seq.ExitedCount(),
		"created_at":      seq.CreatedAt(),
		"updated_at":      seq.UpdatedAt(),
	}
}

func (h *SequenceHandler) enrollmentToResponse(enrollment *sequence.SequenceEnrollment) map[string]interface{} {
	return map[string]interface{}{
		"id":                 enrollment.ID(),
		"sequence_id":        enrollment.SequenceID(),
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
type CreateSequenceRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	TriggerType string                    `json:"trigger_type" binding:"required"`
	Steps       []CreateSequenceStepRequest `json:"steps"`
}

type CreateSequenceStepRequest struct {
	Order           int                    `json:"order" binding:"required"`
	Name            string                 `json:"name" binding:"required"`
	DelayAmount     int                    `json:"delay_amount" binding:"required"`
	DelayUnit       string                 `json:"delay_unit" binding:"required"`
	MessageTemplate MessageTemplateRequest `json:"message_template" binding:"required"`
}

type UpdateSequenceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ExitOnReply *bool   `json:"exit_on_reply"`
}

type EnrollContactRequest struct {
	ContactID string `json:"contact_id" binding:"required"`
}
