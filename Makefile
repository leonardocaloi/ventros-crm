# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Ventros CRM - Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# ESSENTIAL COMMANDS
# Read: guides/MAKEFILE.md for detailed documentation
#
# Quick Start:
#   make build      → Build API binary
#   make infra      → Start infrastructure (Postgres, RabbitMQ, Redis, Temporal)
#   make api        → Run API locally (requires infra)
#   make test       → Run all tests
#   make test-unit  → Run unit tests only (fast)
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Container runtime
CONTAINER_RUNTIME ?= docker
COMPOSE = $(CONTAINER_RUNTIME) compose
COMPOSE_FILE = .deploy/container/compose.api.yaml
ENV_FILE = .deploy/container/.env

# Add Go bin to PATH for development tools
export PATH := $(HOME)/go/bin:$(PATH)

# Build variables
BINARY_NAME = crm-api
BINARY_PATH = bin/$(BINARY_NAME)
MAIN_PATH = cmd/api/main.go

# Test variables
TEST_TIMEOUT = 10m
COVERAGE_DIR = coverage
COVERAGE_FILE = $(COVERAGE_DIR)/coverage.out

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
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "Documentation: $(CYAN)guides/MAKEFILE.md$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔨 Build
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🔨 Build

build: ## Build API binary (output: bin/crm-api)
	@echo "$(BLUE)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)✓ Binary: $(BINARY_PATH)$(RESET)"

build-linux: ## Build API binary for Linux (for Docker)
	@echo "$(BLUE)Building $(BINARY_NAME) for Linux...$(RESET)"
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)✓ Binary: $(BINARY_PATH) (linux/amd64)$(RESET)"

run-binary: build ## Build and run the binary (test production build locally)
	@echo "$(BLUE)Running $(BINARY_NAME) from binary...$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@$(BINARY_PATH)

clean-bin: ## Remove binary
	@echo "$(BLUE)Removing binary...$(RESET)"
	@rm -f $(BINARY_PATH)
	@echo "$(GREEN)✓ Binary removed$(RESET)"

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

infra-logs: ## Show infrastructure logs
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) logs -f

infra-stop: ## Stop infrastructure (keep data)
	@echo "$(BLUE)Stopping infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) stop
	@echo "$(GREEN)✓ Infrastructure stopped$(RESET)"

infra-clean: ## Stop infrastructure and remove volumes (DESTRUCTIVE)
	@echo "$(YELLOW)⚠️  Stopping infrastructure and cleaning volumes...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v
	@echo "$(GREEN)✓ Infrastructure stopped and volumes removed$(RESET)"

fresh: ## ✨ Fresh start (infra up → clean DB → AutoMigrate → API) - FAST DEV
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)✨ FRESH START - Clean slate in seconds$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 1/5: Killing any process on port 8080...$(RESET)"
	@lsof -ti:8080 | xargs -r kill -9 2>/dev/null || true
	@echo "$(GREEN)✓ Port 8080 cleared$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 2/5: Ensuring infrastructure is up...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@sleep 3
	@echo "$(GREEN)✓ Infrastructure ready$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 3/5: Dropping and recreating schema...$(RESET)"
	@PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;" > /dev/null 2>&1
	@echo "$(GREEN)✓ Database cleaned$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 4/5: Running AutoMigrate...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Schema created$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 5/5: Starting API...$(RESET)"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ FRESH START COMPLETE$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@$(MAKE) api

reset-full: ## 🔥 Full reset from scratch (infra + DB + AutoMigrate + API via go run) - DEV
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(YELLOW)🔥 FULL RESET - Starting from scratch (go run - DEV)$(RESET)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 1/5: Stopping and cleaning infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v
	@echo "$(GREEN)✓ Infrastructure cleaned$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 2/5: Starting fresh infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@echo ""
	@echo "$(BLUE)Step 3/5: Waiting for services to be ready...$(RESET)"
	@sleep 8
	@echo "$(GREEN)✓ Services ready$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 4/5: Running GORM AutoMigrate (creating schema)...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Database schema created$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 5/5: Starting API (go run)...$(RESET)"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ RESET COMPLETE - API starting...$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@$(MAKE) api

run-binary-full: ## 🔥 Full reset from scratch (infra + DB + AutoMigrate + Binary) - TEST PROD
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(YELLOW)🔥 FULL RESET - Starting from scratch (binary - PROD TEST)$(RESET)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 1/6: Stopping and cleaning infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v
	@echo "$(GREEN)✓ Infrastructure cleaned$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 2/6: Starting fresh infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@echo ""
	@echo "$(BLUE)Step 3/6: Waiting for services to be ready...$(RESET)"
	@sleep 8
	@echo "$(GREEN)✓ Services ready$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 4/6: Running GORM AutoMigrate (creating schema)...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Database schema created$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 5/6: Building production binary...$(RESET)"
	@$(MAKE) build
	@echo ""
	@echo "$(BLUE)Step 6/6: Starting API from binary...$(RESET)"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ RESET COMPLETE - Running production binary...$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@$(BINARY_PATH)

api: swagger ## Run API locally (requires: make infra)
	@echo "$(BLUE)Starting API...$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@go run $(MAIN_PATH)

swagger: ## Generate Swagger documentation
	@echo "$(BLUE)Generating Swagger docs...$(RESET)"
	@if command -v swag >/dev/null 2>&1; then \
		swag fmt > /dev/null 2>&1; \
		swag init -g $(MAIN_PATH) -o docs --parseDependency --parseInternal > /dev/null 2>&1; \
		echo "$(GREEN)✓ Swagger docs generated$(RESET)"; \
	else \
		echo "$(YELLOW)⚠️  swag not found, run: make deps$(RESET)"; \
	fi

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
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo ""

container-logs: ## Show container logs
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) logs -f ventros-api

container-stop: ## Stop all containers (keep data)
	@echo "$(BLUE)Stopping containers...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) stop
	@echo "$(GREEN)✓ Containers stopped$(RESET)"

container-down: ## Stop and remove all containers
	@echo "$(BLUE)Removing containers...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down
	@echo "$(GREEN)✓ Containers removed$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧪 Testing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🧪 Testing

test: ## Run all tests (unit + integration + e2e)
	@echo "$(BLUE)Running all tests...$(RESET)"
	@go test -v -race -timeout $(TEST_TIMEOUT) ./...

test-unit: ## Run unit tests only (fast, no external dependencies)
	@echo "$(BLUE)Running unit tests...$(RESET)"
	@echo "$(CYAN)Testing: domain + application layers$(RESET)"
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

test-integration: ## Run integration tests (requires: make infra)
	@echo "$(BLUE)Running integration tests...$(RESET)"
	@echo "$(YELLOW)⚠️  Requirements: Infrastructure must be running (make infra)$(RESET)"
	@echo "$(CYAN)Testing: database, messaging, websocket$(RESET)"
	@go test -v -race -timeout $(TEST_TIMEOUT) \
		./tests/integration/... \
		./infrastructure/persistence/... \
		./infrastructure/messaging/... \
		./infrastructure/websocket/...

test-e2e: ## Run E2E tests (requires: make infra + API running)
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@echo "$(YELLOW)⚠️  Requirements: API must be running (make api)$(RESET)"
	@echo "$(CYAN)Testing: full system workflows$(RESET)"
	@go test -v -timeout $(TEST_TIMEOUT) ./tests/e2e/...

msg-e2e-send: ## 📨 E2E test: Send message with system agents (requires: make infra + API running)
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)📨 E2E Test: Message Send with System Agents$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(YELLOW)⚠️  Requirements:$(RESET)"
	@echo "  • Infrastructure running (make infra)"
	@echo "  • API running (make api)"
	@echo "  • WAHA session configured in .env"
	@echo ""
	@./tests/e2e/msg_send_test.sh

msg-e2e-types: ## 📨 E2E test: All message types (text, image, video, audio, etc) - COMPREHENSIVE
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)📨 E2E Test: All Message Types (Send → WAHA → Webhook)$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(CYAN)This test validates the complete message flow:$(RESET)"
	@echo "  1. Send message via API (outbound)"
	@echo "  2. WAHA processes and delivers message"
	@echo "  3. Webhook receives and processes (inbound)"
	@echo ""
	@echo "$(CYAN)Types tested:$(RESET)"
	@echo "  • text      - Text messages"
	@echo "  • image     - Images (JPEG, PNG)"
	@echo "  • video     - Videos (MP4)"
	@echo "  • audio     - Audio files (MP3)"
	@echo "  • document  - Documents (PDF)"
	@echo "  • location  - Geographic location"
	@echo "  • contact   - vCard contacts"
	@echo ""
	@echo "$(YELLOW)⚠️  Requirements:$(RESET)"
	@echo "  • Infrastructure running (make infra)"
	@echo "  • API running (make api)"
	@echo "  • WAHA session configured in .env"
	@echo ""
	@./tests/e2e/msg_types_test.sh

waha-import-e2e: ## 📥 E2E test: WAHA History Import (requires: make infra + API + Temporal)
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)📥 E2E Test: WAHA History Import$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(CYAN)This test validates the complete history import flow:$(RESET)"
	@echo "  1. Create WAHA channel with import configuration"
	@echo "  2. Start history import (triggers Temporal workflow)"
	@echo "  3. Poll import status until completion"
	@echo "  4. Verify contacts, messages, and sessions created"
	@echo ""
	@echo "$(CYAN)Import strategies tested:$(RESET)"
	@echo "  • recent       - Import recent messages only"
	@echo "  • time-limited - Import last 7/30/90 days"
	@echo "  • msg-limited  - Limit messages per chat"
	@echo ""
	@echo "$(YELLOW)⚠️  Requirements:$(RESET)"
	@echo "  • Infrastructure running (make infra)"
	@echo "  • API running (make api)"
	@echo "  • Temporal worker running"
	@echo "  • WAHA instance accessible"
	@echo ""
	@go test -v -timeout 15m ./tests/e2e/waha_history_import_test.go

test-import: ## 🧪 Full import test (FULL RESET + API + worker + import test) - ALL-IN-ONE
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(CYAN)🧪 FULL IMPORT TEST - Complete Reset + Test$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 1/7: Stopping and removing all containers + volumes...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v
	@echo "$(GREEN)✓ Infrastructure cleaned (volumes removed)$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 2/7: Starting fresh infrastructure...$(RESET)"
	@$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d
	@echo ""
	@echo "$(BLUE)Step 3/7: Waiting for services to be ready...$(RESET)"
	@sleep 8
	@echo "$(GREEN)✓ Services ready$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 4/7: Running GORM AutoMigrate (creating fresh schema)...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✓ Database schema created$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 5/7: Starting API in background (includes Temporal worker)...$(RESET)"
	@$(MAKE) swagger > /dev/null 2>&1
	@nohup go run $(MAIN_PATH) > /tmp/ventros-api.log 2>&1 & echo $$! > /tmp/ventros-api.pid
	@echo "$(GREEN)✓ API started (PID: $$(cat /tmp/ventros-api.pid))$(RESET)"
	@echo "$(CYAN)  • Logs: tail -f /tmp/ventros-api.log$(RESET)"
	@echo ""
	@echo "$(BLUE)Step 6/7: Waiting for API to be ready...$(RESET)"
	@for i in $$(seq 1 30); do \
		if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
			echo "$(GREEN)✓ API is ready$(RESET)"; \
			break; \
		fi; \
		sleep 1; \
	done
	@echo ""
	@echo "$(BLUE)Step 7/7: Running import test...$(RESET)"
	@echo "$(CYAN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(CYAN)Loading environment variables from .env...$(RESET)"
	@if [ -f .env ]; then \
		export $$(cat .env | grep -E '^WAHA_' | xargs) && \
		go test -v -timeout 15m ./tests/e2e/waha_history_import_test.go || true; \
	else \
		echo "$(RED)✗ .env file not found!$(RESET)"; \
		go test -v -timeout 15m ./tests/e2e/waha_history_import_test.go || true; \
	fi
	@echo ""
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(YELLOW)🛑 Stopping API...$(RESET)"
	@if [ -f /tmp/ventros-api.pid ]; then \
		kill -TERM $$(cat /tmp/ventros-api.pid) 2>/dev/null || true; \
		rm -f /tmp/ventros-api.pid; \
		echo "$(GREEN)✓ API stopped$(RESET)"; \
	fi
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ IMPORT TEST COMPLETED$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "📄 Logs: /tmp/ventros-api.log"

test-bench: ## Run benchmark tests
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem -timeout 5m ./cmd/benchmark/...

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "$(GREEN)Total Coverage: " $$3 "$(RESET)"}'
	@echo "$(CYAN)Coverage report: $(COVERAGE_DIR)/coverage.html$(RESET)"

test-coverage-unit: ## Run unit tests with coverage
	@echo "$(BLUE)Running unit tests with coverage...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_DIR)/coverage-unit.out -covermode=atomic \
		./internal/domain/... \
		./internal/application/agent/... \
		./internal/application/note/... \
		./internal/application/contact_event/... \
		./internal/application/contact_list/... \
		./internal/application/tracking/... \
		./internal/application/commands/message/...
	@go tool cover -html=$(COVERAGE_DIR)/coverage-unit.out -o $(COVERAGE_DIR)/coverage-unit.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage-unit.out | grep total | awk '{print "$(GREEN)Unit Test Coverage: " $$3 "$(RESET)"}'
	@echo "$(CYAN)Coverage report: $(COVERAGE_DIR)/coverage-unit.html$(RESET)"

clean-coverage: ## Remove coverage reports
	@echo "$(BLUE)Removing coverage reports...$(RESET)"
	@rm -rf $(COVERAGE_DIR)
	@echo "$(GREEN)✓ Coverage reports removed$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔍 Code Quality
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🔍 Code Quality

fmt: ## Format code (go fmt + goimports)
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./... > /dev/null
	@goimports -w . 2>/dev/null || echo "$(YELLOW)⚠️  goimports not found, skipping$(RESET)"
	@echo "$(GREEN)✓ Code formatted$(RESET)"

lint: ## Run golangci-lint
	@echo "$(BLUE)Running linter...$(RESET)"
	@golangci-lint run --timeout 5m || echo "$(YELLOW)⚠️  Some linting issues found$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✓ No issues found$(RESET)"

mod-tidy: ## Clean up go.mod and go.sum
	@echo "$(BLUE)Tidying modules...$(RESET)"
	@go mod tidy
	@echo "$(GREEN)✓ Modules tidied$(RESET)"

mod-download: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	@go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(RESET)"

mod-verify: ## Verify dependencies
	@echo "$(BLUE)Verifying dependencies...$(RESET)"
	@go mod verify
	@echo "$(GREEN)✓ Dependencies verified$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🗄️ Database
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🗄️ Database

migrate-up: ## Run database migrations (up)
	@echo "$(BLUE)Running migrations...$(RESET)"
	@go run cmd/migrate/main.go up
	@echo "$(GREEN)✓ Migrations applied$(RESET)"

migrate-down: ## Rollback last migration
	@echo "$(YELLOW)⚠️  Rolling back last migration...$(RESET)"
	@go run cmd/migrate/main.go down
	@echo "$(GREEN)✓ Migration rolled back$(RESET)"

migrate-status: ## Show migration status
	@go run cmd/migrate/main.go status

migrate-auto: ## Run GORM AutoMigrate (DEV ONLY)
	@echo "$(YELLOW)⚠️  Running GORM AutoMigrate (DEV ONLY)$(RESET)"
	@echo "$(BLUE)This will sync database schema with Go entities...$(RESET)"
	@go run cmd/automigrate/main.go
	@echo "$(GREEN)✅ AutoMigrate completed$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧹 Cleanup
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 🧹 Cleanup

clean: ## Clean everything (containers, volumes, binaries, cache)
	@bash scripts/clean.sh

clean-all: clean clean-coverage clean-bin ## Deep clean (everything including coverage and binaries)
	@echo "$(GREEN)✓ Deep clean completed$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📦 Utilities
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 📦 Utilities

deps: ## Install development dependencies
	@echo "$(BLUE)Installing development dependencies...$(RESET)"
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)✓ Dependencies installed$(RESET)"

version: ## Show Go and tool versions
	@echo "$(BLUE)Versions:$(RESET)"
	@go version
	@echo ""
	@echo "$(CYAN)Project:$(RESET) Ventros CRM"
	@echo "$(CYAN)Module:$(RESET)  github.com/ventros/crm"

check: fmt vet lint ## Run all code quality checks

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📊 Code Analysis (Deterministic Metrics)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

##@ 📊 Code Analysis

analyze: ## Run quick code analysis (bash script, ~2 min)
	@echo "$(BLUE)Running deterministic code analysis...$(RESET)"
	@./scripts/analyze_codebase.sh
	@echo ""
	@echo "$(GREEN)✓ Analysis complete$(RESET)"
	@echo "$(CYAN)Report: ANALYSIS_REPORT.md$(RESET)"

analyze-deep: ## Run deep AST analysis (Go parser, ~30 sec)
	@echo "$(BLUE)Running deep AST analysis...$(RESET)"
	@go run scripts/deep_analyzer.go
	@echo ""
	@echo "$(GREEN)✓ Deep analysis complete$(RESET)"
	@echo "$(CYAN)Report: DEEP_ANALYSIS_REPORT.md$(RESET)"

analyze-all: analyze analyze-deep ## Run all analyses (bash + Go AST)
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(GREEN)✅ ALL ANALYSES COMPLETE$(RESET)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "📄 Reports generated:"
	@echo "  • ANALYSIS_REPORT.md      - Quick metrics (bash)"
	@echo "  • DEEP_ANALYSIS_REPORT.md - AST analysis (Go)"
	@echo "  • ANALYSIS_COMPARISON.md  - Subjective vs Deterministic"
	@echo ""
	@echo "🔍 Key findings:"
	@echo ""
	@grep -A 3 "Optimistic Locking Coverage" ANALYSIS_REPORT.md | head -4
	@echo ""
	@grep "handlers without tenant_id check" DEEP_ANALYSIS_REPORT.md | head -1
	@echo ""
	@echo "Next: Review reports for P0 issues"
	@echo "      $(CYAN)cat DEEP_ANALYSIS_REPORT.md$(RESET)"

analyze-security: analyze-deep ## Show security issues only (BOLA, SQL injection, etc)
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(YELLOW)🔒 SECURITY ANALYSIS SUMMARY$(RESET)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@grep -A 50 "SECURITY ANALYSIS" DEEP_ANALYSIS_REPORT.md | head -50

analyze-ddd: analyze-deep ## Show DDD metrics only (aggregates, events, etc)
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)🏗️  DDD ANALYSIS SUMMARY$(RESET)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@grep -A 80 "DOMAIN-DRIVEN DESIGN" DEEP_ANALYSIS_REPORT.md | head -80

analyze-clean: ## Remove analysis reports
	@echo "$(BLUE)Removing analysis reports...$(RESET)"
	@rm -f ANALYSIS_REPORT.md DEEP_ANALYSIS_REPORT.md
	@echo "$(GREEN)✓ Reports removed$(RESET)"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# End of Makefile
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
