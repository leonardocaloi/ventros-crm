package tracking

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// DomainEvent é um alias para shared.DomainEvent (compatibilidade retroativa).
type DomainEvent = shared.DomainEvent

// TrackingCreatedEvent - Tracking criado no sistema.
type TrackingCreatedEvent struct {
	shared.BaseEvent
	TrackingID uuid.UUID
	ContactID  uuid.UUID
	SessionID  *uuid.UUID
	TenantID   string
	ProjectID  uuid.UUID
	Source     string // meta_ads, google_ads, etc
	Platform   string // instagram, facebook, etc
	CreatedAt  time.Time
}

// NewTrackingCreatedEvent cria um novo evento de tracking criado
func NewTrackingCreatedEvent(trackingID, contactID, projectID uuid.UUID, sessionID *uuid.UUID, tenantID, source, platform string) TrackingCreatedEvent {
	return TrackingCreatedEvent{
		BaseEvent:  shared.NewBaseEvent("tracking.created", time.Now()),
		TrackingID: trackingID,
		ContactID:  contactID,
		SessionID:  sessionID,
		TenantID:   tenantID,
		ProjectID:  projectID,
		Source:     source,
		Platform:   platform,
		CreatedAt:  time.Now(),
	}
}

// TrackingEnrichedEvent - Tracking enriquecido com dados adicionais.
type TrackingEnrichedEvent struct {
	shared.BaseEvent
	TrackingID uuid.UUID
	ContactID  uuid.UUID
	Changes    map[string]interface{}
	EnrichedAt time.Time
}

// NewTrackingEnrichedEvent cria um novo evento de tracking enriquecido
func NewTrackingEnrichedEvent(trackingID, contactID uuid.UUID, changes map[string]interface{}) TrackingEnrichedEvent {
	return TrackingEnrichedEvent{
		BaseEvent:  shared.NewBaseEvent("tracking.enriched", time.Now()),
		TrackingID: trackingID,
		ContactID:  contactID,
		Changes:    changes,
		EnrichedAt: time.Now(),
	}
}
