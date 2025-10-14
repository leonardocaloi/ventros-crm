package contact_list

import (
	"context"
	"errors"
	"testing"

	"github.com/ventros/crm/internal/domain/crm/contact_list"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateContactListUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"
	description := "Test description"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "My List",
		Description:     &description,
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules:     []FilterRuleRequest{},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, uuid.Nil, resp.ContactListID)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateContactListRequest
		expectedErr string
	}{
		{
			name: "empty name",
			request: CreateContactListRequest{
				ProjectID:       uuid.New(),
				TenantID:        "tenant-123",
				Name:            "",
				LogicalOperator: contact_list.LogicalOperatorAND,
				IsStatic:        false,
			},
			expectedErr: "name is required",
		},
		{
			name: "nil project ID",
			request: CreateContactListRequest{
				ProjectID:       uuid.Nil,
				TenantID:        "tenant-123",
				Name:            "My List",
				LogicalOperator: contact_list.LogicalOperatorAND,
				IsStatic:        false,
			},
			expectedErr: "project_id is required",
		},
		{
			name: "empty tenant ID",
			request: CreateContactListRequest{
				ProjectID:       uuid.New(),
				TenantID:        "",
				Name:            "My List",
				LogicalOperator: contact_list.LogicalOperatorAND,
				IsStatic:        false,
			},
			expectedErr: "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockContactListRepository)
			useCase := NewCreateContactListUseCase(mockRepo)
			ctx := context.Background()

			// Act
			resp, err := useCase.Execute(ctx, tt.request)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestCreateContactListUseCase_Execute_WithFilterRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"
	fieldType := "text"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Filtered List",
		LogicalOperator: contact_list.LogicalOperatorOR,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
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
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, uuid.Nil, resp.ContactListID)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_WithPipelineStatusFilter(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	pipelineID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Pipeline List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypePipelineStatus,
				Operator:   contact_list.OperatorEquals,
				FieldKey:   "qualified",
				PipelineID: &pipelineID,
			},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_WithAttributeFilter(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Attribute List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypeAttribute,
				Operator:   contact_list.OperatorContains,
				FieldKey:   "email",
				Value:      "@example.com",
			},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_CustomFieldWithoutFieldType(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Invalid List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypeCustomField,
				Operator:   contact_list.OperatorEquals,
				FieldKey:   "company",
				FieldType:  nil, // Missing field type
				Value:      "Acme Corp",
			},
		},
	}

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "field_type is required for custom_field filters")
}

func TestCreateContactListUseCase_Execute_PipelineStatusWithoutPipelineID(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Invalid Pipeline List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypePipelineStatus,
				Operator:   contact_list.OperatorEquals,
				FieldKey:   "qualified",
				PipelineID: nil, // Missing pipeline ID
			},
		},
	}

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "pipeline_id is required for pipeline_status filters")
}

func TestCreateContactListUseCase_Execute_InvalidFilterRule(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Invalid Filter List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypeTag,
				Operator:   contact_list.OperatorEquals,
				FieldKey:   "",
				Value:      nil,
			},
		},
	}

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestCreateContactListUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "My List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules:     []FilterRuleRequest{},
	}

	expectedErr := errors.New("database error")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(expectedErr)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_StaticList(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Static List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        true,
		FilterRules:     []FilterRuleRequest{},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, uuid.Nil, resp.ContactListID)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_EmptyDescriptionNotSet(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"
	emptyDescription := ""

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "My List",
		Description:     &emptyDescription,
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules:     []FilterRuleRequest{},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_MultipleFilterRules(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	pipelineID := uuid.New()
	tenantID := "tenant-123"
	fieldType := "text"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Complex List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
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
			{
				FilterType: contact_list.FilterTypeAttribute,
				Operator:   contact_list.OperatorContains,
				FieldKey:   "email",
				Value:      "@example.com",
			},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEqual(t, uuid.Nil, resp.ContactListID)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_WithEventFilter(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Event List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypeEvent,
				Operator:   contact_list.OperatorEquals,
				FieldKey:   "page_view",
				Value:      "homepage",
			},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_Execute_WithInteractionFilter(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)
	useCase := NewCreateContactListUseCase(mockRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tenantID := "tenant-123"

	req := CreateContactListRequest{
		ProjectID:       projectID,
		TenantID:        tenantID,
		Name:            "Interaction List",
		LogicalOperator: contact_list.LogicalOperatorAND,
		IsStatic:        false,
		FilterRules: []FilterRuleRequest{
			{
				FilterType: contact_list.FilterTypeInteraction,
				Operator:   contact_list.OperatorGreaterThan,
				FieldKey:   "message_count",
				Value:      10,
			},
		},
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*contact_list.ContactList")).Return(nil)

	// Act
	resp, err := useCase.Execute(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestCreateContactListUseCase_NewUseCase(t *testing.T) {
	// Arrange
	mockRepo := new(MockContactListRepository)

	// Act
	useCase := NewCreateContactListUseCase(mockRepo)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.repo)
}
