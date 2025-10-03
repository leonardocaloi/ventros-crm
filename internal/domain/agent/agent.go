package agent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Agent é o Aggregate Root para agentes do sistema.
// Representa usuários que interagem com contatos e gerenciam conversas.
type Agent struct {
	id          uuid.UUID
	tenantID    string
	name        string
	email       string
	role        Role
	active      bool
	permissions map[string]bool
	settings    map[string]interface{}
	createdAt   time.Time
	updatedAt   time.Time
	lastLoginAt *time.Time

	// Domain Events
	events []DomainEvent
}

// NewAgent cria um novo agente (factory method).
func NewAgent(
	tenantID string,
	name string,
	email string,
	role Role,
) (*Agent, error) {
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if !role.IsValid() {
		return nil, errors.New("invalid role")
	}

	now := time.Now()
	agent := &Agent{
		id:          uuid.New(),
		tenantID:    tenantID,
		name:        name,
		email:       email,
		role:        role,
		active:      true,
		permissions: make(map[string]bool),
		settings:    make(map[string]interface{}),
		createdAt:   now,
		updatedAt:   now,
		events:      []DomainEvent{},
	}

	agent.addEvent(AgentCreatedEvent{
		AgentID:   agent.id,
		TenantID:  tenantID,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: now,
	})

	return agent, nil
}

// ReconstructAgent reconstrói um agente a partir de dados persistidos.
func ReconstructAgent(
	id uuid.UUID,
	tenantID string,
	name string,
	email string,
	role Role,
	active bool,
	permissions map[string]bool,
	settings map[string]interface{},
	createdAt time.Time,
	updatedAt time.Time,
	lastLoginAt *time.Time,
) *Agent {
	return &Agent{
		id:          id,
		tenantID:    tenantID,
		name:        name,
		email:       email,
		role:        role,
		active:      active,
		permissions: permissions,
		settings:    settings,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		lastLoginAt: lastLoginAt,
		events:      []DomainEvent{},
	}
}

// UpdateProfile atualiza informações básicas do agente.
func (a *Agent) UpdateProfile(name, email string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if email == "" {
		return errors.New("email cannot be empty")
	}

	changes := make(map[string]interface{})
	if a.name != name {
		changes["name"] = map[string]string{"old": a.name, "new": name}
		a.name = name
	}
	if a.email != email {
		changes["email"] = map[string]string{"old": a.email, "new": email}
		a.email = email
	}

	if len(changes) > 0 {
		a.updatedAt = time.Now()
		a.addEvent(AgentUpdatedEvent{
			AgentID:   a.id,
			Changes:   changes,
			UpdatedAt: a.updatedAt,
		})
	}

	return nil
}

// Activate ativa o agente.
func (a *Agent) Activate() error {
	if a.active {
		return errors.New("agent is already active")
	}

	a.active = true
	a.updatedAt = time.Now()

	a.addEvent(AgentActivatedEvent{
		AgentID:     a.id,
		ActivatedAt: a.updatedAt,
	})

	return nil
}

// Deactivate desativa o agente.
func (a *Agent) Deactivate() error {
	if !a.active {
		return errors.New("agent is already inactive")
	}

	a.active = false
	a.updatedAt = time.Now()

	a.addEvent(AgentDeactivatedEvent{
		AgentID:       a.id,
		DeactivatedAt: a.updatedAt,
	})

	return nil
}

// RecordLogin registra o último login do agente.
func (a *Agent) RecordLogin() {
	now := time.Now()
	a.lastLoginAt = &now
	a.updatedAt = now

	a.addEvent(AgentLoggedInEvent{
		AgentID:     a.id,
		LoggedInAt:  now,
	})
}

// GrantPermission concede uma permissão ao agente.
func (a *Agent) GrantPermission(permission string) error {
	if permission == "" {
		return errors.New("permission cannot be empty")
	}

	if a.permissions[permission] {
		return nil // já tem a permissão
	}

	a.permissions[permission] = true
	a.updatedAt = time.Now()

	a.addEvent(AgentPermissionGrantedEvent{
		AgentID:    a.id,
		Permission: permission,
		GrantedAt:  a.updatedAt,
	})

	return nil
}

// RevokePermission revoga uma permissão do agente.
func (a *Agent) RevokePermission(permission string) error {
	if permission == "" {
		return errors.New("permission cannot be empty")
	}

	if !a.permissions[permission] {
		return nil // já não tem a permissão
	}

	delete(a.permissions, permission)
	a.updatedAt = time.Now()

	a.addEvent(AgentPermissionRevokedEvent{
		AgentID:    a.id,
		Permission: permission,
		RevokedAt:  a.updatedAt,
	})

	return nil
}

// HasPermission verifica se o agente tem uma permissão.
func (a *Agent) HasPermission(permission string) bool {
	return a.permissions[permission]
}

// UpdateSettings atualiza configurações do agente.
func (a *Agent) UpdateSettings(settings map[string]interface{}) {
	a.settings = settings
	a.updatedAt = time.Now()
}

// Getters
func (a *Agent) ID() uuid.UUID                    { return a.id }
func (a *Agent) TenantID() string                 { return a.tenantID }
func (a *Agent) Name() string                     { return a.name }
func (a *Agent) Email() string                    { return a.email }
func (a *Agent) Role() Role                       { return a.role }
func (a *Agent) IsActive() bool                   { return a.active }
func (a *Agent) Permissions() map[string]bool     { return a.permissions }
func (a *Agent) Settings() map[string]interface{} { return a.settings }
func (a *Agent) CreatedAt() time.Time             { return a.createdAt }
func (a *Agent) UpdatedAt() time.Time             { return a.updatedAt }
func (a *Agent) LastLoginAt() *time.Time          { return a.lastLoginAt }

// Domain Events
func (a *Agent) DomainEvents() []DomainEvent {
	return a.events
}

func (a *Agent) ClearEvents() {
	a.events = []DomainEvent{}
}

func (a *Agent) addEvent(event DomainEvent) {
	a.events = append(a.events, event)
}
