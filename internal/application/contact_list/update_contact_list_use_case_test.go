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

func TestUpdateContactListUseCase_Execute_UpdateName(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Old Name",
		contact_list.LogicalOperatorAND,
		false,
	)

	newName := "New Name"
	req := UpdateContactListRequest{
		ContactListID: contactListID,
		Name:          &newName,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "New Name", existingList.Name())
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_UpdateDescription(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newDescription := "Updated description"
	req := UpdateContactListRequest{
		ContactListID: contactListID,
		Description:   &newDescription,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, existingList.Description())
	assert.Equal(t, "Updated description", *existingList.Description())
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_UpdateLogicalOperator(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newOperator := contact_list.LogicalOperatorOR
	req := UpdateContactListRequest{
		ContactListID:   contactListID,
		LogicalOperator: &newOperator,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, contact_list.LogicalOperatorOR, existingList.LogicalOperator())
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_UpdateFilterRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	// Add initial filter rule
	oldRule, _ := contact_list.NewTagFilterRule(contact_list.OperatorContains, "old")
	existingList.AddFilterRule(oldRule)

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeTag,
			Operator:   contact_list.OperatorContains,
			FieldKey:   "tag",
			Value:      "new",
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_UpdateAll(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Old Name",
		contact_list.LogicalOperatorAND,
		false,
	)

	newName := "New Name"
	newDescription := "New Description"
	newOperator := contact_list.LogicalOperatorOR
	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeTag,
			Operator:   contact_list.OperatorContains,
			FieldKey:   "tag",
			Value:      "vip",
		},
	}

	req := UpdateContactListRequest{
		ContactListID:   contactListID,
		Name:            &newName,
		Description:     &newDescription,
		LogicalOperator: &newOperator,
		FilterRules:     &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "New Name", existingList.Name())
	assert.Equal(t, "New Description", *existingList.Description())
	assert.Equal(t, contact_list.LogicalOperatorOR, existingList.LogicalOperator())
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	emptyName := ""
	req := UpdateContactListRequest{
		ContactListID: contactListID,
		Name:          &emptyName,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_ListNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	expectedErr := errors.New("contact list not found")

	newName := "New Name"
	req := UpdateContactListRequest{
		ContactListID: contactListID,
		Name:          &newName,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(nil, expectedErr)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newName := "New Name"
	req := UpdateContactListRequest{
		ContactListID: contactListID,
		Name:          &newName,
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

func TestUpdateContactListUseCase_Execute_NoChanges(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	req := UpdateContactListRequest{
		ContactListID: contactListID,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_FilterRulesWithCustomField(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	fieldType := "text"
	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeCustomField,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "company",
			FieldType:  &fieldType,
			Value:      "Acme Corp",
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_FilterRulesWithPipelineStatus(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypePipelineStatus,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "qualified",
			PipelineID: &pipelineID,
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_FilterRulesWithAttribute(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeAttribute,
			Operator:   contact_list.OperatorContains,
			FieldKey:   "email",
			Value:      "@example.com",
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_CustomFieldWithoutFieldType(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeCustomField,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "company",
			FieldType:  nil, // Missing field type
			Value:      "Acme Corp",
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field_type is required for custom_field filters")
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_PipelineStatusWithoutPipelineID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypePipelineStatus,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "qualified",
			PipelineID: nil, // Missing pipeline ID
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline_id is required for pipeline_status filters")
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_InvalidFilterRule(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeTag,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "",
			Value:      nil,
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_EmptyFilterRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	// Add initial filter rules
	tagRule, _ := contact_list.NewTagFilterRule(contact_list.OperatorContains, "vip")
	existingList.AddFilterRule(tagRule)

	emptyFilterRules := []FilterRuleRequest{}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &emptyFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_MultipleFilterRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
	ctx := context.Background()

	contactListID := uuid.New()
	projectID := uuid.New()
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	existingList, _ := contact_list.NewContactList(
		projectID,
		tenantID,
		"Test List",
		contact_list.LogicalOperatorAND,
		false,
	)

	fieldType := "text"
	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeTag,
			Operator:   contact_list.OperatorContains,
			FieldKey:   "tag",
			Value:      "vip",
		},
		{
			FilterType: contact_list.FilterTypeCustomField,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "company",
			FieldType:  &fieldType,
			Value:      "Acme Corp",
		},
		{
			FilterType: contact_list.FilterTypePipelineStatus,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "qualified",
			PipelineID: &pipelineID,
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_FilterRulesWithEvent(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeEvent,
			Operator:   contact_list.OperatorEquals,
			FieldKey:   "page_view",
			Value:      "homepage",
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_Execute_FilterRulesWithInteraction(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewUpdateContactListUseCase(mockRepo)
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

	newFilterRules := []FilterRuleRequest{
		{
			FilterType: contact_list.FilterTypeInteraction,
			Operator:   contact_list.OperatorGreaterThan,
			FieldKey:   "message_count",
			Value:      10,
		},
	}

	req := UpdateContactListRequest{
		ContactListID: contactListID,
		FilterRules:   &newFilterRules,
	}

	mockRepo.On("FindByID", ctx, contactListID).Return(existingList, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(existingList.FilterRules()))
	mockRepo.AssertExpectations(t)
}

func TestUpdateContactListUseCase_NewUseCase(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)

	// Act
	useCase := NewUpdateContactListUseCase(mockRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.repo)
}
