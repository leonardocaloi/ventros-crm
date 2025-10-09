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
	CategoryStatus       Category = "status"
	CategoryPipeline     Category = "pipeline"
	CategoryAssignment   Category = "assignment"
	CategoryTag          Category = "tag"
	CategoryNote         Category = "note"
	CategorySession      Category = "session"
	CategoryCustomField  Category = "custom_field"
	CategorySystem       Category = "system"
	CategoryNotification Category = "notification"
	CategoryTracking     Category = "tracking"
)

// IsValid verifica se a categoria é válida.
func (c Category) IsValid() bool {
	switch c {
	case CategoryGeneral, CategoryStatus, CategoryPipeline,
		CategoryAssignment, CategoryTag, CategoryNote,
		CategorySession, CategoryCustomField,
		CategorySystem, CategoryNotification, CategoryTracking:
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
	// Eventos de Status e Pipeline
	EventTypeStatusChanged        = "status_changed"
	EventTypeEnteredPipeline      = "entered_pipeline"
	EventTypeExitedPipeline       = "exited_pipeline"
	EventTypePipelineStageChanged = "pipeline_stage_changed"

	// Eventos de Atribuição
	EventTypeAgentAssigned    = "agent_assigned"
	EventTypeAgentTransferred = "agent_transferred"
	EventTypeAgentUnassigned  = "agent_unassigned"

	// Eventos de Tags
	EventTypeTagAdded   = "tag_added"
	EventTypeTagRemoved = "tag_removed"

	// Eventos de Notas e Anotações
	EventTypeNoteAdded   = "note_added"
	EventTypeNoteUpdated = "note_updated"
	EventTypeNoteDeleted = "note_deleted"

	// Eventos de Sessão
	EventTypeSessionStarted = "session_started"
	EventTypeSessionEnded   = "session_ended"

	// Eventos de Campos Customizados
	EventTypeCustomFieldSet     = "custom_field_set"
	EventTypeCustomFieldCleared = "custom_field_cleared"

	// Eventos de Sistema
	EventTypeWebhookReceived  = "webhook_received"
	EventTypeNotificationSent = "notification_sent"
	EventTypeContactCreated   = "contact_created"
	EventTypeContactUpdated   = "contact_updated"
	EventTypeContactMerged    = "contact_merged"
	EventTypeContactEnriched  = "contact_enriched"
)
