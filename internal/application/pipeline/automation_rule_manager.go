package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
)

// AutomationRuleManager gerencia CRUD e operações complexas de Follow-up Rules
type AutomationRuleManager struct {
	ruleRepo     pipeline.AutomationRepository
	pipelineRepo pipeline.Repository
	validator    RuleValidator
	logger       Logger
}

// RuleValidator valida regras antes de salvar
type RuleValidator interface {
	ValidateRule(rule *pipeline.Automation) error
	ValidateSchedule(schedule *pipeline.ScheduledRuleConfig) error
}

// NewAutomationRuleManager cria novo gerenciador
func NewAutomationRuleManager(
	ruleRepo pipeline.AutomationRepository,
	pipelineRepo pipeline.Repository,
	validator RuleValidator,
	logger Logger,
) *AutomationRuleManager {
	if validator == nil {
		validator = &DefaultRuleValidator{}
	}

	return &AutomationRuleManager{
		ruleRepo:     ruleRepo,
		pipelineRepo: pipelineRepo,
		validator:    validator,
		logger:       logger,
	}
}

// CreateRuleInput input para criar regra
type CreateRuleInput struct {
	PipelineID  uuid.UUID
	TenantID    string
	Name        string
	Description string
	Trigger     pipeline.AutomationTrigger
	Conditions  []pipeline.RuleCondition
	Actions     []pipeline.RuleAction
	Priority    int
	Enabled     bool
	Schedule    *pipeline.ScheduledRuleConfig // opcional, apenas se Trigger = scheduled
}

// CreateRule cria nova regra de follow-up
func (m *AutomationRuleManager) CreateRule(ctx context.Context, input CreateRuleInput) (*pipeline.Automation, error) {
	// Valida pipeline existe
	pipe, err := m.pipelineRepo.FindPipelineByID(ctx, input.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pipeline: %w", err)
	}
	if pipe == nil {
		return nil, errors.New("pipeline not found")
	}

	// Cria regra
	pipelineIDPtr := &input.PipelineID
	rule, err := pipeline.NewAutomation(
		pipeline.AutomationTypePipelineBased, // tipo de automação baseada em pipeline
		input.TenantID,
		input.Name,
		input.Trigger,
		pipelineIDPtr, // ponteiro para pipelineID
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	// Configura campos opcionais
	if input.Description != "" {
		rule.UpdateDescription(input.Description)
	}

	if len(input.Conditions) > 0 {
		rule.SetConditions(input.Conditions)
	}

	if len(input.Actions) > 0 {
		rule.SetActions(input.Actions)
	}

	if input.Priority > 0 {
		if err := rule.SetPriority(input.Priority); err != nil {
			return nil, err
		}
	}

	if !input.Enabled {
		rule.Disable()
	}

	// Valida regra
	if err := m.validator.ValidateRule(rule); err != nil {
		return nil, fmt.Errorf("rule validation failed: %w", err)
	}

	// Se tem schedule, valida
	if input.Schedule != nil && input.Trigger == pipeline.TriggerScheduled {
		if err := m.validator.ValidateSchedule(input.Schedule); err != nil {
			return nil, fmt.Errorf("schedule validation failed: %w", err)
		}
	}

	// Salva
	if err := m.ruleRepo.Save(rule); err != nil {
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	m.logger.Info("follow-up rule created",
		"ruleID", rule.ID(),
		"name", rule.Name(),
		"pipelineID", input.PipelineID,
		"trigger", input.Trigger,
	)

	return rule, nil
}

// UpdateRuleInput input para atualizar regra
type UpdateRuleInput struct {
	Name        *string
	Description *string
	Conditions  []pipeline.RuleCondition
	Actions     []pipeline.RuleAction
	Priority    *int
	Enabled     *bool
}

// UpdateRule atualiza regra existente
func (m *AutomationRuleManager) UpdateRule(ctx context.Context, ruleID uuid.UUID, input UpdateRuleInput) (*pipeline.Automation, error) {
	// Busca regra
	rule, err := m.ruleRepo.FindByID(ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find rule: %w", err)
	}
	if rule == nil {
		return nil, errors.New("rule not found")
	}

	// Aplica mudanças
	if input.Description != nil {
		rule.UpdateDescription(*input.Description)
	}

	if len(input.Conditions) > 0 {
		rule.SetConditions(input.Conditions)
	}

	if len(input.Actions) > 0 {
		rule.SetActions(input.Actions)
	}

	if input.Priority != nil {
		if err := rule.SetPriority(*input.Priority); err != nil {
			return nil, err
		}
	}

	if input.Enabled != nil {
		if *input.Enabled {
			rule.Enable()
		} else {
			rule.Disable()
		}
	}

	// Valida
	if err := m.validator.ValidateRule(rule); err != nil {
		return nil, fmt.Errorf("rule validation failed: %w", err)
	}

	// Salva
	if err := m.ruleRepo.Save(rule); err != nil {
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	m.logger.Info("follow-up rule updated", "ruleID", ruleID)

	return rule, nil
}

// DeleteRule deleta regra
func (m *AutomationRuleManager) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	if err := m.ruleRepo.Delete(ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	m.logger.Info("follow-up rule deleted", "ruleID", ruleID)
	return nil
}

// GetRule busca regra por ID
func (m *AutomationRuleManager) GetRule(ctx context.Context, ruleID uuid.UUID) (*pipeline.Automation, error) {
	rule, err := m.ruleRepo.FindByID(ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find rule: %w", err)
	}
	if rule == nil {
		return nil, errors.New("rule not found")
	}
	return rule, nil
}

// ListRulesByPipeline lista todas as regras de um pipeline
func (m *AutomationRuleManager) ListRulesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Automation, error) {
	rules, err := m.ruleRepo.FindByPipeline(pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to find rules: %w", err)
	}
	return rules, nil
}

// ListEnabledRules lista apenas regras ativas de um pipeline
func (m *AutomationRuleManager) ListEnabledRules(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Automation, error) {
	rules, err := m.ruleRepo.FindEnabledByPipeline(pipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to find enabled rules: %w", err)
	}
	return rules, nil
}

// EnableRule ativa regra
func (m *AutomationRuleManager) EnableRule(ctx context.Context, ruleID uuid.UUID) error {
	rule, err := m.ruleRepo.FindByID(ruleID)
	if err != nil {
		return err
	}
	if rule == nil {
		return errors.New("rule not found")
	}

	rule.Enable()

	if err := m.ruleRepo.Save(rule); err != nil {
		return err
	}

	m.logger.Info("rule enabled", "ruleID", ruleID)
	return nil
}

// DisableRule desativa regra
func (m *AutomationRuleManager) DisableRule(ctx context.Context, ruleID uuid.UUID) error {
	rule, err := m.ruleRepo.FindByID(ruleID)
	if err != nil {
		return err
	}
	if rule == nil {
		return errors.New("rule not found")
	}

	rule.Disable()

	if err := m.ruleRepo.Save(rule); err != nil {
		return err
	}

	m.logger.Info("rule disabled", "ruleID", ruleID)
	return nil
}

// BulkEnableRules ativa múltiplas regras
func (m *AutomationRuleManager) BulkEnableRules(ctx context.Context, ruleIDs []uuid.UUID) error {
	for _, ruleID := range ruleIDs {
		if err := m.EnableRule(ctx, ruleID); err != nil {
			m.logger.Error("failed to enable rule in bulk", "ruleID", ruleID, "error", err)
			// Continua para próximas
		}
	}
	return nil
}

// BulkDisableRules desativa múltiplas regras
func (m *AutomationRuleManager) BulkDisableRules(ctx context.Context, ruleIDs []uuid.UUID) error {
	for _, ruleID := range ruleIDs {
		if err := m.DisableRule(ctx, ruleID); err != nil {
			m.logger.Error("failed to disable rule in bulk", "ruleID", ruleID, "error", err)
		}
	}
	return nil
}

// DuplicateRule duplica uma regra existente
func (m *AutomationRuleManager) DuplicateRule(ctx context.Context, sourceRuleID uuid.UUID, newName string) (*pipeline.Automation, error) {
	// Busca regra fonte
	source, err := m.ruleRepo.FindByID(sourceRuleID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, errors.New("source rule not found")
	}

	// Cria nova regra com mesmos dados
	pipelineID := source.PipelineID()
	if pipelineID == nil || *pipelineID == uuid.Nil {
		return nil, errors.New("source rule has no pipeline ID")
	}

	input := CreateRuleInput{
		PipelineID:  *pipelineID, // dereference pointer
		TenantID:    source.TenantID(),
		Name:        newName,
		Description: source.Description() + " (copy)",
		Trigger:     source.Trigger(),
		Conditions:  source.Conditions(),
		Actions:     source.Actions(),
		Priority:    source.Priority() + 1, // incrementa prioridade
		Enabled:     false,                 // cria desabilitada
	}

	return m.CreateRule(ctx, input)
}

// ExportRule exporta regra como JSON
func (m *AutomationRuleManager) ExportRule(ctx context.Context, ruleID uuid.UUID) (string, error) {
	rule, err := m.ruleRepo.FindByID(ruleID)
	if err != nil {
		return "", err
	}
	if rule == nil {
		return "", errors.New("rule not found")
	}

	export := map[string]interface{}{
		"name":        rule.Name(),
		"description": rule.Description(),
		"trigger":     rule.Trigger(),
		"conditions":  rule.Conditions(),
		"actions":     rule.Actions(),
		"priority":    rule.Priority(),
	}

	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal rule: %w", err)
	}

	return string(jsonData), nil
}

// ImportRuleInput input para importar regra
type ImportRuleInput struct {
	PipelineID uuid.UUID
	TenantID   string
	RuleJSON   string
	Enabled    bool
}

// ImportRule importa regra de JSON
func (m *AutomationRuleManager) ImportRule(ctx context.Context, input ImportRuleInput) (*pipeline.Automation, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(input.RuleJSON), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Extrai campos
	name, _ := data["name"].(string)
	description, _ := data["description"].(string)
	trigger, _ := data["trigger"].(string)
	priority, _ := data["priority"].(float64)

	// Deserializa conditions
	var conditions []pipeline.RuleCondition
	if condData, ok := data["conditions"].([]interface{}); ok {
		condJSON, _ := json.Marshal(condData)
		json.Unmarshal(condJSON, &conditions)
	}

	// Deserializa actions
	var actions []pipeline.RuleAction
	if actData, ok := data["actions"].([]interface{}); ok {
		actJSON, _ := json.Marshal(actData)
		json.Unmarshal(actJSON, &actions)
	}

	// Cria regra
	createInput := CreateRuleInput{
		PipelineID:  input.PipelineID,
		TenantID:    input.TenantID,
		Name:        name,
		Description: description,
		Trigger:     pipeline.AutomationTrigger(trigger),
		Conditions:  conditions,
		Actions:     actions,
		Priority:    int(priority),
		Enabled:     input.Enabled,
	}

	return m.CreateRule(ctx, createInput)
}

// ReorderRules reordena prioridades de regras
func (m *AutomationRuleManager) ReorderRules(ctx context.Context, pipelineID uuid.UUID, ruleOrder []uuid.UUID) error {
	// Busca todas as regras do pipeline
	rules, err := m.ruleRepo.FindByPipeline(pipelineID)
	if err != nil {
		return err
	}

	// Cria mapa ruleID -> regra
	ruleMap := make(map[uuid.UUID]*pipeline.Automation)
	for _, rule := range rules {
		ruleMap[rule.ID()] = rule
	}

	// Aplica nova ordem
	for newPriority, ruleID := range ruleOrder {
		if rule, exists := ruleMap[ruleID]; exists {
			if err := rule.SetPriority(newPriority); err != nil {
				m.logger.Error("failed to set priority", "ruleID", ruleID, "error", err)
				continue
			}

			if err := m.ruleRepo.Save(rule); err != nil {
				m.logger.Error("failed to save rule with new priority", "ruleID", ruleID, "error", err)
			}
		}
	}

	m.logger.Info("rules reordered", "pipelineID", pipelineID, "count", len(ruleOrder))
	return nil
}

// GetRuleStatistics retorna estatísticas de regras
func (m *AutomationRuleManager) GetRuleStatistics(ctx context.Context, pipelineID uuid.UUID) (*RuleStatistics, error) {
	rules, err := m.ruleRepo.FindByPipeline(pipelineID)
	if err != nil {
		return nil, err
	}

	stats := &RuleStatistics{
		Total:             len(rules),
		Enabled:           0,
		Disabled:          0,
		ByTrigger:         make(map[string]int),
		AverageConditions: 0.0,
		AverageActions:    0.0,
	}

	totalConditions := 0
	totalActions := 0

	for _, rule := range rules {
		if rule.IsEnabled() {
			stats.Enabled++
		} else {
			stats.Disabled++
		}

		triggerKey := string(rule.Trigger())
		stats.ByTrigger[triggerKey]++

		totalConditions += len(rule.Conditions())
		totalActions += len(rule.Actions())
	}

	if len(rules) > 0 {
		stats.AverageConditions = float64(totalConditions) / float64(len(rules))
		stats.AverageActions = float64(totalActions) / float64(len(rules))
	}

	return stats, nil
}

// RuleStatistics estatísticas de regras
type RuleStatistics struct {
	Total             int
	Enabled           int
	Disabled          int
	ByTrigger         map[string]int
	AverageConditions float64
	AverageActions    float64
}

// DefaultRuleValidator validador padrão
type DefaultRuleValidator struct {
	triggerRegistry *pipeline.TriggerRegistry
}

// NewDefaultRuleValidator cria validador com registry
func NewDefaultRuleValidator(triggerRegistry *pipeline.TriggerRegistry) *DefaultRuleValidator {
	if triggerRegistry == nil {
		triggerRegistry = pipeline.NewTriggerRegistry()
	}
	return &DefaultRuleValidator{
		triggerRegistry: triggerRegistry,
	}
}

// ValidateRule valida regra
func (v *DefaultRuleValidator) ValidateRule(rule *pipeline.Automation) error {
	// Valida trigger
	if !v.triggerRegistry.IsValidTrigger(string(rule.Trigger())) {
		return fmt.Errorf("invalid trigger: %s (not registered)", rule.Trigger())
	}

	// Valida actions
	if len(rule.Actions()) == 0 {
		return errors.New("rule must have at least one action")
	}

	// Valida cada condição
	for i, condition := range rule.Conditions() {
		if condition.Field == "" {
			return fmt.Errorf("condition %d: field cannot be empty", i)
		}
		if condition.Operator == "" {
			return fmt.Errorf("condition %d: operator cannot be empty", i)
		}
	}

	// Valida cada ação
	for i, action := range rule.Actions() {
		if action.Type == "" {
			return fmt.Errorf("action %d: type cannot be empty", i)
		}
		if action.Delay < 0 {
			return fmt.Errorf("action %d: delay cannot be negative", i)
		}
	}

	return nil
}

// ValidateSchedule valida configuração de schedule
func (v *DefaultRuleValidator) ValidateSchedule(schedule *pipeline.ScheduledRuleConfig) error {
	return schedule.Validate()
}

// TestRuleConditions testa condições de uma regra contra contexto de teste
func (m *AutomationRuleManager) TestRuleConditions(
	ctx context.Context,
	ruleID uuid.UUID,
	testContext map[string]interface{},
) (bool, error) {
	rule, err := m.ruleRepo.FindByID(ruleID)
	if err != nil {
		return false, err
	}
	if rule == nil {
		return false, errors.New("rule not found")
	}

	result := rule.EvaluateConditions(testContext)

	m.logger.Info("rule conditions tested",
		"ruleID", ruleID,
		"result", result,
		"conditionsCount", len(rule.Conditions()),
	)

	return result, nil
}

// ScheduleRuleExecution agenda execução manual de uma regra
func (m *AutomationRuleManager) ScheduleRuleExecution(
	ctx context.Context,
	ruleID uuid.UUID,
	executeAt time.Time,
) error {
	// TODO: Integrar com Temporal para agendamento
	m.logger.Info("rule execution scheduled",
		"ruleID", ruleID,
		"executeAt", executeAt,
	)

	return nil
}
