package session

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/google/uuid"
)

// RecordMessageCommand contém os dados para registrar uma mensagem.
type RecordMessageCommand struct {
	SessionID   uuid.UUID
	FromContact bool
}

// RecordMessageUseCase implementa o caso de uso de registro de mensagem.
type RecordMessageUseCase struct {
	sessionRepo session.Repository
	eventBus    EventBus
}

// NewRecordMessageUseCase cria uma nova instância.
func NewRecordMessageUseCase(
	sessionRepo session.Repository,
	eventBus EventBus,
) *RecordMessageUseCase {
	return &RecordMessageUseCase{
		sessionRepo: sessionRepo,
		eventBus:    eventBus,
	}
}

// Execute executa o caso de uso.
func (uc *RecordMessageUseCase) Execute(ctx context.Context, cmd RecordMessageCommand) error {
	// Validação
	if cmd.SessionID == uuid.Nil {
		return errors.New("sessionID is required")
	}

	// Buscar sessão
	sess, err := uc.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}

	// Lógica de domínio - registrar mensagem
	if err := sess.RecordMessage(cmd.FromContact); err != nil {
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
