package handlers

import (
	"io"
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/internal/domain/crm/channel"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WAHAWebhookHandler struct {
	logger      *zap.Logger
	rawEventBus *messaging.WAHARawEventBus
	channelRepo channel.Repository
}

func NewWAHAWebhookHandler(
	logger *zap.Logger,
	rawEventBus *messaging.WAHARawEventBus,
	channelRepo channel.Repository,
) *WAHAWebhookHandler {
	return &WAHAWebhookHandler{
		logger:      logger,
		rawEventBus: rawEventBus,
		channelRepo: channelRepo,
	}
}

// ReceiveWebhook receives WAHA webhook events using unique webhook ID
//
//	@Summary		Receive WAHA webhook
//	@Description	Recebe eventos de webhook do WAHA usando ID único do webhook (padrão indústria)
//	@Tags			webhooks
//	@Accept			json
//	@Produce		json
//	@Param			webhook_id	path		string					true	"Webhook ID único"
//	@Success		200			{object}	map[string]interface{}	"Event queued"
//	@Failure		400			{object}	map[string]interface{}	"Invalid request"
//	@Failure		404			{object}	map[string]interface{}	"Webhook not found"
//	@Failure		500			{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/webhooks/{webhook_id} [post]
func (h *WAHAWebhookHandler) ReceiveWebhook(c *gin.Context) {
	webhookID := c.Param("webhook_id")

	// 1. Buscar channel pelo webhook_id
	ch, err := h.channelRepo.GetByWebhookID(webhookID)
	if err != nil {
		h.logger.Warn("Webhook not found",
			zap.String("webhook_id", webhookID),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "webhook not found",
			"webhook_id": webhookID,
		})
		return
	}

	// 2. Verificar se é um canal WAHA
	if !ch.IsWAHA() {
		h.logger.Error("Channel is not WAHA type",
			zap.String("webhook_id", webhookID),
			zap.String("channel_id", ch.ID.String()),
			zap.String("channel_type", string(ch.Type)))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "invalid channel type",
			"expected": "waha",
			"got":      string(ch.Type),
		})
		return
	}

	// 3. Extrair session ID da config do channel
	sessionID := ch.ExternalID // ExternalID contém o session_id do WAHA

	// 4. Ler o corpo da requisição
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 5. Extrair headers relevantes
	headers := map[string]string{
		"Content-Type":    c.GetHeader("Content-Type"),
		"User-Agent":      c.GetHeader("User-Agent"),
		"X-Forwarded-For": c.GetHeader("X-Forwarded-For"),
	}

	// 6. Criar evento raw com session ID
	rawEvent := waha.NewWAHARawEvent(
		sessionID, // Session ID extraído do channel
		body,
		headers,
	)

	// 7. Log do evento recebido
	h.logger.Info("WAHA webhook received",
		zap.String("event_id", rawEvent.ID),
		zap.String("webhook_id", webhookID),
		zap.String("channel_id", ch.ID.String()),
		zap.String("session", rawEvent.Session),
		zap.String("content_type", rawEvent.GetContentType()),
		zap.Int("body_size", rawEvent.GetBodySize()))

	// 8. Enfileirar evento raw (NUNCA falha)
	if err := h.rawEventBus.PublishRawEvent(c.Request.Context(), rawEvent); err != nil {
		// Log mas não falha - o evento pode ter sido perdido mas o webhook não quebra
		h.logger.Error("Failed to enqueue raw event",
			zap.Error(err),
			zap.String("event_id", rawEvent.ID))

		// Ainda assim responde sucesso para não quebrar o WAHA
		// O evento será perdido, mas é melhor que quebrar todo o fluxo
	}

	// 9. Resposta imediata - webhook nunca falha
	c.JSON(http.StatusOK, gin.H{
		"status":     "queued",
		"event_id":   rawEvent.ID,
		"webhook_id": webhookID,
		"channel_id": ch.ID.String(),
		"session":    rawEvent.Session,
		"message":    "Event queued for processing",
	})
}

// GetWebhookInfo provides information about the webhook endpoint (generic)
//
//	@Summary		Get webhook info
//	@Description	Retorna informações sobre o endpoint de webhook (padrão indústria)
//	@Tags			webhooks
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Webhook info"
//	@Router			/api/v1/webhooks/info [get]
func (h *WAHAWebhookHandler) GetWebhookInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"endpoint":         "/api/v1/webhooks/{webhook_id}",
		"method":           "POST",
		"content_type":     "application/json",
		"supported_events": waha.GetDefaultWebhookEvents(),
		"description":      "Endpoint para receber eventos de canais (WhatsApp, Telegram, etc.)",
		"pattern":          "Industry standard webhook pattern com ID único",
		"example_usage": map[string]interface{}{
			"url":        "http://localhost:8080/api/v1/webhooks/550e8400-e29b-41d4-a716-446655440000",
			"note":       "O webhook_id é gerado automaticamente ao criar um canal",
			"how_to_get": "Use o endpoint GET /api/v1/crm/channels/:id/webhook-url para obter a URL completa",
			"events": []string{
				"message",
				"message.ack",
				"call.received",
				"session.status",
			},
		},
	})
}
