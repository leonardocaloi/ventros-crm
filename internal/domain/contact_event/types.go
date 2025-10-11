package contact_event

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func (p Priority) String() string {
	return string(p)
}

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

func (c Category) String() string {
	return string(c)
}

type Source string

const (
	SourceSystem      Source = "system"
	SourceAgent       Source = "agent"
	SourceWebhook     Source = "webhook"
	SourceWorkflow    Source = "workflow"
	SourceAutomation  Source = "automation"
	SourceIntegration Source = "integration"
)

func (s Source) IsValid() bool {
	switch s {
	case SourceSystem, SourceAgent, SourceWebhook,
		SourceWorkflow, SourceAutomation, SourceIntegration:
		return true
	default:
		return false
	}
}

func (s Source) String() string {
	return string(s)
}

const (
	EventTypeStatusChanged        = "status_changed"
	EventTypeEnteredPipeline      = "entered_pipeline"
	EventTypeExitedPipeline       = "exited_pipeline"
	EventTypePipelineStageChanged = "pipeline_stage_changed"

	EventTypeAgentAssigned    = "agent_assigned"
	EventTypeAgentTransferred = "agent_transferred"
	EventTypeAgentUnassigned  = "agent_unassigned"

	EventTypeTagAdded   = "tag_added"
	EventTypeTagRemoved = "tag_removed"

	EventTypeNoteAdded   = "note_added"
	EventTypeNoteUpdated = "note_updated"
	EventTypeNoteDeleted = "note_deleted"

	EventTypeSessionStarted = "session_started"
	EventTypeSessionEnded   = "session_ended"

	EventTypeCustomFieldSet     = "custom_field_set"
	EventTypeCustomFieldCleared = "custom_field_cleared"

	EventTypeWebhookReceived  = "webhook_received"
	EventTypeNotificationSent = "notification_sent"
	EventTypeContactCreated   = "contact_created"
	EventTypeContactUpdated   = "contact_updated"
	EventTypeContactMerged    = "contact_merged"
	EventTypeContactEnriched  = "contact_enriched"
)
