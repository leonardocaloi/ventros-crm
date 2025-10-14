package channel

import (
	"github.com/google/uuid"
)

// ActivateChannelCommand comando para ativar um canal
// A ativação é assíncrona: o comando muda status para "activating" e publica evento
// Um worker assíncrono consumirá o evento e executará a validação específica do tipo
type ActivateChannelCommand struct {
	ChannelID uuid.UUID
	TenantID  string
}

// Validate valida o comando
func (c *ActivateChannelCommand) Validate() error {
	if c.ChannelID == uuid.Nil {
		return ErrChannelIDRequired
	}
	if c.TenantID == "" {
		return ErrTenantIDRequired
	}
	return nil
}
