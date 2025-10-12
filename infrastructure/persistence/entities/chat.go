package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChatEntity represents the Chat aggregate in the database
type ChatEntity struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Version       int            `gorm:"default:1;not null"` // Optimistic locking
	ProjectID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_chats_project"`
	TenantID      string         `gorm:"not null;index:idx_chats_tenant;index:idx_chats_tenant_status,priority:1;index:idx_chats_tenant_type,priority:1"`
	ChatType      string         `gorm:"not null;index:idx_chats_type;index:idx_chats_tenant_type,priority:2"`   // individual, group, channel
	ExternalID    *string        `gorm:"type:text;index:idx_chats_external_id;uniqueIndex:uq_chats_external_id"` // External ID from channel (WhatsApp @g.us, Telegram group ID)
	Subject       *string        `gorm:"type:varchar(255);index:idx_chats_subject"`                              // Group/channel name
	Description   *string        `gorm:"type:text"`
	Participants  datatypes.JSON `gorm:"type:jsonb;not null;index:idx_chats_participants,type:gin"`                                 // Array of Participant objects
	Status        string         `gorm:"not null;default:'active';index:idx_chats_status;index:idx_chats_tenant_status,priority:2"` // active, archived, closed
	Metadata      datatypes.JSON `gorm:"type:jsonb;default:'{}'"`                                                                   // Flexible metadata storage
	LastMessageAt *time.Time     `gorm:"index:idx_chats_last_message"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;index:idx_chats_created"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime;index:idx_chats_updated"`
	DeletedAt     gorm.DeletedAt `gorm:"index:idx_chats_deleted"`

	// Relationships
	Project  *ProjectEntity  `gorm:"foreignKey:ProjectID"`
	Messages []MessageEntity `gorm:"foreignKey:ChatID"`
}

// ParticipantJSON represents a participant in JSONB format
type ParticipantJSON struct {
	ID       uuid.UUID  `json:"id"`                // Contact ID or Agent ID
	Type     string     `json:"type"`              // "contact" or "agent"
	JoinedAt time.Time  `json:"joined_at"`         // When joined the chat
	LeftAt   *time.Time `json:"left_at,omitempty"` // When left (for groups/channels)
	IsAdmin  bool       `json:"is_admin"`          // Is admin/moderator (for groups)
}

func (ChatEntity) TableName() string {
	return "chats"
}
