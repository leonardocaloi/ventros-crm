package errors

import (
	"net/http"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

// APIError represents a standardized API error response
type APIError struct {
	// HTTP status code
	Status int `json:"status" example:"400"`

	// Error code for programmatic handling
	Code string `json:"code" example:"VALIDATION_FAILED"`

	// Human-readable error message
	Message string `json:"message" example:"Invalid request data"`

	// Detailed error information (optional, for debugging)
	Details map[string]interface{} `json:"details,omitempty"`

	// Field that caused the error (for validation errors)
	Field string `json:"field,omitempty" example:"email"`

	// Request ID for tracing
	RequestID string `json:"request_id,omitempty" example:"req-123456"`

	// Timestamp of the error
	Timestamp time.Time `json:"timestamp" example:"2025-10-10T15:30:00Z"`

	// Resource information (optional)
	Resource   string `json:"resource,omitempty" example:"contact"`
	ResourceID string `json:"resource_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Help URL (optional)
	Help string `json:"help,omitempty" example:"https://docs.ventros.com/errors/validation"`
}

// ErrorResponse wraps multiple errors (for validation errors with multiple fields)
type ErrorResponse struct {
	Error  *APIError   `json:"error"`
	Errors []*APIError `json:"errors,omitempty"`
}

// NewAPIError creates a new API error from a domain error
func NewAPIError(err error) *APIError {
	if err == nil {
		return nil
	}

	apiErr := &APIError{
		Timestamp: time.Now().UTC(),
	}

	// Check if it's a DomainError
	var domainErr *shared.DomainError
	if !shared.IsDomainError(err, &domainErr) {
		// Generic internal server error
		apiErr.Status = http.StatusInternalServerError
		apiErr.Code = "INTERNAL_ERROR"
		apiErr.Message = "An internal error occurred"
		return apiErr
	}

	// Map domain error to API error
	apiErr.Code = domainErr.Code
	apiErr.Message = domainErr.Message
	apiErr.Details = domainErr.Details
	apiErr.Field = domainErr.Field
	apiErr.Resource = domainErr.Resource
	apiErr.ResourceID = domainErr.ResourceID
	apiErr.Status = mapErrorTypeToHTTPStatus(domainErr.Type)

	return apiErr
}

// WithRequestID adds request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithHelp adds help URL to the error
func (e *APIError) WithHelp(helpURL string) *APIError {
	e.Help = helpURL
	return e
}

// mapErrorTypeToHTTPStatus maps domain error types to HTTP status codes
func mapErrorTypeToHTTPStatus(errorType shared.ErrorType) int {
	switch errorType {
	case shared.ErrorTypeValidation:
		return http.StatusBadRequest // 400
	case shared.ErrorTypeNotFound:
		return http.StatusNotFound // 404
	case shared.ErrorTypeAlreadyExists:
		return http.StatusConflict // 409
	case shared.ErrorTypeConflict:
		return http.StatusConflict // 409
	case shared.ErrorTypeForbidden:
		return http.StatusForbidden // 403
	case shared.ErrorTypeUnauthorized:
		return http.StatusUnauthorized // 401
	case shared.ErrorTypeBadRequest:
		return http.StatusBadRequest // 400
	case shared.ErrorTypePrecondition:
		return http.StatusPreconditionFailed // 412
	case shared.ErrorTypeInvariantViolation:
		return http.StatusUnprocessableEntity // 422
	case shared.ErrorTypeDatabase:
		return http.StatusInternalServerError // 500
	case shared.ErrorTypeCache:
		return http.StatusInternalServerError // 500
	case shared.ErrorTypeMessaging:
		return http.StatusInternalServerError // 500
	case shared.ErrorTypeExternal:
		return http.StatusBadGateway // 502
	case shared.ErrorTypeNetwork:
		return http.StatusBadGateway // 502
	case shared.ErrorTypeInternal:
		return http.StatusInternalServerError // 500
	case shared.ErrorTypeTimeout:
		return http.StatusGatewayTimeout // 504
	case shared.ErrorTypeRateLimit:
		return http.StatusTooManyRequests // 429
	default:
		return http.StatusInternalServerError // 500
	}
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, err error) {
	apiErr := NewAPIError(err)

	// Add request ID from context if available
	if requestID, exists := c.Get("request_id"); exists {
		if reqID, ok := requestID.(string); ok {
			apiErr.WithRequestID(reqID)
		}
	}

	// Add correlation ID from context if available
	if correlationID, exists := c.Get("correlation_id"); exists {
		if corrID, ok := correlationID.(string); ok {
			if apiErr.Details == nil {
				apiErr.Details = make(map[string]interface{})
			}
			apiErr.Details["correlation_id"] = corrID
		}
	}

	c.JSON(apiErr.Status, ErrorResponse{Error: apiErr})
}

// RespondWithValidationErrors sends multiple validation errors
func RespondWithValidationErrors(c *gin.Context, errors []*shared.DomainError) {
	apiErrors := make([]*APIError, len(errors))
	for i, err := range errors {
		apiErrors[i] = NewAPIError(err)
	}

	// Use the first error as the main error
	response := ErrorResponse{
		Error:  apiErrors[0],
		Errors: apiErrors,
	}

	c.JSON(http.StatusBadRequest, response)
}

// Common error responses

// NotFound creates a not found error response
func NotFound(c *gin.Context, resource, resourceID string) {
	err := shared.NewNotFoundError(resource, resourceID)
	RespondWithError(c, err)
}

// BadRequest creates a bad request error response
func BadRequest(c *gin.Context, message string) {
	err := shared.NewBadRequestError(message)
	RespondWithError(c, err)
}

// ValidationError creates a validation error response
func ValidationError(c *gin.Context, field, message string) {
	err := shared.NewValidationError(message, field)
	RespondWithError(c, err)
}

// Unauthorized creates an unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	err := shared.NewUnauthorizedError(message)
	RespondWithError(c, err)
}

// Forbidden creates a forbidden error response
func Forbidden(c *gin.Context, message string) {
	err := shared.NewForbiddenError(message)
	RespondWithError(c, err)
}

// Conflict creates a conflict error response
func Conflict(c *gin.Context, message string) {
	err := shared.NewConflictError(message)
	RespondWithError(c, err)
}

// InternalError creates an internal server error response
func InternalError(c *gin.Context, message string, err error) {
	domainErr := shared.NewInternalError(message, err)
	RespondWithError(c, domainErr)
}
