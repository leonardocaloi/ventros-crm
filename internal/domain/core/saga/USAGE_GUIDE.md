# Saga Pattern - Developer Usage Guide

Este guia mostra como implementar novas Sagas no Ventros CRM.

---

## 🎯 Quick Start: Implementar Nova Saga

### Passo 1: Definir Saga Type

Adicione em `saga_types.go`:

```go
const (
    // ... existing sagas
    MinhaNovaSaga SagaType = "minha_nova_saga"  // ✅ ADD THIS
)

// Defina steps da Saga
const (
    StepPrimeiroEvento  SagaStep = "primeiro_evento"
    StepSegundoEvento   SagaStep = "segundo_evento"
    StepTerceiroEvento  SagaStep = "terceiro_evento"
)
```

### Passo 2: Implementar Use Case

```go
package myusecase

import (
    "context"
    "fmt"
    "github.com/ventros/crm/internal/domain/saga"
)

func (uc *MyUseCase) Execute(ctx context.Context, cmd MyCommand) error {
    // 🎬 Inicia Saga
    ctx = saga.WithSaga(ctx, string(saga.MinhaNovaSaga))
    ctx = saga.WithTenantID(ctx, cmd.TenantID)

    correlationID, _ := saga.GetCorrelationID(ctx)
    fmt.Printf("🎬 Saga started: MinhaNovaSaga (correlation_id: %s)\n", correlationID)

    // Step 1: Primeiro Evento
    ctx = saga.NextStep(ctx, saga.StepPrimeiroEvento)
    result1, err := uc.executarPrimeiroStep(ctx, cmd)
    if err != nil {
        return fmt.Errorf("saga step failed [primeiro_evento]: %w", err)
    }

    // Step 2: Segundo Evento
    ctx = saga.NextStep(ctx, saga.StepSegundoEvento)
    result2, err := uc.executarSegundoStep(ctx, result1)
    if err != nil {
        return fmt.Errorf("saga step failed [segundo_evento]: %w", err)
    }

    // Step 3: Terceiro Evento
    ctx = saga.NextStep(ctx, saga.StepTerceiroEvento)
    err = uc.executarTerceiroStep(ctx, result2)
    if err != nil {
        return fmt.Errorf("saga step failed [terceiro_evento]: %w", err)
    }

    // Publica eventos (correlation_id automático via DomainEventBus)
    uc.eventBus.PublishBatch(ctx, allEvents)

    fmt.Printf("✅ Saga completed: MinhaNovaSaga (correlation_id: %s)\n", correlationID)
    return nil
}

func (uc *MyUseCase) executarPrimeiroStep(ctx context.Context, cmd MyCommand) (*Result, error) {
    // Sua lógica aqui...
    // Quando publicar eventos, eles AUTOMATICAMENTE incluirão correlation_id!
    return result, nil
}
```

### Passo 3: Publicar Eventos (Automático!)

**IMPORTANTE:** Você NÃO precisa adicionar correlation_id manualmente!

```go
// ❌ ERRADO - Não precisa fazer isso!
event := myAggregate.DoSomething()
event.Metadata["correlation_id"] = correlationID

// ✅ CORRETO - DomainEventBus extrai do contexto automaticamente!
event := myAggregate.DoSomething()
uc.eventBus.Publish(ctx, event) // correlation_id injetado automaticamente!
```

### Passo 4: Definir Eventos de Compensação

Adicione em `compensation_events.go`:

```go
const (
    // ... existing compensation events
    CompensateMeuPrimeiroEvento CompensationEventType = "compensate.meu.primeiro_evento"
    CompensateMeuSegundoEvento  CompensationEventType = "compensate.meu.segundo_evento"
)

// Mapeamento de eventos → compensação
var CompensationMapping = map[string]CompensationEventType{
    // ... existing mappings
    "meu.primeiro_evento": CompensateMeuPrimeiroEvento,
    "meu.segundo_evento":  CompensateMeuSegundoEvento,
}
```

### Passo 5: Implementar Handler de Compensação (Opcional)

```go
// No arquivo de inicialização (ex: main.go, wire.go)
func setupSagaCoordinator(
    tracker *saga.SagaTracker,
    outboxRepo outbox.Repository,
) *saga.SagaCoordinator {
    coordinator := saga.NewSagaCoordinator(tracker, outboxRepo)

    // Registra handler de compensação para sua Saga
    coordinator.RegisterCompensationHandler(
        saga.MinhaNovaSaga,
        func(ctx context.Context, execution *saga.SagaExecution) error {
            fmt.Printf("🔄 Starting compensation for MinhaNovaSaga: %s\n", execution.CorrelationID)

            // Compensa na ordem reversa (LIFO)
            for i := len(execution.Events) - 1; i >= 0; i-- {
                event := execution.Events[i]

                // Só compensa eventos processados com sucesso
                if event.Status != outbox.StatusProcessed {
                    continue
                }

                // Dispara compensação baseado no tipo de evento
                switch event.EventType {
                case "meu.primeiro_evento":
                    // Lógica de compensação: desfazer primeiro evento
                    fmt.Printf("   Compensating: meu.primeiro_evento\n")
                    // TODO: Implementar compensação real

                case "meu.segundo_evento":
                    // Lógica de compensação: desfazer segundo evento
                    fmt.Printf("   Compensating: meu.segundo_evento\n")
                    // TODO: Implementar compensação real
                }
            }

            return nil
        },
    )

    return coordinator
}
```

---

## 🔍 Rastreamento de Sagas

### Via Código

```go
import "github.com/ventros/crm/internal/domain/saga"

// Criar tracker
tracker := saga.NewSagaTracker(outboxRepo)

// Buscar execução da Saga
execution, err := tracker.TrackSaga(ctx, correlationID)
if err != nil {
    return err
}

// Verificar status
fmt.Printf("Status: %s\n", execution.Status)          // "in_progress", "completed", "failed"
fmt.Printf("Total Steps: %d\n", execution.TotalSteps)
fmt.Printf("Completed: %d\n", execution.CompletedSteps)
fmt.Printf("Failed: %d\n", execution.FailedSteps)
fmt.Printf("Duration: %v\n", execution.Duration)

// Buscar apenas steps falhados
failedSteps, err := tracker.GetFailedSteps(ctx, correlationID)

// Buscar timeline completa
timeline, err := tracker.GetExecutionTimeline(ctx, correlationID)
for _, entry := range timeline {
    fmt.Printf("Step %d: %s [%s] at %v\n",
        entry.StepNumber, entry.SagaStep, entry.Status, entry.Timestamp)
}
```

### Via SQL

```sql
-- Buscar todos os eventos de uma Saga
SELECT
    event_type,
    status,
    metadata->>'saga_step' as step,
    metadata->>'step_number' as step_number,
    created_at,
    processed_at
FROM outbox_events
WHERE metadata->>'correlation_id' = 'abc-123-xyz'
ORDER BY created_at ASC;

-- Verificar status agregado
SELECT
    metadata->>'saga_type' as saga_type,
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE status = 'processed') as completed,
    COUNT(*) FILTER (WHERE status = 'failed') as failed,
    COUNT(*) FILTER (WHERE status = 'pending') as pending
FROM outbox_events
WHERE metadata->>'correlation_id' = 'abc-123-xyz'
GROUP BY metadata->>'saga_type';
```

---

## 💡 Patterns e Best Practices

### Pattern 1: Fast Path (Choreography)

**Quando usar:**
- Fluxos simples (< 5 steps)
- Baixa latência requerida (< 100ms)
- Steps independentes
- Ex: WAHA webhooks, status changes

**Exemplo:**
```go
// ✅ Fast Path - Direto, sem overhead
ctx = saga.WithSaga(ctx, string(saga.ProcessInboundMessageSaga))
ctx = saga.NextStep(ctx, saga.StepContactCreated)
contact := uc.createContact(ctx, cmd)
uc.eventBus.Publish(ctx, contact.DomainEvents()...)
```

### Pattern 2: Slow Path (Orchestration - Futuro)

**Quando usar:**
- Fluxos complexos (> 5 steps)
- Latência tolerável (> 1s)
- Steps com dependências complexas
- Ex: Onboarding, billing, bulk operations

**Exemplo:**
```go
// ⏳ Slow Path - Via Temporal (futuro)
workflowOptions := client.StartWorkflowOptions{
    ID:        correlationID,
    TaskQueue: "onboarding-queue",
}
execution, err := temporalClient.ExecuteWorkflow(
    ctx,
    workflowOptions,
    OnboardCustomerWorkflow,
    request,
)
```

### Pattern 3: Error Handling

```go
// ✅ Falha em qualquer step interrompe a Saga
ctx = saga.NextStep(ctx, saga.StepContactCreated)
contact, err := uc.createContact(ctx, cmd)
if err != nil {
    // Saga falha, correlation_id já está no contexto para debug
    return fmt.Errorf("saga step failed [contact_created]: %w", err)
}

// Se necessário, pode disparar compensação manualmente
correlationID, _ := saga.GetCorrelationID(ctx)
if err := coordinator.CompensateSaga(ctx, correlationID); err != nil {
    log.Printf("Compensation failed: %v", err)
}
```

### Pattern 4: Optional Steps

```go
// Steps opcionais não devem falhar a Saga
ctx = saga.NextStep(ctx, saga.StepTrackingCreated)
if err := uc.trackAdConversion(ctx, contact); err != nil {
    // Log mas NÃO retorna erro - tracking é opcional
    log.Printf("⚠️  Optional step failed [tracking]: %v", err)
}

// Saga continua...
ctx = saga.NextStep(ctx, saga.StepNextRequiredStep)
```

---

## 🧪 Testing

### Test Structure

```go
func TestMinhaNovaSaga(t *testing.T) {
    // Setup
    ctx := context.Background()
    ctx = saga.WithSaga(ctx, string(saga.MinhaNovaSaga))
    ctx = saga.WithTenantID(ctx, "tenant-123")

    // Execute
    useCase := NewMyUseCase(mockRepos...)
    err := useCase.Execute(ctx, cmd)

    // Assert
    assert.NoError(t, err)

    // Verify correlation_id was created
    correlationID, ok := saga.GetCorrelationID(ctx)
    assert.True(t, ok)
    assert.NotEmpty(t, correlationID)

    // Verify events were published with correlation_id
    events := mockEventBus.PublishedEvents()
    assert.NotEmpty(t, events)

    // TODO: Verify Outbox has events with correlation_id
}
```

### Integration Test

```go
func TestSagaEndToEnd(t *testing.T) {
    // Real database + Outbox
    db := setupTestDatabase(t)
    defer db.Close()

    // Execute Saga
    correlationID := executeSaga(t, db)

    // Query Outbox for Saga events
    var events []OutboxEvent
    db.Where("metadata->>'correlation_id' = ?", correlationID).Find(&events)

    // Verify all steps completed
    assert.Len(t, events, 3) // Expected 3 steps
    for _, event := range events {
        assert.Equal(t, "processed", event.Status)
    }

    // Verify timeline
    tracker := saga.NewSagaTracker(outboxRepo)
    execution, err := tracker.TrackSaga(context.Background(), correlationID)
    assert.NoError(t, err)
    assert.Equal(t, "completed", execution.Status)
    assert.Equal(t, 3, execution.TotalSteps)
    assert.Equal(t, 3, execution.CompletedSteps)
}
```

---

## 📊 Monitoring

### Metrics to Track

1. **Saga Completion Rate**
   ```sql
   SELECT
       COUNT(DISTINCT metadata->>'correlation_id') as total_sagas,
       COUNT(DISTINCT metadata->>'correlation_id') FILTER (
           WHERE status = 'processed'
       ) * 100.0 / COUNT(DISTINCT metadata->>'correlation_id') as completion_rate
   FROM outbox_events
   WHERE metadata->>'correlation_id' IS NOT NULL;
   ```

2. **Saga Duration (P50, P95, P99)**
   ```sql
   WITH saga_durations AS (
       SELECT
           metadata->>'correlation_id' as correlation_id,
           MAX(processed_at) - MIN(created_at) as duration
       FROM outbox_events
       WHERE metadata->>'correlation_id' IS NOT NULL
         AND processed_at IS NOT NULL
       GROUP BY metadata->>'correlation_id'
   )
   SELECT
       percentile_cont(0.5) WITHIN GROUP (ORDER BY duration) as p50,
       percentile_cont(0.95) WITHIN GROUP (ORDER BY duration) as p95,
       percentile_cont(0.99) WITHIN GROUP (ORDER BY duration) as p99
   FROM saga_durations;
   ```

3. **Failed Sagas (needs attention)**
   ```sql
   SELECT
       metadata->>'correlation_id' as correlation_id,
       metadata->>'saga_type' as saga_type,
       COUNT(*) FILTER (WHERE status = 'failed') as failed_steps,
       MAX(last_error) as error_message
   FROM outbox_events
   WHERE metadata->>'correlation_id' IS NOT NULL
     AND status = 'failed'
   GROUP BY 1, 2
   ORDER BY MAX(created_at) DESC
   LIMIT 10;
   ```

---

## 🐛 Debugging

### Problem: Saga não aparece no Outbox

**Causa:** Context sem Saga metadata

**Solução:**
```go
// ❌ ERRADO
ctx := context.Background()
uc.eventBus.Publish(ctx, event) // Sem correlation_id!

// ✅ CORRETO
ctx = saga.WithSaga(ctx, string(saga.MinhaNovaSaga))
uc.eventBus.Publish(ctx, event) // correlation_id injetado!
```

### Problem: Eventos sem correlation_id

**Causa:** EventBus não está extraindo metadata do contexto

**Verificar:** `domain_event_bus.go` deve ter:
```go
func (bus *DomainEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
    // ✅ Extrai Saga metadata
    var sagaMetadata map[string]interface{}
    if sagaMeta := saga.GetMetadata(ctx); sagaMeta != nil {
        sagaMetadata = map[string]interface{}{
            "correlation_id": sagaMeta.CorrelationID,
            // ...
        }
    }
    // ...
}
```

### Problem: Saga "travada" (in_progress por muito tempo)

**Debug Query:**
```sql
-- Sagas em progresso há mais de 5 minutos
SELECT
    metadata->>'correlation_id' as correlation_id,
    metadata->>'saga_type' as saga_type,
    MIN(created_at) as started_at,
    MAX(created_at) as last_event_at,
    NOW() - MAX(created_at) as stuck_duration,
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE status = 'processed') as completed,
    COUNT(*) FILTER (WHERE status = 'pending') as pending
FROM outbox_events
WHERE metadata->>'correlation_id' IS NOT NULL
GROUP BY 1, 2
HAVING NOW() - MAX(created_at) > INTERVAL '5 minutes'
   AND COUNT(*) FILTER (WHERE status = 'pending') > 0
ORDER BY stuck_duration DESC;
```

---

## 📚 Referências

- [SAGA_PATTERN_IMPLEMENTATION.md](../../SAGA_PATTERN_IMPLEMENTATION.md) - Documentação completa
- [saga_coordinator.go](./saga_coordinator.go) - Coordinator implementation
- [saga_tracker.go](./saga_tracker.go) - Tracking utilities
- [compensation_executor.go](./compensation_executor.go) - Compensation logic
- [process_inbound_message.go](../../application/message/process_inbound_message.go) - Exemplo real

**Happy Saga Implementing! 🎬**
