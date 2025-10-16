# Claude Code Agents - Ventros CRM

**Total Agents**: 32 agentes (15 CRM + 4 Global + 7 Meta + 6 Management)
**Purpose**: Análise completa de codebase + gerenciamento de documentação
**Coverage**: 100% (todas as 30 tabelas de análise cobertas)
**Output Structure**: `code-analysis/` (organizado por categoria)
**Naming Pattern**: `{scope}_{category}_{name}.md`

---

## 📊 Nova Estrutura de Categorização

### Prefixos por Escopo

- **`crm_*`** - Específico do Ventros CRM (15 agentes)
- **`global_*`** - Aplicável a qualquer projeto Go (4 agentes)
- **`meta_*`** - Orquestração e desenvolvimento (7 agentes)
- **`mgmt_*`** - Gerenciamento de documentação e estado (6 agentes)

### Benefícios

1. **Clara distinção de escopo** - Fácil identificar agentes reutilizáveis
2. **Melhor organização** - Filtrar por prefixo (`ls crm_*.md`)
3. **Manutenção simplificada** - Atualizar apenas agentes relevantes
4. **Portabilidade** - Agentes `global_*` funcionam em qualquer projeto Go

---

## 🏗️ CRM-Specific Agents (15 agentes)

Análise específica do Ventros CRM (domínio, infraestrutura, AI/ML).

### Domain Analysis (5 agentes)

#### 1. **crm_domain_model_analyzer** 🔴 CRITICAL
- **Output**: `code-analysis/domain-analysis/domain_model_analysis.md`
- **Runtime**: 60-70 min
- **Tabelas**: 1 (Aggregates), 2 (Events), 5 (Children Entities)
- **O que faz**: Analisa 30 agregados DDD, eventos, repositórios, optimistic locking

#### 2. **crm_value_objects_analyzer** ⚪ STANDARD
- **Output**: `code-analysis/domain-analysis/value_objects_analysis.md`
- **Runtime**: 30-40 min
- **Tabela**: 6 (Value Objects)
- **O que faz**: Value objects, primitive obsession, immutability

#### 3. **crm_entity_relationships_analyzer** ⚪ STANDARD
- **Output**: `code-analysis/domain-analysis/entity_relationships_analysis.md`
- **Runtime**: 35-45 min
- **Tabela**: 4 (Entity Relationships)
- **O que faz**: Foreign keys, cardinality, relacionamentos entre agregados

#### 4. **crm_use_cases_analyzer** ⚪ STANDARD
- **Output**: `code-analysis/domain-analysis/use_cases_analysis.md`
- **Runtime**: 40-50 min
- **Tabela**: 10 (Use Cases)
- **O que faz**: CQRS commands/queries, 80+ use cases

#### 5. **crm_events_analyzer** ⚪ STANDARD
- **Output**: `code-analysis/domain-analysis/events_analysis.md`
- **Runtime**: 40-50 min
- **Tabela**: 11 (Domain Events)
- **O que faz**: 182+ domain events, Temporal workflows, Outbox Pattern

---

### Infrastructure (4 agentes)

#### 6. **crm_persistence_analyzer** 🟡 MEDIUM
- **Output**: `code-analysis/infrastructure/persistence_analysis.md`
- **Runtime**: 60-70 min
- **Tabelas**: 3 (Entities), 7 (Normalization), 9 (Migrations)
- **O que faz**: Database schema, GORM entities, migrations, RLS policies

#### 7. **crm_integration_analyzer** 🔴 CRITICAL
- **Output**: `code-analysis/infrastructure/integration_analysis.md`
- **Runtime**: 35-45 min
- **Tabelas**: 8 (Integrations), 12 (Event Bus)
- **O que faz**: WAHA, Stripe, Vertex AI, LlamaParse, circuit breaker

#### 8. **crm_workflows_analyzer** 🟡 MEDIUM
- **Output**: `code-analysis/infrastructure/workflows_analysis.md`
- **Runtime**: 40-50 min
- **O que faz**: Temporal workflows, sagas, long-running processes

#### 9. **crm_infrastructure_analyzer** 🟠 HIGH
- **Output**: `code-analysis/infrastructure/infrastructure_analysis.md`
- **Runtime**: 50-60 min
- **Tabelas**: 29 (Deployment), 30 (Roadmap)
- **O que faz**: Docker, Kubernetes, CI/CD (GitHub Actions + AWX + Helm)

---

### AI/ML (1 agente)

#### 10. **crm_ai_ml_analyzer** 🔴 CRITICAL
- **Output**: `code-analysis/ai-ml/ai_ml_analysis.md`
- **Runtime**: 50-60 min
- **Tabela**: 28 (AI/ML Features)
- **O que faz**: 12 AI providers, message enrichment (100%), memory service (20%)

---

### Quality (5 agentes)

#### 11. **crm_testing_analyzer** 🔴 CRITICAL
- **Output**: `code-analysis/quality/testing_analysis.md`
- **Runtime**: 40-50 min
- **Tabela**: 22 (Test Pyramid)
- **O que faz**: 82% coverage, pirâmide de testes, gaps

#### 12. **crm_security_analyzer** 🔴 CRITICAL 🔒
- **Output**: `code-analysis/quality/security_analysis.md`
- **Runtime**: 70-80 min
- **Tabelas**: 18 (OWASP), 21 (AI Security), 24-27 (Integration Security)
- **O que faz**: 5 P0 vulnerabilities, OWASP Top 10, RBAC

#### 13. **crm_resilience_analyzer** 🟠 HIGH
- **Output**: `code-analysis/quality/resilience_analysis.md`
- **Runtime**: 55-65 min
- **Tabelas**: 19 (Rate Limiting), 20 (Error Handling), 23 (Patterns)
- **O que faz**: Circuit breaker, retry logic, timeouts, rate limiting

#### 14. **crm_data_quality_analyzer** 🟡 MEDIUM
- **Output**: `code-analysis/quality/data_quality_analysis.md`
- **Runtime**: 60-70 min
- **Tabelas**: 13 (Query Perf), 14 (Consistency), 15 (Validations)
- **O que faz**: Query performance, N+1, validations, consistency

#### 15. **crm_api_analyzer** 🟠 HIGH
- **Output**: `code-analysis/infrastructure/api_analysis.md`
- **Runtime**: 45-55 min
- **Tabelas**: 16 (DTOs), 17 (REST Endpoints)
- **O que faz**: 158 endpoints, DTOs, Swagger, HTTP handlers

---

## 🌐 Global Agents (4 agentes)

Reutilizáveis em qualquer projeto Go.

#### 16. **global_deterministic_analyzer** 🔴 CRITICAL ⭐ BASELINE
- **Output**: `code-analysis/architecture/deterministic_metrics.md`
- **Runtime**: 5-10 min
- **Dependencies**: None ← **RODAR PRIMEIRO**
- **O que faz**: Baseline 100% factual (grep/wc/find), valida análises AI

#### 17. **global_code_style_analyzer** 🔵 USER-REQUESTED
- **Output**: `code-analysis/quality/code_style_analysis.md`
- **Runtime**: 40-50 min
- **O que faz**: Go idioms, naming conventions, code organization

#### 18. **global_documentation_analyzer** 🔵 USER-REQUESTED
- **Output**: `code-analysis/quality/documentation_analysis.md`
- **Runtime**: 45-55 min
- **O que faz**: Swagger, godoc, comentários, guias de API

#### 19. **global_solid_principles_analyzer** 🔵 USER-REQUESTED
- **Output**: `code-analysis/quality/solid_principles_analysis.md`
- **Runtime**: 55-65 min
- **O que faz**: S.O.L.I.D. principles audit completo

---

## 🎭 Meta Agents (7 agentes)

Orquestração, desenvolvimento de features, e pós-processamento.

### Analysis Orchestration (1 agente)

#### 20. **meta_orchestrator** 🔴 CRITICAL 🎯
- **Output**: `code-analysis/comprehensive/MASTER_ANALYSIS.md`
- **Runtime**: 2-3 horas (parallelizado)
- **Dependencies**: Todos os 18 agentes de análise
- **O que faz**: Coordena todos os agentes, consolida 30 tabelas

### Development Orchestration (3 agentes) 🆕

#### 21. **meta_dev_orchestrator** 🔴 CRITICAL 🚀 🆕
- **Output**: Feature completa (código + testes + PR)
- **Runtime**: 5 min (verify) to 2 hours (full feature)
- **Triggers**: `/add-feature <description>`
- **O que faz**: Orquestra desenvolvimento completo de features com DDD + Clean Architecture
  - Cria GitHub issue
  - Valida arquitetura (chama meta_feature_architect)
  - Implementa domain + application + infrastructure
  - Escreve testes (82%+ coverage)
  - Code review (chama meta_code_reviewer)
  - Commit + Push + PR
- **Intelligence**: Máxima - 3 modos (full/enhancement/verification)
- **Token Usage**: 5k-100k (otimizado conforme complexidade)

#### 22. **meta_feature_architect** 🔴 CRITICAL 🏗️ 🆕
- **Output**: `/tmp/architecture_plan.md`
- **Runtime**: 5-10 min
- **Called by**: meta_dev_orchestrator
- **O que faz**: Valida arquitetura e cria plano de implementação
  - Valida DDD + Clean Architecture + CQRS + SOLID
  - Gera checklist completo (53 items para full feature)
  - Estima esforço (tempo + tokens + arquivos)
  - Identifica riscos e dependências
  - Define bounded context correto
  - Security analysis (RBAC, BOLA, validação)

#### 23. **meta_code_reviewer** 🟠 HIGH 👁️ 🆕
- **Output**: `/tmp/code_review.md` com score
- **Runtime**: 5-10 min
- **Called by**: meta_dev_orchestrator
- **O que faz**: Review automático de código
  - Domain layer: 25 pontos (business logic, events, version field)
  - Application layer: 20 pontos (command pattern, validation)
  - Infrastructure: 15 pontos (repository, HTTP handler, migration)
  - SOLID: 15 pontos (S.O.L.I.D. principles)
  - Security: 15 pontos (RBAC, BOLA, SQL injection)
  - Testing: 10 pontos (coverage, mocks, e2e)
  - **Pass/Fail**: ≥80% = PASS, 70-79% = WARNING, <70% = FAIL

### Post-Processing (3 agentes)

#### 24. **meta_adr_generator** 🟡 MEDIUM
- **Output**: `docs/adr/*.md`
- **Runtime**: 30-40 min
- **O que faz**: Gera Architecture Decision Records (ADRs)

#### 25. **meta_docs_cleaner** ⚪ LOW
- **Output**: Estrutura limpa
- **Runtime**: 20-30 min
- **O que faz**: Move para archive, limpa temporários

#### 26. **meta_docs_consolidator** 🟠 HIGH
- **Output**: AI_REPORT.md consolidado, TODO.md consolidado
- **Runtime**: 30-40 min
- **O que faz**: Merge de fragmentos (AI_REPORT_PART1-6, etc.)

---

## 🛠️ Management Agents (6 agentes)

Mantêm documentação e estado atualizados.

#### 27. **mgmt_todo_manager** 🟠 HIGH ⭐
- **Output**: `planning/TODO.md`
- **Runtime**: 10-15 min
- **Triggers**: `/update-todo`, pós-análise
- **O que faz**: Consolida TODOs, atualiza status baseado no código, re-prioriza

#### 28. **mgmt_docs_index_manager** 🟡 MEDIUM 📚
- **Output**: README.md files (vários diretórios)
- **Runtime**: 5-10 min
- **Triggers**: `/update-indexes`, pós-consolidação
- **O que faz**: Atualiza índices, detecta novos arquivos, mantém links

#### 29. **mgmt_docs_reorganizer** 🟠 HIGH
- **Output**: Estrutura organizada
- **Runtime**: 15-20 min
- **O que faz**: Segue ORGANIZATION_RULES.md, move arquivos, atualiza referências

#### 30. **mgmt_makefile_updater** 🟡 MEDIUM 🆕
- **Output**: `MAKEFILE.md`
- **Runtime**: 5-10 min
- **Triggers**: `/update-makefile`, após Makefile changes
- **O que faz**: Sincroniza MAKEFILE.md com Makefile, extrai comandos, adiciona exemplos

#### 31. **mgmt_readme_updater** 🟡 MEDIUM 🆕
- **Output**: `README.md`
- **Runtime**: 5-10 min
- **Triggers**: `/update-readme`, após features completas
- **O que faz**: Atualiza badges, métricas, feature status, tech stack

#### 32. **mgmt_dev_guide_updater** 🟡 MEDIUM 🆕
- **Output**: `DEV_GUIDE.md`
- **Runtime**: 10-15 min
- **Triggers**: `/update-dev-guide`, após architecture changes
- **O que faz**: Atualiza exemplos de código (do código real), padrões, workflows

---

## 📋 Ordem de Execução Recomendada

### ⚡ Análise Completa (30 Tabelas)

```bash
# Fase 0: Baseline (5-10 min)
claude-code --agent global_deterministic_analyzer

# Fase 1: Core Analysis (70-80 min, parallel)
claude-code --agent crm_domain_model_analyzer &
claude-code --agent crm_testing_analyzer &
claude-code --agent crm_ai_ml_analyzer &
claude-code --agent crm_security_analyzer &
claude-code --agent crm_integration_analyzer &
claude-code --agent crm_infrastructure_analyzer &
claude-code --agent crm_resilience_analyzer &
claude-code --agent crm_api_analyzer &
wait

# Fase 2: Specialized Analysis (50-70 min, parallel)
claude-code --agent crm_persistence_analyzer &
claude-code --agent crm_data_quality_analyzer &
claude-code --agent global_code_style_analyzer &
claude-code --agent global_documentation_analyzer &
claude-code --agent global_solid_principles_analyzer &
claude-code --agent crm_value_objects_analyzer &
claude-code --agent crm_entity_relationships_analyzer &
claude-code --agent crm_use_cases_analyzer &
claude-code --agent crm_events_analyzer &
wait

# Fase 3: Orchestration (10-15 min)
claude-code --agent meta_orchestrator

# Fase 4: Management (15-20 min)
claude-code --agent mgmt_todo_manager
claude-code --agent mgmt_docs_index_manager

# Fase 5: Post-Processing (50-70 min, optional)
claude-code --agent meta_adr_generator
claude-code --agent meta_docs_consolidator
claude-code --agent meta_docs_cleaner
```

**Total**: 2.5-3 horas (parallelizado) / 15-20 horas (sequencial)

---

## 🛠️ Comandos Úteis

### Listar Agentes por Categoria

```bash
# Todos os agentes
ls .claude/agents/*.md | grep -v README
# Output: 32 agentes

# CRM-specific
ls .claude/agents/crm_*.md
# Output: 15 agentes

# Global (reutilizáveis)
ls .claude/agents/global_*.md
# Output: 4 agentes

# Meta (orchestration + development)
ls .claude/agents/meta_*.md
# Output: 7 agentes

# Management
ls .claude/agents/mgmt_*.md
# Output: 6 agentes
```

### Executar Agentes

```bash
# Análise completa via orchestrator
claude-code --agent meta_orchestrator

# Análise específica (novos nomes!)
claude-code --agent crm_security_analyzer
claude-code --agent global_deterministic_analyzer

# Atualizar documentação
claude-code --agent mgmt_todo_manager
claude-code --agent mgmt_makefile_updater
claude-code --agent mgmt_readme_updater
claude-code --agent mgmt_dev_guide_updater
```

### Slash Commands

```bash
# 🚀 Desenvolvimento de Features (NOVO!)
/add-feature Add a Custom Field aggregate to allow users to create custom fields on contacts
/add-feature Add rate limiting to campaign endpoints
/add-feature Verify the Contact aggregate follows DDD best practices

# Análise completa
/analyze-all

# Atualizar documentação
/update-todo
/update-makefile
/update-readme
/update-dev-guide
/update-indexes

# Consolidar
/consolidate-docs

# Security
/security:p0-check
```

---

## 📊 Cobertura das 30 Tabelas

| Agente (novo nome) | Tabelas | Categoria |
|--------------------|---------|-----------|
| global_deterministic_analyzer | Baseline | Métricas factuais |
| crm_domain_model_analyzer | 1, 2, 5 | Aggregates, Events, Children |
| crm_entity_relationships_analyzer | 4 | Foreign Keys, Cardinality |
| crm_value_objects_analyzer | 6 | Value Objects |
| crm_persistence_analyzer | 3, 7, 9 | Entities, Normalization, Migrations |
| crm_integration_analyzer | 8, 12 | Integrations, Event Bus |
| crm_use_cases_analyzer | 10 | CQRS Commands/Queries |
| crm_events_analyzer | 11 | Domain Events |
| crm_data_quality_analyzer | 13, 14, 15 | Query Perf, Consistency, Validations |
| crm_api_analyzer | 16, 17 | DTOs, REST Endpoints |
| crm_security_analyzer | 18, 21, 24-27 | OWASP, AI Security, Integration Tests |
| crm_resilience_analyzer | 19, 20, 23 | Rate Limiting, Error Handling, Patterns |
| crm_testing_analyzer | 22, 24, 25 | Test Pyramid, Integration Tests, Mocks |
| crm_ai_ml_analyzer | 28 | AI/ML Features, Providers |
| crm_infrastructure_analyzer | 29, 30 | Deployment, CI/CD, Roadmap |

**Análises Extras** (não nas 30 tabelas):
- global_code_style_analyzer
- global_documentation_analyzer
- global_solid_principles_analyzer

**Resultado**: ✅ 100% das 30 tabelas cobertas + 3 análises extras de qualidade

---

## 🔄 Diagrama de Interação

```
┌─────────────────────────────────┐
│ global_deterministic_analyzer   │ ⭐ BASELINE (SEMPRE PRIMEIRO)
└────────────┬────────────────────┘
             │
             ▼
┌────────────────────────────────────────┐
│  18 Agentes de Análise (parallel)     │
│  ├─ 15 CRM-specific (crm_*)           │
│  └─ 3 Global quality (global_*)       │
└────────────┬───────────────────────────┘
             │
             ▼
┌────────────────────────┐
│  meta_orchestrator     │ 🎯 Consolida 30 tabelas
└────────────┬───────────┘
             │
     ┌───────┴───────┐
     │               │
     ▼               ▼
┌──────────┐   ┌──────────────┐
│ mgmt_    │   │ mgmt_docs_   │
│ todo_    │   │ index_       │
│ manager  │   │ manager      │
└──────────┘   └──────────────┘
     │               │
     └───────┬───────┘
             │
             ▼
┌─────────────────────────┐
│  Pós-Processamento      │
│  - meta_adr_generator   │
│  - meta_docs_consolidator│
│  - meta_docs_cleaner    │
└─────────────────────────┘
```

---

## 🎯 Casos de Uso

### 1. Análise Completa (Primeira Vez)
```bash
claude-code --agent meta_orchestrator
# Tempo: ~3 horas
# Output: 30 tabelas + TODO.md atualizado
```

### 2. Re-análise Após Sprint
```bash
claude-code --agent crm_security_analyzer
claude-code --agent crm_testing_analyzer
claude-code --agent mgmt_todo_manager
# Tempo: ~1 hora
```

### 3. Atualizar Documentação
```bash
claude-code --agent mgmt_makefile_updater
claude-code --agent mgmt_readme_updater
claude-code --agent mgmt_dev_guide_updater
# Tempo: ~30 min
```

### 4. Security Audit P0
```bash
claude-code --agent crm_security_analyzer
claude-code --agent mgmt_todo_manager
# Tempo: ~90 min
```

### 5. Adicionar Nova Feature (NOVO!) 🚀
```bash
/add-feature Add a Broadcast feature for sending messages to multiple contacts
# Tempo: 60-90 min (full feature) ou 15-30 min (enhancement)
# Output: Feature completa com:
#   - Domain layer (aggregate, events, repository)
#   - Application layer (commands, handlers, DTOs)
#   - Infrastructure layer (entity, repo impl, HTTP handler, migration)
#   - Tests (unit + integration + e2e)
#   - PR criada e pronta para review
```

---

## 📝 Changelog

### v6.0 (2025-10-15) - CURRENT ✨ 🚀
- ✅ **Added 3 development orchestrators** (feature development workflow)
  - `meta_dev_orchestrator` - Full feature development (analysis → code → tests → PR)
  - `meta_feature_architect` - Architecture validation & planning
  - `meta_code_reviewer` - Automated code review (100-point checklist)
- ✅ **New slash command**: `/add-feature` - AI implements complete features
- ✅ **Intelligence modes**: Full (100k tokens), Enhancement (30k), Verification (10k)
- ✅ **Total: 32 agentes**
  - 15 CRM-specific (`crm_*`)
  - 4 Global reusable (`global_*`)
  - **7 Meta orchestration** (`meta_*`) ← Aumentou de 4 para 7
  - 6 Management (`mgmt_*`)

### v5.0 (2025-10-15)
- ✅ **Renamed all 26 agents** with scope prefixes
- ✅ Added 3 new updater agents (makefile, readme, dev_guide)
- ✅ **Total: 29 agentes**
- ✅ New naming pattern: `{scope}_{category}_{name}.md`
- ✅ Updated YAML `name:` fields in all agents

### v4.0 (2025-10-15)
- 24 agentes (18 análise + 4 meta + 2 gerenciamento)

### v3.0 (2025-10-14)
- 22 agentes (18 análise + 4 meta)

---

**Version**: 6.0 🚀
**Last Updated**: 2025-10-15
**Total Agents**: 32 (15 CRM + 4 Global + 7 Meta + 6 Management)
**Coverage**: 100% das 30 tabelas de análise + desenvolvimento de features
**Output Structure**: `code-analysis/` (organizado por categoria)
**Naming Pattern**: `{scope}_{category}_{name}.md`
**Estimated Runtime**:
- Analysis: 2.5-3 horas (parallelized) / 15-20 horas (sequential)
- Development: 5 min (verify) to 2 hours (full feature)

---

**Maintainer**: Ventros CRM Team
**Status**: ✅ Sistema completo com categorização por escopo
