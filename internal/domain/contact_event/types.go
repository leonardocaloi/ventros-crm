package contact_event

// Priority representa a prioridade de entrega de um evento.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// IsValid verifica se a prioridade é válida.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// String retorna a representação em string da prioridade.
func (p Priority) String() string {
	return string(p)
}

// Category representa a categoria de um evento.
type Category string

const (
	CategoryGeneral      Category = "general"
	CategoryMessage      Category = "message"
	CategoryStatus       Category = "status"
	CategoryNote         Category = "note"
	CategorySystem       Category = "system"
	CategoryNotification Category = "notification"
)

// IsValid verifica se a categoria é válida.
func (c Category) IsValid() bool {
	switch c {
	case CategoryGeneral, CategoryMessage, CategoryStatus, 
		CategoryNote, CategorySystem, CategoryNotification:
		return true
	default:
		return false
	}
}

// String retorna a representação em string da categoria.
func (c Category) String() string {
	return string(c)
}

// Source representa a origem de um evento.
type Source string

const (
	SourceSystem      Source = "system"
	SourceAgent       Source = "agent"
	SourceWebhook     Source = "webhook"
	SourceWorkflow    Source = "workflow"
	SourceAutomation  Source = "automation"
	SourceIntegration Source = "integration"
)

// IsValid verifica se a origem é válida.
func (s Source) IsValid() bool {
	switch s {
	case SourceSystem, SourceAgent, SourceWebhook, 
		SourceWorkflow, SourceAutomation, SourceIntegration:
		return true
	default:
		return false
	}
}

// String retorna a representação em string da origem.
func (s Source) String() string {
	return string(s)
}

// Tipos de eventos comuns para ContactEvent
const (
	EventTypeMessageReceived  = "message_received"
	EventTypeMessageSent      = "message_sent"
	EventTypeStatusChanged    = "status_changed"
	EventTypeNoteAdded        = "note_added"
	EventTypeTagAdded         = "tag_added"
	EventTypeTagRemoved       = "tag_removed"
	EventTypeAgentAssigned    = "agent_assigned"
	EventTypeAgentTransferred = "agent_transferred"
	EventTypeSessionStarted   = "session_started"
	EventTypeSessionEnded     = "session_ended"
	EventTypeCustomFieldSet   = "custom_field_set"
	EventTypeWebhookReceived  = "webhook_received"
	EventTypeNotificationSent = "notification_sent"
)
