package contact_event

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ContactEvent struct {
	id        uuid.UUID
	contactID uuid.UUID
	sessionID *uuid.UUID
	tenantID  string

	eventType string
	category  Category
	priority  Priority

	title       *string
	description *string
	payload     map[string]interface{}
	metadata    map[string]interface{}

	source            Source
	triggeredBy       *uuid.UUID
	integrationSource *string

	isRealtime  bool
	delivered   bool
	deliveredAt *time.Time
	read        bool
	readAt      *time.Time

	visibleToClient bool
	visibleToAgent  bool
	expiresAt       *time.Time

	occurredAt time.Time
	createdAt  time.Time
}

func NewContactEvent(
	contactID uuid.UUID,
	tenantID string,
	eventType string,
	category Category,
	priority Priority,
	source Source,
) (*ContactEvent, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if eventType == "" {
		return nil, errors.New("eventType cannot be empty")
	}
	if !category.IsValid() {
		return nil, errors.New("invalid category")
	}
	if !priority.IsValid() {
		return nil, errors.New("invalid priority")
	}
	if !source.IsValid() {
		return nil, errors.New("invalid source")
	}

	now := time.Now()
	return &ContactEvent{
		id:              uuid.New(),
		contactID:       contactID,
		tenantID:        tenantID,
		eventType:       eventType,
		category:        category,
		priority:        priority,
		source:          source,
		payload:         make(map[string]interface{}),
		metadata:        make(map[string]interface{}),
		isRealtime:      true,
		delivered:       false,
		read:            false,
		visibleToClient: true,
		visibleToAgent:  true,
		occurredAt:      now,
		createdAt:       now,
	}, nil
}

func ReconstructContactEvent(
	id uuid.UUID,
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	tenantID string,
	eventType string,
	category Category,
	priority Priority,
	title *string,
	description *string,
	payload map[string]interface{},
	metadata map[string]interface{},
	source Source,
	triggeredBy *uuid.UUID,
	integrationSource *string,
	isRealtime bool,
	delivered bool,
	deliveredAt *time.Time,
	read bool,
	readAt *time.Time,
	visibleToClient bool,
	visibleToAgent bool,
	expiresAt *time.Time,
	occurredAt time.Time,
	createdAt time.Time,
) *ContactEvent {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &ContactEvent{
		id:                id,
		contactID:         contactID,
		sessionID:         sessionID,
		tenantID:          tenantID,
		eventType:         eventType,
		category:          category,
		priority:          priority,
		title:             title,
		description:       description,
		payload:           payload,
		metadata:          metadata,
		source:            source,
		triggeredBy:       triggeredBy,
		integrationSource: integrationSource,
		isRealtime:        isRealtime,
		delivered:         delivered,
		deliveredAt:       deliveredAt,
		read:              read,
		readAt:            readAt,
		visibleToClient:   visibleToClient,
		visibleToAgent:    visibleToAgent,
		expiresAt:         expiresAt,
		occurredAt:        occurredAt,
		createdAt:         createdAt,
	}
}

func (e *ContactEvent) AttachToSession(sessionID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return errors.New("sessionID cannot be nil")
	}
	e.sessionID = &sessionID
	return nil
}

func (e *ContactEvent) SetTitle(title string) {
	e.title = &title
}

func (e *ContactEvent) SetDescription(description string) {
	e.description = &description
}

func (e *ContactEvent) AddPayloadField(key string, value interface{}) {
	e.payload[key] = value
}

func (e *ContactEvent) AddMetadataField(key string, value interface{}) {
	e.metadata[key] = value
}

func (e *ContactEvent) SetTriggeredBy(agentID uuid.UUID) error {
	if agentID == uuid.Nil {
		return errors.New("agentID cannot be nil")
	}
	e.triggeredBy = &agentID
	return nil
}

func (e *ContactEvent) SetIntegrationSource(source string) {
	e.integrationSource = &source
}

func (e *ContactEvent) MarkAsDelivered() {
	if !e.delivered {
		now := time.Now()
		e.delivered = true
		e.deliveredAt = &now
	}
}

func (e *ContactEvent) MarkAsRead() {
	if !e.read {
		now := time.Now()
		e.read = true
		e.readAt = &now
	}
}

func (e *ContactEvent) SetRealtimeDelivery(enabled bool) {
	e.isRealtime = enabled
}

func (e *ContactEvent) SetVisibility(visibleToClient, visibleToAgent bool) {
	e.visibleToClient = visibleToClient
	e.visibleToAgent = visibleToAgent
}

func (e *ContactEvent) SetExpiresAt(expiresAt time.Time) error {
	if expiresAt.Before(time.Now()) {
		return errors.New("expiresAt cannot be in the past")
	}
	e.expiresAt = &expiresAt
	return nil
}

func (e *ContactEvent) IsExpired() bool {
	if e.expiresAt == nil {
		return false
	}
	return time.Now().After(*e.expiresAt)
}

func (e *ContactEvent) IsDelivered() bool {
	return e.delivered
}

func (e *ContactEvent) IsRead() bool {
	return e.read
}

func (e *ContactEvent) ShouldBeDeliveredInRealtime() bool {
	return e.isRealtime && !e.delivered && !e.IsExpired()
}

func (e *ContactEvent) IsVisibleToClient() bool {
	return e.visibleToClient && !e.IsExpired()
}

func (e *ContactEvent) IsVisibleToAgent() bool {
	return e.visibleToAgent && !e.IsExpired()
}

func (e *ContactEvent) HasSession() bool {
	return e.sessionID != nil
}

func (e *ContactEvent) IsHighPriority() bool {
	return e.priority == PriorityHigh || e.priority == PriorityUrgent
}

func (e *ContactEvent) IsSystemGenerated() bool {
	return e.source == SourceSystem
}

func (e *ContactEvent) ID() uuid.UUID              { return e.id }
func (e *ContactEvent) ContactID() uuid.UUID       { return e.contactID }
func (e *ContactEvent) SessionID() *uuid.UUID      { return e.sessionID }
func (e *ContactEvent) TenantID() string           { return e.tenantID }
func (e *ContactEvent) EventType() string          { return e.eventType }
func (e *ContactEvent) Category() Category         { return e.category }
func (e *ContactEvent) Priority() Priority         { return e.priority }
func (e *ContactEvent) Title() *string             { return e.title }
func (e *ContactEvent) Description() *string       { return e.description }
func (e *ContactEvent) Source() Source             { return e.source }
func (e *ContactEvent) TriggeredBy() *uuid.UUID    { return e.triggeredBy }
func (e *ContactEvent) IntegrationSource() *string { return e.integrationSource }
func (e *ContactEvent) DeliveredAt() *time.Time    { return e.deliveredAt }
func (e *ContactEvent) ReadAt() *time.Time         { return e.readAt }
func (e *ContactEvent) ExpiresAt() *time.Time      { return e.expiresAt }
func (e *ContactEvent) OccurredAt() time.Time      { return e.occurredAt }
func (e *ContactEvent) CreatedAt() time.Time       { return e.createdAt }

func (e *ContactEvent) Payload() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range e.payload {
		copy[k] = v
	}
	return copy
}

func (e *ContactEvent) Metadata() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range e.metadata {
		copy[k] = v
	}
	return copy
}
