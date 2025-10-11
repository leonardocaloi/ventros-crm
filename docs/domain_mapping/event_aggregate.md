# Event Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~200
**Test Coverage**: Unknown

---

## Overview

- **Purpose**: Generic event logging for analytics and integrations
- **Location**: `internal/domain/event/`
- **Entity**: `infrastructure/persistence/entities/event.go`
- **Repository**: Not implemented yet
- **Aggregate Root**: `Event`

**Business Problem**:
The Event aggregate provides **generic event storage** for analytics, webhooks, and third-party integrations. Unlike ContactEvent (which tracks contact-specific activities), Event is a lightweight, flexible event log for system-wide events. Critical for:
- **Analytics** - Track user behavior and system events
- **Webhook delivery** - Log webhook events from external systems
- **Integration events** - Store events from Temporal workflows, cron jobs
- **Event sourcing** - Potential future migration to event sourcing
- **Audit log** - Generic audit trail for non-contact events
- **Debugging** - Track system events for troubleshooting

---

## Domain Model

### Aggregate Root: Event

```go
type Event struct {
    id             uuid.UUID
    contactID      *uuid.UUID  // Optional - event may not relate to contact
    sessionID      *uuid.UUID  // Optional - event may not relate to session
    messageID      *uuid.UUID  // Optional - event may not relate to message
    tenantID       string
    eventType      string      // "contact.created", "message.sent", etc.
    payload        map[string]interface{}  // Event-specific data
    source         EventSource // system, webhook, manual, cron, workflow
    sequenceNumber *int        // Optional sequence for ordering
    timestamp      time.Time   // When event occurred
    createdAt      time.Time   // When event was stored
}
```

### Value Objects

#### EventSource

```go
type EventSource string
const (
    EventSourceSystem   EventSource = "system"    // System-generated
    EventSourceWebhook  EventSource = "webhook"   // External webhook
    EventSourceManual   EventSource = "manual"    // Manual action
    EventSourceCron     EventSource = "cron"      // Cron job
    EventSourceWorkflow EventSource = "workflow"  // Temporal workflow
)
```

### Predefined Event Types

```go
// Contact Events
const (
    EventTypeContactCreated = "contact.created"
    EventTypeContactUpdated = "contact.updated"
    EventTypeContactDeleted = "contact.deleted"
)

// Session Events
const (
    EventTypeSessionStarted    = "session.started"
    EventTypeSessionEnded      = "session.ended"
    EventTypeSessionSummarized = "session.summarized"
)

// Message Events
const (
    EventTypeMessageReceived = "message.received"
    EventTypeMessageSent     = "message.sent"
    EventTypeMessageRead     = "message.read"
    EventTypeMessageFailed   = "message.failed"
)

// Agent Events
const (
    EventTypeAgentAssigned = "agent.assigned"
    EventTypeAgentTransfer = "agent.transfer"
    EventTypeAgentTyping   = "agent.typing"
)

// Other Events
const (
    EventTypeCustomFieldUpdated = "custom_field.updated"
    EventTypeTagAdded           = "tag.added"
    EventTypeTagRemoved         = "tag.removed"
    EventTypeNoteAdded          = "note.added"
)
```

### Business Invariants

1. **Tenant required**
   - `tenantID` required for multi-tenancy
   - `eventType` and `source` required

2. **References are optional**
   - `contactID`, `sessionID`, `messageID` all optional
   - Events can exist without specific entity references

3. **Sequence numbers**
   - Optional for ordering events
   - Useful for event sourcing patterns

4. **Timestamp vs CreatedAt**
   - `timestamp`: When event actually occurred
   - `createdAt`: When event was persisted to database

---

## Events (Not Emitted - This IS the Event)

Event **is itself an event entity**, not an aggregate that emits events. It represents system-wide events that happened.

---

## Repository Interface

```go
type Repository interface {
    Save(ctx context.Context, event *Event) error
    FindByID(ctx context.Context, id uuid.UUID) (*Event, error)
    FindByContact(ctx context.Context, contactID uuid.UUID, limit int) ([]*Event, error)
    FindBySession(ctx context.Context, sessionID uuid.UUID) ([]*Event, error)
    FindByTenantAndType(ctx context.Context, tenantID, eventType string, limit int) ([]*Event, error)
    FindByTimeRange(ctx context.Context, tenantID string, start, end time.Time) ([]*Event, error)
    CountByContact(ctx context.Context, contactID uuid.UUID) (int, error)
    CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error)
}
```

**Note**: Repository interface defined but **GORM implementation not created yet**.

---

## Commands (CQRS)

### ✅ Implemented

None explicitly (events created inline)

### ❌ Suggested

- **CreateEventCommand**
- **BulkCreateEventsCommand**
- **QueryEventsByTypeCommand**
- **ExportEventsCommand**

---

## Use Cases

### ✅ Implemented

None explicitly

### ❌ Suggested

1. **LogEventUseCase** - Log generic system event
2. **GetEventStreamUseCase** - Get events for analytics
3. **ReplayEventsUseCase** - Replay events for debugging
4. **ExportEventsUseCase** - Export events to external system
5. **CountEventsByTypeUseCase** - Analytics aggregations

---

## Differences: Event vs ContactEvent

| Aspect | Event | ContactEvent |
|--------|-------|--------------|
| **Purpose** | Generic system events | Contact-specific activities |
| **Scope** | System-wide | Contact-scoped |
| **Required Links** | None (all optional) | Contact required |
| **Visibility** | Internal only | Can be shown to clients/agents |
| **Delivery** | No delivery tracking | Real-time delivery support |
| **Expiration** | No expiration | Can expire |
| **Priority** | No priority | Has priority levels |
| **Category** | No category | 11 predefined categories |
| **Use Cases** | Analytics, webhooks, integrations | Activity timeline, notifications |

**Rule of Thumb**:
- Use **ContactEvent** for customer-facing activities that agents/contacts should see
- Use **Event** for system events, webhooks, background jobs, analytics

---

## Real-World Usage

### Scenario 1: Webhook Event from External System

```go
// Receive webhook from external CRM
event, _ := event.NewEvent(
    tenantID,
    "webhook.external_crm.contact_updated",
    event.EventSourceWebhook,
    map[string]interface{}{
        "webhook_id":     "wh_123",
        "source_system":  "external_crm",
        "contact_email":  "john@example.com",
        "fields_updated": []string{"name", "company"},
        "raw_payload":    webhookPayload,
    },
)

event.AttachToContact(contactID)
eventRepo.Save(ctx, event)
```

### Scenario 2: Cron Job Event

```go
// Daily cleanup job
event, _ := event.NewEvent(
    tenantID,
    "cron.cleanup.expired_sessions",
    event.EventSourceCron,
    map[string]interface{}{
        "job_name":          "cleanup_expired_sessions",
        "sessions_deleted":  42,
        "execution_time_ms": 1234,
        "next_run":          nextRunTime,
    },
)

eventRepo.Save(ctx, event)
```

### Scenario 3: Temporal Workflow Event

```go
// Workflow completed
event, _ := event.NewEvent(
    tenantID,
    "workflow.session_lifecycle.completed",
    event.EventSourceWorkflow,
    map[string]interface{}{
        "workflow_id":        workflowID,
        "session_id":         sessionID.String(),
        "duration_seconds":   300,
        "activities_run":     5,
        "status":             "success",
    },
)

event.AttachToSession(sessionID)
eventRepo.Save(ctx, event)
```

### Scenario 4: System Event

```go
// Database migration completed
event, _ := event.NewEvent(
    "system",  // System-level event
    "system.migration.completed",
    event.EventSourceSystem,
    map[string]interface{}{
        "migration_version": "000042",
        "migration_name":    "add_chats_table",
        "duration_ms":       543,
        "rows_affected":     0,
    },
)

eventRepo.Save(ctx, event)
```

### Scenario 5: Analytics Event

```go
// Track user behavior for analytics
event, _ := event.NewEvent(
    tenantID,
    "analytics.page_view",
    event.EventSourceSystem,
    map[string]interface{}{
        "page":       "/contacts",
        "user_id":    agentID.String(),
        "session_id": webSessionID,
        "referrer":   "/dashboard",
        "duration":   5000,  // ms
    },
)

eventRepo.Save(ctx, event)
```

---

## API Examples

### Log Event

```http
POST /api/v1/events
{
  "event_type": "webhook.external_crm.contact_updated",
  "source": "webhook",
  "contact_id": "uuid",
  "payload": {
    "webhook_id": "wh_123",
    "source_system": "external_crm",
    "fields_updated": ["name", "company"]
  }
}

Response:
{
  "id": "uuid",
  "event_type": "webhook.external_crm.contact_updated",
  "source": "webhook",
  "timestamp": "2025-10-10T15:00:00Z",
  "created_at": "2025-10-10T15:00:01Z"
}
```

### Get Events by Type

```http
GET /api/v1/events?type=message.sent&limit=100

Response:
{
  "events": [
    {
      "id": "uuid",
      "event_type": "message.sent",
      "source": "system",
      "contact_id": "uuid",
      "session_id": "uuid",
      "message_id": "uuid",
      "payload": {
        "channel": "whatsapp",
        "direction": "outbound"
      },
      "timestamp": "2025-10-10T15:00:00Z"
    }
  ],
  "total": 1234
}
```

### Get Events by Time Range

```http
GET /api/v1/events?start=2025-10-01&end=2025-10-10

Response:
{
  "events": [...],
  "total": 5678,
  "time_range": {
    "start": "2025-10-01T00:00:00Z",
    "end": "2025-10-10T23:59:59Z"
  }
}
```

### Get Contact Events

```http
GET /api/v1/contacts/{id}/events?limit=50

Response:
{
  "events": [
    {
      "event_type": "contact.created",
      "timestamp": "2025-10-01T10:00:00Z"
    },
    {
      "event_type": "session.started",
      "timestamp": "2025-10-01T10:05:00Z"
    },
    {
      "event_type": "message.received",
      "timestamp": "2025-10-01T10:06:00Z"
    }
  ]
}
```

---

## Performance Considerations

### Indexes

```sql
-- Events
CREATE INDEX idx_events_tenant ON events(tenant_id, timestamp DESC);
CREATE INDEX idx_events_type ON events(event_type, timestamp DESC);
CREATE INDEX idx_events_source ON events(source, timestamp DESC);
CREATE INDEX idx_events_contact ON events(contact_id, timestamp DESC);
CREATE INDEX idx_events_session ON events(session_id, timestamp DESC);
CREATE INDEX idx_events_message ON events(message_id);
CREATE INDEX idx_events_time_range ON events(tenant_id, timestamp);

-- Composite for analytics queries
CREATE INDEX idx_events_analytics ON events(tenant_id, event_type, source, timestamp);
```

### Partitioning Strategy

```sql
-- Partition by month for better query performance
CREATE TABLE events (
    id UUID PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    -- ... other columns
) PARTITION BY RANGE (timestamp);

-- Create partitions
CREATE TABLE events_2025_10 PARTITION OF events
    FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');

CREATE TABLE events_2025_11 PARTITION OF events
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
```

### Data Retention

```go
// Archive old events to cold storage
func ArchiveOldEvents(ctx context.Context, before time.Time) error {
    // 1. Export events older than retention period
    events, _ := eventRepo.FindByTimeRange(ctx, tenantID, time.Time{}, before)

    // 2. Write to S3/cold storage
    s3.Upload("events-archive", events)

    // 3. Delete from database
    for _, event := range events {
        eventRepo.Delete(ctx, event.ID())
    }

    return nil
}
```

---

## Use Case: Analytics Aggregations

```go
// Count events by type for dashboard
func GetEventStatistics(ctx context.Context, tenantID string, since time.Time) map[string]int {
    stats := make(map[string]int)

    eventTypes := []string{
        event.EventTypeContactCreated,
        event.EventTypeSessionStarted,
        event.EventTypeMessageReceived,
        event.EventTypeMessageSent,
    }

    for _, eventType := range eventTypes {
        events, _ := eventRepo.FindByTenantAndType(ctx, tenantID, eventType, since, 10000)
        stats[eventType] = len(events)
    }

    return stats
}

// Result:
// {
//   "contact.created": 123,
//   "session.started": 456,
//   "message.received": 1234,
//   "message.sent": 987
// }
```

---

## Use Case: Event Sourcing Pattern

```go
// Rebuild contact state from events
func RebuildContactState(ctx context.Context, contactID uuid.UUID) (*Contact, error) {
    events, _ := eventRepo.FindByContact(ctx, contactID, 10000)

    var contact *Contact

    for _, event := range events {
        switch event.EventType() {
        case event.EventTypeContactCreated:
            name := event.Payload()["name"].(string)
            phone := event.Payload()["phone"].(string)
            contact = NewContact(name, phone, tenantID)

        case event.EventTypeContactUpdated:
            updates := event.Payload()
            contact.Update(updates)

        case event.EventTypeCustomFieldUpdated:
            key := event.Payload()["key"].(string)
            value := event.Payload()["value"]
            contact.SetCustomField(key, value)

        case event.EventTypeTagAdded:
            tag := event.Payload()["tag"].(string)
            contact.AddTag(tag)
        }
    }

    return contact, nil
}
```

---

## Integration with Analytics Systems

### Export to Data Warehouse

```go
// Daily export to BigQuery/Snowflake
func ExportEventsToWarehouse(ctx context.Context, date time.Date) error {
    start := date.StartOfDay()
    end := date.EndOfDay()

    events, _ := eventRepo.FindByTimeRange(ctx, tenantID, start, end)

    // Convert to warehouse schema
    rows := make([]WarehouseRow, len(events))
    for i, event := range events {
        rows[i] = WarehouseRow{
            EventID:    event.ID().String(),
            TenantID:   event.TenantID(),
            EventType:  event.EventType(),
            Source:     event.Source().String(),
            Payload:    json.Marshal(event.Payload()),
            Timestamp:  event.Timestamp(),
        }
    }

    // Bulk insert to warehouse
    return warehouse.BulkInsert("events", rows)
}
```

---

## Implementation Status

### ✅ What's Implemented

1. Domain model (`Event` struct)
2. Value objects (`EventSource`)
3. Business logic methods
4. Repository interface
5. Predefined event type constants

### ❌ What's Missing

1. **GORM repository implementation** - No persistence layer yet
2. **Use cases** - No explicit use cases created
3. **HTTP handlers** - No API endpoints
4. **Migrations** - No database table
5. **Tests** - No unit tests
6. **Workers** - No background processing

---

## Suggested Implementation Roadmap

### Phase 1: Foundation (1-2 days)
- [ ] Create database migration (`CREATE TABLE events`)
- [ ] Implement GormEventRepository
- [ ] Create HTTP handlers
- [ ] Add unit tests

### Phase 2: Use Cases (1-2 days)
- [ ] LogEventUseCase
- [ ] GetEventStreamUseCase
- [ ] CountEventsByTypeUseCase

### Phase 3: Analytics (2-3 days)
- [ ] Aggregation queries
- [ ] Export to data warehouse
- [ ] Dashboard endpoints

### Phase 4: Optimization (1-2 days)
- [ ] Table partitioning
- [ ] Archive old events
- [ ] Performance tuning

---

## References

- [Event Domain](../../internal/domain/event/)
- [Event Repository Interface](../../internal/domain/event/repository.go)
- [Event Types](../../internal/domain/event/types.go)

---

**Next**: [Project Aggregate](project_aggregate.md) →
**Previous**: [ContactEvent Aggregate](contact_event_aggregate.md) ←

---

## Summary

✅ **Event Aggregate Design**:
1. **Generic event storage** - System-wide events, not contact-specific
2. **Flexible payload** - Store any JSON data
3. **Multiple sources** - System, webhook, cron, workflow, manual
4. **Optional references** - Can link to contact/session/message
5. **Analytics-ready** - Designed for aggregations and exports
6. **Event sourcing** - Foundation for future event sourcing migration

❌ **Implementation Status**: Domain model complete, but **persistence layer not implemented yet**.

**Next Steps**: Implement GormEventRepository and create database migration to enable event persistence.

**Use Case**: The Event aggregate serves as a **lightweight, flexible event log** for system events, webhooks, and analytics - complementing ContactEvent which focuses on customer-facing activities.
