# 🗄️ Database Migrations - Ventros CRM

**Last Updated**: 2025-10-10
**Migration Strategy**: SQL files versionados + golang-migrate (Padrão da Indústria)
**Status**: ✅ 28 migrations (todas com .up.sql e .down.sql)

---

## 📋 Overview

Este projeto usa **arquivos SQL versionados** com **golang-migrate** para gerenciar schema do banco de dados.

### Por que este approach?

✅ **Padrão da Indústria** - Usado por Google, Netflix, Uber, Airbnb
✅ **Embedded no binário** - Sem dependências externas em produção
✅ **Versionamento** - Cada migrationtem número de versão sequencial
✅ **Rollback** - Cada .up.sql tem um .down.sql correspondente
✅ **Auditável** - Code review visual, compliance (SOC2, GDPR)
✅ **Type-safe** - Migrations em SQL puro, executadas por biblioteca Go
✅ **Zero-downtime** - Migrations são aplicadas automaticamente na inicialização

---

## 📁 Estrutura de Arquivos

```
infrastructure/database/
├── migrations/                    # 📂 Arquivos SQL (embedded no binário)
│   ├── 000001_initial_schema.up.sql
│   ├── 000001_initial_schema.down.sql
│   ├── 000002_add_contacts.up.sql
│   ├── 000002_add_contacts.down.sql
│   ...
│   ├── 000028_add_saga_metadata_to_outbox.up.sql
│   └── 000028_add_saga_metadata_to_outbox.down.sql
├── migration_runner.go            # 🏃 Migration runner (golang-migrate wrapper)
└── migrations.go                  # 📦 Embed directive + helpers

cmd/migrate/                       # 🛠️ CLI para migrations manuais
└── main.go
```

**Embed Directive:**
```go
//go:embed migrations/*.sql
var migrationsFS embed.FS
```

Todas as migrations são **compiladas no binário** - zero dependências externas!

---

## 🚀 Como Funciona

### 1. Desenvolvimento Local (Automático)

As migrations rodam **automaticamente** quando você inicia a API:

```bash
make api
# ou
go run cmd/api/main.go
```

**Output esperado:**
```
🔄 Applying database migrations...
✅ Migrations applied successfully (version=28)
✅ Database is up to date at version 28
```

### 2. Produção (Automático também!)

Em produção, as migrations também rodam automaticamente na inicialização da API:

```bash
./ventros-api
```

**Vantagens:**
- ✅ Zero-downtime deployments
- ✅ Migrations aplicadas antes da API aceitar requests
- ✅ Rollback automático se migrations falharem (API não sobe)
- ✅ Idempotente (safe para restart)

### 3. Manual (CLI Tool)

Para gerenciamento manual (dev, troubleshooting, recovery):

```bash
# Ver status
go run cmd/migrate/main.go status

# Aplicar todas migrations pendentes
go run cmd/migrate/main.go up

# Rollback última migration
go run cmd/migrate/main.go down

# Aplicar próximas 2 migrations
go run cmd/migrate/main.go steps 2

# Rollback 1 migration
go run cmd/migrate/main.go steps -1

# Forçar versão (recovery apenas!)
go run cmd/migrate/main.go force 28
```

---

## 🏗️ Criando Nova Migration

### Passo 1: Criar arquivos .sql

```bash
# Próximo número: 000029 (increment do último)
touch infrastructure/database/migrations/000029_add_chat_table.up.sql
touch infrastructure/database/migrations/000029_add_chat_table.down.sql
```

### Passo 2: Escrever UP migration

```sql
-- infrastructure/database/migrations/000029_add_chat_table.up.sql

-- Create chats table
CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    chat_type TEXT NOT NULL CHECK (chat_type IN ('individual', 'group', 'channel')),
    subject TEXT,
    participants JSONB NOT NULL DEFAULT '[]'::jsonb,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'closed')),
    metadata JSONB DEFAULT '{}'::jsonb,
    last_message_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_chats_project ON chats(project_id);
CREATE INDEX IF NOT EXISTS idx_chats_tenant ON chats(tenant_id);
CREATE INDEX IF NOT EXISTS idx_chats_last_message ON chats(last_message_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_chats_status ON chats(status) WHERE status != 'closed';

-- Add chat_id to messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS chat_id UUID REFERENCES chats(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_messages_chat ON messages(chat_id) WHERE chat_id IS NOT NULL;
```

### Passo 3: Escrever DOWN migration (CRITICAL!)

```sql
-- infrastructure/database/migrations/000029_add_chat_table.down.sql

-- Remove chat_id from messages
DROP INDEX IF EXISTS idx_messages_chat;
ALTER TABLE messages DROP COLUMN IF EXISTS chat_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_chats_status;
DROP INDEX IF EXISTS idx_chats_last_message;
DROP INDEX IF EXISTS idx_chats_tenant;
DROP INDEX IF EXISTS idx_chats_project;

-- Drop table
DROP TABLE IF EXISTS chats;
```

### Passo 4: Testar localmente

```bash
# Ver status atual
go run cmd/migrate/main.go status
# Version: 28

# Aplicar nova migration
go run cmd/migrate/main.go up
# Version: 29

# Testar rollback
go run cmd/migrate/main.go down
# Version: 28

# Re-aplicar
go run cmd/migrate/main.go up
# Version: 29
```

### Passo 5: Criar GORM entity (se necessário)

```go
// infrastructure/persistence/entities/chat.go
package entities

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
)

type ChatEntity struct {
    ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
    ProjectID      uuid.UUID      `gorm:"type:uuid;not null;index:idx_chats_project"`
    TenantID       string         `gorm:"type:text;not null;index:idx_chats_tenant"`
    ChatType       string         `gorm:"type:text;not null"`
    Subject        *string        `gorm:"type:text"`
    Participants   datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
    Status         string         `gorm:"type:text;not null;default:'active'"`
    Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
    LastMessageAt  *time.Time     `gorm:"index:idx_chats_last_message"`
    CreatedAt      time.Time      `gorm:"not null"`
    UpdatedAt      time.Time      `gorm:"not null"`
}

func (ChatEntity) TableName() string {
    return "chats"
}
```

**IMPORTANTE:** O GORM entity é APENAS para ORM mapping. O schema real é criado pelo SQL migration!

---

## 🔍 Migration Status & Troubleshooting

### Ver Status

```bash
go run cmd/migrate/main.go status
```

**Output esperado:**
```
📊 Migration Status
==================
Version: 28
Dirty: false
Status: ✅ Database is up to date at version 28
```

### Database "Dirty"

Se uma migration falhar no meio:

```
⚠️  WARNING: Database is in DIRTY state!
This means a migration failed mid-way.
```

**Recovery steps:**
1. Inspecionar banco e corrigir schema manualmente
2. Forçar versão: `go run cmd/migrate/main.go force 28`
3. Continuar com migrations normais

### Migration Failed

Se migration falhar, a API **não vai subir**:

```
Failed to apply database migrations: migration 29 failed: ...
```

**Fix:**
1. Corrigir o arquivo .up.sql
2. Re-iniciar API (vai tentar novamente)

---

## 🎯 Best Practices

### ✅ DO:

1. **Sempre criar .up.sql E .down.sql**
2. **Usar IF EXISTS / IF NOT EXISTS** (idempotência)
3. **Testar rollback antes de commit** (`go run cmd/migrate/main.go down`)
4. **Usar CHECK constraints** para validação de dados
5. **Criar indexes em colunas usadas em WHERE/JOIN**
6. **Usar REFERENCES para FK** (integridade referencial)
7. **Comentar migrations complexas**
8. **Commitar .sql files no Git**

### ❌ DON'T:

1. **Nunca modificar migrations já aplicadas em produção**
2. **Nunca pular rollback** (.down.sql é obrigatório!)
3. **Nunca usar DROP sem IF EXISTS** (vai falhar no rollback)
4. **Não usar GORM AutoMigrate** (só SQL migrations!)
5. **Não esquecer indexes** (performance!)

---

## 🐳 Docker & CI/CD

### Dockerfile

```dockerfile
# Build
FROM golang:1.25.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o ventros-api cmd/api/main.go

# Runtime
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ventros-api .
# Migrations estão EMBEDDED no binário!

CMD ["./ventros-api"]
```

**Migrations são embedded** - zero arquivos externos necessários!

### Docker Compose

```yaml
services:
  api:
    build: .
    environment:
      DATABASE_HOST: db
      DATABASE_NAME: ventros_crm
    # Migrations rodam automaticamente na inicialização
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ventros-api
spec:
  template:
    spec:
      containers:
      - name: api
        image: ventros-api:latest
        # Migrations rodam automaticamente no startup
        # Se falhar, pod não fica Ready (health check falha)
```

---

## 📊 Migration Versioning

### Naming Convention

```
000XXX_description.{up|down}.sql
```

- **000XXX**: Zero-padded número sequencial (000001, 000002, ...)
- **description**: Snake_case descrição
- **up**: Aplica mudança
- **down**: Reverte mudança

### Sequencing

Migrations são aplicadas em **ordem numérica**:

```
000001 → 000002 → 000003 → ... → 000028 → 000029
```

golang-migrate rastreia versão atual na tabela `schema_migrations`:

```sql
SELECT version, dirty FROM schema_migrations;
```

---

## 🔐 Production Safety

### Pre-deployment Checklist

- [ ] Migration testada localmente (up + down)
- [ ] Code review da migration SQL
- [ ] Backup do banco antes do deploy
- [ ] Migration é backward-compatible (para zero-downtime)
- [ ] Indexes criados com CONCURRENTLY (se tabela grande)
- [ ] .down.sql testado

### Backward Compatibility

Para zero-downtime deployments:

1. **Add column (nullable)** → Deploy código → Populate → Make NOT NULL
2. **Rename column** → Add new + copy data → Deploy código → Drop old
3. **Remove column** → Deploy código (stop using) → Remove column

### Performance

Para tabelas grandes:

```sql
-- ✅ Non-blocking (pode demorar, mas não trava)
CREATE INDEX CONCURRENTLY idx_messages_created ON messages(created_at);

-- ❌ Blocking (EVITAR em produção!)
CREATE INDEX idx_messages_created ON messages(created_at);
```

---

## 📚 Recursos

- **golang-migrate**: https://github.com/golang-migrate/migrate
- **Migrations Best Practices**: https://planetscale.com/blog/safely-making-database-schema-changes
- **Zero-downtime Migrations**: https://stripe.com/blog/online-migrations
- **DEV_GUIDE.md**: Guia completo de arquitetura

---

## ❓ FAQ

**P: Por que não usar GORM AutoMigrate?**
R: GORM AutoMigrate não é confiável para produção (não suporta rollback, complex alterations, triggers, etc.). SQL migrations dão controle total.

**P: Por que golang-migrate em vez de Atlas/Flyway?**
R: golang-migrate é nativo Go, tem embed support, e é o padrão no ecossistema Go. Atlas é bom mas não tem embedded migrations.

**P: Migrations rodam automaticamente?**
R: Sim! Na inicialização da API. Se falhar, API não sobe. Zero config necessário.

**P: Como fazer rollback em produção?**
R: Use o CLI tool: `go run cmd/migrate/main.go down` OU force uma versão anterior e re-deploy.

**P: E se migration falhar no meio?**
R: Database fica "dirty". Conserte manualmente e force versão: `go run cmd/migrate/main.go force <version>`

---

**Dúvidas?** Consulte DEV_GUIDE.md ou abra uma issue!
