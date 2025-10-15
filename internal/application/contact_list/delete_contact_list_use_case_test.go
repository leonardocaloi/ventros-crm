package contact_list

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/domain/crm/contact_list"
)

func TestDeleteContactListUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewDeleteContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	req := DeleteContactListRequest{
		ContactListID: contactListID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, existingList.IsDeleted())
	mockRepo.AssertExpectations(t)
}

func TestDeleteContactListUseCase_Execute_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewDeleteContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	expectedErr := errors.New("contact list not found")

	req := DeleteContactListRequest{
		ContactListID: contactListID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(nil, expectedErr)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteContactListUseCase_Execute_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewDeleteContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	req := DeleteContactListRequest{
		ContactListID: contactListID,
	}

	expectedErr := errors.New("database error")
	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(expectedErr)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteContactListUseCase_Execute_StaticList(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewDeleteContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	req := DeleteContactListRequest{
		ContactListID: contactListID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, existingList.IsDeleted())
	mockRepo.AssertExpectations(t)
}

func TestDeleteContactListUseCase_Execute_DynamicListWithRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewDeleteContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Dynamic List",
		contact_list.LogicalOperatorOR,
		false,
	)

	// Add filter rules
	tagRule, _ := contact_list.NewTagFilterRule(contact_list.OperatorContains, "vip")
	existingList.AddFilterRule(tagRule)

	req := DeleteContactListRequest{
		ContactListID: contactListID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.True(t, existingList.IsDeleted())
	mockRepo.AssertExpectations(t)
}

func TestDeleteContactListUseCase_Execute_NilContactListID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewDeleteContactListUseCase(mockRepo)
	ctx := context.Background()

	req := DeleteContactListRequest{
		ContactListID: uuid.Nil,
	}

	expectedErr := errors.New("invalid contact list id")
	mockRepo.On("FindByID", ctx, uuid.Nil).Return(nil, expectedErr)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteContactListUseCase_NewUseCase(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)

	// Act
	useCase := NewDeleteContactListUseCase(mockRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.repo)
}
