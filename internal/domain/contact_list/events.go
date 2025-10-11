package contact_list

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ContactListCreatedEvent struct {
	ContactListID uuid.UUID
	ProjectID     uuid.UUID
	TenantID      string
	Name          string
	IsStatic      bool
	CreatedAt     time.Time
}

func (e ContactListCreatedEvent) EventName() string     { return "contact_list.created" }
func (e ContactListCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type ContactListUpdatedEvent struct {
	ContactListID uuid.UUID
	UpdatedFields []string
	UpdatedAt     time.Time
}

func (e ContactListUpdatedEvent) EventName() string     { return "contact_list.updated" }
func (e ContactListUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

type ContactListDeletedEvent struct {
	ContactListID uuid.UUID
	DeletedAt     time.Time
}

func (e ContactListDeletedEvent) EventName() string     { return "contact_list.deleted" }
func (e ContactListDeletedEvent) OccurredAt() time.Time { return e.DeletedAt }

type ContactListFilterRuleAddedEvent struct {
	ContactListID uuid.UUID
	FilterRuleID  uuid.UUID
	FilterType    string
	AddedAt       time.Time
}

func (e ContactListFilterRuleAddedEvent) EventName() string     { return "contact_list.filter_rule_added" }
func (e ContactListFilterRuleAddedEvent) OccurredAt() time.Time { return e.AddedAt }

type ContactListFilterRuleRemovedEvent struct {
	ContactListID uuid.UUID
	FilterRuleID  uuid.UUID
	RemovedAt     time.Time
}

func (e ContactListFilterRuleRemovedEvent) EventName() string {
	return "contact_list.filter_rule_removed"
}
func (e ContactListFilterRuleRemovedEvent) OccurredAt() time.Time { return e.RemovedAt }

type ContactListFilterRulesClearedEvent struct {
	ContactListID uuid.UUID
	ClearedAt     time.Time
}

func (e ContactListFilterRulesClearedEvent) EventName() string {
	return "contact_list.filter_rules_cleared"
}
func (e ContactListFilterRulesClearedEvent) OccurredAt() time.Time { return e.ClearedAt }

type ContactListRecalculatedEvent struct {
	ContactListID uuid.UUID
	ContactCount  int
	CalculatedAt  time.Time
}

func (e ContactListRecalculatedEvent) EventName() string     { return "contact_list.recalculated" }
func (e ContactListRecalculatedEvent) OccurredAt() time.Time { return e.CalculatedAt }

type ContactAddedToListEvent struct {
	ContactListID uuid.UUID
	ContactID     uuid.UUID
	AddedAt       time.Time
}

func (e ContactAddedToListEvent) EventName() string     { return "contact_list.contact_added" }
func (e ContactAddedToListEvent) OccurredAt() time.Time { return e.AddedAt }

type ContactRemovedFromListEvent struct {
	ContactListID uuid.UUID
	ContactID     uuid.UUID
	RemovedAt     time.Time
}

func (e ContactRemovedFromListEvent) EventName() string     { return "contact_list.contact_removed" }
func (e ContactRemovedFromListEvent) OccurredAt() time.Time { return e.RemovedAt }
