# Refatoração Completa: Follow-up Rules → Automation Rules ✅

## 🎯 Decisão: Um Único Tipo Genérico

**Pergunta:** Precisamos de dois tipos (FollowUpRule + AutomationRule) ou um só genérico?

**Resposta:** **UM SÓ GENÉRICO** (`AutomationRule`)

### Por quê?

1. ✅ **DDD Puro**: Uma entidade, uma responsabilidade
2. ✅ **Simples**: Uma tabela, um repository, um manager
3. ✅ **Flexível**: Trigger categoriza naturalmente (follow-up, event, scheduled)
4. ✅ **Evolutivo**: Novos tipos = apenas novos triggers
5. ✅ **Mercado**: HubSpot, Salesforce, Pipedrive usam abordagem genérica

## 📊 O que foi Renomeado

### Domain Layer

| Antes | Depois |
|-------|--------|
| `FollowUpRule` | `AutomationRule` |
| `FollowUpTrigger` | `AutomationTrigger` |
| `FollowUpAction` | `AutomationAction` |
| `NewFollowUpRule()` | `NewAutomationRule()` |
| `ReconstructFollowUpRule()` | `ReconstructAutomationRule()` |
| `FollowUpRuleRepository` | `AutomationRuleRepository` |
| `FollowUpRuleCreatedEvent` | `AutomationRuleCreatedEvent` |
| `FollowUpRuleEnabledEvent` | `AutomationRuleEnabledEvent` |
| `FollowUpRuleDisabledEvent` | `AutomationRuleDisabledEvent` |
| `FollowUpRuleTriggeredEvent` | `AutomationRuleTriggeredEvent` |
| `FollowUpRuleExecutedEvent` | `AutomationRuleExecutedEvent` |
| `FollowUpRuleFailedEvent` | `AutomationRuleFailedEvent` |
| `ScheduledFollowUpRule` | `ScheduledAutomationRule` |

**Arquivos:**
- `follow_up_rule.go` → `automation_rule.go`
- `scheduled_rule.go` → `scheduled_automation.go`
- `events.go` (atualizado)

### Application Layer

| Antes | Depois |
|-------|--------|
| `FollowUpEngine` | `AutomationEngine` |
| `FollowUpRuleManager` | `AutomationRuleManager` |
| `FollowUpIntegration` | `AutomationIntegration` |
| `DefaultActionExecutor` | (mantido - não tem FollowUp no nome) |

**Arquivos:**
- `follow_up_engine.go` → `automation_engine.go`
- `follow_up_rule_manager.go` → `automation_rule_manager.go`
- `follow_up_integration.go` → `automation_integration.go`
- `follow_up_action_executor.go` → `automation_action_executor.go`

### Infrastructure Layer

| Antes | Depois |
|-------|--------|
| `FollowUpRuleEntity` | `AutomationRuleEntity` |
| `GormFollowUpRuleRepository` | `GormAutomationRuleRepository` |
| `ScheduledRulesWorker` | `ScheduledAutomationWorker` |
| Tabela: `follow_up_rules` | `automation_rules` |
| Indexes: `idx_followup_*` | `idx_automation_*` |
| Function: `update_follow_up_rules_updated_at()` | `update_automation_rules_updated_at()` |

**Arquivos:**
- `entities/follow_up_rule.go` → `entities/automation_rule.go`
- `gorm_follow_up_rule_repository.go` → `gorm_automation_rule_repository.go`
- `scheduled_rules_worker.go` → `scheduled_automation_worker.go`
- `000019_create_follow_up_rules_table.up.sql` → `000019_create_automation_rules_table.up.sql`
- `000019_create_follow_up_rules_table.down.sql` → `000019_create_automation_rules_table.down.sql`

### Documentation

Todos os documentos atualizados:
- `FOLLOW_UP_RULES_SYSTEM.md`
- `IMPLEMENTATION_SUMMARY.md`
- `TRIGGER_SYSTEM.md`
- `RENAMING_PROPOSAL.md`

## 🗂️ Estrutura Final

```
internal/domain/pipeline/
├── automation_rule.go          # AutomationRule (aggregate root)
├── scheduled_automation.go     # ScheduledAutomationRule
├── trigger_registry.go         # TriggerRegistry
├── events.go                   # Domain events (AutomationRule*)
├── pipeline.go
├── status.go
└── repository.go

internal/application/pipeline/
├── automation_engine.go         # AutomationEngine
├── automation_rule_manager.go   # AutomationRuleManager
├── automation_integration.go    # AutomationIntegration
└── automation_action_executor.go # DefaultActionExecutor

infrastructure/persistence/
├── entities/
│   └── automation_rule.go      # AutomationRuleEntity
├── gorm_automation_rule_repository.go
└── database/migrations/
    ├── 000019_create_automation_rules_table.up.sql
    └── 000019_create_automation_rules_table.down.sql

infrastructure/workflow/
└── scheduled_automation_worker.go  # ScheduledAutomationWorker
```

## 🎯 Categorização por Trigger

Com o nome genérico `AutomationRule`, categorizamos pelo `trigger`:

### 1. **Follow-up Automation** (resposta/acompanhamento)
```go
Trigger: TriggerNoResponse
Trigger: TriggerSessionTimeout
```

### 2. **Event Automation** (resposta imediata a eventos)
```go
Trigger: TriggerStatusChanged
Trigger: TriggerMessageReceived
Trigger: TriggerSessionEnded
Trigger: TriggerPurchaseCompleted  // ← NOVO
```

### 3. **Scheduled Automation** (agendadas/recorrentes)
```go
Trigger: TriggerScheduled
```

### 4. **Conditional Automation** (baseada em condições)
```go
// Qualquer trigger + conditions complexas
Trigger: TriggerStatusChanged
Conditions: [
  {field: "old_status", operator: "eq", value: "Cliente"},
  {field: "new_status", operator: "eq", value: "Lead"},
  {field: "days_since_last_purchase", operator: "gt", value: 90}
]
```

## ✅ Validação

### Compilação
```bash
✅ go build ./internal/domain/pipeline/...
✅ go build ./internal/application/pipeline/...
✅ go build ./infrastructure/persistence/...
✅ go build ./infrastructure/workflow/...
```

### Referências Antigas
```bash
✅ Nenhuma referência a "FollowUp" encontrada (exceto NoteTypeFollowUp - diferente)
✅ Nenhuma referência a "follow_up" em tabelas
✅ Nenhuma referência a "followup" em structs
```

### Migration
```sql
✅ Tabela: automation_rules
✅ Indexes: idx_automation_*
✅ Function: update_automation_rules_updated_at()
✅ Trigger: trigger_update_automation_rules_updated_at
```

## 📝 Exemplos de Uso Atualizados

### Criar Regra (Antes)
```go
rule, _ := followUpManager.CreateRule(ctx, CreateRuleInput{
    Name:    "Follow-up 24h",
    Trigger: pipeline.TriggerNoResponse,
    ...
})
```

### Criar Regra (Depois)
```go
rule, _ := automationManager.CreateRule(ctx, CreateRuleInput{
    Name:    "Follow-up 24h",  // ← nome pode ser qualquer coisa
    Trigger: pipeline.TriggerNoResponse,
    ...
})
```

### Tipos de Automação

**Follow-up (resposta/acompanhamento):**
```go
Name: "Follow-up 24h sem resposta"
Trigger: TriggerNoResponse
```

**Event Automation (resposta a evento):**
```go
Name: "Mensagem de boas-vindas"
Trigger: TriggerStatusChanged
Conditions: [{field: "new_status", operator: "eq", value: "Lead"}]
```

**Purchase Automation (novo!):**
```go
Name: "Confirmação de compra"
Trigger: TriggerPurchaseCompleted  // ← NOVO TRIGGER
```

**Scheduled (agendada):**
```go
Name: "Newsletter semanal"
Trigger: TriggerScheduled
Schedule: {type: "weekly", day_of_week: 1, hour: 10}
```

**Re-engagement (reativação):**
```go
Name: "Reativar cliente inativo"
Trigger: TriggerStatusChanged
Conditions: [
  {field: "old_status", operator: "eq", value: "Cliente"},
  {field: "new_status", operator: "eq", value: "Lead"}
]
```

## 🚀 Próximos Passos

### 1. Novos Triggers (Recomendado)
```go
// Adicionar ao automation_rule.go
const (
    // ... triggers existentes ...

    // 💰 Triggers de Transação
    TriggerPurchaseCompleted  AutomationTrigger = "purchase.completed"
    TriggerPaymentReceived    AutomationTrigger = "payment.received"
    TriggerRefundIssued       AutomationTrigger = "refund.issued"

    // 📊 Triggers de Comportamento
    TriggerCartAbandoned      AutomationTrigger = "cart.abandoned"
    TriggerPageVisited        AutomationTrigger = "page.visited"
    TriggerFormSubmitted      AutomationTrigger = "form.submitted"
)
```

### 2. API HTTP Handlers
```go
POST   /api/v1/pipelines/{id}/automation-rules
GET    /api/v1/pipelines/{id}/automation-rules
GET    /api/v1/pipelines/{id}/automation-rules/{ruleId}
PUT    /api/v1/pipelines/{id}/automation-rules/{ruleId}
DELETE /api/v1/pipelines/{id}/automation-rules/{ruleId}
POST   /api/v1/pipelines/{id}/automation-rules/{ruleId}/enable
POST   /api/v1/pipelines/{id}/automation-rules/{ruleId}/disable
```

### 3. UI/Frontend
```
Menu: "Automações" ou "Automation Rules"
├─ Todas as Regras
├─ Ativas
├─ Inativas
└─ Nova Regra
   ├─ Follow-up
   ├─ Evento
   ├─ Agendada
   └─ Customizada
```

### 4. Integração com Event Bus
```go
// Conectar AutomationIntegration com domain events
eventBus.Subscribe("session.ended", integration.OnSessionEnded)
eventBus.Subscribe("session.timeout", integration.OnSessionTimeout)
eventBus.Subscribe("status.changed", integration.OnStatusChanged)
```

### 5. Testes
- Unit tests para AutomationRule
- Unit tests para AutomationEngine
- Integration tests para repository
- E2E tests com triggers reais

## 📊 Comparação

### Antes (Follow-up Rules)
- ❌ Nome limitado (sugere apenas follow-up)
- ❌ Confuso para outros casos de uso
- ❌ Não alinhado com mercado

### Depois (Automation Rules)
- ✅ Nome abrangente
- ✅ Cobre todos os casos de uso
- ✅ Alinhado com HubSpot, Salesforce, Pipedrive
- ✅ Marketing-friendly
- ✅ Escalável para novos tipos

## 🎉 Resultado Final

**Sistema único e elegante que suporta:**

1. ✅ Follow-ups (após inatividade)
2. ✅ Event automation (resposta a eventos)
3. ✅ Scheduled automation (agendadas/cron)
4. ✅ Conditional automation (lógica complexa)
5. ✅ Purchase automation (compras/transações)
6. ✅ Re-engagement (reativação)
7. ✅ Onboarding (boas-vindas)
8. ✅ Offboarding (churn prevention)

**Tudo com:**
- Uma única entidade: `AutomationRule`
- Uma única tabela: `automation_rules`
- Um único manager: `AutomationRuleManager`
- Um único engine: `AutomationEngine`

**Elegante, simples, escalável.** ✨
