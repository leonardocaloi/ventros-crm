package channel

import (
	"fmt"

	"github.com/google/uuid"
)

// ImportHistoryCommand comando para iniciar importação de histórico de mensagens
type ImportHistoryCommand struct {
	ChannelID     uuid.UUID
	TenantID      string
	Strategy      string // "time_range", "full", "recent"
	TimeRangeDays int    // Para strategy="time_range"
	Limit         int    // Limite de mensagens por chat
	UserID        uuid.UUID
}

// Validate valida o comando
func (cmd ImportHistoryCommand) Validate() error {
	if cmd.ChannelID == uuid.Nil {
		return fmt.Errorf("channel_id is required")
	}

	if cmd.TenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	if cmd.Strategy == "" {
		return fmt.Errorf("strategy is required")
	}

	validStrategies := map[string]bool{
		"time_range": true,
		"full":       true,
		"recent":     true,
	}

	if !validStrategies[cmd.Strategy] {
		return fmt.Errorf("invalid strategy: %s (must be time_range, full, or recent)", cmd.Strategy)
	}

	if cmd.Strategy == "time_range" && cmd.TimeRangeDays <= 0 {
		return fmt.Errorf("time_range_days must be > 0 when strategy=time_range")
	}

	return nil
}
