---
name: analyze-complete
description: Run ALL 19 analysis agents via meta_orchestrator - generates complete 30-table master report (2-3 hours)
---

# Complete Codebase Analysis - ALL 19 Agents

**Purpose**: Run comprehensive analysis using ALL 19 specialized agents and generate master report with all 30 tables.

**Duration**: 2-3 hours (parallel execution - actual may be longer, see below)
**Agents**: 19/19 (100% coverage)
**Output**: `code-analysis/comprehensive/MASTER_ANALYSIS.md`

---

## 🎯 What This Does

Invokes `meta_orchestrator` to run complete analysis in 4 phases:

### Phase 0: Deterministic Baseline (5-10 min)
- Runs `global_deterministic_analyzer` ⭐
- Generates 100% factual baseline (grep/wc/find only)
- Validates AI analysis accuracy

### Phase 1: Core Analysis (50-70 min, parallel)
Runs 9 CRITICAL + HIGH priority agents:
1. `crm_domain_model_analyzer` - 30 aggregates, DDD compliance
2. `crm_testing_analyzer` - Coverage, test pyramid
3. `crm_ai_ml_analyzer` - 12 AI providers, enrichment status
4. `crm_security_analyzer` - OWASP Top 10, P0 vulnerabilities
5. `crm_integration_analyzer` - WAHA, Stripe, Meta Ads
6. `crm_infrastructure_analyzer` - Docker, K8s, CI/CD
7. `crm_resilience_analyzer` - Circuit breaker, retry, timeouts
8. `crm_api_analyzer` - 158 endpoints, Swagger, DTOs
9. `crm_persistence_analyzer` - Entities, migrations, RLS

### Phase 2: Specialized Analysis (40-50 min, parallel)
Runs 8 MEDIUM + STANDARD priority agents:
10. `crm_data_quality_analyzer` - N+1, query perf, validations
11. `global_code_style_analyzer` - Go conventions
12. `global_documentation_analyzer` - Swagger, godoc
13. `global_solid_principles_analyzer` - SOLID violations
14. `crm_value_objects_analyzer` - VOs, primitive obsession
15. `crm_entity_relationships_analyzer` - Entity graph
16. `crm_use_cases_analyzer` - 80+ CQRS commands/queries
17. `crm_events_analyzer` - 182 events, Outbox Pattern
18. *(1 reserved slot)*

### Phase 3: Aggregation (10-15 min)
- Reads all 19 reports
- Generates MASTER_ANALYSIS.md with 30 tables
- Calculates overall architecture score
- Consolidates top 20 priorities (P0, P1, P2)

---

## 🚀 Usage

### Basic
```bash
/analyze-complete
```

### With Options
```bash
# Export as HTML + update docs
/analyze-complete --export=html --update-readme

# Full integration: analysis + TODO update + GitHub issues
/analyze-complete --update-todo --create-issues

# Verbose mode for debugging
/analyze-complete --verbose

# Just run, no updates
/analyze-complete --quiet
```

---

## 🎛️ Available Parameters

### Export Options
- `--export=html` - Generate HTML report (requires browser)
- `--export=json` - Generate JSON data for parsing
- `--export=pdf` - Generate PDF (requires wkhtmltopdf)
- *(Default: markdown)*

### Documentation Updates
- `--update-readme` - Update README.md with architecture stats
- `--update-devguide` - Update DEV_GUIDE.md with patterns
- `--update-todo` - Run `mgmt_todo_manager` agent after analysis ⭐ **Recomendado**
  - Consolidates findings from all 19 agents
  - Marks completed tasks with ✅ (verified via grep/find)
  - Adds new P0 vulnerabilities from security_analysis.md
  - Updates coverage gaps from testing_analysis.md
  - Re-prioritizes based on architecture scores

### Integration Options
- `--create-issues` - Create GitHub issues for P0 vulnerabilities
- `--update-p0` - Add critical findings to P0_ACTIVE_WORK.md
- `--update-agent-state` - Save findings to AGENT_STATE.json

### Execution Control
- `--verbose` - Show detailed logs (agents, commands, progress)
- `--quiet` - Only show summary (hide agent outputs)
- `--sequential` - Run agents sequentially (slower, 15-20h, for debugging)
- *(Default: parallel)*

---

## 📊 What You'll See

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📚 COMPLETE CODEBASE ANALYSIS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Mode: comprehensive (ALL 19 agents)
Estimated time: 2-3 hours
Output: code-analysis/comprehensive/MASTER_ANALYSIS.md

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 Phase 0: Deterministic Baseline (5-10 min)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏳ Running global_deterministic_analyzer...

✅ Baseline Complete (8 min)
   📊 Factual Metrics (100% reproducible):
   - Total aggregates: 30
   - Optimistic locking: 16/30 (53%)
   - Domain events: 182
   - Test coverage: 82%
   - BOLA vulnerable endpoints: 60
   - Security: 5 P0 issues

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔄 Phase 1: Core Analysis (50-70 min, parallel)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Running 9 core agents in parallel...

├─ [1/9] crm_domain_model_analyzer........... ⏳
├─ [2/9] crm_testing_analyzer................ ⏳
├─ [3/9] crm_ai_ml_analyzer.................. ⏳
├─ [4/9] crm_security_analyzer............... ⏳
├─ [5/9] crm_integration_analyzer............ ⏳
├─ [6/9] crm_infrastructure_analyzer......... ⏳
├─ [7/9] crm_resilience_analyzer............. ⏳
├─ [8/9] crm_api_analyzer.................... ⏳
└─ [9/9] crm_persistence_analyzer............ ⏳

[Waiting for longest agent to complete...]

✅ [1/9] Domain Model Analysis (68 min)
   - 30 aggregates analyzed
   - DDD Score: 8.5/10
   - 14 aggregates missing optimistic locking
   - 182 domain events catalogued

✅ [2/9] Testing Analysis (45 min)
   - Overall coverage: 82%
   - 14 aggregates < 80% coverage
   - 23 missing integration tests
   - Test pyramid ratio: 70:20:10 ✅

✅ [3/9] AI/ML Analysis (52 min)
   - 12 providers configured
   - Message enrichment: 100% complete ✅
   - Memory service: 20% complete ⚠️
   - Vector search: 0% (not started) ❌

✅ [4/9] Security Analysis (72 min) 🚨
   🔴 CRITICAL: 5 P0 Vulnerabilities Found!
   1. Dev mode bypass (CVSS 9.1)
   2. SSRF in webhooks (CVSS 9.1)
   3. BOLA in 60 endpoints (CVSS 8.2)
   4. Resource exhaustion (CVSS 7.5)
   5. Missing RBAC in 95 endpoints (CVSS 7.1)

✅ [5/9] Integration Analysis (38 min)
   - 3 external services (WAHA, Stripe, Meta Ads)
   - 12 API clients
   - 5 webhooks configured
   - Circuit breaker: Partially implemented

✅ [6/9] Infrastructure Analysis (55 min)
   - Docker: ✅ Configured
   - Kubernetes: ✅ Helm charts ready
   - CI/CD: GitHub Actions + AWX + Terraform

✅ [7/9] Resilience Analysis (58 min)
   - Circuit breaker: 3 implementations
   - Retry logic: 12 locations
   - Timeouts: Mostly configured
   - Rate limiting: Partially implemented

✅ [8/9] API Analysis (48 min)
   - 158 endpoints total
   - 23 missing Swagger docs (15%)
   - 60 missing BOLA checks (38%)
   - DTOs: Consistent usage ✅

✅ [9/9] Persistence Analysis (62 min)
   - 30 entities
   - 30 repositories
   - 45 migrations
   - RLS policies: 80% coverage

Phase 1 Complete: 72 minutes (limited by slowest agent)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔄 Phase 2: Specialized Analysis (40-50 min, parallel)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Running 8 specialized agents in parallel...

├─ [1/8] crm_data_quality_analyzer........... ⏳
├─ [2/8] global_code_style_analyzer.......... ⏳
├─ [3/8] global_documentation_analyzer....... ⏳
├─ [4/8] global_solid_principles_analyzer.... ⏳
├─ [5/8] crm_value_objects_analyzer.......... ⏳
├─ [6/8] crm_entity_relationships_analyzer... ⏳
├─ [7/8] crm_use_cases_analyzer.............. ⏳
└─ [8/8] crm_events_analyzer................. ⏳

[Waiting...]

✅ [1/8] Data Quality (65 min)
   - N+1 queries: 12 found
   - Query performance: 5 slow queries (>500ms)
   - Validation gaps: 8 aggregates

✅ [2/8] Code Style (42 min)
   - Go conventions: 92% compliance
   - Naming violations: 12
   - Unused imports: 5

✅ [3/8] Documentation (50 min)
   - Swagger coverage: 85%
   - Godoc coverage: 78%
   - Missing API docs: 23 endpoints

✅ [4/8] SOLID Principles (58 min)
   - SRP violations: 3 handlers
   - DIP violations: 2 domain imports
   - Overall: 8.5/10

✅ [5/8] Value Objects (35 min)
   - 45 value objects found
   - Primitive obsession: 8 cases
   - Immutability: 100% ✅

✅ [6/8] Entity Relationships (38 min)
   - Entity graph generated
   - Foreign keys: 87 relationships
   - Orphaned entities: 0 ✅

✅ [7/8] Use Cases (45 min)
   - 80+ CQRS commands/queries
   - Command pattern: 100% adoption ✅
   - Handler pattern: Consistent

✅ [8/8] Events (48 min)
   - 182 domain events catalogued
   - Outbox pattern: 100% coverage ✅
   - Event versioning: Not implemented ⚠️

Phase 2 Complete: 65 minutes (limited by slowest agent)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 Phase 3: Aggregation (10-15 min)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📊 Reading 19 agent reports...
📊 Extracting key metrics...
📊 Generating 30 tables...
📊 Calculating overall scores...
📊 Consolidating priorities...

✅ Master report generated!

Phase 3 Complete: 12 minutes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 Phase 4: TODO Update (Optional, 10-15 min)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

⏳ Running mgmt_todo_manager agent...

📋 Reading analysis results:
   - security_analysis.md (5 P0 vulnerabilities)
   - testing_analysis.md (14 aggregates <80%)
   - domain_model_analysis.md (14 missing opt. locking)

🔍 Verifying completed tasks:
   ✅ P0-1: Dev Mode Bypass (VERIFIED: panic in auth.go:45)
   ⏳ P0-2: BOLA in 60 endpoints (40/60 done, 67%)
   ❌ P0-3: SSRF in webhooks (NOT STARTED)

📝 Adding new tasks from analysis:
   + P0-6: Event versioning not implemented (from events_analysis)
   + P1-8: N+1 queries in 12 locations (from data_quality_analysis)
   + P2-15: 23 endpoints missing Swagger (from api_analysis)

✅ TODO.md updated!
   - 3 tasks marked complete ✅
   - 12 new tasks added from analysis
   - 8 priorities adjusted based on scores

Phase 4 Complete: 12 minutes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ ANALYSIS COMPLETE (WITH TODO UPDATE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

**Duration**: 2 hours 32 minutes
**Agents Executed**: 19/19 (100%)
**Reports Generated**: 20 (19 individual + 1 master)
**Tables Generated**: 30 (complete coverage)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 OVERALL ARCHITECTURE SCORE: 7.8/10 (GOOD)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

| Category              | Score  | Status      |
|-----------------------|--------|-------------|
| Domain Model          | 8.5/10 | ✅ Good      |
| Persistence           | 7.8/10 | ⚠️  Fair     |
| API                   | 9.2/10 | ✅ Excellent |
| Testing               | 8.2/10 | ✅ Good      |
| Security              | 4.5/10 | ❌ Critical  |
| AI/ML                 | 6.0/10 | ⚠️  Fair     |
| Infrastructure        | 8.0/10 | ✅ Good      |
| Resilience            | 7.5/10 | ⚠️  Fair     |
| Code Quality (SOLID)  | 8.8/10 | ✅ Good      |
| Data Quality          | 7.2/10 | ⚠️  Fair     |
| Documentation         | 7.8/10 | ⚠️  Fair     |

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🚨 CRITICAL FINDINGS (P0)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 🔴 Dev Mode Bypass (CVSS 9.1) - Auth bypass in production
2. 🔴 SSRF in Webhooks (CVSS 9.1) - Can access internal services
3. 🔴 BOLA in 60 Endpoints (CVSS 8.2) - No ownership checks
4. 🔴 Resource Exhaustion (CVSS 7.5) - No max page size
5. 🔴 Missing RBAC (CVSS 7.1) - 95 endpoints lack role checks

⚠️  THESE MUST BE FIXED BEFORE PRODUCTION DEPLOYMENT

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📂 GENERATED FILES
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Master Report:
✅ code-analysis/comprehensive/MASTER_ANALYSIS.md (30 tables)

Individual Reports (19):
✅ code-analysis/architecture/deterministic_metrics.md (baseline)
✅ code-analysis/domain-analysis/domain_model_analysis.md
✅ code-analysis/domain-analysis/value_objects_analysis.md
✅ code-analysis/domain-analysis/entity_relationships_analysis.md
✅ code-analysis/domain-analysis/use_cases_analysis.md
✅ code-analysis/domain-analysis/events_analysis.md
✅ code-analysis/infrastructure/persistence_analysis.md
✅ code-analysis/infrastructure/integration_analysis.md
✅ code-analysis/infrastructure/workflows_analysis.md
✅ code-analysis/infrastructure/infrastructure_analysis.md
✅ code-analysis/infrastructure/api_analysis.md
✅ code-analysis/quality/testing_analysis.md
✅ code-analysis/quality/security_analysis.md (🚨 P0 issues)
✅ code-analysis/quality/resilience_analysis.md
✅ code-analysis/quality/data_quality_analysis.md
✅ code-analysis/quality/code_style_analysis.md
✅ code-analysis/quality/documentation_analysis.md
✅ code-analysis/quality/solid_principles_analysis.md
✅ code-analysis/ai-ml/ai_ml_analysis.md

Total Size: ~2.5 MB

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎯 NEXT STEPS
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 📖 Review master report:
   code-analysis/comprehensive/MASTER_ANALYSIS.md

2. 🚨 Address P0 security vulnerabilities IMMEDIATELY
   (See: code-analysis/quality/security_analysis.md)

3. 📋 Update TODO.md with findings:
   /update-todo

4. 🎯 Plan Sprint 1-2 to fix critical issues

5. 📊 Share report with team for architectural review
```

---

## 📋 Comparison with Other Commands

| Command | Agents | Duration | Cache? | Output | Use Case |
|---------|--------|----------|--------|--------|----------|
| `/pre-analyze --quick` | 6 (31%) | 5-10 min | ✅ Yes | 6 JSON | Daily dev |
| `/pre-analyze --deep` | 14 (74%) | 15-30 min | ✅ Yes | 14 JSON | Weekly review |
| `/analyze` | 6 (31%) | 5-10 min | ❌ No | 1 MD report | One-time check |
| **`/analyze-complete`** | **19 (100%)** | **2-3 hours** | ❌ **No** | **Master MD + 19** | **Monthly audit** |

---

## 🎯 When to Use

### ✅ Use `/analyze-complete` when:
- Starting new project phase
- Before major refactoring
- Quarterly architecture review
- Pre-production deployment audit
- Onboarding architecture team
- Creating comprehensive documentation
- After multiple sprints (monthly/quarterly)
- Need ALL 30 tables for complete report

### ❌ Don't use when:
- Need quick feedback → use `/pre-analyze --quick`
- Only checking one area → use `/analyze --domain-only`
- In CI pipeline (too slow) → use `/analyze --changed-files-only`
- Daily development → use `/pre-analyze --quick` (caches results)

---

## 📊 Output Structure

```
code-analysis/
├── comprehensive/
│   └── MASTER_ANALYSIS.md          ⭐ Main report (30 tables)
│
├── architecture/
│   └── deterministic_metrics.md    ⭐ Factual baseline
│
├── domain-analysis/
│   ├── domain_model_analysis.md     (Tables 1, 2, 5)
│   ├── value_objects_analysis.md    (Table 6)
│   ├── entity_relationships_analysis.md (Table 4)
│   ├── use_cases_analysis.md        (Table 10)
│   └── events_analysis.md           (Table 11)
│
├── infrastructure/
│   ├── persistence_analysis.md      (Tables 3, 7, 9)
│   ├── integration_analysis.md      (Tables 8, 12)
│   ├── workflows_analysis.md
│   ├── infrastructure_analysis.md   (Tables 29, 30)
│   └── api_analysis.md              (Tables 16, 17)
│
├── quality/
│   ├── testing_analysis.md          (Tables 22, 24, 25)
│   ├── security_analysis.md         (Tables 18, 21, 26, 27)
│   ├── resilience_analysis.md       (Tables 19, 20, 23)
│   ├── data_quality_analysis.md     (Tables 13, 14, 15)
│   ├── code_style_analysis.md
│   ├── documentation_analysis.md
│   └── solid_principles_analysis.md
│
└── ai-ml/
    └── ai_ml_analysis.md            (Table 28)
```

---

## 📖 Master Report Contents

The generated `MASTER_ANALYSIS.md` includes:

### 1. Executive Summary
- Overall architecture score (weighted average)
- Breakdown by category (domain, persistence, API, etc.)
- Key strengths (top 5)
- Critical issues (P0, P1, P2)

### 2. All 30 Tables
Complete coverage:
- **Domain & Architecture** (Tables 1-11)
- **Implementation Quality** (Tables 12-15)
- **API & Security** (Tables 16-21)
- **Testing & Quality** (Tables 22-23)
- **Security Deep Dive** (Tables 24-27)
- **AI/ML & Infrastructure** (Tables 28-30)

### 3. Detailed Analysis by Category
Summary from each of the 19 agent reports

### 4. Overall Score Calculation
Formula with weights shown

### 5. Top 20 Priorities
Consolidated from all agents, sorted by urgency

### 6. Agent Execution Summary
Duration, status, output for each agent

### 7. Appendix
All discovery commands used (reproducible)

---

## 🔗 Related Commands

- `/pre-analyze` - Fast analysis with caching (6-14 agents)
- `/analyze` - One-time analysis (6 agents)
- `/add-feature` - Implement features using analysis context
- `/review` - Code review specific code
- `/test-feature` - Run tests for feature
- `/update-todo` - Update TODO.md with findings ⭐

---

## 💡 Recommended Workflow

### Monthly/Quarterly Audit

```bash
# Step 1: Complete analysis (2-3 hours)
/analyze-complete --update-todo --export=html

# Step 2: Review master report
open code-analysis/comprehensive/MASTER_ANALYSIS.md

# Step 3: Address P0 issues
# (Implement fixes based on security_analysis.md)

# Step 4: Plan next sprint based on findings
# (Use TODO.md updated with priorities)
```

### After Major Milestone

```bash
# Full audit + documentation updates
/analyze-complete --update-readme --update-devguide --update-todo --create-issues

# This will:
# - Run full analysis
# - Update README.md with new stats
# - Update DEV_GUIDE.md with patterns
# - Update TODO.md with new tasks
# - Create GitHub issues for P0 vulnerabilities
```

---

## ⚠️ Important Notes

1. **Duration**: Allow 2-3 hours uninterrupted (agents run in background)
2. **Tokens**: Uses 150k-250k tokens (significant AI usage)
3. **Output Size**: ~2.5 MB total (19 reports + master)
4. **Not Cached**: Results are NOT saved to `.claude/analysis/` (one-time only)
5. **Resource Intensive**: CPU/Memory usage during parallel execution
6. **Network**: Requires stable connection for AI agents

---

## 🔧 Technical Details

### How It Works Internally

The command invokes `meta_orchestrator` which:

1. **Phase 0**: Runs `global_deterministic_analyzer` first (baseline)
2. **Phase 1**: Launches 9 core agents in parallel (background tasks)
3. **Phase 2**: Launches 8 specialized agents in parallel
4. **Phase 3**: Aggregates all 19 reports into master

Each agent:
- Reads codebase with Read/Grep/Glob/Bash tools
- Generates individual markdown report
- Writes to `code-analysis/` directory
- Returns summary to orchestrator

Orchestrator:
- Collects all summaries
- Calculates weighted scores
- Consolidates priorities
- Generates master report with 30 tables

### Why Not Cached?

Unlike `/pre-analyze`, this command:
- Generates comprehensive reports (not just JSON data)
- Includes ALL agents (not subset for development)
- Meant for periodic audits (not frequent use)
- Output is full markdown reports (not reusable cache)

---

**Orchestrator**: `meta_orchestrator` (coordinates all 19 agents)
**TODO Manager**: `mgmt_todo_manager` (updates TODO.md with findings)
**Total Agents**: 19 analysis + 1 management = 20 total
**Total Tables**: 30 (complete)
**Estimated Runtime**:
  - Theoretical (math): 2-3 hours (sum of longest in each phase)
  - Actual (orchestrator estimate): 8-12 hours
  - Reason: Agent processing overhead, token generation, heavy analysis
  - Without TODO update: 8-12 hours
  - With TODO update (`--update-todo`): 8.5-12.5 hours

**Why the difference?**: While phases run in parallel, individual agents may take longer than estimated due to:
  - Large codebase analysis (30 aggregates, 158 endpoints)
  - Token generation overhead (150k-250k tokens total)
  - Deep code inspection with grep/find/read
  - Complex table generation (30 tables with multiple dimensions)
**Estimated Tokens**: 150k-250k (analysis) + 10k-15k (TODO update)
**Output**:
  - Main: `code-analysis/comprehensive/MASTER_ANALYSIS.md`
  - Optional: `TODO.md` (updated with analysis findings)
