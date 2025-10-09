# 🏗️ PLANO DE AUDITORIA ARQUITETURAL - Ventros CRM

> **Objetivo**: Avaliar a conformidade da arquitetura do Ventros CRM com os princípios de DDD, Clean Architecture, Event-Driven Architecture e padrões SAGA.

---

## 📋 ÍNDICE

1. [Estrutura de Camadas & DDD](#1-estrutura-de-camadas--ddd)
2. [Camada de Application (Use Cases)](#2-camada-de-application-use-cases)
3. [Camada de Infrastructure](#3-camada-de-infrastructure)
4. [Sagas - Coreografia vs Orquestração](#4-sagas---coreografia-vs-orquestração)
5. [Event-Driven & Mensageria](#5-event-driven--mensageria)
6. [Temporal Workflows](#6-temporal-workflows)
7. [Persistência & Transações](#7-persistência--transações)
8. [SOLID & Clean Architecture](#8-solid--clean-architecture)
9. [Testing Strategy](#9-testing-strategy)
10. [Observabilidade & Resiliência](#10-observabilidade--resiliência)

---

## 1. ESTRUTURA DE CAMADAS & DDD

### 1.1 Pureza das Entidades de Domínio

**Questões de Auditoria:**
- ✅ As entidades do domínio estão 100% livres de dependências externas (sem tags de ORM, JSON, etc)?
- ✅ A camada de domínio possui apenas regras de negócio puras, sem lógica de persistência ou comunicação?
- ✅ Os agregados estão bem definidos com boundaries transacionais claros?
- ⚠️ Cada bounded context é um serviço separado ou estão no mesmo monorepo?

**Análise Atual:**
- ✅ **Entidades puras**: `Contact`, `Session`, `Message` não possuem tags de ORM/JSON
- ✅ **Campos privados**: Todos os campos são privados com getters públicos
- ✅ **Factory methods**: `NewContact()`, `NewSession()`, `NewMessage()`
- ✅ **Reconstruct methods**: Para reconstrução a partir do repositório
- ✅ **Value Objects**: `Email`, `Phone` implementados corretamente
- ✅ **Domain Events**: Eventos gerados internamente pelos agregados
- ⚠️ **Bounded Contexts**: Todos no mesmo monorepo (modular monolith)

**Pontos de Atenção:**
- Verificar se há vazamento de conceitos de infraestrutura no domínio
- Avaliar se os agregados têm tamanho adequado (não muito grandes)
- Confirmar que invariantes são sempre protegidas

---

## 2. CAMADA DE APPLICATION (USE CASES)

### 2.1 Orquestração vs Lógica de Negócio

**Questões de Auditoria:**
- ✅ Os use cases orquestram apenas a lógica de aplicação, delegando regras para o domínio?
- ✅ Eles dependem de interfaces (ports) definidas no domínio?
- ⚠️ Há DTO/Request-Response objects separados das entidades de domínio?
- ⚠️ Como é feito o mapeamento entre DTOs e entidades?

**Análise Atual:**
- ✅ **Use Cases**: `CreateContactUseCase`, `ProcessInboundMessageUseCase` orquestram fluxo
- ✅ **Dependency Inversion**: Use cases dependem de `contact.Repository` (interface no domínio)
- ✅ **Command/Query Objects**: `CreateContactCommand`, `CreateProjectRequest`
- ⚠️ **DTOs misturados**: Alguns DTOs têm tags JSON (não ideal para camada de aplicação)
- ❌ **Assemblers/Mappers**: Não há camada dedicada de mapeamento (feito inline nos use cases)

**Pontos de Atenção:**
- Criar camada de Assemblers/Mappers dedicada
- Separar DTOs de entrada (Commands) de DTOs de saída (Responses)
- Avaliar se há lógica de negócio vazando para os use cases

---

## 3. CAMADA DE INFRASTRUCTURE

### 3.1 Implementação de Adapters

**Questões de Auditoria:**
- ✅ As implementações de repositórios estão na infra injetando dependências via DI?
- ✅ Os adapters do RabbitMQ e Temporal estão implementados como adapters na infra?
- ⚠️ Como é gerenciado o connection pool do Postgres e Redis?
- ⚠️ Há factories ou builders para criação de dependências?

**Análise Atual:**
- ✅ **Repositórios**: `GormContactRepository` implementa `contact.Repository`
- ✅ **Mapeamento Domain↔Entity**: Métodos `domainToEntity()` e `entityToDomain()`
- ✅ **Event Bus Adapter**: `DomainEventBus` implementa interface do domínio
- ✅ **Dependency Injection**: Manual no `main.go` (constructor injection)
- ⚠️ **Connection Pool**: Configurado via GORM, mas sem tuning explícito
- ❌ **Factory Pattern**: Não há factories, tudo criado manualmente no main

**Pontos de Atenção:**
- Criar factories para criação de dependências complexas
- Documentar configuração de connection pools
- Avaliar uso de DI container (wire, dig, fx)

---

## 4. SAGAS - COREOGRAFIA VS ORQUESTRAÇÃO

### 4.1 Padrões de Saga Implementados

**Questões de Auditoria:**
- ✅ **Coreografia (RabbitMQ)**: Quais eventos são publicados? Cada serviço escuta eventos e reage autonomamente?
- ✅ **Orquestração (Temporal)**: Quais workflows são orquestrados pelo Temporal? Por que escolheram híbrido?
- ⚠️ Como garantem idempotência nos handlers de eventos?
- ❌ Há compensação automática em caso de falha nas sagas?

**Análise Atual:**

#### Coreografia (RabbitMQ)
- ✅ **Eventos Publicados**:
  - `domain.events.contact.created`
  - `domain.events.contact.updated`
  - `domain.events.session.started`
  - `domain.events.session.ended`
  - `domain.events.message.created`
- ✅ **Consumers**: `ContactEventConsumer`, `WAHAMessageConsumer`
- ✅ **Dead Letter Queue**: Configurada com 3 retries

#### Orquestração (Temporal)
- ✅ **Workflows**:
  - `SessionLifecycleWorkflow`: Gerencia timeout de sessões
  - `SessionCleanupWorkflow`: Limpeza periódica
  - `WAHAImportWorkflow`: Importação de histórico
- ✅ **Activities**: `EndSessionActivity`, `CleanupSessionsActivity`
- ⚠️ **Compensação**: Não há compensação explícita (apenas logs de erro)

**Pontos de Atenção:**
- Implementar idempotency keys para handlers
- Adicionar compensação automática nos workflows
- Documentar quando usar coreografia vs orquestração

---

## 5. EVENT-DRIVEN & MENSAGERIA

### 5.1 Garantias de Entrega e Consistência

**Questões de Auditoria:**
- ❌ Os eventos são versionados? Como lidam com backward compatibility?
- ✅ Há dead letter queue configurada no RabbitMQ?
- ❌ Os eventos são persistidos antes de publicar (outbox pattern)?
- ⚠️ Como garantem exactly-once delivery ou at-least-once com idempotência?

**Análise Atual:**
- ✅ **DLQ**: Configurada com 3 retries via `DeclareQueueWithDLQ()`
- ✅ **Event Log**: `DomainEventLogRepository` salva eventos (mas não é outbox)
- ❌ **Outbox Pattern**: Não implementado (eventos publicados diretamente)
- ❌ **Versionamento**: Eventos não têm campo de versão
- ⚠️ **Idempotência**: Não há mecanismo explícito

**Pontos de Atenção:**
- Implementar Transactional Outbox Pattern
- Adicionar versionamento de eventos (schema evolution)
- Implementar idempotency keys nos consumers
- Avaliar uso de event sourcing para agregados críticos

---

## 6. TEMPORAL WORKFLOWS

### 6.1 Determinismo e Boas Práticas

**Questões de Auditoria:**
- ✅ Os workflows do Temporal são determinísticos?
- ✅ Há separação clara entre Activities (side effects) e Workflows (lógica)?
- ⚠️ Como lidam com long-running workflows e versioning?
- ✅ Há retry policies e timeouts configurados adequadamente?

**Análise Atual:**
- ✅ **Determinismo**: Workflows usam apenas `workflow.Context` e timers
- ✅ **Activities**: Side effects isolados em `EndSessionActivity`, `CleanupSessionsActivity`
- ✅ **Timeouts**: `StartToCloseTimeout: 30s` configurado
- ✅ **Signals**: `session-activity` para resetar timeout
- ⚠️ **Versioning**: Não há estratégia de versioning documentada
- ⚠️ **Long-running**: Workflows podem durar 30+ minutos (OK, mas sem monitoramento)

**Pontos de Atenção:**
- Documentar estratégia de versioning de workflows
- Adicionar métricas para workflows long-running
- Implementar circuit breakers nas activities

---

## 7. PERSISTÊNCIA & TRANSAÇÕES

### 7.1 Consistência e Transações

**Questões de Auditoria:**
- ⚠️ **PostgreSQL**: Há uso de transactions para garantir consistência nos agregados?
- ✅ **Redis**: É usado para cache, sessions, ou também como event store?
- ⚠️ Como lidam com eventual consistency entre serviços?
- ❌ Há implementação de Unit of Work pattern?

**Análise Atual:**
- ✅ **GORM**: Usado como ORM principal
- ✅ **RLS (Row Level Security)**: Implementado para multi-tenancy
- ⚠️ **Transações**: Não há uso explícito de `db.Transaction()` nos use cases
- ✅ **Redis**: Usado para cache (não como event store)
- ❌ **Unit of Work**: Não implementado

**Pontos de Atenção:**
- Implementar Unit of Work pattern para transações complexas
- Adicionar transações explícitas em use cases críticos
- Documentar estratégia de eventual consistency

---

## 8. SOLID & CLEAN ARCHITECTURE

### 8.1 Princípios SOLID

**Questões de Auditoria:**
- ✅ **Dependency Inversion**: Todas as dependências apontam para o domínio?
- ✅ **Interface Segregation**: As interfaces são específicas por caso de uso?
- ✅ **Single Responsibility**: Cada use case tem uma única responsabilidade?
- ⚠️ Há algum service locator ou as dependências são injetadas via construtor?

**Análise Atual:**
- ✅ **DIP**: Interfaces no domínio, implementações na infra
- ✅ **ISP**: Interfaces pequenas e focadas (`Repository`, `EventBus`)
- ✅ **SRP**: Use cases têm responsabilidade única
- ✅ **Constructor Injection**: Todas as dependências injetadas via construtor
- ⚠️ **Service Locator**: Não usado, mas `main.go` está muito grande

**Pontos de Atenção:**
- Refatorar `main.go` usando DI container ou factories
- Avaliar se há violações de OCP (Open/Closed Principle)

---

## 9. TESTING STRATEGY

### 9.1 Cobertura de Testes

**Questões de Auditoria:**
- ⚠️ Os testes unitários do domínio não dependem de infraestrutura?
- ⚠️ Há testes de integração para repositórios e adapters?
- ❌ Como testam os workflows do Temporal?
- ❌ Há contract tests para os eventos do RabbitMQ?

**Análise Atual:**
- ✅ **E2E Tests**: `api_test.go`, `waha_webhook_test.go`
- ❌ **Unit Tests**: Não encontrados para domínio
- ❌ **Integration Tests**: Não encontrados para repositórios
- ❌ **Workflow Tests**: Não encontrados para Temporal
- ❌ **Contract Tests**: Não implementados

**Pontos de Atenção:**
- Criar testes unitários para agregados do domínio
- Implementar testes de integração para repositórios
- Adicionar testes para workflows do Temporal
- Implementar contract tests para eventos

---

## 10. OBSERVABILIDADE & RESILIÊNCIA

### 10.1 Monitoramento e Resiliência

**Questões de Auditoria:**
- ❌ Há distributed tracing entre os serviços (OpenTelemetry)?
- ⚠️ Como fazem correlation de eventos na saga?
- ❌ Há circuit breakers e retry mechanisms?
- ✅ Metrics e logs estruturados implementados?

**Análise Atual:**
- ✅ **Structured Logging**: Zap logger usado em todo o sistema
- ✅ **Health Checks**: `HealthChecker` para DB, Redis, RabbitMQ, Temporal
- ❌ **Distributed Tracing**: Não implementado
- ⚠️ **Correlation ID**: Não há propagação de correlation ID
- ❌ **Circuit Breakers**: Não implementados
- ✅ **Retry**: RabbitMQ DLQ com 3 retries

**Pontos de Atenção:**
- Implementar OpenTelemetry para tracing
- Adicionar correlation ID em contextos
- Implementar circuit breakers para chamadas externas
- Adicionar métricas Prometheus

---

## 📊 RESUMO EXECUTIVO

### ✅ Pontos Fortes
1. **DDD bem implementado**: Entidades puras, agregados claros, eventos de domínio
2. **Clean Architecture**: Separação clara de camadas, dependency inversion
3. **Event-Driven**: RabbitMQ com DLQ, eventos de domínio bem estruturados
4. **Temporal**: Workflows determinísticos, activities isoladas
5. **Multi-tenancy**: RLS implementado corretamente

### ⚠️ Pontos de Melhoria
1. **Outbox Pattern**: Não implementado (risco de perda de eventos)
2. **Unit of Work**: Transações não gerenciadas explicitamente
3. **Assemblers**: Mapeamento inline nos use cases
4. **DI Container**: Dependências criadas manualmente no main
5. **Versionamento**: Eventos não versionados

### ❌ Pontos Críticos
1. **Testes**: Cobertura insuficiente (sem unit tests de domínio)
2. **Observabilidade**: Sem distributed tracing
3. **Idempotência**: Não garantida nos event handlers
4. **Compensação**: Não implementada nas sagas
5. **Circuit Breakers**: Não implementados

---

## 🎯 PRÓXIMOS PASSOS

Após revisar este plano, seguir para o **TODO.md** para checklist detalhado de ações.

**Priorização Sugerida:**
1. **P0 (Crítico)**: Testes, Outbox Pattern, Idempotência
2. **P1 (Alto)**: Observabilidade, Circuit Breakers, Unit of Work
3. **P2 (Médio)**: Versionamento de eventos, DI Container, Assemblers
4. **P3 (Baixo)**: Refatorações, documentação adicional

---

**Última atualização**: 2025-10-08  
**Autor**: Auditoria Arquitetural Automatizada  
**Versão**: 1.0
