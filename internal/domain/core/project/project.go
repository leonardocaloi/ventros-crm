package project

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	id                    uuid.UUID
	version               int // Optimistic locking - prevents lost updates
	customerID            uuid.UUID
	billingAccountID      uuid.UUID
	tenantID              string
	name                  string
	description           string
	configuration         map[string]interface{}
	active                bool
	sessionTimeoutMinutes int
	agentAssignment       *AgentAssignmentConfig
	createdAt             time.Time
	updatedAt             time.Time

	events []shared.DomainEvent
}

func NewProject(customerID, billingAccountID uuid.UUID, tenantID, name string) (*Project, error) {
	if customerID == uuid.Nil {
		return nil, errors.New("customerID cannot be nil")
	}
	if billingAccountID == uuid.Nil {
		return nil, errors.New("billingAccountID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	project := &Project{
		id:                    uuid.New(),
		version:               1, // Start with version 1 for new aggregates
		customerID:            customerID,
		billingAccountID:      billingAccountID,
		tenantID:              tenantID,
		name:                  name,
		configuration:         make(map[string]interface{}),
		active:                true,
		sessionTimeoutMinutes: 30,
		agentAssignment:       NewAgentAssignmentConfig(),
		createdAt:             now,
		updatedAt:             now,
		events:                []shared.DomainEvent{},
	}

	project.addEvent(NewProjectCreatedEvent(project.id, customerID, billingAccountID, tenantID, name))

	return project, nil
}

func ReconstructProject(
	id uuid.UUID,
	version int, // Optimistic locking version
	customerID uuid.UUID,
	billingAccountID uuid.UUID,
	tenantID string,
	name string,
	description string,
	configuration map[string]interface{},
	active bool,
	sessionTimeoutMinutes int,
	agentAssignment *AgentAssignmentConfig,
	createdAt time.Time,
	updatedAt time.Time,
) *Project {
	if version == 0 {
		version = 1 // Default to version 1 (backwards compatibility)
	}
	if configuration == nil {
		configuration = make(map[string]interface{})
	}

	if sessionTimeoutMinutes <= 0 {
		sessionTimeoutMinutes = 30
	}

	if agentAssignment == nil {
		agentAssignment = NewAgentAssignmentConfig()
	}

	return &Project{
		id:                    id,
		version:               version,
		customerID:            customerID,
		billingAccountID:      billingAccountID,
		tenantID:              tenantID,
		name:                  name,
		description:           description,
		configuration:         configuration,
		active:                active,
		sessionTimeoutMinutes: sessionTimeoutMinutes,
		agentAssignment:       agentAssignment,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
		events:                []shared.DomainEvent{},
	}
}

func (p *Project) Activate() {
	if !p.active {
		p.active = true
		p.updatedAt = time.Now()
	}
}

func (p *Project) Deactivate() {
	if p.active {
		p.active = false
		p.updatedAt = time.Now()
	}
}

func (p *Project) UpdateConfiguration(config map[string]interface{}) {
	p.configuration = config
	p.updatedAt = time.Now()
}

func (p *Project) UpdateDescription(description string) {
	p.description = description
	p.updatedAt = time.Now()
}

func (p *Project) GetConfiguration(key string) (interface{}, bool) {
	val, ok := p.configuration[key]
	return val, ok
}

func (p *Project) SetSessionTimeout(minutes int) {
	if minutes <= 0 {
		minutes = 30
	}
	p.sessionTimeoutMinutes = minutes
	p.updatedAt = time.Now()
}

func (p *Project) GetSessionTimeout() int {
	return p.sessionTimeoutMinutes
}

// Agent Assignment Methods

func (p *Project) GetAgentAssignment() *AgentAssignmentConfig {
	if p.agentAssignment == nil {
		p.agentAssignment = NewAgentAssignmentConfig()
	}
	return p.agentAssignment
}

func (p *Project) SetAgentAssignmentConfig(config *AgentAssignmentConfig) error {
	if config == nil {
		return errors.New("agent assignment config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid agent assignment config: %w", err)
	}

	p.agentAssignment = config
	p.updatedAt = time.Now()

	return nil
}

func (p *Project) AddAssignmentAgent(agentID uuid.UUID) error {
	config := p.GetAgentAssignment()
	if err := config.AddAgent(agentID); err != nil {
		return err
	}
	p.updatedAt = time.Now()
	return nil
}

func (p *Project) RemoveAssignmentAgent(agentID uuid.UUID) error {
	config := p.GetAgentAssignment()
	if err := config.RemoveAgent(agentID); err != nil {
		return err
	}
	p.updatedAt = time.Now()
	return nil
}

func (p *Project) EnableAgentAssignment() {
	config := p.GetAgentAssignment()
	config.Enable()
	p.updatedAt = time.Now()
}

func (p *Project) DisableAgentAssignment() {
	config := p.GetAgentAssignment()
	config.Disable()
	p.updatedAt = time.Now()
}

func (p *Project) SetAssignmentStrategy(strategy AssignmentStrategy) error {
	config := p.GetAgentAssignment()
	if err := config.SetStrategy(strategy); err != nil {
		return err
	}
	p.updatedAt = time.Now()
	return nil
}

// GetNextAssignmentAgent retorna o próximo agente baseado na estratégia configurada
func (p *Project) GetNextAssignmentAgent() (uuid.UUID, error) {
	config := p.GetAgentAssignment()

	if !config.ShouldAutoAssign() {
		return uuid.Nil, errors.New("auto-assignment is not enabled")
	}

	switch config.Strategy {
	case StrategyRoundRobin:
		return config.GetNextAgentRoundRobin()
	case StrategyLeastSessions:
		// Para least_sessions, precisamos de informação externa
		// Por enquanto, retorna o primeiro
		if len(config.AgentIDs) > 0 {
			return config.AgentIDs[0], nil
		}
		return uuid.Nil, errors.New("no agents configured")
	case StrategyManual:
		return uuid.Nil, errors.New("manual assignment strategy - no auto-assignment")
	default:
		return uuid.Nil, fmt.Errorf("unknown assignment strategy: %s", config.Strategy)
	}
}

// ===== Reassignment Rule Management =====

// AddReassignmentRule adiciona uma regra de reatribuição automática ao projeto
func (p *Project) AddReassignmentRule(rule *ReassignmentRule) error {
	config := p.GetAgentAssignment()
	if err := config.AddReassignmentRule(rule); err != nil {
		return err
	}
	p.updatedAt = time.Now()
	return nil
}

// RemoveReassignmentRule remove uma regra de reatribuição por índice
func (p *Project) RemoveReassignmentRule(index int) error {
	config := p.GetAgentAssignment()
	if err := config.RemoveReassignmentRule(index); err != nil {
		return err
	}
	p.updatedAt = time.Now()
	return nil
}

// GetActiveReassignmentRules retorna as regras de reatribuição ativas
func (p *Project) GetActiveReassignmentRules() []*ReassignmentRule {
	config := p.GetAgentAssignment()
	return config.GetActiveReassignmentRules()
}

// ShouldReassignSession verifica se uma sessão deve ser reatribuída
func (p *Project) ShouldReassignSession(lastActivityAt time.Time, reassignmentCount int) (bool, *ReassignmentRule) {
	config := p.GetAgentAssignment()
	return config.ShouldReassignSession(lastActivityAt, reassignmentCount)
}

// ClearReassignmentRules remove todas as regras de reatribuição
func (p *Project) ClearReassignmentRules() {
	config := p.GetAgentAssignment()
	config.ClearReassignmentRules()
	p.updatedAt = time.Now()
}

// SetRequireAtLeastOneAgent define se pelo menos um agente é obrigatório
func (p *Project) SetRequireAtLeastOneAgent(required bool) {
	config := p.GetAgentAssignment()
	config.RequireAtLeastOneAgent = required
	p.updatedAt = time.Now()
}

// RequiresAtLeastOneAgent retorna se o projeto exige pelo menos um agente
func (p *Project) RequiresAtLeastOneAgent() bool {
	config := p.GetAgentAssignment()
	return config.RequireAtLeastOneAgent
}

func (p *Project) ID() uuid.UUID               { return p.id }
func (p *Project) Version() int                { return p.version }
func (p *Project) CustomerID() uuid.UUID       { return p.customerID }
func (p *Project) BillingAccountID() uuid.UUID { return p.billingAccountID }
func (p *Project) TenantID() string            { return p.tenantID }
func (p *Project) Name() string                { return p.name }
func (p *Project) Description() string         { return p.description }
func (p *Project) Configuration() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range p.configuration {
		copy[k] = v
	}
	return copy
}
func (p *Project) IsActive() bool             { return p.active }
func (p *Project) SessionTimeoutMinutes() int { return p.sessionTimeoutMinutes }
func (p *Project) CreatedAt() time.Time       { return p.createdAt }
func (p *Project) UpdatedAt() time.Time       { return p.updatedAt }

func (p *Project) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, p.events...)
}

func (p *Project) ClearEvents() {
	p.events = []shared.DomainEvent{}
}

func (p *Project) addEvent(event shared.DomainEvent) {
	p.events = append(p.events, event)
}

// Compile-time check that Project implements AggregateRoot interface
var _ shared.AggregateRoot = (*Project)(nil)
