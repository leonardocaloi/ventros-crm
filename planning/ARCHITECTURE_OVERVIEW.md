# VENTROS CRM - ARQUITETURA COMPLETA (CORRIGIDA)

**Data**: 2025-10-15
**Versão**: 2.0 (Arquitetura Corrigida)
**Status**: Planning (Sprints 5-30)

---

## 🎯 VISÃO GERAL

**Ventros CRM** é um sistema de CRM multi-canal com IA integrada, composto por:

1. **Ventros CRM (Go)** - Orquestrador Principal & API Backend
2. **Memory Service (Go)** - Serviço de memória híbrida (vector + keyword + graph)
3. **Ventros AI (Python)** - Microserviço de agentes inteligentes
4. **MCP Server (Go)** - Integração com Claude Desktop

---

## 🏗️ ARQUITETURA CORRETA

### Diagrama de Componentes

```
┌─────────────────────────────────────────────────────────────────┐
│                    VENTROS CRM (Go) - ORQUESTRADOR              │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   HTTP API  │  │  WebSocket  │  │   Channel Adapters      │ │
│  │ (Gin/REST)  │  │  (Real-time)│  │ (WhatsApp, IG, FB, etc) │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
│         │                │                     │               │
│         └────────────────┴─────────────────────┘               │
│                          │                                     │
│  ┌──────────────────────┴────────────────────────────────┐    │
│  │           Application Layer (Commands/Queries)         │    │
│  │  - Send Message    - Create Contact   - Start Session │    │
│  │  - Execute Agent   - Search Memories  - List Agents   │    │
│  └──────────────────────┬────────────────────────────────┘    │
│                          │                                     │
│  ┌──────────────────────┴────────────────────────────────┐    │
│  │              Domain Layer (30 Aggregates)              │    │
│  │  - Contact  - Session  - Message  - Channel  - Agent  │    │
│  └──────────────────────┬────────────────────────────────┘    │
│                          │                                     │
│  ┌──────────────────────┴────────────────────────────────┐    │
│  │          Infrastructure Layer (Repositories)           │    │
│  │  - PostgreSQL  - RabbitMQ  - Redis  - Temporal        │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐      │
│  │       MEMORY SERVICE (Embedded Go Service)            │      │
│  │  - Vector Search (pgvector)                          │      │
│  │  - Keyword Search (PostgreSQL FTS)                   │      │
│  │  - Graph Search (temporal knowledge graph)           │      │
│  │  - Hybrid Search (combines all 3)                    │      │
│  │  - Memory Facts Extraction (LLM-based)               │      │
│  └────────────────────────┬─────────────────────────────┘      │
│                            │                                    │
└────────────────────────────┼────────────────────────────────────┘
                             │
                             │ gRPC (bidirectional)
                             │
┌────────────────────────────┼────────────────────────────────────┐
│                            ▼                                    │
│              VENTROS AI (Python ADK Microservice)               │
│                     Agent Library & Executor                    │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐      │
│  │          Agent Registry (Catalog)                     │      │
│  │  - CustomerServiceAgent                              │      │
│  │  - LeadQualifierAgent                                │      │
│  │  - SalesAgent                                        │      │
│  │  - RetentionAgent                                    │      │
│  │  - ChurnPredictionAgent                              │      │
│  │  - SupportTicketAgent                                │      │
│  │  ... (20+ agents)                                    │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐      │
│  │       Agent Orchestration (ADK Framework)             │      │
│  │  - BaseAgent (foundation)                            │      │
│  │  - LlmAgent (ReAct pattern)                          │      │
│  │  - SequentialAgent (chain)                           │      │
│  │  - ParallelAgent (concurrent)                        │      │
│  │  - HierarchicalTaskAgent (coordinator + specialists) │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐      │
│  │            Memory Client (gRPC Client)                │      │
│  │  Calls Go Memory Service for:                        │      │
│  │  - SearchMemories(contactID, query)                  │      │
│  │  - GetContactContext(contactID)                      │      │
│  │  - HybridSearch(query, filters)                      │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                    FRONTEND (React/Next.js)                     │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Dashboard  │  │     Inbox    │  │   Contacts   │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                 │                  │
│         └─────────────────┴─────────────────┘                  │
│                           │                                    │
│                WebSocket + REST API                            │
│                           │                                    │
│                           ▼                                    │
│              Connects to: Go CRM API ONLY                      │
│              (NOT to Python ADK directly)                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                   MCP SERVER (Go) - OPTIONAL                    │
│                  Claude Desktop Integration                     │
│                                                                 │
│  Provides 30 tools to Claude Desktop:                          │
│  - GetContactInfo, SearchContacts, GetSessionHistory           │
│  - SendMessage, CreateContact, UpdatePipeline                  │
│  - RunQuery (BI), SearchMemories, ListAgents                   │
│                                                                 │
│  Connects to: Go CRM API via REST                              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 FLUXO COMPLETO (Mensagem Inbound)

### Cenário: Cliente envia mensagem no WhatsApp

```
1. 📱 WhatsApp Message
   └─> WAHA (WhatsApp API) receives message
       └─> WAHA Webhook → Go CRM API

2. 🟢 Go CRM receives webhook
   └─> Creates Contact (if new)
   └─> Creates/Updates Session
   └─> Creates Message aggregate
   └─> Publishes event: message.received → RabbitMQ Outbox

3. 🧠 Go CRM decides: "Need intelligent response"
   └─> Calls: gRPC ListAvailableAgents() → Python ADK
       └─> Python returns: ["CustomerServiceAgent", "LeadQualifierAgent", ...]

   └─> Go CRM selects: "CustomerServiceAgent"

   └─> Calls: gRPC ExecuteAgent(type="CustomerServiceAgent", context={
           contactID: "uuid",
           message: "Oi, quero comprar produto X",
           channelID: "uuid"
       }) → Python ADK

4. 🐍 Python ADK executes CustomerServiceAgent
   └─> Agent needs context about customer

   └─> Calls: gRPC SearchMemories(contactID, query="previous purchases") → Go Memory Service
       └─> Go returns: {
               memories: [
                   "Purchased product Y 30 days ago",
                   "Complained about delivery delay",
                   "High value customer: R$5000 LTV"
               ],
               embeddings: [...],
               graph: {...}
           }

   └─> Agent processes with Gemini 1.5 Flash + context

   └─> Agent generates response: "Olá! Vi que você já comprou produto Y. Produto X é complementar..."

   └─> Returns to Go CRM: {
           response: "Olá! Vi que você já comprou...",
           intent: "purchase_intent",
           confidence: 0.95,
           suggestedActions: ["create_lead", "update_pipeline"]
       }

5. 🟢 Go CRM processes agent response
   └─> Creates outbound Message aggregate
   └─> Publishes event: message.sent → RabbitMQ Outbox
   └─> Calls WAHA API to send message
   └─> Updates Session last_activity
   └─> Executes suggested actions (create_lead, update_pipeline)
   └─> Emits tracking event for analytics

6. 📱 Message delivered to WhatsApp
   └─> Customer receives intelligent response
   └─> Go CRM receives delivery confirmation from WAHA
   └─> Updates message status to "delivered"
```

---

## 🔄 FLUXO COMPLETO (Mensagem Outbound - Proativa)

### Cenário: Sistema envia campanha proativa

```
1. 🎯 Campaign scheduled in Go CRM
   └─> Temporal Workflow triggers: SendCampaignWorkflow

   └─> Go CRM queries: "Get contacts in segment X"

   └─> For each contact:
       └─> Calls: gRPC ExecuteAgent(type="SalesAgent", context={
               contactID: "uuid",
               campaignID: "uuid",
               objective: "promote_product_Z"
           }) → Python ADK

2. 🐍 Python SalesAgent executes
   └─> Calls: gRPC GetContactContext(contactID) → Go Memory Service
       └─> Returns full context (history, preferences, behavior)

   └─> Agent personalizes message with LLM

   └─> Returns: {
           personalizedMessage: "Oi João, notei que você se interessou por...",
           sendTime: "2024-10-15T14:30:00Z" (best time based on history),
           channel: "whatsapp" (preferred channel)
       }

3. 🟢 Go CRM schedules message
   └─> Creates Message aggregate (status: scheduled)
   └─> Temporal schedules delivery
   └─> At scheduled time:
       └─> Sends via WAHA
       └─> Tracks delivery + read status
       └─> Updates campaign metrics
```

---

## 🔌 API GRPC (Bidirectional Communication)

### Go CRM → Python ADK (Agent Execution)

```protobuf
service AgentService {
  // List all available agents in Python ADK
  rpc ListAvailableAgents(Empty) returns (AgentListResponse);

  // Get capabilities of specific agent
  rpc GetAgentCapabilities(AgentRequest) returns (AgentCapabilitiesResponse);

  // Execute agent with context
  rpc ExecuteAgent(AgentExecutionRequest) returns (AgentExecutionResponse);

  // Stream agent execution (for long-running tasks)
  rpc StreamAgentExecution(AgentExecutionRequest) returns (stream AgentStreamResponse);
}

message AgentListResponse {
  repeated AgentInfo agents = 1;
}

message AgentInfo {
  string type = 1;              // "CustomerServiceAgent"
  string name = 2;              // "Customer Service Specialist"
  string description = 3;
  repeated string capabilities = 4;  // ["intent_classification", "response_generation"]
  int32 avg_latency_ms = 5;     // Performance metric
}

message AgentExecutionRequest {
  string agent_type = 1;         // "CustomerServiceAgent"
  string contact_id = 2;         // UUID
  string message = 3;            // User message
  string channel_id = 4;         // UUID
  map<string, string> context = 5;  // Additional context
}

message AgentExecutionResponse {
  string response = 1;           // Generated response
  string intent = 2;             // Classified intent
  float confidence = 3;          // 0.0 - 1.0
  repeated string suggested_actions = 4;  // ["create_lead", "update_pipeline"]
  map<string, string> metadata = 5;
  int32 latency_ms = 6;
}
```

### Python ADK → Go Memory Service (Memory Access)

```protobuf
service MemoryService {
  // Search memories using hybrid search (vector + keyword + graph)
  rpc SearchMemories(MemorySearchRequest) returns (MemorySearchResponse);

  // Get complete context for a contact
  rpc GetContactContext(ContactContextRequest) returns (ContactContextResponse);

  // Store new memory fact
  rpc StoreMemory(StoreMemoryRequest) returns (StoreMemoryResponse);

  // Get related entities via graph
  rpc GetRelatedEntities(RelatedEntitiesRequest) returns (RelatedEntitiesResponse);
}

message MemorySearchRequest {
  string contact_id = 1;
  string query = 2;              // Natural language query
  int32 limit = 3;               // Max results (default: 10)
  repeated string memory_types = 4;  // ["purchase", "support_ticket", "conversation"]
  MemorySearchMode mode = 5;     // HYBRID, VECTOR_ONLY, KEYWORD_ONLY, GRAPH_ONLY
}

message MemorySearchResponse {
  repeated Memory memories = 1;
  float search_latency_ms = 2;
  int32 total_count = 3;
}

message Memory {
  string id = 1;
  string content = 2;            // "Customer purchased product X on 2024-09-15"
  string memory_type = 3;        // "purchase"
  float relevance_score = 4;     // 0.0 - 1.0
  google.protobuf.Timestamp created_at = 5;
  repeated float embedding = 6;  // Vector (1536 dimensions for text-embedding-3-small)
  map<string, string> metadata = 7;
}

message ContactContextRequest {
  string contact_id = 1;
  bool include_graph = 2;        // Include related entities
  int32 history_limit = 3;       // Last N interactions (default: 50)
}

message ContactContextResponse {
  ContactProfile profile = 1;
  repeated Memory recent_memories = 2;
  repeated Message recent_messages = 3;
  Graph knowledge_graph = 4;     // If include_graph = true
  ContactStats stats = 5;
}
```

---

## 📊 RESPONSABILIDADES DE CADA COMPONENTE

### 1. Go CRM (Orquestrador Principal)

**Responsabilidades**:
- ✅ Gerenciar canais (WhatsApp, Instagram, Facebook)
- ✅ Receber e enviar mensagens (inbound/outbound)
- ✅ Gerenciar entidades (Contact, Session, Message, Channel, Pipeline)
- ✅ Orquestrar workflows (Temporal)
- ✅ Decidir QUANDO usar agentes Python
- ✅ Listar agentes disponíveis (via gRPC)
- ✅ Executar agentes (via gRPC)
- ✅ Fornecer Memory Service para agentes Python
- ✅ Processar eventos (RabbitMQ Outbox Pattern)
- ✅ Gerenciar autenticação e autorização
- ✅ Expor REST API para frontend
- ✅ WebSocket para real-time updates

**NÃO faz**:
- ❌ Processar linguagem natural (delega para Python ADK)
- ❌ Executar LLMs diretamente (delega para Python ADK)
- ❌ Implementar lógica de agentes (delega para Python ADK)

**Tech Stack**:
- Go 1.25.1
- PostgreSQL 15+ (RLS, pgvector)
- RabbitMQ 3.12+ (event bus)
- Redis 7.0+ (cache)
- Temporal (workflows)
- gRPC (client & server)

---

### 2. Memory Service (Embedded Go Service)

**Responsabilidades**:
- ✅ Armazenar embeddings (pgvector)
- ✅ Vector search (similarity search)
- ✅ Keyword search (PostgreSQL FTS)
- ✅ Graph search (temporal knowledge graph)
- ✅ Hybrid search (combina vector + keyword + graph)
- ✅ Memory facts extraction (via LLM)
- ✅ Servir contexto para agentes Python (via gRPC)
- ✅ Servir contexto para Go CRM (in-process)

**NÃO faz**:
- ❌ Processar mensagens
- ❌ Gerenciar canais
- ❌ Executar agentes

**Tech Stack**:
- Go 1.25.1
- PostgreSQL pgvector extension
- OpenAI text-embedding-3-small (embeddings)
- gRPC server

**Status**: 🔴 20% implementado (Sprint 5-11 planejado)

---

### 3. Ventros AI - Python ADK (Agent Library)

**Responsabilidades**:
- ✅ Fornecer catálogo de agentes disponíveis
- ✅ Executar agentes quando chamado pelo Go CRM
- ✅ Processar linguagem natural (intents, entities)
- ✅ Gerar respostas personalizadas (LLM)
- ✅ Executar raciocínio complexo (ReAct, CoT)
- ✅ Chamar ferramentas (tools)
- ✅ Buscar contexto no Memory Service (via gRPC)
- ✅ Retornar resultados estruturados para Go CRM

**NÃO faz**:
- ❌ Gerenciar canais (responsabilidade do Go CRM)
- ❌ Persistir entidades (responsabilidade do Go CRM)
- ❌ Enviar mensagens diretamente (retorna resposta para Go CRM enviar)
- ❌ Gerenciar workflows (responsabilidade do Go CRM via Temporal)
- ❌ Expor API para frontend (frontend conecta ao Go CRM)

**NÃO É**:
- ❌ Frontend
- ❌ Orquestrador principal
- ❌ Source of truth para dados

**É**:
- ✅ Biblioteca de agentes inteligentes
- ✅ Motor de processamento de linguagem natural
- ✅ Executor de lógica de IA
- ✅ Consumidor do Memory Service

**Tech Stack**:
- Python 3.12+
- Google Cloud Agent Development Kit (ADK) 0.5+
- Vertex AI (Gemini 1.5 Flash, Gemini 1.5 Pro)
- gRPC (client & server)
- RabbitMQ (event consumer - opcional)

**Status**: 🔴 0% implementado (Sprint 19-30 planejado)

---

### 4. Frontend (React/Next.js)

**Responsabilidades**:
- ✅ Renderizar UI (dashboard, inbox, contacts)
- ✅ Conectar ao Go CRM via REST API
- ✅ Conectar ao Go CRM via WebSocket (real-time)
- ✅ Exibir mensagens em tempo real
- ✅ Permitir envio de mensagens manuais
- ✅ Gerenciar estado local

**NÃO faz**:
- ❌ Conectar diretamente ao Python ADK
- ❌ Conectar diretamente ao Memory Service
- ❌ Processar lógica de negócio

**Conecta a**:
- ✅ Go CRM API (REST + WebSocket)
- ❌ Python ADK (NUNCA)
- ❌ Memory Service (NUNCA)

---

### 5. MCP Server (Optional - Claude Desktop Integration)

**Responsabilidades**:
- ✅ Expor 30 ferramentas para Claude Desktop
- ✅ Permitir queries BI via Claude
- ✅ Permitir operações CRM via Claude
- ✅ Conectar ao Go CRM API

**NÃO faz**:
- ❌ Substituir Go CRM API
- ❌ Conectar diretamente ao banco de dados

**Status**: 🔴 70% planejado (Sprint 15-18)

---

## 🔀 COMPARAÇÃO: ARQUITETURA ERRADA vs CORRETA

### ❌ ARQUITETURA ERRADA (Antiga)

```
Frontend → Python ADK (Orchestrator) → Go CRM
              ↓
         RabbitMQ (async events)
              ↓
         Go Memory Service
```

**Problemas**:
1. Python como orquestrador central (ERRADO)
2. Frontend conecta ao Python (ERRADO)
3. Python gerencia eventos (ERRADO)
4. Go CRM é subordinado ao Python (ERRADO)

---

### ✅ ARQUITETURA CORRETA

```
Frontend → Go CRM (Orchestrator) ⇄ Python ADK (Agent Library)
              ↓                           ↓
         RabbitMQ                  Go Memory Service
              ↓                           ↑
         PostgreSQL ←────────────────────┘
```

**Correto**:
1. ✅ Go CRM é orquestrador central
2. ✅ Frontend conecta APENAS ao Go CRM
3. ✅ Go CRM decide quando usar agentes Python
4. ✅ Python ADK é biblioteca de agentes (não orquestrador)
5. ✅ Python ADK consome Memory Service do Go CRM
6. ✅ Comunicação bidirecional via gRPC

---

## 📅 TIMELINE DE IMPLEMENTAÇÃO

| Sprint | Feature | Status |
|--------|---------|--------|
| 0-4 | ✅ Go CRM Core (Contact, Session, Message, Channel) | **COMPLETO** |
| 5-11 | 🔄 Memory Service (pgvector, hybrid search) | **PLANEJADO** |
| 12-14 | 🔄 gRPC API (Go ↔ Python bidirectional) | **PLANEJADO** |
| 15-18 | 🔄 MCP Server (Claude Desktop integration) | **PLANEJADO** |
| 19-30 | 🔄 Python ADK (agents + orchestration) | **PLANEJADO** |

---

## 🎯 PRÓXIMOS PASSOS

1. **Consolidar AI_REPORT** (6 partes → 1) em `code-analysis/architecture/`
2. **Corrigir** `planning/ventros-ai/ARCHITECTURE.md` (remover referências a "orchestrator")
3. **Criar** `planning/grpc-api/SPECIFICATION.md` (API gRPC completa)
4. **Criar** comandos slash (.claude/commands/)
5. **Criar** documento final de consolidação

---

**Versão**: 2.0 (Arquitetura Corrigida)
**Última Atualização**: 2025-10-15
**Responsável**: Claude Code (consolidação de documentação)
