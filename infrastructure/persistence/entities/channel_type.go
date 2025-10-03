package entities

import (
	"time"

	"gorm.io/gorm"
)

// ChannelTypeEntity representa a entidade ChannelType no banco de dados
type ChannelTypeEntity struct {
	ID            int                    `gorm:"primary_key;autoIncrement:false"`
	Name          string                 `gorm:"not null;uniqueIndex"`
	Description   string                 `gorm:"type:text"`
	Provider      string                 `gorm:"not null"`
	Configuration map[string]interface{} `gorm:"type:jsonb"`
	Active        bool                   `gorm:"default:true;index"`
	CreatedAt     time.Time              `gorm:"autoCreateTime"`
	UpdatedAt     time.Time              `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt         `gorm:"index"`
}

func (ChannelTypeEntity) TableName() string {
	return "channel_types"
}
