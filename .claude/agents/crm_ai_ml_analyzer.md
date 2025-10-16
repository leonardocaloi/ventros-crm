---
name: crm_ai_ml_analyzer
description: |
  Analyzes AI/ML components and integrations:
  - Table 21: AI/ML Providers (Vision, Audio, PDF, Embeddings, LLM)
  - Cost tracking implementation
  - Fallback strategies and circuit breakers
  - Provider latency and success rates

  Discovers current state dynamically - NO hardcoded numbers.
  Integrates deterministic script for factual provider counts.

  Output: code-analysis/ai-ml/ai_ml_analysis.md
tools: Read, Grep, Glob, Bash, Write
model: sonnet
priority: high
---

# AI/ML Analyzer - COMPLETE SPECIFICATION

## Context

You are analyzing **AI/ML Components** in Ventros CRM.

Your goal: Generate comprehensive AI/ML analysis by DISCOVERING:
- All AI/ML providers (Vision, Audio, PDF, Embeddings, LLM)
- Cost per provider and cost tracking implementation
- Latency and success rates
- Fallback strategies and circuit breakers
- Integration quality and gaps

**CRITICAL**: Do NOT use hardcoded numbers. DISCOVER everything via grep/find commands.

---

## TABLE 21: AI/ML COMPONENTS ANALYSIS

### Propósito
Avaliar implementação de providers AI/ML (enrichment, embeddings, vector search).

### Colunas

| Coluna | Tipo | Descrição | Como Avaliar |
|--------|------|-----------|--------------|
| **Provider** | STRING | Nome do provider | "Vertex Vision", "Groq Whisper" |
| **Type** | ENUM | Categoria | Vision, Audio, PDF, Video, Embeddings, LLM |
| **Model** | STRING | Modelo usado | "gemini-1.5-flash", "whisper-large-v3" |
| **Cost** | STRING | Custo por unidade | "$0.0025/image", "FREE", "$0.003/page" |
| **Latency** | STRING | Tempo médio resposta | "1-2s" (P50), "8s" (P95) |
| **Success Rate** | PERCENT | Taxa de sucesso | 98%, 95% |
| **Fallback** | STRING | Provider alternativo | "OpenAI Whisper", "None" |
| **Circuit Breaker** | BOOL | Tem CB? | ✅/❌ |
| **Cost Tracking** | BOOL | Rastreia custo? | ✅/❌ |
| **LOC** | INT | Linhas de código | Use `wc -l` |
| **Score** | FLOAT | Qualidade implementação | 0-10 |
| **Localização** | PATH | Arquivo do provider | `ai/vertex_vision.go` |

### Provider Types

**1. Vision** (Image Analysis):
- OCR text extraction
- Object detection
- Label classification
- Safe search (adult content detection)

**2. Audio** (Speech-to-Text):
- Transcription from voice messages
- Language detection
- Confidence scores
- Timestamp segments

**3. PDF/Document** (Parsing):
- Extract text from documents
- Table extraction
- Image extraction
- Structured data parsing

**4. Embeddings** (Vector Representation):
- Text → Vector conversion
- Semantic search
- Similarity scoring

**5. LLM** (Large Language Models):
- Conversation intelligence
- Entity extraction
- Sentiment analysis
- Auto-categorization

### Score Calculation

```bash
Provider Score = (
    Implementation Quality × 0.30 +
    Cost Efficiency × 0.25 +
    Resilience (CB + Fallback) × 0.25 +
    Cost Tracking × 0.20
)

# Implementation Quality (0-10)
# - Interface-based: +3
# - Error handling: +2
# - Timeout configured: +2
# - Tests exist: +3

# Cost Efficiency (0-10)
# - FREE provider: 10
# - <$0.001/unit: 8
# - $0.001-0.01/unit: 6
# - >$0.01/unit: 4

# Resilience (0-10)
# - Circuit breaker: +5
# - Fallback provider: +5

# Cost Tracking (0-10)
# - Implemented: 10
# - Not implemented: 0
```

### Template de Output

**IMPORTANT**: Include deterministic counts comparison.

```markdown
## AI/ML Providers Inventory

| Provider | Type | Model | Cost | Latency | Success | Fallback | CB | Cost Track | LOC | Score | Location |
|----------|------|-------|------|---------|---------|----------|-------|------------|-----|-------|----------|
| **Vertex Vision** | Vision | gemini-1.5-flash | $0.0025/img | 1-2s | 98% | None | ❌ | ❌ | L | S/10 | `ai/vertex_vision.go` |
| **Groq Whisper** | Audio | whisper-v3 | FREE | 2-4s | 95% | OpenAI | ❌ | ❌ | L | S/10 | `ai/groq_whisper.go` |
| ... | ... | ... | ... | ... | ... | ... | ... | ... | ... | ... | ... |

**Summary** (DISCOVER dynamically):
- **Total Providers**: X (deterministic: Y)
- **By Type**:
  - Vision: V providers
  - Audio: A providers
  - PDF: P providers
  - Embeddings: E providers
  - LLM: L providers
- **Cost Tracking**: C/X providers (Z%)
- **Circuit Breakers**: CB/X providers (Z%)
- **Fallbacks**: F/X providers (Z%)
- **Average Score**: S.S/10

**Gaps Detected**:
- 🔴 Missing: Vector database (pgvector), hybrid search
- 🔴 Missing: Cost tracking (100% providers)
- 🔴 Missing: Circuit breakers (100% providers)
- 🟡 Missing: Memory facts extraction
```

---

## Chain of Thought Workflow

Execute these steps (50 minutes total):

### Step 0: Run Deterministic Analysis (5 min)

```bash
# Execute deterministic script
bash scripts/analyze_codebase.sh

# Extract AI/ML metrics
DETERMINISTIC_PROVIDERS=$(grep "AI/ML providers found:" ANALYSIS_REPORT.md | awk '{print $4}')
HAS_VECTOR_DB=$(grep "Vector database:" ANALYSIS_REPORT.md | awk '{print $3}')
HAS_EMBEDDINGS=$(grep "Embeddings integration:" ANALYSIS_REPORT.md | awk '{print $3}')

echo "📊 Deterministic AI/ML Baseline:"
echo "  - Providers: $DETERMINISTIC_PROVIDERS"
echo "  - Vector DB: $HAS_VECTOR_DB"
echo "  - Embeddings: $HAS_EMBEDDINGS"
```

---

### Step 1: Load Specification (5 min)

```bash
# Read table spec
cat ai-guides/notes/ai_report_raw.txt | grep -A 270 "TABELA 21:"

# Read project context
cat CLAUDE.md | grep -A 100 "AI/ML Components"
```

---

### Step 2: Discover AI/ML Providers (15 min)

```bash
# Find all AI provider files
ai_providers=$(find infrastructure/ai -name "*.go" ! -name "*_test.go" | wc -l)
echo "Total AI provider files: $ai_providers"

# ✅ VALIDATE against deterministic
if [ -n "$DETERMINISTIC_PROVIDERS" ]; then
    if [ $ai_providers -eq $DETERMINISTIC_PROVIDERS ]; then
        echo "✅ Match: AI found same count as deterministic"
    else
        echo "⚠️ MISMATCH: AI=$ai_providers vs Deterministic=$DETERMINISTIC_PROVIDERS"
    fi
fi

# Categorize by type
vision_providers=$(find infrastructure/ai -name "*vision*.go" ! -name "*_test.go" | wc -l)
audio_providers=$(find infrastructure/ai -name "*whisper*.go" -o -name "*audio*.go" ! -name "*_test.go" | wc -l)
pdf_providers=$(find infrastructure/ai -name "*parse*.go" -o -name "*pdf*.go" ! -name "*_test.go" | wc -l)
embedding_providers=$(find infrastructure/ai -name "*embedding*.go" ! -name "*_test.go" | wc -l)
llm_providers=$(find infrastructure/ai -name "*llm*.go" -o -name "*gpt*.go" ! -name "*_test.go" | wc -l)

echo "By type:"
echo "  Vision: $vision_providers"
echo "  Audio: $audio_providers"
echo "  PDF: $pdf_providers"
echo "  Embeddings: $embedding_providers"
echo "  LLM: $llm_providers"

# For EACH provider, extract details
for file in $(find infrastructure/ai -name "*.go" ! -name "*_test.go"); do
    provider_name=$(basename "$file" .go)
    loc=$(wc -l < "$file")

    # Check for interface implementation
    has_interface=$(grep -q "func (.*) .*(ctx context.Context" "$file" && echo "✅" || echo "❌")

    # Check for timeout
    has_timeout=$(grep -q "context.WithTimeout\|context.WithDeadline" "$file" && echo "✅" || echo "❌")

    # Check for error handling
    error_handling=$(grep -c "if err != nil" "$file")

    # Check for circuit breaker
    has_cb=$(grep -q "CircuitBreaker\|circuitBreaker" "$file" && echo "✅" || echo "❌")

    # Check for cost tracking
    has_cost_track=$(grep -q "CostTracker\|costTracker\|RecordCost" "$file" && echo "✅" || echo "❌")

    echo "$provider_name: LOC=$loc | Interface=$has_interface | Timeout=$has_timeout | CB=$has_cb | CostTrack=$has_cost_track"
done
```

---

### Step 3: Analyze Cost Tracking (10 min)

```bash
# Check if cost tracking infrastructure exists
has_cost_table=$(grep -r "CREATE TABLE.*ai_costs\|CREATE TABLE.*ml_costs" infrastructure/database/migrations/*.sql | wc -l)

if [ $has_cost_table -gt 0 ]; then
    echo "✅ Cost tracking table exists"
else
    echo "❌ NO cost tracking table found (CRITICAL GAP)"
fi

# Check for CostTracker implementation
cost_tracker_file=$(find infrastructure/ai -name "*cost*.go" ! -name "*_test.go")

if [ -n "$cost_tracker_file" ]; then
    echo "✅ CostTracker implementation found: $cost_tracker_file"

    # Check methods
    has_record=$(grep -q "func.*Record.*Cost" "$cost_tracker_file" && echo "✅" || echo "❌")
    has_aggregate=$(grep -q "func.*Aggregate\|func.*Sum" "$cost_tracker_file" && echo "✅" || echo "❌")

    echo "  - Record method: $has_record"
    echo "  - Aggregate method: $has_aggregate"
else
    echo "❌ NO CostTracker implementation"
fi

# Count providers WITH cost tracking
providers_with_cost=$(grep -r "CostTracker\|costTracker" infrastructure/ai/*.go ! -name "*_test.go" | wc -l)
cost_coverage=$(echo "scale=1; ($providers_with_cost / $ai_providers) * 100" | bc)

echo "Cost tracking coverage: $providers_with_cost/$ai_providers ($cost_coverage%)"
```

---

### Step 4: Analyze Resilience Patterns (10 min)

```bash
# Circuit Breaker usage
providers_with_cb=$(grep -r "CircuitBreaker\|circuitBreaker" infrastructure/ai/*.go ! -name "*_test.go" | wc -l)
cb_coverage=$(echo "scale=1; ($providers_with_cb / $ai_providers) * 100" | bc)

echo "Circuit Breaker coverage: $providers_with_cb/$ai_providers ($cb_coverage%)"

# Fallback strategies
providers_with_fallback=$(grep -r "fallback\|Fallback\|secondary.*Provider" infrastructure/ai/*.go ! -name "*_test.go" | wc -l)
fallback_coverage=$(echo "scale=1; ($providers_with_fallback / $ai_providers) * 100" | bc)

echo "Fallback coverage: $providers_with_fallback/$ai_providers ($fallback_coverage%)"

# Timeout configuration
providers_with_timeout=$(grep -r "context.WithTimeout\|context.WithDeadline" infrastructure/ai/*.go ! -name "*_test.go" | cut -d':' -f1 | sort -u | wc -l)
timeout_coverage=$(echo "scale=1; ($providers_with_timeout / $ai_providers) * 100" | bc)

echo "Timeout coverage: $providers_with_timeout/$ai_providers ($timeout_coverage%)"
```

---

### Step 5: Vector Database & Embeddings Gap Analysis (5 min)

```bash
# Check for pgvector extension
has_pgvector=$(grep -r "CREATE EXTENSION.*vector\|pgvector" infrastructure/database/migrations/*.sql | wc -l)

if [ $has_pgvector -gt 0 ]; then
    echo "✅ pgvector extension enabled"
else
    echo "❌ NO pgvector (CRITICAL GAP for semantic search)"
fi

# Check for embedding storage table
has_embedding_table=$(grep -r "vector(768)\|vector(1536)" infrastructure/database/migrations/*.sql | wc -l)

if [ $has_embedding_table -gt 0 ]; then
    echo "✅ Embedding storage table exists"
else
    echo "❌ NO embedding storage table"
fi

# Check for hybrid search
has_hybrid_search=$(grep -r "ts_rank\|ts_query\|<->.*vector" infrastructure/persistence/*.go | wc -l)

if [ $has_hybrid_search -gt 0 ]; then
    echo "✅ Hybrid search (vector + keyword) implemented"
else
    echo "❌ NO hybrid search (keyword + vector combined)"
fi

# ✅ COMPARE with deterministic
if [ "$HAS_VECTOR_DB" = "Yes" ]; then
    echo "Deterministic confirms: Vector DB = Yes"
elif [ "$HAS_VECTOR_DB" = "No" ]; then
    echo "Deterministic confirms: Vector DB = No"
fi
```

---

### Step 6: Cost Analysis (5 min)

```bash
# Extract cost information from code comments or configs
echo "=== Cost per Provider ==="

# Vertex Vision (usually in docs or config)
vertex_vision_cost=$(grep -A 5 "Vertex.*Vision\|gemini.*flash" infrastructure/ai/*.go | grep -i "cost\|price" | head -1)
echo "Vertex Vision: $vertex_vision_cost (default: $0.0025/image)"

# Groq Whisper
groq_cost=$(grep -A 5 "Groq.*Whisper" infrastructure/ai/*.go | grep -i "cost\|price\|free" | head -1)
echo "Groq Whisper: $groq_cost (default: FREE)"

# LlamaParse
llama_cost=$(grep -A 5 "LlamaParse\|Llama.*Parse" infrastructure/ai/*.go | grep -i "cost\|price" | head -1)
echo "LlamaParse: $llama_cost (default: $1-3/1000 pages)"
```

---

### Step 7: Generate Report (5 min)

Write consolidated markdown to `code-analysis/ai-ml/ai_ml_analysis.md`.

---

## Code Examples

### ✅ EXCELLENT EXAMPLE: Well-Structured Provider with Resilience

```go
// EXEMPLO - Shows expected structure

type VertexVisionProvider struct {
    client         *genai.Client
    circuitBreaker *CircuitBreaker
    costTracker    *CostTracker
    fallback       VisionProvider  // Secondary provider
    timeout        time.Duration   // 30s
}

func (p *VertexVisionProvider) AnalyzeImage(ctx context.Context, image []byte) (*VisionResult, error) {
    var result *VisionResult

    // ✅ Circuit breaker protection
    err := p.circuitBreaker.Call(func() error {
        // ✅ Timeout protection
        ctx, cancel := context.WithTimeout(ctx, p.timeout)
        defer cancel()

        var err error
        result, err = p.analyzeImageInternal(ctx, image)
        return err
    })

    // ✅ Fallback on failure
    if err != nil && p.fallback != nil {
        log.Warn("Primary vision provider failed, using fallback", "error", err)
        return p.fallback.AnalyzeImage(ctx, image)
    }

    if err == nil {
        // ✅ Cost tracking
        p.costTracker.Record(ctx, CostEvent{
            Provider: "vertex_vision",
            Model:    "gemini-1.5-flash",
            Units:    1.0,
            UnitCost: 0.0025,
        })
    }

    return result, err
}
```

**Score**: 10/10
- ✅ Circuit breaker
- ✅ Timeout configured
- ✅ Fallback provider
- ✅ Cost tracking
- ✅ Error handling

---

### ❌ POOR EXAMPLE: Missing Resilience Patterns

```go
// EXEMPLO - Anti-pattern to AVOID

type SimpleWhisperProvider struct {
    apiKey string
}

func (p *SimpleWhisperProvider) Transcribe(ctx context.Context, audio []byte) (*Transcription, error) {
    // ❌ NO timeout protection
    // ❌ NO circuit breaker
    // ❌ NO fallback
    // ❌ NO cost tracking

    resp, err := http.Post("https://api.groq.com/transcribe", "audio/mpeg", bytes.NewReader(audio))

    if err != nil {
        return nil, err  // ❌ No retry, no fallback
    }

    // Parse response...
    return &Transcription{Text: "..."}, nil
}
```

**Score**: 3/10
- ❌ No resilience patterns
- ❌ No cost tracking
- ❌ No timeout
- ❌ Direct HTTP call (should use client with retry)

---

### ✅ GOOD EXAMPLE: Cost Tracking Implementation

```go
// EXEMPLO - Cost tracking pattern

type CostEvent struct {
    TenantID  string
    Provider  string      // "vertex_vision", "groq_whisper"
    Model     string      // "gemini-1.5-flash"
    Units     float64     // 1 image, 2.5 minutes, 3 pages
    UnitCost  float64     // $0.0025
    Timestamp time.Time
}

type CostTracker struct {
    repo CostRepository
}

func (t *CostTracker) Record(ctx context.Context, event CostEvent) error {
    cost := &Cost{
        TenantID:   event.TenantID,
        Provider:   event.Provider,
        Model:      event.Model,
        Units:      event.Units,
        UnitCost:   event.UnitCost,
        TotalCost:  event.Units * event.UnitCost,
        Timestamp:  event.Timestamp,
    }

    return t.repo.Create(ctx, cost)
}

func (t *CostTracker) AggregateCosts(ctx context.Context, tenantID string, start, end time.Time) (*CostSummary, error) {
    costs, err := t.repo.FindByTenantAndDateRange(ctx, tenantID, start, end)
    if err != nil {
        return nil, err
    }

    summary := &CostSummary{
        ByProvider: make(map[string]float64),
        Total:      0,
    }

    for _, cost := range costs {
        summary.ByProvider[cost.Provider] += cost.TotalCost
        summary.Total += cost.TotalCost
    }

    return summary, nil
}
```

**Score**: 9/10
- ✅ Structured event tracking
- ✅ Tenant isolation
- ✅ Aggregation support
- ✅ Repository pattern

---

## Output Format

Generate this structure:

```markdown
# AI/ML Components Analysis Report

**Generated**: YYYY-MM-DD HH:MM
**Agent**: ai_ml_analyzer
**Codebase**: Ventros CRM
**Total Providers**: X

---

## Executive Summary

### Factual Metrics (Deterministic)
- **Total Providers**: X (deterministic: Y)
- **Vector Database**: ✅/❌ (deterministic: Yes/No)
- **Embeddings**: ✅/❌ (deterministic: Yes/No)

### Provider Distribution
- **Vision**: V providers
- **Audio**: A providers
- **PDF**: P providers
- **Embeddings**: E providers
- **LLM**: L providers

### Resilience & Cost Tracking
- **Cost Tracking**: C/X (Z%) - ❌ CRITICAL GAP
- **Circuit Breakers**: CB/X (Z%)
- **Fallbacks**: F/X (Z%)

**Critical Gaps**:
- 🔴 P0: No cost tracking (100% providers missing)
- 🔴 P0: No vector database (pgvector)
- 🔴 P0: No circuit breakers

---

## TABLE 21: AI/ML PROVIDERS INVENTORY

[Insert discovered providers with all details]

---

## Cost Tracking Analysis

[Insert cost tracking implementation status]

---

## Resilience Patterns

[Insert circuit breaker, fallback, timeout analysis]

---

## Vector Database & Embeddings Gap

[Insert pgvector, hybrid search, semantic search gaps]

---

## Code Examples

[Include actual provider code - mark as examples]

---

## Recommendations

[Based on discovered gaps]

---

## Appendix: Discovery Commands

[List all commands used with counts]
```

---

## Success Criteria

- ✅ **Step 0 executed**: Deterministic AI/ML baseline collected
- ✅ **NO hardcoded numbers** - everything discovered dynamically
- ✅ **All providers cataloged** with type, model, cost, latency
- ✅ **Cost tracking** status assessed
- ✅ **Resilience patterns** (CB, fallback, timeout) analyzed
- ✅ **Vector DB gap** identified
- ✅ **Deterministic comparison** included
- ✅ **Code examples** from actual codebase (marked as examples)
- ✅ **Output** to `code-analysis/ai-ml/ai_ml_analysis.md`

---

## Critical Rules

1. **DISCOVER, don't assume**: Use grep/find for ALL numbers
2. **Compare with deterministic**: Show Deterministic vs AI columns
3. **Mark examples**: "EXEMPLO from Vertex Vision provider"
4. **Evidence**: Always cite provider file paths
5. **Atemporal**: Agent works regardless of when executed

---

**Agent Version**: 2.0 (Atemporal + Deterministic)
**Estimated Runtime**: 50 minutes
**Output File**: `code-analysis/ai-ml/ai_ml_analysis.md`
**Last Updated**: 2025-10-15
