package agent_session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AgentSession é uma entidade que representa a participação de um agente em uma sessão.
// Implementa relacionamento Many-to-Many entre Agent e Session.
// Importante para Google ADK e orquestração de múltiplos agentes (humanos + IA).
type AgentSession struct {
	id            uuid.UUID
	agentID       uuid.UUID
	sessionID     uuid.UUID
	roleInSession *RoleInSession
	joinedAt      time.Time
	leftAt        *time.Time
	isActive      bool
	metadata      map[string]interface{} // Para integração ADK
	createdAt     time.Time
	updatedAt     time.Time

	// Domain Events
	events []DomainEvent
}

// NewAgentSession cria uma nova participação de agente em sessão.
func NewAgentSession(
	agentID uuid.UUID,
	sessionID uuid.UUID,
	roleInSession *RoleInSession,
) (*AgentSession, error) {
	if agentID == uuid.Nil {
		return nil, errors.New("agentID cannot be nil")
	}
	if sessionID == uuid.Nil {
		return nil, errors.New("sessionID cannot be nil")
	}

	now := time.Now()
	as := &AgentSession{
		id:            uuid.New(),
		agentID:       agentID,
		sessionID:     sessionID,
		roleInSession: roleInSession,
		joinedAt:      now,
		isActive:      true,
		metadata:      make(map[string]interface{}),
		createdAt:     now,
		updatedAt:     now,
		events:        []DomainEvent{},
	}

	as.addEvent(AgentJoinedSessionEvent{
		AgentSessionID: as.id,
		AgentID:        agentID,
		SessionID:      sessionID,
		Role:           roleInSession,
		JoinedAt:       now,
	})

	return as, nil
}

// ReconstructAgentSession reconstrói a partir de dados persistidos.
func ReconstructAgentSession(
	id uuid.UUID,
	agentID uuid.UUID,
	sessionID uuid.UUID,
	roleInSession *RoleInSession,
	joinedAt time.Time,
	leftAt *time.Time,
	isActive bool,
	metadata map[string]interface{},
	createdAt time.Time,
	updatedAt time.Time,
) *AgentSession {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &AgentSession{
		id:            id,
		agentID:       agentID,
		sessionID:     sessionID,
		roleInSession: roleInSession,
		joinedAt:      joinedAt,
		leftAt:        leftAt,
		isActive:      isActive,
		metadata:      metadata,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
		events:        []DomainEvent{},
	}
}

// Leave marca que o agente saiu da sessão.
func (as *AgentSession) Leave() error {
	if !as.isActive {
		return errors.New("agent is not active in this session")
	}

	now := time.Now()
	as.isActive = false
	as.leftAt = &now
	as.updatedAt = now

	as.addEvent(AgentLeftSessionEvent{
		AgentSessionID: as.id,
		AgentID:        as.agentID,
		SessionID:      as.sessionID,
		LeftAt:         now,
	})

	return nil
}

// UpdateMetadata atualiza metadados para integração ADK.
func (as *AgentSession) UpdateMetadata(metadata map[string]interface{}) {
	as.metadata = metadata
	as.updatedAt = time.Now()
}

// ChangeRole muda o papel do agente na sessão.
func (as *AgentSession) ChangeRole(newRole RoleInSession) error {
	oldRole := as.roleInSession
	as.roleInSession = &newRole
	as.updatedAt = time.Now()

	as.addEvent(AgentRoleChangedEvent{
		AgentSessionID: as.id,
		AgentID:        as.agentID,
		SessionID:      as.sessionID,
		OldRole:        oldRole,
		NewRole:        &newRole,
		ChangedAt:      as.updatedAt,
	})

	return nil
}

// Getters
func (as *AgentSession) ID() uuid.UUID                 { return as.id }
func (as *AgentSession) AgentID() uuid.UUID            { return as.agentID }
func (as *AgentSession) SessionID() uuid.UUID          { return as.sessionID }
func (as *AgentSession) RoleInSession() *RoleInSession { return as.roleInSession }
func (as *AgentSession) JoinedAt() time.Time           { return as.joinedAt }
func (as *AgentSession) LeftAt() *time.Time            { return as.leftAt }
func (as *AgentSession) IsActive() bool                { return as.isActive }
func (as *AgentSession) Metadata() map[string]interface{} {
	// Return copy to preserve immutability
	copy := make(map[string]interface{})
	for k, v := range as.metadata {
		copy[k] = v
	}
	return copy
}
func (as *AgentSession) CreatedAt() time.Time { return as.createdAt }
func (as *AgentSession) UpdatedAt() time.Time { return as.updatedAt }

// Domain Events
func (as *AgentSession) DomainEvents() []DomainEvent {
	return as.events
}

func (as *AgentSession) ClearEvents() {
	as.events = []DomainEvent{}
}

func (as *AgentSession) addEvent(event DomainEvent) {
	as.events = append(as.events, event)
}
