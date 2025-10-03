package agent_session

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para eventos de domínio.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// AgentJoinedSessionEvent é emitido quando um agente entra em uma sessão.
type AgentJoinedSessionEvent struct {
	AgentSessionID uuid.UUID
	AgentID        uuid.UUID
	SessionID      uuid.UUID
	Role           *RoleInSession
	JoinedAt       time.Time
}

func (e AgentJoinedSessionEvent) EventName() string {
	return "agent_session.joined"
}

func (e AgentJoinedSessionEvent) OccurredAt() time.Time {
	return e.JoinedAt
}

// AgentLeftSessionEvent é emitido quando um agente sai de uma sessão.
type AgentLeftSessionEvent struct {
	AgentSessionID uuid.UUID
	AgentID        uuid.UUID
	SessionID      uuid.UUID
	LeftAt         time.Time
}

func (e AgentLeftSessionEvent) EventName() string {
	return "agent_session.left"
}

func (e AgentLeftSessionEvent) OccurredAt() time.Time {
	return e.LeftAt
}

// AgentRoleChangedEvent é emitido quando o papel do agente muda na sessão.
type AgentRoleChangedEvent struct {
	AgentSessionID uuid.UUID
	AgentID        uuid.UUID
	SessionID      uuid.UUID
	OldRole        *RoleInSession
	NewRole        *RoleInSession
	ChangedAt      time.Time
}

func (e AgentRoleChangedEvent) EventName() string {
	return "agent_session.role_changed"
}

func (e AgentRoleChangedEvent) OccurredAt() time.Time {
	return e.ChangedAt
}
