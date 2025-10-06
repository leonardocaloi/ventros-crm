package project

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ProjectCreatedEvent struct {
	ProjectID        uuid.UUID
	CustomerID       uuid.UUID
	BillingAccountID uuid.UUID
	TenantID         string
	Name             string
	CreatedAt        time.Time
}

func (e ProjectCreatedEvent) EventName() string     { return "project.created" }
func (e ProjectCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }
