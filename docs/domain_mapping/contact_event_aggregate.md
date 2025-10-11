# ContactEvent Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~300
**Test Coverage**: Unknown

---

## Overview

- **Purpose**: Activity timeline and audit log for contacts
- **Location**: `internal/domain/contact_event/`
- **Entity**: `infrastructure/persistence/entities/contact_event.go`
- **Repository**: `infrastructure/persistence/gorm_contact_event_repository.go`
- **Aggregate Root**: `ContactEvent`

**Business Problem**:
The ContactEvent aggregate provides a **chronological activity log** for contacts, tracking all important events in the customer journey. It enables agents to see complete contact history, provides audit trails, and supports real-time notifications. Critical for:
- **Activity timeline** - Chronological view of all contact interactions
- **Audit trail** - Track who changed what and when
- **Notifications** - Real-time alerts to agents/contacts
- **Customer journey** - Understand contact lifecycle and touchpoints
- **Analytics** - Analyze patterns and behaviors
- **Compliance** - Maintain records for regulatory requirements

---

## Domain Model

### Aggregate Root: ContactEvent

```go
type ContactEvent struct {
    id        uuid.UUID
    contactID uuid.UUID
    sessionID *uuid.UUID  // Optional - event may not be tied to session
    tenantID  string

    // Event Classification
    eventType string    // "status_changed", "agent_assigned", etc.
    category  Category  // general, status, pipeline, assignment, etc.
    priority  Priority  // low, normal, high, urgent

    // Event Content
    title       *string                 // Human-readable title
    description *string                 // Detailed description
    payload     map[string]interface{}  // Event-specific data
    metadata    map[string]interface{}  // Additional context

    // Event Source
    source            Source       // system, agent, webhook, workflow, etc.
    triggeredBy       *uuid.UUID   // Agent who triggered (if applicable)
    integrationSource *string      // External integration name

    // Delivery & Read Status
    isRealtime  bool       // Should be delivered in real-time
    delivered   bool       // Has been delivered
    deliveredAt *time.Time
    read        bool       // Has been read
    readAt      *time.Time

    // Visibility & Expiration
    visibleToClient bool       // Show to contact
    visibleToAgent  bool       // Show to agent
    expiresAt       *time.Time // Event expiration

    occurredAt time.Time  // When event actually occurred
    createdAt  time.Time  // When event was created
}
```

### Value Objects

#### 1. Category

```go
type Category string
const (
    CategoryGeneral      Category = "general"       // Generic events
    CategoryStatus       Category = "status"        // Status changes
    CategoryPipeline     Category = "pipeline"      // Pipeline events
    CategoryAssignment   Category = "assignment"    // Agent assignment
    CategoryTag          Category = "tag"           // Tag changes
    CategoryNote         Category = "note"          // Note events
    CategorySession      Category = "session"       // Session lifecycle
    CategoryCustomField  Category = "custom_field"  // Custom field changes
    CategorySystem       Category = "system"        // System events
    CategoryNotification Category = "notification"  // Notifications
    CategoryTracking     Category = "tracking"      // Tracking events
)
```

#### 2. Priority

```go
type Priority string
const (
    PriorityLow    Priority = "low"     // Low priority
    PriorityNormal Priority = "normal"  // Normal priority (default)
    PriorityHigh   Priority = "high"    // High priority
    PriorityUrgent Priority = "urgent"  // Urgent - immediate attention
)
```

#### 3. Source

```go
type Source string
const (
    SourceSystem      Source = "system"       // System-generated
    SourceAgent       Source = "agent"        // Agent action
    SourceWebhook     Source = "webhook"      // External webhook
    SourceWorkflow    Source = "workflow"     // Temporal workflow
    SourceAutomation  Source = "automation"   // Automation rule
    SourceIntegration Source = "integration"  // External integration
)
```

### Predefined Event Types

```go
// Status Events
const (
    EventTypeStatusChanged        = "status_changed"
    EventTypeEnteredPipeline      = "entered_pipeline"
    EventTypeExitedPipeline       = "exited_pipeline"
    EventTypePipelineStageChanged = "pipeline_stage_changed"
)

// Assignment Events
const (
    EventTypeAgentAssigned    = "agent_assigned"
    EventTypeAgentTransferred = "agent_transferred"
    EventTypeAgentUnassigned  = "agent_unassigned"
)

// Tag Events
const (
    EventTypeTagAdded   = "tag_added"
    EventTypeTagRemoved = "tag_removed"
)

// Note Events
const (
    EventTypeNoteAdded   = "note_added"
    EventTypeNoteUpdated = "note_updated"
    EventTypeNoteDeleted = "note_deleted"
)

// Session Events
const (
    EventTypeSessionStarted = "session_started"
    EventTypeSessionEnded   = "session_ended"
)

// Custom Field Events
const (
    EventTypeCustomFieldSet     = "custom_field_set"
    EventTypeCustomFieldCleared = "custom_field_cleared"
)

// Other Events
const (
    EventTypeWebhookReceived  = "webhook_received"
    EventTypeNotificationSent = "notification_sent"
    EventTypeContactCreated   = "contact_created"
    EventTypeContactUpdated   = "contact_updated"
    EventTypeContactMerged    = "contact_merged"
    EventTypeContactEnriched  = "contact_enriched"
)
```

### Business Invariants

1. **Event must belong to Contact**
   - `contactID` and `tenantID` required
   - `eventType`, `category`, `priority`, `source` required

2. **Session is optional**
   - Events can exist without session (e.g., contact created before first message)

3. **Visibility rules**
   - Can control visibility to client vs agent separately
   - Expired events not shown to either

4. **Delivery tracking**
   - Real-time events must be delivered to active clients
   - Read receipts tracked for notifications

5. **Expiration**
   - Events can have expiration time (e.g., temporary notifications)
   - Expired events can be auto-deleted

---

## Events (Not Emitted - This IS the Event)

ContactEvent **is itself an event entity**, not an aggregate that emits events. It represents business events that happened to contacts.

---

## Repository Interface

```go
type Repository interface {
    Save(ctx context.Context, event *ContactEvent) error
    Update(ctx context.Context, event *ContactEvent) error
    FindByID(ctx context.Context, id uuid.UUID) (*ContactEvent, error)
    FindByContactID(ctx context.Context, contactID uuid.UUID, limit int, offset int) ([]*ContactEvent, error)
    FindByContactIDVisible(ctx context.Context, contactID uuid.UUID, visibleToClient bool, limit int, offset int) ([]*ContactEvent, error)
    FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit int, offset int) ([]*ContactEvent, error)
    FindUndeliveredRealtime(ctx context.Context, limit int) ([]*ContactEvent, error)
    FindUndeliveredForContact(ctx context.Context, contactID uuid.UUID) ([]*ContactEvent, error)
    FindByTenantAndType(ctx context.Context, tenantID string, eventType string, since time.Time, limit int) ([]*ContactEvent, error)
    FindByCategory(ctx context.Context, tenantID string, category Category, since time.Time, limit int) ([]*ContactEvent, error)
    FindExpired(ctx context.Context, before time.Time, limit int) ([]*ContactEvent, error)
    Delete(ctx context.Context, id uuid.UUID) error
    DeleteExpired(ctx context.Context, before time.Time) (int, error)
    CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)
}
```

---

## Commands (CQRS)

### ✅ Implemented

1. **CreateContactEventCommand**
2. **MarkAsDeliveredCommand**
3. **MarkAsReadCommand**

### ❌ Suggested

- **CreateBulkEventsCommand** - Bulk event creation
- **DeleteExpiredEventsCommand** - Cleanup old events
- **ExportContactTimelineCommand** - Export as PDF/CSV

---

## Use Cases

### ✅ Implemented

None explicitly (events created inline in other use cases)

### ❌ Suggested

1. **CreateContactEventUseCase** - Create new contact event
2. **GetContactTimelineUseCase** - Get full activity timeline
3. **DeliverRealtimeEventsUseCase** - Push events via WebSocket
4. **MarkEventAsReadUseCase** - Track event read status
5. **CleanupExpiredEventsUseCase** - Auto-delete expired events
6. **ExportContactTimelineUseCase** - Export to PDF/CSV
7. **SearchContactEventsUseCase** - Search across events

---

## Real-World Usage

### Scenario 1: Pipeline Status Change

```go
// Contact moved from "Lead" to "Qualified"
event, _ := contact_event.NewContactEvent(
    contactID,
    tenantID,
    contact_event.EventTypePipelineStageChanged,
    contact_event.CategoryPipeline,
    contact_event.PriorityNormal,
    contact_event.SourceAgent,
)

event.SetTitle("Pipeline Stage Changed")
event.SetDescription("Contact moved from Lead to Qualified")
event.AddPayloadField("old_status", "Lead")
event.AddPayloadField("new_status", "Qualified")
event.AddPayloadField("pipeline_id", pipelineID)
event.SetTriggeredBy(agentID)
event.SetVisibility(false, true)  // Visible only to agents

eventRepo.Save(ctx, event)
```

### Scenario 2: Agent Assignment

```go
// Agent assigned to contact
event, _ := contact_event.NewContactEvent(
    contactID,
    tenantID,
    contact_event.EventTypeAgentAssigned,
    contact_event.CategoryAssignment,
    contact_event.PriorityNormal,
    contact_event.SourceSystem,
)

event.SetTitle("Agent Assigned")
event.SetDescription("John Doe assigned to this contact")
event.AddPayloadField("agent_id", agentID)
event.AddPayloadField("agent_name", "John Doe")
event.SetVisibility(false, true)  // Visible only to agents

eventRepo.Save(ctx, event)
```

### Scenario 3: Real-time Notification

```go
// Urgent notification to contact
event, _ := contact_event.NewContactEvent(
    contactID,
    tenantID,
    contact_event.EventTypeNotificationSent,
    contact_event.CategoryNotification,
    contact_event.PriorityUrgent,
    contact_event.SourceSystem,
)

event.SetTitle("Payment Reminder")
event.SetDescription("Your invoice #12345 is due tomorrow")
event.AddPayloadField("invoice_id", "12345")
event.AddPayloadField("amount", 199.90)
event.AddPayloadField("due_date", "2025-10-11")
event.SetRealtimeDelivery(true)
event.SetVisibility(true, true)  // Visible to both
event.SetExpiresAt(time.Now().Add(24 * time.Hour))

eventRepo.Save(ctx, event)

// Worker picks up undelivered realtime events
events := eventRepo.FindUndeliveredRealtime(ctx, 100)
for _, evt := range events {
    // Push via WebSocket
    websocketHub.SendToContact(evt.ContactID(), evt)
    evt.MarkAsDelivered()
    eventRepo.Update(ctx, evt)
}
```

### Scenario 4: Session Timeline

```go
// Get all events for a session
events, _ := eventRepo.FindBySessionID(ctx, sessionID, 50, 0)

// Display in agent UI:
// 14:30 - Session Started
// 14:31 - Agent Assigned (John Doe)
// 14:32 - Tag Added: urgent
// 14:35 - Note Added: "Customer wants refund"
// 14:40 - Status Changed: Lead → Qualified
// 14:45 - Session Ended
```

### Scenario 5: Contact Enrichment

```go
// Contact data enriched from external API
event, _ := contact_event.NewContactEvent(
    contactID,
    tenantID,
    contact_event.EventTypeContactEnriched,
    contact_event.CategorySystem,
    contact_event.PriorityLow,
    contact_event.SourceIntegration,
)

event.SetTitle("Contact Data Enriched")
event.SetDescription("Data enriched from Clearbit API")
event.AddPayloadField("enrichment_source", "clearbit")
event.AddPayloadField("fields_updated", []string{"company", "industry", "employee_count"})
event.SetIntegrationSource("clearbit")
event.SetVisibility(false, true)  // Visible only to agents

eventRepo.Save(ctx, event)
```

---

## API Examples

### Get Contact Timeline

```http
GET /api/v1/contacts/{id}/events?limit=50&offset=0

Response:
{
  "events": [
    {
      "id": "uuid",
      "event_type": "pipeline_stage_changed",
      "category": "pipeline",
      "priority": "normal",
      "title": "Pipeline Stage Changed",
      "description": "Contact moved from Lead to Qualified",
      "payload": {
        "old_status": "Lead",
        "new_status": "Qualified",
        "pipeline_id": "uuid"
      },
      "source": "agent",
      "triggered_by": "uuid",
      "visible_to_client": false,
      "visible_to_agent": true,
      "occurred_at": "2025-10-10T14:35:00Z",
      "created_at": "2025-10-10T14:35:01Z"
    },
    {
      "id": "uuid",
      "event_type": "agent_assigned",
      "category": "assignment",
      "priority": "normal",
      "title": "Agent Assigned",
      "payload": {
        "agent_name": "John Doe"
      },
      "occurred_at": "2025-10-10T14:31:00Z"
    }
  ],
  "total": 42,
  "page": 1,
  "limit": 50
}
```

### Get Session Events

```http
GET /api/v1/sessions/{id}/events

Response:
{
  "events": [
    {
      "event_type": "session_started",
      "title": "Session Started",
      "occurred_at": "2025-10-10T14:30:00Z"
    },
    {
      "event_type": "agent_assigned",
      "title": "Agent Assigned",
      "payload": { "agent_name": "John Doe" },
      "occurred_at": "2025-10-10T14:31:00Z"
    },
    {
      "event_type": "session_ended",
      "title": "Session Ended",
      "occurred_at": "2025-10-10T14:45:00Z"
    }
  ]
}
```

### Mark Event as Read

```http
POST /api/v1/contact-events/{id}/mark-read

Response:
{
  "success": true,
  "event_id": "uuid",
  "read_at": "2025-10-10T15:00:00Z"
}
```

### Get Undelivered Events

```http
GET /api/v1/contact-events/undelivered?contact_id=uuid

Response:
{
  "events": [
    {
      "id": "uuid",
      "event_type": "notification_sent",
      "title": "Payment Reminder",
      "priority": "urgent",
      "is_realtime": true,
      "delivered": false,
      "expires_at": "2025-10-11T00:00:00Z"
    }
  ]
}
```

---

## Performance Considerations

### Indexes

```sql
-- Contact Events
CREATE INDEX idx_contact_events_contact ON contact_events(contact_id, occurred_at DESC);
CREATE INDEX idx_contact_events_session ON contact_events(session_id, occurred_at DESC);
CREATE INDEX idx_contact_events_tenant_type ON contact_events(tenant_id, event_type, occurred_at DESC);
CREATE INDEX idx_contact_events_category ON contact_events(tenant_id, category, occurred_at DESC);
CREATE INDEX idx_contact_events_undelivered ON contact_events(is_realtime, delivered)
    WHERE is_realtime = true AND delivered = false;
CREATE INDEX idx_contact_events_expired ON contact_events(expires_at)
    WHERE expires_at IS NOT NULL;
CREATE INDEX idx_contact_events_visibility ON contact_events(contact_id, visible_to_client, occurred_at DESC);
```

### Caching Strategy

```go
// Cache contact timeline (3 min TTL)
cacheKey := fmt.Sprintf("contact:%s:timeline", contactID)
events, err := cache.Get(cacheKey)

// Cache session events (5 min TTL)
cacheKey := fmt.Sprintf("session:%s:events", sessionID)
events, err := cache.Get(cacheKey)
```

### Data Retention

```go
// Cleanup expired events daily
func CleanupExpiredEvents(ctx context.Context) error {
    now := time.Now()

    // Delete events expired more than 7 days ago
    deletedCount, err := eventRepo.DeleteExpired(ctx, now.Add(-7*24*time.Hour))

    log.Info("Expired events deleted", zap.Int("count", deletedCount))
    return err
}
```

---

## Worker: Real-time Event Delivery

```go
type RealtimeEventWorker struct {
    eventRepo     contact_event.Repository
    websocketHub  *websocket.Hub
    pollInterval  time.Duration
}

func (w *RealtimeEventWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.pollInterval)  // e.g., 1 second

    for {
        select {
        case <-ticker.C:
            // Find undelivered realtime events
            events, _ := w.eventRepo.FindUndeliveredRealtime(ctx, 100)

            for _, event := range events {
                // Skip expired events
                if event.IsExpired() {
                    event.MarkAsDelivered()  // Mark as delivered to skip
                    w.eventRepo.Update(ctx, event)
                    continue
                }

                // Send via WebSocket
                err := w.websocketHub.SendToContact(event.ContactID(), event)
                if err != nil {
                    log.Error("failed to deliver event", zap.Error(err))
                    continue
                }

                // Mark as delivered
                event.MarkAsDelivered()
                w.eventRepo.Update(ctx, event)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

---

## Use Case Patterns

### Pattern 1: Create Event from Domain Event

```go
// When domain event occurs, create contact event
func CreateContactEventFromDomainEvent(domainEvent shared.DomainEvent) {
    switch e := domainEvent.(type) {
    case contact.PipelineStatusChangedEvent:
        event, _ := contact_event.NewContactEvent(
            e.ContactID,
            e.TenantID,
            contact_event.EventTypePipelineStageChanged,
            contact_event.CategoryPipeline,
            contact_event.PriorityNormal,
            contact_event.SourceSystem,
        )
        event.AddPayloadField("old_status_id", e.OldStatusID)
        event.AddPayloadField("new_status_id", e.NewStatusID)
        eventRepo.Save(ctx, event)

    case agent.AgentAssignedEvent:
        event, _ := contact_event.NewContactEvent(
            e.ContactID,
            e.TenantID,
            contact_event.EventTypeAgentAssigned,
            contact_event.CategoryAssignment,
            contact_event.PriorityNormal,
            contact_event.SourceSystem,
        )
        event.AddPayloadField("agent_id", e.AgentID)
        eventRepo.Save(ctx, event)
    }
}
```

---

## References

- [ContactEvent Domain](../../internal/domain/contact_event/)
- [ContactEvent Repository](../../infrastructure/persistence/gorm_contact_event_repository.go)
- [ContactEvent Types](../../internal/domain/contact_event/types.go)

---

**Next**: [Event Aggregate](event_aggregate.md) →
**Previous**: [Tracking Aggregate](tracking_aggregate.md) ←

---

## Summary

✅ **ContactEvent Aggregate Features**:
1. **Activity timeline** - Chronological log of all contact events
2. **Multiple categories** - Pipeline, assignment, tags, notes, sessions, etc.
3. **Priority levels** - Low, normal, high, urgent
4. **Real-time delivery** - Push notifications via WebSocket
5. **Visibility control** - Separate visibility for clients vs agents
6. **Expiration** - Auto-expire temporary notifications
7. **Audit trail** - Track who triggered each event
8. **Rich payload** - Custom data per event type

The ContactEvent aggregate provides a **complete activity log** for contacts, enabling agents to understand the full customer journey and providing audit trails for compliance.
