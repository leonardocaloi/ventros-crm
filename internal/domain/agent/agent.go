package agent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AgentType define os tipos de agentes
type AgentType string

const (
	AgentTypeHuman   AgentType = "human"   // Agente humano (atendente/admin)
	AgentTypeAI      AgentType = "ai"      // Agente de IA (externo via provider)
	AgentTypeBot     AgentType = "bot"     // Bot/automação (interno)
	AgentTypeChannel AgentType = "channel" // Canal/dispositivo
)

// AgentStatus define os status possíveis
type AgentStatus string

const (
	AgentStatusAvailable AgentStatus = "available"
	AgentStatusBusy      AgentStatus = "busy"
	AgentStatusAway      AgentStatus = "away"
	AgentStatusOffline   AgentStatus = "offline"
)

// Agent é o Aggregate Root para agentes do sistema.
// Representa entidades que podem interagir com contatos: humanos, IAs, bots ou canais.
type Agent struct {
	id          uuid.UUID
	projectID   uuid.UUID
	userID      *uuid.UUID // Null para agentes não-humanos
	tenantID    string
	name        string
	email       string
	agentType   AgentType
	status      AgentStatus
	role        Role
	active      bool
	config      map[string]interface{} // Configurações específicas (ex: AI provider, model)
	permissions map[string]bool
	settings    map[string]interface{}
	
	// Métricas
	sessionsHandled   int
	averageResponseMs int
	lastActivityAt    *time.Time
	
	createdAt   time.Time
	updatedAt   time.Time
	lastLoginAt *time.Time

	// Domain Events
	events []DomainEvent
}

// NewAgent cria um novo agente (factory method).
func NewAgent(
	projectID uuid.UUID,
	tenantID string,
	name string,
	agentType AgentType,
	userID *uuid.UUID, // Obrigatório para human, null para outros
) (*Agent, error) {
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if agentType == "" {
		agentType = AgentTypeHuman // Padrão
	}
	
	// Validação: agente humano precisa de userID
	if agentType == AgentTypeHuman && (userID == nil || *userID == uuid.Nil) {
		return nil, errors.New("human agent requires a valid userID")
	}

	now := time.Now()
	agent := &Agent{
		id:                uuid.New(),
		projectID:         projectID,
		userID:            userID,
		tenantID:          tenantID,
		name:              name,
		agentType:         agentType,
		status:            AgentStatusOffline,
		role:              RoleHumanAgent, // Padrão para humanos
		active:            true,
		config:            make(map[string]interface{}),
		permissions:       make(map[string]bool),
		settings:          make(map[string]interface{}),
		sessionsHandled:   0,
		averageResponseMs: 0,
		createdAt:         now,
		updatedAt:         now,
		events:            []DomainEvent{},
	}

	agent.addEvent(AgentCreatedEvent{
		AgentID:   agent.id,
		TenantID:  tenantID,
		Name:      name,
		Email:     agent.email,
		Role:      agent.role,
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

// SetStatus atualiza o status do agente
func (a *Agent) SetStatus(status AgentStatus) {
	if a.status != status {
		a.status = status
		a.updatedAt = time.Now()
		a.lastActivityAt = &a.updatedAt
	}
}

// SetConfig atualiza configurações específicas do agente (ex: AI provider)
func (a *Agent) SetConfig(config map[string]interface{}) {
	a.config = config
	a.updatedAt = time.Now()
}

// RecordSessionHandled registra que o agente atendeu uma sessão
func (a *Agent) RecordSessionHandled(responseTimeMs int) {
	a.sessionsHandled++
	
	// Calcula média móvel do tempo de resposta
	if a.averageResponseMs == 0 {
		a.averageResponseMs = responseTimeMs
	} else {
		a.averageResponseMs = (a.averageResponseMs + responseTimeMs) / 2
	}
	
	now := time.Now()
	a.lastActivityAt = &now
	a.updatedAt = now
}

// Getters
func (a *Agent) ID() uuid.UUID                    { return a.id }
func (a *Agent) ProjectID() uuid.UUID             { return a.projectID }
func (a *Agent) UserID() *uuid.UUID               { return a.userID }
func (a *Agent) TenantID() string                 { return a.tenantID }
func (a *Agent) Name() string                     { return a.name }
func (a *Agent) Email() string                    { return a.email }
func (a *Agent) Type() AgentType                  { return a.agentType }
func (a *Agent) Status() AgentStatus              { return a.status }
func (a *Agent) Role() Role                       { return a.role }
func (a *Agent) IsActive() bool                   { return a.active }
func (a *Agent) Config() map[string]interface{}   { return a.config }
func (a *Agent) Permissions() map[string]bool     { return a.permissions }
func (a *Agent) Settings() map[string]interface{} { return a.settings }
func (a *Agent) SessionsHandled() int             { return a.sessionsHandled }
func (a *Agent) AverageResponseMs() int           { return a.averageResponseMs }
func (a *Agent) LastActivityAt() *time.Time       { return a.lastActivityAt }
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
