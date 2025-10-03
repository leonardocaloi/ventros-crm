package agent

// Role representa o papel do agente no sistema.
type Role string

const (
	// Agentes Humanos
	RoleHumanAgent Role = "human_agent" // Agente humano que atende chats
	RoleSupervisor Role = "supervisor"  // Supervisor de agentes humanos
	RoleAdmin      Role = "admin"       // Administrador do sistema

	// Agentes de IA
	RoleAIAgent       Role = "ai_agent"       // Agente de IA que atende autonomamente
	RoleAIAssistant   Role = "ai_assistant"   // IA que auxilia agentes humanos
	RoleChannelBot    Role = "channel_bot"    // Bot que fica por trás dos canais
	RoleWorkflowBot   Role = "workflow_bot"   // Bot de automação de workflows
	RoleAnalyticsBot  Role = "analytics_bot"  // Bot que processa analytics/insights
	RoleSummarizerBot Role = "summarizer_bot" // Bot que gera resumos de sessões
)

// IsValid verifica se o role é válido.
func (r Role) IsValid() bool {
	switch r {
	case RoleHumanAgent, RoleSupervisor, RoleAdmin,
		RoleAIAgent, RoleAIAssistant, RoleChannelBot,
		RoleWorkflowBot, RoleAnalyticsBot, RoleSummarizerBot:
		return true
	default:
		return false
	}
}

// String retorna a representação em string do role.
func (r Role) String() string {
	return string(r)
}

// IsHuman verifica se é um agente humano.
func (r Role) IsHuman() bool {
	return r == RoleHumanAgent || r == RoleSupervisor || r == RoleAdmin
}

// IsAI verifica se é um agente de IA/bot.
func (r Role) IsAI() bool {
	return r == RoleAIAgent || r == RoleAIAssistant || r == RoleChannelBot ||
		r == RoleWorkflowBot || r == RoleAnalyticsBot || r == RoleSummarizerBot
}

// CanAttendSessions verifica se o agente pode atender sessões diretamente.
func (r Role) CanAttendSessions() bool {
	return r == RoleHumanAgent || r == RoleSupervisor || r == RoleAdmin ||
		r == RoleAIAgent || r == RoleChannelBot
}

// CanManageAgents verifica se o role pode gerenciar outros agentes.
func (r Role) CanManageAgents() bool {
	return r == RoleAdmin || r == RoleSupervisor
}

// CanSendMessages verifica se o agente pode enviar mensagens.
func (r Role) CanSendMessages() bool {
	return r == RoleHumanAgent || r == RoleSupervisor || r == RoleAdmin ||
		r == RoleAIAgent || r == RoleAIAssistant || r == RoleChannelBot
}

// RequiresAuthentication verifica se o agente precisa de autenticação humana.
func (r Role) RequiresAuthentication() bool {
	return r.IsHuman()
}
