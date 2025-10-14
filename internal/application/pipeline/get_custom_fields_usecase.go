package pipeline

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
)

// GetCustomFieldsUseCase handles retrieving custom fields for a pipeline
type GetCustomFieldsUseCase struct {
	repository pipeline.Repository
	logger     *logrus.Logger
}

// NewGetCustomFieldsUseCase creates a new instance
func NewGetCustomFieldsUseCase(repository pipeline.Repository, logger *logrus.Logger) *GetCustomFieldsUseCase {
	return &GetCustomFieldsUseCase{
		repository: repository,
		logger:     logger,
	}
}

// GetCustomFieldsQuery represents the query input
type GetCustomFieldsQuery struct {
	PipelineID uuid.UUID
	TenantID   string
}

// Validate validates the query
func (q *GetCustomFieldsQuery) Validate() error {
	if q.PipelineID == uuid.Nil {
		return pipeline.ErrPipelineIDRequired
	}
	if q.TenantID == "" {
		return pipeline.ErrTenantIDRequired
	}
	return nil
}

// Execute executes the use case
func (uc *GetCustomFieldsUseCase) Execute(ctx context.Context, query GetCustomFieldsQuery) ([]*pipeline.PipelineCustomField, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		uc.logger.WithError(err).Error("Invalid query")
		return nil, err
	}

	// Verify pipeline exists and belongs to tenant
	p, err := uc.repository.FindPipelineByID(ctx, query.PipelineID)
	if err != nil {
		uc.logger.WithError(err).WithField("pipeline_id", query.PipelineID).Error("Pipeline not found")
		return nil, pipeline.ErrPipelineNotFound
	}

	if p.TenantID() != query.TenantID {
		uc.logger.WithFields(logrus.Fields{
			"pipeline_id": query.PipelineID,
			"tenant_id":   query.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return nil, fmt.Errorf("access denied")
	}

	// Retrieve custom fields
	fields, err := uc.repository.FindCustomFieldsByPipeline(ctx, query.PipelineID)
	if err != nil {
		uc.logger.WithError(err).Error("Failed to retrieve custom fields")
		return nil, pipeline.ErrCustomFieldOperationFailed
	}

	uc.logger.WithFields(logrus.Fields{
		"pipeline_id":  query.PipelineID,
		"fields_count": len(fields),
	}).Debug("Custom fields retrieved successfully")

	return fields, nil
}

// GetCustomFieldByKeyUseCase handles retrieving a specific custom field by key
type GetCustomFieldByKeyUseCase struct {
	repository pipeline.Repository
	logger     *logrus.Logger
}

// NewGetCustomFieldByKeyUseCase creates a new instance
func NewGetCustomFieldByKeyUseCase(repository pipeline.Repository, logger *logrus.Logger) *GetCustomFieldByKeyUseCase {
	return &GetCustomFieldByKeyUseCase{
		repository: repository,
		logger:     logger,
	}
}

// GetCustomFieldByKeyQuery represents the query input
type GetCustomFieldByKeyQuery struct {
	PipelineID uuid.UUID
	TenantID   string
	Key        string
}

// Validate validates the query
func (q *GetCustomFieldByKeyQuery) Validate() error {
	if q.PipelineID == uuid.Nil {
		return pipeline.ErrPipelineIDRequired
	}
	if q.TenantID == "" {
		return pipeline.ErrTenantIDRequired
	}
	if q.Key == "" {
		return pipeline.ErrCustomFieldKeyRequired
	}
	return nil
}

// Execute executes the use case
func (uc *GetCustomFieldByKeyUseCase) Execute(ctx context.Context, query GetCustomFieldByKeyQuery) (*pipeline.PipelineCustomField, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		uc.logger.WithError(err).Error("Invalid query")
		return nil, err
	}

	// Verify pipeline exists and belongs to tenant
	p, err := uc.repository.FindPipelineByID(ctx, query.PipelineID)
	if err != nil {
		uc.logger.WithError(err).WithField("pipeline_id", query.PipelineID).Error("Pipeline not found")
		return nil, pipeline.ErrPipelineNotFound
	}

	if p.TenantID() != query.TenantID {
		uc.logger.WithFields(logrus.Fields{
			"pipeline_id": query.PipelineID,
			"tenant_id":   query.TenantID,
		}).Warn("Access denied: tenant mismatch")
		return nil, fmt.Errorf("access denied")
	}

	// Retrieve custom field
	field, err := uc.repository.FindCustomFieldByKey(ctx, query.PipelineID, query.Key)
	if err != nil {
		if err == pipeline.ErrCustomFieldNotFound {
			uc.logger.WithFields(logrus.Fields{
				"pipeline_id": query.PipelineID,
				"field_key":   query.Key,
			}).Warn("Custom field not found")
			return nil, err
		}
		uc.logger.WithError(err).Error("Failed to retrieve custom field")
		return nil, pipeline.ErrCustomFieldOperationFailed
	}

	uc.logger.WithFields(logrus.Fields{
		"pipeline_id": query.PipelineID,
		"field_key":   query.Key,
	}).Debug("Custom field retrieved successfully")

	return field, nil
}
