# ✅ CONFIGURAÇÃO FINAL: Consolidação com 4 Horas

**Data**: 2025-10-15
**Status**: ✅ APLICADO E RODANDO

---

## 🎯 ALTERAÇÕES REALIZADAS

### 1. FIX do `last_activity_at` (BUG CRÍTICO)

**Arquivo**: `internal/application/message/process_inbound_message.go`
**Linha**: 242

```diff
- if err := s.RecordMessage(true, cmd.Timestamp); err != nil {
+ if err := s.RecordMessage(true, cmd.ReceivedAt); err != nil {
```

**Problema corrigido**: `cmd.Timestamp` era sempre zero (campo legado nunca preenchido)
**Solução**: Usar `cmd.ReceivedAt` que contém o timestamp histórico correto

---

### 2. TIMEOUT aumentado de 30min → 4h

#### Arquivo 1: `internal/workflows/channel/waha_history_import_workflow.go`
**Linha**: 43

```diff
- input.SessionTimeoutMinutes = 30 // Default: 30 minutos
+ input.SessionTimeoutMinutes = 240 // Default: 4 horas (para máxima consolidação)
```

#### Arquivo 2: `internal/application/message/process_inbound_message.go`
**Linha**: 404

```diff
- timeoutDuration = 30 * time.Minute
+ timeoutDuration = 4 * time.Hour // 240 minutos para máxima consolidação
```

---

## 📊 IMPACTO ESPERADO

### Consolidação: 30 min vs 4 horas

| Cenário | Timeout 30min | Timeout 4h |
|---------|---------------|------------|
| **Gap: 1h30** | 2 sessions ❌ | 1 session ✅ |
| **Gap: 3h45** | 2 sessions ❌ | 1 session ✅ |
| **Gap: 6h** | 2 sessions ❌ | 2 sessions ❌ |

### Resultados Esperados

**ANTES** (com bug):
- 5671 messages = 5671 sessions
- Ratio: 1.0 (zero consolidação)
- last_activity_at: 0001-01-01 (100% bugado)

**DEPOIS** (com fix + 4h timeout):
- 5671 messages → **2500-3400 sessions**
- Ratio: **1.7-2.3** (ótimo!)
- last_activity_at: timestamps reais ✅
- **Redução: 40-60% das sessions!** 🎉

---

## 🔧 COMO APLICAR EM PRODUÇÃO

### 1. Recompilar o código

```bash
make build
```

### 2. Reiniciar API

```bash
pkill -9 -f "crm-api"
./bin/crm-api > /tmp/api.log 2>&1 &
```

### 3. Verificar se está rodando

```bash
curl http://localhost:8080/health
```

### 4. Re-executar import histórico

```bash
# Limpar dados antigos (têm o bug)
PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "
TRUNCATE messages, sessions, contacts CASCADE;"

# Disparar novo import com FIX aplicado
curl -X POST "http://localhost:8080/api/v1/crm/channels/{CHANNEL_ID}/import-history" \
  -H "Content-Type: application/json" \
  -H "X-Dev-User-ID: {USER_ID}" \
  -d '{"time_range_days": 0, "limit_per_chat": 0}'
```

### 5. Aguardar conclusão

Acompanhar via:
- Temporal UI: http://localhost:8282
- Logs: `tail -f /tmp/api.log`

### 6. Verificar resultados

```sql
-- Verificar se last_activity_at foi preenchido corretamente
SELECT
    COUNT(*) as total_sessions,
    COUNT(*) FILTER (WHERE last_activity_at = '0001-01-01 00:00:00') as ❌_bug,
    COUNT(*) FILTER (WHERE last_activity_at > '2024-01-01') as ✅_fixed,
    COUNT(DISTINCT contact_id) as unique_contacts,
    ROUND(COUNT(*)::numeric / COUNT(DISTINCT contact_id), 2) as sessions_per_contact
FROM sessions;

-- Verificar consolidação
SELECT
    COUNT(*) as total_messages,
    COUNT(DISTINCT session_id) as total_sessions,
    ROUND(COUNT(*)::numeric / COUNT(DISTINCT session_id), 2) as messages_per_session
FROM messages;
```

**Valores esperados**:
- `❌_bug = 0` (nenhuma session com zero time)
- `✅_fixed > 0` (todas as sessions com timestamp real)
- `sessions_per_contact ≈ 2-4` (boa consolidação)
- `messages_per_session ≈ 1.7-2.3` (excelente!)

---

## 🎛️ CONFIGURAÇÃO CUSTOMIZADA

Para usar um timeout diferente de 4h:

### Opção 1: Via API (por import)

```bash
curl -X POST "http://localhost:8080/api/v1/crm/channels/{CHANNEL_ID}/import-history" \
  -H "Content-Type: application/json" \
  -d '{
    "time_range_days": 0,
    "limit_per_chat": 0,
    "session_timeout_minutes": 120  # 2 horas
  }'
```

### Opção 2: Via Project (configuração padrão)

```sql
-- Alterar timeout padrão do projeto
UPDATE projects
SET session_timeout_minutes = 480  -- 8 horas
WHERE id = '{PROJECT_ID}';
```

### Opção 3: Via Pipeline (override por funil)

```sql
-- Alterar timeout de um pipeline específico
UPDATE pipelines
SET session_timeout_minutes = 60  -- 1 hora
WHERE id = '{PIPELINE_ID}';
```

**Ordem de prioridade**:
1. Pipeline (mais específico)
2. Channel import parameter
3. Project default
4. Sistema default (4h)

---

## 📋 CHECKLIST DE VALIDAÇÃO

- [x] Fix do `last_activity_at` aplicado (cmd.ReceivedAt)
- [x] Timeout aumentado para 4h (240 min)
- [x] Código recompilado (`make build`)
- [x] API reiniciada e respondendo
- [ ] Import histórico re-executado
- [ ] Validação SQL confirma:
  - [ ] `zero_time_bug_count = 0`
  - [ ] `messages_per_session > 1.5`
  - [ ] `sessions_per_contact < 5`

---

## 🚨 TROUBLESHOOTING

### Problema: Consolidação ainda não funciona

**Verificar**:
```sql
SELECT last_activity_at, COUNT(*)
FROM sessions
GROUP BY last_activity_at
LIMIT 5;
```

**Se ainda mostra zero time**:
- API não foi reiniciada com código novo
- Import foi executado antes do fix
- Solução: Re-executar import histórico

### Problema: Sessions não consolidam mesmo com fix

**Verificar gaps**:
```sql
WITH session_gaps AS (
    SELECT
        s1.id as earlier_id,
        s2.id as later_id,
        s1.last_activity_at,
        s2.started_at,
        EXTRACT(EPOCH FROM (s2.started_at - s1.last_activity_at))/60 as gap_minutes
    FROM sessions s1
    CROSS JOIN sessions s2
    WHERE s1.contact_id = s2.contact_id
      AND s1.last_activity_at < s2.started_at
    LIMIT 10
)
SELECT
    gap_minutes,
    CASE
        WHEN gap_minutes <= 240 THEN '✅ Deveria consolidar'
        ELSE '❌ Gap muito grande'
    END as status
FROM session_gaps;
```

---

**Autor**: Claude Code
**Data**: 2025-10-15
**Versão**: 1.0 (4h timeout)
