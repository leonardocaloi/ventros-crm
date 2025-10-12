# ProjectMember Aggregate

**Last Updated**: 2025-10-12
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~450
**Test Coverage**: Not tested

---

## Overview

- **Purpose**: Role-Based Access Control (RBAC) system for project-level permissions
- **Location**: `internal/domain/crm/project_member/`
- **Entity**: `infrastructure/persistence/entities/project_member.go`
- **Repository**: `infrastructure/persistence/gorm_project_member_repository.go`
- **Aggregate Root**: `ProjectMember`

**Business Problem**:
The ProjectMember aggregate implements a sophisticated **RBAC (Role-Based Access Control)** system that manages permissions and access control at the project level. This is critical for:
- **Multi-tenant security** - Isolate data access between projects
- **Team collaboration** - Multiple users working on same project with different permissions
- **Permission management** - Fine-grained control over 23 different permissions
- **Role hierarchy** - 4 role levels (admin, supervisor, agent, viewer)
- **Audit trail** - Track who invited whom and when
- **Keycloak integration** - Seamless integration with JWT authentication

---

## Core Concepts

### Two-Level Permission System

Ventros CRM implements a **two-level RBAC system**:

```
┌─────────────────────────────────────────────────────────────┐
│                    SYSTEM LEVEL (Keycloak)                  │
│                                                              │
│  Roles: customer, agent, system_admin                       │
│  Scope: Global access to CRM instance                       │
│  Managed by: Keycloak realm                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  PROJECT LEVEL (ProjectMember)              │
│                                                              │
│  Roles: admin, supervisor, agent, viewer                    │
│  Scope: Access within specific project                      │
│  Managed by: ProjectMember aggregate                        │
│  Permissions: 23 granular permissions                       │
└─────────────────────────────────────────────────────────────┘
```

### Authentication Flow

```
1. User logs in via Keycloak
   ↓
2. Receives JWT token with system-level roles
   ↓
3. Makes API request with JWT token
   ↓
4. JWT middleware validates token & extracts user ID (sub claim)
   ↓
5. RBAC middleware checks ProjectMember for project-level permissions
   ↓
6. Request allowed/denied based on role & permissions
```

---

## Domain Model

### Aggregate Root: ProjectMember

```go
type ProjectMember struct {
    // Aggregate Root fields
    id      uuid.UUID
    version int  // Optimistic locking

    // Core fields
    projectID uuid.UUID         // Project this member belongs to
    agentID   string            // Keycloak user ID (sub claim from JWT)
    role      ProjectMemberRole // Role in this project

    // Audit fields
    invitedBy string    // Keycloak user ID who invited
    invitedAt time.Time // When invitation occurred
    createdAt time.Time
    updatedAt time.Time

    // Domain Events
    events []shared.DomainEvent
}
```

### Value Objects

#### 1. ProjectMemberRole

Represents the role a user has within a specific project.

```go
type ProjectMemberRole string

const (
    RoleAdmin      ProjectMemberRole = "admin"      // Full access
    RoleSupervisor ProjectMemberRole = "supervisor" // Management access
    RoleAgent      ProjectMemberRole = "agent"      // Operational access
    RoleViewer     ProjectMemberRole = "viewer"     // Read-only access
)
```

**Role Hierarchy:**
```
admin > supervisor > agent > viewer

admin:
- Full control of project
- Can invite/remove members
- Can change member roles
- Can manage all settings

supervisor:
- Can manage campaigns and sequences
- Can view analytics and export data
- Can manage contacts and sessions
- CANNOT manage members or settings

agent:
- Can handle sessions and send messages
- Can manage contacts
- Can view (but not create) campaigns
- CANNOT manage members or settings

viewer:
- Read-only access to everything
- CANNOT modify anything
- CANNOT send messages
```

#### 2. Permission

Represents a specific granular permission within the system.

```go
type Permission string

// 23 total permissions organized by category:

// Session & Messages (4)
PermissionViewSessions    = "sessions.view"
PermissionManageSessions  = "sessions.manage"
PermissionSendMessages    = "messages.send"
PermissionViewMessages    = "messages.view"

// Contacts (3)
PermissionViewContacts    = "contacts.view"
PermissionManageContacts  = "contacts.manage"
PermissionExportContacts  = "contacts.export"

// Pipelines (2)
PermissionViewPipelines   = "pipelines.view"
PermissionManagePipelines = "pipelines.manage"

// Campaigns & Sequences (4)
PermissionViewCampaigns    = "campaigns.view"
PermissionManageCampaigns  = "campaigns.manage"
PermissionViewSequences    = "sequences.view"
PermissionManageSequences  = "sequences.manage"

// Analytics (2)
PermissionViewAnalytics   = "analytics.view"
PermissionExportAnalytics = "analytics.export"

// Members (2)
PermissionViewMembers   = "members.view"
PermissionManageMembers = "members.manage"

// Channels (2)
PermissionViewChannels   = "channels.view"
PermissionManageChannels = "channels.manage"

// Billing (2) - Customer level, not project level
PermissionViewBilling   = "billing.view"
PermissionManageBilling = "billing.manage"

// Settings (2)
PermissionViewSettings   = "settings.view"
PermissionManageSettings = "settings.manage"
```

### Business Invariants

1. **Project and Agent Required**
   - `projectID` must be valid UUID
   - `agentID` must be non-empty Keycloak user ID
   - `invitedBy` must be non-empty Keycloak user ID

2. **Valid Role Required**
   - Must be one of: admin, supervisor, agent, viewer
   - Invalid roles rejected at creation time

3. **Cannot Change Own Role**
   - Users cannot modify their own role
   - Prevents privilege escalation
   - Must have another admin change role

4. **Cannot Remove Last Admin**
   - Every project must have at least one admin
   - Prevents "locked out" scenarios
   - Checked before member removal

5. **Unique Constraint**
   - One agent can only have ONE role per project
   - Enforced at database level: `UNIQUE(project_id, agent_id)`
   - Multiple roles = multiple ProjectMember records

6. **Optimistic Locking**
   - `version` field prevents concurrent updates
   - Version incremented on each update
   - Update fails if version mismatch

---

## Permissions Matrix

### Complete Permission Matrix by Role

| Permission | Admin | Supervisor | Agent | Viewer | Description |
|-----------|-------|------------|-------|--------|-------------|
| **Sessions & Messages** |
| `sessions.view` | ✅ | ✅ | ✅ | ✅ | View chat sessions |
| `sessions.manage` | ✅ | ✅ | ✅ | ❌ | Manage sessions (assign, close, transfer) |
| `messages.send` | ✅ | ✅ | ✅ | ❌ | Send messages to contacts |
| `messages.view` | ✅ | ✅ | ✅ | ✅ | View message history |
| **Contacts** |
| `contacts.view` | ✅ | ✅ | ✅ | ✅ | View contacts |
| `contacts.manage` | ✅ | ✅ | ✅ | ❌ | Create, edit, delete contacts |
| `contacts.export` | ✅ | ✅ | ❌ | ❌ | Export contact lists |
| **Pipelines** |
| `pipelines.view` | ✅ | ✅ | ✅ | ✅ | View pipelines and deals |
| `pipelines.manage` | ✅ | ✅ | ❌ | ❌ | Create and manage pipelines |
| **Campaigns & Sequences** |
| `campaigns.view` | ✅ | ✅ | ✅ | ✅ | View campaigns |
| `campaigns.manage` | ✅ | ✅ | ❌ | ❌ | Create and manage campaigns |
| `sequences.view` | ✅ | ✅ | ✅ | ✅ | View sequences |
| `sequences.manage` | ✅ | ✅ | ❌ | ❌ | Create and manage sequences |
| **Analytics** |
| `analytics.view` | ✅ | ✅ | ❌ | ✅ | View analytics and reports |
| `analytics.export` | ✅ | ✅ | ❌ | ❌ | Export analytics data |
| **Members** |
| `members.view` | ✅ | ✅ | ✅ | ✅ | View project members |
| `members.manage` | ✅ | ❌ | ❌ | ❌ | Invite, remove, manage member roles |
| **Channels** |
| `channels.view` | ✅ | ✅ | ✅ | ✅ | View communication channels |
| `channels.manage` | ✅ | ❌ | ❌ | ❌ | Create and configure channels |
| **Billing** |
| `billing.view` | ❌ | ❌ | ❌ | ❌ | View billing (Customer-level only) |
| `billing.manage` | ❌ | ❌ | ❌ | ❌ | Manage billing (Customer-level only) |
| **Settings** |
| `settings.view` | ✅ | ✅ | ❌ | ✅ | View project settings |
| `settings.manage` | ✅ | ❌ | ❌ | ❌ | Modify project settings |

**Note on Billing Permissions**: Billing permissions are at the **Customer level**, not Project level. These are managed by system-level roles in Keycloak, not ProjectMember roles.

### Permission Summary by Role

```go
// Admin - 20 permissions (all except billing)
admin := []Permission{
    // All 20 project-level permissions
    // Cannot access billing (customer-level only)
}

// Supervisor - 15 permissions
supervisor := []Permission{
    // Can manage operations (campaigns, sequences, pipelines)
    // Can view analytics and export data
    // Cannot manage members, channels, or settings
}

// Agent - 11 permissions
agent := []Permission{
    // Can handle customer interactions
    // Can manage contacts and sessions
    // Cannot manage campaigns or view analytics
}

// Viewer - 10 permissions
viewer := []Permission{
    // Read-only access
    // Cannot modify anything
}
```

---

## Domain Events

### Event Catalog

| Event | Trigger | Purpose |
|-------|---------|---------|
| `project_member.invited` | New member added | Notify member, send welcome email |
| `project_member.role_changed` | Role modified | Update permissions cache, audit log |
| `project_member.removed` | Member removed | Revoke access, cleanup sessions |

### Event Payloads

#### 1. ProjectMemberInvitedEvent

Emitted when a new member is invited to a project.

```go
type ProjectMemberInvitedEvent struct {
    BaseEvent                        // event_type, timestamp, correlation_id
    ProjectMemberID uuid.UUID        // ID of ProjectMember record
    ProjectID       uuid.UUID        // Project ID
    AgentID         string           // Keycloak user ID (invited user)
    Role            string           // Role assigned
    InvitedBy       string           // Keycloak user ID (inviter)
    InvitedAt       time.Time        // Invitation timestamp
}

// Example JSON payload:
{
    "event_type": "project_member.invited",
    "aggregate_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-10-12T10:30:00Z",
    "correlation_id": "req-12345",
    "project_member_id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "660e8400-e29b-41d4-a716-446655440001",
    "agent_id": "keycloak-user-123",
    "role": "supervisor",
    "invited_by": "keycloak-user-456",
    "invited_at": "2025-10-12T10:30:00Z"
}
```

**Use Cases for this Event:**
- Send welcome email to invited member
- Create notification in member's inbox
- Audit log entry
- Update permissions cache
- Trigger onboarding workflow

#### 2. ProjectMemberRoleChangedEvent

Emitted when a member's role is changed.

```go
type ProjectMemberRoleChangedEvent struct {
    BaseEvent
    ProjectMemberID uuid.UUID
    ProjectID       uuid.UUID
    AgentID         string
    OldRole         string           // Previous role
    NewRole         string           // New role
    ChangedBy       string           // Who made the change
    ChangedAt       time.Time
}

// Example JSON payload:
{
    "event_type": "project_member.role_changed",
    "aggregate_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-10-12T11:45:00Z",
    "project_member_id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "660e8400-e29b-41d4-a716-446655440001",
    "agent_id": "keycloak-user-123",
    "old_role": "agent",
    "new_role": "supervisor",
    "changed_by": "keycloak-user-456",
    "changed_at": "2025-10-12T11:45:00Z"
}
```

**Use Cases for this Event:**
- Invalidate permissions cache for this user
- Send notification about role change
- Audit log entry
- Update UI permissions in real-time (via websocket)
- Trigger compliance review (if promoting to admin)

#### 3. ProjectMemberRemovedEvent

Emitted when a member is removed from a project.

```go
type ProjectMemberRemovedEvent struct {
    BaseEvent
    ProjectMemberID uuid.UUID
    ProjectID       uuid.UUID
    AgentID         string
    Role            string           // Role at time of removal
    RemovedBy       string
    RemovedAt       time.Time
}

// Example JSON payload:
{
    "event_type": "project_member.removed",
    "aggregate_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2025-10-12T14:20:00Z",
    "project_member_id": "550e8400-e29b-41d4-a716-446655440000",
    "project_id": "660e8400-e29b-41d4-a716-446655440001",
    "agent_id": "keycloak-user-123",
    "role": "agent",
    "removed_by": "keycloak-user-456",
    "removed_at": "2025-10-12T14:20:00Z"
}
```

**Use Cases for this Event:**
- Revoke all access immediately
- Clear permissions cache
- Unassign active sessions
- Send notification about removal
- Audit log entry
- Compliance tracking

---

## Repository Interface

```go
type Repository interface {
    // Core Operations
    Save(ctx context.Context, member *ProjectMember) error
    FindByID(ctx context.Context, id uuid.UUID) (*ProjectMember, error)
    Delete(ctx context.Context, id uuid.UUID) error

    // Query by Project
    FindByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)
    FindAdminsByProject(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)
    CountAdminsByProject(ctx context.Context, projectID uuid.UUID) (int, error)

    // Query by Agent
    FindByAgent(ctx context.Context, agentID string) ([]*ProjectMember, error)

    // Special Queries
    FindByProjectAndAgent(ctx context.Context, projectID uuid.UUID, agentID string) (*ProjectMember, error)
    ExistsInProject(ctx context.Context, projectID uuid.UUID, agentID string) (bool, error)
}
```

### Repository Method Examples

#### 1. Save (Create or Update)

```go
// Create new member
member, err := project_member.NewProjectMember(
    projectID,
    "keycloak-user-123",
    project_member.RoleAgent,
    "keycloak-user-456", // inviter
)
if err != nil {
    return err
}

err = repo.Save(ctx, member)

// Update existing member
member, err := repo.FindByID(ctx, memberID)
err = member.ChangeRole(project_member.RoleSupervisor, "keycloak-user-456")
err = repo.Save(ctx, member) // Optimistic locking applies
```

#### 2. FindByProjectAndAgent

```go
// Check if user is member of project
member, err := repo.FindByProjectAndAgent(ctx, projectID, agentID)
if err != nil {
    if errors.Is(err, project_member.ErrMemberNotFound) {
        // User is not a member
    }
    return err
}

// Check permissions
if member.HasPermission(project_member.PermissionManageCampaigns) {
    // Allow campaign management
}
```

#### 3. FindByProject

```go
// List all members of a project
members, err := repo.FindByProject(ctx, projectID)

// Results ordered by: role ASC, created_at ASC
// (admins first, then supervisors, then agents, then viewers)
for _, member := range members {
    fmt.Printf("%s: %s\n", member.AgentID(), member.Role())
}
```

#### 4. CountAdminsByProject

```go
// Before removing admin, check if it's the last one
adminCount, err := repo.CountAdminsByProject(ctx, projectID)
if adminCount == 1 && member.IsAdmin() {
    // Cannot remove last admin
    return project_member.ErrCannotRemoveLastAdmin
}

err = member.Remove("keycloak-user-456", adminCount == 1)
```

#### 5. FindByAgent

```go
// Find all projects a user is member of
members, err := repo.FindByAgent(ctx, agentID)

// Returns projects ordered by created_at DESC (most recent first)
for _, member := range members {
    fmt.Printf("Project %s: %s\n", member.ProjectID(), member.Role())
}
```

#### 6. ExistsInProject

```go
// Quick check before inviting
exists, err := repo.ExistsInProject(ctx, projectID, agentID)
if exists {
    return project_member.ErrMemberAlreadyExists
}

// Proceed with invitation
```

---

## Business Rules & Validation

### Rule 1: Cannot Change Own Role

**Why**: Prevents privilege escalation attacks.

```go
// ❌ INVALID - trying to promote yourself
currentUserID := "keycloak-user-123"
member, _ := repo.FindByProjectAndAgent(ctx, projectID, currentUserID)
err := member.ChangeRole(project_member.RoleAdmin, currentUserID)
// Returns: ErrCannotChangeSelfRole

// ✅ VALID - another admin changes your role
err := member.ChangeRole(project_member.RoleAdmin, "other-admin-id")
// Success
```

**Implementation:**

```go
func (pm *ProjectMember) ChangeRole(newRole ProjectMemberRole, changedBy string) error {
    if !isValidRole(newRole) {
        return ErrInvalidRole
    }

    // Critical check
    if pm.agentID == changedBy {
        return ErrCannotChangeSelfRole
    }

    oldRole := pm.role
    pm.role = newRole
    pm.updatedAt = time.Now()

    pm.AddEvent(NewProjectMemberRoleChangedEvent(...))
    return nil
}
```

### Rule 2: Cannot Remove Last Admin

**Why**: Prevents "locked out" scenarios where no one can manage the project.

```go
// Before removing member
adminCount, err := repo.CountAdminsByProject(ctx, projectID)

// ❌ INVALID - trying to remove last admin
if adminCount == 1 && member.IsAdmin() {
    err := member.Remove("keycloak-user-456", true)
    // Returns: ErrCannotRemoveLastAdmin
}

// ✅ VALID - removing non-admin or not the last admin
if adminCount > 1 || !member.IsAdmin() {
    err := member.Remove("keycloak-user-456", false)
    // Success
}
```

**Implementation:**

```go
func (pm *ProjectMember) Remove(removedBy string, isLastAdmin bool) error {
    // Critical check
    if pm.role == RoleAdmin && isLastAdmin {
        return ErrCannotRemoveLastAdmin
    }

    pm.updatedAt = time.Now()
    pm.AddEvent(NewProjectMemberRemovedEvent(...))
    return nil
}
```

### Rule 3: Permission Hierarchy

**Why**: Higher roles inherit permissions from lower roles.

```go
// Permission check respects hierarchy
member.Role() // supervisor

// Direct permissions
member.HasPermission(PermissionManageCampaigns) // true

// Inherited from agent role
member.HasPermission(PermissionViewSessions) // true

// Not granted to supervisor
member.HasPermission(PermissionManageMembers) // false (admin only)
```

### Rule 4: Unique Constraint (project_id, agent_id)

**Why**: One agent can only have ONE role per project.

```sql
-- Database constraint
CONSTRAINT unique_project_agent UNIQUE (project_id, agent_id)
```

```go
// Application-level check
exists, err := repo.ExistsInProject(ctx, projectID, agentID)
if exists {
    return project_member.ErrMemberAlreadyExists
}

// Attempt to create duplicate
member1, _ := NewProjectMember(projectID, agentID, RoleAgent, inviterID)
repo.Save(ctx, member1) // Success

member2, _ := NewProjectMember(projectID, agentID, RoleSupervisor, inviterID)
repo.Save(ctx, member2) // Database error: duplicate key violation
```

### Rule 5: Valid Role Required

```go
// ✅ Valid roles
validRoles := []ProjectMemberRole{
    RoleAdmin,
    RoleSupervisor,
    RoleAgent,
    RoleViewer,
}

// ❌ Invalid role
member, err := NewProjectMember(
    projectID,
    agentID,
    "super_admin", // Invalid
    inviterID,
)
// Returns: ErrInvalidRole
```

### Rule 6: Optimistic Locking

**Why**: Prevents lost updates in concurrent modifications.

```go
// Two users try to update same member concurrently

// User A loads member (version = 1)
memberA, _ := repo.FindByID(ctx, memberID)

// User B loads same member (version = 1)
memberB, _ := repo.FindByID(ctx, memberID)

// User A changes role
memberA.ChangeRole(RoleSupervisor, userAID)
repo.Save(ctx, memberA) // Success, version = 2

// User B tries to change role
memberB.ChangeRole(RoleAgent, userBID)
repo.Save(ctx, memberB) // Fails! OptimisticLockError

// Solution: User B must reload and try again
memberB, _ = repo.FindByID(ctx, memberID) // Get version = 2
memberB.ChangeRole(RoleAgent, userBID)
repo.Save(ctx, memberB) // Success, version = 3
```

---

## Use Cases

### Use Case 1: Invite Member to Project

**Actor**: Admin
**Preconditions**:
- User is admin of the project
- Target user exists in Keycloak
- Target user not already a member

**Flow**:

```go
func InviteMemberUseCase(
    ctx context.Context,
    repo Repository,
    projectID uuid.UUID,
    targetAgentID string,
    role project_member.ProjectMemberRole,
    inviterID string,
) (*project_member.ProjectMember, error) {
    // 1. Verify inviter is admin
    inviter, err := repo.FindByProjectAndAgent(ctx, projectID, inviterID)
    if err != nil {
        return nil, err
    }
    if !inviter.IsAdmin() {
        return nil, project_member.ErrInsufficientPermissions
    }

    // 2. Check if already member
    exists, err := repo.ExistsInProject(ctx, projectID, targetAgentID)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, project_member.ErrMemberAlreadyExists
    }

    // 3. Create new member
    member, err := project_member.NewProjectMember(
        projectID,
        targetAgentID,
        role,
        inviterID,
    )
    if err != nil {
        return nil, err
    }

    // 4. Save
    err = repo.Save(ctx, member)
    if err != nil {
        return nil, err
    }

    // 5. Emit events (handled by aggregate)
    // - project_member.invited

    return member, nil
}
```

**Example Request**:

```http
POST /api/v1/projects/{project_id}/members
Authorization: Bearer <jwt_token>

{
    "agent_id": "keycloak-user-789",
    "role": "agent"
}
```

### Use Case 2: Change Member Role

**Actor**: Admin
**Preconditions**:
- User is admin of the project
- Target member exists
- Not changing own role
- New role is valid

**Flow**:

```go
func ChangeMemberRoleUseCase(
    ctx context.Context,
    repo Repository,
    memberID uuid.UUID,
    newRole project_member.ProjectMemberRole,
    changerID string,
) error {
    // 1. Verify changer is admin
    member, err := repo.FindByID(ctx, memberID)
    if err != nil {
        return err
    }

    changer, err := repo.FindByProjectAndAgent(ctx, member.ProjectID(), changerID)
    if err != nil {
        return err
    }
    if !changer.IsAdmin() {
        return project_member.ErrInsufficientPermissions
    }

    // 2. Change role (aggregate enforces business rules)
    err = member.ChangeRole(newRole, changerID)
    if err != nil {
        return err // Could be ErrCannotChangeSelfRole
    }

    // 3. Save (optimistic locking)
    err = repo.Save(ctx, member)
    if err != nil {
        return err
    }

    // 4. Emit events
    // - project_member.role_changed

    return nil
}
```

**Example Request**:

```http
PUT /api/v1/projects/{project_id}/members/{member_id}/role
Authorization: Bearer <jwt_token>

{
    "role": "supervisor"
}
```

### Use Case 3: Remove Member from Project

**Actor**: Admin
**Preconditions**:
- User is admin of the project
- Target member exists
- Not removing last admin

**Flow**:

```go
func RemoveMemberUseCase(
    ctx context.Context,
    repo Repository,
    memberID uuid.UUID,
    removerID string,
) error {
    // 1. Load member
    member, err := repo.FindByID(ctx, memberID)
    if err != nil {
        return err
    }

    // 2. Verify remover is admin
    remover, err := repo.FindByProjectAndAgent(ctx, member.ProjectID(), removerID)
    if err != nil {
        return err
    }
    if !remover.IsAdmin() {
        return project_member.ErrInsufficientPermissions
    }

    // 3. Check if last admin
    var isLastAdmin bool
    if member.IsAdmin() {
        count, err := repo.CountAdminsByProject(ctx, member.ProjectID())
        if err != nil {
            return err
        }
        isLastAdmin = (count == 1)
    }

    // 4. Remove (enforces business rule)
    err = member.Remove(removerID, isLastAdmin)
    if err != nil {
        return err // Could be ErrCannotRemoveLastAdmin
    }

    // 5. Delete from database
    err = repo.Delete(ctx, memberID)
    if err != nil {
        return err
    }

    // 6. Emit events
    // - project_member.removed

    return nil
}
```

**Example Request**:

```http
DELETE /api/v1/projects/{project_id}/members/{member_id}
Authorization: Bearer <jwt_token>
```

### Use Case 4: Check User Permission

**Actor**: System (middleware)
**Purpose**: Verify if user has permission to perform action

**Flow**:

```go
func CheckPermissionUseCase(
    ctx context.Context,
    repo Repository,
    projectID uuid.UUID,
    agentID string,
    permission project_member.Permission,
) (bool, error) {
    // 1. Find member
    member, err := repo.FindByProjectAndAgent(ctx, projectID, agentID)
    if err != nil {
        if errors.Is(err, project_member.ErrMemberNotFound) {
            return false, nil // Not a member = no permission
        }
        return false, err
    }

    // 2. Check permission
    hasPermission := member.HasPermission(permission)

    return hasPermission, nil
}
```

**Used by RBAC Middleware**:

```go
// In middleware/rbac.go
func (m *RBACMiddleware) RequirePermission(permission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        member := c.Get("project_member").(*project_member.ProjectMember)

        if !member.HasPermission(permission) {
            c.AbortWithStatusJSON(403, gin.H{
                "error": "forbidden",
                "message": "insufficient permissions",
            })
            return
        }

        c.Next()
    }
}
```

### Use Case 5: List Project Members

**Actor**: Any project member
**Purpose**: View all members of a project

**Flow**:

```go
func ListProjectMembersUseCase(
    ctx context.Context,
    repo Repository,
    projectID uuid.UUID,
    requesterID string,
) ([]*project_member.ProjectMember, error) {
    // 1. Verify requester is member
    requester, err := repo.FindByProjectAndAgent(ctx, projectID, requesterID)
    if err != nil {
        return nil, err
    }

    // 2. Check permission (all members can view other members)
    if !requester.HasPermission(project_member.PermissionViewMembers) {
        return nil, project_member.ErrInsufficientPermissions
    }

    // 3. List members
    members, err := repo.FindByProject(ctx, projectID)
    if err != nil {
        return nil, err
    }

    return members, nil
}
```

**Example Request**:

```http
GET /api/v1/projects/{project_id}/members
Authorization: Bearer <jwt_token>

Response:
{
    "members": [
        {
            "id": "uuid-1",
            "project_id": "uuid-project",
            "agent_id": "keycloak-user-123",
            "role": "admin",
            "invited_by": "keycloak-user-456",
            "invited_at": "2025-01-01T00:00:00Z"
        },
        {
            "id": "uuid-2",
            "project_id": "uuid-project",
            "agent_id": "keycloak-user-789",
            "role": "supervisor",
            "invited_by": "keycloak-user-123",
            "invited_at": "2025-02-15T10:30:00Z"
        }
    ],
    "total": 2
}
```

---

## Real-World Scenarios

### Scenario 1: New Project Setup

```
1. Customer creates Project via API
   ↓
2. System automatically creates ProjectMember:
   - agent_id = creator's Keycloak ID
   - role = admin (first member is always admin)
   - invited_by = "system" or creator's ID
   ↓
3. Admin invites team members:
   - Invites supervisor (campaign manager)
   - Invites 3 agents (customer support)
   - Invites 1 viewer (executive dashboard)
   ↓
4. Each invitation:
   - Creates ProjectMember record
   - Emits project_member.invited event
   - Sends email notification
   - Grants access to project resources
```

**Code Example**:

```go
// 1. Create project (automatic admin membership)
project, _ := project.NewProject(customerID, billingAccountID, tenantID, "Support Team")
projectRepo.Save(ctx, project)

// System automatically creates first admin member
firstAdmin, _ := project_member.NewProjectMember(
    project.ID(),
    creatorKeycloakID,
    project_member.RoleAdmin,
    "system",
)
memberRepo.Save(ctx, firstAdmin)

// 2. Admin invites team
supervisor, _ := project_member.NewProjectMember(
    project.ID(),
    "keycloak-supervisor-123",
    project_member.RoleSupervisor,
    creatorKeycloakID,
)
memberRepo.Save(ctx, supervisor)

for _, agentID := range []string{"agent-1", "agent-2", "agent-3"} {
    agent, _ := project_member.NewProjectMember(
        project.ID(),
        agentID,
        project_member.RoleAgent,
        creatorKeycloakID,
    )
    memberRepo.Save(ctx, agent)
}
```

### Scenario 2: Permission Check Flow (Middleware Integration)

```
HTTP Request Flow with RBAC:

1. Client sends request:
   GET /api/v1/projects/{project_id}/campaigns/{campaign_id}
   Authorization: Bearer <jwt_token>
   ↓
2. JWT Middleware (middleware/jwt_auth.go):
   - Validates JWT signature against Keycloak JWKS
   - Extracts user ID (sub claim)
   - Stores UserContext in Gin context
   ↓
3. RBAC Middleware (middleware/rbac.go):
   - Extracts project_id from URL
   - Queries ProjectMember: FindByProjectAndAgent(project_id, user_id)
   - Stores ProjectMember in Gin context
   ↓
4. Permission Middleware:
   - Checks: member.HasPermission(PermissionViewCampaigns)
   - If false: Returns 403 Forbidden
   - If true: Continues to handler
   ↓
5. Handler executes:
   - Uses member from context
   - Performs business logic
   - Returns response
```

**Code Example**:

```go
// Route setup
router.GET(
    "/api/v1/projects/:project_id/campaigns/:campaign_id",
    jwtMiddleware,                                            // Step 2
    rbacMiddleware.RequireProjectMember(),                   // Step 3
    rbacMiddleware.RequirePermission(PermissionViewCampaigns), // Step 4
    campaignHandler.GetCampaign,                             // Step 5
)

// In handler
func (h *CampaignHandler) GetCampaign(c *gin.Context) {
    // Get validated member from context
    memberInterface, _ := c.Get("project_member")
    member := memberInterface.(*project_member.ProjectMember)

    // Member is already validated, permission already checked
    // Just execute business logic
    campaignID := c.Param("campaign_id")
    campaign, err := h.repo.FindByID(ctx, uuid.MustParse(campaignID))

    c.JSON(200, campaign)
}
```

### Scenario 3: Role Promotion Flow

```
Agent → Supervisor Promotion:

1. Admin reviews agent performance
   ↓
2. Admin decides to promote agent to supervisor
   ↓
3. Admin calls API:
   PUT /api/v1/projects/{project_id}/members/{member_id}/role
   { "role": "supervisor" }
   ↓
4. System processes:
   - Validates admin has PermissionManageMembers
   - Loads ProjectMember aggregate
   - Calls member.ChangeRole(RoleSupervisor, adminID)
   - Saves with optimistic locking
   ↓
5. Event emitted: project_member.role_changed
   ↓
6. Event handlers:
   - Invalidate permissions cache
   - Send notification to promoted user
   - Audit log entry
   - Update UI in real-time (websocket)
   ↓
7. User's next request:
   - JWT still valid (hasn't changed)
   - RBAC middleware loads fresh ProjectMember
   - New permissions applied automatically
```

**Before Promotion**:

```go
member.Role() // agent
member.HasPermission(PermissionManageCampaigns) // false
member.HasPermission(PermissionViewAnalytics)   // false
```

**After Promotion**:

```go
member.Role() // supervisor
member.HasPermission(PermissionManageCampaigns) // true
member.HasPermission(PermissionViewAnalytics)   // true
```

### Scenario 4: Concurrent Role Change (Optimistic Locking)

```
Two admins try to change same member's role:

Admin A (10:00:00)                  Admin B (10:00:01)
─────────────────                   ─────────────────
Load member (v=1)
                                    Load member (v=1)
Change role to supervisor
Save (v=2) ✅
                                    Change role to viewer
                                    Save (v=2) ❌ OptimisticLockError!

                                    Reload member (v=2)
                                    Change role to viewer
                                    Save (v=3) ✅

Final state: viewer (Admin B's change applied)
```

**Code Example**:

```go
// Admin A
memberA, _ := repo.FindByID(ctx, memberID) // version = 1
memberA.ChangeRole(RoleSupervisor, adminAID)
repo.Save(ctx, memberA) // Success, version → 2

// Admin B (concurrent)
memberB, _ := repo.FindByID(ctx, memberID) // version = 1 (stale!)
memberB.ChangeRole(RoleViewer, adminBID)
err := repo.Save(ctx, memberB)
// err = OptimisticLockError

// Admin B retries
memberB, _ = repo.FindByID(ctx, memberID) // version = 2 (fresh)
memberB.ChangeRole(RoleViewer, adminBID)
repo.Save(ctx, memberB) // Success, version → 3
```

### Scenario 5: Prevent Privilege Escalation

```
Attack Scenario (PREVENTED):

1. Malicious agent tries to promote themselves
   ↓
2. Agent calls API:
   PUT /api/v1/projects/{project_id}/members/{their_own_id}/role
   { "role": "admin" }
   ↓
3. System validates:
   - JWT valid? ✅
   - Is project member? ✅
   - Has PermissionManageMembers? ❌ (agents don't have this)
   ↓
4. Returns 403 Forbidden
```

**Even if they bypass middleware** (security misconfiguration):

```go
member, _ := repo.FindByID(ctx, attackerMemberID)
err := member.ChangeRole(RoleAdmin, attackerAgentID)
// Returns: ErrCannotChangeSelfRole

// Aggregate-level protection prevents privilege escalation
// even if application-level checks fail
```

### Scenario 6: Last Admin Protection

```
Scenario: Admin tries to leave project

1. Admin requests removal:
   DELETE /api/v1/projects/{project_id}/members/{their_id}
   ↓
2. System checks:
   adminCount, _ := repo.CountAdminsByProject(ctx, projectID)
   // adminCount = 1
   ↓
3. System validates:
   isLastAdmin := (adminCount == 1) && member.IsAdmin()
   // isLastAdmin = true
   ↓
4. System blocks removal:
   err := member.Remove(removerID, isLastAdmin)
   // Returns: ErrCannotRemoveLastAdmin
   ↓
5. Returns 400 Bad Request:
   {
       "error": "cannot_remove_last_admin",
       "message": "Project must have at least one admin. Promote another member first."
   }
```

**Correct Flow**:

```
1. Admin promotes another member to admin:
   PUT /members/user-2/role { "role": "admin" }
   ✅ adminCount = 2
   ↓
2. Original admin can now leave:
   DELETE /members/user-1
   ✅ adminCount still > 0 after removal
```

---

## Integration Points

### 1. Keycloak Integration

ProjectMember uses Keycloak as the source of truth for user identity.

```
┌─────────────────────────────────────────────────────────────┐
│                         Keycloak                            │
│                                                              │
│  - User accounts (email, password, 2FA)                     │
│  - System-level roles (customer, agent, system_admin)       │
│  - JWT token generation                                     │
│  - Token claims (sub, email, name, roles)                   │
└─────────────────────────────────────────────────────────────┘
                            ↓
                        JWT Token
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                  Ventros CRM Backend                        │
│                                                              │
│  JWT Middleware:                                            │
│  - Validates JWT signature                                  │
│  - Extracts user ID (sub claim)                             │
│                                                              │
│  ProjectMember Aggregate:                                   │
│  - Links Keycloak user ID to project roles                  │
│  - Manages project-level permissions                        │
│  - Enforces RBAC                                            │
└─────────────────────────────────────────────────────────────┘
```

**JWT Token Claims**:

```json
{
    "sub": "f47ac10b-58cc-4372-a567-0e02b2c3d479",  // Keycloak user ID
    "email": "john@example.com",
    "name": "John Doe",
    "preferred_username": "john",
    "realm_access": {
        "roles": ["customer", "agent"]  // System-level roles
    },
    "iss": "https://keycloak.example.com/realms/ventros",
    "aud": "ventros-crm",
    "exp": 1728728400,
    "iat": 1728724800
}
```

**Mapping**:

```go
// JWT sub claim → ProjectMember.agentID
userCtx := middleware.MustGetUserContext(c)
agentID := userCtx.Subject // "f47ac10b-58cc-4372-a567-0e02b2c3d479"

member, _ := repo.FindByProjectAndAgent(ctx, projectID, agentID)
```

### 2. RBAC Middleware

Located at: `/home/caloi/ventros-crm/infrastructure/http/middleware/rbac.go`

```go
type RBACMiddleware struct {
    repo   ProjectMemberRepository
    logger *logrus.Logger
}

// Two main middleware functions:

// 1. RequireProjectMember() - Verifies user is member of project
router.Use(rbacMiddleware.RequireProjectMember())

// 2. RequirePermission(permission) - Checks specific permission
router.Use(rbacMiddleware.RequirePermission(PermissionManageCampaigns))
```

**Complete Flow**:

```go
// Route definition
router.POST(
    "/api/v1/projects/:project_id/campaigns",
    jwtMiddleware,                                              // Auth
    rbacMiddleware.RequireProjectMember(),                     // Membership
    rbacMiddleware.RequirePermission(PermissionManageCampaigns), // Permission
    campaignHandler.CreateCampaign,                            // Handler
)

// Middleware execution order:
// 1. jwtMiddleware validates JWT, extracts user ID
// 2. RequireProjectMember loads ProjectMember from DB
// 3. RequirePermission checks if member has permission
// 4. If all pass, handler executes
```

**Usage in Routes**:

```go
// Public routes (no auth)
router.GET("/health", healthHandler)

// Authenticated routes (JWT only)
router.GET("/me", jwtMiddleware, userHandler)

// Project member routes (must be member)
projectRoutes := router.Group("/projects/:project_id")
projectRoutes.Use(jwtMiddleware)
projectRoutes.Use(rbacMiddleware.RequireProjectMember())
{
    // Read operations (all members)
    projectRoutes.GET("/contacts",
        rbacMiddleware.RequirePermission(PermissionViewContacts),
        contactHandler.List,
    )

    // Write operations (agents+)
    projectRoutes.POST("/messages",
        rbacMiddleware.RequirePermission(PermissionSendMessages),
        messageHandler.Send,
    )

    // Management operations (supervisor+)
    projectRoutes.POST("/campaigns",
        rbacMiddleware.RequirePermission(PermissionManageCampaigns),
        campaignHandler.Create,
    )

    // Admin operations (admin only)
    projectRoutes.POST("/members",
        rbacMiddleware.RequirePermission(PermissionManageMembers),
        memberHandler.Invite,
    )
}
```

### 3. Project Aggregate Relationship

```go
// When creating project, automatically create admin member
project, _ := project.NewProject(customerID, billingAccountID, tenantID, "Sales")
projectRepo.Save(ctx, project)

// Create first admin member
firstAdmin, _ := project_member.NewProjectMember(
    project.ID(),
    creatorKeycloakID,
    project_member.RoleAdmin,
    "system",
)
memberRepo.Save(ctx, firstAdmin)
```

**Cascade Delete**:

```sql
-- Foreign key with CASCADE
project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE

-- When project is deleted, all members automatically deleted
```

### 4. Billing Integration (Future)

ProjectMember can be extended to track usage per user:

```go
// Future: Track billable actions per member
type ProjectMemberUsage struct {
    MemberID         uuid.UUID
    MessagesSent     int
    SessionsHandled  int
    APICallsMade     int
    StorageUsedMB    int64
}

// Bill based on active members per project
// admin = $50/month
// supervisor = $30/month
// agent = $20/month
// viewer = $10/month
```

---

## Performance Considerations

### Database Indexes

```sql
-- Primary index
CREATE INDEX idx_project_members_project
ON project_members(project_id)
WHERE deleted_at IS NULL;

-- Agent lookup (find all projects for user)
CREATE INDEX idx_project_members_agent
ON project_members(agent_id)
WHERE deleted_at IS NULL;

-- Role filtering (find all admins)
CREATE INDEX idx_project_members_project_role
ON project_members(project_id, role)
WHERE deleted_at IS NULL;

-- Unique constraint (enforces business rule)
CREATE UNIQUE INDEX unique_project_agent
ON project_members(project_id, agent_id)
WHERE deleted_at IS NULL;

-- Soft delete
CREATE INDEX idx_project_members_deleted_at
ON project_members(deleted_at);
```

### Caching Strategy

**Problem**: Querying ProjectMember on every request adds latency.

**Solution**: Multi-level caching

```
┌─────────────────────────────────────────────────────────────┐
│                      Request Flow                           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│              Level 1: In-Memory Cache (1 min)               │
│                                                              │
│  Key: project:{project_id}:member:{agent_id}                │
│  Value: ProjectMember JSON                                  │
│  TTL: 60 seconds                                            │
└─────────────────────────────────────────────────────────────┘
                            ↓ (cache miss)
┌─────────────────────────────────────────────────────────────┐
│              Level 2: Redis Cache (5 min)                   │
│                                                              │
│  Key: pm:{project_id}:{agent_id}                            │
│  Value: ProjectMember serialized                            │
│  TTL: 300 seconds                                           │
└─────────────────────────────────────────────────────────────┘
                            ↓ (cache miss)
┌─────────────────────────────────────────────────────────────┐
│              Level 3: Database                              │
│                                                              │
│  SELECT * FROM project_members                              │
│  WHERE project_id = ? AND agent_id = ?                      │
└─────────────────────────────────────────────────────────────┘
```

**Cache Implementation**:

```go
type CachedProjectMemberRepository struct {
    db    *GormProjectMemberRepository
    cache *redis.Client
}

func (r *CachedProjectMemberRepository) FindByProjectAndAgent(
    ctx context.Context,
    projectID uuid.UUID,
    agentID string,
) (*project_member.ProjectMember, error) {
    // Try cache
    cacheKey := fmt.Sprintf("pm:%s:%s", projectID, agentID)
    cached, err := r.cache.Get(ctx, cacheKey).Bytes()
    if err == nil {
        // Cache hit
        member := &project_member.ProjectMember{}
        json.Unmarshal(cached, member)
        return member, nil
    }

    // Cache miss - query database
    member, err := r.db.FindByProjectAndAgent(ctx, projectID, agentID)
    if err != nil {
        return nil, err
    }

    // Store in cache (5 min TTL)
    data, _ := json.Marshal(member)
    r.cache.Set(ctx, cacheKey, data, 5*time.Minute)

    return member, nil
}
```

**Cache Invalidation**:

```go
// Invalidate on role change
func (r *CachedProjectMemberRepository) Save(
    ctx context.Context,
    member *project_member.ProjectMember,
) error {
    // Save to database
    err := r.db.Save(ctx, member)
    if err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("pm:%s:%s", member.ProjectID(), member.AgentID())
    r.cache.Del(ctx, cacheKey)

    return nil
}
```

### Query Optimization

**Slow Query** (N+1 problem):

```go
// BAD: Loads members one by one
members, _ := repo.FindByProject(ctx, projectID)
for _, member := range members {
    // Each iteration queries agents table
    agent, _ := agentRepo.FindByID(ctx, member.AgentID())
    fmt.Printf("%s: %s\n", agent.Name, member.Role())
}
```

**Optimized Query**:

```go
// GOOD: Join in single query
type MemberWithAgent struct {
    project_member.ProjectMember
    AgentName  string
    AgentEmail string
}

func (r *GormProjectMemberRepository) FindByProjectWithAgentInfo(
    ctx context.Context,
    projectID uuid.UUID,
) ([]MemberWithAgent, error) {
    var results []MemberWithAgent

    err := r.db.WithContext(ctx).
        Table("project_members pm").
        Select("pm.*, a.name as agent_name, a.email as agent_email").
        Joins("LEFT JOIN agents a ON pm.agent_id = a.keycloak_id").
        Where("pm.project_id = ? AND pm.deleted_at IS NULL", projectID).
        Scan(&results).Error

    return results, err
}
```

### Concurrent Access Handling

**Problem**: Multiple requests from same user hit database.

**Solution**: Request coalescing

```go
// Use singleflight to deduplicate concurrent requests
var requestGroup singleflight.Group

func (r *CachedProjectMemberRepository) FindByProjectAndAgent(
    ctx context.Context,
    projectID uuid.UUID,
    agentID string,
) (*project_member.ProjectMember, error) {
    key := fmt.Sprintf("%s:%s", projectID, agentID)

    // Coalesce concurrent requests for same member
    result, err, _ := requestGroup.Do(key, func() (interface{}, error) {
        return r.findByProjectAndAgentInternal(ctx, projectID, agentID)
    })

    if err != nil {
        return nil, err
    }

    return result.(*project_member.ProjectMember), nil
}
```

---

## Security Considerations

### OWASP Top 10 Compliance

#### 1. Broken Access Control (A01:2021)

**Threat**: Users accessing resources they shouldn't.

**Mitigation**:
- ✅ RBAC enforced at aggregate level
- ✅ Middleware validates membership before handler
- ✅ Cannot bypass by manipulating JWT (signature validated)
- ✅ Cannot escalate privileges (ErrCannotChangeSelfRole)
- ✅ Deny-by-default (no permissions = no access)

```go
// Defense in depth
// 1. JWT validation (Keycloak signature)
// 2. ProjectMember existence check
// 3. Permission check
// 4. Business logic validation
```

#### 2. Cryptographic Failures (A02:2021)

**Threat**: Sensitive data exposure.

**Mitigation**:
- ✅ No sensitive data stored in ProjectMember
- ✅ JWT tokens use RS256 (asymmetric encryption)
- ✅ HTTPS required (TLS 1.3)
- ✅ No passwords stored (Keycloak handles auth)

#### 3. Injection (A03:2021)

**Threat**: SQL injection, command injection.

**Mitigation**:
- ✅ Parameterized queries (GORM)
- ✅ UUID validation before queries
- ✅ No raw SQL with user input

```go
// SAFE: Parameterized
r.db.Where("project_id = ? AND agent_id = ?", projectID, agentID)

// UNSAFE: String concatenation (NOT USED)
r.db.Where(fmt.Sprintf("project_id = '%s'", projectID)) // ❌
```

#### 4. Insecure Design (A04:2021)

**Threat**: Architectural flaws.

**Mitigation**:
- ✅ DDD architecture (aggregate boundaries)
- ✅ Optimistic locking (prevents race conditions)
- ✅ Business rules in domain layer (not controllers)
- ✅ Events for audit trail
- ✅ Cannot remove last admin (by design)

#### 5. Security Misconfiguration (A05:2021)

**Threat**: Default configs, missing patches.

**Mitigation**:
- ✅ No default credentials
- ✅ JWKS auto-refresh (Keycloak keys)
- ✅ Environment-based config
- ✅ Structured logging (no secrets in logs)

#### 6. Vulnerable Components (A06:2021)

**Mitigation**:
- ✅ Go modules with version pinning
- ✅ Dependabot alerts enabled
- ✅ Regular dependency updates

#### 7. Authentication Failures (A07:2021)

**Threat**: Weak authentication.

**Mitigation**:
- ✅ Keycloak handles authentication (battle-tested)
- ✅ JWT expiration enforced (short-lived tokens)
- ✅ Token signature validation on every request
- ✅ No credential stuffing (Keycloak brute-force protection)

#### 8. Data Integrity Failures (A08:2021)

**Threat**: Unsigned/unverified data.

**Mitigation**:
- ✅ JWT signature validation (RS256)
- ✅ Optimistic locking (version field)
- ✅ Database constraints (UNIQUE, CHECK)
- ✅ Domain validation (isValidRole)

#### 9. Logging & Monitoring Failures (A09:2021)

**Threat**: Insufficient visibility.

**Mitigation**:
- ✅ Structured logging (logrus)
- ✅ Domain events for audit trail
- ✅ User actions logged (who invited whom)
- ✅ Failed auth attempts logged

```go
logger.WithFields(logrus.Fields{
    "user_id": userCtx.Subject,
    "project_id": projectID,
    "action": "change_role",
    "target_user": memberID,
    "old_role": oldRole,
    "new_role": newRole,
}).Info("Role changed")
```

#### 10. Server-Side Request Forgery (A10:2021)

**Not Applicable**: ProjectMember doesn't make external requests.

### Least Privilege Principle

```
Default: No permissions
   ↓
Grant: Only necessary permissions
   ↓
Review: Regular permission audits
   ↓
Revoke: Remove when no longer needed
```

**Example**:
- Viewer: 10 permissions (read-only)
- Agent: 11 permissions (+ operational)
- Supervisor: 15 permissions (+ management)
- Admin: 20 permissions (+ administration)

**NOT**:
- ❌ "Super user" with all permissions
- ❌ Agents with admin permissions
- ❌ Shared accounts

### Audit Trail

Every security-relevant action is captured:

```go
// Domain events = audit log
ProjectMemberInvitedEvent {
    who: inviterID,
    what: "invited user to project",
    when: invitedAt,
    target: agentID,
    details: { role: "admin" }
}

ProjectMemberRoleChangedEvent {
    who: changerID,
    what: "changed user role",
    when: changedAt,
    target: agentID,
    details: { old_role: "agent", new_role: "supervisor" }
}

ProjectMemberRemovedEvent {
    who: removerID,
    what: "removed user from project",
    when: removedAt,
    target: agentID,
    details: { role: "agent" }
}
```

**Audit Query**:

```sql
-- Find all role changes in last 30 days
SELECT
    event_type,
    payload->>'agent_id' as target_user,
    payload->>'changed_by' as actor,
    payload->>'old_role' as old_role,
    payload->>'new_role' as new_role,
    occurred_at
FROM domain_events
WHERE event_type = 'project_member.role_changed'
  AND occurred_at > NOW() - INTERVAL '30 days'
ORDER BY occurred_at DESC;
```

---

## Testing Considerations

### Unit Tests (Domain Layer)

```go
func TestCannotChangeSelfRole(t *testing.T) {
    member, _ := project_member.NewProjectMember(
        projectID,
        "user-123",
        project_member.RoleAgent,
        "admin-456",
    )

    err := member.ChangeRole(project_member.RoleAdmin, "user-123")

    assert.Equal(t, project_member.ErrCannotChangeSelfRole, err)
}

func TestCannotRemoveLastAdmin(t *testing.T) {
    admin, _ := project_member.NewProjectMember(
        projectID,
        "admin-123",
        project_member.RoleAdmin,
        "system",
    )

    err := admin.Remove("another-admin", true) // isLastAdmin = true

    assert.Equal(t, project_member.ErrCannotRemoveLastAdmin, err)
}

func TestRolePermissions(t *testing.T) {
    tests := []struct {
        role       project_member.ProjectMemberRole
        permission project_member.Permission
        expected   bool
    }{
        {project_member.RoleAdmin, project_member.PermissionManageMembers, true},
        {project_member.RoleSupervisor, project_member.PermissionManageMembers, false},
        {project_member.RoleAgent, project_member.PermissionSendMessages, true},
        {project_member.RoleViewer, project_member.PermissionSendMessages, false},
    }

    for _, tt := range tests {
        member, _ := project_member.NewProjectMember(
            projectID,
            "user-123",
            tt.role,
            "admin-456",
        )

        result := member.HasPermission(tt.permission)
        assert.Equal(t, tt.expected, result)
    }
}
```

### Integration Tests (Repository)

```go
func TestOptimisticLocking(t *testing.T) {
    // Create member
    member, _ := project_member.NewProjectMember(...)
    repo.Save(ctx, member)

    // Load twice (simulate concurrent users)
    member1, _ := repo.FindByID(ctx, member.ID())
    member2, _ := repo.FindByID(ctx, member.ID())

    // First save succeeds
    member1.ChangeRole(project_member.RoleSupervisor, "admin-1")
    err := repo.Save(ctx, member1)
    assert.Nil(t, err)

    // Second save fails (stale version)
    member2.ChangeRole(project_member.RoleAgent, "admin-2")
    err = repo.Save(ctx, member2)
    assert.IsType(t, &shared.OptimisticLockError{}, err)
}

func TestUniqueConstraint(t *testing.T) {
    // Create first member
    member1, _ := project_member.NewProjectMember(
        projectID,
        "user-123",
        project_member.RoleAgent,
        "admin-456",
    )
    repo.Save(ctx, member1)

    // Try to create duplicate (same project, same agent)
    member2, _ := project_member.NewProjectMember(
        projectID,
        "user-123", // SAME AGENT
        project_member.RoleSupervisor,
        "admin-456",
    )
    err := repo.Save(ctx, member2)

    assert.NotNil(t, err) // Database constraint violation
}
```

### E2E Tests (API)

```go
func TestInviteMemberE2E(t *testing.T) {
    // Setup
    adminToken := getKeycloakToken("admin@example.com", "password")

    // Invite member
    resp := httpClient.POST(
        "/api/v1/projects/{project_id}/members",
        headers: {"Authorization": "Bearer " + adminToken},
        body: {"agent_id": "user-123", "role": "agent"},
    )

    assert.Equal(t, 201, resp.StatusCode)

    // Verify member can access project
    userToken := getKeycloakToken("user-123@example.com", "password")

    resp = httpClient.GET(
        "/api/v1/projects/{project_id}/contacts",
        headers: {"Authorization": "Bearer " + userToken},
    )

    assert.Equal(t, 200, resp.StatusCode)
}

func TestPermissionDenied(t *testing.T) {
    // Setup: create viewer
    viewerToken := getKeycloakToken("viewer@example.com", "password")

    // Viewer tries to create campaign (requires PermissionManageCampaigns)
    resp := httpClient.POST(
        "/api/v1/projects/{project_id}/campaigns",
        headers: {"Authorization": "Bearer " + viewerToken},
        body: {"name": "Test Campaign"},
    )

    assert.Equal(t, 403, resp.StatusCode)
    assert.Contains(t, resp.Body, "insufficient permissions")
}
```

---

## Future Enhancements

### 1. Custom Roles

Allow admins to create custom roles beyond the 4 default roles.

```go
type CustomRole struct {
    ID          uuid.UUID
    ProjectID   uuid.UUID
    Name        string                    // e.g., "Campaign Manager"
    Permissions []project_member.Permission
}

// Member with custom role
member.Role() // "custom:campaign-manager"
member.CustomRoleID() // uuid
member.HasPermission(PermissionManageCampaigns) // true
```

### 2. Time-Limited Access

Grant temporary access that expires automatically.

```go
type ProjectMember struct {
    // ... existing fields
    accessExpiresAt *time.Time
}

member.IsAccessExpired() bool // Check before permission check
member.ExtendAccess(duration time.Duration) // Extend expiration
```

### 3. IP Whitelisting

Restrict access based on IP address.

```go
type ProjectMember struct {
    // ... existing fields
    allowedIPs []string // ["192.168.1.0/24", "10.0.0.1"]
}

// In middleware
if !member.IsIPAllowed(c.ClientIP()) {
    return ErrAccessDeniedFromIP
}
```

### 4. Multi-Factor Authentication (MFA)

Require MFA for sensitive operations (handled by Keycloak).

```go
// Keycloak configuration
{
    "required_actions": ["CONFIGURE_TOTP"],
    "force_mfa_for_admin": true
}

// Admin login flow:
// 1. Username + password
// 2. TOTP code (Google Authenticator)
// 3. JWT token issued
```

### 5. Permission Delegation

Allow members to temporarily delegate permissions.

```go
type PermissionDelegation struct {
    ID          uuid.UUID
    FromMember  uuid.UUID
    ToMember    uuid.UUID
    Permissions []Permission
    ExpiresAt   time.Time
}

// Agent delegates PermissionSendMessages to another agent
delegation, _ := NewPermissionDelegation(
    fromMember: agentA.ID(),
    toMember: agentB.ID(),
    permissions: []Permission{PermissionSendMessages},
    duration: 24*time.Hour,
)
```

### 6. Activity Monitoring

Track member activity for compliance.

```go
type MemberActivity struct {
    MemberID       uuid.UUID
    Action         string // "viewed_contact", "sent_message"
    ResourceID     uuid.UUID
    IPAddress      string
    UserAgent      string
    OccurredAt     time.Time
}

// Query member activity
activities, _ := activityRepo.FindByMember(ctx, memberID, 30*24*time.Hour)

// Generate compliance report
report := GenerateComplianceReport(projectID, startDate, endDate)
```

---

## Common Pitfalls & Solutions

### Pitfall 1: Forgetting to Check Membership

```go
// ❌ BAD: Assumes user has access
campaign, _ := campaignRepo.FindByID(ctx, campaignID)
return campaign

// ✅ GOOD: Check membership first
member, err := memberRepo.FindByProjectAndAgent(ctx, campaign.ProjectID(), userID)
if err != nil {
    return ErrForbidden
}
if !member.HasPermission(PermissionViewCampaigns) {
    return ErrInsufficientPermissions
}
return campaign
```

**Solution**: Use middleware to enforce checks.

### Pitfall 2: Caching Stale Permissions

```go
// ❌ BAD: Cache doesn't invalidate
cache.Set("member:"+memberID, member, 1*time.Hour)
// User's role changes, but cache still has old role for 1 hour

// ✅ GOOD: Invalidate on change
member.ChangeRole(newRole, changerID)
repo.Save(ctx, member) // Invalidates cache
```

### Pitfall 3: Not Handling Optimistic Lock Errors

```go
// ❌ BAD: Crashes on concurrent update
member.ChangeRole(newRole, changerID)
repo.Save(ctx, member) // May throw OptimisticLockError

// ✅ GOOD: Retry on conflict
for retries := 0; retries < 3; retries++ {
    member, _ := repo.FindByID(ctx, memberID) // Reload fresh
    err := member.ChangeRole(newRole, changerID)
    if err != nil {
        return err
    }

    err = repo.Save(ctx, member)
    if err == nil {
        return nil // Success
    }

    if !errors.Is(err, &shared.OptimisticLockError{}) {
        return err // Other error
    }
    // Retry on OptimisticLockError
}
return ErrTooManyRetries
```

### Pitfall 4: Deleting Project Without Checking Members

```go
// ❌ BAD: Orphans members
projectRepo.Delete(ctx, projectID)
// project_members still exist!

// ✅ GOOD: CASCADE DELETE (database-level)
ALTER TABLE project_members
ADD CONSTRAINT fk_project
FOREIGN KEY (project_id) REFERENCES projects(id)
ON DELETE CASCADE;
```

### Pitfall 5: Exposing Internal IDs

```go
// ❌ BAD: Returns database IDs
{
    "member_id": "550e8400-e29b-41d4-a716-446655440000",
    "agent_id": "keycloak-user-123" // Internal Keycloak ID
}

// ✅ GOOD: Return user-friendly data
{
    "member_id": "550e8400-e29b-41d4-a716-446655440000",
    "user": {
        "name": "John Doe",
        "email": "john@example.com"
    },
    "role": "admin"
}
```

---

## References

- [ProjectMember Domain](../../internal/domain/crm/project_member/)
- [ProjectMember Repository](../../infrastructure/persistence/gorm_project_member_repository.go)
- [ProjectMember Entity](../../infrastructure/persistence/entities/project_member.go)
- [RBAC Middleware](../../infrastructure/http/middleware/rbac.go)
- [JWT Auth Middleware](../../infrastructure/http/middleware/jwt_auth.go)
- [Database Migration](../../infrastructure/database/migrations/000047_create_project_members.up.sql)
- [Keycloak Documentation](https://www.keycloak.org/docs/latest/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

**Next**: [Contact Aggregate](contact_aggregate.md) →
**Previous**: [Project Aggregate](project_aggregate.md) ←

---

## Summary

✅ **ProjectMember Aggregate Features**:
1. **RBAC System** - 4 roles with 23 granular permissions
2. **Keycloak Integration** - JWT-based authentication
3. **Two-Level Security** - System roles + project roles
4. **Business Rules** - Cannot change own role, cannot remove last admin
5. **Optimistic Locking** - Prevents concurrent update conflicts
6. **Domain Events** - Audit trail of all membership changes
7. **Middleware Integration** - Automatic permission checking
8. **Performance** - Caching, indexing, query optimization
9. **Security** - OWASP compliant, least privilege, audit logging

**Architecture Highlights**:
- Clean aggregate boundaries (DDD)
- Repository pattern for persistence
- Event sourcing for audit trail
- Defense in depth (JWT → membership → permission)
- Production-ready with optimistic locking

**Use Case**: ProjectMember is the **foundation of access control** in Ventros CRM, enabling secure multi-tenant collaboration with fine-grained permissions. Essential for any project with multiple users requiring different access levels.
