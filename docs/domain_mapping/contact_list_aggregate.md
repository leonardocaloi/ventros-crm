# ContactList Aggregate

## Overview

The **ContactList** aggregate is Ventros CRM's segmentation and grouping engine for organizing and targeting contacts. It enables powerful audience segmentation through both static (manual) and dynamic (rule-based) list management - similar to HubSpot Lists, Mailchimp Audiences, Salesforce Reports, or ActiveCampaign Lists.

A ContactList can be either **static** (contacts manually added/removed) or **dynamic** (contacts automatically included based on filter rules that are evaluated in real-time), with support for complex filtering by tags, custom fields, pipeline status, contact attributes, events, and behavioral data.

- **Purpose**: Segmentation, grouping, and targeting mechanism for marketing automation
- **Location**: `internal/domain/crm/contact_list/`
- **Entity**: `infrastructure/persistence/entities/contact_list.go`
- **Type**: Core CRM aggregate (CRITICAL for marketing automation and broadcasts)

---

## Domain Model

### Aggregate Root: ContactList

```go
type ContactList struct {
    id               uuid.UUID
    version          int             // Optimistic locking
    projectID        uuid.UUID
    tenantID         string
    name             string
    description      *string
    filterRules      []*FilterRule
    logicalOperator  LogicalOperator // AND or OR
    isStatic         bool
    contactCount     int
    lastCalculatedAt *time.Time
    createdAt        time.Time
    updatedAt        time.Time
    deletedAt        *time.Time
    events           []shared.DomainEvent
}
```

### Value Objects & Entities

#### 1. **LogicalOperator** (Value Object)
```go
type LogicalOperator string

const (
    LogicalOperatorAND LogicalOperator = "AND"
    LogicalOperatorOR  LogicalOperator = "OR"
)
```

**Purpose**: Defines how multiple filter rules are combined
- `AND` - Contact must match ALL filter rules (intersection)
- `OR` - Contact must match ANY filter rule (union)

**Example**:
```go
// AND: Contacts with tag "premium" AND pipeline status "qualified"
// OR:  Contacts with tag "premium" OR pipeline status "qualified"
```

#### 2. **FilterRule** (Entity)
```go
type FilterRule struct {
    id         uuid.UUID
    filterType FilterType      // Type of filter (tag, custom_field, etc.)
    fieldKey   string          // Field being filtered
    fieldType  *shared.FieldType // Type of custom field (if custom_field)
    operator   FilterOperator   // Comparison operator
    value      interface{}      // Value to compare against
    pipelineID *uuid.UUID       // Pipeline ID (if pipeline_status)
    createdAt  time.Time
}
```

#### 3. **FilterType** (Value Object)
```go
type FilterType string

const (
    FilterTypeCustomField    FilterType = "custom_field"    // Custom field values
    FilterTypePipelineStatus FilterType = "pipeline_status" // Pipeline/status
    FilterTypeTag            FilterType = "tag"             // Contact tags
    FilterTypeEvent          FilterType = "event"           // Contact events
    FilterTypeInteraction    FilterType = "interaction"     // Interaction history
    FilterTypeAttribute      FilterType = "attribute"       // Contact attributes
)
```

**Filter Types Explained**:

- **`custom_field`**: Filter by custom field values
  - Example: "company_size > 500", "industry = technology"
  - Requires `fieldType` to be specified
  - Supports all field types (text, number, date, boolean, etc.)

- **`pipeline_status`**: Filter by pipeline and current status
  - Example: "In Sales Pipeline with status = Qualified"
  - Requires `pipelineID` to be specified
  - Common use: "Leads in qualification stage"

- **`tag`**: Filter by contact tags
  - Example: "Has tag 'vip'", "Has tag 'newsletter-subscriber'"
  - Most common filter type
  - Fast queries using JSONB containment operators

- **`event`**: Filter by events in contact timeline
  - Example: "Attended webinar", "Downloaded whitepaper"
  - Currently defined but may need additional implementation

- **`interaction`**: Filter by interaction history
  - Example: "Replied in last 7 days", "Opened last email"
  - Currently defined but may need additional implementation

- **`attribute`**: Filter by core contact attributes
  - Example: "language = pt-BR", "created_at > 2025-01-01"
  - Allowed attributes: name, email, phone, language, timezone, source_channel,
    first_interaction_at, last_interaction_at, created_at, updated_at

#### 4. **FilterOperator** (Value Object)
```go
type FilterOperator string

const (
    OperatorEquals       FilterOperator = "eq"           // Equals
    OperatorNotEquals    FilterOperator = "ne"           // Not equals
    OperatorGreaterThan  FilterOperator = "gt"           // Greater than
    OperatorLessThan     FilterOperator = "lt"           // Less than
    OperatorGreaterEqual FilterOperator = "gte"          // Greater or equal
    OperatorLessEqual    FilterOperator = "lte"          // Less or equal
    OperatorContains     FilterOperator = "contains"     // Contains substring
    OperatorNotContains  FilterOperator = "not_contains" // Not contains
    OperatorStartsWith   FilterOperator = "starts_with"  // Starts with
    OperatorEndsWith     FilterOperator = "ends_with"    // Ends with
    OperatorIn           FilterOperator = "in"           // In array
    OperatorNotIn        FilterOperator = "not_in"       // Not in array
    OperatorIsNull       FilterOperator = "is_null"      // Is null
    OperatorIsNotNull    FilterOperator = "is_not_null"  // Is not null
)
```

**Operator Validation**:
- `is_null`/`is_not_null`: Value must be nil
- `in`/`not_in`: Value must be array/slice
- All others: Value cannot be nil

---

## Business Invariants

### ContactList Invariants

1. **Name Required**: ContactList must have a non-empty name
2. **Project Required**: ProjectID cannot be nil
3. **Tenant Isolation**: ContactList belongs to exactly one tenant
4. **Logical Operator Valid**: Must be AND or OR
5. **Filter Rules Validation**: Each filter rule must be valid for its type
6. **Static vs Dynamic**:
   - Static lists: Contacts added/removed manually via membership table
   - Dynamic lists: Contacts determined by filter rules (evaluated at query time)
7. **Optimistic Locking**: Version field prevents lost updates
8. **Soft Delete**: Deleted lists are marked with deletedAt timestamp

### FilterRule Invariants

1. **Valid Filter Type**: FilterType must be one of the 6 defined types
2. **Valid Operator**: FilterOperator must be one of the 14 defined operators
3. **Field Key Required**: FieldKey cannot be empty
4. **Null Value Rules**:
   - `is_null`/`is_not_null` operators: value must be nil
   - `in`/`not_in` operators: value cannot be nil
   - All other operators: value cannot be nil (except is_null/is_not_null)
5. **Custom Field Type**: If FilterType is `custom_field`, fieldType must be specified
6. **Pipeline Required**: If FilterType is `pipeline_status`, pipelineID must be specified
7. **Attribute Validation**: If FilterType is `attribute`, fieldKey must be in allowed list

---

## Events Emitted

### ContactList Lifecycle Events

1. **`contact_list.created`**
   ```go
   ContactListCreatedEvent {
       ContactListID uuid.UUID
       ProjectID     uuid.UUID
       TenantID      string
       Name          string
       IsStatic      bool
   }
   ```
   **When**: ContactList is created
   **Handlers**: Log creation, initialize analytics tracking

2. **`contact_list.updated`**
   ```go
   ContactListUpdatedEvent {
       ContactListID uuid.UUID
       UpdatedFields []string  // ["name", "description", "logical_operator"]
   }
   ```
   **When**: ContactList name, description, or logical operator is updated
   **Handlers**: Invalidate cache, recalculate count if needed

3. **`contact_list.deleted`**
   ```go
   ContactListDeletedEvent {
       ContactListID uuid.UUID
   }
   ```
   **When**: ContactList is soft deleted
   **Handlers**: Remove from active lists, cleanup cached data, deactivate dependent automations

### Filter Management Events

4. **`contact_list.filter_rule_added`**
   ```go
   ContactListFilterRuleAddedEvent {
       ContactListID uuid.UUID
       FilterRuleID  uuid.UUID
       FilterType    string
   }
   ```
   **When**: New filter rule is added to list (dynamic lists only)
   **Handlers**: Trigger recalculation, invalidate cache

5. **`contact_list.filter_rule_removed`**
   ```go
   ContactListFilterRuleRemovedEvent {
       ContactListID uuid.UUID
       FilterRuleID  uuid.UUID
   }
   ```
   **When**: Filter rule is removed from list
   **Handlers**: Trigger recalculation, invalidate cache

6. **`contact_list.filter_rules_cleared`**
   ```go
   ContactListFilterRulesClearedEvent {
       ContactListID uuid.UUID
   }
   ```
   **When**: All filter rules are cleared from list
   **Handlers**: Trigger recalculation, clear cached results

### Recalculation Events

7. **`contact_list.recalculated`**
   ```go
   ContactListRecalculatedEvent {
       ContactListID uuid.UUID
       ContactCount  int
   }
   ```
   **When**: Contact count is recalculated (dynamic lists)
   **Handlers**: Update UI, trigger dependent workflows, update analytics

### Membership Events (Static Lists Only)

8. **`contact_list.contact_added`**
   ```go
   ContactAddedToListEvent {
       ContactListID uuid.UUID
       ContactID     uuid.UUID
   }
   ```
   **When**: Contact is manually added to static list
   **Handlers**: Update contact timeline, trigger automation rules, increment count

9. **`contact_list.contact_removed`**
   ```go
   ContactRemovedFromListEvent {
       ContactListID uuid.UUID
       ContactID     uuid.UUID
   }
   ```
   **When**: Contact is manually removed from static list
   **Handlers**: Update contact timeline, trigger exit rules, decrement count

**Total Events**: 9 (3 lifecycle + 3 filter management + 1 recalculation + 2 membership)

---

## Repository Interface

### ContactListRepository

```go
type Repository interface {
    // Basic CRUD
    Create(ctx context.Context, list *ContactList) error
    Update(ctx context.Context, list *ContactList) error
    Delete(ctx context.Context, id uuid.UUID) error

    // Queries
    FindByID(ctx context.Context, id uuid.UUID) (*ContactList, error)
    FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*ContactList, error)
    FindByTenantID(ctx context.Context, tenantID string) ([]*ContactList, error)
    ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*ContactList, int, error)

    // Contact queries
    GetContactsInList(ctx context.Context, listID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error)
    RecalculateContactCount(ctx context.Context, listID uuid.UUID) (int, error)

    // Static list membership
    AddContactToStaticList(ctx context.Context, listID, contactID uuid.UUID) error
    RemoveContactFromStaticList(ctx context.Context, listID, contactID uuid.UUID) error
    IsContactInList(ctx context.Context, listID, contactID uuid.UUID) (bool, error)
}
```

**Implementation**: `infrastructure/persistence/gorm_contact_list_repository.go`
- Uses GORM with PostgreSQL
- Optimistic locking via version field
- Transaction support for filter rule management
- Dynamic query building for filter rule evaluation
- JSONB operators for tag filtering
- Subqueries for custom fields and pipeline status
- Efficient pagination

### Key Implementation Details

#### Filter Rule Application

The repository implements sophisticated query building to handle different filter types:

1. **Attribute Filters**: Direct column comparisons
   ```sql
   WHERE name ILIKE '%john%'
   WHERE created_at > '2025-01-01'
   WHERE language = 'pt-BR'
   ```

2. **Tag Filters**: JSONB containment operators
   ```sql
   WHERE tags @> '["premium"]'::jsonb
   WHERE tags @> '["vip","customer"]'::jsonb
   ```

3. **Custom Field Filters**: Subquery on contact_custom_fields table
   ```sql
   WHERE id IN (
       SELECT contact_id FROM contact_custom_fields
       WHERE field_key = 'company_size'
       AND field_value::int > 500
   )
   ```

4. **Pipeline Status Filters**: Subquery on contact_pipeline_statuses table
   ```sql
   WHERE id IN (
       SELECT contact_id FROM contact_pipeline_statuses
       WHERE pipeline_id = ?
       AND current_status = 'qualified'
   )
   ```

#### Static List Membership

Static lists use a separate `contact_list_members` table:
```sql
CREATE TABLE contact_list_members (
    id              uuid PRIMARY KEY,
    contact_list_id uuid NOT NULL,
    contact_id      uuid NOT NULL,
    added_at        timestamp NOT NULL,
    UNIQUE(contact_list_id, contact_id)
)
```

---

## Commands (NOT Implemented)

**TODO**: Create command layer for contact lists:

```go
// ❌ NOT IMPLEMENTED - Suggested commands

type CreateContactListCommand struct {
    ProjectID       uuid.UUID
    TenantID        string
    Name            string
    Description     *string
    LogicalOperator string
    IsStatic        bool
    FilterRules     []FilterRuleDTO
}

type UpdateContactListCommand struct {
    ContactListID   uuid.UUID
    Name            *string
    Description     *string
    LogicalOperator *string
    FilterRules     *[]FilterRuleDTO
}

type DeleteContactListCommand struct {
    ContactListID uuid.UUID
}

type AddContactToListCommand struct {
    ContactListID uuid.UUID
    ContactID     uuid.UUID
}

type RemoveContactFromListCommand struct {
    ContactListID uuid.UUID
    ContactID     uuid.UUID
}

type RecalculateListCommand struct {
    ContactListID uuid.UUID
}
```

---

## Queries (NOT Implemented)

**TODO**: Create query layer for contact lists:

```go
// ❌ NOT IMPLEMENTED - Suggested queries

type ListContactListsQuery struct {
    ProjectID uuid.UUID
    IsStatic  *bool
    Page      int
    Limit     int
}

type GetContactListQuery struct {
    ContactListID uuid.UUID
}

type GetContactsInListQuery struct {
    ContactListID uuid.UUID
    Page          int
    Limit         int
}

type GetListStatsQuery struct {
    ContactListID uuid.UUID
}

type SearchContactListsQuery struct {
    ProjectID uuid.UUID
    SearchTerm string
    Page      int
    Limit     int
}
```

---

## Use Cases

### ✅ Implemented

**Location**: `internal/application/contact_list/`

1. **CreateContactListUseCase** ✅
   - Validates input (name, projectID, tenantID required)
   - Creates ContactList aggregate
   - Adds description if provided
   - Creates and adds filter rules
   - Saves to repository
   - Publishes contact_list.created event
   - Returns ContactListID

2. **UpdateContactListUseCase** ✅
   - Loads existing list
   - Updates name if provided
   - Updates description if provided
   - Updates logical operator if provided
   - Replaces filter rules if provided (clears old, adds new)
   - Saves updated list
   - Publishes contact_list.updated event

3. **ListContactListsUseCase** ✅
   - Queries lists by project with pagination
   - Converts to DTOs with filter rules
   - Returns paginated response with total count

4. **DeleteContactListUseCase** ✅
   - Loads list by ID
   - Calls Delete() method (soft delete)
   - Saves updated list
   - Publishes contact_list.deleted event

### ❌ Suggested Additional Use Cases

```go
// internal/application/contact_list/

// 5. AddContactToStaticListUseCase
//    - Validates list is static
//    - Checks contact exists
//    - Adds contact to list via repository
//    - Publishes contact_list.contact_added event
//    - Recalculates count

// 6. RemoveContactFromStaticListUseCase
//    - Validates list is static
//    - Removes contact from list via repository
//    - Publishes contact_list.contact_removed event
//    - Recalculates count

// 7. GetContactsInListUseCase
//    - Loads list
//    - Gets contact IDs from repository
//    - Loads full contact details
//    - Returns paginated contact list

// 8. RecalculateContactListUseCase
//    - Validates list is dynamic
//    - Recalculates contact count via repository
//    - Updates list with new count
//    - Publishes contact_list.recalculated event

// 9. DuplicateContactListUseCase
//    - Loads existing list
//    - Creates new list with copied filter rules
//    - Appends " (Copy)" to name
//    - Saves new list

// 10. ExportContactsInListUseCase
//     - Gets all contacts in list
//     - Formats as CSV/Excel
//     - Returns file download

// 11. BulkAddContactsToListUseCase
//     - Validates list is static
//     - Validates all contact IDs exist
//     - Adds all contacts in batch
//     - Publishes events for each
//     - Recalculates count

// 12. GetListStatisticsUseCase
//     - Loads list
//     - Calculates growth rate
//     - Calculates engagement metrics
//     - Returns statistics
```

---

## Temporal Workflows (TODO)

**RECOMMENDED**: Background processing for contact list operations

### Suggested Workflows:

1. **RecalculateContactListWorkflow**
   ```go
   // Periodically recalculates dynamic list counts
   //
   // Activities:
   // - FindDynamicListsActivity() - Query all dynamic lists
   // - RecalculateCountActivity() - Recalculate count for each
   // - UpdateListActivity() - Update list with new count
   // - PublishEventActivity() - Publish recalculation event
   //
   // Schedule: Every 15 minutes (configurable)
   // Use Case: Keep dynamic list counts accurate
   ```

2. **BulkImportContactsToListWorkflow**
   ```go
   // Imports large CSV of contacts to static list
   //
   // Activities:
   // - ParseCSVActivity() - Parse and validate CSV
   // - FindOrCreateContactsActivity() - Match or create contacts
   // - AddContactsToListActivity() - Add in batches
   // - PublishEventsActivity() - Publish events in batches
   // - RecalculateCountActivity() - Update list count
   //
   // Handles: Large imports (10k+ contacts), error handling, progress tracking
   ```

3. **ListMigrationWorkflow**
   ```go
   // Migrates contacts from one list to another
   //
   // Activities:
   // - GetSourceContactsActivity() - Get all contacts from source list
   // - AddToTargetListActivity() - Add to target list in batches
   // - RemoveFromSourceActivity() - Optional: remove from source
   // - RecalculateCountsActivity() - Update both list counts
   ```

**Implementation Status**: ❌ NOT IMPLEMENTED

---

## HTTP API (TODO)

### Suggested Endpoints:

```yaml
# ContactList Management
POST   /api/v1/contact-lists                    # Create list
GET    /api/v1/contact-lists                    # List all lists
GET    /api/v1/contact-lists/:id                # Get list details
PUT    /api/v1/contact-lists/:id                # Update list
DELETE /api/v1/contact-lists/:id                # Delete list
POST   /api/v1/contact-lists/:id/duplicate      # Duplicate list
GET    /api/v1/contact-lists/:id/stats          # Get list statistics

# Filter Management
POST   /api/v1/contact-lists/:id/filters        # Add filter rule
DELETE /api/v1/contact-lists/:id/filters/:ruleId # Remove filter rule
PUT    /api/v1/contact-lists/:id/filters        # Replace all filters

# Contact Membership (Static Lists)
GET    /api/v1/contact-lists/:id/contacts       # Get contacts in list
POST   /api/v1/contact-lists/:id/contacts       # Add contact(s) to list
DELETE /api/v1/contact-lists/:id/contacts/:contactId # Remove contact
POST   /api/v1/contact-lists/:id/contacts/bulk  # Bulk add contacts
POST   /api/v1/contact-lists/:id/contacts/import # Import CSV

# Recalculation
POST   /api/v1/contact-lists/:id/recalculate    # Trigger recalculation

# Export
GET    /api/v1/contact-lists/:id/export         # Export contacts as CSV/Excel

# Validation
POST   /api/v1/contact-lists/validate-filters   # Validate filter rules before saving
```

**Implementation Status**: ❌ NOT IMPLEMENTED

---

## Real-World Usage Patterns

### Pattern 1: VIP Customers List (Static)

```go
// Create static list for VIP customers (manually curated)
list, _ := contact_list.NewContactList(
    projectID,
    tenantID,
    "VIP Customers",
    contact_list.LogicalOperatorAND,
    true, // Static
)

// Manually add customers
repo.AddContactToStaticList(ctx, list.ID(), customer1ID)
repo.AddContactToStaticList(ctx, list.ID(), customer2ID)

// Use for exclusive campaigns
broadcast := NewBroadcast("VIP Holiday Offer")
broadcast.SetTargetList(list.ID())
```

### Pattern 2: Qualified Leads List (Dynamic)

```go
// Create dynamic list for qualified leads
list, _ := contact_list.NewContactList(
    projectID,
    tenantID,
    "Qualified Leads",
    contact_list.LogicalOperatorAND, // ALL rules must match
    false, // Dynamic
)

// Rule 1: In Sales Pipeline with "Qualified" status
rule1, _ := contact_list.NewPipelineStatusFilterRule(
    salesPipelineID,
    "Qualified",
    contact_list.OperatorEquals,
)
list.AddFilterRule(rule1)

// Rule 2: Company size > 100 employees
rule2, _ := contact_list.NewCustomFieldFilterRule(
    "company_size",
    shared.FieldTypeNumber,
    contact_list.OperatorGreaterThan,
    100,
)
list.AddFilterRule(rule2)

// Rule 3: Has tag "enterprise"
rule3, _ := contact_list.NewTagFilterRule(
    contact_list.OperatorContains,
    "enterprise",
)
list.AddFilterRule(rule3)

// List automatically updates as contacts match criteria
```

### Pattern 3: Re-engagement List (Dynamic with OR logic)

```go
// Create list for re-engagement campaign
list, _ := contact_list.NewContactList(
    projectID,
    tenantID,
    "Inactive Contacts - Re-engagement",
    contact_list.LogicalOperatorOR, // ANY rule matches
    false, // Dynamic
)

// Rule 1: Last interaction > 30 days ago
rule1, _ := contact_list.NewAttributeFilterRule(
    "last_interaction_at",
    contact_list.OperatorLessThan,
    time.Now().AddDate(0, 0, -30),
)
list.AddFilterRule(rule1)

// Rule 2: Never replied
rule2, _ := contact_list.NewAttributeFilterRule(
    "last_interaction_at",
    contact_list.OperatorIsNull,
    nil,
)
list.AddFilterRule(rule2)
```

### Pattern 4: Multi-Language Segmentation

```go
// Create separate lists for each language
for _, lang := range []string{"en", "pt-BR", "es", "fr"} {
    list, _ := contact_list.NewContactList(
        projectID,
        tenantID,
        fmt.Sprintf("Contacts - %s", lang),
        contact_list.LogicalOperatorAND,
        false, // Dynamic
    )

    rule, _ := contact_list.NewAttributeFilterRule(
        "language",
        contact_list.OperatorEquals,
        lang,
    )
    list.AddFilterRule(rule)

    repo.Create(ctx, list)
}
```

### Pattern 5: Event-Based Segmentation

```go
// Create list of contacts who attended webinar
list, _ := contact_list.NewContactList(
    projectID,
    tenantID,
    "Webinar Attendees - Jan 2025",
    contact_list.LogicalOperatorAND,
    false, // Dynamic
)

// Rule 1: Has tag "webinar-jan-2025"
rule1, _ := contact_list.NewTagFilterRule(
    contact_list.OperatorContains,
    "webinar-jan-2025",
)
list.AddFilterRule(rule1)

// Rule 2: Registered in date range
rule2, _ := contact_list.NewCustomFieldFilterRule(
    "registration_date",
    shared.FieldTypeDate,
    contact_list.OperatorGreaterEqual,
    "2025-01-01",
)
list.AddFilterRule(rule2)
```

---

## Performance Considerations

### Scalability

- **Static Lists**: Fast queries (simple JOIN with membership table)
- **Dynamic Lists**: Query cost depends on filter complexity
- **Contact Count**: Cached and periodically recalculated
- **Filter Evaluation**: Real-time for small lists, background jobs for large lists
- **Pagination**: Always use pagination for large lists (10k+ contacts)

### Database Indexes

```sql
-- ContactList table
CREATE INDEX idx_contact_lists_project_id ON contact_lists(project_id);
CREATE INDEX idx_contact_lists_tenant_id ON contact_lists(tenant_id);
CREATE INDEX idx_contact_lists_deleted_at ON contact_lists(deleted_at);
CREATE INDEX idx_contact_lists_version ON contact_lists(id, version);

-- Filter rules table (in initial schema)
CREATE INDEX idx_contact_list_filters_list_id ON contact_list_filters(contact_list_id);
CREATE INDEX idx_contact_list_filters_type ON contact_list_filters(filter_type);

-- Membership table (static lists)
CREATE INDEX idx_contact_list_members_list_id ON contact_list_members(contact_list_id);
CREATE INDEX idx_contact_list_members_contact_id ON contact_list_members(contact_id);
CREATE UNIQUE INDEX idx_contact_list_members_unique ON contact_list_members(contact_list_id, contact_id);

-- Contacts table (for filter queries)
CREATE INDEX idx_contacts_tags ON contacts USING gin(tags);
CREATE INDEX idx_contacts_language ON contacts(language);
CREATE INDEX idx_contacts_created_at ON contacts(created_at);
CREATE INDEX idx_contacts_last_interaction_at ON contacts(last_interaction_at);
```

### Optimizations

1. **Count Caching**: Cache contact count in ContactList.contactCount
2. **Background Recalculation**: Use Temporal workflow for periodic recalculation
3. **Filter Query Optimization**:
   - Use database indexes for all filtered fields
   - Avoid N+1 queries with proper JOINs
   - Use EXPLAIN ANALYZE to optimize slow queries
4. **Materialized Views**: Consider for very large lists (100k+ contacts)
5. **List Segmentation**: Split very large dynamic lists into smaller sub-lists

### Query Performance Examples

```sql
-- FAST: Tag filter with GIN index
WHERE tags @> '["vip"]'::jsonb

-- FAST: Attribute filter with B-tree index
WHERE language = 'pt-BR' AND created_at > '2025-01-01'

-- MODERATE: Custom field filter (subquery)
WHERE id IN (
    SELECT contact_id FROM contact_custom_fields
    WHERE field_key = 'company_size' AND field_value::int > 100
)

-- SLOW: Multiple OR conditions without proper indexes
WHERE (name ILIKE '%john%' OR email ILIKE '%john%' OR phone ILIKE '%john%')
```

---

## Testing Checklist

### Unit Tests (Domain Layer)

- [x] ContactList creation with valid data
- [x] ContactList creation with invalid data (nil projectID, empty name, etc.)
- [x] Update name (success and validation)
- [x] Update description
- [x] Add filter rule
- [x] Remove filter rule
- [x] Clear all filter rules
- [x] Update logical operator
- [x] Update contact count and lastCalculatedAt
- [x] Soft delete
- [x] Domain events emission (9 events)
- [x] FilterRule creation for all types
- [x] FilterRule validation (operators, null values, etc.)
- [x] LogicalOperator validation
- [x] FilterType validation
- [x] FilterOperator validation

### Integration Tests (Repository Layer)

- [ ] Create list with filter rules
- [ ] Update list with optimistic locking
- [ ] Optimistic locking conflict (version mismatch)
- [ ] Find list by ID with filter rules preloaded
- [ ] Find lists by project
- [ ] Find lists by tenant
- [ ] List by project with pagination
- [ ] Delete list (soft delete)
- [ ] Get contacts in static list
- [ ] Get contacts in dynamic list (tag filter)
- [ ] Get contacts in dynamic list (custom field filter)
- [ ] Get contacts in dynamic list (pipeline status filter)
- [ ] Get contacts in dynamic list (attribute filter)
- [ ] Get contacts with AND logical operator
- [ ] Get contacts with OR logical operator
- [ ] Add contact to static list
- [ ] Remove contact from static list
- [ ] Recalculate contact count
- [ ] Is contact in static list (true)
- [ ] Is contact in static list (false)
- [ ] Is contact in dynamic list (matches rules)
- [ ] Is contact in dynamic list (doesn't match rules)

### End-to-End Tests (Application Layer)

- [ ] Create static list → Add contacts → Verify membership
- [ ] Create dynamic list → Verify contacts match rules
- [ ] Update filter rules → Verify contact list changes
- [ ] Change logical operator → Verify different results
- [ ] Delete list → Verify soft delete and cascading effects
- [ ] Recalculate count → Verify accuracy
- [ ] Export contacts from list
- [ ] Bulk import contacts to static list

---

## Suggested Improvements

### 1. Smart Lists (AI-Powered Segmentation)

```go
type SmartListConfig struct {
    Goal        string  // "high_engagement", "likely_to_convert", "at_risk"
    MinScore    float64 // Minimum prediction score
    MaxContacts int     // Maximum list size
}

// Use AI to predict and segment contacts
// Example: "Contacts likely to convert in next 30 days"
```

### 2. Time-Based Rules

```go
type TimeBasedFilter struct {
    RelativeTime string // "last_7_days", "last_30_days", "this_month"
    FieldName    string
}

// Example: "Last interaction in last 7 days" (auto-updating)
```

### 3. Behavioral Segmentation

```go
type BehaviorFilter struct {
    Action     string // "opened_email", "clicked_link", "replied"
    Frequency  int    // How many times
    TimeWindow string // "last_30_days"
}

// Example: "Opened at least 3 emails in last 30 days"
```

### 4. Exclusion Lists

```go
type ExclusionRule struct {
    ExcludeListID uuid.UUID
    ExcludeType   string // "contacts", "domains", "phone_prefixes"
}

// Example: "All contacts EXCEPT those in 'Unsubscribed' list"
```

### 5. List Snapshots

```go
type ListSnapshot struct {
    ID        uuid.UUID
    ListID    uuid.UUID
    Contacts  []uuid.UUID
    CreatedAt time.Time
    Reason    string // "pre_campaign", "monthly_backup"
}

// Capture point-in-time snapshot of dynamic list
// Useful for campaign attribution and reporting
```

### 6. List Performance Metrics

```go
type ListMetrics struct {
    ListID          uuid.UUID
    TotalContacts   int
    EngagementRate  float64 // Average engagement
    ConversionRate  float64 // Average conversion
    GrowthRate      float64 // Week-over-week growth
    ChurnRate       float64 // Contacts leaving list
}
```

### 7. Nested Filter Groups

```go
type FilterGroup struct {
    Operator LogicalOperator
    Filters  []FilterRule
    Groups   []FilterGroup // Nested groups
}

// Example: (tag="vip" AND language="en") OR (company_size > 500)
```

---

## Related Aggregates

- **Contact**: Primary entity being filtered and grouped
- **CustomField**: Used in FilterTypeCustomField rules
- **Tag**: Used in FilterTypeTag rules
- **Pipeline/PipelineStatus**: Used in FilterTypePipelineStatus rules
- **Broadcast**: Uses ContactLists for targeting
- **Campaign**: Uses ContactLists for enrollment criteria
- **Sequence**: Can be triggered for contacts in specific lists

---

## Industry Comparison

| Feature | Ventros CRM | HubSpot | Mailchimp | ActiveCampaign | Salesforce |
|---------|-------------|---------|-----------|----------------|------------|
| Static Lists | ✅ | ✅ | ✅ | ✅ | ✅ |
| Dynamic Lists | ✅ | ✅ | ✅ | ✅ | ✅ |
| Tag Filtering | ✅ | ✅ | ✅ | ✅ | ✅ |
| Custom Field Filtering | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pipeline Status Filtering | ✅ | ✅ | ❌ | ✅ | ✅ |
| Behavioral Segmentation | ⚠️ | ✅ | ✅ | ✅ | ✅ |
| AI-Powered Lists | ❌ | ✅ | ❌ | ❌ | ✅ |
| List Snapshots | ❌ | ✅ | ✅ | ✅ | ✅ |
| Exclusion Rules | ❌ | ✅ | ✅ | ✅ | ✅ |
| Nested Filter Groups | ❌ | ✅ | ✅ | ✅ | ✅ |
| Real-time Count | ✅ | ✅ | ✅ | ✅ | ✅ |
| List Performance Metrics | ❌ | ✅ | ✅ | ✅ | ✅ |

**Ventros Strengths**:
- Clean DDD architecture
- Type-safe filter rules
- Optimistic locking
- Full event-driven system
- Multi-channel support

**Suggested Additions**:
- Behavioral segmentation (event-based filters)
- AI-powered smart lists
- List snapshots for campaign attribution
- Exclusion rules
- Nested filter groups
- Performance metrics and analytics

---

## Documentation References

- **Domain Events**: `internal/domain/crm/contact_list/events.go`
- **Repository**: `internal/domain/crm/contact_list/repository.go`
- **Filter Rules**: `internal/domain/crm/contact_list/filter_rule.go`
- **Implementation**: `infrastructure/persistence/gorm_contact_list_repository.go`
- **Entities**: `infrastructure/persistence/entities/contact_list.go`
- **Use Cases**: `internal/application/contact_list/`
- **Migration**: `infrastructure/database/migrations/000001_initial_schema.up.sql` (lines 184-197)
- **Optimistic Locking**: `infrastructure/database/migrations/000046_add_optimistic_locking.up.sql` (line 32)

---

## Database Schema

### contact_lists table
```sql
CREATE TABLE contact_lists (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    version            integer NOT NULL DEFAULT 1,
    project_id         uuid NOT NULL,
    tenant_id          text NOT NULL,
    name               text NOT NULL,
    description        text,
    logical_operator   text NOT NULL DEFAULT 'AND',
    is_static          boolean NOT NULL DEFAULT false,
    contact_count      bigint NOT NULL DEFAULT 0,
    last_calculated_at timestamp with time zone,
    created_at         timestamp with time zone,
    updated_at         timestamp with time zone,
    deleted_at         timestamp with time zone
);
```

### contact_list_filters table
```sql
CREATE TABLE contact_list_filters (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_list_id uuid NOT NULL REFERENCES contact_lists(id),
    filter_type     text NOT NULL,
    operator        text NOT NULL,
    field_key       text NOT NULL,
    field_type      text,
    value           text,  -- JSON serialized
    pipeline_id     uuid,
    created_at      timestamp with time zone
);
```

### contact_list_members table
```sql
CREATE TABLE contact_list_members (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_list_id uuid NOT NULL REFERENCES contact_lists(id),
    contact_id      uuid NOT NULL REFERENCES contacts(id),
    added_at        timestamp with time zone,
    UNIQUE(contact_list_id, contact_id)
);
```

---

## Common Questions & Answers

### Q: What's the difference between static and dynamic lists?
**A**: Static lists require manual management (add/remove contacts explicitly). Dynamic lists automatically include contacts matching filter rules in real-time.

### Q: How often are dynamic list counts recalculated?
**A**: Count is cached in `contact_count` field and recalculated:
- On-demand via `RecalculateContactCount()` method
- Automatically when filter rules change
- Periodically via background job (recommended: every 15 minutes)

### Q: Can I combine AND and OR logic?
**A**: Currently, all rules use the same logical operator (all AND or all OR). For complex logic, consider:
1. Creating multiple lists and combining them
2. Using nested filter groups (future enhancement)

### Q: How do I exclude contacts from a list?
**A**: Currently not directly supported. Workarounds:
1. Create inverse rules (e.g., "tag != unsubscribed")
2. Use static list and manually remove contacts
3. Create exclusion list feature (suggested enhancement)

### Q: Can contacts be in multiple lists?
**A**: Yes! Contacts can be in unlimited lists (both static and dynamic).

### Q: What happens when a contact is deleted?
**A**: Contact is soft-deleted, but list memberships remain. When querying list, deleted contacts are automatically excluded.

### Q: How do I target a list in a broadcast?
**A**: Pass the ContactList ID to the Broadcast aggregate, which will query all contacts in the list and send to them.

---

**Last Updated**: 2025-10-12
**Status**: ✅ Domain Complete, ✅ Application Layer Complete, ❌ API Incomplete
**Priority**: HIGH (Critical for marketing automation and segmentation)
**Estimated API Completion**: 3-5 days (REST endpoints + Swagger docs + integration tests)
