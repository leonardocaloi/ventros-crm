package automation

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/note"
	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/google/uuid"
)

// CreateNoteExecutor implementa a ação de criar nota
type CreateNoteExecutor struct {
	noteRepository note.NoteRepository
}

// NewCreateNoteExecutor cria um novo executor de criação de notas
func NewCreateNoteExecutor(noteRepository note.NoteRepository) *CreateNoteExecutor {
	return &CreateNoteExecutor{
		noteRepository: noteRepository,
	}
}

// Type retorna o tipo de ação
func (e *CreateNoteExecutor) Type() pipeline.AutomationAction {
	return pipeline.ActionCreateNote
}

// Validate valida os parâmetros da ação
func (e *CreateNoteExecutor) Validate(params map[string]interface{}) error {
	// Valida entity_type
	entityType, ok := params["entity_type"].(string)
	if !ok || entityType == "" {
		return fmt.Errorf("%w: entity_type", pipeline.ErrMissingRequiredParam)
	}

	if entityType != "agent" && entityType != "contact" && entityType != "session" {
		return fmt.Errorf("invalid entity_type: must be agent, contact, or session")
	}

	// Valida entity_id
	entityIDStr, ok := params["entity_id"].(string)
	if !ok || entityIDStr == "" {
		return fmt.Errorf("%w: entity_id", pipeline.ErrMissingRequiredParam)
	}

	if _, err := uuid.Parse(entityIDStr); err != nil {
		return fmt.Errorf("invalid entity_id: must be a valid UUID")
	}

	// Valida content
	content, ok := params["content"].(string)
	if !ok || content == "" {
		return fmt.Errorf("%w: content", pipeline.ErrMissingRequiredParam)
	}

	return nil
}

// Execute executa a criação da nota
func (e *CreateNoteExecutor) Execute(ctx context.Context, params pipeline.ActionExecutionParams) error {
	// Extrai parâmetros
	entityType := params.Action.Params["entity_type"].(string)
	entityIDStr := params.Action.Params["entity_id"].(string)
	content := params.Action.Params["content"].(string)

	entityID, _ := uuid.Parse(entityIDStr)

	// Título opcional
	title := "Automation Note"
	if t, ok := params.Action.Params["title"].(string); ok && t != "" {
		title = t
	}

	// TODO: Interpolar variáveis no content e title
	// content = interpolateVariables(content, params.Variables)
	// title = interpolateVariables(title, params.Variables)

	// Cria a nota usando o domain
	var noteEntity *note.Note
	var err error

	switch entityType {
	case "agent":
		noteEntity, err = note.NewAgentNote(
			params.TenantID,
			entityID,
			title,
			content,
		)
	case "contact":
		noteEntity, err = note.NewContactNote(
			params.TenantID,
			entityID,
			title,
			content,
		)
	case "session":
		noteEntity, err = note.NewSessionNote(
			params.TenantID,
			entityID,
			title,
			content,
		)
	default:
		return fmt.Errorf("unsupported entity_type: %s", entityType)
	}

	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	// Persiste a nota
	if err := e.noteRepository.Save(noteEntity); err != nil {
		return fmt.Errorf("failed to save note: %w", err)
	}

	return nil
}
