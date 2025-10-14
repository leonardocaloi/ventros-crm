package pipeline_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/pipeline"
)

func TestNewPipelineCustomField(t *testing.T) {
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	t.Run("should create new pipeline custom field with valid data", func(t *testing.T) {
		customField, err := shared.NewCustomField("company_size", shared.FieldTypeText, "100-500")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)

		require.NoError(t, err)
		assert.NotNil(t, pcf)
		assert.NotEqual(t, uuid.Nil, pcf.ID())
		assert.Equal(t, pipelineID, pcf.PipelineID())
		assert.Equal(t, tenantID, pcf.TenantID())
		assert.Equal(t, "company_size", pcf.CustomField().Key())
		assert.Equal(t, shared.FieldTypeText, pcf.CustomField().Type())
		assert.Equal(t, "100-500", pcf.CustomField().Value())
		assert.False(t, pcf.CreatedAt().IsZero())
		assert.False(t, pcf.UpdatedAt().IsZero())
	})

	t.Run("should fail with nil pipeline ID", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(uuid.Nil, tenantID, customField)

		assert.Error(t, err)
		assert.Nil(t, pcf)
		assert.Equal(t, pipeline.ErrPipelineIDRequired, err)
	})

	t.Run("should fail with empty tenant ID", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, "", customField)

		assert.Error(t, err)
		assert.Nil(t, pcf)
		assert.Equal(t, pipeline.ErrTenantIDRequired, err)
	})

	t.Run("should fail with nil custom field", func(t *testing.T) {
		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, nil)

		assert.Error(t, err)
		assert.Nil(t, pcf)
		assert.Equal(t, pipeline.ErrCustomFieldRequired, err)
	})
}

func TestReconstructPipelineCustomField(t *testing.T) {
	id := uuid.New()
	pipelineID := uuid.New()
	tenantID := "tenant-123"
	now := time.Now()

	t.Run("should reconstruct pipeline custom field with all data", func(t *testing.T) {
		customField, err := shared.NewCustomField("annual_revenue", shared.FieldTypeNumber, 1000000.50)
		require.NoError(t, err)

		pcf, err := pipeline.ReconstructPipelineCustomField(id, pipelineID, tenantID, customField, now, now)

		require.NoError(t, err)
		assert.NotNil(t, pcf)
		assert.Equal(t, id, pcf.ID())
		assert.Equal(t, pipelineID, pcf.PipelineID())
		assert.Equal(t, tenantID, pcf.TenantID())
		assert.Equal(t, "annual_revenue", pcf.CustomField().Key())
		assert.Equal(t, shared.FieldTypeNumber, pcf.CustomField().Type())
		assert.Equal(t, 1000000.50, pcf.CustomField().Value())
		assert.Equal(t, now, pcf.CreatedAt())
		assert.Equal(t, now, pcf.UpdatedAt())
	})

	t.Run("should fail with nil ID", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.ReconstructPipelineCustomField(uuid.Nil, pipelineID, tenantID, customField, now, now)

		assert.Error(t, err)
		assert.Nil(t, pcf)
		assert.Equal(t, pipeline.ErrCustomFieldIDRequired, err)
	})

	t.Run("should fail with zero created at", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.ReconstructPipelineCustomField(id, pipelineID, tenantID, customField, time.Time{}, now)

		assert.Error(t, err)
		assert.Nil(t, pcf)
		assert.Equal(t, pipeline.ErrInvalidTimestamp, err)
	})

	t.Run("should fail with zero updated at", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.ReconstructPipelineCustomField(id, pipelineID, tenantID, customField, now, time.Time{})

		assert.Error(t, err)
		assert.Nil(t, pcf)
		assert.Equal(t, pipeline.ErrInvalidTimestamp, err)
	})
}

func TestUpdateValue(t *testing.T) {
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	t.Run("should update value successfully", func(t *testing.T) {
		customField, err := shared.NewCustomField("status", shared.FieldTypeText, "active")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		originalUpdatedAt := pcf.UpdatedAt()
		time.Sleep(1 * time.Millisecond) // Ensure time difference

		newField, err := shared.NewCustomField("status", shared.FieldTypeText, "inactive")
		require.NoError(t, err)

		err = pcf.UpdateValue(newField)

		require.NoError(t, err)
		assert.Equal(t, "inactive", pcf.CustomField().Value())
		assert.True(t, pcf.UpdatedAt().After(originalUpdatedAt))
	})

	t.Run("should fail when updating with nil field", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		err = pcf.UpdateValue(nil)

		assert.Error(t, err)
		assert.Equal(t, pipeline.ErrCustomFieldRequired, err)
	})

	t.Run("should fail when trying to change field key", func(t *testing.T) {
		customField, err := shared.NewCustomField("original_key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		newField, err := shared.NewCustomField("different_key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		err = pcf.UpdateValue(newField)

		assert.Error(t, err)
		assert.Equal(t, pipeline.ErrCustomFieldKeyImmutable, err)
		assert.Equal(t, "original_key", pcf.CustomField().Key()) // Should remain unchanged
	})

	t.Run("should fail when trying to change field type", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		newField, err := shared.NewCustomField("key", shared.FieldTypeNumber, 123)
		require.NoError(t, err)

		err = pcf.UpdateValue(newField)

		assert.Error(t, err)
		assert.Equal(t, pipeline.ErrCustomFieldTypeImmutable, err)
		assert.Equal(t, shared.FieldTypeText, pcf.CustomField().Type()) // Should remain unchanged
	})

	t.Run("should update complex types correctly", func(t *testing.T) {
		// Test with JSON type
		jsonData := map[string]interface{}{
			"nested": map[string]interface{}{
				"field": "value",
			},
		}
		customField, err := shared.NewCustomField("metadata", shared.FieldTypeJSON, jsonData)
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		newJsonData := map[string]interface{}{
			"nested": map[string]interface{}{
				"field": "updated_value",
			},
		}
		newField, err := shared.NewCustomField("metadata", shared.FieldTypeJSON, newJsonData)
		require.NoError(t, err)

		err = pcf.UpdateValue(newField)

		require.NoError(t, err)
		assert.Equal(t, newJsonData, pcf.CustomField().Value())
	})
}

func TestPipelineCustomFieldGetters(t *testing.T) {
	pipelineID := uuid.New()
	tenantID := "tenant-123"
	customField, err := shared.NewCustomField("test_key", shared.FieldTypeBoolean, true)
	require.NoError(t, err)

	pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
	require.NoError(t, err)

	t.Run("ID returns correct value", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, pcf.ID())
	})

	t.Run("PipelineID returns correct value", func(t *testing.T) {
		assert.Equal(t, pipelineID, pcf.PipelineID())
	})

	t.Run("TenantID returns correct value", func(t *testing.T) {
		assert.Equal(t, tenantID, pcf.TenantID())
	})

	t.Run("CustomField returns correct value", func(t *testing.T) {
		cf := pcf.CustomField()
		assert.NotNil(t, cf)
		assert.Equal(t, "test_key", cf.Key())
		assert.Equal(t, shared.FieldTypeBoolean, cf.Type())
		assert.Equal(t, true, cf.Value())
	})

	t.Run("CreatedAt returns valid timestamp", func(t *testing.T) {
		assert.False(t, pcf.CreatedAt().IsZero())
		assert.True(t, pcf.CreatedAt().Before(time.Now().Add(1*time.Second)))
	})

	t.Run("UpdatedAt returns valid timestamp", func(t *testing.T) {
		assert.False(t, pcf.UpdatedAt().IsZero())
		assert.True(t, pcf.UpdatedAt().Before(time.Now().Add(1*time.Second)))
	})
}

func TestPipelineCustomFieldWithDifferentFieldTypes(t *testing.T) {
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	testCases := []struct {
		name      string
		fieldKey  string
		fieldType shared.FieldType
		value     interface{}
	}{
		{
			name:      "text field",
			fieldKey:  "description",
			fieldType: shared.FieldTypeText,
			value:     "Some description text",
		},
		{
			name:      "number field",
			fieldKey:  "priority",
			fieldType: shared.FieldTypeNumber,
			value:     42.5,
		},
		{
			name:      "boolean field",
			fieldKey:  "is_active",
			fieldType: shared.FieldTypeBoolean,
			value:     true,
		},
		{
			name:      "date field",
			fieldKey:  "deadline",
			fieldType: shared.FieldTypeDate,
			value:     time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "email field",
			fieldKey:  "contact_email",
			fieldType: shared.FieldTypeEmail,
			value:     "test@example.com",
		},
		{
			name:      "url field",
			fieldKey:  "website",
			fieldType: shared.FieldTypeURL,
			value:     "https://example.com",
		},
		{
			name:      "phone field",
			fieldKey:  "phone",
			fieldType: shared.FieldTypePhone,
			value:     "+1234567890",
		},
		{
			name:      "select field",
			fieldKey:  "category",
			fieldType: shared.FieldTypeSelect,
			value:     "option1",
		},
		{
			name:      "multi_select field",
			fieldKey:  "tags",
			fieldType: shared.FieldTypeMultiSelect,
			value:     []string{"tag1", "tag2", "tag3"},
		},
		{
			name:      "json field",
			fieldKey:  "metadata",
			fieldType: shared.FieldTypeJSON,
			value: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		},
		{
			name:      "label field",
			fieldKey:  "status_label",
			fieldType: shared.FieldTypeLabel,
			value:     []string{"important", "urgent"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			customField, err := shared.NewCustomField(tc.fieldKey, tc.fieldType, tc.value)
			require.NoError(t, err)

			pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)

			require.NoError(t, err)
			assert.NotNil(t, pcf)
			assert.Equal(t, tc.fieldKey, pcf.CustomField().Key())
			assert.Equal(t, tc.fieldType, pcf.CustomField().Type())
			assert.Equal(t, tc.value, pcf.CustomField().Value())
		})
	}
}

func TestPipelineCustomFieldImmutability(t *testing.T) {
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	t.Run("should not allow changing pipeline ID", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		// PipelineID getter should always return the same value
		assert.Equal(t, pipelineID, pcf.PipelineID())
		// No setter exists, so this is enforced by design
	})

	t.Run("should not allow changing tenant ID", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		// TenantID getter should always return the same value
		assert.Equal(t, tenantID, pcf.TenantID())
		// No setter exists, so this is enforced by design
	})

	t.Run("should not allow changing ID", func(t *testing.T) {
		customField, err := shared.NewCustomField("key", shared.FieldTypeText, "value")
		require.NoError(t, err)

		pcf, err := pipeline.NewPipelineCustomField(pipelineID, tenantID, customField)
		require.NoError(t, err)

		originalID := pcf.ID()
		// No setter exists, ID should remain immutable
		assert.Equal(t, originalID, pcf.ID())
	})
}
