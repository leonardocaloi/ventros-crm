# Ventros CRM - Technical Documentation

Complete technical reference for developers.

---

## 📌 Important Notes

| **Topic** | **Details** |
|-----------|-------------|
| **Architecture** | DDD + Clean Architecture + Event-Driven (119 events) |
| **Test Coverage** | 82% (Domain: 92%, Application: 84%, Infrastructure: 71%) |
| **Event Latency** | <100ms (PostgreSQL LISTEN/NOTIFY, NO polling) |
| **Multi-tenancy** | RLS (Row-Level Security) in PostgreSQL |
| **Migrations** | Automatic on startup (golang-migrate + SQL files) |
| **API Docs** | Swagger UI at `/swagger/index.html` |
| **Main Test** | `make setup-all-complete` - Full E2E test |

### Naming Conventions

| **Pattern** | **Usage** | **Example** |
|-------------|-----------|-------------|
| `New*` | Constructors | `NewContact`, `NewSession` |
| `Reconstruct*` | Load from DB | `ReconstructContact` |
| `Update*`, `Set*` | Mutations | `UpdateName`, `SetEmail` |
| `Add*`, `Remove*` | Collections | `AddTag`, `RemoveTag` |
| `Mark*` | Status changes | `MarkAsDelivered` |
| No `Get` prefix | Getters (idiomatic Go) | `contact.Email()` not `contact.GetEmail()` |

### Implementation Status

| **Component** | **Status** | **Notes** |
|---------------|------------|-----------|
| Contact, Session, Message | ✅ Complete | Full CRUD + events |
| Pipeline, Agent, Channel | ✅ Complete | With automation |
| WAHA Integration | ✅ Complete | All message types |
| Multi-tenancy + RLS | ✅ Complete | PostgreSQL policies |
| Temporal Workflows | ✅ Complete | Session timeouts |
| AI Processing | 🚧 Partial | Events defined, processors pending |
| Meta Ads API | 🚧 Partial | Conversion tracking partial |

---

## 📚 Contents

1. [Domain Layer](#domain-layer)
2. [Application Layer](#application-layer)
3. [Infrastructure Layer](#infrastructure-layer)
4. [API Endpoints](#api-endpoints)
5. [Events](#events)
6. [Workflows](#workflows)

---

## 🏗️ Domain Layer

### Core Aggregates

#### **Contact** (`internal/domain/contact/`)

Customer/lead representation.

**Key Methods**:
```go
NewContact(projectID, tenantID, name string) (*Contact, error)
SetEmail(email string) error
SetPhone(phone string) error
AddTag(tag string)
RemoveTag(tag string)
SetLanguage(language string)
SetTimezone(timezone string)
RecordInteraction()
```

**Events**: `contact.created`, `contact.updated`, `contact.tag_added`, `contact.email_set`, `contact.phone_set`, `contact.pipeline_status_changed`

**Value Objects**: `Email`, `Phone`

---

#### **Session** (`internal/domain/session/`)

Conversation window between contact and team.

**Key Methods**:
```go
NewSession(contactID, tenantID, channelTypeID uuid.UUID, timeout int) (*Session, error)
RecordMessage(fromContact bool, agentID *uuid.UUID) error
AssignAgent(agentID uuid.UUID) error
CheckTimeout() bool
End(reason EndReason) error
```

**Auto-calculated Metrics**:
- `MessageCount`, `MessagesFromContact`, `MessagesFromAgent`
- `AgentResponseTimeSeconds`, `ContactWaitTimeSeconds`
- `DurationSeconds`

**Events**: `session.started`, `session.ended`, `session.message_recorded`, `session.agent_assigned`, `session.abandoned`

---

#### **Message** (`internal/domain/message/`)

Individual message (text, image, audio, video, etc).

**Key Methods**:
```go
NewMessage(contactID, projectID, channelTypeID uuid.UUID, contentType ContentType, fromMe bool) (*Message, error)
SetText(text string) error
SetMediaContent(url, mimetype string) error
MarkAsDelivered()
MarkAsRead()
MarkAsFailed(reason string)
```

**Content Types**: `text`, `image`, `video`, `audio`, `voice`, `document`, `location`, `contact`, `sticker`

**Events**: `message.created`, `message.delivered`, `message.read`, `message.failed`, `message.ai.process_*_requested`

---

#### **Pipeline** (`internal/domain/pipeline/`)

Sales/support funnel with statuses.

**Key Methods**:
```go
NewPipeline(projectID, tenantID uuid.UUID, name string) (*Pipeline, error)
AddStatus(name string, statusType StatusType) (*Status, error)
RemoveStatus(statusID uuid.UUID) error
SetSessionTimeout(minutes int)
```

**Status Types**: `open`, `won`, `lost`

**Events**: `pipeline.created`, `status.created`, `pipeline.status_added`, `contact.entered_pipeline`

---

#### **Automation** (`internal/domain/pipeline/automation_rule.go`)

Trigger-based automation system.

**Triggers** (20+):
- Session: `session.ended`, `session.timeout`, `session.abandoned`
- Message: `message.received`, `no_response.timeout`
- Pipeline: `status.changed`, `contact.entered_pipeline`
- Temporal: `after.delay`, `scheduled`

**Actions** (15+):
- Messaging: `send_message`, `send_template`, `send_email`
- Pipeline: `change_pipeline_status`, `assign_agent`
- Organization: `add_tag`, `remove_tag`, `update_custom_field`
- Integration: `send_webhook`, `trigger_workflow`

---

### Other Aggregates

| Aggregate | Location | Purpose |
|-----------|----------|---------|
| **Project** | `internal/domain/project/` | Workspace/multi-project |
| **Agent** | `internal/domain/agent/` | Human/AI/Bot agents |
| **Channel** | `internal/domain/channel/` | WhatsApp, Instagram, etc |
| **Chat** | `internal/domain/chat/` | Group chats |
| **Note** | `internal/domain/note/` | Internal notes |
| **Tracking** | `internal/domain/tracking/` | UTM, ads attribution |
| **Webhook** | `internal/domain/webhook/` | Webhook subscriptions |

---

## 🎯 Application Layer

### Use Cases

| Use Case | Location | Purpose |
|----------|----------|---------|
| `CreateContactUseCase` | `internal/application/contact/` | Create contact |
| `ChangePipelineStatusUseCase` | `internal/application/contact/` | Move contact in pipeline |
| `ProcessInboundMessageUseCase` | `internal/application/message/` | Process received messages |
| `SendMessageCommand` | `internal/application/commands/message/` | Send message |
| `CreateSessionUseCase` | `internal/application/session/` | Start conversation |
| `CloseSessionUseCase` | `internal/application/session/` | End conversation |

---

## 🔧 Infrastructure Layer

### HTTP Handlers (`infrastructure/http/handlers/`)

| Handler | Endpoints | Purpose |
|---------|-----------|---------|
| **AuthHandler** | `/api/v1/auth/*` | Authentication |
| **ContactHandler** | `/api/v1/crm/contacts` | Contact CRUD |
| **SessionHandler** | `/api/v1/crm/sessions` | Session management |
| **MessageHandler** | `/api/v1/crm/messages` | Message CRUD + send |
| **ChannelHandler** | `/api/v1/crm/channels` | Channel management |
| **PipelineHandler** | `/api/v1/crm/pipelines` | Pipeline CRUD |
| **WAHAWebhookHandler** | `/api/v1/webhooks/waha/:session` | Receive WAHA webhooks |

### Repositories (`infrastructure/persistence/`)

All aggregates have GORM repositories: `GormContactRepository`, `GormSessionRepository`, `GormMessageRepository`, etc.

### Integrations

| Integration | Location | Purpose |
|-------------|----------|---------|
| **WAHA** | `infrastructure/channels/waha/` | WhatsApp HTTP API |
| **RabbitMQ** | `infrastructure/messaging/` | Event bus |
| **Temporal** | `infrastructure/workflow/` | Workflows |
| **Redis** | `infrastructure/cache/` | Cache + WebSocket |

---

## 🌐 API Endpoints

### Authentication

```http
POST   /api/v1/auth/register       # Create user
POST   /api/v1/auth/login          # Login
GET    /api/v1/auth/profile        # Get profile (protected)
POST   /api/v1/auth/api-key        # Generate API key (protected)
```

### Contacts (Protected)

```http
GET    /api/v1/crm/contacts        # List contacts
POST   /api/v1/crm/contacts        # Create contact
GET    /api/v1/crm/contacts/:id    # Get contact
PATCH  /api/v1/crm/contacts/:id    # Update contact
DELETE /api/v1/crm/contacts/:id    # Delete contact
PATCH  /api/v1/crm/contacts/:id/pipeline-status  # Change pipeline status
```

### Sessions (Protected)

```http
GET    /api/v1/crm/sessions        # List sessions
GET    /api/v1/crm/sessions/:id    # Get session
POST   /api/v1/crm/sessions/:id/close  # Close session
```

### Messages (Protected)

```http
GET    /api/v1/crm/messages        # List messages
POST   /api/v1/crm/messages        # Send message
GET    /api/v1/crm/messages/:id    # Get message
POST   /api/v1/crm/messages/:id/mark-delivered  # Mark delivered
```

**Send Message Example**:
```json
POST /api/v1/crm/messages
{
  "contact_id": "uuid",
  "type": "text",
  "content": "Hello! How can I help you?",
  "reply_to_message_id": "uuid"
}
```

### Channels (Protected)

```http
GET    /api/v1/crm/channels             # List channels
POST   /api/v1/crm/channels             # Create channel
GET    /api/v1/crm/channels/:id         # Get channel
POST   /api/v1/crm/channels/:id/qrcode  # Get WhatsApp QR code
POST   /api/v1/crm/channels/:id/import  # Import message history
```

### Pipelines (Protected)

```http
GET    /api/v1/crm/pipelines                 # List pipelines
POST   /api/v1/crm/pipelines                 # Create pipeline
GET    /api/v1/crm/pipelines/:id             # Get pipeline
POST   /api/v1/crm/pipelines/:id/statuses    # Add status
```

### Webhooks (Public)

```http
POST   /api/v1/webhooks/waha/:session    # Receive WAHA webhook
```

### Health (Public)

```http
GET    /health         # Overall health
GET    /ready          # Readiness probe
GET    /live           # Liveness probe
```

---

## 📡 Events

### Event Categories

| Category | Count | Examples |
|----------|-------|----------|
| **Contact** | 22 | `contact.created`, `contact.pipeline_status_changed`, `tracking.message.meta_ads` |
| **Message** | 9 | `message.created`, `message.ai.process_image_requested` |
| **Session** | 9 | `session.started`, `session.ended`, `session.abandoned` |
| **Pipeline** | 24 | `automation_rule.triggered`, `contact.lead_qualified` |
| **Channel** | 8 | `channel.created`, `channel.pipeline.associated` |
| **Agent** | 7 | `agent.created`, `agent.logged_in` |
| **Other** | 50 | Various domain events |

**Total**: 119 events across 17 domains

### Event Flow

1. Domain aggregate emits event
2. Event saved to `outbox_events` table (transactional)
3. PostgreSQL NOTIFY triggers immediately
4. Outbox processor publishes to RabbitMQ (<100ms)
5. Consumers process (webhooks, analytics, integrations)

---

## ⚙️ Workflows

### Session Timeout Workflow (Temporal)

```
1. Session created → Temporal workflow started
2. Workflow waits for session timeout (configurable per pipeline)
3. If no messages received → session ends automatically
4. If messages received → workflow extends timeout
5. On timeout → `session.ended` event emitted
```

### Message Processing Workflow

```
1. WAHA → Webhook → API
2. Find/Create Contact
3. Find/Create Session (starts Temporal workflow)
4. Create Message
5. Emit events → Outbox → RabbitMQ
6. Consumers process (automation, webhooks, etc)
```

---

## 🔐 Security

### Multi-Tenancy (RLS)

PostgreSQL automatically filters by `tenant_id`:

```sql
SET LOCAL "app.tenant_id" = 'tenant-uuid';
-- All queries respect RLS policies
```

### RBAC Roles

| Role | Permissions |
|------|-------------|
| **admin** | Full access |
| **agent** | Manage contacts, messages, sessions |
| **viewer** | Read-only |

### Authentication Methods

1. **Bearer Token**: `Authorization: Bearer {token}`
2. **API Key**: `X-API-Key: {api_key}`
3. **Dev Headers** (dev only): `X-Dev-User-ID: {uuid}`

---

## 🧪 Testing

### Run Tests

```bash
make test                   # Unit tests
make test-coverage          # Coverage report
make test-domain            # Domain tests
make setup-all-complete     # Full E2E test ⭐
```

### E2E Test (`make setup-all-complete`)

Tests the complete system:
1. Create user (auth)
2. Create project
3. Create pipeline with statuses
4. Create WhatsApp channel
5. Send all message types (text, image, audio, video, document, location, contact)
6. Verify database (contacts, sessions, messages, events)

**Requirements**: API running + WAHA with session `5511999999999`

---

## 📚 Related Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design and patterns |
| [DEV_GUIDE.md](DEV_GUIDE.md) | Developer onboarding |
| [README.md](README.md) | Quick start |
| [CHANGELOG.md](CHANGELOG.md) | Version history |
| [MIGRATIONS.md](MIGRATIONS.md) | Database migrations |
| [OUTBOX_NO_POLLING.md](OUTBOX_NO_POLLING.md) | Event processing architecture |
| [WEBSOCKET_API.md](WEBSOCKET_API.md) | Real-time messaging |

**Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

**Last Updated**: 2025-10-11
