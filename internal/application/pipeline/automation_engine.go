package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/google/uuid"
)

// AutomationEngine é responsável por avaliar e executar regras de follow-up
type AutomationEngine struct {
	ruleRepo       pipeline.AutomationRepository
	actionExecutor ActionExecutor
	logger         Logger
}

// Logger interface para logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// ActionExecutor executa as ações das regras
type ActionExecutor interface {
	Execute(ctx context.Context, action pipeline.RuleAction, context ActionContext) error
}

// ActionContext contém o contexto para execução de uma ação
type ActionContext struct {
	SessionID  *uuid.UUID
	ContactID  *uuid.UUID
	ChannelID  *uuid.UUID
	PipelineID *uuid.UUID // ponteiro pois pode ser opcional
	TenantID   string
	RuleID     uuid.UUID
	Trigger    pipeline.AutomationTrigger
	Metadata   map[string]interface{}
}

// NewAutomationEngine cria uma nova instância do engine
func NewAutomationEngine(
	ruleRepo pipeline.AutomationRepository,
	actionExecutor ActionExecutor,
	logger Logger,
) *AutomationEngine {
	if logger == nil {
		logger = &defaultLogger{}
	}

	return &AutomationEngine{
		ruleRepo:       ruleRepo,
		actionExecutor: actionExecutor,
		logger:         logger,
	}
}

// EvaluateAndExecute avalia e executa regras para um evento específico
func (e *AutomationEngine) EvaluateAndExecute(
	ctx context.Context,
	pipelineID uuid.UUID,
	trigger pipeline.AutomationTrigger,
	evalContext map[string]interface{},
	actionCtx ActionContext,
) error {
	// Busca regras ativas para este pipeline e trigger
	rules, err := e.ruleRepo.FindByPipelineAndTrigger(pipelineID, trigger)
	if err != nil {
		e.logger.Error("failed to fetch rules", "pipelineID", pipelineID, "trigger", trigger, "error", err)
		return fmt.Errorf("failed to fetch rules: %w", err)
	}

	// Filtra apenas regras ativas
	activeRules := make([]*pipeline.Automation, 0, len(rules))
	for _, rule := range rules {
		if rule.IsEnabled() {
			activeRules = append(activeRules, rule)
		}
	}

	if len(activeRules) == 0 {
		e.logger.Debug("no active rules found", "pipelineID", pipelineID, "trigger", trigger)
		return nil
	}

	e.logger.Info("evaluating follow-up rules", "pipelineID", pipelineID, "trigger", trigger, "count", len(activeRules))

	// Avalia e executa cada regra (ordenadas por priority)
	executedCount := 0
	for _, rule := range activeRules {
		executed, err := e.evaluateAndExecuteRule(ctx, rule, evalContext, actionCtx)
		if err != nil {
			e.logger.Error("failed to execute rule", "ruleID", rule.ID(), "error", err)
			// Continua para próximas regras mesmo com erro
			continue
		}

		if executed {
			executedCount++
		}
	}

	e.logger.Info("follow-up rules execution completed", "trigger", trigger, "executed", executedCount, "total", len(activeRules))

	return nil
}

// evaluateAndExecuteRule avalia e executa uma regra individual
func (e *AutomationEngine) evaluateAndExecuteRule(
	ctx context.Context,
	rule *pipeline.Automation,
	evalContext map[string]interface{},
	actionCtx ActionContext,
) (bool, error) {
	// 1. Avalia condições
	if !rule.EvaluateConditions(evalContext) {
		e.logger.Debug("rule conditions not met", "ruleID", rule.ID(), "name", rule.Name())
		return false, nil
	}

	e.logger.Info("rule conditions met, executing actions", "ruleID", rule.ID(), "name", rule.Name())

	// 2. Preenche contexto da ação
	actionCtx.RuleID = rule.ID()
	actionCtx.PipelineID = rule.PipelineID()
	actionCtx.TenantID = rule.TenantID()

	// 3. Executa todas as ações da regra
	actions := rule.Actions()
	for i, action := range actions {
		// Se ação tem delay, agenda para execução futura
		if action.Delay > 0 {
			e.logger.Info("scheduling delayed action", "ruleID", rule.ID(), "actionIndex", i, "delayMinutes", action.Delay)
			if err := e.scheduleDelayedAction(ctx, action, actionCtx); err != nil {
				e.logger.Error("failed to schedule delayed action", "error", err)
				return false, fmt.Errorf("failed to schedule action: %w", err)
			}
			continue
		}

		// Executa ação imediatamente
		if err := e.actionExecutor.Execute(ctx, action, actionCtx); err != nil {
			e.logger.Error("failed to execute action", "ruleID", rule.ID(), "actionType", action.Type, "error", err)
			return false, fmt.Errorf("failed to execute action: %w", err)
		}

		e.logger.Debug("action executed successfully", "ruleID", rule.ID(), "actionType", action.Type)
	}

	return true, nil
}

// scheduleDelayedAction agenda uma ação para execução futura
func (e *AutomationEngine) scheduleDelayedAction(
	ctx context.Context,
	action pipeline.RuleAction,
	actionCtx ActionContext,
) error {
	// TODO: Integrar com Temporal Workflow para agendamento
	// Por enquanto, apenas loga
	e.logger.Info("delayed action scheduled",
		"actionType", action.Type,
		"delayMinutes", action.Delay,
		"scheduledFor", time.Now().Add(time.Duration(action.Delay)*time.Minute),
	)

	// Placeholder para integração futura com Temporal
	// workflow.ExecuteChildWorkflow(ctx, DelayedActionWorkflow, ...)

	return nil
}

// ProcessSessionEvent processa eventos de sessão e dispara regras apropriadas
func (e *AutomationEngine) ProcessSessionEvent(
	ctx context.Context,
	pipelineID uuid.UUID,
	trigger pipeline.AutomationTrigger,
	sessionID uuid.UUID,
	contactID uuid.UUID,
	channelID uuid.UUID,
	tenantID string,
	metadata map[string]interface{},
) error {
	// Constrói contexto de avaliação
	evalContext := map[string]interface{}{
		"session_id":  sessionID,
		"contact_id":  contactID,
		"channel_id":  channelID,
		"tenant_id":   tenantID,
		"occurred_at": time.Now(),
	}

	// Mescla com metadata fornecido
	for k, v := range metadata {
		evalContext[k] = v
	}

	// Constrói contexto de ação
	actionCtx := ActionContext{
		SessionID:  &sessionID,
		ContactID:  &contactID,
		ChannelID:  &channelID,
		PipelineID: &pipelineID,
		TenantID:   tenantID,
		Trigger:    trigger,
		Metadata:   metadata,
	}

	return e.EvaluateAndExecute(ctx, pipelineID, trigger, evalContext, actionCtx)
}

// ProcessContactEvent processa eventos de contato (sem sessão ativa)
func (e *AutomationEngine) ProcessContactEvent(
	ctx context.Context,
	pipelineID uuid.UUID,
	trigger pipeline.AutomationTrigger,
	contactID uuid.UUID,
	tenantID string,
	metadata map[string]interface{},
) error {
	evalContext := map[string]interface{}{
		"contact_id":  contactID,
		"tenant_id":   tenantID,
		"occurred_at": time.Now(),
	}

	for k, v := range metadata {
		evalContext[k] = v
	}

	actionCtx := ActionContext{
		ContactID:  &contactID,
		PipelineID: &pipelineID,
		TenantID:   tenantID,
		Trigger:    trigger,
		Metadata:   metadata,
	}

	return e.EvaluateAndExecute(ctx, pipelineID, trigger, evalContext, actionCtx)
}

// ProcessScheduledTrigger processa triggers agendados
func (e *AutomationEngine) ProcessScheduledTrigger(
	ctx context.Context,
	pipelineID uuid.UUID,
	metadata map[string]interface{},
) error {
	evalContext := map[string]interface{}{
		"occurred_at": time.Now(),
	}

	for k, v := range metadata {
		evalContext[k] = v
	}

	// Busca tenant_id do metadata
	tenantID, ok := metadata["tenant_id"].(string)
	if !ok {
		return fmt.Errorf("tenant_id not found in metadata")
	}

	actionCtx := ActionContext{
		PipelineID: &pipelineID,
		TenantID:   tenantID,
		Trigger:    pipeline.TriggerScheduled,
		Metadata:   metadata,
	}

	return e.EvaluateAndExecute(ctx, pipelineID, pipeline.TriggerScheduled, evalContext, actionCtx)
}

// defaultLogger implementação simples de Logger usando log padrão
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *defaultLogger) Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] "+msg, args...)
}
