# Exemplos de Automações Genéricas

Este documento contém exemplos de automações usando o novo sistema genérico de `automation_rules`.

## Tipos de Automação

- `pipeline_automation`: Automações relacionadas a pipelines
- `scheduled_report`: Relatórios agendados
- `time_based_notification`: Notificações baseadas em tempo
- `webhook_automation`: Automações disparadas por webhooks
- `custom`: Automações customizadas

## Exemplo 1: Relatório Diário de Performance de Agentes

Envia relatório diário às 18h com performance dos agentes para o coordenador de vendas.

```json
{
  "automation_type": "scheduled_report",
  "pipeline_id": null,
  "tenant_id": "tenant-123",
  "name": "Relatório Diário de Performance",
  "description": "Gera relatório com métricas dos agentes e envia ao coordenador",
  "trigger": "scheduled",
  "schedule": {
    "type": "cron",
    "cron_expr": "0 18 * * *",
    "timezone": "America/Sao_Paulo"
  },
  "conditions": [],
  "actions": [
    {
      "type": "create_agent_report",
      "params": {
        "agent_id": "all",
        "period": "daily",
        "include_comparisons": true,
        "notify_coordinator": false
      }
    },
    {
      "type": "create_note",
      "params": {
        "entity_type": "agent",
        "entity_id": "{{best_agent_id}}",
        "title": "Top Performance Today",
        "content": "{{performance_summary}}"
      }
    },
    {
      "type": "send_message",
      "params": {
        "channel_id": "{{coordinator_channel_id}}",
        "content": "📊 Relatório diário disponível!\n\n{{report_summary}}\n\n🏆 Destaque: {{best_agent_name}}"
      }
    },
    {
      "type": "send_webhook",
      "params": {
        "url": "https://api.empresa.com/reports/daily",
        "payload": {
          "report_type": "agent_performance",
          "data": "{{report_data}}"
        },
        "headers": {
          "Authorization": "Bearer {{webhook_token}}"
        }
      }
    }
  ],
  "priority": 0,
  "enabled": true
}
```

## Exemplo 2: Notificação de Lead Inativo (Pipeline Automation)

Automação de pipeline que notifica o coordenador quando um lead fica inativo por 3 dias.

```json
{
  "automation_type": "pipeline_automation",
  "pipeline_id": "pipeline-uuid-123",
  "tenant_id": "tenant-123",
  "name": "Alerta de Lead Inativo",
  "description": "Notifica coordenador sobre leads inativos há 3+ dias",
  "trigger": "no_response.timeout",
  "conditions": [
    {
      "field": "days_since_last_message",
      "operator": "gte",
      "value": 3
    },
    {
      "field": "pipeline_status",
      "operator": "in",
      "value": ["Lead", "Qualificado"]
    }
  ],
  "actions": [
    {
      "type": "add_tag",
      "params": {
        "tag": "lead_inativo"
      }
    },
    {
      "type": "create_note",
      "params": {
        "entity_type": "contact",
        "entity_id": "{{contact_id}}",
        "title": "Lead Inativo - Requer Atenção",
        "content": "Lead sem interação há {{days_since_last_message}} dias. Última mensagem: {{last_message}}"
      }
    },
    {
      "type": "notify_coordinator",
      "params": {
        "message": "⚠️ Lead inativo: {{contact_name}} ({{days_since_last_message}} dias)\nAgente: {{agent_name}}",
        "channel": "whatsapp",
        "priority": "high"
      }
    }
  ],
  "priority": 10,
  "enabled": true
}
```

## Exemplo 3: Webhook de Integração Externa

Automação disparada por webhook externo para criar tarefas.

```json
{
  "automation_type": "webhook_automation",
  "pipeline_id": null,
  "tenant_id": "tenant-123",
  "name": "Integração CRM Externo",
  "description": "Processa webhooks do CRM externo",
  "trigger": "webhook.received",
  "conditions": [
    {
      "field": "webhook_source",
      "operator": "eq",
      "value": "external_crm"
    },
    {
      "field": "event_type",
      "operator": "eq",
      "value": "new_lead"
    }
  ],
  "actions": [
    {
      "type": "create_task",
      "params": {
        "title": "Novo Lead: {{lead_name}}",
        "description": "Lead recebido do CRM externo\nEmail: {{lead_email}}\nTelefone: {{lead_phone}}",
        "due_date": "{{tomorrow}}"
      }
    },
    {
      "type": "assign_agent",
      "params": {
        "agent_id": "{{round_robin_agent}}"
      }
    },
    {
      "type": "send_webhook",
      "params": {
        "url": "{{callback_url}}",
        "payload": {
          "status": "processed",
          "contact_id": "{{contact_id}}"
        }
      }
    }
  ],
  "priority": 0,
  "enabled": true
}
```

## Exemplo 4: Notificação Agendada de Lembrete

Envia lembrete semanal para agentes revisarem suas metas.

```json
{
  "automation_type": "time_based_notification",
  "pipeline_id": null,
  "tenant_id": "tenant-123",
  "name": "Lembrete Semanal de Metas",
  "description": "Envia lembrete toda segunda às 9h",
  "trigger": "scheduled",
  "schedule": {
    "type": "cron",
    "cron_expr": "0 9 * * 1",
    "timezone": "America/Sao_Paulo"
  },
  "conditions": [
    {
      "field": "agent_role",
      "operator": "eq",
      "value": "sales"
    }
  ],
  "actions": [
    {
      "type": "notify_agent",
      "params": {
        "agent_id": "{{agent_id}}",
        "message": "🎯 Bom dia! Hora de revisar suas metas da semana.\n\nMeta: {{weekly_goal}}\nProgresso: {{current_progress}}%",
        "channel": "in_app"
      }
    },
    {
      "type": "create_note",
      "params": {
        "entity_type": "agent",
        "entity_id": "{{agent_id}}",
        "title": "Revisão Semanal de Metas",
        "content": "Lembrete enviado em {{timestamp}}"
      }
    }
  ],
  "priority": 0,
  "enabled": true
}
```

## Exemplo 5: Automação Composta (Multiple Actions)

Automação que combina múltiplas ações quando um contato muda de status.

```json
{
  "automation_type": "pipeline_automation",
  "pipeline_id": "pipeline-uuid-123",
  "tenant_id": "tenant-123",
  "name": "Cliente Convertido - Comemoração",
  "description": "Ações quando lead vira cliente",
  "trigger": "status.changed",
  "conditions": [
    {
      "field": "new_status",
      "operator": "eq",
      "value": "Cliente"
    }
  ],
  "actions": [
    {
      "type": "add_tag",
      "params": {
        "tag": "cliente_ativo"
      }
    },
    {
      "type": "send_message",
      "params": {
        "content": "🎉 Bem-vindo à família! Estamos muito felizes em tê-lo como cliente."
      }
    },
    {
      "type": "create_note",
      "params": {
        "entity_type": "contact",
        "entity_id": "{{contact_id}}",
        "title": "Conversão Realizada",
        "content": "Lead convertido em {{conversion_date}}\nAgente: {{agent_name}}\nTempo total: {{days_in_pipeline}} dias"
      }
    },
    {
      "type": "notify_coordinator",
      "params": {
        "message": "🎊 Nova conversão!\n\nCliente: {{contact_name}}\nAgente: {{agent_name}}\nValor: {{deal_value}}",
        "channel": "whatsapp",
        "priority": "medium"
      }
    },
    {
      "type": "send_webhook",
      "params": {
        "url": "https://api.empresa.com/integrations/customer-onboarding",
        "payload": {
          "customer_id": "{{contact_id}}",
          "name": "{{contact_name}}",
          "email": "{{contact_email}}",
          "agent": "{{agent_name}}"
        }
      },
      "delay_minutes": 5
    },
    {
      "type": "create_task",
      "params": {
        "title": "Follow-up Pós-Venda: {{contact_name}}",
        "description": "Verificar satisfação e próximos passos",
        "due_date": "{{7_days_from_now}}"
      },
      "delay_minutes": 1440
    }
  ],
  "priority": 0,
  "enabled": true
}
```

## Variáveis Disponíveis para Interpolação

As seguintes variáveis podem ser usadas nas ações com a sintaxe `{{variable_name}}`:

### Variáveis de Contexto
- `{{tenant_id}}`: ID do tenant
- `{{contact_id}}`: ID do contato
- `{{contact_name}}`: Nome do contato
- `{{contact_email}}`: Email do contato
- `{{contact_phone}}`: Telefone do contato
- `{{agent_id}}`: ID do agente
- `{{agent_name}}`: Nome do agente
- `{{session_id}}`: ID da sessão
- `{{pipeline_id}}`: ID do pipeline
- `{{pipeline_status}}`: Status atual no pipeline

### Variáveis Temporais
- `{{timestamp}}`: Timestamp atual
- `{{date}}`: Data atual (YYYY-MM-DD)
- `{{time}}`: Hora atual (HH:MM:SS)
- `{{tomorrow}}`: Data de amanhã
- `{{7_days_from_now}}`: Data daqui 7 dias

### Variáveis de Métricas
- `{{days_since_last_message}}`: Dias desde última mensagem
- `{{message_count}}`: Contagem de mensagens
- `{{conversion_date}}`: Data de conversão
- `{{days_in_pipeline}}`: Dias no pipeline

### Variáveis de Relatórios
- `{{report_data}}`: Dados do relatório (JSON)
- `{{report_summary}}`: Resumo do relatório
- `{{best_agent_id}}`: ID do melhor agente
- `{{best_agent_name}}`: Nome do melhor agente
- `{{performance_summary}}`: Resumo de performance

## Triggers Disponíveis

### Pipeline Triggers
- `session.ended`: Sessão finalizada
- `session.timeout`: Sessão timeout
- `session.resolved`: Sessão resolvida
- `session.escalated`: Sessão escalada
- `no_response.timeout`: Timeout sem resposta
- `message.received`: Mensagem recebida
- `status.changed`: Status mudou
- `stage.completed`: Etapa completada

### Time-based Triggers
- `scheduled`: Agendado (usa cron)
- `after.delay`: Após delay

### Custom Triggers
- `webhook.received`: Webhook recebido
- `custom`: Customizado

## Categorias de Actions

### Messaging
- `send_message`: Enviar mensagem
- `send_template`: Enviar template
- `send_email`: Enviar email

### Pipeline
- `change_pipeline_status`: Mudar status
- `assign_agent`: Atribuir agente
- `assign_to_queue`: Atribuir à fila

### Organization
- `add_tag`: Adicionar tag
- `remove_tag`: Remover tag
- `update_custom_field`: Atualizar campo customizado

### Tasks & Notes
- `create_task`: Criar tarefa
- `create_note`: Criar nota

### Reports & Analytics
- `create_agent_report`: Gerar relatório de agente

### Integration
- `send_webhook`: Enviar webhook
- `trigger_workflow`: Disparar workflow

### Notifications
- `notify_agent`: Notificar agente
- `notify_coordinator`: Notificar coordenador
