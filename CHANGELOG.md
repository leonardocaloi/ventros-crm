# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Complete CQRS implementation with Commands and Queries
- Unit of Work pattern for atomic transactions
- SAGA pattern with compensation activities
- Assemblers for Domain → DTO conversion
- Event Subscribers for asynchronous processing
- Comprehensive test suite (unit, integration, E2E)

### Changed
- Refactor Handlers to use Use Cases instead of direct Repository access
- Centralize DTOs in `/internal/application/dtos/`
- Improve error handling and logging

### Fixed
- Domain events not being published from HTTP handlers
- Transaction isolation issues in concurrent operations

---

## [0.1.0] - 2025-10-06

### Added
- **Domain-Driven Design** implementation with Aggregates (Contact, Session, Message)
- **Value Objects** for Email and Phone validation
- **Domain Events** pattern for choreography
- **Event-Driven Architecture** with RabbitMQ
  - Dead Letter Queue (DLQ) with 3 retry attempts
  - 15+ event queues for WAHA integration
- **Temporal Workflows** for session lifecycle management
  - SessionLifecycleWorkflow with timeout handling
  - SessionCleanupWorkflow for orphaned sessions
  - Activities with compensation (EndSession, CleanupSessions)
- **WhatsApp Integration** via WAHA HTTP API
  - Inbound message processing
  - Message status tracking (sent, delivered, read)
  - Media support (images, documents, audio)
- **Multi-tenancy** with Row-Level Security (RLS)
  - PostgreSQL RLS policies
  - Tenant isolation at database level
- **RBAC** (Role-Based Access Control)
  - 4 roles: Admin, Manager, User, ReadOnly
  - Resource-based permissions
  - RBAC middleware for routes
- **Authentication & Authorization**
  - JWT-based authentication
  - Auth middleware with context propagation
  - RLS middleware for database session
- **Webhook System**
  - Subscription management
  - Event filtering
  - Retry mechanism with exponential backoff
- **REST API** with Swagger documentation
  - Contact CRUD endpoints
  - Session management endpoints
  - Message endpoints
  - Webhook subscription endpoints
  - Health check endpoint
- **Infrastructure**
  - PostgreSQL 16 with GORM
  - RabbitMQ for message broker
  - Redis for caching
  - Docker & Docker Compose setup
- **Helm Charts** for Kubernetes deployment
  - PostgreSQL init job with best practices
  - Schema validation (values.schema.json)
  - Dynamic NOTES.txt
  - Phase-based deployment support
- **Documentation**
  - Swagger/OpenAPI specs
  - Architecture diagrams
  - Deployment guides

### Changed
- Migrated from lib/pq to GORM for better ORM support
- Standardized error handling across layers
- Improved logging with zap structured logging

### Fixed
- Session timeout not being extended on new messages
- Race conditions in concurrent message processing
- Memory leaks in event subscribers

### Security
- Implemented RLS for tenant isolation
- Added RBAC for fine-grained access control
- Secured sensitive configuration in environment variables
- Added input validation and sanitization

---

## [0.0.1] - 2025-09-15 [YANKED]

### Added
- Initial project structure
- Basic Contact entity
- PostgreSQL connection
- Simple REST API with Gin

### Removed
- This version was yanked due to architectural issues
- Replaced with 0.1.0 with proper DDD implementation

---

## Release Notes

### Version 0.1.0 - "Foundation"
This is the first official release of Ventros CRM. It establishes the architectural foundation with Domain-Driven Design, Event-Driven Architecture, and SAGA pattern using Temporal. The system is production-ready for WhatsApp integrations via WAHA, with full multi-tenancy and RBAC support.

**Key Highlights:**
- ✅ Enterprise-grade architecture (DDD + Event-Driven + SAGA)
- ✅ WhatsApp integration ready
- ✅ Multi-tenant with RLS
- ✅ RBAC for access control
- ✅ Kubernetes-ready with Helm charts
- ⚠️ Handlers still access repositories directly (to be fixed in 0.2.0)
- ⚠️ SAGA without compensation (to be implemented in 0.2.0)

---

## Migration Guides

### Upgrading to 0.1.0
As this is the first release, no migration is required. Follow the [Getting Started](guides/getting-started/) guide for installation.

---

## Links
- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
- [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
- [GitHub Releases](https://github.com/caloi/ventros-crm/releases)

---

[Unreleased]: https://github.com/caloi/ventros-crm/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/caloi/ventros-crm/releases/tag/v0.1.0
[0.0.1]: https://github.com/caloi/ventros-crm/releases/tag/v0.0.1
