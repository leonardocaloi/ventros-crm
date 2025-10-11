package campaign

import "errors"

var (
	// ErrCampaignNotFound is returned when a campaign is not found
	ErrCampaignNotFound = errors.New("campaign not found")

	// ErrEnrollmentNotFound is returned when an enrollment is not found
	ErrEnrollmentNotFound = errors.New("enrollment not found")

	// ErrInvalidStatus is returned when a status transition is invalid
	ErrInvalidStatus = errors.New("invalid status transition")

	// ErrInvalidStepConfig is returned when a step configuration is invalid
	ErrInvalidStepConfig = errors.New("invalid step configuration")

	// ErrDuplicateEnrollment is returned when trying to enroll a contact that's already enrolled
	ErrDuplicateEnrollment = errors.New("contact is already enrolled in this campaign")

	// ErrCampaignNotActive is returned when trying to enroll in a non-active campaign
	ErrCampaignNotActive = errors.New("campaign is not active")

	// ErrNoSteps is returned when trying to activate a campaign with no steps
	ErrNoSteps = errors.New("campaign has no steps")
)
