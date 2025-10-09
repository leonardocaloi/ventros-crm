package note

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNote(t *testing.T) {
	contactID := uuid.New()
	authorID := uuid.New()

	t.Run("valid note", func(t *testing.T) {
		note, err := NewNote("tenant-123", contactID, authorID, "Test note content")
		require.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test note content", note.Content())
	})

	t.Run("validation", func(t *testing.T) {
		_, err := NewNote("", contactID, authorID, "Content")
		assert.Error(t, err)

		_, err = NewNote("tenant", uuid.Nil, authorID, "Content")
		assert.Error(t, err)

		_, err = NewNote("tenant", contactID, uuid.Nil, "Content")
		assert.Error(t, err)

		_, err = NewNote("tenant", contactID, authorID, "")
		assert.Error(t, err)
	})
}
