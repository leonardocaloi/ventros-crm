# 🏗️ GUIA COMPLETO DE DESENVOLVIMENTO - Ventros CRM

**Versão**: 1.0
**Última Atualização**: 2025-10-10
**Objetivo**: Explicar TODA arquitetura e fluxo de desenvolvimento do projeto

---

## 📋 ÍNDICE

1. [Visão Geral da Arquitetura](#visão-geral)
2. [Estrutura de Pastas Completa](#estrutura-de-pastas)
3. [Camadas da Aplicação](#camadas-da-aplicação)
4. [Fluxo Completo de Desenvolvimento](#fluxo-de-desenvolvimento)
5. [Checklist: Criar Nova Feature](#checklist-feature)
6. [Checklist: Criar Novo Agregado](#checklist-agregado)
7. [Padrões e Convenções](#padrões)
8. [Testes](#testes)
9. [Deploy e CI/CD](#deploy)

---

## 📐 VISÃO GERAL DA ARQUITETURA

### **Padrão Arquitetural**: Hexagonal (Ports & Adapters) + DDD + Event-Driven

```
┌─────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE LAYER                      │
│  (Adapters: HTTP, GORM, RabbitMQ, Temporal, Redis, WAHA)   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │            APPLICATION LAYER                          │  │
│  │  (Use Cases, Commands, Queries, DTOs)                │  │
│  │                                                        │  │
│  │  ┌────────────────────────────────────────────────┐  │  │
│  │  │         DOMAIN LAYER                           │  │  │
│  │  │  (Aggregates, Entities, Value Objects,         │  │  │
│  │  │   Domain Events, Repository Interfaces)        │  │  │
│  │  └────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### **Princípios**:
- ✅ **Dependency Inversion**: Domain não depende de nada, Application depende de Domain, Infrastructure depende de tudo
- ✅ **Event-Driven**: Agregados emitem eventos, consumidores reagem
- ✅ **CQRS**: Commands (escrita) separados de Queries (leitura)
- ✅ **Outbox Pattern**: Eventos persistidos transacionalmente
- ✅ **Saga Pattern**: Workflows complexos com compensation

---

## 📁 ESTRUTURA DE PASTAS COMPLETA

```
ventros-crm/
├── cmd/                                    # Entry points
│   ├── api/main.go                        # API server
│   └── migrate/main.go                    # Migrations CLI
│
├── internal/                              # Código privado
│   ├── domain/                            # 🔵 DOMAIN LAYER (core de negócio)
│   │   ├── contact/
│   │   │   ├── contact.go                # Aggregate root
│   │   │   ├── events.go                 # Domain events
│   │   │   ├── repository.go             # Repository interface
│   │   │   ├── types.go                  # Enums, constants
│   │   │   ├── value_objects.go          # Email, Phone, etc.
│   │   │   └── errors.go                 # Domain errors
│   │   ├── session/                      # Mesmo padrão
│   │   ├── message/                      # Mesmo padrão
│   │   ├── pipeline/                     # Mesmo padrão
│   │   ├── agent/                        # Mesmo padrão
│   │   └── [22 agregados no total]
│   │
│   ├── application/                       # 🟢 APPLICATION LAYER (casos de uso)
│   │   ├── commands/                     # CQRS Commands (escrita)
│   │   │   ├── contact/
│   │   │   │   ├── create_contact_command.go
│   │   │   │   ├── update_contact_command.go
│   │   │   │   └── delete_contact_command.go
│   │   │   ├── session/
│   │   │   ├── message/
│   │   │   └── [commands de todos agregados]
│   │   │
│   │   ├── queries/                      # CQRS Queries (leitura)
│   │   │   ├── list_contacts_query.go
│   │   │   ├── search_contacts_query.go
│   │   │   ├── get_contact_stats_query.go
│   │   │   └── [20+ queries]
│   │   │
│   │   ├── contact/                      # Use cases de Contact
│   │   │   ├── create_contact_usecase.go
│   │   │   ├── update_contact_usecase.go
│   │   │   └── fetch_profile_picture_usecase.go
│   │   │
│   │   ├── message/                      # Use cases de Message
│   │   │   ├── process_inbound_message.go
│   │   │   └── send_message_usecase.go
│   │   │
│   │   └── dtos/                         # Data Transfer Objects
│   │       ├── contact_dto.go
│   │       ├── session_dto.go
│   │       └── [DTOs de todas entidades]
│   │
│   └── workflows/                         # Temporal Workflows
│       ├── session/
│       │   ├── session_lifecycle_workflow.go
│       │   └── session_activities.go
│       ├── billing/                       # Sagas de pagamento
│       └── outbox/                        # Outbox processor
│
├── infrastructure/                        # 🟡 INFRASTRUCTURE LAYER (adapters)
│   ├── http/                             # HTTP Adapter (Gin)
│   │   ├── handlers/
│   │   │   ├── contact_handler.go
│   │   │   ├── session_handler.go
│   │   │   ├── message_handler.go
│   │   │   └── [19 handlers]
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── correlation_id.go
│   │   │   ├── rate_limit.go
│   │   │   └── error_handler.go
│   │   ├── routes/
│   │   │   └── routes.go                # Registro de rotas
│   │   └── responses/
│   │       └── envelope.go              # Response padrão
│   │
│   ├── persistence/                      # Database Adapter (GORM)
│   │   ├── entities/                    # GORM Entities
│   │   │   ├── contact.go
│   │   │   ├── session.go
│   │   │   └── [29 entities]
│   │   ├── gorm_contact_repository.go   # Repository implementation
│   │   ├── gorm_session_repository.go
│   │   └── [18 repositories]
│   │
│   ├── messaging/                        # Message Broker Adapter
│   │   ├── rabbitmq.go                  # RabbitMQ client
│   │   ├── domain_event_bus.go          # Event publisher
│   │   ├── postgres_notify_outbox.go    # Outbox processor
│   │   ├── contact_event_consumer.go    # Event consumer
│   │   └── [15+ consumers]
│   │
│   ├── workflow/                         # Temporal Adapter
│   │   ├── temporal.go                  # Temporal client
│   │   ├── session_worker.go            # Session worker
│   │   └── outbox_worker.go             # Outbox worker
│   │
│   ├── cache/                            # Cache Adapter (Redis)
│   │   ├── redis.go                     # Redis client
│   │   ├── repository_cache.go          # Cache layer
│   │   ├── session_cache.go             # Session cache
│   │   └── distributed_lock.go          # Distributed locks
│   │
│   ├── channels/                         # External Services
│   │   └── waha/
│   │       ├── client.go                # WAHA HTTP client
│   │       └── profile_service.go       # Profile fetcher
│   │
│   ├── websocket/                        # WebSocket Adapter
│   │   ├── hub.go                       # WebSocket hub
│   │   ├── client.go                    # WebSocket client
│   │   └── message.go                   # WebSocket messages
│   │
│   ├── database/                         # Database Infrastructure
│   │   ├── migrations/                  # GORM Migrations
│   │   │   ├── migrator.go             # Migration manager
│   │   │   ├── 000001_initial_schema.go
│   │   │   └── [42 migrations]
│   │   └── migrations.go                # Migration checker
│   │
│   └── config/
│       └── config.go                    # Config loader
│
├── docs/                                 # Documentação
│   ├── docs.go                          # Swagger generated
│   ├── swagger.json                     # Swagger spec
│   └── webhook_events.md                # Webhook docs
│
├── tests/                                # Testes E2E/Integração
│   ├── e2e/
│   └── integration/
│
├── TODO_NEW.md                           # Roadmap completo
├── DEV_GUIDE.md                          # Este arquivo
└── README.md                             # Documentação principal
```

---

## 🎯 CAMADAS DA APLICAÇÃO

### **1. DOMAIN LAYER** (Núcleo de Negócio)

**Responsabilidade**: Lógica de negócio pura, sem dependências externas.

**Componentes**:

#### **1.1. Aggregate Root** (Ex: `contact.go`)
```go
// contact.go
package contact

type Contact struct {
    // Private fields (encapsulamento)
    id        uuid.UUID
    name      string
    email     *Email      // Value Object
    phone     *Phone      // Value Object
    events    []DomainEvent  // Domain Events
}

// Factory method
func NewContact(projectID uuid.UUID, tenantID string, name string) (*Contact, error) {
    // Validação de invariantes
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }

    contact := &Contact{
        id:     uuid.New(),
        name:   name,
        events: []DomainEvent{},
    }

    // Emitir evento
    contact.addEvent(NewContactCreatedEvent(contact.id, projectID, tenantID, name))

    return contact, nil
}

// Métodos de negócio
func (c *Contact) UpdateName(name string) error {
    // Validação
    if name == "" {
        return errors.New("name cannot be empty")
    }

    // Mudança de estado
    c.name = name
    c.updatedAt = time.Now()

    // Emitir evento
    c.addEvent(NewContactUpdatedEvent(c.id))

    return nil
}

// Getters (expor estado de forma controlada)
func (c *Contact) ID() uuid.UUID { return c.id }
func (c *Contact) Name() string { return c.name }

// Domain Events
func (c *Contact) DomainEvents() []DomainEvent {
    return append([]DomainEvent{}, c.events...)
}

func (c *Contact) ClearEvents() {
    c.events = []DomainEvent{}
}
```

#### **1.2. Domain Events** (Ex: `events.go`)
```go
// events.go
package contact

type ContactCreatedEvent struct {
    eventID   uuid.UUID
    contactID uuid.UUID
    projectID uuid.UUID
    tenantID  string
    name      string
    timestamp time.Time
}

func NewContactCreatedEvent(contactID, projectID uuid.UUID, tenantID, name string) ContactCreatedEvent {
    return ContactCreatedEvent{
        eventID:   uuid.New(),
        contactID: contactID,
        projectID: projectID,
        tenantID:  tenantID,
        name:      name,
        timestamp: time.Now(),
    }
}

// Implementar interface shared.DomainEvent
func (e ContactCreatedEvent) EventID() uuid.UUID { return e.eventID }
func (e ContactCreatedEvent) EventName() string { return "contact.created" }
func (e ContactCreatedEvent) EventVersion() string { return "1.0" }
func (e ContactCreatedEvent) OccurredAt() time.Time { return e.timestamp }
```

#### **1.3. Repository Interface** (Ex: `repository.go`)
```go
// repository.go
package contact

type Repository interface {
    Save(ctx context.Context, contact *Contact) error
    FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)
    FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*Contact, error)
    FindByEmail(ctx context.Context, projectID uuid.UUID, email string) (*Contact, error)
    // ... outros métodos
}
```

#### **1.4. Value Objects** (Ex: `value_objects.go`)
```go
// value_objects.go
package contact

type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    // Validação
    if email == "" {
        return Email{}, errors.New("email cannot be empty")
    }

    // Regex de validação
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return Email{}, errors.New("invalid email format")
    }

    return Email{value: email}, nil
}

func (e Email) String() string {
    return e.value
}
```

#### **1.5. Domain Errors** (Ex: `errors.go`)
```go
// errors.go
package contact

import "github.com/caloi/ventros-crm/internal/domain/shared"

var (
    ErrContactNotFound = errors.New("contact not found")
)

func NewContactNotFoundError(contactID string) *shared.DomainError {
    err := shared.NewNotFoundError("contact", contactID)
    err.Err = ErrContactNotFound // Wrap sentinel error
    return err
}
```

---

### **2. APPLICATION LAYER** (Casos de Uso)

**Responsabilidade**: Orquestrar agregados, coordenar transações, publicar eventos.

#### **2.1. Commands** (Escrita)
```go
// internal/application/commands/contact/create_contact_command.go
package contact

type CreateContactCommand struct {
    ProjectID  uuid.UUID
    TenantID   string
    Name       string
    Email      *string
    Phone      *string
    Tags       []string
}

func (cmd *CreateContactCommand) Validate() error {
    if cmd.ProjectID == uuid.Nil {
        return errors.New("project_id is required")
    }
    if cmd.TenantID == "" {
        return errors.New("tenant_id is required")
    }
    if cmd.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

type CreateContactCommandHandler struct {
    contactRepo contact.Repository
    eventBus    shared.EventBus
}

func (h *CreateContactCommandHandler) Handle(ctx context.Context, cmd *CreateContactCommand) (*contact.Contact, error) {
    // 1. Validar comando
    if err := cmd.Validate(); err != nil {
        return nil, err
    }

    // 2. Criar agregado
    c, err := contact.NewContact(cmd.ProjectID, cmd.TenantID, cmd.Name)
    if err != nil {
        return nil, err
    }

    // 3. Setar campos opcionais
    if cmd.Email != nil {
        c.SetEmail(*cmd.Email)
    }
    if cmd.Phone != nil {
        c.SetPhone(*cmd.Phone)
    }

    // 4. Salvar no repositório
    if err := h.contactRepo.Save(ctx, c); err != nil {
        return nil, err
    }

    // 5. Publicar eventos de domínio
    events := c.DomainEvents()
    if len(events) > 0 {
        if err := h.eventBus.PublishBatch(ctx, events); err != nil {
            // Log error but don't fail
        }
        c.ClearEvents()
    }

    return c, nil
}
```

#### **2.2. Queries** (Leitura)
```go
// internal/application/queries/list_contacts_query.go
package queries

type ListContactsQuery struct {
    TenantID    string
    ProjectID   *uuid.UUID
    Name        string
    Tags        []string
    Page        int
    Limit       int
    SortBy      string
    SortDir     string
}

type ListContactsQueryHandler struct {
    contactRepo contact.Repository
}

func (h *ListContactsQueryHandler) Handle(ctx context.Context, query *ListContactsQuery) ([]*ContactDTO, int64, error) {
    // Criar filtros
    filters := contact.ContactFilters{
        Name: query.Name,
        Tags: query.Tags,
    }

    // Buscar no repositório
    contacts, total, err := h.contactRepo.FindByTenantWithFilters(
        ctx,
        query.TenantID,
        filters,
        query.Page,
        query.Limit,
        query.SortBy,
        query.SortDir,
    )
    if err != nil {
        return nil, 0, err
    }

    // Converter para DTO
    dtos := make([]*ContactDTO, len(contacts))
    for i, c := range contacts {
        dtos[i] = ToContactDTO(c)
    }

    return dtos, total, nil
}
```

#### **2.3. Use Cases** (Orquestração)
```go
// internal/application/contact/create_contact_usecase.go
package contact

type CreateContactUseCase struct {
    contactRepo contact.Repository
    projectRepo project.Repository
    eventBus    shared.EventBus
}

func (uc *CreateContactUseCase) Execute(ctx context.Context, req CreateContactRequest) (*ContactDTO, error) {
    // 1. Validar projeto existe
    _, err := uc.projectRepo.FindByID(ctx, req.ProjectID)
    if err != nil {
        return nil, errors.New("project not found")
    }

    // 2. Verificar se contato já existe (por phone ou email)
    if req.Phone != nil {
        existing, _ := uc.contactRepo.FindByPhone(ctx, req.ProjectID, *req.Phone)
        if existing != nil {
            return nil, errors.New("contact with this phone already exists")
        }
    }

    // 3. Criar contato
    c, err := contact.NewContact(req.ProjectID, req.TenantID, req.Name)
    if err != nil {
        return nil, err
    }

    // 4. Setar campos opcionais
    if req.Email != nil {
        c.SetEmail(*req.Email)
    }
    if req.Phone != nil {
        c.SetPhone(*req.Phone)
    }
    for _, tag := range req.Tags {
        c.AddTag(tag)
    }

    // 5. Salvar (dentro de transação)
    if err := uc.contactRepo.Save(ctx, c); err != nil {
        return nil, err
    }

    // 6. Publicar eventos via Outbox
    events := c.DomainEvents()
    if len(events) > 0 {
        for _, event := range events {
            if err := uc.eventBus.Publish(ctx, event); err != nil {
                // Log but don't fail
            }
        }
        c.ClearEvents()
    }

    // 7. Retornar DTO
    return ToContactDTO(c), nil
}
```

#### **2.4. DTOs** (Data Transfer Objects)
```go
// internal/application/dtos/contact_dto.go
package dtos

type ContactDTO struct {
    ID                string    `json:"id"`
    ProjectID         string    `json:"project_id"`
    TenantID          string    `json:"tenant_id"`
    Name              string    `json:"name"`
    Email             *string   `json:"email,omitempty"`
    Phone             *string   `json:"phone,omitempty"`
    Tags              []string  `json:"tags,omitempty"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
}

// Mapper: Domain → DTO
func ToContactDTO(c *contact.Contact) *ContactDTO {
    dto := &ContactDTO{
        ID:        c.ID().String(),
        ProjectID: c.ProjectID().String(),
        TenantID:  c.TenantID(),
        Name:      c.Name(),
        Tags:      c.Tags(),
        CreatedAt: c.CreatedAt(),
        UpdatedAt: c.UpdatedAt(),
    }

    if email := c.Email(); email != nil {
        emailStr := email.String()
        dto.Email = &emailStr
    }

    if phone := c.Phone(); phone != nil {
        phoneStr := phone.String()
        dto.Phone = &phoneStr
    }

    return dto
}
```

---

### **3. INFRASTRUCTURE LAYER** (Adapters)

#### **3.1. HTTP Handler** (Gin)
```go
// infrastructure/http/handlers/contact_handler.go
package handlers

type ContactHandler struct {
    createUseCase *contact.CreateContactUseCase
    listQuery     *queries.ListContactsQueryHandler
}

// CreateContact godoc
// @Summary      Criar novo contato
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Param        request body CreateContactRequest true "Dados do contato"
// @Success      201  {object}  APIResponse{data=ContactDTO}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Router       /api/v1/contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
    // 1. Parse request
    var req CreateContactRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, NewErrorResponse("VALIDATION_FAILED", err.Error()))
        return
    }

    // 2. Extract tenant from context
    tenantID := c.GetString("tenant_id")
    req.TenantID = tenantID

    // 3. Execute use case
    dto, err := h.createUseCase.Execute(c.Request.Context(), req)
    if err != nil {
        status, apiErr := MapDomainErrorToHTTP(err)
        c.JSON(status, NewErrorResponse(apiErr.Code, apiErr.Message))
        return
    }

    // 4. Return success
    c.JSON(201, NewSuccessResponse(dto, nil, nil))
}
```

#### **3.2. Repository Implementation** (GORM)
```go
// infrastructure/persistence/gorm_contact_repository.go
package persistence

type GormContactRepository struct {
    db *gorm.DB
}

func (r *GormContactRepository) Save(ctx context.Context, c *contact.Contact) error {
    entity := r.domainToEntity(c)
    return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error) {
    var entity entities.ContactEntity
    err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, contact.NewContactNotFoundError(id.String())
        }
        return nil, err
    }
    return r.entityToDomain(&entity), nil
}

// Mapper: Domain → Entity
func (r *GormContactRepository) domainToEntity(c *contact.Contact) *entities.ContactEntity {
    entity := &entities.ContactEntity{
        ID:        c.ID(),
        ProjectID: c.ProjectID(),
        TenantID:  c.TenantID(),
        Name:      c.Name(),
        Tags:      entities.StringArray(c.Tags()),
        CreatedAt: c.CreatedAt(),
        UpdatedAt: c.UpdatedAt(),
    }

    if email := c.Email(); email != nil {
        entity.Email = email.String()
    }

    if phone := c.Phone(); phone != nil {
        entity.Phone = phone.String()
    }

    return entity
}

// Mapper: Entity → Domain
func (r *GormContactRepository) entityToDomain(entity *entities.ContactEntity) *contact.Contact {
    var email *contact.Email
    if entity.Email != "" {
        if e, err := contact.NewEmail(entity.Email); err == nil {
            email = &e
        }
    }

    var phone *contact.Phone
    if entity.Phone != "" {
        if p, err := contact.NewPhone(entity.Phone); err == nil {
            phone = &p
        }
    }

    return contact.ReconstructContact(
        entity.ID,
        entity.ProjectID,
        entity.TenantID,
        entity.Name,
        email,
        phone,
        nil, nil, // externalID, sourceChannel
        entity.Language,
        nil, // timezone
        []string(entity.Tags),
        nil, nil, nil, nil, // profile picture, interactions
        entity.CreatedAt,
        entity.UpdatedAt,
        nil, // deletedAt
    )
}
```

#### **3.3. GORM Entity**
```go
// infrastructure/persistence/entities/contact.go
package entities

type ContactEntity struct {
    ID                      uuid.UUID `gorm:"type:uuid;primaryKey"`
    ProjectID               uuid.UUID `gorm:"type:uuid;not null;index:idx_contacts_project"`
    TenantID                string    `gorm:"type:text;not null;index:idx_contacts_tenant"`
    Name                    string    `gorm:"type:text;not null;index:idx_contacts_name"`
    Email                   string    `gorm:"type:text;index:idx_contacts_email"`
    Phone                   string    `gorm:"type:text;index:idx_contacts_phone"`
    ExternalID              string    `gorm:"type:text;index:idx_contacts_external_id"`
    SourceChannel           string    `gorm:"type:text"`
    Language                string    `gorm:"type:text;not null;default:'en'"`
    Timezone                string    `gorm:"type:text"`
    Tags                    StringArray `gorm:"type:text[]"` // PostgreSQL array
    ProfilePictureURL       *string   `gorm:"type:text"`
    ProfilePictureFetchedAt *time.Time
    FirstInteractionAt      *time.Time
    LastInteractionAt       *time.Time
    CreatedAt               time.Time `gorm:"not null;index:idx_contacts_created"`
    UpdatedAt               time.Time `gorm:"not null"`
    DeletedAt               gorm.DeletedAt `gorm:"index:idx_contacts_deleted"`
}

func (ContactEntity) TableName() string {
    return "contacts"
}

// Custom GORM type for PostgreSQL array
type StringArray []string

func (a StringArray) GormDataType() string {
    return "text[]"
}
```

#### **3.4. Event Bus** (Outbox)
```go
// infrastructure/messaging/domain_event_bus.go
package messaging

type DomainEventBus struct {
    db         *gorm.DB
    outboxRepo outbox.Repository
}

func (bus *DomainEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
    // 1. Serializar evento
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // 2. Criar evento de outbox
    outboxEvent := &outbox.OutboxEvent{
        ID:            uuid.New(),
        EventID:       event.EventID(),
        EventType:     event.EventName(),
        EventVersion:  event.EventVersion(),
        EventData:     payload,
        TenantID:      extractTenantID(ctx),
        CreatedAt:     time.Now(),
        Status:        outbox.StatusPending,
    }

    // 3. Salvar no outbox (mesma transação do agregado)
    if err := bus.outboxRepo.Save(ctx, outboxEvent); err != nil {
        return err
    }

    // 4. PostgreSQL LISTEN/NOTIFY trigger irá notificar o processor
    // 5. Processor publicará no RabbitMQ

    return nil
}
```

---

## 🔄 FLUXO COMPLETO DE DESENVOLVIMENTO

### **Exemplo: Criar Feature "Adicionar Nota ao Contato"**

#### **PASSO 1: Identificar o Agregado**
- A nota pertence a qual agregado?
  - **Opção A**: Nota é parte de Contact → Adicionar método `Contact.AddNote()`
  - **Opção B**: Nota é agregado separado → Criar `internal/domain/note/`

**Decisão**: Nota é agregado separado (tem lifecycle próprio, pode existir sem Contact)

#### **PASSO 2: Domain Layer**

**2.1. Criar agregado Note**
```go
// internal/domain/note/note.go
package note

type Note struct {
    id        uuid.UUID
    contactID uuid.UUID
    authorID  uuid.UUID // Agent que criou
    text      string
    isPinned  bool
    createdAt time.Time
    updatedAt time.Time
    events    []shared.DomainEvent
}

func NewNote(contactID, authorID uuid.UUID, text string) (*Note, error) {
    if contactID == uuid.Nil {
        return nil, errors.New("contact_id is required")
    }
    if text == "" {
        return nil, errors.New("text cannot be empty")
    }

    note := &Note{
        id:        uuid.New(),
        contactID: contactID,
        authorID:  authorID,
        text:      text,
        isPinned:  false,
        createdAt: time.Now(),
        updatedAt: time.Now(),
        events:    []shared.DomainEvent{},
    }

    note.addEvent(NewNoteAddedEvent(note.id, contactID, authorID))

    return note, nil
}

func (n *Note) Pin() {
    n.isPinned = true
    n.updatedAt = time.Now()
    n.addEvent(NewNotePinnedEvent(n.id))
}

func (n *Note) UpdateText(text string) error {
    if text == "" {
        return errors.New("text cannot be empty")
    }
    n.text = text
    n.updatedAt = time.Now()
    n.addEvent(NewNoteUpdatedEvent(n.id))
    return nil
}
```

**2.2. Criar eventos**
```go
// internal/domain/note/events.go
package note

type NoteAddedEvent struct {
    eventID   uuid.UUID
    noteID    uuid.UUID
    contactID uuid.UUID
    authorID  uuid.UUID
    timestamp time.Time
}

func NewNoteAddedEvent(noteID, contactID, authorID uuid.UUID) NoteAddedEvent {
    return NoteAddedEvent{
        eventID:   uuid.New(),
        noteID:    noteID,
        contactID: contactID,
        authorID:  authorID,
        timestamp: time.Now(),
    }
}

func (e NoteAddedEvent) EventID() uuid.UUID { return e.eventID }
func (e NoteAddedEvent) EventName() string { return "note.added" }
func (e NoteAddedEvent) EventVersion() string { return "1.0" }
func (e NoteAddedEvent) OccurredAt() time.Time { return e.timestamp }
```

**2.3. Criar repository interface**
```go
// internal/domain/note/repository.go
package note

type Repository interface {
    Save(ctx context.Context, note *Note) error
    FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
    FindByContact(ctx context.Context, contactID uuid.UUID) ([]*Note, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

#### **PASSO 3: Application Layer**

**3.1. Criar Command**
```go
// internal/application/commands/note/add_note_command.go
package note

type AddNoteCommand struct {
    ContactID uuid.UUID
    AuthorID  uuid.UUID
    Text      string
}

func (cmd *AddNoteCommand) Validate() error {
    if cmd.ContactID == uuid.Nil {
        return errors.New("contact_id is required")
    }
    if cmd.AuthorID == uuid.Nil {
        return errors.New("author_id is required")
    }
    if cmd.Text == "" {
        return errors.New("text is required")
    }
    return nil
}

type AddNoteCommandHandler struct {
    noteRepo    note.Repository
    contactRepo contact.Repository
    eventBus    shared.EventBus
}

func (h *AddNoteCommandHandler) Handle(ctx context.Context, cmd *AddNoteCommand) (*note.Note, error) {
    // Validar
    if err := cmd.Validate(); err != nil {
        return nil, err
    }

    // Verificar se contato existe
    _, err := h.contactRepo.FindByID(ctx, cmd.ContactID)
    if err != nil {
        return nil, errors.New("contact not found")
    }

    // Criar nota
    n, err := note.NewNote(cmd.ContactID, cmd.AuthorID, cmd.Text)
    if err != nil {
        return nil, err
    }

    // Salvar
    if err := h.noteRepo.Save(ctx, n); err != nil {
        return nil, err
    }

    // Publicar eventos
    events := n.DomainEvents()
    for _, event := range events {
        h.eventBus.Publish(ctx, event)
    }
    n.ClearEvents()

    return n, nil
}
```

**3.2. Criar Query**
```go
// internal/application/queries/list_notes_query.go
package queries

type ListNotesQuery struct {
    ContactID uuid.UUID
    TenantID  string
}

type ListNotesQueryHandler struct {
    noteRepo note.Repository
}

func (h *ListNotesQueryHandler) Handle(ctx context.Context, query *ListNotesQuery) ([]*NoteDTO, error) {
    notes, err := h.noteRepo.FindByContact(ctx, query.ContactID)
    if err != nil {
        return nil, err
    }

    dtos := make([]*NoteDTO, len(notes))
    for i, n := range notes {
        dtos[i] = ToNoteDTO(n)
    }

    return dtos, nil
}
```

**3.3. Criar DTO**
```go
// internal/application/dtos/note_dto.go
package dtos

type NoteDTO struct {
    ID        string    `json:"id"`
    ContactID string    `json:"contact_id"`
    AuthorID  string    `json:"author_id"`
    Text      string    `json:"text"`
    IsPinned  bool      `json:"is_pinned"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func ToNoteDTO(n *note.Note) *NoteDTO {
    return &NoteDTO{
        ID:        n.ID().String(),
        ContactID: n.ContactID().String(),
        AuthorID:  n.AuthorID().String(),
        Text:      n.Text(),
        IsPinned:  n.IsPinned(),
        CreatedAt: n.CreatedAt(),
        UpdatedAt: n.UpdatedAt(),
    }
}
```

#### **PASSO 4: Infrastructure Layer**

**4.1. Criar GORM Entity**
```go
// infrastructure/persistence/entities/note.go
package entities

type NoteEntity struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
    ContactID uuid.UUID `gorm:"type:uuid;not null;index:idx_notes_contact"`
    AuthorID  uuid.UUID `gorm:"type:uuid;not null"`
    Text      string    `gorm:"type:text;not null"`
    IsPinned  bool      `gorm:"default:false"`
    CreatedAt time.Time `gorm:"not null"`
    UpdatedAt time.Time `gorm:"not null"`
}

func (NoteEntity) TableName() string {
    return "notes"
}
```

**4.2. Criar GORM Repository**
```go
// infrastructure/persistence/gorm_note_repository.go
package persistence

type GormNoteRepository struct {
    db *gorm.DB
}

func (r *GormNoteRepository) Save(ctx context.Context, n *note.Note) error {
    entity := &entities.NoteEntity{
        ID:        n.ID(),
        ContactID: n.ContactID(),
        AuthorID:  n.AuthorID(),
        Text:      n.Text(),
        IsPinned:  n.IsPinned(),
        CreatedAt: n.CreatedAt(),
        UpdatedAt: n.UpdatedAt(),
    }
    return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormNoteRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]*note.Note, error) {
    var entities []entities.NoteEntity
    err := r.db.WithContext(ctx).
        Where("contact_id = ?", contactID).
        Order("created_at DESC").
        Find(&entities).Error
    if err != nil {
        return nil, err
    }

    notes := make([]*note.Note, len(entities))
    for i, e := range entities {
        notes[i] = note.ReconstructNote(
            e.ID, e.ContactID, e.AuthorID,
            e.Text, e.IsPinned,
            e.CreatedAt, e.UpdatedAt,
        )
    }
    return notes, nil
}
```

**4.3. Criar Migration GORM**
```go
// infrastructure/database/migrations/000043_add_notes.go
package migrations

type Migration000043AddNotes struct{}

func (m *Migration000043AddNotes) ID() string {
    return "000043_add_notes"
}

func (m *Migration000043AddNotes) Up(db *gorm.DB) error {
    // Criar tabela
    if err := db.AutoMigrate(&entities.NoteEntity{}); err != nil {
        return err
    }

    // Criar índices
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_notes_contact ON notes (contact_id, created_at DESC)",
    }

    for _, idx := range indexes {
        if err := db.Exec(idx).Error; err != nil {
            return err
        }
    }

    return nil
}

func (m *Migration000043AddNotes) Down(db *gorm.DB) error {
    db.Exec("DROP INDEX IF EXISTS idx_notes_contact")
    return db.Migrator().DropTable(&entities.NoteEntity{})
}
```

**4.4. Criar HTTP Handler**
```go
// infrastructure/http/handlers/note_handler.go
package handlers

type NoteHandler struct {
    addNoteCmd  *note.AddNoteCommandHandler
    listQuery   *queries.ListNotesQueryHandler
}

// AddNote godoc
// @Summary      Adicionar nota ao contato
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Param        request body AddNoteRequest true "Dados da nota"
// @Success      201  {object}  APIResponse{data=NoteDTO}
// @Failure      400  {object}  APIResponse{error=APIError}
// @Router       /api/v1/notes [post]
func (h *NoteHandler) AddNote(c *gin.Context) {
    var req AddNoteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, NewErrorResponse("VALIDATION_FAILED", err.Error()))
        return
    }

    authorID := c.GetString("user_id") // Do token JWT

    cmd := note.AddNoteCommand{
        ContactID: req.ContactID,
        AuthorID:  uuid.MustParse(authorID),
        Text:      req.Text,
    }

    n, err := h.addNoteCmd.Handle(c.Request.Context(), &cmd)
    if err != nil {
        status, apiErr := MapDomainErrorToHTTP(err)
        c.JSON(status, NewErrorResponse(apiErr.Code, apiErr.Message))
        return
    }

    c.JSON(201, NewSuccessResponse(ToNoteDTO(n), nil, nil))
}

// ListNotes godoc
// @Summary      Listar notas de um contato
// @Tags         Notes
// @Produce      json
// @Param        contact_id query string true "Contact ID"
// @Success      200  {object}  APIResponse{data=[]NoteDTO}
// @Router       /api/v1/notes [get]
func (h *NoteHandler) ListNotes(c *gin.Context) {
    contactID, err := uuid.Parse(c.Query("contact_id"))
    if err != nil {
        c.JSON(400, NewErrorResponse("INVALID_CONTACT_ID", "Invalid contact_id"))
        return
    }

    query := queries.ListNotesQuery{
        ContactID: contactID,
        TenantID:  c.GetString("tenant_id"),
    }

    notes, err := h.listQuery.Handle(c.Request.Context(), &query)
    if err != nil {
        c.JSON(500, NewErrorResponse("INTERNAL_ERROR", err.Error()))
        return
    }

    c.JSON(200, NewSuccessResponse(notes, nil, nil))
}
```

**4.5. Registrar Rotas**
```go
// infrastructure/http/routes/routes.go
func SetupRoutes(router *gin.Engine, deps *Dependencies) {
    api := router.Group("/api/v1")
    api.Use(middleware.AuthMiddleware())

    // Notes
    notes := api.Group("/notes")
    {
        notes.POST("", deps.Handlers.Note.AddNote)
        notes.GET("", deps.Handlers.Note.ListNotes)
        notes.PUT("/:id", deps.Handlers.Note.UpdateNote)
        notes.DELETE("/:id", deps.Handlers.Note.DeleteNote)
        notes.POST("/:id/pin", deps.Handlers.Note.PinNote)
    }
}
```

**4.6. Atualizar Dependency Injection**
```go
// cmd/api/main.go
func main() {
    // ... setup anterior

    // Repositories
    noteRepo := persistence.NewGormNoteRepository(db)

    // Commands
    addNoteCmd := note.NewAddNoteCommandHandler(noteRepo, contactRepo, eventBus)

    // Queries
    listNotesQuery := queries.NewListNotesQueryHandler(noteRepo)

    // Handlers
    noteHandler := handlers.NewNoteHandler(addNoteCmd, listNotesQuery)

    // Routes
    routes.SetupRoutes(router, &routes.Dependencies{
        Handlers: &routes.Handlers{
            Note: noteHandler,
            // ... outros handlers
        },
    })
}
```

#### **PASSO 5: Documentação**

**5.1. Atualizar Swagger**
```bash
swag init -g cmd/api/main.go
```

**5.2. Documentar Evento Webhook**
```markdown
# docs/webhook_events.md

## note.added
**Disparado quando**: Uma nota é adicionada a um contato
**Payload**:
{
  "event_id": "uuid",
  "event_type": "note.added",
  "event_version": "1.0",
  "timestamp": "2025-01-01T00:00:00Z",
  "data": {
    "note_id": "uuid",
    "contact_id": "uuid",
    "author_id": "uuid",
    "text": "string",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

#### **PASSO 6: Testes**

**6.1. Testes de Domínio**
```go
// internal/domain/note/note_test.go
func TestNewNote(t *testing.T) {
    contactID := uuid.New()
    authorID := uuid.New()

    n, err := note.NewNote(contactID, authorID, "Test note")

    assert.NoError(t, err)
    assert.NotNil(t, n)
    assert.Equal(t, "Test note", n.Text())
    assert.False(t, n.IsPinned())

    // Verificar evento emitido
    events := n.DomainEvents()
    assert.Len(t, events, 1)
    assert.Equal(t, "note.added", events[0].EventName())
}
```

**6.2. Testes de Repository**
```go
// infrastructure/persistence/gorm_note_repository_test.go
func TestGormNoteRepository_Save(t *testing.T) {
    testDB := SetupTestDatabase(t)
    defer testDB.TeardownTestDatabase(t)

    repo := persistence.NewGormNoteRepository(testDB.DB)

    n, _ := note.NewNote(uuid.New(), uuid.New(), "Test")
    err := repo.Save(context.Background(), n)

    assert.NoError(t, err)

    // Verificar no banco
    found, err := repo.FindByID(context.Background(), n.ID())
    assert.NoError(t, err)
    assert.Equal(t, n.ID(), found.ID())
}
```

**6.3. Testes de Handler**
```go
// infrastructure/http/handlers/note_handler_test.go
func TestNoteHandler_AddNote(t *testing.T) {
    // Setup mocks
    mockRepo := &MockNoteRepository{}
    handler := handlers.NewNoteHandler(mockRepo, ...)

    // Create request
    body := `{"contact_id":"...","text":"Test"}`
    req := httptest.NewRequest("POST", "/api/v1/notes", strings.NewReader(body))
    rec := httptest.NewRecorder()

    // Execute
    c, _ := gin.CreateTestContext(rec)
    c.Request = req
    handler.AddNote(c)

    // Assert
    assert.Equal(t, 201, rec.Code)
}
```

---

## ✅ CHECKLIST: CRIAR NOVA FEATURE

Use este checklist para garantir que implementou TUDO corretamente:

### **Domain Layer**
- [ ] Criar/Atualizar agregado em `internal/domain/{aggregate}/`
  - [ ] Aggregate root com métodos de negócio
  - [ ] Validação de invariantes
  - [ ] Emitir eventos de domínio
- [ ] Criar eventos em `internal/domain/{aggregate}/events.go`
  - [ ] Implementar interface `shared.DomainEvent`
  - [ ] Nomear evento corretamente (resource.action)
- [ ] Criar/Atualizar repository interface em `internal/domain/{aggregate}/repository.go`
- [ ] Criar value objects se necessário em `internal/domain/{aggregate}/value_objects.go`
- [ ] Criar domain errors em `internal/domain/{aggregate}/errors.go`

### **Application Layer**
- [ ] Criar Command em `internal/application/commands/{aggregate}/`
  - [ ] Command struct com validação
  - [ ] CommandHandler com lógica de orquestração
- [ ] Criar Query em `internal/application/queries/`
  - [ ] Query struct com filtros
  - [ ] QueryHandler com lógica de busca
- [ ] Criar/Atualizar DTO em `internal/application/dtos/`
  - [ ] Struct com tags JSON
  - [ ] Mapper Domain → DTO
- [ ] Criar Use Case se necessário em `internal/application/{aggregate}/`

### **Infrastructure Layer**
- [ ] Criar GORM Entity em `infrastructure/persistence/entities/`
  - [ ] Struct com tags GORM
  - [ ] TableName()
  - [ ] Indexes via tags ou migration
- [ ] Criar GORM Repository em `infrastructure/persistence/`
  - [ ] Implementar interface do domínio
  - [ ] Mappers: Domain ↔ Entity
- [ ] Criar Migration em `infrastructure/database/migrations/`
  - [ ] Implementar interface Migration
  - [ ] Up() e Down()
  - [ ] Registrar em NewMigrator()
- [ ] Criar HTTP Handler em `infrastructure/http/handlers/`
  - [ ] Swagger annotations completas
  - [ ] Request/Response structs
  - [ ] Error handling com APIResponse
- [ ] Registrar rotas em `infrastructure/http/routes/routes.go`
  - [ ] Adicionar rota ao grupo correto
  - [ ] Aplicar middlewares (auth, rate limit)
- [ ] Atualizar DI em `cmd/api/main.go`
  - [ ] Instanciar repository
  - [ ] Instanciar command/query handlers
  - [ ] Instanciar handler HTTP
  - [ ] Passar para routes

### **Documentação**
- [ ] Swagger docs
  - [ ] Annotations no handler
  - [ ] Regenerar: `swag init -g cmd/api/main.go`
- [ ] Webhook docs em `docs/webhook_events.md`
  - [ ] Documentar eventos emitidos
  - [ ] Exemplo de payload

### **Testes**
- [ ] Testes de domínio em `internal/domain/{aggregate}/{aggregate}_test.go`
- [ ] Testes de repository em `infrastructure/persistence/{repo}_test.go`
- [ ] Testes de handler em `infrastructure/http/handlers/{handler}_test.go`

### **Event Bus**
- [ ] Verificar se evento está mapeado em `domain_event_bus.go`
  - [ ] Adicionar em `mapDomainToBusinessEvents()` se necessário

---

## ✅ CHECKLIST: CRIAR NOVO AGREGADO

### **1. Domain Layer**
```bash
mkdir -p internal/domain/{aggregate_name}
cd internal/domain/{aggregate_name}
touch {aggregate_name}.go events.go repository.go types.go value_objects.go errors.go
```

- [ ] `{aggregate_name}.go`:
  - [ ] Struct privada com campos
  - [ ] Factory method `New{Aggregate}()`
  - [ ] Métodos de negócio que emitem eventos
  - [ ] Getters públicos
  - [ ] DomainEvents() e ClearEvents()
  - [ ] Reconstruct{Aggregate}() para mapper

- [ ] `events.go`:
  - [ ] Structs de eventos (privados)
  - [ ] Factory methods New{Event}()
  - [ ] Implementar shared.DomainEvent

- [ ] `repository.go`:
  - [ ] Interface Repository
  - [ ] Métodos: Save, FindByID, FindBy...

- [ ] `types.go`:
  - [ ] Enums (Status, Type, etc.)
  - [ ] Constants

- [ ] `value_objects.go`:
  - [ ] Value Objects com validação
  - [ ] Immutáveis

- [ ] `errors.go`:
  - [ ] Sentinel errors (ErrXxxNotFound)
  - [ ] Factory methods (NewXxxNotFoundError)

### **2. Application Layer**
```bash
mkdir -p internal/application/commands/{aggregate_name}
mkdir -p internal/application/{aggregate_name}
touch internal/application/dtos/{aggregate_name}_dto.go
```

- [ ] Commands: Create, Update, Delete
- [ ] Queries: List, Search, Get
- [ ] Use Cases se necessário
- [ ] DTOs com mappers

### **3. Infrastructure Layer**
```bash
touch infrastructure/persistence/entities/{aggregate_name}.go
touch infrastructure/persistence/gorm_{aggregate_name}_repository.go
touch infrastructure/database/migrations/000XXX_add_{aggregate_name}s.go
touch infrastructure/http/handlers/{aggregate_name}_handler.go
```

- [ ] GORM Entity
- [ ] GORM Repository
- [ ] Migration GORM
- [ ] HTTP Handler
- [ ] Registrar rotas
- [ ] DI

### **4. Testes**
```bash
touch internal/domain/{aggregate_name}/{aggregate_name}_test.go
touch infrastructure/persistence/gorm_{aggregate_name}_repository_test.go
```

---

## 📐 PADRÕES E CONVENÇÕES

### **Nomenclatura**

| Elemento | Padrão | Exemplo |
|----------|--------|---------|
| Agregado | PascalCase | `Contact`, `Session` |
| Arquivo Go | snake_case | `contact.go`, `events.go` |
| Package | lowercase | `contact`, `session` |
| DTO | {Entity}DTO | `ContactDTO`, `SessionDTO` |
| Entity GORM | {Entity}Entity | `ContactEntity` |
| Repository | Gorm{Entity}Repository | `GormContactRepository` |
| Handler | {Entity}Handler | `ContactHandler` |
| Command | {Action}{Entity}Command | `CreateContactCommand` |
| Query | {Action}{Entity}Query | `ListContactsQuery` |
| Evento | {Entity}{Action}Event | `ContactCreatedEvent` |
| Evento (nome) | {resource}.{action} | `contact.created` |
| Migration | Migration{Number}{Description} | `Migration000001InitialSchema` |

### **Estrutura de Erros**

```go
// Domain errors (retornar para Application)
return nil, errors.New("validation failed")
return nil, contact.NewContactNotFoundError(id)

// Application errors (retornar para Infrastructure)
return nil, errors.New("contact not found")

// Infrastructure (converter para HTTP)
status, apiErr := MapDomainErrorToHTTP(err)
c.JSON(status, NewErrorResponse(apiErr.Code, apiErr.Message))
```

### **Response Padrão HTTP**

```go
// Success
c.JSON(200, APIResponse{
    Data: dto,
    Meta: &ResponseMeta{Page: 1, Total: 100},
    Links: &ResponseLinks{Next: "/api/v1/contacts?page=2"},
})

// Error
c.JSON(400, APIResponse{
    Error: &APIError{
        Code: "VALIDATION_FAILED",
        Message: "Name is required",
        Field: "name",
    },
})
```

### **Paginação Padrão**

- `?page=1` (default: 1)
- `?limit=20` (default: 20, max: 100)
- `?sort_by=created_at` (default: created_at)
- `?sort_dir=desc` (default: desc)

---

## 🧪 TESTES

### **Estrutura de Testes**

```
internal/domain/{aggregate}/{aggregate}_test.go        # Testes unitários
infrastructure/persistence/{repo}_test.go              # Testes integração
infrastructure/http/handlers/{handler}_test.go         # Testes HTTP
tests/e2e/                                             # Testes E2E
```

### **Convenções**

- Usar `testify/assert` e `testify/require`
- Testes de domínio: sem dependências externas
- Testes de repository: usar testcontainers (PostgreSQL)
- Testes de handler: usar httptest

### **Executar Testes**

```bash
# Todos
go test ./...

# Específico
go test ./internal/domain/contact

# Com coverage
go test -cover ./internal/domain/...

# Verbose
go test -v ./...
```

---

## 🚀 DEPLOY E CI/CD

### **Build**

```bash
# Build local
go build -o ventros-api cmd/api/main.go

# Build para produção
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ventros-api cmd/api/main.go
```

### **Docker**

```bash
# Build
docker build -t ventros-crm:latest .

# Run
docker run -p 8080:8080 ventros-crm:latest
```

### **Migrations**

```bash
# Aplicar migrations
go run cmd/migrate/main.go up

# Rollback última
go run cmd/migrate/main.go down

# Status
go run cmd/migrate/main.go status
```

---

## 📚 REFERÊNCIAS

- **DDD**: Evans, Eric. Domain-Driven Design
- **Hexagonal Architecture**: https://alistair.cockburn.us/hexagonal-architecture/
- **CQRS**: https://martinfowler.com/bliki/CQRS.html
- **Event Sourcing**: https://martinfowler.com/eaaDev/EventSourcing.html
- **Outbox Pattern**: https://microservices.io/patterns/data/transactional-outbox.html
- **Saga Pattern**: https://microservices.io/patterns/data/saga.html

---

**Este guia é um documento vivo. Atualize conforme o projeto evolui!**
