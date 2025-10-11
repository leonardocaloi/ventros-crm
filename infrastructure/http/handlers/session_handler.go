package handlers

import (
	"net/http"
	"strconv"
	"strings"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/internal/application/queries"
	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionHandler struct {
	logger                     *zap.Logger
	sessionRepo                session.Repository
	listSessionsQueryHandler   *queries.ListSessionsQueryHandler
	searchSessionsQueryHandler *queries.SearchSessionsQueryHandler
}

func NewSessionHandler(logger *zap.Logger, sessionRepo session.Repository) *SessionHandler {
	return &SessionHandler{
		logger:                     logger,
		sessionRepo:                sessionRepo,
		listSessionsQueryHandler:   queries.NewListSessionsQueryHandler(sessionRepo, logger),
		searchSessionsQueryHandler: queries.NewSearchSessionsQueryHandler(sessionRepo, logger),
	}
}

// ListSessions lists all sessions with optional filters
//
//	@Summary		List sessions
//	@Description	Lista todas as sessões. Quando usado no endpoint global /sessions, requer contact_id ou channel_id
//	@Tags			sessions
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			contact_id	query		string					false	"Filter by contact ID (UUID) - required for global endpoint"
//	@Param			channel_id	query		string					false	"Filter by channel ID (UUID) - required for global endpoint"
//	@Param			status		query		string					false	"Filter by status (active, ended)"
//	@Param			limit		query		int						false	"Limit results"			default(50)
//	@Param			offset		query		int						false	"Offset for pagination"	default(0)
//	@Success		200			{object}	map[string]interface{}	"List of sessions"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/sessions [get]
//	@Router			/api/v1/contacts/{contact_id}/sessions [get]
//	@Router			/api/v1/channels/{channel_id}/sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// Extract filters from query params or path params
	contactIDStr := c.Query("contact_id")
	channelIDStr := c.Query("channel_id")

	// Check if this is a nested route (contact or channel)
	parentID := c.Param("id") // Will be contact_id or channel_id depending on route
	isNestedRoute := parentID != ""

	// For nested routes, determine which parent it is
	if isNestedRoute {
		// Check the full path to determine context
		path := c.FullPath()
		if strings.Contains(path, "/contacts/") {
			contactIDStr = parentID
		} else if strings.Contains(path, "/channels/") {
			channelIDStr = parentID
		}
	}

	// Validate: global endpoint requires at least one filter
	if !isNestedRoute && contactIDStr == "" && channelIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Filter required",
			"hint":  "Please provide contact_id or channel_id query parameter, or use nested routes: /contacts/{id}/sessions or /channels/{id}/sessions",
		})
		return
	}

	// TODO: Implement actual filtering with repository
	// For now, return placeholder
	c.JSON(http.StatusOK, gin.H{
		"message":    "Sessions retrieved successfully",
		"contact_id": contactIDStr,
		"channel_id": channelIDStr,
		"sessions":   []interface{}{},
		"count":      0,
		"note":       "Implementation pending - repository integration needed",
	})
}

// GetSession gets a session by ID
//
//	@Summary		Get session by ID
//	@Description	Obtém detalhes de uma sessão específica
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string					true	"Session ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Session details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid session ID"
//	@Failure		404	{object}	map[string]interface{}	"Session not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/sessions/{id} [get]
//	@Router			/api/v1/contacts/{id}/sessions/{session_id} [get]
//	@Router			/api/v1/channels/{id}/sessions/{session_id} [get]
func (h *SessionHandler) GetSession(c *gin.Context) {
	// Try session_id first (nested route), then id (global route)
	idStr := c.Param("session_id")
	if idStr == "" {
		idStr = c.Param("id")
	}
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
//
//	@Summary		Get session statistics
//	@Description	Obtém estatísticas das sessões por tenant
//	@Tags			sessions
//	@Produce		json
//	@Param			tenant_id	query		string					true	"Tenant ID"
//	@Success		200			{object}	map[string]interface{}	"Session statistics"
//	@Failure		400			{object}	map[string]interface{}	"Missing tenant ID"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/sessions/stats [get]
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
		"tenant_id":       tenantID,
		"active_sessions": activeCount,
		"timestamp":       c.Request.Context().Value("request_time"),
	})
}

// sessionToResponse converts domain session to API response
func (h *SessionHandler) sessionToResponse(sess *session.Session) map[string]interface{} {
	response := map[string]interface{}{
		"id":                    sess.ID(),
		"contact_id":            sess.ContactID(),
		"tenant_id":             sess.TenantID(),
		"channel_type_id":       sess.ChannelTypeID(),
		"started_at":            sess.StartedAt(),
		"ended_at":              sess.EndedAt(),
		"status":                sess.Status(),
		"end_reason":            sess.EndReason(),
		"timeout_duration":      sess.TimeoutDuration().String(),
		"last_activity_at":      sess.LastActivityAt(),
		"message_count":         sess.MessageCount(),
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

// CloseSessionRequest representa a requisição para encerrar uma sessão
type CloseSessionRequest struct {
	Reason string `json:"reason" binding:"required"` // "resolved", "transferred", "escalated", "agent_closed"
	Notes  string `json:"notes"`
}

// CloseSession encerra uma sessão manualmente (por agente)
//
//	@Summary		Close session
//	@Description	Encerra uma sessão manualmente. Apenas agentes podem encerrar sessões.
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		string					true	"Session ID (UUID)"
//	@Param			request	body		CloseSessionRequest		true	"Close session request"
//	@Success		200		{object}	map[string]interface{}	"Session closed successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		404		{object}	map[string]interface{}	"Session not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/sessions/{id}/close [post]
func (h *SessionHandler) CloseSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req CloseSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Busca sessão
	sess, err := h.sessionRepo.FindByID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get session", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Valida se sessão já está encerrada
	if sess.Status() == session.StatusEnded {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session is already ended"})
		return
	}

	// Encerra sessão baseado no reason
	switch req.Reason {
	case "resolved":
		if err := sess.Resolve(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case "escalated":
		if err := sess.Escalate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	case "transferred", "agent_closed":
		// Encerra sessão com reason customizado
		if err := sess.End(session.EndReason(req.Reason)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reason. Must be: resolved, transferred, escalated, or agent_closed"})
		return
	}

	// Salva sessão
	if err := h.sessionRepo.Save(c.Request.Context(), sess); err != nil {
		h.logger.Error("Failed to update session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close session"})
		return
	}

	h.logger.Info("Session closed by agent",
		zap.String("session_id", sessionID.String()),
		zap.String("reason", req.Reason))

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session closed successfully",
		"session_id": sessionID,
		"reason":     req.Reason,
		"status":     sess.Status(),
	})
}

// ListSessionsAdvanced lists sessions with advanced filters, pagination, and sorting
//
//	@Summary		List sessions with advanced filters and pagination
//	@Description	Retrieve a paginated list of sessions with advanced filtering capabilities including contact, pipeline, status, sentiment, and resolution flags. Supports sorting and pagination for efficient data retrieval. Perfect for building session dashboards and reports.
//	@Description
//	@Description	**Filtering Options:**
//	@Description	- Filter by contact to see all sessions for a specific customer
//	@Description	- Filter by pipeline to analyze sessions in a specific workflow
//	@Description	- Filter by status (active/ended) to focus on ongoing or completed sessions
//	@Description	- Filter by sentiment (positive/negative/neutral) for customer satisfaction analysis
//	@Description	- Filter by resolved/escalated/converted flags for outcome tracking
//	@Description	- Filter by date range to analyze sessions within specific time periods
//	@Description	- Filter by message count range to find short or long conversations
//	@Description
//	@Description	**Sorting Options:**
//	@Description	- Sort by started_at, ended_at, message_count, or duration
//	@Description	- Ascending or descending order
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized with composite GORM indexes on tenant+status, tenant+contact, tenant+pipeline
//	@Description	- GIN indexes on JSONB fields (agent_ids, topics, outcome_tags) for fast array searches
//	@Description	- Pagination prevents large result sets
//	@Description	- Maximum 100 results per page
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			contact_id			query		string	false	"Filter by contact UUID - Example: 550e8400-e29b-41d4-a716-446655440000"
//	@Param			pipeline_id			query		string	false	"Filter by pipeline UUID - Example: 660e8400-e29b-41d4-a716-446655440001"
//	@Param			status				query		string	false	"Filter by session status" Enums(active, ended) example(active)
//	@Param			sentiment			query		string	false	"Filter by detected sentiment" Enums(positive, negative, neutral) example(positive)
//	@Param			resolved			query		bool	false	"Filter by resolved flag - true: only resolved sessions, false: only unresolved" example(true)
//	@Param			escalated			query		bool	false	"Filter by escalated flag - true: only escalated sessions" example(false)
//	@Param			converted			query		bool	false	"Filter by converted flag - true: sessions that led to conversions" example(true)
//	@Param			started_after		query		string	false	"Filter sessions started after this timestamp - Format: 2006-01-02T15:04:05Z" example(2024-01-01T00:00:00Z)
//	@Param			started_before		query		string	false	"Filter sessions started before this timestamp" example(2024-12-31T23:59:59Z)
//	@Param			min_messages		query		int		false	"Minimum number of messages in session - Example: 5" example(5)
//	@Param			max_messages		query		int		false	"Maximum number of messages in session - Example: 100" example(100)
//	@Param			page				query		int		false	"Page number for pagination (starts at 1)" default(1) minimum(1) example(1)
//	@Param			limit				query		int		false	"Number of results per page (max 100)" default(20) minimum(1) maximum(100) example(20)
//	@Param			sort_by				query		string	false	"Field to sort by" Enums(started_at, ended_at, message_count, duration_seconds, created_at) default(started_at) example(started_at)
//	@Param			sort_dir			query		string	false	"Sort direction" Enums(asc, desc) default(desc) example(desc)
//	@Success		200					{object}	queries.ListSessionsResponse	"Successfully retrieved sessions with pagination metadata"
//	@Failure		400					{object}	map[string]interface{}			"Bad Request - Invalid parameters (e.g., invalid UUID format, invalid page number, limit exceeds maximum)"
//	@Failure		401					{object}	map[string]interface{}			"Unauthorized - Missing or invalid authentication token"
//	@Failure		403					{object}	map[string]interface{}			"Forbidden - User doesn't have permission to access this tenant's sessions"
//	@Failure		500					{object}	map[string]interface{}			"Internal Server Error - Database connection issues or unexpected errors"
//	@Router			/api/v1/crm/sessions/advanced [get]
func (h *SessionHandler) ListSessionsAdvanced(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse sorting
	sortBy := c.DefaultQuery("sort_by", "started_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")

	// Parse optional UUID filters
	var contactID, pipelineID *uuid.UUID
	if contactIDStr := c.Query("contact_id"); contactIDStr != "" {
		if id, err := uuid.Parse(contactIDStr); err == nil {
			contactID = &id
		}
	}
	if pipelineIDStr := c.Query("pipeline_id"); pipelineIDStr != "" {
		if id, err := uuid.Parse(pipelineIDStr); err == nil {
			pipelineID = &id
		}
	}

	// Parse boolean filters
	var resolved, escalated, converted *bool
	if resolvedStr := c.Query("resolved"); resolvedStr != "" {
		if b, err := strconv.ParseBool(resolvedStr); err == nil {
			resolved = &b
		}
	}
	if escalatedStr := c.Query("escalated"); escalatedStr != "" {
		if b, err := strconv.ParseBool(escalatedStr); err == nil {
			escalated = &b
		}
	}
	if convertedStr := c.Query("converted"); convertedStr != "" {
		if b, err := strconv.ParseBool(convertedStr); err == nil {
			converted = &b
		}
	}

	// Parse string filters
	var status, sentiment *string
	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}
	if sentimentStr := c.Query("sentiment"); sentimentStr != "" {
		sentiment = &sentimentStr
	}

	// Create tenant ID
	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		apierrors.ValidationError(c, "tenant_id", "Invalid tenant ID")
		return
	}

	// Execute query
	query := queries.ListSessionsQuery{
		TenantID:   tenantID,
		ContactID:  contactID,
		PipelineID: pipelineID,
		Status:     status,
		Resolved:   resolved,
		Escalated:  escalated,
		Converted:  converted,
		Sentiment:  sentiment,
		Page:       page,
		Limit:      limit,
		SortBy:     sortBy,
		SortDir:    sortDir,
	}

	response, err := h.listSessionsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list sessions", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve sessions", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchSessions performs full-text search on sessions
//
//	@Summary		Full-text search across sessions
//	@Description	Perform intelligent full-text search across session summaries, topics, key entities, next steps, and outcome tags. Uses PostgreSQL ILIKE for case-insensitive pattern matching with relevance scoring.
//	@Description
//	@Description	**Search Capabilities:**
//	@Description	- Searches across session summary text (AI-generated conversation summaries)
//	@Description	- Searches through detected topics (array of conversation topics)
//	@Description	- Searches through outcome tags (categorization tags added at session end)
//	@Description	- Searches through key entities (people, products, companies mentioned)
//	@Description	- Searches through next steps (action items identified in conversation)
//	@Description
//	@Description	**Match Scoring:**
//	@Description	- Summary matches: 2.0 score (highest priority - main content)
//	@Description	- Topics matches: 1.5 score (high priority - categorization)
//	@Description	- Outcome tags matches: 1.3 score (medium priority - resolution info)
//	@Description	- Key entities/next steps matches: 1.0 score (standard priority)
//	@Description
//	@Description	**Search Examples:**
//	@Description	- Search for "refund" to find all sessions where refunds were discussed
//	@Description	- Search for "escalated" to find problematic sessions
//	@Description	- Search for "product-demo" to find sessions with demo requests
//	@Description	- Search for customer/company names mentioned in conversations
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized with GIN indexes on JSONB fields (topics, outcome_tags, key_entities)
//	@Description	- Results ordered by match score (highest relevance first)
//	@Description	- Maximum 100 results to ensure fast response times
//	@Tags			sessions
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string	true	"Search query - minimum 1 character, case-insensitive - Example: 'refund request' or 'product-demo' or 'escalated'" minlength(1) example(refund request)
//	@Param			limit	query		int		false	"Maximum number of results to return (max 100)" default(20) minimum(1) maximum(100) example(20)
//	@Success		200		{object}	queries.SearchSessionsResponse	"Successfully found matching sessions with relevance scores and matched fields"
//	@Failure		400		{object}	map[string]interface{}			"Bad Request - Missing or invalid search query, limit exceeds maximum"
//	@Failure		401		{object}	map[string]interface{}			"Unauthorized - Missing or invalid authentication token"
//	@Failure		403		{object}	map[string]interface{}			"Forbidden - User doesn't have permission to search this tenant's sessions"
//	@Failure		500		{object}	map[string]interface{}			"Internal Server Error - Database connection issues or search execution errors"
//	@Router			/api/v1/crm/sessions/search [get]
func (h *SessionHandler) SearchSessions(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
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

	// Create tenant ID
	tenantID, err := shared.NewTenantID(authCtx.TenantID)
	if err != nil {
		apierrors.ValidationError(c, "tenant_id", "Invalid tenant ID")
		return
	}

	// Execute search
	query := queries.SearchSessionsQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchSessionsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search sessions", zap.Error(err))
		apierrors.InternalError(c, "Failed to search sessions", err)
		return
	}

	c.JSON(http.StatusOK, response)
}
