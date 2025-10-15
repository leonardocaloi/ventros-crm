# 🚀 Sistema Multi-Agente do Claude Code - Guia Completo

## 📋 Visão Geral

O Claude Code é essencialmente um **sistema multi-agente orquestrado** onde você pode:
- Criar **sub-agentes especializados** com contextos isolados
- Executar até **10 tarefas em paralelo** simultaneamente
- Escalar para **100+ tarefas** através de filas inteligentes
- Cada agente tem sua **própria janela de contexto** (context window)

---

## 🏗️ Arquitetura: Onde Configurar

### 1. **Sub-Agentes** (`.claude/agents/`)

```
ventros-crm/
├── .claude/
│   ├── agents/                    # Agentes do projeto (versionados no git)
│   │   ├── domain-analyzer.md     # Analisa camada de domínio
│   │   ├── infrastructure-reviewer.md  # Revisa infraestrutura
│   │   ├── test-specialist.md     # Especialista em testes
│   │   ├── security-auditor.md    # Auditoria de segurança
│   │   └── orchestrator.md        # Orquestrador master
│   │
│   └── commands/                  # Comandos personalizados (slash commands)
│       ├── analyze-all.md         # /analyze-all
│       ├── security/              # Namespace: /security:*
│       │   ├── audit.md           # /security:audit
│       │   └── p0-check.md        # /security:p0-check
│       └── domain/                # Namespace: /domain:*
│           ├── check.md           # /domain:check
│           └── coverage.md        # /domain:coverage
│
└── ~/.claude/                     # Configuração global do usuário
    ├── agents/                    # Agentes disponíveis em todos os projetos
    └── commands/                  # Comandos globais
```

---

## 📝 Estrutura de um Sub-Agente

### Template Completo (YAML + Markdown)

```markdown
---
name: domain-analyzer
description: |
  Specialized agent for analyzing Domain-Driven Design patterns in the codebase.
  Use when you need to: validate aggregates, check event naming, verify repository
  patterns, or audit domain layer purity (no infrastructure dependencies).
tools: Read, Grep, Glob, Bash
model: sonnet
priority: high
version: "1.0"
author: Ventros Team
---

# Domain Analyzer Agent

You are a **Domain-Driven Design expert** specialized in analyzing Go codebases
following Hexagonal Architecture.

## Your Responsibilities

1. **Aggregate Analysis**
   - Check for `version int` field (optimistic locking)
   - Verify event emission in domain methods
   - Validate repository interfaces

2. **Event Naming Convention**
   - Format: `aggregate.action` (lowercase, past tense)
   - Examples: `contact.created`, `session.ended`

3. **Layer Boundaries**
   - Domain layer must NOT import infrastructure packages
   - Check for violations: `grep -r "infrastructure/" internal/domain/`

## Your Approach

1. Start by reading `AI_REPORT.md` to understand architecture score
2. Use Glob to find all aggregate files: `internal/domain/**/aggregate.go`
3. Use Grep to check for patterns
4. Report findings in structured markdown

## Output Format

```markdown
## Domain Analysis Report

### ✅ Compliant Aggregates (X/30)
- contact: Has version field ✓, Events ✓, Repository ✓
- session: Has version field ✓, Events ✓, Repository ✓

### ❌ Issues Found
- **campaign**: Missing `version int` field (optimistic locking)
- **sequence**: Event naming violation: "SequenceCreated" should be "sequence.created"

### 🔍 Recommendations
1. Add version fields to 14 remaining aggregates (see TODO.md P0-2)
2. Fix event naming in automation bounded context
```

## Tools You Can Use
- **Read**: Read aggregate files, AI_REPORT.md, TODO.md
- **Grep**: Search for patterns across codebase
- **Glob**: Find files matching patterns
- **Bash**: Run go vet, go test for validation

## Critical Rules
- NEVER modify code - only analyze and report
- ALWAYS check AI_REPORT.md first for context
- ALWAYS follow the output format above
- Focus on DDD patterns, not general Go best practices
```

---

## 🎯 Campos da Configuração YAML

| Campo | Obrigatório | Valores | Descrição |
|-------|-------------|---------|-----------|
| `name` | ✅ | lowercase-with-hyphens | Identificador único |
| `description` | ✅ | string multi-linha | Quando usar este agente |
| `tools` | ❌ | Read, Write, Edit, Grep, Glob, Bash | Ferramentas permitidas |
| `model` | ❌ | `sonnet`, `opus`, `haiku`, `inherit` | Modelo a usar |
| `priority` | ❌ | high, medium, low | Prioridade de delegação |
| `version` | ❌ | string | Versionamento do agente |
| `author` | ❌ | string | Quem criou |

---

## ⚡ Paralelização: Como Funciona

### Limites e Comportamento

```bash
# Limite de paralelismo: 10 tarefas simultâneas
# Máximo de tarefas em fila: 100+ (com queueing inteligente)

# Execução automática (Claude decide)
"Explore the codebase in parallel"

# Execução explícita (você controla)
"Using 4 subagents, analyze:
1. Domain layer (domain-analyzer)
2. Application layer (application-reviewer)
3. Infrastructure layer (infrastructure-reviewer)
4. Tests (test-specialist)"

# Paralelismo máximo
"Using 10 subagents in parallel, explore these directories: ..."
```

### Padrões de Orquestração

#### 1. **Investigação Paralela** (Parallel Investigation)
```
Agentes trabalham simultaneamente em aspectos diferentes da mesma tarefa

[Main Agent]
     ├─→ [Backend Specialist] → Analisa internal/
     ├─→ [Frontend Specialist] → Analisa infrastructure/http/
     ├─→ [Test Specialist] → Analisa testes
     └─→ [Docs Specialist] → Analisa documentação
```

#### 2. **Handoff Sequencial** (Sequential Handoff)
```
Saída de um agente é entrada do próximo

[Product Manager] → Define requisitos
       ↓
[UX Designer] → Cria especificações
       ↓
[Senior Engineer] → Implementa
       ↓
[Code Reviewer] → Valida qualidade
```

#### 3. **Orquestrador + Workers** (Orchestrator Pattern)
```
Um agente coordena, outros executam

[Orchestrator Agent]
     ├─→ [Domain Analyzer] → Reporta findings
     ├─→ [Security Auditor] → Reporta vulnerabilities
     ├─→ [Test Runner] → Reporta coverage
     └─→ [Consolidator] → Merge reports → Final report
```

---

## 🎨 Comandos Personalizados (Slash Commands)

### Estrutura de um Comando

```markdown
---
argument-hint: [layer] [depth]
description: Analyze specific architectural layer with configurable depth
---

# Architectural Layer Analysis

Analyze the **$1** layer of the codebase with depth level **$2**.

## Layers Available
- domain: Pure business logic (internal/domain/)
- application: Use cases (internal/application/)
- infrastructure: External concerns (infrastructure/)

## Depth Levels
- 1: Quick scan (file count, LOC)
- 2: Medium (+ imports, dependencies)
- 3: Deep (+ patterns, violations, recommendations)

## Instructions
1. Use Glob to find all files in the specified layer
2. Use Grep to analyze patterns based on depth level
3. Generate report with findings

Start the analysis now.
```

**Uso**: `/analyze-layer domain 3`

### Comandos com Namespaces

```bash
.claude/commands/
├── security/
│   ├── audit.md         # /security:audit
│   ├── p0-check.md      # /security:p0-check
│   └── rbac-verify.md   # /security:rbac-verify
│
├── domain/
│   ├── check.md         # /domain:check
│   ├── coverage.md      # /domain:coverage
│   └── events.md        # /domain:events
│
└── test/
    ├── unit.md          # /test:unit
    ├── integration.md   # /test:integration
    └── e2e.md           # /test:e2e
```

### Comando Master (Orquestrador)

```markdown
---
description: Run complete codebase analysis using multiple specialized agents in parallel
---

# 🎯 Complete Codebase Analysis

Run a comprehensive analysis using specialized agents in parallel.

## Phase 1: Parallel Analysis (10 agents)

Launch these agents simultaneously:

1. **domain-analyzer**: Analyze DDD patterns, aggregates, events
2. **application-reviewer**: Check command/query handlers, use cases
3. **infrastructure-reviewer**: Verify repositories, HTTP handlers, messaging
4. **security-auditor**: Check for P0 vulnerabilities (TODO.md)
5. **test-specialist**: Analyze test coverage, quality
6. **database-specialist**: Check migrations, RLS policies
7. **event-specialist**: Verify Outbox Pattern, event naming
8. **api-specialist**: Check Swagger docs, endpoint coverage
9. **performance-specialist**: Analyze bottlenecks, caching
10. **docs-specialist**: Verify documentation completeness

## Phase 2: Consolidation (1 agent)

After all agents complete, launch **report-consolidator** to:
- Merge all findings into single report
- Prioritize issues by severity (P0 > P1 > P2)
- Generate actionable recommendations
- Create updated TODO.md if needed

## Output Location

Save final report to: `ai-guides/analysis-reports/analysis-YYYY-MM-DD.md`

## Execution

Use the Task tool to launch all agents in parallel. Start now.
```

**Uso**: `/analyze-all`

---

## 🎭 Exemplo Prático: Sistema Multi-Agente para Ventros CRM

### Agentes Especializados

#### 1. **domain-analyzer.md**
```markdown
---
name: domain-analyzer
description: Analyze Domain layer (DDD patterns, aggregates, events, repositories)
tools: Read, Grep, Glob
model: sonnet
---
[Prompt detalhado acima]
```

#### 2. **security-auditor.md**
```markdown
---
name: security-auditor
description: Audit security vulnerabilities, especially P0 issues from TODO.md
tools: Read, Grep, Glob, Bash
model: sonnet
priority: high
---

# Security Auditor Agent

You are a **security expert** focused on identifying vulnerabilities in Go web applications.

## Your Mission

Audit the Ventros CRM codebase for the **5 critical P0 vulnerabilities** listed in TODO.md:

1. **Dev Mode Bypass (CVSS 9.1)** - `middleware/auth.go:41`
2. **SSRF in Webhooks (CVSS 9.1)** - No URL validation
3. **BOLA in 60 GET endpoints (CVSS 8.2)** - No ownership checks
4. **Resource Exhaustion (CVSS 7.5)** - No max page size
5. **RBAC Missing (CVSS 7.1)** - 95 endpoints lack role checks

## Your Approach

1. Read `TODO.md` to get full context on vulnerabilities
2. For each vulnerability:
   - Locate the vulnerable code with Grep/Read
   - Verify if it's still present (may have been fixed)
   - Rate severity (Critical, High, Medium, Low)
   - Suggest specific fix with code example
3. Check for NEW vulnerabilities not in TODO.md

## Output Format

```markdown
## 🔒 Security Audit Report

### Critical Vulnerabilities (P0)

#### 1. Dev Mode Bypass (CVSS 9.1) ❌ STILL PRESENT
**Location**: `infrastructure/http/middleware/auth.go:41`
**Issue**: Production environment can bypass authentication
**Evidence**:
[code snippet]
**Fix**:
[specific code change]

#### 2. SSRF in Webhooks (CVSS 9.1) ✅ FIXED
**Status**: Implemented URL validation in commit `abc123`
```
```

#### 3. **test-specialist.md**
```markdown
---
name: test-specialist
description: Analyze test coverage, quality, and run test suites
tools: Read, Grep, Glob, Bash
model: sonnet
---

# Test Specialist Agent

You analyze test coverage and quality in Go projects.

## Your Tasks

1. **Run Test Suites**
   ```bash
   make test-unit        # Unit tests (~2 min)
   make test-integration # Integration tests (~10 min, requires: make infra)
   make test-e2e        # E2E tests (~10 min, requires: make infra + make api)
   ```

2. **Analyze Coverage**
   ```bash
   make test-coverage    # Generate HTML report
   ```

3. **Check Test Quality**
   - Are tests following table-driven pattern?
   - Are mocks used properly (testify/mock)?
   - Are assertions clear and specific?
   - Are tests testing behavior, not implementation?

## Coverage Goals (from CLAUDE.md)
- Domain Layer: 100% (business-critical)
- Application Layer: 80%+
- Infrastructure Layer: 60%+
- Overall: 82%+ (current: 82%)

## Output Format

```markdown
## 🧪 Test Analysis Report

### Test Execution Results
- Unit: ✅ 61 tests passed (2m 14s)
- Integration: ✅ 2 tests passed (8m 32s)
- E2E: ❌ 1 test failed (see details)

### Coverage Analysis
- Domain: 98% ⚠️ (2% below target)
- Application: 85% ✅
- Infrastructure: 67% ✅
- Overall: 82% ✅

### Failed Test Details
[detailed analysis of failures]

### Recommendations
1. Add tests for Contact.UpdateCustomFields() (domain)
2. Improve error case coverage in MessageHandler
```
```

#### 4. **orchestrator.md** (Agente Master)
```markdown
---
name: orchestrator
description: Orchestrate multiple specialized agents to perform complex analysis workflows
tools: Task, Read, Write
model: opus
priority: highest
---

# Orchestrator Agent

You are the **master orchestrator** that coordinates multiple specialized sub-agents.

## Your Role

1. **Understand the request** - Break down complex tasks
2. **Delegate to specialists** - Launch appropriate sub-agents in parallel
3. **Consolidate results** - Merge findings into coherent report
4. **Manage conflicts** - Resolve overlapping findings
5. **Update documentation** - Keep CLAUDE.md, TODO.md, AI_REPORT.md synchronized

## Available Sub-Agents

| Agent | Specialty | When to Use |
|-------|-----------|-------------|
| domain-analyzer | DDD patterns | Domain layer analysis |
| security-auditor | Security vulns | Security audits |
| test-specialist | Testing | Test coverage/quality |
| infrastructure-reviewer | Infrastructure | DB, HTTP, messaging |
| performance-specialist | Performance | Bottlenecks, optimization |

## Orchestration Patterns

### Pattern 1: Full Codebase Analysis
```
1. Launch 5 agents in parallel (domain, security, test, infra, perf)
2. Wait for all to complete
3. Consolidate reports
4. Generate action items prioritized by severity
5. Update TODO.md if needed
```

### Pattern 2: Security Sprint
```
1. Launch security-auditor for P0 audit
2. Launch test-specialist to verify security test coverage
3. Consolidate findings
4. Generate security roadmap
```

### Pattern 3: Architecture Audit
```
1. Launch domain-analyzer
2. Launch infrastructure-reviewer
3. Check for layer boundary violations
4. Update AI_REPORT.md score if needed
```

## Critical Rules

- **ALWAYS use Task tool** to launch sub-agents
- **Launch agents in PARALLEL** when possible (up to 10 concurrent)
- **Wait for completion** before consolidating
- **Resolve conflicts** - if agents disagree, investigate deeper
- **Update docs** - Keep documentation synchronized
- **Never create new docs** - Update existing CLAUDE.md, TODO.md, AI_REPORT.md

## Example Orchestration

User: "Analyze the entire codebase"

You respond:
"I'll orchestrate a comprehensive analysis using 5 specialized agents in parallel."

[Launch 5 Task tool calls in parallel]
[Wait for all results]
[Consolidate findings]
[Present unified report]
```

---

## 🚀 Como Executar

### Método 1: Delegação Automática
```bash
# Claude decide automaticamente qual agente usar
"Analyze the domain layer for DDD violations"

# Claude internamente:
# - Detecta que é análise de domínio
# - Delega para domain-analyzer agent
# - Retorna resultados
```

### Método 2: Invocação Explícita
```bash
# Você especifica o agente
"Use the security-auditor agent to check for P0 vulnerabilities"

# Ou com múltiplos agentes
"Using these agents in parallel:
1. domain-analyzer - check aggregates
2. security-auditor - check P0s
3. test-specialist - check coverage"
```

### Método 3: Slash Command
```bash
# Usar comando personalizado que orquestra tudo
/analyze-all

# Ou comandos específicos
/security:p0-check
/domain:coverage
/test:integration
```

### Método 4: Programático (Task Tool)
```markdown
# O agente principal usa a Task tool

I'll launch 3 agents in parallel to analyze different layers:

[Task tool call #1: domain-analyzer]
[Task tool call #2: security-auditor]
[Task tool call #3: test-specialist]

[Wait for results]
[Consolidate and present]
```

---

## 📊 Estratégia Multi-Agente para Ventros CRM

### Estrutura Proposta

```
.claude/
├── agents/
│   ├── orchestrator.md              # Master orchestrator
│   ├── domain-analyzer.md           # Domain layer (DDD)
│   ├── application-reviewer.md      # Application layer (CQRS)
│   ├── infrastructure-reviewer.md   # Infrastructure layer
│   ├── security-auditor.md          # Security (P0 focus)
│   ├── test-specialist.md           # Testing & coverage
│   ├── database-specialist.md       # Migrations, RLS, Outbox
│   ├── event-specialist.md          # Event naming, Outbox Pattern
│   ├── api-specialist.md            # Swagger, endpoints
│   ├── performance-specialist.md    # Bottlenecks, caching
│   └── docs-manager.md              # Documentation sync
│
└── commands/
    ├── analyze-all.md               # /analyze-all
    ├── security/
    │   ├── p0-check.md              # /security:p0-check
    │   ├── audit.md                 # /security:audit
    │   └── rbac-verify.md           # /security:rbac-verify
    ├── domain/
    │   ├── check.md                 # /domain:check
    │   ├── coverage.md              # /domain:coverage
    │   └── events.md                # /domain:events
    ├── infra/
    │   ├── migrations.md            # /infra:migrations
    │   └── rls.md                   # /infra:rls
    └── test/
        ├── run-all.md               # /test:run-all
        ├── unit.md                  # /test:unit
        └── coverage.md              # /test:coverage
```

### Workflow de Análise Completa

```bash
# 1. Usuário executa comando master
/analyze-all

# 2. Comando dispara orchestrator.md que:
#    - Lê CLAUDE.md, TODO.md, AI_REPORT.md para contexto
#    - Lança 10 agentes em paralelo:

[Orchestrator]
    ├─→ [domain-analyzer]           # Analisa 30 aggregates
    ├─→ [application-reviewer]      # Analisa 80+ commands
    ├─→ [infrastructure-reviewer]   # Analisa repositories, handlers
    ├─→ [security-auditor]          # Checa 5 P0s
    ├─→ [test-specialist]           # Roda make test, analisa coverage
    ├─→ [database-specialist]       # Valida migrations, RLS
    ├─→ [event-specialist]          # Valida Outbox Pattern, naming
    ├─→ [api-specialist]            # Valida Swagger, 158 endpoints
    ├─→ [performance-specialist]    # Analisa latência, bottlenecks
    └─→ [docs-manager]              # Verifica docs atualizados

# 3. Após todos completarem (2-5 min), orchestrator:
#    - Consolida findings
#    - Prioriza por severidade (P0 > P1 > P2)
#    - Gera relatório unificado
#    - Atualiza TODO.md se necessário

# 4. Output: ai-guides/analysis-reports/analysis-2025-10-15.md
```

---

## 🎯 Agente Especial: docs-manager.md

```markdown
---
name: docs-manager
description: |
  Manages documentation updates across CLAUDE.md, TODO.md, AI_REPORT.md.
  NEVER creates new docs - only updates existing ones. Prevents doc sprawl.
tools: Read, Edit
model: sonnet
priority: high
---

# Documentation Manager Agent

You are the **single source of truth** for documentation updates in Ventros CRM.

## Critical Rules

❌ **NEVER CREATE NEW DOCUMENTATION FILES**
✅ **ONLY UPDATE EXISTING FILES**: CLAUDE.md, TODO.md, AI_REPORT.md

## Your Responsibilities

1. **Keep Docs Synchronized**
   - When architecture changes → Update CLAUDE.md + AI_REPORT.md
   - When tasks complete → Update TODO.md (mark as done)
   - When new patterns emerge → Update CLAUDE.md best practices

2. **Prevent Documentation Sprawl**
   - If someone asks to create new docs → Update existing ones instead
   - Consolidate scattered information into main docs
   - Remove obsolete sections

3. **Maintain Quality**
   - Keep formatting consistent
   - Update "Last Updated" dates
   - Verify accuracy (cross-check with codebase)

## Documentation Hierarchy

| File | Purpose | Update Frequency |
|------|---------|------------------|
| CLAUDE.md | Complete dev guide | Every major change |
| TODO.md | Roadmap + priorities | Every sprint |
| AI_REPORT.md | Architecture audit | Every major refactor |

## Example Updates

### Scenario: Optimistic Locking Complete
```markdown
# Before (TODO.md):
- [ ] P0-2: Add optimistic locking (version field) to 14 aggregates (16/30 = 53%)

# After (TODO.md):
- [x] P0-2: Add optimistic locking (version field) - COMPLETE (30/30 = 100%) ✅

# Also update:
# CLAUDE.md: "Optimistic locking is only 53% complete" → "100% complete"
# AI_REPORT.md: Update architecture score if significant
```

## Output Format

```markdown
## Documentation Updates

### Files Modified
- ✅ CLAUDE.md (updated optimistic locking status)
- ✅ TODO.md (marked P0-2 as complete)
- ⏭️ AI_REPORT.md (no changes needed)

### Changes Made
1. CLAUDE.md line 423: Updated status from 53% to 100%
2. TODO.md line 67: Changed [ ] to [x] for P0-2
3. Updated "Last Updated" dates

### Verification
- Cross-checked with codebase: grep -r "version int" internal/domain/
- Confirmed 30/30 aggregates have version field ✅
```
```

---

## 💡 Dicas e Best Practices

### 1. **Limite de Ferramentas**
```markdown
# ❌ Má prática: Dar todas as ferramentas
tools: Read, Write, Edit, Grep, Glob, Bash, WebFetch, WebSearch

# ✅ Boa prática: Apenas o necessário
tools: Read, Grep, Glob  # Agente read-only de análise
```

### 2. **Descrição Clara**
```markdown
# ❌ Vago
description: Analyze code

# ✅ Específico
description: |
  Analyze Domain layer for DDD patterns including: aggregates with version field,
  event emission, repository interfaces, and layer boundary violations.
  Use when: validating domain purity, checking optimistic locking, auditing events.
```

### 3. **Modelo Apropriado**
```markdown
# Análise rápida, read-only
model: sonnet  # Mais rápido e barato

# Análise complexa com decisões críticas
model: opus    # Mais inteligente

# Herdar modelo do agente principal
model: inherit
```

### 4. **Evitar Conflitos entre Agentes**
```markdown
# No CLAUDE.md ou no prompt do orchestrator:

Sub-agents might overwrite files from each other. To prevent:

1. Read-only agents: Only use Read, Grep, Glob (no Write/Edit)
2. Write agents: Always check git diff before writing
3. Coordination: Orchestrator assigns non-overlapping file sets
4. Communication: Agents report intent before modifying files
```

### 5. **Comandos com Argumentos**
```markdown
---
argument-hint: <domain|application|infrastructure> [quick|deep]
---

Analyze the **$1** layer with **${2:-quick}** analysis level.

# $1 = primeiro argumento (obrigatório)
# ${2:-quick} = segundo argumento com default "quick"
```

---

## 🔧 Comandos Úteis

```bash
# Listar agentes disponíveis
/agents

# Limpar contexto antes de nova tarefa
/clear

# Criar novo agente interativamente
/agents create

# Ver ajuda
/help

# Executar comando customizado
/analyze-all
/security:p0-check
/domain:coverage
```

---

## 📚 Recursos Adicionais

### Repositórios de Exemplo
- [pjt222/claude-code-agents](https://github.com/pjt222/claude-code-agents) - Coleção de agentes especializados
- [qdhenry/Claude-Command-Suite](https://github.com/qdhenry/Claude-Command-Suite) - 148+ comandos profissionais
- [wshobson/agents](https://github.com/wshobson/agents) - Orquestração multi-agente

### Docs Oficiais
- [docs.claude.com/claude-code/sub-agents](https://docs.claude.com/en/docs/claude-code/sub-agents)
- [docs.claude.com/claude-code/slash-commands](https://docs.claude.com/en/docs/claude-code/slash-commands)

### Artigos Técnicos
- [Multi-agent parallel coding](https://medium.com/@codecentrevibe/claude-code-multi-agent-parallel-coding-83271c4675fa)
- [Subagent deep dive](https://cuong.io/blog/2025/06/24-claude-code-subagent-deep-dive)

---

## 🎬 Próximos Passos

Quer que eu:

1. **Crie os agentes especializados** para o Ventros CRM? (domain-analyzer, security-auditor, etc.)
2. **Crie o comando `/analyze-all`** que orquestra tudo?
3. **Crie o `docs-manager`** que gerencia CLAUDE.md, TODO.md, AI_REPORT.md?
4. **Configure o sistema completo** com todos os agentes + comandos?
5. **Teste o sistema** rodando uma análise completa paralela?

---

**Last Updated**: 2025-10-15
**Maintainer**: Ventros CRM Team
**Version**: 1.0
