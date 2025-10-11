package campaign

import (
	"time"

	"github.com/google/uuid"
)

// Campaign events

type CampaignCreatedEvent struct {
	CampaignID  uuid.UUID
	TenantID    string
	Name        string
	Description string
	GoalType    GoalType
	GoalValue   int
	OccurredAt  time.Time
}

type CampaignActivatedEvent struct {
	CampaignID uuid.UUID
	OccurredAt time.Time
}

type CampaignScheduledEvent struct {
	CampaignID uuid.UUID
	StartDate  time.Time
	OccurredAt time.Time
}

type CampaignPausedEvent struct {
	CampaignID uuid.UUID
	OccurredAt time.Time
}

type CampaignResumedEvent struct {
	CampaignID uuid.UUID
	OccurredAt time.Time
}

type CampaignCompletedEvent struct {
	CampaignID uuid.UUID
	OccurredAt time.Time
}

type CampaignArchivedEvent struct {
	CampaignID uuid.UUID
	OccurredAt time.Time
}

type CampaignStepAddedEvent struct {
	CampaignID uuid.UUID
	StepID     uuid.UUID
	StepType   StepType
	Order      int
	OccurredAt time.Time
}

type CampaignStepRemovedEvent struct {
	CampaignID uuid.UUID
	StepID     uuid.UUID
	OccurredAt time.Time
}

// Enrollment events

type ContactEnrolledEvent struct {
	EnrollmentID    uuid.UUID
	CampaignID      uuid.UUID
	ContactID       uuid.UUID
	NextScheduledAt time.Time
	OccurredAt      time.Time
}

type EnrollmentAdvancedEvent struct {
	EnrollmentID     uuid.UUID
	CampaignID       uuid.UUID
	ContactID        uuid.UUID
	CurrentStepOrder int
	NextScheduledAt  *time.Time
	OccurredAt       time.Time
}

type EnrollmentPausedEvent struct {
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
	OccurredAt   time.Time
}

type EnrollmentResumedEvent struct {
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
	OccurredAt   time.Time
}

type EnrollmentCompletedEvent struct {
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
	CompletedAt  time.Time
	OccurredAt   time.Time
}

type EnrollmentExitedEvent struct {
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
	ExitReason   string
	ExitedAt     time.Time
	OccurredAt   time.Time
}
