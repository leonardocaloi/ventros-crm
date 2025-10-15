# VENTROS CRM - PROJECT OVERVIEW

> **Status**: Production (v1.0) - 8.0/10 architectural score
>
> **Stack**: Go 1.25.1 + PostgreSQL 15+ + RabbitMQ 3.12+ + Redis 7.0+ + Temporal
>
> **Última atualização**: 2025-10-15

---

## 🎯 O QUE É O VENTROS CRM?

**Ventros CRM** é um sistema de gerenciamento de relacionamento com clientes (CRM) AI-powered construído com Go, focado em:

1. **Comunicação Multi-Canal**
   - WhatsApp (via WAHA)
   - Instagram Direct
   - Facebook Messenger
   - Email (futuro)
   - SMS (futuro)

2. **Inteligência de Conversação**
   - Transcrição de áudio (Groq Whisper)
   - OCR de imagens (Vertex Vision)
   - Parsing de documentos (LlamaParse)
   - Análise de sentimento
   - Extração de entidades (NER)

3. **Automação Event-Driven**
   - 182+ eventos de domínio
   - Outbox Pattern (<100ms latência)
   - Workflows com Temporal
   - Triggers de pipeline automáticos

4. **Memória Contextual (Memory Service)**
   - pgvector embeddings (768 dimensions)
   - Busca híbrida (vector + keyword + graph)
   - RRF (Reciprocal Rank Fusion) + Cross-Encoder Reranking
   - Knowledge graph de relacionamentos

---

## 🏗️ ARQUITETURA DE ALTO NÍVEL

```
┌─────────────────────────────────────────────────────────────────┐
│                        VENTROS CRM (Go)                         │
│                         MONOLITO MODULAR                        │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  API PRINCIPAL (Port 8080)                                │  │
│  │  - 158 endpoints REST (Gin)                               │  │
│  │  - WebSocket (real-time)                                  │  │
│  │  - Webhooks (WAHA, Stripe, etc)                          │  │
│  │  - JWT Auth + RBAC (Keycloak)                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  MCP SERVER (Port 8081) - Feature integrada               │  │
│  │  - 30+ tools para Python ADK                              │  │
│  │  - CRM operations (contacts, messages, sessions, etc)     │  │
│  │  - Document search (vector + keyword)                     │  │
│  │  - Multimodal context (message groups + enrichments)      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  DOMAIN LAYER (Pure Go, DDD)                              │  │
│  │  - 30 aggregates (3 bounded contexts)                     │  │
│  │  - CRM: Contact, Session, Message, Channel, Pipeline...   │  │
│  │  - Automation: Campaign, Sequence, Broadcast              │  │
│  │  - Core: Project, Billing, User                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  MEMORY SERVICE (Embedded)                                │  │
│  │  - pgvector search (RRF + Cross-Encoder Reranking)        │  │
│  │  - Hybrid search (vector + BM25 + graph)                  │  │
│  │  - Knowledge graph (Neo4j-like em PostgreSQL)             │  │
│  │  - Fact extraction e consolidation                        │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  INFRASTRUCTURE                                            │  │
│  │  - PostgreSQL 15+ (RLS multi-tenancy)                     │  │
│  │  - RabbitMQ 3.12+ (event bus)                             │  │
│  │  - Redis 7.0+ (cache - 80% implementado)                  │  │
│  │  - Temporal (workflows, sagas)                            │  │
│  │  - S3 (media storage)                                      │  │
│  └──────────────────────────────────────────────────────────┘  │
└───────────────────────────────┬───────────────────────────────────┘
                                │
                                │ gRPC (bi-directional)
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PYTHON ADK (Ventros AI)                     │
│                     Microserviço separado (Python)              │
│                                                                   │
│  - Multi-agent orchestration (6+ agents in chain)               │
│  - LLM providers (OpenAI, Anthropic, Groq, Gemini)              │
│  - Agent memory (RAM durante execução)                          │
│  - Tool execution (MCP client)                                   │
│  - Response generation                                           │
│                                                                   │
│  Integração:                                                     │
│  1. Go → Python: ExecuteAgent(type, context)                    │
│  2. Python → Go (MCP): call_tool(tool_name, args)               │
│  3. Python → Go (Memory): GetContactContext(), SearchMemory()   │
│  4. Python retorna response → Go persiste                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📊 STATUS ATUAL (2025-10-15)

### ✅ Implementado (80%)

#### **Backend Core (100%)**
- ✅ Domain layer: 30 aggregates, 182+ eventos
- ✅ Application layer: 80+ commands, 20+ queries (CQRS)
- ✅ Infrastructure: Repositories, HTTP handlers, messaging
- ✅ Outbox Pattern: <100ms event latency
- ✅ Multi-tenancy: RLS (Row-Level Security)
- ✅ Optimistic Locking: 16/30 aggregates (53%)

#### **API (95%)**
- ✅ 158 endpoints REST (Gin)
- ✅ WebSocket (real-time messaging)
- ✅ Swagger docs (auto-generated)
- ✅ JWT Auth (Keycloak integration)
- ⚠️ RBAC: 5% dos endpoints (95% missing - P0!)

#### **Channels (90%)**
- ✅ WhatsApp (WAHA integration - 100%)
- ✅ Message enrichment (imagem, áudio, vídeo, documento)
- ✅ History import (180 days, batch processing)
- ⚠️ Instagram/Facebook: 0% (planned)

#### **Message Enrichment (100%)**
- ✅ Audio → Text (Groq Whisper 216x real-time, FREE)
- ✅ Image → OCR (Vertex Vision Gemini 1.5 Flash)
- ✅ Document → Markdown (LlamaParse)
- ✅ Profile picture scoring (0-10)
- ✅ Automatic provider routing + fallbacks

#### **MCP Server (100%)**
- ✅ 30+ tools para Python ADK
- ✅ CRM operations (contacts, sessions, messages, etc)
- ✅ Document search (vector + keyword + filters)
- ✅ Multimodal context (message groups + enrichments)
- ✅ BI & analytics tools

### ⏳ Em Progresso (15%)

#### **Memory Service (20%)**
- ✅ Schema definido (memory_embeddings, memory_facts, memory_graph)
- ✅ RRF + Cross-Encoder Reranking (moderno 2024-2025)
- ⚠️ Vector search: 0% implementado
- ⚠️ Hybrid search: 0% implementado
- ⚠️ Knowledge graph: 0% implementado
- ⚠️ Fact extraction: 0% implementado

#### **Python ADK (60%)**
- ✅ Multi-agent system (6+ agents in chain)
- ✅ MCP client (calls Go tools)
- ✅ LLM providers integration
- ⚠️ Agent memory persistence: Apenas RAM (não persiste)
- ⚠️ gRPC interface: 0% implementado

#### **Cache (20%)**
- ✅ Redis configured
- ⚠️ Cache integration: 0% dos endpoints

### ❌ Pendente (5%)

#### **Security (P0 - CRÍTICO!)**
- ❌ Dev mode auth bypass (CVSS 9.1) - produção vulnerável!
- ❌ SSRF in webhooks (CVSS 9.1) - sem validação de URL
- ❌ BOLA in 60 GET endpoints (CVSS 8.2) - sem ownership checks
- ❌ Resource exhaustion (CVSS 7.5) - sem max page size
- ❌ RBAC missing (CVSS 7.1) - 95 endpoints sem role checks

#### **AI/ML (60% missing)**
- ❌ Vector search (pgvector setup, mas sem uso)
- ❌ Hybrid search (não implementado)
- ❌ Knowledge graph (schema pronto, mas sem uso)
- ❌ Memory facts extraction (não implementado)
- ❌ Python ADK persistence (tudo em RAM)

#### **Channels Extras**
- ❌ Instagram Direct (0%)
- ❌ Facebook Messenger (0%)
- ❌ Email (0%)
- ❌ SMS (0%)

---

## 🎯 BOUNDED CONTEXTS (DDD)

### 1. **CRM** (23 aggregates)

**Core**:
- Contact (lead management, profile, custom fields)
- Session (conversation tracking, timeout, consolidation)
- Message (multi-channel, enrichment, media)
- Channel (WAHA, Instagram, Facebook, etc)

**Organização**:
- Pipeline (stages, automations)
- Agent (human + AI agents)
- Chat (group conversations)
- ContactList (static + dynamic)

**Tracking**:
- ContactEvent (timeline de eventos)
- Note (annotations)
- Tag (categorization)
- Tracking (ad attribution)

**Webhooks**:
- WebhookSubscription (event subscriptions)

**Outros 10 aggregates** (ver DEV_GUIDE.md)

### 2. **Automation** (3 aggregates)

- Campaign (scheduled messages)
- Sequence (drip campaigns)
- Broadcast (one-time bulk sends)

### 3. **Core** (4 aggregates)

- Project (multi-tenancy)
- Billing (Stripe integration)
- User (authentication, RBAC)
- Saga (orchestration)

---

## 📈 MÉTRICAS

### **Codebase**
- **Total Lines**: ~150,000 (Go + Python + SQL)
- **Go Code**: ~80,000 lines
- **Test Coverage**: 82% (goal: 85%+)
- **Domain Layer**: 100% coverage
- **Application Layer**: 80% coverage
- **Infrastructure**: 60% coverage

### **API**
- **Endpoints**: 158
- **CRM**: 77 endpoints
- **Automation**: 28 endpoints
- **Billing**: 8 endpoints
- **Auth**: 4 endpoints
- **Health/Metrics**: 2 endpoints

### **Database**
- **Tables**: 40+
- **Aggregates**: 30
- **Events**: 182+ domain events
- **Multi-tenancy**: 100% (RLS em todas as tabelas)

### **Performance**
- **API Latency**: <200ms (p95)
- **Event Latency**: <100ms (Outbox Pattern)
- **Query Latency**: <500ms (complex queries)
- **Cache Hit Rate**: N/A (não implementado)

### **AI Enrichment**
- **Audio Transcription**: 216x real-time (Groq Whisper)
- **Image OCR**: $0.00025/image (Vertex Vision)
- **Document Parsing**: $1-3 per 1000 pages (LlamaParse)
- **Profile Picture Scoring**: $0.00025/image (Gemini Vision)

---

## 🚀 DEPLOYMENT

### **Containers**

```yaml
# docker-compose.yml
services:
  ventros-crm:
    image: ventros-crm:latest
    ports:
      - "8080:8080"  # Main API
      - "8081:8081"  # MCP Server (if MCP_ENABLED=true)
    environment:
      - DB_HOST=postgres
      - REDIS_URL=redis://redis:6379
      - RABBITMQ_URL=amqp://rabbitmq:5672
      - MCP_ENABLED=true
      - ENV=production  # IMPORTANTE: Desabilita dev auth bypass
    depends_on:
      - postgres
      - redis
      - rabbitmq
      - temporal

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=ventros_crm
      - POSTGRES_USER=ventros
      - POSTGRES_PASSWORD=***

  redis:
    image: redis:7-alpine

  rabbitmq:
    image: rabbitmq:3.12-management

  temporal:
    image: temporalio/auto-setup:latest

  python-adk:
    image: ventros-ai:latest
    ports:
      - "50051:50051"  # gRPC (futuro)
    environment:
      - VENTROS_CRM_MCP_URL=http://ventros-crm:8081
      - VENTROS_CRM_JWT_SECRET=***
```

### **Environment Variables**

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ventros_crm
DB_USER=ventros
DB_PASSWORD=***

# WAHA (WhatsApp)
WAHA_BASE_URL=https://waha.ventros.cloud
WAHA_API_KEY=***

# AI Providers
GROQ_API_KEY=gsk_***           # Whisper (audio) - FREE
VERTEX_PROJECT_ID=***          # Gemini Vision (images)
LLAMAPARSE_API_KEY=llx_***     # Documents
OPENAI_API_KEY=sk-***          # Whisper fallback

# JWT & Auth
JWT_SECRET=***
KEYCLOAK_URL=https://auth.ventros.cloud

# MCP Server
MCP_ENABLED=true               # Enable MCP server
MCP_PORT=8081                  # MCP server port
MCP_JWT_SECRET=***             # JWT for Python ADK

# Environment
ENV=production                 # CRITICAL: Disables dev auth bypass
```

### **Commands**

```bash
# Build
make build

# Run locally (requires: make infra)
make api

# Tests
make test                    # All tests
make test-unit               # Unit only
make test-integration        # Integration (requires infra)
make test-e2e               # E2E (requires infra + API)

# Coverage
make test-coverage

# Database
make migrate-auto            # GORM AutoMigrate (DEV ONLY)
make migrate-up              # SQL migrations (PRODUCTION)
make migrate-down            # Rollback

# Clean
make clean                   # Remove all data (DESTRUCTIVE)
make reset-full              # Clean + migrate + API
```

---

## 🔗 INTEGRAÇÃO COM OUTROS SERVIÇOS

### **Python ADK ↔ Ventros CRM**

```
FLUXO COMPLETO:

1. 🟢 Cliente envia mensagem no WhatsApp
   ↓
2. 🟢 WAHA webhook → Go CRM
   ↓
3. 🟢 Go CRM:
   - Persiste message no PostgreSQL
   - Dispara enrichment (áudio → texto, etc)
   - Publica event: message.created
   ↓
4. 🟢 Go CRM decide: "Precisa de AI Agent?"
   ↓
5. 🐍 Go → Python ADK (gRPC - futuro):
   ExecuteAgent(type="CustomerServiceAgent", context={...})
   ↓
6. 🐍 Python ADK (MULTI-AGENT):
   - CustomerServiceAgent chama LeadQualifierAgent
   - LeadQualifierAgent chama PricingAgent
   - PricingAgent chama ProposalAgent
   - ProposalAgent chama ApprovalAgent
   - ApprovalAgent chama ResponseGeneratorAgent
   - Total: 6 agentes em cadeia (5-10s)
   ↓
7. 🐍 Durante execução, Python chama Go (MCP):
   - call_tool("get_contact", {contact_id: "..."})
   - call_tool("search_documents", {query: "contrato"})
   - call_tool("get_message_group", {group_id: "..."})
   ↓
8. 🐍 Python retorna response → Go CRM
   ↓
9. 🟢 Go CRM:
   - Persiste response
   - Envia mensagem via WAHA
   - Publica event: message.sent
```

### **Memory Service (Embedded no CRM)**

```
# Python ADK usa Memory Service via MCP:
docs = mcp_client.call_tool("search_documents", {
    "query": "valor do contrato com Company A",
    "contact_id": "contact-uuid",
    "limit": 5
})

# Go CRM executa:
# 1. Generate query embedding (Vertex AI)
# 2. pgvector search (cosine similarity)
# 3. RRF fusion (vector + keyword + graph)
# 4. Cross-Encoder reranking (BAAI/bge-reranker-v2-m3)
# 5. Return top-K results
```

---

## 📚 DOCUMENTAÇÃO

**Raiz do projeto**:
- `README.md` - Overview e quick start
- `CLAUDE.md` - Instruções para Claude Code
- `DEV_GUIDE.md` - Guia completo de desenvolvimento (1,536 lines)
- `TODO.md` - Roadmap e prioridades
- `AI_REPORT.md` - Audit arquitetural (8.0/10)
- `MAKEFILE.md` - Referência de comandos
- `ORGANIZATION_RULES.md` - Regras de organização do projeto

**Planning**:
- `planning/ventros-crm/` - Projeto atual (este documento)
- `planning/ventros-ai/` - Python ADK
- `planning/memory-service/` - Memory Service
- `planning/mcp-server/` - MCP Server

**Guides**:
- `guides/domain_mapping/` - 30 aggregate docs
- `guides/TESTING.md` - Estratégia de testes

---

## 🎯 PRÓXIMOS PASSOS (ver TODO.md)

### **Sprint 0 - Security P0 (URGENTE!)**
- Fix dev mode auth bypass
- Fix SSRF in webhooks
- Add BOLA checks (60 endpoints)
- Add max page size (19 queries)
- Implement RBAC (95 endpoints)

### **Sprint 1 - Memory Service**
- Implement vector search (pgvector)
- Implement hybrid search (RRF + reranker)
- Implement knowledge graph
- Implement fact extraction

### **Sprint 2 - Python ADK**
- Implement gRPC interface (Go ↔ Python)
- Add agent memory persistence
- Add agent execution history
- Add agent performance metrics

### **Sprint 3 - Cache**
- Integrate Redis cache (70+ endpoints)
- Add cache invalidation strategy
- Add cache metrics

### **Sprint 4 - Channels**
- Implement Instagram Direct
- Implement Facebook Messenger
- Implement Email
- Implement SMS

---

## 🏆 SCORE ARQUITETURAL

**Overall**: 8.0/10 (Production-ready backend, AI tem gaps)

**Breakdown**:
- **Domain Layer**: 9.5/10 (Excelente DDD)
- **Application Layer**: 9.0/10 (CQRS bem implementado)
- **Infrastructure**: 8.5/10 (Solid, mas cache missing)
- **Security**: 6.0/10 (5 P0 vulnerabilities!)
- **AI/ML**: 6.0/10 (Message enrichment 100%, memory 20%)
- **Tests**: 8.5/10 (82% coverage)
- **Documentation**: 9.0/10 (Extensive)

---

**Version**: 1.0
**Last Updated**: 2025-10-15
**Maintainer**: Ventros CRM Team
