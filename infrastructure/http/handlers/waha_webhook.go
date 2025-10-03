package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/infrastructure/webhooks"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WAHAWebhookHandler struct {
	logger          *zap.Logger
	rabbitMQ        *messaging.RabbitMQConnection
	webhookNotifier *webhooks.WebhookNotifier
}

func NewWAHAWebhookHandler(logger *zap.Logger, rabbitMQ *messaging.RabbitMQConnection, webhookNotifier *webhooks.WebhookNotifier) *WAHAWebhookHandler {
	return &WAHAWebhookHandler{
		logger:          logger,
		rabbitMQ:        rabbitMQ,
		webhookNotifier: webhookNotifier,
	}
}

// HandleMessage recebe eventos da WAHA e roteia para filas específicas
// @Summary Receive WAHA webhook
// @Description Endpoint único para todos os eventos WAHA (message, ack, session, etc)
// @Tags webhooks
// @Accept json
// @Produce json
// @Param payload body waha.WAHAMessageEvent true "WAHA Event"
// @Success 200 {object} map[string]interface{} "Event queued successfully"
// @Failure 400 {object} map[string]interface{} "Invalid payload"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /webhooks/waha/events [post]
func (h *WAHAWebhookHandler) HandleMessage(c *gin.Context) {
	start := time.Now()
	
	var event waha.WAHAMessageEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("Failed to parse WAHA payload",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Determina fila baseado no tipo de evento e direção da mensagem
	queue := h.routeEventToQueue(event.Event, event.Payload.FromMe)
	if queue == "" {
		h.logger.Warn("Unknown WAHA event type, using default queue",
			zap.String("event_type", event.Event),
			zap.String("event_id", event.ID),
		)
		queue = "waha.events.unknown"
	}

	// Log evento recebido
	h.logger.Info("WAHA webhook received",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Event),
		zap.String("queue", queue),
		zap.String("session", event.Session),
		zap.String("from", event.Payload.From),
		zap.Bool("from_me", event.Payload.FromMe),
	)

	// Serializa evento para JSON
	payload, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal event",
			zap.Error(err),
			zap.String("event_id", event.ID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serialize event"})
		return
	}

	// Publica na fila específica
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	err = h.rabbitMQ.Publish(ctx, queue, payload)
	if err != nil {
		h.logger.Error("Failed to publish to RabbitMQ",
			zap.Error(err),
			zap.String("event_id", event.ID),
			zap.String("queue", queue),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue event"})
		return
	}

	// Webhook notifications are handled by domain events, not WAHA events

	duration := time.Since(start)
	
	h.logger.Info("WAHA event queued successfully",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Event),
		zap.String("queue", queue),
		zap.Duration("duration", duration),
	)

	c.JSON(http.StatusOK, gin.H{
		"status":      "queued",
		"event_id":    event.ID,
		"event_type":  event.Event,
		"queue":       queue,
		"duration_ms": duration.Milliseconds(),
	})
}

// routeEventToQueue determina qual fila usar baseado no tipo de evento WAHA e direção da mensagem
func (h *WAHAWebhookHandler) routeEventToQueue(eventType string, fromMe bool) string {
	switch eventType {
	// Mensagens - Roteamento inteligente baseado na direção
	case "message":
		if fromMe {
			// Mensagens enviadas (outgoing) - para auditoria/backup completo
			return "waha.events.message.any"
		} else {
			// Mensagens recebidas (incoming) - para processamento com buffer
			return "waha.events.message"
		}
	case "message.any":
		// Todas as mensagens independente da direção
		return "waha.events.message.any"
	
	// Confirmações de leitura (ACKs) - Roteamento baseado na direção
	case "message.ack":
		if fromMe {
			// ACKs de mensagens enviadas - para auditoria/backup completo
			return "waha.events.message.any"
		} else {
			// ACKs de mensagens recebidas - para processamento específico
			return "waha.events.ack"
		}
	
	// Status de sessão/conexão
	case "session.status":
		return "waha.events.session.status"
	case "state.change":
		return "waha.events.session.status"
	
	// Presença (online/offline)
	case "presence.update":
		return "waha.events.presence"
	
	// Grupos
	case "group.join":
		return "waha.events.group"
	case "group.leave":
		return "waha.events.group"
	case "group.update":
		return "waha.events.group"
	
	// Chamadas
	case "call.received":
		return "waha.events.call"
	case "call.rejected":
		return "waha.events.call"
	
	// Labels/Tags
	case "label.upsert":
		return "waha.events.label"
	case "label.deleted":
		return "waha.events.label"
	
	default:
		return ""
	}
}

// HandleStatus recebe status de mensagens enviadas (ACK)
// @Summary Receive WAHA message status
// @Description Endpoint para receber confirmações de entrega da WAHA
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /webhooks/waha/status [post]
func (h *WAHAWebhookHandler) HandleStatus(c *gin.Context) {
	var status map[string]interface{}
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	h.logger.Info("WAHA status update received",
		zap.Any("status", status),
	)

	// TODO: Processar status de entrega/leitura
	
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Health check do webhook
// @Summary Webhook health check
// @Description Verifica se o endpoint de webhook está funcionando
// @Tags webhooks
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /webhooks/waha/health [get]
func (h *WAHAWebhookHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "waha_webhook",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
