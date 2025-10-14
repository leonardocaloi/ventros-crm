# TODO_PYTHON - Python Projects Roadmap

## 📋 Python AI/ML Projects Implementation Guide

**Created**: 2025-10-13
**Status**: 🚀 **Starting from Zero** - Complete Implementation Roadmap
**Backend Reference**: Go CRM (production-ready, 9/10 quality)
**AI Status**: 0% implemented (only basic enrichments)

**Related Documentation**:
- `AI_REPORT.md` - Complete AI/ML status report
- `TODO.md` - Main project TODO (backend Go)
- `docs/MCP_SERVER_COMPLETE.md` - MCP Server architecture (1175 lines)
- `docs/PYTHON_ADK_ARCHITECTURE.md` - Python ADK architecture (3000+ lines)
- `docs/AI_MEMORY_GO_ARCHITECTURE.md` - Memory service architecture

---

## 🎯 OVERVIEW

This document provides a **complete from-zero implementation roadmap** for:

1. **MCP Server (Go)** - Model Context Protocol server exposing CRM via tools
2. **Python ADK** - Multi-agent system using Google Cloud ADK 0.5+
3. **gRPC API** - Go ↔ Python communication layer
4. **Memory Service** - Hybrid search (vector + keyword + graph + SQL)

**Total Effort**: **18-27 weeks** (~4-6 months)

**Architecture**:
```
┌─────────────────────────────────────────────────────────────────┐
│                        VENTROS CRM ECOSYSTEM                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐ │
│  │   Go Backend │◄────►│ MCP Server   │◄────►│ Claude       │ │
│  │   (CRM Core) │      │   (Go)       │      │ Desktop      │ │
│  └──────────────┘      └──────────────┘      └──────────────┘ │
│         ▲                                                        │
│         │                                                        │
│         │ gRPC                                                   │
│         │                                                        │
│         ▼                                                        │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐ │
│  │ Memory       │◄────►│ Python ADK   │◄────►│ RabbitMQ     │ │
│  │ Service (Go) │      │ Multi-Agent  │      │ Events       │ │
│  └──────────────┘      └──────────────┘      └──────────────┘ │
│         │                      │                                │
│         ▼                      ▼                                │
│  ┌──────────────┐      ┌──────────────┐                        │
│  │ PostgreSQL   │      │ Temporal     │                        │
│  │ + pgvector   │      │ Workflows    │                        │
│  └──────────────┘      └──────────────┘                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📊 PROJECT PHASES

### Phase 1: Memory Service (Go) - 4 weeks
**Priority**: 🔴 **CRÍTICO**
**Effort**: 3-4 semanas
**Language**: Go
**Deliverable**: Hybrid search para AI agents

### Phase 2: MCP Server (Go) - 3 weeks
**Priority**: 🔴 **CRÍTICO**
**Effort**: 3-4 semanas
**Language**: Go
**Deliverable**: Claude Desktop acessa CRM

### Phase 3: gRPC API (Go ↔ Python) - 2 weeks
**Priority**: 🟠 **ALTO**
**Effort**: 1-2 semanas
**Language**: Go + Python
**Deliverable**: Python ADK chama SearchMemory()

### Phase 4: Python ADK - 6 weeks
**Priority**: 🔴 **CRÍTICO**
**Effort**: 4-6 semanas
**Language**: Python
**Deliverable**: Multi-agent system funcional

### Phase 5: Advanced Memory - 5 weeks
**Priority**: 🟡 **MÉDIO**
**Effort**: 2-3 semanas
**Language**: Go + SQL
**Deliverable**: Knowledge graph + facts extraction

### Phase 6: Templates & Polish - 2 weeks
**Priority**: 🟢 **BAIXO**
**Effort**: 1-2 semanas
**Language**: Go
**Deliverable**: Agent templates registry

---

## 🔴 PHASE 1: MEMORY SERVICE (Go) - 4 WEEKS

### 1.1. Database Schema (3 days)

#### **Migration 000050: memory_embeddings**

```sql
-- infrastructure/database/migrations/000050_create_memory_embeddings.up.sql

-- Install pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Memory embeddings table
CREATE TABLE memory_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Source tracking
    contact_id UUID REFERENCES contacts(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,

    -- Content
    content_type VARCHAR(50) NOT NULL, -- message, note, document, custom
    content_text TEXT NOT NULL,
    content_hash VARCHAR(64) NOT NULL, -- SHA256 for deduplication

    -- Vector embedding
    embedding vector(768) NOT NULL, -- text-embedding-005 dimensionality

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT content_text_not_empty CHECK (length(content_text) > 0),
    CONSTRAINT content_type_valid CHECK (content_type IN ('message', 'note', 'document', 'custom'))
);

-- Indexes
CREATE INDEX idx_memory_embeddings_tenant ON memory_embeddings(tenant_id);
CREATE INDEX idx_memory_embeddings_project ON memory_embeddings(project_id);
CREATE INDEX idx_memory_embeddings_contact ON memory_embeddings(contact_id);
CREATE INDEX idx_memory_embeddings_session ON memory_embeddings(session_id);
CREATE INDEX idx_memory_embeddings_message ON memory_embeddings(message_id);
CREATE INDEX idx_memory_embeddings_content_hash ON memory_embeddings(content_hash);
CREATE INDEX idx_memory_embeddings_created_at ON memory_embeddings(created_at DESC);

-- Vector index (IVFFlat)
-- For better performance, tune 'lists' based on data size:
-- - Small datasets (<100K): lists = 100
-- - Medium datasets (100K-1M): lists = 1000
-- - Large datasets (>1M): lists = 10000
CREATE INDEX idx_memory_embeddings_vector ON memory_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Full-text search index (for keyword search)
ALTER TABLE memory_embeddings ADD COLUMN content_tsv tsvector
GENERATED ALWAYS AS (to_tsvector('portuguese', content_text)) STORED;
CREATE INDEX idx_memory_embeddings_tsv ON memory_embeddings USING GIN(content_tsv);

-- Metadata GIN index (for JSONB queries)
CREATE INDEX idx_memory_embeddings_metadata ON memory_embeddings USING GIN(metadata);

COMMENT ON TABLE memory_embeddings IS 'Vector embeddings storage for hybrid search';
COMMENT ON COLUMN memory_embeddings.embedding IS 'text-embedding-005 (768 dimensions)';
COMMENT ON COLUMN memory_embeddings.content_hash IS 'SHA256 for deduplication';
COMMENT ON INDEX idx_memory_embeddings_vector IS 'IVFFlat index for approximate nearest neighbor search';
```

#### **Migration 000051: memory_facts**

```sql
-- infrastructure/database/migrations/000051_create_memory_facts.up.sql

CREATE TABLE memory_facts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,

    -- Fact details
    fact_type VARCHAR(50) NOT NULL, -- budget, preference, objection, pain_point, decision_maker
    fact_text TEXT NOT NULL,
    fact_category VARCHAR(50), -- sales, retention, support, billing

    -- Confidence & validation
    confidence FLOAT NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    validated BOOLEAN DEFAULT FALSE,

    -- Source tracking
    source_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    source_embedding_id UUID REFERENCES memory_embeddings(id) ON DELETE SET NULL,

    -- Extraction metadata
    extracted_by VARCHAR(100) NOT NULL, -- llm_model_name or 'manual'
    extraction_metadata JSONB DEFAULT '{}',

    -- Timestamps
    extracted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    validated_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Soft delete for contradiction resolution
    deleted_at TIMESTAMP,
    superseded_by_id UUID REFERENCES memory_facts(id) ON DELETE SET NULL,

    CONSTRAINT fact_text_not_empty CHECK (length(fact_text) > 0)
);

-- Indexes
CREATE INDEX idx_memory_facts_tenant ON memory_facts(tenant_id);
CREATE INDEX idx_memory_facts_project ON memory_facts(project_id);
CREATE INDEX idx_memory_facts_contact ON memory_facts(contact_id);
CREATE INDEX idx_memory_facts_type ON memory_facts(fact_type);
CREATE INDEX idx_memory_facts_category ON memory_facts(fact_category);
CREATE INDEX idx_memory_facts_confidence ON memory_facts(confidence DESC);
CREATE INDEX idx_memory_facts_validated ON memory_facts(validated);
CREATE INDEX idx_memory_facts_deleted ON memory_facts(deleted_at) WHERE deleted_at IS NULL;

COMMENT ON TABLE memory_facts IS 'Extracted facts from conversations (NER)';
COMMENT ON COLUMN memory_facts.confidence IS 'LLM confidence score (0.0-1.0)';
COMMENT ON COLUMN memory_facts.superseded_by_id IS 'Contradiction resolution - points to newer fact';
```

#### **Migration 000052: retrieval_strategies**

```sql
-- infrastructure/database/migrations/000052_create_retrieval_strategies.up.sql

CREATE TABLE retrieval_strategies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Strategy identification
    strategy_name VARCHAR(100) NOT NULL UNIQUE,
    strategy_category VARCHAR(50) NOT NULL, -- sales, retention, support, billing

    -- Weights configuration
    vector_weight FLOAT NOT NULL CHECK (vector_weight >= 0 AND vector_weight <= 1),
    keyword_weight FLOAT NOT NULL CHECK (keyword_weight >= 0 AND keyword_weight <= 1),
    graph_weight FLOAT NOT NULL CHECK (graph_weight >= 0 AND graph_weight <= 1),
    sql_baseline_weight FLOAT NOT NULL DEFAULT 0.1,

    -- Options
    use_reranking BOOLEAN DEFAULT FALSE,
    use_facts BOOLEAN DEFAULT TRUE,
    max_results INT DEFAULT 20,

    -- Configuration metadata
    description TEXT,
    config_metadata JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT weights_sum_valid CHECK (
        vector_weight + keyword_weight + graph_weight + sql_baseline_weight <= 1.1
    )
);

-- Default strategies
INSERT INTO retrieval_strategies (strategy_name, strategy_category, vector_weight, keyword_weight, graph_weight, use_reranking, description) VALUES
('sales_prospecting', 'sales', 0.20, 0.30, 0.40, FALSE, 'Focus on campaign attribution via graph traversal'),
('retention_churn', 'retention', 0.50, 0.20, 0.20, TRUE, 'High vector weight for pattern matching, reranking critical'),
('support_technical', 'support', 0.30, 0.50, 0.10, FALSE, 'Keyword-heavy for exact technical terms'),
('support_billing', 'support', 0.25, 0.35, 0.30, FALSE, 'Balanced approach with transaction graph'),
('general_balanced', 'general', 0.33, 0.33, 0.34, FALSE, 'Balanced retrieval for generic queries');

CREATE INDEX idx_retrieval_strategies_category ON retrieval_strategies(strategy_category);

COMMENT ON TABLE retrieval_strategies IS 'Configurable retrieval strategies for hybrid search';
```

---

### 1.2. Domain Layer (5 days)

#### **Files to Create**:

```
internal/domain/memory/
├── memory_embedding.go          ❌ NEW - MemoryEmbedding aggregate root
├── memory_fact.go               ❌ NEW - MemoryFact aggregate root
├── retrieval_strategy.go        ❌ NEW - RetrievalStrategy value object
├── search_result.go             ❌ NEW - SearchResult value object
├── events.go                    ❌ NEW - Memory domain events
├── repository.go                ❌ NEW - Memory repository interface
├── errors.go                    ❌ NEW - Memory errors
└── memory_test.go               ❌ NEW - Unit tests
```

#### **memory_embedding.go**

```go
package memory

import (
    "crypto/sha256"
    "encoding/hex"
    "time"

    "github.com/google/uuid"
    "github.com/ventros/crm/internal/domain/shared"
)

// ContentType represents the type of content stored
type ContentType string

const (
    ContentTypeMessage  ContentType = "message"
    ContentTypeNote     ContentType = "note"
    ContentTypeDocument ContentType = "document"
    ContentTypeCustom   ContentType = "custom"
)

// MemoryEmbedding represents a vector embedding of content for semantic search
type MemoryEmbedding struct {
    // Identity
    id        uuid.UUID
    tenantID  string
    projectID uuid.UUID

    // Source tracking
    contactID *uuid.UUID
    sessionID *uuid.UUID
    messageID *uuid.UUID

    // Content
    contentType ContentType
    contentText string
    contentHash string // SHA256

    // Vector
    embedding []float32 // 768 dimensions for text-embedding-005

    // Metadata
    metadata map[string]interface{}

    // Timestamps
    createdAt time.Time
    updatedAt time.Time

    // Event tracking
    events []shared.DomainEvent
}

// NewMemoryEmbedding creates a new memory embedding
func NewMemoryEmbedding(
    tenantID string,
    projectID uuid.UUID,
    contentType ContentType,
    contentText string,
    embedding []float32,
) (*MemoryEmbedding, error) {
    if tenantID == "" {
        return nil, ErrInvalidTenantID
    }
    if projectID == uuid.Nil {
        return nil, ErrInvalidProjectID
    }
    if contentText == "" {
        return nil, ErrEmptyContent
    }
    if len(embedding) != 768 {
        return nil, ErrInvalidEmbeddingDimension
    }

    now := time.Now()
    me := &MemoryEmbedding{
        id:          uuid.New(),
        tenantID:    tenantID,
        projectID:   projectID,
        contentType: contentType,
        contentText: contentText,
        contentHash: computeContentHash(contentText),
        embedding:   embedding,
        metadata:    make(map[string]interface{}),
        createdAt:   now,
        updatedAt:   now,
        events:      []shared.DomainEvent{},
    }

    me.addEvent(NewMemoryEmbeddingCreatedEvent(me))
    return me, nil
}

// computeContentHash computes SHA256 hash for deduplication
func computeContentHash(content string) string {
    hash := sha256.Sum256([]byte(content))
    return hex.EncodeToString(hash[:])
}

// LinkToContact associates embedding with a contact
func (me *MemoryEmbedding) LinkToContact(contactID uuid.UUID) {
    me.contactID = &contactID
    me.updatedAt = time.Now()
    me.addEvent(NewMemoryEmbeddingLinkedToContactEvent(me, contactID))
}

// LinkToSession associates embedding with a session
func (me *MemoryEmbedding) LinkToSession(sessionID uuid.UUID) {
    me.sessionID = &sessionID
    me.updatedAt = time.Now()
}

// LinkToMessage associates embedding with a message
func (me *MemoryEmbedding) LinkToMessage(messageID uuid.UUID) {
    me.messageID = &messageID
    me.updatedAt = time.Now()
}

// SetMetadata sets metadata key-value
func (me *MemoryEmbedding) SetMetadata(key string, value interface{}) {
    me.metadata[key] = value
    me.updatedAt = time.Now()
}

// GetMetadata retrieves metadata value
func (me *MemoryEmbedding) GetMetadata(key string) (interface{}, bool) {
    val, ok := me.metadata[key]
    return val, ok
}

// Getters
func (me *MemoryEmbedding) ID() uuid.UUID              { return me.id }
func (me *MemoryEmbedding) TenantID() string           { return me.tenantID }
func (me *MemoryEmbedding) ProjectID() uuid.UUID       { return me.projectID }
func (me *MemoryEmbedding) ContactID() *uuid.UUID      { return me.contactID }
func (me *MemoryEmbedding) SessionID() *uuid.UUID      { return me.sessionID }
func (me *MemoryEmbedding) MessageID() *uuid.UUID      { return me.messageID }
func (me *MemoryEmbedding) ContentType() ContentType   { return me.contentType }
func (me *MemoryEmbedding) ContentText() string        { return me.contentText }
func (me *MemoryEmbedding) ContentHash() string        { return me.contentHash }
func (me *MemoryEmbedding) Embedding() []float32       { return me.embedding }
func (me *MemoryEmbedding) Metadata() map[string]interface{} { return me.metadata }
func (me *MemoryEmbedding) CreatedAt() time.Time       { return me.createdAt }
func (me *MemoryEmbedding) UpdatedAt() time.Time       { return me.updatedAt }
func (me *MemoryEmbedding) Events() []shared.DomainEvent { return me.events }
func (me *MemoryEmbedding) ClearEvents()               { me.events = []shared.DomainEvent{} }

func (me *MemoryEmbedding) addEvent(event shared.DomainEvent) {
    me.events = append(me.events, event)
}
```

---

### 1.3. Infrastructure Layer (7 days)

#### **Files to Create**:

```
infrastructure/ai/
├── vertex_embedding_provider.go  ✅ EXISTS (reuse)
├── embedding_service.go          ❌ NEW - Embedding generation service
└── embedding_worker.go           ❌ NEW - Background worker

infrastructure/persistence/
├── gorm_memory_embedding_repository.go  ❌ NEW
├── gorm_memory_fact_repository.go       ❌ NEW
└── entities/memory.go                   ❌ NEW

infrastructure/memory/
├── hybrid_search_service.go      ❌ NEW - Core hybrid search
├── rrf_fusion.go                 ❌ NEW - Reciprocal Rank Fusion
├── vector_search.go              ❌ NEW - pgvector queries
├── keyword_search.go             ❌ NEW - PostgreSQL full-text
├── graph_search.go               ❌ NEW - Graph traversal (placeholder for Phase 5)
├── search_cache.go               ❌ NEW - Redis cache layer
└── reranking_service.go          ❌ NEW - Optional reranking (Jina v2)
```

#### **hybrid_search_service.go**

```go
package memory

import (
    "context"
    "fmt"
    "sort"

    "github.com/google/uuid"
    "github.com/ventros/crm/internal/domain/memory"
)

// HybridSearchService orchestrates hybrid search across multiple strategies
type HybridSearchService struct {
    vectorSearch  *VectorSearchService
    keywordSearch *KeywordSearchService
    graphSearch   *GraphSearchService // Phase 5
    cache         *SearchCache
    rrfFusion     *RRFFusion
}

// SearchRequest encapsulates search parameters
type SearchRequest struct {
    TenantID   string
    ProjectID  uuid.UUID
    Query      string
    ContactID  *uuid.UUID
    SessionID  *uuid.UUID
    Strategy   *memory.RetrievalStrategy
    MaxResults int
}

// SearchResult represents a single search result with score
type SearchResult struct {
    EmbeddingID uuid.UUID
    ContentText string
    ContentType memory.ContentType
    Score       float32
    Source      string // "vector", "keyword", "graph", "sql"
    Metadata    map[string]interface{}
}

// Search performs hybrid search combining multiple strategies
func (s *HybridSearchService) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
    // 1. Check cache
    cacheKey := s.buildCacheKey(req)
    if cached, found := s.cache.Get(ctx, cacheKey); found {
        return cached, nil
    }

    // 2. SQL Baseline - always get last N messages
    sqlResults, err := s.getSQLBaseline(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("sql baseline failed: %w", err)
    }

    // 3. Vector Search
    var vectorResults []*SearchResult
    if req.Strategy.VectorWeight > 0 {
        vectorResults, err = s.vectorSearch.Search(ctx, req)
        if err != nil {
            return nil, fmt.Errorf("vector search failed: %w", err)
        }
    }

    // 4. Keyword Search
    var keywordResults []*SearchResult
    if req.Strategy.KeywordWeight > 0 {
        keywordResults, err = s.keywordSearch.Search(ctx, req)
        if err != nil {
            return nil, fmt.Errorf("keyword search failed: %w", err)
        }
    }

    // 5. Graph Search (Phase 5)
    var graphResults []*SearchResult
    if req.Strategy.GraphWeight > 0 && s.graphSearch != nil {
        graphResults, err = s.graphSearch.Search(ctx, req)
        if err != nil {
            // Non-critical - log and continue
            // logger.Warn("graph search failed", zap.Error(err))
        }
    }

    // 6. RRF Fusion - combine results
    fusedResults := s.rrfFusion.Fuse(
        sqlResults,
        vectorResults,
        keywordResults,
        graphResults,
        req.Strategy,
    )

    // 7. Reranking (optional)
    if req.Strategy.UseReranking {
        // TODO: Implement Jina v2 reranking
    }

    // 8. Limit results
    if len(fusedResults) > req.MaxResults {
        fusedResults = fusedResults[:req.MaxResults]
    }

    // 9. Cache results
    s.cache.Set(ctx, cacheKey, fusedResults, 300) // 5min TTL

    return fusedResults, nil
}

// getSQLBaseline gets last 20 messages as baseline
func (s *HybridSearchService) getSQLBaseline(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
    // Query: SELECT * FROM memory_embeddings
    // WHERE tenant_id = ? AND project_id = ?
    // [AND contact_id = ?]
    // ORDER BY created_at DESC LIMIT 20

    // This ensures we ALWAYS have recent context
    return nil, nil // TODO: Implement
}

func (s *HybridSearchService) buildCacheKey(req *SearchRequest) string {
    return fmt.Sprintf("search:%s:%s:%s", req.TenantID, req.ProjectID, req.Query)
}
```

#### **rrf_fusion.go** (Reciprocal Rank Fusion)

```go
package memory

import (
    "sort"
)

// RRFFusion implements Reciprocal Rank Fusion algorithm
// Paper: https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf
type RRFFusion struct {
    k int // Constant k (typically 60)
}

// NewRRFFusion creates a new RRF fusion instance
func NewRRFFusion() *RRFFusion {
    return &RRFFusion{k: 60}
}

// Fuse combines multiple ranked lists using RRF
func (r *RRFFusion) Fuse(
    sqlResults []*SearchResult,
    vectorResults []*SearchResult,
    keywordResults []*SearchResult,
    graphResults []*SearchResult,
    strategy *memory.RetrievalStrategy,
) []*SearchResult {
    // RRF Score formula: sum_over_all_lists( weight * (1 / (k + rank)) )
    // where rank starts at 1 for the top result

    scoreMap := make(map[string]*SearchResult)

    // SQL baseline (always included, small weight)
    r.addScores(scoreMap, sqlResults, strategy.SQLBaselineWeight, "sql")

    // Vector search
    r.addScores(scoreMap, vectorResults, strategy.VectorWeight, "vector")

    // Keyword search
    r.addScores(scoreMap, keywordResults, strategy.KeywordWeight, "keyword")

    // Graph search
    r.addScores(scoreMap, graphResults, strategy.GraphWeight, "graph")

    // Convert map to slice
    results := make([]*SearchResult, 0, len(scoreMap))
    for _, result := range scoreMap {
        results = append(results, result)
    }

    // Sort by final score (descending)
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    return results
}

// addScores adds RRF scores from a ranked list
func (r *RRFFusion) addScores(scoreMap map[string]*SearchResult, results []*SearchResult, weight float32, source string) {
    for rank, result := range results {
        key := result.EmbeddingID.String()

        // RRF score contribution from this list
        rrfScore := weight * (1.0 / float32(r.k+rank+1))

        if existing, found := scoreMap[key]; found {
            // Already seen - accumulate score
            existing.Score += rrfScore
        } else {
            // First time seeing this result
            newResult := *result // Copy
            newResult.Score = rrfScore
            newResult.Source = source
            scoreMap[key] = &newResult
        }
    }
}
```

---

### 1.4. Application Layer (5 days)

#### **Files to Create**:

```
internal/application/memory/
├── generate_embedding_usecase.go    ❌ NEW - Generate embedding for content
├── search_memory_usecase.go         ❌ NEW - Hybrid search use case
├── extract_facts_usecase.go         ❌ NEW - Extract facts via LLM
├── get_contact_context_usecase.go   ❌ NEW - Get full context for contact
└── ports.go                         ❌ NEW - Port interfaces
```

---

### 1.5. Testing (3 days)

#### **Tests to Create**:
- Unit tests: Domain (memory_test.go)
- Integration tests: Hybrid search with test data
- Benchmark tests: Vector search performance
- E2E tests: Full search flow

**Target Coverage**: 70%+

---

## 🔴 PHASE 2: MCP SERVER (Go) - 3 WEEKS

### 2.1. Project Setup (2 days)

#### **Directory Structure**:

```
cmd/mcp-server/
├── main.go                   ❌ NEW - MCP server entrypoint
└── config.go                 ❌ NEW - Server configuration

infrastructure/mcp/
├── server.go                 ❌ NEW - HTTP server (SSE streaming)
├── auth.go                   ❌ NEW - JWT authentication
├── tool_registry.go          ❌ NEW - Tool registry
├── tool_executor.go          ❌ NEW - Tool execution engine
├── streaming.go              ❌ NEW - SSE streaming handler
└── middleware.go             ❌ NEW - HTTP middleware

infrastructure/mcp/tools/
├── bi_tools.go               ❌ NEW - 7 BI tools
├── agent_analysis_tools.go   ❌ NEW - 5 agent analysis tools
├── crm_operations_tools.go   ❌ NEW - 8 CRM operation tools
├── memory_tools.go           ❌ NEW - 5 memory tools
├── document_tools.go         ❌ NEW - 5 document tools
└── tool_schemas.go           ❌ NEW - JSON schemas for all tools
```

#### **main.go**

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/ventros/crm/infrastructure/mcp"
)

func main() {
    // Load configuration
    cfg, err := LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize MCP server
    mcpServer, err := mcp.NewServer(cfg)
    if err != nil {
        log.Fatalf("Failed to create MCP server: %v", err)
    }

    // HTTP server
    httpServer := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Port),
        Handler:      mcpServer.Handler(),
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    go func() {
        log.Printf("MCP Server listening on port %d", cfg.Port)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("HTTP server error: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down MCP server...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := httpServer.Shutdown(ctx); err != nil {
        log.Fatalf("MCP server forced shutdown: %v", err)
    }

    log.Println("MCP server stopped")
}
```

---

### 2.2. Tool Implementation (10 days)

#### **BI Tools** (7 tools - 2 days)

```go
// infrastructure/mcp/tools/bi_tools.go

package tools

import (
    "context"
    "encoding/json"
)

// GetLeadsCountTool returns count of leads in pipeline
type GetLeadsCountTool struct {
    pipelineRepo pipeline.Repository
}

func (t *GetLeadsCountTool) Name() string { return "get_leads_count" }
func (t *GetLeadsCountTool) Description() string {
    return "Get total count of leads in pipeline, optionally filtered by status"
}

func (t *GetLeadsCountTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "project_id": map[string]string{
                "type":        "string",
                "description": "Project UUID",
            },
            "status_id": map[string]interface{}{
                "type":        "string",
                "description": "Optional: Filter by pipeline status ID",
            },
        },
        "required": []string{"project_id"},
    }
}

func (t *GetLeadsCountTool) Execute(ctx context.Context, input json.RawMessage) (interface{}, error) {
    // Parse input
    var params struct {
        ProjectID string `json:"project_id"`
        StatusID  string `json:"status_id,omitempty"`
    }
    if err := json.Unmarshal(input, &params); err != nil {
        return nil, err
    }

    // Execute query
    count, err := t.pipelineRepo.CountLeadsByStatus(ctx, params.ProjectID, params.StatusID)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "count": count,
        "project_id": params.ProjectID,
        "status_id": params.StatusID,
    }, nil
}
```

**Complete List of BI Tools**:
1. `get_leads_count` - Count leads in pipeline
2. `get_agent_conversion_stats` - Agent conversion metrics
3. `get_top_performing_agent` - Top agent by conversions
4. `get_pipeline_health_score` - Pipeline health (0-100)
5. `get_revenue_forecast` - Revenue forecast
6. `get_session_duration_avg` - Average session duration
7. `get_response_time_percentiles` - P50/P90/P99 response times

---

#### **Memory Tools** (5 tools - 3 days)

```go
// infrastructure/mcp/tools/memory_tools.go

// SearchMemoryTool performs hybrid search in memory
type SearchMemoryTool struct {
    searchService *memory.HybridSearchService
}

func (t *SearchMemoryTool) Name() string { return "search_memory" }
func (t *SearchMemoryTool) Description() string {
    return "Search contact memory using hybrid search (vector + keyword + graph)"
}

func (t *SearchMemoryTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "project_id": map[string]string{
                "type":        "string",
                "description": "Project UUID",
            },
            "query": map[string]string{
                "type":        "string",
                "description": "Search query",
            },
            "contact_id": map[string]interface{}{
                "type":        "string",
                "description": "Optional: Filter by contact",
            },
            "max_results": map[string]interface{}{
                "type":        "integer",
                "description": "Max results (default: 20)",
                "default":     20,
            },
        },
        "required": []string{"project_id", "query"},
    }
}

func (t *SearchMemoryTool) Execute(ctx context.Context, input json.RawMessage) (interface{}, error) {
    // Parse input
    var params struct {
        ProjectID  string `json:"project_id"`
        Query      string `json:"query"`
        ContactID  string `json:"contact_id,omitempty"`
        MaxResults int    `json:"max_results"`
    }
    if err := json.Unmarshal(input, &params); err != nil {
        return nil, err
    }

    if params.MaxResults == 0 {
        params.MaxResults = 20
    }

    // Execute search
    results, err := t.searchService.Search(ctx, &memory.SearchRequest{
        ProjectID:  uuid.MustParse(params.ProjectID),
        Query:      params.Query,
        ContactID:  parseUUIDPtr(params.ContactID),
        MaxResults: params.MaxResults,
    })
    if err != nil {
        return nil, err
    }

    return results, nil
}
```

**Complete List of Memory Tools**:
1. `search_memory` - Hybrid search
2. `get_contact_context` - Full contact context (facts + history)
3. `get_memory_facts` - Extracted facts for contact
4. `store_memory_note` - Store custom note with embedding
5. `delete_memory_embedding` - Delete embedding

---

### 2.3. Authentication & Security (3 days)

#### **JWT Authentication**

```go
// infrastructure/mcp/auth.go

type AuthMiddleware struct {
    jwtSecret []byte
}

func (m *AuthMiddleware) ValidateJWT(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate JWT
        claims, err := m.validateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add claims to context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

### 2.4. SSE Streaming (2 days)

#### **Server-Sent Events for Real-Time Responses**

```go
// infrastructure/mcp/streaming.go

type StreamingHandler struct {
    executor *ToolExecutor
}

func (h *StreamingHandler) HandleToolExecution(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    // Parse request
    var req struct {
        Tool  string          `json:"tool"`
        Input json.RawMessage `json:"input"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Execute tool with streaming
    resultChan := make(chan interface{})
    errorChan := make(chan error)

    go func() {
        result, err := h.executor.Execute(r.Context(), req.Tool, req.Input)
        if err != nil {
            errorChan <- err
            return
        }
        resultChan <- result
    }()

    // Stream results
    select {
    case result := <-resultChan:
        data, _ := json.Marshal(result)
        fmt.Fprintf(w, "data: %s\n\n", data)
        flusher.Flush()
    case err := <-errorChan:
        fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
        flusher.Flush()
    case <-r.Context().Done():
        return
    }
}
```

---

### 2.5. Documentation & Testing (4 days)

- OpenAPI spec for MCP server
- Unit tests for all 30 tools
- Integration tests with Claude Desktop
- Load testing (100 concurrent requests)

---

## 🟠 PHASE 3: gRPC API (Go ↔ Python) - 2 WEEKS

### 3.1. Protocol Buffers (3 days)

#### **Directory Structure**:

```
api/proto/
├── memory_service.proto      ❌ NEW - Memory service definition
├── agent_service.proto       ❌ NEW - Agent service definition (future)
└── common.proto              ❌ NEW - Common types

api/proto/gen/
├── go/                       ❌ GENERATED - Go code
└── python/                   ❌ GENERATED - Python code
```

#### **memory_service.proto**

```protobuf
syntax = "proto3";

package ventros.memory.v1;

option go_package = "github.com/ventros/crm/api/proto/gen/go/memory/v1;memoryv1";

// MemoryService provides access to hybrid search and memory management
service MemoryService {
  // SearchMemory performs hybrid search
  rpc SearchMemory(SearchMemoryRequest) returns (SearchMemoryResponse);

  // StoreEmbedding stores a new embedding
  rpc StoreEmbedding(StoreEmbeddingRequest) returns (StoreEmbeddingResponse);

  // GetContactContext retrieves full context for a contact
  rpc GetContactContext(GetContactContextRequest) returns (GetContactContextResponse);

  // ExtractFacts extracts facts from text via LLM
  rpc ExtractFacts(ExtractFactsRequest) returns (ExtractFactsResponse);

  // GetMemoryFacts retrieves extracted facts for a contact
  rpc GetMemoryFacts(GetMemoryFactsRequest) returns (GetMemoryFactsResponse);
}

// SearchMemoryRequest encapsulates search parameters
message SearchMemoryRequest {
  string tenant_id = 1;
  string project_id = 2;
  string query = 3;
  optional string contact_id = 4;
  optional string session_id = 5;
  string strategy_name = 6; // "sales_prospecting", "retention_churn", etc.
  int32 max_results = 7;
}

// SearchMemoryResponse contains search results
message SearchMemoryResponse {
  repeated SearchResult results = 1;
  SearchMetadata metadata = 2;
}

message SearchResult {
  string embedding_id = 1;
  string content_text = 2;
  string content_type = 3;
  float score = 4;
  string source = 5; // "vector", "keyword", "graph", "sql"
  map<string, string> metadata = 6;
}

message SearchMetadata {
  int32 total_results = 1;
  float search_duration_ms = 2;
  bool from_cache = 3;
}

// StoreEmbeddingRequest stores a new embedding
message StoreEmbeddingRequest {
  string tenant_id = 1;
  string project_id = 2;
  string content_type = 3;
  string content_text = 4;
  repeated float embedding = 5; // 768 dimensions
  optional string contact_id = 6;
  optional string session_id = 7;
  optional string message_id = 8;
  map<string, string> metadata = 9;
}

message StoreEmbeddingResponse {
  string embedding_id = 1;
  string content_hash = 2;
}

// ExtractFactsRequest extracts facts from text
message ExtractFactsRequest {
  string tenant_id = 1;
  string project_id = 2;
  string contact_id = 3;
  string text = 4;
  optional string message_id = 5;
}

message ExtractFactsResponse {
  repeated MemoryFact facts = 1;
}

message MemoryFact {
  string fact_type = 1;
  string fact_text = 2;
  string fact_category = 3;
  float confidence = 4;
}

// GetContactContextRequest retrieves full context
message GetContactContextRequest {
  string tenant_id = 1;
  string project_id = 2;
  string contact_id = 3;
  bool include_facts = 4;
  bool include_recent_messages = 5;
  int32 max_messages = 6;
}

message GetContactContextResponse {
  repeated SearchResult recent_context = 1;
  repeated MemoryFact facts = 2;
  ContactSummary summary = 3;
}

message ContactSummary {
  string contact_id = 1;
  string full_name = 2;
  string email = 3;
  string phone = 4;
  int32 total_messages = 5;
  int32 total_sessions = 6;
  string pipeline_status = 7;
}

// GetMemoryFactsRequest retrieves facts
message GetMemoryFactsRequest {
  string tenant_id = 1;
  string project_id = 2;
  string contact_id = 3;
  optional string fact_type = 4;
  bool validated_only = 5;
}

message GetMemoryFactsResponse {
  repeated MemoryFact facts = 1;
}
```

---

### 3.2. Go Server Implementation (4 days)

#### **Files to Create**:

```
infrastructure/grpc/
├── server.go                 ❌ NEW - gRPC server
├── memory_service_impl.go    ❌ NEW - MemoryService implementation
├── interceptors.go           ❌ NEW - Auth, logging interceptors
└── health.go                 ❌ NEW - Health check service

cmd/grpc-server/
└── main.go                   ❌ NEW - gRPC server entrypoint
```

#### **memory_service_impl.go**

```go
package grpc

import (
    "context"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/ventros/crm/api/proto/gen/go/memory/v1"
    "github.com/ventros/crm/internal/application/memory"
)

type MemoryServiceServer struct {
    pb.UnimplementedMemoryServiceServer
    searchUseCase *memory.SearchMemoryUseCase
    storeUseCase  *memory.StoreEmbeddingUseCase
    contextUseCase *memory.GetContactContextUseCase
    factsUseCase   *memory.ExtractFactsUseCase
}

func (s *MemoryServiceServer) SearchMemory(ctx context.Context, req *pb.SearchMemoryRequest) (*pb.SearchMemoryResponse, error) {
    // Validate request
    if req.TenantId == "" || req.ProjectId == "" || req.Query == "" {
        return nil, status.Error(codes.InvalidArgument, "missing required fields")
    }

    // Execute use case
    results, err := s.searchUseCase.Execute(ctx, &memory.SearchMemoryCommand{
        TenantID:   req.TenantId,
        ProjectID:  req.ProjectId,
        Query:      req.Query,
        ContactID:  req.ContactId,
        SessionID:  req.SessionId,
        Strategy:   req.StrategyName,
        MaxResults: int(req.MaxResults),
    })
    if err != nil {
        return nil, status.Errorf(codes.Internal, "search failed: %v", err)
    }

    // Convert to protobuf
    pbResults := make([]*pb.SearchResult, len(results))
    for i, result := range results {
        pbResults[i] = &pb.SearchResult{
            EmbeddingId: result.EmbeddingID,
            ContentText: result.ContentText,
            ContentType: result.ContentType,
            Score:       result.Score,
            Source:      result.Source,
            Metadata:    result.Metadata,
        }
    }

    return &pb.SearchMemoryResponse{
        Results: pbResults,
        Metadata: &pb.SearchMetadata{
            TotalResults:      int32(len(results)),
            SearchDurationMs:  123.45, // TODO: measure
            FromCache:         false,
        },
    }, nil
}

// ... implement other methods
```

---

### 3.3. Python Client Implementation (4 days)

#### **Directory Structure** (New Python Project):

```
python-adk/
├── pyproject.toml            ❌ NEW - Poetry dependencies
├── poetry.lock               ❌ GENERATED
├── README.md                 ❌ NEW
├── .python-version           ❌ NEW - Python 3.11+
│
├── ventros_adk/
│   ├── __init__.py
│   │
│   ├── grpc/
│   │   ├── __init__.py
│   │   ├── memory_client.py  ❌ NEW - gRPC client wrapper
│   │   └── proto/            ❌ GENERATED - Python protobuf code
│   │
│   ├── agents/               (Phase 4)
│   ├── semantic_router/      (Phase 4)
│   └── tools/                (Phase 4)
│
└── tests/
    ├── __init__.py
    ├── test_memory_client.py ❌ NEW
    └── integration/
        └── test_grpc_integration.py ❌ NEW
```

#### **pyproject.toml**

```toml
[tool.poetry]
name = "ventros-adk"
version = "0.1.0"
description = "Ventros CRM AI Agent Development Kit"
authors = ["Ventros Team <dev@ventros.cloud>"]
readme = "README.md"
python = "^3.11"

[tool.poetry.dependencies]
# Core
python = "^3.11"

# Google Cloud ADK 0.5+
google-genai = "^0.5.0"

# gRPC
grpcio = "^1.60.0"
grpcio-tools = "^1.60.0"
protobuf = "^4.25.0"

# Semantic Router
semantic-router = "^0.0.50"

# LangChain (optional)
langchain = "^0.1.0"
langchain-google-genai = "^0.1.0"

# Observability
opentelemetry-api = "^1.22.0"
opentelemetry-sdk = "^1.22.0"
arize-phoenix = "^4.0.0"

# Messaging
pika = "^1.3.2"  # RabbitMQ client

# Temporal
temporalio = "^1.5.0"

# Utilities
pydantic = "^2.5.0"
pydantic-settings = "^2.1.0"
structlog = "^24.1.0"
rich = "^13.7.0"

[tool.poetry.group.dev.dependencies]
# Testing
pytest = "^7.4.0"
pytest-asyncio = "^0.21.0"
pytest-cov = "^4.1.0"
pytest-mock = "^3.12.0"

# Linting & Formatting
ruff = "^0.1.0"
black = "^23.12.0"
mypy = "^1.7.0"
isort = "^5.13.0"

# Type stubs
types-protobuf = "^4.24.0"

[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"

[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_classes = ["Test*"]
python_functions = ["test_*"]
addopts = "-v --cov=ventros_adk --cov-report=term-missing"

[tool.ruff]
line-length = 100
target-version = "py311"

[tool.black]
line-length = 100
target-version = ["py311"]

[tool.mypy]
python_version = "3.11"
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
```

#### **memory_client.py**

```python
"""gRPC client for Memory Service"""

import grpc
from typing import List, Optional
from dataclasses import dataclass

from ventros_adk.grpc.proto import memory_service_pb2 as pb
from ventros_adk.grpc.proto import memory_service_pb2_grpc as pb_grpc


@dataclass
class SearchResult:
    """Search result from Memory Service"""

    embedding_id: str
    content_text: str
    content_type: str
    score: float
    source: str
    metadata: dict[str, str]


@dataclass
class SearchMetadata:
    """Metadata about search execution"""

    total_results: int
    search_duration_ms: float
    from_cache: bool


class MemoryClient:
    """gRPC client for Memory Service (Go backend)"""

    def __init__(self, host: str = "localhost", port: int = 50051):
        """
        Initialize Memory Service client

        Args:
            host: gRPC server host
            port: gRPC server port
        """
        self.channel = grpc.insecure_channel(f"{host}:{port}")
        self.stub = pb_grpc.MemoryServiceStub(self.channel)

    def search_memory(
        self,
        tenant_id: str,
        project_id: str,
        query: str,
        contact_id: Optional[str] = None,
        session_id: Optional[str] = None,
        strategy_name: str = "general_balanced",
        max_results: int = 20,
    ) -> tuple[List[SearchResult], SearchMetadata]:
        """
        Search memory using hybrid search

        Args:
            tenant_id: Tenant UUID
            project_id: Project UUID
            query: Search query
            contact_id: Optional contact UUID filter
            session_id: Optional session UUID filter
            strategy_name: Retrieval strategy (default: general_balanced)
            max_results: Max results to return

        Returns:
            Tuple of (results, metadata)

        Raises:
            grpc.RpcError: If gRPC call fails
        """
        request = pb.SearchMemoryRequest(
            tenant_id=tenant_id,
            project_id=project_id,
            query=query,
            strategy_name=strategy_name,
            max_results=max_results,
        )

        if contact_id:
            request.contact_id = contact_id
        if session_id:
            request.session_id = session_id

        try:
            response = self.stub.SearchMemory(request)
        except grpc.RpcError as e:
            # Log and re-raise
            print(f"gRPC error: {e.code()} - {e.details()}")
            raise

        # Convert protobuf to dataclasses
        results = [
            SearchResult(
                embedding_id=r.embedding_id,
                content_text=r.content_text,
                content_type=r.content_type,
                score=r.score,
                source=r.source,
                metadata=dict(r.metadata),
            )
            for r in response.results
        ]

        metadata = SearchMetadata(
            total_results=response.metadata.total_results,
            search_duration_ms=response.metadata.search_duration_ms,
            from_cache=response.metadata.from_cache,
        )

        return results, metadata

    def close(self):
        """Close gRPC channel"""
        self.channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


# Example usage
if __name__ == "__main__":
    with MemoryClient() as client:
        results, metadata = client.search_memory(
            tenant_id="tenant-123",
            project_id="project-456",
            query="What is the customer's budget?",
            strategy_name="sales_prospecting",
        )

        print(f"Found {len(results)} results in {metadata.search_duration_ms:.2f}ms")
        for result in results:
            print(f"- [{result.score:.3f}] {result.content_text[:100]}")
```

---

### 3.4. Code Generation (1 day)

#### **Makefile Targets**:

```makefile
# Makefile additions

.PHONY: proto-gen proto-go proto-python

proto-gen: proto-go proto-python

proto-go:
	@echo "Generating Go protobuf code..."
	protoc --go_out=api/proto/gen/go --go_opt=paths=source_relative \
	       --go-grpc_out=api/proto/gen/go --go-grpc_opt=paths=source_relative \
	       api/proto/*.proto

proto-python:
	@echo "Generating Python protobuf code..."
	python -m grpc_tools.protoc \
	    -Iapi/proto \
	    --python_out=python-adk/ventros_adk/grpc/proto \
	    --grpc_python_out=python-adk/ventros_adk/grpc/proto \
	    api/proto/*.proto

grpc-server:
	@echo "Starting gRPC server..."
	go run cmd/grpc-server/main.go
```

---

## 🔴 PHASE 4: PYTHON ADK - 6 WEEKS

### 4.1. Project Setup (3 days)

#### **Initialize Poetry Project**:

```bash
cd python-adk
poetry init
poetry add google-genai semantic-router grpcio pika temporalio pydantic structlog
poetry add --group dev pytest pytest-asyncio pytest-cov ruff black mypy
poetry install
```

---

### 4.2. Multi-Agent System (15 days)

#### **Directory Structure**:

```
python-adk/ventros_adk/
├── __init__.py
│
├── agents/
│   ├── __init__.py
│   ├── base_agent.py             ❌ NEW - Base agent class
│   ├── coordinator_agent.py      ❌ NEW - Coordinator (orchestrator)
│   ├── sales_prospecting_agent.py ❌ NEW - Sales specialist
│   ├── retention_churn_agent.py  ❌ NEW - Retention specialist
│   ├── support_technical_agent.py ❌ NEW - Technical support
│   ├── support_billing_agent.py  ❌ NEW - Billing support
│   └── balanced_agent.py         ❌ NEW - Fallback agent
│
├── semantic_router/
│   ├── __init__.py
│   ├── router.py                 ❌ NEW - Semantic router
│   ├── routes.py                 ❌ NEW - Route definitions
│   └── embeddings.py             ❌ NEW - Embedding provider
│
├── tools/
│   ├── __init__.py
│   ├── tool_registry.py          ❌ NEW - Tool registry
│   ├── crm_tools.py              ❌ NEW - CRM operation tools
│   ├── memory_tools.py           ❌ NEW - Memory access tools
│   └── web_tools.py              ❌ NEW - Web search tools (optional)
│
├── memory/
│   ├── __init__.py
│   ├── memory_service.py         ❌ NEW - Memory facade
│   └── context_builder.py        ❌ NEW - Context builder
│
├── messaging/
│   ├── __init__.py
│   ├── rabbitmq_consumer.py      ❌ NEW - RabbitMQ consumer
│   ├── rabbitmq_publisher.py     ❌ NEW - RabbitMQ publisher
│   └── event_handlers.py         ❌ NEW - Event handler registry
│
├── workflows/
│   ├── __init__.py
│   ├── temporal_client.py        ❌ NEW - Temporal client
│   ├── agent_workflow.py         ❌ NEW - Agent workflow
│   └── activities.py             ❌ NEW - Temporal activities
│
├── observability/
│   ├── __init__.py
│   ├── phoenix_tracer.py         ❌ NEW - Phoenix integration
│   ├── metrics.py                ❌ NEW - Metrics collector
│   └── logger.py                 ❌ NEW - Structured logger
│
└── config/
    ├── __init__.py
    └── settings.py               ❌ NEW - Pydantic settings
```

---

#### **coordinator_agent.py**

```python
"""Coordinator Agent - Orchestrates specialist agents"""

from typing import Optional
from google import genai
from google.genai import types

from ventros_adk.semantic_router import SemanticRouter
from ventros_adk.agents.base_agent import BaseAgent
from ventros_adk.memory import MemoryService


class CoordinatorAgent(BaseAgent):
    """
    Coordinator Agent orchestrates specialist agents using semantic routing

    Flow:
    1. Receive message from RabbitMQ
    2. Build context from Memory Service
    3. Semantic routing to determine specialist
    4. Delegate to specialist agent
    5. Return response
    """

    def __init__(
        self,
        *,
        model_name: str = "gemini-2.0-flash",
        memory_service: MemoryService,
        semantic_router: SemanticRouter,
    ):
        super().__init__(model_name=model_name)
        self.memory_service = memory_service
        self.semantic_router = semantic_router

        # Specialist agents (lazy initialization)
        self._sales_agent: Optional[BaseAgent] = None
        self._retention_agent: Optional[BaseAgent] = None
        self._support_technical_agent: Optional[BaseAgent] = None
        self._support_billing_agent: Optional[BaseAgent] = None
        self._balanced_agent: Optional[BaseAgent] = None

    @property
    def sales_agent(self) -> BaseAgent:
        if self._sales_agent is None:
            from ventros_adk.agents.sales_prospecting_agent import SalesProspectingAgent
            self._sales_agent = SalesProspectingAgent(
                memory_service=self.memory_service
            )
        return self._sales_agent

    # ... similar for other agents

    async def process_message(
        self,
        tenant_id: str,
        project_id: str,
        contact_id: str,
        session_id: str,
        message_text: str,
    ) -> str:
        """
        Process inbound message and generate response

        Args:
            tenant_id: Tenant UUID
            project_id: Project UUID
            contact_id: Contact UUID
            session_id: Session UUID
            message_text: Message content

        Returns:
            Generated response text
        """
        # 1. Build context from Memory Service
        context = await self.memory_service.get_contact_context(
            tenant_id=tenant_id,
            project_id=project_id,
            contact_id=contact_id,
        )

        # 2. Semantic routing to determine intent
        route = await self.semantic_router.route(message_text)

        # 3. Select specialist agent
        specialist = self._select_specialist(route.name)

        # 4. Generate response via specialist
        response = await specialist.generate_response(
            message=message_text,
            context=context,
            session_id=session_id,
        )

        return response

    def _select_specialist(self, route_name: str) -> BaseAgent:
        """Select specialist based on route"""
        if route_name == "sales_prospecting":
            return self.sales_agent
        elif route_name == "retention_churn":
            return self.retention_agent
        elif route_name == "support_technical":
            return self.support_technical_agent
        elif route_name == "support_billing":
            return self.support_billing_agent
        else:
            return self.balanced_agent
```

---

#### **sales_prospecting_agent.py**

```python
"""Sales Prospecting Agent - Specialist for sales conversations"""

from google import genai
from google.genai import types

from ventros_adk.agents.base_agent import BaseAgent
from ventros_adk.memory import MemoryService, ContactContext


SALES_SYSTEM_PROMPT = """
You are a Sales Prospecting Agent for Ventros CRM.

Your role:
- Qualify leads based on budget, authority, need, timeline (BANT)
- Identify pain points and decision makers
- Move contacts through the sales pipeline
- Schedule demos and meetings
- Handle objections professionally

Retrieval Strategy: sales_prospecting
- Graph weight: HIGH (campaign attribution critical)
- Keyword weight: MEDIUM (specific terms like "budget", "demo")
- Vector weight: LOW (semantic similarity less critical)

Tools available:
- qualify_lead(contact_id, qualification_data)
- update_pipeline_stage(contact_id, stage_id)
- schedule_demo(contact_id, datetime)
- create_note(contact_id, note_text)

Always:
- Ask qualifying questions
- Listen for buying signals
- Create notes after important revelations
- Update pipeline stage when appropriate

Context:
{context}
"""


class SalesProspectingAgent(BaseAgent):
    """Sales Prospecting Specialist Agent"""

    def __init__(self, *, memory_service: MemoryService):
        super().__init__(
            model_name="gemini-2.0-flash",
            system_prompt=SALES_SYSTEM_PROMPT,
        )
        self.memory_service = memory_service

    async def generate_response(
        self,
        message: str,
        context: ContactContext,
        session_id: str,
    ) -> str:
        """
        Generate sales-focused response

        Args:
            message: Inbound message
            context: Contact context from Memory Service
            session_id: Session UUID

        Returns:
            Generated response
        """
        # Format context
        formatted_context = self._format_context(context)

        # Build prompt
        full_prompt = self.system_prompt.format(context=formatted_context)

        # Generate response using Google ADK
        response = await self.client.aio.models.generate_content(
            model=self.model_name,
            contents=types.Content(
                parts=[
                    types.Part(text=full_prompt),
                    types.Part(text=f"Contact message: {message}"),
                ],
            ),
            config=types.GenerateContentConfig(
                temperature=0.7,
                max_output_tokens=512,
            ),
        )

        return response.text

    def _format_context(self, context: ContactContext) -> str:
        """Format context for prompt"""
        parts = []

        # Contact summary
        parts.append(f"Contact: {context.summary.full_name}")
        parts.append(f"Pipeline: {context.summary.pipeline_status}")

        # Recent messages
        if context.recent_messages:
            parts.append("\nRecent conversation:")
            for msg in context.recent_messages[-5:]:
                parts.append(f"- {msg.content_text[:100]}")

        # Extracted facts
        if context.facts:
            parts.append("\nKnown facts:")
            for fact in context.facts:
                parts.append(f"- [{fact.fact_type}] {fact.fact_text} (confidence: {fact.confidence:.2f})")

        return "\n".join(parts)
```

---

### 4.3. Semantic Router (4 days)

#### **router.py**

```python
"""Semantic Router for intent classification"""

from typing import List
from dataclasses import dataclass

from semantic_router import SemanticRouter as SR
from semantic_router.encoders import GoogleEncoder


@dataclass
class Route:
    """Route definition"""
    name: str
    score: float


class SemanticRouter:
    """
    Semantic Router classifies message intent to route to specialist agents

    Uses semantic-router library with Google embeddings
    """

    def __init__(self):
        # Initialize Google encoder (text-embedding-005)
        self.encoder = GoogleEncoder(model_name="text-embedding-005")

        # Define routes with example utterances
        self.router = SR(
            encoder=self.encoder,
            routes=[
                {
                    "name": "sales_prospecting",
                    "utterances": [
                        "I'm interested in your product",
                        "Can you tell me about pricing?",
                        "I'd like to schedule a demo",
                        "What's your budget range?",
                        "When can we start?",
                        "I need to discuss this with my team",
                    ],
                },
                {
                    "name": "retention_churn",
                    "utterances": [
                        "I'm thinking of canceling",
                        "This isn't working for us",
                        "We're not seeing the value",
                        "The competitor offers more",
                        "I want to downgrade my plan",
                        "Why should we stay?",
                    ],
                },
                {
                    "name": "support_technical",
                    "utterances": [
                        "The system is not working",
                        "I'm getting an error",
                        "How do I configure X?",
                        "The integration is broken",
                        "I can't connect to the API",
                        "There's a bug in the dashboard",
                    ],
                },
                {
                    "name": "support_billing",
                    "utterances": [
                        "I have a question about my invoice",
                        "Why was I charged twice?",
                        "Can I get a refund?",
                        "How do I update my credit card?",
                        "What's included in my plan?",
                        "I need a copy of my receipt",
                    ],
                },
            ],
        )

    async def route(self, message: str) -> Route:
        """
        Route message to appropriate specialist

        Args:
            message: Inbound message text

        Returns:
            Route with name and confidence score
        """
        result = self.router(message)

        # If no route matched, use balanced agent
        if result is None:
            return Route(name="balanced", score=0.0)

        return Route(name=result.name, score=result.score)
```

---

### 4.4. RabbitMQ Integration (5 days)

#### **rabbitmq_consumer.py**

```python
"""RabbitMQ consumer for inbound messages"""

import asyncio
import json
from typing import Callable, Awaitable

import pika
from pika.adapters.asyncio_connection import AsyncioConnection
from pika.channel import Channel

from ventros_adk.agents import CoordinatorAgent
from ventros_adk.observability import logger


MessageHandler = Callable[[dict], Awaitable[None]]


class RabbitMQConsumer:
    """
    RabbitMQ consumer for message.inbound events

    Consumes from: message.inbound queue
    Publishes to: message.outbound exchange
    """

    def __init__(
        self,
        *,
        host: str = "localhost",
        port: int = 5672,
        username: str = "guest",
        password: str = "guest",
        coordinator_agent: CoordinatorAgent,
    ):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.coordinator = coordinator_agent

        self.connection: Optional[AsyncioConnection] = None
        self.channel: Optional[Channel] = None

    async def connect(self):
        """Establish connection to RabbitMQ"""
        credentials = pika.PlainCredentials(self.username, self.password)
        parameters = pika.ConnectionParameters(
            host=self.host,
            port=self.port,
            credentials=credentials,
        )

        self.connection = await AsyncioConnection.create(parameters)
        self.channel = await self.connection.channel()

        # Declare queue
        await self.channel.queue_declare(
            queue="message.inbound",
            durable=True,
        )

        # Set QoS (prefetch 1 message at a time)
        await self.channel.basic_qos(prefetch_count=1)

        logger.info("Connected to RabbitMQ", host=self.host, queue="message.inbound")

    async def start_consuming(self):
        """Start consuming messages"""
        await self.channel.basic_consume(
            queue="message.inbound",
            on_message_callback=self._on_message,
            auto_ack=False,
        )

        logger.info("Started consuming messages")

        # Keep consuming
        await asyncio.Future()

    async def _on_message(self, channel: Channel, method, properties, body: bytes):
        """Handle incoming message"""
        try:
            # Parse message
            message = json.loads(body.decode())
            logger.info("Received message", message_id=message.get("message_id"))

            # Extract fields
            tenant_id = message["tenant_id"]
            project_id = message["project_id"]
            contact_id = message["contact_id"]
            session_id = message["session_id"]
            message_text = message["content"]

            # Process via Coordinator Agent
            response = await self.coordinator.process_message(
                tenant_id=tenant_id,
                project_id=project_id,
                contact_id=contact_id,
                session_id=session_id,
                message_text=message_text,
            )

            # Publish response to outbound queue
            await self._publish_response(
                message_id=message["message_id"],
                session_id=session_id,
                contact_id=contact_id,
                response_text=response,
            )

            # ACK message
            channel.basic_ack(delivery_tag=method.delivery_tag)
            logger.info("Message processed successfully", message_id=message["message_id"])

        except Exception as e:
            logger.error("Failed to process message", error=str(e))
            # NACK with requeue
            channel.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

    async def _publish_response(
        self,
        message_id: str,
        session_id: str,
        contact_id: str,
        response_text: str,
    ):
        """Publish response to outbound exchange"""
        payload = {
            "message_id": message_id,
            "session_id": session_id,
            "contact_id": contact_id,
            "content": response_text,
            "direction": "outbound",
            "source": "ai_agent",
        }

        await self.channel.basic_publish(
            exchange="message.outbound",
            routing_key="",
            body=json.dumps(payload).encode(),
            properties=pika.BasicProperties(
                content_type="application/json",
                delivery_mode=2,  # Persistent
            ),
        )
```

---

### 4.5. Phoenix Observability (3 days)

#### **phoenix_tracer.py**

```python
"""Phoenix integration for observability"""

from phoenix.trace import using_project
from phoenix.trace.langchain import LangChainInstrumentor


def setup_phoenix(project_name: str = "ventros-adk"):
    """
    Setup Phoenix tracing

    Args:
        project_name: Phoenix project name
    """
    # Auto-instrument LangChain (if used)
    LangChainInstrumentor().instrument()

    # Set project
    using_project(project_name)
```

---

### 4.6. Testing (5 days)

- Unit tests for each agent
- Integration tests with gRPC (mocked)
- E2E tests with RabbitMQ (testcontainers)
- Load tests (100 concurrent messages)

**Target Coverage**: 70%+

---

## 🟡 PHASE 5: ADVANCED MEMORY - 5 WEEKS

### 5.1. Apache AGE Setup (3 days)

#### **Installation**:

```sql
-- Install Apache AGE extension
CREATE EXTENSION IF NOT EXISTS age;

-- Create graph
SELECT create_graph('ventros_graph');
```

#### **Graph Schema**:

```cypher
// Nodes
CREATE (:Contact {id: 'uuid', name: 'string'})
CREATE (:Session {id: 'uuid', status: 'string'})
CREATE (:Message {id: 'uuid', content: 'string'})
CREATE (:Offer {id: 'uuid', value: float})
CREATE (:Campaign {id: 'uuid', name: 'string'})

// Edges
CREATE (c:Contact)-[:HAS_SESSION]->(s:Session)
CREATE (s:Session)-[:CONTAINS_MESSAGE]->(m:Message)
CREATE (c:Contact)-[:RECEIVED_OFFER]->(o:Offer)
CREATE (m:Message)-[:REPLY_TO]->(m2:Message)
CREATE (c:Contact)-[:CAME_FROM]->(camp:Campaign)
```

---

### 5.2. Graph Queries (4 days)

#### **graph_search.go**

```go
package memory

import (
    "context"
    "fmt"
)

// GraphSearchService performs graph traversal queries using Apache AGE
type GraphSearchService struct {
    db *sql.DB
}

// Search performs graph-based search
func (s *GraphSearchService) Search(ctx context.Context, req *SearchRequest) ([]*SearchResult, error) {
    // Example Cypher query: Find related messages via graph
    query := `
        SELECT * FROM cypher('ventros_graph', $$
            MATCH (c:Contact {id: $1})-[:HAS_SESSION]->(s:Session)
                  -[:CONTAINS_MESSAGE]->(m:Message)
            WHERE s.project_id = $2
            RETURN m.content, m.created_at
            ORDER BY m.created_at DESC
            LIMIT 20
        $$) AS (content agtype, created_at agtype);
    `

    // Execute query
    rows, err := s.db.QueryContext(ctx, query, req.ContactID, req.ProjectID)
    if err != nil {
        return nil, fmt.Errorf("graph query failed: %w", err)
    }
    defer rows.Close()

    // Parse results
    results := []*SearchResult{}
    for rows.Next() {
        var content, createdAt string
        if err := rows.Scan(&content, &createdAt); err != nil {
            return nil, err
        }

        results = append(results, &SearchResult{
            ContentText: content,
            Source:      "graph",
            Score:       1.0, // Graph results get base score
        })
    }

    return results, nil
}
```

---

### 5.3. Facts Extraction (7 days)

#### **extract_facts_usecase.go**

```go
package memory

import (
    "context"
    "fmt"

    "github.com/google/genai"
)

// ExtractFactsUseCase extracts facts from text via LLM
type ExtractFactsUseCase struct {
    llmClient    *genai.Client
    factsRepo    memory.FactsRepository
}

// Execute extracts facts
func (uc *ExtractFactsUseCase) Execute(ctx context.Context, cmd *ExtractFactsCommand) ([]*memory.MemoryFact, error) {
    // Prompt for fact extraction (NER)
    prompt := fmt.Sprintf(`
Extract factual information from the following conversation.

Conversation:
%s

Extract the following fact types:
- budget: Customer budget or pricing mentions
- preference: Product preferences or requirements
- objection: Concerns or objections raised
- pain_point: Problems the customer is facing
- decision_maker: Information about decision makers

For each fact, provide:
1. fact_type (from list above)
2. fact_text (concise statement)
3. confidence (0.0-1.0)

Return JSON array:
[
  {"fact_type": "budget", "fact_text": "Budget of $5,000/month", "confidence": 0.92},
  {"fact_type": "pain_point", "fact_text": "ROI is unclear", "confidence": 0.85}
]
`, cmd.Text)

    // Call LLM
    response, err := uc.llmClient.GenerateContent(ctx, prompt)
    if err != nil {
        return nil, fmt.Errorf("LLM call failed: %w", err)
    }

    // Parse JSON response
    var extractedFacts []struct {
        FactType   string  `json:"fact_type"`
        FactText   string  `json:"fact_text"`
        Confidence float64 `json:"confidence"`
    }
    if err := json.Unmarshal([]byte(response.Text), &extractedFacts); err != nil {
        return nil, fmt.Errorf("failed to parse LLM response: %w", err)
    }

    // Create domain facts
    facts := make([]*memory.MemoryFact, 0, len(extractedFacts))
    for _, ef := range extractedFacts {
        fact, err := memory.NewMemoryFact(
            cmd.TenantID,
            cmd.ProjectID,
            cmd.ContactID,
            memory.FactType(ef.FactType),
            ef.FactText,
            float32(ef.Confidence),
        )
        if err != nil {
            continue // Skip invalid facts
        }

        facts = append(facts, fact)
    }

    // Save facts
    for _, fact := range facts {
        if err := uc.factsRepo.Save(ctx, fact); err != nil {
            return nil, fmt.Errorf("failed to save fact: %w", err)
        }
    }

    return facts, nil
}
```

---

### 5.4. Contradiction Resolution (3 days)

When new facts contradict existing facts, resolve via:
1. Confidence scores (higher wins)
2. Recency (newer wins if confidence similar)
3. Mark old fact as superseded

---

## 🟢 PHASE 6: TEMPLATES & POLISH - 2 WEEKS

### 6.1. Agent Templates Registry (5 days)

#### **system_agents.go**

```go
package agent

// SystemAgents provides pre-configured agent templates
var SystemAgents = []AgentTemplate{
    {
        ID:   "agent-sales-prospecting",
        Name: "Sales Prospecting Bot",
        Category: CategorySalesProspecting,
        Description: "Qualifies leads, identifies pain points, moves contacts through pipeline",
        KnowledgeScope: KnowledgeScope{
            IncludeContactHistory: true,
            IncludePipelineInfo:   true,
            IncludeCampaignData:   true,
            MaxHistoryDays:        90,
        },
        MemoryStrategy: MemoryStrategy{
            StrategyName:  "sales_prospecting",
            UseReranking:  false,
            UseFacts:      true,
            MaxResults:    20,
        },
        Personality: Personality{
            Tone:        "professional, consultative",
            Style:       "asks qualifying questions, listens for buying signals",
            Constraints: []string{"never discuss pricing without approval", "always schedule demos"},
        },
    },
    {
        ID:   "agent-retention-churn",
        Name: "Retention & Churn Prevention Bot",
        Category: CategoryRetentionChurn,
        Description: "Prevents churn, identifies at-risk customers, offers solutions",
        KnowledgeScope: KnowledgeScope{
            IncludeContactHistory: true,
            IncludeBillingInfo:    true,
            IncludeUsageMetrics:   true,
            MaxHistoryDays:        180,
        },
        MemoryStrategy: MemoryStrategy{
            StrategyName:  "retention_churn",
            UseReranking:  true, // Critical for pattern matching
            UseFacts:      true,
            MaxResults:    30,
        },
        Personality: Personality{
            Tone:        "empathetic, solution-focused",
            Style:       "acknowledges concerns, offers alternatives",
            Constraints: []string{"never promise discounts without approval", "escalate high-value churn"},
        },
    },
    // ... 8 more templates
}
```

---

### 6.2. Template Instantiation API (3 days)

```go
// POST /api/v1/agents/from-template
func (h *AgentHandler) CreateFromTemplate(c *gin.Context) {
    var req struct {
        TemplateID string            `json:"template_id"`
        Name       string            `json:"name"`
        ProjectID  uuid.UUID         `json:"project_id"`
        Overrides  map[string]string `json:"overrides"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Find template
    template := agent.FindSystemAgentByID(req.TemplateID)
    if template == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
        return
    }

    // Instantiate agent from template
    agent, err := h.agentService.CreateFromTemplate(c.Request.Context(), template, req.Name, req.ProjectID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, agent)
}
```

---

### 6.3. Tuning & A/B Testing (4 days)

- Tune retrieval strategy weights per category
- A/B test different prompt variations
- Measure: response quality, resolution time, CSAT

---

## 📈 SUCCESS METRICS

### Technical Metrics

**Memory Service**:
- ✅ Hybrid search latency: <200ms (P95)
- ✅ Embedding generation: <500ms per document
- ✅ Cache hit rate: >60%
- ✅ Vector search accuracy: >0.85 (relevance score)

**MCP Server**:
- ✅ Tool execution: <1s (P95)
- ✅ Concurrent requests: 100+ supported
- ✅ Uptime: 99.9%
- ✅ JWT validation: 100% secure

**gRPC API**:
- ✅ Request latency: <100ms (P95)
- ✅ Throughput: 1000+ req/s
- ✅ Error rate: <0.1%

**Python ADK**:
- ✅ Message processing: <2s end-to-end
- ✅ Agent routing accuracy: >85%
- ✅ Memory recall: >90%
- ✅ Test coverage: >70%

### Business Metrics

- ✅ AI response quality: >8.0/10 (human eval)
- ✅ Agent automation rate: >60% of conversations
- ✅ Resolution time: -40% vs manual
- ✅ LLM cost: <$0.01 per interaction

---

## 📚 DEVELOPMENT STANDARDS

### Code Quality

**Go**:
- Follow existing DDD patterns
- 70%+ test coverage
- All exports documented (godoc)
- Pass `golangci-lint` with zero warnings

**Python**:
- Type hints everywhere (mypy strict)
- Docstrings (Google style)
- 70%+ test coverage
- Pass `ruff` and `black` formatting

### Git Workflow

```bash
# Feature branch
git checkout -b feature/memory-service-hybrid-search

# Commit convention
git commit -m "feat(memory): implement hybrid search with RRF fusion"
git commit -m "test(memory): add integration tests for vector search"
git commit -m "docs(memory): document retrieval strategies"

# Push and create PR
git push origin feature/memory-service-hybrid-search
```

### Documentation

Every phase must include:
- Architecture Decision Records (ADRs)
- API documentation (OpenAPI/gRPC)
- README with setup instructions
- Code examples
- Migration guides

---

## 🚀 GETTING STARTED

### Prerequisites

**Go Backend**:
- Go 1.23+
- PostgreSQL 16+ (with pgvector + Apache AGE)
- Redis 7+
- RabbitMQ 3.12+
- Temporal Server

**Python ADK**:
- Python 3.11+
- Poetry 1.7+
- protoc 25+
- gRPC tools

### Setup Commands

```bash
# 1. Install PostgreSQL extensions
psql -U postgres -d ventros -c "CREATE EXTENSION IF NOT EXISTS vector;"
psql -U postgres -d ventros -c "CREATE EXTENSION IF NOT EXISTS age;"

# 2. Run migrations (Go)
cd /path/to/ventros-crm
make migrate-up

# 3. Generate protobuf code
make proto-gen

# 4. Setup Python project
cd python-adk
poetry install

# 5. Start services
make api          # Go API (port 8080)
make grpc-server  # gRPC server (port 50051)
make mcp-server   # MCP server (port 9000)

# 6. Start Python ADK
cd python-adk
poetry run python -m ventros_adk.main

# 7. Run tests
make test         # Go tests
cd python-adk && poetry run pytest  # Python tests
```

---

## 📦 DELIVERABLES CHECKLIST

### Phase 1: Memory Service ✅
- [ ] Migration 000050 (memory_embeddings)
- [ ] Migration 000051 (memory_facts)
- [ ] Migration 000052 (retrieval_strategies)
- [ ] Domain layer (memory aggregate)
- [ ] Infrastructure layer (hybrid search, RRF)
- [ ] Application layer (use cases)
- [ ] Tests (70%+ coverage)
- [ ] Documentation

### Phase 2: MCP Server ✅
- [ ] HTTP server with SSE streaming
- [ ] JWT authentication
- [ ] Tool registry + executor
- [ ] 7 BI tools
- [ ] 5 Memory tools
- [ ] 8 CRM operations tools
- [ ] 5 Agent analysis tools
- [ ] 5 Document tools
- [ ] OpenAPI documentation
- [ ] Integration tests

### Phase 3: gRPC API ✅
- [ ] Protobuf definitions
- [ ] Go gRPC server
- [ ] Python gRPC client
- [ ] Interceptors (auth, logging)
- [ ] Health checks
- [ ] Benchmarks

### Phase 4: Python ADK ✅
- [ ] Poetry project setup
- [ ] Base agent class
- [ ] Coordinator agent
- [ ] 5 specialist agents
- [ ] Semantic router
- [ ] Memory service facade
- [ ] RabbitMQ consumer/publisher
- [ ] Temporal workflows
- [ ] Phoenix observability
- [ ] Tests (70%+ coverage)

### Phase 5: Advanced Memory ✅
- [ ] Apache AGE setup
- [ ] Graph schema
- [ ] Graph search queries
- [ ] Facts extraction (NER)
- [ ] Contradiction resolution
- [ ] Reranking (Jina v2)

### Phase 6: Templates & Polish ✅
- [ ] 10+ agent templates
- [ ] Template instantiation API
- [ ] Retrieval strategy tuning
- [ ] A/B testing framework
- [ ] Final documentation

---

## 🎓 LEARNING RESOURCES

### Memory & Search
- [pgvector documentation](https://github.com/pgvector/pgvector)
- [Apache AGE documentation](https://age.apache.org/)
- [Reciprocal Rank Fusion paper](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)
- [Hybrid search patterns](https://www.pinecone.io/learn/hybrid-search/)

### MCP (Model Context Protocol)
- [MCP specification](https://modelcontextprotocol.io/)
- [MCP Go implementation examples](https://github.com/modelcontextprotocol/servers)

### Google ADK
- [Google Generative AI Python SDK](https://ai.google.dev/gemini-api/docs/sdks)
- [ADK 0.5 documentation](https://ai.google.dev/)

### gRPC
- [gRPC Go tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC Python tutorial](https://grpc.io/docs/languages/python/quickstart/)
- [Protocol Buffers guide](https://protobuf.dev/)

### Multi-Agent Systems
- [Semantic Router](https://github.com/aurelio-labs/semantic-router)
- [LangChain agents](https://python.langchain.com/docs/modules/agents/)
- [Phoenix observability](https://docs.arize.com/phoenix/)

---

## 🔍 TROUBLESHOOTING

### Common Issues

**pgvector not installed**:
```bash
# macOS
brew install pgvector

# Ubuntu
sudo apt-get install postgresql-16-pgvector
```

**Apache AGE compilation fails**:
```bash
# Ensure PostgreSQL headers installed
sudo apt-get install postgresql-server-dev-16
```

**gRPC import errors (Python)**:
```bash
# Regenerate protobuf code
make proto-python
```

**RabbitMQ connection refused**:
```bash
# Check RabbitMQ is running
sudo systemctl status rabbitmq-server

# Check credentials in .env
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASS=guest
```

---

## 📞 SUPPORT

For questions or issues:
- GitHub Issues: https://github.com/ventros/crm/issues
- Documentation: https://docs.ventros.cloud
- Email: dev@ventros.cloud

---

**Last Updated**: 2025-10-13
**Status**: Ready for implementation
**Next Steps**: Begin Phase 1 (Memory Service)
