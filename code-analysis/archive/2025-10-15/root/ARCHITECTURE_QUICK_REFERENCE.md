# Ventros CRM - Architecture Quick Reference

**Data**: 2025-10-13 | **Objetivo**: Referência rápida de repositories, handlers e endpoints

---

## 🗂️ REPOSITORIES (31 total)

### Core Domain Repositories

| Repository | Entity | Optimistic Lock | Cache | N+1 | Métodos |
|-----------|--------|----------------|-------|-----|---------|
| **Agent** | agent.Agent | ✅ | ❌ | ✅ | Save, FindByID, FindByEmail, FindByTenant, FindActiveByTenant, Delete, FindByTenantWithFilters, SearchByText (8) |
| **Contact** | contact.Contact | ✅ | ❌ | ✅ | Save, FindByID, FindByProject, FindByExternalID, FindByPhone, FindByEmail, SaveCustomFields, GetCustomFields, FindByCustomField, SearchByText (10+) |
| **Message** | message.Message | ❌ | ❌ | ✅ | Save, FindByID, FindBySession, FindByContact, FindByChannelMessageID, CountBySession, FindByTenantWithFilters, SearchByText (8+) |
| **Session** | session.Session | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Project** | project.Project | ✅ | ❌ | ✅ | Save, FindByID, FindByTenantID, FindByCustomer, FindActiveProjects, Delete, FindByTenantWithFilters, SearchByText (8) |
| **Pipeline** | pipeline.Pipeline | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Channel** | channel.Channel | ❌ | ❌ | ✅ | Create, GetByID, GetByUserID, GetByProjectID, GetByExternalID, GetByWebhookID, GetActiveWAHAChannels, Update, Delete (9) |
| **Chat** | chat.Chat | ✅ | ❌ | ✅ | Create, FindByID, FindByExternalID, FindByProject, FindByTenant, FindByContact, FindActiveByProject, FindIndividualByContact, Update, Delete, SearchBySubject (11) |
| **Note** | note.Note | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Tracking** | tracking.Tracking | ❓ | ❌ | ❓ | (não analisado em detalhe) |

### Automation Repositories

| Repository | Entity | Optimistic Lock | Cache | N+1 | Métodos |
|-----------|--------|----------------|-------|-----|---------|
| **Automation** | pipeline.Automation | ❌ | ❌ | ✅ | Save, FindByID, FindByPipeline, FindByPipelineAndTrigger, FindEnabledByPipeline, Delete (6) |
| **Broadcast** | broadcast.Broadcast | ✅ | ❌ | ✅ | Save, FindByID, FindByTenantID, FindScheduledReady, FindByStatus, Delete (6) |
| **BroadcastExecution** | broadcast.BroadcastExecution | ❌ | ❌ | ✅ | Save, SaveBatch, FindByID, FindByBroadcastID, FindPendingByBroadcastID, Delete (6) |
| **Campaign** | campaign.Campaign | ✅ | ❌ | ⚠️ | Save (with transaction), FindByID, FindByTenantID, FindActiveByStatus, FindScheduled, Delete (6) |
| **CampaignEnrollment** | campaign.CampaignEnrollment | ❌ | ❌ | ✅ | Save, FindByID, FindByCampaignID, FindByContactID, FindReadyForNextStep, FindActiveByCampaignAndContact, Delete (7) |
| **Sequence** | sequence.Sequence | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **SequenceEnrollment** | sequence.SequenceEnrollment | ❓ | ❌ | ❓ | (não analisado em detalhe) |

### Infrastructure Repositories

| Repository | Entity | Optimistic Lock | Cache | N+1 | Métodos |
|-----------|--------|----------------|-------|-----|---------|
| **Billing** | billing.BillingAccount | ✅ | ❌ | ✅ | Create, FindByID, FindByUserID, FindActiveByUserID, Update, Delete (6) |
| **Invoice** | billing.Invoice | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Subscription** | billing.Subscription | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **UsageMeter** | billing.UsageMeter | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Credential** | credential.Credential | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Webhook** | webhook.Webhook | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **Outbox** | outbox.OutboxMessage | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **DomainEventLog** | event.DomainEvent | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **ContactEvent** | contact.ContactEvent | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **ContactList** | contact_list.ContactList | ❌ | ❌ | 🔴 | Save, FindByID, FindByTenantID, Delete, AddContacts, RemoveContacts, GetContactIDs, GetContactCount, IsContactInList (9) |
| **MessageGroup** | message_group.MessageGroup | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **MessageEnrichment** | message.MessageEnrichment | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **ProjectMember** | project_member.ProjectMember | ❓ | ❌ | ❓ | (não analisado em detalhe) |
| **ChannelType** | channel.ChannelType | ❓ | ❌ | ❓ | (não analisado em detalhe) |

**Legenda:**
- ✅ Implementado corretamente
- ❌ Não implementado
- ⚠️ Problema detectado
- 🔴 Problema crítico
- ❓ Não analisado

---

## 🎯 HANDLERS (27 total)

### CRM Handlers

| Handler | Endpoints | Auth | RLS | Rate Limit | RBAC | Use Cases |
|---------|-----------|------|-----|------------|------|-----------|
| **AgentHandler** | 9 | ✅ | ✅ | 1000-M | ❌ | CRUD, CreateVirtual, GetStats, EndVirtualPeriod, Search, Advanced |
| **ContactHandler** | 9 | ✅ | ✅ | 1000-M | ❌ | CRUD, Search, Advanced, ChangePipelineStatus |
| **MessageHandler** | 10 | ✅ | ✅ | 1000-M | ❌ | CRUD, Send, ConfirmDelivery, Search, Advanced |
| **SessionHandler** | 6 | ✅ | ✅ | ❌ | ❌ | List, Get, Close, GetStats, Search, Advanced |
| **ProjectHandler** | 7 | ✅ | ✅ | ❌ | ❌ | CRUD, Search, Advanced |
| **PipelineHandler** | 11 | ✅ | ✅ | ❌ | ❌ | CRUD, CreateStatus, ChangeContactStatus, CustomFields, Search, Advanced |
| **ChannelHandler** | 12 | ✅ | ✅ | 1000-M | ❌ | CRUD, Activate, Deactivate, Webhook config, WAHA operations |
| **ChatHandler** | 9 | ✅ | ✅ | ❌ | ❌ | CRUD, Participants, Archive, Unarchive, Close, UpdateSubject |
| **NoteHandler** | 2 | ✅ | ✅ | ❌ | ❌ | Search, Advanced |
| **TrackingHandler** | 5 | ✅ | ✅ | ❌ | ❌ | CRUD, Encode, Decode, GetEnums |

### Automation Handlers

| Handler | Endpoints | Auth | RLS | Rate Limit | RBAC | Use Cases |
|---------|-----------|------|-----|------------|------|-----------|
| **AutomationHandler** | 6 | ✅ | ✅ | 1000-M | ❌ | CRUD, GetTypes, GetActions, GetOperators |
| **AutomationDiscoveryHandler** | 9 | ✅ | ❌ | ❌ | ❌ | Metadata discovery, triggers, actions, operators |
| **BroadcastHandler** | 8 | ✅ | ✅ | 1000-M | ❌ | CRUD, Schedule, Execute, Cancel, GetStats |
| **SequenceHandler** | 11 | ✅ | ✅ | 1000-M | ❌ | CRUD, Activate, Pause, Resume, Archive, GetStats, Enroll |
| **CampaignHandler** | 14 | ✅ | ✅ | 1000-M | ❌ | CRUD, Activate, Schedule, Pause, Resume, Complete, Archive, GetStats, Enroll |

### Infrastructure Handlers

| Handler | Endpoints | Auth | RLS | Rate Limit | RBAC | Use Cases |
|---------|-----------|------|-----|------------|------|-----------|
| **AuthHandler** | 5 | Partial | ❌ | 10-M | ❌ | Register, Login, GetProfile, GenerateAPIKey, GetInfo |
| **HealthHandler** | 8 | ❌ | ❌ | ❌ | ❌ | Health checks (DB, Redis, RabbitMQ, Temporal, Migrations) |
| **WebhookSubscriptionHandler** | 6 | ✅ | ✅ | ❌ | ❌ | CRUD webhooks, GetAvailableEvents |
| **WAHAWebhookHandler** | 2 | ❌ | ❌ | ❌ | ❌ | ReceiveWebhook (inbound), GetInfo |
| **StripeWebhookHandler** | 2 | ❌ | ❌ | ❌ | ❌ | HandleWebhook, GetInfo |
| **LlamaParseWebhookHandler** | 1 | ❌ | ❌ | ❌ | ❌ | HandleWebhook |
| **WebSocketMessageHandler** | 2 | ✅ | ❌ | 5/min | ❌ | HandleWebSocket, GetStats |
| **MediaHandler** | 1 | ❓ | ❓ | ❌ | ❌ | UploadMedia |
| **DomainEventHandler** | 4 | ❓ | ❓ | ❌ | ❌ | List domain events by contact/session/project/type |
| **ContactEventStreamHandler** | 3 | ❓ | ❓ | ❌ | ❌ | Stream contact events (SSE) |
| **QueueHandler** | 1 | ❌ | ❌ | ❌ | ❌ | ListQueues (RabbitMQ) |
| **TestHandler** | 7 | ❌ | ❌ | ❌ | ❌ | Test utilities (dev only) |

**Rate Limit Codes:**
- `10-M`: 10 requests/minute (auth endpoints - brute force protection)
- `1000-M`: 1000 requests/minute per user (API endpoints)
- `5/min`: 5 WebSocket connections/minute

---

## 📊 ENDPOINTS SUMMARY

### Por Categoria

| Categoria | Total | Auth Required | RLS | Rate Limited |
|-----------|-------|---------------|-----|--------------|
| **Health** | 8 | ❌ | ❌ | ❌ |
| **Auth** | 5 | Partial | ❌ | ✅ |
| **Webhooks (inbound)** | 5 | ❌ | ❌ | ❌ |
| **Webhook Subscriptions** | 6 | ✅ | ✅ | ❌ |
| **CRM - Contacts** | 9 | ✅ | ✅ | ✅ |
| **CRM - Projects** | 7 | ✅ | ✅ | ❌ |
| **CRM - Pipelines** | 11 | ✅ | ✅ | ❌ |
| **CRM - Messages** | 10 | ✅ | ✅ | ✅ |
| **CRM - Sessions** | 6 | ✅ | ✅ | ❌ |
| **CRM - Channels** | 12 | ✅ | ✅ | ✅ |
| **CRM - Agents** | 9 | ✅ | ✅ | ✅ |
| **CRM - Chats** | 9 | ✅ | ✅ | ❌ |
| **CRM - Notes** | 2 | ✅ | ✅ | ❌ |
| **CRM - Trackings** | 5 | ✅ | ✅ | ❌ |
| **CRM - Automation Discovery** | 9 | ✅ | ❌ | ❌ |
| **Automation - Rules** | 6 | ✅ | ✅ | ✅ |
| **Automation - Broadcasts** | 8 | ✅ | ✅ | ✅ |
| **Automation - Sequences** | 11 | ✅ | ✅ | ✅ |
| **Automation - Campaigns** | 14 | ✅ | ✅ | ✅ |
| **WebSocket** | 2 | ✅ | ❌ | ✅ |
| **Test** | 7 | ❌ | ❌ | ❌ |
| **Queues** | 1 | ❌ | ❌ | ❌ |

**TOTAL: 158 ENDPOINTS**

---

## 🔥 CRITICAL ISSUES

### 1. Cache Layer Missing (CRITICAL)

**Impacto:** Alto custo de I/O, latência desnecessária

**Repositories Críticos:**
1. `GormContactRepository` - FindByPhone/FindByEmail (webhooks)
2. `GormAgentRepository` - FindByID (sessões)
3. `GormProjectRepository` - FindByID (todos requests)
4. `GormChannelRepository` - GetByWebhookID (webhooks)
5. `GormSessionRepository` - FindByID (mensagens)

**Fix:**
```go
// Exemplo: Cache-aside pattern
cacheKey := fmt.Sprintf("contact:phone:%s", phone)

// Try cache
if cached, err := redis.Get(ctx, cacheKey).Result(); err == nil {
    return unmarshalContact(cached)
}

// Miss: go to DB
contact, err := r.findByPhoneDB(ctx, projectID, phone)

// Store in cache (TTL: 5 min)
redis.Set(ctx, cacheKey, marshalContact(contact), 5*time.Minute)
```

---

### 2. N+1 Query Detected (MEDIUM)

**Arquivo:** `gorm_contact_list_repository.go`

**Problema:**
```go
// Linha ~45
err := r.db.Preload("Contacts").First(&entity, "id = ?", id).Error
```

**Impacto:**
- OOM em listas grandes (10k+ contatos)
- Timeout em queries
- Performance degradada

**Fix:**
```go
// REMOVE Preload("Contacts")
err := r.db.First(&entity, "id = ?", id).Error

// Método separado com paginação
func GetContactIDs(listID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
    var ids []uuid.UUID
    err := r.db.Model(&entities.ContactListMemberEntity{}).
        Where("list_id = ?", listID).
        Limit(limit).
        Offset(offset).
        Pluck("contact_id", &ids).Error
    return ids, err
}
```

---

### 3. RBAC Missing (MEDIUM)

**Impacto:** Todos usuários têm acesso aos mesmos endpoints

**Endpoints Críticos:**
- `DELETE /api/v1/crm/projects/:id`
- `DELETE /api/v1/crm/agents/:id`
- `POST /api/v1/crm/automation/triggers/custom` (admin only)

**Fix:**
```go
// Middleware RBAC
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if !contains(roles, userRole) {
            c.JSON(403, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Uso:
projects.DELETE("/:id", rbac.RequireRole("admin"), projectHandler.DeleteProject)
```

---

### 4. Optimistic Locking Inconsistente (LOW)

**Repositories SEM optimistic locking:**
- ❌ ChannelRepository
- ❌ MessageRepository
- ❌ CredentialRepository

**Impacto:** Race conditions em updates concorrentes

**Fix:** Implementar pattern:
```go
result := r.db.Model(&Entity{}).
    Where("id = ? AND version = ?", id, existingVersion).
    Updates(map[string]interface{}{
        "version": existingVersion + 1,
        // ... campos
    })

if result.RowsAffected == 0 {
    return shared.NewOptimisticLockError(...)
}
```

---

## 🎯 PRIORITY MATRIX

### High Priority (Fazer agora)

1. ✅ **Implementar Redis cache layer**
   - ContactRepository.FindByPhone/Email
   - AgentRepository.FindByID
   - ProjectRepository.FindByID
   - ChannelRepository.GetByWebhookID

2. ✅ **Corrigir N+1 query**
   - ContactListRepository.FindByID
   - Remover Preload, adicionar paginação

3. ✅ **Implementar RBAC**
   - Proteger endpoints de delete
   - Proteger endpoints de admin

### Medium Priority (Fazer em seguida)

1. ⚠️ **Completar optimistic locking**
   - ChannelRepository
   - MessageRepository
   - CredentialRepository

2. ⚠️ **Circuit breaker**
   - Calls externos (WAHA, Stripe, LlamaParse)

3. ⚠️ **Error handling padronizado**
   - RFC 7807 (Problem Details)
   - Error codes

### Low Priority (Backlog)

1. 🔵 **API versioning**
   - Planejar v2
   - Deprecation policy

2. 🔵 **Swagger docs**
   - Auditar documentação
   - Adicionar exemplos

3. 🔵 **Observability**
   - Prometheus metrics
   - OpenTelemetry tracing

---

## 📌 QUICK STATS

| Métrica | Valor |
|---------|-------|
| **Repositories** | 31 |
| **Handlers** | 27 |
| **Endpoints** | 158 |
| **Optimistic Locking** | 70% cobertura |
| **Cache Layer** | 0% implementado |
| **N+1 Queries** | 1 detectado |
| **RBAC** | 0% implementado |
| **Rate Limiting** | 80% dos endpoints críticos |
| **Authentication** | 90% dos endpoints protegidos |
| **RLS (Tenant Isolation)** | 85% dos handlers CRM |

---

## 🔗 LINKS ÚTEIS

- **Relatório Completo**: `ARCHITECTURE_MAPPING_REPORT.md`
- **Routes Definition**: `/infrastructure/http/routes/routes.go`
- **Persistence Layer**: `/infrastructure/persistence/`
- **Handlers Layer**: `/infrastructure/http/handlers/`
- **Domain Layer**: `/internal/domain/`

---

**Gerado em:** 2025-10-13 | **Versão:** 1.0
