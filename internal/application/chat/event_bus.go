package chat

import (
	"context"

	domainchat "github.com/caloi/ventros-crm/internal/domain/chat"
)

type EventBus interface {
	Publish(ctx context.Context, event domainchat.DomainEvent) error
}
