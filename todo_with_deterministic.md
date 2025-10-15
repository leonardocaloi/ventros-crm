# TODO - Ventros CRM (Deterministic Enhanced)

## 📋 CONSOLIDATED TODO - Based on Complete Architectural Report

**Last Update**: 2025-10-14 (Enhanced with deterministic patterns)
**Report Reference**: `AI_REPORT_PART*.md` (6 parts, 30 evaluation tables)
**Build Status**: ✅ SUCCESS (0 errors, 0 warnings)
**Test Status**: ✅ 100% tests passing
**Determinism Score**: 🎯 7.5/10 (needs improvement in AI components)

---

## 🎯 EXECUTIVE SUMMARY

**Overall Scores**:
- Backend Go: **8.0/10** (B+) - Production-Ready
- Database: **9.2/10** (A) - Excellent
- API Security: **6.0/10** (C+) - **5 P0 Critical Vulnerabilities**
- AI/ML: **2.5/10** (F) - Only enrichment working, 80% missing
- **Determinism**: **7.5/10** (B) - Strong events/DB, weak AI/tests
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
9. ⚠️ **Non-deterministic behavior in tests** (time.Now(), random UUIDs, no fixtures)
10. ⚠️ **AI responses non-deterministic** (need temperature=0, seed parameter)

---

## 🔴 PRIORITY 0: CRITICAL & URGENT (0-4 weeks)

### **SPRINT 0: Deterministic Foundation** (1 week) - NEW

**Purpose**: Establish deterministic behavior across the entire system to improve reliability, testability, and debugging.

#### 0.1. 🎯 **Time Management Patterns** (2 days)

**Problem**: `time.Now()` scattered throughout codebase causes non-deterministic tests and makes debugging difficult.

**Current Anti-patterns**:
```go
// ❌ Non-deterministic - different on every run
contact.CreatedAt = time.Now()

// ❌ Tests fail at midnight
if time.Now().Hour() < 9 { ... }

// ❌ Can't reproduce past bug
logger.Info("Error at", time.Now())
```

**Solution: Clock Interface Pattern**
```go
// Domain layer: internal/domain/shared/clock.go
type Clock interface {
    Now() time.Time
    Since(t time.Time) time.Duration
}

// Production: Real clock
type SystemClock struct{}
func (c *SystemClock) Now() time.Time { return time.Now() }

// Testing: Frozen clock
type FrozenClock struct {
    frozenTime time.Time
}
func (c *FrozenClock) Now() time.Time { return c.frozenTime }
```

**Tasks**:
- [ ] Create `internal/domain/shared/clock.go` interface
- [ ] Inject Clock into all aggregates (Contact, Session, Message, etc.)
- [ ] Inject Clock into command handlers
- [ ] Replace all `time.Now()` calls (estimated 150+ occurrences)
- [ ] Use FrozenClock in all tests
- [ ] Add Clock to application layer services
- [ ] Tests for Clock abstraction
- [ ] Documentation

**Impact**: Tests become 100% deterministic, bugs reproducible

**Effort**: 2 days

---

#### 0.2. 🎯 **UUID Generation Patterns** (1 day)

**Problem**: `uuid.New()` generates random UUIDs, making tests non-deterministic and snapshots impossible.

**Solution: UUID Generator Interface**
```go
// Domain layer: internal/domain/shared/id_generator.go
type IDGenerator interface {
    NewID() uuid.UUID
}

// Production: Random UUID
type RandomIDGenerator struct{}
func (g *RandomIDGenerator) NewID() uuid.UUID { return uuid.New() }

// Testing: Sequential UUID
type SequentialIDGenerator struct {
    counter int64
}
func (g *SequentialIDGenerator) NewID() uuid.UUID {
    g.counter++
    return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", g.counter))
}
```

**Example Usage**:
```go
// Domain aggregate constructor
func NewContact(idGen IDGenerator, clock Clock, name string) *Contact {
    return &Contact{
        id:        idGen.NewID(),
        createdAt: clock.Now(),
        name:      name,
    }
}

// Test
func TestContact(t *testing.T) {
    idGen := &SequentialIDGenerator{}
    clock := &FrozenClock{frozenTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}

    contact := NewContact(idGen, clock, "Alice")

    assert.Equal(t, "00000000-0000-0000-0000-000000000001", contact.ID().String()) // Deterministic!
    assert.Equal(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), contact.CreatedAt())
}
```

**Tasks**:
- [ ] Create IDGenerator interface
- [ ] Inject IDGenerator into all aggregates
- [ ] Replace all `uuid.New()` calls (estimated 80+ occurrences)
- [ ] Use SequentialIDGenerator in tests
- [ ] Tests
- [ ] Documentation

**Effort**: 1 day

---

#### 0.3. 🎯 **Test Fixtures & Golden Files** (2 days)

**Problem**: Tests use inline data, making them brittle and hard to maintain.

**Solution: Centralized Test Fixtures**

**Directory Structure**:
```
tests/
├── fixtures/
│   ├── contacts.json         # 10 sample contacts
│   ├── sessions.json         # 20 sample sessions
│   ├── messages.json         # 50 sample messages
│   ├── campaigns.json        # 5 sample campaigns
│   └── events.json           # 100 domain events
├── golden/
│   ├── contact_dto.golden    # Expected JSON output
│   ├── session_stats.golden  # Expected analytics
│   └── memory_search.golden  # Expected search results
└── testutil/
    ├── fixtures.go           # Fixture loader
    ├── golden.go             # Golden file comparator
    └── builders.go           # Test builders
```

**Example**:
```go
// tests/testutil/fixtures.go
func LoadContactFixture(t *testing.T, name string) *domain.Contact {
    data, _ := os.ReadFile(fmt.Sprintf("fixtures/contacts/%s.json", name))
    var contact domain.Contact
    json.Unmarshal(data, &contact)
    return &contact
}

// tests/testutil/golden.go
func AssertGolden(t *testing.T, name string, actual interface{}) {
    goldenPath := fmt.Sprintf("golden/%s.golden", name)
    actualJSON, _ := json.MarshalIndent(actual, "", "  ")

    if os.Getenv("UPDATE_GOLDEN") == "1" {
        os.WriteFile(goldenPath, actualJSON, 0644)
        return
    }

    expected, _ := os.ReadFile(goldenPath)
    assert.JSONEq(t, string(expected), string(actualJSON))
}

// Usage in test
func TestGetContactDTO(t *testing.T) {
    contact := LoadContactFixture(t, "alice")
    dto := mapper.ToDTO(contact)
    AssertGolden(t, "contact_dto_alice", dto) // Compare with golden file
}
```

**Tasks**:
- [ ] Create `tests/fixtures/` directory
- [ ] Create 10 contact fixtures
- [ ] Create 20 session fixtures
- [ ] Create 50 message fixtures
- [ ] Create 5 campaign fixtures
- [ ] Create golden file helper
- [ ] Create fixture loader helper
- [ ] Create test builders (ContactBuilder, SessionBuilder, etc.)
- [ ] Migrate 20 priority tests to use fixtures
- [ ] Documentation

**Effort**: 2 days

---

#### 0.4. 🎯 **Event Ordering & Idempotency** (2 days)

**Problem**: Event consumers may receive events out of order or duplicate events.

**Current Gaps**:
- No event sequence number
- No idempotency key in handlers
- No event version tracking
- Race conditions possible

**Solution**:

1. **Event Sequence in Outbox**:
```go
// Migration: Add sequence to outbox_events
ALTER TABLE outbox_events ADD COLUMN sequence BIGSERIAL;
CREATE UNIQUE INDEX idx_outbox_sequence ON outbox_events(aggregate_id, sequence);
```

2. **Idempotent Event Handlers**:
```go
// Before processing event, check if already processed
type ProcessedEvent struct {
    EventID   uuid.UUID `gorm:"primary_key"`
    ProcessedAt time.Time
}

func (h *ContactEventHandler) Handle(event DomainEvent) error {
    // Check idempotency
    var processed ProcessedEvent
    if h.db.First(&processed, "event_id = ?", event.EventID()).Error == nil {
        return nil // Already processed, skip
    }

    // Process event
    // ...

    // Mark as processed
    h.db.Create(&ProcessedEvent{EventID: event.EventID(), ProcessedAt: h.clock.Now()})
    return nil
}
```

3. **Event Ordering Guarantees**:
```go
// Consume events per aggregate in order
func (c *Consumer) Start() {
    // One goroutine per aggregate_id ensures order
    c.ch.Consume("events", func(msg amqp.Delivery) {
        aggregateID := msg.Headers["aggregate_id"]
        c.getWorkerForAggregate(aggregateID).Process(msg)
    })
}
```

**Tasks**:
- [ ] Add sequence column to outbox_events
- [ ] Create processed_events table
- [ ] Implement idempotency check in all event handlers (12 handlers)
- [ ] Add aggregate-based routing in RabbitMQ consumer
- [ ] Add event version to DomainEvent interface
- [ ] Tests (out-of-order, duplicates, concurrent)
- [ ] Documentation

**Effort**: 2 days

---

**Total Effort Sprint 0**: 1 week

**Impact**: Foundation for deterministic system behavior

---

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
- [ ] Add IP whitelist for dev mode (127.0.0.1, ::1)
- [ ] Add environment check (GO_ENV=production)
- [ ] Add audit log for dev mode usage
- [ ] Tests (verify panic in production)
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
- [ ] Block cloud metadata (169.254.169.254, fd00:ec2::254)
- [ ] Whitelist schemes (HTTPS only in production)
- [ ] Add DNS rebinding protection
- [ ] Add request timeout (5s)
- [ ] Tests (all blocked IPs, valid URLs)
- [ ] Helper: `isPrivateIP()`, `isCloudMetadata()`, `validateWebhookURL()`

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
- [ ] Add ownership check helper: `middleware/ownership.go`
  ```go
  func CheckOwnership(entity interface{}) gin.HandlerFunc {
      return func(c *gin.Context) {
          authCtx := c.MustGet("auth").(AuthContext)
          tenantID := getTenantID(entity)
          if tenantID != authCtx.TenantID {
              c.JSON(404, gin.H{"error": "not found"}) // 404 prevents info leak
              c.Abort()
              return
          }
          c.Next()
      }
  }
  ```
- [ ] Apply to 60 GET endpoints:
  - `GET /contacts/:id`
  - `GET /messages/:id`
  - `GET /sessions/:id`
  - `GET /campaigns/:id`
  - `GET /pipelines/:id`
  - `GET /chats/:id`
  - ... (54 more)
- [ ] Return 404 (not 403) to avoid info leak
- [ ] Deterministic tests for each endpoint (unauthorized access returns 404)
- [ ] Add security test suite

**Effort**: 1 week (60 handlers × 10 min each + tests)

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
- [ ] Create `shared/pagination.go` with constants
  ```go
  const (
      DefaultPageSize = 20
      MaxPageSize     = 100
      QueryTimeout    = 10 * time.Second
  )

  func ValidatePagination(page, limit int) (int, int, error) {
      if page < 1 {
          page = 1
      }
      if limit < 1 {
          limit = DefaultPageSize
      }
      if limit > MaxPageSize {
          return 0, 0, ErrPageSizeTooLarge
      }
      return page, limit, nil
  }
  ```
- [ ] Enforce in all 19 queries
- [ ] Add query timeouts (10s via context)
- [ ] Add max payload size (10MB) in middleware
- [ ] Add pagination response metadata (total, has_next)
- [ ] Tests (verify max limit enforced)

**Effort**: 3 days (19 queries)

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

**Permission Matrix**:

| Role    | READ | CREATE | UPDATE | DELETE | ASSIGN | EXPORT |
|---------|------|--------|--------|--------|--------|--------|
| Admin   | ✅   | ✅     | ✅     | ✅     | ✅     | ✅     |
| Agent   | ✅   | ✅     | ✅ (own)| ❌    | ✅ (own)| ✅ (own)|
| Viewer  | ✅   | ❌     | ❌     | ❌     | ❌     | ✅     |

**Fix**:
- [ ] Create `middleware/rbac.go`
  ```go
  func Authorize(allowedRoles ...string) gin.HandlerFunc {
      return func(c *gin.Context) {
          authCtx := c.MustGet("auth").(AuthContext)
          for _, role := range allowedRoles {
              if authCtx.Role == role {
                  c.Next()
                  return
              }
          }
          c.JSON(403, gin.H{"error": "forbidden"})
          c.Abort()
      }
  }
  ```
- [ ] Define permission constants
- [ ] Apply RBAC middleware to 95 endpoints
- [ ] Add role column to users table
- [ ] Add role to JWT claims
- [ ] Tests for each role (admin, agent, viewer)
- [ ] Documentation

**Effort**: 2 weeks

---

#### 1.6. 🔴 **Rate Limiting (Redis)** (1 week) - P0
**Location**: `infrastructure/http/middleware/rate_limiter.go`

**Current**: In-memory (non-deterministic, not scalable, easy bypass)

**Fix with Deterministic Redis Implementation**:
```go
// Use Redis sorted sets for deterministic sliding window
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    now := r.clock.Now().UnixNano() // Deterministic clock!
    windowStart := now - window.Nanoseconds()

    pipe := r.redis.TxPipeline()

    // Remove old entries
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(windowStart))

    // Count current requests
    pipe.ZCard(ctx, key)

    // Add current request
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

    // Set expiration
    pipe.Expire(ctx, key, window)

    results, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    count := results[1].(*redis.IntCmd).Val()
    return count < int64(limit), nil
}
```

**Tasks**:
- [ ] Redis integration with Clock interface (deterministic tests!)
- [ ] Sliding window counter algorithm
- [ ] Per-user/tenant limits
- [ ] Different limits per endpoint group:
  - Auth: 5 req/min
  - CRM read: 100 req/min
  - CRM write: 20 req/min
  - Webhooks: 10 req/min
  - AI: 10 req/min
- [ ] Rate limit headers (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset)
- [ ] Deterministic tests with FrozenClock
- [ ] Tests (concurrent requests, different users, exceed limit)

**Effort**: 1 week

**Total Effort Sprint 1-2**: 3-4 weeks

---

### **SPRINT 3-4: Cache Layer + Performance** (2 weeks) - P0

**Reference**: `AI_REPORT_PART3.md` - Table 13 (Queries & Performance)

#### 2.1. 🔴 **Redis Cache Integration** (1 week)
**Status**: Redis client configured but **0% integrated**

**Deterministic Cache Design**:

1. **Cache Key Strategy (Deterministic)**:
```go
// Always generate same cache key for same input
func CacheKey(resource string, params ...interface{}) string {
    // Sort params for deterministic key
    sortedParams := sortParams(params)
    hash := sha256.Sum256([]byte(fmt.Sprintf("%v", sortedParams)))
    return fmt.Sprintf("cache:%s:%x", resource, hash)
}

// Example
key := CacheKey("contact_stats", tenantID, contactID)
// Always produces: cache:contact_stats:a7b3c9...
```

2. **Cache Invalidation (Event-Driven)**:
```go
// Listen to events and invalidate deterministically
func (c *CacheInvalidator) OnContactUpdated(event ContactUpdatedEvent) {
    keys := []string{
        CacheKey("contact_stats", event.TenantID, event.ContactID),
        CacheKey("contacts_list", event.TenantID),
    }
    c.redis.Del(context.Background(), keys...)
}
```

**Tasks**:
- [ ] Create `infrastructure/cache/cache_service.go`
- [ ] Implement deterministic cache key generation
- [ ] Cache middleware with TTL
- [ ] Cache 5 priority queries:
  1. `GetContactStatsQuery` (TTL: 5 min) - currently 423ms avg
  2. `SessionAnalyticsQuery` (TTL: 30 min) - currently 678ms avg
  3. `ListContactsQuery` (TTL: 2 min)
  4. `MessageHistoryQuery` (TTL: 1 min)
  5. `GetActiveSessionsQuery` (TTL: 30 sec)
- [ ] Cache invalidation via events:
  - contact.* → invalidate contact_stats:*
  - session.closed → invalidate session_analytics:*
  - message.created → invalidate message_history:*
- [ ] Cache hit rate monitoring (target: >70%)
- [ ] Deterministic cache tests with FrozenClock
- [ ] Tests (hit, miss, invalidation, expiration)

**Expected Result**: Queries <200ms (from 423ms-678ms), deterministic behavior

**Effort**: 1 week

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

**Fix**: Single JOIN query (deterministic order)
```go
// Single query with deterministic ordering
contacts := db.
    Joins("JOIN contact_list_memberships ON contacts.id = contact_list_memberships.contact_id").
    Where("contact_list_memberships.list_id = ?", listID).
    Order("contacts.created_at ASC, contacts.id ASC"). // Deterministic order!
    Preload("Tags").
    Preload("CustomFields").
    Find(&contacts)
```

**Impact**: 100 contacts = 100 queries → 1 query (100x faster), deterministic result order

**Tasks**:
- [ ] Fix GetContactsInListQuery (single JOIN + deterministic ORDER BY)
- [ ] Verify ConversationThreadQuery (possible N+1)
- [ ] Add query logging to detect N+1 patterns
- [ ] Add EXPLAIN ANALYZE tests
- [ ] Tests (verify 1 query, deterministic order)

**Effort**: 2 days

---

#### 2.3. 🔴 **Materialized View for Analytics** (3 days)
**Location**: `internal/application/queries/session_analytics_query.go`

**Problem**: `SessionAnalyticsQuery` very slow (678ms avg, 1200ms p95), non-deterministic due to concurrent updates

**Fix with Deterministic Materialized View**:
```sql
-- Migration 000049: Materialized view for session analytics
CREATE MATERIALIZED VIEW session_analytics_mv AS
SELECT
    tenant_id,
    DATE(started_at) as date,
    COUNT(*) as total_sessions,
    AVG(EXTRACT(EPOCH FROM (ended_at - started_at))) as avg_duration_seconds,
    COUNT(*) FILTER (WHERE status = 'completed') as completed_sessions,
    COUNT(*) FILTER (WHERE status = 'abandoned') as abandoned_sessions,
    created_at as last_updated
FROM sessions
WHERE deleted_at IS NULL
GROUP BY tenant_id, DATE(started_at)
ORDER BY tenant_id, date DESC; -- Deterministic order

-- Refresh strategy: hourly via cron or event-driven
CREATE UNIQUE INDEX idx_session_analytics_mv ON session_analytics_mv(tenant_id, date);
```

**Tasks**:
- [ ] Migration: `session_analytics_mv` (materialized view)
- [ ] Refresh strategy (hourly via cron + event-driven on session.closed)
- [ ] Query rewrite to use MV
- [ ] Add refresh timestamp tracking
- [ ] Performance tests (<100ms, deterministic results)
- [ ] Tests (verify MV data matches source)

**Effort**: 3 days

**Total Effort Sprint 3-4**: 2 weeks

---

## 🟡 PRIORITY 1: IMPORTANT (4-12 weeks)

### **SPRINT 5-11: Memory Service Foundation** (7 weeks) - P0 for AI

**Reference**: `AI_REPORT_PART5.md` - Table 21 (AI/ML Components)

#### 3.1. **pgvector + Vector Search** (1 week)
**Status**: ❌ 0% implemented

**Deterministic Considerations**:
- Embedding generation: Use `seed` parameter for reproducibility
- Distance calculation: Cosine similarity is deterministic
- Top-K retrieval: Add secondary sort by ID for deterministic ordering

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
      embedding_model VARCHAR(100) NOT NULL, -- Track model version
      embedding_version INT NOT NULL, -- Track version for determinism
      metadata JSONB,
      created_at TIMESTAMP NOT NULL
  );
  CREATE INDEX ON memory_embeddings USING hnsw (embedding vector_cosine_ops);
  CREATE INDEX ON memory_embeddings (tenant_id, contact_id);
  ```
- [ ] `MemoryEmbeddingRepository`
- [ ] `VectorSearchService` with deterministic ordering
  ```go
  // Deterministic top-K: primary sort by distance, secondary by ID
  SELECT id, content_text, embedding <=> $1 as distance
  FROM memory_embeddings
  WHERE tenant_id = $2
  ORDER BY distance ASC, id ASC -- Deterministic!
  LIMIT $3;
  ```
- [ ] Embedding worker (RabbitMQ consumer)
- [ ] Vertex AI integration (text-embedding-005 with version tracking)
- [ ] Embedding cache (same text → same embedding)
- [ ] Tests + benchmarks (<100ms, deterministic results with fixtures)

**Effort**: 1 week

---

#### 3.2. **Hybrid Search Service** (2 weeks)
**Status**: ❌ 0% implemented

**Deterministic Hybrid Search**:

```go
type HybridSearchService struct {
    vectorSearch  *VectorSearchService
    keywordSearch *KeywordSearchService
    graphSearch   *GraphSearchService
    clock         Clock // For deterministic timestamps
}

// Deterministic RRF (Reciprocal Rank Fusion)
func (s *HybridSearchService) Search(ctx context.Context, query string, weights SearchWeights) ([]SearchResult, error) {
    // 1. SQL BASELINE (last 20 messages) - deterministic order
    baseline := s.getBaselineMessages(ctx, 20) // ORDER BY created_at DESC, id ASC

    // 2. VECTOR SEARCH (50% weight) - deterministic with secondary sort
    vectorResults := s.vectorSearch.Search(ctx, query, weights.VectorWeight)

    // 3. KEYWORD SEARCH (20% weight) - deterministic pg_trgm
    keywordResults := s.keywordSearch.Search(ctx, query, weights.KeywordWeight)

    // 4. GRAPH TRAVERSAL (20% weight) - deterministic Cypher ORDER BY
    graphResults := s.graphSearch.Search(ctx, query, weights.GraphWeight)

    // 5. RRF Fusion (deterministic ranking)
    results := s.reciprocalRankFusion(baseline, vectorResults, keywordResults, graphResults)

    // 6. Deterministic sort: primary by score DESC, secondary by timestamp DESC, tertiary by ID ASC
    sort.Slice(results, func(i, j int) bool {
        if results[i].Score != results[j].Score {
            return results[i].Score > results[j].Score
        }
        if !results[i].Timestamp.Equal(results[j].Timestamp) {
            return results[i].Timestamp.After(results[j].Timestamp)
        }
        return results[i].ID.String() < results[j].ID.String()
    })

    return results, nil
}
```

**Tasks**:
- [ ] `HybridSearchService` with deterministic scoring
- [ ] 4 search strategies:
  1. SQL BASELINE (last 20 messages) - ORDER BY created_at DESC, id ASC
  2. VECTOR SEARCH (50% weight) - with secondary ID sort
  3. KEYWORD SEARCH (20% weight) - pg_trgm with deterministic order
  4. GRAPH TRAVERSAL (20% weight) - Apache AGE with ORDER BY
- [ ] RRF (Reciprocal Rank Fusion) with deterministic tie-breaking
- [ ] Optional reranking (Jina v2) with seed parameter
- [ ] Result deduplication (deterministic by ID)
- [ ] Tests (<500ms latency, >90% recall, deterministic order)
- [ ] Golden file tests (same query → same results)

**Effort**: 2 weeks

---

#### 3.3. **Memory Facts Extraction** (2 weeks)
**Status**: ❌ 0% implemented

**Deterministic Facts Extraction**:

```go
// Use LLM with temperature=0 and seed for reproducibility
type FactExtractionService struct {
    llmClient *GeminiClient
    clock     Clock
}

func (s *FactExtractionService) ExtractFacts(ctx context.Context, message string) ([]Fact, error) {
    prompt := s.buildPrompt(message) // Deterministic prompt template

    // Gemini with temperature=0 for deterministic output
    response, err := s.llmClient.Generate(ctx, prompt, GenerateOptions{
        Temperature: 0.0,  // Deterministic!
        Seed:        12345, // Fixed seed for reproducibility
        Model:       "gemini-1.5-flash-002", // Pin version
    })

    facts := s.parseFacts(response)

    // Sort facts deterministically
    sort.Slice(facts, func(i, j int) bool {
        if facts[i].Confidence != facts[j].Confidence {
            return facts[i].Confidence > facts[j].Confidence
        }
        return facts[i].Text < facts[j].Text
    })

    return facts, nil
}
```

**Tasks**:
- [ ] Migration 000051: `memory_facts` table
  ```sql
  CREATE TABLE memory_facts (
      id UUID PRIMARY KEY,
      tenant_id UUID NOT NULL,
      contact_id UUID NOT NULL,
      fact_type VARCHAR(50) NOT NULL, -- budget, preference, objection
      fact_text TEXT NOT NULL,
      confidence FLOAT NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
      llm_model VARCHAR(100) NOT NULL, -- Track model version
      llm_seed INT NOT NULL, -- Track seed for reproducibility
      source_message_id UUID,
      extracted_at TIMESTAMP NOT NULL,
      created_at TIMESTAMP NOT NULL
  );
  CREATE INDEX ON memory_facts (contact_id, fact_type, confidence DESC, id ASC); -- Deterministic order
  ```
- [ ] `FactExtractionService` with deterministic LLM calls
- [ ] LLM-based extraction (Gemini Flash with temperature=0)
- [ ] Fact types: budget, preference, objection, pain_point, timeline, decision_maker
- [ ] Confidence scoring (0.0-1.0)
- [ ] RabbitMQ consumer (message.created)
- [ ] Fact deduplication (deterministic by fact_text hash)
- [ ] Tests with mocked LLM responses (deterministic)
- [ ] Golden file tests

**Example Facts**:
- "Cliente tem budget de R$ 5.000/mês" (confidence: 0.92)
- "Pain point: ROI não está claro" (confidence: 0.85)
- "Objeção: Concorrente 30% mais barato" (confidence: 0.78)
- "Decision maker: CFO" (confidence: 0.95)

**Effort**: 2 weeks

---

#### 3.4. **AI Cost Tracking** (1 week)
**Status**: ❌ 0% implemented - **CRITICAL for billing**

**Deterministic Cost Tracking**:

```go
// Track every AI call deterministically
type CostTracker struct {
    repo  CostRepository
    clock Clock // Deterministic time
}

func (t *CostTracker) Track(ctx context.Context, cost AICost) error {
    cost.ID = t.idGen.NewID() // Deterministic ID in tests
    cost.Timestamp = t.clock.Now() // Deterministic time
    cost.CalculatedAt = t.clock.Now()

    return t.repo.Save(ctx, cost)
}
```

**Tasks**:
- [ ] Migration 000052: `ai_costs` table
  ```sql
  CREATE TABLE ai_costs (
      id UUID PRIMARY KEY,
      tenant_id UUID NOT NULL,
      project_id UUID NOT NULL,
      provider VARCHAR(50) NOT NULL, -- vertex, groq, openai, llamaparse
      model VARCHAR(100) NOT NULL, -- gemini-1.5-flash-002, whisper-large-v3
      operation VARCHAR(50) NOT NULL, -- transcribe, embed, extract, vision
      units DECIMAL(10,2) NOT NULL, -- tokens, seconds, images, pages
      unit_type VARCHAR(20) NOT NULL, -- tokens, seconds, images, pages
      unit_cost DECIMAL(10,6) NOT NULL, -- Cost per unit
      total_cost DECIMAL(10,6) NOT NULL, -- Total cost
      request_id UUID, -- Link to original request
      timestamp TIMESTAMP NOT NULL,
      created_at TIMESTAMP NOT NULL
  );
  CREATE INDEX ON ai_costs (tenant_id, timestamp DESC, id ASC); -- Deterministic order
  CREATE INDEX ON ai_costs (project_id, provider, timestamp DESC);
  ```
- [ ] `CostTracker` service with Clock injection
- [ ] Integrate in all AI providers:
  - Vertex Vision ($0.0025/image)
  - Groq Whisper (FREE - track usage anyway)
  - OpenAI Whisper ($0.006/min)
  - LlamaParse ($0.003/page)
  - Gemini Flash ($0.35/1M input tokens, $1.05/1M output tokens)
  - Text Embeddings ($0.025/1M tokens)
- [ ] Cost aggregation query (deterministic ORDER BY)
- [ ] Budget alerts (threshold-based)
- [ ] Daily cost report (scheduled with deterministic time)
- [ ] Dashboard query with pagination
- [ ] Deterministic tests with FrozenClock
- [ ] Tests (verify cost calculation, aggregation)

**Effort**: 1 week

---

#### 3.5. **Retrieval Strategies Dictionary** (3 days)
**Status**: ❌ 0% implemented

**Deterministic Strategy Selection**:

```go
// Agent category deterministically selects strategy
type StrategySelector struct {
    strategies map[string]RetrievalStrategy
}

func (s *StrategySelector) SelectStrategy(agentCategory string) RetrievalStrategy {
    // Deterministic lookup
    if strategy, exists := s.strategies[agentCategory]; exists {
        return strategy
    }
    return s.strategies["balanced"] // Fallback
}
```

**Tasks**:
- [ ] Migration 000053: `retrieval_strategies` table
  ```sql
  CREATE TABLE retrieval_strategies (
      id UUID PRIMARY KEY,
      agent_category VARCHAR(50) NOT NULL UNIQUE,
      vector_weight FLOAT NOT NULL CHECK (vector_weight >= 0 AND vector_weight <= 1),
      keyword_weight FLOAT NOT NULL CHECK (keyword_weight >= 0 AND keyword_weight <= 1),
      graph_weight FLOAT NOT NULL CHECK (graph_weight >= 0 AND graph_weight <= 1),
      baseline_enabled BOOLEAN NOT NULL DEFAULT true,
      rerank_enabled BOOLEAN NOT NULL DEFAULT false,
      max_results INT NOT NULL DEFAULT 20,
      created_at TIMESTAMP NOT NULL,
      updated_at TIMESTAMP NOT NULL,
      CONSTRAINT weights_sum_check CHECK (
          vector_weight + keyword_weight + graph_weight <= 1.0
      )
  );
  ```
- [ ] Strategy configs per agent category:
  ```go
  "sales_prospecting": {
      VectorWeight:  0.20,
      KeywordWeight: 0.30,
      GraphWeight:   0.40, // HIGH: campaign attribution
      BaselineEnabled: true,
      MaxResults: 20,
  },
  "retention_churn": {
      VectorWeight:  0.50, // HIGH: semantic understanding
      KeywordWeight: 0.20,
      GraphWeight:   0.20,
      BaselineEnabled: true,
      MaxResults: 15,
  },
  "support_technical": {
      VectorWeight:  0.40,
      KeywordWeight: 0.40, // HIGH: exact matches (error codes, SKUs)
      GraphWeight:   0.10,
      BaselineEnabled: true,
      RerankEnabled: true,
      MaxResults: 10,
  },
  ```
- [ ] Dynamic weight adjustment (A/B testing framework)
- [ ] Strategy versioning (track changes)
- [ ] Deterministic selection logic
- [ ] Tests (strategy selection, weight validation)

**Effort**: 3 days

**Total Effort Sprint 5-11**: 7 weeks

---

### **SPRINT 12-14: gRPC API** (3 weeks) - P0 for Python ADK

**Reference**: `AI_REPORT_PART5.md` - Table 25 (gRPC API)

#### 4.1. **Proto Definitions** (3 days)
**Status**: ❌ Directory `api/proto/` doesn't exist

**Deterministic gRPC Design**:
- Use proto3 for deterministic serialization
- Include version field in all messages
- Include request_id for tracing
- Include timestamps for ordering

**Tasks**:
- [ ] Create `api/proto/` directory
- [ ] `api/proto/memory_service.proto`
  ```protobuf
  syntax = "proto3";

  message SearchMemoryRequest {
      string request_id = 1; // For tracing
      string tenant_id = 2;
      string contact_id = 3;
      string query = 4;
      int32 max_results = 5;
      int64 timestamp_nanos = 6; // Deterministic time
  }

  message SearchMemoryResponse {
      string request_id = 1;
      repeated MemoryResult results = 2;
      int64 processing_time_ms = 3;
      string cache_status = 4; // hit/miss
  }
  ```
  - SearchMemory
  - StoreEmbedding
  - ExtractFacts
  - GetContactFacts
- [ ] `api/proto/crm_service.proto` (partial)
  - GetContact
  - ListContacts
  - CreateMessage
  - GetSessionHistory
- [ ] Generate Go code (protoc-gen-go)
- [ ] Generate Python code (protoc-gen-python)
- [ ] Versioning strategy (proto package version)

**Effort**: 3 days

---

#### 4.2. **Go gRPC Server** (1 week)
**Status**: ❌ 0% implemented

**Deterministic Server Design**:
- Inject Clock for deterministic timestamps
- Use request_id for deterministic tracing
- Deterministic error codes

**Tasks**:
- [ ] `infrastructure/grpc/memory_service_server.go`
  ```go
  type MemoryServiceServer struct {
      memoryService *application.MemoryService
      clock         shared.Clock // Deterministic!
      logger        *zap.Logger
  }

  func (s *MemoryServiceServer) SearchMemory(ctx context.Context, req *pb.SearchMemoryRequest) (*pb.SearchMemoryResponse, error) {
      startTime := s.clock.Now() // Deterministic start time

      // Call application layer
      results, err := s.memoryService.Search(ctx, req.TenantId, req.ContactId, req.Query, int(req.MaxResults))

      processingTime := s.clock.Now().Sub(startTime).Milliseconds()

      return &pb.SearchMemoryResponse{
          RequestId:        req.RequestId,
          Results:          s.toProtoResults(results),
          ProcessingTimeMs: processingTime,
      }, nil
  }
  ```
- [ ] `cmd/grpc-server/main.go`
- [ ] Authentication (JWT interceptor)
- [ ] MemoryService implementation
- [ ] Error handling interceptor
- [ ] Logging interceptor
- [ ] Metrics interceptor (Prometheus)
- [ ] Connection pooling
- [ ] Graceful shutdown
- [ ] Tests with deterministic clock

**Effort**: 1 week

---

#### 4.3. **Python gRPC Client** (3 days)
**Status**: ❌ 0% implemented

**Deterministic Client Design**:
- Generate deterministic request_id (UUID v5 from input hash in tests)
- Include timestamp from external source (for determinism)

**Tasks**:
- [ ] `python-adk/ventros_adk/grpc_client/memory_client.py`
  ```python
  class MemoryClient:
      def __init__(self, host: str, port: int, clock: Clock):
          self.channel = grpc.insecure_channel(f"{host}:{port}")
          self.stub = memory_service_pb2_grpc.MemoryServiceStub(self.channel)
          self.clock = clock  # Inject clock for deterministic tests!

      def search_memory(self, tenant_id: str, contact_id: str, query: str, max_results: int = 20) -> List[MemoryResult]:
          request = memory_service_pb2.SearchMemoryRequest(
              request_id=str(uuid.uuid4()),
              tenant_id=tenant_id,
              contact_id=contact_id,
              query=query,
              max_results=max_results,
              timestamp_nanos=self.clock.now().timestamp() * 1e9  # Deterministic!
          )
          response = self.stub.SearchMemory(request)
          return [self._parse_result(r) for r in response.results]
  ```
- [ ] Connection pooling
- [ ] Retry logic (exponential backoff with max retries)
- [ ] Timeout configuration
- [ ] Error handling
- [ ] Logging
- [ ] Tests with mocked gRPC responses

**Effort**: 3 days

---

#### 4.4. **gRPC Interceptors** (3 days)
**Tasks**:
- [ ] Logging interceptor (log all requests/responses)
- [ ] Metrics interceptor (Prometheus)
  - Request count by method
  - Latency histogram
  - Error rate
- [ ] Error handling interceptor (convert domain errors to gRPC status codes)
- [ ] Authentication interceptor (JWT validation)
- [ ] Request ID propagation
- [ ] Deterministic tests

**Effort**: 3 days

**Total Effort Sprint 12-14**: 3 weeks

---

### **SPRINT 15-18: MCP Server** (4 weeks) - P0 for Claude Desktop

**Reference**: `AI_REPORT_PART6.md` - Table 26 (MCP Server)

**Status**: Docs complete (1,175 lines), code 0%

#### 5.1. **MCP Server Setup** (1 week)

**Deterministic MCP Server**:
- Tools return deterministic results (ordered by ID)
- Timestamps use injected Clock
- Cache tool responses for same input

**Tasks**:
- [ ] `infrastructure/mcp/server.go` (HTTP server)
  ```go
  type MCPServer struct {
      toolRegistry *ToolRegistry
      authService  *AuthService
      clock        shared.Clock // Deterministic!
      cache        *redis.Client
  }

  func (s *MCPServer) ExecuteTool(c *gin.Context) {
      toolName := c.Param("tool")
      params := c.MustGet("params").(map[string]interface{})

      // Check cache for deterministic results
      cacheKey := s.buildCacheKey(toolName, params)
      if cached, err := s.cache.Get(c, cacheKey).Result(); err == nil {
          c.JSON(200, cached)
          return
      }

      tool := s.toolRegistry.Get(toolName)
      result, err := tool.Execute(c, params)

      // Cache result
      s.cache.Set(c, cacheKey, result, 5*time.Minute)

      c.JSON(200, result)
  }
  ```
- [ ] `infrastructure/mcp/tool_registry.go`
- [ ] Authentication (JWT + API Keys)
- [ ] MCP protocol endpoints:
  - GET /mcp/tools (discovery)
  - POST /mcp/tools/:tool (execution)
  - GET /mcp/tools/:tool/stream (SSE)
- [ ] Rate limiting per tool
- [ ] Caching layer for tool responses
- [ ] Tests with deterministic clock and cache

**Effort**: 1 week

---

#### 5.2. **Priority MCP Tools** (2 weeks)
**25 tools prioritários** (all with deterministic results)

**Deterministic Tool Implementation Pattern**:
```go
type Tool interface {
    Name() string
    Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

type GetLeadsCountTool struct {
    query *application.GetLeadsCountQuery
    clock shared.Clock
}

func (t *GetLeadsCountTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    tenantID := params["tenant_id"].(string)
    startDate := params["start_date"].(time.Time)
    endDate := params["end_date"].(time.Time)

    // Query with deterministic ordering
    count, err := t.query.Execute(ctx, tenantID, startDate, endDate)

    return map[string]interface{}{
        "count":      count,
        "tenant_id":  tenantID,
        "start_date": startDate,
        "end_date":   endDate,
        "timestamp":  t.clock.Now(), // Deterministic!
    }, err
}
```

**BI Tools (7)** - All with deterministic aggregation and ordering:
- [ ] get_leads_count (COUNT with date range)
- [ ] get_conversion_rate (ROUND to 2 decimals)
- [ ] get_agent_performance (ORDER BY performance DESC, agent_id ASC)
- [ ] get_top_performing_agent (deterministic tie-breaking by agent_id)
- [ ] get_campaign_metrics (ORDER BY campaign_id)
- [ ] get_churn_prediction (deterministic ML model inference)
- [ ] get_revenue_forecast (deterministic time series calculation)

**CRM Operations Tools (8)**:
- [ ] qualify_lead (deterministic scoring)
- [ ] update_pipeline_stage (idempotent)
- [ ] assign_to_agent (deterministic selection algorithm)
- [ ] create_note (deterministic ID generation)
- [ ] schedule_follow_up (deterministic time calculation)
- [ ] send_message (idempotent with deduplication)
- [ ] tag_contact (deterministic tag ordering)
- [ ] export_contacts (deterministic ORDER BY)

**Memory Tools (5)**:
- [ ] search_memory (deterministic hybrid search)
- [ ] get_contact_context (ORDER BY timestamp DESC, id ASC)
- [ ] get_contact_facts (ORDER BY confidence DESC, id ASC)
- [ ] get_session_summary (deterministic summarization with seed)
- [ ] find_similar_contacts (deterministic vector similarity + ID sort)

**Document Tools (5)**:
- [ ] search_documents (deterministic full-text search)
- [ ] get_document_chunks (ORDER BY position ASC)
- [ ] upload_document (deterministic hash-based deduplication)
- [ ] summarize_document (deterministic LLM with seed)
- [ ] answer_from_documents (deterministic RAG)

**Effort**: 2 weeks (25 tools × 2h each)

---

#### 5.3. **Claude Desktop Integration** (1 week)
**Tasks**:
- [ ] Config file: `claude_desktop_config.json`
  ```json
  {
    "mcpServers": {
      "ventros": {
        "url": "http://localhost:8080/mcp",
        "apiKey": "${VENTROS_API_KEY}",
        "tools": [
          "get_leads_count",
          "search_memory",
          "qualify_lead",
          "get_contact_context"
        ]
      }
    }
  }
  ```
- [ ] Test all 25 tools in Claude Desktop
- [ ] Documentation with examples
- [ ] Example prompts:
  - "How many leads did we get this week?"
  - "What's the context for contact Alice?"
  - "Qualify this lead based on their messages"
  - "Show me top performing agents"
- [ ] Deterministic test suite (same prompt → same result)

**Effort**: 1 week

**Total Effort Sprint 15-18**: 4 weeks

---

### **SPRINT 19-24: Python ADK Multi-Agent** (6 weeks) - P0

**Reference**: `AI_REPORT_PART5.md` - Table 24 (Python ADK)

**Status**: Docs complete (3,000+ lines), code 0%

#### 6.1. **Project Setup + Semantic Router** (1 week)

**Deterministic Semantic Router**:
- Fine-tuning with fixed seed
- Model versioning
- Deterministic inference (temperature=0)

**Tasks**:
- [ ] Poetry setup
  ```toml
  [tool.poetry]
  name = "ventros-adk"
  version = "0.1.0"

  [tool.poetry.dependencies]
  python = "^3.11"
  google-cloud-aiplatform = "^1.38.0"
  transformers = "^4.36.0"
  torch = "^2.1.0"
  ```
- [ ] Google Cloud ADK 0.5+
- [ ] DistilBERT fine-tuning (intent classification)
  ```python
  # Deterministic training
  training_args = TrainingArguments(
      seed=42,  # Fixed seed for reproducibility!
      data_seed=42,
      full_determinism=True,
      output_dir="./models/semantic_router_v1",
      num_train_epochs=10,
      per_device_train_batch_size=32,
  )
  ```
- [ ] Training data (10k messages via synthetic generation with seed)
- [ ] 5 intent classes: sales_prospecting, retention_churn, support_technical, support_billing, general
- [ ] Model versioning (v1, v2, etc.)
- [ ] Evaluation metrics (>92% accuracy on test set)
- [ ] Deterministic inference with fixed seed
- [ ] Tests (same input → same intent classification)

**Effort**: 1 week

---

#### 6.2. **CoordinatorAgent** (1 week)

**Deterministic Agent Dispatch**:
```python
class CoordinatorAgent:
    def __init__(self, semantic_router: SemanticRouter, agents: Dict[str, Agent], clock: Clock):
        self.semantic_router = semantic_router
        self.agents = agents
        self.clock = clock  # Deterministic time!

    def route(self, message: str) -> Agent:
        # Deterministic intent classification
        intent, confidence = self.semantic_router.classify(message, temperature=0.0, seed=42)

        # Deterministic fallback logic
        if confidence < 0.75:
            return self.agents["balanced"]  # Fallback

        # Deterministic routing
        return self.agents.get(intent, self.agents["balanced"])
```

**Tasks**:
- [ ] Semantic Router integration
- [ ] Agent dispatch logic (deterministic with confidence threshold)
- [ ] Fallback strategy (BalancedAgent)
- [ ] Agent lifecycle management
- [ ] Routing metrics (track accuracy)
- [ ] Tests with mocked semantic router (deterministic)
- [ ] Golden file tests (same message → same agent)

**Effort**: 1 week

---

#### 6.3. **Specialist Agents** (2 weeks)
**5 agents** (all with deterministic behavior)

**Deterministic Agent Pattern**:
```python
class SalesProspectingAgent(Agent):
    def __init__(self, llm_client: LLMClient, memory_client: MemoryClient, tools: ToolRegistry, clock: Clock):
        self.llm_client = llm_client
        self.memory_client = memory_client
        self.tools = tools
        self.clock = clock
        self.system_prompt = self.load_system_prompt()  # Deterministic prompt

    async def process(self, message: str, context: Context) -> AgentResponse:
        # 1. Retrieve memory (deterministic)
        memory = await self.memory_client.search_memory(
            tenant_id=context.tenant_id,
            contact_id=context.contact_id,
            query=message,
            max_results=20  # Fixed for determinism
        )

        # 2. Build prompt (deterministic)
        prompt = self.build_prompt(message, memory, context)

        # 3. LLM inference (deterministic with temperature=0)
        response = await self.llm_client.generate(
            prompt=prompt,
            temperature=0.0,  # Deterministic!
            seed=42,
            model="gemini-1.5-flash-002"  # Pin version
        )

        # 4. Execute tools (if needed)
        if response.tool_calls:
            tool_results = await self.execute_tools(response.tool_calls)
            # ... process tool results

        return AgentResponse(
            text=response.text,
            confidence=response.confidence,
            timestamp=self.clock.now()  # Deterministic!
        )
```

**Agents to implement**:
- [ ] **SalesProspectingAgent** (lead qualification, pipeline updates)
  - System prompt with deterministic instructions
  - Tools: qualify_lead, update_pipeline_stage, schedule_follow_up
  - Memory strategy: graph_weight=0.40 (campaign attribution)
  - Tests with fixtures

- [ ] **RetentionChurnAgent** (churn prediction, win-back)
  - System prompt focused on retention
  - Tools: get_churn_prediction, send_message, tag_contact
  - Memory strategy: vector_weight=0.50 (semantic understanding)
  - Tests with golden files

- [ ] **SupportTechnicalAgent** (KB search, escalation)
  - System prompt for technical support
  - Tools: search_documents, create_note, assign_to_agent
  - Memory strategy: keyword_weight=0.40 (exact matches for error codes)
  - Tests with deterministic KB search

- [ ] **SupportBillingAgent** (invoices, subscriptions)
  - System prompt for billing queries
  - Tools: get_invoice, get_subscription_status, create_note
  - Memory strategy: balanced weights
  - Tests with billing fixtures

- [ ] **BalancedAgent** (fallback general-purpose)
  - Generic system prompt
  - All tools available
  - Balanced memory strategy
  - Tests covering edge cases

**Each agent includes**:
- Deterministic system prompt (versioned)
- Memory service integration (gRPC)
- Tool registry (30 tools)
- LLM calls with temperature=0 and seed
- Response caching (same input → cached response)
- Tests with fixtures and golden files

**Effort**: 2 weeks

---

#### 6.4. **Tool Registry + RabbitMQ** (1 week)

**Deterministic Tool Execution**:
```python
class ToolRegistry:
    def __init__(self, grpc_client: GRPCClient, clock: Clock):
        self.grpc_client = grpc_client
        self.clock = clock
        self.tools = self._register_tools()

    async def execute(self, tool_name: str, params: dict) -> dict:
        tool = self.tools[tool_name]

        # Add deterministic timestamp
        params["timestamp"] = self.clock.now().isoformat()

        # Execute via gRPC
        result = await tool.execute(params)

        return result
```

**Tasks**:
- [ ] 30 tools wrapped (delegate to Go via gRPC)
- [ ] gRPC calls to Go backend
- [ ] Error handling with retry logic
- [ ] Tool response caching
- [ ] RabbitMQ consumer (message.created)
  - Consume messages deterministically (one per contact at a time)
  - Idempotent processing (check processed_events)
- [ ] RabbitMQ publisher (agent response)
- [ ] Deterministic tests with mocked gRPC
- [ ] Golden file tests

**Effort**: 1 week

---

#### 6.5. **Temporal + Observability** (1 week)

**Deterministic Workflows**:
```python
@workflow.defn
class ProcessMessageWorkflow:
    @workflow.run
    async def run(self, message_id: str) -> str:
        # Deterministic workflow with versioning
        workflow_time = workflow.now()  # Deterministic time in Temporal!

        # Activities are deterministic and retryable
        agent_response = await workflow.execute_activity(
            process_message_activity,
            args=[message_id],
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3)
        )

        return agent_response
```

**Tasks**:
- [ ] Temporal workflow definitions (long-running tasks)
  - ProcessMessageWorkflow
  - CampaignExecutionWorkflow
  - ChurnPredictionWorkflow
- [ ] Temporal activities (idempotent, retryable)
- [ ] Workflow versioning
- [ ] Phoenix observability (tracing)
  - Trace all LLM calls
  - Trace all tool executions
  - Trace memory searches
- [ ] Dashboard (Grafana)
- [ ] Alerts (error rate, latency, cost)
- [ ] Deterministic workflow tests
- [ ] Tests

**Effort**: 1 week

**Total Effort Sprint 19-24**: 6 weeks

---

### **Additional P1 Tasks** (parallel with AI work)

#### 7. **Testing Coverage** (2 weeks)
**Reference**: `AI_REPORT_PART5.md` - Table 22

**Current**: 82% declared, but real ~45% (application layer low)

**Deterministic Testing Strategy**:
- All tests use FrozenClock
- All tests use SequentialIDGenerator
- All tests use fixtures
- All tests verify deterministic ordering
- Golden file tests for complex outputs

**Tasks**:
- [ ] Create test infrastructure (FrozenClock, SequentialIDGenerator, fixtures)
- [ ] 40+ unit tests (19 use cases without tests)
  - All using deterministic helpers
  - All with clear arrange/act/assert structure
- [ ] 10 integration tests (only 2 exist, need 20% pyramid)
  - Database tests with fixtures
  - RabbitMQ tests with deterministic event ordering
  - Cache tests with deterministic keys
- [ ] 5 E2E tests (campaign, sequence, memory, billing)
  - Full stack tests with deterministic data
  - Golden file assertions
- [ ] Add test coverage reporting (show % coverage)
- [ ] Target: 85% overall, 100% deterministic behavior

**Effort**: 2 weeks

---

#### 8. **Resilience Patterns** (2 weeks)
**Reference**: `AI_REPORT_PART5.md` - Table 23

**Current**: Circuit breaker exists but only 10% coverage

**Deterministic Circuit Breaker**:
```go
type CircuitBreaker struct {
    maxFailures  int
    resetTimeout time.Duration
    clock        Clock // Deterministic!
    state        State
    failures     int
    lastFailTime time.Time
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        if cb.clock.Now().Sub(cb.lastFailTime) > cb.resetTimeout {
            cb.state = HalfOpen
        } else {
            return ErrCircuitOpen
        }
    }

    err := fn()

    if err != nil {
        cb.failures++
        cb.lastFailTime = cb.clock.Now()
        if cb.failures >= cb.maxFailures {
            cb.state = Open
        }
        return err
    }

    cb.failures = 0
    cb.state = Closed
    return nil
}
```

**Tasks**:
- [ ] Enhance CircuitBreaker with Clock injection
- [ ] Circuit breaker in 4 external APIs:
  - Vertex AI (image OCR, embeddings)
  - Stripe (billing, subscriptions)
  - WAHA (WhatsApp messaging)
  - LlamaParse (document parsing)
- [ ] Retry logic (exponential backoff with jitter)
  ```go
  func RetryWithBackoff(fn func() error, maxRetries int, clock Clock) error {
      backoff := time.Second
      for i := 0; i < maxRetries; i++ {
          err := fn()
          if err == nil {
              return nil
          }

          // Deterministic backoff (no jitter in tests)
          time.Sleep(backoff)
          backoff *= 2
      }
      return ErrMaxRetriesExceeded
  }
  ```
- [ ] Timeout in all APIs (10s default, configurable)
  ```go
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()
  ```
- [ ] Bulkhead pattern (optional P2)
  - Limit concurrent requests per API
  - Use semaphore with fixed size
- [ ] Metrics (circuit breaker state, retry count, timeout count)
- [ ] Deterministic tests (FrozenClock, verify state transitions)
- [ ] Tests (open, half-open, closed states; retries; timeouts)

**Effort**: 2 weeks

---

#### 9. **Optimistic Locking** (1 week)
**Reference**: `AI_REPORT_PART1.md` - Table 5

**Current**: Only 16/30 aggregates (53%)

**Missing (14 aggregates)**:
- [ ] MessageGroup
- [ ] Note
- [ ] Tracking
- [ ] **UsageMeter** (billing critical! Race condition risk)
- [ ] WebhookSubscription
- [ ] SagaTracker
- [ ] MessageEnrichment
- [ ] ChannelType
- [ ] DomainEventLog
- [ ] OutboxEvent
- [ ] ContactEvent
- [ ] CustomField
- [ ] (2 more - audit codebase)

**Deterministic Optimistic Locking Pattern**:
```go
// Domain aggregate
type UsageMeter struct {
    id          uuid.UUID
    version     int  // REQUIRED for optimistic locking
    usage       int64
    lastUpdated time.Time
}

// Repository
func (r *UsageMeterRepository) Save(ctx context.Context, meter *UsageMeter) error {
    // Deterministic update with version check
    result := r.db.
        Model(&UsageMeterEntity{}).
        Where("id = ? AND version = ?", meter.ID(), meter.Version()).
        Updates(map[string]interface{}{
            "version":      meter.Version() + 1,  // Increment
            "usage":        meter.Usage(),
            "last_updated": meter.LastUpdated(),
        })

    if result.RowsAffected == 0 {
        return ErrConcurrentUpdateConflict
    }

    return nil
}

// Application layer - retry on conflict
func (h *IncrementUsageHandler) Handle(ctx context.Context, cmd IncrementUsageCommand) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        meter, _ := h.repo.FindByID(ctx, cmd.MeterID)
        meter.IncrementUsage(cmd.Amount)

        err := h.repo.Save(ctx, meter)
        if err == nil {
            return nil // Success
        }

        if err != ErrConcurrentUpdateConflict {
            return err // Other error
        }

        // Retry on conflict (deterministic with exponential backoff)
        time.Sleep(time.Millisecond * time.Duration(math.Pow(2, float64(i))))
    }

    return ErrMaxRetriesExceeded
}
```

**Tasks**:
- [ ] Add `version int` field to 14 aggregates
- [ ] Migration for each aggregate (ADD COLUMN version INT NOT NULL DEFAULT 1)
- [ ] Update repository Update() methods
- [ ] Retry logic on optimistic lock failure (exponential backoff)
- [ ] Metrics (conflict count, retry count)
- [ ] Deterministic tests (simulate concurrent updates)
- [ ] Load tests (verify no data loss under concurrency)

**Effort**: 1 week

---

## 🟢 PRIORITY 2: IMPROVEMENTS (12+ weeks)

### 10. **Value Objects** (2 weeks)
**Reference**: `AI_REPORT_PART2.md` - Table 6

**Current**: Only 12 VOs, many primitive obsession cases

**Critical**:
- [ ] **Money VO** (P0): Use int64 (cents) instead of float64 for billing
  ```go
  type Money struct {
      cents    int64  // Deterministic! No floating point errors
      currency string // USD, BRL, EUR
  }

  func NewMoney(cents int64, currency string) Money {
      return Money{cents: cents, currency: currency}
  }

  func (m Money) Add(other Money) (Money, error) {
      if m.currency != other.currency {
          return Money{}, ErrCurrencyMismatch
      }
      return Money{cents: m.cents + other.cents, currency: m.currency}, nil
  }

  func (m Money) ToCents() int64 {
      return m.cents // Always deterministic
  }
  ```
  - Impact: Financial accuracy, no floating point errors
  - Locations: Invoice, Subscription, UsageMeter, Transaction
  - Migration: Convert DECIMAL to BIGINT (cents)
  - Tests (deterministic arithmetic: $1.20 + $0.80 = $2.00)
  - Effort: 1 week

**Important (P1)**:
- [ ] Email VO (RFC 5322 validation)
- [ ] PhoneNumber VO (E.164 validation)
- [ ] URL VO (SSRF prevention, deterministic normalization)
- [ ] Duration VO (type safety, deterministic comparison)
- [ ] Timezone VO (IANA validation)
- [ ] CustomField VO (make immutable)
- [ ] ChannelConfig VO (strong typing instead of map)

**Effort**: 2 weeks total

---

### 11. **Knowledge Graph (Apache AGE)** (2 weeks)
**Reference**: `AI_REPORT_PART5.md` - Table 21

**Deterministic Graph Queries**:
```cypher
-- Always return results in deterministic order
MATCH (c:Contact)-[r:HAS_SESSION]->(s:Session)
WHERE c.tenant_id = $tenant_id
RETURN c, r, s
ORDER BY s.started_at DESC, s.id ASC  -- Deterministic!
LIMIT 20;
```

**Tasks**:
- [ ] Install Apache AGE extension
- [ ] Migration 000054: Graph schema
  - Nodes: Contact, Session, Message, Offer, Campaign, Tag
  - Edges: HAS_SESSION, CONTAINS, RECEIVED_OFFER, IN_CAMPAIGN, REPLY_TO, HAS_TAG
- [ ] Cypher queries (graph traversal with deterministic ORDER BY)
  - Find conversation threads (deterministic traversal)
  - Find similar contacts (deterministic similarity + ID sort)
  - Campaign attribution (deterministic path finding)
- [ ] Graph indexing (for performance)
- [ ] Graph query service
- [ ] Tests with fixtures (golden file tests for graph queries)

**Effort**: 2 weeks

---

### 12. **Observability** (2 weeks)
**Reference**: `AI_REPORT_PART1.md` - Table 1 (score 5.5/10)

**Deterministic Metrics**:
- Counter: Always increment (deterministic)
- Gauge: Set to exact value (deterministic)
- Histogram: Bucket boundaries fixed (deterministic)

**Tasks**:
- [ ] Prometheus metrics
  - Request count by endpoint
  - Latency histogram (p50, p95, p99)
  - Error rate
  - Cache hit rate
  - AI cost per tenant
  - Circuit breaker state
- [ ] OpenTelemetry tracing
  - Trace ID propagation (deterministic with request_id)
  - Span naming convention
  - Trace all external calls (DB, Redis, RabbitMQ, AI)
- [ ] Grafana dashboards
  - API performance
  - Database performance
  - Cache performance
  - AI costs
  - Error rates
- [ ] Alerts (PagerDuty/Slack)
  - Error rate > 5%
  - Latency p95 > 1s
  - Cache hit rate < 50%
  - Daily AI cost > $100
  - Circuit breaker open
- [ ] APM (New Relic/Datadog - optional)
- [ ] Log aggregation (structured JSON logs)

**Effort**: 2 weeks

---

### 13. **Agent Templates** (1 week)
**Reference**: `AI_REPORT_PART6.md` - Roadmap Sprint 30

**Deterministic Agent Templates**:
```go
type AgentTemplate struct {
    ID             uuid.UUID
    Name           string
    Version        int // Versioned for determinism
    SystemPrompt   string // Deterministic prompt
    Tools          []string
    MemoryStrategy RetrievalStrategy
    LLMConfig      LLMConfig // temperature, seed, model
}

// Templates are versioned and immutable
const (
    SalesProspectingV1 = "sales_prospecting_v1"
    RetentionChurnV1   = "retention_churn_v1"
    SupportTechnicalV1 = "support_technical_v1"
)
```

**Tasks**:
- [ ] Agent template registry (versioned)
- [ ] 10 system agent templates:
  1. sales_prospecting_v1
  2. retention_churn_v1
  3. support_technical_v1
  4. support_billing_v1
  5. onboarding_v1
  6. upsell_v1
  7. feedback_collection_v1
  8. appointment_booking_v1
  9. faq_v1
  10. general_assistant_v1
- [ ] Template discovery API (GET /agent-templates)
- [ ] Template instantiation (POST /agents)
- [ ] Template versioning (v1, v2, etc.)
- [ ] Template validation
- [ ] Documentation with examples
- [ ] Tests

**Effort**: 1 week

---

### 14. **Chat Entity** (1 week)
**Status**: ✅ Fully implemented (discovery in report)
**Location**: `internal/domain/crm/chat/chat.go`

**Validation tasks**:
- [ ] Verify Chat tests (exists)
- [ ] Add Chat to API docs (Swagger)
- [ ] Integration tests
  - Create chat
  - Add/remove participants (deterministic order)
  - Send messages (deterministic ordering)
  - Archive chat
- [ ] Documentation update
- [ ] Golden file tests

**Effort**: 3 days (validation only, already implemented)

---

### 15. **WAHA Integration Improvements** (1 week)
**Current**: Only sending messages implemented

**Deterministic WAHA Integration**:
- Message deduplication (idempotent by message_id)
- Event ordering (sequence numbers)
- Deterministic webhook processing

**Tasks**:
- [ ] Fetch message history (GET /api/{session}/messages)
  - Pagination with deterministic ordering
  - Deduplication by message_id
- [ ] Handle all webhook events (idempotent):
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
  - Location (lat/long validation)
  - Contact card
  - Poll
  - Buttons (WhatsApp Business)
  - List (WhatsApp Business)
  - Template messages
- [ ] Webhook signature verification (security)
- [ ] Idempotent webhook processing (check processed_events)
- [ ] Deterministic tests with mocked WAHA API

**Effort**: 1 week

---

## 📅 EXECUTION ROADMAP (31 SPRINTS = 7.75 MONTHS)

### **Phase 0: Deterministic Foundation** (0-1 week) 🎯 NEW
- Sprint 0: Deterministic patterns (Clock, IDGenerator, Fixtures, Events) (1 sem) 🔴

**Milestone M0** (Week 1): Deterministic foundation complete ✅

---

### **Phase 1: Security & Performance** (1-3 months)
- Sprint 1-2: Security Fixes P0 (3-4 sem) 🔴
- Sprint 3-4: Cache + Performance (2 sem) 🔴

**Milestone M1** (Week 7): Production-safe API ✅

---

### **Phase 2: AI Foundation** (3-6 months)
- Sprint 5-11: Memory Service (7 sem) 🔴
- Sprint 12-14: gRPC API (3 sem) 🔴

**Milestone M2** (Week 18): Memory Service + gRPC ready ✅

---

### **Phase 3: AI Tools** (6-8 months)
- Sprint 15-18: MCP Server (4 sem) 🔴
- Sprint 19-24: Python ADK (6 sem) 🔴

**Milestone M3** (Week 31): Multi-agent system production ✅

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
- ✅ Tests: >85% coverage, 100% deterministic
- ✅ Security: 0 P0 vulnerabilities
- ✅ Latency: <200ms (API), <50ms (cache)
- ✅ Cache hit rate: >70%
- ✅ AI/ML Score: >8.0/10
- ✅ **Determinism Score: >9.0/10** (NEW)

### **Determinism KPIs** (NEW)
- ✅ **Tests deterministic: 100%** (all tests use Clock, IDGenerator, fixtures)
- ✅ **Database queries deterministic: 100%** (all queries have ORDER BY)
- ✅ **Event ordering: 100%** (sequence numbers + idempotent handlers)
- ✅ **AI responses reproducible: >95%** (temperature=0, seed parameter)
- ✅ **Cache keys deterministic: 100%** (hash-based, sorted params)
- ✅ **Time operations deterministic: 100%** (Clock interface everywhere)

### **Business KPIs**
- 🤖 AI agents reduce workload 60%
- 🎯 Lead qualification automatic (+30% conversion)
- 💰 Churn prediction (-20% churn)
- 📈 Memory context improves NPS (+15 points)
- 💵 AI cost tracking prevents billing surprises
- 🔍 Debugging 10x faster (deterministic reproduction)

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
8. **Determinism**: Weak in tests, AI, and time-dependent operations

### **Discoveries**
1. ✅ Chat aggregate 100% implemented (contradicts old docs)
2. ✅ 30 aggregates total (not 23)
3. ✅ 158 endpoints (not "50+")
4. ⚠️ 14 aggregates without optimistic locking (47%)
5. ⚠️ 39/44 use cases without tests (89%)
6. ⚠️ Tests non-deterministic (time.Now(), random UUIDs)
7. ⚠️ Database queries without ORDER BY (non-deterministic results)

---

## 🎯 DECISION POINTS

### **For Basic Production (CRM without advanced AI)**:
✅ **READY after 5 weeks** (Sprint 0-4: Determinism + Security + Cache)
- Deterministic foundation
- Message enrichment works
- CRUD complete
- Event-driven OK

### **For Production with Advanced AI** (Multi-agent, Memory):
❌ **7.75 MONTHS** (31 sprints)
- Deterministic foundation (1 week)
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
- Deterministic Testing: https://martinfowler.com/articles/deterministic-tests.html
- Clock Interface Pattern: https://blog.cleancoder.com/uncle-bob/2015/01/08/IntegrationTests.html

---

## 📝 NOTES

### **Priorities Explained**
- 🔴 **P0 (Critical)**: Blockers for production or severe security issues
- 🟡 **P1 (Important)**: Required for advanced features or quality
- 🟢 **P2 (Improvement)**: Nice to have, optimization
- 🎯 **NEW**: Deterministic patterns (foundation for reliability)

### **Effort Estimates**
Based on detailed analysis in architectural report:
- 1 day = 8 hours
- 1 week = 5 days = 40 hours
- Includes: implementation + tests + documentation

### **Resource Recommendation**
- **Backend Go Engineer**: 2 FTEs (security, cache, memory, deterministic patterns)
- **AI/ML Engineer**: 2 FTEs (Python ADK, semantic router, facts)
- **DevOps Engineer**: 0.5 FTE (infra, observability)
- **QA Engineer**: 1 FTE (testing, E2E, integration, deterministic tests)
- **Total**: 5.5 FTEs

### **Determinism Benefits** 🎯
1. **Debugging**: Bugs are 100% reproducible with same input
2. **Testing**: Tests never flake, always pass or fail deterministically
3. **Auditing**: Exact replay of past behavior for compliance
4. **Performance**: Easier to benchmark (no variance from randomness)
5. **AI**: Reproducible LLM outputs for quality assurance
6. **Reliability**: System behavior predictable and consistent

---

**Last Update**: 2025-10-14 (Enhanced with deterministic patterns)
**Next Review**: After Sprint 0 (M0: Deterministic Foundation)
**Maintainer**: Ventros CRM Team
**Status**: ✅ Consolidated TODO with deterministic enhancements
**Determinism Score**: 🎯 Target 9.0/10 (from current 7.5/10)
