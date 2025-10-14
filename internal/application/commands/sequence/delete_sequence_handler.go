package sequence

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/sequence"
)

// DeleteSequenceHandler handler para o comando DeleteSequence
type DeleteSequenceHandler struct {
	repository sequence.Repository
	logger     *logrus.Logger
}

// NewDeleteSequenceHandler cria uma nova instância do handler
func NewDeleteSequenceHandler(repository sequence.Repository, logger *logrus.Logger) *DeleteSequenceHandler {
	return &DeleteSequenceHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de deleção de sequência
func (h *DeleteSequenceHandler) Handle(ctx context.Context, cmd DeleteSequenceCommand) error {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid DeleteSequence command")
		return err
	}

	// Find sequence
	seq, err := h.repository.FindByID(cmd.SequenceID)
	if err != nil {
		h.logger.WithError(err).WithField("sequence_id", cmd.SequenceID).Error("Sequence not found")
		return fmt.Errorf("%w: %v", ErrSequenceNotFound, err)
	}

	if seq == nil {
		h.logger.WithField("sequence_id", cmd.SequenceID).Warn("Sequence not found (nil returned)")
		return ErrSequenceNotFound
	}

	// Validate tenant ownership
	if seq.TenantID() != cmd.TenantID {
		h.logger.WithFields(logrus.Fields{
			"sequence_id":      cmd.SequenceID,
			"sequence_tenant":  seq.TenantID(),
			"requested_tenant": cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return ErrAccessDenied
	}

	// Only allow deletion of draft sequences
	if seq.Status() != sequence.SequenceStatusDraft {
		h.logger.WithFields(logrus.Fields{
			"sequence_id": cmd.SequenceID,
			"status":      seq.Status(),
		}).Warn("Cannot delete non-draft sequence")
		return fmt.Errorf("%w: can only delete sequences in draft status", ErrInvalidSequenceStatus)
	}

	// Delete from repository
	if err := h.repository.Delete(cmd.SequenceID); err != nil {
		h.logger.WithError(err).Error("Failed to delete sequence")
		return fmt.Errorf("%w: %v", ErrSequenceDeleteFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"sequence_id": cmd.SequenceID,
		"tenant_id":   cmd.TenantID,
	}).Info("Sequence deleted successfully")

	return nil
}
