package note

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a interface para persistência de Notes
type Repository interface {
	// Save salva uma nota (create ou update)
	Save(ctx context.Context, note *Note) error

	// FindByID busca uma nota por ID
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)

	// FindByContactID busca notas de um contato
	FindByContactID(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*Note, error)

	// FindBySessionID busca notas de uma sessão
	FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*Note, error)

	// FindPinned busca notas fixadas de um contato
	FindPinned(ctx context.Context, contactID uuid.UUID) ([]*Note, error)

	// Delete deleta uma nota (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// CountByContact conta notas de um contato
	CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)
}
