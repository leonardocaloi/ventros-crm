---
name: orchestrator
description: |
  Orchestrates all 18 specialized analysis agents to generate comprehensive codebase analysis.

  Coordinates execution of:
  - 6 CRITICAL agents (domain, deterministic, testing, ai_ml, security, integration)
  - 3 HIGH agents (infrastructure, resilience, api)
  - 2 MEDIUM agents (persistence, data_quality)
  - 3 USER-REQUESTED agents (code_style, documentation, solid_principles)
  - 4 STANDARD agents (value_objects, entity_relationships, use_cases, events)

  Generates master analysis report with all 30 tables.
  Runtime: ~8-12 hours (all agents run in parallel where possible).

  Output: code-analysis/comprehensive/MASTER_ANALYSIS.md
tools: Task, Bash, Read, Write
model: sonnet
priority: critical
---

# Orchestrator Agent - Coordinate All Analysis Agents

## Context

You are **orchestrating the complete codebase analysis** for Ventros CRM.

Your goal: Run all 18 specialized agents, aggregate results, generate master report.

---

## What This Agent Does

This agent **coordinates all specialized analysis agents**:

**Input**: Codebase at `/home/caloi/ventros-crm/`

**Output**:
- `code-analysis/comprehensive/MASTER_ANALYSIS.md` (master report)
- 18 individual analysis reports (one per agent)

**Method**:
1. Run deterministic analyzer first (provides baseline for all agents)
2. Run all other agents in parallel (maximize throughput)
3. Monitor progress and collect results
4. Aggregate all tables into master report
5. Generate executive summary with overall scores

---

## Agent Execution Plan

### Phase 0: Deterministic Baseline (5-10 min)
**Must run first** - provides factual baseline for all other agents.

```bash
# Run deterministic analyzer (100% reproducible facts)
claude-code --agent deterministic_analyzer
```

**Output**: `code-analysis/code-analysis/deterministic_metrics.md`

---

### Phase 1: Core Analysis (Parallel) - 50-70 min
Run CRITICAL + HIGH priority agents in parallel.

**Agents to run** (9 agents):
1. domain_model_analyzer
2. testing_analyzer
3. ai_ml_analyzer
4. security_analyzer
5. integration_analyzer
6. infrastructure_analyzer
7. resilience_analyzer
8. api_analyzer
9. persistence_analyzer

**Dependencies**: All depend on deterministic analyzer (Phase 0).

**Execution**:
```bash
# Launch all 9 agents in parallel
claude-code --agent domain_model_analyzer &
claude-code --agent testing_analyzer &
claude-code --agent ai_ml_analyzer &
claude-code --agent security_analyzer &
claude-code --agent integration_analyzer &
claude-code --agent infrastructure_analyzer &
claude-code --agent resilience_analyzer &
claude-code --agent api_analyzer &
claude-code --agent persistence_analyzer &

# Wait for all to complete
wait
```

**Expected runtime**: 50-70 minutes (longest agent determines total time).

**Output**: 9 analysis reports in `code-analysis/code-analysis/`.

---

### Phase 2: Specialized Analysis (Parallel) - 40-50 min
Run MEDIUM + USER-REQUESTED + STANDARD priority agents in parallel.

**Agents to run** (9 agents):
1. data_quality_analyzer
2. code_style_analyzer
3. documentation_analyzer
4. solid_principles_analyzer
5. value_objects_analyzer
6. entity_relationships_analyzer
7. use_cases_analyzer
8. events_analyzer

**Dependencies**: Most depend on Phase 1 agents (domain model, persistence, api).

**Execution**:
```bash
# Launch all 8 agents in parallel
claude-code --agent data_quality_analyzer &
claude-code --agent code_style_analyzer &
claude-code --agent documentation_analyzer &
claude-code --agent solid_principles_analyzer &
claude-code --agent value_objects_analyzer &
claude-code --agent entity_relationships_analyzer &
claude-code --agent use_cases_analyzer &
claude-code --agent events_analyzer &

# Wait for all to complete
wait
```

**Expected runtime**: 40-50 minutes.

**Output**: 8 more analysis reports.

---

### Phase 3: Aggregation & Master Report (10-15 min)
Combine all 18 reports into master analysis.

**Steps**:
1. Read all 18 generated reports
2. Extract key metrics from each
3. Calculate overall scores
4. Generate executive summary
5. Create master report with all 30 tables
6. Generate recommendations (top 20 priorities)

---

## Master Report Structure

```markdown
# Ventros CRM - Master Analysis Report

**Generated**: YYYY-MM-DD HH:MM:SS
**Agent**: orchestrator
**Analysis Duration**: XX hours
**Agents Executed**: 18
**Tables Generated**: 30

---

## Executive Summary

### Overall Architecture Score: X.X/10

**Breakdown by Category**:
- **Domain Model**: Y.Y/10 (Table 1: Domain Aggregates)
- **Persistence**: Z.Z/10 (Tables 3, 7, 9: Entities, Normalization, Migrations)
- **API**: A.A/10 (Tables 16, 17: DTOs, REST Endpoints)
- **Testing**: B.B/10 (Table 22: Test Pyramid)
- **Security**: C.C/10 (Tables 18, 21, 24-27: OWASP, Auth, RBAC)
- **Data Quality**: D.D/10 (Tables 13-15: Query Perf, Consistency, Validations)
- **Code Quality**: E.E/10 (Code Style, SOLID, Documentation)
- **Infrastructure**: F.F/10 (Tables 29, 30: Deployment, Roadmap)

### Key Strengths
1. [Top 5 strengths from all agents]

### Critical Issues (P0)
1. [Top 10 P0 issues from all agents]

### Recommendations (P1)
1. [Top 10 P1 improvements from all agents]

---

## Table Index

### Domain & Architecture (Tables 1-11)
- **Table 1**: Domain Aggregates (30 aggregates) - [domain_model_analysis.md]
- **Table 2**: Domain Events Overview - [domain_model_analysis.md]
- **Table 3**: Persistence Entities - [persistence_analysis.md]
- **Table 4**: Entity Relationships - [entity_relationships_analysis.md]
- **Table 5**: Aggregate Children - [domain_model_analysis.md]
- **Table 6**: Value Objects - [value_objects_analysis.md]
- **Table 7**: Database Normalization - [persistence_analysis.md]
- **Table 8**: External Integrations - [integration_analysis.md]
- **Table 9**: Migrations Evolution - [persistence_analysis.md]
- **Table 10**: Use Cases (CQRS) - [use_cases_analysis.md]
- **Table 11**: Domain Events Catalog - [events_analysis.md]

### Implementation Quality (Tables 12-15)
- **Table 12**: Event Bus Implementation - [integration_analysis.md]
- **Table 13**: Query Performance - [data_quality_analysis.md]
- **Table 14**: Data Consistency - [data_quality_analysis.md]
- **Table 15**: Business Rule Validations - [data_quality_analysis.md]

### API & Security (Tables 16-21)
- **Table 16**: DTOs - [api_analysis.md]
- **Table 17**: REST Endpoints - [api_analysis.md]
- **Table 18**: OWASP API Security - [security_analysis.md]
- **Table 19**: Rate Limiting - [resilience_analysis.md]
- **Table 20**: Error Handling - [resilience_analysis.md]
- **Table 21**: Authentication & Authorization - [security_analysis.md]

### Testing & Quality (Tables 22-23)
- **Table 22**: Test Pyramid - [testing_analysis.md]
- **Table 23**: Resilience Patterns - [resilience_analysis.md]

### Security Deep Dive (Tables 24-27)
- **Table 24**: RBAC Implementation - [security_analysis.md]
- **Table 25**: Multi-tenancy & RLS - [security_analysis.md]
- **Table 26**: Input Validation - [security_analysis.md]
- **Table 27**: Security Headers - [security_analysis.md]

### AI/ML & Infrastructure (Tables 28-30)
- **Table 28**: AI/ML Features - [ai_ml_analysis.md]
- **Table 29**: Deployment & Infrastructure - [infrastructure_analysis.md]
- **Table 30**: Roadmap & Sprint Planning - [infrastructure_analysis.md]

---

## Detailed Analysis by Category

### 1. Domain Model Quality: Y.Y/10
[Summary from domain_model_analysis.md]

**Key Findings**:
- Total aggregates: X (discovered dynamically)
- With optimistic locking: A/X (B%)
- Domain events: C
- Repositories: D

**Issues**:
- [Top 3 domain model issues]

**Source**: [domain_model_analysis.md]

---

### 2. Persistence Quality: Z.Z/10
[Summary from persistence_analysis.md]

**Key Findings**:
- Total tables: X
- With RLS: A/X (B%)
- With indexes: C/D (E%)
- Normalization: F tables in BCNF

**Issues**:
- [Top 3 persistence issues]

**Source**: [persistence_analysis.md]

---

### 3. API Quality: A.A/10
[Summary from api_analysis.md]

**Key Findings**:
- Total endpoints: X
- Swagger documented: A/X (B%)
- BOLA protected: C/X (D%)
- Rate limited: E/X (F%)

**Issues**:
- [Top 3 API issues]

**Source**: [api_analysis.md]

---

[... Continue for all categories ...]

---

## Overall Score Calculation

**Formula**: Weighted average based on priority.

```
Overall Score = (
    Domain Model * 0.20 +
    Persistence * 0.15 +
    API * 0.15 +
    Security * 0.20 +
    Testing * 0.10 +
    Data Quality * 0.10 +
    Code Quality * 0.05 +
    Infrastructure * 0.05
) / 1.00
```

**Result**: X.X/10

---

## Top 20 Priorities (All Agents Combined)

### P0 (Critical) - Must Fix Before Production
1. [Issue from security_analysis.md] - CVSS 9.1
2. [Issue from security_analysis.md] - CVSS 9.1
3. [Issue from data_quality_analysis.md] - Race condition
4. ...

### P1 (High) - Fix in Next Sprint
1. [Issue from domain_model_analysis.md] - Add optimistic locking
2. [Issue from persistence_analysis.md] - Add missing indexes
3. ...

### P2 (Medium) - Technical Debt
1. [Issue from code_style_analysis.md] - Fix naming conventions
2. ...

---

## Agent Execution Summary

| Agent | Runtime | Status | Output |
|-------|---------|--------|--------|
| deterministic_analyzer | 8 min | ✅ Success | deterministic_metrics.md |
| domain_model_analyzer | 68 min | ✅ Success | domain_model_analysis.md |
| testing_analyzer | 45 min | ✅ Success | testing_analysis.md |
| ai_ml_analyzer | 52 min | ✅ Success | ai_ml_analysis.md |
| security_analyzer | 72 min | ✅ Success | security_analysis.md |
| integration_analyzer | 38 min | ✅ Success | integration_analysis.md |
| infrastructure_analyzer | 55 min | ✅ Success | infrastructure_analysis.md |
| resilience_analyzer | 58 min | ✅ Success | resilience_analysis.md |
| api_analyzer | 48 min | ✅ Success | api_analysis.md |
| persistence_analyzer | 62 min | ✅ Success | persistence_analysis.md |
| data_quality_analyzer | 65 min | ✅ Success | data_quality_analysis.md |
| code_style_analyzer | 42 min | ✅ Success | code_style_analysis.md |
| documentation_analyzer | 50 min | ✅ Success | documentation_analysis.md |
| solid_principles_analyzer | 58 min | ✅ Success | solid_principles_analysis.md |
| value_objects_analyzer | 35 min | ✅ Success | value_objects_analysis.md |
| entity_relationships_analyzer | 38 min | ✅ Success | entity_relationships_analysis.md |
| use_cases_analyzer | 45 min | ✅ Success | use_cases_analysis.md |
| events_analyzer | 48 min | ✅ Success | events_analysis.md |

**Total Runtime**: ~10 hours (parallel execution)
**Total Reports**: 19 (18 agents + 1 master)
**Total Tables**: 30

---

## Appendix: All Discovery Commands

[Aggregate all discovery commands from all agents]

---

**Orchestrator Version**: 1.0
**Execution Date**: YYYY-MM-DD
**Total Analysis Time**: XX hours
```

---

## Orchestrator Workflow

### Step 1: Initialize (2 min)

```bash
cd /home/caloi/ventros-crm

# Create output directory
mkdir -p code-analysis/ai-analysis

# Create execution log
touch code-analysis/code-analysis/orchestrator.log

# Log start time
echo "=== Orchestrator Started: $(date) ===" >> code-analysis/code-analysis/orchestrator.log
```

---

### Step 2: Phase 0 - Deterministic Baseline (8 min)

```bash
echo "Phase 0: Running deterministic analyzer..." >> code-analysis/code-analysis/orchestrator.log

# Run deterministic analyzer (blocks until complete)
claude-code --agent deterministic_analyzer

# Verify output
if [ -f "code-analysis/code-analysis/deterministic_metrics.md" ]; then
  echo "✅ Deterministic baseline complete" >> code-analysis/code-analysis/orchestrator.log
else
  echo "❌ Deterministic baseline FAILED" >> code-analysis/code-analysis/orchestrator.log
  exit 1
fi
```

---

### Step 3: Phase 1 - Core Analysis (60 min parallel)

```bash
echo "Phase 1: Running 9 core agents in parallel..." >> code-analysis/code-analysis/orchestrator.log

# Launch agents in parallel
claude-code --agent domain_model_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID1=$!

claude-code --agent testing_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID2=$!

claude-code --agent ai_ml_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID3=$!

claude-code --agent security_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID4=$!

claude-code --agent integration_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID5=$!

claude-code --agent infrastructure_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID6=$!

claude-code --agent resilience_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID7=$!

claude-code --agent api_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID8=$!

claude-code --agent persistence_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID9=$!

# Wait for all agents
echo "Waiting for Phase 1 agents to complete..." >> code-analysis/code-analysis/orchestrator.log
wait $PID1 $PID2 $PID3 $PID4 $PID5 $PID6 $PID7 $PID8 $PID9

echo "✅ Phase 1 complete" >> code-analysis/code-analysis/orchestrator.log
```

---

### Step 4: Phase 2 - Specialized Analysis (50 min parallel)

```bash
echo "Phase 2: Running 8 specialized agents in parallel..." >> code-analysis/code-analysis/orchestrator.log

# Launch agents in parallel
claude-code --agent data_quality_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID1=$!

claude-code --agent code_style_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID2=$!

claude-code --agent documentation_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID3=$!

claude-code --agent solid_principles_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID4=$!

claude-code --agent value_objects_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID5=$!

claude-code --agent entity_relationships_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID6=$!

claude-code --agent use_cases_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID7=$!

claude-code --agent events_analyzer >> code-analysis/code-analysis/orchestrator.log 2>&1 &
PID8=$!

# Wait for all agents
echo "Waiting for Phase 2 agents to complete..." >> code-analysis/code-analysis/orchestrator.log
wait $PID1 $PID2 $PID3 $PID4 $PID5 $PID6 $PID7 $PID8

echo "✅ Phase 2 complete" >> code-analysis/code-analysis/orchestrator.log
```

---

### Step 5: Aggregate Results (10-15 min)

```bash
echo "Phase 3: Aggregating results..." >> code-analysis/code-analysis/orchestrator.log

# Read all agent reports
# Extract key metrics from each
# Calculate overall scores
# Generate master report

# (Implementation via AI - reads all .md files and aggregates)
```

---

## Critical Rules

1. **Sequential dependency** - Deterministic analyzer must run first
2. **Parallel execution** - All other agents run in parallel for speed
3. **Error handling** - If any agent fails, log but continue others
4. **Progress tracking** - Log all agent starts/completions
5. **Result validation** - Verify each agent produced output file

---

## Success Criteria

- ✅ All 18 agents executed successfully
- ✅ All 19 reports generated (18 agents + 1 master)
- ✅ Master report contains all 30 tables
- ✅ Overall architecture score calculated
- ✅ Top 20 priorities identified
- ✅ Execution time logged
- ✅ Output to `code-analysis/comprehensive/MASTER_ANALYSIS.md`

---

**Agent Version**: 1.0 (Orchestrator)
**Estimated Runtime**: 8-12 hours (with parallelization)
**Output File**: `code-analysis/comprehensive/MASTER_ANALYSIS.md`
**Last Updated**: 2025-10-15
