# 🏗️ Ventros CRM - Architecture

**Modular Monolith** com DDD + Event-Driven + SAGA (Temporal) + Outbox Pattern + Multi-Tenancy (RLS)

## Stack
- **Backend**: Go 1.25.1, Gin, GORM
- **Database**: PostgreSQL 16 (RLS, 40 migrations)
- **Messaging**: RabbitMQ (15+ queues, DLQ, Outbox Pattern)
- **Workflows**: Temporal (session timeout, automation)
- **Cache**: Redis
- **Integration**: WAHA (WhatsApp HTTP API)
- **Deploy**: Docker/Podman/Buildah, Kubernetes/Helm

## Layers
```
HTTP (Gin) → Application (Use Cases) → Domain (18+ Aggregates, 80+ Events) → Infrastructure (GORM, RabbitMQ, Temporal)
```

## Domain Model (18+ Aggregates)

**Core:** Contact, Session, Message, Pipeline, Channel, Agent, Tracking, Note, ...  
**Support:** WebhookSubscription, ContactList, Credential, BillingAccount, ContactEvent, ChannelType, Project, Customer, User

**Key Features:**
- 80+ domain events (ContactCreated, SessionEnded, MessageDelivered...)
- 13 message content types (text, image, audio, video, voice, document...)
- Automation: 20+ triggers (session.ended, message.received...) + 15+ actions (send_message, change_status...)
- Multi-tenancy: Row-Level Security (RLS) no PostgreSQL

**Relationships:** Customer → Project → Contact → Session → Message

## Event-Driven Architecture

**Outbox Pattern:** Transaction → Outbox Insert → PostgreSQL NOTIFY → RabbitMQ Publish → Subscribers  
**Benefits:** Guaranteed delivery, transactional consistency, no lost events

**Event Types:**
- **Domain Events (80+):** contact.*, session.*, message.*, pipeline.*, agent.*...
- **WAHA Events (15+ queues):** waha.events.message, waha.events.call.received...
- **AI Events:** message.ai.process_image_requested, message.ai.process_video_requested...

## SAGA Pattern (Temporal)

**Session Lifecycle Workflow:**
1. Session Created → Start Timer (30min from Pipeline config)
2. Wait for Timeout OR Signal (new message extends timeout)
3. On Timeout → End Session + Publish enriched SessionEnded event
4. Trigger Automations (if configured)

**Features:** Durable execution, automatic retries, signal handling, observability (Temporal UI)

## Security

**Multi-Tenancy (RLS):** PostgreSQL Row-Level Security auto-filters by tenant_id  
**RBAC:** Admin, Manager, User, ReadOnly roles  
**Auth:** JWT tokens, API keys, OAuth2 (future)  
**Encryption:** AES-256 (credentials), HMAC-SHA256 (webhooks)

## Directory Structure

```
cmd/api/                      # Main API server
internal/
  ├── domain/                # 18+ Aggregates, 80+ Events
  ├── application/           # Use Cases, DTOs, Services
  │   └── queries/          # ❌ EMPTY (TO IMPLEMENT)
  └── testing/              # Test helpers
infrastructure/
  ├── persistence/          # GORM (29 entities, 20 repos)
  ├── messaging/            # RabbitMQ + Outbox Pattern
  ├── http/                 # Gin (28 handlers, middleware)
  ├── channels/             # WAHA, WhatsApp
  ├── workflow/             # Temporal workers
  ├── database/migrations/  # 40 SQL migrations
  └── crypto/               # AES-256 encryption
.deploy/                    # Container + Helm charts
```

## Inbound WhatsApp Message Flow

```
WhatsApp → WAHA → POST /webhooks/:webhook_id
  → WAHAWebhookHandler → RabbitMQ (waha.events.message)
  → WAHAMessageConsumer → ProcessInboundMessageUseCase:
     ├─ Find Channel by webhook_id
     ├─ Find/Create Contact (phone + name)
     ├─ Find/Create Session (starts Temporal workflow)
     ├─ Create Message (13 types)
     ├─ Trigger AI Processing (if enabled)
     └─ Save + Outbox Insert (transaction)
  → Outbox Worker (NOTIFY) → RabbitMQ → Subscribers
  → Temporal (Session Timeout Management)
```

## Scalability

**Horizontal:** Stateless API pods, connection pooling, RabbitMQ consumers, Temporal workers  
**Vertical:** PostgreSQL resources, Redis cluster

## Current Limitations (TO IMPLEMENT)

❌ **CQRS Queries** - `queries/` directory EMPTY  
❌ **Validators** - No centralized validation  
❌ **Mappers** - Domain↔DTO scattered  
❌ **Cache Layer** - Redis unused for repos  
❌ **File Storage** - No abstraction  

✅ **Event Sourcing Ready** - 80+ events, domain_event_log  
✅ **Microservices Ready** - Clear bounded contexts

## Links

- **[DOCS.md](DOCS.md)** - Complete API documentation
- **[README.md](README.md)** - Quick start
- **[TODO.md](TODO.md)** - Tasks/roadmap
- **Swagger**: http://localhost:8080/swagger/index.html
- **Temporal**: http://localhost:8088
- **RabbitMQ**: http://localhost:15672

---

**Version**: 2025-10-10 | Based on actual codebase
