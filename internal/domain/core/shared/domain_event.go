package shared

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string

	EventID() uuid.UUID

	EventVersion() string

	OccurredAt() time.Time
}

type BaseEvent struct {
	eventID      uuid.UUID
	eventName    string
	eventVersion string
	occurredAt   time.Time
}

func NewBaseEvent(eventName string, occurredAt time.Time) BaseEvent {
	return BaseEvent{
		eventID:      uuid.New(),
		eventName:    eventName,
		eventVersion: "v1",
		occurredAt:   occurredAt,
	}
}

func NewBaseEventWithVersion(eventName string, version string, occurredAt time.Time) BaseEvent {
	return BaseEvent{
		eventID:      uuid.New(),
		eventName:    eventName,
		eventVersion: version,
		occurredAt:   occurredAt,
	}
}

func (e BaseEvent) EventID() uuid.UUID    { return e.eventID }
func (e BaseEvent) EventName() string     { return e.eventName }
func (e BaseEvent) EventVersion() string  { return e.eventVersion }
func (e BaseEvent) OccurredAt() time.Time { return e.occurredAt }
