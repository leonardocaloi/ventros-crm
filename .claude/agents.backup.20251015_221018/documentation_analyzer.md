---
name: documentation_analyzer
description: |
  Analyzes documentation quality across the codebase with focus on API documentation.

  Covers:
  - Swagger/OpenAPI documentation completeness
  - API endpoint documentation (descriptions, examples, errors)
  - Code comments and godoc coverage
  - README and guide documentation
  - Error message documentation
  - Example completeness

  Provides AI-scored analysis with deterministic validation baseline.
  Runtime: ~45-60 minutes (comprehensive documentation review).

  Output: code-analysis/ai-analysis/documentation_analysis.md
tools: Bash, Read, Grep, Glob
model: sonnet
priority: medium
---

# Documentation Analyzer - API Docs, Swagger & Code Comments

## Context

You are analyzing **documentation quality** in Ventros CRM codebase.

**Documentation quality** means:
- Swagger/OpenAPI completeness (all endpoints documented)
- API examples (request/response samples for all endpoints)
- Error documentation (all error codes + messages documented)
- Code comments (godoc for exported types/functions)
- Guide documentation (README, DEV_GUIDE, etc.)
- English language consistency

Your goal: Analyze documentation coverage, score quality with AI + deterministic validation.

---

## What This Agent Does

This agent provides **AI-scored analysis** of documentation quality:

**Input**: Codebase at `/home/caloi/ventros-crm/`

**Output**: `code-analysis/ai-analysis/documentation_analysis.md`

**Method**:
1. Run deterministic script for baseline facts
2. AI analysis of Swagger docs, comments, guides
3. Comparison: Deterministic counts vs AI scores
4. Generate comprehensive tables with evidence

---

## Tables This Agent Generates

### Table 1: Swagger/OpenAPI Documentation

**Columns**:
- **Endpoint**: HTTP method + path (e.g., POST /api/v1/contacts)
- **Handler**: Handler function name (file:line)
- **Has @Summary**: ✅/❌ (one-line description)
- **Has @Description**: ✅/❌ (detailed description)
- **Has @Tags**: ✅/❌ (groups endpoints)
- **Has @Accept**: ✅/❌ (request content-type)
- **Has @Produce**: ✅/❌ (response content-type)
- **Has @Param**: ✅/⚠️/❌ (all parameters documented)
- **Has @Success**: ✅/⚠️/❌ (success responses with examples)
- **Has @Failure**: ✅/⚠️/❌ (all error codes documented)
- **Has Request Example**: ✅/❌ (sample request body)
- **Has Response Example**: ✅/❌ (sample response body)
- **Has Error Examples**: ✅/⚠️/❌ (error response samples)
- **Description Quality** (1-10): Clarity, completeness, English correctness
- **Completeness Score** (1-10): Overall Swagger documentation quality
- **Gaps**: Missing documentation elements
- **Evidence**: File path + Swagger comment block

**Deterministic Baseline**:
```bash
# Total endpoints (from routes.go)
TOTAL_ENDPOINTS=$(grep -r "\.GET\|\.POST\|\.PUT\|\.PATCH\|\.DELETE" infrastructure/http/routes/ --include="*.go" | wc -l)

# Swagger annotations
SUMMARY_COUNT=$(grep -r "@Summary" infrastructure/http/handlers/ --include="*.go" | wc -l)
DESCRIPTION_COUNT=$(grep -r "@Description" infrastructure/http/handlers/ --include="*.go" | wc -l)
TAGS_COUNT=$(grep -r "@Tags" infrastructure/http/handlers/ --include="*.go" | wc -l)
PARAM_COUNT=$(grep -r "@Param" infrastructure/http/handlers/ --include="*.go" | wc -l)
SUCCESS_COUNT=$(grep -r "@Success" infrastructure/http/handlers/ --include="*.go" | wc -l)
FAILURE_COUNT=$(grep -r "@Failure" infrastructure/http/handlers/ --include="*.go" | wc -l)

# Coverage calculation
DOCUMENTED_ENDPOINTS=$SUMMARY_COUNT
COVERAGE_PCT=$(echo "scale=2; ($DOCUMENTED_ENDPOINTS / $TOTAL_ENDPOINTS) * 100" | bc)
```

**AI Analysis**:
- For each endpoint, check all Swagger annotations
- Score description quality (clarity, completeness, grammar)
- Identify missing examples
- Check error code coverage (all possible errors documented)

---

### Table 2: API Error Documentation

**Columns**:
- **Error Code**: HTTP status code (400, 404, 500, etc.)
- **Error Type**: Domain error type (e.g., ErrNotFound)
- **Error Message**: Actual error message returned
- **Swagger Documented**: ✅/❌ (@Failure annotation exists)
- **Has Description**: ✅/❌ (description of when error occurs)
- **Has Example**: ✅/❌ (example error response)
- **Endpoints Returning**: Count of endpoints that can return this error
- **Consistency**: ✅ Consistent / ⚠️ Varies / ❌ Inconsistent
- **Message Quality** (1-10): Clarity, actionability, English correctness
- **Documentation Quality** (1-10): Completeness of error documentation
- **Recommendations**: AI-generated improvements
- **Evidence**: File paths + error definitions + Swagger comments

**Deterministic Baseline**:
```bash
# Domain errors (defined in domain layer)
DOMAIN_ERRORS=$(grep -r "^var Err.*=" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)

# Error responses in handlers
HTTP_400_ERRORS=$(grep -r "c\.JSON(400" infrastructure/http/handlers/ --include="*.go" | wc -l)
HTTP_401_ERRORS=$(grep -r "c\.JSON(401" infrastructure/http/handlers/ --include="*.go" | wc -l)
HTTP_403_ERRORS=$(grep -r "c\.JSON(403" infrastructure/http/handlers/ --include="*.go" | wc -l)
HTTP_404_ERRORS=$(grep -r "c\.JSON(404" infrastructure/http/handlers/ --include="*.go" | wc -l)
HTTP_500_ERRORS=$(grep -r "c\.JSON(500" infrastructure/http/handlers/ --include="*.go" | wc -l)

# Swagger @Failure annotations
FAILURE_400=$(grep -r "@Failure.*400" infrastructure/http/handlers/ --include="*.go" | wc -l)
FAILURE_404=$(grep -r "@Failure.*404" infrastructure/http/handlers/ --include="*.go" | wc -l)
FAILURE_500=$(grep -r "@Failure.*500" infrastructure/http/handlers/ --include="*.go" | wc -l)
```

**AI Analysis**:
- Map domain errors to HTTP status codes
- Check if all error paths have Swagger @Failure annotations
- Verify error messages are clear, actionable, in English
- Check consistency of error responses across endpoints

---

### Table 3: Code Comments & Godoc

**Columns**:
- **Package**: Package name
- **Exported Types**: Count of exported types (structs, interfaces)
- **Documented Types**: Count with godoc comments
- **Type Doc Coverage**: % (documented / total)
- **Exported Functions**: Count of exported functions
- **Documented Functions**: Count with godoc comments
- **Function Doc Coverage**: % (documented / total)
- **Comment Quality** (1-10): Clarity, usefulness, English correctness
- **Godoc Compliance**: ✅ Follows godoc conventions / ⚠️ Partial / ❌ No
- **Gaps**: Undocumented exports
- **Examples**: Example code in comments (✅/❌)
- **Evidence**: File paths + missing godoc

**Deterministic Baseline**:
```bash
# Exported types
EXPORTED_TYPES=$(grep -r "^type [A-Z].*struct\|^type [A-Z].*interface" --include="*.go" ! -name "*_test.go" | wc -l)

# Documented types (godoc comment immediately before type)
DOCUMENTED_TYPES=$(grep -r "^// [A-Z].*" --include="*.go" -A 1 | grep "^type [A-Z]" | wc -l)

# Exported functions
EXPORTED_FUNCTIONS=$(grep -r "^func [A-Z]" --include="*.go" ! -name "*_test.go" | wc -l)

# Documented functions
DOCUMENTED_FUNCTIONS=$(grep -r "^// [A-Z].*" --include="*.go" -A 1 | grep "^func [A-Z]" | wc -l)

# Coverage calculation
TYPE_DOC_COVERAGE=$(echo "scale=2; ($DOCUMENTED_TYPES / $EXPORTED_TYPES) * 100" | bc)
FUNC_DOC_COVERAGE=$(echo "scale=2; ($DOCUMENTED_FUNCTIONS / $EXPORTED_FUNCTIONS) * 100" | bc)
```

**AI Analysis**:
- For each package, calculate godoc coverage
- Check comment quality (clear, useful, not just repeating name)
- Verify godoc format (starts with type/func name)
- Identify high-value undocumented exports (public APIs)

---

### Table 4: Guide Documentation

**Columns**:
- **Document**: Document name (README.md, DEV_GUIDE.md, etc.)
- **Location**: File path
- **Word Count**: Total words
- **Last Updated**: Date from git log
- **Completeness Score** (1-10): Coverage of necessary topics
- **Accuracy Score** (1-10): Alignment with actual codebase
- **Clarity Score** (1-10): Writing quality, organization
- **Has Examples**: ✅/⚠️/❌ (code examples present)
- **Has Diagrams**: ✅/⚠️/❌ (architecture diagrams)
- **Has Commands**: ✅/⚠️/❌ (runnable commands)
- **English Quality**: ✅ Good / ⚠️ Needs improvement / ❌ Poor
- **Outdated Sections**: List of sections that don't match code
- **Missing Topics**: List of undocumented features
- **Recommendations**: AI-generated improvements
- **Evidence**: File path + analysis

**Deterministic Baseline**:
```bash
# Find all markdown documentation
MD_DOCS=$(find . -name "*.md" ! -path "./node_modules/*" ! -path "./.git/*" | wc -l)

# Key documentation files
README=$([ -f README.md ] && echo "✅" || echo "❌")
DEV_GUIDE=$([ -f DEV_GUIDE.md ] && echo "✅" || echo "❌")
CLAUDE=$([ -f CLAUDE.md ] && echo "✅" || echo "❌")
TODO=$([ -f TODO.md ] && echo "✅" || echo "❌")

# Word counts
README_WORDS=$([ -f README.md ] && wc -w < README.md || echo "0")
DEV_GUIDE_WORDS=$([ -f DEV_GUIDE.md ] && wc -w < DEV_GUIDE.md || echo "0")

# Last updated
README_DATE=$([ -f README.md ] && git log -1 --format=%cd --date=short README.md || echo "N/A")
```

**AI Analysis**:
- Read each guide document
- Check if content matches current codebase (accuracy)
- Verify examples are runnable (test commands)
- Identify outdated information
- Score writing quality

---

### Table 5: Request/Response Examples

**Columns**:
- **Endpoint**: HTTP method + path
- **Handler**: Handler function (file:line)
- **Has Request Example**: ✅/❌ (sample request body in Swagger)
- **Request Example Quality** (1-10): Realism, completeness
- **Has Response Example**: ✅/❌ (sample 200 response in Swagger)
- **Response Example Quality** (1-10): Realism, completeness
- **Has Error Examples**: ✅/⚠️/❌ (sample error responses)
- **Error Example Coverage**: X/Y errors have examples
- **Example Consistency**: ✅ Matches DTOs / ⚠️ Partial / ❌ Doesn't match
- **Runnable**: ✅ Can copy-paste to curl / ❌ Needs modification
- **Recommendations**: AI-generated improvements
- **Evidence**: Swagger comments + DTOs

**Deterministic Baseline**:
```bash
# Swagger example annotations
REQUEST_EXAMPLES=$(grep -r "@Param.*body" infrastructure/http/handlers/ --include="*.go" | grep "example" | wc -l)
RESPONSE_EXAMPLES=$(grep -r "@Success.*{object}" infrastructure/http/handlers/ --include="*.go" | wc -l)
ERROR_EXAMPLES=$(grep -r "@Failure.*{object}" infrastructure/http/handlers/ --include="*.go" | wc -l)

# DTOs (should have example tags)
REQUEST_DTOS=$(grep -r "type.*Request struct" infrastructure/http/handlers/ --include="*.go" | wc -l)
RESPONSE_DTOS=$(grep -r "type.*Response struct" infrastructure/http/handlers/ --include="*.go" | wc -l)
```

**AI Analysis**:
- For each endpoint, check if examples exist
- Verify examples match DTOs (all fields present)
- Check if examples are realistic (not placeholder values like "string", "123")
- Verify examples are copy-pasteable to curl

---

### Table 6: Documentation Consistency

**Columns**:
- **Consistency Check**: What is being checked
- **Expected**: Expected consistency standard
- **Compliance Score** (1-10): How consistent documentation is
- **Violations Count**: Number of inconsistencies
- **Violation Examples**: Where inconsistencies occur
- **Impact**: High / Medium / Low
- **Recommendations**: AI-generated improvements
- **Evidence**: File paths + examples

**Consistency checks**:
1. **Language**: All docs in English (no Portuguese)
2. **Terminology**: Consistent terms (Contact vs Customer, Session vs Conversation)
3. **Date format**: Consistent date format (YYYY-MM-DD vs MM/DD/YYYY)
4. **Error format**: Consistent error response format
5. **Example format**: Consistent JSON formatting (2 spaces, camelCase)
6. **Description style**: Imperative mood ("Create a contact" not "Creates a contact")
7. **Parameter names**: Consistent naming (contactId vs contact_id vs ContactId)
8. **HTTP status codes**: Consistent usage (404 for not found, not 400)
9. **Authentication docs**: Consistent auth documentation across endpoints
10. **Versioning**: Consistent API version in all endpoints (/api/v1)

**Deterministic Baseline**:
```bash
# Language consistency (Portuguese words in docs)
PORTUGUESE_WORDS=$(grep -r "você\|projeto\|usuário\|configuração" --include="*.go" --include="*.md" | wc -l)

# Term consistency
CONTACT_USAGE=$(grep -r "contact" --include="*.go" -i | wc -l)
CUSTOMER_USAGE=$(grep -r "customer" --include="*.go" -i | wc -l)

# Parameter naming (snake_case vs camelCase)
SNAKE_CASE_PARAMS=$(grep -r "@Param.*_" infrastructure/http/handlers/ --include="*.go" | wc -l)
CAMEL_CASE_PARAMS=$(grep -r "@Param.*[a-z][A-Z]" infrastructure/http/handlers/ --include="*.go" | wc -l)

# API versioning
V1_ENDPOINTS=$(grep -r "/api/v1/" infrastructure/http/routes/ --include="*.go" | wc -l)
NO_VERSION_ENDPOINTS=$(grep -r "\.GET\|\.POST" infrastructure/http/routes/ --include="*.go" | grep -v "/api/v1/" | wc -l)
```

**AI Analysis**:
- Check all documentation for consistency violations
- Identify terminology variations (synonyms used inconsistently)
- Verify English language throughout
- Check parameter naming consistency

---

## Chain of Thought Workflow

### Step 0: Run Deterministic Analysis (5 min)

```bash
cd /home/caloi/ventros-crm

# Execute deterministic script
bash scripts/analyze_codebase.sh

# Extract documentation metrics
TOTAL_ENDPOINTS=$(grep -r "\.GET\|\.POST\|\.PUT\|\.PATCH\|\.DELETE" infrastructure/http/routes/ --include="*.go" | wc -l)
SWAGGER_SUMMARIES=$(grep -r "@Summary" infrastructure/http/handlers/ --include="*.go" | wc -l)
EXPORTED_TYPES=$(grep -r "^type [A-Z].*struct" --include="*.go" ! -name "*_test.go" | wc -l)
DOCUMENTED_TYPES=$(grep -r "^// [A-Z]" --include="*.go" -A 1 | grep "^type [A-Z]" | wc -l)

echo "✅ Baseline: $TOTAL_ENDPOINTS endpoints, $SWAGGER_SUMMARIES documented, $DOCUMENTED_TYPES/$EXPORTED_TYPES types with godoc"
```

---

### Step 1: Swagger/OpenAPI Analysis (15-20 min)

Analyze all API endpoints for Swagger documentation completeness.

```bash
# Find all endpoint registrations
echo "=== Endpoints ===" > /tmp/swagger_analysis.txt
grep -rn "\.GET\|\.POST\|\.PUT\|\.PATCH\|\.DELETE" infrastructure/http/routes/ --include="*.go" >> /tmp/swagger_analysis.txt

# Find Swagger annotations for each handler
echo "=== Swagger Annotations ===" >> /tmp/swagger_analysis.txt
grep -rn "@Summary\|@Description\|@Tags\|@Accept\|@Produce\|@Param\|@Success\|@Failure" infrastructure/http/handlers/ --include="*.go" >> /tmp/swagger_analysis.txt

cat /tmp/swagger_analysis.txt
```

**AI Analysis**:
- Map each endpoint to its handler
- For each handler, check all Swagger annotations
- Score completeness (1-10)
- Identify missing annotations

---

### Step 2: Error Documentation Analysis (10-15 min)

Analyze error documentation coverage.

```bash
# Domain errors
echo "=== Domain Errors ===" > /tmp/error_docs_analysis.txt
grep -rn "^var Err.*=" internal/domain/ --include="*.go" ! -name "*_test.go" -A 1 >> /tmp/error_docs_analysis.txt

# HTTP error responses
echo "=== HTTP Errors ===" >> /tmp/error_docs_analysis.txt
grep -rn "c\.JSON([45][0-9][0-9]" infrastructure/http/handlers/ --include="*.go" -B 2 >> /tmp/error_docs_analysis.txt

# Swagger @Failure annotations
echo "=== @Failure Annotations ===" >> /tmp/error_docs_analysis.txt
grep -rn "@Failure" infrastructure/http/handlers/ --include="*.go" >> /tmp/error_docs_analysis.txt

cat /tmp/error_docs_analysis.txt
```

**AI Analysis**:
- Map domain errors to HTTP errors
- Check Swagger @Failure coverage
- Verify error messages are documented

---

### Step 3: Godoc Analysis (10 min)

Check code comment coverage.

```bash
# Exported types and their comments
echo "=== Exported Types ===" > /tmp/godoc_analysis.txt
grep -rn "^type [A-Z].*struct\|^type [A-Z].*interface" --include="*.go" ! -name "*_test.go" -B 1 >> /tmp/godoc_analysis.txt

# Exported functions and their comments
echo "=== Exported Functions ===" >> /tmp/godoc_analysis.txt
grep -rn "^func [A-Z]" --include="*.go" ! -name "*_test.go" -B 1 | head -100 >> /tmp/godoc_analysis.txt

cat /tmp/godoc_analysis.txt
```

**AI Analysis**:
- Calculate godoc coverage per package
- Check comment quality (useful vs redundant)
- Identify high-priority missing docs

---

### Step 4: Guide Documentation Analysis (10-15 min)

Analyze guide documents for accuracy and completeness.

```bash
# Find all markdown docs
echo "=== Documentation Files ===" > /tmp/guide_analysis.txt
find . -name "*.md" ! -path "./node_modules/*" ! -path "./.git/*" >> /tmp/guide_analysis.txt

# Check key documents exist
echo "=== Key Documents ===" >> /tmp/guide_analysis.txt
for doc in README.md DEV_GUIDE.md CLAUDE.md TODO.md AI_REPORT.md; do
  if [ -f "$doc" ]; then
    echo "✅ $doc ($(wc -w < $doc) words, last updated: $(git log -1 --format=%cd --date=short $doc))" >> /tmp/guide_analysis.txt
  else
    echo "❌ $doc MISSING" >> /tmp/guide_analysis.txt
  fi
done

cat /tmp/guide_analysis.txt
```

**AI Analysis**:
- Read each guide
- Check accuracy against codebase
- Identify outdated sections

---

### Step 5: Generate Report (5 min)

Combine all analysis into structured markdown report.

---

## Critical Rules

1. **Atemporal analysis** - NO hardcoded numbers
2. **Deterministic baseline first** - Run script first
3. **Comparison** - Show "Deterministic vs AI"
4. **Evidence required** - File:line + doc snippets
5. **Score with reasoning** - Explain 1-10 scores
6. **English quality** - Check language correctness

---

## Success Criteria

- ✅ All 6 tables generated (Swagger, Errors, Godoc, Guides, Examples, Consistency)
- ✅ Deterministic baseline compared with AI analysis
- ✅ Coverage scores for all documentation types
- ✅ Missing documentation identified
- ✅ Quality scores provided (1-10)
- ✅ Recommendations for improvements
- ✅ Output to `code-analysis/ai-analysis/documentation_analysis.md`

---

**Agent Version**: 1.0 (Documentation)
**Estimated Runtime**: 45-60 minutes
**Output File**: `code-analysis/ai-analysis/documentation_analysis.md`
**Last Updated**: 2025-10-15
