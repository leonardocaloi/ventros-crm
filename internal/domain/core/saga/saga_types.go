package saga

// SagaType define os tipos de Saga implementados no sistema
type SagaType string

const (
	// Fast Path Sagas (Coreografia via RabbitMQ)
	ProcessInboundMessageSaga    SagaType = "process_inbound_message"     // WAHA webhook → contact → session → message
	ChangePipelineStatusSaga     SagaType = "change_pipeline_status"      // Mudar status → trigger automation → create tasks
	CreateContactWithSessionSaga SagaType = "create_contact_with_session" // Criar contact → start session → send welcome
	TrackAdConversionSaga        SagaType = "track_ad_conversion"         // Tracking → contact event → webhooks
	ProcessWAHAWebhookSaga       SagaType = "process_waha_webhook"        // WAHA event → parse → route → process

	// Slow Path Sagas (Orquestração via Temporal) - Futuro
	OnboardCustomerSaga    SagaType = "onboard_customer"      // Contact + Billing + Trial + Email
	RenewSubscriptionSaga  SagaType = "renew_subscription"    // Check payment + Invoice + Update billing
	ImportWAHAHistorySaga  SagaType = "import_waha_history"   // Fetch messages + create contacts + sessions
	ProcessMediaWithAISaga SagaType = "process_media_with_ai" // Download + Process + Extract + Save

	// Compensation Only (eventos de compensação)
	CompensateSaga SagaType = "compensate" // Compensação genérica
)

// String retorna a representação em string do tipo de Saga
func (s SagaType) String() string {
	return string(s)
}

// IsFastPath retorna true se a Saga deve usar Fast Path (Coreografia)
func (s SagaType) IsFastPath() bool {
	switch s {
	case ProcessInboundMessageSaga,
		ChangePipelineStatusSaga,
		CreateContactWithSessionSaga,
		TrackAdConversionSaga,
		ProcessWAHAWebhookSaga:
		return true
	default:
		return false
	}
}

// IsSlowPath retorna true se a Saga deve usar Slow Path (Temporal)
func (s SagaType) IsSlowPath() bool {
	return !s.IsFastPath() && s != CompensateSaga
}

// SagaStep define os steps de cada Saga
type SagaStep string

// ProcessInboundMessageSaga Steps
const (
	StepWAHAReceived        SagaStep = "waha_received"
	StepContactCreated      SagaStep = "contact_created"
	StepContactFound        SagaStep = "contact_found"
	StepSessionStarted      SagaStep = "session_started"
	StepSessionUpdated      SagaStep = "session_updated"
	StepMessageCreated      SagaStep = "message_created"
	StepMessageSaved        SagaStep = "message_saved"
	StepTrackingCreated     SagaStep = "tracking_created"
	StepAIProcessingStarted SagaStep = "ai_processing_started"
	StepWebhooksNotified    SagaStep = "webhooks_notified"
)

// ChangePipelineStatusSaga Steps
const (
	StepStatusChanged       SagaStep = "status_changed"
	StepAutomationTriggered SagaStep = "automation_triggered"
	StepTasksCreated        SagaStep = "tasks_created"
	StepNotificationsSent   SagaStep = "notifications_sent"
)

// CreateContactWithSessionSaga Steps
const (
	StepContactCreating    SagaStep = "contact_creating"
	StepSessionCreating    SagaStep = "session_creating"
	StepWelcomeMessageSent SagaStep = "welcome_message_sent"
)

// TrackAdConversionSaga Steps
const (
	StepTrackingParsed      SagaStep = "tracking_parsed"
	StepContactEventCreated SagaStep = "contact_event_created"
	StepConversionTracked   SagaStep = "conversion_tracked"
)

// Compensation Steps (reverso)
const (
	StepCompensateContact    SagaStep = "compensate_contact"
	StepCompensateSession    SagaStep = "compensate_session"
	StepCompensateMessage    SagaStep = "compensate_message"
	StepCompensateAutomation SagaStep = "compensate_automation"
	StepCompensateTasks      SagaStep = "compensate_tasks"
)

// String retorna a representação em string do step
func (s SagaStep) String() string {
	return string(s)
}

// GetExpectedSteps retorna os steps esperados para cada tipo de Saga
func GetExpectedSteps(sagaType SagaType) []SagaStep {
	switch sagaType {
	case ProcessInboundMessageSaga:
		return []SagaStep{
			StepWAHAReceived,
			StepContactCreated, // ou StepContactFound
			StepSessionStarted, // ou StepSessionUpdated
			StepMessageSaved,
			// StepAIProcessingStarted (opcional)
		}
	case ChangePipelineStatusSaga:
		return []SagaStep{
			StepStatusChanged,
			StepAutomationTriggered,
			StepTasksCreated,
			StepNotificationsSent,
		}
	case CreateContactWithSessionSaga:
		return []SagaStep{
			StepContactCreating,
			StepSessionCreating,
			StepWelcomeMessageSent,
		}
	case TrackAdConversionSaga:
		return []SagaStep{
			StepTrackingParsed,
			StepContactEventCreated,
			StepConversionTracked,
		}
	default:
		return []SagaStep{}
	}
}
