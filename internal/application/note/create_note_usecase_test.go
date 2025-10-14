package note

import (
	"context"
	"errors"
	"testing"

	"github.com/ventros/crm/infrastructure/messaging"
	"github.com/ventros/crm/internal/domain/crm/note"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// ========== Mocks ==========


type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Save(ctx context.Context, n *note.Note) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*note.Note), args.Error(1)
}

func (m *MockNoteRepository) FindByContactID(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*note.Note, error) {
	args := m.Called(ctx, contactID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
}

func (m *MockNoteRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*note.Note, error) {
	args := m.Called(ctx, sessionID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
}

func (m *MockNoteRepository) FindPinned(ctx context.Context, contactID uuid.UUID) ([]*note.Note, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*note.Note), args.Error(1)
}

func (m *MockNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNoteRepository) CountByContact(ctx context.Context, contactID uuid.UUID) (int, error) {
	args := m.Called(ctx, contactID)
	return args.Int(0), args.Error(1)
}

func (m *MockNoteRepository) FindByTenantWithFilters(ctx context.Context, filters note.NoteFilters) ([]*note.Note, int64, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
}

func (m *MockNoteRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*note.Note, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*note.Note), args.Get(1).(int64), args.Error(2)
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
	return fn(ctx)
}

type SimpleTransactionManager struct{}

func (m *SimpleTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

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

// ========== Constructor Tests ==========

func TestNewCreateNoteUseCase(t *testing.T) {
	// Arrange
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	// Act
	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	// Assert
	assert.NotNil(t, useCase)
	assert.Equal(t, noteRepo, useCase.noteRepo)
	assert.Equal(t, logger, useCase.logger)
	assert.Equal(t, txManager, useCase.txManager)
}

// ========== Success Cases ==========

func TestCreateNoteUseCase_Execute_Success_MinimalFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()
	tenantID := "tenant-1"

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   tenantID,
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "This is a test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID())
	assert.Equal(t, contactID, result.ContactID())
	assert.Equal(t, tenantID, result.TenantID())
	assert.Equal(t, authorID, result.AuthorID())
	assert.Equal(t, note.AuthorTypeAgent, result.AuthorType())
	assert.Equal(t, "John Agent", result.AuthorName())
	assert.Equal(t, "This is a test note", result.Content())
	assert.Equal(t, note.NoteTypeGeneral, result.NoteType())
	assert.Equal(t, note.PriorityNormal, result.Priority())
	assert.False(t, result.VisibleToClient())
	assert.Nil(t, result.SessionID())
	assert.Empty(t, result.Tags())
	assert.Empty(t, result.Mentions())
	assert.Empty(t, result.Attachments())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	sessionID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		SessionID:  &sessionID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Session note",
		NoteType:   note.NoteTypeSessionSummary,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.SessionID())
	assert.Equal(t, sessionID, *result.SessionID())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithPriority(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Urgent note",
		NoteType:   note.NoteTypeEscalation,
		Priority:   note.PriorityUrgent,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, note.PriorityUrgent, result.Priority())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithVisibleToClient(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:       contactID,
		TenantID:        "tenant-1",
		AuthorID:        authorID,
		AuthorType:      note.AuthorTypeAgent,
		AuthorName:      "John Agent",
		Content:         "Client-visible note",
		NoteType:        note.NoteTypeCustomer,
		VisibleToClient: true,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.VisibleToClient())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithTags(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Tagged note",
		NoteType:   note.NoteTypeGeneral,
		Tags:       []string{"important", "follow-up", "customer-request"},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Tags(), 3)
	assert.Contains(t, result.Tags(), "important")
	assert.Contains(t, result.Tags(), "follow-up")
	assert.Contains(t, result.Tags(), "customer-request")

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithMentions(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()
	mentionID1 := uuid.New()
	mentionID2 := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Note with mentions",
		NoteType:   note.NoteTypeGeneral,
		Mentions:   []uuid.UUID{mentionID1, mentionID2},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Mentions(), 2)
	assert.Contains(t, result.Mentions(), mentionID1)
	assert.Contains(t, result.Mentions(), mentionID2)

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithAttachments(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:   contactID,
		TenantID:    "tenant-1",
		AuthorID:    authorID,
		AuthorType:  note.AuthorTypeAgent,
		AuthorName:  "John Agent",
		Content:     "Note with attachments",
		NoteType:    note.NoteTypeGeneral,
		Attachments: []string{"https://example.com/file1.pdf", "https://example.com/file2.jpg"},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Attachments(), 2)
	assert.Contains(t, result.Attachments(), "https://example.com/file1.pdf")
	assert.Contains(t, result.Attachments(), "https://example.com/file2.jpg")

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_WithAllFields(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	sessionID := uuid.New()
	authorID := uuid.New()
	mentionID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:       contactID,
		SessionID:       &sessionID,
		TenantID:        "tenant-1",
		AuthorID:        authorID,
		AuthorType:      note.AuthorTypeSystem,
		AuthorName:      "System",
		Content:         "Complete note",
		NoteType:        note.NoteTypeFollowUp,
		Priority:        note.PriorityHigh,
		VisibleToClient: true,
		Tags:            []string{"tag1", "tag2"},
		Mentions:        []uuid.UUID{mentionID},
		Attachments:     []string{"https://example.com/file.pdf"},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, contactID, result.ContactID())
	assert.Equal(t, sessionID, *result.SessionID())
	assert.Equal(t, "tenant-1", result.TenantID())
	assert.Equal(t, authorID, result.AuthorID())
	assert.Equal(t, note.AuthorTypeSystem, result.AuthorType())
	assert.Equal(t, "System", result.AuthorName())
	assert.Equal(t, "Complete note", result.Content())
	assert.Equal(t, note.NoteTypeFollowUp, result.NoteType())
	assert.Equal(t, note.PriorityHigh, result.Priority())
	assert.True(t, result.VisibleToClient())
	assert.Len(t, result.Tags(), 2)
	assert.Len(t, result.Mentions(), 1)
	assert.Len(t, result.Attachments(), 1)

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_EventsCleared(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Events should be cleared after successful execution
	assert.Empty(t, result.DomainEvents())

	noteRepo.AssertExpectations(t)
}

// ========== Validation Error Cases ==========

func TestCreateNoteUseCase_Execute_Error_InvalidContactID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  uuid.Nil, // Invalid
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create note")
	assert.ErrorIs(t, err, note.ErrInvalidContact)

	noteRepo.AssertNotCalled(t, "Save")
}

func TestCreateNoteUseCase_Execute_Error_InvalidAuthorID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   uuid.Nil, // Invalid
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create note")
	assert.ErrorIs(t, err, note.ErrInvalidAuthor)

	noteRepo.AssertNotCalled(t, "Save")
}

func TestCreateNoteUseCase_Execute_Error_EmptyContent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "", // Empty
		NoteType:   note.NoteTypeGeneral,
	}

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create note")
	assert.ErrorIs(t, err, note.ErrEmptyContent)

	noteRepo.AssertNotCalled(t, "Save")
}

// ========== Repository Error Cases ==========

func TestCreateNoteUseCase_Execute_Error_RepositorySaveFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Mock save error
	saveError := errors.New("database connection error")
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save note")

	noteRepo.AssertExpectations(t)
}

// ========== Event Bus Error Cases ==========
// Note: Event bus error testing is skipped as we cannot mock *messaging.DomainEventBus directly.
// These tests would require integration testing or refactoring the use case to accept an interface.

// ========== Transaction Rollback Cases ==========

func TestCreateNoteUseCase_Execute_Error_TransactionRollbackOnSaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()

	txManager := &MockTransactionManagerWithRollback{
		rolledBack: false,
	}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Mock save error
	saveError := errors.New("database error")
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(saveError)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, txManager.rolledBack, "transaction should have been rolled back")

	noteRepo.AssertExpectations(t)
}

// Note: TransactionRollbackOnPublishError test is skipped as we cannot mock event bus publish errors.

// ========== Different Author Types ==========

func TestCreateNoteUseCase_Execute_Success_AuthorTypeUser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeUser,
		AuthorName: "Customer",
		Content:    "Customer feedback",
		NoteType:   note.NoteTypeCustomer,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, note.AuthorTypeUser, result.AuthorType())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_AuthorTypeSystem(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeSystem,
		AuthorName: "System",
		Content:    "Automated note",
		NoteType:   note.NoteTypeInternal,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, note.AuthorTypeSystem, result.AuthorType())

	noteRepo.AssertExpectations(t)
}

// ========== Different Note Types ==========

func TestCreateNoteUseCase_Execute_Success_AllNoteTypes(t *testing.T) {
	noteTypes := []note.NoteType{
		note.NoteTypeGeneral,
		note.NoteTypeFollowUp,
		note.NoteTypeComplaint,
		note.NoteTypeResolution,
		note.NoteTypeEscalation,
		note.NoteTypeInternal,
		note.NoteTypeCustomer,
		note.NoteTypeSessionSummary,
		note.NoteTypeSessionHandoff,
		note.NoteTypeSessionFeedback,
		note.NoteTypeAdConversion,
		note.NoteTypeAdCampaign,
		note.NoteTypeAdAttribution,
		note.NoteTypeTrackingInsight,
	}

	for _, noteType := range noteTypes {
		t.Run(string(noteType), func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			noteRepo := new(MockNoteRepository)
			eventBus := (*messaging.DomainEventBus)(nil)
			logger := zap.NewNop()
			txManager := &SimpleTransactionManager{}

			useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

			contactID := uuid.New()
			authorID := uuid.New()

			cmd := CreateNoteCommand{
				ContactID:  contactID,
				TenantID:   "tenant-1",
				AuthorID:   authorID,
				AuthorType: note.AuthorTypeAgent,
				AuthorName: "John Agent",
				Content:    "Test note for " + string(noteType),
				NoteType:   noteType,
			}

			// Mock expectations
			noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, noteType, result.NoteType())

			noteRepo.AssertExpectations(t)
		})
	}
}

// ========== Different Priority Levels ==========

func TestCreateNoteUseCase_Execute_Success_AllPriorityLevels(t *testing.T) {
	priorities := []note.Priority{
		note.PriorityLow,
		note.PriorityNormal,
		note.PriorityHigh,
		note.PriorityUrgent,
	}

	for _, priority := range priorities {
		t.Run(string(priority), func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			noteRepo := new(MockNoteRepository)
			eventBus := (*messaging.DomainEventBus)(nil)
			logger := zap.NewNop()
			txManager := &SimpleTransactionManager{}

			useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

			contactID := uuid.New()
			authorID := uuid.New()

			cmd := CreateNoteCommand{
				ContactID:  contactID,
				TenantID:   "tenant-1",
				AuthorID:   authorID,
				AuthorType: note.AuthorTypeAgent,
				AuthorName: "John Agent",
				Content:    "Test note with " + string(priority) + " priority",
				NoteType:   note.NoteTypeGeneral,
				Priority:   priority,
			}

			// Mock expectations
			noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

			// Act
			result, err := useCase.Execute(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, priority, result.Priority())

			noteRepo.AssertExpectations(t)
		})
	}
}

// ========== Edge Cases ==========

func TestCreateNoteUseCase_Execute_Success_EmptyTags(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
		Tags:       []string{},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Tags())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_EmptyMentions(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
		Mentions:   []uuid.UUID{},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Mentions())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_EmptyAttachments(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:   contactID,
		TenantID:    "tenant-1",
		AuthorID:    authorID,
		AuthorType:  note.AuthorTypeAgent,
		AuthorName:  "John Agent",
		Content:     "Test note",
		NoteType:    note.NoteTypeGeneral,
		Attachments: []string{},
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Attachments())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_NilSessionID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		SessionID:  nil, // Explicitly nil
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.SessionID())

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_EmptyPriority(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Test note",
		NoteType:   note.NoteTypeGeneral,
		Priority:   "", // Empty priority should default to normal
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, note.PriorityNormal, result.Priority()) // Should default to normal

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_LongContent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	// Create a long content string (1000+ characters)
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "This is a very long note content. "
	}

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    longContent,
		NoteType:   note.NoteTypeGeneral,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, longContent, result.Content())
	assert.Greater(t, len(result.Content()), 1000)

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_ManyTags(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	// Create many tags
	tags := make([]string, 20)
	for i := 0; i < 20; i++ {
		tags[i] = "tag" + string(rune('A'+i))
	}

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Note with many tags",
		NoteType:   note.NoteTypeGeneral,
		Tags:       tags,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Tags(), 20)

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_ManyMentions(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	// Create many mentions
	mentions := make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		mentions[i] = uuid.New()
	}

	cmd := CreateNoteCommand{
		ContactID:  contactID,
		TenantID:   "tenant-1",
		AuthorID:   authorID,
		AuthorType: note.AuthorTypeAgent,
		AuthorName: "John Agent",
		Content:    "Note with many mentions",
		NoteType:   note.NoteTypeGeneral,
		Mentions:   mentions,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Mentions(), 10)

	noteRepo.AssertExpectations(t)
}

func TestCreateNoteUseCase_Execute_Success_ManyAttachments(t *testing.T) {
	// Arrange
	ctx := context.Background()
	noteRepo := new(MockNoteRepository)
	eventBus := (*messaging.DomainEventBus)(nil)
	logger := zap.NewNop()
	txManager := &SimpleTransactionManager{}

	useCase := NewCreateNoteUseCase(noteRepo, eventBus, logger, txManager)

	contactID := uuid.New()
	authorID := uuid.New()

	// Create many attachments
	attachments := make([]string, 15)
	for i := 0; i < 15; i++ {
		attachments[i] = "https://example.com/file" + string(rune('0'+i)) + ".pdf"
	}

	cmd := CreateNoteCommand{
		ContactID:   contactID,
		TenantID:    "tenant-1",
		AuthorID:    authorID,
		AuthorType:  note.AuthorTypeAgent,
		AuthorName:  "John Agent",
		Content:     "Note with many attachments",
		NoteType:    note.NoteTypeGeneral,
		Attachments: attachments,
	}

	// Mock expectations
	noteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Attachments(), 15)

	noteRepo.AssertExpectations(t)
}
