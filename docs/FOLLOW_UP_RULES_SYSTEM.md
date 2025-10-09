# Follow-up Rules System

## 📋 Visão Geral

O **Follow-up Rules System** é um motor de automação sofisticado que permite criar regras de follow-up automático baseadas em eventos e condições. Sistema completo com triggers, condições, ações, agendamento e execução distribuída.

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    Follow-up Rules System                    │
└─────────────────────────────────────────────────────────────┘

┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Rule Manager    │────▶│  Follow-up       │────▶│  Action          │
│  (CRUD + Ops)    │     │  Engine          │     │  Executor        │
└──────────────────┘     └──────────────────┘     └──────────────────┘
        │                         │                         │
        │                         │                         │
        ▼                         ▼                         ▼
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Rule            │     │  Event           │     │  External        │
│  Repository      │     │  Integration     │     │  Services        │
└──────────────────┘     └──────────────────┘     └──────────────────┘
        │                         │                         │
        │                         │                         │
        ▼                         ▼                         ▼
┌──────────────────────────────────────────────────────────────┐
│                       PostgreSQL + Redis                      │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│              Scheduled Rules Worker (Background)              │
│        Executa regras agendadas (cron, recorrentes)          │
└──────────────────────────────────────────────────────────────┘
```

## 🎯 Componentes Principais

### 1. **AutomationRule** (Domain Model)

Agrega raiz que representa uma regra de follow-up.

**Campos:**
- `id`: UUID único
- `pipelineID`: Pipeline dono da regra
- `tenantID`: Isolamento multi-tenant
- `name`: Nome da regra
- `description`: Descrição opcional
- `trigger`: Evento que dispara a regra
- `conditions`: Array de condições (AND logic)
- `actions`: Array de ações a executar
- `priority`: Ordem de execução (menor = maior prioridade)
- `enabled`: Se a regra está ativa

**Métodos:**
```go
func NewAutomationRule(pipelineID uuid.UUID, tenantID string, name string, trigger AutomationTrigger) (*AutomationRule, error)
func (r *AutomationRule) AddCondition(field, operator string, value interface{}) error
func (r *AutomationRule) AddAction(actionType AutomationAction, params map[string]interface{}, delayMinutes int) error
func (r *AutomationRule) EvaluateConditions(context map[string]interface{}) bool
func (r *AutomationRule) Enable()
func (r *AutomationRule) Disable()
```

### 2. **Triggers** (Eventos)

```go
const (
    // Session Events
    TriggerSessionEnded      = "session.ended"
    TriggerSessionTimeout    = "session.timeout"
    TriggerSessionResolved   = "session.resolved"
    TriggerSessionEscalated  = "session.escalated"

    // Message Events
    TriggerNoResponse        = "no_response.timeout"
    TriggerMessageReceived   = "message.received"

    // Pipeline Events
    TriggerStatusChanged     = "status.changed"
    TriggerStageCompleted    = "stage.completed"

    // Temporal Events
    TriggerAfterDelay        = "after.delay"
    TriggerScheduled         = "scheduled"
)
```

### 3. **Conditions** (Especificação Pattern)

```go
type RuleCondition struct {
    Field    string      `json:"field"`
    Operator string      `json:"operator"`
    Value    interface{} `json:"value"`
}
```

**Operadores suportados:**
- `eq`, `equals`: Igual
- `ne`, `not_equals`: Diferente
- `gt`, `greater_than`: Maior que
- `gte`, `greater_than_or_equal`: Maior ou igual
- `lt`, `less_than`: Menor que
- `lte`, `less_than_or_equal`: Menor ou igual
- `contains`: String contém
- `in`: Valor está no array

**Exemplo:**
```json
{
  "field": "hours_since_last_message",
  "operator": "gt",
  "value": 24
}
```

### 4. **Actions** (Strategy Pattern)

```go
const (
    ActionSendMessage        = "send_message"
    ActionSendTemplate       = "send_template"
    ActionChangeStatus       = "change_pipeline_status"
    ActionAssignAgent        = "assign_agent"
    ActionAssignToQueue      = "assign_to_queue"
    ActionCreateTask         = "create_task"
    ActionSendWebhook        = "send_webhook"
    ActionAddTag             = "add_tag"
    ActionRemoveTag          = "remove_tag"
    ActionUpdateCustomField  = "update_custom_field"
    ActionTriggerWorkflow    = "trigger_workflow"
)
```

**Exemplo de Ação:**
```json
{
  "type": "send_message",
  "params": {
    "content": "Olá! Vi que você não respondeu há 24h. Posso ajudar?"
  },
  "delay_minutes": 0
}
```

### 5. **Scheduled Rules** (Regras Agendadas)

Sistema de agendamento para triggers recorrentes.

**Tipos de Schedule:**
- `once`: Executa uma vez em data/hora específica
- `daily`: Executa diariamente em hora específica
- `weekly`: Executa semanalmente em dia da semana específico
- `monthly`: Executa mensalmente em dia do mês específico
- `cron`: Usa expressão cron customizada

**Exemplo:**
```go
schedule := pipeline.ScheduledRuleConfig{
    Type:       pipeline.ScheduleWeekly,
    DayOfWeek:  ptrInt(1), // segunda-feira
    Hour:       9,
    Minute:     0,
    StartTime:  time.Now(),
    EndTime:    nil, // sem fim
}
```

### 6. **AutomationEngine**

Engine que avalia e executa regras.

**Métodos principais:**
```go
func (e *AutomationEngine) EvaluateAndExecute(
    ctx context.Context,
    pipelineID uuid.UUID,
    trigger AutomationTrigger,
    evalContext map[string]interface{},
    actionCtx ActionContext,
) error

func (e *AutomationEngine) ProcessSessionEvent(...) error
func (e *AutomationEngine) ProcessContactEvent(...) error
func (e *AutomationEngine) ProcessScheduledTrigger(...) error
```

### 7. **AutomationRuleManager**

Gerenciador sofisticado para CRUD e operações complexas.

**Funcionalidades:**
- ✅ CRUD completo (Create, Read, Update, Delete)
- ✅ Enable/Disable em massa
- ✅ Duplicação de regras
- ✅ Export/Import JSON
- ✅ Reordenação de prioridades
- ✅ Estatísticas de regras
- ✅ Teste de condições
- ✅ Agendamento manual
- ✅ Validação sofisticada

**Exemplo de uso:**
```go
manager := pipeline.NewAutomationRuleManager(ruleRepo, pipelineRepo, nil, logger)

// Criar regra
rule, err := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    PipelineID: pipelineID,
    TenantID:   "tenant-123",
    Name:       "Follow-up 24h sem resposta",
    Trigger:    pipeline.TriggerNoResponse,
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
                "content": "Olá! Notei que você não respondeu. Posso ajudar?",
            },
        },
    },
    Enabled: true,
})

// Listar regras ativas
rules, _ := manager.ListEnabledRules(ctx, pipelineID)

// Duplicar regra
copy, _ := manager.DuplicateRule(ctx, rule.ID(), "Follow-up 48h")

// Exportar
json, _ := manager.ExportRule(ctx, rule.ID())

// Estatísticas
stats, _ := manager.GetRuleStatistics(ctx, pipelineID)
fmt.Printf("Total: %d, Enabled: %d\n", stats.Total, stats.Enabled)
```

### 8. **ActionExecutor**

Executa ações concretas usando serviços externos.

**Interfaces necessárias:**
```go
type MessageSender interface
type PipelineStatusChanger interface
type AgentAssigner interface
type QueueAssigner interface
type WebhookSender interface
type TagManager interface
type CustomFieldUpdater interface
type WorkflowTrigger interface
```

### 9. **ScheduledRulesWorker**

Worker background que executa regras agendadas.

**Características:**
- Poll interval configurável (default: 1 minuto)
- Query otimizada com index `idx_automation_scheduled_ready`
- Atualiza `last_executed` e `next_execution` automaticamente
- Desativa regras `once` após execução
- Graceful shutdown

**Startup:**
```go
worker := workflow.NewScheduledRulesWorker(
    db,
    followUpEngine,
    1*time.Minute,
    logger,
)

go worker.Start(ctx)
```

## 📊 Modelo de Dados

### Tabela: `automation_rules`

```sql
CREATE TABLE automation_rules (
    id UUID PRIMARY KEY,
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    trigger VARCHAR(100) NOT NULL,
    conditions JSONB DEFAULT '[]',
    actions JSONB DEFAULT '[]',
    priority INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    schedule JSONB,
    last_executed TIMESTAMP WITH TIME ZONE,
    next_execution TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
```

**Indexes:**
- `idx_automation_pipeline`: (pipeline_id)
- `idx_automation_trigger`: (trigger)
- `idx_automation_enabled`: (enabled)
- `idx_automation_active_rules`: (pipeline_id, trigger, priority) WHERE enabled = true
- `idx_automation_scheduled_ready`: (next_execution, enabled) WHERE trigger = 'scheduled'

## 🔥 Exemplos de Uso

### Exemplo 1: Follow-up 24h sem resposta

```go
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    PipelineID: pipelineID,
    TenantID:   tenantID,
    Name:       "Follow-up 24h inativo",
    Description: "Envia mensagem se contato não responde há 24h",
    Trigger:    pipeline.TriggerNoResponse,
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "hours_since_last_message",
            Operator: "gte",
            Value:    24,
        },
        {
            Field:    "message_count",
            Operator: "gt",
            Value:    0,
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendMessage,
            Params: map[string]interface{}{
                "content": "Olá! Vi que você não respondeu. Ainda precisa de ajuda?",
            },
        },
        {
            Type: pipeline.ActionAddTag,
            Params: map[string]interface{}{
                "tag": "automation_sent",
            },
        },
    },
    Priority: 10,
    Enabled:  true,
})
```

### Exemplo 2: Mover para outro status após 3 dias

```go
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:    "Mover para 'Perdido' após 3 dias",
    Trigger: pipeline.TriggerNoResponse,
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "hours_since_last_message",
            Operator: "gte",
            Value:    72, // 3 dias
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionChangeStatus,
            Params: map[string]interface{}{
                "status_id": lostStatusID.String(),
            },
        },
        {
            Type: pipeline.ActionSendWebhook,
            Params: map[string]interface{}{
                "url": "https://meu-sistema.com/webhook/lead-lost",
            },
        },
    },
    Enabled: true,
})
```

### Exemplo 3: Envio semanal de newsletter

```go
schedule := pipeline.ScheduledRuleConfig{
    Type:      pipeline.ScheduleWeekly,
    DayOfWeek: ptrInt(1), // segunda-feira
    Hour:      10,
    Minute:    0,
}

rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:    "Newsletter semanal",
    Trigger: pipeline.TriggerScheduled,
    Schedule: &schedule,
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendTemplate,
            Params: map[string]interface{}{
                "template_name": "newsletter_template",
                "params": map[string]interface{}{
                    "week": "{{current_week}}",
                },
            },
        },
    },
    Enabled: true,
})
```

### Exemplo 4: Lembrete mensal de renovação

```go
schedule := pipeline.ScheduledRuleConfig{
    Type:       pipeline.ScheduleMonthly,
    DayOfMonth: ptrInt(1), // dia 1 de cada mês
    Hour:       9,
    Minute:     0,
}

rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:     "Lembrete mensal de renovação",
    Trigger:  pipeline.TriggerScheduled,
    Schedule: &schedule,
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "subscription_expiring_in_days",
            Operator: "lte",
            Value:    30,
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendMessage,
            Params: map[string]interface{}{
                "content": "Sua assinatura vence em breve! Renove agora.",
            },
        },
        {
            Type: pipeline.ActionTriggerWorkflow,
            Params: map[string]interface{}{
                "workflow_name": "renewal_reminder_workflow",
            },
        },
    },
    Enabled: true,
})
```

### Exemplo 5: Atribuir agente após 10 mensagens

```go
rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:    "Escalar para humano após 10 mensagens",
    Trigger: pipeline.TriggerMessageReceived,
    Conditions: []pipeline.RuleCondition{
        {
            Field:    "message_count",
            Operator: "gte",
            Value:    10,
        },
        {
            Field:    "agent_id",
            Operator: "eq",
            Value:    nil, // sem agente atribuído
        },
    },
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionAssignToQueue,
            Params: map[string]interface{}{
                "queue_id": humanQueueID.String(),
            },
        },
        {
            Type: pipeline.ActionAddTag,
            Params: map[string]interface{}{
                "tag": "needs_human_agent",
            },
        },
    },
    Priority: 5,
    Enabled:  true,
})
```

## 🔧 Integração com Eventos

### 1. Session Events

```go
integration := pipeline.NewAutomationIntegration(
    engine,
    sessionRepo,
    pipelineRepo,
    logger,
)

// Quando sessão encerra
integration.OnSessionEnded(ctx, sessionID)

// Quando sessão expira por timeout
integration.OnSessionTimeout(ctx, sessionID)

// Quando sessão é resolvida
integration.OnSessionResolved(ctx, sessionID)
```

### 2. Message Events

```go
// Quando mensagem é recebida
integration.OnMessageReceived(ctx, sessionID, messageID)

// Quando contato não responde há X horas
integration.OnNoResponse(ctx, sessionID, hoursSinceLastMessage)
```

### 3. Pipeline Events

```go
// Quando status muda
integration.OnStatusChanged(ctx, contactID, pipelineID, oldStatusID, newStatusID, tenantID)
```

## 📈 Monitoramento e Métricas

```
# Regras
automation_rules_total{pipeline_id, trigger, enabled}
automation_rules_executions_total{rule_id, trigger, status="success|failed"}
automation_rules_execution_duration_seconds{rule_id}

# Condições
automation_conditions_evaluated_total{rule_id}
automation_conditions_matched_total{rule_id}

# Ações
automation_actions_executed_total{rule_id, action_type, status}
automation_actions_duration_seconds{action_type}

# Worker
automation_scheduled_worker_cycles_total
automation_scheduled_rules_processed_total
automation_scheduled_rules_errors_total
```

## 🎛️ Configuração

### Variáveis de Ambiente

```bash
# Engine
FOLLOWUP_ENABLED=true
FOLLOWUP_MAX_CONCURRENT_RULES=10

# Scheduled Worker
FOLLOWUP_WORKER_ENABLED=true
FOLLOWUP_WORKER_POLL_INTERVAL=1m

# Actions
FOLLOWUP_ACTION_TIMEOUT=30s
FOLLOWUP_MESSAGE_SENDER_ENABLED=true
FOLLOWUP_WEBHOOK_TIMEOUT=10s
```

## 🧪 Testes

```go
// Testar condições
match, _ := manager.TestRuleConditions(ctx, ruleID, map[string]interface{}{
    "hours_since_last_message": 25,
    "message_count": 5,
})
fmt.Println("Conditions match:", match)

// Testar ação manualmente
executor.Execute(ctx, pipeline.RuleAction{
    Type: pipeline.ActionSendMessage,
    Params: map[string]interface{}{"content": "Test"},
}, actionCtx)
```

## 🚀 Performance

**Otimizações implementadas:**
- ✅ Indexes parciais para regras ativas
- ✅ Index composto para query pattern comum
- ✅ Index para worker de regras agendadas
- ✅ JSONB para conditions/actions (query eficiente)
- ✅ Priority ordering no banco
- ✅ Batch updates para reordenação

**Benchmarks esperados:**
- Query de regras ativas: < 5ms
- Avaliação de condições: < 1ms
- Execução de regra simples: < 50ms
- Worker cycle: < 100ms (sem regras prontas)

## 🔐 Segurança

- ✅ Tenant isolation em todas as queries
- ✅ Validação de entrada no manager
- ✅ Sanitização de params de ações
- ✅ Rate limiting sugerido para webhooks
- ✅ Audit log de execuções (via domain events)

## 📚 Referências

- **Specification Pattern**: [Martin Fowler](https://martinfowler.com/apsupp/spec.pdf)
- **Strategy Pattern**: Gang of Four Design Patterns
- **Event-Driven Architecture**: Event Sourcing & CQRS
- **Cron Expressions**: [crontab.guru](https://crontab.guru/)
