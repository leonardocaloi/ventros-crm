package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ventros/crm/infrastructure/health"
	"go.uber.org/zap"
)

type HealthHandler struct {
	logger  *zap.Logger
	checker *health.HealthChecker
}

func NewHealthHandler(logger *zap.Logger, checker *health.HealthChecker) *HealthHandler {
	return &HealthHandler{
		logger:  logger,
		checker: checker,
	}
}

type HealthResponse struct {
	Status       string                        `json:"status" example:"healthy"`
	Timestamp    time.Time                     `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Version      string                        `json:"version" example:"0.1.0"`
	Service      string                        `json:"service" example:"ventros-crm"`
	Dependencies map[string]health.CheckResult `json:"dependencies,omitempty"`
}

// Health godoc
//
//	@Summary		Health check
//	@Description	Check if the API is running (basic liveness check)
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	// Basic health check - just returns that the service is alive
	c.JSON(http.StatusOK, HealthResponse{
		Status:    string(health.StatusHealthy),
		Timestamp: time.Now(),
		Version:   "0.1.0",
		Service:   "ventros-crm",
	})
}

// Ready godoc
//
//	@Summary		Readiness check
//	@Description	Check if the API is ready to serve requests (includes all dependency checks)
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Success		503	{object}	HealthResponse
//	@Router			/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()

	// Check all dependencies
	results := h.checker.CheckAll(ctx)
	overallStatus := health.GetOverallStatus(results)

	statusCode := http.StatusOK
	if overallStatus == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == health.StatusDegraded {
		statusCode = http.StatusOK // Still considered ready, but with warnings
	}

	response := HealthResponse{
		Status:       string(overallStatus),
		Timestamp:    time.Now(),
		Version:      "0.1.0",
		Service:      "ventros-crm",
		Dependencies: results,
	}

	h.logger.Debug("Readiness check completed",
		zap.String("status", string(overallStatus)),
		zap.Int("status_code", statusCode),
	)

	c.JSON(statusCode, response)
}

// Live godoc
//
//	@Summary		Liveness check
//	@Description	Check if the API is alive (always returns 200 if service is running)
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	// Liveness check - just returns that the process is running
	// This should not check dependencies
	c.JSON(http.StatusOK, HealthResponse{
		Status:    string(health.StatusHealthy),
		Timestamp: time.Now(),
		Version:   "0.1.0",
		Service:   "ventros-crm",
	})
}

// CheckDatabase godoc
//
//	@Summary		Database health check
//	@Description	Check database connectivity and connection pool status
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	health.CheckResult
//	@Success		503	{object}	health.CheckResult
//	@Router			/health/database [get]
func (h *HealthHandler) CheckDatabase(c *gin.Context) {
	ctx := c.Request.Context()
	result := h.checker.CheckDatabase(ctx)

	statusCode := http.StatusOK
	if result.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, result)
}

// CheckMigrations godoc
//
//	@Summary		Database migrations health check
//	@Description	Check database migration status
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	health.CheckResult
//	@Success		503	{object}	health.CheckResult
//	@Router			/health/migrations [get]
func (h *HealthHandler) CheckMigrations(c *gin.Context) {
	ctx := c.Request.Context()
	result := h.checker.CheckMigrations(ctx)

	statusCode := http.StatusOK
	if result.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, result)
}

// CheckRedis godoc
//
//	@Summary		Redis health check
//	@Description	Check Redis connectivity and status
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	health.CheckResult
//	@Success		503	{object}	health.CheckResult
//	@Router			/health/redis [get]
func (h *HealthHandler) CheckRedis(c *gin.Context) {
	ctx := c.Request.Context()
	result := h.checker.CheckRedis(ctx)

	statusCode := http.StatusOK
	if result.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, result)
}

// CheckRabbitMQ godoc
//
//	@Summary		RabbitMQ health check
//	@Description	Check RabbitMQ connectivity
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	health.CheckResult
//	@Success		503	{object}	health.CheckResult
//	@Router			/health/rabbitmq [get]
func (h *HealthHandler) CheckRabbitMQ(c *gin.Context) {
	ctx := c.Request.Context()
	result := h.checker.CheckRabbitMQ(ctx)

	statusCode := http.StatusOK
	if result.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, result)
}

// CheckTemporal godoc
//
//	@Summary		Temporal health check
//	@Description	Check Temporal workflow engine connectivity
//	@Tags			health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	health.CheckResult
//	@Success		503	{object}	health.CheckResult
//	@Router			/health/temporal [get]
func (h *HealthHandler) CheckTemporal(c *gin.Context) {
	ctx := c.Request.Context()
	result := h.checker.CheckTemporal(ctx)

	statusCode := http.StatusOK
	if result.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, result)
}
