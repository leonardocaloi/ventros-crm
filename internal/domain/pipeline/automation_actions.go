package pipeline

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ActionExecutor é a interface que todos os executores de ações devem implementar
// Cada ação (create_note, send_message, webhook, etc) terá sua própria implementação
type ActionExecutor interface {
	// Execute executa a ação com os parâmetros fornecidos
	Execute(ctx context.Context, params ActionExecutionParams) error

	// Validate valida se os parâmetros são válidos para esta ação
	Validate(params map[string]interface{}) error

	// Type retorna o tipo de ação que este executor implementa
	Type() AutomationAction
}

// ActionExecutionParams contém os dados necessários para executar uma ação
type ActionExecutionParams struct {
	// Ação a ser executada
	Action RuleAction

	// Contexto da execução
	TenantID string
	RuleID   uuid.UUID
	RuleName string

	// Entidades relacionadas (opcionais, dependem do tipo de automação)
	ContactID  *uuid.UUID
	SessionID  *uuid.UUID
	AgentID    *uuid.UUID
	PipelineID *uuid.UUID
	MessageID  *uuid.UUID

	// Dados adicionais para interpolação de variáveis
	Variables map[string]interface{}
}

// ActionExecutorRegistry é um registro de todos os executores disponíveis
type ActionExecutorRegistry interface {
	// Register registra um executor para um tipo de ação
	Register(executor ActionExecutor) error

	// Get obtém o executor para um tipo de ação
	Get(actionType AutomationAction) (ActionExecutor, error)

	// Execute executa uma ação usando o executor apropriado
	Execute(ctx context.Context, params ActionExecutionParams) error
}

// ActionExecutionResult representa o resultado da execução de uma ação
type ActionExecutionResult struct {
	Success bool
	Error   error
	Message string
	Data    map[string]interface{} // dados adicionais retornados pela ação
}

// Common errors
var (
	ErrExecutorNotFound     = errors.New("action executor not found")
	ErrInvalidParams        = errors.New("invalid action parameters")
	ErrExecutionFailed      = errors.New("action execution failed")
	ErrMissingRequiredParam = errors.New("missing required parameter")
)

// VariableInterpolator é responsável por substituir variáveis em strings
// Exemplo: "{{contact_name}}" -> "João Silva"
type VariableInterpolator interface {
	// Interpolate substitui variáveis em uma string
	Interpolate(template string, variables map[string]interface{}) (string, error)

	// InterpolateMap substitui variáveis em todos os valores de um map
	InterpolateMap(data map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error)
}

// ActionExecutorFactory cria executores de ações com suas dependências
type ActionExecutorFactory interface {
	// CreateRegistry cria um novo registro com todos os executores disponíveis
	CreateRegistry() ActionExecutorRegistry

	// CreateExecutor cria um executor específico
	CreateExecutor(actionType AutomationAction) (ActionExecutor, error)
}
