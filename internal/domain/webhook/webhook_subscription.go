package webhook

import (
	"time"

	"github.com/google/uuid"
)

type WebhookSubscription struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ProjectID uuid.UUID
	TenantID  string
	Name      string
	URL       string
	Events    []string

	SubscribeContactEvents bool
	ContactEventTypes      []string
	ContactEventCategories []string

	Active          bool
	Secret          string
	Headers         map[string]string
	RetryCount      int
	TimeoutSeconds  int
	LastTriggeredAt *time.Time
	LastSuccessAt   *time.Time
	LastFailureAt   *time.Time
	SuccessCount    int
	FailureCount    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewWebhookSubscription(userID, projectID uuid.UUID, tenantID, name, url string, events []string) (*WebhookSubscription, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if url == "" {
		return nil, ErrInvalidURL
	}
	if len(events) == 0 {
		return nil, ErrNoEvents
	}

	now := time.Now()
	return &WebhookSubscription{
		ID:             uuid.New(),
		UserID:         userID,
		ProjectID:      projectID,
		TenantID:       tenantID,
		Name:           name,
		URL:            url,
		Events:         events,
		Active:         true,
		RetryCount:     3,
		TimeoutSeconds: 30,
		SuccessCount:   0,
		FailureCount:   0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (w *WebhookSubscription) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	w.Name = name
	w.UpdatedAt = time.Now()
	return nil
}

func (w *WebhookSubscription) UpdateURL(url string) error {
	if url == "" {
		return ErrInvalidURL
	}
	w.URL = url
	w.UpdatedAt = time.Now()
	return nil
}

func (w *WebhookSubscription) UpdateEvents(events []string) error {
	if len(events) == 0 {
		return ErrNoEvents
	}
	w.Events = events
	w.UpdatedAt = time.Now()
	return nil
}

func (w *WebhookSubscription) SetActive() {
	w.Active = true
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) SetInactive() {
	w.Active = false
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) SetSecret(secret string) {
	w.Secret = secret
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) SetHeaders(headers map[string]string) {
	w.Headers = headers
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) SetRetryPolicy(retryCount, timeoutSeconds int) {
	w.RetryCount = retryCount
	w.TimeoutSeconds = timeoutSeconds
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) RecordTrigger(success bool) {
	now := time.Now()
	w.LastTriggeredAt = &now

	if success {
		w.LastSuccessAt = &now
		w.SuccessCount++
	} else {
		w.LastFailureAt = &now
		w.FailureCount++
	}

	w.UpdatedAt = now
}

func (w *WebhookSubscription) IsSubscribedTo(eventType string) bool {
	for _, event := range w.Events {
		if event == eventType || event == "*" {
			return true
		}

		if len(event) > 0 && event[len(event)-1] == '*' {
			prefix := event[:len(event)-1]
			if len(eventType) >= len(prefix) && eventType[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

func (w *WebhookSubscription) ShouldNotify(eventType string) bool {
	return w.Active && w.IsSubscribedTo(eventType)
}

func (w *WebhookSubscription) EnableContactEvents(eventTypes, categories []string) {
	w.SubscribeContactEvents = true
	w.ContactEventTypes = eventTypes
	w.ContactEventCategories = categories
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) DisableContactEvents() {
	w.SubscribeContactEvents = false
	w.ContactEventTypes = nil
	w.ContactEventCategories = nil
	w.UpdatedAt = time.Now()
}

func (w *WebhookSubscription) ShouldReceiveContactEvent(eventType, category string) bool {
	if !w.Active || !w.SubscribeContactEvents {
		return false
	}

	if len(w.ContactEventTypes) == 0 && len(w.ContactEventCategories) == 0 {
		return true
	}

	if len(w.ContactEventTypes) > 0 {
		for _, t := range w.ContactEventTypes {
			if t == eventType || t == "*" {
				return true
			}
		}
	}

	if len(w.ContactEventCategories) > 0 {
		for _, c := range w.ContactEventCategories {
			if c == category || c == "*" {
				return true
			}
		}
	}

	return false
}
