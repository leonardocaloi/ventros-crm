# WAHA Event Routing Logic

## 🎯 Lógica de Roteamento Condicional

Baseado nos eventos reais que você tem, aqui está como o processor decide **o que fazer** com cada tipo:

## 📊 Fluxo de Decisão

```
📥 waha.events.raw
       ↓
   [Parse JSON]
       ↓
   [Validações]
       ↓
  [Switch Event Type] ← CONDICIONAIS AQUI
       ↓
   [Roteamento]
```

## 🔀 Condicionais de Roteamento

### 1. **Mensagens** (`"message"`, `"message.any"`)
```go
case "message", "message.any":
    return p.processMessageEvent(ctx, rawEvent, wahaEvent)
```

**O que faz:**
- ✅ **PROCESSA COMPLETO** via `wahaMessageService.ProcessWAHAMessage()`
- ✅ Salva no banco, cria contato, inicia sessão
- ✅ Extrai dados de conversão de FB Ads
- ❌ Se falhar → vai para `waha.events.message.parsed` (retry)

**Tipos incluídos:**
- 📝 Texto simples
- 🖼️ Imagens 
- 🎵 Áudio normal
- 🎤 **PTT (Push-to-Talk)** ← Era o erro!
- 📄 Documentos
- 👤 Contatos
- 📍 Localização
- 💰 **FB Ads** ← Crítico para negócio!

### 2. **ACKs de Mensagem** (`"message.ack"`)
```go
case "message.ack":
    return p.processMessageAckEvent(ctx, rawEvent, wahaEvent)
```

**O que faz:**
- ⏭️ **SÓ LOGA** (TODO: implementar)
- ➡️ Vai para `waha.events.message.parsed`

### 3. **Chamadas** (`"call.received"`, `"call.accepted"`, `"call.rejected"`)
```go
case "call.received", "call.accepted", "call.rejected":
    return p.processCallEvent(ctx, rawEvent, wahaEvent)
```

**O que faz:**
- ⏭️ **SÓ LOGA** (TODO: implementar)
- ➡️ Vai para `waha.events.call.parsed`

### 4. **Labels/Tags** (`"label.upsert"`, `"label.deleted"`, etc.)
```go
case "label.upsert", "label.deleted", "label.chat.added", "label.chat.deleted":
    return p.processLabelEvent(ctx, rawEvent, wahaEvent)
```

**O que faz:**
- ⏭️ **SÓ LOGA** (TODO: implementar)
- ➡️ Vai para `waha.events.label.parsed`

### 5. **Grupos** (`"group.v2.join"`, `"group.v2.leave"`, etc.)
```go
case "group.v2.join", "group.v2.leave", "group.v2.update", "group.v2.participants":
    return p.processGroupEvent(ctx, rawEvent, wahaEvent)
```

**O que faz:**
- ⏭️ **SÓ LOGA** (TODO: implementar)
- ➡️ Vai para `waha.events.group.parsed`

### 6. **Eventos Desconhecidos** (qualquer outro)
```go
default:
    return p.processUnknownEvent(ctx, rawEvent, wahaEvent)
```

**O que faz:**
- ⚠️ **LOGA WARNING**
- ➡️ Vai para `waha.events.unknown.parsed`

## 🎯 Decisões Importantes

### ✅ **Eventos que SÃO Processados Completamente**
- `"message"` e `"message.any"` → **Processamento completo**
  - Salva no banco
  - Cria/atualiza contato
  - Inicia sessão de atendimento
  - Extrai dados de FB Ads
  - Processa mídia

### ⏭️ **Eventos que SÃO Apenas Enfileirados**
- `"message.ack"` → Para `waha.events.message.parsed`
- `"call.*"` → Para `waha.events.call.parsed`
- `"label.*"` → Para `waha.events.label.parsed`
- `"group.*"` → Para `waha.events.group.parsed`
- Desconhecidos → Para `waha.events.unknown.parsed`

### ❌ **Eventos que VÃO para DLQ**
- Parse errors → Para `waha.events.parse_errors`
- JSON inválido → Para `waha.events.parse_errors`
- Panics → Para `waha.events.parse_errors`

## 🔍 Análise dos Seus Eventos Reais

### 1. **Mensagem de Texto** (`message_text.json`)
```json
{
  "event": "message",
  "payload": {
    "_data": {
      "Info": { "Type": "text" }
    }
  }
}
```
**Rota:** `message` → **PROCESSA COMPLETO** ✅

### 2. **PTT Audio** (`message_recorded_audio.json`)
```json
{
  "event": "message", 
  "payload": {
    "_data": {
      "Info": { "MediaType": "ptt" }
    }
  }
}
```
**Rota:** `message` → **PROCESSA COMPLETO** ✅
**Antes:** ❌ Erro 500 "unsupported media type: ptt"
**Agora:** ✅ Vai para fila, não quebra webhook

### 3. **FB Ads** (`fb_ads_message.json`)
```json
{
  "event": "message",
  "payload": {
    "_data": {
      "Message": {
        "extendedTextMessage": {
          "contextInfo": {
            "conversionSource": "FB_Ads",
            "ctwaClid": "..."
          }
        }
      }
    }
  }
}
```
**Rota:** `message` → **PROCESSA COMPLETO** ✅
**Crítico:** Extrai dados de conversão para métricas de ROI

### 4. **Imagem** (`message_image.json`)
```json
{
  "event": "message",
  "payload": {
    "_data": {
      "Info": { "MediaType": "image" }
    }
  }
}
```
**Rota:** `message` → **PROCESSA COMPLETO** ✅

## 🚨 Pontos de Atenção

### 1. **Só Mensagens São Processadas**
- Apenas eventos `"message"` passam pelo `wahaMessageService`
- Outros tipos só são enfileirados para processamento futuro

### 2. **FB Ads é Crítico**
- Contém `conversionSource: "FB_Ads"`
- Contém `ctwaClid` para tracking
- **DEVE** ser processado completamente

### 3. **PTT Era o Problema**
- Antes quebrava no parsing
- Agora vai para fila raw → processa → sucesso

### 4. **Erros Não Quebram Mais**
- Parse errors → DLQ
- Webhook sempre retorna 200
- Eventos preservados para debug

## 🔧 Configuração Atual

```go
// Só estes tipos são processados completamente:
case "message", "message.any":
    return p.processMessageEvent() // ← PROCESSAMENTO COMPLETO

// Todos os outros só são enfileirados:
case "message.ack":
case "call.*":
case "label.*": 
case "group.*":
default:
    return p.publishToProcessedQueue() // ← SÓ ENFILEIRA
```

## 🎯 Próximos Passos

1. **Implementar processadores específicos** para:
   - ACKs (atualizar status de entrega)
   - Chamadas (registrar tentativas de contato)
   - Labels (organização de conversas)

2. **Monitorar filas** para ver volume de cada tipo

3. **Analisar eventos unknown** para identificar novos tipos
