package pipeline

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ActionExecutor interface {
	Execute(ctx context.Context, params ActionExecutionParams) error

	Validate(params map[string]interface{}) error

	Type() AutomationAction
}

type ActionExecutionParams struct {
	Action RuleAction

	TenantID string
	RuleID   uuid.UUID
	RuleName string

	ContactID  *uuid.UUID
	SessionID  *uuid.UUID
	AgentID    *uuid.UUID
	PipelineID *uuid.UUID
	MessageID  *uuid.UUID

	Variables map[string]interface{}
}

type ActionExecutorRegistry interface {
	Register(executor ActionExecutor) error

	Get(actionType AutomationAction) (ActionExecutor, error)

	Execute(ctx context.Context, params ActionExecutionParams) error
}

type ActionExecutionResult struct {
	Success bool
	Error   error
	Message string
	Data    map[string]interface{}
}

var (
	ErrExecutorNotFound     = errors.New("action executor not found")
	ErrInvalidParams        = errors.New("invalid action parameters")
	ErrExecutionFailed      = errors.New("action execution failed")
	ErrMissingRequiredParam = errors.New("missing required parameter")
)

type VariableInterpolator interface {
	Interpolate(template string, variables map[string]interface{}) (string, error)

	InterpolateMap(data map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error)
}

type ActionExecutorFactory interface {
	CreateRegistry() ActionExecutorRegistry

	CreateExecutor(actionType AutomationAction) (ActionExecutor, error)
}
