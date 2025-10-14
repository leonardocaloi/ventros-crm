package session

import (
	"context"

	domainsession "github.com/ventros/crm/internal/domain/crm/session"
)

// EventBus é a interface para publicar eventos de domínio.
// Implementação pode usar RabbitMQ, Kafka, etc.
type EventBus interface {
	Publish(ctx context.Context, event domainsession.DomainEvent) error
}
