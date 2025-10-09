package note

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/messaging"
	"github.com/caloi/ventros-crm/internal/domain/note"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateNoteUseCase cria uma nova nota
type CreateNoteUseCase struct {
	noteRepo note.Repository
	eventBus *messaging.DomainEventBus
	logger   *zap.Logger
}

// NewCreateNoteUseCase cria uma nova instância do use case
func NewCreateNoteUseCase(
	noteRepo note.Repository,
	eventBus *messaging.DomainEventBus,
	logger *zap.Logger,
) *CreateNoteUseCase {
	return &CreateNoteUseCase{
		noteRepo: noteRepo,
		eventBus: eventBus,
		logger:   logger,
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

	// 3. Salvar nota
	if err := uc.noteRepo.Save(ctx, n); err != nil {
		uc.logger.Error("Failed to save note",
			zap.Error(err),
			zap.String("note_id", n.ID().String()))
		return nil, fmt.Errorf("failed to save note: %w", err)
	}

	// 4. Publicar eventos de domínio
	for _, event := range n.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			uc.logger.Error("Failed to publish domain event",
				zap.Error(err),
				zap.String("note_id", n.ID().String()))
			// Não retorna erro, pois a nota já foi salva
		}
	}
	n.ClearEvents()

	uc.logger.Info("Note created successfully",
		zap.String("note_id", n.ID().String()),
		zap.String("contact_id", cmd.ContactID.String()))

	return n, nil
}
