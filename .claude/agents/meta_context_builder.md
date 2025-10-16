---
name: meta_context_builder
description: |
  Builds complete architectural context before feature implementation.
  Identifies area (import, API, workflow), suggests patterns (Saga, Temporal, Coreography, simple).
  Reads ALL analysis files and creates intelligent recommendations.
tools: Read, Bash, Write
model: sonnet
priority: high
---

# Context Builder - Intelligent Pattern Recommendation

**Purpose**: Build complete context from analysis files and recommend best architectural patterns
**Called by**: `meta_dev_orchestrator` (before implementation)
**Input**: User request + Analysis files (if exist)
**Output**: `/tmp/context_recommendation.md` with pattern suggestions

---

## 🎯 Responsibilities

### 1. Load All Available Context
- Read `.claude/analysis/*.json` (if exists)
- Read `CLAUDE.md`, `DEV_GUIDE.md`, `AI_REPORT.md`
- Read `AGENT_STATE.json`, `P0_ACTIVE_WORK.md`

### 2. Identify Request Area
Classify request into one of these areas:

| Area | Examples | Patterns to Consider |
|------|----------|---------------------|
| **Import** | "Import contacts", "Bulk import", "Historical data" | Temporal Workflow > Saga > Simple |
| **API Endpoint** | "Add GET endpoint", "New REST API" | Simple Handler > CQRS |
| **Background Job** | "Send daily report", "Cleanup old data" | Temporal Workflow > Cron |
| **Event Processing** | "Process webhook", "Handle event" | Coreography > Saga |
| **Aggregate** | "Add new entity", "Create aggregate" | DDD Pattern (always) |
| **Integration** | "Connect to X", "Call external API" | Circuit Breaker + Retry |
| **Workflow** | "Multi-step process", "Complex flow" | Temporal > Saga > Coreography |

### 3. Suggest Pattern Based on Complexity

#### Decision Tree:

```
Is it multi-step with compensation needed?
├─ YES → Is it long-running (>5 min)?
│   ├─ YES → Temporal Workflow (Activity pattern)
│   └─ NO → Saga Pattern (orchestration)
└─ NO → Is it event-driven?
    ├─ YES → Coreography (event bus)
    └─ NO → Simple code (command handler)
```

#### Complexity Scoring:

| Factor | Score | Pattern Suggestion |
|--------|-------|-------------------|
| Steps | 1-2 steps | Simple code |
| Steps | 3-5 steps | Saga pattern |
| Steps | 6+ steps | Temporal Workflow |
| Duration | <1 min | Saga or Simple |
| Duration | 1-5 min | Temporal Activity |
| Duration | >5 min | Temporal Workflow |
| External APIs | 0 | Any pattern |
| External APIs | 1-2 | Circuit Breaker |
| External APIs | 3+ | Temporal (retry built-in) |
| Compensation | Not needed | Simple/Coreography |
| Compensation | Needed | Saga or Temporal |

---

## 📋 Execution Workflow

### Phase 1: Load Context (2-3 min)

```bash
echo "📚 Loading architectural context..."

# 1. Check if analysis exists
if [ -f .claude/analysis/last_run.json ]; then
  ANALYSIS_EXISTS=true
  ANALYSIS_MODE=$(jq -r '.mode' .claude/analysis/last_run.json)
  ANALYSIS_AGE=$(jq -r '.timestamp' .claude/analysis/last_run.json)

  echo "✅ Found pre-analysis (mode: $ANALYSIS_MODE, age: $ANALYSIS_AGE)"

  # Load all analysis files
  DOMAIN_ANALYSIS=$(cat .claude/analysis/domain_model.json 2>/dev/null || echo "{}")
  PERSISTENCE_ANALYSIS=$(cat .claude/analysis/persistence.json 2>/dev/null || echo "{}")
  API_ANALYSIS=$(cat .claude/analysis/api.json 2>/dev/null || echo "{}")
  WORKFLOWS_ANALYSIS=$(cat .claude/analysis/workflows.json 2>/dev/null || echo "{}")
  INTEGRATION_ANALYSIS=$(cat .claude/analysis/integration.json 2>/dev/null || echo "{}")
else
  ANALYSIS_EXISTS=false
  echo "⚠️  No pre-analysis found. Recommending: run /pre-analyze first"
fi

# 2. Load architectural documents
CLAUDE_MD=$(cat /home/caloi/ventros-crm/CLAUDE.md)
DEV_GUIDE=$(cat /home/caloi/ventros-crm/DEV_GUIDE.md)
AI_REPORT=$(cat /home/caloi/ventros-crm/AI_REPORT.md)

# 3. Load state
AGENT_STATE=$(cat /home/caloi/ventros-crm/.claude/AGENT_STATE.json)

echo "✅ Context loaded"
```

---

### Phase 2: Identify Area (1-2 min)

```bash
echo "🔍 Analyzing request area..."

USER_REQUEST="$1"

# Detect keywords for area classification
if echo "$USER_REQUEST" | grep -qiE "import|bulk|historical|migrate"; then
  AREA="import"
  COMPLEXITY="high"
  RECOMMENDED_PATTERN="temporal_workflow"

elif echo "$USER_REQUEST" | grep -qiE "endpoint|api|rest|get|post|http"; then
  AREA="api"
  COMPLEXITY="low"
  RECOMMENDED_PATTERN="simple_handler"

elif echo "$USER_REQUEST" | grep -qiE "background|job|cron|scheduled|daily"; then
  AREA="background_job"
  COMPLEXITY="medium"
  RECOMMENDED_PATTERN="temporal_workflow"

elif echo "$USER_REQUEST" | grep -qiE "webhook|event|process.*event|handle.*event"; then
  AREA="event_processing"
  COMPLEXITY="medium"
  RECOMMENDED_PATTERN="coreography"

elif echo "$USER_REQUEST" | grep -qiE "aggregate|entity|domain|value object"; then
  AREA="aggregate"
  COMPLEXITY="high"
  RECOMMENDED_PATTERN="ddd_pattern"

elif echo "$USER_REQUEST" | grep -qiE "integrate|connect|external|third-party"; then
  AREA="integration"
  COMPLEXITY="high"
  RECOMMENDED_PATTERN="circuit_breaker_retry"

elif echo "$USER_REQUEST" | grep -qiE "workflow|multi-step|process|flow"; then
  AREA="workflow"
  COMPLEXITY="high"
  RECOMMENDED_PATTERN="temporal_workflow"

else
  AREA="general"
  COMPLEXITY="medium"
  RECOMMENDED_PATTERN="auto_detect"
fi

echo "Area: $AREA"
echo "Complexity: $COMPLEXITY"
echo "Initial pattern suggestion: $RECOMMENDED_PATTERN"
```

---

### Phase 3: Deep Analysis (2-3 min)

```bash
echo "🧠 Deep analysis based on existing patterns..."

# Analyze existing patterns from analysis files
if [ "$ANALYSIS_EXISTS" = true ]; then
  # Check existing workflows
  TEMPORAL_WORKFLOWS=$(echo "$WORKFLOWS_ANALYSIS" | jq -r '.temporal_workflows | length')
  SAGAS=$(echo "$WORKFLOWS_ANALYSIS" | jq -r '.sagas | length')
  COREOGRAPHY=$(echo "$WORKFLOWS_ANALYSIS" | jq -r '.coreography_patterns | length')

  echo "Existing patterns:"
  echo "  - Temporal Workflows: $TEMPORAL_WORKFLOWS"
  echo "  - Sagas: $SAGAS"
  echo "  - Coreography: $COREOGRAPHY"

  # Check similar existing features
  if [ "$AREA" = "import" ]; then
    # Check if there's already an import workflow
    IMPORT_WORKFLOWS=$(echo "$WORKFLOWS_ANALYSIS" | jq -r '.temporal_workflows[] | select(.name | contains("import") or contains("Import"))')

    if [ -n "$IMPORT_WORKFLOWS" ]; then
      echo "✅ Found existing import workflow pattern!"
      echo "Recommendation: Follow same pattern"
      RECOMMENDED_PATTERN="temporal_workflow"
      REFERENCE_WORKFLOW=$(echo "$IMPORT_WORKFLOWS" | jq -r '.[0].name')
      REFERENCE_FILE=$(echo "$IMPORT_WORKFLOWS" | jq -r '.[0].file')
    fi
  fi
fi

# Analyze request complexity
STEPS_COUNT=0
EXTERNAL_APIS=0
DURATION_ESTIMATE="unknown"
NEEDS_COMPENSATION=false

# Count steps (estimate)
if echo "$USER_REQUEST" | grep -qE "then|after|next|subsequently"; then
  STEPS_COUNT=$((STEPS_COUNT + 2))
fi

if echo "$USER_REQUEST" | grep -qE "multiple|batch|bulk|all"; then
  STEPS_COUNT=$((STEPS_COUNT + 2))
fi

# Check external APIs
if echo "$USER_REQUEST" | grep -qiE "waha|stripe|meta|facebook|external|api|webhook"; then
  EXTERNAL_APIS=$((EXTERNAL_APIS + 1))
fi

# Check compensation
if echo "$USER_REQUEST" | grep -qiE "rollback|undo|revert|compensate"; then
  NEEDS_COMPENSATION=true
fi

echo "Complexity metrics:"
echo "  - Estimated steps: $STEPS_COUNT"
echo "  - External APIs: $EXTERNAL_APIS"
echo "  - Needs compensation: $NEEDS_COMPENSATION"
```

---

### Phase 4: Pattern Recommendation (1 min)

```bash
echo "🎯 Generating pattern recommendation..."

# Decision logic
if [ "$AREA" = "import" ] && [ $STEPS_COUNT -ge 3 ]; then
  FINAL_PATTERN="temporal_workflow"
  RATIONALE="Multi-step import with potential long duration. Temporal provides: retry, timeout, visibility."

elif [ "$AREA" = "workflow" ] && [ "$NEEDS_COMPENSATION" = true ]; then
  FINAL_PATTERN="saga_pattern"
  RATIONALE="Multi-step workflow with compensation needed. Saga provides orchestrated rollback."

elif [ "$AREA" = "event_processing" ]; then
  FINAL_PATTERN="coreography"
  RATIONALE="Event-driven flow. Coreography provides loose coupling via event bus."

elif [ "$AREA" = "api" ]; then
  FINAL_PATTERN="simple_handler"
  RATIONALE="Simple API endpoint. Command handler pattern sufficient."

elif [ $EXTERNAL_APIS -ge 2 ]; then
  FINAL_PATTERN="temporal_workflow"
  RATIONALE="Multiple external APIs. Temporal provides built-in retry and circuit breaking."

else
  FINAL_PATTERN="$RECOMMENDED_PATTERN"
  RATIONALE="Based on area classification and complexity analysis."
fi

echo "Final pattern: $FINAL_PATTERN"
echo "Rationale: $RATIONALE"
```

---

### Phase 5: Generate Recommendation Document (1 min)

```bash
cat > /tmp/context_recommendation.md << EOF
# Architectural Context & Pattern Recommendation

**Generated**: $(date +"%Y-%m-%d %H:%M:%S")
**Request**: $USER_REQUEST
**Area**: $AREA
**Complexity**: $COMPLEXITY

---

## 📊 Analysis Summary

### Pre-Analysis Status
$(if [ "$ANALYSIS_EXISTS" = true ]; then
  echo "✅ Pre-analysis available (mode: $ANALYSIS_MODE)"
  echo "- Domain: $(echo "$DOMAIN_ANALYSIS" | jq -r '.summary.total_aggregates') aggregates"
  echo "- API: $(echo "$API_ANALYSIS" | jq -r '.summary.total_endpoints') endpoints"
  echo "- Workflows: $TEMPORAL_WORKFLOWS Temporal, $SAGAS Sagas"
else
  echo "⚠️ No pre-analysis found"
  echo "**Recommendation**: Run \`/pre-analyze --quick\` first for better context"
fi)

### Request Classification
- **Area**: $AREA
- **Estimated Steps**: $STEPS_COUNT
- **External APIs**: $EXTERNAL_APIS
- **Needs Compensation**: $NEEDS_COMPENSATION

---

## 🎯 Recommended Pattern: **$FINAL_PATTERN**

### Rationale
$RATIONALE

### Implementation Approach

$(case "$FINAL_PATTERN" in
  "temporal_workflow")
    echo "#### Temporal Workflow Pattern"
    echo ""
    echo "**Structure**:"
    echo "\`\`\`"
    echo "internal/workflows/"
    echo "├── <feature>_workflow.go       # Workflow definition"
    echo "├── <feature>_activities.go     # Activities (steps)"
    echo "└── <feature>_worker.go         # Worker registration"
    echo "\`\`\`"
    echo ""
    echo "**Key Components**:"
    echo "1. **Workflow**: Orchestrates activities"
    echo "2. **Activities**: Individual steps (can retry)"
    echo "3. **Worker**: Executes workflow"
    echo ""
    echo "**Benefits**:"
    echo "- Built-in retry and timeout"
    echo "- Visibility in Temporal UI"
    echo "- Durable execution"
    echo "- Easy to add compensation"
    echo ""
    echo "**Example**: See \`internal/workflows/channel/waha_history_import_workflow.go\`"
    ;;
  "saga_pattern")
    echo "#### Saga Pattern (Orchestration)"
    echo ""
    echo "**Structure**:"
    echo "\`\`\`"
    echo "internal/domain/core/saga/"
    echo "├── saga.go                     # Saga orchestrator"
    echo "├── saga_context.go             # Execution context"
    echo "└── <feature>_saga.go           # Feature-specific saga"
    echo "\`\`\`"
    echo ""
    echo "**Key Components**:"
    echo "1. **Saga**: Orchestrates steps + compensation"
    echo "2. **Steps**: Forward operations"
    echo "3. **Compensations**: Rollback operations"
    echo ""
    echo "**Benefits**:"
    echo "- Explicit compensation logic"
    echo "- Synchronous execution"
    echo "- Simpler than Temporal for short flows"
    echo ""
    echo "**Example**: See \`internal/domain/core/saga/\` directory"
    ;;
  "coreography")
    echo "#### Coreography Pattern (Event-Driven)"
    echo ""
    echo "**Structure**:"
    echo "\`\`\`"
    echo "internal/domain/.../events.go   # Domain events"
    echo "infrastructure/messaging/"
    echo "├── event_bus.go                # Event publisher"
    echo "└── <feature>_consumer.go       # Event subscriber"
    echo "\`\`\`"
    echo ""
    echo "**Key Components**:"
    echo "1. **Events**: Domain events emitted"
    echo "2. **EventBus**: Publishes to RabbitMQ"
    echo "3. **Consumers**: Subscribe and react"
    echo ""
    echo "**Benefits**:"
    echo "- Loose coupling"
    echo "- Async by nature"
    echo "- Easy to add new reactions"
    echo ""
    echo "**Example**: Outbox Pattern + RabbitMQ consumers"
    ;;
  "simple_handler")
    echo "#### Simple Command Handler Pattern"
    echo ""
    echo "**Structure**:"
    echo "\`\`\`"
    echo "internal/application/commands/<feature>/"
    echo "├── create_<feature>_command.go"
    echo "└── create_<feature>_handler.go"
    echo "\`\`\`"
    echo ""
    echo "**Key Components**:"
    echo "1. **Command**: Input validation"
    echo "2. **Handler**: Orchestration"
    echo "3. **Domain**: Business logic"
    echo ""
    echo "**Benefits**:"
    echo "- Simple and fast"
    echo "- No external dependencies"
    echo "- Easy to test"
    echo ""
    echo "**Example**: See \`internal/application/commands/contact/\`"
    ;;
esac)

---

## 📚 Relevant Context from Analysis

### Similar Existing Features
$(if [ "$ANALYSIS_EXISTS" = true ]; then
  if [ -n "$REFERENCE_WORKFLOW" ]; then
    echo "✅ **Reference Implementation Found**:"
    echo "- Workflow: \`$REFERENCE_WORKFLOW\`"
    echo "- File: \`$REFERENCE_FILE\`"
    echo "- **Recommendation**: Follow this pattern for consistency"
  else
    echo "⚠️ No similar feature found. This will be a new pattern."
  fi
else
  echo "⚠️ Run \`/pre-analyze\` to see similar features"
fi)

### Domain Context
$(if [ "$ANALYSIS_EXISTS" = true ]; then
  echo "$DOMAIN_ANALYSIS" | jq -r '.aggregates[] | "- \(.name) (\(.bounded_context)): \(.tests_coverage)% coverage"' | head -5
else
  echo "Run \`/pre-analyze\` to see domain context"
fi)

---

## ✅ Next Steps

1. **Review this recommendation** - Does the pattern make sense?
2. **Check reference implementation** (if provided)
3. **Proceed with /add-feature** - It will use this context
4. **Follow DEV_GUIDE** - For step-by-step implementation

---

**Recommendation Confidence**: $(if [ "$ANALYSIS_EXISTS" = true ]; then echo "HIGH (has pre-analysis)"; else echo "MEDIUM (no pre-analysis)"; fi)
**Generated by**: meta_context_builder v1.0
EOF

cat /tmp/context_recommendation.md
```

---

## 🎯 Output Example

```markdown
# Architectural Context & Pattern Recommendation

**Request**: Import historical WhatsApp messages from WAHA
**Area**: import
**Complexity**: high

---

## 🎯 Recommended Pattern: **Temporal Workflow**

### Rationale
Multi-step import with potential long duration. Temporal provides: retry, timeout, visibility.

### Implementation Approach

#### Temporal Workflow Pattern

**Structure**:
```
internal/workflows/message_import/
├── import_workflow.go       # Workflow definition
├── import_activities.go     # Activities (fetch, process, save)
└── import_worker.go         # Worker registration
```

**Key Components**:
1. **Workflow**: Orchestrates: fetch → process → save → notify
2. **Activities**:
   - FetchMessagesFromWAHA (retriable)
   - ProcessMessagesBatch (retriable)
   - SaveToDatabase (idempotent)
   - NotifyCompletion
3. **Worker**: Executes workflow

**Benefits**:
- Built-in retry (WAHA API can fail)
- Timeout (don't hang forever)
- Visibility in Temporal UI
- Can pause/resume import

**Example**: See `internal/workflows/channel/waha_history_import_workflow.go`

---

## 📚 Relevant Context

### Similar Existing Features
✅ **Reference Implementation Found**:
- Workflow: `WAHAHistoryImportWorkflow`
- File: `internal/workflows/channel/waha_history_import_workflow.go`
- **Recommendation**: Follow this pattern for consistency

### Domain Context
- Contact (crm): 100% coverage
- Message (crm): 95% coverage
- Session (crm): 88% coverage
- Channel (crm): 92% coverage

---

**Recommendation Confidence**: HIGH (has pre-analysis)
```

---

**Agent Version**: 1.0
**Intelligence**: Pattern recognition + similarity matching
**Dependencies**: Requires analysis files for best results
