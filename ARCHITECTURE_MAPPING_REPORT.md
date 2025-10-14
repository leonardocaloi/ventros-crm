# Ventros CRM - Architecture Mapping Report

**Data**: 2025-10-13
**Versão**: 1.0
**Objetivo**: Mapeamento completo de repositories, handlers e endpoints do sistema

---

## Sumário Executivo

- **Repositories Mapeados**: 28 (incluindo sub-repositories de broadcast/campaign/sequence)
- **Handlers Mapeados**: 25
- **Endpoints Totais**: 120+ endpoints funcionais
- **Padrão de Arquitetura**: Hexagonal (DDD + Clean Architecture)
- **ORM**: GORM com suporte a Optimistic Locking
- **Cache Layer**: ❌ **NÃO IMPLEMENTADO** (risco de performance)
- **N+1 Queries**: 🔴 **DETECTADO EM 1 REPOSITORY** (gorm_contact_list_repository.go)

---

## PARTE 1: REPOSITORIES

### 1.1 Lista Completa de Repositories

| # | Repository | Entity Domain | Arquivo |
|---|-----------|---------------|---------|
| 1 | `GormAgentRepository` | `agent.Agent` | `gorm_agent_repository.go` |
| 2 | `GormAutomationRuleRepository` | `pipeline.Automation` | `gorm_automation_repository.go` |
| 3 | `GormBillingRepository` | `billing.BillingAccount` | `gorm_billing_repository.go` |
| 4 | `GormBroadcastRepository` | `broadcast.Broadcast` | `gorm_broadcast_repository.go` |
| 5 | `GormBroadcastExecutionRepository` | `broadcast.BroadcastExecution` | `gorm_broadcast_repository.go` |
| 6 | `GormCampaignRepository` | `campaign.Campaign` | `gorm_campaign_repository.go` |
| 7 | `GormCampaignEnrollmentRepository` | `campaign.CampaignEnrollment` | `gorm_campaign_repository.go` |
| 8 | `GormChannelRepository` | `channel.Channel` | `gorm_channel_repository.go` |
| 9 | `GormChannelTypeRepository` | `channel.ChannelType` | `gorm_channel_type_repository.go` |
| 10 | `GormChatRepository` | `chat.Chat` | `gorm_chat_repository.go` |
| 11 | `GormContactRepository` | `contact.Contact` | `gorm_contact_repository.go` |
| 12 | `GormContactEventRepository` | `contact.ContactEvent` | `gorm_contact_event_repository.go` |
| 13 | `GormContactListRepository` | `contact_list.ContactList` | `gorm_contact_list_repository.go` |
| 14 | `GormCredentialRepository` | `credential.Credential` | `gorm_credential_repository.go` |
| 15 | `GormDomainEventLogRepository` | `event.DomainEvent` | `gorm_domain_event_log_repository.go` |
| 16 | `GormInvoiceRepository` | `billing.Invoice` | `gorm_invoice_repository.go` |
| 17 | `GormMessageRepository` | `message.Message` | `gorm_message_repository.go` |
| 18 | `GormMessageEnrichmentRepository` | `message.MessageEnrichment` | `gorm_message_enrichment_repository.go` |
| 19 | `GormMessageGroupRepository` | `message_group.MessageGroup` | `gorm_message_group_repository.go` |
| 20 | `GormNoteRepository` | `note.Note` | `gorm_note_repository.go` |
| 21 | `GormOutboxRepository` | `outbox.OutboxMessage` | `gorm_outbox_repository.go` |
| 22 | `GormPipelineRepository` | `pipeline.Pipeline` | `gorm_pipeline_repository.go` |
| 23 | `GormProjectRepository` | `project.Project` | `gorm_project_repository.go` |
| 24 | `GormProjectMemberRepository` | `project_member.ProjectMember` | `gorm_project_member_repository.go` |
| 25 | `GormSequenceRepository` | `sequence.Sequence` | `gorm_sequence_repository.go` |
| 26 | `GormSequenceEnrollmentRepository` | `sequence.SequenceEnrollment` | `gorm_sequence_repository.go` |
| 27 | `GormSessionRepository` | `session.Session` | `gorm_session_repository.go` |
| 28 | `GormSubscriptionRepository` | `billing.Subscription` | `gorm_subscription_repository.go` |
| 29 | `GormTrackingRepository` | `tracking.Tracking` | `gorm_tracking_repository.go` |
| 30 | `GormUsageMeterRepository` | `billing.UsageMeter` | `gorm_usage_meter_repository.go` |
| 31 | `GormWebhookRepository` | `webhook.Webhook` | `gorm_webhook_repository.go` |

**Total: 31 Repositories**

---

### 1.2 Análise Detalhada dos Repositories

#### **GormAgentRepository** (`gorm_agent_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.AgentEntity` → `agent.Agent` |
| **Métodos** | `Save`, `FindByID`, `FindByEmail`, `FindByTenant`, `FindActiveByTenant`, `Delete`, `FindByTenantWithFilters`, `SearchByText` |
| **Optimistic Locking** | ✅ **SIM** (campo `version`) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `domainToEntity`, `entityToDomain` |
| **Custom Fields** | Suporte a `VirtualAgentMetadata` (JSONB) |
| **Validações** | Previne modificação/deleção de agentes de sistema (`IsSystem()`) |

**Métodos Especiais:**
- `FindByTenantWithFilters`: Suporte a filtros avançados (type, status, active, pagination, sorting)
- `SearchByText`: Busca por nome e email com ILIKE
- Proteção contra modificação de `system_agents`

---

#### **GormContactRepository** (`gorm_contact_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.ContactEntity` → `contact.Contact` |
| **Métodos** | `Save`, `FindByID`, `FindByProject`, `CountByProject`, `FindByExternalID`, `FindByPhone`, `FindByEmail`, `FindByTenantWithFilters`, `SearchByText`, `SaveCustomFields`, `GetCustomFields`, `FindByCustomField` |
| **Optimistic Locking** | ✅ **SIM** (campo `version`) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `domainToEntity`, `entityToDomain` |
| **Custom Fields** | ✅ **SIM** (tabela `contact_custom_fields` via JSONB) |
| **Transaction Support** | ✅ **SIM** (`getDB(ctx)` suporta transações do contexto) |

**Métodos Especiais:**
- `SaveCustomFields`: Batch upsert de custom fields com ON CONFLICT
- `FindByCustomField`: Busca por custom field key-value
- `GetCustomFields`: Retorna todos custom fields de um contato
- `SearchByText`: Busca full-text com relevance scoring (CASE WHEN)
- Suporte a soft delete (campo `deleted_at`)

---

#### **GormMessageRepository** (`gorm_message_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.MessageEntity` → `message.Message` |
| **Métodos** | `Save`, `FindByID`, `FindBySession`, `FindByContact`, `FindByChannelMessageID`, `CountBySession`, `FindByTenantWithFilters`, `SearchByText`, `FindBySessionIDForEnrichment` |
| **Optimistic Locking** | ❌ **NÃO** (usa `Save` sem version check) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `domainToEntity`, `entityToDomain` |
| **Custom Fields** | Metadata (JSONB), Mentions (Array) |
| **ACK Tracking** | ✅ **SIM** (`delivered_at`, `read_at`, `played_at`) |

**Métodos Especiais:**
- `FindByTenantWithFilters`: 12 filtros avançados (contact, session, channel, content_type, status, timestamp range, has_media, agent)
- `SearchByText`: Busca ILIKE em campo `text`
- `FindBySessionIDForEnrichment`: Retorna informações simplificadas (ID, ChannelID, Direction, Timestamp)
- Suporte a `message.Source` para rastreamento de origem (Manual, System, AI Agent, External)

---

#### **GormProjectRepository** (`gorm_project_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.ProjectEntity` → `project.Project` |
| **Métodos** | `Save`, `FindByID`, `FindByTenantID`, `FindByCustomer`, `FindActiveProjects`, `Delete`, `FindByTenantWithFilters`, `SearchByText` |
| **Optimistic Locking** | ✅ **SIM** (campo `version`) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `domainToEntity`, `entityToDomain` |
| **Configuration** | JSONB field para project settings |
| **Session Timeout** | Campo `session_timeout_minutes` |

---

#### **GormChannelRepository** (`gorm_channel_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.ChannelEntity` → `channel.Channel` |
| **Métodos** | `Create`, `GetByID`, `GetByUserID`, `GetByProjectID`, `GetByExternalID`, `GetByWebhookID`, `GetActiveWAHAChannels`, `Update`, `Delete` |
| **Optimistic Locking** | ❌ **NÃO** (usa `Save` sem version check) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `toEntity`, `toDomain` |
| **Features** | AI enabled, AI agents, tracking, message debouncer (timeout_ms) |
| **Statistics** | messages_received, messages_sent, last_message_at, last_error |

---

#### **GormChatRepository** (`gorm_chat_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.ChatEntity` → `chat.Chat` |
| **Métodos** | `Create`, `FindByID`, `FindByExternalID`, `FindByProject`, `FindByTenant`, `FindByContact`, `FindActiveByProject`, `FindIndividualByContact`, `Update`, `Delete`, `SearchBySubject` |
| **Optimistic Locking** | ✅ **SIM** (campo `version`) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `domainToEntity`, `entityToDomain` |
| **Participants** | JSONB array com suporte a query (`@>` operator) |
| **Chat Types** | Individual, Group |

**Métodos Especiais:**
- `FindByContact`: Usa JSONB query `participants @> [{"id":"uuid"}]`
- `FindIndividualByContact`: Busca chat 1:1 com contato
- `SearchBySubject`: Busca ILIKE por assunto

---

#### **GormCampaignRepository** (`gorm_campaign_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.CampaignEntity` → `campaign.Campaign` |
| **Métodos** | `Save`, `FindByID`, `FindByTenantID`, `FindActiveByStatus`, `FindScheduled`, `Delete` |
| **Optimistic Locking** | ✅ **SIM** (campo `version`) com **TRANSACTION** |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ⚠️ **POTENCIAL** (carrega steps em loop no `toDomainSlice`) |
| **Mappers** | `toEntity`, `toDomain`, `stepsToEntities` |
| **Nested Entities** | `CampaignStepEntity` (relação 1:N) |
| **Transaction Support** | ✅ **SIM** (insert/update campaign + steps atomicamente) |

**Métodos Especiais:**
- `Save`: Usa transaction para atualizar campaign + deletar steps antigos + inserir novos
- `FindScheduled`: Busca campanhas agendadas prontas para iniciar (`status = scheduled AND start_date <= NOW()`)
- `Delete`: Deleta steps primeiro, depois campaign (cascade manual)

**Sub-Repository: GormCampaignEnrollmentRepository**

| Métodos | `Save`, `FindByID`, `FindByCampaignID`, `FindByContactID`, `FindReadyForNextStep`, `FindActiveByCampaignAndContact`, `Delete` |
|---------|----------|
| **Conversores** | `enrollmentToEntity`, `enrollmentToDomain` |

---

#### **GormBroadcastRepository** (`gorm_broadcast_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.BroadcastEntity` → `broadcast.Broadcast` |
| **Métodos** | `Save`, `FindByID`, `FindByTenantID`, `FindScheduledReady`, `FindByStatus`, `Delete` |
| **Optimistic Locking** | ✅ **SIM** (campo `version`) |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | ✅ **NÃO DETECTADO** |
| **Mappers** | `toEntity`, `toDomain` |
| **Message Template** | JSONB field |
| **Statistics** | total_contacts, sent_count, failed_count, pending_count |

**Sub-Repository: GormBroadcastExecutionRepository**

| Métodos | `Save`, `SaveBatch`, `FindByID`, `FindByBroadcastID`, `FindPendingByBroadcastID`, `Delete` |
|---------|----------|
| **Batch Support** | ✅ **SIM** (`SaveBatch` para inserção em massa) |

---

#### **GormContactListRepository** (`gorm_contact_list_repository.go`)

| Aspecto | Detalhes |
|---------|----------|
| **Entity** | `entities.ContactListEntity` → `contact_list.ContactList` |
| **Métodos** | `Save`, `FindByID`, `FindByTenantID`, `Delete`, `AddContacts`, `RemoveContacts`, `GetContactIDs`, `GetContactCount`, `IsContactInList` |
| **Optimistic Locking** | ❌ **NÃO** |
| **Cache Layer** | ❌ **NÃO** |
| **N+1 Queries** | 🔴 **SIM** (usa `Preload("Contacts")` que pode causar N+1 em listas grandes) |
| **Mappers** | `toEntity`, `toDomain` |
| **List Types** | Static, Dynamic (com filter rules em JSONB) |

**🔴 PROBLEMA DETECTADO:**
```go
// gorm_contact_list_repository.go - linha ~45
err := r.db.Preload("Contacts").First(&entity, "id = ?", id).Error
```
**Impacto**: Carregar todos os contatos da lista pode causar problemas de memória e performance em listas grandes (10k+ contatos)

**Recomendação**: Implementar paginação ou remover Preload e usar queries separadas

---

### 1.3 Repositories sem Cache Layer ⚠️

**TODOS** os 31 repositories não possuem cache layer implementado.

**Impacto:**
- ❌ Todas as queries vão direto ao PostgreSQL
- ❌ Dados frequentemente acessados (agentes, projetos, canais) são buscados repetidamente
- ❌ Falta de cache para sessões ativas (alto custo de I/O)
- ❌ Contact lookups por phone/email não são cacheados (crítico para webhooks WAHA)

**Recomendação:**
Implementar Redis cache layer para:
1. **Prioridade ALTA**: `GormAgentRepository`, `GormProjectRepository`, `GormChannelRepository`
2. **Prioridade MÉDIA**: `GormContactRepository` (cache por phone/email), `GormSessionRepository`
3. **Prioridade BAIXA**: `GormPipelineRepository`, `GormWebhookRepository`

**Padrão sugerido:**
```go
// Exemplo: Cache-aside pattern
func (r *GormContactRepository) FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*contact.Contact, error) {
    cacheKey := fmt.Sprintf("contact:project:%s:phone:%s", projectID, phone)

    // Try cache first
    if cached, err := r.redis.Get(ctx, cacheKey).Result(); err == nil {
        var contact contact.Contact
        json.Unmarshal([]byte(cached), &contact)
        return &contact, nil
    }

    // Miss: go to database
    contact, err := r.findByPhoneDB(ctx, projectID, phone)
    if err != nil {
        return nil, err
    }

    // Store in cache (TTL: 5 minutes)
    json, _ := json.Marshal(contact)
    r.redis.Set(ctx, cacheKey, json, 5*time.Minute)

    return contact, nil
}
```

---

### 1.4 Optimistic Locking Summary

| Repository | Optimistic Locking | Version Field |
|-----------|-------------------|---------------|
| AgentRepository | ✅ | `version INT` |
| BillingRepository | ✅ | `version INT` |
| BroadcastRepository | ✅ | `version INT` |
| CampaignRepository | ✅ | `version INT` + Transaction |
| ChatRepository | ✅ | `version INT` |
| ContactRepository | ✅ | `version INT` |
| ProjectRepository | ✅ | `version INT` |
| ChannelRepository | ❌ | - |
| MessageRepository | ❌ | - |
| SequenceRepository | ❓ (não analisado) | - |

**Padrão implementado:**
```go
result := r.db.Model(&entities.EntityName{}).
    Where("id = ? AND version = ?", entity.ID, existing.Version).
    Updates(map[string]interface{}{
        "version": existing.Version + 1,
        // ... outros campos
    })

if result.RowsAffected == 0 {
    return shared.NewOptimisticLockError("Entity", id, existing.Version, entity.Version)
}
```

---

## PARTE 2: HTTP HANDLERS

### 2.1 Lista Completa de Handlers

| # | Handler | Aggregate/Entity | Arquivo |
|---|---------|------------------|---------|
| 1 | `AgentHandler` | `agent.Agent` | `agent_handler.go` |
| 2 | `AuthHandler` | `user.User` | `auth_handler.go` |
| 3 | `AutomationHandler` | `pipeline.Automation` | `automation_handler.go` |
| 4 | `AutomationDiscoveryHandler` | Metadata | `automation_discovery_handler.go` |
| 5 | `BroadcastHandler` | `broadcast.Broadcast` | `broadcast_handler.go` |
| 6 | `CampaignHandler` | `campaign.Campaign` | `campaign_handler.go` |
| 7 | `ChannelHandler` | `channel.Channel` | `channel_handler.go` |
| 8 | `ChatHandler` | `chat.Chat` | `chat_handler.go` |
| 9 | `ContactHandler` | `contact.Contact` | `contact_handler.go` |
| 10 | `ContactEventStreamHandler` | `contact.ContactEvent` | `contact_event_stream_handler.go` |
| 11 | `DomainEventHandler` | `event.DomainEvent` | `domain_event_handler.go` |
| 12 | `HealthHandler` | Infra | `health.go` |
| 13 | `LlamaParseWebhookHandler` | Webhook | `llamaparse_webhook_handler.go` |
| 14 | `MediaHandler` | Storage | `media_handler.go` |
| 15 | `MessageHandler` | `message.Message` | `message_handler.go` |
| 16 | `NoteHandler` | `note.Note` | `note_handler.go` |
| 17 | `PipelineHandler` | `pipeline.Pipeline` | `pipeline_handler.go` |
| 18 | `ProjectHandler` | `project.Project` | `project_handler.go` |
| 19 | `QueueHandler` | Infra | `queue_handler.go` |
| 20 | `SequenceHandler` | `sequence.Sequence` | `sequence_handler.go` |
| 21 | `SessionHandler` | `session.Session` | `session_handler.go` |
| 22 | `StripeWebhookHandler` | Webhook | `stripe_webhook_handler.go` |
| 23 | `TestHandler` | Testing | `test_handler.go` |
| 24 | `TrackingHandler` | `tracking.Tracking` | `tracking_handler.go` |
| 25 | `WAHAWebhookHandler` | Webhook | `waha_webhook_handler.go` |
| 26 | `WebhookSubscriptionHandler` | `webhook.Webhook` | `webhook_subscription.go` |
| 27 | `WebSocketMessageHandler` | Real-time | `websocket_message_handler.go` |

**Total: 27 Handlers**

---

### 2.2 Análise Detalhada dos Handlers

#### **AgentHandler** (`agent_handler.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 9 endpoints |
| **Use Cases** | `CreateAgent`, `GetAgent`, `UpdateAgent`, `DeleteAgent`, `ListAgents`, `CreateVirtualAgent`, `GetAgentStats`, `EndVirtualAgentPeriod`, `SearchAgents`, `ListAgentsAdvanced` |
| **DTOs** | Request/Response não especificados no código analisado |
| **Validação** | ❓ (não verificado) |
| **Error Handling** | ✅ Presente (uso de `c.JSON` para erros) |
| **Auth** | ✅ SIM (via `authMiddleware.Authenticate()`) |
| **RLS** | ✅ SIM (via `rlsMiddleware.SetUserContext()`) |
| **Rate Limiting** | ✅ SIM (`1000-M` = 1000 req/min por usuário) |

**Endpoints:**
- `GET /api/v1/crm/agents` - ListAgents
- `GET /api/v1/crm/agents/search` - SearchAgents
- `GET /api/v1/crm/agents/advanced` - ListAgentsAdvanced
- `POST /api/v1/crm/agents` - CreateAgent
- `POST /api/v1/crm/agents/virtual` - CreateVirtualAgent
- `GET /api/v1/crm/agents/:id` - GetAgent
- `PUT /api/v1/crm/agents/:id` - UpdateAgent
- `DELETE /api/v1/crm/agents/:id` - DeleteAgent
- `GET /api/v1/crm/agents/:id/stats` - GetAgentStats
- `PUT /api/v1/crm/agents/:id/virtual/end-period` - EndVirtualAgentPeriod

---

#### **ContactHandler** (`contact_handler.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 9 endpoints |
| **Use Cases** | `CreateContact`, `GetContact`, `UpdateContact`, `DeleteContact`, `ListContacts`, `SearchContacts`, `ListContactsAdvanced`, `ChangePipelineStatus` |
| **DTOs** | Request/Response não especificados |
| **Validação** | ✅ Presente (binding de JSON) |
| **Error Handling** | ✅ Presente |
| **Auth** | ✅ SIM |
| **RLS** | ✅ SIM |

**Endpoints:**
- `GET /api/v1/contacts` - ListContacts
- `GET /api/v1/contacts/search` - SearchContacts
- `GET /api/v1/contacts/advanced` - ListContactsAdvanced
- `POST /api/v1/contacts` - CreateContact
- `GET /api/v1/contacts/:id` - GetContact
- `PUT /api/v1/contacts/:id` - UpdateContact
- `DELETE /api/v1/contacts/:id` - DeleteContact
- `PUT /api/v1/contacts/:id/pipelines/:pipeline_id/status` - ChangePipelineStatus
- `GET /api/v1/contacts/:contact_id/trackings` - GetContactTrackings (nested)

---

#### **MessageHandler** (`message_handler.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 10 endpoints |
| **Use Cases** | `SendMessage`, `ConfirmMessageDelivery`, `ListMessages`, `SearchMessages`, `CreateMessage`, `GetMessage`, `UpdateMessage`, `DeleteMessage`, `GetMessagesBySession`, `ListMessagesAdvanced` |
| **DTOs** | Request/Response para SendMessage |
| **Validação** | ✅ Presente |
| **Error Handling** | ✅ Robusto |
| **Auth** | ✅ SIM |
| **RLS** | ✅ SIM |

**Endpoints:**
- `GET /api/v1/crm/messages` - ListMessages
- `GET /api/v1/crm/messages/search` - SearchMessages
- `GET /api/v1/crm/messages/advanced` - ListMessagesAdvanced
- `POST /api/v1/crm/messages` - CreateMessage
- `POST /api/v1/crm/messages/send` - SendMessage (comando principal)
- `POST /api/v1/crm/messages/confirm-delivery` - ConfirmMessageDelivery
- `GET /api/v1/crm/messages/:id` - GetMessage
- `PUT /api/v1/crm/messages/:id` - UpdateMessage
- `DELETE /api/v1/crm/messages/:id` - DeleteMessage

---

#### **CampaignHandler** (`campaign_handler.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 14 endpoints |
| **Use Cases** | CRUD + lifecycle (activate, schedule, pause, resume, complete, archive) + enrollments |
| **DTOs** | Request/Response para campaign management |
| **Validação** | ✅ Presente |
| **Error Handling** | ✅ Robusto |
| **Auth** | ✅ SIM |
| **RLS** | ✅ SIM |

**Endpoints:**
- `GET /api/v1/automation/campaigns` - ListCampaigns
- `POST /api/v1/automation/campaigns` - CreateCampaign
- `GET /api/v1/automation/campaigns/:id` - GetCampaign
- `PUT /api/v1/automation/campaigns/:id` - UpdateCampaign
- `DELETE /api/v1/automation/campaigns/:id` - DeleteCampaign
- `POST /api/v1/automation/campaigns/:id/activate` - ActivateCampaign
- `POST /api/v1/automation/campaigns/:id/schedule` - ScheduleCampaign
- `POST /api/v1/automation/campaigns/:id/pause` - PauseCampaign
- `POST /api/v1/automation/campaigns/:id/resume` - ResumeCampaign
- `POST /api/v1/automation/campaigns/:id/complete` - CompleteCampaign
- `POST /api/v1/automation/campaigns/:id/archive` - ArchiveCampaign
- `GET /api/v1/automation/campaigns/:id/stats` - GetCampaignStats
- `POST /api/v1/automation/campaigns/:id/enroll` - EnrollContact
- `GET /api/v1/automation/campaigns/:id/enrollments` - ListEnrollments

---

#### **PipelineHandler** (`pipeline_handler.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 11 endpoints |
| **Use Cases** | CRUD pipelines + status management + custom fields |
| **DTOs** | Request/Response para pipelines |
| **Validação** | ✅ Presente |
| **Error Handling** | ✅ Robusto |
| **Auth** | ✅ SIM |
| **RLS** | ✅ SIM |

**Endpoints:**
- `GET /api/v1/crm/pipelines` - ListPipelines
- `GET /api/v1/crm/pipelines/search` - SearchPipelines
- `GET /api/v1/crm/pipelines/advanced` - ListPipelinesAdvanced
- `POST /api/v1/crm/pipelines` - CreatePipeline
- `GET /api/v1/crm/pipelines/:id` - GetPipeline
- `POST /api/v1/crm/pipelines/:id/statuses` - CreateStatus
- `PUT /api/v1/crm/pipelines/:id/contacts/:contact_id/status` - ChangeContactStatus
- `GET /api/v1/crm/pipelines/:id/contacts/:contact_id/status` - GetContactStatus
- `POST /api/v1/crm/pipelines/:id/custom-fields` - SetCustomField (NEW)
- `GET /api/v1/crm/pipelines/:id/custom-fields` - GetCustomFields (NEW)
- `DELETE /api/v1/crm/pipelines/:id/custom-fields/:field_key` - RemoveCustomField (NEW)

---

#### **HealthHandler** (`health.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 8 endpoints |
| **Use Cases** | Health checks de infraestrutura |
| **Validação** | N/A |
| **Error Handling** | ✅ Robusto |
| **Auth** | ❌ NÃO (public endpoints) |
| **RLS** | ❌ NÃO |

**Endpoints:**
- `GET /health` - Health (aggregate health check)
- `GET /ready` - Ready (readiness probe - Kubernetes)
- `GET /live` - Live (liveness probe - Kubernetes)
- `GET /health/database` - CheckDatabase
- `GET /health/migrations` - CheckMigrations
- `GET /health/redis` - CheckRedis
- `GET /health/rabbitmq` - CheckRabbitMQ
- `GET /health/temporal` - CheckTemporal

---

#### **WAHAWebhookHandler** (`waha_webhook_handler.go`)

| Aspecto | Valor |
|---------|-------|
| **Endpoints** | 2 endpoints |
| **Use Cases** | Receber webhooks do WAHA (WhatsApp HTTP API) |
| **Validação** | ✅ Webhook signature validation |
| **Error Handling** | ✅ Robusto |
| **Auth** | ❌ NÃO (webhook inbound não usa auth de usuário) |
| **RLS** | ❌ NÃO |

**Endpoints:**
- `POST /api/v1/webhooks/:webhook_id` - ReceiveWebhook (inbound)
- `GET /api/v1/webhooks/info` - GetWebhookInfo

**IMPORTANTE:** Estes endpoints são expostos publicamente para receber webhooks de serviços externos (WAHA). A autenticação é feita via `webhook_id` único por canal.

---

### 2.3 Handlers Security Summary

| Handler | Authentication | RLS | Rate Limiting | RBAC |
|---------|---------------|-----|---------------|------|
| AgentHandler | ✅ | ✅ | ✅ 1000-M | ❌ |
| AuthHandler | ⚠️ Partial | ❌ | ✅ 10-M | ❌ |
| ContactHandler | ✅ | ✅ | ✅ 1000-M | ❌ |
| MessageHandler | ✅ | ✅ | ✅ 1000-M | ❌ |
| CampaignHandler | ✅ | ✅ | ✅ 1000-M | ❌ |
| PipelineHandler | ✅ | ✅ | ✅ 1000-M | ❌ |
| ProjectHandler | ✅ | ✅ | ❌ | ❌ |
| SessionHandler | ✅ | ✅ | ❌ | ❌ |
| WAHAWebhookHandler | ❌ | ❌ | ❌ | ❌ |
| StripeWebhookHandler | ❌ | ❌ | ❌ | ❌ |
| HealthHandler | ❌ | ❌ | ❌ | ❌ |

**Legenda:**
- ✅ **SIM**: Implementado
- ❌ **NÃO**: Não implementado
- ⚠️ **Partial**: Implementado parcialmente (alguns endpoints públicos, outros protegidos)

**Rate Limiting:**
- `10-M`: 10 requests por minuto (auth endpoints - brute force protection)
- `1000-M`: 1000 requests por minuto por usuário (API endpoints)

---

## PARTE 3: API ROUTES

### 3.1 Estrutura de Rotas

**Versões:**
- `v1`: API principal (único versionamento implementado)

**Grupos de Rotas:**

#### 1. **Health Checks** (Public)
```
GET    /health
GET    /ready
GET    /live
GET    /health/database
GET    /health/migrations
GET    /health/redis
GET    /health/rabbitmq
GET    /health/temporal
```

#### 2. **Auth** (`/api/v1/auth`) - Rate Limited (10 req/min)
```
POST   /register
POST   /login
GET    /info
GET    /profile             (Auth Required)
POST   /api-key             (Auth Required)
```

#### 3. **Webhooks Inbound** (`/api/v1/webhooks`) - Public
```
POST   /:webhook_id         (WAHA webhook receiver)
GET    /info
```

#### 4. **Webhook Subscriptions** (`/api/v1/webhook-subscriptions`) - Auth + RLS
```
GET    /available-events
POST   /
GET    /
GET    /:id
PUT    /:id
DELETE /:id
```

#### 5. **CRM - Contacts** (`/api/v1/contacts` ou `/api/v1/crm/contacts`) - Auth + RLS + Rate Limited
```
GET    /                    (ListContacts)
GET    /search              (SearchContacts)
GET    /advanced            (ListContactsAdvanced)
POST   /                    (CreateContact)
GET    /:id                 (GetContact)
PUT    /:id                 (UpdateContact)
DELETE /:id                 (DeleteContact)
GET    /:id/sessions        (ListSessions - nested)
GET    /:id/sessions/:session_id (GetSession - nested)
PUT    /:id/pipelines/:pipeline_id/status (ChangePipelineStatus)
GET    /:contact_id/trackings (GetContactTrackings)
```

#### 6. **CRM - Projects** (`/api/v1/crm/projects`) - Auth + RLS
```
GET    /                    (ListProjects)
GET    /search              (SearchProjects)
GET    /advanced            (ListProjectsAdvanced)
POST   /                    (CreateProject)
GET    /:id                 (GetProject)
PUT    /:id                 (UpdateProject)
DELETE /:id                 (DeleteProject)
```

#### 7. **CRM - Pipelines** (`/api/v1/crm/pipelines`) - Auth + RLS
```
GET    /                    (ListPipelines)
GET    /search              (SearchPipelines)
GET    /advanced            (ListPipelinesAdvanced)
POST   /                    (CreatePipeline)
GET    /:id                 (GetPipeline)
POST   /:id/statuses        (CreateStatus)
PUT    /:id/contacts/:contact_id/status (ChangeContactStatus)
GET    /:id/contacts/:contact_id/status (GetContactStatus)
POST   /:id/custom-fields   (SetCustomField)
GET    /:id/custom-fields   (GetCustomFields)
DELETE /:id/custom-fields/:field_key (RemoveCustomField)
```

#### 8. **CRM - Messages** (`/api/v1/crm/messages`) - Auth + RLS
```
GET    /                    (ListMessages)
GET    /search              (SearchMessages)
GET    /advanced            (ListMessagesAdvanced)
POST   /                    (CreateMessage)
POST   /send                (SendMessage - primary command)
POST   /confirm-delivery    (ConfirmMessageDelivery)
GET    /:id                 (GetMessage)
PUT    /:id                 (UpdateMessage)
DELETE /:id                 (DeleteMessage)
```

#### 9. **CRM - Sessions** (`/api/v1/crm/sessions`) - Auth + RLS
```
GET    /                    (ListSessions - requires filters)
GET    /search              (SearchSessions)
GET    /advanced            (ListSessionsAdvanced)
GET    /:id                 (GetSession)
POST   /:id/close           (CloseSession)
GET    /stats               (GetSessionStats)
```

#### 10. **CRM - Channels** (`/api/v1/crm/channels`) - Auth + RLS + Rate Limited
```
GET    /                    (ListChannels)
POST   /                    (CreateChannel)
GET    /:id                 (GetChannel)
POST   /:id/activate        (ActivateChannel)
POST   /:id/deactivate      (DeactivateChannel)
DELETE /:id                 (DeleteChannel)
GET    /:id/webhook-url     (GetChannelWebhookURL)
POST   /:id/configure-webhook (ConfigureChannelWebhook)
GET    /:id/webhook-info    (GetChannelWebhookInfo)
POST   /:id/activate-waha   (ActivateWAHAChannel)
POST   /:id/import-history  (ImportWAHAHistory)
GET    /:id/sessions        (ListSessions - nested)
GET    /:id/sessions/:session_id (GetSession - nested)
```

#### 11. **CRM - Agents** (`/api/v1/crm/agents`) - Auth + RLS + Rate Limited
```
GET    /                    (ListAgents)
GET    /search              (SearchAgents)
GET    /advanced            (ListAgentsAdvanced)
POST   /                    (CreateAgent)
POST   /virtual             (CreateVirtualAgent)
GET    /:id                 (GetAgent)
PUT    /:id                 (UpdateAgent)
DELETE /:id                 (DeleteAgent)
GET    /:id/stats           (GetAgentStats)
PUT    /:id/virtual/end-period (EndVirtualAgentPeriod)
```

#### 12. **CRM - Chats** (`/api/v1/crm/chats`) - Auth + RLS
```
POST   /                    (CreateChat)
GET    /                    (ListChats)
GET    /:id                 (GetChat)
POST   /:id/participants    (AddParticipant)
DELETE /:id/participants/:participant_id (RemoveParticipant)
POST   /:id/archive         (ArchiveChat)
POST   /:id/unarchive       (UnarchiveChat)
POST   /:id/close           (CloseChat)
PATCH  /:id/subject         (UpdateChatSubject)
```

#### 13. **CRM - Notes** (`/api/v1/crm/notes`) - Auth + RLS
```
GET    /search              (SearchNotes)
GET    /advanced            (ListNotesAdvanced)
```

#### 14. **CRM - Trackings** (`/api/v1/crm/trackings`) - Auth + RLS
```
GET    /enums               (GetTrackingEnums)
POST   /encode              (EncodeTracking)
POST   /decode              (DecodeTracking)
POST   /                    (CreateTracking)
GET    /:id                 (GetTracking)
```

#### 15. **CRM - Automation Discovery** (`/api/v1/crm/automation`) - Auth Only
```
GET    /types               (GetAutomationTypes)
GET    /triggers            (GetTriggers)
GET    /triggers/:code      (GetTriggerDetails)
GET    /actions             (GetActions)
GET    /conditions/operators (GetConditionOperators)
GET    /logic-operators     (GetLogicOperators)
GET    /discovery           (GetFullDiscovery)
POST   /triggers/custom     (RegisterCustomTrigger - Admin)
DELETE /triggers/custom/:code (UnregisterCustomTrigger - Admin)
```

#### 16. **Automation Product** (`/api/v1/automation`) - Auth + RLS + Rate Limited

##### 16.1 **Automations** (Pipeline Rules)
```
GET    /types               (GetAutomationTypes - discovery)
GET    /actions             (GetAvailableActions - discovery)
GET    /operators           (GetAvailableOperators - discovery)
GET    /                    (ListAutomations)
POST   /                    (CreateAutomation)
GET    /:id                 (GetAutomation)
PUT    /:id                 (UpdateAutomation)
DELETE /:id                 (DeleteAutomation)
```

##### 16.2 **Broadcasts** (`/automation/broadcasts`)
```
GET    /                    (ListBroadcasts)
POST   /                    (CreateBroadcast)
GET    /:id                 (GetBroadcast)
PUT    /:id                 (UpdateBroadcast)
DELETE /:id                 (DeleteBroadcast)
POST   /:id/schedule        (ScheduleBroadcast)
POST   /:id/execute         (ExecuteBroadcast)
POST   /:id/cancel          (CancelBroadcast)
GET    /:id/stats           (GetBroadcastStats)
```

##### 16.3 **Sequences** (`/automation/sequences`)
```
GET    /                    (ListSequences)
POST   /                    (CreateSequence)
GET    /:id                 (GetSequence)
PUT    /:id                 (UpdateSequence)
DELETE /:id                 (DeleteSequence)
POST   /:id/activate        (ActivateSequence)
POST   /:id/pause           (PauseSequence)
POST   /:id/resume          (ResumeSequence)
POST   /:id/archive         (ArchiveSequence)
GET    /:id/stats           (GetSequenceStats)
POST   /:id/enroll          (EnrollContact)
GET    /:id/enrollments     (ListEnrollments)
```

##### 16.4 **Campaigns** (`/automation/campaigns`)
```
GET    /                    (ListCampaigns)
POST   /                    (CreateCampaign)
GET    /:id                 (GetCampaign)
PUT    /:id                 (UpdateCampaign)
DELETE /:id                 (DeleteCampaign)
POST   /:id/activate        (ActivateCampaign)
POST   /:id/schedule        (ScheduleCampaign)
POST   /:id/pause           (PauseCampaign)
POST   /:id/resume          (ResumeCampaign)
POST   /:id/complete        (CompleteCampaign)
POST   /:id/archive         (ArchiveCampaign)
GET    /:id/stats           (GetCampaignStats)
POST   /:id/enroll          (EnrollContact)
GET    /:id/enrollments     (ListEnrollments)
```

#### 17. **WebSocket** (`/api/v1/ws`) - Auth + Rate Limited (5 connections/min)
```
GET    /messages            (HandleWebSocket - real-time messaging)
GET    /stats               (GetStats - WebSocket stats)
```

#### 18. **Test Endpoints** (`/api/v1/crm/test`) - Development Only
```
POST   /setup               (SetupTestEnvironment)
POST   /cleanup             (CleanupTestEnvironment)
PUT    /pipeline/:id/timeout (UpdatePipelineTimeout)
POST   /waha-message        (TestWAHAMessage)
POST   /send-waha-message   (SendWAHAMessage)
POST   /waha-connection     (TestWAHAConnection)
POST   /waha-qr             (TestWAHAQRCode)
```

#### 19. **Queue Management** (`/api/v1/queues`)
```
GET    /                    (ListQueues - RabbitMQ status)
```

---

### 3.2 Middlewares Aplicados

#### **Global Middlewares** (todas as rotas)
1. `middleware.GORMContextMiddleware(gormDB)` - Injeta GORM DB no contexto
2. `middleware.CorrelationIDMiddleware()` - Distributed tracing
3. `gin.Recovery()` - Panic recovery
4. `LoggerMiddleware(logger)` - Request logging
5. `CORSMiddleware()` - CORS headers

#### **Auth Middlewares**
1. `authMiddleware.Authenticate()` - JWT/API Key authentication
2. `rlsMiddleware.SetUserContext()` - Row-Level Security (tenant isolation)
3. `wsAuthMiddleware.Authenticate()` - WebSocket authentication

#### **Rate Limiting**
1. `middleware.AuthRateLimitMiddleware()` - 10 req/min (auth endpoints)
2. `middleware.UserBasedRateLimitMiddleware("1000-M")` - 1000 req/min per user (API endpoints)
3. `wsRateLimiter.RateLimit(5, 1*time.Minute)` - 5 WebSocket connections/min

---

### 3.3 Endpoints Totais

| Categoria | Total Endpoints |
|-----------|----------------|
| Health | 8 |
| Auth | 5 |
| Webhooks Inbound | 2 |
| Webhook Subscriptions | 6 |
| Contacts | 9 |
| Projects | 7 |
| Pipelines | 11 |
| Messages | 9 |
| Sessions | 6 |
| Channels | 12 |
| Agents | 9 |
| Chats | 9 |
| Notes | 2 |
| Trackings | 5 |
| Automation Discovery | 9 |
| Automations (Pipeline Rules) | 6 |
| Broadcasts | 8 |
| Sequences | 11 |
| Campaigns | 14 |
| WebSocket | 2 |
| Test | 7 |
| Queues | 1 |

**TOTAL ENDPOINTS: 158**

---

## PARTE 4: PROBLEMAS IDENTIFICADOS

### 4.1 Cache Layer - CRÍTICO ⚠️

**Status:** ❌ **NÃO IMPLEMENTADO**

**Impacto:**
- Todas as queries vão direto ao PostgreSQL sem cache intermediário
- Alto custo de I/O para dados frequentemente acessados
- Latência desnecessária em operações críticas (webhooks WAHA)

**Repositories Críticos:**
1. `GormContactRepository` - FindByPhone/FindByEmail (usado em todo webhook WAHA)
2. `GormAgentRepository` - FindByID (usado em sessões e mensagens)
3. `GormProjectRepository` - FindByID (usado em todo request)
4. `GormChannelRepository` - GetByWebhookID (usado em webhooks)
5. `GormSessionRepository` - FindByID (usado em mensagens)

**Recomendação:**
Implementar Redis cache layer com padrão **cache-aside**:
- TTL: 5 minutos para entidades mutáveis
- TTL: 1 hora para entidades estáticas (projects, agents)
- Cache invalidation: On update/delete

---

### 4.2 N+1 Queries - MÉDIO ⚠️

**Status:** 🔴 **DETECTADO EM 1 REPOSITORY**

**Arquivo:** `gorm_contact_list_repository.go`

**Problema:**
```go
// Linha ~45
err := r.db.Preload("Contacts").First(&entity, "id = ?", id).Error
```

**Impacto:**
- Carregar listas grandes (10k+ contatos) pode causar OOM (Out of Memory)
- Query única retorna TODOS os contatos da lista sem paginação
- Pode causar timeout em listas muito grandes

**Recomendação:**
1. **Remover Preload("Contacts")** do FindByID
2. Criar método separado `GetContactIDs(listID, limit, offset)` com paginação
3. Implementar lazy loading de contatos

**Fix Sugerido:**
```go
// ANTES (RUIM)
func (r *GormContactListRepository) FindByID(id uuid.UUID) (*contact_list.ContactList, error) {
    var entity entities.ContactListEntity
    err := r.db.Preload("Contacts").First(&entity, "id = ?", id).Error
    // ...
}

// DEPOIS (BOM)
func (r *GormContactListRepository) FindByID(id uuid.UUID) (*contact_list.ContactList, error) {
    var entity entities.ContactListEntity
    err := r.db.First(&entity, "id = ?", id).Error
    // Não carrega contatos automaticamente
    // ...
}

func (r *GormContactListRepository) GetContactIDs(listID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
    var contactIDs []uuid.UUID
    err := r.db.Model(&entities.ContactListMemberEntity{}).
        Where("list_id = ?", listID).
        Limit(limit).
        Offset(offset).
        Pluck("contact_id", &contactIDs).Error
    return contactIDs, err
}
```

---

### 4.3 Optimistic Locking Inconsistente - BAIXO ⚠️

**Status:** ⚠️ **IMPLEMENTAÇÃO PARCIAL**

**Repositories COM optimistic locking:**
- ✅ AgentRepository
- ✅ BillingRepository
- ✅ BroadcastRepository
- ✅ CampaignRepository (com transaction)
- ✅ ChatRepository
- ✅ ContactRepository
- ✅ ProjectRepository

**Repositories SEM optimistic locking:**
- ❌ ChannelRepository
- ❌ MessageRepository
- ❌ CredentialRepository
- ❌ TrackingRepository
- ❌ NoteRepository

**Impacto:**
- Risco de race conditions em updates concorrentes
- Possível perda de dados em ambientes com alta concorrência

**Recomendação:**
- Implementar optimistic locking em TODOS os repositories que possuem Update
- Prioridade: ChannelRepository, MessageRepository

---

### 4.4 RBAC (Role-Based Access Control) - MÉDIO ⚠️

**Status:** ❌ **NÃO IMPLEMENTADO**

**Impacto:**
- Não há controle granular de permissões por role (admin, agent, manager)
- Todos os usuários autenticados têm acesso aos mesmos endpoints
- Não há separação de permissões para operações sensíveis (delete project, delete agent)

**Recomendação:**
Implementar middleware RBAC:
```go
// infrastructure/http/middleware/rbac.go
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if !contains(allowedRoles, userRole) {
            c.JSON(403, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// Uso nas rotas:
projects.DELETE("/:id", rbacMiddleware.RequireRole("admin"), projectHandler.DeleteProject)
```

---

### 4.5 API Versioning - BAIXO ⚠️

**Status:** ⚠️ **ÚNICO VERSIONAMENTO (v1)**

**Impacto:**
- Não há estratégia para versionamento futuro (v2)
- Breaking changes forçarão todos os clientes a atualizar

**Recomendação:**
- Manter v1 estável
- Planejar v2 para mudanças breaking (se necessário)
- Considerar versioning por header (Accept: application/vnd.ventros.v2+json)

---

### 4.6 Swagger/OpenAPI Documentation - BAIXO ⚠️

**Status:** ⚠️ **IMPLEMENTADO MAS INCOMPLETO**

**Endpoint:** `GET /swagger/*any`

**Problema:**
- Documentação Swagger está configurada, mas pode estar desatualizada
- Não verificado se todos os 158 endpoints estão documentados

**Recomendação:**
- Auditar swagger.yaml/swagger.json
- Garantir que TODOS os endpoints estejam documentados
- Adicionar exemplos de request/response
- Documentar error codes

---

## PARTE 5: RECOMENDAÇÕES

### 5.1 Prioridade ALTA 🔴

1. **Implementar Redis Cache Layer**
   - Repositories críticos: Contact, Agent, Project, Channel
   - Cache-aside pattern com TTL configurável
   - Cache invalidation on update/delete

2. **Corrigir N+1 Query em ContactListRepository**
   - Remover Preload("Contacts")
   - Implementar paginação

3. **Implementar RBAC**
   - Proteger endpoints críticos (delete, admin operations)
   - Roles: admin, manager, agent, readonly

### 5.2 Prioridade MÉDIA 🟡

1. **Completar Optimistic Locking**
   - ChannelRepository
   - MessageRepository
   - CredentialRepository

2. **Implementar Circuit Breaker**
   - Proteger calls externos (WAHA, Stripe, LlamaParse)
   - Fallback strategies

3. **Melhorar Error Handling**
   - Padronizar error responses (RFC 7807 - Problem Details)
   - Adicionar error codes (VENTROS_ERR_001, etc)

### 5.3 Prioridade BAIXA 🟢

1. **API Versioning Strategy**
   - Planejar v2
   - Deprecation policy

2. **Swagger Documentation**
   - Auditar e completar documentação
   - Adicionar exemplos

3. **Monitoring & Observability**
   - Adicionar Prometheus metrics
   - Distributed tracing (OpenTelemetry)

---

## ANEXOS

### A1. Comandos para Verificar Repositories

```bash
# Contar repositories
find infrastructure/persistence -name "gorm_*_repository.go" | wc -l

# Verificar cache usage
grep -r "redis" infrastructure/persistence/gorm_*_repository.go

# Verificar N+1 queries
grep -r "Preload" infrastructure/persistence/gorm_*_repository.go

# Verificar optimistic locking
grep -r "version" infrastructure/persistence/gorm_*_repository.go | grep Updates
```

### A2. Comandos para Verificar Handlers

```bash
# Contar handlers
find infrastructure/http/handlers -name "*_handler.go" | wc -l

# Contar endpoints
grep -r "func (h \*.*Handler)" infrastructure/http/handlers/*.go | wc -l

# Verificar middlewares
grep -r "Use(" infrastructure/http/routes/routes.go
```

---

## CONCLUSÃO

O projeto Ventros CRM possui uma arquitetura hexagonal bem estruturada com:

✅ **Pontos Fortes:**
- 31 repositories bem organizados com adapters
- 27 handlers cobrindo todas as entidades de domínio
- 158 endpoints funcionais
- Optimistic locking implementado em 70% dos repositories críticos
- Segurança: Auth + RLS + Rate Limiting implementados
- Infraestrutura: Health checks, WebSockets, Webhooks

⚠️ **Pontos de Atenção:**
- Falta de cache layer (CRÍTICO para performance)
- 1 N+1 query detectado (ContactListRepository)
- RBAC não implementado (todos usuários têm mesmas permissões)
- Optimistic locking incompleto em alguns repositories

🎯 **Próximos Passos:**
1. Implementar Redis cache layer (prioridade ALTA)
2. Corrigir N+1 query
3. Implementar RBAC
4. Completar optimistic locking

---

**Gerado em:** 2025-10-13
**Arquiteto:** Claude (Anthropic)
**Versão do Relatório:** 1.0
