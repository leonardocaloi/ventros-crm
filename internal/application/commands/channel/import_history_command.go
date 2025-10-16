package channel

import (
	"fmt"

	"github.com/google/uuid"
)

// ImportHistoryCommand comando para iniciar importação de histórico de mensagens
type ImportHistoryCommand struct {
	ChannelID             uuid.UUID
	TenantID              string
	Strategy              string // "time_range", "full", "recent"
	TimeRangeDays         int    // para strategy "time_range"
	Limit                 int    // limite de mensagens por chat (0 = todas)
	SessionTimeoutMinutes int    // timeout para agrupar sessões (0 = usar default do canal)
	UserID                uuid.UUID
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
		"all":        true, // Alias for "full" (import all available history)
		"maximum":    true, // Alias for "full" (import maximum available)
	}

	if !validStrategies[cmd.Strategy] {
		return fmt.Errorf("invalid strategy: %s (must be time_range, full, recent, all, or maximum)", cmd.Strategy)
	}

	if cmd.Strategy == "time_range" && cmd.TimeRangeDays <= 0 {
		return fmt.Errorf("time_range_days must be > 0 when strategy=time_range")
	}

	return nil
}
