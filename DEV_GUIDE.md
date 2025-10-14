# Developer Guide - Ventros CRM

**Guia Completo para Implementação de Features**

**Versão**: 1.0
**Última Atualização**: 2025-10-12
**Status**: ✅ Production-Ready

---

## 📋 Índice

1. [Visão Geral da Arquitetura](#visão-geral-da-arquitetura)
2. [Checklist Completo para Nova Feature](#checklist-completo-para-nova-feature)
3. [Estrutura de Camadas DDD](#estrutura-de-camadas-ddd)
4. [Padrões de Código](#padrões-de-código)
5. [Estratégia de Testes](#estratégia-de-testes)
6. [Documentação Obrigatória](#documentação-obrigatória)
7. [Referências Importantes](#referências-importantes)

---

## 🏗️ Visão Geral da Arquitetura

### Stack Tecnológico

```
Go 1.25.1+
├── Framework HTTP: Gin
├── ORM: GORM
├── Database: PostgreSQL 15+ (Row-Level Security)
├── Message Broker: RabbitMQ 3.12+ (Outbox Pattern)
├── Cache: Redis 7.0+
└── Workflows: Temporal
```

### Padrões Arquiteturais

```
✅ Domain-Driven Design (DDD)
✅ Hexagonal Architecture (Ports & Adapters)
✅ Event-Driven Architecture (104+ domain events)
✅ CQRS (Command Handler Pattern)
✅ Outbox Pattern (Transactional events)
✅ Saga Pattern (Orchestration + Choreography)
✅ Multi-tenancy (tenant_id + RLS)
```

### Estrutura de Diretórios

```
ventros-crm/
├── cmd/                          # Entry points
│   ├── api/                      # API server
│   └── migrate/                  # Migration CLI
│
├── internal/                     # Private application code
│   ├── domain/                   # DOMAIN LAYER (Pure business logic)
│   │   ├── crm/                  # CRM bounded context
│   │   │   ├── contact/          # Contact aggregate
│   │   │   ├── session/          # Session aggregate
│   │   │   ├── message/          # Message aggregate
│   │   │   ├── channel/          # Channel aggregate
│   │   │   ├── pipeline/         # Pipeline aggregate
│   │   │   ├── agent/            # Agent aggregate
│   │   │   └── chat/             # Chat aggregate
│   │   ├── automation/           # Automation bounded context
│   │   │   ├── campaign/         # Campaign aggregate
│   │   │   └── sequence/         # Sequence aggregate
│   │   └── core/                 # Core bounded context
│   │       ├── billing/          # Billing aggregate
│   │       ├── project/          # Project aggregate
│   │       └── shared/           # Shared domain primitives
│   │
│   ├── application/              # APPLICATION LAYER (Use cases)
│   │   ├── commands/             # Write operations (CQRS)
│   │   │   ├── contact/          # Contact commands
│   │   │   ├── session/          # Session commands
│   │   │   ├── message/          # Message commands
│   │   │   └── ...
│   │   └── queries/              # Read operations (CQRS)
│   │
│   └── infrastructure/           # INFRASTRUCTURE LAYER (External concerns)
│       ├── http/                 # HTTP handlers (Presentation)
│       │   ├── handlers/         # Gin handlers
│       │   ├── middleware/       # HTTP middleware
│       │   └── routes/           # Route definitions
│       ├── persistence/          # Database (Repositories)
│       │   ├── entities/         # GORM entities
│       │   └── gorm_*_repository.go
│       ├── messaging/            # RabbitMQ (Event Bus)
│       ├── cache/                # Redis
│       ├── workflow/             # Temporal
│       └── channels/             # External integrations (WAHA, etc)
│
├── guides/                       # Documentation
│   ├── domain_mapping/           # 23 aggregate docs
│   ├── MAKEFILE.md               # Development commands
│   └── ACTORS.md                 # System actors
│
├── migrations/                   # SQL migrations
├── P0.md                         # Handler refactoring (DONE)
├── AI_REPORT.md                  # Architectural audit
└── TODO.md                       # Roadmap
```

**Documentação Completa**:
- [Arquitecture Overview](AI_REPORT.md) - Rating 8.2/10
- [23 Domain Aggregates](guides/domain_mapping/) - Complete DDD mapping
- [Handler Pattern](P0.md) - Command Handler implementation (100% done)

---

## ✅ Checklist Completo para Nova Feature

Use este checklist **SEMPRE** que implementar uma nova feature:

### **Fase 1: Análise e Design** 📐

- [ ] **1.1. Entender o Problema de Negócio**
  - [ ] Qual problema estamos resolvendo?
  - [ ] Quem são os atores envolvidos? (ver [ACTORS.md](guides/ACTORS.md))
  - [ ] Qual o fluxo de negócio completo?
  - [ ] Há integrações externas envolvidas?

- [ ] **1.2. Identificar o Bounded Context**
  - [ ] CRM (Contact, Session, Message, Channel, Pipeline, Agent, Chat)?
  - [ ] Automation (Campaign, Sequence)?
  - [ ] Core (Billing, Project, Customer)?
  - [ ] Novo bounded context? (criar estrutura completa)

- [ ] **1.3. Identificar o Aggregate**
  - [ ] Qual aggregate é responsável? (ver [guides/domain_mapping/](guides/domain_mapping/))
  - [ ] Precisa criar novo aggregate?
  - [ ] Quais invariantes de negócio devem ser protegidas?
  - [ ] Qual o aggregate root?

- [ ] **1.4. Definir Eventos de Domínio**
  - [ ] Quais eventos serão emitidos?
  - [ ] Nomenclatura: `aggregate.action` (ex: `contact.created`, `session.ended`)
  - [ ] Payload mínimo necessário
  - [ ] Quem consome esses eventos?

---

### **Fase 2: Domain Layer** 🎯

**Localização**: `internal/domain/{bounded_context}/{aggregate}/`

- [ ] **2.1. Criar/Atualizar Aggregate Root**
  ```go
  // internal/domain/crm/contact/contact.go

  type Contact struct {
      id        uuid.UUID    // Sempre privado
      version   int          // ✅ Optimistic locking obrigatório
      projectID uuid.UUID    // Multi-tenancy obrigatório
      tenantID  string       // Multi-tenancy obrigatório

      // Business fields (privados)
      name      string
      email     *Email       // Value Object
      phone     *Phone       // Value Object

      // Audit fields
      createdAt time.Time
      updatedAt time.Time
      deletedAt *time.Time   // Soft delete

      // Event sourcing
      events    []DomainEvent
  }
  ```

- [ ] **2.2. Criar Value Objects** (se necessário)
  ```go
  // internal/domain/crm/contact/value_objects.go

  type Email struct {
      Value string
  }

  func NewEmail(value string) (Email, error) {
      // ✅ Validação no construtor
      if !isValidEmail(value) {
          return Email{}, ErrInvalidEmail
      }
      return Email{Value: value}, nil
  }
  ```

- [ ] **2.3. Implementar Business Methods**
  ```go
  // ✅ Factory method (construtor)
  func NewContact(projectID uuid.UUID, tenantID, name string) (*Contact, error) {
      // Validações de negócio
      if name == "" {
          return nil, ErrNameRequired
      }

      c := &Contact{
          id:        uuid.New(),
          version:   1, // ✅ Optimistic locking
          projectID: projectID,
          tenantID:  tenantID,
          name:      name,
          createdAt: time.Now(),
          updatedAt: time.Now(),
          events:    []DomainEvent{},
      }

      // ✅ Emitir evento
      c.addEvent(NewContactCreatedEvent(c))

      return c, nil
  }

  // ✅ Business methods (não setters genéricos!)
  func (c *Contact) UpdateName(newName string) error {
      if newName == "" {
          return ErrNameRequired
      }

      oldName := c.name
      c.name = newName
      c.updatedAt = time.Now()

      c.addEvent(NewContactNameChangedEvent(c, oldName, newName))

      return nil
  }

  // ✅ Getters públicos
  func (c *Contact) ID() uuid.UUID       { return c.id }
  func (c *Contact) Version() int        { return c.version }
  func (c *Contact) Name() string        { return c.name }
  func (c *Contact) DomainEvents() []DomainEvent { return c.events }
  ```

- [ ] **2.4. Definir Domain Events**
  ```go
  // internal/domain/crm/contact/events.go

  type ContactCreatedEvent struct {
      ContactID uuid.UUID
      ProjectID uuid.UUID
      TenantID  string
      Name      string
      EventMeta EventMetadata
  }

  func (e ContactCreatedEvent) EventType() string {
      return "contact.created" // ✅ Nomenclatura padrão
  }
  ```

- [ ] **2.5. Definir Repository Interface**
  ```go
  // internal/domain/crm/contact/repository.go

  type Repository interface {
      Save(ctx context.Context, contact *Contact) error
      FindByID(ctx context.Context, id uuid.UUID) (*Contact, error)
      // ... outros métodos
  }
  ```

- [ ] **2.6. Definir Errors**
  ```go
  // internal/domain/crm/contact/errors.go

  var (
      ErrContactNotFound = errors.New("contact not found")
      ErrNameRequired    = errors.New("name is required")
  )

  func NewContactNotFoundError(id string) *shared.DomainError {
      err := shared.NewNotFoundError("contact", id)
      err.Err = ErrContactNotFound // ✅ Wrap sentinel error
      return err
  }
  ```

- [ ] **2.7. Testes Unitários de Domínio**
  ```go
  // internal/domain/crm/contact/contact_test.go

  func TestNewContact(t *testing.T) {
      // Arrange
      projectID := uuid.New()
      tenantID := "tenant-123"
      name := "John Doe"

      // Act
      contact, err := NewContact(projectID, tenantID, name)

      // Assert
      assert.NoError(t, err)
      assert.NotNil(t, contact)
      assert.Equal(t, name, contact.Name())
      assert.Equal(t, 1, contact.Version()) // ✅ Optimistic locking
      assert.Len(t, contact.DomainEvents(), 1) // ✅ Event emitted
      assert.Equal(t, "contact.created", contact.DomainEvents()[0].EventType())
  }

  func TestUpdateName_EmptyName(t *testing.T) {
      // Arrange
      contact, _ := NewContact(uuid.New(), "tenant-123", "John Doe")

      // Act
      err := contact.UpdateName("")

      // Assert
      assert.Error(t, err)
      assert.Equal(t, ErrNameRequired, err)
  }
  ```

**✅ Meta**: 100% cobertura de testes unitários no domain layer

---

### **Fase 3: Application Layer** 🎨

**Localização**: `internal/application/commands/{aggregate}/`

- [ ] **3.1. Criar Command Struct**
  ```go
  // internal/application/commands/contact/create_contact_command.go

  package contact

  import "github.com/google/uuid"

  type CreateContactCommand struct {
      ProjectID uuid.UUID
      TenantID  string
      Name      string
      Email     *string
      Phone     *string
      Tags      []string
  }

  // ✅ Validação no command
  func (c *CreateContactCommand) Validate() error {
      if c.ProjectID == uuid.Nil {
          return ErrProjectIDRequired
      }
      if c.TenantID == "" {
          return ErrTenantIDRequired
      }
      if c.Name == "" {
          return ErrNameRequired
      }
      return nil
  }
  ```

- [ ] **3.2. Criar Command Handler**
  ```go
  // internal/application/commands/contact/create_contact_handler.go

  package contact

  import (
      "context"

      domainContact "github.com/ventros/crm/internal/domain/crm/contact"
      "github.com/ventros/crm/internal/domain/core/shared"
  )

  type CreateContactHandler struct {
      contactRepo domainContact.Repository
      eventBus    shared.EventBus
  }

  func NewCreateContactHandler(
      contactRepo domainContact.Repository,
      eventBus shared.EventBus,
  ) *CreateContactHandler {
      return &CreateContactHandler{
          contactRepo: contactRepo,
          eventBus:    eventBus,
      }
  }

  func (h *CreateContactHandler) Handle(
      ctx context.Context,
      cmd CreateContactCommand,
  ) (*domainContact.Contact, error) {
      // 1. ✅ Validar command
      if err := cmd.Validate(); err != nil {
          return nil, err
      }

      // 2. ✅ Criar domain aggregate
      contact, err := domainContact.NewContact(
          cmd.ProjectID,
          cmd.TenantID,
          cmd.Name,
      )
      if err != nil {
          return nil, err
      }

      // 3. ✅ Aplicar operações adicionais
      if cmd.Email != nil && *cmd.Email != "" {
          if err := contact.SetEmail(*cmd.Email); err != nil {
              return nil, err
          }
      }

      if cmd.Phone != nil && *cmd.Phone != "" {
          if err := contact.SetPhone(*cmd.Phone); err != nil {
              return nil, err
          }
      }

      // 4. ✅ Persistir (com optimistic locking)
      if err := h.contactRepo.Save(ctx, contact); err != nil {
          return nil, err
      }

      // 5. ✅ Publicar eventos (via Outbox Pattern)
      for _, event := range contact.DomainEvents() {
          if err := h.eventBus.Publish(ctx, event); err != nil {
              // Log error but don't fail (Outbox guarantees eventual delivery)
          }
      }

      return contact, nil
  }
  ```

- [ ] **3.3. Criar Errors Específicos**
  ```go
  // internal/application/commands/contact/errors.go

  package contact

  import "errors"

  var (
      ErrProjectIDRequired = errors.New("project ID is required")
      ErrTenantIDRequired  = errors.New("tenant ID is required")
      ErrNameRequired      = errors.New("name is required")
  )
  ```

- [ ] **3.4. Testes Unitários do Command Handler**
  ```go
  // internal/application/commands/contact/create_contact_handler_test.go

  package contact_test

  func TestCreateContactHandler_Handle(t *testing.T) {
      t.Run("should create contact successfully", func(t *testing.T) {
          // Arrange
          mockRepo := new(MockContactRepository)
          mockEventBus := new(MockEventBus)
          handler := NewCreateContactHandler(mockRepo, mockEventBus)

          cmd := CreateContactCommand{
              ProjectID: uuid.New(),
              TenantID:  "tenant-123",
              Name:      "John Doe",
          }

          mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
          mockEventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)

          // Act
          contact, err := handler.Handle(context.Background(), cmd)

          // Assert
          assert.NoError(t, err)
          assert.NotNil(t, contact)
          assert.Equal(t, "John Doe", contact.Name())
          mockRepo.AssertExpectations(t)
          mockEventBus.AssertExpectations(t)
      })

      t.Run("should return error when name is empty", func(t *testing.T) {
          // ... test validation
      })
  }
  ```

**✅ Meta**: 80%+ cobertura de testes no application layer

---

### **Fase 4: Infrastructure Layer** 🔧

#### **4.1. Persistence (Repository Implementation)**

**Localização**: `infrastructure/persistence/`

- [ ] **4.1.1. Criar GORM Entity**
  ```go
  // infrastructure/persistence/entities/contact_entity.go

  package entities

  import (
      "time"
      "github.com/google/uuid"
  )

  type ContactEntity struct {
      ID        uuid.UUID  `gorm:"type:uuid;primary_key"`
      Version   int        `gorm:"not null;default:1"` // ✅ Optimistic locking
      ProjectID uuid.UUID  `gorm:"type:uuid;not null;index"`
      TenantID  string     `gorm:"type:text;not null;index"`

      Name  string  `gorm:"type:text;not null"`
      Email *string `gorm:"type:text"`
      Phone *string `gorm:"type:text;index"`

      CreatedAt time.Time
      UpdatedAt time.Time
      DeletedAt *time.Time `gorm:"index"` // ✅ Soft delete
  }

  func (ContactEntity) TableName() string {
      return "contacts"
  }
  ```

- [ ] **4.1.2. Criar Repository Implementation**
  ```go
  // infrastructure/persistence/gorm_contact_repository.go

  package persistence

  import (
      "context"
      "errors"

      "gorm.io/gorm"
      domainContact "github.com/ventros/crm/internal/domain/crm/contact"
      "github.com/ventros/crm/infrastructure/persistence/entities"
  )

  type GormContactRepository struct {
      db *gorm.DB
  }

  func NewGormContactRepository(db *gorm.DB) *GormContactRepository {
      return &GormContactRepository{db: db}
  }

  // ✅ Save with optimistic locking
  func (r *GormContactRepository) Save(
      ctx context.Context,
      contact *domainContact.Contact,
  ) error {
      entity := r.toEntity(contact)

      if contact.Version() == 1 {
          // INSERT (new record)
          return r.db.WithContext(ctx).Create(entity).Error
      }

      // UPDATE (existing record with optimistic locking)
      result := r.db.WithContext(ctx).
          Model(&entities.ContactEntity{}).
          Where("id = ? AND version = ?", entity.ID, contact.Version()).
          Updates(map[string]interface{}{
              "name":       entity.Name,
              "email":      entity.Email,
              "phone":      entity.Phone,
              "version":    contact.Version() + 1, // ✅ Increment version
              "updated_at": entity.UpdatedAt,
          })

      if result.Error != nil {
          return result.Error
      }

      if result.RowsAffected == 0 {
          return domainContact.ErrConcurrentUpdateConflict
      }

      return nil
  }

  func (r *GormContactRepository) FindByID(
      ctx context.Context,
      id uuid.UUID,
  ) (*domainContact.Contact, error) {
      var entity entities.ContactEntity

      err := r.db.WithContext(ctx).
          Where("id = ? AND deleted_at IS NULL", id).
          First(&entity).Error

      if err != nil {
          if errors.Is(err, gorm.ErrRecordNotFound) {
              return nil, domainContact.NewContactNotFoundError(id.String())
          }
          return nil, err
      }

      return r.toDomain(&entity)
  }

  // ✅ Mapper: Entity → Domain
  func (r *GormContactRepository) toDomain(
      entity *entities.ContactEntity,
  ) (*domainContact.Contact, error) {
      // Use reflection or constructor to rebuild domain aggregate
      // ...
  }

  // ✅ Mapper: Domain → Entity
  func (r *GormContactRepository) toEntity(
      contact *domainContact.Contact,
  ) *entities.ContactEntity {
      return &entities.ContactEntity{
          ID:        contact.ID(),
          Version:   contact.Version(),
          ProjectID: contact.ProjectID(),
          TenantID:  contact.TenantID(),
          Name:      contact.Name(),
          // ...
      }
  }
  ```

- [ ] **4.1.3. Testes de Integração do Repository**
  ```go
  // infrastructure/persistence/gorm_contact_repository_test.go

  package persistence_test

  func TestGormContactRepository_Save(t *testing.T) {
      // ✅ Setup: Real database (testcontainers)
      db := setupTestDatabase(t)
      repo := NewGormContactRepository(db)

      t.Run("should save new contact", func(t *testing.T) {
          // ...
      })

      t.Run("should detect concurrent update conflict", func(t *testing.T) {
          // Test optimistic locking
      })
  }
  ```

**✅ Executar**: `make test-integration`

---

#### **4.2. HTTP Layer (Presentation)**

**Localização**: `infrastructure/http/handlers/`

- [ ] **4.2.1. Criar HTTP Handler**
  ```go
  // infrastructure/http/handlers/contact_handler.go

  package handlers

  import (
      "net/http"

      "github.com/gin-gonic/gin"
      "github.com/ventros/crm/internal/application/commands/contact"
  )

  type ContactHandler struct {
      createContactHandler *contact.CreateContactHandler
      // ... outros handlers
  }

  func NewContactHandler(
      createContactHandler *contact.CreateContactHandler,
  ) *ContactHandler {
      return &ContactHandler{
          createContactHandler: createContactHandler,
      }
  }

  // ✅ Handler HTTP (apenas adaptador!)
  // @Summary Create a new contact
  // @Tags Contacts
  // @Accept json
  // @Produce json
  // @Param request body CreateContactRequest true "Contact data"
  // @Success 201 {object} ContactResponse
  // @Failure 400 {object} ErrorResponse
  // @Router /api/v1/contacts [post]
  func (h *ContactHandler) CreateContact(c *gin.Context) {
      // 1. ✅ Parse request
      var req CreateContactRequest
      if err := c.ShouldBindJSON(&req); err != nil {
          c.JSON(http.StatusBadRequest, ErrorResponse{
              Error: "Invalid request body",
          })
          return
      }

      // 2. ✅ Extract auth context
      projectID := c.GetString("project_id") // from JWT
      tenantID := c.GetString("tenant_id")   // from JWT

      // 3. ✅ Build command
      cmd := contact.CreateContactCommand{
          ProjectID: uuid.MustParse(projectID),
          TenantID:  tenantID,
          Name:      req.Name,
          Email:     req.Email,
          Phone:     req.Phone,
          Tags:      req.Tags,
      }

      // 4. ✅ Delegate to command handler
      domainContact, err := h.createContactHandler.Handle(c.Request.Context(), cmd)
      if err != nil {
          // Map domain errors to HTTP status codes
          c.JSON(mapErrorToHTTPStatus(err), ErrorResponse{
              Error: err.Error(),
          })
          return
      }

      // 5. ✅ Convert to response DTO
      response := h.toResponse(domainContact)

      c.JSON(http.StatusCreated, response)
  }

  // ✅ DTOs (Request/Response)
  type CreateContactRequest struct {
      Name  string   `json:"name" binding:"required"`
      Email *string  `json:"email,omitempty"`
      Phone *string  `json:"phone,omitempty"`
      Tags  []string `json:"tags,omitempty"`
  }

  type ContactResponse struct {
      ID        string    `json:"id"`
      Name      string    `json:"name"`
      Email     *string   `json:"email,omitempty"`
      Phone     *string   `json:"phone,omitempty"`
      Tags      []string  `json:"tags,omitempty"`
      CreatedAt time.Time `json:"created_at"`
  }

  func (h *ContactHandler) toResponse(
      contact *domainContact.Contact,
  ) ContactResponse {
      return ContactResponse{
          ID:        contact.ID().String(),
          Name:      contact.Name(),
          CreatedAt: contact.CreatedAt(),
          // ...
      }
  }
  ```

- [ ] **4.2.2. Registrar Rotas**
  ```go
  // infrastructure/http/routes/contact_routes.go

  package routes

  func RegisterContactRoutes(
      router *gin.RouterGroup,
      handler *handlers.ContactHandler,
      authMiddleware gin.HandlerFunc,
  ) {
      contacts := router.Group("/contacts")
      contacts.Use(authMiddleware) // ✅ Authentication required
      {
          contacts.POST("", handler.CreateContact)
          contacts.GET("/:id", handler.GetContact)
          contacts.PUT("/:id", handler.UpdateContact)
          contacts.DELETE("/:id", handler.DeleteContact)
      }
  }
  ```

- [ ] **4.2.3. Testes E2E**
  ```go
  // tests/e2e/contact_test.go

  package e2e_test

  func TestCreateContact_E2E(t *testing.T) {
      // ✅ Setup: Full stack running
      apiURL := "http://localhost:8080"

      t.Run("should create contact via API", func(t *testing.T) {
          // Arrange
          token := getAuthToken(t) // Helper function

          payload := map[string]interface{}{
              "name":  "John Doe",
              "email": "john@example.com",
          }

          // Act
          resp, err := httpPost(apiURL+"/api/v1/contacts", token, payload)

          // Assert
          assert.NoError(t, err)
          assert.Equal(t, 201, resp.StatusCode)

          var result map[string]interface{}
          json.NewDecoder(resp.Body).Decode(&result)
          assert.Equal(t, "John Doe", result["name"])
      })
  }
  ```

**✅ Executar**: `make test-e2e` (requer `make infra` + `make api`)

---

#### **4.3. Database Migration**

**Localização**: `infrastructure/database/migrations/`

- [ ] **4.3.1. Criar Migration UP**
  ```sql
  -- infrastructure/database/migrations/000050_create_contacts.up.sql

  CREATE TABLE IF NOT EXISTS contacts (
      id UUID PRIMARY KEY,
      version INTEGER NOT NULL DEFAULT 1,         -- ✅ Optimistic locking
      project_id UUID NOT NULL,
      tenant_id TEXT NOT NULL,

      name TEXT NOT NULL,
      email TEXT,
      phone TEXT,

      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
      deleted_at TIMESTAMP,                       -- ✅ Soft delete

      -- ✅ Foreign keys
      CONSTRAINT fk_contacts_project FOREIGN KEY (project_id)
          REFERENCES projects(id) ON DELETE CASCADE,

      -- ✅ Indexes
      CONSTRAINT contacts_name_not_empty CHECK (name <> '')
  );

  -- ✅ Indexes for performance
  CREATE INDEX idx_contacts_project ON contacts(project_id);
  CREATE INDEX idx_contacts_tenant ON contacts(tenant_id);
  CREATE INDEX idx_contacts_phone ON contacts(phone) WHERE phone IS NOT NULL;
  CREATE INDEX idx_contacts_email ON contacts(email) WHERE email IS NOT NULL;
  CREATE INDEX idx_contacts_deleted ON contacts(deleted_at) WHERE deleted_at IS NULL;

  -- ✅ Row-Level Security (Multi-tenancy)
  ALTER TABLE contacts ENABLE ROW LEVEL SECURITY;

  CREATE POLICY contacts_tenant_isolation ON contacts
      USING (tenant_id = current_setting('app.current_tenant')::TEXT);
  ```

- [ ] **4.3.2. Criar Migration DOWN**
  ```sql
  -- infrastructure/database/migrations/000050_create_contacts.down.sql

  DROP POLICY IF EXISTS contacts_tenant_isolation ON contacts;
  DROP TABLE IF EXISTS contacts CASCADE;
  ```

- [ ] **4.3.3. Executar Migration**
  ```bash
  # Aplica migration
  make migrate-up

  # Ou manualmente:
  go run cmd/migrate/main.go up
  ```

**✅ Ver**: [MIGRATIONS.md](MIGRATIONS.md) para guia completo

---

### **Fase 5: Event Publishing & Consumers** 📢

#### **5.1. Configurar Outbox Pattern**

**Já implementado!** Events são publicados via:
- `EventBus.Publish()` → Salva em `outbox_events` table
- `OutboxWorker` → Processa e publica no RabbitMQ
- PostgreSQL NOTIFY → Notifica worker (latency <100ms)

- [ ] **5.1.1. Verificar Event Naming**
  ```go
  // ✅ Padrão: aggregate.action
  "contact.created"
  "contact.updated"
  "contact.deleted"
  "session.started"
  "session.ended"
  "message.sent"
  ```

- [ ] **5.1.2. Definir Event Payload**
  ```go
  type ContactCreatedEvent struct {
      ContactID uuid.UUID `json:"contact_id"`
      ProjectID uuid.UUID `json:"project_id"`
      TenantID  string    `json:"tenant_id"`
      Name      string    `json:"name"`
      Email     *string   `json:"email,omitempty"`

      EventMeta EventMetadata `json:"_meta"`
  }
  ```

#### **5.2. Criar Event Consumer** (se necessário)

**Localização**: `infrastructure/messaging/consumers/`

- [ ] **5.2.1. Criar Consumer**
  ```go
  // infrastructure/messaging/consumers/contact_created_consumer.go

  package consumers

  import (
      "context"
      "encoding/json"

      "github.com/ventros/crm/internal/domain/crm/contact"
  )

  type ContactCreatedConsumer struct {
      // Dependencies (use cases, services, etc)
  }

  func (c *ContactCreatedConsumer) Handle(
      ctx context.Context,
      event contact.ContactCreatedEvent,
  ) error {
      // 1. ✅ Idempotency check (processed_events table)
      if c.isProcessed(ctx, event.EventMeta.EventID) {
          return nil // Already processed
      }

      // 2. ✅ Business logic
      // Example: Send welcome email, create CRM record, etc

      // 3. ✅ Mark as processed
      c.markAsProcessed(ctx, event.EventMeta.EventID, "contact_created_consumer")

      return nil
  }
  ```

- [ ] **5.2.2. Registrar Consumer no RabbitMQ**
  ```go
  // cmd/api/main.go

  messagingClient.Subscribe(
      "domain_events",              // Exchange
      "contact.created",            // Routing key
      "contact_created_consumer",   // Queue name
      contactCreatedConsumer.Handle,
  )
  ```

**✅ Ver**: Outbox Pattern implementado em `infrastructure/messaging/outbox/`

---

### **Fase 6: Documentação** 📚

- [ ] **6.1. Swagger Documentation**
  ```go
  // ✅ Adicionar comentários Swagger no handler

  // @Summary Create a new contact
  // @Description Creates a new contact in the CRM system
  // @Tags Contacts
  // @Accept json
  // @Produce json
  // @Param request body CreateContactRequest true "Contact data"
  // @Success 201 {object} ContactResponse "Contact created successfully"
  // @Failure 400 {object} ErrorResponse "Invalid request"
  // @Failure 401 {object} ErrorResponse "Unauthorized"
  // @Security BearerAuth
  // @Router /api/v1/contacts [post]
  func (h *ContactHandler) CreateContact(c *gin.Context) {
      // ...
  }
  ```

- [ ] **6.2. Atualizar Domain Aggregate Doc**
  ```markdown
  <!-- guides/domain_mapping/contact_aggregate.md -->

  # Contact Aggregate

  ## Commands
  - ✅ CreateContactCommand - Creates new contact
  - ✅ UpdateContactCommand - Updates contact details
  - ✅ DeleteContactCommand - Soft deletes contact

  ## Events
  - contact.created
  - contact.updated
  - contact.deleted

  ## Use Cases
  - CreateContactUseCase (internal/application/commands/contact/)
  - UpdateContactUseCase (internal/application/commands/contact/)
  ```

- [ ] **6.3. Atualizar TODO.md** (se feature grande)
  ```markdown
  ## ✅ RECENTLY COMPLETED FEATURES

  ### **★ Contact Management** ✅ COMPLETED (2025-10-12)
  - CRUD operations
  - Optimistic locking
  - Event publishing via Outbox
  ```

- [ ] **6.4. Atualizar README.md** (se feature importante)

---

## 🧪 Estratégia de Testes

### Test Pyramid (Mike Cohn, 2009)

```
           /\
          /E2E\      ← 10% (5 tests) - Full stack integration
         /------\
        /Integ. \   ← 20% (2 tests) - Database, external services
       /----------\
      /   Unit    \  ← 70% (61 tests) - Domain + Application logic
     /______________\
```

### Tipos de Teste

#### **1. Unit Tests** (70% - Fast)

**Onde**: `*_test.go` no mesmo package

**O que testar**:
- ✅ Domain layer (aggregates, value objects, business rules)
- ✅ Application layer (command handlers, validators)
- ✅ Pure functions

**Como executar**:
```bash
make test-unit  # ~2 minutos, sem dependências externas
```

**Exemplo**:
```go
func TestContact_UpdateName(t *testing.T) {
    // Arrange
    contact, _ := NewContact(uuid.New(), "tenant", "John")

    // Act
    err := contact.UpdateName("Jane")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Jane", contact.Name())
}
```

---

#### **2. Integration Tests** (20% - Medium)

**Onde**: `infrastructure/persistence/*_test.go`

**O que testar**:
- ✅ Repository implementations (GORM + PostgreSQL)
- ✅ Database migrations
- ✅ External service integrations

**Como executar**:
```bash
make infra              # Inicia PostgreSQL, RabbitMQ, Redis
make test-integration   # ~10 minutos
```

**Exemplo**:
```go
func TestGormContactRepository_Save(t *testing.T) {
    // Setup: Real PostgreSQL database
    db := setupTestDatabase(t)
    repo := NewGormContactRepository(db)

    // Test
    contact, _ := contact.NewContact(uuid.New(), "tenant", "John")
    err := repo.Save(context.Background(), contact)

    assert.NoError(t, err)
}
```

---

#### **3. E2E Tests** (10% - Slow)

**Onde**: `tests/e2e/*_test.go`

**O que testar**:
- ✅ HTTP endpoints completos
- ✅ Fluxos de negócio end-to-end
- ✅ Integrações completas (API → DB → RabbitMQ → Worker)

**Como executar**:
```bash
make infra  # Inicia infraestrutura
make api    # Inicia API em outra janela
make test-e2e  # ~10 minutos
```

**Exemplo**:
```go
func TestCreateContact_E2E(t *testing.T) {
    // Full HTTP request to running API
    token := getAuthToken(t)
    resp := httpPost("http://localhost:8080/api/v1/contacts", token, payload)

    assert.Equal(t, 201, resp.StatusCode)
}
```

---

### Coverage Goals

- ✅ **Domain Layer**: 100% coverage (business-critical)
- ✅ **Application Layer**: 80%+ coverage
- ✅ **Infrastructure Layer**: 60%+ coverage
- ✅ **Overall**: 82%+ coverage

**Verificar cobertura**:
```bash
make test-coverage  # Gera relatório HTML
```

**Ver mais**: [guides/TESTING.md](guides/TESTING.md)

---

## 📖 Padrões de Código

### Naming Conventions

```go
// ✅ Aggregates: PascalCase singular
Contact, Session, Message, Pipeline

// ✅ Commands: {Action}{Aggregate}Command
CreateContactCommand, UpdateSessionCommand

// ✅ Command Handlers: {Action}{Aggregate}Handler
CreateContactHandler, UpdateSessionHandler

// ✅ Events: {aggregate}.{action} (lowercase)
"contact.created", "session.ended", "message.sent"

// ✅ Repository Interface: Repository
type Repository interface { ... }

// ✅ Repository Implementation: Gorm{Aggregate}Repository
GormContactRepository, GormSessionRepository

// ✅ Errors: Err{Description}
ErrContactNotFound, ErrInvalidEmail, ErrNameRequired

// ✅ DTOs: {Action}{Aggregate}Request/Response
CreateContactRequest, ContactResponse

// ✅ Handlers: {Aggregate}Handler
ContactHandler, SessionHandler

// ✅ Value Objects: PascalCase
Email, Phone, Money, HexColor
```

---

### Error Handling

```go
// ✅ Domain errors (sentinel errors)
var (
    ErrContactNotFound = errors.New("contact not found")
    ErrInvalidEmail    = errors.New("invalid email format")
)

// ✅ Wrapped errors (com contexto)
func NewContactNotFoundError(id string) *shared.DomainError {
    err := shared.NewNotFoundError("contact", id)
    err.Err = ErrContactNotFound // Wrap sentinel
    return err
}

// ✅ Usar errors.Is() para comparação
if errors.Is(err, contact.ErrContactNotFound) {
    // Handle not found
}

// ✅ Mapear domain errors para HTTP status
func mapErrorToHTTPStatus(err error) int {
    switch {
    case errors.Is(err, contact.ErrContactNotFound):
        return http.StatusNotFound
    case errors.Is(err, contact.ErrInvalidEmail):
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}
```

---

### Event Naming

```go
// ✅ Formato: {aggregate}.{action} (past tense)
"contact.created"
"contact.updated"
"contact.deleted"
"session.started"
"session.ended"
"message.sent"
"campaign.activated"

// ❌ Evitar:
"create_contact"  // Wrong format
"ContactCreated"  // Wrong casing
"contact_create"  // Wrong tense
```

---

### Optimistic Locking (Obrigatório!)

```go
// ✅ SEMPRE adicionar campo version
type Contact struct {
    id      uuid.UUID
    version int  // ✅ Starts at 1, increments on each update
    // ...
}

// ✅ Constructor inicia com version = 1
func NewContact(...) (*Contact, error) {
    return &Contact{
        id:      uuid.New(),
        version: 1,  // ✅
        // ...
    }, nil
}

// ✅ Repository Save verifica version
func (r *GormContactRepository) Save(ctx context.Context, c *Contact) error {
    result := r.db.
        Where("id = ? AND version = ?", c.ID(), c.Version()).
        Updates(map[string]interface{}{
            "name":    c.Name(),
            "version": c.Version() + 1,  // ✅ Increment
        })

    if result.RowsAffected == 0 {
        return ErrConcurrentUpdateConflict  // ✅ Conflict detected
    }

    return nil
}
```

**Por quê?** Previne lost updates em ambientes concorrentes.

**Ver mais**: [AI_REPORT.md - GAP 1](AI_REPORT.md#gap-1-optimistic-locking)

---

## 📚 Referências Importantes

### Documentação do Projeto

| Documento | Descrição | Quando Consultar |
|-----------|-----------|------------------|
| [README.md](README.md) | Visão geral do projeto | Início de qualquer feature |
| [TODO.md](TODO.md) | Roadmap e prioridades | Verificar escopo e próximas features |
| [P0.md](P0.md) | Handler refactoring (100% done) | Referência de padrão Command Handler |
| [AI_REPORT.md](AI_REPORT.md) | Auditoria arquitetural (8.2/10) | Entender qualidade e gaps |
| [MAKEFILE.md](MAKEFILE.md) | Comandos de desenvolvimento | Referência rápida de comandos |
| [MIGRATIONS.md](MIGRATIONS.md) | Guia de migrations SQL | Ao criar/modificar schema |
| [guides/MAKEFILE.md](guides/MAKEFILE.md) | Guia completo do Makefile | Comandos avançados |
| [guides/ACTORS.md](guides/ACTORS.md) | Atores do sistema | Entender permissões e capabilities |
| [guides/TESTING.md](guides/TESTING.md) | Estratégia de testes | Ao escrever testes |
| [guides/domain_mapping/](guides/domain_mapping/) | 23 Domain aggregates | Entender domínio antes de codificar |

---

### Domain Aggregates (23 Total)

**Core CRM** (5):
- [Contact](guides/domain_mapping/contact_aggregate.md)
- [Session](guides/domain_mapping/session_aggregate.md)
- [Message](guides/domain_mapping/message_aggregate.md)
- [Pipeline](guides/domain_mapping/pipeline_aggregate.md)
- [Agent](guides/domain_mapping/agent_aggregate.md)

**Communication** (3):
- [Channel](guides/domain_mapping/channel_aggregate.md)
- [ChannelType](guides/domain_mapping/channel_type_aggregate.md)
- [Broadcast](guides/domain_mapping/broadcast_aggregate.md)

**Analytics** (3):
- [Tracking](guides/domain_mapping/tracking_aggregate.md)
- [ContactEvent](guides/domain_mapping/contact_event_aggregate.md)
- [Event](guides/domain_mapping/event_aggregate.md)

**Auth & Multi-tenancy** (3):
- [Project](guides/domain_mapping/project_aggregate.md)
- [Customer](guides/domain_mapping/customer_aggregate.md)
- [Credential](guides/domain_mapping/credential_aggregate.md)

**Billing** (1):
- [Billing](guides/domain_mapping/billing_aggregate.md)

**Webhooks** (1):
- [Webhook](guides/domain_mapping/webhook_aggregate.md)

**Supporting** (4):
- [Note](guides/domain_mapping/note_aggregate.md)
- [ContactList](guides/domain_mapping/contact_list_aggregate.md)
- [AgentSession](guides/domain_mapping/agent_session_aggregate.md)
- [Saga](guides/domain_mapping/saga_aggregate.md)

**NEW** (1):
- [Chat](guides/domain_mapping/chat_aggregate.md) - CRITICAL

---

### Referências Externas

**Padrões Arquiteturais**:
- [Domain-Driven Design](https://martinfowler.com/tags/domain%20driven%20design.html) - Martin Fowler
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) - Alistair Cockburn
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) - Uncle Bob
- [CQRS Pattern](https://martinfowler.com/bliki/CQRS.html) - Martin Fowler
- [Event-Driven Architecture](https://martinfowler.com/articles/201701-event-driven.html) - Martin Fowler
- [Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html) - Chris Richardson
- [Saga Pattern](https://microservices.io/patterns/data/saga.html) - Chris Richardson

**Testing**:
- [Test Pyramid](https://martinfowler.com/bliki/TestPyramid.html) - Martin Fowler
- [Testing Best Practices](https://github.com/goldbergyoni/javascript-testing-best-practices) - Yoni Goldberg

**Go Best Practices**:
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

## 🎯 Checklist Final

Antes de abrir Pull Request:

### Code Quality
- [ ] ✅ Código segue padrões de nomenclatura
- [ ] ✅ Todas as camadas DDD implementadas
- [ ] ✅ Optimistic locking adicionado (se aggregate)
- [ ] ✅ Soft delete implementado (se necessário)
- [ ] ✅ Multi-tenancy (tenant_id) em todas as tabelas
- [ ] ✅ RLS (Row-Level Security) configurado
- [ ] ✅ Command Handler Pattern seguido
- [ ] ✅ Eventos de domínio emitidos corretamente
- [ ] ✅ Outbox Pattern usado para event publishing

### Tests
- [ ] ✅ Unit tests escritos (domain + application)
- [ ] ✅ Integration tests escritos (repository)
- [ ] ✅ E2E tests escritos (HTTP endpoints)
- [ ] ✅ Coverage: Domain 100%, Application 80%+, Overall 82%+
- [ ] ✅ `make test` passa sem erros
- [ ] ✅ `make test-unit` passa (~2 min)
- [ ] ✅ `make test-integration` passa (~10 min)
- [ ] ✅ `make test-e2e` passa (~10 min)

### Documentation
- [ ] ✅ Swagger comments adicionados
- [ ] ✅ Domain aggregate doc atualizado (guides/domain_mapping/)
- [ ] ✅ README.md atualizado (se feature importante)
- [ ] ✅ TODO.md atualizado (se feature grande)
- [ ] ✅ Migration UP/DOWN criada
- [ ] ✅ Code comments em português ou inglês

### Database
- [ ] ✅ Migration criada (XXX_description.up.sql + .down.sql)
- [ ] ✅ Migration testada (up + down)
- [ ] ✅ Indexes adicionados (performance)
- [ ] ✅ Foreign keys definidas
- [ ] ✅ Constraints check adicionadas
- [ ] ✅ RLS policies configuradas

### Build & Deploy
- [ ] ✅ `make build` passa sem erros
- [ ] ✅ `make fmt` executado
- [ ] ✅ `go vet` passa sem warnings
- [ ] ✅ No compilation errors
- [ ] ✅ No unused imports

---

## 🚀 Comandos Úteis

```bash
# Development
make dev              # Full stack: infra + API
make api              # Run API only
make infra            # Start infrastructure (PostgreSQL, RabbitMQ, Redis, Temporal)
make infra-stop       # Stop infrastructure
make infra-reset      # Clean + restart infrastructure

# Testing
make test             # Run all tests
make test-unit        # Fast unit tests (~2 min)
make test-integration # Integration tests (~10 min) - Requires: make infra
make test-e2e         # E2E tests (~10 min) - Requires: make infra + make api
make test-coverage    # Generate coverage report (HTML)

# Code Quality
make fmt              # Format code
make lint             # Run linter
make vet              # Run go vet

# Database
make migrate-up       # Apply all migrations
make migrate-down     # Rollback last migration
make migrate-status   # Check migration status

# Build
make build            # Build binary
make clean            # Clean build artifacts

# Documentation
make swagger          # Generate Swagger docs
```

**Ver mais**: [MAKEFILE.md](MAKEFILE.md)

---

## 💡 Dicas Importantes

### ✅ DO's

1. **SEMPRE use Optimistic Locking** (version field) em aggregates
2. **SEMPRE emita domain events** em mudanças de estado
3. **SEMPRE use Outbox Pattern** para garantir entrega de eventos
4. **SEMPRE implemente Soft Delete** (deleted_at)
5. **SEMPRE adicione tenant_id** para multi-tenancy
6. **SEMPRE configure RLS** em tabelas multi-tenant
7. **SEMPRE escreva testes** (unit + integration + e2e)
8. **SEMPRE documente** no Swagger e domain aggregate docs
9. **SEMPRE crie migrations** (up + down)
10. **SEMPRE siga Command Handler Pattern** (ver P0.md)

### ❌ DON'Ts

1. ❌ Não manipule aggregates diretamente no handler HTTP
2. ❌ Não exponha domain entities via API (use DTOs)
3. ❌ Não faça hard delete (sempre soft delete)
4. ❌ Não publique eventos diretamente (use EventBus + Outbox)
5. ❌ Não use GORM AutoMigrate em produção (use SQL migrations)
6. ❌ Não ignore erros de optimistic locking
7. ❌ Não pule testes unitários
8. ❌ Não commite código sem `make fmt`
9. ❌ Não use setters genéricos (crie business methods)
10. ❌ Não quebre a regra de dependência (Domain → Application → Infrastructure)

---

## 📞 Suporte

**Dúvidas sobre arquitetura?**
- Consulte: [AI_REPORT.md](AI_REPORT.md) - Auditoria completa com 8.2/10

**Dúvidas sobre domain model?**
- Consulte: [guides/domain_mapping/](guides/domain_mapping/) - 23 aggregates documentados

**Dúvidas sobre padrão de handlers?**
- Consulte: [P0.md](P0.md) - 100% dos handlers refatorados com Command Pattern

**Dúvidas sobre comandos?**
- Consulte: [MAKEFILE.md](MAKEFILE.md) ou [guides/MAKEFILE.md](guides/MAKEFILE.md)

**Issues & Bugs**:
- [GitHub Issues](https://github.com/ventros/crm/issues)

**Email**:
- dev@ventros.ai

---

## 📝 Exemplo Completo - Step by Step

Ver [P0.md](P0.md) para exemplo completo de refatoração seguindo todos os padrões.

**Template de Use Case**:
- Seção "Template de Use Case" em P0.md (linha 217)

**Template de Teste**:
- Seção "Template de Teste" em P0.md (linha 320)

---

**Mantido por**: Ventros CRM Team
**Versão**: 1.0
**Última Atualização**: 2025-10-12
**Status**: ✅ Production-Ready

---

**🎯 Lembre-se**: Qualidade > Velocidade. Siga os padrões, escreva testes, documente bem!
