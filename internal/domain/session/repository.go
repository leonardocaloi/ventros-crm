package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// Repository define as operações de persistência para Session.
// Esta é uma INTERFACE - a implementação está em infrastructure/persistence.
type Repository interface {
	// Save persiste uma sessão (create ou update).
	Save(ctx context.Context, session *Session) error

	// FindByID busca uma sessão por ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Session, error)

	// FindActiveByContact busca a sessão ativa de um contato em um canal.
	FindActiveByContact(ctx context.Context, contactID uuid.UUID, channelTypeID *int) (*Session, error)

	// FindInactiveSessions busca sessões que ultrapassaram o timeout.
	FindInactiveSessions(ctx context.Context, tenantID string) ([]*Session, error)

	// FindSessionsRequiringSummary busca sessões encerradas sem resumo.
	FindSessionsRequiringSummary(ctx context.Context, tenantID string, limit int) ([]*Session, error)

	// CountActiveByTenant conta sessões ativas por tenant.
	CountActiveByTenant(ctx context.Context, tenantID string) (int, error)

	// FindActiveBeforeTime busca sessões ativas com última atividade antes do tempo especificado.
	FindActiveBeforeTime(ctx context.Context, cutoffTime time.Time) ([]*Session, error)
}
