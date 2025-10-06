# 🏗️ Architecture Overview - Ventros CRM

This document provides a high-level overview of the Ventros CRM architecture. For detailed guides, see the [/guides/architecture](guides/architecture/) directory.

---

## 📐 Architectural Principles

Ventros CRM is built on three foundational patterns:

1. **Domain-Driven Design (DDD)** - Rich domain models with business logic encapsulation
2. **Event-Driven Architecture** - Asynchronous communication via events
3. **SAGA Pattern** - Distributed transactions with Temporal workflows

---

## 🔷 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Layer                            │
│  (Handlers, Middleware, Routes, DTOs)                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                   Application Layer                          │
│  (Use Cases, Commands, Queries, Services, Assemblers)       │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                     Domain Layer                             │
│  (Aggregates, Entities, Value Objects, Domain Events)       │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 Infrastructure Layer                         │
│  (Repositories, Event Bus, External APIs, Database)         │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎯 Domain Model

### Aggregates (Roots)

**Contact** - Central entity representing a customer/lead
- Identity: UUID
- Value Objects: Email, Phone
- Invariants: Name required, unique phone per project
- Events: ContactCreated, ContactUpdated, ContactDeleted

**Session** - Conversation with timeout management
- Identity: UUID
- Lifecycle: Active → Ended (via Temporal workflow)
- Metrics: Message counts, duration, sentiment
- Events: SessionStarted, SessionEnded, MessageRecorded

**Message** - Individual message in a conversation
- Identity: UUID
- Types: Text, Image, Document, Audio, Video
- Status: Sent → Delivered → Read
- Events: MessageCreated, MessageDelivered, MessageRead

### Relationships
```
Project (1) ──── (N) Contact
Contact (1) ──── (N) Session
Session (1) ──── (N) Message
Channel (1) ──── (N) Message
```

---

## 🔄 Event-Driven Flow

### Event Types

#### Domain Events (Internal)
```
domain.events.contact.created
domain.events.contact.updated
domain.events.session.started
domain.events.session.ended
domain.events.message.created
domain.events.message.delivered
```

#### Integration Events (External - WAHA)
```
waha.events.message
waha.events.message.ack
waha.events.call.received
waha.events.label.upsert
```

### Event Flow

```
┌──────────┐      ┌──────────────┐      ┌──────────────┐
│   API    │─────▶│   Use Case   │─────▶│  Aggregate   │
└──────────┘      └──────────────┘      └──────┬───────┘
                                               │ Raise Event
                                               ▼
                  ┌──────────────────────────────────────┐
                  │         Domain Event Bus             │
                  │          (RabbitMQ)                  │
                  └──────┬───────────────────────┬───────┘
                         │                       │
                         ▼                       ▼
                  ┌─────────────┐        ┌──────────────┐
                  │  Subscribers│        │   Webhooks   │
                  │  (Internal) │        │  (External)  │
                  └─────────────┘        └──────────────┘
```

---

## ⚙️ SAGA Pattern with Temporal

### Session Lifecycle Workflow

```go
ProcessInboundMessage Workflow:
  1. Find or Create Contact    ✓
     └─ Compensation: Delete if new
  
  2. Find or Create Session    ✓
     └─ Compensation: Delete if new
  
  3. Create Message            ✓
     └─ Compensation: Delete message
  
  4. Publish Events            ✓
     └─ No compensation needed
```

### Workflow Features
- ✅ **Durable Execution** - Survives process restarts
- ✅ **Automatic Retries** - Configurable retry policies
- ✅ **Signal Handling** - Extend timeout dynamically
- ✅ **Scheduled Cleanup** - Periodic orphaned session cleanup

---

## 🗂️ Directory Structure

```
ventros-crm/
├── cmd/                    # Application entry points
│   └── api/               # Main API server
│       └── main.go
│
├── internal/              # Private application code
│   ├── domain/           # Domain Layer (DDD)
│   │   ├── contact/      # Contact Aggregate
│   │   ├── session/      # Session Aggregate
│   │   ├── message/      # Message Aggregate
│   │   └── shared/       # Shared domain interfaces
│   │
│   ├── application/      # Application Layer
│   │   ├── contact/      # Contact use cases
│   │   ├── session/      # Session use cases
│   │   ├── message/      # Message use cases
│   │   ├── dtos/         # Data Transfer Objects
│   │   └── assemblers/   # Domain → DTO conversion
│   │
│   └── workflows/        # Temporal Workflows
│       └── session/      # Session lifecycle workflows
│
├── infrastructure/       # Infrastructure Layer
│   ├── persistence/     # Database repositories (GORM)
│   │   ├── entities/    # ORM entities
│   │   └── gorm_*_repository.go
│   │
│   ├── messaging/       # Event Bus (RabbitMQ)
│   │   ├── domain_event_bus.go
│   │   └── waha_message_consumer.go
│   │
│   ├── http/           # HTTP Layer (Gin)
│   │   ├── handlers/   # Request handlers
│   │   ├── middleware/ # Auth, RBAC, RLS
│   │   └── routes/     # Route configuration
│   │
│   ├── channels/       # External integrations
│   │   └── waha/      # WhatsApp HTTP API
│   │
│   └── workflow/       # Temporal client/workers
│
├── guides/             # Documentation
├── deployments/        # Deployment configs
│   ├── docker/        # Docker setup
│   └── helm/          # Kubernetes Helm charts
│
└── scripts/           # Build/maintenance scripts
```

---

## 🔐 Security Architecture

### Multi-Tenancy
```
User Request
    ↓
Auth Middleware → Extract user_id, tenant_id
    ↓
RLS Middleware → SET user_id in PostgreSQL session
    ↓
PostgreSQL RLS Policies → Filter by tenant_id automatically
```

### RBAC (Role-Based Access Control)
```
Roles:
  - Admin      → Full access
  - Manager    → Manage team, view analytics
  - User       → CRUD own resources
  - ReadOnly   → View only

Resources:
  - Contact, Session, Message, Webhook, Analytics, User

Operations:
  - Create, Read, Update, Delete, List, Export
```

---

## 📊 Data Flow Example

### Inbound WhatsApp Message Flow

```
1. WAHA sends webhook
   POST /webhooks/waha
   
2. Handler validates signature
   
3. Use Case: ProcessInboundMessage
   ├─ Find or Create Contact
   ├─ Find or Create Session (starts Temporal workflow)
   ├─ Create Message
   └─ Publish Domain Events
   
4. Event Bus publishes to RabbitMQ
   
5. Subscribers process events:
   ├─ Webhook Notifier → External webhooks
   ├─ Analytics Service → Update metrics
   └─ AI Service → Generate summary (future)
   
6. Temporal Workflow manages session timeout
   ├─ Wait for timeout duration (30 min)
   ├─ If new message → extend timeout
   └─ On timeout → End session
```

---

## 🧩 Technology Stack

### Backend
- **Go 1.25.1** - Primary language
- **Gin** - HTTP framework
- **GORM** - ORM for PostgreSQL

### Infrastructure
- **PostgreSQL 16** - Primary database with RLS
- **RabbitMQ** - Message broker
- **Redis** - Cache and sessions
- **Temporal** - Workflow orchestration

### DevOps
- **Docker** - Containerization
- **Kubernetes** - Orchestration
- **Helm** - Package management

---

## 🎯 Design Decisions

### Why DDD?
- **Rich Domain Models** - Business logic stays in the domain
- **Ubiquitous Language** - Shared language between devs and business
- **Aggregate Boundaries** - Clear transaction boundaries

### Why Event-Driven?
- **Loose Coupling** - Services don't depend on each other
- **Scalability** - Easy to add new consumers
- **Audit Trail** - Complete history of events

### Why Temporal for SAGA?
- **Durable Execution** - No lost transactions
- **Easy Compensation** - Built-in rollback support
- **Observability** - View workflow history in UI

### Why NOT Microservices?
- **Team Size** - Small team, monolith is simpler
- **Performance** - No network overhead
- **Complexity** - Avoid distributed system problems
- **Future-Ready** - Modular monolith can split later

---

## 📈 Scalability

### Horizontal Scaling
- **Stateless API** - Scale pods horizontally
- **Connection Pooling** - Efficient database connections
- **Message Queues** - Distribute work across consumers

### Vertical Scaling
- **Database** - Increase PostgreSQL resources
- **Cache** - Add Redis cluster
- **Workers** - More Temporal workers

### Future Optimizations
- **CQRS** - Separate read/write databases
- **Event Sourcing** - Store events instead of state
- **Microservices** - Split bounded contexts

---

## 🔗 Further Reading

- [Domain-Driven Design Guide](guides/architecture/ddd.md)
- [Event-Driven Architecture Guide](guides/architecture/event-driven.md)
- [SAGA Pattern Guide](guides/architecture/saga.md)
- [Security & Multi-tenancy](guides/architecture/security.md)

---

**For questions or clarifications, see [CONTRIBUTING.md](CONTRIBUTING.md)**
