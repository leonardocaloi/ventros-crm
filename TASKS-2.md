# 🤖 PROMPT: ANÁLISE ARQUITETURAL DDD COMPLETA

> **Instruções:** Copie o prompt abaixo e execute. A IA sobrescreverá este arquivo com a análise completa.

---

## 📝 PROMPT PARA IA (COPIE TUDO ABAIXO)

```
# ANÁLISE ARQUITETURAL DDD - VENTROS CRM

Você é um arquiteto de software especialista em Domain-Driven Design (DDD), Clean Architecture e padrões táticos/estratégicos.

**IMPORTANTE:** Você tem acesso completo ao código-fonte do projeto em `/home/caloi/ventros-crm/`. 

**TAREFA:** Analise TODO o código-fonte e gere uma avaliação arquitetural COMPLETA e DETALHADA seguindo EXATAMENTE a estrutura abaixo. Sobrescreva o arquivo `/home/caloi/ventros-crm/TASKS-2.md` com sua análise.

---

## ⚠️ REGRA CRÍTICA: TRAZER TUDO

**VOCÊ DEVE LISTAR:**

1. ✅ **O QUE JÁ ESTÁ IMPLEMENTADO** - Tudo que existe no código, mesmo que:
   - Não tenha testes
   - Não tenha lógica completa
   - Seja apenas estrutura/interface
   - Tenha comentários TODO/FIXME
   - Esteja parcialmente implementado

2. ❌ **O QUE NÃO ESTÁ IMPLEMENTADO** - Tudo que deveria existir mas não existe:
   - Agregados ausentes
   - Value Objects faltantes
   - Domain Services não implementados
   - Specifications ausentes
   - Factories ausentes
   - Testes ausentes
   - Documentação ausente

3. ⚠️ **O QUE ESTÁ INCOMPLETO** - Código que existe mas precisa de trabalho:
   - Métodos vazios
   - TODOs no código
   - Validações faltantes
   - Tratamento de erros incompleto

**FORMATO PARA CADA ELEMENTO:**

```markdown
#### [Nome do Elemento]
**Status:** ✅ Implementado / ⚠️ Parcial / ❌ Ausente
**Localização:** [caminho do arquivo] (se existir)
**Implementação:** [% de completude estimado]
**Detalhes:** [o que tem e o que falta]
```

**EXEMPLO:**
```markdown
#### Agregado: Contact
**Status:** ✅ Implementado
**Localização:** `/internal/domain/contact/contact.go`
**Implementação:** 95%
**Detalhes:** 
- ✅ Entidade Contact completa
- ✅ Value Objects: Email, Phone
- ✅ Domain Events: ContactCreated, ContactUpdated
- ✅ Repository Interface
- ⚠️ Falta: Specification para filtros complexos
- ❌ Falta: Testes unitários para método UpdateProfilePicture()
```

---

## 📋 CONTEXTO DO SISTEMA

**Nome:** Ventros CRM  
**Domínio:** Customer Relationship Management (CRM)  
**Stack:** Go (Golang), GORM, GIN, PostgreSQL, RabbitMQ, Temporal, Redis  

**Funcionalidades Principais:**
- Gestão de contatos e conversas multicanal
- Pipeline de vendas e automações
- Processamento de mensagens com IA
- Tracking e analytics
- Billing e subscriptions

**Estrutura de Pastas Esperada:**
```
/home/caloi/ventros-crm/
├── internal/
│   ├── domain/           # Camada de Domínio (agregados, VOs, events)
│   └── application/      # Camada de Aplicação (use cases, DTOs)
├── infrastructure/       # Camada de Infraestrutura (repos, DB, HTTP)
├── cmd/                  # Entry points
└── ...
```

---

## 🎯 ESTRUTURA OBRIGATÓRIA DA ANÁLISE

Sua análise DEVE seguir EXATAMENTE esta estrutura. Não pule seções.

---

### 1. SUMÁRIO EXECUTIVO

Crie uma tabela com notas gerais:

| Camada | Nota | Status | Observações |
|--------|------|--------|-------------|
| **Domínio** | X/10 | ✅/⚠️/❌ | [observação breve] |
| **Aplicação** | X/10 | ✅/⚠️/❌ | [observação breve] |
| **Infraestrutura** | X/10 | ✅/⚠️/❌ | [observação breve] |
| **Interface** | X/10 | ✅/⚠️/❌ | [observação breve] |
| **Eventos** | X/10 | ✅/⚠️/❌ | [observação breve] |

**Pontuação Geral: X.X/10**

---

### 2. BOUNDED CONTEXTS IDENTIFICADOS

Liste TODOS os bounded contexts encontrados no código:

Para cada bounded context:
- **Nome do Bounded Context**
- **Agregados principais**
- **Status da implementação** (completo/parcial/inicial)
- **Nota** (0-10)
- **Observações**

Exemplo:
```
### BC: Contact Management
- **Agregados:** Contact, ContactList, ContactEvent
- **Status:** Completo
- **Nota:** 9.0/10
- **Observações:** Bem implementado com VOs (Email, Phone)
```

---

### 3. CAMADA DE DOMÍNIO - ANÁLISE DETALHADA

**Instruções:** Explore `/internal/domain/` e liste TODOS os elementos encontrados.

#### 3.1. AGREGADOS (Aggregate Roots)

Para CADA agregado encontrado (E CADA AGREGADO AUSENTE), crie uma seção assim:

```markdown
#### Agregado: [Nome]
**Status:** ✅ Implementado / ⚠️ Parcial / ❌ Ausente
**Localização:** `/internal/domain/[pasta]/[arquivo].go` (se existir)
**Implementação:** [X%]

**Entidades:**
- [Nome da Entity Root] ✅/⚠️/❌
- [Outras entities, se houver] ✅/⚠️/❌

**Value Objects:**
- [Nome do VO] ✅ Implementado / ❌ Ausente (deveria ter)
- [Outro VO] ✅/❌

**Enums/Types:**
- [Nome do Enum/Type] ✅ (valores: x, y, z) / ❌ Ausente

**Domain Events:**
- [NomeDoEvento] ✅ Implementado / ⚠️ Definido mas não usado / ❌ Ausente
- [OutroEvento] ✅/⚠️/❌

**Repository Interface:**
- [NomeDoRepository] ✅ Completo / ⚠️ Parcial / ❌ Ausente
- Localização: `/internal/domain/[pasta]/repository.go`
- Métodos: [liste todos os métodos da interface]

**Repository Implementation:**
- [NomeGormRepository] ✅ Implementado / ⚠️ Parcial / ❌ Ausente
- Localização: `/infrastructure/persistence/gorm_[nome]_repository.go`
- Métodos implementados: [X/Y]

**Métodos de Negócio:**
Liste TODOS os métodos públicos (implementados E vazios):
- `MetodoX(params) error` ✅ Implementado
- `MetodoY(params)` ⚠️ Implementado mas sem validação
- `MetodoZ(params)` ❌ Definido mas vazio (TODO)

**Invariantes Protegidas:**
- ✅ [Invariante 1 - implementada]
- ⚠️ [Invariante 2 - parcialmente validada]
- ❌ [Invariante 3 - não validada]

**Testes:**
- Testes unitários: ✅ Sim / ⚠️ Parcial / ❌ Não
- Cobertura estimada: [X%]
- Localização: [caminho do arquivo _test.go]

**Nota:** X/10 ✅/⚠️/❌

**O que TEM (Pontos Fortes):**
- [Ponto forte 1 com exemplo de código]
- [Ponto forte 2]

**O que FALTA (Pontos de Melhoria):**
- ❌ [Ausente: descrição do que não existe]
- ⚠️ [Incompleto: descrição do que precisa melhorar]
- 💡 [Sugestão: descrição da melhoria]
```

**IMPORTANTE:** 
1. Faça isso para TODOS os agregados encontrados em `/internal/domain/`
2. Liste também agregados que DEVERIAM existir mas NÃO existem
3. Para cada método, indique se está implementado, parcial ou vazio

---

#### 3.2. VALUE OBJECTS

Liste TODOS os Value Objects (implementados E ausentes):

**Value Objects IMPLEMENTADOS:**

| Value Object | Status | Localização | Validações | Métodos | Testes | Nota |
|--------------|--------|-------------|------------|---------|--------|------|
| Email | ✅ Completo | `/internal/domain/contact/value_objects.go` | Regex, lowercase | Equals(), String() | ✅ Sim | 9.5/10 |
| Phone | ✅ Completo | `/internal/domain/contact/value_objects.go` | Limpeza, tamanho | Equals(), String() | ⚠️ Parcial | 9.0/10 |
| [Outro VO] | ✅/⚠️/❌ | [caminho] | [validações] | [métodos] | ✅/⚠️/❌ | X/10 |

**Total de VOs Implementados:** [número]

---

**Value Objects AUSENTES (Oportunidades):**

Liste campos que DEVERIAM ser VOs mas são strings/primitivos:

| VO Sugerido | Domínio | Campo Atual | Tipo Atual | Justificativa | Prioridade |
|-------------|---------|-------------|------------|---------------|------------|
| MessageText | Message | text | *string | Validar tamanho máximo (4096 chars) | 🔴 Alta |
| MediaURL | Message | mediaURL | *string | Validar formato de URL | 🟡 Média |
| HexColor | Pipeline | color | string | Validar formato hexadecimal (#RRGGBB) | 🟡 Média |
| Money | Billing | amount | float64 | Valor + moeda (evitar erros) | 🔴 Alta |
| Timezone | Contact | timezone | *string | Validar timezone IANA | 🟢 Baixa |
| [Outro] | [domínio] | [campo] | [tipo] | [justificativa] | 🔴/🟡/🟢 |

**Total de VOs Ausentes:** [número]

---

#### 3.3. DOMAIN SERVICES

Liste TODOS os Domain Services encontrados (ou indique ausência):

| Domain Service | Localização | Responsabilidade | Nota |
|----------------|-------------|------------------|------|
| [Nome] | [caminho] | [descrição] | X/10 |

**Domain Services Ausentes (Oportunidades):**
- `PasswordPolicyService` - validar políticas de senha
- [Outros...]

---

#### 3.4. SPECIFICATIONS

Liste TODAS as Specifications encontradas (ou indique ausência):

| Specification | Localização | Uso | Nota |
|---------------|-------------|-----|------|
| [Nome] | [caminho] | [descrição] | X/10 |

**Status:** ✅ Implementado / ❌ Ausente

---

#### 3.5. FACTORIES

Liste TODAS as Factories encontradas:

| Factory | Localização | Tipo | Nota |
|---------|-------------|------|------|
| [Nome] | [caminho] | Explícita/Implícita (New*) | X/10 |

**Observação:** Se usar padrão `New*`, indique se é suficiente ou se precisa de Factory explícita.

---

#### 3.6. DOMAIN EVENTS

Liste TODOS os Domain Events encontrados:

| Event | Localização | Agregado | Payload | Nota |
|-------|-------------|----------|---------|------|
| ContactCreated | `/internal/domain/contact/events.go` | Contact | contactID, name, ... | 9/10 |
| [Outro] | [caminho] | [agregado] | [campos] | X/10 |

**Total de Domain Events:** [número]

---

#### 3.7. RESUMO DA CAMADA DE DOMÍNIO

| Elemento | Quantidade | Status |
|----------|------------|--------|
| Aggregate Roots | X | ✅/⚠️/❌ |
| Entities | X | ✅/⚠️/❌ |
| Value Objects | X | ✅/⚠️/❌ |
| Domain Events | X | ✅/⚠️/❌ |
| Repository Interfaces | X | ✅/⚠️/❌ |
| Domain Services | X | ✅/⚠️/❌ |
| Specifications | X | ✅/⚠️/❌ |
| Factories | X | ✅/⚠️/❌ |

**Nota Geral da Camada de Domínio: X/10**

---

### 4. CAMADA DE APLICAÇÃO - ANÁLISE DETALHADA

**Instruções:** Explore `/internal/application/` e liste TODOS os use cases e serviços.

#### 4.1. USE CASES / APPLICATION SERVICES

Agrupe por bounded context e liste TODOS os use cases:

**Formato:**
```markdown
#### Bounded Context: [Nome]

| Use Case | Arquivo | Tipo | Dependências | Nota |
|----------|---------|------|--------------|------|
| CreateContact | `create_contact.go` | Command | ContactRepository, EventBus | 9/10 |
| GetContact | `get_contact.go` | Query | ContactRepository | 8.5/10 |
| [Outro] | [arquivo] | Command/Query/Service | [deps] | X/10 |
```

Faça isso para TODOS os bounded contexts em `/internal/application/`.

---

#### 4.2. DTOs (Data Transfer Objects)

Liste TODOS os DTOs encontrados:

| DTO | Localização | Uso | Campos Principais | Nota |
|-----|-------------|-----|-------------------|------|
| ContactDTO | `/internal/application/dtos/contact_dtos.go` | Request/Response | id, name, email, phone | 9/10 |
| [Outro] | [caminho] | [uso] | [campos] | X/10 |

**Total de DTOs:** [número]

---

#### 4.3. PORTS (INTERFACES)

Liste TODAS as interfaces de portas (Hexagonal Architecture):

| Port | Localização | Métodos | Implementações | Nota |
|------|-------------|---------|----------------|------|
| MessageSender | `/internal/application/message/ports.go` | SendText(), SendMedia() | WAHAAdapter | 8.5/10 |
| [Outro] | [caminho] | [métodos] | [impls] | X/10 |

**Ports Ausentes (Oportunidades):**
- `EmailSender` - envio de emails
- `SMSSender` - envio de SMS
- [Outros...]

---

#### 4.4. CQRS (Command Query Responsibility Segregation)

Avalie a separação de Commands e Queries:

**Estrutura de Pastas:**
- `/internal/application/commands/` - ✅ Existe / ❌ Não existe / ⚠️ Vazio
- `/internal/application/queries/` - ✅ Existe / ❌ Não existe / ⚠️ Vazio

**Nomenclatura:**
- Commands terminam com `Command`? ✅/❌
- Queries terminam com `Query`? ✅/❌
- Handlers separados? ✅/❌

**Nota CQRS:** X/10

**Recomendação:** [Se não implementado, sugerir estrutura]

---

#### 4.5. EVENT HANDLERS / SUBSCRIBERS

Liste TODOS os event handlers/subscribers encontrados:

| Handler | Localização | Event Subscrito | Ação | Nota |
|---------|-------------|-----------------|------|------|
| [Nome] | [caminho] | [event] | [ação] | X/10 |

---

#### 4.6. RESUMO DA CAMADA DE APLICAÇÃO

| Elemento | Quantidade | Status |
|----------|------------|--------|
| Use Cases | X | ✅/⚠️/❌ |
| DTOs | X | ✅/⚠️/❌ |
| Ports | X | ✅/⚠️/❌ |
| Event Handlers | X | ✅/⚠️/❌ |
| CQRS Explícito | - | ✅/⚠️/❌ |

**Nota Geral da Camada de Aplicação: X/10**

---

### 5. CAMADA DE INFRAESTRUTURA - ANÁLISE DETALHADA

**Instruções:** Explore `/infrastructure/` e mapeie TODAS as implementações.

#### 5.1. REPOSITORIES (Implementações)

Liste TODOS os repositories implementados:

| Repository | Arquivo | Agregado | Métodos Principais | Nota |
|------------|---------|----------|-------------------|------|
| GormContactRepository | `/infrastructure/persistence/gorm_contact_repository.go` | Contact | Save(), FindByID(), FindByEmail() | 9.5/10 |
| [Outro] | [caminho] | [agregado] | [métodos] | X/10 |

**Total de Repositories:** [número]

**Avaliação:**
- Todos os agregados têm repository? ✅/❌
- Repositories implementam interface do domínio? ✅/❌
- Uso correto de transações? ✅/❌

---

#### 5.2. ENTIDADES GORM (Persistência)

Liste TODAS as entidades GORM encontradas em `/infrastructure/persistence/entities/`:

| Entidade GORM | Arquivo | Agregado Correspondente | Relacionamentos | Nota |
|---------------|---------|-------------------------|-----------------|------|
| Contact | `contact.go` | domain/contact.Contact | HasMany(Messages) | 9/10 |
| [Outra] | [arquivo] | [agregado] | [rels] | X/10 |

**Total de Entidades GORM:** [número]

**Mapeamento Domain ↔ Persistence:**
- Mappers explícitos (`domainToEntity`, `entityToDomain`)? ✅/❌
- Separação clara entre modelo de domínio e persistência? ✅/❌

---

#### 5.3. MIGRAÇÕES SQL

Analise TODAS as migrações em `/infrastructure/database/migrations/`:

**Total de Migrações:** [número]

**Qualidade das Migrações:**
- [ ] Foreign Keys bem definidas
- [ ] Índices otimizados
- [ ] Constraints (NOT NULL, UNIQUE, CHECK)
- [ ] Tipos de dados adequados
- [ ] Rollback (down migrations) implementado

**Destaques:**
Liste migrações importantes:
- `000XXX_create_contacts_table.up.sql` - [descrição]
- `000XXX_add_outbox_trigger.up.sql` - [descrição]
- [Outras...]

**Consistência:**
- Migrações sincronizadas com entidades GORM? ✅/❌
- Versionamento sequencial correto? ✅/❌

**Nota Migrações:** X/10

---

#### 5.4. EVENT BUS & OUTBOX PATTERN

Avalie a implementação de eventos:

**Outbox Pattern:**
- Tabela `outbox_events` existe? ✅/❌
- Trigger PostgreSQL `NOTIFY` implementado? ✅/❌
- Processor para publicar eventos? ✅/❌
- Localização: `/infrastructure/messaging/outbox_processor.go`

**Message Bus (RabbitMQ):**
- Conexão configurada? ✅/❌
- Filas declaradas? ✅/❌
- Exchanges configurados? ✅/❌

**Filas Declaradas:**
Liste TODAS as filas encontradas no código:
```
domain.events.contact.created
domain.events.contact.updated
domain.events.message.created
[Outras...]
```

**Total de Filas:** [número]

**Idempotência:**
- Tabela `processed_events` existe? ✅/❌
- Deduplicação de eventos implementada? ✅/❌

**Nota Event Bus:** X/10

---

#### 5.5. HTTP HANDLERS (Interface REST)

Liste TODOS os handlers em `/infrastructure/http/handlers/`:

| Handler | Arquivo | Endpoints | Métodos HTTP | Nota |
|---------|---------|-----------|--------------|------|
| ContactHandler | `contact_handler.go` | /contacts | GET, POST, PUT, DELETE | 8.5/10 |
| [Outro] | [arquivo] | [endpoints] | [métodos] | X/10 |

**Total de Handlers:** [número]

---

#### 5.6. MIDDLEWARE

Liste TODOS os middlewares em `/infrastructure/http/middleware/`:

| Middleware | Arquivo | Função | Nota |
|------------|---------|--------|------|
| AuthMiddleware | `auth.go` | Validação JWT | 9/10 |
| RBACMiddleware | `rbac.go` | Controle de acesso | 8.5/10 |
| RLSMiddleware | `rls.go` | Row-Level Security (tenant_id) | 9.5/10 |
| [Outro] | [arquivo] | [função] | X/10 |

---

#### 5.7. INTEGRAÇÕES EXTERNAS

Liste TODAS as integrações em `/infrastructure/channels/` ou `/infrastructure/`:

| Integração | Localização | ACL Implementado | Nota |
|------------|-------------|------------------|------|
| WAHA (WhatsApp) | `/infrastructure/channels/waha/` | ✅ Sim | 9/10 |
| [Outra] | [caminho] | ✅/❌ | X/10 |

**ACL (Anti-Corruption Layer):**
- Mapeamento de payloads externos para domínio? ✅/❌
- Tratamento de erros específicos? ✅/❌

---

#### 5.8. SEGURANÇA & CRIPTOGRAFIA

**Criptografia:**
- Implementação: `/infrastructure/crypto/aes_encryptor.go`
- Uso: Credenciais sensíveis (OAuth tokens, API keys)
- Algoritmo: AES-256-GCM
- Testes: ✅/❌

**Row-Level Security (RLS):**
- Middleware implementado? ✅/❌
- Filtro automático por `tenant_id`? ✅/❌
- Aplicado em todos os repositories? ✅/❌

**Nota Segurança:** X/10

---

#### 5.9. RESUMO DA CAMADA DE INFRAESTRUTURA

| Elemento | Quantidade | Status |
|----------|------------|--------|
| Repositories | X | ✅/⚠️/❌ |
| Entidades GORM | X | ✅/⚠️/❌ |
| Migrações SQL | X | ✅/⚠️/❌ |
| Handlers HTTP | X | ✅/⚠️/❌ |
| Middlewares | X | ✅/⚠️/❌ |
| Integrações | X | ✅/⚠️/❌ |
| Event Bus | - | ✅/⚠️/❌ |
| Outbox Pattern | - | ✅/⚠️/❌ |

**Nota Geral da Camada de Infraestrutura: X/10**

---

### 6. TIPOS, ENUMS E MÁQUINAS DE ESTADO

**Instruções:** Identifique TODOS os enums/types e máquinas de estado no código.

#### 6.1. ENUMS RICOS (Smart Enums)

Liste TODOS os enums/types encontrados:

**Formato:**
```markdown
#### Enum: [Nome]
**Localização:** `/internal/domain/[pasta]/types.go`

**Valores:**
```go
const (
    Valor1 TipoEnum = "valor1"
    Valor2 TipoEnum = "valor2"
    // ...
)
```

**Métodos:**
- `IsValid() bool` - ✅/❌
- `Parse(s string) (TipoEnum, error)` - ✅/❌
- [Outros métodos específicos]

**Nota:** X/10

**Sugestões:**
- [Se faltam métodos, sugerir]
```

Faça isso para TODOS os enums encontrados.

**Total de Enums:** [número]

---

#### 6.2. MÁQUINAS DE ESTADO

Identifique TODAS as máquinas de estado (implícitas ou explícitas):

**Formato:**
```markdown
#### Máquina de Estado: [Agregado].[Campo]

**Agregado:** [Nome]  
**Campo de Status:** [nome do campo]  
**Tipo:** [tipo do enum]

**Diagrama de Transições:**
```
┌─────────┐
│ estado1 │
└────┬────┘
     │
     ├──[ação1]──────> estado2
     │
     └──[ação2]──────> estado3
```

**Transições Válidas:**
- `estado1 -> estado2` (via ação1)
- `estado1 -> estado3` (via ação2)

**Transições Inválidas:**
- `estado2 -> estado1` ❌
- `estado3 -> estado1` ❌

**Implementação:**
- Método `CanTransitionTo(newStatus Status) bool` existe? ✅/❌
- Validação de transições no código? ✅/❌
- Tipo: Explícita / Implícita

**Nota:** X/10

**Sugestão:**
[Se implícita, sugerir criar método CanTransitionTo() e ValidTransitions()]
```

Faça isso para TODAS as máquinas de estado encontradas (ex: Session.Status, Message.Status, Pipeline.Active, etc.).

**Total de Máquinas de Estado:** [número]

---

### 7. ANÁLISE DE CONSISTÊNCIA

#### 7.1. NOMENCLATURA

Avalie a consistência de nomenclatura em TODO o código:

**Construtores:**
- Padrão usado: `New*` / `Create*` / Misto
- Consistente em todos os agregados? ✅/❌
- Exemplos: [liste alguns]

**Reconstrutores:**
- Padrão usado: `Reconstruct*` / `From*` / Outro
- Consistente? ✅/❌
- Exemplos: [liste alguns]

**Getters:**
- Seguem padrão Go (sem prefixo `Get`)? ✅/❌
- Exemplos: `ID()`, `Name()`, `Email()`

**Métodos de Negócio:**
- Usam verbos claros? ✅/❌
- Exemplos: `UpdateName()`, `MarkAsRead()`, `Close()`

**Nota Nomenclatura:** X/10

---

#### 7.2. PADRÕES ARQUITETURAIS

Verifique a implementação dos padrões:

| Padrão | Implementado | Qualidade | Observações |
|--------|--------------|-----------|-------------|
| **Repository Pattern** | ✅/❌ | X/10 | Interface no domínio, impl na infra |
| **Dependency Inversion** | ✅/❌ | X/10 | Portas e adaptadores |
| **Domain Events** | ✅/❌ | X/10 | Todos os agregados emitem? |
| **Encapsulamento** | ✅/❌ | X/10 | Campos privados + getters |
| **Invariantes** | ✅/❌ | X/10 | Validadas nos construtores |
| **Outbox Pattern** | ✅/❌ | X/10 | Consistência eventual |
| **CQRS** | ✅/❌ | X/10 | Separação explícita |
| **ACL** | ✅/❌ | X/10 | Anti-Corruption Layer |

---

#### 7.3. ESTRUTURA DE PASTAS

Avalie a organização do projeto:

```
/home/caloi/ventros-crm/
├── internal/
│   ├── domain/              ✅/⚠️/❌ [nota]
│   │   ├── [bounded_context]/
│   │   └── shared/
│   └── application/         ✅/⚠️/❌ [nota]
│       ├── [bounded_context]/
│       ├── commands/        ✅/⚠️/❌ [vazio?]
│       ├── queries/         ✅/⚠️/❌ [vazio?]
│       └── dtos/
├── infrastructure/          ✅/⚠️/❌ [nota]
│   ├── persistence/
│   ├── http/
│   ├── messaging/
│   └── channels/
└── cmd/                     ✅/⚠️/❌ [nota]
```

**Problemas Identificados:**
- [Liste pastas vazias ou mal organizadas]
- [Liste arquivos fora do lugar]

**Nota Estrutura:** X/10

---

### 8. OPORTUNIDADES DE MELHORIA

Liste sugestões PRIORIZADAS e ESPECÍFICAS:

#### 🔴 PRIORIDADE ALTA (Impacto Crítico)

**1. [Título da Melhoria]**
- **Problema Atual:** [Descreva o problema encontrado no código]
- **Localização:** [Arquivo/pasta específica]
- **Impacto:** [Consequências do problema]
- **Solução Sugerida:** [Como resolver]
- **Exemplo de Código:**
```go
// Antes (problemático)
[código atual]

// Depois (sugerido)
[código melhorado]
```

**2. [Outra Melhoria]**
[Mesmo formato...]

---

#### 🟡 PRIORIDADE MÉDIA (Melhoria Significativa)

**1. [Título]**
- **Problema Atual:** ...
- **Localização:** ...
- **Impacto:** ...
- **Solução Sugerida:** ...

[Continue para todas as melhorias de prioridade média]

---

#### 🟢 PRIORIDADE BAIXA (Refinamento)

**1. [Título]**
- **Problema Atual:** ...
- **Localização:** ...
- **Impacto:** ...
- **Solução Sugerida:** ...

[Continue para todas as melhorias de prioridade baixa]

---

### 9. RESUMO EXECUTIVO FINAL

#### 9.1. TABELA DE NOTAS POR CATEGORIA

| Categoria | Nota | Status | Justificativa |
|-----------|------|--------|---------------|
| **Agregados & Entidades** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Value Objects** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Domain Events** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Repositories** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Use Cases** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **DTOs** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Handlers** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Migrações** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Event Bus** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Segurança** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Testes** | X/10 | ✅/⚠️/❌ | [breve justificativa] |
| **Documentação** | X/10 | ✅/⚠️/❌ | [breve justificativa] |

### **NOTA GERAL: X.X/10**

---

#### 9.2. PONTOS FORTES (TOP 5)

1. **[Ponto Forte 1]**
   - Descrição detalhada
   - Exemplo no código
   - Por que é excelente

2. **[Ponto Forte 2]**
   [Mesmo formato...]

[Continue até 5 pontos fortes]

---

#### 9.3. PONTOS CRÍTICOS (TOP 5)

1. **[Ponto Crítico 1]**
   - Descrição do problema
   - Impacto no sistema
   - Urgência de correção

2. **[Ponto Crítico 2]**
   [Mesmo formato...]

[Continue até 5 pontos críticos]

---

#### 9.4. CONCLUSÃO

**Estado Atual da Arquitetura:**
[Parágrafo descrevendo o estado geral do sistema]

**Conformidade com DDD:**
[Avaliação de quanto o sistema segue os princípios de DDD (Eric Evans, Vaughn Vernon)]

**Conformidade com Clean Architecture:**
[Avaliação da separação de camadas e dependências]

**Recomendação Final:**
- [ ] ✅ **Pronto para produção** - Sistema maduro, poucas melhorias necessárias
- [ ] ⚠️ **Pronto com ressalvas** - Funcional, mas precisa de melhorias incrementais
- [ ] ❌ **Precisa refatoração** - Problemas arquiteturais críticos

**Próximos Passos Sugeridos:**
1. [Ação prioritária 1]
2. [Ação prioritária 2]
3. [Ação prioritária 3]

---

## 📚 REFERÊNCIAS UTILIZADAS

- **Domain-Driven Design** (Eric Evans, 2003)
- **Implementing Domain-Driven Design** (Vaughn Vernon, 2013)
- **Clean Architecture** (Robert C. Martin, 2017)
- **Patterns of Enterprise Application Architecture** (Martin Fowler, 2002)

---

## 🔍 INSTRUÇÕES CRÍTICAS PARA A IA

### IMPORTANTE: LEIA ANTES DE COMEÇAR

1. **Você TEM acesso ao código:** Explore `/home/caloi/ventros-crm/` completamente
2. **Não assuma nada:** Leia TODOS os arquivos relevantes antes de avaliar
3. **Seja específico:** Cite arquivos, linhas, nomes de classes/métodos
4. **Seja completo:** Não pule bounded contexts, agregados ou elementos
5. **Seja rigoroso:** Avalie com base em DDD puro (Eric Evans, Vaughn Vernon)
6. **Seja construtivo:** Toda crítica DEVE vir com sugestão de melhoria
7. **Use exemplos:** Mostre código real do projeto (bom e problemático)
8. **Priorize:** Foque no que realmente importa para a arquitetura

### FORMATO DE SAÍDA

- **Arquivo:** Sobrescreva `/home/caloi/ventros-crm/TASKS-2.md` com sua análise
- **Formato:** Markdown completo e navegável
- **Tamanho:** 1500-2500 linhas (análise detalhada)
- **Elementos:** Índice clicável, emojis, tabelas, blocos de código, diagramas ASCII

### ESTRUTURA OBRIGATÓRIA

Siga EXATAMENTE a estrutura definida acima:
1. Sumário Executivo
2. Bounded Contexts Identificados
3. Camada de Domínio (3.1 a 3.7)
4. Camada de Aplicação (4.1 a 4.6)
5. Camada de Infraestrutura (5.1 a 5.9)
6. Tipos, Enums e Máquinas de Estado (6.1 a 6.2)
7. Análise de Consistência (7.1 a 7.3)
8. Oportunidades de Melhoria (Prioridade Alta/Média/Baixa)
9. Resumo Executivo Final (9.1 a 9.4)

---

## 📋 CHECKLIST DE COMPLETUDE

Antes de finalizar, verifique se sua análise inclui:

### Camada de Domínio
- [ ] Todos os bounded contexts identificados (espera-se 6-10)
- [ ] Todos os agregados mapeados com detalhes completos (espera-se 15+)
- [ ] Todos os value objects listados (espera-se 6+)
- [ ] Todos os domain events documentados (espera-se 45+)
- [ ] Todos os repository interfaces listados (espera-se 15+)
- [ ] Domain services identificados (ou ausência indicada)
- [ ] Specifications identificadas (ou ausência indicada)
- [ ] Factories identificadas (explícitas ou implícitas via New*)

### Camada de Aplicação
- [ ] Todos os use cases mapeados por bounded context (espera-se 30+)
- [ ] Todos os DTOs listados (espera-se 5+)
- [ ] Todos os ports/interfaces identificados
- [ ] CQRS avaliado (estrutura de pastas, nomenclatura)
- [ ] Event handlers/subscribers listados

### Camada de Infraestrutura
- [ ] Todos os repositories implementados (espera-se 17+)
- [ ] Todas as entidades GORM mapeadas (espera-se 26+)
- [ ] Todas as migrações SQL analisadas (espera-se 26+)
- [ ] Event Bus & Outbox Pattern avaliados
- [ ] Filas RabbitMQ listadas (espera-se 45+)
- [ ] Handlers HTTP listados (espera-se 18+)
- [ ] Middlewares listados (espera-se 4+)
- [ ] Integrações externas avaliadas
- [ ] Segurança & criptografia analisadas

### Tipos & Estado
- [ ] Todos os enums documentados com métodos (espera-se 10+)
- [ ] Todas as máquinas de estado identificadas (espera-se 3+)
- [ ] Diagramas ASCII para transições de estado

### Análise & Recomendações
- [ ] Nomenclatura avaliada (construtores, getters, métodos)
- [ ] Padrões arquiteturais verificados (8 padrões listados)
- [ ] Estrutura de pastas avaliada
- [ ] Oportunidades de melhoria priorizadas (Alta/Média/Baixa)
- [ ] Cada melhoria com: problema, localização, impacto, solução, exemplo

### Resumo Final
- [ ] Tabela de notas por categoria (12 categorias)
- [ ] Nota geral calculada
- [ ] Top 5 pontos fortes com justificativas
- [ ] Top 5 pontos críticos com impacto
- [ ] Conclusão com recomendação (pronto/ressalvas/refatoração)
- [ ] Próximos passos sugeridos

---

## ⚠️ AVISOS FINAIS

### REGRAS CRÍTICAS:

1. **✅ LISTE TUDO QUE EXISTE** - Mesmo que:
   - Não tenha testes
   - Seja apenas estrutura
   - Tenha TODOs
   - Esteja incompleto
   - Não tenha lógica

2. **❌ LISTE TUDO QUE FALTA** - Indique:
   - Agregados ausentes
   - VOs que deveriam existir
   - Métodos vazios
   - Testes ausentes
   - Validações faltantes

3. **⚠️ INDIQUE O QUE ESTÁ PARCIAL** - Mostre:
   - Código com TODOs
   - Métodos sem validação
   - Testes incompletos
   - Documentação parcial

4. **NÃO invente dados:** Se não encontrar algo, indique ausência explicitamente

5. **NÃO seja genérico:** Cite arquivos, linhas de código, nomes de métodos

6. **NÃO pule seções:** Mesmo que vazia, indique "❌ Não encontrado"

7. **SIM, seja crítico:** Avalie com rigor, mas seja construtivo

8. **SIM, dê exemplos:** Mostre código real do projeto (bom E problemático)

9. **SIM, use % de completude:** Estime implementação (ex: "85% implementado")

10. **SIM, priorize melhorias:** Use 🔴 Alta / 🟡 Média / 🟢 Baixa

---

## 🎯 COMEÇE AGORA

**LEMBRE-SE:** Você deve trazer **TUDO** - o que existe, o que falta, o que está incompleto.

Explore o código em `/home/caloi/ventros-crm/` e gere a análise completa sobrescrevendo este arquivo.

**BOA SORTE!** 🚀
```
