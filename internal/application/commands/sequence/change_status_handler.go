package sequence

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/sequence"
)

// ChangeSequenceStatusHandler handler para o comando ChangeSequenceStatus
type ChangeSequenceStatusHandler struct {
	repository sequence.Repository
	logger     *logrus.Logger
}

// NewChangeSequenceStatusHandler cria uma nova instância do handler
func NewChangeSequenceStatusHandler(repository sequence.Repository, logger *logrus.Logger) *ChangeSequenceStatusHandler {
	return &ChangeSequenceStatusHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de mudança de status de sequência
func (h *ChangeSequenceStatusHandler) Handle(ctx context.Context, cmd ChangeSequenceStatusCommand) (*sequence.Sequence, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid ChangeSequenceStatus command")
		return nil, err
	}

	// Find sequence
	seq, err := h.repository.FindByID(cmd.SequenceID)
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

	// Execute action based on command
	switch cmd.Action {
	case StatusActionActivate:
		if err := seq.Activate(); err != nil {
			h.logger.WithError(err).Error("Failed to activate sequence")
			return nil, fmt.Errorf("%w: %v", ErrSequenceActivationFailed, err)
		}
	case StatusActionPause:
		if err := seq.Pause(); err != nil {
			h.logger.WithError(err).Error("Failed to pause sequence")
			return nil, fmt.Errorf("%w: %v", ErrSequencePauseFailed, err)
		}
	case StatusActionResume:
		if err := seq.Resume(); err != nil {
			h.logger.WithError(err).Error("Failed to resume sequence")
			return nil, fmt.Errorf("%w: %v", ErrSequenceResumeFailed, err)
		}
	case StatusActionArchive:
		if err := seq.Archive(); err != nil {
			h.logger.WithError(err).Error("Failed to archive sequence")
			return nil, fmt.Errorf("%w: %v", ErrSequenceArchiveFailed, err)
		}
	default:
		return nil, ErrInvalidSequenceStatus
	}

	// Save to repository
	if err := h.repository.Save(seq); err != nil {
		h.logger.WithError(err).Error("Failed to save sequence after status change")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"sequence_id": seq.ID(),
		"tenant_id":   seq.TenantID(),
		"action":      cmd.Action,
		"new_status":  seq.Status(),
	}).Info("Sequence status changed successfully")

	return seq, nil
}
