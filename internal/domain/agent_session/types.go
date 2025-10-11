package agent_session

type RoleInSession string

const (
	RolePrimary  RoleInSession = "primary"
	RoleSupport  RoleInSession = "support"
	RoleObserver RoleInSession = "observer"
	RoleHandoff  RoleInSession = "handoff"

	RoleAIAssistant RoleInSession = "ai_assistant"
	RoleAIPrimary   RoleInSession = "ai_primary"
	RoleBot         RoleInSession = "bot"
)

func (r RoleInSession) IsValid() bool {
	switch r {
	case RolePrimary, RoleSupport, RoleObserver, RoleHandoff,
		RoleAIAssistant, RoleAIPrimary, RoleBot:
		return true
	default:
		return false
	}
}

func (r RoleInSession) String() string {
	return string(r)
}
