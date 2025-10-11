package chat

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Factory Methods

func TestNewIndividualChat(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()

	t.Run("creates valid individual chat", func(t *testing.T) {
		chat, err := NewIndividualChat(projectID, tenantID, contactID)

		require.NoError(t, err)
		assert.NotNil(t, chat)
		assert.NotEqual(t, uuid.Nil, chat.ID())
		assert.Equal(t, projectID, chat.ProjectID())
		assert.Equal(t, tenantID, chat.TenantID())
		assert.Equal(t, ChatTypeIndividual, chat.ChatType())
		assert.Equal(t, ChatStatusActive, chat.Status())
		assert.Nil(t, chat.Subject())
		assert.Nil(t, chat.Description())
		assert.NotNil(t, chat.Metadata())
		assert.Len(t, chat.Metadata(), 0)

		// Check participant
		participants := chat.Participants()
		assert.Len(t, participants, 1)
		assert.Equal(t, contactID, participants[0].ID)
		assert.Equal(t, ParticipantTypeContact, participants[0].Type)
		assert.False(t, participants[0].IsAdmin)
		assert.Nil(t, participants[0].LeftAt)

		// Check event
		events := chat.DomainEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "chat.created", events[0].EventType())

		// Check timestamps
		assert.False(t, chat.CreatedAt().IsZero())
		assert.False(t, chat.UpdatedAt().IsZero())
		assert.Nil(t, chat.LastMessageAt())
	})

	t.Run("returns error when projectID is nil", func(t *testing.T) {
		chat, err := NewIndividualChat(uuid.Nil, tenantID, contactID)

		assert.Error(t, err)
		assert.Equal(t, ErrProjectIDRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when tenantID is empty", func(t *testing.T) {
		chat, err := NewIndividualChat(projectID, "", contactID)

		assert.Error(t, err)
		assert.Equal(t, ErrTenantIDRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when contactID is nil", func(t *testing.T) {
		chat, err := NewIndividualChat(projectID, tenantID, uuid.Nil)

		assert.Error(t, err)
		assert.Equal(t, ErrContactIDRequired, err)
		assert.Nil(t, chat)
	})
}

func TestNewGroupChat(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	subject := "Team Discussion"
	creatorID := uuid.New()

	t.Run("creates valid group chat", func(t *testing.T) {
		chat, err := NewGroupChat(projectID, tenantID, subject, creatorID, nil)

		require.NoError(t, err)
		assert.NotNil(t, chat)
		assert.NotEqual(t, uuid.Nil, chat.ID())
		assert.Equal(t, projectID, chat.ProjectID())
		assert.Equal(t, tenantID, chat.TenantID())
		assert.Equal(t, ChatTypeGroup, chat.ChatType())
		assert.Equal(t, ChatStatusActive, chat.Status())
		assert.NotNil(t, chat.Subject())
		assert.Equal(t, subject, *chat.Subject())

		// Check creator participant
		participants := chat.Participants()
		assert.Len(t, participants, 1)
		assert.Equal(t, creatorID, participants[0].ID)
		assert.Equal(t, ParticipantTypeContact, participants[0].Type)
		assert.True(t, participants[0].IsAdmin) // Creator is admin

		// Check event
		events := chat.DomainEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "chat.created", events[0].EventType())
	})

	t.Run("returns error when projectID is nil", func(t *testing.T) {
		chat, err := NewGroupChat(uuid.Nil, tenantID, subject, creatorID, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrProjectIDRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when tenantID is empty", func(t *testing.T) {
		chat, err := NewGroupChat(projectID, "", subject, creatorID, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrTenantIDRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when subject is empty", func(t *testing.T) {
		chat, err := NewGroupChat(projectID, tenantID, "", creatorID, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrSubjectRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when creatorID is nil", func(t *testing.T) {
		chat, err := NewGroupChat(projectID, tenantID, subject, uuid.Nil, nil)

		assert.Error(t, err)
		assert.Equal(t, ErrCreatorIDRequired, err)
		assert.Nil(t, chat)
	})
}

func TestNewChannelChat(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	subject := "Announcements Channel"

	t.Run("creates valid channel chat", func(t *testing.T) {
		chat, err := NewChannelChat(projectID, tenantID, subject)

		require.NoError(t, err)
		assert.NotNil(t, chat)
		assert.NotEqual(t, uuid.Nil, chat.ID())
		assert.Equal(t, projectID, chat.ProjectID())
		assert.Equal(t, tenantID, chat.TenantID())
		assert.Equal(t, ChatTypeChannel, chat.ChatType())
		assert.Equal(t, ChatStatusActive, chat.Status())
		assert.NotNil(t, chat.Subject())
		assert.Equal(t, subject, *chat.Subject())

		// Channels may have no participants
		participants := chat.Participants()
		assert.Len(t, participants, 0)

		// Check event
		events := chat.DomainEvents()
		assert.Len(t, events, 1)
		assert.Equal(t, "chat.created", events[0].EventType())
	})

	t.Run("returns error when projectID is nil", func(t *testing.T) {
		chat, err := NewChannelChat(uuid.Nil, tenantID, subject)

		assert.Error(t, err)
		assert.Equal(t, ErrProjectIDRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when tenantID is empty", func(t *testing.T) {
		chat, err := NewChannelChat(projectID, "", subject)

		assert.Error(t, err)
		assert.Equal(t, ErrTenantIDRequired, err)
		assert.Nil(t, chat)
	})

	t.Run("returns error when subject is empty", func(t *testing.T) {
		chat, err := NewChannelChat(projectID, tenantID, "")

		assert.Error(t, err)
		assert.Equal(t, ErrSubjectRequired, err)
		assert.Nil(t, chat)
	})
}

// Test ReconstructChat

func TestReconstructChat(t *testing.T) {
	id := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	chatType := ChatTypeGroup
	subject := "Test Group"
	description := "Test Description"
	participants := []Participant{
		{ID: uuid.New(), Type: ParticipantTypeContact, JoinedAt: time.Now(), IsAdmin: true},
		{ID: uuid.New(), Type: ParticipantTypeAgent, JoinedAt: time.Now(), IsAdmin: false},
	}
	status := ChatStatusActive
	metadata := map[string]interface{}{"key": "value"}
	lastMessageAt := time.Now()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	externalID := "test-external-id"
	chat := ReconstructChat(
		id,
		projectID,
		tenantID,
		chatType,
		&externalID,
		&subject,
		&description,
		participants,
		status,
		metadata,
		&lastMessageAt,
		createdAt,
		updatedAt,
	)

	assert.NotNil(t, chat)
	assert.Equal(t, id, chat.ID())
	assert.Equal(t, projectID, chat.ProjectID())
	assert.Equal(t, tenantID, chat.TenantID())
	assert.Equal(t, chatType, chat.ChatType())
	assert.Equal(t, subject, *chat.Subject())
	assert.Equal(t, description, *chat.Description())
	assert.Len(t, chat.Participants(), 2)
	assert.Equal(t, status, chat.Status())
	assert.Len(t, chat.Metadata(), 1)
	assert.NotNil(t, chat.LastMessageAt())
	assert.Equal(t, createdAt, chat.CreatedAt())
	assert.Equal(t, updatedAt, chat.UpdatedAt())
	assert.Len(t, chat.DomainEvents(), 0) // No events on reconstruction
}

func TestReconstructChat_InitializesEmptyMetadata(t *testing.T) {
	chat := ReconstructChat(
		uuid.New(),
		uuid.New(),
		"tenant-123",
		ChatTypeIndividual,
		nil,
		nil,
		nil, // nil description
		[]Participant{},
		ChatStatusActive,
		nil, // nil metadata
		nil,
		time.Now(),
		time.Now(),
	)

	assert.NotNil(t, chat.Metadata())
	assert.Len(t, chat.Metadata(), 0)
}

// Test AddParticipant

func TestChat_AddParticipant(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()

	t.Run("adds contact participant to individual chat", func(t *testing.T) {
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)
		agentID := uuid.New()

		err := chat.AddParticipant(agentID, ParticipantTypeAgent)

		assert.NoError(t, err)
		assert.Len(t, chat.Participants(), 2)
		assert.True(t, chat.IsParticipant(agentID))

		// Check event
		events := chat.DomainEvents()
		assert.Len(t, events, 2) // created + participant_added
		assert.Equal(t, "chat.participant_added", events[1].EventType())
	})

	t.Run("adds multiple participants to group chat", func(t *testing.T) {
		chat, _ := NewGroupChat(projectID, tenantID, "Group", contactID, nil)
		participant1ID := uuid.New()
		participant2ID := uuid.New()

		err1 := chat.AddParticipant(participant1ID, ParticipantTypeContact)
		err2 := chat.AddParticipant(participant2ID, ParticipantTypeAgent)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Len(t, chat.Participants(), 3) // creator + 2 new
	})

	t.Run("returns error when adding duplicate participant", func(t *testing.T) {
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)

		err := chat.AddParticipant(contactID, ParticipantTypeContact)

		assert.Error(t, err)
		assert.Equal(t, ErrParticipantAlreadyExists, err)
		assert.Len(t, chat.Participants(), 1) // unchanged
	})

	t.Run("returns error when adding second contact to individual chat", func(t *testing.T) {
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)
		anotherContactID := uuid.New()

		err := chat.AddParticipant(anotherContactID, ParticipantTypeContact)

		assert.Error(t, err)
		assert.Equal(t, ErrIndividualChatLimitReached, err)
		assert.Len(t, chat.Participants(), 1) // unchanged
	})

	t.Run("returns error when adding to closed chat", func(t *testing.T) {
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)
		chat.Close()
		agentID := uuid.New()

		err := chat.AddParticipant(agentID, ParticipantTypeAgent)

		assert.Error(t, err)
		assert.Equal(t, ErrChatClosed, err)
	})
}

// Test RemoveParticipant

func TestChat_RemoveParticipant(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	creatorID := uuid.New()

	t.Run("removes participant from group chat", func(t *testing.T) {
		chat, _ := NewGroupChat(projectID, tenantID, "Group", creatorID, nil)
		participantID := uuid.New()
		chat.AddParticipant(participantID, ParticipantTypeContact)

		err := chat.RemoveParticipant(participantID)

		assert.NoError(t, err)
		assert.Len(t, chat.Participants(), 1) // only creator remains
		assert.False(t, chat.IsParticipant(participantID))

		// Check event
		events := chat.DomainEvents()
		assert.Equal(t, "chat.participant_removed", events[len(events)-1].EventType())
	})

	t.Run("returns error when removing from individual chat", func(t *testing.T) {
		contactID := uuid.New()
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)

		err := chat.RemoveParticipant(contactID)

		assert.Error(t, err)
		assert.Equal(t, ErrCannotRemoveFromIndividual, err)
		assert.Len(t, chat.Participants(), 1) // unchanged
	})

	t.Run("returns error when participant not found", func(t *testing.T) {
		chat, _ := NewGroupChat(projectID, tenantID, "Group", creatorID, nil)
		nonExistentID := uuid.New()

		err := chat.RemoveParticipant(nonExistentID)

		assert.Error(t, err)
		assert.Equal(t, ErrParticipantNotFound, err)
	})
}

// Test Status Transitions

func TestChat_Archive(t *testing.T) {
	chat, _ := NewIndividualChat(uuid.New(), "tenant-123", uuid.New())
	initialUpdatedAt := chat.UpdatedAt()
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	chat.Archive()

	assert.Equal(t, ChatStatusArchived, chat.Status())
	assert.True(t, chat.UpdatedAt().After(initialUpdatedAt))

	// Check event
	events := chat.DomainEvents()
	assert.Equal(t, "chat.archived", events[len(events)-1].EventType())
}

func TestChat_Unarchive(t *testing.T) {
	chat, _ := NewIndividualChat(uuid.New(), "tenant-123", uuid.New())
	chat.Archive()
	time.Sleep(10 * time.Millisecond)

	chat.Unarchive()

	assert.Equal(t, ChatStatusActive, chat.Status())

	// Check event
	events := chat.DomainEvents()
	assert.Equal(t, "chat.unarchived", events[len(events)-1].EventType())
}

func TestChat_Close(t *testing.T) {
	chat, _ := NewIndividualChat(uuid.New(), "tenant-123", uuid.New())
	initialUpdatedAt := chat.UpdatedAt()
	time.Sleep(10 * time.Millisecond)

	chat.Close()

	assert.Equal(t, ChatStatusClosed, chat.Status())
	assert.True(t, chat.UpdatedAt().After(initialUpdatedAt))

	// Check event
	events := chat.DomainEvents()
	assert.Equal(t, "chat.closed", events[len(events)-1].EventType())
}

// Test UpdateLastMessageAt

func TestChat_UpdateLastMessageAt(t *testing.T) {
	chat, _ := NewIndividualChat(uuid.New(), "tenant-123", uuid.New())
	assert.Nil(t, chat.LastMessageAt())

	messageTime := time.Now()
	chat.UpdateLastMessageAt(messageTime)

	assert.NotNil(t, chat.LastMessageAt())
	assert.Equal(t, messageTime.Unix(), chat.LastMessageAt().Unix())
}

// Test UpdateSubject

func TestChat_UpdateSubject(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	creatorID := uuid.New()

	t.Run("updates subject for group chat", func(t *testing.T) {
		chat, _ := NewGroupChat(projectID, tenantID, "Old Subject", creatorID, nil)
		newSubject := "New Subject"

		err := chat.UpdateSubject(newSubject)

		assert.NoError(t, err)
		assert.Equal(t, newSubject, *chat.Subject())

		// Check event
		events := chat.DomainEvents()
		assert.Equal(t, "chat.subject_updated", events[len(events)-1].EventType())
	})

	t.Run("updates subject for channel chat", func(t *testing.T) {
		chat, _ := NewChannelChat(projectID, tenantID, "Old Channel")
		newSubject := "New Channel"

		err := chat.UpdateSubject(newSubject)

		assert.NoError(t, err)
		assert.Equal(t, newSubject, *chat.Subject())
	})

	t.Run("returns error for individual chat", func(t *testing.T) {
		contactID := uuid.New()
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)

		err := chat.UpdateSubject("Some Subject")

		assert.Error(t, err)
		assert.Equal(t, ErrIndividualChatNoSubject, err)
		assert.Nil(t, chat.Subject())
	})
}

// Test UpdateDescription

func TestChat_UpdateDescription(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	creatorID := uuid.New()

	t.Run("updates description for group chat", func(t *testing.T) {
		chat, _ := NewGroupChat(projectID, tenantID, "Group", creatorID, nil)
		description := "This is a test group"

		err := chat.UpdateDescription(description)

		assert.NoError(t, err)
		assert.NotNil(t, chat.Description())
		assert.Equal(t, description, *chat.Description())

		// Check event
		events := chat.DomainEvents()
		assert.Equal(t, "chat.description_updated", events[len(events)-1].EventType())
	})

	t.Run("returns error for individual chat", func(t *testing.T) {
		contactID := uuid.New()
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)

		err := chat.UpdateDescription("Some description")

		assert.Error(t, err)
		assert.Equal(t, ErrIndividualChatNoSubject, err)
		assert.Nil(t, chat.Description())
	})
}

// Test Query Methods

func TestChat_IsParticipant(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()
	chat, _ := NewIndividualChat(projectID, tenantID, contactID)

	t.Run("returns true for existing participant", func(t *testing.T) {
		assert.True(t, chat.IsParticipant(contactID))
	})

	t.Run("returns false for non-participant", func(t *testing.T) {
		nonParticipantID := uuid.New()
		assert.False(t, chat.IsParticipant(nonParticipantID))
	})
}

func TestChat_GetContactParticipants(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	creatorID := uuid.New()
	chat, _ := NewGroupChat(projectID, tenantID, "Group", creatorID, nil)

	contactID := uuid.New()
	agentID := uuid.New()
	chat.AddParticipant(contactID, ParticipantTypeContact)
	chat.AddParticipant(agentID, ParticipantTypeAgent)

	contacts := chat.GetContactParticipants()

	assert.Len(t, contacts, 2) // creator + added contact
	for _, p := range contacts {
		assert.Equal(t, ParticipantTypeContact, p.Type)
	}
}

func TestChat_GetAgentParticipants(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()
	chat, _ := NewIndividualChat(projectID, tenantID, contactID)

	agent1ID := uuid.New()
	agent2ID := uuid.New()
	chat.AddParticipant(agent1ID, ParticipantTypeAgent)
	chat.AddParticipant(agent2ID, ParticipantTypeAgent)

	agents := chat.GetAgentParticipants()

	assert.Len(t, agents, 2)
	for _, p := range agents {
		assert.Equal(t, ParticipantTypeAgent, p.Type)
	}
}

// Test Immutability

func TestChat_Participants_ReturnsImmutableCopy(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()
	chat, _ := NewIndividualChat(projectID, tenantID, contactID)

	participants1 := chat.Participants()
	participants2 := chat.Participants()

	// Modifying returned slice should not affect internal state
	participants1[0].IsAdmin = true
	participants2[0].IsAdmin = false

	// Original chat participants should remain unchanged
	originalParticipants := chat.Participants()
	assert.False(t, originalParticipants[0].IsAdmin)
}

func TestChat_Metadata_ReturnsImmutableCopy(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()
	chat, _ := NewIndividualChat(projectID, tenantID, contactID)

	metadata := chat.Metadata()
	metadata["test_key"] = "test_value"

	// Original chat metadata should remain unchanged
	originalMetadata := chat.Metadata()
	assert.Len(t, originalMetadata, 0)
	_, exists := originalMetadata["test_key"]
	assert.False(t, exists)
}

// Test Domain Events

func TestChat_DomainEvents(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	contactID := uuid.New()

	t.Run("accumulates multiple events", func(t *testing.T) {
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)
		agentID := uuid.New()

		chat.AddParticipant(agentID, ParticipantTypeAgent)
		chat.Archive()
		chat.Unarchive()
		chat.Close()

		events := chat.DomainEvents()
		assert.Len(t, events, 5) // created + participant_added + archived + unarchived + closed

		assert.Equal(t, "chat.created", events[0].EventType())
		assert.Equal(t, "chat.participant_added", events[1].EventType())
		assert.Equal(t, "chat.archived", events[2].EventType())
		assert.Equal(t, "chat.unarchived", events[3].EventType())
		assert.Equal(t, "chat.closed", events[4].EventType())
	})

	t.Run("clears events", func(t *testing.T) {
		chat, _ := NewIndividualChat(projectID, tenantID, contactID)
		assert.Len(t, chat.DomainEvents(), 1)

		chat.ClearEvents()

		assert.Len(t, chat.DomainEvents(), 0)
	})
}

// Test Edge Cases

func TestChat_ComplexScenario(t *testing.T) {
	projectID := uuid.New()
	tenantID := "tenant-123"
	creatorID := uuid.New()

	// Create group chat
	chat, err := NewGroupChat(projectID, tenantID, "Project Team", creatorID, nil)
	require.NoError(t, err)

	// Add multiple participants
	member1 := uuid.New()
	member2 := uuid.New()
	agent := uuid.New()

	chat.AddParticipant(member1, ParticipantTypeContact)
	chat.AddParticipant(member2, ParticipantTypeContact)
	chat.AddParticipant(agent, ParticipantTypeAgent)

	assert.Len(t, chat.Participants(), 4) // creator + 3 added

	// Update subject and description
	chat.UpdateSubject("Updated Project Team")
	chat.UpdateDescription("Team for project X")

	assert.Equal(t, "Updated Project Team", *chat.Subject())
	assert.Equal(t, "Team for project X", *chat.Description())

	// Remove a member
	chat.RemoveParticipant(member1)
	assert.Len(t, chat.Participants(), 3)
	assert.False(t, chat.IsParticipant(member1))

	// Update last message time
	messageTime := time.Now()
	chat.UpdateLastMessageAt(messageTime)
	assert.NotNil(t, chat.LastMessageAt())

	// Archive and unarchive
	chat.Archive()
	assert.Equal(t, ChatStatusArchived, chat.Status())

	chat.Unarchive()
	assert.Equal(t, ChatStatusActive, chat.Status())

	// Close chat
	chat.Close()
	assert.Equal(t, ChatStatusClosed, chat.Status())

	// Cannot add participants to closed chat
	newMember := uuid.New()
	err = chat.AddParticipant(newMember, ParticipantTypeContact)
	assert.Error(t, err)
	assert.Equal(t, ErrChatClosed, err)

	// Verify all events
	events := chat.DomainEvents()
	assert.NotEmpty(t, events)
	// created + 3 participants_added + subject_updated + description_updated +
	// participant_removed + archived + unarchived + closed = 10 events
	assert.Len(t, events, 10)
}
