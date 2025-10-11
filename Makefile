# Ventros CRM - Makefile
# Professional development and deployment automation

# Container runtime (docker or podman)
CONTAINER_RUNTIME ?= docker
COMPOSE = $(CONTAINER_RUNTIME) compose

# Color output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
RESET := \033[0m

.DEFAULT_GOAL := help
.PHONY: help

##@ 📚 Help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BLUE)Ventros CRM - Available Commands$(RESET)\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 🚀 Quick Start

dev: infra api ## [RECOMMENDED] Start infrastructure + API (separate terminals)

infra: ## Start infrastructure only (PostgreSQL, RabbitMQ, Redis, Temporal)
	@echo "$(BLUE)Starting Infrastructure...$(RESET)"
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml up -d
	@echo "$(GREEN)✓ Infrastructure ready$(RESET)"
	@echo ""
	@echo "Services:"
	@echo "  • PostgreSQL: localhost:5432"
	@echo "  • RabbitMQ:   localhost:5672 (UI: http://localhost:15672)"
	@echo "  • Redis:      localhost:6379"
	@echo "  • Temporal:   localhost:7233 (UI: http://localhost:8088)"
	@echo ""
	@echo "Next: $(GREEN)make api$(RESET)"

api: swagger ## Run API (requires infra running)
	@echo "$(BLUE)Starting API...$(RESET)"
	@echo ""
	@echo "Endpoints:"
	@echo "  • API:     http://localhost:8080"
	@echo "  • Swagger: http://localhost:8080/swagger/index.html"
	@echo "  • Health:  http://localhost:8080/health"
	@echo ""
	@go run cmd/api/main.go

##@ 🧪 Testing

test: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(RESET)"
	@go test -v -race ./...

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(RESET)"

test-domain: ## Run domain tests with coverage
	@echo "$(BLUE)Running domain tests...$(RESET)"
	@go test -v -race -coverprofile=coverage-domain.out ./internal/domain/...
	@go tool cover -func=coverage-domain.out

setup-all-complete: ## ⭐ FULL E2E TEST: User + Project + Pipeline + Channel + All Message Types + Verification
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo "$(BLUE)       🚀 FULL E2E SYSTEM TEST$(RESET)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(RESET)"
	@echo ""
	@echo "$(GREEN)This will test the COMPLETE system:$(RESET)"
	@echo "  1. Create user (auth)"
	@echo "  2. Create project"
	@echo "  3. Create pipeline with statuses"
	@echo "  4. Create WhatsApp channel"
	@echo "  5. Send all message types (text, image, audio, video, document, location, contact)"
	@echo "  6. Verify database (contacts, sessions, messages, events)"
	@echo ""
	@echo "$(YELLOW)Requirements:$(RESET)"
	@echo "  • API running (make api)"
	@echo "  • WAHA running with session '5511999999999'"
	@echo ""
	@read -p "Press Enter to continue or Ctrl+C to cancel..."
	@echo ""
	@bash scripts/setup-all-complete.sh

##@ 🛠️ Development

build: ## Build binary
	@echo "$(BLUE)Building binary...$(RESET)"
	@go build -o bin/api cmd/api/main.go
	@echo "$(GREEN)✓ Binary: bin/api$(RESET)"

swagger: ## Generate Swagger documentation
	@echo "$(BLUE)Generating Swagger docs...$(RESET)"
	@swag fmt
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "$(GREEN)✓ Swagger docs generated$(RESET)"

lint: ## Run linters
	@echo "$(BLUE)Running linters...$(RESET)"
	@golangci-lint run ./...

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	@go fmt ./...
	@goimports -w .

deps: ## Update dependencies
	@echo "$(BLUE)Updating dependencies...$(RESET)"
	@go mod tidy
	@go mod vendor

##@ 🐳 Docker/Podman

container: ## Start full containerized stack
	@echo "$(BLUE)Starting containerized stack...$(RESET)"
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml up -d --build
	@echo "$(GREEN)✓ Stack ready at http://localhost:8080$(RESET)"

container-stop: ## Stop containers
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml down

container-logs: ## Show container logs
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml logs -f

##@ ☸️ Kubernetes

k8s: ## Deploy to Kubernetes (requires minikube/k3s)
	@echo "$(BLUE)Deploying to Kubernetes...$(RESET)"
	@helm upgrade --install ventros-crm .deploy/helm/ventros-crm \
		--namespace ventros-crm --create-namespace \
		--wait --timeout 5m
	@echo "$(GREEN)✓ Deployed to Kubernetes$(RESET)"
	@echo ""
	@echo "Access API:"
	@echo "  kubectl port-forward -n ventros-crm svc/ventros-crm 8080:8080"

k8s-delete: ## Delete from Kubernetes
	@helm uninstall ventros-crm --namespace ventros-crm || true
	@kubectl delete namespace ventros-crm || true

k8s-logs: ## Show Kubernetes logs
	@kubectl logs -n ventros-crm -l app=ventros-crm -f

k8s-status: ## Show Kubernetes status
	@kubectl get all -n ventros-crm

##@ 🗄️ Database

db-migrate: ## Run database migrations
	@echo "$(BLUE)Running migrations...$(RESET)"
	@echo "$(YELLOW)Note: Migrations run automatically on API startup$(RESET)"
	@echo "Use this only for manual migration testing"
	@go run cmd/migrate/main.go up

db-rollback: ## Rollback last migration
	@echo "$(BLUE)Rolling back migration...$(RESET)"
	@go run cmd/migrate/main.go down

db-status: ## Show migration status
	@go run cmd/migrate/main.go status

db-seed: ## Seed database with test data
	@echo "$(BLUE)Seeding database...$(RESET)"
	@go run cmd/seed/main.go

##@ 🛑 Stop & Clean

infra-stop: ## Stop infrastructure
	@echo "$(BLUE)Stopping infrastructure...$(RESET)"
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml down
	@echo "$(GREEN)✓ Infrastructure stopped$(RESET)"

infra-clean: ## Stop and remove volumes (DESTRUCTIVE - deletes all data)
	@echo "$(RED)⚠️  WARNING: This will DELETE ALL DATA!$(RESET)"
	@echo "  • PostgreSQL (tables, data)"
	@echo "  • RabbitMQ (queues, messages)"
	@echo "  • Redis (cache)"
	@echo "  • Temporal (workflows)"
	@echo ""
	@read -p "Press Enter to continue or Ctrl+C to cancel..."
	@echo ""
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml down -v
	@echo "$(GREEN)✓ All data removed$(RESET)"

infra-reset: infra-clean infra ## Clean and restart infrastructure (fresh start)

clean: ## Remove generated files
	@echo "$(BLUE)Cleaning generated files...$(RESET)"
	@rm -f bin/api coverage.out coverage.html
	@rm -rf vendor/
	@echo "$(GREEN)✓ Clean$(RESET)"

##@ 📊 Monitoring

health: ## Check API health
	@echo "$(BLUE)Checking API health...$(RESET)"
	@curl -s http://localhost:8080/health | jq . || echo "$(RED)API not reachable$(RESET)"

logs: ## Show API logs (if running in background)
	@tail -f logs/api.log 2>/dev/null || echo "$(YELLOW)No log file found. Use 'make api' to start API$(RESET)"

ps: ## List running containers
	@$(COMPOSE) --env-file .deploy/container/.env -f .deploy/container/compose.api.yaml ps
