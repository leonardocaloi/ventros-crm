---
name: workflows_analyzer
description: |
  Analyzes Temporal workflows and sagas (Table 12).
  Output: code-analysis/domain/workflows_analysis.md
tools: Read, Grep, Glob, Bash, Write
model: sonnet
priority: medium
---

# Workflows Analyzer

Analyzes Temporal workflows for orchestration and saga patterns.

## Workflow
```bash
# Find all workflows
find internal/workflows -name "*_workflow.go"

# Find all activities
find internal/workflows -name "*_activities.go"

# Check saga compensations
grep -r "Compensate\|Rollback" internal/workflows/
```

## Output
Table 12: All workflows with activities, compensations, error handling.

**Runtime**: 40 minutes
