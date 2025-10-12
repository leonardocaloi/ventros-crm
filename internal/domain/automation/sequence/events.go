package sequence

import (
	"time"

	"github.com/google/uuid"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
)

// Sequence Events

type SequenceCreatedEvent struct {
	shared.BaseEvent
	SequenceID uuid.UUID
	TenantID   string
	Name       string
}

func NewSequenceCreatedEvent(sequenceID uuid.UUID, tenantID, name string) SequenceCreatedEvent {
	return SequenceCreatedEvent{
		BaseEvent:  shared.NewBaseEvent("sequence.created", time.Now()),
		SequenceID: sequenceID,
		TenantID:   tenantID,
		Name:       name,
	}
}

type SequenceActivatedEvent struct {
	shared.BaseEvent
	SequenceID uuid.UUID
}

func NewSequenceActivatedEvent(sequenceID uuid.UUID) SequenceActivatedEvent {
	return SequenceActivatedEvent{
		BaseEvent:  shared.NewBaseEvent("sequence.activated", time.Now()),
		SequenceID: sequenceID,
	}
}

type SequencePausedEvent struct {
	shared.BaseEvent
	SequenceID uuid.UUID
}

func NewSequencePausedEvent(sequenceID uuid.UUID) SequencePausedEvent {
	return SequencePausedEvent{
		BaseEvent:  shared.NewBaseEvent("sequence.paused", time.Now()),
		SequenceID: sequenceID,
	}
}

type SequenceResumedEvent struct {
	shared.BaseEvent
	SequenceID uuid.UUID
}

func NewSequenceResumedEvent(sequenceID uuid.UUID) SequenceResumedEvent {
	return SequenceResumedEvent{
		BaseEvent:  shared.NewBaseEvent("sequence.resumed", time.Now()),
		SequenceID: sequenceID,
	}
}

type SequenceArchivedEvent struct {
	shared.BaseEvent
	SequenceID uuid.UUID
}

func NewSequenceArchivedEvent(sequenceID uuid.UUID) SequenceArchivedEvent {
	return SequenceArchivedEvent{
		BaseEvent:  shared.NewBaseEvent("sequence.archived", time.Now()),
		SequenceID: sequenceID,
	}
}

type SequenceStepAddedEvent struct {
	shared.BaseEvent
	SequenceID uuid.UUID
	StepID     uuid.UUID
	Order      int
}

func NewSequenceStepAddedEvent(sequenceID, stepID uuid.UUID, order int) SequenceStepAddedEvent {
	return SequenceStepAddedEvent{
		BaseEvent:  shared.NewBaseEvent("sequence.step_added", time.Now()),
		SequenceID: sequenceID,
		StepID:     stepID,
		Order:      order,
	}
}

// Enrollment Events

type ContactEnrolledEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	SequenceID   uuid.UUID
	ContactID    uuid.UUID
}

func NewContactEnrolledEvent(enrollmentID, sequenceID, contactID uuid.UUID) ContactEnrolledEvent {
	return ContactEnrolledEvent{
		BaseEvent:    shared.NewBaseEvent("sequence.contact_enrolled", time.Now()),
		EnrollmentID: enrollmentID,
		SequenceID:   sequenceID,
		ContactID:    contactID,
	}
}

type EnrollmentAdvancedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	NewStepOrder int
}

func NewEnrollmentAdvancedEvent(enrollmentID uuid.UUID, newStepOrder int) EnrollmentAdvancedEvent {
	return EnrollmentAdvancedEvent{
		BaseEvent:    shared.NewBaseEvent("sequence.enrollment_advanced", time.Now()),
		EnrollmentID: enrollmentID,
		NewStepOrder: newStepOrder,
	}
}

type EnrollmentCompletedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
}

func NewEnrollmentCompletedEvent(enrollmentID uuid.UUID) EnrollmentCompletedEvent {
	return EnrollmentCompletedEvent{
		BaseEvent:    shared.NewBaseEvent("sequence.enrollment_completed", time.Now()),
		EnrollmentID: enrollmentID,
	}
}

type EnrollmentExitedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
	Reason       string
}

func NewEnrollmentExitedEvent(enrollmentID uuid.UUID, reason string) EnrollmentExitedEvent {
	return EnrollmentExitedEvent{
		BaseEvent:    shared.NewBaseEvent("sequence.enrollment_exited", time.Now()),
		EnrollmentID: enrollmentID,
		Reason:       reason,
	}
}

type EnrollmentPausedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
}

func NewEnrollmentPausedEvent(enrollmentID uuid.UUID) EnrollmentPausedEvent {
	return EnrollmentPausedEvent{
		BaseEvent:    shared.NewBaseEvent("sequence.enrollment_paused", time.Now()),
		EnrollmentID: enrollmentID,
	}
}

type EnrollmentResumedEvent struct {
	shared.BaseEvent
	EnrollmentID uuid.UUID
}

func NewEnrollmentResumedEvent(enrollmentID uuid.UUID) EnrollmentResumedEvent {
	return EnrollmentResumedEvent{
		BaseEvent:    shared.NewBaseEvent("sequence.enrollment_resumed", time.Now()),
		EnrollmentID: enrollmentID,
	}
}
