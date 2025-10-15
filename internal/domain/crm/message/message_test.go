package message_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	domain "github.com/ventros/crm/internal/domain/core"
	"github.com/ventros/crm/internal/domain/crm/message"
)

// ===========================
// 1.4.1 - Testes de Factory Method
// ===========================

func TestNewMessage_Success(t *testing.T) {
	// Arrange
	contactID := domain.NewTestUUID()
	projectID := domain.NewTestUUID()
	customerID := domain.NewTestUUID()

	// Act
	msg, err := message.NewMessage(
		contactID,
		projectID,
		customerID,
		message.ContentTypeText,
		false, // inbound
	)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, msg.ID())
	assert.Equal(t, contactID, msg.ContactID())
	assert.Equal(t, projectID, msg.ProjectID())
	assert.Equal(t, customerID, msg.CustomerID())
	assert.Equal(t, message.ContentTypeText, msg.ContentType())
	assert.False(t, msg.FromMe())
	assert.Equal(t, message.StatusSent, msg.Status())
	domain.AssertTimeNotZero(t, msg.Timestamp(), "Timestamp")
}

func TestNewMessage_EmptyContactID(t *testing.T) {
	// Arrange
	contactID := uuid.Nil
	projectID := domain.NewTestUUID()
	customerID := domain.NewTestUUID()

	// Act
	msg, err := message.NewMessage(contactID, projectID, customerID, message.ContentTypeText, false)

	// Assert
	require.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "contactID cannot be nil")
}

func TestNewMessage_EmptyProjectID(t *testing.T) {
	// Arrange
	contactID := domain.NewTestUUID()
	projectID := uuid.Nil
	customerID := domain.NewTestUUID()

	// Act
	msg, err := message.NewMessage(contactID, projectID, customerID, message.ContentTypeText, false)

	// Assert
	require.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "projectID cannot be nil")
}

func TestNewMessage_InvalidContentType(t *testing.T) {
	// Arrange
	contactID := domain.NewTestUUID()
	projectID := domain.NewTestUUID()
	customerID := domain.NewTestUUID()
	invalidContentType := message.ContentType("invalid")

	// Act
	msg, err := message.NewMessage(contactID, projectID, customerID, invalidContentType, false)

	// Assert
	require.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "invalid content type")
}

func TestNewMessage_GeneratesEvent(t *testing.T) {
	// Arrange
	contactID := domain.NewTestUUID()
	projectID := domain.NewTestUUID()
	customerID := domain.NewTestUUID()

	// Act
	msg, err := message.NewMessage(contactID, projectID, customerID, message.ContentTypeText, false)

	// Assert
	require.NoError(t, err)
	events := msg.DomainEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(message.MessageCreatedEvent)
	require.True(t, ok, "Event should be MessageCreatedEvent")
	assert.Equal(t, msg.ID(), event.MessageID)
	assert.Equal(t, contactID, event.ContactID)
	assert.False(t, event.FromMe)
	domain.AssertTimeNotZero(t, event.CreatedAt, "Event CreatedAt")
}

// ===========================
// 1.4.2 - Testes de Content Type
// ===========================

func TestSetText_ValidTextMessage(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		false,
	)
	text := "Hello, world!"

	// Act
	err := msg.SetText(text)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, msg.Text())
	assert.Equal(t, text, *msg.Text())
}

func TestSetText_NonTextMessage(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeImage, // Media type, not text
		false,
	)

	// Act
	err := msg.SetText("test")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set text on non-text message")
}

func TestSetMediaContent_ValidMediaMessage(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeImage,
		false,
	)
	url := "https://example.com/image.jpg"
	mimetype := "image/jpeg"

	// Act
	err := msg.SetMediaContent(url, mimetype)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, msg.MediaURL())
	assert.Equal(t, url, *msg.MediaURL())
	require.NotNil(t, msg.MediaMimetype())
	assert.Equal(t, mimetype, *msg.MediaMimetype())
	assert.True(t, msg.HasMediaURL())
}

func TestSetMediaContent_NonMediaMessage(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText, // Text type, not media
		false,
	)

	// Act
	err := msg.SetMediaContent("https://example.com/file.jpg", "image/jpeg")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set media content on non-media message")
}

func TestContentType_IsValid(t *testing.T) {
	validTypes := []message.ContentType{
		message.ContentTypeText,
		message.ContentTypeImage,
		message.ContentTypeVideo,
		message.ContentTypeAudio,
		message.ContentTypeVoice,
		message.ContentTypeDocument,
		message.ContentTypeLocation,
		message.ContentTypeContact,
		message.ContentTypeSticker,
		message.ContentTypeSystem,
	}

	for _, ct := range validTypes {
		t.Run(string(ct), func(t *testing.T) {
			assert.True(t, ct.IsValid(), "%s should be valid", ct)
		})
	}

	// Invalid type
	invalidType := message.ContentType("invalid")
	assert.False(t, invalidType.IsValid())
}

func TestContentType_IsText(t *testing.T) {
	assert.True(t, message.ContentTypeText.IsText())
	assert.False(t, message.ContentTypeImage.IsText())
	assert.False(t, message.ContentTypeVideo.IsText())
}

func TestContentType_IsMedia(t *testing.T) {
	mediaTypes := []message.ContentType{
		message.ContentTypeImage,
		message.ContentTypeVideo,
		message.ContentTypeAudio,
		message.ContentTypeVoice,
		message.ContentTypeDocument,
		message.ContentTypeSticker,
	}

	for _, ct := range mediaTypes {
		t.Run(string(ct), func(t *testing.T) {
			assert.True(t, ct.IsMedia(), "%s should be media type", ct)
		})
	}

	// Non-media types
	assert.False(t, message.ContentTypeText.IsMedia())
	assert.False(t, message.ContentTypeLocation.IsMedia())
	assert.False(t, message.ContentTypeContact.IsMedia())
}

// ===========================
// 1.4.3 - Testes de Status Transitions
// ===========================

func TestMarkAsDelivered_Success(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true, // outbound
	)
	msg.ClearEvents() // Limpa evento de criação
	assert.Equal(t, message.StatusSent, msg.Status())

	// Act
	msg.MarkAsDelivered()

	// Assert
	assert.Equal(t, message.StatusDelivered, msg.Status())
}

func TestMarkAsDelivered_SetsTimestamp(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true,
	)
	assert.Nil(t, msg.DeliveredAt())

	// Act
	msg.MarkAsDelivered()

	// Assert
	require.NotNil(t, msg.DeliveredAt())
	domain.AssertTimeNotZero(t, *msg.DeliveredAt(), "DeliveredAt")
}

func TestMarkAsDelivered_GeneratesEvent(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true,
	)
	msg.ClearEvents()

	// Act
	msg.MarkAsDelivered()

	// Assert
	events := msg.DomainEvents()
	require.Len(t, events, 1)

	event, ok := events[0].(message.MessageDeliveredEvent)
	require.True(t, ok, "Event should be MessageDeliveredEvent")
	assert.Equal(t, msg.ID(), event.MessageID)
	domain.AssertTimeNotZero(t, event.DeliveredAt, "Event DeliveredAt")
}

func TestMarkAsRead_Success(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true,
	)
	assert.Equal(t, message.StatusSent, msg.Status())

	// Act
	msg.MarkAsRead()

	// Assert
	assert.Equal(t, message.StatusRead, msg.Status())
}

func TestMarkAsRead_SetsTimestamp(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true,
	)
	assert.Nil(t, msg.ReadAt())

	// Act
	msg.MarkAsRead()

	// Assert
	require.NotNil(t, msg.ReadAt())
	domain.AssertTimeNotZero(t, *msg.ReadAt(), "ReadAt")
}

func TestMarkAsFailed_Success(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true,
	)
	assert.Equal(t, message.StatusSent, msg.Status())

	// Act
	msg.MarkAsFailed()

	// Assert
	assert.Equal(t, message.StatusFailed, msg.Status())
}

// ===========================
// Additional Tests
// ===========================

func TestAssignToSession(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		false,
	)
	sessionID := domain.NewTestUUID()
	assert.Nil(t, msg.SessionID())

	// Act
	msg.AssignToSession(sessionID)

	// Assert
	require.NotNil(t, msg.SessionID())
	assert.Equal(t, sessionID, *msg.SessionID())
}

func TestAssignToChannel(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		false,
	)
	channelID := domain.NewTestUUID()
	channelTypeID := domain.IntPtr(1)

	// Act
	msg.AssignToChannel(channelID, channelTypeID)

	// Assert
	assert.Equal(t, channelID, msg.ChannelID())
	require.NotNil(t, msg.ChannelTypeID())
	assert.Equal(t, 1, *msg.ChannelTypeID())
}

func TestSetChannelMessageID(t *testing.T) {
	// Arrange
	msg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		false,
	)
	externalID := "wamid.123456"

	// Act
	msg.SetChannelMessageID(externalID)

	// Assert
	require.NotNil(t, msg.ChannelMessageID())
	assert.Equal(t, externalID, *msg.ChannelMessageID())
}

func TestIsInbound_IsOutbound(t *testing.T) {
	// Arrange
	inboundMsg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		false, // fromMe = false
	)
	outboundMsg, _ := message.NewMessage(
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		domain.NewTestUUID(),
		message.ContentTypeText,
		true, // fromMe = true
	)

	// Assert
	assert.True(t, inboundMsg.IsInbound())
	assert.False(t, inboundMsg.IsOutbound())

	assert.True(t, outboundMsg.IsOutbound())
	assert.False(t, outboundMsg.IsInbound())
}

func TestReconstructMessage(t *testing.T) {
	// Arrange
	id := domain.NewTestUUID()
	timestamp := time.Now()
	customerID := domain.NewTestUUID()
	projectID := domain.NewTestUUID()
	contactID := domain.NewTestUUID()
	sessionID := domain.NewTestUUID()
	channelID := domain.NewTestUUID()
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
		true, // fromMe
		channelID,
		contactID,
		&sessionID,
		message.ContentTypeText,
		&text,
		nil, // mediaURL
		nil, // mediaMimetype
		nil, // channelMessageID
		nil, // replyToID
		message.StatusRead,
		nil, // language
		nil, // agentID
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
	assert.Equal(t, channelID, msg.ChannelID())
	require.NotNil(t, msg.SessionID())
	assert.Equal(t, sessionID, *msg.SessionID())
	assert.True(t, msg.FromMe())
	assert.Equal(t, message.StatusRead, msg.Status())
	require.NotNil(t, msg.Text())
	assert.Equal(t, text, *msg.Text())
	require.NotNil(t, msg.DeliveredAt())
	require.NotNil(t, msg.ReadAt())

	// Metadata deve ser copiado
	retrievedMetadata := msg.Metadata()
	assert.Equal(t, "value", retrievedMetadata["key"])

	// Não deve ter eventos após reconstituir
	assert.Empty(t, msg.DomainEvents())
}
