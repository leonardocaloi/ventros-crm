package campaign

import (
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

// Campaign events

type CampaignCreatedEvent struct {
	shared.BaseEvent
	CampaignID  uuid.UUID
	TenantID    string
	Name        string
	Description string
	GoalType    GoalType
	GoalValue   int
}

func NewCampaignCreatedEvent(campaignID uuid.UUID, tenantID, name, description string, goalType GoalType, goalValue int) CampaignCreatedEvent {
	return CampaignCreatedEvent{
		BaseEvent:   shared.NewBaseEvent("campaign.created", time.Now()),
		CampaignID:  campaignID,
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		GoalType:    goalType,
		GoalValue:   goalValue,
	}
}

type CampaignActivatedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
}

func NewCampaignActivatedEvent(campaignID uuid.UUID) CampaignActivatedEvent {
	return CampaignActivatedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.activated", time.Now()),
		CampaignID: campaignID,
	}
}

type CampaignScheduledEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
	StartDate  time.Time
}

func NewCampaignScheduledEvent(campaignID uuid.UUID, startDate time.Time) CampaignScheduledEvent {
	return CampaignScheduledEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.scheduled", time.Now()),
		CampaignID: campaignID,
		StartDate:  startDate,
	}
}

type CampaignPausedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
}

func NewCampaignPausedEvent(campaignID uuid.UUID) CampaignPausedEvent {
	return CampaignPausedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.paused", time.Now()),
		CampaignID: campaignID,
	}
}

type CampaignResumedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
}

func NewCampaignResumedEvent(campaignID uuid.UUID) CampaignResumedEvent {
	return CampaignResumedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.resumed", time.Now()),
		CampaignID: campaignID,
	}
}

type CampaignCompletedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
}

func NewCampaignCompletedEvent(campaignID uuid.UUID) CampaignCompletedEvent {
	return CampaignCompletedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.completed", time.Now()),
		CampaignID: campaignID,
	}
}

type CampaignArchivedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
}

func NewCampaignArchivedEvent(campaignID uuid.UUID) CampaignArchivedEvent {
	return CampaignArchivedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.archived", time.Now()),
		CampaignID: campaignID,
	}
}

type CampaignStepAddedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
	StepID     uuid.UUID
	StepType   StepType
	Order      int
}

func NewCampaignStepAddedEvent(campaignID, stepID uuid.UUID, stepType StepType, order int) CampaignStepAddedEvent {
	return CampaignStepAddedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.step_added", time.Now()),
		CampaignID: campaignID,
		StepID:     stepID,
		StepType:   stepType,
		Order:      order,
	}
}

type CampaignStepRemovedEvent struct {
	shared.BaseEvent
	CampaignID uuid.UUID
	StepID     uuid.UUID
}

func NewCampaignStepRemovedEvent(campaignID, stepID uuid.UUID) CampaignStepRemovedEvent {
	return CampaignStepRemovedEvent{
		BaseEvent:  shared.NewBaseEvent("campaign.step_removed", time.Now()),
		CampaignID: campaignID,
		StepID:     stepID,
	}
}

// Enrollment events

type ContactEnrolledEvent struct {
	shared.BaseEvent
	EnrollmentID    uuid.UUID
	CampaignID      uuid.UUID
	ContactID       uuid.UUID
	NextScheduledAt time.Time
}

func NewContactEnrolledEvent(enrollmentID, campaignID, contactID uuid.UUID, nextScheduledAt time.Time) ContactEnrolledEvent {
	return ContactEnrolledEvent{
		BaseEvent:       shared.NewBaseEvent("campaign.contact_enrolled", time.Now()),
		EnrollmentID:    enrollmentID,
		CampaignID:      campaignID,
		ContactID:       contactID,
		NextScheduledAt: nextScheduledAt,
	}
}

type EnrollmentAdvancedEvent struct {
	shared.BaseEvent
	EnrollmentID     uuid.UUID
	CampaignID       uuid.UUID
	ContactID        uuid.UUID
	CurrentStepOrder int
	NextScheduledAt  *time.Time
}

func NewEnrollmentAdvancedEvent(enrollmentID, campaignID, contactID uuid.UUID, currentStepOrder int, nextScheduledAt *time.Time) EnrollmentAdvancedEvent {
	return EnrollmentAdvancedEvent{
		BaseEvent:        shared.NewBaseEvent("campaign.enrollment_advanced", time.Now()),
		EnrollmentID:     enrollmentID,
		CampaignID:       campaignID,
		ContactID:        contactID,
		CurrentStepOrder: currentStepOrder,
		NextScheduledAt:  nextScheduledAt,
	}
}

type EnrollmentPausedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
}

func NewEnrollmentPausedEvent(enrollmentID, campaignID, contactID uuid.UUID) EnrollmentPausedEvent {
	return EnrollmentPausedEvent{
		BaseEvent:    shared.NewBaseEvent("campaign.enrollment_paused", time.Now()),
		EnrollmentID: enrollmentID,
		CampaignID:   campaignID,
		ContactID:    contactID,
	}
}

type EnrollmentResumedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
}

func NewEnrollmentResumedEvent(enrollmentID, campaignID, contactID uuid.UUID) EnrollmentResumedEvent {
	return EnrollmentResumedEvent{
		BaseEvent:    shared.NewBaseEvent("campaign.enrollment_resumed", time.Now()),
		EnrollmentID: enrollmentID,
		CampaignID:   campaignID,
		ContactID:    contactID,
	}
}

type EnrollmentCompletedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
	CompletedAt  time.Time
}

func NewEnrollmentCompletedEvent(enrollmentID, campaignID, contactID uuid.UUID, completedAt time.Time) EnrollmentCompletedEvent {
	return EnrollmentCompletedEvent{
		BaseEvent:    shared.NewBaseEvent("campaign.enrollment_completed", time.Now()),
		EnrollmentID: enrollmentID,
		CampaignID:   campaignID,
		ContactID:    contactID,
		CompletedAt:  completedAt,
	}
}

type EnrollmentExitedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	CampaignID   uuid.UUID
	ContactID    uuid.UUID
	ExitReason   string
	ExitedAt     time.Time
}

func NewEnrollmentExitedEvent(enrollmentID, campaignID, contactID uuid.UUID, exitReason string, exitedAt time.Time) EnrollmentExitedEvent {
	return EnrollmentExitedEvent{
		BaseEvent:    shared.NewBaseEvent("campaign.enrollment_exited", time.Now()),
		EnrollmentID: enrollmentID,
		CampaignID:   campaignID,
		ContactID:    contactID,
		ExitReason:   exitReason,
		ExitedAt:     exitedAt,
	}
}
