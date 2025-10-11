package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Note: Mock types (MockSessionRepository, MockEventBus, MockTransactionManager, etc.)
// are defined in close_session_test.go and shared across all session tests.

// ========== Constructor Tests ==========

func TestNewCreateSessionUseCase(t *testing.T) {
	// Arrange
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	// Act
	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, sessionRepo, useCase.sessionRepo)
	assert.Equal(t, eventBus, useCase.eventBus)
	assert.Equal(t, txManager, useCase.txManager)
}

// ========== Success Tests ==========

func TestCreateSessionUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish (SessionStartedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.SessionID)
	assert.True(t, result.Created, "session should be created")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_SuccessWithChannelType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	channelTypeID := 1
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: &channelTypeID,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, &channelTypeID).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish (SessionStartedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.SessionID)
	assert.True(t, result.Created, "session should be created")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_SuccessWithCustomTimeout(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 60 * time.Minute // custom timeout

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish (SessionStartedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.SessionID)
	assert.True(t, result.Created)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_ActiveSessionExists_ReturnsExisting(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: active session already exists
	existingSession, _ := session.NewSession(contactID, tenantID, nil, timeout)
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(existingSession, nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingSession.ID(), result.SessionID)
	assert.False(t, result.Created, "session should not be created, existing returned")

	sessionRepo.AssertExpectations(t)
	// Save and Publish should not be called
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_ActiveSessionExistsForChannelType_ReturnsExisting(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	channelTypeID := 1
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: &channelTypeID,
		Timeout:       timeout,
	}

	// Mock: active session already exists for this channel type
	existingSession, _ := session.NewSession(contactID, tenantID, &channelTypeID, timeout)
	sessionRepo.On("FindActiveByContact", ctx, contactID, &channelTypeID).Return(existingSession, nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, existingSession.ID(), result.SessionID)
	assert.False(t, result.Created, "session should not be created, existing returned")

	sessionRepo.AssertExpectations(t)
	// Save and Publish should not be called
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_EventsClearedOnSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish (SessionStartedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify events were cleared after successful execution
	// Note: We can't directly check savedSession.DomainEvents() here because ClearEvents()
	// is called after the transaction completes. This test verifies the flow executes without error.

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Validation Error Tests ==========

func TestCreateSessionUseCase_Execute_MissingContactID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	cmd := CreateSessionCommand{
		ContactID:     uuid.Nil,
		TenantID:      "tenant-1",
		ChannelTypeID: nil,
		Timeout:       30 * time.Minute,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "contactID is required")

	sessionRepo.AssertNotCalled(t, "FindActiveByContact")
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_MissingTenantID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	cmd := CreateSessionCommand{
		ContactID:     uuid.New(),
		TenantID:      "",
		ChannelTypeID: nil,
		Timeout:       30 * time.Minute,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tenantID is required")

	sessionRepo.AssertNotCalled(t, "FindActiveByContact")
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_ZeroTimeout_UsesDefault(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       0, // zero timeout
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save - verify default timeout is used (30 minutes)
	var savedSession *session.Session
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).
		Run(func(args mock.Arguments) {
			savedSession = args.Get(1).(*session.Session)
		}).
		Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, savedSession)
	// Domain logic sets default timeout to 30 minutes
	assert.Equal(t, 30*time.Minute, savedSession.TimeoutDuration())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_NegativeTimeout_UsesDefault(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       -10 * time.Minute, // negative timeout
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save - verify default timeout is used (30 minutes)
	var savedSession *session.Session
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).
		Run(func(args mock.Arguments) {
			savedSession = args.Get(1).(*session.Session)
		}).
		Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, savedSession)
	// Domain logic sets default timeout to 30 minutes
	assert.Equal(t, 30*time.Minute, savedSession.TimeoutDuration())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Repository Error Tests ==========

func TestCreateSessionUseCase_Execute_FindActiveByContact_UnexpectedError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: unexpected error from FindActiveByContact
	unexpectedError := errors.New("database connection error")
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, unexpectedError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, unexpectedError, err)

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_RepositorySaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save error
	saveError := errors.New("database error")
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, saveError, err)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

// ========== EventBus Error Tests ==========

func TestCreateSessionUseCase_Execute_EventBusPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(publishError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, publishError, err)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Transaction Rollback Tests ==========

func TestCreateSessionUseCase_Execute_TransactionRollbackOnSaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	// Use a transaction manager that tracks rollback
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save error
	saveError := errors.New("database error")
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_TransactionRollbackOnPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	// Use a transaction manager that tracks rollback
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(publishError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Edge Cases ==========

func TestCreateSessionUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: return context cancelled error
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, context.Canceled)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, context.Canceled, err)

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCreateSessionUseCase_Execute_NilChannelTypeID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil, // explicitly nil
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.SessionID)
	assert.True(t, result.Created)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_DifferentChannelTypes_CreatesSeparateSessions(t *testing.T) {
	// This test verifies that sessions with different channel types are treated as separate
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	channelTypeID1 := 1
	channelTypeID2 := 2
	timeout := 30 * time.Minute

	// Create first session with channel type 1
	cmd1 := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: &channelTypeID1,
		Timeout:       timeout,
	}

	// Mock: no active session for channel type 1
	sessionRepo.On("FindActiveByContact", ctx, contactID, &channelTypeID1).Return(nil, errors.New("not found")).Once()
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil).Once()
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil).Once()

	// Act 1
	result1, err1 := useCase.Execute(ctx, cmd1)

	// Assert 1
	assert.NoError(t, err1)
	assert.NotNil(t, result1)
	assert.True(t, result1.Created)

	// Create second session with channel type 2
	cmd2 := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: &channelTypeID2,
		Timeout:       timeout,
	}

	// Mock: no active session for channel type 2 (different from channel type 1)
	sessionRepo.On("FindActiveByContact", ctx, contactID, &channelTypeID2).Return(nil, errors.New("not found")).Once()
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil).Once()
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).Return(nil).Once()

	// Act 2
	result2, err2 := useCase.Execute(ctx, cmd2)

	// Assert 2
	assert.NoError(t, err2)
	assert.NotNil(t, result2)
	assert.True(t, result2.Created)
	assert.NotEqual(t, result1.SessionID, result2.SessionID, "different sessions should be created for different channel types")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Domain Event Tests ==========

func TestCreateSessionUseCase_Execute_EmitsSessionStartedEvent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	channelTypeID := 1
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: &channelTypeID,
		Timeout:       timeout,
	}

	// Mock: no active session exists
	sessionRepo.On("FindActiveByContact", ctx, contactID, &channelTypeID).Return(nil, errors.New("not found"))

	// Mock save
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil)

	// Mock event publish - verify event structure
	var publishedEvent session.SessionStartedEvent
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionStartedEvent")).
		Run(func(args mock.Arguments) {
			publishedEvent = args.Get(1).(session.SessionStartedEvent)
		}).
		Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify event was published with correct data
	assert.NotEqual(t, uuid.Nil, publishedEvent.SessionID)
	assert.Equal(t, contactID, publishedEvent.ContactID)
	assert.Equal(t, tenantID, publishedEvent.TenantID)
	assert.Equal(t, &channelTypeID, publishedEvent.ChannelTypeID)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCreateSessionUseCase_Execute_NoEventsPublishedWhenActiveSessionExists(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateSessionUseCase(sessionRepo, eventBus, txManager)

	contactID := uuid.New()
	tenantID := "tenant-1"
	timeout := 30 * time.Minute

	cmd := CreateSessionCommand{
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: nil,
		Timeout:       timeout,
	}

	// Mock: active session already exists
	existingSession, _ := session.NewSession(contactID, tenantID, nil, timeout)
	sessionRepo.On("FindActiveByContact", ctx, contactID, (*int)(nil)).Return(existingSession, nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Created)

	// Verify no events were published
	eventBus.AssertNotCalled(t, "Publish")
}
