package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

// CircuitBreakerConfig configura o circuit breaker
type CircuitBreakerConfig struct {
	Name          string        // Nome do circuit breaker
	MaxRequests   uint32        // Máximo de requests em half-open state
	Interval      time.Duration // Intervalo para resetar contadores
	Timeout       time.Duration // Timeout para voltar de open → half-open
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// CircuitBreaker wrapper para gobreaker com logging e métricas
type CircuitBreaker struct {
	breaker *gobreaker.CircuitBreaker
	logger  *zap.Logger
	name    string
}

// NewCircuitBreaker cria um novo circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig, logger *zap.Logger) *CircuitBreaker {
	// Default config se não fornecido
	if config.MaxRequests == 0 {
		config.MaxRequests = 3
	}
	if config.Interval == 0 {
		config.Interval = 60 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Default ReadyToTrip: 5 falhas consecutivas
	if config.ReadyToTrip == nil {
		config.ReadyToTrip = func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		}
	}

	// Logging de mudanças de estado
	onStateChange := config.OnStateChange
	config.OnStateChange = func(name string, from gobreaker.State, to gobreaker.State) {
		logger.Warn("Circuit breaker state changed",
			zap.String("circuit_breaker", name),
			zap.String("from", from.String()),
			zap.String("to", to.String()),
		)
		if onStateChange != nil {
			onStateChange(name, from, to)
		}
	}

	settings := gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	}

	return &CircuitBreaker{
		breaker: gobreaker.NewCircuitBreaker(settings),
		logger:  logger,
		name:    config.Name,
	}
}

// Execute executa uma função com circuit breaker
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := cb.breaker.Execute(fn)
	if err != nil {
		cb.logger.Error("Circuit breaker execution failed",
			zap.String("circuit_breaker", cb.name),
			zap.Error(err),
		)
	}
	return result, err
}

// ExecuteWithContext executa uma função com context e circuit breaker
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	// Verifica se context já foi cancelado
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result, err := cb.breaker.Execute(func() (interface{}, error) {
		return fn(ctx)
	})

	if err != nil {
		cb.logger.Error("Circuit breaker execution failed",
			zap.String("circuit_breaker", cb.name),
			zap.Error(err),
		)
	}

	return result, err
}

// State retorna o estado atual do circuit breaker
func (cb *CircuitBreaker) State() gobreaker.State {
	return cb.breaker.State()
}

// Counts retorna as estatísticas do circuit breaker
func (cb *CircuitBreaker) Counts() gobreaker.Counts {
	return cb.breaker.Counts()
}

// Name retorna o nome do circuit breaker
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// CircuitBreakerManager gerencia múltiplos circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	logger   *zap.Logger
}

// NewCircuitBreakerManager cria um novo manager
func NewCircuitBreakerManager(logger *zap.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   logger,
	}
}

// Register registra um novo circuit breaker
func (m *CircuitBreakerManager) Register(name string, config CircuitBreakerConfig) *CircuitBreaker {
	config.Name = name
	cb := NewCircuitBreaker(config, m.logger)
	m.breakers[name] = cb
	return cb
}

// Get retorna um circuit breaker pelo nome
func (m *CircuitBreakerManager) Get(name string) (*CircuitBreaker, error) {
	cb, ok := m.breakers[name]
	if !ok {
		return nil, fmt.Errorf("circuit breaker not found: %s", name)
	}
	return cb, nil
}

// GetOrCreate retorna ou cria um circuit breaker com config padrão
func (m *CircuitBreakerManager) GetOrCreate(name string) *CircuitBreaker {
	if cb, ok := m.breakers[name]; ok {
		return cb
	}

	// Cria com config padrão
	config := CircuitBreakerConfig{
		Name:        name,
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
	}

	return m.Register(name, config)
}

// HealthStatus retorna o status de todos os circuit breakers
func (m *CircuitBreakerManager) HealthStatus() map[string]interface{} {
	status := make(map[string]interface{})

	for name, cb := range m.breakers {
		state := cb.State()
		counts := cb.Counts()

		status[name] = map[string]interface{}{
			"state":          state.String(),
			"requests":       counts.Requests,
			"total_failures": counts.TotalFailures,
			"consecutive_failures": counts.ConsecutiveFailures,
			"total_successes": counts.TotalSuccesses,
			"consecutive_successes": counts.ConsecutiveSuccesses,
		}
	}

	return status
}
