package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Project é o Aggregate Root para workspaces/projetos (multi-tenancy).
type Project struct {
	id            uuid.UUID
	customerID    uuid.UUID
	tenantID      string
	name          string
	description   string
	configuration map[string]interface{}
	active        bool
	createdAt     time.Time
	updatedAt     time.Time
	
	events []DomainEvent
}

// NewProject cria um novo projeto.
func NewProject(customerID uuid.UUID, tenantID, name string) (*Project, error) {
	if customerID == uuid.Nil {
		return nil, errors.New("customerID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	project := &Project{
		id:            uuid.New(),
		customerID:    customerID,
		tenantID:      tenantID,
		name:          name,
		configuration: make(map[string]interface{}),
		active:        true,
		createdAt:     now,
		updatedAt:     now,
		events:        []DomainEvent{},
	}

	project.addEvent(ProjectCreatedEvent{
		ProjectID:  project.id,
		CustomerID: customerID,
		TenantID:   tenantID,
		Name:       name,
		CreatedAt:  now,
	})

	return project, nil
}

// ReconstructProject reconstrói um Project a partir de dados persistidos.
func ReconstructProject(
	id uuid.UUID,
	customerID uuid.UUID,
	tenantID string,
	name string,
	description string,
	configuration map[string]interface{},
	active bool,
	createdAt time.Time,
	updatedAt time.Time,
) *Project {
	if configuration == nil {
		configuration = make(map[string]interface{})
	}

	return &Project{
		id:            id,
		customerID:    customerID,
		tenantID:      tenantID,
		name:          name,
		description:   description,
		configuration: configuration,
		active:        active,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		events:        []DomainEvent{},
	}
}

// Activate ativa o projeto.
func (p *Project) Activate() {
	if !p.active {
		p.active = true
		p.updatedAt = time.Now()
	}
}

// Deactivate desativa o projeto.
func (p *Project) Deactivate() {
	if p.active {
		p.active = false
		p.updatedAt = time.Now()
	}
}

// UpdateConfiguration atualiza a configuração do projeto.
func (p *Project) UpdateConfiguration(config map[string]interface{}) {
	p.configuration = config
	p.updatedAt = time.Now()
}

// GetConfiguration retorna uma configuração específica.
func (p *Project) GetConfiguration(key string) (interface{}, bool) {
	val, ok := p.configuration[key]
	return val, ok
}

// SetSessionTimeout define o timeout padrão de sessões.
func (p *Project) SetSessionTimeout(minutes int) {
	if minutes <= 0 {
		minutes = 30
	}
	p.configuration["session_timeout_minutes"] = minutes
	p.updatedAt = time.Now()
}

// GetSessionTimeout retorna o timeout de sessões (default 30).
func (p *Project) GetSessionTimeout() int {
	if val, ok := p.configuration["session_timeout_minutes"]; ok {
		if timeout, ok := val.(int); ok {
			return timeout
		}
	}
	return 30
}

// Getters
func (p *Project) ID() uuid.UUID                    { return p.id }
func (p *Project) CustomerID() uuid.UUID            { return p.customerID }
func (p *Project) TenantID() string                 { return p.tenantID }
func (p *Project) Name() string                     { return p.name }
func (p *Project) Description() string              { return p.description }
func (p *Project) Configuration() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range p.configuration {
		copy[k] = v
	}
	return copy
}
func (p *Project) IsActive() bool       { return p.active }
func (p *Project) CreatedAt() time.Time { return p.createdAt }
func (p *Project) UpdatedAt() time.Time { return p.updatedAt }

// Domain Events
func (p *Project) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, p.events...)
}

func (p *Project) ClearEvents() {
	p.events = []DomainEvent{}
}

func (p *Project) addEvent(event DomainEvent) {
	p.events = append(p.events, event)
}
