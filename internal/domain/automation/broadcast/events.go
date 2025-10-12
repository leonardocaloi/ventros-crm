package broadcast

import (
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

// BroadcastCreatedEvent is emitted when a broadcast is created
type BroadcastCreatedEvent struct {
	shared.BaseEvent
	BroadcastID uuid.UUID
	TenantID    string
	Name        string
	ListID      uuid.UUID
}

func NewBroadcastCreatedEvent(broadcastID uuid.UUID, tenantID, name string, listID uuid.UUID) BroadcastCreatedEvent {
	return BroadcastCreatedEvent{
		BaseEvent:   shared.NewBaseEvent("broadcast.created", time.Now()),
		BroadcastID: broadcastID,
		TenantID:    tenantID,
		Name:        name,
		ListID:      listID,
	}
}

// BroadcastScheduledEvent is emitted when a broadcast is scheduled
type BroadcastScheduledEvent struct {
	shared.BaseEvent
	BroadcastID  uuid.UUID
	ScheduledFor time.Time
}

func NewBroadcastScheduledEvent(broadcastID uuid.UUID, scheduledFor time.Time) BroadcastScheduledEvent {
	return BroadcastScheduledEvent{
		BaseEvent:    shared.NewBaseEvent("broadcast.scheduled", time.Now()),
		BroadcastID:  broadcastID,
		ScheduledFor: scheduledFor,
	}
}

// BroadcastStartedEvent is emitted when a broadcast starts executing
type BroadcastStartedEvent struct {
	shared.BaseEvent
	BroadcastID uuid.UUID
}

func NewBroadcastStartedEvent(broadcastID uuid.UUID) BroadcastStartedEvent {
	return BroadcastStartedEvent{
		BaseEvent:   shared.NewBaseEvent("broadcast.started", time.Now()),
		BroadcastID: broadcastID,
	}
}

// BroadcastCompletedEvent is emitted when a broadcast completes
type BroadcastCompletedEvent struct {
	shared.BaseEvent
	BroadcastID uuid.UUID
	TotalSent   int
	TotalFailed int
}

func NewBroadcastCompletedEvent(broadcastID uuid.UUID, totalSent, totalFailed int) BroadcastCompletedEvent {
	return BroadcastCompletedEvent{
		BaseEvent:   shared.NewBaseEvent("broadcast.completed", time.Now()),
		BroadcastID: broadcastID,
		TotalSent:   totalSent,
		TotalFailed: totalFailed,
	}
}

// BroadcastCancelledEvent is emitted when a broadcast is cancelled
type BroadcastCancelledEvent struct {
	shared.BaseEvent
	BroadcastID uuid.UUID
}

func NewBroadcastCancelledEvent(broadcastID uuid.UUID) BroadcastCancelledEvent {
	return BroadcastCancelledEvent{
		BaseEvent:   shared.NewBaseEvent("broadcast.cancelled", time.Now()),
		BroadcastID: broadcastID,
	}
}

// BroadcastFailedEvent is emitted when a broadcast fails
type BroadcastFailedEvent struct {
	shared.BaseEvent
	BroadcastID uuid.UUID
	Reason      string
}

func NewBroadcastFailedEvent(broadcastID uuid.UUID, reason string) BroadcastFailedEvent {
	return BroadcastFailedEvent{
		BaseEvent:   shared.NewBaseEvent("broadcast.failed", time.Now()),
		BroadcastID: broadcastID,
		Reason:      reason,
	}
}
