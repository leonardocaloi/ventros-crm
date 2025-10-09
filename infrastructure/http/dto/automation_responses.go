package dto

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/pipeline"
	"github.com/google/uuid"
)

// AutomationTypeResponse representa um tipo de automação
type AutomationTypeResponse struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Icon        string   `json:"icon,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

// TriggerResponse representa um trigger disponível
type TriggerResponse struct {
	Code        string             `json:"code"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	IsSystem    bool               `json:"is_system"`
	Parameters  []TriggerParameter `json:"parameters,omitempty"`
}

// TriggerParameter descreve um parâmetro disponível no contexto do trigger
type TriggerParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// ActionResponse representa uma ação disponível
type ActionResponse struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Parameters  []ActionParameterDTO   `json:"parameters"`
	Example     map[string]interface{} `json:"example,omitempty"`
}

// ActionParameterDTO descreve um parâmetro de ação
type ActionParameterDTO struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// ConditionOperatorResponse representa um operador de condição
type ConditionOperatorResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example,omitempty"`
}

// AutomationDiscoveryResponse resposta completa de discovery
type AutomationDiscoveryResponse struct {
	Types      []AutomationTypeResponse    `json:"types"`
	Triggers   []TriggerResponse           `json:"triggers"`
	Actions    []ActionResponse            `json:"actions"`
	Operators  []ConditionOperatorResponse `json:"operators"`
	LogicTypes []LogicOperatorResponse     `json:"logic_types"`
}

// LogicOperatorResponse descreve operadores lógicos (AND/OR)
type LogicOperatorResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AutomationResponse representa uma regra de automação
type AutomationResponse struct {
	ID          uuid.UUID                `json:"id"`
	PipelineID  *uuid.UUID               `json:"pipeline_id,omitempty"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Trigger     string                   `json:"trigger"`
	Conditions  []pipeline.RuleCondition `json:"conditions"`
	Actions     []pipeline.RuleAction    `json:"actions"`
	Priority    int                      `json:"priority"`
	Enabled     bool                     `json:"enabled"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// GetAutomationTypes retorna todos os tipos de automação
func GetAutomationTypes() []AutomationTypeResponse {
	return []AutomationTypeResponse{
		{
			Code:        string(pipeline.AutomationTypeFollowUp),
			Name:        "Follow-up",
			Description: "Automações de acompanhamento após inatividade ou falta de resposta",
			Icon:        "clock",
			Examples: []string{
				"Enviar mensagem após 24h sem resposta",
				"Lembrete de pagamento pendente",
				"Recuperação de carrinho abandonado",
			},
		},
		{
			Code:        string(pipeline.AutomationTypeEvent),
			Name:        "Evento",
			Description: "Automações disparadas por eventos específicos do sistema",
			Icon:        "zap",
			Examples: []string{
				"Confirmação de compra imediata",
				"Notificação de mudança de status",
				"Alerta de nova mensagem recebida",
			},
		},
		{
			Code:        string(pipeline.AutomationTypeScheduled),
			Name:        "Agendada",
			Description: "Automações recorrentes ou agendadas para horários específicos",
			Icon:        "calendar",
			Examples: []string{
				"Newsletter semanal às segundas 10h",
				"Relatório mensal automático",
				"Lembrete diário de tarefas",
			},
		},
		{
			Code:        string(pipeline.AutomationTypeReengagement),
			Name:        "Reativação",
			Description: "Automações para reengajar clientes inativos ou em churn",
			Icon:        "refresh",
			Examples: []string{
				"Cliente voltou para Lead após 90 dias",
				"Campanha de reativação para inativos",
				"Oferta especial para ex-clientes",
			},
		},
		{
			Code:        string(pipeline.AutomationTypeOnboarding),
			Name:        "Onboarding",
			Description: "Automações de boas-vindas e integração de novos contatos",
			Icon:        "user-plus",
			Examples: []string{
				"Mensagem de boas-vindas ao novo lead",
				"Sequência de onboarding em 5 dias",
				"Tutorial automático do produto",
			},
		},
		{
			Code:        string(pipeline.AutomationTypeCustom),
			Name:        "Customizada",
			Description: "Automações personalizadas com lógica específica do seu negócio",
			Icon:        "settings",
			Examples: []string{
				"Fluxo específico de aprovação",
				"Integrações customizadas",
				"Lógica de negócio única",
			},
		},
	}
}

// GetLogicOperators retorna operadores lógicos disponíveis
func GetLogicOperators() []LogicOperatorResponse {
	return []LogicOperatorResponse{
		{
			Code:        string(pipeline.LogicAND),
			Name:        "E (AND)",
			Description: "Todas as condições devem ser verdadeiras",
		},
		{
			Code:        string(pipeline.LogicOR),
			Name:        "OU (OR)",
			Description: "Pelo menos uma condição deve ser verdadeira",
		},
	}
}

// ToConditionOperatorResponse converte domain para DTO
func ToConditionOperatorResponse(op pipeline.ConditionOperator) ConditionOperatorResponse {
	return ConditionOperatorResponse{
		Code:        op.Code,
		Name:        op.Name,
		Description: op.Description,
		Example:     op.Example,
	}
}

// ToActionResponse converte domain para DTO
func ToActionResponse(action pipeline.ActionMetadata) ActionResponse {
	params := make([]ActionParameterDTO, len(action.Parameters))
	for i, p := range action.Parameters {
		params[i] = ActionParameterDTO{
			Name:        p.Name,
			Type:        p.Type,
			Required:    p.Required,
			Description: p.Description,
			Default:     p.Default,
		}
	}

	return ActionResponse{
		Code:        action.Code,
		Name:        action.Name,
		Description: action.Description,
		Category:    action.Category,
		Parameters:  params,
		Example:     action.Example,
	}
}

// ToAutomationResponse converte domain para DTO
func ToAutomationResponse(rule *pipeline.Automation) AutomationResponse {
	return AutomationResponse{
		ID:          rule.ID(),
		PipelineID:  rule.PipelineID(),
		Name:        rule.Name(),
		Description: rule.Description(),
		Trigger:     string(rule.Trigger()),
		Conditions:  rule.Conditions(),
		Actions:     rule.Actions(),
		Priority:    rule.Priority(),
		Enabled:     rule.IsEnabled(),
		CreatedAt:   rule.CreatedAt(),
		UpdatedAt:   rule.UpdatedAt(),
	}
}
