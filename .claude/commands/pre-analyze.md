---
name: pre-analyze
description: Run comprehensive codebase analysis and save results for future commands to use (2 modes - quick/deep)
---

# Pre-Analyze Command

**Purpose**: Analyze the entire codebase and save results in `.claude/analysis/` for other commands to use.

**Why**: Gives `/add-feature` and other commands FULL context of the codebase before implementing.

---

## 🚀 Usage

### Quick Mode (5-10 min)
```bash
/pre-analyze
# OR
/pre-analyze --quick
```

**What it does**:
- Runs 6 analyzers in parallel
- Saves results to `.claude/analysis/`
- Updates AGENT_STATE.json

**Analyzers**:
1. `crm_domain_model_analyzer` - 30 aggregates, events, value objects
2. `crm_persistence_analyzer` - Entities, repositories, migrations
3. `crm_api_analyzer` - Endpoints, Swagger, handlers
4. `crm_testing_analyzer` - Coverage, missing tests
5. `crm_workflows_analyzer` - Temporal workflows, sagas
6. `crm_integration_analyzer` - External integrations (WAHA, Stripe, etc)

**Duration**: 5-10 minutes
**Tokens**: ~15k-20k

---

### Deep Mode (15-30 min)
```bash
/pre-analyze --deep
```

**What it does**: Everything in quick mode PLUS:
- Security deep dive (OWASP, CVE)
- Code quality audit (SOLID, technical debt)
- Data quality analysis
- Resilience patterns
- Event-driven architecture audit
- Value objects analysis
- Entity relationships mapping
- Documentation quality

**Additional Analyzers** (8 more):
7. `crm_security_analyzer` - OWASP, P0 vulnerabilities
8. `global_code_style_analyzer` - Go code style, conventions
9. `global_solid_principles_analyzer` - SOLID violations
10. `crm_data_quality_analyzer` - Data validation, consistency
11. `crm_resilience_analyzer` - Error handling, retries
12. `crm_events_analyzer` - Domain events, outbox pattern
13. `crm_value_objects_analyzer` - Value objects, primitives
14. `crm_entity_relationships_analyzer` - Entity relationships

**Duration**: 15-30 minutes
**Tokens**: ~40k-60k

---

## 📂 Output Structure

All results saved in `.claude/analysis/`:

```
.claude/analysis/
├── domain_model.json          # 30 aggregates, bounded contexts
├── persistence.json            # Entities, repos, migrations
├── api.json                    # 158 endpoints, Swagger status
├── testing.json                # Coverage %, missing tests
├── workflows.json              # Temporal workflows, sagas
├── integration.json            # WAHA, Stripe, Meta Ads
├── security.json               # P0 vulnerabilities (deep only)
├── code_quality.json           # SOLID, tech debt (deep only)
├── data_quality.json           # Validation gaps (deep only)
├── resilience.json             # Error handling (deep only)
├── events.json                 # 182+ events (deep only)
├── value_objects.json          # VOs analysis (deep only)
├── relationships.json          # Entity graphs (deep only)
└── last_run.json               # Metadata
```

### Example: `domain_model.json`
```json
{
  "timestamp": "2025-10-16T10:30:00Z",
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
    },
    {
      "name": "Campaign",
      "bounded_context": "automation",
      "path": "internal/domain/automation/campaign",
      "has_version_field": false,
      "events": ["campaign.created", "campaign.activated", "campaign.paused"],
      "value_objects": [],
      "repository_interface": "internal/domain/automation/campaign/repository.go",
      "tests_coverage": 65,
      "missing_tests": ["ActivateCampaign error cases", "Concurrent updates"]
    }
  ],
  "summary": {
    "total_aggregates": 30,
    "aggregates_with_version_field": 16,
    "aggregates_missing_version_field": 14,
    "total_events": 182,
    "bounded_contexts": ["crm", "automation", "core"]
  }
}
```

---

## 🔄 How Other Commands Use This

### `/add-feature` (automatic)
```bash
/add-feature Add Custom Fields aggregate

# Internally checks:
if [ -f .claude/analysis/last_run.json ]; then
  # Has analysis! Load context
  DOMAIN_ANALYSIS=$(cat .claude/analysis/domain_model.json)
  PERSISTENCE_ANALYSIS=$(cat .claude/analysis/persistence.json)

  # Now has FULL context of:
  # - 30 existing aggregates
  # - Patterns used (DDD, CQRS, Event-Driven)
  # - Which aggregates have version field
  # - Test coverage per aggregate
  # - etc
fi
```

### Manual check
```bash
# See what analysis exists
ls -lh .claude/analysis/

# See summary
cat .claude/analysis/last_run.json

# See domain analysis
cat .claude/analysis/domain_model.json | jq '.summary'
```

---

## 🎯 When to Run

### Run `/pre-analyze --quick` when:
- Starting a new feature
- Haven't analyzed in a while (>1 week)
- Major changes merged to main
- Onboarding new developer (AI or human)

### Run `/pre-analyze --deep` when:
- Before major refactoring
- Before security audit
- Before production deploy
- Quarterly architecture review

---

## 🔧 Implementation

This command invokes:

1. **`meta_orchestrator`** - Coordinates all analyzers
2. **Parallel execution** - All analyzers run at same time (if `--parallel`)
3. **Result aggregation** - Collects all outputs
4. **File writing** - Saves to `.claude/analysis/`
5. **State update** - Updates `AGENT_STATE.json`

---

## 📊 Expected Output

```bash
$ /pre-analyze --quick

📚 Pre-Analysis Starting...
Mode: quick (6 analyzers)
Estimated time: 5-10 minutes

🔄 Running analyzers in parallel...
├─ [1/6] crm_domain_model_analyzer... ⏳
├─ [2/6] crm_persistence_analyzer... ⏳
├─ [3/6] crm_api_analyzer... ⏳
├─ [4/6] crm_testing_analyzer... ⏳
├─ [5/6] crm_workflows_analyzer... ⏳
└─ [6/6] crm_integration_analyzer... ⏳

✅ [1/6] Domain Model Analysis complete (2.3 min)
   - 30 aggregates found
   - 14 missing version field
   - 182 events defined

✅ [2/6] Persistence Analysis complete (1.8 min)
   - 30 entities found
   - 30 repositories found
   - 45 migrations found

✅ [3/6] API Analysis complete (2.1 min)
   - 158 endpoints found
   - 23 missing Swagger docs
   - 60 missing BOLA checks

✅ [4/6] Testing Analysis complete (3.2 min)
   - Overall coverage: 82%
   - 14 aggregates < 80% coverage
   - 23 missing integration tests

✅ [5/6] Workflows Analysis complete (1.5 min)
   - 3 Temporal workflows found
   - 0 sagas found (using Temporal for all)
   - 1 coreography pattern (WAHA history import)

✅ [6/6] Integration Analysis complete (2.0 min)
   - 3 external services (WAHA, Stripe, Meta Ads)
   - 12 API clients
   - 5 webhooks configured

💾 Saving results to .claude/analysis/...
✅ Analysis complete!

📊 Summary saved to: .claude/analysis/last_run.json
📂 6 analysis files created in .claude/analysis/

🎯 Next steps:
1. Run /add-feature with full context
2. Review security.json for P0 issues (if deep mode)
3. Check testing.json for missing tests

Duration: 6.2 minutes
Tokens used: 18,450
```

---

## 🔗 Related Commands

- `/add-feature` - Uses analysis automatically
- `/analyze` - One-time analysis (doesn't save)
- `/test-feature` - Uses testing.json for context
- `/review` - Uses all analysis for baseline

---

**Agent Invoked**: `meta_orchestrator` (coordinates all analyzers)
**Saves to**: `.claude/analysis/*.json`
**Updates**: `AGENT_STATE.json`, `P0_ACTIVE_WORK.md` (if issues found)
**Tokens**: 15k-60k (depends on mode)
**Duration**: 5-30 min (depends on mode)
