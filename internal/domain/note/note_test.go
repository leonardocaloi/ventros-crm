package note

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNote(t *testing.T) {
	contactID := uuid.New()
	authorID := uuid.New()

	t.Run("valid note", func(t *testing.T) {
		note, err := NewNote(
			contactID,
			"tenant-123",
			authorID,
			AuthorTypeUser,
			"John Doe",
			"Test note content",
			NoteTypeGeneral,
		)
		require.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, "Test note content", note.Content())
		assert.Equal(t, contactID, note.ContactID())
		assert.Equal(t, "tenant-123", note.TenantID())
		assert.Equal(t, authorID, note.AuthorID())
		assert.Equal(t, "John Doe", note.AuthorName())
		assert.Equal(t, NoteTypeGeneral, note.NoteType())
		assert.Equal(t, PriorityNormal, note.Priority())
		assert.False(t, note.VisibleToClient())
		assert.False(t, note.Pinned())
		assert.Empty(t, note.Tags())
		assert.Empty(t, note.Mentions())
		assert.Empty(t, note.Attachments())
		assert.NotZero(t, note.CreatedAt())
		assert.NotZero(t, note.UpdatedAt())
		assert.Nil(t, note.DeletedAt())
		assert.Nil(t, note.SessionID())

		// Check domain event
		events := note.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(NoteAddedEvent)
		require.True(t, ok)
		assert.Equal(t, note.ID(), event.NoteID)
		assert.Equal(t, contactID, event.ContactID)
		assert.Equal(t, "note.added", event.EventType())
		assert.Equal(t, note.ID(), event.AggregateID())
	})

	t.Run("validation - invalid contact ID", func(t *testing.T) {
		_, err := NewNote(uuid.Nil, "tenant", authorID, AuthorTypeUser, "John", "Content", NoteTypeGeneral)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidContact)
	})

	t.Run("validation - invalid author ID", func(t *testing.T) {
		_, err := NewNote(contactID, "tenant", uuid.Nil, AuthorTypeUser, "John", "Content", NoteTypeGeneral)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAuthor)
	})

	t.Run("validation - empty content", func(t *testing.T) {
		_, err := NewNote(contactID, "tenant", authorID, AuthorTypeUser, "John", "", NoteTypeGeneral)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyContent)
	})

	t.Run("different author types", func(t *testing.T) {
		authorTypes := []AuthorType{AuthorTypeAgent, AuthorTypeSystem, AuthorTypeUser}
		for _, authorType := range authorTypes {
			note, err := NewNote(contactID, "tenant", authorID, authorType, "Author", "Content", NoteTypeGeneral)
			require.NoError(t, err)
			assert.Equal(t, authorType, note.AuthorType())
		}
	})

	t.Run("different note types", func(t *testing.T) {
		noteTypes := []NoteType{
			NoteTypeGeneral, NoteTypeFollowUp, NoteTypeComplaint, NoteTypeResolution,
			NoteTypeEscalation, NoteTypeInternal, NoteTypeCustomer, NoteTypeSessionSummary,
			NoteTypeSessionHandoff, NoteTypeSessionFeedback, NoteTypeAdConversion,
			NoteTypeAdCampaign, NoteTypeAdAttribution, NoteTypeTrackingInsight,
		}
		for _, noteType := range noteTypes {
			note, err := NewNote(contactID, "tenant", authorID, AuthorTypeUser, "Author", "Content", noteType)
			require.NoError(t, err)
			assert.Equal(t, noteType, note.NoteType())
		}
	})
}

func TestReconstructNote(t *testing.T) {
	id := uuid.New()
	contactID := uuid.New()
	sessionID := uuid.New()
	authorID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now().Add(-1 * time.Hour)
	deletedAt := time.Now()

	t.Run("reconstruct with all fields", func(t *testing.T) {
		note := ReconstructNote(
			id, contactID, &sessionID, "tenant-123",
			authorID, AuthorTypeAgent, "Agent Name",
			"Reconstructed content", NoteTypeFollowUp,
			PriorityHigh, true, true,
			[]string{"tag1", "tag2"},
			[]uuid.UUID{uuid.New()},
			[]string{"https://example.com/file.pdf"},
			createdAt, updatedAt, &deletedAt,
		)

		assert.Equal(t, id, note.ID())
		assert.Equal(t, contactID, note.ContactID())
		assert.Equal(t, &sessionID, note.SessionID())
		assert.Equal(t, "tenant-123", note.TenantID())
		assert.Equal(t, authorID, note.AuthorID())
		assert.Equal(t, AuthorTypeAgent, note.AuthorType())
		assert.Equal(t, "Agent Name", note.AuthorName())
		assert.Equal(t, "Reconstructed content", note.Content())
		assert.Equal(t, NoteTypeFollowUp, note.NoteType())
		assert.Equal(t, PriorityHigh, note.Priority())
		assert.True(t, note.VisibleToClient())
		assert.True(t, note.Pinned())
		assert.Len(t, note.Tags(), 2)
		assert.Len(t, note.Mentions(), 1)
		assert.Len(t, note.Attachments(), 1)
		assert.Equal(t, createdAt, note.CreatedAt())
		assert.Equal(t, updatedAt, note.UpdatedAt())
		assert.Equal(t, &deletedAt, note.DeletedAt())
		assert.Empty(t, note.DomainEvents()) // No events on reconstruction
	})

	t.Run("reconstruct with nil collections", func(t *testing.T) {
		note := ReconstructNote(
			id, contactID, nil, "tenant-123",
			authorID, AuthorTypeAgent, "Agent",
			"Content", NoteTypeGeneral,
			PriorityNormal, false, false,
			nil, nil, nil,
			createdAt, updatedAt, nil,
		)

		assert.Empty(t, note.Tags())
		assert.Empty(t, note.Mentions())
		assert.Empty(t, note.Attachments())
		assert.Nil(t, note.SessionID())
		assert.Nil(t, note.DeletedAt())
	})
}

func TestNote_AttachToSession(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)
	note.ClearEvents()

	sessionID := uuid.New()
	note.AttachToSession(sessionID)

	assert.NotNil(t, note.SessionID())
	assert.Equal(t, sessionID, *note.SessionID())
}

func TestNote_UpdateContent(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Original content", NoteTypeGeneral)
	require.NoError(t, err)
	note.ClearEvents()

	updatedBy := uuid.New()

	t.Run("successful update", func(t *testing.T) {
		err := note.UpdateContent("Updated content", updatedBy)
		require.NoError(t, err)
		assert.Equal(t, "Updated content", note.Content())

		events := note.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(NoteUpdatedEvent)
		require.True(t, ok)
		assert.Equal(t, "Original content", event.OldContent)
		assert.Equal(t, "Updated content", event.NewContent)
		assert.Equal(t, updatedBy, event.UpdatedBy)
		assert.Equal(t, "note.updated", event.EventType())
		assert.Equal(t, note.ID(), event.AggregateID())
	})

	t.Run("empty content", func(t *testing.T) {
		err := note.UpdateContent("", updatedBy)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyContent)
	})
}

func TestNote_SetPriority(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	priorities := []Priority{PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent}
	for _, priority := range priorities {
		note.SetPriority(priority)
		assert.Equal(t, priority, note.Priority())
	}
}

func TestNote_SetVisibility(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	assert.False(t, note.VisibleToClient())

	note.SetVisibility(true)
	assert.True(t, note.VisibleToClient())

	note.SetVisibility(false)
	assert.False(t, note.VisibleToClient())
}

func TestNote_PinUnpin(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)
	note.ClearEvents()

	pinnedBy := uuid.New()

	t.Run("pin note", func(t *testing.T) {
		assert.False(t, note.Pinned())

		note.Pin(pinnedBy)
		assert.True(t, note.Pinned())

		events := note.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(NotePinnedEvent)
		require.True(t, ok)
		assert.Equal(t, note.ID(), event.NoteID)
		assert.Equal(t, pinnedBy, event.PinnedBy)
		assert.Equal(t, "note.pinned", event.EventType())
		assert.Equal(t, note.ID(), event.AggregateID())
	})

	t.Run("pin already pinned note", func(t *testing.T) {
		note.ClearEvents()
		note.Pin(pinnedBy)

		// Should not generate duplicate event
		events := note.DomainEvents()
		assert.Len(t, events, 0)
	})

	t.Run("unpin note", func(t *testing.T) {
		note.Unpin()
		assert.False(t, note.Pinned())
	})
}

func TestNote_Tags(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	t.Run("add tags", func(t *testing.T) {
		note.AddTag("urgent")
		note.AddTag("customer-complaint")

		tags := note.Tags()
		assert.Len(t, tags, 2)
		assert.Contains(t, tags, "urgent")
		assert.Contains(t, tags, "customer-complaint")
	})

	t.Run("add empty tag", func(t *testing.T) {
		initialLen := len(note.Tags())
		note.AddTag("")
		assert.Len(t, note.Tags(), initialLen) // Should not add empty tag
	})

	t.Run("remove tag", func(t *testing.T) {
		note.RemoveTag("urgent")
		tags := note.Tags()
		assert.Len(t, tags, 1)
		assert.NotContains(t, tags, "urgent")
	})

	t.Run("remove non-existent tag", func(t *testing.T) {
		initialLen := len(note.Tags())
		note.RemoveTag("non-existent")
		assert.Len(t, note.Tags(), initialLen)
	})
}

func TestNote_Mentions(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	t.Run("mention agents", func(t *testing.T) {
		agent1 := uuid.New()
		agent2 := uuid.New()

		note.MentionAgent(agent1)
		note.MentionAgent(agent2)

		mentions := note.Mentions()
		assert.Len(t, mentions, 2)
		assert.Contains(t, mentions, agent1)
		assert.Contains(t, mentions, agent2)
	})

	t.Run("mention nil agent", func(t *testing.T) {
		initialLen := len(note.Mentions())
		note.MentionAgent(uuid.Nil)
		assert.Len(t, note.Mentions(), initialLen) // Should not add nil
	})
}

func TestNote_Attachments(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	t.Run("add attachments", func(t *testing.T) {
		note.AddAttachment("https://example.com/file1.pdf")
		note.AddAttachment("https://example.com/image.png")

		attachments := note.Attachments()
		assert.Len(t, attachments, 2)
		assert.Contains(t, attachments, "https://example.com/file1.pdf")
		assert.Contains(t, attachments, "https://example.com/image.png")
	})

	t.Run("add empty attachment", func(t *testing.T) {
		initialLen := len(note.Attachments())
		note.AddAttachment("")
		assert.Len(t, note.Attachments(), initialLen) // Should not add empty
	})
}

func TestNote_Delete(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)
	note.ClearEvents()

	deletedBy := uuid.New()

	t.Run("delete note", func(t *testing.T) {
		assert.False(t, note.IsDeleted())
		assert.Nil(t, note.DeletedAt())

		note.Delete(deletedBy)

		assert.True(t, note.IsDeleted())
		assert.NotNil(t, note.DeletedAt())

		events := note.DomainEvents()
		require.Len(t, events, 1)
		event, ok := events[0].(NoteDeletedEvent)
		require.True(t, ok)
		assert.Equal(t, note.ID(), event.NoteID)
		assert.Equal(t, deletedBy, event.DeletedBy)
		assert.Equal(t, "note.deleted", event.EventType())
		assert.Equal(t, note.ID(), event.AggregateID())
	})

	t.Run("delete already deleted note", func(t *testing.T) {
		note.ClearEvents()
		oldDeletedAt := note.DeletedAt()

		note.Delete(deletedBy)

		// Should not change deletedAt or generate event
		assert.Equal(t, oldDeletedAt, note.DeletedAt())
		events := note.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestNote_EventManagement(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	t.Run("clear events", func(t *testing.T) {
		assert.Len(t, note.DomainEvents(), 1) // Creation event

		note.ClearEvents()
		assert.Len(t, note.DomainEvents(), 0)
	})

	t.Run("events are immutable copies", func(t *testing.T) {
		note.UpdateContent("New content", uuid.New())

		events1 := note.DomainEvents()
		events2 := note.DomainEvents()

		// Should be different slices (copies)
		assert.NotSame(t, &events1, &events2)
		assert.Equal(t, len(events1), len(events2))
	})
}

func TestNote_GettersCopies(t *testing.T) {
	note, err := NewNote(uuid.New(), "tenant", uuid.New(), AuthorTypeUser, "User", "Content", NoteTypeGeneral)
	require.NoError(t, err)

	note.AddTag("tag1")
	note.MentionAgent(uuid.New())
	note.AddAttachment("file.pdf")

	t.Run("tags returns copy", func(t *testing.T) {
		tags1 := note.Tags()
		tags2 := note.Tags()

		// Should be different slices
		assert.NotSame(t, &tags1, &tags2)
		assert.Equal(t, tags1, tags2)
	})

	t.Run("mentions returns copy", func(t *testing.T) {
		mentions1 := note.Mentions()
		mentions2 := note.Mentions()

		assert.NotSame(t, &mentions1, &mentions2)
		assert.Equal(t, mentions1, mentions2)
	})

	t.Run("attachments returns copy", func(t *testing.T) {
		attachments1 := note.Attachments()
		attachments2 := note.Attachments()

		assert.NotSame(t, &attachments1, &attachments2)
		assert.Equal(t, attachments1, attachments2)
	})
}

func TestNote_Constants(t *testing.T) {
	// Author types
	assert.Equal(t, AuthorType("agent"), AuthorTypeAgent)
	assert.Equal(t, AuthorType("system"), AuthorTypeSystem)
	assert.Equal(t, AuthorType("user"), AuthorTypeUser)

	// Note types
	assert.Equal(t, NoteType("general"), NoteTypeGeneral)
	assert.Equal(t, NoteType("follow_up"), NoteTypeFollowUp)
	assert.Equal(t, NoteType("complaint"), NoteTypeComplaint)

	// Priorities
	assert.Equal(t, Priority("low"), PriorityLow)
	assert.Equal(t, Priority("normal"), PriorityNormal)
	assert.Equal(t, Priority("high"), PriorityHigh)
	assert.Equal(t, Priority("urgent"), PriorityUrgent)
}
