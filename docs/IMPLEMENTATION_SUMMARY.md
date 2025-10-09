# Implementation Summary - AI Agent & Message Debouncer

## 📦 O que foi implementado

### 1. Message Debouncer System ✅

**Arquivos criados:**
- `infrastructure/messaging/message_debouncer_v2.go` - Debouncer estilo n8n
- `infrastructure/messaging/message_batch_processor.go` - Processor desacoplado
- `infrastructure/messaging/debouncer_integration.go` - Integração completa
- `infrastructure/messaging/debouncer_example.go` - Exemplos de uso
- `docs/MESSAGE_DEBOUNCER_USAGE.md` - Documentação completa

**Funcionalidades:**
- ✅ Push mensagens para Redis LIST
- ✅ Pull + Switch Logic (Nothing/Proceed/Wait)
- ✅ Retry loop com max_retries (10)
- ✅ Timeout baseado em timestamp da última mensagem (15s default)
- ✅ Processor desacoplado (opcional)
- ✅ Strategies: Concatenator, Validator, Enricher, Sender
- ✅ Integração com WAHA
- ✅ 3 modos de uso: Only/Simple/AI

**Fluxo:**
```
WAHA Event → DebouncerIntegration
           → MessageDebouncerV2.PushAndCheck()
           → Push to Redis LIST
           → Loop: Pull → Check Decision → Wait/Proceed
           → MessageBatchProcessor (opcional)
           → AI/Webhook/Custom (opcional)
```

### 2. Channel-Pipeline Integration ✅

**Arquivos modificados/criados:**
- `infrastructure/database/migrations/000018_add_channel_pipeline_association.up.sql`
- `infrastructure/persistence/entities/channel.go` - Adicionados campos
- `internal/domain/channel/channel.go` - Novos métodos
- `internal/domain/channel/events.go` - Novos eventos

**Funcionalidades:**
- ✅ Campo `pipeline_id` (opcional) no Channel
- ✅ Campo `default_session_timeout_minutes` no Channel
- ✅ Métodos: `AssociatePipeline()`, `DisassociatePipeline()`, `HasPipeline()`, `SetDefaultTimeout()`
- ✅ Eventos de domínio

### 3. Session Timeout Resolver ✅

**Arquivos criados:**
- `internal/application/session/session_timeout_resolver.go`

**Funcionalidades:**
- ✅ Hierarquia: Pipeline > Channel > System Default (30min)
- ✅ Método `ResolveForChannel()` - retorna timeout + pipelineID
- ✅ Método `ResolveWithDetails()` - retorna info detalhada + source
- ✅ Método `ResolveWithFallback()` - fallback customizado
- ✅ Fallback graceful em caso de erro

**Hierarquia de Resolução:**
```
1. Pipeline.SessionTimeoutMinutes (se Channel.PipelineID != nil)
2. Channel.DefaultSessionTimeoutMinutes (se > 0)
3. Sistema Default: 30 minutos
```

### 4. Follow-up Rules System ✅

**Arquivos criados - Domain Layer:**
- `internal/domain/pipeline/automation_rule.go` - Domain model completo
- `internal/domain/pipeline/scheduled_rule.go` - Regras agendadas/recorrentes
- `internal/domain/pipeline/events.go` - Eventos de Follow-up Rules (updated)

**Arquivos criados - Application Layer:**
- `internal/application/pipeline/automation_engine.go` - Engine de execução
- `internal/application/pipeline/automation_action_executor.go` - Executor de ações
- `internal/application/pipeline/automation_rule_manager.go` - Gerenciador CRUD sofisticado
- `internal/application/pipeline/automation_integration.go` - Integração com eventos

**Arquivos criados - Infrastructure Layer:**
- `infrastructure/persistence/entities/automation_rule.go` - Entity com schedule
- `infrastructure/persistence/gorm_automation_rule_repository.go` - Repository
- `infrastructure/database/migrations/000019_create_automation_rules_table.up.sql`
- `infrastructure/database/migrations/000019_create_automation_rules_table.down.sql`
- `infrastructure/workflow/scheduled_rules_worker.go` - Worker para regras agendadas

**Funcionalidades:**
- ✅ **Triggers**: session.ended, session.timeout, no_response.timeout, status.changed, scheduled
- ✅ **Conditions**: Specification Pattern com operadores (eq, gt, lt, contains, in)
- ✅ **Actions**: 11 tipos (send_message, change_status, assign_agent, webhook, tags, etc)
- ✅ **Scheduled Rules**: once, daily, weekly, monthly, cron
- ✅ **AutomationEngine**: Avalia condições e executa ações
- ✅ **ActionExecutor**: Executa ações concretas com serviços externos
- ✅ **RuleManager**: CRUD + duplicate + export/import + reorder + statistics + test
- ✅ **ScheduledRulesWorker**: Background worker com poll interval configurável
- ✅ **Priority-based execution**: Ordem de execução configurável
- ✅ **Delayed actions**: Ações com delay em minutos
- ✅ **Graceful fallback**: Continua executando mesmo com erros

**Arquitetura:**
```
Domain Events → AutomationIntegration
              → AutomationEngine.EvaluateAndExecute()
              → Rule.EvaluateConditions() (Specification Pattern)
              → ActionExecutor.Execute() (Strategy Pattern)
              → External Services (MessageSender, WebhookSender, etc)

Scheduled Rules → ScheduledRulesWorker (Background)
                → Query: next_execution <= now AND enabled = true
                → AutomationEngine.EvaluateAndExecute()
                → Update: last_executed, next_execution
```

**Scheduled Types:**
- `once`: Executa uma vez em timestamp específico
- `daily`: Todos os dias em hour:minute
- `weekly`: Toda semana em day_of_week, hour:minute
- `monthly`: Todo mês em day_of_month, hour:minute
- `cron`: Expressão cron customizada (placeholder)

**RuleManager Features:**
- Create/Update/Delete rules
- Enable/Disable (single + bulk)
- Duplicate rule
- Export/Import JSON
- Reorder priorities
- Get statistics (total, enabled, by trigger, avg conditions/actions)
- Test conditions with mock context
- Schedule manual execution

### 5. Documentation ✅

**Arquivos criados:**
- `docs/AI_AGENT_DEBOUNCER_DESIGN.md` - Design completo da arquitetura
- `docs/MESSAGE_DEBOUNCER_USAGE.md` - Guia de uso detalhado
- `docs/FOLLOW_UP_RULES_SYSTEM.md` - Documentação completa de Follow-up Rules
- `docs/IMPLEMENTATION_SUMMARY.md` - Este arquivo

## 🔄 Próximos passos

### 6. AI Agent Integration Interface (Pendente)
- Interface genérica AIProvider
- Implementação OpenAI
- Implementação Anthropic
- AI Agent Coordinator
- Integração com MessageBatchProcessor

### 7. Testes (Pendente)
- Testes unitários do Debouncer
- Testes unitários do Follow-up Engine
- Testes do Timeout Resolver
- Testes de integração Follow-up Rules
- Testes E2E

### 8. Integrações Finais (Pendente)
- Conectar AutomationIntegration com domain event bus
- Integrar ScheduledRulesWorker no main.go
- Handler HTTP para gerenciar regras (CRUD API)
- Temporal workflows para delayed actions

## 📋 Como usar

### Follow-up Rules - Exemplos

#### Criar regra de follow-up 24h

```go
manager := pipeline.NewAutomationRuleManager(ruleRepo, pipelineRepo, nil, logger)

rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    PipelineID: pipelineID,
    TenantID:   "tenant-123",
    Name:       "Follow-up 24h inativo",
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
                "content": "Olá! Vi que você não respondeu. Posso ajudar?",
            },
        },
    },
    Enabled: true,
})
```

#### Criar regra agendada semanal

```go
schedule := pipeline.ScheduledRuleConfig{
    Type:      pipeline.ScheduleWeekly,
    DayOfWeek: ptrInt(1), // segunda-feira
    Hour:      10,
    Minute:    0,
}

rule, _ := manager.CreateRule(ctx, pipeline.CreateRuleInput{
    Name:     "Newsletter semanal",
    Trigger:  pipeline.TriggerScheduled,
    Schedule: &schedule,
    Actions: []pipeline.RuleAction{
        {
            Type: pipeline.ActionSendTemplate,
            Params: map[string]interface{}{
                "template_name": "newsletter",
            },
        },
    },
})
```

#### Integrar com eventos de sessão

```go
integration := pipeline.NewAutomationIntegration(
    engine,
    sessionRepo,
    pipelineRepo,
    logger,
)

// Quando sessão encerra
integration.OnSessionEnded(ctx, sessionID)

// Quando contato não responde
integration.OnNoResponse(ctx, sessionID, 24.0)
```

#### Iniciar Scheduled Worker

```go
worker := workflow.NewScheduledRulesWorker(
    db,
    followUpEngine,
    1*time.Minute,
    logger,
)

go worker.Start(ctx)
```

## 📋 Debouncer - Como usar

### Modo 1: Apenas Debouncing (Manual)

```go
redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
debouncer := messaging.NewMessageDebouncerV2(redisClient, 15*time.Second, nil)

// Push manual
debouncer.Push(ctx, sessionKey, msg)

// Pull manual quando quiser processar
messages, _ := debouncer.Pull(ctx, sessionKey)

// Processar como quiser
for _, m := range messages {
    // seu código
}
```

### Modo 2: Debouncing + Processamento Automático

```go
processor := messaging.NewMessageBatchProcessor(
    messaging.SimpleConcatenator{},
    messaging.NoopValidator{},
    nil, // sem enricher
    nil, // sem sender (processar internamente)
)

integration := messaging.NewDebouncerIntegration(redisClient, processor)

// Processa automaticamente: push + loop + flush + process
integration.ProcessWAHAMessage(ctx, wahaEvent)
```

### Modo 3: Debouncing + AI (OpenAI/Anthropic)

```go
// Implementa sender customizado
type OpenAISender struct {
    client *openai.Client
}

func (s *OpenAISender) Send(ctx context.Context, sessionKey string, content string, metadata interface{}) error {
    // Chama OpenAI
    resp, _ := s.client.CreateChatCompletion(...)

    // Envia resposta ao contato
    // ...

    return nil
}

// Cria processor com AI
aiSender := &OpenAISender{client: openai.NewClient(apiKey)}
processor := messaging.NewMessageBatchProcessor(
    messaging.MediaAwareConcatenator{}, // detecta mídia
    messaging.MinMessageValidator{MinCount: 1},
    nil,
    aiSender, // envia para IA
)

integration := messaging.NewDebouncerIntegration(redisClient, processor)
integration.ProcessWAHAMessage(ctx, wahaEvent)
```

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────────────────────────┐
│                    Camadas da Solução                        │
└─────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│  Application Layer                                           │
│  • DebouncerIntegration                                      │
│  • SessionTimeoutResolver                                    │
│  • AIAgentCoordinator (TODO)                                │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│  Domain Layer                                                │
│  • Channel (com PipelineID)                                  │
│  • Pipeline (com SessionTimeoutMinutes)                      │
│  • Session                                                   │
│  • AutomationRule (TODO)                                       │
└──────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│  Infrastructure Layer                                        │
│  • MessageDebouncerV2 (Redis LIST)                          │
│  • MessageBatchProcessor (Strategy Pattern)                 │
│  • Repositories                                              │
│  • WAHA Integration                                          │
└──────────────────────────────────────────────────────────────┘
```

## 🔧 Configuração

### Variáveis de Ambiente

```bash
# Debouncer
DEBOUNCER_ENABLED=true
DEBOUNCER_WAIT_DURATION=15s
DEBOUNCER_MAX_RETRIES=10

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# AI (opcional)
AI_ENABLED=true
AI_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4

# Pipeline defaults
DEFAULT_SESSION_TIMEOUT_MINUTES=30
```

### Database Migrations

Executar migrations:
```bash
make migrate-up
```

Migration `000018` adiciona:
- `channels.pipeline_id` (UUID, nullable, FK para pipelines)
- `channels.default_session_timeout_minutes` (INT, default 30)

## 📊 Métricas Sugeridas

```
# Debouncer
debouncer_messages_pushed_total{session_key}
debouncer_messages_processed_total{session_key}
debouncer_batch_size{session_key}
debouncer_wait_duration_seconds{session_key}
debouncer_retries_total{session_key}

# Session Timeout
session_timeout_source{source="pipeline|channel|default"}
session_created_with_pipeline_total
session_created_without_pipeline_total

# AI Agent (TODO)
ai_agent_requests_total{provider, model}
ai_agent_response_time_seconds{provider}
ai_agent_tokens_used{provider, model}
ai_agent_errors_total{provider, error_type}
```

## 🎯 Status das Implementações

### Concluído ✅
1. ✅ **Message Debouncer** - Sistema completo com Redis LIST, n8n flow, 3 modos
2. ✅ **Session Timeout Resolver** - Hierarquia Pipeline > Channel > Default
3. ✅ **Channel-Pipeline Association** - Campo opcional + eventos
4. ✅ **Follow-up Rules System** - Engine + Manager + Scheduled Worker + 11 actions
5. ✅ **Scheduled Rules** - once/daily/weekly/monthly/cron com worker background
6. ✅ **Rule Manager** - CRUD completo + duplicate + export/import + stats
7. ✅ **Documentation** - 3 docs completos (Design, Usage, System)

### Em Andamento 🔄
- Correção de erros de compilação
- Ajustes de tipos e imports

### Pendente ⏳
8. ⏳ **AI Agent Integration Interface** - OpenAI/Anthropic providers
9. ⏳ **HTTP Handlers** - API REST para gerenciar regras
10. ⏳ **Event Bus Integration** - Conectar eventos de domínio
11. ⏳ **Temporal Integration** - Delayed actions workflows
12. ⏳ **Testes** - Unit + Integration + E2E
13. ⏳ **Métricas** - Prometheus metrics
14. ⏳ **Swagger Documentation** - OpenAPI specs
