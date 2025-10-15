# Planning - Future Features

**Status**: Architecture & Planning Phase
**Scope**: Features NOT yet implemented
**Timeline**: Sprint 5-30 (7.5 months)

---

## 🔮 Overview

This directory contains **planning documentation** for features that will be implemented in future sprints.

**Important**: These are NOT part of the current codebase. See `TODO.md` for implementation timeline.

---

## 📁 Directories

### 1. **ventros-ai/** - Python Microservice (ADK + Multi-Agent)
**Status**: Planning (0% implemented)
**Sprint**: 19-30 (6 weeks)
**Type**: Separate Python microservice

**What it is**:
- Python microservice (separate from Go backend)
- Google Cloud ADK 0.5+ (Agent Development Kit)
- Multi-agent system (5 specialist agents)
- Semantic router (DistilBERT intent classification)
- Temporal workflows for long-running tasks

**Components**:
- CoordinatorAgent (dispatches to specialists)
- SalesProspectingAgent
- RetentionChurnAgent
- SupportTechnicalAgent
- SupportBillingAgent
- BalancedAgent (fallback)

**Communication**: gRPC with Go backend

---

### 2. **memory-service/** - Vector Search & Hybrid Memory
**Status**: Planning (0% implemented)
**Sprint**: 5-11 (7 weeks)
**Type**: Go service extension

**What it is**:
- pgvector extension (PostgreSQL)
- Vector search (text-embedding-005)
- Hybrid search (vector + keyword + graph)
- Memory facts extraction (LLM-based)
- AI cost tracking

**Components**:
- VectorSearchService (pgvector + embeddings)
- HybridSearchService (4 strategies: baseline, vector, keyword, graph)
- FactExtractionService (Gemini Flash)
- CostTracker (billing integration)

---

### 3. **mcp-server/** - Claude Desktop Integration
**Status**: Planning (0% implemented)
**Sprint**: 15-18 (4 weeks)
**Type**: Go HTTP server (MCP protocol)

**What it is**:
- MCP (Model Context Protocol) server
- 30 tools for Claude Desktop
- Real-time CRM operations via Claude chat
- BI queries, contact management, memory search

**Tool Categories**:
- BI Tools (7): get_leads_count, get_conversion_rate, etc.
- CRM Operations (8): qualify_lead, update_pipeline_stage, etc.
- Memory Tools (5): search_memory, get_contact_context, etc.
- Document Tools (5): search_documents, summarize_document, etc.

---

### 4. **grpc-api/** - Go ↔ Python Communication
**Status**: Planning (0% implemented)
**Sprint**: 12-14 (3 weeks)
**Type**: gRPC service

**What it is**:
- gRPC API for Go ↔ Python communication
- Proto definitions (memory_service.proto, crm_service.proto)
- Bidirectional communication
- Authentication (JWT)

**Services**:
- MemoryService (search, store embeddings, extract facts)
- CRMService (partial, for Python ADK to call Go backend)

---

## 🗓️ Implementation Timeline

See `TODO.md` for detailed roadmap.

**Phase 2: AI Foundation** (Sprint 5-14, 10 weeks)
- Sprint 5-11: Memory Service (7 weeks)
- Sprint 12-14: gRPC API (3 weeks)

**Phase 3: AI Tools** (Sprint 15-30, 16 weeks)
- Sprint 15-18: MCP Server (4 weeks)
- Sprint 19-30: ventros-ai microservice (12 weeks)

---

## ⚠️ Important Notes

### NOT Part of Current Codebase
Files in `planning/` are **documentation only**. They describe future architecture but don't represent implemented code.

### See TODO_CURRENT.md for Code Gaps
For improvements to **existing code**, see `TODO_CURRENT.md` (ignores planning/).

### Consolidated from Multiple Sources
Documentation here was consolidated from:
- `docs/PYTHON_ADK_ARCHITECTURE.md` (+ PART2, PART3)
- `docs/AI_MEMORY_GO_ARCHITECTURE.md` (+ PART2, PART3)
- `docs/MCP_SERVER_COMPLETE.md`
- Various planning documents

---

## 📚 Related Documentation

- [TODO.md](../TODO.md) - Complete roadmap (includes planning/)
- [TODO_CURRENT.md](../TODO_CURRENT.md) - Current code only (ignores planning/)
- [Code Analysis](../code-analysis/) - Current state analysis

---

**Created**: 2025-10-15
**Purpose**: Future features planning & architecture
**Maintainer**: docs_index_manager agent
