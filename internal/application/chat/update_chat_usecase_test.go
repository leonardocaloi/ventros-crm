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

func TestNewUpdateChatUseCase(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)

	uc := NewUpdateChatUseCase(repo, eventBus)

	assert.NotNil(t, uc)
	assert.Equal(t, repo, uc.chatRepo)
	assert.Equal(t, eventBus, uc.eventBus)
}

func TestUpdateChatUseCase_UpdateSubject_Success_GroupChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"
	newSubject := "New Subject"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, newSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_Success_ChannelChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Channel"
	newSubject := "New Channel"

	// Create a channel chat
	chat, _ := domainchat.NewChannelChat(projectID, tenantID, oldSubject)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, newSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	input := UpdateChatSubjectInput{
		ChatID:  uuid.Nil,
		Subject: "New Subject",
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestUpdateChatUseCase_UpdateSubject_MissingSubject(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: "",
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "subject is required")
}

func TestUpdateChatUseCase_UpdateSubject_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()

	// Mock repository returning not found
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: "New Subject",
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_IndividualChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

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

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: "New Subject",
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update subject")

	repo.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_RepositoryUpdateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"
	newSubject := "New Subject"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repoError := errors.New("database connection failed")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_EventBusError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"
	newSubject := "New Subject"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
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

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_SameSubject(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	subject := "Same Subject"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: subject, // Same subject
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, subject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_WithSpecialCharacters(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"
	newSubject := "New Subject with 特殊字符 and émojis 🎉"

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_VeryLongSubject(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"

	// Create a very long subject
	longSubject := ""
	for i := 0; i < 500; i++ {
		longSubject += "A"
	}

	// Create a group chat
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: longSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, longSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_RepositoryFindError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()

	// Mock repository error
	repoError := errors.New("database connection failed")
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, repoError)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: "New Subject",
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_ClosedChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"
	newSubject := "New Subject"

	// Create a group chat and close it
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
	chat.Close()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	// Should succeed - domain doesn't prevent updating closed chat subject
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestUpdateChatUseCase_UpdateSubject_ArchivedChat(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewUpdateChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	oldSubject := "Old Subject"
	newSubject := "New Subject"

	// Create a group chat and archive it
	chat, _ := domainchat.NewGroupChat(projectID, tenantID, oldSubject, creatorID, nil)
	chat.Archive()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UpdateChatSubjectInput{
		ChatID:  chatID,
		Subject: newSubject,
	}

	result, err := uc.UpdateSubject(context.Background(), input)

	// Should succeed - domain doesn't prevent updating archived chat subject
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newSubject, *result.Chat.Subject)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
