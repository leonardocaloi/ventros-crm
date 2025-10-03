package contact

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrContactNotFound = errors.New("contact not found")
)

// Repository define as operações de persistência para Contact.
type Repository interface {
	// Save persiste um contato (create ou update).
	Save(ctx context.Context, contact *Contact) error

	// FindByID busca um contato por ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)

	// FindByPhone busca contato por telefone no projeto.
	FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*Contact, error)

	// FindByEmail busca contato por email no projeto.
	FindByEmail(ctx context.Context, projectID uuid.UUID, email string) (*Contact, error)

	// FindByExternalID busca contato por external_id no projeto.
	FindByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*Contact, error)

	// FindByProject lista contatos de um projeto.
	FindByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*Contact, error)

	// CountByProject conta contatos de um projeto.
	CountByProject(ctx context.Context, projectID uuid.UUID) (int, error)
}
