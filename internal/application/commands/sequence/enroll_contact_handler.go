package sequence

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/sequence"
)

// EnrollContactHandler handler para o comando EnrollContact
type EnrollContactHandler struct {
	sequenceRepository   sequence.Repository
	enrollmentRepository sequence.EnrollmentRepository
	logger               *logrus.Logger
}

// NewEnrollContactHandler cria uma nova instância do handler
func NewEnrollContactHandler(
	sequenceRepository sequence.Repository,
	enrollmentRepository sequence.EnrollmentRepository,
	logger *logrus.Logger,
) *EnrollContactHandler {
	return &EnrollContactHandler{
		sequenceRepository:   sequenceRepository,
		enrollmentRepository: enrollmentRepository,
		logger:               logger,
	}
}

// Handle executa o comando de inscrição de contato em sequência
func (h *EnrollContactHandler) Handle(ctx context.Context, cmd EnrollContactCommand) (*sequence.SequenceEnrollment, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid EnrollContact command")
		return nil, err
	}

	// Find sequence
	seq, err := h.sequenceRepository.FindByID(cmd.SequenceID)
	if err != nil {
		h.logger.WithError(err).WithField("sequence_id", cmd.SequenceID).Error("Sequence not found")
		return nil, fmt.Errorf("%w: %v", ErrSequenceNotFound, err)
	}

	if seq == nil {
		h.logger.WithField("sequence_id", cmd.SequenceID).Warn("Sequence not found (nil returned)")
		return nil, ErrSequenceNotFound
	}

	// Validate tenant ownership
	if seq.TenantID() != cmd.TenantID {
		h.logger.WithFields(logrus.Fields{
			"sequence_id":      cmd.SequenceID,
			"sequence_tenant":  seq.TenantID(),
			"requested_tenant": cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return nil, ErrAccessDenied
	}

	// Check if sequence is active
	if seq.Status() != sequence.SequenceStatusActive {
		h.logger.WithFields(logrus.Fields{
			"sequence_id": cmd.SequenceID,
			"status":      seq.Status(),
		}).Warn("Sequence must be active to enroll contacts")
		return nil, fmt.Errorf("%w: sequence must be active to enroll contacts", ErrInvalidSequenceStatus)
	}

	// Check if already enrolled
	existing, err := h.enrollmentRepository.FindActiveBySequenceAndContact(cmd.SequenceID, cmd.ContactID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check existing enrollment")
		return nil, fmt.Errorf("%w: failed to check enrollment: %v", ErrEnrollmentFailed, err)
	}
	if existing != nil {
		h.logger.WithFields(logrus.Fields{
			"sequence_id": cmd.SequenceID,
			"contact_id":  cmd.ContactID,
		}).Warn("Contact is already enrolled in sequence")
		return nil, ErrContactAlreadyEnrolled
	}

	// Get first step delay
	firstStep, err := seq.GetStepByOrder(0)
	if err != nil || firstStep == nil {
		h.logger.WithField("sequence_id", cmd.SequenceID).Warn("Sequence has no steps")
		return nil, ErrSequenceHasNoSteps
	}

	// Create enrollment
	enrollment, err := sequence.NewSequenceEnrollment(
		cmd.SequenceID,
		cmd.ContactID,
		firstStep.GetDelayDuration(),
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create enrollment")
		return nil, fmt.Errorf("%w: %v", ErrEnrollmentFailed, err)
	}

	// Save enrollment
	if err := h.enrollmentRepository.Save(enrollment); err != nil {
		h.logger.WithError(err).Error("Failed to save enrollment")
		return nil, fmt.Errorf("%w: %v", ErrEnrollmentFailed, err)
	}

	// Update sequence stats
	seq.IncrementEnrolled()
	if err := h.sequenceRepository.Save(seq); err != nil {
		h.logger.WithError(err).Warn("Failed to update sequence stats (enrollment was saved)")
		// Don't return error here - enrollment was successful
	}

	h.logger.WithFields(logrus.Fields{
		"enrollment_id": enrollment.ID(),
		"sequence_id":   cmd.SequenceID,
		"contact_id":    cmd.ContactID,
		"tenant_id":     cmd.TenantID,
	}).Info("Contact enrolled successfully")

	return enrollment, nil
}
