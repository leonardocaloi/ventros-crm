package contact_list

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

type ContactListCreatedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	ProjectID     uuid.UUID
	TenantID      string
	Name          string
	IsStatic      bool
}

func NewContactListCreatedEvent(contactListID, projectID uuid.UUID, tenantID, name string, isStatic bool) ContactListCreatedEvent {
	return ContactListCreatedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.created", time.Now()),
		ContactListID: contactListID,
		ProjectID:     projectID,
		TenantID:      tenantID,
		Name:          name,
		IsStatic:      isStatic,
	}
}

type ContactListUpdatedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	UpdatedFields []string
}

func NewContactListUpdatedEvent(contactListID uuid.UUID, updatedFields []string) ContactListUpdatedEvent {
	return ContactListUpdatedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.updated", time.Now()),
		ContactListID: contactListID,
		UpdatedFields: updatedFields,
	}
}

type ContactListDeletedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
}

func NewContactListDeletedEvent(contactListID uuid.UUID) ContactListDeletedEvent {
	return ContactListDeletedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.deleted", time.Now()),
		ContactListID: contactListID,
	}
}

type ContactListFilterRuleAddedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	FilterRuleID  uuid.UUID
	FilterType    string
}

func NewContactListFilterRuleAddedEvent(contactListID, filterRuleID uuid.UUID, filterType string) ContactListFilterRuleAddedEvent {
	return ContactListFilterRuleAddedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.filter_rule_added", time.Now()),
		ContactListID: contactListID,
		FilterRuleID:  filterRuleID,
		FilterType:    filterType,
	}
}

type ContactListFilterRuleRemovedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	FilterRuleID  uuid.UUID
}

func NewContactListFilterRuleRemovedEvent(contactListID, filterRuleID uuid.UUID) ContactListFilterRuleRemovedEvent {
	return ContactListFilterRuleRemovedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.filter_rule_removed", time.Now()),
		ContactListID: contactListID,
		FilterRuleID:  filterRuleID,
	}
}

type ContactListFilterRulesClearedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
}

func NewContactListFilterRulesClearedEvent(contactListID uuid.UUID) ContactListFilterRulesClearedEvent {
	return ContactListFilterRulesClearedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.filter_rules_cleared", time.Now()),
		ContactListID: contactListID,
	}
}

type ContactListRecalculatedEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	ContactCount  int
}

func NewContactListRecalculatedEvent(contactListID uuid.UUID, contactCount int) ContactListRecalculatedEvent {
	return ContactListRecalculatedEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.recalculated", time.Now()),
		ContactListID: contactListID,
		ContactCount:  contactCount,
	}
}

type ContactAddedToListEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	ContactID     uuid.UUID
}

func NewContactAddedToListEvent(contactListID, contactID uuid.UUID) ContactAddedToListEvent {
	return ContactAddedToListEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.contact_added", time.Now()),
		ContactListID: contactListID,
		ContactID:     contactID,
	}
}

type ContactRemovedFromListEvent struct {
	shared.BaseEvent
	ContactListID uuid.UUID
	ContactID     uuid.UUID
}

func NewContactRemovedFromListEvent(contactListID, contactID uuid.UUID) ContactRemovedFromListEvent {
	return ContactRemovedFromListEvent{
		BaseEvent:     shared.NewBaseEvent("contact_list.contact_removed", time.Now()),
		ContactListID: contactListID,
		ContactID:     contactID,
	}
}
