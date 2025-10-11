package agent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAgentNotFound = errors.New("agent not found")
)

type AgentType string

const (
	AgentTypeHuman   AgentType = "human"
	AgentTypeAI      AgentType = "ai"
	AgentTypeBot     AgentType = "bot"
	AgentTypeChannel AgentType = "channel"
)

type AgentStatus string

const (
	AgentStatusAvailable AgentStatus = "available"
	AgentStatusBusy      AgentStatus = "busy"
	AgentStatusAway      AgentStatus = "away"
	AgentStatusOffline   AgentStatus = "offline"
)

type Agent struct {
	id          uuid.UUID
	projectID   uuid.UUID
	userID      *uuid.UUID
	tenantID    string
	name        string
	email       string
	agentType   AgentType
	status      AgentStatus
	role        Role
	active      bool
	config      map[string]interface{}
	permissions map[string]bool
	settings    map[string]interface{}

	sessionsHandled   int
	averageResponseMs int
	lastActivityAt    *time.Time

	createdAt   time.Time
	updatedAt   time.Time
	lastLoginAt *time.Time

	events []DomainEvent
}

func NewAgent(
	projectID uuid.UUID,
	tenantID string,
	name string,
	agentType AgentType,
	userID *uuid.UUID,
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
		agentType = AgentTypeHuman
	}

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
		role:              RoleHumanAgent,
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

	agent.addEvent(NewAgentCreatedEvent(agent.id, tenantID, name, agent.email, agent.role))

	return agent, nil
}

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
		a.addEvent(NewAgentUpdatedEvent(a.id, changes))
	}

	return nil
}

func (a *Agent) Activate() error {
	if a.active {
		return errors.New("agent is already active")
	}

	a.active = true
	a.updatedAt = time.Now()

	a.addEvent(NewAgentActivatedEvent(a.id))

	return nil
}

func (a *Agent) Deactivate() error {
	if !a.active {
		return errors.New("agent is already inactive")
	}

	a.active = false
	a.updatedAt = time.Now()

	a.addEvent(NewAgentDeactivatedEvent(a.id))

	return nil
}

func (a *Agent) RecordLogin() {
	now := time.Now()
	a.lastLoginAt = &now
	a.updatedAt = now

	a.addEvent(NewAgentLoggedInEvent(a.id))
}

func (a *Agent) GrantPermission(permission string) error {
	if permission == "" {
		return errors.New("permission cannot be empty")
	}

	if a.permissions[permission] {
		return nil
	}

	a.permissions[permission] = true
	a.updatedAt = time.Now()

	a.addEvent(NewAgentPermissionGrantedEvent(a.id, permission))

	return nil
}

func (a *Agent) RevokePermission(permission string) error {
	if permission == "" {
		return errors.New("permission cannot be empty")
	}

	if !a.permissions[permission] {
		return nil
	}

	delete(a.permissions, permission)
	a.updatedAt = time.Now()

	a.addEvent(NewAgentPermissionRevokedEvent(a.id, permission))

	return nil
}

func (a *Agent) HasPermission(permission string) bool {
	return a.permissions[permission]
}

func (a *Agent) UpdateSettings(settings map[string]interface{}) {
	a.settings = settings
	a.updatedAt = time.Now()
}

func (a *Agent) SetStatus(status AgentStatus) {
	if a.status != status {
		a.status = status
		a.updatedAt = time.Now()
		a.lastActivityAt = &a.updatedAt
	}
}

func (a *Agent) SetConfig(config map[string]interface{}) {
	a.config = config
	a.updatedAt = time.Now()
}

func (a *Agent) RecordSessionHandled(responseTimeMs int) {
	a.sessionsHandled++

	if a.averageResponseMs == 0 {
		a.averageResponseMs = responseTimeMs
	} else {
		a.averageResponseMs = (a.averageResponseMs + responseTimeMs) / 2
	}

	now := time.Now()
	a.lastActivityAt = &now
	a.updatedAt = now
}

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

func (a *Agent) DomainEvents() []DomainEvent {
	return a.events
}

func (a *Agent) ClearEvents() {
	a.events = []DomainEvent{}
}

func (a *Agent) addEvent(event DomainEvent) {
	a.events = append(a.events, event)
}
