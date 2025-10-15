# VENTROS CRM - INTEGRATION GUIDE

> **Como Ventros CRM se integra com Python ADK, Memory Service e MCP Server**
>
> **Última atualização**: 2025-10-15

---

## 🎯 VISÃO GERAL

Ventros CRM é o **ORQUESTRADOR CENTRAL** de todo o ecossistema Ventros:

```
┌─────────────────────────────────────────────────────────────────┐
│                     VENTROS CRM (Go)                            │
│                     ORQUESTRADOR CENTRAL                        │
│                                                                   │
│  - Persiste TODOS os dados (PostgreSQL)                         │
│  - Orquestra TODAS as operações                                 │
│  - Expõe API REST (port 8080) para frontend                     │
│  - Expõe MCP Server (port 8081) para Python ADK                 │
│  - Contém Memory Service (embedded)                             │
│  - Gerencia multi-tenancy, auth, RBAC                           │
└───────────────────────┬───────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
   Frontend      Python ADK      External Services
   (React)       (Microservice)  (WAHA, Stripe, etc)
```

---

## 🔗 INTEGRAÇÃO 1: Ventros CRM ↔ Python ADK

### **Arquitetura**

```
┌─────────────────────────────────────────────────────────────────┐
│                     VENTROS CRM (Go)                            │
│                                                                   │
│  1. Cliente envia mensagem no WhatsApp                          │
│  2. WAHA webhook → Go CRM                                       │
│  3. Go persiste message + dispara enrichment                    │
│  4. Go decide: "Precisa de AI Agent?"                           │
│     ↓                                                            │
│  5. Go → Python ADK (gRPC - futuro):                            │
│     ExecuteAgent(type="CustomerService", context={...})         │
└───────────────────────┬─────────────────────────────────────────┘
                        │
                        │ gRPC (Port 50051)
                        │ Request: AgentExecutionRequest
                        │ Response: AgentExecutionResponse
                        │
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                     PYTHON ADK (Python)                         │
│                     Microserviço separado                       │
│                                                                   │
│  6. Recebe request de execução                                  │
│  7. Inicializa CustomerServiceAgent                             │
│  8. CustomerServiceAgent chama sub-agents (6 agents em cadeia): │
│     - LeadQualifierAgent                                        │
│     - PricingAgent                                              │
│     - ProposalAgent                                             │
│     - ApprovalAgent                                             │
│     - ResponseGeneratorAgent                                    │
│     - Total: 5-10 segundos de execução                          │
│                                                                   │
│  9. Durante execução, Python chama Go (MCP Tools):              │
│     ↓                                                            │
│     HTTP POST http://ventros-crm:8081/v1/mcp/execute            │
│     Body: {                                                      │
│       "tool_name": "get_contact",                               │
│       "arguments": {"contact_id": "uuid"}                       │
│     }                                                            │
│     ← Response: {"tool_name": "...", "result": {...}}           │
│                                                                   │
│  10. Python retorna response → Go CRM                           │
└───────────────────────┬─────────────────────────────────────────┘
                        │
                        │ gRPC Response
                        │
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                     VENTROS CRM (Go)                            │
│                                                                   │
│  11. Recebe response do Python ADK                              │
│  12. Persiste response no PostgreSQL                            │
│  13. Envia mensagem via WAHA                                    │
│  14. Publica events: message.sent, agent.response_generated     │
└─────────────────────────────────────────────────────────────────┘
```

### **Protocolo gRPC (Futuro - 0% implementado)**

```protobuf
// internal/grpc/agent.proto

syntax = "proto3";

package ventros.agent;

service AgentService {
  // Execute agent with context
  rpc ExecuteAgent(AgentExecutionRequest) returns (AgentExecutionResponse);

  // Stream agent execution (for long-running agents)
  rpc ExecuteAgentStream(AgentExecutionRequest) returns (stream AgentExecutionEvent);

  // Get agent execution status
  rpc GetExecutionStatus(ExecutionStatusRequest) returns (ExecutionStatusResponse);

  // Cancel agent execution
  rpc CancelExecution(CancelExecutionRequest) returns (CancelExecutionResponse);
}

message AgentExecutionRequest {
  // Agent info
  string agent_type = 1;           // "CustomerService", "LeadQualifier", etc
  string execution_id = 2;         // UUID for tracking

  // Context
  string tenant_id = 3;
  string project_id = 4;
  string contact_id = 5;
  string session_id = 6;
  string message_id = 7;

  // Input
  string input_text = 10;
  map<string, string> metadata = 11;

  // Configuration
  int32 timeout_seconds = 20;      // Default: 30s
  bool streaming = 21;             // Enable streaming responses
}

message AgentExecutionResponse {
  // Execution info
  string execution_id = 1;
  string agent_type = 2;
  string status = 3;               // success, error, timeout

  // Output
  string response_text = 10;
  repeated AgentAction actions = 11;
  map<string, string> metadata = 12;

  // Stats
  int32 execution_time_ms = 20;
  int32 sub_agents_called = 21;
  repeated string sub_agent_types = 22;
  int32 mcp_tools_called = 23;
  repeated string mcp_tool_names = 24;

  // Cost
  int32 tokens_used = 30;
  double cost_usd = 31;

  // Error (if any)
  string error_message = 40;
  string error_code = 41;
}

message AgentAction {
  string action_type = 1;          // "update_contact", "create_event", etc
  string entity_type = 2;          // "contact", "session", "message"
  string entity_id = 3;
  map<string, string> data = 4;
}
```

### **Exemplo de Chamada (Go → Python)**

```go
// internal/application/agent/execute_agent_usecase.go

package agent

import (
	"context"
	"time"

	"google.golang.org/grpc"
	pb "github.com/ventros-crm/internal/grpc/proto"
)

type ExecuteAgentUseCase struct {
	agentClient pb.AgentServiceClient
}

func NewExecuteAgentUseCase(conn *grpc.ClientConn) *ExecuteAgentUseCase {
	return &ExecuteAgentUseCase{
		agentClient: pb.NewAgentServiceClient(conn),
	}
}

func (uc *ExecuteAgentUseCase) Execute(
	ctx context.Context,
	agentType string,
	contactID, sessionID, messageID string,
	inputText string,
) (*pb.AgentExecutionResponse, error) {
	// Build request
	req := &pb.AgentExecutionRequest{
		AgentType:      agentType,
		ExecutionId:    uuid.New().String(),
		TenantId:       ctx.Value("tenant_id").(string),
		ProjectId:      ctx.Value("project_id").(string),
		ContactId:      contactID,
		SessionId:      sessionID,
		MessageId:      messageID,
		InputText:      inputText,
		TimeoutSeconds: 30,
		Streaming:      false,
	}

	// Call Python ADK
	ctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()

	resp, err := uc.agentClient.ExecuteAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute agent: %w", err)
	}

	return resp, nil
}
```

### **Python ADK recebe e chama MCP Tools**

```python
# python-adk/ventros/agent_service.py

import grpc
from mcp_client import MCPClient
from ventros.proto import agent_pb2, agent_pb2_grpc

class AgentService(agent_pb2_grpc.AgentServiceServicer):
    def __init__(self, mcp_url: str, mcp_token: str):
        self.mcp_client = MCPClient(base_url=mcp_url, auth_token=mcp_token)

    def ExecuteAgent(self, request, context):
        # Initialize agent
        agent = self._get_agent(request.agent_type)

        # Build context for agent
        agent_context = {
            "tenant_id": request.tenant_id,
            "contact_id": request.contact_id,
            "session_id": request.session_id,
            "message_id": request.message_id,
        }

        # Agent needs contact data? Call MCP tool
        if request.contact_id:
            contact = self.mcp_client.call_tool("get_contact", {
                "contact_id": request.contact_id
            })
            agent_context["contact"] = contact["result"]

        # Agent needs conversation history? Call MCP tool
        messages = self.mcp_client.call_tool("get_messages", {
            "session_id": request.session_id,
            "limit": 50
        })
        agent_context["history"] = messages["result"]["messages"]

        # Agent needs documents? Call MCP tool
        documents = self.mcp_client.call_tool("search_documents", {
            "query": request.input_text,
            "contact_id": request.contact_id,
            "limit": 5
        })
        agent_context["documents"] = documents["result"]["documents"]

        # Execute agent (may call 6+ sub-agents internally)
        start_time = time.time()
        response = agent.execute(request.input_text, agent_context)
        execution_time_ms = int((time.time() - start_time) * 1000)

        # Return response to Go
        return agent_pb2.AgentExecutionResponse(
            execution_id=request.execution_id,
            agent_type=request.agent_type,
            status="success",
            response_text=response["text"],
            execution_time_ms=execution_time_ms,
            sub_agents_called=response["sub_agents_called"],
            mcp_tools_called=response["mcp_tools_called"],
            tokens_used=response["tokens_used"],
            cost_usd=response["cost_usd"],
        )
```

### **Estado da Implementação**

| Componente | Status | Prioridade |
|------------|--------|------------|
| gRPC Interface (Go) | ❌ 0% | P1 (Sprint 2) |
| gRPC Interface (Python) | ❌ 0% | P1 (Sprint 2) |
| AgentExecutionRequest/Response | ❌ 0% | P1 (Sprint 2) |
| MCP Client (Python) | ✅ 100% | Done |
| MCP Tools (Go) | ✅ 100% | Done |
| Agent Execution Tracking | ❌ 0% | P2 (Sprint 3) |
| Agent Performance Metrics | ❌ 0% | P2 (Sprint 3) |

---

## 🔗 INTEGRAÇÃO 2: Ventros CRM ↔ Memory Service

### **Arquitetura**

**IMPORTANTE**: Memory Service NÃO é um serviço separado. É uma **feature embedded** dentro do Ventros CRM.

```
┌─────────────────────────────────────────────────────────────────┐
│                     VENTROS CRM (Go)                            │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  MEMORY SERVICE (Embedded)                              │   │
│  │                                                          │   │
│  │  - pgvector embeddings (768 dimensions)                 │   │
│  │  - Hybrid search (vector + keyword + graph)             │   │
│  │  - RRF (Reciprocal Rank Fusion) + Cross-Encoder        │   │
│  │  - Knowledge graph (PostgreSQL)                         │   │
│  │  - Fact extraction & consolidation                      │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                   │
│  Exposto via MCP Tools (Port 8081):                             │
│  - search_documents (vector + keyword + filters)                │
│  - get_document (full document reconstruction)                  │
│  - get_document_references (find all docs mentioning entity)    │
│  - get_message_group (multimodal context)                       │
│  - search_memory (facts + embeddings)                           │
└─────────────────────────────────────────────────────────────────┘
```

### **Schema PostgreSQL**

```sql
-- Embeddings (vector storage)
CREATE TABLE memory_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    project_id UUID NOT NULL,

    -- Content
    content_type VARCHAR(50) NOT NULL,      -- document, message, session, fact
    content_subtype VARCHAR(50),            -- contract, invoice, audio, etc
    content_text TEXT NOT NULL,
    embedding vector(768) NOT NULL,         -- pgvector

    -- Metadata (JSONB for flexibility)
    metadata JSONB NOT NULL DEFAULT '{}',
    -- Examples:
    -- Document: {document_title, document_type, page_number, chunk_index, entities, references}
    -- Message: {message_id, session_id, contact_id, channel_id, sentiment}
    -- Fact: {fact_type, confidence, extracted_at, source_ids}

    -- References
    contact_id UUID,
    session_id UUID,
    message_id UUID,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Indexes
    INDEX idx_memory_embeddings_tenant (tenant_id),
    INDEX idx_memory_embeddings_type (content_type),
    INDEX idx_memory_embeddings_contact (contact_id),
    INDEX idx_memory_embeddings_session (session_id)
);

-- Vector index (HNSW for fast similarity search)
CREATE INDEX idx_memory_embeddings_vector
    ON memory_embeddings
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- GIN index for metadata JSONB queries
CREATE INDEX idx_memory_embeddings_metadata
    ON memory_embeddings
    USING GIN (metadata);

-- Facts (extracted knowledge)
CREATE TABLE memory_facts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    project_id UUID NOT NULL,

    -- Fact content
    fact_type VARCHAR(50) NOT NULL,         -- preference, constraint, attribute, event
    fact_text TEXT NOT NULL,                -- "João prefere café sem açúcar"
    fact_key VARCHAR(255) NOT NULL,         -- "preference.drink.coffee.sugar"
    fact_value TEXT,                        -- "without_sugar"

    -- Confidence
    confidence FLOAT NOT NULL DEFAULT 1.0,  -- 0.0 to 1.0
    source_count INT NOT NULL DEFAULT 1,    -- How many times observed

    -- References
    contact_id UUID NOT NULL,
    source_ids UUID[] NOT NULL,             -- Array of message/session IDs

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_observed_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE (tenant_id, contact_id, fact_key),

    -- Indexes
    INDEX idx_memory_facts_tenant (tenant_id),
    INDEX idx_memory_facts_contact (contact_id),
    INDEX idx_memory_facts_type (fact_type),
    INDEX idx_memory_facts_key (fact_key)
);

-- Knowledge graph (relationships)
CREATE TABLE memory_graph (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    project_id UUID NOT NULL,

    -- Nodes
    node_from_type VARCHAR(50) NOT NULL,    -- contact, company, product, etc
    node_from_id UUID NOT NULL,
    node_to_type VARCHAR(50) NOT NULL,
    node_to_id UUID NOT NULL,

    -- Relationship
    relationship_type VARCHAR(50) NOT NULL, -- works_at, manages, bought, interested_in
    relationship_strength FLOAT DEFAULT 1.0, -- 0.0 to 1.0

    -- Metadata
    metadata JSONB NOT NULL DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Constraints
    UNIQUE (tenant_id, node_from_type, node_from_id, node_to_type, node_to_id, relationship_type),

    -- Indexes
    INDEX idx_memory_graph_tenant (tenant_id),
    INDEX idx_memory_graph_from (node_from_type, node_from_id),
    INDEX idx_memory_graph_to (node_to_type, node_to_id),
    INDEX idx_memory_graph_relationship (relationship_type)
);
```

### **Hybrid Search Implementation (RRF + Reranker)**

```go
// internal/application/memory/hybrid_search.go

package memory

import (
	"context"
	"sort"
)

type HybridSearchService struct {
	db               *gorm.DB
	embeddingService EmbeddingService
	reranker         CrossEncoderReranker
}

type SearchMethod string

const (
	MethodVector       SearchMethod = "vector"         // pgvector cosine similarity
	MethodKeyword      SearchMethod = "keyword"        // PostgreSQL FTS
	MethodBM25         SearchMethod = "bm25"           // BM25 ranking
	MethodGraph        SearchMethod = "graph"          // Knowledge graph traversal
	MethodRecent       SearchMethod = "recent"         // Time-based
	MethodColBERT      SearchMethod = "colbert"        // ColBERT v2 (futuro)
	MethodHybridSparse SearchMethod = "hybrid_sparse"  // SPLADE (futuro)
)

type HybridSearchConfig struct {
	Methods  []SearchMethod           `json:"methods"`
	K        float64                  `json:"k"`           // RRF constant (default: 60)
	Reranker *CrossEncoderConfig      `json:"reranker"`
	LLMJudge *LLMJudgeConfig          `json:"llm_judge"`
}

func (s *HybridSearchService) Search(
	ctx context.Context,
	query string,
	config HybridSearchConfig,
	limit int,
) ([]SearchResult, error) {
	// 1. Execute each search method in parallel
	resultsChan := make(chan MethodResults, len(config.Methods))

	for _, method := range config.Methods {
		go func(m SearchMethod) {
			results := s.executeMethod(ctx, query, m, limit*2) // Get 2x for fusion
			resultsChan <- MethodResults{Method: m, Results: results}
		}(method)
	}

	// Collect results from all methods
	allMethodResults := make(map[SearchMethod][]SearchResult)
	for i := 0; i < len(config.Methods); i++ {
		methodResults := <-resultsChan
		allMethodResults[methodResults.Method] = methodResults.Results
	}

	// 2. Apply Reciprocal Rank Fusion (RRF)
	fusedResults := s.applyRRF(allMethodResults, config.K)

	// 3. Rerank with Cross-Encoder (if enabled)
	if config.Reranker != nil && config.Reranker.Enabled {
		rerankedResults, err := s.reranker.Rerank(ctx, query, fusedResults, config.Reranker)
		if err != nil {
			return nil, err
		}
		fusedResults = rerankedResults
	}

	// 4. LLM-as-Judge (if enabled, for complex cases)
	if config.LLMJudge != nil && config.LLMJudge.Enabled {
		judgedResults, err := s.llmJudge(ctx, query, fusedResults, config.LLMJudge)
		if err != nil {
			return nil, err
		}
		fusedResults = judgedResults
	}

	// 5. Return top-K
	if len(fusedResults) > limit {
		fusedResults = fusedResults[:limit]
	}

	return fusedResults, nil
}

// RRF Formula: score = Σ(1 / (k + rank_i)) for each method i
func (s *HybridSearchService) applyRRF(
	methodResults map[SearchMethod][]SearchResult,
	k float64,
) []SearchResult {
	// Aggregate scores by document ID
	docScores := make(map[string]float64)

	for method, results := range methodResults {
		for rank, result := range results {
			rrfScore := 1.0 / (k + float64(rank+1))
			docScores[result.ID] += rrfScore
		}
	}

	// Convert to slice and sort by score
	var fusedResults []SearchResult
	for docID, score := range docScores {
		// Get original result (from any method)
		var originalResult SearchResult
		for _, results := range methodResults {
			for _, r := range results {
				if r.ID == docID {
					originalResult = r
					break
				}
			}
		}

		originalResult.Score = score
		fusedResults = append(fusedResults, originalResult)
	}

	// Sort by RRF score (descending)
	sort.Slice(fusedResults, func(i, j int) bool {
		return fusedResults[i].Score > fusedResults[j].Score
	})

	return fusedResults
}
```

### **MCP Tools para Memory Service**

```go
// infrastructure/mcp/tools/memory_tools.go

// search_documents: Hybrid search (vector + keyword + filters)
func (s *MemoryService) SearchDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query := getStringArg(args, "query", "")
	contactID := getStringArg(args, "contact_id", "")
	limit := getIntArg(args, "limit", 10)

	// Configure hybrid search
	config := HybridSearchConfig{
		Methods: []SearchMethod{MethodVector, MethodKeyword, MethodBM25},
		K:       60.0,
		Reranker: &CrossEncoderConfig{
			Enabled:   true,
			Model:     "BAAI/bge-reranker-v2-m3",
			TopK:      100,
			FinalK:    limit,
			Threshold: 0.3,
		},
	}

	// Execute search
	results, err := s.hybridSearch.Search(ctx, query, config, limit)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"documents": results,
		"total":     len(results),
		"query":     query,
	}, nil
}

// search_memory: Search facts + embeddings
func (s *MemoryService) SearchMemory(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query := getStringArg(args, "query", "")
	contactID := getStringArg(args, "contact_id", "")
	memoryTypes := getStringArrayArg(args, "memory_types", []string{"fact", "document", "message"})

	// Search facts (keyword)
	var facts []MemoryFact
	s.db.Where("tenant_id = ? AND contact_id = ?", tenantID, contactID).
		Where("fact_text ILIKE ?", "%"+query+"%").
		Order("confidence DESC, last_observed_at DESC").
		Limit(10).
		Find(&facts)

	// Search embeddings (hybrid)
	embeddingResults, _ := s.SearchDocuments(ctx, map[string]interface{}{
		"query":       query,
		"contact_id":  contactID,
		"limit":       10,
	})

	return map[string]interface{}{
		"facts":      facts,
		"embeddings": embeddingResults,
		"query":      query,
	}, nil
}
```

### **Estado da Implementação**

| Componente | Status | Prioridade |
|------------|--------|------------|
| Schema PostgreSQL | ✅ 100% | Done |
| pgvector Extension | ✅ 100% | Done |
| Embedding Generation | ⏳ 50% | P1 (Sprint 1) |
| Vector Search | ❌ 0% | P1 (Sprint 1) |
| Hybrid Search (RRF) | ❌ 0% | P1 (Sprint 1) |
| Cross-Encoder Reranking | ❌ 0% | P1 (Sprint 1) |
| Knowledge Graph | ❌ 0% | P1 (Sprint 1) |
| Fact Extraction | ❌ 0% | P2 (Sprint 2) |
| MCP Tools | ✅ 100% | Done |

---

## 🔗 INTEGRAÇÃO 3: MCP Server (Embedded no CRM)

### **Arquitetura**

MCP Server é uma **feature embedded** dentro do Ventros CRM, NÃO um serviço separado.

```
┌─────────────────────────────────────────────────────────────────┐
│                     VENTROS CRM (Go)                            │
│                     Single Binary, Two Ports                    │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  MAIN API (Port 8080) - Gin Router                      │   │
│  │  - 158 endpoints REST                                    │   │
│  │  - WebSocket                                             │   │
│  │  - Webhooks (WAHA, Stripe)                              │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  MCP SERVER (Port 8081) - Chi Router                    │   │
│  │  - 30+ tools para Python ADK                            │   │
│  │  - Same database connection                             │   │
│  │  - Same repositories                                     │   │
│  │  - Same domain aggregates                               │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### **MCP Tools Disponíveis (30+)**

```go
// infrastructure/mcp/registry.go

var ToolRegistry = map[string]Tool{
	// CRM Operations (10)
	"get_contacts":       {Handler: crmService.GetContacts, Description: "List contacts"},
	"get_contact":        {Handler: crmService.GetContact, Description: "Get specific contact"},
	"get_contact_lists":  {Handler: crmService.GetContactLists, Description: "List contact lists"},
	"get_list_contacts":  {Handler: crmService.GetListContacts, Description: "Get contacts in list"},
	"get_pipelines":      {Handler: crmService.GetPipelines, Description: "List pipelines"},
	"get_channels":       {Handler: crmService.GetChannels, Description: "List channels"},
	"get_agents":         {Handler: crmService.GetAgents, Description: "List agents"},
	"get_sessions":       {Handler: crmService.GetSessions, Description: "List sessions"},
	"get_messages":       {Handler: crmService.GetMessages, Description: "List messages"},
	"get_contact_events": {Handler: crmService.GetContactEvents, Description: "Get contact timeline"},

	// Multimodal Context (2)
	"get_message_group":   {Handler: msgGroupService.GetMessageGroup, Description: "Get message group with enrichments"},
	"list_message_groups": {Handler: msgGroupService.ListMessageGroups, Description: "List message groups"},

	// Document Operations (3)
	"search_documents":        {Handler: docService.SearchDocuments, Description: "Hybrid search documents"},
	"get_document":            {Handler: docService.GetDocument, Description: "Get full document"},
	"get_document_references": {Handler: docService.GetDocumentReferences, Description: "Find docs mentioning entity"},

	// Memory Operations (2)
	"search_memory":    {Handler: memoryService.SearchMemory, Description: "Search facts + embeddings"},
	"get_contact_facts": {Handler: memoryService.GetContactFacts, Description: "Get extracted facts"},

	// BI & Analytics (5)
	"get_leads_count":        {Handler: biService.GetLeadsCount, Description: "Count leads by filter"},
	"get_agent_stats":        {Handler: biService.GetAgentStats, Description: "Get agent performance"},
	"get_top_agent":          {Handler: biService.GetTopAgent, Description: "Get best performing agent"},
	"analyze_agent_messages": {Handler: biService.AnalyzeAgentMessages, Description: "Analyze agent messages"},
	"compare_agents":         {Handler: biService.CompareAgents, Description: "Compare agent performance"},

	// CRM Mutations (8)
	"update_contact":          {Handler: mutationService.UpdateContact, Description: "Update contact"},
	"update_pipeline_stage":   {Handler: mutationService.UpdatePipelineStage, Description: "Move contact in pipeline"},
	"assign_to_agent":         {Handler: mutationService.AssignToAgent, Description: "Assign contact to agent"},
	"qualify_lead":            {Handler: mutationService.QualifyLead, Description: "Qualify contact as lead"},
	"create_note":             {Handler: mutationService.CreateNote, Description: "Add note to contact"},
	"add_tag":                 {Handler: mutationService.AddTag, Description: "Add tag to contact"},
	"add_to_list":             {Handler: mutationService.AddToList, Description: "Add contact to list"},
	"send_message":            {Handler: mutationService.SendMessage, Description: "Send message to contact"},
}
```

### **Exemplo: Python ADK chama MCP Tool**

```python
# python-adk/ventros/mcp_client.py

import requests
from typing import Dict, Any

class MCPClient:
    def __init__(self, base_url: str, auth_token: str):
        self.base_url = base_url
        self.headers = {
            "Authorization": f"Bearer {auth_token}",
            "Content-Type": "application/json",
        }

    def call_tool(self, tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """Call MCP tool"""
        url = f"{self.base_url}/v1/mcp/execute"

        payload = {
            "tool_name": tool_name,
            "arguments": arguments,
        }

        response = requests.post(url, json=payload, headers=self.headers, timeout=30)
        response.raise_for_status()

        return response.json()

    def list_tools(self) -> Dict[str, Any]:
        """List available tools"""
        url = f"{self.base_url}/v1/mcp/tools"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()


# Example usage in agent
class CustomerServiceAgent:
    def __init__(self, mcp_client: MCPClient):
        self.mcp = mcp_client

    def execute(self, input_text: str, context: Dict[str, Any]) -> Dict[str, Any]:
        # Get contact data
        contact = self.mcp.call_tool("get_contact", {
            "contact_id": context["contact_id"]
        })

        # Get conversation history
        messages = self.mcp.call_tool("get_messages", {
            "session_id": context["session_id"],
            "limit": 50
        })

        # Search documents
        documents = self.mcp.call_tool("search_documents", {
            "query": input_text,
            "contact_id": context["contact_id"],
            "limit": 5
        })

        # Search memory facts
        memory = self.mcp.call_tool("search_memory", {
            "query": input_text,
            "contact_id": context["contact_id"]
        })

        # Generate response using LLM with context
        response = self.llm.generate(
            input_text=input_text,
            context={
                "contact": contact["result"],
                "history": messages["result"]["messages"],
                "documents": documents["result"]["documents"],
                "memory": memory["result"],
            }
        )

        return {
            "text": response,
            "mcp_tools_called": 4,
            "tokens_used": response.tokens_used,
        }
```

---

## 🔄 FLUXO COMPLETO: Cliente → Resposta AI

```
1. 👤 Cliente envia mensagem no WhatsApp
   "Qual o valor do contrato que enviamos?"
   ↓

2. 📱 WAHA (WhatsApp API) recebe e envia webhook
   POST http://ventros-crm:8080/webhooks/waha
   Body: {
     message: {
       from: "5511999999999",
       text: "Qual o valor do contrato que enviamos?",
       type: "text"
     }
   }
   ↓

3. 🟢 Go CRM: ProcessInboundMessage
   - Identifica/cria Contact (WhatsappIdentifiers)
   - Identifica/cria Session
   - Persiste Message no PostgreSQL
   - Publica event: message.created (Outbox Pattern)
   ↓

4. 🟢 Go CRM: AI Decision Engine
   - "Precisa de AI Agent?" → SIM
   - Qual agent? → CustomerServiceAgent
   ↓

5. 🟢 Go → Python ADK (gRPC - futuro)
   ExecuteAgent(
     type="CustomerServiceAgent",
     context={
       contact_id: "uuid",
       session_id: "uuid",
       message_id: "uuid",
       input_text: "Qual o valor do contrato que enviamos?"
     }
   )
   ↓

6. 🐍 Python ADK: CustomerServiceAgent.execute()
   - Inicializa agent
   - Chama sub-agents (LeadQualifier, Pricing, etc) - 6 em cadeia
   - Total execution time: 5-10s
   ↓

7. 🐍 Python ADK chama MCP Tools (via HTTP):

   7.1. GET contact data:
        POST http://ventros-crm:8081/v1/mcp/execute
        Body: {
          "tool_name": "get_contact",
          "arguments": {"contact_id": "uuid"}
        }
        ← Response: {contact: {name: "João Silva", ...}}

   7.2. GET conversation history:
        POST http://ventros-crm:8081/v1/mcp/execute
        Body: {
          "tool_name": "get_messages",
          "arguments": {"session_id": "uuid", "limit": 50}
        }
        ← Response: {messages: [{...}, {...}]}

   7.3. SEARCH documents (hybrid: vector + keyword):
        POST http://ventros-crm:8081/v1/mcp/execute
        Body: {
          "tool_name": "search_documents",
          "arguments": {
            "query": "contrato valor",
            "contact_id": "uuid",
            "limit": 5
          }
        }
        ← Response: {documents: [
          {
            content: "Valor: R$ 10.000,00 mensais...",
            document_title: "Contrato.pdf",
            page_number: 3,
            similarity: 0.89
          }
        ]}

   7.4. SEARCH memory facts:
        POST http://ventros-crm:8081/v1/mcp/execute
        Body: {
          "tool_name": "search_memory",
          "arguments": {
            "query": "contrato",
            "contact_id": "uuid"
          }
        }
        ← Response: {
          facts: [
            {fact_text: "João tem contrato de R$ 10k", confidence: 0.95}
          ],
          embeddings: [...]
        }
   ↓

8. 🐍 Python ADK: LLM generates response
   - Context: contact + history + documents + memory
   - Prompt: "Responda a pergunta usando o contexto..."
   - LLM: Gemini 1.5 Flash / Claude Sonnet
   - Output: "O valor do contrato é R$ 10.000,00 mensais, conforme documento enviado em 15/10/2024."
   ↓

9. 🐍 Python ADK retorna para Go (gRPC)
   AgentExecutionResponse {
     status: "success",
     response_text: "O valor do contrato...",
     execution_time_ms: 8500,
     sub_agents_called: 6,
     mcp_tools_called: 4,
     tokens_used: 2500,
     cost_usd: 0.0125
   }
   ↓

10. 🟢 Go CRM: ProcessAgentResponse
    - Persiste response no PostgreSQL
    - Cria Message (from_me=true, content=response_text)
    - Publica events: agent.response_generated, message.created
    ↓

11. 🟢 Go CRM: SendMessage via WAHA
    POST https://waha.ventros.cloud/api/sendText
    Body: {
      session: "5511999999999",
      text: "O valor do contrato é R$ 10.000,00 mensais..."
    }
    ↓

12. 📱 WAHA envia para WhatsApp
    ↓

13. 👤 Cliente recebe resposta no WhatsApp
    "O valor do contrato é R$ 10.000,00 mensais, conforme documento enviado em 15/10/2024."

14. 🟢 Go CRM: Atualiza Message status
    - message.sent → message.delivered (webhook WAHA)
    - Publica event: message.delivered

TEMPO TOTAL: ~12-15 segundos (incluindo AI processing)
```

---

## 📊 RESUMO DE INTEGRAÇÕES

| Integração | Status | Protocolo | Prioridade |
|------------|--------|-----------|------------|
| **Go → Python ADK** | ❌ 0% | gRPC (Port 50051) | P1 (Sprint 2) |
| **Python → Go (MCP)** | ✅ 100% | HTTP (Port 8081) | Done |
| **Memory Service** | ⏳ 20% | Embedded (same DB) | P1 (Sprint 1) |
| **WAHA Webhooks** | ✅ 100% | HTTP webhooks | Done |
| **Stripe Webhooks** | ✅ 100% | HTTP webhooks | Done |
| **Temporal Workflows** | ✅ 90% | gRPC | Done |
| **Frontend ↔ API** | ✅ 100% | REST + WebSocket | Done |

---

**Version**: 1.0
**Last Updated**: 2025-10-15
