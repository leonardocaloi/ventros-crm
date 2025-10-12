# Sequence Aggregate

## Overview

The **Sequence** aggregate powers Ventros CRM's automated drip campaign functionality. It orchestrates time-based, sequential message delivery to contacts, enabling nurture campaigns, onboarding flows, and follow-up sequences - similar to Mailchimp Automation, Drip campaigns, or HubSpot Email Sequences.

A Sequence is a linear series of messages sent to enrolled contacts over time, with configurable delays between each message, conditional logic for step execution, and automatic exit conditions. Unlike Campaigns (which support complex branching and multi-channel workflows), Sequences focus on simple, time-based message delivery.

- **Purpose**: Automated sequential messaging with time delays and conditions
- **Location**: `internal/domain/automation/sequence/`
- **Entity**: `infrastructure/persistence/entities/sequence.go`
- **Type**: Core automation aggregate (HIGH priority for nurture marketing)

---

## Domain Model

### Aggregate Root: Sequence

```go
type Sequence struct {
    id             uuid.UUID
    version        int // Optimistic locking
    tenantID       string
    name           string
    description    string
    status         SequenceStatus
    steps          []SequenceStep // Ordered steps

    // Entry conditions
    triggerType TriggerType            // How contacts enter
    triggerData map[string]interface{} // Trigger-specific config

    // Exit conditions
    exitOnReply bool // Exit if contact replies

    // Stats
    totalEnrolled  int
    activeCount    int
    completedCount int
    exitedCount    int

    createdAt time.Time
    updatedAt time.Time

    events []shared.DomainEvent
}
```

### Value Objects & Entities

#### 1. **SequenceStatus** (Value Object)
```go
type SequenceStatus string

const (
    SequenceStatusDraft    SequenceStatus = "draft"    // Being configured
    SequenceStatusActive   SequenceStatus = "active"   // Accepting enrollments
    SequenceStatusPaused   SequenceStatus = "paused"   // Temporarily stopped
    SequenceStatusArchived SequenceStatus = "archived" // Archived (soft delete)
)
```

**Status Transitions**:
- draft → active (via Activate)
- active → paused (via Pause)
- paused → active (via Resume)
- any → archived (via Archive)

#### 2. **TriggerType** (Value Object)
```go
type TriggerType string

const (
    TriggerTypeManual          TriggerType = "manual"           // Manual enrollment
    TriggerTypeTagAdded        TriggerType = "tag_added"        // When tag added to contact
    TriggerTypeListJoined      TriggerType = "list_joined"      // When contact joins list
    TriggerTypeFormSubmit      TriggerType = "form_submit"      // When form submitted
    TriggerTypePipelineEntered TriggerType = "pipeline_entered" // When contact enters pipeline
)
```

**Trigger Data Examples**:
```go
// tag_added trigger
triggerData: {
    "tag_id": "uuid",
    "tag_name": "lead-magnet-download"
}

// list_joined trigger
triggerData: {
    "list_id": "uuid",
    "list_name": "newsletter-subscribers"
}

// form_submit trigger
triggerData: {
    "form_id": "uuid",
    "form_name": "free-trial-signup"
}
```

#### 3. **SequenceStep** (Entity)
```go
type SequenceStep struct {
    ID              uuid.UUID
    Order           int             // Execution order (0, 1, 2, ...)
    Name            string          // Step name (e.g., "Welcome Message")
    DelayAmount     int             // Time to wait (5, 24, 7)
    DelayUnit       DelayUnit       // minutes, hours, days
    MessageTemplate MessageTemplate // Message to send
    Conditions      []StepCondition // Execution conditions
    CreatedAt       time.Time
}
```

**DelayUnit**:
```go
type DelayUnit string

const (
    DelayUnitMinutes DelayUnit = "minutes"
    DelayUnitHours   DelayUnit = "hours"
    DelayUnitDays    DelayUnit = "days"
)
```

**MessageTemplate**:
```go
type MessageTemplate struct {
    Type       string            `json:"type"` // text, template, media
    Content    string            `json:"content"`
    TemplateID *string           `json:"template_id,omitempty"` // WhatsApp template
    Variables  map[string]string `json:"variables,omitempty"`   // Template variables
    MediaURL   *string           `json:"media_url,omitempty"`   // Image/video URL
}
```

**Example Templates**:
```json
// Simple text message
{
    "type": "text",
    "content": "Hi {{contact.name}}, thanks for signing up!",
    "variables": {
        "contact.name": "John"
    }
}

// WhatsApp template message
{
    "type": "template",
    "template_id": "welcome_message_v1",
    "variables": {
        "1": "John",
        "2": "Premium Plan"
    }
}

// Media message
{
    "type": "media",
    "content": "Check out this product demo!",
    "media_url": "https://cdn.example.com/demo.mp4"
}
```

**StepCondition**:
```go
type StepCondition struct {
    Type     ConditionType `json:"type"`
    Operator string        `json:"operator"` // equals, contains, greater_than
    Value    string        `json:"value"`
}

type ConditionType string

const (
    ConditionTypeTag            ConditionType = "tag"             // Has specific tag
    ConditionTypeCustomField    ConditionType = "custom_field"    // Custom field value
    ConditionTypePipelineStatus ConditionType = "pipeline_status" // Pipeline status
    ConditionTypeLastActivity   ConditionType = "last_activity"   // Last activity date
)
```

**Condition Examples**:
```json
// Only send if contact has "vip" tag
{
    "type": "tag",
    "operator": "has",
    "value": "vip"
}

// Only send if company size > 50
{
    "type": "custom_field",
    "operator": "greater_than",
    "value": "50",
    "field": "company_size"
}

// Only send if in "qualified" pipeline stage
{
    "type": "pipeline_status",
    "operator": "equals",
    "value": "qualified"
}
```

#### 4. **SequenceEnrollment** (Separate Aggregate)
```go
type SequenceEnrollment struct {
    id               uuid.UUID
    sequenceID       uuid.UUID
    contactID        uuid.UUID
    status           EnrollmentStatus // active, paused, completed, exited
    currentStepOrder int              // Current step (0, 1, 2, ...)
    nextScheduledAt  *time.Time       // When to send next message

    // Exit tracking
    exitedAt   *time.Time
    exitReason *string

    // Completion tracking
    completedAt *time.Time

    enrolledAt time.Time
    updatedAt  time.Time

    events []shared.DomainEvent
}
```

**Enrollment Statuses**:
```go
type EnrollmentStatus string

const (
    EnrollmentStatusActive    EnrollmentStatus = "active"    // Progressing through steps
    EnrollmentStatusPaused    EnrollmentStatus = "paused"    // Temporarily stopped
    EnrollmentStatusCompleted EnrollmentStatus = "completed" // Finished all steps
    EnrollmentStatusExited    EnrollmentStatus = "exited"    // Exited early
)
```

**Exit Reasons**:
- `"contact_replied"` - Contact replied to message (if exitOnReply=true)
- `"contact_unsubscribed"` - Contact unsubscribed
- `"contact_deleted"` - Contact was deleted
- `"manual_exit"` - Manually exited by user
- `"condition_not_met"` - Step condition not met
- `"sequence_paused"` - Parent sequence paused
- `"error"` - Technical error occurred

---

## Business Invariants

### Sequence Invariants

1. **Name Required**: Sequence must have a non-empty name
2. **Tenant Isolation**: Sequence belongs to exactly one tenant
3. **Status Transitions**: Only valid transitions allowed (see SequenceStatus)
4. **Steps Required for Activation**: Cannot activate sequence without steps
5. **Step Order Uniqueness**: Each step must have unique order number (0, 1, 2, ...)
6. **Draft-Only Modifications**: Steps can only be added/removed/updated in draft status
7. **Non-Negative Delays**: DelayAmount must be >= 0
8. **Valid Delay Units**: DelayUnit must be minutes, hours, or days
9. **Optimistic Locking**: Version field prevents concurrent update conflicts
10. **Exit On Reply Default**: exitOnReply defaults to true (best practice)

### Enrollment Invariants

1. **Valid IDs**: sequenceID and contactID cannot be nil
2. **Active Advancement**: Only active enrollments can advance to next step
3. **No Double Completion**: Cannot complete already completed enrollment
4. **No Double Exit**: Cannot exit already exited/completed enrollment
5. **Status Consistency**: nextScheduledAt is nil when completed/exited
6. **Unique Active Enrollment**: One contact can only have one active enrollment per sequence
7. **Step Order Monotonic**: currentStepOrder only increases (never decreases)

---

## Events Emitted

### Sequence Lifecycle Events

1. **`sequence.created`**
   ```go
   SequenceCreatedEvent {
       SequenceID uuid.UUID
       TenantID   string
       Name       string
   }
   ```
   **When**: Sequence is created in draft status
   **Handlers**: Log creation, update analytics dashboard

2. **`sequence.activated`**
   ```go
   SequenceActivatedEvent {
       SequenceID uuid.UUID
   }
   ```
   **When**: Sequence transitions from draft to active
   **Handlers**: Start accepting enrollments, notify admins

3. **`sequence.paused`**
   ```go
   SequencePausedEvent {
       SequenceID uuid.UUID
   }
   ```
   **When**: Active sequence is paused
   **Handlers**: Pause all active enrollments, stop processing

4. **`sequence.resumed`**
   ```go
   SequenceResumedEvent {
       SequenceID uuid.UUID
   }
   ```
   **When**: Paused sequence is resumed
   **Handlers**: Resume active enrollments, restart processing

5. **`sequence.archived`**
   ```go
   SequenceArchivedEvent {
       SequenceID uuid.UUID
   }
   ```
   **When**: Sequence is archived (soft delete)
   **Handlers**: Remove from active lists, update analytics

### Step Management Events

6. **`sequence.step_added`**
   ```go
   SequenceStepAddedEvent {
       SequenceID uuid.UUID
       StepID     uuid.UUID
       Order      int
   }
   ```
   **When**: New step is added to sequence (draft only)
   **Handlers**: Update sequence preview, recalculate total duration

### Enrollment Events

7. **`sequence.contact_enrolled`**
   ```go
   ContactEnrolledEvent {
       EnrollmentID uuid.UUID
       SequenceID   uuid.UUID
       ContactID    uuid.UUID
   }
   ```
   **When**: Contact is enrolled in sequence
   **Handlers**: Schedule first step, update contact timeline, increment sequence stats

8. **`sequence.enrollment_advanced`**
   ```go
   EnrollmentAdvancedEvent {
       EnrollmentID uuid.UUID
       NewStepOrder int
   }
   ```
   **When**: Enrollment advances to next step
   **Handlers**: Schedule next step execution, update progress tracking

9. **`sequence.enrollment_completed`**
   ```go
   EnrollmentCompletedEvent {
       EnrollmentID uuid.UUID
   }
   ```
   **When**: Enrollment completes all steps successfully
   **Handlers**: Update sequence stats, mark contact as nurtured, trigger conversion tracking

10. **`sequence.enrollment_exited`**
    ```go
    EnrollmentExitedEvent {
        EnrollmentID uuid.UUID
        Reason       string
    }
    ```
    **When**: Enrollment exits early (reply, unsubscribe, error)
    **Handlers**: Log exit reason, update sequence stats, clean up scheduled jobs

11. **`sequence.enrollment_paused`**
    ```go
    EnrollmentPausedEvent {
        EnrollmentID uuid.UUID
    }
    ```
    **When**: Individual enrollment is paused
    **Handlers**: Cancel scheduled next step

12. **`sequence.enrollment_resumed`**
    ```go
    EnrollmentResumedEvent {
        EnrollmentID uuid.UUID
    }
    ```
    **When**: Paused enrollment is resumed
    **Handlers**: Reschedule next step

**Total Events**: 12 (5 sequence lifecycle + 1 step management + 6 enrollment)

---

## Repository Interface

### Sequence Repository

```go
type Repository interface {
    Save(sequence *Sequence) error
    FindByID(id uuid.UUID) (*Sequence, error)
    FindByTenantID(tenantID string) ([]*Sequence, error)
    FindActiveByTriggerType(triggerType TriggerType) ([]*Sequence, error)
    FindByStatus(status SequenceStatus) ([]*Sequence, error)
    Delete(id uuid.UUID) error
}
```

**Query Methods**:
- `FindByTenantID`: Get all sequences for tenant (filtered by tenant)
- `FindActiveByTriggerType`: Find active sequences with specific trigger (for auto-enrollment)
- `FindByStatus`: Filter by status (draft, active, paused, archived)

### Enrollment Repository

```go
type EnrollmentRepository interface {
    Save(enrollment *SequenceEnrollment) error
    FindByID(id uuid.UUID) (*SequenceEnrollment, error)
    FindBySequenceID(sequenceID uuid.UUID) ([]*SequenceEnrollment, error)
    FindByContactID(contactID uuid.UUID) ([]*SequenceEnrollment, error)
    FindReadyForNextStep() ([]*SequenceEnrollment, error)
    FindActiveBySequenceAndContact(sequenceID, contactID uuid.UUID) (*SequenceEnrollment, error)
    Delete(id uuid.UUID) error
}
```

**Critical Query**: `FindReadyForNextStep()` - Returns enrollments where:
- status = 'active'
- nextScheduledAt <= NOW()

This query powers the sequence execution worker.

**Implementation**: `infrastructure/persistence/gorm_sequence_repository.go`
- Uses GORM with PostgreSQL
- Optimistic locking via version field
- Transaction support for step management
- Cascade deletes for steps and enrollments
- Efficient indexes (see migration)

---

## Commands (NOT Implemented)

**TODO**: Create command layer for sequences:

```go
// NOT IMPLEMENTED - Suggested commands

type CreateSequenceCommand struct {
    TenantID    string
    Name        string
    Description string
    TriggerType string
}

type AddSequenceStepCommand struct {
    SequenceID      uuid.UUID
    Order           int
    Name            string
    DelayAmount     int
    DelayUnit       string
    MessageTemplate MessageTemplate
    Conditions      []StepCondition
}

type UpdateSequenceStepCommand struct {
    SequenceID      uuid.UUID
    StepID          uuid.UUID
    Name            *string
    DelayAmount     *int
    DelayUnit       *string
    MessageTemplate *MessageTemplate
}

type RemoveSequenceStepCommand struct {
    SequenceID uuid.UUID
    StepID     uuid.UUID
}

type ActivateSequenceCommand struct {
    SequenceID uuid.UUID
}

type PauseSequenceCommand struct {
    SequenceID uuid.UUID
}

type EnrollContactCommand struct {
    SequenceID uuid.UUID
    ContactID  uuid.UUID
}

type ExitEnrollmentCommand struct {
    EnrollmentID uuid.UUID
    Reason       string
}
```

---

## Queries (NOT Implemented)

**TODO**: Create query layer for sequences:

```go
// NOT IMPLEMENTED - Suggested queries

type ListSequencesQuery struct {
    TenantID string
    Status   *SequenceStatus
    Page     int
    Limit    int
}

type GetSequenceStatsQuery struct {
    SequenceID uuid.UUID
}

type ListSequenceEnrollmentsQuery struct {
    SequenceID uuid.UUID
    Status     *EnrollmentStatus
    Page       int
    Limit      int
}

type GetContactEnrollmentsQuery struct {
    ContactID uuid.UUID
    Status    *EnrollmentStatus
}

type GetEnrollmentHistoryQuery struct {
    EnrollmentID uuid.UUID
}
```

---

## Use Cases

### Implemented (HTTP Handler Layer)

**Location**: `infrastructure/http/handlers/sequence_handler.go`

**Status**: HTTP handlers implemented, but application layer use cases are MISSING.

Current handlers directly manipulate domain aggregates (NOT following clean architecture):

1. **ListSequences** - List sequences with pagination and status filter
2. **CreateSequence** - Create new sequence with steps
3. **GetSequence** - Get sequence by ID
4. **UpdateSequence** - Update sequence name/description/exitOnReply (draft only)
5. **ActivateSequence** - Activate sequence to start accepting enrollments
6. **PauseSequence** - Pause active sequence
7. **ResumeSequence** - Resume paused sequence
8. **ArchiveSequence** - Archive sequence
9. **DeleteSequence** - Delete sequence (draft only)
10. **GetSequenceStats** - Get sequence statistics
11. **EnrollContact** - Enroll contact in sequence
12. **ListEnrollments** - List enrollments for sequence

### Suggested Use Cases (Application Layer)

**TODO**: Move business logic from handlers to application layer:

```go
// internal/application/automation/sequence/

// 1. CreateSequenceUseCase
//    - Validates input
//    - Creates Sequence aggregate
//    - Saves to repository
//    - Publishes sequence.created event
//    - Returns SequenceDTO

// 2. AddStepToSequenceUseCase
//    - Loads sequence
//    - Validates sequence is in draft status
//    - Validates step configuration
//    - Adds step via Sequence.AddStep()
//    - Saves sequence
//    - Publishes sequence.step_added event
//    - Returns updated SequenceDTO

// 3. ActivateSequenceUseCase
//    - Loads sequence
//    - Validates sequence has steps
//    - Calls Sequence.Activate()
//    - Saves sequence
//    - Publishes sequence.activated event
//    - Starts processing enrollments

// 4. EnrollContactInSequenceUseCase (CRITICAL)
//    - Loads sequence (must be active)
//    - Loads contact (validate exists)
//    - Checks if already enrolled (one active enrollment per sequence)
//    - Gets first step delay
//    - Creates SequenceEnrollment
//    - Saves enrollment
//    - Publishes sequence.contact_enrolled event
//    - Schedules first step execution
//    - Updates sequence stats (IncrementEnrolled)

// 5. ExecuteSequenceStepUseCase (CRITICAL - MISSING!)
//    - Loads enrollment (via FindReadyForNextStep)
//    - Loads sequence
//    - Gets current step by order
//    - Validates step conditions (tag, custom_field, etc.)
//    - IF conditions not met → Exit enrollment
//    - IF conditions met → Send message via messaging service
//    - Advances enrollment to next step
//    - IF no next step → Complete enrollment
//    - IF has next step → Schedule next execution
//    - Publishes sequence.enrollment_advanced event
//    - Updates sequence stats

// 6. HandleContactReplyUseCase (CRITICAL - MISSING!)
//    - Triggered when contact replies to message
//    - Loads all active enrollments for contact
//    - For each enrollment:
//      * Load sequence
//      * IF sequence.exitOnReply = true → Exit enrollment with reason "contact_replied"
//    - Publishes sequence.enrollment_exited events
//    - Updates sequence stats

// 7. PauseSequenceUseCase
//    - Loads sequence
//    - Calls Sequence.Pause()
//    - Pauses all active enrollments (batch update)
//    - Saves sequence
//    - Publishes sequence.paused event

// 8. CompleteEnrollmentUseCase
//    - Loads enrollment
//    - Calls enrollment.Complete()
//    - Loads sequence
//    - Updates sequence stats (MarkCompleted)
//    - Saves enrollment and sequence
//    - Publishes sequence.enrollment_completed event

// 9. ExitEnrollmentUseCase
//    - Loads enrollment
//    - Calls enrollment.Exit(reason)
//    - Loads sequence
//    - Updates sequence stats (MarkExited)
//    - Saves enrollment and sequence
//    - Publishes sequence.enrollment_exited event

// 10. GetSequenceStatsUseCase
//     - Loads sequence
//     - Calculates completion rate, avg time to complete, etc.
//     - Returns SequenceStatsDTO
```

---

## Temporal Workflows (TODO)

**CRITICAL**: Sequence execution should be orchestrated by Temporal workflows for reliability.

### Suggested Workflows:

1. **SequenceEnrollmentWorkflow**
   ```go
   // Orchestrates a single contact's journey through a sequence
   //
   // Input:
   //   - EnrollmentID uuid.UUID
   //   - SequenceID uuid.UUID
   //   - ContactID uuid.UUID
   //
   // Activities:
   //   - LoadSequenceActivity() - Load sequence definition
   //   - LoadEnrollmentActivity() - Load enrollment state
   //   - GetCurrentStepActivity() - Get step to execute
   //   - ValidateConditionsActivity() - Check if conditions met
   //   - SendMessageActivity() - Send message via messaging service
   //   - AdvanceEnrollmentActivity() - Move to next step
   //   - CompleteEnrollmentActivity() - Mark as completed
   //   - ExitEnrollmentActivity() - Exit early
   //
   // Flow:
   //   1. Loop through steps (0, 1, 2, ...)
   //   2. For each step:
   //      a. Sleep(step.DelayDuration) - Wait for delay
   //      b. Validate conditions
   //      c. IF conditions not met → Exit
   //      d. Send message
   //      e. Advance to next step
   //      f. IF no next step → Complete
   //   3. Handle signals: pause, resume, exit
   //
   // Signals:
   //   - PauseSignal - Pause enrollment
   //   - ResumeSignal - Resume enrollment
   //   - ExitSignal(reason) - Exit enrollment
   //
   // Error Handling:
   //   - Retry failed message sends (3 attempts)
   //   - Exit on unrecoverable errors
   //   - Log all errors for debugging
   ```

2. **SequenceActivationWorkflow**
   ```go
   // Activates a sequence and starts processing enrollments
   //
   // Input:
   //   - SequenceID uuid.UUID
   //
   // Activities:
   //   - ActivateSequenceActivity() - Activate sequence
   //   - FindEligibleContactsActivity() - Query contacts matching trigger
   //   - EnrollContactsActivity() - Batch enroll contacts
   //   - StartEnrollmentWorkflowsActivity() - Start individual workflows
   //
   // Flow:
   //   1. Activate sequence
   //   2. IF triggerType is auto (tag_added, list_joined, etc.):
   //      a. Find eligible contacts
   //      b. Batch enroll contacts
   //      c. Start SequenceEnrollmentWorkflow for each
   //   3. Setup trigger listener for future enrollments
   ```

3. **SequenceExecutorWorkflow** (Background Worker)
   ```go
   // Background worker that processes ready enrollments
   //
   // Schedule: Runs every 1 minute (cron)
   //
   // Activities:
   //   - FindReadyEnrollmentsActivity() - Query enrollments ready for next step
   //   - ExecuteStepActivity() - Execute step for enrollment
   //
   // Flow:
   //   1. Query FindReadyForNextStep()
   //   2. For each enrollment:
   //      a. Load sequence
   //      b. Get current step
   //      c. Validate conditions
   //      d. Send message
   //      e. Advance enrollment
   //   3. Sleep 1 minute
   //   4. Repeat
   //
   // Alternative: Use Temporal Schedules instead of cron
   ```

**Implementation Status**: NOT IMPLEMENTED

---

## HTTP API (Implemented)

### Endpoints:

```yaml
# Sequence Management
GET    /api/v1/automation/sequences              # List sequences
POST   /api/v1/automation/sequences              # Create sequence
GET    /api/v1/automation/sequences/:id          # Get sequence
PUT    /api/v1/automation/sequences/:id          # Update sequence
DELETE /api/v1/automation/sequences/:id          # Delete sequence (draft only)

# Sequence Actions
POST   /api/v1/automation/sequences/:id/activate # Activate sequence
POST   /api/v1/automation/sequences/:id/pause    # Pause sequence
POST   /api/v1/automation/sequences/:id/resume   # Resume sequence
POST   /api/v1/automation/sequences/:id/archive  # Archive sequence

# Sequence Statistics
GET    /api/v1/automation/sequences/:id/stats    # Get sequence stats

# Enrollment Management
POST   /api/v1/automation/sequences/:id/enroll   # Enroll contact
GET    /api/v1/automation/sequences/:id/enrollments # List enrollments
```

**Implementation**: `infrastructure/http/handlers/sequence_handler.go`

**API Documentation**: Swagger annotations in handler (see `@Router` tags)

---

## Real-World Usage Patterns

### Pattern 1: Welcome Onboarding Sequence

```go
// 5-step onboarding sequence for new users

// Step 0: Immediate welcome message
step0 := NewSequenceStep(
    0,
    "Welcome Message",
    0,           // No delay (immediate)
    DelayUnitMinutes,
    MessageTemplate{
        Type: "text",
        Content: "Welcome to Ventros CRM! 👋 We're excited to have you on board.",
    },
)

// Step 1: Setup guide (1 day later)
step1 := NewSequenceStep(
    1,
    "Setup Guide",
    1,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "Here's a quick guide to get you started: [link]",
    },
)

// Step 2: Feature highlight (3 days later)
step2 := NewSequenceStep(
    2,
    "Feature Highlight",
    3,
    DelayUnitDays,
    MessageTemplate{
        Type: "media",
        Content: "Check out our most popular feature!",
        MediaURL: &demoVideoURL,
    },
)

// Step 3: Check-in (7 days later, only if not replied)
step3 := NewSequenceStep(
    3,
    "Check-in Message",
    7,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "How's your experience so far? Reply if you need any help!",
    },
)
// Add condition: Only send if contact hasn't replied yet
step3.Conditions = []StepCondition{
    {
        Type:     ConditionTypeLastActivity,
        Operator: "is_null",
        Value:    "last_reply_at",
    },
}

// Step 4: Upsell (14 days later, only for active users)
step4 := NewSequenceStep(
    4,
    "Upgrade Offer",
    14,
    DelayUnitDays,
    MessageTemplate{
        Type: "template",
        TemplateID: &upgradeTemplateID,
        Variables: map[string]string{
            "discount_code": "UPGRADE20",
        },
    },
)

sequence.AddStep(step0)
sequence.AddStep(step1)
sequence.AddStep(step2)
sequence.AddStep(step3)
sequence.AddStep(step4)
```

### Pattern 2: Lead Nurture Sequence

```go
// 3-step lead nurture sequence with exit on reply

sequence, _ := NewSequence(
    tenantID,
    "Lead Nurture",
    "Nurture cold leads to warm prospects",
    TriggerTypeTagAdded, // Trigger when "cold-lead" tag added
)

sequence.UpdateExitOnReply(true) // Exit if lead replies

// Step 0: Introduction (5 minutes)
step0 := NewSequenceStep(
    0,
    "Introduction",
    5,
    DelayUnitMinutes,
    MessageTemplate{
        Type: "text",
        Content: "Hi {{contact.name}}, I saw you're interested in {{product}}. Would you like to schedule a demo?",
        Variables: map[string]string{
            "contact.name": "John",
            "product": "Ventros CRM",
        },
    },
)

// Step 1: Case study (2 days, only if not replied)
step1 := NewSequenceStep(
    1,
    "Case Study",
    2,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "Here's how we helped {{company}} increase conversions by 50%: [link]",
    },
)

// Step 2: Final offer (5 days)
step2 := NewSequenceStep(
    2,
    "Final Offer",
    5,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "Last chance! Book a demo this week and get 20% off your first month.",
    },
)

sequence.AddStep(step0)
sequence.AddStep(step1)
sequence.AddStep(step2)
```

### Pattern 3: Re-engagement Sequence

```go
// 4-step re-engagement for inactive contacts

sequence, _ := NewSequence(
    tenantID,
    "Win-back Sequence",
    "Re-engage inactive contacts",
    TriggerTypeManual, // Manually enroll inactive contacts
)

// Step 0: "We miss you" (immediate)
step0 := NewSequenceStep(
    0,
    "We Miss You",
    0,
    DelayUnitMinutes,
    MessageTemplate{
        Type: "text",
        Content: "We noticed you haven't been around lately. Is everything okay?",
    },
)

// Step 1: Special offer (3 days)
step1 := NewSequenceStep(
    1,
    "Special Offer",
    3,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "As a token of appreciation, here's 25% off your next purchase: CODE25",
    },
)

// Step 2: Feedback request (7 days)
step2 := NewSequenceStep(
    2,
    "Feedback Request",
    7,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "Would you mind sharing why you stopped using our service? [survey link]",
    },
)

// Step 3: Final goodbye (14 days)
step3 := NewSequenceStep(
    3,
    "Final Goodbye",
    14,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "We're sad to see you go. If you ever need us, we're just a message away.",
    },
)

sequence.AddStep(step0)
sequence.AddStep(step1)
sequence.AddStep(step2)
sequence.AddStep(step3)
```

### Pattern 4: Event Follow-up Sequence

```go
// 3-step follow-up after webinar/event attendance

sequence, _ := NewSequence(
    tenantID,
    "Webinar Follow-up",
    "Follow up with webinar attendees",
    TriggerTypeFormSubmit, // Trigger when webinar registration form submitted
)

// Step 0: Thank you + recording (1 hour after webinar)
step0 := NewSequenceStep(
    0,
    "Thank You",
    1,
    DelayUnitHours,
    MessageTemplate{
        Type: "text",
        Content: "Thanks for attending today's webinar! Here's the recording: [link]",
    },
)

// Step 1: Resources (2 days)
step1 := NewSequenceStep(
    1,
    "Additional Resources",
    2,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "Here are the slides and additional resources we mentioned: [link]",
    },
)

// Step 2: Book consultation (5 days, only for engaged attendees)
step2 := NewSequenceStep(
    2,
    "Book Consultation",
    5,
    DelayUnitDays,
    MessageTemplate{
        Type: "text",
        Content: "Want to discuss how we can help you? Book a free consultation: [calendar link]",
    },
)
// Only send if contact engaged (opened/clicked previous messages)
step2.Conditions = []StepCondition{
    {
        Type:     ConditionTypeCustomField,
        Operator: "greater_than",
        Value:    "0",
        Field:    "engagement_score",
    },
}

sequence.AddStep(step0)
sequence.AddStep(step1)
sequence.AddStep(step2)
```

---

## Performance Considerations

### Scalability

- **Enrollments**: Can scale to millions of concurrent enrollments
- **Step Execution**: Process 1000+ messages per minute
- **Query Efficiency**: Use indexed queries (nextScheduledAt, status)
- **Batch Processing**: Process enrollments in batches of 100-500
- **Caching**: Cache sequence definitions (rarely change once active)

### Database Indexes

```sql
-- Already exist in migration 000042:

-- Sequences
CREATE INDEX idx_sequences_tenant ON sequences(tenant_id);
CREATE INDEX idx_sequences_status ON sequences(status);
CREATE INDEX idx_sequences_trigger_type ON sequences(trigger_type);

-- Sequence Steps
CREATE INDEX idx_sequence_steps_sequence_id ON sequence_steps(sequence_id);
CREATE UNIQUE INDEX idx_sequence_steps_sequence_order ON sequence_steps(sequence_id, "order");

-- Sequence Enrollments
CREATE INDEX idx_enrollments_sequence_id ON sequence_enrollments(sequence_id);
CREATE INDEX idx_enrollments_contact_id ON sequence_enrollments(contact_id);
CREATE INDEX idx_enrollments_status ON sequence_enrollments(status);
CREATE INDEX idx_enrollments_next_scheduled ON sequence_enrollments(next_scheduled_at)
    WHERE next_scheduled_at IS NOT NULL;

-- Unique constraint: One active enrollment per sequence+contact
CREATE UNIQUE INDEX idx_enrollments_sequence_contact_unique
    ON sequence_enrollments(sequence_id, contact_id)
    WHERE status = 'active';
```

### Optimizations

1. **Worker Query Optimization**:
   ```sql
   -- FindReadyForNextStep() query
   SELECT * FROM sequence_enrollments
   WHERE status = 'active'
     AND next_scheduled_at <= NOW()
   ORDER BY next_scheduled_at ASC
   LIMIT 500;
   ```
   - Uses partial index on next_scheduled_at (WHERE clause)
   - Batches 500 enrollments per execution
   - Ordered by next_scheduled_at (FIFO processing)

2. **Sequence Caching**:
   ```go
   // Cache sequence definitions in Redis
   type SequenceCache struct {
       sequences map[uuid.UUID]*Sequence
       ttl       time.Duration // 1 hour
   }
   // Invalidate cache on sequence update
   ```

3. **Condition Pre-compilation**:
   ```go
   // Pre-compile condition logic for faster evaluation
   type CompiledCondition struct {
       Type     ConditionType
       Evaluator func(contact *Contact) bool
   }
   ```

4. **Bulk Enrollment**:
   ```go
   // EnrollContactsInBatch(sequenceID, contactIDs []uuid.UUID)
   // Insert multiple enrollments in single transaction
   tx.Begin()
   for _, contactID := range contactIDs {
       enrollment := NewSequenceEnrollment(...)
       tx.Save(enrollment)
   }
   tx.Commit()
   ```

5. **Metrics Aggregation**:
   ```sql
   -- Materialized view for sequence statistics
   CREATE MATERIALIZED VIEW sequence_stats AS
   SELECT
       sequence_id,
       COUNT(*) FILTER (WHERE status = 'active') AS active_count,
       COUNT(*) FILTER (WHERE status = 'completed') AS completed_count,
       COUNT(*) FILTER (WHERE status = 'exited') AS exited_count,
       AVG(EXTRACT(EPOCH FROM (completed_at - enrolled_at))) AS avg_duration_seconds
   FROM sequence_enrollments
   GROUP BY sequence_id;

   -- Refresh every 5 minutes
   REFRESH MATERIALIZED VIEW sequence_stats;
   ```

---

## Testing Checklist

### Unit Tests (Domain Layer)

- [x] Sequence creation with valid data
- [x] Sequence status transitions (draft → active, pause/resume)
- [x] Step management (add, update, remove)
- [x] Step order uniqueness validation
- [x] Draft-only modification enforcement
- [x] Stats tracking (IncrementEnrolled, MarkCompleted, MarkExited)
- [x] Statistics calculation (completion rate)
- [x] Enrollment creation with valid data
- [x] Enrollment advancement (next step, completion)
- [x] Enrollment exit with reason
- [x] Enrollment pause/resume
- [x] IsReadyForNextStep() logic
- [x] GetDelayDuration() calculation (minutes/hours/days)
- [x] Domain events emission (12 events)

### Integration Tests (Repository Layer)

- [ ] Save and load sequence with steps
- [ ] Optimistic locking (version conflicts)
- [ ] Find sequences by tenant
- [ ] Find sequences by status
- [ ] Find sequences by trigger type
- [ ] Cascade delete (sequence → steps, enrollments)
- [ ] Enrollment queries (by sequence, by contact, ready for next step)
- [ ] Find active enrollment for sequence+contact
- [ ] Unique constraint enforcement (one active enrollment per sequence+contact)

### End-to-End Tests (Application Layer)

- [ ] Create sequence → Add steps → Activate
- [ ] Enroll contact → Execute steps → Complete
- [ ] Pause sequence → Resume sequence
- [ ] Enrollment exits on contact reply (if exitOnReply=true)
- [ ] Step condition validation (tag, custom_field, etc.)
- [ ] Multiple enrollments for same contact (different sequences)
- [ ] Sequence statistics accuracy
- [ ] Temporal workflow execution
- [ ] Error handling (message send failures, condition errors)

### Performance Tests

- [ ] Process 10,000 enrollments in < 10 minutes
- [ ] FindReadyForNextStep() query < 100ms
- [ ] Concurrent enrollment creation (no deadlocks)
- [ ] Batch enrollment (1000 contacts) < 5 seconds

---

## Suggested Improvements

### 1. Smart Sending (Send Time Optimization)

```go
type SmartSendingConfig struct {
    Enabled         bool
    RespectTimezone bool
    PreferredHours  []int // [9, 10, 11, ..., 17]
    PreferredDays   []time.Weekday // Monday-Friday
}

// Adjust nextScheduledAt based on contact's timezone and preferred hours
func (e *SequenceEnrollment) CalculateSmartSendTime(
    delay time.Duration,
    config SmartSendingConfig,
    contactTimezone string,
) time.Time {
    scheduledTime := time.Now().Add(delay)

    if !config.Enabled {
        return scheduledTime
    }

    // Adjust to contact's timezone
    if config.RespectTimezone {
        loc, _ := time.LoadLocation(contactTimezone)
        scheduledTime = scheduledTime.In(loc)
    }

    // Adjust to preferred hours
    hour := scheduledTime.Hour()
    if !contains(config.PreferredHours, hour) {
        // Move to next preferred hour
        scheduledTime = nextPreferredHour(scheduledTime, config.PreferredHours)
    }

    // Adjust to preferred days
    day := scheduledTime.Weekday()
    if !contains(config.PreferredDays, day) {
        // Move to next preferred day
        scheduledTime = nextPreferredDay(scheduledTime, config.PreferredDays)
    }

    return scheduledTime
}
```

### 2. A/B Testing Support

```go
type SequenceVariant struct {
    ID          uuid.UUID
    Name        string
    Steps       []SequenceStep
    SplitPercent float64 // 50.0 = 50%
}

type Sequence struct {
    // ... existing fields ...
    variants []SequenceVariant // A/B test variants
}

// Enroll contact in random variant based on split percentage
func (s *Sequence) EnrollInVariant(contactID uuid.UUID) *SequenceVariant {
    rand := random.Float64() * 100
    cumulative := 0.0
    for _, variant := range s.variants {
        cumulative += variant.SplitPercent
        if rand < cumulative {
            return &variant
        }
    }
    return &s.variants[0] // Default to first variant
}
```

### 3. Dynamic Content Personalization

```go
type DynamicVariable struct {
    Name   string // {{contact.name}}, {{company.size}}
    Source string // contact, company, custom_field
    Field  string // name, size, industry
}

// Resolve variables from contact data
func (mt *MessageTemplate) Resolve(contact *Contact) string {
    content := mt.Content
    for key, value := range mt.Variables {
        // Support dynamic resolution
        if strings.HasPrefix(value, "{{") {
            resolvedValue := resolveFromContact(contact, value)
            content = strings.Replace(content, key, resolvedValue, -1)
        } else {
            content = strings.Replace(content, key, value, -1)
        }
    }
    return content
}
```

### 4. Re-enrollment Rules

```go
type ReEnrollmentConfig struct {
    Enabled         bool
    AllowAfterDays  int  // Allow re-enrollment after X days
    AllowAfterExit  bool // Allow re-enrollment after exit
    AllowAfterComplete bool // Allow re-enrollment after completion
}

// Check if contact can be re-enrolled
func (s *Sequence) CanReEnroll(contactID uuid.UUID, lastEnrollment *SequenceEnrollment) bool {
    if !s.reEnrollmentConfig.Enabled {
        return false
    }

    if lastEnrollment.Status() == EnrollmentStatusExited && !s.reEnrollmentConfig.AllowAfterExit {
        return false
    }

    if lastEnrollment.Status() == EnrollmentStatusCompleted && !s.reEnrollmentConfig.AllowAfterComplete {
        return false
    }

    daysSinceLastEnrollment := time.Since(lastEnrollment.UpdatedAt()).Hours() / 24
    if daysSinceLastEnrollment < float64(s.reEnrollmentConfig.AllowAfterDays) {
        return false
    }

    return true
}
```

### 5. Step Templates Library

```go
type StepTemplate struct {
    ID          uuid.UUID
    Name        string
    Description string
    Category    string // onboarding, nurture, re-engagement
    DelayAmount int
    DelayUnit   DelayUnit
    MessageTemplate MessageTemplate
}

// Pre-built templates
var StepTemplates = []StepTemplate{
    {
        Name: "Welcome Message",
        Category: "onboarding",
        DelayAmount: 0,
        DelayUnit: DelayUnitMinutes,
        MessageTemplate: MessageTemplate{
            Type: "text",
            Content: "Welcome to {{company.name}}! We're excited to have you.",
        },
    },
    {
        Name: "Setup Guide",
        Category: "onboarding",
        DelayAmount: 1,
        DelayUnit: DelayUnitDays,
        MessageTemplate: MessageTemplate{
            Type: "text",
            Content: "Here's a quick guide to get you started: [link]",
        },
    },
    // ... more templates
}
```

### 6. Sequence Cloning

```go
func (s *Sequence) Clone(newName string, tenantID string) (*Sequence, error) {
    clone := &Sequence{
        id:             uuid.New(),
        version:        1,
        tenantID:       tenantID,
        name:           newName,
        description:    s.description + " (cloned)",
        status:         SequenceStatusDraft, // Always draft
        triggerType:    s.triggerType,
        triggerData:    copyMap(s.triggerData),
        exitOnReply:    s.exitOnReply,
        steps:          []SequenceStep{},
        createdAt:      time.Now(),
        updatedAt:      time.Now(),
        events:         []shared.DomainEvent{},
    }

    // Clone steps
    for _, step := range s.steps {
        clonedStep := SequenceStep{
            ID:              uuid.New(),
            Order:           step.Order,
            Name:            step.Name,
            DelayAmount:     step.DelayAmount,
            DelayUnit:       step.DelayUnit,
            MessageTemplate: step.MessageTemplate,
            Conditions:      append([]StepCondition{}, step.Conditions...),
            CreatedAt:       time.Now(),
        }
        clone.steps = append(clone.steps, clonedStep)
    }

    clone.addEvent(NewSequenceCreatedEvent(clone.id, tenantID, newName))

    return clone, nil
}
```

---

## Related Aggregates

- **Campaign**: Can reference Sequence in StepTypeSequence steps (campaigns can trigger sequences)
- **Contact**: Enrolled in sequences via SequenceEnrollment
- **Message**: Sent as part of sequence steps (via messaging service)
- **Tag**: Used in sequence triggers (TriggerTypeTagAdded) and step conditions
- **Broadcast**: Similar but for one-time mass messaging (vs. sequential)

**Integration Points**:
1. Campaign → Sequence (step.Config.SequenceID)
2. Tag Added Event → Sequence Auto-enrollment
3. Contact Reply Event → Sequence Exit (if exitOnReply=true)
4. Sequence Step → Message Sending (via messaging service)

---

## Industry Comparison

| Feature | Ventros CRM | Mailchimp | Drip | HubSpot | ActiveCampaign |
|---------|-------------|-----------|------|---------|----------------|
| Sequential messaging | ✅ | ✅ | ✅ | ✅ | ✅ |
| Time delays | ✅ | ✅ | ✅ | ✅ | ✅ |
| Step conditions | ✅ | ✅ | ✅ | ✅ | ✅ |
| Exit on reply | ✅ | ❌ | ✅ | ✅ | ✅ |
| A/B testing | ❌ | ✅ | ✅ | ✅ | ✅ |
| Smart sending | ❌ | ✅ | ✅ | ✅ | ✅ |
| Re-enrollment rules | ❌ | ✅ | ✅ | ✅ | ✅ |
| Multi-channel | ✅ | ❌ | ❌ | ✅ | ✅ |
| Template library | ❌ | ✅ | ✅ | ✅ | ✅ |
| Visual builder | ❌ | ✅ | ✅ | ✅ | ✅ |
| Analytics | ⚠️ | ✅ | ✅ | ✅ | ✅ |

**Ventros Strengths**:
- Clean DDD architecture (domain-driven design)
- Multi-channel support (WhatsApp, SMS, Email)
- Event-driven system (full event sourcing)
- Temporal-based reliability (durable execution)
- Optimistic locking (prevents lost updates)

**Suggested Additions**:
- A/B testing for sequences
- Smart sending (timezone-aware, preferred hours)
- Re-enrollment rules
- Template library
- Visual sequence builder (frontend)
- Advanced analytics (open rate, reply rate, conversion tracking)

---

## Documentation References

- **Domain Events**: `internal/domain/automation/sequence/events.go`
- **Domain Model**: `internal/domain/automation/sequence/sequence.go`
- **Enrollment**: `internal/domain/automation/sequence/sequence_enrollment.go`
- **Steps**: `internal/domain/automation/sequence/sequence_step.go`
- **Repository Interface**: Defined in `sequence.go` (lines 375-382)
- **Implementation**: `infrastructure/persistence/gorm_sequence_repository.go`
- **Entities**: `infrastructure/persistence/entities/sequence.go`
- **Migration**: `infrastructure/database/migrations/000042_create_sequences.up.sql`
- **HTTP Handlers**: `infrastructure/http/handlers/sequence_handler.go`

---

## Differences from Campaign Aggregate

| Aspect | Sequence | Campaign |
|--------|----------|----------|
| **Complexity** | Simple, linear flow | Complex, branching workflows |
| **Step Types** | Messages only | Broadcast, sequence, delay, condition, wait |
| **Conditional Logic** | Per-step conditions | Per-step + branching |
| **Scheduling** | Relative delays (from enrollment) | Absolute + relative scheduling |
| **Entry** | Multiple triggers | Manual enrollment + criteria |
| **Exit** | Exit on reply (default) | Manual exit + goal completion |
| **Use Case** | Nurture, onboarding, follow-ups | Complex automation workflows |
| **Goal Tracking** | Stats only (no goals) | Goal types + conversion tracking |
| **Multi-channel** | ✅ (WhatsApp, SMS, Email) | ✅ (via broadcasts/sequences) |

**When to Use Sequence vs Campaign**:
- **Use Sequence**: Simple drip campaigns, onboarding, follow-ups
- **Use Campaign**: Complex workflows with branching, multi-step automation, goal-based campaigns

**Example**: Welcome onboarding with 5 messages → **Sequence**
**Example**: Lead nurture with conditional branching, wait for reply, multiple channels → **Campaign**

---

**Last Updated**: 2025-10-12
**Status**: ✅ Domain Complete, ✅ HTTP Handlers Implemented, ❌ Application Layer Missing, ❌ Temporal Workflows Missing
**Priority**: HIGH (Core automation feature)
**Estimated Completion**: 1 week (application layer + Temporal workflows + advanced features)
