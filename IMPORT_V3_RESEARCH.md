# V3 Import Optimization Research

## Problema Atual

**V1 (Implementado)**: Processa chat por chat, 20 em paralelo
- **Requests WAHA**: 1 por chat (N+1 total: 1 lista + N chats)
- **Transações DB**: 1 por chat
- **Limite**: 100 mensagens por request WAHA
- **Performance**: ~9 minutos para 1,172 chats (5,683 mensagens)

## Constraint Crítica: WAHA API Limit

```
📌 LIMITE DO WAHA: 100 mensagens por requisição
```

Isso significa que:
- Para um chat com 500 mensagens, precisamos de 5 requests (100+100+100+100+100)
- Não podemos buscar "todas as mensagens de todos os chats" em 1 request
- **PORÉM**: Podemos fazer múltiplos requests em paralelo!

## Opções de Otimização

### Opção A: Aumentar Paralelismo de Fetches (Quick Win)

**Atual**: 20 chats em paralelo (1 fetch por chat)

**Proposta**: 50-100 chats em paralelo

```go
// Aumentar de 20 → 100 chats simultâneos
maxConcurrentChats := 100  // Era 20

// Vantagem: Reduz tempo de 9min → ~2min (5x mais rápido)
// Desvantagem: Aumenta carga no WAHA e PostgreSQL
```

**Viabilidade**: ✅ FÁCIL (mudar 1 linha)
**Impacto**: 🚀 ALTO (5x mais rápido)
**Risco**: ⚠️ MÉDIO (pode sobrecarregar WAHA)

---

### Opção B: Bulk Processing com Chunking (Médio Esforço)

**Atual**: Processa 1 chat → 1 transação DB

**Proposta**: Processa N chats → 1 transação DB

```
┌────────────────────────────────────────────────────────────┐
│ Chunk 1 (100 chats):                                       │
│   1. Fetch 100 chats em paralelo (100 requests simultâneos)│
│   2. Agregar TODOS em memória (~500 mensagens)            │
│   3. Processar em 1 transação DB                           │
│   4. CHECKPOINT                                             │
├────────────────────────────────────────────────────────────┤
│ Chunk 2 (100 chats): ...                                   │
└────────────────────────────────────────────────────────────┘

Resultado: 1,172 transações → 12 transações (98% redução)
```

**Vantagens:**
- Reduz transações DB em 100x
- Checkpoints a cada chunk (não perde tudo se falhar)
- Uso controlado de memória

**Desvantagens:**
- Complexidade maior
- Precisa buffer/queue management
- Requer teste de stress

**Viabilidade**: ⚠️ MÉDIO (2-4 horas implementação)
**Impacto**: 🚀🚀 MUITO ALTO (10x mais rápido)
**Risco**: ⚠️ MÉDIO (bugs em edge cases)

---

### Opção C: Pipeline Streaming (Alta Complexidade)

**Conceito**: Producer-Consumer Pattern com Goroutines

```go
// Stage 1: Fetcher (produz mensagens)
fetchers := make(chan []WAHAMessage, 100)
for _, chat := range chats {
    go func(chat) {
        msgs := fetchMessages(chat)
        fetchers <- msgs
    }(chat)
}

// Stage 2: Aggregator (agrupa por chunk)
chunks := make(chan []ImportMessage, 10)
go func() {
    buffer := []ImportMessage{}
    for msgs := range fetchers {
        buffer = append(buffer, msgs...)
        if len(buffer) >= 500 {
            chunks <- buffer
            buffer = []ImportMessage{}
        }
    }
}()

// Stage 3: Processor (processa chunks)
for chunk := range chunks {
    importBatchUC.Execute(ctx, chunk)  // 1 transação por chunk
}
```

**Vantagens:**
- Máxima eficiência (fetch + process simultâneos)
- Não espera todos os chats antes de processar
- Stream processing (low latency)

**Desvantagens:**
- Complexidade ALTA
- Bugs difíceis de debugar
- Gerenciamento de goroutines/channels complexo

**Viabilidade**: ❌ COMPLEXO (1-2 dias implementação)
**Impacto**: 🚀🚀🚀 ALTÍSSIMO (15x mais rápido)
**Risco**: 🔴 ALTO (race conditions, deadlocks)

---

### Opção D: Hybrid Quick Win + Batch (RECOMENDADO)

**Combinação de A + B simplificado**

```go
// 1. Quick Win: Aumentar paralelismo
maxConcurrentChats := 100  // Era 20

// 2. Batch simplificado: Processar múltiplos chats em 1 transação
// (mas sem buffer complexo, apenas "flush" a cada N chats)
```

**Algoritmo:**

```
Para cada chunk de 50 chats:
  1. Fetch 50 chats em PARALELO (50 goroutines)
  2. Aguardar TODOS completarem (sync.WaitGroup)
  3. Agregar mensagens em memória
  4. Processar em 1 TRANSAÇÃO DB
  5. CHECKPOINT
```

**Código simplificado:**

```go
chunkSize := 50
for i := 0; i < len(chats); i += chunkSize {
    end := min(i+chunkSize, len(chats))
    chunk := chats[i:end]

    // Fetch paralelo
    var wg sync.WaitGroup
    allMessages := make(chan []ImportMessage, len(chunk))

    for _, chat := range chunk {
        wg.Add(1)
        go func(c Chat) {
            defer wg.Done()
            msgs := fetchMessages(c)
            allMessages <- msgs
        }(chat)
    }

    wg.Wait()
    close(allMessages)

    // Agregar
    var batch []ImportMessage
    for msgs := range allMessages {
        batch = append(batch, msgs...)
    }

    // Processar em 1 transação
    importBatchUC.Execute(ctx, ImportBatchInput{
        Messages: batch,
        // ...
    })
}
```

**Viabilidade**: ✅ FÁCIL (1-2 horas)
**Impacto**: 🚀🚀 ALTO (8-10x mais rápido)
**Risco**: 🟢 BAIXO (teste simples valida)

---

## Benchmarks Estimados

### Dataset: 1,172 chats, 5,683 mensagens

| Abordagem | Tempo | Transações DB | Requests WAHA | Complexidade |
|-----------|-------|---------------|---------------|--------------|
| V0 (Original) | ~48 min | 5,683 | 5,683 | Simples |
| V1 (Atual) | ~9 min | 1,172 | 1,173 | Simples |
| A (Paralelo 100) | ~2 min | 1,172 | 1,173 | Trivial |
| B (Chunk 100) | ~1 min | 12 | 1,173 | Médio |
| C (Streaming) | ~30s | ~10 | 1,173 | Alto |
| **D (Híbrido)** | **~1.5 min** | **24** | **1,173** | **Baixo** |

---

## Recomendação Final

### Fase 1 (AGORA): Quick Win - Opção A
```go
// Em waha_history_import_workflow.go
maxConcurrentChats := 100  // Era 20
```

**Impacto**: 9min → 2min (5x mais rápido)
**Esforço**: 5 minutos
**Risco**: Baixo

---

### Fase 2 (HOJE): Híbrido - Opção D
Implementar chunking simplificado com fetch paralelo.

**Impacto**: 9min → 1.5min (6x mais rápido, 48min → 8min no total)
**Esforço**: 1-2 horas
**Risco**: Baixo

---

### Fase 3 (FUTURO): Streaming - Opção C
Se ainda precisar mais performance (ex: imports de 10k+ chats).

**Impacto**: 9min → 30s (18x mais rápido)
**Esforço**: 1-2 dias
**Risco**: Médio

---

## Implementação Recomendada: Opção D

Vou implementar o hybrid approach (Quick Win + Batching) agora:

1. ✅ Aumentar paralelismo (5 min)
2. ✅ Implementar chunked batching (1-2h)
3. ✅ Testar com 30 dias (validação rápida)
4. ✅ Comparar V1 vs V3

**Objetivo**: Reduzir tempo de 9min → 1.5min (6x improvement)
