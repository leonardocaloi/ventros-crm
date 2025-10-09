package dtos

import (
	"time"

	"github.com/google/uuid"
)

// SessionSummaryDTO - DTO resumido para sessões
type SessionSummaryDTO struct {
	ID             uuid.UUID  `json:"id"`
	ContactID      uuid.UUID  `json:"contact_id"`
	ContactName    string     `json:"contact_name"`
	ChannelTypeID  *int       `json:"channel_type_id,omitempty"`
	Status         string     `json:"status"`
	StartedAt      time.Time  `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at,omitempty"`
	MessageCount   int        `json:"message_count"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

// SessionDetailDTO - DTO completo para detalhes da sessão
type SessionDetailDTO struct {
	ID                  uuid.UUID              `json:"id"`
	ContactID           uuid.UUID              `json:"contact_id"`
	TenantID            string                 `json:"tenant_id"`
	ChannelTypeID       *int                   `json:"channel_type_id,omitempty"`
	Status              string                 `json:"status"`
	StartedAt           time.Time              `json:"started_at"`
	EndedAt             *time.Time             `json:"ended_at,omitempty"`
	EndReason           *string                `json:"end_reason,omitempty"`
	TimeoutDuration     int64                  `json:"timeout_duration_ns"`
	LastActivityAt      *time.Time             `json:"last_activity_at,omitempty"`
	MessageCount        int                    `json:"message_count"`
	MessagesFromContact int                    `json:"messages_from_contact"`
	MessagesFromAgent   int                    `json:"messages_from_agent"`
	DurationSeconds     *int                   `json:"duration_seconds,omitempty"`
	AgentIDs            []uuid.UUID            `json:"agent_ids,omitempty"`
	AgentTransfers      int                    `json:"agent_transfers"`
	Summary             *string                `json:"summary,omitempty"`
	Sentiment           *string                `json:"sentiment,omitempty"`
	SentimentScore      *float64               `json:"sentiment_score,omitempty"`
	Topics              []string               `json:"topics,omitempty"`
	NextSteps           []string               `json:"next_steps,omitempty"`
	KeyEntities         map[string]interface{} `json:"key_entities,omitempty"`
	Resolved            bool                   `json:"resolved"`
	Escalated           bool                   `json:"escalated"`
	Converted           bool                   `json:"converted"`
	OutcomeTags         []string               `json:"outcome_tags,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`

	// Relacionamentos
	Contact        *ContactSummaryDTO  `json:"contact,omitempty"`
	RecentMessages []MessageSummaryDTO `json:"recent_messages,omitempty"`
	Agents         []AgentSummaryDTO   `json:"agents,omitempty"`
}

// ContactSummaryDTO - DTO resumido para contato (usado em sessões)
type ContactSummaryDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email *string   `json:"email,omitempty"`
	Phone *string   `json:"phone,omitempty"`
}

// SessionStatisticsDTO - Estatísticas de sessões
type SessionStatisticsDTO struct {
	TotalSessions       int     `json:"total_sessions"`
	ActiveSessions      int     `json:"active_sessions"`
	AverageDuration     float64 `json:"average_duration_seconds"`
	AverageMessageCount float64 `json:"average_message_count"`
	ResolutionRate      float64 `json:"resolution_rate"`
	EscalationRate      float64 `json:"escalation_rate"`
	ConversionRate      float64 `json:"conversion_rate"`
}

// SessionFilters - Filtros para busca de sessões
type SessionFilters struct {
	ContactID     *uuid.UUID `form:"contact_id"`
	ChannelTypeID *int       `form:"channel_type_id"`
	Status        *string    `form:"status"`
	StartedAfter  *time.Time `form:"started_after"`
	StartedBefore *time.Time `form:"started_before"`
	Resolved      *bool      `form:"resolved"`
	Escalated     *bool      `form:"escalated"`
	Converted     *bool      `form:"converted"`
	HasAgents     *bool      `form:"has_agents"`
	Limit         int        `form:"limit" binding:"min=1,max=100"`
	Offset        int        `form:"offset" binding:"min=0"`
	SortBy        string     `form:"sort_by"`    // started_at, ended_at, message_count
	SortOrder     string     `form:"sort_order"` // asc, desc
}
