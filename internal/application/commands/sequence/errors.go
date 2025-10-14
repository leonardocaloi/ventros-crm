package sequence

import "errors"

var (
	// Command validation errors
	ErrTenantIDRequired    = errors.New("tenant_id is required")
	ErrSequenceIDRequired  = errors.New("sequence_id is required")
	ErrNameRequired        = errors.New("name is required")
	ErrTriggerTypeRequired = errors.New("trigger_type is required")
	ErrContactIDRequired   = errors.New("contact_id is required")
	ErrNoFieldsToUpdate    = errors.New("no fields provided for update")

	// Business logic errors
	ErrSequenceNotFound         = errors.New("sequence not found")
	ErrAccessDenied             = errors.New("access denied: sequence belongs to different tenant")
	ErrSequenceCreationFailed   = errors.New("failed to create sequence")
	ErrSequenceUpdateFailed     = errors.New("failed to update sequence")
	ErrSequenceActivationFailed = errors.New("failed to activate sequence")
	ErrSequencePauseFailed      = errors.New("failed to pause sequence")
	ErrSequenceResumeFailed     = errors.New("failed to resume sequence")
	ErrSequenceArchiveFailed    = errors.New("failed to archive sequence")
	ErrSequenceDeleteFailed     = errors.New("failed to delete sequence")
	ErrRepositorySaveFailed     = errors.New("failed to save sequence")
	ErrInvalidSequenceStatus    = errors.New("invalid sequence status for this operation")
	ErrContactAlreadyEnrolled   = errors.New("contact is already enrolled in this sequence")
	ErrEnrollmentFailed         = errors.New("failed to enroll contact")
	ErrSequenceHasNoSteps       = errors.New("sequence has no steps")
)
