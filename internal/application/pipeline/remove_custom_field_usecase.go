package pipeline

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
)

// RemoveCustomFieldUseCase handles removing a custom field from a pipeline
type RemoveCustomFieldUseCase struct {
	repository pipeline.Repository
	logger     *logrus.Logger
}

// NewRemoveCustomFieldUseCase creates a new instance
func NewRemoveCustomFieldUseCase(repository pipeline.Repository, logger *logrus.Logger) *RemoveCustomFieldUseCase {
	return &RemoveCustomFieldUseCase{
		repository: repository,
		logger:     logger,
	}
}

// RemoveCustomFieldCommand represents the input for removing a custom field
type RemoveCustomFieldCommand struct {
	PipelineID uuid.UUID
	TenantID   string
	Key        string
}

// Validate validates the command
func (c *RemoveCustomFieldCommand) Validate() error {
	if c.PipelineID == uuid.Nil {
		return pipeline.ErrPipelineIDRequired
	}
	if c.TenantID == "" {
		return pipeline.ErrTenantIDRequired
	}
	if c.Key == "" {
		return pipeline.ErrCustomFieldKeyRequired
	}
	return nil
}

// Execute executes the use case
func (uc *RemoveCustomFieldUseCase) Execute(ctx context.Context, cmd RemoveCustomFieldCommand) error {
	// Validate command
	if err := cmd.Validate(); err != nil {
		uc.logger.WithError(err).Error("Invalid command")
		return err
	}

	// Verify pipeline exists and belongs to tenant
	p, err := uc.repository.FindPipelineByID(ctx, cmd.PipelineID)
	if err != nil {
		uc.logger.WithError(err).WithField("pipeline_id", cmd.PipelineID).Error("Pipeline not found")
		return pipeline.ErrPipelineNotFound
	}

	if p.TenantID() != cmd.TenantID {
		uc.logger.WithFields(logrus.Fields{
			"pipeline_id": cmd.PipelineID,
			"tenant_id":   cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return fmt.Errorf("access denied")
	}

	// Verify custom field exists
	_, err = uc.repository.FindCustomFieldByKey(ctx, cmd.PipelineID, cmd.Key)
	if err != nil {
		if err == pipeline.ErrCustomFieldNotFound {
			uc.logger.WithFields(logrus.Fields{
				"pipeline_id": cmd.PipelineID,
				"field_key":   cmd.Key,
			}).Warn("Custom field not found")
			return err
		}
		uc.logger.WithError(err).Error("Failed to check custom field existence")
		return pipeline.ErrCustomFieldOperationFailed
	}

	// Delete custom field
	if err := uc.repository.DeleteCustomFieldByKey(ctx, cmd.PipelineID, cmd.Key); err != nil {
		uc.logger.WithError(err).Error("Failed to delete custom field")
		return pipeline.ErrCustomFieldOperationFailed
	}

	uc.logger.WithFields(logrus.Fields{
		"pipeline_id": cmd.PipelineID,
		"field_key":   cmd.Key,
	}).Info("Custom field removed successfully")

	return nil
}
