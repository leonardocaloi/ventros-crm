# Automation API Reference

## Visão Geral

A API de Automação permite descobrir e configurar regras de automação para pipelines, incluindo tipos de automação, triggers, ações, condições e operadores lógicos.

**Base URL:** `/api/v1/automation`

**Autenticação:** Todas as rotas requerem autenticação via Bearer token

---

## Endpoints de Discovery

### GET /types

Lista todos os tipos de automação disponíveis.

**Resposta 200:**
```json
[
  {
    "code": "follow_up",
    "name": "Follow-up",
    "description": "Automações de acompanhamento após inatividade ou falta de resposta",
    "icon": "clock",
    "examples": [
      "Enviar mensagem após 24h sem resposta",
      "Lembrete de pagamento pendente",
      "Recuperação de carrinho abandonado"
    ]
  },
  {
    "code": "event",
    "name": "Evento",
    "description": "Automações disparadas por eventos específicos do sistema",
    "icon": "zap",
    "examples": [
      "Confirmação de compra imediata",
      "Notificação de mudança de status"
    ]
  },
  {
    "code": "scheduled",
    "name": "Agendada",
    "description": "Automações recorrentes ou agendadas para horários específicos",
    "icon": "calendar",
    "examples": [
      "Newsletter semanal às segundas 10h",
      "Relatório mensal automático"
    ]
  },
  {
    "code": "reengagement",
    "name": "Reativação",
    "description": "Automações para reengajar clientes inativos ou em churn",
    "icon": "refresh",
    "examples": [
      "Cliente voltou para Lead após 90 dias",
      "Campanha de reativação para inativos"
    ]
  },
  {
    "code": "onboarding",
    "name": "Onboarding",
    "description": "Automações de boas-vindas e integração de novos contatos",
    "icon": "user-plus",
    "examples": [
      "Mensagem de boas-vindas ao novo lead",
      "Sequência de onboarding em 5 dias"
    ]
  },
  {
    "code": "custom",
    "name": "Customizada",
    "description": "Automações personalizadas com lógica específica do seu negócio",
    "icon": "settings",
    "examples": []
  }
]
```

---

### GET /triggers

Lista todos os triggers disponíveis (system + custom).

**Query Parameters:**
- `category` (opcional): Filtra por categoria (`session`, `message`, `pipeline`, `temporal`, `transaction`, `behavior`)

**Resposta 200:**
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
          "description": "ID da sessão",
          "example": ""
        },
        {
          "name": "message_count",
          "type": "int",
          "description": "Total de mensagens na sessão",
          "example": ""
        }
      ]
    },
    {
      "code": "purchase.completed",
      "name": "Compra Concluída",
      "description": "Disparado quando uma compra é finalizada com sucesso",
      "category": "transaction",
      "is_system": true,
      "parameters": [
        {
          "name": "order_id",
          "type": "uuid",
          "description": "",
          "example": ""
        },
        {
          "name": "amount",
          "type": "float",
          "description": "Valor da compra",
          "example": ""
        },
        {
          "name": "payment_method",
          "type": "string",
          "description": "Método de pagamento utilizado",
          "example": ""
        }
      ]
    }
  ],
  "custom_triggers": []
}
```

**Exemplo de uso:**
```bash
# Todos os triggers
curl -H "Authorization: Bearer $TOKEN" \
  https://api.ventros.com/api/v1/automation/triggers

# Apenas triggers de transação
curl -H "Authorization: Bearer $TOKEN" \
  https://api.ventros.com/api/v1/automation/triggers?category=transaction
```

---

### GET /triggers/:code

Retorna detalhes de um trigger específico.

**Path Parameters:**
- `code`: Código do trigger (ex: `session.ended`, `purchase.completed`)

**Resposta 200:**
```json
{
  "code": "no_response.timeout",
  "name": "Sem Resposta",
  "description": "Disparado quando cliente não responde há X tempo",
  "category": "message",
  "is_system": true,
  "parameters": [
    {
      "name": "hours_since_last_message",
      "type": "float",
      "description": "",
      "example": ""
    },
    {
      "name": "message_count",
      "type": "int",
      "description": "",
      "example": ""
    }
  ]
}
```

**Resposta 404:**
```json
{
  "error": "trigger not found"
}
```

---

### GET /actions

Lista todas as ações disponíveis.

**Query Parameters:**
- `category` (opcional): Filtra por categoria (`messaging`, `pipeline`, `assignment`, `tasks`, `integration`, `organization`, `data`, `workflow`)

**Resposta 200:**
```json
[
  {
    "code": "send_message",
    "name": "Enviar Mensagem",
    "description": "Envia mensagem de texto para o contato",
    "category": "messaging",
    "parameters": [
      {
        "name": "content",
        "type": "string",
        "required": true,
        "description": "Conteúdo da mensagem",
        "default": null
      }
    ],
    "example": {
      "content": "Olá! Como posso ajudar?"
    }
  },
  {
    "code": "change_pipeline_status",
    "name": "Mudar Status",
    "description": "Altera o status do contato no pipeline",
    "category": "pipeline",
    "parameters": [
      {
        "name": "status_id",
        "type": "string",
        "required": true,
        "description": "UUID do novo status",
        "default": null
      }
    ],
    "example": {
      "status_id": "uuid-here"
    }
  },
  {
    "code": "send_webhook",
    "name": "Enviar Webhook",
    "description": "Dispara webhook para URL externa",
    "category": "integration",
    "parameters": [
      {
        "name": "url",
        "type": "string",
        "required": true,
        "description": "URL do webhook"
      },
      {
        "name": "payload",
        "type": "object",
        "required": false,
        "description": "Dados a enviar"
      },
      {
        "name": "headers",
        "type": "object",
        "required": false,
        "description": "Headers HTTP customizados"
      }
    ],
    "example": {
      "url": "https://api.exemplo.com/webhook",
      "payload": {
        "event": "status_changed"
      }
    }
  }
]
```

---

### GET /conditions/operators

Lista todos os operadores de condição disponíveis.

**Resposta 200:**
```json
[
  {
    "code": "eq",
    "name": "Igual a",
    "description": "Valor igual ao especificado",
    "example": "status == 'Lead'"
  },
  {
    "code": "gt",
    "name": "Maior que",
    "description": "Valor maior que o especificado",
    "example": "message_count > 5"
  },
  {
    "code": "contains",
    "name": "Contém",
    "description": "String contém o valor especificado",
    "example": "message contains 'urgente'"
  },
  {
    "code": "in",
    "name": "Está em",
    "description": "Valor está na lista especificada",
    "example": "status in ['Lead', 'Qualificado']"
  }
]
```

---

### GET /logic-operators

Lista operadores lógicos para combinar condições (AND/OR).

**Resposta 200:**
```json
[
  {
    "code": "AND",
    "name": "E (AND)",
    "description": "Todas as condições devem ser verdadeiras"
  },
  {
    "code": "OR",
    "name": "OU (OR)",
    "description": "Pelo menos uma condição deve ser verdadeira"
  }
]
```

---

### GET /discovery

Retorna TODOS os metadados de automação em uma única chamada (types, triggers, actions, operators, logic).

**Resposta 200:**
```json
{
  "types": [...],
  "triggers": [...],
  "actions": [...],
  "operators": [...],
  "logic_types": [...]
}
```

**Exemplo de uso:**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  https://api.ventros.com/api/v1/automation/discovery
```

**Quando usar:** Ideal para carregar todas as opções de uma vez ao abrir a interface de criação de automações no frontend.

---

## Endpoints de Gerenciamento (Custom Triggers)

### POST /triggers/custom

Registra um trigger customizado.

**Permissão:** Requer permissão de admin

**Request Body:**
```json
{
  "code": "custom.payment_received",
  "name": "Pagamento Recebido",
  "description": "Disparado quando webhook de pagamento confirma recebimento",
  "parameters": [
    {
      "name": "payment_id",
      "type": "uuid",
      "description": "ID da transação",
      "example": ""
    },
    {
      "name": "amount",
      "type": "float",
      "description": "Valor pago",
      "example": ""
    },
    {
      "name": "payment_method",
      "type": "string",
      "description": "Método de pagamento (pix, card, etc)",
      "example": ""
    }
  ]
}
```

**Resposta 201:**
```json
{
  "message": "custom trigger registered successfully",
  "trigger": {
    "code": "custom.payment_received",
    "name": "Pagamento Recebido",
    "description": "Disparado quando webhook de pagamento confirma recebimento",
    "category": "custom",
    "is_system": false,
    "parameters": [...]
  }
}
```

**Resposta 400:**
```json
{
  "error": "custom triggers must start with 'custom.' prefix"
}
```

**Regras de validação:**
- ✅ Código DEVE começar com `custom.`
- ❌ NÃO pode sobrescrever triggers do sistema
- ✅ Ilimitados por tenant

---

### DELETE /triggers/custom/:code

Remove um trigger customizado.

**Permissão:** Requer permissão de admin

**Path Parameters:**
- `code`: Código do trigger customizado (ex: `custom.payment_received`)

**Resposta 200:**
```json
{
  "message": "custom trigger unregistered successfully"
}
```

**Resposta 400:**
```json
{
  "error": "cannot unregister system trigger"
}
```

---

## Categorias de Triggers

### Session (4 triggers)
- `session.ended` - Sessão encerrada normalmente
- `session.timeout` - Sessão expirou por inatividade
- `session.resolved` - Sessão marcada como resolvida
- `session.escalated` - Sessão escalada para outro nível

### Message (2 triggers)
- `no_response.timeout` - Cliente não responde há X tempo
- `message.received` - Nova mensagem recebida

### Pipeline (2 triggers)
- `status.changed` - Status do contato mudou
- `stage.completed` - Etapa do pipeline concluída

### Temporal (2 triggers)
- `after.delay` - Após delay específico desde evento
- `scheduled` - Horários agendados (cron, recorrente)

### Transaction (5 triggers) 🆕
- `purchase.completed` - Compra finalizada com sucesso
- `payment.received` - Pagamento confirmado
- `refund.issued` - Reembolso processado
- `cart.abandoned` - Carrinho abandonado
- `order.shipped` - Pedido despachado para entrega

### Behavior (3 triggers) 🆕
- `page.visited` - Contato visitou página específica
- `form.submitted` - Formulário submetido
- `file.downloaded` - Arquivo/recurso baixado

### Custom (∞ triggers)
- `custom.*` - Triggers customizados definidos pelo usuário

---

## Categorias de Actions

### Messaging (2 ações)
- `send_message` - Enviar mensagem de texto
- `send_template` - Enviar template pré-definido

### Pipeline (1 ação)
- `change_pipeline_status` - Mudar status no pipeline

### Assignment (2 ações)
- `assign_agent` - Atribuir a agente específico
- `assign_to_queue` - Atribuir à fila de atendimento

### Tasks (1 ação)
- `create_task` - Criar tarefa relacionada

### Integration (1 ação)
- `send_webhook` - Disparar webhook para URL externa

### Organization (2 ações)
- `add_tag` - Adicionar tag ao contato
- `remove_tag` - Remover tag do contato

### Data (1 ação)
- `update_custom_field` - Atualizar campo customizado

### Workflow (1 ação)
- `trigger_workflow` - Iniciar workflow Temporal

---

## Exemplos de Uso

### 1. Frontend - Carregar Opções para UI

```javascript
// Carregar todas as opções de uma vez
const response = await fetch('/api/v1/automation/discovery', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

const {
  types,        // Para dropdown "Tipo de Automação"
  triggers,     // Para dropdown "Quando Disparar"
  actions,      // Para lista de ações disponíveis
  operators,    // Para dropdown de operadores de condição
  logic_types   // Para dropdown AND/OR
} = await response.json();

// Renderizar UI com todas as opções
```

### 2. Criar Automação de Follow-up 24h

```javascript
// Primeiro, consultar triggers de mensagem
const triggers = await fetch('/api/v1/automation/triggers?category=message')
  .then(r => r.json());

// Criar regra com trigger no_response.timeout
const rule = {
  name: "Follow-up 24h sem resposta",
  trigger: "no_response.timeout",
  conditions: [
    {
      field: "hours_since_last_message",
      operator: "gte",
      value: 24
    }
  ],
  actions: [
    {
      type: "send_message",
      params: {
        content: "Olá! Vi que você não respondeu. Posso ajudar?"
      }
    }
  ]
};

// Criar via API de regras (POST /api/v1/pipelines/:id/automation-rules)
```

### 3. Criar Automação de Compra

```javascript
// Consultar actions de messaging
const actions = await fetch('/api/v1/automation/actions?category=messaging')
  .then(r => r.json());

// Criar regra com trigger de transação
const rule = {
  name: "Confirmação de compra",
  trigger: "purchase.completed",
  conditions: [], // Sem condições, sempre dispara
  actions: [
    {
      type: "send_template",
      params: {
        template_name: "purchase_confirmation",
        params: { order_id: "{{order_id}}" }
      }
    },
    {
      type: "change_pipeline_status",
      params: {
        status_id: "cliente_ativo_status_uuid"
      }
    }
  ]
};
```

### 4. Condições Complexas com AND/OR

```javascript
// Consultar operadores lógicos
const logicOps = await fetch('/api/v1/automation/logic-operators')
  .then(r => r.json());

// Criar condição complexa:
// (status == 'Lead' AND days_inactive > 30) OR (status == 'Cliente' AND days_inactive > 90)
const conditionGroup = {
  logic: "OR",
  conditions: [],
  groups: [
    {
      logic: "AND",
      conditions: [
        { field: "status", operator: "eq", value: "Lead" },
        { field: "days_inactive", operator: "gt", value: 30 }
      ],
      groups: []
    },
    {
      logic: "AND",
      conditions: [
        { field: "status", operator: "eq", value: "Cliente" },
        { field: "days_inactive", operator: "gt", value: 90 }
      ],
      groups: []
    }
  ]
};
```

---

## Sumário de Triggers (20 total)

| Categoria | Quantidade | Triggers |
|-----------|------------|----------|
| **Session** | 4 | `session.ended`, `session.timeout`, `session.resolved`, `session.escalated` |
| **Message** | 2 | `no_response.timeout`, `message.received` |
| **Pipeline** | 2 | `status.changed`, `stage.completed` |
| **Temporal** | 2 | `after.delay`, `scheduled` |
| **Transaction** | 5 | `purchase.completed`, `payment.received`, `refund.issued`, `cart.abandoned`, `order.shipped` |
| **Behavior** | 3 | `page.visited`, `form.submitted`, `file.downloaded` |
| **Custom** | ∞ | `custom.*` (definidos pelo usuário) |

---

## Sumário de Actions (11 total)

| Categoria | Quantidade | Actions |
|-----------|------------|---------|
| **Messaging** | 2 | `send_message`, `send_template` |
| **Pipeline** | 1 | `change_pipeline_status` |
| **Assignment** | 2 | `assign_agent`, `assign_to_queue` |
| **Tasks** | 1 | `create_task` |
| **Integration** | 1 | `send_webhook` |
| **Organization** | 2 | `add_tag`, `remove_tag` |
| **Data** | 1 | `update_custom_field` |
| **Workflow** | 1 | `trigger_workflow` |

---

## Sumário de Operators (8 total)

| Código | Nome | Descrição |
|--------|------|-----------|
| `eq` | Igual a | Valor igual ao especificado |
| `ne` | Diferente de | Valor diferente do especificado |
| `gt` | Maior que | Valor maior que o especificado |
| `gte` | Maior ou igual | Valor maior ou igual ao especificado |
| `lt` | Menor que | Valor menor que o especificado |
| `lte` | Menor ou igual | Valor menor ou igual ao especificado |
| `contains` | Contém | String contém o valor especificado |
| `in` | Está em | Valor está na lista especificada |

---

## Notas Importantes

### Permissões
- Endpoints de **Discovery** (GET): Apenas autenticação requerida
- Endpoints de **Custom Triggers** (POST/DELETE): Requerem permissão de admin

### Rate Limiting
- Discovery endpoints têm cache agressivo (recomendado)
- Custom trigger registration: máx 100 por tenant

### Versionamento
- API versão: `v1`
- Breaking changes: nova versão será criada (`v2`)

### Frontend Guidelines
1. **Cache os metadados** (types, triggers, actions, operators) no localStorage
2. **Use `/discovery`** para carga inicial da UI
3. **Recarregue apenas quando necessário** (ex: após criar custom trigger)
4. **Valide inputs** usando metadados antes de enviar regra

### Backend Integration
- Triggers são **avaliados em ordem de prioridade**
- Actions são **executadas sequencialmente**
- Erros em uma ação **NÃO interrompem** as demais (graceful degradation)
