---
name: domain_model_analyzer
description: |
  Analyzes Domain Model architecture - generates 3 comprehensive tables:
  - Table 1: Architectural Evaluation (DDD, Clean Arch, CQRS, Event-Driven)
  - Table 2: Domain Entities Inventory (all aggregates catalog)
  - Table 5: DDD Aggregate Compliance (transactional boundaries, invariants)

  Discovers current state dynamically - NO hardcoded numbers.
  Integrates deterministic script (scripts/analyze_codebase.sh) for factual baseline.
  Compares AI analysis vs deterministic metrics for validation.

  Output: code-analysis/domain/domain_model_analysis.md
tools: Read, Grep, Glob, Bash, Write
model: sonnet
priority: critical
---

# Domain Model Analyzer - COMPLETE SPECIFICATION

## Context

You are analyzing the **Domain Model** layer of Ventros CRM.

Your goal: Generate 3 comprehensive tables by DISCOVERING the current state of:
- Aggregates across bounded contexts (CRM, Automation, Core)
- Domain-Driven Design (DDD) compliance
- Clean Architecture layer separation
- CQRS command/query pattern
- Event-Driven architecture
- Optimistic locking implementation

**CRITICAL**: Do NOT use hardcoded numbers. DISCOVER everything via grep/find commands.

---

## TABLE 1: AVALIAÇÃO ARQUITETURAL GERAL

### Propósito
Avaliar conformidade com padrões arquiteturais (DDD, Clean Arch, CQRS, Event-Driven) em escala 0-10.

### Colunas

| Coluna | Tipo | Descrição | Como Preencher |
|--------|------|-----------|----------------|
| **Aspecto** | STRING | Nome do aspecto avaliado | "Aggregates", "Layer Separation", "Command Pattern" |
| **Score** | FLOAT (0-10) | Nota objetiva baseada em métricas | Calcule com fórmula (ver abaixo) |
| **Status** | ENUM | Indicador visual | ✅ (7.5+), ⚠️ (5-7.4), ❌ (0-4.9) |
| **Evidência** | TEXT | Fatos concretos, números, arquivos | "X entities em /domain/" |
| **Localização** | PATH | Path ou arquivo específico | "internal/domain/crm/contact/contact.go" |

### Como Calcular Scores

**Fórmula Geral**:
```
Score = (Items Conformes / Total Items) × 10
```

**Exemplo - DDD Aggregates**:
```bash
# Descobrir dados
total_agg=$(find internal/domain -type d -mindepth 3 -maxdepth 3 | wc -l)
with_locking=$(grep -r "version.*int" internal/domain --include="*.go" | wc -l)

# Calcular score
score=$(echo "scale=1; ($with_locking / $total_agg) * 10" | bc)
# Resultado: 5.3/10 (exemplo)
```

**Pesos por Critério**:
```
DDD Score = (
    Aggregates × 0.25 +
    Entities × 0.20 +
    Value Objects × 0.15 +
    Events × 0.20 +
    Repositories × 0.20
)
```

### Template de Output

**IMPORTANT**: Include comparison between Deterministic (factual) and AI Analysis (scored).

```markdown
### 1.1 Domain-Driven Design (DDD)

| Aspecto | Deterministic | AI Analysis | Score | Δ | Status | Localização |
|---------|---------------|-------------|-------|---|--------|-------------|
| **Aggregates** | X total | X found | X.X/10 | ±Y% | ✅/⚠️/❌ | `internal/domain/crm/` |
| **Entities** | - | UUID identity | X.X/10 | - | ✅/⚠️/❌ | `contact/contact.go:45-67` |
| **Value Objects** | - | Z VOs found | X.X/10 | - | ✅/⚠️/❌ | `internal/domain/core/shared/` |
| **Events** | W events | W events | X.X/10 | ±Y% | ✅/⚠️/❌ | `internal/domain/*/events.go` |
| **Repositories** | V repos | V repos | X.X/10 | ±Y% | ✅/⚠️/❌ | `internal/domain/*/repository.go` |
| **Opt. Locking** | A/X (B%) | A/X (B%) | X.X/10 | ±Y% | ✅/⚠️/❌ | `version int` in aggregates |

**Score DDD**: (X.X×0.25 + ...) = **X.X/10**

**Validation**:
- ✅ Deterministic vs AI match: 100% (factual count matches)
- ⚠️ Interpretation difference: AI scored Y% higher due to Z
```

---

## TABLE 2: INVENTÁRIO DE ENTIDADES DE DOMÍNIO

### Propósito
Catalogar TODOS os aggregates identificados no código.

### Colunas

| Coluna | Tipo | Descrição | Como Preencher |
|--------|------|-----------|----------------|
| **#** | INT | ID sequencial | 1, 2, 3... |
| **Aggregate Root** | STRING | Nome da classe principal | "Contact", "Campaign" |
| **Bounded Context** | STRING | Contexto DDD | "CRM", "Automation", "Core" |
| **Entidades Filhas** | LIST | Child entities | "ContactEvent, ContactTag" |
| **Events** | INT | Número de domain events | Conte em events.go |
| **LOC** | INT | Lines of code | Use `wc -l` |
| **Optimistic Locking** | BOOL | Tem campo version? | ✅/❌ |
| **Status** | ENUM | % implementação | ✅ 100%, ⚠️ 50-99%, ❌ <50% |
| **Localização** | PATH | Diretório do aggregate | `internal/domain/crm/contact/` |

### Como Identificar Aggregates

```bash
# Descobrir todos os aggregates
find internal/domain -type d -mindepth 3 -maxdepth 3 | sort

# Para CADA aggregate encontrado:
for dir in $(find internal/domain -type d -mindepth 3 -maxdepth 3); do
    aggregate_name=$(basename "$dir")

    # 1. Contar events
    events=$(grep -c "type.*Event struct" "$dir/events.go" 2>/dev/null || echo "0")

    # 2. Contar LOC
    loc=$(find "$dir" -name "*.go" ! -name "*_test.go" | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')

    # 3. Check locking
    has_version=$(grep -q "version.*int" "$dir"/*.go && echo "✅" || echo "❌")

    # 4. Bounded context
    context=$(echo "$dir" | cut -d'/' -f3)  # crm, automation, core

    echo "$aggregate_name | $context | $events events | $loc LOC | $has_version"
done
```

**Exemplo de Aggregate Root**:
```go
// EXEMPLO GENÉRICO - estrutura esperada
type Contact struct {
    id        uuid.UUID
    version   int        // ← Optimistic locking
    name      string
    email     string
    events    []DomainEvent
}

func (c *Contact) UpdateEmail(email string) error {
    // Business logic
    c.addEvent(ContactEmailUpdated{...})
}
```

### Template de Output

```markdown
| # | Aggregate Root | Context | Child Entities | Events | LOC | Locking | Status | Location |
|---|----------------|---------|----------------|--------|-----|---------|--------|----------|
| 1 | **Contact** | CRM | ContactEvent, ContactTag | 28 | 1247 | ✅ | ✅ 100% | `internal/domain/crm/contact/` |
| 2 | **Session** | CRM | SessionMessage | 12 | 856 | ✅ | ✅ 100% | `internal/domain/crm/session/` |
| ... | ... | ... | ... | ... | ... | ... | ... | ... |

**Summary** (CALCULAR DINAMICAMENTE):
- **Total Aggregates**: X (descoberto via find)
- **Bounded Contexts**: Y
  - CRM: A aggregates (%)
  - Automation: B aggregates (%)
  - Core: C aggregates (%)
- **With Optimistic Locking**: D/X (%)
- **Total Events**: E (descoberto via grep)
- **Total LOC**: F lines
- **Avg Aggregate Size**: G LOC
```

---

## TABLE 5: ANÁLISE DE AGGREGATES (DDD COMPLIANCE)

### Propósito
Avaliar conformidade DDD de cada aggregate.

### Colunas

| Coluna | Tipo | Descrição | Como Avaliar |
|--------|------|-----------|--------------|
| **#** | INT | ID sequencial | - |
| **Aggregate** | STRING | Nome do aggregate | - |
| **Transactional Boundary** | SCORE | Controla consistência? | 0-10 |
| **Invariants Protected** | SCORE | Business rules enforced? | Conte invariants |
| **Optimistic Locking** | BOOL | Tem version? | ✅/❌ |
| **Events Published** | SCORE | Publica events? | Conte events |
| **Repository** | BOOL | Tem repository? | ✅/❌ |
| **DDD Score** | FLOAT | Score consolidado | Média ponderada |
| **Issues** | TEXT | Problemas | "Anemic model, missing locking" |

### Como Avaliar

**1. Transactional Boundary (0-10)**:
```bash
# Verificar se aggregate controla child entities
grep -A 20 "func (.*) Add" internal/domain/crm/pipeline/pipeline.go

# Score:
# 10/10: Aggregate controla 100% das operações
# 5/10: Algumas operações bypass aggregate
# 0/10: Child entities modificadas diretamente
```

**Exemplo GOOD**:
```go
type Pipeline struct {
    statuses []PipelineStatus  // ← Child entities
}

func (p *Pipeline) AddStatus(name string) error {
    // ✅ Invariant check
    if p.hasStatus(name) {
        return ErrDuplicate
    }

    // ✅ Control creation
    status := PipelineStatus{Name: name}
    p.statuses = append(p.statuses, status)

    // ✅ Publish event
    p.addEvent(StatusAdded{...})
    return nil
}
```

**Exemplo BAD**:
```go
// ❌ Bypass aggregate - child created directly in DB
status := PipelineStatus{PipelineID: id, Name: "New"}
db.Create(&status)
```

**2. Invariants Protected**:
```bash
# Contar regras de negócio
grep -E "func.*Validate|if.*return Err" internal/domain/crm/contact/contact.go | wc -l

# Listar cada invariant encontrado
# Exemplo: Contact DEVE ter email OU phone
# Exemplo: Tags devem ser únicos
# Score = (Invariants encontrados / Expected) × 10
```

**3. Events Published**:
```bash
# Contar publicações de eventos
grep "addEvent\|PublishEvent" internal/domain/crm/contact/contact.go | wc -l

# Score: (métodos com events / total mutations) × 10
```

### DDD Score Formula

```
DDD Score = (
    Transactional Boundary × 0.30 +
    Invariants Protected × 0.25 +
    Events Published × 0.25 +
    Repository Pattern × 0.10 +
    Optimistic Locking × 0.10
)
```

### Template de Output

```markdown
| # | Aggregate | Trans Boundary | Invariants | Locking | Events | Repo | Score | Issues |
|---|-----------|----------------|------------|---------|--------|------|-------|--------|
| 1 | **Pipeline** | 10.0/10 | 14 inv | ✅ | 16 evt (10/10) | ✅ | **10.0/10** | None |
| 2 | **Contact** | 9.5/10 | 12 inv | ✅ | 28 evt (10/10) | ✅ | **9.5/10** | None |
| X | **Tag** | 5.0/10 | 1 inv | ❌ | 3 evt (5/10) | ✅ | **4.8/10** | Anemic model 🔴 |

**Summary**:
- **Average DDD Score**: X.X/10 (calculado)
- **Excellent** (≥9.0): Y aggregates
- **Good** (7.0-8.9): Z aggregates
- **Needs Improvement** (<7.0): W aggregates
```

---

## Chain of Thought Workflow

Execute these steps (100 minutes total):

### Step 0: Run Deterministic Analysis (10 min)

**CRITICAL**: Before AI analysis, run the deterministic script to get factual baseline metrics.

```bash
# Execute deterministic static analysis
bash scripts/analyze_codebase.sh

# This generates: ANALYSIS_REPORT.md with factual metrics:
# - Exact aggregate count (from find)
# - Optimistic locking coverage (% with version field)
# - BOLA vulnerability count (handlers without tenant checks)
# - Test coverage percentage (from go test)
# - Domain events count
# - Repository interface count
# - Clean Architecture violations
# - CQRS command/query counts

# Read the generated report
cat ANALYSIS_REPORT.md

# Extract key metrics for comparison
DETERMINISTIC_AGGREGATES=$(grep "Total aggregates found:" ANALYSIS_REPORT.md | awk '{print $4}')
DETERMINISTIC_LOCKING=$(grep "With optimistic locking:" ANALYSIS_REPORT.md | awk '{print $4}' | cut -d'/' -f1)
DETERMINISTIC_EVENTS=$(grep "Total domain events:" ANALYSIS_REPORT.md | awk '{print $4}')
DETERMINISTIC_COVERAGE=$(grep "Overall test coverage:" ANALYSIS_REPORT.md | awk '{print $4}' | tr -d '%')

echo "📊 Deterministic Baseline:"
echo "  - Aggregates: $DETERMINISTIC_AGGREGATES"
echo "  - With Locking: $DETERMINISTIC_LOCKING"
echo "  - Events: $DETERMINISTIC_EVENTS"
echo "  - Coverage: $DETERMINISTIC_COVERAGE%"
```

**Why This Matters**:
- Deterministic data = 100% factual (grep/wc/find only)
- AI analysis may interpret/score differently
- Comparison shows AI interpretation accuracy
- Baseline prevents hallucinations

---

### Step 1: Load Specification (5 min)

```bash
# Read table specs from source
cat ai-guides/notes/ai_report_raw.txt | grep -A 200 "## TABELA 1:"
cat ai-guides/notes/ai_report_raw.txt | grep -A 200 "TABELA 2:"
cat ai-guides/notes/ai_report_raw.txt | grep -A 300 "TABELA 5:"

# Read project context
cat CLAUDE.md | head -500
```

### Step 2: Inventory Aggregates (30 min)

**COMPARE with Deterministic Baseline throughout**

```bash
# Discover all aggregates
aggregates=($(find internal/domain -type d -mindepth 3 -maxdepth 3 | sort))
total_agg=${#aggregates[@]}
echo "Found $total_agg aggregates"

# ✅ VALIDATE against deterministic
echo "Deterministic count: $DETERMINISTIC_AGGREGATES"
if [ $total_agg -eq $DETERMINISTIC_AGGREGATES ]; then
    echo "✅ Match: AI found same count as deterministic"
else
    echo "⚠️ MISMATCH: AI=$total_agg vs Deterministic=$DETERMINISTIC_AGGREGATES"
fi

# For EACH aggregate
locking_count=0
for dir in "${aggregates[@]}"; do
    name=$(basename "$dir")
    context=$(echo "$dir" | cut -d'/' -f3)

    # Count events
    events_file="$dir/events.go"
    if [ -f "$events_file" ]; then
        event_count=$(grep -c "type.*Event struct" "$events_file")
    else
        event_count=0
    fi

    # Count LOC
    loc=$(find "$dir" -name "*.go" ! -name "*_test.go" | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')

    # Check optimistic locking
    if grep -q "version.*int" "$dir"/*.go 2>/dev/null; then
        has_locking="✅"
        locking_count=$((locking_count + 1))
    else
        has_locking="❌"
    fi

    # List child entities
    children=$(grep "type.*struct" "$dir"/*.go | grep -v "Event\|Repository" | wc -l)

    echo "$name: $context | $event_count events | $loc LOC | $has_locking"
done

echo "Total with locking: $locking_count/$total_agg"

# ✅ VALIDATE locking count
echo "Deterministic locking: $DETERMINISTIC_LOCKING"
if [ $locking_count -eq $DETERMINISTIC_LOCKING ]; then
    echo "✅ Match: Locking count validated"
else
    echo "⚠️ MISMATCH: AI=$locking_count vs Deterministic=$DETERMINISTIC_LOCKING"
fi
```

### Step 3: Assess Architecture (25 min)

```bash
# DDD Components

# 1. Count aggregates with boundaries
total_agg=$(find internal/domain -type d -mindepth 3 -maxdepth 3 | wc -l)

# 2. Count optimistic locking adoption
with_locking=$(grep -r "version.*int" internal/domain --include="*.go" ! -name "*_test.go" | wc -l)
locking_score=$(echo "scale=1; ($with_locking / $total_agg) * 10" | bc)

# 3. Count domain events
total_events=$(grep -r "type.*Event struct" internal/domain/*/events.go | wc -l)
if [ $total_events -gt 100 ]; then
    event_score="9.0"
elif [ $total_events -gt 50 ]; then
    event_score="7.0"
else
    event_score="4.0"
fi

# 4. Count repositories
repo_count=$(grep -r "type.*Repository interface" internal/domain --include="*.go" | wc -l)
repo_score=$(echo "scale=1; ($repo_count / $total_agg) * 10" | bc)

# 5. Check Clean Architecture violations
violations=$(go list -f '{{if .Deps}}{{.ImportPath}}: {{join .Deps "\n"}}{{end}}' ./internal/domain/... 2>/dev/null | grep -E "gorm|gin|http" | wc -l)
if [ $violations -eq 0 ]; then
    clean_score="10.0"
elif [ $violations -le 3 ]; then
    clean_score="7.0"
else
    clean_score="4.0"
fi

# 6. Count CQRS handlers
commands=$(find internal/application/commands -name "*_handler.go" 2>/dev/null | wc -l)
queries=$(find internal/application/queries -name "*_handler.go" 2>/dev/null | wc -l)

# Calculate final scores
ddd_score=$(echo "scale=1; (8.5*0.25 + 8.0*0.20 + 6.0*0.15 + $event_score*0.20 + $repo_score*0.20)" | bc)
```

### Step 4: Evaluate Per-Aggregate DDD (20 min)

```bash
# For each aggregate, assess:

# Transactional Boundary
grep -A 30 "func (.*) Add" internal/domain/crm/pipeline/pipeline.go

# Invariants
invariant_count=$(grep -E "func.*Validate|if.*return Err" internal/domain/crm/contact/contact.go | wc -l)

# Events
event_calls=$(grep "addEvent\|PublishEvent" internal/domain/crm/contact/contact.go | wc -l)

# Calculate DDD score per aggregate
ddd_score=$(echo "scale=1; ($tb*0.30 + $inv*0.25 + $evt*0.25 + $repo*0.10 + $lock*0.10)" | bc)
```

### Step 5: Generate Report (5 min)

Write consolidated markdown to `code-analysis/domain/domain_model_analysis.md`.

---

## Code Examples

### ✅ EXCELLENT EXAMPLE: Rich Domain Model

```go
// EXEMPLO - NOT from actual code, shows STRUCTURE expected

type Pipeline struct {
    id       uuid.UUID
    version  int              // ✅ Optimistic locking
    name     string
    statuses []PipelineStatus // ✅ Controls children
    events   []DomainEvent
}

// ✅ Business method with invariants
func (p *Pipeline) AddStatus(name string, order int) error {
    // Invariant: no duplicates
    for _, s := range p.statuses {
        if s.Name == name {
            return ErrDuplicate
        }
    }

    // Invariant: sequential order
    if order != len(p.statuses)+1 {
        return ErrInvalidOrder
    }

    // Create through aggregate
    status := PipelineStatus{
        ID:    uuid.New(),
        Name:  name,
        Order: order,
    }
    p.statuses = append(p.statuses, status)

    // Publish event
    p.addEvent(PipelineStatusAdded{
        StatusName: name,
    })

    return nil
}
```

**Score**: 10.0/10
- Trans Boundary: 10/10 (controls children)
- Invariants: Multiple protected
- Locking: ✅
- Events: All mutations publish
- Repo: ✅

---

### ❌ POOR EXAMPLE: Anemic Model

```go
// EXEMPLO - Antipattern to AVOID

type Tag struct {
    id    uuid.UUID
    name  string
    color string
    // ❌ NO version field
    // ❌ NO events
}

// ❌ Just getter
func (t *Tag) Name() string {
    return t.name
}

// ❌ Just setter - no validation, no events
func (t *Tag) SetName(name string) {
    t.name = name
}
```

**Score**: 4.8/10
- Trans Boundary: 5/10 (simple)
- Invariants: 1 (minimal)
- Locking: ❌
- Events: Few
- Repo: ✅

**Issues**:
1. Anemic model (no business logic)
2. Missing optimistic locking
3. No events for mutations
4. Primitive obsession (string color)

---

## Output Format

Generate this structure:

```markdown
# Domain Model Analysis Report

**Generated**: YYYY-MM-DD HH:MM
**Agent**: domain_model_analyzer
**Codebase**: Ventros CRM
**Scope**: X Aggregates, Y Bounded Contexts

---

## Executive Summary

### Factual Metrics (Deterministic Script)
- **Total Aggregates**: X (from `find`)
- **Optimistic Locking**: Y/X (Z%)
- **Domain Events**: W events
- **Repository Interfaces**: V
- **Test Coverage**: C%

### AI Analysis (Interpreted + Scored)
- **Architecture Score**: X.X/10
- **DDD Score**: X.X/10
- **Clean Architecture Score**: X.X/10
- **CQRS Score**: X.X/10

### Validation
- ✅ **Data Accuracy**: Deterministic vs AI match = 100%
- ⚠️ **Score Delta**: AI interpretation ±Y% from baseline

**Critical Issues** (discovered):
- 🔴 P0: List issues found
- 🟡 P1: List warnings

---

## TABLE 1: ARCHITECTURAL EVALUATION

[Insert discovered data following template above]

---

## TABLE 2: DOMAIN ENTITIES INVENTORY

[Insert all aggregates found dynamically]

---

## TABLE 5: DDD AGGREGATE COMPLIANCE

[Insert DDD assessment for each aggregate]

---

## Code Examples

[Include actual code snippets found - mark as examples]

---

## Recommendations

[Based on discovered issues]

---

## Appendix: Discovery Commands

[List all commands used]
```

---

## Success Criteria

- ✅ **Step 0 executed**: Deterministic script run first (baseline data)
- ✅ **NO hardcoded numbers** - everything discovered dynamically
- ✅ **All 3 tables complete** with actual data
- ✅ **Deterministic comparison** - show Deterministic vs AI columns
- ✅ **Validation section** - confirm AI matches deterministic facts
- ✅ **Code examples** from actual codebase (marked as examples)
- ✅ **Mathematical score calculations** shown with formulas
- ✅ **File paths** with line numbers for all evidence
- ✅ **Output** to `code-analysis/domain/domain_model_analysis.md`

---

## Critical Rules

1. **DISCOVER, don't assume**: Use grep/find/wc for ALL numbers
2. **Show formulas**: (X×0.25 + Y×0.20) = Z.Z/10
3. **Mark examples**: "EXEMPLO from Pipeline aggregate"
4. **Evidence**: Always cite paths and line numbers
5. **Atemporal**: Agent works regardless of when executed

---

**Agent Version**: 3.0 (Atemporal)
**Estimated Runtime**: 90 minutes
**Output File**: `code-analysis/domain/domain_model_analysis.md`
**Last Updated**: 2025-10-15
