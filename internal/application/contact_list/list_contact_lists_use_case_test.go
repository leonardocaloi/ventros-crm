package contact_list

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ventros/crm/internal/domain/crm/contact_list"
)

func TestListContactListsUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()

	list1, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"List 1",
		contact_list.LogicalOperatorAND,
		false,
	)

	list2, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"List 2",
		contact_list.LogicalOperatorOR,
		true,
	)

	lists := []*contact_list.ContactList{list1, list2}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(lists, 2, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 2, len(resp.ContactLists))
	assert.Equal(t, "List 1", resp.ContactLists[0].Name)
	assert.Equal(t, "List 2", resp.ContactLists[1].Name)
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	lists := []*contact_list.ContactList{}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(lists, 0, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, 0, len(resp.ContactLists))
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_WithPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()

	list1, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"List 1",
		contact_list.LogicalOperatorAND,
		false,
	)

	lists := []*contact_list.ContactList{list1}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     1,
		Offset:    5,
	}

	mockRepo.On("ListByProject", ctx, projectID, 1, 5).Return(lists, 10, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 10, resp.Total)
	assert.Equal(t, 1, len(resp.ContactLists))
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_WithDescription(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()

	list1, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"List with Description",
		contact_list.LogicalOperatorAND,
		false,
	)
	list1.UpdateDescription("This is a test description")

	lists := []*contact_list.ContactList{list1}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(lists, 1, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.ContactLists))
	assert.NotNil(t, resp.ContactLists[0].Description)
	assert.Equal(t, "This is a test description", *resp.ContactLists[0].Description)
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_WithFilterRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()

	list1, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Filtered List",
		contact_list.LogicalOperatorOR,
		false,
	)

	// Add filter rules
	tagRule, _ := contact_list.NewTagFilterRule(contact_list.OperatorContains, "vip")
	list1.AddFilterRule(tagRule)

	attrRule, _ := contact_list.NewAttributeFilterRule("email", contact_list.OperatorContains, "@example.com")
	list1.AddFilterRule(attrRule)

	lists := []*contact_list.ContactList{list1}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(lists, 1, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.ContactLists))
	assert.Equal(t, 2, len(resp.ContactLists[0].FilterRules))
	assert.Equal(t, "tag", string(resp.ContactLists[0].FilterRules[0].FilterType))
	assert.Equal(t, "attribute", string(resp.ContactLists[0].FilterRules[1].FilterType))
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_StaticAndDynamicLists(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()

	staticList, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Static List",
		contact_list.LogicalOperatorAND,
		true,
	)

	dynamicList, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Dynamic List",
		contact_list.LogicalOperatorOR,
		false,
	)

	lists := []*contact_list.ContactList{staticList, dynamicList}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(lists, 2, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.ContactLists))
	assert.True(t, resp.ContactLists[0].IsStatic)
	assert.False(t, resp.ContactLists[1].IsStatic)
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	expectedErr := errors.New("database error")

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(nil, 0, expectedErr)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_LogicalOperators(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()

	andList, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"AND List",
		contact_list.LogicalOperatorAND,
		false,
	)

	orList, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"OR List",
		contact_list.LogicalOperatorOR,
		false,
	)

	lists := []*contact_list.ContactList{andList, orList}

	req := ListContactListsRequest{
		ProjectID: projectID,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, projectID, 10, 0).Return(lists, 2, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.ContactLists))
	assert.Equal(t, "AND", resp.ContactLists[0].LogicalOperator)
	assert.Equal(t, "OR", resp.ContactLists[1].LogicalOperator)
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_Execute_NilProjectID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)
	ctx := context.Background()

	req := ListContactListsRequest{
		ProjectID: uuid.Nil,
		Limit:     10,
		Offset:    0,
	}

	mockRepo.On("ListByProject", ctx, uuid.Nil, 10, 0).Return([]*contact_list.ContactList{}, 0, nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.ContactLists))
	mockRepo.AssertExpectations(t)
}

func TestListContactListsUseCase_toDTO_WithLastCalculatedAt(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)

	projectID := uuid.New()
	list, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	// Simulate recalculation
	list.UpdateContactCount(42)

	// Act
	dto := useCase.toDTO(list)

	// Assert
	assert.NotNil(t, dto)
	assert.Equal(t, 42, dto.ContactCount)
	assert.NotNil(t, dto.LastCalculatedAt)
}

func TestListContactListsUseCase_toDTO_WithoutLastCalculatedAt(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)

	projectID := uuid.New()
	list, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	// Act
	dto := useCase.toDTO(list)

	// Assert
	assert.NotNil(t, dto)
	assert.Equal(t, 0, dto.ContactCount)
	assert.Nil(t, dto.LastCalculatedAt)
}

func TestListContactListsUseCase_toDTO_TimestampsFormat(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)

	projectID := uuid.New()
	list, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	// Act
	dto := useCase.toDTO(list)

	// Assert
	assert.NotNil(t, dto)

	// Check timestamp format (RFC3339)
	_, err := time.Parse("2006-01-02T15:04:05Z07:00", dto.CreatedAt)
	assert.NoError(t, err)

	_, err = time.Parse("2006-01-02T15:04:05Z07:00", dto.UpdatedAt)
	assert.NoError(t, err)
}

func TestListContactListsUseCase_toDTO_FilterRuleFieldType(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)

	projectID := uuid.New()
	list, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	// Add custom field filter with field type
	customFieldRule, _ := contact_list.NewCustomFieldFilterRule(
		"company",
		"text",
		contact_list.OperatorEquals,
		"Acme Corp",
	)
	list.AddFilterRule(customFieldRule)

	// Act
	dto := useCase.toDTO(list)

	// Assert
	assert.NotNil(t, dto)
	assert.Equal(t, 1, len(dto.FilterRules))
	assert.NotNil(t, dto.FilterRules[0].FieldType)
	assert.Equal(t, "text", *dto.FilterRules[0].FieldType)
}

func TestListContactListsUseCase_toDTO_FilterRulePipelineID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewListContactListsUseCase(mockRepo)

	projectID := uuid.New()
	pipelineID := uuid.New()

	list, _ := contact_list.NewContactList(
		projectID,
		"tenant-123",
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	// Add pipeline status filter
	pipelineRule, _ := contact_list.NewPipelineStatusFilterRule(
		pipelineID,
		"qualified",
		contact_list.OperatorEquals,
	)
	list.AddFilterRule(pipelineRule)

	// Act
	dto := useCase.toDTO(list)

	// Assert
	assert.NotNil(t, dto)
	assert.Equal(t, 1, len(dto.FilterRules))
	assert.NotNil(t, dto.FilterRules[0].PipelineID)
	assert.Equal(t, pipelineID, *dto.FilterRules[0].PipelineID)
}

func TestListContactListsUseCase_NewUseCase(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)

	// Act
	useCase := NewListContactListsUseCase(mockRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.repo)
}
