package dto

import (
	"time"

	"github.com/google/uuid"
)

// ContactResponse representa um contato na API
type ContactResponse struct {
	ID                   uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProjectID            uuid.UUID              `json:"project_id" example:"11111111-1111-1111-1111-111111111111"`
	TenantID             string                 `json:"tenant_id" example:"default"`
	Name                 string                 `json:"name" example:"João Silva"`
	Email                string                 `json:"email,omitempty" example:"joao@example.com"`
	Phone                string                 `json:"phone,omitempty" example:"5511999999999"`
	ExternalID           string                 `json:"external_id,omitempty" example:"ext-123"`
	SourceChannel        string                 `json:"source_channel,omitempty" example:"whatsapp"`
	Language             string                 `json:"language" example:"pt-BR"`
	Timezone             string                 `json:"timezone" example:"America/Sao_Paulo"`
	Tags                 []string               `json:"tags,omitempty" example:"vip,lead"`
	FirstInteractionAt   *time.Time             `json:"first_interaction_at,omitempty"`
	LastInteractionAt    *time.Time             `json:"last_interaction_at,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
} // @name Contact

// SessionResponse representa uma sessão na API
type SessionResponse struct {
	ID                   uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ContactID            uuid.UUID              `json:"contact_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID             string                 `json:"tenant_id" example:"default"`
	ChannelTypeID        *int                   `json:"channel_type_id,omitempty" example:"1"`
	StartedAt            time.Time              `json:"started_at"`
	EndedAt              *time.Time             `json:"ended_at,omitempty"`
	Status               string                 `json:"status" example:"active"`
	EndReason            *string                `json:"end_reason,omitempty" example:"timeout"`
	TimeoutDuration      int64                  `json:"timeout_duration" example:"1800000000000"`
	LastActivityAt       time.Time              `json:"last_activity_at"`
	MessageCount         int                    `json:"message_count" example:"5"`
	MessagesFromContact  int                    `json:"messages_from_contact" example:"3"`
	MessagesFromAgent    int                    `json:"messages_from_agent" example:"2"`
	DurationSeconds      int                    `json:"duration_seconds" example:"120"`
	Summary              *string                `json:"summary,omitempty"`
	Sentiment            *string                `json:"sentiment,omitempty" example:"positive"`
	SentimentScore       *float64               `json:"sentiment_score,omitempty" example:"0.85"`
	Topics               []string               `json:"topics,omitempty"`
	Resolved             bool                   `json:"resolved" example:"false"`
	Escalated            bool                   `json:"escalated" example:"false"`
	Converted            bool                   `json:"converted" example:"false"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
} // @name Session

// MessageResponse representa uma mensagem na API
type MessageResponse struct {
	ID               uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ContactID        uuid.UUID              `json:"contact_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SessionID        *uuid.UUID             `json:"session_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProjectID        uuid.UUID              `json:"project_id" example:"11111111-1111-1111-1111-111111111111"`
	ExternalID       string                 `json:"external_id,omitempty" example:"msg-ext-123"`
	ContentType      string                 `json:"content_type" example:"text"`
	Direction        string                 `json:"direction" example:"inbound"`
	FromMe           bool                   `json:"from_me" example:"false"`
	Text             *string                `json:"text,omitempty" example:"Olá! Preciso de ajuda."`
	MediaURL         *string                `json:"media_url,omitempty"`
	MediaMimetype    *string                `json:"media_mimetype,omitempty"`
	Status           string                 `json:"status" example:"delivered"`
	SentAt           *time.Time             `json:"sent_at,omitempty"`
	DeliveredAt      *time.Time             `json:"delivered_at,omitempty"`
	ReadAt           *time.Time             `json:"read_at,omitempty"`
	FailedAt         *time.Time             `json:"failed_at,omitempty"`
	ErrorMessage     *string                `json:"error_message,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
} // @name Message

// PipelineResponse representa um pipeline na API
type PipelineResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProjectID   uuid.UUID `json:"project_id" example:"11111111-1111-1111-1111-111111111111"`
	TenantID    string    `json:"tenant_id" example:"default"`
	Name        string    `json:"name" example:"Pipeline de Vendas"`
	Description string    `json:"description,omitempty" example:"Pipeline principal de vendas"`
	Color       string    `json:"color" example:"#3B82F6"`
	Position    int       `json:"position" example:"1"`
	Active      bool      `json:"active" example:"true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} // @name Pipeline

// PipelineStatusResponse representa um status de pipeline na API
type PipelineStatusResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PipelineID  uuid.UUID `json:"pipeline_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"Novo Lead"`
	Description string    `json:"description,omitempty" example:"Contato recém chegado"`
	Color       string    `json:"color" example:"#10B981"`
	StatusType  string    `json:"status_type" example:"open"`
	Position    int       `json:"position" example:"1"`
	Active      bool      `json:"active" example:"true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
} // @name PipelineStatus

// ProjectResponse representa um projeto na API
type ProjectResponse struct {
	ID          uuid.UUID              `json:"id" example:"11111111-1111-1111-1111-111111111111"`
	UserID      uuid.UUID              `json:"user_id" example:"22222222-2222-2222-2222-222222222222"`
	Name        string                 `json:"name" example:"Meu Projeto"`
	Description string                 `json:"description,omitempty" example:"Projeto de vendas"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	Active      bool                   `json:"active" example:"true"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
} // @name Project

// WebhookSubscriptionResponse representa uma inscrição de webhook na API
type WebhookSubscriptionResponse struct {
	ID               uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name             string     `json:"name" example:"Webhook N8N"`
	URL              string     `json:"url" example:"https://webhook.site/unique-url"`
	Events           []string   `json:"events" example:"contact.created,session.started"`
	Active           bool       `json:"active" example:"true"`
	RetryCount       int        `json:"retry_count" example:"3"`
	TimeoutSeconds   int        `json:"timeout_seconds" example:"30"`
	LastTriggeredAt  *time.Time `json:"last_triggered_at,omitempty"`
	LastSuccessAt    *time.Time `json:"last_success_at,omitempty"`
	LastFailureAt    *time.Time `json:"last_failure_at,omitempty"`
	SuccessCount     int        `json:"success_count" example:"100"`
	FailureCount     int        `json:"failure_count" example:"2"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
} // @name WebhookSubscription

// AgentResponse representa um agente na API
type AgentResponse struct {
	ID          uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID    string                 `json:"tenant_id" example:"default"`
	Name        string                 `json:"name" example:"Maria Atendente"`
	Email       string                 `json:"email" example:"maria@ventros.com"`
	Role        string                 `json:"role" example:"agent"`
	Status      string                 `json:"status" example:"active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	LastLoginAt *time.Time             `json:"last_login_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
} // @name Agent

// ErrorResponse representa uma resposta de erro da API
type ErrorResponse struct {
	Error   string                 `json:"error" example:"Validation failed"`
	Message string                 `json:"message,omitempty" example:"Invalid email format"`
	Details map[string]interface{} `json:"details,omitempty"`
} // @name Error

// SuccessResponse representa uma resposta de sucesso genérica
type SuccessResponse struct {
	Success bool                   `json:"success" example:"true"`
	Message string                 `json:"message,omitempty" example:"Operation completed successfully"`
	Data    map[string]interface{} `json:"data,omitempty"`
} // @name Success

// ListResponse representa uma resposta de lista paginada
type ListResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total" example:"100"`
	Page       int         `json:"page" example:"1"`
	PageSize   int         `json:"page_size" example:"20"`
	TotalPages int         `json:"total_pages" example:"5"`
} // @name List
