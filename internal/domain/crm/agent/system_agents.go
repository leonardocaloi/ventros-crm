package agent

import "github.com/google/uuid"

// System Agent UUIDs - Range reservado: 00000000-0000-0000-0000-0000000000XX (00-99)
// Estes agents são criados automaticamente via migration e são imutáveis
var (
	SystemAgentBroadcast = uuid.MustParse("00000000-0000-0000-0000-000000000001") // Campanhas broadcast
	SystemAgentSequence  = uuid.MustParse("00000000-0000-0000-0000-000000000002") // Sequências de automação
	SystemAgentTrigger   = uuid.MustParse("00000000-0000-0000-0000-000000000003") // Triggers/regras de pipeline
	SystemAgentWebhook   = uuid.MustParse("00000000-0000-0000-0000-000000000004") // Respostas via webhook
	SystemAgentScheduled = uuid.MustParse("00000000-0000-0000-0000-000000000005") // Mensagens agendadas
	SystemAgentTest      = uuid.MustParse("00000000-0000-0000-0000-000000000010") // Testes E2E e envios de teste
	SystemAgentDefault   = uuid.MustParse("00000000-0000-0000-0000-000000000099") // Fallback genérico
)

// IsSystemAgentID verifica se o UUID está no range reservado de system agents (00-99)
func IsSystemAgentID(id uuid.UUID) bool {
	bytes := id[:]

	// Primeiros 15 bytes devem ser zero
	for i := 0; i < 15; i++ {
		if bytes[i] != 0 {
			return false
		}
	}

	// Último byte deve estar entre 0-99
	return bytes[15] <= 99
}

// ValidSystemAgentIDs retorna lista de todos os system agents válidos
func ValidSystemAgentIDs() []uuid.UUID {
	return []uuid.UUID{
		SystemAgentBroadcast,
		SystemAgentSequence,
		SystemAgentTrigger,
		SystemAgentWebhook,
		SystemAgentScheduled,
		SystemAgentTest,
		SystemAgentDefault,
	}
}

// GetSystemAgentName retorna o nome do system agent baseado no ID
func GetSystemAgentName(id uuid.UUID) string {
	switch id {
	case SystemAgentBroadcast:
		return "System - Broadcast"
	case SystemAgentSequence:
		return "System - Sequence"
	case SystemAgentTrigger:
		return "System - Trigger"
	case SystemAgentWebhook:
		return "System - Webhook"
	case SystemAgentScheduled:
		return "System - Scheduled"
	case SystemAgentTest:
		return "System - Test"
	case SystemAgentDefault:
		return "System - Default"
	default:
		return "Unknown System Agent"
	}
}
