# 🚀 Getting Started with Ventros CRM

This guide will help you set up Ventros CRM for local development.

---

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

### Required
- **Go 1.25.1+** - [Download](https://golang.org/dl/)
- **Docker 24.0+** & **Docker Compose v2** - [Download](https://docs.docker.com/get-docker/)
- **Git** - [Download](https://git-scm.com/downloads)

### Optional (for manual setup)
- **PostgreSQL 16** - [Download](https://www.postgresql.org/download/)
- **RabbitMQ 3.12+** - [Download](https://www.rabbitmq.com/download.html)
- **Redis 7.0+** - [Download](https://redis.io/download)
- **Temporal Server** - [Install Guide](https://docs.temporal.io/cli)

---

## ⚡ Quick Start (5 minutes)

### 1. Clone the Repository

```bash
git clone https://github.com/caloi/ventros-crm.git
cd ventros-crm
```

### 2. Run Setup Script

```bash
# Run automated setup (creates .env, starts infrastructure, runs migrations)
make setup

# Or manually:
cp .env.example .env
make infra-up
make migrate
make seed
```

### 3. Start the API

```bash
make run
# or
go run cmd/api/main.go
```

### 4. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Swagger UI
open http://localhost:8080/swagger/index.html
```

✅ **You're ready to go!** Jump to [First Steps](#-first-steps) to create your first contact.

---

## 🔧 Detailed Setup

### Step 1: Environment Configuration

Copy the example environment file and configure:

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```bash
# Server
SERVER_PORT=8080
SERVER_ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=ventros
DB_PASSWORD=ventros123
DB_NAME=ventros_crm
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Temporal
TEMPORAL_HOST=localhost:7233
TEMPORAL_NAMESPACE=default

# JWT
JWT_SECRET=your-secret-key-change-in-production

# WAHA (WhatsApp HTTP API)
WAHA_BASE_URL=http://localhost:3000
WAHA_API_KEY=your-waha-api-key
```

### Step 2: Start Infrastructure

Using Docker Compose (recommended):

```bash
make infra-up
```

This starts:
- PostgreSQL 16 on port 5432
- RabbitMQ on port 5672 (Management UI: 15672)
- Redis on port 6379
- Temporal Server on port 7233 (Web UI: 8233)

Verify services are running:

```bash
docker-compose ps
```

### Step 3: Database Migration

Run GORM auto-migrations:

```bash
make migrate
```

This creates all tables with RLS policies enabled.

### Step 4: Seed Initial Data

Seed the database with:
- Admin user
- Default project
- Channel types
- Sample contacts (optional)

```bash
make seed
```

### Step 5: Run the Application

```bash
# Development mode with hot reload
make run

# Or build and run
make build
./bin/ventros-crm
```

The API will start on `http://localhost:8080`

---

## 🎯 First Steps

### 1. Get an Auth Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@ventros.com",
    "password": "admin123"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "...",
    "email": "admin@ventros.com",
    "role": "admin"
  }
}
```

Save the token for subsequent requests.

### 2. Create a Project

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My First Project",
    "description": "Test project"
  }'
```

### 3. Create a Contact

```bash
curl -X POST "http://localhost:8080/api/v1/contacts?project_id=YOUR_PROJECT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+5511999999999",
    "tags": ["lead", "whatsapp"]
  }'
```

### 4. List Contacts

```bash
curl "http://localhost:8080/api/v1/contacts?project_id=YOUR_PROJECT_ID" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Test WhatsApp Integration (Optional)

If you have WAHA running:

```bash
# Simulate incoming message from WAHA
curl -X POST http://localhost:8080/webhooks/waha \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message",
    "session": "default",
    "payload": {
      "from": "5511999999999@c.us",
      "body": "Hello!",
      "timestamp": 1234567890
    }
  }'
```

---

## 🧪 Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/domain/contact/...

# Run integration tests
make test-integration

# Run E2E tests (requires infrastructure)
make test-e2e
```

---

## 🐛 Troubleshooting

### Database Connection Error

**Problem**: `pq: database "ventros_crm" does not exist`

**Solution**:
```bash
# Create database manually
docker exec -it ventros-postgres psql -U ventros -c "CREATE DATABASE ventros_crm;"

# Then run migrations
make migrate
```

### RabbitMQ Connection Error

**Problem**: `dial tcp: connection refused`

**Solution**:
```bash
# Check RabbitMQ is running
docker-compose ps rabbitmq

# Restart if needed
docker-compose restart rabbitmq

# Check logs
docker-compose logs rabbitmq
```

### Temporal Not Starting

**Problem**: Temporal workflows not executing

**Solution**:
```bash
# Check Temporal is running
docker-compose ps temporal

# Access Temporal Web UI
open http://localhost:8233

# Check worker logs
docker-compose logs temporal-worker
```

### Port Already in Use

**Problem**: `bind: address already in use`

**Solution**:
```bash
# Find process using port
lsof -i :8080

# Kill process (replace PID)
kill -9 PID

# Or change port in .env
SERVER_PORT=8081
```

### Go Module Issues

**Problem**: `go: module not found`

**Solution**:
```bash
# Download dependencies
go mod download

# Tidy modules
go mod tidy

# Verify modules
go mod verify
```

For more troubleshooting tips, see [troubleshooting.md](troubleshooting.md)

---

## 🔄 Development Workflow

### Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes and test**
   ```bash
   make test
   make lint
   ```

3. **Commit with conventional commits**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/my-feature
   ```

### Hot Reload (Development)

Install Air for hot reload:
```bash
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

### Debugging with Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug API
dlv debug cmd/api/main.go

# Or attach to running process
dlv attach PID
```

---

## 📚 Next Steps

Now that you have Ventros CRM running:

1. **Explore the API** - Check out [Swagger UI](http://localhost:8080/swagger/index.html)
2. **Read Architecture** - Understand the [architecture](../architecture/README.md)
3. **Try Examples** - Check [code examples](../code-examples/)
4. **Deploy** - Follow [deployment guide](../deployment/)

---

## 💡 Useful Commands

```bash
# Development
make run              # Run application
make build            # Build binary
make test             # Run tests
make lint             # Run linters
make fmt              # Format code

# Infrastructure
make infra-up         # Start all services
make infra-down       # Stop all services
make infra-logs       # View logs

# Database
make migrate          # Run migrations
make migrate-down     # Rollback migrations
make seed             # Seed data
make db-reset         # Drop and recreate DB

# Docker
make docker-build     # Build Docker image
make docker-run       # Run in Docker
make docker-clean     # Clean Docker artifacts

# Kubernetes
make k8s-deploy       # Deploy to K8s
make k8s-delete       # Delete from K8s
make k8s-logs         # View K8s logs

# Documentation
make swagger          # Generate Swagger docs
make docs             # Generate all docs

# Help
make help             # Show all commands
```

---

## 🆘 Need Help?

- 📖 [Documentation](../README.md)
- 🐛 [Report a Bug](https://github.com/caloi/ventros-crm/issues/new?template=bug_report.md)
- 💡 [Request a Feature](https://github.com/caloi/ventros-crm/issues/new?template=feature_request.md)
- 💬 [GitHub Discussions](https://github.com/caloi/ventros-crm/discussions)
- 📧 [Email Support](mailto:support@ventros.com)

---

**Happy coding! 🚀**
