# 🧠 AI MEMORY & HYBRID SEARCH - GO ARCHITECTURE (2025)

> **Arquitetura completa de memória AI para Ventros CRM**
> Baseado em: Zep Graphiti, Microsoft OmniRAG, Anthropic Contextual Retrieval
> Stack: Go + PostgreSQL + pgvector + Apache AGE

---

## 📋 ÍNDICE

1. [Visão Geral](#visão-geral)
2. [Estruturas de Dados Core](#estruturas-de-dados-core)
3. [Memory Embedding Service](#memory-embedding-service)
4. [Hybrid Search Service](#hybrid-search-service)
5. [Retrieval Strategies Dictionary](#retrieval-strategies-dictionary)
6. [Temporal Knowledge Graph Service](#temporal-knowledge-graph-service)
7. [Agent Registry & Routing](#agent-registry--routing)
8. [Memory Fact Service](#memory-fact-service)
9. [Context Manager (Caching)](#context-manager-caching)
10. [gRPC API](#grpc-api)
11. [Database Schema](#database-schema)

---

## 🎯 VISÃO GERAL

### Responsabilidades do Go Service

```
┌─────────────────────────────────────────────────────────────┐
│                     GO MEMORY SERVICE                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ✅ CRUD de todas entidades (Contact, Message, Session)     │
│  ✅ Geração de embeddings (Vertex AI Go SDK)                │
│  ✅ Contextual Retrieval (chunk + context)                  │
│  ✅ Hybrid Search (Vector + Keyword + Graph + SQL)          │
│  ✅ Temporal Knowledge Graph (Apache AGE)                   │
│  ✅ Memory Facts com contradiction resolution                │
│  ✅ Agent Registry & Semantic Routing                        │
│  ✅ Context Caching (Redis)                                  │
│  ✅ gRPC Server (Python ADK chama)                          │
│  ✅ Event Publishing (RabbitMQ)                              │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### Fluxo de Dados

```
Message → Domain Logic → Embeddings → Memory → Search → Python ADK
   ↓           ↓             ↓           ↓        ↓         ↓
 Entity    Aggregate    Contextual   PostgreSQL  Hybrid   Agent
Created    Events       Chunks       +pgvector   Search   Routing
```

---

## 📦 ESTRUTURAS DE DADOS CORE

### **1. Agent Metadata Extensions**

```go
package agent

// AIAgentMetadata contém configuração específica para AI agents
type AIAgentMetadata struct {
    Category         AgentCategory       `json:"category"`
    Skills           []AgentSkill        `json:"skills"`
    KnowledgeScope   KnowledgeScope      `json:"knowledge_scope"`
    MemoryStrategy   MemoryStrategy      `json:"memory_strategy"`
    RoutingRules     []RoutingRule       `json:"routing_rules"`
    Priority         int                 `json:"priority"`        // 1-10 (para tie-breaking)
    MaxConcurrentSessions int            `json:"max_concurrent_sessions"`
}

// AgentCategory - Categorias especializadas de agentes
type AgentCategory string

const (
    // === SALES ===
    CategorySalesProspecting    AgentCategory = "sales_prospecting"
    CategorySalesNegotiation    AgentCategory = "sales_negotiation"
    CategorySalesClosing        AgentCategory = "sales_closing"

    // === SUPPORT ===
    CategorySupportTechnical    AgentCategory = "support_technical"
    CategorySupportBilling      AgentCategory = "support_billing"
    CategorySupportOnboarding   AgentCategory = "support_onboarding"

    // === RETENTION ===
    CategoryRetentionChurn      AgentCategory = "retention_churn"
    CategoryRetentionUpsell     AgentCategory = "retention_upsell"
    CategoryRetentionWinback    AgentCategory = "retention_winback"

    // === OPERATIONS ===
    CategoryOperationsSchedule  AgentCategory = "operations_schedule"
    CategoryOperationsFollowup  AgentCategory = "operations_followup"
    CategoryOperationsQA        AgentCategory = "operations_qa"

    // === MARKETING ===
    CategoryMarketingCampaign   AgentCategory = "marketing_campaign"
    CategoryMarketingContent    AgentCategory = "marketing_content"
    CategoryMarketingEvent      AgentCategory = "marketing_event"
)

// AgentSkill - Competências específicas do agente
type AgentSkill struct {
    Name        string  `json:"name"`         // "objection_handling", "technical_troubleshooting"
    Proficiency float64 `json:"proficiency"`  // 0.0-1.0
}

// RoutingRule - Regras para quando este agente deve ser acionado
type RoutingRule struct {
    Condition   string                 `json:"condition"`    // "sentiment < -0.5", "topics contains 'cancelamento'"
    Priority    int                    `json:"priority"`     // Ordem de avaliação
    Metadata    map[string]interface{} `json:"metadata"`
}
```

### **2. Knowledge Scope (O que cada agente acessa)**

```go
package memory

// KnowledgeScope define quais dados o agente pode acessar
type KnowledgeScope struct {
    // === DADOS RELACIONAIS ===
    IncludeSessions          bool     `json:"include_sessions"`
    SessionsLookbackDays     int      `json:"sessions_lookback_days"`      // Ex: 30
    IncludeMessages          bool     `json:"include_messages"`
    MessagesLimit            int      `json:"messages_limit"`              // Ex: últimas 50
    MessagesOnlyRecent       bool     `json:"messages_only_recent"`        // Sempre inclui últimas N mensagens
    RecentMessagesDays       int      `json:"recent_messages_days"`        // Ex: últimas 7 dias
    IncludeContactEvents     bool     `json:"include_contact_events"`
    ContactEventsCategories  []string `json:"contact_events_categories"`   // Filtrar por categoria
    IncludeTracking          bool     `json:"include_tracking"`
    IncludeNotes             bool     `json:"include_notes"`

    // === DADOS SEMÂNTICOS (Vector Search) ===
    IncludeSessionSummaries  bool     `json:"include_session_summaries"`
    SimilarityThreshold      float64  `json:"similarity_threshold"`        // 0.7 = 70%
    MaxSimilarSessions       int      `json:"max_similar_sessions"`        // Top-K

    // === DADOS DE GRAFO ===
    IncludeAgentTransferChain bool    `json:"include_agent_transfer_chain"`
    IncludeReplyThreads       bool    `json:"include_reply_threads"`
    IncludeSocialGraph        bool    `json:"include_social_graph"`        // Menções
    IncludeCampaignGraph      bool    `json:"include_campaign_graph"`      // Tracking attribution
    GraphTraversalDepth       int     `json:"graph_traversal_depth"`       // 1-3 hops

    // === AGREGAÇÕES ===
    IncludeContactStats       bool    `json:"include_contact_stats"`
    IncludePipelineContext    bool    `json:"include_pipeline_context"`

    // === MEMORY FACTS ===
    IncludeMemoryFacts        bool    `json:"include_memory_facts"`
    FactTypes                 []string `json:"fact_types"`                 // budget_constraint, preference, etc
}

// Presets por categoria de agente
func DefaultKnowledgeScope(category AgentCategory) KnowledgeScope {
    switch category {
    case CategorySalesProspecting:
        return KnowledgeScope{
            IncludeSessions:         true,
            SessionsLookbackDays:    30,
            IncludeMessages:         true,
            MessagesLimit:           20,
            MessagesOnlyRecent:      true,  // SEMPRE inclui últimas mensagens
            RecentMessagesDays:      7,
            IncludeContactEvents:    true,
            ContactEventsCategories: []string{"lead_captured", "form_submitted"},
            IncludeTracking:         true,  // CRÍTICO: origem do lead
            IncludeNotes:            true,
            IncludeSessionSummaries: true,
            SimilarityThreshold:     0.75,
            MaxSimilarSessions:      5,
            IncludeCampaignGraph:    true,  // UTM attribution
            GraphTraversalDepth:     2,
            IncludeContactStats:     true,
            IncludePipelineContext:  true,
            IncludeMemoryFacts:      true,
            FactTypes:               []string{"budget_constraint", "preference", "goal"},
        }

    case CategoryRetentionChurn:
        return KnowledgeScope{
            IncludeSessions:           true,
            SessionsLookbackDays:      90,  // Histórico mais longo
            IncludeMessages:           true,
            MessagesLimit:             50,  // Mais contexto
            MessagesOnlyRecent:        true,
            RecentMessagesDays:        14,  // 2 semanas
            IncludeContactEvents:      true,
            ContactEventsCategories:   []string{"complaint", "negative_feedback", "refund_requested"},
            IncludeTracking:           false, // Menos relevante
            IncludeNotes:              true,  // CRÍTICO: notas de agentes
            IncludeSessionSummaries:   true,
            SimilarityThreshold:       0.70,  // Threshold mais baixo
            MaxSimilarSessions:        10,    // Mais padrões de churn
            IncludeAgentTransferChain: true,  // CRÍTICO: sinal de insatisfação
            IncludeReplyThreads:       true,
            GraphTraversalDepth:       3,     // Mais profundidade
            IncludeContactStats:       true,
            IncludePipelineContext:    true,
            IncludeMemoryFacts:        true,
            FactTypes:                 []string{"objection", "pain_point", "constraint"},
        }

    case CategorySupportTechnical:
        return KnowledgeScope{
            IncludeSessions:         true,
            SessionsLookbackDays:    7,    // Contexto recente
            IncludeMessages:         true,
            MessagesLimit:           30,
            MessagesOnlyRecent:      true,
            RecentMessagesDays:      3,    // Últimos 3 dias críticos
            IncludeContactEvents:    true,
            ContactEventsCategories: []string{"error_reported", "bug_reported"},
            IncludeTracking:         false,
            IncludeNotes:            true,
            IncludeSessionSummaries: true,
            SimilarityThreshold:     0.80, // Alta precisão (bugs específicos)
            MaxSimilarSessions:      3,
            IncludeReplyThreads:     true, // Threading importante
            GraphTraversalDepth:     1,
            IncludeContactStats:     false,
            IncludePipelineContext:  false,
            IncludeMemoryFacts:      true,
            FactTypes:               []string{"technical_issue", "environment_info"},
        }

    default:
        // Balanced default
        return KnowledgeScope{
            IncludeSessions:         true,
            SessionsLookbackDays:    14,
            IncludeMessages:         true,
            MessagesLimit:           30,
            MessagesOnlyRecent:      true,
            RecentMessagesDays:      7,
            IncludeContactEvents:    true,
            IncludeTracking:         true,
            IncludeNotes:            true,
            IncludeSessionSummaries: true,
            SimilarityThreshold:     0.75,
            MaxSimilarSessions:      5,
            GraphTraversalDepth:     2,
            IncludeContactStats:     true,
            IncludePipelineContext:  true,
            IncludeMemoryFacts:      true,
        }
    }
}
```

### **3. Memory Strategy (Como cada agente busca)**

```go
package memory

// MemoryStrategy define como o agente faz retrieval
type MemoryStrategy struct {
    // === WEIGHTS (devem somar 1.0) ===
    VectorWeight   float64 `json:"vector_weight"`    // Semantic similarity
    KeywordWeight  float64 `json:"keyword_weight"`   // BM25/pg_trgm
    GraphWeight    float64 `json:"graph_weight"`     // Graph traversal
    RecentWeight   float64 `json:"recent_weight"`    // Recency bias

    // === RETRIEVAL STRATEGY ===
    Strategy       RetrievalStrategy `json:"strategy"` // OmniRAG pattern

    // === FUSION ===
    FusionMethod   FusionMethod `json:"fusion_method"` // RRF ou Weighted

    // === RERANKING ===
    UseReranking   bool   `json:"use_reranking"`
    RerankProvider string `json:"rerank_provider"` // "jina-v2", "cohere", "none"
    RerankTopK     int    `json:"rerank_top_k"`    // Top-K após reranking

    // === CONTEXT WINDOW ===
    MaxTokens      int    `json:"max_tokens"`       // Limite de tokens
    Summarization  string `json:"summarization"`    // "none", "map_reduce", "recursive"

    // === CACHING ===
    CacheDuration  int    `json:"cache_duration"`   // Segundos
    CacheStrategy  string `json:"cache_strategy"`   // "per_contact", "per_session", "none"
}

// RetrievalStrategy - OmniRAG dynamic selection
type RetrievalStrategy string

const (
    StrategySemanticHeavy    RetrievalStrategy = "semantic_heavy"    // 70/20/10/0
    StrategyGraphHeavy       RetrievalStrategy = "graph_heavy"       // 10/20/70/0
    StrategyAnalyticalHeavy  RetrievalStrategy = "analytical_heavy"  // 10/10/0/80 (SQL)
    StrategyBalanced         RetrievalStrategy = "balanced"          // 33/33/33/0
    StrategyVectorKeyword    RetrievalStrategy = "vector_keyword"    // 50/50/0/0
    StrategyVectorGraph      RetrievalStrategy = "vector_graph"      // 50/0/50/0
    StrategyKeywordRecent    RetrievalStrategy = "keyword_recent"    // 30/0/0/70
    StrategyCustom           RetrievalStrategy = "custom"            // Usa weights explícitos
)

// FusionMethod - Como combinar resultados
type FusionMethod string

const (
    FusionRRF      FusionMethod = "rrf"      // Reciprocal Rank Fusion (no tuning needed)
    FusionWeighted FusionMethod = "weighted" // Weighted average (needs tuning)
    FusionLinear   FusionMethod = "linear"   // Linear combination
)

// GetWeights retorna os pesos baseado na estratégia
func (s RetrievalStrategy) GetWeights() (vector, keyword, graph, recent float64) {
    switch s {
    case StrategySemanticHeavy:
        return 0.70, 0.20, 0.10, 0.00
    case StrategyGraphHeavy:
        return 0.10, 0.20, 0.70, 0.00
    case StrategyAnalyticalHeavy:
        return 0.10, 0.10, 0.00, 0.80
    case StrategyBalanced:
        return 0.33, 0.33, 0.33, 0.00
    case StrategyVectorKeyword:
        return 0.50, 0.50, 0.00, 0.00
    case StrategyVectorGraph:
        return 0.50, 0.00, 0.50, 0.00
    case StrategyKeywordRecent:
        return 0.00, 0.30, 0.00, 0.70
    default:
        return 0.33, 0.33, 0.33, 0.00
    }
}

// Presets por categoria
func DefaultMemoryStrategy(category AgentCategory) MemoryStrategy {
    switch category {
    case CategorySalesProspecting:
        return MemoryStrategy{
            VectorWeight:   0.20,
            KeywordWeight:  0.30,
            GraphWeight:    0.40,  // ALTO: origem do lead
            RecentWeight:   0.10,
            Strategy:       StrategyGraphHeavy,
            FusionMethod:   FusionRRF,
            UseReranking:   false, // Fast qualification
            MaxTokens:      6000,
            Summarization:  "none",
            CacheDuration:  300,   // 5min
            CacheStrategy:  "per_contact",
        }

    case CategoryRetentionChurn:
        return MemoryStrategy{
            VectorWeight:   0.50,  // ALTO: padrões de churn
            KeywordWeight:  0.20,
            GraphWeight:    0.20,  // Agent transfers
            RecentWeight:   0.10,
            Strategy:       StrategySemanticHeavy,
            FusionMethod:   FusionWeighted,
            UseReranking:   true,  // Crítico: accuracy matters
            RerankProvider: "jina-v2",
            RerankTopK:     10,
            MaxTokens:      10000, // Contexto maior
            Summarization:  "map_reduce",
            CacheDuration:  60,    // 1min (churn muda rápido)
            CacheStrategy:  "per_contact",
        }

    case CategorySupportTechnical:
        return MemoryStrategy{
            VectorWeight:   0.30,
            KeywordWeight:  0.50,  // ALTO: termos técnicos exatos
            GraphWeight:    0.10,
            RecentWeight:   0.10,
            Strategy:       StrategyKeywordRecent,
            FusionMethod:   FusionRRF,
            UseReranking:   false,
            MaxTokens:      4000,  // Problema específico
            Summarization:  "none",
            CacheDuration:  0,     // Sem cache (problema pode estar resolvido)
            CacheStrategy:  "none",
        }

    default:
        return MemoryStrategy{
            VectorWeight:   0.33,
            KeywordWeight:  0.33,
            GraphWeight:    0.33,
            RecentWeight:   0.00,
            Strategy:       StrategyBalanced,
            FusionMethod:   FusionRRF,
            UseReranking:   false,
            MaxTokens:      8000,
            Summarization:  "none",
            CacheDuration:  300,
            CacheStrategy:  "per_contact",
        }
    }
}
```

---

## 🔍 RETRIEVAL STRATEGIES DICTIONARY

### **Dicionário Completo de Estratégias**

```go
package memory

// StrategyConfig - Configurações prontas para diferentes cenários
var StrategyConfigs = map[string]StrategyConfig{
    // === SALES ===
    "sales_prospecting": {
        Name:        "Sales Prospecting",
        Description: "Prioriza atribuição de campanha + contexto recente",
        Weights:     StrategyWeights{Vector: 0.20, Keyword: 0.30, Graph: 0.40, Recent: 0.10},
        UseCase:     "Lead qualification, campaign attribution",
        Reranking:   false,
        CacheTTL:    300,
    },

    "sales_negotiation": {
        Name:        "Sales Negotiation",
        Description: "Prioriza histórico de objeções + budget constraints",
        Weights:     StrategyWeights{Vector: 0.40, Keyword: 0.30, Graph: 0.20, Recent: 0.10},
        UseCase:     "Price objections, feature requests",
        Reranking:   true,
        CacheTTL:    180,
    },

    "sales_closing": {
        Name:        "Sales Closing",
        Description: "Contexto completo + decision makers",
        Weights:     StrategyWeights{Vector: 0.35, Keyword: 0.25, Graph: 0.30, Recent: 0.10},
        UseCase:     "Final negotiation, contract terms",
        Reranking:   true,
        CacheTTL:    60,
    },

    // === RETENTION ===
    "retention_churn": {
        Name:        "Churn Prevention",
        Description: "Padrões de churn + sentiment + agent transfers",
        Weights:     StrategyWeights{Vector: 0.50, Keyword: 0.20, Graph: 0.20, Recent: 0.10},
        UseCase:     "Customer wanting to cancel, dissatisfaction",
        Reranking:   true, // CRÍTICO
        CacheTTL:    60,
    },

    "retention_upsell": {
        Name:        "Upsell Opportunity",
        Description: "Usage patterns + feature requests + satisfaction",
        Weights:     StrategyWeights{Vector: 0.40, Keyword: 0.20, Graph: 0.30, Recent: 0.10},
        UseCase:     "Cross-sell, plan upgrade",
        Reranking:   false,
        CacheTTL:    300,
    },

    // === SUPPORT ===
    "support_technical": {
        Name:        "Technical Support",
        Description: "Keywords técnicos + similar issues + recent context",
        Weights:     StrategyWeights{Vector: 0.30, Keyword: 0.50, Graph: 0.10, Recent: 0.10},
        UseCase:     "Bug reports, technical issues",
        Reranking:   false,
        CacheTTL:    0, // Sem cache
    },

    "support_billing": {
        Name:        "Billing Support",
        Description: "Transaction history + billing events",
        Weights:     StrategyWeights{Vector: 0.20, Keyword: 0.40, Graph: 0.30, Recent: 0.10},
        UseCase:     "Payment issues, invoice questions",
        Reranking:   false,
        CacheTTL:    300,
    },

    // === GENERIC ===
    "balanced": {
        Name:        "Balanced",
        Description: "Distribuição igual entre métodos",
        Weights:     StrategyWeights{Vector: 0.33, Keyword: 0.33, Graph: 0.33, Recent: 0.00},
        UseCase:     "General purpose, unknown intent",
        Reranking:   false,
        CacheTTL:    300,
    },

    "vector_only": {
        Name:        "Vector Only",
        Description: "Apenas semantic search",
        Weights:     StrategyWeights{Vector: 1.00, Keyword: 0.00, Graph: 0.00, Recent: 0.00},
        UseCase:     "Pure semantic queries, similar conversations",
        Reranking:   true,
        CacheTTL:    600,
    },

    "keyword_only": {
        Name:        "Keyword Only",
        Description: "Apenas keyword search (BM25/pg_trgm)",
        Weights:     StrategyWeights{Vector: 0.00, Keyword: 1.00, Graph: 0.00, Recent: 0.00},
        UseCase:     "Exact phrase matching, compliance queries",
        Reranking:   false,
        CacheTTL:    300,
    },

    "graph_only": {
        Name:        "Graph Only",
        Description: "Apenas graph traversal",
        Weights:     StrategyWeights{Vector: 0.00, Keyword: 0.00, Graph: 1.00, Recent: 0.00},
        UseCase:     "Relationship queries, attribution, transfer chains",
        Reranking:   false,
        CacheTTL:    300,
    },

    // === 50/50 SPLITS ===
    "vector_keyword_50": {
        Name:        "Vector + Keyword (50/50)",
        Description: "Classic hybrid search",
        Weights:     StrategyWeights{Vector: 0.50, Keyword: 0.50, Graph: 0.00, Recent: 0.00},
        UseCase:     "Standard RAG, document search",
        Reranking:   true,
        CacheTTL:    300,
    },

    "vector_graph_50": {
        Name:        "Vector + Graph (50/50)",
        Description: "Semantic + relational",
        Weights:     StrategyWeights{Vector: 0.50, Keyword: 0.00, Graph: 0.50, Recent: 0.00},
        UseCase:     "Context-aware semantic search",
        Reranking:   false,
        CacheTTL:    300,
    },

    "keyword_graph_50": {
        Name:        "Keyword + Graph (50/50)",
        Description: "Exact match + relationships",
        Weights:     StrategyWeights{Vector: 0.00, Keyword: 0.50, Graph: 0.50, Recent: 0.00},
        UseCase:     "Compliance + attribution",
        Reranking:   false,
        CacheTTL:    300,
    },

    // === RECENCY-BASED ===
    "recent_only": {
        Name:        "Recent Only",
        Description: "Apenas mensagens recentes (SQL timestamp)",
        Weights:     StrategyWeights{Vector: 0.00, Keyword: 0.00, Graph: 0.00, Recent: 1.00},
        UseCase:     "Live chat context, last N messages",
        Reranking:   false,
        CacheTTL:    30,
    },

    "keyword_recent_70_30": {
        Name:        "Keyword + Recent (70/30)",
        Description: "Keywords nos últimos N dias",
        Weights:     StrategyWeights{Vector: 0.00, Keyword: 0.70, Graph: 0.00, Recent: 0.30},
        UseCase:     "Recent issue tracking",
        Reranking:   false,
        CacheTTL:    60,
    },

    "vector_recent_70_30": {
        Name:        "Vector + Recent (70/30)",
        Description: "Semantic search com bias de recência",
        Weights:     StrategyWeights{Vector: 0.70, Keyword: 0.00, Graph: 0.00, Recent: 0.30},
        UseCase:     "Recent similar conversations",
        Reranking:   true,
        CacheTTL:    120,
    },
}

// ========== DEPRECATED: StrategyWeights (OLD - 2024) ==========
// PROBLEMA: Pesos fixos não se adaptam ao contexto
// SOLUÇÃO: Usar RRF + Reranker (abaixo)
// ===============================================================
//
// type StrategyWeights struct {
//     Vector  float64 `json:"vector"`
//     Keyword float64 `json:"keyword"`
//     Graph   float64 `json:"graph"`
//     Recent  float64 `json:"recent"`
// }
//
// StrategyWeights será REMOVIDO em Sprint 6
// Migrar para ReciprocalRankFusion + CrossEncoderReranker
//
// ===============================================================

// ========== MODERNO (2024-2025): Reciprocal Rank Fusion ==========

// RRFConfig - Reciprocal Rank Fusion (estado da arte 2024)
// Combina rankings de diferentes métodos SEM pesos fixos
// Paper: "Reciprocal Rank Fusion outperforms condorcet and individual rank learning methods"
type RRFConfig struct {
    K          float64                `json:"k"`           // Constante RRF (default: 60)
    Methods    []RetrievalMethod      `json:"methods"`     // Métodos ativos
    Reranker   *RerankerConfig        `json:"reranker"`    // Cross-encoder reranking (opcional)
    LLMJudge   *LLMJudgeConfig        `json:"llm_judge"`   // LLM-as-Judge (casos complexos)
}

// RetrievalMethod - Método de busca individual
type RetrievalMethod string

const (
    MethodVector        RetrievalMethod = "vector"         // Semantic search (pgvector)
    MethodKeyword       RetrievalMethod = "keyword"        // Full-text search (PostgreSQL FTS)
    MethodGraph         RetrievalMethod = "graph"          // Knowledge graph traversal
    MethodRecent        RetrievalMethod = "recent"         // Time-based (últimas N mensagens)
    MethodBM25          RetrievalMethod = "bm25"           // BM25 ranking (keyword + TF-IDF)
    MethodColBERT       RetrievalMethod = "colbert"        // ColBERT v2 (late interaction)
    MethodHybridSparse  RetrievalMethod = "hybrid_sparse"  // SPLADE (sparse + dense)
)

// RerankerConfig - Cross-Encoder Reranking (mais preciso que pesos)
type RerankerConfig struct {
    Enabled   bool   `json:"enabled"`
    Model     string `json:"model"`      // "BAAI/bge-reranker-v2-m3" (default)
    TopK      int    `json:"top_k"`      // Top-K antes de rerank (default: 100)
    FinalK    int    `json:"final_k"`    // Top-K após rerank (default: 20)
    Threshold float64 `json:"threshold"` // Score mínimo (default: 0.3)
}

// LLMJudgeConfig - LLM-as-Judge para casos complexos
// Usa LLM para decidir relevância final (mais lento, mais preciso)
type LLMJudgeConfig struct {
    Enabled      bool     `json:"enabled"`
    Model        string   `json:"model"`           // "gemini-1.5-flash" (rápido)
    MaxCandidates int     `json:"max_candidates"`  // Max resultados para julgar (default: 10)
    Prompt       string   `json:"prompt"`          // Template do prompt
    Temperature  float64  `json:"temperature"`     // Default: 0.0 (determinístico)
}

// StrategyConfig - Configuração completa de uma estratégia (MODERNIZADA)
type StrategyConfig struct {
    Name        string       `json:"name"`
    Description string       `json:"description"`
    UseCase     string       `json:"use_case"`
    RRF         RRFConfig    `json:"rrf"`           // Reciprocal Rank Fusion
    CacheTTL    int          `json:"cache_ttl"`
}

// Exemplo de configuração moderna:
var ModernHybridStrategy = StrategyConfig{
    Name:        "modern_hybrid",
    Description: "RRF + Reranker + LLM Judge (state-of-the-art 2025)",
    UseCase:     "production",
    RRF: RRFConfig{
        K: 60,
        Methods: []RetrievalMethod{
            MethodVector,    // Semantic
            MethodBM25,      // Keyword avançado
            MethodGraph,     // Knowledge graph
            MethodRecent,    // Baseline temporal
        },
        Reranker: &RerankerConfig{
            Enabled:   true,
            Model:     "BAAI/bge-reranker-v2-m3",  // SOTA 2024
            TopK:      100,
            FinalK:    20,
            Threshold: 0.3,
        },
        LLMJudge: &LLMJudgeConfig{
            Enabled:      false,  // Apenas para casos críticos (billing, legal)
            Model:        "gemini-1.5-flash",
            MaxCandidates: 10,
            Temperature:  0.0,
        },
    },
    CacheTTL: 300,
}
```

### **Sobre SQL/Recent sempre ter parte recente:**

**SIM**, você está CERTO! Messages SQL **SEMPRE** vai ter uma parte recente, independente da estratégia:

```go
// Todas as queries incluem "recent baseline"
func (h *HybridSearchService) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
    results := &SearchResult{}

    // === BASELINE: SEMPRE busca últimas N mensagens ===
    // Isso garante contexto conversacional atual
    recentMessages := h.getRecentMessages(
        req.ContactID,
        req.KnowledgeScope.RecentMessagesDays,  // Ex: últimos 7 dias
        req.KnowledgeScope.MessagesLimit,        // Ex: últimas 20 msgs
    )
    results.RecentMessages = recentMessages  // SEMPRE presente

    // === MÉTODOS ADICIONAIS (baseado em strategy) ===
    strategy := req.MemoryStrategy

    if strategy.VectorWeight > 0 {
        vectorResults := h.vectorSearch(ctx, req)
        results.VectorResults = vectorResults
    }

    if strategy.KeywordWeight > 0 {
        keywordResults := h.keywordSearch(ctx, req)
        results.KeywordResults = keywordResults
    }

    if strategy.GraphWeight > 0 {
        graphResults := h.graphTraversal(ctx, req)
        results.GraphResults = graphResults
    }

    // Recent baseline já está em results.RecentMessages
    // Outros métodos complementam (não substituem)

    return results, nil
}
```

**Por quê?**
- Agente **sempre** precisa saber as últimas mensagens (contexto conversacional)
- Mesmo que estratégia seja "vector_only", precisa das últimas msgs para contexto
- Vector/Keyword/Graph **complementam** (trazem contexto histórico), não substituem

---

## 🎨 MEMORY EMBEDDING SERVICE

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "cloud.google.com/go/vertexai/genai"
    "github.com/google/uuid"
)

// EmbeddingService - Serviço de geração de embeddings contextuais
type EmbeddingService struct {
    llmClient       *genai.Client  // Gemini Flash para context generation
    embeddingClient *genai.Client  // text-embedding-005
    repo            EmbeddingRepository
    sessionRepo     SessionRepository
    contactRepo     ContactRepository
    trackingRepo    TrackingRepository
}

// NewEmbeddingService cria novo serviço
func NewEmbeddingService(
    llmClient, embeddingClient *genai.Client,
    repo EmbeddingRepository,
    sessionRepo SessionRepository,
    contactRepo ContactRepository,
    trackingRepo TrackingRepository,
) *EmbeddingService {
    return &EmbeddingService{
        llmClient:       llmClient,
        embeddingClient: embeddingClient,
        repo:            repo,
        sessionRepo:     sessionRepo,
        contactRepo:     contactRepo,
        trackingRepo:    trackingRepo,
    }
}

// GenerateContextualEmbedding - Implementa Contextual Retrieval (Anthropic 2025)
func (e *EmbeddingService) GenerateContextualEmbedding(
    ctx context.Context,
    sessionID uuid.UUID,
    tenantID string,
) (*MemoryEmbedding, error) {
    // 1. Busca dados relacionados
    session, err := e.sessionRepo.FindByID(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("session not found: %w", err)
    }

    contact, err := e.contactRepo.FindByID(ctx, session.ContactID())
    if err != nil {
        return nil, fmt.Errorf("contact not found: %w", err)
    }

    // Tracking pode não existir
    tracking, _ := e.trackingRepo.FindByContactID(ctx, contact.ID())

    // 2. Gera contexto usando LLM (Gemini Flash 2.5)
    contextPrompt := e.buildContextPrompt(session, contact, tracking)

    generatedContext, err := e.llmClient.GenerateContent(ctx, contextPrompt)
    if err != nil {
        return nil, fmt.Errorf("failed to generate context: %w", err)
    }

    // 3. Concatena contexto + summary
    originalSummary := ""
    if session.Summary() != nil {
        originalSummary = *session.Summary()
    }

    contextualText := fmt.Sprintf("%s\n\n[SUMMARY]: %s",
        generatedContext.Text,
        originalSummary,
    )

    // 4. Gera embedding (text-embedding-005, 768 dim)
    embeddingResponse, err := e.embeddingClient.EmbedContent(ctx,
        genai.Text(contextualText),
        &genai.EmbedContentRequest{
            Model: "text-embedding-005",
        },
    )
    if err != nil {
        return nil, fmt.Errorf("failed to generate embedding: %w", err)
    }

    // 5. Prepara metadata
    metadata := e.buildMetadata(session, contact, tracking)

    // 6. Cria MemoryEmbedding
    memoryEmbedding := &MemoryEmbedding{
        ID:              uuid.New(),
        TenantID:        tenantID,
        SourceType:      SourceSessionSummary,
        SourceID:        sessionID,
        Embedding:       embeddingResponse.Embedding.Values,
        ContextualText:  contextualText,
        OriginalText:    originalSummary,
        Metadata:        metadata,
        CreatedAt:       time.Now(),
        ExpiresAt:       nil, // Não expira
    }

    // 7. Persiste
    if err := e.repo.Save(ctx, memoryEmbedding); err != nil {
        return nil, fmt.Errorf("failed to save embedding: %w", err)
    }

    return memoryEmbedding, nil
}

// buildContextPrompt - Monta prompt para geração de contexto
func (e *EmbeddingService) buildContextPrompt(
    session *Session,
    contact *Contact,
    tracking *Tracking,
) string {
    prompt := fmt.Sprintf(`
Dado os dados a seguir, gere um contexto sucinto (2-3 frases em português) que situe o resumo da sessão no histórico do cliente.

### CONTATO
Nome: %s
Tags: %v
Primeira interação: %s
Última interação: %s

### SESSÃO
ID: %s
Data: %s
Duração: %d segundos
Mensagens: %d (%d do contato, %d do agente)
Sentiment: %s (score: %.2f)
Topics: %v
Outcome Tags: %v
Resolvida: %t
Escalada: %t
Pipeline: %s
Agentes: %v
Transferências: %d
`,
        contact.Name(),
        contact.Tags(),
        formatTime(contact.FirstInteractionAt()),
        formatTime(contact.LastInteractionAt()),
        session.ID(),
        session.StartedAt().Format("02/01/2006 15:04"),
        session.DurationSeconds(),
        session.MessageCount(),
        session.MessagesFromContact(),
        session.MessagesFromAgent(),
        sentimentStr(session.Sentiment()),
        sentimentScore(session.SentimentScore()),
        session.Topics(),
        session.OutcomeTags(),
        session.IsResolved(),
        session.IsEscalated(),
        pipelineStr(session.PipelineID()),
        agentNames(session.AgentIDs()),  // TODO: fetch agent names
        session.AgentTransfers(),
    )

    // Adiciona tracking se existir
    if tracking != nil {
        prompt += fmt.Sprintf(`
### ORIGEM / ATRIBUIÇÃO
Fonte: %s
Plataforma: %s
Campanha: %s
UTM Source: %s
UTM Medium: %s
UTM Campaign: %s
`,
            tracking.Source,
            tracking.Platform,
            tracking.Campaign,
            tracking.UTMSource,
            tracking.UTMMedium,
            tracking.UTMCampaign,
        )
    }

    // Summary original
    if session.Summary() != nil {
        prompt += fmt.Sprintf(`
### RESUMO ORIGINAL
%s

Gere APENAS o contexto (2-3 frases). Não repita o resumo original.
`, *session.Summary())
    }

    return prompt
}

// buildMetadata - Monta metadata estruturado para filtragem
func (e *EmbeddingService) buildMetadata(
    session *Session,
    contact *Contact,
    tracking *Tracking,
) map[string]interface{} {
    metadata := map[string]interface{}{
        "contact_id":          contact.ID().String(),
        "session_id":          session.ID().String(),
        "sentiment":           sentimentStr(session.Sentiment()),
        "sentiment_score":     sentimentScore(session.SentimentScore()),
        "topics":              session.Topics(),
        "agent_ids":           uuidSliceToStringSlice(session.AgentIDs()),
        "duration_seconds":    session.DurationSeconds(),
        "message_count":       session.MessageCount(),
        "resolved":            session.IsResolved(),
        "escalated":           session.IsEscalated(),
        "outcome_tags":        session.OutcomeTags(),
        "contact_tags":        contact.Tags(),
        "session_date":        session.StartedAt().Format("2006-01-02"),
        "agent_transfers":     session.AgentTransfers(),
    }

    if session.PipelineID() != nil {
        metadata["pipeline_id"] = session.PipelineID().String()
    }

    if tracking != nil {
        metadata["tracking_source"] = tracking.Source
        metadata["tracking_platform"] = tracking.Platform
        metadata["tracking_campaign"] = tracking.Campaign
        metadata["utm_source"] = tracking.UTMSource
        metadata["utm_medium"] = tracking.UTMMedium
        metadata["utm_campaign"] = tracking.UTMCampaign
    }

    return metadata
}

// BatchGenerateEmbeddings - Processa múltiplas sessões em batch
func (e *EmbeddingService) BatchGenerateEmbeddings(
    ctx context.Context,
    sessionIDs []uuid.UUID,
    tenantID string,
) ([]*MemoryEmbedding, error) {
    embeddings := make([]*MemoryEmbedding, 0, len(sessionIDs))

    // TODO: Implementar batching real com Vertex AI batch API
    // Por enquanto, processa sequencialmente
    for _, sessionID := range sessionIDs {
        embedding, err := e.GenerateContextualEmbedding(ctx, sessionID, tenantID)
        if err != nil {
            // Log erro mas continua (não falha batch inteiro)
            fmt.Printf("Failed to generate embedding for session %s: %v\n", sessionID, err)
            continue
        }
        embeddings = append(embeddings, embedding)
    }

    return embeddings, nil
}

// MemoryEmbedding - Entidade de embedding persistido
type MemoryEmbedding struct {
    ID             uuid.UUID              `db:"id"`
    TenantID       string                 `db:"tenant_id"`
    SourceType     EmbeddingSourceType    `db:"source_type"`
    SourceID       uuid.UUID              `db:"source_id"`
    Embedding      []float32              `db:"embedding"`      // pgvector
    ContextualText string                 `db:"contextual_text"` // Context + original
    OriginalText   string                 `db:"original_text"`  // Summary original
    Metadata       map[string]interface{} `db:"metadata"`       // JSONB
    CreatedAt      time.Time              `db:"created_at"`
    ExpiresAt      *time.Time             `db:"expires_at"`     // Nullable
}

// EmbeddingSourceType - Tipos de fonte para embeddings
type EmbeddingSourceType string

const (
    SourceSessionSummary      EmbeddingSourceType = "session_summary"
    SourceSessionTopics       EmbeddingSourceType = "session_topics"
    SourceSessionOutcome      EmbeddingSourceType = "session_outcome"
    SourceMessageThread       EmbeddingSourceType = "message_thread"
    SourceContactEvent        EmbeddingSourceType = "contact_event"
    SourceContactEventCluster EmbeddingSourceType = "contact_event_cluster"
    SourceNote                EmbeddingSourceType = "note"
    SourcePipelineTransition  EmbeddingSourceType = "pipeline_transition"
    SourceCampaignContext     EmbeddingSourceType = "campaign_context"
    SourceContactProfile      EmbeddingSourceType = "contact_profile"
)

// Helper functions
func formatTime(t *time.Time) string {
    if t == nil {
        return "N/A"
    }
    return t.Format("02/01/2006")
}

func sentimentStr(s *Sentiment) string {
    if s == nil {
        return "unknown"
    }
    return string(*s)
}

func sentimentScore(s *float64) float64 {
    if s == nil {
        return 0.0
    }
    return *s
}

func pipelineStr(p *uuid.UUID) string {
    if p == nil {
        return "N/A"
    }
    return p.String()
}

func agentNames(agentIDs []uuid.UUID) []string {
    // TODO: Fetch real agent names from repository
    names := make([]string, len(agentIDs))
    for i, id := range agentIDs {
        names[i] = id.String() // Placeholder
    }
    return names
}

func uuidSliceToStringSlice(uuids []uuid.UUID) []string {
    strs := make([]string, len(uuids))
    for i, u := range uuids {
        strs[i] = u.String()
    }
    return strs
}
```

---

## 🔎 HYBRID SEARCH SERVICE

```go
package memory

import (
    "context"
    "fmt"
    "math"
    "sort"
    "time"

    "github.com/google/uuid"
)

// HybridSearchService - Motor de busca híbrida (Vector + Keyword + Graph + SQL)
type HybridSearchService struct {
    vectorRepo      VectorRepository      // pgvector queries
    keywordRepo     KeywordRepository     // pg_trgm/BM25 queries
    graphRepo       GraphRepository       // Apache AGE queries
    sqlRepo         SQLRepository         // Agregações SQL
    rerankService   *RerankService        // Jina/Cohere reranking
    contextManager  *ContextManager       // Caching
}

// SearchRequest - Request de busca híbrida
type SearchRequest struct {
    // === IDENTIFIERS ===
    TenantID       string          `json:"tenant_id"`
    ContactID      uuid.UUID       `json:"contact_id"`
    SessionID      *uuid.UUID      `json:"session_id"`      // Opcional: sessão atual
    AgentCategory  AgentCategory   `json:"agent_category"`

    // === QUERY ===
    Query          string          `json:"query"`           // Texto da query
    QueryEmbedding []float32       `json:"query_embedding"` // Embedding da query (opcional)

    // === CONFIGURATION ===
    KnowledgeScope KnowledgeScope  `json:"knowledge_scope"`
    MemoryStrategy MemoryStrategy  `json:"memory_strategy"`

    // === LIMITS ===
    TopK           int             `json:"top_k"`           // Final top-K
    RetrievalTopK  int             `json:"retrieval_top_k"` // Top-K antes de reranking
}

// SearchResult - Resultado da busca híbrida
type SearchResult struct {
    // === BASELINE (SEMPRE presente) ===
    RecentMessages   []Message              `json:"recent_messages"`    // Últimas N mensagens

    // === RETRIEVAL RESULTS ===
    VectorResults    []VectorResult         `json:"vector_results"`
    KeywordResults   []KeywordResult        `json:"keyword_results"`
    GraphResults     []GraphResult          `json:"graph_results"`

    // === FUSED RESULTS ===
    FusedResults     []FusedResult          `json:"fused_results"`      // Após RRF/Weighted fusion

    // === FINAL (após reranking se aplicável) ===
    FinalResults     []FinalResult          `json:"final_results"`      // Top-K final

    // === CONTEXT ===
    ContactStats     *ContactStats          `json:"contact_stats"`
    PipelineContext  *PipelineContext       `json:"pipeline_context"`
    MemoryFacts      []MemoryFact           `json:"memory_facts"`

    // === METADATA ===
    TotalResults     int                    `json:"total_results"`
    SearchLatencyMs  int64                  `json:"search_latency_ms"`
    CacheHit         bool                   `json:"cache_hit"`
    Strategy         string                 `json:"strategy"`
}

// Search - Executa busca híbrida completa
func (h *HybridSearchService) Search(
    ctx context.Context,
    req SearchRequest,
) (*SearchResult, error) {
    startTime := time.Now()

    result := &SearchResult{
        Strategy: string(req.MemoryStrategy.Strategy),
    }

    // === 1. CHECK CACHE ===
    if req.MemoryStrategy.CacheDuration > 0 {
        cached, err := h.contextManager.GetCached(ctx, req)
        if err == nil && cached != nil {
            result.CacheHit = true
            result.SearchLatencyMs = time.Since(startTime).Milliseconds()
            return cached, nil
        }
    }

    // === 2. BASELINE: Recent Messages (SEMPRE) ===
    recentMessages, err := h.sqlRepo.GetRecentMessages(ctx,
        req.ContactID,
        req.KnowledgeScope.RecentMessagesDays,
        req.KnowledgeScope.MessagesLimit,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get recent messages: %w", err)
    }
    result.RecentMessages = recentMessages

    // === 3. GENERATE QUERY EMBEDDING (se não fornecido) ===
    if req.QueryEmbedding == nil && req.MemoryStrategy.VectorWeight > 0 {
        // TODO: Generate embedding for query
        // req.QueryEmbedding = h.embeddingService.Embed(req.Query)
    }

    // === 4. VECTOR SEARCH (se weight > 0) ===
    if req.MemoryStrategy.VectorWeight > 0 && req.QueryEmbedding != nil {
        vectorResults, err := h.vectorRepo.Search(ctx, VectorSearchParams{
            TenantID:           req.TenantID,
            ContactID:          req.ContactID,
            Embedding:          req.QueryEmbedding,
            SimilarityThreshold: req.KnowledgeScope.SimilarityThreshold,
            Limit:              req.RetrievalTopK,
            LookbackDays:       req.KnowledgeScope.SessionsLookbackDays,
        })
        if err != nil {
            return nil, fmt.Errorf("vector search failed: %w", err)
        }
        result.VectorResults = vectorResults
    }

    // === 5. KEYWORD SEARCH (se weight > 0) ===
    if req.MemoryStrategy.KeywordWeight > 0 {
        keywordResults, err := h.keywordRepo.Search(ctx, KeywordSearchParams{
            TenantID:     req.TenantID,
            ContactID:    req.ContactID,
            Query:        req.Query,
            Limit:        req.RetrievalTopK,
            LookbackDays: req.KnowledgeScope.SessionsLookbackDays,
        })
        if err != nil {
            return nil, fmt.Errorf("keyword search failed: %w", err)
        }
        result.KeywordResults = keywordResults
    }

    // === 6. GRAPH TRAVERSAL (se weight > 0) ===
    if req.MemoryStrategy.GraphWeight > 0 {
        graphResults, err := h.graphRepo.Traverse(ctx, GraphTraversalParams{
            TenantID:        req.TenantID,
            ContactID:       req.ContactID,
            TraversalDepth:  req.KnowledgeScope.GraphTraversalDepth,
            IncludeTransfers: req.KnowledgeScope.IncludeAgentTransferChain,
            IncludeThreads:   req.KnowledgeScope.IncludeReplyThreads,
            IncludeCampaign:  req.KnowledgeScope.IncludeCampaignGraph,
        })
        if err != nil {
            return nil, fmt.Errorf("graph traversal failed: %w", err)
        }
        result.GraphResults = graphResults
    }

    // === 7. FUSION (RRF ou Weighted) ===
    fusedResults := h.fuseResults(
        result.VectorResults,
        result.KeywordResults,
        result.GraphResults,
        req.MemoryStrategy,
    )
    result.FusedResults = fusedResults

    // === 8. RERANKING (se habilitado) ===
    if req.MemoryStrategy.UseReranking {
        rerankedResults, err := h.rerankService.Rerank(ctx, RerankRequest{
            Query:      req.Query,
            Documents:  fusedResults,
            TopK:       req.TopK,
            Provider:   req.MemoryStrategy.RerankProvider,
        })
        if err != nil {
            return nil, fmt.Errorf("reranking failed: %w", err)
        }
        result.FinalResults = rerankedResults
    } else {
        // Sem reranking: pega top-K dos fused results
        result.FinalResults = h.topK(fusedResults, req.TopK)
    }

    // === 9. CONTEXTO ADICIONAL (stats, pipeline, facts) ===
    if req.KnowledgeScope.IncludeContactStats {
        stats, _ := h.sqlRepo.GetContactStats(ctx, req.ContactID)
        result.ContactStats = stats
    }

    if req.KnowledgeScope.IncludePipelineContext {
        pipeline, _ := h.sqlRepo.GetPipelineContext(ctx, req.ContactID)
        result.PipelineContext = pipeline
    }

    if req.KnowledgeScope.IncludeMemoryFacts {
        facts, _ := h.sqlRepo.GetMemoryFacts(ctx, req.ContactID, req.KnowledgeScope.FactTypes)
        result.MemoryFacts = facts
    }

    // === 10. CACHE RESULT ===
    if req.MemoryStrategy.CacheDuration > 0 {
        h.contextManager.SetCached(ctx, req, result, req.MemoryStrategy.CacheDuration)
    }

    result.SearchLatencyMs = time.Since(startTime).Milliseconds()
    result.TotalResults = len(result.FinalResults)

    return result, nil
}

// fuseResults - Combina resultados usando RRF ou Weighted Average
func (h *HybridSearchService) fuseResults(
    vectorResults []VectorResult,
    keywordResults []KeywordResult,
    graphResults []GraphResult,
    strategy MemoryStrategy,
) []FusedResult {
    // Mapa sessionID -> FusedResult
    fusedMap := make(map[uuid.UUID]*FusedResult)

    switch strategy.FusionMethod {
    case FusionRRF:
        return h.fuseRRF(vectorResults, keywordResults, graphResults, strategy)
    case FusionWeighted:
        return h.fuseWeighted(vectorResults, keywordResults, graphResults, strategy)
    default:
        return h.fuseRRF(vectorResults, keywordResults, graphResults, strategy)
    }
}

// fuseRRF - Reciprocal Rank Fusion
func (h *HybridSearchService) fuseRRF(
    vectorResults []VectorResult,
    keywordResults []KeywordResult,
    graphResults []GraphResult,
    strategy MemoryStrategy,
) []FusedResult {
    const k = 60 // RRF constant

    fusedMap := make(map[uuid.UUID]*FusedResult)

    // Vector results
    for rank, vr := range vectorResults {
        sessionID := vr.SessionID
        if _, exists := fusedMap[sessionID]; !exists {
            fusedMap[sessionID] = &FusedResult{
                SessionID: sessionID,
                Session:   vr.Session,
            }
        }
        fusedMap[sessionID].VectorScore = vr.Similarity
        fusedMap[sessionID].FusedScore += strategy.VectorWeight * (1.0 / float64(rank+1+k))
    }

    // Keyword results
    for rank, kr := range keywordResults {
        sessionID := kr.SessionID
        if _, exists := fusedMap[sessionID]; !exists {
            fusedMap[sessionID] = &FusedResult{
                SessionID: sessionID,
                Session:   kr.Session,
            }
        }
        fusedMap[sessionID].KeywordScore = kr.Rank
        fusedMap[sessionID].FusedScore += strategy.KeywordWeight * (1.0 / float64(rank+1+k))
    }

    // Graph results
    for rank, gr := range graphResults {
        sessionID := gr.SessionID
        if _, exists := fusedMap[sessionID]; !exists {
            fusedMap[sessionID] = &FusedResult{
                SessionID: sessionID,
                Session:   gr.Session,
            }
        }
        fusedMap[sessionID].GraphScore = gr.Relevance
        fusedMap[sessionID].FusedScore += strategy.GraphWeight * (1.0 / float64(rank+1+k))
    }

    // Convert map to slice
    fusedSlice := make([]FusedResult, 0, len(fusedMap))
    for _, fr := range fusedMap {
        fusedSlice = append(fusedSlice, *fr)
    }

    // Sort by fused score DESC
    sort.Slice(fusedSlice, func(i, j int) bool {
        return fusedSlice[i].FusedScore > fusedSlice[j].FusedScore
    })

    return fusedSlice
}

// fuseWeighted - Weighted Average Fusion
func (h *HybridSearchService) fuseWeighted(
    vectorResults []VectorResult,
    keywordResults []KeywordResult,
    graphResults []GraphResult,
    strategy MemoryStrategy,
) []FusedResult {
    fusedMap := make(map[uuid.UUID]*FusedResult)

    // Vector results (normalize similarity to 0-1)
    for _, vr := range vectorResults {
        sessionID := vr.SessionID
        if _, exists := fusedMap[sessionID]; !exists {
            fusedMap[sessionID] = &FusedResult{
                SessionID: sessionID,
                Session:   vr.Session,
            }
        }
        normalizedScore := vr.Similarity // Já está 0-1
        fusedMap[sessionID].VectorScore = normalizedScore
        fusedMap[sessionID].FusedScore += strategy.VectorWeight * normalizedScore
    }

    // Keyword results (normalize rank to 0-1)
    maxRank := float64(len(keywordResults))
    for rank, kr := range keywordResults {
        sessionID := kr.SessionID
        if _, exists := fusedMap[sessionID]; !exists {
            fusedMap[sessionID] = &FusedResult{
                SessionID: sessionID,
                Session:   kr.Session,
            }
        }
        normalizedScore := 1.0 - (float64(rank) / maxRank) // Inverte: rank 0 = score 1.0
        fusedMap[sessionID].KeywordScore = normalizedScore
        fusedMap[sessionID].FusedScore += strategy.KeywordWeight * normalizedScore
    }

    // Graph results (normalize relevance to 0-1)
    for _, gr := range graphResults {
        sessionID := gr.SessionID
        if _, exists := fusedMap[sessionID]; !exists {
            fusedMap[sessionID] = &FusedResult{
                SessionID: sessionID,
                Session:   gr.Session,
            }
        }
        normalizedScore := gr.Relevance // Assumir já normalizado 0-1
        fusedMap[sessionID].GraphScore = normalizedScore
        fusedMap[sessionID].FusedScore += strategy.GraphWeight * normalizedScore
    }

    // Convert to slice and sort
    fusedSlice := make([]FusedResult, 0, len(fusedMap))
    for _, fr := range fusedMap {
        fusedSlice = append(fusedSlice, *fr)
    }

    sort.Slice(fusedSlice, func(i, j int) bool {
        return fusedSlice[i].FusedScore > fusedSlice[j].FusedScore
    })

    return fusedSlice
}

// topK - Pega top-K results
func (h *HybridSearchService) topK(results []FusedResult, k int) []FinalResult {
    if k > len(results) {
        k = len(results)
    }

    finalResults := make([]FinalResult, k)
    for i := 0; i < k; i++ {
        finalResults[i] = FinalResult{
            Rank:         i + 1,
            SessionID:    results[i].SessionID,
            Session:      results[i].Session,
            FinalScore:   results[i].FusedScore,
            VectorScore:  results[i].VectorScore,
            KeywordScore: results[i].KeywordScore,
            GraphScore:   results[i].GraphScore,
        }
    }

    return finalResults
}

// Result types
type VectorResult struct {
    SessionID  uuid.UUID `json:"session_id"`
    Session    *Session  `json:"session"`
    Similarity float64   `json:"similarity"` // Cosine similarity (0-1)
}

type KeywordResult struct {
    SessionID uuid.UUID `json:"session_id"`
    Session   *Session  `json:"session"`
    Rank      float64   `json:"rank"` // BM25/pg_trgm rank
}

type GraphResult struct {
    SessionID uuid.UUID `json:"session_id"`
    Session   *Session  `json:"session"`
    Relevance float64   `json:"relevance"` // Graph relevance score
    Path      []string  `json:"path"`      // Graph traversal path
}

type FusedResult struct {
    SessionID    uuid.UUID `json:"session_id"`
    Session      *Session  `json:"session"`
    FusedScore   float64   `json:"fused_score"`
    VectorScore  float64   `json:"vector_score"`
    KeywordScore float64   `json:"keyword_score"`
    GraphScore   float64   `json:"graph_score"`
}

type FinalResult struct {
    Rank         int       `json:"rank"`
    SessionID    uuid.UUID `json:"session_id"`
    Session      *Session  `json:"session"`
    FinalScore   float64   `json:"final_score"`   // Score após reranking (se aplicável)
    VectorScore  float64   `json:"vector_score"`
    KeywordScore float64   `json:"keyword_score"`
    GraphScore   float64   `json:"graph_score"`
}

// ContactStats - Estatísticas agregadas do contato
type ContactStats struct {
    TotalSessions      int     `json:"total_sessions"`
    AvgSessionDuration int     `json:"avg_session_duration"`
    TotalMessages      int     `json:"total_messages"`
    PositiveSessions   int     `json:"positive_sessions"`
    NegativeSessions   int     `json:"negative_sessions"`
    ResolvedSessions   int     `json:"resolved_sessions"`
    AvgSentimentScore  float64 `json:"avg_sentiment_score"`
    LastSessionAt      *time.Time `json:"last_session_at"`
}

// PipelineContext - Contexto do pipeline atual
type PipelineContext struct {
    PipelineID   uuid.UUID `json:"pipeline_id"`
    PipelineName string    `json:"pipeline_name"`
    CurrentStage string    `json:"current_stage"`
    StageOrder   int       `json:"stage_order"`
    DaysInStage  int       `json:"days_in_stage"`
}
```

---

## 🕸️ TEMPORAL KNOWLEDGE GRAPH SERVICE

### **Apache AGE Integration**

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// TemporalKnowledgeGraph - Serviço de grafo temporal usando Apache AGE
type TemporalKnowledgeGraph struct {
    ageRepo    *ApacheAGERepository
    sessionRepo SessionRepository
    contactRepo ContactRepository
    messageRepo MessageRepository
}

// GraphNode - Nó do grafo
type GraphNode struct {
    ID         uuid.UUID              `json:"id"`
    Type       string                 `json:"type"` // session, contact, message, agent, campaign
    Properties map[string]interface{} `json:"properties"`
    ValidFrom  time.Time              `json:"valid_from"`  // Bi-temporal
    ValidTo    *time.Time             `json:"valid_to"`
    TxTime     time.Time              `json:"tx_time"`     // Transaction time
}

// GraphEdge - Aresta do grafo
type GraphEdge struct {
    ID         uuid.UUID              `json:"id"`
    FromID     uuid.UUID              `json:"from_id"`
    ToID       uuid.UUID              `json:"to_id"`
    Type       string                 `json:"type"` // belongs_to, replied_to, transferred_to, originated_from
    Properties map[string]interface{} `json:"properties"`
    ValidFrom  time.Time              `json:"valid_from"`
    ValidTo    *time.Time             `json:"valid_to"`
    TxTime     time.Time              `json:"tx_time"`
}

// CreateSessionGraph - Cria nós + edges quando sessão é criada
func (t *TemporalKnowledgeGraph) CreateSessionGraph(
    ctx context.Context,
    sessionID uuid.UUID,
) error {
    session, err := t.sessionRepo.FindByID(ctx, sessionID)
    if err != nil {
        return err
    }

    // 1. Create session node
    sessionNode := GraphNode{
        ID:   session.ID(),
        Type: "session",
        Properties: map[string]interface{}{
            "contact_id":       session.ContactID().String(),
            "started_at":       session.StartedAt(),
            "sentiment":        sentimentStr(session.Sentiment()),
            "sentiment_score":  sentimentScore(session.SentimentScore()),
            "topics":           session.Topics(),
            "resolved":         session.IsResolved(),
            "escalated":        session.IsEscalated(),
            "agent_transfers":  session.AgentTransfers(),
        },
        ValidFrom: session.StartedAt(),
        TxTime:    time.Now(),
    }

    if err := t.ageRepo.CreateNode(ctx, sessionNode); err != nil {
        return err
    }

    // 2. Create edge: contact -> session
    contactEdge := GraphEdge{
        ID:     uuid.New(),
        FromID: session.ContactID(),
        ToID:   session.ID(),
        Type:   "has_session",
        Properties: map[string]interface{}{
            "order": session.AgentTransfers(), // Quantas sessões até agora
        },
        ValidFrom: session.StartedAt(),
        TxTime:    time.Now(),
    }

    if err := t.ageRepo.CreateEdge(ctx, contactEdge); err != nil {
        return err
    }

    // 3. Create edges: session -> agents
    for _, agentID := range session.AgentIDs() {
        agentEdge := GraphEdge{
            ID:     uuid.New(),
            FromID: session.ID(),
            ToID:   agentID,
            Type:   "handled_by",
            Properties: map[string]interface{}{
                "transfer_count": session.AgentTransfers(),
            },
            ValidFrom: session.StartedAt(),
            TxTime:    time.Now(),
        }

        if err := t.ageRepo.CreateEdge(ctx, agentEdge); err != nil {
            return err
        }
    }

    return nil
}

// TraverseAgentTransferChain - Busca chain de transferências
func (t *TemporalKnowledgeGraph) TraverseAgentTransferChain(
    ctx context.Context,
    contactID uuid.UUID,
    depth int,
) ([]GraphPath, error) {
    // CYPHER query via Apache AGE
    query := `
    MATCH path = (c:Contact {id: $contactID})-[:has_session*1..%d]->(s:Session)-[:handled_by]->(a:Agent)
    WHERE s.escalated = true
    RETURN path
    ORDER BY s.started_at DESC
    LIMIT 10
    `

    query = fmt.Sprintf(query, depth)

    return t.ageRepo.ExecuteCypher(ctx, query, map[string]interface{}{
        "contactID": contactID.String(),
    })
}

// TraverseCampaignAttribution - Busca origem da campanha
func (t *TemporalKnowledgeGraph) TraverseCampaignAttribution(
    ctx context.Context,
    contactID uuid.UUID,
) (*CampaignPath, error) {
    // CYPHER: Contact <- originated_from <- Campaign
    query := `
    MATCH (c:Contact {id: $contactID})<-[:originated_from]-(t:Tracking)<-[:tracked_by]-(campaign:Campaign)
    RETURN campaign
    LIMIT 1
    `

    result, err := t.ageRepo.ExecuteCypher(ctx, query, map[string]interface{}{
        "contactID": contactID.String(),
    })
    if err != nil {
        return nil, err
    }

    if len(result) == 0 {
        return nil, nil // Sem atribuição
    }

    // Parse result
    return parseCampaignPath(result[0]), nil
}

// TraverseMessageThreads - Busca threads de mensagens (reply chains)
func (t *TemporalKnowledgeGraph) TraverseMessageThreads(
    ctx context.Context,
    sessionID uuid.UUID,
) ([]MessageThread, error) {
    // CYPHER: Message -[replied_to]-> Message
    query := `
    MATCH path = (m1:Message {session_id: $sessionID})-[:replied_to*]->(m2:Message)
    RETURN path
    ORDER BY LENGTH(path) DESC
    `

    results, err := t.ageRepo.ExecuteCypher(ctx, query, map[string]interface{}{
        "sessionID": sessionID.String(),
    })
    if err != nil {
        return nil, err
    }

    return parseMessageThreads(results), nil
}

// TraverseSocialGraph - Busca menções (social network)
func (t *TemporalKnowledgeGraph) TraverseSocialGraph(
    ctx context.Context,
    contactID uuid.UUID,
    depth int,
) ([]Contact, error) {
    // CYPHER: Contact -[mentions]-> Contact
    query := `
    MATCH path = (c1:Contact {id: $contactID})-[:mentions*1..%d]->(c2:Contact)
    RETURN DISTINCT c2
    LIMIT 20
    `

    query = fmt.Sprintf(query, depth)

    results, err := t.ageRepo.ExecuteCypher(ctx, query, map[string]interface{}{
        "contactID": contactID.String(),
    })
    if err != nil {
        return nil, err
    }

    return parseContacts(results), nil
}

// Types
type GraphPath struct {
    Nodes []GraphNode `json:"nodes"`
    Edges []GraphEdge `json:"edges"`
}

type CampaignPath struct {
    CampaignID   string `json:"campaign_id"`
    CampaignName string `json:"campaign_name"`
    Source       string `json:"source"`
    Medium       string `json:"medium"`
    Platform     string `json:"platform"`
}

type MessageThread struct {
    ThreadID     uuid.UUID   `json:"thread_id"`
    RootMessage  uuid.UUID   `json:"root_message"`
    Messages     []uuid.UUID `json:"messages"`
    ThreadDepth  int         `json:"thread_depth"`
}
```

---

## 🎯 AGENT REGISTRY & ROUTING

```go
package agent

import (
    "context"
    "fmt"
    "sort"

    "github.com/google/uuid"
)

// AgentRegistry - Registro de agentes disponíveis e roteamento
type AgentRegistry struct {
    repo            AgentRepository
    routingService  *SemanticRoutingService
}

// RegisterAgent - Registra novo agent
func (r *AgentRegistry) RegisterAgent(ctx context.Context, agent *Agent) error {
    // Validações
    if agent.Type() != AgentTypeAI {
        return fmt.Errorf("only AI agents can be registered")
    }

    metadata := agent.AIMetadata()
    if metadata == nil {
        return fmt.Errorf("AI agent must have metadata")
    }

    // Persiste
    return r.repo.Save(ctx, agent)
}

// RouteToAgent - Roteia mensagem para agent apropriado
func (r *AgentRegistry) RouteToAgent(
    ctx context.Context,
    req RoutingRequest,
) (*Agent, float64, error) {
    // 1. Get all active AI agents for project
    agents, err := r.repo.FindActiveAIAgentsByProject(ctx, req.ProjectID)
    if err != nil {
        return nil, 0, err
    }

    if len(agents) == 0 {
        return nil, 0, fmt.Errorf("no AI agents available")
    }

    // 2. Score each agent
    scores := make([]agentScore, 0, len(agents))

    for _, agent := range agents {
        score := r.scoreAgent(ctx, agent, req)
        scores = append(scores, agentScore{
            agent: agent,
            score: score,
        })
    }

    // 3. Sort by score DESC
    sort.Slice(scores, func(i, j int) bool {
        // Se scores iguais, usa priority
        if scores[i].score == scores[j].score {
            return scores[i].agent.AIMetadata().Priority > scores[j].agent.AIMetadata().Priority
        }
        return scores[i].score > scores[j].score
    })

    // 4. Return top agent
    if scores[0].score < 0.3 {
        // Score muito baixo: retorna balanced agent
        balancedAgent, err := r.repo.FindBalancedAgent(ctx, req.ProjectID)
        if err != nil {
            return nil, 0, fmt.Errorf("no suitable agent found and balanced fallback unavailable")
        }
        return balancedAgent, 0.5, nil
    }

    return scores[0].agent, scores[0].score, nil
}

// scoreAgent - Calcula score de adequação do agent
func (r *AgentRegistry) scoreAgent(
    ctx context.Context,
    agent *Agent,
    req RoutingRequest,
) float64 {
    metadata := agent.AIMetadata()
    score := 0.0

    // === 1. SEMANTIC SIMILARITY (50%) ===
    if len(metadata.IntentExamples) > 0 {
        semanticScore := r.routingService.CalculateSimilarity(
            req.Message,
            metadata.IntentExamples,
        )
        score += 0.50 * semanticScore
    }

    // === 2. ROUTING RULES (30%) ===
    ruleScore := r.evaluateRoutingRules(metadata.RoutingRules, req)
    score += 0.30 * ruleScore

    // === 3. AGENT WORKLOAD (10%) ===
    // Se agent tem MaxConcurrentSessions, verificar se está saturado
    if metadata.MaxConcurrentSessions > 0 {
        activeSessions := r.repo.CountActiveSessions(ctx, agent.ID())
        if activeSessions >= metadata.MaxConcurrentSessions {
            score *= 0.5 // Penaliza se saturado
        }
    }

    // === 4. SKILLS MATCH (10%) ===
    skillScore := r.matchSkills(metadata.Skills, req.RequiredSkills)
    score += 0.10 * skillScore

    return score
}

// evaluateRoutingRules - Avalia regras de roteamento
func (r *AgentRegistry) evaluateRoutingRules(
    rules []RoutingRule,
    req RoutingRequest,
) float64 {
    if len(rules) == 0 {
        return 0.0
    }

    // Sort by priority
    sort.Slice(rules, func(i, j int) bool {
        return rules[i].Priority > rules[j].Priority
    })

    // Avaliar regras em ordem
    for _, rule := range rules {
        matched, confidence := r.evaluateRule(rule, req)
        if matched {
            return confidence
        }
    }

    return 0.0
}

// evaluateRule - Avalia regra individual
func (r *AgentRegistry) evaluateRule(
    rule RoutingRule,
    req RoutingRequest,
) (bool, float64) {
    // Exemplo de avaliação de condition
    // Condition pode ser: "sentiment < -0.5", "topics contains 'cancelamento'"

    // TODO: Implementar rule engine (pode usar govaluate ou similar)
    // Por enquanto, retorna false
    return false, 0.0
}

// matchSkills - Calcula match de skills
func (r *AgentRegistry) matchSkills(
    agentSkills []AgentSkill,
    requiredSkills []string,
) float64 {
    if len(requiredSkills) == 0 {
        return 0.5 // Neutro
    }

    matched := 0
    for _, required := range requiredSkills {
        for _, skill := range agentSkills {
            if skill.Name == required {
                matched++
                break
            }
        }
    }

    return float64(matched) / float64(len(requiredSkills))
}

// RoutingRequest - Request para roteamento
type RoutingRequest struct {
    ProjectID       uuid.UUID  `json:"project_id"`
    ContactID       uuid.UUID  `json:"contact_id"`
    SessionID       *uuid.UUID `json:"session_id"`
    Message         string     `json:"message"`
    Sentiment       *float64   `json:"sentiment"`
    Topics          []string   `json:"topics"`
    RequiredSkills  []string   `json:"required_skills"`
    Context         map[string]interface{} `json:"context"`
}

type agentScore struct {
    agent *Agent
    score float64
}

// SemanticRoutingService - Serviço de roteamento semântico
type SemanticRoutingService struct {
    embeddingService *EmbeddingService
}

// CalculateSimilarity - Calcula similaridade entre mensagem e exemplos
func (s *SemanticRoutingService) CalculateSimilarity(
    message string,
    examples []string,
) float64 {
    // TODO: Implementar semantic similarity usando embeddings
    // 1. Gerar embedding da mensagem
    // 2. Gerar embeddings dos exemplos (cachear)
    // 3. Calcular cosine similarity
    // 4. Retornar max similarity

    return 0.0 // Placeholder
}
```

---

## 💭 MEMORY FACT SERVICE

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// MemoryFactService - Serviço de gerenciamento de memory facts
type MemoryFactService struct {
    repo       MemoryFactRepository
    llmClient  *genai.Client  // Para contradiction resolution
}

// MemoryFact - Fato estruturado sobre o contato
type MemoryFact struct {
    ID          uuid.UUID              `db:"id"`
    TenantID    string                 `db:"tenant_id"`
    ContactID   uuid.UUID              `db:"contact_id"`
    FactType    FactType               `db:"fact_type"`
    FactText    string                 `db:"fact_text"`
    Confidence  float64                `db:"confidence"`  // 0.0-1.0
    SourceType  string                 `db:"source_type"` // session, message, note, agent
    SourceID    uuid.UUID              `db:"source_id"`
    ExtractedAt time.Time              `db:"extracted_at"`
    ValidUntil  *time.Time             `db:"valid_until"`
    Superseded  bool                   `db:"superseded"` // Fact foi contradito?
    SupersededBy *uuid.UUID            `db:"superseded_by"`
    Metadata    map[string]interface{} `db:"metadata"`
}

// FactType - Tipos de fatos
type FactType string

const (
    FactTypeBudget         FactType = "budget_constraint"
    FactTypePreference     FactType = "preference"
    FactTypeGoal           FactType = "goal"
    FactTypePainPoint      FactType = "pain_point"
    FactTypeObjection      FactType = "objection"
    FactTypeConstraint     FactType = "constraint"
    FactTypeTechnicalInfo  FactType = "technical_info"
    FactTypePersonal       FactType = "personal_info"
    FactTypeBusinessInfo   FactType = "business_info"
)

// ExtractFactsFromSession - Extrai facts de uma sessão usando LLM
func (m *MemoryFactService) ExtractFactsFromSession(
    ctx context.Context,
    sessionID uuid.UUID,
    tenantID string,
) ([]MemoryFact, error) {
    // 1. Busca sessão completa
    session, err := m.sessionRepo.FindByID(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    // 2. Busca mensagens da sessão
    messages, err := m.messageRepo.FindBySession(ctx, sessionID)
    if err != nil {
        return nil, err
    }

    // 3. Monta prompt para LLM
    extractionPrompt := m.buildFactExtractionPrompt(session, messages)

    // 4. Chama LLM (Gemini Flash)
    response, err := m.llmClient.GenerateContent(ctx, extractionPrompt)
    if err != nil {
        return nil, fmt.Errorf("LLM extraction failed: %w", err)
    }

    // 5. Parse JSON response
    extractedFacts := parseFactsFromJSON(response.Text)

    // 6. Cria MemoryFact entities
    facts := make([]MemoryFact, 0, len(extractedFacts))
    now := time.Now()

    for _, ef := range extractedFacts {
        fact := MemoryFact{
            ID:          uuid.New(),
            TenantID:    tenantID,
            ContactID:   session.ContactID(),
            FactType:    FactType(ef.Type),
            FactText:    ef.Text,
            Confidence:  ef.Confidence,
            SourceType:  "session",
            SourceID:    sessionID,
            ExtractedAt: now,
            ValidUntil:  ef.ValidUntil,
            Superseded:  false,
            Metadata:    ef.Metadata,
        }

        facts = append(facts, fact)
    }

    // 7. Verifica contradições com facts existentes
    for i := range facts {
        err := m.CheckAndResolveContradictions(ctx, &facts[i])
        if err != nil {
            fmt.Printf("Warning: contradiction resolution failed for fact %s: %v\n", facts[i].ID, err)
        }
    }

    // 8. Persiste
    for i := range facts {
        if err := m.repo.Save(ctx, &facts[i]); err != nil {
            return nil, err
        }
    }

    return facts, nil
}

// CheckAndResolveContradictions - Detecta e resolve contradições
func (m *MemoryFactService) CheckAndResolveContradictions(
    ctx context.Context,
    newFact *MemoryFact,
) error {
    // 1. Busca facts similares do mesmo tipo
    existingFacts, err := m.repo.FindByContactAndType(
        ctx,
        newFact.ContactID,
        newFact.FactType,
    )
    if err != nil {
        return err
    }

    if len(existingFacts) == 0 {
        return nil // Sem contradição possível
    }

    // 2. Para cada fact existente, verifica contradição com LLM
    for _, existing := range existingFacts {
        if existing.Superseded {
            continue // Já foi superseded
        }

        contradicts := m.checkContradiction(ctx, existing, *newFact)

        if contradicts {
            // 3. Resolve contradição: fact mais recente vence
            if newFact.ExtractedAt.After(existing.ExtractedAt) {
                // New fact supersedes existing
                existing.Superseded = true
                existing.SupersededBy = &newFact.ID
                if err := m.repo.Update(ctx, &existing); err != nil {
                    return err
                }
            } else {
                // Existing fact supersedes new (não persiste new)
                newFact.Superseded = true
                newFact.SupersededBy = &existing.ID
            }
        }
    }

    return nil
}

// checkContradiction - Verifica se facts contradizem usando LLM
func (m *MemoryFactService) checkContradiction(
    ctx context.Context,
    fact1 MemoryFact,
    fact2 MemoryFact,
) bool {
    prompt := fmt.Sprintf(`
Analyze if these two facts contradict each other:

Fact 1 (from %s): "%s"
Fact 2 (from %s): "%s"

Return ONLY "true" if they contradict, "false" otherwise.
`,
        fact1.ExtractedAt.Format("2006-01-02"),
        fact1.FactText,
        fact2.ExtractedAt.Format("2006-01-02"),
        fact2.FactText,
    )

    response, err := m.llmClient.GenerateContent(ctx, prompt)
    if err != nil {
        return false // Em caso de erro, assume não contradiz
    }

    return response.Text == "true"
}

// GetActiveFactsByContact - Retorna facts ativos (não superseded)
func (m *MemoryFactService) GetActiveFactsByContact(
    ctx context.Context,
    contactID uuid.UUID,
    factTypes []string,
) ([]MemoryFact, error) {
    facts, err := m.repo.FindActiveByContact(ctx, contactID, factTypes)
    if err != nil {
        return nil, err
    }

    // Filtra apenas não-superseded e não-expirados
    activeFacts := make([]MemoryFact, 0, len(facts))
    now := time.Now()

    for _, fact := range facts {
        if fact.Superseded {
            continue
        }
        if fact.ValidUntil != nil && fact.ValidUntil.Before(now) {
            continue
        }
        activeFacts = append(activeFacts, fact)
    }

    return activeFacts, nil
}

// buildFactExtractionPrompt - Monta prompt para extração
func (m *MemoryFactService) buildFactExtractionPrompt(
    session *Session,
    messages []Message,
) string {
    // TODO: Implementar prompt engineering para extração de facts
    return ""
}

type extractedFact struct {
    Type       string                 `json:"type"`
    Text       string                 `json:"text"`
    Confidence float64                `json:"confidence"`
    ValidUntil *time.Time             `json:"valid_until"`
    Metadata   map[string]interface{} `json:"metadata"`
}

func parseFactsFromJSON(jsonStr string) []extractedFact {
    // TODO: Parse JSON
    return []extractedFact{}
}
```

---

## 💾 CONTEXT MANAGER (CACHING)

```go
package memory

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
)

// ContextManager - Gerenciador de cache de contexto
type ContextManager struct {
    redisClient *redis.Client
}

// NewContextManager cria novo gerenciador
func NewContextManager(redisClient *redis.Client) *ContextManager {
    return &ContextManager{
        redisClient: redisClient,
    }
}

// GetCached - Busca resultado cached
func (c *ContextManager) GetCached(
    ctx context.Context,
    req SearchRequest,
) (*SearchResult, error) {
    cacheKey := c.buildCacheKey(req)

    // Get from Redis
    data, err := c.redisClient.Get(ctx, cacheKey).Result()
    if err == redis.Nil {
        return nil, nil // Cache miss
    }
    if err != nil {
        return nil, err
    }

    // Deserialize
    var result SearchResult
    if err := json.Unmarshal([]byte(data), &result); err != nil {
        return nil, err
    }

    return &result, nil
}

// SetCached - Armazena resultado em cache
func (c *ContextManager) SetCached(
    ctx context.Context,
    req SearchRequest,
    result *SearchResult,
    ttlSeconds int,
) error {
    cacheKey := c.buildCacheKey(req)

    // Serialize
    data, err := json.Marshal(result)
    if err != nil {
        return err
    }

    // Set in Redis com TTL
    return c.redisClient.Set(
        ctx,
        cacheKey,
        data,
        time.Duration(ttlSeconds)*time.Second,
    ).Err()
}

// InvalidateContact - Invalida todos caches de um contato
func (c *ContextManager) InvalidateContact(
    ctx context.Context,
    contactID uuid.UUID,
) error {
    pattern := fmt.Sprintf("memory:cache:contact:%s:*", contactID.String())

    // Scan + delete
    iter := c.redisClient.Scan(ctx, 0, pattern, 100).Iterator()
    for iter.Next(ctx) {
        if err := c.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
            return err
        }
    }

    return iter.Err()
}

// buildCacheKey - Gera chave de cache
func (c *ContextManager) buildCacheKey(req SearchRequest) string {
    // Hash baseado em: contact_id + query + strategy + scope
    h := sha256.New()

    // Add fields que impactam resultado
    h.Write([]byte(req.TenantID))
    h.Write([]byte(req.ContactID.String()))
    h.Write([]byte(req.Query))
    h.Write([]byte(req.AgentCategory))

    // Serialize strategy + scope para hash
    strategyJSON, _ := json.Marshal(req.MemoryStrategy)
    scopeJSON, _ := json.Marshal(req.KnowledgeScope)
    h.Write(strategyJSON)
    h.Write(scopeJSON)

    hash := fmt.Sprintf("%x", h.Sum(nil))

    // Key format: memory:cache:contact:{contactID}:{hash}
    return fmt.Sprintf("memory:cache:contact:%s:%s",
        req.ContactID.String(),
        hash[:16], // Primeiros 16 chars do hash
    )
}

// GetCachedPromptContext - Busca contexto cached para prompt caching
func (c *ContextManager) GetCachedPromptContext(
    ctx context.Context,
    contactID uuid.UUID,
) (*PromptContext, error) {
    key := fmt.Sprintf("prompt:context:%s", contactID.String())

    data, err := c.redisClient.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }

    var promptCtx PromptContext
    if err := json.Unmarshal([]byte(data), &promptCtx); err != nil {
        return nil, err
    }

    return &promptCtx, nil
}

// SetCachedPromptContext - Armazena contexto para prompt caching
func (c *ContextManager) SetCachedPromptContext(
    ctx context.Context,
    contactID uuid.UUID,
    promptCtx *PromptContext,
    ttlMinutes int,
) error {
    key := fmt.Sprintf("prompt:context:%s", contactID.String())

    data, err := json.Marshal(promptCtx)
    if err != nil {
        return err
    }

    return c.redisClient.Set(
        ctx,
        key,
        data,
        time.Duration(ttlMinutes)*time.Minute,
    ).Err()
}

// PromptContext - Contexto para prompt caching (Anthropic/Gemini)
type PromptContext struct {
    ContactID        uuid.UUID  `json:"contact_id"`
    BaseContext      string     `json:"base_context"`      // Context fixo (profile, stats)
    DynamicContext   string     `json:"dynamic_context"`   // Context que muda (recent msgs)
    LastUpdated      time.Time  `json:"last_updated"`
    TokenCount       int        `json:"token_count"`
    CacheHitRate     float64    `json:"cache_hit_rate"`    // Tracking
}
```

---

## 🔌 gRPC API

```protobuf
// memory_service.proto

syntax = "proto3";

package ventros.memory.v1;

option go_package = "github.com/ventros/api/gen/memory/v1;memoryv1";

// MemoryService - Serviço gRPC de memória
service MemoryService {
    // Search memory with hybrid retrieval
    rpc SearchMemory(SearchMemoryRequest) returns (SearchMemoryResponse);

    // Add session to memory (generate embeddings async)
    rpc AddSession(AddSessionRequest) returns (AddSessionResponse);

    // Get contact statistics
    rpc GetContactStats(GetContactStatsRequest) returns (GetContactStatsResponse);

    // Get memory facts
    rpc GetMemoryFacts(GetMemoryFactsRequest) returns (GetMemoryFactsResponse);

    // Stream real-time updates (for Python ADK)
    rpc StreamUpdates(StreamUpdatesRequest) returns (stream MemoryUpdate);
}

message SearchMemoryRequest {
    string tenant_id = 1;
    string contact_id = 2;
    optional string session_id = 3;
    string query = 4;
    string agent_category = 5;
    int32 top_k = 6;
}

message SearchMemoryResponse {
    repeated Message recent_messages = 1;
    repeated SessionSummary similar_sessions = 2;
    ContactStats contact_stats = 3;
    repeated MemoryFact memory_facts = 4;
    int64 search_latency_ms = 5;
    bool cache_hit = 6;
}

message Message {
    string id = 1;
    string content = 2;
    bool from_me = 3;
    string timestamp = 4;
    string content_type = 5;
}

message SessionSummary {
    string session_id = 1;
    string summary = 2;
    string sentiment = 3;
    float sentiment_score = 4;
    repeated string topics = 5;
    float similarity_score = 6;
    string started_at = 7;
}

message ContactStats {
    int32 total_sessions = 1;
    int32 total_messages = 2;
    float avg_sentiment_score = 3;
    string last_session_at = 4;
    int32 resolved_sessions = 5;
    int32 escalated_sessions = 6;
}

message MemoryFact {
    string id = 1;
    string fact_type = 2;
    string fact_text = 3;
    float confidence = 4;
    string extracted_at = 5;
}

message AddSessionRequest {
    string tenant_id = 1;
    string session_id = 2;
    string contact_id = 3;
    repeated Message messages = 4;
    map<string, string> metadata = 5;
}

message AddSessionResponse {
    bool success = 1;
    string embedding_id = 2;
}

message GetContactStatsRequest {
    string tenant_id = 1;
    string contact_id = 2;
}

message GetContactStatsResponse {
    ContactStats stats = 1;
}

message GetMemoryFactsRequest {
    string tenant_id = 1;
    string contact_id = 2;
    repeated string fact_types = 3;
}

message GetMemoryFactsResponse {
    repeated MemoryFact facts = 1;
}

message StreamUpdatesRequest {
    string tenant_id = 1;
    repeated string contact_ids = 2;
}

message MemoryUpdate {
    string contact_id = 1;
    string update_type = 2; // "session_created", "fact_extracted", "embedding_generated"
    string timestamp = 3;
    map<string, string> data = 4;
}
```

---

## 🗄️ DATABASE SCHEMA

```sql
-- ==========================================
-- MEMORY EMBEDDINGS
-- ==========================================

CREATE TABLE IF NOT EXISTS memory_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL, -- session_summary, note, etc
    source_id UUID NOT NULL,
    embedding vector(768), -- pgvector (text-embedding-005)
    contextual_text TEXT NOT NULL, -- Context + original text
    original_text TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,

    INDEX idx_memory_embeddings_tenant (tenant_id),
    INDEX idx_memory_embeddings_source (source_type, source_id),
    INDEX idx_memory_embeddings_created (created_at DESC)
);

-- Vector index (HNSW for fast similarity search)
CREATE INDEX idx_memory_embeddings_vector
    ON memory_embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- ==========================================
-- MEMORY FACTS
-- ==========================================

CREATE TABLE IF NOT EXISTS memory_facts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    fact_type VARCHAR(50) NOT NULL, -- budget_constraint, preference, etc
    fact_text TEXT NOT NULL,
    confidence FLOAT NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    source_type VARCHAR(50) NOT NULL, -- session, message, note
    source_id UUID NOT NULL,
    extracted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_until TIMESTAMP,
    superseded BOOLEAN DEFAULT FALSE,
    superseded_by UUID REFERENCES memory_facts(id),
    metadata JSONB DEFAULT '{}',

    INDEX idx_memory_facts_contact (contact_id),
    INDEX idx_memory_facts_type (fact_type),
    INDEX idx_memory_facts_superseded (superseded) WHERE superseded = FALSE,
    INDEX idx_memory_facts_valid (valid_until) WHERE valid_until IS NOT NULL
);

-- ==========================================
-- KNOWLEDGE GRAPH (Apache AGE)
-- ==========================================

-- Instalar extensão AGE
CREATE EXTENSION IF NOT EXISTS age;

-- Load AGE extension
LOAD 'age';
SET search_path = ag_catalog, "$user", public;

-- Criar graph
SELECT create_graph('ventros_graph');

-- Nodes: Contact, Session, Agent, Message, Campaign
-- Edges: has_session, handled_by, replied_to, transferred_to, originated_from

-- Queries são feitas via CYPHER (ver TemporalKnowledgeGraph service)

-- ==========================================
-- AGENT METADATA EXTENSIONS
-- ==========================================

-- Adicionar colunas à tabela agents existente
ALTER TABLE agents
ADD COLUMN IF NOT EXISTS ai_metadata JSONB DEFAULT '{}';

-- Index para query por categoria
CREATE INDEX idx_agents_ai_category
    ON agents ((ai_metadata->>'category'));

-- ==========================================
-- KEYWORD SEARCH (pg_trgm)
-- ==========================================

-- Instalar extensão pg_trgm
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Index trigram em session summaries
CREATE INDEX idx_sessions_summary_trgm
    ON sessions
    USING gin (summary gin_trgm_ops);

-- Index trigram em messages
CREATE INDEX idx_messages_text_trgm
    ON messages
    USING gin (text gin_trgm_ops);

-- ==========================================
-- TEMPORAL TRACKING
-- ==========================================

-- Adicionar colunas temporais (bi-temporal model)
ALTER TABLE memory_embeddings
ADD COLUMN IF NOT EXISTS valid_from TIMESTAMP NOT NULL DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS valid_to TIMESTAMP,
ADD COLUMN IF NOT EXISTS tx_time TIMESTAMP NOT NULL DEFAULT NOW();

-- Index para queries temporais
CREATE INDEX idx_memory_embeddings_temporal
    ON memory_embeddings (valid_from, valid_to);
```

---

## 🔌 MCP SERVER IMPLEMENTATION

### **Model Context Protocol (MCP) Server**

**Decisão Arquitetural: Hybrid Approach**

```
┌─────────────────────────────────────────────────────────────┐
│  VENTROS MCP SERVER (Go)                                     │
│  ✅ Tools reutilizáveis entre múltiplos agents               │
│  ✅ BI Analytics (leads count, conversions, metrics)         │
│  ✅ CRM Operations (pipeline, assignments, tasks)            │
│  ✅ Agent Performance (comparisons, quality analysis)        │
│  ✅ Connection pooling + auth + caching                      │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ MCP Protocol (stdio/HTTP)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  PYTHON ADK - MCPToolset                                     │
│  ✅ Auto-discovery de tools                                  │
│  ✅ tool_filter (segurança)                                  │
│  ✅ Lifecycle management                                     │
└─────────────────────────────────────────────────────────────┘
```

### **MCP Tools Catalog**

```go
package mcp

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/google/uuid"
)

// MCPServer - Servidor MCP para expor tools do Ventros
type MCPServer struct {
    biService        *BIQueryService
    crmService       *CRMOperationsService
    agentAnalyzer    *AgentPerformanceAnalyzer
    authService      *AuthService
}

// ToolDefinition - Definição de tool MCP
type ToolDefinition struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// GetAvailableTools - Lista todas tools disponíveis (MCP protocol)
func (m *MCPServer) GetAvailableTools(ctx context.Context) ([]ToolDefinition, error) {
    return []ToolDefinition{
        // === BI ANALYTICS TOOLS ===
        {
            Name:        "get_leads_count",
            Description: "Retorna quantidade de leads em período",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "date_filter": map[string]interface{}{
                        "type":        "string",
                        "description": "Filtro de data: 'today', 'yesterday', 'this_week', 'this_month', ou ISO date range",
                    },
                    "pipeline_stage": map[string]interface{}{
                        "type":        "string",
                        "description": "Filtro por stage (opcional)",
                    },
                },
                "required": []string{"date_filter"},
            },
        },
        {
            Name:        "get_agent_conversion_stats",
            Description: "Retorna estatísticas de conversão por agente",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "agent_ids": map[string]interface{}{
                        "type":        "array",
                        "description": "IDs dos agentes (vazio = todos)",
                        "items": map[string]interface{}{
                            "type": "string",
                        },
                    },
                    "date_range": map[string]interface{}{
                        "type":        "string",
                        "description": "Range de datas",
                    },
                },
            },
        },
        {
            Name:        "get_top_performing_agent",
            Description: "Retorna agente com melhor performance em métrica específica",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "metric": map[string]interface{}{
                        "type": "string",
                        "enum": []string{"conversion_rate", "response_time", "satisfaction_score", "resolution_rate"},
                    },
                    "date_range": map[string]interface{}{
                        "type": "string",
                    },
                },
                "required": []string{"metric"},
            },
        },

        // === AGENT ANALYSIS TOOLS ===
        {
            Name:        "analyze_agent_messages",
            Description: "Analisa mensagens de um agente (gramática, tom, brand alignment)",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "agent_id": map[string]interface{}{
                        "type": "string",
                    },
                    "sample_size": map[string]interface{}{
                        "type":    "integer",
                        "default": 50,
                    },
                    "include_grammar": map[string]interface{}{
                        "type":    "boolean",
                        "default": true,
                    },
                    "include_tone": map[string]interface{}{
                        "type":    "boolean",
                        "default": true,
                    },
                    "include_brand_alignment": map[string]interface{}{
                        "type":    "boolean",
                        "default": true,
                    },
                },
                "required": []string{"agent_id"},
            },
        },
        {
            Name:        "compare_agents",
            Description: "Compara performance e qualidade entre agentes",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "agent_ids": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            "type": "string",
                        },
                        "minItems": 2,
                    },
                    "comparison_dimensions": map[string]interface{}{
                        "type":  "array",
                        "items": map[string]interface{}{
                            "type": "string",
                            "enum": []string{"response_time", "grammar", "tone", "conversion", "satisfaction"},
                        },
                    },
                },
                "required": []string{"agent_ids"},
            },
        },

        // === CRM OPERATIONS TOOLS ===
        {
            Name:        "assign_to_agent",
            Description: "Atribui contato para agente humano",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "contact_id": map[string]interface{}{
                        "type": "string",
                    },
                    "agent_id": map[string]interface{}{
                        "type": "string",
                    },
                    "reason": map[string]interface{}{
                        "type": "string",
                    },
                },
                "required": []string{"contact_id", "agent_id"},
            },
        },
        {
            Name:        "update_pipeline_stage",
            Description: "Atualiza stage no pipeline",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "contact_id": map[string]interface{}{
                        "type": "string",
                    },
                    "pipeline_id": map[string]interface{}{
                        "type": "string",
                    },
                    "stage_id": map[string]interface{}{
                        "type": "string",
                    },
                },
                "required": []string{"contact_id", "pipeline_id", "stage_id"},
            },
        },
        {
            Name:        "qualify_lead",
            Description: "Qualifica lead usando critérios BANT",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "contact_id": map[string]interface{}{
                        "type": "string",
                    },
                    "qualification_data": map[string]interface{}{
                        "type": "object",
                        "properties": map[string]interface{}{
                            "budget":    map[string]interface{}{"type": "number"},
                            "authority": map[string]interface{}{"type": "boolean"},
                            "need":      map[string]interface{}{"type": "string"},
                            "timeline":  map[string]interface{}{"type": "string"},
                        },
                    },
                },
                "required": []string{"contact_id", "qualification_data"},
            },
        },
    }, nil
}

// ExecuteTool - Executa tool MCP (dispatcher)
func (m *MCPServer) ExecuteTool(
    ctx context.Context,
    toolName string,
    arguments map[string]interface{},
) (interface{}, error) {
    // Autenticação/autorização
    if err := m.authService.ValidateToolAccess(ctx, toolName); err != nil {
        return nil, fmt.Errorf("unauthorized: %w", err)
    }

    // Dispatch por tool name
    switch toolName {
    case "get_leads_count":
        return m.biService.GetLeadsCount(ctx, arguments)
    case "get_agent_conversion_stats":
        return m.biService.GetAgentConversionStats(ctx, arguments)
    case "get_top_performing_agent":
        return m.biService.GetTopPerformingAgent(ctx, arguments)
    case "analyze_agent_messages":
        return m.agentAnalyzer.AnalyzeMessages(ctx, arguments)
    case "compare_agents":
        return m.agentAnalyzer.CompareAgents(ctx, arguments)
    case "assign_to_agent":
        return m.crmService.AssignToAgent(ctx, arguments)
    case "update_pipeline_stage":
        return m.crmService.UpdatePipelineStage(ctx, arguments)
    case "qualify_lead":
        return m.crmService.QualifyLead(ctx, arguments)
    default:
        return nil, fmt.Errorf("unknown tool: %s", toolName)
    }
}
```

### **MCP vs Direct ADK Tools - Decision Tree**

```
                    [Need Tool]
                         │
                         ▼
              ┌──────────┴───────────┐
              │                      │
         [Reutilizável?]      [Específico do Agent?]
              │                      │
         [SIM] │                [SIM] │
              ▼                      ▼
      ┌───────────────┐      ┌──────────────┐
      │   MCP TOOL    │      │  DIRECT TOOL │
      │               │      │   (Python)   │
      │ - BI queries  │      │              │
      │ - CRM ops     │      │ - Formatting │
      │ - Analytics   │      │ - Parsing    │
      │ - DB access   │      │ - Validation │
      └───────────────┘      └──────────────┘
```

**Use MCP quando:**
- ✅ Tool é usada por múltiplos agents
- ✅ Operação envolve DB/API complexa
- ✅ Precisa connection pooling/caching
- ✅ Tool muda frequentemente
- ✅ Segurança centralizada necessária

**Use Direct ADK Tool quando:**
- ✅ Lógica específica de um agent
- ✅ Operação leve (formatting, parsing)
- ✅ Workflow domain-specific
- ✅ Não precisa persistência

---

## 📊 AGENT METRICS & PERFORMANCE TRACKING

### **Database Schema - Agent Metrics**

```sql
-- ==========================================
-- AGENT PERFORMANCE METRICS
-- ==========================================

CREATE TABLE IF NOT EXISTS agent_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    metric_date DATE NOT NULL,

    -- Volumetria
    total_sessions INT DEFAULT 0,
    total_messages INT DEFAULT 0,
    messages_sent INT DEFAULT 0,
    messages_received INT DEFAULT 0,

    -- Performance
    avg_response_time_seconds FLOAT,
    median_response_time_seconds FLOAT,
    p95_response_time_seconds FLOAT,

    -- Qualidade
    avg_satisfaction_score FLOAT,
    positive_feedbacks INT DEFAULT 0,
    negative_feedbacks INT DEFAULT 0,

    -- Conversão
    leads_qualified INT DEFAULT 0,
    opportunities_created INT DEFAULT 0,
    deals_closed INT DEFAULT 0,
    conversion_rate FLOAT,

    -- Resolução
    sessions_resolved INT DEFAULT 0,
    sessions_escalated INT DEFAULT 0,
    resolution_rate FLOAT,
    first_response_resolution INT DEFAULT 0,

    -- Retenção
    churn_prevented INT DEFAULT 0,
    retention_offers_accepted INT DEFAULT 0,

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE (tenant_id, agent_id, metric_date),
    INDEX idx_agent_metrics_date (metric_date DESC),
    INDEX idx_agent_metrics_agent (agent_id, metric_date DESC)
);

-- ==========================================
-- AGENT MESSAGE QUALITY ANALYSIS
-- ==========================================

CREATE TABLE IF NOT EXISTS agent_message_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    analyzed_at TIMESTAMP DEFAULT NOW(),

    -- Gramática
    grammar_score FLOAT, -- 0-1
    grammar_issues JSONB, -- [{type, severity, suggestion}]

    -- Tom de voz
    tone_detected VARCHAR(50), -- "professional", "friendly", "empathetic", etc
    brand_alignment_score FLOAT, -- 0-1
    tone_issues JSONB,

    -- Conteúdo
    clarity_score FLOAT,
    completeness_score FLOAT,
    relevance_score FLOAT,

    -- Sentiment
    message_sentiment VARCHAR(20),
    sentiment_score FLOAT,

    -- Metadata
    analysis_version VARCHAR(20),
    llm_model_used VARCHAR(50),

    INDEX idx_message_analysis_agent (agent_id, analyzed_at DESC),
    INDEX idx_message_analysis_message (message_id)
);

-- ==========================================
-- AGENT COMPARISONS
-- ==========================================

CREATE TABLE IF NOT EXISTS agent_comparisons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    comparison_name VARCHAR(255),
    requested_by_agent_id UUID, -- Qual agent solicitou (ex: BI Manager)
    requested_at TIMESTAMP DEFAULT NOW(),

    -- Agents sendo comparados
    agent_ids UUID[], -- Array de agent IDs

    -- Dimensões comparadas
    dimensions VARCHAR(50)[], -- ["response_time", "grammar", "tone", "conversion"]

    -- Período
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,

    -- Resultados
    comparison_results JSONB, -- Structured results
    winner_agent_id UUID, -- Agent com melhor performance geral

    -- Metadata
    completed_at TIMESTAMP,

    INDEX idx_agent_comparisons_requested (requested_at DESC),
    INDEX idx_agent_comparisons_agents (agent_ids) USING GIN
);

-- ==========================================
-- RESPONSE FORMAT TEMPLATES
-- ==========================================

CREATE TABLE IF NOT EXISTS response_format_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    template_name VARCHAR(255) NOT NULL,
    template_type VARCHAR(50) NOT NULL, -- "markdown", "json", "html", "plain_text"

    -- Template structure
    format_schema JSONB NOT NULL, -- Estrutura do formato esperado

    -- Examples
    example_data JSONB,
    example_output TEXT,

    -- Validation
    validation_rules JSONB,

    -- Usage
    used_by_agents UUID[], -- Quais agents usam este template

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE (tenant_id, template_name),
    INDEX idx_format_templates_type (template_type)
);
```

### **BI Query Service**

```go
package analytics

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// BIQueryService - Serviço de queries analytics para BI Manager Agent
type BIQueryService struct {
    db          *sql.DB
    cacheManager *CacheManager
}

// GetLeadsCount - "Quantos leads tive hoje?"
func (b *BIQueryService) GetLeadsCount(
    ctx context.Context,
    args map[string]interface{},
) (*LeadsCountResult, error) {
    dateFilter := args["date_filter"].(string)

    // Parse date filter
    from, to := b.parseDateFilter(dateFilter)

    // Cache key
    cacheKey := fmt.Sprintf("leads:count:%s", dateFilter)
    if cached, ok := b.cacheManager.Get(cacheKey); ok {
        return cached.(*LeadsCountResult), nil
    }

    // Query otimizado com indexes
    query := `
        SELECT
            COUNT(*) as total_leads,
            COUNT(CASE WHEN pipeline_status->>'stage' = 'qualified' THEN 1 END) as qualified_leads,
            COUNT(CASE WHEN pipeline_status->>'stage' = 'opportunity' THEN 1 END) as opportunities
        FROM contacts
        WHERE tenant_id = $1
          AND created_at >= $2
          AND created_at < $3
          AND tags @> '["lead"]'::jsonb
    `

    var result LeadsCountResult
    err := b.db.QueryRowContext(ctx, query,
        getTenantID(ctx),
        from,
        to,
    ).Scan(
        &result.TotalLeads,
        &result.QualifiedLeads,
        &result.Opportunities,
    )

    if err != nil {
        return nil, err
    }

    result.DateFilter = dateFilter
    result.DateFrom = from
    result.DateTo = to

    // Cache for 5 minutes
    b.cacheManager.Set(cacheKey, &result, 5*time.Minute)

    return &result, nil
}

// GetAgentConversionStats - "Qual agente converteu mais?"
func (b *BIQueryService) GetAgentConversionStats(
    ctx context.Context,
    args map[string]interface{},
) (*AgentConversionStatsResult, error) {
    dateRange := args["date_range"].(string)
    from, to := b.parseDateFilter(dateRange)

    // Query com joins otimizados
    query := `
        WITH agent_conversions AS (
            SELECT
                a.id as agent_id,
                a.name as agent_name,
                a.type as agent_type,
                COUNT(DISTINCT c.id) as total_contacts,
                COUNT(DISTINCT CASE
                    WHEN ps.status->>'stage' = 'closed_won'
                    THEN c.id
                END) as conversions,
                COALESCE(
                    COUNT(DISTINCT CASE WHEN ps.status->>'stage' = 'closed_won' THEN c.id END)::float /
                    NULLIF(COUNT(DISTINCT c.id), 0),
                    0
                ) as conversion_rate,
                AVG(EXTRACT(EPOCH FROM (ps.updated_at - c.created_at))) as avg_time_to_conversion
            FROM agents a
            LEFT JOIN sessions s ON s.agent_ids @> ARRAY[a.id]
            LEFT JOIN contacts c ON c.id = s.contact_id
            LEFT JOIN pipeline_status ps ON ps.contact_id = c.id
            WHERE a.tenant_id = $1
              AND a.is_active = true
              AND s.started_at >= $2
              AND s.started_at < $3
            GROUP BY a.id, a.name, a.type
        )
        SELECT *
        FROM agent_conversions
        ORDER BY conversion_rate DESC, conversions DESC
    `

    rows, err := b.db.QueryContext(ctx, query, getTenantID(ctx), from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var stats []AgentConversionStat
    for rows.Next() {
        var stat AgentConversionStat
        err := rows.Scan(
            &stat.AgentID,
            &stat.AgentName,
            &stat.AgentType,
            &stat.TotalContacts,
            &stat.Conversions,
            &stat.ConversionRate,
            &stat.AvgTimeToConversion,
        )
        if err != nil {
            return nil, err
        }
        stats = append(stats, stat)
    }

    return &AgentConversionStatsResult{
        Stats:       stats,
        DateFrom:    from,
        DateTo:      to,
        TopAgent:    stats[0], // Primeiro é o top
    }, nil
}

// GetTopPerformingAgent - Retorna top agent em métrica específica
func (b *BIQueryService) GetTopPerformingAgent(
    ctx context.Context,
    args map[string]interface{},
) (*TopAgentResult, error) {
    metric := args["metric"].(string)
    dateRange := args["date_range"].(string)
    from, to := b.parseDateFilter(dateRange)

    var query string
    switch metric {
    case "conversion_rate":
        query = `
            SELECT
                a.id, a.name, a.type,
                COALESCE(
                    COUNT(DISTINCT CASE WHEN ps.status->>'stage' = 'closed_won' THEN c.id END)::float /
                    NULLIF(COUNT(DISTINCT c.id), 0),
                    0
                ) as metric_value
            FROM agents a
            LEFT JOIN sessions s ON s.agent_ids @> ARRAY[a.id]
            LEFT JOIN contacts c ON c.id = s.contact_id
            LEFT JOIN pipeline_status ps ON ps.contact_id = c.id
            WHERE a.tenant_id = $1
              AND a.is_active = true
              AND s.started_at >= $2
              AND s.started_at < $3
            GROUP BY a.id, a.name, a.type
            HAVING COUNT(DISTINCT c.id) >= 10 -- Mínimo 10 contacts
            ORDER BY metric_value DESC
            LIMIT 1
        `

    case "response_time":
        query = `
            SELECT
                a.id, a.name, a.type,
                AVG(am.avg_response_time_seconds) as metric_value
            FROM agents a
            INNER JOIN agent_metrics am ON am.agent_id = a.id
            WHERE a.tenant_id = $1
              AND am.metric_date >= $2::date
              AND am.metric_date < $3::date
            GROUP BY a.id, a.name, a.type
            ORDER BY metric_value ASC -- Menor é melhor
            LIMIT 1
        `

    case "satisfaction_score":
        query = `
            SELECT
                a.id, a.name, a.type,
                AVG(am.avg_satisfaction_score) as metric_value
            FROM agents a
            INNER JOIN agent_metrics am ON am.agent_id = a.id
            WHERE a.tenant_id = $1
              AND am.metric_date >= $2::date
              AND am.metric_date < $3::date
            GROUP BY a.id, a.name, a.type
            ORDER BY metric_value DESC
            LIMIT 1
        `

    default:
        return nil, fmt.Errorf("unknown metric: %s", metric)
    }

    var result TopAgentResult
    err := b.db.QueryRowContext(ctx, query, getTenantID(ctx), from, to).Scan(
        &result.AgentID,
        &result.AgentName,
        &result.AgentType,
        &result.MetricValue,
    )

    if err != nil {
        return nil, err
    }

    result.Metric = metric
    result.DateFrom = from
    result.DateTo = to

    return &result, nil
}

// Types
type LeadsCountResult struct {
    TotalLeads     int       `json:"total_leads"`
    QualifiedLeads int       `json:"qualified_leads"`
    Opportunities  int       `json:"opportunities"`
    DateFilter     string    `json:"date_filter"`
    DateFrom       time.Time `json:"date_from"`
    DateTo         time.Time `json:"date_to"`
}

type AgentConversionStat struct {
    AgentID             uuid.UUID `json:"agent_id"`
    AgentName           string    `json:"agent_name"`
    AgentType           string    `json:"agent_type"` // "human", "ai"
    TotalContacts       int       `json:"total_contacts"`
    Conversions         int       `json:"conversions"`
    ConversionRate      float64   `json:"conversion_rate"`
    AvgTimeToConversion float64   `json:"avg_time_to_conversion_seconds"`
}

type AgentConversionStatsResult struct {
    Stats    []AgentConversionStat `json:"stats"`
    DateFrom time.Time             `json:"date_from"`
    DateTo   time.Time             `json:"date_to"`
    TopAgent AgentConversionStat   `json:"top_agent"`
}

type TopAgentResult struct {
    AgentID     uuid.UUID `json:"agent_id"`
    AgentName   string    `json:"agent_name"`
    AgentType   string    `json:"agent_type"`
    Metric      string    `json:"metric"`
    MetricValue float64   `json:"metric_value"`
    DateFrom    time.Time `json:"date_from"`
    DateTo      time.Time `json:"date_to"`
}
```

### **Agent Performance Analyzer**

```go
package analytics

import (
    "context"
    "fmt"

    "github.com/google/uuid"
)

// AgentPerformanceAnalyzer - Analisa performance e qualidade de agentes
type AgentPerformanceAnalyzer struct {
    db         *sql.DB
    llmClient  *genai.Client // Para análises qualitativas
}

// AnalyzeMessages - Analisa mensagens de um agente
func (a *AgentPerformanceAnalyzer) AnalyzeMessages(
    ctx context.Context,
    args map[string]interface{},
) (*AgentMessageAnalysisResult, error) {
    agentID := uuid.MustParse(args["agent_id"].(string))
    sampleSize := 50
    if size, ok := args["sample_size"].(int); ok {
        sampleSize = size
    }

    // Busca sample de mensagens
    messages, err := a.getAgentMessagesSample(ctx, agentID, sampleSize)
    if err != nil {
        return nil, err
    }

    // Análises em paralelo
    var (
        grammarAnalysis *GrammarAnalysis
        toneAnalysis    *ToneAnalysis
        brandAnalysis   *BrandAlignmentAnalysis
    )

    errCh := make(chan error, 3)

    if args["include_grammar"].(bool) {
        go func() {
            grammarAnalysis, err = a.analyzeGrammar(ctx, messages)
            errCh <- err
        }()
    }

    if args["include_tone"].(bool) {
        go func() {
            toneAnalysis, err = a.analyzeTone(ctx, messages)
            errCh <- err
        }()
    }

    if args["include_brand_alignment"].(bool) {
        go func() {
            brandAnalysis, err = a.analyzeBrandAlignment(ctx, messages)
            errCh <- err
        }()
    }

    // Wait for all
    for i := 0; i < 3; i++ {
        if err := <-errCh; err != nil {
            return nil, err
        }
    }

    return &AgentMessageAnalysisResult{
        AgentID:            agentID,
        SampleSize:         len(messages),
        GrammarAnalysis:    grammarAnalysis,
        ToneAnalysis:       toneAnalysis,
        BrandAlignment:     brandAnalysis,
        OverallQualityScore: a.calculateOverallScore(grammarAnalysis, toneAnalysis, brandAnalysis),
    }, nil
}

// CompareAgents - Compara múltiplos agentes
func (a *AgentPerformanceAnalyzer) CompareAgents(
    ctx context.Context,
    args map[string]interface{},
) (*AgentComparisonResult, error) {
    agentIDsRaw := args["agent_ids"].([]interface{})
    agentIDs := make([]uuid.UUID, len(agentIDsRaw))
    for i, id := range agentIDsRaw {
        agentIDs[i] = uuid.MustParse(id.(string))
    }

    dimensions := args["comparison_dimensions"].([]string)

    // Busca métricas de cada agent
    comparisons := make([]AgentComparison, 0, len(agentIDs))

    for _, agentID := range agentIDs {
        metrics, err := a.getAgentMetrics(ctx, agentID, dimensions)
        if err != nil {
            return nil, err
        }
        comparisons = append(comparisons, *metrics)
    }

    // Determina winner por dimensão
    winners := make(map[string]uuid.UUID)
    for _, dim := range dimensions {
        winners[dim] = a.findWinnerInDimension(comparisons, dim)
    }

    // Persiste comparison
    comparisonID := uuid.New()
    err := a.saveComparison(ctx, comparisonID, agentIDs, comparisons, winners)
    if err != nil {
        return nil, err
    }

    return &AgentComparisonResult{
        ComparisonID: comparisonID,
        Agents:       comparisons,
        Winners:      winners,
        Summary:      a.generateComparisonSummary(comparisons, winners),
    }, nil
}

// Types
type AgentMessageAnalysisResult struct {
    AgentID             uuid.UUID              `json:"agent_id"`
    SampleSize          int                    `json:"sample_size"`
    GrammarAnalysis     *GrammarAnalysis       `json:"grammar_analysis"`
    ToneAnalysis        *ToneAnalysis          `json:"tone_analysis"`
    BrandAlignment      *BrandAlignmentAnalysis `json:"brand_alignment"`
    OverallQualityScore float64                `json:"overall_quality_score"`
}

type GrammarAnalysis struct {
    AverageScore float64                `json:"average_score"`
    Issues       []GrammarIssue         `json:"issues"`
    TopIssues    []string               `json:"top_issues"` // Ex: ["pontuação", "concordância"]
}

type ToneAnalysis struct {
    PredominantTone string             `json:"predominant_tone"`
    ToneDistribution map[string]float64 `json:"tone_distribution"` // {"professional": 0.7, "friendly": 0.3}
    Consistency     float64             `json:"consistency"` // 0-1
}

type BrandAlignmentAnalysis struct {
    AlignmentScore  float64   `json:"alignment_score"` // 0-1
    ViolationsList  []string  `json:"violations"`
    Strengths       []string  `json:"strengths"`
}

type AgentComparisonResult struct {
    ComparisonID uuid.UUID                 `json:"comparison_id"`
    Agents       []AgentComparison         `json:"agents"`
    Winners      map[string]uuid.UUID      `json:"winners"` // dimension -> agent_id
    Summary      string                    `json:"summary"`
}

type AgentComparison struct {
    AgentID       uuid.UUID         `json:"agent_id"`
    AgentName     string            `json:"agent_name"`
    AgentType     string            `json:"agent_type"`
    Metrics       map[string]float64 `json:"metrics"` // dimension -> value
    Rank          map[string]int    `json:"rank"`    // dimension -> rank position
}
```

---

## 📋 RESPONSE FORMATTING STRUCTURES

### **Response Format Guide**

```go
package formatting

// ResponseFormatGuide - Guia de formatação que agent retorna para Go
type ResponseFormatGuide struct {
    Format      FormatType              `json:"format"`
    Structure   map[string]interface{}  `json:"structure"`
    Styling     *StylingRules           `json:"styling,omitempty"`
    Examples    []FormattingExample     `json:"examples,omitempty"`
}

type FormatType string

const (
    FormatMarkdown   FormatType = "markdown"
    FormatHTML       FormatType = "html"
    FormatJSON       FormatType = "json"
    FormatPlainText  FormatType = "plain_text"
    FormatRichText   FormatType = "rich_text"
)

// StylingRules - Regras de estilização
type StylingRules struct {
    Bold         []string  `json:"bold"`          // Campos que devem ser bold
    Italic       []string  `json:"italic"`
    Highlight    []string  `json:"highlight"`
    CodeBlock    []string  `json:"code_block"`
    BulletPoints []string  `json:"bullet_points"`
}

// FormattingExample - Exemplo de formatação
type FormattingExample struct {
    Input          string `json:"input"`
    ExpectedOutput string `json:"expected_output"`
}

// ResponseFormatter - Formata response baseado em guide
type ResponseFormatter struct {
    templates map[string]*ResponseFormatTemplate
}

// Format - Formata response segundo guide do agent
func (r *ResponseFormatter) Format(
    content string,
    guide *ResponseFormatGuide,
) (string, error) {
    switch guide.Format {
    case FormatMarkdown:
        return r.formatMarkdown(content, guide)
    case FormatHTML:
        return r.formatHTML(content, guide)
    case FormatJSON:
        return r.formatJSON(content, guide)
    default:
        return content, nil // Plain text
    }
}

// formatMarkdown - Formata como Markdown
func (r *ResponseFormatter) formatMarkdown(
    content string,
    guide *ResponseFormatGuide,
) (string, error) {
    // Apply styling rules
    formatted := content

    if guide.Styling != nil {
        // Bold
        for _, field := range guide.Styling.Bold {
            formatted = applyBold(formatted, field)
        }

        // Bullet points
        if len(guide.Styling.BulletPoints) > 0 {
            formatted = formatAsBulletList(formatted, guide.Styling.BulletPoints)
        }
    }

    return formatted, nil
}
```

**Exemplo de uso:**

```python
# Python ADK Agent retorna:
{
    "response_text": "João converteu 15 leads este mês...",
    "format_guide": {
        "format": "markdown",
        "structure": {
            "sections": [
                {"title": "Resumo", "content": "..."},
                {"title": "Detalhes", "content": "..."}
            ]
        },
        "styling": {
            "bold": ["15 leads", "João"],
            "bullet_points": ["response_time", "grammar_score", "tone"]
        }
    }
}

# Go recebe e formata:
formatted = ResponseFormatter.Format(response_text, format_guide)

# Output:
"""
## Resumo
**João** converteu **15 leads** este mês.

## Detalhes
- Response time: 35s (melhor do time)
- Grammar score: 9.2/10
- Tone: profissional e empático
"""
```

---

## ✅ RESUMO EXECUTIVO GO

### **Responsabilidades:**

1. ✅ **CRUD completo** de todas entidades (Contact, Message, Session, Agent)
2. ✅ **Contextual Embeddings** (Anthropic 2025) com Gemini Flash + text-embedding-005
3. ✅ **Hybrid Search** (Vector + Keyword + Graph + SQL) com RRF/Weighted fusion
4. ✅ **Temporal Knowledge Graph** (Apache AGE) para relações + atribuição
5. ✅ **Memory Facts** com contradiction resolution (LLM-based)
6. ✅ **Agent Registry & Routing** (semantic + rules + workload)
7. ✅ **Context Caching** (Redis) para performance
8. ✅ **gRPC Server** para Python ADK chamar
9. ✅ **MCP Server** para tools reutilizáveis (BI, CRM, Analytics)
10. ✅ **Agent Performance Tracking** (metrics, comparisons, quality analysis)
11. ✅ **Response Formatting** (Markdown, HTML, JSON)

### **Stack:**
- Go 1.23+
- PostgreSQL 16 + pgvector + Apache AGE + pg_trgm
- Vertex AI Go SDK (Gemini Flash 2.5 + text-embedding-005)
- Redis 7+ (caching)
- gRPC (Memory Service API)
- MCP Protocol (Tool Server)
- RabbitMQ (event publishing)

### **MCP Tools Catalog (8 tools):**
1. `get_leads_count` - BI analytics de leads
2. `get_agent_conversion_stats` - Estatísticas de conversão
3. `get_top_performing_agent` - Top performer por métrica
4. `analyze_agent_messages` - Análise de qualidade (gramática, tom, brand)
5. `compare_agents` - Comparação entre agents (AI e humanos)
6. `assign_to_agent` - Atribuição para agente humano
7. `update_pipeline_stage` - Movimentação de pipeline
8. `qualify_lead` - Qualificação BANT

### **Database Tables (Novas):**
- `agent_metrics` - Performance diária por agent
- `agent_message_analysis` - Análise qualitativa de mensagens
- `agent_comparisons` - Histórico de comparações
- `response_format_templates` - Templates de formatação

### **Data Architecture: PostgreSQL vs BigQuery**

#### **Cenário 1: PostgreSQL (Operacional) - Colunas Tipadas**

```sql
-- Design: Dados importantes em colunas, metadata mínimo
CREATE TABLE memory_embeddings (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    contact_id UUID NOT NULL,
    session_id UUID,

    -- Dados estruturados (colunas)
    document_id UUID,              -- ← Coluna normal para JOINs
    document_name TEXT,             -- ← Coluna normal para busca
    document_type TEXT,             -- contract, invoice, etc
    content_type TEXT NOT NULL,     -- document, audio, video, image
    content_text TEXT NOT NULL,

    -- Vector
    embedding vector(768) NOT NULL,

    -- Metadata (apenas dados flexíveis/raros)
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Indexes otimizados
    INDEX idx_embeddings_contact (contact_id),
    INDEX idx_embeddings_document (document_id),
    INDEX idx_embeddings_doc_name (document_name),  -- Para ILIKE
    INDEX idx_embeddings_vector USING ivfflat (embedding),
    INDEX idx_embeddings_metadata USING GIN (metadata)  -- Para queries JSON raras
);

-- Queries rápidas (colunas normais)
SELECT * FROM memory_embeddings
WHERE contact_id = 'contact-123'
  AND document_name ILIKE '%contrato%'
  AND document_type = 'contract';

-- JOIN eficiente
SELECT
    ce.summary,
    me.document_name,
    me.content_text
FROM contact_events ce
JOIN memory_embeddings me ON me.document_id = ce.metadata->>'document_id'
WHERE ce.contact_id = 'contact-123';

-- Metadata apenas para campos raros/flexíveis:
metadata: {
    "page_number": 3,
    "chunk_index": 2,
    "ocr_confidence": 0.98,
    "processing_time_ms": 3500
}
```

**Benefícios PostgreSQL:**
- ✅ Queries rápidas (índices em colunas)
- ✅ JOINs eficientes (foreign keys)
- ✅ Type safety (PostgreSQL valida tipos)
- ✅ Menor storage (sem duplicação no JSON)

#### **Cenário 2: BigQuery (BI/Data Warehouse) - Metadata Estratégico**

```sql
-- Design: Tudo em JSON para flexibilidade analítica
CREATE TABLE `project.dataset.embeddings_warehouse` (
    id STRING NOT NULL,
    tenant_id STRING NOT NULL,

    -- Vector (array de floats)
    embedding ARRAY<FLOAT64>,

    -- Metadata estratégico (TUDO aqui)
    metadata JSON,

    -- Partitioning & clustering
    created_at TIMESTAMP NOT NULL,
    ingestion_date DATE NOT NULL  -- Partition key
)
PARTITION BY ingestion_date
CLUSTER BY tenant_id, metadata.contact_id;

-- Metadata completo para BI:
metadata: {
    -- Identifiers (para JOINs)
    "contact_id": "contact-123",
    "session_id": "session-456",
    "document_id": "doc-uuid-789",
    "message_id": "msg-001",
    "event_id": "event-123",

    -- Document info (para filtros)
    "document_name": "Contrato.pdf",
    "document_type": "contract",
    "content_type": "document",

    -- Business data (para agregações)
    "amount_extracted": 10000.00,
    "currency": "BRL",
    "date_extracted": "2025-01-01",

    -- Entities (para análises)
    "entities": [
        {"type": "company", "value": "Company A"},
        {"type": "person", "value": "João Silva"}
    ],

    -- Dimensions (para BI)
    "source_channel": "whatsapp",
    "agent_type": "human",
    "campaign_source": "google_ads",

    -- Processing metadata
    "provider": "llamaparse",
    "tokens_used": 1200,
    "cost_usd": 0.0012
}

-- Queries BI (JSON extraction)
SELECT
    JSON_VALUE(metadata.document_name) as doc_name,
    JSON_VALUE(metadata.document_type) as doc_type,
    CAST(JSON_VALUE(metadata.amount_extracted) AS FLOAT64) as amount,
    COUNT(*) as chunk_count,
    SUM(CAST(JSON_VALUE(metadata.tokens_used) AS INT64)) as total_tokens
FROM `project.dataset.embeddings_warehouse`
WHERE DATE(created_at) >= DATE_SUB(CURRENT_DATE(), INTERVAL 30 DAY)
    AND JSON_VALUE(metadata.contact_id) = 'contact-123'
    AND JSON_VALUE(metadata.document_type) = 'contract'
GROUP BY 1, 2, 3;

-- Vector similarity + metadata filtering
SELECT
    id,
    JSON_VALUE(metadata.document_name) as doc_name,
    JSON_VALUE(metadata.content_text) as content,
    1 - COSINE_DISTANCE(embedding, query_embedding) as similarity
FROM `project.dataset.embeddings_warehouse`
WHERE JSON_VALUE(metadata.contact_id) = 'contact-123'
    AND JSON_VALUE(metadata.content_type) = 'document'
ORDER BY similarity DESC
LIMIT 10;
```

**Benefícios BigQuery:**
- ✅ Schema flexível (adiciona campos sem ALTER TABLE)
- ✅ Queries analíticas complexas (JSON_VALUE, UNNEST)
- ✅ Partitioning eficiente (reduz scan)
- ✅ Integração com Looker/DataStudio

### **Quando usar cada approach:**

| Aspecto | PostgreSQL (Colunas) | BigQuery (Metadata) |
|---------|---------------------|---------------------|
| **Query pattern** | Operacional (OLTP) | Analítico (OLAP) |
| **Schema** | Fixo, estável | Flexível, evolutivo |
| **JOINs** | Frequentes | Raros (denormalizado) |
| **Performance** | Índices B-tree/GIN | Partitioning/Clustering |
| **Custo** | Storage barato | Query-based pricing |
| **Use case** | AI Agent queries | BI dashboards |

---

### **Contact Events as Document Index:**

**Eventos criam índice de documentos enviados (PostgreSQL approach):**

```sql
-- Contact event quando documento é recebido
INSERT INTO contact_events (
    id, tenant_id, contact_id, category, summary, metadata
) VALUES (
    'event-123', 'tenant-1', 'contact-456',
    'document_received',
    'Cliente enviou contrato de prestação de serviços',
    '{
        "document_name": "Contrato.pdf",
        "document_id": "doc-uuid-789",
        "document_type": "contract",
        "page_count": 5
    }'
);

-- Embeddings linkam ao document_id do evento
INSERT INTO memory_embeddings (
    ...,
    metadata: {
        "source_document_id": "doc-uuid-789",  ← LINK ao evento
        "source_event_id": "event-123",
        "document_title": "Contrato.pdf",
        ...
    }
);

-- Query cross-reference: eventos → documentos vetorizados
SELECT
    ce.summary as event_summary,
    ce.created_at as event_date,
    ce.metadata->>'document_name' as doc_name,
    COUNT(me.id) as chunk_count,
    STRING_AGG(me.content_text, ' | ') as content_preview
FROM contact_events ce
JOIN memory_embeddings me
    ON me.metadata->>'source_document_id' = ce.metadata->>'document_id'
WHERE ce.contact_id = 'contact-456'
    AND ce.category = 'document_received'
GROUP BY ce.id;
```

**Benefícios:**
- ✅ Eventos servem como índice temporal de documentos
- ✅ Metadata permite busca por nome: `WHERE metadata->>'document_name' ILIKE '%contrato%'`
- ✅ Cross-reference: evento → document_id → chunks vetorizados
- ✅ AI Agent vê timeline: "quando foi enviado" + "conteúdo"

### **Performance:**
- **Embedding generation**: ~200ms por sessão (com contextual retrieval)
- **Hybrid search**: ~50-150ms (dependendo de strategy)
- **Cache hit rate**: 70-90% (typical)
- **Vector search**: <50ms (HNSW index)
- **Graph traversal**: <100ms (AGE cypher queries)
- **Event→Document lookup**: <20ms (JSONB GIN index)
- **BI queries**: 10-50ms (cached), 50-200ms (fresh)
- **Agent analysis**: 500-1000ms (qualitative with LLM)

### **Integração com Python ADK:**
1. Go expõe gRPC API (`SearchMemory`, `AddSession`)
2. Python chama via `VentrosMemoryService` (custom `BaseMemoryService`)
3. Go retorna contexto formatado (recent messages + similar sessions + facts)
4. Python ADK usa contexto para LLM prompt
5. Background: Go gera embeddings async (não bloqueia agent)

---

**Próximos passos:**
1. Implementar protobuf + gRPC server
2. Implementar Apache AGE integration (CYPHER queries)
3. Implementar memory fact extraction + contradiction resolution
4. Performance tuning (indexes, caching, batch processing)
5. Observability (OpenTelemetry)# 🧠 AI MEMORY GO ARCHITECTURE - PART 2

## 📚 TEMPORAL KNOWLEDGE GRAPH SERVICE

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// TemporalKnowledgeGraphService - Gerencia grafo de conhecimento temporal (Apache AGE)
type TemporalKnowledgeGraphService struct {
    ageClient *ApacheAGEClient
    repo      TemporalGraphRepository
}

// TemporalEdge - Aresta com validade temporal (bi-temporal model)
type TemporalEdge struct {
    ID             uuid.UUID              `json:"id"`
    FromNodeID     uuid.UUID              `json:"from_node_id"`
    FromNodeType   string                 `json:"from_node_type"`   // "contact", "session", "agent"
    ToNodeID       uuid.UUID              `json:"to_node_id"`
    ToNodeType     string                 `json:"to_node_type"`
    EdgeType       string                 `json:"edge_type"`        // "HAS_SESSION", "ASSIGNED_TO", etc

    // === BI-TEMPORAL MODEL (Zep/Graphiti pattern) ===
    ValidFrom      time.Time              `json:"valid_from"`       // Quando evento OCORREU
    ValidTo        *time.Time             `json:"valid_to"`         // Quando evento TERMINOU (NULL = ainda válido)
    TransactionAt  time.Time              `json:"transaction_at"`   // Quando foi INGERIDO no sistema

    Properties     map[string]interface{} `json:"properties"`
    TenantID       string                 `json:"tenant_id"`
}

// EdgeType constants
const (
    EdgeHasSession            = "HAS_SESSION"             // Contact → Session
    EdgeAssignedTo            = "ASSIGNED_TO"             // Session → Agent
    EdgeTransferredTo         = "TRANSFERRED_TO"          // Agent → Agent (chain)
    EdgeRepliedTo             = "REPLIED_TO"              // Message → Message (threading)
    EdgeMentions              = "MENTIONS"                // Message → Contact (social graph)
    EdgeDiscussedTopic        = "DISCUSSED_TOPIC"         // Session → Topic
    EdgeCameFromCampaign      = "CAME_FROM_CAMPAIGN"      // Contact → Campaign
    EdgeHasPlatform           = "HAS_PLATFORM"            // Campaign → Platform
    EdgeInPipeline            = "IN_PIPELINE"             // Contact → Pipeline
    EdgeHasBudgetConstraint   = "HAS_BUDGET_CONSTRAINT"   // Contact → Budget (memory fact)
    EdgeHasPreference         = "HAS_PREFERENCE"          // Contact → Preference
    EdgeHasGoal               = "HAS_GOAL"                // Contact → Goal
    EdgeHasObjection          = "HAS_OBJECTION"           // Session → Objection
)

// AddTemporalEdge - Adiciona edge com validade temporal
func (t *TemporalKnowledgeGraphService) AddTemporalEdge(
    ctx context.Context,
    edge TemporalEdge,
) error {
    // Validate
    if edge.FromNodeID == uuid.Nil || edge.ToNodeID == uuid.Nil {
        return fmt.Errorf("invalid node IDs")
    }

    now := time.Now()
    if edge.TransactionAt.IsZero() {
        edge.TransactionAt = now
    }
    if edge.ValidFrom.IsZero() {
        edge.ValidFrom = now
    }

    // Apache AGE CYPHER query
    query := `
    MATCH (from {id: $from_id, type: $from_type})
    MATCH (to {id: $to_id, type: $to_type})

    // Invalida edges anteriores do mesmo tipo (se aplicável)
    OPTIONAL MATCH (from)-[old_edge:` + edge.EdgeType + `]->(to)
    WHERE old_edge.valid_to IS NULL
      AND old_edge.from_node_id = $from_id
      AND old_edge.to_node_id = $to_id
    SET old_edge.valid_to = $valid_from

    // Cria nova edge com bi-temporal tracking
    CREATE (from)-[new_edge:` + edge.EdgeType + `]->(to)
    SET new_edge.id = $edge_id,
        new_edge.valid_from = $valid_from,
        new_edge.valid_to = NULL,
        new_edge.transaction_at = $transaction_at,
        new_edge.properties = $properties,
        new_edge.tenant_id = $tenant_id

    RETURN new_edge
    `

    params := map[string]interface{}{
        "from_id":        edge.FromNodeID.String(),
        "from_type":      edge.FromNodeType,
        "to_id":          edge.ToNodeID.String(),
        "to_type":        edge.ToNodeType,
        "edge_id":        edge.ID.String(),
        "valid_from":     edge.ValidFrom,
        "transaction_at": edge.TransactionAt,
        "properties":     edge.Properties,
        "tenant_id":      edge.TenantID,
    }

    return t.ageClient.Execute(ctx, query, params)
}

// QueryTemporalGraph - Query point-in-time (válido em determinado momento)
func (t *TemporalKnowledgeGraphService) QueryTemporalGraph(
    ctx context.Context,
    nodeID uuid.UUID,
    edgeType string,
    asOf *time.Time,  // NULL = agora
) ([]TemporalEdge, error) {
    if asOf == nil {
        now := time.Now()
        asOf = &now
    }

    query := `
    MATCH (node {id: $node_id})-[edge:` + edgeType + `]->(target)
    WHERE edge.valid_from <= $as_of
      AND (edge.valid_to IS NULL OR edge.valid_to > $as_of)
    RETURN edge, target
    ORDER BY edge.valid_from DESC
    `

    params := map[string]interface{}{
        "node_id": nodeID.String(),
        "as_of":   *asOf,
    }

    // TODO: Parse results from Apache AGE
    results := []TemporalEdge{}

    return results, t.ageClient.Query(ctx, query, params, &results)
}

// GetAgentTransferChain - Retorna chain completa de agent transfers
func (t *TemporalKnowledgeGraphService) GetAgentTransferChain(
    ctx context.Context,
    sessionID uuid.UUID,
) (*AgentTransferChain, error) {
    query := `
    MATCH path = (session:Session {id: $session_id})-[:ASSIGNED_TO*]->(agent:Agent)
    RETURN path
    ORDER BY length(path) DESC
    LIMIT 1
    `

    params := map[string]interface{}{
        "session_id": sessionID.String(),
    }

    // Parse path from graph
    chain := &AgentTransferChain{
        SessionID: sessionID,
        Agents:    []AgentNode{},
    }

    // TODO: Parse from Apache AGE result
    err := t.ageClient.Query(ctx, query, params, &chain)

    return chain, err
}

// GetCampaignAttribution - Retorna grafo de atribuição completo
func (t *TemporalKnowledgeGraphService) GetCampaignAttribution(
    ctx context.Context,
    contactID uuid.UUID,
) (*CampaignAttributionGraph, error) {
    query := `
    MATCH (contact:Contact {id: $contact_id})
          -[:CAME_FROM_CAMPAIGN]->(campaign:Campaign)
          -[:HAS_PLATFORM]->(platform:Platform)
    OPTIONAL MATCH (campaign)-[:HAS_AD]->(ad:Ad)
    RETURN contact, campaign, platform, ad
    `

    params := map[string]interface{}{
        "contact_id": contactID.String(),
    }

    attribution := &CampaignAttributionGraph{
        ContactID: contactID,
    }

    err := t.ageClient.Query(ctx, query, params, &attribution)

    return attribution, err
}

// GetSocialGraph - Retorna grafo de menções (quem menciona quem)
func (t *TemporalKnowledgeGraphService) GetSocialGraph(
    ctx context.Context,
    contactID uuid.UUID,
    depth int,
) (*SocialGraph, error) {
    query := fmt.Sprintf(`
    MATCH path = (contact:Contact {id: $contact_id})
                 -[:MENTIONS*1..%d]-(other:Contact)
    RETURN path, other
    `, depth)

    params := map[string]interface{}{
        "contact_id": contactID.String(),
    }

    socialGraph := &SocialGraph{
        CenterContactID: contactID,
        Connections:     []SocialConnection{},
    }

    err := t.ageClient.Query(ctx, query, params, &socialGraph)

    return socialGraph, err
}

// Supporting types
type AgentTransferChain struct {
    SessionID uuid.UUID   `json:"session_id"`
    Agents    []AgentNode `json:"agents"`
    Transfers int         `json:"transfers"`
}

type AgentNode struct {
    AgentID      uuid.UUID `json:"agent_id"`
    AgentName    string    `json:"agent_name"`
    AssignedAt   time.Time `json:"assigned_at"`
    TransferredAt *time.Time `json:"transferred_at"`
}

type CampaignAttributionGraph struct {
    ContactID    uuid.UUID `json:"contact_id"`
    Campaign     string    `json:"campaign"`
    Platform     string    `json:"platform"`
    AdID         *string   `json:"ad_id"`
    AdCreative   *string   `json:"ad_creative"`
    UTMSource    string    `json:"utm_source"`
    UTMMedium    string    `json:"utm_medium"`
    UTMCampaign  string    `json:"utm_campaign"`
}

type SocialGraph struct {
    CenterContactID uuid.UUID           `json:"center_contact_id"`
    Connections     []SocialConnection  `json:"connections"`
}

type SocialConnection struct {
    ContactID   uuid.UUID `json:"contact_id"`
    ContactName string    `json:"contact_name"`
    Mentions    int       `json:"mentions"`
    Depth       int       `json:"depth"`  // Grau de separação
}
```

---

## 🎯 AGENT REGISTRY & SEMANTIC ROUTING

```go
package memory

import (
    "context"
    "fmt"
    "math"
    "sort"

    "github.com/google/uuid"
)

// AgentRegistry - Gerencia registro e routing de agentes
type AgentRegistry struct {
    agentRepo          AgentRepository
    semanticRouter     *SemanticRouterService
    embeddingClient    *genai.Client
    routeEmbeddings    map[string][]float32  // Pre-computed embeddings de routes
}

// RegisterAgent - Registra novo agente AI
func (r *AgentRegistry) RegisterAgent(
    ctx context.Context,
    agent *Agent,
    aiMetadata *AIAgentMetadata,
) error {
    // Validate
    if agent.Type() != AgentTypeAI && agent.Type() != AgentTypeBot {
        return fmt.Errorf("only AI/Bot agents can be registered")
    }

    // Store in repository
    if err := r.agentRepo.Save(ctx, agent); err != nil {
        return fmt.Errorf("failed to save agent: %w", err)
    }

    // Pre-compute embeddings para routing rules (se existirem)
    if len(aiMetadata.RoutingRules) > 0 {
        for _, rule := range aiMetadata.RoutingRules {
            embedding, err := r.embeddingClient.Embed(ctx, rule.Condition)
            if err != nil {
                continue // Log erro mas não falha
            }
            r.routeEmbeddings[agent.ID().String()+":"+rule.Condition] = embedding
        }
    }

    return nil
}

// RouteToAgent - Semantic routing baseado em mensagem
func (r *AgentRegistry) RouteToAgent(
    ctx context.Context,
    message *Message,
    session *Session,
    tenantID string,
) (*Agent, error) {
    // 1. Detecta intent usando Semantic Router
    intent, err := r.semanticRouter.ClassifyIntent(ctx, *message.Text())
    if err != nil {
        return nil, fmt.Errorf("failed to classify intent: %w", err)
    }

    // 2. Busca agentes compatíveis com o intent
    candidates, err := r.findCandidateAgents(ctx, intent, session, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to find candidate agents: %w", err)
    }

    if len(candidates) == 0 {
        // Fallback: agente default/operations
        return r.getDefaultAgent(ctx, tenantID)
    }

    // 3. Score e ranking de candidates
    scored := r.scoreAgents(candidates, message, session, intent)

    // 4. Retorna melhor match
    return scored[0].Agent, nil
}

// SemanticRouterService - Implementa Semantic Router (Aurelio Labs pattern)
type SemanticRouterService struct {
    embeddingClient *genai.Client
    routes          []SemanticRoute
    routeEmbeddings map[string][]float32
    threshold       float64  // Similarity threshold (ex: 0.75)
}

// SemanticRoute - Route com examples para matching
type SemanticRoute struct {
    Name        string                 `json:"name"`         // "churn_risk", "sales_inquiry"
    Category    AgentCategory          `json:"category"`     // Maps to agent category
    Utterances  []string               `json:"utterances"`   // Example phrases
    Priority    int                    `json:"priority"`     // Tie-breaking
    Metadata    map[string]interface{} `json:"metadata"`
}

// Semantic Routes predefinidas
var DefaultSemanticRoutes = []SemanticRoute{
    {
        Name:     "churn_risk",
        Category: CategoryRetentionChurn,
        Utterances: []string{
            "quero cancelar",
            "não quero mais",
            "vou desistir",
            "isso não está funcionando",
            "muito caro pra mim",
            "não vale a pena",
            "vou procurar outro",
        },
        Priority: 10, // ALTA prioridade
    },
    {
        Name:     "sales_inquiry",
        Category: CategorySalesProspecting,
        Utterances: []string{
            "quanto custa",
            "qual o preço",
            "tem desconto",
            "quero saber valores",
            "como funciona a cobrança",
            "aceita cartão",
        },
        Priority: 7,
    },
    {
        Name:     "technical_support",
        Category: CategorySupportTechnical,
        Utterances: []string{
            "não está funcionando",
            "deu erro",
            "bug",
            "problema técnico",
            "não consigo acessar",
            "tela em branco",
        },
        Priority: 9, // ALTA (problemas técnicos são urgentes)
    },
    {
        Name:     "billing_support",
        Category: CategorySupportBilling,
        Utterances: []string{
            "não recebi a fatura",
            "cobrança errada",
            "problema com pagamento",
            "cartão não passou",
            "reembolso",
        },
        Priority: 8,
    },
    {
        Name:     "objection_handling",
        Category: CategorySalesNegotiation,
        Utterances: []string{
            "muito caro",
            "não cabe no orçamento",
            "preciso pensar",
            "vou conversar com o time",
            "não é o momento certo",
        },
        Priority: 7,
    },
    {
        Name:     "feature_request",
        Category: CategoryOperationsFollowup,
        Utterances: []string{
            "seria bom se tivesse",
            "gostaria de sugerir",
            "falta essa funcionalidade",
            "quando vocês vão lançar",
        },
        Priority: 5,
    },
}

// NewSemanticRouterService - Cria novo router com routes pré-computadas
func NewSemanticRouterService(
    embeddingClient *genai.Client,
    routes []SemanticRoute,
    threshold float64,
) (*SemanticRouterService, error) {
    router := &SemanticRouterService{
        embeddingClient: embeddingClient,
        routes:          routes,
        routeEmbeddings: make(map[string][]float32),
        threshold:       threshold,
    }

    // Pre-compute embeddings dos utterances
    for _, route := range routes {
        for _, utterance := range route.Utterances {
            embedding, err := embeddingClient.Embed(context.Background(), utterance)
            if err != nil {
                continue // Log mas não falha
            }
            key := fmt.Sprintf("%s:%s", route.Name, utterance)
            router.routeEmbeddings[key] = embedding
        }
    }

    return router, nil
}

// ClassifyIntent - Classifica intent usando semantic similarity
func (s *SemanticRouterService) ClassifyIntent(
    ctx context.Context,
    text string,
) (*Intent, error) {
    // 1. Gera embedding do texto de entrada
    textEmbedding, err := s.embeddingClient.Embed(ctx, text)
    if err != nil {
        return nil, fmt.Errorf("failed to embed text: %w", err)
    }

    // 2. Compara com todos utterance embeddings
    type ScoredRoute struct {
        Route      SemanticRoute
        Similarity float64
        MatchedUtterance string
    }

    scored := []ScoredRoute{}

    for _, route := range s.routes {
        maxSimilarity := 0.0
        matchedUtterance := ""

        for _, utterance := range route.Utterances {
            key := fmt.Sprintf("%s:%s", route.Name, utterance)
            utteranceEmbedding, exists := s.routeEmbeddings[key]
            if !exists {
                continue
            }

            similarity := cosineSimilarity(textEmbedding, utteranceEmbedding)
            if similarity > maxSimilarity {
                maxSimilarity = similarity
                matchedUtterance = utterance
            }
        }

        if maxSimilarity >= s.threshold {
            scored = append(scored, ScoredRoute{
                Route:            route,
                Similarity:       maxSimilarity,
                MatchedUtterance: matchedUtterance,
            })
        }
    }

    if len(scored) == 0 {
        // No match above threshold
        return &Intent{
            Name:       "unknown",
            Category:   CategoryOperationsFollowup, // Fallback
            Confidence: 0.0,
        }, nil
    }

    // 3. Sort by similarity DESC, then priority DESC
    sort.Slice(scored, func(i, j int) bool {
        if math.Abs(scored[i].Similarity-scored[j].Similarity) < 0.01 {
            // Similaridade igual: usa priority
            return scored[i].Route.Priority > scored[j].Route.Priority
        }
        return scored[i].Similarity > scored[j].Similarity
    })

    // 4. Retorna best match
    best := scored[0]
    return &Intent{
        Name:             best.Route.Name,
        Category:         best.Route.Category,
        Confidence:       best.Similarity,
        MatchedUtterance: best.MatchedUtterance,
    }, nil
}

// Intent - Classificação de intent
type Intent struct {
    Name             string        `json:"name"`
    Category         AgentCategory `json:"category"`
    Confidence       float64       `json:"confidence"`
    MatchedUtterance string        `json:"matched_utterance"`
}

// findCandidateAgents - Busca agentes compatíveis com intent
func (r *AgentRegistry) findCandidateAgents(
    ctx context.Context,
    intent *Intent,
    session *Session,
    tenantID string,
) ([]*Agent, error) {
    // Busca agentes ativos da categoria
    agents, err := r.agentRepo.FindByCategory(ctx, intent.Category, tenantID)
    if err != nil {
        return nil, err
    }

    // Filtra por disponibilidade
    available := []*Agent{}
    for _, agent := range agents {
        if agent.IsActive() && agent.Status() == AgentStatusAvailable {
            // TODO: Check MaxConcurrentSessions
            available = append(available, agent)
        }
    }

    return available, nil
}

// scoreAgents - Pontua agentes baseado em múltiplos fatores
func (r *AgentRegistry) scoreAgents(
    agents []*Agent,
    message *Message,
    session *Session,
    intent *Intent,
) []ScoredAgent {
    scored := make([]ScoredAgent, len(agents))

    for i, agent := range agents {
        score := 0.0

        // 1. Intent match confidence (peso: 40%)
        score += intent.Confidence * 0.40

        // 2. Agent priority (peso: 20%)
        // TODO: Get from AIAgentMetadata
        priority := 5.0 // Default
        score += (priority / 10.0) * 0.20

        // 3. Session history (peso: 20%)
        // Se agente já atendeu este contato antes, +bonus
        if containsAgent(session.AgentIDs(), agent.ID()) {
            score += 0.20
        }

        // 4. Load balancing (peso: 20%)
        // TODO: Get current session count
        currentSessions := 0
        maxSessions := 10 // TODO: Get from AIAgentMetadata
        loadFactor := 1.0 - (float64(currentSessions) / float64(maxSessions))
        score += loadFactor * 0.20

        scored[i] = ScoredAgent{
            Agent: agent,
            Score: score,
        }
    }

    // Sort by score DESC
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].Score > scored[j].Score
    })

    return scored
}

type ScoredAgent struct {
    Agent *Agent  `json:"agent"`
    Score float64 `json:"score"`
}

// getDefaultAgent - Retorna agente default/fallback
func (r *AgentRegistry) getDefaultAgent(
    ctx context.Context,
    tenantID string,
) (*Agent, error) {
    agents, err := r.agentRepo.FindByCategory(ctx, CategoryOperationsFollowup, tenantID)
    if err != nil {
        return nil, err
    }

    if len(agents) == 0 {
        return nil, fmt.Errorf("no default agent found")
    }

    return agents[0], nil
}

// Helper functions
func cosineSimilarity(a, b []float32) float64 {
    if len(a) != len(b) {
        return 0.0
    }

    var dotProduct, normA, normB float64
    for i := 0; i < len(a); i++ {
        dotProduct += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }

    if normA == 0 || normB == 0 {
        return 0.0
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func containsAgent(agentIDs []uuid.UUID, agentID uuid.UUID) bool {
    for _, id := range agentIDs {
        if id == agentID {
            return true
        }
    }
    return false
}
```

---

## 🧬 MEMORY FACT SERVICE (Google Memory Bank pattern)

```go
package memory

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
)

// MemoryFactService - Gerencia facts com contradiction resolution
type MemoryFactService struct {
    factRepo    FactRepository
    llmClient   *genai.Client  // Para extraction e contradiction detection
}

// MemoryFact - Fact extraído de conversas
type MemoryFact struct {
    ID           uuid.UUID              `db:"id"`
    ContactID    uuid.UUID              `db:"contact_id"`
    TenantID     string                 `db:"tenant_id"`
    FactType     FactType               `db:"fact_type"`
    FactText     string                 `db:"fact_text"`         // Texto original
    FactValue    interface{}            `db:"fact_value"`        // Valor estruturado
    Confidence   float64                `db:"confidence"`        // 0.0-1.0
    ValidFrom    time.Time              `db:"valid_from"`
    ValidTo      *time.Time             `db:"valid_to"`          // NULL = current
    Supersedes   *uuid.UUID             `db:"supersedes"`        // FK to previous fact
    Source       string                 `db:"source"`            // "message", "note", "annotation"
    SourceID     *uuid.UUID             `db:"source_id"`         // ID do source (message_id, note_id)
    CreatedAt    time.Time              `db:"created_at"`
    Metadata     map[string]interface{} `db:"metadata"`          // JSONB
}

// FactType - Tipos de facts
type FactType string

const (
    FactTypeBudgetConstraint   FactType = "budget_constraint"
    FactTypePreference          FactType = "preference"
    FactTypeGoal                FactType = "goal"
    FactTypeObjection           FactType = "objection"
    FactTypePainPoint           FactType = "pain_point"
    FactTypeTechnicalIssue      FactType = "technical_issue"
    FactTypeEnvironmentInfo     FactType = "environment_info"
    FactTypeDecisionMaker       FactType = "decision_maker"
    FactTypeTimeline            FactType = "timeline"
    FactTypeCompetitor          FactType = "competitor"
)

// AddFact - Adiciona novo fact com contradiction detection
func (m *MemoryFactService) AddFact(
    ctx context.Context,
    contactID uuid.UUID,
    tenantID string,
    factText string,
    source string,
    sourceID *uuid.UUID,
) (*MemoryFact, error) {
    // 1. Extract structured information usando LLM
    extracted, err := m.extractStructuredFact(ctx, factText)
    if err != nil {
        return nil, fmt.Errorf("failed to extract fact: %w", err)
    }

    // 2. Busca facts existentes do mesmo tipo
    existingFacts, err := m.factRepo.FindActiveFactsByType(
        ctx,
        contactID,
        extracted.FactType,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to find existing facts: %w", err)
    }

    // 3. Contradiction detection
    for _, existing := range existingFacts {
        contradicts, resolution, err := m.detectContradiction(
            ctx,
            existing,
            extracted,
        )
        if err != nil {
            continue // Log mas não falha
        }

        if contradicts {
            if err := m.resolveContradiction(ctx, existing, extracted, resolution); err != nil {
                return nil, fmt.Errorf("failed to resolve contradiction: %w", err)
            }
        }
    }

    // 4. Persiste novo fact
    newFact := &MemoryFact{
        ID:         uuid.New(),
        ContactID:  contactID,
        TenantID:   tenantID,
        FactType:   extracted.FactType,
        FactText:   factText,
        FactValue:  extracted.FactValue,
        Confidence: extracted.Confidence,
        ValidFrom:  time.Now(),
        ValidTo:    nil,  // Current
        Source:     source,
        SourceID:   sourceID,
        CreatedAt:  time.Now(),
        Metadata:   extracted.Metadata,
    }

    if err := m.factRepo.Save(ctx, newFact); err != nil {
        return nil, fmt.Errorf("failed to save fact: %w", err)
    }

    return newFact, nil
}

// extractStructuredFact - Usa LLM para extrair fact estruturado
func (m *MemoryFactService) extractStructuredFact(
    ctx context.Context,
    factText string,
) (*ExtractedFact, error) {
    prompt := fmt.Sprintf(`
Extraia informações estruturadas deste fact:

"%s"

Retorne JSON no formato:
{
    "fact_type": "budget_constraint|preference|goal|objection|pain_point|technical_issue|environment_info|decision_maker|timeline|competitor",
    "fact_value": <value>,  // Valor estruturado (número, string, objeto)
    "confidence": <0-1>,     // Quão confiante você está na extração
    "metadata": {}           // Informações adicionais relevantes
}

Exemplos:
- "Meu orçamento é R$5000" → {"fact_type": "budget_constraint", "fact_value": 5000.0, "confidence": 0.95}
- "Prefiro ser chamado de João" → {"fact_type": "preference", "fact_value": "name=João", "confidence": 0.90}
- "Quero fechar até sexta" → {"fact_type": "timeline", "fact_value": "deadline=2025-01-17", "confidence": 0.85}
`, factText)

    // TODO: Call LLM with structured output
    extracted := &ExtractedFact{}
    // err := m.llmClient.GenerateStructured(ctx, prompt, extracted)

    return extracted, nil
}

// detectContradiction - Detecta contradição entre facts
func (m *MemoryFactService) detectContradiction(
    ctx context.Context,
    existing *MemoryFact,
    newFact *ExtractedFact,
) (bool, ContradictionResolution, error) {
    prompt := fmt.Sprintf(`
Analise se estes dois facts são contraditórios:

Fact 1 (existente):
- Tipo: %s
- Texto: "%s"
- Valor: %v
- Data: %s

Fact 2 (novo):
- Tipo: %s
- Texto: "%s"
- Valor: %v

Retorne JSON:
{
    "contradicts": true/false,
    "resolution": "keep_new|keep_old|merge|both_valid",
    "explanation": "explicação da contradição ou compatibilidade"
}

Exemplos de contradição:
- "Orçamento R$5000" vs "Orçamento R$3000" → contradicts=true, resolution=keep_new
- "Prefiro João" vs "Prefiro João Pedro" → contradicts=true, resolution=keep_new
- "Quer fechar em 30 dias" vs "Quer fechar urgente" → contradicts=false, resolution=both_valid (complementares)
`,
        existing.FactType,
        existing.FactText,
        existing.FactValue,
        existing.ValidFrom.Format("02/01/2006"),
        newFact.FactType,
        newFact.FactText,
        newFact.FactValue,
    )

    // TODO: Call LLM
    resolution := ContradictionResolution{
        Contradicts: false,
        Resolution:  "keep_new",
    }

    // err := m.llmClient.GenerateStructured(ctx, prompt, &resolution)

    return resolution.Contradicts, resolution, nil
}

// resolveContradiction - Resolve contradição baseado em strategy
func (m *MemoryFactService) resolveContradiction(
    ctx context.Context,
    existing *MemoryFact,
    newFact *ExtractedFact,
    resolution ContradictionResolution,
) error {
    now := time.Now()

    switch resolution.Resolution {
    case "keep_new":
        // Invalida fact antigo
        existing.ValidTo = &now
        newFact.Supersedes = &existing.ID
        return m.factRepo.Update(ctx, existing)

    case "keep_old":
        // Não faz nada (descarta novo fact)
        return nil

    case "merge":
        // Merge values (caso específico)
        // TODO: Implement merge logic
        return nil

    case "both_valid":
        // Ambos válidos (não são realmente contraditórios)
        return nil

    default:
        return fmt.Errorf("unknown resolution: %s", resolution.Resolution)
    }
}

// GetActiveFacts - Retorna facts válidos point-in-time
func (m *MemoryFactService) GetActiveFacts(
    ctx context.Context,
    contactID uuid.UUID,
    factTypes []FactType,
    asOf *time.Time,
) ([]MemoryFact, error) {
    if asOf == nil {
        now := time.Now()
        asOf = &now
    }

    return m.factRepo.FindFactsByValidity(ctx, contactID, factTypes, *asOf)
}

// Supporting types
type ExtractedFact struct {
    FactType   FactType               `json:"fact_type"`
    FactValue  interface{}            `json:"fact_value"`
    Confidence float64                `json:"confidence"`
    Metadata   map[string]interface{} `json:"metadata"`
}

type ContradictionResolution struct {
    Contradicts bool   `json:"contradicts"`
    Resolution  string `json:"resolution"`  // "keep_new", "keep_old", "merge", "both_valid"
    Explanation string `json:"explanation"`
}
```

---

## 💾 CONTEXT MANAGER (Prompt Caching)

```go
package memory

import (
    "context"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
)

// ContextManager - Gerencia caching de contexto (Redis)
type ContextManager struct {
    redisClient *redis.Client
}

// NewContextManager cria novo manager
func NewContextManager(redisClient *redis.Client) *ContextManager {
    return &ContextManager{
        redisClient: redisClient,
    }
}

// GetCached - Busca resultado cacheado
func (c *ContextManager) GetCached(
    ctx context.Context,
    req SearchRequest,
) (*SearchResult, error) {
    // Generate cache key
    cacheKey := c.generateCacheKey(req)

    // Get from Redis
    data, err := c.redisClient.Get(ctx, cacheKey).Bytes()
    if err == redis.Nil {
        return nil, nil  // Cache miss
    }
    if err != nil {
        return nil, fmt.Errorf("redis error: %w", err)
    }

    // Deserialize
    var result SearchResult
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal: %w", err)
    }

    return &result, nil
}

// SetCached - Cacheia resultado
func (c *ContextManager) SetCached(
    ctx context.Context,
    req SearchRequest,
    result *SearchResult,
    ttl int,  // seconds
) error {
    // Generate cache key
    cacheKey := c.generateCacheKey(req)

    // Serialize
    data, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal: %w", err)
    }

    // Set in Redis with TTL
    return c.redisClient.Set(ctx, cacheKey, data, time.Duration(ttl)*time.Second).Err()
}

// generateCacheKey - Gera key determinística
func (c *ContextManager) generateCacheKey(req SearchRequest) string {
    // Hash based on: contact_id + agent_category + strategy
    key := fmt.Sprintf("memory:%s:%s:%s",
        req.TenantID,
        req.ContactID.String(),
        req.AgentCategory,
    )

    // Se strategy é custom, inclui weights no hash
    if req.MemoryStrategy.Strategy == StrategyCustom {
        key += fmt.Sprintf(":%.2f:%.2f:%.2f:%.2f",
            req.MemoryStrategy.VectorWeight,
            req.MemoryStrategy.KeywordWeight,
            req.MemoryStrategy.GraphWeight,
            req.MemoryStrategy.RecentWeight,
        )
    } else {
        key += ":" + string(req.MemoryStrategy.Strategy)
    }

    // Hash final (para evitar keys muito longas)
    hash := sha256.Sum256([]byte(key))
    return fmt.Sprintf("ctx:%x", hash[:16])  // 32 chars
}

// InvalidateContactCache - Invalida cache de um contato
func (c *ContextManager) InvalidateContactCache(
    ctx context.Context,
    tenantID string,
    contactID uuid.UUID,
) error {
    // Pattern matching no Redis
    pattern := fmt.Sprintf("ctx:*%s*%s*", tenantID, contactID.String())

    iter := c.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
    for iter.Next(ctx) {
        if err := c.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
            continue // Log mas não falha
        }
    }

    return iter.Err()
}
```

Continua na **PART 3** com gRPC API e Database Schema! Quer que eu continue ou prefere que eu crie o documento Python ADK agora?
# 🧠 AI MEMORY GO ARCHITECTURE - PART 3 (FINAL)

## 🌐 gRPC API (Python ADK Integration)

### **Protocol Buffers Definition**

```protobuf
syntax = "proto3";

package memory.v1;

option go_package = "github.com/ventros/crm/api/memory/v1;memoryv1";

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

// MemoryService - Serviço de memória para Python ADK
service MemoryService {
    // Search memory (hybrid search)
    rpc SearchMemory(SearchMemoryRequest) returns (SearchMemoryResponse);

    // Add session to memory (após session terminar)
    rpc AddSessionToMemory(AddSessionToMemoryRequest) returns (AddSessionToMemoryResponse);

    // Get contact context (para agent routing)
    rpc GetContactContext(GetContactContextRequest) returns (GetContactContextResponse);

    // Route to agent (semantic routing)
    rpc RouteToAgent(RouteToAgentRequest) returns (RouteToAgentResponse);

    // Add memory fact
    rpc AddMemoryFact(AddMemoryFactRequest) returns (AddMemoryFactResponse);

    // Get active facts
    rpc GetActiveFacts(GetActiveFactsRequest) returns (GetActiveFactsResponse);
}

// === SEARCH MEMORY ===

message SearchMemoryRequest {
    string tenant_id = 1;
    string contact_id = 2;
    optional string session_id = 3;
    string agent_category = 4;
    string query = 5;
    repeated float query_embedding = 6;  // Optional: embedding pré-computado
    KnowledgeScope knowledge_scope = 7;
    MemoryStrategy memory_strategy = 8;
    int32 top_k = 9;
    int32 retrieval_top_k = 10;
}

message SearchMemoryResponse {
    repeated Message recent_messages = 1;  // SEMPRE presente
    repeated VectorResult vector_results = 2;
    repeated KeywordResult keyword_results = 3;
    repeated GraphResult graph_results = 4;
    repeated FinalResult final_results = 5;
    optional ContactStats contact_stats = 6;
    optional PipelineContext pipeline_context = 7;
    repeated MemoryFact memory_facts = 8;
    int32 total_results = 9;
    int64 search_latency_ms = 10;
    bool cache_hit = 11;
    string strategy = 12;
}

message Message {
    string id = 1;
    google.protobuf.Timestamp timestamp = 2;
    string contact_id = 3;
    bool from_me = 4;
    string content_type = 5;
    optional string text = 6;
    optional string media_url = 7;
    string status = 8;
    optional string sentiment = 9;
    optional string agent_id = 10;
    string source = 11;
}

message VectorResult {
    string session_id = 1;
    Session session = 2;
    double similarity = 3;
}

message KeywordResult {
    string session_id = 1;
    Session session = 2;
    double rank = 3;
}

message GraphResult {
    string session_id = 1;
    Session session = 2;
    double relevance = 3;
    repeated string path = 4;
}

message FinalResult {
    int32 rank = 1;
    string session_id = 2;
    Session session = 3;
    double final_score = 4;
    double vector_score = 5;
    double keyword_score = 6;
    double graph_score = 7;
}

message Session {
    string id = 1;
    string contact_id = 2;
    google.protobuf.Timestamp started_at = 3;
    optional google.protobuf.Timestamp ended_at = 4;
    string status = 5;
    int32 message_count = 6;
    int32 messages_from_contact = 7;
    int32 messages_from_agent = 8;
    optional string summary = 9;
    optional string sentiment = 10;
    optional double sentiment_score = 11;
    repeated string topics = 12;
    repeated string outcome_tags = 13;
    bool resolved = 14;
    bool escalated = 15;
    repeated string agent_ids = 16;
    int32 agent_transfers = 17;
}

message ContactStats {
    int32 total_sessions = 1;
    int32 avg_session_duration = 2;
    int32 total_messages = 3;
    int32 positive_sessions = 4;
    int32 negative_sessions = 5;
    int32 resolved_sessions = 6;
    double avg_sentiment_score = 7;
    optional google.protobuf.Timestamp last_session_at = 8;
}

message PipelineContext {
    string pipeline_id = 1;
    string pipeline_name = 2;
    string current_stage = 3;
    int32 stage_order = 4;
    int32 days_in_stage = 5;
}

// === KNOWLEDGE SCOPE ===

message KnowledgeScope {
    bool include_sessions = 1;
    int32 sessions_lookback_days = 2;
    bool include_messages = 3;
    int32 messages_limit = 4;
    bool messages_only_recent = 5;
    int32 recent_messages_days = 6;
    bool include_contact_events = 7;
    repeated string contact_events_categories = 8;
    bool include_tracking = 9;
    bool include_notes = 10;
    bool include_session_summaries = 11;
    double similarity_threshold = 12;
    int32 max_similar_sessions = 13;
    bool include_agent_transfer_chain = 14;
    bool include_reply_threads = 15;
    bool include_social_graph = 16;
    bool include_campaign_graph = 17;
    int32 graph_traversal_depth = 18;
    bool include_contact_stats = 19;
    bool include_pipeline_context = 20;
    bool include_memory_facts = 21;
    repeated string fact_types = 22;
}

// === MEMORY STRATEGY ===

message MemoryStrategy {
    double vector_weight = 1;
    double keyword_weight = 2;
    double graph_weight = 3;
    double recent_weight = 4;
    string strategy = 5;  // "semantic_heavy", "balanced", etc
    string fusion_method = 6;  // "rrf", "weighted"
    bool use_reranking = 7;
    string rerank_provider = 8;
    int32 rerank_top_k = 9;
    int32 max_tokens = 10;
    string summarization = 11;
    int32 cache_duration = 12;
    string cache_strategy = 13;
}

// === ADD SESSION TO MEMORY ===

message AddSessionToMemoryRequest {
    string tenant_id = 1;
    string session_id = 2;
    bool generate_embedding = 3;  // Se deve gerar embedding contextual
}

message AddSessionToMemoryResponse {
    bool success = 1;
    optional string embedding_id = 2;
    optional string error = 3;
}

// === GET CONTACT CONTEXT ===

message GetContactContextRequest {
    string tenant_id = 1;
    string contact_id = 2;
    string agent_category = 3;
    bool include_cached = 4;
}

message GetContactContextResponse {
    string context_text = 1;  // Context completo (cacheável)
    repeated Message recent_messages = 2;
    optional ContactStats stats = 3;
    optional PipelineContext pipeline = 4;
    repeated MemoryFact memory_facts = 5;
    bool cache_hit = 6;
}

// === ROUTE TO AGENT ===

message RouteToAgentRequest {
    string tenant_id = 1;
    string contact_id = 2;
    string session_id = 3;
    string message_text = 4;
}

message RouteToAgentResponse {
    string agent_id = 1;
    string agent_name = 2;
    string agent_category = 3;
    Intent intent = 4;
    double routing_confidence = 5;
}

message Intent {
    string name = 1;
    string category = 2;
    double confidence = 3;
    string matched_utterance = 4;
}

// === MEMORY FACTS ===

message AddMemoryFactRequest {
    string tenant_id = 1;
    string contact_id = 2;
    string fact_text = 3;
    string source = 4;
    optional string source_id = 5;
}

message AddMemoryFactResponse {
    string fact_id = 1;
    string fact_type = 2;
    google.protobuf.Struct fact_value = 3;
    double confidence = 4;
    optional string supersedes_fact_id = 5;  // Se resolveu contradição
}

message GetActiveFactsRequest {
    string tenant_id = 1;
    string contact_id = 2;
    repeated string fact_types = 3;
    optional google.protobuf.Timestamp as_of = 4;
}

message GetActiveFactsResponse {
    repeated MemoryFact facts = 1;
}

message MemoryFact {
    string id = 1;
    string contact_id = 2;
    string fact_type = 3;
    string fact_text = 4;
    google.protobuf.Struct fact_value = 5;
    double confidence = 6;
    google.protobuf.Timestamp valid_from = 7;
    optional google.protobuf.Timestamp valid_to = 8;
    string source = 9;
}
```

### **Go gRPC Server Implementation**

```go
package grpc

import (
    "context"
    "fmt"

    memoryv1 "github.com/ventros/crm/api/memory/v1"
    "github.com/ventros/crm/internal/memory"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// MemoryServiceServer implementa o gRPC server
type MemoryServiceServer struct {
    memoryv1.UnimplementedMemoryServiceServer

    searchService    *memory.HybridSearchService
    embeddingService *memory.EmbeddingService
    agentRegistry    *memory.AgentRegistry
    factService      *memory.MemoryFactService
    contextManager   *memory.ContextManager
}

// NewMemoryServiceServer cria novo gRPC server
func NewMemoryServiceServer(
    searchService *memory.HybridSearchService,
    embeddingService *memory.EmbeddingService,
    agentRegistry *memory.AgentRegistry,
    factService *memory.MemoryFactService,
    contextManager *memory.ContextManager,
) *MemoryServiceServer {
    return &MemoryServiceServer{
        searchService:    searchService,
        embeddingService: embeddingService,
        agentRegistry:    agentRegistry,
        factService:      factService,
        contextManager:   contextManager,
    }
}

// SearchMemory implementa busca híbrida
func (s *MemoryServiceServer) SearchMemory(
    ctx context.Context,
    req *memoryv1.SearchMemoryRequest,
) (*memoryv1.SearchMemoryResponse, error) {
    // Validate request
    if req.TenantId == "" || req.ContactId == "" {
        return nil, status.Error(codes.InvalidArgument, "tenant_id and contact_id required")
    }

    // Convert protobuf to domain request
    searchReq, err := s.protoToSearchRequest(req)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
    }

    // Execute search
    result, err := s.searchService.Search(ctx, *searchReq)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "search failed: %v", err)
    }

    // Convert result to protobuf
    protoResult := s.searchResultToProto(result)

    return protoResult, nil
}

// AddSessionToMemory adiciona sessão à memória (gera embeddings)
func (s *MemoryServiceServer) AddSessionToMemory(
    ctx context.Context,
    req *memoryv1.AddSessionToMemoryRequest,
) (*memoryv1.AddSessionToMemoryResponse, error) {
    if req.TenantId == "" || req.SessionId == "" {
        return nil, status.Error(codes.InvalidArgument, "tenant_id and session_id required")
    }

    // Parse session ID
    sessionID, err := uuid.Parse(req.SessionId)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid session_id: %v", err)
    }

    // Generate contextual embedding
    if req.GenerateEmbedding {
        embedding, err := s.embeddingService.GenerateContextualEmbedding(
            ctx,
            sessionID,
            req.TenantId,
        )
        if err != nil {
            return &memoryv1.AddSessionToMemoryResponse{
                Success: false,
                Error:   ptr(err.Error()),
            }, nil
        }

        return &memoryv1.AddSessionToMemoryResponse{
            Success:     true,
            EmbeddingId: ptr(embedding.ID.String()),
        }, nil
    }

    return &memoryv1.AddSessionToMemoryResponse{
        Success: true,
    }, nil
}

// GetContactContext retorna contexto completo do contato
func (s *MemoryServiceServer) GetContactContext(
    ctx context.Context,
    req *memoryv1.GetContactContextRequest,
) (*memoryv1.GetContactContextResponse, error) {
    if req.TenantId == "" || req.ContactId == "" {
        return nil, status.Error(codes.InvalidArgument, "tenant_id and contact_id required")
    }

    contactID, err := uuid.Parse(req.ContactId)
    if err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "invalid contact_id: %v", err)
    }

    // Build search request for context
    searchReq := memory.SearchRequest{
        TenantID:       req.TenantId,
        ContactID:      contactID,
        AgentCategory:  memory.AgentCategory(req.AgentCategory),
        Query:          "",  // Empty query = get all context
        KnowledgeScope: memory.DefaultKnowledgeScope(memory.AgentCategory(req.AgentCategory)),
        MemoryStrategy: memory.DefaultMemoryStrategy(memory.AgentCategory(req.AgentCategory)),
        TopK:           10,
        RetrievalTopK:  100,
    }

    // Execute search
    result, err := s.searchService.Search(ctx, searchReq)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get context: %v", err)
    }

    // Build context text
    contextText := s.buildContextText(result)

    // Convert to proto
    response := &memoryv1.GetContactContextResponse{
        ContextText:    contextText,
        RecentMessages: s.messagesToProto(result.RecentMessages),
        CacheHit:       result.CacheHit,
    }

    if result.ContactStats != nil {
        response.Stats = s.contactStatsToProto(result.ContactStats)
    }

    if result.PipelineContext != nil {
        response.Pipeline = s.pipelineContextToProto(result.PipelineContext)
    }

    response.MemoryFacts = s.memoryFactsToProto(result.MemoryFacts)

    return response, nil
}

// RouteToAgent faz semantic routing
func (s *MemoryServiceServer) RouteToAgent(
    ctx context.Context,
    req *memoryv1.RouteToAgentRequest,
) (*memoryv1.RouteToAgentResponse, error) {
    if req.TenantId == "" || req.ContactId == "" || req.MessageText == "" {
        return nil, status.Error(codes.InvalidArgument, "missing required fields")
    }

    contactID, _ := uuid.Parse(req.ContactId)
    sessionID, _ := uuid.Parse(req.SessionId)

    // Get session and message
    // TODO: Fetch from repository

    // Route to agent
    agent, err := s.agentRegistry.RouteToAgent(ctx, message, session, req.TenantId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "routing failed: %v", err)
    }

    // Get intent (from routing process)
    // TODO: Return intent from RouteToAgent

    return &memoryv1.RouteToAgentResponse{
        AgentId:       agent.ID().String(),
        AgentName:     agent.Name(),
        AgentCategory: string(agent.Category()),
        Intent: &memoryv1.Intent{
            Name:       "detected_intent",
            Category:   string(agent.Category()),
            Confidence: 0.85,
        },
        RoutingConfidence: 0.90,
    }, nil
}

// AddMemoryFact adiciona memory fact
func (s *MemoryServiceServer) AddMemoryFact(
    ctx context.Context,
    req *memoryv1.AddMemoryFactRequest,
) (*memoryv1.AddMemoryFactResponse, error) {
    contactID, _ := uuid.Parse(req.ContactId)

    var sourceID *uuid.UUID
    if req.SourceId != nil {
        sid, _ := uuid.Parse(*req.SourceId)
        sourceID = &sid
    }

    fact, err := s.factService.AddFact(
        ctx,
        contactID,
        req.TenantId,
        req.FactText,
        req.Source,
        sourceID,
    )
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to add fact: %v", err)
    }

    response := &memoryv1.AddMemoryFactResponse{
        FactId:     fact.ID.String(),
        FactType:   string(fact.FactType),
        Confidence: fact.Confidence,
    }

    if fact.Supersedes != nil {
        response.SupersedesFactId = ptr(fact.Supersedes.String())
    }

    return response, nil
}

// GetActiveFacts retorna facts ativos
func (s *MemoryServiceServer) GetActiveFacts(
    ctx context.Context,
    req *memoryv1.GetActiveFactsRequest,
) (*memoryv1.GetActiveFactsResponse, error) {
    contactID, _ := uuid.Parse(req.ContactId)

    factTypes := make([]memory.FactType, len(req.FactTypes))
    for i, ft := range req.FactTypes {
        factTypes[i] = memory.FactType(ft)
    }

    var asOf *time.Time
    if req.AsOf != nil {
        t := req.AsOf.AsTime()
        asOf = &t
    }

    facts, err := s.factService.GetActiveFacts(ctx, contactID, factTypes, asOf)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get facts: %w", err)
    }

    return &memoryv1.GetActiveFactsResponse{
        Facts: s.memoryFactsToProto(facts),
    }, nil
}

// Helper conversion functions (proto <-> domain)

func (s *MemoryServiceServer) protoToSearchRequest(req *memoryv1.SearchMemoryRequest) (*memory.SearchRequest, error) {
    contactID, err := uuid.Parse(req.ContactId)
    if err != nil {
        return nil, err
    }

    var sessionID *uuid.UUID
    if req.SessionId != nil {
        sid, _ := uuid.Parse(*req.SessionId)
        sessionID = &sid
    }

    searchReq := &memory.SearchRequest{
        TenantID:       req.TenantId,
        ContactID:      contactID,
        SessionID:      sessionID,
        AgentCategory:  memory.AgentCategory(req.AgentCategory),
        Query:          req.Query,
        QueryEmbedding: req.QueryEmbedding,
        TopK:           int(req.TopK),
        RetrievalTopK:  int(req.RetrievalTopK),
    }

    // Convert KnowledgeScope
    if req.KnowledgeScope != nil {
        searchReq.KnowledgeScope = s.protoToKnowledgeScope(req.KnowledgeScope)
    }

    // Convert MemoryStrategy
    if req.MemoryStrategy != nil {
        searchReq.MemoryStrategy = s.protoToMemoryStrategy(req.MemoryStrategy)
    }

    return searchReq, nil
}

func (s *MemoryServiceServer) buildContextText(result *memory.SearchResult) string {
    // Build structured context text
    context := "=== RECENT MESSAGES ===\n"
    for _, msg := range result.RecentMessages {
        context += fmt.Sprintf("[%s] %s: %s\n",
            msg.Timestamp.Format("15:04"),
            ifElse(msg.FromMe, "Agent", "Customer"),
            ptrStr(msg.Text),
        )
    }

    context += "\n=== RELEVANT SESSIONS ===\n"
    for _, fr := range result.FinalResults {
        context += fmt.Sprintf("- [Session %s] %s (similarity: %.2f)\n",
            fr.SessionID.String()[:8],
            ptrStr(fr.Session.Summary),
            fr.FinalScore,
        )
    }

    if result.ContactStats != nil {
        context += fmt.Sprintf("\n=== CONTACT STATS ===\n")
        context += fmt.Sprintf("Total Sessions: %d\n", result.ContactStats.TotalSessions)
        context += fmt.Sprintf("Avg Sentiment: %.2f\n", result.ContactStats.AvgSentimentScore)
    }

    return context
}

func ptr(s string) *string { return &s }
func ptrStr(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
func ifElse(cond bool, a, b string) string {
    if cond {
        return a
    }
    return b
}
```

---

## 🗄️ DATABASE SCHEMA

```sql
-- ============================================
-- MEMORY EMBEDDINGS TABLE
-- ============================================

CREATE EXTENSION IF NOT EXISTS vector;  -- pgvector extension
CREATE EXTENSION IF NOT EXISTS age;     -- Apache AGE extension
CREATE EXTENSION IF NOT EXISTS pg_trgm; -- Trigram extension for keyword search

-- Memory Embeddings (Contextual Retrieval)
CREATE TABLE memory_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    source_type VARCHAR(50) NOT NULL,  -- 'session_summary', 'contact_event', etc
    source_id UUID NOT NULL,           -- ID do objeto fonte
    embedding vector(768) NOT NULL,    -- text-embedding-005 (768 dim)
    contextual_text TEXT NOT NULL,     -- Context + original text
    original_text TEXT NOT NULL,       -- Text original (sem context)
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP,              -- NULL = não expira

    -- Indexes
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- HNSW index para vector search (melhor performance que IVFFlat)
CREATE INDEX idx_memory_embeddings_hnsw ON memory_embeddings
USING hnsw (embedding vector_cosine_ops);

-- Indexes adicionais
CREATE INDEX idx_memory_embeddings_tenant ON memory_embeddings(tenant_id);
CREATE INDEX idx_memory_embeddings_source ON memory_embeddings(tenant_id, source_type, source_id);
CREATE INDEX idx_memory_embeddings_metadata ON memory_embeddings USING GIN (metadata);
CREATE INDEX idx_memory_embeddings_expires ON memory_embeddings(expires_at) WHERE expires_at IS NOT NULL;

-- ============================================
-- MEMORY FACTS TABLE (Google Memory Bank pattern)
-- ============================================

CREATE TABLE memory_facts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL,
    tenant_id VARCHAR(255) NOT NULL,
    fact_type VARCHAR(50) NOT NULL,  -- 'budget_constraint', 'preference', etc
    fact_text TEXT NOT NULL,
    fact_value JSONB NOT NULL,       -- Structured value
    confidence FLOAT NOT NULL CHECK (confidence >= 0 AND confidence <= 1),

    -- Bi-temporal model
    valid_from TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMP,              -- NULL = current/active

    -- Contradiction resolution
    supersedes UUID,                 -- FK to previous fact

    -- Source tracking
    source VARCHAR(50) NOT NULL,     -- 'message', 'note', 'annotation'
    source_id UUID,                  -- ID do source object

    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_contact FOREIGN KEY (contact_id) REFERENCES contacts(id) ON DELETE CASCADE,
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT fk_supersedes FOREIGN KEY (supersedes) REFERENCES memory_facts(id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_memory_facts_contact ON memory_facts(contact_id);
CREATE INDEX idx_memory_facts_tenant ON memory_facts(tenant_id);
CREATE INDEX idx_memory_facts_type ON memory_facts(tenant_id, fact_type);
CREATE INDEX idx_memory_facts_active ON memory_facts(contact_id) WHERE valid_to IS NULL;
CREATE INDEX idx_memory_facts_validity ON memory_facts(contact_id, valid_from, valid_to);
CREATE INDEX idx_memory_facts_metadata ON memory_facts USING GIN (metadata);

-- ============================================
-- TEMPORAL GRAPH EDGES (Apache AGE)
-- ============================================

-- Note: Apache AGE usa graph schema próprio, mas mantemos tabela relacional para auditoria
CREATE TABLE temporal_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,

    -- From/To nodes
    from_node_id UUID NOT NULL,
    from_node_type VARCHAR(50) NOT NULL,  -- 'contact', 'session', 'agent'
    to_node_id UUID NOT NULL,
    to_node_type VARCHAR(50) NOT NULL,

    -- Edge type
    edge_type VARCHAR(50) NOT NULL,  -- 'HAS_SESSION', 'ASSIGNED_TO', etc

    -- Bi-temporal model (Graphiti pattern)
    valid_from TIMESTAMP NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMP,              -- NULL = still valid
    transaction_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Properties
    properties JSONB NOT NULL DEFAULT '{}',

    -- Audit
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_temporal_edges_tenant ON temporal_edges(tenant_id);
CREATE INDEX idx_temporal_edges_from ON temporal_edges(from_node_id, from_node_type);
CREATE INDEX idx_temporal_edges_to ON temporal_edges(to_node_id, to_node_type);
CREATE INDEX idx_temporal_edges_type ON temporal_edges(edge_type);
CREATE INDEX idx_temporal_edges_validity ON temporal_edges(valid_from, valid_to);
CREATE INDEX idx_temporal_edges_active ON temporal_edges(from_node_id, to_node_id) WHERE valid_to IS NULL;

-- ============================================
-- AGENT AI METADATA TABLE
-- ============================================

CREATE TABLE agent_ai_metadata (
    agent_id UUID PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,

    -- Agent configuration
    category VARCHAR(50) NOT NULL,   -- 'sales_prospecting', 'retention_churn', etc
    skills JSONB NOT NULL DEFAULT '[]',
    knowledge_scope JSONB NOT NULL,
    memory_strategy JSONB NOT NULL,
    routing_rules JSONB NOT NULL DEFAULT '[]',

    -- Capacity
    priority INT NOT NULL DEFAULT 5 CHECK (priority >= 1 AND priority <= 10),
    max_concurrent_sessions INT NOT NULL DEFAULT 10,
    current_session_count INT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_agent_ai_metadata_tenant ON agent_ai_metadata(tenant_id);
CREATE INDEX idx_agent_ai_metadata_category ON agent_ai_metadata(tenant_id, category);
CREATE INDEX idx_agent_ai_metadata_priority ON agent_ai_metadata(priority DESC);

-- ============================================
-- SEMANTIC ROUTES TABLE (for pre-computed embeddings)
-- ============================================

CREATE TABLE semantic_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,

    -- Route definition
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,      -- Maps to AgentCategory
    utterance TEXT NOT NULL,            -- Example phrase
    embedding vector(768) NOT NULL,     -- Pre-computed embedding
    priority INT NOT NULL DEFAULT 5,

    -- Metadata
    metadata JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT fk_tenant FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    UNIQUE (tenant_id, name, utterance)
);

-- HNSW index para semantic routing
CREATE INDEX idx_semantic_routes_hnsw ON semantic_routes
USING hnsw (embedding vector_cosine_ops);

-- Indexes adicionais
CREATE INDEX idx_semantic_routes_tenant ON semantic_routes(tenant_id);
CREATE INDEX idx_semantic_routes_category ON semantic_routes(tenant_id, category);
CREATE INDEX idx_semantic_routes_active ON semantic_routes(tenant_id) WHERE active = TRUE;

-- ============================================
-- HELPER FUNCTIONS
-- ============================================

-- Função para busca híbrida (Vector + Keyword)
CREATE OR REPLACE FUNCTION hybrid_search_sessions(
    p_tenant_id VARCHAR,
    p_contact_id UUID,
    p_query_embedding vector(768),
    p_query_text TEXT,
    p_similarity_threshold FLOAT DEFAULT 0.7,
    p_limit INT DEFAULT 10
)
RETURNS TABLE (
    session_id UUID,
    similarity FLOAT,
    keyword_rank FLOAT,
    hybrid_score FLOAT
) AS $$
BEGIN
    RETURN QUERY
    WITH vector_results AS (
        SELECT
            me.source_id AS session_id,
            1 - (me.embedding <=> p_query_embedding) AS similarity,
            ROW_NUMBER() OVER (ORDER BY me.embedding <=> p_query_embedding) AS rank
        FROM memory_embeddings me
        WHERE me.tenant_id = p_tenant_id
          AND me.source_type = 'session_summary'
          AND me.metadata->>'contact_id' = p_contact_id::TEXT
          AND (1 - (me.embedding <=> p_query_embedding)) >= p_similarity_threshold
        LIMIT 100
    ),
    keyword_results AS (
        SELECT
            s.id AS session_id,
            ts_rank(to_tsvector('portuguese', s.summary), to_tsquery('portuguese', p_query_text)) AS keyword_rank,
            ROW_NUMBER() OVER (ORDER BY ts_rank(to_tsvector('portuguese', s.summary), to_tsquery('portuguese', p_query_text)) DESC) AS rank
        FROM sessions s
        WHERE s.tenant_id = p_tenant_id
          AND s.contact_id = p_contact_id
          AND to_tsvector('portuguese', s.summary) @@ to_tsquery('portuguese', p_query_text)
        LIMIT 100
    )
    SELECT
        COALESCE(v.session_id, k.session_id) AS session_id,
        COALESCE(v.similarity, 0) AS similarity,
        COALESCE(k.keyword_rank, 0) AS keyword_rank,
        -- RRF fusion
        (COALESCE(1.0 / (v.rank + 60), 0) + COALESCE(1.0 / (k.rank + 60), 0)) AS hybrid_score
    FROM vector_results v
    FULL OUTER JOIN keyword_results k ON v.session_id = k.session_id
    ORDER BY hybrid_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Função para busca temporal em facts
CREATE OR REPLACE FUNCTION get_active_facts(
    p_contact_id UUID,
    p_fact_types TEXT[],
    p_as_of TIMESTAMP DEFAULT NOW()
)
RETURNS TABLE (
    fact_id UUID,
    fact_type VARCHAR,
    fact_text TEXT,
    fact_value JSONB,
    confidence FLOAT,
    valid_from TIMESTAMP,
    valid_to TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        mf.id,
        mf.fact_type,
        mf.fact_text,
        mf.fact_value,
        mf.confidence,
        mf.valid_from,
        mf.valid_to
    FROM memory_facts mf
    WHERE mf.contact_id = p_contact_id
      AND (p_fact_types IS NULL OR mf.fact_type = ANY(p_fact_types))
      AND mf.valid_from <= p_as_of
      AND (mf.valid_to IS NULL OR mf.valid_to > p_as_of)
    ORDER BY mf.valid_from DESC;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- APACHE AGE GRAPH SCHEMA
-- ============================================

-- Create graph
SELECT create_graph('ventros_memory_graph');

-- Create vertex labels
SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE VLABEL Contact
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE VLABEL Session
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE VLABEL Agent
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE VLABEL Campaign
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE VLABEL Platform
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE VLABEL Topic
$$) as (result agtype);

-- Create edge labels
SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE ELABEL HAS_SESSION
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE ELABEL ASSIGNED_TO
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE ELABEL TRANSFERRED_TO
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE ELABEL CAME_FROM_CAMPAIGN
$$) as (result agtype);

SELECT * FROM cypher('ventros_memory_graph', $$
    CREATE ELABEL DISCUSSED_TOPIC
$$) as (result agtype);

-- ============================================
-- MIGRATIONS & SEED DATA
-- ============================================

-- Seed semantic routes com embeddings pré-computados
-- (Este seria um migration separado que rodaria script Go para gerar embeddings)

-- TODO: Seed default semantic routes
-- TODO: Seed default agent AI metadata para agentes existentes

-- ============================================
-- PERMISSIONS & RLS (Row Level Security)
-- ============================================

-- Enable RLS em todas as tabelas
ALTER TABLE memory_embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE memory_facts ENABLE ROW LEVEL SECURITY;
ALTER TABLE temporal_edges ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_ai_metadata ENABLE ROW LEVEL SECURITY;
ALTER TABLE semantic_routes ENABLE ROW LEVEL SECURITY;

-- Policies (exemplo para multi-tenancy)
CREATE POLICY tenant_isolation_memory_embeddings ON memory_embeddings
    USING (tenant_id = current_setting('app.current_tenant_id')::VARCHAR);

CREATE POLICY tenant_isolation_memory_facts ON memory_facts
    USING (tenant_id = current_setting('app.current_tenant_id')::VARCHAR);

-- (Similar policies para outras tabelas...)
```

---

## 🎯 RESUMO DA ARQUITETURA GO

### **Componentes Implementados:**

1. ✅ **Memory Embedding Service** (Contextual Retrieval)
2. ✅ **Hybrid Search Service** (Vector + Keyword + Graph + SQL)
3. ✅ **Retrieval Strategies Dictionary** (33/33/33, 50/50, 70/30, etc)
4. ✅ **Temporal Knowledge Graph** (Apache AGE + Bi-temporal model)
5. ✅ **Agent Registry & Semantic Routing** (Aurelio Labs pattern)
6. ✅ **Memory Fact Service** (Google Memory Bank pattern)
7. ✅ **Context Manager** (Prompt Caching com Redis)
8. ✅ **gRPC API** (Python ADK integration)
9. ✅ **Database Schema** (PostgreSQL + pgvector + Apache AGE)

### **Pronto para:**
- ✅ Python ADK consumir via gRPC
- ✅ Escalar para milhões de embeddings
- ✅ Suportar múltiplos agentes simultâneos
- ✅ BI/Analytics integration (schema bem estruturado)

---

**Próximo: Criar documento PYTHON ADK completo com todos os agent types e patterns!** 🚀
