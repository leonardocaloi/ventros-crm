package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// CorrelationIDHeader is the HTTP header for correlation ID
	CorrelationIDHeader = "X-Correlation-ID"

	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey = "correlation_id"
)

// CorrelationIDMiddleware extracts or generates a correlation ID for request tracing
// The correlation ID is used to trace requests across services and through the event system
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get correlation ID from request header
		correlationID := c.GetHeader(CorrelationIDHeader)

		// If not present, generate a new UUID
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Set correlation ID in response header
		c.Header(CorrelationIDHeader, correlationID)

		// Store in Gin context
		c.Set(CorrelationIDKey, correlationID)

		// Store in request context for downstream use
		ctx := context.WithValue(c.Request.Context(), CorrelationIDKey, correlationID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}

	return ""
}

// GetCorrelationIDFromGin extracts correlation ID from Gin context
func GetCorrelationIDFromGin(c *gin.Context) string {
	if value, exists := c.Get(CorrelationIDKey); exists {
		if correlationID, ok := value.(string); ok {
			return correlationID
		}
	}
	return ""
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}
