package session

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/google/uuid"
)

// CreateSessionCommand contém os dados para criar uma sessão.
type CreateSessionCommand struct {
	ContactID     uuid.UUID
	TenantID      string
	ChannelTypeID *int
	Timeout       time.Duration
}

// CreateSessionResult retorna o resultado da criação.
type CreateSessionResult struct {
	SessionID uuid.UUID
	Created   bool // false se sessão ativa já existia
}

// CreateSessionUseCase implementa o caso de uso de criação de sessão.
type CreateSessionUseCase struct {
	sessionRepo session.Repository
	eventBus    EventBus
}

// NewCreateSessionUseCase cria uma nova instância.
func NewCreateSessionUseCase(
	sessionRepo session.Repository,
	eventBus EventBus,
) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
	}
}

// Execute executa o caso de uso.
func (uc *CreateSessionUseCase) Execute(ctx context.Context, cmd CreateSessionCommand) (*CreateSessionResult, error) {
	// Validação
	if cmd.ContactID == uuid.Nil {
		return nil, errors.New("contactID is required")
	}
	if cmd.TenantID == "" {
		return nil, errors.New("tenantID is required")
	}

	// Verificar se já existe sessão ativa
	activeSession, err := uc.sessionRepo.FindActiveByContact(ctx, cmd.ContactID, cmd.ChannelTypeID)
	if err != nil && !isNotFoundError(err) {
		return nil, err
	}

	// Se já existe sessão ativa, retorna ela
	if activeSession != nil {
		return &CreateSessionResult{
			SessionID: activeSession.ID(),
			Created:   false,
		}, nil
	}

	// Criar nova sessão
	newSession, err := session.NewSession(
		cmd.ContactID,
		cmd.TenantID,
		cmd.ChannelTypeID,
		cmd.Timeout,
	)
	if err != nil {
		return nil, err
	}

	// Persistir
	if err := uc.sessionRepo.Save(ctx, newSession); err != nil {
		return nil, err
	}

	// Publicar eventos de domínio
	for _, event := range newSession.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail
			// Em produção, usar logging adequado
		}
	}

	newSession.ClearEvents()

	return &CreateSessionResult{
		SessionID: newSession.ID(),
		Created:   true,
	}, nil
}

// isNotFoundError verifica se é erro de "não encontrado".
// Implementação depende do repositório (pode usar errors.Is).
func isNotFoundError(err error) bool {
	// TODO: implementar conforme seu padrão de erros
	return err.Error() == "not found"
}
