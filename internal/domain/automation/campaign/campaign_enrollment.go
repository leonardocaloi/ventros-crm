package campaign

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// CampaignEnrollment tracks a contact's participation in a campaign
type CampaignEnrollment struct {
	id               uuid.UUID
	campaignID       uuid.UUID
	contactID        uuid.UUID
	status           EnrollmentStatus
	currentStepOrder int
	nextScheduledAt  *time.Time
	exitedAt         *time.Time
	exitReason       *string
	completedAt      *time.Time
	enrolledAt       time.Time
	updatedAt        time.Time
	events           []interface{}
}

type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusPaused    EnrollmentStatus = "paused"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
	EnrollmentStatusExited    EnrollmentStatus = "exited"
)

// NewCampaignEnrollment creates a new campaign enrollment
func NewCampaignEnrollment(campaignID, contactID uuid.UUID, firstStepDelay time.Duration) (*CampaignEnrollment, error) {
	if campaignID == uuid.Nil {
		return nil, errors.New("campaign_id is required")
	}
	if contactID == uuid.Nil {
		return nil, errors.New("contact_id is required")
	}

	now := time.Now()
	nextScheduled := now.Add(firstStepDelay)

	enrollment := &CampaignEnrollment{
		id:               uuid.New(),
		campaignID:       campaignID,
		contactID:        contactID,
		status:           EnrollmentStatusActive,
		currentStepOrder: 0,
		nextScheduledAt:  &nextScheduled,
		enrolledAt:       now,
		updatedAt:        now,
		events:           []interface{}{},
	}

	// Emit domain event
	enrollment.addEvent(NewContactEnrolledEvent(
		enrollment.id,
		campaignID,
		contactID,
		nextScheduled,
	))

	return enrollment, nil
}

// ReconstructEnrollment reconstructs an enrollment from persistence
func ReconstructEnrollment(
	id uuid.UUID,
	campaignID uuid.UUID,
	contactID uuid.UUID,
	status EnrollmentStatus,
	currentStepOrder int,
	nextScheduledAt *time.Time,
	exitedAt *time.Time,
	exitReason *string,
	completedAt *time.Time,
	enrolledAt time.Time,
	updatedAt time.Time,
) *CampaignEnrollment {
	return &CampaignEnrollment{
		id:               id,
		campaignID:       campaignID,
		contactID:        contactID,
		status:           status,
		currentStepOrder: currentStepOrder,
		nextScheduledAt:  nextScheduledAt,
		exitedAt:         exitedAt,
		exitReason:       exitReason,
		completedAt:      completedAt,
		enrolledAt:       enrolledAt,
		updatedAt:        updatedAt,
		events:           []interface{}{},
	}
}

// AdvanceToNextStep advances the enrollment to the next step
func (e *CampaignEnrollment) AdvanceToNextStep(nextStepDelay time.Duration, hasNextStep bool) error {
	if e.status != EnrollmentStatusActive {
		return errors.New("can only advance active enrollments")
	}

	e.currentStepOrder++
	e.updatedAt = time.Now()

	if hasNextStep {
		nextScheduled := time.Now().Add(nextStepDelay)
		e.nextScheduledAt = &nextScheduled
	} else {
		// No more steps, mark as completed
		return e.Complete()
	}

	e.addEvent(NewEnrollmentAdvancedEvent(
		e.id,
		e.campaignID,
		e.contactID,
		e.currentStepOrder,
		e.nextScheduledAt,
	))

	return nil
}

// IsReadyForNextStep checks if the enrollment is ready for the next step
func (e *CampaignEnrollment) IsReadyForNextStep() bool {
	if e.status != EnrollmentStatusActive {
		return false
	}
	if e.nextScheduledAt == nil {
		return false
	}
	return time.Now().After(*e.nextScheduledAt) || time.Now().Equal(*e.nextScheduledAt)
}

// Pause pauses the enrollment
func (e *CampaignEnrollment) Pause() error {
	if e.status != EnrollmentStatusActive {
		return errors.New("can only pause active enrollments")
	}

	e.status = EnrollmentStatusPaused
	e.updatedAt = time.Now()

	e.addEvent(NewEnrollmentPausedEvent(e.id, e.campaignID, e.contactID))

	return nil
}

// Resume resumes a paused enrollment
func (e *CampaignEnrollment) Resume() error {
	if e.status != EnrollmentStatusPaused {
		return errors.New("can only resume paused enrollments")
	}

	e.status = EnrollmentStatusActive
	e.updatedAt = time.Now()

	e.addEvent(NewEnrollmentResumedEvent(e.id, e.campaignID, e.contactID))

	return nil
}

// Complete marks the enrollment as completed
func (e *CampaignEnrollment) Complete() error {
	if e.status == EnrollmentStatusCompleted {
		return errors.New("enrollment is already completed")
	}

	e.status = EnrollmentStatusCompleted
	now := time.Now()
	e.completedAt = &now
	e.nextScheduledAt = nil
	e.updatedAt = now

	e.addEvent(NewEnrollmentCompletedEvent(e.id, e.campaignID, e.contactID, now))

	return nil
}

// Exit exits the enrollment with a reason
func (e *CampaignEnrollment) Exit(reason string) error {
	if e.status == EnrollmentStatusExited || e.status == EnrollmentStatusCompleted {
		return errors.New("enrollment is already exited or completed")
	}

	e.status = EnrollmentStatusExited
	now := time.Now()
	e.exitedAt = &now
	e.exitReason = &reason
	e.nextScheduledAt = nil
	e.updatedAt = now

	e.addEvent(NewEnrollmentExitedEvent(e.id, e.campaignID, e.contactID, reason, now))

	return nil
}

// Getters

func (e *CampaignEnrollment) ID() uuid.UUID               { return e.id }
func (e *CampaignEnrollment) CampaignID() uuid.UUID       { return e.campaignID }
func (e *CampaignEnrollment) ContactID() uuid.UUID        { return e.contactID }
func (e *CampaignEnrollment) Status() EnrollmentStatus    { return e.status }
func (e *CampaignEnrollment) CurrentStepOrder() int       { return e.currentStepOrder }
func (e *CampaignEnrollment) NextScheduledAt() *time.Time { return e.nextScheduledAt }
func (e *CampaignEnrollment) ExitedAt() *time.Time        { return e.exitedAt }
func (e *CampaignEnrollment) ExitReason() *string         { return e.exitReason }
func (e *CampaignEnrollment) CompletedAt() *time.Time     { return e.completedAt }
func (e *CampaignEnrollment) EnrolledAt() time.Time       { return e.enrolledAt }
func (e *CampaignEnrollment) UpdatedAt() time.Time        { return e.updatedAt }
func (e *CampaignEnrollment) DomainEvents() []interface{} { return e.events }

func (e *CampaignEnrollment) ClearEvents() {
	e.events = []interface{}{}
}

func (e *CampaignEnrollment) addEvent(event interface{}) {
	e.events = append(e.events, event)
}
