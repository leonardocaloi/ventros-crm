# Agent Context Protocol

**Purpose**: Define o contexto obrigatório que TODOS os agentes devem ler antes de executar qualquer tarefa.

**Rule**: SEMPRE ler esses arquivos antes de começar qualquer análise ou implementação.

---

## 📚 Arquivos de Contexto Obrigatório

### 1. **CLAUDE.md** - Instruções Gerais
**Path**: `/home/caloi/ventros-crm/CLAUDE.md`
**Contém**:
- Visão geral do projeto
- Padrões arquiteturais (DDD, Clean Architecture, CQRS)
- Comandos essenciais (make, git, etc)
- Sistema de IA (slash commands, agentes)
- Guidelines críticos (ALWAYS Do / NEVER Do)

**Quando ler**: SEMPRE, em TODOS os agentes

---

### 2. **DEV_GUIDE.md** - Guia de Desenvolvimento
**Path**: `/home/caloi/ventros-crm/DEV_GUIDE.md`
**Contém**:
- Passo-a-passo para adicionar features
- Estrutura de diretórios completa
- Padrões de código (Domain, Application, Infrastructure)
- Exemplos práticos
- Checklist detalhado

**Quando ler**:
- Antes de implementar features (`meta_dev_orchestrator`)
- Antes de validar arquitetura (`meta_feature_architect`)
- Durante análise de código (`crm_domain_model_analyzer`, etc)

---

### 3. **AI_REPORT.md** - Auditoria Arquitetural
**Path**: `/home/caloi/ventros-crm/AI_REPORT.md`
**Contém**:
- Score de arquitetura (8.0/10)
- Pontos fortes e fracos
- Padrões já implementados
- Gaps identificados
- Recomendações

**Quando ler**:
- Antes de análise completa (`/analyze`)
- Antes de validação arquitetural (`meta_feature_architect`)
- Durante code review (`meta_code_reviewer`)

---

### 4. **TODO.md** - Roadmap e Prioridades
**Path**: `/home/caloi/ventros-crm/TODO.md`
**Contém**:
- P0 (vulnerabilidades críticas) - 5 items
- P1 (importante) - Sprint atual
- P2 (nice to have)
- Progresso de implementação

**Quando ler**:
- Antes de análise de segurança (`crm_security_analyzer`)
- Antes de planejamento (`meta_feature_architect`)
- Para evitar retrabalho (ver o que já está em TODO)

---

### 5. **P0_ACTIVE_WORK.md** - Trabalho Ativo
**Path**: `/home/caloi/ventros-crm/.claude/P0_ACTIVE_WORK.md`
**Contém**:
- Branches ativas
- Trabalho em progresso
- Blockers

**Quando ler**:
- Antes de iniciar novo trabalho (verificar se não há conflitos)
- Durante coordenação (`meta_dev_orchestrator`)

---

### 6. **AGENT_STATE.json** - Estado Compartilhado
**Path**: `/home/caloi/ventros-crm/.claude/AGENT_STATE.json`
**Contém**:
- Resultados de análises anteriores
- Resultados de testes
- Descobertas de outros agentes

**Quando ler**: SEMPRE, para ter contexto do que outros agentes já fizeram

---

## 🔄 Protocolo de Execução

### Para TODOS os Agentes:

```python
# FASE 0: SEMPRE fazer isso primeiro
print("📚 Loading architectural context...")

# 1. CLAUDE.md - Visão geral + padrões
claude_md = Read(file_path="/home/caloi/ventros-crm/CLAUDE.md")

# 2. AGENT_STATE.json - Contexto de outros agentes
agent_state = Read(file_path="/home/caloi/ventros-crm/.claude/AGENT_STATE.json")

# 3. P0_ACTIVE_WORK.md - Trabalho em progresso
p0_work = Read(file_path="/home/caloi/ventros-crm/.claude/P0_ACTIVE_WORK.md")

print("✅ Context loaded. Proceeding with task...")

# FASE 1: Agora pode executar a tarefa
# ... resto do agente
```

### Para Agentes de Implementação/Análise:

```python
# Além dos arquivos acima, também ler:

# 4. DEV_GUIDE.md - Guia detalhado
dev_guide = Read(file_path="/home/caloi/ventros-crm/DEV_GUIDE.md")

# 5. AI_REPORT.md - Auditoria arquitetural
ai_report = Read(file_path="/home/caloi/ventros-crm/AI_REPORT.md")

# 6. TODO.md - Prioridades e gaps conhecidos
todo = Read(file_path="/home/caloi/ventros-crm/TODO.md")

print("✅ Full architectural context loaded.")
```

### Para Agentes de Segurança:

```python
# Além dos arquivos base, também:

# TODO.md - P0 vulnerabilities
todo = Read(file_path="/home/caloi/ventros-crm/TODO.md")
# Extrair seção de P0 security issues
```

---

## 📋 Checklist por Tipo de Agente

### Meta Agents (Orchestration)
- [x] CLAUDE.md
- [x] AGENT_STATE.json
- [x] P0_ACTIVE_WORK.md
- [x] DEV_GUIDE.md
- [x] AI_REPORT.md
- [x] TODO.md

### CRM Analyzers
- [x] CLAUDE.md
- [x] AGENT_STATE.json
- [x] DEV_GUIDE.md (especialmente seção de domain patterns)
- [x] AI_REPORT.md (para contexto de qualidade)
- [ ] TODO.md (opcional)

### Security Analyzers
- [x] CLAUDE.md (seção de security)
- [x] AGENT_STATE.json
- [x] TODO.md (seção P0 vulnerabilities)
- [x] AI_REPORT.md (security score)

### Testing Analyzers
- [x] CLAUDE.md (seção de testing)
- [x] AGENT_STATE.json (test results anteriores)
- [x] DEV_GUIDE.md (estratégia de testes)

### Code Reviewer
- [x] CLAUDE.md (padrões)
- [x] DEV_GUIDE.md (checklist completo)
- [x] AI_REPORT.md (baseline de qualidade)
- [x] AGENT_STATE.json

---

## 🎯 Benefícios

### 1. Contexto Completo
- Agentes sabem os padrões do projeto
- Sabem o que já foi feito
- Sabem os gaps conhecidos

### 2. Consistência
- Todos seguem os mesmos padrões
- Não há conflitos de abordagem

### 3. Evita Retrabalho
- Vê o que está em TODO
- Vê o que outros agentes já analisaram

### 4. Melhor Qualidade
- Decisões informadas por contexto completo
- Alinhamento com arquitetura existente

---

## 🔧 Implementação nos Agentes

Vou atualizar os principais agentes para seguir este protocolo:

1. `meta_dev_orchestrator.md` - Adicionar leitura de contexto
2. `meta_feature_architect.md` - Adicionar leitura de contexto
3. `meta_code_reviewer.md` - Adicionar leitura de contexto
4. Todos os analyzers (crm_*, global_*) - Adicionar leitura de contexto

---

**Version**: 1.0
**Created**: 2025-10-16
**Purpose**: Garantir que todos os agentes têm contexto arquitetural completo antes de executar
**Compliance**: MANDATORY para todos os 32 agentes
