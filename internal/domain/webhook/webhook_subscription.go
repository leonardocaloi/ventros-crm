package webhook

import (
	"time"

	"github.com/google/uuid"
)

// WebhookSubscription é a entidade de domínio para inscrições de webhook
type WebhookSubscription struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	ProjectID       uuid.UUID
	TenantID        string
	Name            string
	URL             string
	Events          []string
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

// NewWebhookSubscription cria uma nova inscrição de webhook
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

// UpdateName atualiza o nome do webhook
func (w *WebhookSubscription) UpdateName(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	w.Name = name
	w.UpdatedAt = time.Now()
	return nil
}

// UpdateURL atualiza a URL do webhook
func (w *WebhookSubscription) UpdateURL(url string) error {
	if url == "" {
		return ErrInvalidURL
	}
	w.URL = url
	w.UpdatedAt = time.Now()
	return nil
}

// UpdateEvents atualiza os eventos do webhook
func (w *WebhookSubscription) UpdateEvents(events []string) error {
	if len(events) == 0 {
		return ErrNoEvents
	}
	w.Events = events
	w.UpdatedAt = time.Now()
	return nil
}

// SetActive ativa o webhook
func (w *WebhookSubscription) SetActive() {
	w.Active = true
	w.UpdatedAt = time.Now()
}

// SetInactive desativa o webhook
func (w *WebhookSubscription) SetInactive() {
	w.Active = false
	w.UpdatedAt = time.Now()
}

// SetSecret define o secret para HMAC
func (w *WebhookSubscription) SetSecret(secret string) {
	w.Secret = secret
	w.UpdatedAt = time.Now()
}

// SetHeaders define headers customizados
func (w *WebhookSubscription) SetHeaders(headers map[string]string) {
	w.Headers = headers
	w.UpdatedAt = time.Now()
}

// SetRetryPolicy define a política de retry
func (w *WebhookSubscription) SetRetryPolicy(retryCount, timeoutSeconds int) {
	w.RetryCount = retryCount
	w.TimeoutSeconds = timeoutSeconds
	w.UpdatedAt = time.Now()
}

// RecordTrigger registra que o webhook foi disparado
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

// IsSubscribedTo verifica se o webhook está inscrito em um evento
func (w *WebhookSubscription) IsSubscribedTo(eventType string) bool {
	for _, event := range w.Events {
		if event == eventType || event == "*" {
			return true
		}
		// Support wildcard matching (e.g., "message.*" matches "message.ack")
		if len(event) > 0 && event[len(event)-1] == '*' {
			prefix := event[:len(event)-1]
			if len(eventType) >= len(prefix) && eventType[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// ShouldNotify verifica se o webhook deve ser notificado
func (w *WebhookSubscription) ShouldNotify(eventType string) bool {
	return w.Active && w.IsSubscribedTo(eventType)
}
