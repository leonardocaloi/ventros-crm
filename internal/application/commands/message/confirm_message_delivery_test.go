package message

import (
	"context"
	"errors"
	"testing"
	"time"

	domainMessage "github.com/ventros/crm/internal/domain/crm/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageFinder is a mock implementation of MessageFinder interface
type MockMessageFinder struct {
	mock.Mock
}

func (m *MockMessageFinder) FindByID(ctx context.Context, id uuid.UUID) (*domainMessage.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainMessage.Message), args.Error(1)
}

func (m *MockMessageFinder) FindByChannelMessageID(ctx context.Context, channelMessageID string) (*domainMessage.Message, error) {
	args := m.Called(ctx, channelMessageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainMessage.Message), args.Error(1)
}

func (m *MockMessageFinder) Save(ctx context.Context, msg *domainMessage.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

// TestConfirmMessageDeliveryCommand_Validate tests all validation scenarios
func TestConfirmMessageDeliveryCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *ConfirmMessageDeliveryCommand
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid delivered status",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "delivered",
			},
			wantErr: false,
		},
		{
			name: "valid read status",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "read",
			},
			wantErr: false,
		},
		{
			name: "valid failed status",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "failed",
			},
			wantErr: false,
		},
		{
			name: "valid with delivered_at timestamp",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:   uuid.New(),
				ExternalID:  "wamid.123456",
				Status:      "delivered",
				DeliveredAt: timePtr(time.Now()),
			},
			wantErr: false,
		},
		{
			name: "valid with read_at timestamp",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "read",
				ReadAt:     timePtr(time.Now()),
			},
			wantErr: false,
		},
		{
			name: "valid with failure_reason",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:     uuid.New(),
				ExternalID:    "wamid.123456",
				Status:        "failed",
				FailureReason: stringPtr("Network error"),
			},
			wantErr: false,
		},
		{
			name: "missing message_id",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.Nil,
				ExternalID: "wamid.123456",
				Status:     "delivered",
			},
			wantErr: true,
			errMsg:  "message_id is required",
		},
		{
			name: "missing external_id",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "",
				Status:     "delivered",
			},
			wantErr: true,
			errMsg:  "external_id is required",
		},
		{
			name: "missing status",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "",
			},
			wantErr: true,
			errMsg:  "status is required",
		},
		{
			name: "invalid status - pending",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "pending",
			},
			wantErr: true,
			errMsg:  "invalid status: must be delivered, read, or failed",
		},
		{
			name: "invalid status - sent",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "sent",
			},
			wantErr: true,
			errMsg:  "invalid status: must be delivered, read, or failed",
		},
		{
			name: "invalid status - random string",
			cmd: &ConfirmMessageDeliveryCommand{
				MessageID:  uuid.New(),
				ExternalID: "wamid.123456",
				Status:     "invalid_status",
			},
			wantErr: true,
			errMsg:  "invalid status: must be delivered, read, or failed",
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

// TestConfirmMessageDeliveryHandler_Handle_Success_Delivered tests successful delivery confirmation
func TestConfirmMessageDeliveryHandler_Handle_Success_Delivered(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domainMessage.StatusDelivered, testMessage.Status())
	assert.NotNil(t, testMessage.DeliveredAt())
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_Success_Read tests successful read confirmation
func TestConfirmMessageDeliveryHandler_Handle_Success_Read(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "read",
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domainMessage.StatusRead, testMessage.Status())
	assert.NotNil(t, testMessage.ReadAt())
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_Success_Failed tests successful failure confirmation
func TestConfirmMessageDeliveryHandler_Handle_Success_Failed(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"
	failureReason := "Network timeout"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:     messageID,
		ExternalID:    externalID,
		Status:        "failed",
		FailureReason: &failureReason,
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domainMessage.StatusFailed, testMessage.Status())
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_FindByExternalID tests finding message by external ID when FindByID fails
func TestConfirmMessageDeliveryHandler_Handle_FindByExternalID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	// FindByID returns error, but FindByChannelMessageID succeeds
	messageRepo.On("FindByID", ctx, messageID).Return(nil, errors.New("not found"))
	messageRepo.On("FindByChannelMessageID", ctx, externalID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domainMessage.StatusDelivered, testMessage.Status())
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_FindByExternalID_WhenMessageNil tests finding by external ID when FindByID returns nil
func TestConfirmMessageDeliveryHandler_Handle_FindByExternalID_WhenMessageNil(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	// FindByID returns nil message, then FindByChannelMessageID succeeds
	messageRepo.On("FindByID", ctx, messageID).Return(nil, nil)
	messageRepo.On("FindByChannelMessageID", ctx, externalID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domainMessage.StatusDelivered, testMessage.Status())
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_MessageNotFound_ByID tests message not found by ID
func TestConfirmMessageDeliveryHandler_Handle_MessageNotFound_ByID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Both FindByID and FindByChannelMessageID return errors
	messageRepo.On("FindByID", ctx, messageID).Return(nil, errors.New("not found"))
	messageRepo.On("FindByChannelMessageID", ctx, externalID).Return(nil, errors.New("not found"))

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_MessageNotFound_BothNil tests when both lookups return nil
func TestConfirmMessageDeliveryHandler_Handle_MessageNotFound_BothNil(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Both FindByID and FindByChannelMessageID return nil message
	messageRepo.On("FindByID", ctx, messageID).Return(nil, nil)
	messageRepo.On("FindByChannelMessageID", ctx, externalID).Return(nil, nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_InvalidCommand tests handling invalid command
func TestConfirmMessageDeliveryHandler_Handle_InvalidCommand(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  uuid.Nil, // Invalid
		ExternalID: "wamid.123456",
		Status:     "delivered",
	}

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message_id is required")
	// No repository calls should be made
	messageRepo.AssertNotCalled(t, "FindByID")
	messageRepo.AssertNotCalled(t, "FindByChannelMessageID")
	messageRepo.AssertNotCalled(t, "Save")
}

// TestConfirmMessageDeliveryHandler_Handle_SaveError tests error during save
func TestConfirmMessageDeliveryHandler_Handle_SaveError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(errors.New("database error"))

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_InvalidStatusInSwitch tests invalid status after validation (edge case)
func TestConfirmMessageDeliveryHandler_Handle_InvalidStatusInSwitch(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	externalID := "wamid.123456"

	// Create command with invalid status that will fail validation
	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "unknown", // This will fail validation
	}

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	// Should fail at validation before reaching the switch
	assert.Contains(t, err.Error(), "invalid status")
}

// TestConfirmMessageDeliveryHandler_Handle_WithTimestamps tests handling with optional timestamps
func TestConfirmMessageDeliveryHandler_Handle_WithTimestamps(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()
	externalID := "wamid.123456"
	deliveredAt := time.Now().Add(-5 * time.Minute)
	readAt := time.Now()

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:   messageID,
		ExternalID:  externalID,
		Status:      "read",
		DeliveredAt: &deliveredAt,
		ReadAt:      &readAt,
	}

	// Create a test message
	testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

	messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
	messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domainMessage.StatusRead, testMessage.Status())
	// Note: The actual timestamps are set by the domain methods, not from the command
	// This test verifies that having these fields in the command doesn't cause issues
	messageRepo.AssertExpectations(t)
}

// TestConfirmMessageDeliveryHandler_Handle_AllStatusTransitions tests all status transitions
func TestConfirmMessageDeliveryHandler_Handle_AllStatusTransitions(t *testing.T) {
	statuses := []struct {
		status         string
		expectedStatus domainMessage.Status
	}{
		{"delivered", domainMessage.StatusDelivered},
		{"read", domainMessage.StatusRead},
		{"failed", domainMessage.StatusFailed},
	}

	for _, tc := range statuses {
		t.Run("status_"+tc.status, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			messageRepo := new(MockMessageFinder)
			handler := NewConfirmMessageDeliveryHandler(messageRepo)

			messageID := uuid.New()
			contactID := uuid.New()
			projectID := uuid.New()
			customerID := uuid.New()
			externalID := "wamid.123456"

			cmd := &ConfirmMessageDeliveryCommand{
				MessageID:  messageID,
				ExternalID: externalID,
				Status:     tc.status,
			}

			// Create a test message
			testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

			messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
			messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

			// Act
			err := handler.Handle(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, testMessage.Status())
			messageRepo.AssertExpectations(t)
		})
	}
}

// TestConfirmMessageDeliveryHandler_Handle_NilMessageID tests with nil UUID for MessageID
func TestConfirmMessageDeliveryHandler_Handle_NilMessageID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  uuid.Nil,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert - Should fail validation before reaching repository
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message_id is required")
}

// TestConfirmMessageDeliveryHandler_Handle_EmptyExternalID tests with empty external ID
func TestConfirmMessageDeliveryHandler_Handle_EmptyExternalID(t *testing.T) {
	// Arrange
	ctx := context.Background()
	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: "",
		Status:     "delivered",
	}

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "external_id is required")
}

// TestConfirmMessageDeliveryHandler_Handle_ContextCancellation tests handling with cancelled context
func TestConfirmMessageDeliveryHandler_Handle_ContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	messageRepo := new(MockMessageFinder)
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	messageID := uuid.New()
	externalID := "wamid.123456"

	cmd := &ConfirmMessageDeliveryCommand{
		MessageID:  messageID,
		ExternalID: externalID,
		Status:     "delivered",
	}

	// Repository returns context cancelled error for both lookups
	messageRepo.On("FindByID", ctx, messageID).Return(nil, context.Canceled)
	messageRepo.On("FindByChannelMessageID", ctx, externalID).Return(nil, context.Canceled)

	// Act
	err := handler.Handle(ctx, cmd)

	// Assert
	assert.Error(t, err)
	// Should fail when trying to find the message
	assert.Contains(t, err.Error(), "message not found")
	messageRepo.AssertExpectations(t)
}

// TestNewConfirmMessageDeliveryHandler tests handler constructor
func TestNewConfirmMessageDeliveryHandler(t *testing.T) {
	// Arrange
	messageRepo := new(MockMessageFinder)

	// Act
	handler := NewConfirmMessageDeliveryHandler(messageRepo)

	// Assert
	assert.NotNil(t, handler)
	assert.Equal(t, messageRepo, handler.messageRepo)
}

// TestConfirmMessageDeliveryHandler_Handle_DomainEvents tests that domain events are generated
func TestConfirmMessageDeliveryHandler_Handle_DomainEvents(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		checkEvent    func(*domainMessage.Message) bool
		expectedCount int
	}{
		{
			name:   "delivered status generates event",
			status: "delivered",
			checkEvent: func(msg *domainMessage.Message) bool {
				// Message should have events (1 from creation + 1 from delivery)
				events := msg.DomainEvents()
				return len(events) >= 2
			},
			expectedCount: 2,
		},
		{
			name:   "read status generates event",
			status: "read",
			checkEvent: func(msg *domainMessage.Message) bool {
				// Message should have events (1 from creation + 1 from read)
				events := msg.DomainEvents()
				return len(events) >= 2
			},
			expectedCount: 2,
		},
		{
			name:   "failed status does not generate additional event",
			status: "failed",
			checkEvent: func(msg *domainMessage.Message) bool {
				// Message should only have creation event
				events := msg.DomainEvents()
				return len(events) == 1
			},
			expectedCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			messageRepo := new(MockMessageFinder)
			handler := NewConfirmMessageDeliveryHandler(messageRepo)

			messageID := uuid.New()
			contactID := uuid.New()
			projectID := uuid.New()
			customerID := uuid.New()
			externalID := "wamid.123456"

			cmd := &ConfirmMessageDeliveryCommand{
				MessageID:  messageID,
				ExternalID: externalID,
				Status:     tc.status,
			}

			// Create a test message
			testMessage, _ := domainMessage.NewMessage(contactID, projectID, customerID, domainMessage.ContentTypeText, true)

			messageRepo.On("FindByID", ctx, messageID).Return(testMessage, nil)
			messageRepo.On("Save", ctx, mock.AnythingOfType("*message.Message")).Return(nil)

			// Act
			err := handler.Handle(ctx, cmd)

			// Assert
			assert.NoError(t, err)
			assert.True(t, tc.checkEvent(testMessage), "Expected domain events not found")
			events := testMessage.DomainEvents()
			assert.Equal(t, tc.expectedCount, len(events), "Unexpected number of domain events")
			messageRepo.AssertExpectations(t)
		})
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func stringPtr(s string) *string {
	return &s
}
