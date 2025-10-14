package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ventros/crm/internal/domain/crm/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Note: Mock types (MockSessionRepository, MockEventBus) are defined in close_session_test.go
// and shared across all session tests.

// ========== Test Helpers ==========

// buildTestSession creates a session for testing with optional configuration
func buildTestSession(contactID uuid.UUID, tenantID string, status session.Status) *session.Session {
	sess, _ := session.NewSession(contactID, tenantID, nil, 30*time.Minute)
	sess.ClearEvents() // Clear SessionStartedEvent from creation

	if status == session.StatusEnded {
		_ = sess.End(session.ReasonManualClose)
		sess.ClearEvents() // Clear SessionEndedEvent
	}

	return sess
}

// ========== Constructor Tests ==========

func TestNewRecordMessageUseCase(t *testing.T) {
	// Arrange
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	// Act
	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, sessionRepo, useCase.sessionRepo)
	assert.Equal(t, eventBus, useCase.eventBus)
}

// ========== Success Tests ==========

func TestRecordMessageUseCase_Execute_Success_FromContact(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish (MessageRecordedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, activeSession.MessageCount(), "message count should be incremented")
	assert.Equal(t, 1, activeSession.MessagesFromContact(), "messages from contact should be incremented")
	assert.Equal(t, 0, activeSession.MessagesFromAgent(), "messages from agent should remain 0")
	assert.NotNil(t, activeSession.FirstContactMessageAt(), "first contact message timestamp should be set")
	assert.Equal(t, messageTimestamp, *activeSession.FirstContactMessageAt())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_Success_FromAgent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      false,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish (MessageRecordedEvent)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, activeSession.MessageCount(), "message count should be incremented")
	assert.Equal(t, 0, activeSession.MessagesFromContact(), "messages from contact should remain 0")
	assert.Equal(t, 1, activeSession.MessagesFromAgent(), "messages from agent should be incremented")
	assert.NotNil(t, activeSession.FirstAgentResponseAt(), "first agent response timestamp should be set")
	assert.Equal(t, messageTimestamp, *activeSession.FirstAgentResponseAt())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_Success_MultipleMessages(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	// Record first message from contact
	cmd1 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: time.Now(),
	}

	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil).Once()
	sessionRepo.On("Save", ctx, activeSession).Return(nil).Once()
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil).Once()

	err1 := useCase.Execute(ctx, cmd1)
	assert.NoError(t, err1)

	// Clear events after first message
	activeSession.ClearEvents()

	// Record second message from agent
	cmd2 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      false,
		MessageTimestamp: time.Now(),
	}

	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil).Once()
	sessionRepo.On("Save", ctx, activeSession).Return(nil).Once()
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil).Once()

	err2 := useCase.Execute(ctx, cmd2)

	// Assert
	assert.NoError(t, err2)
	assert.Equal(t, 2, activeSession.MessageCount(), "message count should be 2")
	assert.Equal(t, 1, activeSession.MessagesFromContact(), "messages from contact should be 1")
	assert.Equal(t, 1, activeSession.MessagesFromAgent(), "messages from agent should be 1")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_Success_WithDifferentTimestamps(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create timestamps
	timestamp1 := time.Now().Add(-10 * time.Minute)
	timestamp2 := time.Now().Add(-5 * time.Minute)
	timestamp3 := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	// Record message with timestamp1
	cmd1 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: timestamp1,
	}

	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil).Times(3)
	sessionRepo.On("Save", ctx, activeSession).Return(nil).Times(3)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil).Times(3)

	err := useCase.Execute(ctx, cmd1)
	assert.NoError(t, err)
	activeSession.ClearEvents()

	// Record message with timestamp2
	cmd2 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      false,
		MessageTimestamp: timestamp2,
	}

	err = useCase.Execute(ctx, cmd2)
	assert.NoError(t, err)
	activeSession.ClearEvents()

	// Record message with timestamp3
	cmd3 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: timestamp3,
	}

	err = useCase.Execute(ctx, cmd3)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 3, activeSession.MessageCount(), "message count should be 3")
	assert.NotNil(t, activeSession.FirstContactMessageAt())
	assert.Equal(t, timestamp1, *activeSession.FirstContactMessageAt(), "first contact message should use earliest timestamp")
	assert.NotNil(t, activeSession.FirstAgentResponseAt())
	assert.Equal(t, timestamp2, *activeSession.FirstAgentResponseAt(), "first agent response should use agent's timestamp")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_Success_EventsClearedAfterPublish(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, activeSession.DomainEvents(), "events should be cleared after successful execution")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Validation Error Tests ==========

func TestRecordMessageUseCase_Execute_ValidationError_MissingSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	cmd := RecordMessageCommand{
		SessionID:        uuid.Nil,
		FromContact:      true,
		MessageTimestamp: time.Now(),
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

// ========== Repository Error Tests ==========

func TestRecordMessageUseCase_Execute_SessionNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: time.Now(),
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

func TestRecordMessageUseCase_Execute_RepositoryFindError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: time.Now(),
	}

	// Mock FindByID - database error
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

func TestRecordMessageUseCase_Execute_RepositorySaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save error
	saveError := errors.New("database save error")
	sessionRepo.On("Save", ctx, activeSession).Return(saveError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, saveError, err)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertNotCalled(t, "Publish")
}

// ========== Domain Error Tests ==========

func TestRecordMessageUseCase_Execute_DomainError_SessionNotActive(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an ended session
	endedSession := buildTestSession(contactID, tenantID, session.StatusEnded)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(endedSession, nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add message to non-active session")

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

// ========== EventBus Error Tests ==========

func TestRecordMessageUseCase_Execute_EventBusError_DoesNotFailUseCase(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish error - should be logged but not fail the use case
	publishError := errors.New("event bus error")
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(publishError)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	// Event bus errors are logged but should NOT fail the use case
	// based on the code comment: "// Log error"
	assert.NoError(t, err, "event bus error should be logged but not fail the use case")
	assert.Empty(t, activeSession.DomainEvents(), "events should still be cleared")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Multiple Events Tests ==========

func TestRecordMessageUseCase_Execute_WithMultipleEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish - RecordMessage generates one MessageRecordedEvent
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)

	// Verify Publish was called once for MessageRecordedEvent
	eventBus.AssertNumberOfCalls(t, "Publish", 1)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_PublishesCorrectEventType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      false,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish - verify event structure
	var publishedEvent session.MessageRecordedEvent
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).
		Run(func(args mock.Arguments) {
			publishedEvent = args.Get(1).(session.MessageRecordedEvent)
		}).
		Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)

	// Verify event was published with correct data
	assert.Equal(t, activeSession.ID(), publishedEvent.SessionID)
	assert.False(t, publishedEvent.FromContact, "event should indicate message is not from contact")
	assert.NotZero(t, publishedEvent.RecordedAt)

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

// ========== Edge Cases ==========

func TestRecordMessageUseCase_Execute_WithZeroTimestamp(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: time.Time{}, // zero timestamp
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, activeSession.MessageCount())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_ContextCancelled(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: time.Now(),
	}

	// Mock FindByID - return context cancelled error
	sessionRepo.On("FindByID", ctx, sessionID).Return(nil, context.Canceled)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	sessionRepo.AssertExpectations(t)
	sessionRepo.AssertNotCalled(t, "Save")
	eventBus.AssertNotCalled(t, "Publish")
}

func TestRecordMessageUseCase_Execute_UpdatesLastActivity(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"
	messageTimestamp := time.Now()

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)
	previousLastActivity := activeSession.LastActivityAt()

	// Wait a bit to ensure lastActivityAt changes
	time.Sleep(10 * time.Millisecond)

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: messageTimestamp,
	}

	// Mock FindByID
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)

	// Mock Save
	sessionRepo.On("Save", ctx, activeSession).Return(nil)

	// Mock event publish
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.True(t, activeSession.LastActivityAt().After(previousLastActivity), "lastActivityAt should be updated")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_CalculatesResponseTime(t *testing.T) {
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	// First message from contact
	contactMessageTime := time.Now().Add(-5 * time.Minute)
	cmd1 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: contactMessageTime,
	}

	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil).Times(2)
	sessionRepo.On("Save", ctx, activeSession).Return(nil).Times(2)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil).Times(2)

	err := useCase.Execute(ctx, cmd1)
	assert.NoError(t, err)
	// Note: The domain logic calculates agentResponseTimeSeconds based on time.Now()
	// when the first contact message is recorded, so it may not be nil

	activeSession.ClearEvents()

	// Agent responds
	agentMessageTime := time.Now()
	cmd2 := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      false,
		MessageTimestamp: agentMessageTime,
	}

	err = useCase.Execute(ctx, cmd2)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, activeSession.AgentResponseTimeSeconds(), "agent response time should be calculated")
	assert.NotNil(t, activeSession.FirstContactMessageAt(), "first contact message timestamp should be set")
	assert.NotNil(t, activeSession.FirstAgentResponseAt(), "first agent response timestamp should be set")

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestRecordMessageUseCase_Execute_NilContext(t *testing.T) {
	// Arrange
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()

	cmd := RecordMessageCommand{
		SessionID:        sessionID,
		FromContact:      true,
		MessageTimestamp: time.Now(),
	}

	// Mock FindByID with nil context - should handle gracefully
	sessionRepo.On("FindByID", mock.Anything, sessionID).Return(nil, errors.New("context error"))

	// Act
	err := useCase.Execute(nil, cmd)

	// Assert
	assert.Error(t, err)

	sessionRepo.AssertExpectations(t)
}

// ========== Integration-Like Tests ==========

func TestRecordMessageUseCase_Execute_ConversationFlow(t *testing.T) {
	// This test simulates a realistic conversation flow
	// Arrange
	ctx := context.Background()
	sessionRepo := new(MockSessionRepository)
	eventBus := new(MockEventBus)

	useCase := NewRecordMessageUseCase(sessionRepo, eventBus)

	sessionID := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-1"

	// Create an active session
	activeSession := buildTestSession(contactID, tenantID, session.StatusActive)

	// Mock repository and event bus for multiple calls
	sessionRepo.On("FindByID", ctx, sessionID).Return(activeSession, nil)
	sessionRepo.On("Save", ctx, activeSession).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("session.MessageRecordedEvent")).Return(nil)

	// Simulate conversation: Contact -> Agent -> Contact -> Agent
	messages := []struct {
		fromContact bool
		timestamp   time.Time
	}{
		{true, time.Now().Add(-10 * time.Minute)},
		{false, time.Now().Add(-8 * time.Minute)},
		{true, time.Now().Add(-5 * time.Minute)},
		{false, time.Now().Add(-2 * time.Minute)},
		{true, time.Now()},
	}

	for _, msg := range messages {
		cmd := RecordMessageCommand{
			SessionID:        sessionID,
			FromContact:      msg.fromContact,
			MessageTimestamp: msg.timestamp,
		}

		err := useCase.Execute(ctx, cmd)
		assert.NoError(t, err)
		activeSession.ClearEvents()
	}

	// Assert final state
	assert.Equal(t, 5, activeSession.MessageCount(), "should have recorded 5 messages")
	assert.Equal(t, 3, activeSession.MessagesFromContact(), "should have 3 messages from contact")
	assert.Equal(t, 2, activeSession.MessagesFromAgent(), "should have 2 messages from agent")
	assert.NotNil(t, activeSession.FirstContactMessageAt())
	assert.NotNil(t, activeSession.FirstAgentResponseAt())
	assert.NotNil(t, activeSession.AgentResponseTimeSeconds())

	sessionRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
