---
name: test-feature
description: Run tests for specific feature/aggregate with real-time execution and coverage reports
---

# Test Feature Command

**Real-time test execution** for specific features, aggregates, or layers with coverage analysis.

**NEW**: Actually RUNS `go test` and shows real results!

---

## 🎯 What This Does

Executes tests with intelligent analysis integration:

1. **🆕 Loads Analysis Context** - Reads `.claude/analysis/testing.json` to see known gaps
2. **🆕 Uses Domain Context** - Reads `domain_model.json` to understand structure
3. **Runs Go Tests** - Actually executes `go test` with real-time output
4. **Generates Coverage** - HTML coverage reports
5. **Identifies Gaps** - Shows untested code (enhanced with analysis)
6. **Updates P0** - Tracks test failures in P0 file
7. **Updates Agent State** - Shares test results with other agents

**NEW**: If you ran `/pre-analyze`, this command will use those results to:
- Show missing tests identified in analysis
- Recommend which tests to write next
- Compare current coverage vs baseline
- Highlight critical gaps (P0 areas with low coverage)

---

## 🚀 Usage

### Test Specific Aggregate
```bash
/test-feature Contact

# Runs:
# - go test ./internal/domain/crm/contact/...
# - go test ./internal/application/commands/contact/...
# - go test ./infrastructure/persistence/gorm_contact_repository_test.go
# - go test ./infrastructure/http/handlers/contact_handler_test.go
```

### Test Specific Layer
```bash
/test-feature --layer=domain

# Runs all domain tests:
# go test ./internal/domain/...
```

### With Parameters
```bash
/test-feature Contact --coverage --realtime --update-p0

/test-feature --layer=domain --verbose --fail-fast

/test-feature Campaign --integration-only --verbose

/test-feature --all --coverage --export=html
```

---

## 🎛️ Available Parameters

### Scope Control
- `<AGGREGATE_NAME>` - Test specific aggregate (e.g., `Contact`, `Campaign`)
- `--layer=LAYER` - Test specific layer: `domain`, `application`, `infrastructure`, `all`
- `--bounded-context=NAME` - Test bounded context: `crm`, `automation`, `core`
- `--all` - Run all tests (equivalent to `make test`)

### Test Type
- `--unit-only` - Only unit tests (fast, no DB)
- `--integration-only` - Only integration tests (requires DB)
- `--e2e-only` - Only E2E tests (requires API running)
- *(Default: All test types)*

### Execution Control
- `--realtime` - Stream test output in real-time (default: true)
- `--verbose` / `-v` - Verbose test output (`go test -v`)
- `--fail-fast` - Stop on first failure
- `--parallel=N` - Run tests in parallel (N workers)
- `--timeout=DURATION` - Test timeout (default: 10m)

### Coverage
- `--coverage` - Generate coverage report
- `--coverage-html` - Generate HTML coverage report (opens in browser)
- `--coverage-target=N` - Fail if coverage < N% (default: 82)
- `--show-uncovered` - List all uncovered lines

### Output Control
- `--export=FORMAT` - Export results: `json`, `junit`, `html`
- `--output=PATH` - Custom output path
- `--quiet` - Only show pass/fail summary
- `--update-p0` - Update P0 file with failures

### Filtering
- `--run=REGEX` - Run tests matching regex (e.g., `--run=TestCreate`)
- `--skip=REGEX` - Skip tests matching regex
- `--changed-only` - Only test files changed in git

---

## 🧠 Analysis Integration (NEW)

If you ran `/pre-analyze`, this command automatically loads context:

### Analysis Files Used

1. **`.claude/analysis/testing.json`**:
   - Known test gaps
   - Coverage baseline per aggregate
   - Missing test categories (unit, integration, e2e)
   - Recommended test priorities

2. **`.claude/analysis/domain_model.json`**:
   - Aggregate structure
   - Events to test
   - Value objects needing validation tests
   - Repository interfaces to test

3. **`.claude/analysis/api.json`**:
   - Endpoints to test
   - Missing E2E tests
   - BOLA checks needed

### How It Enhances Testing

**Before Analysis** (basic):
```bash
/test-feature Contact --coverage
# Output:
# ✅ Domain: 15/15 tests passed (100%)
# ⚠️  Application: 8/10 tests passed (85%)
```

**After Analysis** (`/pre-analyze` was run):
```bash
/test-feature Contact --coverage
# Output:
# 📊 Loading test analysis context...
# ✅ Found pre-analysis (mode: quick, age: 2h ago)
#
# 📋 Known Gaps from Analysis:
#   - Missing: TestContact_MergeContacts (concurrency test)
#   - Missing: TestContact_Validation (edge cases)
#   - Coverage gap: contact/aggregate.go:142-145 (error handling)
#
# 🧪 Running Tests...
# ✅ Domain: 15/15 tests passed (100%)
# ⚠️  Application: 8/10 tests passed (85%)
#
# 🔍 Gap Analysis:
#   Priority 1 (P0): Add TestContact_MergeContacts (concurrent updates)
#   Priority 2: Cover error handling in aggregate.go:142-145
#   Priority 3: Add edge case validation tests
#
# 💡 Recommendation: Write 2 missing tests to reach 95% coverage
```

### Analysis-Driven Test Recommendations

The command uses analysis to prioritize what tests to write:

| Priority | Criteria | Example |
|----------|----------|---------|
| **P0** | Security-critical + <80% coverage | Auth middleware, RBAC checks |
| **P1** | Core domain + <90% coverage | Contact.Merge(), Session.ConsolidateWith() |
| **P2** | Application layer + <85% coverage | Command handlers |
| **P3** | Infrastructure + <70% coverage | Repositories, HTTP handlers |

---

## 📋 Parameter Examples

### Example 1: Test Aggregate with Coverage
```bash
/test-feature Contact --coverage --realtime

# Output:
# 🧪 Testing Contact Aggregate
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
# 📦 Domain Layer Tests
# Running: go test -v -cover ./internal/domain/crm/contact/...
#
# === RUN   TestNewContact
# --- PASS: TestNewContact (0.00s)
# === RUN   TestContact_UpdateName
# --- PASS: TestContact_UpdateName (0.00s)
# === RUN   TestContact_AddTag
# --- PASS: TestContact_AddTag (0.00s)
# [... 12 more tests ...]
# PASS
# coverage: 100.0% of statements
# ok      internal/domain/crm/contact     0.125s  coverage: 100.0% of statements
#
# ✅ Domain: 15/15 tests passed (100% coverage)
#
# 📦 Application Layer Tests
# Running: go test -v -cover ./internal/application/commands/contact/...
#
# === RUN   TestCreateContactHandler_Handle
# --- PASS: TestCreateContactHandler_Handle (0.05s)
# === RUN   TestUpdateContactHandler_Handle
# --- PASS: TestUpdateContactHandler_Handle (0.03s)
# [... 8 more tests ...]
# PASS
# coverage: 85.4% of statements
# ok      internal/application/commands/contact   0.234s  coverage: 85.4% of statements
#
# ✅ Application: 10/10 tests passed (85.4% coverage)
#
# 📦 Infrastructure Layer Tests
# Running: go test -v -cover ./infrastructure/persistence/gorm_contact_repository_test.go
#
# === RUN   TestGormContactRepository_Save
# --- PASS: TestGormContactRepository_Save (0.12s)
# === RUN   TestGormContactRepository_FindByID
# --- PASS: TestGormContactRepository_FindByID (0.08s)
# [... 5 more tests ...]
# PASS
# coverage: 78.2% of statements
# ok      infrastructure/persistence      0.456s  coverage: 78.2% of statements
#
# ✅ Infrastructure: 7/7 tests passed (78.2% coverage)
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📊 SUMMARY
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# Total Tests: 32/32 passed ✅
# Total Coverage: 88.7% (target: 82%) ✅
# Duration: 0.815s
# Status: PASS ✅
#
# Coverage report: /tmp/contact_coverage.html
```

### Example 2: Fast Domain-Only Test
```bash
/test-feature --layer=domain --quiet

# Output:
# ✅ Domain: 61/61 tests passed (95.2% coverage) [1.2s]
```

### Example 3: Integration Tests (Requires DB)
```bash
/test-feature Contact --integration-only --verbose

# This will:
# 1. Check if PostgreSQL is running (make infra)
# 2. Run integration tests with real DB
# 3. Show detailed output
# 4. Clean up test data
```

### Example 4: Find Failing Tests
```bash
/test-feature Campaign --fail-fast --verbose --update-p0

# This will:
# 1. Run Campaign tests
# 2. Stop on first failure
# 3. Show verbose output
# 4. Add failure to P0 file:
#
#    ### Branch: `main`
#    #### Test Failures:
#    - ❌ TestCampaign_Activate: panic: runtime error (line 45)
#    - Next: Fix panic in Campaign.Activate()
```

### Example 5: Coverage Gap Analysis
```bash
/test-feature Contact --coverage-html --show-uncovered

# This will:
# 1. Generate HTML coverage report
# 2. Open in browser
# 3. List uncovered lines:
#
#    Uncovered Lines:
#    - internal/domain/crm/contact/aggregate.go:142-145 (error handling)
#    - internal/domain/crm/contact/aggregate.go:201 (edge case)
```

### Example 6: Test All Changes (Git Integration)
```bash
/test-feature --changed-only --coverage --realtime

# This will:
# 1. git diff --name-only
# 2. Find all *_test.go files for changed files
# 3. Run only those tests
# 4. Fast feedback loop for active development
```

### Example 7: CI Integration (Strict Mode)
```bash
/test-feature --all --coverage --coverage-target=82 --export=junit --fail-fast

# This will:
# 1. Run all tests
# 2. Generate JUnit XML (for CI)
# 3. Fail if coverage < 82%
# 4. Exit 1 on any failure (blocks merge)
```

---

## 🔄 Real-Time Execution Flow

```bash
# User runs:
/test-feature Contact --coverage --realtime --update-p0

# System does:
1. Update P0 file: "🔵 Testing Contact aggregate..."
2. Run: go test -v -cover ./internal/domain/crm/contact/...
3. Stream output to user in real-time
4. Capture coverage percentage
5. Update AGENT_STATE.json with results
6. Run: go test -v -cover ./internal/application/commands/contact/...
7. Stream output
8. Run: go test -v -cover ./infrastructure/.../contact...
9. Stream output
10. Generate coverage report
11. Update P0 file with results:
    - ✅ Domain: 15/15 (100%)
    - ✅ Application: 10/10 (85.4%)
    - ✅ Infrastructure: 7/7 (78.2%)
    - Overall: 88.7% ✅
12. Clean up P0 (remove "Testing..." entry)
```

---

## 📊 Output Formats

### Console Output (Default)
Real-time streaming of `go test -v` output

### JSON Export (`--export=json`)
```json
{
  "aggregate": "Contact",
  "timestamp": "2025-10-16T10:30:00Z",
  "duration_seconds": 0.815,
  "summary": {
    "total_tests": 32,
    "passed": 32,
    "failed": 0,
    "skipped": 0,
    "coverage_percentage": 88.7
  },
  "layers": {
    "domain": {
      "tests": 15,
      "passed": 15,
      "coverage": 100.0,
      "duration": 0.125
    },
    "application": {
      "tests": 10,
      "passed": 10,
      "coverage": 85.4,
      "duration": 0.234
    },
    "infrastructure": {
      "tests": 7,
      "passed": 7,
      "coverage": 78.2,
      "duration": 0.456
    }
  },
  "failures": [],
  "uncovered_lines": [
    "internal/domain/crm/contact/aggregate.go:142-145",
    "internal/domain/crm/contact/aggregate.go:201"
  ]
}
```

### JUnit XML (`--export=junit`) - For CI
```xml
<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="Contact" tests="32" failures="0" time="0.815">
    <testcase name="TestNewContact" classname="domain" time="0.001"/>
    <testcase name="TestContact_UpdateName" classname="domain" time="0.002"/>
    ...
  </testsuite>
</testsuites>
```

---

## 🔗 Integration with /add-feature

Real-time testing during feature development:

```bash
/add-feature Add Broadcast aggregate --run-tests-realtime

# This will:
# 1. Implement domain layer
# 2. Call: /test-feature Broadcast --layer=domain --coverage --realtime
# 3. Show results immediately
# 4. If tests fail: ask user to review before continuing
# 5. Implement application layer
# 6. Call: /test-feature Broadcast --layer=application --coverage --realtime
# 7. ... and so on
```

---

## 🎯 Use Cases

### 1. TDD Workflow
```bash
# Write failing test first
# Then implement
/test-feature Contact --run=TestContact_Merge --fail-fast

# Iterate until pass
```

### 2. Regression Testing
```bash
# After bug fix
/test-feature Contact --coverage --coverage-target=90

# Ensure no coverage drop
```

### 3. CI Pipeline
```bash
# In GitHub Actions
/test-feature --all --export=junit --coverage-target=82
```

### 4. Quick Smoke Test
```bash
# Before committing
/test-feature --changed-only --fail-fast
```

### 5. Deep Dive Investigation
```bash
# Understand test failures
/test-feature Campaign --verbose --run=TestCampaign_Activate
```

---

## 🔗 Related Commands

- `/add-feature` - Implement feature with tests
- `/review` - Code review
- `/analyze` - Full codebase analysis
- `make test` - Traditional make-based testing

---

**Execution**: REAL - Actually runs `go test` commands
**Runtime**: 0.1s - 10min (depends on scope)
**Tokens**: 2k-5k (just for parsing and reporting)
**Output**: Real-time `go test` output + reports
**P0 Integration**: Updates `.claude/P0_ACTIVE_WORK.md` with results
**Agent State**: Updates `.claude/AGENT_STATE.json` with test results
