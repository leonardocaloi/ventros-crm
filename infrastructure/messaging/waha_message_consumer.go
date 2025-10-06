package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/application/message"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
)

// WAHAMessageConsumer consome eventos de mensagem do WAHA via RabbitMQ.
type WAHAMessageConsumer struct {
	wahaMessageService *message.WAHAMessageService
}

// NewWAHAMessageConsumer cria um novo consumer de mensagens WAHA.
func NewWAHAMessageConsumer(
	wahaMessageService *message.WAHAMessageService,
) *WAHAMessageConsumer {
	return &WAHAMessageConsumer{
		wahaMessageService: wahaMessageService,
	}
}

// ProcessMessage processa uma mensagem do RabbitMQ.
// Implementa a interface Consumer.
func (c *WAHAMessageConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	// Deserializa evento do WAHA
	var wahaEvent waha.WAHAMessageEvent
	if err := json.Unmarshal(delivery.Body, &wahaEvent); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	
	fmt.Printf("Processing WAHA event from queue: id=%s, event=%s, session=%s\n",
		wahaEvent.ID,
		wahaEvent.Event,
		wahaEvent.Session,
	)
	
	// Delega todo processamento para o service
	return c.wahaMessageService.ProcessWAHAMessage(ctx, wahaEvent)
}

// Start inicia o consumer.
func (c *WAHAMessageConsumer) Start(ctx context.Context, rabbitConn *RabbitMQConnection) error {
	queueName := "waha.events.message"
	consumerTag := fmt.Sprintf("ventros-crm-message-consumer-%s", uuid.New().String()[:8])
	
	fmt.Printf("Starting WAHA message consumer: queue=%s, tag=%s\n", queueName, consumerTag)
	
	return rabbitConn.StartConsumer(ctx, queueName, consumerTag, c, 10)
}
