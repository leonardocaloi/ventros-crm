# Tarefa em Andamento: Correção de Import de Histórico WAHA

## Status Atual (2025-10-15 00:50)

### ✅ Problema 1: tenant_id vazio - RESOLVIDO
**Sintoma**: Mensagens sendo inseridas com `tenant_id = ''` (vazio)

**Causa Raiz**: `gorm_message_repository.go` linha 291 - função `domainToEntity()` não extraía tenant_id do contexto saga

**Solução Implementada**:
```go
// infrastructure/persistence/gorm_message_repository.go:291-296
func (r *GormMessageRepository) domainToEntity(ctx context.Context, m *message.Message) *entities.MessageEntity {
	// ✅ Extract tenant_id from saga context
	tenantID, _ := saga.GetTenantID(ctx)

	entity := &entities.MessageEntity{
		TenantID: tenantID, // ✅ Extracted from saga context
		// ... rest of fields
	}
}
```

**Modificações**:
1. Adicionado import: `"github.com/ventros/crm/internal/domain/core/saga"`
2. Modificado `Save()` para passar context: `entity := r.domainToEntity(ctx, m)`
3. Modificado `domainToEntity()` para aceitar context e extrair tenant_id

**Resultado**: ✅ tenant_id agora é setado corretamente (ex: `'user-1d98bbd3'`)

---

### ⚠️ Problema 2: Foreign Key Violation - EM ANDAMENTO

**Sintoma**:
```
ERROR: insert or update on table "messages" violates foreign key constraint "fk_sessions_messages" (SQLSTATE 23503)
```

**Evidências dos Logs**:
```sql
-- Session INSERT (sucesso)
INSERT INTO "sessions" (...,"id") VALUES (...,'99999781-0a2a-417e-b3bc-d478b62b08b3')

-- Message INSERT (falha) - referenciando session que não existe no banco
INSERT INTO "messages" (...,"session_id"...) VALUES (...,'4436300e-2fc5-4470-87f6-26db230aa5c9')
```

**Causa Raiz Identificada**:
1. Session E Message estão na MESMA transação
2. Message INSERT falha por algum motivo
3. **Transaction ROLLBACK** → Session também é desfeito
4. Sessions não são commitadas no banco (verificado: 0 rows com tenant_id correto)

**Próximos Passos**:
1. Investigar `internal/application/message/process_inbound_message.go` para entender transações
2. Garantir que Session seja commitada ANTES de criar Message
3. Opções:
   - Separar em 2 transações (commit session, depois message)
   - Garantir ordem correta de persistência dentro da mesma transação
   - Verificar se há rollback explícito no código

**Arquivos Relevantes**:
- `internal/application/message/process_inbound_message.go` (use case principal)
- `infrastructure/persistence/gorm_session_repository.go` (linha 35-104: Save com transaction)
- `infrastructure/persistence/gorm_message_repository.go` (linha 23-26: Save)
- `internal/workflows/channel/waha_history_import_activities.go` (linha 556: chamada do use case)

**Teste Realizado**:
- Script: `/tmp/test_tenant_id_fix.sh`
- Resultado: 0 mensagens importadas, 866 erros de FK
- Banco: 0 sessions com tenant_id = 'user-1d98bbd3'

---

## Próxima Ação

Ler `process_inbound_message.go` linha 200-250 para entender gerenciamento de transações e ordem de persistência.
