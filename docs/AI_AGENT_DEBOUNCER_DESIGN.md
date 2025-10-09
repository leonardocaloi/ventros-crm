# Design: AI Agent Message Debouncer & Pipeline Integration

## Visão Geral

Este documento descreve a arquitetura para implementar:

1. **Message Debouncer com Redis**: Sistema que agrupa mensagens rápidas e fragmentadas
2. **Pipeline-Channel Integration**: Relacionamento opcional entre Channel e Pipeline
3. **Session Timeout Hierarchy**: Timeout do Pipeline sobrescreve o do Channel
4. **Follow-up Rules**: Regras automáticas de follow-up no Pipeline
5. **AI Agent Integration**: Interface para agentes externos de IA

## 1. Message Debouncer Architecture

### 1.1 Problema

Mensagens chegam muito rápido e fragmentadas (ex: usuário digitando em partes). Precisamos:
- Agrupar mensagens por sessão/contato
- Manter ordem temporal
- Processar ou marcar como "processadas" (debate arquitetural)
- Concatenar em texto ou JSON para envio ao agente IA

### 1.2 Solução: Redis-based Debouncer

```
┌─────────────────────────────────────────────────────────────────┐
│                      Message Flow                                │
└─────────────────────────────────────────────────────────────────┘

Webhook/Queue → WAHAMessageConsumer → MessageDebouncer (Redis)
                                            │
                                            ▼
                                    [Buffer Window: 2s]
                                            │
                                            ▼
                                    Concatenate & Process
                                            │
                                            ▼
                                    AI Agent Interface
```

#### Estrutura Redis

**Chave**: `debouncer:session:{session_id}:messages`

**Estrutura**: Sorted Set (ZADD) com score = timestamp

```json
{
  "score": 1697123456789,
  "value": {
    "message_id": "uuid",
    "text": "parte da mensagem",
    "type": "text|audio|image|document",
    "timestamp": 1697123456789,
    "from_contact": true,
    "metadata": {}
  }
}
```

**TTL**: 5 minutos (após última mensagem)

#### Componentes

```go
// infrastructure/messaging/message_debouncer.go

type MessageDebouncer struct {
    redis           *redis.Client
    windowDuration  time.Duration // 2s default
    flushCallback   func(sessionID uuid.UUID, messages []DebouncedMessage) error
}

type DebouncedMessage struct {
    MessageID   uuid.UUID              `json:"message_id"`
    Text        string                 `json:"text"`
    Type        string                 `json:"type"`
    Timestamp   int64                  `json:"timestamp"`
    FromContact bool                   `json:"from_contact"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Adiciona mensagem ao buffer
func (d *MessageDebouncer) AddMessage(ctx context.Context, sessionID uuid.UUID, msg DebouncedMessage) error

// Verifica se deve fazer flush (chamado por timer ou threshold)
func (d *MessageDebouncer) CheckAndFlush(ctx context.Context, sessionID uuid.UUID) error

// Worker que monitora buffers e faz flush periódico
func (d *MessageDebouncer) StartFlushWorker(ctx context.Context)
```

#### Estratégias de Flush

1. **Time-based**: Após 2s sem novas mensagens
2. **Count-based**: Após 10 mensagens no buffer
3. **Manual**: API endpoint para forçar flush

#### Formato de Saída

**Opção 1: Concatenação Simples**
```
Mensagem 1
Mensagem 2
Mensagem 3
```

**Opção 2: JSON Estruturado** (Recomendado)
```json
{
  "session_id": "uuid",
  "messages": [
    {"text": "...", "timestamp": 123, "type": "text"},
    {"text": "...", "timestamp": 124, "type": "text"}
  ],
  "session_context": {
    "contact": {...},
    "pipeline": {...},
    "custom_fields": {...},
    "session_summary": "..."
  }
}
```

## 2. Channel-Pipeline Integration

### 2.1 Current State

- **Channel**: Independente, tem timeout padrão
- **Session**: Pode ou não ter Pipeline
- **Pipeline**: Define timeout e fluxo

### 2.2 New Relationship

```
Channel (opcional) ──┐
                     ├──> Session ──> Pipeline (obrigatório na prática)
Pipeline ────────────┘
```

### 2.3 Schema Changes

#### ChannelEntity

```go
// infrastructure/persistence/entities/channel.go

type ChannelEntity struct {
    // ... campos existentes ...

    // Novo campo: Pipeline associado (opcional)
    PipelineID *uuid.UUID `gorm:"type:uuid;index"`

    // Timeout padrão do canal (em minutos)
    DefaultSessionTimeoutMinutes int `gorm:"default:30"`

    // Relacionamento
    Pipeline *PipelineEntity `gorm:"foreignKey:PipelineID"`
}
```

#### Channel Domain

```go
// internal/domain/channel/channel.go

type Channel struct {
    // ... campos existentes ...

    PipelineID                   *uuid.UUID
    DefaultSessionTimeoutMinutes int // 30min default
}

func (c *Channel) AssociatePipeline(pipelineID uuid.UUID) error
func (c *Channel) DisassociatePipeline() error
func (c *Channel) GetEffectiveTimeout() int // Retorna timeout do pipeline ou do canal
```

### 2.4 Timeout Resolution Logic

**Hierarquia de Timeout**:

```
1. Pipeline.SessionTimeoutMinutes (se Channel.PipelineID != nil)
2. Channel.DefaultSessionTimeoutMinutes
3. Sistema Default: 30 minutos
```

**Implementação**:

```go
// internal/application/session/session_timeout_resolver.go

type SessionTimeoutResolver struct {
    channelRepo  channel.Repository
    pipelineRepo pipeline.Repository
}

func (r *SessionTimeoutResolver) ResolveTimeout(
    channelID uuid.UUID,
) (time.Duration, error) {
    channel, err := r.channelRepo.GetByID(channelID)
    if err != nil {
        return 30 * time.Minute, nil // fallback
    }

    // Se tem pipeline associado, usa timeout do pipeline
    if channel.PipelineID != nil {
        pipeline, err := r.pipelineRepo.GetByID(*channel.PipelineID)
        if err == nil && pipeline != nil {
            return time.Duration(pipeline.SessionTimeoutMinutes()) * time.Minute, nil
        }
    }

    // Senão, usa timeout do canal
    if channel.DefaultSessionTimeoutMinutes > 0 {
        return time.Duration(channel.DefaultSessionTimeoutMinutes) * time.Minute, nil
    }

    // Fallback sistema
    return 30 * time.Minute, nil
}
```

## 3. Follow-up Rules System

### 3.1 Conceito

Follow-ups são ações automáticas disparadas por eventos de sessão. Exemplos:
- Enviar mensagem após X tempo de inatividade
- Enviar pesquisa de satisfação após sessão encerrar
- Reagendar contato se não houver resposta em Y horas
- Escalar para humano se IA não conseguiu resolver

### 3.2 Architecture Pattern: Rules Engine

Baseado em **Specification Pattern** + **Chain of Responsibility**

```
┌────────────────────────────────────────────────────────┐
│                  Follow-up Rule                         │
├────────────────────────────────────────────────────────┤
│ • Trigger: SessionEnded / InactivityTimeout / etc      │
│ • Condition: After 24h / If unresolved / etc           │
│ • Action: SendMessage / ChangeStatus / AssignAgent     │
└────────────────────────────────────────────────────────┘
```

### 3.3 Domain Model

```go
// internal/domain/pipeline/automation_rule.go

type AutomationTrigger string

const (
    TriggerSessionEnded      AutomationTrigger = "session.ended"
    TriggerInactivityTimeout AutomationTrigger = "inactivity.timeout"
    TriggerUnresolved        AutomationTrigger = "session.unresolved"
    TriggerNoResponse        AutomationTrigger = "no_response.timeout"
)

type AutomationAction string

const (
    ActionSendMessage    AutomationAction = "send_message"
    ActionChangeStatus   AutomationAction = "change_pipeline_status"
    ActionAssignAgent    AutomationAction = "assign_agent"
    ActionCreateTask     AutomationAction = "create_task"
    ActionSendWebhook    AutomationAction = "send_webhook"
)

type AutomationRule struct {
    id         uuid.UUID
    pipelineID uuid.UUID
    name       string
    trigger    AutomationTrigger
    conditions []RuleCondition // Ex: after_minutes, if_field_equals
    actions    []RuleAction
    priority   int  // ordem de execução
    enabled    bool
    createdAt  time.Time
    updatedAt  time.Time
}

type RuleCondition struct {
    Field    string      `json:"field"`     // ex: "minutes_since_last_message"
    Operator string      `json:"operator"`  // eq, gt, lt, contains
    Value    interface{} `json:"value"`
}

type RuleAction struct {
    Type   AutomationAction         `json:"type"`
    Params map[string]interface{} `json:"params"`
}

// Exemplo de regra
rule := AutomationRule{
    Name:    "Send satisfaction survey",
    Trigger: TriggerSessionEnded,
    Conditions: []RuleCondition{
        {Field: "resolved", Operator: "eq", Value: true},
        {Field: "message_count", Operator: "gt", Value: 3},
    },
    Actions: []RuleAction{
        {
            Type: ActionSendMessage,
            Params: map[string]interface{}{
                "delay_minutes": 5,
                "template": "Como foi sua experiência? 😊",
            },
        },
    },
}
```

### 3.4 Entities

```go
// infrastructure/persistence/entities/automation_rule.go

type AutomationRuleEntity struct {
    ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    PipelineID uuid.UUID      `gorm:"type:uuid;not null;index"`
    TenantID   string         `gorm:"not null;index"`
    Name       string         `gorm:"not null"`
    Trigger    string         `gorm:"not null;index"`
    Conditions datatypes.JSON `gorm:"type:jsonb"` // Array de RuleCondition
    Actions    datatypes.JSON `gorm:"type:jsonb"` // Array de RuleAction
    Priority   int            `gorm:"default:0"`
    Enabled    bool           `gorm:"default:true;index"`
    CreatedAt  time.Time      `gorm:"autoCreateTime"`
    UpdatedAt  time.Time      `gorm:"autoUpdateTime"`

    Pipeline PipelineEntity `gorm:"foreignKey:PipelineID"`
}
```

### 3.5 Execution Engine

```go
// internal/application/pipeline/automation_engine.go

type AutomationEngine struct {
    ruleRepo           AutomationRuleRepository
    messageService     MessageService
    contactService     ContactService
    webhookDispatcher  WebhookDispatcher
}

// Avalia e executa regras para um evento de sessão
func (e *AutomationEngine) ProcessSessionEvent(
    ctx context.Context,
    event session.DomainEvent,
) error {
    // 1. Busca regras do pipeline
    rules, err := e.ruleRepo.FindByPipelineAndTrigger(
        session.PipelineID(),
        mapEventToTrigger(event),
    )

    // 2. Filtra regras que passam nas condições
    applicableRules := e.evaluateConditions(rules, event)

    // 3. Ordena por prioridade
    sort.Slice(applicableRules, func(i, j int) bool {
        return applicableRules[i].Priority() < applicableRules[j].Priority()
    })

    // 4. Executa ações
    for _, rule := range applicableRules {
        if err := e.executeActions(ctx, rule, event); err != nil {
            log.Errorf("Failed to execute rule %s: %v", rule.Name(), err)
        }
    }

    return nil
}

func (e *AutomationEngine) evaluateConditions(
    rules []*AutomationRule,
    event session.DomainEvent,
) []*AutomationRule {
    // Implementa lógica de avaliação de condições
}

func (e *AutomationEngine) executeActions(
    ctx context.Context,
    rule *AutomationRule,
    event session.DomainEvent,
) error {
    for _, action := range rule.Actions() {
        switch action.Type {
        case ActionSendMessage:
            // Envia mensagem com delay (via worker/scheduler)
        case ActionChangeStatus:
            // Atualiza status do pipeline do contato
        case ActionAssignAgent:
            // Atribui agente específico
        }
    }
    return nil
}
```

### 3.6 Temporal Workflows Integration

Para ações com delay (ex: "enviar após 24h"), usar Temporal:

```go
// internal/workflows/automation/automation_workflow.go

func AutomationWorkflow(ctx workflow.Context, req AutomationRequest) error {
    // Aguarda delay
    if req.DelayMinutes > 0 {
        workflow.Sleep(ctx, time.Duration(req.DelayMinutes) * time.Minute)
    }

    // Executa ação
    return workflow.ExecuteActivity(ctx, ExecuteAutomationAction, req).Get(ctx, nil)
}
```

## 4. AI Agent Integration Interface

### 4.1 Objetivo

Interface genérica para integrar com múltiplos provedores de IA (OpenAI, Anthropic, etc).

### 4.2 Architecture

```
Session → AIAgentCoordinator → AIAgentProvider (interface)
                                      ↓
                        ┌─────────────┴─────────────┐
                        ▼                           ▼
                  OpenAIProvider          AnthropicProvider
```

### 4.3 Domain Interface

```go
// internal/domain/agent/ai_provider.go

type AIProvider interface {
    // Envia mensagens concatenadas e contexto para IA
    SendMessage(ctx context.Context, req AIMessageRequest) (*AIMessageResponse, error)

    // Verifica health do provider
    HealthCheck(ctx context.Context) error
}

type AIMessageRequest struct {
    SessionID      uuid.UUID              `json:"session_id"`
    Messages       []ConversationMessage  `json:"messages"` // histórico
    Context        AIContext              `json:"context"`
    Config         AIConfig               `json:"config"`
}

type ConversationMessage struct {
    Role      string    `json:"role"` // user, assistant, system
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}

type AIContext struct {
    ContactInfo       ContactSummary         `json:"contact"`
    SessionSummary    string                 `json:"session_summary,omitempty"`
    PipelineInfo      PipelineSummary        `json:"pipeline"`
    CustomFields      map[string]interface{} `json:"custom_fields"`
    PreviousSessions  []SessionSummary       `json:"previous_sessions,omitempty"`
}

type AIConfig struct {
    Model       string  `json:"model"`       // gpt-4, claude-3, etc
    Temperature float64 `json:"temperature"`
    MaxTokens   int     `json:"max_tokens"`
    SystemPrompt string `json:"system_prompt,omitempty"`
}

type AIMessageResponse struct {
    Content     string                 `json:"content"`
    Metadata    map[string]interface{} `json:"metadata"`
    TokensUsed  int                    `json:"tokens_used"`
    ProcessedAt time.Time              `json:"processed_at"`
}
```

### 4.4 Implementation Example: OpenAI

```go
// infrastructure/ai/openai_provider.go

type OpenAIProvider struct {
    client *openai.Client
    apiKey string
}

func (p *OpenAIProvider) SendMessage(
    ctx context.Context,
    req AIMessageRequest,
) (*AIMessageResponse, error) {
    // Converte para formato OpenAI
    messages := make([]openai.ChatCompletionMessage, 0, len(req.Messages)+1)

    // System prompt com contexto
    systemPrompt := p.buildSystemPrompt(req.Context, req.Config)
    messages = append(messages, openai.ChatCompletionMessage{
        Role:    "system",
        Content: systemPrompt,
    })

    // Adiciona histórico de conversas
    for _, msg := range req.Messages {
        messages = append(messages, openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }

    // Chama API OpenAI
    resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:       req.Config.Model,
        Messages:    messages,
        Temperature: float32(req.Config.Temperature),
        MaxTokens:   req.Config.MaxTokens,
    })
    if err != nil {
        return nil, err
    }

    return &AIMessageResponse{
        Content:     resp.Choices[0].Message.Content,
        TokensUsed:  resp.Usage.TotalTokens,
        ProcessedAt: time.Now(),
    }, nil
}

func (p *OpenAIProvider) buildSystemPrompt(
    ctx AIContext,
    cfg AIConfig,
) string {
    var b strings.Builder

    if cfg.SystemPrompt != "" {
        b.WriteString(cfg.SystemPrompt)
        b.WriteString("\n\n")
    }

    b.WriteString("Contexto da Conversa:\n")
    b.WriteString(fmt.Sprintf("- Contato: %s\n", ctx.ContactInfo.Name))
    b.WriteString(fmt.Sprintf("- Pipeline: %s\n", ctx.PipelineInfo.Name))

    if ctx.SessionSummary != "" {
        b.WriteString(fmt.Sprintf("- Resumo: %s\n", ctx.SessionSummary))
    }

    return b.String()
}
```

### 4.5 AI Agent Coordinator

```go
// internal/application/agent/ai_agent_coordinator.go

type AIAgentCoordinator struct {
    debouncer      *MessageDebouncer
    providers      map[string]AIProvider // "openai", "anthropic", etc
    sessionRepo    session.Repository
    contactRepo    contact.Repository
    messageService MessageService
}

// Processa mensagens debouncadas e envia para IA
func (c *AIAgentCoordinator) ProcessDebouncedMessages(
    ctx context.Context,
    sessionID uuid.UUID,
    messages []DebouncedMessage,
) error {
    // 1. Carrega contexto da sessão
    sess, err := c.sessionRepo.GetByID(ctx, sessionID)
    if err != nil {
        return err
    }

    // 2. Verifica se sessão tem agente IA configurado
    if !c.shouldProcessWithAI(sess) {
        return nil
    }

    // 3. Monta request para IA
    aiReq := c.buildAIRequest(ctx, sess, messages)

    // 4. Seleciona provider
    provider := c.selectProvider(sess)

    // 5. Envia para IA
    resp, err := provider.SendMessage(ctx, aiReq)
    if err != nil {
        return err
    }

    // 6. Envia resposta ao contato
    return c.messageService.SendMessage(ctx, SendMessageRequest{
        SessionID: sessionID,
        Content:   resp.Content,
        FromAgent: true,
    })
}

func (c *AIAgentCoordinator) buildAIRequest(
    ctx context.Context,
    sess *session.Session,
    messages []DebouncedMessage,
) AIMessageRequest {
    // Monta contexto completo com histórico, custom fields, etc
    // ...
}
```

## 5. Pipeline Entity Updates

```go
// internal/domain/pipeline/pipeline.go

type Pipeline struct {
    // ... campos existentes ...

    // AI Features
    enableAISummary bool
    aiProvider      *string // "openai", "anthropic"
    aiModel         *string // "gpt-4", "claude-3"
    aiSystemPrompt  *string

    // Follow-up Rules (lazy loaded)
    followUpRules []*AutomationRule

    // Auto-message on session start
    welcomeMessage      *string
    welcomeMessageDelay int // segundos
}

// Métodos
func (p *Pipeline) SetAIAgent(provider, model string, systemPrompt *string) error
func (p *Pipeline) DisableAIAgent()
func (p *Pipeline) AddAutomationRule(rule *AutomationRule) error
func (p *Pipeline) RemoveAutomationRule(ruleID uuid.UUID) error
func (p *Pipeline) GetAutomationRules() []*AutomationRule
func (p *Pipeline) SetWelcomeMessage(message string, delaySeconds int)
```

## 6. Migration Plan

### 6.1 Database Migrations

```sql
-- 000018_add_channel_pipeline_association.up.sql
ALTER TABLE channels
ADD COLUMN pipeline_id UUID REFERENCES pipelines(id) ON DELETE SET NULL,
ADD COLUMN default_session_timeout_minutes INT DEFAULT 30;

CREATE INDEX idx_channels_pipeline_id ON channels(pipeline_id);

-- 000019_add_pipeline_ai_config.up.sql
ALTER TABLE pipelines
ADD COLUMN ai_provider VARCHAR(50),
ADD COLUMN ai_model VARCHAR(100),
ADD COLUMN ai_system_prompt TEXT,
ADD COLUMN welcome_message TEXT,
ADD COLUMN welcome_message_delay INT DEFAULT 0;

-- 000020_create_automation_rules_table.up.sql
CREATE TABLE automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    trigger VARCHAR(50) NOT NULL,
    conditions JSONB,
    actions JSONB NOT NULL,
    priority INT DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_automation_rules_pipeline_id ON automation_rules(pipeline_id);
CREATE INDEX idx_automation_rules_trigger ON automation_rules(trigger);
CREATE INDEX idx_automation_rules_enabled ON automation_rules(enabled);
```

### 6.2 Implementation Order

1. ✅ **Fase 1**: Message Debouncer (standalone, testável)
2. ✅ **Fase 2**: Channel-Pipeline association + Timeout resolver
3. ✅ **Fase 3**: Follow-up Rules domain model + basic engine
4. ✅ **Fase 4**: AI Agent Integration interface + OpenAI provider
5. ✅ **Fase 5**: AI Agent Coordinator (integra debouncer + AI)
6. ✅ **Fase 6**: Follow-up rules com Temporal workflows
7. ✅ **Fase 7**: Testes E2E + documentação

## 7. Configuration Examples

### 7.1 Channel with Pipeline

```json
{
  "name": "WhatsApp Vendas",
  "type": "waha",
  "pipeline_id": "pipeline-uuid",
  "default_session_timeout_minutes": 30,
  "config": {
    "session_id": "vendas-001"
  }
}
```

### 7.2 Pipeline with AI Agent

```json
{
  "name": "Atendimento Automatizado",
  "session_timeout_minutes": 45,
  "ai_provider": "openai",
  "ai_model": "gpt-4",
  "ai_system_prompt": "Você é um assistente de vendas...",
  "welcome_message": "Olá! Como posso ajudar?",
  "welcome_message_delay": 2
}
```

### 7.3 Follow-up Rule Example

```json
{
  "name": "Pesquisa de satisfação",
  "trigger": "session.ended",
  "conditions": [
    {"field": "resolved", "operator": "eq", "value": true},
    {"field": "message_count", "operator": "gt", "value": 3}
  ],
  "actions": [
    {
      "type": "send_message",
      "params": {
        "delay_minutes": 5,
        "template": "Como foi sua experiência? 😊\n1 - Péssimo\n2 - Ruim\n3 - Bom\n4 - Ótimo\n5 - Excelente"
      }
    }
  ],
  "priority": 1,
  "enabled": true
}
```

## 8. Testing Strategy

### 8.1 Unit Tests

- MessageDebouncer: test buffering, flushing
- AutomationEngine: test condition evaluation
- AIProviders: mock API calls

### 8.2 Integration Tests

- Channel-Pipeline-Session timeout resolution
- End-to-end message flow com debouncer
- Follow-up rule execution

### 8.3 E2E Tests

- Enviar mensagens rápidas → verificar debouncing → verificar resposta IA
- Criar sessão com pipeline → verificar timeout correto
- Disparar evento → verificar follow-up executado

## 9. Monitoring & Observability

### Metrics

- `debouncer.messages.buffered`
- `debouncer.flush.duration`
- `ai_agent.response.time`
- `ai_agent.tokens.used`
- `automation.rules.executed`
- `automation.actions.failed`

### Logs

```
[Debouncer] Buffered message session=XXX count=3
[Debouncer] Flushing session=XXX messages=5 age=2.1s
[AI Agent] Sent to OpenAI session=XXX tokens=1234
[Follow-up] Executed rule "Send survey" session=XXX
```

## 10. Next Steps

1. Revisar este design com time
2. Criar PRs incrementais (uma fase por vez)
3. Documentar APIs públicas (Swagger)
4. Criar exemplos de uso (cookbook)
5. Monitorar métricas em produção
