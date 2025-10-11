# Broadcast System - Disparos em Massa

## Conceito

Sistema de **disparos em massa** (broadcast) para enviar mensagens para listas de contatos, com suporte a:
- Disparos imediatos
- Disparos agendados (scheduled)
- Personalização por contato (variáveis)
- Rate limiting (controle de velocidade)
- Status tracking (pendente, em andamento, concluído, falhou)

---

## Domínio

### ContactList (Lista de Contatos)

```go
// ContactList representa uma lista de contatos
type ContactList struct {
    id          uuid.UUID
    tenantID    string
    name        string
    description string
    tags        []string      // Tags para filtrar contatos
    filters     ListFilters   // Filtros dinâmicos
    contactIDs  []uuid.UUID   // IDs fixos (snapshot)
    isDynamic   bool          // Se true, recalcula contatos dinamicamente
    createdAt   time.Time
    updatedAt   time.Time
}

// ListFilters filtros para listas dinâmicas
type ListFilters struct {
    PipelineID *uuid.UUID `json:"pipeline_id,omitempty"`
    StatusID   *uuid.UUID `json:"status_id,omitempty"`
    Tags       []string   `json:"tags,omitempty"`
    CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}
```

**Exemplos:**
- Lista estática: "Clientes VIP" (100 IDs fixos)
- Lista dinâmica: "Leads Qualificados" (filtro: `pipeline_id=X AND status_id=Y`)

---

### Broadcast (Disparo)

```go
// Broadcast representa um disparo em massa
type Broadcast struct {
    id              uuid.UUID
    tenantID        string
    name            string
    listID          uuid.UUID        // Lista de contatos alvo
    messageTemplate MessageTemplate  // Template da mensagem
    status          BroadcastStatus
    scheduledFor    *time.Time       // Quando disparar (nil = imediato)
    startedAt       *time.Time
    completedAt     *time.Time

    // Stats
    totalContacts   int
    sentCount       int
    failedCount     int
    pendingCount    int

    // Rate limiting
    rateLimit       int  // mensagens por minuto (0 = sem limite)

    createdAt       time.Time
    updatedAt       time.Time

    events []DomainEvent
}

type BroadcastStatus string

const (
    BroadcastStatusDraft      BroadcastStatus = "draft"       // Rascunho
    BroadcastStatusScheduled  BroadcastStatus = "scheduled"   // Agendado
    BroadcastStatusRunning    BroadcastStatus = "running"     // Em execução
    BroadcastStatusCompleted  BroadcastStatus = "completed"   // Concluído
    BroadcastStatusFailed     BroadcastStatus = "failed"      // Falhou
    BroadcastStatusCancelled  BroadcastStatus = "cancelled"   // Cancelado
)

// MessageTemplate template da mensagem com variáveis
type MessageTemplate struct {
    Type        string                 `json:"type"`  // text, template, media
    Content     string                 `json:"content"`
    TemplateID  *string                `json:"template_id,omitempty"`
    Variables   map[string]string      `json:"variables,omitempty"`
    MediaURL    *string                `json:"media_url,omitempty"`
}
```

**Fluxo de vida:**
```
draft → scheduled → running → completed
                   ↓
                cancelled
```

---

### BroadcastExecution (Execução Individual)

```go
// BroadcastExecution rastreia envio para cada contato
type BroadcastExecution struct {
    id          uuid.UUID
    broadcastID uuid.UUID
    contactID   uuid.UUID
    status      ExecutionStatus
    messageID   *uuid.UUID   // ID da mensagem enviada
    error       *string
    sentAt      *time.Time
    createdAt   time.Time
}

type ExecutionStatus string

const (
    ExecutionStatusPending   ExecutionStatus = "pending"
    ExecutionStatusSending   ExecutionStatus = "sending"
    ExecutionStatusSent      ExecutionStatus = "sent"
    ExecutionStatusFailed    ExecutionStatus = "failed"
    ExecutionStatusSkipped   ExecutionStatus = "skipped"
)
```

---

## Use Cases

### 1. Criar Lista Estática

```go
list, _ := listManager.CreateStaticList(ctx, CreateStaticListInput{
    TenantID:    "tenant-123",
    Name:        "Clientes VIP Janeiro",
    Description: "Clientes que compraram > R$1000 em Janeiro",
    ContactIDs:  []uuid.UUID{uuid1, uuid2, uuid3},
    Tags:        []string{"vip", "janeiro"},
})
```

### 2. Criar Lista Dinâmica

```go
list, _ := listManager.CreateDynamicList(ctx, CreateDynamicListInput{
    TenantID:    "tenant-123",
    Name:        "Leads Qualificados",
    Description: "Leads no status 'Qualificado'",
    Filters: ListFilters{
        PipelineID: &pipelineID,
        StatusID:   &qualifiedStatusID,
    },
})

// Lista dinâmica recalcula contatos em tempo real
contacts, _ := list.ResolveContacts(contactRepo)
```

### 3. Criar Disparo Imediato

```go
broadcast, _ := broadcastManager.CreateBroadcast(ctx, CreateBroadcastInput{
    TenantID: "tenant-123",
    Name:     "Promoção Flash",
    ListID:   listID,
    MessageTemplate: MessageTemplate{
        Type:    "text",
        Content: "Olá {{name}}! Promoção exclusiva: 50% OFF até hoje!",
        Variables: map[string]string{
            "name": "contact.name",  // Variável do contato
        },
    },
    ScheduledFor: nil,  // Imediato
    RateLimit:    60,   // 60 msgs/min
})

// Executar imediatamente
broadcastWorker.ExecuteBroadcast(ctx, broadcast.ID())
```

### 4. Criar Disparo Agendado

```go
scheduledTime := time.Now().Add(24 * time.Hour)

broadcast, _ := broadcastManager.CreateBroadcast(ctx, CreateBroadcastInput{
    Name:     "Newsletter Semanal",
    ListID:   listID,
    MessageTemplate: MessageTemplate{
        Type:       "template",
        TemplateID: ptr("newsletter_template"),
        Variables: map[string]string{
            "week": "current_week",
        },
    },
    ScheduledFor: &scheduledTime,  // Agendado para amanhã
    RateLimit:    100,
})

// Worker verifica broadcasts agendados periodicamente
// Quando scheduledFor <= now, executa automaticamente
```

### 5. Cancelar Disparo

```go
broadcast.Cancel()
// Status: scheduled → cancelled
```

### 6. Ver Status de Execução

```go
stats := broadcast.GetStats()
// {
//   totalContacts: 1000,
//   sentCount: 850,
//   failedCount: 50,
//   pendingCount: 100,
//   progress: 85%
// }
```

---

## Workers

### BroadcastSchedulerWorker

Verifica broadcasts agendados e os dispara no horário correto:

```go
type BroadcastSchedulerWorker struct {
    db              *gorm.DB
    broadcastWorker *BroadcastExecutionWorker
    pollInterval    time.Duration
}

func (w *BroadcastSchedulerWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.pollInterval)  // Ex: 1 minuto

    for {
        select {
        case <-ticker.C:
            // Busca broadcasts agendados prontos
            broadcasts := w.findReadyBroadcasts()

            for _, b := range broadcasts {
                // Marca como running
                b.Start()

                // Dispara execução
                go w.broadcastWorker.ExecuteBroadcast(ctx, b.ID())
            }
        }
    }
}

func (w *BroadcastSchedulerWorker) findReadyBroadcasts() []Broadcast {
    // SELECT * FROM broadcasts
    // WHERE status = 'scheduled'
    // AND scheduled_for <= NOW()
    // ORDER BY scheduled_for ASC
}
```

### BroadcastExecutionWorker

Executa o disparo para todos os contatos:

```go
type BroadcastExecutionWorker struct {
    contactRepo ContactRepository
    messageSender MessageSender
    rateLimiter RateLimiter
}

func (w *BroadcastExecutionWorker) ExecuteBroadcast(ctx context.Context, broadcastID uuid.UUID) error {
    broadcast, _ := w.repo.FindByID(broadcastID)

    // 1. Resolve contatos da lista
    list, _ := w.listRepo.FindByID(broadcast.ListID())
    contacts := list.ResolveContacts(w.contactRepo)

    broadcast.UpdateTotalContacts(len(contacts))

    // 2. Cria execuções
    executions := make([]BroadcastExecution, len(contacts))
    for i, contact := range contacts {
        executions[i] = NewBroadcastExecution(broadcast.ID(), contact.ID())
    }
    w.execRepo.SaveBatch(executions)

    // 3. Executa com rate limiting
    rateLimiter := NewRateLimiter(broadcast.RateLimit())

    for _, execution := range executions {
        rateLimiter.Wait()  // Respeita rate limit

        contact := w.contactRepo.FindByID(execution.ContactID())

        // Renderiza mensagem com variáveis
        message := w.renderMessage(broadcast.MessageTemplate(), contact)

        // Envia
        messageID, err := w.messageSender.Send(contact, message)

        if err != nil {
            execution.MarkFailed(err.Error())
            broadcast.IncrementFailed()
        } else {
            execution.MarkSent(messageID)
            broadcast.IncrementSent()
        }

        w.execRepo.Save(execution)
        w.broadcastRepo.Save(broadcast)
    }

    // 4. Marca como completo
    broadcast.Complete()
    w.broadcastRepo.Save(broadcast)

    return nil
}
```

---

## API Endpoints

### Contact Lists

```
POST   /api/v1/contact-lists                    # Criar lista
GET    /api/v1/contact-lists                    # Listar listas
GET    /api/v1/contact-lists/:id                # Ver lista
PUT    /api/v1/contact-lists/:id                # Atualizar lista
DELETE /api/v1/contact-lists/:id                # Deletar lista
GET    /api/v1/contact-lists/:id/contacts       # Ver contatos da lista
POST   /api/v1/contact-lists/:id/contacts       # Adicionar contatos
DELETE /api/v1/contact-lists/:id/contacts/:cid  # Remover contato
```

### Broadcasts

```
POST   /api/v1/broadcasts                       # Criar disparo
GET    /api/v1/broadcasts                       # Listar disparos
GET    /api/v1/broadcasts/:id                   # Ver disparo
PUT    /api/v1/broadcasts/:id                   # Atualizar (só draft)
DELETE /api/v1/broadcasts/:id                   # Deletar (só draft)
POST   /api/v1/broadcasts/:id/schedule          # Agendar
POST   /api/v1/broadcasts/:id/execute           # Executar imediato
POST   /api/v1/broadcasts/:id/cancel            # Cancelar
GET    /api/v1/broadcasts/:id/stats             # Estatísticas
GET    /api/v1/broadcasts/:id/executions        # Execuções individuais
```

---

## Exemplos de Uso

### Exemplo 1: Campanha de Promoção (Imediato)

```bash
# 1. Criar lista de clientes ativos
curl -X POST /api/v1/contact-lists \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Clientes Ativos",
    "is_dynamic": true,
    "filters": {
      "pipeline_id": "uuid",
      "status_id": "uuid",
      "tags": ["cliente", "ativo"]
    }
  }'

# 2. Criar disparo imediato
curl -X POST /api/v1/broadcasts \
  -d '{
    "name": "Promoção Black Friday",
    "list_id": "list-uuid",
    "message_template": {
      "type": "template",
      "template_id": "promo_bf",
      "variables": {
        "name": "contact.name",
        "discount": "50"
      }
    },
    "rate_limit": 100
  }'

# 3. Executar imediatamente
curl -X POST /api/v1/broadcasts/{id}/execute
```

### Exemplo 2: Newsletter Semanal (Agendado)

```bash
# 1. Criar lista de inscritos
curl -X POST /api/v1/contact-lists \
  -d '{
    "name": "Inscritos Newsletter",
    "is_dynamic": true,
    "filters": {
      "tags": ["newsletter"]
    }
  }'

# 2. Criar disparo agendado para segunda-feira 10h
curl -X POST /api/v1/broadcasts \
  -d '{
    "name": "Newsletter Semanal",
    "list_id": "list-uuid",
    "message_template": {
      "type": "template",
      "template_id": "newsletter_weekly"
    },
    "scheduled_for": "2025-01-20T10:00:00Z",
    "rate_limit": 200
  }'
```

### Exemplo 3: Recuperação de Carrinho (Dinâmico + Agendado)

```bash
# 1. Criar lista dinâmica de carrinhos abandonados
curl -X POST /api/v1/contact-lists \
  -d '{
    "name": "Carrinhos Abandonados Hoje",
    "is_dynamic": true,
    "filters": {
      "custom_fields": {
        "cart_status": "abandoned",
        "abandoned_date": "today"
      }
    }
  }'

# 2. Agendar disparo para 2h após abandono
curl -X POST /api/v1/broadcasts \
  -d '{
    "name": "Recuperação Carrinho 2h",
    "list_id": "list-uuid",
    "message_template": {
      "type": "text",
      "content": "Olá {{name}}! Seu carrinho está te esperando. Finalize agora e ganhe 10% OFF!"
    },
    "scheduled_for": "2025-01-15T14:00:00Z"
  }'
```

---

## Integração com Automation Rules

Broadcasts podem ser disparados via **Automation Rules**:

```go
// Action: trigger_broadcast
{
  "type": "trigger_broadcast",
  "params": {
    "broadcast_id": "uuid",
    "override_schedule": false  // se true, ignora agendamento e dispara imediatamente
  }
}
```

**Exemplo de regra:**
```json
{
  "name": "Enviar newsletter quando status muda para Cliente",
  "trigger": "status.changed",
  "conditions": [
    { "field": "new_status_name", "operator": "eq", "value": "Cliente" }
  ],
  "actions": [
    {
      "type": "trigger_broadcast",
      "params": {
        "broadcast_id": "newsletter_boas_vindas_uuid"
      }
    }
  ]
}
```

---

## Rate Limiting

Controle de velocidade para evitar bloqueios:

```go
type RateLimiter struct {
    rate     int           // mensagens por minuto
    interval time.Duration // intervalo entre mensagens
}

func NewRateLimiter(msgsPerMinute int) *RateLimiter {
    if msgsPerMinute == 0 {
        return &RateLimiter{rate: 0} // sem limite
    }

    interval := time.Minute / time.Duration(msgsPerMinute)
    return &RateLimiter{
        rate:     msgsPerMinute,
        interval: interval,
    }
}

func (r *RateLimiter) Wait() {
    if r.rate == 0 {
        return  // sem limite
    }
    time.Sleep(r.interval)
}
```

**Exemplos:**
- 60 msgs/min → espera 1s entre cada mensagem
- 100 msgs/min → espera 600ms entre cada mensagem
- 0 (sem limite) → dispara o mais rápido possível

---

## Variáveis de Template

Variáveis disponíveis para substituição:

```
{{contact.name}}           → Nome do contato
{{contact.email}}          → Email do contato
{{contact.phone}}          → Telefone do contato
{{contact.custom.X}}       → Campo customizado X
{{broadcast.name}}         → Nome do disparo
{{current_date}}           → Data atual
{{current_time}}           → Hora atual
{{unsubscribe_link}}       → Link para descadastrar
```

**Exemplo:**
```
Olá {{contact.name}}!

Sua última compra foi em {{contact.custom.last_purchase_date}}.
Aproveite nossa promoção exclusiva!

Para não receber mais: {{unsubscribe_link}}
```

---

## Conclusão

Sistema completo de **broadcasts** (disparos em massa) com:

✅ Listas estáticas e dinâmicas
✅ Disparos imediatos e agendados
✅ Templates com variáveis personalizadas
✅ Rate limiting configurável
✅ Tracking detalhado por contato
✅ Integração com Automation Rules
✅ Workers para execução em background

**Próximos passos:**
1. Implementar domain models (ContactList, Broadcast, BroadcastExecution)
2. Implementar repositories
3. Implementar workers (Scheduler + Execution)
4. Implementar HTTP handlers
5. Adicionar métricas e logs
6. Testes E2E
