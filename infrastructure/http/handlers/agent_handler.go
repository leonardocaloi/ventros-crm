package handlers

import (
	"net/http"
	"strconv"
	"time"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	"github.com/caloi/ventros-crm/internal/application/queries"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/agent"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AgentHandler struct {
	logger                   *zap.Logger
	agentRepo                agent.Repository
	listAgentsQueryHandler   *queries.ListAgentsQueryHandler
	searchAgentsQueryHandler *queries.SearchAgentsQueryHandler
}

func NewAgentHandler(logger *zap.Logger, agentRepo agent.Repository) *AgentHandler {
	return &AgentHandler{
		logger:                   logger,
		agentRepo:                agentRepo,
		listAgentsQueryHandler:   queries.NewListAgentsQueryHandler(agentRepo, logger),
		searchAgentsQueryHandler: queries.NewSearchAgentsQueryHandler(agentRepo, logger),
	}
}

// CreateAgentRequest representa o payload para criar um agente
type CreateAgentRequest struct {
	Name        string   `json:"name" binding:"required" example:"João Silva"`
	Email       string   `json:"email" binding:"required" example:"joao@empresa.com"`
	Phone       string   `json:"phone" example:"+5511999999999"`
	Role        string   `json:"role" example:"agent"`
	Department  string   `json:"department" example:"vendas"`
	Skills      []string `json:"skills" example:"vendas,suporte"`
	Languages   []string `json:"languages" example:"pt,en"`
	MaxSessions int      `json:"max_sessions" example:"5"`
	TenantID    string   `json:"tenant_id" binding:"required" example:"tenant_123"`
}

// CreateVirtualAgentRequest representa o payload para criar um agente virtual
type CreateVirtualAgentRequest struct {
	ProjectID            string  `json:"project_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepresentsPersonName string  `json:"represents_person_name" binding:"required" example:"Maria da Silva"`
	PeriodStart          string  `json:"period_start" binding:"required" example:"2023-01-01T00:00:00Z"`
	Reason               string  `json:"reason" binding:"required" example:"device_attribution"`
	SourceDevice         *string `json:"source_device,omitempty" example:"whatsapp:5511999999999"`
	Notes                string  `json:"notes" example:"Historical agent from old system"`
}

// EndVirtualAgentPeriodRequest representa o payload para finalizar um período de agente virtual
type EndVirtualAgentPeriodRequest struct {
	PeriodEnd string `json:"period_end" binding:"required" example:"2024-01-01T00:00:00Z"`
}

// UpdateAgentRequest representa o payload para atualizar um agente
type UpdateAgentRequest struct {
	Name        *string  `json:"name,omitempty"`
	Email       *string  `json:"email,omitempty"`
	Phone       *string  `json:"phone,omitempty"`
	Role        *string  `json:"role,omitempty"`
	Department  *string  `json:"department,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	Languages   []string `json:"languages,omitempty"`
	MaxSessions *int     `json:"max_sessions,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

// ListAgents lists all agents with optional filters
//
//	@Summary		List agents
//	@Description	Lista todos os agentes com filtros opcionais
//	@Tags			CRM - Agents
//	@Produce		json
//	@Param			tenant_id	query		string					false	"Filter by tenant ID"
//	@Param			active		query		bool					false	"Filter by active status"
//	@Param			role		query		string					false	"Filter by role"
//	@Param			department	query		string					false	"Filter by department"
//	@Param			limit		query		int						false	"Limit results"			default(50)
//	@Param			offset		query		int						false	"Offset for pagination"	default(0)
//	@Success		200			{array}		map[string]interface{}	"List of agents"
//	@Failure		400			{object}	map[string]interface{}	"Invalid parameters"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/agents [get]
func (h *AgentHandler) ListAgents(c *gin.Context) {
	// TODO: Implement proper agent listing with filters
	c.JSON(http.StatusOK, gin.H{
		"message": "Agent listing not yet implemented",
		"note":    "Use GET /api/v1/agents/{id} to get specific agent",
	})
}

// CreateAgent creates a new agent
//
//	@Summary		Create agent
//	@Description	Cria um novo agente
//	@Tags			CRM - Agents
//	@Accept			json
//	@Produce		json
//	@Param			agent	body		CreateAgentRequest		true	"Agent data"
//	@Success		201		{object}	map[string]interface{}	"Agent created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/agents [post]
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse agent request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement agent creation
	c.JSON(http.StatusCreated, gin.H{
		"message":      "Agent creation not yet implemented",
		"name":         req.Name,
		"email":        req.Email,
		"role":         req.Role,
		"department":   req.Department,
		"max_sessions": req.MaxSessions,
		"tenant_id":    req.TenantID,
	})
}

// GetAgent gets an agent by ID
//
//	@Summary		Get agent by ID
//	@Description	Obtém detalhes de um agente específico
//	@Tags			CRM - Agents
//	@Produce		json
//	@Param			id	path		string					true	"Agent ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Agent details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid agent ID"
//	@Failure		404	{object}	map[string]interface{}	"Agent not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/agents/{id} [get]
func (h *AgentHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("Invalid agent ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// TODO: Implement agent retrieval
	c.JSON(http.StatusOK, gin.H{
		"message":  "Agent retrieval not yet implemented",
		"agent_id": agentID,
	})
}

// UpdateAgent updates an agent
//
//	@Summary		Update agent
//	@Description	Atualiza um agente existente
//	@Tags			CRM - Agents
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Agent ID (UUID)"
//	@Param			agent	body		UpdateAgentRequest		true	"Agent update data"
//	@Success		200		{object}	map[string]interface{}	"Agent updated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		404		{object}	map[string]interface{}	"Agent not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/agents/{id} [put]
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	idStr := c.Param("id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// TODO: Implement agent update
	c.JSON(http.StatusOK, gin.H{
		"message":  "Agent update not yet implemented",
		"agent_id": agentID,
	})
}

// DeleteAgent deletes an agent
//
//	@Summary		Delete agent
//	@Description	Remove um agente (soft delete)
//	@Tags			CRM - Agents
//	@Produce		json
//	@Param			id	path	string	true	"Agent ID (UUID)"
//	@Success		204	"Agent deleted successfully"
//	@Failure		400	{object}	map[string]interface{}	"Invalid agent ID"
//	@Failure		404	{object}	map[string]interface{}	"Agent not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/agents/{id} [delete]
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// TODO: Implement agent deletion
	c.JSON(http.StatusOK, gin.H{
		"message":  "Agent deletion not yet implemented",
		"agent_id": agentID,
	})
}

// GetAgentStats gets agent statistics
//
//	@Summary		Get agent statistics
//	@Description	Obtém estatísticas de um agente
//	@Tags			CRM - Agents
//	@Produce		json
//	@Param			id	path		string					true	"Agent ID (UUID)"
//	@Success		200	{object}	map[string]interface{}	"Agent statistics"
//	@Failure		400	{object}	map[string]interface{}	"Invalid agent ID"
//	@Failure		404	{object}	map[string]interface{}	"Agent not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/agents/{id}/stats [get]
func (h *AgentHandler) GetAgentStats(c *gin.Context) {
	idStr := c.Param("id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID format"})
		return
	}

	// TODO: Implement agent statistics
	c.JSON(http.StatusOK, gin.H{
		"message":         "Agent statistics not yet implemented",
		"agent_id":        agentID,
		"active_sessions": 0,
		"total_sessions":  0,
		"avg_rating":      0.0,
	})
}

// ListAgentsAdvanced lists agents with advanced filters, pagination, and sorting
//
//	@Summary		List agents with advanced filters and pagination
//	@Description	Retrieve all agents (AI agents and human support staff) with comprehensive filtering capabilities. Agents handle customer conversations either autonomously (AI) or manually (human). Essential for team management, capacity planning, and performance monitoring.
//	@Description
//	@Description	**Filtering Capabilities:**
//	@Description	- Filter by project_id to view agents assigned to specific business units
//	@Description	- Filter by type to distinguish AI agents from human agents
//	@Description	- Filter by status (online, offline, busy) for real-time availability tracking
//	@Description	- Filter by active status to show/hide deactivated agents
//	@Description
//	@Description	**Common Use Cases:**
//	@Description	- Load all active agents for the team dashboard
//	@Description	- Build agent selector dropdowns for manual conversation assignment
//	@Description	- Monitor real-time agent availability and capacity
//	@Description	- Track agent performance and productivity metrics
//	@Description	- Identify offline or busy agents for workload balancing
//	@Description	- Generate agent reports by project or department
//	@Description	- Audit agent configurations and permissions
//	@Description
//	@Description	**Sorting Options:**
//	@Description	- Sort by name (alphabetical order)
//	@Description	- Sort by created_at (onboarding order)
//	@Description	- Ascending or descending order
//	@Description
//	@Description	**Performance:**
//	@Description	- Optimized GORM indexes on tenant+type for fast agent type queries
//	@Description	- Composite indexes on tenant+status for real-time availability checks
//	@Description	- Small result sets (typically < 200 agents per tenant) for instant responses
//	@Tags			CRM - Agents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			project_id	query		string						false	"Filter by project UUID - Example: 550e8400-e29b-41d4-a716-446655440000"
//	@Param			type		query		string						false	"Filter by agent type"												Enums(ai, human)				example(human)
//	@Param			status		query		string						false	"Filter by availability status"										Enums(online, offline, busy)	example(online)
//	@Param			active		query		bool						false	"Filter by active status - true: only active, false: only inactive"	example(true)
//	@Param			page		query		int							false	"Page number for pagination (starts at 1)"							default(1)				minimum(1)			example(1)
//	@Param			limit		query		int							false	"Results per page (max 100)"										default(20)				minimum(1)			maximum(100)	example(20)
//	@Param			sort_by		query		string						false	"Field to sort by"													Enums(name, created_at)	default(created_at)	example(name)
//	@Param			sort_dir	query		string						false	"Sort direction"													Enums(asc, desc)		default(desc)		example(asc)
//	@Success		200			{object}	queries.ListAgentsResponse	"Successfully retrieved agents with full details"
//	@Failure		400			{object}	map[string]interface{}		"Bad Request - Invalid UUID or parameter format"
//	@Failure		401			{object}	map[string]interface{}		"Unauthorized - Authentication required"
//	@Failure		403			{object}	map[string]interface{}		"Forbidden - No access to this tenant's agents"
//	@Failure		500			{object}	map[string]interface{}		"Internal Server Error"
//	@Router			/api/v1/crm/agents/advanced [get]
func (h *AgentHandler) ListAgentsAdvanced(c *gin.Context) {
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

	// Parse type filter
	var agentType *agent.AgentType
	if typeStr := c.Query("type"); typeStr != "" {
		t := agent.AgentType(typeStr)
		agentType = &t
	}

	// Parse status filter
	var status *agent.AgentStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := agent.AgentStatus(statusStr)
		status = &s
	}

	// Parse active filter
	var active *bool
	if activeStr := c.Query("active"); activeStr != "" {
		if a, err := strconv.ParseBool(activeStr); err == nil {
			active = &a
		}
	}

	query := queries.ListAgentsQuery{
		TenantID:  tenantID,
		ProjectID: projectID,
		Type:      agentType,
		Status:    status,
		Active:    active,
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortDir:   sortDir,
	}

	response, err := h.listAgentsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list agents", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve agents", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// SearchAgents performs full-text search on agents
//
//	@Summary		Search agents by name and email
//	@Description	Full-text search across agent names and email addresses. Perfect for quickly finding specific team members in organizations with many agents.
//	@Description
//	@Description	**Search Capabilities:**
//	@Description	- Searches agent names (primary field)
//	@Description	- Searches agent email addresses (secondary field)
//	@Description	- Case-insensitive ILIKE matching
//	@Description
//	@Description	**Match Scoring:**
//	@Description	- Name matches: 1.5 score (higher priority)
//	@Description	- Email matches: 1.2 score (lower priority)
//	@Description
//	@Description	**Search Examples:**
//	@Description	- "João" - Find agents named João
//	@Description	- "support" - Find support team members
//	@Description	- "@gmail.com" - Find agents with Gmail addresses
//	@Description	- "sales" - Find sales team agents
//	@Tags			CRM - Agents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string							true	"Search query - name or email"	minlength(1)	example(João Silva)
//	@Param			limit	query		int								false	"Maximum results (max 100)"		default(20)		minimum(1)	maximum(100)	example(10)
//	@Success		200		{object}	queries.SearchAgentsResponse	"Found agents with match scores"
//	@Failure		400		{object}	map[string]interface{}			"Bad Request - Empty search query"
//	@Failure		401		{object}	map[string]interface{}			"Unauthorized"
//	@Failure		403		{object}	map[string]interface{}			"Forbidden"
//	@Failure		500		{object}	map[string]interface{}			"Internal Server Error"
//	@Router			/api/v1/crm/agents/search [get]
func (h *AgentHandler) SearchAgents(c *gin.Context) {
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

	query := queries.SearchAgentsQuery{
		TenantID:   tenantID,
		SearchText: searchText,
		Limit:      limit,
	}

	response, err := h.searchAgentsQueryHandler.Handle(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search agents", zap.Error(err))
		apierrors.InternalError(c, "Failed to search agents", err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateVirtualAgent creates a new virtual agent
//
//	@Summary		Create virtual agent
//	@Description	Creates a virtual agent to represent a person from the past. Virtual agents cannot send messages or be manually assigned - they exist only for metrics segmentation and historical tracking.
//	@Description
//	@Description	**Use Cases:**
//	@Description	- Attribute old messages from external systems (device attribution)
//	@Description	- Track historical contacts when phone numbers change hands
//	@Description	- Segment metrics by time periods and historical users
//	@Description	- Maintain conversation history without granting system access
//	@Description
//	@Description	**Virtual Agent Restrictions:**
//	@Description	- Cannot send messages
//	@Description	- Cannot be manually assigned to conversations
//	@Description	- Do not count in agent performance metrics
//	@Description	- Always appear as offline
//	@Description	- Name is prefixed with "[Virtual]"
//	@Tags			CRM - Agents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			agent	body		CreateVirtualAgentRequest	true	"Virtual agent data"
//	@Success		201		{object}	map[string]interface{}		"Virtual agent created successfully"
//	@Failure		400		{object}	map[string]interface{}		"Invalid request"
//	@Failure		401		{object}	map[string]interface{}		"Unauthorized"
//	@Failure		500		{object}	map[string]interface{}		"Internal server error"
//	@Router			/api/v1/crm/agents/virtual [post]
func (h *AgentHandler) CreateVirtualAgent(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	var req CreateVirtualAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse virtual agent request", zap.Error(err))
		apierrors.ValidationError(c, "body", "Invalid request: "+err.Error())
		return
	}

	// Parse project ID
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		apierrors.ValidationError(c, "project_id", "Invalid project ID format")
		return
	}

	// Parse period_start
	periodStart, err := time.Parse(time.RFC3339, req.PeriodStart)
	if err != nil {
		apierrors.ValidationError(c, "period_start", "Invalid date format (use RFC3339, e.g., 2023-01-01T00:00:00Z)")
		return
	}

	// Create virtual agent
	virtualAgent, err := agent.NewVirtualAgent(
		projectID,
		authCtx.TenantID,
		req.RepresentsPersonName,
		periodStart,
		req.Reason,
		req.SourceDevice,
		req.Notes,
	)
	if err != nil {
		h.logger.Error("Failed to create virtual agent", zap.Error(err))
		apierrors.InternalError(c, "Failed to create virtual agent", err)
		return
	}

	// Save to database
	if err := h.agentRepo.Save(c.Request.Context(), virtualAgent); err != nil {
		h.logger.Error("Failed to save virtual agent", zap.Error(err))
		apierrors.InternalError(c, "Failed to save virtual agent", err)
		return
	}

	h.logger.Info("Virtual agent created successfully",
		zap.String("agent_id", virtualAgent.ID().String()),
		zap.String("represents_person", req.RepresentsPersonName),
		zap.String("reason", req.Reason),
	)

	c.JSON(http.StatusCreated, gin.H{
		"id":                     virtualAgent.ID(),
		"name":                   virtualAgent.Name(),
		"type":                   virtualAgent.Type(),
		"represents_person_name": req.RepresentsPersonName,
		"period_start":           periodStart,
		"reason":                 req.Reason,
		"source_device":          req.SourceDevice,
		"notes":                  req.Notes,
		"created_at":             virtualAgent.CreatedAt(),
	})
}

// EndVirtualAgentPeriod marks the end of a virtual agent's represented period
//
//	@Summary		End virtual agent period
//	@Description	Marks the end date for the period that a virtual agent represents. Useful when a phone number is transferred to a different person.
//	@Description
//	@Description	**Use Case:**
//	@Description	When a WhatsApp number changes ownership, you can:
//	@Description	1. End the period for the old virtual agent (this endpoint)
//	@Description	2. Create a new virtual agent for the new owner
//	@Description	3. Maintain clean historical segmentation in analytics
//	@Tags			CRM - Agents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Virtual agent ID (UUID)"
//	@Param			period	body		EndVirtualAgentPeriodRequest	true	"Period end date"
//	@Success		200		{object}	map[string]interface{}			"Period ended successfully"
//	@Failure		400		{object}	map[string]interface{}			"Invalid request or agent is not virtual"
//	@Failure		401		{object}	map[string]interface{}			"Unauthorized"
//	@Failure		404		{object}	map[string]interface{}			"Agent not found"
//	@Failure		500		{object}	map[string]interface{}			"Internal server error"
//	@Router			/api/v1/crm/agents/{id}/virtual/end-period [put]
func (h *AgentHandler) EndVirtualAgentPeriod(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		apierrors.Unauthorized(c, "Authentication required")
		return
	}

	// Parse agent ID
	idStr := c.Param("id")
	agentID, err := uuid.Parse(idStr)
	if err != nil {
		apierrors.ValidationError(c, "id", "Invalid agent ID format")
		return
	}

	var req EndVirtualAgentPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse request", zap.Error(err))
		apierrors.ValidationError(c, "body", "Invalid request: "+err.Error())
		return
	}

	// Parse period_end
	periodEnd, err := time.Parse(time.RFC3339, req.PeriodEnd)
	if err != nil {
		apierrors.ValidationError(c, "period_end", "Invalid date format (use RFC3339, e.g., 2024-01-01T00:00:00Z)")
		return
	}

	// Find agent
	virtualAgent, err := h.agentRepo.FindByID(c.Request.Context(), agentID)
	if err != nil {
		if err == agent.ErrAgentNotFound {
			apierrors.NotFound(c, "agent", agentID.String())
			return
		}
		h.logger.Error("Failed to find agent", zap.Error(err))
		apierrors.InternalError(c, "Failed to retrieve agent", err)
		return
	}

	// Verify tenant ownership
	if virtualAgent.TenantID() != authCtx.TenantID {
		apierrors.Forbidden(c, "You do not have access to this agent")
		return
	}

	// Verify it's a virtual agent
	if !virtualAgent.IsVirtual() {
		apierrors.ValidationError(c, "id", "Agent is not a virtual agent")
		return
	}

	// End the period
	if err := virtualAgent.EndVirtualAgentPeriod(periodEnd); err != nil {
		h.logger.Error("Failed to end virtual agent period", zap.Error(err))
		apierrors.ValidationError(c, "period_end", err.Error())
		return
	}

	// Save changes
	if err := h.agentRepo.Save(c.Request.Context(), virtualAgent); err != nil {
		h.logger.Error("Failed to save agent", zap.Error(err))
		apierrors.InternalError(c, "Failed to update agent", err)
		return
	}

	h.logger.Info("Virtual agent period ended successfully",
		zap.String("agent_id", virtualAgent.ID().String()),
		zap.Time("period_end", periodEnd),
	)

	metadata := virtualAgent.VirtualMetadata()
	c.JSON(http.StatusOK, gin.H{
		"id":                     virtualAgent.ID(),
		"name":                   virtualAgent.Name(),
		"represents_person_name": metadata.RepresentsPersonName,
		"period_start":           metadata.PeriodStart,
		"period_end":             metadata.PeriodEnd,
		"reason":                 metadata.Reason,
		"updated_at":             virtualAgent.UpdatedAt(),
	})
}
