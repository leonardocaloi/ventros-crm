package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AgentHandler struct {
	logger *zap.Logger
}

func NewAgentHandler(logger *zap.Logger) *AgentHandler {
	return &AgentHandler{
		logger: logger,
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
//	@Tags			agents
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
//	@Tags			agents
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
//	@Tags			agents
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
//	@Tags			agents
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
//	@Tags			agents
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
//	@Tags			agents
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
