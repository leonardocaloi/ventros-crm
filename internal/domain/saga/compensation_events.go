package saga

// CompensationEventType define os tipos de eventos de compensação.
// Eventos de compensação são disparados quando uma Saga falha e precisa
// desfazer operações já realizadas (ordem reversa - LIFO).
type CompensationEventType string

const (
	// Contact compensation events
	CompensateContactCreated         CompensationEventType = "compensate.contact.created"           // Deleta contato criado
	CompensateContactUpdated         CompensationEventType = "compensate.contact.updated"           // Reverte atualização de contato
	CompensateContactMerged          CompensationEventType = "compensate.contact.merged"            // Desfaz merge de contatos
	CompensateContactPipelineChanged CompensationEventType = "compensate.contact.pipeline_changed"  // Reverte mudança de pipeline

	// Session compensation events
	CompensateSessionStarted  CompensationEventType = "compensate.session.started"   // Encerra sessão criada
	CompensateSessionAssigned CompensationEventType = "compensate.session.assigned"  // Remove agente atribuído
	CompensateSessionEscalated CompensationEventType = "compensate.session.escalated" // Reverte escalação

	// Message compensation events
	CompensateMessageCreated CompensationEventType = "compensate.message.created" // Deleta mensagem criada
	CompensateMessageSent    CompensationEventType = "compensate.message.sent"    // Marca mensagem como falhada

	// Pipeline compensation events
	CompensatePipelineCreated        CompensationEventType = "compensate.pipeline.created"         // Deleta pipeline criado
	CompensateStatusCreated          CompensationEventType = "compensate.status.created"           // Deleta status criado
	CompensatePipelineStatusChanged  CompensationEventType = "compensate.pipeline.status_changed"  // Reverte mudança de status

	// Note compensation events
	CompensateNoteAdded   CompensationEventType = "compensate.note.added"   // Deleta nota criada
	CompensateNoteUpdated CompensationEventType = "compensate.note.updated" // Reverte atualização de nota

	// Channel compensation events
	CompensateChannelCreated   CompensationEventType = "compensate.channel.created"   // Deleta canal criado
	CompensateChannelActivated CompensationEventType = "compensate.channel.activated" // Desativa canal

	// Tracking compensation events
	CompensateTrackingCreated CompensationEventType = "compensate.tracking.created" // Deleta tracking criado

	// Automation compensation events
	CompensateAutomationExecuted CompensationEventType = "compensate.automation.executed" // Reverte ação de automação

	// Agent compensation events
	CompensateAgentCreated        CompensationEventType = "compensate.agent.created"         // Deleta agente criado
	CompensateAgentPermissionGranted CompensationEventType = "compensate.agent.permission_granted" // Revoga permissão concedida

	// Billing compensation events
	CompensateBillingAccountCreated CompensationEventType = "compensate.billing.account_created" // Cancela conta de billing criada
	CompensatePaymentMethodActivated CompensationEventType = "compensate.billing.payment_method_activated" // Desativa método de pagamento
)

// CompensationMapping mapeia eventos de domínio para seus eventos de compensação.
// Usado pelo SagaCoordinator para determinar qual evento de compensação disparar.
var CompensationMapping = map[string]CompensationEventType{
	// Contact events
	"contact.created":                CompensateContactCreated,
	"contact.updated":                CompensateContactUpdated,
	"contact.merged":                 CompensateContactMerged,
	"contact.pipeline_status_changed": CompensateContactPipelineChanged,

	// Session events
	"session.started":    CompensateSessionStarted,
	"session.agent_assigned": CompensateSessionAssigned,
	"session.escalated":  CompensateSessionEscalated,

	// Message events
	"message.created": CompensateMessageCreated,
	"message.sent":    CompensateMessageSent,

	// Pipeline events
	"pipeline.created":              CompensatePipelineCreated,
	"status.created":                CompensateStatusCreated,
	"contact.status_changed":        CompensatePipelineStatusChanged,

	// Note events
	"note.added":   CompensateNoteAdded,
	"note.updated": CompensateNoteUpdated,

	// Channel events
	"channel.created":   CompensateChannelCreated,
	"channel.activated": CompensateChannelActivated,

	// Tracking events
	"tracking.created": CompensateTrackingCreated,

	// Automation events
	"automation_rule.executed": CompensateAutomationExecuted,

	// Agent events
	"agent.created":            CompensateAgentCreated,
	"agent.permission_granted": CompensateAgentPermissionGranted,

	// Billing events
	"billing.account_created":           CompensateBillingAccountCreated,
	"billing.payment_method_activated": CompensatePaymentMethodActivated,
}

// GetCompensationEvent retorna o evento de compensação para um evento de domínio.
// Retorna empty string se não há compensação necessária.
func GetCompensationEvent(domainEventType string) CompensationEventType {
	if compensation, exists := CompensationMapping[domainEventType]; exists {
		return compensation
	}
	return ""
}

// NeedsCompensation verifica se um evento de domínio necessita compensação.
func NeedsCompensation(domainEventType string) bool {
	_, exists := CompensationMapping[domainEventType]
	return exists
}

// CompensationAction define uma ação de compensação a ser executada.
type CompensationAction struct {
	EventType     CompensationEventType  `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	Reason        string                 `json:"reason"`
	OriginalEvent string                 `json:"original_event"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CompensationStrategy define como compensar uma Saga.
type CompensationStrategy string

const (
	// ReverseOrder compensa na ordem reversa (LIFO - Last In First Out)
	// Exemplo: Se Saga executou [A, B, C], compensa [C, B, A]
	ReverseOrder CompensationStrategy = "reverse_order"

	// AllAtOnce dispara todas as compensações em paralelo
	// Usado quando as compensações são independentes
	AllAtOnce CompensationStrategy = "all_at_once"

	// Selective compensa apenas eventos específicos (definidos por selector)
	// Útil quando nem todos os steps precisam ser compensados
	Selective CompensationStrategy = "selective"
)

// DefaultCompensationStrategy é a estratégia padrão (ordem reversa - LIFO).
const DefaultCompensationStrategy = ReverseOrder

// CompensationConfig configura como uma Saga deve ser compensada.
type CompensationConfig struct {
	Strategy      CompensationStrategy   `json:"strategy"`
	MaxRetries    int                    `json:"max_retries"`
	EventSelector func(eventType string) bool `json:"-"` // Usado apenas com Selective strategy
}

// DefaultCompensationConfig retorna configuração padrão de compensação.
func DefaultCompensationConfig() CompensationConfig {
	return CompensationConfig{
		Strategy:   DefaultCompensationStrategy,
		MaxRetries: 3,
	}
}
