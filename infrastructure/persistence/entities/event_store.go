package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContactEventStoreEntity represents an event in the event store (Event Sourcing pattern).
// This is an append-only immutable log of all changes to Contact aggregates.
type ContactEventStoreEntity struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Aggregate identification
	AggregateID   uuid.UUID `gorm:"type:uuid;not null;index:idx_contact_events_aggregate,priority:1"`
	AggregateType string    `gorm:"type:varchar(50);not null;default:'contact'"`

	// Event metadata
	EventType      string `gorm:"type:varchar(100);not null;index:idx_contact_events_type,priority:1"`
	EventVersion   string `gorm:"type:varchar(10);not null;default:'v1'"`
	SequenceNumber int64  `gorm:"not null;index:idx_contact_events_aggregate,priority:2;uniqueIndex:unique_aggregate_sequence,priority:2"`

	// Event data
	EventData map[string]interface{} `gorm:"type:jsonb;not null"`
	Metadata  map[string]interface{} `gorm:"type:jsonb"`

	// Temporal
	OccurredAt time.Time `gorm:"not null;index:idx_contact_events_type,priority:2,sort:desc;index:idx_contact_events_occurred,sort:desc"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`

	// Multi-tenancy
	TenantID  string     `gorm:"type:varchar(255);not null;index:idx_contact_events_tenant,priority:1"`
	ProjectID *uuid.UUID `gorm:"type:uuid"`

	// Causation and correlation
	CausationID   *uuid.UUID `gorm:"type:uuid"`
	CorrelationID *uuid.UUID `gorm:"type:uuid;index:idx_contact_events_correlation"`
}

// TableName specifies the table name for GORM
func (ContactEventStoreEntity) TableName() string {
	return "contact_event_store"
}

// BeforeCreate sets UUID if not already set
func (e *ContactEventStoreEntity) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	return nil
}

// ContactSnapshotEntity represents a snapshot of a Contact aggregate state.
// Snapshots are used to optimize performance by avoiding replay of all events.
type ContactSnapshotEntity struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`

	// Aggregate identification
	AggregateID uuid.UUID `gorm:"type:uuid;not null;index:idx_contact_snapshots_aggregate,priority:1"`

	// Snapshot data
	SnapshotData       map[string]interface{} `gorm:"type:jsonb;not null"`
	LastSequenceNumber int64                  `gorm:"not null;index:idx_contact_snapshots_aggregate,priority:2,sort:desc;uniqueIndex:unique_aggregate_snapshot,priority:2"`

	// Temporal
	CreatedAt time.Time `gorm:"not null;default:now();index:idx_contact_snapshots_tenant,priority:2,sort:desc"`

	// Multi-tenancy
	TenantID string `gorm:"type:varchar(255);not null;index:idx_contact_snapshots_tenant,priority:1"`
}

// TableName specifies the table name for GORM
func (ContactSnapshotEntity) TableName() string {
	return "contact_snapshots"
}

// BeforeCreate sets UUID if not already set
func (s *ContactSnapshotEntity) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	return nil
}
