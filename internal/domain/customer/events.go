package customer

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type CustomerCreatedEvent struct {
	CustomerID uuid.UUID
	Name       string
	Email      string
	CreatedAt  time.Time
}

func (e CustomerCreatedEvent) EventName() string     { return "customer.created" }
func (e CustomerCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type CustomerActivatedEvent struct {
	CustomerID  uuid.UUID
	ActivatedAt time.Time
}

func (e CustomerActivatedEvent) EventName() string     { return "customer.activated" }
func (e CustomerActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

type CustomerSuspendedEvent struct {
	CustomerID  uuid.UUID
	SuspendedAt time.Time
}

func (e CustomerSuspendedEvent) EventName() string     { return "customer.suspended" }
func (e CustomerSuspendedEvent) OccurredAt() time.Time { return e.SuspendedAt }
