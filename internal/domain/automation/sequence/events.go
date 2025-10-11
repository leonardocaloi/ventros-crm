package sequence

import (
	"time"

	"github.com/google/uuid"
)

// Sequence Events

type SequenceCreatedEvent struct {
	SequenceID uuid.UUID
	TenantID   string
	Name       string
	Timestamp  time.Time
}

func (e *SequenceCreatedEvent) EventType() string         { return "sequence.created" }
func (e *SequenceCreatedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *SequenceCreatedEvent) AggregateID() uuid.UUID    { return e.SequenceID }

type SequenceActivatedEvent struct {
	SequenceID uuid.UUID
	Timestamp  time.Time
}

func (e *SequenceActivatedEvent) EventType() string         { return "sequence.activated" }
func (e *SequenceActivatedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *SequenceActivatedEvent) AggregateID() uuid.UUID    { return e.SequenceID }

type SequencePausedEvent struct {
	SequenceID uuid.UUID
	Timestamp  time.Time
}

func (e *SequencePausedEvent) EventType() string         { return "sequence.paused" }
func (e *SequencePausedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *SequencePausedEvent) AggregateID() uuid.UUID    { return e.SequenceID }

type SequenceResumedEvent struct {
	SequenceID uuid.UUID
	Timestamp  time.Time
}

func (e *SequenceResumedEvent) EventType() string         { return "sequence.resumed" }
func (e *SequenceResumedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *SequenceResumedEvent) AggregateID() uuid.UUID    { return e.SequenceID }

type SequenceArchivedEvent struct {
	SequenceID uuid.UUID
	Timestamp  time.Time
}

func (e *SequenceArchivedEvent) EventType() string         { return "sequence.archived" }
func (e *SequenceArchivedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *SequenceArchivedEvent) AggregateID() uuid.UUID    { return e.SequenceID }

type SequenceStepAddedEvent struct {
	SequenceID uuid.UUID
	StepID     uuid.UUID
	Order      int
	Timestamp  time.Time
}

func (e *SequenceStepAddedEvent) EventType() string         { return "sequence.step_added" }
func (e *SequenceStepAddedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *SequenceStepAddedEvent) AggregateID() uuid.UUID    { return e.SequenceID }

// Enrollment Events

type ContactEnrolledEvent struct {
	EnrollmentID uuid.UUID
	SequenceID   uuid.UUID
	ContactID    uuid.UUID
	Timestamp    time.Time
}

func (e *ContactEnrolledEvent) EventType() string         { return "sequence.contact_enrolled" }
func (e *ContactEnrolledEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *ContactEnrolledEvent) AggregateID() uuid.UUID    { return e.EnrollmentID }

type EnrollmentAdvancedEvent struct {
	EnrollmentID uuid.UUID
	NewStepOrder int
	Timestamp    time.Time
}

func (e *EnrollmentAdvancedEvent) EventType() string         { return "sequence.enrollment_advanced" }
func (e *EnrollmentAdvancedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *EnrollmentAdvancedEvent) AggregateID() uuid.UUID    { return e.EnrollmentID }

type EnrollmentCompletedEvent struct {
	EnrollmentID uuid.UUID
	Timestamp    time.Time
}

func (e *EnrollmentCompletedEvent) EventType() string         { return "sequence.enrollment_completed" }
func (e *EnrollmentCompletedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *EnrollmentCompletedEvent) AggregateID() uuid.UUID    { return e.EnrollmentID }

type EnrollmentExitedEvent struct {
	EnrollmentID uuid.UUID
	Reason       string
	Timestamp    time.Time
}

func (e *EnrollmentExitedEvent) EventType() string         { return "sequence.enrollment_exited" }
func (e *EnrollmentExitedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *EnrollmentExitedEvent) AggregateID() uuid.UUID    { return e.EnrollmentID }

type EnrollmentPausedEvent struct {
	EnrollmentID uuid.UUID
	Timestamp    time.Time
}

func (e *EnrollmentPausedEvent) EventType() string         { return "sequence.enrollment_paused" }
func (e *EnrollmentPausedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *EnrollmentPausedEvent) AggregateID() uuid.UUID    { return e.EnrollmentID }

type EnrollmentResumedEvent struct {
	EnrollmentID uuid.UUID
	Timestamp    time.Time
}

func (e *EnrollmentResumedEvent) EventType() string         { return "sequence.enrollment_resumed" }
func (e *EnrollmentResumedEvent) EventTimestamp() time.Time { return e.Timestamp }
func (e *EnrollmentResumedEvent) AggregateID() uuid.UUID    { return e.EnrollmentID }
