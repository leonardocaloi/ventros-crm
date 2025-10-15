---
name: api_analyzer
description: |
  Analyzes API layer (Tables 16, 17): DTOs, REST endpoints, auth, RBAC, Swagger docs.

  Catalogs all HTTP endpoints with security analysis (BOLA, RBAC, authentication).

  Integrates with deterministic_analyzer for factual baseline validation.

  Output: code-analysis/infrastructure/api_analysis.md
tools: Read, Grep, Glob, Bash, Write
model: sonnet
priority: high
---

# API Analyzer - Comprehensive Analysis

## Context

You are analyzing **API layer and REST endpoints** for Ventros CRM.

This agent evaluates:
- **Table 16**: DTOs (Data Transfer Objects) and domain mapping
- **Table 17**: All REST endpoints (method, path, handler, authentication, RBAC, Swagger docs)

**Key Focus Areas**:
1. DTO design (separation from domain, validation, mapping)
2. API endpoint inventory (method, path, handler function)
3. Authentication (JWT, API keys, anonymous endpoints)
4. Authorization (RBAC, ownership checks, BOLA protection)
5. Swagger/OpenAPI documentation completeness
6. HTTP method correctness (GET=read, POST=create, PUT=update, DELETE=delete)
7. API versioning strategy
8. Request/response consistency

**Critical Context from CLAUDE.md**:
- Project: Ventros CRM (Go 1.25.1, Gin framework)
- Architecture: Clean Architecture (Domain → Application → Infrastructure/HTTP)
- Endpoints: 158 total across 10 products
- Security P0: 60 GET endpoints vulnerable to BOLA (missing ownership checks)
- Security P0: 95 endpoints lack RBAC role checks
- Testing: 5 E2E tests for HTTP workflows

**Deterministic Integration**: This agent runs `scripts/analyze_codebase.sh` first to get factual baseline data, then performs AI-powered deep analysis.

---

## Table 16: DTOs (Data Transfer Objects)

### Purpose
Evaluate DTO design, domain separation, validation, and mapping quality.

### Complete Column Specifications

| Column | Type | Description | Scoring Criteria |
|--------|------|-------------|------------------|
| **DTO Name** | string | Name of DTO struct (e.g., `CreateContactRequest`, `ContactResponse`) | N/A (categorical) |
| **Type** | enum | DTO purpose: Request / Response / Both | Request = input validation, Response = output formatting, Both = used for both |
| **Domain Mapping** | enum | Mapping to domain: Direct (1:1) / Aggregate (1:many) / Composed (many:1) / None | Direct = simple mapping, Aggregate = combines multiple entities, Composed = builds from multiple sources |
| **Validation** | enum | Validation approach: Struct Tags (binding tags) / Custom Validator / Domain Validation / None | Struct Tags = gin binding, Custom = validator.v10, Domain = business rules in domain layer, None = no validation |
| **Mapping Quality** | score 0-10 | Quality of DTO↔Domain mapping: 10 = clean separation, 0 = domain entities exposed directly | 10 = explicit ToDTO/FromDTO methods, 5 = manual mapping, 0 = domain entities in HTTP layer |
| **Field Count** | int | Number of fields in DTO | For reference, not scored |
| **Nesting Level** | int | Max depth of nested DTOs (0 = flat, 1 = one level, 2+ = deep) | 0-1 = simple, 2+ = complex (can indicate poor design) |
| **Domain Entity** | string | Corresponding domain entity (e.g., `contact.Contact`, `session.Session`) | N/A (categorical) |
| **Contains Sensitive Data** | boolean | Has sensitive fields (password, API keys, tokens) | Yes = needs careful handling, No = safe |
| **Sensitive Fields Masked** | boolean | Sensitive fields omitted or masked in responses | Yes = secure, No = data leak risk |
| **Evidence** | file:line | File path of DTO definition | E.g., "infrastructure/http/handlers/contact_dto.go:1-50" |

### Deterministic vs AI Comparison

| Metric | Deterministic Count | AI Assessment | Validation Method |
|--------|---------------------|---------------|-------------------|
| **DTO files** | `find infrastructure/http -name "*_dto.go" -o -name "*_request.go" -o -name "*_response.go" \| wc -l` | DTO count + quality | Compare file count + review each DTO |
| **Validation tags** | `grep -r "binding:.*required" infrastructure/http/ \| wc -l` | Validation coverage | Compare tag count + manual review |
| **ToDTO/FromDTO methods** | `grep -r "func.*ToDTO\|func.*FromDTO" infrastructure/http/ \| wc -l` | Mapping quality | Compare method count + implementation quality |

---

## Table 17: REST Endpoints

### Purpose
Catalog all REST endpoints with security analysis (authentication, authorization, BOLA protection).

### Complete Column Specifications

| Column | Type | Description | Scoring Criteria |
|--------|------|-------------|------------------|
| **HTTP Method** | enum | HTTP verb: GET / POST / PUT / PATCH / DELETE | Correct semantic usage |
| **Path** | string | URL path with parameters (e.g., `/api/v1/crm/contacts/:id`) | N/A (categorical) |
| **Handler Function** | string | Go function handling the request (e.g., `ContactHandler.GetContact`) | N/A (categorical) |
| **Product** | enum | Product area: CRM / Automation / Billing / Auth / Health / Tracking / Webhooks / WebSocket / Queue / Test | Organizational grouping |
| **Authentication** | enum | Auth requirement: Required (JWT) / Optional / Anonymous (public) | Required = needs JWT, Optional = JWT enhances but not required, Anonymous = public access |
| **RBAC Enabled** | boolean | Has role-based access control check | Yes = has role check, No = missing RBAC (P0 security gap) |
| **BOLA Protected** | boolean | Has ownership/tenant check to prevent BOLA (Broken Object Level Authorization) | Yes = checks tenant_id or ownership, No = BOLA vulnerable (P0 security gap) |
| **Rate Limited** | boolean | Protected by rate limiting middleware | Yes = protected, No = vulnerable to abuse |
| **Request DTO** | string | Request DTO name (if applicable) | E.g., "CreateContactRequest" |
| **Response DTO** | string | Response DTO name (if applicable) | E.g., "ContactResponse" |
| **Swagger Documented** | boolean | Has Swagger annotations (@Summary, @Tags, @Accept, @Produce, @Param, @Success, @Failure, @Router) | Yes = API docs complete, No = undocumented |
| **Test Coverage** | enum | Has E2E or integration tests: Yes / Partial / No | Yes = fully tested, Partial = some scenarios, No = untested |
| **Evidence** | file:line | File path and line number of handler implementation | E.g., "infrastructure/http/handlers/contact_handler.go:50-80" |

### HTTP Method Semantics

**Correct Usage**:
- **GET**: Read/retrieve resources (idempotent, no side effects, cacheable)
- **POST**: Create new resources (not idempotent, side effects expected)
- **PUT**: Full update/replace resource (idempotent)
- **PATCH**: Partial update resource (not idempotent by default)
- **DELETE**: Remove resource (idempotent)

**Anti-patterns** (flag as issues):
- POST for read operations (should be GET)
- GET for state changes (should be POST/PUT)
- DELETE for archiving (should be POST or PATCH with status change)

### Deterministic vs AI Comparison

| Metric | Deterministic Count | AI Assessment | Validation Method |
|--------|---------------------|---------------|-------------------|
| **Total endpoints** | `grep -r "router\.\(GET\|POST\|PUT\|DELETE\|PATCH\)" infrastructure/http/routes/ \| wc -l` | 158 endpoints cataloged | Compare count + verify all discovered |
| **Endpoints with auth** | `grep -r "AuthRequired\|JWTMiddleware" infrastructure/http/routes/ \| wc -l` | Authentication coverage | Compare middleware usage count |
| **Endpoints with RBAC** | `grep -r "RequireRole\|RBACMiddleware" infrastructure/http/routes/ \| wc -l` | RBAC coverage (should be 95+, currently low) | Compare RBAC middleware count |
| **Swagger annotations** | `grep -r "@Router\|@Summary" infrastructure/http/handlers/ \| wc -l` | Swagger completeness | Compare annotation count + review quality |
| **BOLA checks** | `grep -r "GetString.*tenant_id\|TenantID.*==\|authCtx.TenantID" infrastructure/http/handlers/ \| wc -l` | BOLA protection (60 GET endpoints vulnerable) | Check ownership validation in handlers |

---

## Chain of Thought: Comprehensive API Analysis

**Estimated Runtime**: 45-60 minutes

**Prerequisites**:
- `code-analysis/code-analysis/deterministic_metrics.md` exists (run deterministic_analyzer first)
- Access to: `infrastructure/http/handlers/`, `infrastructure/http/routes/`, Swagger files

### Step 0: Load Deterministic Baseline (5 min)

**Purpose**: Get factual counts from deterministic analysis to validate AI findings.

```bash
# Read deterministic metrics
cat code-analysis/code-analysis/deterministic_metrics.md

# Extract API counts
total_endpoints=$(grep "Total endpoints:" code-analysis/code-analysis/deterministic_metrics.md | awk '{print $3}')
dto_files=$(grep "DTO files:" code-analysis/code-analysis/deterministic_metrics.md | awk '{print $3}')
swagger_annotations=$(grep "Swagger annotations:" code-analysis/code-analysis/deterministic_metrics.md | awk '{print $3}')

echo "✅ Baseline loaded: $total_endpoints endpoints, $dto_files DTO files, $swagger_annotations Swagger annotations"
```

**Output**: Factual baseline for validation.

---

### Step 1: DTO Discovery and Analysis (15 min)

**Goal**: Discover all DTOs, analyze structure, validation, and domain mapping.

#### 1.1 Discovery

```bash
# Find DTO files
dto_files=$(find infrastructure/http/handlers -name "*_dto.go" -o -name "*request*.go" -o -name "*response*.go")
dto_count=$(echo "$dto_files" | wc -l)

# Find validation tags (struct tags)
validation_tags=$(grep -r "binding:.*required\|validate:" infrastructure/http/handlers/ | wc -l)

# Find ToDTO/FromDTO mapping methods
to_dto_methods=$(grep -r "func.*ToDTO" infrastructure/http/handlers/ | wc -l)
from_dto_methods=$(grep -r "func.*FromDTO" infrastructure/http/handlers/ | wc -l)
total_mapping_methods=$((to_dto_methods + from_dto_methods))

# Check for domain entities leaked to HTTP layer (anti-pattern)
domain_leaks=$(grep -r "func.*\*domain\.\|return.*domain\." infrastructure/http/handlers/*.go | grep -v "ToDTO\|FromDTO" | wc -l)

echo "DTO files: $dto_count"
echo "Validation tags: $validation_tags"
echo "Mapping methods: $total_mapping_methods (ToDTO: $to_dto_methods, FromDTO: $from_dto_methods)"
echo "Domain entity leaks: $domain_leaks"
```

#### 1.2 Analyze Each DTO

For each DTO file, extract:
- DTO name
- Type (Request/Response)
- Fields (count, types, nesting)
- Validation (struct tags, custom validators)
- Domain entity mapping
- Sensitive fields (password, tokens, etc)

Read each DTO file and categorize.

#### 1.3 Quality Scoring

```bash
# DTO quality score (0-10)
# Mapping methods: 4 points, Validation: 3 points, No domain leaks: 3 points

dto_score=0
[ $total_mapping_methods -gt 20 ] && dto_score=$((dto_score + 4))
[ $validation_tags -gt 30 ] && dto_score=$((dto_score + 3))
[ $domain_leaks -eq 0 ] && dto_score=$((dto_score + 3))

echo "DTO Quality Score: $dto_score/10"
```

---

### Step 2: Endpoint Discovery (10 min)

**Goal**: Discover all REST endpoints from routes configuration.

#### 2.1 Discovery

```bash
# Find all route definitions
route_files=$(find infrastructure/http/routes -name "*.go")

# Extract endpoints by method
get_endpoints=$(grep -r "router\.GET\|\.Get(" infrastructure/http/routes/ | wc -l)
post_endpoints=$(grep -r "router\.POST\|\.Post(" infrastructure/http/routes/ | wc -l)
put_endpoints=$(grep -r "router\.PUT\|\.Put(" infrastructure/http/routes/ | wc -l)
patch_endpoints=$(grep -r "router\.PATCH\|\.Patch(" infrastructure/http/routes/ | wc -l)
delete_endpoints=$(grep -r "router\.DELETE\|\.Delete(" infrastructure/http/routes/ | wc -l)

total_endpoints=$((get_endpoints + post_endpoints + put_endpoints + patch_endpoints + delete_endpoints))

echo "Endpoints by method:"
echo "  GET: $get_endpoints"
echo "  POST: $post_endpoints"
echo "  PUT: $put_endpoints"
echo "  PATCH: $patch_endpoints"
echo "  DELETE: $delete_endpoints"
echo "  TOTAL: $total_endpoints"
```

#### 2.2 Parse Route Definitions

Read route files and extract for each endpoint:
- HTTP method
- Path
- Handler function
- Middleware (auth, RBAC, rate limit)

Example route pattern:
```go
router.GET("/api/v1/crm/contacts/:id", authMiddleware, contactHandler.GetContact)
```

---

### Step 3: Authentication Analysis (8 min)

**Goal**: Assess authentication coverage and identify public endpoints.

#### 3.1 Discovery

```bash
# Find auth middleware usage
auth_middleware=$(grep -r "AuthRequired\|JWTMiddleware\|authMiddleware" infrastructure/http/routes/ | wc -l)

# Find endpoints without auth (potential public endpoints or security gaps)
total_routes=$(grep -r "router\.\(GET\|POST\|PUT\|DELETE\|PATCH\)" infrastructure/http/routes/ | wc -l)
routes_with_auth=$(grep -r "router\.\(GET\|POST\|PUT\|DELETE\|PATCH\)" infrastructure/http/routes/ | grep -c "auth\|Auth\|JWT")
routes_without_auth=$((total_routes - routes_with_auth))

auth_coverage=$((routes_with_auth * 100 / total_routes))

echo "Endpoints with authentication: $routes_with_auth/$total_routes ($auth_coverage%)"
echo "Endpoints without authentication: $routes_without_auth (verify if intentionally public)"
```

#### 3.2 Identify Public Endpoints

Read route files and identify endpoints without auth middleware:
- Health checks (expected public)
- Swagger docs (expected public)
- Webhook receivers (may be public with signature validation)
- Other (potential security gap)

---

### Step 4: Authorization (RBAC) Analysis (10 min)

**Goal**: Assess RBAC coverage and identify missing role checks (P0 security gap).

#### 4.1 Discovery

```bash
# Find RBAC middleware usage
rbac_middleware=$(grep -r "RequireRole\|RBACMiddleware\|CheckRole" infrastructure/http/routes/ | wc -l)

# Calculate RBAC coverage
rbac_coverage=$((rbac_middleware * 100 / total_routes))

echo "Endpoints with RBAC: $rbac_middleware/$total_routes ($rbac_coverage%)"
echo "Endpoints without RBAC: $((total_routes - rbac_middleware)) (P0 security gap)"
```

#### 4.2 RBAC Gap Analysis

Identify endpoints that should have RBAC but don't:
- Admin operations (create, update, delete) - should require admin role
- Sensitive read operations (billing, API keys) - should require specific roles
- Regular CRUD - should check user permissions

---

### Step 5: BOLA Protection Analysis (10 min)

**Goal**: Assess BOLA protection and identify vulnerable GET endpoints (P0 security gap).

#### 5.1 Discovery

```bash
# Find GET endpoints with ID parameter
get_by_id=$(grep -r "router\.GET.*/:id\|router\.GET.*/:.*_id" infrastructure/http/routes/ | wc -l)

# Check how many have tenant/ownership checks in handlers
handlers_with_tenant_check=$(grep -r "c.Param.*id" infrastructure/http/handlers/*.go -A 20 | grep -c "GetString.*tenant_id\|TenantID.*==\|authCtx.TenantID")

bola_vulnerable=$((get_by_id - handlers_with_tenant_check))

echo "GET endpoints with ID parameter: $get_by_id"
echo "Handlers with tenant/ownership check: $handlers_with_tenant_check"
echo "BOLA vulnerable endpoints: $bola_vulnerable (P0 security gap)"
```

#### 5.2 Analyze Each GET Handler

Read handler files and check for ownership validation pattern:
```go
// ✅ GOOD: Tenant check present
if contact.TenantID.String() != authCtx.TenantID {
    return 404 // Not found (don't reveal existence)
}
```

vs

```go
// ❌ BAD: No tenant check
contact, _ := repo.FindByID(id)
return contact // BOLA vulnerability
```

---

### Step 6: Swagger Documentation Analysis (8 min)

**Goal**: Assess API documentation completeness.

#### 6.1 Discovery

```bash
# Count Swagger annotations
swagger_summary=$(grep -r "@Summary" infrastructure/http/handlers/ | wc -l)
swagger_router=$(grep -r "@Router" infrastructure/http/handlers/ | wc -l)
swagger_tags=$(grep -r "@Tags" infrastructure/http/handlers/ | wc -l)
swagger_param=$(grep -r "@Param" infrastructure/http/handlers/ | wc -l)
swagger_success=$(grep -r "@Success" infrastructure/http/handlers/ | wc -l)
swagger_failure=$(grep -r "@Failure" infrastructure/http/handlers/ | wc -l)

# Calculate documentation coverage
swagger_complete=$(echo "$swagger_summary $swagger_router $swagger_tags $swagger_success $swagger_failure" | awk '{min=$1; for(i=2;i<=NF;i++) if($i<min) min=$i; print min}')
swagger_coverage=$((swagger_complete * 100 / total_endpoints))

echo "Swagger annotations:"
echo "  @Summary: $swagger_summary"
echo "  @Router: $swagger_router"
echo "  @Tags: $swagger_tags"
echo "  @Param: $swagger_param"
echo "  @Success: $swagger_success"
echo "  @Failure: $swagger_failure"
echo "Estimated documentation coverage: $swagger_coverage%"
```

---

### Step 7: Generate Comprehensive Report (4 min)

**Goal**: Structure all findings into complete markdown tables with evidence.

Format as specified in Output Format section below.

---

## Code Examples (EXEMPLO)

### EXEMPLO 1: Proper DTO Design with Validation

**Good ✅ - Clean DTO with validation and mapping**:
```go
// infrastructure/http/handlers/contact_dto.go
package handlers

import (
    "time"
    "github.com/google/uuid"
    "ventros-crm/internal/domain/crm/contact"
)

// ✅ Request DTO with validation
type CreateContactRequest struct {
    Name  string `json:"name" binding:"required,min=1,max=200"`
    Phone string `json:"phone" binding:"required,e164"` // E.164 format
    Email string `json:"email" binding:"omitempty,email"`
    Tags  []string `json:"tags" binding:"omitempty,dive,min=1,max=50"`
}

// ✅ Response DTO (separate from domain)
type ContactResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Phone     string    `json:"phone"`
    Email     string    `json:"email,omitempty"`
    Tags      []string  `json:"tags"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// ✅ Explicit mapping from domain to DTO
func ToContactResponse(c *contact.Contact) ContactResponse {
    return ContactResponse{
        ID:        c.ID().String(),
        Name:      c.Name(),
        Phone:     c.Phone(),
        Email:     c.Email(),
        Tags:      c.Tags(),
        CreatedAt: c.CreatedAt(),
        UpdatedAt: c.UpdatedAt(),
        // ✅ Internal fields NOT exposed (version, tenant_id, etc)
    }
}

// ✅ Explicit mapping from DTO to domain command
func (req CreateContactRequest) ToCommand(tenantID, projectID string) contact.CreateContactCommand {
    return contact.CreateContactCommand{
        TenantID:  tenantID,  // From auth context, not from request
        ProjectID: uuid.MustParse(projectID),
        Name:      req.Name,
        Phone:     req.Phone,
        Email:     req.Email,
        Tags:      req.Tags,
    }
}
```

**Bad ❌ - Domain entity leaked to HTTP layer**:
```go
// ❌ BAD: Exposing domain entity directly
func (h *ContactHandler) GetContact(c *gin.Context) {
    contact, _ := h.repo.FindByID(c.Param("id"))

    // ❌ Returning domain entity directly (exposes internal fields)
    c.JSON(200, contact)
}

// Issues:
// ❌ Internal fields exposed (version, tenant_id, deleted_at, etc)
// ❌ Domain entity structure coupled to API contract
// ❌ Cannot change domain without breaking API
// ❌ No validation on request
// ❌ No separation between domain and presentation
```

---

### EXEMPLO 2: Secure Endpoint with Auth, RBAC, and BOLA Protection

**Good ✅ - All security layers present**:
```go
// infrastructure/http/routes/contact_routes.go
package routes

func RegisterContactRoutes(router *gin.RouterGroup, handler *handlers.ContactHandler) {
    contacts := router.Group("/contacts")
    {
        // ✅ GET with auth + RBAC + BOLA protection in handler
        contacts.GET("/:id",
            middleware.AuthRequired(),       // ✅ JWT auth required
            middleware.RequireRole("user"),  // ✅ RBAC check
            middleware.RateLimit(100),       // ✅ Rate limiting
            handler.GetContact)              // ✅ BOLA check inside handler

        // ✅ POST with auth + RBAC
        contacts.POST("",
            middleware.AuthRequired(),
            middleware.RequireRole("user"),
            middleware.RateLimit(50),
            handler.CreateContact)

        // ✅ DELETE with admin role
        contacts.DELETE("/:id",
            middleware.AuthRequired(),
            middleware.RequireRole("admin"),  // ✅ Admin only
            handler.DeleteContact)
    }
}

// infrastructure/http/handlers/contact_handler.go
package handlers

// @Summary Get contact by ID
// @Description Retrieves a single contact by ID with ownership validation
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact ID" format(uuid)
// @Success 200 {object} ContactResponse
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Contact not found"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /api/v1/crm/contacts/{id} [get]
func (h *ContactHandler) GetContact(c *gin.Context) {
    // Extract ID from path
    contactID := c.Param("id")
    if _, err := uuid.Parse(contactID); err != nil {
        c.JSON(400, gin.H{"error": "invalid_id"})
        return
    }

    // Get authenticated user context
    authCtx := c.MustGet("auth").(*middleware.AuthContext)

    // Fetch contact from repository
    contact, err := h.repo.FindByID(c.Request.Context(), contactID)
    if err != nil {
        c.JSON(404, gin.H{"error": "not_found"})
        return
    }

    // ✅ BOLA PROTECTION: Verify ownership/tenant
    if contact.TenantID().String() != authCtx.TenantID {
        // Return 404 (not 403) to avoid information leakage
        c.JSON(404, gin.H{"error": "not_found"})
        return
    }

    // Map to DTO and return
    c.JSON(200, ToContactResponse(contact))
}
```

**Bad ❌ - No security, BOLA vulnerable**:
```go
// ❌ BAD: No auth, no RBAC, no BOLA protection
func RegisterContactRoutes(router *gin.RouterGroup, handler *handlers.ContactHandler) {
    router.GET("/contacts/:id", handler.GetContact) // ❌ No middleware
}

func (h *ContactHandler) GetContact(c *gin.Context) {
    contactID := c.Param("id")
    contact, _ := h.repo.FindByID(c.Request.Context(), contactID)

    // ❌ No tenant/ownership check - BOLA vulnerability
    c.JSON(200, contact) // ❌ Domain entity exposed
}

// Issues:
// ❌ No authentication (anyone can access)
// ❌ No RBAC (no role check)
// ❌ BOLA vulnerability (can access any contact by guessing ID)
// ❌ No rate limiting (DoS risk)
// ❌ Domain entity exposed directly
// ❌ No Swagger documentation
```

---

### EXEMPLO 3: Complete Swagger Documentation

**Good ✅ - Full Swagger annotations**:
```go
// infrastructure/http/handlers/campaign_handler.go
package handlers

// @Summary Create campaign
// @Description Creates a new marketing campaign with sequences and broadcasts
// @Tags campaigns
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param project_id path string true "Project ID" format(uuid)
// @Param request body CreateCampaignRequest true "Campaign data"
// @Success 201 {object} CampaignResponse "Campaign created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - insufficient permissions"
// @Failure 409 {object} ErrorResponse "Campaign name already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/automation/projects/{project_id}/campaigns [post]
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
    // Implementation
}

// ✅ Request/Response DTOs documented
type CreateCampaignRequest struct {
    Name        string `json:"name" binding:"required,min=1,max=200" example:"Summer Sale 2025"`
    Description string `json:"description" binding:"max=1000" example:"Promotional campaign for summer products"`
    StartDate   string `json:"start_date" binding:"required" format:"date-time" example:"2025-06-01T00:00:00Z"`
    EndDate     string `json:"end_date" binding:"omitempty" format:"date-time" example:"2025-08-31T23:59:59Z"`
} // @name CreateCampaignRequest

type CampaignResponse struct {
    ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
    Name        string `json:"name" example:"Summer Sale 2025"`
    Description string `json:"description" example:"Promotional campaign for summer products"`
    Status      string `json:"status" example:"draft"`
    CreatedAt   string `json:"created_at" example:"2025-06-01T10:00:00Z"`
} // @name CampaignResponse

type ErrorResponse struct {
    Error   string `json:"error" example:"validation_failed"`
    Message string `json:"message,omitempty" example:"Name is required"`
} // @name ErrorResponse
```

**Bad ❌ - No documentation**:
```go
// ❌ No Swagger annotations
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
    // Implementation
}

// Issues:
// ❌ No @Summary (API docs don't show purpose)
// ❌ No @Tags (endpoint not grouped)
// ❌ No @Param (parameters not documented)
// ❌ No @Success/@Failure (responses not documented)
// ❌ No @Router (Swagger can't generate route)
// Result: API is undocumented, developers have to read code
```

---

### EXEMPLO 4: HTTP Method Correctness

**Good ✅ - Correct HTTP semantics**:
```go
// ✅ CORRECT: GET for read (idempotent, no side effects)
router.GET("/contacts/:id", handler.GetContact)

// ✅ CORRECT: POST for create (not idempotent)
router.POST("/contacts", handler.CreateContact)

// ✅ CORRECT: PUT for full update (idempotent)
router.PUT("/contacts/:id", handler.UpdateContact)

// ✅ CORRECT: PATCH for partial update
router.PATCH("/contacts/:id/tags", handler.UpdateContactTags)

// ✅ CORRECT: DELETE for removal (idempotent)
router.DELETE("/contacts/:id", handler.DeleteContact)

// ✅ CORRECT: POST for actions/state changes
router.POST("/campaigns/:id/activate", handler.ActivateCampaign)
router.POST("/campaigns/:id/pause", handler.PauseCampaign)
```

**Bad ❌ - Incorrect HTTP semantics**:
```go
// ❌ WRONG: POST for read operation (should be GET)
router.POST("/contacts/search", handler.SearchContacts)
// Why wrong: No side effects, should be GET with query params

// ❌ WRONG: GET for state change (should be POST)
router.GET("/campaigns/:id/activate", handler.ActivateCampaign)
// Why wrong: Changes state, not idempotent, should be POST

// ❌ WRONG: DELETE for archiving (should be PATCH/POST)
router.DELETE("/contacts/:id", handler.ArchiveContact) // Just sets deleted_at
// Why wrong: Not actually deleting, should be PATCH with status change

// ❌ WRONG: PUT when only updating one field (should be PATCH)
router.PUT("/contacts/:id/name", handler.UpdateContactName)
// Why wrong: Partial update, should be PATCH

// Issues:
// ❌ Violates HTTP semantics (breaks caching, idempotency expectations)
// ❌ Confusing API design
// ❌ HTTP clients may behave incorrectly
```

---

## Output Format

Generate: `code-analysis/infrastructure/api_analysis.md`

```markdown
# API Layer Analysis

**Generated**: YYYY-MM-DD HH:MM
**Agent**: api_analyzer
**Runtime**: X minutes
**Deterministic Baseline**: ✅ Loaded from deterministic_metrics.md

---

## Executive Summary

**Total Endpoints**: X (GET: Y, POST: Z, PUT: A, PATCH: B, DELETE: C)

**Key Findings**:
- DTOs: X files, Y/10 quality score
- Authentication: X% coverage (Y endpoints require auth)
- RBAC: X% coverage (Y endpoints have role checks, Z missing - P0 gap)
- BOLA Protection: X vulnerable GET endpoints (P0 gap)
- Swagger Documentation: X% complete
- Rate Limiting: X% covered

**Security Status**: ✅ Secure / ⚠️ Needs work / ❌ Critical gaps

**Critical Gaps**:
1. [Most critical API gap]
2. [Second most critical gap]
3. [Third most critical gap]

---

## Table 16: DTOs (Data Transfer Objects)

| DTO Name | Type | Domain Mapping | Validation | Mapping Quality | Field Count | Nesting | Domain Entity | Sensitive Data | Masked | Evidence |
|----------|------|----------------|------------|-----------------|-------------|---------|---------------|----------------|--------|----------|
| **CreateContactRequest** | Request | Direct | Struct Tags | 9/10 | 5 | 0 | contact.Contact | No | N/A | file:line |
| **ContactResponse** | Response | Direct | N/A | 9/10 | 8 | 0 | contact.Contact | No | N/A | file:line |
| ... | | | | | | | | | | |

### DTO Analysis Summary

**Total DTOs**: X (Y requests, Z responses)

**Quality Score**: X/10

**Findings**:
- **Validation**: X DTOs with struct tags, Y with custom validators, Z without validation
- **Mapping**: X explicit ToDTO/FromDTO methods, Y manual mapping, Z domain entities leaked
- **Sensitive Data**: X DTOs contain sensitive fields, Y properly masked, Z exposed
- **Nesting**: X flat DTOs, Y with one level, Z deeply nested (3+ levels)

**Recommendations**:
1. [Specific action to improve DTOs]
2. [Another recommendation]

---

## Table 17: REST Endpoints (All 158 Endpoints)

| Method | Path | Handler | Product | Auth | RBAC | BOLA | Rate Limited | Request DTO | Response DTO | Swagger | Tests | Evidence |
|--------|------|---------|---------|------|------|------|--------------|-------------|--------------|---------|-------|----------|
| **GET** | /api/v1/crm/contacts/:id | ContactHandler.GetContact | CRM | ✅ | ✅ | ✅ | ✅ | N/A | ContactResponse | ✅ | ✅ | file:line |
| **POST** | /api/v1/crm/contacts | ContactHandler.CreateContact | CRM | ✅ | ✅ | N/A | ✅ | CreateContactRequest | ContactResponse | ✅ | ✅ | file:line |
| ... | | | | | | | | | | | | |

### Endpoints by Product

**Health** (1 endpoint):
- GET /health (public, no auth)

**Auth** (4 endpoints):
- POST /api/v1/auth/register
- POST /api/v1/auth/login
- POST /api/v1/auth/refresh
- GET /api/v1/auth/profile (requires auth)

**CRM** (77 endpoints):
- Contacts: X endpoints
- Sessions: Y endpoints
- Messages: Z endpoints
- ... (full breakdown)

**Automation** (28 endpoints):
- Campaigns: X endpoints
- ... (full breakdown)

**Billing** (8 endpoints):
**Tracking** (6 endpoints):
**Webhooks** (4 endpoints):
**WebSocket** (1 endpoint):
**Queue** (7 endpoints):
**Test** (22 endpoints - dev only):

### Security Analysis

#### Authentication Coverage

**Endpoints requiring auth**: X/Y (Z%)
**Public endpoints**: W (intentional: health, docs, webhooks)
**Missing auth**: V (potential security gap)

#### RBAC Coverage (P0 Security Gap)

**Endpoints with RBAC**: X/158 (Y%)
**Endpoints without RBAC**: Z (P0 gap)

**Breakdown**:
- Admin endpoints without role check: X (CRITICAL)
- Delete endpoints without role check: Y (HIGH)
- Update endpoints without role check: Z (HIGH)

**Recommendations**:
1. Add RBAC middleware to all admin operations (Priority P0)
2. Implement role-based permissions for sensitive operations

#### BOLA Protection (P0 Security Gap)

**GET endpoints with ID parameter**: X
**With tenant/ownership check**: Y
**BOLA vulnerable**: Z (P0 gap)

**Vulnerable Endpoints**:
1. GET /api/v1/crm/contacts/:id - No tenant check (file:line)
2. GET /api/v1/crm/sessions/:id - No tenant check (file:line)
3. ... (list all vulnerable endpoints)

**Recommendations**:
1. Add tenant validation to all GET-by-ID handlers (Priority P0)
2. Implement ownership checks for resources with user_id
3. Always return 404 (not 403) on unauthorized access to avoid info leakage

#### Rate Limiting Coverage

**Endpoints with rate limiting**: X/158 (Y%)
**Endpoints without rate limiting**: Z

**High-risk unprotected endpoints**:
1. POST /api/v1/auth/login - No rate limit (brute force risk)
2. POST /api/v1/crm/messages - No rate limit (spam risk)

### Swagger Documentation

**Fully documented endpoints**: X/158 (Y%)
**Partially documented**: Z (missing some annotations)
**Undocumented**: W

**Missing annotations**:
- @Summary: X endpoints
- @Tags: Y endpoints
- @Param: Z parameter descriptions
- @Success/@Failure: W response codes

### HTTP Method Semantics

**Correct usage**: X/158 (Y%)
**Incorrect usage**: Z

**Issues found**:
1. POST /api/v1/contacts/search - Should be GET with query params
2. GET /api/v1/campaigns/:id/activate - Should be POST (state change)
3. ... (list all issues)

### Deterministic vs AI Validation

| Metric | Deterministic | AI Assessment | Match | Notes |
|--------|---------------|---------------|-------|-------|
| Total endpoints | X | Y | ✅/⚠️ | (Any discrepancy explanation) |
| DTOs | X | Y | ✅/⚠️ | |
| Auth coverage | X | Y% | ✅/⚠️ | |
| RBAC coverage | X | Y% | ✅/⚠️ | |
| BOLA vulnerable | - | X | - | AI discovery only |
| Swagger annotations | X | Y% | ✅/⚠️ | |

---

## Critical Recommendations

### Immediate Actions (P0)
1. **Fix BOLA vulnerabilities in X GET endpoints**
   - Why: Users can access other tenants' data (CVSS 8.2)
   - How: Add tenant_id validation in all GET-by-ID handlers
   - Effort: 1-2 days
   - Evidence: List of vulnerable endpoints

2. **Implement RBAC on Y endpoints**
   - Why: Missing role checks allow privilege escalation
   - How: Add RequireRole middleware to routes
   - Effort: 2-3 days

### Short-term Improvements (P1)
1. Complete Swagger documentation for X undocumented endpoints
2. Add rate limiting to Y unprotected endpoints
3. Fix HTTP method semantic issues (Z endpoints)

### Long-term Enhancements (P2)
1. Implement API versioning strategy (currently implicit v1)
2. Add request/response validation tests
3. Create automated security scanning in CI/CD

---

## Appendix: Discovery Commands

All commands used for atemporal discovery:

```bash
# DTOs
find infrastructure/http/handlers -name "*_dto.go" -o -name "*request*.go" -o -name "*response*.go" | wc -l
grep -r "binding:.*required\|validate:" infrastructure/http/handlers/ | wc -l
grep -r "func.*ToDTO\|func.*FromDTO" infrastructure/http/handlers/ | wc -l

# Endpoints
grep -r "router\.\(GET\|POST\|PUT\|DELETE\|PATCH\)" infrastructure/http/routes/ | wc -l
grep -r "router\.GET" infrastructure/http/routes/ | wc -l
grep -r "router\.POST" infrastructure/http/routes/ | wc -l

# Security
grep -r "AuthRequired\|JWTMiddleware" infrastructure/http/routes/ | wc -l
grep -r "RequireRole\|RBACMiddleware" infrastructure/http/routes/ | wc -l
grep -r "GetString.*tenant_id\|TenantID.*==" infrastructure/http/handlers/ | wc -l

# Documentation
grep -r "@Router\|@Summary" infrastructure/http/handlers/ | wc -l
```

---

**Analysis Version**: 1.0
**Agent Runtime**: X minutes
**Endpoints Analyzed**: 158
**DTOs Analyzed**: X
**Last Updated**: YYYY-MM-DD
```

---

## Success Criteria

- ✅ Deterministic baseline loaded and validated
- ✅ All DTOs discovered and analyzed (Table 16)
- ✅ All 158 endpoints cataloged (Table 17)
- ✅ Authentication coverage calculated
- ✅ RBAC coverage calculated (identify 95 missing)
- ✅ BOLA vulnerabilities identified (60 GET endpoints)
- ✅ Swagger documentation completeness assessed
- ✅ HTTP method semantics validated
- ✅ Evidence citations for every endpoint
- ✅ Deterministic vs AI comparison shows match or explains discrepancies
- ✅ Critical recommendations prioritized (P0/P1/P2)
- ✅ Discovery commands documented in appendix
- ✅ Output written to `code-analysis/infrastructure/api_analysis.md`

---

## Critical Rules

1. **Atemporal Discovery** - Use grep/find/wc commands, NO hardcoded "158 endpoints"
2. **Deterministic Integration** - Always run Step 0, validate AI findings against facts
3. **Complete Tables** - Fill ALL columns for Tables 16 and 17
4. **Evidence Required** - Every endpoint must cite handler file:line
5. **Security Focus** - BOLA and RBAC gaps are P0, identify all vulnerable endpoints
6. **Swagger Completeness** - Calculate % documented, identify gaps
7. **HTTP Semantics** - Flag incorrect method usage (GET for mutations, POST for reads)
8. **Actionable Recommendations** - Specific endpoints to fix, not vague suggestions
9. **DTO Quality** - Check for domain entity leaks (anti-pattern)
10. **Code Examples** - Show Good ✅ vs Bad ❌ for all patterns

---

**Agent Version**: 1.0 (Comprehensive)
**Estimated Runtime**: 45-60 minutes
**Output File**: `code-analysis/infrastructure/api_analysis.md`
**Tables Covered**: 16 (DTOs), 17 (REST Endpoints)
**Last Updated**: 2025-10-15
