package tracking

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

type DomainEvent = shared.DomainEvent

type TrackingCreatedEvent struct {
	shared.BaseEvent
	TrackingID uuid.UUID
	ContactID  uuid.UUID
	SessionID  *uuid.UUID
	TenantID   string
	ProjectID  uuid.UUID
	Source     string
	Platform   string
	CreatedAt  time.Time
}

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

type TrackingEnrichedEvent struct {
	shared.BaseEvent
	TrackingID uuid.UUID
	ContactID  uuid.UUID
	Changes    map[string]interface{}
	EnrichedAt time.Time
}

func NewTrackingEnrichedEvent(trackingID, contactID uuid.UUID, changes map[string]interface{}) TrackingEnrichedEvent {
	return TrackingEnrichedEvent{
		BaseEvent:  shared.NewBaseEvent("tracking.enriched", time.Now()),
		TrackingID: trackingID,
		ContactID:  contactID,
		Changes:    changes,
		EnrichedAt: time.Now(),
	}
}
