package note

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/internal/domain/crm/note"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TransactionManager gerencia transações de banco de dados.
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// CreateNoteUseCase cria uma nova nota
type CreateNoteUseCase struct {
	noteRepo  note.Repository
	eventBus  *messaging.DomainEventBus
	logger    *zap.Logger
	txManager TransactionManager
}

// NewCreateNoteUseCase cria uma nova instância do use case
func NewCreateNoteUseCase(
	noteRepo note.Repository,
	eventBus *messaging.DomainEventBus,
	logger *zap.Logger,
	txManager TransactionManager,
) *CreateNoteUseCase {
	return &CreateNoteUseCase{
		noteRepo:  noteRepo,
		eventBus:  eventBus,
		logger:    logger,
		txManager: txManager,
	}
}

// CreateNoteCommand comando para criar nota
type CreateNoteCommand struct {
	ContactID       uuid.UUID
	SessionID       *uuid.UUID
	TenantID        string
	AuthorID        uuid.UUID
	AuthorType      note.AuthorType
	AuthorName      string
	Content         string
	NoteType        note.NoteType
	Priority        note.Priority
	VisibleToClient bool
	Tags            []string
	Mentions        []uuid.UUID
	Attachments     []string
}

// Execute executa o use case
func (uc *CreateNoteUseCase) Execute(ctx context.Context, cmd CreateNoteCommand) (*note.Note, error) {
	// 1. Criar nota
	n, err := note.NewNote(
		cmd.ContactID,
		cmd.TenantID,
		cmd.AuthorID,
		cmd.AuthorType,
		cmd.AuthorName,
		cmd.Content,
		cmd.NoteType,
	)
	if err != nil {
		uc.logger.Error("Failed to create note",
			zap.Error(err),
			zap.String("contact_id", cmd.ContactID.String()))
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	// 2. Configurar campos opcionais
	if cmd.SessionID != nil {
		n.AttachToSession(*cmd.SessionID)
	}

	if cmd.Priority != "" {
		n.SetPriority(cmd.Priority)
	}

	n.SetVisibility(cmd.VisibleToClient)

	for _, tag := range cmd.Tags {
		n.AddTag(tag)
	}

	for _, mention := range cmd.Mentions {
		n.MentionAgent(mention)
	}

	for _, attachment := range cmd.Attachments {
		n.AddAttachment(attachment)
	}

	// 3-4. ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 3. Salvar nota (usa transação do contexto)
		if err := uc.noteRepo.Save(txCtx, n); err != nil {
			uc.logger.Error("Failed to save note",
				zap.Error(err),
				zap.String("note_id", n.ID().String()))
			return fmt.Errorf("failed to save note: %w", err)
		}

		// 4. Publicar eventos de domínio (usa mesma transação)
		if uc.eventBus != nil {
			for _, event := range n.DomainEvents() {
				if err := uc.eventBus.Publish(txCtx, event); err != nil {
					uc.logger.Error("Failed to publish domain event",
						zap.Error(err),
						zap.String("note_id", n.ID().String()))
					return fmt.Errorf("failed to publish event: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	n.ClearEvents()

	uc.logger.Info("Note created successfully",
		zap.String("note_id", n.ID().String()),
		zap.String("contact_id", cmd.ContactID.String()))

	return n, nil
}
