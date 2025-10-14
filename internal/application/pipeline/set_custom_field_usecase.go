package pipeline

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
)

// SetCustomFieldUseCase handles setting/updating a custom field for a pipeline
type SetCustomFieldUseCase struct {
	repository pipeline.Repository
	logger     *logrus.Logger
}

// NewSetCustomFieldUseCase creates a new instance
func NewSetCustomFieldUseCase(repository pipeline.Repository, logger *logrus.Logger) *SetCustomFieldUseCase {
	return &SetCustomFieldUseCase{
		repository: repository,
		logger:     logger,
	}
}

// SetCustomFieldCommand represents the input for setting a custom field
type SetCustomFieldCommand struct {
	PipelineID uuid.UUID
	TenantID   string
	Key        string
	Type       shared.FieldType
	Value      interface{}
}

// Validate validates the command
func (c *SetCustomFieldCommand) Validate() error {
	if c.PipelineID == uuid.Nil {
		return pipeline.ErrPipelineIDRequired
	}
	if c.TenantID == "" {
		return pipeline.ErrTenantIDRequired
	}
	if c.Key == "" {
		return pipeline.ErrCustomFieldKeyRequired
	}
	if !c.Type.IsValid() {
		return pipeline.ErrCustomFieldInvalidType
	}
	return nil
}

// Execute executes the use case
func (uc *SetCustomFieldUseCase) Execute(ctx context.Context, cmd SetCustomFieldCommand) (*pipeline.PipelineCustomField, error) {
	// Validate command
	if err := cmd.Validate(); err != nil {
		uc.logger.WithError(err).Error("Invalid command")
		return nil, err
	}

	// Verify pipeline exists and belongs to tenant
	p, err := uc.repository.FindPipelineByID(ctx, cmd.PipelineID)
	if err != nil {
		uc.logger.WithError(err).WithField("pipeline_id", cmd.PipelineID).Error("Pipeline not found")
		return nil, pipeline.ErrPipelineNotFound
	}

	if p.TenantID() != cmd.TenantID {
		uc.logger.WithFields(logrus.Fields{
			"pipeline_id": cmd.PipelineID,
			"tenant_id":   cmd.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return nil, fmt.Errorf("access denied")
	}

	// Check if custom field already exists
	existing, err := uc.repository.FindCustomFieldByKey(ctx, cmd.PipelineID, cmd.Key)
	if err != nil && err != pipeline.ErrCustomFieldNotFound {
		uc.logger.WithError(err).Error("Failed to check existing custom field")
		return nil, err
	}

	var customFieldEntity *pipeline.PipelineCustomField

	if existing != nil {
		// Update existing field
		newField, err := shared.NewCustomField(cmd.Key, cmd.Type, cmd.Value)
		if err != nil {
			uc.logger.WithError(err).Error("Failed to create custom field value object")
			return nil, pipeline.ErrCustomFieldInvalidValue
		}

		if err := existing.UpdateValue(newField); err != nil {
			uc.logger.WithError(err).Error("Failed to update custom field value")
			return nil, err
		}

		customFieldEntity = existing
	} else {
		// Create new field
		newField, err := shared.NewCustomField(cmd.Key, cmd.Type, cmd.Value)
		if err != nil {
			uc.logger.WithError(err).Error("Failed to create custom field value object")
			return nil, pipeline.ErrCustomFieldInvalidValue
		}

		customFieldEntity, err = pipeline.NewPipelineCustomField(cmd.PipelineID, cmd.TenantID, newField)
		if err != nil {
			uc.logger.WithError(err).Error("Failed to create pipeline custom field")
			return nil, err
		}
	}

	// Save to repository
	if err := uc.repository.SaveCustomField(ctx, customFieldEntity); err != nil {
		uc.logger.WithError(err).Error("Failed to save custom field")
		return nil, pipeline.ErrCustomFieldOperationFailed
	}

	uc.logger.WithFields(logrus.Fields{
		"pipeline_id": cmd.PipelineID,
		"field_key":   cmd.Key,
		"field_type":  cmd.Type,
	}).Info("Custom field set successfully")

	return customFieldEntity, nil
}
