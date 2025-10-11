package entities

import (
	"time"

	"github.com/google/uuid"
)

// BroadcastEntity represents a broadcast in the database
type BroadcastEntity struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
	TenantID        string     `gorm:"type:varchar(255);not null;index:idx_broadcasts_tenant"`
	Name            string     `gorm:"type:varchar(255);not null"`
	ListID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_broadcasts_list"`
	MessageTemplate []byte     `gorm:"type:jsonb;not null"`
	Status          string     `gorm:"type:varchar(50);not null;index:idx_broadcasts_status"`
	ScheduledFor    *time.Time `gorm:"type:timestamp"`
	StartedAt       *time.Time `gorm:"type:timestamp"`
	CompletedAt     *time.Time `gorm:"type:timestamp"`
	TotalContacts   int        `gorm:"type:int;default:0"`
	SentCount       int        `gorm:"type:int;default:0"`
	FailedCount     int        `gorm:"type:int;default:0"`
	PendingCount    int        `gorm:"type:int;default:0"`
	RateLimit       int        `gorm:"type:int;default:0"`
	CreatedAt       time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (BroadcastEntity) TableName() string {
	return "broadcasts"
}

// BroadcastExecutionEntity represents a broadcast execution in the database
type BroadcastExecutionEntity struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"`
	BroadcastID uuid.UUID  `gorm:"type:uuid;not null;index:idx_executions_broadcast"`
	ContactID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_executions_contact"`
	Status      string     `gorm:"type:varchar(50);not null;index:idx_executions_status"`
	MessageID   *uuid.UUID `gorm:"type:uuid"`
	Error       *string    `gorm:"type:text"`
	SentAt      *time.Time `gorm:"type:timestamp"`
	CreatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (BroadcastExecutionEntity) TableName() string {
	return "broadcast_executions"
}
