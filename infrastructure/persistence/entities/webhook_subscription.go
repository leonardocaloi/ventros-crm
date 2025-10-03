package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// WebhookSubscriptionEntity representa a entidade WebhookSubscription no banco de dados
type WebhookSubscriptionEntity struct {
	ID              uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name            string                 `gorm:"not null"`
	URL             string                 `gorm:"not null"`
	Events          pq.StringArray         `gorm:"type:text[]"`
	Active          bool                   `gorm:"default:true;index"`
	Secret          string                 `gorm:""`
	Headers         []byte                 `gorm:"type:jsonb"`
	RetryCount      int                    `gorm:"default:3"`
	TimeoutSeconds  int                    `gorm:"default:30"`
	LastTriggeredAt *time.Time             `gorm:""`
	LastSuccessAt   *time.Time             `gorm:""`
	LastFailureAt   *time.Time             `gorm:""`
	SuccessCount    int                    `gorm:"default:0"`
	FailureCount    int                    `gorm:"default:0"`
	CreatedAt       time.Time              `gorm:"autoCreateTime"`
	UpdatedAt       time.Time              `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt         `gorm:"index"`
}

func (WebhookSubscriptionEntity) TableName() string {
	return "webhook_subscriptions"
}
