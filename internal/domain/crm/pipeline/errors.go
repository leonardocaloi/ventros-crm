package pipeline

import "errors"

// Domain errors for Pipeline
var (
	// Pipeline errors
	ErrPipelineIDRequired   = errors.New("pipeline ID is required")
	ErrTenantIDRequired     = errors.New("tenant ID is required")
	ErrPipelineNotFound     = errors.New("pipeline not found")
	ErrPipelineNameRequired = errors.New("pipeline name is required")

	// Status errors
	ErrStatusNotFound     = errors.New("status not found")
	ErrStatusNameRequired = errors.New("status name is required")
	ErrStatusTypeRequired = errors.New("status type is required")

	// Custom field errors
	ErrCustomFieldRequired        = errors.New("custom field is required")
	ErrCustomFieldIDRequired      = errors.New("custom field ID is required")
	ErrCustomFieldNotFound        = errors.New("custom field not found")
	ErrCustomFieldKeyImmutable    = errors.New("custom field key cannot be changed")
	ErrCustomFieldTypeImmutable   = errors.New("custom field type cannot be changed")
	ErrCustomFieldAlreadyExists   = errors.New("custom field with this key already exists")
	ErrCustomFieldKeyRequired     = errors.New("custom field key is required")
	ErrCustomFieldInvalidType     = errors.New("invalid custom field type")
	ErrCustomFieldInvalidValue    = errors.New("invalid custom field value for type")
	ErrCustomFieldOperationFailed = errors.New("custom field operation failed")
	ErrInvalidTimestamp           = errors.New("invalid timestamp")
)
