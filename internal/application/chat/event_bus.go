package chat

import (
	"context"

	domainchat "github.com/caloi/ventros-crm/internal/domain/crm/chat"
)

type EventBus interface {
	Publish(ctx context.Context, event domainchat.DomainEvent) error
}
