package sequence

import "github.com/google/uuid"

// EnrollContactCommand comando para inscrever um contato em uma sequência
type EnrollContactCommand struct {
	SequenceID uuid.UUID
	ContactID  uuid.UUID
	TenantID   string
}

// Validate valida o comando
func (c *EnrollContactCommand) Validate() error {
	if c.SequenceID == uuid.Nil {
		return ErrSequenceIDRequired
	}
	if c.ContactID == uuid.Nil {
		return ErrContactIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	return nil
}
