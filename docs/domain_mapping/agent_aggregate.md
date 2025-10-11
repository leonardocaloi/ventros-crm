# Agent Aggregate

**Last Updated**: 2025-10-10
**Status**: ✅ Complete and Production-Ready
**Lines of Code**: ~600
**Test Coverage**: Partial

---

## Overview

- **Purpose**: Represents human agents, AI bots, and automation agents
- **Location**: `internal/domain/agent/`
- **Entity**: `infrastructure/persistence/entities/agent.go`
- **Repository**: `infrastructure/persistence/gorm_agent_repository.go`
- **Aggregate Root**: `Agent`

**Business Problem**:
The Agent aggregate represents **individuals or bots** who interact with contacts. This includes human customer service reps, AI chatbots, and automated systems. Critical for:
- **Agent management** - Track team members and their performance
- **Workload distribution** - Assign conversations to available agents
- **Performance monitoring** - Measure response times and quality
- **Permission control** - Control access to features and data
- **AI agents** - Deploy bots alongside human agents

---

## Domain Model

### Aggregate Root: Agent

```go
type Agent struct {
    id          uuid.UUID
    projectID   uuid.UUID
    userID      *uuid.UUID     // Required for human agents, null for bots
    tenantID    string
    name        string
    email       string
    agentType   AgentType      // human, ai, bot, channel
    status      AgentStatus    // available, busy, away, offline
    role        Role           // human_agent, supervisor, admin, bot
    active      bool
    config      map[string]interface{}  // Bot config (API keys, models, etc)
    permissions map[string]bool         // Granular permissions
    settings    map[string]interface{}  // UI preferences

    // Performance metrics
    sessionsHandled   int
    averageResponseMs int
    lastActivityAt    *time.Time

    createdAt   time.Time
    updatedAt   time.Time
    lastLoginAt *time.Time
}
```

### Value Objects

#### 1. AgentType
```go
type AgentType string
const (
    AgentTypeHuman   AgentType = "human"    // Customer service rep
    AgentTypeAI      AgentType = "ai"       // GPT-4, Claude, etc
    AgentTypeBot     AgentType = "bot"      // Rule-based bot
    AgentTypeChannel AgentType = "channel"  // System agent for channel
)
```

#### 2. AgentStatus
```go
type AgentStatus string
const (
    AgentStatusAvailable AgentStatus = "available"  // Ready for new conversations
    AgentStatusBusy      AgentStatus = "busy"       // Handling conversations
    AgentStatusAway      AgentStatus = "away"       // Break/lunch
    AgentStatusOffline   AgentStatus = "offline"    // Not logged in
)
```

#### 3. Role (ai_provider.go)
```go
type Role string
const (
    RoleHumanAgent  Role = "human_agent"   // Standard agent
    RoleSupervisor  Role = "supervisor"    // Can see all conversations
    RoleAdmin       Role = "admin"         // Full access
    RoleBot         Role = "bot"           // Automated responses
)
```

### Business Invariants

1. **Agent must belong to Project**
   - `projectID` and `tenantID` required
   - `name` required

2. **Human agents require User**
   - `AgentTypeHuman` must have `userID`
   - Bots/AI have null `userID`

3. **Status transitions**
   - Only active agents can be available/busy/away
   - Inactive agents are offline

4. **Permissions are additive**
   - Permissions granted explicitly
   - No permissions by default (except role-based)

---

## Events Emitted

| Event | When | Purpose |
|-------|------|---------|
| `agent.created` | New agent | Initialize agent |
| `agent.updated` | Profile modified | Sync UI |
| `agent.activated` | Agent enabled | Allow login |
| `agent.deactivated` | Agent disabled | Block access |
| `agent.logged_in` | Agent logs in | Track activity |
| `agent.permission_granted` | Permission added | Update access |
| `agent.permission_revoked` | Permission removed | Revoke access |

---

## Repository Interface

```go
type Repository interface {
    Save(ctx context.Context, agent *Agent) error
    FindByID(ctx context.Context, id uuid.UUID) (*Agent, error)
    FindByEmail(ctx context.Context, tenantID, email string) (*Agent, error)
    FindByTenant(ctx context.Context, tenantID string) ([]*Agent, error)
    FindActiveByTenant(ctx context.Context, tenantID string) ([]*Agent, error)
    Delete(ctx context.Context, id uuid.UUID) error

    // Advanced queries
    FindByTenantWithFilters(ctx context.Context, filters AgentFilters) ([]*Agent, int64, error)
    SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*Agent, int64, error)
}
```

---

## Commands (CQRS)

### ✅ Implemented

1. **CreateAgentCommand**
2. **UpdateAgentCommand**
3. **ActivateAgentCommand**
4. **DeactivateAgentCommand**
5. **GrantPermissionCommand**

### ❌ Suggested

- **AssignSessionToAgentCommand**
- **TransferSessionCommand**
- **SetAgentStatusCommand** - Change availability
- **RecordAgentPerformanceCommand** - Update metrics
- **GenerateAgentReportCommand** - Daily/weekly reports

---

## Use Cases

### ✅ Implemented

None explicitly (basic CRUD via handlers)

### ❌ Suggested

1. **AssignSessionToAgentUseCase** - Load balancing logic
2. **CalculateAgentPerformanceUseCase** - Generate metrics
3. **FindAvailableAgentUseCase** - Round-robin assignment
4. **DeployAIAgentUseCase** - Setup AI bot with config
5. **GenerateAgentReportUseCase** - Daily performance summary

---

## Agent Types in Detail

### 1. Human Agent (`AgentTypeHuman`)

```go
agent, _ := NewAgent(projectID, tenantID, "John Doe", AgentTypeHuman, &userID)
agent.GrantPermission("view_contacts")
agent.GrantPermission("send_messages")
agent.SetStatus(AgentStatusAvailable)
```

**Use Cases**: Customer service reps, sales agents

### 2. AI Agent (`AgentTypeAI`)

```go
aiAgent, _ := NewAgent(projectID, tenantID, "GPT-4 Assistant", AgentTypeAI, nil)
aiAgent.SetConfig(map[string]interface{}{
    "model": "gpt-4",
    "temperature": 0.7,
    "max_tokens": 1000,
    "api_key": "sk-...",
})
```

**Use Cases**: AI chatbots, auto-responses

### 3. Bot Agent (`AgentTypeBot`)

```go
bot, _ := NewAgent(projectID, tenantID, "Welcome Bot", AgentTypeBot, nil)
bot.SetConfig(map[string]interface{}{
    "rules": []map[string]string{
        {"trigger": "hello", "response": "Hi! How can I help?"},
        {"trigger": "bye", "response": "Goodbye!"},
    },
})
```

**Use Cases**: Rule-based automation, FAQ bots

### 4. Channel Agent (`AgentTypeChannel`)

```go
channelAgent, _ := NewAgent(projectID, tenantID, "WhatsApp Bot", AgentTypeChannel, nil)
channelAgent.SetConfig(map[string]interface{}{
    "channel_id": channelID,
    "auto_response": true,
})
```

**Use Cases**: Channel-specific automation (WhatsApp business bot)

---

## Performance Metrics

### Tracked Metrics

```go
type AgentMetrics struct {
    SessionsHandled   int       // Total conversations
    AverageResponseMs int       // Avg response time
    LastActivityAt    time.Time // Last action
}

// Update on session completion
agent.RecordSessionHandled(responseTimeMs)
```

### Suggested Metrics (Not Implemented)

```go
type AdvancedMetrics struct {
    FirstResponseMs       int     // Time to first reply
    ResolutionTimeMs      int     // Time to resolve
    CustomerSatisfaction  float64 // CSAT score (1-5)
    ConversionsGenerated  int     // Sales closed
    EscalationRate        float64 // % of escalated sessions
}
```

---

## Permission System

### Available Permissions

```go
// View permissions
"view_contacts"
"view_sessions"
"view_messages"
"view_analytics"

// Action permissions
"send_messages"
"create_contacts"
"edit_contacts"
"delete_contacts"
"assign_sessions"
"close_sessions"

// Admin permissions
"manage_agents"
"manage_pipelines"
"manage_automations"
"manage_settings"
```

### Permission Checks

```go
if agent.HasPermission("send_messages") {
    // Allow sending message
}

// Grant permission
agent.GrantPermission("view_analytics")

// Revoke permission
agent.RevokePermission("delete_contacts")
```

---

## Real-World Usage

### Agent Assignment Logic

```
1. New message arrives from contact
2. Check if contact has active session
   - YES: Route to assigned agent
   - NO: Find available agent
3. Find available agent:
   a. Get all active agents with status=available
   b. Sort by sessions_handled (ascending)
   c. Assign to agent with fewest active sessions
4. Update agent status to "busy" if threshold reached
5. Record assignment in session
```

### Agent Status Workflow

```
Agent Login:
→ RecordLogin()
→ SetStatus(AgentStatusAvailable)
→ Show in available agents pool

Agent Handling Session:
→ Session assigned
→ SetStatus(AgentStatusBusy) if > 3 active sessions

Agent Break:
→ SetStatus(AgentStatusAway)
→ No new assignments

Agent Logout:
→ SetStatus(AgentStatusOffline)
→ Transfer active sessions
```

---

## API Examples

### Create Human Agent

```http
POST /api/v1/agents
{
  "name": "John Doe",
  "email": "john@company.com",
  "agent_type": "human",
  "role": "human_agent",
  "user_id": "uuid"
}
```

### Create AI Bot

```http
POST /api/v1/agents
{
  "name": "GPT-4 Assistant",
  "agent_type": "ai",
  "role": "bot",
  "config": {
    "model": "gpt-4",
    "temperature": 0.7,
    "api_key": "sk-..."
  }
}
```

### Grant Permission

```http
POST /api/v1/agents/{id}/permissions
{
  "permission": "view_analytics"
}
```

### Update Status

```http
PUT /api/v1/agents/{id}/status
{
  "status": "available"
}
```

---

## References

- [Agent Domain](../../internal/domain/agent/)
- [Agent Repository](../../infrastructure/persistence/gorm_agent_repository.go)
- [Agent Handler](../../infrastructure/http/handlers/agent_handler.go)
- [AI Provider](../../internal/domain/agent/ai_provider.go)

---

**Next**: [Channel Aggregate](channel_aggregate.md) →
**Previous**: [Pipeline Aggregate](pipeline_aggregate.md) ←

---

## Summary

✅ **Core CRM Aggregates Complete**:
1. Contact - Customer/lead management
2. Session - Conversation tracking
3. Message - Chat messages
4. Pipeline - Sales funnel & automation
5. Agent - Team & bot management

These 5 aggregates form the **foundation** of the Ventros CRM system.
