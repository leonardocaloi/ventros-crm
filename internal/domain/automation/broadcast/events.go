package broadcast

import (
	"time"

	"github.com/google/uuid"
)

// BroadcastCreatedEvent is emitted when a broadcast is created
type BroadcastCreatedEvent struct {
	BroadcastID uuid.UUID
	TenantID    string
	Name        string
	ListID      uuid.UUID
	Timestamp   time.Time
}

func (e *BroadcastCreatedEvent) EventType() string {
	return "automation.broadcast.created"
}

func (e *BroadcastCreatedEvent) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *BroadcastCreatedEvent) AggregateID() uuid.UUID {
	return e.BroadcastID
}

// BroadcastScheduledEvent is emitted when a broadcast is scheduled
type BroadcastScheduledEvent struct {
	BroadcastID  uuid.UUID
	ScheduledFor time.Time
	Timestamp    time.Time
}

func (e *BroadcastScheduledEvent) EventType() string {
	return "automation.broadcast.scheduled"
}

func (e *BroadcastScheduledEvent) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *BroadcastScheduledEvent) AggregateID() uuid.UUID {
	return e.BroadcastID
}

// BroadcastStartedEvent is emitted when a broadcast starts executing
type BroadcastStartedEvent struct {
	BroadcastID uuid.UUID
	Timestamp   time.Time
}

func (e *BroadcastStartedEvent) EventType() string {
	return "automation.broadcast.started"
}

func (e *BroadcastStartedEvent) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *BroadcastStartedEvent) AggregateID() uuid.UUID {
	return e.BroadcastID
}

// BroadcastCompletedEvent is emitted when a broadcast completes
type BroadcastCompletedEvent struct {
	BroadcastID uuid.UUID
	TotalSent   int
	TotalFailed int
	Timestamp   time.Time
}

func (e *BroadcastCompletedEvent) EventType() string {
	return "automation.broadcast.completed"
}

func (e *BroadcastCompletedEvent) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *BroadcastCompletedEvent) AggregateID() uuid.UUID {
	return e.BroadcastID
}

// BroadcastCancelledEvent is emitted when a broadcast is cancelled
type BroadcastCancelledEvent struct {
	BroadcastID uuid.UUID
	Timestamp   time.Time
}

func (e *BroadcastCancelledEvent) EventType() string {
	return "automation.broadcast.cancelled"
}

func (e *BroadcastCancelledEvent) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *BroadcastCancelledEvent) AggregateID() uuid.UUID {
	return e.BroadcastID
}

// BroadcastFailedEvent is emitted when a broadcast fails
type BroadcastFailedEvent struct {
	BroadcastID uuid.UUID
	Reason      string
	Timestamp   time.Time
}

func (e *BroadcastFailedEvent) EventType() string {
	return "automation.broadcast.failed"
}

func (e *BroadcastFailedEvent) EventTimestamp() time.Time {
	return e.Timestamp
}

func (e *BroadcastFailedEvent) AggregateID() uuid.UUID {
	return e.BroadcastID
}
