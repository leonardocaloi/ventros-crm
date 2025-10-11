package credential

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, credential *Credential) error

	FindByID(ctx context.Context, id uuid.UUID) (*Credential, error)

	FindByTenantAndType(ctx context.Context, tenantID string, credType CredentialType) ([]*Credential, error)

	FindByTenantAndName(ctx context.Context, tenantID string, name string) (*Credential, error)

	FindByProjectAndType(ctx context.Context, projectID uuid.UUID, credType CredentialType) ([]*Credential, error)

	FindActiveByTenant(ctx context.Context, tenantID string) ([]*Credential, error)

	FindExpiring(ctx context.Context, withinMinutes int) ([]*Credential, error)

	Delete(ctx context.Context, id uuid.UUID) error
}
