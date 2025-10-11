# 🗺️ Domain Aggregates Mapping - Ventros CRM

**Last Updated**: 2025-10-10
**Total Aggregates**: 23
**Architecture**: Domain-Driven Design (DDD) + Event-Driven

---

## 📋 Overview

Este documento mapeia **todos os 23 Domain Aggregates** do Ventros CRM, fornecendo uma visão completa da arquitetura de domínio.

Cada aggregate é documentado individualmente com:
- Purpose e responsabilidades
- Domain model (root, value objects, invariants)
- Events emitidos
- Repository interface
- Commands/Queries (CQRS)
- Use cases implementados
- Sugestões de melhorias

---

## 🏗️ Architecture Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                    HEXAGONAL ARCHITECTURE                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌───────────────────────────────────────────────────┐     │
│  │         DOMAIN LAYER (Pure Business Logic)        │     │
│  │                                                    │     │
│  │  ┌──────────────┐  ┌──────────────┐             │     │
│  │  │  Aggregate   │  │    Events    │             │     │
│  │  │    Root      │──│   (104+)     │             │     │
│  │  └──────────────┘  └──────────────┘             │     │
│  │         │                  │                      │     │
│  │         ├─ Value Objects   │                      │     │
│  │         ├─ Invariants      │                      │     │
│  │         └─ Business Rules  │                      │     │
│  │                            │                      │     │
│  └────────────────────────────┼──────────────────────┘     │
│                               │                            │
│  ┌────────────────────────────┼──────────────────────┐     │
│  │      APPLICATION LAYER     │                      │     │
│  │                            ▼                      │     │
│  │  Use Cases ───────▶ Domain Events                │     │
│  │  Commands                  │                      │     │
│  │  Queries                   │                      │     │
│  └────────────────────────────┼──────────────────────┘     │
│                               │                            │
│  ┌────────────────────────────┼──────────────────────┐     │
│  │    INFRASTRUCTURE LAYER    │                      │     │
│  │                            ▼                      │     │
│  │  Repositories ◀────── GORM Entities              │     │
│  │  Event Bus    ◀────── RabbitMQ (15+ queues)      │     │
│  │  HTTP/REST    ◀────── Gin Handlers               │     │
│  │  Workflows    ◀────── Temporal (7 workflows)     │     │
│  └──────────────────────────────────────────────────┘     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 📊 Aggregate Categories

### 🔷 Core CRM Aggregates (Alta Prioridade)

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Contact](contact_aggregate.md) | Customer/Lead management | ✅ Complete | 8+ | `internal/domain/contact/` |
| [Session](session_aggregate.md) | Conversation sessions | ✅ Complete | 6+ | `internal/domain/session/` |
| [Message](message_aggregate.md) | Chat messages | ✅ Complete | 5+ | `internal/domain/message/` |
| [Pipeline](pipeline_aggregate.md) | Sales pipeline & automation | ✅ Complete | 7+ | `internal/domain/pipeline/` |
| [Agent](agent_aggregate.md) | Human/Bot agents | ✅ Complete | 4+ | `internal/domain/agent/` |

### 📨 Communication Aggregates

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Channel](channel_aggregate.md) | Communication channels (WhatsApp, etc) | ✅ Complete | 5+ | `internal/domain/channel/` |
| [ChannelType](channel_type_aggregate.md) | Channel configurations | ✅ Complete | 2+ | `internal/domain/channel_type/` |
| [Broadcast](broadcast_aggregate.md) | Mass messaging | ⚠️ Partial | 3+ | `internal/domain/broadcast/` |

### 📊 Analytics & Tracking

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Tracking](tracking_aggregate.md) | Attribution tracking (UTMs) | ✅ Complete | 4+ | `internal/domain/tracking/` |
| [ContactEvent](contact_event_aggregate.md) | Contact timeline/activity | ✅ Complete | 2+ | `internal/domain/contact_event/` |
| [Event](event_aggregate.md) | Domain event log | ✅ Complete | - | `internal/domain/event/` |

### 🔐 Auth & Multi-tenancy

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [User](user_aggregate.md) | System users (formerly Customer) | ✅ Complete | 3+ | `internal/domain/user/` |
| [Project](project_aggregate.md) | Tenant/workspace | ✅ Complete | 3+ | `internal/domain/project/` |
| [Credential](credential_aggregate.md) | OAuth tokens & credentials | ✅ Complete | 4+ | `internal/domain/credential/` |

### 💰 Billing & Payment

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Billing](billing_aggregate.md) | Billing accounts | ⚠️ Partial | 5+ | `internal/domain/billing/` |

### 🔔 Notifications & Webhooks

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Webhook](webhook_aggregate.md) | Webhook subscriptions | ✅ Complete | 3+ | `internal/domain/webhook/` |

### 📝 Supporting Aggregates

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Note](note_aggregate.md) | Internal notes | ✅ Complete | 2+ | `internal/domain/note/` |
| [ContactList](contact_list_aggregate.md) | Dynamic contact segments | ✅ Complete | 4+ | `internal/domain/contact_list/` |
| [AgentSession](agent_session_aggregate.md) | Agent online status | ✅ Complete | 3+ | `internal/domain/agent_session/` |

### 🔄 Infrastructure Aggregates

| Aggregate | Purpose | Status | Events | Location |
|-----------|---------|--------|--------|----------|
| [Outbox](outbox_aggregate.md) | Transactional outbox pattern | ✅ Complete | - | `internal/domain/outbox/` |
| [Saga](saga_aggregate.md) | Distributed transactions | ✅ Complete | 4+ | `internal/domain/saga/` |

### 🛠️ Shared Components

| Component | Purpose | Location |
|-----------|---------|----------|
| [Shared](shared_aggregate.md) | Value Objects, Base Types | `internal/domain/shared/` |

---

## 🎯 Status Legend

- ✅ **Complete**: Aggregate fully implemented with tests
- ⚠️ **Partial**: Basic implementation, missing features
- ❌ **TODO**: Not implemented yet

---

## 📈 Statistics

```
Total Aggregates:        23
Core CRM:                5
Communication:           3
Analytics:               3
Auth & Multi-tenancy:    3
Billing:                 1
Notifications:           1
Supporting:              4
Infrastructure:          2
Shared:                  1

Total Domain Events:     104+
Total Repositories:      18
Total Use Cases:         40+
Total Workflows:         7
```

---

## 🔗 Cross-Aggregate Relationships

### Primary Relationships

```
User (Tenant)
  └─► Project
       ├─► Contact
       │    ├─► Session
       │    │    └─► Message
       │    ├─► ContactEvent (timeline)
       │    ├─► Tracking (attribution)
       │    └─► Note
       ├─► Pipeline
       │    └─► PipelineStatus
       ├─► Channel
       │    └─► ChannelType
       └─► Agent
            └─► AgentSession

Outbox ◄── All Aggregates (publish events)
Saga ◄─── Orchestrated workflows
```

### Event Flow

```
Domain Aggregate
  └─► Emit Event
       └─► Outbox (Transactional)
            └─► PostgreSQL LISTEN/NOTIFY
                 └─► RabbitMQ
                      ├─► Consumers (other aggregates)
                      ├─► Webhooks (external systems)
                      └─► Temporal (workflows/sagas)
```

---

## 📚 Documentation Index

### Core CRM
1. [Contact Aggregate](contact_aggregate.md) - Customer/Lead management
2. [Session Aggregate](session_aggregate.md) - Conversation sessions
3. [Message Aggregate](message_aggregate.md) - Chat messages
4. [Pipeline Aggregate](pipeline_aggregate.md) - Sales funnel
5. [Agent Aggregate](agent_aggregate.md) - Human/Bot agents

### Communication
6. [Channel Aggregate](channel_aggregate.md) - Communication channels
7. [ChannelType Aggregate](channel_type_aggregate.md) - Channel configs
8. [Broadcast Aggregate](broadcast_aggregate.md) - Mass messaging

### Analytics & Tracking
9. [Tracking Aggregate](tracking_aggregate.md) - Attribution (UTMs)
10. [ContactEvent Aggregate](contact_event_aggregate.md) - Activity timeline
11. [Event Aggregate](event_aggregate.md) - Domain events log

### Auth & Multi-tenancy
12. [User Aggregate](user_aggregate.md) - System users
13. [Project Aggregate](project_aggregate.md) - Tenants/workspaces
14. [Credential Aggregate](credential_aggregate.md) - OAuth & secrets

### Billing
15. [Billing Aggregate](billing_aggregate.md) - Billing accounts

### Notifications
16. [Webhook Aggregate](webhook_aggregate.md) - Webhook subscriptions

### Supporting
17. [Note Aggregate](note_aggregate.md) - Internal notes
18. [ContactList Aggregate](contact_list_aggregate.md) - Dynamic segments
19. [AgentSession Aggregate](agent_session_aggregate.md) - Agent status

### Infrastructure
20. [Outbox Aggregate](outbox_aggregate.md) - Event publishing
21. [Saga Aggregate](saga_aggregate.md) - Distributed transactions

### Shared
22. [Shared Components](shared_aggregate.md) - Value Objects & base types

---

## 🚀 Next Steps

1. ✅ Complete aggregate documentation (this directory)
2. ❌ Research famous CRM APIs (HubSpot, Salesforce, etc)
3. ❌ Implement missing use cases
4. ❌ Add new Chat aggregate (CRITICAL)
5. ❌ Complete WAHA integration
6. ❌ Implement Redis caching layer

---

**For detailed information about each aggregate, click on the individual documentation files above.**
