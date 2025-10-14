package sequence

import "github.com/google/uuid"

// UpdateSequenceCommand comando para atualizar uma sequência
type UpdateSequenceCommand struct {
	SequenceID  uuid.UUID
	TenantID    string
	Name        *string
	Description *string
	ExitOnReply *bool
}

// Validate valida o comando
func (c *UpdateSequenceCommand) Validate() error {
	if c.SequenceID == uuid.Nil {
		return ErrSequenceIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}

	// At least one field must be provided
	if c.Name == nil && c.Description == nil && c.ExitOnReply == nil {
		return ErrNoFieldsToUpdate
	}

	return nil
}
