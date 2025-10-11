package message

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/application/message"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	domainMessage "github.com/caloi/ventros-crm/internal/domain/crm/message"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*contact.Contact), args.Error(1)
}

type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) GetActiveSessionByContact(ctx context.Context, contactID uuid.UUID) (*session.Session, error) {
	args := m.Called(ctx, contactID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

func (m *MockSessionRepository) Save(ctx context.Context, sess *session.Session) error {
	args := m.Called(ctx, sess)
	return args.Error(0)
}

type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Save(ctx context.Context, msg *domainMessage.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

type MockMessageSender struct {
	mock.Mock
}

func (m *MockMessageSender) SendMessage(ctx context.Context, msg *message.OutboundMessage) (*message.SendResult, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*message.SendResult), args.Error(1)
}

func (m *MockMessageSender) SendBulkMessages(ctx context.Context, messages []*message.OutboundMessage) ([]*message.SendResult, error) {
	args := m.Called(ctx, messages)
	return args.Get(0).([]*message.SendResult), args.Error(1)
}

func (m *MockMessageSender) GetSupportedTypes() []message.MessageType {
	args := m.Called()
	return args.Get(0).([]message.MessageType)
}

func (m *MockMessageSender) ValidateMessage(msg *message.OutboundMessage) error {
	args := m.Called(msg)
	return args.Error(0)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// For testing, just execute the function directly without real transaction
	return fn(ctx)
}

// Tests
func TestSendMessageCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *SendMessageCommand
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text message",
			cmd: &SendMessageCommand{
				ContactID:   uuid.New(),
				ChannelID:   uuid.New(),
				ContentType: domainMessage.ContentTypeText,
				Text:        strPtr("Hello World"),
				TenantID:    "tenant-1",
				ProjectID:   uuid.New(),
				CustomerID:  uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "missing contact_id",
			cmd: &SendMessageCommand{
				ContactID:   uuid.Nil,
				ChannelID:   uuid.New(),
				ContentType: domainMessage.ContentTypeText,
				Text:        strPtr("Hello"),
				TenantID:    "tenant-1",
				ProjectID:   uuid.New(),
				CustomerID:  uuid.New(),
			},
			wantErr: true,
			errMsg:  "contact_id is required",
		},
		{
			name: "missing channel_id",
			cmd: &SendMessageCommand{
				ContactID:   uuid.New(),
				ChannelID:   uuid.Nil,
				ContentType: domainMessage.ContentTypeText,
				Text:        strPtr("Hello"),
				TenantID:    "tenant-1",
				ProjectID:   uuid.New(),
				CustomerID:  uuid.New(),
			},
			wantErr: true,
			errMsg:  "channel_id is required",
		},
		{
			name: "text message without text",
			cmd: &SendMessageCommand{
				ContactID:   uuid.New(),
				ChannelID:   uuid.New(),
				ContentType: domainMessage.ContentTypeText,
				TenantID:    "tenant-1",
				ProjectID:   uuid.New(),
				CustomerID:  uuid.New(),
			},
			wantErr: true,
			errMsg:  "text is required for text messages",
		},
		{
			name: "media message without media_url",
			cmd: &SendMessageCommand{
				ContactID:   uuid.New(),
				ChannelID:   uuid.New(),
				ContentType: domainMessage.ContentTypeImage,
				TenantID:    "tenant-1",
				ProjectID:   uuid.New(),
				CustomerID:  uuid.New(),
			},
			wantErr: true,
			errMsg:  "media_url is required for media messages",
		},
		{
			name: "missing tenant_id",
			cmd: &SendMessageCommand{
				ContactID:   uuid.New(),
				ChannelID:   uuid.New(),
				ContentType: domainMessage.ContentTypeText,
				Text:        strPtr("Hello"),
				ProjectID:   uuid.New(),
				CustomerID:  uuid.New(),
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSendMessageHandler_Handle_ContactNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   uuid.New(),
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   uuid.New(),
		CustomerID:  uuid.New(),
	}

	contactRepo.On("FindByID", ctx, contactID).Return(nil, assert.AnError)

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "contact not found")
	contactRepo.AssertExpectations(t)
}

func TestSendMessageHandler_Handle_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	channelID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()

	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   channelID,
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello World"),
		TenantID:    "tenant-1",
		ProjectID:   projectID,
		CustomerID:  customerID,
	}

	// Mock contact
	testContact, _ := contact.NewContact(projectID, "John Doe", "tenant-1")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Mock active session
	testSession, _ := session.NewSession(contactID, "tenant-1", nil, 30*time.Minute)
	sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(testSession, nil)

	// Mock message save
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil).Times(2)

	// Mock message sending
	externalID := "wamid.123456"
	sendResult := &message.SendResult{
		ExternalID: &externalID,
		Status:     "sent",
	}
	messageSender.On("SendMessage", ctx, mock.AnythingOfType("*message.OutboundMessage")).Return(sendResult, nil)

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sent", result.Status)
	assert.NotNil(t, result.ExternalID)
	assert.Equal(t, externalID, *result.ExternalID)

	contactRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	messageSender.AssertExpectations(t)
}

func TestSendMessageCommand_Validate_InvalidContentType(t *testing.T) {
	cmd := &SendMessageCommand{
		ContactID:   uuid.New(),
		ChannelID:   uuid.New(),
		ContentType: domainMessage.ContentType("invalid_type"),
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   uuid.New(),
		CustomerID:  uuid.New(),
	}

	err := cmd.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid content_type")
}

func TestSendMessageCommand_Validate_MissingProjectID(t *testing.T) {
	cmd := &SendMessageCommand{
		ContactID:   uuid.New(),
		ChannelID:   uuid.New(),
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   uuid.Nil,
		CustomerID:  uuid.New(),
	}

	err := cmd.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project_id is required")
}

func TestSendMessageCommand_Validate_MissingCustomerID(t *testing.T) {
	cmd := &SendMessageCommand{
		ContactID:   uuid.New(),
		ChannelID:   uuid.New(),
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   uuid.New(),
		CustomerID:  uuid.Nil,
	}

	err := cmd.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "customer_id is required")
}

func TestSendMessageHandler_Handle_NoActiveSession_CreatesNewSession(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	channelID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()

	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   channelID,
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello World"),
		TenantID:    "tenant-1",
		ProjectID:   projectID,
		CustomerID:  customerID,
	}

	// Mock contact
	testContact, _ := contact.NewContact(projectID, "John Doe", "tenant-1")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Mock NO active session - should create new one
	sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(nil, errors.New("no active session"))

	// Mock session save (new session creation)
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).Return(nil).Times(1)

	// Mock message save (twice - create + update)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil).Times(2)

	// Mock message sending
	externalID := "wamid.123456"
	sendResult := &message.SendResult{
		ExternalID: &externalID,
		Status:     "sent",
	}
	messageSender.On("SendMessage", ctx, mock.AnythingOfType("*message.OutboundMessage")).Return(sendResult, nil)

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sent", result.Status)

	contactRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	messageSender.AssertExpectations(t)
}

func TestSendMessageHandler_Handle_MediaMessage_Image(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	channelID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()

	mediaURL := "https://example.com/image.jpg"
	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   channelID,
		ContentType: domainMessage.ContentTypeImage,
		MediaURL:    &mediaURL,
		TenantID:    "tenant-1",
		ProjectID:   projectID,
		CustomerID:  customerID,
	}

	// Mock contact
	testContact, _ := contact.NewContact(projectID, "John Doe", "tenant-1")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Mock active session
	testSession, _ := session.NewSession(contactID, "tenant-1", nil, 30*time.Minute)
	sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(testSession, nil)

	// Mock message save
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil).Times(2)

	// Mock message sending
	externalID := "wamid.image123"
	sendResult := &message.SendResult{
		ExternalID: &externalID,
		Status:     "sent",
	}
	messageSender.On("SendMessage", ctx, mock.MatchedBy(func(msg *message.OutboundMessage) bool {
		return msg.Type == message.MessageTypeImage && msg.MediaURL != nil && *msg.MediaURL == mediaURL
	})).Return(sendResult, nil)

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "sent", result.Status)

	messageSender.AssertExpectations(t)
}

func TestSendMessageHandler_Handle_SendMessageFails(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	channelID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()

	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   channelID,
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello World"),
		TenantID:    "tenant-1",
		ProjectID:   projectID,
		CustomerID:  customerID,
	}

	// Mock contact
	testContact, _ := contact.NewContact(projectID, "John Doe", "tenant-1")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Mock active session
	testSession, _ := session.NewSession(contactID, "tenant-1", nil, 30*time.Minute)
	sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(testSession, nil)

	// Mock message save (create + mark as failed)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil).Times(2)

	// Mock message sending FAILS
	messageSender.On("SendMessage", ctx, mock.AnythingOfType("*message.OutboundMessage")).
		Return(nil, errors.New("channel connection failed"))

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, string(domainMessage.StatusFailed), result.Status)
	assert.NotNil(t, result.Error)
	assert.Contains(t, *result.Error, "channel connection failed")

	contactRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	messageRepo.AssertExpectations(t)
	messageSender.AssertExpectations(t)
}

func TestSendMessageHandler_Handle_AllContentTypes(t *testing.T) {
	testCases := []struct {
		name        string
		contentType domainMessage.ContentType
		msgType     message.MessageType
		text        *string
		mediaURL    *string
	}{
		{
			name:        "Audio",
			contentType: domainMessage.ContentTypeAudio,
			msgType:     message.MessageTypeAudio,
			mediaURL:    strPtr("https://example.com/audio.mp3"),
		},
		{
			name:        "Video",
			contentType: domainMessage.ContentTypeVideo,
			msgType:     message.MessageTypeVideo,
			mediaURL:    strPtr("https://example.com/video.mp4"),
		},
		{
			name:        "Document",
			contentType: domainMessage.ContentTypeDocument,
			msgType:     message.MessageTypeDocument,
			mediaURL:    strPtr("https://example.com/doc.pdf"),
		},
		{
			name:        "Location",
			contentType: domainMessage.ContentTypeLocation,
			msgType:     message.MessageTypeLocation,
			text:        strPtr("latitude:40.7128,longitude:-74.0060"),
		},
		{
			name:        "Contact",
			contentType: domainMessage.ContentTypeContact,
			msgType:     message.MessageTypeContact,
			text:        strPtr("+1234567890"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			contactRepo := new(MockContactRepository)
			sessionRepo := new(MockSessionRepository)
			messageRepo := new(MockMessageRepository)
			messageSender := new(MockMessageSender)
			txManager := new(MockTransactionManager)

			handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

			contactID := uuid.New()
			channelID := uuid.New()
			projectID := uuid.New()
			customerID := uuid.New()

			cmd := &SendMessageCommand{
				ContactID:   contactID,
				ChannelID:   channelID,
				ContentType: tc.contentType,
				Text:        tc.text,
				MediaURL:    tc.mediaURL,
				TenantID:    "tenant-1",
				ProjectID:   projectID,
				CustomerID:  customerID,
			}

			// Mock contact
			testContact, _ := contact.NewContact(projectID, "John Doe", "tenant-1")
			contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

			// Mock active session
			testSession, _ := session.NewSession(contactID, "tenant-1", nil, 30*time.Minute)
			sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(testSession, nil)

			// Mock message save
			messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil).Times(2)

			// Mock message sending
			externalID := "wamid." + tc.name
			sendResult := &message.SendResult{
				ExternalID: &externalID,
				Status:     "sent",
			}
			messageSender.On("SendMessage", ctx, mock.MatchedBy(func(msg *message.OutboundMessage) bool {
				return msg.Type == tc.msgType
			})).Return(sendResult, nil)

			// Act
			result, err := handler.Handle(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "sent", result.Status)

			messageSender.AssertExpectations(t)
		})
	}
}

func TestSendMessageHandler_ConvertToOutboundMessage_DefaultPriority(t *testing.T) {
	handler := &SendMessageHandler{}

	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	channelID := uuid.New()

	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   channelID,
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   projectID,
		CustomerID:  customerID,
		Priority:    "", // Empty priority should default to Normal
	}

	msg, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)
	testContact, _ := contact.NewContact(projectID, "John Doe", "tenant-1")

	outbound := handler.convertToOutboundMessage(cmd, msg, testContact)

	assert.Equal(t, message.PriorityNormal, outbound.Priority)
}

func TestSendMessageHandler_Handle_MessageSaveError_FirstTransaction(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   uuid.New(),
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   uuid.New(),
		CustomerID:  uuid.New(),
	}

	// Mock contact
	testContact, _ := contact.NewContact(cmd.ProjectID, "John Doe", "tenant-1")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Mock active session
	testSession, _ := session.NewSession(contactID, "tenant-1", nil, 30*time.Minute)
	sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(testSession, nil)

	// Mock message save ERROR
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).
		Return(errors.New("database connection lost"))

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection lost")
}

func TestSendMessageHandler_Handle_SessionSaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	contactRepo := new(MockContactRepository)
	sessionRepo := new(MockSessionRepository)
	messageRepo := new(MockMessageRepository)
	messageSender := new(MockMessageSender)
	txManager := new(MockTransactionManager)

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender, txManager)

	contactID := uuid.New()
	cmd := &SendMessageCommand{
		ContactID:   contactID,
		ChannelID:   uuid.New(),
		ContentType: domainMessage.ContentTypeText,
		Text:        strPtr("Hello"),
		TenantID:    "tenant-1",
		ProjectID:   uuid.New(),
		CustomerID:  uuid.New(),
	}

	// Mock contact
	testContact, _ := contact.NewContact(cmd.ProjectID, "John Doe", "tenant-1")
	contactRepo.On("FindByID", ctx, contactID).Return(testContact, nil)

	// Mock NO active session
	sessionRepo.On("GetActiveSessionByContact", ctx, contactID).Return(nil, errors.New("no session"))

	// Mock session save ERROR
	sessionRepo.On("Save", ctx, mock.AnythingOfType("*session.Session")).
		Return(errors.New("failed to save session"))

	// Act
	result, err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save session")
}

// Helper functions
func strPtr(s string) *string {
	return &s
}
