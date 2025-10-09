# Integração de Processamento de IA com Temporal

## 📋 Visão Geral

Este documento explica como integrar o processamento de mídia com IA usando eventos de domínio que disparam workflows do Temporal.

## 🎯 Fluxo de Processamento

```
1. Mensagem recebida (webhook WAHA)
   ↓
2. Message criada no domínio
   ↓
3. Verificar configuração do canal
   ↓
4. Disparar evento de processamento de IA
   ↓
5. Evento publicado no event bus
   ↓
6. Worker do Temporal consome evento
   ↓
7. Workflow de processamento executado
   ↓
8. Resultado salvo (transcrição, OCR, etc)
```

## 🔧 Implementação no Message Service

### Exemplo de Uso

```go
package message

import (
    "context"
    "github.com/caloi/ventros-crm/internal/domain/channel"
    "github.com/caloi/ventros-crm/internal/domain/message"
)

func (s *MessageService) CreateMessage(ctx context.Context, req CreateMessageRequest) error {
    // 1. Criar a mensagem
    msg, err := message.NewMessage(
        req.ContactID,
        req.ProjectID,
        req.CustomerID,
        message.ContentType(req.ContentType),
        req.FromMe,
    )
    if err != nil {
        return err
    }

    // Configurar campos adicionais
    msg.SetChannelID(req.ChannelID)
    msg.SetSessionID(req.SessionID)
    msg.SetText(req.Text)
    msg.SetMediaURL(req.MediaURL)
    msg.SetMediaMimetype(req.MediaMimetype)

    // 2. Buscar configuração do canal
    ch, err := s.channelRepo.GetByID(req.ChannelID)
    if err != nil {
        return err
    }

    // 3. Se o canal tem IA habilitada, disparar processamento
    if ch.ShouldProcessAI() {
        aiConfig := ch.GetAIProcessingConfig()
        msg.RequestAIProcessing(message.AIProcessingConfig{
            ProcessImage: aiConfig.ProcessImage,
            ProcessVideo: aiConfig.ProcessVideo,
            ProcessAudio: aiConfig.ProcessAudio,
            ProcessVoice: aiConfig.ProcessVoice,
        })
    }

    // 4. Salvar mensagem
    if err := s.repo.Save(msg); err != nil {
        return err
    }

    // 5. Publicar eventos de domínio (incluindo eventos de IA)
    for _, event := range msg.DomainEvents() {
        if err := s.eventBus.Publish(ctx, event); err != nil {
            s.logger.Error("Failed to publish event", zap.Error(err))
        }
    }

    msg.ClearEvents()
    return nil
}
```

## 📨 Eventos de IA Disponíveis

### 1. `message.ai.process_image_requested`

**Quando disparado:**
- ContentType = "image"
- Canal tem `ai_process_image = true`
- Mensagem tem `media_url` preenchida

**Payload:**
```go
type AIProcessImageRequestedEvent struct {
    MessageID   uuid.UUID
    ChannelID   uuid.UUID
    ContactID   uuid.UUID
    SessionID   uuid.UUID
    ImageURL    string
    MimeType    string
    RequestedAt time.Time
}
```

**Workflow Temporal:**
- OCR (extração de texto)
- Reconhecimento de objetos
- Análise de conteúdo visual

---

### 2. `message.ai.process_video_requested`

**Quando disparado:**
- ContentType = "video"
- Canal tem `ai_process_video = true`
- Mensagem tem `media_url` preenchida

**Payload:**
```go
type AIProcessVideoRequestedEvent struct {
    MessageID   uuid.UUID
    ChannelID   uuid.UUID
    ContactID   uuid.UUID
    SessionID   uuid.UUID
    VideoURL    string
    MimeType    string
    Duration    int // segundos
    RequestedAt time.Time
}
```

**Workflow Temporal:**
- Transcrição de áudio do vídeo
- Análise de conteúdo visual
- Extração de frames-chave

---

### 3. `message.ai.process_audio_requested`

**Quando disparado:**
- ContentType = "audio"
- Canal tem `ai_process_audio = true`
- Mensagem tem `media_url` preenchida

**Payload:**
```go
type AIProcessAudioRequestedEvent struct {
    MessageID   uuid.UUID
    ChannelID   uuid.UUID
    ContactID   uuid.UUID
    SessionID   uuid.UUID
    AudioURL    string
    MimeType    string
    Duration    int // segundos
    RequestedAt time.Time
}
```

**Workflow Temporal:**
- Transcrição de áudio (Whisper API)
- Análise de sentimento
- Detecção de idioma

---

### 4. `message.ai.process_voice_requested` ⭐

**Quando disparado:**
- ContentType = "voice" (PTT do WhatsApp)
- Canal tem `ai_process_voice = true`
- Mensagem tem `media_url` preenchida

**Payload:**
```go
type AIProcessVoiceRequestedEvent struct {
    MessageID   uuid.UUID
    ChannelID   uuid.UUID
    ContactID   uuid.UUID
    SessionID   uuid.UUID
    VoiceURL    string
    MimeType    string
    Duration    int // segundos
    RequestedAt time.Time
}
```

**Workflow Temporal:**
- Transcrição automática de PTT
- **Resolve o erro "unsupported media type: ptt"**
- Análise de sentimento
- Salvar transcrição no campo `text` da mensagem

---

## 🔄 Exemplo de Workflow Temporal

### Worker que consome eventos

```go
package workflows

import (
    "context"
    "go.temporal.io/sdk/workflow"
)

// ProcessVoiceWorkflow processa mensagens de voz/PTT
func ProcessVoiceWorkflow(ctx workflow.Context, event AIProcessVoiceRequestedEvent) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Processing voice message", "message_id", event.MessageID)

    // 1. Download do áudio
    var audioData []byte
    err := workflow.ExecuteActivity(ctx, DownloadMediaActivity, event.VoiceURL).Get(ctx, &audioData)
    if err != nil {
        return err
    }

    // 2. Transcrição com Whisper
    var transcription string
    err = workflow.ExecuteActivity(ctx, TranscribeAudioActivity, audioData).Get(ctx, &transcription)
    if err != nil {
        return err
    }

    // 3. Análise de sentimento
    var sentiment string
    err = workflow.ExecuteActivity(ctx, AnalyzeSentimentActivity, transcription).Get(ctx, &sentiment)
    if err != nil {
        return err
    }

    // 4. Salvar resultado
    result := ProcessingResult{
        MessageID:     event.MessageID,
        Transcription: transcription,
        Sentiment:     sentiment,
    }
    
    err = workflow.ExecuteActivity(ctx, SaveProcessingResultActivity, result).Get(ctx, nil)
    return err
}
```

### Registrar worker

```go
package main

import (
    "go.temporal.io/sdk/client"
    "go.temporal.io/sdk/worker"
)

func main() {
    c, _ := client.NewClient(client.Options{})
    defer c.Close()

    w := worker.New(c, "ai-processing-queue", worker.Options{})

    // Registrar workflows
    w.RegisterWorkflow(ProcessVoiceWorkflow)
    w.RegisterWorkflow(ProcessImageWorkflow)
    w.RegisterWorkflow(ProcessVideoWorkflow)
    w.RegisterWorkflow(ProcessAudioWorkflow)

    // Registrar activities
    w.RegisterActivity(DownloadMediaActivity)
    w.RegisterActivity(TranscribeAudioActivity)
    w.RegisterActivity(AnalyzeSentimentActivity)
    w.RegisterActivity(SaveProcessingResultActivity)

    w.Run(worker.InterruptCh())
}
```

## 🎛️ Configuração no Frontend

O usuário configura no modal de criação de canal:

```
☑ Canal Inteligente
  └─ Habilita processamento inteligente

☐ Agentes IA
  └─ Respostas automáticas

┌─────────────────────────────────────────┐
│ Processamento de Mídia com IA          │
│                                         │
│ ☑ Processar Imagens                    │
│ ☐ Processar Vídeos                     │
│ ☐ Processar Áudios                     │
│ ☑ Processar Mensagens de Voz/PTT       │
│ ─────────────────────────────────────   │
│ ☑ Resumir Sessões Automaticamente      │
└─────────────────────────────────────────┘
```

## 📊 Banco de Dados

### Tabela `channels`

```sql
ai_enabled BOOLEAN DEFAULT FALSE
ai_agents_enabled BOOLEAN DEFAULT FALSE
ai_process_image BOOLEAN DEFAULT FALSE
ai_process_video BOOLEAN DEFAULT FALSE
ai_process_audio BOOLEAN DEFAULT FALSE
ai_process_voice BOOLEAN DEFAULT FALSE
ai_summarize_sessions BOOLEAN DEFAULT FALSE
```

### Tabela `messages` (campos para resultado)

```sql
-- Adicionar colunas para armazenar resultados do processamento
ai_transcription TEXT -- Transcrição de áudio/voz
ai_ocr_text TEXT -- Texto extraído de imagens
ai_sentiment VARCHAR(50) -- Sentimento detectado
ai_processed_at TIMESTAMP -- Quando foi processado
ai_processing_status VARCHAR(50) -- pending, processing, completed, failed
```

## 🚀 Próximos Passos

1. **Implementar workers do Temporal** para cada tipo de processamento
2. **Integrar com APIs de IA:**
   - OpenAI Whisper (transcrição)
   - OpenAI Vision (análise de imagens)
   - Google Cloud Vision (OCR)
3. **Criar activities** para download de mídia, processamento e salvamento
4. **Implementar retry policies** para falhas temporárias
5. **Adicionar métricas** de processamento (tempo, custo, taxa de sucesso)

## ⚠️ Considerações Importantes

- **Custo:** Processamento de IA tem custo por requisição
- **Performance:** Usar filas assíncronas (Temporal) para não bloquear webhooks
- **Privacidade:** Garantir que dados sensíveis sejam tratados adequadamente
- **Rate Limits:** APIs de IA têm limites de requisições por minuto
- **Fallback:** Ter estratégia para quando IA falhar

## 📝 Exemplo Completo de Uso

```go
// No webhook handler do WAHA
func (h *WAHAWebhookHandler) ProcessMessage(ctx context.Context, payload WAHAPayload) error {
    // 1. Criar mensagem
    msg := message.NewMessage(...)
    msg.SetMediaURL(payload.MediaURL)
    msg.SetContentType(message.ContentTypeVoice) // PTT

    // 2. Buscar canal
    channel, _ := h.channelRepo.GetByExternalID(payload.SessionID)

    // 3. Se canal tem processamento de voz habilitado
    if channel.AIProcessVoice {
        // Evento será disparado automaticamente
        msg.RequestAIProcessing(channel.GetAIProcessingConfig())
    }

    // 4. Salvar e publicar eventos
    h.messageRepo.Save(msg)
    h.eventBus.PublishAll(msg.DomainEvents())
    
    return nil
}
```

---

**Status:** ✅ Eventos de domínio implementados  
**Próximo:** Implementar workers do Temporal
