# Task: Refatoração - SQL para Go Puro (Clean Architecture)

## Status: ✅ 80% CONCLUÍDO - Falta apenas atualizar Activity e recompilar

## Motivação

Usuário pediu "código impecável" seguindo Clean Architecture 100%, mesmo que seja 3x mais lento que SQL.
Com Kubernetes, escalabilidade horizontal resolve performance.

## O que já foi feito:

### 1. ✅ Domain Layer
**Arquivo**: `/home/caloi/ventros-crm/internal/domain/crm/session/session.go`

Adicionado método (linhas 380-420):
```go
func (s *Session) ShouldConsolidateWith(other *Session, timeout time.Duration) bool {
    // Regra de negócio: Sessions do mesmo contact com gap <= timeout devem consolidar
    // ...
}
```

### 2. ✅ Repository Interfaces

**Arquivo**: `/home/caloi/ventros-crm/internal/domain/crm/session/repository.go`

Adicionados métodos (linhas 58-62):
```go
FindByChannelPaginated(ctx context.Context, channelID uuid.UUID, limit int, offset int) ([]*Session, error)
CountByChannel(ctx context.Context, channelID uuid.UUID) (int64, error)
DeleteBatch(ctx context.Context, sessionIDs []uuid.UUID) error
```

**Arquivo**: `/home/caloi/ventros-crm/internal/domain/crm/message/repository.go`

Adicionado método (linha 51):
```go
UpdateSessionIDForSession(ctx context.Context, oldSessionID, newSessionID uuid.UUID) (int64, error)
```

### 3. ✅ Repository Implementations

**Arquivo**: `/home/caloi/ventros-crm/infrastructure/persistence/gorm_session_repository.go`

Implementados métodos (linhas 330-385):
- `FindByChannelPaginated` - Busca sessions ordenadas por contact_id, started_at
- `CountByChannel` - Conta sessions de um canal
- `DeleteBatch` - Deleta múltiplas sessions (para orphans)

**Arquivo**: `/home/caloi/ventros-crm/infrastructure/persistence/gorm_message_repository.go`

Implementado método (linhas 302-317):
- `UpdateSessionIDForSession` - Move mensagens de uma session para outra

### 4. ✅ Application Layer (Use Case)

**Arquivo**: `/home/caloi/ventros-crm/internal/application/session/consolidate_sessions_usecase.go` (NOVO - 268 linhas)

Use case completo com:
- Processamento em batches (controle de memória)
- Lógica de consolidação usando `Session.ShouldConsolidateWith()` (domínio)
- Atualização de mensagens via repository
- Deleção de sessions órfãs
- Logging detalhado

## O que falta fazer:

### Passo 1: Atualizar Activity

**Arquivo**: `/home/caloi/ventros-crm/internal/workflows/channel/waha_history_import_activities.go`

**Adicionar import** (linha ~12):
```go
sessionapp "github.com/ventros/crm/internal/application/session"
```

**Substituir método** `ConsolidateHistorySessionsActivity` (linhas 718-929):

**ANTES** (211 linhas de SQL):
```go
func (a *WAHAHistoryImportActivities) ConsolidateHistorySessionsActivity(...) {
	// SQL window functions...
	consolidationSQL := `WITH message_groups AS ...` // 80+ linhas SQL
	// Mais SQL para update, delete, stats...
}
```

**DEPOIS** (30 linhas Go puro):
```go
func (a *WAHAHistoryImportActivities) ConsolidateHistorySessionsActivity(ctx context.Context, input ConsolidateHistorySessionsActivityInput) (*ConsolidateHistorySessionsActivityResult, error) {
	a.logger.Info("🔄 Starting session consolidation (Go pure - Clean Architecture)",
		zap.String("channel_id", input.ChannelID),
		zap.Int("timeout_minutes", input.SessionTimeoutMinutes))

	channelID, err := uuid.Parse(input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel_id: %w", err)
	}

	// ✅ Create use case with injected repositories
	consolidateUC := sessionapp.NewConsolidateSessionsUseCase(
		a.sessionRepo,
		a.messageRepo,
		a.logger,
	)

	// ✅ Execute consolidation using pure domain logic
	consolidateInput := sessionapp.ConsolidateInput{
		ChannelID:             channelID,
		SessionTimeoutMinutes: input.SessionTimeoutMinutes,
		BatchSize:             5000, // Process 5k sessions per batch to control memory
	}

	result, err := consolidateUC.Execute(ctx, consolidateInput)
	if err != nil {
		return nil, fmt.Errorf("consolidation failed: %w", err)
	}

	// ✅ Convert result
	return &ConsolidateHistorySessionsActivityResult{
		ChannelID:       input.ChannelID,
		SessionsBefore:  int(result.SessionsBefore),
		SessionsAfter:   int(result.SessionsAfter),
		SessionsDeleted: result.SessionsDeleted,
		MessagesUpdated: result.MessagesUpdated,
	}, nil
}
```

### Passo 2: Remover dependência `db *gorm.DB`

Não é mais necessário, pois toda lógica está em repositories.

**Arquivo**: `/home/caloi/ventros-crm/internal/workflows/channel/waha_history_import_activities.go`
- **Linha 31**: Remover campo `db *gorm.DB`
- **Linha 44**: Remover parâmetro `db *gorm.DB`
- **Linha ~55**: Remover assignment `db: db,`

**Arquivo**: `/home/caloi/ventros-crm/internal/workflows/channel/waha_import_worker.go`
- **Linha 37**: Remover parâmetro `db *gorm.DB`
- **Linha ~50**: Remover argumento `db` ao criar activities

**Arquivo**: `/home/caloi/ventros-crm/cmd/api/main.go`
- **Linha ~604**: Remover argumento `gormDB` ao criar worker

### Passo 3: Recompilar e testar

```bash
# Recompilar
make build

# Matar API antiga
pkill -f crm-api

# Iniciar nova API
./bin/crm-api

# Testar E2E (em outro terminal)
bash /tmp/test_consolidation_e2e.sh
```

## Verificação de Sucesso

✅ **Código compila sem erros**
✅ **Teste E2E mostra consolidação funcionando** (ratio > 1.5 msgs/session)
✅ **Logs mostram "Go pure implementation"**
✅ **Nenhuma query SQL direto na activity** (apenas via repositories)

## Performance Esperada

| Métrica | SQL | Go Batches | Diferença |
|---------|-----|------------|-----------|
| 100k msgs | 2-5s | 5-15s | 3x mais lento |
| 500k msgs | 10-25s | 25-75s | 3x mais lento |
| Memória | Constante | Constante | Mesma (batches) |
| Escalabilidade | Vertical | Horizontal (K8s) | **MELHOR** |
| Testabilidade | Difícil | Fácil (mocks) | **MELHOR** |
| Manutenibilidade | Obscuro | Explícito | **MELHOR** |

## Benefícios Arquiteturais

1. ✅ **Clean Architecture 100%**: Lógica de negócio no domínio
2. ✅ **SOLID**: Separação de responsabilidades perfeita
3. ✅ **DDD**: Business rules explícitos (`ShouldConsolidateWith`)
4. ✅ **Testável**: Testes unitários puros (sem DB)
5. ✅ **Manutenível**: Código Go explícito vs SQL obscuro
6. ✅ **Escalável**: Kubernetes resolve performance

## Trade-offs Aceitos

- ❌ **3x mais lento** que SQL (mas ainda ~10s para 100k mensagens)
- ✅ **Código impecável** (prioridade do usuário)
- ✅ **Escalável horizontalmente** com K8s

---

**Próxima ação**: Aplicar Passo 1 (atualizar Activity) + Passo 2 (remover db) + Passo 3 (recompilar/testar)
