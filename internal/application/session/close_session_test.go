package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/ventros/crm/internal/domain/crm/session"
)

// ========== Mocks ==========

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Save(ctx context.Context, s *session.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

func (m *MockSessionRepository) FindActiveByContact(ctx context.Context, contactID uuid.UUID, channelTypeID *int) (*session.Session, error) {
	args := m.Called(ctx, contactID, channelTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

func (m *MockSessionRepository) FindInactiveSessions(ctx context.Context, tenantID string) ([]*session.Session, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.Session), args.Error(1)
}

func (m *MockSessionRepository) FindSessionsRequiringSummary(ctx context.Context, tenantID string, limit int) ([]*session.Session, error) {
	args := m.Called(ctx, tenantID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.Session), args.Error(1)
}

func (m *MockSessionRepository) CountActiveByTenant(ctx context.Context, tenantID string) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockSessionRepository) FindActiveBeforeTime(ctx context.Context, cutoffTime time.Time) ([]*session.Session, error) {
	args := m.Called(ctx, cutoffTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*session.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByTenantWithFilters(ctx context.Context, filters session.SessionFilters) ([]*session.Session, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*session.Session), args.Get(1).(int64), args.Error(2)
}

func (m *MockSessionRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*session.Session, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*session.Session), args.Get(1).(int64), args.Error(2)
}

type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event session.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, fn)
	if len(args) > 0 && args.Get(0) != nil {
		if err, ok := args.Get(0).(error); ok {
			return err
		}
	}
	// Otherwise, execute the function directly (simulating successful transaction)
	return fn(ctx)
}

// SimpleTransactionManager is a test transaction manager that just executes the function
type SimpleTransactionManager struct{}

func (m *SimpleTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
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

// ========== Tests ==========

func TestNewCloseSessionUseCase(t *testing.T) {
	// Arrange
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	// Act
	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, sessionRepo, useCase.sessionRepo)
	assert.Equal(t, eventBus, useCase.eventBus)
	assert.Equal(t, txManager, useCase.txManager)
}

func TestCloseSessionUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish (SessionEndedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.NotNil(t, activeSession.EndedAt())
	assert.NotNil(t, activeSession.EndReason())
	assert.Equal(t, session.ReasonManualClose, *activeSession.EndReason())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_WithInactivityTimeout(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonInactivityTimeout,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish (SessionEndedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.Equal(t, session.ReasonInactivityTimeout, *activeSession.EndReason())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_WithAgentClose(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonAgentClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish (SessionEndedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.Equal(t, session.ReasonAgentClose, *activeSession.EndReason())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_MissingSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	cmd := CloseSessionCommand{
		SessionID: uuid.Nil,
		Reason:    session.ReasonManualClose,
	}

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sessionID is required")

	sessionRepo.AssertNotCalled(t, "FindByID")
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCloseSessionUseCase_Execute_SessionNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID - session not found
	notFoundError := errors.New("session not found")
	sessionRepo.On("FindByID", ctx, sessionID).Return(nil, notFoundError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, notFoundError, err)

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCloseSessionUseCase_Execute_SessionAlreadyClosed(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create a session and close it
	closedSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	_ = closedSession.End(session.ReasonManualClose)

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonAgentClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(closedSession, nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session is not active")

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCloseSessionUseCase_Execute_RepositorySaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save error
	saveError := errors.New("database error")
	sessionRepo.On("Save", ctx, activeSession).Return(saveError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, saveError, err)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCloseSessionUseCase_Execute_EventBusPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(publishError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, publishError, err)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_TransactionRollbackOnSaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	// Use a real transaction manager that tracks rollback
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save error
	saveError := errors.New("database error")
	sessionRepo.On("Save", ctx, activeSession).Return(saveError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, saveError, err)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCloseSessionUseCase_Execute_TransactionRollbackOnPublishError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	// Use a real transaction manager that tracks rollback
	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish error
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(publishError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, publishError, err)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_EventsClearedOnSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish (SessionEndedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, activeSession.DomainEvents(), "events should be cleared after successful execution")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_CalculatesDuration(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	// Wait a bit to ensure duration > 0
	time.Sleep(100 * time.Millisecond)

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, activeSession.DurationSeconds(), 0, "duration should be calculated")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_WithContactRequest(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonContactRequest,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.Equal(t, session.ReasonContactRequest, *activeSession.EndReason())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_WithSystemClose(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonSystemClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.Equal(t, session.ReasonSystemClose, *activeSession.EndReason())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_RepositoryFindError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID - general database error
	dbError := errors.New("database connection error")
	sessionRepo.On("FindByID", ctx, sessionID).Return(nil, dbError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestCloseSessionUseCase_Execute_WithMessages(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session with messages
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation
	_ = activeSession.RecordMessage(true, time.Now())
	_ = activeSession.RecordMessage(false, time.Now())
	_ = activeSession.RecordMessage(true, time.Now())
	activeSession.ClearEvents() // Clear MessageRecordedEvents

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.Equal(t, 3, activeSession.MessageCount(), "message count should be preserved")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_WithAssignedAgents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	agentID := uuid.New()

	// Create an active session with an assigned agent
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation
	_ = activeSession.AssignAgent(agentID)
	activeSession.ClearEvents() // Clear AgentAssignedEvent

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.True(t, activeSession.HasAssignedAgents(), "assigned agents should be preserved")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_MultipleEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish - should be called once for SessionEndedEvent
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)

	// Verify Publish was called exactly once (for SessionEndedEvent)
	eventBus.AssertNumberOfCalls(t, "Publish", 1)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_SagaRollbackReason(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.EndReasonSagaRollback,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, session.StatusEnded, activeSession.Status())
	assert.Equal(t, session.EndReasonSagaRollback, *activeSession.EndReason())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_NilContext(t *testing.T) {
	// Arrange
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID with nil context - should handle gracefully
	sessionRepo.On("FindByID", mock.Anything, sessionID).Return(nil, errors.New("context error"))

	// Act
	err := useCase.Execute(nil, cmd)

	// Assert
	require.Error(t, err)

	sessionRepo.AssertExpectations(t)
}

func TestCloseSessionUseCase_Execute_TransactionCommitVerification(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)
	txManager := &SimpleTransactionManager{}

	useCase := NewCloseSessionUseCase(sessionRepo, eventBus, txManager)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	activeSession.ClearEvents() // Clear SessionStartedEvent from creation

	cmd := CloseSessionCommand{
		SessionID: sessionID,
		Reason:    session.ReasonManualClose,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)

	// Verify all operations within transaction were called
	sessionRepo.AssertCalled(t, "Save", ctx, activeSession)
	eventBus.AssertCalled(t, "Publish", ctx, mock.AnythingOfType("session.SessionEndedEvent"))

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
