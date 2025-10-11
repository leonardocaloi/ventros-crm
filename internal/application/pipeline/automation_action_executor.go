package pipeline

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
)

// DefaultActionExecutor implementação padrão do ActionExecutor
type DefaultActionExecutor struct {
	messageSender         MessageSender
	pipelineStatusChanger PipelineStatusChanger
	agentAssigner         AgentAssigner
	queueAssigner         QueueAssigner
	webhookSender         WebhookSender
	tagManager            TagManager
	customFieldUpdater    CustomFieldUpdater
	workflowTrigger       WorkflowTrigger
	logger                Logger
}

// Interfaces para serviços externos

// MessageSender envia mensagens para contatos
type MessageSender interface {
	SendMessage(ctx context.Context, contactID uuid.UUID, channelID uuid.UUID, content string) error
	SendTemplate(ctx context.Context, contactID uuid.UUID, channelID uuid.UUID, templateName string, params map[string]interface{}) error
}

// PipelineStatusChanger altera status do contato no pipeline
type PipelineStatusChanger interface {
	ChangeStatus(ctx context.Context, contactID uuid.UUID, pipelineID uuid.UUID, newStatusID uuid.UUID) error
}

// AgentAssigner atribui agente a uma sessão/contato
type AgentAssigner interface {
	AssignAgent(ctx context.Context, sessionID uuid.UUID, agentID uuid.UUID) error
}

// QueueAssigner atribui sessão a uma fila
type QueueAssigner interface {
	AssignToQueue(ctx context.Context, sessionID uuid.UUID, queueID uuid.UUID) error
}

// WebhookSender envia webhooks
type WebhookSender interface {
	SendWebhook(ctx context.Context, url string, payload map[string]interface{}) error
}

// TagManager gerencia tags de contatos
type TagManager interface {
	AddTag(ctx context.Context, contactID uuid.UUID, tag string) error
	RemoveTag(ctx context.Context, contactID uuid.UUID, tag string) error
}

// CustomFieldUpdater atualiza campos customizados
type CustomFieldUpdater interface {
	UpdateCustomField(ctx context.Context, contactID uuid.UUID, fieldName string, value interface{}) error
}

// WorkflowTrigger dispara workflows
type WorkflowTrigger interface {
	TriggerWorkflow(ctx context.Context, workflowName string, input map[string]interface{}) error
}

// NewDefaultActionExecutor cria novo executor
func NewDefaultActionExecutor(
	messageSender MessageSender,
	pipelineStatusChanger PipelineStatusChanger,
	agentAssigner AgentAssigner,
	queueAssigner QueueAssigner,
	webhookSender WebhookSender,
	tagManager TagManager,
	customFieldUpdater CustomFieldUpdater,
	workflowTrigger WorkflowTrigger,
	logger Logger,
) *DefaultActionExecutor {
	return &DefaultActionExecutor{
		messageSender:         messageSender,
		pipelineStatusChanger: pipelineStatusChanger,
		agentAssigner:         agentAssigner,
		queueAssigner:         queueAssigner,
		webhookSender:         webhookSender,
		tagManager:            tagManager,
		customFieldUpdater:    customFieldUpdater,
		workflowTrigger:       workflowTrigger,
		logger:                logger,
	}
}

// Execute executa uma ação específica
func (e *DefaultActionExecutor) Execute(
	ctx context.Context,
	action pipeline.RuleAction,
	actionCtx ActionContext,
) error {
	e.logger.Debug("executing action", "type", action.Type, "ruleID", actionCtx.RuleID)

	switch action.Type {
	case pipeline.ActionSendMessage:
		return e.executeSendMessage(ctx, action, actionCtx)
	case pipeline.ActionSendTemplate:
		return e.executeSendTemplate(ctx, action, actionCtx)
	case pipeline.ActionChangeStatus:
		return e.executeChangeStatus(ctx, action, actionCtx)
	case pipeline.ActionAssignAgent:
		return e.executeAssignAgent(ctx, action, actionCtx)
	case pipeline.ActionAssignToQueue:
		return e.executeAssignToQueue(ctx, action, actionCtx)
	case pipeline.ActionSendWebhook:
		return e.executeSendWebhook(ctx, action, actionCtx)
	case pipeline.ActionAddTag:
		return e.executeAddTag(ctx, action, actionCtx)
	case pipeline.ActionRemoveTag:
		return e.executeRemoveTag(ctx, action, actionCtx)
	case pipeline.ActionUpdateCustomField:
		return e.executeUpdateCustomField(ctx, action, actionCtx)
	case pipeline.ActionTriggerWorkflow:
		return e.executeTriggerWorkflow(ctx, action, actionCtx)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (e *DefaultActionExecutor) executeSendMessage(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.ContactID == nil {
		return fmt.Errorf("contactID is required for send_message action")
	}
	if actionCtx.ChannelID == nil {
		return fmt.Errorf("channelID is required for send_message action")
	}

	content, ok := action.Params["content"].(string)
	if !ok {
		return fmt.Errorf("content parameter is required for send_message action")
	}

	if e.messageSender == nil {
		return fmt.Errorf("messageSender not configured")
	}

	return e.messageSender.SendMessage(ctx, *actionCtx.ContactID, *actionCtx.ChannelID, content)
}

func (e *DefaultActionExecutor) executeSendTemplate(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.ContactID == nil {
		return fmt.Errorf("contactID is required for send_template action")
	}
	if actionCtx.ChannelID == nil {
		return fmt.Errorf("channelID is required for send_template action")
	}

	templateName, ok := action.Params["template_name"].(string)
	if !ok {
		return fmt.Errorf("template_name parameter is required for send_template action")
	}

	templateParams, _ := action.Params["params"].(map[string]interface{})

	if e.messageSender == nil {
		return fmt.Errorf("messageSender not configured")
	}

	return e.messageSender.SendTemplate(ctx, *actionCtx.ContactID, *actionCtx.ChannelID, templateName, templateParams)
}

func (e *DefaultActionExecutor) executeChangeStatus(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.ContactID == nil {
		return fmt.Errorf("contactID is required for change_status action")
	}

	statusIDStr, ok := action.Params["status_id"].(string)
	if !ok {
		return fmt.Errorf("status_id parameter is required for change_status action")
	}

	statusID, err := uuid.Parse(statusIDStr)
	if err != nil {
		return fmt.Errorf("invalid status_id: %w", err)
	}

	if e.pipelineStatusChanger == nil {
		return fmt.Errorf("pipelineStatusChanger not configured")
	}

	if actionCtx.PipelineID == nil {
		return fmt.Errorf("pipelineID is required for change_status action")
	}

	return e.pipelineStatusChanger.ChangeStatus(ctx, *actionCtx.ContactID, *actionCtx.PipelineID, statusID)
}

func (e *DefaultActionExecutor) executeAssignAgent(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.SessionID == nil {
		return fmt.Errorf("sessionID is required for assign_agent action")
	}

	agentIDStr, ok := action.Params["agent_id"].(string)
	if !ok {
		return fmt.Errorf("agent_id parameter is required for assign_agent action")
	}

	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		return fmt.Errorf("invalid agent_id: %w", err)
	}

	if e.agentAssigner == nil {
		return fmt.Errorf("agentAssigner not configured")
	}

	return e.agentAssigner.AssignAgent(ctx, *actionCtx.SessionID, agentID)
}

func (e *DefaultActionExecutor) executeAssignToQueue(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.SessionID == nil {
		return fmt.Errorf("sessionID is required for assign_to_queue action")
	}

	queueIDStr, ok := action.Params["queue_id"].(string)
	if !ok {
		return fmt.Errorf("queue_id parameter is required for assign_to_queue action")
	}

	queueID, err := uuid.Parse(queueIDStr)
	if err != nil {
		return fmt.Errorf("invalid queue_id: %w", err)
	}

	if e.queueAssigner == nil {
		return fmt.Errorf("queueAssigner not configured")
	}

	return e.queueAssigner.AssignToQueue(ctx, *actionCtx.SessionID, queueID)
}

func (e *DefaultActionExecutor) executeSendWebhook(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	url, ok := action.Params["url"].(string)
	if !ok {
		return fmt.Errorf("url parameter is required for send_webhook action")
	}

	payload, _ := action.Params["payload"].(map[string]interface{})
	if payload == nil {
		// Se não forneceu payload customizado, usa contexto da ação
		payload = map[string]interface{}{
			"rule_id":   actionCtx.RuleID.String(),
			"tenant_id": actionCtx.TenantID,
			"trigger":   string(actionCtx.Trigger),
		}
		if actionCtx.PipelineID != nil {
			payload["pipeline_id"] = actionCtx.PipelineID.String()
		}
		if actionCtx.SessionID != nil {
			payload["session_id"] = actionCtx.SessionID.String()
		}
		if actionCtx.ContactID != nil {
			payload["contact_id"] = actionCtx.ContactID.String()
		}
	}

	if e.webhookSender == nil {
		return fmt.Errorf("webhookSender not configured")
	}

	return e.webhookSender.SendWebhook(ctx, url, payload)
}

func (e *DefaultActionExecutor) executeAddTag(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.ContactID == nil {
		return fmt.Errorf("contactID is required for add_tag action")
	}

	tag, ok := action.Params["tag"].(string)
	if !ok {
		return fmt.Errorf("tag parameter is required for add_tag action")
	}

	if e.tagManager == nil {
		return fmt.Errorf("tagManager not configured")
	}

	return e.tagManager.AddTag(ctx, *actionCtx.ContactID, tag)
}

func (e *DefaultActionExecutor) executeRemoveTag(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.ContactID == nil {
		return fmt.Errorf("contactID is required for remove_tag action")
	}

	tag, ok := action.Params["tag"].(string)
	if !ok {
		return fmt.Errorf("tag parameter is required for remove_tag action")
	}

	if e.tagManager == nil {
		return fmt.Errorf("tagManager not configured")
	}

	return e.tagManager.RemoveTag(ctx, *actionCtx.ContactID, tag)
}

func (e *DefaultActionExecutor) executeUpdateCustomField(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	if actionCtx.ContactID == nil {
		return fmt.Errorf("contactID is required for update_custom_field action")
	}

	fieldName, ok := action.Params["field_name"].(string)
	if !ok {
		return fmt.Errorf("field_name parameter is required for update_custom_field action")
	}

	value, ok := action.Params["value"]
	if !ok {
		return fmt.Errorf("value parameter is required for update_custom_field action")
	}

	if e.customFieldUpdater == nil {
		return fmt.Errorf("customFieldUpdater not configured")
	}

	return e.customFieldUpdater.UpdateCustomField(ctx, *actionCtx.ContactID, fieldName, value)
}

func (e *DefaultActionExecutor) executeTriggerWorkflow(ctx context.Context, action pipeline.RuleAction, actionCtx ActionContext) error {
	workflowName, ok := action.Params["workflow_name"].(string)
	if !ok {
		return fmt.Errorf("workflow_name parameter is required for trigger_workflow action")
	}

	input, _ := action.Params["input"].(map[string]interface{})
	if input == nil {
		// Usa contexto como input padrão
		input = map[string]interface{}{
			"rule_id":   actionCtx.RuleID.String(),
			"tenant_id": actionCtx.TenantID,
		}
		if actionCtx.PipelineID != nil {
			input["pipeline_id"] = actionCtx.PipelineID.String()
		}
		if actionCtx.SessionID != nil {
			input["session_id"] = actionCtx.SessionID.String()
		}
		if actionCtx.ContactID != nil {
			input["contact_id"] = actionCtx.ContactID.String()
		}
	}

	if e.workflowTrigger == nil {
		return fmt.Errorf("workflowTrigger not configured")
	}

	return e.workflowTrigger.TriggerWorkflow(ctx, workflowName, input)
}
