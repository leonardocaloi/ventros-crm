# 🏛️ DOCUMENTAÇÃO - CAMADA DE DOMÍNIO

**Localização**: `internal/domain/`
**Propósito**: Contém a lógica de negócio pura, independente de frameworks e infraestrutura

---

## 📋 ÍNDICE

1. [Aggregate Roots](#aggregate-roots)
2. [Value Objects](#value-objects)
3. [Domain Events](#domain-events)
4. [Domain Services](#domain-services)
5. [Repositories (Interfaces)](#repositories)
6. [Regras de Negócio](#regras-de-negócio)

---

## 1. AGGREGATE ROOTS

### 1.1 Contact (Contato)

**Arquivo**: `internal/domain/contact/contact.go`

**Responsabilidade**: Gerenciar informações e ciclo de vida de contatos (clientes/leads)

**Invariantes**:
- ✅ Nome não pode ser vazio
- ✅ Email deve ser válido (se fornecido)
- ✅ Phone deve ser normalizado
- ✅ Soft delete (deletedAt != nil)
- ✅ Tags não podem ter duplicatas

**Métodos Principais**:

```go
// Factory Method
NewContact(projectID, tenantID, name) (*Contact, error)

// Comandos
contact.SetEmail(email string) error
contact.SetPhone(phone string) error
contact.UpdateName(name string) error
contact.AddTag(tag string)
contact.RemoveTag(tag string)
contact.SetProfilePicture(url string)
contact.RecordInteraction()
contact.SoftDelete() error

// Queries
contact.ID() uuid.UUID
contact.Name() string
contact.Email() *Email
contact.Phone() *Phone
contact.IsDeleted() bool
contact.Tags() []string
```

**Domain Events Publicados**:
- `ContactCreatedEvent`: Quando contato é criado
- `ContactUpdatedEvent`: Quando informações mudam
- `ContactDeletedEvent`: Quando soft delete ocorre
- `ContactProfilePictureUpdatedEvent`: Quando foto de perfil é atualizada
- `ContactPipelineStatusChangedEvent`: Quando status no pipeline muda

**Regras de Negócio**:

1. **RN-001**: Contato DEVE ter nome não vazio
2. **RN-002**: Email DEVE ser válido (formato xxx@yyy.zzz)
3. **RN-003**: Phone DEVE ser normalizado (+55XXXXXXXXXXX)
4. **RN-004**: Tags não podem ter duplicatas
5. **RN-005**: Soft delete mantém dados para auditoria
6. **RN-006**: ProfilePicture sincronizada do WhatsApp
7. **RN-007**: FirstInteractionAt nunca pode ser alterado após set
8. **RN-008**: LastInteractionAt atualiza em toda interação

---

### 1.2 Session (Sessão de Atendimento)

**Arquivo**: `internal/domain/session/session.go`

**Responsabilidade**: Gerenciar conversas/atendimentos com timeout automático

**Invariantes**:
- ✅ ContactID obrigatório
- ✅ TenantID obrigatório
- ✅ TimeoutDuration > 0
- ✅ Status válido (Active, Ended)
- ✅ EndReason obrigatório quando Status = Ended

**Métodos Principais**:

```go
// Factory Methods
NewSession(contactID, tenantID, channelTypeID, timeoutDuration) (*Session, error)
NewSessionWithPipeline(contactID, tenantID, channelTypeID, pipelineID, timeout) (*Session, error)

// Comandos
session.RecordMessage(fromContact bool, timestamp) error
session.AssignAgent(agentID) error
session.End(reason EndReason) error
session.Resolve() error
session.Escalate() error
session.SetSummary(summary, sentiment, score, topics, nextSteps)

// Queries
session.ID() uuid.UUID
session.ContactID() uuid.UUID
session.IsActive() bool
session.MessageCount() int
session.AgentResponseTimeSeconds() *int
session.ShouldGenerateSummary() bool
```

**Domain Events Publicados**:
- `SessionStartedEvent`: Nova sessão criada
- `SessionEndedEvent`: Sessão encerrada (manual ou timeout)
- `MessageRecordedEvent`: Mensagem adicionada à sessão
- `AgentAssignedEvent`: Agente atribuído
- `SessionResolvedEvent`: Sessão marcada como resolvida
- `SessionEscalatedEvent`: Sessão escalada
- `SessionSummarizedEvent`: IA gerou resumo

**Regras de Negócio**:

1. **RN-S01**: Timeout vem do Pipeline.SessionTimeoutMinutes (default: 30min)
2. **RN-S02**: LastActivityAt atualiza em TODA mensagem (reset timeout)
3. **RN-S03**: AgentResponseTimeSeconds = tempo até primeira resposta do agente
4. **RN-S04**: ContactWaitTimeSeconds = tempo até primeira resposta do contato
5. **RN-S05**: AgentTransfers incrementa quando troca de agente
6. **RN-S06**: Sessão SOMENTE pode ser resolvida após EndedAt != nil
7. **RN-S07**: Sessão ativa NÃO pode ser deletada
8. **RN-S08**: Summary gerado SOMENTE se messageCount >= 3

---

### 1.3 Message (Mensagem)

**Arquivo**: `internal/domain/message/message.go`

**Responsabilidade**: Representar mensagens individuais (text, media, etc)

**Invariantes**:
- ✅ ContactID obrigatório
- ✅ ProjectID obrigatório
- ✅ CustomerID obrigatório
- ✅ ContentType válido (Text, Image, Audio, Video, Document, etc)
- ✅ ChannelID obrigatório (toda mensagem vem de um canal)

**Métodos Principais**:

```go
// Factory Method
NewMessage(contactID, projectID, customerID, contentType, fromMe) (*Message, error)

// Comandos
message.SetText(text string) error
message.SetMediaContent(url, mimetype string) error
message.AssignToChannel(channelID, channelTypeID)
message.AssignToSession(sessionID)
message.SetChannelMessageID(channelMessageID) // Para dedup
message.MarkAsDelivered()
message.MarkAsRead()
message.MarkAsFailed()

// Queries
message.ID() uuid.UUID
message.IsInbound() bool // Do contato
message.IsOutbound() bool // Para o contato
message.HasMediaURL() bool
message.Status() Status
```

**Domain Events Publicados**:
- `MessageCreatedEvent`: Nova mensagem
- `MessageDeliveredEvent`: ACK de entrega
- `MessageReadEvent`: Lido pelo destinatário

**Regras de Negócio**:

1. **RN-M01**: ContentType.IsText() → DEVE ter text != nil
2. **RN-M02**: ContentType.IsMedia() → DEVE ter mediaURL != nil
3. **RN-M03**: ChannelMessageID usado para deduplicação
4. **RN-M04**: FromMe = false → Inbound, FromMe = true → Outbound
5. **RN-M05**: DeliveredAt SEMPRE < ReadAt (se ambos existem)
6. **RN-M06**: Status transitions: Sent → Delivered → Read
7. **RN-M07**: Failed messages NÃO transitam para Delivered

---

### 1.4 Agent (Agente)

**Arquivo**: `internal/domain/agent/agent.go`

**Responsabilidade**: Representar agentes (humanos, IA, bots) que interagem com contatos

**Invariantes**:
- ✅ ProjectID obrigatório
- ✅ TenantID obrigatório
- ✅ Name não vazio
- ✅ AgentType válido (Human, AI, Bot, Channel)
- ✅ Human agents DEVEM ter UserID
- ✅ AI agents DEVEM ter config (provider, model)

**Métodos Principais**:

```go
// Factory Method
NewAgent(projectID, tenantID, name, agentType, userID) (*Agent, error)

// Comandos
agent.UpdateProfile(name, email) error
agent.Activate() error
agent.Deactivate() error
agent.RecordLogin()
agent.GrantPermission(permission) error
agent.RevokePermission(permission) error
agent.SetStatus(status)
agent.SetConfig(config) // Para AI agents
agent.RecordSessionHandled(responseTimeMs)

// Queries
agent.ID() uuid.UUID
agent.Type() AgentType
agent.IsActive() bool
agent.HasPermission(permission) bool
agent.SessionsHandled() int
agent.AverageResponseMs() int
```

**Domain Events Publicados**:
- `AgentCreatedEvent`
- `AgentUpdatedEvent`
- `AgentActivatedEvent`
- `AgentDeactivatedEvent`
- `AgentLoggedInEvent`
- `AgentPermissionGrantedEvent`
- `AgentPermissionRevokedEvent`

**Regras de Negócio**:

1. **RN-A01**: AgentType = Human → userID obrigatório
2. **RN-A02**: AgentType = AI → config.provider e config.model obrigatórios
3. **RN-A03**: AgentType = Bot → config.workflow obrigatório
4. **RN-A04**: AgentType = Channel → representa canal (não usuário)
5. **RN-A05**: AverageResponseMs = média móvel (não acumula histórico todo)
6. **RN-A06**: LastActivityAt atualiza em RecordSessionHandled()
7. **RN-A07**: Status = Available, Busy, Away, Offline

---

### 1.5 Pipeline (Funil de Vendas)

**Arquivo**: `internal/domain/pipeline/pipeline.go`

**Responsabilidade**: Gerenciar fluxos de contatos com múltiplos status/estágios

**Invariantes**:
- ✅ Name não vazio
- ✅ ProjectID obrigatório
- ✅ Statuses não podem estar vazios
- ✅ Status order DEVE ser único
- ✅ SessionTimeoutMinutes > 0

**Métodos Principais**:

```go
// Factory Method
NewPipeline(projectID, tenantID, name) (*Pipeline, error)

// Comandos
pipeline.AddStatus(name, order, color) error
pipeline.UpdateStatusOrder(statusID, newOrder) error
pipeline.RemoveStatus(statusID) error
pipeline.SetSessionTimeout(minutes int)
pipeline.Activate()
pipeline.Deactivate()

// Queries
pipeline.ID() uuid.UUID
pipeline.Name() string
pipeline.Statuses() []*PipelineStatus
pipeline.IsActive() bool
pipeline.GetStatus(statusID) *PipelineStatus
```

**Domain Events Publicados**:
- `PipelineCreatedEvent`
- `PipelineStatusAddedEvent`
- `PipelineStatusUpdatedEvent`
- `ContactEnteredPipelineEvent`
- `ContactExitedPipelineEvent`
- `ContactStatusChangedEvent`

**Regras de Negócio**:

1. **RN-P01**: SessionTimeoutMinutes define timeout de Session
2. **RN-P02**: Status.Order DEVE ser único dentro do pipeline
3. **RN-P03**: Status NÃO pode ser removido se tiver contatos
4. **RN-P04**: Pipeline inativo NÃO aceita novos contatos
5. **RN-P05**: Mudança de status SEMPRE gera ContactStatusChangedEvent

---

### 1.6 Tracking (Rastreamento UTM)

**Arquivo**: `internal/domain/tracking/tracking.go`

**Responsabilidade**: Rastrear origem de contatos (ads, campanhas, landing pages)

**Invariantes**:
- ✅ ContactID obrigatório
- ✅ SessionID obrigatório
- ✅ MessageID obrigatório
- ✅ Source não vazio
- ✅ UTM parameters validados

**Métodos Principais**:

```go
// Builder Pattern
builder := NewTrackingBuilder(contactID, sessionID, messageID)
builder.WithUTMSource("google")
builder.WithUTMCampaign("summer-2024")
builder.WithGCLID("abc123")
builder.Build() (*Tracking, error)

// Comandos
tracking.EnrichWithFBCLID(fbclid)
tracking.EnrichWithGCLID(gclid)
tracking.SetConversionValue(value)

// Queries
tracking.ID() uuid.UUID
tracking.GetUTMSource() string
tracking.IsMetaAd() bool
tracking.IsGoogleAd() bool
```

**Domain Events Publicados**:
- `TrackingCreatedEvent`
- `TrackingEnrichedEvent`
- `AdConversionTrackedEvent`

**Regras de Negócio**:

1. **RN-T01**: utm_source = "meta" → tracking de Meta Ads
2. **RN-T02**: utm_source = "google" → tracking de Google Ads
3. **RN-T03**: fbclid detectado → enrichment automático Meta
4. **RN-T04**: gclid detectado → enrichment automático Google
5. **RN-T05**: Conversion value SOMENTE para ads pagos

---

## 2. VALUE OBJECTS

### 2.1 Email

**Arquivo**: `internal/domain/contact/value_objects.go`

```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    // Validação de formato
    // Normalização (lowercase)
}
```

**Invariantes**:
- ✅ Formato válido (regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
- ✅ Case insensitive (sempre lowercase)
- ✅ Immutable

### 2.2 Phone

**Arquivo**: `internal/domain/contact/value_objects.go`

```go
type Phone struct {
    value string // Formato: +55XXXXXXXXXXX
}

func NewPhone(phone string) (Phone, error) {
    // Normalização (+55)
    // Remoção de caracteres especiais
}
```

**Invariantes**:
- ✅ Formato internacional (+DDI)
- ✅ Apenas dígitos (remove espaços, parênteses, hífens)
- ✅ Immutable

### 2.3 TenantID

**Arquivo**: `internal/domain/shared/tenant_id.go`

```go
type TenantID struct {
    value string
}
```

**Invariantes**:
- ✅ Não vazio
- ✅ Type-safe (não é string comum)
- ✅ Usado em RLS (Row Level Security)

### 2.4 CustomField

**Arquivo**: `internal/domain/shared/custom_field.go`

```go
type CustomField struct {
    Key   string
    Value interface{}
    Type  CustomFieldType // String, Number, Boolean, Date
}
```

**Invariantes**:
- ✅ Key não vazio
- ✅ Type válido
- ✅ Value compatível com Type

---

## 3. DOMAIN EVENTS

### 3.1 BaseEvent (Padrão)

**Arquivo**: `internal/domain/shared/domain_event.go`

```go
type DomainEvent interface {
    EventName() string       // Ex: "contact.created"
    EventID() uuid.UUID      // ID único (idempotência)
    EventVersion() string    // Ex: "v1" (schema evolution)
    OccurredAt() time.Time   // Timestamp
}

type BaseEvent struct {
    eventID      uuid.UUID
    eventName    string
    eventVersion string
    occurredAt   time.Time
}
```

**Todos os eventos DEVEM**:
- ✅ Embedar `BaseEvent`
- ✅ Ter construtor `NewXxxEvent()`
- ✅ Ser imutáveis
- ✅ Carregar estado necessário (Event Carried State Transfer)

### 3.2 Eventos por Agregado

**Contact**:
- `ContactCreatedEvent`
- `ContactUpdatedEvent`
- `ContactDeletedEvent`
- `ContactProfilePictureUpdatedEvent`
- `ContactPipelineStatusChangedEvent`
- `AdConversionTrackedEvent`

**Session**:
- `SessionStartedEvent`
- `SessionEndedEvent`
- `MessageRecordedEvent`
- `AgentAssignedEvent`
- `SessionResolvedEvent`
- `SessionEscalatedEvent`
- `SessionSummarizedEvent`

**Message**:
- `MessageCreatedEvent`
- `MessageDeliveredEvent`
- `MessageReadEvent`

**Agent**:
- `AgentCreatedEvent`
- `AgentUpdatedEvent`
- `AgentActivatedEvent`
- `AgentDeactivatedEvent`
- `AgentLoggedInEvent`
- `AgentPermissionGrantedEvent`
- `AgentPermissionRevokedEvent`

**Pipeline**:
- `PipelineCreatedEvent`
- `PipelineStatusAddedEvent`
- `ContactEnteredPipelineEvent`
- `ContactExitedPipelineEvent`
- `ContactStatusChangedEvent`

**Tracking**:
- `TrackingCreatedEvent`
- `TrackingEnrichedEvent`

**Note**:
- `NoteAddedEvent`
- `NoteUpdatedEvent`
- `NoteDeletedEvent`
- `NotePinnedEvent`

---

## 4. DOMAIN SERVICES

### 4.1 Quando Usar Domain Services

**Critério**: Lógica de negócio que **NÃO** pertence a um único agregado

**Exemplos**:
- ❌ `Contact.UpdateName()` → Pertence ao Contact (método do agregado)
- ✅ `TransferContactBetweenProjects(contactID, fromProject, toProject)` → Domain Service (afeta múltiplos agregados)

**Implementados**:

```go
// internal/domain/messaging/message_sender.go
type MessageSender interface {
    SendMessage(ctx, to, text) error
    SendMedia(ctx, to, mediaType, url) error
}
```

---

## 5. REPOSITORIES (INTERFACES)

**Padrão**: Definidas no domínio, implementadas na infra

```go
// internal/domain/contact/repository.go
type Repository interface {
    Save(ctx context.Context, contact *Contact) error
    FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)
    FindByPhone(ctx context.Context, phone Phone) (*Contact, error)
    FindByEmail(ctx context.Context, email Email) (*Contact, error)
    List(ctx context.Context, filters ContactFilters) ([]*Contact, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

**Repositórios Definidos**:
- `contact.Repository`
- `session.Repository`
- `message.Repository`
- `agent.Repository`
- `pipeline.Repository`
- `tracking.Repository`
- `note.Repository`
- `webhook.Repository`
- `outbox.Repository`

---

## 6. REGRAS DE NEGÓCIO (RESUMO)

### 6.1 Contact (Contato)

| ID | Regra |
|----|-------|
| RN-001 | Nome obrigatório e não vazio |
| RN-002 | Email deve ser válido (se fornecido) |
| RN-003 | Phone normalizado (+55XXXXXXXXXXX) |
| RN-004 | Tags sem duplicatas |
| RN-005 | Soft delete mantém dados |
| RN-006 | ProfilePicture sincronizada do WhatsApp |
| RN-007 | FirstInteractionAt imutável |
| RN-008 | LastInteractionAt atualiza em toda interação |

### 6.2 Session (Sessão)

| ID | Regra |
|----|-------|
| RN-S01 | Timeout vem do Pipeline (default: 30min) |
| RN-S02 | LastActivityAt reset em toda mensagem |
| RN-S03 | AgentResponseTime = tempo até 1ª resposta |
| RN-S04 | ContactWaitTime = tempo até 1ª resposta do contato |
| RN-S05 | AgentTransfers incrementa em troca |
| RN-S06 | Resolve SOMENTE após EndedAt != nil |
| RN-S07 | Sessão ativa NÃO pode ser deletada |
| RN-S08 | Summary gerado se messageCount >= 3 |

### 6.3 Message (Mensagem)

| ID | Regra |
|----|-------|
| RN-M01 | Text message → text obrigatório |
| RN-M02 | Media message → mediaURL obrigatório |
| RN-M03 | ChannelMessageID para dedup |
| RN-M04 | FromMe define direção |
| RN-M05 | DeliveredAt < ReadAt |
| RN-M06 | Status: Sent → Delivered → Read |
| RN-M07 | Failed não transita para Delivered |

### 6.4 Agent (Agente)

| ID | Regra |
|----|-------|
| RN-A01 | Human → UserID obrigatório |
| RN-A02 | AI → config (provider, model) obrigatório |
| RN-A03 | Bot → config.workflow obrigatório |
| RN-A04 | Channel → representa canal, não usuário |
| RN-A05 | AverageResponseMs = média móvel |
| RN-A06 | LastActivityAt atualiza em handling |
| RN-A07 | Status: Available, Busy, Away, Offline |

### 6.5 Pipeline (Funil)

| ID | Regra |
|----|-------|
| RN-P01 | SessionTimeoutMinutes define timeout de Session |
| RN-P02 | Status.Order único no pipeline |
| RN-P03 | Status com contatos NÃO pode ser removido |
| RN-P04 | Pipeline inativo NÃO aceita novos contatos |
| RN-P05 | Mudança status SEMPRE gera evento |

### 6.6 Tracking (Rastreamento)

| ID | Regra |
|----|-------|
| RN-T01 | utm_source = "meta" → Meta Ads |
| RN-T02 | utm_source = "google" → Google Ads |
| RN-T03 | fbclid → enrichment Meta |
| RN-T04 | gclid → enrichment Google |
| RN-T05 | Conversion value SOMENTE ads pagos |

---

## 📖 BOUNDED CONTEXTS

```
┌─────────────────────────────────────────────────────────┐
│  CONTACT MANAGEMENT                                     │
│  - Contact (AR)                                         │
│  - ContactEvent                                         │
│  - ContactList                                          │
│  - CustomFields                                         │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  SESSION MANAGEMENT                                     │
│  - Session (AR)                                         │
│  - Message                                              │
│  - AgentSession                                         │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  AGENT MANAGEMENT                                       │
│  - Agent (AR)                                           │
│  - Permissions                                          │
│  - AI Provider Integration                              │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  PIPELINE MANAGEMENT                                    │
│  - Pipeline (AR)                                        │
│  - PipelineStatus                                       │
│  - Status Transitions                                   │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  CHANNEL MANAGEMENT                                     │
│  - Channel (AR)                                         │
│  - ChannelType                                          │
│  - WAHA Integration                                     │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  TRACKING & ATTRIBUTION                                 │
│  - Tracking (AR)                                        │
│  - UTM Parameters                                       │
│  - Ad Conversion                                        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  BILLING                                                │
│  - BillingAccount (AR)                                  │
│  - Customer                                             │
│  - Project                                              │
└─────────────────────────────────────────────────────────┘
```

---

## ✅ CHECKLIST DE QUALIDADE

### Aggregate Roots
- [x] Factory methods bem definidos
- [x] Reconstruct methods para hidratação
- [x] Invariantes protegidas
- [x] Comandos retornam errors
- [x] Queries nunca modificam estado
- [x] Domain events publicados corretamente
- [x] Getters imutáveis (retornam cópias)

### Value Objects
- [x] Imutáveis
- [x] Validação no construtor
- [x] Equality baseado em valor
- [x] Sem ID próprio

### Domain Events
- [x] Implementam DomainEvent
- [x] Têm EventID (idempotência)
- [x] Têm EventVersion (schema evolution)
- [x] Construtores padronizados
- [x] Carregam estado necessário

### Repositories
- [x] Interfaces no domínio
- [x] Implementações na infra
- [x] Retornam agregados completos
- [x] Save/FindByID/Delete padrão

---

**Próximo**: Ver [Camada de Aplicação](./CAMADA_APLICACAO.md)
