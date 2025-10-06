package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	webhookapp "github.com/caloi/ventros-crm/internal/application/webhook"
	"github.com/caloi/ventros-crm/internal/domain/webhook"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WebhookSubscriptionHandler struct {
	logger  *zap.Logger
	useCase *webhookapp.ManageSubscriptionUseCase
}

func NewWebhookSubscriptionHandler(logger *zap.Logger, useCase *webhookapp.ManageSubscriptionUseCase) *WebhookSubscriptionHandler {
	return &WebhookSubscriptionHandler{
		logger:  logger,
		useCase: useCase,
	}
}

type CreateWebhookRequest struct {
	Name           string            `json:"name" binding:"required" example:"N8N Webhook"`
	URL            string            `json:"url" binding:"required,url" example:"https://n8n.example.com/webhook/waha-events"`
	Events         []string          `json:"events" binding:"required" example:"message,ack,call.received"`
	Secret         string            `json:"secret,omitempty" example:"my-secret-key"`
	Headers        map[string]string `json:"headers,omitempty" example:"Authorization:Bearer token123"`
	RetryCount     *int              `json:"retry_count,omitempty" example:"3"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty" example:"30"`
}

type UpdateWebhookRequest struct {
	Name           *string           `json:"name,omitempty"`
	URL            *string           `json:"url,omitempty"`
	Events         []string          `json:"events,omitempty"`
	Active         *bool             `json:"active,omitempty"`
	Secret         *string           `json:"secret,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	RetryCount     *int              `json:"retry_count,omitempty"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty"`
}

// CreateWebhook creates a new webhook subscription
// @Summary Create webhook subscription
// @Description Cria uma nova inscrição de webhook para receber eventos do WAHA
// @Tags webhooks
// @Accept json
// @Produce json
// @Param webhook body CreateWebhookRequest true "Webhook subscription data"
// @Success 201 {object} map[string]interface{} "Webhook created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhook-subscriptions [post]
func (h *WebhookSubscriptionHandler) CreateWebhook(c *gin.Context) {
	// Obter contexto do usuário autenticado
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse webhook request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	retryCount := 3
	if req.RetryCount != nil {
		retryCount = *req.RetryCount
	}

	timeoutSeconds := 30
	if req.TimeoutSeconds != nil {
		timeoutSeconds = *req.TimeoutSeconds
	}

	dto := webhookapp.CreateWebhookDTO{
		UserID:         authCtx.UserID,
		ProjectID:      authCtx.ProjectID,
		TenantID:       authCtx.TenantID,
		Name:           req.Name,
		URL:            req.URL,
		Events:         req.Events,
		Secret:         req.Secret,
		Headers:        req.Headers,
		RetryCount:     retryCount,
		TimeoutSeconds: timeoutSeconds,
	}

	result, err := h.useCase.CreateWebhook(c.Request.Context(), dto)
	if err != nil {
		h.logger.Error("Failed to create webhook", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ListWebhooks lists all webhook subscriptions
// @Summary List webhook subscriptions
// @Description Lista todas as inscrições de webhooks
// @Tags webhooks
// @Produce json
// @Param active query bool false "Filter by active status"
// @Success 200 {array} map[string]interface{} "List of webhooks"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhook-subscriptions [get]
func (h *WebhookSubscriptionHandler) ListWebhooks(c *gin.Context) {
	var activeOnly *bool
	if activeStr := c.Query("active"); activeStr != "" {
		if activeStr == "true" {
			active := true
			activeOnly = &active
		} else if activeStr == "false" {
			active := false
			activeOnly = &active
		}
	}

	webhooks, err := h.useCase.ListWebhooks(c.Request.Context(), activeOnly)
	if err != nil {
		h.logger.Error("Failed to list webhooks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// GetWebhook gets a webhook subscription by ID
// @Summary Get webhook subscription
// @Description Obtém detalhes de uma inscrição de webhook
// @Tags webhooks
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Success 200 {object} map[string]interface{} "Webhook details"
// @Failure 404 {object} map[string]interface{} "Webhook not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhook-subscriptions/{id} [get]
func (h *WebhookSubscriptionHandler) GetWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	result, err := h.useCase.GetWebhook(c.Request.Context(), id)
	if err != nil {
		if err == webhook.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
			return
		}
		h.logger.Error("Failed to get webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get webhook"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateWebhook updates a webhook subscription
// @Summary Update webhook subscription
// @Description Atualiza uma inscrição de webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Param webhook body UpdateWebhookRequest true "Webhook update data"
// @Success 200 {object} map[string]interface{} "Webhook updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Webhook not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhook-subscriptions/{id} [put]
func (h *WebhookSubscriptionHandler) UpdateWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	dto := webhookapp.UpdateWebhookDTO{
		Name:           req.Name,
		URL:            req.URL,
		Events:         req.Events,
		Active:         req.Active,
		Secret:         req.Secret,
		Headers:        req.Headers,
		RetryCount:     req.RetryCount,
		TimeoutSeconds: req.TimeoutSeconds,
	}

	result, err := h.useCase.UpdateWebhook(c.Request.Context(), id, dto)
	if err != nil {
		if err == webhook.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
			return
		}
		h.logger.Error("Failed to update webhook", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteWebhook deletes a webhook subscription
// @Summary Delete webhook subscription
// @Description Remove uma inscrição de webhook
// @Tags webhooks
// @Produce json
// @Param id path string true "Webhook ID (UUID)"
// @Success 204 "Webhook deleted successfully"
// @Failure 404 {object} map[string]interface{} "Webhook not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhook-subscriptions/{id} [delete]
func (h *WebhookSubscriptionHandler) DeleteWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	err = h.useCase.DeleteWebhook(c.Request.Context(), id)
	if err != nil {
		if err == webhook.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
			return
		}
		h.logger.Error("Failed to delete webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAvailableEvents returns list of available WAHA events
// @Summary Get available WAHA events
// @Description Lista todos os eventos WAHA disponíveis para inscrição
// @Tags webhooks
// @Produce json
// @Success 200 {object} map[string]interface{} "Available events"
// @Router /api/v1/webhook-subscriptions/available-events [get]
func (h *WebhookSubscriptionHandler) GetAvailableEvents(c *gin.Context) {
	events := h.useCase.GetAvailableEvents()
	
	c.JSON(http.StatusOK, gin.H{
		"events":       events,
		"queue_prefix": "waha.events",
	})
}
