# 🚀 Opções de Implementação - Detalhamento Completo

**Contexto:** Implementar sistema de Credentials + Meta Integration + Notifiers

---

## 1️⃣ Domain Model `Credential` Completo 🔐

### 📋 O QUE É

Criar a **camada de domínio completa** para gerenciar credenciais de forma segura.

### 🏗️ ESTRUTURA A CRIAR

```
internal/domain/credential/
├── credential.go              # Aggregate Root (250 linhas)
├── credential_type.go         # Value Object - tipos de credenciais (80 linhas)
├── encrypted_value.go         # Value Object - valores criptografados (40 linhas)
├── oauth_token.go             # Value Object - tokens OAuth (120 linhas)
├── repository.go              # Interface do repository (50 linhas)
├── events.go                  # Domain Events (100 linhas)
└── credential_test.go         # Testes unitários (300 linhas)

infrastructure/persistence/entities/
└── credential.go              # Entity ORM (80 linhas)

infrastructure/persistence/
└── gorm_credential_repository.go  # Implementação repository (200 linhas)

infrastructure/crypto/
├── encryptor.go               # Interface (20 linhas)
└── aes_encryptor.go           # Implementação AES-256-GCM (150 linhas)
```

**TOTAL:** ~1.390 linhas de código

### 🎯 FUNCIONALIDADES

#### A. **Credential Aggregate** (`credential.go`)

**Métodos principais:**
```go
// Criação
NewCredential(tenantID, type, name, value, encryptor) → *Credential

// OAuth Management
SetOAuthToken(accessToken, refreshToken, expiresIn) → error
RefreshOAuthToken(newAccessToken, expiresIn) → error
GetAccessToken(encryptor) → (string, error)
GetRefreshToken(encryptor) → (string, error)

// Lifecycle
IsExpired() → bool
NeedsRefresh() → bool  // true se falta < 30min para expirar
MarkAsUsed()
Activate()
Deactivate()

// Security
Decrypt(encryptor) → (string, error)
```

**Exemplo de uso:**
```go
// 1. Criar credencial
encryptor := NewAESEncryptor(encryptionKey)

cred, err := credential.NewCredential(
    "tenant-123",
    credential.CredentialTypeMetaWhatsApp,
    "Meta WhatsApp Production",
    "secret-value-here",
    encryptor,
)

// 2. Adicionar tokens OAuth
err = cred.SetOAuthToken(
    "EAAa1b2c3d4e5f6...",  // access token
    "refresh-token-xyz",    // refresh token
    3600,                   // expires in 1h
    encryptor,
)

// 3. Usar token
if cred.NeedsRefresh() {
    // renovar token
}

accessToken, err := cred.GetAccessToken(encryptor)
// Use accessToken para chamar Meta API
```

#### B. **Credential Types** (`credential_type.go`)

**Tipos suportados:**
```go
const (
    // Meta
    CredentialTypeMetaWhatsApp      = "meta_whatsapp_cloud"
    CredentialTypeMetaConversions   = "meta_conversions_api"
    CredentialTypeMetaAds           = "meta_ads"

    // Google
    CredentialTypeGoogleAds         = "google_ads"
    CredentialTypeGoogleAnalytics   = "google_analytics"

    // Generic
    CredentialTypeWebhook           = "webhook_auth"
    CredentialTypeAPIKey            = "api_key"
    CredentialTypeBasicAuth         = "basic_auth"

    // Internal
    CredentialTypeWAHA              = "waha_instance"
)
```

**Métodos:**
```go
type.RequiresOAuth() → bool
type.GetScopes() → []string
type.IsValid() → bool
```

#### C. **AES Encryptor** (`aes_encryptor.go`)

**Algoritmo:** AES-256-GCM (Galois/Counter Mode)

**Features:**
- ✅ Autenticação (detecta adulteração)
- ✅ Nonce único por mensagem
- ✅ Base64 encoding
- ✅ Resistant to padding oracle attacks

```go
encryptor := NewAESEncryptor(32-byte-key)

// Encrypt
encrypted, err := encryptor.Encrypt("my-secret-token")
// encrypted = EncryptedValue{
//   ciphertext: "aGVsbG8gd29ybGQ=",
//   nonce: "abc123xyz"
// }

// Decrypt
plaintext, err := encryptor.Decrypt(encrypted)
// plaintext = "my-secret-token"
```

#### D. **Domain Events** (`events.go`)

```go
// Events gerados:
CredentialCreatedEvent
CredentialUpdatedEvent
CredentialActivatedEvent
CredentialDeactivatedEvent
OAuthTokenRefreshedEvent
CredentialUsedEvent
CredentialExpiredEvent
```

### 📦 MIGRATIONS NECESSÁRIAS

```sql
-- 000023_create_credentials_table.up.sql
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    project_id UUID,
    credential_type VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Encrypted data
    encrypted_value_ciphertext TEXT NOT NULL,
    encrypted_value_nonce TEXT NOT NULL,

    -- OAuth tokens (encrypted)
    oauth_access_token_ciphertext TEXT,
    oauth_access_token_nonce TEXT,
    oauth_refresh_token_ciphertext TEXT,
    oauth_refresh_token_nonce TEXT,

    -- Metadata
    metadata JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_credentials_tenant ON credentials(tenant_id);
CREATE INDEX idx_credentials_type ON credentials(credential_type);
CREATE INDEX idx_credentials_active ON credentials(is_active) WHERE is_active = true;
CREATE INDEX idx_credentials_expires ON credentials(expires_at) WHERE expires_at IS NOT NULL;
```

### ⏱️ TEMPO ESTIMADO
**4-6 horas** (com testes)

---

## 2️⃣ Meta OAuth Endpoints (`/authorize`, `/callback`) 🔑

### 📋 O QUE É

Criar **endpoints HTTP** para o fluxo OAuth completo com Meta.

### 🏗️ ESTRUTURA A CRIAR

```
internal/application/oauth/
├── meta_oauth_service.go      # Service principal (350 linhas)
├── state_store.go             # Interface para state storage (40 linhas)
└── redis_state_store.go       # Implementação Redis (100 linhas)

infrastructure/http/handlers/
└── oauth_handler.go           # HTTP handlers (200 linhas)

infrastructure/http/routes/
└── oauth_routes.go            # Rotas (30 linhas)
```

**TOTAL:** ~720 linhas de código

### 🎯 ENDPOINTS

#### A. **GET /api/oauth/meta/authorize**

**Propósito:** Iniciar fluxo OAuth, redirecionar para Facebook Login

**Flow:**
```
1. Frontend chama: GET /api/oauth/meta/authorize?tenant_id=xyz
2. Backend:
   - Gera state (random 32 chars) → salva em Redis (5min TTL)
   - Gera code_verifier (random 64 chars)
   - Gera code_challenge = SHA256(code_verifier)
   - Salva state+code_verifier no Redis
3. Backend retorna URL de redirect:
   https://www.facebook.com/v18.0/dialog/oauth?
     client_id=YOUR_APP_ID
     &redirect_uri=https://yourapp.com/api/oauth/meta/callback
     &scope=whatsapp_business_messaging,whatsapp_business_management
     &state=abc123xyz...
     &code_challenge=def456...
     &code_challenge_method=S256
4. Frontend redireciona usuário para essa URL
5. Usuário faz login na Meta e aprova permissões
```

**Código:**
```go
// GET /api/oauth/meta/authorize
func (h *OAuthHandler) InitiateMetaOAuth(c *gin.Context) {
    tenantID := c.Query("tenant_id")

    // Gera auth URL
    authURL, err := h.oauthService.GenerateAuthURL(tenantID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // Retorna URL para frontend redirecionar
    c.JSON(200, gin.H{
        "auth_url": authURL,
    })
}
```

#### B. **GET /api/oauth/meta/callback**

**Propósito:** Receber callback da Meta após aprovação

**Flow:**
```
1. Meta redireciona:
   https://yourapp.com/api/oauth/meta/callback?code=AUTH_CODE&state=abc123

2. Backend:
   - Valida state (busca no Redis)
   - Obtém code_verifier do Redis
   - Troca code por tokens (POST para Meta API)
   - Cria Credential (criptografada)
   - Salva no banco
   - Remove state do Redis

3. Backend redireciona para frontend:
   https://app.yourapp.com/settings/integrations?success=true
```

**Código:**
```go
// GET /api/oauth/meta/callback
func (h *OAuthHandler) HandleMetaCallback(c *gin.Context) {
    code := c.Query("code")
    state := c.Query("state")

    // Processa callback
    credential, err := h.oauthService.HandleCallback(c.Request.Context(), code, state)
    if err != nil {
        // Redireciona com erro
        c.Redirect(302, "https://app.yourapp.com/settings/integrations?error="+err.Error())
        return
    }

    // Sucesso - redireciona
    c.Redirect(302, "https://app.yourapp.com/settings/integrations?success=true&credential_id="+credential.ID().String())
}
```

### 🔐 SECURITY

**PKCE (Proof Key for Code Exchange):**
```
code_verifier = random_string(64)
code_challenge = base64url(SHA256(code_verifier))

Enviar code_challenge → Meta
Meta retorna code
Enviar code + code_verifier → Meta
Meta valida: SHA256(code_verifier) == code_challenge ✅
```

**State (Anti-CSRF):**
```
state = random_string(32)
Salvar state no Redis (5min TTL)
Enviar state → Meta
Meta retorna state
Validar: state existe no Redis? ✅
```

### 📦 DEPENDÊNCIAS

```bash
go get github.com/go-redis/redis/v9  # State storage
```

**Redis config:**
```yaml
redis:
  host: localhost
  port: 6379
  db: 0
  password: ""
```

### ⏱️ TEMPO ESTIMADO
**3-4 horas** (com testes)

---

## 3️⃣ Migration para Expandir `outbox_events` 📊

### 📋 O QUE É

Adicionar **novos campos** à tabela `outbox_events` para suportar notificadores externos.

### 🏗️ ARQUIVOS A CRIAR

```
infrastructure/database/migrations/
├── 000022_add_outbox_event_types.up.sql       # Migration UP (40 linhas)
└── 000022_add_outbox_event_types.down.sql     # Migration DOWN (20 linhas)
```

### 📊 MUDANÇAS NA TABELA

**Campos NOVOS:**

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `event_category` | VARCHAR(50) | Categoria: `domain_event`, `message_delivery`, `webhook`, `meta_conversion` |
| `credential_id` | UUID | ID da credencial (para notifiers) |
| `delivery_attempts` | INT | Contador de tentativas de entrega |
| `next_retry_at` | TIMESTAMP | Quando tentar próxima entrega |

**Migration UP:**
```sql
-- 000022_add_outbox_event_types.up.sql

-- 1. Adiciona event_category
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS event_category VARCHAR(50) NOT NULL DEFAULT 'domain_event';

COMMENT ON COLUMN outbox_events.event_category IS
'Categoria do evento: domain_event (RabbitMQ), message_delivery (WAHA), webhook, meta_conversion, google_ads';

-- 2. Adiciona credential_id (para notifiers externos)
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS credential_id UUID;

ALTER TABLE outbox_events
ADD CONSTRAINT fk_outbox_credential
FOREIGN KEY (credential_id)
REFERENCES credentials(id)
ON DELETE SET NULL;

COMMENT ON COLUMN outbox_events.credential_id IS
'ID da credencial usada para notifiers (Meta, Google Ads, webhooks autenticados)';

-- 3. Adiciona delivery_attempts
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS delivery_attempts INT NOT NULL DEFAULT 0;

-- 4. Adiciona next_retry_at
ALTER TABLE outbox_events
ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP WITH TIME ZONE;

-- 5. Índices para performance
CREATE INDEX IF NOT EXISTS idx_outbox_event_category
ON outbox_events(event_category, status)
WHERE status IN ('pending', 'failed');

CREATE INDEX IF NOT EXISTS idx_outbox_credential
ON outbox_events(credential_id)
WHERE credential_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_retry
ON outbox_events(next_retry_at, status)
WHERE status = 'failed' AND next_retry_at IS NOT NULL;

-- 6. Atualiza eventos existentes
UPDATE outbox_events
SET event_category = 'domain_event'
WHERE event_category IS NULL OR event_category = '';
```

**Migration DOWN:**
```sql
-- 000022_add_outbox_event_types.down.sql

DROP INDEX IF EXISTS idx_outbox_retry;
DROP INDEX IF EXISTS idx_outbox_credential;
DROP INDEX IF EXISTS idx_outbox_event_category;

ALTER TABLE outbox_events DROP CONSTRAINT IF EXISTS fk_outbox_credential;

ALTER TABLE outbox_events DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS delivery_attempts;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS credential_id;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS event_category;
```

### 🔄 UPDATE DO DOMAIN MODEL

```go
// internal/domain/outbox/outbox.go
type OutboxEvent struct {
    // ... campos existentes ...

    // NOVOS CAMPOS:
    EventCategory    EventCategory  // domain_event, message_delivery, etc
    CredentialID     *uuid.UUID     // para notifiers
    DeliveryAttempts int
    NextRetryAt      *time.Time
}

type EventCategory string

const (
    EventCategoryDomain          EventCategory = "domain_event"
    EventCategoryMessageDelivery EventCategory = "message_delivery"
    EventCategoryWebhook         EventCategory = "webhook"
    EventCategoryMetaConversion  EventCategory = "meta_conversion"
    EventCategoryGoogleAds       EventCategory = "google_ads"
)
```

### ⏱️ TEMPO ESTIMADO
**1-2 horas**

---

## 4️⃣ Worker para Processar Notifiers 🔔

### 📋 O QUE É

Criar um **background worker** que:
1. Lê eventos do `outbox_events`
2. Identifica o tipo de evento
3. Roteia para o notificador correto (Meta, Google, Webhook)
4. Envia para API externa
5. Atualiza status (processed/failed)
6. Implementa retry com exponential backoff

### 🏗️ ESTRUTURA COMPLETA

```
internal/application/notifier/
├── notifier_registry.go       # Registry pattern (100 linhas)
├── meta_notifier.go           # Meta Conversions API (250 linhas)
├── google_notifier.go         # Google Ads API (250 linhas)
├── webhook_notifier.go        # Generic webhooks (150 linhas)
└── retry_strategy.go          # Retry logic (80 linhas)

infrastructure/worker/
├── notifier_worker.go         # Worker principal (400 linhas)
├── worker_config.go           # Configuração (50 linhas)
└── worker_metrics.go          # Métricas (Prometheus) (100 linhas)

cmd/notifier-worker/
└── main.go                    # Entrypoint (80 lininas)
```

**TOTAL:** ~1.460 linhas de código

---

### 🔧 ARQUITETURA DETALHADA

```
┌─────────────────────────────────────────────────────────────────┐
│                      NOTIFIER WORKER                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Loop Infinito (a cada 5 segundos):                          │
│     ┌──────────────────────────────────────────────────┐        │
│     │ SELECT * FROM outbox_events                      │        │
│     │ WHERE status = 'pending'                         │        │
│     │   AND (                                          │        │
│     │      event_category IN ('webhook',               │        │
│     │                         'meta_conversion',        │        │
│     │                         'google_ads')             │        │
│     │      OR (status = 'failed'                       │        │
│     │          AND next_retry_at <= NOW())             │        │
│     │   )                                              │        │
│     │ ORDER BY created_at ASC                          │        │
│     │ LIMIT 100                                        │        │
│     └──────────────────────────────────────────────────┘        │
│                        ↓                                         │
│     ┌──────────────────────────────────────────────────┐        │
│     │ Para cada evento:                                │        │
│     │                                                  │        │
│     │  2. UPDATE status = 'processing' (lock)          │        │
│     │                                                  │        │
│     │  3. Identifica tipo (event_category)             │        │
│     │                                                  │        │
│     │  4. Roteia para Notifier correto:               │        │
│     │     ┌───────────────────────────────────┐        │        │
│     │     │ NotifierRegistry.Get(type)        │        │        │
│     │     │   ├─ meta_conversion              │        │        │
│     │     │   │   → MetaConversionsNotifier   │        │        │
│     │     │   ├─ google_ads                   │        │        │
│     │     │   │   → GoogleAdsNotifier         │        │        │
│     │     │   └─ webhook                      │        │        │
│     │     │       → WebhookNotifier           │        │        │
│     │     └───────────────────────────────────┘        │        │
│     │                                                  │        │
│     │  5. Busca Credential (se credential_id)          │        │
│     │                                                  │        │
│     │  6. Notifier.Notify(event, credential)           │        │
│     │     ┌────────────────────────────────────┐       │        │
│     │     │ • Valida dados                     │       │        │
│     │     │ • Descriptografa access token      │       │        │
│     │     │ • Chama API externa (HTTP POST)    │       │        │
│     │     │ • Retorna sucesso/erro             │       │        │
│     │     └────────────────────────────────────┘       │        │
│     │                                                  │        │
│     │  7. Atualiza status:                             │        │
│     │     ├─ Sucesso:                                  │        │
│     │     │   UPDATE status = 'processed',             │        │
│     │     │          processed_at = NOW()              │        │
│     │     │                                            │        │
│     │     └─ Erro:                                     │        │
│     │         delivery_attempts++                      │        │
│     │         IF attempts < max_retries:               │        │
│     │            status = 'failed'                     │        │
│     │            next_retry_at = NOW() +               │        │
│     │               (backoff * 2^attempts)             │        │
│     │         ELSE:                                    │        │
│     │            status = 'dead_letter'                │        │
│     │            → Move para DLQ                       │        │
│     └──────────────────────────────────────────────────┘        │
│                                                                  │
│  8. Métricas (Prometheus):                                      │
│     • notifier_events_processed_total{type, status}             │
│     • notifier_processing_duration_seconds{type}                │
│     • notifier_errors_total{type, error_type}                   │
│     • notifier_queue_size{category}                             │
│                                                                  │
│  9. Graceful Shutdown:                                          │
│     • SIGTERM → Finaliza eventos em processamento              │
│     • Timeout 30s → Force quit                                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

### 📝 CÓDIGO DETALHADO

#### A. **Notifier Interface**

```go
// internal/application/notifier/notifier.go
package notifier

import (
    "context"
    "github.com/google/uuid"
)

type Notifier interface {
    // Notify envia evento para sistema externo
    Notify(ctx context.Context, event NotificationEvent) error

    // Type retorna o tipo do notifier
    Type() NotifierType

    // Validate valida evento antes de enviar
    Validate(event NotificationEvent) error
}

type NotificationEvent struct {
    EventID      uuid.UUID
    TenantID     string
    EventType    string
    EventData    map[string]interface{}
    CredentialID *uuid.UUID
}

type NotifierType string

const (
    NotifierTypeMeta    NotifierType = "meta_conversions"
    NotifierTypeGoogle  NotifierType = "google_ads"
    NotifierTypeWebhook NotifierType = "webhook"
)
```

#### B. **Meta Conversions Notifier**

```go
// internal/application/notifier/meta_notifier.go
package notifier

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type MetaConversionsNotifier struct {
    credentialRepo credential.Repository
    encryptor      credential.Encryptor
    pixelID        string
    httpClient     *http.Client
}

func (n *MetaConversionsNotifier) Notify(
    ctx context.Context,
    event NotificationEvent,
) error {
    // 1. Busca credencial
    cred, err := n.credentialRepo.FindByID(*event.CredentialID)
    if err != nil {
        return fmt.Errorf("credential not found: %w", err)
    }

    // 2. Checa se expirou/precisa renovar
    if cred.IsExpired() {
        return fmt.Errorf("credential expired")
    }

    if cred.NeedsRefresh() {
        // Worker de refresh vai cuidar disso
        return fmt.Errorf("credential needs refresh")
    }

    // 3. Descriptografa access token
    accessToken, err := cred.GetAccessToken(n.encryptor)
    if err != nil {
        return fmt.Errorf("failed to decrypt token: %w", err)
    }

    // 4. Monta payload Meta Conversions API
    payload := map[string]interface{}{
        "data": []map[string]interface{}{
            {
                "event_name":       event.EventData["event_name"],
                "event_time":       event.EventData["event_time"],
                "event_id":         event.EventID.String(),
                "event_source_url": event.EventData["source_url"],
                "user_data":        event.EventData["user_data"],
                "custom_data":      event.EventData["custom_data"],
                "action_source":    "website",
            },
        },
    }

    // 5. Envia para Meta
    url := fmt.Sprintf(
        "https://graph.facebook.com/v18.0/%s/events?access_token=%s",
        n.pixelID,
        accessToken,
    )

    jsonData, _ := json.Marshal(payload)

    req, err := http.NewRequestWithContext(
        ctx,
        "POST",
        url,
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := n.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("http request failed: %w", err)
    }
    defer resp.Body.Close()

    // 6. Valida resposta
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        var errResp map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&errResp)
        return fmt.Errorf(
            "Meta API error (status %d): %v",
            resp.StatusCode,
            errResp,
        )
    }

    // 7. Marca credencial como usada
    cred.MarkAsUsed()
    _ = n.credentialRepo.Save(cred)

    return nil
}
```

#### C. **Notifier Registry**

```go
// internal/application/notifier/notifier_registry.go
package notifier

import (
    "context"
    "fmt"
    "sync"
)

type NotifierRegistry struct {
    notifiers map[NotifierType]Notifier
    mu        sync.RWMutex
}

func NewNotifierRegistry() *NotifierRegistry {
    return &NotifierRegistry{
        notifiers: make(map[NotifierType]Notifier),
    }
}

func (r *NotifierRegistry) Register(notifier Notifier) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.notifiers[notifier.Type()] = notifier
}

func (r *NotifierRegistry) Get(notifierType NotifierType) (Notifier, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    notifier, exists := r.notifiers[notifierType]
    if !exists {
        return nil, fmt.Errorf("notifier not found: %s", notifierType)
    }

    return notifier, nil
}

func (r *NotifierRegistry) Notify(
    ctx context.Context,
    notifierType NotifierType,
    event NotificationEvent,
) error {
    notifier, err := r.Get(notifierType)
    if err != nil {
        return err
    }

    // Valida antes de enviar
    if err := notifier.Validate(event); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // Envia
    return notifier.Notify(ctx, event)
}
```

#### D. **Worker Principal**

```go
// infrastructure/worker/notifier_worker.go
package worker

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/cawolfkreo/ventros-crm/internal/application/notifier"
    "github.com/cawolfkreo/ventros-crm/internal/domain/outbox"
)

type NotifierWorker struct {
    outboxRepo       outbox.Repository
    notifierRegistry *notifier.NotifierRegistry
    config           WorkerConfig
    metrics          *WorkerMetrics
    stopChan         chan struct{}
}

type WorkerConfig struct {
    PollInterval    time.Duration  // 5 segundos
    BatchSize       int            // 100 eventos
    MaxRetries      int            // 3 tentativas
    RetryBackoff    time.Duration  // 1 minuto
    ProcessTimeout  time.Duration  // 30 segundos por evento
}

func (w *NotifierWorker) Start(ctx context.Context) error {
    log.Println("🚀 Notifier Worker started")

    ticker := time.NewTicker(w.config.PollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            log.Println("🛑 Notifier Worker stopping...")
            return nil

        case <-ticker.C:
            w.processBatch(ctx)
        }
    }
}

func (w *NotifierWorker) processBatch(ctx context.Context) {
    // 1. Busca eventos pendentes
    events, err := w.outboxRepo.GetPendingNotifierEvents(
        ctx,
        w.config.BatchSize,
    )
    if err != nil {
        log.Printf("❌ Error fetching events: %v", err)
        return
    }

    if len(events) == 0 {
        return
    }

    log.Printf("📦 Processing %d notifier events", len(events))

    // 2. Processa cada evento
    for _, event := range events {
        w.processEvent(ctx, event)
    }
}

func (w *NotifierWorker) processEvent(
    ctx context.Context,
    event *outbox.OutboxEvent,
) {
    eventCtx, cancel := context.WithTimeout(ctx, w.config.ProcessTimeout)
    defer cancel()

    startTime := time.Now()

    // 1. Marca como processing (pessimistic lock)
    if err := w.outboxRepo.MarkAsProcessing(eventCtx, event.EventID); err != nil {
        log.Printf("⚠️  Failed to lock event %s: %v", event.EventID, err)
        return
    }

    // 2. Identifica tipo de notifier
    notifierType, err := w.getNotifierType(event.EventCategory)
    if err != nil {
        w.markAsFailed(eventCtx, event, err.Error())
        return
    }

    // 3. Converte event data
    var eventData map[string]interface{}
    if err := json.Unmarshal(event.EventData, &eventData); err != nil {
        w.markAsFailed(eventCtx, event, fmt.Sprintf("invalid event data: %v", err))
        return
    }

    // 4. Cria NotificationEvent
    notifEvent := notifier.NotificationEvent{
        EventID:      event.EventID,
        TenantID:     *event.TenantID,
        EventType:    event.EventType,
        EventData:    eventData,
        CredentialID: event.CredentialID,
    }

    // 5. Envia via notifier
    err = w.notifierRegistry.Notify(eventCtx, notifierType, notifEvent)

    duration := time.Since(startTime)

    // 6. Atualiza status
    if err != nil {
        log.Printf("❌ Event %s failed: %v", event.EventID, err)

        // Incrementa tentativas
        event.DeliveryAttempts++

        if event.DeliveryAttempts < w.config.MaxRetries {
            // Agenda retry com exponential backoff
            backoff := w.config.RetryBackoff * time.Duration(1 << event.DeliveryAttempts)
            nextRetry := time.Now().Add(backoff)

            w.outboxRepo.MarkForRetry(eventCtx, event.EventID, err.Error(), nextRetry)

            log.Printf("🔄 Event %s scheduled for retry in %v (attempt %d/%d)",
                event.EventID,
                backoff,
                event.DeliveryAttempts,
                w.config.MaxRetries,
            )
        } else {
            // Max retries atingido → Dead Letter Queue
            w.outboxRepo.MarkAsDeadLetter(eventCtx, event.EventID, err.Error())

            log.Printf("💀 Event %s moved to DLQ after %d attempts",
                event.EventID,
                event.DeliveryAttempts,
            )

            // TODO: Alertar via Slack/email
        }

        w.metrics.RecordError(string(notifierType), err)
    } else {
        log.Printf("✅ Event %s processed successfully", event.EventID)

        w.outboxRepo.MarkAsProcessed(eventCtx, event.EventID)
        w.metrics.RecordSuccess(string(notifierType), duration)
    }
}

func (w *NotifierWorker) getNotifierType(
    category outbox.EventCategory,
) (notifier.NotifierType, error) {
    switch category {
    case outbox.EventCategoryMetaConversion:
        return notifier.NotifierTypeMeta, nil
    case outbox.EventCategoryGoogleAds:
        return notifier.NotifierTypeGoogle, nil
    case outbox.EventCategoryWebhook:
        return notifier.NotifierTypeWebhook, nil
    default:
        return "", fmt.Errorf("unknown category: %s", category)
    }
}
```

#### E. **Main Entrypoint**

```go
// cmd/notifier-worker/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/cawolfkreo/ventros-crm/infrastructure/worker"
    "github.com/cawolfkreo/ventros-crm/internal/application/notifier"
)

func main() {
    log.Println("🚀 Starting Notifier Worker...")

    // 1. Inicializa dependências
    db := initDatabase()
    encryptor := initEncryptor()

    outboxRepo := persistence.NewGormOutboxRepository(db)
    credentialRepo := persistence.NewGormCredentialRepository(db)

    // 2. Cria notifiers
    registry := notifier.NewNotifierRegistry()

    metaNotifier := notifier.NewMetaConversionsNotifier(
        credentialRepo,
        encryptor,
        os.Getenv("META_PIXEL_ID"),
    )
    registry.Register(metaNotifier)

    googleNotifier := notifier.NewGoogleAdsNotifier(credentialRepo, encryptor)
    registry.Register(googleNotifier)

    webhookNotifier := notifier.NewWebhookNotifier(credentialRepo, encryptor)
    registry.Register(webhookNotifier)

    // 3. Cria worker
    config := worker.WorkerConfig{
        PollInterval:   5 * time.Second,
        BatchSize:      100,
        MaxRetries:     3,
        RetryBackoff:   1 * time.Minute,
        ProcessTimeout: 30 * time.Second,
    }

    metrics := worker.NewWorkerMetrics()

    notifierWorker := worker.NewNotifierWorker(
        outboxRepo,
        registry,
        config,
        metrics,
    )

    // 4. Inicia worker
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        if err := notifierWorker.Start(ctx); err != nil {
            log.Fatalf("Worker error: %v", err)
        }
    }()

    // 5. Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    <-sigChan
    log.Println("🛑 Shutdown signal received")

    cancel()

    // Aguarda 30s para processar eventos em andamento
    time.Sleep(30 * time.Second)

    log.Println("👋 Notifier Worker stopped")
}
```

---

### 🎯 FEATURES DO WORKER

#### ✅ **Retry com Exponential Backoff**

```
Tentativa 1: erro → retry em 1 min
Tentativa 2: erro → retry em 2 min
Tentativa 3: erro → retry em 4 min
Tentativa 4: DLQ (Dead Letter Queue)
```

#### ✅ **Graceful Shutdown**

```
SIGTERM recebido
  → Para de buscar novos eventos
  → Aguarda eventos em processamento (30s timeout)
  → Sai
```

#### ✅ **Métricas (Prometheus)**

```go
// Métricas expostas em /metrics
notifier_events_processed_total{type="meta_conversions", status="success"} 1234
notifier_events_processed_total{type="meta_conversions", status="failed"} 12
notifier_processing_duration_seconds{type="meta_conversions"} 0.245
notifier_queue_size{category="meta_conversion"} 42
```

#### ✅ **Dead Letter Queue**

Eventos que falharam após `max_retries`:
```sql
SELECT * FROM outbox_events
WHERE status = 'dead_letter'
ORDER BY created_at DESC;
```

---

### 📦 DEPLOYMENT

**Docker Compose:**
```yaml
services:
  notifier-worker:
    build: .
    command: /app/notifier-worker
    environment:
      - DATABASE_URL=postgres://...
      - ENCRYPTION_KEY=base64...
      - META_PIXEL_ID=123456789
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
```

**Kubernetes:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: notifier-worker
spec:
  replicas: 2  # Escala horizontal
  selector:
    matchLabels:
      app: notifier-worker
  template:
    spec:
      containers:
      - name: worker
        image: ventros-crm/notifier-worker:latest
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

---

### ⏱️ TEMPO ESTIMADO
**8-10 horas** (com testes completos)

---

## 📊 RESUMO COMPARATIVO

| Opção | Complexidade | Tempo | Valor | Prioridade |
|-------|--------------|-------|-------|------------|
| **1. Credential Domain** | Média | 4-6h | 🔴 Crítico | 1º |
| **2. Meta OAuth** | Média | 3-4h | 🔴 Crítico | 2º |
| **3. Migration Outbox** | Baixa | 1-2h | 🟡 Alta | 3º |
| **4. Notifier Worker** | Alta | 8-10h | 🔴 Crítico | 4º |

**TOTAL:** 16-22 horas (~3-4 dias de trabalho)

---

**Qual você quer implementar primeiro?** 🚀