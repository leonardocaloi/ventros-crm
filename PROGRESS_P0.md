# P0 IMPLEMENTATION PROGRESS - Optimistic Locking + Aggregate Root + Unit of Work

**Data Início**: 2025-10-12
**Sprint**: 1-2 (Semanas 1-4)
**Objetivo**: Corrigir gaps críticos de consistência e integridade de dados

---

## ✅ CONCLUÍDO

### 1. Erro de Optimistic Locking ✅
- **Arquivo**: `internal/domain/core/shared/errors.go`
- **Mudanças**:
  - Adicionado `ErrorTypeOptimisticLock`
  - Criado `NewOptimisticLockError()` constructor
  - Criado `IsOptimisticLockError()` helper
- **Status**: ✅ **COMPLETO**

### 2. Interface AggregateRoot ✅
- **Arquivo**: `internal/domain/core/shared/aggregate.go` (NOVO)
- **Conteúdo**:
  ```go
  type AggregateRoot interface {
      ID() uuid.UUID
      Version() int
      DomainEvents() []DomainEvent
      ClearEvents()
  }
  ```
- **Documentação**: Inclui princípios DDD, referências (Evans, Vernon)
- **Status**: ✅ **COMPLETO**

### 3. Migration de Optimistic Locking ✅
- **Arquivos**:
  - `infrastructure/database/migrations/000046_add_optimistic_locking.up.sql`
  - `infrastructure/database/migrations/000046_add_optimistic_locking.down.sql`
- **Tabelas Modificadas**: 15 aggregate roots
  - contacts, sessions, channels, agents, pipelines, chats, projects
  - billing_accounts, campaigns, sequences, broadcasts, credentials
  - contact_lists, pipelines_statuses, webhook_subscriptions
- **Mudanças**:
  - Adicionada coluna `version INTEGER DEFAULT 1 NOT NULL` em todas
  - Criados índices compostos `(id, version)` para performance
  - Comentários documentando propósito
- **Status**: ✅ **COMPLETO** (migration criada, não executada ainda)

### 4. Aggregate Contact Atualizado ✅
- **Arquivo**: `internal/domain/crm/contact/contact.go`
- **Mudanças**:
  - Adicionado campo `version int`
  - Construtor `NewContact()` inicializa version = 1
  - `ReconstructContact()` aceita version como parâmetro
  - Adicionado método `Version() int`
  - Import de `shared` package
  - Verificação compile-time: `var _ shared.AggregateRoot = (*Contact)(nil)`
- **Status**: ✅ **COMPLETO**

---

## 🔄 EM PROGRESSO

### 5. Adicionar version aos demais agregados
**Agregados Pendentes**: 14 agregados faltantes

#### Lista de Agregados a Atualizar:

| # | Aggregate | Arquivo | Status |
|---|-----------|---------|--------|
| 1 | ✅ Contact | `crm/contact/contact.go` | ✅ COMPLETO |
| 2 | ⏳ Session | `crm/session/session.go` | 🔄 PENDENTE |
| 3 | ⏳ Channel | `crm/channel/channel.go` | 🔄 PENDENTE |
| 4 | ⏳ Agent | `crm/agent/agent.go` | 🔄 PENDENTE |
| 5 | ⏳ Pipeline | `crm/pipeline/pipeline.go` | 🔄 PENDENTE |
| 6 | ⏳ Chat | `crm/chat/chat.go` | 🔄 PENDENTE |
| 7 | ⏳ Project | `core/project/project.go` | 🔄 PENDENTE |
| 8 | ⏳ BillingAccount | `core/billing/billing_account.go` | 🔄 PENDENTE |
| 9 | ⏳ Campaign | `automation/campaign/campaign.go` | 🔄 PENDENTE |
| 10 | ⏳ Sequence | `automation/sequence/sequence.go` | 🔄 PENDENTE |
| 11 | ⏳ Broadcast | `automation/broadcast/broadcast.go` | 🔄 PENDENTE |
| 12 | ⏳ Credential | `crm/credential/credential.go` | 🔄 PENDENTE |
| 13 | ⏳ ContactList | `crm/contact_list/contact_list.go` | 🔄 PENDENTE |
| 14 | ⏳ Webhook | `crm/webhook/webhook.go` | 🔄 PENDENTE |

#### Template para Atualização de Agregado:

Para cada agregado, seguir estes passos:

**1. Adicionar campo version**:
```go
type <Aggregate> struct {
    id      uuid.UUID
    version int    // Optimistic locking - prevents lost updates
    // ... outros campos
}
```

**2. Atualizar construtor**:
```go
func New<Aggregate>(...) (*<Aggregate>, error) {
    // ... validações
    aggregate := &<Aggregate>{
        id:      uuid.New(),
        version: 1, // Start with version 1 for new aggregates
        // ... outros campos
    }
    return aggregate, nil
}
```

**3. Atualizar Reconstruct**:
```go
func Reconstruct<Aggregate>(
    id uuid.UUID,
    version int, // Optimistic locking version
    // ... outros parâmetros
) *<Aggregate> {
    if version == 0 {
        version = 1 // Default to version 1 (backwards compatibility)
    }
    return &<Aggregate>{
        id:      id,
        version: version,
        // ... outros campos
    }
}
```

**4. Adicionar método Version()**:
```go
// Aggregate Root implementation
func (a *<Aggregate>) ID() uuid.UUID { return a.id }
func (a *<Aggregate>) Version() int  { return a.version }
```

**5. Adicionar import e verificação de interface**:
```go
import (
    // ... outros imports
    "github.com/ventros-crm/internal/domain/core/shared"
)

// No final do arquivo:
var _ shared.AggregateRoot = (*<Aggregate>)(nil)
```

---

## 📋 PENDENTE

### 6. Modificar Repositórios para Optimistic Locking
**Status**: 🔄 PENDENTE

#### Repositórios a Modificar:

| # | Repository | Arquivo | Adapter | Status |
|---|------------|---------|---------|--------|
| 1 | ContactRepository | `gorm_contact_repository.go` | `contact_adapter.go` | 🔄 PENDENTE |
| 2 | SessionRepository | `gorm_session_repository.go` | `session_adapter.go` | 🔄 PENDENTE |
| 3 | ChannelRepository | `gorm_channel_repository.go` | `channel_adapter.go` | 🔄 PENDENTE |
| 4 | AgentRepository | `gorm_agent_repository.go` | `agent_adapter.go` | 🔄 PENDENTE |
| 5 | PipelineRepository | `gorm_pipeline_repository.go` | `pipeline_adapter.go` | 🔄 PENDENTE |
| 6 | ChatRepository | `gorm_chat_repository.go` | `chat_adapter.go` | 🔄 PENDENTE |
| 7 | ProjectRepository | `gorm_project_repository.go` | `project_adapter.go` | 🔄 PENDENTE |
| 8 | BillingAccountRepository | `gorm_billing_account_repository.go` | `billing_adapter.go` | 🔄 PENDENTE |
| 9 | CampaignRepository | `gorm_campaign_repository.go` | `campaign_adapter.go` | 🔄 PENDENTE |
| 10 | SequenceRepository | `gorm_sequence_repository.go` | `sequence_adapter.go` | 🔄 PENDENTE |

#### Template para Atualização de Repository:

**A. Atualizar Adapter (toDomain)**:
```go
func (a *ContactAdapter) toDomain(entity ContactEntity) *contact.Contact {
    return contact.ReconstructContact(
        entity.ID,
        entity.Version, // ✅ ADICIONAR ISTO
        entity.ProjectID,
        // ... outros parâmetros
    )
}
```

**B. Atualizar Adapter (toEntity)**:
```go
func (a *ContactAdapter) toEntity(c *contact.Contact) ContactEntity {
    return ContactEntity{
        ID:      c.ID(),
        Version: c.Version(), // ✅ ADICIONAR ISTO
        // ... outros campos
    }
}
```

**C. Atualizar Repository Save() com Optimistic Locking**:
```go
func (r *GORMContactRepository) Save(ctx context.Context, c *contact.Contact) error {
    entity := r.adapter.toEntity(c)

    // Optimistic locking: WHERE id = ? AND version = ?
    result := r.db.WithContext(ctx).
        Model(&ContactEntity{}).
        Where("id = ? AND version = ?", c.ID(), c.Version()).
        Updates(map[string]interface{}{
            "name":       entity.Name,
            "email":      entity.Email,
            // ... todos os campos mutáveis
            "version":    c.Version() + 1, // ✅ INCREMENT VERSION
            "updated_at": time.Now(),
        })

    if result.Error != nil {
        return result.Error
    }

    // Check if update succeeded (version matched)
    if result.RowsAffected == 0 {
        return shared.NewOptimisticLockError(
            "contact",
            c.ID().String(),
            c.Version(),
            -1, // actual version unknown
        )
    }

    return nil
}
```

---

### 7. Interface Unit of Work
**Status**: 🔄 PENDENTE

**Arquivo a Criar**: `infrastructure/persistence/unit_of_work.go`

```go
package persistence

import "context"

// UnitOfWork represents a database transaction boundary
type UnitOfWork interface {
    // Begin starts a new transaction
    Begin(ctx context.Context) error

    // Commit commits the transaction
    Commit() error

    // Rollback rolls back the transaction
    Rollback() error

    // Repository accessors - all operate within the same transaction
    ContactRepository() ContactRepository
    SessionRepository() SessionRepository
    MessageRepository() MessageRepository
    ChannelRepository() ChannelRepository
    AgentRepository() AgentRepository
    PipelineRepository() PipelineRepository
    CampaignRepository() CampaignRepository
    SequenceRepository() SequenceRepository
    BillingAccountRepository() BillingAccountRepository
    ProjectRepository() ProjectRepository
}
```

---

### 8. Implementação GORM Unit of Work
**Status**: 🔄 PENDENTE

**Arquivo a Criar**: `infrastructure/persistence/gorm_unit_of_work.go`

```go
package persistence

import (
    "context"
    "gorm.io/gorm"
)

type GORMUnitOfWork struct {
    db *gorm.DB
    tx *gorm.DB
}

func NewGORMUnitOfWork(db *gorm.DB) *GORMUnitOfWork {
    return &GORMUnitOfWork{db: db}
}

func (uow *GORMUnitOfWork) Begin(ctx context.Context) error {
    uow.tx = uow.db.WithContext(ctx).Begin()
    return uow.tx.Error
}

func (uow *GORMUnitOfWork) Commit() error {
    if uow.tx == nil {
        return nil
    }
    err := uow.tx.Commit().Error
    uow.tx = nil
    return err
}

func (uow *GORMUnitOfWork) Rollback() error {
    if uow.tx != nil {
        err := uow.tx.Rollback().Error
        uow.tx = nil
        return err
    }
    return nil
}

// Repository accessors - return repositories using the transaction
func (uow *GORMUnitOfWork) ContactRepository() ContactRepository {
    return NewGORMContactRepository(uow.tx)
}

// ... implementar para todos repositórios
```

---

### 9. Refatorar Use Cases para Unit of Work
**Status**: 🔄 PENDENTE

#### Exemplo de Refatoração:

**ANTES** (múltiplas transações):
```go
func (uc *ProcessInboundMessageUseCase) Execute(msg WAHAMessage) error {
    contact := uc.findOrCreateContact(msg)
    uc.contactRepo.Save(contact) // Transaction 1

    session := uc.findOrCreateSession(contact.ID())
    uc.sessionRepo.Save(session) // Transaction 2

    message := createMessage(msg, session.ID())
    uc.messageRepo.Save(message) // Transaction 3 - pode falhar!

    return nil
}
```

**DEPOIS** (transação única):
```go
func (uc *ProcessInboundMessageUseCase) Execute(msg WAHAMessage) error {
    uow := uc.uowFactory.NewUnitOfWork()

    if err := uow.Begin(ctx); err != nil {
        return err
    }
    defer uow.Rollback() // Rollback if not committed

    contact := uc.findOrCreateContact(msg, uow.ContactRepository())
    session := uc.findOrCreateSession(contact.ID(), uow.SessionRepository())
    message := createMessage(msg, session.ID(), uow.MessageRepository())

    // Atomic commit - all or nothing
    return uow.Commit()
}
```

---

### 10. ADR - Aggregate Boundaries
**Status**: 🔄 PENDENTE

**Arquivo a Criar**: `docs/architecture/decisions/001-aggregate-boundaries.md`

**Conteúdo mínimo**:
- Listar todos aggregates e seus limites transacionais
- Justificar decisões (por que Message não é aggregate root)
- Documentar invariantes protegidos em cada aggregate
- Referências: Evans (DDD), Vernon (IDDD)

---

### 11. Testes de Concorrência
**Status**: 🔄 PENDENTE

**Arquivo a Criar**: `internal/domain/crm/contact/contact_concurrency_test.go`

```go
func TestContact_OptimisticLocking_PreventsConcurrentUpdates(t *testing.T) {
    // Setup
    repo := setupTestRepository()
    contact, _ := contact.NewContact(uuid.New(), "tenant1", "John")
    repo.Save(context.Background(), contact)

    // Simulate concurrent updates
    c1, _ := repo.FindByID(context.Background(), contact.ID()) // version 1
    c2, _ := repo.FindByID(context.Background(), contact.ID()) // version 1

    // User 1 updates
    c1.UpdateName("Alice")
    err1 := repo.Save(context.Background(), c1) // version 2
    assert.NoError(t, err1)

    // User 2 tries to update with stale version
    c2.UpdateName("Bob")
    err2 := repo.Save(context.Background(), c2) // version 1 (stale!)

    // Should fail with optimistic lock error
    assert.Error(t, err2)
    assert.True(t, shared.IsOptimisticLockError(err2))

    // Verify final state
    final, _ := repo.FindByID(context.Background(), contact.ID())
    assert.Equal(t, "Alice", final.Name()) // User 1's change preserved
    assert.Equal(t, 2, final.Version())
}
```

---

## 📊 MÉTRICAS DE PROGRESSO

| Categoria | Completo | Pendente | Total | % Completo |
|-----------|----------|----------|-------|------------|
| **Erros & Interfaces** | 2 | 0 | 2 | 100% ✅ |
| **Migrations** | 1 | 0 | 1 | 100% ✅ |
| **Agregados (domain)** | 1 | 14 | 15 | 7% 🔴 |
| **Repositórios** | 0 | 10 | 10 | 0% 🔴 |
| **Unit of Work** | 0 | 2 | 2 | 0% 🔴 |
| **Use Cases** | 0 | ? | ? | 0% 🔴 |
| **Documentação (ADR)** | 0 | 1 | 1 | 0% 🔴 |
| **Testes** | 0 | 1 | 1 | 0% 🔴 |
| **TOTAL** | 4 | 28+ | 32+ | 13% 🔴 |

---

## 🎯 PRÓXIMOS PASSOS RECOMENDADOS

### **Sequência Ótima de Implementação**:

1. **Semana 1** - Atualizar todos agregados (14 pendentes)
   - Use o template fornecido
   - Verifique compile-time com `var _`
   - Teste compilação incremental

2. **Semana 2** - Executar migration + Atualizar adapters
   - Rodar `000046_add_optimistic_locking.up.sql`
   - Atualizar todos adapters (toDomain, toEntity)
   - Verificar mappers

3. **Semana 3** - Atualizar repositórios + Unit of Work
   - Modificar Save() em todos repositórios
   - Implementar UnitOfWork
   - Testar transações atômicas

4. **Semana 4** - Refatorar use cases + Testes + ADR
   - Migrar use cases para UoW
   - Escrever testes de concorrência
   - Documentar aggregate boundaries

---

## ⚠️ RISCOS E MITIGAÇÕES

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| Breaking changes em adapters | Alta | Alto | Testar com dados reais, rollback pronto |
| Performance degradation | Média | Médio | Benchmark antes/depois, índices criados |
| Complexidade de refatoração | Alta | Alto | Fazer incrementalmente, CI/CD em cada step |
| Bugs de concorrência | Média | Alto | Testes de concorrência abrangentes |

---

## 📚 REFERÊNCIAS

- **AI_REPORT.md** - Seção GAP 1, GAP 2, GAP 3 (linhas 586-726)
- **Eric Evans** - Domain-Driven Design (2003) - Cap. 6 (Aggregates)
- **Vaughn Vernon** - IDDD (2013) - Cap. 10 (Aggregates), Cap. 14 (Repositories)
- **Martin Fowler** - Patterns of EA (2002) - Unit of Work pattern

---

**Última Atualização**: 2025-10-12
**Próxima Revisão**: 2025-10-13
