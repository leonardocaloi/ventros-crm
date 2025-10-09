# Trigger System - Follow-up Rules

## 📋 O que são Triggers?

**Triggers** são eventos que disparam a avaliação de Follow-up Rules. Quando um trigger ocorre, o sistema:

1. Busca todas as regras ativas daquele Pipeline com aquele trigger
2. Avalia as condições de cada regra
3. Executa as ações das regras cujas condições foram satisfeitas

## 🎯 Tipos de Triggers

### **System Triggers (Hard-coded)**

10 triggers pré-definidos no código, sempre disponíveis para todos os tenants:

#### 🔵 **Session Triggers (4)**

| Trigger | Código | Quando Dispara |
|---------|--------|----------------|
| **Sessão Encerrada** | `session.ended` | Sessão encerra normalmente |
| **Sessão Expirou** | `session.timeout` | Sessão expira por inatividade |
| **Sessão Resolvida** | `session.resolved` | Sessão marcada como resolvida |
| **Sessão Escalada** | `session.escalated` | Sessão escalada para outro nível |

#### 💬 **Message Triggers (2)**

| Trigger | Código | Quando Dispara |
|---------|--------|----------------|
| **Sem Resposta** | `no_response.timeout` | Cliente não responde há X tempo |
| **Mensagem Recebida** | `message.received` | Nova mensagem é recebida |

#### 🎯 **Pipeline Triggers (2)**

| Trigger | Código | Quando Dispara |
|---------|--------|----------------|
| **Status Mudou** | `status.changed` | Status do contato muda no pipeline |
| **Etapa Concluída** | `stage.completed` | Etapa do pipeline é concluída |

#### ⏰ **Temporal Triggers (2)**

| Trigger | Código | Quando Dispara |
|---------|--------|----------------|
| **Após Delay** | `after.delay` | Após delay específico desde evento |
| **Agendado** | `scheduled` | Em horários agendados (cron, recorrente) |

### **Custom Triggers (User-defined)**

Triggers customizados que podem ser criados por admins/tenants para casos específicos.

**Regras:**
- ✅ Devem começar com prefixo `custom.`
- ✅ Exemplo: `custom.payment_received`, `custom.nps_sent`
- ✅ Ilimitados por tenant
- ❌ Não podem sobrescrever system triggers

## 📊 Parâmetros Disponíveis por Trigger

Cada trigger disponibiliza parâmetros específicos no contexto de avaliação das condições:

### `session.ended`

```json
{
  "session_id": "uuid",
  "contact_id": "uuid",
  "channel_id": "uuid",
  "session_duration_minutes": 45.5,
  "message_count": 12,
  "resolved": true,
  "agent_id": "uuid"
}
```

**Exemplo de condição:**
```json
{
  "field": "session_duration_minutes",
  "operator": "gt",
  "value": 30
}
```

### `no_response.timeout`

```json
{
  "session_id": "uuid",
  "contact_id": "uuid",
  "hours_since_last_message": 25.3,
  "last_message_at": "2025-01-15T10:30:00Z",
  "message_count": 5
}
```

**Exemplo de condição:**
```json
{
  "field": "hours_since_last_message",
  "operator": "gte",
  "value": 24
}
```

### `status.changed`

```json
{
  "contact_id": "uuid",
  "pipeline_id": "uuid",
  "old_status_id": "uuid",
  "new_status_id": "uuid",
  "old_status_name": "Lead",
  "new_status_name": "Qualificado"
}
```

### `scheduled`

```json
{
  "scheduled_at": "2025-01-15T10:00:00Z",
  "schedule_type": "weekly",
  "day_of_week": 1,
  "hour": 10,
  "minute": 0
}
```

## 🔧 Como Usar System Triggers

### Listar triggers disponíveis

```go
registry := pipeline.NewTriggerRegistry()

// Listar todos os system triggers
systemTriggers := registry.ListSystemTriggers()
for _, trigger := range systemTriggers {
    fmt.Printf("%s (%s): %s\n",
        trigger.Name,
        trigger.Code,
        trigger.Description,
    )
}

// Listar por categoria
sessionTriggers := registry.ListTriggersByCategory(pipeline.CategorySession)
```

### Ver parâmetros de um trigger

```go
params, _ := registry.GetParametersForTrigger("no_response.timeout")
for _, param := range params {
    fmt.Printf("- %s (%s): %s\n",
        param.Name,
        param.Type,
        param.Description,
    )
}

// Output:
// - hours_since_last_message (float): Horas desde última mensagem
// - last_message_at (timestamp): Timestamp da última mensagem
// - message_count (int): Total de mensagens na sessão
```

### Criar regra com system trigger

```go
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    PipelineID: pipelineID,
    TenantID:   "tenant-123",
    Name:       "Follow-up 24h sem resposta",
    Trigger:    pipeline.TriggerNoResponse, // ← System trigger
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "hours_since_last_message",
            Operator: "gte",
            Value:    24,
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendMessage,
            Params: map[string]interface{}{
                "content": "Olá! Ainda precisa de ajuda?",
            },
        },
    },
})
```

## 🎨 Como Usar Custom Triggers

### 1. Registrar custom trigger

```go
registry := pipeline.NewTriggerRegistry()

err := registry.RegisterCustomTrigger(pipeline.TriggerMetadata{
    Code:        "custom.payment_received",
    Name:        "Pagamento Recebido",
    Description: "Disparado quando webhook de pagamento confirma recebimento",
    Parameters: []pipeline.TriggerParameter{
        {
            Name:        "payment_id",
            Type:        "uuid",
            Description: "ID da transação",
        },
        {
            Name:        "amount",
            Type:        "float",
            Description: "Valor pago",
        },
        {
            Name:        "payment_method",
            Type:        "string",
            Description: "Método de pagamento (pix, card, etc)",
        },
    },
})

if err != nil {
    // Erro: trigger já existe ou nome inválido
}
```

### 2. Criar regra com custom trigger

```go
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    PipelineID: pipelineID,
    Name:       "Agradecer pagamento PIX",
    Trigger:    pipeline.AutomationTrigger("custom.payment_received"), // ← Custom trigger
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "payment_method",
            Operator: "eq",
            Value:    "pix",
        },
        {
            Field:    "amount",
            Operator: "gt",
            Value:    100.0,
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendMessage,
            Params: map[string]interface{}{
                "content": "Pagamento via PIX confirmado! Obrigado pela compra. 🎉",
            },
        },
        {
            Type: pipeline.ActionChangeStatus,
            Params: map[string]interface{}{
                "status_id": paidStatusID.String(),
            },
        },
    },
})
```

### 3. Disparar custom trigger

```go
// Via webhook ou evento customizado
engine := pipeline.NewAutomationEngine(ruleRepo, actionExecutor, logger)

// Contexto com parâmetros customizados
evalContext := map[string]interface{}{
    "payment_id":     paymentID,
    "amount":         150.50,
    "payment_method": "pix",
    "occurred_at":    time.Now(),
}

actionCtx := pipeline.ActionContext{
    ContactID:  &contactID,
    PipelineID: pipelineID,
    TenantID:   tenantID,
    Trigger:    "custom.payment_received",
    Metadata:   evalContext,
}

// Dispara avaliação de regras
engine.EvaluateAndExecute(
    ctx,
    pipelineID,
    "custom.payment_received",
    evalContext,
    actionCtx,
)
```

## 🔐 Validação e Segurança

### Validação de triggers

O `DefaultRuleValidator` valida automaticamente se o trigger existe:

```go
validator := pipeline.NewDefaultRuleValidator(registry)

// Tenta criar regra com trigger inválido
rule, err := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Trigger: "invalid.trigger", // ← Não registrado
    // ...
})
// Erro: "invalid trigger: invalid.trigger (not registered)"
```

### Regras de segurança

1. ✅ **Custom triggers DEVEM começar com `custom.`**
   ```go
   registry.RegisterCustomTrigger(pipeline.TriggerMetadata{
       Code: "payment_received", // ❌ Inválido
   })
   // Erro: "custom triggers must start with 'custom.' prefix"
   ```

2. ✅ **Não pode sobrescrever system triggers**
   ```go
   registry.RegisterCustomTrigger(pipeline.TriggerMetadata{
       Code: "session.ended", // ❌ Inválido (system trigger)
   })
   // Erro: "cannot override system trigger: session.ended"
   ```

3. ✅ **Não pode remover system triggers**
   ```go
   registry.UnregisterCustomTrigger("session.ended")
   // Erro: "cannot unregister system trigger"
   ```

## 📚 Exemplos Práticos

### Exemplo 1: E-commerce - Carrinho Abandonado

```go
// System trigger: no_response.timeout
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:    "Recuperar carrinho abandonado",
    Trigger: pipeline.TriggerNoResponse,
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "hours_since_last_message",
            Operator: "eq",
            Value:    1, // exatamente 1 hora
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendTemplate,
            Params: map[string]interface{}{
                "template_name": "carrinho_abandonado",
                "params": map[string]interface{}{
                    "cupom": "VOLTA10",
                },
            },
        },
    },
})
```

### Exemplo 2: SaaS - Trial Expirando

```go
// Custom trigger: custom.trial_expiring
registry.RegisterCustomTrigger(pipeline.TriggerMetadata{
    Code:        "custom.trial_expiring",
    Name:        "Trial Expirando",
    Description: "Disparado 3 dias antes do trial expirar",
    Parameters: []pipeline.TriggerParameter{
        {Name: "days_remaining", Type: "int"},
        {Name: "user_email", Type: "string"},
        {Name: "plan_name", Type: "string"},
    },
})

rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:    "Lembrete trial expirando",
    Trigger: "custom.trial_expiring",
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "days_remaining",
            Operator: "eq",
            Value:    3,
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendTemplate,
            Params: map[string]interface{}{
                "template_name": "trial_expiring_3days",
            },
        },
        {
            Type: pipeline.ActionAddTag,
            Params: map[string]interface{}{
                "tag": "trial_expiring_soon",
            },
        },
    },
})
```

### Exemplo 3: Suporte - Escalação Automática

```go
// System trigger: message.received
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:    "Escalar após 5 mensagens",
    Trigger: pipeline.TriggerMessageReceived,
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "message_count",
            Operator: "gte",
            Value:    5,
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionAssignToQueue,
            Params: map[string]interface{}{
                "queue_id": specialistQueueID.String(),
            },
        },
        {
            Type: pipeline.ActionSendWebhook,
            Params: map[string]interface{}{
                "url": "https://api.slack.com/webhook",
                "payload": map[string]interface{}{
                    "text": "Cliente precisa de especialista (5+ mensagens)",
                },
            },
        },
    },
})
```

## 🎯 API HTTP para Triggers

### Listar triggers disponíveis

```http
GET /api/v1/triggers
Authorization: Bearer {token}
```

**Response:**
```json
{
  "system_triggers": [
    {
      "code": "session.ended",
      "name": "Sessão Encerrada",
      "description": "Disparado quando uma sessão é encerrada normalmente",
      "category": "session",
      "is_system": true,
      "parameters": [
        {
          "name": "session_id",
          "type": "uuid",
          "description": "ID da sessão"
        },
        {
          "name": "message_count",
          "type": "int",
          "description": "Total de mensagens na sessão"
        }
      ]
    }
  ],
  "custom_triggers": [
    {
      "code": "custom.payment_received",
      "name": "Pagamento Recebido",
      "category": "custom",
      "is_system": false
    }
  ]
}
```

### Registrar custom trigger

```http
POST /api/v1/triggers/custom
Authorization: Bearer {token}
Content-Type: application/json

{
  "code": "custom.nps_survey_completed",
  "name": "Pesquisa NPS Concluída",
  "description": "Disparado quando cliente completa pesquisa NPS",
  "parameters": [
    {
      "name": "score",
      "type": "int",
      "description": "Score NPS (0-10)"
    },
    {
      "name": "feedback",
      "type": "string",
      "description": "Feedback opcional"
    }
  ]
}
```

### Ver parâmetros de trigger

```http
GET /api/v1/triggers/{trigger_code}/parameters
Authorization: Bearer {token}
```

## 🔄 Fluxo Completo

```
1. Evento acontece no sistema
   └─ Ex: Cliente não responde há 24h

2. Sistema identifica o trigger
   └─ trigger = "no_response.timeout"

3. TriggerRegistry valida
   └─ registry.IsValidTrigger("no_response.timeout") → true

4. Sistema monta contexto de avaliação
   └─ evalContext = {
        "hours_since_last_message": 24.5,
        "message_count": 5,
        "contact_id": "uuid",
        ...
      }

5. AutomationEngine busca regras
   └─ SELECT * FROM automation_rules
      WHERE pipeline_id = ?
      AND trigger = 'no_response.timeout'
      AND enabled = true
      ORDER BY priority ASC

6. Para cada regra:
   ├─ Avalia condições: rule.EvaluateConditions(evalContext)
   │  └─ Ex: hours_since_last_message >= 24? → true
   │
   └─ Executa ações: actionExecutor.Execute(action, actionCtx)
      └─ Ex: SendMessage("Olá! Ainda precisa de ajuda?")
```

## 📊 Resumo

| Aspecto | System Triggers | Custom Triggers |
|---------|-----------------|-----------------|
| **Quantidade** | 10 (fixo) | Ilimitado |
| **Definição** | Hard-coded no Go | Registrados via API |
| **Prefixo** | Vários (`session.`, `message.`, etc) | Obrigatório `custom.` |
| **Validação** | Compile-time | Runtime |
| **Listeners** | Implementados no código | Disparo manual/webhook |
| **Modificação** | Requer deploy | Via API |
| **Uso** | Todos os tenants | Por tenant |

**Recomendação:** Use **System Triggers** sempre que possível. Use **Custom Triggers** apenas para casos muito específicos do seu negócio.
