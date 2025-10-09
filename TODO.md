# ✅ TODO - Melhorias Arquiteturais Ventros CRM

> **Status**: 🟡 Em Desenvolvimento Ativo
> **Última atualização**: 2025-10-08
> **Formato**: Tarefas micro-segmentadas para execução incremental

---

## 📊 AVALIAÇÃO ARQUITETURAL - NOTAS TÉCNICAS

### Análise Completa DDD, SOLID, Saga Pattern e Event-Driven Architecture

#### **1. Domain-Driven Design (DDD) - Nota: 8.5/10**

**Pontos Fortes:**
- ✅ **Separação de camadas exemplar**: `internal/domain`, `internal/application`, `infrastructure` perfeitamente isolados
- ✅ **Agregados bem modelados**: Contact, Session, Message com boundaries claros e invariantes protegidos
- ✅ **Value Objects implementados**: Email e Phone com validação encapsulada e imutabilidade
- ✅ **Factory Methods consistentes**: `NewContact()`, `NewSession()`, `NewMessage()` com validações
- ✅ **Reconstituição de agregados**: Métodos `Reconstruct*()` separados da criação de negócio
- ✅ **Domain Events**: Eventos gerados pelos agregados (`ContactCreatedEvent`, `SessionStartedEvent`)
- ✅ **Encapsulamento rigoroso**: Campos privados com getters públicos, sem setters diretos
- ✅ **Ubiquitous Language**: Nomenclatura de negócio clara (Pipeline, Session, Contact, Channel)

**Pontos de Melhoria:**
- ⚠️ **Falta de Event Sourcing**: Eventos são publicados mas não há event store completo
- ⚠️ **Specifications Pattern ausente**: Queries complexas ainda nas repositories sem abstrações de domínio
- ⚠️ **Alguns eventos sem EventID**: Falta rastreabilidade única para idempotência (previsto no TODO)

**Arquivos Analisados:**
- `internal/domain/contact/contact.go:1-318` - Agregado Contact
- `internal/domain/session/session.go:1-464` - Agregado Session
- `internal/domain/message/message.go:1-258` - Agregado Message
- `internal/domain/contact/value_objects.go:1-73` - Value Objects Email e Phone

---

#### **2. Princípios SOLID - Nota: 8.0/10**

**S - Single Responsibility Principle: 9/10**
- ✅ Cada agregado tem responsabilidade única e coesa
- ✅ Use cases separados por funcionalidade (`ProcessInboundMessageUseCase`, `CreateContactUseCase`)
- ✅ Repositories com responsabilidade única de persistência
- ⚠️ `ProcessInboundMessageUseCase` tem múltiplas responsabilidades (criar contact, session, message, tracking) - poderia ser decomposto

**O - Open/Closed Principle: 8/10**
- ✅ Interfaces para repositories permitem extensão sem modificação
- ✅ Event-driven architecture permite adicionar novos consumers sem alterar publishers
- ✅ Strategy pattern em `MessageSenderFactory` para diferentes canais
- ⚠️ Alguns switches em `mapDomainToBusinessEvents` poderiam ser registry-based

**L - Liskov Substitution Principle: 9/10**
- ✅ Interfaces `Repository`, `EventBus`, `Consumer` são bem definidas
- ✅ Implementações substituíveis sem quebrar contratos
- ✅ Domain events implementam interface `DomainEvent` corretamente

**I - Interface Segregation Principle: 7/10**
- ✅ Interfaces focadas: `EventBus`, `Repository`, `Consumer`
- ⚠️ Repository do Contact tem muitos métodos (`FindByPhone`, `FindByEmail`, `FindByProject`, etc) - poderia ser segregado
- ⚠️ `EventBus` mistura publish single e batch - poderia ter interfaces separadas

**D - Dependency Inversion Principle: 9/10**
- ✅ Camada de domínio não depende de infraestrutura
- ✅ Use cases dependem de interfaces, não de implementações concretas
- ✅ Dependency Injection via construtores
- ✅ Temporal workflows usam interfaces `EventBus` ao invés de implementação concreta

**Arquivos Analisados:**
- `internal/application/message/process_inbound_message.go:1-448`
- `infrastructure/persistence/gorm_contact_repository.go`
- `infrastructure/messaging/domain_event_bus.go:1-265`

---

#### **3. Saga Pattern - Nota: 7.5/10**

**3.1 Orquestração com Temporal: 8/10**

**Pontos Fortes:**
- ✅ **Workflows bem estruturados**: `SessionLifecycleWorkflow` gerencia ciclo de vida completo
- ✅ **Activities isoladas**: `EndSessionActivity`, `CleanupSessionsActivity` com responsabilidades claras
- ✅ **Timeout management**: Timer + Signals para reset dinâmico de timeout
- ✅ **Graceful degradation**: Workflow continua mesmo se activity falha
- ✅ **Idempotência nas activities**: Verifica se sessão já foi encerrada antes de processar

**Pontos de Melhoria:**
- ⚠️ **Falta compensação explícita**: Não há compensation activities implementadas (previsto no TODO P1)
- ⚠️ **Saga state não persistido**: Falta tabela `saga_state` para tracking (previsto no TODO)
- ⚠️ **Retry policies básicos**: ActivityOptions com timeout fixo, sem exponential backoff configurado

**Arquivos Analisados:**
- `internal/workflows/session/session_lifecycle_workflow.go:1-133`
- `internal/workflows/session/session_activities.go:1-264`

**3.2 Coreografia com RabbitMQ: 7/10**

**Pontos Fortes:**
- ✅ **Event-driven choreography**: Eventos de domínio disparam consumers independentes
- ✅ **Dead Letter Queues (DLQ)**: Todas as queues têm DLQ configurada
- ✅ **Retry logic**: Máximo 3 tentativas antes de ir para DLQ
- ✅ **Quorum queues**: Alta disponibilidade com `x-queue-type: quorum`
- ✅ **Consumer decoupling**: ContactEventConsumer processa eventos sem conhecer publishers

**Pontos de Melhoria:**
- ❌ **Falta Outbox Pattern**: Publicação de eventos não é transacional (previsto no TODO P0 - crítico)
- ❌ **Falta idempotência**: Eventos podem ser processados duplicados (previsto no TODO P0)
- ⚠️ **Correlation ID ausente**: Falta propagação de ID de correlação entre eventos (previsto no TODO P1)
- ⚠️ **Circuit breaker ausente**: Sem proteção contra falhas em cascata (previsto no TODO P1)

**Arquivos Analisados:**
- `infrastructure/messaging/contact_event_consumer.go:1-680`
- `infrastructure/messaging/rabbitmq.go:1-631`
- `infrastructure/messaging/domain_event_bus.go:1-265`

---

#### **4. Arquitetura de Eventos - Nota: 7.0/10**

**Pontos Fortes:**
- ✅ **Domain events bem nomeados**: `ContactCreatedEvent`, `SessionStartedEvent`, `MessageCreatedEvent`
- ✅ **Event mapping**: `mapDomainToBusinessEvents()` converte eventos internos para webhooks
- ✅ **Event log repository**: Logs de eventos para auditoria
- ✅ **Webhook integration**: Eventos disparam webhooks externos automaticamente
- ✅ **Batch publishing**: `PublishBatch()` para publicar múltiplos eventos

**Pontos de Melhoria:**
- ❌ **Eventos sem ID único**: Falta `EventID` para rastreabilidade e idempotência
- ❌ **Eventos sem versão**: Falta `EventVersion` para schema evolution
- ⚠️ **Eventos sem timestamp consistente**: Alguns eventos usam `time.Now()` na publicação, não na geração
- ⚠️ **TenantID missing**: `SessionEndedEvent` e `ContactUpdatedEvent` sem tenantID

---

#### **5. Testes - Nota: 6.0/10**

**Testes Implementados (10 arquivos):**
- ✅ `internal/domain/contact/contact_test.go` - Testes de Contact aggregate
- ✅ `internal/domain/contact/email_test.go` - Value Object Email
- ✅ `internal/domain/contact/phone_test.go` - Value Object Phone
- ✅ `internal/domain/contact/ad_conversion_event_test.go` - Evento de conversão
- ✅ `internal/domain/session/session_test.go` - Testes de Session aggregate
- ✅ `internal/domain/message/message_test.go` - Testes de Message aggregate
- ✅ `infrastructure/persistence/gorm_contact_repository_test.go` - Repository integration test
- ✅ `infrastructure/channels/waha/message_adapter_test.go` - Adapter test
- ✅ `tests/e2e/waha_webhook_test.go` - E2E webhook test
- ✅ `tests/e2e/api_test.go` - E2E API test

**Pontos de Melhoria:**
- ❌ **Coverage baixo**: Faltam testes para 90% dos use cases
- ❌ **Faltam testes de repositories**: Apenas ContactRepository tem testes
- ❌ **Faltam testes de integration**: Setup com testcontainers previsto mas não completo
- ❌ **Faltam testes de idempotência**: Crítico para event-driven systems
- ❌ **Faltam testes de RLS**: Security-critical, previsto no TODO P0

**Cobertura Estimada:** ~15-20% (muitas funcionalidades sem testes)

---

#### **6. Observabilidade - Nota: 4.0/10**

**Pontos Fortes:**
- ✅ Logging estruturado com Zap em alguns componentes
- ✅ Health checks implementados

**Pontos de Melhoria:**
- ❌ **Sem distributed tracing**: OpenTelemetry previsto no TODO P1 mas não implementado
- ❌ **Sem métricas Prometheus**: Métricas de negócio e técnicas previstas no TODO P2
- ❌ **Sem correlation ID**: Impossível rastrear requests através do sistema
- ❌ **Logs inconsistentes**: Alguns componentes usam `fmt.Printf`, outros Zap

---

#### **7. Resiliência - Nota: 6.5/10**

**Pontos Fortes:**
- ✅ **Auto-reconnect RabbitMQ**: Reconexão automática com retry
- ✅ **Dead Letter Queues**: Mensagens falhas vão para DLQ
- ✅ **Retry logic**: 3 tentativas antes de DLQ
- ✅ **Graceful shutdown**: Temporal workflows podem ser cancelados

**Pontos de Melhoria:**
- ❌ **Sem circuit breakers**: Falhas em serviços externos podem causar cascata
- ❌ **Sem rate limiting**: Endpoints desprotegidos
- ❌ **Sem timeouts configuráveis**: Timeouts hardcoded em activities
- ⚠️ **Connection pooling default**: Database pool não otimizado

---

### **NOTA GERAL DO PROJETO: 7.2/10**

**Média Ponderada:**
- DDD (peso 25%): 8.5 × 0.25 = 2.12
- SOLID (peso 20%): 8.0 × 0.20 = 1.60
- Saga Pattern (peso 20%): 7.5 × 0.20 = 1.50
- Event Architecture (peso 15%): 7.0 × 0.15 = 1.05
- Testes (peso 10%): 6.0 × 0.10 = 0.60
- Observabilidade (peso 5%): 4.0 × 0.05 = 0.20
- Resiliência (peso 5%): 6.5 × 0.05 = 0.33

**TOTAL: 7.4/10**

---

### **PRIORIZAÇÃO DE MELHORIAS**

**🔴 CRÍTICO (P0) - Implementar Imediatamente:**
1. **Outbox Pattern** (TODO item 3) - Garantir consistência transacional de eventos
2. **Idempotência** (TODO item 4) - Prevenir processamento duplicado
3. **Testes de domínio completos** (TODO item 1.2-1.4) - Garantir qualidade
4. **Event ID nos eventos** (TODO item 4.2) - Rastreabilidade

**🟡 ALTO (P1) - Próximos Sprints:**
5. **OpenTelemetry** (TODO item 5) - Distributed tracing
6. **Correlation ID** (TODO item 6) - Request tracking
7. **Circuit Breakers** (TODO item 7) - Proteção contra falhas
8. **Compensation Sagas** (TODO item 9) - Rollback transacional

**🟢 MÉDIO (P2) - Backlog:**
9. **Métricas Prometheus** (TODO item 15) - Observabilidade
10. **Event Versioning** (TODO item 10) - Schema evolution
11. **Contract Tests** (TODO item 14) - Event contracts

---

## 📝 JUSTIFICATIVAS TÉCNICAS DAS NOTAS

### Por que DDD 8.5/10?
A arquitetura segue fielmente os padrões DDD com separação clara de camadas, agregados bem modelados com invariantes protegidos, value objects imutáveis, factory methods e domain events. Os agregados Contact, Session e Message demonstram encapsulamento exemplar e lógica de negócio pura. Penalizado pela ausência de Event Sourcing completo e Specifications Pattern.

### Por que SOLID 8.0/10?
Código bem estruturado com Single Responsibility evidente, interfaces segregadas e Dependency Inversion consistente. Repositories e Use Cases seguem SRP. Open/Closed bem aplicado com event-driven architecture. Penalizado por algumas repositories com muitos métodos (Contact) e alguns use cases com múltiplas responsabilidades.

### Por que Saga Pattern 7.5/10?
Temporal workflows bem implementados com timeout management e activities isoladas. Coreografia via RabbitMQ funcional com DLQs e retries. Penalizado gravemente pela ausência de Outbox Pattern (inconsistência transacional) e falta de compensation sagas (impossível rollback distribuído).

### Por que Event Architecture 7.0/10?
Eventos de domínio bem nomeados e event-driven architecture funcional. Event mapping para webhooks bem pensado. Penalizado pela falta de EventID único (impossibilita idempotência), falta de versionamento (impossibilita schema evolution) e alguns eventos sem tenantID.

### Por que Testes 6.0/10?
Testes de domínio implementados para agregados principais (Contact, Session, Message) e value objects (Email, Phone). Integration test básico de repository. Penalizado severamente por coverage baixíssimo (~15%), falta de testes de use cases, falta de testes de idempotência e RLS.

### Por que Observabilidade 4.0/10?
Apenas logging básico com Zap e health checks. Nota baixa justificada pela ausência de distributed tracing (OpenTelemetry), métricas Prometheus, correlation ID e logs inconsistentes (mix de fmt.Printf e Zap).

### Por que Resiliência 6.5/10?
Boa implementação de retry com RabbitMQ e DLQs. Auto-reconnect funcional. Penalizado pela ausência de circuit breakers (crítico para serviços externos como WAHA), rate limiting e timeouts não configuráveis.

---

## 📋 LEGENDA

- [ ] **Pendente** - Não iniciado
- [⏳] **Em Progresso** - Sendo trabalhado
- [✅] **Concluído** - Finalizado
- [🔴] **Crítico** - Prioridade máxima (P0)
- [🟡] **Alto** - Prioridade alta (P1)
- [🟢] **Médio** - Prioridade média (P2)
- [⚪] **Baixo** - Prioridade baixa (P3)

**Estimativa de tempo**: 🕐 = 1-2h | 🕑 = 2-4h | 🕒 = 4-8h | 🕓 = 1-2 dias

---

## 🔴 P0 - CRÍTICO (Fazer Primeiro)

### 1. Testing Strategy - Testes Unitários de Domínio

#### 1.1 Setup de Testes 🕐
- [✅] 🔴 Instalar dependências de teste
  ```bash
  go get github.com/stretchr/testify/assert
  go get github.com/stretchr/testify/require
  ```
- [✅] 🔴 Criar helper de testes `internal/domain/test_helpers.go`
- [✅] 🔴 Configurar `go test` no Makefile com coverage

**Arquivos:**
- `go.mod` (atualizar)
- `internal/domain/test_helpers.go` (criar)
- `Makefile` (adicionar target `test-domain`)

---

#### 1.2 Testes de Contact Aggregate 🕑 - ✅ COMPLETO

##### 1.2.1 Testes de Factory Method - ✅ COMPLETO
- [✅] 🔴 Criar arquivo `internal/domain/contact/contact_test.go` - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewContact_Success` - criação válida - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewContact_EmptyProjectID` - erro quando projectID é nil - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewContact_EmptyTenantID` - erro quando tenantID vazio - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewContact_EmptyName` - erro quando name vazio - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewContact_GeneratesEvent` - verifica evento ContactCreatedEvent - IMPLEMENTADO

**Arquivo:** `internal/domain/contact/contact_test.go`

##### 1.2.2 Testes de Email Value Object - ✅ COMPLETO
- [✅] 🔴 Teste: `TestSetEmail_ValidEmail` - aceita email válido - IMPLEMENTADO (email_test.go)
- [✅] 🔴 Teste: `TestSetEmail_InvalidFormat` - rejeita email inválido - IMPLEMENTADO (email_test.go)
- [✅] 🔴 Teste: `TestSetEmail_UpdatesTimestamp` - verifica updatedAt - IMPLEMENTADO (email_test.go)

##### 1.2.3 Testes de Phone Value Object - ✅ COMPLETO
- [✅] 🔴 Teste: `TestSetPhone_ValidPhone` - aceita telefone válido - IMPLEMENTADO (phone_test.go)
- [✅] 🔴 Teste: `TestSetPhone_InvalidFormat` - rejeita telefone inválido - IMPLEMENTADO (phone_test.go)
- [✅] 🔴 Teste: `TestSetPhone_UpdatesTimestamp` - verifica updatedAt - IMPLEMENTADO (phone_test.go)

##### 1.2.4 Testes de Métodos de Negócio
- [✅] 🔴 Teste: `TestUpdateName_Success` - atualiza nome
- [✅] 🔴 Teste: `TestUpdateName_EmptyName` - rejeita nome vazio
- [✅] 🔴 Teste: `TestUpdateName_GeneratesEvent` - gera ContactUpdatedEvent
- [✅] 🔴 Teste: `TestAddTag_NewTag` - adiciona tag nova
- [✅] 🔴 Teste: `TestAddTag_DuplicateTag` - ignora tag duplicada
- [✅] 🔴 Teste: `TestRemoveTag_ExistingTag` - remove tag existente
- [✅] 🔴 Teste: `TestRemoveTag_NonExistingTag` - não falha se tag não existe
- [✅] 🔴 Teste: `TestRecordInteraction_FirstTime` - define firstInteractionAt
- [✅] 🔴 Teste: `TestRecordInteraction_UpdatesLastInteraction` - atualiza lastInteractionAt

##### 1.2.5 Testes de Soft Delete
- [✅] 🔴 Teste: `TestSoftDelete_Success` - marca como deletado
- [✅] 🔴 Teste: `TestSoftDelete_AlreadyDeleted` - erro se já deletado
- [✅] 🔴 Teste: `TestSoftDelete_GeneratesEvent` - gera ContactDeletedEvent
- [✅] 🔴 Teste: `TestIsDeleted_ReturnsTrueAfterDelete`

---

#### 1.3 Testes de Session Aggregate 🕑 - ✅ COMPLETO

##### 1.3.1 Testes de Factory Method - ✅ COMPLETO
- [✅] 🔴 Criar arquivo `internal/domain/session/session_test.go` - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewSession_Success` - criação válida - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewSession_EmptyContactID` - erro quando contactID nil - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewSession_EmptyTenantID` - erro quando tenantID vazio - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewSession_DefaultTimeout` - usa 30min se timeout <= 0 - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewSessionWithPipeline_Success` - criação com pipeline - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewSessionWithPipeline_NilPipelineID` - erro se pipeline nil - IMPLEMENTADO

**Arquivo:** `internal/domain/session/session_test.go`

##### 1.3.2 Testes de Lifecycle
- [✅] 🔴 Teste: `TestRecordMessage_FromContact` - incrementa messagesFromContact
- [✅] 🔴 Teste: `TestRecordMessage_FromAgent` - incrementa messagesFromAgent
- [✅] 🔴 Teste: `TestRecordMessage_NonActiveSession` - erro se sessão não ativa
- [✅] 🔴 Teste: `TestRecordMessage_UpdatesLastActivity` - atualiza lastActivityAt
- [✅] 🔴 Teste: `TestEnd_Success` - encerra sessão ativa
- [✅] 🔴 Teste: `TestEnd_AlreadyEnded` - erro se já encerrada
- [✅] 🔴 Teste: `TestEnd_CalculatesDuration` - calcula durationSeconds
- [✅] 🔴 Teste: `TestEnd_GeneratesEvent` - gera SessionEndedEvent

##### 1.3.3 Testes de Timeout
- [✅] 🔴 Teste: `TestCheckTimeout_NotExpired` - retorna false se não expirou
- [✅] 🔴 Teste: `TestCheckTimeout_Expired` - retorna true e encerra sessão
- [✅] 🔴 Teste: `TestCheckTimeout_NonActiveSession` - retorna false

##### 1.3.4 Testes de Métricas de Resposta
- [✅] 🔴 Teste: `TestRecordMessage_FirstContactMessage` - define firstContactMessageAt
- [✅] 🔴 Teste: `TestRecordMessage_FirstAgentResponse` - define firstAgentResponseAt
- [✅] 🔴 Teste: `TestRecordMessage_CalculatesAgentResponseTime` - calcula tempo de resposta
- [✅] 🔴 Teste: `TestRecordMessage_CalculatesContactWaitTime` - calcula tempo de espera

##### 1.3.5 Testes de Agentes
- [✅] 🔴 Teste: `TestAssignAgent_FirstAgent` - atribui primeiro agente
- [✅] 🔴 Teste: `TestAssignAgent_Transfer` - incrementa agentTransfers
- [✅] 🔴 Teste: `TestAssignAgent_DuplicateAgent` - não adiciona duplicado
- [✅] 🔴 Teste: `TestAssignAgent_NonActiveSession` - erro se não ativa

---

#### 1.4 Testes de Message Aggregate 🕐 - ✅ COMPLETO

##### 1.4.1 Testes de Factory Method - ✅ COMPLETO
- [✅] 🔴 Criar arquivo `internal/domain/message/message_test.go` - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewMessage_Success` - criação válida - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewMessage_EmptyContactID` - erro quando contactID nil - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewMessage_EmptyProjectID` - erro quando projectID nil - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewMessage_InvalidContentType` - erro quando contentType inválido - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewMessage_GeneratesEvent` - gera MessageCreatedEvent - IMPLEMENTADO

**Arquivo:** `internal/domain/message/message_test.go`

##### 1.4.2 Testes de Content Type
- [✅] 🔴 Teste: `TestSetText_ValidTextMessage` - define texto em mensagem text
- [✅] 🔴 Teste: `TestSetText_NonTextMessage` - erro se não for text
- [✅] 🔴 Teste: `TestSetMediaContent_ValidMediaMessage` - define URL e mimetype
- [✅] 🔴 Teste: `TestSetMediaContent_NonMediaMessage` - erro se não for media

##### 1.4.3 Testes de Status Transitions
- [✅] 🔴 Teste: `TestMarkAsDelivered_Success` - muda status para Delivered
- [✅] 🔴 Teste: `TestMarkAsDelivered_SetsTimestamp` - define deliveredAt
- [✅] 🔴 Teste: `TestMarkAsDelivered_GeneratesEvent` - gera MessageDeliveredEvent
- [✅] 🔴 Teste: `TestMarkAsRead_Success` - muda status para Read
- [✅] 🔴 Teste: `TestMarkAsRead_SetsTimestamp` - define readAt
- [✅] 🔴 Teste: `TestMarkAsFailed_Success` - muda status para Failed

---

#### 1.5 Testes de Value Objects 🕐 - ✅ COMPLETO

##### 1.5.1 Email Value Object - ✅ COMPLETO
- [✅] 🔴 Criar arquivo `internal/domain/contact/email_test.go` - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewEmail_ValidEmail` - aceita emails válidos - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewEmail_InvalidFormat` - rejeita formato inválido - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewEmail_EmptyString` - rejeita string vazia - IMPLEMENTADO
- [✅] 🔴 Teste: `TestEmail_String` - retorna string corretamente - IMPLEMENTADO

**Arquivo:** `internal/domain/contact/email_test.go` ✅

##### 1.5.2 Phone Value Object - ✅ COMPLETO
- [✅] 🔴 Criar arquivo `internal/domain/contact/phone_test.go` - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewPhone_ValidPhone` - aceita telefones válidos - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewPhone_InvalidFormat` - rejeita formato inválido - IMPLEMENTADO
- [✅] 🔴 Teste: `TestNewPhone_EmptyString` - rejeita string vazia - IMPLEMENTADO
- [✅] 🔴 Teste: `TestPhone_String` - retorna string corretamente - IMPLEMENTADO

**Arquivo:** `internal/domain/contact/phone_test.go` ✅

---

### 2. Testes de Integração - Repositórios

#### 2.1 Setup de Testes de Integração 🕑 - ✅ PARCIALMENTE COMPLETO
- [✅] 🔴 Instalar testcontainers - IMPLEMENTADO
  ```bash
  go get github.com/testcontainers/testcontainers-go
  go get github.com/testcontainers/testcontainers-go/modules/postgres
  ```
- [✅] 🔴 Criar helper `infrastructure/persistence/test_helpers.go` - IMPLEMENTADO
- [✅] 🔴 Criar função `SetupTestDatabase()` - inicia container PostgreSQL - IMPLEMENTADO
- [✅] 🔴 Criar função `TeardownTestDatabase()` - para container - IMPLEMENTADO
- [✅] 🔴 Criar função `SeedTestData()` - popula dados de teste - IMPLEMENTADO

**Arquivo:** `infrastructure/persistence/test_helpers.go` ✅

---

#### 2.2 Testes de GormContactRepository 🕑 - ✅ COMPLETO

##### 2.2.1 Setup - ✅ COMPLETO
- [✅] 🔴 Criar arquivo `infrastructure/persistence/gorm_contact_repository_test.go` - IMPLEMENTADO
- [✅] 🔴 Criar `TestMain()` para setup/teardown global - IMPLEMENTADO
- [✅] 🔴 Criar helper `createTestContact()` - cria contato de teste - IMPLEMENTADO

**Arquivo:** `infrastructure/persistence/gorm_contact_repository_test.go` ✅

##### 2.2.2 Testes de Save - ✅ COMPLETO
- [✅] 🔴 Teste: `TestGormContactRepository_Save_NewContact` - insere novo - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_Save_UpdateContact` - atualiza existente - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_Save_PreservesID` - não muda ID - IMPLEMENTADO

##### 2.2.3 Testes de FindByID - ✅ COMPLETO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByID_Exists` - encontra contato - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByID_NotFound` - retorna erro - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByID_ReconstructsDomain` - mapeia corretamente - IMPLEMENTADO

##### 2.2.4 Testes de FindByPhone - ✅ COMPLETO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByPhone_Exists` - encontra por telefone - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByPhone_NotFound` - retorna erro - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByPhone_IgnoresDeleted` - ignora soft deleted - IMPLEMENTADO

##### 2.2.5 Testes de FindByEmail - ✅ COMPLETO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByEmail_Exists` - encontra por email - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByEmail_NotFound` - retorna erro - IMPLEMENTADO

##### 2.2.6 Testes de Paginação - ✅ COMPLETO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByProject_WithLimit` - respeita limit - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_FindByProject_WithOffset` - respeita offset - IMPLEMENTADO
- [✅] 🔴 Teste: `TestGormContactRepository_CountByProject` - conta corretamente - IMPLEMENTADO

---

#### 2.3 Testes de RLS (Row Level Security) 🕐
- [ ] 🔴 Teste: `TestRLS_IsolatesTenants` - tenant A não vê dados de tenant B
- [ ] 🔴 Teste: `TestRLS_WithoutUserID` - falha se user_id não definido
- [ ] 🔴 Teste: `TestRLS_Callbacks` - verifica callbacks GORM funcionam

**Arquivo:** `infrastructure/persistence/rls_test.go` (criar)

---

### 3. Outbox Pattern (Transactional Outbox)

#### 3.1 Database Schema 🕐
- [ ] 🔴 Criar migration `migrations/20250108_create_outbox_events.sql`
- [ ] 🔴 Definir tabela `outbox_events`:
  ```sql
  CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,
    CONSTRAINT chk_status CHECK (status IN ('pending', 'processing', 'processed', 'failed'))
  );
  ```
- [ ] 🔴 Criar índices:
  ```sql
  CREATE INDEX idx_outbox_status ON outbox_events(status);
  CREATE INDEX idx_outbox_created_at ON outbox_events(created_at);
  ```

**Arquivo:** `migrations/20250108_create_outbox_events.sql`

---

#### 3.2 Domain Interface 🕐
- [ ] 🔴 Criar interface `OutboxRepository` em `internal/domain/shared/outbox_repository.go`
- [ ] 🔴 Definir método `Save(ctx, aggregateID, aggregateType, eventType, eventData)`
- [ ] 🔴 Definir método `GetPendingEvents(ctx, limit) ([]OutboxEvent, error)`
- [ ] 🔴 Definir método `MarkAsProcessed(ctx, eventID) error`
- [ ] 🔴 Definir método `MarkAsFailed(ctx, eventID, error) error`

**Arquivo:** `internal/domain/shared/outbox_repository.go`

---

#### 3.3 Infrastructure Implementation 🕑
- [ ] 🔴 Criar `infrastructure/persistence/entities/outbox_event.go`
- [ ] 🔴 Definir struct `OutboxEventEntity` com tags GORM
- [ ] 🔴 Criar `infrastructure/persistence/gorm_outbox_repository.go`
- [ ] 🔴 Implementar `Save()` - insere evento no outbox
- [ ] 🔴 Implementar `GetPendingEvents()` - busca eventos pending
- [ ] 🔴 Implementar `MarkAsProcessed()` - atualiza status para processed
- [ ] 🔴 Implementar `MarkAsFailed()` - atualiza status e incrementa retry_count

**Arquivos:**
- `infrastructure/persistence/entities/outbox_event.go`
- `infrastructure/persistence/gorm_outbox_repository.go`

---

#### 3.4 Outbox Processor Worker 🕑
- [ ] 🔴 Criar `infrastructure/messaging/outbox_processor.go`
- [ ] 🔴 Implementar `OutboxProcessor` struct com dependencies
- [ ] 🔴 Implementar `Start()` - inicia worker em goroutine
- [ ] 🔴 Implementar `processEvents()` - loop principal
  - [ ] Buscar eventos pending (batch de 10)
  - [ ] Publicar cada evento no RabbitMQ
  - [ ] Marcar como processed ou failed
  - [ ] Sleep 1 segundo entre batches
- [ ] 🔴 Implementar `Stop()` - graceful shutdown
- [ ] 🔴 Adicionar retry logic (max 3 tentativas)
- [ ] 🔴 Adicionar logging estruturado

**Arquivo:** `infrastructure/messaging/outbox_processor.go`

---

#### 3.5 Integração com Use Cases 🕑
- [ ] 🔴 Modificar `CreateContactUseCase` para usar outbox
  - [ ] Injetar `OutboxRepository` no construtor
  - [ ] Salvar eventos no outbox em vez de publicar diretamente
  - [ ] Usar transação GORM para atomicidade
- [ ] 🔴 Modificar `ProcessInboundMessageUseCase` para usar outbox
- [ ] 🔴 Criar helper `SaveEventsToOutbox()` para reutilizar lógica

**Arquivos a modificar:**
- `internal/application/contact/create_contact.go`
- `internal/application/message/process_inbound_message_usecase.go`

---

#### 3.6 Startup Integration 🕐
- [ ] 🔴 Modificar `cmd/api/main.go`
- [ ] 🔴 Instanciar `OutboxRepository`
- [ ] 🔴 Instanciar `OutboxProcessor`
- [ ] 🔴 Iniciar `OutboxProcessor.Start()` em goroutine
- [ ] 🔴 Adicionar graceful shutdown do processor

**Arquivo:** `cmd/api/main.go`

---

### 4. Idempotência em Event Handlers

#### 4.1 Database Schema 🕐
- [ ] 🔴 Criar migration `migrations/20250108_create_processed_events.sql`
- [ ] 🔴 Definir tabela `processed_events`:
  ```sql
  CREATE TABLE processed_events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    handler_name VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(event_id, handler_name)
  );
  ```
- [ ] 🔴 Criar índice:
  ```sql
  CREATE INDEX idx_processed_events_lookup ON processed_events(event_id, handler_name);
  ```

**Arquivo:** `migrations/20250108_create_processed_events.sql`

---

#### 4.2 Adicionar EventID aos Domain Events 🕐
- [ ] 🔴 Modificar `internal/domain/shared/domain_event.go`
- [ ] 🔴 Adicionar método `EventID() uuid.UUID` na interface
- [ ] 🔴 Modificar `ContactCreatedEvent` para incluir `eventID uuid.UUID`
- [ ] 🔴 Modificar `ContactUpdatedEvent` para incluir `eventID uuid.UUID`
- [ ] 🔴 Modificar `SessionStartedEvent` para incluir `eventID uuid.UUID`
- [ ] 🔴 Modificar `SessionEndedEvent` para incluir `eventID uuid.UUID`
- [ ] 🔴 Modificar `MessageCreatedEvent` para incluir `eventID uuid.UUID`
- [ ] 🔴 Gerar UUID automaticamente nos construtores de eventos

**Arquivos a modificar:**
- `internal/domain/shared/domain_event.go`
- `internal/domain/contact/events.go`
- `internal/domain/session/events.go`
- `internal/domain/message/events.go`

---

#### 4.3 Idempotency Checker Service 🕑
- [ ] 🔴 Criar `infrastructure/persistence/idempotency_checker.go`
- [ ] 🔴 Criar struct `IdempotencyChecker` com DB dependency
- [ ] 🔴 Implementar `IsProcessed(ctx, eventID, handlerName) (bool, error)`
  - [ ] Query na tabela `processed_events`
  - [ ] Retorna true se já processado
- [ ] 🔴 Implementar `MarkAsProcessed(ctx, eventID, handlerName) error`
  - [ ] Insert na tabela `processed_events`
  - [ ] Usar `ON CONFLICT DO NOTHING` para evitar duplicatas
- [ ] 🔴 Adicionar logging estruturado

**Arquivo:** `infrastructure/persistence/idempotency_checker.go`

---

#### 4.4 Integração com Consumers 🕑

##### 4.4.1 ContactEventConsumer
- [ ] 🔴 Modificar `infrastructure/messaging/contact_event_consumer.go`
- [ ] 🔴 Injetar `IdempotencyChecker` no construtor
- [ ] 🔴 Adicionar check de idempotência no início do handler:
  ```go
  if processed, _ := checker.IsProcessed(ctx, event.EventID(), "ContactEventConsumer"); processed {
    return nil // Skip
  }
  ```
- [ ] 🔴 Marcar como processado após sucesso:
  ```go
  checker.MarkAsProcessed(ctx, event.EventID(), "ContactEventConsumer")
  ```

**Arquivo:** `infrastructure/messaging/contact_event_consumer.go`

##### 4.4.2 WAHAMessageConsumer
- [ ] 🔴 Modificar `infrastructure/messaging/waha_message_consumer.go`
- [ ] 🔴 Injetar `IdempotencyChecker` no construtor
- [ ] 🔴 Adicionar check de idempotência
- [ ] 🔴 Marcar como processado após sucesso

**Arquivo:** `infrastructure/messaging/waha_message_consumer.go`

---

#### 4.5 Testes de Idempotência 🕐
- [ ] 🔴 Criar `infrastructure/persistence/idempotency_checker_test.go`
- [ ] 🔴 Teste: `TestIsProcessed_NotProcessed` - retorna false
- [ ] 🔴 Teste: `TestIsProcessed_AlreadyProcessed` - retorna true
- [ ] 🔴 Teste: `TestMarkAsProcessed_Success` - insere registro
- [ ] 🔴 Teste: `TestMarkAsProcessed_Duplicate` - não falha em duplicata

**Arquivo:** `infrastructure/persistence/idempotency_checker_test.go`

---

## 🟡 P1 - ALTO (Fazer em Seguida)

### 5. Observabilidade (OpenTelemetry)

#### 5.1 Setup OpenTelemetry 🕑
- [ ] 🟡 Instalar dependências
  ```bash
  go get go.opentelemetry.io/otel
  go get go.opentelemetry.io/otel/exporters/jaeger
  go get go.opentelemetry.io/otel/sdk/trace
  go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp
  ```
- [ ] 🟡 Criar `infrastructure/observability/tracing.go`
- [ ] 🟡 Implementar `InitTracer()` - configura Jaeger exporter
- [ ] 🟡 Implementar `ShutdownTracer()` - graceful shutdown
- [ ] 🟡 Adicionar variáveis de ambiente para Jaeger endpoint

**Arquivo:** `infrastructure/observability/tracing.go`

---

#### 5.2 HTTP Middleware 🕐
- [ ] 🟡 Criar `infrastructure/observability/http_middleware.go`
- [ ] 🟡 Implementar middleware `TracingMiddleware()` usando `otelhttp`
- [ ] 🟡 Adicionar span attributes (method, path, status_code)
- [ ] 🟡 Integrar no Gin router em `cmd/api/main.go`

**Arquivo:** `infrastructure/observability/http_middleware.go`

---

#### 5.3 Tracing em Use Cases 🕑
- [ ] 🟡 Criar helper `StartSpan(ctx, operationName)` em `observability/tracing.go`
- [ ] 🟡 Modificar `CreateContactUseCase.Execute()`
  - [ ] Adicionar span "CreateContact"
  - [ ] Adicionar attributes (projectID, tenantID)
  - [ ] Defer span.End()
- [ ] 🟡 Modificar `ProcessInboundMessageUseCase.Execute()`
  - [ ] Adicionar span "ProcessInboundMessage"
  - [ ] Adicionar child spans para cada etapa

**Arquivos a modificar:**
- `internal/application/contact/create_contact.go`
- `internal/application/message/process_inbound_message_usecase.go`

---

#### 5.4 Tracing em Repositórios 🕐
- [ ] 🟡 Modificar `GormContactRepository.Save()`
  - [ ] Adicionar span "ContactRepository.Save"
- [ ] 🟡 Modificar `GormContactRepository.FindByID()`
  - [ ] Adicionar span "ContactRepository.FindByID"
- [ ] 🟡 Aplicar mesmo padrão para outros repositórios

**Arquivos a modificar:**
- `infrastructure/persistence/gorm_contact_repository.go`
- `infrastructure/persistence/gorm_session_repository.go`

---

#### 5.5 Tracing em Event Bus 🕐
- [ ] 🟡 Modificar `DomainEventBus.Publish()`
- [ ] 🟡 Adicionar span "EventBus.Publish"
- [ ] 🟡 Injetar trace context nos headers do RabbitMQ
- [ ] 🟡 Extrair trace context nos consumers

**Arquivo:** `infrastructure/messaging/domain_event_bus.go`

---

### 6. Correlation ID

#### 6.1 Context Key 🕐
- [ ] 🟡 Criar `infrastructure/observability/context.go`
- [ ] 🟡 Definir `type contextKey string`
- [ ] 🟡 Definir constante `correlationIDKey = contextKey("correlation_id")`
- [ ] 🟡 Implementar `GetCorrelationID(ctx) string`
- [ ] 🟡 Implementar `SetCorrelationID(ctx, correlationID) context.Context`

**Arquivo:** `infrastructure/observability/context.go`

---

#### 6.2 HTTP Middleware 🕐
- [ ] 🟡 Criar `infrastructure/http/middleware/correlation_id.go`
- [ ] 🟡 Implementar `CorrelationIDMiddleware()`
  - [ ] Extrair header `X-Correlation-ID`
  - [ ] Se não existir, gerar novo UUID
  - [ ] Injetar no context
  - [ ] Adicionar ao response header
- [ ] 🟡 Integrar no Gin router

**Arquivo:** `infrastructure/http/middleware/correlation_id.go`

---

#### 6.3 Logging com Correlation ID 🕐
- [ ] 🟡 Modificar todos os logs para incluir correlation ID
- [ ] 🟡 Criar helper `LogWithCorrelation(ctx, logger)` que retorna logger com field
- [ ] 🟡 Exemplo:
  ```go
  logger := LogWithCorrelation(ctx, baseLogger)
  logger.Info("Processing message")
  ```

**Arquivo:** `infrastructure/observability/logging.go` (criar)

---

#### 6.4 Propagação via RabbitMQ 🕐
- [ ] 🟡 Modificar `DomainEventBus.Publish()`
- [ ] 🟡 Adicionar correlation ID nos headers AMQP
- [ ] 🟡 Modificar consumers para extrair correlation ID
- [ ] 🟡 Injetar no context do handler

**Arquivos a modificar:**
- `infrastructure/messaging/domain_event_bus.go`
- `infrastructure/messaging/contact_event_consumer.go`

---

### 7. Circuit Breakers

#### 7.1 Setup Circuit Breaker 🕑
- [ ] 🟡 Instalar dependência
  ```bash
  go get github.com/sony/gobreaker
  ```
- [ ] 🟡 Criar `infrastructure/resilience/circuit_breaker.go`
- [ ] 🟡 Implementar `NewCircuitBreaker(name, settings)` factory
- [ ] 🟡 Definir configuração padrão:
  - [ ] MaxRequests: 3
  - [ ] Interval: 60s
  - [ ] Timeout: 30s
  - [ ] ReadyToTrip: 5 falhas consecutivas

**Arquivo:** `infrastructure/resilience/circuit_breaker.go`

---

#### 7.2 Circuit Breaker para WAHA API 🕑
- [ ] 🟡 Modificar `infrastructure/channels/waha/client.go`
- [ ] 🟡 Adicionar field `cb *gobreaker.CircuitBreaker` no struct
- [ ] 🟡 Envolver todas as chamadas HTTP com circuit breaker:
  ```go
  result, err := w.cb.Execute(func() (interface{}, error) {
    return w.httpClient.Do(req)
  })
  ```
- [ ] 🟡 Adicionar logging quando circuit abre/fecha
- [ ] 🟡 Adicionar métricas

**Arquivo:** `infrastructure/channels/waha/client.go`

---

#### 7.3 Circuit Breaker para Webhooks 🕐
- [ ] 🟡 Modificar `infrastructure/webhooks/webhook_notifier.go`
- [ ] 🟡 Criar circuit breaker por webhook URL
- [ ] 🟡 Envolver chamadas HTTP com circuit breaker
- [ ] 🟡 Adicionar fallback quando circuit está aberto

**Arquivo:** `infrastructure/webhooks/webhook_notifier.go`

---

#### 7.4 Métricas de Circuit Breaker 🕐
- [ ] 🟡 Criar métricas Prometheus:
  - [ ] `circuit_breaker_state{name}` - gauge (0=closed, 1=open, 2=half-open)
  - [ ] `circuit_breaker_requests_total{name,result}` - counter
  - [ ] `circuit_breaker_failures_total{name}` - counter

**Arquivo:** `infrastructure/resilience/metrics.go` (criar)

---

### 8. Unit of Work Pattern

#### 8.1 Domain Interface 🕐
- [ ] 🟡 Criar `internal/domain/shared/unit_of_work.go`
- [ ] 🟡 Definir interface `UnitOfWork`:
  ```go
  type UnitOfWork interface {
    Begin(ctx context.Context) error
    Commit() error
    Rollback() error
    ContactRepository() contact.Repository
    SessionRepository() session.Repository
    MessageRepository() message.Repository
  }
  ```

**Arquivo:** `internal/domain/shared/unit_of_work.go`

---

#### 8.2 GORM Implementation 🕑
- [ ] 🟡 Criar `infrastructure/persistence/gorm_unit_of_work.go`
- [ ] 🟡 Implementar struct `GormUnitOfWork` com `*gorm.DB` e `*gorm.DB` (tx)
- [ ] 🟡 Implementar `Begin()` - inicia transação
- [ ] 🟡 Implementar `Commit()` - commita transação
- [ ] 🟡 Implementar `Rollback()` - rollback transação
- [ ] 🟡 Implementar getters de repositórios usando tx
- [ ] 🟡 Adicionar panic recovery em Rollback

**Arquivo:** `infrastructure/persistence/gorm_unit_of_work.go`

---

#### 8.3 Integração com Use Cases 🕑

##### 8.3.1 ProcessInboundMessageUseCase
- [ ] 🟡 Modificar construtor para receber `UnitOfWork`
- [ ] 🟡 Refatorar `Execute()` para usar transação:
  ```go
  uow.Begin(ctx)
  defer func() {
    if r := recover(); r != nil {
      uow.Rollback()
      panic(r)
    }
  }()
  
  // ... lógica de negócio usando uow.ContactRepository(), etc
  
  if err != nil {
    uow.Rollback()
    return err
  }
  
  uow.Commit()
  ```

**Arquivo:** `internal/application/message/process_inbound_message_usecase.go`

##### 8.3.2 CreateContactUseCase
- [ ] 🟡 Aplicar mesmo padrão de UoW

**Arquivo:** `internal/application/contact/create_contact.go`

---

#### 8.4 Testes de Unit of Work 🕐
- [ ] 🟡 Criar `infrastructure/persistence/gorm_unit_of_work_test.go`
- [ ] 🟡 Teste: `TestUnitOfWork_Commit` - verifica commit
- [ ] 🟡 Teste: `TestUnitOfWork_Rollback` - verifica rollback
- [ ] 🟡 Teste: `TestUnitOfWork_RollbackOnError` - rollback automático

**Arquivo:** `infrastructure/persistence/gorm_unit_of_work_test.go`

---

### 9. Compensação em Sagas

#### 9.1 Documentação de Estratégia 🕑
- [ ] 🟡 Criar `docs/saga_compensation_strategy.md`
- [ ] 🟡 Documentar fluxos de compensação:
  - [ ] ProcessInboundMessage workflow
  - [ ] SessionLifecycle workflow
- [ ] 🟡 Definir quando compensar vs quando apenas logar erro
- [ ] 🟡 Criar diagramas de fluxo

**Arquivo:** `docs/saga_compensation_strategy.md`

---

#### 9.2 Compensation Activities 🕑
- [ ] 🟡 Criar `internal/workflows/session/compensation_activities.go`
- [ ] 🟡 Implementar `DeleteContactActivity` - compensa criação de contato
- [ ] 🟡 Implementar `DeleteSessionActivity` - compensa criação de sessão
- [ ] 🟡 Implementar `DeleteMessageActivity` - compensa criação de mensagem
- [ ] 🟡 Adicionar logging estruturado

**Arquivo:** `internal/workflows/session/compensation_activities.go`

---

#### 9.3 Workflow com Compensação 🕒
- [ ] 🟡 Modificar `SessionLifecycleWorkflow`
- [ ] 🟡 Adicionar saga state tracking
- [ ] 🟡 Implementar lógica de compensação em caso de falha:
  ```go
  // Pseudo-código
  contactCreated := false
  sessionCreated := false
  
  defer func() {
    if err != nil {
      if sessionCreated {
        workflow.ExecuteActivity(ctx, DeleteSessionActivity, ...)
      }
      if contactCreated {
        workflow.ExecuteActivity(ctx, DeleteContactActivity, ...)
      }
    }
  }()
  ```
- [ ] 🟡 Adicionar retry policies para compensation activities

**Arquivo:** `internal/workflows/session/session_lifecycle_workflow.go`

---

#### 9.4 Saga State Tracking 🕐
- [ ] 🟡 Criar tabela `saga_state` para tracking:
  ```sql
  CREATE TABLE saga_state (
    id UUID PRIMARY KEY,
    workflow_id VARCHAR(255) NOT NULL,
    saga_type VARCHAR(100) NOT NULL,
    state JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
  );
  ```
- [ ] 🟡 Criar repository para saga state

**Arquivo:** `migrations/20250108_create_saga_state.sql`

---

## 🟢 P2 - MÉDIO (Fazer Depois)

### 10. Versionamento de Eventos

#### 10.1 Schema Evolution Strategy 🕐
- [ ] 🟢 Criar `docs/event_versioning_strategy.md`
- [ ] 🟢 Documentar estratégia de versionamento
- [ ] 🟢 Definir regras de compatibilidade:
  - [ ] Backward compatible: adicionar campos opcionais
  - [ ] Breaking change: incrementar versão
- [ ] 🟢 Documentar processo de migração

**Arquivo:** `docs/event_versioning_strategy.md`

---

#### 10.2 Adicionar Version aos Events 🕐
- [ ] 🟢 Modificar `internal/domain/shared/domain_event.go`
- [ ] 🟢 Adicionar método `EventVersion() string` na interface
- [ ] 🟢 Modificar todos os eventos para incluir versão:
  - [ ] `ContactCreatedEvent` → v1
  - [ ] `SessionStartedEvent` → v1
  - [ ] `MessageCreatedEvent` → v1

**Arquivos a modificar:**
- `internal/domain/shared/domain_event.go`
- `internal/domain/contact/events.go`
- `internal/domain/session/events.go`
- `internal/domain/message/events.go`

---

#### 10.3 Event Upcasters 🕑
- [ ] 🟢 Criar `infrastructure/messaging/event_upcaster.go`
- [ ] 🟢 Implementar interface `EventUpcaster`:
  ```go
  type EventUpcaster interface {
    CanUpcast(eventType string, version string) bool
    Upcast(eventData []byte, fromVersion string) ([]byte, error)
  }
  ```
- [ ] 🟢 Implementar upcasters específicos (quando necessário)
- [ ] 🟢 Integrar no consumer para upcast automático

**Arquivo:** `infrastructure/messaging/event_upcaster.go`

---

#### 10.4 Testes de Compatibilidade 🕐
- [ ] 🟢 Criar `infrastructure/messaging/event_compatibility_test.go`
- [ ] 🟢 Teste: eventos v1 podem ser deserializados
- [ ] 🟢 Teste: eventos v2 com campos novos não quebram v1
- [ ] 🟢 Teste: upcaster converte v1 para v2 corretamente

**Arquivo:** `infrastructure/messaging/event_compatibility_test.go`

---

### 11. Assemblers/Mappers Layer

#### 11.1 Contact Assembler 🕑
- [ ] 🟢 Criar `internal/application/assemblers/contact_assembler.go`
- [ ] 🟢 Implementar `ToDTO(contact *domain.Contact) ContactDTO`
- [ ] 🟢 Implementar `ToListDTO(contacts []*domain.Contact) []ContactDTO`
- [ ] 🟢 Implementar `FromCreateCommand(cmd CreateContactCommand) *domain.Contact`
- [ ] 🟢 Criar testes unitários

**Arquivo:** `internal/application/assemblers/contact_assembler.go`

---

#### 11.2 Session Assembler 🕐
- [ ] 🟢 Criar `internal/application/assemblers/session_assembler.go`
- [ ] 🟢 Implementar `ToDTO(session *domain.Session) SessionDTO`
- [ ] 🟢 Implementar `ToDetailDTO(session *domain.Session) SessionDetailDTO`
- [ ] 🟢 Criar testes unitários

**Arquivo:** `internal/application/assemblers/session_assembler.go`

---

#### 11.3 Message Assembler 🕐
- [ ] 🟢 Criar `internal/application/assemblers/message_assembler.go`
- [ ] 🟢 Implementar `ToDTO(message *domain.Message) MessageDTO`
- [ ] 🟢 Implementar `ToListDTO(messages []*domain.Message) []MessageDTO`
- [ ] 🟢 Criar testes unitários

**Arquivo:** `internal/application/assemblers/message_assembler.go`

---

#### 11.4 Refatorar Use Cases 🕑
- [ ] 🟢 Modificar `CreateContactUseCase` para usar assembler
- [ ] 🟢 Modificar handlers HTTP para usar assemblers
- [ ] 🟢 Remover lógica de mapeamento inline

**Arquivos a modificar:**
- `internal/application/contact/create_contact.go`
- `infrastructure/http/handlers/contact_handler.go`

---

### 12. Dependency Injection Container

#### 12.1 Avaliar DI Frameworks 🕐
- [ ] 🟢 Pesquisar pros/cons de wire vs dig vs fx
- [ ] 🟢 Criar POC com wire
- [ ] 🟢 Criar POC com dig
- [ ] 🟢 Decidir qual usar (recomendação: wire por ser compile-time)

**Arquivo:** `docs/adr/004-dependency-injection-framework.md` (criar)

---

#### 12.2 Implementar Wire 🕒
- [ ] 🟢 Instalar wire
  ```bash
  go install github.com/google/wire/cmd/wire@latest
  ```
- [ ] 🟢 Criar `cmd/api/wire.go` com providers
- [ ] 🟢 Criar `cmd/api/wire_gen.go` (gerado)
- [ ] 🟢 Organizar providers por camada:
  - [ ] Infrastructure providers
  - [ ] Application providers
  - [ ] Handler providers

**Arquivos:**
- `cmd/api/wire.go`
- `cmd/api/providers/infrastructure.go`
- `cmd/api/providers/application.go`

---

#### 12.3 Refatorar main.go 🕑
- [ ] 🟢 Simplificar `main.go` usando wire
- [ ] 🟢 Mover toda criação de dependências para providers
- [ ] 🟢 Reduzir `main.go` para ~50 linhas

**Arquivo:** `cmd/api/main.go`

---

### 13. Connection Pool Tuning

#### 13.1 Documentar Configuração Atual 🕐
- [ ] 🟢 Criar `docs/database_tuning.md`
- [ ] 🟢 Documentar configuração atual de pools
- [ ] 🟢 Documentar benchmarks de performance

**Arquivo:** `docs/database_tuning.md`

---

#### 13.2 Configuração Explícita 🕐
- [ ] 🟢 Modificar `infrastructure/persistence/database.go`
- [ ] 🟢 Adicionar configuração de pool:
  ```go
  sqlDB.SetMaxOpenConns(25)
  sqlDB.SetMaxIdleConns(5)
  sqlDB.SetConnMaxLifetime(5 * time.Minute)
  sqlDB.SetConnMaxIdleTime(10 * time.Minute)
  ```
- [ ] 🟢 Tornar configurável via env vars

**Arquivo:** `infrastructure/persistence/database.go`

---

#### 13.3 Métricas de Pool 🕐
- [ ] 🟢 Adicionar métricas Prometheus:
  - [ ] `db_connections_open`
  - [ ] `db_connections_idle`
  - [ ] `db_connections_in_use`
  - [ ] `db_connections_wait_duration`
- [ ] 🟢 Expor via endpoint `/metrics`

**Arquivo:** `infrastructure/observability/db_metrics.go` (criar)

---

### 14. Contract Tests para Eventos

#### 14.1 Setup Pact 🕑
- [ ] 🟢 Instalar Pact
  ```bash
  go get github.com/pact-foundation/pact-go/v2
  ```
- [ ] 🟢 Criar `tests/contracts/setup_test.go`
- [ ] 🟢 Configurar Pact broker (ou usar local)

**Arquivo:** `tests/contracts/setup_test.go`

---

#### 14.2 Contract Tests - Contact Events 🕑
- [ ] 🟢 Criar `tests/contracts/contact_events_test.go`
- [ ] 🟢 Definir contract para `contact.created`
- [ ] 🟢 Definir contract para `contact.updated`
- [ ] 🟢 Implementar provider verification

**Arquivo:** `tests/contracts/contact_events_test.go`

---

#### 14.3 Contract Tests - Session Events 🕐
- [ ] 🟢 Criar `tests/contracts/session_events_test.go`
- [ ] 🟢 Definir contract para `session.started`
- [ ] 🟢 Definir contract para `session.ended`

**Arquivo:** `tests/contracts/session_events_test.go`

---

#### 14.4 CI/CD Integration 🕐
- [ ] 🟢 Adicionar step de contract tests no CI
- [ ] 🟢 Publicar contracts no Pact broker
- [ ] 🟢 Adicionar verificação de breaking changes

**Arquivo:** `.github/workflows/ci.yml` (modificar)

---

### 15. Métricas Prometheus

#### 15.1 Setup Prometheus 🕐
- [ ] 🟢 Instalar dependência
  ```bash
  go get github.com/prometheus/client_golang/prometheus
  go get github.com/prometheus/client_golang/prometheus/promhttp
  ```
- [ ] 🟢 Criar `infrastructure/observability/metrics.go`
- [ ] 🟢 Implementar `InitMetrics()` - registra métricas

**Arquivo:** `infrastructure/observability/metrics.go`

---

#### 15.2 Business Metrics 🕑
- [ ] 🟢 Criar métricas de negócio:
  ```go
  contacts_created_total counter
  contacts_active gauge
  sessions_started_total counter
  sessions_active gauge
  sessions_duration_seconds histogram
  messages_processed_total counter{type, direction}
  messages_failed_total counter{reason}
  ```
- [ ] 🟢 Instrumentar use cases

**Arquivo:** `infrastructure/observability/business_metrics.go` (criar)

---

#### 15.3 Technical Metrics 🕐
- [ ] 🟢 Criar métricas técnicas:
  ```go
  event_processing_duration_seconds histogram{event_type}
  event_processing_errors_total counter{event_type}
  repository_operation_duration_seconds histogram{operation}
  http_request_duration_seconds histogram{method, path, status}
  ```

**Arquivo:** `infrastructure/observability/technical_metrics.go` (criar)

---

#### 15.4 Metrics Endpoint 🕐
- [ ] 🟢 Criar `infrastructure/http/handlers/metrics_handler.go`
- [ ] 🟢 Expor endpoint `GET /metrics`
- [ ] 🟢 Adicionar rota no router

**Arquivo:** `infrastructure/http/handlers/metrics_handler.go`

---

#### 15.5 Grafana Dashboards 🕑
- [ ] 🟢 Criar `monitoring/grafana/dashboards/ventros-crm.json`
- [ ] 🟢 Dashboard de negócio (contacts, sessions, messages)
- [ ] 🟢 Dashboard técnico (latency, errors, throughput)
- [ ] 🟢 Dashboard de infraestrutura (DB, Redis, RabbitMQ)

**Arquivo:** `monitoring/grafana/dashboards/ventros-crm.json` (criar)

---

## ⚪ P3 - BAIXO (Backlog)

### 16. Refatorações de Código

#### 16.1 Extrair Factories 🕐
- [ ] ⚪ Criar `cmd/api/factories/repository_factory.go`
- [ ] ⚪ Criar `cmd/api/factories/usecase_factory.go`
- [ ] ⚪ Criar `cmd/api/factories/handler_factory.go`
- [ ] ⚪ Refatorar `main.go` para usar factories

**Arquivos a criar:**
- `cmd/api/factories/repository_factory.go`
- `cmd/api/factories/usecase_factory.go`
- `cmd/api/factories/handler_factory.go`

---

#### 16.2 Consolidar DTOs 🕐
- [ ] ⚪ Criar package `internal/application/dtos`
- [ ] ⚪ Mover todos os DTOs para este package
- [ ] ⚪ Organizar por bounded context:
  - [ ] `contact_dtos.go`
  - [ ] `session_dtos.go`
  - [ ] `message_dtos.go`

**Arquivos a criar:**
- `internal/application/dtos/contact_dtos.go`
- `internal/application/dtos/session_dtos.go`
- `internal/application/dtos/message_dtos.go`

---

#### 16.3 Revisar Agregados 🕑
- [ ] ⚪ Analisar tamanho do agregado `Session` (457 linhas)
- [ ] ⚪ Avaliar se deve ser dividido
- [ ] ⚪ Considerar extrair `SessionMetrics` como value object
- [ ] ⚪ Considerar extrair `SessionAgents` como entidade

**Arquivo:** `docs/aggregate_review.md` (criar)

---

### 17. Documentação Arquitetural

#### 17.1 Architecture Decision Records 🕑
- [ ] ⚪ Criar `docs/adr/001-modular-monolith.md`
  - [ ] Por que não microservices?
  - [ ] Quando migrar para microservices?
- [ ] ⚪ Criar `docs/adr/002-saga-hybrid-approach.md`
  - [ ] Coreografia vs Orquestração
  - [ ] Quando usar cada um
- [ ] ⚪ Criar `docs/adr/003-multi-tenancy-rls.md`
  - [ ] Por que RLS?
  - [ ] Alternativas consideradas

**Arquivos a criar:**
- `docs/adr/001-modular-monolith.md`
- `docs/adr/002-saga-hybrid-approach.md`
- `docs/adr/003-multi-tenancy-rls.md`

---

#### 17.2 Diagramas de Fluxo 🕑
- [ ] ⚪ Criar diagrama de fluxo de mensagem inbound
- [ ] ⚪ Criar diagrama de saga ProcessInboundMessage
- [ ] ⚪ Criar diagrama de SessionLifecycle workflow
- [ ] ⚪ Usar Mermaid ou PlantUML

**Arquivos a criar:**
- `docs/diagrams/inbound_message_flow.md`
- `docs/diagrams/saga_flows.md`
- `docs/diagrams/session_lifecycle.md`

---

#### 17.3 Guia de Onboarding 🕑
- [ ] ⚪ Criar `docs/ONBOARDING.md`
- [ ] ⚪ Seções:
  - [ ] Setup do ambiente local
  - [ ] Arquitetura overview
  - [ ] Como adicionar novo agregado
  - [ ] Como adicionar novo use case
  - [ ] Como adicionar novo evento
  - [ ] Padrões de código
  - [ ] Como rodar testes

**Arquivo:** `docs/ONBOARDING.md`

---

### 18. Event Sourcing (Avaliação Futura)

#### 18.1 Research & POC 🕒
- [ ] ⚪ Pesquisar Event Sourcing patterns
- [ ] ⚪ Avaliar libraries (EventStore, custom)
- [ ] ⚪ Criar POC com agregado `Contact`
- [ ] ⚪ Documentar trade-offs:
  - [ ] Pros: audit trail completo, time travel, replay
  - [ ] Cons: complexidade, storage, eventual consistency

**Arquivo:** `docs/event_sourcing_evaluation.md` (criar)

---

#### 18.2 Event Store Design 🕑
- [ ] ⚪ Desenhar schema de event store
- [ ] ⚪ Definir estratégia de snapshots
- [ ] ⚪ Definir estratégia de projeções
- [ ] ⚪ Avaliar impacto em queries

**Arquivo:** `docs/event_store_design.md` (criar)

---

### 19. Performance Optimization

#### 19.1 Database Indexes 🕐
- [ ] ⚪ Analisar queries lentas com `EXPLAIN ANALYZE`
- [ ] ⚪ Adicionar índices faltantes:
  - [ ] `contacts(project_id, phone)`
  - [ ] `sessions(contact_id, status)`
  - [ ] `messages(session_id, timestamp)`
- [ ] ⚪ Documentar estratégia de indexação

**Arquivo:** `docs/database_indexes.md` (criar)

---

#### 19.2 Caching Strategy 🕑
- [ ] ⚪ Identificar queries cacheable
- [ ] ⚪ Implementar cache para:
  - [ ] Channel types (raramente muda)
  - [ ] Pipelines (raramente muda)
  - [ ] Project config (raramente muda)
- [ ] ⚪ Definir TTL por tipo de dado
- [ ] ⚪ Implementar cache invalidation

**Arquivo:** `infrastructure/cache/cache_strategy.go` (criar)

---

#### 19.3 Query Optimization 🕐
- [ ] ⚪ Revisar N+1 queries
- [ ] ⚪ Adicionar eager loading onde necessário
- [ ] ⚪ Implementar pagination em todas as listagens
- [ ] ⚪ Adicionar query timeouts

---

### 20. Security Hardening

#### 20.1 Input Validation 🕐
- [ ] ⚪ Adicionar validação em todos os handlers
- [ ] ⚪ Usar biblioteca de validação (go-playground/validator)
- [ ] ⚪ Sanitizar inputs para prevenir injection
- [ ] ⚪ Validar tamanho de payloads

---

#### 20.2 Rate Limiting 🕐
- [ ] ⚪ Implementar rate limiting por tenant
- [ ] ⚪ Implementar rate limiting por IP
- [ ] ⚪ Configurar limites por endpoint
- [ ] ⚪ Adicionar headers de rate limit

**Arquivo:** `infrastructure/http/middleware/rate_limiter.go` (criar)

---

#### 20.3 Secrets Management 🕐
- [ ] ⚪ Migrar secrets para vault/secrets manager
- [ ] ⚪ Remover secrets hardcoded
- [ ] ⚪ Implementar rotation de secrets
- [ ] ⚪ Auditar uso de secrets

---

## 📊 PROGRESSO GERAL

```
P0 - Crítico:     [░░░░░░░░░░] 0/95 tarefas (0%)
P1 - Alto:        [░░░░░░░░░░] 0/78 tarefas (0%)
P2 - Médio:       [░░░░░░░░░░] 0/62 tarefas (0%)
P3 - Baixo:       [░░░░░░░░░░] 0/35 tarefas (0%)
───────────────────────────────────────────────
TOTAL:            [░░░░░░░░░░] 0/270 tarefas (0%)
```

**Estimativa total**: ~40-50 sprints (80-100 semanas) para completar tudo

---

## 🎯 SPRINT PLANNING SUGERIDO

### **Sprint 1** (2 semanas) - 🔴 Fundação de Qualidade
**Objetivo**: Estabelecer base sólida de testes

- [ ] 1.1 Setup de Testes (1-2h)
- [ ] 1.2 Testes de Contact Aggregate (2-4h)
- [ ] 1.3 Testes de Session Aggregate (2-4h)
- [ ] 1.4 Testes de Message Aggregate (1-2h)
- [ ] 1.5 Testes de Value Objects (1-2h)

**Total estimado**: 7-14h

---

### **Sprint 2** (2 semanas) - 🔴 Testes de Integração
**Objetivo**: Garantir que repositórios funcionam corretamente

- [ ] 2.1 Setup de Testes de Integração (2-4h)
- [ ] 2.2 Testes de GormContactRepository (2-4h)
- [ ] 2.3 Testes de RLS (1-2h)

**Total estimado**: 5-10h

---

### **Sprint 3** (2 semanas) - 🔴 Outbox Pattern
**Objetivo**: Garantir entrega confiável de eventos

- [ ] 3.1 Database Schema (1-2h)
- [ ] 3.2 Domain Interface (1-2h)
- [ ] 3.3 Infrastructure Implementation (2-4h)
- [ ] 3.4 Outbox Processor Worker (2-4h)
- [ ] 3.5 Integração com Use Cases (2-4h)
- [ ] 3.6 Startup Integration (1-2h)

**Total estimado**: 9-18h

---

### **Sprint 4** (2 semanas) - 🔴 Idempotência
**Objetivo**: Prevenir processamento duplicado de eventos

- [ ] 4.1 Database Schema (1-2h)
- [ ] 4.2 Adicionar EventID aos Domain Events (1-2h)
- [ ] 4.3 Idempotency Checker Service (2-4h)
- [ ] 4.4 Integração com Consumers (2-4h)
- [ ] 4.5 Testes de Idempotência (1-2h)

**Total estimado**: 7-14h

---

### **Sprint 5** (2 semanas) - 🟡 OpenTelemetry
**Objetivo**: Adicionar distributed tracing

- [ ] 5.1 Setup OpenTelemetry (2-4h)
- [ ] 5.2 HTTP Middleware (1-2h)
- [ ] 5.3 Tracing em Use Cases (2-4h)
- [ ] 5.4 Tracing em Repositórios (1-2h)
- [ ] 5.5 Tracing em Event Bus (1-2h)

**Total estimado**: 7-14h

---

### **Sprint 6** (2 semanas) - 🟡 Correlation ID
**Objetivo**: Rastrear requisições através do sistema

- [ ] 6.1 Context Key (1-2h)
- [ ] 6.2 HTTP Middleware (1-2h)
- [ ] 6.3 Logging com Correlation ID (1-2h)
- [ ] 6.4 Propagação via RabbitMQ (1-2h)

**Total estimado**: 4-8h

---

### **Sprint 7-8** (4 semanas) - 🟡 Circuit Breakers & Unit of Work
**Objetivo**: Aumentar resiliência do sistema

**Sprint 7:**
- [ ] 7.1 Setup Circuit Breaker (2-4h)
- [ ] 7.2 Circuit Breaker para WAHA API (2-4h)
- [ ] 7.3 Circuit Breaker para Webhooks (1-2h)
- [ ] 7.4 Métricas de Circuit Breaker (1-2h)

**Sprint 8:**
- [ ] 8.1 Domain Interface UoW (1-2h)
- [ ] 8.2 GORM Implementation (2-4h)
- [ ] 8.3 Integração com Use Cases (2-4h)
- [ ] 8.4 Testes de Unit of Work (1-2h)

**Total estimado**: 12-24h

---

## 📝 GUIA DE USO

### Como Começar
1. **Escolha uma tarefa P0** (crítica)
2. **Crie uma branch**: `feature/p0-1.1-setup-testes`
3. **Marque como [⏳]** no TODO
4. **Implemente** seguindo o detalhamento
5. **Crie testes** (sempre!)
6. **Abra PR** com descrição clara
7. **Após merge, marque [✅]**

### Convenções de Branch
- `feature/p0-X.Y-nome-curto` - Features P0
- `feature/p1-X.Y-nome-curto` - Features P1
- `refactor/nome-curto` - Refatorações
- `docs/nome-curto` - Documentação
- `test/nome-curto` - Testes

### Convenções de Commit
```
feat(domain): add Contact aggregate tests
test(infra): add GormContactRepository integration tests
refactor(app): extract contact assembler
docs(adr): add decision record for outbox pattern
fix(infra): prevent duplicate event processing
```

### Checklist de PR
- [ ] Código implementado e funcionando
- [ ] Testes unitários criados (cobertura > 80%)
- [ ] Testes de integração (se aplicável)
- [ ] Documentação atualizada
- [ ] Logs estruturados adicionados
- [ ] Sem breaking changes (ou documentado)
- [ ] CI/CD passando

---

## 🎓 RECURSOS DE APRENDIZADO

### Livros Recomendados
- **Domain-Driven Design** - Eric Evans
- **Implementing Domain-Driven Design** - Vaughn Vernon
- **Clean Architecture** - Robert C. Martin
- **Building Event-Driven Microservices** - Adam Bellemare

### Artigos & Blogs
- [Martin Fowler - Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html)
- [Microsoft - Saga Pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/saga)
- [Microservices.io - Patterns](https://microservices.io/patterns/)

### Cursos
- Udemy: Microservices with DDD, SAGA, Outbox & Kafka
- Pluralsight: Domain-Driven Design Fundamentals

---

## 📞 SUPORTE

### Dúvidas sobre Arquitetura
- Consultar `ARCHITECTURE.md`
- Consultar `PLAN.md` para análise detalhada
- Abrir issue com label `question`

### Reportar Problemas
- Abrir issue com label `bug`
- Incluir logs e contexto
- Seguir template de issue

---

**Última atualização**: 2025-10-08  
**Próxima revisão**: A cada sprint  
**Responsável**: Time de Desenvolvimento  
**Versão**: 2.0 (Micro-segmentado)
