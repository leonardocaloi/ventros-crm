package project_member

import (
	"errors"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

// ProjectMemberRole representa os roles disponíveis em um projeto
type ProjectMemberRole string

const (
	RoleAdmin      ProjectMemberRole = "admin"      // Full access - pode convidar, remover, alterar roles
	RoleSupervisor ProjectMemberRole = "supervisor" // Pode ver analytics, gerenciar campaigns/sequences
	RoleAgent      ProjectMemberRole = "agent"      // Pode interagir com sessions/messages, ver contacts
	RoleViewer     ProjectMemberRole = "viewer"     // Read-only access
)

// ProjectMember representa a associação entre um Agent (usuário Keycloak) e um Project
// É um aggregate que gerencia permissões a nível de projeto
type ProjectMember struct {
	// Aggregate Root fields
	id      uuid.UUID
	version int

	// Core fields
	projectID uuid.UUID         // ID do projeto
	agentID   string            // ID do usuário no Keycloak (sub claim do JWT)
	role      ProjectMemberRole // Role do membro neste projeto

	// Audit fields
	invitedBy string    // ID do usuário que convidou
	invitedAt time.Time // Quando foi convidado
	createdAt time.Time
	updatedAt time.Time

	// Domain Events
	events []shared.DomainEvent
}

// Errors
var (
	ErrInvalidRole             = errors.New("invalid project member role")
	ErrAgentIDRequired         = errors.New("agent ID is required")
	ErrProjectIDRequired       = errors.New("project ID is required")
	ErrInvitedByRequired       = errors.New("invited by is required")
	ErrCannotChangeSelfRole    = errors.New("cannot change your own role")
	ErrCannotRemoveLastAdmin   = errors.New("cannot remove the last admin from project")
	ErrMemberAlreadyExists     = errors.New("member already exists in this project")
	ErrMemberNotFound          = errors.New("member not found")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

// NewProjectMember cria um novo membro de projeto
func NewProjectMember(
	projectID uuid.UUID,
	agentID string,
	role ProjectMemberRole,
	invitedBy string,
) (*ProjectMember, error) {
	// Validations
	if projectID == uuid.Nil {
		return nil, ErrProjectIDRequired
	}
	if agentID == "" {
		return nil, ErrAgentIDRequired
	}
	if !isValidRole(role) {
		return nil, ErrInvalidRole
	}
	if invitedBy == "" {
		return nil, ErrInvitedByRequired
	}

	now := time.Now()
	pm := &ProjectMember{
		id:        uuid.New(),
		version:   1,
		projectID: projectID,
		agentID:   agentID,
		role:      role,
		invitedBy: invitedBy,
		invitedAt: now,
		createdAt: now,
		updatedAt: now,
		events:    []shared.DomainEvent{},
	}

	// Domain event
	pm.AddEvent(NewProjectMemberInvitedEvent(
		pm.id,
		pm.projectID,
		pm.agentID,
		string(pm.role),
		pm.invitedBy,
		pm.invitedAt,
	))

	return pm, nil
}

// ReconstructProjectMember reconstrói um ProjectMember a partir do banco
func ReconstructProjectMember(
	id uuid.UUID,
	version int,
	projectID uuid.UUID,
	agentID string,
	role ProjectMemberRole,
	invitedBy string,
	invitedAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *ProjectMember {
	if version == 0 {
		version = 1
	}

	return &ProjectMember{
		id:        id,
		version:   version,
		projectID: projectID,
		agentID:   agentID,
		role:      role,
		invitedBy: invitedBy,
		invitedAt: invitedAt,
		createdAt: createdAt,
		updatedAt: updatedAt,
		events:    []shared.DomainEvent{},
	}
}

// ChangeRole altera o role do membro
func (pm *ProjectMember) ChangeRole(newRole ProjectMemberRole, changedBy string) error {
	if !isValidRole(newRole) {
		return ErrInvalidRole
	}

	// Não pode mudar o próprio role
	if pm.agentID == changedBy {
		return ErrCannotChangeSelfRole
	}

	oldRole := pm.role
	pm.role = newRole
	pm.updatedAt = time.Now()

	// Domain event
	pm.AddEvent(NewProjectMemberRoleChangedEvent(
		pm.id,
		pm.projectID,
		pm.agentID,
		string(oldRole),
		string(newRole),
		changedBy,
		pm.updatedAt,
	))

	return nil
}

// Remove marca o membro para remoção
func (pm *ProjectMember) Remove(removedBy string, isLastAdmin bool) error {
	// Não pode remover o último admin
	if pm.role == RoleAdmin && isLastAdmin {
		return ErrCannotRemoveLastAdmin
	}

	pm.updatedAt = time.Now()

	// Domain event
	pm.AddEvent(NewProjectMemberRemovedEvent(
		pm.id,
		pm.projectID,
		pm.agentID,
		string(pm.role),
		removedBy,
		pm.updatedAt,
	))

	return nil
}

// HasPermission verifica se o membro tem permissão para uma ação
func (pm *ProjectMember) HasPermission(permission Permission) bool {
	return RoleHasPermission(pm.role, permission)
}

// IsAdmin verifica se o membro é admin
func (pm *ProjectMember) IsAdmin() bool {
	return pm.role == RoleAdmin
}

// IsSupervisor verifica se o membro é supervisor ou superior
func (pm *ProjectMember) IsSupervisor() bool {
	return pm.role == RoleAdmin || pm.role == RoleSupervisor
}

// CanManageMembers verifica se pode gerenciar outros membros
func (pm *ProjectMember) CanManageMembers() bool {
	return pm.role == RoleAdmin
}

// CanManageCampaigns verifica se pode gerenciar campaigns
func (pm *ProjectMember) CanManageCampaigns() bool {
	return pm.role == RoleAdmin || pm.role == RoleSupervisor
}

// CanInteractWithSessions verifica se pode interagir com sessions
func (pm *ProjectMember) CanInteractWithSessions() bool {
	return pm.role != RoleViewer
}

// Getters
func (pm *ProjectMember) ID() uuid.UUID                 { return pm.id }
func (pm *ProjectMember) Version() int                  { return pm.version }
func (pm *ProjectMember) ProjectID() uuid.UUID          { return pm.projectID }
func (pm *ProjectMember) AgentID() string               { return pm.agentID }
func (pm *ProjectMember) Role() ProjectMemberRole       { return pm.role }
func (pm *ProjectMember) InvitedBy() string             { return pm.invitedBy }
func (pm *ProjectMember) InvitedAt() time.Time          { return pm.invitedAt }
func (pm *ProjectMember) CreatedAt() time.Time          { return pm.createdAt }
func (pm *ProjectMember) UpdatedAt() time.Time          { return pm.updatedAt }
func (pm *ProjectMember) DomainEvents() []shared.DomainEvent { return pm.events }

// ClearEvents limpa os eventos de domínio
func (pm *ProjectMember) ClearEvents() {
	pm.events = []shared.DomainEvent{}
}

// AddEvent adiciona um evento de domínio
func (pm *ProjectMember) AddEvent(event shared.DomainEvent) {
	pm.events = append(pm.events, event)
}

// Helper functions
func isValidRole(role ProjectMemberRole) bool {
	switch role {
	case RoleAdmin, RoleSupervisor, RoleAgent, RoleViewer:
		return true
	default:
		return false
	}
}

// ValidRoles retorna a lista de roles válidos
func ValidRoles() []ProjectMemberRole {
	return []ProjectMemberRole{
		RoleAdmin,
		RoleSupervisor,
		RoleAgent,
		RoleViewer,
	}
}
