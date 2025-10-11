package contact

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========== Test Helpers ==========

func createTestContact(projectID uuid.UUID, tenantID string) *contact.Contact {
	c, _ := contact.NewContact(projectID, tenantID, "Test Contact")
	return c
}

func createTestPipeline(projectID uuid.UUID, tenantID string) *pipeline.Pipeline {
	p, _ := pipeline.NewPipeline(projectID, tenantID, "Test Pipeline")
	return p
}

func createTestStatus(pipelineID uuid.UUID, name string, statusType pipeline.StatusType) *pipeline.Status {
	s, _ := pipeline.NewStatus(pipelineID, name, statusType)
	return s
}

func createInactivePipeline(projectID uuid.UUID, tenantID string) *pipeline.Pipeline {
	p := createTestPipeline(projectID, tenantID)
	p.Deactivate()
	return p
}

func createInactiveStatus(pipelineID uuid.UUID, name string) *pipeline.Status {
	s := createTestStatus(pipelineID, name, pipeline.StatusTypeOpen)
	s.Deactivate()
	return s
}

// ========== Tests ==========

func TestChangePipelineStatusUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(nil, errors.New("not found"))
	pipelineRepo.On("SetContactStatus", ctx, contactID, pipelineID, statusID).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("*contact.ContactPipelineStatusChangedEvent")).Return(nil)
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, contactID, output.ContactID)
	assert.Equal(t, pipelineID, output.PipelineID)
	assert.Nil(t, output.PreviousStatusID)
	assert.Equal(t, "", output.PreviousStatusName)
	assert.Equal(t, statusID, output.NewStatusID)
	assert.Equal(t, "New Lead", output.NewStatusName)
	assert.NotEmpty(t, output.ChangedAt)

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_SuccessWithPreviousStatus(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	previousStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	newStatus := createTestStatus(pipelineID, "Qualified", pipeline.StatusTypeActive)

	previousStatusID := previousStatus.ID()
	newStatusID := newStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   newStatusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{previousStatus, newStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(previousStatus, nil)
	pipelineRepo.On("SetContactStatus", ctx, contactID, pipelineID, newStatusID).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("*contact.ContactPipelineStatusChangedEvent")).Return(nil)
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, contactID, output.ContactID)
	assert.Equal(t, pipelineID, output.PipelineID)
	assert.NotNil(t, output.PreviousStatusID)
	assert.Equal(t, previousStatusID, *output.PreviousStatusID)
	assert.Equal(t, "New Lead", output.PreviousStatusName)
	assert.Equal(t, newStatusID, output.NewStatusID)
	assert.Equal(t, "Qualified", output.NewStatusName)

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	input := ChangePipelineStatusInput{
		ContactID:  uuid.Nil,
		PipelineID: uuid.New(),
		StatusID:   uuid.New(),
		TenantID:   "tenant-1",
		ProjectID:  uuid.New(),
	}

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "contact_id is required", err.Error())

	contactRepo.AssertNotCalled(t, "FindByID")
}

func TestChangePipelineStatusUseCase_Execute_MissingPipelineID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	input := ChangePipelineStatusInput{
		ContactID:  uuid.New(),
		PipelineID: uuid.Nil,
		StatusID:   uuid.New(),
		TenantID:   "tenant-1",
		ProjectID:  uuid.New(),
	}

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "pipeline_id is required", err.Error())

	contactRepo.AssertNotCalled(t, "FindByID")
}

func TestChangePipelineStatusUseCase_Execute_MissingStatusID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	input := ChangePipelineStatusInput{
		ContactID:  uuid.New(),
		PipelineID: uuid.New(),
		StatusID:   uuid.Nil,
		TenantID:   "tenant-1",
		ProjectID:  uuid.New(),
	}

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "status_id is required", err.Error())

	contactRepo.AssertNotCalled(t, "FindByID")
}

func TestChangePipelineStatusUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	input := ChangePipelineStatusInput{
		ContactID:  uuid.New(),
		PipelineID: uuid.New(),
		StatusID:   uuid.New(),
		TenantID:   "",
		ProjectID:  uuid.New(),
	}

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "tenant_id is required", err.Error())

	contactRepo.AssertNotCalled(t, "FindByID")
}

func TestChangePipelineStatusUseCase_Execute_MissingProjectID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	input := ChangePipelineStatusInput{
		ContactID:  uuid.New(),
		PipelineID: uuid.New(),
		StatusID:   uuid.New(),
		TenantID:   "tenant-1",
		ProjectID:  uuid.Nil,
	}

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "project_id is required", err.Error())

	contactRepo.AssertNotCalled(t, "FindByID")
}

func TestChangePipelineStatusUseCase_Execute_ContactNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	contactID := uuid.New()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: uuid.New(),
		StatusID:   uuid.New(),
		TenantID:   "tenant-1",
		ProjectID:  uuid.New(),
	}

	contactRepo.On("FindByID", ctx, contactID).Return(nil, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "contact not found", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertNotCalled(t, "GetPipelineWithStatuses")
}

func TestChangePipelineStatusUseCase_Execute_ContactFindError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	contactID := uuid.New()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: uuid.New(),
		StatusID:   uuid.New(),
		TenantID:   "tenant-1",
		ProjectID:  uuid.New(),
	}

	dbError := errors.New("database connection error")
	contactRepo.On("FindByID", ctx, contactID).Return(nil, dbError)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "failed to find contact")

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertNotCalled(t, "GetPipelineWithStatuses")
}

func TestChangePipelineStatusUseCase_Execute_ContactDoesNotBelongToProject(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	differentProjectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()

	testContact := createTestContact(differentProjectID, tenantID)

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: uuid.New(),
		StatusID:   uuid.New(),
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "contact does not belong to this project", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertNotCalled(t, "GetPipelineWithStatuses")
}

func TestChangePipelineStatusUseCase_Execute_PipelineNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   uuid.New(),
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(nil, nil, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "pipeline not found", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_PipelineFindError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   uuid.New(),
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	dbError := errors.New("database error")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(nil, nil, dbError)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "failed to find pipeline")

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_PipelineDoesNotBelongToProject(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	differentProjectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(differentProjectID, tenantID)

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   uuid.New(),
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{}, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "pipeline does not belong to this project", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_PipelineNotActive(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createInactivePipeline(projectID, tenantID)

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   uuid.New(),
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{}, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "pipeline is not active", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_StatusNotFoundInPipeline(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()
	differentStatusID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   differentStatusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "status not found in pipeline", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_StatusNotActive(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createInactiveStatus(pipelineID, "Archived Lead")
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "status is not active", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_ContactAlreadyInStatus(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(testStatus, nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, "contact is already in this status", err.Error())

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	pipelineRepo.AssertNotCalled(t, "SetContactStatus")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestChangePipelineStatusUseCase_Execute_SetContactStatusError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	dbError := errors.New("database constraint violation")

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(nil, errors.New("not found"))
	pipelineRepo.On("SetContactStatus", ctx, contactID, pipelineID, statusID).Return(dbError)
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "failed to set contact status")

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestChangePipelineStatusUseCase_Execute_PublishEventError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	eventError := errors.New("event bus connection error")

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(nil, errors.New("not found"))
	pipelineRepo.On("SetContactStatus", ctx, contactID, pipelineID, statusID).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("*contact.ContactPipelineStatusChangedEvent")).Return(eventError)
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Contains(t, err.Error(), "failed to publish event")

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_TransactionError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	txError := errors.New("transaction commit failed")

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(nil, errors.New("not found"))
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(txError)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, txError, err)

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Execute_WithChangedByAndReason(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()
	changedBy := uuid.New()
	reason := "Contact requested information about premium services"

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	testStatus := createTestStatus(pipelineID, "Qualified", pipeline.StatusTypeActive)
	statusID := testStatus.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		ChangedBy:  &changedBy,
		Reason:     reason,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{testStatus}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(nil, errors.New("not found"))
	pipelineRepo.On("SetContactStatus", ctx, contactID, pipelineID, statusID).Return(nil)
	eventBus.On("Publish", ctx, mock.MatchedBy(func(event contact.DomainEvent) bool {
		if evt, ok := event.(*contact.ContactPipelineStatusChangedEvent); ok {
			return evt.ChangedBy != nil &&
				*evt.ChangedBy == changedBy &&
				evt.Reason == reason
		}
		return false
	})).Return(nil)
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, contactID, output.ContactID)
	assert.Equal(t, pipelineID, output.PipelineID)
	assert.Equal(t, statusID, output.NewStatusID)
	assert.Equal(t, "Qualified", output.NewStatusName)

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestChangePipelineStatusUseCase_Constructor(t *testing.T) {
	// Arrange
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	// Act
	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	// Assert
	assert.NotNil(t, useCase)
	assert.NotNil(t, useCase.contactRepo)
	assert.NotNil(t, useCase.pipelineRepo)
	assert.NotNil(t, useCase.eventBus)
	assert.NotNil(t, useCase.txManager)
}

func TestChangePipelineStatusUseCase_Execute_MultipleStatuses(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	pipelineRepo := new(MockPipelineRepository)
	eventBus := new(MockEventBus)
	txManager := new(MockTransactionManager)

	useCase := NewChangePipelineStatusUseCase(contactRepo, pipelineRepo, eventBus, txManager)

	projectID := uuid.New()
	tenantID := "tenant-1"
	contactID := uuid.New()
	pipelineID := uuid.New()

	testContact := createTestContact(projectID, tenantID)
	testPipeline := createTestPipeline(projectID, tenantID)
	status1 := createTestStatus(pipelineID, "New Lead", pipeline.StatusTypeOpen)
	status2 := createTestStatus(pipelineID, "Qualified", pipeline.StatusTypeActive)
	status3 := createTestStatus(pipelineID, "Converted", pipeline.StatusTypeClosed)

	status2ID := status2.ID()

	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   status2ID,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)
	pipelineRepo.On("GetPipelineWithStatuses", ctx, pipelineID).Return(testPipeline, []*pipeline.Status{status1, status2, status3}, nil)
	pipelineRepo.On("GetContactStatus", ctx, contactID, pipelineID).Return(nil, errors.New("not found"))
	pipelineRepo.On("SetContactStatus", ctx, contactID, pipelineID, status2ID).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("*contact.ContactPipelineStatusChangedEvent")).Return(nil)
	txManager.On("ExecuteInTransaction", ctx, mock.AnythingOfType("func(context.Context) error")).Return(nil)

	// Act
	output, err := useCase.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Qualified", output.NewStatusName)

	contactRepo.AssertExpectations(t)
	pipelineRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestChangePipelineStatusOutput_Fields(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	pipelineID := uuid.New()
	previousStatusID := uuid.New()
	newStatusID := uuid.New()
	changedAt := time.Now().Format("2006-01-02T15:04:05Z07:00")

	// Act
	output := &ChangePipelineStatusOutput{
		ContactID:          contactID,
		PipelineID:         pipelineID,
		PreviousStatusID:   &previousStatusID,
		PreviousStatusName: "Old Status",
		NewStatusID:        newStatusID,
		NewStatusName:      "New Status",
		ChangedAt:          changedAt,
	}

	// Assert
	assert.Equal(t, contactID, output.ContactID)
	assert.Equal(t, pipelineID, output.PipelineID)
	assert.NotNil(t, output.PreviousStatusID)
	assert.Equal(t, previousStatusID, *output.PreviousStatusID)
	assert.Equal(t, "Old Status", output.PreviousStatusName)
	assert.Equal(t, newStatusID, output.NewStatusID)
	assert.Equal(t, "New Status", output.NewStatusName)
	assert.Equal(t, changedAt, output.ChangedAt)
}

func TestChangePipelineStatusInput_AllFields(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	pipelineID := uuid.New()
	statusID := uuid.New()
	changedBy := uuid.New()
	reason := "Test reason"
	tenantID := "tenant-1"
	projectID := uuid.New()

	// Act
	input := ChangePipelineStatusInput{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		ChangedBy:  &changedBy,
		Reason:     reason,
		TenantID:   tenantID,
		ProjectID:  projectID,
	}

	// Assert
	assert.Equal(t, contactID, input.ContactID)
	assert.Equal(t, pipelineID, input.PipelineID)
	assert.Equal(t, statusID, input.StatusID)
	assert.NotNil(t, input.ChangedBy)
	assert.Equal(t, changedBy, *input.ChangedBy)
	assert.Equal(t, reason, input.Reason)
	assert.Equal(t, tenantID, input.TenantID)
	assert.Equal(t, projectID, input.ProjectID)
}
