# Análise do Fluxo de Mensagens WAHA → CRM

## 📋 Resumo Executivo

Todas as mensagens são armazenadas na tabela `messages` independente do tipo (texto, imagem, áudio, documento, localização, contato, etc.). O sistema segue DDD com camadas bem separadas.

## 🔄 Fluxo Completo de Processamento

```
WAHA Webhook/RabbitMQ
    ↓
WAHAMessageEvent (Infrastructure Layer)
    ↓
MessageAdapter.ToContentType() → Determina tipo
    ↓
WAHAMessageService.ProcessWAHAMessage()
    ↓
ProcessInboundMessageCommand (Application Layer)
    ↓
ProcessInboundMessageUseCase.Execute()
    ↓
1. FindOrCreate Contact
2. FindOrCreate Session
3. Create Message (Domain)
4. Record in Session
5. Save to DB (messages table)
    ↓
MessageEntity (Persistence Layer)
```

## 📊 Mapeamento de Campos por Tipo de Mensagem

### Tabela `messages` - Campos Principais

| Campo DB | Tipo | Descrição | Preenchido em |
|----------|------|-----------|---------------|
| `id` | uuid | ID único da mensagem | Todos |
| `timestamp` | timestamp | Data/hora da mensagem | Todos |
| `user_id` | uuid | Dono do workspace | Todos |
| `project_id` | uuid | Projeto | Todos |
| `channel_type_id` | int | Tipo do canal (WhatsApp=1) | Todos |
| `from_me` | bool | Enviada por mim? | Todos |
| `channel_id` | uuid | Canal específico | Todos |
| `contact_id` | uuid | Contato | Todos |
| `session_id` | uuid | Sessão ativa | Todos |
| `content_type` | string | Tipo: text/image/audio/voice/video/document/location/contact/sticker | Todos |
| `text` | text | Conteúdo textual | text, captions |
| `media_url` | string | URL da mídia | image/audio/video/document/sticker |
| `media_mimetype` | string | Tipo MIME | image/audio/video/document/sticker |
| `channel_message_id` | string | ID externo | Todos |
| `reply_to_id` | uuid | Resposta a outra msg | Quando aplicável |
| `status` | string | sent/delivered/read/failed | Todos |
| `metadata` | jsonb | Dados extras | Todos |

### Tipos de Mensagem Suportados

#### 1. **TEXT** (`content_type: "text"`)
**Exemplos:** `message_text.json`, `message_image_text.json`

**Campos preenchidos:**
- `text`: Conteúdo da mensagem
- `content_type`: "text"
- Campos comuns (timestamp, contact_id, session_id, etc.)

**Extração:**
```
payload.body → text
payload._data.Message.conversation → text
payload._data.Message.extendedTextMessage.text → text
```

**Tracking de Ads:**
- Se `extendedTextMessage.contextInfo` existe → pode ter dados de conversão
- `metadata` armazena: conversion_source, ctwa_clid, ad_source_url, etc.

---

#### 2. **IMAGE** (`content_type: "image"`)
**Exemplos:** `message_image.json`, `message_image_text.json`

**Campos preenchidos:**
- `media_url`: URL da imagem (S3/GCS)
- `media_mimetype`: "image/jpeg", "image/png", etc.
- `text`: Caption (se houver)
- `content_type`: "image"

**Extração:**
```
payload.media.url → media_url
payload.media.mimetype → media_mimetype
payload._data.Message.imageMessage.caption → text
```

---

#### 3. **AUDIO** (`content_type: "audio"`)
**Exemplo:** `message_audio.json`

**Campos preenchidos:**
- `media_url`: URL do áudio
- `media_mimetype`: "audio/ogg; codecs=opus", "audio/mp4", etc.
- `content_type`: "audio"

**Extração:**
```
payload.media.url → media_url
payload.media.mimetype → media_mimetype
payload._data.Info.MediaType: "audio"
```

---

#### 4. **VOICE/PTT** (`content_type: "voice"`)
**Exemplo:** `message_recorded_audio.json`

**Campos preenchidos:**
- `media_url`: URL do áudio gravado
- `media_mimetype`: "audio/ogg; codecs=opus"
- `content_type`: "voice"

**Identificação:**
```
payload._data.Info.MediaType: "ptt"
payload._data.Message.audioMessage.PTT: true
```

**⚠️ PROBLEMA IDENTIFICADO:**
- O adapter atual não detecta PTT corretamente
- Método `isPTT()` retorna sempre `false`
- Causa erro: "unsupported media type: ptt"

---

#### 5. **VIDEO** (`content_type: "video"`)
**Campos preenchidos:**
- `media_url`: URL do vídeo
- `media_mimetype`: "video/mp4", etc.
- `text`: Caption (se houver)
- `content_type`: "video"

**Extração:**
```
payload._data.Info.MediaType: "video"
payload._data.Message.videoMessage
```

---

#### 6. **DOCUMENT** (`content_type: "document"`)
**Exemplo:** `message_document_pdf.json`, `message_document_image_heic.json`

**Campos preenchidos:**
- `media_url`: URL do documento
- `media_mimetype`: "application/pdf", "image/heic", etc.
- `text`: Caption (se houver)
- `content_type`: "document"
- `metadata`: Pode incluir `filename`, `pageCount`

**Extração:**
```
payload.media.url → media_url
payload.media.mimetype → media_mimetype
payload.media.filename → metadata.filename
payload._data.Message.documentMessage
```

---

#### 7. **LOCATION** (`content_type: "location"`)
**Exemplo:** `message_location.json`

**Campos preenchidos:**
- `content_type`: "location"
- `metadata`: { latitude, longitude, address }

**Extração:**
```
payload._data.Message.locationMessage.degreesLatitude → metadata.latitude
payload._data.Message.locationMessage.degreesLongitude → metadata.longitude
```

**⚠️ PROBLEMA IDENTIFICADO:**
- Adapter não extrai coordenadas de localização
- Dados ficam perdidos, não são salvos

---

#### 8. **CONTACT/VCARD** (`content_type: "contact"`)
**Exemplo:** `message_contact.json`

**Campos preenchidos:**
- `content_type`: "contact"
- `metadata`: { displayName, vcard }

**Extração:**
```
payload._data.Info.MediaType: "vcard"
payload._data.Message.contactMessage.displayName → metadata.displayName
payload._data.Message.contactMessage.vcard → metadata.vcard
```

**⚠️ PROBLEMA IDENTIFICADO:**
- Adapter não suporta tipo "vcard"/"contact"
- Causa erro: "unsupported media type: vcard"

---

#### 9. **STICKER** (`content_type: "sticker"`)
**Campos preenchidos:**
- `media_url`: URL do sticker
- `media_mimetype`: "image/webp"
- `content_type`: "sticker"

**Extração:**
```
payload._data.Info.MediaType: "sticker"
payload._data.Message.stickerMessage
```

---

## 🐛 Problemas Identificados

### 1. **PTT (Push-to-Talk) não é detectado**
**Arquivo:** `message_adapter.go:193-198`

```go
func (a *MessageAdapter) isPTT(event WAHAMessageEvent) bool {
    // WAHA/WhatsApp geralmente marca PTT com flags específicas
    // Adicione lógica aqui baseado na documentação do WAHA
    // Por enquanto, retorna false para áudios normais
    return false  // ❌ SEMPRE FALSE!
}
```

**Impacto:**
- Mensagens de voz (PTT) causam erro: "unsupported media type: ptt"
- Webhook retorna 500

**Solução:**
```go
func (a *MessageAdapter) isPTT(event WAHAMessageEvent) bool {
    msg := event.Payload.Data.Message
    if msg.AudioMessage != nil {
        // Verifica campo PTT na estrutura interna
        // Nota: Precisamos adicionar campo PTT na struct WAHAMediaMessage
        return true // Temporário: checar payload._data.Info.MediaType == "ptt"
    }
    return event.Payload.Data.Info.MediaType == "ptt"
}
```

---

### 2. **Location não é extraído**
**Arquivo:** `message_adapter.go`

**Problema:**
- Não há método `ExtractLocation()` no adapter
- Coordenadas ficam perdidas

**Solução necessária:**
```go
func (a *MessageAdapter) ExtractLocationData(event WAHAMessageEvent) map[string]interface{} {
    msg := event.Payload.Data.Message
    if msg.LocationMessage != nil {
        return map[string]interface{}{
            "latitude":  msg.LocationMessage.DegreesLatitude,
            "longitude": msg.LocationMessage.DegreesLongitude,
        }
    }
    return nil
}
```

---

### 3. **Contact/VCard não é suportado**
**Arquivo:** `message_adapter.go:131-190`

**Problema:**
- `ToContentType()` não trata `MediaType: "vcard"`
- Causa erro: "unsupported media type: vcard"

**Solução:**
```go
case "vcard":
    return message.ContentTypeContact, nil
```

---

### 4. **Struct WAHAMessage incompleta**
**Arquivo:** `message_adapter.go:80-89`

**Problema:**
- Falta `LocationMessage` e `ContactMessage`
- Falta campo `PTT` em `WAHAMediaMessage`

**Solução:**
```go
type WAHAMessage struct {
    Conversation     *string            `json:"conversation"`
    ImageMessage     *WAHAMediaMessage  `json:"imageMessage"`
    VideoMessage     *WAHAMediaMessage  `json:"videoMessage"`
    AudioMessage     *WAHAMediaMessage  `json:"audioMessage"`
    DocumentMessage  *WAHAMediaMessage  `json:"documentMessage"`
    StickerMessage   *WAHAMediaMessage  `json:"stickerMessage"`
    LocationMessage  *WAHALocationMessage `json:"locationMessage"`  // ✅ ADICIONAR
    ContactMessage   *WAHAContactMessage  `json:"contactMessage"`   // ✅ ADICIONAR
    ExtendedTextMsg  *WAHAExtendedText  `json:"extendedTextMessage"`
}

type WAHAMediaMessage struct {
    URL      string `json:"URL"`
    Mimetype string `json:"mimetype"`
    Caption  string `json:"caption,omitempty"`
    PTT      bool   `json:"PTT,omitempty"`  // ✅ ADICIONAR
}

type WAHALocationMessage struct {
    DegreesLatitude  float64 `json:"degreesLatitude"`
    DegreesLongitude float64 `json:"degreesLongitude"`
}

type WAHAContactMessage struct {
    DisplayName string `json:"displayName"`
    VCard       string `json:"vcard"`
}
```

---

### 5. **Metadata não é populado com dados específicos**
**Arquivo:** `process_inbound_message.go:250-289`

**Problema:**
- Metadata genérico, não inclui dados específicos do tipo
- Exemplo: filename de documento, coordenadas de localização, vcard de contato

**Solução:**
- Adicionar método `ExtractMetadata()` no adapter
- Preencher metadata específico por tipo no use case

---

## 🏗️ Arquitetura DDD - Análise

### ✅ Pontos Positivos

1. **Separação de Camadas Clara:**
   - Domain: `message.Message` (aggregate root)
   - Application: `ProcessInboundMessageUseCase`
   - Infrastructure: `MessageAdapter`, `WAHAMessageEvent`

2. **Adapter Pattern:**
   - `MessageAdapter` isola complexidade do WAHA
   - Domain não conhece estrutura externa

3. **Repository Pattern:**
   - `MessageRepository` abstrai persistência
   - Domain não conhece GORM/SQL

4. **Domain Events:**
   - `MessageCreatedEvent`, `MessageDeliveredEvent`
   - Choreography via EventBus

5. **Value Objects:**
   - `ContentType`, `Status` com validação

### ⚠️ Pontos de Melhoria

1. **Adapter Incompleto:**
   - Faltam tipos: location, contact, PTT
   - Métodos de extração incompletos

2. **Metadata Genérico:**
   - Não aproveita campos específicos
   - Perde informações valiosas (filename, coords, etc.)

3. **Error Handling:**
   - Erros de tipo não suportado causam 500
   - Deveria logar e continuar (graceful degradation)

4. **Testes:**
   - `message_adapter_test.go` existe mas precisa cobrir novos tipos

5. **Documentação:**
   - Falta doc sobre tipos suportados
   - Falta exemplos de cada tipo

---

## 📝 Recomendações

### Prioridade ALTA

1. **Corrigir detecção de PTT**
   - Implementar `isPTT()` corretamente
   - Adicionar campo `PTT` na struct

2. **Adicionar suporte a Contact/VCard**
   - Adicionar case no `ToContentType()`
   - Criar structs necessárias

3. **Adicionar suporte a Location**
   - Criar `WAHALocationMessage`
   - Implementar extração de coordenadas

### Prioridade MÉDIA

4. **Melhorar metadata**
   - Adicionar `ExtractMetadata()` genérico
   - Preencher dados específicos por tipo

5. **Graceful degradation**
   - Não falhar webhook em tipo desconhecido
   - Salvar como "system" com payload raw

### Prioridade BAIXA

6. **Testes**
   - Adicionar testes para todos os tipos
   - Usar JSONs de `events_waha/` como fixtures

7. **Documentação**
   - Documentar tipos suportados
   - Adicionar diagramas de fluxo

---

## 🔍 Tracking de Conversão (Ads)

### Campos Extraídos

Quando mensagem vem de anúncio (Facebook/Instagram Ads):

```json
{
  "conversion_source": "ad",
  "conversion_app": "facebook",
  "external_source": "instagram",
  "external_medium": "story",
  "ad_source_type": "ad",
  "ad_source_id": "123456",
  "ctwa_clid": "click-id-do-ad"
}
```

**Armazenamento:**
- `metadata.tracking` na tabela messages
- Evento `AdConversionTrackedEvent` publicado
- Timeline do contato registra origem

**Uso:**
- Relatórios de ROI de ads
- Atribuição de conversão
- Análise de canais

---

## 📊 Ordem de Processamento

```
1. Webhook recebe evento WAHA
2. Valida estrutura JSON
3. Busca canal por session (ExternalID)
4. Adapter extrai dados
5. Cria ProcessInboundMessageCommand
6. Use Case:
   a. FindOrCreate Contact (por telefone)
   b. FindOrCreate Session (por contact + canal)
   c. Create Message (domain)
   d. Assign to Channel + Session
   e. Set content (text/media)
   f. Save to DB
   g. Update session metrics
   h. Update contact last_interaction
   i. Publish domain events
   j. Track ad conversion (se aplicável)
7. Retorna 200 OK
```

---

## 🎯 Conclusão

O sistema está **bem arquitetado** seguindo DDD, mas o **adapter está incompleto**. 

**Principais gaps:**
- PTT não detectado → erro 500
- Location não extraído → dados perdidos
- Contact não suportado → erro 500
- Metadata genérico → perde contexto

**Próximos passos:**
1. Corrigir adapter (structs + métodos)
2. Adicionar testes
3. Melhorar error handling
4. Documentar tipos suportados
