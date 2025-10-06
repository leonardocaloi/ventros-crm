package contact_event

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository define a interface de persistência para ContactEvent.
type Repository interface {
	// Save persiste um novo evento de contato.
	Save(ctx context.Context, event *ContactEvent) error

	// Update atualiza um evento de contato existente.
	Update(ctx context.Context, event *ContactEvent) error

	// FindByID busca um evento por ID.
	FindByID(ctx context.Context, id uuid.UUID) (*ContactEvent, error)

	// FindByContactID busca todos os eventos de um contato, ordenados por tempo.
	FindByContactID(ctx context.Context, contactID uuid.UUID, limit int, offset int) ([]*ContactEvent, error)

	// FindByContactIDVisible busca eventos visíveis para o cliente de um contato.
	FindByContactIDVisible(ctx context.Context, contactID uuid.UUID, visibleToClient bool, limit int, offset int) ([]*ContactEvent, error)

	// FindBySessionID busca todos os eventos de uma sessão.
	FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, offset int) ([]*ContactEvent, error)

	// FindUndeliveredRealtime busca eventos em tempo real não entregues.
	FindUndeliveredRealtime(ctx context.Context, limit int) ([]*ContactEvent, error)

	// FindUndeliveredForContact busca eventos não entregues de um contato específico.
	FindUndeliveredForContact(ctx context.Context, contactID uuid.UUID) ([]*ContactEvent, error)

	// FindByTenantAndType busca eventos por tenant e tipo.
	FindByTenantAndType(ctx context.Context, tenantID string, eventType string, since time.Time, limit int) ([]*ContactEvent, error)

	// FindByCategory busca eventos por categoria.
	FindByCategory(ctx context.Context, tenantID string, category Category, since time.Time, limit int) ([]*ContactEvent, error)

	// FindExpired busca eventos expirados para limpeza.
	FindExpired(ctx context.Context, before time.Time, limit int) ([]*ContactEvent, error)

	// Delete remove um evento.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteExpired remove eventos expirados em batch.
	DeleteExpired(ctx context.Context, before time.Time) (int, error)

	// CountByContact conta eventos de um contato.
	CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)
}
