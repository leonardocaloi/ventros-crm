# 🧠 VENTROS AI ARCHITECTURE - EXECUTIVE SUMMARY

> **Como tudo funciona em conjunto: Go CRM + Python ADK + Memory Layer**

---

## 🎯 VISÃO GERAL

### Arquitetura Loosely Coupled

```
┌─────────────────────────────────────────────────────────────────┐
│  VENTROS CRM (Go) - SERVIÇO PRIMÁRIO                            │
│  ✅ Roda INDEPENDENTE (não depende do AI Service)               │
│  ✅ Domain Model rico (sessions, contacts, messages, agents)    │
│  ✅ Memory Layer (embeddings, hybrid search, knowledge graph)   │
│  ✅ gRPC Server (expõe memória para Python)                     │
│  ✅ Event Publisher (RabbitMQ)                                   │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           │ RabbitMQ (async, fire-and-forget)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  VENTROS AI (Python ADK) - SERVIÇO OPCIONAL                     │
│  ✅ Roda INDEPENDENTE (degrada gracefully se Go está down)      │
│  ✅ Multi-agent orchestration (Coordinator + Specialists)       │
│  ✅ Memory-aware agents (chama Go via gRPC)                     │
│  ✅ Temporal workflows (long-running processes)                 │
│  ✅ Phoenix observability (LLM tracing)                          │
└─────────────────────────────────────────────────────────────────┘
```

**Princípio:** Loose coupling + eventual consistency + graceful degradation

---

## 📊 FLUXO COMPLETO: MENSAGEM → RESPOSTA AI

### 1. **Mensagem Chega (WAHA → Go)**

```
Cliente envia: "Quero cancelar minha conta, muito caro"
    ↓
WAHA Webhook → Go API
    ↓
Go cria Message entity
    ↓
Go publica evento: MessageReceived → RabbitMQ
```

**Go persiste:**
- Message (id, text, contactID, sessionID, timestamp, fromMe=false)
- Atualiza Session (adiciona msg, incrementa message_count)

---

### 2. **Python ADK Consome Evento**

```
RabbitMQ → Python EventConsumer
    ↓
EventConsumer.handle_message_received()
    ↓
1. Load/create ADK Session
2. Semantic Router → classifica intent
3. Agent routing decision
```

**Semantic Router:**
```python
message: "Quero cancelar minha conta, muito caro"
    ↓
Embedding similarity com intent_examples
    ↓
Top match: "retention_churn" (confidence: 0.89)
    ↓
session.state["agent_category"] = "retention_churn"
```

---

### 3. **Agent Busca Memória (Python → Go via gRPC)**

```
Python: RetentionChurnAgent.run()
    ↓
Agent chama search_memory tool (implicit via BaseMemoryService)
    ↓
VentrosMemoryService.search_memory()
    ↓
gRPC call → Go Memory Service
```

**Request gRPC:**
```protobuf
SearchMemoryRequest {
    tenant_id: "tenant-123"
    contact_id: "contact-456"
    query: "Customer dissatisfaction and pricing concerns"
    agent_category: "retention_churn"
    top_k: 10
}
```

---

### 4. **Go Executa Hybrid Search**

```
Go Memory Service recebe request
    ↓
1. BASELINE: Busca últimas 20 mensagens (SQL) ✅ SEMPRE
    ↓
2. VECTOR SEARCH (50%): Similar sessions via pgvector
    ↓
3. KEYWORD SEARCH (20%): "cancelar" + "caro" via pg_trgm
    ↓
4. GRAPH TRAVERSAL (20%): Agent transfer chain via Apache AGE
    ↓
5. FUSION (RRF): Combina resultados
    ↓
6. RERANKING (Jina v2): Top-10 → Top-5
    ↓
7. AGREGAÇÕES: Contact stats + Memory facts
```

**Go retorna contexto formatado:**
```
=== RECENT MESSAGES ===
[2025-10-13 14:32] Contact: "Quero cancelar minha conta, muito caro"
[2025-10-13 13:15] Agent: "Como posso ajudar?"
[... últimas 20 mensagens]

=== SIMILAR PAST SESSIONS ===
1. Cliente queria cancelar por preço, retido com 25% desconto (similarity: 0.89)
2. Reclamação sobre valor alto, não retido (similarity: 0.78)
3. Cliente insatisfeito com ROI (similarity: 0.72)

=== CONTACT CONTEXT ===
- Total sessions: 12
- Avg sentiment: -0.35 (trending negative)
- Last interaction: 2025-10-10
- Campaign source: Google Ads (utm_campaign=retencao_q4)
- Pipeline stage: active_customer (90 days)

=== MEMORY FACTS ===
- Budget: R$ 5000/mês (confidence: 0.85, from: session-789)
- Pain point: "ROI não está claro" (confidence: 0.92, from: session-801)
- Objection: "Concorrente oferece 30% mais barato" (confidence: 0.78, from: session-801)
```

---

### 5. **Agent Processa com LLM + Tools**

```
RetentionChurnAgent recebe contexto da memória
    ↓
LLM (Gemini 2.0 Flash) processa:
    - System instruction (retenção specialist)
    - Memory context (histórico formatado)
    - User message ("Quero cancelar...")
    ↓
LLM decide: Criar oferta de retenção (30% desconto)
    ↓
Agent chama tool: create_retention_offer
    ↓
Tool chama Go API via gRPC:
    CreateRetentionOfferRequest {
        contact_id: "contact-456"
        discount_percent: 25  // Agent respeitou constraint (max 30%)
        duration_months: 3
        reason: "price_objection"
    }
    ↓
Go cria offer record + retorna confirmation
    ↓
Agent gera resposta final
```

**Resposta do agent:**
```
"Entendo sua preocupação, Leonardo. Vi que você está conosco há 90 dias
e valorizamos muito sua parceria. Podemos oferecer 25% de desconto nos
próximos 3 meses enquanto trabalhamos para demonstrar melhor o ROI.
O que acha?"
```

---

### 6. **Python Publica Resposta (Python → RabbitMQ → Go)**

```
Agent retorna resposta
    ↓
Python publica evento: OutboundMessage → RabbitMQ
    ↓
{
    "event_type": "message.send",
    "payload": {
        "contact_id": "contact-456",
        "session_id": "session-789",
        "text": "Entendo sua preocupação...",
        "source": "bot",
        "agent_id": "agent-retention-churn",
        "metadata": {
            "agent_category": "retention_churn",
            "confidence": 0.89,
            "offer_id": "offer-123"
        }
    }
}
```

---

### 7. **Go Envia Mensagem (RabbitMQ → Go → WAHA)**

```
Go EventConsumer consome OutboundMessage
    ↓
Go cria Message entity (fromMe=true)
    ↓
Go envia via WAHA Client:
    POST /api/sendText
    {
        "session": "phone@c.us",
        "text": "Entendo sua preocupação..."
    }
    ↓
WAHA → WhatsApp → Cliente recebe mensagem
```

---

### 8. **Background: Go Atualiza Memória**

```
Async (não bloqueia response):
    ↓
1. Gera embedding contextual da sessão:
    - LLM (Gemini Flash) gera contexto enriquecido
    - Embedding service (text-embedding-005) gera vetor 768D
    - Persiste em memory_embeddings (pgvector)
    ↓
2. Extrai memory facts (LLM):
    - "Cliente tem budget de R$ 5000"
    - "Pain point: ROI não está claro"
    - Verifica contradições com facts existentes
    - Persiste em memory_facts
    ↓
3. Atualiza knowledge graph (Apache AGE):
    - Cria nodes: Session, Offer
    - Cria edges: Contact -[has_session]-> Session
    - Cria edges: Session -[received_offer]-> Offer
    ↓
Memória atualizada para próximas interações ✅
```

---

## 🏗️ TEMPORAL WORKFLOWS (Long-Running Processes)

### Exemplo: Lead Nurturing (30 dias)

```
ContactCreated event → RabbitMQ
    ↓
Python Temporal Worker inicia workflow:
    LeadNurturingWorkflow(contact_id, project_id)
    ↓
Day 1: SendWelcomeEmail (activity)
    ↓
Wait 3 days...
    ↓
Day 3: CheckEngagement (activity)
    IF engagement_score < 5:
        TriggerAIOutreach (activity)
            ↓
            ADK Agent: SalesProspectingAgent
                - Busca memória (gRPC → Go)
                - Analisa comportamento
                - Gera mensagem personalizada
                - Publica OutboundMessage → RabbitMQ
    ↓
Wait 4 days...
    ↓
Day 7: AILeadQualification (activity)
    ADK Agent: SalesProspectingAgent
        - Analisa histórico (memória)
        - Qualifica usando BANT
        - Retorna: {is_qualified: true, score: 8.5}
    IF qualified:
        MovePipelineStage("qualified")
    ↓
Wait 7 days...
    ↓
Day 14: AIRetentionCheck (activity)
    ↓
Wait 16 days...
    ↓
Day 30: AssignToHumanAgent (activity)
        ↓
    Workflow completo ✅
```

**Temporal mantém estado durante 30 dias, sobrevive a deploys/crashes**

---

## 🎯 AGENT ENTITY CREATION

### Quem Cria o Quê?

```
┌─────────────────────────────────────────────────────────────┐
│  GO - OWNS AGENT ENTITIES                                    │
│  ✅ Agent aggregate (agents table)                           │
│  ✅ CRUD operations                                           │
│  ✅ AIAgentMetadata (JSONB column)                           │
│  ✅ Discovery & routing                                       │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ gRPC: ListAgentTemplates()
                       │        GetAgentTemplate(id)
                       ▼
┌─────────────────────────────────────────────────────────────┐
│  PYTHON - EXPÕE TEMPLATES GENÉRICOS                          │
│  ✅ AgentTemplateRegistry                                    │
│  ✅ Templates pré-configurados:                              │
│     - sales_prospecting                                      │
│     - retention_churn                                        │
│     - support_technical                                      │
│     - support_billing                                        │
│     - balanced (fallback)                                    │
│  ✅ NO persistence (stateless)                               │
└─────────────────────────────────────────────────────────────┘
```

### Fluxo de Criação

```
Admin quer criar novo agent de retenção
    ↓
Frontend: "Criar agent baseado em template"
    ↓
Go API: ListAvailableAgentTemplates()
    ↓
gRPC call → Python: ListAgentTemplates()
    ↓
Python retorna templates disponíveis
    ↓
Admin seleciona: "retention_churn"
    ↓
Go API: CreateAgentFromTemplate("retention_churn", "Retenção Bot", project_id)
    ↓
gRPC call → Python: GetAgentTemplate("retention_churn")
    ↓
Python retorna config completo:
    {
        "template_id": "retention_churn",
        "name": "Retention & Churn Prevention",
        "category": "retention",
        "knowledge_scope": {
            "lookback_days": 90,
            "include_agent_transfers": true,
            ...
        },
        "retrieval_strategy": "retention_churn", // 50% vector, 20% keyword, 20% graph
        "instruction_prompt": "Você é um especialista em retenção...",
        "required_tools": ["search_memory", "create_retention_offer", ...]
    }
    ↓
Go cria Agent entity:
    agent := domain.NewAgent("Retenção Bot", AgentTypeAI, project_id)
    agent.SetAIMetadata(convertedMetadata)
    agentRepo.Save(agent)
    ↓
Agent criado e ativo ✅
```

---

## 🔭 PHOENIX OBSERVABILITY

### O Que é Rastreado?

```
┌─────────────────────────────────────────────────────────────┐
│  PHOENIX DASHBOARD (http://localhost:6006)                   │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  📊 AGENT FLOW VISUALIZATION                                 │
│      CoordinatorAgent (125ms)                                │
│        ├─ search_memory (45ms)                               │
│        ├─ RetentionChurnAgent (380ms)                        │
│        │   ├─ search_memory (45ms)                           │
│        │   ├─ LLM call (280ms) - 1,245 tokens, $0.003       │
│        │   └─ create_retention_offer (35ms)                  │
│        └─ publish_message (15ms)                             │
│                                                               │
│  💰 TOKEN USAGE & COST                                        │
│      Total: 1,245 tokens                                     │
│      Input: 980 tokens (cached: 750, 90% saving)            │
│      Output: 265 tokens                                      │
│      Cost: $0.003                                            │
│                                                               │
│  🎯 RETRIEVAL QUALITY                                         │
│      Query: "Customer dissatisfaction and pricing"           │
│      Retrieved: 5 sessions                                   │
│      Avg similarity: 0.82                                    │
│      Reranking: 0.89 → 0.91 (+2% improvement)               │
│                                                               │
│  🚨 HALLUCINATION DETECTION                                   │
│      Response: "25% desconto por 3 meses"                    │
│      Grounded in context: ✅ Yes (discount policy: max 30%)  │
│      Hallucination score: 0.02 (low)                        │
│                                                               │
│  📈 EMBEDDING SPACE (UMAP)                                    │
│      [Visualization of all session embeddings clustered]     │
│      Cluster 1: Price objections (45 sessions)              │
│      Cluster 2: Technical issues (67 sessions)              │
│      Cluster 3: Onboarding questions (32 sessions)          │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

---

## ⚙️ SAGA PATTERNS (Temporal + ADK)

### Exemplo: Customer Onboarding

```
SubscriptionCreated event → RabbitMQ
    ↓
Python Temporal Worker inicia saga:
    CustomerOnboardingSaga(customer_id)
    ↓
Try:
    Step 1: CreateAccount
        ↓ Success → Register compensation: DeleteAccount
    Step 2: SetupBilling
        ↓ Success → Register compensation: CancelBilling
    Step 3: AIWelcome (ADK agent)
        ↓ ADK Agent: OnboardingAgent
            - Busca perfil do cliente (memória)
            - Gera boas-vindas personalizada
            - Publica mensagem
        ↓ Success → Register compensation: SendApology
    Step 4: ProvisionResources
        ↓ Success → Register compensation: DeprovisionResources
    Step 5: ActivateSubscription
        ↓ Success → Register compensation: SuspendSubscription
        ↓
    Saga completo ✅

Catch Exception:
    Execute compensations em ordem reversa:
        SuspendSubscription → DeprovisionResources → SendApology → CancelBilling → DeleteAccount
        ↓
    Rollback completo ✅
```

---

## 🔄 DEGRADAÇÃO GRACEFUL

### Cenário: AI Service está DOWN

```
1. Cliente envia mensagem → WAHA → Go ✅
2. Go salva message + session ✅
3. Go publica MessageReceived → RabbitMQ ✅
4. Evento fica na fila (Python não consome) ⏳
5. Sistema continua funcionando (sem AI responses)

Quando AI Service volta:
    ↓
1. Python conecta no RabbitMQ ✅
2. Consome backlog de eventos ✅
3. Processa em ordem ✅
4. Sistema volta ao normal ✅
```

**Usuário não perde mensagens, apenas delay na resposta AI**

---

## 📊 RETRIEVAL STRATEGIES DICTIONARY

### Exemplos Práticos

| Strategy | Weights | Use Case | Exemplo |
|----------|---------|----------|---------|
| **retention_churn** | 50% vector, 20% keyword, 20% graph, 10% recent | Cliente quer cancelar | "Muito caro" → Busca padrões similares de churn |
| **sales_prospecting** | 20% vector, 30% keyword, 40% graph, 10% recent | Lead qualification | Prioriza origem (UTM, campaign attribution) |
| **support_technical** | 30% vector, 50% keyword, 10% graph, 10% recent | Bug reports | Keywords exatos ("erro 500", "não carrega") |
| **balanced** | 33% vector, 33% keyword, 33% graph | Fallback genérico | Quando intent não é claro |
| **vector_keyword_50** | 50% vector, 50% keyword | Standard RAG | Classic hybrid search |

**SQL Baseline:** SEMPRE inclui últimas 20 mensagens (independente da strategy)

---

## 🗄️ DATA ARCHITECTURE: PostgreSQL vs BigQuery

### **Cenário 1: PostgreSQL (Operacional)**

```sql
-- Colunas tipadas para queries rápidas
CREATE TABLE memory_embeddings (
    id UUID PRIMARY KEY,
    tenant_id UUID,
    contact_id UUID,  -- ← Coluna
    document_id UUID,  -- ← Coluna (para JOINs)
    document_name TEXT,  -- ← Coluna (para ILIKE)
    content_type TEXT,
    embedding vector(768),
    metadata JSONB DEFAULT '{}'  -- Mínimo
);

-- Query rápida
SELECT * FROM memory_embeddings
WHERE contact_id = 'x' AND document_name ILIKE '%contrato%';
```

**Use cases:** AI Agent queries, real-time search, JOINs

### **Cenário 2: BigQuery (BI/Analytics)**

```sql
-- Metadata estratégico para flexibilidade
CREATE TABLE embeddings_warehouse (
    id STRING,
    embedding ARRAY<FLOAT64>,
    metadata JSON,  -- TUDO aqui
    created_at TIMESTAMP
)
PARTITION BY DATE(created_at);

-- metadata:
{
    "contact_id": "x",
    "document_id": "y",
    "document_name": "Contrato.pdf",
    "amount_extracted": 10000,
    "entities": [...],
    "campaign_source": "google_ads"
}

-- Query BI
SELECT
    JSON_VALUE(metadata.document_type) as type,
    COUNT(*) as count
FROM embeddings_warehouse
WHERE JSON_VALUE(metadata.contact_id) = 'x'
GROUP BY type;
```

**Use cases:** BI dashboards, historical analysis, data science

---

## 📄 CONTACT EVENTS → DOCUMENTS INTEGRATION

### **Eventos como Índice de Documentos (PostgreSQL)**

```
FLUXO:
1. Cliente envia PDF → WAHA → Go cria Message
   ↓
2. Go cria Contact Event:
   category: "document_received"
   summary: "Cliente enviou contrato"
   metadata: {
     document_name: "Contrato.pdf",
     document_id: "doc-uuid-789",
     document_type: "contract"
   }
   ↓
3. PDF → OCR → Chunks → Embeddings
   metadata: {
     source_document_id: "doc-uuid-789",  ← LINK ao evento
     source_event_id: "event-123",
     document_title: "Contrato.pdf",
     ...
   }
   ↓
4. AI Agent busca contexto:
   GetMemoryContext() retorna:
   - contact_events: ["Cliente enviou contrato em 2025-01-10"]
   - documents: [chunks vetorizados do Contrato.pdf]
   - Cross-reference: evento → document_id → embeddings
```

### **Benefícios:**

✅ **Timeline Visual**: Agent vê "quando" documentos foram enviados
✅ **Busca por Nome**: `WHERE metadata->>'document_name' ILIKE '%contrato%'`
✅ **Cross-Reference**: Evento → document_id → chunks vetorizados
✅ **Contexto Completo**: Evento (when) + Embeddings (what)

### **MCP Tool: get_contact_events_with_documents**

```python
# Python ADK
events = mcp_client.call_tool("get_contact_events_with_documents", {
    "contact_id": "contact-456",
    "event_categories": ["document_received"],
    "lookback_days": 30
})

# Result:
{
    "events": [
        {
            "event_id": "event-123",
            "summary": "Cliente enviou contrato",
            "created_at": "2025-01-10",
            "document_name": "Contrato.pdf",
            "document_id": "doc-uuid-789",
            "chunk_count": 15,
            "top_chunks": ["Valor: R$ 10k...", "Vigência: 12 meses..."]
        }
    ]
}
```

---

## 🚀 PERFORMANCE METRICS

### Latências Típicas

```
┌─────────────────────────────────┬──────────┬────────────┐
│ Operation                        │ Latency  │ Notes      │
├─────────────────────────────────┼──────────┼────────────┤
│ Message persistence (Go)         │ 5-10ms   │ PostgreSQL │
│ Embedding generation             │ 150-250ms│ Gemini     │
│ Hybrid search                    │ 50-150ms │ Cached     │
│   - Vector search (pgvector)     │ 20-50ms  │ HNSW index │
│   - Keyword search (pg_trgm)     │ 10-30ms  │ GIN index  │
│   - Graph traversal (AGE)        │ 50-100ms │ CYPHER     │
│ Reranking (Jina v2)             │ 80-120ms │ Optional   │
│ LLM inference (Gemini Flash)     │ 200-500ms│ Streaming  │
│ Full agent response              │ 500-1000ms│ End-to-end│
└─────────────────────────────────┴──────────┴────────────┘
```

### Cache Hit Rates

```
- Memory search cache: 70-90%
- Prompt caching (Gemini): 85-95%
- Embedding cache: 95%+ (sessões antigas)
```

### Cost Optimization

```
- Prompt caching: 90% redução em cached tokens
- Contextual retrieval: 67% menos erros (menos retries)
- RRF fusion: Sem tuning necessário (vs weighted)
- Reranking: Apenas quando necessário (high-value queries)
```

---

## ✅ STACK COMPLETO

### Go (Ventros CRM)
```
- Go 1.23+
- PostgreSQL 16 + pgvector + Apache AGE + pg_trgm
- Vertex AI Go SDK (Gemini Flash 2.5 + text-embedding-005)
- Redis 7+ (caching)
- gRPC (server)
- RabbitMQ (event publisher)
- Temporal (workflow client)
```

### Python (Ventros AI)
```
- Python 3.12+
- Google Cloud ADK 0.5+
- Vertex AI Python SDK
- gRPC (client para Go)
- RabbitMQ (event consumer)
- Temporal SDK (workflow + activity worker)
- Phoenix (observability)
- FastAPI (REST API)
```

---

## 🎯 RESUMO EM 5 PONTOS

1. **Loose Coupling:** Go e Python rodam independentes, comunicam via eventos/gRPC
2. **Memory-Aware Agents:** Python ADK agents buscam contexto híbrido no Go (vector + keyword + graph)
3. **Event-Driven:** RabbitMQ para async, gRPC para sync, Temporal para long-running
4. **Graceful Degradation:** Se AI está down, CRM continua funcionando normalmente
5. **Observability:** Phoenix rastreia tudo (LLM calls, embeddings, hallucinations, cost)

---

## 🔜 PRÓXIMOS PASSOS

### Fase 1: MVP (2-3 semanas)
1. ✅ Implementar gRPC API (Go)
2. ✅ Implementar VentrosMemoryService (Python)
3. ✅ Criar CoordinatorAgent + 1 specialist (RetentionChurnAgent)
4. ✅ Event consumer (RabbitMQ)
5. ✅ Phoenix setup

### Fase 2: Scale (4-6 semanas)
1. ✅ Adicionar mais specialists (Sales, Support, Operations)
2. ✅ Implementar Apache AGE knowledge graph
3. ✅ Memory facts extraction + contradiction resolution
4. ✅ Temporal workflows (Lead Nurturing, Onboarding)
5. ✅ Agent templates UI

### Fase 3: Production (8+ semanas)
1. ✅ Performance tuning (indexes, batch processing)
2. ✅ Monitoring & alerts
3. ✅ A/B testing de strategies
4. ✅ Cost optimization
5. ✅ Documentation & training

---

**Arquitetura moderna, elegante e production-ready! 🚀**
