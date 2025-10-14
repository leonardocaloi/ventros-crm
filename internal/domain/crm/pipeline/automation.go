package pipeline

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type Automation struct {
	id             uuid.UUID
	automationType AutomationType
	pipelineID     *uuid.UUID
	tenantID       string
	name           string
	description    string
	trigger        AutomationTrigger
	conditions     []RuleCondition
	actions        []RuleAction
	priority       int
	enabled        bool
	createdAt      time.Time
	updatedAt      time.Time

	events []shared.DomainEvent
}

type AutomationType string

const (
	// Pipeline-based automations (require pipelineID)
	AutomationTypePipelineBased AutomationType = "pipeline_based"
	AutomationTypeFollowUp      AutomationType = "follow_up"
	AutomationTypeReengagement  AutomationType = "reengagement"
	AutomationTypeOnboarding    AutomationType = "onboarding"

	// Event-driven automations
	AutomationTypeEvent AutomationType = "event"

	// Scheduled automations
	AutomationTypeScheduled        AutomationType = "scheduled"
	AutomationTypeScheduledReport  AutomationType = "scheduled_report"
	AutomationTypeTimeNotification AutomationType = "time_notification"

	// Integration automations
	AutomationTypeWebhook AutomationType = "webhook"

	// Custom automations
	AutomationTypeCustom AutomationType = "custom"
)

type AutomationTrigger string

const (
	TriggerSessionEnded     AutomationTrigger = "session.ended"
	TriggerSessionTimeout   AutomationTrigger = "session.timeout"
	TriggerSessionResolved  AutomationTrigger = "session.resolved"
	TriggerSessionEscalated AutomationTrigger = "session.escalated"

	TriggerNoResponse      AutomationTrigger = "no_response.timeout"
	TriggerMessageReceived AutomationTrigger = "message.received"

	TriggerStatusChanged  AutomationTrigger = "status.changed"
	TriggerStageCompleted AutomationTrigger = "stage.completed"

	TriggerAfterDelay AutomationTrigger = "after.delay"
	TriggerScheduled  AutomationTrigger = "scheduled"

	TriggerPurchaseCompleted AutomationTrigger = "purchase.completed"
	TriggerPaymentReceived   AutomationTrigger = "payment.received"
	TriggerRefundIssued      AutomationTrigger = "refund.issued"
	TriggerCartAbandoned     AutomationTrigger = "cart.abandoned"
	TriggerOrderShipped      AutomationTrigger = "order.shipped"

	TriggerPageVisited    AutomationTrigger = "page.visited"
	TriggerFormSubmitted  AutomationTrigger = "form.submitted"
	TriggerFileDownloaded AutomationTrigger = "file.downloaded"
)

type LogicOperator string

const (
	LogicAND LogicOperator = "AND"
	LogicOR  LogicOperator = "OR"
)

type RuleCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type ConditionGroup struct {
	Logic      LogicOperator    `json:"logic"`
	Conditions []RuleCondition  `json:"conditions"`
	Groups     []ConditionGroup `json:"groups"`
}

type ConditionOperator struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

func GetAvailableOperators() []ConditionOperator {
	return []ConditionOperator{
		{Code: "eq", Name: "Igual a", Description: "Valor igual ao especificado", Example: "status == 'Lead'"},
		{Code: "ne", Name: "Diferente de", Description: "Valor diferente do especificado", Example: "status != 'Cliente'"},
		{Code: "gt", Name: "Maior que", Description: "Valor maior que o especificado", Example: "message_count > 5"},
		{Code: "gte", Name: "Maior ou igual", Description: "Valor maior ou igual ao especificado", Example: "hours >= 24"},
		{Code: "lt", Name: "Menor que", Description: "Valor menor que o especificado", Example: "days < 7"},
		{Code: "lte", Name: "Menor ou igual", Description: "Valor menor ou igual ao especificado", Example: "amount <= 100"},
		{Code: "contains", Name: "Contém", Description: "String contém o valor especificado", Example: "message contains 'urgente'"},
		{Code: "in", Name: "Está em", Description: "Valor está na lista especificada", Example: "status in ['Lead', 'Qualificado']"},
	}
}

type AutomationAction string

const (
	ActionSendMessage  AutomationAction = "send_message"
	ActionSendTemplate AutomationAction = "send_template"

	ActionChangeStatus  AutomationAction = "change_pipeline_status"
	ActionAssignAgent   AutomationAction = "assign_agent"
	ActionAssignToQueue AutomationAction = "assign_to_queue"

	ActionCreateTask        AutomationAction = "create_task"
	ActionAddTag            AutomationAction = "add_tag"
	ActionRemoveTag         AutomationAction = "remove_tag"
	ActionUpdateCustomField AutomationAction = "update_custom_field"

	ActionCreateNote        AutomationAction = "create_note"
	ActionCreateAgentReport AutomationAction = "create_agent_report"

	ActionSendWebhook     AutomationAction = "send_webhook"
	ActionTriggerWorkflow AutomationAction = "trigger_workflow"

	ActionNotifyAgent       AutomationAction = "notify_agent"
	ActionNotifyCoordinator AutomationAction = "notify_coordinator"
	ActionSendEmail         AutomationAction = "send_email"
)

type RuleAction struct {
	Type   AutomationAction       `json:"type"`
	Params map[string]interface{} `json:"params"`
	Delay  int                    `json:"delay_minutes,omitempty"`
}

type ActionMetadata struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Parameters  []ActionParameter      `json:"parameters"`
	Example     map[string]interface{} `json:"example"`
}

type ActionParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

func GetAvailableActions() []ActionMetadata {
	return []ActionMetadata{
		{
			Code:        string(ActionSendMessage),
			Name:        "Enviar Mensagem",
			Description: "Envia a text message para o contato",
			Category:    "messaging",
			Parameters: []ActionParameter{
				{Name: "content", Type: "string", Required: true, Description: "Conteúdo da mensagem"},
			},
			Example: map[string]interface{}{"content": "Olá! Como posso ajudar?"},
		},
		{
			Code:        string(ActionSendTemplate),
			Name:        "Enviar Template",
			Description: "Envia mensagem usando template pré-definido",
			Category:    "messaging",
			Parameters: []ActionParameter{
				{Name: "template_name", Type: "string", Required: true, Description: "Nome do template"},
				{Name: "params", Type: "object", Required: false, Description: "Parâmetros do template"},
			},
			Example: map[string]interface{}{"template_name": "welcome", "params": map[string]string{"name": "João"}},
		},
		{
			Code:        string(ActionChangeStatus),
			Name:        "Mudar Status",
			Description: "Altera o status do contato no pipeline",
			Category:    "pipeline",
			Parameters: []ActionParameter{
				{Name: "status_id", Type: "string", Required: true, Description: "UUID do novo status"},
			},
			Example: map[string]interface{}{"status_id": "uuid-here"},
		},
		{
			Code:        string(ActionAssignAgent),
			Name:        "Atribuir Agente",
			Description: "Atribui contato a um agente específico",
			Category:    "assignment",
			Parameters: []ActionParameter{
				{Name: "agent_id", Type: "string", Required: true, Description: "UUID do agente"},
			},
			Example: map[string]interface{}{"agent_id": "uuid-here"},
		},
		{
			Code:        string(ActionAssignToQueue),
			Name:        "Atribuir à Fila",
			Description: "Atribui contato a uma fila de atendimento",
			Category:    "assignment",
			Parameters: []ActionParameter{
				{Name: "queue_id", Type: "string", Required: true, Description: "UUID da fila"},
			},
			Example: map[string]interface{}{"queue_id": "uuid-here"},
		},
		{
			Code:        string(ActionCreateTask),
			Name:        "Criar Tarefa",
			Description: "Cria tarefa relacionada ao contato",
			Category:    "tasks",
			Parameters: []ActionParameter{
				{Name: "title", Type: "string", Required: true, Description: "Título da tarefa"},
				{Name: "description", Type: "string", Required: false, Description: "Descrição detalhada"},
				{Name: "due_date", Type: "string", Required: false, Description: "Data de vencimento (ISO 8601)"},
			},
			Example: map[string]interface{}{"title": "Ligar para cliente", "description": "Follow-up da proposta"},
		},
		{
			Code:        string(ActionSendWebhook),
			Name:        "Enviar Webhook",
			Description: "Dispara webhook para URL externa",
			Category:    "integration",
			Parameters: []ActionParameter{
				{Name: "url", Type: "string", Required: true, Description: "URL do webhook"},
				{Name: "payload", Type: "object", Required: false, Description: "Dados a enviar"},
				{Name: "headers", Type: "object", Required: false, Description: "Headers HTTP customizados"},
			},
			Example: map[string]interface{}{"url": "https://api.exemplo.com/webhook", "payload": map[string]string{"event": "status_changed"}},
		},
		{
			Code:        string(ActionAddTag),
			Name:        "Adicionar Tag",
			Description: "Adiciona tag ao contato",
			Category:    "organization",
			Parameters: []ActionParameter{
				{Name: "tag", Type: "string", Required: true, Description: "Nome da tag"},
			},
			Example: map[string]interface{}{"tag": "vip"},
		},
		{
			Code:        string(ActionRemoveTag),
			Name:        "Remover Tag",
			Description: "Remove tag do contato",
			Category:    "organization",
			Parameters: []ActionParameter{
				{Name: "tag", Type: "string", Required: true, Description: "Nome da tag"},
			},
			Example: map[string]interface{}{"tag": "lead_frio"},
		},
		{
			Code:        string(ActionUpdateCustomField),
			Name:        "Atualizar Campo Customizado",
			Description: "Atualiza valor de campo customizado",
			Category:    "data",
			Parameters: []ActionParameter{
				{Name: "field_name", Type: "string", Required: true, Description: "Nome do campo"},
				{Name: "value", Type: "string", Required: true, Description: "Novo valor"},
			},
			Example: map[string]interface{}{"field_name": "last_contact", "value": "2025-01-15"},
		},
		{
			Code:        string(ActionTriggerWorkflow),
			Name:        "Disparar Workflow",
			Description: "Inicia workflow Temporal",
			Category:    "workflow",
			Parameters: []ActionParameter{
				{Name: "workflow_name", Type: "string", Required: true, Description: "Nome do workflow"},
				{Name: "params", Type: "object", Required: false, Description: "Parâmetros do workflow"},
			},
			Example: map[string]interface{}{"workflow_name": "email_campaign", "params": map[string]string{"campaign_id": "123"}},
		},
		{
			Code:        string(ActionCreateNote),
			Name:        "Criar Nota",
			Description: "Cria nota vinculada a agent/contact/session",
			Category:    "notes",
			Parameters: []ActionParameter{
				{Name: "entity_type", Type: "string", Required: true, Description: "Tipo de entidade (agent, contact, session)"},
				{Name: "entity_id", Type: "string", Required: true, Description: "UUID da entidade"},
				{Name: "content", Type: "string", Required: true, Description: "Conteúdo da nota"},
				{Name: "title", Type: "string", Required: false, Description: "Título da nota"},
			},
			Example: map[string]interface{}{"entity_type": "agent", "entity_id": "{{agent_id}}", "content": "Performance report", "title": "Daily Report"},
		},
		{
			Code:        string(ActionCreateAgentReport),
			Name:        "Gerar Relatório de Agente",
			Description: "Gera relatório de performance com IA",
			Category:    "reports",
			Parameters: []ActionParameter{
				{Name: "agent_id", Type: "string", Required: true, Description: "UUID do agente"},
				{Name: "period", Type: "string", Required: false, Description: "Período (daily, weekly, monthly)", Default: "daily"},
				{Name: "include_comparisons", Type: "boolean", Required: false, Description: "Incluir comparações entre agentes", Default: true},
				{Name: "notify_coordinator", Type: "boolean", Required: false, Description: "Notificar coordenador", Default: false},
			},
			Example: map[string]interface{}{"agent_id": "{{agent_id}}", "period": "daily", "include_comparisons": true, "notify_coordinator": true},
		},
		{
			Code:        string(ActionNotifyAgent),
			Name:        "Notificar Agente",
			Description: "Envia notificação para agente específico",
			Category:    "notifications",
			Parameters: []ActionParameter{
				{Name: "agent_id", Type: "string", Required: true, Description: "UUID do agente"},
				{Name: "message", Type: "string", Required: true, Description: "Mensagem da notificação"},
				{Name: "channel", Type: "string", Required: false, Description: "Canal de notificação (whatsapp, email, in_app)", Default: "in_app"},
			},
			Example: map[string]interface{}{"agent_id": "{{agent_id}}", "message": "Novo lead atribuído", "channel": "whatsapp"},
		},
		{
			Code:        string(ActionNotifyCoordinator),
			Name:        "Notificar Coordenador",
			Description: "Envia notificação para coordenador de vendas",
			Category:    "notifications",
			Parameters: []ActionParameter{
				{Name: "coordinator_id", Type: "string", Required: false, Description: "UUID do coordenador (opcional, usa padrão do projeto)"},
				{Name: "message", Type: "string", Required: true, Description: "Mensagem da notificação"},
				{Name: "channel", Type: "string", Required: false, Description: "Canal de notificação (whatsapp, email, in_app)", Default: "in_app"},
				{Name: "priority", Type: "string", Required: false, Description: "Prioridade (low, medium, high)", Default: "medium"},
			},
			Example: map[string]interface{}{"message": "Relatório diário disponível", "channel": "whatsapp", "priority": "high"},
		},
		{
			Code:        string(ActionSendEmail),
			Name:        "Enviar Email",
			Description: "Envia email para destinatário específico",
			Category:    "messaging",
			Parameters: []ActionParameter{
				{Name: "to", Type: "string", Required: true, Description: "Email do destinatário"},
				{Name: "subject", Type: "string", Required: true, Description: "Assunto do email"},
				{Name: "body", Type: "string", Required: true, Description: "Corpo do email"},
				{Name: "cc", Type: "array", Required: false, Description: "Lista de emails em cópia"},
			},
			Example: map[string]interface{}{"to": "coordenador@empresa.com", "subject": "Relatório Diário", "body": "{{report_content}}"},
		},
	}
}

func NewAutomation(
	automationType AutomationType,
	tenantID string,
	name string,
	trigger AutomationTrigger,
	pipelineID *uuid.UUID,
) (*Automation, error) {
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if trigger == "" {
		return nil, errors.New("trigger cannot be empty")
	}
	if automationType == "" {
		return nil, errors.New("automationType cannot be empty")
	}

	// Pipeline-based automations require a valid pipelineID
	if automationType == AutomationTypePipelineBased || automationType == AutomationTypeFollowUp ||
		automationType == AutomationTypeReengagement || automationType == AutomationTypeOnboarding {
		if pipelineID == nil || *pipelineID == uuid.Nil {
			return nil, errors.New("pipeline-based automations require a valid pipelineID")
		}
	}

	now := time.Now()
	rule := &Automation{
		id:             uuid.New(),
		automationType: automationType,
		pipelineID:     pipelineID,
		tenantID:       tenantID,
		name:           name,
		trigger:        trigger,
		conditions:     []RuleCondition{},
		actions:        []RuleAction{},
		priority:       0,
		enabled:        true,
		createdAt:      now,
		updatedAt:      now,
		events:         []shared.DomainEvent{},
	}

	var eventPipelineID uuid.UUID
	if pipelineID != nil {
		eventPipelineID = *pipelineID
	}

	rule.addEvent(NewAutomationCreatedEvent(rule.id, eventPipelineID, tenantID, name, trigger))

	return rule, nil
}

func ReconstructAutomation(
	id uuid.UUID,
	automationType AutomationType,
	pipelineID *uuid.UUID,
	tenantID, name, description string,
	trigger AutomationTrigger,
	conditions []RuleCondition,
	actions []RuleAction,
	priority int,
	enabled bool,
	createdAt, updatedAt time.Time,
) *Automation {
	return &Automation{
		id:             id,
		automationType: automationType,
		pipelineID:     pipelineID,
		tenantID:       tenantID,
		name:           name,
		description:    description,
		trigger:        trigger,
		conditions:     conditions,
		actions:        actions,
		priority:       priority,
		enabled:        enabled,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		events:         []shared.DomainEvent{},
	}
}

func (r *Automation) AddCondition(field, operator string, value interface{}) error {
	if field == "" {
		return errors.New("field cannot be empty")
	}
	if operator == "" {
		return errors.New("operator cannot be empty")
	}

	condition := RuleCondition{
		Field:    field,
		Operator: operator,
		Value:    value,
	}

	r.conditions = append(r.conditions, condition)
	r.updatedAt = time.Now()

	return nil
}

func (r *Automation) AddAction(actionType AutomationAction, params map[string]interface{}, delayMinutes int) error {
	if actionType == "" {
		return errors.New("action type cannot be empty")
	}

	action := RuleAction{
		Type:   actionType,
		Params: params,
		Delay:  delayMinutes,
	}

	r.actions = append(r.actions, action)
	r.updatedAt = time.Now()

	return nil
}

func (r *Automation) SetConditions(conditions []RuleCondition) {
	r.conditions = conditions
	r.updatedAt = time.Now()
}

func (r *Automation) SetActions(actions []RuleAction) {
	r.actions = actions
	r.updatedAt = time.Now()
}

func (r *Automation) UpdateDescription(description string) {
	r.description = description
	r.updatedAt = time.Now()
}

func (r *Automation) SetPriority(priority int) error {
	if priority < 0 {
		return errors.New("priority cannot be negative")
	}

	r.priority = priority
	r.updatedAt = time.Now()

	return nil
}

func (r *Automation) Enable() {
	if !r.enabled {
		r.enabled = true
		r.updatedAt = time.Now()

		r.addEvent(NewAutomationEnabledEvent(r.id))
	}
}

func (r *Automation) Disable() {
	if r.enabled {
		r.enabled = false
		r.updatedAt = time.Now()

		r.addEvent(NewAutomationDisabledEvent(r.id))
	}
}

func (r *Automation) EvaluateConditions(context map[string]interface{}) bool {

	if len(r.conditions) == 0 {
		return true
	}

	for _, condition := range r.conditions {
		if !evaluateCondition(condition, context) {
			return false
		}
	}

	return true
}

func EvaluateConditionGroup(group ConditionGroup, context map[string]interface{}) bool {

	if len(group.Conditions) == 0 && len(group.Groups) == 0 {
		return true
	}

	results := make([]bool, 0)

	for _, condition := range group.Conditions {
		results = append(results, evaluateCondition(condition, context))
	}

	for _, subGroup := range group.Groups {
		results = append(results, EvaluateConditionGroup(subGroup, context))
	}

	if group.Logic == LogicOR {

		for _, result := range results {
			if result {
				return true
			}
		}
		return false
	}

	for _, result := range results {
		if !result {
			return false
		}
	}
	return true
}

func evaluateCondition(condition RuleCondition, context map[string]interface{}) bool {
	fieldValue, exists := context[condition.Field]
	if !exists {
		return false
	}

	switch condition.Operator {
	case "eq", "equals":
		return fieldValue == condition.Value

	case "ne", "not_equals":
		return fieldValue != condition.Value

	case "gt", "greater_than":
		return compareNumeric(fieldValue, condition.Value, ">")

	case "gte", "greater_than_or_equal":
		return compareNumeric(fieldValue, condition.Value, ">=")

	case "lt", "less_than":
		return compareNumeric(fieldValue, condition.Value, "<")

	case "lte", "less_than_or_equal":
		return compareNumeric(fieldValue, condition.Value, "<=")

	case "contains":
		return containsString(fieldValue, condition.Value)

	case "in":
		return inSlice(fieldValue, condition.Value)

	default:
		return false
	}
}

func compareNumeric(a, b interface{}, op string) bool {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch op {
	case ">":
		return aFloat > bFloat
	case ">=":
		return aFloat >= bFloat
	case "<":
		return aFloat < bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

func containsString(haystack, needle interface{}) bool {
	haystackStr, ok1 := haystack.(string)
	needleStr, ok2 := needle.(string)

	if !ok1 || !ok2 {
		return false
	}

	return len(haystackStr) > 0 && len(needleStr) > 0 &&
		(haystackStr == needleStr || len(haystackStr) >= len(needleStr) &&
			haystackStr[:len(needleStr)] == needleStr)
}

func inSlice(value interface{}, slice interface{}) bool {
	sliceVal, ok := slice.([]interface{})
	if !ok {
		return false
	}

	for _, item := range sliceVal {
		if item == value {
			return true
		}
	}

	return false
}

func (r *Automation) ID() uuid.UUID               { return r.id }
func (r *Automation) Type() AutomationType        { return r.automationType }
func (r *Automation) PipelineID() *uuid.UUID      { return r.pipelineID }
func (r *Automation) TenantID() string            { return r.tenantID }
func (r *Automation) Name() string                { return r.name }
func (r *Automation) Description() string         { return r.description }
func (r *Automation) Trigger() AutomationTrigger  { return r.trigger }
func (r *Automation) Conditions() []RuleCondition { return append([]RuleCondition{}, r.conditions...) }
func (r *Automation) Actions() []RuleAction       { return append([]RuleAction{}, r.actions...) }
func (r *Automation) Priority() int               { return r.priority }
func (r *Automation) IsEnabled() bool             { return r.enabled }
func (r *Automation) CreatedAt() time.Time        { return r.createdAt }
func (r *Automation) UpdatedAt() time.Time        { return r.updatedAt }

func (r *Automation) DomainEvents() []shared.DomainEvent {
	return append([]shared.DomainEvent{}, r.events...)
}

func (r *Automation) ClearEvents() {
	r.events = []shared.DomainEvent{}
}

func (r *Automation) addEvent(event shared.DomainEvent) {
	r.events = append(r.events, event)
}

type AutomationRepository interface {
	Save(rule *Automation) error
	FindByID(id uuid.UUID) (*Automation, error)
	FindByPipeline(pipelineID uuid.UUID) ([]*Automation, error)
	FindByPipelineAndTrigger(pipelineID uuid.UUID, trigger AutomationTrigger) ([]*Automation, error)
	FindEnabledByPipeline(pipelineID uuid.UUID) ([]*Automation, error)
	Delete(id uuid.UUID) error
}
