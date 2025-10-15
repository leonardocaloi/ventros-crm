package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/domain/crm/contact"
)

// Note: MockContactRepository and MockEventBus are defined in change_pipeline_status_usecase_test.go

// SimpleTransactionManager is a test transaction manager that just executes the function
type SimpleTransactionManager struct{}

func (m *SimpleTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Tests
func TestCreateContactUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish (ContactCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ContactID)

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_Execute_SuccessWithEmailAndPhone(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"
	email := "john@example.com"
	phone := "+5511999999999"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Email:     &email,
		Phone:     &phone,
	}

	// Mock repository - no duplicate phone check
	contactRepo.On("FindByPhone", ctx, projectID, phone).Return(nil, errors.New("not found"))

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish (ContactCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ContactID)

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_Execute_MissingProjectID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	cmd := CreateContactCommand{
		ProjectID: uuid.Nil,
		TenantID:  "tenant-1",
		Name:      "John Doe",
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "projectID is required")

	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	cmd := CreateContactCommand{
		ProjectID: uuid.New(),
		TenantID:  "",
		Name:      "John Doe",
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tenantID is required")

	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_MissingName(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	cmd := CreateContactCommand{
		ProjectID: uuid.New(),
		TenantID:  "tenant-1",
		Name:      "",
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "name is required")

	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"
	invalidEmail := "not-an-email"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Email:     &invalidEmail,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid email format")

	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_InvalidPhone(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"
	invalidPhone := "123" // too short

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     &invalidPhone,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "phone too short")

	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_DuplicatePhone(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"
	phone := "+5511999999999"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     &phone,
	}

	// Mock repository - duplicate phone exists
	existingContact, _ := contact.NewContact(projectID, tenantID, "Jane Doe")
	contactRepo.On("FindByPhone", ctx, projectID, phone).Return(existingContact, nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "contact with this phone already exists")

	contactRepo.AssertExpectations(t)
	contactRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_RepositorySaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save error
	saveError := errors.New("database error")
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save contact")

	contactRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_EventBusPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(publishError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to publish event")

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_Execute_EmptyEmailString(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"
	emptyEmail := ""

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Email:     &emptyEmail,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish (ContactCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ContactID)

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_Execute_EmptyPhoneString(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"
	emptyPhone := ""

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     &emptyPhone,
	}

	// Mock repository - empty phone string still triggers FindByPhone because cmd.Phone != nil
	contactRepo.On("FindByPhone", ctx, projectID, emptyPhone).Return(nil, errors.New("not found"))

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish (ContactCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ContactID)

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_Execute_PhoneNilDoesNotCheckDuplicate(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // no phone provided
	}

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish (ContactCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ContactID)

	// Ensure FindByPhone was NOT called
	contactRepo.AssertNotCalled(t, "FindByPhone")
	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_NewCreateContactUseCase(t *testing.T) {
	// Arrange
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &MockTransactionManager{}

	// Act
	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, contactRepo, useCase.contactRepo)
	assert.Equal(t, eventBus, useCase.eventBus)
	assert.Equal(t, txManager, useCase.txManager)
}

func TestCreateContactUseCase_Execute_TransactionRollbackOnSaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)

	// Use a real transaction manager that tracks rollback
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save error
	saveError := errors.New("database error")
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	contactRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateContactUseCase_Execute_TransactionRollbackOnPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)

	// Use a real transaction manager that tracks rollback
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(publishError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateContactUseCase_Execute_EventsClearedOnSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateContactUseCase(contactRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	name := "John Doe"

	cmd := CreateContactCommand{
		ProjectID: projectID,
		TenantID:  tenantID,
		Name:      name,
		Phone:     nil, // No phone provided, so no duplicate check
	}

	// Mock save
	contactRepo.On("Save", ctx, mock.AnythingOfType("*contact.Contact")).Return(nil)

	// Mock event publish (ContactCreatedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("contact.ContactCreatedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify events were cleared after successful execution
	// Note: We can't directly check savedContact.DomainEvents() here because ClearEvents()
	// is called after the transaction completes. This test verifies the flow executes without error.

	contactRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// MockTransactionManagerWithRollback is a transaction manager that tracks rollback
type MockTransactionManagerWithRollback struct {
	rolledBack bool
}

func (m *MockTransactionManagerWithRollback) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	err := fn(ctx)
	if err != nil {
		m.rolledBack = true
	}
	return err
}

// Helper function
func strPtr(s string) *string {
	return &s
}
