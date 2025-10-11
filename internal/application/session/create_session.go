package session

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/internal/application/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
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
	txManager   shared.TransactionManager
}

// NewCreateSessionUseCase cria uma nova instância.
func NewCreateSessionUseCase(
	sessionRepo session.Repository,
	eventBus EventBus,
	txManager shared.TransactionManager,
) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
		txManager:   txManager,
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

	// Verificar se já existe sessão ativa (fora da transação - read-only)
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

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	var sessionID uuid.UUID
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Persistir sessão (usa transação do contexto)
		if err := uc.sessionRepo.Save(txCtx, newSession); err != nil {
			return err
		}

		// 2. Publicar eventos no outbox (usa mesma transação)
		for _, event := range newSession.DomainEvents() {
			if err := uc.eventBus.Publish(txCtx, event); err != nil {
				return err
			}
		}

		sessionID = newSession.ID()
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Limpar eventos após sucesso
	newSession.ClearEvents()

	return &CreateSessionResult{
		SessionID: sessionID,
		Created:   true,
	}, nil
}

// isNotFoundError verifica se é erro de "não encontrado".
// Implementação depende do repositório (pode usar errors.Is).
func isNotFoundError(err error) bool {
	// TODO: implementar conforme seu padrão de erros
	return err.Error() == "not found"
}
