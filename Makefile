.PHONY: help run build test clean docker-up docker-down migrate seedr-logs clean test dev-up dev-down dev-restart dev-logs dev-clean

# Load asdf
SHELL := /bin/bash
.SHELLFLAGS := -c '. $$HOME/.asdf/asdf.sh && exec bash -c "$$@"' --

run:
	@echo "🔄 Checking if Ent code is up to date..."
	@make -s ent-generate
	@echo "Starting API server..."
	go run cmd/api/main.go

dev-run:
	@echo "🔄 Preparing dev environment..."
	@echo "🔄 Running GORM database migrations..."
	@go run cmd/migrate-gorm/main.go
	@echo "🔒 Setting up RLS (Row Level Security)..."
	@PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -f scripts/setup-rls.sql -q
	@echo "📚 Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Ready! Starting API server..."
	@echo "   Swagger UI: http://localhost:8080/swagger/index.html"
	@echo ""
	go run cmd/api/main.go

swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs
	@echo "✅ Swagger docs available at http://localhost:8080/swagger/index.html"

swagger-install:
	@echo "Installing Swagger CLI..."
	go install github.com/swaggo/swag/cmd/swag@latest

# Development environment (without API)
dev-down:
	@echo "Stopping dev services..."
	docker-compose -f deployments/docker/docker-compose.dev.yml down

dev-up:
	@echo "Starting dev services (postgres, rabbitmq, redis, temporal)..."
	docker-compose -f deployments/docker/docker-compose.dev.yml up -d
	@echo ""
	@echo "✓ Dev services started!"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - RabbitMQ: localhost:5672 (Management UI: http://localhost:15672)"
	@echo "  - Redis: localhost:6379"
	@echo "  - Temporal: localhost:7233 (UI: http://localhost:8088)"
	@echo ""
	@echo "Run 'make run' to start the API locally"
	@echo "Run 'make health' to check service health"

dev-restart:
	@echo "Restarting dev services..."
	docker-compose -f deployments/docker/docker-compose.dev.yml restart

dev-logs:
	docker-compose -f deployments/docker/docker-compose.dev.yml logs -f

dev-clean:
	@echo "Stopping and removing dev containers and volumes..."
	docker-compose -f deployments/docker/docker-compose.dev.yml down -v

# Production environment (with API)
docker-up:
	@echo "Starting Docker containers..."
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-rebuild:
	@echo "Rebuilding and starting Docker containers..."
	docker-compose -f deployments/docker/docker-compose.yml up -d --build

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose -f deployments/docker/docker-compose.yml down

docker-clean:
	@echo "Stopping and removing Docker containers, volumes..."
	docker-compose -f deployments/docker/docker-compose.yml down -v

docker-logs:
	docker-compose -f deployments/docker/docker-compose.yml logs -f

docker-restart:
	@echo "Restarting Docker containers..."
	docker-compose -f deployments/docker/docker-compose.yml restart

clean:
	@echo "Cleaning build artifacts..."
	rm -rf docs/swagger
	rm -f cmd/api/api

test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building binary..."
	go build -o ventros-crm cmd/api/main.go

deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Install development tools (containerized approach)
install-tools: ## Install development tools locally
	@echo "🔧 Installing development tools..."
	@mkdir -p bin/
	@GOBIN=$(PWD)/bin go install ariga.io/entimport/cmd/entimport@latest
	@GOBIN=$(PWD)/bin go install entgo.io/ent/cmd/ent@latest
	@GOBIN=$(PWD)/bin go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Tools installed in ./bin/"
	@echo "💡 Add ./bin to your PATH or use make commands"

# Ent Code Generation
ent-init:
	@echo "Initializing Ent..."
	@mkdir -p ent/schema
	@go run -mod=mod entgo.io/ent/cmd/ent init --target ent/schema User

ent-generate:
	@echo "Generating Ent code..."
	@go generate ./ent/schema

ent-new:
	@echo "Creating new Ent schema..."
	@echo "Usage: make ent-new SCHEMA=Contact"
	@if [ -z "$(SCHEMA)" ]; then \
		echo "Error: SCHEMA is required. Example: make ent-new SCHEMA=Contact"; \
		exit 1; \
	fi
	@go run -mod=mod entgo.io/ent/cmd/ent init --target ent/schema $(SCHEMA)

# Database migrations
migrate: ## Run database migrations
	@echo "🔄 Running database migrations..."
	@make -s ent-generate
	@go run cmd/migrate/main.go

# RLS Setup
setup-rls: ## Setup Row Level Security
	@echo "🔒 Setting up RLS (Row Level Security)..."
	@PGPASSWORD=ventros123 psql -h localhost -U ventros -d ventros_crm -f scripts/setup-rls.sql
	@echo "✅ RLS configured successfully!"

# Import schemas from existing database
import-schemas: ## Import Ent schemas from existing database
	@echo "📥 Importing schemas from database..."
	@if [ -f "./bin/entimport" ]; then \
		./bin/entimport \
			-dsn "postgres://ventros:ventros123@localhost:5432/ventros_crm?sslmode=disable" \
			-schema-path "./ent/schema"; \
	else \
		go run ariga.io/entimport/cmd/entimport@latest \
			-dsn "postgres://ventros:ventros123@localhost:5432/ventros_crm?sslmode=disable" \
			-schema-path "./ent/schema"; \
	fi
	@echo "✅ Schemas imported successfully!"

# Database Sync (overrides DB with Domain schema)
db-sync:
	@echo "⚠️  This will DROP tables not declared in Ent schema!"
	@echo ""
	@echo "Press Ctrl+C to cancel, or wait 5 seconds to continue..."
	@sleep 5
	@echo ""
	@echo "🔄 Running database sync..."
	@make -s ent-generate
	@go run cmd/migrate/main.go

seed: ## Run database seeds
	@echo "Running database seeds..."
	@chmod +x scripts/run-seeds.sh
	@./scripts/run-seeds.sh
	@echo "✅ Database seeded successfully!"

# Atlas Installation
atlas-install:
	@echo "Installing Atlas CLI..."
	@atlas version

# Database migrations
migrate-create:
	@echo "Creating new migration..."
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Example: make migrate-create NAME=add_users"; \
		exit 1; \
	fi
	@atlas migrate diff $(NAME) \
		--dir "file://ent/migrate/migrations" \
		--to "ent://ent/schema" \
		--dev-url "docker://postgres/15/test?search_path=public"

migrate-apply:
	@echo "Applying migrations..."
	@atlas migrate apply \
		--dir "file://ent/migrate/migrations" \
		--url "postgres://ventros:ventros123@localhost:5432/ventros_crm?sslmode=disable"

migrate-status:
	@echo "Checking migration status..."
	@atlas migrate status \
		--dir "file://ent/migrate/migrations" \
		--url "postgres://ventros:ventros123@localhost:5432/ventros_crm?sslmode=disable"

migrate-down:
	@echo "Rolling back last migration..."
	@atlas migrate down \
		--dir "file://ent/migrate/migrations" \
		--url "postgres://ventros:ventros123@localhost:5432/ventros_crm?sslmode=disable"

health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/ready | jq . || echo "Service not running or unhealthy"

health-live:
	@curl -s http://localhost:8080/live | jq .

health-basic:
	@curl -s http://localhost:8080/health | jq .

health-db:
	@echo "Checking database health..."
	@curl -s http://localhost:8080/health/database | jq .

health-migrations:
	@echo "Checking migrations status..."
	@curl -s http://localhost:8080/health/migrations | jq .

health-redis:
	@echo "Checking Redis health..."
	@curl -s http://localhost:8080/health/redis | jq .

health-rabbitmq:
	@echo "Checking RabbitMQ health..."
	@curl -s http://localhost:8080/health/rabbitmq | jq .

health-temporal:
	@echo "Checking Temporal health..."
	@curl -s http://localhost:8080/health/temporal | jq .

# Webhook management
webhook-events:
	@echo "Available WAHA events for webhook subscription:"
	@curl -s http://localhost:8080/api/v1/webhook-subscriptions/available-events | jq .

webhook-list:
	@echo "Listing webhook subscriptions:"
	@curl -s http://localhost:8080/api/v1/webhook-subscriptions | jq .

webhook-list-active:
	@echo "Listing active webhook subscriptions:"
	@curl -s http://localhost:8080/api/v1/webhook-subscriptions?active=true | jq .

webhook-create:
	@echo "Creating webhook subscription..."
	@if [ -z "$(URL)" ]; then \
		echo "Error: URL is required. Example: make webhook-create URL=https://n8n.example.com/webhook/waha EVENTS=message,call.received"; \
		exit 1; \
	fi
	@EVENTS_JSON=$$(echo "$(EVENTS)" | sed 's/,/","/g' | sed 's/^/"/' | sed 's/$$/"/'); \
	curl -X POST http://localhost:8080/api/v1/webhook-subscriptions \
		-H "Content-Type: application/json" \
		-d "{ \
			\"name\": \"$(NAME)\", \
			\"url\": \"$(URL)\", \
			\"events\": [$$EVENTS_JSON] \
		}" | jq .

webhook-get:
	@echo "Getting webhook subscription..."
	@if [ -z "$(ID)" ]; then \
		echo "Error: ID is required. Example: make webhook-get ID=uuid-here"; \
		exit 1; \
	fi
	@curl -s http://localhost:8080/api/v1/webhook-subscriptions/$(ID) | jq .

webhook-delete:
	@echo "Deleting webhook subscription..."
	@if [ -z "$(ID)" ]; then \
		echo "Error: ID is required. Example: make webhook-delete ID=uuid-here"; \
		exit 1; \
	fi
	@curl -X DELETE http://localhost:8080/api/v1/webhook-subscriptions/$(ID)
	@echo "✅ Webhook deleted"

webhook-test-n8n:
	@echo "Creating test N8N webhook subscription..."
	@curl -X POST http://localhost:8080/api/v1/webhook-subscriptions \
		-H "Content-Type: application/json" \
		-d '{ \
			"name": "N8N Test Webhook", \
			"url": "http://localhost:5678/webhook-test/waha", \
			"events": ["message", "message.ack", "call.received"], \
			"retry_count": 3, \
			"timeout_seconds": 30 \
		}' | jq .

# Queue management
queues:
	@echo "Listing RabbitMQ queues:"
	@curl -s http://localhost:8080/api/v1/queues | jq .

help:
	@echo "Available targets:"
	@echo ""
	@echo "Development:"
	@echo "  dev-down        - Stop dev services (first step)"
	@echo "  dev-up          - Start dev services: postgres, rabbitmq, redis, temporal (without API)"
	@echo "  dev-restart     - Restart dev services"
	@echo "  dev-logs        - View dev services logs"
	@echo "  dev-clean       - Stop and remove dev containers + volumes"
	@echo "  run             - Run the API server locally (with ent-generate)"
	@echo "  dev-run         - Full dev setup: ent-generate + migrate + swagger + run"
	@echo ""
	@echo "Production:"
	@echo "  docker-up       - Start all Docker containers (including API)"
	@echo "  docker-rebuild  - Rebuild and start Docker containers"
	@echo "  docker-down     - Stop Docker containers"
	@echo "  docker-clean    - Stop and remove containers + volumes"
	@echo "  docker-restart  - Restart Docker containers"
	@echo "  docker-logs     - View Docker logs"
	@echo ""
	@echo "Ent & Migrations:"
	@echo "  ent-init        - Initialize Ent (first time only)"
	@echo "  ent-new SCHEMA=Name - Create new Ent schema"
	@echo "  ent-generate    - Generate Ent code from schemas"
	@echo "  migrate         - Run database migrations (ent-generate + migrate)"
	@echo "  setup-rls       - Setup Row Level Security (RLS) policies"
	@echo "  import-schemas  - Import Ent schemas from existing database"
	@echo "  db-sync         - ⚠️  Sync DB with domain (drops non-declared tables!)"
	@echo "  atlas-install   - Install Atlas CLI"
	@echo "  migrate-create NAME=name - Create new migration"
	@echo "  migrate-apply   - Apply pending migrations"
	@echo "  migrate-status  - Check migration status"
	@echo "  migrate-down    - Rollback last migration"
	@echo ""
	@echo "Health Checks:"
	@echo "  health          - Check all dependencies (aggregated)"
	@echo "  health-live     - Liveness check (is service running?)"
	@echo "  health-basic    - Basic health check"
	@echo "  health-db       - Check database only"
	@echo "  health-migrations - Check migration status"
	@echo "  health-redis    - Check Redis only"
	@echo "  health-rabbitmq - Check RabbitMQ only"
	@echo "  health-temporal - Check Temporal only"
	@echo ""
	@echo "Queue Management:"
	@echo "  queues          - List all RabbitMQ queues with stats"
	@echo ""
	@echo "Other:"
	@echo "  swagger         - Generate Swagger docs"
	@echo "  build           - Build binary"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  install-tools   - Install dev tools locally (containerized approach)"
