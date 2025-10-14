# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Project Overview

**Ventros CRM** is an AI-powered customer relationship management system built with Go, focusing on multi-channel communication (WhatsApp, Instagram, Facebook), conversation intelligence, and event-driven automation.

**Tech Stack**: Go 1.25.1, PostgreSQL 15+ (RLS), RabbitMQ 3.12+, Redis 7.0+, Temporal

**Architecture Score**: 8.0/10 (Production-ready backend, see AI_REPORT.md)

---

## Essential Development Commands

### Daily Development Workflow

```bash
# Start infrastructure (PostgreSQL, RabbitMQ, Redis, Temporal, Keycloak)
make infra

# Run API locally (requires make infra)
make api

# Full reset from scratch (clean DB + migrations + API)
make reset-full

# Stop infrastructure but keep data
make infra-stop

# Clean everything (DESTRUCTIVE - removes all data)
make clean
```

### Testing

```bash
# Run all tests (unit + integration + e2e)
make test

# Unit tests only (~2 min, no dependencies)
make test-unit

# Integration tests (~10 min, requires: make infra)
make test-integration

# E2E tests (~10 min, requires: make infra + make api in separate terminal)
make test-e2e

# Test coverage report
make test-coverage
```

### Code Quality

```bash
# Format code (ALWAYS run before commit)
make fmt

# Run linter
make lint

# Run go vet
make vet

# Generate Swagger docs
make swagger
```

### Database

```bash
# Run GORM AutoMigrate (DEV ONLY - creates schema automatically)
make migrate-auto

# Apply SQL migrations (PRODUCTION)
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status
```

### Build

```bash
# Build binary for local testing
make build

# Build for Linux (Docker)
make build-linux

# Build and run binary (test production build)
make run-binary
```

---

## Architecture Overview

### Design Patterns

**Ventros CRM follows strict architectural patterns**:

1. **Domain-Driven Design (DDD)** - 30 aggregates across 3 bounded contexts
2. **Hexagonal Architecture** - Clear separation: Domain → Application → Infrastructure
3. **Event-Driven Architecture** - 182+ domain events with Outbox Pattern (<100ms latency)
4. **CQRS** - Commands (80+) and Queries (20+) separated
5. **Saga Pattern** - Orchestration for complex workflows (Temporal)
6. **Multi-tenancy** - Row-Level Security (RLS) with tenant_id everywhere

### Directory Structure

```
ventros-crm/
├── cmd/
│   ├── api/              # Main API server entry point
│   ├── automigrate/      # GORM AutoMigrate tool (dev only)
│   └── migrate/          # SQL migrations runner
│
├── internal/
│   ├── domain/           # DOMAIN LAYER (Pure business logic)
│   │   ├── crm/          # CRM bounded context (23 aggregates)
│   │   │   ├── contact/  # Contact aggregate (events, repository, errors)
│   │   │   ├── session/  # Session aggregate
│   │   │   ├── message/  # Message aggregate
│   │   │   ├── channel/  # Channel aggregate
│   │   │   ├── pipeline/ # Pipeline aggregate with automation rules
│   │   │   ├── agent/    # Agent aggregate
│   │   │   ├── chat/     # Chat aggregate (group conversations)
│   │   │   └── ...       # 16 more aggregates
│   │   ├── automation/   # Automation bounded context (3 aggregates)
│   │   │   ├── campaign/ # Campaign aggregate
│   │   │   ├── sequence/ # Sequence aggregate
│   │   │   └── broadcast/# Broadcast aggregate
│   │   └── core/         # Core bounded context (4 aggregates)
│   │       ├── billing/  # Billing aggregate (Stripe integration)
│   │       ├── project/  # Project aggregate (multi-tenancy)
│   │       └── shared/   # Shared domain primitives
│   │
│   ├── application/      # APPLICATION LAYER (Use cases)
│   │   ├── commands/     # Write operations (CQRS)
│   │   │   ├── contact/  # Contact command handlers
│   │   │   ├── message/  # Message command handlers (send, confirm delivery)
│   │   │   └── ...
│   │   └── queries/      # Read operations (CQRS)
│   │
│   └── infrastructure/   # INFRASTRUCTURE LAYER (External concerns)
│       ├── http/         # HTTP handlers (Presentation)
│       │   ├── handlers/ # Gin handlers (thin adapters, delegate to commands)
│       │   ├── middleware/# Auth, RLS, rate limiting
│       │   └── routes/   # Route definitions
│       ├── persistence/  # Database (Repositories)
│       │   ├── entities/ # GORM entities
│       │   └── gorm_*_repository.go  # Repository implementations
│       ├── messaging/    # RabbitMQ (Event Bus, Outbox Pattern)
│       ├── channels/     # External integrations (WAHA, etc)
│       └── workflow/     # Temporal workflows
│
├── infrastructure/database/migrations/  # SQL migrations (versioned)
├── guides/              # Documentation
│   ├── domain_mapping/  # 23 aggregate docs
│   └── MAKEFILE.md      # Complete Makefile guide
│
├── DEV_GUIDE.md         # Complete developer guide (START HERE!)
├── TODO.md              # Roadmap with priorities (see security P0!)
├── AI_REPORT.md         # Architectural audit (8.0/10 score)
└── MAKEFILE.md          # Quick command reference
```

### Layer Dependency Rules

**CRITICAL: Never violate these rules**:

```
Domain ← Application ← Infrastructure
(No dependencies)  (Depends on Domain)  (Depends on Domain + Application)
```

- **Domain layer**: Pure Go, no external dependencies, only business logic
- **Application layer**: Orchestrates domain aggregates, calls repositories via interfaces
- **Infrastructure layer**: Implements interfaces (DB, HTTP, messaging, external APIs)

---

## Key Architectural Patterns

### 1. Command Handler Pattern (100% adoption)

All write operations follow this pattern:

```go
// 1. Command struct (input validation)
type CreateContactCommand struct {
    ProjectID uuid.UUID
    TenantID  string
    Name      string
}

// 2. Command handler (orchestration)
type CreateContactHandler struct {
    contactRepo domain.Repository
    eventBus    shared.EventBus
}

func (h *CreateContactHandler) Handle(ctx context.Context, cmd CreateContactCommand) (*Contact, error) {
    // Validate command
    // Create domain aggregate
    // Persist via repository
    // Publish events via event bus
    return contact, nil
}

// 3. HTTP handler (thin adapter)
func (h *ContactHandler) CreateContact(c *gin.Context) {
    // Parse request → Build command → Delegate to handler → Return response
}
```

### 2. Optimistic Locking (MANDATORY for all aggregates)

```go
type Contact struct {
    id      uuid.UUID
    version int  // REQUIRED: Starts at 1, increments on each update
    // ... other fields
}

// Repository Save() must check version
func (r *Repository) Save(ctx context.Context, contact *Contact) error {
    result := r.db.
        Where("id = ? AND version = ?", contact.ID(), contact.Version()).
        Updates(map[string]interface{}{
            "version": contact.Version() + 1,  // Increment
            // ... other fields
        })

    if result.RowsAffected == 0 {
        return ErrConcurrentUpdateConflict  // Conflict detected
    }
    return nil
}
```

**Status**: 16/30 aggregates (53%) - 14 still need version field (see TODO.md)

### 3. Event Publishing via Outbox Pattern

```go
// Events are automatically published via Outbox Pattern
// 1. Domain aggregate emits events
contact := NewContact(...)  // Emits contact.created event

// 2. Repository Save stores aggregate + events in same transaction
repo.Save(ctx, contact)  // Atomic: INSERT contact + INSERT outbox_events

// 3. PostgreSQL NOTIFY triggers immediate processing (<100ms)
// 4. Outbox worker publishes to RabbitMQ
// 5. Consumers process events asynchronously
```

**NEVER publish events directly to RabbitMQ** - always use event bus which persists to outbox first.

### 4. Multi-Tenancy with Row-Level Security (RLS)

```go
// Every table MUST have tenant_id
type ContactEntity struct {
    ID       uuid.UUID `gorm:"primary_key"`
    TenantID string    `gorm:"type:text;not null;index"`  // REQUIRED
    // ... other fields
}

// RLS middleware sets tenant context automatically
func (m *RLSMiddleware) Handle(c *gin.Context) {
    tenantID := c.GetString("tenant_id")  // From JWT
    // Set PostgreSQL session variable
    db.Exec("SET app.current_tenant = ?", tenantID)
}

// PostgreSQL policy enforces isolation (defined in migrations)
CREATE POLICY contacts_tenant_isolation ON contacts
    USING (tenant_id = current_setting('app.current_tenant')::TEXT);
```

---

## Critical Development Guidelines

### ALWAYS Do

1. **Run `make fmt` before every commit** - Code must be formatted
2. **Add `version int` field to new aggregates** - Optimistic locking is mandatory
3. **Add `tenant_id` to new tables** - Multi-tenancy is mandatory
4. **Use Command Handler Pattern** - See P0.md for template
5. **Emit domain events** - All state changes must emit events
6. **Write tests** - Unit tests for domain (100%), application (80%+), integration tests for repositories
7. **Follow event naming**: `aggregate.action` (e.g., `contact.created`, `session.ended`)
8. **Use Outbox Pattern** - Never publish events directly to RabbitMQ
9. **Implement soft delete** - Add `deleted_at` field, never hard delete
10. **Configure RLS policies** - Every multi-tenant table needs RLS policy in migration

### NEVER Do

1. ❌ **Manipulate domain aggregates directly in HTTP handlers** - Use command handlers
2. ❌ **Expose domain entities via API** - Always use DTOs (request/response)
3. ❌ **Hard delete records** - Always soft delete with `deleted_at`
4. ❌ **Publish events directly to RabbitMQ** - Use EventBus + Outbox
5. ❌ **Use GORM AutoMigrate in production** - Use SQL migrations only
6. ❌ **Ignore optimistic locking errors** - Implement retry logic
7. ❌ **Skip tests** - Tests are mandatory for all layers
8. ❌ **Commit without `make fmt`** - Code must be formatted
9. ❌ **Use generic setters** - Create business methods with validation
10. ❌ **Break dependency rules** - Domain must not depend on Infrastructure

---

## Testing Strategy

**Test Pyramid** (Mike Cohn, 2009):

```
        /\
       /E2E\      ← 10% (5 tests) - Full stack
      /------\
     /Integr.\   ← 20% (2 tests) - Requires: make infra
    /----------\
   /   Unit    \  ← 70% (61 tests) - Fast, no dependencies
  /______________\
```

### Run Tests

```bash
# Unit tests (~2 min) - Domain + Application logic
make test-unit

# Integration tests (~10 min) - Database + RabbitMQ
# Requires: make infra (PostgreSQL, RabbitMQ, Redis must be running)
make test-integration

# E2E tests (~10 min) - Full HTTP workflows
# Requires: make infra + make api (in separate terminal)
make test-e2e

# Coverage report (HTML)
make test-coverage
```

### Coverage Goals

- **Domain Layer**: 100% (business-critical)
- **Application Layer**: 80%+
- **Infrastructure Layer**: 60%+
- **Overall**: 82%+ (current: 82%)

---

## Database Migrations

### IMPORTANT

- **DEV**: Use `make migrate-auto` (GORM AutoMigrate, quick schema sync)
- **PRODUCTION**: Use SQL migrations only (`make migrate-up`)

### Creating Migrations

```bash
# 1. Create migration files (XXX_description.up.sql + .down.sql)
touch infrastructure/database/migrations/000050_add_custom_fields.up.sql
touch infrastructure/database/migrations/000050_add_custom_fields.down.sql

# 2. Write UP migration
cat > infrastructure/database/migrations/000050_add_custom_fields.up.sql <<EOF
CREATE TABLE custom_fields (
    id UUID PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,  -- Optimistic locking
    tenant_id TEXT NOT NULL,             -- Multi-tenancy
    project_id UUID NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,                -- Soft delete

    CONSTRAINT fk_custom_fields_project FOREIGN KEY (project_id)
        REFERENCES projects(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_custom_fields_tenant ON custom_fields(tenant_id);
CREATE INDEX idx_custom_fields_project ON custom_fields(project_id);
CREATE INDEX idx_custom_fields_deleted ON custom_fields(deleted_at) WHERE deleted_at IS NULL;

-- Row-Level Security
ALTER TABLE custom_fields ENABLE ROW LEVEL SECURITY;

CREATE POLICY custom_fields_tenant_isolation ON custom_fields
    USING (tenant_id = current_setting('app.current_tenant')::TEXT);
EOF

# 3. Write DOWN migration
cat > infrastructure/database/migrations/000050_add_custom_fields.down.sql <<EOF
DROP POLICY IF EXISTS custom_fields_tenant_isolation ON custom_fields;
DROP TABLE IF EXISTS custom_fields CASCADE;
EOF

# 4. Apply migration
make migrate-up

# 5. Test rollback
make migrate-down
```

---

## Common Workflows

### Adding a New Aggregate (Full DDD Pattern)

See **DEV_GUIDE.md** for complete step-by-step guide (1,536 lines).

**Quick checklist**:
1. Domain layer: Aggregate root + events + repository interface + errors
2. Application layer: Command + command handler + DTOs
3. Infrastructure layer: GORM entity + repository implementation + HTTP handler
4. Database: SQL migration (up + down)
5. Tests: Unit (domain + application) + integration (repository) + E2E (HTTP)

### Adding a New HTTP Endpoint

```bash
# 1. Create command/query in application layer
# 2. Create handler in application layer
# 3. Create HTTP handler in infrastructure/http/handlers/
# 4. Add Swagger comments (@Summary, @Tags, @Accept, @Produce, @Param, @Success, @Failure, @Router)
# 5. Register route in infrastructure/http/routes/routes.go
# 6. Generate Swagger docs
make swagger
# 7. Test with curl or E2E test
```

### Publishing a Domain Event

```go
// 1. Define event in domain layer
type ContactCreatedEvent struct {
    ContactID uuid.UUID
    Name      string
    EventMeta EventMetadata
}

func (e ContactCreatedEvent) EventType() string {
    return "contact.created"  // Format: aggregate.action
}

// 2. Emit in aggregate
func NewContact(...) (*Contact, error) {
    contact := &Contact{...}
    contact.addEvent(NewContactCreatedEvent(contact))  // Emit
    return contact, nil
}

// 3. Publish in command handler
func (h *CreateContactHandler) Handle(...) {
    contact, _ := domain.NewContact(...)
    h.contactRepo.Save(ctx, contact)  // Atomic: aggregate + events

    for _, event := range contact.DomainEvents() {
        h.eventBus.Publish(ctx, event)  // Outbox Pattern
    }
}

// 4. Consumer (if needed) subscribes in RabbitMQ
// See: infrastructure/messaging/contact_event_consumer.go
```

---

## Critical Security Notes

**⚠️ 5 CRITICAL P0 vulnerabilities exist (see TODO.md)**:

1. **Dev Mode Bypass** (CVSS 9.1) - `middleware/auth.go:41` allows auth bypass in production
2. **SSRF in Webhooks** (CVSS 9.1) - No URL validation, can access internal services
3. **BOLA in 60 GET endpoints** (CVSS 8.2) - No ownership checks
4. **Resource Exhaustion** (CVSS 7.5) - No max page size (19 queries vulnerable)
5. **RBAC Missing** (CVSS 7.1) - 95 endpoints lack role checks

**DO NOT deploy to production until Sprint 1-2 security fixes are completed.**

---

## Documentation References

| Document | Purpose | When to Use |
|----------|---------|-------------|
| `README.md` | Project overview | First time setup |
| `DEV_GUIDE.md` | Complete developer guide | Implementing features |
| `TODO.md` | Roadmap and priorities | Planning work |
| `AI_REPORT.md` | Architectural audit | Understanding quality |
| `MAKEFILE.md` | Command reference | Quick lookup |
| `P0.md` | Handler refactoring (done) | Reference for patterns |
| `guides/domain_mapping/` | 23 aggregate docs | Understanding domain |
| `guides/TESTING.md` | Testing strategy | Writing tests |

---

## API Endpoints

**Total**: 158 endpoints across 10 products

- **Health**: 1 endpoint (`/health`)
- **Auth**: 4 endpoints (register, login, refresh, profile)
- **CRM**: 77 endpoints (contacts, sessions, messages, channels, pipelines, chats, agents, notes)
- **Automation**: 28 endpoints (campaigns, sequences, broadcasts, automation rules)
- **Billing**: 8 endpoints (Stripe integration)
- **Tracking**: 6 endpoints (ad conversion attribution)
- **Webhooks**: 4 endpoints (subscriptions)
- **WebSocket**: 1 endpoint (real-time messaging)
- **Queue**: 7 endpoints (RabbitMQ admin)
- **Test**: 22 endpoints (dev only)

**Swagger UI**: http://localhost:8080/swagger/index.html

---

## Environment Variables

**Critical variables** (see `.env`):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ventros_crm

# WAHA (WhatsApp API)
WAHA_BASE_URL=https://waha.ventros.cloud
WAHA_API_KEY=your-key

# AI Providers (optional)
GROQ_API_KEY=gsk_...           # Whisper (audio) - FREE, priority 1
VERTEX_PROJECT_ID=...          # Gemini Vision (images)
LLAMAPARSE_API_KEY=llx_...     # Documents (PDF, Word, etc)
OPENAI_API_KEY=sk-...          # Whisper fallback

# JWT
JWT_SECRET=your-secret-key

# Environment
ENV=development  # CRITICAL: Must be "production" in prod (disables dev auth bypass)
```

---

## Common Issues & Solutions

### "Port 8080 already in use"

```bash
lsof -i :8080
kill -9 <PID>
# OR
make clean  # Kills all processes + cleans containers
```

### "Cannot connect to database"

```bash
# Check if infrastructure is running
docker ps

# Restart infrastructure
make infra-stop
make infra
```

### "Migrations failed"

```bash
# Clean and restart
make clean
make infra
make api  # Migrations run automatically on startup
```

### "Tests failing"

```bash
# Ensure infrastructure is running
make infra

# Wait for services to be ready
sleep 10

# Run tests
make test-unit        # No dependencies
make test-integration # Requires infra
make test-e2e        # Requires infra + API in separate terminal
```

---

## AI/ML Components

**Status**: Message enrichment 100% complete, memory service 80% missing

### Implemented

- **Message Enrichment Service** (12 providers)
  - Audio transcription: Groq Whisper (FREE, 216x real-time) → OpenAI Whisper fallback
  - Image OCR: Vertex Vision (Gemini 1.5 Flash, $0.00025/image)
  - Document parsing: LlamaParse ($1-3 per 1000 pages)
  - Profile picture scoring: Gemini Vision (0-10 score)
  - Automatic provider routing with fallbacks

### Missing (see TODO.md)

- Vector search (pgvector + embeddings)
- Hybrid search (vector + keyword + graph)
- Memory facts extraction
- Python ADK (multi-agent system)
- MCP Server (Claude Desktop integration)
- gRPC API (Go ↔ Python communication)

---

## Quick Reference: Event Naming

**Format**: `aggregate.action` (lowercase, past tense)

**Examples**:
```
contact.created
contact.updated
contact.deleted
session.started
session.ended
message.sent
message.delivered
campaign.activated
pipeline.status_changed
agent.assigned
```

**Bad examples**:
```
❌ create_contact    (Wrong format)
❌ ContactCreated    (Wrong casing)
❌ contact_create    (Wrong tense)
```

---

## Dependency Versions

**Go**: 1.25.1+
**PostgreSQL**: 15+
**RabbitMQ**: 3.12+
**Redis**: 7.0+
**Temporal**: Latest

**Key Go packages**:
- Gin (HTTP framework)
- GORM (ORM)
- Temporal SDK (workflows)
- uuid (Google)
- logrus + zap (logging)
- testify (testing)

---

## Performance Targets

- **API Latency**: <200ms (p95)
- **Cache Hit Rate**: >70% (when implemented)
- **Event Latency**: <100ms (Outbox Pattern with NOTIFY)
- **Query Latency**: <500ms (complex queries)
- **Test Coverage**: 82%+

---

## Additional Notes

1. **The Chat aggregate is fully implemented** despite not being in old documentation
2. **Total aggregates: 30** (not 23 as previously documented)
3. **Handler pattern refactoring is 100% complete** (see P0.md)
4. **Optimistic locking is only 53% complete** (16/30 aggregates)
5. **Cache integration is 0%** despite Redis being configured
6. **Security has 5 P0 vulnerabilities** - address before production

---

**Last Updated**: 2025-10-13
**Maintainer**: Ventros CRM Team
**Architecture Score**: 8.0/10 (Backend solid, AI has gaps)
