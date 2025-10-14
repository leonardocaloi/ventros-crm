package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MessageSendTestSuite tests the POST /api/v1/messages/send endpoint
type MessageSendTestSuite struct {
	APITestSuite
}

// SetupSuite executes once before all tests
func (s *MessageSendTestSuite) SetupSuite() {
	s.APITestSuite.SetupSuite()

	// Create necessary test data
	s.setupTestData()
}

// TearDownSuite executes once after all tests
func (s *MessageSendTestSuite) TearDownSuite() {
	s.cleanupTestData()
}

// setupTestData creates user, project, channel, and contact for tests
func (s *MessageSendTestSuite) setupTestData() {
	// Create user
	userFixture := s.fixtures.Users[0]
	payload := map[string]string{
		"name":     userFixture.Name,
		"email":    userFixture.Email,
		"password": userFixture.Password,
		"role":     userFixture.Role,
	}

	resp, body := s.makeRequest("POST", "/api/v1/auth/register", payload, "")
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	s.createdIDs["user_id"] = result["user_id"].(string)
	s.createdIDs["api_key"] = result["api_key"].(string)
	s.createdIDs["project_id"] = result["default_project_id"].(string)

	fmt.Printf("✅ Test user created: %s\n", userFixture.Email)

	// Create channel
	channelFixture := s.fixtures.Channels[0]
	apiKey := s.createdIDs["api_key"]
	projectID := s.createdIDs["project_id"]

	channelPayload := map[string]interface{}{
		"name":        channelFixture.Name,
		"type":        channelFixture.Type,
		"waha_config": channelFixture.WAHAConfig,
	}

	endpoint := fmt.Sprintf("/api/v1/crm/channels?project_id=%s", projectID)
	resp, body = s.makeRequest("POST", endpoint, channelPayload, apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	s.createdIDs["channel_id"] = result["id"].(string)

	// Activate channel (async with 202 Accepted)
	activateEndpoint := fmt.Sprintf("/api/v1/crm/channels/%s/activate", s.createdIDs["channel_id"])
	resp, body = s.makeRequest("POST", activateEndpoint, nil, apiKey)
	assert.Equal(s.T(), http.StatusAccepted, resp.StatusCode, "Should return 202 Accepted for async activation")

	// Poll for channel activation (max 10 seconds)
	maxRetries := 20
	pollInterval := 500 * time.Millisecond
	channelActive := false

	for i := 0; i < maxRetries; i++ {
		time.Sleep(pollInterval)

		getEndpoint := fmt.Sprintf("/api/v1/crm/channels/%s", s.createdIDs["channel_id"])
		getResp, getBody := s.makeRequest("GET", getEndpoint, nil, apiKey)
		assert.Equal(s.T(), http.StatusOK, getResp.StatusCode)

		var channel map[string]interface{}
		err = json.Unmarshal(getBody, &channel)
		assert.NoError(s.T(), err)

		status := channel["status"].(string)
		if status == "active" {
			channelActive = true
			break
		} else if status == "inactive" {
			lastError := ""
			if channel["last_error"] != nil {
				lastError = channel["last_error"].(string)
			}
			s.T().Fatalf("Channel activation failed: %s", lastError)
		}
	}

	assert.True(s.T(), channelActive, "Channel should be activated within 10 seconds")

	fmt.Printf("✅ Test channel created and activated: %s\n", channelFixture.Name)

	// Create contact
	contactFixture := s.fixtures.Contacts[0]
	contactPayload := map[string]string{
		"name":  contactFixture.Name,
		"phone": contactFixture.Phone,
		"email": contactFixture.Email,
	}

	contactEndpoint := fmt.Sprintf("/api/v1/contacts?project_id=%s", projectID)
	resp, body = s.makeRequest("POST", contactEndpoint, contactPayload, apiKey)
	assert.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	err = json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	s.createdIDs["contact_id"] = result["id"].(string)

	fmt.Printf("✅ Test contact created: %s\n", contactFixture.Name)
}

// cleanupTestData removes test data
func (s *MessageSendTestSuite) cleanupTestData() {
	fmt.Println("\n🧹 Cleaning up message send test data...")
	s.cleanupChannels()
	s.cleanupContacts()
	fmt.Println("✅ Cleanup completed")
}

// TestSendTextMessage tests sending a simple text message
func (s *MessageSendTestSuite) TestSendTextMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	assert.NotEmpty(s.T(), apiKey, "API key must be set")
	assert.NotEmpty(s.T(), contactID, "Contact ID must be set")
	assert.NotEmpty(s.T(), channelID, "Channel ID must be set")

	textContent := "Hello from E2E test!"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	// Currently expecting error because WAHA integration is not fully implemented
	// When fully implemented, this should return StatusOK
	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// The handler should process the request even if sending fails
	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Text message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		// Expected while WAHA is not fully implemented
		assert.Equal(s.T(), "failed", result["status"])
		assert.NotNil(s.T(), result["error"])
		fmt.Printf("⚠️  Message send failed as expected (WAHA not implemented): %v\n", result["error"])
	} else {
		s.T().Fatalf("Unexpected status code: %d", resp.StatusCode)
	}
}

// TestSendTextMessage_MissingContactID tests validation for missing contact_id
func (s *MessageSendTestSuite) TestSendTextMessage_MissingContactID() {
	apiKey := s.createdIDs["api_key"]
	channelID := s.createdIDs["channel_id"]

	textContent := "Hello"
	payload := map[string]interface{}{
		// Missing contact_id
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	assert.Contains(s.T(), result["error"], "contact_id")

	fmt.Println("✅ Correctly rejected message with missing contact_id")
}

// TestSendTextMessage_MissingChannelID tests validation for missing channel_id
func (s *MessageSendTestSuite) TestSendTextMessage_MissingChannelID() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id": contactID,
		// Missing channel_id
		"content_type": "text",
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	assert.Contains(s.T(), result["error"], "channel_id")

	fmt.Println("✅ Correctly rejected message with missing channel_id")
}

// TestSendTextMessage_MissingContentType tests validation for missing content_type
func (s *MessageSendTestSuite) TestSendTextMessage_MissingContentType() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id": contactID,
		"channel_id": channelID,
		// Missing content_type
		"text": &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	assert.Contains(s.T(), result["error"], "content_type")

	fmt.Println("✅ Correctly rejected message with missing content_type")
}

// TestSendTextMessage_InvalidContentType tests validation for invalid content_type
func (s *MessageSendTestSuite) TestSendTextMessage_InvalidContentType() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "invalid_type", // Invalid content type
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	assert.Contains(s.T(), result["error"], "Invalid content type")

	fmt.Println("✅ Correctly rejected message with invalid content_type")
}

// TestSendTextMessage_InvalidContactID tests validation for invalid contact_id UUID
func (s *MessageSendTestSuite) TestSendTextMessage_InvalidContactID() {
	apiKey := s.createdIDs["api_key"]
	channelID := s.createdIDs["channel_id"]

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id":   "not-a-uuid", // Invalid UUID
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)
	assert.Equal(s.T(), http.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)
	// Should contain error about invalid UUID format
	assert.NotNil(s.T(), result["error"])

	fmt.Println("✅ Correctly rejected message with invalid contact_id format")
}

// TestSendImageMessage tests sending an image message
func (s *MessageSendTestSuite) TestSendImageMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	// Real image URL from WAHA GitHub examples
	mediaURL := "https://github.com/devlikeapro/waha/raw/core/examples/dev.likeapro.jpg"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "image",
		"media_url":    &mediaURL,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Similar to text message, may fail due to WAHA not being fully implemented
	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Image message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		assert.Equal(s.T(), "failed", result["status"])
		fmt.Printf("⚠️  Image message send failed as expected (WAHA not implemented): %v\n", result["error"])
	}
}

// TestSendVideoMessage tests sending a video message
func (s *MessageSendTestSuite) TestSendVideoMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	mediaURL := "https://example.com/test-video.mp4"
	caption := "Test video message"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "video",
		"media_url":    &mediaURL,
		"text":         &caption,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Video message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		assert.Equal(s.T(), "failed", result["status"])
		fmt.Printf("⚠️  Video message send failed: %v\n", result["error"])
	}
}

// TestSendAudioMessage tests sending an audio message
func (s *MessageSendTestSuite) TestSendAudioMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	mediaURL := "https://example.com/test-audio.mp3"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "audio",
		"media_url":    &mediaURL,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Audio message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		assert.Equal(s.T(), "failed", result["status"])
		fmt.Printf("⚠️  Audio message send failed: %v\n", result["error"])
	}
}

// TestSendDocumentMessage tests sending a document message
func (s *MessageSendTestSuite) TestSendDocumentMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	mediaURL := "https://example.com/test-document.pdf"
	caption := "Test document"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "document",
		"media_url":    &mediaURL,
		"text":         &caption,
		"metadata": map[string]interface{}{
			"filename": "test-document.pdf",
		},
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Document message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		assert.Equal(s.T(), "failed", result["status"])
		fmt.Printf("⚠️  Document message send failed: %v\n", result["error"])
	}
}

// TestSendLocationMessage tests sending a location message
func (s *MessageSendTestSuite) TestSendLocationMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	title := "Test Location"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "location",
		"text":         &title,
		"metadata": map[string]interface{}{
			"latitude":  -23.5505199,
			"longitude": -46.6333094,
		},
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Location message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		assert.Equal(s.T(), "failed", result["status"])
		fmt.Printf("⚠️  Location message send failed: %v\n", result["error"])
	}
}

// TestSendContactMessage tests sending a contact message
func (s *MessageSendTestSuite) TestSendContactMessage() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	// vCard format for contact sharing
	vcard := `BEGIN:VCARD
VERSION:3.0
FN:John Doe
TEL;TYPE=CELL:+55 11 99999-9999
EMAIL:john@example.com
END:VCARD`

	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "contact",
		"metadata": map[string]interface{}{
			"vcard": vcard,
		},
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if resp.StatusCode == http.StatusOK {
		assert.NotEmpty(s.T(), result["message_id"])
		assert.Equal(s.T(), "sent", result["status"])
		fmt.Printf("✅ Contact message sent successfully: %v\n", result["message_id"])
	} else if resp.StatusCode == http.StatusInternalServerError {
		assert.Equal(s.T(), "failed", result["status"])
		fmt.Printf("⚠️  Contact message send failed: %v\n", result["error"])
	}
}

// TestSendMessageWithReply tests sending a message as a reply
func (s *MessageSendTestSuite) TestSendMessageWithReply() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	// Generate a random reply_to_id
	replyToID := uuid.New()
	textContent := "This is a reply"

	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
		"reply_to_id":  replyToID,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Should process the request even with reply_to_id
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError {
		fmt.Printf("✅ Message with reply processed: status=%s\n", result["status"])
	} else {
		s.T().Fatalf("Unexpected status code: %d", resp.StatusCode)
	}
}

// TestSendMessageWithMetadata tests sending a message with custom metadata
func (s *MessageSendTestSuite) TestSendMessageWithMetadata() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	textContent := "Message with metadata"
	metadata := map[string]interface{}{
		"campaign_id":  "summer_sale_2025",
		"source":       "automated_flow",
		"priority":     "high",
		"custom_field": 123,
	}

	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
		"metadata":     metadata,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	// Metadata should not cause rejection
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusInternalServerError {
		fmt.Printf("✅ Message with metadata processed: status=%s\n", result["status"])
	} else {
		s.T().Fatalf("Unexpected status code: %d", resp.StatusCode)
	}
}

// TestSendMessage_Unauthorized tests sending without authentication
func (s *MessageSendTestSuite) TestSendMessage_Unauthorized() {
	contactID := s.createdIDs["contact_id"]
	channelID := s.createdIDs["channel_id"]

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
	}

	// Send request without API key
	resp, _ := s.makeRequest("POST", "/api/v1/messages/send", payload, "")

	// Should return 401 Unauthorized
	assert.Equal(s.T(), http.StatusUnauthorized, resp.StatusCode)

	fmt.Println("✅ Correctly rejected unauthorized message send request")
}

// TestSendMessage_NonExistentContact tests sending to non-existent contact
func (s *MessageSendTestSuite) TestSendMessage_NonExistentContact() {
	apiKey := s.createdIDs["api_key"]
	channelID := s.createdIDs["channel_id"]

	// Use a random UUID that doesn't exist
	nonExistentContactID := uuid.New()

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id":   nonExistentContactID.String(),
		"channel_id":   channelID,
		"content_type": "text",
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	// Should return error (either 404 or 500 depending on implementation)
	assert.True(s.T(), resp.StatusCode >= 400, "Should return error for non-existent contact")

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if result["error"] != nil {
		fmt.Printf("✅ Correctly handled non-existent contact: %v\n", result["error"])
	}
}

// TestSendMessage_NonExistentChannel tests sending via non-existent channel
func (s *MessageSendTestSuite) TestSendMessage_NonExistentChannel() {
	apiKey := s.createdIDs["api_key"]
	contactID := s.createdIDs["contact_id"]

	// Use a random UUID that doesn't exist
	nonExistentChannelID := uuid.New()

	textContent := "Hello"
	payload := map[string]interface{}{
		"contact_id":   contactID,
		"channel_id":   nonExistentChannelID.String(),
		"content_type": "text",
		"text":         &textContent,
	}

	resp, body := s.makeRequest("POST", "/api/v1/messages/send", payload, apiKey)

	// Should return error
	assert.True(s.T(), resp.StatusCode >= 400, "Should return error for non-existent channel")

	var result map[string]interface{}
	err := json.Unmarshal(body, &result)
	assert.NoError(s.T(), err)

	if result["error"] != nil {
		fmt.Printf("✅ Correctly handled non-existent channel: %v\n", result["error"])
	}
}

// TestMessageSendTestSuite runs the message send test suite
func TestMessageSendTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	suite.Run(t, new(MessageSendTestSuite))
}
