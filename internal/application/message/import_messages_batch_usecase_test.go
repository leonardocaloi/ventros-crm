package message_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ventros/crm/internal/application/message"
	contactdomain "github.com/ventros/crm/internal/domain/crm/contact"
	messagedomain "github.com/ventros/crm/internal/domain/crm/message"
	sessiondomain "github.com/ventros/crm/internal/domain/crm/session"
	shareddomain "github.com/ventros/crm/internal/domain/core/shared"
	"go.uber.org/zap"
)

// Mock implementations
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) Save(ctx context.Context, contact *contactdomain.Contact) error {
	args := m.Called(ctx, contact)
	return args.Error(0)
}

func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*contactdomain.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*contactdomain.Contact, error) {
	args := m.Called(ctx, projectID, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByPhones(ctx context.Context, projectID uuid.UUID, phones []string) (map[string]*contactdomain.Contact, error) {
	args := m.Called(ctx, projectID, phones)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByEmail(ctx context.Context, projectID uuid.UUID, email string) (*contactdomain.Contact, error) {
	args := m.Called(ctx, projectID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*contactdomain.Contact, error) {
	args := m.Called(ctx, projectID, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) FindByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*contactdomain.Contact, error) {
	args := m.Called(ctx, projectID, limit, offset)
	return args.Get(0).([]*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) CountByProject(ctx context.Context, projectID uuid.UUID) (int, error) {
	args := m.Called(ctx, projectID)
	return args.Int(0), args.Error(1)
}

func (m *MockContactRepository) FindByTenantWithFilters(ctx context.Context, tenantID string, filters contactdomain.ContactFilters, page, limit int, sortBy, sortDir string) ([]*contactdomain.Contact, int64, error) {
	args := m.Called(ctx, tenantID, filters, page, limit, sortBy, sortDir)
	return args.Get(0).([]*contactdomain.Contact), int64(args.Int(1)), args.Error(2)
}

func (m *MockContactRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int) ([]*contactdomain.Contact, error) {
	args := m.Called(ctx, tenantID, searchText, limit)
	return args.Get(0).([]*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) SaveCustomFields(ctx context.Context, contactID uuid.UUID, fields map[string]string) error {
	args := m.Called(ctx, contactID, fields)
	return args.Error(0)
}

func (m *MockContactRepository) FindByCustomField(ctx context.Context, tenantID, key, value string) (*contactdomain.Contact, error) {
	args := m.Called(ctx, tenantID, key, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contactdomain.Contact), args.Error(1)
}

func (m *MockContactRepository) GetCustomFields(ctx context.Context, contactID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, contactID)
	return args.Get(0).(map[string]string), args.Error(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Save(ctx context.Context, session *sessiondomain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*sessiondomain.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) FindActiveByContact(ctx context.Context, contactID uuid.UUID, channelTypeID *int) (*sessiondomain.Session, error) {
	args := m.Called(ctx, contactID, channelTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByContact(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, contactID, limit, offset)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) FindExpiredSessions(ctx context.Context, batchSize int) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, batchSize)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByChannel(ctx context.Context, channelID uuid.UUID, limit, offset int) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, channelID, limit, offset)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) CountByChannel(ctx context.Context, channelID uuid.UUID) (int64, error) {
	args := m.Called(ctx, channelID)
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockSessionRepository) CountActiveByTenant(ctx context.Context, tenantID string) (int, error) {
	args := m.Called(ctx, tenantID)
	return args.Int(0), args.Error(1)
}

func (m *MockSessionRepository) FindByTenantWithFilters(ctx context.Context, filters sessiondomain.SessionFilters) ([]*sessiondomain.Session, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*sessiondomain.Session), int64(args.Int(1)), args.Error(2)
}

func (m *MockSessionRepository) DeleteBatch(ctx context.Context, sessionIDs []uuid.UUID) error {
	args := m.Called(ctx, sessionIDs)
	return args.Error(0)
}

func (m *MockSessionRepository) FindActiveBeforeTime(ctx context.Context, cutoffTime time.Time) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, cutoffTime)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) FindInactiveSessions(ctx context.Context, tenantID string) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) FindSessionsRequiringSummary(ctx context.Context, tenantID string, limit int) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, tenantID, limit)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*sessiondomain.Session, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	return args.Get(0).([]*sessiondomain.Session), int64(args.Int(1)), args.Error(2)
}

func (m *MockSessionRepository) FindByChannelPaginated(ctx context.Context, channelID uuid.UUID, limit int, offset int) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, channelID, limit, offset)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

func (m *MockSessionRepository) GetContactIDsByChannel(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, channelID)
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockSessionRepository) FindByChannelAndContacts(ctx context.Context, channelID uuid.UUID, contactIDs []uuid.UUID) ([]*sessiondomain.Session, error) {
	args := m.Called(ctx, channelID, contactIDs)
	return args.Get(0).([]*sessiondomain.Session), args.Error(1)
}

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Save(ctx context.Context, msg *messagedomain.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*messagedomain.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messagedomain.Message), args.Error(1)
}

func (m *MockMessageRepository) FindByChannelMessageID(ctx context.Context, channelMessageID string) (*messagedomain.Message, error) {
	args := m.Called(ctx, channelMessageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messagedomain.Message), args.Error(1)
}

func (m *MockMessageRepository) FindBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*messagedomain.Message, error) {
	args := m.Called(ctx, sessionID, limit, offset)
	return args.Get(0).([]*messagedomain.Message), args.Error(1)
}

func (m *MockMessageRepository) FindByContact(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*messagedomain.Message, error) {
	args := m.Called(ctx, contactID, limit, offset)
	return args.Get(0).([]*messagedomain.Message), args.Error(1)
}

func (m *MockMessageRepository) CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error) {
	args := m.Called(ctx, sessionID)
	return args.Int(0), args.Error(1)
}

func (m *MockMessageRepository) FindByChannelAndMessageID(ctx context.Context, channelID uuid.UUID, channelMessageID string) (*messagedomain.Message, error) {
	args := m.Called(ctx, channelID, channelMessageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*messagedomain.Message), args.Error(1)
}

func (m *MockMessageRepository) UpdateMessagesBulk(ctx context.Context, sessionID uuid.UUID, newSessionID uuid.UUID) (int, error) {
	args := m.Called(ctx, sessionID, newSessionID)
	return args.Int(0), args.Error(1)
}

func (m *MockMessageRepository) FindByTenantWithFilters(ctx context.Context, filters messagedomain.MessageFilters) ([]*messagedomain.Message, int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*messagedomain.Message), int64(args.Int(1)), args.Error(2)
}

func (m *MockMessageRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*messagedomain.Message, int64, error) {
	args := m.Called(ctx, tenantID, searchText, limit, offset)
	return args.Get(0).([]*messagedomain.Message), int64(args.Int(1)), args.Error(2)
}

func (m *MockMessageRepository) UpdateSessionIDForSession(ctx context.Context, oldSessionID, newSessionID uuid.UUID) (int64, error) {
	args := m.Called(ctx, oldSessionID, newSessionID)
	return int64(args.Int(0)), args.Error(1)
}

type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event shareddomain.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) PublishBatch(ctx context.Context, events []shareddomain.DomainEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

type MockSessionTimeoutResolver struct {
	mock.Mock
}

func (m *MockSessionTimeoutResolver) ResolveForChannel(ctx context.Context, channelID uuid.UUID) (time.Duration, *uuid.UUID, error) {
	args := m.Called(ctx, channelID)
	return args.Get(0).(time.Duration), args.Get(1).(*uuid.UUID), args.Error(2)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(context.Context) error) error {
	// For simplicity, just execute the function directly in tests
	return fn(ctx)
}

// Tests
func TestImportMessagesBatchUseCase_SingleContact_SingleSession(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	eventBus := new(MockEventBus)
	timeoutResolver := new(MockSessionTimeoutResolver)
	txManager := &MockTransactionManager{}

	// Create use case
	uc := message.NewImportMessagesBatchUseCase(
		contactRepo,
		sessionRepo,
		messageRepo,
		eventBus,
		timeoutResolver,
		txManager,
		logger,
	)

	// Test data
	projectID := uuid.New()
	channelID := uuid.New()
	tenantID := "tenant-123"
	customerID := uuid.New()
	phone := "5511999999999"

	// Mock: No existing contacts
	contactRepo.On("FindByPhones", mock.Anything, projectID, []string{phone}).
		Return(make(map[string]*contactdomain.Contact), nil)

	// Mock: Contact save
	contactRepo.On("Save", mock.Anything, mock.AnythingOfType("*contact.Contact")).
		Return(nil)

	// Mock: Timeout resolver
	timeout := 60 * time.Minute
	timeoutResolver.On("ResolveForChannel", mock.Anything, channelID).
		Return(timeout, (*uuid.UUID)(nil), nil)

	// Mock: No active session
	sessionRepo.On("FindActiveByContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*int")).
		Return(nil, sessiondomain.ErrSessionNotFound)

	// Mock: Session save
	sessionRepo.On("Save", mock.Anything, mock.AnythingOfType("*session.Session")).
		Return(nil)

	// Mock: No duplicate messages
	messageRepo.On("FindByChannelMessageID", mock.Anything, mock.AnythingOfType("string")).
		Return(nil, messagedomain.ErrMessageNotFound)

	// Mock: Message save
	messageRepo.On("Save", mock.Anything, mock.AnythingOfType("*message.Message")).
		Return(nil)

	// Mock: Event publishing
	eventBus.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]shared.DomainEvent")).
		Return(nil)

	// Create 3 messages with 30-minute gap (should be in same session)
	baseTime := time.Now().Add(-2 * time.Hour)
	input := message.ImportBatchInput{
		ChannelID:             channelID,
		ProjectID:             projectID,
		TenantID:              tenantID,
		CustomerID:            customerID,
		ChannelTypeID:         1,
		SessionTimeoutMinutes: 60, // 1 hour timeout
		Messages: []message.ImportMessage{
			{
				ExternalID:   "msg-1",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Hello",
				Timestamp:    baseTime,
				FromMe:       false,
			},
			{
				ExternalID:   "msg-2",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "How are you?",
				Timestamp:    baseTime.Add(30 * time.Minute), // 30 min later (same session)
				FromMe:       false,
			},
			{
				ExternalID:   "msg-3",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Goodbye",
				Timestamp:    baseTime.Add(45 * time.Minute), // 45 min later (same session)
				FromMe:       true,
			},
		},
	}

	// Execute
	result, err := uc.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ContactsCreated, "Should create 1 contact")
	assert.Equal(t, 1, result.SessionsCreated, "Should create 1 session (all messages within timeout)")
	assert.Equal(t, 3, result.MessagesCreated, "Should create 3 messages")
	assert.Equal(t, 0, result.Duplicates, "Should have no duplicates")

	// Verify mocks
	contactRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestImportMessagesBatchUseCase_SingleContact_MultipleSessions(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	eventBus := new(MockEventBus)
	timeoutResolver := new(MockSessionTimeoutResolver)
	txManager := &MockTransactionManager{}

	uc := message.NewImportMessagesBatchUseCase(
		contactRepo,
		sessionRepo,
		messageRepo,
		eventBus,
		timeoutResolver,
		txManager,
		logger,
	)

	// Test data
	projectID := uuid.New()
	channelID := uuid.New()
	tenantID := "tenant-123"
	customerID := uuid.New()
	phone := "5511999999999"

	// Mocks
	contactRepo.On("FindByPhones", mock.Anything, projectID, []string{phone}).
		Return(make(map[string]*contactdomain.Contact), nil)
	contactRepo.On("Save", mock.Anything, mock.AnythingOfType("*contact.Contact")).
		Return(nil)

	// Mock: Timeout resolver
	timeout := 60 * time.Minute
	timeoutResolver.On("ResolveForChannel", mock.Anything, channelID).
		Return(timeout, (*uuid.UUID)(nil), nil)

	sessionRepo.On("FindActiveByContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*int")).
		Return(nil, sessiondomain.ErrSessionNotFound)
	sessionRepo.On("Save", mock.Anything, mock.AnythingOfType("*session.Session")).
		Return(nil)
	messageRepo.On("FindByChannelMessageID", mock.Anything, mock.AnythingOfType("string")).
		Return(nil, messagedomain.ErrMessageNotFound)
	messageRepo.On("Save", mock.Anything, mock.AnythingOfType("*message.Message")).
		Return(nil)
	eventBus.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]shared.DomainEvent")).
		Return(nil)

	// Create messages with 90-minute gap (should create 2 sessions)
	baseTime := time.Now().Add(-3 * time.Hour)
	input := message.ImportBatchInput{
		ChannelID:             channelID,
		ProjectID:             projectID,
		TenantID:              tenantID,
		CustomerID:            customerID,
		ChannelTypeID:         1,
		SessionTimeoutMinutes: 60, // 1 hour timeout
		Messages: []message.ImportMessage{
			{
				ExternalID:   "msg-1",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Session 1 - Message 1",
				Timestamp:    baseTime,
				FromMe:       false,
			},
			{
				ExternalID:   "msg-2",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Session 2 - Message 1 (90min gap = new session)",
				Timestamp:    baseTime.Add(90 * time.Minute), // 90 min gap > 60 min timeout
				FromMe:       false,
			},
		},
	}

	// Execute
	result, err := uc.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ContactsCreated, "Should create 1 contact")
	assert.Equal(t, 2, result.SessionsCreated, "Should create 2 sessions (gap > timeout)")
	assert.Equal(t, 2, result.MessagesCreated, "Should create 2 messages")
}

func TestImportMessagesBatchUseCase_MultipleContacts(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	eventBus := new(MockEventBus)
	timeoutResolver := new(MockSessionTimeoutResolver)
	txManager := &MockTransactionManager{}

	uc := message.NewImportMessagesBatchUseCase(
		contactRepo,
		sessionRepo,
		messageRepo,
		eventBus,
		timeoutResolver,
		txManager,
		logger,
	)

	// Test data
	projectID := uuid.New()
	channelID := uuid.New()
	tenantID := "tenant-123"
	customerID := uuid.New()
	phone1 := "5511999999991"
	phone2 := "5511999999992"

	// Mocks
	contactRepo.On("FindByPhones", mock.Anything, projectID, mock.MatchedBy(func(phones []string) bool {
		return len(phones) == 2
	})).Return(make(map[string]*contactdomain.Contact), nil)
	contactRepo.On("Save", mock.Anything, mock.AnythingOfType("*contact.Contact")).
		Return(nil)

	// Mock: Timeout resolver
	timeout := 60 * time.Minute
	timeoutResolver.On("ResolveForChannel", mock.Anything, channelID).
		Return(timeout, (*uuid.UUID)(nil), nil)

	sessionRepo.On("FindActiveByContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*int")).
		Return(nil, sessiondomain.ErrSessionNotFound)
	sessionRepo.On("Save", mock.Anything, mock.AnythingOfType("*session.Session")).
		Return(nil)
	messageRepo.On("FindByChannelMessageID", mock.Anything, mock.AnythingOfType("string")).
		Return(nil, messagedomain.ErrMessageNotFound)
	messageRepo.On("Save", mock.Anything, mock.AnythingOfType("*message.Message")).
		Return(nil)
	eventBus.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]shared.DomainEvent")).
		Return(nil)

	baseTime := time.Now().Add(-1 * time.Hour)
	input := message.ImportBatchInput{
		ChannelID:             channelID,
		ProjectID:             projectID,
		TenantID:              tenantID,
		CustomerID:            customerID,
		ChannelTypeID:         1,
		SessionTimeoutMinutes: 60,
		Messages: []message.ImportMessage{
			{
				ExternalID:   "msg-1",
				ContactPhone: phone1,
				ContactName:  "User 1",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Hello from user 1",
				Timestamp:    baseTime,
				FromMe:       false,
			},
			{
				ExternalID:   "msg-2",
				ContactPhone: phone2,
				ContactName:  "User 2",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Hello from user 2",
				Timestamp:    baseTime.Add(10 * time.Minute),
				FromMe:       false,
			},
			{
				ExternalID:   "msg-3",
				ContactPhone: phone1,
				ContactName:  "User 1",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Second message from user 1",
				Timestamp:    baseTime.Add(20 * time.Minute),
				FromMe:       false,
			},
		},
	}

	// Execute
	result, err := uc.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ContactsCreated, "Should create 2 contacts")
	assert.Equal(t, 2, result.SessionsCreated, "Should create 2 sessions (one per contact)")
	assert.Equal(t, 3, result.MessagesCreated, "Should create 3 messages")
}

func TestImportMessagesBatchUseCase_Deduplication(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	eventBus := new(MockEventBus)
	timeoutResolver := new(MockSessionTimeoutResolver)
	txManager := &MockTransactionManager{}

	uc := message.NewImportMessagesBatchUseCase(
		contactRepo,
		sessionRepo,
		messageRepo,
		eventBus,
		timeoutResolver,
		txManager,
		logger,
	)

	projectID := uuid.New()
	channelID := uuid.New()
	phone := "5511999999999"

	// Mocks
	contactRepo.On("FindByPhones", mock.Anything, projectID, []string{phone}).
		Return(make(map[string]*contactdomain.Contact), nil)
	contactRepo.On("Save", mock.Anything, mock.AnythingOfType("*contact.Contact")).
		Return(nil)

	// Mock: Timeout resolver
	timeout := 60 * time.Minute
	timeoutResolver.On("ResolveForChannel", mock.Anything, channelID).
		Return(timeout, (*uuid.UUID)(nil), nil)

	sessionRepo.On("FindActiveByContact", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("*int")).
		Return(nil, sessiondomain.ErrSessionNotFound)
	sessionRepo.On("Save", mock.Anything, mock.AnythingOfType("*session.Session")).
		Return(nil)

	// Mock: First message is new, second is duplicate
	messageRepo.On("FindByChannelMessageID", mock.Anything, "msg-1").
		Return(nil, messagedomain.ErrMessageNotFound).Once()
	messageRepo.On("FindByChannelMessageID", mock.Anything, "msg-2").
		Return(&messagedomain.Message{}, nil).Once() // Duplicate!

	messageRepo.On("Save", mock.Anything, mock.AnythingOfType("*message.Message")).
		Return(nil)
	eventBus.On("PublishBatch", mock.Anything, mock.AnythingOfType("[]shared.DomainEvent")).
		Return(nil)

	baseTime := time.Now()
	input := message.ImportBatchInput{
		ChannelID:             channelID,
		ProjectID:             projectID,
		TenantID:              "tenant-123",
		CustomerID:            uuid.New(),
		ChannelTypeID:         1,
		SessionTimeoutMinutes: 60,
		Messages: []message.ImportMessage{
			{
				ExternalID:   "msg-1",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "New message",
				Timestamp:    baseTime,
				FromMe:       false,
			},
			{
				ExternalID:   "msg-2",
				ContactPhone: phone,
				ContactName:  "Test User",
				ContentType:  messagedomain.ContentTypeText,
				Text:         "Duplicate message",
				Timestamp:    baseTime.Add(1 * time.Minute),
				FromMe:       false,
			},
		},
	}

	// Execute
	result, err := uc.Execute(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.MessagesCreated, "Should create only 1 message (other is duplicate)")
	assert.Equal(t, 1, result.Duplicates, "Should detect 1 duplicate")
}
