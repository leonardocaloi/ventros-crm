package dto

import "github.com/google/uuid"

// CreateContactRequest representa o payload para criar um contato
type CreateContactRequest struct {
	Name          string                 `json:"name" binding:"required" example:"João Silva"`
	Email         string                 `json:"email,omitempty" example:"joao@example.com"`
	Phone         string                 `json:"phone,omitempty" example:"5511999999999"`
	ExternalID    string                 `json:"external_id,omitempty" example:"ext-123"`
	SourceChannel string                 `json:"source_channel,omitempty" example:"whatsapp"`
	Language      string                 `json:"language,omitempty" example:"pt-BR"`
	Timezone      string                 `json:"timezone,omitempty" example:"America/Sao_Paulo"`
	Tags          []string               `json:"tags,omitempty" example:"vip,lead"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
} //	@name	CreateContactRequest

// UpdateContactRequest representa o payload para atualizar um contato
type UpdateContactRequest struct {
	Name          *string                `json:"name,omitempty" example:"João Silva"`
	Email         *string                `json:"email,omitempty" example:"joao@example.com"`
	Phone         *string                `json:"phone,omitempty" example:"5511999999999"`
	SourceChannel *string                `json:"source_channel,omitempty" example:"whatsapp"`
	Language      *string                `json:"language,omitempty" example:"pt-BR"`
	Timezone      *string                `json:"timezone,omitempty" example:"America/Sao_Paulo"`
	Tags          *[]string              `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
} //	@name	UpdateContactRequest

// CreatePipelineRequest representa o payload para criar um pipeline
type CreatePipelineRequest struct {
	Name        string `json:"name" binding:"required" example:"Pipeline de Vendas"`
	Description string `json:"description,omitempty" example:"Pipeline principal de vendas"`
	Color       string `json:"color,omitempty" example:"#3B82F6"`
	Position    int    `json:"position,omitempty" example:"1"`
} //	@name	CreatePipelineRequest

// CreatePipelineStatusRequest representa o payload para criar um status de pipeline
type CreatePipelineStatusRequest struct {
	Name        string `json:"name" binding:"required" example:"Novo Lead"`
	Description string `json:"description,omitempty" example:"Contato recém chegado"`
	Color       string `json:"color,omitempty" example:"#10B981"`
	StatusType  string `json:"status_type" binding:"required" example:"open"`
	Position    int    `json:"position,omitempty" example:"1"`
} //	@name	CreatePipelineStatusRequest

// ChangeContactStatusRequest representa o payload para mudar status de um contato
type ChangeContactStatusRequest struct {
	StatusID uuid.UUID              `json:"status_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Notes    string                 `json:"notes,omitempty" example:"Cliente demonstrou interesse"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
} //	@name	ChangeContactStatusRequest

// CreateProjectRequest representa o payload para criar um projeto
type CreateProjectRequest struct {
	Name        string                 `json:"name" binding:"required" example:"Meu Projeto"`
	Description string                 `json:"description,omitempty" example:"Projeto de vendas"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
} //	@name	CreateProjectRequest

// CreateWebhookSubscriptionRequest representa o payload para criar um webhook
type CreateWebhookSubscriptionRequest struct {
	Name           string            `json:"name" binding:"required" example:"Webhook N8N"`
	URL            string            `json:"url" binding:"required" example:"https://webhook.site/unique-url"`
	Events         []string          `json:"events" binding:"required" example:"contact.created,session.started"`
	Secret         string            `json:"secret,omitempty" example:"my-secret-key"`
	Headers        map[string]string `json:"headers,omitempty"`
	RetryCount     int               `json:"retry_count,omitempty" example:"3"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty" example:"30"`
} //	@name	CreateWebhookSubscriptionRequest

// UpdateWebhookSubscriptionRequest representa o payload para atualizar um webhook
type UpdateWebhookSubscriptionRequest struct {
	Name           *string            `json:"name,omitempty" example:"Webhook N8N"`
	URL            *string            `json:"url,omitempty" example:"https://webhook.site/unique-url"`
	Events         *[]string          `json:"events,omitempty"`
	Active         *bool              `json:"active,omitempty" example:"true"`
	Secret         *string            `json:"secret,omitempty"`
	Headers        *map[string]string `json:"headers,omitempty"`
	RetryCount     *int               `json:"retry_count,omitempty" example:"3"`
	TimeoutSeconds *int               `json:"timeout_seconds,omitempty" example:"30"`
} //	@name	UpdateWebhookSubscriptionRequest

// CreateAgentRequest representa o payload para criar um agente
type CreateAgentRequest struct {
	Name     string                 `json:"name" binding:"required" example:"Maria Atendente"`
	Email    string                 `json:"email" binding:"required" example:"maria@ventros.com"`
	Role     string                 `json:"role,omitempty" example:"agent"`
	Status   string                 `json:"status,omitempty" example:"active"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
} //	@name	CreateAgentRequest

// UpdateAgentRequest representa o payload para atualizar um agente
type UpdateAgentRequest struct {
	Name     *string                `json:"name,omitempty" example:"Maria Atendente"`
	Email    *string                `json:"email,omitempty" example:"maria@ventros.com"`
	Role     *string                `json:"role,omitempty" example:"agent"`
	Status   *string                `json:"status,omitempty" example:"active"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
} //	@name	UpdateAgentRequest

// SendMessageRequest representa o payload para enviar uma mensagem
type SendMessageRequest struct {
	ContactID     uuid.UUID              `json:"contact_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ContentType   string                 `json:"content_type" binding:"required" example:"text"`
	Text          string                 `json:"text,omitempty" example:"Olá! Como posso ajudar?"`
	MediaURL      string                 `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	MediaMimetype string                 `json:"media_mimetype,omitempty" example:"image/jpeg"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
} //	@name	SendMessageRequest
