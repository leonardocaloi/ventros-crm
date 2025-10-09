# Message Debouncer - Guia de Uso

## Visão Geral

O Message Debouncer implementa a lógica do n8n para agrupar mensagens rápidas e fragmentadas, com **camadas desacopladas** para máxima flexibilidade.

## Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    Message Flow                              │
└─────────────────────────────────────────────────────────────┘

WAHA Event → DebouncerIntegration → MessageDebouncerV2 → Redis
                                            │
                                            ▼
                                    [Push + Check Loop]
                                            │
                                            ▼
                                    Pull + Switch Logic
                                            │
                         ┌──────────────────┼──────────────────┐
                         ▼                  ▼                  ▼
                    Nothing            Proceed             Wait 15s
                   (duplicada)      (processar)          (retry loop)
                                            │
                                            ▼
                              MessageBatchProcessor (OPCIONAL)
                                            │
                      ┌─────────────────────┼─────────────────────┐
                      ▼                     ▼                     ▼
               Concatenator            Enricher              Sender
            (texto ou JSON)      (adiciona contexto)    (OpenAI/etc)
```

## Modos de Uso

### Modo 1: Apenas Debouncing (Você Controla Tudo)

Use quando quiser **apenas agrupar mensagens** sem processamento automático.

```go
package main

import (
    "context"
    "github.com/redis/go-redis/v9"
    "github.com/caloi/ventros-crm/infrastructure/messaging"
)

func main() {
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Debouncer SEM processor (nil)
    debouncer := messaging.NewMessageDebouncerV2(redisClient, 15*time.Second, nil)

    ctx := context.Background()
    sessionKey := "contact123:waha:channel456"

    // 1. Push mensagem
    msg := messaging.BufferedMessage{
        MessageID:   "msg_001",
        Text:        "Olá",
        Type:        "text",
        Timestamp:   time.Now().UnixMilli(),
        FromContact: true,
        ContactID:   "contact123",
    }

    // 2. Push manual (não inicia loop automático)
    debouncer.Push(ctx, sessionKey, msg)

    // 3. Quando quiser processar, pull manual
    messages, _ := debouncer.Pull(ctx, sessionKey)

    // 4. Processa como quiser
    for _, m := range messages {
        // Seu código aqui
    }

    // 5. Limpa buffer
    debouncer.ClearBuffer(ctx, sessionKey)
}
```

**Quando usar**:
- Quer controle total
- Processamento customizado complexo
- Integração com sistema legado

---

### Modo 2: Debouncing + Processamento Simples

Use quando quiser **agrupar E processar automaticamente**, mas sem IA.

```go
package main

func main() {
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Cria processor básico
    processor := messaging.NewMessageBatchProcessor(
        messaging.SimpleConcatenator{},  // Concatena texto com \n
        messaging.NoopValidator{},       // Não valida
        nil,                             // Sem enrichment
        nil,                             // Sem sender (processa internamente)
    )

    // Integração completa
    integration := messaging.NewDebouncerIntegration(redisClient, processor)

    ctx := context.Background()

    // Processa mensagem WAHA (automático: push + loop + process)
    err := integration.ProcessWAHAMessage(ctx, wahaEvent)
    if err != nil {
        log.Fatal(err)
    }

    // Ou versão genérica
    err = integration.ProcessMessage(
        ctx,
        "contact123",        // contactID
        "waha",              // channel type
        "channel456",        // channel ID
        "msg_001",           // message ID
        "Olá, tudo bem?",    // text
        "text",              // type
        time.Now().UnixMilli(), // timestamp
        true,                // fromContact
        nil,                 // metadata
    )
}
```

**Quando usar**:
- Quer automação básica
- Apenas concatenar mensagens
- Enviar para webhook/banco/fila

---

### Modo 3: Debouncing + AI (OpenAI/Anthropic)

Use quando quiser **agrupar E enviar para IA automaticamente**.

```go
package main

// 1. Implementa seu Sender customizado
type OpenAISender struct {
    client *openai.Client
}

func (s *OpenAISender) Send(ctx context.Context, sessionKey string, content string, metadata interface{}) error {
    // Envia para OpenAI
    resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    "system",
                Content: "Você é um assistente...",
            },
            {
                Role:    "user",
                Content: content, // Mensagens concatenadas
            },
        },
    })
    if err != nil {
        return err
    }

    // Processa resposta
    aiResponse := resp.Choices[0].Message.Content

    // Envia resposta ao contato (via WAHA/WhatsApp/etc)
    // ... seu código aqui ...

    return nil
}

func main() {
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Cria AI sender
    aiSender := &OpenAISender{
        client: openai.NewClient(apiKey),
    }

    // Cria processor com AI
    processor := messaging.NewMessageBatchProcessor(
        messaging.MediaAwareConcatenator{}, // Detecta mídia
        messaging.MinMessageValidator{MinCount: 1},
        nil,      // Enricher customizado (opcional)
        aiSender, // Envia para IA
    )

    integration := messaging.NewDebouncerIntegration(redisClient, processor)

    // Usa no consumer
    ctx := context.Background()
    integration.ProcessWAHAMessage(ctx, wahaEvent)
}
```

**Quando usar**:
- Integração com OpenAI/Anthropic/Claude
- Respostas automáticas inteligentes
- Chatbots com IA

---

## Integração com WAHAMessageConsumer

### Antes (sem debouncer)

```go
func (c *WAHAMessageConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
    var wahaEvent waha.WAHAMessageEvent
    json.Unmarshal(delivery.Body, &wahaEvent)

    // Processava direto
    return c.wahaMessageService.ProcessWAHAMessage(ctx, wahaEvent)
}
```

### Depois (com debouncer)

```go
type WAHAMessageConsumer struct {
    wahaMessageService *message.WAHAMessageService
    idempotencyChecker IdempotencyChecker
    debouncerIntegration *messaging.DebouncerIntegration // NOVO
    consumerName       string
}

func (c *WAHAMessageConsumer) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
    var wahaEvent waha.WAHAMessageEvent
    json.Unmarshal(delivery.Body, &wahaEvent)

    // Usa debouncer ao invés de processar direto
    if c.debouncerIntegration != nil {
        return c.debouncerIntegration.ProcessWAHAMessage(ctx, wahaEvent)
    }

    // Fallback: processa direto (compatibilidade)
    return c.wahaMessageService.ProcessWAHAMessage(ctx, wahaEvent)
}
```

---

## Configuração no main.go

```go
package main

func setupDebouncer(redisClient *redis.Client, config *config.Config) *messaging.DebouncerIntegration {
    // Lê configuração
    enableAI := config.GetBool("AI_ENABLED")

    if !enableAI {
        // Modo simples (sem IA)
        processor := messaging.NewMessageBatchProcessor(
            messaging.SimpleConcatenator{},
            messaging.NoopValidator{},
            nil,
            nil,
        )
        return messaging.NewDebouncerIntegration(redisClient, processor)
    }

    // Modo AI
    aiProvider := config.GetString("AI_PROVIDER") // "openai" ou "anthropic"

    var sender messaging.MessageSender
    switch aiProvider {
    case "openai":
        sender = NewOpenAISender(config.GetString("OPENAI_API_KEY"))
    case "anthropic":
        sender = NewAnthropicSender(config.GetString("ANTHROPIC_API_KEY"))
    default:
        sender = nil
    }

    processor := messaging.NewMessageBatchProcessor(
        messaging.MediaAwareConcatenator{},
        messaging.MinMessageValidator{MinCount: 1},
        nil,
        sender,
    )

    return messaging.NewDebouncerIntegration(redisClient, processor)
}

func main() {
    // ... setup Redis ...

    debouncerIntegration := setupDebouncer(redisClient, cfg)

    // Injeta no consumer
    wahaConsumer := messaging.NewWAHAMessageConsumer(
        wahaMessageService,
        idempotencyChecker,
        debouncerIntegration, // NOVO parâmetro
    )

    // ... start consumer ...
}
```

---

## Variáveis de Ambiente

```bash
# Debouncer
DEBOUNCER_ENABLED=true
DEBOUNCER_WAIT_DURATION=15s  # Tempo de espera (padrão n8n)
DEBOUNCER_MAX_RETRIES=10     # Max tentativas antes de forçar

# AI (opcional)
AI_ENABLED=true
AI_PROVIDER=openai           # ou "anthropic"
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

---

## Fluxo de Decisão (Switch Logic)

```go
// Replica exatamente o Switch do n8n:

1. Pull mensagens do Redis
2. Verifica primeira mensagem da lista
   - Se ≠ mensagem atual → NOTHING (duplicada, ignora)
3. Verifica última mensagem da lista
   - Se timestamp > 15s atrás → PROCEED (processa)
4. Senão → WAIT 15s e volta ao passo 1
5. Após max_retries (10) → força PROCEED
```

---

## Testes

```go
func TestDebouncer(t *testing.T) {
    redisClient := setupTestRedis(t)
    ctx := context.Background()

    debouncer := messaging.NewMessageDebouncerV2(redisClient, 100*time.Millisecond, nil)

    sessionKey := "test:session:1"

    // Push mensagens rápido
    for i := 0; i < 5; i++ {
        msg := messaging.BufferedMessage{
            MessageID: fmt.Sprintf("msg_%d", i),
            Text:      fmt.Sprintf("Texto %d", i),
            Timestamp: time.Now().UnixMilli(),
        }
        debouncer.Push(ctx, sessionKey, msg)
    }

    // Aguarda timeout
    time.Sleep(150 * time.Millisecond)

    // Pull e verifica
    messages, err := debouncer.Pull(ctx, sessionKey)
    require.NoError(t, err)
    assert.Len(t, messages, 5)
}
```

---

## Métricas e Observabilidade

```go
// Adicione métricas (Prometheus):

debouncer_messages_pushed_total{session_key}
debouncer_messages_processed_total{session_key}
debouncer_wait_duration_seconds{session_key}
debouncer_batch_size{session_key}
debouncer_retries_total{session_key}
```

---

## FAQ

**Q: Preciso usar o processor?**
A: Não! Você pode criar o debouncer com `nil` e processar manualmente.

**Q: Posso mudar o tempo de espera?**
A: Sim, passe `time.Duration` customizado no construtor.

**Q: Como integrar com meu AI provider customizado?**
A: Implemente a interface `MessageSender` e passe no processor.

**Q: E se eu não quiser usar Redis?**
A: Você teria que implementar outra storage layer, mas Redis é padrão da indústria para isso.

**Q: Funciona com mensagens de áudio/vídeo?**
A: Sim! Use `MediaAwareConcatenator` que detecta automaticamente.
