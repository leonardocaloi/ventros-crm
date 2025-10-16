---
name: analyze
description: Run comprehensive codebase analysis (domain, persistence, API, testing, security) with parameters
---

# Analyze Command

**Deep codebase analysis** without modifying any code. Generates reports and updates documentation.

---

## 🎯 What This Does

Runs specialized analyzers and generates comprehensive reports:

1. **Domain Model Analysis** - All 30 aggregates, events, value objects
2. **Persistence Analysis** - Repositories, entities, migrations
3. **API Analysis** - Endpoints, Swagger, missing docs
4. **Testing Analysis** - Coverage, missing tests, quality
5. **Security Analysis** - OWASP vulnerabilities, RBAC, BOLA
6. **Code Quality Analysis** - SOLID violations, technical debt

**Updates**: `.claude/AGENT_STATE.json` with findings

---

## 🚀 Usage

### Full Analysis (All Analyzers)
```bash
/analyze

# Runs all 6 analyzers sequentially
# Generates: /tmp/analysis_report.md
# Duration: ~5-10 min
# Tokens: ~15k-20k
```

### With Parameters
```bash
/analyze --parallel --update-devguide --export=json

/analyze --domain-only --verbose

/analyze --security --strict

/analyze --before-implement --update-p0
```

---

## 🎛️ Available Parameters

### Scope Control
- `--domain-only` - Only analyze domain layer
- `--persistence-only` - Only analyze repositories/entities
- `--api-only` - Only analyze HTTP endpoints
- `--testing-only` - Only analyze tests/coverage
- `--security-only` - Only run security analysis
- `--quality-only` - Only analyze code quality (SOLID, etc)
- *(Default: Run all analyzers)*

### Execution Control
- `--parallel` - Run all analyzers in parallel (faster, 2-3 min)
- `--sequential` - Run one at a time (default, more readable logs)
- `--verbose` - Show detailed logs
- `--quiet` - Only show summary

### Output Control
- `--export=FORMAT` - Export format: `md` (default), `json`, `html`, `pdf`
- `--output=PATH` - Custom output path (default: `/tmp/analysis_report.md`)
- `--update-devguide` - Update DEV_GUIDE.md with findings
- `--update-readme` - Update README.md with stats

### Integration Control
- `--update-p0` - Add findings to P0 file as TODOs
- `--create-issues` - Create GitHub issues for P0 vulnerabilities
- `--before-implement` - Run before implementing feature (saves to agent state)

### Filtering
- `--aggregate=NAME` - Analyze specific aggregate only
- `--bounded-context=NAME` - Analyze specific bounded context (crm, automation, core)
- `--since-commit=SHA` - Only analyze changes since commit
- `--changed-files-only` - Only analyze git-modified files

### Strictness
- `--strict` - Fail if any P0 issues found
- `--warnings-as-errors` - Treat warnings as errors
- `--ignore-p2` - Ignore P2 (nice-to-have) issues

---

## 📋 Parameter Examples

### Example 1: Quick Domain Check
```bash
/analyze --domain-only --quiet

# Output:
# ✅ 30 aggregates found
# ⚠️  14 aggregates missing version field (optimistic locking)
# ✅ 182 domain events defined
# ⚠️  5 aggregates with business logic in handlers
```

### Example 2: Security Audit Before Deploy
```bash
/analyze --security-only --strict --create-issues

# This will:
# 1. Run crm_security_analyzer
# 2. Find 5 P0 vulnerabilities (from TODO.md)
# 3. Create 5 GitHub issues
# 4. FAIL with exit code 1 (blocks deployment)
```

### Example 3: Pre-Implementation Analysis
```bash
/analyze --before-implement --update-p0 --parallel

# This will:
# 1. Run all 6 analyzers in parallel (~3 min)
# 2. Update .claude/AGENT_STATE.json with findings
# 3. Add P0 items to .claude/P0_ACTIVE_WORK.md
# 4. Ready for: /add-feature <request> --use-analysis
```

### Example 4: Changed Files Only (Fast CI Check)
```bash
/analyze --changed-files-only --since-commit=HEAD~1 --strict

# This will:
# 1. Get files changed in last commit
# 2. Analyze only those files
# 3. Fail if any P0 issues found
# 4. Use in CI: make analyze-changes
```

### Example 5: Complete Report with All Outputs
```bash
/analyze --parallel --export=html --update-devguide --update-readme --verbose

# This will:
# 1. Run all analyzers in parallel
# 2. Generate HTML report: /tmp/analysis_report.html
# 3. Update DEV_GUIDE.md with findings
# 4. Update README.md with stats
# 5. Show detailed logs
```

### Example 6: Specific Aggregate Deep Dive
```bash
/analyze --aggregate=Contact --verbose

# This will:
# 1. Analyze internal/domain/crm/contact/
# 2. Check persistence (ContactEntity, repository)
# 3. Check API (contact_handler.go)
# 4. Check tests (coverage, missing tests)
# 5. Security audit (RBAC, BOLA)
# 6. Generate detailed report for Contact only
```

---

## 📊 Output Format

### Markdown Report (Default)
```markdown
# Codebase Analysis Report

**Generated**: 2025-10-16 10:30:00
**Duration**: 3.5 minutes
**Analyzers**: 6 (all)
**Mode**: Parallel

---

## 📋 Summary

| Category | Score | Status | Issues |
|----------|-------|--------|--------|
| Domain Model | 85/100 | ✅ Good | 3 warnings |
| Persistence | 78/100 | ⚠️  Fair | 2 P1 issues |
| API | 92/100 | ✅ Excellent | 0 |
| Testing | 82/100 | ✅ Good | Coverage target met |
| Security | 45/100 | ❌ Critical | 5 P0 vulnerabilities |
| Code Quality | 88/100 | ✅ Good | 2 SOLID violations |
| **OVERALL** | **78/100** | ⚠️  **Fair** | **5 P0, 2 P1, 3 P2** |

---

## 🔴 P0 Issues (CRITICAL - Fix Immediately)

1. **Dev Mode Auth Bypass** (CVSS 9.1)
   - File: `infrastructure/http/middleware/auth.go:41`
   - Issue: `ENV=development` bypasses auth in production
   - Fix: Remove dev mode check, use feature flags

2. **SSRF in Webhooks** (CVSS 9.1)
   - File: `infrastructure/webhooks/webhook_handler.go:125`
   - Issue: No URL validation, can access internal services
   - Fix: Whitelist allowed domains

3. **BOLA in 60 GET endpoints** (CVSS 8.2)
   - Files: All `*_handler.go` files
   - Issue: No ownership verification
   - Fix: Add `CheckOwnership()` middleware

4. **Resource Exhaustion** (CVSS 7.5)
   - Files: 19 queries with pagination
   - Issue: No max page size (can request 1M records)
   - Fix: Cap at 1000 records

5. **RBAC Missing** (CVSS 7.1)
   - Files: 95 endpoints
   - Issue: No role checks
   - Fix: Add `RequireRole()` middleware

---

## 🟡 P1 Issues (Important - Fix This Sprint)

[...]

## ⚪ P2 Issues (Nice to Have - Future Sprint)

[...]

## ✅ Strengths

[...]

## 💡 Recommendations

[...]
```

### JSON Export (`--export=json`)
```json
{
  "generated_at": "2025-10-16T10:30:00Z",
  "duration_seconds": 210,
  "analyzers": ["domain", "persistence", "api", "testing", "security", "quality"],
  "overall_score": 78,
  "summary": {
    "domain": {"score": 85, "status": "good", "issues": 3},
    "persistence": {"score": 78, "status": "fair", "issues": 2},
    "api": {"score": 92, "status": "excellent", "issues": 0},
    "testing": {"score": 82, "status": "good", "issues": 0},
    "security": {"score": 45, "status": "critical", "issues": 5},
    "quality": {"score": 88, "status": "good", "issues": 2}
  },
  "issues": {
    "p0": [
      {
        "title": "Dev Mode Auth Bypass",
        "file": "infrastructure/http/middleware/auth.go",
        "line": 41,
        "cvss": 9.1,
        "description": "ENV=development bypasses auth in production",
        "fix": "Remove dev mode check, use feature flags"
      }
    ],
    "p1": [...],
    "p2": [...]
  }
}
```

---

## 🔄 Integration with /add-feature

Run analysis first, then use findings in implementation:

```bash
# Step 1: Analyze
/analyze --before-implement --update-p0 --parallel

# Step 2: Implement (uses analysis from agent state)
/add-feature Add Custom Fields aggregate --use-analysis

# The orchestrator will:
# - Read .claude/AGENT_STATE.json
# - Know about 30 existing aggregates
# - Know about persistence patterns
# - Know about API patterns
# - Implement with full context
```

---

## 🎯 Use Cases

### 1. Pre-Deploy Health Check
```bash
/analyze --security-only --strict

# Use in CI/CD:
# - Exit 0 if pass
# - Exit 1 if P0 found (blocks deploy)
```

### 2. Weekly Code Review
```bash
/analyze --parallel --export=html --update-readme

# Generate report for team review
```

### 3. Onboarding New Developer
```bash
/analyze --verbose --export=html

# Comprehensive codebase overview
```

### 4. Before Refactoring
```bash
/analyze --aggregate=Contact --verbose

# Understand current state before changing
```

### 5. CI Integration
```bash
/analyze --changed-files-only --since-commit=$CI_COMMIT_SHA --strict

# Only analyze changes, fail on P0
```

---

## 🔗 Related Commands

- `/add-feature` - Implement feature (uses analysis)
- `/review` - Code review existing code
- `/test-feature` - Run tests for specific feature
- `/update-todo` - Update TODO.md with findings

---

**Agents Invoked**:
- `crm_domain_model_analyzer`
- `crm_persistence_analyzer`
- `crm_api_analyzer`
- `crm_testing_analyzer`
- `crm_security_analyzer`
- `global_code_style_analyzer`

**Runtime**: 3-10 min (parallel: 2-3 min)
**Tokens**: 15k-25k (depends on scope)
**Output**: `/tmp/analysis_report.md` (or custom path)
