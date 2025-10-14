# 📚 Makefile - Documentation

**Ventros CRM** - Essential Development Commands

---

## 🎯 Quick Reference

```bash
make help          # Show all commands
make clean         # Clean everything (DESTRUCTIVE)
make infra         # Start infrastructure
make api           # Run API locally
make container     # Start everything containerized
make test          # Run tests
make test-e2e      # Run E2E tests
```

---

## 📖 Table of Contents

1. [Development Modes](#-development-modes)
2. [Commands Reference](#-commands-reference)
3. [Workflows](#-workflows)
4. [Troubleshooting](#-troubleshooting)

---

## 🔧 Development Modes

Ventros CRM supports **2 development modes**:

### **Mode 1: Local Development** (Recommended)
- **Infrastructure**: Docker containers (Postgres, RabbitMQ, Redis, Temporal)
- **API**: Go process on host machine
- **Advantages**: Fast hot-reload, easy debugging
- **Use when**: Day-to-day development

### **Mode 2: Full Containerized**
- **Everything**: Docker containers (infra + API)
- **Advantages**: Production-like environment, consistent deployment
- **Use when**: Testing deployment, CI/CD, integration testing

---

## 📝 Commands Reference

### 🧹 Cleanup

#### `make clean`
**Stop and remove EVERYTHING** (API processes + containers + volumes + generated files)

**What it does:**
1. Kills running API processes (local)
2. Stops all Docker containers
3. Removes Docker volumes (**DELETES ALL DATA**)
4. Removes generated files (bin/, coverage)

**When to use:**
- Fresh start needed
- End of work day
- Before switching branches
- When something is broken

**Example:**
```bash
make clean
# ⚠️  CLEANING: API + Infrastructure + Data + Files
# [1/4] Stopping API... ✓
# [2/4] Removing infrastructure... ✓
# [3/4] Cleaning files... ✓
# [4/4] Verifying... ✓
# ✓ CLEAN COMPLETE
```

---

### 🚀 Local Development (Mode 1)

#### `make infra`
**Start infrastructure** (Postgres, RabbitMQ, Redis, Temporal)

**What it starts:**
- PostgreSQL 16 (port 5432)
- RabbitMQ 3.13 (port 5672, UI: 15672)
- Redis 7 (port 6379)
- Temporal (port 7233, UI: 8088)

**Example:**
```bash
make infra
# Starting Infrastructure...
# ✓ Infrastructure ready
#
# Services:
#   • PostgreSQL: localhost:5432
#   • RabbitMQ:   localhost:5672 (UI: http://localhost:15672)
#   • Redis:      localhost:6379
#   • Temporal:   localhost:7233 (UI: http://localhost:8088)
#
# Next: make api
```

**After running:**
- Check health: `docker ps`
- Access RabbitMQ UI: http://localhost:15672 (user: `guest`, pass: `guest`)
- Access Temporal UI: http://localhost:8088

---

#### `make api`
**Run API locally** (requires `make infra` running)

**What it does:**
1. Generates Swagger docs
2. Runs migrations (auto)
3. Starts API server on port 8080

**Example:**
```bash
make api
# Starting API...
#
# Endpoints:
#   • API:     http://localhost:8080
#   • Swagger: http://localhost:8080/swagger/index.html
#   • Health:  http://localhost:8080/health
#
# [GIN-debug] Listening on :8080
```

**After running:**
- Access API: http://localhost:8080
- Access Swagger: http://localhost:8080/swagger/index.html
- Check health: `curl http://localhost:8080/health`

**Hot reload:**
- Just Ctrl+C and run `make api` again
- Or use `air` for automatic reload

---

#### `make build`
**Build API binary**

**Output:** `bin/api`

**Example:**
```bash
make build
# Building binary...
# ✓ Binary: bin/api

# Run it:
./bin/api
```

---

#### `make swagger`
**Generate Swagger documentation**

**What it does:**
- Formats Go comments
- Generates Swagger JSON/YAML
- Updates `docs/` directory

**Example:**
```bash
make swagger
# ✓ Swagger docs generated

# Files updated:
# - docs/docs.go
# - docs/swagger.json
# - docs/swagger.yaml
```

---

#### `make fmt`
**Format code** (go fmt + goimports)

**What it does:**
- Formats all `.go` files
- Organizes imports
- Removes unused imports

**Example:**
```bash
make fmt
# Formatting code...
# ✓ Code formatted
```

---

### 🐳 Container Mode (Mode 2)

#### `make container`
**Start EVERYTHING containerized** (infra + API)

**What it starts:**
- All infrastructure (Postgres, RabbitMQ, Redis, Temporal)
- API container (built from source)

**Example:**
```bash
make container
# Starting containerized stack...
# [+] Building API image...
# [+] Starting containers...
#
# ✓ Stack ready
#
#   • API: http://localhost:8080
#   • Swagger: http://localhost:8080/swagger/index.html
```

**When to use:**
- Testing deployment
- CI/CD pipelines
- Production-like testing
- Sharing environment with team

---

#### `make container-down`
**Stop all containers** (keeps data)

**Example:**
```bash
make container-down
# Stopping containers...
# ✓ Containers stopped
```

**Note:** Data is preserved in volumes. To delete data, use `make clean`.

---

### 🧪 Testing

#### `make test`
**Run unit tests**

**Example:**
```bash
make test
# Running tests...
# ok  	github.com/ventros/crm/internal/domain/contact	0.123s
# ok  	github.com/ventros/crm/internal/domain/message	0.089s
```

---

#### `make test-e2e`
**Run full E2E test** (User → Project → Pipeline → Channel → Messages → Verification)

**What it tests:**
1. Create user (authentication)
2. Create project
3. Create pipeline + statuses
4. Create WhatsApp channel
5. Send all message types (text, image, audio, video, document, location, contact)
6. Verify database (contacts, sessions, messages, events)

**Requirements:**
- API running (`make api`)
- WAHA session `5511999999999` active

**Example:**
```bash
make test-e2e
# 🚀 FULL E2E SYSTEM TEST
#
# This will test:
#   1. Create user (auth)
#   2. Create project
#   3. Create pipeline + statuses
#   4. Create WhatsApp channel
#   5. Send all message types
#   6. Verify database
#
# Requirements:
#   • API running (make api)
#   • WAHA session '5511999999999'
#
# Press Enter to continue or Ctrl+C to cancel...
```

---

## 🔄 Workflows

### 📅 Daily Development Workflow

```bash
# Morning: Start fresh
make clean
make infra
make api

# Code, test, repeat...
# Ctrl+C to stop API
make api  # Restart after changes

# End of day
make clean  # Optional: clean everything
```

---

### 🆕 First Time Setup

```bash
# 1. Clone repository
git clone <repo>
cd ventros-crm

# 2. Install dependencies
go mod download

# 3. Start infrastructure
make infra

# 4. Run API
make api

# 5. Access Swagger
open http://localhost:8080/swagger/index.html

# 6. Run E2E test (optional)
make test-e2e
```

---

### 🐛 Debugging Workflow

```bash
# 1. Clean everything
make clean

# 2. Start infra
make infra

# 3. Build binary
make build

# 4. Run with debugger
dlv exec ./bin/api

# Or use VSCode debugger with:
# - Executable: ./bin/api
# - Working directory: ${workspaceFolder}
```

---

### 🚀 Production Build Workflow

```bash
# 1. Clean
make clean

# 2. Format code
make fmt

# 3. Run tests
make test

# 4. Build binary
make build

# 5. Test containerized
make container

# 6. If OK, tag and push
docker tag ventros-crm:latest ventros-crm:v1.0.0
docker push ventros-crm:v1.0.0
```

---

### 🔄 Switch Branch Workflow

```bash
# Before switching
make clean

# Switch branch
git checkout feature/new-feature

# Restart
make infra
make api
```

---

## 🆘 Troubleshooting

### Problem: "Port 8080 already in use"

**Solution:**
```bash
# Kill process using port 8080
lsof -i :8080
kill -9 <PID>

# Or use make clean
make clean
```

---

### Problem: "Cannot connect to Docker daemon"

**Solution:**
```bash
# Start Docker Desktop (macOS/Windows)
# Or start Docker service (Linux)
sudo systemctl start docker

# Verify
docker ps
```

---

### Problem: "Migrations failed"

**Solution:**
```bash
# Clean database
make clean

# Restart
make infra
make api

# Migrations run automatically on API startup
```

---

### Problem: "WAHA session not found"

**Solution:**
```bash
# Check WAHA is running
curl http://localhost:3000/api/sessions

# Create session 5511999999999
# (See WAHA documentation)
```

---

### Problem: "Tests failing"

**Solution:**
```bash
# Clean everything
make clean

# Restart infra
make infra

# Wait 10 seconds for services to be ready
sleep 10

# Run tests
make test
```

---

### Problem: "Out of disk space"

**Solution:**
```bash
# Remove all Docker volumes
make clean

# Remove unused Docker images
docker system prune -a

# Check disk space
df -h
```

---

## 📊 Command Comparison

| Command | Infrastructure | API | Data | Use Case |
|---------|---------------|-----|------|----------|
| `make infra` | ✅ Starts | ❌ Manual | ✅ Keeps | Daily dev |
| `make api` | ⚠️ Requires | ✅ Starts | ✅ Keeps | Daily dev |
| `make container` | ✅ Starts | ✅ Starts | ✅ Keeps | Production-like |
| `make container-down` | ❌ Stops | ❌ Stops | ✅ Keeps | Pause work |
| `make clean` | ❌ Removes | ❌ Kills | ❌ Deletes | Fresh start |

---

## 🎯 Best Practices

### ✅ DO

- Run `make clean` before switching branches
- Run `make fmt` before committing
- Run `make test` before pushing
- Use `make infra` + `make api` for daily dev
- Use `make container` for deployment testing

### ❌ DON'T

- Don't run `make api` without `make infra`
- Don't run multiple `make api` instances (port conflict)
- Don't commit without running `make fmt`
- Don't push without running `make test`

---

## 📖 Additional Documentation

- [README.md](README.md) - Project overview
- [ARCHITECTURE.md](ARCHITECTURE.md) - Architecture documentation
- [MIGRATIONS.md](MIGRATIONS.md) - Database migrations guide
- [DEV_GUIDE.md](DEV_GUIDE.md) - Development guide

---

## 🤝 Contributing

When adding new make targets:

1. Keep it essential (avoid bloat)
2. Document in this file
3. Add `## Comment` for `make help`
4. Test on clean environment
5. Update workflows if needed

---

**Last Updated**: 2025-10-12
**Version**: 2.0 (Simplified)
