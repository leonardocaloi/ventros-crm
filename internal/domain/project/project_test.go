package project

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProject(t *testing.T) {
	userID := uuid.New()
	billingAccountID := uuid.New()

	t.Run("valid project", func(t *testing.T) {
		project, err := NewProject(userID, billingAccountID, "tenant-123", "Test Project")
		require.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Test Project", project.Name())
		assert.True(t, project.IsActive())
	})

	t.Run("validation errors", func(t *testing.T) {
		_, err := NewProject(uuid.Nil, billingAccountID, "tenant", "Project")
		assert.Error(t, err)

		_, err = NewProject(userID, uuid.Nil, "tenant", "Project")
		assert.Error(t, err)

		_, err = NewProject(userID, billingAccountID, "", "Project")
		assert.Error(t, err)

		_, err = NewProject(userID, billingAccountID, "tenant", "")
		assert.Error(t, err)
	})

	t.Run("activate/deactivate", func(t *testing.T) {
		project, _ := NewProject(userID, billingAccountID, "tenant", "Project")

		project.Deactivate()
		assert.False(t, project.IsActive())

		project.Activate()
		assert.True(t, project.IsActive())
	})
}
