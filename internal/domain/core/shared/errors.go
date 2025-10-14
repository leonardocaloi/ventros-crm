package shared

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// Domain/Business Logic Errors
	ErrorTypeValidation         ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound           ErrorType = "NOT_FOUND"
	ErrorTypeAlreadyExists      ErrorType = "ALREADY_EXISTS"
	ErrorTypeConflict           ErrorType = "CONFLICT"
	ErrorTypeOptimisticLock     ErrorType = "OPTIMISTIC_LOCK_CONFLICT"
	ErrorTypeForbidden          ErrorType = "FORBIDDEN"
	ErrorTypeUnauthorized       ErrorType = "UNAUTHORIZED"
	ErrorTypeBadRequest         ErrorType = "BAD_REQUEST"
	ErrorTypePrecondition       ErrorType = "PRECONDITION_FAILED"
	ErrorTypeInvariantViolation ErrorType = "INVARIANT_VIOLATION"

	// Infrastructure Errors
	ErrorTypeDatabase  ErrorType = "DATABASE_ERROR"
	ErrorTypeCache     ErrorType = "CACHE_ERROR"
	ErrorTypeMessaging ErrorType = "MESSAGING_ERROR"
	ErrorTypeExternal  ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeNetwork   ErrorType = "NETWORK_ERROR"

	// Application Errors
	ErrorTypeInternal  ErrorType = "INTERNAL_ERROR"
	ErrorTypeTimeout   ErrorType = "TIMEOUT"
	ErrorTypeRateLimit ErrorType = "RATE_LIMIT_EXCEEDED"
	ErrorTypeUnknown   ErrorType = "UNKNOWN_ERROR"
)

// DomainError represents a domain-level error with rich context
type DomainError struct {
	Type       ErrorType
	Message    string
	Code       string
	Details    map[string]interface{}
	Err        error  // underlying error
	Field      string // for validation errors
	Resource   string // resource identifier (e.g., "contact", "session")
	ResourceID string // specific resource ID
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is allows error comparison
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Type == t.Type && e.Code == t.Code
}

// WithDetail adds a detail to the error
func (e *DomainError) WithDetail(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithField sets the field for validation errors
func (e *DomainError) WithField(field string) *DomainError {
	e.Field = field
	return e
}

// WithResource sets the resource information
func (e *DomainError) WithResource(resource, resourceID string) *DomainError {
	e.Resource = resource
	e.ResourceID = resourceID
	return e
}

// Error constructors

// NewValidationError creates a validation error
func NewValidationError(message, field string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Message: message,
		Code:    "VALIDATION_FAILED",
		Field:   field,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, resourceID string) *DomainError {
	return &DomainError{
		Type:       ErrorTypeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		Code:       "RESOURCE_NOT_FOUND",
		Resource:   resource,
		ResourceID: resourceID,
	}
}

// NewAlreadyExistsError creates an already exists error
func NewAlreadyExistsError(resource, identifier string) *DomainError {
	err := &DomainError{
		Type:     ErrorTypeAlreadyExists,
		Message:  fmt.Sprintf("%s already exists", resource),
		Code:     "RESOURCE_ALREADY_EXISTS",
		Resource: resource,
	}
	return err.WithDetail("identifier", identifier)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeConflict,
		Message: message,
		Code:    "CONFLICT",
	}
}

// NewOptimisticLockError creates an optimistic locking conflict error
func NewOptimisticLockError(resource, resourceID string, expectedVersion, actualVersion int) *DomainError {
	err := &DomainError{
		Type:       ErrorTypeOptimisticLock,
		Message:    fmt.Sprintf("%s was modified by another transaction (version mismatch)", resource),
		Code:       "OPTIMISTIC_LOCK_CONFLICT",
		Resource:   resource,
		ResourceID: resourceID,
	}
	return err.
		WithDetail("expected_version", expectedVersion).
		WithDetail("actual_version", actualVersion)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Code:    "FORBIDDEN",
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Code:    "UNAUTHORIZED",
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeBadRequest,
		Message: message,
		Code:    "BAD_REQUEST",
	}
}

// NewPreconditionError creates a precondition failed error
func NewPreconditionError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypePrecondition,
		Message: message,
		Code:    "PRECONDITION_FAILED",
	}
}

// NewInvariantViolationError creates an invariant violation error
func NewInvariantViolationError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInvariantViolation,
		Message: message,
		Code:    "INVARIANT_VIOLATION",
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeDatabase,
		Message: message,
		Code:    "DATABASE_ERROR",
		Err:     err,
	}
}

// NewCacheError creates a cache error
func NewCacheError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeCache,
		Message: message,
		Code:    "CACHE_ERROR",
		Err:     err,
	}
}

// NewExternalServiceError creates an external service error
func NewExternalServiceError(service, message string, err error) *DomainError {
	domainErr := &DomainError{
		Type:    ErrorTypeExternal,
		Message: message,
		Code:    "EXTERNAL_SERVICE_ERROR",
		Err:     err,
	}
	return domainErr.WithDetail("service", service)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *DomainError {
	err := &DomainError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("operation timed out: %s", operation),
		Code:    "TIMEOUT",
	}
	return err.WithDetail("operation", operation)
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeRateLimit,
		Message: message,
		Code:    "RATE_LIMIT_EXCEEDED",
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string, err error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternal,
		Message: message,
		Code:    "INTERNAL_ERROR",
		Err:     err,
	}
}

// WrapError wraps an existing error with context
func WrapError(err error, message string) *DomainError {
	if err == nil {
		return nil
	}

	// If already a DomainError, preserve type but add context
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return &DomainError{
			Type:       domainErr.Type,
			Message:    message + ": " + domainErr.Message,
			Code:       domainErr.Code,
			Details:    domainErr.Details,
			Err:        domainErr.Err,
			Field:      domainErr.Field,
			Resource:   domainErr.Resource,
			ResourceID: domainErr.ResourceID,
		}
	}

	// Wrap as internal error
	return NewInternalError(message, err)
}

// Helper functions to check error types

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == ErrorTypeValidation
	}
	return false
}

// IsAlreadyExistsError checks if error is an already exists error
func IsAlreadyExistsError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == ErrorTypeAlreadyExists
	}
	return false
}

// IsForbiddenError checks if error is a forbidden error
func IsForbiddenError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == ErrorTypeForbidden
	}
	return false
}

// IsUnauthorizedError checks if error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == ErrorTypeUnauthorized
	}
	return false
}

// IsOptimisticLockError checks if error is an optimistic locking conflict error
func IsOptimisticLockError(err error) bool {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Type == ErrorTypeOptimisticLock
	}
	return false
}

// IsDomainError checks if an error is a DomainError and optionally unwraps it
func IsDomainError(err error, target **DomainError) bool {
	return errors.As(err, target)
}

// Custom Field Errors
var (
	ErrCustomFieldKeyRequired     = errors.New("custom field key cannot be empty")
	ErrInvalidCustomFieldType     = errors.New("invalid custom field type")
	ErrCustomFieldReadOnly        = errors.New("custom field is read-only")
	ErrCustomFieldRequired        = errors.New("custom field is required")
	ErrInvalidCustomFieldValue    = errors.New("invalid custom field value")
	ErrCustomFieldNil             = errors.New("custom field cannot be nil")
	ErrCustomFieldNotFound        = errors.New("custom field not found")
	ErrCustomFieldCannotBeRemoved = errors.New("custom field cannot be removed")
)
