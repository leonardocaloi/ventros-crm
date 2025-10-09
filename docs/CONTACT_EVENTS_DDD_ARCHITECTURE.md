# Contact Events - Arquitetura DDD

## Visão Geral

Sistema de eventos de contato seguindo **Domain-Driven Design (DDD)** para streaming via SSE.

## Arquitetura em Camadas (DDD)

```
┌─────────────────────────────────────────────────────────────────┐
│                      PRESENTATION LAYER                          │
│                    (Interface/Handlers)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
    ▼                         ▼                         ▼
┌──────────┐          ┌──────────────┐         ┌──────────────┐
│   SSE    │          │   REST API   │         │   GraphQL    │
│ Handler  │          │   Endpoints  │         │  (futuro)    │
└──────────┘          └──────────────┘         └──────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     APPLICATION LAYER                            │
│                    (Use Cases/Services)                          │
└─────────────────────────────────────────────────────────────────┘
                              │
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
    ▼                         ▼                         ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ CreateContact    │  │ ContactEvent     │  │ DomainEvent      │
│ EventUseCase     │  │ Consumer         │  │ Handlers         │
└──────────────────┘  └──────────────────┘  └──────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       DOMAIN LAYER                               │
│                  (Entities/Aggregates/Events)                    │
└─────────────────────────────────────────────────────────────────┘
                              │
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
    ▼                         ▼                         ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ ContactEvent     │  │ Domain Events    │  │ Value Objects    │
│ (Aggregate)      │  │ (Contact, etc)   │  │ (Category, etc)  │
└──────────────────┘  └──────────────────┘  └──────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   INFRASTRUCTURE LAYER                           │
│              (Persistence/Messaging/External)                    │
└─────────────────────────────────────────────────────────────────┘
                              │
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
    ▼                         ▼                         ▼
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ PostgreSQL       │  │ RabbitMQ         │  │ Redis Cache      │
│ (contact_events) │  │ (Domain Events)  │  │ (opcional)       │
└──────────────────┘  └──────────────────┘  └──────────────────┘
```

## Fluxo Completo (Event-Driven Architecture)

### 1. Ação de Domínio Acontece

```go
// Domain Layer - Contact Aggregate
contact.ChangeStatus(newStatusID)

// Aggregate levanta evento de domínio
contact.RaiseEvent(ContactStatusChangedEvent{
    ContactID: contact.ID(),
    NewStatusID: newStatusID,
    OccurredAt: time.Now(),
})
```

### 2. Use Case Publica Evento

```go
// Application Layer - Use Case
func (uc *ChangeContactStatusUseCase) Execute(ctx context.Context, cmd Command) error {
    // Busca aggregate
    contact := uc.repo.FindByID(ctx, cmd.ContactID)
    
    // Executa lógica de negócio
    contact.ChangeStatus(cmd.NewStatusID)
    
    // Persiste mudanças
    uc.repo.Save(ctx, contact)
    
    // Publica eventos de domínio
    for _, event := range contact.DomainEvents() {
        uc.eventBus.Publish(ctx, event)
    }
    
    contact.ClearEvents()
    return nil
}
```

### 3. RabbitMQ Distribui Evento

```
Infrastructure Layer - Message Broker

┌─────────────────┐
│  Domain Event   │
│  Published      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   RabbitMQ      │
│   Exchange      │
└────────┬────────┘
         │
         ├──────────────────────────────────┐
         │                                  │
         ▼                                  ▼
┌─────────────────┐              ┌─────────────────┐
│  Webhook        │              │  ContactEvent   │
│  Notifier       │              │  Consumer       │
│  (External)     │              │  (Timeline)     │
└─────────────────┘              └─────────────────┘
```

### 4. Consumer Cria Contact Event

```go
// Infrastructure Layer - Consumer
func (c *ContactEventConsumer) handleContactStatusChanged(
    ctx context.Context, 
    event pipeline.ContactStatusChangedEvent,
) error {
    // Traduz Domain Event para Contact Event (timeline)
    cmd := contacteventapp.CreateContactEventCommand{
        ContactID:   event.ContactID,
        TenantID:    event.TenantID,
        EventType:   contact_event.EventTypeStatusChanged,
        Category:    contact_event.CategoryStatus,
        Priority:    contact_event.PriorityHigh,
        Source:      contact_event.SourceSystem,
        Title:       &title,
        Description: &description,
        Payload:     payload,
        IsRealtime:  true,
    }
    
    // Executa Use Case
    _, err := c.createContactEventUseCase.Execute(ctx, cmd)
    return err
}
```

### 5. Use Case Cria Aggregate e Persiste

```go
// Application Layer - CreateContactEventUseCase
func (uc *CreateContactEventUseCase) Execute(
    ctx context.Context, 
    cmd CreateContactEventCommand,
) (*contact_event.ContactEvent, error) {
    // Cria Aggregate Root (Domain Layer)
    event, err := contact_event.NewContactEvent(
        cmd.ContactID,
        cmd.TenantID,
        cmd.EventType,
        cmd.Category,
        cmd.Priority,
        cmd.Source,
    )
    
    // Configura propriedades
    event.SetTitle(*cmd.Title)
    event.AddPayloadField("status", cmd.Payload["status"])
    event.SetRealtimeDelivery(cmd.IsRealtime)
    
    // Persiste via Repository (Infrastructure Layer)
    if err := uc.repo.Save(ctx, event); err != nil {
        return nil, err
    }
    
    return event, nil
}
```

### 6. Repository Salva na Tabela

```go
// Infrastructure Layer - GormContactEventRepository
func (r *GormContactEventRepository) Save(
    ctx context.Context, 
    event *contact_event.ContactEvent,
) error {
    // Converte Domain Model para Entity
    entity := r.domainToEntity(event)
    
    // Persiste no PostgreSQL
    return r.db.WithContext(ctx).Create(entity).Error
}

// Salvo na tabela: contact_events
```

### 7. SSE Handler Lê da Tabela

```go
// Presentation Layer - SSE Handler
func (h *ContactEventStreamHandler) StreamContactEvents(c *gin.Context) {
    // Autentica e valida
    authCtx := c.Get("auth")
    
    // Polling a cada 3 segundos
    ticker := time.NewTicker(3 * time.Second)
    
    for {
        // Busca novos eventos via Repository
        events, _ := h.eventRepo.FindByContactID(
            ctx, 
            contactID, 
            limit, 
            offset,
        )
        
        // Envia via SSE
        for _, event := range events {
            h.sendSSEEvent(c, "contact_event", event)
        }
    }
}
```

### 8. Frontend Recebe Evento

```javascript
// Browser - EventSource API
const eventSource = new EventSource(
    `/api/v1/contacts/${contactId}/events/stream`,
    { withCredentials: true }
);

eventSource.addEventListener('contact_event', (e) => {
    const event = JSON.parse(e.data);
    
    // Atualiza UI
    addToTimeline(event);
    showNotification(event.title);
});
```

## Separação de Responsabilidades (DDD)

### Domain Layer (Regras de Negócio)

**ContactEvent Aggregate:**
- ✅ Encapsula regras de negócio
- ✅ Valida invariantes
- ✅ Não conhece infraestrutura
- ✅ Imutável após criação (Event Sourcing friendly)

```go
// internal/domain/contact_event/contact_event.go
type ContactEvent struct {
    id        uuid.UUID
    contactID uuid.UUID
    eventType string
    category  Category  // Value Object
    priority  Priority  // Value Object
    // ...
}

// Regras de negócio no aggregate
func (e *ContactEvent) SetExpiresAt(expiresAt time.Time) error {
    if expiresAt.Before(time.Now()) {
        return errors.New("expiresAt cannot be in the past")
    }
    e.expiresAt = &expiresAt
    return nil
}
```

**Value Objects:**
- `Category`: status, pipeline, assignment, tag, note, session
- `Priority`: low, normal, high, urgent
- `Source`: system, agent, webhook, workflow

### Application Layer (Orquestração)

**Use Cases:**
- `CreateContactEventUseCase`: Cria eventos de contato
- Coordena entre Domain e Infrastructure
- Não contém lógica de negócio
- Transacional

**Consumers:**
- `ContactEventConsumer`: Escuta Domain Events
- Traduz eventos de domínio para eventos de timeline
- Desacopla domínios (Contact, Pipeline, Session)

### Infrastructure Layer (Detalhes Técnicos)

**Repositories:**
- `GormContactEventRepository`: Persistência PostgreSQL
- Implementa interface do Domain
- Converte entre Domain Model e Entity

**Messaging:**
- `RabbitMQConnection`: Gerencia conexões
- `DomainEventBus`: Publica eventos
- `ContactEventConsumer`: Consome eventos

**Handlers:**
- `ContactEventStreamHandler`: SSE endpoint
- Segurança (auth, CSRF, XSS)
- Serialização JSON

## Benefícios da Arquitetura

### 1. Separação de Concerns
- Domain não conhece infraestrutura
- Fácil testar lógica de negócio
- Trocar PostgreSQL por MongoDB? Só muda Infrastructure

### 2. Event-Driven
- Desacoplamento entre bounded contexts
- Escalabilidade horizontal
- Auditoria completa (event log)

### 3. CQRS Implícito
- **Write**: Domain Events → RabbitMQ
- **Read**: Contact Events → PostgreSQL → SSE
- Otimizado para cada caso de uso

### 4. Testabilidade
```go
// Testar Domain sem infraestrutura
func TestContactEvent_SetExpiresAt(t *testing.T) {
    event := contact_event.NewContactEvent(...)
    
    // Regra de negócio: não pode expirar no passado
    err := event.SetExpiresAt(time.Now().Add(-1 * time.Hour))
    
    assert.Error(t, err)
}
```

### 5. Evolução
- Adicionar novo tipo de evento? Só adiciona consumer
- Mudar formato SSE? Só muda handler
- Domain permanece estável

## Próximos Passos

1. ✅ Domain Layer implementado
2. ✅ Application Layer implementado  
3. ✅ Infrastructure Layer implementado
4. ⚠️ Corrigir erros de compilação (eventos de sessão)
5. ⚠️ Registrar rotas SSE
6. ⚠️ Adicionar testes unitários
7. ⚠️ Adicionar testes de integração
8. ⚠️ Documentar APIs (Swagger)

## Segurança (Defense in Depth)

- **Domain**: Validações de invariantes
- **Application**: Validações de comando
- **Infrastructure**: Auth, CSRF, XSS, Rate Limiting
- **Presentation**: Sanitização de output

Cada camada adiciona uma camada de proteção! 🔒
