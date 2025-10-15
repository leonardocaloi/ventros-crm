# Template de Prompt para Novas Features

**Use este template toda vez que solicitar uma nova feature**

---

## 📋 Template de Solicitação

```
# Feature Request: [NOME DA FEATURE]

## 📖 Contexto Obrigatório

Antes de começar a implementação, VOCÊ DEVE:

1. **ANALISAR** o código existente relacionado
2. **CONSULTAR** as documentações relevantes listadas abaixo
3. **IDENTIFICAR** o bounded context e aggregate corretos
4. **PROPOR** a solução arquitetural ANTES de codificar
5. **CONFIRMAR** comigo a abordagem antes de implementar

## 📚 Documentações de Referência (LEIA ANTES!)

### Documentação Principal
- [ ] `README.md` - Visão geral do projeto
- [ ] `DEV_GUIDE.md` - Guia completo de desenvolvimento (CRÍTICO!)
- [ ] `AI_REPORT.md` - Auditoria arquitetural (8.2/10)
- [ ] `P0.md` - Padrão Command Handler (100% implementado)
- [ ] `TODO.md` - Roadmap e prioridades

### Domain Mapping (23 Aggregates)
- [ ] `guides/domain_mapping/README.md` - Overview de todos aggregates
- [ ] `guides/domain_mapping/[AGGREGATE]_aggregate.md` - Aggregate específico relacionado

### Guias Técnicos
- [ ] `guides/ACTORS.md` - Atores do sistema e permissões
- [ ] `guides/MAKEFILE.md` - Comandos de desenvolvimento
- [ ] `guides/TESTING.md` - Estratégia de testes
- [ ] `MIGRATIONS.md` - Guia de migrations SQL

## 🎯 Descrição da Feature

**O que preciso implementar:**

[DESCREVA AQUI A FEATURE EM DETALHES]

**Problema de negócio que resolve:**

[EXPLIQUE O PROBLEMA DE NEGÓCIO]

**Atores envolvidos:**

[QUEM USA ESTA FEATURE? Ex: Admin, Agent, Contact, System]

**Fluxo de negócio esperado:**

1. [PASSO 1]
2. [PASSO 2]
3. [PASSO 3]

**Integrações externas necessárias:**

[WAHA? Stripe? Temporal? Nenhuma?]

---

## ⚠️ INSTRUÇÕES PARA A IA (IMPORTANTE!)

### Fase 1: ANÁLISE (Obrigatória antes de codificar!)

**VOCÊ DEVE COMPLETAR ESTA ANÁLISE ANTES DE PROPOR CÓDIGO:**

1. **Identificar Bounded Context**
   - [ ] CRM (Contact, Session, Message, Channel, Pipeline, Agent, Chat)?
   - [ ] Automation (Campaign, Sequence)?
   - [ ] Core (Billing, Project, Customer)?
   - [ ] Novo context?

2. **Identificar Aggregate Responsável**
   - [ ] Qual aggregate root?
   - [ ] Consultar: `guides/domain_mapping/[aggregate]_aggregate.md`
   - [ ] Precisa criar novo aggregate?
   - [ ] Quais invariantes de negócio proteger?

3. **Verificar Código Existente**
   - [ ] Existe código similar? Onde?
   - [ ] Que padrões já estão implementados?
   - [ ] Há migrations relacionadas?
   - [ ] Há testes existentes para consultar?

4. **Analisar Domain Events**
   - [ ] Quais eventos serão emitidos?
   - [ ] Nomenclatura: `aggregate.action` (ex: `contact.created`)
   - [ ] Consultar eventos existentes em `internal/domain/[context]/[aggregate]/events.go`
   - [ ] Quem consome esses eventos?

5. **Verificar Dependências**
   - [ ] Precisa de novo repositório?
   - [ ] Precisa de integration externa?
   - [ ] Precisa de Temporal workflow?
   - [ ] Precisa de cache Redis?

### Fase 2: PROPOSTA (Apresente ANTES de codificar!)

**APRESENTE ESTA PROPOSTA PARA EU APROVAR:**

```
## 📋 Proposta de Implementação

### 1. Bounded Context e Aggregate
- **Context**: [CRM | Automation | Core]
- **Aggregate**: [Contact | Session | Message | etc]
- **Aggregate Root**: [Nome do aggregate root]
- **Arquivo**: `internal/domain/[context]/[aggregate]/[aggregate].go`

### 2. Camadas DDD a Implementar

#### Domain Layer (internal/domain/)
- [ ] Aggregate root: `[aggregate].go`
- [ ] Value objects: `value_objects.go` (se necessário)
- [ ] Domain events: `events.go`
- [ ] Repository interface: `repository.go`
- [ ] Errors: `errors.go`
- [ ] Tests: `[aggregate]_test.go`

#### Application Layer (internal/application/)
- [ ] Command: `commands/[aggregate]/[action]_[aggregate]_command.go`
- [ ] Command Handler: `commands/[aggregate]/[action]_[aggregate]_handler.go`
- [ ] Tests: `commands/[aggregate]/[action]_[aggregate]_handler_test.go`

#### Infrastructure Layer (infrastructure/)
- [ ] GORM Entity: `persistence/entities/[aggregate]_entity.go`
- [ ] Repository Impl: `persistence/gorm_[aggregate]_repository.go`
- [ ] Repository Tests: `persistence/gorm_[aggregate]_repository_test.go`
- [ ] HTTP Handler: `http/handlers/[aggregate]_handler.go`
- [ ] Routes: `http/routes/[aggregate]_routes.go`
- [ ] Migration: `database/migrations/[number]_[description].up.sql`
- [ ] Migration Down: `database/migrations/[number]_[description].down.sql`

### 3. Domain Events a Emitir
- `[aggregate].[action1]` - [Quando ocorre]
- `[aggregate].[action2]` - [Quando ocorre]

### 4. Database Schema
```sql
-- Proposta de tabela
CREATE TABLE [table_name] (
    id UUID PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,  -- ✅ Optimistic locking
    project_id UUID NOT NULL,
    tenant_id TEXT NOT NULL,

    -- Business fields
    [field1] [type],

    -- Audit
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP  -- ✅ Soft delete
);
```

### 5. Estratégia de Testes
- **Unit Tests**: [X testes no domain + application]
- **Integration Tests**: [Y testes no repository]
- **E2E Tests**: [Z testes no HTTP endpoint]
- **Coverage Goal**: Domain 100%, Application 80%+

### 6. Riscos e Considerações
- [RISCO 1]: [Como mitigar]
- [RISCO 2]: [Como mitigar]

**AGUARDO SUA APROVAÇÃO PARA PROSSEGUIR COM A IMPLEMENTAÇÃO.**
```

### Fase 3: IMPLEMENTAÇÃO (Após aprovação!)

**SÓ IMPLEMENTE APÓS EU APROVAR A PROPOSTA ACIMA!**

**Durante a implementação, SIGA RIGOROSAMENTE:**

#### ✅ Checklist Obrigatório (DEV_GUIDE.md)

1. **Domain Layer**
   - [ ] Aggregate com `version` field (optimistic locking)
   - [ ] Factory method `New[Aggregate]()`
   - [ ] Business methods (não setters genéricos!)
   - [ ] Domain events emitidos em mudanças de estado
   - [ ] Value objects validados no construtor
   - [ ] Getters públicos para campos privados
   - [ ] Tests unitários (100% coverage no domain)

2. **Application Layer**
   - [ ] Command struct com `Validate()` method
   - [ ] Command handler com dependências injetadas
   - [ ] Lógica de negócio NO handler, não no HTTP handler
   - [ ] Event publishing via `EventBus.Publish()`
   - [ ] Tests unitários com mocks (80%+ coverage)

3. **Infrastructure Layer**
   - [ ] GORM entity com tags corretas
   - [ ] Repository com optimistic locking no `Save()`
   - [ ] Mappers: `toDomain()` e `toEntity()`
   - [ ] HTTP handler APENAS como adaptador (sem lógica!)
   - [ ] DTOs: `[Action]Request` e `[Aggregate]Response`
   - [ ] Swagger comments completos
   - [ ] Tests de integração (repository)
   - [ ] Tests E2E (HTTP endpoint)

4. **Database**
   - [ ] Migration `.up.sql` criada
   - [ ] Migration `.down.sql` criada
   - [ ] `version` column adicionada (optimistic locking)
   - [ ] `deleted_at` column adicionada (soft delete)
   - [ ] `tenant_id` column adicionada (multi-tenancy)
   - [ ] Indexes criados (performance)
   - [ ] Foreign keys definidas
   - [ ] RLS policy configurada (multi-tenancy)

5. **Events**
   - [ ] Nomenclatura: `aggregate.action` (lowercase)
   - [ ] Payload com todos campos necessários
   - [ ] Event metadata incluído
   - [ ] Published via Outbox Pattern
   - [ ] Consumer implementado (se necessário)

6. **Tests**
   - [ ] Unit tests escritos (domain + application)
   - [ ] Integration tests escritos (repository)
   - [ ] E2E tests escritos (HTTP)
   - [ ] `make test-unit` passa
   - [ ] `make test-integration` passa
   - [ ] `make test-e2e` passa
   - [ ] Coverage: Domain 100%, Application 80%+

7. **Documentation**
   - [ ] Swagger comments adicionados
   - [ ] Domain aggregate doc atualizado (`guides/domain_mapping/`)
   - [ ] README.md atualizado (se feature importante)
   - [ ] TODO.md atualizado (se feature grande)

8. **Code Quality**
   - [ ] `make fmt` executado
   - [ ] `make build` passa sem erros
   - [ ] `go vet` passa sem warnings
   - [ ] Nomenclatura seguindo padrões (DEV_GUIDE.md)
   - [ ] Sem unused imports

### Fase 4: REVIEW (Antes de finalizar!)

**CHECKLIST FINAL (APRESENTE PARA EU REVISAR):**

```
## ✅ Implementação Completa - Review Checklist

### Arquivos Criados/Modificados

**Domain Layer**:
- [ ] `internal/domain/[context]/[aggregate]/[aggregate].go`
- [ ] `internal/domain/[context]/[aggregate]/events.go`
- [ ] `internal/domain/[context]/[aggregate]/repository.go`
- [ ] `internal/domain/[context]/[aggregate]/[aggregate]_test.go`

**Application Layer**:
- [ ] `internal/application/commands/[aggregate]/[command]_command.go`
- [ ] `internal/application/commands/[aggregate]/[command]_handler.go`
- [ ] `internal/application/commands/[aggregate]/[command]_handler_test.go`

**Infrastructure Layer**:
- [ ] `infrastructure/persistence/entities/[aggregate]_entity.go`
- [ ] `infrastructure/persistence/gorm_[aggregate]_repository.go`
- [ ] `infrastructure/persistence/gorm_[aggregate]_repository_test.go`
- [ ] `infrastructure/http/handlers/[aggregate]_handler.go`
- [ ] `infrastructure/http/routes/[aggregate]_routes.go`

**Database**:
- [ ] `infrastructure/database/migrations/[number]_[name].up.sql`
- [ ] `infrastructure/database/migrations/[number]_[name].down.sql`

**Documentation**:
- [ ] `guides/domain_mapping/[aggregate]_aggregate.md` (atualizado)

### Tests Executados
- [ ] `make test-unit` - ✅ PASSOU
- [ ] `make test-integration` - ✅ PASSOU
- [ ] `make test-e2e` - ✅ PASSOU
- [ ] `make test-coverage` - Coverage: [X]%

### Build & Quality
- [ ] `make build` - ✅ PASSOU
- [ ] `make fmt` - ✅ EXECUTADO
- [ ] `go vet` - ✅ SEM WARNINGS

### Métricas
- **Linhas de código**: [X] linhas
- **Test coverage**: [Y]%
- **Arquivos criados**: [Z] arquivos
- **Arquivos modificados**: [W] arquivos

**IMPLEMENTAÇÃO PRONTA PARA REVIEW!**
```

---

## 📝 Exemplo de Uso

### ❌ ERRADO (Não faça assim!)

```
"Crie um CRUD de contatos"
```

### ✅ CORRETO (Faça assim!)

```
# Feature Request: CRUD de Contatos

## 📖 Contexto Obrigatório

Antes de começar, VOCÊ DEVE:
1. Analisar `guides/domain_mapping/contact_aggregate.md`
2. Consultar `DEV_GUIDE.md` para padrões
3. Ver exemplo em `P0.md` (Contact Handler já refatorado)
4. Propor solução arquitetural ANTES de codificar

## 📚 Documentações de Referência (LEIA!)
- [x] DEV_GUIDE.md
- [x] guides/domain_mapping/contact_aggregate.md
- [x] P0.md

## 🎯 Descrição da Feature

**O que preciso implementar:**
Sistema completo de CRUD para gerenciamento de contatos no CRM.

**Problema de negócio:**
Agentes precisam criar, visualizar, editar e excluir contatos manualmente.

**Atores envolvidos:**
- Admin: Acesso total
- Agent: Criar, visualizar, editar contatos do seu projeto
- Viewer: Apenas visualizar

**Fluxo de negócio:**
1. Agent acessa tela de contatos
2. Clica em "Novo Contato"
3. Preenche: Nome (obrigatório), Email, Phone, Tags
4. Sistema valida dados
5. Sistema cria contato
6. Sistema emite evento `contact.created`
7. Sistema exibe mensagem de sucesso

**Integrações externas:**
Nenhuma (apenas banco de dados e RabbitMQ para eventos)

---

## ⚠️ INSTRUÇÕES PARA A IA

[SEGUIR TODAS AS FASES: ANÁLISE → PROPOSTA → APROVAÇÃO → IMPLEMENTAÇÃO → REVIEW]
```

---

## 🎯 Comandos Úteis Durante Desenvolvimento

```bash
# Análise
make test-unit              # Verificar se não quebrou nada
make build                  # Verificar se compila

# Desenvolvimento
make infra                  # Iniciar infraestrutura
make api                    # Iniciar API
make migrate-up             # Aplicar migrations

# Testing
make test                   # Todos os testes
make test-coverage          # Ver coverage

# Quality
make fmt                    # Formatar código
make vet                    # Verificar warnings
```

---

## 📞 Dúvidas?

**Arquitetura**: Consulte [AI_REPORT.md](AI_REPORT.md)
**Domain Model**: Consulte [guides/domain_mapping/](guides/domain_mapping/)
**Padrões**: Consulte [DEV_GUIDE.md](DEV_GUIDE.md)
**Comandos**: Consulte [MAKEFILE.md](MAKEFILE.md)

---

**Versão**: 1.0
**Última Atualização**: 2025-10-12
**Status**: ✅ Production-Ready
