---
name: update-todo
description: Synchronize TODO.md with codebase state and analysis results
---

# Update TODO Command

Run the `mgmt_todo_manager` agent to consolidate and synchronize TODO.md with current codebase state.

## What This Does

1. **Consolidates Multiple TODO Sources**:
   - Merges `TODO.md`, `todo_go_pure_consolidation.md`, and analysis-generated tasks
   - Removes duplicates and organizes by priority (P0, P1, P2)

2. **Detects Completed Tasks**:
   - Runs deterministic checks (grep/find) to verify task completion
   - Marks completed tasks with ✅ and adds evidence (file:line)
   - Examples:
     ```bash
     # P0-1: Dev Mode Bypass
     grep -q "panic.*devMode" infrastructure/http/middleware/auth.go

     # Optimistic Locking
     find internal/domain -name "*.go" -exec grep -l "version.*int" {} \; | wc -l
     ```

3. **Re-prioritizes Based on Analysis Results**:
   - Reads latest security/testing/architecture analyses
   - Adds new P0 vulnerabilities from security_analysis.md
   - Updates coverage gaps from testing_analysis.md
   - Adjusts effort estimates based on progress

4. **Cross-references with Git Activity**:
   - Checks when code was last modified vs TODO update
   - Flags stale tasks if code changed after TODO update
   - Adds completion dates from git log

## When to Use

- After completing a major feature or fix
- After running `/full-analysis` (security/testing/architecture changes)
- Weekly review (if TODO.md > 7 days old)
- Before planning next sprint
- After merging PRs that affect roadmap

## Output

Updates `/home/caloi/ventros-crm/TODO.md` with:
- Current date in "Last Update"
- Accurate task status (verified via code analysis)
- Completed tasks marked with [x] and evidence
- New tasks from analyses added with proper priority
- Progress percentages updated

## Example Output

```markdown
## 📋 CONSOLIDATED TODO

**Last Update**: 2025-10-15 (Auto-updated by mgmt_todo_manager)
**Status**: ✅ Synced with codebase (P0-1 COMPLETE, P0-2 IN PROGRESS)
**Next Review**: 2025-10-22 (7 days)

## 🔴 PRIORITY 0: CRITICAL & URGENT

### Sprint 1-2: Security Fixes (3-4 weeks)

#### 1.1. ✅ **Dev Mode Bypass** (COMPLETE - 2025-10-14)
**Status**: Fixed in commit abc123
**Verification**: middleware/auth.go:45 now panics if devMode in production ✅

#### 1.2. 🔄 **BOLA in 60 GET Endpoints** (IN PROGRESS - 40/60 done)
**Status**: 40 handlers fixed, 20 remaining
**Remaining**:
- [ ] GET /campaigns/:id (infrastructure/http/handlers/campaign_handler.go:123)
- [ ] GET /sequences/:id (infrastructure/http/handlers/sequence_handler.go:89)
**Effort**: 2-3 days remaining
```

## Agent Invoked

Triggers: `/home/caloi/ventros-crm/.claude/agents/mgmt_todo_manager.md`

**Runtime**: ~10-15 minutes

**Reads**:
- TODO.md (current master)
- code-analysis/quality/security_analysis.md
- code-analysis/quality/testing_analysis.md
- code-analysis/architecture/AI_REPORT.md
- Git logs for recent changes

**Writes**:
- TODO.md (updated)
- code-analysis/archive/YYYY-MM-DD/todos/ (old variants)
