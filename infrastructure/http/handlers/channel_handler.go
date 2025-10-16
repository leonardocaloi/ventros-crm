package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/http/middleware"
	channelapp "github.com/ventros/crm/internal/application/channel"
	channelcmd "github.com/ventros/crm/internal/application/commands/channel"
	channelworkflow "github.com/ventros/crm/internal/workflows/channel"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type ChannelHandler struct {
	logger                 *zap.Logger
	channelService         *channelapp.ChannelService
	activateChannelHandler *channelcmd.ActivateChannelHandler
	importHistoryHandler   *channelcmd.ImportHistoryHandler
	temporalClient         interface{}
}

func NewChannelHandler(
	logger *zap.Logger,
	channelService *channelapp.ChannelService,
	activateChannelHandler *channelcmd.ActivateChannelHandler,
	importHistoryHandler *channelcmd.ImportHistoryHandler,
	temporalClient interface{},
) *ChannelHandler {
	return &ChannelHandler{
		logger:                 logger,
		channelService:         channelService,
		activateChannelHandler: activateChannelHandler,
		importHistoryHandler:   importHistoryHandler,
		temporalClient:         temporalClient,
	}
}

// CreateChannelRequest represents the request to create a channel.
type CreateChannelRequest struct {
	Name                  string `json:"name" binding:"required" example:"WhatsApp Principal"`
	Type                  string `json:"type" binding:"required" example:"waha"`
	SessionTimeoutMinutes *int   `json:"session_timeout_minutes,omitempty" example:"30"`
	AllowGroups           *bool  `json:"allow_groups,omitempty" example:"false"`
	TrackingEnabled       *bool  `json:"tracking_enabled,omitempty" example:"true"`

	// AI Configuration
	AIEnabled         *bool `json:"ai_enabled,omitempty" example:"false"`
	AIAgentsEnabled   *bool `json:"ai_agents_enabled,omitempty" example:"false"`
	DebounceTimeoutMs *int  `json:"debounce_timeout_ms,omitempty" example:"15000"`

	WAHAConfig *CreateWAHAConfigRequest `json:"waha_config,omitempty"`
}

// CreateWAHAConfigRequest represents WAHA configuration.
type CreateWAHAConfigRequest struct {
	BaseURL    string `json:"base_url" binding:"required" example:"http://localhost:3000"`
	APIKey     string `json:"api_key" example:"your-waha-api-key"`
	Token      string `json:"token" example:"your-waha-token"`
	SessionID  string `json:"session_id" example:"default"`
	WebhookURL string `json:"webhook_url" example:"http://localhost:8080/api/v1/webhooks/waha"`
}

// CreateChannel creates a new communication channel
//
//	@Summary		Create channel
//	@Description	Create a new communication channel (WAHA, WhatsApp, etc.)
//	@Tags			CRM - Channels
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			channel	body		CreateChannelRequest	true	"Channel data"
//	@Success		201		{object}	map[string]interface{}	"Channel created successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		401		{object}	map[string]interface{}	"Authentication required"
//	@Failure		403		{object}	map[string]interface{}	"Access denied"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse channel request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	serviceReq := channelapp.CreateChannelRequest{
		UserID:                authCtx.UserID,
		ProjectID:             authCtx.ProjectID,
		TenantID:              authCtx.TenantID,
		Name:                  req.Name,
		Type:                  req.Type,
		SessionTimeoutMinutes: req.SessionTimeoutMinutes,
		AllowGroups:           req.AllowGroups,
		TrackingEnabled:       req.TrackingEnabled,
		AIEnabled:             req.AIEnabled,
		AIAgentsEnabled:       req.AIAgentsEnabled,
		DebounceTimeoutMs:     req.DebounceTimeoutMs,
	}

	if req.WAHAConfig != nil {
		serviceReq.WAHAConfig = &channelapp.WAHAConfigRequest{
			BaseURL:    req.WAHAConfig.BaseURL,
			APIKey:     req.WAHAConfig.APIKey,
			Token:      req.WAHAConfig.Token,
			SessionID:  req.WAHAConfig.SessionID,
			WebhookURL: req.WAHAConfig.WebhookURL,
		}
	}

	response, err := h.channelService.CreateChannel(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to create channel", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Channel created successfully",
		"id":      response.ID,
		"channel": response,
	})
}

// UpdateChannelRequest represents the request to update a channel
type UpdateChannelRequest struct {
	SessionTimeoutMinutes *int `json:"session_timeout_minutes,omitempty" example:"120"`
}

// UpdateChannel updates an existing channel
//
//	@Summary		Update channel
//	@Description	Update channel configuration (e.g., session timeout)
//	@Tags			CRM - Channels
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		string					true	"Channel ID"
//	@Param			channel	body		UpdateChannelRequest	true	"Channel update data"
//	@Success		200		{object}	map[string]interface{}	"Channel updated successfully"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		401		{object}	map[string]interface{}	"Authentication required"
//	@Failure		404		{object}	map[string]interface{}	"Channel not found"
//	@Failure		500		{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id} [patch]
func (h *ChannelHandler) UpdateChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID := c.Param("id")

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	serviceReq := channelapp.UpdateChannelRequest{
		ChannelID:             channelID,
		UserID:                authCtx.UserID,
		SessionTimeoutMinutes: req.SessionTimeoutMinutes,
	}

	err := h.channelService.UpdateChannel(c.Request.Context(), serviceReq)
	if err != nil {
		h.logger.Error("Failed to update channel", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Channel updated successfully",
	})
}

// ListChannels lists all channels for the authenticated user
//
//	@Summary		List channels
//	@Description	List all channels for authenticated user
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	map[string]interface{}	"Channels list"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channels, err := h.channelService.GetChannelsByUser(c.Request.Context(), authCtx.UserID)
	if err != nil {
		h.logger.Error("Failed to list channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"total":    len(channels),
	})
}

// GetChannel gets a specific channel
//
//	@Summary		Get channel
//	@Description	Get details of a specific channel
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"Channel details"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id} [get]
func (h *ChannelHandler) GetChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		h.logger.Error("Failed to get channel", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel": channel,
	})
}

// ActivateChannel activates a channel asynchronously
//
//	@Summary		Activate channel
//	@Description	Request channel activation (async processing via events)
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		202	{object}	map[string]interface{}	"Activation requested (processing asynchronously)"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID or already active"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id}/activate [post]
func (h *ChannelHandler) ActivateChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Verify channel exists and user owns it
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// 🚀 FAST SYNC ACTIVATION: Direct WAHA health check (1 request → active)
	// For test/dev environments, we activate synchronously
	// TODO: Restore async Temporal workflow for webhook setup in production
	/*
	// ASYNC VERSION (commented out for now - for webhook setup)
	cmd := channelcmd.ActivateChannelCommand{
		ChannelID: channelID,
		TenantID:  authCtx.TenantID,
	}
	if err := h.activateChannelHandler.Handle(c.Request.Context(), cmd); err != nil {
		// ... error handling
	}
	*/

	// Synchronous activation (fast)
	if err := h.channelService.ActivateChannel(c.Request.Context(), channelID); err != nil {
		h.logger.Error("Failed to activate channel",
			zap.Error(err),
			zap.String("channel_id", channelID.String()))

		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to activate channel: " + err.Error()})
		return
	}

	// Return 200 OK - activation is SYNCHRONOUS now
	c.JSON(http.StatusOK, gin.H{
		"message":    "Channel activated successfully",
		"channel_id": channelID,
		"status":     "active",
		"note":       "Channel is now active (synchronous activation)",
	})
}

// DeactivateChannel deactivates a channel
//
//	@Summary		Deactivate channel
//	@Description	Deactivate a communication channel
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"Channel deactivated"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id}/deactivate [post]
func (h *ChannelHandler) DeactivateChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.channelService.DeactivateChannel(c.Request.Context(), channelID); err != nil {
		h.logger.Error("Failed to deactivate channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Channel deactivated successfully",
	})
}

// DeleteChannel deletes a channel
//
//	@Summary		Delete channel
//	@Description	Delete a communication channel
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"Channel deleted"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.channelService.DeleteChannel(c.Request.Context(), channelID); err != nil {
		h.logger.Error("Failed to delete channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Channel deleted successfully",
	})
}

// GetChannelWebhookURL gets the webhook URL for configuring the external channel
//
//	@Summary		Get channel webhook URL
//	@Description	Return the webhook URL to be configured in external channel (WAHA, WhatsApp, etc)
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"Webhook URL"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Router			/api/v1/channels/{id}/webhook-url [get]
func (h *ChannelHandler) GetChannelWebhookURL(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	baseURL := c.GetHeader("X-Base-URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	webhookURL, err := h.channelService.GetWebhookURL(c.Request.Context(), channelID, baseURL)
	if err != nil {
		h.logger.Error("Failed to get webhook URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel_id":   channelID,
		"channel_name": channel.Name,
		"channel_type": channel.Type,
		"webhook_url":  webhookURL,
		"instructions": map[string]string{
			"waha":     "Configure this URL in WAHA session webhooks",
			"whatsapp": "Configure this URL in WhatsApp Business API webhooks",
			"telegram": "Use this URL when setting up Telegram bot webhook",
		}[channel.Type],
	})
}

// ConfigureChannelWebhook configures the webhook automatically in the external channel
//
//	@Summary		Configure channel webhook
//	@Description	Automatically configure webhook in external channel (e.g. WAHA)
//	@Tags			CRM - Channels
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		string					true	"Channel ID"
//	@Param			request	body		ConfigureWebhookRequest	false	"Webhook configuration (optional, uses default if not provided)"
//	@Success		200		{object}	map[string]interface{}	"Webhook configured"
//	@Failure		400		{object}	map[string]interface{}	"Invalid request"
//	@Failure		401		{object}	map[string]interface{}	"Authentication required"
//	@Failure		404		{object}	map[string]interface{}	"Channel not found"
//	@Failure		500		{object}	map[string]interface{}	"Configuration failed"
//	@Router			/api/v1/channels/{id}/configure-webhook [post]
func (h *ChannelHandler) ConfigureChannelWebhook(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req ConfigureWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.BaseURL = "http://localhost:8080"
	}

	if req.BaseURL == "" {
		req.BaseURL = c.GetHeader("X-Base-URL")
		if req.BaseURL == "" {
			req.BaseURL = "http://localhost:8080"
		}
	}

	webhookURL, err := h.channelService.GetWebhookURL(c.Request.Context(), channelID, req.BaseURL)
	if err != nil {
		h.logger.Error("Failed to get webhook URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.channelService.ConfigureWebhook(c.Request.Context(), channelID, webhookURL); err != nil {
		h.logger.Error("Failed to configure webhook", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure webhook: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Webhook configured successfully",
		"channel_id":   channelID,
		"channel_name": channel.Name,
		"webhook_url":  webhookURL,
	})
}

// GetChannelWebhookInfo gets information about the channel webhook
//
//	@Summary		Get channel webhook info
//	@Description	Return detailed information about channel webhook
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"Webhook info"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Router			/api/v1/channels/{id}/webhook-info [get]
func (h *ChannelHandler) GetChannelWebhookInfo(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	info, err := h.channelService.GetWebhookInfo(c.Request.Context(), channelID)
	if err != nil {
		h.logger.Error("Failed to get webhook info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ConfigureWebhookRequest represents the request to configure webhook.
type ConfigureWebhookRequest struct {
	BaseURL string `json:"base_url" example:"https://api.ventros.com"`
}

// GetChannelQRCode gets the QR code of a WAHA channel.
// TEMPORARILY DISABLED - NEEDS SERVICE METHODS IMPLEMENTATION
/*
// @Summary Get channel QR code
// @Description Obtém o QR code para conectar um canal WAHA
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "QR code data"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID or not WAHA type"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels/{id}/qr [get]
func (h *ChannelHandler) GetChannelQRCode(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !channel.IsWAHA() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel is not WAHA type"})
		return
	}

	status := channel.GetWAHASessionStatus()

	response := gin.H{
		"channel_id": channel.ID,
		"session_id": channel.ExternalID,
		"status":     string(status),
	}

	if status == channel.WAHASessionStatusWorking {
		response["message"] = "Channel is already connected"
		c.JSON(http.StatusOK, response)
		return
	}

	if channel.IsWAHAQRCodeValid() {
		response["qr_code"] = channel.GetWAHAQRCode()
		if generatedAt, ok := channel.Config["qr_generated_at"].(int64); ok {
			response["qr_generated_at"] = generatedAt
			response["qr_expires_at"] = generatedAt + 120
		}
		c.JSON(http.StatusOK, response)
		return
	}

	if channel.NeedsNewQRCode() {
		response["message"] = "QR code expired or not available. Please request a new session from WAHA."
		response["needs_new_qr"] = true
		c.JSON(http.StatusOK, response)
		return
	}

	response["message"] = "Channel not ready for QR code"
	c.JSON(http.StatusOK, response)
}
*/

// ActivateWAHAChannel activates a specific WAHA channel
//
//	@Summary		Activate WAHA channel
//	@Description	Activate and initialize a WAHA session for a channel
//	@Tags			CRM - Channels
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"WAHA channel activated"
//	@Failure		400	{object}	map[string]interface{}	"Invalid request or not WAHA channel"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id}/activate-waha [post]
func (h *ChannelHandler) ActivateWAHAChannel(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if channel.Type != "waha" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel is not WAHA type"})
		return
	}

	if err := h.channelService.ActivateChannel(c.Request.Context(), channelID); err != nil {
		h.logger.Error("Failed to activate WAHA channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate WAHA channel: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "WAHA channel activated successfully",
		"channel_id": channelID,
		"next_steps": []string{
			"Scan QR code if not already authenticated",
			"Check webhook configuration",
			"Send test message",
		},
	})
}

// ImportWAHAHistoryRequest represents the request to import history.
type ImportWAHAHistoryRequest struct {
	Strategy      string `json:"strategy" example:"recent"`
	Limit         int    `json:"limit" example:"100"`
	TimeRangeDays int    `json:"time_range_days" example:"7"` // Dias para filtrar mensagens (0 = usar config do canal)
}

// ImportWAHAHistory imports message history from a WAHA channel
//
//	@Summary		Import WAHA message history
//	@Description	Import message history from a WAHA channel (chats and messages)
//	@Tags			CRM - Channels
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id		path		string						true	"Channel ID"
//	@Param			request	body		ImportWAHAHistoryRequest	false	"Import configuration (optional)"
//	@Success		202		{object}	map[string]interface{}		"Import started (async)"
//	@Failure		400		{object}	map[string]interface{}		"Invalid request or not WAHA channel"
//	@Failure		401		{object}	map[string]interface{}		"Authentication required"
//	@Failure		404		{object}	map[string]interface{}		"Channel not found"
//	@Failure		500		{object}	map[string]interface{}		"Internal server error"
//	@Router			/api/v1/channels/{id}/import-history [post]
func (h *ChannelHandler) ImportWAHAHistory(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Verify channel exists and user owns it
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse request body
	var req ImportWAHAHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Se JSON parsing falhar, usar valores padrão
		req.Strategy = "recent"
		// ❌ NÃO definir req.Limit = 100
		// deixar limit == 0 (SEM LIMITE)
	}

	if req.Strategy == "" {
		req.Strategy = "recent"
	}
	// limit == 0 significa SEM LIMITE (buscar todas as mensagens disponíveis)

	// Determinar TimeRangeDays: priorizar request, depois config do canal
	timeRangeDays := req.TimeRangeDays
	if timeRangeDays == 0 && channel.HistoryImportMaxDays != nil && *channel.HistoryImportMaxDays > 0 {
		timeRangeDays = *channel.HistoryImportMaxDays
	}

	// 🔥 FIX Bug 2: Session timeout is now loaded by workflow from channel config
	// No need to pass it here - workflow will fetch it via GetChannelConfigActivity
	// Create command (follows ActivateChannel pattern)
	cmd := channelcmd.ImportHistoryCommand{
		ChannelID:             channelID,
		TenantID:              authCtx.TenantID,
		Strategy:              req.Strategy,
		TimeRangeDays:         timeRangeDays,
		Limit:                 req.Limit,
		SessionTimeoutMinutes: 0, // Workflow will fetch from channel config
		UserID:                authCtx.UserID,
	}

	// Execute command (async processing via events)
	correlationID, err := h.importHistoryHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		h.logger.Error("Failed to request history import",
			zap.Error(err),
			zap.String("channel_id", channelID.String()))

		// Handle specific errors
		if err == channelcmd.ErrChannelNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
			return
		}
		if err == channelcmd.ErrRepositorySaveFailed {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save import request"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("History import requested successfully (async processing)",
		zap.String("channel_id", channelID.String()),
		zap.String("correlation_id", correlationID),
		zap.String("strategy", req.Strategy),
		zap.Int("time_range_days", timeRangeDays),
		zap.Int("limit", req.Limit))

	// Return 202 Accepted - import is async
	c.JSON(http.StatusAccepted, gin.H{
		"message":         "History import requested",
		"channel_id":      channelID,
		"correlation_id":  correlationID,
		"strategy":        req.Strategy,
		"limit":           req.Limit,
		"time_range_days": timeRangeDays,
		"status":          "processing",
		"note":            "Import is processing asynchronously. Poll /channels/{id} or /channels/{id}/import-status to check progress.",
	})
}

// GetWAHAImportStatus gets the status of a WAHA history import workflow
//
//	@Summary		Get WAHA import status
//	@Description	Get the current status and progress of a WAHA history import
//	@Tags			CRM - Channels
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	path		string					true	"Channel ID"
//	@Success		200	{object}	map[string]interface{}	"Import status"
//	@Failure		400	{object}	map[string]interface{}	"Invalid channel ID"
//	@Failure		401	{object}	map[string]interface{}	"Authentication required"
//	@Failure		404	{object}	map[string]interface{}	"Channel not found or no import running"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/channels/{id}/import-status [get]
func (h *ChannelHandler) GetWAHAImportStatus(c *gin.Context) {
	authCtx, exists := middleware.GetAuthContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if h.temporalClient == nil {
		h.logger.Error("Temporal client not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Workflow engine not configured"})
		return
	}

	temporalClient, ok := h.temporalClient.(client.Client)
	if !ok {
		h.logger.Error("Invalid Temporal client type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid workflow engine configuration"})
		return
	}

	workflowID := fmt.Sprintf("waha-import-%s", channelID.String())

	// Try to describe the workflow
	descResp, err := temporalClient.DescribeWorkflowExecution(c.Request.Context(), workflowID, "")
	if err != nil {
		// No workflow found - check channel for historical import data
		response := gin.H{
			"channel_id":  channelID,
			"workflow_id": workflowID,
			"status":      "no_import_running",
		}

		// Add channel's last import info if available
		if channel.LastImportDate != nil {
			response["last_import_date"] = channel.LastImportDate
			response["history_import_status"] = channel.HistoryImportStatus
			response["history_import_stats"] = channel.HistoryImportStats
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// Get workflow execution info
	workflowInfo := descResp.WorkflowExecutionInfo
	workflowStatus := workflowInfo.Status.String()

	response := gin.H{
		"channel_id":  channelID,
		"workflow_id": workflowID,
		"run_id":      workflowInfo.Execution.RunId,
		"status":      workflowStatus,
		"start_time":  workflowInfo.StartTime,
	}

	// Add close time if workflow is completed
	if workflowInfo.CloseTime != nil {
		response["close_time"] = workflowInfo.CloseTime
	}

	// Query workflow for current progress
	queryResp, err := temporalClient.QueryWorkflow(c.Request.Context(), workflowID, "", "import-status")
	if err == nil {
		var result *channelworkflow.WAHAHistoryImportWorkflowResult
		if err := queryResp.Get(&result); err == nil && result != nil {
			response["progress"] = gin.H{
				"chats_processed":   result.ChatsProcessed,
				"messages_imported": result.MessagesImported,
				"sessions_created":  result.SessionsCreated,
				"contacts_created":  result.ContactsCreated,
				"errors":            result.Errors,
				"started_at":        result.StartedAt,
				"completed_at":      result.CompletedAt,
				"import_status":     result.Status,
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
