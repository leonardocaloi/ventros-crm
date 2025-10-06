package agent

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para eventos de domínio.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// AgentCreatedEvent - Agente criado no sistema.
type AgentCreatedEvent struct {
	AgentID   uuid.UUID
	TenantID  string
	Name      string
	Email     string
	Role      Role
	CreatedAt time.Time
}

func (e AgentCreatedEvent) EventName() string     { return "agent.created" }
func (e AgentCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// AgentUpdatedEvent - Informações do agente atualizadas.
type AgentUpdatedEvent struct {
	AgentID   uuid.UUID
	Changes   map[string]interface{}
	UpdatedAt time.Time
}

func (e AgentUpdatedEvent) EventName() string     { return "agent.updated" }
func (e AgentUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

// AgentActivatedEvent - Agente ativado.
type AgentActivatedEvent struct {
	AgentID     uuid.UUID
	ActivatedAt time.Time
}

func (e AgentActivatedEvent) EventName() string     { return "agent.activated" }
func (e AgentActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

// AgentDeactivatedEvent - Agente desativado.
type AgentDeactivatedEvent struct {
	AgentID       uuid.UUID
	DeactivatedAt time.Time
}

func (e AgentDeactivatedEvent) EventName() string     { return "agent.deactivated" }
func (e AgentDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

// AgentLoggedInEvent - Agente fez login.
type AgentLoggedInEvent struct {
	AgentID    uuid.UUID
	LoggedInAt time.Time
}

func (e AgentLoggedInEvent) EventName() string     { return "agent.logged_in" }
func (e AgentLoggedInEvent) OccurredAt() time.Time { return e.LoggedInAt }

// AgentPermissionGrantedEvent - Permissão concedida ao agente.
type AgentPermissionGrantedEvent struct {
	AgentID    uuid.UUID
	Permission string
	GrantedAt  time.Time
}

func (e AgentPermissionGrantedEvent) EventName() string     { return "agent.permission_granted" }
func (e AgentPermissionGrantedEvent) OccurredAt() time.Time { return e.GrantedAt }

// AgentPermissionRevokedEvent - Permissão revogada do agente.
type AgentPermissionRevokedEvent struct {
	AgentID    uuid.UUID
	Permission string
	RevokedAt  time.Time
}

func (e AgentPermissionRevokedEvent) EventName() string     { return "agent.permission_revoked" }
func (e AgentPermissionRevokedEvent) OccurredAt() time.Time { return e.RevokedAt }
