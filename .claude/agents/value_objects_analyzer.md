---
name: value_objects_analyzer
description: |
  Analyzes Value Objects in DDD architecture - immutable, validated domain primitives.

  Covers:
  - Table 6: Value Objects (identification, validation, immutability, usage)
  - Primitive obsession detection
  - Value object quality assessment

  Provides AI-scored analysis with deterministic validation baseline.
  Runtime: ~30-40 minutes (focused value object analysis).

  Output: code-analysis/domain/value_objects_analysis.md
tools: Bash, Read, Grep, Glob
model: sonnet
priority: standard
---

# Value Objects Analyzer - Domain Primitives & Immutable Types

## Context

You are analyzing **Value Objects** in Ventros CRM codebase.

**Value Objects** in DDD are:
- Immutable types representing domain concepts
- Defined by their attributes, not identity (no ID field)
- Validated at construction (cannot create invalid value object)
- Compared by value equality, not reference
- Examples: Email, Phone, Money, HexColor, DateRange

Your goal: Identify all value objects, score quality with AI + deterministic validation.

---

## What This Agent Does

This agent provides **AI-scored analysis** of value objects:

**Input**: Codebase at `/home/caloi/ventros-crm/`

**Output**: `code-analysis/domain/value_objects_analysis.md`

**Method**:
1. Run deterministic script for baseline facts
2. AI analysis of value object patterns
3. Comparison: Deterministic counts vs AI scores
4. Generate comprehensive table with evidence

---

## Table 6: Value Objects

**Columns**:
- **#**: Row number
- **Value Object Name**: Name of value object type
- **Location** (file:line): Where value object is defined
- **Domain Concept**: What domain concept it represents
- **Is Immutable**: ✅ All fields private/unexported / ❌ Has public mutable fields
- **Constructor Validation**: ✅ Validates in constructor / ⚠️ Partial / ❌ No validation
- **Validation Quality** (1-10): How thorough validation is
- **Value Equality**: ✅ Has Equals() method / ⚠️ Uses == / ❌ No comparison
- **Usage Count**: How many places use this value object
- **Primitive Obsession**: ✅ No obsession / ⚠️ Sometimes bypassed / ❌ Primitives used directly
- **Status**: ✅ Implemented / ⚠️ Partial / ❌ Missing
- **Quality Score** (1-10): Overall value object quality
- **Improvements**: AI-suggested improvements
- **Evidence**: File path + value object definition

**Deterministic Baseline**:
```bash
# Potential value objects (types in domain/shared or domain/*/value_objects.go)
VALUE_OBJECT_FILES=$(find internal/domain/ -name "*value_objects.go" -o -name "shared.go" 2>/dev/null | wc -l)

# Types that look like value objects (no ID field, simple types)
POTENTIAL_VOS=$(grep -r "^type [A-Z].*struct {" internal/domain/ --include="*.go" ! -name "*_test.go" -A 5 | grep -v "id.*uuid\|ID.*uuid" | grep -c "^type")

# Known value objects
EMAIL_VO=$(grep -r "type Email struct\|type EmailAddress" internal/domain/ --include="*.go" | wc -l)
PHONE_VO=$(grep -r "type Phone struct\|type PhoneNumber" internal/domain/ --include="*.go" | wc -l)
MONEY_VO=$(grep -r "type Money struct" internal/domain/ --include="*.go" | wc -l)
HEX_COLOR_VO=$(grep -r "type HexColor struct\|type Color" internal/domain/ --include="*.go" | wc -l)

# Primitive obsession detection (using string/int directly instead of value objects)
EMAIL_AS_STRING=$(grep -r "email.*string\|Email.*string" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)
PHONE_AS_STRING=$(grep -r "phone.*string\|Phone.*string" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)
```

**AI Analysis**:
- Identify all value objects in domain layer
- Check immutability (all fields private, no setters)
- Verify constructor validation
- Check value equality implementation
- Detect primitive obsession (using primitives instead of value objects)
- Score quality (1-10)

---

## Value Object Quality Checklist

For each value object, check:

1. **Immutability** (✅ required):
   ```go
   type Email struct {
       value string  // ✅ Unexported field (immutable)
   }
   ```

2. **Constructor Validation** (✅ required):
   ```go
   func NewEmail(email string) (Email, error) {
       if !isValidEmail(email) {
           return Email{}, ErrInvalidEmail
       }
       return Email{value: email}, nil
   }
   ```

3. **Value Equality** (✅ recommended):
   ```go
   func (e Email) Equals(other Email) bool {
       return e.value == other.value
   }
   ```

4. **No Setters** (✅ required)
5. **Getter** (✅ required for accessing value)
6. **String representation** (⚠️ optional but recommended)

---

## Chain of Thought Workflow

### Step 0: Run Deterministic Analysis (5 min)

```bash
cd /home/caloi/ventros-crm

# Execute deterministic script
bash scripts/analyze_codebase.sh

# Extract value object metrics
VALUE_OBJECT_FILES=$(find internal/domain/ -name "*value_objects.go" -o -name "shared.go" 2>/dev/null | wc -l)
MONEY_VO=$(grep -r "type Money struct" internal/domain/ --include="*.go" | wc -l)
HEX_COLOR_VO=$(grep -r "type HexColor struct" internal/domain/ --include="*.go" | wc -l)
EMAIL_AS_STRING=$(grep -r "email.*string" internal/domain/ --include="*.go" ! -name "*_test.go" | wc -l)

echo "✅ Baseline: $VALUE_OBJECT_FILES VO files, Money: $MONEY_VO, HexColor: $HEX_COLOR_VO, Email primitive obsession: $EMAIL_AS_STRING"
```

---

### Step 1: Identify Explicit Value Objects (10 min)

Find types explicitly defined as value objects.

```bash
# Find value object files
echo "=== Value Object Files ===" > /tmp/vo_analysis.txt
find internal/domain/ -name "*value_objects.go" -o -name "shared.go" -o -path "*/shared/*.go" 2>/dev/null >> /tmp/vo_analysis.txt

# Extract value object definitions
echo "=== Value Object Definitions ===" >> /tmp/vo_analysis.txt
for file in $(find internal/domain/ -name "*value_objects.go" -o -path "*/shared/*.go" 2>/dev/null | grep -v "_test.go"); do
  if [ -f "$file" ]; then
    echo "--- $file ---" >> /tmp/vo_analysis.txt
    grep -n "^type [A-Z].*struct {" "$file" -A 10 >> /tmp/vo_analysis.txt
  fi
done

cat /tmp/vo_analysis.txt
```

**AI Analysis**: For each type, check if it's truly a value object.

---

### Step 2: Detect Primitive Obsession (5-10 min)

Find places using primitives instead of value objects.

```bash
# Find primitive usage where value objects should be used
echo "=== Primitive Obsession ===" > /tmp/primitive_obsession.txt

# Email as string
echo "=== Email (should be value object) ===" >> /tmp/primitive_obsession.txt
grep -rn "email.*string\|Email.*string" internal/domain/ --include="*.go" ! -name "*_test.go" | head -20 >> /tmp/primitive_obsession.txt

# Phone as string
echo "=== Phone (should be value object) ===" >> /tmp/primitive_obsession.txt
grep -rn "phone.*string\|Phone.*string" internal/domain/ --include="*.go" ! -name "*_test.go" | head -20 >> /tmp/primitive_obsession.txt

# Money as float64
echo "=== Money (should be value object) ===" >> /tmp/primitive_obsession.txt
grep -rn "amount.*float64\|price.*float64\|balance.*float64" internal/domain/ --include="*.go" ! -name "*_test.go" | head -20 >> /tmp/primitive_obsession.txt

cat /tmp/primitive_obsession.txt
```

**AI Analysis**: Identify all primitive obsession instances, suggest value object conversions.

---

### Step 3: Analyze Value Object Quality (10 min)

For each identified value object, check quality.

```bash
# Check immutability (unexported fields)
echo "=== Immutability Check ===" > /tmp/vo_quality.txt
for file in $(find internal/domain/ -path "*/shared/*.go" 2>/dev/null | grep -v "_test.go"); do
  if [ -f "$file" ]; then
    echo "--- $file ---" >> /tmp/vo_quality.txt
    grep -n "^type [A-Z].*struct {" "$file" -A 5 >> /tmp/vo_quality.txt
  fi
done

# Check constructor validation
echo "=== Constructor Validation ===" >> /tmp/vo_quality.txt
grep -rn "^func New[A-Z].*(" internal/domain/core/shared/ --include="*.go" -A 10 2>/dev/null >> /tmp/vo_quality.txt

# Check value equality
echo "=== Equals Methods ===" >> /tmp/vo_quality.txt
grep -rn "func.*Equals(" internal/domain/ --include="*.go" -A 5 | head -50 >> /tmp/vo_quality.txt

cat /tmp/vo_quality.txt
```

**AI Analysis**: Score each value object (immutability, validation, equality).

---

### Step 4: Generate Report (5 min)

Combine all analysis into structured markdown report.

---

## Critical Rules

1. **Atemporal analysis** - NO hardcoded numbers
2. **Deterministic baseline first** - Run script first
3. **Comparison** - Show "Deterministic vs AI"
4. **Evidence required** - File:line + code snippets
5. **Score with reasoning** - Explain 1-10 scores
6. **Identify primitive obsession** - Find where VOs should be used

---

## Success Criteria

- ✅ Table 6 generated (Value Objects)
- ✅ Deterministic baseline compared with AI analysis
- ✅ All value objects identified
- ✅ Primitive obsession detected
- ✅ Quality scores provided (1-10)
- ✅ Recommendations for improvements
- ✅ Output to `code-analysis/domain/value_objects_analysis.md`

---

**Agent Version**: 1.0 (Value Objects)
**Estimated Runtime**: 30-40 minutes
**Output File**: `code-analysis/domain/value_objects_analysis.md`
**Last Updated**: 2025-10-15
