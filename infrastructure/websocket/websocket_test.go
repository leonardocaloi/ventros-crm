package websocket

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewWSMessage(t *testing.T) {
	payload := map[string]interface{}{
		"test": "data",
	}

	msg := NewWSMessage(MessageTypePing, payload)

	assert.Equal(t, MessageTypePing, msg.Type)
	assert.Equal(t, payload, msg.Payload)
	assert.NotEmpty(t, msg.MessageID)
	assert.NotEmpty(t, msg.Timestamp)
}

func TestNewErrorMessage(t *testing.T) {
	code := "test_error"
	message := "This is a test error"

	msg := NewErrorMessage(code, message)

	assert.Equal(t, MessageTypeError, msg.Type)
	assert.NotEmpty(t, msg.MessageID)
	assert.NotEmpty(t, msg.Timestamp)

	errorPayload, ok := msg.Payload.(ErrorPayload)
	assert.True(t, ok)
	assert.Equal(t, code, errorPayload.Code)
	assert.Equal(t, message, errorPayload.Message)
}

func TestWSMessage_ToJSON(t *testing.T) {
	msg := NewWSMessage(MessageTypePing, nil)

	jsonData, err := msg.ToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)
	assert.Contains(t, string(jsonData), "ping")
}

func TestFromJSON(t *testing.T) {
	jsonData := `{"type":"ping","timestamp":"2025-10-10T15:30:00Z","message_id":"msg-123"}`

	msg, err := FromJSON([]byte(jsonData))

	assert.NoError(t, err)
	assert.Equal(t, MessageTypePing, msg.Type)
	assert.Equal(t, "msg-123", msg.MessageID)
}

func TestWSMessage_ParsePayload(t *testing.T) {
	sendPayload := SendMessagePayload{
		SessionID:   uuid.New(),
		ContactID:   uuid.New(),
		Text:        "Test message",
		ContentType: "text",
	}

	msg := NewWSMessage(MessageTypeSendMessage, sendPayload)

	var parsed SendMessagePayload
	err := msg.ParsePayload(&parsed)

	assert.NoError(t, err)
	assert.Equal(t, sendPayload.Text, parsed.Text)
	assert.Equal(t, sendPayload.ContentType, parsed.ContentType)
}
