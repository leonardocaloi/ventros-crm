package event

// EventSource representa a origem de um evento.
type EventSource string

const (
	EventSourceSystem   EventSource = "system"
	EventSourceWebhook  EventSource = "webhook"
	EventSourceManual   EventSource = "manual"
	EventSourceCron     EventSource = "cron"
	EventSourceWorkflow EventSource = "workflow"
)

// IsValid verifica se o source é válido.
func (s EventSource) IsValid() bool {
	switch s {
	case EventSourceSystem, EventSourceWebhook, EventSourceManual,
		EventSourceCron, EventSourceWorkflow:
		return true
	default:
		return false
	}
}

// String retorna a representação em string do source.
func (s EventSource) String() string {
	return string(s)
}

// Tipos de eventos comuns (constantes para referência)
const (
	// Contact events
	EventTypeContactCreated = "contact.created"
	EventTypeContactUpdated = "contact.updated"
	EventTypeContactDeleted = "contact.deleted"

	// Session events
	EventTypeSessionStarted    = "session.started"
	EventTypeSessionEnded      = "session.ended"
	EventTypeSessionSummarized = "session.summarized"

	// Message events
	EventTypeMessageReceived = "message.received"
	EventTypeMessageSent     = "message.sent"
	EventTypeMessageRead     = "message.read"
	EventTypeMessageFailed   = "message.failed"

	// Agent events
	EventTypeAgentAssigned = "agent.assigned"
	EventTypeAgentTransfer = "agent.transfer"
	EventTypeAgentTyping   = "agent.typing"

	// Custom events
	EventTypeCustomFieldUpdated = "custom_field.updated"
	EventTypeTagAdded           = "tag.added"
	EventTypeTagRemoved         = "tag.removed"
	EventTypeNoteAdded          = "note.added"
)
