package chat

import (
	"context"
	"errors"
	"testing"

	domainchat "github.com/ventros/crm/internal/domain/crm/chat"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewArchiveChatUseCase(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)

	uc := NewArchiveChatUseCase(repo, eventBus)

	assert.NotNil(t, uc)
	assert.Equal(t, repo, uc.chatRepo)
	assert.Equal(t, eventBus, uc.eventBus)
}

func TestArchiveChatUseCase_Archive_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := ArchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Archive(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, "archived", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Archive_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	input := ArchiveChatInput{
		ChatID: uuid.Nil,
	}

	result, err := uc.Archive(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestArchiveChatUseCase_Archive_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()

	// Mock repository returning not found
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := ArchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Archive(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestArchiveChatUseCase_Archive_RepositoryUpdateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repoError := errors.New("database connection failed")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	input := ArchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Archive(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestArchiveChatUseCase_Archive_EventBusError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
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

	input := ArchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Archive(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "archived", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Unarchive_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create a chat and archive it
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.Archive()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UnarchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Unarchive(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, "active", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Unarchive_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	input := UnarchiveChatInput{
		ChatID: uuid.Nil,
	}

	result, err := uc.Unarchive(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestArchiveChatUseCase_Unarchive_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()

	// Mock repository returning not found
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := UnarchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Unarchive(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestArchiveChatUseCase_Unarchive_RepositoryUpdateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create a chat and archive it
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.Archive()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repoError := errors.New("database connection failed")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	input := UnarchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Unarchive(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestArchiveChatUseCase_Unarchive_EventBusError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create a chat and archive it
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.Archive()
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

	input := UnarchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Unarchive(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "active", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Close_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := CloseChatInput{
		ChatID: chatID,
	}

	result, err := uc.Close(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, "closed", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Close_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	input := CloseChatInput{
		ChatID: uuid.Nil,
	}

	result, err := uc.Close(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestArchiveChatUseCase_Close_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()

	// Mock repository returning not found
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := CloseChatInput{
		ChatID: chatID,
	}

	result, err := uc.Close(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestArchiveChatUseCase_Close_RepositoryUpdateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repoError := errors.New("database connection failed")
	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	input := CloseChatInput{
		ChatID: chatID,
	}

	result, err := uc.Close(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestArchiveChatUseCase_Close_EventBusError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
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

	input := CloseChatInput{
		ChatID: chatID,
	}

	result, err := uc.Close(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "closed", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Close_AlreadyClosed(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create a chat and close it
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.Close()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := CloseChatInput{
		ChatID: chatID,
	}

	result, err := uc.Close(context.Background(), input)

	// Should succeed (idempotent operation)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "closed", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Archive_AlreadyArchived(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create a chat and archive it
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.Archive()
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := ArchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Archive(context.Background(), input)

	// Should succeed (idempotent operation)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "archived", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestArchiveChatUseCase_Unarchive_AlreadyActive(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewArchiveChatUseCase(repo, eventBus)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create an active chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat.ClearEvents()

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	repo.On("Update", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	input := UnarchiveChatInput{
		ChatID: chatID,
	}

	result, err := uc.Unarchive(context.Background(), input)

	// Should succeed (idempotent operation)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "active", result.Chat.Status)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
