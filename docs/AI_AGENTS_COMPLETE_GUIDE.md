# AI Agents - Complete System Guide

**Version**: 2.0
**Created**: 2025-10-16
**Purpose**: Complete guide to the AI-powered development system with visible agent coordination

---

## 📚 Table of Contents

1. [System Overview](#system-overview)
2. [How It Works](#how-it-works)
3. [Slash Commands (User Interface)](#slash-commands)
4. [AI Agents (32 Agents)](#ai-agents)
5. [Agent Coordination Chain](#agent-coordination-chain)
6. [Analysis-First Workflow](#analysis-first-workflow)
7. [State Management](#state-management)
8. [Complete Examples](#complete-examples)
9. [Quick Reference](#quick-reference)

---

## System Overview

### What Is This?

An intelligent AI development system that:
- ✅ **Actually RUNS code and tests** (not just generates)
- ✅ **Coordinates 32 specialized agents** working together
- ✅ **Analyzes before implementing** (optional pre-analysis)
- ✅ **Tracks work in real-time** (P0 file + Agent State)
- ✅ **Recommends architectural patterns** (Saga, Temporal, Choreography, Simple)
- ✅ **Shares context between agents** (all agents know about each other)

### Key Stats

- **32 Specialized Agents**: 15 CRM, 4 Global, 7 Meta, 6 Management
- **4 Main Slash Commands**: `/add-feature`, `/pre-analyze`, `/test-feature`, `/review`
- **5,800+ Lines of Agent Definitions**
- **2 Analysis Modes**: Quick (5-10 min) vs Deep (15-30 min)
- **30+ Parameters** for fine-grained control

---

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    USER (You)                               │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Types slash command
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              SLASH COMMAND (.claude/commands/*.md)          │
│  (/add-feature, /pre-analyze, /test-feature, /review)       │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Becomes prompt for Claude
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   CLAUDE (Me)                               │
│  Reads command, invokes agents via Task tool                │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Calls agents
                        ▼
┌─────────────────────────────────────────────────────────────┐
│            AI AGENTS (.claude/agents/*.md)                  │
│  32 specialized agents coordinating via Task tool           │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        │ Reads/writes state
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              STATE FILES                                     │
│  - P0_ACTIVE_WORK.md (real-time tracking)                   │
│  - AGENT_STATE.json (shared knowledge)                      │
│  - .claude/analysis/*.json (pre-analysis cache)             │
└─────────────────────────────────────────────────────────────┘
```

### Key Concepts

#### 1. Slash Commands are Markdown Files

Location: `.claude/commands/<name>.md`

Example: `/add-feature` → `.claude/commands/add-feature.md`

When you type `/add-feature Add Custom Fields`, the system:
1. Reads `add-feature.md`
2. Expands it into a full prompt
3. Sends to Claude (me)
4. I execute the instructions

#### 2. Agents are Markdown Files

Location: `.claude/agents/<name>.md`

Example: `meta_dev_orchestrator` → `.claude/agents/meta_dev_orchestrator.md`

When an agent is invoked:
1. Claude (me) uses the **Task tool**
2. Task tool creates a new Claude session with agent's markdown as prompt
3. Agent executes autonomously
4. Agent returns results via final message
5. Parent session receives results

#### 3. Agents Call Other Agents via Task Tool

**IMPORTANT**: Agents do NOT share context automatically. Each agent invocation is isolated.

**How agents communicate**:
- Read shared files: `AGENT_STATE.json`, `P0_ACTIVE_WORK.md`
- Write to shared files before exiting
- Pass results via `/tmp/` files or direct output

---

## Slash Commands

### `/add-feature` - Intelligent Feature Implementation

**Purpose**: Implement features from scratch or enhance existing ones

**Basic Usage**:
```bash
/add-feature Add Custom Fields aggregate

/add-feature Add endpoint to list all active campaigns

/add-feature Implement contact import from CSV
```

**With Analysis**:
```bash
# Run analysis first, then implement
/add-feature Import historical WhatsApp messages --analyze-first

# This will:
# 1. Call /pre-analyze (if not already run)
# 2. Call meta_context_builder (pattern recommendation)
# 3. Show recommended pattern (e.g., Temporal Workflow)
# 4. Ask for confirmation
# 5. Implement using recommended pattern
```

**With Real-Time Testing**:
```bash
/add-feature Add Broadcast aggregate --run-tests-realtime

# This will:
# 1. Implement domain layer
# 2. Call: /test-feature Broadcast --layer=domain
# 3. Show test results immediately
# 4. If tests pass → continue to application layer
# 5. If tests fail → ask user to review before continuing
```

**30+ Parameters**: See `AI_DEVELOPMENT_SYSTEM.md` for full list

---

### `/pre-analyze` - Pre-Implementation Analysis

**Purpose**: Analyze entire codebase and cache results for other commands

**Quick Mode** (5-10 min, 6 analyzers):
```bash
/pre-analyze
# OR
/pre-analyze --quick
```

**Analyzers Run**:
1. `crm_domain_model_analyzer` - 30 aggregates, events, value objects
2. `crm_persistence_analyzer` - Entities, repositories, migrations
3. `crm_api_analyzer` - 158 endpoints, Swagger status
4. `crm_testing_analyzer` - Coverage %, missing tests
5. `crm_workflows_analyzer` - Temporal workflows, sagas
6. `crm_integration_analyzer` - WAHA, Stripe, Meta Ads

**Output**: `.claude/analysis/*.json` files

**Deep Mode** (15-30 min, 14 analyzers):
```bash
/pre-analyze --deep
```

**Additional Analyzers** (8 more):
7. `crm_security_analyzer` - P0 vulnerabilities
8. `global_code_style_analyzer` - Go conventions
9. `global_solid_principles_analyzer` - SOLID violations
10. `crm_data_quality_analyzer` - Validation gaps
11. `crm_resilience_analyzer` - Error handling
12. `crm_events_analyzer` - 182+ events
13. `crm_value_objects_analyzer` - Value objects
14. `crm_entity_relationships_analyzer` - Entity relationships

**Why Run This**:
- `/add-feature` uses analysis to recommend patterns
- `/test-feature` uses analysis to show known gaps
- `/review` uses analysis for baseline comparison
- Results are cached (reused until you run again)

---

### `/test-feature` - Real-Time Test Execution

**Purpose**: Actually RUN `go test` and show results

**Basic Usage**:
```bash
/test-feature Contact

# Runs:
# - go test ./internal/domain/crm/contact/...
# - go test ./internal/application/commands/contact/...
# - go test ./infrastructure/persistence/gorm_contact_repository_test.go
```

**With Analysis Integration** (NEW):
```bash
# If you ran /pre-analyze first:
/test-feature Contact --coverage

# Output will include:
# 📊 Loading test analysis context...
# ✅ Found pre-analysis (mode: quick, age: 2h ago)
#
# 📋 Known Gaps from Analysis:
#   - Missing: TestContact_MergeContacts (concurrency test)
#   - Coverage gap: contact/aggregate.go:142-145
#
# 🧪 Running Tests...
# [actual go test output]
#
# 🔍 Gap Analysis:
#   Priority 1 (P0): Add TestContact_MergeContacts
#   Priority 2: Cover error handling in aggregate.go:142-145
```

**Parameters**:
- `--coverage` - Generate coverage report
- `--realtime` - Stream output in real-time
- `--layer=domain|application|infrastructure` - Test specific layer
- `--integration-only` - Only integration tests (requires DB)
- See more in `test-feature.md`

---

### `/review` - Automated Code Review

**Purpose**: Review code with 100-point scoring system

**Usage**:
```bash
/review

/review --strict  # 90% threshold instead of 80%

/review --fix     # Auto-fix issues
```

**Scoring**:
- Domain (25) + Application (20) + Infrastructure (15)
- SOLID (15) + Security (15) + Testing (10)
- **Total**: 100 points
- **Pass**: 80% (or 90% with `--strict`)

---

## AI Agents

### Categories

#### Meta Agents (7) - High-Level Coordination
| Agent | Purpose | Called By |
|-------|---------|-----------|
| `meta_dev_orchestrator` | Main feature implementation | `/add-feature` |
| `meta_context_builder` | Pattern recommendation | `meta_dev_orchestrator` |
| `meta_feature_architect` | Architecture planning | `meta_dev_orchestrator` |
| `meta_code_reviewer` | Code review | `/review`, `meta_dev_orchestrator` |

#### CRM Analyzers (15) - Domain-Specific Analysis
| Agent | Purpose | Called By |
|-------|---------|-----------|
| `crm_domain_model_analyzer` | Analyze 30 aggregates | `/pre-analyze`, `meta_dev_orchestrator` |
| `crm_persistence_analyzer` | Analyze entities & repos | `/pre-analyze` |
| `crm_api_analyzer` | Analyze 158 endpoints | `/pre-analyze` |
| `crm_testing_analyzer` | Analyze test coverage | `/pre-analyze`, `/test-feature` |
| `crm_workflows_analyzer` | Analyze Temporal workflows | `/pre-analyze` |
| `crm_integration_analyzer` | Analyze external integrations | `/pre-analyze` |
| `crm_security_analyzer` | Find P0 vulnerabilities | `/pre-analyze --deep` |
| ... (8 more) | See AI_DEVELOPMENT_SYSTEM.md | ... |

#### Global Analyzers (4) - Cross-Cutting Concerns
| Agent | Purpose |
|-------|---------|
| `global_code_style_analyzer` | Go code conventions |
| `global_solid_principles_analyzer` | SOLID violations |
| `global_documentation_analyzer` | Doc quality |
| `global_deterministic_analyzer` | Determinism checks |

---

## Agent Coordination Chain

### 🔍 How to SEE Agents Calling Each Other

When you run a command, you'll see clear logging of agent coordination:

```bash
/add-feature Import historical contacts --analyze-first
```

**Console Output** (with visible agent coordination):

```
📋 Slash Command: /add-feature
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📝 Request: Import historical contacts
🎯 Parameters detected:
   - analyze-first: true
   - run-tests-realtime: true (default)
   - update-p0: true (default)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🤖 AGENT COORDINATION CHAIN
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Level 1: /add-feature
   │
   ├─► 🤖 Calling: meta_dev_orchestrator
   │   Purpose: Main feature orchestration
   │   Input: "Import historical contacts" + params
   │   Status: Running...
   │
   │   Level 2: meta_dev_orchestrator
   │   │
   │   ├─► 🤖 Calling: meta_context_builder
   │   │   Purpose: Load analysis & recommend pattern
   │   │   Input: User request + analysis files
   │   │   Status: Running...
   │   │
   │   │   📚 Loading architectural context...
   │   │   ✅ Found pre-analysis (mode: quick, age: 1h ago)
   │   │   ✅ Loaded: domain_model.json (30 aggregates)
   │   │   ✅ Loaded: workflows.json (3 Temporal workflows)
   │   │
   │   │   🔍 Analyzing request area...
   │   │   Area: import
   │   │   Complexity: high
   │   │   Initial pattern: temporal_workflow
   │   │
   │   │   🧠 Deep analysis...
   │   │   Existing patterns:
   │   │     - Temporal Workflows: 3
   │   │     - Sagas: 0
   │   │
   │   │   ✅ Found similar: WAHAHistoryImportWorkflow
   │   │   Reference: internal/workflows/channel/waha_history_import_workflow.go
   │   │
   │   │   🎯 Final Recommendation: Temporal Workflow
   │   │   Rationale: Multi-step import with external API.
   │   │             Temporal provides retry, timeout, visibility.
   │   │             Follow WAHAHistoryImportWorkflow pattern.
   │   │
   │   │   ✅ meta_context_builder COMPLETE (duration: 2.3 min)
   │   │   Output: /tmp/context_recommendation.md
   │   │
   │   ├─► 🤖 Calling: meta_feature_architect
   │   │   Purpose: Create detailed architecture plan
   │   │   Input: Request + pattern recommendation + analysis
   │   │   Status: Running...
   │   │
   │   │   📐 Creating architecture plan...
   │   │   Pattern: Temporal Workflow
   │   │   Reference: WAHAHistoryImportWorkflow
   │   │
   │   │   📋 Plan Generated:
   │   │   Files to Create: 8
   │   │   Files to Modify: 3
   │   │   Estimated Time: 45 min
   │   │   Estimated Tokens: 25,000
   │   │
   │   │   ✅ meta_feature_architect COMPLETE (duration: 3.1 min)
   │   │   Output: /tmp/architecture_plan.md
   │   │
   │   ├─► 📊 User Confirmation Required
   │   │   [Shows plan]
   │   │   Continue? (y/N) y
   │   │
   │   ├─► 🔧 Implementation Phase
   │   │   [AI writes code]
   │   │
   │   ├─► 🤖 Calling: /test-feature ContactImport --layer=domain --realtime
   │   │   Purpose: Test domain layer
   │   │   Status: Running...
   │   │
   │   │   🧪 Running: go test ./internal/domain/crm/contact_import/...
   │   │   === RUN   TestContactImport_Process
   │   │   --- PASS: TestContactImport_Process (0.01s)
   │   │   PASS
   │   │   coverage: 100.0% of statements
   │   │
   │   │   ✅ Domain: 5/5 tests passed (100% coverage)
   │   │
   │   ├─► 🤖 Calling: meta_code_reviewer
   │   │   Purpose: Review all implemented code
   │   │   Input: Domain + Application + Infrastructure files
   │   │   Status: Running...
   │   │
   │   │   Level 3: meta_code_reviewer
   │   │   │
   │   │   ├─► 🤖 Calling: crm_domain_model_analyzer
   │   │   │   Purpose: Validate domain patterns
   │   │   │   Status: Running...
   │   │   │   ✅ Domain patterns: PASS (8.5/10)
   │   │   │
   │   │   ├─► 🤖 Calling: global_solid_principles_analyzer
   │   │   │   Purpose: Check SOLID violations
   │   │   │   Status: Running...
   │   │   │   ✅ SOLID: PASS (9/10)
   │   │   │
   │   │   ✅ Code Review: 87/100 (PASS)
   │   │
   │   ✅ meta_dev_orchestrator COMPLETE (duration: 42 min)
   │   Output: Feature implemented + tested + reviewed
   │
   ✅ /add-feature COMPLETE

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 AGENT COORDINATION SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Agents Invoked: 5
├─ meta_dev_orchestrator (Level 1)
├─ meta_context_builder (Level 2)
├─ meta_feature_architect (Level 2)
├─ meta_code_reviewer (Level 2)
│  ├─ crm_domain_model_analyzer (Level 3)
│  └─ global_solid_principles_analyzer (Level 3)

Total Duration: 42 minutes
Total Tokens: ~28,000
Files Created: 8
Files Modified: 3
Tests Added: 12 (100% coverage)
Code Review Score: 87/100 (PASS)

✅ Branch: feature/import-historical-contacts
✅ Commit: feat: Add contact import workflow
✅ Push: origin/feature/import-historical-contacts
```

---

## Analysis-First Workflow

### The Problem (Before v2.0)

When implementing features, agents didn't have full context:
- Don't know existing patterns (should I use Saga or Temporal?)
- Don't know similar features (can I copy this pattern?)
- Don't know test gaps (what tests are missing?)
- Don't know security issues (what are the P0 vulnerabilities?)

### The Solution (v2.0)

**Two-Phase Workflow**:

#### Phase 1: Pre-Analysis (Optional but Recommended)

```bash
/pre-analyze --quick  # 5-10 min
```

This runs 6-14 analyzers in parallel and saves results to `.claude/analysis/*.json`

**Benefits**:
- Results are cached (fast for subsequent commands)
- All agents can read these files
- Pattern recommendations are data-driven (not guesses)

#### Phase 2: Implementation with Context

```bash
/add-feature Import contacts
```

**What happens internally**:

```python
# Step 1: Check for pre-analysis
if os.path.exists('.claude/analysis/last_run.json'):
    analysis_exists = True
    domain_analysis = load('.claude/analysis/domain_model.json')
    workflows_analysis = load('.claude/analysis/workflows.json')
    # ... load all analysis files
else:
    analysis_exists = False
    # Will work without analysis, but less intelligent

# Step 2: Call meta_context_builder with analysis
context = meta_context_builder(
    user_request="Import contacts",
    analysis_files={
        'domain': domain_analysis,
        'workflows': workflows_analysis,
        # ...
    }
)

# Step 3: Context builder analyzes and recommends
# Area: import
# Existing similar feature: WAHAHistoryImportWorkflow
# Recommended pattern: Temporal Workflow
# Rationale: Multi-step, external API, needs retry

# Step 4: Show recommendation to user
print(f"""
🎯 Recommended Pattern: Temporal Workflow

Why?
- You have 3 existing Temporal Workflows
- Found similar feature: WAHAHistoryImportWorkflow
- Import operations need: retry, timeout, visibility
- Follow this pattern for consistency

Reference Implementation:
{workflows_analysis['temporal_workflows'][0]['file']}

Continue? (y/N)
""")

# Step 5: If user confirms, implement using recommended pattern
```

### Example: Contact Import Analysis

**User runs**:
```bash
/pre-analyze --quick
# Wait 7 minutes
/add-feature Import contacts from CSV with validation
```

**System detects**:
- Area: `import` (keyword: "import")
- Steps: 4+ (fetch → validate → transform → save)
- External APIs: 0 (CSV is local)
- Duration: Likely >1 min (batch processing)
- Existing pattern: WAHAHistoryImportWorkflow (similar!)

**Recommendation**:
```
Pattern: Temporal Workflow
Rationale:
  - Multi-step process (4+ steps)
  - Batch processing (could take minutes)
  - Need progress tracking
  - Similar to WAHAHistoryImportWorkflow

Reference: internal/workflows/channel/waha_history_import_workflow.go

Structure:
  internal/workflows/contact_import/
  ├── import_workflow.go
  ├── import_activities.go (FetchCSV, ValidateRows, TransformRows, SaveBatch)
  └── import_worker.go
```

---

## State Management

### 3 State Files

#### 1. `.claude/P0_ACTIVE_WORK.md` - Real-Time Tracking

**Purpose**: Track active development work per branch (can have multiple)

**Updated by**: All agents that implement features

**Structure**:
```markdown
# Active Development Work (P0)

**Last Updated**: 2025-10-16 14:30
**Total Active Branches**: 2

---

### Branch: `feature/custom-fields`
**Created**: 2025-10-16 10:00
**Developer**: meta_dev_orchestrator
**Status**: 🟡 In Progress

#### Current Request:
Add Custom Field aggregate

#### What's Being Done:
- [x] Domain layer (100% coverage)
- [x] Application layer (85% coverage)
- [ ] Infrastructure layer (60% done)

#### Test Results:
✅ Domain: 15/15 tests passed (100%)
✅ Application: 10/10 tests passed (85%)
⏳ Infrastructure: Not yet run

#### Next Steps:
1. Complete infrastructure layer
2. Run integration tests
3. Code review

---

### Branch: `feature/contact-import`
**Created**: 2025-10-16 14:15
**Developer**: meta_dev_orchestrator
**Status**: 🟢 Ready for Review

#### Current Request:
Import contacts from CSV

#### What's Being Done:
- [x] Temporal Workflow (WAHAHistoryImportWorkflow pattern)
- [x] All tests passing
- [x] Code review: 87/100

#### Test Results:
✅ All layers: 45/45 tests passed (95% coverage)

#### Next Steps:
1. Create PR
2. Wait for CI
3. Merge
```

**Why It's Important**:
- Shows exactly what's being worked on
- Tracks progress in real-time
- Supports multiple concurrent branches
- Helps user understand agent's current focus

---

#### 2. `.claude/AGENT_STATE.json` - Shared Knowledge

**Purpose**: Enable all agents to know what others have discovered

**Updated by**: All agents

**Structure**:
```json
{
  "version": "1.0",
  "last_updated": "2025-10-16T14:30:00Z",
  "active_branches": [
    "feature/custom-fields",
    "feature/contact-import"
  ],
  "current_context": {
    "working_branch": "feature/contact-import",
    "last_request": "Import contacts from CSV",
    "mode": "full_feature",
    "phase": "testing"
  },
  "agents": {
    "meta_dev_orchestrator": {
      "last_run": "2025-10-16T14:15:00Z",
      "status": "active",
      "current_task": "Import contacts from CSV",
      "findings": [
        "Recommended pattern: Temporal Workflow",
        "Similar feature: WAHAHistoryImportWorkflow",
        "Test coverage: 95%"
      ]
    },
    "meta_context_builder": {
      "last_run": "2025-10-16T14:16:00Z",
      "status": "completed",
      "findings": [
        "Area: import",
        "Pattern: temporal_workflow",
        "Reference: WAHAHistoryImportWorkflow"
      ]
    },
    "meta_code_reviewer": {
      "last_run": "2025-10-16T14:28:00Z",
      "status": "completed",
      "findings": [
        "Score: 87/100",
        "Domain: 9/10",
        "SOLID: 8.5/10"
      ]
    }
  },
  "test_results": {
    "unit_tests": {
      "status": "passed",
      "passed": 45,
      "failed": 0,
      "coverage": 95.2,
      "last_run": "2025-10-16T14:25:00Z"
    },
    "integration_tests": {
      "status": "not_run",
      "last_run": null
    }
  },
  "shared_knowledge": {
    "current_aggregates": 30,
    "bounded_contexts": ["crm", "automation", "core"],
    "total_endpoints": 158,
    "temporal_workflows": 3,
    "known_patterns": {
      "import": "temporal_workflow",
      "api_endpoint": "simple_handler",
      "background_job": "temporal_workflow"
    }
  }
}
```

**Why It's Important**:
- All agents read this BEFORE starting work
- Prevents duplicate work
- Shares discoveries (pattern recommendations, test results, etc.)
- Enables intelligent decision-making based on what others found

---

#### 3. `.claude/analysis/*.json` - Analysis Cache

**Purpose**: Cache results from `/pre-analyze` for reuse

**Created by**: `/pre-analyze` command
**Read by**: `/add-feature`, `/test-feature`, `/review`, `meta_context_builder`

**Files**:
- `last_run.json` - Metadata (when, mode, duration)
- `domain_model.json` - 30 aggregates, events, value objects
- `persistence.json` - Entities, repositories, migrations
- `api.json` - 158 endpoints, Swagger status
- `testing.json` - Coverage per aggregate, missing tests
- `workflows.json` - Temporal workflows, sagas
- `integration.json` - External integrations (WAHA, Stripe, etc.)
- `security.json` - P0 vulnerabilities (deep mode only)
- ... (7 more in deep mode)

**Example: `domain_model.json`**:
```json
{
  "timestamp": "2025-10-16T10:00:00Z",
  "mode": "quick",
  "aggregates": [
    {
      "name": "Contact",
      "bounded_context": "crm",
      "path": "internal/domain/crm/contact",
      "has_version_field": true,
      "events": ["contact.created", "contact.updated", "contact.deleted"],
      "value_objects": ["WhatsAppNumber", "EmailAddress"],
      "repository_interface": "internal/domain/crm/contact/repository.go",
      "tests_coverage": 100,
      "missing_tests": []
    }
  ],
  "summary": {
    "total_aggregates": 30,
    "aggregates_with_version_field": 16,
    "aggregates_missing_version_field": 14
  }
}
```

---

## Complete Examples

### Example 1: Add Feature with Full Analysis

**Scenario**: User wants to add a new feature but isn't sure what pattern to use

**Commands**:
```bash
# Step 1: Analyze codebase (one-time, results cached)
/pre-analyze --quick

# Step 2: Implement with analysis
/add-feature Import historical WhatsApp messages --analyze-first --run-tests-realtime
```

**What Happens**:

1. **`/add-feature` reads analysis** from `.claude/analysis/`
2. **Calls `meta_context_builder`**:
   - Detects area: "import"
   - Finds similar feature: WAHAHistoryImportWorkflow
   - Recommends: Temporal Workflow pattern
3. **Calls `meta_feature_architect`**:
   - Creates detailed plan following Temporal pattern
   - References WAHAHistoryImportWorkflow as template
4. **Shows plan to user** → User confirms
5. **Implements**:
   - Domain layer → Tests → ✅ Pass
   - Application layer → Tests → ✅ Pass
   - Infrastructure layer → Tests → ✅ Pass
6. **Calls `meta_code_reviewer`**:
   - Score: 88/100 → PASS
7. **Commits + Creates PR**

**Result**: Feature implemented following existing patterns, fully tested, code-reviewed

---

### Example 2: Test-Driven Development with Analysis

**Scenario**: User wants to see test gaps before writing new tests

**Commands**:
```bash
# Step 1: Analyze testing gaps
/pre-analyze --quick

# Step 2: See what tests are missing
/test-feature Contact --coverage

# Output:
# 📋 Known Gaps from Analysis:
#   - Missing: TestContact_MergeContacts (P1)
#   - Coverage gap: contact/aggregate.go:142-145
#
# Current Coverage: 88% (target: 90%)
# Recommendation: Write 2 missing tests

# Step 3: Write the missing tests
# [AI or human writes tests]

# Step 4: Re-run to verify
/test-feature Contact --coverage

# Output:
# ✅ Coverage: 95% (target: 90%)
# 🎉 All P1 gaps resolved!
```

---

### Example 3: Parallel Analysis (Deep Mode)

**Scenario**: Before major refactoring, user wants complete analysis

**Command**:
```bash
/pre-analyze --deep --parallel
```

**What Happens** (all in parallel):

```
🔄 Running 14 analyzers in parallel...

├─ [1/14] crm_domain_model_analyzer... ⏳
├─ [2/14] crm_persistence_analyzer... ⏳
├─ [3/14] crm_api_analyzer... ⏳
├─ [4/14] crm_testing_analyzer... ⏳
├─ [5/14] crm_workflows_analyzer... ⏳
├─ [6/14] crm_integration_analyzer... ⏳
├─ [7/14] crm_security_analyzer... ⏳
├─ [8/14] global_code_style_analyzer... ⏳
├─ [9/14] global_solid_principles_analyzer... ⏳
├─ [10/14] crm_data_quality_analyzer... ⏳
├─ [11/14] crm_resilience_analyzer... ⏳
├─ [12/14] crm_events_analyzer... ⏳
├─ [13/14] crm_value_objects_analyzer... ⏳
└─ [14/14] crm_entity_relationships_analyzer... ⏳

[2.1 min] ✅ Domain Model (30 aggregates, 182 events)
[1.8 min] ✅ Persistence (30 entities, 45 migrations)
[2.3 min] ✅ API (158 endpoints, 23 missing Swagger)
[3.1 min] ✅ Testing (82% coverage, 14 aggregates <80%)
[1.5 min] ✅ Workflows (3 Temporal, 0 Sagas)
[2.0 min] ✅ Integration (WAHA, Stripe, Meta Ads)
[4.2 min] ✅ Security (5 P0 vulnerabilities found!)
[1.9 min] ✅ Code Style (92% compliance)
[2.5 min] ✅ SOLID (3 violations in handlers)
[2.8 min] ✅ Data Quality (8 validation gaps)
[2.2 min] ✅ Resilience (12 retry opportunities)
[2.7 min] ✅ Events (182 events, Outbox 100% coverage)
[2.1 min] ✅ Value Objects (45 VOs, 8 primitives)
[3.0 min] ✅ Relationships (Entity graph generated)

💾 Saving 14 analysis files to .claude/analysis/...
✅ Deep analysis complete!

Duration: 18.2 minutes (parallel execution)
Files created: 14
Total size: 2.3 MB

🚨 CRITICAL FINDINGS:
   - 5 P0 Security Vulnerabilities (see security.json)
   - 14 Aggregates <80% test coverage
   - 23 Endpoints missing Swagger docs
   - 60 Endpoints missing BOLA checks

📊 Recommendations:
   1. Address P0 security issues IMMEDIATELY
   2. Add missing tests (run: /test-feature --show-gaps)
   3. Generate Swagger docs (run: make swagger)
   4. Add BOLA checks to endpoints
```

**Result**: Complete codebase analysis cached for next 7 days

---

## Quick Reference

### Command Cheat Sheet

```bash
# Analysis (run once, cache for days)
/pre-analyze                                 # Quick mode (6 analyzers, 5-10 min)
/pre-analyze --deep                          # Deep mode (14 analyzers, 15-30 min)
/pre-analyze --parallel                      # Run analyzers in parallel

# Implementation
/add-feature <description>                   # Basic implementation
/add-feature <desc> --analyze-first          # Load analysis first
/add-feature <desc> --run-tests-realtime     # Test after each layer
/add-feature <desc> --mode=enhancement       # Fast mode for small changes

# Testing
/test-feature <Aggregate>                    # Test specific aggregate
/test-feature --layer=domain                 # Test domain layer only
/test-feature --coverage --show-uncovered    # Coverage + gap analysis
/test-feature --changed-only                 # Only test changed files

# Code Review
/review                                      # Full review (80% threshold)
/review --strict                             # Strict mode (90% threshold)
/review --fix                                # Auto-fix issues
```

### Agent Invocation Chain

**Level 1** (User-Facing Commands):
- `/add-feature`
- `/pre-analyze`
- `/test-feature`
- `/review`

**Level 2** (Orchestrators):
- `meta_dev_orchestrator`
- `meta_context_builder`
- `meta_feature_architect`
- `meta_code_reviewer`

**Level 3** (Specialized Analyzers):
- `crm_domain_model_analyzer`
- `crm_testing_analyzer`
- `global_solid_principles_analyzer`
- ... (29 more)

### State Files

| File | Purpose | Updated By | Read By |
|------|---------|------------|---------|
| `P0_ACTIVE_WORK.md` | Track active work | All feature-implementing agents | All agents |
| `AGENT_STATE.json` | Shared knowledge | All agents | All agents |
| `.claude/analysis/*.json` | Cached analysis | `/pre-analyze` | `/add-feature`, `/test-feature`, `/review` |

---

## 🎯 Success Metrics

With this system, you should see:

✅ **Faster Development**: Analysis-first means better decisions
✅ **Better Quality**: Real-time testing catches issues early
✅ **Consistent Patterns**: AI recommends patterns based on existing code
✅ **Full Visibility**: See exactly which agents are doing what
✅ **Context Sharing**: All agents know about each other's findings
✅ **Reduced Errors**: Pre-analysis catches issues before implementation

---

## 📖 Related Documentation

- **AI_DEVELOPMENT_SYSTEM.md** - Full parameter reference
- **DEV_ORCHESTRATION_SUMMARY.md** - System architecture overview
- **CLAUDE.md** - Project-specific instructions
- **DEV_GUIDE.md** - Manual development guide
- **.claude/commands/*.md** - Individual slash command docs
- **.claude/agents/*.md** - Individual agent specs

---

**Last Updated**: 2025-10-16
**Version**: 2.0
**Total Lines**: 1,100+
**Maintainer**: Ventros CRM Team + Claude Code
