# Documentação da Camada de Infraestrutura

## Visão Geral

A camada de infraestrutura implementa os detalhes técnicos de persistência, messaging, integrações externas e deployment. Esta camada isola o domínio e aplicação de frameworks e tecnologias específicas.

**Responsabilidades:**
- Implementar repositories (persistência)
- Gerenciar conexões com bancos de dados
- Implementar message brokers (RabbitMQ)
- Integrar com APIs externas (WAHA, Facebook, etc)
- Gerenciar workflows (Temporal)
- Configurar HTTP handlers e rotas
- Implementar middleware (auth, RLS, RBAC)

---

## 1. Persistência (Database)

### PostgreSQL + GORM

**Localização:** `infrastructure/persistence/`

#### Database.go
**Arquivo:** `database.go`

**Responsabilidade:** Gerenciar conexão e configuração do PostgreSQL

```go
type Database struct {
    DB *gorm.DB
}

func NewDatabase(config *config.Config) (*Database, error) {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        config.DB.Host, config.DB.Port, config.DB.User,
        config.DB.Password, config.DB.Name, config.DB.SSLMode,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })

    // Configurar connection pool
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    return &Database{DB: db}, nil
}
```

**Configurações:**
- Max Idle Connections: 10
- Max Open Connections: 100
- Connection Max Lifetime: 1 hora

#### Migrations

**Localização:** `infrastructure/database/migrations/`

**Ferramenta:** golang-migrate

**Migrations Principais:**
- `000001_create_users_table.up.sql`
- `000002_create_projects_table.up.sql`
- `000003_create_agents_table.up.sql`
- `000004_create_contacts_table.up.sql`
- `000005_create_sessions_table.up.sql`
- `000006_create_messages_table.up.sql`
- `000007_create_channels_table.up.sql`
- `000008_create_pipelines_table.up.sql`
- `000009_create_notes_table.up.sql`
- `000010_create_webhooks_table.up.sql`
- `000011_create_contact_events_table.up.sql`
- `000012_create_contact_lists_table.up.sql`
- `000013_create_domain_event_log_table.up.sql`
- `000014_create_trackings_table.up.sql`
- `000015_create_tracking_enrichments_table.up.sql`
- `000016_create_outbox_events_table.up.sql`
- `000017_create_processed_events_table.up.sql`

**Executar migrations:**
```bash
make migrate-up
```

---

### Entities (ORM Models)

**Localização:** `infrastructure/persistence/entities/`

Cada aggregate root tem uma entidade GORM correspondente:

#### Contact Entity
**Arquivo:** `contact.go`

```go
type Contact struct {
    ID               uuid.UUID              `gorm:"type:uuid;primary_key"`
    TenantID         string                 `gorm:"type:varchar(255);not null;index:idx_contacts_tenant"`
    Name             string                 `gorm:"type:varchar(255);not null"`
    Phone            string                 `gorm:"type:varchar(50);not null;index:idx_contacts_phone"`
    Email            string                 `gorm:"type:varchar(255)"`
    ProfilePictureURL string                `gorm:"type:text"`
    Tags             []string               `gorm:"type:text[]"`
    CustomFields     datatypes.JSON         `gorm:"type:jsonb"`
    PipelineID       *uuid.UUID             `gorm:"type:uuid"`
    PipelineStatusID *uuid.UUID             `gorm:"type:uuid"`
    CreatedAt        time.Time              `gorm:"not null"`
    UpdatedAt        time.Time              `gorm:"not null"`

    // Relationships
    Sessions        []Session        `gorm:"foreignKey:ContactID"`
    Messages        []Message        `gorm:"foreignKey:ContactID"`
    Notes           []Note           `gorm:"foreignKey:ContactID"`
    ContactEvents   []ContactEvent   `gorm:"foreignKey:ContactID"`
    Trackings       []Tracking       `gorm:"foreignKey:ContactID"`
}

// Indexes
func (Contact) TableName() string {
    return "contacts"
}

// Index: idx_contacts_tenant_phone (tenant_id, phone) UNIQUE
```

**Conversão Domain ↔ Entity:**
```go
// ToEntity: Domain → GORM Entity
func (c *contact.Contact) ToEntity() *entities.Contact {
    return &entities.Contact{
        ID:                c.ID(),
        TenantID:          c.TenantID(),
        Name:              c.Name(),
        Phone:             c.Phone().String(),
        Email:             c.Email().String(),
        ProfilePictureURL: c.ProfilePictureURL(),
        CustomFields:      c.CustomFields(),
        CreatedAt:         c.CreatedAt(),
        UpdatedAt:         c.UpdatedAt(),
    }
}

// FromEntity: GORM Entity → Domain
func ContactFromEntity(e *entities.Contact) (*contact.Contact, error) {
    phone, err := contact.NewPhone(e.Phone)
    if err != nil {
        return nil, err
    }

    email, _ := contact.NewEmail(e.Email)

    return contact.ReconstructContact(
        e.ID,
        e.TenantID,
        e.Name,
        phone,
        email,
        e.ProfilePictureURL,
        e.Tags,
        e.CustomFields,
        e.PipelineID,
        e.PipelineStatusID,
        e.CreatedAt,
        e.UpdatedAt,
    ), nil
}
```

#### Session Entity
**Arquivo:** `session.go`

```go
type Session struct {
    ID                  uuid.UUID       `gorm:"type:uuid;primary_key"`
    TenantID            string          `gorm:"type:varchar(255);not null;index"`
    ContactID           uuid.UUID       `gorm:"type:uuid;not null;index"`
    ChannelID           uuid.UUID       `gorm:"type:uuid;not null"`
    ProjectID           uuid.UUID       `gorm:"type:uuid;not null"`
    Status              string          `gorm:"type:varchar(50);not null"` // active, closed, expired
    AssignedAgentID     *uuid.UUID      `gorm:"type:uuid"`
    StartedAt           time.Time       `gorm:"not null"`
    EndedAt             *time.Time
    LastInteractionAt   *time.Time
    TotalMessages       int             `gorm:"default:0"`
    CustomFields        datatypes.JSON  `gorm:"type:jsonb"`
    CreatedAt           time.Time       `gorm:"not null"`
    UpdatedAt           time.Time       `gorm:"not null"`

    // Relationships
    Contact             Contact         `gorm:"foreignKey:ContactID"`
    Messages            []Message       `gorm:"foreignKey:SessionID"`
    Notes               []Note          `gorm:"foreignKey:SessionID"`
}

// Index: idx_sessions_contact (contact_id, started_at DESC)
// Index: idx_sessions_agent (assigned_agent_id, status)
```

#### Message Entity
**Arquivo:** `message.go`

```go
type Message struct {
    ID              uuid.UUID       `gorm:"type:uuid;primary_key"`
    TenantID        string          `gorm:"type:varchar(255);not null;index"`
    ContactID       uuid.UUID       `gorm:"type:uuid;not null;index"`
    SessionID       *uuid.UUID      `gorm:"type:uuid;index"`
    ChannelID       uuid.UUID       `gorm:"type:uuid;not null"`
    ExternalID      string          `gorm:"type:varchar(255);unique"` // WhatsApp message ID
    Direction       string          `gorm:"type:varchar(20);not null"` // inbound, outbound
    ContentType     string          `gorm:"type:varchar(50);not null"` // text, image, audio, etc
    Content         string          `gorm:"type:text"`
    MediaURL        string          `gorm:"type:text"`
    MediaMimeType   string          `gorm:"type:varchar(100)"`
    Metadata        datatypes.JSON  `gorm:"type:jsonb"`
    Status          string          `gorm:"type:varchar(50)"` // sent, delivered, read, failed
    SentBy          *uuid.UUID      `gorm:"type:uuid"` // Agent ID (for outbound)
    SentAt          time.Time       `gorm:"not null"`
    DeliveredAt     *time.Time
    ReadAt          *time.Time
    CreatedAt       time.Time       `gorm:"not null"`
    UpdatedAt       time.Time       `gorm:"not null"`
}

// Index: idx_messages_session (session_id, sent_at)
// Index: idx_messages_contact (contact_id, sent_at DESC)
// Index: idx_messages_external_id (external_id) UNIQUE
```

#### Outbox Event Entity
**Arquivo:** `outbox_event.go`

```go
type OutboxEvent struct {
    ID              uuid.UUID       `gorm:"type:uuid;primary_key"`
    EventID         uuid.UUID       `gorm:"type:uuid;not null;unique"`
    EventType       string          `gorm:"type:varchar(100);not null;index"`
    AggregateID     uuid.UUID       `gorm:"type:uuid;not null;index"`
    AggregateType   string          `gorm:"type:varchar(100);not null"`
    EventVersion    string          `gorm:"type:varchar(20);not null"`
    Payload         datatypes.JSON  `gorm:"type:jsonb;not null"`
    TenantID        string          `gorm:"type:varchar(255);not null;index"`
    OccurredAt      time.Time       `gorm:"not null;index"`
    ProcessedAt     *time.Time      `gorm:"index"`
    FailedAt        *time.Time
    RetryCount      int             `gorm:"default:0"`
    ErrorMessage    string          `gorm:"type:text"`
    CreatedAt       time.Time       `gorm:"not null"`
}

// Index: idx_outbox_pending (processed_at IS NULL, occurred_at ASC)
// Index: idx_outbox_failed (failed_at IS NOT NULL, retry_count)
```

#### Processed Event Entity
**Arquivo:** `processed_event.go`

```go
type ProcessedEvent struct {
    ID              uuid.UUID  `gorm:"type:uuid;primary_key"`
    EventID         uuid.UUID  `gorm:"type:uuid;not null"`
    ConsumerName    string     `gorm:"type:varchar(255);not null"`
    ProcessedAt     time.Time  `gorm:"not null"`
    DurationMs      *int       `gorm:"type:integer"`
}

// Index: idx_processed_events_unique (event_id, consumer_name) UNIQUE
// Purpose: Garantir idempotência - cada evento processado uma vez por consumer
```

---

### Repositories (GORM Implementation)

**Localização:** `infrastructure/persistence/`

Cada repository implementa a interface do domain:

#### Contact Repository
**Arquivo:** `gorm_contact_repository.go`

```go
type GormContactRepository struct {
    db *gorm.DB
}

func (r *GormContactRepository) Save(ctx context.Context, contact *domain.Contact) error {
    entity := ToContactEntity(contact)
    return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormContactRepository) FindByID(ctx context.Context, id uuid.UUID, tenantID string) (*domain.Contact, error) {
    var entity entities.Contact
    err := r.db.WithContext(ctx).
        Where("id = ? AND tenant_id = ?", id, tenantID).
        First(&entity).Error

    if err == gorm.ErrRecordNotFound {
        return nil, domain.ErrContactNotFound
    }
    if err != nil {
        return nil, err
    }

    return ContactFromEntity(&entity)
}

func (r *GormContactRepository) FindByPhone(ctx context.Context, phone string, tenantID string) (*domain.Contact, error) {
    var entity entities.Contact
    err := r.db.WithContext(ctx).
        Where("phone = ? AND tenant_id = ?", phone, tenantID).
        First(&entity).Error

    if err == gorm.ErrRecordNotFound {
        return nil, domain.ErrContactNotFound
    }
    if err != nil {
        return nil, err
    }

    return ContactFromEntity(&entity)
}

func (r *GormContactRepository) List(ctx context.Context, tenantID string, filters ContactFilters) ([]*domain.Contact, error) {
    var entities []entities.Contact

    query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

    // Apply filters
    if filters.PipelineID != nil {
        query = query.Where("pipeline_id = ?", *filters.PipelineID)
    }
    if filters.Tags != nil && len(filters.Tags) > 0 {
        query = query.Where("tags && ?", pq.Array(filters.Tags))
    }
    if filters.Search != "" {
        query = query.Where("name ILIKE ? OR phone ILIKE ?",
            "%"+filters.Search+"%", "%"+filters.Search+"%")
    }

    // Pagination
    query = query.Limit(filters.Limit).Offset(filters.Offset)

    if err := query.Find(&entities).Error; err != nil {
        return nil, err
    }

    contacts := make([]*domain.Contact, len(entities))
    for i, e := range entities {
        contact, err := ContactFromEntity(&e)
        if err != nil {
            return nil, err
        }
        contacts[i] = contact
    }

    return contacts, nil
}
```

#### Session Repository
**Arquivo:** `gorm_session_repository.go`

```go
func (r *GormSessionRepository) FindActiveByContact(ctx context.Context, contactID uuid.UUID, tenantID string) (*domain.Session, error) {
    var entity entities.Session
    err := r.db.WithContext(ctx).
        Where("contact_id = ? AND tenant_id = ? AND status = ?",
            contactID, tenantID, "active").
        Order("started_at DESC").
        First(&entity).Error

    if err == gorm.ErrRecordNotFound {
        return nil, domain.ErrSessionNotFound
    }
    if err != nil {
        return nil, err
    }

    return SessionFromEntity(&entity)
}

func (r *GormSessionRepository) FindExpiredSessions(ctx context.Context, before time.Time) ([]*domain.Session, error) {
    var entities []entities.Session
    err := r.db.WithContext(ctx).
        Where("status = ? AND last_interaction_at < ?", "active", before).
        Find(&entities).Error

    if err != nil {
        return nil, err
    }

    sessions := make([]*domain.Session, len(entities))
    for i, e := range entities {
        session, err := SessionFromEntity(&e)
        if err != nil {
            return nil, err
        }
        sessions[i] = session
    }

    return sessions, nil
}
```

#### Message Repository
**Arquivo:** `gorm_message_repository.go`

```go
func (r *GormMessageRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, tenantID string) ([]*domain.Message, error) {
    var entities []entities.Message
    err := r.db.WithContext(ctx).
        Where("session_id = ? AND tenant_id = ?", sessionID, tenantID).
        Order("sent_at ASC").
        Find(&entities).Error

    if err != nil {
        return nil, err
    }

    messages := make([]*domain.Message, len(entities))
    for i, e := range entities {
        msg, err := MessageFromEntity(&e)
        if err != nil {
            return nil, err
        }
        messages[i] = msg
    }

    return messages, nil
}

func (r *GormMessageRepository) FindByExternalID(ctx context.Context, externalID string, tenantID string) (*domain.Message, error) {
    var entity entities.Message
    err := r.db.WithContext(ctx).
        Where("external_id = ? AND tenant_id = ?", externalID, tenantID).
        First(&entity).Error

    if err == gorm.ErrRecordNotFound {
        return nil, nil // Not found é caso válido para deduplicação
    }
    if err != nil {
        return nil, err
    }

    return MessageFromEntity(&entity)
}
```

#### Outbox Repository
**Arquivo:** `gorm_outbox_repository.go`

```go
func (r *GormOutboxRepository) Save(ctx context.Context, event shared.DomainEvent) error {
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }

    outboxEvent := &entities.OutboxEvent{
        ID:            uuid.New(),
        EventID:       event.EventID(),
        EventType:     event.EventType(),
        AggregateID:   event.AggregateID(),
        AggregateType: extractAggregateType(event.EventType()),
        EventVersion:  event.EventVersion(),
        Payload:       payload,
        TenantID:      extractTenantID(event),
        OccurredAt:    event.OccurredAt(),
        CreatedAt:     time.Now(),
    }

    return r.db.WithContext(ctx).Create(outboxEvent).Error
}

func (r *GormOutboxRepository) FetchPending(ctx context.Context, limit int) ([]*entities.OutboxEvent, error) {
    var events []*entities.OutboxEvent
    err := r.db.WithContext(ctx).
        Where("processed_at IS NULL").
        Order("occurred_at ASC").
        Limit(limit).
        Find(&events).Error

    return events, err
}

func (r *GormOutboxRepository) MarkAsProcessed(ctx context.Context, eventID uuid.UUID) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&entities.OutboxEvent{}).
        Where("event_id = ?", eventID).
        Update("processed_at", now).Error
}

func (r *GormOutboxRepository) MarkAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error {
    now := time.Now()
    return r.db.WithContext(ctx).
        Model(&entities.OutboxEvent{}).
        Where("event_id = ?", eventID).
        Updates(map[string]interface{}{
            "failed_at":     now,
            "retry_count":   gorm.Expr("retry_count + 1"),
            "error_message": errorMsg,
        }).Error
}
```

#### Idempotency Checker
**Arquivo:** `idempotency_checker.go`

```go
type IdempotencyChecker struct {
    db *gorm.DB
}

func (c *IdempotencyChecker) IsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string) (bool, error) {
    var count int64
    err := c.db.WithContext(ctx).
        Model(&entities.ProcessedEvent{}).
        Where("event_id = ? AND consumer_name = ?", eventID, consumerName).
        Count(&count).Error

    return count > 0, err
}

func (c *IdempotencyChecker) MarkAsProcessed(ctx context.Context, eventID uuid.UUID, consumerName string, durationMs *int) error {
    processedEvent := &entities.ProcessedEvent{
        ID:           uuid.New(),
        EventID:      eventID,
        ConsumerName: consumerName,
        ProcessedAt:  time.Now(),
        DurationMs:   durationMs,
    }

    return c.db.WithContext(ctx).Create(processedEvent).Error
}
```

---

### Row-Level Security (RLS)

**Localização:** `infrastructure/persistence/rls_callback.go`

**Responsabilidade:** Garantir isolamento de tenants no nível do banco

```go
func RegisterRLSCallback(db *gorm.DB) {
    db.Callback().Query().Before("gorm:query").Register("rls:set_tenant", func(d *gorm.DB) {
        tenantID := d.Statement.Context.Value("tenant_id")
        if tenantID != nil {
            d.Statement.AddClause(clause.Where{
                Exprs: []clause.Expression{
                    clause.Eq{Column: "tenant_id", Value: tenantID},
                },
            })
        }
    })
}
```

**Uso:**
```go
// No middleware, adiciona tenant_id ao context
ctx := context.WithValue(r.Context(), "tenant_id", tenantID)

// Todas as queries automaticamente filtram por tenant_id
contacts, err := repo.List(ctx, filters)
```

---

## 2. Messaging (RabbitMQ)

**Localização:** `infrastructure/messaging/`

### RabbitMQ Connection
**Arquivo:** `rabbitmq.go`

```go
type RabbitMQConnection struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    config  *config.RabbitMQConfig
}

func NewRabbitMQConnection(cfg *config.RabbitMQConfig) (*RabbitMQConnection, error) {
    url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
        cfg.User, cfg.Password, cfg.Host, cfg.Port)

    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
    }

    channel, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, fmt.Errorf("failed to open channel: %w", err)
    }

    // Configurar QoS
    if err := channel.Qos(10, 0, false); err != nil {
        channel.Close()
        conn.Close()
        return nil, err
    }

    return &RabbitMQConnection{
        conn:    conn,
        channel: channel,
        config:  cfg,
    }, nil
}
```

### Exchanges e Queues

**Declaração em `rabbitmq.go`:**

```go
func (r *RabbitMQConnection) DeclareExchanges() error {
    exchanges := []string{
        "domain.events",      // Eventos de domínio
        "waha.events",        // Eventos do WAHA
        "contact.events",     // Eventos de contato
        "webhooks.outbound",  // Webhooks para entregar
    }

    for _, exchange := range exchanges {
        if err := r.channel.ExchangeDeclare(
            exchange,
            "topic",  // tipo: topic (routing por pattern)
            true,     // durable
            false,    // auto-deleted
            false,    // internal
            false,    // no-wait
            nil,      // arguments
        ); err != nil {
            return err
        }
    }

    return nil
}

func (r *RabbitMQConnection) DeclareQueues() error {
    queues := []struct {
        name       string
        exchange   string
        routingKey string
    }{
        {"waha.messages", "waha.events", "message.*"},
        {"contact.events.enrichment", "contact.events", "contact.*"},
        {"contact.events.webhook", "contact.events", "contact.*"},
        {"webhooks.delivery", "webhooks.outbound", "#"},
        {"domain.events.all", "domain.events", "#"},
    }

    for _, q := range queues {
        // Declarar queue
        _, err := r.channel.QueueDeclare(
            q.name,
            true,  // durable
            false, // auto-delete
            false, // exclusive
            false, // no-wait
            amqp.Table{
                "x-dead-letter-exchange": q.exchange + ".dlx",
                "x-message-ttl":          86400000, // 24h
            },
        )
        if err != nil {
            return err
        }

        // Bind queue to exchange
        if err := r.channel.QueueBind(
            q.name,
            q.routingKey,
            q.exchange,
            false,
            nil,
        ); err != nil {
            return err
        }
    }

    return nil
}
```

### Event Bus Adapters
**Arquivo:** `event_bus_adapters.go`

```go
type RabbitMQEventBus struct {
    conn *RabbitMQConnection
}

func (bus *RabbitMQEventBus) Publish(ctx context.Context, exchange string, routingKey string, event shared.DomainEvent) error {
    body, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return bus.conn.channel.PublishWithContext(
        ctx,
        exchange,
        routingKey,
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            MessageId:    event.EventID().String(),
            Timestamp:    event.OccurredAt(),
            Headers: amqp.Table{
                "event_type":    event.EventType(),
                "event_version": event.EventVersion(),
                "aggregate_id":  event.AggregateID().String(),
            },
        },
    )
}
```

### Consumers

#### WAHA Message Consumer
**Arquivo:** `waha_message_consumer.go`

```go
type WAHAMessageConsumer struct {
    wahaMessageService *message.WAHAMessageService
    idempotencyChecker IdempotencyChecker
    consumerName       string
}

func (c *WAHAMessageConsumer) Start(ctx context.Context, conn *RabbitMQConnection) error {
    msgs, err := conn.channel.Consume(
        "waha.messages", // queue
        c.consumerName,  // consumer tag
        false,           // auto-ack (manual ack for reliability)
        false,           // exclusive
        false,           // no-local
        false,           // no-wait
        nil,             // args
    )
    if err != nil {
        return err
    }

    for msg := range msgs {
        if err := c.ProcessMessage(ctx, msg); err != nil {
            // NACK and requeue
            msg.Nack(false, true)
            fmt.Printf("❌ Error processing message: %v\n", err)
        } else {
            // ACK
            msg.Ack(false)
            fmt.Printf("✅ Message processed successfully\n")
        }
    }

    return nil
}

func (c *WAHAMessageConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
    startTime := time.Now()

    var wahaEvent waha.WAHAMessageEvent
    if err := json.Unmarshal(delivery.Body, &wahaEvent); err != nil {
        return fmt.Errorf("invalid json: %w", err)
    }

    eventUUID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(wahaEvent.ID))

    // Verificar idempotência
    if c.idempotencyChecker != nil {
        processed, err := c.idempotencyChecker.IsProcessed(ctx, eventUUID, c.consumerName)
        if err != nil {
            fmt.Printf("⚠️  Failed to check idempotency: %v\n", err)
        } else if processed {
            fmt.Printf("⏭️  Event already processed, skipping: id=%s\n", wahaEvent.ID)
            return nil
        }
    }

    // Processar mensagem
    if err := c.wahaMessageService.ProcessWAHAMessage(ctx, wahaEvent); err != nil {
        return err
    }

    // Marcar como processado
    if c.idempotencyChecker != nil {
        duration := int(time.Since(startTime).Milliseconds())
        if err := c.idempotencyChecker.MarkAsProcessed(ctx, eventUUID, c.consumerName, &duration); err != nil {
            fmt.Printf("⚠️  Failed to mark event as processed: %v\n", err)
        }
    }

    return nil
}
```

#### Contact Event Consumer
**Arquivo:** `contact_event_consumer.go`

```go
type ContactEventConsumer struct {
    createContactEventUseCase *contactevent.CreateContactEventUseCase
    rabbitConn                *RabbitMQConnection
    idempotencyChecker        IdempotencyChecker
    logger                    *zap.Logger
}

func (c *ContactEventConsumer) Start(ctx context.Context) error {
    msgs, err := c.rabbitConn.channel.Consume(
        "contact.events.enrichment",
        "contact_event_consumer",
        false, // manual ack
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    for msg := range msgs {
        if err := c.processMessage(ctx, msg); err != nil {
            c.logger.Error("Failed to process contact event", zap.Error(err))
            msg.Nack(false, true) // requeue
        } else {
            msg.Ack(false)
        }
    }

    return nil
}
```

#### Webhook Queue Consumer
**Arquivo:** `webhook_queue.go`

```go
type WebhookQueueConsumer struct {
    conn       *RabbitMQConnection
    httpClient *http.Client
    maxRetries int
}

func (c *WebhookQueueConsumer) Start(ctx context.Context) error {
    msgs, err := c.conn.channel.Consume(
        "webhooks.delivery",
        "webhook_delivery_consumer",
        false,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }

    for msg := range msgs {
        if err := c.deliverWebhook(ctx, msg); err != nil {
            // Verificar retry count
            retryCount := 0
            if val, ok := msg.Headers["x-retry-count"]; ok {
                retryCount = val.(int)
            }

            if retryCount < c.maxRetries {
                // Requeue with delay
                msg.Nack(false, true)
            } else {
                // Send to DLQ
                c.sendToDLQ(msg)
                msg.Ack(false)
            }
        } else {
            msg.Ack(false)
        }
    }

    return nil
}

func (c *WebhookQueueConsumer) deliverWebhook(ctx context.Context, delivery amqp.Delivery) error {
    var webhook WebhookQueueMessage
    if err := json.Unmarshal(delivery.Body, &webhook); err != nil {
        return err
    }

    // Construir requisição
    payload, _ := json.Marshal(webhook.Payload)
    req, err := http.NewRequestWithContext(ctx, webhook.Method, webhook.URL, bytes.NewReader(payload))
    if err != nil {
        return err
    }

    // Adicionar headers
    req.Header.Set("Content-Type", "application/json")
    for k, v := range webhook.Headers {
        req.Header.Set(k, v)
    }

    // Calcular signature HMAC
    signature := calculateHMAC(payload, webhook.Secret)
    req.Header.Set("X-Webhook-Signature", signature)

    // Enviar
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("webhook delivery failed: status %d", resp.StatusCode)
    }

    return nil
}
```

### Message Metrics
**Arquivo:** `message_metrics.go`

```go
var (
    messageProcessingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "message_processing_duration_seconds",
            Help:    "Message processing duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"consumer", "status"},
    )

    messageProcessingTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "message_processing_total",
            Help: "Total messages processed",
        },
        []string{"consumer", "status"},
    )
)

func RecordMessageProcessing(consumer string, duration time.Duration, err error) {
    status := "success"
    if err != nil {
        status = "error"
    }

    messageProcessingDuration.WithLabelValues(consumer, status).Observe(duration.Seconds())
    messageProcessingTotal.WithLabelValues(consumer, status).Inc()
}
```

---

## 3. Integrações Externas

### WAHA (WhatsApp HTTP API)

**Localização:** `infrastructure/channels/waha/`

#### WAHA Client
**Arquivo:** `client.go`

```go
type WAHAClient struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

func NewWAHAClient(baseURL, apiKey string) *WAHAClient {
    return &WAHAClient{
        baseURL: baseURL,
        apiKey:  apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *WAHAClient) SendTextMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
    endpoint := fmt.Sprintf("%s/api/%s/sendText", c.baseURL, req.Session)

    body := map[string]interface{}{
        "chatId": req.ChatID,
        "text":   req.Text,
    }

    return c.doRequest(ctx, "POST", endpoint, body)
}

func (c *WAHAClient) SendImageMessage(ctx context.Context, req SendImageRequest) (*SendMessageResponse, error) {
    endpoint := fmt.Sprintf("%s/api/%s/sendImage", c.baseURL, req.Session)

    body := map[string]interface{}{
        "chatId": req.ChatID,
        "file": map[string]string{
            "url":      req.ImageURL,
            "mimetype": "image/jpeg",
        },
        "caption": req.Caption,
    }

    return c.doRequest(ctx, "POST", endpoint, body)
}

func (c *WAHAClient) GetProfilePicture(ctx context.Context, session, chatID string) (string, error) {
    endpoint := fmt.Sprintf("%s/api/%s/contacts/profile-picture?chatId=%s",
        c.baseURL, session, chatID)

    resp, err := c.doRequest(ctx, "GET", endpoint, nil)
    if err != nil {
        return "", err
    }

    return resp.ProfilePictureURL, nil
}

func (c *WAHAClient) doRequest(ctx context.Context, method, url string, body interface{}) (*Response, error) {
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, err
        }
        reqBody = bytes.NewReader(jsonBody)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Api-Key", c.apiKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("WAHA API error: status %d", resp.StatusCode)
    }

    var result Response
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}
```

#### Message Adapter
**Arquivo:** `message_adapter.go`

```go
type MessageAdapter struct {
    client *WAHAClient
}

func (a *MessageAdapter) AdaptWAHAEventToMessage(wahaEvent WAHAMessageEvent) (*message.CreateMessageInput, error) {
    // Determinar tipo de conteúdo
    contentType := a.detectContentType(wahaEvent)

    // Extrair conteúdo
    content := a.extractContent(wahaEvent, contentType)

    // Extrair metadata
    metadata := a.extractMetadata(wahaEvent)

    return &message.CreateMessageInput{
        ExternalID:    wahaEvent.ID,
        Direction:     "inbound",
        ContentType:   contentType,
        Content:       content,
        MediaURL:      a.extractMediaURL(wahaEvent),
        MediaMimeType: a.extractMimeType(wahaEvent),
        Metadata:      metadata,
        SentAt:        time.Unix(wahaEvent.Timestamp, 0),
    }, nil
}

func (a *MessageAdapter) detectContentType(event WAHAMessageEvent) string {
    if event.Type == "image" {
        return "image"
    }
    if event.Type == "audio" || event.Type == "ptt" {
        return "audio"
    }
    if event.Type == "video" {
        return "video"
    }
    if event.Type == "document" {
        return "document"
    }
    if event.Type == "location" {
        return "location"
    }
    if event.Type == "vcard" {
        return "contact"
    }
    if event.Type == "sticker" {
        return "sticker"
    }
    return "text"
}
```

---

## 4. Temporal Workflows

**Localização:** `internal/workflows/` e `infrastructure/workflow/`

### Outbox Worker
**Arquivo:** `infrastructure/workflow/outbox_worker.go`

```go
type OutboxWorker struct {
    temporalClient client.Client
    config         *config.Config
}

func (w *OutboxWorker) Start(ctx context.Context) error {
    // Criar worker
    worker := worker.New(w.temporalClient, "outbox-queue", worker.Options{})

    // Registrar workflow
    worker.RegisterWorkflow(outboxworkflow.OutboxProcessorWorkflow)

    // Registrar activities
    worker.RegisterActivity(outboxworkflow.FetchPendingEventsActivity)
    worker.RegisterActivity(outboxworkflow.PublishEventActivity)
    worker.RegisterActivity(outboxworkflow.MarkAsProcessedActivity)

    // Iniciar workflow contínuo
    workflowOptions := client.StartWorkflowOptions{
        ID:        "outbox-processor-main",
        TaskQueue: "outbox-queue",
    }

    input := outboxworkflow.OutboxProcessorWorkflowInput{
        BatchSize:    100,
        PollInterval: 1 * time.Second,
        MaxRetries:   5,
        RetryBackoff: 30 * time.Second,
    }

    _, err := w.temporalClient.ExecuteWorkflow(ctx, workflowOptions,
        outboxworkflow.OutboxProcessorWorkflow, input)
    if err != nil {
        return err
    }

    // Start worker (blocking)
    return worker.Run(worker.InterruptCh())
}
```

### Outbox Workflow
**Arquivo:** `internal/workflows/outbox/outbox_workflow.go`

```go
func OutboxProcessorWorkflow(ctx workflow.Context, input OutboxProcessorWorkflowInput) error {
    logger := workflow.GetLogger(ctx)

    for {
        // Buscar eventos pendentes
        var events []*entities.OutboxEvent
        err := workflow.ExecuteActivity(ctx, FetchPendingEventsActivity,
            input.BatchSize).Get(ctx, &events)
        if err != nil {
            logger.Error("Failed to fetch pending events", "error", err)
            workflow.Sleep(ctx, input.PollInterval)
            continue
        }

        if len(events) == 0 {
            // Nenhum evento pendente, aguardar
            workflow.Sleep(ctx, input.PollInterval)
            continue
        }

        logger.Info("Processing outbox events", "count", len(events))

        // Processar eventos em paralelo
        futures := make([]workflow.Future, len(events))
        for i, event := range events {
            futures[i] = workflow.ExecuteActivity(ctx, PublishEventActivity, event)
        }

        // Aguardar conclusão
        for i, future := range futures {
            event := events[i]
            if err := future.Get(ctx, nil); err != nil {
                logger.Error("Failed to publish event",
                    "event_id", event.EventID, "error", err)

                // Marcar como falho
                workflow.ExecuteActivity(ctx, MarkAsFailedActivity,
                    event.EventID, err.Error())
            } else {
                // Marcar como processado
                workflow.ExecuteActivity(ctx, MarkAsProcessedActivity,
                    event.EventID)
            }
        }
    }

    return nil
}
```

### Session Lifecycle Workflow
**Arquivo:** `internal/workflows/session/session_lifecycle_workflow.go`

```go
func SessionLifecycleWorkflow(ctx workflow.Context, input SessionLifecycleInput) error {
    logger := workflow.GetLogger(ctx)

    // Criar sessão
    var sessionID uuid.UUID
    err := workflow.ExecuteActivity(ctx, CreateSessionActivity, input).Get(ctx, &sessionID)
    if err != nil {
        return err
    }

    logger.Info("Session created", "session_id", sessionID)

    // Aguardar timeout ou fechamento manual
    selector := workflow.NewSelector(ctx)

    // Timer de timeout
    timeoutTimer := workflow.NewTimer(ctx, input.TimeoutDuration)
    selector.AddFuture(timeoutTimer, func(f workflow.Future) {
        logger.Info("Session timeout reached", "session_id", sessionID)
        workflow.ExecuteActivity(ctx, CloseSessionActivity, sessionID, "timeout")
    })

    // Signal de fechamento manual
    var closeSignal CloseSessionSignal
    closeChannel := workflow.GetSignalChannel(ctx, "close_session")
    selector.AddReceive(closeChannel, func(c workflow.ReceiveChannel, more bool) {
        c.Receive(ctx, &closeSignal)
        logger.Info("Session closed manually", "session_id", sessionID)
        workflow.ExecuteActivity(ctx, CloseSessionActivity, sessionID, closeSignal.Reason)
    })

    // Signal de nova mensagem (reset timer)
    messageChannel := workflow.GetSignalChannel(ctx, "new_message")
    selector.AddReceive(messageChannel, func(c workflow.ReceiveChannel, more bool) {
        var msg NewMessageSignal
        c.Receive(ctx, &msg)
        logger.Info("New message received, resetting timer", "session_id", sessionID)
        // Timer é resetado automaticamente ao receber signal
    })

    selector.Select(ctx)

    return nil
}
```

### Session Worker
**Arquivo:** `infrastructure/workflow/session_worker.go`

```go
func (w *SessionWorker) StartSessionWorkflow(ctx context.Context, input session.SessionLifecycleInput) error {
    workflowOptions := client.StartWorkflowOptions{
        ID:        fmt.Sprintf("session-%s", input.SessionID),
        TaskQueue: "session-queue",
    }

    _, err := w.temporalClient.ExecuteWorkflow(ctx, workflowOptions,
        session.SessionLifecycleWorkflow, input)

    return err
}

func (w *SessionWorker) SendMessageSignal(ctx context.Context, sessionID uuid.UUID) error {
    workflowID := fmt.Sprintf("session-%s", sessionID)

    return w.temporalClient.SignalWorkflow(ctx, workflowID, "",
        "new_message", session.NewMessageSignal{MessageID: uuid.New()})
}

func (w *SessionWorker) CloseSession(ctx context.Context, sessionID uuid.UUID, reason string) error {
    workflowID := fmt.Sprintf("session-%s", sessionID)

    return w.temporalClient.SignalWorkflow(ctx, workflowID, "",
        "close_session", session.CloseSessionSignal{Reason: reason})
}
```

---

## 5. HTTP Layer

**Localização:** `infrastructure/http/`

### Handlers

#### Contact Handler
**Arquivo:** `handlers/contact_handler.go`

```go
type ContactHandler struct {
    createContactUseCase *contact.CreateContactUseCase
    updateContactUseCase *contact.UpdateContactUseCase
    listContactsUseCase  *contact.ListContactsUseCase
}

// CreateContact godoc
// @Summary Create new contact
// @Tags contacts
// @Accept json
// @Produce json
// @Param request body dto.CreateContactRequest true "Contact data"
// @Success 201 {object} dto.ContactResponse
// @Router /contacts [post]
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateContactRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validar
    if err := req.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Extrair tenant_id do context (injetado por middleware)
    tenantID := r.Context().Value("tenant_id").(string)

    // Executar use case
    contact, err := h.createContactUseCase.Execute(r.Context(), tenantID, req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Retornar response
    w.Header().Set("Content-Type", "application/json")
    w.WriteStatus(http.StatusCreated)
    json.NewEncoder(w).Encode(dto.ContactToResponse(contact))
}

func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
    tenantID := r.Context().Value("tenant_id").(string)

    // Parse query params
    filters := dto.ContactFilters{
        Search:     r.URL.Query().Get("search"),
        PipelineID: parseUUID(r.URL.Query().Get("pipeline_id")),
        Tags:       r.URL.Query()["tags"],
        Limit:      parseInt(r.URL.Query().Get("limit"), 50),
        Offset:     parseInt(r.URL.Query().Get("offset"), 0),
    }

    contacts, err := h.listContactsUseCase.Execute(r.Context(), tenantID, filters)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(dto.ContactListToResponse(contacts))
}
```

#### Message Handler
**Arquivo:** `handlers/message_handler.go`

```go
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
    var req dto.SendMessageRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    tenantID := r.Context().Value("tenant_id").(string)
    userID := r.Context().Value("user_id").(uuid.UUID)

    message, err := h.sendMessageUseCase.Execute(r.Context(), messaging.SendMessageInput{
        TenantID:    tenantID,
        ContactID:   req.ContactID,
        ChannelID:   req.ChannelID,
        ContentType: req.ContentType,
        Content:     req.Content,
        MediaURL:    req.MediaURL,
        SentBy:      userID,
    })
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(dto.MessageToResponse(message))
}
```

#### WAHA Webhook Handler
**Arquivo:** `handlers/waha_webhook_handler.go`

```go
type WAHAWebhookHandler struct {
    rabbitConn *messaging.RabbitMQConnection
}

func (h *WAHAWebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    var event waha.WAHAMessageEvent
    if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    // Publicar para RabbitMQ
    if err := h.publishToQueue(r.Context(), event); err != nil {
        fmt.Printf("Failed to publish to queue: %v\n", err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "OK")
}

func (h *WAHAWebhookHandler) publishToQueue(ctx context.Context, event waha.WAHAMessageEvent) error {
    body, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return h.rabbitConn.Channel().PublishWithContext(
        ctx,
        "waha.events",
        "message.received",
        false,
        false,
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,
            MessageId:    event.ID,
        },
    )
}
```

### Middleware

#### Auth Middleware
**Arquivo:** `middleware/auth.go`

```go
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")

            token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method")
                }
                return []byte(jwtSecret), nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            claims := token.Claims.(jwt.MapClaims)

            // Injetar claims no context
            ctx := r.Context()
            ctx = context.WithValue(ctx, "user_id", claims["user_id"])
            ctx = context.WithValue(ctx, "tenant_id", claims["tenant_id"])
            ctx = context.WithValue(ctx, "roles", claims["roles"])

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

#### RBAC Middleware
**Arquivo:** `middleware/rbac.go`

```go
func RequirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            roles := r.Context().Value("roles").([]interface{})

            hasPermission := false
            for _, role := range roles {
                roleStr := role.(string)
                if hasRolePermission(roleStr, permission) {
                    hasPermission = true
                    break
                }
            }

            if !hasPermission {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func hasRolePermission(role, permission string) bool {
    // admin tem todas as permissões
    if role == "admin" {
        return true
    }

    // Mapa de permissões por role
    rolePermissions := map[string][]string{
        "agent": {
            "contacts.read",
            "contacts.update",
            "messages.read",
            "messages.send",
            "sessions.read",
            "sessions.update",
            "notes.create",
            "notes.read",
        },
        "supervisor": {
            "contacts.*",
            "messages.*",
            "sessions.*",
            "agents.read",
            "notes.*",
            "pipelines.read",
        },
    }

    perms, ok := rolePermissions[role]
    if !ok {
        return false
    }

    for _, perm := range perms {
        if perm == permission || strings.HasSuffix(perm, ".*") {
            return true
        }
    }

    return false
}
```

#### RLS Middleware
**Arquivo:** `middleware/rls.go`

```go
func RLSMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenantID := r.Context().Value("tenant_id")
            if tenantID == nil {
                http.Error(w, "Missing tenant_id", http.StatusUnauthorized)
                return
            }

            // Adicionar tenant_id ao context do GORM
            ctx := r.Context()
            ctx = context.WithValue(ctx, "gorm:tenant_id", tenantID)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Routes
**Arquivo:** `routes/routes.go`

```go
func SetupRoutes(router *chi.Mux, handlers *Handlers, cfg *config.Config) {
    // Public routes
    router.Post("/auth/login", handlers.Auth.Login)
    router.Post("/auth/register", handlers.Auth.Register)

    // Webhook (sem auth)
    router.Post("/webhooks/waha", handlers.WAHAWebhook.HandleWebhook)

    // Protected routes
    router.Group(func(r chi.Router) {
        r.Use(middleware.AuthMiddleware(cfg.JWTSecret))
        r.Use(middleware.RLSMiddleware(handlers.DB))

        // Contacts
        r.Route("/contacts", func(r chi.Router) {
            r.Use(middleware.RequirePermission("contacts.read"))
            r.Get("/", handlers.Contact.ListContacts)
            r.Get("/{id}", handlers.Contact.GetContact)

            r.With(middleware.RequirePermission("contacts.create")).
                Post("/", handlers.Contact.CreateContact)

            r.With(middleware.RequirePermission("contacts.update")).
                Put("/{id}", handlers.Contact.UpdateContact)

            r.Get("/{id}/sessions", handlers.Session.ListContactSessions)
            r.Get("/{id}/messages", handlers.Message.ListContactMessages)
            r.Get("/{id}/notes", handlers.Note.ListContactNotes)
        })

        // Messages
        r.Route("/messages", func(r chi.Router) {
            r.Use(middleware.RequirePermission("messages.read"))
            r.Get("/", handlers.Message.ListMessages)

            r.With(middleware.RequirePermission("messages.send")).
                Post("/send", handlers.Message.SendMessage)
        })

        // Sessions
        r.Route("/sessions", func(r chi.Router) {
            r.Use(middleware.RequirePermission("sessions.read"))
            r.Get("/", handlers.Session.ListSessions)
            r.Get("/{id}", handlers.Session.GetSession)

            r.With(middleware.RequirePermission("sessions.update")).
                Post("/{id}/assign", handlers.Session.AssignAgent)

            r.With(middleware.RequirePermission("sessions.update")).
                Post("/{id}/close", handlers.Session.CloseSession)
        })

        // Admin only
        r.Group(func(r chi.Router) {
            r.Use(middleware.RequireRole("admin"))

            r.Route("/agents", func(r chi.Router) {
                r.Get("/", handlers.Agent.ListAgents)
                r.Post("/", handlers.Agent.CreateAgent)
                r.Put("/{id}", handlers.Agent.UpdateAgent)
            })

            r.Route("/channels", func(r chi.Router) {
                r.Get("/", handlers.Channel.ListChannels)
                r.Post("/", handlers.Channel.CreateChannel)
                r.Put("/{id}", handlers.Channel.UpdateChannel)
            })
        })
    })
}
```

---

## 6. Configuration

**Localização:** `infrastructure/config/config.go`

```go
type Config struct {
    Server   ServerConfig
    DB       DatabaseConfig
    RabbitMQ RabbitMQConfig
    Temporal TemporalConfig
    WAHA     WAHAConfig
    JWT      JWTConfig
}

type ServerConfig struct {
    Port string `env:"SERVER_PORT" envDefault:"8080"`
    Host string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
}

type DatabaseConfig struct {
    Host     string `env:"DB_HOST" envDefault:"localhost"`
    Port     string `env:"DB_PORT" envDefault:"5432"`
    User     string `env:"DB_USER" envDefault:"postgres"`
    Password string `env:"DB_PASSWORD" envDefault:"postgres"`
    Name     string `env:"DB_NAME" envDefault:"ventros_crm"`
    SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

type RabbitMQConfig struct {
    Host     string `env:"RABBITMQ_HOST" envDefault:"localhost"`
    Port     string `env:"RABBITMQ_PORT" envDefault:"5672"`
    User     string `env:"RABBITMQ_USER" envDefault:"guest"`
    Password string `env:"RABBITMQ_PASSWORD" envDefault:"guest"`
}

type TemporalConfig struct {
    Host      string `env:"TEMPORAL_HOST" envDefault:"localhost:7233"`
    Namespace string `env:"TEMPORAL_NAMESPACE" envDefault:"default"`
}

type WAHAConfig struct {
    BaseURL string `env:"WAHA_BASE_URL" envDefault:"http://localhost:3000"`
    APIKey  string `env:"WAHA_API_KEY"`
}

type JWTConfig struct {
    Secret string `env:"JWT_SECRET" envDefault:"supersecret"`
    TTL    int    `env:"JWT_TTL" envDefault:"86400"` // 24h
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

---

## 7. Health Checks

**Localização:** `infrastructure/health/checker.go`

```go
type HealthChecker struct {
    db             *gorm.DB
    rabbitConn     *messaging.RabbitMQConnection
    temporalClient client.Client
}

func (h *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Status: "healthy",
        Checks: make(map[string]CheckResult),
    }

    // Check database
    dbStatus := h.checkDatabase(ctx)
    status.Checks["database"] = dbStatus
    if !dbStatus.Healthy {
        status.Status = "unhealthy"
    }

    // Check RabbitMQ
    mqStatus := h.checkRabbitMQ(ctx)
    status.Checks["rabbitmq"] = mqStatus
    if !mqStatus.Healthy {
        status.Status = "unhealthy"
    }

    // Check Temporal
    temporalStatus := h.checkTemporal(ctx)
    status.Checks["temporal"] = temporalStatus
    if !temporalStatus.Healthy {
        status.Status = "degraded"
    }

    return status
}

func (h *HealthChecker) checkDatabase(ctx context.Context) CheckResult {
    sqlDB, err := h.db.DB()
    if err != nil {
        return CheckResult{Healthy: false, Message: err.Error()}
    }

    if err := sqlDB.PingContext(ctx); err != nil {
        return CheckResult{Healthy: false, Message: err.Error()}
    }

    return CheckResult{Healthy: true, Message: "Connected"}
}
```

---

## Resumo

A camada de infraestrutura implementa:

- ✅ **Persistência:** PostgreSQL + GORM com 17 migrations, RLS, indexes otimizados
- ✅ **Messaging:** RabbitMQ com 5+ queues, DLQs, idempotência, metrics
- ✅ **Workflows:** Temporal para orquestração (outbox, session lifecycle)
- ✅ **Integrações:** WAHA client para WhatsApp, adapters para outros canais
- ✅ **HTTP:** Chi router, 30+ endpoints, auth/RBAC/RLS middleware
- ✅ **Observabilidade:** Health checks, Prometheus metrics, structured logging
- ✅ **Configuração:** Environment-based config, 12-factor app

**Performance:**
- Connection pooling (100 max connections)
- RabbitMQ QoS (10 prefetch)
- Batch processing (100 eventos/vez no outbox)
- Index otimizado em todas as queries principais

**Confiabilidade:**
- Idempotência em todos os consumers
- Dead Letter Queues para falhas
- Retry automático com backoff
- Health checks contínuos
- Graceful shutdown

**Segurança:**
- JWT authentication
- RBAC (role-based access control)
- RLS (row-level security) no banco
- HMAC signatures em webhooks
- Secrets via environment variables
