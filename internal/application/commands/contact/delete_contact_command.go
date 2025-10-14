package contact

import (
	"github.com/google/uuid"
)

// DeleteContactCommand comando para deletar (soft delete) um contato
type DeleteContactCommand struct {
	ContactID uuid.UUID
	TenantID  string
}

// Validate valida o comando
func (c *DeleteContactCommand) Validate() error {
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	if c.ContactID == uuid.Nil {
		return ErrContactIDRequired
	}
	return nil
}
