package webhook

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/webhook"
	"github.com/google/uuid"
)

// CreateWebhookDTO representa os dados para criar um webhook
type CreateWebhookDTO struct {
	Name           string
	URL            string
	Events         []string
	Secret         string
	Headers        map[string]string
	RetryCount     int
	TimeoutSeconds int
}

// UpdateWebhookDTO representa os dados para atualizar um webhook
type UpdateWebhookDTO struct {
	Name           *string
	URL            *string
	Events         []string
	Active         *bool
	Secret         *string
	Headers        map[string]string
	RetryCount     *int
	TimeoutSeconds *int
}

// WebhookDTO representa um webhook para resposta
type WebhookDTO struct {
	ID              uuid.UUID         `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name            string            `json:"name" example:"N8N Webhook"`
	URL             string            `json:"url" example:"https://n8n.example.com/webhook/waha-events"`
	Events          []string          `json:"events" example:"message,ack,call.received"`
	Active          bool              `json:"active" example:"true"`
	Headers         map[string]string `json:"headers,omitempty" swaggertype:"object"`
	RetryCount      int               `json:"retry_count" example:"3"`
	TimeoutSeconds  int               `json:"timeout_seconds" example:"30"`
	LastTriggeredAt *time.Time        `json:"last_triggered_at,omitempty" example:"2024-01-01T00:00:00Z"`
	LastSuccessAt   *time.Time        `json:"last_success_at,omitempty" example:"2024-01-01T00:00:00Z"`
	LastFailureAt   *time.Time        `json:"last_failure_at,omitempty" example:"2024-01-01T00:00:00Z"`
	SuccessCount    int               `json:"success_count" example:"150"`
	FailureCount    int               `json:"failure_count" example:"2"`
	CreatedAt       time.Time         `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time         `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ToDTO converte entidade de domínio para DTO
func ToDTO(w *webhook.WebhookSubscription) WebhookDTO {
	return WebhookDTO{
		ID:              w.ID,
		Name:            w.Name,
		URL:             w.URL,
		Events:          w.Events,
		Active:          w.Active,
		Headers:         w.Headers,
		RetryCount:      w.RetryCount,
		TimeoutSeconds:  w.TimeoutSeconds,
		LastTriggeredAt: w.LastTriggeredAt,
		LastSuccessAt:   w.LastSuccessAt,
		LastFailureAt:   w.LastFailureAt,
		SuccessCount:    w.SuccessCount,
		FailureCount:    w.FailureCount,
		CreatedAt:       w.CreatedAt,
		UpdatedAt:       w.UpdatedAt,
	}
}

// ToDTOList converte lista de entidades para DTOs
func ToDTOList(webhooks []*webhook.WebhookSubscription) []WebhookDTO {
	dtos := make([]WebhookDTO, len(webhooks))
	for i, w := range webhooks {
		dtos[i] = ToDTO(w)
	}
	return dtos
}
