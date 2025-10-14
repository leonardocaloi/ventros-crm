package messaging

import (
	"context"
	"time"

	"github.com/sony/gobreaker"
	"github.com/ventros/crm/infrastructure/resilience"
	"go.uber.org/zap"
)

// RabbitMQWithCircuitBreaker wrapper do RabbitMQConnection com circuit breaker
type RabbitMQWithCircuitBreaker struct {
	conn           *RabbitMQConnection
	circuitBreaker *resilience.CircuitBreaker
	logger         *zap.Logger
}

// NewRabbitMQWithCircuitBreaker cria uma conexão RabbitMQ com circuit breaker
func NewRabbitMQWithCircuitBreaker(conn *RabbitMQConnection, logger *zap.Logger) *RabbitMQWithCircuitBreaker {
	config := resilience.CircuitBreakerConfig{
		Name:        "rabbitmq",
		MaxRequests: 5,                // Permite 5 requests em half-open
		Interval:    60 * time.Second, // Reseta contadores a cada 60s
		Timeout:     30 * time.Second, // Volta para half-open após 30s
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Abre se 60% das requests falharem E tiver pelo menos 10 requests
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 10 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("RabbitMQ circuit breaker state changed",
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	cb := resilience.NewCircuitBreaker(config, logger)

	return &RabbitMQWithCircuitBreaker{
		conn:           conn,
		circuitBreaker: cb,
		logger:         logger,
	}
}

// Publish publica uma mensagem com circuit breaker
func (r *RabbitMQWithCircuitBreaker) Publish(ctx context.Context, queue string, body []byte) error {
	_, err := r.circuitBreaker.ExecuteWithContext(ctx, func(ctx context.Context) (interface{}, error) {
		return nil, r.conn.Publish(ctx, queue, body)
	})
	return err
}

// DeclareQueue declara uma fila com circuit breaker
func (r *RabbitMQWithCircuitBreaker) DeclareQueue(name string) error {
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, r.conn.DeclareQueue(name)
	})
	return err
}

// DeclareQueueWithDLQ declara uma fila com DLQ usando circuit breaker
func (r *RabbitMQWithCircuitBreaker) DeclareQueueWithDLQ(name string, maxRetries int) error {
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, r.conn.DeclareQueueWithDLQ(name, maxRetries)
	})
	return err
}

// StartConsumer inicia um consumer (não usa circuit breaker - é blocking)
// O consumer já tem sua própria resiliência via reconnect automático
func (r *RabbitMQWithCircuitBreaker) StartConsumer(
	ctx context.Context,
	queueName string,
	consumerTag string,
	consumer Consumer,
	prefetchCount int,
) error {
	// StartConsumer é blocking e gerencia sua própria resiliência via reconnect
	return r.conn.StartConsumer(ctx, queueName, consumerTag, consumer, prefetchCount)
}

// SetupAllQueues configura todas as queues com circuit breaker
func (r *RabbitMQWithCircuitBreaker) SetupAllQueues() error {
	_, err := r.circuitBreaker.Execute(func() (interface{}, error) {
		return nil, r.conn.SetupAllQueues()
	})
	return err
}

// Close fecha a conexão
func (r *RabbitMQWithCircuitBreaker) Close() error {
	return r.conn.Close()
}

// State retorna o estado do circuit breaker
func (r *RabbitMQWithCircuitBreaker) State() gobreaker.State {
	return r.circuitBreaker.State()
}

// IsHealthy verifica se o RabbitMQ está saudável (circuit closed ou half-open)
func (r *RabbitMQWithCircuitBreaker) IsHealthy() bool {
	state := r.circuitBreaker.State()
	return state == gobreaker.StateClosed || state == gobreaker.StateHalfOpen
}

// GetUnderlyingConnection retorna a conexão subjacente (para casos especiais)
func (r *RabbitMQWithCircuitBreaker) GetUnderlyingConnection() *RabbitMQConnection {
	return r.conn
}
