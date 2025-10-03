package messaging

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	domaincontact "github.com/caloi/ventros-crm/internal/domain/contact"
	domainsession "github.com/caloi/ventros-crm/internal/domain/session"
)

// ContactEventBusAdapter adapta DomainEventBus para contact.EventBus
type ContactEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewContactEventBusAdapter(domainEventBus *DomainEventBus) *ContactEventBusAdapter {
	return &ContactEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *ContactEventBusAdapter) Publish(ctx context.Context, event domaincontact.DomainEvent) error {
	// Cast para shared.DomainEvent
	sharedEvent, ok := event.(shared.DomainEvent)
	if !ok {
		// Se não implementa shared.DomainEvent, criar um wrapper
		sharedEvent = &DomainEventWrapper{
			eventName:  event.EventName(),
			occurredAt: event.OccurredAt(),
			data:       event,
		}
	}
	return a.domainEventBus.Publish(ctx, sharedEvent)
}

// SessionEventBusAdapter adapta DomainEventBus para session.EventBus
type SessionEventBusAdapter struct {
	domainEventBus *DomainEventBus
}

func NewSessionEventBusAdapter(domainEventBus *DomainEventBus) *SessionEventBusAdapter {
	return &SessionEventBusAdapter{domainEventBus: domainEventBus}
}

func (a *SessionEventBusAdapter) Publish(ctx context.Context, event domainsession.DomainEvent) error {
	// Cast para shared.DomainEvent
	sharedEvent, ok := event.(shared.DomainEvent)
	if !ok {
		// Se não implementa shared.DomainEvent, criar um wrapper
		sharedEvent = &DomainEventWrapper{
			eventName:  event.EventName(),
			occurredAt: event.OccurredAt(),
			data:       event,
		}
	}
	return a.domainEventBus.Publish(ctx, sharedEvent)
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
