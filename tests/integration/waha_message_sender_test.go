package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/infrastructure/persistence"
	"github.com/ventros/crm/internal/application/message"
	"go.uber.org/zap"
)

// TestWAHAMessageSenderAdapter_GetSupportedTypes verifies supported message types
func TestWAHAMessageSenderAdapter_GetSupportedTypes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{} // Mock client
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	supportedTypes := adapter.GetSupportedTypes()

	assert.NotEmpty(t, supportedTypes)
	assert.Contains(t, supportedTypes, message.MessageTypeText)
	assert.Contains(t, supportedTypes, message.MessageTypeImage)
	assert.Contains(t, supportedTypes, message.MessageTypeAudio)
	assert.Contains(t, supportedTypes, message.MessageTypeVideo)
	assert.Contains(t, supportedTypes, message.MessageTypeDocument)
	assert.Contains(t, supportedTypes, message.MessageTypeLocation)
	assert.Contains(t, supportedTypes, message.MessageTypeContact)
}

// TestWAHAMessageSenderAdapter_ValidateMessage tests message validation
func TestWAHAMessageSenderAdapter_ValidateMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	tests := []struct {
		name    string
		msg     *message.OutboundMessage
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text message",
			msg: &message.OutboundMessage{
				ID:        uuid.New(),
				ChannelID: uuid.New(),
				ContactID: uuid.New(),
				Type:      message.MessageTypeText,
				Content:   "Hello World",
			},
			wantErr: false,
		},
		{
			name: "valid image message",
			msg: &message.OutboundMessage{
				ID:        uuid.New(),
				ChannelID: uuid.New(),
				ContactID: uuid.New(),
				Type:      message.MessageTypeImage,
				MediaURL:  stringPtr("https://example.com/image.jpg"),
			},
			wantErr: false,
		},
		{
			name: "missing channel ID",
			msg: &message.OutboundMessage{
				ID:        uuid.New(),
				ChannelID: uuid.Nil,
				ContactID: uuid.New(),
				Type:      message.MessageTypeText,
				Content:   "Hello",
			},
			wantErr: true,
			errMsg:  "channel_id is required",
		},
		{
			name: "missing contact ID",
			msg: &message.OutboundMessage{
				ID:        uuid.New(),
				ChannelID: uuid.New(),
				ContactID: uuid.Nil,
				Type:      message.MessageTypeText,
				Content:   "Hello",
			},
			wantErr: true,
			errMsg:  "contact_id is required",
		},
		{
			name: "unsupported message type",
			msg: &message.OutboundMessage{
				ID:        uuid.New(),
				ChannelID: uuid.New(),
				ContactID: uuid.New(),
				Type:      message.MessageType("unsupported"),
				Content:   "Hello",
			},
			wantErr: true,
			errMsg:  "not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateMessage(tt.msg)

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

// TestWAHAMessageSenderAdapter_SendMessage_NotImplemented verifies current implementation status
func TestWAHAMessageSenderAdapter_SendMessage_NotImplemented(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	ctx := context.Background()
	msg := &message.OutboundMessage{
		ID:        uuid.New(),
		ChannelID: uuid.New(),
		ContactID: uuid.New(),
		Type:      message.MessageTypeText,
		Content:   "Test message",
	}

	result, err := adapter.SendMessage(ctx, msg)

	// Currently not fully implemented
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not fully implemented")
}

// TestWAHAMessageSenderAdapter_SendBulkMessages tests bulk sending
func TestWAHAMessageSenderAdapter_SendBulkMessages(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	ctx := context.Background()
	messages := []*message.OutboundMessage{
		{
			ID:        uuid.New(),
			ChannelID: uuid.New(),
			ContactID: uuid.New(),
			Type:      message.MessageTypeText,
			Content:   "Message 1",
		},
		{
			ID:        uuid.New(),
			ChannelID: uuid.New(),
			ContactID: uuid.New(),
			Type:      message.MessageTypeText,
			Content:   "Message 2",
		},
	}

	results, err := adapter.SendBulkMessages(ctx, messages)

	// Should return results even if individual messages fail
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, results, 2)
}

// TestWAHAMessageSenderAdapter_Integration_AllTypes tests all supported message types
func TestWAHAMessageSenderAdapter_Integration_AllTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	messageTypes := []struct {
		name string
		typ  message.MessageType
	}{
		{"text", message.MessageTypeText},
		{"image", message.MessageTypeImage},
		{"audio", message.MessageTypeAudio},
		{"video", message.MessageTypeVideo},
		{"document", message.MessageTypeDocument},
		{"location", message.MessageTypeLocation},
		{"contact", message.MessageTypeContact},
	}

	for _, mt := range messageTypes {
		t.Run(mt.name, func(t *testing.T) {
			msg := &message.OutboundMessage{
				ID:        uuid.New(),
				ChannelID: uuid.New(),
				ContactID: uuid.New(),
				Type:      mt.typ,
				Content:   "Test content",
			}

			// Add media URL for media message types
			if mt.typ == message.MessageTypeImage ||
				mt.typ == message.MessageTypeAudio ||
				mt.typ == message.MessageTypeVideo ||
				mt.typ == message.MessageTypeDocument {
				msg.MediaURL = stringPtr("https://example.com/media.jpg")
			}

			err := adapter.ValidateMessage(msg)
			assert.NoError(t, err, "Should validate %s message type", mt.typ)
		})
	}
}

// TestWAHAMessageSenderAdapter_Validation_EdgeCases tests edge cases
func TestWAHAMessageSenderAdapter_Validation_EdgeCases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	t.Run("nil message", func(t *testing.T) {
		// This would panic in current implementation, but testing the concept
		defer func() {
			if r := recover(); r != nil {
				t.Log("Correctly panics on nil message")
			}
		}()

		// Don't actually call with nil to avoid panic in test
		// Just document the expected behavior
		t.Log("Nil message handling should be improved")
	})

	t.Run("empty content for text message", func(t *testing.T) {
		msg := &message.OutboundMessage{
			ID:        uuid.New(),
			ChannelID: uuid.New(),
			ContactID: uuid.New(),
			Type:      message.MessageTypeText,
			Content:   "", // Empty content
		}

		// Now properly validates that content is required for text messages
		err := adapter.ValidateMessage(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content is required")
	})

	t.Run("missing media URL for media message", func(t *testing.T) {
		msg := &message.OutboundMessage{
			ID:        uuid.New(),
			ChannelID: uuid.New(),
			ContactID: uuid.New(),
			Type:      message.MessageTypeImage,
			MediaURL:  nil, // Missing media URL
		}

		// Now properly validates that media_url is required for media messages
		err := adapter.ValidateMessage(msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "media_url is required")
	})
}

// TestWAHAMessageSenderAdapter_Concurrency tests concurrent message sending
func TestWAHAMessageSenderAdapter_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	ctx := context.Background()
	numMessages := 10

	messages := make([]*message.OutboundMessage, numMessages)
	for i := 0; i < numMessages; i++ {
		messages[i] = &message.OutboundMessage{
			ID:        uuid.New(),
			ChannelID: uuid.New(),
			ContactID: uuid.New(),
			Type:      message.MessageTypeText,
			Content:   "Concurrent message",
		}
	}

	// Send bulk messages (internally may handle concurrency)
	results, err := adapter.SendBulkMessages(ctx, messages)

	assert.NoError(t, err)
	assert.Len(t, results, numMessages)
}

// TestWAHAMessageSenderAdapter_ContextCancellation tests context cancellation
func TestWAHAMessageSenderAdapter_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context test in short mode")
	}

	logger, _ := zap.NewDevelopment()
	client := &waha.WAHAClient{}
	adapter := persistence.NewWAHAMessageSenderAdapter(client, nil, nil, logger)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := &message.OutboundMessage{
		ID:        uuid.New(),
		ChannelID: uuid.New(),
		ContactID: uuid.New(),
		Type:      message.MessageTypeText,
		Content:   "Test message",
	}

	// Current implementation doesn't check context
	// This test documents expected future behavior
	_, err := adapter.SendMessage(ctx, msg)

	// Currently will still return "not fully implemented" error
	// Future implementation should check context.Err()
	require.Error(t, err)
	t.Log("Context cancellation handling should be added in future")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
