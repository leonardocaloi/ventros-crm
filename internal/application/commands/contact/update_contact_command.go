package contact

import (
	"github.com/google/uuid"
)

// UpdateContactCommand comando para atualizar um contato existente
type UpdateContactCommand struct {
	ContactID     uuid.UUID
	TenantID      string
	Name          *string
	Email         *string
	Phone         *string
	ExternalID    *string
	SourceChannel *string
	Language      *string
	Timezone      *string
	Tags          []string // Se fornecido, substitui todas as tags
	CustomFields  map[string]string
}

// Validate valida o comando
func (c *UpdateContactCommand) Validate() error {
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	if c.ContactID == uuid.Nil {
		return ErrContactIDRequired
	}

	// Verificar se pelo menos um campo foi fornecido
	if c.Name == nil && c.Email == nil && c.Phone == nil && c.ExternalID == nil &&
		c.SourceChannel == nil && c.Language == nil && c.Timezone == nil &&
		len(c.Tags) == 0 && len(c.CustomFields) == 0 {
		return ErrNoFieldsToUpdate
	}

	return nil
}
