package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// MessageEntity representa a entidade Message no banco de dados
type MessageEntity struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID      string     `gorm:"not null;index:idx_messages_tenant;index:idx_messages_tenant_session,priority:1;index:idx_messages_tenant_contact,priority:1;index:idx_messages_tenant_timestamp,priority:1"`
	Timestamp     time.Time  `gorm:"not null;index:idx_messages_timestamp;index:idx_messages_tenant_timestamp,priority:2"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_messages_user;column:user_id"` // User ID (dono do workspace)
	ProjectID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_messages_project"`
	ChannelTypeID *int       `gorm:"index:idx_messages_channel_type"`
	FromMe        bool       `gorm:"default:false;index:idx_messages_from_me"`
	ChannelID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_messages_channel"` // OBRIGATÓRIO
	ChatID        *uuid.UUID `gorm:"type:uuid;index:idx_messages_chat_id"` // Chat this message belongs to
	ContactID     uuid.UUID  `gorm:"type:uuid;not null;index:idx_messages_contact;index:idx_messages_tenant_contact,priority:2"`
	SessionID     *uuid.UUID `gorm:"type:uuid;index:idx_messages_session;index:idx_messages_tenant_session,priority:2"`
	ContentType   string     `gorm:"default:'text';not null;index:idx_messages_content_type"`
	Text          *string    `gorm:"type:text"`

	// Media fields
	MediaURL         *string    `gorm:""`
	MediaMimetype    *string    `gorm:""`
	ChannelMessageID *string    `gorm:"index"`
	ReplyToID        *uuid.UUID `gorm:"type:uuid"`

	Status      string                 `gorm:"default:'sent';index:idx_messages_status"`
	Language    *string                `gorm:""`
	AgentID     *uuid.UUID             `gorm:"type:uuid;index:idx_messages_agent"`
	Metadata    map[string]interface{} `gorm:"type:jsonb;index:idx_messages_metadata,type:gin"`
	Mentions    pq.StringArray         `gorm:"type:text[]"` // IDs externos mencionados (formato WAHA: "phone@c.us")
	DeliveredAt *time.Time             `gorm:"index:idx_messages_delivered_at"`
	ReadAt      *time.Time             `gorm:"index:idx_messages_read_at"`

	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_messages_created"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_messages_updated"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_messages_deleted"`

	// Relacionamentos
	Contact ContactEntity  `gorm:"foreignKey:ContactID"`
	Chat    *ChatEntity    `gorm:"foreignKey:ChatID"`
	Session *SessionEntity `gorm:"foreignKey:SessionID"`
	Project ProjectEntity  `gorm:"foreignKey:ProjectID"`
	Channel ChannelEntity  `gorm:"foreignKey:ChannelID;constraint:OnDelete:RESTRICT"`
}

func (MessageEntity) TableName() string {
	return "messages"
}
