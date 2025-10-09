package contact

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// DomainEvent é um alias para shared.DomainEvent (compatibilidade retroativa).
type DomainEvent = shared.DomainEvent

// ContactCreatedEvent - Contato criado.
type ContactCreatedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	ProjectID uuid.UUID
	TenantID  string
	Name      string
	CreatedAt time.Time
}

func NewContactCreatedEvent(contactID, projectID uuid.UUID, tenantID, name string) ContactCreatedEvent {
	return ContactCreatedEvent{
		BaseEvent: shared.NewBaseEvent("contact.created", time.Now()),
		ContactID: contactID,
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// ContactUpdatedEvent - Contato atualizado.
type ContactUpdatedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	UpdatedAt time.Time
}

func NewContactUpdatedEvent(contactID uuid.UUID) ContactUpdatedEvent {
	return ContactUpdatedEvent{
		BaseEvent: shared.NewBaseEvent("contact.updated", time.Now()),
		ContactID: contactID,
		UpdatedAt: time.Now(),
	}
}

// ContactProfilePictureUpdatedEvent - Foto de perfil do contato foi atualizada
type ContactProfilePictureUpdatedEvent struct {
	shared.BaseEvent
	ContactID         uuid.UUID
	TenantID          string
	ProfilePictureURL string
	FetchedAt         time.Time
}

func NewContactProfilePictureUpdatedEvent(contactID uuid.UUID, tenantID, profilePictureURL string) ContactProfilePictureUpdatedEvent {
	return ContactProfilePictureUpdatedEvent{
		BaseEvent:         shared.NewBaseEvent("contact.profile_picture_updated", time.Now()),
		ContactID:         contactID,
		TenantID:          tenantID,
		ProfilePictureURL: profilePictureURL,
		FetchedAt:         time.Now(),
	}
}

// ContactDeletedEvent - Contato deletado (soft delete).
type ContactDeletedEvent struct {
	shared.BaseEvent
	ContactID uuid.UUID
	DeletedAt time.Time
}

func NewContactDeletedEvent(contactID uuid.UUID) ContactDeletedEvent {
	return ContactDeletedEvent{
		BaseEvent: shared.NewBaseEvent("contact.deleted", time.Now()),
		ContactID: contactID,
		DeletedAt: time.Now(),
	}
}

// ContactMergedEvent - Contatos duplicados foram merged.
type ContactMergedEvent struct {
	shared.BaseEvent
	PrimaryContactID uuid.UUID
	MergedContactIDs []uuid.UUID
	MergeStrategy    string
	MergedAt         time.Time
}

func NewContactMergedEvent(primaryContactID uuid.UUID, mergedContactIDs []uuid.UUID, mergeStrategy string) ContactMergedEvent {
	return ContactMergedEvent{
		BaseEvent:        shared.NewBaseEvent("contact.merged", time.Now()),
		PrimaryContactID: primaryContactID,
		MergedContactIDs: mergedContactIDs,
		MergeStrategy:    mergeStrategy,
		MergedAt:         time.Now(),
	}
}

// ContactEnrichedEvent - Dados externos adicionados ao contato.
type ContactEnrichedEvent struct {
	shared.BaseEvent
	ContactID        uuid.UUID
	EnrichmentSource string
	EnrichedData     map[string]interface{}
	EnrichedAt       time.Time
}

func NewContactEnrichedEvent(contactID uuid.UUID, enrichmentSource string, enrichedData map[string]interface{}) ContactEnrichedEvent {
	return ContactEnrichedEvent{
		BaseEvent:        shared.NewBaseEvent("contact.enriched", time.Now()),
		ContactID:        contactID,
		EnrichmentSource: enrichmentSource,
		EnrichedData:     enrichedData,
		EnrichedAt:       time.Now(),
	}
}
