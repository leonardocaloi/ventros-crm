package project

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

type DomainEvent = shared.DomainEvent

type ProjectCreatedEvent struct {
	shared.BaseEvent
	ProjectID        uuid.UUID
	CustomerID       uuid.UUID
	BillingAccountID uuid.UUID
	TenantID         string
	Name             string
}

func NewProjectCreatedEvent(projectID, customerID, billingAccountID uuid.UUID, tenantID, name string) ProjectCreatedEvent {
	return ProjectCreatedEvent{
		BaseEvent:        shared.NewBaseEvent("project.created", time.Now()),
		ProjectID:        projectID,
		CustomerID:       customerID,
		BillingAccountID: billingAccountID,
		TenantID:         tenantID,
		Name:             name,
	}
}
