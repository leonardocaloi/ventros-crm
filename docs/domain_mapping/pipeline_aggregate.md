# Pipeline Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~2,500 (with automation rules)
**Test Coverage**: Partial (7+ tests passing)

---

## Overview

- **Purpose**: Manages sales/support pipelines with automation
- **Location**: `internal/domain/pipeline/`
- **Entity**: `infrastructure/persistence/entities/pipeline.go`
- **Repository**: `infrastructure/persistence/gorm_pipeline_repository.go`
- **Aggregate Root**: `Pipeline`

**Business Problem**:
The Pipeline aggregate represents **sales funnels and support workflows** with automated actions at each stage. Pipelines organize contacts through sequential statuses (Lead → Qualified → Customer) and trigger automations based on status changes or events. Critical for:
- **Sales management** - Track deals through sales stages
- **Support workflows** - Route tickets through resolution process
- **Automation** - Trigger actions automatically (send message, assign agent, create task)
- **Analytics** - Conversion rates, stage durations, bottleneck identification

---

## Domain Model

### Aggregate Root: Pipeline

```go
type Pipeline struct {
    id                    uuid.UUID
    projectID             uuid.UUID
    tenantID              string
    name                  string
    description           string
    color                 string        // UI color
    position              int           // Display order
    active                bool
    sessionTimeoutMinutes *int          // Override default session timeout
    statuses              []*Status     // Pipeline stages
}
```

### Entity: Status

```go
type Status struct {
    id          uuid.UUID
    pipelineID  uuid.UUID
    name        string
    description string
    color       string
    statusType  StatusType // open, active, closed
    position    int        // Order in pipeline
    active      bool
}

type StatusType string
const (
    StatusTypeOpen   StatusType = "open"    // Initial stages (Lead, New)
    StatusTypeActive StatusType = "active"  // Active work (Qualified, In Progress)
    StatusTypeClosed StatusType = "closed"  // Final stages (Customer, Won, Lost)
)
```

### Entity: Automation (Automation Rules)

```go
type Automation struct {
    id             uuid.UUID
    automationType AutomationType      // pipeline_automation, scheduled, etc
    pipelineID     *uuid.UUID          // Optional: scoped to pipeline
    tenantID       string
    name           string
    description    string
    trigger        AutomationTrigger   // What starts the automation
    conditions     []RuleCondition     // When to execute (IF conditions)
    actions        []RuleAction        // What to do (THEN actions)
    priority       int                 // Execution order
    enabled        bool
}

// Triggers (20+ available)
type AutomationTrigger string
const (
    TriggerSessionEnded     AutomationTrigger = "session.ended"
    TriggerMessageReceived  AutomationTrigger = "message.received"
    TriggerStatusChanged    AutomationTrigger = "status.changed"
    TriggerAfterDelay       AutomationTrigger = "after.delay"
    TriggerPurchaseCompleted AutomationTrigger = "purchase.completed"
    TriggerCartAbandoned    AutomationTrigger = "cart.abandoned"
    // ... and 15+ more
)

// Conditions (8 operators: eq, ne, gt, gte, lt, lte, contains, in)
type RuleCondition struct {
    Field    string      // e.g., "message_count", "status", "hours_since_last_contact"
    Operator string      // e.g., "gt", "eq", "contains"
    Value    interface{} // e.g., 5, "Lead", "urgent"
}

// Actions (15+ available)
type AutomationAction string
const (
    ActionSendMessage       AutomationAction = "send_message"
    ActionChangeStatus      AutomationAction = "change_pipeline_status"
    ActionAssignAgent       AutomationAction = "assign_agent"
    ActionCreateTask        AutomationAction = "create_task"
    ActionAddTag            AutomationAction = "add_tag"
    ActionSendWebhook       AutomationAction = "send_webhook"
    ActionTriggerWorkflow   AutomationAction = "trigger_workflow"
    ActionCreateNote        AutomationAction = "create_note"
    ActionNotifyAgent       AutomationAction = "notify_agent"
    // ... and 10+ more
)

type RuleAction struct {
    Type   AutomationAction
    Params map[string]interface{}
    Delay  int  // Optional delay in minutes
}
```

### Business Invariants

1. **Pipeline must belong to Project**
   - `projectID` and `tenantID` required
   - `name` required

2. **Status belongs to Pipeline**
   - Each status has unique name within pipeline
   - Status types define behavior (open/active/closed)

3. **Automation execution**
   - Conditions evaluated with AND logic (all must match)
   - Actions executed sequentially with optional delays
   - Priority determines execution order when multiple match

4. **Status transitions**
   - Contacts move forward through statuses
   - History tracked in `contact_status_history` table

---

## Events Emitted

The Pipeline aggregate emits **15+ domain events**:

| Event | When | Purpose |
|-------|------|---------|
| `pipeline.created` | New pipeline | Initialize pipeline |
| `pipeline.updated` | Pipeline modified | Track changes |
| `pipeline.activated` | Pipeline enabled | Enable automation |
| `pipeline.deactivated` | Pipeline disabled | Disable automation |
| `status.created` | New status added | Create stage |
| `status.updated` | Status modified | Track changes |
| `pipeline.status_added` | Status added to pipeline | Update UI |
| `pipeline.status_removed` | Status removed | Update UI |
| `contact.status_changed` | Contact moved between statuses | Trigger automation |
| `contact.entered_pipeline` | Contact enters pipeline | Trigger onboarding |
| `contact.exited_pipeline` | Contact exits pipeline | Trigger closure |
| `automation.created` | New automation rule | Register rule |
| `automation.enabled` | Automation activated | Start monitoring |
| `automation.disabled` | Automation deactivated | Stop monitoring |
| `automation_rule.triggered` | Rule conditions matched | Log trigger |
| `automation_rule.executed` | Actions completed | Log execution |
| `automation_rule.failed` | Execution failed | Alert admin |

---

## Repository Interface

```go
type Repository interface {
    // Pipelines
    SavePipeline(ctx context.Context, pipeline *Pipeline) error
    FindPipelineByID(ctx context.Context, id uuid.UUID) (*Pipeline, error)
    FindPipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*Pipeline, error)
    FindActivePipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*Pipeline, error)
    DeletePipeline(ctx context.Context, id uuid.UUID) error

    // Statuses
    SaveStatus(ctx context.Context, status *Status) error
    FindStatusByID(ctx context.Context, id uuid.UUID) (*Status, error)
    FindStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*Status, error)
    DeleteStatus(ctx context.Context, id uuid.UUID) error

    // Contact Status Management
    SetContactStatus(ctx context.Context, contactID, pipelineID, statusID uuid.UUID) error
    GetContactStatus(ctx context.Context, contactID, pipelineID uuid.UUID) (*Status, error)
    GetContactsByStatus(ctx context.Context, pipelineID, statusID uuid.UUID) ([]uuid.UUID, error)
    GetContactStatusHistory(ctx context.Context, contactID, pipelineID uuid.UUID) ([]*ContactStatusHistory, error)

    // Advanced queries
    FindByTenantWithFilters(ctx context.Context, filters PipelineFilters) ([]*Pipeline, int64, error)
    SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Pipeline, int64, error)
}
```

---

## Commands (CQRS)

### ✅ Implemented

1. **CreatePipelineCommand**
2. **UpdatePipelineCommand**
3. **AddStatusCommand**
4. **RemoveStatusCommand**
5. **CreateAutomationCommand**
6. **ChangePipelineStatusCommand** (`internal/application/contact/change_pipeline_status_usecase.go`)

### ❌ Suggested

- **BulkMoveContactsCommand** - Move multiple contacts at once
- **ClonePipelineCommand** - Duplicate pipeline with statuses
- **ArchivePipelineCommand** - Archive with historical data
- **TestAutomationCommand** - Dry-run automation rules

---

## Automation System

### Example: Welcome Message on New Lead

```go
automation := &Automation{
    name: "Welcome Message for New Leads",
    trigger: TriggerStatusChanged,
    conditions: []RuleCondition{
        {Field: "new_status_name", Operator: "eq", Value: "Lead"},
    },
    actions: []RuleAction{
        {
            Type: ActionSendMessage,
            Params: map[string]interface{}{
                "content": "Welcome! How can I help you?",
            },
            Delay: 0,
        },
        {
            Type: ActionAddTag,
            Params: map[string]interface{}{
                "tag": "new_lead",
            },
            Delay: 0,
        },
    },
}
```

### Example: Follow-up after 24h without response

```go
automation := &Automation{
    name: "Follow-up after 24h inactivity",
    trigger: TriggerAfterDelay,
    conditions: []RuleCondition{
        {Field: "hours_since_last_message", Operator: "gte", Value: 24},
        {Field: "last_message_from", Operator: "eq", Value: "agent"},
    },
    actions: []RuleAction{
        {
            Type: ActionSendMessage,
            Params: map[string]interface{}{
                "content": "Still interested? Let me know if you need help!",
            },
            Delay: 0,
        },
    },
}
```

### Available Actions Summary

| Category | Actions |
|----------|---------|
| **Messaging** | send_message, send_template, send_email |
| **Pipeline** | change_pipeline_status |
| **Assignment** | assign_agent, assign_to_queue |
| **Tasks** | create_task |
| **Organization** | add_tag, remove_tag, update_custom_field |
| **Notes** | create_note, create_agent_report |
| **Integration** | send_webhook, trigger_workflow |
| **Notifications** | notify_agent, notify_coordinator |

---

## Use Cases

### ✅ Implemented

1. **ChangePipelineStatusUseCase** - Move contact between statuses
2. **CreateNoteExecutor** - Create notes from automation

### ❌ Suggested

3. **ExecuteAutomationUseCase** - Run automation actions
4. **EvaluateAutomationsUseCase** - Check which automations should trigger
5. **AnalyzePipelinePerformanceUseCase** - Calculate conversion rates
6. **ForecastRevenueUseCase** - Predict revenue from pipeline
7. **IdentifyBottlenecksUseCase** - Find stuck contacts

---

## Performance Considerations

### Indexes

```sql
-- Pipelines
CREATE INDEX idx_pipelines_project ON pipelines(project_id);
CREATE INDEX idx_pipelines_active ON pipelines(active) WHERE active = true;

-- Statuses
CREATE INDEX idx_statuses_pipeline ON statuses(pipeline_id);
CREATE INDEX idx_statuses_type ON statuses(status_type);

-- Contact Status (CRITICAL for queries)
CREATE INDEX idx_contact_status_contact ON contact_pipeline_status(contact_id, pipeline_id);
CREATE INDEX idx_contact_status_status ON contact_pipeline_status(pipeline_id, status_id);

-- Automation execution
CREATE INDEX idx_automation_pipeline ON automation_rules(pipeline_id, enabled)
    WHERE enabled = true;
```

---

## Real-World Examples

### Sales Pipeline

```
Pipeline: "Sales Funnel"
Statuses:
1. Lead (open)          → Automation: Send welcome message
2. Qualified (active)   → Automation: Assign to sales rep
3. Proposal (active)    → Automation: Send proposal template
4. Negotiation (active) → Automation: Notify manager if >$10k
5. Won (closed)         → Automation: Send thank you + onboarding
6. Lost (closed)        → Automation: Add to re-engagement campaign
```

### Support Pipeline

```
Pipeline: "Support Tickets"
Statuses:
1. New (open)           → Automation: Auto-assign to available agent
2. In Progress (active) → Automation: Set SLA timer
3. Waiting (active)     → Automation: Follow-up after 2 hours
4. Resolved (closed)    → Automation: Send satisfaction survey
5. Closed (closed)      → Automation: Archive after 30 days
```

---

## API Examples

### Create Pipeline

```http
POST /api/v1/pipelines
{
  "name": "Sales Funnel",
  "description": "Main sales pipeline",
  "color": "#4CAF50"
}
```

### Add Status

```http
POST /api/v1/pipelines/{id}/statuses
{
  "name": "Qualified Lead",
  "status_type": "active",
  "color": "#2196F3"
}
```

### Create Automation

```http
POST /api/v1/automations
{
  "pipeline_id": "uuid",
  "name": "Welcome Message",
  "trigger": "status.changed",
  "conditions": [
    {"field": "new_status_name", "operator": "eq", "value": "Lead"}
  ],
  "actions": [
    {"type": "send_message", "params": {"content": "Welcome!"}}
  ]
}
```

### Move Contact to Status

```http
POST /api/v1/contacts/{contact_id}/change-status
{
  "pipeline_id": "uuid",
  "new_status_id": "uuid",
  "notes": "Qualified after demo call"
}
```

---

## References

- [Pipeline Domain](../../internal/domain/pipeline/)
- [Automation Rule](../../internal/domain/pipeline/automation_rule.go)
- [Status](../../internal/domain/pipeline/status.go)
- [Pipeline Repository](../../infrastructure/persistence/gorm_pipeline_repository.go)
- [Change Pipeline Status Use Case](../../internal/application/contact/change_pipeline_status_usecase.go)

---

**Next**: [Agent Aggregate](agent_aggregate.md) →
**Previous**: [Message Aggregate](message_aggregate.md) ←
