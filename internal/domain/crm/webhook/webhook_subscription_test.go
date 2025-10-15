package webhook

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookSubscription_Valid(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"
	name := "Production Webhook"
	url := "https://example.com/webhooks"
	events := []string{"contact.created", "message.received"}

	webhook, err := NewWebhookSubscription(userID, projectID, tenantID, name, url, events)

	require.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.NotEqual(t, uuid.Nil, webhook.ID)
	assert.Equal(t, userID, webhook.UserID)
	assert.Equal(t, projectID, webhook.ProjectID)
	assert.Equal(t, tenantID, webhook.TenantID)
	assert.Equal(t, name, webhook.Name)
	assert.Equal(t, url, webhook.URL)
	assert.Equal(t, events, webhook.Events)
	assert.True(t, webhook.Active)
	assert.Equal(t, 3, webhook.RetryCount)
	assert.Equal(t, 30, webhook.TimeoutSeconds)
	assert.Equal(t, 0, webhook.SuccessCount)
	assert.Equal(t, 0, webhook.FailureCount)
	assert.NotZero(t, webhook.CreatedAt)
	assert.NotZero(t, webhook.UpdatedAt)
}

func TestNewWebhookSubscription_Invalid(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	tenantID := "tenant-123"

	tests := []struct {
		name      string
		wName     string
		url       string
		events    []string
		expectErr error
	}{
		{
			name:      "empty name",
			wName:     "",
			url:       "https://example.com/webhooks",
			events:    []string{"contact.created"},
			expectErr: ErrInvalidName,
		},
		{
			name:      "empty URL",
			wName:     "Test Webhook",
			url:       "",
			events:    []string{"contact.created"},
			expectErr: ErrInvalidURL,
		},
		{
			name:      "no events",
			wName:     "Test Webhook",
			url:       "https://example.com/webhooks",
			events:    []string{},
			expectErr: ErrNoEvents,
		},
		{
			name:      "nil events",
			wName:     "Test Webhook",
			url:       "https://example.com/webhooks",
			events:    nil,
			expectErr: ErrNoEvents,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhook, err := NewWebhookSubscription(userID, projectID, tenantID, tt.wName, tt.url, tt.events)

			assert.Error(t, err)
			assert.Equal(t, tt.expectErr, err)
			assert.Nil(t, webhook)
		})
	}
}

func TestWebhookSubscription_UpdateName(t *testing.T) {
	webhook := createTestWebhook(t)

	t.Run("update with valid name", func(t *testing.T) {
		newName := "Updated Webhook Name"
		err := webhook.UpdateName(newName)

		require.NoError(t, err)
		assert.Equal(t, newName, webhook.Name)
	})

	t.Run("update with empty name fails", func(t *testing.T) {
		err := webhook.UpdateName("")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidName, err)
	})
}

func TestWebhookSubscription_UpdateURL(t *testing.T) {
	webhook := createTestWebhook(t)

	t.Run("update with valid URL", func(t *testing.T) {
		newURL := "https://new-domain.com/webhooks"
		err := webhook.UpdateURL(newURL)

		require.NoError(t, err)
		assert.Equal(t, newURL, webhook.URL)
	})

	t.Run("update with empty URL fails", func(t *testing.T) {
		err := webhook.UpdateURL("")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidURL, err)
	})
}

func TestWebhookSubscription_UpdateEvents(t *testing.T) {
	webhook := createTestWebhook(t)

	t.Run("update with valid events", func(t *testing.T) {
		newEvents := []string{"contact.updated", "session.ended", "message.sent"}
		err := webhook.UpdateEvents(newEvents)

		require.NoError(t, err)
		assert.Equal(t, newEvents, webhook.Events)
	})

	t.Run("update with empty events fails", func(t *testing.T) {
		err := webhook.UpdateEvents([]string{})

		assert.Error(t, err)
		assert.Equal(t, ErrNoEvents, err)
	})
}

func TestWebhookSubscription_SetActive_SetInactive(t *testing.T) {
	webhook := createTestWebhook(t)

	assert.True(t, webhook.Active)

	t.Run("set inactive", func(t *testing.T) {
		webhook.SetInactive()
		assert.False(t, webhook.Active)
	})

	t.Run("set active", func(t *testing.T) {
		webhook.SetActive()
		assert.True(t, webhook.Active)
	})
}

func TestWebhookSubscription_SetSecret(t *testing.T) {
	webhook := createTestWebhook(t)

	secret := "super-secret-hmac-key"
	webhook.SetSecret(secret)

	assert.Equal(t, secret, webhook.Secret)
}

func TestWebhookSubscription_SetHeaders(t *testing.T) {
	webhook := createTestWebhook(t)

	headers := map[string]string{
		"Authorization":   "Bearer token123",
		"X-Custom-Header": "custom-value",
		"Content-Type":    "application/json",
	}

	webhook.SetHeaders(headers)

	assert.Equal(t, headers, webhook.Headers)
	assert.Equal(t, "Bearer token123", webhook.Headers["Authorization"])
}

func TestWebhookSubscription_SetRetryPolicy(t *testing.T) {
	webhook := createTestWebhook(t)

	assert.Equal(t, 3, webhook.RetryCount)
	assert.Equal(t, 30, webhook.TimeoutSeconds)

	webhook.SetRetryPolicy(5, 60)

	assert.Equal(t, 5, webhook.RetryCount)
	assert.Equal(t, 60, webhook.TimeoutSeconds)
}

func TestWebhookSubscription_RecordTrigger(t *testing.T) {
	webhook := createTestWebhook(t)

	assert.Nil(t, webhook.LastTriggeredAt)
	assert.Nil(t, webhook.LastSuccessAt)
	assert.Nil(t, webhook.LastFailureAt)
	assert.Equal(t, 0, webhook.SuccessCount)
	assert.Equal(t, 0, webhook.FailureCount)

	t.Run("record successful trigger", func(t *testing.T) {
		webhook.RecordTrigger(true)

		assert.NotNil(t, webhook.LastTriggeredAt)
		assert.NotNil(t, webhook.LastSuccessAt)
		assert.Nil(t, webhook.LastFailureAt)
		assert.Equal(t, 1, webhook.SuccessCount)
		assert.Equal(t, 0, webhook.FailureCount)
	})

	t.Run("record failed trigger", func(t *testing.T) {
		webhook.RecordTrigger(false)

		assert.NotNil(t, webhook.LastTriggeredAt)
		assert.NotNil(t, webhook.LastFailureAt)
		assert.Equal(t, 1, webhook.SuccessCount)
		assert.Equal(t, 1, webhook.FailureCount)
	})

	t.Run("record multiple triggers", func(t *testing.T) {
		webhook.RecordTrigger(true)
		webhook.RecordTrigger(true)
		webhook.RecordTrigger(false)

		assert.Equal(t, 3, webhook.SuccessCount)
		assert.Equal(t, 2, webhook.FailureCount)
	})
}

func TestWebhookSubscription_IsSubscribedTo(t *testing.T) {
	tests := []struct {
		name          string
		events        []string
		testEvent     string
		expectMatches bool
	}{
		{
			name:          "exact match",
			events:        []string{"contact.created", "message.received"},
			testEvent:     "contact.created",
			expectMatches: true,
		},
		{
			name:          "no match",
			events:        []string{"contact.created", "message.received"},
			testEvent:     "session.ended",
			expectMatches: false,
		},
		{
			name:          "wildcard all events",
			events:        []string{"*"},
			testEvent:     "any.event.type",
			expectMatches: true,
		},
		{
			name:          "prefix wildcard match",
			events:        []string{"message.*"},
			testEvent:     "message.received",
			expectMatches: true,
		},
		{
			name:          "prefix wildcard match nested",
			events:        []string{"message.*"},
			testEvent:     "message.status.updated",
			expectMatches: true,
		},
		{
			name:          "prefix wildcard no match",
			events:        []string{"message.*"},
			testEvent:     "contact.created",
			expectMatches: false,
		},
		{
			name:          "multiple wildcards",
			events:        []string{"contact.*", "message.*"},
			testEvent:     "message.sent",
			expectMatches: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhook := createTestWebhookWithEvents(t, tt.events)

			result := webhook.IsSubscribedTo(tt.testEvent)

			assert.Equal(t, tt.expectMatches, result)
		})
	}
}

func TestWebhookSubscription_ShouldNotify(t *testing.T) {
	t.Run("active webhook subscribed to event", func(t *testing.T) {
		webhook := createTestWebhookWithEvents(t, []string{"contact.created"})
		assert.True(t, webhook.ShouldNotify("contact.created"))
	})

	t.Run("active webhook not subscribed to event", func(t *testing.T) {
		webhook := createTestWebhookWithEvents(t, []string{"contact.created"})
		assert.False(t, webhook.ShouldNotify("message.received"))
	})

	t.Run("inactive webhook", func(t *testing.T) {
		webhook := createTestWebhookWithEvents(t, []string{"contact.created"})
		webhook.SetInactive()
		assert.False(t, webhook.ShouldNotify("contact.created"))
	})
}

func TestWebhookSubscription_ContactEvents(t *testing.T) {
	webhook := createTestWebhook(t)

	assert.False(t, webhook.SubscribeContactEvents)

	t.Run("enable contact events", func(t *testing.T) {
		eventTypes := []string{"contact_created", "session_started"}
		categories := []string{"system", "session"}

		webhook.EnableContactEvents(eventTypes, categories)

		assert.True(t, webhook.SubscribeContactEvents)
		assert.Equal(t, eventTypes, webhook.ContactEventTypes)
		assert.Equal(t, categories, webhook.ContactEventCategories)
	})

	t.Run("disable contact events", func(t *testing.T) {
		webhook.DisableContactEvents()

		assert.False(t, webhook.SubscribeContactEvents)
		assert.Nil(t, webhook.ContactEventTypes)
		assert.Nil(t, webhook.ContactEventCategories)
	})
}

func TestWebhookSubscription_ShouldReceiveContactEvent(t *testing.T) {
	t.Run("inactive webhook doesn't receive events", func(t *testing.T) {
		webhook := createTestWebhook(t)
		webhook.EnableContactEvents([]string{"contact_created"}, nil)
		webhook.SetInactive()

		assert.False(t, webhook.ShouldReceiveContactEvent("contact_created", "system"))
	})

	t.Run("webhook not subscribed to contact events", func(t *testing.T) {
		webhook := createTestWebhook(t)

		assert.False(t, webhook.ShouldReceiveContactEvent("contact_created", "system"))
	})

	t.Run("receive all events when no filters", func(t *testing.T) {
		webhook := createTestWebhook(t)
		webhook.EnableContactEvents(nil, nil)

		assert.True(t, webhook.ShouldReceiveContactEvent("contact_created", "system"))
		assert.True(t, webhook.ShouldReceiveContactEvent("session_started", "session"))
		assert.True(t, webhook.ShouldReceiveContactEvent("note_added", "note"))
	})

	t.Run("filter by event type", func(t *testing.T) {
		webhook := createTestWebhook(t)
		webhook.EnableContactEvents([]string{"contact_created", "session_started"}, nil)

		assert.True(t, webhook.ShouldReceiveContactEvent("contact_created", "system"))
		assert.True(t, webhook.ShouldReceiveContactEvent("session_started", "session"))
		assert.False(t, webhook.ShouldReceiveContactEvent("note_added", "note"))
	})

	t.Run("filter by category", func(t *testing.T) {
		webhook := createTestWebhook(t)
		webhook.EnableContactEvents(nil, []string{"system", "session"})

		assert.True(t, webhook.ShouldReceiveContactEvent("contact_created", "system"))
		assert.True(t, webhook.ShouldReceiveContactEvent("session_started", "session"))
		assert.False(t, webhook.ShouldReceiveContactEvent("note_added", "note"))
	})

	t.Run("wildcard event type", func(t *testing.T) {
		webhook := createTestWebhook(t)
		webhook.EnableContactEvents([]string{"*"}, nil)

		assert.True(t, webhook.ShouldReceiveContactEvent("any_event", "any_category"))
	})

	t.Run("wildcard category", func(t *testing.T) {
		webhook := createTestWebhook(t)
		webhook.EnableContactEvents(nil, []string{"*"})

		assert.True(t, webhook.ShouldReceiveContactEvent("any_event", "any_category"))
	})
}

func TestWebhookSubscription_UpdatesTimestamp(t *testing.T) {
	webhook := createTestWebhook(t)
	originalUpdatedAt := webhook.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name   string
		action func()
	}{
		{
			name: "UpdateName",
			action: func() {
				webhook.UpdateName("New Name")
			},
		},
		{
			name: "UpdateURL",
			action: func() {
				webhook.UpdateURL("https://new-url.com")
			},
		},
		{
			name: "UpdateEvents",
			action: func() {
				webhook.UpdateEvents([]string{"new.event"})
			},
		},
		{
			name: "SetActive",
			action: func() {
				webhook.SetActive()
			},
		},
		{
			name: "SetInactive",
			action: func() {
				webhook.SetInactive()
			},
		},
		{
			name: "SetSecret",
			action: func() {
				webhook.SetSecret("new-secret")
			},
		},
		{
			name: "SetHeaders",
			action: func() {
				webhook.SetHeaders(map[string]string{"key": "value"})
			},
		},
		{
			name: "SetRetryPolicy",
			action: func() {
				webhook.SetRetryPolicy(10, 120)
			},
		},
		{
			name: "RecordTrigger",
			action: func() {
				webhook.RecordTrigger(true)
			},
		},
		{
			name: "EnableContactEvents",
			action: func() {
				webhook.EnableContactEvents([]string{"event"}, []string{"category"})
			},
		},
		{
			name: "DisableContactEvents",
			action: func() {
				webhook.DisableContactEvents()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeUpdate := webhook.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			tt.action()

			assert.True(t, webhook.UpdatedAt.After(beforeUpdate),
				"UpdatedAt should be updated after %s", tt.name)
		})
	}

	assert.True(t, webhook.UpdatedAt.After(originalUpdatedAt))
}

// Helper functions
func createTestWebhook(t *testing.T) *WebhookSubscription {
	return createTestWebhookWithEvents(t, []string{"contact.created", "message.received"})
}

func createTestWebhookWithEvents(t *testing.T, events []string) *WebhookSubscription {
	webhook, err := NewWebhookSubscription(
		uuid.New(),
		uuid.New(),
		"tenant-123",
		"Test Webhook",
		"https://example.com/webhooks",
		events,
	)
	require.NoError(t, err)
	return webhook
}
