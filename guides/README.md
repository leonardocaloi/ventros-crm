# 📚 Ventros CRM - Documentation Guides

Welcome to the Ventros CRM documentation! This directory contains comprehensive guides to help you understand, develop, and deploy the system.

---

## 📖 Available Guides

### 🏗️ [Architecture](architecture/)
Learn about the system's architecture, design patterns, and technical decisions.

- **[Architecture Overview](architecture/README.md)** - High-level architecture with DDD, Event-Driven, and SAGA patterns
- **[Domain-Driven Design](architecture/ddd.md)** - Aggregates, Entities, Value Objects
- **[Event-Driven Architecture](architecture/event-driven.md)** - Event flow, RabbitMQ, choreography
- **[SAGA Pattern](architecture/saga.md)** - Temporal workflows and distributed transactions
- **[Diagrams](architecture/diagrams/)** - Visual representations of the system
- **[ADRs](architecture/decisions/)** - Architecture Decision Records

### 🚀 [Getting Started](getting-started/)
Quick start guides to get you up and running.

- **[Local Development Setup](getting-started/README.md)** - Complete setup guide
- **[Quickstart](getting-started/quickstart.md)** - Get running in 5 minutes
- **[Troubleshooting](getting-started/troubleshooting.md)** - Common issues and solutions
- **[Environment Variables](getting-started/environment.md)** - Configuration reference

### 🚢 [Deployment](deployment/)
Production deployment guides.

- **[Kubernetes with Helm](deployment/README.md)** - Deploy to Kubernetes
- **[Docker Compose](deployment/docker.md)** - Deploy with Docker
- **[Production Checklist](deployment/production-checklist.md)** - Pre-launch verification
- **[Scaling Guide](deployment/scaling.md)** - Horizontal and vertical scaling

### 💻 [Code Examples](code-examples/)
Practical code examples and patterns.

- **[RBAC Example](code-examples/rbac_example.go)** - Role-Based Access Control
- **[Use Case Example](code-examples/use_case_example.md)** - Implementing use cases
- **[Event Handler Example](code-examples/event_handler_example.md)** - Domain event handling
- **[Testing Patterns](code-examples/testing_patterns.md)** - Unit, integration, E2E tests

### 🔌 [API Guide](api/)
REST API documentation and examples.

- **[API Overview](api/README.md)** - RESTful API design
- **[Authentication](api/authentication.md)** - JWT auth flow
- **[Webhooks](api/webhooks.md)** - Webhook subscriptions
- **[Endpoints Reference](api/endpoints.md)** - Complete endpoint list

---

## 📋 Quick Links

| Resource | Description | Link |
|----------|-------------|------|
| **Main README** | Project overview | [../README.md](../README.md) |
| **CHANGELOG** | Version history | [../CHANGELOG.md](../CHANGELOG.md) |
| **CONTRIBUTING** | Contribution guide | [../CONTRIBUTING.md](../CONTRIBUTING.md) |
| **TASKS** | Task roadmap | [../TASKS.md](../TASKS.md) |
| **Swagger UI** | Interactive API docs | http://localhost:8080/swagger/index.html |

---

## 🎯 Documentation Philosophy

Our documentation follows these principles:

1. **Clarity First** - Clear, concise, and easy to understand
2. **Code Examples** - Real, working code examples
3. **Visual Aids** - Diagrams and flowcharts where helpful
4. **Up-to-Date** - Kept in sync with code changes
5. **Searchable** - Well-organized and easy to navigate

---

## 🤝 Contributing to Documentation

Found a typo? Want to add an example? Contributions welcome!

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on:
- Improving existing documentation
- Adding new guides
- Creating diagrams
- Translating content

---

## 📞 Get Help

If you can't find what you're looking for:

1. Check the [Troubleshooting Guide](getting-started/troubleshooting.md)
2. Search [GitHub Issues](https://github.com/caloi/ventros-crm/issues)
3. Ask in [GitHub Discussions](https://github.com/caloi/ventros-crm/discussions)
4. Email support: support@ventros.com

---

**Happy coding! 🚀**
