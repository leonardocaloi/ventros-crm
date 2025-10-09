# Proposta de Renomeação: Follow-up Rules → Automation Rules

## 🤔 Problema

O nome **"Follow-up Rules"** é muito limitado e confuso:

### O que "Follow-up" sugere (incorretamente):
- ❌ Apenas mensagens de acompanhamento após inatividade
- ❌ Responder quando cliente não responde
- ❌ "Seguir" algo que já aconteceu

### O que o sistema REALMENTE faz:
- ✅ Automação de eventos (compra, status, mensagens)
- ✅ Ações imediatas em eventos (não apenas "follow-up")
- ✅ Workflows complexos (if-then-else logic)
- ✅ Agendamentos recorrentes (cron, diário, semanal)
- ✅ Mudanças de status automáticas
- ✅ Atribuição de agentes
- ✅ Webhooks
- ✅ Tags e campos customizados

## ❓ Exemplos que NÃO são "Follow-up"

### Exemplo 1: Mensagem Imediata após Compra
```
Trigger: status.changed (→ "Cliente")
Ação: send_message("Obrigado pela compra! Seu pedido #123 está confirmado.")
```
**Isso é:** Automação de evento de compra
**NÃO é:** Follow-up (não está "seguindo" nada, é imediato)

### Exemplo 2: Boas-vindas ao entrar no pipeline
```
Trigger: status.changed (→ "Lead Novo")
Ação: send_template("welcome_message")
```
**Isso é:** Onboarding automation
**NÃO é:** Follow-up

### Exemplo 3: Cliente volta para Lead
```
Trigger: status.changed (Cliente → Lead)
Condition: last_purchase_days > 90
Ação: send_message("Sentimos sua falta! Volta pra gente 💙")
```
**Isso é:** Re-engagement / Reativação
**Pode ser considerado:** Follow-up? Talvez... mas é confuso

### Exemplo 4: Newsletter Semanal
```
Trigger: scheduled (weekly, monday, 10:00)
Ação: send_template("newsletter")
```
**Isso é:** Scheduled automation / Campaign
**NÃO é:** Follow-up

## ✅ Proposta de Novo Nome

### **Opção 1: "Automation Rules" (Simples e claro)**

```
Pipeline
├─ Automation Rules (N)
│  ├─ "Mensagem de boas-vindas"
│  ├─ "Escalar após 5 mensagens"
│  ├─ "Newsletter semanal"
│  └─ "Recuperar carrinho abandonado"
```

**Vantagens:**
- ✅ Abrangente: cobre todos os casos de uso
- ✅ Intuitivo: todo mundo entende "automation"
- ✅ Simples: apenas 2 palavras

**Inglês:** Pipeline Automation Rules

### **Opção 2: "Workflow Rules"**

```
Pipeline
├─ Workflow Rules (N)
│  ├─ Rule: "Onboarding novo lead"
│  ├─ Rule: "Follow-up 24h"
│  └─ Rule: "Relatório mensal"
```

**Vantagens:**
- ✅ Conceito conhecido (Salesforce usa)
- ✅ Implica sequência de ações

**Desvantagens:**
- ⚠️ "Workflow" pode ser confundido com Temporal workflows

### **Opção 3: "Trigger Rules" ou "Event Rules"**

```
Pipeline
├─ Trigger Rules (N)
│  └─ Trigger: session.ended
│     └─ Rule: "Pedir feedback"
```

**Vantagens:**
- ✅ Foco no conceito central (triggers)

**Desvantagens:**
- ⚠️ Muito técnico para usuários finais

## 🎯 Recomendação Final

### **Nome Recomendado: "Automation Rules"**

**Por quê:**
1. Abrangente: cobre todos os casos de uso
2. Intuitivo: usuários entendem imediatamente
3. Marketing-friendly: "automatize seu pipeline!"
4. Flexível: permite evoluir o conceito
5. Usado por concorrentes: HubSpot, Pipedrive, etc.

## 📝 Mudanças Necessárias

### Domain Layer

**Antes:**
```go
// internal/domain/pipeline/automation_rule.go
type AutomationRule struct { ... }
type AutomationTrigger string
type AutomationAction string
```

**Depois:**
```go
// internal/domain/pipeline/automation_rule.go
type AutomationRule struct { ... }
type AutomationTrigger string
type AutomationAction string
```

### Application Layer

**Antes:**
```go
type AutomationEngine struct { ... }
type AutomationRuleManager struct { ... }
type AutomationIntegration struct { ... }
```

**Depois:**
```go
type AutomationEngine struct { ... }
type AutomationRuleManager struct { ... }
type AutomationIntegration struct { ... }
```

### Infrastructure Layer

**Antes:**
```sql
CREATE TABLE automation_rules (...)
```

**Depois:**
```sql
CREATE TABLE automation_rules (...)
-- ou manter compatibilidade:
CREATE TABLE pipeline_automation_rules (...)
```

### API Endpoints

**Antes:**
```
POST /api/v1/pipelines/{id}/automation-rules
GET  /api/v1/pipelines/{id}/automation-rules
```

**Depois:**
```
POST /api/v1/pipelines/{id}/automation-rules
GET  /api/v1/pipelines/{id}/automation-rules
-- ou simplesmente:
POST /api/v1/pipelines/{id}/rules
GET  /api/v1/pipelines/{id}/rules
```

### UI/Frontend

**Antes:**
- Menu: "Follow-up Rules"
- Botão: "Nova Regra de Follow-up"
- Título: "Gerenciar Follow-ups"

**Depois:**
- Menu: "Automation Rules" ou "Automações"
- Botão: "Nova Automação"
- Título: "Gerenciar Automações"

## 🌍 Comparação com Concorrentes

### HubSpot
```
Workflows → "Set up triggers and actions"
```

### Salesforce
```
Process Builder → "Automate your business processes"
Flow → "Build flows that..."
```

### Pipedrive
```
Workflow Automation → "Automate repetitive tasks"
```

### Intercom
```
Custom Bots & Workflows
```

### Zendesk
```
Triggers → "Automatic actions when..."
Automations → "Time-based actions"
```

**Todos usam:** "Automation", "Workflows", "Triggers"
**Ninguém usa:** "Follow-up Rules"

## 📊 Categorização dos Casos de Uso

Com o novo nome, podemos categorizar melhor:

### 1. **Response Automation** (o que era "follow-up")
- Responder após X tempo sem resposta
- Escalar se não resolver
- Pedir feedback após resolução

### 2. **Event Automation** (novo!)
- Mensagem imediata ao comprar
- Boas-vindas ao virar lead
- Notificar ao mudar status

### 3. **Scheduled Automation** (novo!)
- Newsletter semanal
- Relatório mensal
- Lembrete recorrente

### 4. **Condition-Based Automation** (novo!)
- Se carrinho > R$100 → cupom de frete grátis
- Se 5+ mensagens → escalar para humano
- Se inativo 30 dias → campanha reativação

## 🎨 Exemplos com Novo Nome

### Exemplo 1: Automação de Compra
```javascript
// ANTES (confuso)
POST /automation-rules
{
  "name": "Mensagem pós-compra",
  "trigger": "status.changed",
  "actions": [{
    "type": "send_message",
    "params": {"content": "Obrigado pela compra!"}
  }]
}

// DEPOIS (claro)
POST /automation-rules
{
  "name": "Automação de Compra",
  "trigger": "status.changed",
  "actions": [{
    "type": "send_message",
    "params": {"content": "Obrigado pela compra!"}
  }]
}
```

### Exemplo 2: Automação de Reativação
```javascript
// ANTES (ok, esse faz sentido como follow-up)
{
  "name": "Follow-up cliente inativo",
  "trigger": "no_response.timeout",
  ...
}

// DEPOIS (ainda funciona, mas mais abrangente)
{
  "name": "Reativação de Cliente Inativo",
  "trigger": "no_response.timeout",
  ...
}
```

## 🚀 Plano de Migração

### Fase 1: Aliasing (Retrocompatibilidade)
```go
// Mantém ambos os nomes temporariamente
type AutomationRule = AutomationRule
type AutomationEngine = AutomationEngine

// API aceita ambos
POST /automation-rules  // deprecated
POST /automation-rules // novo
```

### Fase 2: Migração Gradual
- Atualizar documentação
- Atualizar UI
- Avisar usuários sobre deprecation
- Logs de uso do endpoint antigo

### Fase 3: Remoção
- Após 3-6 meses, remover aliases
- Manter apenas `automation-rules`

## 💬 Glossário

| Conceito Antigo | Conceito Novo | Descrição |
|-----------------|---------------|-----------|
| Follow-up Rule | Automation Rule | Regra de automação do pipeline |
| Follow-up Engine | Automation Engine | Motor que executa automações |
| Follow-up Manager | Automation Manager | Gerenciador de regras |
| Follow-up Action | Automation Action | Ação executada pela regra |
| Follow-up Trigger | Automation Trigger | Evento que dispara regra |

## ✅ Decisão

**Proposta:** Renomear para **"Automation Rules"** ou **"Pipeline Automation Rules"**

**Benefícios:**
1. ✅ Nome mais preciso e abrangente
2. ✅ Alinhado com concorrentes
3. ✅ Melhor UX (usuários entendem imediatamente)
4. ✅ Permite crescimento do conceito
5. ✅ Marketing-friendly

**Custos:**
- ⚠️ Refactoring de código (com aliases, é suave)
- ⚠️ Atualizar documentação
- ⚠️ Avisar usuários (se já houver)

**Quando fazer:**
- 🟢 AGORA (antes de lançar em produção)
- 🔴 Não fazer se já tem muitos usuários usando "follow-up"
