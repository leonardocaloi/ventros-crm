package entities

import (
	"time"

	"github.com/google/uuid"
)

// CampaignEntity represents a campaign in the database
type CampaignEntity struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key"`
	TenantID         string     `gorm:"type:varchar(255);not null;index:idx_campaigns_tenant"`
	Name             string     `gorm:"type:varchar(255);not null"`
	Description      string     `gorm:"type:text"`
	Status           string     `gorm:"type:varchar(50);not null;index:idx_campaigns_status"`
	GoalType         string     `gorm:"type:varchar(50);not null"`
	GoalValue        int        `gorm:"type:int;not null;default:0"`
	ContactsReached  int        `gorm:"type:int;default:0"`
	ConversionsCount int        `gorm:"type:int;default:0"`
	StartDate        *time.Time `gorm:"type:timestamp"`
	EndDate          *time.Time `gorm:"type:timestamp"`
	CreatedAt        time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (CampaignEntity) TableName() string {
	return "campaigns"
}

// CampaignStepEntity represents a campaign step in the database
type CampaignStepEntity struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	CampaignID uuid.UUID `gorm:"type:uuid;not null;index:idx_campaign_steps_campaign_id"`
	Order      int       `gorm:"type:int;not null"`
	Name       string    `gorm:"type:varchar(255);not null"`
	Type       string    `gorm:"type:varchar(50);not null"`  // broadcast, sequence, delay, condition, wait
	Config     []byte    `gorm:"type:jsonb;not null"`        // step configuration (broadcast_id, sequence_id, etc.)
	Conditions []byte    `gorm:"type:jsonb"`                 // execution conditions
	CreatedAt  time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (CampaignStepEntity) TableName() string {
	return "campaign_steps"
}

// CampaignEnrollmentEntity represents a contact's enrollment in a campaign
type CampaignEnrollmentEntity struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key"`
	CampaignID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_campaign_enrollments_campaign_id"`
	ContactID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_campaign_enrollments_contact_id"`
	Status           string     `gorm:"type:varchar(50);not null;index:idx_campaign_enrollments_status"`
	CurrentStepOrder int        `gorm:"type:int;not null;default:0"`
	NextScheduledAt  *time.Time `gorm:"type:timestamp;index:idx_campaign_enrollments_next_scheduled"`
	ExitedAt         *time.Time `gorm:"type:timestamp"`
	ExitReason       *string    `gorm:"type:text"`
	CompletedAt      *time.Time `gorm:"type:timestamp"`
	EnrolledAt       time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (CampaignEnrollmentEntity) TableName() string {
	return "campaign_enrollments"
}
