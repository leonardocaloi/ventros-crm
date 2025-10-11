package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	apierrors "github.com/caloi/ventros-crm/infrastructure/http/errors"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlerMiddleware handles errors from handlers and converts them to API responses
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Log the error with context
			logError(c, logger, err)

			// If response was already written, don't send another one
			if c.Writer.Written() {
				return
			}

			// Send error response
			apierrors.RespondWithError(c, err)
		}
	}
}

// RecoveryMiddleware recovers from panics and converts them to 500 errors
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger.Error("Panic recovered",
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)

				// Create internal error
				var panicErr error
				switch e := err.(type) {
				case error:
					panicErr = e
				case string:
					panicErr = errors.New(e)
				default:
					panicErr = fmt.Errorf("panic: %v", e)
				}

				internalErr := shared.NewInternalError("Internal server error", panicErr)

				// Send error response
				apierrors.RespondWithError(c, internalErr)

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}

// logError logs the error with appropriate level and context
func logError(c *gin.Context, logger *zap.Logger, err error) {
	// Extract request context
	fields := []zap.Field{
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("client_ip", c.ClientIP()),
	}

	// Add request ID if available
	if requestID, exists := c.Get("request_id"); exists {
		if reqID, ok := requestID.(string); ok {
			fields = append(fields, zap.String("request_id", reqID))
		}
	}

	// Add correlation ID if available
	if correlationID, exists := c.Get("correlation_id"); exists {
		if corrID, ok := correlationID.(string); ok {
			fields = append(fields, zap.String("correlation_id", corrID))
		}
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		fields = append(fields, zap.Any("user_id", userID))
	}
	if tenantID, exists := c.Get("tenant_id"); exists {
		fields = append(fields, zap.Any("tenant_id", tenantID))
	}

	// Check error type and log with appropriate level
	var domainErr *shared.DomainError
	if shared.IsDomainError(err, &domainErr) {
		fields = append(fields,
			zap.String("error_type", string(domainErr.Type)),
			zap.String("error_code", domainErr.Code),
		)

		if domainErr.Resource != "" {
			fields = append(fields, zap.String("resource", domainErr.Resource))
		}
		if domainErr.ResourceID != "" {
			fields = append(fields, zap.String("resource_id", domainErr.ResourceID))
		}
		if domainErr.Field != "" {
			fields = append(fields, zap.String("field", domainErr.Field))
		}

		// Log level based on error type
		switch domainErr.Type {
		case shared.ErrorTypeValidation,
			shared.ErrorTypeNotFound,
			shared.ErrorTypeBadRequest,
			shared.ErrorTypeAlreadyExists:
			// Client errors - log as info/warn
			logger.Info("Client error", append(fields, zap.Error(err))...)

		case shared.ErrorTypeForbidden,
			shared.ErrorTypeUnauthorized:
			// Auth errors - log as warning
			logger.Warn("Authorization error", append(fields, zap.Error(err))...)

		case shared.ErrorTypeConflict,
			shared.ErrorTypePrecondition,
			shared.ErrorTypeInvariantViolation:
			// Business logic errors - log as warning
			logger.Warn("Business logic error", append(fields, zap.Error(err))...)

		case shared.ErrorTypeRateLimit:
			// Rate limit - log as info
			logger.Info("Rate limit exceeded", append(fields, zap.Error(err))...)

		case shared.ErrorTypeTimeout:
			// Timeout - log as warning
			logger.Warn("Operation timeout", append(fields, zap.Error(err))...)

		case shared.ErrorTypeDatabase,
			shared.ErrorTypeCache,
			shared.ErrorTypeMessaging,
			shared.ErrorTypeExternal,
			shared.ErrorTypeNetwork,
			shared.ErrorTypeInternal:
			// Infrastructure/system errors - log as error
			logger.Error("System error", append(fields, zap.Error(err))...)

		default:
			logger.Error("Unknown error", append(fields, zap.Error(err))...)
		}
	} else {
		// Non-domain errors - log as error
		logger.Error("Unexpected error", append(fields, zap.Error(err))...)
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := shared.NewNotFoundError("endpoint", c.Request.URL.Path)
		err.Message = fmt.Sprintf("Endpoint not found: %s %s", c.Request.Method, c.Request.URL.Path)
		apierrors.RespondWithError(c, err)
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := shared.NewBadRequestError(
			fmt.Sprintf("Method %s not allowed for endpoint %s", c.Request.Method, c.Request.URL.Path),
		)
		err.Code = "METHOD_NOT_ALLOWED"

		c.JSON(http.StatusMethodNotAllowed, apierrors.NewAPIError(err))
	}
}
