package credential

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para persistência de credenciais
type Repository interface {
	// Save persiste uma credencial (cria ou atualiza)
	Save(ctx context.Context, credential *Credential) error

	// FindByID busca uma credencial por ID
	FindByID(ctx context.Context, id uuid.UUID) (*Credential, error)

	// FindByTenantAndType busca credenciais por tenant e tipo
	FindByTenantAndType(ctx context.Context, tenantID string, credType CredentialType) ([]*Credential, error)

	// FindByTenantAndName busca uma credencial específica por nome
	FindByTenantAndName(ctx context.Context, tenantID string, name string) (*Credential, error)

	// FindByProjectAndType busca credenciais de um projeto específico
	FindByProjectAndType(ctx context.Context, projectID uuid.UUID, credType CredentialType) ([]*Credential, error)

	// FindActiveByTenant busca todas as credenciais ativas de um tenant
	FindActiveByTenant(ctx context.Context, tenantID string) ([]*Credential, error)

	// FindExpiring busca credenciais que expiram em breve (para renovação)
	FindExpiring(ctx context.Context, withinMinutes int) ([]*Credential, error)

	// Delete remove uma credencial (soft delete via Deactivate)
	Delete(ctx context.Context, id uuid.UUID) error
}
