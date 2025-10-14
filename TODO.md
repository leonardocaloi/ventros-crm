# TODO - Ventros CRM

## 📋 CONSOLIDATED TODO - Based on Complete Architectural Report

**Last Update**: 2025-10-13 (After complete 200k+ lines architectural audit)
**Report Reference**: `AI_REPORT_PART*.md` (6 parts, 30 evaluation tables)
**Build Status**: ✅ SUCCESS (0 errors, 0 warnings)
**Test Status**: ✅ 100% tests passing

---

## 🎯 EXECUTIVE SUMMARY

**Overall Scores**:
- Backend Go: **8.0/10** (B+) - Production-Ready
- Database: **9.2/10** (A) - Excellent
- API Security: **6.0/10** (C+) - **5 P0 Critical Vulnerabilities**
- AI/ML: **2.5/10** (F) - Only enrichment working, 80% missing
- Overall System: **5.3/10** (C) - Backend solid, AI critical gaps

**Critical Findings**:
1. ✅ **Chat aggregate 100% implemented** (contradicts previous docs)
2. ✅ **30 aggregates identified** (not 23 as documented)
3. ✅ **158 endpoints catalogued** (not "50+")
4. ✅ **182 domain events** (100% consistent pattern)
5. ❌ **5 security vulnerabilities P0** (SSRF CVSS 9.1, Dev Bypass CVSS 9.1, BOLA CVSS 8.2)
6. ❌ **Memory Service 80% missing** (vector DB, hybrid search, facts)
7. ❌ **0% cache integration** (Redis configured but not used)
8. ❌ **Python ADK 0%** (multi-agent not started)

---

## 🔴 PRIORITY 0: CRITICAL & URGENT (0-4 weeks)

### **SPRINT 1-2: Security Fixes** (3-4 weeks) - BLOCKER FOR PRODUCTION

**Reference**: `AI_REPORT_PART4.md` - Table 18 (OWASP Top 10)

#### 1.1. 🚨 **Dev Mode Bypass** (1 day) - CVSS 9.1 CRITICAL
**Location**: `infrastructure/http/middleware/auth.go:41`

**Vulnerability**:
```go
// ❌ CRITICAL: Dev mode allows authentication bypass via header
if a.devMode {
    if authCtx := a.handleDevAuth(c); authCtx != nil {
        c.Set("auth", authCtx)
        return // Bypass authentication!
    }
}
```

**Exploit**:
```bash
curl -H "X-Dev-User-ID: any-uuid" \
     -H "X-Dev-Tenant-ID: victim-tenant" \
     https://api.ventros.ai/api/v1/crm/contacts
# Response: 200 OK with ALL contacts ❌
```

**Fix**:
- [ ] Disable dev mode in production (panic if enabled)
- [ ] Add IP whitelist for dev mode
- [ ] Add environment check (GO_ENV=production)
- [ ] Tests
- [ ] Deploy urgently

**Effort**: 1 day

---

#### 1.2. 🚨 **SSRF in Webhooks** (3 days) - CVSS 9.1 CRITICAL
**Location**: `internal/domain/crm/webhook/webhook_subscription.go:36`

**Vulnerability**:
```go
// ❌ SSRF: Can access AWS metadata, internal services
func NewWebhookSubscription(url string, events []string) (*WebhookSubscription, error) {
    if url == "" {
        return nil, ErrInvalidURL
    }
    return &WebhookSubscription{URL: url}, nil // No validation!
}
```

**Exploit**:
```bash
curl -X POST /api/v1/webhooks -d '{
  "url": "http://169.254.169.254/latest/meta-data/iam/security-credentials/",
  "events": ["contact.created"]
}'
# Server fetches AWS credentials ❌
```

**Fix**:
- [ ] URL validation (scheme, host, IP)
- [ ] Block private IPs (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8)
- [ ] Block cloud metadata (169.254.169.254)
- [ ] Whitelist schemes (HTTPS only)
- [ ] Tests
- [ ] Helper: `isPrivateIP()`, `isCloudMetadata()`

**Effort**: 3 days

---

#### 1.3. 🔴 **BOLA in 60 GET Endpoints** (1 week) - CVSS 8.2 HIGH
**Location**: ~60 handlers in `infrastructure/http/handlers/*_handler.go`

**Vulnerability**:
```go
// ❌ NO ownership check - any authenticated user can access ANY contact
func (h *ContactHandler) GetContact(c *gin.Context) {
    contactID := c.Param("id")
    domainContact, _ := h.contactRepo.FindByID(ctx, contactID)
    c.JSON(200, h.mapper.ToDTO(domainContact)) // ❌ Leak!
}
```

**Exploit**: Attacker (tenant A) accessing victim (tenant B) data

**Fix**:
- [ ] Add ownership check in 60 GET endpoints:
  - `GET /contacts/:id`
  - `GET /messages/:id`
  - `GET /sessions/:id`
  - `GET /campaigns/:id`
  - `GET /pipelines/:id`
  - `GET /chats/:id`
  - ... (54 more)
- [ ] Helper function: `checkOwnership(entity.TenantID, authCtx.TenantID)`
- [ ] Return 404 (not 403) to avoid info leak
- [ ] Tests for each endpoint

**Effort**: 1 semana (60 handlers × 10 min each + tests)

---

#### 1.4. 🔴 **Resource Exhaustion** (3 days) - CVSS 7.5 HIGH
**Locations**:
- `internal/application/queries/list_contacts_query.go:67`
- All 19 query handlers

**Vulnerability**:
```go
// ❌ NO max limit - can request 1M contacts
func (q *ListContactsQuery) Execute(ctx context.Context, page, limit int) ([]ContactDTO, error) {
    offset := (page - 1) * limit
    contacts := q.db.Offset(offset).Limit(limit).Find(&contacts)
    return contacts, nil
}
```

**Exploit**:
```bash
curl "/api/v1/crm/contacts?page=1&limit=1000000"
# Server: OutOfMemory ❌
```

**Fix**:
- [ ] Enforce `MaxPageSize = 100` in all queries
- [ ] Add query timeouts (10s via context)
- [ ] Add max payload size (10MB)
- [ ] Tests

**Effort**: 3 dias (19 queries)

---

#### 1.5. 🔴 **RBAC Missing in 95 Endpoints** (2 weeks) - CVSS 7.1 HIGH
**Location**: `infrastructure/http/routes/routes.go`

**Vulnerability**:
```go
// ❌ NO RBAC - any authenticated user can delete
contactRoutes.DELETE("/:id", contactHandler.DeleteContact)

// ✅ Should be:
contactRoutes.DELETE("/:id",
    rbac.Authorize("admin", "agent"),
    contactHandler.DeleteContact)
```

**Missing RBAC** (95 endpoints):
- DELETE operations → admin only
- POST /campaigns → agent+
- PUT /pipelines → admin only
- DELETE /automations → admin only

**Fix**:
- [ ] Apply RBAC middleware to 95 endpoints
- [ ] Define permission matrix (admin, agent, viewer)
- [ ] Tests for each role

**Effort**: 2 semanas

---

#### 1.6. 🔴 **Rate Limiting (Redis)** (1 week) - P0
**Location**: `infrastructure/http/middleware/rate_limiter.go`

**Current**: In-memory (não escalável, fácil bypass)

**Fix**:
- [ ] Redis integration
- [ ] Sliding window counter
- [ ] Per-user/tenant limits
- [ ] Different limits per endpoint group:
  - Auth: 5 req/min
  - CRM read: 100 req/min
  - CRM write: 20 req/min
  - Webhooks: 10 req/min
  - AI: 10 req/min
- [ ] Tests

**Effort**: 1 semana

**Total Effort Sprint 1-2**: 3-4 semanas

---

### **SPRINT 3-4: Cache Layer + Performance** (2 weeks) - P0

**Reference**: `AI_REPORT_PART3.md` - Table 13 (Queries & Performance)

#### 2.1. 🔴 **Redis Cache Integration** (1 week)
**Status**: Redis client configured but **0% integrated**

**Tasks**:
- [ ] Cache middleware
- [ ] Cache 5 priority queries:
  1. `GetContactStatsQuery` (TTL: 5 min) - currently 423ms avg
  2. `SessionAnalyticsQuery` (TTL: 30 min) - currently 678ms avg
  3. `ListContactsQuery` (TTL: 2 min)
  4. `MessageHistoryQuery` (TTL: 1 min)
  5. `GetActiveSessionsQuery` (TTL: 30 sec)
- [ ] Cache invalidation via events:
  - contact.* → invalidate contact_stats:*
  - session.closed → invalidate session_analytics:*
- [ ] Cache hit rate monitoring (target: >70%)
- [ ] Tests

**Expected Result**: Queries <200ms (from 423ms-678ms)

**Effort**: 1 semana

---

#### 2.2. 🔴 **Fix N+1 Query Bug** (2 days)
**Location**: `infrastructure/persistence/gorm_contact_list_repository.go:234`

**Bug**:
```go
// Query 1: Get IDs
contactIDs := db.Table("contact_list_memberships").
    Where("list_id = ?", listID).
    Pluck("contact_id", &ids)

// Query 2+: N queries for each contact (N+1)
for _, id := range ids {
    contact := db.First(&Contact{}, id) // ❌ N queries!
    contacts = append(contacts, contact)
}
```

**Fix**: Single JOIN query
```go
contacts := db.Joins("JOIN contact_list_memberships ON ...").
    Where("list_id = ?", listID).
    Preload("Tags").
    Preload("CustomFields").
    Find(&contacts)
```

**Impact**: 100 contacts = 100 queries → 1 query (100x faster)

**Tasks**:
- [ ] Fix GetContactsInListQuery (single JOIN)
- [ ] Verify ConversationThreadQuery (possible N+1)
- [ ] Tests

**Effort**: 2 dias

---

#### 2.3. 🔴 **Materialized View for Analytics** (3 days)
**Location**: `internal/application/queries/session_analytics_query.go`

**Problem**: `SessionAnalyticsQuery` muito lento (678ms avg, 1200ms p95)

**Fix**:
- [ ] Migration: `session_analytics_mv` (materialized view)
- [ ] Refresh strategy (hourly via cron)
- [ ] Query rewrite to use MV
- [ ] Performance tests (<100ms)

**Effort**: 3 dias

**Total Effort Sprint 3-4**: 2 semanas

---

## 🟡 PRIORITY 1: IMPORTANT (4-12 weeks)

### **SPRINT 5-11: Memory Service Foundation** (7 weeks) - P0 for AI

**Reference**: `AI_REPORT_PART5.md` - Table 21 (AI/ML Components)

#### 3.1. **pgvector + Vector Search** (1 week)
**Status**: ❌ 0% implemented

**Tasks**:
- [ ] Install pgvector extension
- [ ] Migration 000050: `memory_embeddings` table
  ```sql
  CREATE TABLE memory_embeddings (
      id UUID PRIMARY KEY,
      tenant_id UUID NOT NULL,
      contact_id UUID,
      session_id UUID,
      message_id UUID,
      content_type VARCHAR(50) NOT NULL,
      content_text TEXT NOT NULL,
      embedding vector(768), -- text-embedding-005
      metadata JSONB,
      created_at TIMESTAMP
  );
  CREATE INDEX ON memory_embeddings USING hnsw (embedding vector_cosine_ops);
  ```
- [ ] `MemoryEmbeddingRepository`
- [ ] `VectorSearchService` (cosine similarity, top-K)
- [ ] Embedding worker (RabbitMQ consumer)
- [ ] Vertex AI integration (text-embedding-005)
- [ ] Tests + benchmarks (<100ms)

**Effort**: 1 semana

---

#### 3.2. **Hybrid Search Service** (2 weeks)
**Status**: ❌ 0% implemented

**Tasks**:
- [ ] `HybridSearchService`
- [ ] 4 search strategies:
  1. SQL BASELINE (last 20 messages)
  2. VECTOR SEARCH (50% weight)
  3. KEYWORD SEARCH (20% weight) - pg_trgm
  4. GRAPH TRAVERSAL (20% weight) - Apache AGE
- [ ] RRF (Reciprocal Rank Fusion)
- [ ] Optional reranking (Jina v2)
- [ ] Tests (<500ms latency, >90% recall)

**Effort**: 2 semanas

---

#### 3.3. **Memory Facts Extraction** (2 weeks)
**Status**: ❌ 0% implemented

**Tasks**:
- [ ] Migration 000051: `memory_facts` table
  ```sql
  CREATE TABLE memory_facts (
      id UUID PRIMARY KEY,
      tenant_id UUID NOT NULL,
      contact_id UUID NOT NULL,
      fact_type VARCHAR(50) NOT NULL, -- budget, preference, objection
      fact_text TEXT NOT NULL,
      confidence FLOAT NOT NULL,
      source_message_id UUID,
      extracted_at TIMESTAMP
  );
  ```
- [ ] `FactExtractionService`
- [ ] LLM-based extraction (Gemini Flash)
- [ ] Fact types: budget, preference, objection, pain_point, timeline
- [ ] RabbitMQ consumer (message.created)
- [ ] Tests

**Example Facts**:
- "Cliente tem budget de R$ 5.000/mês" (confidence: 0.92)
- "Pain point: ROI não está claro" (confidence: 0.85)
- "Objeção: Concorrente 30% mais barato" (confidence: 0.78)

**Effort**: 2 semanas

---

#### 3.4. **AI Cost Tracking** (1 week)
**Status**: ❌ 0% implemented - **CRITICAL for billing**

**Tasks**:
- [ ] Migration: `ai_costs` table
  ```sql
  CREATE TABLE ai_costs (
      id UUID PRIMARY KEY,
      tenant_id UUID NOT NULL,
      provider VARCHAR(50) NOT NULL,
      model VARCHAR(100) NOT NULL,
      units DECIMAL(10,2) NOT NULL,
      unit_cost DECIMAL(10,6) NOT NULL,
      total_cost DECIMAL(10,6) NOT NULL,
      timestamp TIMESTAMP NOT NULL
  );
  ```
- [ ] `CostTracker` service
- [ ] Integrate in all AI providers:
  - Vertex Vision ($0.0025/image)
  - Groq Whisper (FREE)
  - OpenAI Whisper ($0.006/min)
  - LlamaParse ($0.003/page)
  - Gemini Flash ($0.35/1M tokens)
- [ ] Dashboard query
- [ ] Budget alerts
- [ ] Tests

**Effort**: 1 semana

---

#### 3.5. **Retrieval Strategies Dictionary** (3 days)
**Status**: ❌ 0% implemented

**Tasks**:
- [ ] Migration 000052: `retrieval_strategies` table
- [ ] Strategy configs per agent category:
  ```go
  "sales_prospecting": {
      VectorWeight:  0.20,
      KeywordWeight: 0.30,
      GraphWeight:   0.40, // HIGH: campaign attribution
  }
  ```
- [ ] Dynamic weight adjustment
- [ ] A/B testing framework
- [ ] Tests

**Effort**: 3 dias

**Total Effort Sprint 5-11**: 7 semanas

---

### **SPRINT 12-14: gRPC API** (3 weeks) - P0 for Python ADK

**Reference**: `AI_REPORT_PART5.md` - Table 25 (gRPC API)

#### 4.1. **Proto Definitions** (3 days)
**Status**: ❌ Directory `api/proto/` doesn't exist

**Tasks**:
- [ ] `api/proto/memory_service.proto`
  - SearchMemory
  - StoreEmbedding
  - ExtractFacts
  - GetContactFacts
- [ ] `api/proto/crm_service.proto` (partial)
- [ ] Generate Go code (protoc-gen-go)
- [ ] Generate Python code (protoc-gen-python)

**Effort**: 3 dias

---

#### 4.2. **Go gRPC Server** (1 week)
**Status**: ❌ 0% implemented

**Tasks**:
- [ ] `infrastructure/grpc/memory_service_server.go`
- [ ] `cmd/grpc-server/main.go`
- [ ] Authentication (JWT)
- [ ] MemoryService implementation
- [ ] Tests

**Effort**: 1 semana

---

#### 4.3. **Python gRPC Client** (3 days)
**Status**: ❌ 0% implemented

**Tasks**:
- [ ] `python-adk/ventros_adk/memory_client.py`
- [ ] Connection pooling
- [ ] Retry logic
- [ ] Tests

**Effort**: 3 dias

---

#### 4.4. **gRPC Interceptors** (3 days)
**Tasks**:
- [ ] Logging interceptor
- [ ] Metrics interceptor (Prometheus)
- [ ] Error handling
- [ ] Tests

**Effort**: 3 dias

**Total Effort Sprint 12-14**: 3 semanas

---

### **SPRINT 15-18: MCP Server** (4 weeks) - P0 for Claude Desktop

**Reference**: `AI_REPORT_PART6.md` - Table 26 (MCP Server)

**Status**: Docs complete (1,175 lines), code 0%

#### 5.1. **MCP Server Setup** (1 week)
**Tasks**:
- [ ] `infrastructure/mcp/server.go` (HTTP server)
- [ ] `infrastructure/mcp/tool_registry.go`
- [ ] Authentication (JWT + API Keys)
- [ ] MCP protocol endpoints:
  - GET /mcp/tools (discovery)
  - POST /mcp/tools/:tool (execution)
  - GET /mcp/tools/:tool/stream (SSE)
- [ ] Tests

**Effort**: 1 semana

---

#### 5.2. **Priority MCP Tools** (2 weeks)
**25 tools prioritários**:

**BI Tools (7)**:
- [ ] get_leads_count
- [ ] get_conversion_rate
- [ ] get_agent_performance
- [ ] get_top_performing_agent
- [ ] get_campaign_metrics
- [ ] get_churn_prediction
- [ ] get_revenue_forecast

**CRM Operations Tools (8)**:
- [ ] qualify_lead
- [ ] update_pipeline_stage
- [ ] assign_to_agent
- [ ] create_note
- [ ] schedule_follow_up
- [ ] send_message
- [ ] tag_contact
- [ ] export_contacts

**Memory Tools (5)**:
- [ ] search_memory
- [ ] get_contact_context
- [ ] get_contact_facts
- [ ] get_session_summary
- [ ] find_similar_contacts

**Document Tools (5)**:
- [ ] search_documents
- [ ] get_document_chunks
- [ ] upload_document
- [ ] summarize_document
- [ ] answer_from_documents

**Effort**: 2 semanas (25 tools × 2h each)

---

#### 5.3. **Claude Desktop Integration** (1 week)
**Tasks**:
- [ ] Config file: `claude_desktop_config.json`
- [ ] Test all 25 tools
- [ ] Documentation
- [ ] Example prompts

**Effort**: 1 semana

**Total Effort Sprint 15-18**: 4 semanas

---

### **SPRINT 19-24: Python ADK Multi-Agent** (6 weeks) - P0

**Reference**: `AI_REPORT_PART5.md` - Table 24 (Python ADK)

**Status**: Docs complete (3,000+ lines), code 0%

#### 6.1. **Project Setup + Semantic Router** (1 week)
**Tasks**:
- [ ] Poetry setup
- [ ] Google Cloud ADK 0.5+
- [ ] DistilBERT fine-tuning (intent classification)
- [ ] Training data (10k messages via synthetic)
- [ ] 5 intent classes: sales_prospecting, retention_churn, support_technical, support_billing, general
- [ ] Tests (>92% accuracy)

**Effort**: 1 semana

---

#### 6.2. **CoordinatorAgent** (1 week)
**Tasks**:
- [ ] Semantic Router integration
- [ ] Agent dispatch logic
- [ ] Fallback strategy (BalancedAgent)
- [ ] Tests

**Effort**: 1 semana

---

#### 6.3. **Specialist Agents** (2 weeks)
**5 agents**:
- [ ] SalesProspectingAgent (lead qualification, pipeline updates)
- [ ] RetentionChurnAgent (churn prediction, win-back)
- [ ] SupportTechnicalAgent (KB search, escalation)
- [ ] SupportBillingAgent (invoices, subscriptions)
- [ ] BalancedAgent (fallback general-purpose)

**Each agent includes**:
- System prompt
- Memory service integration (gRPC)
- Tool registry (30 tools)
- Tests

**Effort**: 2 semanas

---

#### 6.4. **Tool Registry + RabbitMQ** (1 week)
**Tasks**:
- [ ] 30 tools wrapped
- [ ] gRPC calls to Go backend
- [ ] Error handling
- [ ] RabbitMQ consumer (message.created)
- [ ] RabbitMQ publisher (agent response)
- [ ] Tests

**Effort**: 1 semana

---

#### 6.5. **Temporal + Observability** (1 week)
**Tasks**:
- [ ] Temporal workflow definitions (long-running tasks)
- [ ] Phoenix observability (tracing)
- [ ] Dashboard
- [ ] Alerts
- [ ] Tests

**Effort**: 1 semana

**Total Effort Sprint 19-24**: 6 semanas

---

### **Additional P1 Tasks** (parallel with AI work)

#### 7. **Testing Coverage** (2 weeks)
**Reference**: `AI_REPORT_PART5.md` - Table 22

**Current**: 82% declared, but real ~45% (application layer low)

**Tasks**:
- [ ] 40+ unit tests (19 use cases without tests)
- [ ] 10 integration tests (only 2 exist, need 20% pyramid)
- [ ] 5 E2E tests (campaign, sequence, memory, billing)
- [ ] Target: 85% overall

**Effort**: 2 semanas

---

#### 8. **Resilience Patterns** (2 weeks)
**Reference**: `AI_REPORT_PART5.md` - Table 23

**Current**: Circuit breaker exists but only 10% coverage

**Tasks**:
- [ ] Circuit breaker in 4 external APIs:
  - Vertex AI
  - Stripe
  - WAHA
  - LlamaParse
- [ ] Retry logic (exponential backoff)
- [ ] Timeout in all APIs (10s default)
- [ ] Bulkhead pattern (optional P2)
- [ ] Tests

**Effort**: 2 semanas

---

#### 9. **Optimistic Locking** (1 week)
**Reference**: `AI_REPORT_PART1.md` - Table 5

**Current**: Only 16/30 aggregates (53%)

**Missing (14 aggregates)**:
- [ ] MessageGroup
- [ ] Note
- [ ] Tracking
- [ ] UsageMeter (billing critical!)
- [ ] WebhookSubscription
- [ ] SagaTracker
- [ ] MessageEnrichment
- [ ] ChannelType
- [ ] DomainEventLog
- [ ] OutboxEvent
- [ ] ContactEvent
- [ ] CustomField
- [ ] (2 more)

**Tasks**:
- [ ] Add `version int` field to 14 aggregates
- [ ] Migration for each
- [ ] Update repository Update() methods
- [ ] Retry logic on optimistic lock failure
- [ ] Tests

**Effort**: 1 semana

---

## 🟢 PRIORITY 2: IMPROVEMENTS (12+ weeks)

### 10. **Value Objects** (2 weeks)
**Reference**: `AI_REPORT_PART2.md` - Table 6

**Current**: Only 12 VOs, many primitive obsession cases

**Critical**:
- [ ] **Money VO** (P0): Use int64 (cents) instead of float64 for billing
  - Impact: Financial accuracy
  - Locations: Invoice, Subscription, UsageMeter
  - Effort: 1 semana

**Important (P1)**:
- [ ] Email VO (RFC 5322 validation)
- [ ] PhoneNumber VO (E.164 validation)
- [ ] URL VO (SSRF prevention)
- [ ] Duration VO (type safety)
- [ ] Timezone VO (IANA validation)
- [ ] CustomField VO (make immutable)
- [ ] ChannelConfig VO (strong typing instead of map)

**Effort**: 2 semanas total

---

### 11. **Knowledge Graph (Apache AGE)** (2 weeks)
**Reference**: `AI_REPORT_PART5.md` - Table 21

**Tasks**:
- [ ] Install Apache AGE extension
- [ ] Graph schema:
  - Nodes: Contact, Session, Message, Offer, Campaign
  - Edges: HAS_SESSION, CONTAINS, RECEIVED_OFFER, IN_CAMPAIGN, REPLY_TO
- [ ] Cypher queries (graph traversal)
- [ ] Find similar contacts
- [ ] Tests

**Effort**: 2 semanas

---

### 12. **Observability** (2 weeks)
**Reference**: `AI_REPORT_PART1.md` - Table 1 (score 5.5/10)

**Tasks**:
- [ ] Prometheus metrics
- [ ] OpenTelemetry tracing
- [ ] Grafana dashboards
- [ ] Alerts (error rate, latency, costs)
- [ ] APM (New Relic/Datadog - optional)

**Effort**: 2 semanas

---

### 13. **Agent Templates** (1 week)
**Reference**: `AI_REPORT_PART6.md` - Roadmap Sprint 30

**Tasks**:
- [ ] 10 system agent templates registry
- [ ] Template discovery API
- [ ] Template instantiation
- [ ] Documentation

**Effort**: 1 semana

---

### 14. **Chat Entity** (1 week)
**Status**: ✅ Fully implemented (discovery in report)
**Location**: `internal/domain/crm/chat/chat.go`

**Validation tasks**:
- [ ] Verify Chat tests (exists)
- [ ] Add Chat to API docs
- [ ] Integration tests
- [ ] Documentation update

**Effort**: 3 dias (validation only, já implementado)

---

### 15. **WAHA Integration Improvements** (1 week)
**Current**: Only sending messages implemented

**Tasks**:
- [ ] Fetch message history (GET /api/{session}/messages)
- [ ] Handle all webhook events:
  - message.ack (delivery status)
  - message.revoked (deleted)
  - state.change (session state)
  - group.join/leave
  - call.received/accepted/rejected
  - presence.update
  - chat.archived
  - contact.changed
  - label.upsert
- [ ] Send additional message types:
  - Location
  - Contact card
  - Poll
  - Buttons (WhatsApp Business)
  - List (WhatsApp Business)
- [ ] Tests

**Effort**: 1 semana

---

## 📅 EXECUTION ROADMAP (30 SPRINTS = 7.5 MONTHS)

### **Phase 1: Security & Performance** (0-2 months)
- Sprint 1-2: Security Fixes P0 (3-4 sem) 🔴
- Sprint 3-4: Cache + Performance (2 sem) 🔴

**Milestone M1** (Week 6): Production-safe API ✅

---

### **Phase 2: AI Foundation** (2-5 months)
- Sprint 5-11: Memory Service (7 sem) 🔴
- Sprint 12-14: gRPC API (3 sem) 🔴

**Milestone M2** (Week 17): Memory Service + gRPC ready ✅

---

### **Phase 3: AI Tools** (5-7 months)
- Sprint 15-18: MCP Server (4 sem) 🔴
- Sprint 19-24: Python ADK (6 sem) 🔴

**Milestone M3** (Week 30): Multi-agent system production ✅

---

### **Phase 4: Polish** (parallel - ongoing)
- Testing coverage (2 sem) 🟡
- Resilience patterns (2 sem) 🟡
- Optimistic locking (1 sem) 🟡
- Value Objects (2 sem) 🟢
- Knowledge Graph (2 sem) 🟢
- Observability (2 sem) 🟢

---

## 📊 SUCCESS METRICS

### **Technical KPIs**
- ✅ Build: SUCCESS (0 errors)
- ✅ Tests: >85% coverage
- ✅ Security: 0 P0 vulnerabilities
- ✅ Latency: <200ms (API), <50ms (cache)
- ✅ Cache hit rate: >70%
- ✅ AI/ML Score: >8.0/10

### **Business KPIs**
- 🤖 AI agents reduce workload 60%
- 🎯 Lead qualification automatic (+30% conversion)
- 💰 Churn prediction (-20% churn)
- 📈 Memory context improves NPS (+15 points)
- 💵 AI cost tracking prevents billing surprises

---

## 🔍 KEY INSIGHTS FROM REPORT

### **Architectural Strengths** ✅
1. DDD + Clean Architecture (8.5/10)
2. Event-Driven (8.5/10) - 182 events, outbox <100ms
3. Database (9.2/10) - 49 migrations, 350+ indexes, 3NF
4. Message Enrichment (8.5/10) - 12 providers production-ready
5. CQRS (8.0/10) - 80+ commands, 20+ queries

### **Critical Gaps** ❌
1. **Security**: 5 P0 vulnerabilities (CVSS 9.1, 8.2, 7.5, 7.1)
2. **Cache**: 0% integration (Redis configured but unused)
3. **Memory Service**: 80% missing (vector DB, hybrid search, facts)
4. **Python ADK**: 0% (multi-agent not started)
5. **MCP Server**: 0% (30 tools planned, 0 implemented)
6. **gRPC API**: 0% (Go ↔ Python communication)
7. **AI Cost Tracking**: 0% (billing surprises risk)

### **Discoveries**
1. ✅ Chat aggregate 100% implemented (contradicts old docs)
2. ✅ 30 aggregates total (not 23)
3. ✅ 158 endpoints (not "50+")
4. ⚠️ 14 aggregates without optimistic locking (47%)
5. ⚠️ 39/44 use cases without tests (89%)

---

## 🎯 DECISION POINTS

### **For Basic Production (CRM without advanced AI)**:
✅ **READY after 4 weeks** (Sprint 1-4: Security + Cache)
- Message enrichment works
- CRUD complete
- Event-driven OK

### **For Production with Advanced AI** (Multi-agent, Memory):
❌ **6 MONTHS** (24 sprints)
- Security fixes (4 weeks)
- Cache layer (2 weeks)
- Memory Service (7 weeks)
- gRPC API (3 weeks)
- MCP Server (4 weeks)
- Python ADK (6 weeks)

---

## 📚 REFERENCES

### **Architecture Report**
- `AI_REPORT_PART1.md` - Backend Go, Domain, Persistence (Tables 1-5)
- `AI_REPORT_PART2.md` - Value Objects, Normalization, Migrations (Tables 6-10)
- `AI_REPORT_PART3.md` - Events, Workflows, Consistency (Tables 11-15)
- `AI_REPORT_PART4.md` - API, Security OWASP, DTOs (Tables 16-20)
- `AI_REPORT_PART5.md` - AI/ML, Testing, Python ADK, gRPC (Tables 21-25)
- `AI_REPORT_PART6.md` - MCP Server, Roadmap, Scores (Tables 26-30)

### **Technical Docs**
- `TODO_PYTHON.md` - Python ADK detailed roadmap
- `docs/MCP_SERVER_COMPLETE.md` - MCP Server architecture (1,175 lines)
- `docs/PYTHON_ADK_ARCHITECTURE.md` - Multi-agent system (3 parts, 3,000+ lines)

### **External References**
- OWASP API Security: https://owasp.org/API-Security/
- Google Cloud ADK: https://cloud.google.com/vertex-ai/docs/agent-builder
- Temporal Docs: https://docs.temporal.io/
- Apache AGE: https://age.apache.org/

---

## 📝 NOTES

### **Priorities Explained**
- 🔴 **P0 (Critical)**: Blockers for production or severe security issues
- 🟡 **P1 (Important)**: Required for advanced features or quality
- 🟢 **P2 (Improvement)**: Nice to have, optimization

### **Effort Estimates**
Based on detailed analysis in architectural report:
- 1 day = 8 hours
- 1 week = 5 days = 40 hours
- Includes: implementation + tests + documentation

### **Resource Recommendation**
- **Backend Go Engineer**: 2 FTEs (security, cache, memory)
- **AI/ML Engineer**: 2 FTEs (Python ADK, semantic router, facts)
- **DevOps Engineer**: 0.5 FTE (infra, observability)
- **QA Engineer**: 1 FTE (testing, E2E, integration)
- **Total**: 5.5 FTEs

---

**Last Update**: 2025-10-13 (Post-architectural audit)
**Next Review**: After Sprint 4 (M1: Production-safe API)
**Maintainer**: Ventros CRM Team
**Status**: ✅ Consolidated TODO based on complete codebase analysis
