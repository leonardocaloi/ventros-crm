package entities

import (
	"time"

	"github.com/google/uuid"
)

// AIAgentHistoryEntity representa o histórico de envio concatenado para AI Agent
type AIAgentHistoryEntity struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey"`
	GroupID             uuid.UUID `gorm:"type:uuid;not null;index:idx_ai_agent_history_group_id"`
	SessionID           uuid.UUID `gorm:"type:uuid;not null;index:idx_ai_agent_history_session_id"`
	ContactID           uuid.UUID `gorm:"type:uuid;not null;index:idx_ai_agent_history_contact"`
	ChannelID           uuid.UUID `gorm:"type:uuid;not null"`
	TenantID            string    `gorm:"type:varchar(255);not null;index:idx_ai_agent_history_tenant"`
	ConcatenatedContent string    `gorm:"type:text;not null"`
	MessageCount        int       `gorm:"not null"`
	EnrichmentCount     int       `gorm:"not null"`
	SentToAI            bool      `gorm:"not null;default:false;index:idx_ai_agent_history_sent_to_ai"`
	AIResponse          *string   `gorm:"type:text"`
	AIProvider          *string   `gorm:"type:varchar(50)"`
	AIModel             *string   `gorm:"type:varchar(100)"`
	ProcessingTimeMs    *int      `gorm:"type:int"`
	CreatedAt           time.Time `gorm:"not null;default:now()"`
	SentAt              *time.Time
	ResponseReceivedAt  *time.Time

	// Relações
	MessageGroup *MessageGroupEntity `gorm:"foreignKey:GroupID;references:ID"`
	Session      *SessionEntity      `gorm:"foreignKey:SessionID;references:ID"`
	Contact      *ContactEntity      `gorm:"foreignKey:ContactID;references:ID"`
	Channel      *ChannelEntity      `gorm:"foreignKey:ChannelID;references:ID"`
}

func (AIAgentHistoryEntity) TableName() string {
	return "agent_ai_interactions"
}
