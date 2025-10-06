package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	messageapp "github.com/caloi/ventros-crm/internal/application/message"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WAHAWebhookHandler struct {
	logger             *zap.Logger
	wahaMessageService *messageapp.WAHAMessageService
}

func NewWAHAWebhookHandler(
	logger *zap.Logger,
	wahaMessageService *messageapp.WAHAMessageService,
) *WAHAWebhookHandler {
	return &WAHAWebhookHandler{
		logger:             logger,
		wahaMessageService: wahaMessageService,
	}
}

// ReceiveWebhook receives WAHA webhook events
// @Summary Receive WAHA webhook
// @Description Recebe eventos de webhook do WAHA (mensagens, chamadas, etc.)
// @Tags webhooks
// @Accept json
// @Produce json
// @Param session query string false "Session ID"
// @Success 200 {object} map[string]interface{} "Event processed"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/webhooks/waha [post]
func (h *WAHAWebhookHandler) ReceiveWebhook(c *gin.Context) {
	// Ler o corpo da requisição
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Log do evento recebido
	h.logger.Info("WAHA webhook received",
		zap.String("session", c.Query("session")),
		zap.String("content_type", c.GetHeader("Content-Type")),
		zap.Int("body_size", len(body)))

	// Parsear o evento
	event, err := waha.ParseWebhookEvent(body)
	if err != nil {
		h.logger.Error("Failed to parse webhook event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook format"})
		return
	}

	// Log detalhado do evento
	h.logger.Info("WAHA event parsed",
		zap.String("event_type", event.Event),
		zap.String("session", event.Session),
		zap.Any("payload", event.Payload))

	// Processar baseado no tipo de evento
	switch event.Event {
	case "message", "message.any":
		if err := h.processMessageEvent(c, event); err != nil {
			h.logger.Error("Failed to process message event", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
			return
		}

	case "message.ack":
		h.logger.Debug("Message ACK received", zap.String("session", event.Session))
		// TODO: Atualizar status da mensagem

	case "call.received", "call.accepted", "call.rejected":
		h.logger.Info("Call event received",
			zap.String("event", event.Event),
			zap.String("session", event.Session))
		// TODO: Processar eventos de chamada

	case "label.upsert", "label.deleted", "label.chat.added", "label.chat.deleted":
		h.logger.Info("Label event received",
			zap.String("event", event.Event),
			zap.String("session", event.Session))
		// TODO: Processar eventos de labels/tags

	case "group.v2.join", "group.v2.leave", "group.v2.update", "group.v2.participants":
		h.logger.Info("Group event received",
			zap.String("event", event.Event),
			zap.String("session", event.Session))
		// TODO: Processar eventos de grupo

	default:
		h.logger.Warn("Unknown WAHA event type",
			zap.String("event", event.Event),
			zap.String("session", event.Session))
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "processed",
		"event":   event.Event,
		"session": event.Session,
	})
}

// processMessageEvent processa eventos de mensagem
// Agora delega toda a lógica para o WAHAMessageService
func (h *WAHAWebhookHandler) processMessageEvent(c *gin.Context, event *waha.WAHAWebhookEvent) error {
	// Converter payload para WAHAMessageEvent completo
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload waha.WAHAPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	wahaEvent := waha.WAHAMessageEvent{
		ID:        event.Event, // Usar event type como ID temporário
		Timestamp: 0,           // Será extraído do payload
		Event:     event.Event,
		Session:   event.Session,
		Payload:   payload,
	}

	// Delegar processamento completo para o service
	return h.wahaMessageService.ProcessWAHAMessage(c.Request.Context(), wahaEvent)
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
		"endpoint":        "/api/v1/webhooks/waha",
		"method":          "POST",
		"content_type":    "application/json",
		"supported_events": waha.GetDefaultWebhookEvents(),
		"description":     "Endpoint para receber eventos do WAHA (WhatsApp HTTP API)",
		"example_usage": map[string]interface{}{
			"url": "http://localhost:8080/api/v1/webhooks/waha",
			"events": []string{
				"message",
				"message.ack",
				"call.received",
			},
		},
	})
}
