package health

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Status     Status                 `json:"status" example:"healthy"`
	Message    string                 `json:"message,omitempty" example:"database is operational"`
	Timestamp  time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	DurationMs int64                  `json:"duration_ms" example:"15"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// HealthChecker provides health check functionality for all dependencies
type HealthChecker struct {
	db             *sql.DB
	redisClient    *redis.Client
	rabbitmqURL    string
	temporalClient client.Client
	gormDB         *gorm.DB
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(
	db *sql.DB,
	redisClient *redis.Client,
	rabbitmqURL string,
	temporalClient client.Client,
	gormDB *gorm.DB,
) *HealthChecker {
	return &HealthChecker{
		db:             db,
		redisClient:    redisClient,
		rabbitmqURL:    rabbitmqURL,
		temporalClient: temporalClient,
		gormDB:         gormDB,
	}
}

// CheckDatabase verifies database connectivity and status
func (hc *HealthChecker) CheckDatabase(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// Check if DB is nil
	if hc.db == nil {
		result.Status = StatusUnhealthy
		result.Message = "database not configured"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	// Ping database
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := hc.db.PingContext(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("database ping failed: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	// Get database stats
	stats := hc.db.Stats()
	result.Metadata["open_connections"] = stats.OpenConnections
	result.Metadata["in_use"] = stats.InUse
	result.Metadata["idle"] = stats.Idle
	result.Metadata["max_open_connections"] = stats.MaxOpenConnections

	result.Status = StatusHealthy
	result.Message = "database is operational"
	result.DurationMs = time.Since(start).Milliseconds()

	return result
}

// CheckMigrations verifies database migration status
func (hc *HealthChecker) CheckMigrations(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	if hc.db == nil {
		result.Status = StatusUnhealthy
		result.Message = "database not configured"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	if hc.gormDB == nil {
		result.Status = StatusUnhealthy
		result.Message = "gorm db not configured"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check if main tables exist (sessions table as indicator)
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name IN ('sessions', 'contacts', 'messages')
		)
	`
	if err := hc.db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("migration check failed: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	if !exists {
		result.Status = StatusDegraded
		result.Message = "schema not initialized - run migrations"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	result.Status = StatusHealthy
	result.Message = "database schema is up to date"
	result.DurationMs = time.Since(start).Milliseconds()

	return result
}

// CheckRedis verifies Redis connectivity
func (hc *HealthChecker) CheckRedis(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	if hc.redisClient == nil {
		result.Status = StatusUnhealthy
		result.Message = "redis not configured"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Ping Redis
	if err := hc.redisClient.Ping(ctx).Err(); err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("redis ping failed: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	// Get Redis info
	if err := hc.redisClient.Info(ctx, "server").Err(); err == nil {
		result.Metadata["info"] = "available"
	}

	// Get database size
	if dbSize, err := hc.redisClient.DBSize(ctx).Result(); err == nil {
		result.Metadata["keys"] = dbSize
	}

	result.Status = StatusHealthy
	result.Message = "redis is operational"
	result.DurationMs = time.Since(start).Milliseconds()

	return result
}

// CheckRabbitMQ verifies RabbitMQ connectivity and queue setup
func (hc *HealthChecker) CheckRabbitMQ(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	if hc.rabbitmqURL == "" {
		result.Status = StatusUnhealthy
		result.Message = "rabbitmq not configured"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	// Create connection with timeout
	done := make(chan error, 1)
	var conn *amqp.Connection

	go func() {
		var err error
		conn, err = amqp.Dial(hc.rabbitmqURL)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			result.Status = StatusUnhealthy
			result.Message = fmt.Sprintf("rabbitmq connection failed: %v", err)
			result.DurationMs = time.Since(start).Milliseconds()
			return result
		}
		defer conn.Close()

		// Try to open a channel
		ch, err := conn.Channel()
		if err != nil {
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("rabbitmq channel creation failed: %v", err)
			result.DurationMs = time.Since(start).Milliseconds()
			return result
		}
		defer ch.Close()

		// Check if critical queues exist
		criticalQueues := []string{
			"waha.events.message",
			"waha.events.ack", 
			"domain.events.contact.created",
			"domain.events.session.started",
		}

		var existingQueues []string
		var missingQueues []string

		for _, queueName := range criticalQueues {
			if _, err := ch.QueueInspect(queueName); err != nil {
				missingQueues = append(missingQueues, queueName)
			} else {
				existingQueues = append(existingQueues, queueName)
			}
		}

		result.Metadata["existing_queues"] = existingQueues
		result.Metadata["existing_count"] = len(existingQueues)

		if len(missingQueues) > 0 {
			result.Status = StatusDegraded
			result.Message = fmt.Sprintf("rabbitmq operational but %d critical queues missing", len(missingQueues))
			result.Metadata["missing_queues"] = missingQueues
			result.Metadata["missing_count"] = len(missingQueues)
		} else {
			result.Status = StatusHealthy
			result.Message = "rabbitmq operational with all critical queues"
		}

		result.DurationMs = time.Since(start).Milliseconds()

	case <-time.After(2 * time.Second):
		result.Status = StatusUnhealthy
		result.Message = "rabbitmq connection timeout"
		result.DurationMs = time.Since(start).Milliseconds()
	}

	return result
}

// CheckTemporal verifies Temporal connectivity
func (hc *HealthChecker) CheckTemporal(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	if hc.temporalClient == nil {
		result.Status = StatusUnhealthy
		result.Message = "temporal not configured"
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check Temporal health by checking workflow service
	_, err := hc.temporalClient.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("temporal health check failed: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	result.Status = StatusHealthy
	result.Message = "temporal is operational"
	result.DurationMs = time.Since(start).Milliseconds()

	return result
}

// CheckAll performs all health checks
func (hc *HealthChecker) CheckAll(ctx context.Context) map[string]CheckResult {
	results := make(map[string]CheckResult)

	results["database"] = hc.CheckDatabase(ctx)
	results["migrations"] = hc.CheckMigrations(ctx)
	results["redis"] = hc.CheckRedis(ctx)
	results["rabbitmq"] = hc.CheckRabbitMQ(ctx)
	results["temporal"] = hc.CheckTemporal(ctx)

	return results
}

// GetOverallStatus determines the overall system status
func GetOverallStatus(results map[string]CheckResult) Status {
	hasUnhealthy := false
	hasDegraded := false

	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}
