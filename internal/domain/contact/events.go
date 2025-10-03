package contact

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para eventos de domínio.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// ContactCreatedEvent - Contato criado.
type ContactCreatedEvent struct {
	ContactID uuid.UUID
	ProjectID uuid.UUID
	TenantID  string
	Name      string
	CreatedAt time.Time
}

func (e ContactCreatedEvent) EventName() string     { return "contact.created" }
func (e ContactCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// ContactUpdatedEvent - Contato atualizado.
type ContactUpdatedEvent struct {
	ContactID uuid.UUID
	UpdatedAt time.Time
}

func (e ContactUpdatedEvent) EventName() string     { return "contact.updated" }
func (e ContactUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

// ContactDeletedEvent - Contato deletado (soft delete).
type ContactDeletedEvent struct {
	ContactID uuid.UUID
	DeletedAt time.Time
}

func (e ContactDeletedEvent) EventName() string     { return "contact.deleted" }
func (e ContactDeletedEvent) OccurredAt() time.Time { return e.DeletedAt }

// ContactMergedEvent - Contatos duplicados foram merged.
type ContactMergedEvent struct {
	PrimaryContactID   uuid.UUID
	MergedContactIDs   []uuid.UUID
	MergeStrategy      string
	MergedAt           time.Time
}

func (e ContactMergedEvent) EventName() string     { return "contact.merged" }
func (e ContactMergedEvent) OccurredAt() time.Time { return e.MergedAt }

// ContactEnrichedEvent - Dados externos adicionados ao contato.
type ContactEnrichedEvent struct {
	ContactID         uuid.UUID
	EnrichmentSource  string
	EnrichedData      map[string]interface{}
	EnrichedAt        time.Time
}

func (e ContactEnrichedEvent) EventName() string     { return "contact.enriched" }
func (e ContactEnrichedEvent) OccurredAt() time.Time { return e.EnrichedAt }
