# Saga Orchestration - Temporal Workflows

## 📁 **Estrutura Implementada**

```
internal/workflows/saga/
├── saga_types.go                           ✅ CRIADO
├── process_inbound_message_saga.go         ✅ CRIADO
├── process_inbound_message_activities.go   🔄 PRÓXIMO
├── send_message_saga.go                    ⏳ PENDENTE
├── send_message_activities.go              ⏳ PENDENTE
└── README.md                               ✅ ESTE ARQUIVO
```

## ✅ **O que foi implementado:**

### 1️⃣ **saga_types.go**
Tipos compartilhados:
- `ProcessInboundMessageInput` - Input da saga
- `SagaState` - Estado mantido durante execução
- `ContactCreatedResult`, `SessionCreatedResult`, `MessageCreatedResult` - Resultados das activities
- `SendMessageInput`, `SendMessageState` - Para saga de envio

### 2️⃣ **process_inbound_message_saga.go**
Workflow principal com:
- ✅ 4 steps principais (Contact → Session → Message → Events)
- ✅ 2 steps opcionais (Debouncer, Tracking)
- ✅ Compensação automática em ordem REVERSA (LIFO)
- ✅ Retry configurável (3x, backoff exponencial)
- ✅ Logging detalhado

**Fluxo de compensação:**
```
Falha → DeleteMessage → CloseSession (se criado) → DeleteContact (se criado)
```

---

## 🔄 **Próximos Passos**

### **Activities a implementar** (`process_inbound_message_activities.go`):

#### **Forward Activities:**
1. **FindOrCreateContactActivity** - Busca ou cria contato
2. **FindOrCreateSessionActivity** - Busca ou cria sessão
3. **CreateMessageActivity** - Cria mensagem
4. **PublishDomainEventsActivity** - Publica eventos no outbox
5. **ProcessMessageDebouncerActivity** - Agrupa mensagens (opcional)
6. **TrackAdConversionActivity** - Rastreia conversão (opcional)

#### **Compensation Activities:**
7. **DeleteContactActivity** - Soft delete de contato
8. **CloseSessionActivity** - Fecha sessão forçadamente
9. **DeleteMessageActivity** - Soft delete de mensagem

---

## 🎯 **Integração com Use Case**

### **Antes (Transação direta):**
```go
func (uc *ProcessInboundMessageUseCase) Execute(ctx context.Context, cmd ProcessInboundMessageCommand) error {
    return uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
        // ... lógica ...
    })
}
```

### **Depois (Saga Orchestration):**
```go
func (uc *ProcessInboundMessageUseCase) Execute(ctx context.Context, cmd ProcessInboundMessageCommand) error {
    // Feature flag para migração gradual
    if uc.config.UseSagaOrchestration {
        return uc.executeViaSaga(ctx, cmd)
    }
    return uc.executeViaTransaction(ctx, cmd)
}

func (uc *ProcessInboundMessageUseCase) executeViaSaga(ctx context.Context, cmd ProcessInboundMessageCommand) error {
    input := saga.ProcessInboundMessageInput{
        ChannelMessageID: cmd.ChannelMessageID,
        ContactPhone:     cmd.ContactPhone,
        // ... map fields ...
    }

    workflowOptions := client.StartWorkflowOptions{
        ID:        fmt.Sprintf("process-inbound-%s", cmd.ChannelMessageID),
        TaskQueue: "message-processing",
    }

    we, err := uc.temporalClient.ExecuteWorkflow(ctx, workflowOptions, saga.ProcessInboundMessageSaga, input)
    if err != nil {
        return fmt.Errorf("failed to start saga: %w", err)
    }

    // Aguarda conclusão (síncrono para webhooks)
    return we.Get(ctx, nil)
}
```

---

## 📊 **Benefícios**

| Aspecto | Antes (Transaction) | Depois (Saga) |
|---------|---------------------|---------------|
| **Visibilidade** | Logs dispersos | Temporal UI com status real-time |
| **Retry** | Manual | Automático (3x, backoff) |
| **Compensação** | Manual | Automática (LIFO) |
| **Debug** | Difícil | Temporal UI + event history |
| **Timeout** | Fixo | Configurável por step |
| **Latência** | ~100ms | ~150ms (+50ms overhead Temporal) |

---

## 🧪 **Testing**

### **Unit Test Example:**
```go
func TestProcessInboundMessageSaga(t *testing.T) {
    suite := &testsuite.WorkflowTestSuite{}
    env := suite.NewTestWorkflowEnvironment()

    // Mock activities
    env.OnActivity("FindOrCreateContactActivity", mock.Anything, mock.Anything).
        Return(&saga.ContactCreatedResult{ContactID: testContactID, WasCreated: true}, nil)

    env.ExecuteWorkflow(saga.ProcessInboundMessageSaga, testInput)

    require.True(t, env.IsWorkflowCompleted())
    require.NoError(t, env.GetWorkflowError())
}
```

### **Compensation Test:**
```go
func TestProcessInboundMessageSaga_CompensationOnFailure(t *testing.T) {
    // Simula falha no Step 3
    env.OnActivity("CreateMessageActivity", mock.Anything, mock.Anything).
        Return(nil, errors.New("database error"))

    // Verifica compensação
    env.OnActivity("CloseSessionActivity", mock.Anything, mock.Anything).Return(nil)
    env.OnActivity("DeleteContactActivity", mock.Anything, mock.Anything).Return(nil)

    env.ExecuteWorkflow(saga.ProcessInboundMessageSaga, testInput)

    // Assert compensação executada
    env.AssertCalled(t, "CloseSessionActivity", mock.Anything, mock.Anything)
    env.AssertCalled(t, "DeleteContactActivity", mock.Anything, mock.Anything)
}
```

---

## 🚀 **Deployment Strategy**

### **Fase 1: Canary (10% tráfego)**
```env
USE_SAGA_ORCHESTRATION=true
SAGA_CANARY_PERCENTAGE=10
```

### **Fase 2: Ramp up (50% → 100%)**
```
Semana 1: 10% → 25%
Semana 2: 25% → 50%
Semana 3: 50% → 100%
```

### **Rollback:**
```env
USE_SAGA_ORCHESTRATION=false
```

---

## ⚙️ **Worker Registration**

```go
// cmd/worker/main.go
func registerSagaWorkers(client client.Client) {
    w := worker.New(client, "message-processing", worker.Options{})

    // Register workflow
    w.RegisterWorkflow(saga.ProcessInboundMessageSaga)

    // Register activities
    w.RegisterActivity(FindOrCreateContactActivity)
    w.RegisterActivity(FindOrCreateSessionActivity)
    w.RegisterActivity(CreateMessageActivity)
    w.RegisterActivity(PublishDomainEventsActivity)
    w.RegisterActivity(ProcessMessageDebouncerActivity)
    w.RegisterActivity(TrackAdConversionActivity)
    w.RegisterActivity(DeleteContactActivity)
    w.RegisterActivity(CloseSessionActivity)
    w.RegisterActivity(DeleteMessageActivity)

    err := w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatal("Saga worker failed:", err)
    }
}
```

---

## 📈 **Monitoramento**

### **Métricas Temporal UI:**
- Workflow execution time
- Activity retry count
- Compensation execution rate
- Failed workflows (alertar)

### **Dashboards:**
- Taxa de sucesso: >= 99.9%
- Latência P50/P95/P99
- Compensação rate: < 0.1%

---

**Status**: 🔄 2/3 arquivos implementados
**Próximo**: Implementar activities
**Estimativa**: 2-3 horas para completar
