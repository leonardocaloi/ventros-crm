# Proposta de Refatoração: Automation Domain

## 🚨 Problema Identificado

Atualmente, `AutomationRule` está **acoplado ao Pipeline** (`pipeline/automation_rule.go`), mas isso cria limitações:

### Problemas Atuais:
1. **Acoplamento incorreto:** Automation só funciona no contexto de Pipeline
2. **Falta de flexibilidade:** Não permite automações globais (diárias, reports, etc)
3. **Violação de DDD:** Automation deveria ser seu próprio Bounded Context

### Exemplo Real do Problema:
```
❌ Cenário Atual (limitado):
- Automação APENAS quando status do pipeline muda
- Automação APENAS dentro de um pipeline específico

✅ Cenário Desejado:
- Todo dia às 18h: gerar relatório de performance de agentes
- Comparar agentes, ranqueá-los
- Enviar via webhook para coordenador
- Notificar top 3 agentes

^ Isso NÃO depende de pipeline!
```

---

## 🎯 Solução Proposta: Novo Bounded Context "Automation"

### Nova Estrutura DDD

```
internal/domain/
├── automation/                    # 🆕 NOVO BOUNDED CONTEXT
│   ├── automation.go             # Aggregate Root principal
│   ├── trigger.go                # Value Objects de triggers
│   ├── condition.go              # Value Objects de condições
│   ├── action.go                 # Value Objects de ações
│   ├── schedule.go               # Scheduling logic
│   ├── execution_history.go     # Tracking de execuções
│   ├── events.go                 # Domain Events
│   └── repository.go             # Repository interface
│
├── pipeline/
│   ├── pipeline.go
│   ├── status.go
│   └── pipeline_automation.go    # 🆕 Especialização para pipeline
│
└── agent/
    └── agent_automation.go        # 🆕 Especialização para agentes
```

---

## 📐 Design Proposto

### 1. Automation Aggregate Root (Universal)

```go
package automation

// Automation é o Aggregate Root UNIVERSAL para qualquer tipo de automação
type Automation struct {
    id          uuid.UUID
    tenantID    string
    projectID   uuid.UUID
    name        string
    description string

    // Tipo de automação define o domínio
    automationType AutomationType

    // Trigger: quando executar
    trigger Trigger

    // Conditions: verificações antes de executar
    conditions ConditionGroup

    // Actions: o que fazer
    actions []Action

    // Scheduling (opcional)
    schedule *Schedule

    // Metadata
    enabled   bool
    priority  int
    createdAt time.Time
    updatedAt time.Time

    // Execution tracking
    lastExecutedAt *time.Time
    executionCount int

    events []DomainEvent
}

// AutomationType define o escopo da automação
type AutomationType string

const (
    // Pipeline-scoped automations
    AutomationTypePipeline AutomationType = "pipeline"

    // Global automations (não dependem de entidade específica)
    AutomationTypeScheduled AutomationType = "scheduled"
    AutomationTypeReport    AutomationType = "report"

    // Agent-scoped automations
    AutomationTypeAgentPerformance AutomationType = "agent_performance"
    AutomationTypeAgentRanking     AutomationType = "agent_ranking"

    // Contact-scoped automations
    AutomationTypeContactFollowUp  AutomationType = "contact_followup"
    AutomationTypeContactSegment   AutomationType = "contact_segment"

    // Custom
    AutomationTypeCustom AutomationType = "custom"
)
```

### 2. Trigger System (Desacoplado)

```go
// Trigger define QUANDO a automação deve executar
type Trigger struct {
    Type      TriggerType
    EventType string           // ex: "contact.created", "session.ended"
    Schedule  *ScheduleConfig  // para triggers temporais
    Filters   map[string]interface{} // filtros adicionais
}

type TriggerType string

const (
    // Event-based (reativo)
    TriggerTypeEvent TriggerType = "event"

    // Time-based (proativo)
    TriggerTypeSchedule TriggerType = "schedule"
    TriggerTypeCron     TriggerType = "cron"

    // Condition-based (monitoramento)
    TriggerTypeCondition TriggerType = "condition"

    // Manual
    TriggerTypeManual TriggerType = "manual"
)

// Exemplos de Eventos que podem disparar automações
const (
    // Pipeline Events
    EventPipelineStatusChanged = "pipeline.status_changed"
    EventPipelineStageComplete = "pipeline.stage_completed"

    // Session Events
    EventSessionEnded   = "session.ended"
    EventSessionTimeout = "session.timeout"

    // Contact Events
    EventContactCreated = "contact.created"
    EventContactUpdated = "contact.updated"

    // Agent Events
    EventAgentPerformanceChanged = "agent.performance_changed"
    EventAgentShiftEnded        = "agent.shift_ended"

    // Time Events
    EventDailyReport   = "time.daily_report"
    EventWeeklyReview  = "time.weekly_review"
    EventMonthlyClose  = "time.monthly_close"
)
```

### 3. Condition System (Flexível)

```go
// ConditionGroup permite composição complexa
type ConditionGroup struct {
    Logic      LogicOperator  // AND / OR
    Conditions []Condition
    Groups     []ConditionGroup // nested
}

type Condition struct {
    Field    string      // campo a verificar
    Operator Operator    // comparação
    Value    interface{} // valor esperado
    Context  string      // contexto: "contact", "agent", "session", etc
}

// Exemplos de campos por contexto:
// Context: "contact"
//   - "message_count"
//   - "last_interaction_hours_ago"
//   - "pipeline_status"
//
// Context: "agent"
//   - "performance_score"
//   - "handled_sessions_today"
//   - "average_response_time"
//
// Context: "time"
//   - "hour_of_day"
//   - "day_of_week"
//   - "day_of_month"
```

### 4. Action System (Extensível)

```go
// Action define O QUE fazer quando automation executa
type Action struct {
    Type     ActionType
    Params   map[string]interface{}
    Delay    time.Duration // delay antes de executar
    Priority int           // ordem de execução
}

type ActionType string

const (
    // Messaging
    ActionSendMessage   ActionType = "send_message"
    ActionSendTemplate  ActionType = "send_template"
    ActionSendEmail     ActionType = "send_email"

    // Pipeline
    ActionChangePipelineStatus ActionType = "change_pipeline_status"
    ActionMoveToStage          ActionType = "move_to_stage"

    // Assignment
    ActionAssignAgent    ActionType = "assign_agent"
    ActionAssignToQueue  ActionType = "assign_to_queue"

    // Data
    ActionUpdateField    ActionType = "update_field"
    ActionAddTag         ActionType = "add_tag"
    ActionRemoveTag      ActionType = "remove_tag"

    // Reports & Analytics
    ActionGenerateReport ActionType = "generate_report"
    ActionCalculateMetrics ActionType = "calculate_metrics"
    ActionRankEntities   ActionType = "rank_entities"

    // Notifications
    ActionSendNotification ActionType = "send_notification"
    ActionSendWebhook      ActionType = "send_webhook"

    // Workflow
    ActionTriggerWorkflow ActionType = "trigger_workflow"
)
```

### 5. Schedule System (Para Automações Temporais)

```go
// Schedule define agendamento recorrente
type Schedule struct {
    Type       ScheduleType
    CronExpr   string
    Timezone   string
    StartDate  time.Time
    EndDate    *time.Time

    // Helpers
    Hour       int  // 0-23
    Minute     int  // 0-59
    DayOfWeek  *int // 0-6 (Sun-Sat)
    DayOfMonth *int // 1-31
}

type ScheduleType string

const (
    ScheduleOnce     ScheduleType = "once"
    ScheduleMinutely ScheduleType = "minutely"
    ScheduleHourly   ScheduleType = "hourly"
    ScheduleDaily    ScheduleType = "daily"
    ScheduleWeekly   ScheduleType = "weekly"
    ScheduleMonthly  ScheduleType = "monthly"
    ScheduleCron     ScheduleType = "cron"
)
```

---

## 🎬 Casos de Uso Suportados

### Caso 1: Pipeline Automation (Atual)
```go
// Automação específica de pipeline
automation := NewAutomation(
    tenantID,
    projectID,
    "Follow-up após 24h sem resposta",
    AutomationTypePipeline,
)

// Trigger: evento de pipeline
automation.SetTrigger(Trigger{
    Type:      TriggerTypeEvent,
    EventType: EventSessionEnded,
})

// Condição: verificar inatividade
automation.AddCondition(Condition{
    Context:  "session",
    Field:    "hours_since_last_message",
    Operator: OperatorGreaterThan,
    Value:    24,
})

// Ação: enviar mensagem
automation.AddAction(Action{
    Type: ActionSendTemplate,
    Params: map[string]interface{}{
        "template": "followup_24h",
    },
})
```

### Caso 2: Daily Agent Report (NOVO!)
```go
// Automação global agendada
automation := NewAutomation(
    tenantID,
    projectID,
    "Relatório Diário de Performance dos Agentes",
    AutomationTypeReport,
)

// Trigger: todo dia às 18h
automation.SetTrigger(Trigger{
    Type: TriggerTypeSchedule,
    Schedule: &ScheduleConfig{
        Type:     ScheduleDaily,
        Hour:     18,
        Minute:   0,
        Timezone: "America/Sao_Paulo",
    },
})

// Sem condições (sempre executa)

// Ações:
automation.AddAction(Action{
    Type: ActionCalculateMetrics,
    Params: map[string]interface{}{
        "entity": "agent",
        "metrics": []string{
            "sessions_handled",
            "avg_response_time",
            "satisfaction_score",
        },
        "period": "today",
    },
})

automation.AddAction(Action{
    Type: ActionRankEntities,
    Params: map[string]interface{}{
        "entity": "agent",
        "by": "performance_score",
        "limit": 10,
    },
})

automation.AddAction(Action{
    Type: ActionGenerateReport,
    Params: map[string]interface{}{
        "template": "agent_performance_daily",
        "format": "pdf",
    },
})

automation.AddAction(Action{
    Type: ActionSendWebhook,
    Params: map[string]interface{}{
        "url": "https://api.empresa.com/reports/agents",
        "include_report": true,
    },
})

automation.AddAction(Action{
    Type: ActionSendNotification,
    Params: map[string]interface{}{
        "to": "coordinator_role",
        "subject": "Relatório Diário - Performance dos Agentes",
        "body": "Veja o relatório em anexo",
    },
})
```

### Caso 3: Agent Performance Milestone (NOVO!)
```go
// Automação baseada em condição de agente
automation := NewAutomation(
    tenantID,
    projectID,
    "Notificar quando agente atinge 100 sessões",
    AutomationTypeAgentPerformance,
)

// Trigger: evento de agente
automation.SetTrigger(Trigger{
    Type:      TriggerTypeEvent,
    EventType: EventAgentPerformanceChanged,
})

// Condição: 100+ sessões hoje
automation.AddCondition(Condition{
    Context:  "agent",
    Field:    "sessions_handled_today",
    Operator: OperatorGreaterThanOrEqual,
    Value:    100,
})

// Ação: parabenizar agente
automation.AddAction(Action{
    Type: ActionSendNotification,
    Params: map[string]interface{}{
        "to": "agent",
        "message": "Parabéns! Você atingiu 100 sessões hoje! 🎉",
    },
})

automation.AddAction(Action{
    Type: ActionAddTag,
    Params: map[string]interface{}{
        "entity": "agent",
        "tag": "high_performer",
    },
})
```

---

## 📊 Comparação: Antes vs Depois

### Antes (Limitado)
```
Pipeline Domain
  └── AutomationRule (acoplado)
       ├── Trigger: apenas eventos de pipeline
       ├── Conditions: apenas dados de pipeline
       └── Actions: apenas ações de pipeline
```

### Depois (Flexível)
```
Automation Domain (Universal)
  ├── Automation (Aggregate Root)
  │    ├── Type: pipeline | scheduled | report | agent | contact | custom
  │    ├── Trigger: events | schedule | condition | manual
  │    ├── Conditions: qualquer contexto (contact, agent, session, time, etc)
  │    └── Actions: qualquer ação (messaging, data, reports, webhooks, etc)
  │
  ├── PipelineAutomation (Especialização)
  │    └── Helpers específicos para pipeline
  │
  ├── AgentAutomation (Especialização)
  │    └── Helpers específicos para agentes
  │
  └── ReportAutomation (Especialização)
       └── Helpers específicos para relatórios
```

---

## 🔄 Plano de Migração

### Fase 1: Criar Novo Domain
- [ ] Criar `internal/domain/automation/`
- [ ] Implementar Automation aggregate
- [ ] Implementar Trigger, Condition, Action value objects
- [ ] Implementar Schedule system
- [ ] Escrever testes completos

### Fase 2: Migrar Pipeline Automations
- [ ] Criar `PipelineAutomation` como especialização
- [ ] Migrar dados existentes
- [ ] Atualizar application layer
- [ ] Manter backward compatibility

### Fase 3: Implementar Novos Tipos
- [ ] Implementar `ScheduledAutomation`
- [ ] Implementar `AgentAutomation`
- [ ] Implementar `ReportAutomation`

### Fase 4: Temporal Integration
- [ ] Criar workflows para execução de automações
- [ ] Scheduled automation worker
- [ ] Condition monitoring worker
- [ ] Execution history tracking

### Fase 5: Deprecate Old
- [ ] Marcar `pipeline/automation_rule.go` como deprecated
- [ ] Remover após migração completa

---

## 🎯 Benefícios da Refatoração

### 1. Desacoplamento
- Automation não depende mais de Pipeline
- Cada bounded context tem responsabilidade clara

### 2. Flexibilidade
- Suporta automações globais (reports, scheduled tasks)
- Suporta automações de qualquer entidade (agent, contact, session)
- Triggers customizáveis

### 3. Extensibilidade
- Fácil adicionar novos tipos de automação
- Fácil adicionar novos triggers
- Fácil adicionar novas ações

### 4. Manutenibilidade
- Código mais limpo e organizado
- Testes mais fáceis (contextos isolados)
- Lógica de negócio mais clara

### 5. Casos de Uso Reais
- ✅ Relatórios diários automáticos
- ✅ Ranking de agentes
- ✅ Notificações baseadas em métricas
- ✅ Follow-ups inteligentes
- ✅ Webhooks para integrações externas

---

## 🚀 Próximos Passos

1. **Validar Proposta** com time
2. **Criar Spike** para testar conceito
3. **Implementar Fase 1** (novo domain)
4. **Migrar Gradualmente** (não big bang)
5. **Documentar Patterns** para extensões futuras

---

## 📝 Notas de Implementação

### Repository Pattern
```go
type AutomationRepository interface {
    Save(automation *Automation) error
    FindByID(id uuid.UUID) (*Automation, error)
    FindByType(automationType AutomationType) ([]*Automation, error)
    FindByTrigger(triggerType TriggerType) ([]*Automation, error)
    FindScheduledForExecution(now time.Time) ([]*Automation, error)
    FindByEntity(entityType string, entityID uuid.UUID) ([]*Automation, error)
    Delete(id uuid.UUID) error
}
```

### Application Service
```go
type AutomationService interface {
    // Criação
    CreateAutomation(dto CreateAutomationDTO) (*Automation, error)

    // Execução
    ExecuteAutomation(automationID uuid.UUID, context ExecutionContext) error
    EvaluateConditions(automationID uuid.UUID, context map[string]interface{}) (bool, error)

    // Scheduling
    ScheduleAutomation(automationID uuid.UUID) error
    ProcessScheduledAutomations() error

    // Event handling
    HandleDomainEvent(event DomainEvent) error
}
```

---

**Conclusão:** Esta refatoração transforma `AutomationRule` de uma feature acoplada a Pipeline em um **Bounded Context Universal** capaz de automatizar qualquer aspecto do sistema, seguindo princípios DDD e mantendo alta coesão e baixo acoplamento.

**Autor:** Time de Engenharia
**Data:** 2025-10-09
**Status:** 🟡 Proposta para Discussão
