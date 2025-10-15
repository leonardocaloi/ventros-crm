# 📋 VENTROS CRM - ROADMAP ATUALIZADO

**Data**: 2025-10-13
**Versão**: 2.0 (com WAHA History Sync P0)

---

## 🔴 TAREFAS P0 CRÍTICAS (ANTES DE PRODUÇÃO)

### SPRINT 0: 🔴 WAHA History Sync (2 semanas) - **NOVO P0**

**Priority**: 🔴 **P0 CRÍTICO** - Funcionalidade essencial

**Objetivo**: Sincronização completa de histórico WAHA (incoming + outgoing messages, contacts, agent assignment).

**Tasks**:

#### Week 1: Foundation & WAHA Integration
- [ ] Migration 000053 (history import fields)
- [ ] Channel aggregate updates (enable/start/complete/fail import)
- [ ] WAHA History Client (fetch messages API)
- [ ] Pagination + retry logic
- [ ] Tests unitários

#### Week 2: Application Service & E2E
- [ ] WahaHistoryImportService (orchestration)
- [ ] Contact auto-creation (deduplication by phone)
- [ ] Agent assignment (channel.default_agent_id)
- [ ] Message status mapping (ack → status)
- [ ] Temporal workflow (long-running)
- [ ] HTTP handlers (start/status endpoints)
- [ ] **E2E Tests**:
  - [ ] Import 100+ messages (incoming + outgoing)
  - [ ] Deduplication (reimport idempotent)
  - [ ] Contact creation + update
  - [ ] Agent assignment verification
  - [ ] Media enrichment trigger
  - [ ] Incremental sync (last_import_date)

**Deliverables**:
- ✅ WAHA history sync production-ready
- ✅ E2E tests passing (100% coverage)
- ✅ Incremental sync funcionando
- ✅ Import speed: >100 messages/min
- ✅ Deduplication: 100% accuracy

**Spec Detalhada**: Ver `P0_WAHA_HISTORY_SYNC.md`

---

### SPRINT 1-2: Security Fixes (3-4 semanas)

**Objetivo**: Corrigir 5 vulnerabilidades críticas.

#### Sprint 1 (2 semanas)
1. **Dev Mode Bypass** (1 dia) - CVSS 9.1
2. **SSRF em Webhooks** (3 dias) - CVSS 9.1
3. **BOLA em GET endpoints** (1 semana) - CVSS 8.2

#### Sprint 2 (2 semanas)
4. **Resource Exhaustion** (3 dias) - CVSS 7.5
5. **RBAC Missing** (1 semana) - CVSS 7.1
6. **Rate Limiting Redis** (3 dias)

**Deliverables**: API segura para production ✅

---

### SPRINT 3-4: Cache Layer + N+1 Fixes (2 semanas)

**Objetivo**: Resolver performance críticos.

1. **Redis Cache Integration** (5 dias)
2. **N+1 Query Fix** (2 dias)
3. **Materialized View** (5 dias)

**Deliverables**: Queries <200ms, cache hit >70% ✅

---

## 📊 TIMELINE ATUALIZADO

| Sprint | Fase | Duration | Priority | Deliverable |
|--------|------|----------|----------|-------------|
| **0** | **WAHA History Sync** | **2 sem** | 🔴 **P0 NOVO** | **Sync completo** |
| 1-2 | Security Fixes | 3-4 sem | 🔴 P0 | API segura |
| 3-4 | Cache + Performance | 2 sem | 🔴 P0 | Queries <200ms |
| 5-8 | Memory Service | 4 sem | 🔴 P0 | Hybrid search |
| 9-11 | Facts + Cost | 3 sem | 🔴 P0 | Facts + billing |
| 12-14 | gRPC API | 3 sem | 🔴 P0 | Go ↔ Python |
| 15-18 | MCP Server | 4 sem | 🔴 P0 | 25 tools |
| 19-24 | Python ADK | 6 sem | 🔴 P0 | Multi-agent |
| 25-26 | Testing | 2 sem | 🟡 P1 | Coverage >85% |
| 27-30 | Advanced | 4 sem | 🟡 P1 | Enterprise |

**Total Duration**: **32 sprints** (32 semanas = 8 meses)

**P0 Features**: Sprints 0-24 (26 semanas = 6.5 meses)

---

## 🎯 MILESTONES ATUALIZADOS

| Milestone | Sprint | Date (est.) | Key Deliverable |
|-----------|--------|-------------|-----------------|
| **M0: WAHA History Ready** | 0 | Week 2 | ✅ **NOVO** - History sync functional |
| **M1: Security Hardened** | 2 | Week 6 | Production-safe API |
| **M2: Performance Optimized** | 4 | Week 8 | Cache layer live |
| **M3: Memory Service Live** | 11 | Week 16 | Hybrid search + facts |
| **M4: gRPC API Ready** | 14 | Week 19 | Python ↔ Go communication |
| **M5: MCP Server Beta** | 18 | Week 24 | Claude Desktop integration |
| **M6: Multi-Agent GA** | 24 | Week 32 | AI agents production |
| **M7: Enterprise Ready** | 30 | Week 40 | Full feature set |

---

## 💡 DECISÃO: WAHA HISTORY SYNC PRIMEIRO

**Justificativa**:
1. ✅ **Funcionalidade core** - Importar histórico é essencial para CRM
2. ✅ **Bloqueador de adoção** - Clientes não adotam sem histórico
3. ✅ **E2E crítico** - Testa pipeline completo (WAHA → DB → enrichment)
4. ✅ **Foundation** - Base para agent assignment e sessions
5. ✅ **Rápido** - 2 semanas, não bloqueia security fixes (pode ser paralelo)

**Sugestão**:
- **Semana 1-2**: WAHA History Sync (1 dev)
- **Semana 1-4**: Security Fixes (1 dev) **EM PARALELO**
- **Total**: 4 semanas com 2 devs (ao invés de 6 semanas sequencial)

---

## 📝 PRÓXIMOS PASSOS

1. **Agora**: Review spec completa em `P0_WAHA_HISTORY_SYNC.md`
2. **Segunda**: Kickoff Sprint 0 (WAHA History Sync)
3. **Paralelo**: Iniciar security fixes
4. **Semana 3**: Deploy WAHA sync em staging
5. **Semana 4**: Deploy security fixes em staging
6. **Semana 5**: Deploy conjunto em production

---

## 🚀 ORDEM DE EXECUÇÃO RECOMENDADA

### Opção A: Sequencial (conservador)
```
Sprint 0 (2 sem) → Sprint 1-2 (4 sem) → Sprint 3-4 (2 sem) → ...
Total: 32 semanas
```

### Opção B: Paralelo (recomendado) ✅
```
Semana 1-2:  WAHA History Sync (Dev 1) + Security Fixes (Dev 2)
Semana 3-4:  Security Fixes continuação (Dev 1 + Dev 2)
Semana 5-6:  Cache Layer (Dev 1 + Dev 2)
Total: 30 semanas (2 semanas economizadas)
```

**Recomendação**: **Opção B** (paralelo) - economiza 2 semanas, testa integração cedo.

---

**Status**: 🟡 Aguardando aprovação para iniciar Sprint 0

**Próxima Revisão**: Após Sprint 0 (WAHA History Sync completo)

---

**Ver também**:
- `P0_WAHA_HISTORY_SYNC.md` - Spec completa da implementação
- `AI_REPORT_PART4.md` - Análise de segurança (OWASP)
- `AI_REPORT_PART6.md` - Roadmap completo original
