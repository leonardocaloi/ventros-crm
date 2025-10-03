package message_test

import (
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessage_Success(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	projectID := uuid.New()
	customerID := uuid.New()

	// Act
	msg, err := message.NewMessage(
		contactID,
		projectID,
		customerID,
		message.MessageTypeText,
		false, // inbound
	)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, msg.ID())
	assert.Equal(t, contactID, msg.ContactID())
	assert.Equal(t, projectID, msg.ProjectID())
	assert.Equal(t, customerID, msg.CustomerID())
	assert.Equal(t, message.MessageTypeText, msg.Type())
	assert.False(t, msg.FromMe())
	assert.Equal(t, message.StatusSent, msg.Status())
	
	// Deve ter evento MessageCreated
	events := msg.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "message.created", events[0].EventName())
}

func TestNewMessage_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		contactID  uuid.UUID
		projectID  uuid.UUID
		customerID uuid.UUID
		wantErr    string
	}{
		{
			name:       "nil contact ID",
			contactID:  uuid.Nil,
			projectID:  uuid.New(),
			customerID: uuid.New(),
			wantErr:    "contactID cannot be nil",
		},
		{
			name:       "nil project ID",
			contactID:  uuid.New(),
			projectID:  uuid.Nil,
			customerID: uuid.New(),
			wantErr:    "projectID cannot be nil",
		},
		{
			name:       "nil customer ID",
			contactID:  uuid.New(),
			projectID:  uuid.New(),
			customerID: uuid.Nil,
			wantErr:    "customerID cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := message.NewMessage(
				tt.contactID,
				tt.projectID,
				tt.customerID,
				message.MessageTypeText,
				false,
			)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Nil(t, msg)
		})
	}
}

func TestMessage_SetText_Success(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, false)
	text := "Hello, world!"

	// Act
	err := msg.SetText(text)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, msg.Text())
	assert.Equal(t, text, *msg.Text())
}

func TestMessage_SetText_OnMediaMessage_Error(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeMedia, false)

	// Act
	err := msg.SetText("test")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set text on non-text message")
}

func TestMessage_SetMedia_Success(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeMedia, false)

	// Act
	err := msg.SetMedia(message.MediaTypeImage, "image/jpeg", "https://example.com/image.jpg")

	// Assert
	require.NoError(t, err)
	assert.True(t, msg.HasMedia())
	assert.NotNil(t, msg.MediaType())
	assert.Equal(t, "image", *msg.MediaType())
	assert.NotNil(t, msg.MediaMimetype())
	assert.Equal(t, "image/jpeg", *msg.MediaMimetype())
	assert.NotNil(t, msg.MediaURL())
	assert.Equal(t, "https://example.com/image.jpg", *msg.MediaURL())
	
	// Verifica GetMediaType
	mediaType, err := msg.GetMediaType()
	require.NoError(t, err)
	assert.Equal(t, message.MediaTypeImage, mediaType)
}

func TestMessage_SetMedia_OnTextMessage_Error(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, false)

	// Act
	err := msg.SetMedia(message.MediaTypeImage, "image/jpeg", "https://example.com/image.jpg")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set media on non-media message")
}

func TestMessage_MediaTypes(t *testing.T) {
	tests := []struct {
		name      string
		mediaType message.MediaType
	}{
		{"image", message.MediaTypeImage},
		{"video", message.MediaTypeVideo},
		{"audio", message.MediaTypeAudio},
		{"document", message.MediaTypeDocument},
		{"sticker", message.MediaTypeSticker},
		{"location", message.MediaTypeLocation},
		{"contact", message.MediaTypeContact},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeMedia, false)

			// Act
			err := msg.SetMedia(tt.mediaType, "test/mimetype", "https://example.com/file")

			// Assert
			require.NoError(t, err)
			assert.True(t, msg.HasMedia())
			
			retrievedType, err := msg.GetMediaType()
			require.NoError(t, err)
			assert.Equal(t, tt.mediaType, retrievedType)
		})
	}
}

func TestMessage_AssignToSession(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, false)
	sessionID := uuid.New()

	// Act
	msg.AssignToSession(sessionID)

	// Assert
	assert.NotNil(t, msg.SessionID())
	assert.Equal(t, sessionID, *msg.SessionID())
}

func TestMessage_MarkAsDelivered(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, true)
	msg.ClearEvents() // Limpa evento de criação

	// Act
	msg.MarkAsDelivered()

	// Assert
	assert.Equal(t, message.StatusDelivered, msg.Status())
	assert.NotNil(t, msg.DeliveredAt())
	
	// Deve emitir evento
	events := msg.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "message.delivered", events[0].EventName())
}

func TestMessage_MarkAsRead(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, true)
	msg.ClearEvents()

	// Act
	msg.MarkAsRead()

	// Assert
	assert.Equal(t, message.StatusRead, msg.Status())
	assert.NotNil(t, msg.ReadAt())
	
	// Deve emitir evento
	events := msg.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "message.read", events[0].EventName())
}

func TestMessage_MarkAsFailed(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, true)

	// Act
	msg.MarkAsFailed()

	// Assert
	assert.Equal(t, message.StatusFailed, msg.Status())
}

func TestMessage_IsInbound_IsOutbound(t *testing.T) {
	// Arrange
	inboundMsg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, false)
	outboundMsg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, true)

	// Assert
	assert.True(t, inboundMsg.IsInbound())
	assert.False(t, inboundMsg.IsOutbound())
	
	assert.True(t, outboundMsg.IsOutbound())
	assert.False(t, outboundMsg.IsInbound())
}

func TestReconstructMessage(t *testing.T) {
	// Arrange
	id := uuid.New()
	timestamp := time.Now()
	customerID := uuid.New()
	projectID := uuid.New()
	contactID := uuid.New()
	sessionID := uuid.New()
	channelTypeID := 1
	text := "Test message"
	deliveredAt := time.Now()
	readAt := time.Now()
	metadata := map[string]interface{}{"key": "value"}

	// Act
	msg := message.ReconstructMessage(
		id,
		timestamp,
		customerID,
		projectID,
		&channelTypeID,
		true,
		nil,
		contactID,
		&sessionID,
		message.MessageTypeText,
		&text,
		nil,
		nil,
		nil,
		nil,
		nil,
		message.StatusRead,
		nil,
		nil,
		metadata,
		&deliveredAt,
		&readAt,
	)

	// Assert
	assert.Equal(t, id, msg.ID())
	assert.Equal(t, timestamp, msg.Timestamp())
	assert.Equal(t, customerID, msg.CustomerID())
	assert.Equal(t, projectID, msg.ProjectID())
	assert.Equal(t, contactID, msg.ContactID())
	assert.NotNil(t, msg.SessionID())
	assert.Equal(t, sessionID, *msg.SessionID())
	assert.True(t, msg.FromMe())
	assert.Equal(t, message.StatusRead, msg.Status())
	assert.NotNil(t, msg.Text())
	assert.Equal(t, text, *msg.Text())
	assert.NotNil(t, msg.DeliveredAt())
	assert.NotNil(t, msg.ReadAt())
	
	// Metadata deve ser copiado
	retrievedMetadata := msg.Metadata()
	assert.Equal(t, "value", retrievedMetadata["key"])
	
	// Não deve ter eventos após reconstituir
	assert.Empty(t, msg.DomainEvents())
}

func TestMessage_ClearEvents(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(uuid.New(), uuid.New(), uuid.New(), message.MessageTypeText, false)
	assert.Len(t, msg.DomainEvents(), 1)

	// Act
	msg.ClearEvents()

	// Assert
	assert.Empty(t, msg.DomainEvents())
}
