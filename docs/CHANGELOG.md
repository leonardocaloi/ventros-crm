# Changelog

Todas as mudanças importantes do Ventros CRM serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Versionamento Semântico](https://semver.org/lang/pt-BR/).

---

## [Unreleased]

### Em Desenvolvimento
- Consolidação de sessions (30 min gap)
- Correção de last_activity_at em importação histórica
- AI Memory Service (vetor search + hybrid search)

---

## [2.0.0] - 2025-10-16

### ✨ Added - Sistema AI de Desenvolvimento

#### Slash Commands
- **`/add-feature`** - Implementação inteligente de features com DDD + Clean Architecture
  - 30+ parâmetros para controle fino
  - Execução real de testes (`go test`)
  - Tracking em tempo real (P0 file)
  - 3 modos: Full (100k tokens), Enhancement (30k), Verification (10k)

- **`/pre-analyze`** - Análise pré-implementação com cache
  - Modo Quick: 6 analyzers, 5-10 min
  - Modo Deep: 14 analyzers, 15-30 min
  - Resultados salvos em `.claude/analysis/*.json`

- **`/test-feature`** - Execução real de testes com streaming
  - Integração com análise (mostra gaps conhecidos)
  - Priorização de testes (P0 > P1 > P2 > P3)
  - Coverage reports em HTML

- **`/review`** - Code review automatizado
  - Sistema de pontuação 100 pontos
  - Threshold: 80% (ou 90% com --strict)

#### AI Agents (32 total)
- **7 Meta Agents** (orquestração + desenvolvimento):
  - `meta_dev_orchestrator` - Orquestrador principal de features
  - `meta_context_builder` - Recomendação inteligente de padrões 🆕
  - `meta_feature_architect` - Validação de arquitetura
  - `meta_code_reviewer` - Code review automatizado
  - `meta_orchestrator` - Coordenação de análises
  - `meta_adr_generator` - Geração de ADRs
  - `meta_docs_consolidator` - Consolidação de docs

- **15 CRM Agents** (análise específica):
  - Domain: `crm_domain_model_analyzer`, `crm_value_objects_analyzer`, etc.
  - Infrastructure: `crm_persistence_analyzer`, `crm_integration_analyzer`, etc.
  - Quality: `crm_testing_analyzer`, `crm_security_analyzer`, etc.

- **4 Global Agents** (reutilizáveis):
  - `global_code_style_analyzer`
  - `global_solid_principles_analyzer`
  - `global_deterministic_analyzer`
  - `global_documentation_analyzer`

- **6 Management Agents** (manutenção):
  - `mgmt_todo_manager`, `mgmt_readme_updater`, etc.

#### State Management
- **`.claude/P0_ACTIVE_WORK.md`** - Tracking de trabalho ativo por branch
- **`.claude/AGENT_STATE.json`** - Estado compartilhado entre agentes
- **`.claude/analysis/*.json`** - Cache de análises (quick/deep modes)

#### Workflow Analysis-First
- Análise antes de implementar (optional mas recomendado)
- Recomendação inteligente de padrões:
  - Import → Temporal Workflow
  - API endpoint → Simple Handler
  - Background job → Temporal Workflow
  - Event processing → Choreography
- Detecção de features similares existentes

#### Visible Agent Coordination
- Cadeia hierárquica visível de chamadas entre agentes
- Logs detalhados: "🤖 [Level 2] meta_dev_orchestrator → meta_context_builder"
- Duração e tokens por agente

### 📚 Documentation
- **`docs/AI_AGENTS_COMPLETE_GUIDE.md`** (1,100+ linhas) - Guia completo do sistema AI
- **`docs/CHANGELOG.md`** (este arquivo) - Histórico de mudanças
- **`docs/archive/`** - Summaries históricos preservados
- **`DOCUMENTATION_AUDIT_AND_CLEANUP.md`** - Auditoria de toda documentação

### 🔧 Changed
- Consolidada documentação AI (3 arquivos → 1 guia completo)
- Reorganizados arquivos .md (12 na raiz → 4 essenciais)
- Estrutura de diretórios: raiz (essenciais) + docs/ (suplementar) + docs/archive/ (histórico)
- CLAUDE.md atualizado com referências ao sistema AI

### 🗑️ Removed
- Arquivos temporários de design (MAKEFILE_COMMANDS_REVIEW.md, MAKEFILE_DESIGN_FINAL.md)
- Documentação redundante (AI_DEVELOPMENT_SYSTEM.md, DEV_ORCHESTRATION_SUMMARY.md consolidados)
- Backup directory (.claude/agents.backup.20251015_221018/)

---

## [1.0.0] - 2025-10-15

### Added - Handler Refactoring (Sprint 0)
- **Command Pattern 100%**: Todos os handlers refatorados
- **Campaign Handlers**: 40% completo (5/13 métodos)
  - CreateCampaign: 80 linhas → 20 linhas (75% redução)
  - State commands: Activate, Pause, Complete delegam para handlers
- **Security**: RBAC + JWT + Project Member + Security Headers
- **Rate Limiting**: Middleware corrigido para API correta

### Added - Otimistic Locking
- **16/30 aggregates** com campo `version`
- Contact, Message, Session, Channel com locking implementado
- 14 aggregates ainda pendentes (ver TODO.md)

### Added - Stripe Billing Integration
- **Billing Aggregate** implementado
- Subscription management via Stripe
- Webhook handling para eventos Stripe

### Fixed
- **Rate Limiter**: Uso correto da API do middleware
- **RateLimiter struct**: Removido (obsoleto)
- **routes.go**: Atualizado para usar funções diretas de middleware

---

## [0.9.0] - 2025-10-14

### Added - WAHA History Import
- **Workflow Temporal**: Importação histórica de mensagens WhatsApp
- **90 days de histórico**: Processamento em batches
- **Consolidação de sessões**: Gap de 30 min
- **Enriquecimento de mensagens**: Audio (Whisper), Imagens (Gemini Vision), Docs (LlamaParse)

### Added - Message Enrichment (12 providers)
- **Audio**: Groq Whisper (FREE, 216x real-time) + OpenAI fallback
- **Images**: Vertex Vision (Gemini 1.5 Flash, $0.00025/imagem)
- **Documents**: LlamaParse ($1-3 per 1000 pages)
- **Profile Pictures**: Gemini Vision (0-10 score)
- Automatic provider routing + fallbacks

---

## [0.8.0] - 2025-10-13

### Added - Domain Events + Outbox Pattern
- **182+ domain events** mapeados
- **Outbox Pattern** com PostgreSQL NOTIFY (<100ms latency)
- **RabbitMQ Integration** para event bus
- Atomic persistence: aggregate + events em mesma transação

### Added - Multi-Tenancy (RLS)
- **Row-Level Security** em todas as tabelas
- **tenant_id obrigatório** em todos os aggregates
- PostgreSQL policies para isolamento por tenant
- Middleware RLS automático

---

## [0.7.0] - 2025-10-12

### Added - 30 Aggregates DDD
- **CRM**: Contact, Session, Message, Channel, Pipeline, Agent, Chat, Note, Tracking, Webhook (10 aggregates)
- **Automation**: Campaign, Sequence, Broadcast (3 aggregates)
- **Core**: Billing, Project, User, ProjectMember (4 aggregates)
- **Clean Architecture**: Domain → Application → Infrastructure
- **CQRS**: 80+ commands, 20+ queries

### Added - Testing Infrastructure
- **82% overall coverage** (61 unit tests)
- Test pyramid: 70% unit, 20% integration, 10% E2E
- GORM AutoMigrate + SQL migrations dual support

---

## [0.6.0] - 2025-10-11

### Added - API Foundation
- **158 endpoints** across 10 products
- **Swagger documentation** via swaggo
- **Gin HTTP framework** com middlewares (Auth, RLS, CORS, Rate Limiting)
- **Health endpoint** + Queue admin endpoints

### Added - Infrastructure
- **PostgreSQL 15+** com RLS
- **RabbitMQ 3.12+** para messaging
- **Redis 7.0+** para caching (configurado mas não usado ainda)
- **Temporal** para workflows
- **Docker Compose** para desenvolvimento local

---

## Tipos de Mudanças

- `Added` - Novas funcionalidades
- `Changed` - Mudanças em funcionalidades existentes
- `Deprecated` - Funcionalidades obsoletas (mas ainda funcionam)
- `Removed` - Funcionalidades removidas
- `Fixed` - Correções de bugs
- `Security` - Correções de segurança

---

## Estatísticas do Projeto

### Código
- **Linguagem**: Go 1.25.1
- **Arquitetura**: DDD + Clean Architecture + CQRS + Event-Driven
- **Score Arquitetural**: 8.0/10 (backend sólido, AI com gaps)

### Coverage
- **Overall**: 82%+
- **Domain**: 100% (business-critical)
- **Application**: 80%+
- **Infrastructure**: 60%+

### API
- **Total Endpoints**: 158
- **Swagger Coverage**: 135/158 (85%)
- **BOLA Checks**: 98/158 (62%) - 60 endpoints vulneráveis

### Security (⚠️ 5 P0 Vulnerabilities)
1. Dev Mode Bypass (CVSS 9.1)
2. SSRF in Webhooks (CVSS 9.1)
3. BOLA in 60 GET endpoints (CVSS 8.2)
4. Resource Exhaustion (CVSS 7.5)
5. RBAC Missing (CVSS 7.1)

**⚠️ NÃO deployar em produção até Sprint 1-2 de segurança completo**

---

## Roadmap

Ver `TODO.md` para detalhes completos.

### Sprint 1 (P0 - Security)
- [ ] Remover dev mode auth bypass
- [ ] Adicionar SSRF protection em webhooks
- [ ] Implementar BOLA checks em todos endpoints
- [ ] Adicionar max page size (resource exhaustion)
- [ ] Implementar RBAC completo

### Sprint 2 (P0 - Optimistic Locking)
- [ ] Adicionar version field nos 14 aggregates faltantes
- [ ] Testes de concorrência

### Sprint 3 (P1 - AI Memory)
- [ ] Vector search (pgvector)
- [ ] Hybrid search
- [ ] Memory facts extraction
- [ ] Python ADK (multi-agent)

---

**Maintainer**: Ventros CRM Team
**License**: Proprietary
**Status**: 🚧 Em desenvolvimento ativo
