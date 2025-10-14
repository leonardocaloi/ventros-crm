package session

import (
	"github.com/google/uuid"
)

// CloseSessionCommand comando para encerrar uma sessão
type CloseSessionCommand struct {
	SessionID uuid.UUID
	Reason    string // "resolved", "transferred", "escalated", "agent_closed"
	Notes     string
}

// Validate valida o comando
func (c *CloseSessionCommand) Validate() error {
	if c.SessionID == uuid.Nil {
		return ErrSessionIDRequired
	}
	if c.Reason == "" {
		return ErrReasonRequired
	}

	// Validate reason
	validReasons := map[string]bool{
		"resolved":     true,
		"transferred":  true,
		"escalated":    true,
		"agent_closed": true,
	}
	if !validReasons[c.Reason] {
		return ErrInvalidReason
	}

	return nil
}
