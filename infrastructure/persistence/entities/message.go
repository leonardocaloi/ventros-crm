package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageEntity representa a entidade Message no banco de dados
type MessageEntity struct {
	ID               uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Timestamp        time.Time              `gorm:"not null;index"`
	UserID           uuid.UUID              `gorm:"type:uuid;not null;index;column:user_id"` // User ID (dono do workspace)
	ProjectID        uuid.UUID              `gorm:"type:uuid;not null;index"`
	ChannelTypeID    *int                   `gorm:"index"`
	FromMe           bool                   `gorm:"default:false;index"`
	ChannelID        uuid.UUID              `gorm:"type:uuid;not null;index"` // OBRIGATÓRIO
	ContactID        uuid.UUID              `gorm:"type:uuid;not null;index"`
	SessionID        *uuid.UUID             `gorm:"type:uuid;index"`
	ContentType      string                 `gorm:"default:'text';not null;index"`
	Text             *string                `gorm:"type:text"`
	
	// Media fields
	MediaURL         *string                `gorm:""`
	MediaMimetype    *string                `gorm:""`
	ChannelMessageID *string                `gorm:"index"`
	ReplyToID        *uuid.UUID             `gorm:"type:uuid"`
	
	Status           string                 `gorm:"default:'sent';index"`
	Language         *string                `gorm:""`
	AgentID          *uuid.UUID             `gorm:"type:uuid"`
	Metadata         map[string]interface{} `gorm:"type:jsonb"`
	DeliveredAt      *time.Time             `gorm:""`
	ReadAt           *time.Time             `gorm:""`
	
	CreatedAt        time.Time              `gorm:"autoCreateTime"`
	UpdatedAt        time.Time              `gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt         `gorm:"index"`

	// Relacionamentos
	Contact ContactEntity  `gorm:"foreignKey:ContactID"`
	Session *SessionEntity `gorm:"foreignKey:SessionID"`
	Project ProjectEntity  `gorm:"foreignKey:ProjectID"`
	Channel ChannelEntity `gorm:"foreignKey:ChannelID;constraint:OnDelete:RESTRICT"`
}

func (MessageEntity) TableName() string {
	return "messages"
}
