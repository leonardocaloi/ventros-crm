package event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository define o contrato de persistência para Events.
type Repository interface {
	// Save persiste um evento.
	Save(ctx context.Context, event *Event) error

	// FindByID busca um evento por ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Event, error)

	// FindByContact busca eventos de um contato ordenados por timestamp.
	FindByContact(ctx context.Context, contactID uuid.UUID, limit int) ([]*Event, error)

	// FindBySession busca eventos de uma sessão ordenados por sequence_number.
	FindBySession(ctx context.Context, sessionID uuid.UUID) ([]*Event, error)

	// FindByTenantAndType busca eventos por tenant e tipo.
	FindByTenantAndType(ctx context.Context, tenantID, eventType string, limit int) ([]*Event, error)

	// FindByTimeRange busca eventos em um intervalo de tempo.
	FindByTimeRange(ctx context.Context, tenantID string, start, end time.Time) ([]*Event, error)

	// CountByContact conta eventos de um contato.
	CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)

	// CountBySession conta eventos de uma sessão.
	CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error)
}
