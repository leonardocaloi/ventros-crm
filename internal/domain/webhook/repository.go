package webhook

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para persistência de webhook subscriptions
type Repository interface {
	// Create cria uma nova inscrição de webhook
	Create(ctx context.Context, webhook *WebhookSubscription) error

	// FindByID busca um webhook por ID
	FindByID(ctx context.Context, id uuid.UUID) (*WebhookSubscription, error)

	// FindAll busca todos os webhooks
	FindAll(ctx context.Context) ([]*WebhookSubscription, error)

	// FindActiveByEvent busca webhooks ativos inscritos em um evento
	FindActiveByEvent(ctx context.Context, eventType string) ([]*WebhookSubscription, error)

	// FindByActive busca webhooks por status ativo
	FindByActive(ctx context.Context, active bool) ([]*WebhookSubscription, error)

	// Update atualiza um webhook existente
	Update(ctx context.Context, webhook *WebhookSubscription) error

	// Delete remove um webhook
	Delete(ctx context.Context, id uuid.UUID) error

	// RecordTrigger atualiza estatísticas de disparo do webhook
	RecordTrigger(ctx context.Context, id uuid.UUID, success bool) error
}
