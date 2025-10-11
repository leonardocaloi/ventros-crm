# Broadcast Aggregate

**Last Updated**: 2025-10-10
**Status**: ❌ NOT IMPLEMENTED (Design Only)
**Lines of Code**: 0 (Design document exists)
**Test Coverage**: N/A

---

## Overview

- **Purpose**: Mass message broadcasting to contact lists
- **Location**: `internal/domain/broadcast/` (planned)
- **Entity**: Not implemented yet
- **Repository**: Not implemented yet
- **Aggregate Roots**: `ContactList`, `Broadcast`, `BroadcastExecution`

**Business Problem**:
The Broadcast aggregate will enable **mass message campaigns** to targeted contact lists. It supports scheduled and immediate broadcasts with personalized templates, rate limiting, and detailed tracking. Critical for:
- **Marketing campaigns** - Send promotional messages to customer segments
- **Newsletters** - Scheduled weekly/monthly newsletters
- **Announcements** - System-wide notifications
- **Cart recovery** - Automated abandoned cart messages
- **Customer engagement** - Re-engagement campaigns for inactive contacts

**Implementation Status**: 📋 Design document exists at `/internal/domain/broadcast/broadcast.md` but no code implemented yet.

---

## Domain Model (Planned)

### Aggregate Root 1: ContactList

```go
type ContactList struct {
    id          uuid.UUID
    tenantID    string
    name        string
    description string
    tags        []string        // Tags for filtering
    filters     ListFilters     // Dynamic filter criteria
    contactIDs  []uuid.UUID     // Fixed IDs (snapshot)
    isDynamic   bool            // If true, recalculates contacts dynamically
    createdAt   time.Time
    updatedAt   time.Time
}

// ListFilters - Criteria for dynamic lists
type ListFilters struct {
    PipelineID   *uuid.UUID             `json:"pipeline_id,omitempty"`
    StatusID     *uuid.UUID             `json:"status_id,omitempty"`
    Tags         []string               `json:"tags,omitempty"`
    CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}
```

**Types of Lists**:
1. **Static Lists** - Fixed contact IDs (e.g., "VIP Customers - 100 IDs")
2. **Dynamic Lists** - Query-based (e.g., "Qualified Leads" = `pipeline_id=X AND status_id=Y`)

### Aggregate Root 2: Broadcast

```go
type Broadcast struct {
    id              uuid.UUID
    tenantID        string
    name            string
    listID          uuid.UUID         // Target contact list
    messageTemplate MessageTemplate   // Message template with variables
    status          BroadcastStatus
    scheduledFor    *time.Time        // When to send (nil = immediate)
    startedAt       *time.Time
    completedAt     *time.Time

    // Statistics
    totalContacts   int
    sentCount       int
    failedCount     int
    pendingCount    int

    // Rate limiting
    rateLimit       int  // Messages per minute (0 = unlimited)

    createdAt       time.Time
    updatedAt       time.Time

    events []DomainEvent
}

type BroadcastStatus string
const (
    BroadcastStatusDraft      BroadcastStatus = "draft"      // Draft
    BroadcastStatusScheduled  BroadcastStatus = "scheduled"  // Scheduled
    BroadcastStatusRunning    BroadcastStatus = "running"    // In progress
    BroadcastStatusCompleted  BroadcastStatus = "completed"  // Completed
    BroadcastStatusFailed     BroadcastStatus = "failed"     // Failed
    BroadcastStatusCancelled  BroadcastStatus = "cancelled"  // Cancelled
)

// MessageTemplate - Message template with variables
type MessageTemplate struct {
    Type       string            `json:"type"`        // text, template, media
    Content    string            `json:"content"`
    TemplateID *string           `json:"template_id,omitempty"`
    Variables  map[string]string `json:"variables,omitempty"`
    MediaURL   *string           `json:"media_url,omitempty"`
}
```

**Lifecycle**:
```
draft → scheduled → running → completed
                  ↓
               cancelled
```

### Entity: BroadcastExecution

```go
// BroadcastExecution - Tracks individual send per contact
type BroadcastExecution struct {
    id          uuid.UUID
    broadcastID uuid.UUID
    contactID   uuid.UUID
    status      ExecutionStatus
    messageID   *uuid.UUID   // ID of sent message
    error       *string
    sentAt      *time.Time
    createdAt   time.Time
}

type ExecutionStatus string
const (
    ExecutionStatusPending   ExecutionStatus = "pending"
    ExecutionStatusSending   ExecutionStatus = "sending"
    ExecutionStatusSent      ExecutionStatus = "sent"
    ExecutionStatusFailed    ExecutionStatus = "failed"
    ExecutionStatusSkipped   ExecutionStatus = "skipped"
)
```

---

## Events (Planned)

### ContactList Events

| Event | When | Purpose |
|-------|------|---------|
| `contact_list.created` | New list created | Initialize list |
| `contact_list.updated` | List modified | Sync changes |
| `contact_list.contacts_added` | Contacts added | Track additions |
| `contact_list.contacts_removed` | Contacts removed | Track removals |
| `contact_list.deleted` | List deleted | Cleanup |

### Broadcast Events

| Event | When | Purpose |
|-------|------|---------|
| `broadcast.created` | New broadcast created | Initialize broadcast |
| `broadcast.scheduled` | Broadcast scheduled | Set timer |
| `broadcast.started` | Execution started | Track start |
| `broadcast.message_sent` | Message sent to contact | Track progress |
| `broadcast.message_failed` | Message failed | Log error |
| `broadcast.completed` | All messages processed | Finalize |
| `broadcast.cancelled` | Broadcast cancelled | Stop execution |

---

## Repository Interfaces (Planned)

### ContactListRepository

```go
type ContactListRepository interface {
    Save(ctx context.Context, list *ContactList) error
    FindByID(ctx context.Context, id uuid.UUID) (*ContactList, error)
    FindByTenant(ctx context.Context, tenantID string) ([]*ContactList, error)
    Delete(ctx context.Context, id uuid.UUID) error

    // Contact management
    AddContacts(ctx context.Context, listID uuid.UUID, contactIDs []uuid.UUID) error
    RemoveContacts(ctx context.Context, listID uuid.UUID, contactIDs []uuid.UUID) error
    ResolveContacts(ctx context.Context, listID uuid.UUID) ([]*Contact, error)
}
```

### BroadcastRepository

```go
type BroadcastRepository interface {
    Save(ctx context.Context, broadcast *Broadcast) error
    FindByID(ctx context.Context, id uuid.UUID) (*Broadcast, error)
    FindByTenant(ctx context.Context, tenantID string) ([]*Broadcast, error)
    FindScheduledReady(ctx context.Context) ([]*Broadcast, error) // For scheduler worker
    Delete(ctx context.Context, id uuid.UUID) error
}
```

### BroadcastExecutionRepository

```go
type BroadcastExecutionRepository interface {
    SaveBatch(ctx context.Context, executions []*BroadcastExecution) error
    Save(ctx context.Context, execution *BroadcastExecution) error
    FindByBroadcast(ctx context.Context, broadcastID uuid.UUID) ([]*BroadcastExecution, error)
    FindPending(ctx context.Context, broadcastID uuid.UUID) ([]*BroadcastExecution, error)
}
```

---

## Commands (Planned)

### ContactList Commands

1. **CreateStaticListCommand** - Create list with fixed contacts
2. **CreateDynamicListCommand** - Create list with filters
3. **AddContactsToListCommand** - Add contacts to list
4. **RemoveContactsFromListCommand** - Remove contacts
5. **UpdateListFiltersCommand** - Update dynamic filters

### Broadcast Commands

1. **CreateBroadcastCommand** - Create new broadcast
2. **ScheduleBroadcastCommand** - Schedule for later
3. **ExecuteBroadcastCommand** - Execute immediately
4. **CancelBroadcastCommand** - Cancel scheduled/running broadcast
5. **UpdateBroadcastTemplateCommand** - Update message template (draft only)

---

## Use Cases (Planned)

### ContactList Use Cases

1. **CreateStaticListUseCase** - Create list with contact IDs
2. **CreateDynamicListUseCase** - Create list with filters
3. **ResolveContactsUseCase** - Get contacts from dynamic list
4. **AddContactsToListUseCase** - Add contacts
5. **RemoveContactsFromListUseCase** - Remove contacts

### Broadcast Use Cases

1. **CreateBroadcastUseCase** - Create new broadcast
2. **ScheduleBroadcastUseCase** - Schedule for future
3. **ExecuteBroadcastUseCase** - Execute immediately
4. **CancelBroadcastUseCase** - Cancel broadcast
5. **GetBroadcastStatsUseCase** - Retrieve statistics
6. **ListBroadcastExecutionsUseCase** - List individual sends

---

## Workers (Planned)

### 1. BroadcastSchedulerWorker

Monitors scheduled broadcasts and triggers execution at the right time:

```go
type BroadcastSchedulerWorker struct {
    broadcastRepo   BroadcastRepository
    executionWorker *BroadcastExecutionWorker
    pollInterval    time.Duration
}

func (w *BroadcastSchedulerWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.pollInterval)  // e.g., 1 minute

    for {
        select {
        case <-ticker.C:
            // Find ready broadcasts
            broadcasts := w.broadcastRepo.FindScheduledReady(ctx)

            for _, b := range broadcasts {
                b.Start()
                w.broadcastRepo.Save(ctx, b)

                // Execute in background
                go w.executionWorker.ExecuteBroadcast(ctx, b.ID())
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### 2. BroadcastExecutionWorker

Executes broadcast for all contacts with rate limiting:

```go
type BroadcastExecutionWorker struct {
    broadcastRepo   BroadcastRepository
    contactListRepo ContactListRepository
    messageSender   MessageSender
}

func (w *BroadcastExecutionWorker) ExecuteBroadcast(ctx context.Context, broadcastID uuid.UUID) error {
    broadcast, _ := w.broadcastRepo.FindByID(ctx, broadcastID)

    // 1. Resolve contacts from list
    list, _ := w.contactListRepo.FindByID(ctx, broadcast.ListID())
    contacts, _ := w.contactListRepo.ResolveContacts(ctx, list.ID())

    broadcast.UpdateTotalContacts(len(contacts))

    // 2. Create executions
    executions := make([]*BroadcastExecution, len(contacts))
    for i, contact := range contacts {
        executions[i] = NewBroadcastExecution(broadcast.ID(), contact.ID())
    }
    w.execRepo.SaveBatch(ctx, executions)

    // 3. Execute with rate limiting
    rateLimiter := NewRateLimiter(broadcast.RateLimit())

    for _, execution := range executions {
        rateLimiter.Wait()

        contact := contacts[execution.ContactID()]

        // Render message with variables
        message := w.renderMessage(broadcast.MessageTemplate(), contact)

        // Send
        messageID, err := w.messageSender.Send(ctx, contact, message)

        if err != nil {
            execution.MarkFailed(err.Error())
            broadcast.IncrementFailed()
        } else {
            execution.MarkSent(messageID)
            broadcast.IncrementSent()
        }

        w.execRepo.Save(ctx, execution)
        w.broadcastRepo.Save(ctx, broadcast)
    }

    // 4. Complete
    broadcast.Complete()
    w.broadcastRepo.Save(ctx, broadcast)

    return nil
}
```

---

## Rate Limiting (Planned)

```go
type RateLimiter struct {
    rate     int           // messages per minute
    interval time.Duration // interval between messages
}

func NewRateLimiter(msgsPerMinute int) *RateLimiter {
    if msgsPerMinute == 0 {
        return &RateLimiter{rate: 0} // unlimited
    }

    interval := time.Minute / time.Duration(msgsPerMinute)
    return &RateLimiter{
        rate:     msgsPerMinute,
        interval: interval,
    }
}

func (r *RateLimiter) Wait() {
    if r.rate == 0 {
        return  // no limit
    }
    time.Sleep(r.interval)
}
```

**Examples**:
- 60 msgs/min → waits 1s between each message
- 100 msgs/min → waits 600ms between each message
- 0 (unlimited) → sends as fast as possible

---

## Template Variables (Planned)

Available variables for substitution:

```
{{contact.name}}           → Contact name
{{contact.email}}          → Contact email
{{contact.phone}}          → Contact phone
{{contact.custom.X}}       → Custom field X
{{broadcast.name}}         → Broadcast name
{{current_date}}           → Current date
{{current_time}}           → Current time
{{unsubscribe_link}}       → Unsubscribe link
```

**Example Template**:
```
Hello {{contact.name}}!

Your last purchase was on {{contact.custom.last_purchase_date}}.
Take advantage of our exclusive promotion!

To unsubscribe: {{unsubscribe_link}}
```

---

## API Endpoints (Planned)

### Contact Lists

```
POST   /api/v1/contact-lists                    # Create list
GET    /api/v1/contact-lists                    # List lists
GET    /api/v1/contact-lists/:id                # Get list
PUT    /api/v1/contact-lists/:id                # Update list
DELETE /api/v1/contact-lists/:id                # Delete list
GET    /api/v1/contact-lists/:id/contacts       # Get list contacts
POST   /api/v1/contact-lists/:id/contacts       # Add contacts
DELETE /api/v1/contact-lists/:id/contacts/:cid  # Remove contact
```

### Broadcasts

```
POST   /api/v1/broadcasts                       # Create broadcast
GET    /api/v1/broadcasts                       # List broadcasts
GET    /api/v1/broadcasts/:id                   # Get broadcast
PUT    /api/v1/broadcasts/:id                   # Update (draft only)
DELETE /api/v1/broadcasts/:id                   # Delete (draft only)
POST   /api/v1/broadcasts/:id/schedule          # Schedule
POST   /api/v1/broadcasts/:id/execute           # Execute immediately
POST   /api/v1/broadcasts/:id/cancel            # Cancel
GET    /api/v1/broadcasts/:id/stats             # Statistics
GET    /api/v1/broadcasts/:id/executions        # Individual executions
```

---

## Use Case Examples (Planned)

### Example 1: Immediate Promotion Campaign

```bash
# 1. Create active customers list
curl -X POST /api/v1/contact-lists \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Active Customers",
    "is_dynamic": true,
    "filters": {
      "pipeline_id": "uuid",
      "status_id": "uuid",
      "tags": ["customer", "active"]
    }
  }'

# 2. Create immediate broadcast
curl -X POST /api/v1/broadcasts \
  -d '{
    "name": "Black Friday Promotion",
    "list_id": "list-uuid",
    "message_template": {
      "type": "template",
      "template_id": "promo_bf",
      "variables": {
        "name": "contact.name",
        "discount": "50"
      }
    },
    "rate_limit": 100
  }'

# 3. Execute immediately
curl -X POST /api/v1/broadcasts/{id}/execute
```

### Example 2: Scheduled Weekly Newsletter

```bash
# 1. Create newsletter subscribers list
curl -X POST /api/v1/contact-lists \
  -d '{
    "name": "Newsletter Subscribers",
    "is_dynamic": true,
    "filters": {
      "tags": ["newsletter"]
    }
  }'

# 2. Schedule for Monday 10am
curl -X POST /api/v1/broadcasts \
  -d '{
    "name": "Weekly Newsletter",
    "list_id": "list-uuid",
    "message_template": {
      "type": "template",
      "template_id": "newsletter_weekly"
    },
    "scheduled_for": "2025-01-20T10:00:00Z",
    "rate_limit": 200
  }'
```

### Example 3: Cart Abandonment Recovery

```bash
# 1. Create dynamic list of abandoned carts
curl -X POST /api/v1/contact-lists \
  -d '{
    "name": "Abandoned Carts Today",
    "is_dynamic": true,
    "filters": {
      "custom_fields": {
        "cart_status": "abandoned",
        "abandoned_date": "today"
      }
    }
  }'

# 2. Schedule for 2 hours after abandonment
curl -X POST /api/v1/broadcasts \
  -d '{
    "name": "Cart Recovery 2h",
    "list_id": "list-uuid",
    "message_template": {
      "type": "text",
      "content": "Hi {{name}}! Your cart is waiting. Complete now and get 10% OFF!"
    },
    "scheduled_for": "2025-01-15T14:00:00Z"
  }'
```

---

## Integration with Automation Rules (Planned)

Broadcasts can be triggered via **Automation Rules**:

```go
// Action: trigger_broadcast
{
  "type": "trigger_broadcast",
  "params": {
    "broadcast_id": "uuid",
    "override_schedule": false  // if true, ignores schedule and fires immediately
  }
}
```

**Example Rule**:
```json
{
  "name": "Send newsletter when status changes to Customer",
  "trigger": "status.changed",
  "conditions": [
    { "field": "new_status_name", "operator": "eq", "value": "Customer" }
  ],
  "actions": [
    {
      "type": "trigger_broadcast",
      "params": {
        "broadcast_id": "welcome_newsletter_uuid"
      }
    }
  ]
}
```

---

## Performance Considerations (Planned)

### Indexes

```sql
-- Contact Lists
CREATE INDEX idx_contact_lists_tenant ON contact_lists(tenant_id);
CREATE INDEX idx_contact_lists_dynamic ON contact_lists(is_dynamic);

-- List Contacts (junction table)
CREATE INDEX idx_list_contacts_list ON list_contacts(list_id);
CREATE INDEX idx_list_contacts_contact ON list_contacts(contact_id);

-- Broadcasts
CREATE INDEX idx_broadcasts_tenant ON broadcasts(tenant_id);
CREATE INDEX idx_broadcasts_status ON broadcasts(status);
CREATE INDEX idx_broadcasts_scheduled ON broadcasts(status, scheduled_for)
    WHERE status = 'scheduled';

-- Broadcast Executions
CREATE INDEX idx_executions_broadcast ON broadcast_executions(broadcast_id);
CREATE INDEX idx_executions_contact ON broadcast_executions(contact_id);
CREATE INDEX idx_executions_status ON broadcast_executions(status);
```

### Caching Strategy

```go
// Cache dynamic list results (5 min TTL)
cacheKey := fmt.Sprintf("contact_list:%s:contacts", listID)
contacts, err := cache.Get(cacheKey)

// Cache broadcast stats (1 min TTL)
statsKey := fmt.Sprintf("broadcast:%s:stats", broadcastID)
stats, err := cache.Get(statsKey)
```

---

## Implementation Roadmap

### Phase 1: Core Domain Models ❌
- [ ] ContactList aggregate
- [ ] Broadcast aggregate
- [ ] BroadcastExecution entity
- [ ] Domain events
- [ ] Repository interfaces

### Phase 2: Persistence ❌
- [ ] Database migrations
- [ ] GORM repositories
- [ ] Entity mappings

### Phase 3: Workers ❌
- [ ] BroadcastSchedulerWorker
- [ ] BroadcastExecutionWorker
- [ ] RateLimiter implementation

### Phase 4: API ❌
- [ ] ContactList handlers
- [ ] Broadcast handlers
- [ ] API documentation

### Phase 5: Integration ❌
- [ ] Automation rule integration
- [ ] Template variable rendering
- [ ] Message sender integration

### Phase 6: Testing ❌
- [ ] Unit tests
- [ ] Integration tests
- [ ] E2E tests

---

## References

- [Broadcast Design Document](../../internal/domain/broadcast/broadcast.md) (Portuguese)
- [Pipeline Automation](./pipeline_aggregate.md) - Related automation system
- [Message Aggregate](./message_aggregate.md) - Message sending

---

**Next**: [ContactEvent Aggregate](contact_event_aggregate.md) →
**Previous**: [ChannelType Aggregate](channel_type_aggregate.md) ←

---

## Summary

❌ **Broadcast Aggregate Status**: NOT IMPLEMENTED

**Planned Features**:
1. Contact list management (static and dynamic)
2. Mass message broadcasting
3. Scheduled and immediate broadcasts
4. Template variables and personalization
5. Rate limiting to prevent blocking
6. Detailed execution tracking
7. Integration with automation rules
8. Background workers for execution

The Broadcast aggregate is **fully designed** but awaiting implementation. It will enable marketing campaigns, newsletters, and automated customer engagement at scale.

**Next Steps**: Implement Phase 1 (Core Domain Models) when prioritized.
