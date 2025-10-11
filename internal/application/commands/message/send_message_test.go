package message

import (
	"context"
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/application/message"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	domainMessage "github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/caloi/ventros-crm/internal/domain/session"
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

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender)

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

	handler := NewSendMessageHandler(contactRepo, sessionRepo, messageRepo, messageSender)

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

// Helper functions
func strPtr(s string) *string {
	return &s
}
