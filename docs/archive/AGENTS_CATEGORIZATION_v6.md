# 🏷️ Agent Categorization & Renaming Plan

**Date**: 2025-10-15
**Purpose**: Categorize 27 agents with clear prefixes based on scope, category, and architecture

---

## 📊 Categorization Criteria

### 1. **Scope Prefix**
- `crm_` - Ventros CRM specific (domain model, business logic, CRM patterns)
- `global_` - Applicable to any Go codebase (code style, SOLID, documentation)
- `meta_` - Orchestration and post-processing (coordinates other agents)
- `mgmt_` - Management/maintenance (keeps docs/todos updated)

### 2. **Category Suffix**
- `_analyzer` - Analyzes code and generates reports
- `_manager` - Manages/maintains documentation or state
- `_generator` - Generates new content (ADRs, docs)
- `_cleaner` - Cleanup and organization tasks

### 3. **Architecture Type**
- **Skeleton** - Simple agent with basic analysis (< 300 lines, mostly grep/bash)
- **Mixed** - Combines deterministic + AI analysis (300-800 lines)
- **Complex** - Deep analysis with multi-phase logic (> 800 lines)

---

## 🎯 New Naming Pattern

```
{scope}_{category}_{name}_{type}.md
```

**Examples**:
- `crm_domain_model_analyzer.md` - CRM-specific domain analysis
- `global_code_style_analyzer.md` - Applicable to any Go project
- `meta_orchestrator.md` - Coordinates all agents
- `mgmt_todo_manager.md` - Manages TODO.md

---

## 📋 Current Agents → New Names

### 🔴 CRM-Specific Agents (15 agents)

**Domain Analysis (5 agents)**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `domain_model_analyzer.md` | `crm_domain_model_analyzer.md` | Complex | ~900 | CRITICAL |
| `value_objects_analyzer.md` | `crm_value_objects_analyzer.md` | Mixed | ~400 | STANDARD |
| `entity_relationships_analyzer.md` | `crm_entity_relationships_analyzer.md` | Mixed | ~450 | STANDARD |
| `use_cases_analyzer.md` | `crm_use_cases_analyzer.md` | Mixed | ~500 | STANDARD |
| `events_analyzer.md` | `crm_events_analyzer.md` | Mixed | ~550 | STANDARD |

**Infrastructure (4 agents)**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `persistence_analyzer.md` | `crm_persistence_analyzer.md` | Complex | ~850 | MEDIUM |
| `integration_analyzer.md` | `crm_integration_analyzer.md` | Complex | ~700 | CRITICAL |
| `workflows_analyzer.md` | `crm_workflows_analyzer.md` | Mixed | ~600 | MEDIUM |
| `infrastructure_analyzer.md` | `crm_infrastructure_analyzer.md` | Complex | ~900 | HIGH |

**AI/ML (1 agent)**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `ai_ml_analyzer.md` | `crm_ai_ml_analyzer.md` | Complex | ~800 | CRITICAL |

**Quality (5 agents)**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `testing_analyzer.md` | `crm_testing_analyzer.md` | Complex | ~900 | CRITICAL |
| `security_analyzer.md` | `crm_security_analyzer.md` | Complex | ~1000 | CRITICAL |
| `resilience_analyzer.md` | `crm_resilience_analyzer.md` | Complex | ~800 | HIGH |
| `data_quality_analyzer.md` | `crm_data_quality_analyzer.md` | Complex | ~750 | MEDIUM |
| `api_analyzer.md` | `crm_api_analyzer.md` | Complex | ~700 | HIGH |

---

### 🌐 Global Agents (4 agents)

**Applicable to any Go codebase**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `code_style_analyzer.md` | `global_code_style_analyzer.md` | Mixed | ~600 | USER-REQ |
| `documentation_analyzer.md` | `global_documentation_analyzer.md` | Mixed | ~650 | USER-REQ |
| `solid_principles_analyzer.md` | `global_solid_principles_analyzer.md` | Mixed | ~700 | USER-REQ |
| `deterministic_analyzer.md` | `global_deterministic_analyzer.md` | Skeleton | ~200 | CRITICAL |

---

### 🎭 Meta Agents (4 agents)

**Orchestration & Post-Processing**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `orchestrator.md` | `meta_orchestrator.md` | Complex | ~600 | CRITICAL |
| `adr_generator.md` | `meta_adr_generator.md` | Mixed | ~400 | MEDIUM |
| `docs_cleanup.md` | `meta_docs_cleaner.md` | Skeleton | ~250 | LOW |
| `docs_consolidator.md` | `meta_docs_consolidator.md` | Mixed | ~500 | HIGH |

---

### 🛠️ Management Agents (4 agents)

**Documentation & State Management**:
| Current Name | New Name | Arch | Lines | Priority |
|--------------|----------|------|-------|----------|
| `todo_manager.md` | `mgmt_todo_manager.md` | Mixed | ~450 | HIGH |
| `docs_index_manager.md` | `mgmt_docs_index_manager.md` | Skeleton | ~300 | MEDIUM |
| `docs_reorganizer.md` | `mgmt_docs_reorganizer.md` | Mixed | ~385 | HIGH |
| *(NEW)* | `mgmt_makefile_updater.md` | Skeleton | ~200 | MEDIUM |
| *(NEW)* | `mgmt_readme_updater.md` | Skeleton | ~200 | MEDIUM |
| *(NEW)* | `mgmt_dev_guide_updater.md` | Skeleton | ~200 | MEDIUM |

---

## 📊 Statistics

### By Scope
- **CRM-Specific**: 15 agents (56%)
- **Global**: 4 agents (15%)
- **Meta**: 4 agents (15%)
- **Management**: 4 agents (15%) *+ 3 new = 7 (26%)*

### By Architecture
- **Skeleton**: 5 agents (19%) - Simple, grep-based
- **Mixed**: 12 agents (44%) - Deterministic + AI
- **Complex**: 10 agents (37%) - Deep multi-phase analysis

### By Priority
- **CRITICAL**: 7 agents (26%)
- **HIGH**: 4 agents (15%)
- **MEDIUM**: 6 agents (22%)
- **USER-REQUESTED**: 3 agents (11%)
- **STANDARD**: 4 agents (15%)
- **LOW**: 1 agent (4%)

---

## 🔄 Migration Steps

### Step 1: Rename Files (27 agents)
```bash
# CRM-Specific (15 agents)
mv .claude/agents/domain_model_analyzer.md .claude/agents/crm_domain_model_analyzer.md
mv .claude/agents/value_objects_analyzer.md .claude/agents/crm_value_objects_analyzer.md
mv .claude/agents/entity_relationships_analyzer.md .claude/agents/crm_entity_relationships_analyzer.md
mv .claude/agents/use_cases_analyzer.md .claude/agents/crm_use_cases_analyzer.md
mv .claude/agents/events_analyzer.md .claude/agents/crm_events_analyzer.md
mv .claude/agents/persistence_analyzer.md .claude/agents/crm_persistence_analyzer.md
mv .claude/agents/integration_analyzer.md .claude/agents/crm_integration_analyzer.md
mv .claude/agents/workflows_analyzer.md .claude/agents/crm_workflows_analyzer.md
mv .claude/agents/infrastructure_analyzer.md .claude/agents/crm_infrastructure_analyzer.md
mv .claude/agents/ai_ml_analyzer.md .claude/agents/crm_ai_ml_analyzer.md
mv .claude/agents/testing_analyzer.md .claude/agents/crm_testing_analyzer.md
mv .claude/agents/security_analyzer.md .claude/agents/crm_security_analyzer.md
mv .claude/agents/resilience_analyzer.md .claude/agents/crm_resilience_analyzer.md
mv .claude/agents/data_quality_analyzer.md .claude/agents/crm_data_quality_analyzer.md
mv .claude/agents/api_analyzer.md .claude/agents/crm_api_analyzer.md

# Global (4 agents)
mv .claude/agents/code_style_analyzer.md .claude/agents/global_code_style_analyzer.md
mv .claude/agents/documentation_analyzer.md .claude/agents/global_documentation_analyzer.md
mv .claude/agents/solid_principles_analyzer.md .claude/agents/global_solid_principles_analyzer.md
mv .claude/agents/deterministic_analyzer.md .claude/agents/global_deterministic_analyzer.md

# Meta (4 agents)
mv .claude/agents/orchestrator.md .claude/agents/meta_orchestrator.md
mv .claude/agents/adr_generator.md .claude/agents/meta_adr_generator.md
mv .claude/agents/docs_cleanup.md .claude/agents/meta_docs_cleaner.md
mv .claude/agents/docs_consolidator.md .claude/agents/meta_docs_consolidator.md

# Management (4 agents)
mv .claude/agents/todo_manager.md .claude/agents/mgmt_todo_manager.md
mv .claude/agents/docs_index_manager.md .claude/agents/mgmt_docs_index_manager.md
mv .claude/agents/docs_reorganizer.md .claude/agents/mgmt_docs_reorganizer.md
```

### Step 2: Update Agent `name:` Field
Update YAML frontmatter in each agent file:
```yaml
---
name: crm_domain_model_analyzer  # Update this
description: |
  ...
---
```

### Step 3: Update References
Files that reference agents:
- `.claude/agents/README.md` - Update all agent names
- `.claude/agents/meta_orchestrator.md` - Update agent list
- `.claude/commands/*.md` - Update slash commands if they reference agents

### Step 4: Create New Updater Agents
Create 3 new management agents:
- `mgmt_makefile_updater.md` - Keeps MAKEFILE.md in sync with Makefile
- `mgmt_readme_updater.md` - Keeps README.md updated
- `mgmt_dev_guide_updater.md` - Keeps DEV_GUIDE.md updated

---

## 🎯 Benefits of New Naming

### 1. **Clear Scope**
- `crm_*` - "This is Ventros CRM specific"
- `global_*` - "I can use this on any Go project"
- `meta_*` - "This orchestrates other agents"
- `mgmt_*` - "This maintains documentation"

### 2. **Easy Filtering**
```bash
# All CRM-specific agents
ls .claude/agents/crm_*.md

# All global agents (reusable)
ls .claude/agents/global_*.md

# All meta agents
ls .claude/agents/meta_*.md

# All management agents
ls .claude/agents/mgmt_*.md
```

### 3. **Better Organization**
```
.claude/agents/
├── crm_*.md           (15 agents) - Ventros CRM specific
├── global_*.md        (4 agents)  - Reusable for any Go project
├── meta_*.md          (4 agents)  - Orchestration
├── mgmt_*.md          (7 agents)  - Docs/state management
└── README.md
```

### 4. **Clearer Purpose**
Old: `documentation_analyzer.md` - "Analyzes what? CRM docs? General docs?"
New: `global_documentation_analyzer.md` - "Analyzes Go documentation (godoc, swagger) - works on any project"

---

## ✅ Validation Checklist

After renaming:
- [ ] All 27 agents renamed with correct prefixes
- [ ] YAML `name:` field updated in each agent
- [ ] `.claude/agents/README.md` updated
- [ ] `meta_orchestrator.md` agent list updated
- [ ] Slash commands updated (if any reference agents directly)
- [ ] Test: `ls .claude/agents/*.md | wc -l` returns 28 (27 agents + README)
- [ ] Test: `ls .claude/agents/crm_*.md | wc -l` returns 15
- [ ] Test: `ls .claude/agents/global_*.md | wc -l` returns 4
- [ ] Test: `ls .claude/agents/meta_*.md | wc -l` returns 4
- [ ] Test: `ls .claude/agents/mgmt_*.md | wc -l` returns 7 (4 existing + 3 new)

---

## 📝 Next Steps

1. ✅ Review this categorization plan
2. ⏳ Execute file renaming (Step 1)
3. ⏳ Update YAML name fields (Step 2)
4. ⏳ Update references (Step 3)
5. ⏳ Create new updater agents (Step 4)
6. ⏳ Update README.md and other docs
7. ⏳ Validate all changes

---

**Status**: DRAFT - Awaiting approval
**Total Agents**: 27 current + 3 new = 30 agents
**Naming Pattern**: `{scope}_{category}_{name}_{type}.md`
