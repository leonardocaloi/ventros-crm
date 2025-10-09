package automation

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/google/uuid"
)

// AutomationService é responsável por executar automações
type AutomationService struct {
	automationRepository pipeline.AutomationRepository
	executorRegistry     pipeline.ActionExecutorRegistry
}

// NewAutomationService cria um novo serviço de automação
func NewAutomationService(
	automationRepository pipeline.AutomationRepository,
	executorRegistry pipeline.ActionExecutorRegistry,
) *AutomationService {
	return &AutomationService{
		automationRepository: automationRepository,
		executorRegistry:     executorRegistry,
	}
}

// ExecuteRuleParams contém os parâmetros para executar uma regra
type ExecuteRuleParams struct {
	RuleID     uuid.UUID
	TenantID   string
	ContactID  *uuid.UUID
	SessionID  *uuid.UUID
	AgentID    *uuid.UUID
	PipelineID *uuid.UUID
	MessageID  *uuid.UUID
	Context    map[string]interface{} // para avaliação de condições
	Variables  map[string]interface{} // para interpolação de variáveis
}

// ExecuteRule executa uma regra de automação específica
func (s *AutomationService) ExecuteRule(ctx context.Context, params ExecuteRuleParams) error {
	// Busca a regra
	rule, err := s.automationRepository.FindByID(params.RuleID)
	if err != nil {
		return fmt.Errorf("failed to find rule: %w", err)
	}

	// Verifica se a regra está ativa
	if !rule.IsEnabled() {
		return fmt.Errorf("rule %s is disabled", params.RuleID)
	}

	// Valida tenant
	if rule.TenantID() != params.TenantID {
		return fmt.Errorf("rule %s does not belong to tenant %s", params.RuleID, params.TenantID)
	}

	// Avalia condições
	if !rule.EvaluateConditions(params.Context) {
		log.Printf("Rule %s conditions not met, skipping execution", params.RuleID)
		return nil
	}

	// Executa todas as ações
	for i, action := range rule.Actions() {
		// Aplica delay se configurado
		if action.Delay > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(action.Delay) * time.Minute):
				// Delay aplicado
			}
		}

		// Prepara parâmetros de execução
		execParams := pipeline.ActionExecutionParams{
			Action:     action,
			TenantID:   params.TenantID,
			RuleID:     params.RuleID,
			RuleName:   rule.Name(),
			ContactID:  params.ContactID,
			SessionID:  params.SessionID,
			AgentID:    params.AgentID,
			PipelineID: params.PipelineID,
			MessageID:  params.MessageID,
			Variables:  params.Variables,
		}

		// Executa a ação
		if err := s.executorRegistry.Execute(ctx, execParams); err != nil {
			log.Printf("Failed to execute action %d of rule %s: %v", i, params.RuleID, err)
			// Continua executando as outras ações mesmo se uma falhar
			// TODO: Adicionar configuração para stop-on-error
		}
	}

	return nil
}

// ExecuteRulesForTrigger executa todas as regras ativas para um trigger específico
func (s *AutomationService) ExecuteRulesForTrigger(
	ctx context.Context,
	pipelineID uuid.UUID,
	trigger pipeline.AutomationTrigger,
	params ExecuteRuleParams,
) error {
	// Busca regras ativas para o trigger
	rules, err := s.automationRepository.FindByPipelineAndTrigger(pipelineID, trigger)
	if err != nil {
		return fmt.Errorf("failed to find rules: %w", err)
	}

	// Filtra apenas regras ativas
	activeRules := make([]*pipeline.Automation, 0)
	for _, rule := range rules {
		if rule.IsEnabled() {
			activeRules = append(activeRules, rule)
		}
	}

	// Ordena por prioridade (já vem ordenado do repository)
	// Executa cada regra
	for _, rule := range activeRules {
		params.RuleID = rule.ID()
		params.TenantID = rule.TenantID()

		if err := s.ExecuteRule(ctx, params); err != nil {
			log.Printf("Failed to execute rule %s: %v", rule.ID(), err)
			// Continua executando as outras regras
		}
	}

	return nil
}

// ExecuteGenericAutomation executa automações que não são de pipeline
// Por exemplo: scheduled_report, time_based_notification, etc
func (s *AutomationService) ExecuteGenericAutomation(
	ctx context.Context,
	params ExecuteRuleParams,
) error {
	return s.ExecuteRule(ctx, params)
}
