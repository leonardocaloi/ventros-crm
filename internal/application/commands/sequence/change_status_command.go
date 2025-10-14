package sequence

import "github.com/google/uuid"

// SequenceStatusAction representa a ação de mudança de status
type SequenceStatusAction string

const (
	StatusActionActivate SequenceStatusAction = "activate"
	StatusActionPause    SequenceStatusAction = "pause"
	StatusActionResume   SequenceStatusAction = "resume"
	StatusActionArchive  SequenceStatusAction = "archive"
)

// ChangeSequenceStatusCommand comando para mudar o status de uma sequência
type ChangeSequenceStatusCommand struct {
	SequenceID uuid.UUID
	TenantID   string
	Action     SequenceStatusAction
}

// Validate valida o comando
func (c *ChangeSequenceStatusCommand) Validate() error {
	if c.SequenceID == uuid.Nil {
		return ErrSequenceIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}

	// Validate action
	validActions := map[SequenceStatusAction]bool{
		StatusActionActivate: true,
		StatusActionPause:    true,
		StatusActionResume:   true,
		StatusActionArchive:  true,
	}
	if !validActions[c.Action] {
		return ErrInvalidSequenceStatus
	}

	return nil
}
