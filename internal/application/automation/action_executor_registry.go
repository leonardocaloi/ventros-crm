package automation

import (
	"context"
	"fmt"
	"sync"

	"github.com/caloi/ventros-crm/internal/domain/pipeline"
)

// actionExecutorRegistry implementa ActionExecutorRegistry
type actionExecutorRegistry struct {
	executors map[pipeline.AutomationAction]pipeline.ActionExecutor
	mu        sync.RWMutex
}

// NewActionExecutorRegistry cria um novo registro de executores
func NewActionExecutorRegistry() pipeline.ActionExecutorRegistry {
	return &actionExecutorRegistry{
		executors: make(map[pipeline.AutomationAction]pipeline.ActionExecutor),
	}
}

// Register registra um executor para um tipo de ação
func (r *actionExecutorRegistry) Register(executor pipeline.ActionExecutor) error {
	if executor == nil {
		return fmt.Errorf("executor cannot be nil")
	}

	actionType := executor.Type()
	if actionType == "" {
		return fmt.Errorf("executor type cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.executors[actionType]; exists {
		return fmt.Errorf("executor for action type %s already registered", actionType)
	}

	r.executors[actionType] = executor
	return nil
}

// Get obtém o executor para um tipo de ação
func (r *actionExecutorRegistry) Get(actionType pipeline.AutomationAction) (pipeline.ActionExecutor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	executor, exists := r.executors[actionType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", pipeline.ErrExecutorNotFound, actionType)
	}

	return executor, nil
}

// Execute executa uma ação usando o executor apropriado
func (r *actionExecutorRegistry) Execute(ctx context.Context, params pipeline.ActionExecutionParams) error {
	executor, err := r.Get(params.Action.Type)
	if err != nil {
		return err
	}

	// Valida parâmetros antes de executar
	if err := executor.Validate(params.Action.Params); err != nil {
		return fmt.Errorf("%w: %v", pipeline.ErrInvalidParams, err)
	}

	// Executa a ação
	if err := executor.Execute(ctx, params); err != nil {
		return fmt.Errorf("%w: %v", pipeline.ErrExecutionFailed, err)
	}

	return nil
}
