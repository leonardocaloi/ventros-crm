package entities

import (
	"time"

	"github.com/google/uuid"
)

// SequenceEntity represents a sequence in the database
type SequenceEntity struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key"`
	TenantID       string     `gorm:"type:varchar(255);not null;index:idx_sequences_tenant"`
	Name           string     `gorm:"type:varchar(255);not null"`
	Description    string     `gorm:"type:text"`
	Status         string     `gorm:"type:varchar(50);not null;index:idx_sequences_status"`
	TriggerType    string     `gorm:"type:varchar(50);not null;index:idx_sequences_trigger_type"`
	TriggerData    []byte     `gorm:"type:jsonb"` // JSON data for trigger configuration
	ExitOnReply    bool       `gorm:"type:boolean;default:true"`
	TotalEnrolled  int        `gorm:"type:int;default:0"`
	ActiveCount    int        `gorm:"type:int;default:0"`
	CompletedCount int        `gorm:"type:int;default:0"`
	ExitedCount    int        `gorm:"type:int;default:0"`
	CreatedAt      time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (SequenceEntity) TableName() string {
	return "sequences"
}

// SequenceStepEntity represents a sequence step in the database
type SequenceStepEntity struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key"`
	SequenceID      uuid.UUID `gorm:"type:uuid;not null;index:idx_sequence_steps_sequence_id"`
	Order           int       `gorm:"type:int;not null"`
	Name            string    `gorm:"type:varchar(255);not null"`
	DelayAmount     int       `gorm:"type:int;not null"`
	DelayUnit       string    `gorm:"type:varchar(20);not null"` // minutes, hours, days
	MessageTemplate []byte    `gorm:"type:jsonb;not null"`
	Conditions      []byte    `gorm:"type:jsonb"` // JSON array of conditions
	CreatedAt       time.Time `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (SequenceStepEntity) TableName() string {
	return "sequence_steps"
}

// SequenceEnrollmentEntity represents a contact's enrollment in a sequence
type SequenceEnrollmentEntity struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key"`
	SequenceID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_enrollments_sequence_id"`
	ContactID        uuid.UUID  `gorm:"type:uuid;not null;index:idx_enrollments_contact_id"`
	Status           string     `gorm:"type:varchar(50);not null;index:idx_enrollments_status"`
	CurrentStepOrder int        `gorm:"type:int;not null;default:0"`
	NextScheduledAt  *time.Time `gorm:"type:timestamp;index:idx_enrollments_next_scheduled"`
	ExitedAt         *time.Time `gorm:"type:timestamp"`
	ExitReason       *string    `gorm:"type:text"`
	CompletedAt      *time.Time `gorm:"type:timestamp"`
	EnrolledAt       time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time  `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
}

// TableName specifies the table name for GORM
func (SequenceEnrollmentEntity) TableName() string {
	return "sequence_enrollments"
}
