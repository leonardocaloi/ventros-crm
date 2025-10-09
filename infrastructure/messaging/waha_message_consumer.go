package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/application/message"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// WAHAMessageConsumer consome eventos de mensagem do WAHA via RabbitMQ.
type WAHAMessageConsumer struct {
	wahaMessageService *message.WAHAMessageService
	idempotencyChecker IdempotencyChecker
	consumerName       string
}

// NewWAHAMessageConsumer cria um novo consumer de mensagens WAHA.
func NewWAHAMessageConsumer(
	wahaMessageService *message.WAHAMessageService,
	idempotencyChecker IdempotencyChecker,
) *WAHAMessageConsumer {
	return &WAHAMessageConsumer{
		wahaMessageService: wahaMessageService,
		idempotencyChecker: idempotencyChecker,
		consumerName:       "waha_message_consumer",
	}
}

// ProcessMessage processa uma mensagem do RabbitMQ.
// Implementa a interface Consumer.
func (c *WAHAMessageConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	startTime := time.Now()

	// Deserializa evento do WAHA
	var wahaEvent waha.WAHAMessageEvent
	if err := json.Unmarshal(delivery.Body, &wahaEvent); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	// Gera UUID a partir do ID do evento WAHA (usa hash estável)
	eventUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(wahaEvent.ID))

	// Verifica se já foi processado (idempotência)
	if c.idempotencyChecker != nil {
		processed, err := c.idempotencyChecker.IsProcessed(ctx, eventUUID, c.consumerName)
		if err != nil {
			fmt.Printf("⚠️  Failed to check idempotency: %v\n", err)
			// Continua processamento (fail-open)
		} else if processed {
			fmt.Printf("⏭️  Event already processed, skipping: id=%s\n", wahaEvent.ID)
			return nil // ACK sem reprocessar
		}
	}

	fmt.Printf("Processing WAHA event from queue: id=%s, event=%s, session=%s\n",
		wahaEvent.ID,
		wahaEvent.Event,
		wahaEvent.Session,
	)

	// Delega todo processamento para o service
	if err := c.wahaMessageService.ProcessWAHAMessage(ctx, wahaEvent); err != nil {
		return err
	}

	// Marca como processado após sucesso
	if c.idempotencyChecker != nil {
		duration := int(time.Since(startTime).Milliseconds())
		if err := c.idempotencyChecker.MarkAsProcessed(ctx, eventUUID, c.consumerName, &duration); err != nil {
			fmt.Printf("⚠️  Failed to mark event as processed: %v\n", err)
			// Não retorna erro - evento foi processado com sucesso
		}
	}

	return nil
}

// Start inicia o consumer.
func (c *WAHAMessageConsumer) Start(ctx context.Context, rabbitConn *RabbitMQConnection) error {
	queueName := "waha.events.message"
	consumerTag := fmt.Sprintf("ventros-crm-message-consumer-%s", uuid.New().String()[:8])

	fmt.Printf("Starting WAHA message consumer: queue=%s, tag=%s\n", queueName, consumerTag)

	return rabbitConn.StartConsumer(ctx, queueName, consumerTag, c, 10)
}
