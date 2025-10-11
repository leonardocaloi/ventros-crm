package webhook

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, webhook *WebhookSubscription) error

	FindByID(ctx context.Context, id uuid.UUID) (*WebhookSubscription, error)

	FindAll(ctx context.Context) ([]*WebhookSubscription, error)

	FindActiveByEvent(ctx context.Context, eventType string) ([]*WebhookSubscription, error)

	FindByActive(ctx context.Context, active bool) ([]*WebhookSubscription, error)

	Update(ctx context.Context, webhook *WebhookSubscription) error

	Delete(ctx context.Context, id uuid.UUID) error

	RecordTrigger(ctx context.Context, id uuid.UUID, success bool) error
}
