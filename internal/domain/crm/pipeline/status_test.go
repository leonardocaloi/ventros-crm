package pipeline

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatus(t *testing.T) {
	pipelineID := uuid.New()

	tests := []struct {
		name       string
		pipelineID uuid.UUID
		statusName string
		statusType StatusType
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid open status",
			pipelineID: pipelineID,
			statusName: "New Lead",
			statusType: StatusTypeOpen,
			wantErr:    false,
		},
		{
			name:       "valid active status",
			pipelineID: pipelineID,
			statusName: "In Progress",
			statusType: StatusTypeActive,
			wantErr:    false,
		},
		{
			name:       "valid closed status",
			pipelineID: pipelineID,
			statusName: "Won",
			statusType: StatusTypeClosed,
			wantErr:    false,
		},
		{
			name:       "nil pipelineID",
			pipelineID: uuid.Nil,
			statusName: "Status",
			statusType: StatusTypeOpen,
			wantErr:    true,
			errMsg:     "pipelineID cannot be nil",
		},
		{
			name:       "empty name",
			pipelineID: pipelineID,
			statusName: "",
			statusType: StatusTypeOpen,
			wantErr:    true,
			errMsg:     "name cannot be empty",
		},
		{
			name:       "empty statusType",
			pipelineID: pipelineID,
			statusName: "Status",
			statusType: "",
			wantErr:    true,
			errMsg:     "statusType cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewStatus(tt.pipelineID, tt.statusName, tt.statusType)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, status)
			} else {
				require.NoError(t, err)
				require.NotNil(t, status)

				assert.NotEqual(t, uuid.Nil, status.ID())
				assert.Equal(t, tt.pipelineID, status.PipelineID())
				assert.Equal(t, tt.statusName, status.Name())
				assert.Equal(t, tt.statusType, status.StatusType())
				assert.Equal(t, 0, status.Position())
				assert.True(t, status.IsActiveStatus())
				assert.NotZero(t, status.CreatedAt())
				assert.NotZero(t, status.UpdatedAt())

				// Check domain event
				events := status.DomainEvents()
				require.Len(t, events, 1)
				event, ok := events[0].(StatusCreatedEvent)
				require.True(t, ok)
				assert.Equal(t, status.ID(), event.StatusID)
				assert.Equal(t, tt.pipelineID, event.PipelineID)
				assert.Equal(t, tt.statusName, event.Name)
				assert.Equal(t, tt.statusType, event.StatusType)
			}
		})
	}
}

func TestStatus_UpdateName(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Original Name", StatusTypeOpen)
	require.NoError(t, err)
	status.ClearEvents()

	t.Run("successful name update", func(t *testing.T) {
		err := status.UpdateName("New Name")
		require.NoError(t, err)

		assert.Equal(t, "New Name", status.Name())

		events := status.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(StatusUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "name", event.Field)
		assert.Equal(t, "Original Name", event.OldValue)
		assert.Equal(t, "New Name", event.NewValue)
	})

	t.Run("empty name", func(t *testing.T) {
		err := status.UpdateName("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})
}

func TestStatus_UpdateDescription(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Status", StatusTypeOpen)
	require.NoError(t, err)
	status.ClearEvents()

	status.UpdateDescription("New description")
	assert.Equal(t, "New description", status.Description())

	events := status.DomainEvents()
	require.Len(t, events, 1)
	event, ok := events[0].(StatusUpdatedEvent)
	require.True(t, ok)
	assert.Equal(t, "description", event.Field)
}

func TestStatus_UpdateColor(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Status", StatusTypeOpen)
	require.NoError(t, err)
	status.ClearEvents()

	status.UpdateColor("#FF5733")
	assert.Equal(t, "#FF5733", status.Color())

	events := status.DomainEvents()
	require.Len(t, events, 1)
	event, ok := events[0].(StatusUpdatedEvent)
	require.True(t, ok)
	assert.Equal(t, "color", event.Field)
	assert.Equal(t, "#FF5733", event.NewValue)
}

func TestStatus_UpdatePosition(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Status", StatusTypeOpen)
	require.NoError(t, err)
	status.ClearEvents()

	status.UpdatePosition(5)
	assert.Equal(t, 5, status.Position())

	events := status.DomainEvents()
	require.Len(t, events, 1)
	event, ok := events[0].(StatusUpdatedEvent)
	require.True(t, ok)
	assert.Equal(t, "position", event.Field)
	assert.Equal(t, 0, event.OldValue)
	assert.Equal(t, 5, event.NewValue)
}

func TestStatus_UpdateType(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Status", StatusTypeOpen)
	require.NoError(t, err)
	status.ClearEvents()

	t.Run("update to active type", func(t *testing.T) {
		err := status.UpdateType(StatusTypeActive)
		require.NoError(t, err)

		assert.Equal(t, StatusTypeActive, status.StatusType())

		events := status.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(StatusUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "status_type", event.Field)
		assert.Equal(t, string(StatusTypeOpen), event.OldValue)
		assert.Equal(t, string(StatusTypeActive), event.NewValue)
	})

	t.Run("empty status type", func(t *testing.T) {
		status.ClearEvents()
		err := status.UpdateType("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "statusType cannot be empty")
	})
}

func TestStatus_ActivateDeactivate(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Status", StatusTypeOpen)
	require.NoError(t, err)
	status.ClearEvents()

	t.Run("deactivate active status", func(t *testing.T) {
		assert.True(t, status.IsActiveStatus())

		status.Deactivate()
		assert.False(t, status.IsActiveStatus())

		events := status.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(StatusDeactivatedEvent)
		require.True(t, ok)
	})

	t.Run("deactivate already inactive status", func(t *testing.T) {
		status.ClearEvents()
		status.Deactivate()

		// Should not generate duplicate event
		events := status.DomainEvents()
		assert.Len(t, events, 0)
	})

	t.Run("activate inactive status", func(t *testing.T) {
		status.ClearEvents()
		status.Activate()
		assert.True(t, status.IsActiveStatus())

		events := status.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(StatusActivatedEvent)
		require.True(t, ok)
	})

	t.Run("activate already active status", func(t *testing.T) {
		status.ClearEvents()
		status.Activate()

		// Should not generate duplicate event
		events := status.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestStatus_TypeCheckers(t *testing.T) {
	pipelineID := uuid.New()

	t.Run("IsOpen", func(t *testing.T) {
		status, err := NewStatus(pipelineID, "New", StatusTypeOpen)
		require.NoError(t, err)

		assert.True(t, status.IsOpen())
		assert.False(t, status.IsActiveType())
		assert.False(t, status.IsClosed())
	})

	t.Run("IsActiveType", func(t *testing.T) {
		status, err := NewStatus(pipelineID, "In Progress", StatusTypeActive)
		require.NoError(t, err)

		assert.False(t, status.IsOpen())
		assert.True(t, status.IsActiveType())
		assert.False(t, status.IsClosed())
	})

	t.Run("IsClosed", func(t *testing.T) {
		status, err := NewStatus(pipelineID, "Won", StatusTypeClosed)
		require.NoError(t, err)

		assert.False(t, status.IsOpen())
		assert.False(t, status.IsActiveType())
		assert.True(t, status.IsClosed())
	})
}

func TestReconstructStatus(t *testing.T) {
	id := uuid.New()
	pipelineID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	t.Run("reconstruct with full data", func(t *testing.T) {
		status := ReconstructStatus(
			id, pipelineID,
			"Reconstructed Status",
			"Test description",
			"#FF0000",
			StatusTypeActive,
			3,
			true,
			createdAt, updatedAt,
		)

		assert.Equal(t, id, status.ID())
		assert.Equal(t, pipelineID, status.PipelineID())
		assert.Equal(t, "Reconstructed Status", status.Name())
		assert.Equal(t, "Test description", status.Description())
		assert.Equal(t, "#FF0000", status.Color())
		assert.Equal(t, StatusTypeActive, status.StatusType())
		assert.Equal(t, 3, status.Position())
		assert.True(t, status.IsActiveStatus())
		assert.Equal(t, createdAt, status.CreatedAt())
		assert.Equal(t, updatedAt, status.UpdatedAt())
		assert.Len(t, status.DomainEvents(), 0) // No events on reconstruction
	})

	t.Run("reconstruct inactive status", func(t *testing.T) {
		status := ReconstructStatus(
			id, pipelineID,
			"Inactive Status",
			"",
			"",
			StatusTypeOpen,
			0,
			false, // inactive
			createdAt, updatedAt,
		)

		assert.False(t, status.IsActiveStatus())
	})
}

func TestStatus_EventManagement(t *testing.T) {
	status, err := NewStatus(uuid.New(), "Status", StatusTypeOpen)
	require.NoError(t, err)

	t.Run("clear events", func(t *testing.T) {
		assert.Len(t, status.DomainEvents(), 1) // Creation event

		status.ClearEvents()
		assert.Len(t, status.DomainEvents(), 0)
	})

	t.Run("multiple operations generate multiple events", func(t *testing.T) {
		status.ClearEvents()

		_ = status.UpdateName("New Name")
		status.UpdateDescription("New desc")
		status.UpdateColor("#000000")
		status.UpdatePosition(2)

		events := status.DomainEvents()
		assert.Len(t, events, 4)
	})

	t.Run("events are immutable copies", func(t *testing.T) {
		status.ClearEvents()
		status.UpdateColor("#FFF")

		events1 := status.DomainEvents()
		events2 := status.DomainEvents()

		// Should be different slices (copies)
		assert.NotSame(t, &events1, &events2)
		assert.Equal(t, len(events1), len(events2))
	})
}

func TestStatus_AllGetters(t *testing.T) {
	pipelineID := uuid.New()
	status, err := NewStatus(pipelineID, "Test Status", StatusTypeActive)
	require.NoError(t, err)

	status.UpdateDescription("Test Description")
	status.UpdateColor("#ABCDEF")
	status.UpdatePosition(7)

	t.Run("verify all getters", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, status.ID())
		assert.Equal(t, pipelineID, status.PipelineID())
		assert.Equal(t, "Test Status", status.Name())
		assert.Equal(t, "Test Description", status.Description())
		assert.Equal(t, "#ABCDEF", status.Color())
		assert.Equal(t, StatusTypeActive, status.StatusType())
		assert.Equal(t, 7, status.Position())
		assert.True(t, status.IsActiveStatus())
		assert.NotZero(t, status.CreatedAt())
		assert.NotZero(t, status.UpdatedAt())
	})
}
