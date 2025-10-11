package event

type EventSource string

const (
	EventSourceSystem   EventSource = "system"
	EventSourceWebhook  EventSource = "webhook"
	EventSourceManual   EventSource = "manual"
	EventSourceCron     EventSource = "cron"
	EventSourceWorkflow EventSource = "workflow"
)

func (s EventSource) IsValid() bool {
	switch s {
	case EventSourceSystem, EventSourceWebhook, EventSourceManual,
		EventSourceCron, EventSourceWorkflow:
		return true
	default:
		return false
	}
}

func (s EventSource) String() string {
	return string(s)
}

const (
	EventTypeContactCreated = "contact.created"
	EventTypeContactUpdated = "contact.updated"
	EventTypeContactDeleted = "contact.deleted"

	EventTypeSessionStarted    = "session.started"
	EventTypeSessionEnded      = "session.ended"
	EventTypeSessionSummarized = "session.summarized"

	EventTypeMessageReceived = "message.received"
	EventTypeMessageSent     = "message.sent"
	EventTypeMessageRead     = "message.read"
	EventTypeMessageFailed   = "message.failed"

	EventTypeAgentAssigned = "agent.assigned"
	EventTypeAgentTransfer = "agent.transfer"
	EventTypeAgentTyping   = "agent.typing"

	EventTypeCustomFieldUpdated = "custom_field.updated"
	EventTypeTagAdded           = "tag.added"
	EventTypeTagRemoved         = "tag.removed"
	EventTypeNoteAdded          = "note.added"
)
