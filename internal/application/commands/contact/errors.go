package contact

import "errors"

var (
	// Command validation errors
	ErrTenantIDRequired    = errors.New("tenant_id is required")
	ErrProjectIDRequired   = errors.New("project_id is required")
	ErrContactIDRequired   = errors.New("contact_id is required")
	ErrContactNameRequired = errors.New("contact name is required")
	ErrNoFieldsToUpdate    = errors.New("no fields provided for update")
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrInvalidPhone        = errors.New("invalid phone format")

	// Business logic errors
	ErrContactCreationFailed = errors.New("failed to create contact")
	ErrRepositorySaveFailed  = errors.New("failed to save contact")
	ErrContactNotFound       = errors.New("contact not found")
	ErrAccessDenied          = errors.New("access denied: contact belongs to different tenant")
	ErrContactUpdateFailed   = errors.New("failed to update contact")
	ErrContactDeleteFailed   = errors.New("failed to delete contact")
)
