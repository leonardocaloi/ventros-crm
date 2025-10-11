package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCircuitBreaker_Success(t *testing.T) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "test-success",
		MaxRequests: 3,
		Interval:    1 * time.Second,
		Timeout:     1 * time.Second,
	}

	cb := NewCircuitBreaker(config, logger)

	// Deve executar com sucesso
	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, gobreaker.StateClosed, cb.State())
}

func TestCircuitBreaker_Failure(t *testing.T) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "test-failure",
		MaxRequests: 1,
		Interval:    1 * time.Second,
		Timeout:     1 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	cb := NewCircuitBreaker(config, logger)

	// Simula 3 falhas consecutivas
	for i := 0; i < 3; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, errors.New("simulated error")
		})
		assert.Error(t, err)
	}

	// Circuit breaker deve estar OPEN
	assert.Equal(t, gobreaker.StateOpen, cb.State())

	// Próxima execução deve falhar imediatamente
	_, err := cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})
	assert.Error(t, err)
	assert.Equal(t, gobreaker.ErrOpenState, err)
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "test-half-open",
		MaxRequests: 2,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	}

	cb := NewCircuitBreaker(config, logger)

	// Simula 2 falhas para abrir o circuito
	for i := 0; i < 2; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, errors.New("error")
		})
	}

	assert.Equal(t, gobreaker.StateOpen, cb.State())

	// Aguarda timeout para ir para half-open
	time.Sleep(150 * time.Millisecond)

	// Próxima execução deve funcionar (half-open state)
	result, err := cb.Execute(func() (interface{}, error) {
		return "recovered", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "recovered", result)
}

func TestCircuitBreaker_WithContext(t *testing.T) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "test-context",
		MaxRequests: 3,
	}

	cb := NewCircuitBreaker(config, logger)

	// Context normal
	ctx := context.Background()
	result, err := cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
		return "with context", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "with context", result)

	// Context cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela imediatamente

	_, err = cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, nil
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestCircuitBreakerManager(t *testing.T) {
	logger := zap.NewNop()
	manager := NewCircuitBreakerManager(logger)

	// Registra circuit breaker
	config := CircuitBreakerConfig{
		MaxRequests: 5,
	}
	cb1 := manager.Register("database", config)

	assert.NotNil(t, cb1)
	assert.Equal(t, "database", cb1.Name())

	// Get circuit breaker existente
	cb2, err := manager.Get("database")
	assert.NoError(t, err)
	assert.Equal(t, cb1, cb2)

	// Get circuit breaker inexistente
	_, err = manager.Get("nonexistent")
	assert.Error(t, err)

	// GetOrCreate
	cb3 := manager.GetOrCreate("rabbitmq")
	assert.NotNil(t, cb3)
	assert.Equal(t, "rabbitmq", cb3.Name())

	// HealthStatus
	status := manager.HealthStatus()
	assert.Len(t, status, 2) // database + rabbitmq
	assert.Contains(t, status, "database")
	assert.Contains(t, status, "rabbitmq")
}

func TestCircuitBreaker_StateChanges(t *testing.T) {
	stateChanges := []string{}

	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "test-state-changes",
		MaxRequests: 1,
		Interval:    100 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			stateChanges = append(stateChanges, from.String()+"->"+to.String())
		},
	}

	cb := NewCircuitBreaker(config, logger)

	// Inicial: Closed
	assert.Equal(t, gobreaker.StateClosed, cb.State())

	// 2 falhas → Open
	for i := 0; i < 2; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, errors.New("error")
		})
	}

	assert.Equal(t, gobreaker.StateOpen, cb.State())
	assert.Contains(t, stateChanges, "closed->open")

	// Aguarda timeout → Half-Open
	time.Sleep(150 * time.Millisecond)

	cb.Execute(func() (interface{}, error) {
		return "ok", nil
	})

	// Deve ter ido para half-open e depois closed
	assert.Contains(t, stateChanges, "open->half-open")
}

func TestCircuitBreaker_Counts(t *testing.T) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "test-counts",
		MaxRequests: 10,
	}

	cb := NewCircuitBreaker(config, logger)

	// 3 sucessos
	for i := 0; i < 3; i++ {
		cb.Execute(func() (interface{}, error) {
			return "ok", nil
		})
	}

	// 2 falhas
	for i := 0; i < 2; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, errors.New("error")
		})
	}

	counts := cb.Counts()
	assert.Equal(t, uint32(5), counts.Requests)
	assert.Equal(t, uint32(3), counts.TotalSuccesses)
	assert.Equal(t, uint32(2), counts.TotalFailures)
}

// Benchmark do circuit breaker
func BenchmarkCircuitBreaker_Execute(b *testing.B) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "benchmark",
		MaxRequests: 100,
	}

	cb := NewCircuitBreaker(config, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(func() (interface{}, error) {
			return "ok", nil
		})
	}
}

func BenchmarkCircuitBreaker_ExecuteWithContext(b *testing.B) {
	logger := zap.NewNop()
	config := CircuitBreakerConfig{
		Name:        "benchmark-ctx",
		MaxRequests: 100,
	}

	cb := NewCircuitBreaker(config, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
			return "ok", nil
		})
	}
}
