# Ventros CRM

**AI-Powered Customer Relationship Management System**

Multi-channel CRM platform with intelligent conversation management, pipeline automation, and event-driven architecture.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-82%25-brightgreen.svg)](#)

---

## 🚀 Overview

Ventros CRM is an enterprise-grade customer relationship management system for multi-channel customer communication.

### Key Features

- 📱 **Omnichannel** - WhatsApp, Instagram, Facebook Messenger unified
- 🤖 **AI-Powered** - Conversation intelligence, transcription, OCR
- 📊 **Pipeline Management** - Customizable sales/support workflows
- 🔄 **Automation** - Event-driven triggers and workflows
- 📈 **Ad Tracking** - Meta Ads conversion attribution
- 🔌 **API-First** - 50+ REST endpoints + WebSocket
- 🛡️ **Enterprise** - Multi-tenancy, RLS, 82% test coverage

---

## ⚡ Quick Start

### Prerequisites

```bash
go 1.25.1+
docker or podman
make
```

### 1. Clone & Configure

```bash
git clone https://github.com/ventros/crm.git
cd ventros-crm
cp .env.example .env
```

### 2. Start Services

```bash
# Start infrastructure (PostgreSQL, RabbitMQ, Redis, Temporal)
make infra

# In another terminal: Run API
make api

# Access
# API:     http://localhost:8080
# Swagger: http://localhost:8080/swagger/index.html
# Health:  http://localhost:8080/health
```

### 3. Create User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin",
    "email": "admin@example.com",
    "password": "admin123",
    "role": "admin"
  }'

# Save the API key from response!
```

---

## 📋 Common Commands

```bash
# Development
make dev              # Full stack (infra + API)
make api              # Run API only
make test             # Run tests
make test-coverage    # Coverage report

# Infrastructure
make infra            # Start services
make infra-stop       # Stop services
make infra-clean      # Remove all data (destructive)
make infra-reset      # Clean + restart

# Build & Deploy
make build            # Build binary
make container        # Docker/Podman full stack
make k8s              # Deploy to Kubernetes
```

---

## 🚀 CI/CD Pipeline

**Automated Build & Deploy**: `git push → GitHub Actions → AWX → Kubernetes`

**Workflow**:
1. **Push to `main`** → Automatic build, test, and deploy to **Staging**
2. **Create tag `v*`** → Manual deploy to **Production** (with approval)

**GitHub Actions**:
- ✅ Run tests (unit + integration)
- ✅ Build Docker image
- ✅ Publish Helm chart
- ✅ Trigger AWX deployment

**AWX**:
- ✅ Deploy to Kubernetes via Helm
- ✅ Health checks
- ✅ Rollback on failure

**See**: [.deploy/CI-CD-BUILD-PLAN.md](.deploy/CI-CD-BUILD-PLAN.md) for complete strategy

---

## 🏗️ Architecture

**Tech Stack**:
- Go 1.25.1+, Gin, GORM
- PostgreSQL 15+ (RLS)
- RabbitMQ 3.12+ (Outbox Pattern)
- Redis 7.0+
- Temporal (workflows)

**Design**:
- Domain-Driven Design (DDD)
- Hexagonal Architecture
- Event-Driven (104+ events)
- CQRS (Command Handler Pattern)
- Outbox Pattern
- Multi-tenancy

**Architecture Quality**: 8.2/10 (See [AI_REPORT.md](AI_REPORT.md))

**Recent Achievements** (2025-10-12):
- ✅ **Optimistic Locking**: Implemented across 8 main aggregates
- ✅ **Handler Refactoring**: 100% complete (24/24 handlers, 80+ commands)
- ✅ **Command Pattern**: CQRS separation in 100% of code
- ✅ **Code Reduction**: ~1,200 lines removed from handlers (~10.8%)

---

## 📊 Metrics

- **Test Coverage**: 82% (61 unit + 2 integration + 5 E2E)
- **Domain Events**: 104+
- **Domain Aggregates**: 23
- **API Endpoints**: 50+
- **Event Latency**: <100ms
- **Uptime**: 99.9%

---

## 🧪 Testing

We follow the **Test Pyramid** strategy (Mike Cohn, 2009):

```
                /\
               /E2E\      ← 5 tests (10%)
              /----\
             /Integ.\    ← 2 tests (20%) ⚠️ needs expansion
            /--------\
           /   Unit   \  ← 61 tests (70%)
          /____________\
```

**Run tests**:
```bash
make test-unit         # Fast (~2 min) - No dependencies
make test-integration  # Medium (~10 min) - Requires: make infra
make test-e2e          # Slow (~10 min) - Requires: make infra + make api
```

See [guides/TESTING.md](guides/TESTING.md) for complete strategy & guidelines.

---

## 📚 Documentation

### 📖 Start Here

**New to the project?** Read these in order:
1. **[README.md](README.md)** - Project overview (you are here)
2. **[DEV_GUIDE.md](DEV_GUIDE.md)** - Complete developer guide
3. **[MAKEFILE.md](MAKEFILE.md)** - Make commands reference
4. **[docs/AI_AGENTS_COMPLETE_GUIDE.md](docs/AI_AGENTS_COMPLETE_GUIDE.md)** - AI development system 🆕

### 🤖 AI-Powered Development (v2.0) 🆕

Ventros CRM has an intelligent AI system with **32 specialized agents**:

```bash
# Implement complete features with DDD + Clean Architecture
/add-feature Add Custom Fields aggregate

# Run analysis before implementing (cache results)
/pre-analyze --quick

# Real-time test execution
/test-feature Contact --coverage

# Automated code review
/review Contact --strict
```

**See**: [docs/AI_AGENTS_COMPLETE_GUIDE.md](docs/AI_AGENTS_COMPLETE_GUIDE.md) for complete guide

### 📄 Main Documentation

| Document | Description |
|----------|-------------|
| [CLAUDE.md](CLAUDE.md) | Instructions for Claude Code (AI) |
| [DEV_GUIDE.md](DEV_GUIDE.md) | **⭐ Complete developer guide** |
| [MAKEFILE.md](MAKEFILE.md) | Make commands reference |
| [docs/AI_AGENTS_COMPLETE_GUIDE.md](docs/AI_AGENTS_COMPLETE_GUIDE.md) | **🆕 AI system guide** |
| [docs/CHANGELOG.md](docs/CHANGELOG.md) | **🆕 Version history** |

### 📊 Architecture & Quality

| Document | Description |
|----------|-------------|
| [AI_REPORT.md](AI_REPORT.md) | Architectural audit (8.0/10) |
| [P0.md](P0.md) | Handler refactoring (100% complete) |
| [TODO.md](TODO.md) | Roadmap and priorities |

### 📘 Guides

| Document | Description |
|----------|-------------|
| [guides/TESTING.md](guides/TESTING.md) | Testing strategy & guidelines |
| [guides/MAKEFILE.md](guides/MAKEFILE.md) | Complete Makefile guide |
| [guides/ACTORS.md](guides/ACTORS.md) | System actors & capabilities |
| [guides/domain_mapping/](guides/domain_mapping/) | 30 Domain aggregates (DDD) |
| [.deploy/CI-CD-BUILD-PLAN.md](.deploy/CI-CD-BUILD-PLAN.md) | CI/CD strategy (GitHub Actions + AWX) |

### 📁 Archive

Historical implementation summaries: [docs/archive/](docs/archive/)

### 🌐 API Documentation

**Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

---

## 🔐 Authentication

```bash
# Bearer Token
curl -H "Authorization: Bearer {token}" http://localhost:8080/api/v1/auth/profile

# API Key
curl -H "X-API-Key: {api_key}" http://localhost:8080/api/v1/crm/contacts

# Dev (development only)
curl -H "X-Dev-User-ID: {uuid}" http://localhost:8080/api/v1/auth/profile
```

**Roles**: `admin`, `agent`, `viewer`

---

## 🤝 Contributing

Before contributing:
- Read [guides/MAKEFILE.md](guides/MAKEFILE.md) for development commands
- Read [guides/domain_mapping/](guides/domain_mapping/) for domain model
- Run `make test-unit` before committing
- Run `make fmt` to format code

---

## 📄 License

MIT License - see [LICENSE](LICENSE)

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/ventros/crm/issues)
- **Docs**: [guides/](guides/)
- **Email**: dev@ventros.ai

---

**Made with ❤️ by Ventros Team**
