# PROMPT: Avaliação Arquitetural Completa + Análise de Modelo de Dados + AI/ML + API Security - Ventros CRM

## CONTEXTO
Você é um **arquiteto de software sênior** especializado em:
- Sistemas distribuídos, DDD, CQRS, Event-Driven Architecture, Cloud Native patterns
- **Modelagem de Dados/Domínio** (PostgreSQL, normalização, integridade referencial)
- **AI/ML Engineering** (vector search, hybrid retrieval, multi-agent systems, LLM integration)
- **Python AI Agents** (Google ADK, Semantic Router, gRPC, RabbitMQ, Temporal) **seguindo DDD/Clean Arch**
- **API Design & Security** (RESTful, OWASP Top 10, authentication, rate limiting)
- **Testing Strategy** (unit, integration, E2E, coverage analysis)

Seu objetivo é realizar uma **auditoria arquitetural PROFUNDA e PRECISA** do projeto **Ventros CRM**, analisando:
1. **Backend Go** (200.000+ linhas, 600+ arquivos, 49 migrations) - DDD, CQRS, Event-Driven
2. **Modelo de Dados** (PostgreSQL schema, entidades, relacionamentos, normalização)
3. **API REST** (endpoints, security, error handling, documentation)
4. **AI/ML Features** (message enrichment, memory service, MCP server, Python ADK)
5. **Testing Coverage** (Go + Python, error handling, resilience)
6. **Python Project** (futuro multi-agent system com **DDD/Clean Arch/CQRS**)

---

## 🚨 INSTRUÇÃO CRÍTICA: LEIA 100% DO CÓDIGO GO

**ANTES de preencher qualquer tabela, você DEVE:**

1. **Ler TODOS os arquivos .go** do projeto:
   ```
   internal/domain/           # 94 arquivos - LEIA TODOS
   internal/application/      # Todos use cases
   infrastructure/           # 136 arquivos - LEIA TODOS
   cmd/                      # Entrypoints
   ```

2. **Analisar TODAS as 49 migrations**:
   ```
   infrastructure/database/migrations/000001_*.sql até 000049_*.sql
   ```

3. **Ler TODA a documentação**:
   ```
   AI_REPORT.md (580 linhas)
   TODO.md (1117 linhas)
   TODO_PYTHON.md (2797 linhas)
   docs/domain_mapping/ (15.000+ linhas - 23 aggregates)
   docs/MCP_SERVER_COMPLETE.md (1175 linhas)
   docs/PYTHON_ADK_ARCHITECTURE*.md (3000+ linhas)
   ```

4. **Mapear TODOS os 104+ Domain Events**
5. **Mapear TODOS os 26 GORM Repositories**
6. **Mapear TODOS os 27 HTTP Handlers**
7. **Mapear TODOS os relacionamentos entre entidades**

**NÃO faça suposições. NÃO preencha tabelas sem ler o código real.**

---

## ARQUIVOS DE REFERÊNCIA (LEIA OBRIGATORIAMENTE)

### Documentação Existente
```
AI_REPORT.md              # Status atual de AI/ML (2.5/10 - apenas enrichments básicos)
TODO.md                   # Roadmap principal (backend Go, Chat entity, WAHA integration)
TODO_PYTHON.md            # Roadmap Python ADK (6 fases, 18-27 semanas)

docs/
├── MCP_SERVER_COMPLETE.md               # MCP Server architecture (1175 linhas)
├── PYTHON_ADK_ARCHITECTURE.md           # Python ADK design (1000+ linhas)
├── PYTHON_ADK_ARCHITECTURE_PART2.md
├── PYTHON_ADK_ARCHITECTURE_PART3.md
├── AI_MEMORY_GO_ARCHITECTURE.md         # Memory service architecture
├── INTEGRATION_PLAN_MEMORY_GROUPS_DOCS.md
└── domain_mapping/                      # 23 aggregates documentados (15.000+ linhas)
    ├── README.md
    ├── contact_aggregate.md (500+ linhas)
    ├── session_aggregate.md (600+ linhas)
    ├── message_aggregate.md (700+ linhas)
    ├── pipeline_aggregate.md (500+ linhas)
    ├── agent_aggregate.md (400+ linhas)
    ├── channel_aggregate.md (600+ linhas)
    ├── billing_aggregate.md (900+ linhas)
    ├── webhook_aggregate.md (1100+ linhas)
    └── ... (15 more aggregates)
```

### Contexto Atual (2025-10-13)
**Backend Go**: 9.0/10 - Production-ready, enterprise-grade
- ✅ DDD + Clean Architecture
- ✅ CQRS (80+ commands, 20+ queries)
- ✅ Event-Driven (104+ events)
- ✅ Saga + Outbox Pattern (LISTEN/NOTIFY, <100ms)
- ✅ Optimistic Locking (8 agregados)
- ✅ 82% test coverage
- ✅ 26 GORM repositories, 27 HTTP handlers

**AI/ML Features**: 2.5/10 - Apenas enrichments básicos
- ✅ Message enrichment (12 providers: Gemini Vision, Groq Whisper, LlamaParse, FFmpeg)
- ❌ Memory Service (0% - hybrid search, vector embeddings, facts extraction)
- ❌ MCP Server (0% - apenas 1175 linhas de documentação)
- ❌ Python ADK (0% - apenas 3000+ linhas de documentação)
- ❌ gRPC API (0%)
- ❌ Knowledge Graph (Apache AGE 0%)
- ❌ Agent Templates (0%)

**Python Project**: 0% implementado, **MAS DEVE seguir DDD/Clean Arch/CQRS**
- Planejado: Multi-agent system (CoordinatorAgent + 5 specialists)
- Stack: Google ADK 0.5+, Semantic Router, gRPC, RabbitMQ, Temporal, Phoenix
- **CRITICAL**: Python ADK DEVE usar mesma arquitetura do Go:
  - Domain layer (agents como aggregates)
  - Application layer (use cases)
  - Infrastructure layer (gRPC, RabbitMQ, Temporal)
  - Separation of concerns (domain ≠ infrastructure)
- Effort: 18-27 semanas (4-6 meses)

---

## CRITÉRIOS DE AVALIAÇÃO (Referências Canônicas)

### SOLID Principles (Uncle Bob - Clean Code)
- **S**ingle Responsibility: Cada struct/função/class tem uma única razão para mudar?
- **O**pen/Closed: Extensível sem modificação?
- **L**iskov Substitution: Subtipos substituíveis?
- **I**nterface Segregation: Interfaces coesas e específicas?
- **D**ependency Inversion: Depende de abstrações, não de concreções?

### Domain-Driven Design (Eric Evans + Vaughn Vernon)
- Bounded Contexts bem definidos e isolados?
- Entities, Value Objects, Aggregates corretamente modelados?
- Domain Events capturando mudanças de estado?
- Ubiquitous Language refletido no código?
- Repositories abstraindo persistência?
- Domain Services vs Application Services claramente separados?
- **Invariantes de domínio protegidos nos Aggregates?**
- **Identity vs Equality corretamente implementados?**
- **Aggregates pequenos (2-3 entidades)?**
- **Transactional boundaries corretos?**

### Clean Architecture (Robert Martin)
- Camadas concêntricas respeitadas (Domain → Application → Infrastructure → API)?
- Regra de dependência: camadas internas não conhecem externas?
- Use Cases encapsulam lógica de aplicação?
- Entities do domínio puras, sem dependências de frameworks?
- Independência de frameworks, UI, DB?

### CQRS (Microsoft Azure Architecture)
- Separação clara entre Commands (write) e Queries (read)?
- Models diferentes para leitura e escrita?
- Eventual consistency gerenciado corretamente?
- Projeções/Read Models implementados?

### Event-Driven Architecture
- Domain Events vs Integration Events separados?
- Event Bus implementado (RabbitMQ)?
- Event Handlers desacoplados?
- Event Sourcing aplicado (parcial ou total)?
- Idempotência garantida?

### Saga Pattern (Microservices.io)
- Transações distribuídas coordenadas?
- Orquestração (centralizada) ou Coreografia (descentralizada)?
- Compensação de falhas implementada?

### Outbox Pattern (Microservices.io)
- Transações locais + publicação de eventos atômicas?
- Polling ou Transaction Log Tailing (LISTEN/NOTIFY)?
- Garantia de at-least-once delivery?

### Temporal Workflows
- Workflows orquestrando processos de longa duração?
- Activities idempotentes e retriáveis?
- Signal/Query patterns usados?
- Compensação/Saga implementada via Temporal?

### AI/ML Best Practices
- **Vector Search**: pgvector com índices IVFFlat otimizados?
- **Hybrid Retrieval**: Combinação de vector + keyword + graph + SQL?
- **RRF Fusion**: Reciprocal Rank Fusion implementado?
- **Memory Management**: Context caching, deduplication (SHA256)?
- **LLM Integration**: Vertex AI, error handling, retry logic?
- **Multi-Agent Systems**: Coordinator pattern, semantic routing?
- **Observability**: Phoenix tracing, metrics, logging?
- **Cost Management**: Context caching, token counting?

### RESTful API Design (Roy Fielding + Best Practices)
- **Resource-oriented**: URLs representam recursos (não ações)?
- **HTTP Methods**: GET/POST/PUT/PATCH/DELETE corretos?
- **Status Codes**: 2xx/3xx/4xx/5xx apropriados?
- **Idempotência**: GET/PUT/DELETE idempotentes?
- **Stateless**: Servidor não mantém estado de sessão?
- **HATEOAS**: Hypermedia as the Engine of Application State?
- **Versioning**: Estratégia de versionamento (URL/Header)?
- **Pagination**: Cursor-based ou offset-based?
- **Filtering**: Query parameters consistentes?
- **Sorting**: Sintaxe clara e documentada?

### API Security (OWASP Top 10 API Security 2023)
- **API1:2023 - Broken Object Level Authorization (BOLA)**: Verificação de ownership?
- **API2:2023 - Broken Authentication**: JWT, OAuth2, session management?
- **API3:2023 - Broken Object Property Level Authorization**: Mass assignment prevention?
- **API4:2023 - Unrestricted Resource Consumption**: Rate limiting, pagination limits?
- **API5:2023 - Broken Function Level Authorization (BFLA)**: RBAC implementado?
- **API6:2023 - Unrestricted Access to Sensitive Business Flows**: Anti-automation?
- **API7:2023 - Server Side Request Forgery (SSRF)**: Input validation?
- **API8:2023 - Security Misconfiguration**: CORS, headers, TLS?
- **API9:2023 - Improper Inventory Management**: API discovery, deprecation?
- **API10:2023 - Unsafe Consumption of APIs**: Validação de third-party APIs?

### Testing Strategy
- **Unit Tests**: Domain layer (70%+ coverage)?
- **Integration Tests**: Repository, HTTP, gRPC, RabbitMQ?
- **E2E Tests**: Full flow (webhook → processing → response)?
- **Error Handling**: Circuit breakers, retries, fallbacks?
- **Load Tests**: Performance benchmarks (locust, k6)?
- **Security Tests**: OWASP ZAP, penetration testing?

### Cloud Native & 12 Factor App
- Codebase único rastreado em VCS?
- Dependências explicitamente declaradas?
- Config em variáveis de ambiente?
- Backing services como recursos anexados?
- Stateless processes?
- Logs como event streams?
- Graceful shutdown?
- Dev/prod parity?

---

## TABELAS OBRIGATÓRIAS (30 TABELAS)

### TABELA 1: Avaliação Arquitetural Geral (0-10)

| Aspecto Arquitetural | Go Backend | Python ADK (Planejado) | Observações Críticas |
|---------------------|------------|------------------------|----------------------|
| **SOLID Principles** | | N/A | Verificar SRP em handlers grandes |
| **DDD - Bounded Contexts** | | N/A | Quantos BCs identificados? |
| **DDD - Aggregates & Entities** | | N/A | 23 aggregates documentados |
| **DDD - Value Objects** | | N/A | Primitive Obsession? |
| **DDD - Domain Events** | | | 104+ events, compartilhados via RabbitMQ |
| **DDD - Repositories** | | N/A | 26 repositories |
| **DDD - Invariantes de Domínio** | | N/A | Protegidos nos construtores? |
| **Clean Architecture - Camadas** | | | Verificar separação Go ↔ Python |
| **Use Cases / Application Services** | | | Quantos use cases? |
| **DTOs / API Contracts** | | | |
| **CQRS - Separação Command/Query** | | N/A | 80+ commands, 20+ queries |
| **CQRS - Read Models** | | N/A | Projeções implementadas? |
| **Event-Driven Architecture** | | | Go publica, Python consome |
| **Event Bus (RabbitMQ)** | | | 15+ queues, DLQ, retry |
| **Saga Pattern - Orquestração** | | | Temporal workflows |
| **Saga Pattern - Coreografia** | | | RabbitMQ events |
| **Outbox Pattern** | | N/A | LISTEN/NOTIFY, <100ms |
| **Temporal Workflows** | | | 7 workflows Go, planejado Python |
| **Temporal Activities** | | | Idempotentes? |
| **Postgres - Transações/Consistência** | | Via gRPC | 49 migrations |
| **Redis - Caching Strategy** | | | Repository cache implementado? |
| **Cloud Native - 12 Factors** | | | |
| **Error Handling & Resilience** | | | Circuit breakers? |
| **Observability (Logs/Metrics/Traces)** | | Phoenix | Structured logging? |
| **Testing Strategy** | | | 82% Go, 0% Python |
| **AI/ML - Vector Search (pgvector)** | | Via gRPC | Tabela faltando |
| **AI/ML - Hybrid Retrieval** | | Via gRPC | RRF Fusion faltando |
| **AI/ML - Memory Service** | | Via gRPC | 0% implementado |
| **AI/ML - LLM Integration** | | Google ADK | Vertex AI Go, Gemini Python |
| **AI/ML - Multi-Agent System** | N/A | | CoordinatorAgent + 5 specialists |
| **AI/ML - Semantic Routing** | N/A | | semantic-router lib |
| **AI/ML - Observability (Phoenix)** | N/A | | arize-phoenix |
| **gRPC API (Go ↔ Python)** | | | Communication layer 0% |
| **MCP Server (Go)** | | N/A | Claude Desktop 0% |
| **Modelo de Dados - Design** | | N/A | |
| **Modelo de Dados - Normalização** | | N/A | Formas normais? |
| **Modelo de Dados - Integridade** | | N/A | FK, constraints? |
| **Mapeamento ORM/Persistência** | | Via gRPC | GORM adapters |

**Legenda de Notas:**
- 0-3: Crítico/Ausente
- 4-5: Parcial/Inconsistente
- 6-7: Adequado/Funcional
- 8-9: Bom/Bem Estruturado
- 10: Excelente/Referência

---

### TABELA 2: Inventário e Análise de Entidades de Domínio (Go)

| Entidade de Domínio | Bounded Context | Tipo | Identidade | Invariantes Protegidos? | Complexidade | Rich/Anemic | Arquivo | Optimistic Lock? |
|---------------------|-----------------|------|------------|-------------------------|--------------|-------------|---------|------------------|
| Contact | CRM | Aggregate Root | UUID | | | | internal/domain/crm/contact/contact.go | |
| Message | CRM | Aggregate Root | UUID | | | | internal/domain/crm/message/message.go | |
| Session | CRM | Aggregate Root | UUID | | | | internal/domain/crm/session/session.go | |
| Project | Core | Aggregate Root | UUID | | | | internal/domain/core/project/project.go | |
| Agent | CRM | Aggregate Root | UUID | | | | internal/domain/crm/agent/agent.go | |
| Channel | CRM | Aggregate Root | UUID | | | | internal/domain/crm/channel/channel.go | |
| ChannelType | CRM | Aggregate Root | UUID | | | | internal/domain/crm/channel/channel_type.go | |
| Pipeline | CRM | Aggregate Root | UUID | | | | internal/domain/crm/pipeline/pipeline.go | |
| Chat | CRM | ❌ PLANEJADO | UUID | ❌ | ❌ | ❌ | ❌ NÃO EXISTE | ❌ |
| BillingAccount | Core/Billing | Aggregate Root | UUID | | | | internal/domain/core/billing/billing_account.go | |
| Subscription | Core/Billing | Entity | UUID | | | | internal/domain/core/billing/subscription.go | |
| Invoice | Core/Billing | Entity | UUID | | | | internal/domain/core/billing/invoice.go | |
| UsageMeter | Core/Billing | Entity | UUID | | | | internal/domain/core/billing/usage_meter.go | |
| Campaign | Automation | Aggregate Root | UUID | | | | internal/domain/automation/campaign/campaign.go | |
| Broadcast | Automation | Aggregate Root | UUID | | | | internal/domain/automation/broadcast/broadcast.go | |
| Sequence | Automation | Aggregate Root | UUID | | | | internal/domain/automation/sequence/sequence.go | |
| ContactList | CRM | Aggregate Root | UUID | | | | internal/domain/crm/contact_list/contact_list.go | |
| Note | CRM | Aggregate Root | UUID | | | | internal/domain/crm/note/note.go | |
| Tracking | CRM | Aggregate Root | UUID | | | | internal/domain/crm/tracking/tracking.go | |
| Credential | CRM | Aggregate Root | UUID | | | | internal/domain/crm/credential/credential.go | |
| Webhook | CRM | Aggregate Root | UUID | | | | internal/domain/crm/webhook/webhook.go | |
| MessageGroup | CRM | Aggregate Root | UUID | | | | internal/domain/crm/message_group/message_group.go | |
| ProjectMember | Core | Entity | UUID | | | | internal/domain/crm/project_member/project_member.go | |
| AgentSession | CRM | Entity | UUID | | | | (buscar arquivo) | |

**INSTRUÇÕES**:
- Leia CADA arquivo .go listado
- Verifique se invariantes são protegidos (constructor + métodos)
- Verifique Rich (tem comportamento) vs Anemic (só getters/setters)
- Verifique Optimistic Locking (campo `version`)
- Busque TODOS os aggregates (podem ter mais que 23)

---

### TABELA 3: Inventário e Análise de Entidades de Persistência (DB Schema)

| Tabela (DB) | Entidade de Domínio | Campos (count) | Índices (count) | Constraints (FK/UK/Check) | Soft Delete? | Auditoria? | Migration | Problemas |
|-------------|---------------------|----------------|-----------------|---------------------------|--------------|------------|-----------|-----------|
| contacts | Contact | | | | | | 000001 | |
| messages | Message | | | | | | 000002 | |
| sessions | Session | | | | | | 000003 | |
| projects | Project | | | | | | 000004 | |
| agents | Agent | | | | | | 000005 | |
| channels | Channel | | | | | | 000006 | |
| channel_types | ChannelType | | | | | | 000007 | |
| pipelines | Pipeline | | | | | | 000008 | |
| chats | ❌ Chat | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ NÃO EXISTE | 🔴 CRÍTICO |
| billing_accounts | BillingAccount | | | | | | 000020 | |
| subscriptions | Subscription | | | | | | 000021 | |
| invoices | Invoice | | | | | | 000022 | |
| usage_meters | UsageMeter | | | | | | 000023 | |
| campaigns | Campaign | | | | | | 000030 | |
| broadcasts | Broadcast | | | | | | 000031 | |
| sequences | Sequence | | | | | | 000032 | |
| contact_lists | ContactList | | | | | | 000010 | |
| notes | Note | | | | | | 000011 | |
| trackings | Tracking | | | | | | 000012 | |
| credentials | Credential | | | | | | 000013 | |
| webhooks | Webhook | | | | | | 000014 | |
| message_groups | MessageGroup | | | | | | 000036 | |
| message_enrichments | MessageEnrichment | | | | | | 000039 | |
| memory_embeddings | ❌ MemoryEmbedding | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ NÃO EXISTE | 🔴 CRÍTICO |
| memory_facts | ❌ MemoryFact | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ NÃO EXISTE | 🔴 CRÍTICO |
| retrieval_strategies | ❌ RetrievalStrategy | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ NÃO EXISTE | 🔴 CRÍTICO |
| domain_event_log | DomainEvent | | | | | | 000040 | |
| outbox_events | OutboxEvent | | | | | | 000041 | |
| ... (todas tabelas) | | | | | | | | |

**INSTRUÇÕES**:
- Leia TODAS as 49 migrations (000001-000049)
- Conte campos reais de cada tabela
- Conte índices (GIN, BTREE, UNIQUE, etc)
- Identifique constraints (FOREIGN KEY, UNIQUE, CHECK)
- Verifique soft delete (deleted_at)
- Verifique auditoria (created_at, updated_at)

---

### TABELA 4: Análise de Relacionamentos entre Entidades

| Entidade A | Entidade B | Tipo | Cardinalidade | Cascade Delete? | Integridade (DB/App)? | Navegabilidade | Problemas |
|------------|------------|------|---------------|-----------------|----------------------|----------------|-----------|
| Contact | Project | Association | N:1 | NO | DB (FK) | Unidirecional | |
| Message | Contact | Association | N:1 | NO | DB (FK) | Unidirecional | |
| Message | Session | Association | N:1 (optional) | NO | DB (FK NULL) | Unidirecional | |
| Message | Chat | ❌ FALTA | ❌ N:1 | ❌ | ❌ | ❌ | 🔴 FK faltando |
| Message | Agent | Association | N:1 (optional) | NO | DB (FK NULL) | Unidirecional | |
| Session | Contact | Association | N:1 | NO | DB (FK) | Unidirecional | |
| Session | Agent | Association | N:1 (optional) | NO | DB (FK NULL) | Unidirecional | |
| Channel | Project | Association | N:1 | CASCADE | DB (FK) | Unidirecional | |
| Channel | ChannelType | Association | N:1 | NO | DB (FK) | Unidirecional | |
| Subscription | BillingAccount | Composition | N:1 | CASCADE | DB (FK) | Bidirecional | |
| Invoice | BillingAccount | Composition | N:1 | CASCADE | DB (FK) | Unidirecional | |
| Campaign | Pipeline | Association | N:1 | NO | DB (FK) | Unidirecional | |
| Broadcast | ContactList | Association | N:1 | NO | DB (FK) | Unidirecional | |
| MessageGroup | Session | Association | N:1 | NO | DB (FK) | Unidirecional | |
| MessageEnrichment | Message | Association | N:1 | CASCADE | DB (FK) | Unidirecional | |
| Webhook | Project | Association | N:1 | CASCADE | DB (FK) | Unidirecional | |
| ... (todos relacionamentos) | | | | | | | |

**INSTRUÇÕES**:
- Leia migrations para identificar ALL foreign keys
- Verifique ON DELETE CASCADE/SET NULL/NO ACTION
- Identifique relacionamentos bidirecionais (rare in Go)
- Identifique navegabilidade (Go structs têm ponteiros?)

---

### TABELA 5: Análise de Aggregates (DDD)

| Aggregate Root | Entidades Filhas | Invariantes Principais | Transactional Boundary OK? | Tamanho | Otimização Necessária? |
|----------------|------------------|------------------------|----------------------------|---------|------------------------|
| Contact | ContactDetails (embedded) | Email/Phone únicos por tenant | ✅ | Pequeno | ✅ OK |
| Message | MessageContent (embedded) | Content não vazio | ✅ | Pequeno | ✅ OK |
| Session | SessionMetrics? | Timeout correto | ✅ | Pequeno | ✅ OK |
| Pipeline | PipelineStatus (list) | Status order correto | ✅ | Médio | ⚠️ Verificar |
| BillingAccount | Subscription, Invoice, UsageMeter | Billing logic | ⚠️ | Grande | 🔴 Pode ser muito grande |
| Campaign | CampaignRules? | State machine | ✅ | Médio | ⚠️ Verificar |
| MessageGroup | Messages (list) | Debounce timeout | ✅ | Médio | ⚠️ Pode crescer |
| ... (todos) | | | | | |

**INSTRUÇÕES**:
- Identifique entidades filhas dentro do aggregate
- Verifique se transação única protege invariantes
- **Aggregates grandes (>5 entidades)** são red flag
- Sugira quebrar aggregates muito grandes

---

### TABELA 6: Análise de Value Objects

| Value Object | Propriedades | Imutável? | Validação no Constructor? | Usado em Entities | Deveria ser VO mas não é? |
|--------------|--------------|-----------|---------------------------|-------------------|---------------------------|
| Email | string | ✅ | ✅ | Contact | |
| Phone | string | ✅ | ✅ | Contact | |
| Address | street, city, zip | | | Contact? | ⚠️ Verificar se existe |
| Money | amount, currency | | | Invoice? | ⚠️ Verificar se existe |
| MessageContent | text, type | | | Message | ⚠️ Verificar |
| SessionTimeout | duration | | | Session | ⚠️ Verificar |
| PipelineStage | name, order | | | Pipeline | ⚠️ Verificar |
| ... (todos VOs) | | | | | |

**INSTRUÇÕES**:
- Busque VOs em `internal/domain/*/` arquivos separados ou embedded
- Verifique imutabilidade (no setters, apenas constructor)
- Identifique **Primitive Obsession** (usar string ao invés de VO)
- Sugira VOs que deveriam existir

---

### TABELA 7: Análise de Normalização do Banco de Dados

| Tabela | Forma Normal Atual | Redundâncias Identificadas | Desnormalização Intencional? | Justificativa | Ação Recomendada |
|--------|-------------------|----------------------------|------------------------------|---------------|------------------|
| contacts | 3NF | Nenhuma | N/A | | ✅ OK |
| messages | 3NF | Nenhuma | N/A | | ✅ OK |
| sessions | 3NF | Nenhuma | N/A | | ✅ OK |
| pipelines | 3NF | PipelineStatus JSON | ⚠️ Sim | Performance reads | ⚠️ Verificar queries |
| message_groups | 3NF | Messages list cached? | ⚠️ | | ⚠️ Verificar |
| billing_accounts | 3NF | | | | |
| ... (todas tabelas) | | | | | |

**INSTRUÇÕES**:
- **1NF**: Valores atômicos (sem arrays em colunas não-JSONB)
- **2NF**: Sem dependência parcial da PK
- **3NF**: Sem dependência transitiva
- **BCNF**: Boyce-Codd Normal Form
- Desnormalização aceitável para: read models, caches, performance crítica

---

### TABELA 8: Análise de Mapeamento Domínio ↔ Persistência

| Entidade de Domínio | Entidade de Persistência (GORM) | Impedance Mismatch? | Mapper Implementado? | Qualidade Mapper | N+1 Queries? |
|---------------------|--------------------------------|---------------------|----------------------|------------------|--------------|
| Contact | ContactEntity | ⚠️ VOs? | ✅ | | |
| Message | MessageEntity | ⚠️ VOs? | ✅ | | ⚠️ Verificar |
| Session | SessionEntity | ⚠️ Metrics? | ✅ | | |
| Pipeline | PipelineEntity | ⚠️ Status list | ✅ | | ⚠️ Verificar |
| BillingAccount | BillingAccountEntity | ⚠️ Aggregates filhos | ✅ | | 🔴 Possível |
| Campaign | CampaignEntity | | ✅ | | |
| MessageGroup | MessageGroupEntity | | ✅ | | |
| ... (todos) | | | | | |

**INSTRUÇÕES**:
- Leia `infrastructure/persistence/entities/*.go`
- Leia mappers em `infrastructure/persistence/*_adapter.go`
- Verifique N+1 queries (usar GORM Preload?)
- Verifique se VOs são corretamente convertidos

---

### TABELA 9: Análise de Migrations e Evolução de Schema

| Migration | Data/Versão | Operação | Reversível? | Zero Downtime? | Problemas |
|-----------|-------------|----------|-------------|----------------|-----------|
| 000001_create_projects | | CREATE TABLE | ✅ | ✅ | |
| 000002_create_contacts | | CREATE TABLE | ✅ | ✅ | |
| 000036_create_message_groups | | CREATE TABLE | ✅ | ✅ | |
| 000038_add_debounce_timeout | | ALTER TABLE | ✅ | ✅ | |
| 000039_create_message_enrichments | | CREATE TABLE | ✅ | ✅ | |
| 000043_add_optimistic_locking | | ALTER TABLE | ✅ | ⚠️ | Lock contention? |
| 000048_add_system_agents | | INSERT | ✅ | ✅ | |
| 000049_add_played_at | | ALTER TABLE | ✅ | ✅ | |
| ❌ 000050_memory_embeddings | ❌ | ❌ | ❌ | ❌ | 🔴 NÃO EXISTE |
| ❌ 000051_memory_facts | ❌ | ❌ | ❌ | ❌ | 🔴 NÃO EXISTE |
| ❌ 000052_retrieval_strategies | ❌ | ❌ | ❌ | ❌ | 🔴 NÃO EXISTE |
| ❌ 000053_create_chats | ❌ | ❌ | ❌ | ❌ | 🔴 NÃO EXISTE |
| ... (todas 49 migrations) | | | | | |

**INSTRUÇÕES**:
- Leia TODAS as 49 migrations (.up.sql e .down.sql)
- Verifique reversibilidade (.down.sql existe e funciona?)
- Identifique migrations que podem causar downtime (ALTER TABLE locks)

---

### TABELA 10: Inventário de Use Cases

| Use Case | Camada | Status | Entidades | Aciona Eventos? | Usa Saga? | Usa Temporal? | Transação DB? | Complexidade |
|----------|--------|--------|-----------|-----------------|-----------|---------------|---------------|--------------|
| CreateContactUseCase | Application | ✅ | Contact | ✅ | ❌ | ❌ | ✅ | Baixa |
| UpdateContactUseCase | Application | ✅ | Contact | ✅ | ❌ | ❌ | ✅ | Baixa |
| CreateSessionUseCase | Application | ✅ | Session | ✅ | ❌ | ✅ | ✅ | Média |
| CloseSessionUseCase | Application | ✅ | Session | ✅ | ❌ | ✅ | ✅ | Média |
| SendMessageCommand | Application | ✅ | Message | ✅ | ✅ | ✅ | ✅ | Alta |
| ProcessInboundMessageSaga | Application | ✅ | Message, Contact, Session | ✅ | ✅ | ✅ | ✅ | Muito Alta |
| CreateCampaignHandler | Application | ✅ | Campaign | ✅ | ❌ | ❌ | ✅ | Média |
| UpdatePipelineStatusUseCase | Application | ✅ | Contact, Pipeline | ✅ | ❌ | ❌ | ✅ | Média |
| ... (todos use cases) | | | | | | | | |

**INSTRUÇÕES**:
- Busque em `internal/application/*/` TODOS os use cases
- Busque em `internal/application/commands/` TODOS os command handlers
- Busque em `internal/application/queries/` TODAS as queries
- Conte: quantos use cases existem?

---

### TABELA 11: Inventário de Domain Events

| Domain Event | Bounded Context | Entidade Origem | Publicado Via | Handlers | Armazenado (Outbox)? | Propaga? |
|--------------|-----------------|-----------------|---------------|----------|----------------------|----------|
| ContactCreated | CRM | Contact | Outbox | | ✅ | ✅ |
| ContactUpdated | CRM | Contact | Outbox | | ✅ | ✅ |
| MessageSent | CRM | Message | Outbox | | ✅ | ✅ |
| MessageReceived | CRM | Message | Outbox | | ✅ | ✅ |
| SessionCreated | CRM | Session | Outbox | | ✅ | ✅ |
| SessionClosed | CRM | Session | Outbox | | ✅ | ✅ |
| CampaignStateChanged | Automation | Campaign | Outbox | | ✅ | ✅ |
| ... (todos 104+ events) | | | | | | |

**INSTRUÇÕES**:
- Busque em `internal/domain/*/events.go` TODOS os events
- Verifique se são publicados via Outbox Pattern
- Identifique handlers em `infrastructure/messaging/*_consumer.go`
- **AI_REPORT.md diz 104+ events** - mapeie TODOS

---

### TABELA 12: Inventário de Integration/Application Events

| Integration Event | Origem | Destino(s) | Exchange/Queue | Retry Policy | DLQ | Idempotente? |
|-------------------|--------|------------|----------------|--------------|-----|--------------|
| message.inbound | WAHA Webhook | ProcessInboundMessageSaga | message.inbound | 3x exponential | ✅ | ✅ |
| message.outbound | SendMessageCommand | WAHA Client | message.outbound | 3x exponential | ✅ | ✅ |
| message.enrichment.requested | MessageDebouncerService | EnrichmentWorker | enrichment.queue | 3x | ✅ | ✅ |
| contact.qualified | Pipeline | CRM | contact.events | 3x | ✅ | ✅ |
| ... (todos) | | | | | | |

**INSTRUÇÕES**:
- Leia `infrastructure/messaging/` para identificar TODAS as queues
- Verifique configuração de retry (RabbitMQ)
- Verifique DLQ (Dead Letter Queue)
- **AI_REPORT.md diz 15+ queues** - mapeie TODAS

---

### TABELA 13: Mapeamento de Eventos para Projetos/Módulos

| Evento | Publicador (Módulo) | Consumidor(es) (Módulo) | Tipo (Domain/Integration) | Entidades Afetadas |
|--------|---------------------|-------------------------|---------------------------|--------------------|
| ContactCreated | Contact Aggregate | Lead Qualification, Email Service | Domain | Contact |
| MessageReceived | Message Aggregate | ProcessInbound Saga, Enrichment | Domain | Message, Contact, Session |
| SessionClosed | Session Aggregate | Analytics, Billing | Domain | Session, Agent |
| CampaignStateChanged | Campaign Aggregate | Broadcast Scheduler | Domain | Campaign, Broadcast |
| ... (todos) | | | | |

**INSTRUÇÕES**:
- Identifique publishers lendo domain aggregates
- Identifique consumers lendo `infrastructure/messaging/*_consumer.go`

---

### TABELA 14: Análise de Temporal Workflows

| Workflow | Activities | Entidades | Duração Típica | Compensação? | Signal/Query? | Caso de Uso |
|----------|-----------|----------|----------------|--------------|---------------|-------------|
| SessionManagementWorkflow | EndSessionActivity, CleanupSessionsActivity | Session | 15min-24h | ✅ | ✅ | Timeout de sessão |
| ProcessInboundMessageSaga | EnrichMessage, SaveMessage, NotifyAgent | Message, Contact, Session | 5-30s | ✅ | ❌ | Webhook processing |
| MessageEnrichmentWorkflow | TranscribeAudio, OCRDocument, AnalyzeImage | Message, MessageEnrichment | 2-10s | ✅ | ❌ | AI enrichment |
| ... (todos 7 workflows) | | | | | | |

**INSTRUÇÕES**:
- Leia `internal/workflows/` para identificar workflows
- Leia `internal/workflows/*/activities.go` para activities
- **AI_REPORT.md diz 7 workflows** - mapeie TODOS

---

### TABELA 15: Análise de Queries e Performance

| Query Crítica | Entidades | Índices Usados? | N+1? | Eager/Lazy? | Paginação? | Tempo Esperado | Otimização? |
|---------------|-----------|-----------------|------|-------------|------------|----------------|-------------|
| ListContactsByProject | Contact | idx_contacts_project | ❌ | Lazy | ✅ Cursor | <50ms | ✅ OK |
| GetConversationHistory | Message, Contact | idx_messages_session | ⚠️ | Eager | ✅ Offset | <100ms | ⚠️ Verificar |
| SearchContactsByPhone | Contact | idx_contacts_phone | ❌ | Lazy | ✅ | <30ms | ✅ OK |
| GetPipelineWithStatuses | Pipeline, PipelineStatus | idx_pipelines_project | ⚠️ | Eager | ❌ | <80ms | ⚠️ N+1? |
| GetActiveSessionsByAgent | Session | idx_sessions_agent_status | ❌ | Lazy | ✅ | <50ms | ✅ OK |
| ... (queries principais) | | | | | | | |

**INSTRUÇÕES**:
- Leia `internal/application/queries/*.go` para identificar queries
- Verifique se há índices apropriados (leia migrations)
- Identifique N+1 queries (loop chamando repository)

---

### TABELA 16: Consistência de Dados e Transações

| Operação Crítica | Entidades | Padrão | Garantias | Riscos | Transação Atômica? |
|------------------|-----------|--------|-----------|--------|-------------------|
| Criar Contato + Publicar Evento | Contact, domain_events | Outbox Pattern | At-least-once | ✅ Nenhum | ✅ Sim |
| Processar Webhook + Salvar Mensagem | Message, Contact, Session | Saga (Temporal) | Compensação | ⚠️ Idempotência | ✅ Sim |
| Atualizar Pipeline + Enviar Email | Contact, Pipeline | Event-Driven | Eventual consistency | ⚠️ Email pode falhar | ✅ Sim (DB) |
| Gerar Embedding + Armazenar | ❌ MemoryEmbedding | ❌ N/A | ❌ | 🔴 Não implementado | N/A |
| Python Agent → gRPC → DB | Contact, Message | ❌ N/A | ❌ | 🔴 Não implementado | N/A |
| ... (operações críticas) | | | | | |

**INSTRUÇÕES**:
- Identifique operações que envolvem múltiplas entidades
- Verifique se usam transação DB (GORM Begin/Commit)
- Verifique padrões (Outbox, Saga, Event-Driven)

---

### TABELA 17: Análise de Validações e Business Rules

| Regra de Negócio | Localização | Entidades | Implementada? | Falha em qual Cenário? | Deveria estar em? |
|------------------|-------------|-----------|---------------|------------------------|-------------------|
| Email único por projeto | Domain | Contact | ✅ | Duplicate email | ✅ Domain OK |
| Mensagem não pode ser vazia | Domain | Message | ✅ | Empty content | ✅ Domain OK |
| Session timeout válido | Domain | Session | ✅ | Invalid duration | ✅ Domain OK |
| Pipeline stage order | Domain | Pipeline | ⚠️ | Verificar | ⚠️ |
| Campaign state machine | Domain | Campaign | ⚠️ | Verificar | ⚠️ |
| ... (todas regras) | | | | | |

**INSTRUÇÕES**:
- Leia domain aggregates para identificar business rules
- Verifique se estão no **Domain Layer** (não em handlers)
- Identifique regras que estão no lugar errado (API/Infrastructure)

---

### TABELA 18: Análise de DTOs e Serialização

| DTO | Entidade Origem | Usado em | Campos Expostos | Validações | Over/Under-fetching? | Dados Sensíveis? |
|-----|-----------------|----------|-----------------|------------|----------------------|------------------|
| ContactDTO | Contact | API | id, name, email, phone | ✅ | ⚠️ | ⚠️ Email/Phone |
| MessageDTO | Message | API | id, content, direction | ✅ | ✅ OK | ❌ |
| SessionDTO | Session | API | id, status, metrics | ✅ | ✅ OK | ❌ |
| PipelineDTO | Pipeline | API | id, name, statuses | ✅ | ⚠️ Over? | ❌ |
| ... (todos DTOs) | | | | | | |

**INSTRUÇÕES**:
- Leia `infrastructure/http/dto/*.go`
- Verifique se DTOs expõem apenas campos necessários
- Identifique dados sensíveis (passwords, tokens, etc)

---

### TABELA 19: API Endpoints Inventory & RESTful Design

| Endpoint | HTTP Method | Resource | RESTful? | Idempotente? | Status Codes | Authentication | Authorization (RBAC) | Pagination | Versioning |
|----------|-------------|----------|----------|--------------|--------------|----------------|---------------------|------------|------------|
| POST /api/v1/contacts | POST | Contact | ✅ | ❌ | 201, 400, 409 | JWT | ✅ | N/A | ✅ v1 |
| GET /api/v1/contacts | GET | Contact | ✅ | ✅ | 200, 401 | JWT | ✅ | ✅ Cursor | ✅ v1 |
| GET /api/v1/contacts/:id | GET | Contact | ✅ | ✅ | 200, 404 | JWT | ✅ | N/A | ✅ v1 |
| PUT /api/v1/contacts/:id | PUT | Contact | ✅ | ✅ | 200, 404 | JWT | ✅ | N/A | ✅ v1 |
| PATCH /api/v1/contacts/:id | PATCH | Contact | ✅ | ❌ | 200, 404 | JWT | ✅ | N/A | ✅ v1 |
| DELETE /api/v1/contacts/:id | DELETE | Contact | ✅ | ✅ | 204, 404 | JWT | ✅ | N/A | ✅ v1 |
| POST /api/v1/messages/send | POST | Message | ⚠️ Action-based | ❌ | 201, 400 | JWT | ✅ | N/A | ✅ v1 |
| GET /api/v1/sessions/active | GET | Session | ⚠️ Filter in URL? | ✅ | 200 | JWT | ✅ | ✅ | ✅ v1 |
| POST /api/webhooks/waha | POST | Webhook | N/A | ⚠️ | 200, 400 | HMAC | ❌ Public | N/A | N/A |
| ... (TODOS endpoints) | | | | | | | | | |

**INSTRUÇÕES**:
- Leia `infrastructure/http/routes/routes.go` para mapear TODOS endpoints
- Leia `infrastructure/http/handlers/*.go` para verificar implementação
- Verifique:
  - RESTful design (resource-oriented, não action-based)
  - Idempotência (GET/PUT/DELETE devem ser idempotentes)
  - Status codes corretos (2xx success, 4xx client error, 5xx server error)
  - Authentication (JWT? API Key? OAuth2?)
  - Authorization (RBAC implementado? middleware/rbac.go)
  - Pagination (cursor-based ou offset-based?)
  - Versioning (URL path /v1? Header? Query param?)

**Anti-patterns para identificar**:
- ❌ `/api/contacts/getAllActive` (ação no URL, deveria ser GET /api/contacts?status=active)
- ❌ POST para operações de leitura
- ❌ GET para operações que mudam estado
- ❌ Falta de versionamento
- ❌ Paginação inconsistente

---

### TABELA 20: API Security Assessment (OWASP Top 10 API Security 2023)

| OWASP API Security Risk | Mitigação Implementada? | Localização | Evidência de Proteção | Vulnerabilidades Identificadas | Ação Corretiva |
|------------------------|-------------------------|-------------|----------------------|-------------------------------|----------------|
| **API1:2023 - BOLA (Broken Object Level Authorization)** | | middleware/rbac.go? | Verificar ownership checks | | |
| **API2:2023 - Broken Authentication** | | middleware/jwt_auth.go | JWT validation, expiry | | |
| **API3:2023 - Broken Object Property Level Authorization** | | | Mass assignment prevention? | | |
| **API4:2023 - Unrestricted Resource Consumption** | | middleware/rate_limit.go | Rate limiting config | | |
| **API5:2023 - BFLA (Broken Function Level Authorization)** | | middleware/rbac.go | Role-based access control | | |
| **API6:2023 - Unrestricted Access to Sensitive Business Flows** | | | Anti-automation? CAPTCHA? | | |
| **API7:2023 - SSRF (Server Side Request Forgery)** | | | Input validation, URL whitelist | | |
| **API8:2023 - Security Misconfiguration** | | | CORS, headers, TLS config | | |
| **API9:2023 - Improper Inventory Management** | | | API docs, deprecation policy | | |
| **API10:2023 - Unsafe Consumption of APIs** | | | Third-party API validation (WAHA, Stripe, Vertex) | | |

**INSTRUÇÕES**:
- **API1 (BOLA)**: Verifique se há checks de ownership antes de retornar recursos
  - Exemplo: `GET /contacts/:id` deve verificar se contact pertence ao project do user
  - Busque em handlers: `if contact.ProjectID != user.ProjectID { return 403 }`

- **API2 (Authentication)**: Leia `infrastructure/http/middleware/jwt_auth.go`
  - JWT signature validation?
  - Token expiry verificado?
  - Refresh token implementado?

- **API3 (Mass Assignment)**: Verifique se DTOs limitam campos que podem ser atualizados
  - Exemplo: User não pode setar `is_admin: true` via API

- **API4 (Resource Consumption)**: Leia `infrastructure/http/middleware/rate_limit.go`
  - Rate limiting por IP? Por user?
  - Limites de pagination (max 100 itens?)
  - Timeout em requests longos?

- **API5 (BFLA)**: Leia `infrastructure/http/middleware/rbac.go`
  - RBAC implementado?
  - Roles: admin, agent, viewer?
  - Permissões verificadas antes de actions críticas?

- **API7 (SSRF)**: Verifique input validation em:
  - Webhook URLs (user pode configurar webhook malicioso?)
  - Image/Document URLs (enriquecimento de mídia)

- **API8 (Misconfiguration)**:
  - CORS configurado? (`Access-Control-Allow-Origin: *` é perigoso)
  - Security headers? (`X-Frame-Options`, `X-Content-Type-Options`, `Strict-Transport-Security`)
  - TLS enforced?

- **API10 (Unsafe Consumption)**:
  - WAHA client: valida responses? Timeout? Retry?
  - Stripe client: signature verification?
  - Vertex AI: input sanitization?

**Vulnerabilidades Comuns para Buscar**:
- ❌ Falta de ownership check (qualquer user pode acessar qualquer contact)
- ❌ JWT sem expiry ou com expiry muito longo (>24h)
- ❌ Rate limiting desabilitado ou muito permissivo
- ❌ CORS `Access-Control-Allow-Origin: *`
- ❌ Senhas/tokens em logs
- ❌ SQL injection (mesmo com ORM, verificar raw queries)
- ❌ SSRF via webhook URL configurável

---

### TABELA 21: API Rate Limiting & DDoS Protection

| Endpoint/Resource | Rate Limit Configurado? | Limites (req/s ou req/min) | Escopo (IP/User/Global) | Circuit Breaker? | Throttling Strategy | DDoS Mitigation |
|-------------------|------------------------|----------------------------|------------------------|------------------|---------------------|-----------------|
| POST /api/v1/contacts | | | | | | |
| GET /api/v1/contacts | | | | | | |
| POST /api/v1/messages/send | | | | | | |
| POST /api/webhooks/waha | | | | | | |
| POST /api/webhooks/stripe | | | | | | |
| GET /api/v1/sessions | | | | | | |
| ... (todos endpoints críticos) | | | | | | |

**INSTRUÇÕES**:
- Leia `infrastructure/http/middleware/rate_limit.go`
- Identifique:
  - Limites por IP? Por user? Global?
  - Algoritmo: Token Bucket? Leaky Bucket? Fixed Window? Sliding Window?
  - Headers de resposta: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `Retry-After`?
  - Endpoints públicos (webhooks) têm rate limiting mais agressivo?
  - Circuit breaker implementado? (`infrastructure/messaging/rabbitmq_circuit_breaker.go`)

**Ataques para Mitigar**:
- 🔴 **DDoS** via webhook floods (WAHA, Stripe)
- 🔴 **Brute-force** em autenticação
- 🔴 **Resource exhaustion** em queries pesadas (GET /messages)
- 🔴 **API abuse** (scraping contacts)

---

### TABELA 22: API Error Handling & Response Contracts

| Error Type | HTTP Status | Error Response Format | User-Friendly Message? | Logged? | Stack Trace Exposed? | Retry Strategy |
|------------|-------------|----------------------|------------------------|---------|---------------------|----------------|
| Validation Error | 400 | | | | | |
| Authentication Error | 401 | | | | | |
| Authorization Error | 403 | | | | | |
| Not Found | 404 | | | | | |
| Conflict (Duplicate) | 409 | | | | | |
| Rate Limit Exceeded | 429 | | | | | |
| Internal Server Error | 500 | | | | | |
| Service Unavailable | 503 | | | | | |

**INSTRUÇÕES**:
- Leia `infrastructure/http/errors/api_error.go`
- Leia `infrastructure/http/middleware/error_handler.go`
- Verifique formato de erro padrão:
  ```json
  {
    "error": {
      "code": "RESOURCE_NOT_FOUND",
      "message": "Contact not found",
      "details": {},
      "request_id": "uuid"
    }
  }
  ```
- Verifique se stack traces são expostos (🔴 security risk)
- Verifique se erros internos são loggados mas não expostos ao cliente

**Anti-patterns**:
- ❌ Stack trace no response (expõe código interno)
- ❌ Mensagens técnicas para user ("sql: no rows in result set")
- ❌ Erro genérico ("Internal Server Error") sem details
- ❌ Falta de `request_id` para troubleshooting

---

### TABELA 23: Inventário de AI/ML Components

| Component | Status | Language | Dependencies | Implementado? | Gaps Críticos |
|-----------|--------|----------|--------------|---------------|---------------|
| **Message Enrichment** | ✅ Funcional | Go | Vertex AI, Groq, LlamaParse, FFmpeg | 85% | Resultados não usados em memory |
| **Memory Embeddings** | ❌ Ausente | Go | Vertex AI, pgvector | 0% | Tabela não existe |
| **Hybrid Search** | ❌ Ausente | Go | pgvector, PostgreSQL FTS | 0% | RRF Fusion faltando |
| **Vector Search** | ❌ Ausente | Go | pgvector IVFFlat | 0% | Queries faltando |
| **Keyword Search** | ❌ Ausente | Go | PostgreSQL FTS (tsvector) | 0% | Full-text index faltando |
| **Graph Search** | ❌ Ausente | Go | Apache AGE | 0% | AGE não instalado |
| **Memory Facts Extraction** | ❌ Ausente | Go | Vertex AI (NER) | 0% | Tabela não existe |
| **gRPC Server** | ❌ Ausente | Go | gRPC, protobuf | 0% | Proto definitions faltando |
| **MCP Server** | ❌ Ausente | Go | HTTP, SSE | 0% | Zero código (apenas 1175 linhas docs) |
| **Python ADK** | ❌ Ausente | Python | Google ADK 0.5+ | 0% | Projeto não iniciado |
| **Coordinator Agent** | ❌ Ausente | Python | Google ADK | 0% | Não implementado |
| **Specialist Agents (5x)** | ❌ Ausente | Python | Google ADK | 0% | Não implementado |
| **Semantic Router** | ❌ Ausente | Python | semantic-router | 0% | Não implementado |
| **RabbitMQ Consumer (Python)** | ❌ Ausente | Python | pika | 0% | Não implementado |
| **gRPC Client (Python)** | ❌ Ausente | Python | grpcio | 0% | Proto generation faltando |
| **Phoenix Observability** | ❌ Ausente | Python | arize-phoenix | 0% | Não configurado |

**INSTRUÇÕES**:
- Leia `infrastructure/ai/*.go` para components existentes
- Verifique cada provider listado em AI_REPORT.md
- Identifique gaps críticos (memory, gRPC, Python ADK)

---

### TABELA 24: Análise de Testing Coverage (Go + Python)

| Layer/Component | Go Backend | Python ADK | Type Coverage | Missing Tests | Priority |
|-----------------|------------|------------|---------------|---------------|----------|
| **Domain Layer** | 82% ✅ | N/A | Unit | Customer (23.6%), Project (42.3%), Shared (46.1%) | P1 |
| **Application Layer** | ? | N/A | Unit + Integration | Use cases sem tests | P0 |
| **Infrastructure - Repositories** | ✅ Tests passing | N/A | Integration | Repository mocks | P1 |
| **Infrastructure - HTTP Handlers** | ? | N/A | Integration | Handlers sem tests | P0 |
| **Infrastructure - RabbitMQ** | ✅ 7/7 passing | N/A | Integration | Event consumers | P1 |
| **Infrastructure - Temporal** | ✅ 3/3 passing | N/A | Integration | Workflows complexos | P1 |
| **API - E2E** | Parcial | N/A | E2E | Cobertura incompleta | P0 |
| **AI - Message Enrichment** | ? | N/A | Integration | Providers sem tests | P1 |
| **AI - Memory Service** | N/A | N/A | Integration | Não implementado | P0 |
| **AI - gRPC API** | N/A | N/A | Integration | Não implementado | P0 |
| **Python - Multi-Agent** | N/A | 0% | Unit + Integration | Não implementado | P0 |
| **Python - Semantic Router** | N/A | 0% | Unit | Não implementado | P0 |
| **Python - gRPC Client** | N/A | 0% | Integration | Não implementado | P0 |
| **Python - RabbitMQ** | N/A | 0% | Integration | Não implementado | P0 |
| **Security Tests** | ❌ 0% | ❌ 0% | Security | OWASP ZAP, penetration | P0 |
| **Load Tests** | ❌ 0% | ❌ 0% | Performance | Benchmarks (locust, k6) | P2 |

**INSTRUÇÕES**:
- Rode `go test -cover ./...` para obter coverage real
- Identifique arquivos sem testes (`*_test.go` faltando)
- **Target: 70%+ em todas as camadas**

---

### TABELA 25: Análise de Error Handling & Resilience (Go + Python)

| Component | Circuit Breaker | Retry Logic | Fallback | Timeout | Dead Letter Queue | Error Logging | Grade (0-10) |
|-----------|----------------|-------------|----------|---------|-------------------|---------------|--------------|
| **RabbitMQ Consumer (Go)** | ✅ Sim | ✅ Sim (3x exponential) | ✅ Sim | ✅ Sim | ✅ DLQ | ✅ Structured | 10 |
| **HTTP Handlers (Go)** | ? | ? | ? | ✅ Context timeout | N/A | ✅ Structured | ? |
| **gRPC Server (Go)** | N/A | N/A | N/A | N/A | N/A | N/A | 0 |
| **Temporal Activities (Go)** | ✅ Built-in | ✅ Built-in | ✅ Compensation | ✅ Sim | N/A | ✅ Structured | 9 |
| **Vertex AI (Go)** | ? | ? | ? | ? | N/A | ? | ? |
| **PostgreSQL Queries (Go)** | ? | ✅ GORM retries | ? | ✅ Context | N/A | ? | ? |
| **Redis Operations (Go)** | ? | ? | ✅ Graceful degradation | ? | N/A | ? | ? |
| **WAHA Client (Go)** | ? | ? | ? | ✅ HTTP timeout | N/A | ? | ? |
| **Python - RabbitMQ Consumer** | N/A | N/A | N/A | N/A | N/A | N/A | 0 |
| **Python - gRPC Client** | N/A | N/A | N/A | N/A | N/A | N/A | 0 |
| **Python - LLM Calls (ADK)** | N/A | N/A | N/A | N/A | N/A | N/A | 0 |

**INSTRUÇÕES**:
- Leia `infrastructure/messaging/rabbitmq_circuit_breaker.go` (já implementado)
- Verifique error handling em cada component critical
- Identifique missing circuit breakers, retries, fallbacks

---

### TABELA 26: Python ADK Architecture Assessment (Design Quality)

| Aspect | Status | Design Quality | Implementação | Gaps Críticos |
|--------|--------|----------------|---------------|---------------|
| **Project Structure** | 📋 Documented | ✅ Excellent | ❌ 0% | Projeto não iniciado |
| **DDD/Clean Arch Compliance** | 📋 Documented | ⚠️ **MUST VERIFY** | ❌ 0% | **CRÍTICO**: Python DEVE seguir DDD |
| **Domain Layer (Agents as Aggregates)** | 📋 Documented | ⚠️ **NEEDS REVIEW** | ❌ 0% | Agents devem ser aggregates? |
| **Application Layer (Use Cases)** | 📋 Documented | ⚠️ **NEEDS DESIGN** | ❌ 0% | Process message = use case? |
| **Infrastructure Layer** | 📋 Documented | ✅ Good | ❌ 0% | gRPC, RabbitMQ, Temporal |
| **Multi-Agent System** | 📋 Documented | ✅ Excellent | ❌ 0% | CoordinatorAgent + 5 specialists |
| **Base Agent Class** | 📋 Documented | ✅ Good | ❌ 0% | Interface clara |
| **Semantic Router** | 📋 Documented | ✅ Excellent | ❌ 0% | Routes: sales, retention, support |
| **Memory Service Facade** | 📋 Documented | ✅ Good | ❌ 0% | gRPC client wrapper |
| **Tool Registry** | 📋 Documented | ✅ Good | ❌ 0% | CRM + Memory tools |
| **RabbitMQ Consumer** | 📋 Documented | ✅ Good | ❌ 0% | message.inbound queue |
| **RabbitMQ Publisher** | 📋 Documented | ✅ Good | ❌ 0% | message.outbound exchange |
| **Temporal Workflows** | 📋 Documented | ✅ Good | ❌ 0% | Agent workflows |
| **Phoenix Observability** | 📋 Documented | ✅ Good | ❌ 0% | Tracing setup |
| **Testing Strategy** | 📋 Documented | ✅ Good | ❌ 0% | Unit + Integration + E2E |
| **Dependency Management** | 📋 Documented | ✅ Excellent | ❌ 0% | Poetry + pyproject.toml |
| **Type Safety** | 📋 Documented | ✅ Excellent | ❌ 0% | mypy strict mode |
| **Error Handling** | 📋 Documented | ⚠️ Incomplete | ❌ 0% | Circuit breaker, retry faltando |
| **Configuration** | 📋 Documented | ✅ Good | ❌ 0% | Pydantic Settings |
| **Cost Management (LLM)** | 📋 Documented | ⚠️ Incomplete | ❌ 0% | Context caching, token counting faltando |

**CRÍTICO**: Python ADK DEVE seguir mesmos padrões do Go:
- ✅ **Domain Layer**: Agents como aggregates, invariantes protegidos
- ✅ **Application Layer**: Use cases (ProcessMessageUseCase, RouteToSpecialistUseCase)
- ✅ **Infrastructure Layer**: gRPC, RabbitMQ, Temporal adapters
- ✅ **Separation of Concerns**: Domain puro, sem dependências de infra
- ✅ **Testing**: 70%+ coverage (unit + integration + E2E)
- ✅ **Error Handling**: Circuit breakers, retries, fallbacks

**INSTRUÇÕES**:
- Leia `docs/PYTHON_ADK_ARCHITECTURE*.md` (3000+ linhas)
- **VALIDE se design está conforme DDD/Clean Arch**
- Sugira refatorações se necessário
- **PESQUISE Google ADK 0.5+ docs** (https://ai.google.dev/)

---

### TABELA 27: gRPC API Design Assessment

| Aspect | Proto Defined? | Go Server | Python Client | Security | Performance | Documentation |
|--------|---------------|-----------|---------------|----------|-------------|---------------|
| **memory_service.proto** | ❌ No | ❌ No | ❌ No | N/A | N/A | ✅ TODO_PYTHON.md |
| **SearchMemory RPC** | ❌ No | ❌ No | ❌ No | N/A | N/A | ✅ Designed |
| **StoreEmbedding RPC** | ❌ No | ❌ No | ❌ No | N/A | N/A | ✅ Designed |
| **GetContactContext RPC** | ❌ No | ❌ No | ❌ No | N/A | N/A | ✅ Designed |
| **ExtractFacts RPC** | ❌ No | ❌ No | ❌ No | N/A | N/A | ✅ Designed |
| **GetMemoryFacts RPC** | ❌ No | ❌ No | ❌ No | N/A | N/A | ✅ Designed |
| **Authentication (JWT)** | ❌ No | ❌ No | ❌ No | ⚠️ Critical | N/A | ⚠️ Missing |
| **Interceptors (Auth, Logging)** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Missing |
| **Health Checks** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Missing |
| **Error Handling** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Missing |
| **Streaming (bi-directional)** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Not planned |
| **Connection Pooling** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Missing |
| **Retry Policy** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Missing |
| **Deadline/Timeout** | ❌ No | ❌ No | ❌ No | N/A | N/A | ⚠️ Missing |

**INSTRUÇÕES**:
- Design documentado em `TODO_PYTHON.md` (Phase 3, linhas 1080-1615)
- **Effort: 1-2 semanas**
- **MUST HAVE**: Authentication (JWT interceptor), Error handling, Health checks
- **NICE TO HAVE**: Streaming, Connection pooling

---

### TABELA 28: MCP Server Design Assessment

| Tool Category | Tools Planned | Implemented | Documentation | Priority |
|--------------|---------------|-------------|---------------|----------|
| **BI Tools** | 7 | ❌ 0 | ✅ MCP_SERVER_COMPLETE.md | P0 |
| **Agent Analysis Tools** | 5 | ❌ 0 | ✅ MCP_SERVER_COMPLETE.md | P1 |
| **CRM Operations Tools** | 8 | ❌ 0 | ✅ MCP_SERVER_COMPLETE.md | P0 |
| **Memory Tools** | 5 | ❌ 0 | ✅ MCP_SERVER_COMPLETE.md | P0 |
| **Document Tools** | 5 | ❌ 0 | ✅ MCP_SERVER_COMPLETE.md | P2 |
| **HTTP Server** | 1 | ❌ 0 | ✅ Architecture | P0 |
| **SSE Streaming** | 1 | ❌ 0 | ✅ Architecture | P0 |
| **JWT Auth** | 1 | ❌ 0 | ✅ Architecture | P0 |
| **Tool Registry** | 1 | ❌ 0 | ✅ Architecture | P0 |
| **Tool Executor** | 1 | ❌ 0 | ✅ Architecture | P0 |

**Status**: 0% implementado (apenas 1175 linhas de documentação)
**Effort**: 3-4 semanas (conforme AI_REPORT.md, Phase 2)

---

### TABELA 29: Documentação do Google ADK (Validation Checklist)

**IMPORTANTE**: Pesquise ATIVAMENTE na documentação oficial para validar:

| Aspecto | Referência Oficial | Status no Projeto | Conformidade |
|---------|-------------------|-------------------|--------------|
| **Google ADK 0.5+ Features** | https://ai.google.dev/gemini-api/docs | 📋 Documented | ⚠️ Validar versão |
| **Function Calling** | https://ai.google.dev/gemini-api/docs/function-calling | 📋 Tool registry planned | ⚠️ Validar design |
| **Streaming Responses** | https://ai.google.dev/gemini-api/docs/streaming | ⚠️ Missing | 🔴 ADICIONAR |
| **Context Caching** | https://ai.google.dev/gemini-api/docs/caching | ⚠️ Missing | 🔴 CRÍTICO para custos |
| **Safety Settings** | https://ai.google.dev/gemini-api/docs/safety-settings | ⚠️ Missing | ⚠️ ADICIONAR |
| **System Instructions** | https://ai.google.dev/gemini-api/docs/system-instructions | ✅ Defined (SALES_SYSTEM_PROMPT) | ✅ OK |
| **Multi-turn Conversations** | https://ai.google.dev/gemini-api/docs/multi-turn | ✅ Designed | ✅ OK |
| **Token Counting** | https://ai.google.dev/gemini-api/docs/tokens | ⚠️ Missing | 🔴 ADICIONAR para custos |
| **Error Handling** | https://ai.google.dev/gemini-api/docs/error-handling | ⚠️ Incomplete | 🔴 CRÍTICO |
| **Best Practices** | https://ai.google.dev/gemini-api/docs/best-practices | ⚠️ Not reviewed | ⚠️ REVISAR |
| **Rate Limiting** | https://ai.google.dev/gemini-api/docs/rate-limits | ⚠️ Missing | ⚠️ ADICIONAR |
| **Prompt Design** | https://ai.google.dev/gemini-api/docs/prompting | ✅ Defined | ✅ OK |

**ACTION**: Antes de implementar Python ADK, **LER e VALIDAR** TODAS as referências acima.

---

### TABELA 30: Análise de Integridade e Consistência de Dados

| Operação Crítica | Entidades | Padrão | Garantias | Riscos | Transação Atômica? |
|------------------|-----------|--------|-----------|--------|-------------------|
| Criar Contato + Publicar Evento | Contact, OutboxEvent | Outbox Pattern | At-least-once | ✅ Nenhum | ✅ Sim (LISTEN/NOTIFY) |
| Processar Webhook + Salvar Mensagem | Message, Contact, Session | Saga (Temporal) | Compensação | ⚠️ Idempotência | ✅ Sim |
| Atualizar Pipeline + Notificar | Contact, Pipeline | Event-Driven | Eventual consistency | ⚠️ Notificação pode falhar | ✅ Sim (DB) |
| Gerar Embedding + Armazenar | ❌ MemoryEmbedding | ❌ N/A | ❌ | 🔴 Não implementado | N/A |
| Python Agent → gRPC → DB | Contact, Message | ❌ N/A | ❌ | 🔴 Não implementado | N/A |
| Extrair Facts via LLM | ❌ MemoryFact | ❌ N/A | ❌ | 🔴 Não implementado | N/A |
| Hybrid Search (vector+keyword+graph) | ❌ MemoryEmbedding | ❌ N/A | ❌ | 🔴 Não implementado | N/A |

---

## SEÇÃO DE DESCOBERTAS E RECOMENDAÇÕES

### 3.1 Análise Crítica do Modelo de Dados

#### 3.1.1 Modelo de Domínio (DDD) - Go
- **Aggregates identificados e sua qualidade** (23 aggregates documentados em docs/domain_mapping/)
- **Value Objects existentes vs necessários** (Primitive Obsession?)
- **Invariantes de domínio: onde estão protegidos?** (constructors? métodos?)
- **Anemic Domain Model vs Rich Domain Model** (behaviors vs getters/setters)
- **Separação de concerns: Domain vs Persistence** (GORM tags no domain?)
- **Novo Aggregate: Chat** (planejado mas não implementado - CRÍTICO)

#### 3.1.2 Modelo de Persistência (Database) - Go
- **Normalização: formas normais atingidas** (3NF mínimo?)
- **Índices: cobertura e otimização** (300+ indexes conforme AI_REPORT.md)
- **Constraints: integridade referencial** (FK, UNIQUE, CHECK)
- **Tipos de dados: escolhas apropriadas?** (UUID vs INT, JSONB usage)
- **Migrations: estratégia e reversibilidade** (49 migrations, .down.sql implementados?)
- **AI Schema**: memory_embeddings, memory_facts (NÃO EXISTEM - gap P0)

#### 3.1.3 Impedance Mismatch (OO vs Relacional) - Go
- **Qualidade dos Mappers** (26 GORM repositories - clean?)
- **Problemas de N+1 queries** (Preload usage?)
- **Lazy loading vs Eager loading** (default behavior?)
- **Projeções e DTOs** (over-fetching?)

#### 3.1.4 Integridade e Consistência - Go
- **Transactional boundaries corretos?** (Aggregate = Transaction?)
- **Eventual consistency gerenciada?** (Outbox Pattern: ✅ LISTEN/NOTIFY, <100ms)
- **Optimistic Locking?** (Implementado em 8 agregados - quais?)
- **Idempotência em operações** (message processing idempotente?)

#### 3.1.5 AI/ML Data Architecture - Go
- **Vector Embeddings**: pgvector setup? (❌ Tabela memory_embeddings NÃO EXISTE)
- **Hybrid Search Strategy**: vector + keyword + graph + SQL? (❌ NÃO)
- **Memory Facts Schema**: Extração via LLM? (❌ Tabela memory_facts NÃO EXISTE)
- **Knowledge Graph**: Apache AGE? (❌ Não instalado)
- **Deduplication**: SHA256 content hashing? (❌ NÃO)

#### 3.1.6 API Design & Security - Go
- **RESTful compliance**: Resource-oriented? (verificar endpoints action-based)
- **OWASP Top 10 API Security**: Mitigações implementadas? (BOLA, BFLA, Rate Limiting)
- **Authentication**: JWT validation robust? (middleware/jwt_auth.go)
- **Authorization**: RBAC implementado? (middleware/rbac.go)
- **Error Handling**: Formato padronizado? Stack traces expostos? (errors/api_error.go)

#### 3.1.7 Python ADK Architecture (Planejado)
- **DDD/Clean Arch Compliance**: **CRÍTICO** - Python DEVE seguir mesmos padrões do Go
- **Domain Layer**: Agents como Aggregates? Invariantes?
- **Application Layer**: Use Cases (ProcessMessageUseCase?)
- **Infrastructure Layer**: gRPC, RabbitMQ, Temporal adapters
- **Separation of Concerns**: Domain puro sem dependências infra
- **Testing**: 70%+ coverage planejado?
- **Error Handling**: Circuit breakers, retries, fallbacks planejados?
- **Cost Management**: Context caching, token counting planejados?

### 3.2 Pontos Fortes

Liste aspectos arquiteturais e de modelagem bem implementados com **evidências do código real**.

**Exemplos esperados** (verificar no código):
- ✅ **Outbox Pattern**: LISTEN/NOTIFY, <100ms latency (infrastructure/messaging/postgres_notify_outbox.go)
- ✅ **Optimistic Locking**: 8 aggregates (quais? verificar migrations)
- ✅ **Test Coverage**: 82% domain layer (go test -cover)
- ✅ **104 Domain Events**: Well-named, first-class citizens (internal/domain/*/events.go)
- ✅ **Event-Driven**: RabbitMQ 15+ queues, DLQ, retry (infrastructure/messaging/)
- ✅ **Message Enrichment**: 12 providers funcionais (infrastructure/ai/)
- ✅ **Temporal Workflows**: 7 workflows, compensação (internal/workflows/)
- ✅ **DDD Documentation**: 23 aggregates, 15.000+ linhas (docs/domain_mapping/)

### 3.3 Gaps Críticos (Prioridade P0)

Liste problemas que podem causar:
- **Perda de dados**
- **Inconsistência de estado**
- **Corrupção de dados**
- **Violação de invariantes**
- **Falhas em cascata**
- **Indisponibilidade**
- **Custos LLM descontrolados**
- **Performance degradada**
- **Vulnerabilidades de segurança**

Para cada gap:
- **Descrição**: O que está faltando/errado?
- **Evidência**: Arquivo/linha, tabela DB, migration faltando
- **Impacto**: Qual o risco?
- **Exemplo de Falha**: Cenário concreto
- **Ação Corretiva**: Como resolver?
- **Esforço**: Semanas

**Exemplos esperados**:

1. **Memory Embeddings Table Missing** (P0 - 1 semana)
   - **Evidência**: Tabela não existe, buscar em migrations 000001-000049
   - **Impacto**: AI agents sem contexto semântico, Python ADK não funciona
   - **Ação**: Migration 000050 + MemoryEmbeddingRepository + Vector indexes (IVFFlat)

2. **Chat Entity Not Implemented** (P0 - 1 semana)
   - **Evidência**: internal/domain/chat/ não existe, messages sem chat_id FK
   - **Impacto**: Mensagens sem contexto de chat (WhatsApp groups, Telegram channels)
   - **Ação**: Chat aggregate + migration 000053 + message.chat_id FK

3. **gRPC API Not Implemented** (P0 - 2 semanas)
   - **Evidência**: api/proto/ vazio, nenhum .proto file
   - **Impacto**: Python ADK não pode acessar Memory Service
   - **Ação**: Proto definitions + Go server + Python client + auth interceptor

4. **API Security - BOLA Not Verified** (P0 - 1 semana)
   - **Evidência**: Verificar ownership checks em handlers (exemplo: GET /contacts/:id)
   - **Impacto**: Qualquer user pode acessar contacts de outros tenants
   - **Ação**: Adicionar checks `if resource.ProjectID != user.ProjectID { return 403 }`

5. **Error Handling in LLM Calls** (P0 - 1 semana)
   - **Evidência**: Verificar infrastructure/ai/*_provider.go (circuit breaker, retry?)
   - **Impacto**: Custos descontrolados, falhas silenciosas, timeout sem fallback
   - **Ação**: Circuit breaker + exponential backoff + fallback responses

6. **Testing Coverage for AI Features** (P0 - 2 semanas)
   - **Evidência**: Buscar *_test.go em infrastructure/ai/ (existem?)
   - **Impacto**: AI code não testado, bugs em produção, custos inesperados
   - **Ação**: Unit + Integration tests (mock LLM responses)

7. **Python ADK - DDD/Clean Arch Compliance** (P0 - 0 semanas - design review)
   - **Evidência**: Analisar docs/PYTHON_ADK_ARCHITECTURE*.md
   - **Impacto**: Python ADK pode violar separation of concerns, dificultar manutenção
   - **Ação**: Refatorar design para incluir Domain/Application/Infrastructure layers

8. **Rate Limiting Not Configured** (P0 - 3 dias)
   - **Evidência**: Verificar infrastructure/http/middleware/rate_limit.go (configurado?)
   - **Impacto**: DDoS via webhook floods, API abuse, resource exhaustion
   - **Ação**: Implementar rate limiting (Token Bucket) por IP + por user

9. **Context Caching (LLM Cost Management)** (P0 - 1 semana)
   - **Evidência**: Verificar se Google ADK context caching está planejado
   - **Impacto**: Custos LLM altos (repetir contexto em cada request)
   - **Ação**: Implementar context caching conforme Google ADK docs

10. **Security Headers Missing** (P0 - 2 dias)
    - **Evidência**: Verificar main.go ou middleware (CORS, X-Frame-Options, CSP)
    - **Impacto**: XSS, clickjacking, MIME sniffing attacks
    - **Ação**: Adicionar security headers middleware

### 3.4 Melhorias Importantes (Prioridade P1)

Liste problemas de qualidade, manutenibilidade e performance.

### 3.5 Otimizações (Prioridade P2)

Liste melhorias incrementais de arquitetura e modelo de dados.

### 3.6 Python ADK Roadmap Validation

- **Design Review**: Arquitetura multi-agent está correta?
- **Google ADK 0.5+ Compatibility**: Features usadas disponíveis?
- **DDD/Clean Arch**: Domain layer separado de Infrastructure?
- **Testing Strategy**: 70%+ coverage viável?
- **Error Handling**: Circuit breakers e retries planejados?
- **Observability**: Phoenix integration adequada?
- **Cost Management**: Context caching + token counting implementados?

---

## AVALIAÇÃO DE SAÚDE GERAL

### Score Geral por Dimensão (0-10)

**Go Backend**:
- **Arquitetura de Domínio (DDD)**: ___ /10
- **Modelagem de Entidades**: ___ /10
- **Design de Banco de Dados**: ___ /10
- **Mapeamento Domínio ↔ Persistência**: ___ /10
- **Separação de Concerns (Clean Arch)**: ___ /10
- **Integridade e Consistência de Dados**: ___ /10
- **Event-Driven Maturity**: ___ /10
- **Resiliência e Tolerância a Falhas**: ___ /10
- **Performance (Queries e Transações)**: ___ /10
- **Observability**: ___ /10
- **Testing Coverage**: ___ /10
- **API Design (RESTful)**: ___ /10
- **API Security (OWASP)**: ___ /10

**AI/ML Features (Go)**:
- **Message Enrichment**: ___ /10 (esperado: 8-9)
- **Memory Service**: ___ /10 (esperado: 0-2)
- **gRPC API**: ___ /10 (esperado: 0)
- **MCP Server**: ___ /10 (esperado: 0)

**Python ADK (Planejado)**:
- **Design Quality (based on docs)**: ___ /10
- **DDD/Clean Arch Compliance**: ___ /10
- **Implementation**: ___ /10 (esperado: 0)

**Overall**:
- **Cloud Readiness**: ___ /10
- **Production Readiness (Go)**: ___ /10
- **AI/ML Readiness (Go + Python)**: ___ /10

### Status de Saúde Final

🟢 **SAUDÁVEL (8-10)**: Arquitetura e modelo de dados sólidos, poucas melhorias necessárias
🟡 **ATENÇÃO (5-7)**: Funcional mas com gaps importantes de modelagem/consistência/segurança
🔴 **CRÍTICO (0-4)**: Necessita refatoração significativa de arquitetura e dados

**Veredicto**: [🟢/🟡/🔴] + Justificativa em 3-5 parágrafos cobrindo:

1. **Qualidade do modelo de domínio (Go)**: Aggregates, VOs, invariantes
2. **Qualidade do modelo de persistência (Go)**: Normalização, integridade, indexes
3. **Consistência e integridade (Go)**: Outbox, Optimistic Locking, transações
4. **AI/ML Readiness (Go + Python)**: Memory service, gRPC, Python ADK design
5. **API Security**: OWASP compliance, rate limiting, authentication
6. **Riscos principais**: Top 5 gaps críticos (P0)
7. **Esforço total estimado**: Semanas para atingir 8.5/10 em todas dimensões

---

## ROADMAP DE MELHORIAS (6-12 meses)

### Sprint 1-2 (P0 - Crítico - AI/ML Foundation)
- [ ] **Memory Service** (4 semanas): Migration 000050-052, hybrid search, RRF fusion
- [ ] **Chat Entity** (1 semana): internal/domain/chat/ + migration 000053
- [ ] **Error Handling LLM** (1 semana): Circuit breakers, retry, fallback
- [ ] **API Security - BOLA** (3 dias): Ownership checks em todos handlers
- [ ] **Rate Limiting** (3 dias): Token Bucket por IP + user

### Sprint 3-4 (P0 - Crítico - Python Integration)
- [ ] **gRPC API** (2 semanas): Proto definitions, Go server, Python client, auth
- [ ] **MCP Server** (3 semanas): HTTP server, 30+ tools, SSE streaming
- [ ] **Security Headers** (2 dias): CORS, X-Frame-Options, CSP
- [ ] **Context Caching** (1 semana): LLM cost optimization

### Sprint 5-10 (P0 - Crítico - Python ADK)
- [ ] **Python ADK Setup** (1 semana): Poetry, structure, DDD/Clean Arch review
- [ ] **Multi-Agent System** (3 semanas): CoordinatorAgent + 5 specialists
- [ ] **Semantic Router** (1 semana): Routes + testing
- [ ] **RabbitMQ Integration** (1 semana): Consumer + Publisher + error handling

### Sprint 11-12 (P1 - Importante - Testing & Observability)
- [ ] **Testing Coverage Go** (2 semanas): 70%+ (application layer, handlers, AI)
- [ ] **Testing Coverage Python** (1 semana): 70%+ (agents, router, gRPC client)
- [ ] **Phoenix Observability** (1 semana): Tracing + Metrics
- [ ] **Security Testing** (1 semana): OWASP ZAP, penetration tests

### Sprint 13-18 (P1 - Advanced AI Features)
- [ ] **Knowledge Graph** (3 semanas): Apache AGE + graph queries
- [ ] **Memory Facts Extraction** (2 semanas): LLM-based NER
- [ ] **Agent Templates** (1 semana): System agents registry

### Sprint 19-24 (P2 - Optimization)
- [ ] **Performance Tuning** (2 semanas): Retrieval strategies, caching
- [ ] **A/B Testing Framework** (1 semana): Prompt variations
- [ ] **Load Testing** (1 semana): k6, locust benchmarks
- [ ] **Cost Optimization** (1 semana): Token counting, embedding deduplication

**Total Effort**: **22-30 semanas** (~5.5-7.5 meses) para atingir 8.5/10 em todas dimensões

---

## CHECKLIST DE VALIDAÇÃO

### Para cada Entidade de Domínio (Go):
- [ ] Tem identidade clara (Entity) ou igualdade por valor (VO)?
- [ ] Invariantes protegidos no construtor/métodos?
- [ ] Sem dependências de infraestrutura (GORM tags, etc.)?
- [ ] Métodos de domínio (comportamento) ou só getters/setters?
- [ ] Emite Domain Events quando necessário?
- [ ] Pertence a um Aggregate Root?
- [ ] Transactional boundary adequado?
- [ ] Optimistic Locking se necessário?
- [ ] Testes unitários (70%+)?

### Para cada Tabela de Banco:
- [ ] Chave primária definida e apropriada (UUID)?
- [ ] Foreign keys com ON DELETE/UPDATE corretos?
- [ ] Índices em colunas de busca/join?
- [ ] Check constraints para regras simples?
- [ ] Timestamps (created_at, updated_at)?
- [ ] Soft delete (deleted_at) se necessário?
- [ ] Normalizada (3NF mínimo) ou desnormalização justificada?
- [ ] Vector indexes (pgvector) se necessário?
- [ ] GIN indexes (JSONB, full-text) se necessário?

### Para cada Endpoint API:
- [ ] RESTful design (resource-oriented)?
- [ ] HTTP methods corretos (GET/POST/PUT/DELETE)?
- [ ] Status codes apropriados (2xx/4xx/5xx)?
- [ ] Authentication (JWT)?
- [ ] Authorization (RBAC ownership check)?
- [ ] Rate limiting configurado?
- [ ] Error response padronizado?
- [ ] Idempotência (GET/PUT/DELETE)?
- [ ] Paginação (cursor-based)?
- [ ] Documentado (Swagger)?

### Para cada AI/ML Component:
- [ ] Design documentado (architecture, data flow)?
- [ ] Error handling (circuit breaker, retry, fallback)?
- [ ] Observability (logs, metrics, traces)?
- [ ] Testing (unit, integration, E2E)?
- [ ] Cost management (context caching, token counting)?
- [ ] Performance benchmarks (latency, throughput)?
- [ ] Security (auth, input validation)?

### Para Python ADK:
- [ ] **DDD/Clean Arch compliance** (Domain/Application/Infrastructure)?
- [ ] Type hints everywhere (mypy strict)?
- [ ] Docstrings (Google style)?
- [ ] Error handling (circuit breaker, retry)?
- [ ] Testing (pytest, 70%+ coverage)?
- [ ] Dependencies pinned (Poetry)?
- [ ] gRPC client tested (mocked + real)?
- [ ] RabbitMQ consumer tested (testcontainers)?
- [ ] LLM calls mocked in tests?
- [ ] Phoenix tracing configured?
- [ ] Cost management (context caching, token counting)?

---

## REFERÊNCIAS UTILIZADAS NA AVALIAÇÃO

### DDD & Architecture
- **Eric Evans** - Domain-Driven Design (2003)
- **Vaughn Vernon** - Implementing Domain-Driven Design (2013)
- **Robert Martin** - Clean Architecture (2017)
- **Martin Fowler** - Patterns of Enterprise Application Architecture (2002)

### Distributed Systems
- **Microsoft Azure Architecture Patterns** (CQRS, Saga, Event Sourcing)
- **Microservices.io** - Patterns (Saga, Outbox, Database per Service)
- **12 Factor App** (https://12factor.net)
- **Temporal Best Practices** (https://learn.temporal.io/)

### Database Design
- **Database Design Principles** - Normal Forms (Codd, Date)
- **PostgreSQL Documentation** - Indexes, Constraints, RLS
- **pgvector Documentation** (https://github.com/pgvector/pgvector)
- **Apache AGE Documentation** (https://age.apache.org/)

### AI/ML
- **Reciprocal Rank Fusion** (https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)
- **Hybrid Search Patterns** (https://www.pinecone.io/learn/hybrid-search/)
- **Google Generative AI Python SDK** (https://ai.google.dev/gemini-api/docs)
- **Semantic Router** (https://github.com/aurelio-labs/semantic-router)
- **Phoenix Observability** (https://docs.arize.com/phoenix/)

### API Design & Security
- **Roy Fielding** - REST Dissertation (2000)
- **OWASP Top 10 API Security 2023** (https://owasp.org/API-Security/editions/2023/en/0x11-t10/)
- **OpenAPI Specification** (https://swagger.io/specification/)
- **RESTful API Design Best Practices** (https://restfulapi.net/)

### Protocols & Standards
- **MCP Specification** (https://modelcontextprotocol.io/)
- **gRPC Best Practices** (https://grpc.io/docs/guides/best-practices/)
- **Protocol Buffers Guide** (https://protobuf.dev/)

---

## INSTRUÇÕES FINAIS PARA A IA

### Processo de Análise

1. **LEIA 100% DO CÓDIGO GO** antes de preencher qualquer tabela:
   - `internal/domain/` (94 arquivos)
   - `internal/application/` (todos use cases, commands, queries)
   - `infrastructure/` (136 arquivos)
   - `cmd/` (entrypoints)
   - TODAS as 49 migrations (.up.sql e .down.sql)

2. **Pesquise ativamente na documentação oficial**:
   - Google ADK 0.5+ (https://ai.google.dev/)
   - Semantic Router (https://github.com/aurelio-labs/semantic-router)
   - Phoenix Observability (https://docs.arize.com/phoenix/)
   - MCP Protocol (https://modelcontextprotocol.io/)
   - OWASP API Security (https://owasp.org/API-Security/)

3. **Leia TODA a documentação interna**:
   - AI_REPORT.md (580 linhas)
   - TODO.md (1117 linhas)
   - TODO_PYTHON.md (2797 linhas)
   - docs/domain_mapping/ (15.000+ linhas)
   - docs/MCP_SERVER_COMPLETE.md (1175 linhas)
   - docs/PYTHON_ADK_ARCHITECTURE*.md (3000+ linhas)

4. **Mapeie TODAS as entidades, eventos, use cases, endpoints**
5. **Verifique TODOS os relacionamentos** (FK, constraints)
6. **NÃO assuma implementações** - verifique código real
7. **Cite arquivos e linhas específicas** como evidência
8. **Compare com best practices** (DDD, Clean Arch, OWASP, RESTful)
9. **Seja crítico mas construtivo** - problemas + soluções
10. **Notas OBJETIVAS** (0-10 comparativo a referências)
11. **EXCLUA frontend** - apenas backend (Go + Python futuro)
12. **Priorize**: Corretude dados → Security → Error handling → Testing → Features
13. **Estime esforço em semanas** para cada gap
14. **Valide Python ADK design** contra DDD/Clean Arch (CRÍTICO)

### Formato de Saída Esperado

1. **TODAS as 30 tabelas preenchidas COMPLETAMENTE**
2. **Seção 3.1-3.6** (análise crítica) com 5+ páginas
3. **Gaps críticos com evidências** (arquivo:linha, migration, tabela DB)
4. **Roadmap priorizado** (6-12 meses, semanas)
5. **Score final com justificativa** (3-5 parágrafos)
6. **Exemplos de código ruim e como corrigir**
7. **Análise comparativa**: Go (atual) vs Python (planejado)
8. **Validação Python ADK** contra Google ADK docs
9. **Testing roadmap completo** (Go + Python, todas camadas)
10. **Security assessment** (OWASP Top 10 API)
11. **Cost management strategy** (LLM, embeddings, caching)

### Análise Holística

Ao final, responda:

1. **Alinhamento**: Go backend vs Python ADK vs AI features bem integrados?
2. **Consistência**: DDD, CQRS, Event-Driven aplicados igualmente?
3. **Conflitos**: Go vs Python, sync vs async, monolito vs microservices?
4. **Ordem de implementação**: Memory → gRPC → MCP → Python ADK correta?
5. **Esforços realistas**: 22-30 semanas viável?
6. **Dependências**: Circulares? Bloqueadores?
7. **Testing priorizado**: 70%+ em TODAS camadas?
8. **Error handling robusto**: Circuit breakers em TODOS componentes críticos?
9. **Observability adequada**: Logs, metrics, traces Go + Python?
10. **Cost management planejado**: Context caching, token counting, deduplication?
11. **Security robusta**: OWASP mitigado? Rate limiting? RBAC?
12. **API design RESTful**: Resource-oriented? Status codes corretos?

### Outputs Finais

Salve em **AI_REPORT.md** (sobrescrever):
- Avaliação arquitetural COMPLETA (30 tabelas)
- Go Backend (9.0/10) + análise detalhada
- API Security (OWASP compliance)
- AI/ML Features (2.5/10) + gaps críticos
- Python ADK (0.0/10) + design validation
- Roadmap 6-12 meses
- Scores finais + justificativas (3-5 parágrafos)

---

**Fim do Prompt**

**Versão**: 3.0 COMPLETA
**Data**: 2025-10-13
**Autores**: Ventros CRM Team
**Escopo**: Backend Go + Python ADK + AI/ML + API Security + Testing + Data Modeling
**Total de Tabelas**: 30 (18 originais + 8 AI/ML/Testing + 4 API Security)
**Objetivo**: Avaliação 360° com roadmap acionável, scores objetivos, security assessment
**Diferencial**: INSTRUÇÃO CRÍTICA - Ler 100% do código Go + Python DEVE seguir DDD/Clean Arch
