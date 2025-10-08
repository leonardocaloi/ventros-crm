package handlers

import (
	"net/http"
	"strconv"

	"github.com/caloi/ventros-crm/infrastructure/persistence"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DomainEventHandler struct {
	eventLogRepo *persistence.DomainEventLogRepository
	logger       *zap.Logger
}

func NewDomainEventHandler(eventLogRepo *persistence.DomainEventLogRepository, logger *zap.Logger) *DomainEventHandler {
	return &DomainEventHandler{
		eventLogRepo: eventLogRepo,
		logger:       logger,
	}
}

// ListDomainEventsByContact lista eventos de domínio de um contato
// @Summary List domain events by contact
// @Description Lista todos os eventos de domínio disparados para um contato
// @Tags domain-events
// @Produce json
// @Param contact_id path string true "Contact ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/contacts/{contact_id}/domain-events [get]
func (h *DomainEventHandler) ListDomainEventsByContact(c *gin.Context) {
	contactIDStr := c.Param("contact_id")
	contactID, err := uuid.Parse(contactIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid contact ID"})
		return
	}

	events, err := h.eventLogRepo.FindByAggregateID(c.Request.Context(), contactID)
	if err != nil {
		h.logger.Error("Failed to fetch domain events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contact_id": contactID,
		"events":     events,
		"count":      len(events),
	})
}

// ListDomainEventsBySession lista eventos de domínio de uma sessão
// @Summary List domain events by session
// @Description Lista todos os eventos de domínio disparados para uma sessão
// @Tags domain-events
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sessions/{session_id}/domain-events [get]
func (h *DomainEventHandler) ListDomainEventsBySession(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	events, err := h.eventLogRepo.FindByAggregateID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to fetch domain events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"events":     events,
		"count":      len(events),
	})
}

// ListDomainEventsByProject lista eventos de domínio de um projeto
// @Summary List domain events by project
// @Description Lista todos os eventos de domínio disparados em um projeto
// @Tags domain-events
// @Produce json
// @Param project_id query string true "Project ID"
// @Param limit query int false "Limit (default: 100)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/domain-events [get]
func (h *DomainEventHandler) ListDomainEventsByProject(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	events, err := h.eventLogRepo.FindByProjectID(c.Request.Context(), projectID, limit)
	if err != nil {
		h.logger.Error("Failed to fetch domain events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project_id": projectID,
		"events":     events,
		"count":      len(events),
		"limit":      limit,
	})
}

// ListDomainEventsByType lista eventos de domínio por tipo
// @Summary List domain events by type
// @Description Lista eventos de domínio filtrados por tipo
// @Tags domain-events
// @Produce json
// @Param event_type query string true "Event Type (e.g., contact.created, session.started)"
// @Param limit query int false "Limit (default: 100)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/domain-events/by-type [get]
func (h *DomainEventHandler) ListDomainEventsByType(c *gin.Context) {
	eventType := c.Query("event_type")
	if eventType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_type is required"})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		// Parse limit
	}

	events, err := h.eventLogRepo.FindByEventType(c.Request.Context(), eventType, limit)
	if err != nil {
		h.logger.Error("Failed to fetch domain events", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_type": eventType,
		"events":     events,
		"count":      len(events),
		"limit":      limit,
	})
}
