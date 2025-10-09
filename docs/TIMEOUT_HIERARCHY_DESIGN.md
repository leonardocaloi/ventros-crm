# Session Timeout Hierarchy Design

## Overview

O sistema de timeout de sessões segue uma **hierarquia elegante de overrides** em 3 níveis:

```
Project (base) → Channel (override) → Pipeline (final override)
```

Cada nível pode opcionalmente sobrescrever o nível anterior, permitindo configuração granular enquanto mantém defaults sensatos.

## Hierarquia Detalhada

### 1. **Project (Base Level)**
- **Campo**: `projects.session_timeout_minutes` (NOT NULL, default: 30)
- **Propósito**: Timeout padrão para TODO o projeto/CRM
- **Configurável**: Via settings do CRM
- **Escopo**: Todos os canais e pipelines herdam este valor por padrão

```sql
-- Exemplo: Configurar timeout de 60 minutos para todo o projeto
UPDATE projects
SET session_timeout_minutes = 60
WHERE id = 'project-uuid';
```

### 2. **Channel (Override Level)**
- **Campo**: `channels.session_timeout_minutes` (NULL allowed)
- **Propósito**: Override opcional do timeout para canais específicos
- **Configurável**: Via API ao criar/editar canal
- **Escopo**: Sobrescreve o timeout do projeto para este canal
- **Regra**: `NULL` = herda do projeto

```sql
-- Exemplo: WhatsApp pode ter timeout diferente de Telegram
UPDATE channels
SET session_timeout_minutes = 15
WHERE id = 'whatsapp-channel-uuid';
-- Telegram herda do projeto (NULL)
```

### 3. **Pipeline (Final Override)**
- **Campo**: `pipelines.session_timeout_minutes` (NULL allowed)
- **Propósito**: Override final e mais específico
- **Configurável**: Via API ao criar/editar pipeline
- **Escopo**: Sobrescreve canal E projeto para este pipeline
- **Regra**: `NULL` = herda do canal ou projeto

```sql
-- Exemplo: Pipeline de suporte pode ter timeout maior
UPDATE pipelines
SET session_timeout_minutes = 120
WHERE id = 'support-pipeline-uuid';
```

## Fluxo de Resolução

```typescript
function resolveSessionTimeout(projectID, channelID, pipelineID?): number {
  // 1. Busca configurações
  const project = getProject(projectID);
  const channel = getChannel(channelID);
  const pipeline = pipelineID ? getPipeline(pipelineID) : null;

  // 2. Aplica hierarquia (cascade de overrides)
  let timeout = project.session_timeout_minutes; // Base: sempre tem valor

  if (channel.session_timeout_minutes !== null) {
    timeout = channel.session_timeout_minutes; // Override: canal
  }

  if (pipeline?.session_timeout_minutes !== null) {
    timeout = pipeline.session_timeout_minutes; // Final override: pipeline
  }

  return timeout;
}
```

## Exemplos de Uso

### Cenário 1: Projeto com timeout padrão
```
Project: 30 min (base)
Channel: NULL (herda)
Pipeline: NULL (herda)
→ Resultado: 30 minutos
```

### Cenário 2: Canal com override
```
Project: 30 min (base)
Channel: 15 min (override)
Pipeline: NULL (herda do canal)
→ Resultado: 15 minutos
```

### Cenário 3: Pipeline com override final
```
Project: 30 min (base)
Channel: 15 min (override)
Pipeline: 60 min (final override)
→ Resultado: 60 minutos
```

### Cenário 4: Mistura de heranças
```
Project: 30 min (base)
Channel: NULL (herda projeto → 30 min)
Pipeline: 45 min (override)
→ Resultado: 45 minutos
```

## Implementação no Código

### Método de Resolução
`internal/application/message/process_inbound_message.go:resolveSessionTimeout()`

```go
func (uc *ProcessInboundMessageUseCase) resolveSessionTimeout(
  ctx context.Context,
  projectID,
  channelID uuid.UUID,
) (int, error) {
  // Query otimizada que busca project e channel em uma só query
  result := uc.db.Raw(`
    SELECT
      c.session_timeout_minutes as channel_timeout,
      COALESCE(p.session_timeout_minutes, 30) as project_timeout
    FROM channels c
    JOIN projects p ON p.id = c.project_id
    WHERE c.id = ? AND p.id = ?
  `, channelID, projectID)

  // Aplica hierarquia
  if result.ChannelTimeout != nil {
    return *result.ChannelTimeout // Override do canal
  }

  return result.ProjectTimeout // Base do projeto
}
```

### Pipeline Override
O override do pipeline é aplicado DEPOIS de resolver project/channel:

```go
pipelineInfo := uc.findActivePipelineWithTimeout(ctx, projectID)

if pipelineInfo.TimeoutOverride != nil {
  timeout = *pipelineInfo.TimeoutOverride // Final override
}
```

## Migrations

### Migration 000024: Project Base Timeout
```sql
ALTER TABLE projects
ADD COLUMN session_timeout_minutes INT NOT NULL DEFAULT 30;
```

### Migration 000025: Hierarchy Overrides
```sql
-- Channel override (nullable)
ALTER TABLE channels
ADD COLUMN session_timeout_minutes INT NULL;

-- Pipeline override (nullable)
ALTER TABLE pipelines
ADD COLUMN session_timeout_minutes INT NULL;
```

## Benefícios da Arquitetura

1. **Flexibilidade**: Cada nível pode customizar conforme necessário
2. **Simplicidade**: `NULL` = herda automaticamente
3. **Performance**: Single query resolve project + channel
4. **Manutenção**: Fácil de entender a precedência
5. **Default Sensato**: 30 min em todos os níveis
6. **Configurável**: UI pode expor cada nível independentemente

## UI/UX Sugerido

### Settings do Projeto
```
┌─────────────────────────────────────────┐
│ Session Timeout Configuration           │
├─────────────────────────────────────────┤
│ Default Timeout:  [30] minutes          │
│ ℹ️ Applied to all sessions unless       │
│    overridden by channel or pipeline    │
└─────────────────────────────────────────┘
```

### Configuração de Canal
```
┌─────────────────────────────────────────┐
│ Channel Settings                        │
├─────────────────────────────────────────┤
│ Session Timeout:                        │
│ ( ) Use project default (30 min)       │
│ (•) Custom: [15] minutes                │
└─────────────────────────────────────────┘
```

### Configuração de Pipeline
```
┌─────────────────────────────────────────┐
│ Pipeline Settings                       │
├─────────────────────────────────────────┤
│ Session Timeout:                        │
│ ( ) Use channel/project default         │
│ (•) Custom: [60] minutes                │
└─────────────────────────────────────────┘
```

## Testing

```bash
# Test 1: Project base timeout
make test-timeout-project

# Test 2: Channel override
make test-timeout-channel

# Test 3: Pipeline final override
make test-timeout-pipeline
```
