package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// MessageGroupEntity representa um grupo de mensagens agrupadas pelo debouncer
type MessageGroupEntity struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ContactID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_message_groups_contact_channel"`
	ChannelID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_message_groups_contact_channel"`
	SessionID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_message_groups_session"`
	TenantID    string         `gorm:"type:varchar(255);not null;index:idx_message_groups_tenant"`
	MessageIDs  pq.StringArray `gorm:"type:text[];not null"` // Array de UUIDs como strings
	Status      string         `gorm:"type:varchar(50);not null;index:idx_message_groups_status"`
	StartedAt   time.Time      `gorm:"not null"`
	CompletedAt *time.Time
	ExpiresAt   time.Time `gorm:"not null;index:idx_message_groups_expires_at"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (MessageGroupEntity) TableName() string {
	return "message_groups"
}
