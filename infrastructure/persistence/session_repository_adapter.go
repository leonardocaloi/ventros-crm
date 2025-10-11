package persistence

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/google/uuid"
)

// SessionRepositoryAdapter adapta o session.Repository do domínio
// adicionando métodos convenientes para comandos/queries
type SessionRepositoryAdapter struct {
	repo session.Repository
}

// NewSessionRepositoryAdapter cria um novo adapter
func NewSessionRepositoryAdapter(repo session.Repository) *SessionRepositoryAdapter {
	return &SessionRepositoryAdapter{repo: repo}
}

// GetActiveSessionByContact busca a sessão ativa de um contato
func (a *SessionRepositoryAdapter) GetActiveSessionByContact(ctx context.Context, contactID uuid.UUID) (*session.Session, error) {
	// Usa o repositório GORM que já implementa essa lógica
	if gormRepo, ok := a.repo.(*GormSessionRepository); ok {
		return gormRepo.FindActiveByContact(ctx, contactID, nil)
	}

	// Fallback: não encontra
	return nil, nil
}

// Save persiste uma sessão
func (a *SessionRepositoryAdapter) Save(ctx context.Context, sess *session.Session) error {
	return a.repo.Save(ctx, sess)
}
