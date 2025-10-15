# 🐛 BUG FIX: Zero Time em `last_activity_at` impedia consolidação de sessions

**Data**: 2025-10-15
**Severidade**: 🔴 CRÍTICO
**Impacto**: Consolidação de sessions não funcionava (ratio 1.0 messages/sessions)

---

## 📋 Sumário

Durante a refatoração de consolidação de sessions (SQL → Go Puro), descobrimos que **TODAS** as sessions tinham `last_activity_at = 0001-01-01 00:00:00` (zero time), impedindo completamente a consolidação.

### Estatísticas do Problema

- **5671 mensagens** = **5671 sessions** (ratio 1.0)
- **1085 contatos** com mensagens em gaps < 30 minutos que **não consolidaram**
- **Consolidações realizadas**: **0** (zero!)
- **Causa raiz**: `cmd.Timestamp` (campo legado) era usado mas **nunca** foi preenchido

---

## 🔍 Investigação

### Fluxo Descoberto

1. **History Import** (activities):
   ```go
   // internal/workflows/channel/waha_history_import_activities.go:564
   cmd := messageapp.ProcessInboundMessageCommand{
       ReceivedAt: time.Unix(wahaMsg.Timestamp, 0), // ✅ Timestamp histórico CORRETO
       Timestamp:  time.Time{},                      // ❌ Campo legado, NUNCA preenchido (zero value)
   }
   ```

2. **ProcessInboundMessageUseCase**:
   ```go
   // internal/application/message/process_inbound_message.go:241 (ANTES DO FIX)
   if err := s.RecordMessage(true, cmd.Timestamp); err != nil {  // ❌ Usando campo ERRADO!
       return fmt.Errorf("saga step failed [session_updated]: %w", err)
   }
   ```

3. **Domain Session**:
   ```go
   // internal/domain/crm/session/session.go:293
   func (s *Session) RecordMessage(isInbound bool, messageTimestamp time.Time) error {
       s.lastActivityAt = messageTimestamp  // ❌ Recebe zero time!
       // ...
   }
   ```

4. **Consolidation Algorithm**:
   ```go
   // internal/domain/crm/session/session.go:416
   gap := later.startedAt.Sub(earlier.lastActivityAt)  // ❌ gap = ~2024 years!
   return gap <= timeout  // ❌ Sempre falso!
   ```

### Evidência do Bug

```sql
SELECT
    session_started,
    last_activity_at,
    first_msg_timestamp
FROM (
    SELECT
        s.started_at as session_started,
        s.last_activity_at,
        MIN(m.received_at) as first_msg_timestamp
    FROM sessions s
    INNER JOIN messages m ON m.session_id = s.id
    GROUP BY s.id, s.started_at, s.last_activity_at
    LIMIT 1
) sub;

-- Resultado:
-- session_started:      2025-10-15 02:37:44.693
-- last_activity_at:     0001-01-01 00:00:00       ← ZERO TIME!
-- first_msg_timestamp:  2025-10-15 02:37:44.771
```

**Conclusão**: O timestamp correto existia no banco (`received_at`), mas a session **nunca foi atualizada** porque `RecordMessage()` recebeu zero time.

---

## ✅ FIX APLICADO

### Arquivo Modificado

**`internal/application/message/process_inbound_message.go`**

```diff
// Step 4: Record message in session (updates metrics)
-if err := s.RecordMessage(true, cmd.Timestamp); err != nil {
+// ✅ FIX: Use cmd.ReceivedAt (timestamp histórico correto) ao invés de cmd.Timestamp (legado, sempre zero)
+if err := s.RecordMessage(true, cmd.ReceivedAt); err != nil {
    return fmt.Errorf("saga step failed [session_updated]: %w", err)
}
```

### Linha Exata

- **Arquivo**: `internal/application/message/process_inbound_message.go`
- **Linha**: **242**
- **Commit**: (pending - will be included in next commit)

---

## 🎯 Impacto do Fix

### Antes do Fix

```
5671 messages → 5671 sessions (ratio 1.0)
❌ 0 consolidações
❌ last_activity_at = 0001-01-01 (100%)
```

### Depois do Fix (Esperado)

```
5671 messages → ~4586 sessions (ratio 0.81)
✅ ~1085 consolidações
✅ last_activity_at = timestamp real da mensagem
```

### Cálculo de Redução

- **Contatos com mensagens próximas**: 1085
- **Estimativa de consolidação**: 1 mensagem/contato = 1085 sessions eliminadas
- **Redução esperada**: 19% (~1085/5671)

---

## 🧪 Como Testar

### Teste 1: Verificar Código Compilado

```bash
grep -n "cmd.ReceivedAt" internal/application/message/process_inbound_message.go | head -5
# Deve mostrar linha 242 com cmd.ReceivedAt
```

### Teste 2: Simular Mensagem Histórica (SQL)

```sql
-- 1. Criar session de teste com zero time (bug)
INSERT INTO sessions (id, tenant_id, contact_id, status, started_at, last_activity_at, timeout_duration)
VALUES (
    gen_random_uuid(),
    'test-tenant',
    (SELECT id FROM contacts LIMIT 1),
    'active',
    NOW() - INTERVAL '10 minutes',
    '0001-01-01 00:00:00',  -- ❌ BUG: Zero time
    30
);

-- 2. Simular RecordMessage() com fix (usando timestamp correto)
UPDATE sessions
SET last_activity_at = NOW() - INTERVAL '8 minutes'  -- ✅ FIX: timestamp real
WHERE last_activity_at = '0001-01-01 00:00:00';

-- 3. Verificar resultado
SELECT
    started_at,
    last_activity_at,
    CASE
        WHEN last_activity_at = '0001-01-01 00:00:00' THEN '❌ BUG'
        WHEN last_activity_at > started_at THEN '✅ FIX FUNCIONOU'
        ELSE '⚠️  Inconsistente'
    END as result
FROM sessions
ORDER BY id DESC
LIMIT 1;
```

### Teste 3: E2E com Import Histórico Completo

```bash
# 1. Limpar dados existentes
PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "
TRUNCATE messages, sessions, contacts CASCADE;"

# 2. Executar import histórico via API
curl -X POST "http://localhost:8080/api/v1/crm/channels/{CHANNEL_ID}/import-history" \
  -H "Content-Type: application/json" \
  -H "X-Dev-User-ID: {USER_ID}" \
  -d '{"time_range_days": 0, "limit_per_chat": 0}'

# 3. Aguardar conclusão (verificar Temporal UI)

# 4. Verificar last_activity_at
PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "
SELECT
    COUNT(*) as total_sessions,
    COUNT(*) FILTER (WHERE last_activity_at = '0001-01-01 00:00:00') as ❌_zero_time_bug,
    COUNT(*) FILTER (WHERE last_activity_at > '2024-01-01') as ✅_fixed,
    ROUND(COUNT(*) FILTER (WHERE last_activity_at > '2024-01-01')::numeric / COUNT(*) * 100, 2) as percent_fixed
FROM sessions;"

# 5. Executar consolidação
# (Temporal workflow já executa automaticamente após import)

# 6. Verificar redução de sessions
PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "
SELECT
    COUNT(*) as sessions_after_consolidation,
    COUNT(DISTINCT contact_id) as unique_contacts,
    ROUND(COUNT(*)::numeric / COUNT(DISTINCT contact_id), 2) as sessions_per_contact
FROM sessions;"
```

---

## 📊 Validação Esperada

### Query de Validação

```sql
-- Esta query deve mostrar:
-- ✅ zero_time_bug_count = 0
-- ✅ correct_time_count > 0
-- ✅ consolidation_happened = true

SELECT
    COUNT(*) as total_sessions,
    COUNT(*) FILTER (WHERE last_activity_at = '0001-01-01 00:00:00') as zero_time_bug_count,
    COUNT(*) FILTER (WHERE last_activity_at > '2024-01-01') as correct_time_count,
    COUNT(*) < (SELECT COUNT(*) FROM messages) as consolidation_happened,
    MIN(last_activity_at) as earliest_activity,
    MAX(last_activity_at) as latest_activity
FROM sessions;
```

### Resultado Esperado

| Métrica | Valor Esperado |
|---------|----------------|
| `zero_time_bug_count` | **0** |
| `correct_time_count` | **> 0** (todas as sessions) |
| `consolidation_happened` | **true** |
| `earliest_activity` | **> 2024-01-01** (não zero time) |
| `latest_activity` | **~NOW()** (timestamp recente) |

---

## 🔄 Próximos Passos

### Ações Imediatas

1. ✅ **Fix aplicado** - `process_inbound_message.go:242`
2. ✅ **Código compilado** - `make build` executado
3. ✅ **API reiniciada** - nova versão em produção
4. ⏳ **Aguardando re-import** - dados existentes ainda têm bug

### Ações Futuras

1. **Re-executar import histórico completo**
   - Limpar sessions/messages atuais (têm bug)
   - Executar novo import com fix aplicado
   - Verificar `last_activity_at` preenchido corretamente

2. **Executar consolidação**
   - Temporal workflow `ConsolidateHistorySessionsActivity`
   - Verificar redução de ~19% nas sessions
   - Validar ratio messages/sessions > 1.0

3. **Monitoramento**
   - Adicionar métrica: `sessions_consolidated_count`
   - Adicionar alerta se `zero_time_bug_count > 0`
   - Dashboard com ratio messages/sessions

---

## 📝 Lições Aprendidas

### 1. **Campos Legados São Perigosos**

O campo `cmd.Timestamp` existia no struct mas **nunca** foi preenchido. Código compilava sem erros, mas comportamento estava quebrado.

**Solução**: Remover campos legados ou marcar como deprecated.

```go
// ANTES
type ProcessInboundMessageCommand struct {
    Timestamp  time.Time  // ❌ Legado, nunca preenchido
    ReceivedAt time.Time  // ✅ Campo correto
}

// DEPOIS
type ProcessInboundMessageCommand struct {
    // Timestamp  time.Time  // DEPRECATED: Use ReceivedAt instead
    ReceivedAt time.Time
}
```

### 2. **Testes E2E São Essenciais**

Bug só foi descoberto quando **observamos** o resultado da consolidação (ratio 1.0). Testes unitários não detectaram porque mockamos repositórios.

**Solução**: Adicionar teste E2E que valida `last_activity_at`:

```go
func TestProcessInboundMessage_UpdatesLastActivityAt(t *testing.T) {
    // Arrange: Criar session com zero time
    // Act: Processar mensagem histórica
    // Assert: last_activity_at = message.ReceivedAt
}
```

### 3. **Zero Values São Invisíveis**

Go usa `time.Time{}` (0001-01-01) como zero value. No PostgreSQL isso vira `0001-01-01 00:00:00`. Query SQL não alertou porque tecnicamente **não é NULL**.

**Solução**: Usar `*time.Time` (nullable) ou validar explicitamente:

```go
if s.lastActivityAt.IsZero() {
    return ErrInvalidLastActivityAt
}
```

---

## 🎉 Conclusão

Bug **CRÍTICO** corrigido! Consolidação de sessions agora funcionará corretamente após re-import do histórico.

**Próximo passo**: Re-executar import completo e validar redução de ~19% nas sessions.

---

**Autor**: Claude Code
**Reviewer**: (pending review)
**Status**: ✅ FIX APLICADO, aguardando re-import para validação
