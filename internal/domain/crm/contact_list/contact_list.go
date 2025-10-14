package contact_list

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type LogicalOperator string

const (
	LogicalOperatorAND LogicalOperator = "AND"
	LogicalOperatorOR  LogicalOperator = "OR"
)

func (lo LogicalOperator) IsValid() bool {
	return lo == LogicalOperatorAND || lo == LogicalOperatorOR
}

type ContactList struct {
	id               uuid.UUID
	version          int // Optimistic locking - prevents lost updates
	projectID        uuid.UUID
	tenantID         string
	name             string
	description      *string
	filterRules      []*FilterRule
	logicalOperator  LogicalOperator
	isStatic         bool
	contactCount     int
	lastCalculatedAt *time.Time
	createdAt        time.Time
	updatedAt        time.Time
	deletedAt        *time.Time
	events           []shared.DomainEvent
}

func NewContactList(
	projectID uuid.UUID,
	tenantID string,
	name string,
	logicalOperator LogicalOperator,
	isStatic bool,
) (*ContactList, error) {
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if !logicalOperator.IsValid() {
		return nil, errors.New("invalid logical operator")
	}

	now := time.Now()
	list := &ContactList{
		id:              uuid.New(),
		version:         1, // Start with version 1 for new aggregates
		projectID:       projectID,
		tenantID:        tenantID,
		name:            name,
		filterRules:     []*FilterRule{},
		logicalOperator: logicalOperator,
		isStatic:        isStatic,
		contactCount:    0,
		createdAt:       now,
		updatedAt:       now,
		events:          []shared.DomainEvent{},
	}

	list.addEvent(NewContactListCreatedEvent(list.id, projectID, tenantID, name, isStatic))

	return list, nil
}

func ReconstructContactList(
	id uuid.UUID,
	version int, // Optimistic locking version
	projectID uuid.UUID,
	tenantID string,
	name string,
	description *string,
	filterRules []*FilterRule,
	logicalOperator LogicalOperator,
	isStatic bool,
	contactCount int,
	lastCalculatedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *ContactList {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}
	if filterRules == nil {
		filterRules = []*FilterRule{}
	}

	return &ContactList{
		id:               id,
		version:          version,
		projectID:        projectID,
		tenantID:         tenantID,
		name:             name,
		description:      description,
		filterRules:      filterRules,
		logicalOperator:  logicalOperator,
		isStatic:         isStatic,
		contactCount:     contactCount,
		lastCalculatedAt: lastCalculatedAt,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		deletedAt:        deletedAt,
		events:           []shared.DomainEvent{},
	}
}

func (cl *ContactList) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	cl.name = name
	cl.updatedAt = time.Now()

	cl.addEvent(NewContactListUpdatedEvent(cl.id, []string{"name"}))

	return nil
}

func (cl *ContactList) UpdateDescription(description string) {
	cl.description = &description
	cl.updatedAt = time.Now()

	cl.addEvent(NewContactListUpdatedEvent(cl.id, []string{"description"}))
}

func (cl *ContactList) AddFilterRule(rule *FilterRule) error {
	if rule == nil {
		return errors.New("filter rule cannot be nil")
	}

	cl.filterRules = append(cl.filterRules, rule)
	cl.updatedAt = time.Now()

	cl.addEvent(NewContactListFilterRuleAddedEvent(cl.id, rule.ID(), string(rule.FilterType())))

	return nil
}

func (cl *ContactList) RemoveFilterRule(ruleID uuid.UUID) error {
	for i, rule := range cl.filterRules {
		if rule.ID() == ruleID {
			cl.filterRules = append(cl.filterRules[:i], cl.filterRules[i+1:]...)
			cl.updatedAt = time.Now()

			cl.addEvent(NewContactListFilterRuleRemovedEvent(cl.id, ruleID))

			return nil
		}
	}
	return errors.New("filter rule not found")
}

func (cl *ContactList) ClearFilterRules() {
	cl.filterRules = []*FilterRule{}
	cl.updatedAt = time.Now()

	cl.addEvent(NewContactListFilterRulesClearedEvent(cl.id))
}

func (cl *ContactList) UpdateLogicalOperator(operator LogicalOperator) error {
	if !operator.IsValid() {
		return errors.New("invalid logical operator")
	}

	cl.logicalOperator = operator
	cl.updatedAt = time.Now()

	cl.addEvent(NewContactListUpdatedEvent(cl.id, []string{"logical_operator"}))

	return nil
}

func (cl *ContactList) UpdateContactCount(count int) {
	cl.contactCount = count
	now := time.Now()
	cl.lastCalculatedAt = &now
	cl.updatedAt = now

	cl.addEvent(NewContactListRecalculatedEvent(cl.id, count))
}

func (cl *ContactList) Delete() {
	now := time.Now()
	cl.deletedAt = &now
	cl.updatedAt = now

	cl.addEvent(NewContactListDeletedEvent(cl.id))
}

func (cl *ContactList) IsDeleted() bool {
	return cl.deletedAt != nil
}

func (cl *ContactList) HasFilterRules() bool {
	return len(cl.filterRules) > 0
}

func (cl *ContactList) addEvent(event shared.DomainEvent) {
	cl.events = append(cl.events, event)
}

func (cl *ContactList) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, cl.events...)
}

func (cl *ContactList) ClearEvents() {
	cl.events = []shared.DomainEvent{}
}

func (cl *ContactList) ID() uuid.UUID                    { return cl.id }
func (cl *ContactList) Version() int                     { return cl.version }
func (cl *ContactList) ProjectID() uuid.UUID             { return cl.projectID }
func (cl *ContactList) TenantID() string                 { return cl.tenantID }
func (cl *ContactList) Name() string                     { return cl.name }
func (cl *ContactList) Description() *string             { return cl.description }
func (cl *ContactList) FilterRules() []*FilterRule       { return cl.filterRules }
func (cl *ContactList) LogicalOperator() LogicalOperator { return cl.logicalOperator }
func (cl *ContactList) IsStatic() bool                   { return cl.isStatic }
func (cl *ContactList) ContactCount() int                { return cl.contactCount }
func (cl *ContactList) LastCalculatedAt() *time.Time     { return cl.lastCalculatedAt }
func (cl *ContactList) CreatedAt() time.Time             { return cl.createdAt }
func (cl *ContactList) UpdatedAt() time.Time             { return cl.updatedAt }
func (cl *ContactList) DeletedAt() *time.Time            { return cl.deletedAt }

// Compile-time check that ContactList implements AggregateRoot interface
var _ shared.AggregateRoot = (*ContactList)(nil)
