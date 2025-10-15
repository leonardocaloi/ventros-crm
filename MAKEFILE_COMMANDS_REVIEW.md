# 🔧 MAKEFILE COMMANDS REVIEW

> **Decisões necessárias**: Quais comandos manter, quais novos, novos nomes

**Data**: 2025-10-15
**Status**: AGUARDANDO REVISÃO

---

## 📊 COMANDOS ATUAIS (Makefile)

### ✅ Comandos GO

```makefile
build                 ## Build API binary
build-linux           ## Build for Linux (Docker)
run-binary            ## Build and run binary
clean-bin             ## Remove binary
```

**DECISÃO NECESSÁRIA**:
- [ ] Manter como está?
- [ ] Mudar para subcomandos: `make go.build`, `make go.run`, `make go.clean`?
- [ ] Novos comandos GO necessários?

---

### ✅ Comandos INFRA

```makefile
infra                 ## Start infrastructure
infra-logs            ## Show logs
infra-stop            ## Stop (keep data)
infra-clean           ## Stop + remove volumes
```

**DECISÃO NECESSÁRIA**:
- [ ] Manter como está?
- [ ] Mudar para: `make infra.start`, `make infra.logs`, `make infra.stop`, `make infra.clean`?
- [ ] Adicionar: `make infra.restart`, `make infra.status`?

---

### ✅ Comandos RESET/FRESH

```makefile
fresh                 ## ✨ Fresh start (fast dev)
reset-full            ## 🔥 Full reset (dev)
run-binary-full       ## 🔥 Full reset (test prod)
```

**DECISÃO NECESSÁRIA**:
- [ ] Manter nomes atuais?
- [ ] Mudar para: `make dev.fresh`, `make dev.reset`, `make prod.reset`?
- [ ] Simplificar para apenas: `make reset` e `make reset.full`?

---

### ✅ Comandos API

```makefile
api                   ## Run API locally
swagger               ## Generate Swagger docs
```

**DECISÃO NECESSÁRIA**:
- [ ] Manter como está?
- [ ] Mudar para: `make go.api`, `make go.swagger`?
- [ ] Adicionar: `make api.dev` (com hot reload)?

---

### ⚠️ Comandos CONTAINER

```makefile
container             ## Start EVERYTHING containerized
container-logs        ## Show logs
container-stop        ## Stop (keep data)
container-down        ## Stop + remove
```

**DECISÃO NECESSÁRIA**:
- [ ] Manter como está?
- [ ] Mudar para: `make docker.up`, `make docker.logs`, `make docker.stop`, `make docker.down`?
- [ ] Ou: `make deploy.local`, `make deploy.stop`?

---

### 🧪 Comandos TESTE (CRÍTICO - Usar discover.sh)

**ATUAL**:
```makefile
test                  ## Run all tests
test-unit             ## Unit tests only
test-integration      ## Integration tests
test-import           ## Full import test
test-bench            ## Benchmark tests
test-coverage         ## Coverage report
test-coverage-unit    ## Unit coverage
clean-coverage        ## Clean coverage
```

**PROPOSTA INTELIGENTE** (com `tests/scripts/discover.sh`):

```makefile
# Descoberta automática de testes
test.discover         ## List all available tests
test.stats            ## Show test statistics

# Testes gerais
test                  ## Run all tests
test.unit             ## All unit tests
test.integration      ## All integration tests
test.e2e              ## All e2e tests

# Testes específicos (auto-descobertos)
test.unit.domain      ## Unit tests: internal/domain/*
test.unit.application ## Unit tests: internal/application/*
test.unit.infra       ## Unit tests: infrastructure/*

test.integration.waha ## Integration: WAHA
test.integration.db   ## Integration: Database
test.integration.mq   ## Integration: RabbitMQ

test.e2e.waha         ## E2E: WAHA flow
test.e2e.campaign     ## E2E: Campaign flow
test.e2e.sequence     ## E2E: Sequence flow

# Cobertura
test.coverage         ## Coverage report (all)
test.coverage.unit    ## Coverage (unit only)
test.coverage.html    ## Coverage HTML report

# Benchmarks
test.bench            ## Run benchmarks
test.bench.domain     ## Benchmarks: domain
```

**DECISÃO NECESSÁRIA**:
- [ ] Usar estrutura proposta?
- [ ] Quais testes E2E específicos você quer (além de waha)?
- [ ] Precisamos de `test.watch` (hot reload)?

---

### 🔧 Comandos CODE QUALITY

```makefile
fmt                   ## Format code
lint                  ## Run golangci-lint
vet                   ## Run go vet
mod-tidy              ## Clean go.mod
```

**DECISÃO NECESSÁRIA**:
- [ ] Manter como está?
- [ ] Mudar para: `make quality.fmt`, `make quality.lint`, etc?
- [ ] Adicionar: `make quality.all` (fmt + lint + vet)?

---

## 🆕 NOVOS COMANDOS PROPOSTOS

### Análise (Claude)

```makefile
analyze               ## Run deterministic analysis
analyze.quick         ## Quick analysis (bash)
analyze.deep          ## Deep analysis (Go AST)
analyze.security      ## Security analysis only
```

### Database

```makefile
db.migrate.up         ## Apply migrations
db.migrate.down       ## Rollback migration
db.migrate.status     ## Migration status
db.migrate.create     ## Create new migration
db.seed               ## Seed database
db.reset              ## Reset database
```

### Deploy

```makefile
deploy.dev            ## Deploy to dev environment
deploy.staging        ## Deploy to staging
deploy.prod           ## Deploy to production
deploy.status         ## Check deployment status
deploy.logs           ## Show deployment logs
```

### Monitoring

```makefile
monitor.logs          ## Tail logs
monitor.metrics       ## Show metrics
monitor.health        ## Health check
```

---

## ❓ PERGUNTAS PARA VOCÊ

### 1. Subcomandos vs Hífens?

**Opção A - Subcomandos** (recomendado):
```makefile
make go.build
make test.unit.domain
make infra.start
make deploy.prod
```

**Opção B - Hífens** (atual):
```makefile
make build
make test-unit-domain
make infra
make deploy-prod
```

**SUA DECISÃO**: [ ] A  [ ] B  [ ] Misto (alguns com subcomandos, outros não)

---

### 2. Comandos de Teste - Quais E2E?

Além de `test.e2e.waha`, quais outros E2E você quer?

**Sugestões**:
- [ ] `test.e2e.campaign` - Fluxo completo de campanha
- [ ] `test.e2e.sequence` - Fluxo de sequência (drip campaign)
- [ ] `test.e2e.broadcast` - Broadcast de mensagens
- [ ] `test.e2e.pipeline` - Movimentação em pipeline
- [ ] `test.e2e.chat` - Chat em grupo
- [ ] `test.e2e.webhook` - Webhooks externos
- [ ] Outros: __________________

---

### 3. Comandos GO - Quais manter/adicionar?

**Manter**:
- [ ] `make build` (ou `make go.build`)
- [ ] `make api` (ou `make go.api`)
- [ ] `make run-binary` (ou `make go.run.binary`)

**Adicionar**:
- [ ] `make go.test` - Alias para test.unit
- [ ] `make go.watch` - Hot reload (air/gow)
- [ ] `make go.mod.tidy` - go mod tidy
- [ ] `make go.vendor` - go mod vendor
- [ ] Outros: __________________

---

### 4. Database - Quais comandos?

- [ ] `make db.migrate.up`
- [ ] `make db.migrate.down`
- [ ] `make db.migrate.create NAME=create_users`
- [ ] `make db.seed`
- [ ] `make db.reset` (drop + create + migrate + seed)
- [ ] `make db.console` (psql)
- [ ] `make db.backup`
- [ ] `make db.restore`
- [ ] Outros: __________________

---

### 5. Deploy - Necessário?

- [ ] Sim, adicionar comandos de deploy
- [ ] Não, deploy é via CI/CD apenas
- [ ] Apenas para desenvolvimento local (Docker Compose)

Se SIM, quais ambientes?
- [ ] dev
- [ ] staging
- [ ] production
- [ ] Outros: __________________

---

### 6. Monitoring - Necessário?

- [ ] Sim, adicionar comandos de monitoring
- [ ] Não, usar ferramentas externas apenas

Se SIM, quais comandos?
- [ ] `make logs` - Tail logs
- [ ] `make metrics` - Prometheus metrics
- [ ] `make health` - Health check
- [ ] `make trace` - Distributed tracing
- [ ] Outros: __________________

---

## 📝 COMANDOS QUE VÃO SAIR

Estes comandos serão **removidos** ou **renomeados**:

### Remover
- [ ] `fresh` → substituído por `dev.fresh` ou `reset`
- [ ] `reset-full` → substituído por `reset.full` ou `dev.reset.full`
- [ ] `run-binary-full` → substituído por `prod.reset` ou `go.run.prod`

### Renomear
- [ ] `infra` → `infra.start`
- [ ] `api` → `go.api` ou manter como `api`
- [ ] `test-unit` → `test.unit`
- [ ] `test-integration` → `test.integration`

**Confirmação necessária**!

---

## 🎯 ESTRUTURA FINAL PROPOSTA

```makefile
# ============================================
# GO Commands
# ============================================
go.build              ## Build API binary
go.build.linux        ## Build for Linux
go.run                ## Run API (go run)
go.run.binary         ## Run binary
go.clean              ## Clean binaries
go.api                ## Run API with swagger (alias)

# ============================================
# Infrastructure
# ============================================
infra.start           ## Start infrastructure
infra.stop            ## Stop infrastructure
infra.clean           ## Clean volumes
infra.logs            ## Show logs
infra.restart         ## Restart infrastructure

# ============================================
# Database
# ============================================
db.migrate.up         ## Apply migrations
db.migrate.down       ## Rollback
db.migrate.status     ## Status
db.reset              ## Reset database

# ============================================
# Tests (Auto-discovery)
# ============================================
test                  ## Run all tests
test.unit             ## Unit tests
test.integration      ## Integration tests
test.e2e              ## E2E tests

test.unit.domain      ## Unit: domain layer
test.unit.application ## Unit: application layer
test.e2e.waha         ## E2E: WAHA integration

test.coverage         ## Coverage report
test.bench            ## Benchmarks
test.discover         ## List available tests

# ============================================
# Code Quality
# ============================================
quality.fmt           ## Format code
quality.lint          ## Lint code
quality.vet           ## Go vet
quality.all           ## All quality checks

# ============================================
# Analysis (Claude)
# ============================================
analyze               ## Run analysis
analyze.quick         ## Quick (bash)
analyze.deep          ## Deep (Go AST)

# ============================================
# Development
# ============================================
dev.fresh             ## Fresh start
dev.reset             ## Full reset
dev.watch             ## Hot reload

# ============================================
# Docker
# ============================================
docker.up             ## Start containers
docker.down           ## Stop containers
docker.logs           ## Show logs
```

**Você aprova esta estrutura?** [ ] Sim  [ ] Não  [ ] Modificar

---

## ✅ PRÓXIMOS PASSOS

Depois das suas decisões:

1. **Atualizar Makefile** com comandos aprovados
2. **Criar scripts** necessários (go/, db/, deploy/, etc)
3. **Atualizar agente** `crm_docs_makefile_updater`
4. **Gerar MAKEFILE.md** com documentação completa
5. **Testar todos os comandos**

---

**AGUARDANDO SUA REVISÃO!**

Marque suas escolhas e me avise quando estiver pronto.
