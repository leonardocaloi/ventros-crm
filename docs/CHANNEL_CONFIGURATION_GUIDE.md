# 📡 Channel Configuration Guide - Complete Reference

**Version**: 1.0  
**Last Updated**: 2025-10-16  
**Purpose**: Comprehensive guide for all channel types, configurations, and validation

---

## 📋 Supported Channel Types

| Type | Status | Webhook Support | Import Support | Description |
|------|--------|-----------------|----------------|-------------|
| `waha` | ✅ Production | ✅ Yes | ✅ Yes | Manual WAHA (bring your own server) |
| `whatsapp_business` | 🟡 Planned | ✅ Yes | ✅ Yes | Auto-managed WhatsApp (system creates session) |
| `whatsapp` | 🔴 Not Implemented | ✅ Yes | ❌ No | WhatsApp Cloud API |
| `messenger` | 🔴 Not Implemented | ✅ Yes | ❌ No | Facebook Messenger |
| `instagram` | 🔴 Not Implemented | ✅ Yes | ❌ No | Instagram DM |
| `telegram` | 🔴 Not Implemented | ✅ Yes | ❌ No | Telegram Bot |
| `wechat` | 🔴 Not Implemented | ✅ Yes | ❌ No | WeChat Official Account |
| `twilio_sms` | 🔴 Not Implemented | ✅ Yes | ❌ No | Twilio SMS |
| `web_form` | 🔴 Not Implemented | ✅ Yes | ❌ No | Web Form / Webhook |

---

## 🏗️ Channel Fields (Generic - All Types)

### ✅ OBRIGATÓRIOS (Nível de Canal)

| Campo | Tipo | Descrição | Exemplo |
|-------|------|-----------|---------|
| `name` | string | Nome do canal | `"WhatsApp Suporte"` |
| `type` | enum | Tipo do canal | `"waha"` |

**Campos Automáticos** (preenchidos pelo sistema):
- `id` (UUID - gerado automaticamente)
- `user_id` (UUID - do JWT)
- `project_id` (UUID - do JWT)
- `tenant_id` (string - do JWT)
- `status` (`"inactive"` inicial)
- `created_at` (timestamp)
- `updated_at` (timestamp)

---

### ⚙️ OPCIONAIS (Nível de Canal)

| Campo | Tipo | Default | Descrição | Dependências |
|-------|------|---------|-----------|-------------|
| `session_timeout_minutes` | int | `30` | Tempo sem mensagem para encerrar sessão | Nenhuma |
| `allow_groups` | bool | `false` | Aceitar mensagens de grupos WhatsApp | Nenhuma |
| `tracking_enabled` | bool | `true` | Rastrear origem das mensagens | Nenhuma |
| `pipeline_id` | UUID | `null` | Pipeline padrão do canal (pode escolher outro depois) | Nenhuma |
| `ai_enabled` | bool | `false` | ✅ Ativa processamento de tipos de conteúdo (text, audio, image, video, document) | Nenhuma |
| `ai_agents_enabled` | bool | `false` | ✅ Ativa agentes IA + debouncer + **memória obrigatória** | Requer: `ai_enabled=true` |
| `debounce_timeout_ms` | int | `15000` | ⚙️ Tempo para agrupar mensagens (buffer inteligente) | Ativo apenas se `ai_agents_enabled=true` |

**⚠️ IMPORTANTE - Hierarquia de Dependências**:
```
ai_enabled = false
  └─> Nenhum processamento de IA
  └─> Mensagens passam direto (sem buffer)

ai_enabled = true
  ├─> Ativa processamento por tipo de conteúdo:
  │   ├─> text (Claude/GPT) - passa direto
  │   ├─> audio (Whisper)
  │   ├─> image (Gemini Vision)
  │   ├─> video (extração)
  │   ├─> document (LlamaParse) - PDF → Memória
  │   └─> voice/PTT (Whisper prioritário)
  └─> Se mensagem é SÓ texto → passa direto
  └─> Se mensagem tem mídia ativada → aguarda no buffer

ai_agents_enabled = true
  ├─> REQUER: ai_enabled = true
  ├─> ATIVA: debounce_timeout_ms (default: 15000ms)
  ├─> ATIVA: Buffer inteligente (agrupa mensagens na ordem)
  ├─> OBRIGATÓRIO: Sistema de Memória (configurado a nível de agente)
  └─> Agentes podem:
      ├─> Consultar memória (vector + keyword + graph)
      ├─> Criar fatos de memória
      └─> Usar contexto histórico do contato
```

**Fluxo de Mensagens**:
```
1. Mensagem chega do webhook
   ↓
2. ai_enabled?
   ├─> NÃO: Processa direto (sem IA)
   └─> SIM: Verifica tipo de conteúdo
       ↓
3. Tipo de conteúdo habilitado?
   ├─> text: Passa direto (sem buffer)
   ├─> mídia (audio/image/video/doc):
       ↓
4. ai_agents_enabled?
   ├─> NÃO: Processa mídia direto
   └─> SIM: Aguarda no buffer (debounce_timeout_ms)
       ├─> Agrupa mensagens sequenciais
       ├─> Ordena corretamente
       └─> Envia para agente IA
           ↓
5. Agente IA processa:
   ├─> Consulta memória do contato
   ├─> Processa com contexto
   ├─> Gera resposta
   └─> Cria fatos de memória
```

---

### 🤖 AI Processing Config (por Tipo de Conteúdo)

**Apenas ativo se `ai_enabled = true`**

Cada tipo de conteúdo pode ser configurado individualmente:

| Tipo | Provider | Model | Priority | Debounce | Max Size | Notas |
|------|----------|-------|----------|----------|----------|-------|
| `text` | `anthropic` | `claude-3-5-sonnet` | 5 | 1000ms | 1MB | Passa direto (sem buffer) |
| `audio` | `openai` | `whisper-1` | 8 | 500ms | 25MB | Split long audio: sim |
| `voice` (PTT) | `openai` | `whisper-1` | **10** | 100ms | 25MB | **Máxima prioridade** |
| `image` | `google` | `gemini-1.5-pro` | 7 | 1000ms | 10MB | Vision processing |
| `video` | `openai` | `gpt-4-vision` | 3 | 5000ms | 100MB | Extração + frames |
| `document` | `llamaparse` | `default` | 6 | 2000ms | 50MB | **PDF → Memória** |

**Configuração por Tipo** (opcional):
```json
{
  "ai_enabled": true,
  "ai_processing_config": {
    "text": {
      "enabled": true,
      "provider": "anthropic",
      "model": "claude-3-5-sonnet-20241022",
      "priority": 5,
      "debounce_ms": 1000
    },
    "voice": {
      "enabled": true,
      "provider": "openai",
      "model": "whisper-1",
      "priority": 10,
      "debounce_ms": 100,
      "max_size_bytes": 26214400
    },
    "document": {
      "enabled": true,
      "provider": "llamaparse",
      "priority": 6,
      "debounce_ms": 2000
    }
  }
}
```

**Se não configurado**: Usa defaults acima.

---

### 📥 OPCIONAIS (History Import)

| Campo | Tipo | Default | Descrição |
|-------|------|---------|-----------|
| `history_import_enabled` | bool | `false` | Habilitar importação de histórico |
| `history_import_max_days` | int? | `null` | Limite de dias (null = ilimitado) |
| `history_import_max_messages_chat` | int? | `null` | Limite msgs/chat (null = ilimitado) |
| `default_agent_id` | UUID? | `null` | Agente padrão para import |

---

### 🔒 READ-ONLY (Calculados pelo Sistema)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `status` | enum | `inactive`, `activating`, `active`, `error` |
| `external_id` | string | ID externo (ex: session_id do WAHA) |
| `webhook_id` | string | ID do webhook configurado |
| `webhook_url` | string | URL do webhook |
| `webhook_active` | bool | Se webhook está funcionando |
| `webhook_configured_at` | timestamp? | Quando webhook foi configurado |
| `messages_received` | int | Total de mensagens recebidas |
| `messages_sent` | int | Total de mensagens enviadas |
| `last_message_at` | timestamp? | Última mensagem |
| `last_error_at` | timestamp? | Último erro |
| `last_error` | string | Mensagem do último erro |
| `history_import_status` | enum | `idle`, `importing`, `completed`, `failed` |
| `history_import_messages_count` | int | Total de mensagens importadas |
| `last_import_date` | timestamp? | Última importação |

---

## 📱 WAHA Channel Configuration

### ✅ OBRIGATÓRIOS (WAHA Adapter)

| Campo | Tipo | Validação | Descrição | Exemplo |
|-------|------|-----------|-----------|---------|
| `base_url` | string | URL válida | Servidor WAHA | `"https://waha.ventros.cloud"` |
| `api_key` **OU** `token` | string | Não vazio | Autenticação WAHA | `"4bffec302d..."` |
| `session_id` | string | Não vazio | ID da sessão WhatsApp | `"freefaro-b2b-comercial"` |

**Regras de Validação**:
```go
// Domain: WAHAActivationStrategy.CanActivate()

if config.BaseURL == "" {
    return error("WAHA base_url is required")
}

if config.Auth.APIKey == "" && config.Auth.Token == "" {
    return error("WAHA authentication (api_key or token) is required")
}

if config.SessionID == "" {
    return error("WAHA session_id is required")
}
```

---

### ⚙️ OPCIONAIS (WAHA Adapter)

| Campo | Tipo | Default | Descrição | Quando Usar |
|-------|------|---------|-----------|-------------|
| `webhook_url` | string | `null` | URL para receber webhooks | **Produção**: Obrigatório<br>**Teste Import**: Opcional |
| `import_strategy` | enum | `"none"` | `none`, `new_only`, `all` | Controle de importação |
| `import_completed` | bool | `false` | Se import já foi feito | Interno (calculado) |

---

## 🔄 Fluxo Completo de Criação de Canal

### 1️⃣ **Criar Canal** (POST `/api/v1/crm/channels`)

```json
// Request
{
  // ✅ OBRIGATÓRIO - Nível Canal
  "name": "FreeFaro B2B Comercial",
  "type": "waha",
  
  // ⚙️ OPCIONAL - Nível Canal
  "session_timeout_minutes": 120,  // 2 horas
  "allow_groups": false,
  "tracking_enabled": true,
  "history_import_enabled": true,
  "history_import_max_days": 30,
  
  // ✅ OBRIGATÓRIO - Adapter WAHA
  "waha_config": {
    "base_url": "https://waha.ventros.cloud",
    "api_key": "4bffec302d5f4312b8b73700da3ff3cb",
    "session_id": "freefaro-b2b-comercial",
    
    // ⚙️ OPCIONAL - Webhook (depende do cenário)
    "webhook_url": "https://xyz.ngrok.io/api/v1/webhooks/waha"
  }
}

// Response
{
  "message": "Channel created successfully",
  "id": "uuid-do-canal",
  "channel": {
    "id": "uuid",
    "name": "FreeFaro B2B Comercial",
    "type": "waha",
    "status": "inactive",  // Começa inativo
    "external_id": "freefaro-b2b-comercial",
    "created_at": "2025-10-16T06:00:00Z",
    // ...
  }
}
```

---

### 2️⃣ **Ativar Canal** (POST `/api/v1/crm/channels/:id/activate`)

**Fluxo Interno**:

```
1. Validação Pré-Ativação (CanActivate)
   ├─ Validar base_url ✅
   ├─ Validar auth (api_key ou token) ✅
   ├─ Validar session_id ✅
   └─ Status: inactive → activating

2. Event Published: channel.activation.requested
   └─ RabbitMQ Consumer processa assincronamente

3. Health Check WAHA (ASYNC)
   ├─ GET /api/sessions/{session_id}/status
   ├─ Verificar status == "WORKING" ✅
   └─ Se falhar → status: error

4. (OPCIONAL) Validação de Webhook
   ├─ SE webhook_url configurado
   ├─ E SE skip_webhook_validation == false
   ├─ ENTÃO: Enviar mensagem teste
   ├─ Aguardar webhook voltar (10s timeout)
   └─ Se falhar → status: error

5. Ativação Completa
   ├─ Status: activating → active ✅
   └─ Event Published: channel.activated
```

**Request**:
```bash
POST /api/v1/crm/channels/{channel_id}/activate
Authorization: Bearer {api_key}
```

**Response** (202 Accepted - Async):
```json
{
  "message": "Channel activation requested",
  "status": "activating",
  "correlation_id": "uuid"
}
```

---

### 3️⃣ **Monitorar Ativação** (GET `/api/v1/crm/channels/:id`)

```bash
GET /api/v1/crm/channels/{channel_id}
Authorization: Bearer {api_key}

# Polling a cada 1-2 segundos
# Até status == "active" ou "error"
```

**Response**:
```json
{
  "id": "uuid",
  "name": "FreeFaro B2B Comercial",
  "status": "active",  // ← Mudou!
  "webhook_active": true,
  "last_message_at": null,
  // ...
}
```

---

## 🎯 Validação de Webhook - Quando Aplicar?

### Por Tipo de Canal:

| Tipo de Canal | Webhook Obrigatório? | Validação Ping-Pong? | Notas |
|---------------|---------------------|----------------------|-------|
| `waha` | ⚠️ Depende | ✅ Sim (futuro) | Produção: SIM<br>Teste Import: NÃO |
| `whatsapp` (Cloud) | ✅ Sempre | ✅ Sim | Cloud API requer webhook |
| `messenger` | ✅ Sempre | ✅ Sim | Facebook requer webhook |
| `instagram` | ✅ Sempre | ✅ Sim | Facebook requer webhook |
| `telegram` | ✅ Sempre | ✅ Sim | Telegram requer webhook |
| `wechat` | ✅ Sempre | ❌ Não | WeChat não valida |
| `twilio_sms` | ✅ Sempre | ✅ Sim | Twilio requer webhook |
| `web_form` | ✅ Sempre | ❌ Não | Apenas recebe POST |

---

### Implementação Sugerida (Multi-Canal):

```go
// internal/application/channel/activation/strategy.go

type ActivationStrategy interface {
    CanActivate(ctx context.Context, ch *channel.Channel) error
    Activate(ctx context.Context, ch *channel.Channel) error
    HealthCheck(ctx context.Context, ch *channel.Channel) (bool, string, error)
    
    // NOVO: Validação de webhook
    RequiresWebhook() bool  // Se webhook é obrigatório
    ValidateWebhook(ctx context.Context, ch *channel.Channel) error  // Ping-pong test
    
    Compensate(ctx context.Context, ch *channel.Channel) error
}

// Implementações:
// - WAHAActivationStrategy      → RequiresWebhook() = false (opcional)
// - WhatsAppCloudStrategy       → RequiresWebhook() = true (obrigatório)
// - MessengerStrategy           → RequiresWebhook() = true (obrigatório)
// - TelegramStrategy            → RequiresWebhook() = true (obrigatório)
```

---

## 📊 Cenários de Uso - WAHA

### Cenário 1: **Produção (com webhook)**

```json
{
  "name": "Suporte WhatsApp",
  "type": "waha",
  "session_timeout_minutes": 120,
  "history_import_enabled": false,
  "waha_config": {
    "base_url": "https://waha.prod.com",
    "api_key": "xxx",
    "session_id": "suporte-prod",
    "webhook_url": "https://crm.prod.com/api/v1/webhooks/waha"  // ✅ COM webhook
  }
}
```

**Ativação**:
- ✅ Valida base_url, auth, session_id
- ✅ Health check WAHA
- ✅ **Valida webhook** (ping-pong test)
- ✅ Status: `active`

---

### Cenário 2: **Teste de Import (sem webhook)**

```json
{
  "name": "FreeFaro Import Test",
  "type": "waha",
  "session_timeout_minutes": 120,
  "history_import_enabled": true,
  "history_import_max_days": 30,
  "waha_config": {
    "base_url": "https://waha.ventros.cloud",
    "api_key": "xxx",
    "session_id": "freefaro-b2b-comercial"
    // ❌ SEM webhook_url (ou null)
  }
}
```

**Ativação**:
- ✅ Valida base_url, auth, session_id
- ✅ Health check WAHA
- ⚠️ **Skip validação webhook** (não configurado)
- ✅ Status: `active`

**Import**:
```bash
POST /api/v1/crm/channels/{id}/import-history
{
  "strategy": "time_range",
  "time_range_days": 30,
  "limit": 0
}
```

---

### Cenário 3: **Desenvolvimento Local (com tunnel)**

```json
{
  "name": "Dev Local",
  "type": "waha",
  "waha_config": {
    "base_url": "https://waha.ventros.cloud",
    "api_key": "xxx",
    "session_id": "dev-test",
    "webhook_url": "https://abc.ngrok.io/api/v1/webhooks/waha"  // ✅ Tunnel
  }
}
```

**Setup**:
```bash
# Terminal 1: Tunnel
make crm.run.tunnel
# Output: https://abc.ngrok.io

# Terminal 2: Criar canal com webhook apontando para tunnel
```

---

## 🔧 Próxima Implementação Sugerida

### 1. **Flag `skip_webhook_validation`** (Temporária)

```json
{
  "name": "Test Channel",
  "type": "waha",
  "skip_webhook_validation": true,  // ← Bypass temporário
  "waha_config": {
    "webhook_url": "https://xyz.com/webhook"
  }
}
```

---

### 2. **Webhook Ping-Pong Validation** (Produção)

```go
// internal/application/channel/activation/waha_strategy.go

func (s *WAHAActivationStrategy) ValidateWebhook(
    ctx context.Context, 
    ch *channel.Channel,
) error {
    config, _ := ch.GetWAHAConfig()
    
    // 1. Gerar callback ID único
    callbackID := uuid.New().String()
    
    // 2. Enviar mensagem de teste
    testMsg := fmt.Sprintf("🔔 Webhook Test - %s", callbackID)
    wahaClient.SendText(ctx, waha.SendTextRequest{
        ChatID: config.SessionID + "@c.us",  // Para si mesmo
        Text:   testMsg,
    })
    
    // 3. Aguardar webhook voltar (timeout 10s)
    select {
    case <-webhookConfirmed:
        return nil  // ✅ Webhook OK
    case <-time.After(10 * time.Second):
        return error("webhook timeout")  // ❌ Falhou
    }
}
```

---

## 📚 Referências

- **Domain**: `internal/domain/crm/channel/channel.go`
- **Handler**: `infrastructure/http/handlers/channel_handler.go`
- **Activation**: `internal/application/channel/activation/waha_strategy.go`
- **WAHA Client**: `infrastructure/channels/waha/client.go`
- **Tests**: `tests/e2e/waha_history_import_test.go`

---

**Version**: 1.0  
**Status**: Production-Ready (WAHA only)  
**Next**: Implement webhook validation for all channel types
