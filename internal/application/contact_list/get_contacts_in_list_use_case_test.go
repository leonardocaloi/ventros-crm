package contact_list

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetContactsInListUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         10,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 10, 0).Return(contactIDs, 3, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 3, len(resp.ContactIDs))
	assert.Equal(t, contactIDs, resp.ContactIDs)
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_EmptyList(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         10,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 10, 0).Return(contactIDs, 0, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, 0, len(resp.ContactIDs))
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_WithPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
	}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         2,
		Offset:        10,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 2, 10).Return(contactIDs, 25, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 25, resp.Total)
	assert.Equal(t, 2, len(resp.ContactIDs))
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_LargeLimit(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{
		uuid.New(),
		uuid.New(),
		uuid.New(),
	}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         1000,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 1000, 0).Return(contactIDs, 3, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, resp.Total)
	assert.Equal(t, 3, len(resp.ContactIDs))
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_ZeroLimit(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         0,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 0, 0).Return(contactIDs, 50, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 50, resp.Total)
	assert.Equal(t, 0, len(resp.ContactIDs))
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	expectedErr := errors.New("database error")

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         10,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 10, 0).Return(nil, 0, expectedErr)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_ListNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	expectedErr := errors.New("contact list not found")

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         10,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 10, 0).Return(nil, 0, expectedErr)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_SingleContact(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{uuid.New()}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         10,
		Offset:        0,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 10, 0).Return(contactIDs, 1, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Total)
	assert.Equal(t, 1, len(resp.ContactIDs))
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_HighOffset(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	contactIDs := []uuid.UUID{}

	req := GetContactsInListRequest{
		ContactListID: contactListID,
		Limit:         10,
		Offset:        1000,
	}

	mockRepo.On("GetContactsInList", ctx, contactListID, 10, 1000).Return(contactIDs, 50, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 50, resp.Total)
	assert.Equal(t, 0, len(resp.ContactIDs))
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_Execute_NilContactListID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewGetContactsInListUseCase(mockRepo)
	ctx := context.Background()

	req := GetContactsInListRequest{
		ContactListID: uuid.Nil,
		Limit:         10,
		Offset:        0,
	}

	expectedErr := errors.New("invalid contact list id")
	mockRepo.On("GetContactsInList", ctx, uuid.Nil, 10, 0).Return(nil, 0, expectedErr)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestGetContactsInListUseCase_NewUseCase(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)

	// Act
	useCase := NewGetContactsInListUseCase(mockRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.repo)
}
