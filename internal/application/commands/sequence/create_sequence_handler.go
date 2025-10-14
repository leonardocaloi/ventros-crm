package sequence

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/automation/sequence"
)

// CreateSequenceHandler handler para o comando CreateSequence
type CreateSequenceHandler struct {
	repository sequence.Repository
	logger     *logrus.Logger
}

// NewCreateSequenceHandler cria uma nova instância do handler
func NewCreateSequenceHandler(repository sequence.Repository, logger *logrus.Logger) *CreateSequenceHandler {
	return &CreateSequenceHandler{
		repository: repository,
		logger:     logger,
	}
}

// Handle executa o comando de criação de sequência
func (h *CreateSequenceHandler) Handle(ctx context.Context, cmd CreateSequenceCommand) (*sequence.Sequence, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("Invalid CreateSequence command")
		return nil, err
	}

	// Create domain sequence
	seq, err := sequence.NewSequence(
		cmd.TenantID,
		cmd.Name,
		cmd.Description,
		sequence.TriggerType(cmd.TriggerType),
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create sequence domain object")
		return nil, fmt.Errorf("%w: %v", ErrSequenceCreationFailed, err)
	}

	// Add steps if provided
	for _, stepInput := range cmd.Steps {
		// Convert string to *string for optional fields
		var templateID *string
		if stepInput.MessageTemplate.TemplateID != "" {
			templateID = &stepInput.MessageTemplate.TemplateID
		}

		var mediaURL *string
		if stepInput.MessageTemplate.MediaURL != "" {
			mediaURL = &stepInput.MessageTemplate.MediaURL
		}

		template := sequence.MessageTemplate{
			Type:       stepInput.MessageTemplate.Type,
			Content:    stepInput.MessageTemplate.Content,
			TemplateID: templateID,
			Variables:  stepInput.MessageTemplate.Variables,
			MediaURL:   mediaURL,
		}

		step := sequence.NewSequenceStep(
			stepInput.Order,
			stepInput.Name,
			stepInput.DelayAmount,
			sequence.DelayUnit(stepInput.DelayUnit),
			template,
		)

		if err := seq.AddStep(step); err != nil {
			h.logger.WithError(err).WithField("step_name", stepInput.Name).Error("Failed to add step to sequence")
			return nil, fmt.Errorf("%w: failed to add step: %v", ErrSequenceCreationFailed, err)
		}
	}

	// Save to repository
	if err := h.repository.Save(seq); err != nil {
		h.logger.WithError(err).Error("Failed to save sequence to repository")
		return nil, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	h.logger.WithFields(logrus.Fields{
		"sequence_id": seq.ID(),
		"tenant_id":   seq.TenantID(),
		"name":        seq.Name(),
		"steps_count": len(cmd.Steps),
	}).Info("Sequence created successfully")

	return seq, nil
}
