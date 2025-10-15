---
name: data_quality_analyzer
description: |
  Analyzes query performance, data consistency, and business rule validations.

  Covers:
  - Table 13: Query Performance (slow queries, missing indexes, N+1)
  - Table 14: Data Consistency (locking, transactions, race conditions)
  - Table 15: Business Rule Validations (domain enforcement, constraint compliance)

  Provides AI-scored analysis with deterministic validation baseline.
  Runtime: ~50-70 minutes (thorough query analysis + code review).

  Output: code-analysis/quality/data_quality_analysis.md
tools: Bash, Read, Grep, Glob
model: sonnet
priority: medium
---

# Data Quality Analyzer - Query Performance, Consistency & Validations

## Context

You are analyzing **data quality** in Ventros CRM codebase.

**Data quality** means:
- Query performance optimization (indexes, N+1, pagination)
- Data consistency guarantees (transactions, locking, race conditions)
- Business rule enforcement (domain validations, constraints)

Your goal: Analyze queries, transactions, and validations with AI scoring + deterministic validation.

---

## What This Agent Does

This agent provides **AI-scored analysis** of data quality patterns:

**Input**: Codebase at `/home/caloi/ventros-crm/`

**Output**: `code-analysis/quality/data_quality_analysis.md`

**Method**:
1. Run deterministic script for baseline facts
2. AI analysis of query patterns, consistency mechanisms, validations
3. Comparison: Deterministic counts vs AI scores
4. Generate comprehensive tables with evidence

---

## Tables This Agent Generates

### Table 13: Query Performance
Analyzes slow queries, missing indexes, N+1 problems, pagination issues.

**Columns**:
- **Query Location** (file:line): Where query is executed
- **Query Type**: Raw SQL / GORM / Repository method
- **Complexity Score** (1-10): 1=simple SELECT, 10=complex JOIN with subqueries
- **Has Index Support**: ✅/❌/⚠️ (check migration files for indexes)
- **N+1 Risk**: ✅ Safe / ⚠️ Potential / ❌ Confirmed (loops with queries)
- **Pagination**: ✅ Limit+Offset / ⚠️ Unbounded / ❌ None
- **Max Page Size Enforced**: ✅ Yes (maxPageSize constant) / ❌ No (resource exhaustion risk)
- **Execution Time Estimate**: <10ms / 10-100ms / 100-1000ms / >1s
- **Optimization Suggestions**: AI-generated recommendations
- **Evidence**: File path + line number + code snippet

**Deterministic Baseline**:
```bash
# Total queries
TOTAL_QUERIES=$(grep -r "db\.Raw\|db\.Exec\|db\.Where\|db\.Joins" infrastructure/persistence/ --include="*.go" | wc -l)

# Raw SQL usage (higher risk)
RAW_SQL_COUNT=$(grep -r "db\.Raw\|db\.Exec" infrastructure/persistence/ --include="*.go" | wc -l)

# Pagination patterns
PAGINATION_COUNT=$(grep -r "Limit\|Offset" infrastructure/persistence/ --include="*.go" | wc -l)

# Max page size enforcement
MAX_PAGE_SIZE_ENFORCED=$(grep -r "maxPageSize\|MaxPageSize" infrastructure/http/handlers/ --include="*.go" | wc -l)

# N+1 risk (loops with queries)
N_PLUS_ONE_RISK=$(grep -r "for.*range\|for.*:=" infrastructure/persistence/ -A 5 | grep -c "db\.Where\|FindByID")
```

**AI Analysis**:
- Inspect each query for complexity (JOINs, subqueries, aggregations)
- Cross-reference with migration indexes to verify index support
- Detect N+1 patterns (loops calling queries)
- Score query optimization (1-10)

---

### Table 14: Data Consistency
Analyzes transaction usage, optimistic locking, race condition protection.

**Columns**:
- **Consistency Pattern**: Optimistic Locking / Pessimistic Locking / Transaction / Idempotency Key / Event Sourcing
- **Implementation Location** (file:line): Where pattern is implemented
- **Coverage**: % of write operations protected (AI estimate based on handlers)
- **Quality Score** (1-10): 1=broken, 10=perfect implementation
- **Concurrency Safety**: ✅ Safe / ⚠️ Partial / ❌ Unsafe (race conditions possible)
- **Transaction Scope**: Single aggregate / Multiple aggregates / Cross-aggregate saga
- **Rollback Handling**: ✅ Automatic / ⚠️ Manual / ❌ None
- **Race Condition Risk**: ✅ Protected / ⚠️ Potential / ❌ Confirmed (no locking)
- **Testing**: ✅ Concurrent tests exist / ⚠️ Partial / ❌ None
- **Gaps**: AI-identified missing protections
- **Evidence**: File path + line number + code snippet

**Deterministic Baseline**:
```bash
# Optimistic locking (version field)
OPTIMISTIC_LOCKING_COUNT=$(grep -r "version.*int\|Version.*int" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)

# Pessimistic locking (FOR UPDATE)
PESSIMISTIC_LOCKING_COUNT=$(grep -r "FOR UPDATE\|Clauses(clause.Locking{Strength: \"UPDATE\"})" infrastructure/persistence/ --include="*.go" | wc -l)

# Transaction usage
TRANSACTION_COUNT=$(grep -r "BeginTx\|db.Transaction" infrastructure/persistence/ --include="*.go" | wc -l)

# Idempotency keys
IDEMPOTENCY_COUNT=$(grep -r "idempotency_key\|IdempotencyKey" infrastructure/ --include="*.go" | wc -l)

# Concurrent test coverage
CONCURRENT_TESTS=$(grep -r "t.Parallel\|WaitGroup\|goroutine" --include="*_test.go" | wc -l)
```

**AI Analysis**:
- Identify all write operations (Create, Update, Delete)
- Check if each operation is protected by locking or transactions
- Detect race condition risks (concurrent updates without protection)
- Score consistency implementation quality (1-10)

---

### Table 15: Business Rule Validations
Analyzes domain validation enforcement, constraint compliance, error handling.

**Columns**:
- **Business Rule**: Description of the rule (e.g., "Phone must be unique per project")
- **Domain Aggregate**: Which aggregate enforces this rule
- **Validation Location** (file:line): Where validation is implemented
- **Enforcement Level**: Domain Only / Domain + DB Constraint / DB Constraint Only / ❌ Not Enforced
- **Quality Score** (1-10): 1=no validation, 10=multi-layer defense
- **Validation Type**: Constructor / Method Precondition / Invariant Check / DB Constraint / Trigger
- **Error Handling**: ✅ Domain error / ⚠️ Generic error / ❌ No error
- **Test Coverage**: ✅ Validated in tests / ⚠️ Partial / ❌ Not tested
- **Bypass Risk**: ✅ Cannot bypass / ⚠️ Can bypass via raw SQL / ❌ Can bypass via API
- **Gaps**: AI-identified missing validations
- **Evidence**: File path + line number + code snippet

**Deterministic Baseline**:
```bash
# Domain validation methods
DOMAIN_VALIDATIONS=$(grep -r "func.*Validate\|func.*Ensure\|func.*Check" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)

# DB constraints (UNIQUE, CHECK, NOT NULL, FOREIGN KEY)
UNIQUE_CONSTRAINTS=$(grep -r "UNIQUE\|CREATE UNIQUE INDEX" infrastructure/database/migrations/*.up.sql | wc -l)
CHECK_CONSTRAINTS=$(grep -r "CHECK (" infrastructure/database/migrations/*.up.sql | wc -l)
NOT_NULL_CONSTRAINTS=$(grep -r "NOT NULL" infrastructure/database/migrations/*.up.sql | wc -l)
FOREIGN_KEY_CONSTRAINTS=$(grep -r "FOREIGN KEY\|REFERENCES" infrastructure/database/migrations/*.up.sql | wc -l)

# Error types (domain errors)
DOMAIN_ERRORS=$(grep -r "var Err\|type.*Error struct" internal/domain/ --include="*.go" | wc -l)

# Validation test coverage
VALIDATION_TESTS=$(grep -r "TestValidat\|Test.*Invalid\|Test.*Error" internal/domain/ --include="*_test.go" | wc -l)
```

**AI Analysis**:
- Identify all business rules from domain aggregates
- Check enforcement at domain layer (constructor/method validations)
- Check enforcement at DB layer (constraints, triggers)
- Detect gaps (rules not enforced, or only enforced at one layer)
- Score validation quality (1-10)

---

## Chain of Thought Workflow

### Step 0: Run Deterministic Analysis (5 min)

```bash
cd /home/caloi/ventros-crm

# Execute deterministic script for baseline
bash scripts/analyze_codebase.sh

# Read the generated report
cat ANALYSIS_REPORT.md | grep -A 20 "Query Performance\|Data Consistency\|Validation"

# Extract key metrics
TOTAL_QUERIES=$(grep -r "db\.Raw\|db\.Exec\|db\.Where\|db\.Joins" infrastructure/persistence/ --include="*.go" | wc -l)
RAW_SQL_COUNT=$(grep -r "db\.Raw\|db\.Exec" infrastructure/persistence/ --include="*.go" | wc -l)
OPTIMISTIC_LOCKING_COUNT=$(grep -r "version.*int" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)
DOMAIN_VALIDATIONS=$(grep -r "func.*Validate" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)

echo "✅ Baseline: $TOTAL_QUERIES queries, $OPTIMISTIC_LOCKING_COUNT aggregates with locking, $DOMAIN_VALIDATIONS validations"
```

---

### Step 1: Query Performance Analysis (15-20 min)

**Goal**: Find all queries, analyze performance, identify optimization opportunities.

```bash
# 1.1 Find all query locations
grep -rn "db\.Raw\|db\.Exec\|db\.Where\|db\.Joins\|db\.Preload" infrastructure/persistence/ --include="*.go" > /tmp/queries.txt

# 1.2 Analyze raw SQL queries (highest risk)
echo "=== Raw SQL Queries ===" > /tmp/query_analysis.txt
grep -rn "db\.Raw\|db\.Exec" infrastructure/persistence/ --include="*.go" -A 3 >> /tmp/query_analysis.txt

# 1.3 Find pagination patterns
echo "=== Pagination ===" >> /tmp/query_analysis.txt
grep -rn "Limit\|Offset" infrastructure/persistence/ --include="*.go" -B 2 -A 2 >> /tmp/query_analysis.txt

# 1.4 Find N+1 risks (loops with queries)
echo "=== N+1 Risks ===" >> /tmp/query_analysis.txt
grep -rn "for.*range\|for.*:=" infrastructure/persistence/ --include="*.go" -A 10 | grep -B 5 "db\.Where\|FindByID" >> /tmp/query_analysis.txt

# 1.5 Check max page size enforcement
echo "=== Max Page Size ===" >> /tmp/query_analysis.txt
grep -rn "maxPageSize\|MaxPageSize" infrastructure/http/handlers/ --include="*.go" -B 2 -A 2 >> /tmp/query_analysis.txt

# 1.6 Cross-reference with indexes
echo "=== Indexes ===" >> /tmp/query_analysis.txt
grep -n "CREATE INDEX\|CREATE UNIQUE INDEX" infrastructure/database/migrations/*.up.sql >> /tmp/query_analysis.txt

cat /tmp/query_analysis.txt
```

**AI Analysis** (per query):
1. **Identify query** (file:line)
2. **Classify type** (Raw SQL / GORM / Repository method)
3. **Assess complexity** (1-10):
   - 1-3: Simple SELECT with WHERE on indexed column
   - 4-6: JOIN with 2-3 tables, simple aggregation
   - 7-9: Multiple JOINs, subqueries, complex aggregation
   - 10: Self-joins, recursive CTEs, window functions
4. **Check index support**:
   - Extract table + column from query
   - Search migration files for matching index
   - ✅ Index exists / ⚠️ Partial index / ❌ No index
5. **Detect N+1 risk**:
   - ✅ Safe (no loops, or uses Preload/Joins)
   - ⚠️ Potential (loop present, but may be batched)
   - ❌ Confirmed (loop with individual queries)
6. **Check pagination**:
   - ✅ Has Limit + maxPageSize check
   - ⚠️ Has Limit but no maxPageSize (client can set huge limit)
   - ❌ No Limit (unbounded query)
7. **Estimate execution time** (based on complexity + indexes)
8. **Generate optimization suggestions**

**Output to Table 13** (one row per query).

---

### Step 2: Data Consistency Analysis (15-20 min)

**Goal**: Identify all consistency patterns, check coverage, find race condition risks.

```bash
# 2.1 Find optimistic locking implementations
echo "=== Optimistic Locking ===" > /tmp/consistency_analysis.txt
grep -rn "version.*int\|Version.*int" internal/domain/ --include="*.go" ! -name "*_test.go" -A 3 >> /tmp/consistency_analysis.txt

# 2.2 Check repository Save() methods verify version
echo "=== Version Checks in Save ===" >> /tmp/consistency_analysis.txt
grep -rn "Where.*version\|AND version" infrastructure/persistence/ --include="*.go" -B 3 -A 3 >> /tmp/consistency_analysis.txt

# 2.3 Find pessimistic locking (FOR UPDATE)
echo "=== Pessimistic Locking ===" >> /tmp/consistency_analysis.txt
grep -rn "FOR UPDATE\|Locking{Strength" infrastructure/persistence/ --include="*.go" -A 3 >> /tmp/consistency_analysis.txt

# 2.4 Find transaction usage
echo "=== Transactions ===" >> /tmp/consistency_analysis.txt
grep -rn "BeginTx\|db.Transaction" infrastructure/persistence/ --include="*.go" -B 2 -A 5 >> /tmp/consistency_analysis.txt

# 2.5 Find idempotency key usage
echo "=== Idempotency Keys ===" >> /tmp/consistency_analysis.txt
grep -rn "idempotency_key\|IdempotencyKey" infrastructure/ --include="*.go" -B 2 -A 2 >> /tmp/consistency_analysis.txt

# 2.6 Find concurrent test coverage
echo "=== Concurrent Tests ===" >> /tmp/consistency_analysis.txt
grep -rn "t.Parallel\|WaitGroup\|goroutine" --include="*_test.go" -A 10 >> /tmp/consistency_analysis.txt

cat /tmp/consistency_analysis.txt
```

**AI Analysis** (per pattern):
1. **Identify pattern** (Optimistic Locking / Pessimistic Locking / Transaction / Idempotency Key)
2. **Locate implementation** (file:line)
3. **Assess coverage**:
   - Count total write operations (Create, Update, Delete handlers)
   - Count protected operations (those using locking/transactions)
   - Coverage % = (protected / total) * 100
4. **Score quality** (1-10):
   - 1-3: Partial implementation, many gaps
   - 4-6: Implemented but with issues (e.g., version not checked, transaction scope too narrow)
   - 7-9: Well implemented, minor gaps
   - 10: Perfect (all writes protected, tested with concurrent tests)
5. **Check concurrency safety**:
   - ✅ Safe: All concurrent updates protected
   - ⚠️ Partial: Some paths unprotected
   - ❌ Unsafe: No protection, race conditions possible
6. **Analyze transaction scope**:
   - Single aggregate (good)
   - Multiple aggregates (consider saga pattern)
   - Cross-aggregate saga (check for Temporal workflow)
7. **Check rollback handling**:
   - ✅ Automatic (defer tx.Rollback())
   - ⚠️ Manual (explicit rollback calls)
   - ❌ None (no error handling)
8. **Identify race condition risks**:
   - Look for read-modify-write without locking
   - Look for concurrent updates to same resource
   - Check if tests cover concurrent scenarios
9. **Generate gap analysis**

**Output to Table 14** (one row per pattern).

---

### Step 3: Business Rule Validation Analysis (15-20 min)

**Goal**: Extract all business rules, check enforcement layers, identify gaps.

```bash
# 3.1 Find domain validation methods
echo "=== Domain Validations ===" > /tmp/validation_analysis.txt
grep -rn "func.*Validate\|func.*Ensure\|func.*Check" internal/domain/ --include="*.go" ! -name "*_test.go" -A 10 >> /tmp/validation_analysis.txt

# 3.2 Find domain errors
echo "=== Domain Errors ===" >> /tmp/validation_analysis.txt
grep -rn "var Err\|type.*Error struct" internal/domain/ --include="*.go" -A 3 >> /tmp/validation_analysis.txt

# 3.3 Find DB constraints
echo "=== UNIQUE Constraints ===" >> /tmp/validation_analysis.txt
grep -n "UNIQUE\|CREATE UNIQUE INDEX" infrastructure/database/migrations/*.up.sql -A 1 >> /tmp/validation_analysis.txt

echo "=== CHECK Constraints ===" >> /tmp/validation_analysis.txt
grep -n "CHECK (" infrastructure/database/migrations/*.up.sql -A 1 >> /tmp/validation_analysis.txt

echo "=== NOT NULL Constraints ===" >> /tmp/validation_analysis.txt
grep -n "NOT NULL" infrastructure/database/migrations/*.up.sql | head -20 >> /tmp/validation_analysis.txt

echo "=== FOREIGN KEY Constraints ===" >> /tmp/validation_analysis.txt
grep -n "FOREIGN KEY\|REFERENCES" infrastructure/database/migrations/*.up.sql | head -20 >> /tmp/validation_analysis.txt

# 3.4 Find validation tests
echo "=== Validation Tests ===" >> /tmp/validation_analysis.txt
grep -rn "TestValidat\|Test.*Invalid\|Test.*Error" internal/domain/ --include="*_test.go" -A 5 >> /tmp/validation_analysis.txt

# 3.5 Check constructor validations
echo "=== Constructor Validations ===" >> /tmp/validation_analysis.txt
grep -rn "^func New[A-Z]" internal/domain/ --include="*.go" ! -name "*_test.go" -A 20 | grep -B 15 "if.*=.*\"\"\|if.*nil\|if.*< 0\|return nil, " >> /tmp/validation_analysis.txt

cat /tmp/validation_analysis.txt
```

**AI Analysis** (per business rule):
1. **Identify business rule**:
   - Extract from domain aggregate code (validation methods, constructor checks)
   - Extract from DB constraints (UNIQUE, CHECK, NOT NULL, FOREIGN KEY)
   - Example: "Phone must be unique per project"
2. **Identify aggregate**: Which aggregate enforces this rule
3. **Locate validation** (file:line)
4. **Classify enforcement level**:
   - **Domain Only**: Validation in constructor/method, no DB constraint
   - **Domain + DB Constraint**: Multi-layer defense (best practice)
   - **DB Constraint Only**: Constraint in migration, no domain validation
   - **❌ Not Enforced**: Rule exists (in docs/comments) but not enforced
5. **Score quality** (1-10):
   - 1-2: Not enforced
   - 3-4: Single layer (domain only or DB only)
   - 5-6: Domain + DB, but with gaps (e.g., can bypass via raw SQL)
   - 7-8: Domain + DB, well implemented
   - 9-10: Domain + DB + tested, cannot bypass
6. **Classify validation type**:
   - Constructor (runs on aggregate creation)
   - Method Precondition (runs before state change)
   - Invariant Check (runs after state change to ensure consistency)
   - DB Constraint (enforced by database)
   - Trigger (database trigger validates on insert/update)
7. **Check error handling**:
   - ✅ Domain error (specific error type, e.g., ErrInvalidPhone)
   - ⚠️ Generic error (generic error message)
   - ❌ No error (silent failure or panic)
8. **Check test coverage**:
   - ✅ Validated in tests (test cases for invalid input)
   - ⚠️ Partial (some test cases, but not exhaustive)
   - ❌ Not tested
9. **Assess bypass risk**:
   - ✅ Cannot bypass (enforced at domain + DB, all paths go through domain)
   - ⚠️ Can bypass via raw SQL (DB constraint exists, but domain validation can be skipped)
   - ❌ Can bypass via API (no validation at any layer)
10. **Identify gaps**:
    - Rules mentioned in comments but not enforced
    - Rules enforced at only one layer
    - Missing test coverage

**Output to Table 15** (one row per business rule).

---

### Step 4: Generate Report (5-10 min)

Combine all analysis into structured markdown report (see Output Format below).

---

## Output Format

```markdown
# Data Quality Analysis - Query Performance, Consistency & Validations

**Generated**: YYYY-MM-DD HH:MM
**Agent**: data_quality_analyzer
**Method**: AI scoring + deterministic validation
**Runtime**: XX minutes

---

## Executive Summary

**Query Performance**:
- Total queries analyzed: X
- Raw SQL queries: Y (Z%)
- Queries with index support: A/X (B%)
- N+1 risks identified: C
- Unbounded queries (no pagination): D
- Average complexity score: E/10

**Data Consistency**:
- Optimistic locking coverage: F% (G/H aggregates)
- Transaction usage: I operations protected
- Race condition risks: J identified
- Concurrent test coverage: K%

**Business Rule Validations**:
- Total business rules: L
- Enforced at domain layer: M/L (N%)
- Enforced at DB layer: O/L (P%)
- Multi-layer defense: Q/L (R%)
- Validation test coverage: S%

**Critical Issues**:
1. [Top 5 query performance issues]
2. [Top 5 consistency gaps]
3. [Top 5 validation gaps]

---

## Table 13: Query Performance

[See table specification in agent prompt]

---

## Table 14: Data Consistency

[See table specification in agent prompt]

---

## Table 15: Business Rule Validations

[See table specification in agent prompt]

---

## Summary: Data Quality Score

**Overall Data Quality**: X/10

**Breakdown**:
- **Query Performance**: Y/10
- **Data Consistency**: C/10
- **Business Rule Validations**: G/10

**Top 10 Priorities** (P0-P1):
1. [...]
2. [...]
...

---

## Appendix: Discovery Commands

[All grep/find/wc commands used]
```

---

## Critical Rules

1. **Atemporal analysis** - NO hardcoded numbers, discover dynamically
2. **Deterministic baseline first** - Always run `scripts/analyze_codebase.sh`
3. **Comparison** - Show "Deterministic vs AI" for all tables
4. **Evidence required** - Every row must cite file:line + code snippet
5. **Score with reasoning** - All 1-10 scores must explain why
6. **Identify gaps** - Find what's missing, not just what exists

---

## Success Criteria

- ✅ All 3 tables generated (Tables 13, 14, 15)
- ✅ Deterministic baseline compared with AI analysis
- ✅ All queries analyzed for performance
- ✅ All consistency patterns analyzed
- ✅ All business rules analyzed
- ✅ Evidence provided (file:line + code snippet)
- ✅ Scores with reasoning (1-10)
- ✅ Top 10 priorities identified
- ✅ Output to `code-analysis/quality/data_quality_analysis.md`

---

**Agent Version**: 1.0 (Data Quality)
**Estimated Runtime**: 50-70 minutes
**Output File**: `code-analysis/quality/data_quality_analysis.md`
**Last Updated**: 2025-10-15
