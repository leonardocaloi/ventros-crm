# Code Analysis - Ventros CRM

**Last Update**: 2025-10-15
**Total Agents**: 25 (18 analysis + 4 meta + 3 management)
**Coverage**: 100% (all 30 analysis tables)

---

## 📁 Directory Structure

### 1. **architecture/** - Baseline & Architecture Metrics
**Agents**: 1
- `deterministic_analyzer` → `deterministic_metrics.md`

**Contains**:
- Factual baseline (grep/find/wc counts)
- Architecture scores by layer
- AI_REPORT.md (consolidated from 7 parts)

---

### 2. **domain/** - Domain-Driven Design Analysis
**Agents**: 6
- `domain_model_analyzer` → `domain_model_analysis.md`
- `entity_relationships_analyzer` → `entity_relationships_analysis.md`
- `value_objects_analyzer` → `value_objects_analysis.md`
- `use_cases_analyzer` → `use_cases_analysis.md`
- `events_analyzer` → `events_analysis.md`
- `workflows_analyzer` → `workflows_analysis.md` (Temporal)

**Contains**:
- Aggregates (30 total)
- Domain events (182+)
- Value objects (12+)
- CQRS commands/queries (80+/20+)
- Entity relationships
- Temporal workflows

---

### 3. **infrastructure/** - Infrastructure & Integrations
**Agents**: 4
- `persistence_analyzer` → `persistence_analysis.md`
- `api_analyzer` → `api_analysis.md`
- `integration_analyzer` → `integration_analysis.md`
- `infrastructure_analyzer` → `infrastructure_analysis.md`

**Contains**:
- Database schema (49 migrations, 350+ indexes)
- REST API endpoints (158 total)
- External integrations (WAHA, Stripe, Vertex AI, RabbitMQ)
- Docker/K8s, CI/CD, monitoring

---

### 4. **quality/** - Code Quality & Security
**Agents**: 7
- `testing_analyzer` → `testing_analysis.md`
- `security_analyzer` → `security_analysis.md`
- `resilience_analyzer` → `resilience_analysis.md`
- `code_style_analyzer` → `code_style_analysis.md`
- `documentation_analyzer` → `documentation_analysis.md`
- `solid_principles_analyzer` → `solid_principles_analysis.md`
- `data_quality_analyzer` → `data_quality_analysis.md`

**Contains**:
- Test coverage (unit/integration/e2e)
- Security audit (OWASP Top 10, P0 vulnerabilities)
- Resilience patterns (circuit breaker, retry, timeout)
- Go idioms & naming conventions
- Swagger/godoc quality
- S.O.L.I.D. principles compliance
- Query performance & data consistency

---

### 5. **ai-ml/** - AI/ML Features
**Agents**: 1
- `ai_ml_analyzer` → `ai_ml_analysis.md`

**Contains**:
- AI providers (12 total: Groq, Vertex, OpenAI, LlamaParse)
- Message enrichment (100% complete)
- Future features gaps (Memory Service, Python ADK)
- Cost tracking

---

### 6. **comprehensive/** - Master Reports
**Agents**: 1 (orchestrator)
- `orchestrator` → `MASTER_ANALYSIS.md`

**Contains**:
- Consolidated report with all 30 tables
- Cross-agent findings
- Overall architecture score
- Executive summary

---

### 7. **adr/** - Architecture Decision Records
**Agents**: 1
- `adr_generator` → `NNNN-title.md`

**Contains**:
- ADR-0001: Adopt DDD
- ADR-0002: Hexagonal Architecture
- ADR-0003: Event-Driven + Outbox Pattern
- ... (17+ ADRs total)

---

### 8. **archive/** - Historical Analysis
**Agents**: 1 (docs_cleanup)

**Contains**:
- Old analysis reports (dated folders)
- Superseded documentation
- Historical AI reports

---

## 🔄 Agent Output Mapping

| Agent | Output Path |
|-------|-------------|
| deterministic_analyzer | architecture/deterministic_metrics.md |
| domain_model_analyzer | domain/domain_model_analysis.md |
| entity_relationships_analyzer | domain/entity_relationships_analysis.md |
| value_objects_analyzer | domain/value_objects_analysis.md |
| use_cases_analyzer | domain/use_cases_analysis.md |
| events_analyzer | domain/events_analysis.md |
| workflows_analyzer | domain/workflows_analysis.md |
| persistence_analyzer | infrastructure/persistence_analysis.md |
| api_analyzer | infrastructure/api_analysis.md |
| integration_analyzer | infrastructure/integration_analysis.md |
| infrastructure_analyzer | infrastructure/infrastructure_analysis.md |
| testing_analyzer | quality/testing_analysis.md |
| security_analyzer | quality/security_analysis.md |
| resilience_analyzer | quality/resilience_analysis.md |
| code_style_analyzer | quality/code_style_analysis.md |
| documentation_analyzer | quality/documentation_analysis.md |
| solid_principles_analyzer | quality/solid_principles_analysis.md |
| data_quality_analyzer | quality/data_quality_analysis.md |
| ai_ml_analyzer | ai-ml/ai_ml_analysis.md |
| orchestrator | comprehensive/MASTER_ANALYSIS.md |
| adr_generator | adr/NNNN-*.md |

---

## 📊 Analysis Coverage (30 Tables)

### Domain Model (Tables 1-6, 10-11)
- ✅ Table 1: Aggregates (domain_model_analyzer)
- ✅ Table 2: Domain Events (domain_model_analyzer, events_analyzer)
- ✅ Table 4: Entity Relationships (entity_relationships_analyzer)
- ✅ Table 5: Aggregate Children (domain_model_analyzer)
- ✅ Table 6: Value Objects (value_objects_analyzer)
- ✅ Table 10: Use Cases (use_cases_analyzer)
- ✅ Table 11: Domain Events Detail (events_analyzer)

### Infrastructure (Tables 3, 7-9, 12, 16-17, 26-30)
- ✅ Table 3: Entities (persistence_analyzer)
- ✅ Table 7: Normalization (persistence_analyzer)
- ✅ Table 8: External Integrations (integration_analyzer)
- ✅ Table 9: Migrations (persistence_analyzer)
- ✅ Table 12: Event Bus (integration_analyzer)
- ✅ Table 16: DTOs (api_analyzer)
- ✅ Table 17: REST Endpoints (api_analyzer)
- ✅ Table 26: Integrations Detail (integration_analyzer)
- ✅ Table 27: gRPC vs REST (integration_analyzer)
- ✅ Table 28: Cache Strategy (integration_analyzer)
- ✅ Table 29: Deployment (infrastructure_analyzer)
- ✅ Table 30: Roadmap (infrastructure_analyzer)

### Quality (Tables 13-15, 18-25)
- ✅ Table 13: Query Performance (data_quality_analyzer)
- ✅ Table 14: Data Consistency (data_quality_analyzer)
- ✅ Table 15: Validations (data_quality_analyzer)
- ✅ Table 18: OWASP (security_analyzer)
- ✅ Table 19: Rate Limiting (resilience_analyzer)
- ✅ Table 20: Error Handling (resilience_analyzer)
- ✅ Table 21: AI Security (security_analyzer)
- ✅ Table 22: Test Pyramid (testing_analyzer)
- ✅ Table 23: Resilience Patterns (resilience_analyzer)
- ✅ Table 24: Integration Tests (testing_analyzer, security_analyzer)
- ✅ Table 25: Mock Quality (testing_analyzer)

---

## 🚀 How to Use

### Run Full Analysis
```bash
# Complete analysis (all 30 tables)
claude-code --agent orchestrator
# Output: code-analysis/comprehensive/MASTER_ANALYSIS.md
```

### Run Specific Analysis
```bash
# Security audit
claude-code --agent security_analyzer
# Output: code-analysis/quality/security_analysis.md

# Domain model analysis
claude-code --agent domain_model_analyzer
# Output: code-analysis/domain/domain_model_analysis.md
```

### Update Indexes
```bash
# Update all README.md files
claude-code --agent docs_index_manager
```

---

## 📚 Related Documentation

- [Agent Catalog](../.claude/agents/README.md) - All 25 agents
- [TODO.md](../TODO.md) - Roadmap (30 sprints)
- [Planning](../planning/) - Future features (ventros-ai, memory-service)

---

**Auto-generated**: 2025-10-15
**Maintainer**: docs_index_manager agent
