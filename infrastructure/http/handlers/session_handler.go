package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionHandler struct {
	logger      *zap.Logger
	sessionRepo session.Repository
}

func NewSessionHandler(logger *zap.Logger, sessionRepo session.Repository) *SessionHandler {
	return &SessionHandler{
		logger:      logger,
		sessionRepo: sessionRepo,
	}
}

// ListSessions lists all sessions with optional filters
// @Summary List sessions
// @Description Lista todas as sessões com filtros opcionais
// @Tags sessions
// @Produce json
// @Param tenant_id query string false "Filter by tenant ID"
// @Param contact_id query string false "Filter by contact ID (UUID)"
// @Param status query string false "Filter by status (active, ended)"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} map[string]interface{} "List of sessions"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// For now, return a simple message since we need to implement pagination
	// TODO: Implement proper session listing with filters
	c.JSON(http.StatusOK, gin.H{
		"message": "Session listing not yet implemented",
		"note":    "Use GET /api/v1/sessions/{id} to get specific session",
	})
}

// GetSession gets a session by ID
// @Summary Get session by ID
// @Description Obtém detalhes de uma sessão específica
// @Tags sessions
// @Produce json
// @Param id path string true "Session ID (UUID)"
// @Success 200 {object} map[string]interface{} "Session details"
// @Failure 400 {object} map[string]interface{} "Invalid session ID"
// @Failure 404 {object} map[string]interface{} "Session not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/sessions/{id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid session ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID format"})
		return
	}

	sess, err := h.sessionRepo.FindByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to find session", zap.String("session_id", sessionID.String()), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session"})
		return
	}

	if sess == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Convert domain session to API response
	response := h.sessionToResponse(sess)
	c.JSON(http.StatusOK, response)
}

// GetSessionStats gets session statistics
// @Summary Get session statistics
// @Description Obtém estatísticas das sessões por tenant
// @Tags sessions
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "Session statistics"
// @Failure 400 {object} map[string]interface{} "Missing tenant ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/sessions/stats [get]
func (h *SessionHandler) GetSessionStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant_id is required"})
		return
	}

	activeCount, err := h.sessionRepo.CountActiveByTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to count active sessions", zap.String("tenant_id", tenantID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant_id":      tenantID,
		"active_sessions": activeCount,
		"timestamp":      c.Request.Context().Value("request_time"),
	})
}

// sessionToResponse converts domain session to API response
func (h *SessionHandler) sessionToResponse(sess *session.Session) map[string]interface{} {
	response := map[string]interface{}{
		"id":                sess.ID(),
		"contact_id":        sess.ContactID(),
		"tenant_id":         sess.TenantID(),
		"channel_type_id":   sess.ChannelTypeID(),
		"started_at":        sess.StartedAt(),
		"ended_at":          sess.EndedAt(),
		"status":            sess.Status(),
		"end_reason":        sess.EndReason(),
		"timeout_duration":  sess.TimeoutDuration().String(),
		"last_activity_at":  sess.LastActivityAt(),
		"message_count":     sess.MessageCount(),
		"messages_from_contact": sess.MessagesFromContact(),
		"messages_from_agent":   sess.MessagesFromAgent(),
		"duration_seconds":      sess.DurationSeconds(),
		"agent_ids":             sess.AgentIDs(),
		"agent_transfers":       sess.AgentTransfers(),
		"summary":               sess.Summary(),
		"sentiment":             sess.Sentiment(),
		"sentiment_score":       sess.SentimentScore(),
		"topics":                sess.Topics(),
		"next_steps":            sess.NextSteps(),
		"key_entities":          sess.KeyEntities(),
		"resolved":              sess.IsResolved(),
		"escalated":             sess.IsEscalated(),
		"converted":             sess.IsConverted(),
		"outcome_tags":          sess.OutcomeTags(),
	}

	return response
}
