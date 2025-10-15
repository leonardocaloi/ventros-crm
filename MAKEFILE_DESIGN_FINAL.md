# 🎯 MAKEFILE DESIGN FINAL - Ventros CRM

> **Design elegante e padronizado para comandos Make**
>
> **Data**: 2025-10-15
> **Status**: PROPOSTA FINAL

---

## 📋 ARQUITETURA CI/CD (Existente)

```
┌──────────────────────────────────────────────────────────────┐
│                    DESENVOLVIMENTO LOCAL                      │
│                                                                │
│  make test → make build → make docker.build → make helm.package│
└─────────────────────────┬──────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────────┐
│               GITHUB ACTIONS (CI)                             │
│                                                                │
│  1. Run tests                                                 │
│  2. Build & push Docker image                                 │
│  3. Package & push Helm chart                                 │
│  4. Trigger AWX via API                                       │
└─────────────────────────┬──────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────────┐
│                  AWX + ANSIBLE (CD)                           │
│                                                                │
│  1. helm repo update                                          │
│  2. helm upgrade --install                                    │
│  3. wait readiness                                            │
│  4. smoke tests                                               │
│  5. backup/restore (Zalando Postgres Operator)                │
└──────────────────────────────────────────────────────────────┘
```

---

## 🎨 PADRÃO DE NOMENCLATURA

### Hierarquia
```
{categoria}.{ação}[.modificador]
```

### Exemplos
```makefile
infra.up                # Categoria: infra, Ação: up
crm.run.force           # Categoria: crm, Ação: run, Modificador: force
test.unit.domain        # Categoria: test, Ação: unit, Modificador: domain
```

### Categorias
- `infra.*` - Infraestrutura (Postgres, RabbitMQ, Redis, Temporal, Keycloak)
- `crm.*` - Aplicação CRM (Go API)
- `test.*` - Testes
- `db.*` - Database (migrations, seed, backup)
- `docker.*` - Docker (build, push)
- `helm.*` - Helm (package, push, deploy)
- `k8s.*` - Kubernetes (minikube, deploy)
- `deploy.*` - Deploy (dev, staging, prod)
- `quality.*` - Code quality (fmt, lint, vet)
- `analyze.*` - Análise (Claude)

---

## 📦 COMANDOS PROPOSTOS

### 1. INFRA (Infraestrutura)

```makefile
# ============================================
# INFRA - Infrastructure Management
# ============================================

infra.up              ## Start infrastructure (Postgres, RabbitMQ, Redis, Temporal, Keycloak)
                      # docker-compose -f docker-compose.infra.yml up -d

infra.down            ## Stop infrastructure (keep volumes)
                      # docker-compose -f docker-compose.infra.yml down

infra.delete          ## Delete all volumes (DESTRUCTIVE)
                      # docker-compose -f docker-compose.infra.yml down -v

infra.logs            ## Show infrastructure logs
                      # docker-compose -f docker-compose.infra.yml logs -f

infra.restart         ## Restart infrastructure
                      # make infra.down && make infra.up

infra.status          ## Show infrastructure status
                      # docker-compose -f docker-compose.infra.yml ps
```

**Implementação**:
- Script: `scripts/make/infra/up.sh`, `down.sh`, `delete.sh`
- Docker Compose: `docker-compose.infra.yml` (apenas infra, sem CRM)

---

### 2. CRM (Aplicação Go)

```makefile
# ============================================
# CRM - Go Application Management
# ============================================

## --- Run (go run) ---

crm.run               ## Run CRM (go run cmd/api/main.go)
                      # Requires: make infra.up

crm.run.force         ## Kill port 8080 and run CRM
                      # lsof -ti:8080 | xargs kill -9 || true
                      # go run cmd/api/main.go

## --- Binary ---

crm.build             ## Build CRM binary (bin/crm-api)
                      # go build -o bin/crm-api cmd/api/main.go

crm.run.binary        ## Run CRM binary
                      # ./bin/crm-api

crm.run.binary.force  ## Rebuild + run binary
                      # make crm.build && make crm.run.binary.force.kill
                      # ./bin/crm-api

## --- Combined (Infra + CRM) ---

crm.infra.up          ## Start infra + run CRM (force)
                      # make infra.up && make crm.run.force

crm.infra.up.reset    ## Delete infra + start + run CRM
                      # make infra.delete && make infra.up && make crm.run.force

## --- Container ---

crm.container.build   ## Build Docker image (ventros-crm:latest)
                      # docker build -t ventros-crm:latest .

crm.container.run     ## Run Docker container (builds if needed)
                      # if ! docker images | grep -q ventros-crm; then
                      #   make crm.container.build
                      # fi
                      # docker run -p 8080:8080 ventros-crm:latest

crm.container.run.force  ## Rebuild + run container
                      # make crm.container.build
                      # docker rm -f ventros-crm || true
                      # docker run -p 8080:8080 --name ventros-crm ventros-crm:latest

## --- Docker Compose ---

crm.compose.up        ## Start infra + CRM via Docker Compose
                      # docker-compose up -d

crm.compose.up.reset  ## Rebuild + delete infra + compose up
                      # make infra.delete
                      # docker-compose up --build -d

crm.compose.down      ## Stop Docker Compose
                      # docker-compose down

crm.compose.logs      ## Show Docker Compose logs
                      # docker-compose logs -f
```

**Implementação**:
- Scripts: `scripts/make/go/run.sh`, `build.sh`, `kill-port.sh`
- Docker: `Dockerfile`, `docker-compose.yml` (full stack)
- Docker Compose Infra: `docker-compose.infra.yml` (apenas infra)

---

### 3. TESTS (Descoberta Inteligente)

```makefile
# ============================================
# TEST - Intelligent Test Discovery
# ============================================

## --- Discovery ---

test.discover         ## List all available tests
                      # ./tests/scripts/discover.sh stats

test.stats            ## Show test statistics
                      # ./tests/scripts/discover.sh stats all

## --- General ---

test                  ## Run all tests (unit + integration + e2e)
                      # make test.unit && make test.integration && make test.e2e

test.unit             ## Run all unit tests (fast, no dependencies)
                      # go test -v -short ./internal/... -count=1

test.integration      ## Run all integration tests (requires: infra.up)
                      # go test -v ./tests/integration/... -count=1

test.e2e              ## Run all E2E tests (requires: infra.up + crm.run)
                      # go test -v ./tests/e2e/... -count=1

## --- Unit Tests (Auto-discovered) ---

test.unit.domain      ## Unit tests: internal/domain/*
                      # go test -v ./internal/domain/... -count=1

test.unit.application ## Unit tests: internal/application/*
                      # go test -v ./internal/application/... -count=1

test.unit.infra       ## Unit tests: infrastructure/*
                      # go test -v ./infrastructure/... -count=1 -short

## --- Integration Tests ---

test.integration.waha ## Integration: WAHA
                      # go test -v ./tests/integration/waha/... -count=1

test.integration.db   ## Integration: Database
                      # go test -v ./tests/integration/db/... -count=1

test.integration.mq   ## Integration: RabbitMQ
                      # go test -v ./tests/integration/mq/... -count=1

## --- E2E Tests (Specific Flows) ---

test.e2e.waha         ## E2E: WAHA integration flow
                      # ./tests/scripts/test-e2e-waha.sh

test.e2e.campaign     ## E2E: Campaign flow
                      # go test -v ./tests/e2e/campaign/... -count=1

test.e2e.sequence     ## E2E: Sequence flow
                      # go test -v ./tests/e2e/sequence/... -count=1

test.e2e.broadcast    ## E2E: Broadcast flow
                      # go test -v ./tests/e2e/broadcast/... -count=1

test.e2e.pipeline     ## E2E: Pipeline flow
                      # go test -v ./tests/e2e/pipeline/... -count=1

## --- Coverage ---

test.coverage         ## Coverage report (all tests)
                      # go test -v -coverprofile=coverage.out ./...
                      # go tool cover -html=coverage.out -o coverage.html

test.coverage.unit    ## Coverage (unit tests only)
                      # go test -v -coverprofile=coverage.unit.out ./internal/... -short
                      # go tool cover -html=coverage.unit.out -o coverage.unit.html

test.coverage.html    ## Open coverage HTML report
                      # open coverage.html  # macOS
                      # xdg-open coverage.html  # Linux

## --- Benchmarks ---

test.bench            ## Run all benchmarks
                      # go test -bench=. -benchmem ./...

test.bench.domain     ## Benchmarks: domain layer
                      # go test -bench=. -benchmem ./internal/domain/...
```

**Implementação**:
- Script: `tests/scripts/discover.sh` (já criado!)
- Tests: Organizados em `tests/{integration,e2e}/`
- Unit tests: Em cada package (`*_test.go`)

---

### 4. DATABASE (Migrations)

```makefile
# ============================================
# DB - Database Management
# ============================================

## --- Migrations (Centralized - Used by CI/CD) ---

db.migrate.up         ## Apply migrations (PRODUCTION)
                      # migrate -path infrastructure/database/migrations \
                      #         -database "postgres://..." up

db.migrate.down       ## Rollback last migration
                      # migrate -path infrastructure/database/migrations \
                      #         -database "postgres://..." down 1

db.migrate.status     ## Show migration status
                      # migrate -path infrastructure/database/migrations \
                      #         -database "postgres://..." version

db.migrate.create     ## Create new migration (Usage: make db.migrate.create NAME=add_users)
                      # migrate create -ext sql -dir infrastructure/database/migrations \
                      #         -seq $(NAME)

## --- Development ---

db.seed               ## Seed database with test data
                      # go run cmd/seed/main.go

db.reset              ## Reset database (drop + migrate + seed)
                      # make db.drop && make db.migrate.up && make db.seed

db.console            ## Open PostgreSQL console
                      # PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm

## --- Backup/Restore (Local Dev) ---

db.backup             ## Backup database to file
                      # pg_dump -h localhost -U ventros ventros_crm > backup.sql

db.restore            ## Restore database from file
                      # psql -h localhost -U ventros ventros_crm < backup.sql
```

**Implementação**:
- Scripts: `scripts/make/db/migrate.sh`, `seed.sh`, `reset.sh`
- Migrations: `infrastructure/database/migrations/*.sql`
- Seed: `cmd/seed/main.go` (opcional)

**IMPORTANTE**:
- ✅ `db.migrate.*` é centralizado no Makefile
- ✅ GitHub Actions usa `make db.migrate.up`
- ✅ AWX/Ansible usa `make db.migrate.up` via Makefile

---

### 5. DOCKER (Build & Push)

```makefile
# ============================================
# DOCKER - Image Management
# ============================================

docker.build          ## Build Docker image (ventros-crm:latest)
                      # docker build -t ventros-crm:latest .

docker.build.tag      ## Build with custom tag (Usage: make docker.build.tag TAG=v1.2.3)
                      # docker build -t ventros-crm:$(TAG) .

docker.push           ## Push to registry (requires: login)
                      # docker tag ventros-crm:latest ghcr.io/ventros/crm:latest
                      # docker push ghcr.io/ventros/crm:latest

docker.push.tag       ## Push specific tag (Usage: make docker.push.tag TAG=v1.2.3)
                      # docker tag ventros-crm:$(TAG) ghcr.io/ventros/crm:$(TAG)
                      # docker push ghcr.io/ventros/crm:$(TAG)

docker.login          ## Login to GitHub Container Registry
                      # echo "$GITHUB_TOKEN" | docker login ghcr.io -u USERNAME --password-stdin
```

**Implementação**:
- Script: `scripts/make/docker/build.sh`, `push.sh`
- Registry: GitHub Container Registry (ghcr.io)

---

### 6. HELM (Package & Deploy)

```makefile
# ============================================
# HELM - Chart Management
# ============================================

## --- Package ---

helm.package          ## Package Helm chart
                      # helm package .deploy/helm/ventros-crm -d .deploy/helm/

helm.push             ## Push chart to registry
                      # helm push .deploy/helm/ventros-crm-*.tgz oci://ghcr.io/ventros/charts

## --- Local Deploy (Minikube/Kind) ---

helm.install.dev      ## Install chart to dev namespace
                      # helm install ventros-crm .deploy/helm/ventros-crm \
                      #   -n dev --create-namespace -f .deploy/helm/values-dev.yaml

helm.upgrade.dev      ## Upgrade chart in dev
                      # helm upgrade ventros-crm .deploy/helm/ventros-crm \
                      #   -n dev -f .deploy/helm/values-dev.yaml

helm.uninstall.dev    ## Uninstall chart from dev
                      # helm uninstall ventros-crm -n dev

## --- CI/CD (Staging/Prod via AWX) ---
# Nota: Staging/Prod são deployados via AWX/Ansible
# GitHub Actions apenas faz: make helm.push
# AWX então faz: helm upgrade --install
```

**Implementação**:
- Chart: `.deploy/helm/ventros-crm/`
- Values: `values-dev.yaml`, `values-staging.yaml`, `values-prod.yaml`
- Scripts: `scripts/make/helm/package.sh`, `push.sh`

---

### 7. KUBERNETES (Minikube)

```makefile
# ============================================
# K8S - Kubernetes Management (Minikube)
# ============================================

## --- Minikube ---

k8s.minikube.start    ## Start Minikube
                      # minikube start --cpus=4 --memory=8192

k8s.minikube.stop     ## Stop Minikube
                      # minikube stop

k8s.minikube.delete   ## Delete Minikube cluster
                      # minikube delete

k8s.minikube.dashboard ## Open Minikube dashboard
                      # minikube dashboard

## --- Deploy to Minikube (Full Flow) ---

k8s.deploy.minikube   ## Full deploy to Minikube (build → push → helm)
                      # Pre-requisites:
                      # 1. Build image: make docker.build.tag TAG=minikube
                      # 2. Load image: minikube image load ventros-crm:minikube
                      # 3. Deploy helm: make helm.install.dev
                      #
                      # Script: ./scripts/make/k8s/deploy-minikube.sh

k8s.deploy.minikube.reset ## Rebuild + redeploy to Minikube
                      # make docker.build.tag TAG=minikube
                      # minikube image rm ventros-crm:minikube || true
                      # minikube image load ventros-crm:minikube
                      # helm uninstall ventros-crm -n dev || true
                      # make helm.install.dev
```

**Implementação**:
- Script: `scripts/make/k8s/deploy-minikube.sh` (full flow)
- Usa Helm values: `.deploy/helm/values-dev.yaml`

---

### 8. DEPLOY (Ambientes)

```makefile
# ============================================
# DEPLOY - Environment Deployment
# ============================================

## --- Development ---

deploy.dev            ## Deploy to development (local Docker Compose)
                      # make crm.compose.up

deploy.dev.reset      ## Reset + deploy to development
                      # make crm.compose.up.reset

## --- Staging (via AWX) ---

deploy.staging        ## Deploy to staging (triggers AWX)
                      # curl -X POST https://awx.domain.com/api/v2/job_templates/ID/launch/ \
                      #   -H "Authorization: Bearer $AWX_TOKEN" \
                      #   -d '{"extra_vars": "{\"environment\": \"staging\"}"}'

deploy.staging.full   ## Full deploy to staging (DB refresh + deploy)
                      # Same as above, but extra_vars: {"db_refresh": true}

## --- Production (via AWX with approval) ---

deploy.prod           ## Deploy to production (triggers AWX with approval)
                      # curl -X POST https://awx.domain.com/api/v2/job_templates/ID/launch/ \
                      #   -H "Authorization: Bearer $AWX_TOKEN" \
                      #   -d '{"extra_vars": "{\"environment\": \"production\"}"}'
                      # Requires manual approval in AWX

## --- Rollback ---

deploy.rollback.staging ## Rollback staging to previous version
                      # helm rollback ventros-crm -n staging

deploy.rollback.prod  ## Rollback production to previous version
                      # helm rollback ventros-crm -n production
```

**Implementação**:
- Scripts: `scripts/make/deploy/staging.sh`, `prod.sh`
- AWX: Playbooks em `.deploy/ansible/`
- Helm: Charts em `.deploy/helm/`

**IMPORTANTE**:
- ✅ Dev: Docker Compose local
- ✅ Staging/Prod: AWX + Ansible + Helm
- ✅ CI/CD: GitHub Actions → make → AWX API

---

### 9. QUALITY (Code Quality)

```makefile
# ============================================
# QUALITY - Code Quality Checks
# ============================================

quality.fmt           ## Format code (gofmt + goimports)
                      # gofmt -w .
                      # goimports -w .

quality.lint          ## Run golangci-lint
                      # golangci-lint run ./...

quality.vet           ## Run go vet
                      # go vet ./...

quality.all           ## Run all quality checks
                      # make quality.fmt && make quality.lint && make quality.vet

quality.mod.tidy      ## Clean go.mod and go.sum
                      # go mod tidy
```

**Implementação**:
- Scripts: `scripts/make/quality/fmt.sh`, `lint.sh`, `vet.sh`

---

### 10. ANALYZE (Claude Code)

```makefile
# ============================================
# ANALYZE - Code Analysis (Claude)
# ============================================

analyze               ## Run full analysis
                      # make analyze.quick && make analyze.deep

analyze.quick         ## Quick analysis (bash)
                      # ./scripts/claude/analyze_codebase.sh

analyze.deep          ## Deep analysis (Go AST)
                      # go run scripts/claude/deep_analyzer.go

analyze.security      ## Security analysis only
                      # go run scripts/claude/deep_analyzer.go --security-only
```

**Implementação**:
- Scripts: `scripts/claude/analyze_codebase.sh`, `deep_analyzer.go`

---

## 🎯 ALIASES (Conveniência)

```makefile
# ============================================
# ALIASES - Convenience Commands
# ============================================

# Backwards compatibility + convenience
build                 ## Alias for: crm.build
	@make crm.build

run                   ## Alias for: crm.run
	@make crm.run

api                   ## Alias for: crm.run (with swagger)
	@make crm.run

test                  ## Run all tests
	@make test.unit && make test.integration && make test.e2e

fmt                   ## Alias for: quality.fmt
	@make quality.fmt

lint                  ## Alias for: quality.lint
	@make quality.lint
```

---

## 📊 ESTRUTURA DE SCRIPTS

```
scripts/
├── claude/                    # Análise de código (IA)
│   ├── analyze_codebase.sh
│   └── deep_analyzer.go
│
├── make/
│   ├── go/                    # Build, run Go
│   │   ├── build.sh
│   │   ├── run.sh
│   │   └── kill-port.sh
│   │
│   ├── infra/                 # Infraestrutura
│   │   ├── up.sh
│   │   ├── down.sh
│   │   └── delete.sh
│   │
│   ├── db/                    # Database
│   │   ├── migrate.sh
│   │   ├── seed.sh
│   │   └── reset.sh
│   │
│   ├── docker/                # Docker
│   │   ├── build.sh
│   │   └── push.sh
│   │
│   ├── helm/                  # Helm
│   │   ├── package.sh
│   │   └── push.sh
│   │
│   ├── k8s/                   # Kubernetes
│   │   ├── deploy-minikube.sh
│   │   └── minikube-start.sh
│   │
│   └── deploy/                # Deploy
│       ├── staging.sh
│       └── prod.sh
│
└── dev/                       # Development utilities
    ├── generate-domain-tests.sh
    ├── create-webhook-all-events.sh
    └── ...

tests/
└── scripts/                   # Test utilities
    ├── discover.sh            # Test discovery (intelligent)
    ├── test-e2e-waha.sh
    └── ...
```

---

## ✅ VALIDAÇÃO

### Checklist de Qualidade

- [ ] Todos os comandos seguem padrão `{categoria}.{ação}[.modificador]`
- [ ] Scripts isolados em `scripts/make/{categoria}/`
- [ ] Testes centralizados em `tests/`
- [ ] Discovery inteligente via `discover.sh`
- [ ] DB migrations centralizadas no Makefile (CI/CD)
- [ ] Minikube deploy completo (build → push → helm)
- [ ] Docker Compose para dev local
- [ ] AWX integration para staging/prod
- [ ] Aliases para backwards compatibility
- [ ] Documentação inline (comentários `##`)

---

## 🔗 INTEGRAÇÃO CI/CD

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI/CD

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Start infrastructure
        run: make infra.up

      - name: Run tests
        run: make test

      - name: Coverage
        run: make test.coverage

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: make docker.build.tag TAG=${{ github.sha }}

      - name: Push to registry
        run: make docker.push.tag TAG=${{ github.sha }}

      - name: Package Helm chart
        run: make helm.package

      - name: Push Helm chart
        run: make helm.push

  deploy-staging:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - name: Trigger AWX Deploy
        run: make deploy.staging

  deploy-prod:
    needs: deploy-staging
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://crm.ventros.cloud
    steps:
      - name: Trigger AWX Deploy (with approval)
        run: make deploy.prod
```

---

## 📝 PRÓXIMOS PASSOS

1. **Implementar Makefile** com todos os comandos
2. **Criar scripts** em `scripts/make/{categoria}/`
3. **Testar comandos** localmente
4. **Atualizar agente** `crm_docs_makefile_updater`
5. **Gerar MAKEFILE.md** com documentação completa
6. **Validar CI/CD** no GitHub Actions

---

## ❓ PERGUNTAS FINAIS

### 1. Registry Docker
- [ ] GitHub Container Registry (ghcr.io)
- [ ] Docker Hub
- [ ] AWS ECR
- [ ] Outro: __________

### 2. Helm Chart Registry
- [ ] GitHub Container Registry (OCI)
- [ ] ChartMuseum
- [ ] Harbor
- [ ] Outro: __________

### 3. AWX URL
- Produção: `https://awx.ventros.cloud` (exemplo)
- Job Template IDs:
  - Staging: ________
  - Production: ________

### 4. Ambientes Kubernetes
- [ ] Minikube (local dev)
- [ ] K3s/K3d (local dev)
- [ ] EKS (AWS)
- [ ] GKE (Google Cloud)
- [ ] AKS (Azure)
- [ ] RKE2 (Rancher)
- [ ] Outro: __________

---

**DESIGN COMPLETO!** ✅

Esta estrutura é:
- ✅ **Elegante**: Padrão consistente em todos os comandos
- ✅ **Padronizada**: Segue best practices da indústria
- ✅ **Escalável**: Fácil adicionar novos comandos
- ✅ **CI/CD Ready**: Integra com GitHub Actions + AWX
- ✅ **Developer Friendly**: Aliases e discover inteligente

**Aprovado?** Se sim, vou implementar! 🚀
