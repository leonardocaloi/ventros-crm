package project_member

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// ProjectMemberInvitedEvent é disparado quando um novo membro é convidado
type ProjectMemberInvitedEvent struct {
	shared.BaseEvent
	ProjectMemberID uuid.UUID `json:"project_member_id"`
	ProjectID       uuid.UUID `json:"project_id"`
	AgentID         string    `json:"agent_id"`
	Role            string    `json:"role"`
	InvitedBy       string    `json:"invited_by"`
	InvitedAt       time.Time `json:"invited_at"`
}

func NewProjectMemberInvitedEvent(
	projectMemberID, projectID uuid.UUID,
	agentID, role, invitedBy string,
	invitedAt time.Time,
) *ProjectMemberInvitedEvent {
	return &ProjectMemberInvitedEvent{
		BaseEvent:       shared.NewBaseEvent("project_member.invited", invitedAt),
		ProjectMemberID: projectMemberID,
		ProjectID:       projectID,
		AgentID:         agentID,
		Role:            role,
		InvitedBy:       invitedBy,
		InvitedAt:       invitedAt,
	}
}

// ProjectMemberRoleChangedEvent é disparado quando o role de um membro muda
type ProjectMemberRoleChangedEvent struct {
	shared.BaseEvent
	ProjectMemberID uuid.UUID `json:"project_member_id"`
	ProjectID       uuid.UUID `json:"project_id"`
	AgentID         string    `json:"agent_id"`
	OldRole         string    `json:"old_role"`
	NewRole         string    `json:"new_role"`
	ChangedBy       string    `json:"changed_by"`
	ChangedAt       time.Time `json:"changed_at"`
}

func NewProjectMemberRoleChangedEvent(
	projectMemberID, projectID uuid.UUID,
	agentID, oldRole, newRole, changedBy string,
	changedAt time.Time,
) *ProjectMemberRoleChangedEvent {
	return &ProjectMemberRoleChangedEvent{
		BaseEvent:       shared.NewBaseEvent("project_member.role_changed", changedAt),
		ProjectMemberID: projectMemberID,
		ProjectID:       projectID,
		AgentID:         agentID,
		OldRole:         oldRole,
		NewRole:         newRole,
		ChangedBy:       changedBy,
		ChangedAt:       changedAt,
	}
}

// ProjectMemberRemovedEvent é disparado quando um membro é removido
type ProjectMemberRemovedEvent struct {
	shared.BaseEvent
	ProjectMemberID uuid.UUID `json:"project_member_id"`
	ProjectID       uuid.UUID `json:"project_id"`
	AgentID         string    `json:"agent_id"`
	Role            string    `json:"role"`
	RemovedBy       string    `json:"removed_by"`
	RemovedAt       time.Time `json:"removed_at"`
}

func NewProjectMemberRemovedEvent(
	projectMemberID, projectID uuid.UUID,
	agentID, role, removedBy string,
	removedAt time.Time,
) *ProjectMemberRemovedEvent {
	return &ProjectMemberRemovedEvent{
		BaseEvent:       shared.NewBaseEvent("project_member.removed", removedAt),
		ProjectMemberID: projectMemberID,
		ProjectID:       projectID,
		AgentID:         agentID,
		Role:            role,
		RemovedBy:       removedBy,
		RemovedAt:       removedAt,
	}
}
