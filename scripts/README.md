# 🔍 Ventros CRM - Deterministic Code Analysis Tools

This directory contains tools for **factual, code-based analysis** of the Ventros CRM codebase.

Unlike subjective AI scores (like "Backend: 9.0/10"), these tools extract **measurable metrics** directly from the code using static analysis and AST parsing.

---

## 📊 Available Tools

### 1. `analyze_codebase.sh` - Quick Analysis (Bash)

**Purpose**: Fast overview using grep, find, and basic pattern matching.

**Runtime**: ~2-3 minutes (skips test coverage by default)

**Generates**: `ANALYSIS_REPORT.md`

**Usage**:
```bash
./scripts/analyze_codebase.sh
```

**What it analyzes**:
- ✅ Codebase structure (files, lines, directories)
- ✅ DDD patterns (aggregates, events, repositories)
- ✅ Optimistic locking coverage (version fields)
- ✅ CQRS separation (commands vs queries)
- ✅ Event-driven architecture (events, outbox)
- ✅ Persistence layer (repositories, migrations, RLS)
- ✅ HTTP layer (handlers, middleware, Swagger)
- ✅ Security analysis (BOLA, SQL injection, resource exhaustion)
- ⚠️  Test coverage (optional, can timeout)
- ✅ AI/ML features (implemented vs planned)

---

### 2. `deep_analyzer.go` - Deep AST Analysis (Go)

**Purpose**: Precise analysis using Go's AST (Abstract Syntax Tree) parser.

**Runtime**: ~30 seconds

**Generates**: `DEEP_ANALYSIS_REPORT.md`

**Usage**:
```bash
go run scripts/deep_analyzer.go
```

**What it analyzes**:
- ✅ **Optimistic locking**: Finds exact aggregates with/without `version int` field
- ✅ **Clean Architecture violations**: Detects domain layer importing infrastructure
- ✅ **Security issues**: Handlers without `tenant_id` or `user_id` checks (BOLA)
- ✅ **CQRS patterns**: Command handlers vs Query handlers
- ✅ **Domain events**: Counts events per aggregate
- ✅ **Repository interfaces**: Finds all repository definitions
- ✅ **SQL injection risks**: Files using `db.Raw()` or `db.Exec()`

**Why it's better**:
- **More accurate**: Parses Go AST, not just text matching
- **Fewer false positives**: Understands code structure
- **Faster**: No external tools (grep, find)

---

## 📋 Generated Reports

### `ANALYSIS_REPORT.md` (Bash script)

```markdown
## 2. 🏗️ DOMAIN-DRIVEN DESIGN (DDD)

| Metric | Count | Status |
|--------|-------|--------|
| Aggregates with optimistic locking | 19 / 30 | ⚠️  |

**Aggregates MISSING version field** (🔴 HIGH PRIORITY):
- 🔴 `crm/message`
- 🔴 `crm/channel`
- 🔴 `crm/note`
...
```

### `DEEP_ANALYSIS_REPORT.md` (Go analyzer)

```markdown
## 4. 🔒 SECURITY ANALYSIS

### API1:2023 - Broken Object Level Authorization (BOLA)

**🔴 143 handlers without tenant_id check**:

- 🔴 `CreateContact (in contact_handler.go)`
- 🔴 `UpdateContact (in contact_handler.go)`
- 🔴 `DeleteContact (in contact_handler.go)`
...
```

---

## 🎯 Use Cases

### Before Production Deployment

```bash
# Run both analyses
./scripts/analyze_codebase.sh
go run scripts/deep_analyzer.go

# Review security issues
grep -A 10 "SECURITY ANALYSIS" DEEP_ANALYSIS_REPORT.md

# Check optimistic locking coverage
grep -A 20 "Optimistic Locking" ANALYSIS_REPORT.md
```

### During Code Review

```bash
# Check if new aggregate has version field
go run scripts/deep_analyzer.go
grep "your_aggregate_name" DEEP_ANALYSIS_REPORT.md

# Verify handler has tenant_id check
grep "YourHandlerName" DEEP_ANALYSIS_REPORT.md
```

### For Architecture Documentation

```bash
# Generate metrics for README
./scripts/analyze_codebase.sh
head -50 ANALYSIS_REPORT.md  # Show executive summary
```

---

## 📊 Key Metrics Tracked

### 1. Optimistic Locking Coverage

**Why it matters**: Prevents data corruption in concurrent updates.

**Current state** (from analysis):
- ✅ 13 aggregates WITH `version int`
- 🔴 20 aggregates WITHOUT `version int`
- 📊 Coverage: **39.4%**
- 🎯 Target: **100%**

### 2. BOLA Protection (API Security)

**Why it matters**: OWASP API Security #1 risk - unauthorized data access.

**Current state**:
- ✅ 35 handlers WITH `tenant_id` check
- 🔴 143 handlers WITHOUT `tenant_id` check
- 📊 Coverage: **19.6%**
- 🎯 Target: **100%**

### 3. Clean Architecture Compliance

**Why it matters**: Domain layer must not depend on infrastructure.

**Current state**:
- ✅ **0 violations detected**
- Domain layer correctly isolated

### 4. Test Coverage

**Why it matters**: Confidence in refactoring and changes.

**Current state**:
- 📊 82% overall coverage (from `go test -cover`)
- ✅ Above 80% target

### 5. CQRS Separation

**Why it matters**: Read/write separation for scalability.

**Current state**:
- ✅ 18 Command handlers
- ✅ 20 Query handlers
- ✅ Properly separated

---

## 🔧 Customizing the Analysis

### Add New Check to Bash Script

Edit `analyze_codebase.sh`:

```bash
# Example: Count handlers with rate limiting
HANDLERS_WITH_RATELIMIT=$(count_matches "RateLimit" infrastructure/http/handlers)

cat >> "$REPORT_FILE" <<EOF
### Rate Limiting

| Metric | Count |
|--------|-------|
| Handlers with rate limiting | $HANDLERS_WITH_RATELIMIT |
EOF
```

### Add New Check to Go Analyzer

Edit `deep_analyzer.go`:

```go
// Example: Find handlers without logging
func checkHandlerLogging(node *ast.File, filePath string, result *AnalysisResult) {
    ast.Inspect(node, func(n ast.Node) bool {
        funcDecl, ok := n.(*ast.FuncDecl)
        if !ok {
            return true
        }

        hasLogging := false
        ast.Inspect(funcDecl.Body, func(bodyNode ast.Node) bool {
            // Check for log.Info, log.Error, etc.
            // ...
            return true
        })

        if !hasLogging {
            result.HandlersWithoutLogging = append(...)
        }
        return true
    })
}
```

---

## 📈 Comparison: AI Scores vs Deterministic Analysis

### Before (AI_REPORT.md - Subjective)

```markdown
| Category | Score | Status |
|----------|-------|--------|
| Backend Architecture | 9.0/10 | ✅ Excellent |
| Message Enrichment | 8.5/10 | ✅ Complete |
| Memory Service | 2.0/10 | 🔴 Critical |
```

**Problem**: What does "9.0/10" mean? How was it calculated?

### After (ANALYSIS_REPORT.md - Factual)

```markdown
| Metric | Count | Coverage |
|--------|-------|----------|
| Aggregates with optimistic locking | 13/33 | 39.4% |
| Handlers with BOLA protection | 35/179 | 19.6% |
| Queries with LIMIT clause | 42/90 | 46.7% |
| Test coverage | 82% | ✅ Above target |
```

**Benefit**: Clear, actionable metrics that can be tracked over time.

---

## 🎯 Recommended Workflow

### 1. Baseline (First Run)

```bash
./scripts/analyze_codebase.sh
go run scripts/deep_analyzer.go

# Save baseline
cp ANALYSIS_REPORT.md reports/baseline-2025-10-14.md
cp DEEP_ANALYSIS_REPORT.md reports/baseline-deep-2025-10-14.md
```

### 2. Sprint Planning

```bash
# Review P0 issues
grep -A 50 "P0 - CRITICAL" DEEP_ANALYSIS_REPORT.md

# Create tasks in TODO.md based on findings
```

### 3. Post-Sprint Analysis

```bash
# Re-run analysis
./scripts/analyze_codebase.sh
go run scripts/deep_analyzer.go

# Compare with baseline
diff reports/baseline-2025-10-14.md ANALYSIS_REPORT.md
```

### 4. Continuous Integration (CI)

Add to `.github/workflows/analysis.yml`:

```yaml
- name: Run Code Analysis
  run: |
    ./scripts/analyze_codebase.sh
    go run scripts/deep_analyzer.go

- name: Check Optimistic Locking Coverage
  run: |
    COVERAGE=$(grep "Optimistic Locking Coverage" ANALYSIS_REPORT.md | grep -o "[0-9]*\.[0-9]*%")
    if [ "${COVERAGE%.*}" -lt 80 ]; then
      echo "Error: Optimistic locking coverage below 80%"
      exit 1
    fi
```

---

## 🚀 Future Enhancements

### Planned Features

1. **HTML Dashboard**
   - Interactive charts (Chart.js)
   - Drill-down into specific issues
   - Historical trend analysis

2. **JSON Export**
   - Machine-readable output
   - Integration with monitoring tools
   - Programmatic analysis

3. **Diff Mode**
   - Compare two reports
   - Highlight improvements/regressions
   - Generate changelog

4. **IDE Integration**
   - VS Code extension
   - Real-time warnings
   - Quick fixes

5. **Custom Rules Engine**
   - Define custom patterns to detect
   - Team-specific best practices
   - Automatic PR comments

---

## 📚 References

### Clean Architecture
- Robert Martin - "Clean Architecture" (2017)
- Domain layer should not depend on infrastructure

### OWASP API Security
- [OWASP API Security Top 10 (2023)](https://owasp.org/API-Security/)
- API1: Broken Object Level Authorization (BOLA)
- API4: Unrestricted Resource Consumption

### DDD Patterns
- Eric Evans - "Domain-Driven Design" (2003)
- Vaughn Vernon - "Implementing Domain-Driven Design" (2013)

### Optimistic Locking
- Martin Fowler - [Optimistic Offline Lock](https://martinfowler.com/eaaCatalog/optimisticOfflineLock.html)
- Prevents lost updates in concurrent systems

---

## 🤝 Contributing

To add new analysis features:

1. **Fork the repo**
2. **Add your check** to `analyze_codebase.sh` or `deep_analyzer.go`
3. **Test thoroughly**:
   ```bash
   ./scripts/analyze_codebase.sh
   go run scripts/deep_analyzer.go
   ```
4. **Document the new metric** in this README
5. **Submit PR** with example output

---

## 📞 Support

Questions? Found a bug?

- Open an issue: [GitHub Issues](https://github.com/ventros-crm/issues)
- Slack: #crm-dev
- Email: dev@ventros.com

---

**Last Updated**: 2025-10-14
**Maintainer**: Ventros CRM Team
**License**: Proprietary
