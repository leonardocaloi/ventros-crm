package session

import (
	"context"

	domainsession "github.com/caloi/ventros-crm/internal/domain/session"
)

// EventBus é a interface para publicar eventos de domínio.
// Implementação pode usar RabbitMQ, Kafka, etc.
type EventBus interface {
	Publish(ctx context.Context, event domainsession.DomainEvent) error
}
