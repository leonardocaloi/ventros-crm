package messaging

import (
	"context"

	"github.com/ventros/crm/internal/domain/core/shared"
	domaincontact "github.com/ventros/crm/internal/domain/crm/contact"
	domainsession "github.com/ventros/crm/internal/domain/crm/session"
)

// ContactEventBusAdapter adapta DomainEventBus para contact.EventBus
type ContactEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewContactEventBusAdapter(domainEventBus *DomainEventBus) *ContactEventBusAdapter {
	return &ContactEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *ContactEventBusAdapter) Publish(ctx context.Context, event domaincontact.DomainEvent) error {
	// contact.DomainEvent é agora um alias para shared.DomainEvent, então pode passar direto
	return a.domainEventBus.Publish(ctx, event)
}

// SessionEventBusAdapter adapta DomainEventBus para session.EventBus
type SessionEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewSessionEventBusAdapter(domainEventBus *DomainEventBus) *SessionEventBusAdapter {
	return &SessionEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *SessionEventBusAdapter) Publish(ctx context.Context, event domainsession.DomainEvent) error {
	// session.DomainEvent é agora um alias para shared.DomainEvent, então pode passar direto
	return a.domainEventBus.Publish(ctx, event)
}

// MessageEventBusAdapter adapta DomainEventBus para message.EventBus
type MessageEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewMessageEventBusAdapter(domainEventBus *DomainEventBus) *MessageEventBusAdapter {
	return &MessageEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *MessageEventBusAdapter) Publish(ctx context.Context, event shared.DomainEvent) error {
	return a.domainEventBus.Publish(ctx, event)
}

func (a *MessageEventBusAdapter) PublishBatch(ctx context.Context, events []shared.DomainEvent) error {
	return a.domainEventBus.PublishBatch(ctx, events)
}

// ChatEventBusAdapter adapta DomainEventBus para chat.EventBus
type ChatEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewChatEventBusAdapter(domainEventBus *DomainEventBus) *ChatEventBusAdapter {
	return &ChatEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *ChatEventBusAdapter) Publish(ctx context.Context, event shared.DomainEvent) error {
	// chat.DomainEvent é agora compatível com shared.DomainEvent, então pode passar direto
	return a.domainEventBus.Publish(ctx, event)
}

// SagaEventBusAdapter adapta DomainEventBus para saga.EventBus (interface{} based)
type SagaEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewSagaEventBusAdapter(domainEventBus *DomainEventBus) *SagaEventBusAdapter {
	return &SagaEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *SagaEventBusAdapter) Publish(ctx context.Context, event interface{}) error {
	// Convert interface{} to shared.DomainEvent
	if domainEvent, ok := event.(shared.DomainEvent); ok {
		return a.domainEventBus.Publish(ctx, domainEvent)
	}
	return nil // Silently ignore non-DomainEvent events
}

func (a *SagaEventBusAdapter) PublishBatch(ctx context.Context, events []interface{}) error {
	// Convert []interface{} to []shared.DomainEvent
	domainEvents := make([]shared.DomainEvent, 0, len(events))
	for _, event := range events {
		if domainEvent, ok := event.(shared.DomainEvent); ok {
			domainEvents = append(domainEvents, domainEvent)
		}
	}
	return a.domainEventBus.PublishBatch(ctx, domainEvents)
}
