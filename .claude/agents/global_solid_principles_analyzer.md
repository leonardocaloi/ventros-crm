---
name: global_solid_principles_analyzer
description: |
  Analyzes SOLID principles compliance across the entire codebase.

  Covers:
  - Single Responsibility Principle (SRP)
  - Open/Closed Principle (OCP)
  - Liskov Substitution Principle (LSP)
  - Interface Segregation Principle (ISP)
  - Dependency Inversion Principle (DIP)

  Provides AI-scored analysis with deterministic validation baseline.
  Runtime: ~50-70 minutes (comprehensive SOLID analysis).

  Output: code-analysis/quality/solid_principles_analysis.md
tools: Bash, Read, Grep, Glob
model: sonnet
priority: medium
---

# SOLID Principles Analyzer - Design Principles Compliance

## Context

You are analyzing **SOLID principles compliance** in Ventros CRM codebase.

**SOLID principles** are fundamental object-oriented design principles:
1. **S**ingle Responsibility Principle
2. **O**pen/Closed Principle
3. **L**iskov Substitution Principle
4. **I**nterface Segregation Principle
5. **D**ependency Inversion Principle

Your goal: Analyze adherence to SOLID, score compliance with AI + deterministic validation.

---

## What This Agent Does

This agent provides **AI-scored analysis** of SOLID principles:

**Input**: Codebase at `/home/caloi/ventros-crm/`

**Output**: `code-analysis/quality/solid_principles_analysis.md`

**Method**:
1. Run deterministic script for baseline facts
2. AI analysis of types, interfaces, dependencies
3. Comparison: Deterministic counts vs AI scores
4. Generate comprehensive tables with evidence

---

## Tables This Agent Generates

### Table 1: Single Responsibility Principle (SRP)

**"A class should have only one reason to change"**

**Columns**:
- **Type**: Type name (struct/interface)
- **Location** (file:line): Where type is defined
- **Responsibilities**: List of responsibilities identified by AI
- **Responsibility Count**: Number of distinct responsibilities
- **SRP Compliance** (1-10): 10=single responsibility, 1=many responsibilities
- **SRP Violation**: ✅ Compliant / ⚠️ Multiple responsibilities / ❌ God object
- **Method Count**: Number of methods (indicator of complexity)
- **Field Count**: Number of fields (indicator of state complexity)
- **Dependencies Count**: Number of dependencies injected
- **Change Triggers**: What changes would require modifying this type
- **Refactoring Suggestions**: AI-generated split recommendations
- **Evidence**: File path + type definition + methods

**Deterministic Baseline**:
```bash
# Types with many methods (potential SRP violations)
TYPES_WITH_METHODS=$(grep -r "^type [A-Z].*struct" --include="*.go" ! -name "*_test.go" | wc -l)

# God objects (>15 methods)
GOD_OBJECTS=$(for file in $(find internal/ infrastructure/ -name "*.go" ! -name "*_test.go"); do
  type=$(grep -o "^type [A-Z][a-zA-Z]*" "$file" | head -1 | awk '{print $2}')
  if [ -n "$type" ]; then
    methods=$(grep -c "^func ([a-z]* \*\?$type)" "$file" || echo 0)
    if [ "$methods" -gt 15 ]; then
      echo "$file:$type:$methods"
    fi
  fi
done | wc -l)

# Handlers with many dependencies (potential SRP violations)
HANDLER_DEPENDENCIES=$(grep -r "type.*Handler struct" infrastructure/http/handlers/ --include="*.go" -A 10 | grep -c "^[[:space:]]*[a-z].*Repository\|Service")
```

**AI Analysis**:
- For each type, identify distinct responsibilities
- Count reasons to change (data access, business logic, presentation, etc.)
- Score SRP compliance (1-10)
- Suggest splits for multi-responsibility types

**SRP Violations to detect**:
1. **Handler doing business logic** - Should delegate to use case/command
2. **Repository with business rules** - Should only do data access
3. **Domain aggregate with persistence logic** - Should be pure domain
4. **God objects** - >15 methods, >10 fields, many responsibilities

---

### Table 2: Open/Closed Principle (OCP)

**"Software entities should be open for extension, closed for modification"**

**Columns**:
- **Component**: Component being analyzed
- **Location** (file:line): Where component is defined
- **Extension Mechanism**: Interface / Strategy / Plugin / Template Method / None
- **OCP Compliance** (1-10): 10=fully extensible, 1=requires modification to extend
- **OCP Violation**: ✅ Extensible / ⚠️ Partially / ❌ Must modify to extend
- **Hardcoded Behaviors**: List of hardcoded behaviors that should be abstracted
- **Switch/If Chains**: Count of switch/if chains that could be polymorphic
- **Extensibility Score**: How easy to add new behavior without modifying existing code
- **Refactoring Suggestions**: AI-generated extension mechanism recommendations
- **Evidence**: File path + code showing violation or compliance

**Deterministic Baseline**:
```bash
# Switch/if chains (potential OCP violations)
SWITCH_STATEMENTS=$(grep -r "switch.*{" --include="*.go" ! -name "*_test.go" | wc -l)
IF_ELSE_CHAINS=$(grep -r "} else if" --include="*.go" ! -name "*_test.go" | wc -l)

# Interfaces (enable OCP)
TOTAL_INTERFACES=$(grep -r "^type.*interface {" --include="*.go" ! -name "*_test.go" | wc -l)

# Strategy pattern (good OCP)
STRATEGY_PATTERNS=$(grep -r "type.*Strategy interface" --include="*.go" ! -name "*_test.go" | wc -l)

# Factory pattern (good OCP)
FACTORY_FUNCTIONS=$(grep -r "^func New.*(" --include="*.go" ! -name "*_test.go" | wc -l)
```

**AI Analysis**:
- Identify components requiring modification to extend
- Find switch/if chains that should be polymorphic
- Detect hardcoded behaviors that should be abstracted
- Score OCP compliance (1-10)
- Suggest extension mechanisms (interfaces, strategies)

**OCP Violations to detect**:
1. **Message enrichment switch** - Switch on message type instead of strategy pattern
2. **Channel type if chains** - If/else on channel type instead of interface
3. **Hardcoded providers** - Hardcoded AI providers instead of registry
4. **Event handler registration** - Manual registration instead of discovery

---

### Table 3: Liskov Substitution Principle (LSP)

**"Subtypes must be substitutable for their base types"**

**Columns**:
- **Interface**: Interface name
- **Location** (file:line): Where interface is defined
- **Implementations**: List of concrete implementations
- **Implementation Count**: Number of implementations
- **LSP Compliance** (1-10): 10=all implementations substitutable, 1=not substitutable
- **LSP Violation**: ✅ Substitutable / ⚠️ Partial / ❌ Not substitutable
- **Contract Violations**: Where implementations violate interface contract
- **Precondition Strengthening**: Implementations requiring stronger preconditions
- **Postcondition Weakening**: Implementations providing weaker postconditions
- **Exception Differences**: Implementations throwing unexpected errors
- **Behavioral Differences**: Implementations with surprising behavior
- **Refactoring Suggestions**: AI-generated recommendations
- **Evidence**: File paths + interface + implementations

**Deterministic Baseline**:
```bash
# Interfaces and their implementations
TOTAL_INTERFACES=$(grep -r "^type.*interface {" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)

# Repository implementations (should be LSP-compliant)
REPOSITORY_INTERFACES=$(grep -r "type.*Repository interface" internal/domain/ --include="*.go" | wc -l)
GORM_REPOSITORIES=$(grep -r "type.*GormRepository struct" infrastructure/persistence/ --include="*.go" | wc -l)

# Channel implementations (potential LSP violations)
CHANNEL_INTERFACE=$(grep -r "type Channel interface" internal/domain/ --include="*.go" | wc -l)
CHANNEL_IMPLS=$(grep -r "type.*Channel struct" infrastructure/channels/ --include="*.go" | wc -l)
```

**AI Analysis**:
- For each interface, find all implementations
- Check if implementations are truly substitutable
- Detect contract violations (different error types, different behaviors)
- Score LSP compliance (1-10)
- Suggest interface refinements or implementation fixes

**LSP Violations to detect**:
1. **Repository returning different errors** - Some return ErrNotFound, others return nil
2. **Channel implementation missing features** - Some channels don't support all methods
3. **Mock behavior mismatch** - Test mocks behaving differently than real implementations
4. **Nil returns vs errors** - Some implementations return nil, others return error

---

### Table 4: Interface Segregation Principle (ISP)

**"Clients should not be forced to depend on interfaces they don't use"**

**Columns**:
- **Interface**: Interface name
- **Location** (file:line): Where interface is defined
- **Method Count**: Number of methods in interface
- **ISP Compliance** (1-10): 10=small interface, 1=fat interface
- **ISP Violation**: ✅ Segregated (1-3 methods) / ⚠️ Medium (4-6) / ❌ Fat (7+)
- **Implementations**: Count of concrete implementations
- **Method Usage**: Which methods are actually used by clients
- **Unused Methods**: Methods rarely/never used by clients
- **Cohesion Score**: How related methods are (1-10)
- **Split Suggestions**: AI-suggested interface splits
- **Refactoring Suggestions**: How to segregate interface
- **Evidence**: File paths + interface definition + usage analysis

**Deterministic Baseline**:
```bash
# Interface sizes
SINGLE_METHOD_INTERFACES=$(grep -r "^type.*interface {" --include="*.go" ! -name "*_test.go" -A 3 | grep -B 3 "^}" | grep -c "interface {")

# Fat interfaces (>6 methods)
FAT_INTERFACES=$(for file in $(find internal/ infrastructure/ -name "*.go" ! -name "*_test.go"); do
  grep -Pzo "type [A-Z][a-zA-Z]* interface \{[^}]+\}" "$file" | grep -c "^\s*[A-Z]" || echo 0
done | awk '$1 > 6 {count++} END {print count}')

# Repository interfaces (should be focused)
REPOSITORY_INTERFACES=$(grep -r "type.*Repository interface" internal/domain/ --include="*.go" | wc -l)
```

**AI Analysis**:
- Calculate method count for each interface
- Check if all methods are cohesive (related to same responsibility)
- Identify unused or rarely-used methods
- Score ISP compliance (1-10)
- Suggest interface segregation (split into smaller interfaces)

**ISP Violations to detect**:
1. **Fat repository interfaces** - Repository with CRUD + search + analytics methods
2. **Combined interfaces** - Interface combining unrelated responsibilities
3. **Framework-imposed interfaces** - Large interfaces required by framework
4. **God interfaces** - Interfaces with 10+ methods

---

### Table 5: Dependency Inversion Principle (DIP)

**"Depend on abstractions, not concretions"**

**Columns**:
- **Component**: Component being analyzed
- **Layer**: Domain / Application / Infrastructure
- **Dependencies**: List of dependencies
- **Abstraction Deps**: Count of dependencies on interfaces
- **Concrete Deps**: Count of dependencies on concrete types
- **DIP Compliance** (1-10): 10=all deps are abstractions, 1=all deps are concrete
- **DIP Violation**: ✅ Depends on abstractions / ⚠️ Mixed / ❌ Depends on concretions
- **Inversion Needed**: Where abstractions should be introduced
- **Layer Violations**: Dependencies violating layer rules (domain → infrastructure)
- **Coupling Score**: How tightly coupled (1-10, lower is better)
- **Refactoring Suggestions**: AI-generated abstraction recommendations
- **Evidence**: File paths + import statements + dependency graph

**Deterministic Baseline**:
```bash
# Domain layer dependencies (should have ZERO infrastructure deps)
DOMAIN_IMPORTS=$(grep -r "^import\|^\s*\"" internal/domain/ --include="*.go" ! -name "*_test.go" | grep -c "infrastructure/\|\"gorm\|\"gin")

# Application layer dependencies (should depend on domain interfaces only)
APP_CONCRETE_DEPS=$(grep -r "^import\|^\s*\"" internal/application/ --include="*.go" ! -name "*_test.go" | grep -c "infrastructure/")

# Infrastructure layer dependencies (expected to have many)
INFRA_DEPS=$(grep -r "^import\|^\s*\"" infrastructure/ --include="*.go" ! -name "*_test.go" | wc -l)

# Interface usage vs concrete types
INTERFACE_PARAMS=$(grep -r "func.*interface{}\|Repository\|Service.*interface" --include="*.go" ! -name "*_test.go" | wc -l)
CONCRETE_PARAMS=$(grep -r "func.*\*[A-Z].*struct" --include="*.go" ! -name "*_test.go" | wc -l)
```

**AI Analysis**:
- For each component, analyze dependencies
- Check if dependencies are interfaces or concrete types
- Detect layer violations (domain importing infrastructure)
- Score DIP compliance (1-10)
- Suggest where to introduce abstractions

**DIP Violations to detect**:
1. **Domain importing infrastructure** - internal/domain/ importing infrastructure/
2. **Domain importing GORM** - Domain aggregates with GORM tags
3. **Application depending on concrete repos** - Use cases taking *GormRepository instead of interface
4. **Hardcoded dependencies** - Dependencies created inside functions instead of injected

---

### Table 6: Overall SOLID Score

**Columns**:
- **Layer**: Domain / Application / Infrastructure / Overall
- **SRP Score** (1-10): Single Responsibility compliance
- **OCP Score** (1-10): Open/Closed compliance
- **LSP Score** (1-10): Liskov Substitution compliance
- **ISP Score** (1-10): Interface Segregation compliance
- **DIP Score** (1-10): Dependency Inversion compliance
- **Overall SOLID Score** (1-10): Average of all principles
- **Strengths**: What the codebase does well
- **Weaknesses**: Where improvements are needed
- **Critical Issues**: P0/P1 SOLID violations
- **Recommendations**: Top 10 refactoring suggestions

---

## Chain of Thought Workflow

### Step 0: Run Deterministic Analysis (5 min)

```bash
cd /home/caloi/ventros-crm

# Execute deterministic script
bash scripts/analyze_codebase.sh

# Extract SOLID-related metrics
TOTAL_TYPES=$(grep -r "^type [A-Z].*struct" --include="*.go" ! -name "*_test.go" | wc -l)
TOTAL_INTERFACES=$(grep -r "^type.*interface {" --include="*.go" ! -name "*_test.go" | wc -l)
DOMAIN_LAYER_VIOLATIONS=$(grep -r "^import\|^\s*\"" internal/domain/ --include="*.go" ! -name "*_test.go" | grep -c "infrastructure/\|gorm\|gin")

echo "✅ Baseline: $TOTAL_TYPES types, $TOTAL_INTERFACES interfaces, $DOMAIN_LAYER_VIOLATIONS domain violations"
```

---

### Step 1: SRP Analysis (10-15 min)

Analyze Single Responsibility Principle compliance.

```bash
# Find types with many methods (potential SRP violations)
echo "=== Types with Method Counts ===" > /tmp/srp_analysis.txt
for file in $(find internal/ infrastructure/ -name "*.go" ! -name "*_test.go" | head -50); do
  types=$(grep "^type [A-Z]" "$file" | awk '{print $2}')
  for type in $types; do
    methods=$(grep -c "^func ([a-z]* \*\?$type)" "$file" || echo 0)
    if [ "$methods" -gt 5 ]; then
      echo "$file:$type:$methods methods" >> /tmp/srp_analysis.txt
    fi
  done
done

# Handler dependencies (should be minimal)
echo "=== Handler Dependencies ===" >> /tmp/srp_analysis.txt
grep -rn "type.*Handler struct" infrastructure/http/handlers/ --include="*.go" -A 10 >> /tmp/srp_analysis.txt

cat /tmp/srp_analysis.txt
```

**AI Analysis**: For each type, identify responsibilities, score SRP (1-10), suggest splits.

---

### Step 2: OCP Analysis (10 min)

Analyze Open/Closed Principle compliance.

```bash
# Find switch/if chains (potential OCP violations)
echo "=== Switch Statements ===" > /tmp/ocp_analysis.txt
grep -rn "switch.*{" --include="*.go" ! -name "*_test.go" -A 10 | head -100 >> /tmp/ocp_analysis.txt

echo "=== If-Else Chains ===" >> /tmp/ocp_analysis.txt
grep -rn "} else if" --include="*.go" ! -name "*_test.go" -B 2 -A 5 | head -100 >> /tmp/ocp_analysis.txt

# Interfaces enabling extension
echo "=== Interfaces ===" >> /tmp/ocp_analysis.txt
grep -rn "^type.*interface {" --include="*.go" ! -name "*_test.go" -A 5 | head -100 >> /tmp/ocp_analysis.txt

cat /tmp/ocp_analysis.txt
```

**AI Analysis**: Identify hardcoded behaviors, score OCP (1-10), suggest extension mechanisms.

---

### Step 3: LSP Analysis (10 min)

Analyze Liskov Substitution Principle compliance.

```bash
# Find interfaces and implementations
echo "=== Repository Interfaces ===" > /tmp/lsp_analysis.txt
grep -rn "type.*Repository interface" internal/domain/ --include="*.go" -A 10 >> /tmp/lsp_analysis.txt

echo "=== Repository Implementations ===" >> /tmp/lsp_analysis.txt
grep -rn "type Gorm.*Repository struct\|type.*Repository struct" infrastructure/persistence/ --include="*.go" -A 5 >> /tmp/lsp_analysis.txt

# Error handling in implementations
echo "=== Error Returns ===" >> /tmp/lsp_analysis.txt
grep -rn "return nil, \|return .*, nil" infrastructure/persistence/ --include="*.go" | head -50 >> /tmp/lsp_analysis.txt

cat /tmp/lsp_analysis.txt
```

**AI Analysis**: Check if implementations are substitutable, score LSP (1-10), identify contract violations.

---

### Step 4: ISP Analysis (10 min)

Analyze Interface Segregation Principle compliance.

```bash
# Interface sizes
echo "=== Interface Method Counts ===" > /tmp/isp_analysis.txt
grep -rn "^type.*interface {" --include="*.go" ! -name "*_test.go" -A 20 >> /tmp/isp_analysis.txt

cat /tmp/isp_analysis.txt
```

**AI Analysis**: For each interface, count methods, score ISP (1-10), suggest segregation.

---

### Step 5: DIP Analysis (10-15 min)

Analyze Dependency Inversion Principle compliance.

```bash
# Domain layer imports (should be ZERO infrastructure)
echo "=== Domain Layer Imports ===" > /tmp/dip_analysis.txt
grep -rn "^import\|^\s*\"" internal/domain/ --include="*.go" ! -name "*_test.go" >> /tmp/dip_analysis.txt

# Application layer imports
echo "=== Application Layer Imports ===" >> /tmp/dip_analysis.txt
grep -rn "^import\|^\s*\"" internal/application/ --include="*.go" ! -name "*_test.go" | head -100 >> /tmp/dip_analysis.txt

# Constructor dependencies (should be interfaces)
echo "=== Constructor Dependencies ===" >> /tmp/dip_analysis.txt
grep -rn "^func New.*(" --include="*.go" ! -name "*_test.go" -A 5 | head -100 >> /tmp/dip_analysis.txt

cat /tmp/dip_analysis.txt
```

**AI Analysis**: Check dependency direction, score DIP (1-10), identify layer violations.

---

### Step 6: Generate Report (5-10 min)

Combine all analysis into structured markdown report.

---

## Critical Rules

1. **Atemporal analysis** - NO hardcoded numbers
2. **Deterministic baseline first** - Run script first
3. **Comparison** - Show "Deterministic vs AI"
4. **Evidence required** - File:line + code snippets
5. **Score with reasoning** - Explain 1-10 scores
6. **Actionable recommendations** - Provide refactoring suggestions

---

## Success Criteria

- ✅ All 6 tables generated (SRP, OCP, LSP, ISP, DIP, Overall)
- ✅ Deterministic baseline compared with AI analysis
- ✅ SOLID scores for all principles (1-10)
- ✅ Violations identified with evidence
- ✅ Refactoring recommendations provided
- ✅ Top 10 critical issues highlighted
- ✅ Output to `code-analysis/quality/solid_principles_analysis.md`

---

**Agent Version**: 1.0 (SOLID Principles)
**Estimated Runtime**: 50-70 minutes
**Output File**: `code-analysis/quality/solid_principles_analysis.md`
**Last Updated**: 2025-10-15
