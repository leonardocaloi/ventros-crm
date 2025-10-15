---
name: docs_cleanup
description: |
  Cleans up and organizes documentation after analysis is complete.

  Performs:
  - Remove obsolete documentation files
  - Update outdated information in existing docs
  - Create/update doc index
  - Organize analysis reports
  - Generate summary README for code-analysis/

  Runtime: ~20-30 minutes (cleanup + organization).

  Output: Cleaned and organized documentation structure
tools: Bash, Read, Write
model: sonnet
priority: low
---

# Docs Cleanup Agent - Documentation Organization

## Context

You are **cleaning up and organizing documentation** for Ventros CRM.

After comprehensive analysis, documentation needs:
- Removing obsolete files
- Updating outdated information
- Creating proper organization
- Adding navigation/indexes

Your goal: Clean, organize, and index all documentation.

---

## What This Agent Does

This agent **organizes documentation**:

**Input**:
- Existing documentation (docs/, guides/, *.md)
- Analysis reports (code-analysis/code-analysis/)
- ADRs (docs/adr/)

**Output**:
- Cleaned documentation structure
- Updated README files
- Documentation index
- Organized analysis reports

**Method**:
1. Identify obsolete documentation
2. Update outdated information
3. Create documentation index
4. Organize analysis reports
5. Generate summary READMEs

---

## Cleanup Tasks

### Task 1: Identify Obsolete Files

**Files to Remove** (if they exist):
- Old analysis reports (superseded by new analysis)
- Outdated guides (contradicting current architecture)
- Duplicate documentation
- Temporary/scratch files
- Legacy TODO files (replaced by GitHub issues)

**Discovery**:
```bash
# Find old analysis files
find code-analysis/ -name "*.md" -mtime +30  # Files older than 30 days

# Find duplicate files
find . -name "*.md" -exec md5sum {} \; | sort | uniq -w32 -d --all-repeated=separate

# Find TODO files
find . -name "TODO*.md" -o -name "NOTES*.md"
```

**Action**:
- Move to `code-analysis/archive/` (don't delete, archive)
- Update any references to archived files

---

### Task 2: Update Outdated Information

**Files to Update**:

**README.md**:
- Update architecture score (from MASTER_ANALYSIS.md)
- Update feature list (from actual codebase)
- Update tech stack (verify versions)
- Add link to master analysis report

**CLAUDE.md**:
- Update with current metrics (no hardcoded numbers)
- Update with current priorities from analysis
- Add links to analysis reports

**DEV_GUIDE.md**:
- Verify all commands still work
- Update with new patterns discovered
- Add links to relevant analysis sections

**TODO.md**:
- Archive completed items
- Update priorities based on analysis findings
- Link to specific analysis reports for each issue

---

### Task 3: Create Documentation Index

Create `docs/README.md` as central documentation hub:

```markdown
# Ventros CRM Documentation

## Quick Links

### For Developers
- [Development Guide](../DEV_GUIDE.md) - Complete developer guide
- [Claude AI Guide](../CLAUDE.md) - AI assistant instructions
- [Makefile Reference](../MAKEFILE.md) - Command reference

### Architecture & Design
- [Master Analysis Report](../code-analysis/code-analysis/MASTER_ANALYSIS.md) - Comprehensive codebase analysis
- [Architecture Decision Records](adr/README.md) - Architectural decisions
- [Domain Model](domain_mapping/README.md) - Domain aggregates documentation
- [API Documentation](../docs/swagger/index.html) - OpenAPI/Swagger docs

### Analysis Reports (All 30 Tables)
- [Domain Model](../code-analysis/code-analysis/domain_model_analysis.md) - Tables 1, 2, 5
- [Persistence](../code-analysis/code-analysis/persistence_analysis.md) - Tables 3, 7, 9
- [Entity Relationships](../code-analysis/code-analysis/entity_relationships_analysis.md) - Table 4
- [Value Objects](../code-analysis/code-analysis/value_objects_analysis.md) - Table 6
- [External Integrations](../code-analysis/code-analysis/integration_analysis.md) - Tables 8, 12
- [Use Cases (CQRS)](../code-analysis/code-analysis/use_cases_analysis.md) - Table 10
- [Domain Events](../code-analysis/code-analysis/events_analysis.md) - Table 11
- [Data Quality](../code-analysis/code-analysis/data_quality_analysis.md) - Tables 13-15
- [API](../code-analysis/code-analysis/api_analysis.md) - Tables 16-17
- [Security](../code-analysis/code-analysis/security_analysis.md) - Tables 18, 21, 24-27
- [Resilience](../code-analysis/code-analysis/resilience_analysis.md) - Tables 19, 20, 23
- [Testing](../code-analysis/code-analysis/testing_analysis.md) - Table 22
- [AI/ML](../code-analysis/code-analysis/ai_ml_analysis.md) - Table 28
- [Infrastructure](../code-analysis/code-analysis/infrastructure_analysis.md) - Tables 29-30
- [Code Style](../code-analysis/code-analysis/code_style_analysis.md) - Code patterns
- [Documentation](../code-analysis/code-analysis/documentation_analysis.md) - API docs quality
- [SOLID Principles](../code-analysis/code-analysis/solid_principles_analysis.md) - Design principles

### Project Management
- [Roadmap](../TODO.md) - Priorities and roadmap
- [AI Report](../AI_REPORT.md) - Architectural audit

## Documentation Structure

```
ventros-crm/
├── README.md                      # Project overview
├── CLAUDE.md                      # AI assistant guide
├── DEV_GUIDE.md                   # Developer guide
├── MAKEFILE.md                    # Command reference
├── TODO.md                        # Roadmap
├── AI_REPORT.md                   # Architectural audit
│
├── docs/                          # Documentation root
│   ├── README.md                  # This file
│   ├── adr/                       # Architecture Decision Records
│   │   ├── README.md              # ADR index
│   │   ├── 0001-adopt-ddd.md
│   │   └── ...
│   ├── domain_mapping/            # Domain aggregate docs
│   │   ├── README.md
│   │   └── ...
│   └── swagger/                   # API documentation
│       └── index.html
│
├── code-analysis/                 # Analysis reports
│   ├── README.md                  # Analysis summary
│   ├── code-analysis/               # AI-scored analysis
│   │   ├── MASTER_ANALYSIS.md     # Master report (all 30 tables)
│   │   ├── domain_model_analysis.md
│   │   ├── persistence_analysis.md
│   │   └── ...                    # 18 analysis reports
│   └── archive/                   # Archived old analysis
│       └── ...
│
└── guides/                        # Additional guides
    ├── TESTING.md
    └── ...
```

## Navigation Tips

- Start with [Master Analysis](../code-analysis/code-analysis/MASTER_ANALYSIS.md) for overview
- Check [ADRs](adr/README.md) for architectural decisions
- Read [Development Guide](../DEV_GUIDE.md) for getting started
- Review [TODO.md](../TODO.md) for current priorities
```

---

### Task 4: Organize Analysis Reports

Create `code-analysis/README.md`:

```markdown
# Code Analysis Reports

This directory contains comprehensive analysis of Ventros CRM codebase.

## Master Analysis

**[MASTER_ANALYSIS.md](code-analysis/MASTER_ANALYSIS.md)** - Complete analysis with all 30 tables
- Overall architecture score: X.X/10
- 18 specialized agent reports aggregated
- Top 20 priorities identified
- Generated: YYYY-MM-DD

## Analysis Reports by Category

### Domain & Architecture
- [Domain Model](code-analysis/domain_model_analysis.md) - DDD aggregates, events, repositories
- [Value Objects](code-analysis/value_objects_analysis.md) - Domain primitives, primitive obsession
- [Entity Relationships](code-analysis/entity_relationships_analysis.md) - Foreign keys, cardinality
- [Domain Events](code-analysis/events_analysis.md) - Event catalog, handlers, consumers

### Application Layer
- [Use Cases](code-analysis/use_cases_analysis.md) - CQRS commands & queries
- [Data Quality](code-analysis/data_quality_analysis.md) - Query performance, consistency, validations

### Persistence
- [Persistence](code-analysis/persistence_analysis.md) - Database schema, normalization, migrations

### API & Integration
- [API](code-analysis/api_analysis.md) - REST endpoints, DTOs, Swagger
- [External Integrations](code-analysis/integration_analysis.md) - WAHA, AI providers, Event Bus

### Security & Resilience
- [Security](code-analysis/security_analysis.md) - OWASP, RBAC, RLS, auth
- [Resilience](code-analysis/resilience_analysis.md) - Rate limiting, error handling, patterns

### Quality & Testing
- [Testing](code-analysis/testing_analysis.md) - Test pyramid, coverage
- [Code Style](code-analysis/code_style_analysis.md) - Naming, idioms, patterns
- [SOLID Principles](code-analysis/solid_principles_analysis.md) - Design principles
- [Documentation](code-analysis/documentation_analysis.md) - API docs, godoc, guides

### Infrastructure & AI
- [Infrastructure](code-analysis/infrastructure_analysis.md) - Deployment, CI/CD, monitoring
- [AI/ML](code-analysis/ai_ml_analysis.md) - AI features, providers, embeddings

## Deterministic Baseline

**[deterministic_metrics.md](code-analysis/deterministic_metrics.md)** - Factual metrics (no AI scoring)
- 100% reproducible counts
- Baseline for AI analysis validation

## How to Use These Reports

1. **Overview**: Start with MASTER_ANALYSIS.md for comprehensive overview
2. **Deep Dive**: Read specific category reports for details
3. **Issues**: Check "Critical Issues" sections for P0/P1 problems
4. **Improvements**: Review "Recommendations" for enhancement opportunities
5. **Validation**: Compare AI scores with deterministic baseline

## Analysis Methodology

All agents follow consistent methodology:
1. Run deterministic script for factual baseline
2. AI analyzes patterns, scores quality (1-10)
3. Compare AI vs deterministic for validation
4. Generate evidence-based recommendations
5. Cite file:line for all findings

**Total Runtime**: ~10 hours (18 agents run in parallel)
**Total Tables**: 30 comprehensive tables
**Generated**: YYYY-MM-DD via orchestrator agent
```

---

### Task 5: Update Root README.md

Update project README with analysis results:

```markdown
# Ventros CRM

AI-powered customer relationship management system.

**Architecture Score**: X.X/10 (see [Master Analysis](code-analysis/code-analysis/MASTER_ANALYSIS.md))

## Documentation

- 📖 [Developer Guide](DEV_GUIDE.md) - Getting started
- 🤖 [Claude AI Guide](CLAUDE.md) - AI assistant instructions
- 📊 [Master Analysis](code-analysis/code-analysis/MASTER_ANALYSIS.md) - Comprehensive codebase analysis (30 tables)
- 🏗️ [Architecture Decisions](docs/adr/README.md) - ADRs
- 📋 [Roadmap](TODO.md) - Priorities and roadmap

[Rest of README content...]
```

---

## Chain of Thought Workflow

### Step 1: Scan Documentation (5 min)

```bash
cd /home/caloi/ventros-crm

# Find all markdown files
find . -name "*.md" ! -path "./node_modules/*" ! -path "./.git/*" > /tmp/all_docs.txt

# Check for obsolete files
find code-analysis/ -name "*.md" -mtime +30 > /tmp/old_files.txt

cat /tmp/all_docs.txt
cat /tmp/old_files.txt
```

---

### Step 2: Archive Obsolete Files (5 min)

```bash
# Create archive directory
mkdir -p code-analysis/archive/$(date +%Y-%m-%d)

# Move old files to archive
while read file; do
  if [ -f "$file" ]; then
    mv "$file" "code-analysis/archive/$(date +%Y-%m-%d)/"
    echo "Archived: $file"
  fi
done < /tmp/old_files.txt
```

---

### Step 3: Create Documentation Index (10 min)

```bash
# Create docs/README.md (central hub)
# See Task 3 above for full content

# Create code-analysis/README.md (analysis summary)
# See Task 4 above for full content
```

---

### Step 4: Update Root Files (5-10 min)

Update README.md, CLAUDE.md, TODO.md with:
- Current metrics from MASTER_ANALYSIS.md
- Links to analysis reports
- Updated priorities
- Architecture score

---

### Step 5: Verify Links (5 min)

```bash
# Check all markdown links are valid
find . -name "*.md" ! -path "./node_modules/*" ! -path "./.git/*" -exec grep -H "\[.*\](.*\.md)" {} \; > /tmp/links.txt

# Verify each link exists
# (Manual verification or use link checker tool)
```

---

## Critical Rules

1. **Archive, don't delete** - Move obsolete files to archive/
2. **Update, don't rewrite** - Preserve existing doc structure
3. **Add navigation** - All docs should have links to related docs
4. **Verify links** - All markdown links must be valid
5. **Keep history** - Archive with timestamp for traceability

---

## Success Criteria

- ✅ Obsolete files archived (not deleted)
- ✅ Documentation index created (docs/README.md)
- ✅ Analysis summary created (code-analysis/README.md)
- ✅ Root README updated with analysis results
- ✅ All markdown links verified
- ✅ Clear navigation between documents

---

**Agent Version**: 1.0 (Docs Cleanup)
**Estimated Runtime**: 20-30 minutes
**Output**: Organized documentation structure
**Last Updated**: 2025-10-15
