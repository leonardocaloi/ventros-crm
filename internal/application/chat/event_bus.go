package chat

import (
	"context"

	"github.com/ventros/crm/internal/domain/core/shared"
)

type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
}
