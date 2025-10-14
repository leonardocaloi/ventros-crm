package contact

import (
	"github.com/google/uuid"
)

// CreateContactCommand comando para criar um novo contato
type CreateContactCommand struct {
	ProjectID     uuid.UUID
	TenantID      string
	Name          string
	Email         string
	Phone         string
	ExternalID    string
	SourceChannel string
	Language      string
	Timezone      string
	Tags          []string
	CustomFields  map[string]string
}

// Validate valida o comando
func (c *CreateContactCommand) Validate() error {
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	if c.ProjectID == uuid.Nil {
		return ErrProjectIDRequired
	}
	if c.Name == "" {
		return ErrContactNameRequired
	}
	return nil
}
