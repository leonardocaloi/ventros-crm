package contact

import (
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/shared"
)

// Legacy errors (kept for backward compatibility)
var (
	ErrContactNotFound = errors.New("contact not found")
)

// NewContactNotFoundError creates a not found error for a contact
func NewContactNotFoundError(contactID string) *shared.DomainError {
	err := shared.NewNotFoundError("contact", contactID)
	err.Err = ErrContactNotFound // Wrap the sentinel error for errors.Is() compatibility
	return err
}

// NewContactAlreadyExistsError creates an already exists error
func NewContactAlreadyExistsError(identifier string) *shared.DomainError {
	return shared.NewAlreadyExistsError("contact", identifier)
}

// NewContactValidationError creates a validation error for contact fields
func NewContactValidationError(field, message string) *shared.DomainError {
	return shared.NewValidationError(message, field).WithResource("contact", "")
}

// NewContactInvariantViolation creates an invariant violation error
func NewContactInvariantViolation(message string) *shared.DomainError {
	return shared.NewInvariantViolationError(message).WithResource("contact", "")
}
