package session

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/application/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
)

// CloseSessionCommand contém os dados para fechar uma sessão.
type CloseSessionCommand struct {
	SessionID uuid.UUID
	Reason    session.EndReason
}

// CloseSessionUseCase implementa o caso de uso de fechamento de sessão.
type CloseSessionUseCase struct {
	sessionRepo session.Repository
	eventBus    EventBus
	txManager   shared.TransactionManager
}

// NewCloseSessionUseCase cria uma nova instância.
func NewCloseSessionUseCase(
	sessionRepo session.Repository,
	eventBus EventBus,
	txManager shared.TransactionManager,
) *CloseSessionUseCase {
	return &CloseSessionUseCase{
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
		txManager:   txManager,
	}
}

// Execute executa o caso de uso.
func (uc *CloseSessionUseCase) Execute(ctx context.Context, cmd CloseSessionCommand) error {
	// Validação
	if cmd.SessionID == uuid.Nil {
		return errors.New("sessionID is required")
	}

	// Buscar sessão (fora da transação - read-only)
	sess, err := uc.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}

	// Lógica de domínio - encerrar sessão
	if err := sess.End(cmd.Reason); err != nil {
		return err
	}

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Persistir sessão (usa transação do contexto)
		if err := uc.sessionRepo.Save(txCtx, sess); err != nil {
			return err
		}

		// 2. Publicar eventos no outbox (usa mesma transação)
		for _, event := range sess.DomainEvents() {
			if err := uc.eventBus.Publish(txCtx, event); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Limpar eventos após sucesso
	sess.ClearEvents()

	return nil
}
