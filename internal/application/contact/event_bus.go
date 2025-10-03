package contact

import (
	"context"

	domaincontact "github.com/caloi/ventros-crm/internal/domain/contact"
)

// EventBus é a interface para publicar eventos de domínio.
type EventBus interface {
	Publish(ctx context.Context, event domaincontact.DomainEvent) error
}
