# Relatório de Cobertura de Entidades - Ventros CRM

Verificação completa de todas as entidades do sistema em todas as camadas (Domain, Application, Infrastructure).

## ✅ Entidades Completas (Domain + Application + Infrastructure)

| Entidade | Domain | Application | Entity (Infra) | Repository (Infra) | Status |
|----------|--------|-------------|----------------|-------------------|--------|
| **Agent** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **BillingAccount** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Channel** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **ChannelType** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Contact** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **ContactEvent** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **ContactList** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Credential** | ✅ | ❌ | ✅ | ✅ | **PARCIAL** (falta Application) |
| **Message** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Note** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Pipeline** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Project** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Session** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Tracking** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |
| **Webhook** | ✅ | ✅ | ✅ | ✅ | **COMPLETO** |

## 📦 Entidades de Infraestrutura (Sem Domain - OK)

Estas entidades são técnicas e não precisam de modelo de domínio:

| Entidade | Domain | Application | Entity (Infra) | Repository (Infra) | Status |
|----------|--------|-------------|----------------|-------------------|--------|
| **OutboxEvent** | ✅ (interface) | N/A | ✅ | ✅ | **OK** (Outbox Pattern) |
| **ProcessedEvent** | N/A | N/A | ✅ | N/A | **OK** (Idempotency) |
| **DomainEventLog** | N/A | N/A | ✅ | ✅ | **OK** (Audit Log) |

## 🔧 Entidades Especiais

| Entidade | Domain | Application | Entity (Infra) | Repository (Infra) | Observação |
|----------|--------|-------------|----------------|-------------------|------------|
| **AutomationRule** | ✅ (em Pipeline) | ✅ | ✅ | ✅ | **COMPLETO** |
| **AgentSession** | ✅ | ❌ | ✅ | ❌ | **PARCIAL** (falta Application) |
| **TrackingEnrichment** | ✅ (em Tracking) | ✅ | ✅ | ✅ | **COMPLETO** (parte de Tracking) |

## ⚠️ Entidades com Pendências

### 1. Credential (Falta Application Layer)

**Domain**: ✅ Completo
- `/internal/domain/credential/credential.go`
- `/internal/domain/credential/repository.go`
- Tipos: OAuth, API Key, etc.

**Infrastructure**: ✅ Completo
- `entities/credential.go`
- `gorm_credential_repository.go`

**Application**: ❌ **FALTANDO**
- Não existe `/internal/application/credential/`
- **Ações necessárias**:
  - Criar use cases: `CreateCredentialUseCase`, `GetCredentialUseCase`, `DeleteCredentialUseCase`
  - Gerenciamento de credenciais OAuth
  - Renovação de tokens

### 2. AgentSession (Falta Application Layer)

**Domain**: ✅ Completo
- `/internal/domain/agent_session/agent_session.go`
- `/internal/domain/agent_session/repository.go`

**Infrastructure**: ✅ Completo
- `entities/agent_session.go`

**Application**: ❌ **FALTANDO**
- Não existe `/internal/application/agent_session/`
- **Ações necessárias**:
  - Criar use cases para gerenciar sessões de agentes
  - Atribuição de agentes a sessões
  - Histórico de atendimentos

**Repository**: ❌ **FALTANDO**
- Não existe `gorm_agent_session_repository.go`

## 📋 Checklist de Implementação

### Prioridade ALTA ⚠️

- [ ] **Credential Application Layer**
  - [ ] Criar `/internal/application/credential/create_credential.go`
  - [ ] Criar `/internal/application/credential/get_credential.go`
  - [ ] Criar `/internal/application/credential/delete_credential.go`
  - [ ] Criar `/internal/application/credential/refresh_oauth_token.go`

- [ ] **AgentSession Application Layer**
  - [ ] Criar `/internal/application/agent_session/assign_agent.go`
  - [ ] Criar `/internal/application/agent_session/end_agent_session.go`
  - [ ] Criar `/infrastructure/persistence/gorm_agent_session_repository.go`

### Prioridade MÉDIA

- [ ] **Broadcast** (Entidade no Domain mas sem implementação completa)
  - Diretório existe: `/internal/domain/broadcast/`
  - Verificar se deve ser implementado ou removido

## 🎯 Resumo Geral

| Status | Quantidade | Entidades |
|--------|------------|-----------|
| **✅ COMPLETO** | 15 | Agent, BillingAccount, Channel, ChannelType, Contact, ContactEvent, ContactList, Message, Note, Pipeline, Project, Session, Tracking, Webhook, AutomationRule |
| **⚠️ PARCIAL** | 2 | Credential (falta Application), AgentSession (falta Application + Repository) |
| **📦 INFRAESTRUTURA** | 3 | OutboxEvent, ProcessedEvent, DomainEventLog (não precisam Domain) |

## ✅ Verificações de Qualidade

### Domain Layer
- ✅ Todos os agregados têm eventos de domínio
- ✅ Todos os agregados têm repositórios definidos
- ✅ Value Objects implementados (Email, Phone, TenantID, etc.)

### Application Layer
- ✅ Use cases seguem padrão Command/Query
- ✅ DTOs para requests/responses
- ⚠️ Faltam use cases para Credential e AgentSession

### Infrastructure Layer
- ✅ Todas as entities GORM criadas
- ✅ Migrations criadas (000001 até 000023)
- ✅ Repositories implementam interfaces do Domain
- ✅ RLS configurado para multi-tenancy

## 🔗 Associações entre Entidades

### Relacionamentos 1:N
- **User** → Projects (1:N)
- **Project** → Pipelines (1:N)
- **Project** → Contacts (1:N)
- **Project** → Channels (1:N)
- **Pipeline** → Statuses (1:N)
- **Contact** → Sessions (1:N)
- **Contact** → Messages (1:N)
- **Contact** → ContactEvents (1:N)
- **Session** → Messages (1:N)

### Relacionamentos N:M
- **Contact** ↔ **Pipeline** (através de `contact_pipeline_statuses`)
- **ContactList** ↔ **Contact** (através de tabela pivot)

### Relacionamentos Especiais
- **Channel** → **Pipeline** (opcional, para auto-atribuição)
- **Tracking** → **TrackingEnrichment** (1:N, enriquecimento de dados de anúncios)

## 🚀 Próximos Passos

1. **Implementar Credential Application Layer** (CRÍTICO para OAuth do Meta)
2. **Implementar AgentSession Application Layer** (para atribuição de agentes)
3. **Revisar entidade Broadcast** (decidir se mantém ou remove)
4. **Adicionar handlers HTTP** para as novas entidades
5. **Documentar APIs no Swagger**

---
**Relatório gerado em**: 2025-10-09
**Versão do sistema**: 1.0.0
