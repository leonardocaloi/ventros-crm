package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SessionEntity representa a entidade Session no banco de dados
type SessionEntity struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ContactID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_sessions_contact;index:idx_sessions_tenant_contact,priority:2"`
	TenantID        string     `gorm:"not null;index:idx_sessions_tenant;index:idx_sessions_tenant_status,priority:1;index:idx_sessions_tenant_contact,priority:1;index:idx_sessions_tenant_started,priority:1"`
	ChannelTypeID   *int       `gorm:"index:idx_sessions_channel_type"`
	PipelineID      *uuid.UUID `gorm:"type:uuid;index:idx_sessions_pipeline"` // Pipeline que define timeout e fluxo
	StartedAt       time.Time  `gorm:"not null;index:idx_sessions_started;index:idx_sessions_tenant_started,priority:2"`
	EndedAt         *time.Time `gorm:"index:idx_sessions_ended"`
	Status          string     `gorm:"default:'active';index:idx_sessions_status;index:idx_sessions_tenant_status,priority:2"`
	EndReason       *string    `gorm:""`
	TimeoutDuration int64      `gorm:"default:1800000000000"` // nanoseconds
	LastActivityAt  time.Time  `gorm:"not null;index:idx_sessions_last_activity"`

	// Metrics
	MessageCount        int `gorm:"default:0"`
	MessagesFromContact int `gorm:"default:0"`
	MessagesFromAgent   int `gorm:"default:0"`
	DurationSeconds     int `gorm:"default:0"`

	// Response Time Metrics (para lead score e feedback comercial)
	FirstContactMessageAt    *time.Time `gorm:""` // Primeira mensagem do contato
	FirstAgentResponseAt     *time.Time `gorm:""` // Primeira resposta do agente
	AgentResponseTimeSeconds *int       `gorm:""` // Tempo de espera até primeira resposta do agente
	ContactWaitTimeSeconds   *int       `gorm:""` // Tempo de espera do contato (se agente iniciou)

	// Agents
	AgentIDs       []uuid.UUID `gorm:"type:jsonb;index:idx_sessions_agent_ids,type:gin"`
	AgentTransfers int         `gorm:"default:0"`

	// AI/Analytics
	Summary        *string        `gorm:"type:text"`
	Sentiment      *string        `gorm:"index:idx_sessions_sentiment"`
	SentimentScore *float64       `gorm:""`
	Topics         []string       `gorm:"type:jsonb;index:idx_sessions_topics,type:gin"`
	NextSteps      []string       `gorm:"type:jsonb"`
	KeyEntities    datatypes.JSON `gorm:"type:jsonb;index:idx_sessions_key_entities,type:gin"`

	// Business flags
	Resolved    bool     `gorm:"default:false;index:idx_sessions_resolved"`
	Escalated   bool     `gorm:"default:false;index:idx_sessions_escalated"`
	Converted   bool     `gorm:"default:false;index:idx_sessions_converted"`
	OutcomeTags []string `gorm:"type:jsonb;index:idx_sessions_outcome_tags,type:gin"`

	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_sessions_created"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_sessions_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_sessions_deleted"`

	// Relacionamentos
	Contact      ContactEntity              `gorm:"foreignKey:ContactID"`
	Messages     []MessageEntity            `gorm:"foreignKey:SessionID"`
	CustomFields []SessionCustomFieldEntity `gorm:"foreignKey:SessionID"`
}

func (SessionEntity) TableName() string {
	return "sessions"
}
