package contact

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContact_Comprehensive(t *testing.T) {
	projectID := uuid.New()

	t.Run("NewContact validation", func(t *testing.T) {
		contact, err := NewContact(projectID, "tenant-123", "John Doe")
		require.NoError(t, err)
		assert.Equal(t, "John Doe", contact.Name())
		assert.Equal(t, "en", contact.Language())
	})

	t.Run("SetEmail", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		err := contact.SetEmail("john@example.com")
		require.NoError(t, err)
		assert.NotNil(t, contact.Email())
	})

	t.Run("SetPhone", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		err := contact.SetPhone("+5511999999999")
		require.NoError(t, err)
		assert.NotNil(t, contact.Phone())
	})

	t.Run("Tags management", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		contact.AddTag("vip")
		contact.AddTag("customer")
		assert.Len(t, contact.Tags(), 2)

		contact.RemoveTag("vip")
		assert.Len(t, contact.Tags(), 1)

		contact.ClearTags()
		assert.Len(t, contact.Tags(), 0)
	})

	t.Run("Soft delete", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		assert.False(t, contact.IsDeleted())

		err := contact.SoftDelete()
		require.NoError(t, err)
		assert.True(t, contact.IsDeleted())

		// Cannot delete twice
		err = contact.SoftDelete()
		require.Error(t, err)
	})

	t.Run("RecordInteraction", func(t *testing.T) {
		contact, _ := NewContact(projectID, "tenant-123", "John")
		assert.Nil(t, contact.FirstInteractionAt())

		contact.RecordInteraction()
		assert.NotNil(t, contact.FirstInteractionAt())
		assert.NotNil(t, contact.LastInteractionAt())
	})
}
