package dtos

import (
	"time"

	"github.com/google/uuid"
)

// MessageSummaryDTO - DTO resumido para mensagens
type MessageSummaryDTO struct {
	ID          uuid.UUID  `json:"id"`
	Timestamp   time.Time  `json:"timestamp"`
	FromMe      bool       `json:"from_me"`
	ContentType string     `json:"content_type"`
	Text        *string    `json:"text,omitempty"`
	Status      string     `json:"status"`
	AgentID     *uuid.UUID `json:"agent_id,omitempty"`
}

// MessageDetailDTO - DTO completo para detalhes da mensagem
type MessageDetailDTO struct {
	ID               uuid.UUID              `json:"id"`
	Timestamp        time.Time              `json:"timestamp"`
	UserID           uuid.UUID              `json:"user_id"` // User ID (dono do workspace)
	ProjectID        uuid.UUID              `json:"project_id"`
	ChannelTypeID    *int                   `json:"channel_type_id,omitempty"`
	FromMe           bool                   `json:"from_me"`
	ChannelID        *uuid.UUID             `json:"channel_id,omitempty"`
	ContactID        uuid.UUID              `json:"contact_id"`
	SessionID        *uuid.UUID             `json:"session_id,omitempty"`
	ContentType      string                 `json:"content_type"`
	Text             *string                `json:"text,omitempty"`
	MediaURL         *string                `json:"media_url,omitempty"`
	MediaMimetype    *string                `json:"media_mimetype,omitempty"`
	ChannelMessageID *string                `json:"channel_message_id,omitempty"`
	ReplyToID        *uuid.UUID             `json:"reply_to_id,omitempty"`
	Status           string                 `json:"status"`
	Language         *string                `json:"language,omitempty"`
	AgentID          *uuid.UUID             `json:"agent_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	DeliveredAt      *time.Time             `json:"delivered_at,omitempty"`
	ReadAt           *time.Time             `json:"read_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`

	// Relacionamentos
	Contact *ContactSummaryDTO `json:"contact,omitempty"`
	Agent   *AgentSummaryDTO   `json:"agent,omitempty"`
	ReplyTo *MessageSummaryDTO `json:"reply_to,omitempty"`
}

// SendMessageDTO - DTO para envio de mensagem
type SendMessageDTO struct {
	ContactID   uuid.UUID              `json:"contact_id" binding:"required"`
	ContentType string                 `json:"content_type" binding:"required"`
	Text        *string                `json:"text,omitempty"`
	MediaURL    *string                `json:"media_url,omitempty"`
	ReplyToID   *uuid.UUID             `json:"reply_to_id,omitempty"`
	AgentID     *uuid.UUID             `json:"agent_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MessageFilters - Filtros para busca de mensagens
type MessageFilters struct {
	ContactID     *uuid.UUID `form:"contact_id"`
	SessionID     *uuid.UUID `form:"session_id"`
	ChannelTypeID *int       `form:"channel_type_id"`
	FromMe        *bool      `form:"from_me"`
	ContentType   *string    `form:"content_type"`
	Status        *string    `form:"status"`
	AgentID       *uuid.UUID `form:"agent_id"`
	AfterTime     *time.Time `form:"after_time"`
	BeforeTime    *time.Time `form:"before_time"`
	Search        *string    `form:"search"` // Busca no texto
	Limit         int        `form:"limit" binding:"min=1,max=100"`
	Offset        int        `form:"offset" binding:"min=0"`
	SortOrder     string     `form:"sort_order"` // asc, desc
}
