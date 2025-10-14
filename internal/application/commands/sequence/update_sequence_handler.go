package sequence

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/sequence"
)

// UpdateSequenceHandler handler para o comando UpdateSequence
type UpdateSequenceHandler struct {
	repository sequence.Repository
	logger     *logrus.Logger
}

// NewUpdateSequenceHandler cria uma nova instância do handler
func NewUpdateSequenceHandler(repository sequence.Repository, logger *logrus.Logger) *UpdateSequenceHandler {
	return &UpdateSequenceHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de atualização de sequência
func (h *UpdateSequenceHandler) Handle(ctx context.Context, cmd UpdateSequenceCommand) (*sequence.Sequence, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid UpdateSequence command")
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

	// Update name if provided
	if cmd.Name != nil {
		if err := seq.UpdateName(*cmd.Name); err != nil {
			h.logger.WithError(err).Error("Failed to update sequence name")
			return nil, fmt.Errorf("%w: %v", ErrSequenceUpdateFailed, err)
		}
	}

	// Update description if provided
	if cmd.Description != nil {
		seq.UpdateDescription(*cmd.Description)
	}

	// Update exit_on_reply if provided
	if cmd.ExitOnReply != nil {
		seq.UpdateExitOnReply(*cmd.ExitOnReply)
	}

	// Save to repository
	if err := h.repository.Save(seq); err != nil {
		h.logger.WithError(err).Error("Failed to save updated sequence")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"sequence_id": seq.ID(),
		"tenant_id":   seq.TenantID(),
	}).Info("Sequence updated successfully")

	return seq, nil
}
