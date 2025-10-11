package contact_list

import (
	"context"
	"errors"
	"testing"

	"github.com/caloi/ventros-crm/internal/domain/crm/contact_list"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestManageStaticListUseCase_AddContact_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true, // Static list
	)

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(false, nil)
	mockRepo.On("AddContactToStaticList", ctx, contactListID, contactID).Return(nil)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_ListNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	expectedErr := errors.New("contact list not found")

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(nil, expectedErr)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_DynamicListError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dynamicList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Dynamic List",
		contact_list.LogicalOperatorAND,
		false, // Dynamic list
	)

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(dynamicList, nil)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot manually add contacts to dynamic list")
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_AlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(true, nil)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contact already in list")
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_IsContactInListError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	expectedErr := errors.New("database error")
	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(false, expectedErr)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	expectedErr := errors.New("database error")
	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(false, nil)
	mockRepo.On("AddContactToStaticList", ctx, contactListID, contactID).Return(expectedErr)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(true, nil)
	mockRepo.On("RemoveContactFromStaticList", ctx, contactListID, contactID).Return(nil)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_ListNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	expectedErr := errors.New("contact list not found")

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(nil, expectedErr)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_DynamicListError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	dynamicList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Dynamic List",
		contact_list.LogicalOperatorAND,
		false,
	)

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(dynamicList, nil)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot manually remove contacts from dynamic list")
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_NotInList(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(false, nil)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contact not in list")
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_IsContactInListError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	expectedErr := errors.New("database error")
	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(false, expectedErr)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     contactID,
	}

	expectedErr := errors.New("database error")
	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, contactID).Return(true, nil)
	mockRepo.On("RemoveContactFromStaticList", ctx, contactListID, contactID).Return(expectedErr)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_NilContactListID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactID := uuid.New()
	expectedErr := errors.New("invalid contact list id")

	req := AddContactToListRequest{
		ContactListID: uuid.Nil,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, uuid.Nil).Return(nil, expectedErr)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_AddContact_NilContactID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := AddContactToListRequest{
		ContactListID: contactListID,
		ContactID:     uuid.Nil,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, uuid.Nil).Return(false, nil)
	mockRepo.On("AddContactToStaticList", ctx, contactListID, uuid.Nil).Return(nil)

	// Act
	err := useCase.AddContact(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_NilContactListID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactID := uuid.New()
	expectedErr := errors.New("invalid contact list id")

	req := RemoveContactFromListRequest{
		ContactListID: uuid.Nil,
		ContactID:     contactID,
	}

	mockRepo.On("FindByID", ctx, uuid.Nil).Return(nil, expectedErr)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_RemoveContact_NilContactID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewManageStaticListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	staticList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := RemoveContactFromListRequest{
		ContactListID: contactListID,
		ContactID:     uuid.Nil,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(staticList, nil)
	mockRepo.On("IsContactInList", ctx, contactListID, uuid.Nil).Return(true, nil)
	mockRepo.On("RemoveContactFromStaticList", ctx, contactListID, uuid.Nil).Return(nil)

	// Act
	err := useCase.RemoveContact(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestManageStaticListUseCase_NewUseCase(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)

	// Act
	useCase := NewManageStaticListUseCase(mockRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.repo)
}
