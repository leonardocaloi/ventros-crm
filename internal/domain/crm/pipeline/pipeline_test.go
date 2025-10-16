package pipeline

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

func TestNewPipeline(t *testing.T) {
	tests := []struct {
		name         string
		projectID    uuid.UUID
		tenantID     string
		pipelineName string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid pipeline creation",
			projectID:    uuid.New(),
			tenantID:     "tenant-123",
			pipelineName: "Sales Pipeline",
			wantErr:      false,
		},
		{
			name:         "nil projectID",
			projectID:    uuid.Nil,
			tenantID:     "tenant-123",
			pipelineName: "Sales Pipeline",
			wantErr:      true,
			errMsg:       "projectID cannot be nil",
		},
		{
			name:         "empty tenantID",
			projectID:    uuid.New(),
			tenantID:     "",
			pipelineName: "Sales Pipeline",
			wantErr:      true,
			errMsg:       "tenantID cannot be empty",
		},
		{
			name:         "empty name",
			projectID:    uuid.New(),
			tenantID:     "tenant-123",
			pipelineName: "",
			wantErr:      true,
			errMsg:       "name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := NewPipeline(tt.projectID, tt.tenantID, tt.pipelineName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, pipeline)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pipeline)

				assert.NotEqual(t, uuid.Nil, pipeline.ID())
				assert.Equal(t, tt.projectID, pipeline.ProjectID())
				assert.Equal(t, tt.tenantID, pipeline.TenantID())
				assert.Equal(t, tt.pipelineName, pipeline.Name())
				assert.True(t, pipeline.IsActive())
				// sessionTimeoutMinutes is nil by default (inherits from channel/project)
				assert.Nil(t, pipeline.SessionTimeoutMinutes())
				assert.Equal(t, 0, pipeline.Position())
				assert.NotZero(t, pipeline.CreatedAt())
				assert.NotZero(t, pipeline.UpdatedAt())

				// Check domain event
				events := pipeline.DomainEvents()
				require.Len(t, events, 1)
				event, ok := events[0].(PipelineCreatedEvent)
				require.True(t, ok)
				assert.Equal(t, pipeline.ID(), event.PipelineID)
				assert.Equal(t, tt.projectID, event.ProjectID)
			}
		})
	}
}

func TestPipeline_UpdateName(t *testing.T) {
	pipeline, err := NewPipeline(uuid.New(), "tenant-123", "Original Name")
	require.NoError(t, err)
	pipeline.ClearEvents() // Clear creation event

	t.Run("successful name update", func(t *testing.T) {
		err := pipeline.UpdateName("New Name")
		require.NoError(t, err)

		assert.Equal(t, "New Name", pipeline.Name())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(PipelineUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "name", event.Field)
		assert.Equal(t, "Original Name", event.OldValue)
		assert.Equal(t, "New Name", event.NewValue)
	})

	t.Run("empty name", func(t *testing.T) {
		err := pipeline.UpdateName("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})
}

func TestPipeline_SessionTimeout(t *testing.T) {
	pipeline, err := NewPipeline(uuid.New(), "tenant-123", "Test Pipeline")
	require.NoError(t, err)

	t.Run("set valid timeout", func(t *testing.T) {
		err := pipeline.SetSessionTimeout(intPtr(60))
		require.NoError(t, err)
		require.NotNil(t, pipeline.SessionTimeoutMinutes())
		assert.Equal(t, 60, *pipeline.SessionTimeoutMinutes())
	})

	t.Run("zero timeout", func(t *testing.T) {
		err := pipeline.SetSessionTimeout(intPtr(0))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})

	t.Run("negative timeout", func(t *testing.T) {
		err := pipeline.SetSessionTimeout(intPtr(-10))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})

	t.Run("timeout exceeds maximum", func(t *testing.T) {
		err := pipeline.SetSessionTimeout(intPtr(1500))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed 1440 minutes")
	})

	t.Run("maximum allowed timeout", func(t *testing.T) {
		err := pipeline.SetSessionTimeout(intPtr(1440))
		require.NoError(t, err)
		require.NotNil(t, pipeline.SessionTimeoutMinutes())
		assert.Equal(t, 1440, *pipeline.SessionTimeoutMinutes())
	})
}

func TestPipeline_ActivateDeactivate(t *testing.T) {
	pipeline, err := NewPipeline(uuid.New(), "tenant-123", "Test Pipeline")
	require.NoError(t, err)
	pipeline.ClearEvents()

	t.Run("deactivate active pipeline", func(t *testing.T) {
		assert.True(t, pipeline.IsActive())

		pipeline.Deactivate()
		assert.False(t, pipeline.IsActive())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(PipelineDeactivatedEvent)
		require.True(t, ok)
	})

	t.Run("deactivate already inactive pipeline", func(t *testing.T) {
		pipeline.ClearEvents()
		pipeline.Deactivate()

		// Should not generate duplicate event
		events := pipeline.DomainEvents()
		assert.Len(t, events, 0)
	})

	t.Run("activate inactive pipeline", func(t *testing.T) {
		pipeline.ClearEvents()
		pipeline.Activate()
		assert.True(t, pipeline.IsActive())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(PipelineActivatedEvent)
		require.True(t, ok)
	})

	t.Run("activate already active pipeline", func(t *testing.T) {
		pipeline.ClearEvents()
		pipeline.Activate()

		// Should not generate duplicate event
		events := pipeline.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestPipeline_StatusManagement(t *testing.T) {
	pipeline, err := NewPipeline(uuid.New(), "tenant-123", "Test Pipeline")
	require.NoError(t, err)
	pipeline.ClearEvents()

	status1, err := NewStatus(pipeline.ID(), "New", StatusTypeOpen)
	require.NoError(t, err)
	status1.UpdatePosition(1)

	status2, err := NewStatus(pipeline.ID(), "In Progress", StatusTypeActive)
	require.NoError(t, err)
	status2.UpdatePosition(2)

	t.Run("add status", func(t *testing.T) {
		err := pipeline.AddStatus(status1)
		require.NoError(t, err)

		statuses := pipeline.Statuses()
		assert.Len(t, statuses, 1)
		assert.Equal(t, "New", statuses[0].Name())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(StatusAddedToPipelineEvent)
		require.True(t, ok)
		assert.Equal(t, status1.ID(), event.StatusID)
	})

	t.Run("add nil status", func(t *testing.T) {
		err := pipeline.AddStatus(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status cannot be nil")
	})

	t.Run("add duplicate status name", func(t *testing.T) {
		duplicateStatus, err := NewStatus(pipeline.ID(), "New", StatusTypeActive)
		require.NoError(t, err)

		err = pipeline.AddStatus(duplicateStatus)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status with this name already exists")
	})

	t.Run("add second status", func(t *testing.T) {
		pipeline.ClearEvents()
		err := pipeline.AddStatus(status2)
		require.NoError(t, err)

		statuses := pipeline.Statuses()
		assert.Len(t, statuses, 2)
	})

	t.Run("get status by ID", func(t *testing.T) {
		found := pipeline.GetStatusByID(status1.ID())
		require.NotNil(t, found)
		assert.Equal(t, "New", found.Name())

		notFound := pipeline.GetStatusByID(uuid.New())
		assert.Nil(t, notFound)
	})

	t.Run("get status by name", func(t *testing.T) {
		found := pipeline.GetStatusByName("In Progress")
		require.NotNil(t, found)
		assert.Equal(t, status2.ID(), found.ID())

		notFound := pipeline.GetStatusByName("Nonexistent")
		assert.Nil(t, notFound)
	})

	t.Run("remove status", func(t *testing.T) {
		pipeline.ClearEvents()
		err := pipeline.RemoveStatus(status1.ID())
		require.NoError(t, err)

		statuses := pipeline.Statuses()
		assert.Len(t, statuses, 1)
		assert.Equal(t, "In Progress", statuses[0].Name())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(StatusRemovedFromPipelineEvent)
		require.True(t, ok)
		assert.Equal(t, status1.ID(), event.StatusID)
	})

	t.Run("remove nonexistent status", func(t *testing.T) {
		err := pipeline.RemoveStatus(uuid.New())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status not found")
	})
}

func TestPipeline_UpdateDescriptionAndColor(t *testing.T) {
	pipeline, err := NewPipeline(uuid.New(), "tenant-123", "Test Pipeline")
	require.NoError(t, err)
	pipeline.ClearEvents()

	t.Run("update description", func(t *testing.T) {
		pipeline.UpdateDescription("New description")
		assert.Equal(t, "New description", pipeline.Description())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(PipelineUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "description", event.Field)
	})

	t.Run("update color", func(t *testing.T) {
		pipeline.ClearEvents()
		pipeline.UpdateColor("#FF5733")
		assert.Equal(t, "#FF5733", pipeline.Color())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(PipelineUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "color", event.Field)
	})

	t.Run("update position", func(t *testing.T) {
		pipeline.ClearEvents()
		pipeline.UpdatePosition(5)
		assert.Equal(t, 5, pipeline.Position())

		events := pipeline.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(PipelineUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "position", event.Field)
	})
}

func TestPipeline_ReconstructPipeline(t *testing.T) {
	id := uuid.New()
	projectID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	t.Run("reconstruct with valid data", func(t *testing.T) {
		pipeline := ReconstructPipeline(
			id, projectID,
			1, // version
			"tenant-123", "Reconstructed Pipeline", "Test description", "#FF0000",
			2, true, intPtr(45), nil, // nil leadQualificationConfig
			createdAt, updatedAt,
		)

		assert.Equal(t, id, pipeline.ID())
		assert.Equal(t, projectID, pipeline.ProjectID())
		assert.Equal(t, "tenant-123", pipeline.TenantID())
		assert.Equal(t, "Reconstructed Pipeline", pipeline.Name())
		assert.Equal(t, "Test description", pipeline.Description())
		assert.Equal(t, "#FF0000", pipeline.Color())
		assert.Equal(t, 2, pipeline.Position())
		assert.True(t, pipeline.IsActive())
		require.NotNil(t, pipeline.SessionTimeoutMinutes())
		assert.Equal(t, 45, *pipeline.SessionTimeoutMinutes())
		assert.Equal(t, createdAt, pipeline.CreatedAt())
		assert.Equal(t, updatedAt, pipeline.UpdatedAt())
		assert.Len(t, pipeline.DomainEvents(), 0) // No events on reconstruction
	})

	t.Run("reconstruct with zero timeout preserves value", func(t *testing.T) {
		pipeline := ReconstructPipeline(
			id, projectID,
			1, // version
			"tenant-123", "Pipeline", "", "",
			0, true, intPtr(0), nil, // zero timeout, nil leadQualificationConfig
			createdAt, updatedAt,
		)

		require.NotNil(t, pipeline.SessionTimeoutMinutes())
		assert.Equal(t, 0, *pipeline.SessionTimeoutMinutes())
	})

	t.Run("reconstruct with negative timeout preserves value", func(t *testing.T) {
		pipeline := ReconstructPipeline(
			id, projectID,
			1, // version
			"tenant-123", "Pipeline", "", "",
			0, true, intPtr(-10), nil, // negative timeout, nil leadQualificationConfig
			createdAt, updatedAt,
		)

		require.NotNil(t, pipeline.SessionTimeoutMinutes())
		assert.Equal(t, -10, *pipeline.SessionTimeoutMinutes())
	})
}

func TestPipeline_EventManagement(t *testing.T) {
	pipeline, err := NewPipeline(uuid.New(), "tenant-123", "Test Pipeline")
	require.NoError(t, err)

	t.Run("clear events", func(t *testing.T) {
		assert.Len(t, pipeline.DomainEvents(), 1) // Creation event

		pipeline.ClearEvents()
		assert.Len(t, pipeline.DomainEvents(), 0)
	})

	t.Run("multiple operations generate multiple events", func(t *testing.T) {
		pipeline.ClearEvents()

		pipeline.UpdateDescription("New desc")
		pipeline.UpdateColor("#000000")
		pipeline.Deactivate()

		events := pipeline.DomainEvents()
		assert.Len(t, events, 3)
	})

	t.Run("events are immutable copies", func(t *testing.T) {
		pipeline.ClearEvents()
		pipeline.UpdateName("Test")

		events1 := pipeline.DomainEvents()
		events2 := pipeline.DomainEvents()

		// Should be different slices (copies)
		assert.NotSame(t, &events1, &events2)
		assert.Equal(t, len(events1), len(events2))
	})
}
