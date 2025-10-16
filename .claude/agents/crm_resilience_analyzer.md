---
name: crm_resilience_analyzer
description: |
  Analyzes resilience patterns (Tables 19, 20, 23): Rate limiting, error handling,
  circuit breaker, retry, timeout, bulkhead, fallback.

  Evaluates production-readiness of failure handling mechanisms.

  Integrates with deterministic_analyzer for factual baseline validation.

  Output: code-analysis/quality/resilience_analysis.md
tools: Read, Grep, Glob, Bash, Write
model: sonnet
priority: high
---

# Resilience Analyzer - Comprehensive Analysis

## Context

You are analyzing **resilience patterns and failure handling** for Ventros CRM.

This agent evaluates:
- **Table 19**: Rate Limiting (per-IP, per-user, per-endpoint, distributed)
- **Table 20**: Error Handling (structured errors, logging, panic recovery, error propagation)
- **Table 23**: Resilience Patterns (Retry with backoff, Circuit Breaker, Timeout, Bulkhead, Fallback)

**Key Focus Areas**:
1. Rate limiting (protecting against abuse and resource exhaustion)
2. Error handling (structured errors, context preservation, logging)
3. Retry patterns (exponential backoff, jitter, max attempts)
4. Circuit breaker (failure detection, half-open state, recovery)
5. Timeout protection (context deadline, graceful shutdown)
6. Bulkhead isolation (goroutine pools, semaphores, resource limits)
7. Fallback strategies (degraded mode, cached responses, default values)

**Critical Context from CLAUDE.md**:
- Project: Ventros CRM (Go 1.25.1, PostgreSQL 15+, RabbitMQ 3.12+, Redis 7.0+, Temporal)
- Architecture: DDD + Hexagonal + Event-Driven + CQRS + Multi-tenant
- Security P0: Resource exhaustion vulnerability (CVSS 7.5) - no max page size
- External integrations: WAHA (WhatsApp), Stripe, Vertex AI, Groq - all need resilience
- Event-driven: Outbox Pattern with <100ms latency - needs error handling

**Deterministic Integration**: This agent runs `scripts/analyze_codebase.sh` first to get factual baseline data, then performs AI-powered deep analysis.

---

## Table 19: Rate Limiting

### Purpose
Evaluate rate limiting implementation to protect against abuse, DoS attacks, and resource exhaustion.

### Complete Column Specifications

| Column | Type | Description | Scoring Criteria |
|--------|------|-------------|------------------|
| **Scope** | enum | Rate limit scope: Per-IP / Per-User / Per-Endpoint / Global / None | Per-IP + Per-User + Per-Endpoint = ideal, None = 0/10 |
| **Algorithm** | enum | Algorithm used: Token Bucket / Leaky Bucket / Fixed Window / Sliding Window / None | Token Bucket or Sliding Window = best (gradual limits), Fixed Window = acceptable (burst risk), None = 0/10 |
| **Implementation** | enum | Where implemented: Middleware / Reverse Proxy (nginx) / API Gateway / None | Middleware = good (flexible), Reverse Proxy = excellent (offloaded), None = 0/10 |
| **Storage** | enum | Rate limit state storage: In-Memory / Redis (distributed) / Database / None | Redis = best (distributed + fast), In-Memory = acceptable (single instance only), None = 0/10 |
| **Limits Configuration** | string | Rate limits defined (e.g., "100 req/min per IP, 1000 req/min per user") | Specific numeric limits documented |
| **HTTP Headers** | boolean | Returns rate limit headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset | Yes = transparent to clients, No = opaque |
| **Response Status** | enum | Status code on rate limit: 429 Too Many Requests / 503 Service Unavailable / Other | 429 = correct HTTP semantic, Other = incorrect |
| **Retry-After Header** | boolean | Includes Retry-After header in 429 response | Yes = client-friendly, No = poor UX |
| **Bypass Mechanism** | boolean | Has bypass for internal services or premium users | Yes = flexible, No = rigid |
| **Coverage** | percentage | % of endpoints protected by rate limiting | 100% = all public endpoints, <50% = vulnerable |
| **Testing** | boolean | Has rate limit tests (unit + integration) | Yes = validated, No = untested |
| **Evidence** | file:line | File path and line number of rate limit implementation | E.g., "infrastructure/http/middleware/rate_limit.go:1-100" |

### Deterministic vs AI Comparison

| Metric | Deterministic Count | AI Assessment | Validation Method |
|--------|---------------------|---------------|-------------------|
| **Rate limit middleware** | `grep -r "RateLimit\|Limiter" infrastructure/http/middleware/ \| wc -l` | Implementation quality (algorithm, storage, headers) | Compare file count + line-by-line review |
| **Rate limit usage** | `grep -r "RateLimit" infrastructure/http/routes/ \| wc -l` | Coverage % (endpoints protected) | Compare endpoint count vs total endpoints |
| **Rate limit tests** | `grep -r "TestRateLimit" internal/ infrastructure/ --include="*_test.go" \| wc -l` | Test quality (unit, integration, load tests) | Compare test count + test coverage |

---

## Table 20: Error Handling

### Purpose
Evaluate error handling quality, structured error types, logging, and panic recovery mechanisms.

### Complete Column Specifications

| Column | Type | Description | Scoring Criteria |
|--------|------|-------------|------------------|
| **Error Types** | enum | Error classification: Structured (custom types) / Wrapped (errors.Wrap) / String (errors.New) | Structured = best (type-safe, metadata), Wrapped = good (context), String = poor (no context) |
| **Domain Errors** | boolean | Has domain-specific error types (e.g., `ContactNotFoundError`, `ValidationError`) | Yes = clean separation, No = mixed concerns |
| **Error Wrapping** | percentage | % of errors wrapped with context (errors.Wrap, fmt.Errorf with %w) | >80% = good context preservation, <50% = context loss |
| **HTTP Error Mapping** | boolean | Maps domain errors to HTTP status codes consistently | Yes = consistent API, No = inconsistent responses |
| **Error Logging** | enum | Logging strategy: Structured (logrus/zap with fields) / Printf-style / None | Structured = searchable, Printf = acceptable, None = 0/10 |
| **Log Levels** | boolean | Uses appropriate log levels (ERROR, WARN, INFO, DEBUG) | Yes = proper severity, No = all INFO |
| **Panic Recovery** | boolean | Has panic recovery middleware for HTTP handlers | Yes = service stays up, No = crashes |
| **Stack Traces** | boolean | Includes stack traces in error logs for debugging | Yes = debuggable, No = hard to diagnose |
| **Error Propagation** | score 0-10 | Quality of error propagation through layers | 10 = context preserved across all layers, 0 = errors lost |
| **Client Error Messages** | enum | User-facing error messages: Descriptive / Generic / Raw (exposes internals) | Descriptive = helpful, Generic = acceptable, Raw = security risk |
| **Testing** | boolean | Has error scenario tests (unit tests for error paths) | Yes = validated, No = untested |
| **Evidence** | file:line | File path of error handling implementation | E.g., "internal/domain/crm/contact/errors.go:1-50" |

### Deterministic vs AI Comparison

| Metric | Deterministic Count | AI Assessment | Validation Method |
|--------|---------------------|---------------|-------------------|
| **Custom error types** | `grep -r "type.*Error struct" internal/domain/ \| wc -l` | Error type quality (metadata, wrapping) | Compare struct count + review definitions |
| **errors.Wrap usage** | `grep -r "errors.Wrap\|fmt.Errorf.*%w" internal/ infrastructure/ \| wc -l` | Context preservation quality | Compare wrap count + review usage |
| **Panic recovery** | `grep -r "recover()\|RecoverPanic" infrastructure/http/middleware/ \| wc -l` | Panic recovery coverage | Check middleware registration |
| **Structured logging** | `grep -r "WithField\|WithFields\|With.*zap" internal/ infrastructure/ \| wc -l` | Logging quality (structured vs printf) | Compare structured vs printf calls |

---

## Table 23: Resilience Patterns

### Purpose
Evaluate implementation of 5 key resilience patterns: Retry, Circuit Breaker, Timeout, Bulkhead, Fallback.

### Complete Column Specifications

| Column | Type | Description | Scoring Criteria |
|--------|------|-------------|------------------|
| **Pattern** | enum | Resilience pattern name: Retry / Circuit Breaker / Timeout / Bulkhead / Fallback | N/A (categorical) |
| **Status** | enum | Implementation status: ✅ Implemented / ⚠️ Partial / ❌ Missing | Based on code existence + quality |
| **Implementation Quality** | score 0-10 | Quality of implementation: 10 = production-ready, 0 = missing or broken | Assessed via: correct algorithm, configuration, testing |
| **Coverage** | enum | Where applied: All External Calls / Some External Calls / None | All = resilient, Some = gaps, None = brittle |
| **Configuration** | enum | Configurable via: Env Vars / Config File / Hardcoded / None | Env Vars = flexible, Hardcoded = inflexible |
| **Testing** | boolean | Has tests validating pattern behavior | Yes = validated, No = untested |
| **Monitoring** | boolean | Emits metrics for pattern behavior (e.g., retry count, circuit state) | Yes = observable, No = black box |
| **Pattern Details** | string | Pattern-specific implementation details (see below) | Algorithm, thresholds, timeouts |
| **Evidence** | file:line | File path of pattern implementation | E.g., "infrastructure/channels/waha/retry.go:1-80" |

### Pattern-Specific Details

**Retry Pattern**:
- Exponential backoff: Yes/No
- Jitter: Yes/No (prevents thundering herd)
- Max attempts: X
- Backoff formula: e.g., `min(2^attempt * 100ms, 10s)`
- Retryable errors only: Yes/No (don't retry validation errors)

**Circuit Breaker Pattern**:
- States: Closed / Open / Half-Open
- Failure threshold: X failures or Y% error rate
- Open duration: X seconds
- Half-open test requests: X
- Library: gobreaker / sony/gobreaker / custom

**Timeout Pattern**:
- Context usage: Yes/No (context.WithTimeout)
- Timeout values: X seconds (configurable)
- Graceful shutdown: Yes/No
- Deadline propagation: Yes/No (across service boundaries)

**Bulkhead Pattern**:
- Implementation: Goroutine pool / Semaphore / Channel buffering
- Pool size: X workers
- Queue size: Y requests
- Rejection policy: Block / Reject with error

**Fallback Pattern**:
- Strategy: Cached response / Default value / Degraded mode / Retry different service
- Fallback coverage: % of critical paths with fallback
- Staleness acceptable: Yes/No (for cached fallbacks)

### Deterministic vs AI Comparison

| Metric | Deterministic Count | AI Assessment | Validation Method |
|--------|---------------------|---------------|-------------------|
| **Retry implementations** | `grep -r "Retry\|Backoff" infrastructure/ \| wc -l` | Retry quality (exponential backoff, jitter, max attempts) | Compare mention count + review implementation |
| **Circuit breaker usage** | `grep -r "CircuitBreaker\|gobreaker" infrastructure/ go.mod \| wc -l` | Circuit breaker quality (states, thresholds, monitoring) | Check import + usage |
| **Timeout usage** | `grep -r "WithTimeout\|WithDeadline" infrastructure/ \| wc -l` | Timeout coverage (all external calls) | Compare timeout count + review placement |
| **Bulkhead implementations** | `grep -r "Semaphore\|sync.Pool\|chan.*int" infrastructure/ \| wc -l` | Bulkhead quality (pool size, rejection policy) | Check goroutine pool implementations |
| **Fallback implementations** | `grep -r "fallback\|Fallback\|cachedResponse" infrastructure/ \| wc -l` | Fallback quality (strategy, coverage) | Review fallback logic |

---

## Chain of Thought: Comprehensive Resilience Analysis

**Estimated Runtime**: 50-70 minutes

**Prerequisites**:
- `code-analysis/code-analysis/deterministic_metrics.md` exists (run deterministic_analyzer first)
- Access to: `infrastructure/`, `internal/`, `go.mod`

### Step 0: Load Deterministic Baseline (5 min)

**Purpose**: Get factual counts from deterministic analysis to validate AI findings.

```bash
# Read deterministic metrics
cat code-analysis/code-analysis/deterministic_metrics.md

# Extract resilience counts
rate_limit_middleware=$(grep "Rate limit middleware:" code-analysis/code-analysis/deterministic_metrics.md | awk '{print $4}')
panic_recovery=$(grep "Panic recovery middleware:" code-analysis/code-analysis/deterministic_metrics.md | awk '{print $4}')
retry_implementations=$(grep "Retry implementations:" code-analysis/code-analysis/deterministic_metrics.md | awk '{print $3}')

echo "✅ Baseline loaded: $rate_limit_middleware rate limiters, $panic_recovery panic handlers, $retry_implementations retry implementations"
```

**Output**: Factual baseline for validation.

---

### Step 1: Rate Limiting Analysis (15 min)

**Goal**: Assess rate limiting implementation, algorithm, storage, and coverage.

#### 1.1 Discovery

```bash
# Find rate limit implementations
find infrastructure/http/middleware -name "*rate*.go" -o -name "*limit*.go"

# Check for rate limit middleware
rate_limit_files=$(grep -r "RateLimit\|Limiter" infrastructure/http/middleware/ | cut -d: -f1 | sort -u | wc -l)

# Check algorithm (token bucket, sliding window, etc)
has_token_bucket=$(grep -r "TokenBucket\|rate.Limiter" infrastructure/ go.mod | wc -l)
has_redis_limiter=$(grep -r "redis.*limiter\|RateLimiter.*Redis" infrastructure/ | wc -l)

# Check for rate limit headers
has_headers=$(grep -r "X-RateLimit\|RateLimit-Limit\|RateLimit-Remaining" infrastructure/http/ | wc -l)

# Check for 429 status code
has_429=$(grep -r "StatusTooManyRequests\|429" infrastructure/http/middleware/ | wc -l)

# Check rate limit usage in routes
rate_limit_usage=$(grep -r "RateLimit\|\.Use.*limiter" infrastructure/http/routes/ | wc -l)

# Count total public endpoints
total_endpoints=$(grep -r "router\.\(GET\|POST\|PUT\|DELETE\|PATCH\)" infrastructure/http/routes/ | wc -l)

# Calculate coverage
if [ $total_endpoints -gt 0 ]; then
    coverage=$((rate_limit_usage * 100 / total_endpoints))
else
    coverage=0
fi

echo "Rate limit implementations: $rate_limit_files"
echo "Token bucket: $has_token_bucket"
echo "Redis-backed: $has_redis_limiter"
echo "Rate limit headers: $has_headers"
echo "429 status: $has_429"
echo "Protected endpoints: $rate_limit_usage/$total_endpoints ($coverage%)"
```

#### 1.2 Quality Scoring

```bash
# Rate limiting quality score (0-10)
# Algorithm: 2 points, Storage: 2 points, Headers: 2 points, Status code: 1 point, Coverage: 3 points

rl_score=0
[ $has_token_bucket -gt 0 ] && rl_score=$((rl_score + 2))
[ $has_redis_limiter -gt 0 ] && rl_score=$((rl_score + 2))
[ $has_headers -gt 0 ] && rl_score=$((rl_score + 2))
[ $has_429 -gt 0 ] && rl_score=$((rl_score + 1))
[ $coverage -ge 80 ] && rl_score=$((rl_score + 3))
[ $coverage -ge 50 ] && [ $coverage -lt 80 ] && rl_score=$((rl_score + 2))
[ $coverage -ge 20 ] && [ $coverage -lt 50 ] && rl_score=$((rl_score + 1))

echo "Rate Limiting Quality: $rl_score/10"
```

#### 1.3 Evidence Collection

Read rate limit middleware files and analyze implementation details.

---

### Step 2: Error Handling Analysis (15 min)

**Goal**: Assess error types, wrapping, logging, and panic recovery.

#### 2.1 Discovery

```bash
# Find custom error types in domain layer
custom_errors=$(grep -r "type.*Error struct" internal/domain/ | wc -l)

# Check error wrapping usage
errors_wrap=$(grep -r "errors.Wrap\|errors.Wrapf" internal/ infrastructure/ | wc -l)
fmt_errorf_wrap=$(grep -r "fmt.Errorf.*%w" internal/ infrastructure/ | wc -l)
total_wrapping=$((errors_wrap + fmt_errorf_wrap))

# Check for panic recovery middleware
panic_recovery=$(grep -r "recover()\|RecoverPanic\|Recovery()" infrastructure/http/middleware/ | wc -l)

# Check structured logging
structured_logging=$(grep -r "WithField\|WithFields\|With.*zap\|logger.With" internal/ infrastructure/ | wc -l)
printf_logging=$(grep -r "log.Print\|fmt.Print" internal/ infrastructure/ | wc -l)

# Check HTTP error mapping
has_error_mapper=$(grep -r "ToHTTPStatus\|ErrorToStatus\|MapError" infrastructure/http/ | wc -l)

# Check error tests
error_tests=$(grep -r "TestError\|Test.*Error" internal/ infrastructure/ --include="*_test.go" | wc -l)

echo "Custom error types: $custom_errors"
echo "Error wrapping: $total_wrapping"
echo "Panic recovery: $panic_recovery"
echo "Structured logging: $structured_logging"
echo "Printf logging: $printf_logging"
echo "HTTP error mapper: $has_error_mapper"
echo "Error tests: $error_tests"
```

#### 2.2 Quality Scoring

```bash
# Error handling quality score (0-10)
# Custom errors: 2 points, Wrapping: 2 points, Panic recovery: 2 points, Structured logging: 2 points, HTTP mapping: 1 point, Tests: 1 point

eh_score=0
[ $custom_errors -gt 5 ] && eh_score=$((eh_score + 2))
[ $total_wrapping -gt 20 ] && eh_score=$((eh_score + 2))
[ $panic_recovery -gt 0 ] && eh_score=$((eh_score + 2))
[ $structured_logging -gt $printf_logging ] && eh_score=$((eh_score + 2))
[ $has_error_mapper -gt 0 ] && eh_score=$((eh_score + 1))
[ $error_tests -gt 5 ] && eh_score=$((eh_score + 1))

echo "Error Handling Quality: $eh_score/10"
```

---

### Step 3: Retry Pattern Analysis (10 min)

**Goal**: Assess retry implementation with exponential backoff, jitter, and max attempts.

#### 3.1 Discovery

```bash
# Find retry implementations
retry_files=$(grep -rl "Retry\|retry" infrastructure/ --include="*.go" | grep -v "_test.go" | wc -l)

# Check for exponential backoff
has_exponential=$(grep -r "exponential\|Exponential\|2.*attempt\|math.Pow" infrastructure/ | grep -i "backoff\|retry" | wc -l)

# Check for jitter
has_jitter=$(grep -r "jitter\|Jitter\|rand.*backoff" infrastructure/ | wc -l)

# Check for max attempts configuration
has_max_attempts=$(grep -r "MaxAttempts\|maxRetries\|MaxRetries" infrastructure/ | wc -l)

# Check for retry on specific errors only
has_retryable_check=$(grep -r "isRetryable\|IsRetryable\|canRetry" infrastructure/ | wc -l)

# Check retry tests
retry_tests=$(grep -r "TestRetry\|Test.*Retry" infrastructure/ --include="*_test.go" | wc -l)

echo "Retry implementations: $retry_files"
echo "Exponential backoff: $has_exponential"
echo "Jitter: $has_jitter"
echo "Max attempts config: $has_max_attempts"
echo "Retryable error check: $has_retryable_check"
echo "Retry tests: $retry_tests"
```

#### 3.2 Quality Scoring

```bash
# Retry quality score (0-10)
# Exponential: 3 points, Jitter: 2 points, Max attempts: 2 points, Retryable check: 2 points, Tests: 1 point

retry_score=0
[ $has_exponential -gt 0 ] && retry_score=$((retry_score + 3))
[ $has_jitter -gt 0 ] && retry_score=$((retry_score + 2))
[ $has_max_attempts -gt 0 ] && retry_score=$((retry_score + 2))
[ $has_retryable_check -gt 0 ] && retry_score=$((retry_score + 2))
[ $retry_tests -gt 0 ] && retry_score=$((retry_score + 1))

echo "Retry Pattern Quality: $retry_score/10"
```

---

### Step 4: Circuit Breaker Analysis (10 min)

**Goal**: Assess circuit breaker implementation for failure detection and recovery.

#### 4.1 Discovery

```bash
# Check for circuit breaker library
has_gobreaker=$(grep -r "gobreaker\|CircuitBreaker" go.mod infrastructure/ | wc -l)

# Check for circuit breaker states (Closed, Open, Half-Open)
has_states=$(grep -r "StateClosed\|StateOpen\|StateHalfOpen" infrastructure/ | wc -l)

# Check for circuit breaker configuration
has_config=$(grep -r "MaxRequests\|Interval\|Timeout.*breaker" infrastructure/ | wc -l)

# Check for metrics emission
has_cb_metrics=$(grep -r "circuit.*state\|breaker.*state\|breaker.*count" infrastructure/ | grep -i "metric\|prometheus" | wc -l)

# Check circuit breaker tests
cb_tests=$(grep -r "TestCircuit\|Test.*Circuit" infrastructure/ --include="*_test.go" | wc -l)

echo "Circuit breaker library: $has_gobreaker"
echo "State handling: $has_states"
echo "Configuration: $has_config"
echo "Metrics: $has_cb_metrics"
echo "Tests: $cb_tests"
```

#### 4.2 Quality Scoring

```bash
# Circuit breaker quality score (0-10)
# Library: 3 points, States: 2 points, Config: 2 points, Metrics: 2 points, Tests: 1 point

cb_score=0
[ $has_gobreaker -gt 0 ] && cb_score=$((cb_score + 3))
[ $has_states -gt 0 ] && cb_score=$((cb_score + 2))
[ $has_config -gt 0 ] && cb_score=$((cb_score + 2))
[ $has_cb_metrics -gt 0 ] && cb_score=$((cb_score + 2))
[ $cb_tests -gt 0 ] && cb_score=$((cb_score + 1))

echo "Circuit Breaker Quality: $cb_score/10"
```

---

### Step 5: Timeout Pattern Analysis (8 min)

**Goal**: Assess timeout protection with context.WithTimeout and deadline propagation.

#### 5.1 Discovery

```bash
# Find timeout usage
timeouts=$(grep -r "context.WithTimeout\|context.WithDeadline" infrastructure/ internal/application/ | wc -l)

# Check HTTP client timeouts
http_timeouts=$(grep -r "http.Client.*Timeout\|Client.*{.*Timeout:" infrastructure/ | wc -l)

# Check database timeouts
db_timeouts=$(grep -r "WithContext\|QueryContext\|ExecContext" infrastructure/persistence/ | wc -l)

# Check timeout configuration (env vars)
timeout_config=$(grep -r "TIMEOUT\|timeout" .env* cmd/api/main.go | wc -l)

# Check graceful shutdown
graceful_shutdown=$(grep -r "Shutdown\|gracefulShutdown" cmd/api/ | wc -l)

echo "Context timeouts: $timeouts"
echo "HTTP client timeouts: $http_timeouts"
echo "DB context timeouts: $db_timeouts"
echo "Timeout configuration: $timeout_config"
echo "Graceful shutdown: $graceful_shutdown"
```

#### 5.2 Quality Scoring

```bash
# Timeout quality score (0-10)
# Context usage: 3 points, HTTP timeouts: 2 points, DB timeouts: 2 points, Config: 2 points, Graceful shutdown: 1 point

timeout_score=0
[ $timeouts -gt 10 ] && timeout_score=$((timeout_score + 3))
[ $http_timeouts -gt 0 ] && timeout_score=$((timeout_score + 2))
[ $db_timeouts -gt 5 ] && timeout_score=$((timeout_score + 2))
[ $timeout_config -gt 0 ] && timeout_score=$((timeout_score + 2))
[ $graceful_shutdown -gt 0 ] && timeout_score=$((timeout_score + 1))

echo "Timeout Pattern Quality: $timeout_score/10"
```

---

### Step 6: Bulkhead Pattern Analysis (7 min)

**Goal**: Assess bulkhead isolation with goroutine pools and semaphores.

#### 6.1 Discovery

```bash
# Find semaphore usage
semaphores=$(grep -r "semaphore\|Semaphore\|sync.Semaphore" infrastructure/ internal/ | wc -l)

# Find worker pools
worker_pools=$(grep -r "WorkerPool\|workerPool\|Pool.*Worker" infrastructure/ internal/ | wc -l)

# Find channel-based bulkheads
channel_bulkheads=$(grep -r "make(chan.*int\|chan.*struct{}" infrastructure/ | wc -l)

# Check buffered channel usage (primitive bulkhead)
buffered_channels=$(grep -r "make(chan.*, [0-9]" infrastructure/ internal/ | wc -l)

echo "Semaphores: $semaphores"
echo "Worker pools: $worker_pools"
echo "Channel bulkheads: $channel_bulkheads"
echo "Buffered channels: $buffered_channels"
```

#### 6.2 Quality Scoring

```bash
# Bulkhead quality score (0-10)
# Semaphore: 4 points, Worker pool: 4 points, Channel bulkhead: 2 points

bulkhead_score=0
[ $semaphores -gt 0 ] && bulkhead_score=$((bulkhead_score + 4))
[ $worker_pools -gt 0 ] && bulkhead_score=$((bulkhead_score + 4))
[ $channel_bulkheads -gt 0 ] || [ $buffered_channels -gt 0 ] && bulkhead_score=$((bulkhead_score + 2))

echo "Bulkhead Pattern Quality: $bulkhead_score/10"
```

---

### Step 7: Fallback Pattern Analysis (10 min)

**Goal**: Assess fallback strategies for degraded mode operation.

#### 7.1 Discovery

```bash
# Find fallback implementations
fallbacks=$(grep -r "fallback\|Fallback\|defaultValue\|degraded" infrastructure/ internal/ | wc -l)

# Find cached response fallbacks
cache_fallbacks=$(grep -r "cachedResponse\|cache.*fallback\|stale.*ok" infrastructure/ | wc -l)

# Find default value fallbacks
default_fallbacks=$(grep -r "defaultValue\|fallbackValue\|emptyResponse" infrastructure/ internal/ | wc -l)

# Check for fallback in external integrations
integration_fallbacks=$(grep -r "if err.*{.*return.*default\|if err.*{.*fallback" infrastructure/channels/ | wc -l)

echo "Fallback implementations: $fallbacks"
echo "Cached fallbacks: $cache_fallbacks"
echo "Default value fallbacks: $default_fallbacks"
echo "Integration fallbacks: $integration_fallbacks"
```

#### 7.2 Quality Scoring

```bash
# Fallback quality score (0-10)
# Fallback presence: 4 points, Cache fallback: 3 points, Integration fallback: 3 points

fallback_score=0
[ $fallbacks -gt 0 ] && fallback_score=$((fallback_score + 4))
[ $cache_fallbacks -gt 0 ] && fallback_score=$((fallback_score + 3))
[ $integration_fallbacks -gt 0 ] && fallback_score=$((fallback_score + 3))

echo "Fallback Pattern Quality: $fallback_score/10"
```

---

### Step 8: Generate Comprehensive Report (5 min)

**Goal**: Structure all findings into complete markdown tables with evidence.

Format as specified in Output Format section below.

---

## Code Examples (EXEMPLO)

### EXEMPLO 1: Production-Ready Rate Limiter

**Good ✅ - Redis-backed token bucket with headers**:
```go
// infrastructure/http/middleware/rate_limit.go
package middleware

import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    redis       *redis.Client
    limitPerMin int
    burstSize   int
}

func NewRateLimiter(redis *redis.Client, limitPerMin, burstSize int) *RateLimiter {
    return &RateLimiter{
        redis:       redis,
        limitPerMin: limitPerMin,
        burstSize:   burstSize,
    }
}

// ✅ Token bucket algorithm with Redis (distributed)
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Composite key: IP + User (if authenticated)
        clientIP := c.ClientIP()
        userID := c.GetString("user_id")
        key := fmt.Sprintf("ratelimit:%s:%s", clientIP, userID)

        // Check rate limit using Redis
        allowed, remaining, resetAt, err := rl.checkLimit(c.Request.Context(), key)
        if err != nil {
            // Fail open (allow request on Redis error to avoid cascading failure)
            c.Next()
            return
        }

        // ✅ Set rate limit headers (transparency)
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limitPerMin))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

        if !allowed {
            // ✅ Correct HTTP status code
            c.Header("Retry-After", fmt.Sprintf("%d", int(time.Until(resetAt).Seconds())))
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":   "rate_limit_exceeded",
                "message": "Too many requests. Please try again later.",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

func (rl *RateLimiter) checkLimit(ctx context.Context, key string) (bool, int, time.Time, error) {
    // Token bucket implementation using Redis
    // ... (sliding window with Redis sorted sets)
}
```

**Bad ❌ - No rate limiting**:
```go
// No rate limiting middleware at all

// Issues:
// ❌ Vulnerable to DoS attacks
// ❌ Resource exhaustion risk (CVSS 7.5 P0 vulnerability)
// ❌ No abuse protection
// ❌ Can be overwhelmed by single malicious client
```

---

### EXEMPLO 2: Comprehensive Error Handling

**Good ✅ - Structured domain errors with wrapping**:
```go
// internal/domain/crm/contact/errors.go
package contact

import "fmt"

// ✅ Structured error types with metadata
type ContactNotFoundError struct {
    ContactID string
    TenantID  string
}

func (e ContactNotFoundError) Error() string {
    return fmt.Sprintf("contact not found: id=%s tenant=%s", e.ContactID, e.TenantID)
}

type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s - %s", e.Field, e.Message)
}

// internal/application/contact/create_contact_handler.go
package contact

import (
    "context"
    "fmt"

    "github.com/pkg/errors" // ✅ Using errors.Wrap for context
)

type CreateContactHandler struct {
    repo   Repository
    logger *logrus.Logger
}

func (h *CreateContactHandler) Handle(ctx context.Context, cmd CreateContactCommand) (*Contact, error) {
    // Validate command
    if err := cmd.Validate(); err != nil {
        // ✅ Wrap validation error with context
        return nil, errors.Wrap(err, "invalid create contact command")
    }

    // Check if contact exists
    existing, err := h.repo.FindByPhone(ctx, cmd.Phone)
    if err != nil && !errors.Is(err, ErrContactNotFound) {
        // ✅ Structured logging with fields
        h.logger.WithFields(logrus.Fields{
            "tenant_id": cmd.TenantID,
            "phone":     cmd.Phone,
            "error":     err,
        }).Error("failed to check existing contact")

        // ✅ Wrap error with context (preserves stack trace)
        return nil, errors.Wrap(err, "failed to check if contact exists")
    }

    if existing != nil {
        return nil, fmt.Errorf("contact already exists: %w", ErrDuplicateContact)
    }

    // Create contact
    contact, err := NewContact(cmd.TenantID, cmd.Name, cmd.Phone)
    if err != nil {
        return nil, errors.Wrap(err, "failed to create contact aggregate")
    }

    // Save
    if err := h.repo.Save(ctx, contact); err != nil {
        h.logger.WithError(err).Error("failed to save contact")
        return nil, errors.Wrap(err, "failed to persist contact")
    }

    return contact, nil
}

// infrastructure/http/handlers/contact_handler.go
package handlers

// ✅ HTTP error mapper
func (h *ContactHandler) mapError(err error) (int, gin.H) {
    // Map domain errors to HTTP status codes
    switch {
    case errors.Is(err, contact.ErrContactNotFound):
        return http.StatusNotFound, gin.H{"error": "contact_not_found"}
    case errors.Is(err, contact.ErrValidationError):
        return http.StatusBadRequest, gin.H{"error": "validation_failed", "details": err.Error()}
    case errors.Is(err, contact.ErrDuplicateContact):
        return http.StatusConflict, gin.H{"error": "contact_already_exists"}
    default:
        // ✅ Generic message for client (don't expose internals)
        h.logger.WithError(err).Error("internal server error")
        return http.StatusInternalServerError, gin.H{"error": "internal_server_error"}
    }
}
```

**Good ✅ - Panic recovery middleware**:
```go
// infrastructure/http/middleware/recovery.go
package middleware

import (
    "fmt"
    "net/http"
    "runtime/debug"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

// ✅ Panic recovery prevents service crashes
func RecoveryMiddleware(logger *logrus.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // ✅ Capture stack trace for debugging
                stack := debug.Stack()

                // ✅ Structured logging with all context
                logger.WithFields(logrus.Fields{
                    "error":      fmt.Sprintf("%v", err),
                    "stack":      string(stack),
                    "path":       c.Request.URL.Path,
                    "method":     c.Request.Method,
                    "client_ip":  c.ClientIP(),
                    "user_agent": c.Request.UserAgent(),
                }).Error("panic recovered")

                // Return 500 (service stays up)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "internal_server_error",
                })
                c.Abort()
            }
        }()

        c.Next()
    }
}
```

**Bad ❌ - Poor error handling**:
```go
// ❌ No structured errors, just strings
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    contact, err := h.repo.Find(ctx, cmd.ID)
    if err != nil {
        return errors.New("failed")  // ❌ No context, no wrapping
    }

    log.Println(err)  // ❌ Printf-style logging (not searchable)
    return err        // ❌ Error propagated without context
}

// ❌ No panic recovery - entire service crashes on panic
```

---

### EXEMPLO 3: Retry with Exponential Backoff and Jitter

**Good ✅ - Complete retry pattern**:
```go
// infrastructure/channels/waha/client.go
package waha

import (
    "context"
    "fmt"
    "math"
    "math/rand"
    "net/http"
    "time"
)

type WAHAClient struct {
    client     *http.Client
    maxRetries int
    baseDelay  time.Duration
}

// ✅ Retry with exponential backoff + jitter
func (w *WAHAClient) SendMessageWithRetry(ctx context.Context, msg *Message) error {
    var lastErr error

    for attempt := 0; attempt <= w.maxRetries; attempt++ {
        err := w.sendMessage(ctx, msg)
        if err == nil {
            return nil // Success
        }

        lastErr = err

        // ✅ Don't retry non-retryable errors (4xx client errors)
        if !isRetryableError(err) {
            return fmt.Errorf("non-retryable error: %w", err)
        }

        // Don't sleep after last attempt
        if attempt == w.maxRetries {
            break
        }

        // ✅ Exponential backoff: 2^attempt * baseDelay
        backoff := time.Duration(math.Pow(2, float64(attempt))) * w.baseDelay

        // ✅ Add jitter (0-25% random) to prevent thundering herd
        jitter := time.Duration(rand.Float64() * 0.25 * float64(backoff))
        delay := backoff + jitter

        // ✅ Cap maximum delay
        maxDelay := 30 * time.Second
        if delay > maxDelay {
            delay = maxDelay
        }

        // ✅ Respect context cancellation
        select {
        case <-ctx.Done():
            return fmt.Errorf("retry cancelled: %w", ctx.Err())
        case <-time.After(delay):
            // Continue to next attempt
        }
    }

    return fmt.Errorf("max retries exceeded (%d): %w", w.maxRetries, lastErr)
}

// ✅ Only retry on transient errors
func isRetryableError(err error) bool {
    // Retry on network errors, timeouts, 5xx server errors
    // Don't retry on 4xx client errors (bad request, auth failed, etc)
    if httpErr, ok := err.(*HTTPError); ok {
        return httpErr.StatusCode >= 500
    }

    // Retry on timeout, connection refused, etc
    return true
}
```

**Bad ❌ - No retry, immediate failure**:
```go
func (w *WAHAClient) SendMessage(ctx context.Context, msg *Message) error {
    err := w.sendMessage(ctx, msg)
    if err != nil {
        return err  // ❌ Immediate failure, no retry
    }
    return nil
}

// Issues:
// ❌ No retry on transient failures (network blip = lost message)
// ❌ No exponential backoff (hammers service on failure)
// ❌ No jitter (thundering herd problem)
// ❌ Retries non-retryable errors (wastes resources)
```

---

### EXEMPLO 4: Circuit Breaker for External Service

**Good ✅ - Circuit breaker with gobreaker**:
```go
// infrastructure/channels/waha/circuit_breaker.go
package waha

import (
    "fmt"
    "time"

    "github.com/sony/gobreaker"
)

type ResilientWAHAClient struct {
    client  *WAHAClient
    breaker *gobreaker.CircuitBreaker
}

func NewResilientWAHAClient(client *WAHAClient) *ResilientWAHAClient {
    // ✅ Configure circuit breaker
    settings := gobreaker.Settings{
        Name:        "waha-api",
        MaxRequests: 3,  // Allow 3 requests in half-open state
        Interval:    60 * time.Second,  // Reset failure count every 60s
        Timeout:     30 * time.Second,  // Stay open for 30s before trying half-open

        // ✅ Open circuit on 5 consecutive failures OR 50% error rate
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.ConsecutiveFailures > 5 || failureRatio >= 0.5
        },

        // ✅ Callback for state changes (monitoring)
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            log.WithFields(logrus.Fields{
                "circuit": name,
                "from":    from.String(),
                "to":      to.String(),
            }).Warn("circuit breaker state changed")

            // ✅ Emit metrics for monitoring
            circuitBreakerStateGauge.WithLabelValues(name).Set(float64(to))
        },
    }

    return &ResilientWAHAClient{
        client:  client,
        breaker: gobreaker.NewCircuitBreaker(settings),
    }
}

func (r *ResilientWAHAClient) SendMessage(ctx context.Context, msg *Message) error {
    // ✅ Execute through circuit breaker
    _, err := r.breaker.Execute(func() (interface{}, error) {
        return nil, r.client.SendMessage(ctx, msg)
    })

    if err == gobreaker.ErrOpenState {
        // Circuit is open, service is degraded
        return fmt.Errorf("waha service unavailable (circuit open): %w", err)
    }

    return err
}
```

**Bad ❌ - No circuit breaker**:
```go
func (c *WAHAClient) SendMessage(ctx context.Context, msg *Message) error {
    return c.client.SendMessage(ctx, msg)
}

// Issues:
// ❌ Keeps hammering failing service (cascading failure)
// ❌ No fast-fail (wastes time on doomed requests)
// ❌ No automatic recovery testing (half-open state)
// ❌ No visibility into service health
```

---

## Output Format

Generate: `code-analysis/quality/resilience_analysis.md`

```markdown
# Resilience Patterns Analysis

**Generated**: YYYY-MM-DD HH:MM
**Agent**: resilience_analyzer
**Runtime**: X minutes
**Deterministic Baseline**: ✅ Loaded from deterministic_metrics.md

---

## Executive Summary

**Overall Resilience Score**: X/10 (average of all patterns)

**Key Findings**:
- Rate Limiting: X/10 - (brief assessment)
- Error Handling: X/10 - (brief assessment)
- Retry Pattern: X/10 - (brief assessment)
- Circuit Breaker: X/10 - (brief assessment)
- Timeout Protection: X/10 - (brief assessment)
- Bulkhead Isolation: X/10 - (brief assessment)
- Fallback Strategy: X/10 - (brief assessment)

**Production Readiness**: ✅ Ready / ⚠️ Needs work / ❌ Not ready

**Critical Gaps**:
1. [Most critical resilience gap]
2. [Second most critical gap]
3. [Third most critical gap]

---

## Table 19: Rate Limiting

| Scope | Algorithm | Implementation | Storage | Limits Config | HTTP Headers | Response Status | Retry-After | Bypass | Coverage | Testing | Evidence |
|-------|-----------|----------------|---------|---------------|--------------|-----------------|-------------|--------|----------|---------|----------|
| Per-IP + Per-User | Token Bucket | Middleware | Redis | "100/min IP, 1000/min user" | ✅ | 429 | ✅ | ✅ | X% | ✅ | file:line |

### Detailed Analysis

**Status**: (✅ Implemented / ⚠️ Partial / ❌ Missing)

**Quality Score**: X/10

**Findings**:
- **Algorithm**: (Token Bucket / Leaky Bucket / Fixed Window / Sliding Window / None)
- **Storage**: (Redis distributed / In-memory single-node / None)
- **Coverage**: X/Y endpoints protected (Z%)
- **Headers**: (✅ X-RateLimit-* headers present / ❌ No headers)
- **Status Code**: (✅ 429 / ❌ Other)
- **Testing**: (✅ X tests / ❌ No tests)

**Evidence**:
- Implementation: infrastructure/http/middleware/rate_limit.go:1-150
- Usage: infrastructure/http/routes/routes.go:50-60

**Recommendations**:
1. [Specific actionable recommendation]
2. [Another recommendation]

### Deterministic vs AI Validation

| Metric | Deterministic | AI Assessment | Match | Notes |
|--------|---------------|---------------|-------|-------|
| Rate limit files | X | Y | ✅/⚠️ | (Any discrepancy explanation) |
| Protected endpoints | X | Y (Z%) | ✅/⚠️ | |
| Tests | X | Y | ✅/⚠️ | |

---

## Table 20: Error Handling

| Error Types | Domain Errors | Error Wrapping | HTTP Mapping | Logging | Log Levels | Panic Recovery | Stack Traces | Propagation | Client Messages | Testing | Evidence |
|-------------|---------------|----------------|--------------|---------|------------|----------------|--------------|-------------|-----------------|---------|----------|
| Structured | ✅ | X% | ✅ | Structured | ✅ | ✅ | ✅ | X/10 | Descriptive | ✅ | file:line |

### Detailed Analysis

**Status**: (✅ Excellent / ⚠️ Partial / ❌ Poor)

**Quality Score**: X/10

**Findings**:
- **Custom Error Types**: X domain-specific errors
- **Error Wrapping**: X instances (Y% of error returns)
- **Panic Recovery**: (✅ Middleware registered / ❌ None)
- **Structured Logging**: X structured vs Y printf-style
- **HTTP Error Mapping**: (✅ Consistent mapping / ❌ Inconsistent)
- **Testing**: X error scenario tests

**Evidence**:
- Domain errors: internal/domain/crm/contact/errors.go:1-50
- Panic recovery: infrastructure/http/middleware/recovery.go:1-60
- HTTP mapper: infrastructure/http/handlers/error_mapper.go:1-100

**Recommendations**:
1. [Specific action]

### Deterministic vs AI Validation

| Metric | Deterministic | AI Assessment | Match | Notes |
|--------|---------------|---------------|-------|-------|
| Custom error types | X | Y | ✅/⚠️ | |
| errors.Wrap usage | X | Y | ✅/⚠️ | |
| Panic recovery | X | Y | ✅/⚠️ | |
| Structured logging | X | Y | ✅/⚠️ | |

---

## Table 23: Resilience Patterns

| Pattern | Status | Quality | Coverage | Configuration | Testing | Monitoring | Details | Evidence |
|---------|--------|---------|----------|---------------|---------|------------|---------|----------|
| **Retry** | ✅/⚠️/❌ | X/10 | All/Some/None | Env Vars | ✅/❌ | ✅/❌ | Exponential backoff + jitter, max 5 attempts | file:line |
| **Circuit Breaker** | | | | | | | 5 failures OR 50% error rate, 30s timeout | |
| **Timeout** | | | | | | | context.WithTimeout, X seconds | |
| **Bulkhead** | | | | | | | Goroutine pool, X workers | |
| **Fallback** | | | | | | | Cached response + default values | |

### Pattern 1: Retry

**Status**: (✅ Implemented / ⚠️ Partial / ❌ Missing)
**Quality**: X/10

**Implementation Details**:
- **Exponential Backoff**: (✅ Yes / ❌ No)
- **Jitter**: (✅ Yes / ❌ No)
- **Max Attempts**: X
- **Backoff Formula**: (e.g., "min(2^attempt * 100ms, 10s)")
- **Retryable Errors Only**: (✅ Yes / ❌ No - retries everything)

**Coverage**: (All external calls / Some external calls / None)

**Evidence**: infrastructure/channels/waha/retry.go:1-80

**Recommendations**:
1. [Specific action]

### Pattern 2: Circuit Breaker

**Status**: (✅ Implemented / ⚠️ Partial / ❌ Missing)
**Quality**: X/10

**Implementation Details**:
- **Library**: (gobreaker / sony/gobreaker / custom / none)
- **States**: (✅ Closed/Open/Half-Open / ❌ None)
- **Failure Threshold**: X failures or Y% error rate
- **Open Duration**: X seconds
- **Half-Open Test Requests**: X

**Coverage**: (All external calls / Some external calls / None)

**Evidence**: infrastructure/channels/waha/circuit_breaker.go:1-100

**Recommendations**:
1. [Specific action]

### Pattern 3: Timeout

**Status**: (✅ Implemented / ⚠️ Partial / ❌ Missing)
**Quality**: X/10

**Implementation Details**:
- **Context Usage**: X instances of context.WithTimeout
- **HTTP Client Timeouts**: (✅ Yes / ❌ No)
- **Database Timeouts**: (✅ Yes / ❌ No)
- **Timeout Values**: X seconds (configurable via env)
- **Graceful Shutdown**: (✅ Yes / ❌ No)

**Coverage**: (All external calls / Some external calls / None)

**Evidence**: infrastructure/channels/waha/client.go:50-60

**Recommendations**:
1. [Specific action]

### Pattern 4: Bulkhead

**Status**: (✅ Implemented / ⚠️ Partial / ❌ Missing)
**Quality**: X/10

**Implementation Details**:
- **Implementation**: (Goroutine pool / Semaphore / Channel buffering / None)
- **Pool Size**: X workers
- **Queue Size**: Y requests
- **Rejection Policy**: (Block / Reject with error)

**Coverage**: (Critical paths / Some paths / None)

**Evidence**: infrastructure/worker/pool.go:1-100

**Recommendations**:
1. [Specific action]

### Pattern 5: Fallback

**Status**: (✅ Implemented / ⚠️ Partial / ❌ Missing)
**Quality**: X/10

**Implementation Details**:
- **Strategy**: (Cached response / Default value / Degraded mode / None)
- **Fallback Coverage**: X% of critical paths
- **Cache Staleness**: (Acceptable / Not acceptable)

**Coverage**: (Critical paths / Some paths / None)

**Evidence**: infrastructure/channels/waha/fallback.go:1-80

**Recommendations**:
1. [Specific action]

### Deterministic vs AI Validation

| Metric | Deterministic | AI Assessment | Match | Notes |
|--------|---------------|---------------|-------|-------|
| Retry implementations | X | Y | ✅/⚠️ | |
| Circuit breakers | X | Y | ✅/⚠️ | |
| Timeout usage | X | Y | ✅/⚠️ | |
| Bulkhead implementations | X | Y | ✅/⚠️ | |
| Fallback implementations | X | Y | ✅/⚠️ | |

---

## Critical Recommendations

### Immediate Actions (P0)
1. **[Most critical resilience gap]**
   - Why: (Impact on production)
   - How: (Specific implementation steps)
   - Effort: X days
   - Evidence: file:line

2. **[Second critical action]**

### Short-term Improvements (P1)
1. [Action]
2. [Action]

### Long-term Enhancements (P2)
1. [Action]
2. [Action]

---

## Appendix: Discovery Commands

All commands used for atemporal discovery:

```bash
# Rate limiting
grep -r "RateLimit\|Limiter" infrastructure/http/middleware/ | wc -l
grep -r "X-RateLimit\|429" infrastructure/http/ | wc -l

# Error handling
grep -r "type.*Error struct" internal/domain/ | wc -l
grep -r "errors.Wrap\|fmt.Errorf.*%w" internal/ infrastructure/ | wc -l
grep -r "recover()" infrastructure/http/middleware/ | wc -l

# Retry
grep -r "Retry\|retry\|Backoff" infrastructure/ --include="*.go" | wc -l
grep -r "exponential\|Exponential\|jitter" infrastructure/ | wc -l

# Circuit breaker
grep -r "gobreaker\|CircuitBreaker" go.mod infrastructure/ | wc -l

# Timeout
grep -r "context.WithTimeout\|context.WithDeadline" infrastructure/ | wc -l

# Bulkhead
grep -r "semaphore\|Semaphore\|WorkerPool" infrastructure/ | wc -l

# Fallback
grep -r "fallback\|Fallback\|defaultValue" infrastructure/ | wc -l
```

---

**Analysis Version**: 1.0
**Agent Runtime**: X minutes
**Patterns Analyzed**: 7 (rate limiting, error handling, retry, circuit breaker, timeout, bulkhead, fallback)
**Last Updated**: YYYY-MM-DD
```

---

## Success Criteria

- ✅ Deterministic baseline loaded and validated
- ✅ Rate limiting discovered and scored (Table 19)
- ✅ Error handling discovered and scored (Table 20)
- ✅ All 5 resilience patterns discovered and scored (Table 23)
- ✅ Quality scores calculated for each pattern
- ✅ Coverage assessed (% of code using each pattern)
- ✅ Evidence citations for every assessment
- ✅ Deterministic vs AI comparison shows match or explains discrepancies
- ✅ Critical recommendations prioritized (P0/P1/P2)
- ✅ Discovery commands documented in appendix
- ✅ Output written to `code-analysis/quality/resilience_analysis.md`

---

## Critical Rules

1. **Atemporal Discovery** - Use grep/find/wc commands, NO hardcoded numbers
2. **Deterministic Integration** - Always run Step 0, validate AI findings against facts
3. **Complete Tables** - Fill ALL columns for Tables 19, 20, 23
4. **Evidence Required** - Every assessment must cite file:line
5. **Pattern Quality** - Score each pattern 0-10 with clear criteria
6. **Coverage Assessment** - Calculate % of code using each pattern
7. **Actionable Recommendations** - Specific steps, not vague suggestions
8. **Security Focus** - Rate limiting is P0 security issue (resource exhaustion)
9. **Production Readiness** - Be honest about gaps (external calls without retry/circuit breaker = brittle)
10. **Code Examples** - Show Good ✅ vs Bad ❌ for all patterns

---

**Agent Version**: 1.0 (Comprehensive)
**Estimated Runtime**: 50-70 minutes
**Output File**: `code-analysis/quality/resilience_analysis.md`
**Tables Covered**: 19 (Rate Limiting), 20 (Error Handling), 23 (Resilience Patterns)
**Last Updated**: 2025-10-15
