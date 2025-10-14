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

func TestNewFindChatUseCase(t *testing.T) {
	repo := new(MockChatRepository)

	uc := NewFindChatUseCase(repo)

	assert.NotNil(t, uc)
	assert.Equal(t, repo, uc.chatRepo)
}

func TestFindChatUseCase_FindByID_Success(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	chatID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create a test chat
	chat, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)

	// Mock repository
	repo.On("FindByID", mock.Anything, chatID).
		Return(chat, nil)

	input := FindChatInput{
		ChatID: chatID,
	}

	result, err := uc.FindByID(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Chat)
	assert.Equal(t, chat.ID(), result.Chat.ID)
	assert.Equal(t, projectID, result.Chat.ProjectID)
	assert.Equal(t, tenantID, result.Chat.TenantID)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_FindByID_MissingChatID(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	input := FindChatInput{
		ChatID: uuid.Nil,
	}

	result, err := uc.FindByID(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "chat_id is required")
}

func TestFindChatUseCase_FindByID_ChatNotFound(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	chatID := uuid.New()

	// Mock repository returning not found error
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, domainchat.ErrChatNotFound)

	input := FindChatInput{
		ChatID: chatID,
	}

	result, err := uc.FindByID(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_FindByID_RepositoryError(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	chatID := uuid.New()

	// Mock repository error
	repoError := errors.New("database connection failed")
	repo.On("FindByID", mock.Anything, chatID).
		Return(nil, repoError)

	input := FindChatInput{
		ChatID: chatID,
	}

	result, err := uc.FindByID(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find chat")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_ByContactID_Success(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	// Create test chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat2, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)

	chats := []*domainchat.Chat{chat1, chat2}

	// Mock repository
	repo.On("FindByContact", mock.Anything, contactID).
		Return(chats, nil)

	input := ListChatsInput{
		ContactID: &contactID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 2)
	assert.Equal(t, 2, result.Total)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_ByProjectID_Success(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID1 := uuid.New()
	contactID2 := uuid.New()
	tenantID := "tenant-123"

	// Create test chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID1)
	chat2, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID2)

	chats := []*domainchat.Chat{chat1, chat2}

	// Mock repository
	repo.On("FindByProject", mock.Anything, projectID).
		Return(chats, nil)

	input := ListChatsInput{
		ProjectID: &projectID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 2)
	assert.Equal(t, 2, result.Total)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_ByProjectID_ActiveOnly(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID1 := uuid.New()
	contactID2 := uuid.New()
	tenantID := "tenant-123"

	// Create test chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID1)
	chat2, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID2)

	chats := []*domainchat.Chat{chat1, chat2}

	status := "active"

	// Mock repository - should call FindActiveByProject
	repo.On("FindActiveByProject", mock.Anything, projectID).
		Return(chats, nil)

	input := ListChatsInput{
		ProjectID: &projectID,
		Status:    &status,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 2)
	assert.Equal(t, 2, result.Total)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_ByTenantID_Success(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID1 := uuid.New()
	contactID2 := uuid.New()
	tenantID := "tenant-123"

	// Create test chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID1)
	chat2, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID2)

	chats := []*domainchat.Chat{chat1, chat2}

	// Mock repository
	repo.On("FindByTenant", mock.Anything, tenantID).
		Return(chats, nil)

	input := ListChatsInput{
		TenantID: &tenantID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 2)
	assert.Equal(t, 2, result.Total)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_NoFilters(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	input := ListChatsInput{}

	result, err := uc.ListChats(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one filter")
}

func TestFindChatUseCase_ListChats_WithStatusFilter(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID1 := uuid.New()
	contactID2 := uuid.New()
	tenantID := "tenant-123"

	// Create test chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID1)
	chat2, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID2)
	chat2.Archive() // Archive the second chat

	chats := []*domainchat.Chat{chat1, chat2}

	// Mock repository
	repo.On("FindByProject", mock.Anything, projectID).
		Return(chats, nil)

	status := "archived"
	input := ListChatsInput{
		ProjectID: &projectID,
		Status:    &status,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 1) // Only archived chat
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "archived", result.Chats[0].Status)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_WithChatTypeFilter(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create different types of chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat2, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)

	chats := []*domainchat.Chat{chat1, chat2}

	// Mock repository
	repo.On("FindByProject", mock.Anything, projectID).
		Return(chats, nil)

	chatType := "group"
	input := ListChatsInput{
		ProjectID: &projectID,
		ChatType:  &chatType,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 1) // Only group chat
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "group", result.Chats[0].ChatType)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_WithMultipleFilters(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID := uuid.New()
	creatorID := uuid.New()
	tenantID := "tenant-123"
	subject := "Test Group"

	// Create different types of chats
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat2, _ := domainchat.NewGroupChat(projectID, tenantID, subject, creatorID, nil)
	chat2.Archive()
	chat3, _ := domainchat.NewGroupChat(projectID, tenantID, "Another Group", creatorID, nil)

	chats := []*domainchat.Chat{chat1, chat2, chat3}

	// Mock repository - When status is "active", FindActiveByProject is called
	chatType := "group"
	status := "active"

	repo.On("FindActiveByProject", mock.Anything, projectID).
		Return(chats, nil)

	input := ListChatsInput{
		ProjectID: &projectID,
		ChatType:  &chatType,
		Status:    &status,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 1) // Only active group chat
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, "group", result.Chats[0].ChatType)
	assert.Equal(t, "active", result.Chats[0].Status)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_EmptyResult(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	contactID := uuid.New()

	// Mock repository returning empty list
	repo.On("FindByContact", mock.Anything, contactID).
		Return([]*domainchat.Chat{}, nil)

	input := ListChatsInput{
		ContactID: &contactID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 0)
	assert.Equal(t, 0, result.Total)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_RepositoryError_ByContact(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	contactID := uuid.New()

	// Mock repository error
	repoError := errors.New("database connection failed")
	repo.On("FindByContact", mock.Anything, contactID).
		Return(nil, repoError)

	input := ListChatsInput{
		ContactID: &contactID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list chats")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_RepositoryError_ByProject(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()

	// Mock repository error
	repoError := errors.New("database connection failed")
	repo.On("FindByProject", mock.Anything, projectID).
		Return(nil, repoError)

	input := ListChatsInput{
		ProjectID: &projectID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list chats")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_RepositoryError_ByTenant(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	tenantID := "tenant-123"

	// Mock repository error
	repoError := errors.New("database connection failed")
	repo.On("FindByTenant", mock.Anything, tenantID).
		Return(nil, repoError)

	input := ListChatsInput{
		TenantID: &tenantID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list chats")
	assert.Contains(t, err.Error(), "database connection failed")

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_NilChatsResult(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	contactID := uuid.New()

	// Mock repository returning nil
	repo.On("FindByContact", mock.Anything, contactID).
		Return(nil, nil)

	input := ListChatsInput{
		ContactID: &contactID,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Chats) // Should handle nil gracefully
	assert.Equal(t, 0, result.Total)

	repo.AssertExpectations(t)
}

func TestFindChatUseCase_ListChats_FilterNoMatches(t *testing.T) {
	repo := new(MockChatRepository)
	uc := NewFindChatUseCase(repo)

	projectID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"

	// Create test chats (all individual, all active)
	chat1, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)
	chat2, _ := domainchat.NewIndividualChat(projectID, tenantID, contactID)

	chats := []*domainchat.Chat{chat1, chat2}

	// Mock repository
	repo.On("FindByProject", mock.Anything, projectID).
		Return(chats, nil)

	// Filter for group chats (none exist)
	chatType := "group"
	input := ListChatsInput{
		ProjectID: &projectID,
		ChatType:  &chatType,
	}

	result, err := uc.ListChats(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Chats, 0)
	assert.Equal(t, 0, result.Total)

	repo.AssertExpectations(t)
}
