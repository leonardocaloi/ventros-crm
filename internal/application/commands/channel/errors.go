package channel

import "errors"

var (
	// Validation errors
	ErrChannelIDRequired = errors.New("channel ID is required")
	ErrTenantIDRequired  = errors.New("tenant ID is required")

	// Business errors
	ErrChannelNotFound             = errors.New("channel not found")
	ErrChannelAlreadyActive        = errors.New("channel is already active")
	ErrChannelAlreadyActivating    = errors.New("channel is already being activated")
	ErrInvalidChannelForActivation = errors.New("channel cannot be activated in current state")
	ErrRepositorySaveFailed        = errors.New("failed to save channel to repository")
	ErrEventPublishFailed          = errors.New("failed to publish domain event")
)
