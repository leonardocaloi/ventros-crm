package sequence

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// SequenceEnrollment representa a participação de um contato em uma sequência
type SequenceEnrollment struct {
	id              uuid.UUID
	sequenceID      uuid.UUID
	contactID       uuid.UUID
	status          EnrollmentStatus
	currentStepOrder int        // Ordem do step atual
	nextScheduledAt *time.Time // Quando enviar próxima mensagem

	// Exit tracking
	exitedAt     *time.Time
	exitReason   *string

	// Completion tracking
	completedAt  *time.Time

	enrolledAt   time.Time
	updatedAt    time.Time

	events []DomainEvent
}

type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"    // Ativo na sequência
	EnrollmentStatusPaused    EnrollmentStatus = "paused"    // Pausado
	EnrollmentStatusCompleted EnrollmentStatus = "completed" // Completou todos os steps
	EnrollmentStatusExited    EnrollmentStatus = "exited"    // Saiu antes de completar
)

// NewSequenceEnrollment creates a new enrollment
func NewSequenceEnrollment(
	sequenceID uuid.UUID,
	contactID uuid.UUID,
	firstStepDelay time.Duration,
) (*SequenceEnrollment, error) {
	if sequenceID == uuid.Nil {
		return nil, errors.New("sequenceID cannot be empty")
	}
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be empty")
	}

	now := time.Now()
	nextScheduled := now.Add(firstStepDelay)

	enrollment := &SequenceEnrollment{
		id:               uuid.New(),
		sequenceID:       sequenceID,
		contactID:        contactID,
		status:           EnrollmentStatusActive,
		currentStepOrder: 0, // Começa no step 0
		nextScheduledAt:  &nextScheduled,
		enrolledAt:       now,
		updatedAt:        now,
		events:           []DomainEvent{},
	}

	enrollment.addEvent(&ContactEnrolledEvent{
		EnrollmentID: enrollment.id,
		SequenceID:   sequenceID,
		ContactID:    contactID,
		Timestamp:    now,
	})

	return enrollment, nil
}

// ReconstructEnrollment reconstructs an enrollment from persistence
func ReconstructEnrollment(
	id uuid.UUID,
	sequenceID uuid.UUID,
	contactID uuid.UUID,
	status EnrollmentStatus,
	currentStepOrder int,
	nextScheduledAt *time.Time,
	exitedAt *time.Time,
	exitReason *string,
	completedAt *time.Time,
	enrolledAt, updatedAt time.Time,
) *SequenceEnrollment {
	return &SequenceEnrollment{
		id:               id,
		sequenceID:       sequenceID,
		contactID:        contactID,
		status:           status,
		currentStepOrder: currentStepOrder,
		nextScheduledAt:  nextScheduledAt,
		exitedAt:         exitedAt,
		exitReason:       exitReason,
		completedAt:      completedAt,
		enrolledAt:       enrolledAt,
		updatedAt:        updatedAt,
		events:           []DomainEvent{},
	}
}

// AdvanceToNextStep moves to the next step
func (e *SequenceEnrollment) AdvanceToNextStep(nextStepDelay time.Duration, hasNextStep bool) error {
	if e.status != EnrollmentStatusActive {
		return errors.New("can only advance active enrollments")
	}

	e.currentStepOrder++

	if hasNextStep {
		// Tem próximo step, agendar
		nextScheduled := time.Now().Add(nextStepDelay)
		e.nextScheduledAt = &nextScheduled
	} else {
		// Não tem mais steps, marcar como completo
		e.nextScheduledAt = nil
		return e.Complete()
	}

	e.updatedAt = time.Now()

	e.addEvent(&EnrollmentAdvancedEvent{
		EnrollmentID: e.id,
		NewStepOrder: e.currentStepOrder,
		Timestamp:    time.Now(),
	})

	return nil
}

// Pause pauses the enrollment
func (e *SequenceEnrollment) Pause() error {
	if e.status != EnrollmentStatusActive {
		return errors.New("can only pause active enrollments")
	}

	e.status = EnrollmentStatusPaused
	e.updatedAt = time.Now()

	e.addEvent(&EnrollmentPausedEvent{
		EnrollmentID: e.id,
		Timestamp:    time.Now(),
	})

	return nil
}

// Resume resumes a paused enrollment
func (e *SequenceEnrollment) Resume() error {
	if e.status != EnrollmentStatusPaused {
		return errors.New("can only resume paused enrollments")
	}

	e.status = EnrollmentStatusActive
	e.updatedAt = time.Now()

	e.addEvent(&EnrollmentResumedEvent{
		EnrollmentID: e.id,
		Timestamp:    time.Now(),
	})

	return nil
}

// Complete marks the enrollment as completed
func (e *SequenceEnrollment) Complete() error {
	if e.status == EnrollmentStatusCompleted {
		return errors.New("enrollment already completed")
	}

	now := time.Now()
	e.status = EnrollmentStatusCompleted
	e.completedAt = &now
	e.nextScheduledAt = nil
	e.updatedAt = now

	e.addEvent(&EnrollmentCompletedEvent{
		EnrollmentID: e.id,
		Timestamp:    now,
	})

	return nil
}

// Exit exits the enrollment before completion
func (e *SequenceEnrollment) Exit(reason string) error {
	if e.status == EnrollmentStatusCompleted || e.status == EnrollmentStatusExited {
		return errors.New("enrollment already finished")
	}

	now := time.Now()
	e.status = EnrollmentStatusExited
	e.exitedAt = &now
	e.exitReason = &reason
	e.nextScheduledAt = nil
	e.updatedAt = now

	e.addEvent(&EnrollmentExitedEvent{
		EnrollmentID: e.id,
		Reason:       reason,
		Timestamp:    now,
	})

	return nil
}

// IsReadyForNextStep checks if it's time to send the next step
func (e *SequenceEnrollment) IsReadyForNextStep() bool {
	if e.status != EnrollmentStatusActive {
		return false
	}
	if e.nextScheduledAt == nil {
		return false
	}
	return e.nextScheduledAt.Before(time.Now()) || e.nextScheduledAt.Equal(time.Now())
}

// Getters
func (e *SequenceEnrollment) ID() uuid.UUID                  { return e.id }
func (e *SequenceEnrollment) SequenceID() uuid.UUID          { return e.sequenceID }
func (e *SequenceEnrollment) ContactID() uuid.UUID           { return e.contactID }
func (e *SequenceEnrollment) Status() EnrollmentStatus       { return e.status }
func (e *SequenceEnrollment) CurrentStepOrder() int          { return e.currentStepOrder }
func (e *SequenceEnrollment) NextScheduledAt() *time.Time    { return e.nextScheduledAt }
func (e *SequenceEnrollment) ExitedAt() *time.Time           { return e.exitedAt }
func (e *SequenceEnrollment) ExitReason() *string            { return e.exitReason }
func (e *SequenceEnrollment) CompletedAt() *time.Time        { return e.completedAt }
func (e *SequenceEnrollment) EnrolledAt() time.Time          { return e.enrolledAt }
func (e *SequenceEnrollment) UpdatedAt() time.Time           { return e.updatedAt }

func (e *SequenceEnrollment) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, e.events...)
}

func (e *SequenceEnrollment) ClearEvents() {
	e.events = []DomainEvent{}
}

func (e *SequenceEnrollment) addEvent(event DomainEvent) {
	e.events = append(e.events, event)
}

// EnrollmentRepository interface
type EnrollmentRepository interface {
	Save(enrollment *SequenceEnrollment) error
	FindByID(id uuid.UUID) (*SequenceEnrollment, error)
	FindBySequenceID(sequenceID uuid.UUID) ([]*SequenceEnrollment, error)
	FindByContactID(contactID uuid.UUID) ([]*SequenceEnrollment, error)
	FindReadyForNextStep() ([]*SequenceEnrollment, error)
	FindActiveBySequenceAndContact(sequenceID, contactID uuid.UUID) (*SequenceEnrollment, error)
	Delete(id uuid.UUID) error
}
