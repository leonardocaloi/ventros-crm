#!/bin/bash
set -euo pipefail

# Generate Codebase Metrics (Fast & Deterministic)
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

OUTPUT_FILE=".claude/analysis/codebase_metrics.json"
mkdir -p .claude/analysis

START_TIME=$(date +%s)

echo "📊 Generating codebase metrics (fast mode)..."

# Domain - Count aggregate directories
CRM_AGGREGATES=$(find internal/domain/crm -maxdepth 1 -mindepth 1 -type d | wc -l)
AUTOMATION_AGGREGATES=$(find internal/domain/automation -maxdepth 1 -mindepth 1 -type d 2>/dev/null | wc -l)
CORE_AGGREGATES=$(find internal/domain/core -maxdepth 1 -mindepth 1 -type d 2>/dev/null | wc -l)
TOTAL_AGGREGATES=$((CRM_AGGREGATES + AUTOMATION_AGGREGATES + CORE_AGGREGATES))

# API - Count endpoints in routes.go
TOTAL_ENDPOINTS=$(cat infrastructure/http/routes/routes.go | grep -E "router\.(GET|POST|PUT|DELETE|PATCH)" | wc -l)

# CQRS
TOTAL_COMMANDS=$(find internal/application/commands -name "*_handler.go" 2>/dev/null | wc -l)
TOTAL_QUERIES=$(find internal/application/queries -name "*.go" -type f 2>/dev/null | wc -l)

# Testing
TOTAL_TESTS=$(find . -name "*_test.go" -exec cat {} \; | grep -c "^func Test" || echo "0")

# Agents
TOTAL_AGENTS=$(find .claude/agents -name "*.md" ! -name "README.md" | wc -l)
CRM_AGENTS=$(find .claude/agents -name "crm_*.md" | wc -l)
GLOBAL_AGENTS=$(find .claude/agents -name "global_*.md" | wc -l)
META_AGENTS=$(find .claude/agents -name "meta_*.md" | wc -l)
MGMT_AGENTS=$(find .claude/agents -name "mgmt_*.md" | wc -l)

# Infrastructure
TOTAL_MIGRATIONS=$(find infrastructure/database/migrations -name "*.up.sql" 2>/dev/null | wc -l)

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Generate JSON
cat > "$OUTPUT_FILE" <<EOF
{
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "scan_duration_seconds": $DURATION,
  "domain": {
    "aggregates": {
      "total": $TOTAL_AGGREGATES,
      "by_bounded_context": {
        "crm": $CRM_AGGREGATES,
        "automation": $AUTOMATION_AGGREGATES,
        "core": $CORE_AGGREGATES
      }
    }
  },
  "api": {
    "endpoints": {
      "total": $TOTAL_ENDPOINTS
    }
  },
  "cqrs": {
    "commands": $TOTAL_COMMANDS,
    "queries": $TOTAL_QUERIES
  },
  "testing": {
    "total_tests": $TOTAL_TESTS
  },
  "agents": {
    "total": $TOTAL_AGENTS,
    "by_category": {
      "crm": $CRM_AGENTS,
      "global": $GLOBAL_AGENTS,
      "meta": $META_AGENTS,
      "mgmt": $MGMT_AGENTS
    }
  },
  "infrastructure": {
    "migrations": $TOTAL_MIGRATIONS
  }
}
EOF

echo "✅ Metrics: $OUTPUT_FILE"
echo "📊 Aggregates: $TOTAL_AGGREGATES | Endpoints: $TOTAL_ENDPOINTS | Agents: $TOTAL_AGENTS | Tests: $TOTAL_TESTS"
