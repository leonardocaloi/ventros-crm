# Ventros CRM

[![Go Version](https://img.shields.io/badge/Go-1.25.1-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **CRM moderno** com Domain-Driven Design (DDD), Event-Driven Architecture e SAGA pattern usando Temporal.

Sistema de gerenciamento de relacionamento com clientes integrado com WhatsApp (via WAHA), arquitetura baseada em eventos (RabbitMQ) e workflows duráveis (Temporal).

---

## 🛠️ Tecnologias

- **Go 1.25.1** + Gin (API REST)
- **PostgreSQL 16** (banco principal com Row-Level Security)
- **RabbitMQ** (event bus com DLQ)
- **Redis** (cache/sessions)
- **Temporal** (workflows e SAGAs)
- **Containers** (Docker/Podman/Buildah - agnóstico OCI)
- **Kubernetes/Helm** (deploy produção)

**Arquitetura**: DDD + Event-Driven + SAGA + Multi-tenancy

---

## 🚀 Quick Start - Desenvolvimento Local

---

### 1. Clone e Configure
```bash
git clone https://github.com/caloi/ventros-crm.git
cd ventros-crm
cp .env.example .env  # Edite conforme necessário
```

### 2. Desenvolvimento Local (Modo Debug) ⭐ RECOMENDADO

**Opção A: Tudo junto (automático)**
```bash
make dev  # Sobe infra + API em sequência
```

**Opção B: Separado (para debug)**
```bash
# 1. Sobe APENAS infraestrutura (PostgreSQL, RabbitMQ, Redis, Temporal)
make infra

# 2. Em outro terminal: roda APENAS a API
make api
```

**Acesse:**
- API: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html
- RabbitMQ: http://localhost:15672 (guest/guest)
- Temporal: http://localhost:8088

### 3. Containers Full Stack (Docker/Podman)
```bash
# UM COMANDO: build + sobe tudo containerizado
make container

# Testa
curl http://localhost:8080/health

# Com Podman
CONTAINER_RUNTIME=podman make container
```

### 4. Kubernetes com Helm
```bash
# Inicie Minikube
minikube start

# UM COMANDO: deploy completo no K8s
make k8s

# Acesse API
kubectl port-forward -n ventros-crm svc/ventros-crm 8080:8080

# Ver status
make k8s-pods
make k8s-logs
```

---

## 📋 Comandos Make

```bash
# 🚀 Desenvolvimento
make infra            # [INFRA] Sobe SÓ infraestrutura
make api              # [API] Roda SÓ a API (requer infra)
make dev              # [DEV] Infra + API (automático)

# 🐳 Outros Ambientes
make container        # [CONTAINER] Full containerizado
make k8s              # [K8S] Deploy no Minikube

# 🛑 Parar/Limpar
make infra-stop       # Para infraestrutura
make infra-clean      # Para + apaga volumes (DESTRUTIVO)
make infra-reset      # Limpa + sobe de novo (fresh start)
make container-stop   # Para containers
make k8s-delete       # Remove do K8s

# 📊 Logs
make infra-logs       # Logs da infra
make container-logs   # Logs dos containers
make k8s-logs         # Logs do K8s

# 🛠️  Utilitários
make build            # Compila binário
make test             # Roda testes unitários
make test-e2e         # Roda testes E2E (requer API rodando)
make swagger          # Gera docs
make health           # Checa API
make help             # Ajuda completa
```

---

## 📚 Documentação Completa

- **[Guia de Instalação](guides/getting-started/)** - Setup detalhado
- **[Arquitetura](ARCHITECTURE.md)** - DDD, Event-Driven, SAGA
- **[Tarefas e Roadmap](TASKS.md)** - Próximas features
- **[Contribuir](CONTRIBUTING.md)** - Guidelines para devs

---

## 🏗️ Arquitetura (Resumo)

### Domain-Driven Design
```
internal/domain/    → Aggregates (Contact, Session, Message)
internal/application/ → Use Cases, DTOs, Services
infrastructure/     → Repositories, Event Bus, HTTP
```

### Event-Driven
- Domain Events publicados no RabbitMQ após commits
- 15+ filas WAHA (message, call, label, group events)
- Dead Letter Queue (DLQ) com 3 retries

### SAGA com Temporal
- `SessionLifecycleWorkflow` gerencia timeout de conversas
- Activities com compensação para rollback
- Cleanup automático de sessões órfãs

Ver [ARCHITECTURE.md](ARCHITECTURE.md) para detalhes.

---

## 🔐 Features

- ✅ **Contact/Session/Message** - Aggregates DDD completos
- ✅ **WhatsApp via WAHA** - Mensagens inbound/outbound
- ✅ **Multi-tenancy** - Row-Level Security no PostgreSQL
- ✅ **RBAC** - 4 roles (Admin, Manager, User, ReadOnly)
- ✅ **Webhooks** - Sistema de subscrição para eventos
- ✅ **Temporal Workflows** - SAGA para operações distribuídas

---

## 📄 License

MIT License - veja [LICENSE](LICENSE)

---

## 👥 Autor

**Leoanrdo Caloi** - [@caloi](https://github.com/leonardocaloi)

---

**Dúvidas?** Abra uma [issue](https://github.com/caloi/ventros-crm/issues) ou veja [CONTRIBUTING.md](CONTRIBUTING.md)
