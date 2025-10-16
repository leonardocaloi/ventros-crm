---
name: update-indexes
description: Update all documentation indexes (README.md files)
---

# Update Indexes Command

Run the `mgmt_docs_index_manager` agent to automatically update all README.md indexes across the project.

## What This Does

1. **Updates Root README.md**:
   - Scans root directory for markdown files
   - Updates "Project Documentation" section
   - Adds file sizes, line counts, last modified dates
   - Maintains accurate file count

2. **Updates code-analysis/README.md**:
   - Scans all subdirectories (architecture, domain, infrastructure, quality, ai-ml)
   - Updates analysis catalog with latest outputs
   - Tracks coverage (currently 100% - all 30 analysis tables)
   - Adds execution runtimes and dependencies

3. **Updates planning/README.md**:
   - Scans all subdirectories (ventros-ai, memory-service, mcp-server, grpc-api)
   - Lists all architecture documents
   - Shows implementation status and sprint timeline
   - Tracks documentation sizes

4. **Updates .claude/agents/README.md**:
   - Scans all agent files
   - Updates agent catalog with correct counts
   - Verifies output paths are correct
   - Maintains cross-references between agents

5. **Detects New Files**:
   - Finds files not yet indexed
   - Extracts metadata (title, description, size)
   - Adds to appropriate sections
   - Flags for manual review if needed

## When to Use

- After consolidating documentation
- After running `/full-analysis` (new analysis files created)
- After creating new agents
- After adding new planning documents
- Monthly maintenance (ensure all indexes are accurate)

## Output

Updates these files:
- `/home/caloi/ventros-crm/README.md`
- `/home/caloi/ventros-crm/code-analysis/README.md`
- `/home/caloi/ventros-crm/planning/README.md`
- `/home/caloi/ventros-crm/.claude/agents/README.md`

## Example Output

```markdown
## 📚 Project Documentation

**Total Files**: 15 markdown files in root

| Document | Purpose | Size | Lines |
|----------|---------|------|-------|
| README.md | Project overview and setup | 12K | 342 |
| DEV_GUIDE.md | Complete developer guide | 45K | 1,536 |
| TODO.md | Roadmap and priorities | 18K | 612 |
| AI_REPORT.md | Architectural audit (8.0/10) | 89K | 2,847 |
| CLAUDE.md | Claude Code instructions | 15K | 478 |
...

**Last Updated**: 2025-10-15 by mgmt_docs_index_manager
**Next Review**: 2025-11-15 (monthly)
```

## Agent Invoked

Triggers: `/home/caloi/ventros-crm/.claude/agents/mgmt_docs_index_manager.md`

**Runtime**: ~5-8 minutes

**Reads**:
- All *.md files in root, code-analysis/, planning/, .claude/agents/
- Git metadata (file sizes, dates)

**Writes**:
- README.md (root)
- code-analysis/README.md
- planning/README.md
- .claude/agents/README.md

## Related Commands

- `/update-todo` - Synchronize TODO.md after index update
- `/consolidate-docs` - Consolidate fragmented documentation first
