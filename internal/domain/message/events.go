package message

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type MessageCreatedEvent struct {
	MessageID uuid.UUID
	ContactID uuid.UUID
	FromMe    bool
	CreatedAt time.Time
}

func (e MessageCreatedEvent) EventName() string     { return "message.created" }
func (e MessageCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type MessageDeliveredEvent struct {
	MessageID   uuid.UUID
	DeliveredAt time.Time
}

func (e MessageDeliveredEvent) EventName() string     { return "message.delivered" }
func (e MessageDeliveredEvent) OccurredAt() time.Time { return e.DeliveredAt }

type MessageReadEvent struct {
	MessageID uuid.UUID
	ReadAt    time.Time
}

func (e MessageReadEvent) EventName() string     { return "message.read" }
func (e MessageReadEvent) OccurredAt() time.Time { return e.ReadAt }

// MessageFailedEvent - Mensagem falhou ao ser enviada.
type MessageFailedEvent struct {
	MessageID   uuid.UUID
	FailureReason string
	FailedAt    time.Time
}

func (e MessageFailedEvent) EventName() string     { return "message.failed" }
func (e MessageFailedEvent) OccurredAt() time.Time { return e.FailedAt }
