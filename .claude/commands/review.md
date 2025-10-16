---
name: review
description: Automated code review with 100-point scoring system (DDD, Clean Architecture, SOLID, Security)
---

# Review Command

**Automated code review** for DDD, Clean Architecture, SOLID, and security compliance.

**Generates**: `/tmp/code_review.md` with score and actionable recommendations

---

## 🎯 What This Does

Invokes **`meta_code_reviewer`** agent to review code:

1. **Reads Files** - Domain, application, infrastructure layers
2. **Validates Architecture** - DDD, Clean Architecture, CQRS
3. **Checks SOLID** - Single Responsibility, Open/Closed, etc.
4. **Security Audit** - OWASP (RBAC, BOLA, SQL injection)
5. **Testing Coverage** - Unit, integration, E2E
6. **Generates Report** - 100-point score with pass/fail decision

**Updates**: `.claude/AGENT_STATE.json` with review score

---

## 🚀 Usage

### Review Specific Aggregate
```bash
/review Contact

# Reviews:
# - internal/domain/crm/contact/*.go
# - internal/application/commands/contact/*.go
# - infrastructure/persistence/gorm_contact_repository.go
# - infrastructure/http/handlers/contact_handler.go
```

### Review All Changes (Git)
```bash
/review --changed-only

# Reviews only files in: git diff --name-only
```

### With Parameters
```bash
/review Contact --strict --update-p0

/review --aggregate=Campaign --security-focus

/review --changed-only --fail-below=90

/review --all --export=json --verbose
```

---

## 🎛️ Available Parameters

### Scope Control
- `<AGGREGATE_NAME>` - Review specific aggregate (e.g., `Contact`, `Campaign`)
- `--all` - Review entire codebase (slow, 30+ min)
- `--changed-only` - Review only git-modified files (fast)
- `--aggregate=NAME` - Explicit aggregate name
- `--layer=LAYER` - Review specific layer: `domain`, `application`, `infrastructure`

### Review Focus
- `--security-focus` - Deep security audit (OWASP, CVE)
- `--architecture-focus` - Deep architecture audit (DDD, Clean Arch)
- `--testing-focus` - Deep testing audit (coverage, quality)
- *(Default: Balanced review)*

### Strictness
- `--strict` - Require 90%+ score (default: 80%+)
- `--fail-below=N` - Fail if score < N%
- `--warnings-as-errors` - Treat warnings as blocking issues
- `--ignore-p2` - Ignore P2 (nice-to-have) issues

### Output Control
- `--export=FORMAT` - Export format: `md` (default), `json`, `html`
- `--output=PATH` - Custom output path (default: `/tmp/code_review.md`)
- `--verbose` - Show detailed checks
- `--quiet` - Only show score and pass/fail

### Integration Control
- `--update-p0` - Add P0 issues to P0 file
- `--create-issues` - Create GitHub issues for P0 items
- `--block-merge` - Exit 1 if fail (for CI)

### Automated Fixes
- `--auto-fix` - Automatically fix simple issues (fmt, imports)
- `--suggest-fixes` - Generate fix suggestions (don't apply)

---

## 📋 Parameter Examples

### Example 1: Quick Review
```bash
/review Contact --quiet

# Output:
# Contact: 85/100 (85%) ✅ PASS
```

### Example 2: Strict Security Audit
```bash
/review Campaign --security-focus --strict --update-p0

# This will:
# 1. Deep security audit of Campaign
# 2. Require 90%+ score
# 3. Add any P0 vulnerabilities to P0 file
# 4. Generate detailed report
```

### Example 3: Pre-Commit Review
```bash
/review --changed-only --fail-below=80 --block-merge

# This will:
# 1. Review only changed files
# 2. Fail if score < 80%
# 3. Exit 1 (blocks commit in pre-commit hook)
```

### Example 4: Full Codebase Audit (Weekly)
```bash
/review --all --export=html --verbose

# This will:
# 1. Review all 30 aggregates
# 2. Generate HTML report
# 3. Duration: ~30-45 min
# 4. Tokens: ~50k-80k
```

### Example 5: Review with Auto-Fix
```bash
/review Contact --auto-fix --verbose

# This will:
# 1. Review Contact aggregate
# 2. Automatically fix:
#    - go fmt issues
#    - missing imports
#    - simple godoc comments
# 3. Show what was fixed
# 4. Re-review to verify
```

---

## 📊 100-Point Scoring System

### Domain Layer (25 points)
- **Aggregate Root** (5 pts) - Has ID, version, domain events
- **Business Logic** (5 pts) - In domain, not in handlers
- **Events** (3 pts) - Naming: `aggregate.action`, proper metadata
- **Repository Interface** (3 pts) - In domain, not infrastructure
- **Value Objects** (3 pts) - No primitive obsession
- **Factory Methods** (2 pts) - `NewAggregate()` pattern
- **Invariants** (2 pts) - Enforced in aggregate
- **No External Deps** (2 pts) - Pure Go, no imports from infra/app

### Application Layer (20 points)
- **Command Pattern** (5 pts) - Command struct + handler
- **Validation** (3 pts) - In command, not handler
- **No Business Logic** (5 pts) - Delegates to domain
- **Event Publishing** (3 pts) - Via EventBus after persistence
- **DTOs** (2 pts) - Request/Response separation
- **Error Handling** (2 pts) - No panics, proper errors

### Infrastructure Layer (15 points)
- **Repository Implementation** (4 pts) - Implements domain interface
- **HTTP Handler** (3 pts) - Thin adapter, delegates to commands
- **Swagger** (2 pts) - Complete annotations
- **Migration** (3 pts) - Up + Down, idempotent
- **RLS Policy** (3 pts) - Tenant isolation

### SOLID Principles (15 points)
- **Single Responsibility** (3 pts) - One reason to change
- **Open/Closed** (3 pts) - Extend via composition
- **Liskov Substitution** (3 pts) - Interfaces used correctly
- **Interface Segregation** (3 pts) - Focused interfaces
- **Dependency Inversion** (3 pts) - Depend on abstractions

### Security (15 points)
- **RBAC** (3 pts) - Role check in handler
- **BOLA** (3 pts) - Ownership verification
- **Input Validation** (3 pts) - SQL injection, XSS prevention
- **Rate Limiting** (2 pts) - Applied to endpoints
- **Tenant Isolation** (2 pts) - RLS enforced
- **Sensitive Data** (2 pts) - Masked in logs

### Testing (10 points)
- **Domain Tests** (3 pts) - Unit tests with 100% coverage
- **Application Tests** (3 pts) - Handler tests with mocks
- **Integration Tests** (2 pts) - Repository with real DB
- **E2E Tests** (2 pts) - Full HTTP flow

---

## 📊 Output Example

```markdown
# Code Review Report

**Aggregate**: Contact
**Reviewed**: 2025-10-16 10:30:00
**Files**: 12 files
**Total Score**: 85/100 (85%) ✅ **PASS**

---

## 📊 Score Breakdown

| Category | Score | Max | Percentage |
|----------|-------|-----|------------|
| Domain Layer | 23 | 25 | 92% ✅ |
| Application Layer | 18 | 20 | 90% ✅ |
| Infrastructure | 13 | 15 | 87% ✅ |
| SOLID Principles | 12 | 15 | 80% ✅ |
| Security | 12 | 15 | 80% ✅ |
| Testing | 7 | 10 | 70% ⚠️  |
| **TOTAL** | **85** | **100** | **85%** |

---

## ✅ PASS

Code is production-ready with excellent quality.

---

## 🔍 Detailed Findings

### ✅ Strengths (What's Good)

1. **Domain Layer** (23/25) - Excellent DDD implementation
   - ✅ Aggregate has ID, version field, domain events
   - ✅ Business logic in domain, not handlers
   - ✅ Events: `contact.created`, `contact.updated`, `contact.deleted`
   - ✅ Repository interface in domain layer
   - ✅ Value objects: `WhatsAppNumber`, `EmailAddress`
   - ✅ Factory method: `NewContact()`
   - ⚠️  Missing: Invariant for email format validation

2. **Application Layer** (18/20) - Good command pattern usage
   - ✅ Commands: Create, Update, Delete
   - ✅ Handlers delegate to domain
   - ✅ DTOs: CreateContactRequest, ContactResponse
   - ✅ Event publishing via EventBus
   - ⚠️  Missing: Input validation in command (relies on domain)

3. **Infrastructure** (13/15) - Good implementation
   - ✅ Repository implements domain interface
   - ✅ HTTP handler is thin (delegates to commands)
   - ✅ Swagger annotations complete
   - ⚠️  Migration missing index on `phone_number`

4. **SOLID** (12/15) - Mostly compliant
   - ✅ Single Responsibility
   - ✅ Dependency Inversion (uses interfaces)
   - ⚠️  Some functions > 50 lines (refactor recommended)

5. **Security** (12/15) - Good, with gaps
   - ✅ RBAC check in handler
   - ✅ Tenant isolation via RLS
   - ⚠️  Missing BOLA check (no ownership verification)

6. **Testing** (7/10) - Needs improvement
   - ✅ Domain: 100% coverage
   - ⚠️  Application: 65% coverage (target: 80%)
   - ❌ Missing E2E tests

---

## 🔴 P0 Issues (CRITICAL - Fix Before Merge)

None found ✅

---

## 🟡 P1 Issues (Important - Fix in This PR)

1. **Missing BOLA Protection** (Security)
   - File: `infrastructure/http/handlers/contact_handler.go:42`
   - Issue: GetContact() doesn't verify ownership
   - Fix:
     ```go
     // Add ownership check
     if err := h.rbac.CheckOwnership(ctx, contactID, userID); err != nil {
         return c.JSON(403, gin.H{"error": "Forbidden"})
     }
     ```

2. **Low Application Test Coverage** (Testing)
   - File: `internal/application/commands/contact/`
   - Coverage: 65% (target: 80%)
   - Missing tests:
     - `UpdateContactHandler` error cases
     - `DeleteContactHandler` concurrent updates
   - Fix: Add 3 more test cases

---

## ⚪ P2 Issues (Nice to Have - Future PR)

1. **Missing Index** (Performance)
   - File: `infrastructure/database/migrations/XXX_add_contacts.up.sql`
   - Add: `CREATE INDEX idx_contacts_phone ON contacts(phone_number);`

2. **Long Functions** (Code Quality)
   - File: `internal/domain/crm/contact/aggregate.go:87-142` (55 lines)
   - Refactor: Extract `validateContactData()` helper

3. **Email Invariant** (Domain)
   - File: `internal/domain/crm/contact/aggregate.go`
   - Add email format validation in aggregate

---

## 💡 Recommendations

### 1. Immediate Actions (P1)
- Add BOLA check to GetContact handler (5 min)
- Add 3 test cases to reach 80% coverage (15 min)

### 2. Before Next Release (P2)
- Add phone_number index (1 min)
- Refactor long functions (30 min)
- Add email validation invariant (10 min)

### 3. Good Practices to Continue
- Keep using value objects (excellent!)
- Maintain 100% domain test coverage
- Keep handlers thin (great job!)

---

**Reviewer**: meta_code_reviewer v1.0
**Confidence**: High
**Re-Review**: Not needed (unless P1 issues fixed)
```

---

## 🔄 Integration with /add-feature

Automatic review during feature development:

```bash
/add-feature Add Custom Fields aggregate

# After implementation, automatically calls:
/review CustomField --strict --update-p0

# If score < 80%:
# - Show report
# - Ask user: "Fix issues or continue anyway?"
# - Update P0 file with issues to fix
```

---

## 🎯 Use Cases

### 1. Pre-Commit Review
```bash
# In git pre-commit hook
/review --changed-only --fail-below=80 --block-merge
```

### 2. PR Review
```bash
# In GitHub Actions
/review --changed-only --export=json --create-issues
```

### 3. Weekly Code Quality Check
```bash
# Cron job
/review --all --export=html --email-report
```

### 4. Onboarding Code Review
```bash
# New developer's first PR
/review --changed-only --verbose --suggest-fixes
```

### 5. Security Audit
```bash
# Before production deploy
/review --all --security-focus --strict --fail-below=90
```

---

## 🔗 Related Commands

- `/add-feature` - Implement feature (includes review)
- `/analyze` - Full codebase analysis
- `/test-feature` - Run tests
- `make lint` - Traditional linting

---

**Agent Invoked**: `meta_code_reviewer.md`
**Runtime**: 1-5 min (per aggregate), 30-45 min (full codebase)
**Tokens**: 3k-10k (per aggregate), 50k-80k (full codebase)
**Output**: `/tmp/code_review.md` with score and recommendations
**Pass Threshold**: 80% (configurable with `--fail-below=N`)
