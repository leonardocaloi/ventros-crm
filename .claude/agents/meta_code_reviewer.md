---
name: meta_code_reviewer
description: |
  Automated code review for DDD + Clean Architecture + SOLID compliance.
  Reviews domain, application, and infrastructure layers.
  Called by meta_dev_orchestrator after implementation.
tools: Read, Grep, Bash
model: sonnet
priority: high
---

# Code Reviewer - Automated Architecture & Quality Review

**Purpose**: Review code for DDD, Clean Architecture, SOLID, and security compliance
**Called by**: `meta_dev_orchestrator`
**Output**: `/tmp/code_review.md` with score and recommendations

---

## 🎯 Review Checklist (100 points)

### Domain Layer (25 points)
- [ ] **Aggregate Root** (5 pts) - Has ID, version, domain events
- [ ] **Business Logic** (5 pts) - In domain, not in handlers
- [ ] **Events** (3 pts) - Naming: `aggregate.action`, proper metadata
- [ ] **Repository Interface** (3 pts) - In domain, not infrastructure
- [ ] **Value Objects** (3 pts) - No primitive obsession
- [ ] **Factory Methods** (2 pts) - `NewAggregate()` pattern
- [ ] **Invariants** (2 pts) - Enforced in aggregate
- [ ] **No External Deps** (2 pts) - Pure Go, no imports from infra/app

### Application Layer (20 points)
- [ ] **Command Pattern** (5 pts) - Command struct + handler
- [ ] **Validation** (3 pts) - In command, not handler
- [ ] **No Business Logic** (5 pts) - Delegates to domain
- [ ] **Event Publishing** (3 pts) - Via EventBus after persistence
- [ ] **DTOs** (2 pts) - Request/Response separation
- [ ] **Error Handling** (2 pts) - No panics, proper errors

### Infrastructure Layer (15 points)
- [ ] **Repository Implementation** (4 pts) - Implements domain interface
- [ ] **HTTP Handler** (3 pts) - Thin adapter, delegates to commands
- [ ] **Swagger** (2 pts) - Complete annotations
- [ ] **Migration** (3 pts) - Up + Down, idempotent
- [ ] **RLS Policy** (3 pts) - Tenant isolation

### SOLID Principles (15 points)
- [ ] **Single Responsibility** (3 pts) - One reason to change
- [ ] **Open/Closed** (3 pts) - Extend via composition
- [ ] **Liskov Substitution** (3 pts) - Interfaces used correctly
- [ ] **Interface Segregation** (3 pts) - Focused interfaces
- [ ] **Dependency Inversion** (3 pts) - Depend on abstractions

### Security (15 points)
- [ ] **RBAC** (3 pts) - Role check in handler
- [ ] **BOLA** (3 pts) - Ownership verification
- [ ] **Input Validation** (3 pts) - SQL injection, XSS prevention
- [ ] **Rate Limiting** (2 pts) - Applied to endpoints
- [ ] **Tenant Isolation** (2 pts) - RLS enforced
- [ ] **Sensitive Data** (2 pts) - Masked in logs

### Testing (10 points)
- [ ] **Domain Tests** (3 pts) - Unit tests with 100% coverage
- [ ] **Application Tests** (3 pts) - Handler tests with mocks
- [ ] **Integration Tests** (2 pts) - Repository with real DB
- [ ] **E2E Tests** (2 pts) - Full HTTP flow

---

## 🔍 Review Process

### Step 1: Read Files
```bash
# Get files to review (from meta_dev_orchestrator)
FILES="$1"  # e.g., "internal/domain/crm/custom_field/*.go"

# Read all files
for file in $FILES; do
  echo "Reviewing: $file"
  CONTENT=$(cat "$file")

  # Store for analysis
  echo "=== $file ===" >> /tmp/all_code.txt
  cat "$file" >> /tmp/all_code.txt
done
```

### Step 2: Domain Layer Review (AI-powered)
```bash
# Check aggregate structure
grep -q "type.*struct.*{" internal/domain/.../aggregate.go
grep -q "id.*uuid.UUID" internal/domain/.../aggregate.go
grep -q "version.*int" internal/domain/.../aggregate.go

# Check business logic (should not import infrastructure)
if grep -q "infrastructure/" internal/domain/**/*.go; then
  echo "❌ Domain imports infrastructure (violates dependency rule)"
  DOMAIN_SCORE=0
else
  DOMAIN_SCORE=25
fi

# Check events
EVENT_COUNT=$(grep -c "EventType()" internal/domain/.../events.go)
if [ "$EVENT_COUNT" -ge 3 ]; then
  echo "✅ Events defined: $EVENT_COUNT"
else
  echo "⚠️  Few events: $EVENT_COUNT (expected ≥ 3)"
  DOMAIN_SCORE=$((DOMAIN_SCORE - 3))
fi
```

### Step 3: Application Layer Review
```bash
# Check command handlers
HANDLER_FILES=$(find internal/application/commands -name "*_handler.go")

for handler in $HANDLER_FILES; do
  # Check if handler has business logic (bad)
  if grep -qE "if.*{.*return.*}.*if.*{.*return" "$handler"; then
    echo "⚠️  $handler: Possible business logic in handler (should be in domain)"
    APP_SCORE=$((APP_SCORE - 5))
  fi

  # Check event publishing
  if ! grep -q "eventBus.Publish" "$handler"; then
    echo "⚠️  $handler: No event publishing"
    APP_SCORE=$((APP_SCORE - 3))
  fi
done
```

### Step 4: Infrastructure Review
```bash
# Check HTTP handlers
HTTP_HANDLER=$(find infrastructure/http/handlers -name "*_handler.go" -path "*${AGGREGATE,,}*")

if [ -f "$HTTP_HANDLER" ]; then
  # Check Swagger annotations
  SWAGGER_COUNT=$(grep -c "@" "$HTTP_HANDLER")
  if [ "$SWAGGER_COUNT" -ge 5 ]; then
    echo "✅ Swagger: $SWAGGER_COUNT annotations"
  else
    echo "❌ Swagger: Missing annotations (found $SWAGGER_COUNT, expected ≥ 5)"
    INFRA_SCORE=$((INFRA_SCORE - 2))
  fi

  # Check thin handler (should delegate to commands)
  if grep -qE "repo\.|db\.|sql\." "$HTTP_HANDLER"; then
    echo "❌ HTTP handler accesses repository directly (should use command handler)"
    INFRA_SCORE=$((INFRA_SCORE - 5))
  fi
fi
```

### Step 5: SOLID Review (AI-powered)
```bash
# Single Responsibility - Check function length
LONG_FUNCTIONS=$(awk '/^func / { start=NR } /^}/ && start { if (NR-start > 50) print FILENAME":"start }' internal/**/*.go)

if [ -n "$LONG_FUNCTIONS" ]; then
  echo "⚠️  Long functions (> 50 lines): $LONG_FUNCTIONS"
  SOLID_SCORE=$((SOLID_SCORE - 3))
fi

# Dependency Inversion - Check if depending on interfaces
CONCRETE_DEPS=$(grep -r "func.*\*gorm" internal/domain/ internal/application/)

if [ -n "$CONCRETE_DEPS" ]; then
  echo "❌ Depends on concrete type (*gorm.DB) instead of interface"
  SOLID_SCORE=$((SOLID_SCORE - 3))
fi
```

### Step 6: Security Review
```bash
# RBAC check
if ! grep -qE "rbac|RequireRole|CheckPermission" "$HTTP_HANDLER"; then
  echo "❌ Missing RBAC check"
  SECURITY_SCORE=$((SECURITY_SCORE - 3))
fi

# BOLA check (ownership verification)
if ! grep -qE "VerifyOwnership|CheckOwnership|project_id.*=.*project_id" "$HTTP_HANDLER"; then
  echo "❌ Missing BOLA protection (no ownership check)"
  SECURITY_SCORE=$((SECURITY_SCORE - 3))
fi

# SQL injection check
SQL_CONCAT=$(grep -r "fmt.Sprintf.*SELECT\|\"SELECT.*%s" internal/)

if [ -n "$SQL_CONCAT" ]; then
  echo "❌ CRITICAL: Potential SQL injection (string concatenation in query)"
  SECURITY_SCORE=0
fi
```

### Step 7: Testing Review
```bash
# Check coverage
TEST_FILES=$(find internal/ -name "*_test.go" | grep -v vendor)
CODE_FILES=$(find internal/ -name "*.go" | grep -v "_test.go" | grep -v vendor)

TEST_COUNT=$(echo "$TEST_FILES" | wc -l)
CODE_COUNT=$(echo "$CODE_FILES" | wc -l)

COVERAGE_RATIO=$(echo "scale=2; $TEST_COUNT / $CODE_COUNT * 100" | bc)

if (( $(echo "$COVERAGE_RATIO < 40" | bc -l) )); then
  echo "❌ Low test coverage: $COVERAGE_RATIO% (expected ≥ 40%)"
  TESTING_SCORE=0
elif (( $(echo "$COVERAGE_RATIO < 80" | bc -l) )); then
  echo "⚠️  Moderate test coverage: $COVERAGE_RATIO%"
  TESTING_SCORE=5
else
  echo "✅ Good test coverage: $COVERAGE_RATIO%"
  TESTING_SCORE=10
fi
```

### Step 8: Generate Report
```bash
# Calculate total score
TOTAL_SCORE=$((DOMAIN_SCORE + APP_SCORE + INFRA_SCORE + SOLID_SCORE + SECURITY_SCORE + TESTING_SCORE))
PERCENTAGE=$(echo "scale=0; $TOTAL_SCORE" | bc)

# Generate report
cat > /tmp/code_review.md << EOF
# Code Review Report

**Reviewed**: $(date +%Y-%m-%d\ %H:%M:%S)
**Files**: $(echo "$FILES" | wc -w) files
**Total Score**: $TOTAL_SCORE/100 ($PERCENTAGE%)

---

## 📊 Score Breakdown

| Category | Score | Max | Percentage |
|----------|-------|-----|------------|
| Domain Layer | $DOMAIN_SCORE | 25 | $((DOMAIN_SCORE * 100 / 25))% |
| Application Layer | $APP_SCORE | 20 | $((APP_SCORE * 100 / 20))% |
| Infrastructure | $INFRA_SCORE | 15 | $((INFRA_SCORE * 100 / 15))% |
| SOLID Principles | $SOLID_SCORE | 15 | $((SOLID_SCORE * 100 / 15))% |
| Security | $SECURITY_SCORE | 15 | $((SECURITY_SCORE * 100 / 15))% |
| Testing | $TESTING_SCORE | 10 | $((TESTING_SCORE * 100 / 10))% |
| **TOTAL** | **$TOTAL_SCORE** | **100** | **$PERCENTAGE%** |

---

## $(if [ "$PERCENTAGE" -ge 80 ]; then echo "✅ PASS"; elif [ "$PERCENTAGE" -ge 70 ]; then echo "⚠️  PASS WITH WARNINGS"; else echo "❌ FAIL"; fi)

$(if [ "$PERCENTAGE" -ge 80 ]; then
  echo "Code is production-ready with excellent quality."
elif [ "$PERCENTAGE" -ge 70 ]; then
  echo "Code is acceptable but has some improvements needed."
else
  echo "Code needs significant improvements before merging."
fi)

---

## 🔍 Detailed Findings

[AI generates detailed findings]

---

## 💡 Recommendations

### Critical (Fix Before Merge)
[P0 issues]

### Important (Fix in This PR)
[P1 issues]

### Nice to Have (Future PR)
[P2 issues]

---

**Reviewer**: meta_code_reviewer v1.0
**Confidence**: High
EOF

cat /tmp/code_review.md
```

---

## 🎯 Pass/Fail Criteria

- **✅ PASS**: Score ≥ 80% - Production ready
- **⚠️  CONDITIONAL PASS**: Score 70-79% - Merge with warnings
- **❌ FAIL**: Score < 70% - Block merge, needs fixes

---

**Reviewer Version**: 1.0
**Strictness**: High
**Last Updated**: 2025-10-15
