# WAHA Raw Events Architecture

## 🎯 Objetivo

Implementar uma arquitetura resiliente de eventos WAHA que **nunca perde eventos** e resolve os problemas de parsing que causavam erros 500 no webhook.

## 🏗️ Arquitetura

### Fluxo Anterior (Problemático)
```
WAHA Webhook → Parse → ❌ Erro → 500 Response
```

### Novo Fluxo (Resiliente)
```
WAHA Webhook → 📥 waha.events.raw → Consumer → Parse → 📤 Filas Específicas
```

## 📋 Componentes

### 1. Fila de Entrada (Raw Events)
- **Nome**: `waha.events.raw`
- **Função**: Recebe todos os eventos WAHA sem parsing
- **Payload**: JSON bruto do webhook
- **Garantia**: Nunca falha - sempre aceita

### 2. Estruturas de Dados

#### WAHARawEvent
```go
type WAHARawEvent struct {
    ID        string            `json:"id"`        // UUID único
    Timestamp time.Time         `json:"timestamp"` // Quando foi recebido
    Session   string            `json:"session"`   // Session ID do WAHA
    Body      []byte            `json:"body"`      // JSON bruto
    Headers   map[string]string `json:"headers"`   // Headers HTTP
    Source    string            `json:"source"`    // "webhook", "retry"
    Metadata  map[string]string `json:"metadata"`  // Debug info
}
```

#### WAHAProcessedEvent
```go
type WAHAProcessedEvent struct {
    RawEventID string                 `json:"raw_event_id"`
    EventType  string                 `json:"event_type"`
    Session    string                 `json:"session"`
    ParsedAt   time.Time              `json:"parsed_at"`
    Payload    map[string]interface{} `json:"payload"`
    Metadata   map[string]interface{} `json:"metadata"`
}
```

### 3. Filas de Saída (Eventos Processados)
- `waha.events.message.parsed` - Mensagens válidas
- `waha.events.call.parsed` - Chamadas válidas
- `waha.events.presence.parsed` - Presença válida
- `waha.events.group.parsed` - Eventos de grupo válidos
- `waha.events.label.parsed` - Eventos de label válidos
- `waha.events.unknown.parsed` - Eventos desconhecidos mas válidos
- `waha.events.parse_errors` - Erros de parsing (DLQ)

## 🔧 Implementação

### 1. Webhook Handler (Novo)
```go
func (h *WAHAWebhookHandler) ReceiveWebhook(c *gin.Context) {
    // Lê body bruto
    body, _ := io.ReadAll(c.Request.Body)
    
    // Cria evento raw
    rawEvent := waha.NewWAHARawEvent(
        c.Query("session"),
        body,
        extractHeaders(c),
    )
    
    // Enfileira SEMPRE (nunca falha)
    h.rawEventBus.PublishRawEvent(ctx, rawEvent)
    
    // Resposta imediata
    c.JSON(200, gin.H{
        "status": "queued",
        "event_id": rawEvent.ID,
    })
}
```

### 2. Raw Event Processor
```go
func (p *WAHARawEventProcessor) ProcessMessage(ctx context.Context, delivery amqp.Delivery) error {
    var rawEvent waha.WAHARawEvent
    json.Unmarshal(delivery.Body, &rawEvent)
    
    // Parse com recovery
    wahaEvent, err := waha.ParseWebhookEvent(rawEvent.Body)
    if err != nil {
        return p.handleParseError(ctx, rawEvent, err)
    }
    
    // Roteamento baseado no tipo
    return p.routeEvent(ctx, rawEvent, wahaEvent)
}
```

### 3. Roteamento Inteligente
```go
func (p *WAHARawEventProcessor) routeEvent(ctx context.Context, rawEvent waha.WAHARawEvent, wahaEvent *waha.WAHAWebhookEvent) error {
    switch wahaEvent.Event {
    case "message", "message.any":
        return p.processMessageEvent(ctx, rawEvent, wahaEvent)
    case "call.received", "call.accepted", "call.rejected":
        return p.processCallEvent(ctx, rawEvent, wahaEvent)
    // ... outros tipos
    default:
        return p.processUnknownEvent(ctx, rawEvent, wahaEvent)
    }
}
```

## 🛡️ Tratamento de Erros

### 1. Parse Errors
- Eventos que falham no parsing vão para `waha.events.parse_errors`
- Incluem contexto completo para debug
- Não quebram o fluxo principal

### 2. Panic Recovery
```go
defer func() {
    if r := recover(); r != nil {
        parseError := waha.WAHAParseError{
            RawEventID: rawEvent.ID,
            Error:      fmt.Sprintf("panic: %v", r),
            ErrorType:  "panic",
            OccurredAt: time.Now(),
            RawBody:    rawEvent.Body,
        }
        p.eventBus.PublishParseError(ctx, parseError)
    }
}()
```

### 3. Dead Letter Queues
- Todas as filas têm DLQ automática
- Retry automático (3 tentativas)
- Eventos problemáticos não bloqueiam o fluxo

## 🚀 Benefícios

### 1. Resiliência Total
- ✅ Webhook **nunca** retorna 500
- ✅ Eventos **nunca** são perdidos
- ✅ Parse errors não quebram o sistema

### 2. Observabilidade
- 📊 Todos os eventos têm ID único
- 📊 Rastreamento completo do fluxo
- 📊 Métricas detalhadas por tipo

### 3. Escalabilidade
- ⚡ Processamento assíncrono
- ⚡ Múltiplos consumers
- ⚡ Backpressure automático

### 4. Debugging
- 🔍 Eventos raw preservados
- 🔍 Stack traces completos
- 🔍 Contexto de erro detalhado

## 📊 Resolução dos Problemas Originais

### 1. "unsupported media type: ptt"
- **Antes**: Webhook falhava com 500
- **Agora**: Evento vai para `waha.events.parse_errors`, webhook retorna 200

### 2. "json: cannot unmarshal object into Go struct field WAHAPayload.replyTo"
- **Antes**: Webhook falhava com 500
- **Agora**: Evento vai para `waha.events.parse_errors`, webhook retorna 200

### 3. Perda de Eventos
- **Antes**: Eventos perdidos em caso de erro
- **Agora**: Todos os eventos preservados na fila raw

## 🔄 Migração

### Fase 1: Implementação Paralela
- ✅ Nova arquitetura implementada
- ✅ Webhook usa nova arquitetura
- ✅ Sistema legado mantido para compatibilidade

### Fase 2: Validação (Próximos Passos)
- [ ] Monitorar métricas das filas
- [ ] Validar processamento de eventos
- [ ] Comparar com sistema legado

### Fase 3: Migração Completa (Futuro)
- [ ] Migrar todos os consumers
- [ ] Remover sistema legado
- [ ] Otimizar performance

## 🛠️ Comandos Úteis

### Verificar Filas
```bash
# Via API do sistema
curl http://localhost:8080/api/v1/admin/queues

# Via RabbitMQ Management
rabbitmqctl list_queues name messages consumers
```

### Monitorar Logs
```bash
# Eventos raw recebidos
grep "WAHA webhook received" logs/app.log

# Eventos processados
grep "Raw event processed" logs/app.log

# Erros de parsing
grep "Parse error in raw event" logs/app.log
```

## 📈 Métricas Importantes

1. **Taxa de Eventos Raw**: Quantos eventos chegam por segundo
2. **Taxa de Parse Success**: % de eventos parseados com sucesso
3. **Taxa de Parse Errors**: % de eventos que falham no parsing
4. **Latência de Processamento**: Tempo entre recebimento e processamento
5. **Tamanho das Filas**: Backlog de eventos pendentes

## 🎯 Próximos Passos

1. **Monitoramento**: Implementar dashboards para as métricas
2. **Alertas**: Configurar alertas para filas cheias ou erros
3. **Otimização**: Ajustar número de consumers baseado na carga
4. **Análise**: Analisar padrões de erros para melhorar parsing
5. **Documentação**: Criar runbooks para operação
