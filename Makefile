# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Ventros CRM - Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# ESSENTIAL COMMANDS ONLY
# Read: MAKEFILE.md for detailed documentation
#
# Quick Start:
#   make clean    → Clean everything (API + containers + data)
#   make infra    → Start infrastructure (Postgres, RabbitMQ, Redis, Temporal)
#   make api      → Run API locally (requires infra running)
#   make container → Start EVERYTHING containerized (infra + API)
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Container runtime
CONTAINER_RUNTIME ?= docker
COMPOSE = $(CONTAINER_RUNTIME) compose
COMPOSE_FILE = .deploy/container/compose.api.yaml
ENV_FILE = .deploy/container/.env

# Colors
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
RESET := \033[0m

.DEFAULT_GOAL := help
.PHONY: help

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📚 Help
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

help: ## Show available commands
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)  Ventros CRM - Essential Commands$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-18s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "Documentation: $(YELLOW)MAKEFILE.md$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧹 Clean
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🧹 Cleanup

clean: ## Clean EVERYTHING (API + containers + volumes + files) - DESTRUCTIVE
	@bash scripts/clean.sh

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🚀 Development (Local)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🚀 Local Development

infra: ## Start infrastructure (Postgres, RabbitMQ, Redis, Temporal, Keycloak)
	@echo "$(BLUE)Starting Infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@echo ""
	@echo "$(YELLOW)Initializing Keycloak...$(RESET)"
	@bash .deploy/container/scripts/init-keycloak.sh
	@echo ""
	@echo "$(GREEN)✓ Infrastructure ready$(RESET)"
	@echo ""
	@echo "Services:"
	@echo "  • PostgreSQL: localhost:5432"
	@echo "  • RabbitMQ:   localhost:5672 (UI: http://localhost:15672)"
	@echo "  • Redis:      localhost:6379"
	@echo "  • Temporal:   localhost:7233 (UI: http://localhost:8088)"
	@echo "  • Keycloak:   http://localhost:8180 (admin/admin123)"
	@echo ""
	@echo "Next: $(GREEN)make api$(RESET)"

infra-clean: ## Stop infrastructure and clean volumes (DESTRUCTIVE)
	@echo "$(YELLOW)⚠️  Stopping infrastructure and cleaning volumes...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v
	@echo "$(GREEN)✓ Infrastructure stopped and volumes removed$(RESET)"

api: swagger ## Run API locally (requires: make infra)
	@echo "$(BLUE)Starting API...$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@go run cmd/api/main.go

build: ## Build API binary (output: bin/api)
	@echo "$(BLUE)Building binary...$(RESET)"
	@go build -o bin/api cmd/api/main.go
	@echo "$(GREEN)✓ Binary: bin/api$(RESET)"

swagger: ## Generate Swagger docs
	@swag fmt > /dev/null 2>&1
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal > /dev/null 2>&1
	@echo "$(GREEN)✓ Swagger docs generated$(RESET)"

fmt: ## Format code (go fmt + goimports)
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./... > /dev/null
	@goimports -w . 2>/dev/null || true
	@echo "$(GREEN)✓ Code formatted$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🐳 Container (Everything Dockerized)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🐳 Container (All Services)

container: ## Start EVERYTHING containerized (infra + API)
	@echo "$(BLUE)Starting containerized stack...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d --build
	@echo ""
	@echo "$(GREEN)✓ Stack ready$(RESET)"
	@echo ""
	@echo "  • API: http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo ""

container-down: ## Stop all containers
	@echo "$(BLUE)Stopping containers...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down
	@echo "$(GREEN)✓ Containers stopped$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧪 Testing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🗄️ Database

migrate-auto: ## Run GORM AutoMigrate (DEV ONLY - creates/updates schema)
	@echo "$(YELLOW)⚠️  Running GORM AutoMigrate (DEV ONLY)$(RESET)"
	@echo "$(BLUE)This will sync database schema with Go entities...$(RESET)"
	@go run cmd/migrate/automigrate.go
	@echo "$(GREEN)✅ AutoMigrate completed$(RESET)"

##@ 🧪 Testing

test: ## Run unit tests
	@echo "$(BLUE)Running tests...$(RESET)"
	@go test -v -race ./...

test-waha: ## Run WAHA webhook tests (uses events_waha/*.json)
	@echo "$(BLUE)Running WAHA webhook tests...$(RESET)"
	@echo "$(YELLOW)Requirements: API running (make api)$(RESET)"
	@go test -v -timeout 10m -run TestWAHAWebhookTestSuite ./tests/e2e/

test-waha-session: ## Run WAHA E2E test with session_id (usage: SESSION=5511999999999 make test-waha-session)
	@if [ -z "$(SESSION)" ]; then \
		echo "$(RED)❌ Error: SESSION is required$(RESET)"; \
		echo ""; \
		echo "Usage:"; \
		echo "  $(GREEN)SESSION=5511999999999 make test-waha-session$(RESET)"; \
		echo ""; \
		exit 1; \
	fi
	@bash scripts/test-e2e-waha.sh $(SESSION)

test-e2e: ## Run E2E test (User + Project + Pipeline + Channel + Messages)
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)       🚀 FULL E2E SYSTEM TEST$(RESET)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(GREEN)This will test:$(RESET)"
	@echo "  1. Create user (auth)"
	@echo "  2. Create project"
	@echo "  3. Create pipeline + statuses"
	@echo "  4. Create WhatsApp channel"
	@echo "  5. Send all message types"
	@echo "  6. Verify database"
	@echo ""
	@echo "$(YELLOW)Requirements:$(RESET)"
	@echo "  • API running (make api)"
	@echo "  • WAHA session '5511999999999'"
	@echo ""
	@read -p "Press Enter to continue or Ctrl+C to cancel..."
	@echo ""
	@bash scripts/setup-all-complete.sh

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# End of Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
