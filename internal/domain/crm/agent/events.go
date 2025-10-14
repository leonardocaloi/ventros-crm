package agent

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type DomainEvent = shared.DomainEvent

type AgentCreatedEvent struct {
	shared.BaseEvent
	AgentID  uuid.UUID
	TenantID string
	Name     string
	Email    string
	Role     Role
}

func NewAgentCreatedEvent(agentID uuid.UUID, tenantID, name, email string, role Role) AgentCreatedEvent {
	return AgentCreatedEvent{
		BaseEvent: shared.NewBaseEvent("agent.created", time.Now()),
		AgentID:   agentID,
		TenantID:  tenantID,
		Name:      name,
		Email:     email,
		Role:      role,
	}
}

type AgentUpdatedEvent struct {
	shared.BaseEvent
	AgentID uuid.UUID
	Changes map[string]interface{}
}

func NewAgentUpdatedEvent(agentID uuid.UUID, changes map[string]interface{}) AgentUpdatedEvent {
	return AgentUpdatedEvent{
		BaseEvent: shared.NewBaseEvent("agent.updated", time.Now()),
		AgentID:   agentID,
		Changes:   changes,
	}
}

type AgentActivatedEvent struct {
	shared.BaseEvent
	AgentID uuid.UUID
}

func NewAgentActivatedEvent(agentID uuid.UUID) AgentActivatedEvent {
	return AgentActivatedEvent{
		BaseEvent: shared.NewBaseEvent("agent.activated", time.Now()),
		AgentID:   agentID,
	}
}

type AgentDeactivatedEvent struct {
	shared.BaseEvent
	AgentID uuid.UUID
}

func NewAgentDeactivatedEvent(agentID uuid.UUID) AgentDeactivatedEvent {
	return AgentDeactivatedEvent{
		BaseEvent: shared.NewBaseEvent("agent.deactivated", time.Now()),
		AgentID:   agentID,
	}
}

type AgentLoggedInEvent struct {
	shared.BaseEvent
	AgentID uuid.UUID
}

func NewAgentLoggedInEvent(agentID uuid.UUID) AgentLoggedInEvent {
	return AgentLoggedInEvent{
		BaseEvent: shared.NewBaseEvent("agent.logged_in", time.Now()),
		AgentID:   agentID,
	}
}

type AgentPermissionGrantedEvent struct {
	shared.BaseEvent
	AgentID    uuid.UUID
	Permission string
}

func NewAgentPermissionGrantedEvent(agentID uuid.UUID, permission string) AgentPermissionGrantedEvent {
	return AgentPermissionGrantedEvent{
		BaseEvent:  shared.NewBaseEvent("agent.permission_granted", time.Now()),
		AgentID:    agentID,
		Permission: permission,
	}
}

type AgentPermissionRevokedEvent struct {
	shared.BaseEvent
	AgentID    uuid.UUID
	Permission string
}

func NewAgentPermissionRevokedEvent(agentID uuid.UUID, permission string) AgentPermissionRevokedEvent {
	return AgentPermissionRevokedEvent{
		BaseEvent:  shared.NewBaseEvent("agent.permission_revoked", time.Now()),
		AgentID:    agentID,
		Permission: permission,
	}
}
