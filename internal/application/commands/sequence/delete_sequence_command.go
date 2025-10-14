package sequence

import "github.com/google/uuid"

// DeleteSequenceCommand comando para deletar uma sequência
type DeleteSequenceCommand struct {
	SequenceID uuid.UUID
	TenantID   string
}

// Validate valida o comando
func (c *DeleteSequenceCommand) Validate() error {
	if c.SequenceID == uuid.Nil {
		return ErrSequenceIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	return nil
}
