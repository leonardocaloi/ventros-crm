# Claude Code Agents - Ventros CRM

**Total Agents**: 24 agentes (18 análise + 4 meta + 2 gerenciamento)
**Purpose**: Análise completa de codebase + gerenciamento de documentação
**Coverage**: 100% (todas as 30 tabelas de análise cobertas)
**Output Structure**: `code-analysis/` (organizado por categoria)

---

## 📊 Visão Geral

### Sistema Multi-Agente

Este projeto usa **24 agentes especializados** que trabalham em conjunto:

1. **18 Agentes de Análise**: Analisam diferentes aspectos do código (domínio, infraestrutura, segurança, testes, etc.)
2. **4 Agentes Meta**: Orquestração e pós-processamento (orchestrator, adr_generator, docs_cleanup, docs_consolidator)
3. **2 Agentes de Gerenciamento**: Mantêm documentação atualizada (todo_manager, docs_index_manager)

### Estrutura de Output

Todos os agentes geram outputs em `code-analysis/` organizado por categoria:

```
code-analysis/
├── architecture/          # Métricas arquiteturais e AI_REPORT consolidado
├── domain-analysis/       # Análises DDD (aggregates, eventos, value objects)
├── infrastructure/        # Persistence, API, integrações, deploy
├── quality/              # Testes, segurança, code style, SOLID
├── ai-ml/                # Features AI/ML, providers, custos
├── comprehensive/        # Reports completos (orchestrator output)
└── archive/              # Análises antigas com timestamp
```

---

## 🔴 CRITICAL Priority (6 agentes)

Devem rodar primeiro, fornecem baseline essencial:

### 1. **deterministic_analyzer** ⭐ BASELINE
**Propósito**: Gera baseline 100% factual (contagens determinísticas)
- **Runtime**: 5-10 min
- **Output**: `code-analysis/architecture/deterministic_metrics.md`
- **Dependencies**: None ← **RODAR PRIMEIRO**
- **Tools**: Bash (grep, find, wc, cloc)
- **Tables**: Baseline factual para validar análises AI

**O que faz**:
- Conta arquivos, linhas, funções (grep/find/wc)
- Valida análises AI (determinístico vs AI score)
- Gera métricas atemporais (sem números hardcoded)

---

### 2. **domain_model_analyzer**
**Propósito**: Analisa agregados DDD, eventos, repositórios
- **Runtime**: 60-70 min
- **Output**: `code-analysis/domain-analysis/domain_model_analysis.md`
- **Tables**: 1 (Aggregates), 2 (Events), 5 (Children Entities)
- **Tools**: Read, Grep, Glob, Bash

**O que faz**:
- Identifica todos os agregados (internal/domain/**/aggregate.go)
- Valida padrões DDD (version field, eventos, repositórios)
- Score de qualidade 1-10 com evidências

---

### 3. **testing_analyzer**
**Propósito**: Analisa pirâmide de testes, cobertura, qualidade
- **Runtime**: 40-50 min
- **Output**: `code-analysis/quality/testing_analysis.md`
- **Tables**: 22 (Test Pyramid)
- **Tools**: Read, Grep, Glob, Bash

**O que faz**:
- Executa `make test-coverage`
- Analisa unit/integration/e2e (proporções)
- Identifica use cases sem testes
- Score de cobertura por camada

---

### 4. **ai_ml_analyzer**
**Propósito**: Analisa features AI/ML, providers, custos
- **Runtime**: 50-60 min
- **Output**: `code-analysis/ai-ml/ai_ml_analysis.md`
- **Tables**: 28 (AI/ML Features)
- **Tools**: Read, Grep, Glob, Bash

**O que faz**:
- Mapeia 12 AI providers (Groq, Vertex, OpenAI, LlamaParse)
- Analisa message enrichment (100% implementado)
- Identifica gaps (Memory Service 80% missing)
- Cost tracking readiness

---

### 5. **security_analyzer** 🔒
**Propósito**: Auditoria de segurança OWASP Top 10
- **Runtime**: 70-80 min
- **Output**: `code-analysis/quality/security_analysis.md`
- **Tables**: 18 (OWASP), 21 (AI Security), 24-27 (Integration Security)
- **Tools**: Read, Grep, Glob, Bash

**O que faz**:
- Identifica P0 vulnerabilidades (SSRF, BOLA, Auth Bypass)
- OWASP Top 10 coverage
- Security headers, RBAC, rate limiting
- **IMPORTANTE**: Atualiza TODO.md via todo_manager (P0 encontrados)

---

### 6. **integration_analyzer**
**Propósito**: Analisa integrações com APIs externas
- **Runtime**: 35-45 min
- **Output**: `code-analysis/infrastructure/integration_analysis.md`
- **Tables**: 8 (Integrations), 12 (Event Bus)
- **Tools**: Read, Grep, Glob, Bash

**O que faz**:
- Mapeia WAHA, Stripe, Vertex AI, LlamaParse
- Circuit breaker, retry logic, timeouts
- Event bus (RabbitMQ + Outbox Pattern)

---

## 🟠 HIGH Priority (3 agentes)

Importantes para production readiness:

### 7. **infrastructure_analyzer**
**Propósito**: Deploy, CI/CD, infraestrutura
- **Runtime**: 50-60 min
- **Output**: `code-analysis/infrastructure/infrastructure_analysis.md`
- **Tables**: 29 (Deployment), 30 (Roadmap)
- **Tools**: Read, Grep, Glob, Bash

---

### 8. **resilience_analyzer**
**Propósito**: Rate limiting, error handling, resilience patterns
- **Runtime**: 55-65 min
- **Output**: `code-analysis/quality/resilience_analysis.md`
- **Tables**: 19 (Rate Limiting), 20 (Error Handling), 23 (Patterns)
- **Tools**: Read, Grep, Glob, Bash

---

### 9. **api_analyzer**
**Propósito**: REST endpoints, DTOs, Swagger
- **Runtime**: 45-55 min
- **Output**: `code-analysis/infrastructure/api_analysis.md`
- **Tables**: 16 (DTOs), 17 (REST Endpoints)
- **Tools**: Read, Grep, Glob, Bash

---

## 🟡 MEDIUM Priority (2 agentes)

Database e data quality:

### 10. **persistence_analyzer**
**Propósito**: Database schema, migrations, repositories
- **Runtime**: 60-70 min
- **Output**: `code-analysis/infrastructure/persistence_analysis.md`
- **Tables**: 3 (Entities), 7 (Normalization), 9 (Migrations)
- **Tools**: Read, Grep, Glob, Bash

---

### 11. **data_quality_analyzer**
**Propósito**: Query performance, consistency, validations
- **Runtime**: 60-70 min
- **Output**: `code-analysis/quality/data_quality_analysis.md`
- **Tables**: 13 (Query Perf), 14 (Consistency), 15 (Validations)
- **Tools**: Read, Grep, Glob, Bash

---

## 🔵 USER-REQUESTED (3 agentes)

Code quality (solicitados pelo usuário):

### 12. **code_style_analyzer**
**Propósito**: Go idioms, naming conventions, organização
- **Runtime**: 40-50 min
- **Output**: `code-analysis/quality/code_style_analysis.md`
- **Tables**: 6 tabelas (Idioms, Naming, Organization, Errors, Interfaces, Consistency)
- **Tools**: Read, Grep, Glob, Bash

---

### 13. **documentation_analyzer**
**Propósito**: Swagger, godoc, guias de API
- **Runtime**: 45-55 min
- **Output**: `code-analysis/quality/documentation_analysis.md`
- **Tables**: 6 tabelas (Swagger, Errors, Comments, Guides, Examples, Consistency)
- **Tools**: Read, Grep, Glob, Bash

---

### 14. **solid_principles_analyzer**
**Propósito**: S.O.L.I.D. principles audit
- **Runtime**: 55-65 min
- **Output**: `code-analysis/quality/solid_principles_analysis.md`
- **Tables**: 6 tabelas (SRP, OCP, LSP, ISP, DIP, Overall Score)
- **Tools**: Read, Grep, Glob, Bash

---

## ⚪ STANDARD Priority (4 agentes)

Detalhes do modelo de domínio:

### 15. **value_objects_analyzer**
**Propósito**: Value objects, primitive obsession
- **Runtime**: 30-40 min
- **Output**: `code-analysis/domain-analysis/value_objects_analysis.md`
- **Table**: 6 (Value Objects)
- **Tools**: Read, Grep, Glob

---

### 16. **entity_relationships_analyzer**
**Propósito**: Foreign keys, cardinality, relacionamentos
- **Runtime**: 35-45 min
- **Output**: `code-analysis/domain-analysis/entity_relationships_analysis.md`
- **Table**: 4 (Entity Relationships)
- **Tools**: Read, Grep, Glob

---

### 17. **use_cases_analyzer**
**Propósito**: CQRS commands/queries, use cases
- **Runtime**: 40-50 min
- **Output**: `code-analysis/domain-analysis/use_cases_analysis.md`
- **Table**: 10 (Use Cases)
- **Tools**: Read, Grep, Glob

---

### 18. **events_analyzer**
**Propósito**: Domain events, Temporal workflows
- **Runtime**: 40-50 min
- **Output**: `code-analysis/domain-analysis/events_analysis.md`
- **Tables**: 11 (Domain Events)
- **Tools**: Read, Grep, Glob, Bash

---

## 🟣 META Agents (4 agentes)

Orquestração e pós-processamento:

### 19. **orchestrator** 🎯
**Propósito**: Coordena todos os 18 agentes de análise
- **Runtime**: 8-12 horas (parallelized to 2-3 hours)
- **Output**: `code-analysis/comprehensive/MASTER_ANALYSIS.md` (todas as 30 tabelas)
- **Dependencies**: Todos os 18 agentes acima
- **Priority**: CRITICAL
- **Tools**: Task (lança sub-agentes em paralelo)

**O que faz**:
1. Lança deterministic_analyzer (baseline)
2. Lança 18 agentes especializados em paralelo (batches de 10)
3. Aguarda todas as análises
4. Consolida em MASTER_ANALYSIS.md
5. Dispara todo_manager (atualiza TODO.md)
6. Dispara docs_index_manager (atualiza índices)

---

### 20. **adr_generator**
**Propósito**: Gera Architecture Decision Records (ADRs)
- **Runtime**: 30-40 min
- **Output**: `docs/adr/*.md` (17+ ADR files)
- **Dependencies**: Todas as análises completas
- **Priority**: MEDIUM
- **Tools**: Read, Write, Bash

**O que faz**:
- Lê análises consolidadas
- Gera ADRs para decisões arquiteturais (DDD, Hexagonal, Event-Driven, etc.)
- Formato: `0001-adopt-ddd.md`, `0002-hexagonal-architecture.md`

---

### 21. **docs_cleanup**
**Propósito**: Organiza documentação pós-análise
- **Runtime**: 20-30 min
- **Output**: Estrutura de docs limpa
- **Dependencies**: Todas as análises completas
- **Priority**: LOW
- **Tools**: Bash, Read, Write

**O que faz**:
- Move análises antigas para `code-analysis/archive/YYYY-MM-DD/`
- Limpa arquivos temporários
- Organiza estrutura de pastas

---

### 22. **docs_consolidator**
**Propósito**: Consolida documentação fragmentada
- **Runtime**: 30-40 min
- **Output**: AI_REPORT.md consolidado, TODO.md consolidado
- **Dependencies**: None (pode rodar a qualquer momento)
- **Priority**: HIGH
- **Tools**: Read, Edit, Write, Bash

**O que faz**:
- Merge AI_REPORT_PART1-6 → AI_REPORT.md consolidado
- Merge TODO.md (consolidated), todo_*.md → TODO.md consolidado
- Consolida docs fragmentadas (Python ADK, AI Memory, MCP Server)
- Arquiva fragmentos antigos

---

## 🆕 MANAGEMENT Agents (2 agentes)

Mantêm documentação atualizada automaticamente:

### 23. **todo_manager** ⭐ NOVO
**Propósito**: Mantém TODO.md consolidado e sincronizado com codebase
- **Runtime**: 10-15 min
- **Output**: `TODO.md` (raiz, sempre atualizado)
- **Dependencies**: Análises completas (security_analyzer, testing_analyzer)
- **Priority**: HIGH
- **Tools**: Read, Edit, Grep, Glob, Bash
- **Triggers**: `/update-todo`, pós-análise (automático), semanal (review)

**O que faz**:
1. **Consolida TODOs**: Merge de TODO.md, TODO.md (consolidated), todo_*.md
2. **Atualiza status**: Marca tarefas como completas baseado no código
   - Ex: P0-1 Security Fix → Verifica se middleware/auth.go foi corrigido
   - Ex: Optimistic Locking → Conta aggregates com `version int` field
3. **Re-prioriza**: Baseado em análises
   - security_analyzer encontra P0 → Adiciona ao Sprint 1-2
   - testing_analyzer identifica gaps → Adiciona tarefas de testes
4. **Sincroniza**: Cross-reference com análises
   - code-analysis/quality/security_analysis.md (P0 vulns)
   - code-analysis/quality/testing_analysis.md (coverage gaps)

**Exemplo de uso**:
```bash
# Manual
/update-todo

# Automático (dispara após orchestrator)
# Detecta: security_analyzer encontrou nova P0 SSRF
# Ação: Adiciona "Fix SSRF in webhooks" ao TODO.md Sprint 1-2
```

---

### 24. **docs_index_manager** 📚 NOVO
**Propósito**: Mantém índices de documentação atualizados
- **Runtime**: 5-10 min
- **Output**: README.md files (raiz, code-analysis/, docs/, docs/future/)
- **Dependencies**: None
- **Priority**: MEDIUM
- **Tools**: Read, Edit, Glob, Bash
- **Triggers**: `/update-indexes`, pós-consolidação, pós-análise

**O que faz**:
1. **Escaneia diretórios**: Detecta novos arquivos .md
2. **Atualiza índices**:
   - `README.md` (raiz) → Quick links para docs principais
   - `code-analysis/README.md` → Índice de todas as análises
   - `docs/README.md` → Hub de documentação
   - `docs/future/README.md` → Roadmap de features
3. **Detecta mudanças**: Adiciona novos arquivos automaticamente
4. **Mantém consistência**: Links corretos, sem broken references

**Exemplo de uso**:
```bash
# Manual
/update-indexes

# Automático (dispara após docs_consolidator)
# Detecta: Nova análise em code-analysis/ai-ml/ai_cost_tracking.md
# Ação: Adiciona ao code-analysis/README.md automaticamente
```

---

## 📋 Ordem de Execução

### ⚡ Análise Completa (30 Tabelas)

#### **Fase 0: Baseline** (5-10 min) ⭐
```bash
# SEMPRE rodar primeiro - baseline determinístico
claude-code --agent deterministic_analyzer
```

#### **Fase 1: Análise Core** (Parallel) - 70-80 min
```bash
# CRITICAL + HIGH priority (9 agentes em paralelo)
claude-code --agent domain_model_analyzer &
claude-code --agent testing_analyzer &
claude-code --agent ai_ml_analyzer &
claude-code --agent security_analyzer &
claude-code --agent integration_analyzer &
claude-code --agent infrastructure_analyzer &
claude-code --agent resilience_analyzer &
claude-code --agent api_analyzer &
wait
```

#### **Fase 2: Análise Especializada** (Parallel) - 50-70 min
```bash
# MEDIUM + USER-REQUESTED + STANDARD priority (9 agentes em paralelo)
claude-code --agent persistence_analyzer &
claude-code --agent data_quality_analyzer &
claude-code --agent code_style_analyzer &
claude-code --agent documentation_analyzer &
claude-code --agent solid_principles_analyzer &
claude-code --agent value_objects_analyzer &
claude-code --agent entity_relationships_analyzer &
claude-code --agent use_cases_analyzer &
claude-code --agent events_analyzer &
wait
```

#### **Fase 3: Orquestração** (10-15 min)
```bash
# Orchestrator agrega todos os resultados
claude-code --agent orchestrator

# Output: code-analysis/comprehensive/MASTER_ANALYSIS.md
```

#### **Fase 4: Gerenciamento** (15-20 min)
```bash
# Atualiza TODO.md baseado nas análises
claude-code --agent todo_manager

# Atualiza todos os índices
claude-code --agent docs_index_manager
```

#### **Fase 5: Pós-Processamento** (Opcional) - 50-70 min
```bash
# Gera ADRs
claude-code --agent adr_generator

# Consolida documentação fragmentada
claude-code --agent docs_consolidator

# Limpa e arquiva
claude-code --agent docs_cleanup
```

---

## ⏱️ Runtime Total

- **Fase 0**: 5-10 min (baseline)
- **Fase 1**: 70-80 min (parallelized - core analysis)
- **Fase 2**: 50-70 min (parallelized - specialized analysis)
- **Fase 3**: 10-15 min (orchestration)
- **Fase 4**: 15-20 min (management)
- **Fase 5**: 50-70 min (post-processing, opcional)

**Total Mínimo** (Fase 0-4): ~2.5-3 horas (parallelizado)
**Total Completo** (Fase 0-5): ~3.5-5 horas (parallelizado)

**Sem paralelização**: ~15-20 horas

---

## 📁 Estrutura de Output Completa

```
code-analysis/
├── README.md                             # Índice de análises (mantido por docs_index_manager)
│
├── architecture/                         # Métricas arquiteturais
│   ├── AI_REPORT.md                     # Report consolidado (7 partes → 1)
│   ├── deterministic_metrics.md         # Baseline factual
│   └── architecture_scores.md           # Scores por camada
│
├── domain-analysis/                      # Análises DDD
│   ├── domain_model_analysis.md         # Aggregates, eventos, repos
│   ├── value_objects_analysis.md        # Value objects
│   ├── entity_relationships_analysis.md # Relacionamentos
│   ├── use_cases_analysis.md            # CQRS commands/queries
│   └── events_analysis.md               # Domain events
│
├── infrastructure/                       # Infraestrutura
│   ├── persistence_analysis.md          # Database, migrations
│   ├── api_analysis.md                  # REST endpoints, DTOs
│   ├── integration_analysis.md          # APIs externas
│   └── infrastructure_analysis.md       # Deploy, CI/CD
│
├── quality/                              # Qualidade de código
│   ├── testing_analysis.md              # Cobertura de testes
│   ├── security_analysis.md             # OWASP, vulnerabilidades
│   ├── code_style_analysis.md           # Go idioms, naming
│   ├── documentation_analysis.md        # Swagger, godoc
│   ├── solid_principles_analysis.md     # S.O.L.I.D.
│   ├── resilience_analysis.md           # Rate limiting, circuit breaker
│   └── data_quality_analysis.md         # Query performance
│
├── ai-ml/                                # AI/ML
│   ├── ai_ml_analysis.md                # Features, providers
│   └── ai_cost_tracking.md              # Custos por provider
│
├── comprehensive/                        # Reports completos
│   └── MASTER_ANALYSIS.md               # Output do orchestrator (30 tabelas)
│
└── archive/                              # Análises antigas
    └── 2025-10-15/
        ├── ai_reports/                   # AI_REPORT_PART1-6
        ├── todos/                        # TODO variants
        └── analysis_reports/             # Old reports
```

---

## 🎯 Princípios de Design dos Agentes

Todos os 24 agentes seguem metodologia consistente:

### 1. **Deterministic Baseline First** 🔢
- Rodar grep/find/wc para contagens factuais
- Validar análises AI com números determinísticos
- Exemplo: "AI detectou 30 aggregates" ✅ vs "grep encontrou 30 arquivos" ✅

### 2. **AI Analysis** 🤖
- Analisar padrões, qualidade, compliance
- Score de 1-10 com justificativa
- Identificar violations, gaps, best practices

### 3. **Comparison** ⚖️
- Mostrar "Deterministic vs AI" side-by-side
- Validar consistência
- Explicar discrepâncias

### 4. **Evidence Required** 📍
- Toda finding precisa de file:line citation
- Exemplos de código (✅ Good vs ❌ Bad)
- Links para documentação relevante

### 5. **Atemporal Design** ⏳
- **NUNCA** hardcode números no prompt
- Uso: "All aggregates" ✅ ao invés de "30 aggregates" ❌
- Números descobertos dinamicamente via grep/glob

### 6. **Code Examples** 💻
- Sempre mostrar exemplos práticos
- Formato: ✅ Correct vs ❌ Incorrect
- Com explicação do "why"

### 7. **Scoring with Reasoning** 📊
- Todo score 1-10 precisa de justificativa
- Critérios claros e consistentes
- Comparar com industry best practices

---

## 🔄 Diagrama de Interação dos Agentes

```
                    ┌─────────────────────────┐
                    │ deterministic_analyzer  │ ⭐ BASELINE
                    │   (SEMPRE PRIMEIRO)     │
                    └───────────┬─────────────┘
                                │
                                ▼
                ┌───────────────────────────────────────┐
                │   18 Agentes de Análise               │
                │   (executam em paralelo)              │
                │                                       │
                │   🔴 CRITICAL (6 agentes)             │
                │   - domain_model_analyzer             │
                │   - testing_analyzer                  │
                │   - ai_ml_analyzer                    │
                │   - security_analyzer                 │
                │   - integration_analyzer              │
                │   - persistence_analyzer              │
                │                                       │
                │   🟠 HIGH (3 agentes)                 │
                │   🟡 MEDIUM (2 agentes)               │
                │   🔵 USER-REQUESTED (3 agentes)       │
                │   ⚪ STANDARD (4 agentes)             │
                └───────────┬───────────────────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ orchestrator  │ 🎯 Agrega todos os resultados
                    └───────┬───────┘
                            │
                ┌───────────┴───────────┐
                │                       │
                ▼                       ▼
        ┌──────────────┐        ┌──────────────────┐
        │ todo_manager │ 🆕     │ docs_index_      │ 🆕
        │              │        │ manager          │
        │ - Atualiza   │        │ - Atualiza       │
        │   TODO.md    │        │   índices        │
        └──────────────┘        └──────────────────┘
                │                       │
                └───────────┬───────────┘
                            │
                            ▼
                ┌─────────────────────────────┐
                │  Pós-Processamento          │
                │  (Opcional)                 │
                │                             │
                │  - adr_generator            │
                │  - docs_consolidator        │
                │  - docs_cleanup             │
                └─────────────────────────────┘
```

---

## 🛠️ Comandos Úteis

### Listar Agentes
```bash
# Todos os agentes
ls -1 .claude/agents/*.md | grep -v README

# Contar agentes
ls -1 .claude/agents/*.md | grep -v README | wc -l
# Output: 24

# Ver detalhes de um agente
cat .claude/agents/domain_model_analyzer.md
```

### Executar Agentes
```bash
# Análise completa (30 tabelas) via orchestrator
claude-code --agent orchestrator

# Análise específica
claude-code --agent security_analyzer

# Baseline (sempre primeiro)
claude-code --agent deterministic_analyzer

# Atualizar TODO.md
claude-code --agent todo_manager

# Atualizar índices
claude-code --agent docs_index_manager

# Consolidar docs fragmentadas
claude-code --agent docs_consolidator
```

### Slash Commands
```bash
# Análise completa
/analyze-all

# Atualizar TODO
/update-todo

# Atualizar índices
/update-indexes

# Consolidar documentação
/consolidate-docs

# Check de segurança P0
/security:p0-check
```

---

## 📊 Cobertura das 30 Tabelas

### Especialização por Agente

| Agente | Tabelas | Categoria |
|--------|---------|-----------|
| deterministic_analyzer | Baseline | Métricas factuais |
| domain_model_analyzer | 1, 2, 5 | Aggregates, Events, Children |
| entity_relationships_analyzer | 4 | Foreign Keys, Cardinality |
| value_objects_analyzer | 6 | Value Objects |
| persistence_analyzer | 3, 7, 9 | Entities, Normalization, Migrations |
| integration_analyzer | 8, 12 | Integrations, Event Bus |
| use_cases_analyzer | 10 | CQRS Commands/Queries |
| events_analyzer | 11 | Domain Events |
| data_quality_analyzer | 13, 14, 15 | Query Perf, Consistency, Validations |
| api_analyzer | 16, 17 | DTOs, REST Endpoints |
| security_analyzer | 18, 21, 24-27 | OWASP, AI Security, Integration Tests |
| resilience_analyzer | 19, 20, 23 | Rate Limiting, Error Handling, Patterns |
| testing_analyzer | 22, 24, 25 | Test Pyramid, Integration Tests, Mocks |
| ai_ml_analyzer | 28 | AI/ML Features, Providers |
| infrastructure_analyzer | 29, 30 | Deployment, CI/CD, Roadmap |

### Análises Adicionais de Qualidade

| Agente | Foco |
|--------|------|
| code_style_analyzer | Go idioms, naming, organization |
| documentation_analyzer | Swagger, godoc, guides |
| solid_principles_analyzer | S.O.L.I.D. principles |

**Resultado**: 100% das 30 tabelas cobertas + análises extras de qualidade ✅

---

## 🔧 Manutenção

### Adicionar Novo Agente

1. **Criar arquivo**: `.claude/agents/{category}_{name}_analyzer.md`
2. **Seguir template**:
```markdown
---
name: {category}_{name}_analyzer
description: |
  {quando usar este agente}
tools: Read, Grep, Glob, Bash
model: sonnet
priority: {high|medium|low}
---

# {Name} Analyzer Agent

## Output Location
`code-analysis/{category}/{name}_analysis.md`

[... resto do prompt seguindo padrão ouro]
```

3. **Atualizar este README**: Adicionar na seção de prioridade correta
4. **Atualizar orchestrator.md**: Incluir novo agente na lista
5. **Testar**: Rodar agente standalone
6. **Validar output**: Verificar se gera em code-analysis/

### Atualizar Agente Existente

1. **Ler agente**: `cat .claude/agents/{agent}.md`
2. **Editar**: Atualizar prompt/tools/output path
3. **Testar**: Rodar agente
4. **Validar**: Comparar output antes/depois

### Deletar Agente (Raro)

1. **Remover arquivo**: `rm .claude/agents/{agent}.md`
2. **Atualizar README**: Remover da lista
3. **Atualizar orchestrator.md**: Remover das dependências
4. **Update docs_index_manager**: Para não incluir mais

---

## 🎯 Casos de Uso

### 1. Análise Completa de Codebase (Primeira Vez)
```bash
# Tempo: ~3 horas (parallelized)
claude-code --agent orchestrator

# Output:
# - code-analysis/comprehensive/MASTER_ANALYSIS.md (30 tabelas)
# - TODO.md atualizado com P0s encontrados
# - Índices atualizados
```

### 2. Re-análise Após Mudanças (Sprint Review)
```bash
# Re-rodar análises críticas (30 min)
claude-code --agent security_analyzer
claude-code --agent testing_analyzer

# Atualizar TODO baseado em novas findings
claude-code --agent todo_manager

# Output: TODO.md sincronizado com estado atual
```

### 3. Consolidar Documentação Fragmentada
```bash
# Merge AI_REPORT_PART1-6, TODO.md (consolidated), etc.
claude-code --agent docs_consolidator

# Atualizar índices
claude-code --agent docs_index_manager

# Output:
# - code-analysis/architecture/AI_REPORT.md (consolidado)
# - TODO.md (consolidado)
# - Índices atualizados
```

### 4. Auditoria de Segurança P0
```bash
# Análise de segurança focada
claude-code --agent security_analyzer

# Se P0 encontrado, atualizar TODO automaticamente
claude-code --agent todo_manager

# Output:
# - code-analysis/quality/security_analysis.md
# - TODO.md com nova P0 no Sprint 1-2
```

### 5. Validar Coverage de Testes
```bash
# Análise de testes
claude-code --agent testing_analyzer

# Output:
# - code-analysis/quality/testing_analysis.md
# - Identificar use cases sem testes
# - TODO.md atualizado com tarefas de testes
```

---

## 📚 Referências

### Documentação Interna
- `ai-guides/claude-code-guide.md` - Sistema multi-agente completo
- `ai-guides/prompt-engineering-guide.md` - 15+ técnicas de prompting
- `TODO.md` - Roadmap consolidado (mantido por todo_manager)
- `CLAUDE.md` - Instruções para Claude Code

### Agentes Relacionados
- Todos os 24 agentes em `.claude/agents/`
- Slash commands em `.claude/commands/`

### Outputs Gerados
- `code-analysis/` - Todas as análises
- `docs/adr/` - Architecture Decision Records
- `docs/future/` - Documentação de features planejadas

---

## 📝 Changelog

### v4.0 (2025-10-15) - CURRENT
- ✅ Adicionado `todo_manager` (agente 23)
- ✅ Adicionado `docs_index_manager` (agente 24)
- ✅ Atualizada estrutura de output para `code-analysis/`
- ✅ Reorganizada categorização de agentes
- ✅ Total: 24 agentes (18 análise + 4 meta + 2 gerenciamento)

### v3.0 (2025-10-14)
- ✅ 22 agentes (18 análise + 4 meta)
- ✅ Cobertura 100% das 30 tabelas
- ✅ Padrão ouro completo implementado
- ✅ Outputs em `code-analysis/`

### v2.0 (2025-10-13)
- ✅ 18 agentes de análise especializados
- ✅ 4 agentes meta (orchestrator, adr_generator, docs_cleanup, docs_consolidator)

### v1.0 (2025-10-12)
- ✅ Sistema inicial de agentes

---

**Version**: 4.0
**Last Updated**: 2025-10-15
**Total Agents**: 24 (18 análise + 4 meta + 2 gerenciamento)
**Coverage**: 100% das 30 tabelas de análise
**Output Structure**: `code-analysis/` (organizado por categoria)
**Estimated Runtime**: 2.5-3 horas (parallelized) / 15-20 horas (sequential)

---

**Maintainer**: Ventros CRM Team
**Status**: ✅ Sistema completo de análise + gerenciamento de documentação
