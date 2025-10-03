package session

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/session"
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
}

// NewCloseSessionUseCase cria uma nova instância.
func NewCloseSessionUseCase(
	sessionRepo session.Repository,
	eventBus EventBus,
) *CloseSessionUseCase {
	return &CloseSessionUseCase{
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
	}
}

// Execute executa o caso de uso.
func (uc *CloseSessionUseCase) Execute(ctx context.Context, cmd CloseSessionCommand) error {
	// Validação
	if cmd.SessionID == uuid.Nil {
		return errors.New("sessionID is required")
	}

	// Buscar sessão
	sess, err := uc.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}

	// Lógica de domínio - encerrar sessão
	if err := sess.End(cmd.Reason); err != nil {
		return err
	}

	// Persistir
	if err := uc.sessionRepo.Save(ctx, sess); err != nil {
		return err
	}

	// Publicar eventos
	for _, event := range sess.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			// Log error
		}
	}

	sess.ClearEvents()

	return nil
}
