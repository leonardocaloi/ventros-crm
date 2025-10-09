package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// Webhook queue names
	WebhookQueueName    = "webhooks.outbound"
	WebhookQueueDLQName = "webhooks.outbound.dlq"

	// Retry configuration
	WebhookMaxRetries = 3
)

// WebhookQueueMessage represents a webhook message in the queue
type WebhookQueueMessage struct {
	WebhookID   string                 `json:"webhook_id"`
	URL         string                 `json:"url"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers"`
	Payload     map[string]interface{} `json:"payload"`
	MaxRetries  int                    `json:"max_retries"`
	TimeoutSecs int                    `json:"timeout_secs"`
	EventType   string                 `json:"event_type"`
	EventID     string                 `json:"event_id"`
	TenantID    string                 `json:"tenant_id"`
	EnqueuedAt  time.Time              `json:"enqueued_at"`
}

// WebhookQueuePublisher publishes webhook delivery tasks to RabbitMQ
type WebhookQueuePublisher struct {
	conn *RabbitMQConnection
}

// NewWebhookQueuePublisher creates a new webhook queue publisher
func NewWebhookQueuePublisher(conn *RabbitMQConnection) (*WebhookQueuePublisher, error) {
	publisher := &WebhookQueuePublisher{
		conn: conn,
	}

	// Setup webhook queues
	if err := publisher.setupQueues(); err != nil {
		return nil, fmt.Errorf("failed to setup webhook queues: %w", err)
	}

	return publisher, nil
}

// setupQueues declares webhook queues with DLQ
func (p *WebhookQueuePublisher) setupQueues() error {
	return p.conn.DeclareQueueWithDLQ(WebhookQueueName, WebhookMaxRetries)
}

// PublishWebhook publishes a webhook delivery task to the queue
func (p *WebhookQueuePublisher) PublishWebhook(ctx context.Context, msg WebhookQueueMessage) error {
	// Set enqueued timestamp
	msg.EnqueuedAt = time.Now().UTC()

	// Marshal message to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook message: %w", err)
	}

	// Publish to queue
	if err := p.conn.Publish(ctx, WebhookQueueName, body); err != nil {
		return fmt.Errorf("failed to publish webhook to queue: %w", err)
	}

	return nil
}

// WebhookQueueConsumer consumes and processes webhook delivery tasks
type WebhookQueueConsumer struct {
	conn      *RabbitMQConnection
	processor WebhookProcessor
}

// WebhookProcessor interface for processing webhook deliveries
type WebhookProcessor interface {
	ProcessWebhook(ctx context.Context, msg WebhookQueueMessage) error
}

// NewWebhookQueueConsumer creates a new webhook queue consumer
func NewWebhookQueueConsumer(conn *RabbitMQConnection, processor WebhookProcessor) (*WebhookQueueConsumer, error) {
	consumer := &WebhookQueueConsumer{
		conn:      conn,
		processor: processor,
	}

	// Setup webhook queues
	if err := conn.DeclareQueueWithDLQ(WebhookQueueName, WebhookMaxRetries); err != nil {
		return nil, fmt.Errorf("failed to setup webhook queues: %w", err)
	}

	return consumer, nil
}

// Start starts consuming webhook messages
func (c *WebhookQueueConsumer) Start(ctx context.Context) error {
	return c.conn.StartConsumer(
		ctx,
		WebhookQueueName,
		"webhook-consumer",
		c,
		10, // prefetch count
	)
}

// ProcessMessage implements Consumer interface
func (c *WebhookQueueConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
	// Parse message
	var msg WebhookQueueMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		fmt.Printf("Failed to unmarshal webhook message: %v\n", err)
		return fmt.Errorf("failed to unmarshal webhook message: %w", err)
	}

	fmt.Printf("Processing webhook delivery: webhook_id=%s, url=%s, event_type=%s\n",
		msg.WebhookID, msg.URL, msg.EventType)

	// Process webhook
	if err := c.processor.ProcessWebhook(ctx, msg); err != nil {
		fmt.Printf("Failed to process webhook: webhook_id=%s, error=%v\n", msg.WebhookID, err)
		return err
	}

	fmt.Printf("Webhook delivered successfully: webhook_id=%s\n", msg.WebhookID)
	return nil
}

// WebhookQueueStats returns statistics about webhook queues
type WebhookQueueStats struct {
	QueueName       string `json:"queue_name"`
	PendingMessages int    `json:"pending_messages"`
	ActiveConsumers int    `json:"active_consumers"`
	DLQMessages     int    `json:"dlq_messages"`
	ProcessingRate  string `json:"processing_rate"`
	AverageLatency  string `json:"average_latency"`
	SuccessRate     string `json:"success_rate"`
	LastProcessedAt string `json:"last_processed_at"`
}

// GetStats returns webhook queue statistics
func (p *WebhookQueuePublisher) GetStats() (*WebhookQueueStats, error) {
	channel := p.conn.Channel()
	if channel == nil {
		return nil, fmt.Errorf("channel is not available")
	}

	// Get main queue info
	mainQueue, err := channel.QueueInspect(WebhookQueueName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect main queue: %w", err)
	}

	// Get DLQ info
	dlqQueue, err := channel.QueueInspect(WebhookQueueDLQName)
	dlqMessages := 0
	if err == nil {
		dlqMessages = dlqQueue.Messages
	}

	stats := &WebhookQueueStats{
		QueueName:       WebhookQueueName,
		PendingMessages: mainQueue.Messages,
		ActiveConsumers: mainQueue.Consumers,
		DLQMessages:     dlqMessages,
		ProcessingRate:  "N/A", // Would need metrics system
		AverageLatency:  "N/A", // Would need metrics system
		SuccessRate:     "N/A", // Would need metrics system
		LastProcessedAt: "N/A", // Would need metrics system
	}

	return stats, nil
}
