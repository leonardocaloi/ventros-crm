# 🎯 Guia de Engenharia de Prompt para Claude Code

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [Técnicas Fundamentais](#técnicas-fundamentais)
3. [Técnicas Avançadas](#técnicas-avançadas)
4. [Técnicas Específicas do Claude](#técnicas-específicas-do-claude)
5. [Aplicação em Agentes Paralelos](#aplicação-em-agentes-paralelos)
6. [Templates Reutilizáveis](#templates-reutilizáveis)
7. [Best Practices](#best-practices)

---

## 📚 Visão Geral

Este guia apresenta **15+ técnicas de prompt engineering** aplicadas ao Claude Code, com foco especial em:
- **Sub-agentes especializados** com prompts otimizados
- **Execução paralela** de múltiplas tarefas
- **Estruturação de comandos** personalizados
- **Orquestração** de sistemas multi-agente

### Matriz de Técnicas

| Técnica | Complexidade | Uso em Agentes | Paralelização | Melhor Para |
|---------|--------------|----------------|---------------|-------------|
| Zero-Shot | Baixa | ✅ | ✅ | Tarefas simples |
| Few-Shot | Média | ✅ | ✅ | Padrões repetitivos |
| Chain of Thought (CoT) | Média | ✅ | ❌ | Raciocínio passo-a-passo |
| Skeleton Prompting | Alta | ✅ | ✅ | Conteúdo longo estruturado |
| Tree of Thoughts (ToT) | Alta | ✅ | ✅ | Exploração de alternativas |
| ReAct | Média | ✅ | ❌ | Ações iterativas |
| Self-Consistency | Alta | ✅ | ✅ | Validação de respostas |
| Prompt Chaining | Média | ✅ | ✅ | Workflows complexos |
| Meta Prompting | Alta | ✅ | ❌ | Auto-refinamento |
| XML Tags | Baixa | ✅ | ✅ | Estruturação de dados |
| Extended Thinking | Média | ✅ | ❌ | Problemas complexos |
| Recursive Prompting | Alta | ✅ | ✅ | Dados volumosos |

---

## 🎓 Técnicas Fundamentais

### 1. Zero-Shot Prompting

**Descrição**: Fazer uma pergunta ou apresentar uma tarefa sem fornecer exemplos.

**Quando usar**: Tarefas simples que o modelo já compreende bem.

**Exemplo Básico**:
```markdown
Analyze the Contact aggregate and check if it has a version field for optimistic locking.
```

**Exemplo em Agente Claude Code**:
```markdown
---
name: quick-analyzer
description: Quick analysis without examples
tools: Read, Grep
model: sonnet
---

You are a code analyzer. Given a file path, analyze its structure and report findings.

Your analysis should include:
- File purpose
- Key components
- Potential issues

Be concise and direct.
```

**Prós**: Simples, rápido, econômico
**Contras**: Menos preciso para tarefas complexas ou específicas

---

### 2. Few-Shot Prompting

**Descrição**: Fornecer 2-5 exemplos do formato/padrão desejado antes da tarefa real.

**Quando usar**: Quando você precisa que o modelo siga um padrão específico.

**Exemplo em Agente Claude Code**:
```markdown
---
name: event-validator
description: Validate event naming conventions using examples
tools: Grep, Read
model: sonnet
---

# Event Naming Validator Agent

You validate event naming conventions in the codebase.

## Correct Event Naming Examples

**Example 1**: ✅ CORRECT
```go
type ContactCreatedEvent struct {
    ContactID uuid.UUID
}
func (e ContactCreatedEvent) EventType() string {
    return "contact.created"  // Format: aggregate.action (lowercase, past tense)
}
```

**Example 2**: ✅ CORRECT
```go
type SessionEndedEvent struct {
    SessionID uuid.UUID
}
func (e SessionEndedEvent) EventType() string {
    return "session.ended"  // Correct format
}
```

**Example 3**: ❌ INCORRECT
```go
type CampaignActivatedEvent struct {
    CampaignID uuid.UUID
}
func (e CampaignActivatedEvent) EventType() string {
    return "CampaignActivated"  // Wrong: CamelCase instead of lowercase
}
```

**Example 4**: ❌ INCORRECT
```go
type MessageSendEvent struct {
    MessageID uuid.UUID
}
func (e MessageSendEvent) EventType() string {
    return "message.send"  // Wrong: should be past tense "message.sent"
}
```

## Your Task

Now analyze the event naming in the provided codebase and report violations following the pattern shown above.

## Output Format

Use the same format as the examples:
- ✅ for correct events
- ❌ for violations
- Explain why each violation is wrong
```

**Prós**: Muito mais preciso, estabelece padrões claros
**Contras**: Usa mais tokens, requer bons exemplos

---

### 3. Chain of Thought (CoT) Prompting

**Descrição**: Instruir o modelo a "pensar em voz alta", mostrando os passos intermediários do raciocínio.

**Quando usar**: Problemas que requerem raciocínio lógico passo-a-passo.

**Exemplo em Agente Claude Code**:
```markdown
---
name: architecture-auditor
description: Audit architecture using step-by-step reasoning
tools: Read, Grep, Glob
model: sonnet
---

# Architecture Auditor Agent

You audit codebases for architectural violations using systematic reasoning.

## Your Reasoning Process

For each analysis, **think step-by-step**:

### Step 1: Understand the Architecture
First, read AI_REPORT.md and CLAUDE.md to understand:
- What architectural pattern is used? (Hexagonal, Clean, etc.)
- What are the layer boundaries?
- What are the rules?

**Reasoning**: I need context before I can identify violations.

### Step 2: Identify Layer Dependencies
Check imports in each layer:
- Domain layer should NOT import infrastructure
- Application layer can import domain
- Infrastructure layer can import both

**Reasoning**: Dependency direction violations break Clean Architecture.

### Step 3: Find Violations
Use Grep to search for violations:
```bash
# Check if domain imports infrastructure (violation)
grep -r "import.*infrastructure" internal/domain/
```

**Reasoning**: Any match here is a critical violation.

### Step 4: Analyze Each Violation
For each violation found:
- Why is this a violation?
- What's the impact?
- How to fix it?

**Reasoning**: Understanding WHY helps prevent future violations.

### Step 5: Report Findings
Consolidate all findings with:
- Severity (Critical, High, Medium, Low)
- Location (file:line)
- Explanation
- Suggested fix

## Your Task

Follow this step-by-step reasoning process to audit the codebase. **Show your thinking for each step**.
```

**Exemplo de Output Esperado**:
```markdown
## Step 1: Understanding Architecture
I read AI_REPORT.md and found:
- Pattern: Hexagonal Architecture + DDD
- Layers: Domain → Application → Infrastructure
- Rule: Domain must be pure (no external dependencies)

## Step 2: Checking Dependencies
I'm searching for violations in the domain layer...
[grep command output]

## Step 3: Violations Found
Found 2 violations:
1. internal/domain/crm/contact/contact.go:15
2. internal/domain/automation/campaign/campaign.go:23

## Step 4: Analysis
Violation 1: contact.go imports infrastructure/messaging
- **Why violation**: Domain depends on infrastructure
- **Impact**: Breaks layer isolation, makes domain not portable
- **Fix**: Move messaging to application layer, use events instead

[continue for each violation...]
```

**Prós**: Melhora raciocínio complexo, transparência
**Contras**: Usa muitos tokens, não paraleliza bem

---

## 🚀 Técnicas Avançadas

### 4. Skeleton Prompting (Skeleton-of-Thought)

**Descrição**: Processo em 2 fases:
1. **Fase 1**: Gerar um "esqueleto" (estrutura de alto nível)
2. **Fase 2**: Expandir cada ponto do esqueleto **em paralelo**

**Quando usar**: Conteúdo longo e estruturado (relatórios, documentação, análises abrangentes).

**Performance**: 2x mais rápido em 60% dos casos (paralelização)

**Exemplo em Agente Claude Code**:

#### Fase 1: Skeleton Generator
```markdown
---
name: skeleton-generator
description: Phase 1 - Generate high-level skeleton for comprehensive analysis
tools: Read
model: sonnet
---

# Skeleton Generator Agent

You create **concise skeletons** (outlines) for comprehensive analysis tasks.

## Instructions

Given a codebase analysis request, generate ONLY the skeleton (not full content).

**Format**: Numbered list with 3-5 words per point.

**Example Request**: "Analyze the entire Ventros CRM codebase"

**Example Skeleton**:
1. Domain layer DDD patterns
2. Application layer CQRS implementation
3. Infrastructure layer integrations
4. Security vulnerabilities audit
5. Test coverage analysis
6. Performance bottlenecks
7. Documentation completeness
8. Event-driven architecture
9. Database schema design
10. API endpoint consistency

**Your Task**: Generate skeleton for the given analysis request. **Maximum 10 points, 3-5 words each**.
```

#### Fase 2: Point Expander (PARALLEL)
```markdown
---
name: point-expander
description: Phase 2 - Expand one skeleton point in detail (runs in parallel)
tools: Read, Grep, Glob, Bash
model: sonnet
---

# Point Expander Agent

You expand **one and only one** point from a skeleton into detailed analysis.

## Instructions

You will receive:
- **Skeleton point index**: Which point to expand (e.g., "Point 3")
- **Point description**: What the point is about (e.g., "Infrastructure layer integrations")
- **Context**: Full skeleton for reference

**Your Job**: Expand ONLY your assigned point into 1-2 paragraphs with:
- Detailed analysis
- Evidence (code snippets, grep results)
- Findings (issues, recommendations)

**Keep it concise**: 1-2 paragraphs maximum.

## Example

**Input**:
- Point Index: 3
- Description: "Infrastructure layer integrations"
- Skeleton: [1. Domain..., 2. Application..., 3. Infrastructure..., ...]

**Output**:
```markdown
## 3. Infrastructure Layer Integrations

The infrastructure layer implements 5 external integrations: WAHA (WhatsApp), RabbitMQ (messaging), PostgreSQL (persistence), Redis (caching), and Temporal (workflows). The WAHA integration is well-structured with proper error handling and retry logic (infrastructure/channels/waha/client.go). However, Redis integration exists but is **never used** (0% adoption), representing technical debt. Recommendation: Either implement caching layer or remove Redis dependency to reduce operational complexity.
```
```

#### Orquestrador de Skeleton (Combina Fase 1 + 2)
```markdown
---
name: skeleton-orchestrator
description: Orchestrate skeleton-based analysis (2-phase parallel processing)
tools: Task, Read, Write
model: opus
---

# Skeleton Orchestrator

You coordinate skeleton-based analysis using parallel processing.

## Workflow

### Phase 1: Generate Skeleton
1. Launch **skeleton-generator** agent to create outline
2. Wait for skeleton (10 points)

### Phase 2: Parallel Expansion
3. Launch **10 point-expander agents IN PARALLEL**, each assigned one point
4. Wait for all expansions to complete

### Phase 3: Consolidation
5. Merge all expanded points into final report
6. Add introduction and conclusion
7. Save to ai-guides/analysis-reports/

## Execution

When user requests comprehensive analysis:
1. Use Task tool to launch skeleton-generator
2. Parse skeleton into individual points
3. Use Task tool to launch 10 point-expander agents in parallel (one per point)
4. Consolidate results
5. Present final report

**Performance**: 2x faster than sequential analysis due to parallelization.
```

**Uso Prático**:
```bash
User: "Analyze the entire Ventros CRM codebase comprehensively"

Orchestrator:
1. Launches skeleton-generator → Gets 10-point outline
2. Launches 10 point-expander agents in parallel (all at once)
3. Waits ~2 min (vs ~10 min sequential)
4. Merges results into final report
```

**Prós**: 2x mais rápido, estrutura clara, paraleliza bem
**Contras**: Mais complexo de implementar, custos de token maiores

---

### 5. Tree of Thoughts (ToT)

**Descrição**: Explorar múltiplos caminhos de raciocínio simultaneamente, com avaliação e poda.

**Quando usar**: Problemas que requerem exploração de alternativas (design decisions, debugging).

**Exemplo em Agente Claude Code**:
```markdown
---
name: refactoring-explorer
description: Explore multiple refactoring strategies using Tree of Thoughts
tools: Read, Grep
model: opus
---

# Refactoring Explorer Agent (Tree of Thoughts)

You explore multiple refactoring strategies and evaluate trade-offs.

## Process

Given a code refactoring task, you:

### Step 1: Generate Alternative Approaches (Breadth)
Brainstorm 3-5 different refactoring strategies.

**Example Task**: "Refactor the Contact aggregate to improve testability"

**Thought 1**: Extract business logic into separate methods
**Thought 2**: Introduce repository interface with mock
**Thought 3**: Use dependency injection for event bus
**Thought 4**: Apply Strategy pattern for validation rules
**Thought 5**: Implement Builder pattern for Contact creation

### Step 2: Evaluate Each Thought
Rate each approach on:
- **Testability improvement**: 1-10
- **Complexity added**: 1-10 (lower is better)
- **Breaking changes**: 1-10 (lower is better)
- **Alignment with DDD**: 1-10

**Example Evaluation**:
```
Thought 1: Extract methods
- Testability: 7/10
- Complexity: 2/10 ✅
- Breaking changes: 1/10 ✅
- DDD alignment: 8/10
- **Score**: 18/40 (higher is better)

Thought 2: Repository interface
- Testability: 9/10 ✅
- Complexity: 4/10
- Breaking changes: 3/10
- DDD alignment: 10/10 ✅
- **Score**: 22/40

[continue for all thoughts...]
```

### Step 3: Prune Weak Approaches
Eliminate thoughts with score < 15/40.

### Step 4: Expand Best Thoughts (Depth)
For top 2 thoughts, explore implementation details:
- What files to change?
- What patterns to use?
- What tests to add?

### Step 5: Recommend Best Path
Choose the highest-scoring approach with detailed implementation plan.

## Output Format

```markdown
## Refactoring Exploration (Tree of Thoughts)

### Alternative Approaches
[List 3-5 thoughts]

### Evaluation Matrix
| Thought | Testability | Complexity | Breaking | DDD | Total |
|---------|-------------|------------|----------|-----|-------|
| 1. Extract methods | 7 | 2 | 1 | 8 | 18 |
| 2. Repository interface | 9 | 4 | 3 | 10 | 22 ✅ |
[...]

### Pruned Approaches
- Thought 4 (score: 12/40) - Too complex for little gain

### Detailed Plan (Best Approach: Thought 2)
[Implementation details]

### Recommendation
Use **Thought 2: Repository interface** because it maximizes testability while maintaining DDD principles.
```
```

**Uso com Múltiplos Agentes**:
```markdown
# Orchestrator launches 5 agents in parallel, each exploring one "thought"
# Then consolidates and ranks results
```

**Prós**: Explora alternativas, avalia trade-offs
**Contras**: Muito custoso em tokens, requer modelo inteligente (Opus)

---

### 6. ReAct (Reasoning + Acting)

**Descrição**: Ciclo de **Thought → Action → Observation → Thought → ...** até resolver.

**Quando usar**: Debugging, investigação iterativa, tarefas que requerem feedback do ambiente.

**Exemplo em Agente Claude Code**:
```markdown
---
name: bug-investigator
description: Investigate bugs using ReAct pattern (iterative reasoning + action)
tools: Read, Grep, Bash
model: sonnet
---

# Bug Investigator Agent (ReAct Pattern)

You investigate bugs using iterative reasoning and action.

## ReAct Loop

Repeat this cycle until bug is identified:

1. **Thought**: What do I know? What should I check next?
2. **Action**: Run a command or read a file
3. **Observation**: What did I learn?
4. **Thought**: Does this explain the bug? Or do I need more info?

## Example

**User Report**: "Contact creation fails with 500 error"

---

**Thought 1**: I need to find where contact creation happens.
**Action 1**: `grep -r "CreateContact" internal/`
**Observation 1**: Found `internal/application/contact/create_contact.go` and handler in `infrastructure/http/handlers/contact_handler.go`

**Thought 2**: Let me check the handler for error handling.
**Action 2**: Read `infrastructure/http/handlers/contact_handler.go`
**Observation 2**: Handler calls `CreateContactHandler.Handle()` but doesn't check for validation errors before returning.

**Thought 3**: This could be the issue. Let me check the command handler.
**Action 3**: Read `internal/application/contact/create_contact.go`
**Observation 3**: Command handler returns `ErrInvalidInput` but handler doesn't map it to 400 (returns 500 instead).

**Thought 4**: Found the bug! Handler needs to map `ErrInvalidInput` to HTTP 400.
**Action 4**: No action needed, bug identified.

---

## Output Format

Present your investigation as:
```markdown
## Bug Investigation (ReAct)

**Thought 1**: [reasoning]
**Action 1**: [command/file read]
**Observation 1**: [what you learned]

**Thought 2**: [reasoning]
**Action 2**: [command/file read]
**Observation 2**: [what you learned]

[continue until bug found...]

## Root Cause
[Explanation]

## Fix
[Suggested code change]
```
```

**Prós**: Excelente para debugging, auto-corrige com feedback
**Contras**: Iterativo (não paraleliza), muitos tokens

---

### 7. Self-Consistency

**Descrição**: Gerar **múltiplas respostas independentes** para a mesma pergunta, depois escolher a mais frequente/consistente.

**Quando usar**: Validação de análises críticas, decisões importantes.

**Exemplo em Agente Claude Code**:
```markdown
---
name: security-validator
description: Validate security findings using self-consistency (multiple runs)
tools: Read, Grep
model: sonnet
---

# Security Validator Agent (Self-Consistency)

You validate security findings by running multiple independent analyses.

## Process

Given a security concern, you:

1. **Run 1**: Analyze independently (fresh context)
2. **Run 2**: Analyze independently (don't look at Run 1)
3. **Run 3**: Analyze independently (don't look at Run 1 or 2)

Then compare results and report only **consistent findings** (appear in 2+ runs).

## Example

**Task**: "Check if middleware/auth.go has security issues"

---

**Run 1 Output**:
- Issue A: Dev mode bypass on line 41 (CRITICAL)
- Issue B: Missing rate limiting (MEDIUM)
- Issue C: Weak JWT validation (HIGH)

**Run 2 Output**:
- Issue A: Dev mode bypass on line 41 (CRITICAL)
- Issue C: Weak JWT validation (HIGH)
- Issue D: No CSRF protection (MEDIUM)

**Run 3 Output**:
- Issue A: Dev mode bypass on line 41 (CRITICAL)
- Issue C: JWT validation issues (HIGH)

---

**Consistent Findings** (appear in 2+ runs):
- ✅ Issue A: Dev mode bypass (3/3 runs) → **HIGH CONFIDENCE**
- ✅ Issue C: JWT validation (3/3 runs) → **HIGH CONFIDENCE**

**Inconsistent Findings** (appear in only 1 run):
- ⚠️ Issue B: Rate limiting (1/3 runs) → **LOW CONFIDENCE**
- ⚠️ Issue D: CSRF (1/3 runs) → **LOW CONFIDENCE**

## Output Format

```markdown
## Security Analysis (Self-Consistency)

### Run 1 Findings
[list issues]

### Run 2 Findings
[list issues]

### Run 3 Findings
[list issues]

### Consistent Findings (High Confidence)
[issues that appear in 2+ runs]

### Inconsistent Findings (Low Confidence)
[issues that appear in only 1 run - may be false positives]

### Recommendation
Focus on consistent findings first.
```
```

**Implementação Paralela**:
```markdown
# Orchestrator launches 3 instances of security-validator in parallel
# Each runs independently with isolated context
# Then consolidates results
```

**Prós**: Reduz falsos positivos, aumenta confiança
**Contras**: 3x custo de tokens, requer isolamento de contexto

---

### 8. Prompt Chaining

**Descrição**: Conectar múltiplos prompts onde a saída de um é entrada do próximo.

**Tipos**:
- **Sequential**: A → B → C (linear)
- **Branching**: A → [B1, B2, B3] (paralelo)
- **Recursive**: A → [A1, A2] → [A1.1, A1.2, A2.1, A2.2] (fractal)
- **Conditional**: A → (if X then B else C) (dinâmico)

**Exemplo em Agente Claude Code (Sequential)**:
```markdown
---
name: chain-orchestrator
description: Orchestrate sequential prompt chain for comprehensive analysis
tools: Task
model: opus
---

# Chain Orchestrator

Execute analysis in sequential stages:

## Stage 1: Data Collection
Launch **data-collector** agent to gather:
- File structure
- Dependencies
- Test coverage stats

**Output**: JSON with collected data

## Stage 2: Pattern Analysis
Launch **pattern-analyzer** agent with Stage 1 output to identify:
- Design patterns used
- Anti-patterns found
- Architectural violations

**Output**: List of findings

## Stage 3: Recommendation Generation
Launch **recommender** agent with Stage 2 output to generate:
- Prioritized fixes
- Refactoring suggestions
- Architecture improvements

**Output**: Actionable roadmap

## Stage 4: Report Writing
Launch **report-writer** agent with all previous outputs to create:
- Executive summary
- Detailed findings
- Implementation plan

**Output**: Final markdown report

## Execution

Use Task tool sequentially (wait for each stage to complete before starting next).
```

**Exemplo Branching (Parallel)**:
```markdown
## Stage 1: Parallel Investigation
Launch 5 agents in parallel:
- domain-analyzer
- security-auditor
- test-specialist
- performance-analyzer
- docs-checker

## Stage 2: Consolidation
Launch consolidator with all 5 outputs to merge into unified report.
```

**Prós**: Modularidade, reutilização, clareza
**Contras**: Overhead de coordenação

---

### 9. Recursive Prompting (Recursion of Thought)

**Descrição**: Dividir grandes inputs em chunks pequenos, processar recursivamente, depois agregar.

**Quando usar**: Analisar grandes arquivos, múltiplos módulos, datasets volumosos.

**Exemplo em Agente Claude Code**:
```markdown
---
name: recursive-analyzer
description: Recursively analyze large directory structures
tools: Glob, Read, Task
model: sonnet
---

# Recursive Analyzer Agent

You analyze large codebases by recursively dividing them into smaller chunks.

## Recursive Strategy

### Level 1: Directory Level
For each top-level directory (e.g., `internal/domain/`, `internal/application/`):
- Launch sub-agent to analyze that directory

### Level 2: Module Level
Each directory agent further divides into modules:
- `internal/domain/crm/` → Launch agent for `contact/`, `session/`, `message/`, etc.

### Level 3: File Level
Each module agent analyzes individual files.

### Aggregation
Results bubble up:
- File results → Module summary
- Module summaries → Directory summary
- Directory summaries → Final report

## Example

**Input**: Analyze `internal/domain/` (100+ files)

**Recursion**:
```
recursive-analyzer (internal/domain/)
    ├─→ crm-analyzer (internal/domain/crm/)
    │   ├─→ contact-analyzer (internal/domain/crm/contact/)
    │   │   ├─→ Analyze contact.go → "Has version field ✅"
    │   │   ├─→ Analyze events.go → "3 events defined ✅"
    │   │   └─→ Analyze repository.go → "Interface defined ✅"
    │   │   └─→ **Summary**: "Contact aggregate: 100% compliant"
    │   │
    │   ├─→ session-analyzer (internal/domain/crm/session/)
    │   │   └─→ **Summary**: "Session aggregate: Missing version field ❌"
    │   │
    │   └─→ **CRM Summary**: "23 aggregates analyzed, 3 issues found"
    │
    ├─→ automation-analyzer (internal/domain/automation/)
    │   └─→ **Automation Summary**: "3 aggregates, all compliant ✅"
    │
    └─→ **FINAL REPORT**: "26 total aggregates, 3 issues, 88% compliance"
```

## Implementation

Use Task tool to launch sub-agents recursively. Each level waits for its children before aggregating.
```

**Prós**: Escala para grandes volumes, paraleliza bem
**Contras**: Complexo de implementar, requer agregação cuidadosa

---

### 10. Meta Prompting

**Descrição**: O próprio modelo gera e refina seus prompts.

**Processo**:
1. Gerar prompt inicial
2. Executar e obter resultado
3. Gerar feedback sobre o prompt
4. Refinar prompt
5. Repetir até atingir qualidade desejada

**Exemplo em Agente Claude Code**:
```markdown
---
name: meta-prompter
description: Self-improve prompts using meta-prompting
tools: Task, Write
model: opus
---

# Meta Prompter Agent

You iteratively improve prompts through self-refinement.

## Process

### Iteration 1: Initial Prompt
Generate initial prompt for the task.

**Example Task**: "Create agent that validates DDD patterns"

**Initial Prompt**:
```markdown
You are a DDD validator. Check if aggregates follow DDD patterns.
```

### Iteration 2: Execute & Critique
- Execute initial prompt on test data
- Analyze results
- Identify weaknesses

**Critique**:
```
Weaknesses:
1. Too vague - doesn't specify which DDD patterns
2. No output format defined
3. No examples provided
4. Doesn't mention specific checks (events, version field, etc.)
```

### Iteration 3: Refine Prompt
```markdown
You are a Domain-Driven Design validator for Go codebases.

Check these DDD patterns:
1. Aggregate has version field (optimistic locking)
2. Aggregate emits domain events
3. Repository interface exists
4. No infrastructure dependencies

Output format:
- ✅ Pattern name: Compliant
- ❌ Pattern name: Violation (explanation)

Example:
✅ Optimistic Locking: Contact aggregate has version field
❌ Events: No events emitted in NewContact()
```

### Iteration 4: Re-execute & Re-critique
Continue until prompt quality meets threshold.

## Output

Final refined prompt saved to `.claude/agents/ddd-validator.md`
```

**Prós**: Auto-otimização, melhoria contínua
**Contras**: Muito custoso, requer múltiplas execuções

---

## 🎨 Técnicas Específicas do Claude

### 11. XML Tags (Estruturação)

**Descrição**: Usar tags XML para estruturar prompts e outputs.

**Por que funciona**: Claude foi treinado com XML tags, responde muito bem a essa estrutura.

**Exemplo em Agente Claude Code**:
```markdown
---
name: structured-analyzer
description: Use XML tags for clear structure
tools: Read, Grep
model: sonnet
---

# Structured Analyzer Agent

You analyze code using XML tags for clarity.

## Input Format

You receive tasks in this format:

```xml
<task>
  <type>security_audit</type>
  <target>
    <file>infrastructure/http/middleware/auth.go</file>
    <focus>authentication bypass vulnerabilities</focus>
  </target>
  <context>
    <severity>P0</severity>
    <reference>TODO.md line 45</reference>
  </context>
</task>
```

## Output Format

Your response must follow this structure:

```xml
<analysis>
  <summary>
    Brief overview of findings
  </summary>

  <findings>
    <finding severity="critical">
      <title>Dev Mode Authentication Bypass</title>
      <location>
        <file>infrastructure/http/middleware/auth.go</file>
        <line>41</line>
      </location>
      <description>
        The middleware bypasses authentication when ENV != "production",
        but this check can be circumvented by setting ENV to any non-empty
        value other than "production".
      </description>
      <evidence>
        <code>
          if os.Getenv("ENV") != "production" {
              return c.Next() // BYPASS!
          }
        </code>
      </evidence>
      <impact>
        Allows unauthorized access to all endpoints in staging/dev environments
        if attacker can influence ENV variable.
      </impact>
      <recommendation>
        <fix>
          Use explicit allowlist: if ENV == "development" && ALLOW_DEV_BYPASS == "true"
        </fix>
        <priority>P0</priority>
      </recommendation>
    </finding>

    <finding severity="high">
      [...]
    </finding>
  </findings>

  <statistics>
    <total_findings>3</total_findings>
    <critical>1</critical>
    <high>2</high>
    <medium>0</medium>
  </statistics>

  <next_steps>
    <step priority="1">Fix critical finding immediately</step>
    <step priority="2">Review all ENV checks in codebase</step>
  </next_steps>
</analysis>
```

## Benefits

- **Parseability**: Easy to extract specific sections programmatically
- **Clarity**: Clear hierarchy and relationships
- **Consistency**: Enforced structure
```

**Uso em Comandos**:
```markdown
---
description: Analyze security with structured XML output
---

Analyze the security of the following component:

<component>
  <name>$1</name>
  <type>$2</type>
</component>

Use XML tags in your response for easy parsing.
```

**Prós**: Estrutura clara, parseável, suportado nativamente
**Contras**: Mais verboso que JSON

---

### 12. Prefill (Forçar Formato de Saída)

**Descrição**: Pré-preencher o início da resposta do assistente para forçar formato.

**Quando usar**: Quando você quer JSON puro, sem texto introdutório.

**Limitação**: **NÃO funciona com Extended Thinking ativado**.

**Exemplo Conceitual** (API):
```python
# User message
user: "Analyze contact.go and return JSON"

# Prefill assistant response (force JSON output)
assistant: "{"

# Claude continua a partir daqui, completando o JSON
# Output: {"aggregate": "Contact", "has_version": true, ...}
```

**Exemplo em Slash Command**:
```markdown
---
description: Get JSON analysis of aggregate
---

Analyze the $1 aggregate and return results in JSON format.

**Start your response with**:
```json
{
```

This ensures clean JSON output without preamble.
```

**Prós**: Output limpo, sem texto desnecessário
**Contras**: Não funciona com Extended Thinking

---

### 13. Extended Thinking

**Descrição**: Claude aloca tokens extras para "pensar" antes de responder.

**Quando usar**: Provas matemáticas, planejamento de código, lógica complexa.

**Configuração**: Via API ou interface (toggle + budget de tokens).

**Exemplo em Agente Claude Code**:
```markdown
---
name: algorithm-designer
description: Design algorithms using extended thinking
model: sonnet
---

# Algorithm Designer Agent

You design algorithms with deep reasoning.

## Instructions

**Use extended thinking mode** (this is configured externally, not in prompt).

Given an algorithmic problem:
1. **Think deeply** about edge cases, trade-offs, complexity
2. **Explore alternatives** in your thinking block
3. **Deliver concise final answer** after thinking

## Example

**User**: "Design algorithm to detect duplicate sessions in O(n) time"

**Claude's Internal Thinking** (in thinking block, not shown to user):
```
Let me think about this...
- Need O(n) time, so can't use nested loops
- Could use HashMap to track seen session IDs
- But what if sessions have same ID but different timestamps?
- Need composite key: (session_id, contact_id)
- Edge case: What if session_id is nil?
- Should we consider timezone in timestamp comparison?
[...extensive thinking...]
```

**Claude's Final Answer** (shown to user):
```go
func detectDuplicates(sessions []Session) []Session {
    seen := make(map[string]bool)
    duplicates := []Session{}

    for _, s := range sessions {
        key := fmt.Sprintf("%s:%s", s.ID, s.ContactID)
        if seen[key] {
            duplicates = append(duplicates, s)
        }
        seen[key] = true
    }
    return duplicates // O(n) time, O(n) space
}
```

## Prompting Tips for Extended Thinking

- ✅ High-level instructions: "Think deeply about edge cases"
- ✅ Multishot examples of reasoning patterns
- ❌ Step-by-step prescriptive guidance (limits creativity)
- ❌ Prefilling (not allowed with extended thinking)
```

**Prós**: Raciocínio mais profundo, melhores soluções
**Contras**: Usa mais tokens, não permite prefill

---

### 14. Multishot Prompting (Advanced Few-Shot)

**Descrição**: Few-shot avançado com múltiplos exemplos diversos.

**Quando usar**: Tarefas que requerem padrões complexos com variações.

**Exemplo em Agente Claude Code**:
```markdown
---
name: test-generator
description: Generate tests using diverse examples (multishot)
tools: Read, Write
model: sonnet
---

# Test Generator Agent

You generate Go tests using table-driven patterns.

## Example 1: Simple Validation Test

**Code**:
```go
func (c *Contact) SetName(name string) error {
    if name == "" {
        return ErrInvalidName
    }
    c.name = name
    return nil
}
```

**Test**:
```go
func TestContact_SetName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        {"valid name", "John Doe", nil},
        {"empty name", "", ErrInvalidName},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := &Contact{}
            err := c.SetName(tt.input)
            assert.Equal(t, tt.wantErr, err)
        })
    }
}
```

## Example 2: Repository Test with Mock

**Code**:
```go
type ContactRepository interface {
    Save(ctx context.Context, contact *Contact) error
}
```

**Test**:
```go
func TestCreateContactHandler_Handle(t *testing.T) {
    tests := []struct {
        name      string
        setupMock func(*MockRepository)
        wantErr   bool
    }{
        {
            name: "success",
            setupMock: func(m *MockRepository) {
                m.On("Save", mock.Anything, mock.Anything).Return(nil)
            },
            wantErr: false,
        },
        {
            name: "repository error",
            setupMock: func(m *MockRepository) {
                m.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := new(MockRepository)
            tt.setupMock(repo)
            handler := NewCreateContactHandler(repo)
            // [test execution...]
        })
    }
}
```

## Example 3: Event Test

**Code**:
```go
func (c *Contact) Create() {
    c.addEvent(NewContactCreatedEvent(c))
}
```

**Test**:
```go
func TestContact_Create_EmitsEvent(t *testing.T) {
    c := NewContact("John", "john@example.com")

    events := c.DomainEvents()

    require.Len(t, events, 1)
    event, ok := events[0].(ContactCreatedEvent)
    require.True(t, ok)
    assert.Equal(t, "contact.created", event.EventType())
}
```

## Your Task

Now generate tests for the provided code using the patterns shown above.
Match the complexity of the test to the complexity of the code.
```

**Prós**: Aprende padrões complexos com variações
**Contras**: Usa muitos tokens com múltiplos exemplos

---

## 🔄 Aplicação em Agentes Paralelos

### Técnicas que Paralelizam Bem

| Técnica | Paralelização | Como Implementar |
|---------|---------------|------------------|
| **Skeleton Prompting** | ✅ Excelente | Fase 2: Expandir cada ponto em agente separado |
| **Tree of Thoughts** | ✅ Excelente | Cada "thought" em agente separado |
| **Self-Consistency** | ✅ Excelente | 3+ runs independentes em paralelo |
| **Few-Shot** | ✅ Boa | Cada agente usa mesmos exemplos |
| **Recursive** | ✅ Boa | Cada chunk processado em paralelo |
| **Prompt Chaining (Branching)** | ✅ Boa | Múltiplos branches em paralelo |
| **Zero-Shot** | ✅ Básica | Múltiplos agentes, tarefas independentes |

### Técnicas que NÃO Paralelizam

| Técnica | Por que não paraleliza | Alternativa |
|---------|------------------------|-------------|
| **Chain of Thought** | Raciocínio é sequencial | Use por agente, não entre agentes |
| **ReAct** | Requer feedback iterativo | Um agente faz loop interno |
| **Prompt Chaining (Sequential)** | Cada etapa depende da anterior | Use branching onde possível |
| **Extended Thinking** | Pensamento interno do modelo | Ative por agente, não global |

---

### Pattern: Orquestrador + Workers Paralelos

**Estrutura**:
```
[Orchestrator Agent]
     │
     ├─→ [Worker 1] ──┐
     ├─→ [Worker 2] ──┤
     ├─→ [Worker 3] ──┼──→ [Consolidator Agent] ──→ Final Report
     ├─→ [Worker 4] ──┤
     └─→ [Worker 5] ──┘
```

**Exemplo Completo**:
```markdown
---
name: parallel-orchestrator
description: Orchestrate 10 parallel agents for comprehensive analysis
tools: Task, Write
model: opus
---

# Parallel Orchestrator Agent

You coordinate 10 specialized agents running in parallel.

## Phase 1: Parallel Dispatch

Launch these agents **simultaneously** using Task tool:

### Analysis Agents (10 parallel)
1. **domain-analyzer** → Analyze internal/domain/
2. **application-analyzer** → Analyze internal/application/
3. **infrastructure-analyzer** → Analyze infrastructure/
4. **security-auditor** → Check P0 vulnerabilities
5. **test-coverage-checker** → Run make test-coverage
6. **database-schema-auditor** → Validate migrations + RLS
7. **event-pattern-checker** → Verify Outbox Pattern
8. **api-documentation-validator** → Check Swagger completeness
9. **performance-profiler** → Identify bottlenecks
10. **dependency-auditor** → Check go.mod for outdated deps

## Phase 2: Wait for Completion

All agents run in parallel (max 10 concurrent). Wait for all to finish (~2-3 min).

## Phase 3: Consolidation

Launch **report-consolidator** agent with all 10 outputs:
- Merge findings
- Remove duplicates
- Prioritize by severity (P0 > P1 > P2)
- Generate executive summary
- Create action items

## Phase 4: Output

Save to `ai-guides/analysis-reports/analysis-{date}.md`

## Execution Template

```markdown
I'm launching 10 specialized agents in parallel to analyze the codebase:

[Task tool call: domain-analyzer]
[Task tool call: application-analyzer]
[Task tool call: infrastructure-analyzer]
[Task tool call: security-auditor]
[Task tool call: test-coverage-checker]
[Task tool call: database-schema-auditor]
[Task tool call: event-pattern-checker]
[Task tool call: api-documentation-validator]
[Task tool call: performance-profiler]
[Task tool call: dependency-auditor]

Waiting for all agents to complete... (ETA: 2-3 minutes)

[All completed]

Now launching consolidator to merge results...

[Final report generated]
```
```

**Performance**: 10x faster que análise sequencial (2-3 min vs 20-30 min)

---

### Pattern: Map-Reduce com Agentes

**Aplicação**: Analisar grandes volumes de arquivos

**Map Phase** (Parallel):
```markdown
---
name: file-mapper
description: Map phase - analyze one file
tools: Read, Grep
model: sonnet
---

# File Mapper Agent

You analyze **one file** and return structured findings.

**Input**: File path
**Output**: JSON
```json
{
  "file": "path/to/file.go",
  "loc": 250,
  "functions": 12,
  "tests": 8,
  "coverage": 67,
  "issues": [
    {"type": "missing-error-handling", "line": 45}
  ]
}
```
```

**Reduce Phase** (Aggregation):
```markdown
---
name: file-reducer
description: Reduce phase - aggregate all file analyses
tools: Write
model: sonnet
---

# File Reducer Agent

You receive JSON outputs from multiple file-mapper agents.

**Task**: Aggregate into summary:
- Total LOC
- Total functions
- Average coverage
- All issues grouped by type

**Output**: Consolidated report
```

**Orchestrator**:
```markdown
# 1. Find all Go files
files = glob("**/*.go")  # 200 files

# 2. Map phase: Launch 200 file-mapper agents in parallel (batches of 10)
results = []
for batch in chunks(files, 10):
    batch_results = parallel_map(file-mapper, batch)
    results.extend(batch_results)

# 3. Reduce phase: Aggregate
final_report = file-reducer(results)
```

---

## 📦 Templates Reutilizáveis

### Template 1: Agente Analisador com Few-Shot

```markdown
---
name: {agent-name}
description: {when to use this agent}
tools: Read, Grep, Glob
model: sonnet
---

# {Agent Name}

You are a {domain} expert specialized in {specialty}.

## Examples of Good {Pattern}

**Example 1**: ✅ CORRECT
{code example}
**Why correct**: {explanation}

**Example 2**: ✅ CORRECT
{code example}
**Why correct**: {explanation}

## Examples of Bad {Pattern}

**Example 3**: ❌ INCORRECT
{code example}
**Why wrong**: {explanation}

**Example 4**: ❌ INCORRECT
{code example}
**Why wrong**: {explanation}

## Your Task

Analyze the codebase for {pattern} and report findings using the format above.

## Output Format

```markdown
## {Pattern} Analysis Report

### ✅ Compliant Code
- {file}:{line} - {description}

### ❌ Violations
- {file}:{line} - {description} - {why wrong} - {how to fix}

### Statistics
- Total checked: X
- Compliant: Y (Z%)
- Violations: W
```
```

---

### Template 2: Agente Orquestrador Paralelo

```markdown
---
name: {orchestrator-name}
description: Orchestrate {N} parallel agents for {task}
tools: Task, Write
model: opus
---

# {Orchestrator Name}

Coordinate {N} specialized agents running in parallel.

## Phase 1: Parallel Dispatch

Launch these agents simultaneously:

{list of agents with descriptions}

## Phase 2: Wait for Completion

All {N} agents run in parallel. Wait for completion.

## Phase 3: Consolidation

Launch consolidator with all outputs to:
- Merge findings
- Remove duplicates
- Prioritize
- Generate summary

## Phase 4: Output

Save to {output path}
```

---

### Template 3: Agente com Chain of Thought

```markdown
---
name: {agent-name}
description: {when to use} - uses step-by-step reasoning
tools: Read, Grep, Bash
model: sonnet
---

# {Agent Name} (Chain of Thought)

You analyze {domain} using systematic step-by-step reasoning.

## Reasoning Process

### Step 1: {step name}
**Goal**: {what to achieve}
**Action**: {what to do}
**Reasoning**: {why this step}

### Step 2: {step name}
**Goal**: {what to achieve}
**Action**: {what to do}
**Reasoning**: {why this step}

[continue for all steps...]

## Your Task

Follow this reasoning process step-by-step. **Show your thinking for each step**.

## Output Format

```markdown
## {Analysis Name} (Chain of Thought)

### Step 1: {step name}
**Goal**: {goal}
**Action**: {what I did}
**Result**: {what I found}
**Reasoning**: {why this matters}

[continue for each step...]

## Final Conclusion
{consolidated findings}
```
```

---

### Template 4: Agente com XML Structured Output

```markdown
---
name: {agent-name}
description: {when to use} - returns structured XML
tools: Read, Grep
model: sonnet
---

# {Agent Name} (Structured Output)

You analyze {domain} and return structured XML output.

## Input Format

```xml
<task>
  <type>{task type}</type>
  <target>{what to analyze}</target>
  <parameters>
    <param name="{name}">{value}</param>
  </parameters>
</task>
```

## Output Format

Your response MUST be valid XML:

```xml
<analysis>
  <summary>
    {brief overview}
  </summary>

  <findings>
    <finding severity="{critical|high|medium|low}">
      <title>{finding title}</title>
      <location>
        <file>{file path}</file>
        <line>{line number}</line>
      </location>
      <description>{detailed description}</description>
      <recommendation>{how to fix}</recommendation>
    </finding>
  </findings>

  <statistics>
    <total_findings>{count}</total_findings>
  </statistics>
</analysis>
```

## Benefits

- Programmatically parseable
- Clear structure
- Easy to extract specific sections
```

---

### Template 5: Comando com Skeleton Prompting

```markdown
---
description: {command description} - uses skeleton prompting for speed
---

# {Command Name}

This command uses **skeleton prompting** (2-phase parallel processing) for fast comprehensive analysis.

## Phase 1: Generate Skeleton

First, generate a high-level outline:

**Task**: {main task}

**Generate skeleton with 5-10 points (3-5 words each)**:
1. {point 1}
2. {point 2}
[...]

## Phase 2: Expand in Parallel

For each skeleton point, launch a specialized agent to expand in detail.

**Parallelization**: All points expanded simultaneously (5-10 agents in parallel)

## Execution

Run this command with: `/{command-name} {args}`

Expected completion time: {time} (vs {sequential time} sequentially)
```

---

## ✅ Best Practices

### 1. **Escolha a Técnica Certa**

| Se você quer... | Use... |
|-----------------|--------|
| Resposta rápida e simples | Zero-Shot |
| Seguir padrão específico | Few-Shot / Multishot |
| Raciocínio transparente | Chain of Thought |
| Conteúdo longo estruturado | Skeleton Prompting |
| Explorar alternativas | Tree of Thoughts |
| Debugging iterativo | ReAct |
| Validar análise crítica | Self-Consistency |
| Workflow multi-etapa | Prompt Chaining |
| Processar grande volume | Recursive Prompting |
| Output estruturado | XML Tags + Prefill |
| Problema complexo | Extended Thinking |

---

### 2. **Combine Técnicas**

**Exemplo**: Few-Shot + Chain of Thought
```markdown
## Examples (Few-Shot)
Example 1: {pattern}
Example 2: {pattern}

## Your Task (Chain of Thought)
Step 1: Understand context
Step 2: Apply pattern from examples
Step 3: Verify correctness
```

**Exemplo**: Skeleton + Self-Consistency
```markdown
Phase 1: Generate skeleton
Phase 2: Expand each point 3 times independently (self-consistency)
Phase 3: Choose most consistent expansion for each point
```

---

### 3. **Otimize para Paralelização**

**❌ Evite dependências sequenciais**:
```markdown
# BAD: Sequential chain
1. Analyze domain → wait
2. Analyze application (depends on step 1) → wait
3. Analyze infrastructure (depends on step 2) → wait
```

**✅ Prefira análises independentes**:
```markdown
# GOOD: Parallel branches
1. Analyze domain ┐
2. Analyze application ├─→ All run in parallel → Consolidate
3. Analyze infrastructure ┘
```

---

### 4. **Estruture Prompts Claramente**

**Ordem recomendada**:
```markdown
1. **Role/Identity**: "You are a {expert} specialized in {domain}"
2. **Context**: "This codebase uses {architecture} with {patterns}"
3. **Task**: "Your task is to {specific goal}"
4. **Constraints**: "Do NOT {forbidden action}. ALWAYS {required action}"
5. **Examples** (if Few-Shot): Example 1, 2, 3...
6. **Process** (if CoT/ReAct): Step 1, Step 2, Step 3...
7. **Output Format**: "Your response must be {format}"
8. **Quality Criteria**: "Good output has {qualities}"
```

---

### 5. **Use XML Tags para Separar Seções**

```markdown
<role>
You are a security expert
</role>

<context>
This is a Go web application using Gin framework
</context>

<task>
Audit for authentication vulnerabilities
</task>

<constraints>
- Do NOT modify code
- ALWAYS provide evidence (code snippets)
</constraints>

<output_format>
```xml
<findings>
  <finding severity="...">
    ...
  </finding>
</findings>
```
</output_format>
```

**Benefício**: Claude entende estrutura melhor, você pode referir "as described in <task>"

---

### 6. **Seja Específico, Não Vago**

**❌ Vago**:
```markdown
Analyze the code and find issues.
```

**✅ Específico**:
```markdown
Analyze internal/domain/crm/contact/contact.go for:
1. Optimistic locking: Check if Contact struct has `version int` field
2. Event emission: Check if NewContact() emits ContactCreatedEvent
3. Repository interface: Check if ContactRepository interface is defined
4. Layer purity: Check if contact.go imports any infrastructure packages (violation)

Report findings in this format:
- ✅ {check}: Compliant ({explanation})
- ❌ {check}: Violation ({explanation} + {how to fix})
```

---

### 7. **Defina Success Criteria**

```markdown
## Success Criteria

Your analysis is successful if:
1. ✅ All 30 aggregates are checked (100% coverage)
2. ✅ Each finding includes: file path, line number, explanation, fix
3. ✅ Findings are prioritized (Critical > High > Medium > Low)
4. ✅ No false positives (every issue is real)
5. ✅ Output is parseable JSON/XML

If any criterion is not met, the analysis is incomplete.
```

---

### 8. **Itere e Refine**

```markdown
## Iteration Strategy

**First pass**: Quick scan with Zero-Shot
**If results unclear**: Add Few-Shot examples
**If still unclear**: Add Chain of Thought reasoning
**If critical decision**: Add Self-Consistency (3 runs)

Iterate until quality meets success criteria.
```

---

### 9. **Documente Técnicas Usadas**

```markdown
---
name: security-auditor
description: Security audit using Self-Consistency + Few-Shot
techniques:
  - Few-Shot: Examples of vulnerabilities
  - Self-Consistency: 3 independent runs
  - XML Tags: Structured output
tools: Read, Grep
model: sonnet
---
```

**Benefício**: Fácil entender e modificar depois

---

### 10. **Teste com Dados Reais**

```markdown
## Testing

Before deploying this agent:
1. Test with 3-5 real files from codebase
2. Verify output format is correct
3. Check for false positives/negatives
4. Measure execution time
5. Adjust prompt based on results

Iteration is key to prompt quality.
```

---

## 🎓 Estudo de Caso: Sistema Multi-Agente Completo

### Cenário

**Task**: "Analyze the entire Ventros CRM codebase comprehensively and generate actionable roadmap"

### Solução: Combinação de Técnicas

#### Agente 1: Orchestrator (Meta Prompting + Prompt Chaining)
```markdown
---
name: master-orchestrator
description: Top-level orchestrator using meta-prompting
tools: Task, Write
model: opus
techniques:
  - Meta Prompting: Self-optimizes workflow
  - Prompt Chaining: Sequential stages
---

# Master Orchestrator

## Stage 1: Planning (Meta Prompting)
- Analyze task requirements
- Generate optimal workflow
- Refine workflow based on codebase size

## Stage 2: Parallel Analysis (Branching Chain)
- Launch 10 specialized agents in parallel

## Stage 3: Consolidation
- Merge findings
- Generate roadmap

## Stage 4: Validation (Self-Consistency)
- Re-run critical findings 3 times
- Keep only consistent results
```

#### Agente 2-11: Specialized Analyzers (Skeleton + Few-Shot)
```markdown
---
name: domain-analyzer
description: Analyze domain layer using Skeleton + Few-Shot
tools: Read, Grep, Glob
model: sonnet
techniques:
  - Skeleton Prompting: Generate outline → expand in parallel
  - Few-Shot: Examples of DDD patterns
---

# Domain Analyzer

## Phase 1: Skeleton
Generate outline of all 30 aggregates

## Phase 2: Expand (Parallel)
For each aggregate, check:
- Version field (Example: {example})
- Events (Example: {example})
- Repository (Example: {example})

Launch 30 sub-agents in parallel (batches of 10)
```

#### Agente 12: Consolidator (Tree of Thoughts)
```markdown
---
name: report-consolidator
description: Consolidate findings using Tree of Thoughts
tools: Write
model: opus
techniques:
  - Tree of Thoughts: Explore multiple report structures
---

# Report Consolidator

## Thought 1: Group by Layer
Domain → Application → Infrastructure

## Thought 2: Group by Severity
P0 → P1 → P2

## Thought 3: Group by Type
Architecture → Security → Performance

## Evaluation
Choose best structure based on:
- Clarity for developers
- Actionability
- Priority visibility

## Output
Generate report using best structure
```

### Resultado

**Execução**:
1. Master orchestrator planeja (30s)
2. Lança 10 agentes paralelos (2 min)
3. Cada agente usa skeleton (sub-parallelization) (2 min)
4. Consolidator usa ToT para estruturar (30s)

**Total**: ~5 minutos (vs 30+ min sequencial)

**Técnicas usadas**: 6 diferentes (Meta, Chaining, Skeleton, Few-Shot, ToT, Parallel)

**Output**: Relatório estruturado com roadmap priorizado

---

## 📚 Referências

### Documentação Oficial
- [Claude Prompt Engineering](https://docs.claude.com/en/docs/build-with-claude/prompt-engineering)
- [Claude 4 Best Practices](https://docs.claude.com/en/docs/build-with-claude/prompt-engineering/claude-4-best-practices)
- [Extended Thinking Tips](https://docs.claude.com/en/docs/build-with-claude/prompt-engineering/extended-thinking-tips)
- [XML Tags Guide](https://docs.claude.com/en/docs/build-with-claude/prompt-engineering/use-xml-tags)

### Guias de Técnicas
- [Prompt Engineering Guide](https://www.promptingguide.ai/)
- [Learn Prompting](https://learnprompting.org/)
- [Chain of Thought](https://www.promptingguide.ai/techniques/cot)
- [Tree of Thoughts](https://www.promptingguide.ai/techniques/tot)
- [ReAct Prompting](https://www.promptingguide.ai/techniques/react)
- [Skeleton-of-Thought](https://learnprompting.org/docs/advanced/decomposition/skeleton_of_thoughts)

### Artigos e Papers
- [Self-Consistency Improves Chain of Thought](https://arxiv.org/abs/2203.11171)
- [ReAct: Synergizing Reasoning and Acting](https://arxiv.org/abs/2210.03629)
- [Tree of Thoughts: Deliberate Problem Solving](https://arxiv.org/abs/2305.10601)
- [Meta Prompting for AI Systems](https://arxiv.org/abs/2311.11482)

### Recursos da Comunidade
- [Awesome Claude Prompts](https://github.com/langgptai/awesome-claude-prompts)
- [Claude Cookbooks](https://github.com/anthropics/anthropic-cookbook)
- [Anthropic Prompt Tutorial](https://github.com/anthropics/prompt-eng-interactive-tutorial)

---

## 🎬 Próximos Passos

Agora que você domina as técnicas, você pode:

1. **Criar agentes especializados** usando templates deste guia
2. **Combinar técnicas** para casos complexos
3. **Otimizar para paralelização** usando padrões apresentados
4. **Implementar workflows** completos com orquestração
5. **Iterar e refinar** prompts baseado em resultados reais

**Recomendação**: Comece simples (Zero-Shot → Few-Shot → CoT) e aumente complexidade conforme necessário.

---

**Last Updated**: 2025-10-15
**Maintainer**: Ventros CRM Team
**Version**: 1.0
**Related Guides**:
- `claude-code-guide.md` - Sistema multi-agente do Claude Code
- `CLAUDE.md` - Guia completo do desenvolvedor
