package handlers

import (
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/http/middleware"
	channelapp "github.com/caloi/ventros-crm/internal/application/channel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ChannelHandler struct {
	logger         *zap.Logger
	channelService *channelapp.ChannelService
}

func NewChannelHandler(logger *zap.Logger, channelService *channelapp.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		logger:         logger,
		channelService: channelService,
	}
}

// CreateChannelRequest representa o payload para criar um canal
type CreateChannelRequest struct {
	Name       string                        `json:"name" binding:"required" example:"WhatsApp Principal"`
	Type       string                        `json:"type" binding:"required" example:"waha"`
	WAHAConfig *CreateWAHAConfigRequest      `json:"waha_config,omitempty"`
}

// CreateWAHAConfigRequest representa a configuração WAHA
type CreateWAHAConfigRequest struct {
	BaseURL    string `json:"base_url" binding:"required" example:"http://localhost:3000"`
	APIKey     string `json:"api_key" example:"your-waha-api-key"`     // Chave da API para autenticação
	Token      string `json:"token" example:"your-waha-token"`         // Token de acesso (alternativo)
	SessionID  string `json:"session_id" example:"default"`            // ID da sessão (equivale ao ExternalID)
	WebhookURL string `json:"webhook_url" example:"http://localhost:8080/api/v1/webhooks/waha"`
}

// CreateChannel creates a new communication channel
// @Summary Create channel
// @Description Cria um novo canal de comunicação (WAHA, WhatsApp, etc.)
// @Tags channels
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param channel body CreateChannelRequest true "Channel data"
// @Success 201 {object} map[string]interface{} "Channel created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 403 {object} map[string]interface{} "Access denied"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels [post]
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	// Obter contexto do usuário autenticado
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

	// Preparar request para o service
	serviceReq := channelapp.CreateChannelRequest{
		UserID:    authCtx.UserID,
		ProjectID: authCtx.ProjectID,
		TenantID:  authCtx.TenantID,
		Name:      req.Name,
		Type:      req.Type,
	}

	// Configuração WAHA se fornecida
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

// ListChannels lists all channels for the authenticated user
// @Summary List channels
// @Description Lista todos os canais do usuário autenticado
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Channels list"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels [get]
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
// @Summary Get channel
// @Description Obtém detalhes de um canal específico
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "Channel details"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels/{id} [get]
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

	// Verificar se o canal pertence ao usuário (RLS já filtra, mas double-check)
	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"channel": channel,
	})
}

// ActivateChannel activates a channel
// @Summary Activate channel
// @Description Ativa um canal de comunicação
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "Channel activated"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels/{id}/activate [post]
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

	// Verificar se o canal existe e pertence ao usuário
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.channelService.ActivateChannel(c.Request.Context(), channelID); err != nil {
		h.logger.Error("Failed to activate channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Channel activated successfully",
	})
}

// DeactivateChannel deactivates a channel
// @Summary Deactivate channel
// @Description Desativa um canal de comunicação
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "Channel deactivated"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels/{id}/deactivate [post]
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

	// Verificar se o canal existe e pertence ao usuário
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
// @Summary Delete channel
// @Description Deleta um canal de comunicação
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "Channel deleted"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/channels/{id} [delete]
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

	// Verificar se o canal existe e pertence ao usuário
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

// GetChannelWebhookURL obtém a URL do webhook para configurar no canal externo
// @Summary Get channel webhook URL
// @Description Retorna a URL do webhook que deve ser configurada no canal externo (WAHA, WhatsApp, etc)
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "Webhook URL"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Router /api/v1/channels/{id}/webhook-url [get]
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

	// Verificar se o canal existe e pertence ao usuário
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Obter base URL da aplicação (pode vir de env ou header)
	baseURL := c.GetHeader("X-Base-URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Fallback
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
			"waha":      "Configure this URL in WAHA session webhooks",
			"whatsapp":  "Configure this URL in WhatsApp Business API webhooks",
			"telegram":  "Use this URL when setting up Telegram bot webhook",
		}[channel.Type],
	})
}

// ConfigureChannelWebhook configura o webhook automaticamente no canal externo
// @Summary Configure channel webhook
// @Description Configura automaticamente o webhook no canal externo (ex: WAHA)
// @Tags channels
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Param request body ConfigureWebhookRequest false "Webhook configuration (optional, uses default if not provided)"
// @Success 200 {object} map[string]interface{} "Webhook configured"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Configuration failed"
// @Router /api/v1/channels/{id}/configure-webhook [post]
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

	// Verificar se o canal existe e pertence ao usuário
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse request (opcional)
	var req ConfigureWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Se não tiver body, usa valores padrão
		req.BaseURL = "http://localhost:8080"
	}

	// Se não forneceu base URL, tenta obter do header ou usa padrão
	if req.BaseURL == "" {
		req.BaseURL = c.GetHeader("X-Base-URL")
		if req.BaseURL == "" {
			req.BaseURL = "http://localhost:8080"
		}
	}

	// Gerar URL do webhook
	webhookURL, err := h.channelService.GetWebhookURL(c.Request.Context(), channelID, req.BaseURL)
	if err != nil {
		h.logger.Error("Failed to get webhook URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Configurar webhook no canal externo
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

// GetChannelWebhookInfo obtém informações sobre o webhook do canal
// @Summary Get channel webhook info
// @Description Retorna informações detalhadas sobre o webhook do canal
// @Tags channels
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]interface{} "Webhook info"
// @Failure 400 {object} map[string]interface{} "Invalid channel ID"
// @Failure 401 {object} map[string]interface{} "Authentication required"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Router /api/v1/channels/{id}/webhook-info [get]
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

	// Verificar se o canal existe e pertence ao usuário
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Obter informações do webhook
	info, err := h.channelService.GetWebhookInfo(c.Request.Context(), channelID)
	if err != nil {
		h.logger.Error("Failed to get webhook info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ConfigureWebhookRequest representa o request para configurar webhook
type ConfigureWebhookRequest struct {
	BaseURL string `json:"base_url" example:"https://api.ventros.com"`
}

// GetChannelQRCode obtém o QR code de um canal WAHA
// TEMPORARIAMENTE COMENTADO - PRECISA IMPLEMENTAR MÉTODOS NO SERVICE
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

	// Verificar se o canal existe e pertence ao usuário
	channel, err := h.channelService.GetChannel(c.Request.Context(), channelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	if channel.UserID != authCtx.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Verificar se é canal WAHA
	if !channel.IsWAHA() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel is not WAHA type"})
		return
	}

	// Verificar status da sessão
	status := channel.GetWAHASessionStatus()
	
	response := gin.H{
		"channel_id": channel.ID,
		"session_id": channel.ExternalID,
		"status":     string(status),
	}

	// Se já está conectado, não precisa de QR code
	if status == channel.WAHASessionStatusWorking {
		response["message"] = "Channel is already connected"
		c.JSON(http.StatusOK, response)
		return
	}

	// Se tem QR code válido, retorna ele
	if channel.IsWAHAQRCodeValid() {
		response["qr_code"] = channel.GetWAHAQRCode()
		if generatedAt, ok := channel.Config["qr_generated_at"].(int64); ok {
			response["qr_generated_at"] = generatedAt
			response["qr_expires_at"] = generatedAt + 120 // 2 minutos
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Se precisa de novo QR code, indica que deve ser solicitado via WAHA
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
