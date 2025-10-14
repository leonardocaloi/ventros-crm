package pipeline

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// PipelineCustomField represents a custom field attached to a pipeline
type PipelineCustomField struct {
	id          uuid.UUID
	pipelineID  uuid.UUID
	tenantID    string
	customField *shared.CustomField
	createdAt   time.Time
	updatedAt   time.Time
}

// NewPipelineCustomField creates a new custom field for a pipeline
func NewPipelineCustomField(
	pipelineID uuid.UUID,
	tenantID string,
	customField *shared.CustomField,
) (*PipelineCustomField, error) {
	if pipelineID == uuid.Nil {
		return nil, ErrPipelineIDRequired
	}
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if customField == nil {
		return nil, ErrCustomFieldRequired
	}

	now := time.Now()
	return &PipelineCustomField{
		id:          uuid.New(),
		pipelineID:  pipelineID,
		tenantID:    tenantID,
		customField: customField,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// ReconstructPipelineCustomField reconstructs a custom field from persistence
func ReconstructPipelineCustomField(
	id uuid.UUID,
	pipelineID uuid.UUID,
	tenantID string,
	customField *shared.CustomField,
	createdAt time.Time,
	updatedAt time.Time,
) (*PipelineCustomField, error) {
	if id == uuid.Nil {
		return nil, ErrCustomFieldIDRequired
	}
	if pipelineID == uuid.Nil {
		return nil, ErrPipelineIDRequired
	}
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if customField == nil {
		return nil, ErrCustomFieldRequired
	}
	if createdAt.IsZero() {
		return nil, ErrInvalidTimestamp
	}
	if updatedAt.IsZero() {
		return nil, ErrInvalidTimestamp
	}

	return &PipelineCustomField{
		id:          id,
		pipelineID:  pipelineID,
		tenantID:    tenantID,
		customField: customField,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

// UpdateValue updates the value of the custom field
// Note: Key and Type cannot be changed
func (pcf *PipelineCustomField) UpdateValue(newField *shared.CustomField) error {
	if newField == nil {
		return ErrCustomFieldRequired
	}

	if pcf.customField.Key() != newField.Key() {
		return ErrCustomFieldKeyImmutable
	}
	if pcf.customField.Type() != newField.Type() {
		return ErrCustomFieldTypeImmutable
	}

	pcf.customField = newField
	pcf.updatedAt = time.Now()
	return nil
}

// Getters
func (pcf *PipelineCustomField) ID() uuid.UUID                    { return pcf.id }
func (pcf *PipelineCustomField) PipelineID() uuid.UUID            { return pcf.pipelineID }
func (pcf *PipelineCustomField) TenantID() string                 { return pcf.tenantID }
func (pcf *PipelineCustomField) CustomField() *shared.CustomField { return pcf.customField }
func (pcf *PipelineCustomField) FieldKey() string                 { return pcf.customField.Key() }
func (pcf *PipelineCustomField) FieldType() shared.FieldType      { return pcf.customField.Type() }
func (pcf *PipelineCustomField) FieldValue() interface{}          { return pcf.customField.Value() }
func (pcf *PipelineCustomField) CreatedAt() time.Time             { return pcf.createdAt }
func (pcf *PipelineCustomField) UpdatedAt() time.Time             { return pcf.updatedAt }
