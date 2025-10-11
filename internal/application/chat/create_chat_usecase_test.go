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

func TestNewCreateChatUseCase(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)

	uc := NewCreateChatUseCase(repo, eventBus)

	assert.NotNil(t, uc)
	assert.Equal(t, repo, uc.chatRepo)
	assert.Equal(t, eventBus, uc.eventBus)
}

func TestCreateChatUseCase_Execute_IndividualChat_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  tenantID,
		ChatType:  "individual",
		ContactID: &contactID,
	}

	// Mock that no existing chat is found
	repo.On("FindIndividualByContact", mock.Anything, contactID, projectID).
		Return(nil, domainchat.ErrChatNotFound)

	// Mock successful creation
	repo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	result, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, projectID, result.Chat.ProjectID)
	assert.Equal(t, tenantID, result.Chat.TenantID)
	assert.Equal(t, "individual", result.Chat.ChatType)
	assert.Equal(t, "active", result.Chat.Status)
	assert.Len(t, result.Chat.Participants, 1)
	assert.Equal(t, contactID, result.Chat.Participants[0].ID)
	assert.Equal(t, "contact", result.Chat.Participants[0].Type)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateChatUseCase_Execute_IndividualChat_AlreadyExists(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  tenantID,
		ChatType:  "individual",
		ContactID: &contactID,
	}

	// Create existing chat
	existingChat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)

	// Mock that existing chat is found
	repo.On("FindIndividualByContact", mock.Anything, contactID, projectID).
		Return(existingChat, nil)

	result, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, existingChat.ID(), result.Chat.ID)

	repo.AssertExpectations(t)
	// Should not call Create or Publish since chat already exists
	repo.AssertNotCalled(t, "Create")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateChatUseCase_Execute_IndividualChat_MissingProjectID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	contactID := uuid.New()

	input := CreateChatInput{
		ProjectID: uuid.Nil,
		TenantID:  "tenant-123",
		ChatType:  "individual",
		ContactID: &contactID,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "project_id is required")
}

func TestCreateChatUseCase_Execute_IndividualChat_MissingTenantID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "",
		ChatType:  "individual",
		ContactID: &contactID,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tenant_id is required")
}

func TestCreateChatUseCase_Execute_IndividualChat_MissingChatType(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "",
		ContactID: &contactID,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_type is required")
}

func TestCreateChatUseCase_Execute_IndividualChat_MissingContactID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "individual",
		ContactID: nil,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "contact_id is required for individual chats")
}

func TestCreateChatUseCase_Execute_IndividualChat_NilContactID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	nilContactID := uuid.Nil

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "individual",
		ContactID: &nilContactID,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "contact_id is required for individual chats")
}

func TestCreateChatUseCase_Execute_GroupChat_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"
	externalID := "group@g.us"

	input := CreateChatInput{
		ProjectID:  projectID,
		TenantID:   tenantID,
		ChatType:   "group",
		CreatorID:  &creatorID,
		Subject:    &subject,
		ExternalID: &externalID,
	}

	// Mock successful creation
	repo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	result, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, projectID, result.Chat.ProjectID)
	assert.Equal(t, tenantID, result.Chat.TenantID)
	assert.Equal(t, "group", result.Chat.ChatType)
	assert.Equal(t, "active", result.Chat.Status)
	assert.Equal(t, subject, *result.Chat.Subject)
	assert.Equal(t, externalID, *result.Chat.ExternalID)
	assert.Len(t, result.Chat.Participants, 1)
	assert.Equal(t, creatorID, result.Chat.Participants[0].ID)
	assert.Equal(t, "contact", result.Chat.Participants[0].Type)
	assert.True(t, result.Chat.Participants[0].IsAdmin)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateChatUseCase_Execute_GroupChat_MissingSubject(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	creatorID := uuid.New()

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "group",
		CreatorID: &creatorID,
		Subject:   nil,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "subject is required for group chats")
}

func TestCreateChatUseCase_Execute_GroupChat_EmptySubject(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	creatorID := uuid.New()
	emptySubject := ""

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "group",
		CreatorID: &creatorID,
		Subject:   &emptySubject,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "subject is required for group chats")
}

func TestCreateChatUseCase_Execute_GroupChat_MissingCreatorID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	subject := "Test Group"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "group",
		CreatorID: nil,
		Subject:   &subject,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "creator_id is required for group chats")
}

func TestCreateChatUseCase_Execute_GroupChat_NilCreatorID(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	subject := "Test Group"
	nilCreatorID := uuid.Nil

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "group",
		CreatorID: &nilCreatorID,
		Subject:   &subject,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "creator_id is required for group chats")
}

func TestCreateChatUseCase_Execute_ChannelChat_Success(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Channel"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  tenantID,
		ChatType:  "channel",
		Subject:   &subject,
	}

	// Mock successful creation
	repo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event publishing
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	result, err := uc.Execute(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, projectID, result.Chat.ProjectID)
	assert.Equal(t, tenantID, result.Chat.TenantID)
	assert.Equal(t, "channel", result.Chat.ChatType)
	assert.Equal(t, "active", result.Chat.Status)
	assert.Equal(t, subject, *result.Chat.Subject)
	assert.Len(t, result.Chat.Participants, 0) // Channels have no initial participants

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateChatUseCase_Execute_ChannelChat_MissingSubject(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "channel",
		Subject:   nil,
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "subject is required for channel chats")
}

func TestCreateChatUseCase_Execute_InvalidChatType(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  "tenant-123",
		ChatType:  "invalid",
	}

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid chat_type: invalid")
}

func TestCreateChatUseCase_Execute_RepositoryCreateError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  tenantID,
		ChatType:  "individual",
		ContactID: &contactID,
	}

	// Mock that no existing chat is found
	repo.On("FindIndividualByContact", mock.Anything, contactID, projectID).
		Return(nil, domainchat.ErrChatNotFound)

	// Mock repository error
	repoError := errors.New("database connection failed")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(repoError)

	result, err := uc.Execute(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestCreateChatUseCase_Execute_EventBusPublishError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  tenantID,
		ChatType:  "individual",
		ContactID: &contactID,
	}

	// Mock that no existing chat is found
	repo.On("FindIndividualByContact", mock.Anything, contactID, projectID).
		Return(nil, domainchat.ErrChatNotFound)

	// Mock successful creation
	repo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	// Mock event bus error (should not fail the operation)
	eventBusError := errors.New("event bus unavailable")
	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(eventBusError)

	result, err := uc.Execute(context.Background(), input)

	// Should succeed even if event publishing fails
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateChatUseCase_Execute_FindIndividualByContact_UnexpectedError(t *testing.T) {
	repo := new(MockChatRepository)
	eventBus := new(MockEventBus)
	uc := NewCreateChatUseCase(repo, eventBus)

	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	input := CreateChatInput{
		ProjectID: projectID,
		TenantID:  tenantID,
		ChatType:  "individual",
		ContactID: &contactID,
	}

	// Mock unexpected error from repository
	unexpectedError := errors.New("unexpected database error")
	repo.On("FindIndividualByContact", mock.Anything, contactID, projectID).
		Return(nil, unexpectedError)

	// Even with error, the use case continues to create the chat
	repo.On("Create", mock.Anything, mock.AnythingOfType("*chat.Chat")).
		Return(nil)

	eventBus.On("Publish", mock.Anything, mock.Anything).
		Return(nil)

	result, err := uc.Execute(context.Background(), input)

	// Should succeed and create new chat
	require.NoError(t, err)
	require.NotNil(t, result)

	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
