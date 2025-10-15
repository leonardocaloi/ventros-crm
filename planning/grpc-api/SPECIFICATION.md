# gRPC API SPECIFICATION - Ventros CRM

**Versão**: 1.0
**Status**: Planning (Sprint 12-14)
**Protocolo**: gRPC (HTTP/2)
**Serialização**: Protocol Buffers v3

---

## 🎯 VISÃO GERAL

A API gRPC permite comunicação **bidirecional** entre:

1. **Go CRM → Python ADK**: Execução de agentes inteligentes
2. **Python ADK → Go Memory Service**: Acesso a contexto e memórias

```
┌──────────────────┐                    ┌──────────────────┐
│                  │   ExecuteAgent()   │                  │
│     Go CRM       │ ──────────────────>│   Python ADK     │
│  (gRPC Client)   │                    │  (gRPC Server)   │
│                  │<── AgentResponse ─ │                  │
└──────────────────┘                    └─────────┬────────┘
         ▲                                        │
         │                                        │
         │  SearchMemories()                      │
         │<────────────────────────────────────────
         │
┌────────┴─────────┐
│  Memory Service  │
│  (gRPC Server)   │
└──────────────────┘
```

---

## 📋 ÍNDICE

1. [Agent Service (Go → Python)](#1-agent-service-go--python)
2. [Memory Service (Python → Go)](#2-memory-service-python--go)
3. [Protocol Buffers Definitions](#3-protocol-buffers-definitions)
4. [Error Handling](#4-error-handling)
5. [Performance & Optimization](#5-performance--optimization)
6. [Examples](#6-examples)
7. [Deployment](#7-deployment)

---

## 1. AGENT SERVICE (Go → Python)

### 1.1. ListAvailableAgents

**Propósito**: Obter catálogo de agentes disponíveis no Python ADK

**Request**:
```protobuf
message ListAgentsRequest {
  // Empty - retorna todos os agentes
}
```

**Response**:
```protobuf
message ListAgentsResponse {
  repeated AgentInfo agents = 1;
  int32 total_count = 2;
}

message AgentInfo {
  string type = 1;                    // "CustomerServiceAgent"
  string name = 2;                    // "Customer Service Specialist"
  string description = 3;             // "Handles customer inquiries..."
  repeated string capabilities = 4;   // ["intent_classification", "response_generation"]
  repeated string required_inputs = 5; // ["contact_id", "message", "channel_id"]
  int32 avg_latency_ms = 6;           // Performance metric (historical)
  AgentCategory category = 7;         // SALES, SUPPORT, MARKETING, RETENTION
}

enum AgentCategory {
  SALES = 0;
  SUPPORT = 1;
  MARKETING = 2;
  RETENTION = 3;
  ANALYTICS = 4;
}
```

**Exemplo de Uso**:
```go
// Go CRM
resp, err := agentClient.ListAvailableAgents(ctx, &pb.ListAgentsRequest{})
if err != nil {
    return err
}

for _, agent := range resp.Agents {
    fmt.Printf("Agent: %s (%s)\n", agent.Name, agent.Type)
    fmt.Printf("  Capabilities: %v\n", agent.Capabilities)
    fmt.Printf("  Avg Latency: %dms\n", agent.AvgLatencyMs)
}

// Output:
// Agent: Customer Service Specialist (CustomerServiceAgent)
//   Capabilities: [intent_classification response_generation sentiment_analysis]
//   Avg Latency: 850ms
//
// Agent: Lead Qualifier (LeadQualifierAgent)
//   Capabilities: [lead_scoring qualification_questions urgency_detection]
//   Avg Latency: 920ms
```

---

### 1.2. GetAgentCapabilities

**Propósito**: Obter detalhes de um agente específico

**Request**:
```protobuf
message GetAgentCapabilitiesRequest {
  string agent_type = 1;  // "CustomerServiceAgent"
}
```

**Response**:
```protobuf
message GetAgentCapabilitiesResponse {
  AgentInfo agent = 1;
  AgentSchema schema = 2;
  repeated Example examples = 3;
}

message AgentSchema {
  repeated Field input_fields = 1;
  repeated Field output_fields = 2;
}

message Field {
  string name = 1;         // "contact_id"
  string type = 2;         // "string" | "int" | "float" | "object"
  bool required = 3;
  string description = 4;
  string default_value = 5;
}

message Example {
  map<string, string> input = 1;
  map<string, string> output = 2;
  string description = 3;
}
```

**Exemplo de Uso**:
```go
// Go CRM
resp, err := agentClient.GetAgentCapabilities(ctx, &pb.GetAgentCapabilitiesRequest{
    AgentType: "CustomerServiceAgent",
})

fmt.Printf("Agent: %s\n", resp.Agent.Name)
fmt.Printf("Input Fields:\n")
for _, field := range resp.Schema.InputFields {
    required := ""
    if field.Required {
        required = " (required)"
    }
    fmt.Printf("  - %s (%s)%s: %s\n",
        field.Name, field.Type, required, field.Description)
}

// Output:
// Agent: Customer Service Specialist
// Input Fields:
//   - contact_id (string) (required): UUID of the contact
//   - message (string) (required): User message to process
//   - channel_id (string) (required): Channel UUID (WhatsApp, Instagram, etc)
//   - session_id (string): Current session UUID (optional)
//   - context (object): Additional context (optional)
```

---

### 1.3. ExecuteAgent

**Propósito**: Executar agente com contexto fornecido

**Request**:
```protobuf
message ExecuteAgentRequest {
  string agent_type = 1;              // "CustomerServiceAgent"
  string contact_id = 2;              // UUID
  string message = 3;                 // User message
  string channel_id = 4;              // UUID
  string session_id = 5;              // UUID (optional)
  map<string, string> context = 6;    // Additional context
  ExecutionOptions options = 7;
}

message ExecutionOptions {
  int32 timeout_ms = 1;               // Max execution time (default: 30000)
  bool stream_response = 2;           // Use StreamAgentExecution instead
  repeated string tools = 3;          // Whitelist tools (empty = all allowed)
  int32 max_tokens = 4;               // LLM max tokens (default: 1024)
  float temperature = 5;              // LLM temperature (default: 0.7)
}
```

**Response**:
```protobuf
message ExecuteAgentResponse {
  string response = 1;                // Generated response text
  string intent = 2;                  // Classified intent
  float confidence = 3;               // 0.0 - 1.0
  repeated string suggested_actions = 4;  // ["create_lead", "update_pipeline"]
  map<string, string> metadata = 5;   // Additional metadata
  int32 latency_ms = 6;               // Execution time
  repeated ToolCall tool_calls = 7;   // Tools used during execution
  AgentThoughts thoughts = 8;         // Reasoning steps (optional)
}

message ToolCall {
  string tool_name = 1;               // "search_memories"
  map<string, string> input = 2;
  string output = 3;
  int32 latency_ms = 4;
}

message AgentThoughts {
  repeated string reasoning_steps = 1;
  string final_answer = 2;
}
```

**Exemplo de Uso**:
```go
// Go CRM
resp, err := agentClient.ExecuteAgent(ctx, &pb.ExecuteAgentRequest{
    AgentType: "CustomerServiceAgent",
    ContactId: "550e8400-e29b-41d4-a716-446655440000",
    Message: "Oi, quero saber sobre o produto X",
    ChannelId: "660e8400-e29b-41d4-a716-446655440001",
    Context: map[string]string{
        "last_purchase_date": "2024-09-15",
        "customer_tier": "premium",
    },
    Options: &pb.ExecutionOptions{
        TimeoutMs: 30000,
        Temperature: 0.7,
    },
})

if err != nil {
    return err
}

fmt.Printf("Response: %s\n", resp.Response)
fmt.Printf("Intent: %s (confidence: %.2f)\n", resp.Intent, resp.Confidence)
fmt.Printf("Suggested Actions: %v\n", resp.SuggestedActions)
fmt.Printf("Latency: %dms\n", resp.LatencyMs)

// Output:
// Response: Olá! Vi que você é cliente premium e comprou recentemente.
//           O produto X é complementar ao produto Y. Posso te enviar...
// Intent: purchase_intent (confidence: 0.95)
// Suggested Actions: [create_lead update_pipeline_status send_product_catalog]
// Latency: 850ms
```

---

### 1.4. StreamAgentExecution

**Propósito**: Stream de execução de agente (Server-Side Streaming)

**Request**:
```protobuf
message StreamAgentRequest {
  // Same as ExecuteAgentRequest
  string agent_type = 1;
  string contact_id = 2;
  string message = 3;
  // ... outros campos
}
```

**Response Stream**:
```protobuf
message AgentStreamChunk {
  ChunkType type = 1;
  string content = 2;
  map<string, string> metadata = 3;
}

enum ChunkType {
  THOUGHT = 0;      // Reasoning step
  TOOL_CALL = 1;    // Tool execution
  PARTIAL = 2;      // Partial response (incremental)
  FINAL = 3;        // Final response
  ERROR = 4;        // Error occurred
}
```

**Exemplo de Uso**:
```go
// Go CRM
stream, err := agentClient.StreamAgentExecution(ctx, &pb.StreamAgentRequest{
    AgentType: "CustomerServiceAgent",
    ContactId: contactID,
    Message: userMessage,
})

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        return err
    }

    switch chunk.Type {
    case pb.ChunkType_THOUGHT:
        fmt.Printf("[THINKING] %s\n", chunk.Content)
    case pb.ChunkType_TOOL_CALL:
        fmt.Printf("[TOOL] %s\n", chunk.Content)
    case pb.ChunkType_PARTIAL:
        fmt.Printf("%s", chunk.Content)  // Incremental response
    case pb.ChunkType_FINAL:
        fmt.Printf("\n[DONE] Confidence: %s\n", chunk.Metadata["confidence"])
    }
}

// Output:
// [THINKING] User is asking about product X, need to check purchase history
// [TOOL] Calling search_memories(contact_id=..., query="purchases")
// [TOOL] Found: Customer purchased product Y 30 days ago
// [THINKING] Product X complements product Y, good upsell opportunity
// Olá! Vi que você é cliente premium e comprou recentemente...
// [DONE] Confidence: 0.95
```

---

## 2. MEMORY SERVICE (Python → Go)

### 2.1. SearchMemories

**Propósito**: Buscar memórias usando hybrid search (vector + keyword + graph)

**Request**:
```protobuf
message SearchMemoriesRequest {
  string contact_id = 1;              // UUID
  string query = 2;                   // Natural language query
  int32 limit = 3;                    // Max results (default: 10)
  repeated string memory_types = 4;   // ["purchase", "conversation", "support_ticket"]
  SearchMode mode = 5;                // HYBRID, VECTOR_ONLY, KEYWORD_ONLY, GRAPH_ONLY
  repeated string include_fields = 6; // ["content", "embedding", "metadata"]
}

enum SearchMode {
  HYBRID = 0;         // Vector + keyword + graph (best quality)
  VECTOR_ONLY = 1;    // Semantic similarity only (fast)
  KEYWORD_ONLY = 2;   // Full-text search (exact matches)
  GRAPH_ONLY = 3;     // Knowledge graph traversal
}
```

**Response**:
```protobuf
message SearchMemoriesResponse {
  repeated Memory memories = 1;
  float search_latency_ms = 2;
  int32 total_count = 3;
  SearchMetadata metadata = 4;
}

message Memory {
  string id = 1;                      // UUID
  string content = 2;                 // "Customer purchased product X on 2024-09-15"
  string memory_type = 3;             // "purchase"
  float relevance_score = 4;          // 0.0 - 1.0
  google.protobuf.Timestamp created_at = 5;
  repeated float embedding = 6;       // Vector (1536 dims for text-embedding-3-small)
  map<string, string> metadata = 7;   // Additional data
  repeated RelatedEntity related = 8;
}

message RelatedEntity {
  string entity_type = 1;             // "product", "campaign", "ticket"
  string entity_id = 2;               // UUID
  string relationship = 3;            // "purchased", "mentioned", "complained_about"
  float strength = 4;                 // 0.0 - 1.0
}

message SearchMetadata {
  int32 vector_matches = 1;
  int32 keyword_matches = 2;
  int32 graph_matches = 3;
  float vector_weight = 4;
  float keyword_weight = 5;
  float graph_weight = 6;
}
```

**Exemplo de Uso**:
```python
# Python ADK Agent
import grpc
from ventros.memory_service import memory_pb2, memory_pb2_grpc

# Durante execução do agente
channel = grpc.insecure_channel('localhost:50052')
memory_client = memory_pb2_grpc.MemoryServiceStub(channel)

response = memory_client.SearchMemories(
    memory_pb2.SearchMemoriesRequest(
        contact_id=contact_id,
        query="recent purchases and complaints",
        limit=10,
        memory_types=["purchase", "support_ticket"],
        mode=memory_pb2.SearchMode.HYBRID
    )
)

for memory in response.memories:
    print(f"[{memory.memory_type}] {memory.content}")
    print(f"  Relevance: {memory.relevance_score:.2f}")
    print(f"  Related: {[f'{r.entity_type}:{r.entity_id}' for r in memory.related]}")

# Output:
# [purchase] Customer purchased product Y on 2024-09-15
#   Relevance: 0.95
#   Related: ['product:uuid-123', 'campaign:uuid-456']
#
# [support_ticket] Complained about delivery delay on 2024-09-18
#   Relevance: 0.87
#   Related: ['ticket:uuid-789', 'product:uuid-123']
```

---

### 2.2. GetContactContext

**Propósito**: Obter contexto completo de um contato (perfil + memórias + grafo)

**Request**:
```protobuf
message GetContactContextRequest {
  string contact_id = 1;
  bool include_graph = 2;             // Include knowledge graph (expensive)
  int32 history_limit = 3;            // Last N interactions (default: 50)
  repeated string entity_types = 4;   // ["message", "purchase", "ticket"]
}
```

**Response**:
```protobuf
message GetContactContextResponse {
  ContactProfile profile = 1;
  repeated Memory recent_memories = 2;
  repeated Message recent_messages = 3;
  KnowledgeGraph graph = 4;           // If include_graph = true
  ContactStats stats = 5;
  float latency_ms = 6;
}

message ContactProfile {
  string contact_id = 1;
  string name = 2;
  string phone = 3;
  string email = 4;
  string tier = 5;                    // "premium", "standard", "free"
  repeated string tags = 6;
  map<string, string> custom_fields = 7;
}

message Message {
  string id = 1;
  string content = 2;
  bool is_inbound = 3;
  google.protobuf.Timestamp sent_at = 4;
  string channel = 5;
}

message KnowledgeGraph {
  repeated Node nodes = 1;
  repeated Edge edges = 2;
}

message Node {
  string id = 1;
  string type = 2;                    // "contact", "product", "campaign"
  map<string, string> properties = 3;
}

message Edge {
  string from_id = 1;
  string to_id = 2;
  string relationship = 3;            // "purchased", "interested_in"
  float weight = 4;
  google.protobuf.Timestamp created_at = 5;
}

message ContactStats {
  int32 total_messages = 1;
  int32 total_purchases = 2;
  float lifetime_value = 3;
  google.protobuf.Timestamp first_contact = 4;
  google.protobuf.Timestamp last_activity = 5;
  float churn_risk = 6;               // 0.0 - 1.0
}
```

**Exemplo de Uso**:
```python
# Python ADK Agent
response = memory_client.GetContactContext(
    memory_pb2.GetContactContextRequest(
        contact_id=contact_id,
        include_graph=True,
        history_limit=20
    )
)

print(f"Contact: {response.profile.name} ({response.profile.tier})")
print(f"LTV: R${response.stats.lifetime_value:.2f}")
print(f"Churn Risk: {response.stats.churn_risk:.2%}")
print(f"\nRecent Memories ({len(response.recent_memories)}):")
for mem in response.recent_memories[:5]:
    print(f"  - {mem.content}")

print(f"\nKnowledge Graph:")
print(f"  Nodes: {len(response.graph.nodes)}")
print(f"  Edges: {len(response.graph.edges)}")
for edge in response.graph.edges[:5]:
    print(f"  - {edge.from_id} --[{edge.relationship}]--> {edge.to_id}")

# Output:
# Contact: João Silva (premium)
# LTV: R$5000.00
# Churn Risk: 12.50%
#
# Recent Memories (10):
#   - Customer purchased product Y on 2024-09-15
#   - Complained about delivery delay on 2024-09-18
#   - Support ticket resolved on 2024-09-20
#   - Received discount coupon on 2024-09-22
#   - Opened email campaign "Fall Sale" on 2024-09-25
#
# Knowledge Graph:
#   Nodes: 15
#   Edges: 23
#   - contact:uuid --[purchased]--> product:uuid
#   - contact:uuid --[complained_about]--> ticket:uuid
#   - contact:uuid --[interested_in]--> category:electronics
```

---

### 2.3. StoreMemory

**Propósito**: Armazenar nova memória (criada pelo agente)

**Request**:
```protobuf
message StoreMemoryRequest {
  string contact_id = 1;
  string content = 2;                 // Memory text
  string memory_type = 3;             // "agent_insight", "preference", "fact"
  map<string, string> metadata = 4;
  bool auto_embed = 5;                // Auto-generate embedding (default: true)
  repeated RelatedEntity related = 6; // Link to existing entities
}
```

**Response**:
```protobuf
message StoreMemoryResponse {
  string memory_id = 1;               // UUID of created memory
  bool success = 2;
  float latency_ms = 3;
}
```

**Exemplo de Uso**:
```python
# Python ADK Agent armazena insight
response = memory_client.StoreMemory(
    memory_pb2.StoreMemoryRequest(
        contact_id=contact_id,
        content="Customer prefers communication after 2pm",
        memory_type="preference",
        metadata={
            "source": "agent_inference",
            "confidence": "0.89"
        },
        auto_embed=True
    )
)

print(f"Memory stored: {response.memory_id}")
# Output: Memory stored: 770e8400-e29b-41d4-a716-446655440002
```

---

### 2.4. GetRelatedEntities

**Propósito**: Buscar entidades relacionadas via knowledge graph

**Request**:
```protobuf
message GetRelatedEntitiesRequest {
  string entity_id = 1;               // Starting node (contact_id, product_id, etc)
  string entity_type = 2;             // "contact", "product", "campaign"
  repeated string target_types = 3;   // ["product", "ticket"]
  int32 max_depth = 4;                // Graph traversal depth (default: 2)
  int32 limit = 5;                    // Max results (default: 20)
}
```

**Response**:
```protobuf
message GetRelatedEntitiesResponse {
  repeated RelatedEntityResult entities = 1;
  float latency_ms = 2;
}

message RelatedEntityResult {
  Node entity = 1;
  repeated Edge path = 2;             // Path from source to this entity
  float relevance = 3;                // 0.0 - 1.0
}
```

---

## 3. PROTOCOL BUFFERS DEFINITIONS

### 3.1. Arquivo: go_to_python.proto

```protobuf
syntax = "proto3";

package ventros.agents;

option go_package = "github.com/ventros/crm/internal/grpc/agent";

import "google/protobuf/timestamp.proto";

// Agent Service (Go CRM → Python ADK)
service AgentService {
  rpc ListAvailableAgents(ListAgentsRequest) returns (ListAgentsResponse);
  rpc GetAgentCapabilities(GetAgentCapabilitiesRequest) returns (GetAgentCapabilitiesResponse);
  rpc ExecuteAgent(ExecuteAgentRequest) returns (ExecuteAgentResponse);
  rpc StreamAgentExecution(StreamAgentRequest) returns (stream AgentStreamChunk);
}

// [... messages definidas anteriormente ...]
```

### 3.2. Arquivo: python_to_go.proto

```protobuf
syntax = "proto3";

package ventros.memory;

option go_package = "github.com/ventros/crm/internal/grpc/memory";

import "google/protobuf/timestamp.proto";

// Memory Service (Python ADK → Go CRM)
service MemoryService {
  rpc SearchMemories(SearchMemoriesRequest) returns (SearchMemoriesResponse);
  rpc GetContactContext(GetContactContextRequest) returns (GetContactContextResponse);
  rpc StoreMemory(StoreMemoryRequest) returns (StoreMemoryResponse);
  rpc GetRelatedEntities(GetRelatedEntitiesRequest) returns (GetRelatedEntitiesResponse);
}

// [... messages definidas anteriormente ...]
```

---

## 4. ERROR HANDLING

### 4.1. gRPC Status Codes

| Code | Quando Usar | Exemplo |
|------|-------------|---------|
| `OK` (0) | Sucesso | Agent executado com sucesso |
| `INVALID_ARGUMENT` (3) | Input inválido | Missing contact_id |
| `NOT_FOUND` (5) | Entidade não existe | Agent type not found |
| `DEADLINE_EXCEEDED` (4) | Timeout | Agent execution > 30s |
| `RESOURCE_EXHAUSTED` (8) | Rate limit | Too many requests |
| `INTERNAL` (13) | Erro interno | LLM API error |
| `UNAVAILABLE` (14) | Serviço offline | Python ADK down |

### 4.2. Error Details

```protobuf
message ErrorDetails {
  string code = 1;              // "AGENT_NOT_FOUND"
  string message = 2;           // "Agent type 'XYZ' does not exist"
  map<string, string> metadata = 3;
  repeated string suggestions = 4;  // ["Try CustomerServiceAgent instead"]
}
```

### 4.3. Retry Strategy

**Go CRM** (client):
```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

func executeAgentWithRetry(ctx context.Context, req *pb.ExecuteAgentRequest) (*pb.ExecuteAgentResponse, error) {
    var resp *pb.ExecuteAgentResponse
    var err error

    for attempt := 1; attempt <= 3; attempt++ {
        resp, err = agentClient.ExecuteAgent(ctx, req)
        if err == nil {
            return resp, nil
        }

        st, ok := status.FromError(err)
        if !ok {
            return nil, err
        }

        switch st.Code() {
        case codes.DeadlineExceeded, codes.Unavailable:
            // Retry
            time.Sleep(time.Duration(attempt) * time.Second)
            continue
        case codes.InvalidArgument, codes.NotFound:
            // Don't retry
            return nil, err
        default:
            return nil, err
        }
    }

    return nil, err
}
```

---

## 5. PERFORMANCE & OPTIMIZATION

### 5.1. Connection Pooling

**Go CRM**:
```go
// Shared connection pool (reuse connections)
conn, err := grpc.Dial(
    "localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }),
)
defer conn.Close()

agentClient := pb.NewAgentServiceClient(conn)
```

**Python ADK**:
```python
from concurrent import futures
import grpc

server = grpc.server(
    futures.ThreadPoolExecutor(max_workers=10),
    options=[
        ('grpc.max_send_message_length', 50 * 1024 * 1024),  # 50MB
        ('grpc.max_receive_message_length', 50 * 1024 * 1024),
        ('grpc.keepalive_time_ms', 10000),
        ('grpc.keepalive_timeout_ms', 3000),
    ]
)
```

### 5.2. Timeouts

**Agent Execution**:
- Default: 30s
- Max: 60s (long-running tasks)
- Streaming: No timeout (client-controlled)

**Memory Search**:
- Hybrid: 500ms
- Vector-only: 200ms
- Keyword-only: 100ms

### 5.3. Compression

```go
// Go CRM
import "google.golang.org/grpc/encoding/gzip"

resp, err := agentClient.ExecuteAgent(
    ctx,
    req,
    grpc.UseCompressor(gzip.Name),  // Enable gzip compression
)
```

### 5.4. Batching

```protobuf
// Batch execution (future feature)
message ExecuteAgentBatchRequest {
  repeated ExecuteAgentRequest requests = 1;
}

message ExecuteAgentBatchResponse {
  repeated ExecuteAgentResponse responses = 1;
}
```

---

## 6. EXAMPLES

### 6.1. Full Workflow Example

**Go CRM**:
```go
// 1. List available agents
listResp, _ := agentClient.ListAvailableAgents(ctx, &pb.ListAgentsRequest{})
fmt.Printf("Available agents: %d\n", listResp.TotalCount)

// 2. Select CustomerServiceAgent
selectedAgent := "CustomerServiceAgent"

// 3. Execute agent
execResp, err := agentClient.ExecuteAgent(ctx, &pb.ExecuteAgentRequest{
    AgentType: selectedAgent,
    ContactId: contactID,
    Message: "Oi, quero comprar produto X",
    ChannelId: channelID,
    Context: map[string]string{
        "last_purchase": "product_Y",
        "tier": "premium",
    },
})

if err != nil {
    return err
}

// 4. Process response
fmt.Printf("Response: %s\n", execResp.Response)

// 5. Execute suggested actions
for _, action := range execResp.SuggestedActions {
    switch action {
    case "create_lead":
        createLead(contactID, execResp.Metadata)
    case "update_pipeline":
        updatePipelineStatus(contactID, "qualified")
    }
}

// 6. Send message via WAHA
wahaClient.SendMessage(channelID, contactID, execResp.Response)
```

**Python ADK Agent**:
```python
class CustomerServiceAgent(LlmAgent):
    def run(self, contact_id: str, message: str, **kwargs):
        # 1. Get context from Memory Service
        context = self.memory_client.GetContactContext(
            memory_pb2.GetContactContextRequest(
                contact_id=contact_id,
                history_limit=20
            )
        )

        # 2. Build prompt with context
        prompt = f"""
        Customer: {context.profile.name} ({context.profile.tier})
        LTV: ${context.stats.lifetime_value}
        Recent Messages: {[m.content for m in context.recent_messages[:5]]}

        User Message: {message}

        Generate a helpful response.
        """

        # 3. Call LLM
        response = self.llm.generate(prompt)

        # 4. Store insight
        if response.confidence > 0.8:
            self.memory_client.StoreMemory(
                memory_pb2.StoreMemoryRequest(
                    contact_id=contact_id,
                    content=f"Agent classified intent as: {response.intent}",
                    memory_type="agent_insight"
                )
            )

        # 5. Return to Go CRM
        return {
            "response": response.text,
            "intent": response.intent,
            "confidence": response.confidence,
            "suggested_actions": ["create_lead", "update_pipeline"]
        }
```

---

## 7. DEPLOYMENT

### 7.1. Go CRM (gRPC Client)

```yaml
# docker-compose.yml
services:
  crm-api:
    build: .
    ports:
      - "8080:8080"      # REST API
    environment:
      PYTHON_ADK_GRPC: "python-adk:50051"
      MEMORY_SERVICE_GRPC: "localhost:50052"  # In-process (embedded)
```

### 7.2. Python ADK (gRPC Server)

```yaml
# docker-compose.yml
services:
  python-adk:
    build: ./ventros-ai
    ports:
      - "50051:50051"    # gRPC Agent Service
    environment:
      GO_MEMORY_SERVICE: "crm-api:50052"
      VERTEX_PROJECT_ID: "..."
```

### 7.3. Health Checks

```protobuf
service HealthService {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {
  string service = 1;  // "agent", "memory"
}

message HealthCheckResponse {
  enum Status {
    SERVING = 0;
    NOT_SERVING = 1;
  }
  Status status = 1;
}
```

---

## 📊 SUMMARY

| Feature | Status | Sprint |
|---------|--------|--------|
| Agent Service (Go → Python) | 🔴 Planned | 12-14 |
| Memory Service (Python → Go) | 🔴 Planned | 5-11 (dependency) |
| Protocol Buffers | 🔴 Not Started | 12 |
| Error Handling | 🔴 Not Started | 13 |
| Performance Optimization | 🔴 Not Started | 14 |
| Deployment | 🔴 Not Started | 14 |

**Dependencies**:
1. Memory Service must be implemented first (Sprint 5-11)
2. Python ADK foundation (Sprint 19-30)
3. gRPC API (Sprint 12-14) bridges the two

---

**Versão**: 1.0
**Última Atualização**: 2025-10-15
**Status**: Planning (Sprint 12-14)
**Responsável**: Claude Code (consolidação de documentação)
