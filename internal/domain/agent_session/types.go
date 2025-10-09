package agent_session

// RoleInSession define o papel de um agente dentro de uma sessão específica.
type RoleInSession string

const (
	// Papéis primários
	RolePrimary  RoleInSession = "primary"  // Agente principal da sessão
	RoleSupport  RoleInSession = "support"  // Agente de suporte
	RoleObserver RoleInSession = "observer" // Apenas observando (supervisor, analytics)
	RoleHandoff  RoleInSession = "handoff"  // Recebendo transferência

	// Papéis de IA (Google ADK)
	RoleAIAssistant RoleInSession = "ai_assistant" // AI assistindo agente humano
	RoleAIPrimary   RoleInSession = "ai_primary"   // AI como agente principal
	RoleBot         RoleInSession = "bot"          // Bot de canal
)

// IsValid verifica se o papel é válido.
func (r RoleInSession) IsValid() bool {
	switch r {
	case RolePrimary, RoleSupport, RoleObserver, RoleHandoff,
		RoleAIAssistant, RoleAIPrimary, RoleBot:
		return true
	default:
		return false
	}
}

// String retorna a string representation.
func (r RoleInSession) String() string {
	return string(r)
}
