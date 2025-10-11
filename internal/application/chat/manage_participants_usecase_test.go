package chat

import (
	"context"
	"errors"
	"testing"

	domainchat "github.com/caloi/ventros-crm/internal/domain/crm/chat"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewManageParticipantsUseCase(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)

	uc := NewManageParticipantsUseCase(repo, eventBus)

	assert.NotNil(t, uc)
	assert.Equal(t, repo, uc.chatRepo)
	assert.Equal(t, eventBus, uc.eventBus)
}

func TestManageParticipantsUseCase_AddParticipant_Success_Contact(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	newParticipantID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.ClearEvents() // Clear creation events

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   newParticipantID,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Len(t, result.Chat.Participants, 2) // Creator + new participant

	// Verify the new participant was added
	found := false
	for _, p := range result.Chat.Participants {
		if p.ID == newParticipantID {
			found = true
			assert.Equal(t, "contact", p.Type)
			assert.False(t, p.IsAdmin)
			break
		}
	}
	assert.True(t, found, "New participant should be in the chat")

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestManageParticipantsUseCase_AddParticipant_Success_Agent(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	agentID := uuid.New()
	tenantID := "tenant-123"

	// Create an individual chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   agentID,
		ParticipantType: "agent",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chat.Participants, 2) // Contact + agent

	// Verify the agent was added
	found := false
	for _, p := range result.Chat.Participants {
		if p.ID == agentID {
			found = true
			assert.Equal(t, "agent", p.Type)
			break
		}
	}
	assert.True(t, found)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestManageParticipantsUseCase_AddParticipant_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	participantID := uuid.New()

	input := AddParticipantInput{
		ChatID:          uuid.Nil,
		ParticipantID:   participantID,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestManageParticipantsUseCase_AddParticipant_MissingParticipantID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   uuid.Nil,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "participant_id is required")
}

func TestManageParticipantsUseCase_AddParticipant_MissingParticipantType(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	participantID := uuid.New()

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   participantID,
		ParticipantType: "",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "participant_type is required")
}

func TestManageParticipantsUseCase_AddParticipant_InvalidParticipantType(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	participantID := uuid.New()

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   participantID,
		ParticipantType: "invalid",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid participant_type: invalid")
}

func TestManageParticipantsUseCase_AddParticipant_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	participantID := uuid.New()

	// Mock repository returning not found
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   participantID,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_AddParticipant_ParticipantAlreadyExists(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   creatorID, // Try to add creator again
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to add participant")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_AddParticipant_ClosedChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	newParticipantID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat and close it
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.Close()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   newParticipantID,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to add participant")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_AddParticipant_RepositoryUpdateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	newParticipantID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repoError := errors.New("database connection failed")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   newParticipantID,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_AddParticipant_EventBusError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	newParticipantID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event bus error (should not fail the operation)
	eventBusError := errors.New("event bus unavailable")
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(eventBusError)

	input := AddParticipantInput{
		ChatID:          chatID,
		ParticipantID:   newParticipantID,
		ParticipantType: "contact",
	}

	result, err := uc.AddParticipant(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestManageParticipantsUseCase_RemoveParticipant_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	participantToRemove := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat and add a participant
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.AddParticipant(participantToRemove, domainchat.ParticipantTypeContact)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: participantToRemove,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chat.Participants, 1) // Only creator remains

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestManageParticipantsUseCase_RemoveParticipant_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	participantID := uuid.New()

	input := RemoveParticipantInput{
		ChatID:        uuid.Nil,
		ParticipantID: participantID,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestManageParticipantsUseCase_RemoveParticipant_MissingParticipantID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: uuid.Nil,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "participant_id is required")
}

func TestManageParticipantsUseCase_RemoveParticipant_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	participantID := uuid.New()

	// Mock repository returning not found
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: participantID,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_RemoveParticipant_FromIndividualChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an individual chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: contactID,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to remove participant")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_RemoveParticipant_ParticipantNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	nonExistentParticipant := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: nonExistentParticipant,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to remove participant")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_RemoveParticipant_RepositoryUpdateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	participantToRemove := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat and add a participant
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.AddParticipant(participantToRemove, domainchat.ParticipantTypeContact)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repoError := errors.New("database connection failed")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: participantToRemove,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestManageParticipantsUseCase_RemoveParticipant_EventBusError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewManageParticipantsUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	participantToRemove := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create a group chat and add a participant
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.AddParticipant(participantToRemove, domainchat.ParticipantTypeContact)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event bus error (should not fail the operation)
	eventBusError := errors.New("event bus unavailable")
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(eventBusError)

	input := RemoveParticipantInput{
		ChatID:        chatID,
		ParticipantID: participantToRemove,
	}

	result, err := uc.RemoveParticipant(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
