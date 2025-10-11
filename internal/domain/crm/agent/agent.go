package agent

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrAgentNotFound = errors.New("agent not found")
)

// Permission constants
const (
	PermissionReassignSessions  = "reassign_sessions"
	PermissionManageAgents      = "manage_agents"
	PermissionViewAllSessions   = "view_all_sessions"
	PermissionSendMessages      = "send_messages"
	PermissionAccessAnalytics   = "access_analytics"
	PermissionManageAutomations = "manage_automations"
)

type AgentType string

const (
	AgentTypeHuman   AgentType = "human"
	AgentTypeAI      AgentType = "ai"
	AgentTypeBot     AgentType = "bot"
	AgentTypeChannel AgentType = "channel"
	AgentTypeVirtual AgentType = "virtual" // Representa pessoas do passado (histórico), não pode enviar mensagens
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

	// Virtual agent metadata (for historical representation)
	virtualMetadata *VirtualAgentMetadata

	createdAt   time.Time
	updatedAt   time.Time
	lastLoginAt *time.Time

	events []DomainEvent
}

// VirtualAgentMetadata contém metadados para agentes virtuais
// Usado para representar pessoas do passado e segmentação de métricas
type VirtualAgentMetadata struct {
	RepresentsPersonName string     // Nome da pessoa representada
	PeriodStart          time.Time  // Início do período representado
	PeriodEnd            *time.Time // Fim do período (nil se ainda ativo)
	Reason               string     // Razão da criação (ex: "device_attribution", "historical_contact")
	SourceDevice         *string    // Device ID original (se aplicável)
	Notes                string     // Notas adicionais
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
	if agentType == AgentTypeVirtual {
		return nil, errors.New("use NewVirtualAgent() to create virtual agents")
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

// NewVirtualAgent cria um agente virtual que representa uma pessoa do passado
// Agentes virtuais:
// - NÃO podem enviar mensagens
// - NÃO podem ser atribuídos manualmente
// - Servem apenas para segmentação de métricas e histórico
// - Sobrescrevem o conceito de "device" para mensagens antigas
func NewVirtualAgent(
	projectID uuid.UUID,
	tenantID string,
	representsPersonName string,
	periodStart time.Time,
	reason string,
	sourceDevice *string,
	notes string,
) (*Agent, error) {
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if representsPersonName == "" {
		return nil, errors.New("representsPersonName cannot be empty")
	}
	if reason == "" {
		return nil, errors.New("reason cannot be empty")
	}

	now := time.Now()
	virtualName := "[Virtual] " + representsPersonName

	agent := &Agent{
		id:                uuid.New(),
		projectID:         projectID,
		userID:            nil, // Virtual agents não têm userID
		tenantID:          tenantID,
		name:              virtualName,
		email:             "", // Virtual agents não têm email
		agentType:         AgentTypeVirtual,
		status:            AgentStatusOffline, // Sempre offline
		role:              RoleHumanAgent,     // Representa um humano do passado
		active:            true,
		config:            make(map[string]interface{}),
		permissions:       make(map[string]bool),
		settings:          make(map[string]interface{}),
		sessionsHandled:   0,
		averageResponseMs: 0,
		virtualMetadata: &VirtualAgentMetadata{
			RepresentsPersonName: representsPersonName,
			PeriodStart:          periodStart,
			PeriodEnd:            nil,
			Reason:               reason,
			SourceDevice:         sourceDevice,
			Notes:                notes,
		},
		createdAt: now,
		updatedAt: now,
		events:    []DomainEvent{},
	}

	agent.addEvent(NewAgentCreatedEvent(agent.id, tenantID, virtualName, agent.email, agent.role))

	return agent, nil
}

// EndVirtualAgentPeriod marca o fim do período que o agente virtual representa
// Útil quando um número passa para outra pessoa
func (a *Agent) EndVirtualAgentPeriod(endDate time.Time) error {
	if a.agentType != AgentTypeVirtual {
		return errors.New("only virtual agents can have period ended")
	}
	if a.virtualMetadata == nil {
		return errors.New("virtual agent has no metadata")
	}
	if endDate.Before(a.virtualMetadata.PeriodStart) {
		return errors.New("end date cannot be before start date")
	}

	a.virtualMetadata.PeriodEnd = &endDate
	a.updatedAt = time.Now()

	return nil
}

// IsVirtual verifica se o agente é virtual
func (a *Agent) IsVirtual() bool {
	return a.agentType == AgentTypeVirtual
}

// CanSendMessages verifica se o agente pode enviar mensagens
// Agentes virtuais NÃO podem enviar mensagens
func (a *Agent) CanSendMessages() bool {
	return a.agentType != AgentTypeVirtual && a.active
}

// CanBeManuallyAssigned verifica se o agente pode ser atribuído manualmente
// Agentes virtuais NÃO podem ser atribuídos manualmente
func (a *Agent) CanBeManuallyAssigned() bool {
	return a.agentType != AgentTypeVirtual && a.active
}

// ShouldCountInMetrics verifica se o agente deve ser contado em métricas de desempenho
// Agentes virtuais NÃO são contados em métricas de desempenho de agentes reais
func (a *Agent) ShouldCountInMetrics() bool {
	return a.agentType != AgentTypeVirtual
}

// CanReassignSessions verifica se o agente tem permissão para reatribuir sessões
// Usado para controlar quais agentes podem atribuir sessões a outros agentes
func (a *Agent) CanReassignSessions() bool {
	// Agentes virtuais nunca podem reatribuir sessões
	if a.agentType == AgentTypeVirtual {
		return false
	}
	// Verifica se tem permissão explícita ou se é supervisor/admin
	return a.HasPermission(PermissionReassignSessions) || a.HasPermission(PermissionManageAgents)
}

func ReconstructAgent(
	id uuid.UUID,
	projectID uuid.UUID,
	userID *uuid.UUID,
	tenantID string,
	name string,
	email string,
	agentType AgentType,
	status AgentStatus,
	role Role,
	active bool,
	config map[string]interface{},
	permissions map[string]bool,
	settings map[string]interface{},
	sessionsHandled int,
	averageResponseMs int,
	lastActivityAt *time.Time,
	virtualMetadata *VirtualAgentMetadata,
	createdAt time.Time,
	updatedAt time.Time,
	lastLoginAt *time.Time,
) *Agent {
	return &Agent{
		id:                id,
		projectID:         projectID,
		userID:            userID,
		tenantID:          tenantID,
		name:              name,
		email:             email,
		agentType:         agentType,
		status:            status,
		role:              role,
		active:            active,
		config:            config,
		permissions:       permissions,
		settings:          settings,
		sessionsHandled:   sessionsHandled,
		averageResponseMs: averageResponseMs,
		lastActivityAt:    lastActivityAt,
		virtualMetadata:   virtualMetadata,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		lastLoginAt:       lastLoginAt,
		events:            []DomainEvent{},
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
	// Agentes virtuais não acumulam métricas de desempenho
	if a.agentType == AgentTypeVirtual {
		return
	}

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

func (a *Agent) ID() uuid.UUID                      { return a.id }
func (a *Agent) ProjectID() uuid.UUID               { return a.projectID }
func (a *Agent) UserID() *uuid.UUID                 { return a.userID }
func (a *Agent) TenantID() string                   { return a.tenantID }
func (a *Agent) Name() string                       { return a.name }
func (a *Agent) Email() string                      { return a.email }
func (a *Agent) Type() AgentType                    { return a.agentType }
func (a *Agent) Status() AgentStatus                { return a.status }
func (a *Agent) Role() Role                         { return a.role }
func (a *Agent) IsActive() bool                     { return a.active }
func (a *Agent) Config() map[string]interface{}     { return a.config }
func (a *Agent) Permissions() map[string]bool       { return a.permissions }
func (a *Agent) Settings() map[string]interface{}   { return a.settings }
func (a *Agent) SessionsHandled() int               { return a.sessionsHandled }
func (a *Agent) AverageResponseMs() int             { return a.averageResponseMs }
func (a *Agent) LastActivityAt() *time.Time         { return a.lastActivityAt }
func (a *Agent) VirtualMetadata() *VirtualAgentMetadata { return a.virtualMetadata }
func (a *Agent) CreatedAt() time.Time               { return a.createdAt }
func (a *Agent) UpdatedAt() time.Time               { return a.updatedAt }
func (a *Agent) LastLoginAt() *time.Time            { return a.lastLoginAt }

func (a *Agent) DomainEvents() []DomainEvent {
	return a.events
}

func (a *Agent) ClearEvents() {
	a.events = []DomainEvent{}
}

func (a *Agent) addEvent(event DomainEvent) {
	a.events = append(a.events, event)
}
