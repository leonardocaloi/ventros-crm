package project_member

import (
	"context"

	"github.com/google/uuid"
)

// Repository define as operações de persistência para ProjectMember
type Repository interface {
	// Save persiste um ProjectMember (create ou update)
	Save(ctx context.Context, member *ProjectMember) error

	// FindByID busca um ProjectMember por ID
	FindByID(ctx context.Context, id uuid.UUID) (*ProjectMember, error)

	// FindByProjectAndAgent busca um membro específico de um projeto
	FindByProjectAndAgent(ctx context.Context, projectID uuid.UUID, agentID string) (*ProjectMember, error)

	// FindByProject busca todos os membros de um projeto
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)

	// FindByAgent busca todos os projetos de um agent
	FindByAgent(ctx context.Context, agentID string) ([]*ProjectMember, error)

	// FindAdminsByProject busca todos os admins de um projeto
	FindAdminsByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)

	// CountAdminsByProject conta quantos admins tem em um projeto
	CountAdminsByProject(ctx context.Context, projectID uuid.UUID) (int, error)

	// ExistsInProject verifica se um agent já é membro de um projeto
	ExistsInProject(ctx context.Context, projectID uuid.UUID, agentID string) (bool, error)

	// Delete remove um ProjectMember
	Delete(ctx context.Context, id uuid.UUID) error
}
