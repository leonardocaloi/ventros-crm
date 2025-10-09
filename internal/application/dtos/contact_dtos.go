package dtos

import (
	"time"

	"github.com/google/uuid"
)

// ContactListDTO - DTO otimizado para listagem de contatos
type ContactListDTO struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Email             *string    `json:"email,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	ExternalID        *string    `json:"external_id,omitempty"`
	Tags              []string   `json:"tags,omitempty"`
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"`
	SessionCount      int        `json:"session_count"`
	MessageCount      int        `json:"message_count"`
	PipelineStatus    *string    `json:"pipeline_status,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ContactDetailDTO - DTO completo para detalhes do contato
type ContactDetailDTO struct {
	ID                 uuid.UUID              `json:"id"`
	ProjectID          uuid.UUID              `json:"project_id"`
	TenantID           string                 `json:"tenant_id"`
	Name               string                 `json:"name"`
	Email              *string                `json:"email,omitempty"`
	Phone              *string                `json:"phone,omitempty"`
	ExternalID         *string                `json:"external_id,omitempty"`
	SourceChannel      *string                `json:"source_channel,omitempty"`
	Language           *string                `json:"language,omitempty"`
	Timezone           *string                `json:"timezone,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	CustomFields       map[string]interface{} `json:"custom_fields,omitempty"`
	FirstInteractionAt *time.Time             `json:"first_interaction_at,omitempty"`
	LastInteractionAt  *time.Time             `json:"last_interaction_at,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`

	// Agregados relacionados
	ActiveSessions []SessionSummaryDTO   `json:"active_sessions,omitempty"`
	RecentMessages []MessageSummaryDTO   `json:"recent_messages,omitempty"`
	PipelineStatus *PipelineStatusDTO    `json:"pipeline_status,omitempty"`
	Statistics     *ContactStatisticsDTO `json:"statistics,omitempty"`
}

// ContactStatisticsDTO - Estatísticas agregadas do contato
type ContactStatisticsDTO struct {
	TotalSessions        int        `json:"total_sessions"`
	TotalMessages        int        `json:"total_messages"`
	MessagesFromContact  int        `json:"messages_from_contact"`
	MessagesToContact    int        `json:"messages_to_contact"`
	AverageResponseTime  *float64   `json:"average_response_time_seconds,omitempty"`
	LastSessionStartedAt *time.Time `json:"last_session_started_at,omitempty"`
	LastSessionEndedAt   *time.Time `json:"last_session_ended_at,omitempty"`
}

// CreateContactDTO - DTO para criação de contato
type CreateContactDTO struct {
	ProjectID     uuid.UUID              `json:"project_id" binding:"required"`
	Name          string                 `json:"name" binding:"required"`
	Email         *string                `json:"email,omitempty"`
	Phone         *string                `json:"phone,omitempty"`
	ExternalID    *string                `json:"external_id,omitempty"`
	SourceChannel *string                `json:"source_channel,omitempty"`
	Language      *string                `json:"language,omitempty"`
	Timezone      *string                `json:"timezone,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	CustomFields  map[string]interface{} `json:"custom_fields,omitempty"`
}

// UpdateContactDTO - DTO para atualização de contato
type UpdateContactDTO struct {
	Name         *string                `json:"name,omitempty"`
	Email        *string                `json:"email,omitempty"`
	Phone        *string                `json:"phone,omitempty"`
	Language     *string                `json:"language,omitempty"`
	Timezone     *string                `json:"timezone,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// ContactFilters - Filtros para busca de contatos
type ContactFilters struct {
	ProjectID         *uuid.UUID `form:"project_id"`
	Search            *string    `form:"search"` // Busca por nome, email, phone
	Tags              []string   `form:"tags"`
	SourceChannel     *string    `form:"source_channel"`
	HasActiveSessions *bool      `form:"has_active_sessions"`
	CreatedAfter      *time.Time `form:"created_after"`
	CreatedBefore     *time.Time `form:"created_before"`
	Limit             int        `form:"limit" binding:"min=1,max=100"`
	Offset            int        `form:"offset" binding:"min=0"`
	SortBy            string     `form:"sort_by"`    // name, created_at, last_interaction_at
	SortOrder         string     `form:"sort_order"` // asc, desc
}
