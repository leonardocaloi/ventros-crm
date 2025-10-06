package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/caloi/ventros-crm/infrastructure/webhooks"
)

// DomainEventBus publica eventos de domínio para RabbitMQ.
type DomainEventBus struct {
	conn            *RabbitMQConnection
	webhookNotifier *webhooks.WebhookNotifier
}

// NewDomainEventBus cria um novo event bus.
func NewDomainEventBus(conn *RabbitMQConnection, webhookNotifier *webhooks.WebhookNotifier) *DomainEventBus {
	return &DomainEventBus{
		conn:            conn,
		webhookNotifier: webhookNotifier,
	}
}

// Publish publica um evento de domínio no RabbitMQ.
// Serializa o evento como JSON e envia para uma exchange fanout.
func (bus *DomainEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
	// Serializa evento como JSON
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Nome da routing key baseado no evento
	routingKey := event.EventName()
	
	// Publica na exchange de eventos de domínio
	// Usa exchange fanout para broadcast
	err = bus.conn.Publish(ctx, fmt.Sprintf("domain.events.%s", routingKey), payload)
	if err != nil {
		return err
	}
	
	// Notifica webhooks inscritos neste evento de domínio
	if bus.webhookNotifier != nil {
		go bus.webhookNotifier.NotifyWebhooks(context.Background(), routingKey, event)
	}
	
	return nil
}

// PublishBatch publica múltiplos eventos em batch.
func (bus *DomainEventBus) PublishBatch(ctx context.Context, events []shared.DomainEvent) error {
	for _, event := range events {
		if err := bus.Publish(ctx, event); err != nil {
			// Log error but continue with other events
			fmt.Printf("Failed to publish event %s: %v\n", event.EventName(), err)
		}
		
		// Small delay to avoid overwhelming RabbitMQ
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// SetupEventQueues configura as queues necessárias para eventos de domínio.
func (bus *DomainEventBus) SetupEventQueues() error {
	// Queue para eventos de contato
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.contact.created", 3); err != nil {
		return err
	}
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.contact.updated", 3); err != nil {
		return err
	}
	
	// Queue para eventos de sessão
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.session.started", 3); err != nil {
		return err
	}
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.session.ended", 3); err != nil {
		return err
	}
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.message.recorded", 3); err != nil {
		return err
	}
	
	// Queue para eventos de mensagem
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.message.created", 3); err != nil {
		return err
	}
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.message.delivered", 3); err != nil {
		return err
	}
	
	// Queue para eventos de conversão de ads
	if err := bus.conn.DeclareQueueWithDLQ("domain.events.ad_conversion.tracked", 3); err != nil {
		return err
	}
	
	return nil
}
