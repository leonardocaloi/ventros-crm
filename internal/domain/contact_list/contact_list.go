package contact_list

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// LogicalOperator define o operador lógico entre regras de filtro
type LogicalOperator string

const (
	LogicalOperatorAND LogicalOperator = "AND"
	LogicalOperatorOR  LogicalOperator = "OR"
)

// IsValid verifica se o operador lógico é válido
func (lo LogicalOperator) IsValid() bool {
	return lo == LogicalOperatorAND || lo == LogicalOperatorOR
}

// ContactList é o Aggregate Root para listas de contatos
type ContactList struct {
	id               uuid.UUID
	projectID        uuid.UUID
	tenantID         string
	name             string
	description      *string
	filterRules      []*FilterRule
	logicalOperator  LogicalOperator // AND ou OR entre as regras
	isStatic         bool            // Se true, a lista é estática (não atualiza automaticamente)
	contactCount     int             // Número de contatos na lista (cache)
	lastCalculatedAt *time.Time      // Última vez que a lista foi recalculada
	createdAt        time.Time
	updatedAt        time.Time
	deletedAt        *time.Time
	events           []DomainEvent
}

// NewContactList cria uma nova lista de contatos
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
		projectID:       projectID,
		tenantID:        tenantID,
		name:            name,
		filterRules:     []*FilterRule{},
		logicalOperator: logicalOperator,
		isStatic:        isStatic,
		contactCount:    0,
		createdAt:       now,
		updatedAt:       now,
		events:          []DomainEvent{},
	}

	list.addEvent(ContactListCreatedEvent{
		ContactListID: list.id,
		ProjectID:     projectID,
		TenantID:      tenantID,
		Name:          name,
		IsStatic:      isStatic,
		CreatedAt:     now,
	})

	return list, nil
}

// ReconstructContactList reconstrói uma lista de contatos a partir de dados persistidos
func ReconstructContactList(
	id uuid.UUID,
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
	if filterRules == nil {
		filterRules = []*FilterRule{}
	}

	return &ContactList{
		id:               id,
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
		events:           []DomainEvent{},
	}
}

// UpdateName atualiza o nome da lista
func (cl *ContactList) UpdateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	cl.name = name
	cl.updatedAt = time.Now()

	cl.addEvent(ContactListUpdatedEvent{
		ContactListID: cl.id,
		UpdatedFields: []string{"name"},
		UpdatedAt:     cl.updatedAt,
	})

	return nil
}

// UpdateDescription atualiza a descrição da lista
func (cl *ContactList) UpdateDescription(description string) {
	cl.description = &description
	cl.updatedAt = time.Now()

	cl.addEvent(ContactListUpdatedEvent{
		ContactListID: cl.id,
		UpdatedFields: []string{"description"},
		UpdatedAt:     cl.updatedAt,
	})
}

// AddFilterRule adiciona uma nova regra de filtro
func (cl *ContactList) AddFilterRule(rule *FilterRule) error {
	if rule == nil {
		return errors.New("filter rule cannot be nil")
	}

	cl.filterRules = append(cl.filterRules, rule)
	cl.updatedAt = time.Now()

	cl.addEvent(ContactListFilterRuleAddedEvent{
		ContactListID: cl.id,
		FilterRuleID:  rule.ID(),
		FilterType:    string(rule.FilterType()),
		AddedAt:       cl.updatedAt,
	})

	return nil
}

// RemoveFilterRule remove uma regra de filtro pelo ID
func (cl *ContactList) RemoveFilterRule(ruleID uuid.UUID) error {
	for i, rule := range cl.filterRules {
		if rule.ID() == ruleID {
			cl.filterRules = append(cl.filterRules[:i], cl.filterRules[i+1:]...)
			cl.updatedAt = time.Now()

			cl.addEvent(ContactListFilterRuleRemovedEvent{
				ContactListID: cl.id,
				FilterRuleID:  ruleID,
				RemovedAt:     cl.updatedAt,
			})

			return nil
		}
	}
	return errors.New("filter rule not found")
}

// ClearFilterRules remove todas as regras de filtro
func (cl *ContactList) ClearFilterRules() {
	cl.filterRules = []*FilterRule{}
	cl.updatedAt = time.Now()

	cl.addEvent(ContactListFilterRulesClearedEvent{
		ContactListID: cl.id,
		ClearedAt:     cl.updatedAt,
	})
}

// UpdateLogicalOperator atualiza o operador lógico entre as regras
func (cl *ContactList) UpdateLogicalOperator(operator LogicalOperator) error {
	if !operator.IsValid() {
		return errors.New("invalid logical operator")
	}

	cl.logicalOperator = operator
	cl.updatedAt = time.Now()

	cl.addEvent(ContactListUpdatedEvent{
		ContactListID: cl.id,
		UpdatedFields: []string{"logical_operator"},
		UpdatedAt:     cl.updatedAt,
	})

	return nil
}

// UpdateContactCount atualiza o contador de contatos
func (cl *ContactList) UpdateContactCount(count int) {
	cl.contactCount = count
	now := time.Now()
	cl.lastCalculatedAt = &now
	cl.updatedAt = now

	cl.addEvent(ContactListRecalculatedEvent{
		ContactListID: cl.id,
		ContactCount:  count,
		CalculatedAt:  now,
	})
}

// Delete marca a lista como deletada (soft delete)
func (cl *ContactList) Delete() {
	now := time.Now()
	cl.deletedAt = &now
	cl.updatedAt = now

	cl.addEvent(ContactListDeletedEvent{
		ContactListID: cl.id,
		DeletedAt:     now,
	})
}

// IsDeleted verifica se a lista está deletada
func (cl *ContactList) IsDeleted() bool {
	return cl.deletedAt != nil
}

// HasFilterRules verifica se a lista possui regras de filtro
func (cl *ContactList) HasFilterRules() bool {
	return len(cl.filterRules) > 0
}

// addEvent adiciona um evento de domínio
func (cl *ContactList) addEvent(event DomainEvent) {
	cl.events = append(cl.events, event)
}

// GetEvents retorna todos os eventos de domínio
func (cl *ContactList) GetEvents() []DomainEvent {
	return cl.events
}

// ClearEvents limpa todos os eventos de domínio
func (cl *ContactList) ClearEvents() {
	cl.events = []DomainEvent{}
}

// Getters
func (cl *ContactList) ID() uuid.UUID                    { return cl.id }
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
