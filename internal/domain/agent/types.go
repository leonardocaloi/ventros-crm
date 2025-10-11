package agent

type Role string

const (
	RoleHumanAgent Role = "human_agent"
	RoleSupervisor Role = "supervisor"
	RoleAdmin      Role = "admin"

	RoleAIAgent       Role = "ai_agent"
	RoleAIAssistant   Role = "ai_assistant"
	RoleChannelBot    Role = "channel_bot"
	RoleWorkflowBot   Role = "workflow_bot"
	RoleAnalyticsBot  Role = "analytics_bot"
	RoleSummarizerBot Role = "summarizer_bot"
)

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

func (r Role) String() string {
	return string(r)
}

func (r Role) IsHuman() bool {
	return r == RoleHumanAgent || r == RoleSupervisor || r == RoleAdmin
}

func (r Role) IsAI() bool {
	return r == RoleAIAgent || r == RoleAIAssistant || r == RoleChannelBot ||
		r == RoleWorkflowBot || r == RoleAnalyticsBot || r == RoleSummarizerBot
}

func (r Role) CanAttendSessions() bool {
	return r == RoleHumanAgent || r == RoleSupervisor || r == RoleAdmin ||
		r == RoleAIAgent || r == RoleChannelBot
}

func (r Role) CanManageAgents() bool {
	return r == RoleAdmin || r == RoleSupervisor
}

func (r Role) CanSendMessages() bool {
	return r == RoleHumanAgent || r == RoleSupervisor || r == RoleAdmin ||
		r == RoleAIAgent || r == RoleAIAssistant || r == RoleChannelBot
}

func (r Role) RequiresAuthentication() bool {
	return r.IsHuman()
}
