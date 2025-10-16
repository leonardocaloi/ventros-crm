# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Ventros CRM - Makefile (Elegant & Standardized)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# Pattern: {category}.{action}[.modifier]
# Categories: infra, crm, test, db, docker, helm, k8s, deploy
#
# Documentation: MAKEFILE.md
#
# Quick Start:
#   make infra.up      → Start infrastructure
#   make crm.run       → Run CRM API
#   make test.unit     → Run unit tests
#   make help          → Show all commands
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Variables
CONTAINER_RUNTIME ?= docker
COMPOSE = $(CONTAINER_RUNTIME) compose
COMPOSE_FILE = .deploy/container/compose.api.yaml
ENV_FILE = .deploy/container/.env
BINARY_NAME = crm-api
BINARY_PATH = bin/$(BINARY_NAME)
MAIN_PATH = cmd/api/main.go

# Colors
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
CYAN := \033[0;36m
RESET := \033[0m

.DEFAULT_GOAL := help
.PHONY: help

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📚 Help
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

help: ## Show this help message
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)  Ventros CRM - Development Commands$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_\.-]+:.*##/ { printf "  $(GREEN)%-25s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "Documentation: $(CYAN)MAKEFILE.md$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 1. INFRA - Infrastructure Management
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🏗️  Infrastructure

infra.up: ## Start infrastructure (Postgres, RabbitMQ, Redis, Temporal, Keycloak)
	@echo "$(BLUE)Starting infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@sleep 3
	@bash .deploy/container/scripts/init-keycloak.sh || true
	@echo "$(GREEN)✓ Infrastructure ready$(RESET)"
	@echo ""
	@echo "Services:"
	@echo "  • PostgreSQL: localhost:5432"
	@echo "  • RabbitMQ:   localhost:5672 (UI: http://localhost:15672)"
	@echo "  • Redis:      localhost:6379"
	@echo "  • Temporal:   localhost:7233 (UI: http://localhost:8088)"
	@echo "  • Keycloak:   http://localhost:8180"
	@echo ""

infra.down: ## Stop infrastructure (keep volumes)
	@echo "$(BLUE)Stopping infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) stop
	@echo "$(GREEN)✓ Infrastructure stopped$(RESET)"

infra.delete: ## Delete all volumes (DESTRUCTIVE)
	@echo "$(YELLOW)⚠️  Stopping infrastructure and removing volumes...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v
	@echo "$(GREEN)✓ Infrastructure cleaned$(RESET)"

infra.logs: ## Show infrastructure logs
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) logs -f

infra.restart: ## Restart infrastructure
	@$(MAKE) infra.down
	@$(MAKE) infra.up

infra.status: ## Show infrastructure status
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) ps

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 2. CRM - Go Application Management
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🚀 CRM Application

crm.run: ## Run CRM (go run cmd/api/main.go)
	@echo "$(BLUE)Starting CRM API...$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@go run $(MAIN_PATH)

crm.run.tunnel: ## Run CRM + Cloudflare Tunnel (for webhooks)
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)🌐 Starting CRM with Public Tunnel (for webhooks)$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(YELLOW)⚠️  Installing cloudflared if needed...$(RESET)"
	@which cloudflared >/dev/null 2>&1 || (echo "$(RED)cloudflared not found. Install with:$(RESET)"; echo "  wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb"; echo "  sudo dpkg -i cloudflared-linux-amd64.deb"; exit 1)
	@echo "$(GREEN)✓ cloudflared found$(RESET)"
	@echo ""
	@echo "$(BLUE)Starting API in background...$(RESET)"
	@lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
	@nohup go run $(MAIN_PATH) > /tmp/crm-api.log 2>&1 &
	@sleep 3
	@echo "$(GREEN)✓ API started at http://localhost:8080$(RESET)"
	@echo ""
	@echo "$(BLUE)Starting Cloudflare Tunnel...$(RESET)"
	@echo "$(YELLOW)📡 Public URL will appear below. Use it for webhooks!$(RESET)"
	@echo ""
	@cloudflared tunnel --url http://localhost:8080

crm.run.force: ## Kill port 8080 and run CRM
	@echo "$(BLUE)Killing port 8080...$(RESET)"
	@lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
	@echo "$(GREEN)✓ Port 8080 cleared$(RESET)"
	@$(MAKE) crm.run

crm.build: ## Build CRM binary (bin/crm-api)
	@echo "$(BLUE)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)✓ Binary: $(BINARY_PATH)$(RESET)"

crm.run.binary: crm.build ## Run CRM binary
	@echo "$(BLUE)Running $(BINARY_NAME) from binary...$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo ""
	@$(BINARY_PATH)

crm.run.binary.force: ## Rebuild + run binary
	@lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
	@$(MAKE) crm.build
	@$(MAKE) crm.run.binary

crm.infra.up: ## Start infra + run CRM (force)
	@$(MAKE) infra.up
	@$(MAKE) crm.run.force

crm.infra.up.reset: ## Delete infra + start + run CRM
	@$(MAKE) infra.delete
	@$(MAKE) infra.up
	@echo "$(BLUE)Running GORM AutoMigrate...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Schema created$(RESET)"
	@$(MAKE) crm.run.force

crm.container.build: ## Build Docker image (ventros-crm:latest)
	@echo "$(BLUE)Building Docker image...$(RESET)"
	@docker build -t ventros-crm:latest .
	@echo "$(GREEN)✓ Image built: ventros-crm:latest$(RESET)"

crm.container.run: ## Run Docker container (builds if needed)
	@if ! docker images | grep -q ventros-crm; then \
		$(MAKE) crm.container.build; \
	fi
	@echo "$(BLUE)Running container...$(RESET)"
	@docker run -p 8080:8080 --env-file .env ventros-crm:latest

crm.container.run.force: ## Rebuild + run container
	@$(MAKE) crm.container.build
	@docker rm -f ventros-crm 2>/dev/null || true
	@echo "$(BLUE)Running container...$(RESET)"
	@docker run -p 8080:8080 --name ventros-crm --env-file .env ventros-crm:latest

crm.compose.up: ## Start infra + CRM via Docker Compose
	@echo "$(BLUE)Starting Docker Compose stack...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@echo "$(GREEN)✓ Stack ready$(RESET)"

crm.compose.up.reset: ## Rebuild + delete infra + compose up
	@$(MAKE) infra.delete
	@echo "$(BLUE)Rebuilding and starting stack...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up --build -d
	@echo "$(GREEN)✓ Stack ready$(RESET)"

crm.compose.down: ## Stop Docker Compose
	@echo "$(BLUE)Stopping Docker Compose...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down
	@echo "$(GREEN)✓ Stack stopped$(RESET)"

crm.compose.logs: ## Show Docker Compose logs
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) logs -f

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 3. TEST - Intelligent Test Discovery
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🧪 Testing

test.discover: ## List all available tests
	@./tests/scripts/discover.sh list

test.stats: ## Show test statistics
	@./tests/scripts/discover.sh stats

test: ## Run all tests (unit + integration + e2e)
	@echo "$(BLUE)Running all tests...$(RESET)"
	@go test -v -race -timeout 10m ./...

test.unit: ## Run all unit tests (fast, no dependencies)
	@echo "$(BLUE)Running unit tests...$(RESET)"
	@go test -v -race -timeout 2m \
		./internal/domain/... \
		./internal/application/agent/... \
		./internal/application/note/... \
		./internal/application/contact_event/... \
		./internal/application/contact_list/... \
		./internal/application/tracking/... \
		./internal/application/commands/message/... \
		./infrastructure/crypto/... \
		./infrastructure/resilience/... \
		./infrastructure/channels/waha/...

test.integration: ## Run all integration tests (requires: infra.up)
	@echo "$(BLUE)Running integration tests...$(RESET)"
	@echo "$(YELLOW)⚠️  Requirements: Infrastructure must be running (make infra.up)$(RESET)"
	@go test -v -race -timeout 10m \
		./tests/integration/... \
		./infrastructure/persistence/... \
		./infrastructure/messaging/... \
		./infrastructure/websocket/...

test.e2e: ## Run all E2E tests (requires: infra.up + crm.run)
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@echo "$(YELLOW)⚠️  Requirements: API must be running (make crm.run)$(RESET)"
	@go test -v -timeout 10m ./tests/e2e/...

test.unit.domain: ## Unit tests: internal/domain/*
	@echo "$(BLUE)Running domain unit tests...$(RESET)"
	@go test -v ./internal/domain/... -count=1

test.unit.application: ## Unit tests: internal/application/*
	@echo "$(BLUE)Running application unit tests...$(RESET)"
	@go test -v ./internal/application/... -count=1 -short

test.unit.infra: ## Unit tests: infrastructure/*
	@echo "$(BLUE)Running infrastructure unit tests...$(RESET)"
	@go test -v ./infrastructure/... -count=1 -short

test.integration.waha: ## Integration: WAHA
	@echo "$(BLUE)Running WAHA integration tests...$(RESET)"
	@go test -v ./infrastructure/channels/waha/... -count=1

test.integration.db: ## Integration: Database
	@echo "$(BLUE)Running database integration tests...$(RESET)"
	@go test -v ./infrastructure/persistence/... -count=1

test.integration.mq: ## Integration: RabbitMQ
	@echo "$(BLUE)Running RabbitMQ integration tests...$(RESET)"
	@go test -v ./infrastructure/messaging/... -count=1

test.e2e.waha: ## E2E: WAHA integration flow
	@echo "$(BLUE)Running WAHA E2E tests...$(RESET)"
	@./tests/scripts/test-e2e-waha.sh

test.e2e.campaign: ## E2E: Campaign flow
	@echo "$(BLUE)Running Campaign E2E tests...$(RESET)"
	@go test -v ./tests/e2e/campaign/... -count=1

test.e2e.sequence: ## E2E: Sequence flow
	@echo "$(BLUE)Running Sequence E2E tests...$(RESET)"
	@go test -v ./tests/e2e/sequence/... -count=1

test.e2e.broadcast: ## E2E: Broadcast flow
	@echo "$(BLUE)Running Broadcast E2E tests...$(RESET)"
	@go test -v ./tests/e2e/broadcast/... -count=1

test.e2e.pipeline: ## E2E: Pipeline flow
	@echo "$(BLUE)Running Pipeline E2E tests...$(RESET)"
	@go test -v ./tests/e2e/pipeline/... -count=1

test.e2e.import: ## E2E: WAHA History Import (full test)
	@echo "$(BLUE)Running WAHA History Import E2E test...$(RESET)"
	@go test -v -timeout 10m -run TestWAHAHistoryImportTestSuite ./tests/e2e/

test.e2e.reset.import: ## Reset infra + run E2E import test
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)🔄 Full E2E Reset + Import Test$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@$(MAKE) infra.delete
	@$(MAKE) infra.up
	@echo "$(BLUE)Running GORM AutoMigrate...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Schema created$(RESET)"
	@echo "$(BLUE)Starting CRM API in background...$(RESET)"
	@lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
	@nohup go run cmd/api/main.go > /tmp/crm-api.log 2>&1 &
	@sleep 5
	@echo "$(GREEN)✓ API started$(RESET)"
	@echo ""
	@echo "$(BLUE)Cleaning test cache...$(RESET)"
	@go clean -testcache
	@echo "$(GREEN)✓ Cache cleaned$(RESET)"
	@echo ""
	@echo "$(BLUE)Running E2E Import Test...$(RESET)"
	@go test -v -timeout 10m -run TestWAHAHistoryImportTestSuite ./tests/e2e/ || (echo "$(RED)Test failed. API logs:$(RESET)"; tail -50 /tmp/crm-api.log; exit 1)
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ E2E Import Test completed!$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"

test.e2e.reset: ## Reset infra + run all E2E tests (generic)
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)🔄 Full E2E Reset + All Tests$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@$(MAKE) infra.delete
	@$(MAKE) infra.up
	@echo "$(BLUE)Running GORM AutoMigrate...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Schema created$(RESET)"
	@echo "$(BLUE)Starting CRM API in background...$(RESET)"
	@lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
	@nohup go run cmd/api/main.go > /tmp/crm-api.log 2>&1 &
	@sleep 5
	@echo "$(GREEN)✓ API started$(RESET)"
	@echo ""
	@echo "$(BLUE)Running All E2E Tests...$(RESET)"
	@go test -v -timeout 15m ./tests/e2e/... || (echo "$(RED)Tests failed. API logs:$(RESET)"; tail -50 /tmp/crm-api.log; exit 1)
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ All E2E Tests completed!$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"

test.coverage: ## Coverage report (all tests)
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@mkdir -p coverage
	@go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@go tool cover -func=coverage/coverage.out | grep total | awk '{print "$(GREEN)Total Coverage: " $$3 "$(RESET)"}'
	@echo "$(CYAN)Coverage report: coverage/coverage.html$(RESET)"

test.coverage.unit: ## Coverage (unit tests only)
	@echo "$(BLUE)Running unit tests with coverage...$(RESET)"
	@mkdir -p coverage
	@go test -v -race -coverprofile=coverage/coverage-unit.out -covermode=atomic \
		./internal/domain/... \
		./internal/application/agent/... \
		./internal/application/note/... \
		./internal/application/contact_event/... \
		./internal/application/contact_list/... \
		./internal/application/tracking/... \
		./internal/application/commands/message/...
	@go tool cover -html=coverage/coverage-unit.out -o coverage/coverage-unit.html
	@go tool cover -func=coverage/coverage-unit.out | grep total | awk '{print "$(GREEN)Unit Test Coverage: " $$3 "$(RESET)"}'
	@echo "$(CYAN)Coverage report: coverage/coverage-unit.html$(RESET)"

test.coverage.html: ## Open coverage HTML report
	@if [ -f coverage/coverage.html ]; then \
		xdg-open coverage/coverage.html 2>/dev/null || open coverage/coverage.html 2>/dev/null; \
	else \
		echo "$(RED)✗ Coverage report not found. Run: make test.coverage$(RESET)"; \
	fi

test.bench: ## Run all benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem -timeout 5m ./cmd/benchmark/...

test.bench.domain: ## Benchmarks: domain layer
	@echo "$(BLUE)Running domain benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./internal/domain/...

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 4. DB - Database Management
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🗄️  Database

db.migrate.up: ## Apply migrations (PRODUCTION)
	@echo "$(BLUE)Running migrations...$(RESET)"
	@go run cmd/migrate/main.go up
	@echo "$(GREEN)✓ Migrations applied$(RESET)"

db.migrate.down: ## Rollback last migration
	@echo "$(YELLOW)⚠️  Rolling back last migration...$(RESET)"
	@go run cmd/migrate/main.go down
	@echo "$(GREEN)✓ Migration rolled back$(RESET)"

db.migrate.status: ## Show migration status
	@go run cmd/migrate/main.go status

db.migrate.create: ## Create new migration (Usage: make db.migrate.create NAME=add_users)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)✗ NAME is required. Usage: make db.migrate.create NAME=add_users$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)$(RESET)"
	@migrate create -ext sql -dir infrastructure/database/migrations -seq $(NAME)
	@echo "$(GREEN)✓ Migration created$(RESET)"

db.seed: ## Seed database with test data
	@echo "$(BLUE)Seeding database...$(RESET)"
	@if [ -f cmd/seed/main.go ]; then \
		go run cmd/seed/main.go; \
		echo "$(GREEN)✓ Database seeded$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  Seed script not found (cmd/seed/main.go)$(RESET)"; \
	fi

db.reset: ## Reset database (drop + migrate + seed)
	@echo "$(YELLOW)⚠️  Resetting database...$(RESET)"
	@PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;" > /dev/null 2>&1
	@$(MAKE) db.migrate.up
	@$(MAKE) db.seed
	@echo "$(GREEN)✓ Database reset complete$(RESET)"

db.console: ## Open PostgreSQL console
	@echo "$(BLUE)Opening PostgreSQL console...$(RESET)"
	@PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 5. DOCKER - Image Management
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🐳 Docker

docker.build: ## Build Docker image (ventros-crm:latest)
	@echo "$(BLUE)Building Docker image...$(RESET)"
	@docker build -t ventros-crm:latest .
	@echo "$(GREEN)✓ Image built: ventros-crm:latest$(RESET)"

docker.build.tag: ## Build with custom tag (Usage: make docker.build.tag TAG=v1.2.3)
	@if [ -z "$(TAG)" ]; then \
		echo "$(RED)✗ TAG is required. Usage: make docker.build.tag TAG=v1.2.3$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Building Docker image with tag: $(TAG)$(RESET)"
	@docker build -t ventros-crm:$(TAG) .
	@echo "$(GREEN)✓ Image built: ventros-crm:$(TAG)$(RESET)"

docker.push: ## Push to registry (requires: login)
	@echo "$(BLUE)Pushing image to registry...$(RESET)"
	@docker tag ventros-crm:latest ghcr.io/ventros/crm:latest
	@docker push ghcr.io/ventros/crm:latest
	@echo "$(GREEN)✓ Image pushed$(RESET)"

docker.push.tag: ## Push specific tag (Usage: make docker.push.tag TAG=v1.2.3)
	@if [ -z "$(TAG)" ]; then \
		echo "$(RED)✗ TAG is required. Usage: make docker.push.tag TAG=v1.2.3$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Pushing image with tag: $(TAG)$(RESET)"
	@docker tag ventros-crm:$(TAG) ghcr.io/ventros/crm:$(TAG)
	@docker push ghcr.io/ventros/crm:$(TAG)
	@echo "$(GREEN)✓ Image pushed: ghcr.io/ventros/crm:$(TAG)$(RESET)"

docker.login: ## Login to GitHub Container Registry
	@echo "$(BLUE)Logging in to GitHub Container Registry...$(RESET)"
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "$(RED)✗ GITHUB_TOKEN is required$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GITHUB_TOKEN)" | docker login ghcr.io -u $(GITHUB_USER) --password-stdin
	@echo "$(GREEN)✓ Logged in$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 6. HELM - Chart Management
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ ⎈ Helm

helm.package: ## Package Helm chart
	@echo "$(BLUE)Packaging Helm chart...$(RESET)"
	@helm package .deploy/helm/ventros-crm -d .deploy/helm/
	@echo "$(GREEN)✓ Chart packaged$(RESET)"

helm.push: ## Push chart to registry
	@echo "$(BLUE)Pushing Helm chart to registry...$(RESET)"
	@CHART_FILE=$$(ls .deploy/helm/ventros-crm-*.tgz | head -1); \
	if [ -z "$$CHART_FILE" ]; then \
		echo "$(RED)✗ Chart not found. Run: make helm.package$(RESET)"; \
		exit 1; \
	fi; \
	helm push $$CHART_FILE oci://ghcr.io/ventros/charts
	@echo "$(GREEN)✓ Chart pushed$(RESET)"

helm.install.dev: ## Install chart to dev namespace
	@echo "$(BLUE)Installing chart to dev namespace...$(RESET)"
	@helm install ventros-crm .deploy/helm/ventros-crm \
		-n dev --create-namespace -f .deploy/helm/values-dev.yaml
	@echo "$(GREEN)✓ Chart installed$(RESET)"

helm.upgrade.dev: ## Upgrade chart in dev
	@echo "$(BLUE)Upgrading chart in dev namespace...$(RESET)"
	@helm upgrade ventros-crm .deploy/helm/ventros-crm \
		-n dev -f .deploy/helm/values-dev.yaml
	@echo "$(GREEN)✓ Chart upgraded$(RESET)"

helm.uninstall.dev: ## Uninstall chart from dev
	@echo "$(BLUE)Uninstalling chart from dev namespace...$(RESET)"
	@helm uninstall ventros-crm -n dev
	@echo "$(GREEN)✓ Chart uninstalled$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 7. K8S - Kubernetes Management (Minikube)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ ☸️  Kubernetes

k8s.minikube.start: ## Start Minikube
	@echo "$(BLUE)Starting Minikube...$(RESET)"
	@minikube start --cpus=4 --memory=8192
	@echo "$(GREEN)✓ Minikube started$(RESET)"

k8s.minikube.stop: ## Stop Minikube
	@echo "$(BLUE)Stopping Minikube...$(RESET)"
	@minikube stop
	@echo "$(GREEN)✓ Minikube stopped$(RESET)"

k8s.minikube.delete: ## Delete Minikube cluster
	@echo "$(YELLOW)⚠️  Deleting Minikube cluster...$(RESET)"
	@minikube delete
	@echo "$(GREEN)✓ Minikube cluster deleted$(RESET)"

k8s.minikube.dashboard: ## Open Minikube dashboard
	@minikube dashboard

k8s.deploy.minikube: ## Full deploy to Minikube (build → load → helm)
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)☸️  Full Minikube Deploy$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 1/4: Building image...$(RESET)"
	@$(MAKE) docker.build.tag TAG=minikube
	@echo ""
	@echo "$(BLUE)Step 2/4: Loading image to Minikube...$(RESET)"
	@minikube image load ventros-crm:minikube
	@echo "$(GREEN)✓ Image loaded to Minikube$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 3/4: Packaging Helm chart...$(RESET)"
	@$(MAKE) helm.package
	@echo ""
	@echo "$(BLUE)Step 4/4: Installing Helm chart...$(RESET)"
	@$(MAKE) helm.install.dev
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ Minikube deploy complete$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"

k8s.deploy.minikube.reset: ## Rebuild + redeploy to Minikube
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)☸️  Minikube Reset & Deploy$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 1/5: Uninstalling chart...$(RESET)"
	@helm uninstall ventros-crm -n dev || true
	@echo ""
	@echo "$(BLUE)Step 2/5: Removing old image...$(RESET)"
	@minikube image rm ventros-crm:minikube || true
	@echo ""
	@echo "$(BLUE)Step 3/5: Building new image...$(RESET)"
	@$(MAKE) docker.build.tag TAG=minikube
	@echo ""
	@echo "$(BLUE)Step 4/5: Loading image to Minikube...$(RESET)"
	@minikube image load ventros-crm:minikube
	@echo ""
	@echo "$(BLUE)Step 5/5: Installing Helm chart...$(RESET)"
	@$(MAKE) helm.install.dev
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ Minikube reset complete$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 8. DEPLOY - Environment Deployment
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🚀 Deployment

deploy.dev: ## Deploy to development (local Docker Compose)
	@$(MAKE) crm.compose.up

deploy.dev.reset: ## Reset + deploy to development
	@$(MAKE) crm.compose.up.reset

deploy.staging: ## Deploy to staging (triggers AWX)
	@echo "$(BLUE)Triggering staging deployment via AWX...$(RESET)"
	@if [ -z "$(AWX_TOKEN)" ]; then \
		echo "$(RED)✗ AWX_TOKEN is required$(RESET)"; \
		exit 1; \
	fi
	@curl -X POST https://awx.ventros.cloud/api/v2/job_templates/$(AWX_STAGING_JOB_ID)/launch/ \
		-H "Authorization: Bearer $(AWX_TOKEN)" \
		-d '{"extra_vars": "{\"environment\": \"staging\"}"}'
	@echo "$(GREEN)✓ Staging deployment triggered$(RESET)"

deploy.staging.full: ## Full deploy to staging (DB refresh + deploy)
	@echo "$(BLUE)Triggering full staging deployment via AWX...$(RESET)"
	@if [ -z "$(AWX_TOKEN)" ]; then \
		echo "$(RED)✗ AWX_TOKEN is required$(RESET)"; \
		exit 1; \
	fi
	@curl -X POST https://awx.ventros.cloud/api/v2/job_templates/$(AWX_STAGING_JOB_ID)/launch/ \
		-H "Authorization: Bearer $(AWX_TOKEN)" \
		-d '{"extra_vars": "{\"environment\": \"staging\", \"db_refresh\": true}"}'
	@echo "$(GREEN)✓ Full staging deployment triggered$(RESET)"

deploy.prod: ## Deploy to production (triggers AWX with approval)
	@echo "$(YELLOW)⚠️  Triggering production deployment via AWX (requires approval)...$(RESET)"
	@if [ -z "$(AWX_TOKEN)" ]; then \
		echo "$(RED)✗ AWX_TOKEN is required$(RESET)"; \
		exit 1; \
	fi
	@curl -X POST https://awx.ventros.cloud/api/v2/job_templates/$(AWX_PROD_JOB_ID)/launch/ \
		-H "Authorization: Bearer $(AWX_TOKEN)" \
		-d '{"extra_vars": "{\"environment\": \"production\"}"}'
	@echo "$(GREEN)✓ Production deployment triggered (awaiting approval)$(RESET)"

deploy.rollback.staging: ## Rollback staging to previous version
	@echo "$(YELLOW)⚠️  Rolling back staging...$(RESET)"
	@helm rollback ventros-crm -n staging
	@echo "$(GREEN)✓ Staging rolled back$(RESET)"

deploy.rollback.prod: ## Rollback production to previous version
	@echo "$(YELLOW)⚠️  Rolling back production...$(RESET)"
	@helm rollback ventros-crm -n production
	@echo "$(GREEN)✓ Production rolled back$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# ALIASES - Backwards Compatibility
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🔗 Aliases (Backwards Compatibility)

build: crm.build ## Alias for: crm.build

run: crm.run ## Alias for: crm.run

api: crm.run ## Alias for: crm.run

infra: infra.up ## Alias for: infra.up

# Simple commands (not categorized)
fmt: ## Format code (gofmt + goimports)
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./... > /dev/null
	@goimports -w . 2>/dev/null || echo "$(YELLOW)⚠️  goimports not found$(RESET)"
	@echo "$(GREEN)✓ Code formatted$(RESET)"

lint: ## Run golangci-lint
	@echo "$(BLUE)Running linter...$(RESET)"
	@golangci-lint run --timeout 5m || echo "$(YELLOW)⚠️  Some linting issues found$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✓ No issues found$(RESET)"

swagger: ## Generate Swagger documentation
	@echo "$(BLUE)Generating Swagger docs...$(RESET)"
	@if command -v swag >/dev/null 2>&1; then \
		swag fmt > /dev/null 2>&1; \
		swag init -g $(MAIN_PATH) -o docs --parseDependency --parseInternal > /dev/null 2>&1; \
		echo "$(GREEN)✓ Swagger docs generated$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  swag not found$(RESET)"; \
	fi

clean: ## Clean everything (containers, volumes, binaries, cache)
	@bash scripts/make/infra/clean.sh

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# End of Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
