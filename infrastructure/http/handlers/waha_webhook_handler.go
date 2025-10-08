package handlers

import (
	"io"
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WAHAWebhookHandler struct {
	logger       *zap.Logger
	rawEventBus  *messaging.WAHARawEventBus
}

func NewWAHAWebhookHandler(
	logger *zap.Logger,
	rawEventBus *messaging.WAHARawEventBus,
) *WAHAWebhookHandler {
	return &WAHAWebhookHandler{
		logger:      logger,
		rawEventBus: rawEventBus,
	}
}

// ReceiveWebhook receives WAHA webhook events
// @Summary Receive WAHA webhook
// @Description Recebe eventos de webhook do WAHA (mensagens, chamadas, etc.)
// @Tags webhooks
// @Accept json
// @Produce json
// @Param session path string true "Session ID"
// @Success 200 {object} map[string]interface{} "Event queued"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhooks/waha/{session} [post]
func (h *WAHAWebhookHandler) ReceiveWebhook(c *gin.Context) {
	// Ler o corpo da requisição
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Extrair headers relevantes
	headers := map[string]string{
		"Content-Type":   c.GetHeader("Content-Type"),
		"User-Agent":     c.GetHeader("User-Agent"),
		"X-Forwarded-For": c.GetHeader("X-Forwarded-For"),
	}

	// Criar evento raw
	rawEvent := waha.NewWAHARawEvent(
		c.Param("session"), // Session ID do path param
		body,
		headers,
	)

	// Log do evento recebido
	h.logger.Info("WAHA webhook received",
		zap.String("event_id", rawEvent.ID),
		zap.String("session", rawEvent.Session),
		zap.String("content_type", rawEvent.GetContentType()),
		zap.Int("body_size", rawEvent.GetBodySize()))

	// Enfileirar evento raw (NUNCA falha)
	if err := h.rawEventBus.PublishRawEvent(c.Request.Context(), rawEvent); err != nil {
		// Log mas não falha - o evento pode ter sido perdido mas o webhook não quebra
		h.logger.Error("Failed to enqueue raw event", 
			zap.Error(err),
			zap.String("event_id", rawEvent.ID))
		
		// Ainda assim responde sucesso para não quebrar o WAHA
		// O evento será perdido, mas é melhor que quebrar todo o fluxo
	}

	// Resposta imediata - webhook nunca falha
	c.JSON(http.StatusOK, gin.H{
		"status":   "queued",
		"event_id": rawEvent.ID,
		"session":  rawEvent.Session,
		"message":  "Event queued for processing",
	})
}


// GetWebhookInfo provides information about the webhook endpoint
// @Summary Get webhook info
// @Description Retorna informações sobre o endpoint de webhook WAHA
// @Tags webhooks
// @Produce json
// @Success 200 {object} map[string]interface{} "Webhook info"
// @Router /api/v1/webhooks/waha [get]
func (h *WAHAWebhookHandler) GetWebhookInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"endpoint":        "/api/v1/webhooks/waha/{session}",
		"method":          "POST",
		"content_type":    "application/json",
		"supported_events": waha.GetDefaultWebhookEvents(),
		"description":     "Endpoint para receber eventos do WAHA (WhatsApp HTTP API)",
		"example_usage": map[string]interface{}{
			"url": "http://localhost:8080/api/v1/webhooks/waha/your-session-id",
			"events": []string{
				"message",
				"message.ack",
				"call.received",
			},
		},
	})
}
