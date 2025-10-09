package messaging

import (
	"context"

	domaincontact "github.com/caloi/ventros-crm/internal/domain/contact"
	domainsession "github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/caloi/ventros-crm/internal/domain/shared"
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
