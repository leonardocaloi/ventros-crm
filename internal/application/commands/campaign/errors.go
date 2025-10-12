package campaign

import "errors"

var (
	// Command validation errors
	ErrTenantIDRequired     = errors.New("tenant_id is required")
	ErrCampaignIDRequired   = errors.New("campaign_id is required")
	ErrCampaignNameRequired = errors.New("campaign name is required")
	ErrGoalTypeRequired     = errors.New("goal_type is required")
	ErrNoFieldsToUpdate     = errors.New("no fields provided for update")

	// Business logic errors
	ErrCampaignCreationFailed = errors.New("failed to create campaign")
	ErrStepCreationFailed     = errors.New("failed to add step to campaign")
	ErrRepositorySaveFailed   = errors.New("failed to save campaign")
	ErrCampaignNotFound       = errors.New("campaign not found")
	ErrAccessDenied           = errors.New("access denied: campaign belongs to different tenant")
	ErrCampaignUpdateFailed   = errors.New("failed to update campaign")
	ErrCampaignActivateFailed = errors.New("failed to activate campaign")
	ErrCampaignPauseFailed    = errors.New("failed to pause campaign")
	ErrCampaignCompleteFailed = errors.New("failed to complete campaign")
)
