# Campaign Aggregate

## Overview

The **Campaign** aggregate is the heart of Ventros CRM's marketing automation engine. It orchestrates complex, multi-step marketing campaigns that can include broadcasts, message sequences, conditional logic, delays, and wait steps - similar to HubSpot Workflows, ActiveCampaign Automation, or Mailchimp Customer Journeys.

A Campaign is a structured series of automated actions executed in sequence against enrolled contacts, with support for conditional branching, goal tracking, and comprehensive analytics.

- **Purpose**: Multi-step marketing automation with conditional logic and goal tracking
- **Location**: `internal/domain/automation/campaign/`
- **Entity**: `infrastructure/persistence/entities/campaign.go`
- **Type**: Core automation aggregate (CRITICAL for marketing workflows)

---

## Domain Model

### Aggregate Root: Campaign

```go
type Campaign struct {
    id          uuid.UUID
    version     int // Optimistic locking
    tenantID    string
    name        string
    description string
    status      CampaignStatus
    steps       []CampaignStep

    // Goals and metrics
    goalType         GoalType
    goalValue        int
    contactsReached  int
    conversionsCount int

    // Scheduling
    startDate *time.Time
    endDate   *time.Time

    // Timestamps
    createdAt time.Time
    updatedAt time.Time

    events []shared.DomainEvent
}
```

### Value Objects & Entities

#### 1. **CampaignStatus** (Value Object)
```go
type CampaignStatus string

const (
    CampaignStatusDraft     CampaignStatus = "draft"     // Being configured
    CampaignStatusScheduled CampaignStatus = "scheduled" // Scheduled to start
    CampaignStatusActive    CampaignStatus = "active"    // Currently running
    CampaignStatusPaused    CampaignStatus = "paused"    // Temporarily stopped
    CampaignStatusCompleted CampaignStatus = "completed" // Finished
    CampaignStatusArchived  CampaignStatus = "archived"  // Archived
)
```

#### 2. **GoalType** (Value Object)
```go
type GoalType string

const (
    GoalTypeReachContacts GoalType = "reach_contacts" // Reach X contacts
    GoalTypeConversions   GoalType = "conversions"   // Get X conversions
    GoalTypeEngagement    GoalType = "engagement"    // Get X engagements
)
```

#### 3. **CampaignStep** (Entity)
```go
type CampaignStep struct {
    ID         uuid.UUID
    Order      int           // Execution order (0, 1, 2, ...)
    Name       string
    Type       StepType      // broadcast, sequence, delay, condition, wait
    Config     StepConfig    // Type-specific configuration
    Conditions []StepCondition // Execution conditions
    CreatedAt  time.Time
}
```

**Step Types**:
- `broadcast` - Send broadcast message to all enrolled contacts
- `sequence` - Start a message sequence for contact
- `delay` - Wait for specific time (minutes/hours/days)
- `condition` - Conditional branching based on contact data
- `wait` - Wait for user action (reply, click, open)

**Step Configuration**:
```go
type StepConfig struct {
    // For broadcast steps
    BroadcastID *uuid.UUID `json:"broadcast_id,omitempty"`

    // For sequence steps
    SequenceID *uuid.UUID `json:"sequence_id,omitempty"`

    // For delay/wait steps
    DelayAmount *int    `json:"delay_amount,omitempty"` // 5, 24, 7
    DelayUnit   *string `json:"delay_unit,omitempty"`   // minutes, hours, days

    // For condition steps
    ConditionType *string                `json:"condition_type,omitempty"` // tag_has, field_equals
    ConditionData map[string]interface{} `json:"condition_data,omitempty"`

    // For wait steps (wait for user action)
    WaitFor     *string `json:"wait_for,omitempty"`     // reply, click, open
    WaitTimeout *int    `json:"wait_timeout,omitempty"` // timeout in hours
    TimeoutStep *int    `json:"timeout_step,omitempty"` // step to jump to on timeout
}
```

**Step Conditions**:
```go
type StepCondition struct {
    Type     string      `json:"type"`     // tag_has, field_equals, pipeline_status
    Field    string      `json:"field"`    // field to check
    Operator string      `json:"operator"` // equals, contains, greater_than
    Value    interface{} `json:"value"`    // value to compare against
    Metadata map[string]interface{} `json:"metadata"`
}
```

#### 4. **CampaignEnrollment** (Separate Aggregate)
```go
type CampaignEnrollment struct {
    id               uuid.UUID
    campaignID       uuid.UUID
    contactID        uuid.UUID
    status           EnrollmentStatus // active, paused, completed, exited
    currentStepOrder int              // Current step being executed
    nextScheduledAt  *time.Time       // When to execute next step
    exitedAt         *time.Time
    exitReason       *string
    completedAt      *time.Time
    enrolledAt       time.Time
    updatedAt        time.Time
}
```

**Enrollment Statuses**:
- `active` - Currently executing steps
- `paused` - Temporarily stopped
- `completed` - Finished all steps
- `exited` - Exited early (unsubscribe, error, etc.)

---

## Business Invariants

### Campaign Invariants

1. **Name Required**: Campaign must have a non-empty name
2. **Tenant Isolation**: Campaign belongs to exactly one tenant
3. **Status Transitions**: Only valid state transitions allowed:
   - draft → scheduled, active
   - scheduled → active
   - active → paused, completed
   - paused → active
   - any → archived
4. **Steps Required for Activation**: Cannot activate/schedule campaign with no steps
5. **Future Scheduling**: Schedule date must be in the future
6. **Step Order Uniqueness**: Each step must have unique order number
7. **Draft-Only Modifications**: Steps can only be added/removed in draft status
8. **Goal Value Non-Negative**: Goal value must be >= 0
9. **Optimistic Locking**: Version field prevents lost updates

### Enrollment Invariants

1. **Valid IDs**: campaignID and contactID cannot be nil
2. **Active Advancement**: Only active enrollments can advance to next step
3. **No Double Completion**: Cannot complete already completed enrollment
4. **Status Consistency**: Exit/complete sets nextScheduledAt to nil

---

## Events Emitted

### Campaign Lifecycle Events

1. **`campaign.created`**
   ```go
   CampaignCreatedEvent {
       CampaignID  uuid.UUID
       TenantID    string
       Name        string
       Description string
       GoalType    GoalType
       GoalValue   int
   }
   ```
   **When**: Campaign is created in draft status
   **Handlers**: Log creation, update analytics dashboard

2. **`campaign.activated`**
   ```go
   CampaignActivatedEvent {
       CampaignID uuid.UUID
   }
   ```
   **When**: Campaign transitions from draft/scheduled to active
   **Handlers**: Start enrollment worker, notify admins

3. **`campaign.scheduled`**
   ```go
   CampaignScheduledEvent {
       CampaignID uuid.UUID
       StartDate  time.Time
   }
   ```
   **When**: Campaign is scheduled to start at specific time
   **Handlers**: Schedule activation job

4. **`campaign.paused`**
   ```go
   CampaignPausedEvent {
       CampaignID uuid.UUID
   }
   ```
   **When**: Active campaign is paused
   **Handlers**: Pause enrollment worker, pause all enrollments

5. **`campaign.resumed`**
   ```go
   CampaignResumedEvent {
       CampaignID uuid.UUID
   }
   ```
   **When**: Paused campaign is resumed
   **Handlers**: Resume enrollment worker, resume enrollments

6. **`campaign.completed`**
   ```go
   CampaignCompletedEvent {
       CampaignID uuid.UUID
   }
   ```
   **When**: Campaign reaches completion (all goals met or ended)
   **Handlers**: Generate final report, notify admins, archive data

7. **`campaign.archived`**
   ```go
   CampaignArchivedEvent {
       CampaignID uuid.UUID
   }
   ```
   **When**: Campaign is archived (soft delete)
   **Handlers**: Remove from active lists, update analytics

### Step Management Events

8. **`campaign.step_added`**
   ```go
   CampaignStepAddedEvent {
       CampaignID uuid.UUID
       StepID     uuid.UUID
       StepType   StepType
       Order      int
   }
   ```
   **When**: New step is added to campaign (draft only)

9. **`campaign.step_removed`**
   ```go
   CampaignStepRemovedEvent {
       CampaignID uuid.UUID
       StepID     uuid.UUID
   }
   ```
   **When**: Step is removed from campaign (draft only)

### Enrollment Events

10. **`campaign.contact_enrolled`**
    ```go
    ContactEnrolledEvent {
        EnrollmentID    uuid.UUID
        CampaignID      uuid.UUID
        ContactID       uuid.UUID
        NextScheduledAt time.Time
    }
    ```
    **When**: Contact is enrolled in campaign
    **Handlers**: Schedule first step execution, update contact timeline

11. **`campaign.enrollment_advanced`**
    ```go
    EnrollmentAdvancedEvent {
        EnrollmentID     uuid.UUID
        CampaignID       uuid.UUID
        ContactID        uuid.UUID
        CurrentStepOrder int
        NextScheduledAt  *time.Time
    }
    ```
    **When**: Enrollment advances to next step
    **Handlers**: Schedule next step, update progress tracking

12. **`campaign.enrollment_paused`**
    ```go
    EnrollmentPausedEvent {
        EnrollmentID uuid.UUID
        CampaignID   uuid.UUID
        ContactID    uuid.UUID
    }
    ```
    **When**: Enrollment is paused

13. **`campaign.enrollment_resumed`**
    ```go
    EnrollmentResumedEvent {
        EnrollmentID uuid.UUID
        CampaignID   uuid.UUID
        ContactID    uuid.UUID
    }
    ```
    **When**: Paused enrollment is resumed

14. **`campaign.enrollment_completed`**
    ```go
    EnrollmentCompletedEvent {
        EnrollmentID uuid.UUID
        CampaignID   uuid.UUID
        ContactID    uuid.UUID
        CompletedAt  time.Time
    }
    ```
    **When**: Enrollment completes all steps successfully
    **Handlers**: Update campaign metrics, contact timeline

15. **`campaign.enrollment_exited`**
    ```go
    EnrollmentExitedEvent {
        EnrollmentID uuid.UUID
        CampaignID   uuid.UUID
        ContactID    uuid.UUID
        ExitReason   string
        ExitedAt     time.Time
    }
    ```
    **When**: Enrollment exits early (unsubscribe, error, contact deleted)
    **Handlers**: Log exit reason, update metrics

**Total Events**: 15 (9 campaign lifecycle + 2 step management + 6 enrollment)

---

## Repository Interface

### CampaignRepository
```go
type Repository interface {
    Save(campaign *Campaign) error
    FindByID(id uuid.UUID) (*Campaign, error)
    FindByTenantID(tenantID string) ([]*Campaign, error)
    FindActiveByStatus(status CampaignStatus) ([]*Campaign, error)
    FindScheduled() ([]*Campaign, error) // Campaigns ready to start
    Delete(id uuid.UUID) error
}
```

### EnrollmentRepository
```go
type EnrollmentRepository interface {
    Save(enrollment *CampaignEnrollment) error
    FindByID(id uuid.UUID) (*CampaignEnrollment, error)
    FindByCampaignID(campaignID uuid.UUID) ([]*CampaignEnrollment, error)
    FindByContactID(contactID uuid.UUID) ([]*CampaignEnrollment, error)
    FindReadyForNextStep() ([]*CampaignEnrollment, error) // Ready to execute next step
    FindActiveByCampaignAndContact(campaignID, contactID uuid.UUID) (*CampaignEnrollment, error)
    Delete(id uuid.UUID) error
}
```

**Implementation**: `infrastructure/persistence/gorm_campaign_repository.go`
- Uses GORM with PostgreSQL
- Optimistic locking via version field
- Transaction support for step management
- Cascade deletes for steps

---

## Commands (NOT Implemented)

**TODO**: Create command layer for campaigns:

```go
// ❌ NOT IMPLEMENTED - Suggested commands

type CreateCampaignCommand struct {
    TenantID    string
    Name        string
    Description string
    GoalType    string
    GoalValue   int
}

type AddCampaignStepCommand struct {
    CampaignID uuid.UUID
    Order      int
    Name       string
    StepType   string
    Config     map[string]interface{}
}

type ActivateCampaignCommand struct {
    CampaignID uuid.UUID
}

type EnrollContactCommand struct {
    CampaignID uuid.UUID
    ContactID  uuid.UUID
}
```

---

## Queries (NOT Implemented)

**TODO**: Create query layer for campaigns:

```go
// ❌ NOT IMPLEMENTED - Suggested queries

type ListCampaignsQuery struct {
    TenantID string
    Status   *CampaignStatus
    Page     int
    Limit    int
}

type GetCampaignStatsQuery struct {
    CampaignID uuid.UUID
}

type ListCampaignEnrollmentsQuery struct {
    CampaignID uuid.UUID
    Status     *EnrollmentStatus
    Page       int
    Limit      int
}
```

---

## Use Cases

### ✅ Implemented (Partial)

**Location**: `internal/application/automation/` (TODO: verify if exists)

**Current Status**: Domain layer is complete, but application layer use cases are likely MISSING or incomplete.

### ❌ Suggested Use Cases

```go
// internal/application/automation/campaign/

// 1. CreateCampaignUseCase
//    - Validates input
//    - Creates Campaign aggregate
//    - Saves to repository
//    - Publishes campaign.created event
//    - Returns CampaignID

// 2. AddStepToCampaignUseCase
//    - Loads campaign
//    - Validates step configuration
//    - Adds step via Campaign.AddStep()
//    - Saves campaign
//    - Publishes campaign.step_added event

// 3. ActivateCampaignUseCase
//    - Loads campaign
//    - Validates campaign has steps
//    - Calls Campaign.Activate()
//    - Saves campaign
//    - Publishes campaign.activated event
//    - Starts enrollment worker

// 4. EnrollContactInCampaignUseCase
//    - Loads campaign (must be active)
//    - Loads contact
//    - Checks if already enrolled
//    - Creates CampaignEnrollment
//    - Calculates first step delay
//    - Saves enrollment
//    - Publishes campaign.contact_enrolled event
//    - Schedules first step execution

// 5. ExecuteCampaignStepUseCase (CRITICAL)
//    - Loads enrollment
//    - Loads campaign
//    - Gets current step by order
//    - Validates step conditions
//    - Executes step based on type:
//      * broadcast → Send broadcast to contact
//      * sequence → Enroll contact in sequence
//      * delay → Schedule next step with delay
//      * condition → Evaluate condition, branch accordingly
//      * wait → Wait for user action (reply/click/open)
//    - Advances enrollment to next step
//    - Saves enrollment
//    - Publishes campaign.enrollment_advanced event

// 6. PauseCampaignUseCase
//    - Loads campaign
//    - Calls Campaign.Pause()
//    - Pauses all active enrollments
//    - Saves campaign
//    - Publishes campaign.paused event

// 7. CompleteCampaignUseCase
//    - Loads campaign
//    - Calls Campaign.Complete()
//    - Completes all active enrollments
//    - Generates final report
//    - Saves campaign
//    - Publishes campaign.completed event

// 8. GetCampaignStatsUseCase
//    - Loads campaign
//    - Calls Campaign.GetStats()
//    - Returns CampaignStats (contacts reached, conversions, rates)
```

---

## Temporal Workflows (TODO)

**CRITICAL**: Campaign execution should be orchestrated by Temporal workflows for reliability and durability.

### Suggested Workflows:

1. **CampaignEnrollmentWorkflow**
   ```go
   // Orchestrates a single contact's journey through a campaign
   //
   // Activities:
   // - LoadCampaignActivity() - Load campaign definition
   // - ValidateEnrollmentActivity() - Check if contact eligible
   // - ExecuteStepActivity() - Execute current step
   // - EvaluateConditionsActivity() - Evaluate step conditions
   // - ScheduleDelayActivity() - Schedule delay/wait steps
   // - AdvanceEnrollmentActivity() - Move to next step
   // - CompleteEnrollmentActivity() - Mark as completed
   //
   // Handles:
   // - Step-by-step execution
   // - Conditional branching
   // - Delays and wait steps
   // - Error handling and retries
   // - Early exit (unsubscribe)
   ```

2. **CampaignActivationWorkflow**
   ```go
   // Activates a campaign and enrolls eligible contacts
   //
   // Activities:
   // - ActivateCampaignActivity() - Activate campaign
   // - FindEligibleContactsActivity() - Query contacts matching criteria
   // - EnrollContactsActivity() - Batch enroll contacts
   // - StartEnrollmentWorkflowsActivity() - Start individual workflows
   ```

3. **CampaignSchedulerWorkflow**
   ```go
   // Background worker that processes scheduled campaigns
   //
   // Activities:
   // - FindScheduledCampaignsActivity() - Query campaigns ready to start
   // - ActivateCampaignActivity() - Activate each campaign
   // - StartCampaignActivationWorkflowActivity() - Start activation workflow
   //
   // Runs: Every 1 minute (cron schedule)
   ```

**Implementation Status**: ❌ NOT IMPLEMENTED

---

## HTTP API (TODO)

### Suggested Endpoints:

```yaml
POST   /api/v1/campaigns                    # Create campaign
GET    /api/v1/campaigns                    # List campaigns
GET    /api/v1/campaigns/:id                # Get campaign details
PUT    /api/v1/campaigns/:id                # Update campaign (draft only)
DELETE /api/v1/campaigns/:id                # Delete campaign
POST   /api/v1/campaigns/:id/activate       # Activate campaign
POST   /api/v1/campaigns/:id/schedule       # Schedule campaign
POST   /api/v1/campaigns/:id/pause          # Pause campaign
POST   /api/v1/campaigns/:id/resume         # Resume campaign
GET    /api/v1/campaigns/:id/stats          # Get campaign statistics

POST   /api/v1/campaigns/:id/steps          # Add step to campaign
DELETE /api/v1/campaigns/:id/steps/:stepId  # Remove step
PUT    /api/v1/campaigns/:id/steps/:stepId  # Update step

POST   /api/v1/campaigns/:id/enroll         # Enroll contacts
GET    /api/v1/campaigns/:id/enrollments    # List enrollments
GET    /api/v1/enrollments/:id              # Get enrollment details
POST   /api/v1/enrollments/:id/pause        # Pause enrollment
POST   /api/v1/enrollments/:id/resume       # Resume enrollment
POST   /api/v1/enrollments/:id/exit         # Exit enrollment
```

**Implementation Status**: ❌ NOT IMPLEMENTED

---

## Real-World Usage Patterns

### Pattern 1: Welcome Series Campaign

```go
// Step 1: Delay - Wait 5 minutes after signup
step1 := NewCampaignStep(0, "Initial Delay", StepTypeDelay, StepConfig{
    DelayAmount: &five,
    DelayUnit:   &minutes,
})

// Step 2: Broadcast - Send welcome message
step2 := NewCampaignStep(1, "Welcome Message", StepTypeBroadcast, StepConfig{
    BroadcastID: &welcomeBroadcastID,
})

// Step 3: Wait - Wait for reply (timeout 24 hours)
step3 := NewCampaignStep(2, "Wait for Reply", StepTypeWait, StepConfig{
    WaitFor:     &reply,
    WaitTimeout: &twentyFour,
    TimeoutStep: &four, // Jump to step 4 if no reply
})

// Step 4: Condition - Check if replied
step4 := NewCampaignStep(3, "Check Reply", StepTypeCondition, StepConfig{
    ConditionType: &fieldEquals,
    ConditionData: map[string]interface{}{
        "field": "last_reply_at",
        "operator": "is_not_null",
    },
})

// Step 5: Sequence - Start onboarding sequence
step5 := NewCampaignStep(4, "Onboarding Sequence", StepTypeSequence, StepConfig{
    SequenceID: &onboardingSequenceID,
})

campaign.AddStep(step1)
campaign.AddStep(step2)
campaign.AddStep(step3)
campaign.AddStep(step4)
campaign.AddStep(step5)
```

### Pattern 2: Re-engagement Campaign

```go
// Target: Contacts inactive for 30+ days

// Step 1: Broadcast - "We miss you!" message
// Step 2: Delay - Wait 3 days
// Step 3: Condition - Check if opened message
// Step 4a: If opened - Send special offer
// Step 4b: If not opened - Send different message
// Step 5: Delay - Wait 7 days
// Step 6: Final broadcast or exit campaign
```

### Pattern 3: Product Launch Campaign

```go
// Target: All active contacts with "product_interest" tag

// Step 1: Broadcast - Pre-launch teaser
// Step 2: Delay - Wait until launch date
// Step 3: Broadcast - Launch announcement
// Step 4: Wait - Wait for link click (24 hours)
// Step 5: Condition - Check if clicked
// Step 6a: If clicked - Mark conversion
// Step 6b: If not clicked - Send reminder
```

---

## Performance Considerations

### Scalability
- **Enrollments**: Can scale to millions of concurrent enrollments
- **Step Execution**: Use Temporal workflows for distributed execution
- **Caching**: Cache campaign definitions (steps rarely change once active)
- **Batching**: Process enrollments in batches for efficiency

### Database Indexes
```sql
-- Already exist in migration:
CREATE INDEX idx_campaigns_tenant ON campaigns(tenant_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaign_enrollments_campaign_id ON campaign_enrollments(campaign_id);
CREATE INDEX idx_campaign_enrollments_contact_id ON campaign_enrollments(contact_id);
CREATE INDEX idx_campaign_enrollments_status ON campaign_enrollments(status);
CREATE INDEX idx_campaign_enrollments_next_scheduled ON campaign_enrollments(next_scheduled_at)
    WHERE status = 'active';
```

### Optimizations
1. **Bulk Enrollment**: Enroll multiple contacts in single transaction
2. **Step Caching**: Cache step definitions in memory
3. **Condition Evaluation**: Pre-compile condition logic
4. **Metrics Aggregation**: Use materialized views for campaign statistics
5. **Archival**: Archive completed enrollments after 90 days

---

## Testing Checklist

### Unit Tests (Domain Layer)
- [x] Campaign creation with valid data
- [x] Campaign status transitions (draft → active, pause/resume, etc.)
- [x] Step management (add, remove, update, validate order)
- [x] Goal tracking (increment contacts/conversions)
- [x] Statistics calculation (conversion rate, progress rate)
- [x] Enrollment creation and state transitions
- [x] Enrollment advancement (next step, completion, exit)
- [x] Step validation (config requirements)
- [x] Delay duration calculation (minutes/hours/days)
- [x] Domain events emission (14 events)

### Integration Tests (Repository Layer)
- [ ] Save and load campaign with steps
- [ ] Optimistic locking (version conflicts)
- [ ] Find campaigns by tenant
- [ ] Find campaigns by status
- [ ] Find scheduled campaigns
- [ ] Cascade delete (campaign → steps)
- [ ] Enrollment queries (by campaign, by contact, ready for next step)
- [ ] Find active enrollment for campaign+contact

### End-to-End Tests (Application Layer)
- [ ] Create campaign → Add steps → Activate
- [ ] Enroll contacts → Execute steps → Complete
- [ ] Pause campaign → Resume campaign
- [ ] Conditional branching logic
- [ ] Wait for user action (reply/click)
- [ ] Campaign statistics accuracy
- [ ] Temporal workflow execution

---

## Suggested Improvements

### 1. A/B Testing Support
```go
type ABTestStep struct {
    VariantA StepConfig
    VariantB StepConfig
    SplitPercentage float64 // 50/50, 70/30, etc.
}
```

### 2. Advanced Goal Tracking
```go
type Goal struct {
    Type      GoalType
    Value     int
    Operator  string // >=, >, ==
    Metric    string // contacts_reached, reply_rate, conversion_rate
    AutoStop  bool   // Auto-stop campaign when goal reached
}
```

### 3. Step Templates
```go
// Pre-defined step templates for common patterns
type StepTemplate struct {
    Name        string
    Description string
    Type        StepType
    DefaultConfig StepConfig
}

// Examples: "Welcome Message", "Re-engagement", "Product Recommendation"
```

### 4. Campaign Cloning
```go
func (c *Campaign) Clone(newName string) *Campaign {
    // Clone campaign with all steps
    // Useful for A/B testing or variations
}
```

### 5. Smart Scheduling
```go
// Send messages during contact's active hours
type SmartScheduling struct {
    RespectTimezone bool
    PreferredHours  []int // [9, 10, 11, ..., 17]
    PreferredDays   []string // ["Monday", "Tuesday", ...]
}
```

### 6. Exit Rules
```go
type ExitRule struct {
    Type      string // unsubscribe, goal_reached, tag_removed
    Automatic bool   // Auto-exit when condition met
}
```

---

## Related Aggregates

- **Broadcast**: Referenced by StepTypeBroadcast steps
- **Sequence**: Referenced by StepTypeSequence steps
- **Contact**: Enrolled in campaigns via CampaignEnrollment
- **Message**: Sent as part of broadcast/sequence steps
- **Tag**: Used in conditional logic and enrollment criteria

---

## Industry Comparison

| Feature | Ventros CRM | HubSpot | ActiveCampaign | Mailchimp |
|---------|-------------|---------|----------------|-----------|
| Multi-step workflows | ✅ | ✅ | ✅ | ✅ |
| Conditional branching | ✅ | ✅ | ✅ | ✅ |
| A/B testing | ❌ | ✅ | ✅ | ✅ |
| Goal tracking | ✅ | ✅ | ✅ | ✅ |
| Wait for action | ✅ | ✅ | ✅ | ✅ |
| Time delays | ✅ | ✅ | ✅ | ✅ |
| Enrollment limits | ❌ | ✅ | ✅ | ✅ |
| Visual builder | ❌ | ✅ | ✅ | ✅ |
| Re-enrollment rules | ❌ | ✅ | ✅ | ✅ |

**Ventros Strengths**:
- Clean DDD architecture
- Temporal-based reliability
- Full event-driven system
- Multi-channel support (not just email)

**Suggested Additions**:
- A/B testing
- Visual workflow builder (frontend)
- Re-enrollment rules
- Enrollment limits (daily/total)

---

## Documentation References

- **Domain Events**: `internal/domain/automation/campaign/events.go`
- **Repository**: `internal/domain/automation/campaign/repository.go`
- **Implementation**: `infrastructure/persistence/gorm_campaign_repository.go`
- **Entities**: `infrastructure/persistence/entities/campaign.go`
- **Migration**: `infrastructure/database/migrations/000042_create_campaigns.up.sql`
- **TODO.md**: Section "3. NEW ENTITY: Chat (CRITICAL)" mentions Campaign integration

---

**Last Updated**: 2025-10-12
**Status**: ✅ Domain Complete, ❌ Application/API Incomplete
**Priority**: CRITICAL (Core automation feature)
**Estimated Completion**: 2 weeks (application layer + Temporal workflows + API + frontend builder)
