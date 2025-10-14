package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ventros/crm/infrastructure/messaging"
	"go.uber.org/zap"
)

type QueueHandler struct {
	logger   *zap.Logger
	rabbitMQ *messaging.RabbitMQConnection
}

func NewQueueHandler(logger *zap.Logger, rabbitMQ *messaging.RabbitMQConnection) *QueueHandler {
	return &QueueHandler{
		logger:   logger,
		rabbitMQ: rabbitMQ,
	}
}

// ListQueues lista todas as filas RabbitMQ com suas estatísticas
//
//	@Summary		List RabbitMQ queues
//	@Description	Lista todas as filas do RabbitMQ com número de mensagens e consumers
//	@Tags			queues
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Queue statistics"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/queues [get]
func (h *QueueHandler) ListQueues(c *gin.Context) {
	queues, err := h.rabbitMQ.ListQueues()
	if err != nil {
		h.logger.Error("Failed to list queues", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list queues"})
		return
	}

	// Agrupa filas por tipo
	messageQueues := []messaging.QueueInfo{}
	dlqQueues := []messaging.QueueInfo{}
	totalMessages := 0
	totalConsumers := 0

	for _, queue := range queues {
		if queue.IsDLQ {
			dlqQueues = append(dlqQueues, queue)
		} else {
			messageQueues = append(messageQueues, queue)
		}
		totalMessages += queue.Messages
		totalConsumers += queue.Consumers
	}

	c.JSON(http.StatusOK, gin.H{
		"total_queues":       len(queues),
		"total_messages":     totalMessages,
		"total_consumers":    totalConsumers,
		"queues":             messageQueues,
		"dead_letter_queues": dlqQueues,
	})
}
