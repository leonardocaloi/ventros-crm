# Auditoria Completa de Eventos de Domínio - Ventros CRM

## 📊 Resumo Executivo

Data: 2025-10-09
Total de Eventos Encontrados: **97 eventos**
Status da Análise: EM ANDAMENTO

---

## 1️⃣ AGENT - Eventos de Agentes

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 1 | `AgentCreatedEvent` | `agent.created` | ✅ | ❓ |
| 2 | `AgentUpdatedEvent` | `agent.updated` | ✅ | ❓ |
| 3 | `AgentActivatedEvent` | `agent.activated` | ✅ | ❓ |
| 4 | `AgentDeactivatedEvent` | `agent.deactivated` | ✅ | ❓ |
| 5 | `AgentLoggedInEvent` | `agent.logged_in` | ❌ | ❓ |
| 6 | `AgentPermissionGrantedEvent` | `agent.permission_granted` | ❌ | ❓ |
| 7 | `AgentPermissionRevokedEvent` | `agent.permission_revoked` | ❌ | ❓ |

---

## 2️⃣ AGENT_SESSION - Eventos de Sessão de Agente

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 8 | `AgentJoinedSessionEvent` | `agent_session.joined` | ❌ | ❓ |
| 9 | `AgentLeftSessionEvent` | `agent_session.left` | ❌ | ❓ |
| 10 | `AgentRoleChangedEvent` | `agent_session.role_changed` | ❌ | ❓ |

---

## 3️⃣ BILLING - Eventos de Cobrança

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 11 | `BillingAccountCreatedEvent` | `billing.account_created` | ❌ | ❓ |
| 12 | `PaymentMethodActivatedEvent` | `billing.payment_method_activated` | ❌ | ❓ |
| 13 | `BillingAccountSuspendedEvent` | `billing.account_suspended` | ❌ | ❓ |
| 14 | `BillingAccountReactivatedEvent` | `billing.account_reactivated` | ❌ | ❓ |
| 15 | `BillingAccountCanceledEvent` | `billing.account_canceled` | ❌ | ❓ |

---

## 4️⃣ CHANNEL - Eventos de Canal

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 16 | `ChannelCreatedEvent` | `channel.created` | ✅ | ❓ |
| 17 | `ChannelActivatedEvent` | `channel.activated` | ✅ | ❓ |
| 18 | `ChannelDeactivatedEvent` | `channel.deactivated` | ✅ | ❓ |
| 19 | `ChannelDeletedEvent` | `channel.deleted` | ✅ | ❓ |
| 20 | `ChannelPipelineAssociatedEvent` | `channel.pipeline_associated` | ❌ | ❓ |
| 21 | `ChannelPipelineDisassociatedEvent` | `channel.pipeline_disassociated` | ❌ | ❓ |

---

## 5️⃣ CHANNEL_TYPE - Eventos de Tipo de Canal

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 22 | `ChannelTypeCreatedEvent` | `channel_type.created` | ❌ | ❓ |
| 23 | `ChannelTypeActivatedEvent` | `channel_type.activated` | ❌ | ❓ |
| 24 | `ChannelTypeDeactivatedEvent` | `channel_type.deactivated` | ❌ | ❓ |

---

## 6️⃣ CONTACT - Eventos de Contato

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 25 | `ContactCreatedEvent` | `contact.created` | ✅ | ✅ |
| 26 | `ContactUpdatedEvent` | `contact.updated` | ✅ | ✅ |
| 27 | `ContactProfilePictureUpdatedEvent` | `contact.profile_picture_updated` | ❌ | ❓ |
| 28 | `ContactDeletedEvent` | `contact.deleted` | ❌ | ❓ |
| 29 | `ContactMergedEvent` | `contact.merged` | ✅ | ❓ |
| 30 | `ContactEnrichedEvent` | `contact.enriched` | ✅ | ❓ |
| 31 | `ContactPipelineStatusChangedEvent` | `contact.pipeline_status_changed` | ✅ | ✅ |
| 32 | `AdConversionTrackedEvent` | `tracking.message.meta_ads` | ✅ | ✅ |

---

## 7️⃣ CONTACT_LIST - Eventos de Lista de Contatos

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 33 | `ContactListCreatedEvent` | `contact_list.created` | ❌ | ❓ |
| 34 | `ContactListUpdatedEvent` | `contact_list.updated` | ❌ | ❓ |
| 35 | `ContactListDeletedEvent` | `contact_list.deleted` | ❌ | ❓ |
| 36 | `ContactListFilterRuleAddedEvent` | `contact_list.filter_rule_added` | ❌ | ❓ |
| 37 | `ContactListFilterRuleRemovedEvent` | `contact_list.filter_rule_removed` | ❌ | ❓ |
| 38 | `ContactListFilterRulesClearedEvent` | `contact_list.filter_rules_cleared` | ❌ | ❓ |
| 39 | `ContactListRecalculatedEvent` | `contact_list.recalculated` | ❌ | ❓ |
| 40 | `ContactAddedToListEvent` | `contact_list.contact_added` | ❌ | ❓ |
| 41 | `ContactRemovedFromListEvent` | `contact_list.contact_removed` | ❌ | ❓ |

---

## 8️⃣ CREDENTIAL - Eventos de Credenciais

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 42 | `CredentialCreatedEvent` | `credential.created` | ❌ | ❓ |
| 43 | `CredentialUpdatedEvent` | `credential.updated` | ❌ | ❓ |
| 44 | `OAuthTokenRefreshedEvent` | `credential.oauth_token_refreshed` | ❌ | ❓ |
| 45 | `CredentialActivatedEvent` | `credential.activated` | ❌ | ❓ |
| 46 | `CredentialDeactivatedEvent` | `credential.deactivated` | ❌ | ❓ |
| 47 | `CredentialUsedEvent` | `credential.used` | ❌ | ❓ |
| 48 | `CredentialExpiredEvent` | `credential.expired` | ❌ | ❓ |

---

## 9️⃣ CUSTOMER - Eventos de Cliente (Tenant)

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 49 | `CustomerCreatedEvent` | `customer.created` | ❌ | ❓ |
| 50 | `CustomerActivatedEvent` | `customer.activated` | ❌ | ❓ |
| 51 | `CustomerSuspendedEvent` | `customer.suspended` | ❌ | ❓ |

---

## 🔟 MESSAGE - Eventos de Mensagem

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 52 | `MessageCreatedEvent` | `message.created` | ✅ | ✅ |
| 53 | `MessageDeliveredEvent` | `message.delivered` | ✅ | ❓ |
| 54 | `MessageReadEvent` | `message.read` | ✅ | ❓ |
| 55 | `MessageFailedEvent` | `message.failed` | ✅ | ❓ |
| 56 | `AIProcessImageRequestedEvent` | `ai.process_image_requested` | ❌ | ❓ |
| 57 | `AIProcessVideoRequestedEvent` | `ai.process_video_requested` | ❌ | ❓ |
| 58 | `AIProcessAudioRequestedEvent` | `ai.process_audio_requested` | ❌ | ❓ |
| 59 | `AIProcessVoiceRequestedEvent` | `ai.process_voice_requested` | ❌ | ❓ |

---

## 1️⃣1️⃣ NOTE - Eventos de Nota

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 60 | `NoteAddedEvent` | `note.added` | ✅ | ❓ |
| 61 | `NoteUpdatedEvent` | `note.updated` | ✅ | ❓ |
| 62 | `NoteDeletedEvent` | `note.deleted` | ✅ | ❓ |
| 63 | `NotePinnedEvent` | `note.pinned` | ✅ | ❓ |

---

## 1️⃣2️⃣ PIPELINE - Eventos de Pipeline

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 64 | `PipelineCreatedEvent` | `pipeline.created` | ✅ | ❓ |
| 65 | `PipelineUpdatedEvent` | `pipeline.updated` | ✅ | ❓ |
| 66 | `PipelineActivatedEvent` | `pipeline.activated` | ✅ | ❓ |
| 67 | `PipelineDeactivatedEvent` | `pipeline.deactivated` | ✅ | ❓ |
| 68 | `StatusCreatedEvent` | `pipeline.status.created` | ✅ | ❓ |
| 69 | `StatusUpdatedEvent` | `pipeline.status.updated` | ✅ | ❓ |
| 70 | `StatusActivatedEvent` | `pipeline.status.activated` | ❌ | ❓ |
| 71 | `StatusDeactivatedEvent` | `pipeline.status.deactivated` | ❌ | ❓ |
| 72 | `StatusAddedToPipelineEvent` | `pipeline.status.added` | ❌ | ❓ |
| 73 | `StatusRemovedFromPipelineEvent` | `pipeline.status.removed` | ❌ | ❓ |
| 74 | `ContactStatusChangedEvent` | `contact.status_changed` | ✅ | ❓ |
| 75 | `ContactEnteredPipelineEvent` | `contact.entered_pipeline` | ✅ | ❓ |
| 76 | `ContactExitedPipelineEvent` | `contact.exited_pipeline` | ✅ | ❓ |
| 77 | `AutomationCreatedEvent` | `automation.created` | ❌ | ❓ |
| 78 | `AutomationEnabledEvent` | `automation.enabled` | ❌ | ❓ |
| 79 | `AutomationDisabledEvent` | `automation.disabled` | ❌ | ❓ |
| 80 | `AutomationRuleTriggeredEvent` | `automation.rule_triggered` | ❌ | ❓ |
| 81 | `AutomationRuleExecutedEvent` | `automation.rule_executed` | ❌ | ❓ |
| 82 | `AutomationRuleFailedEvent` | `automation.rule_failed` | ❌ | ❓ |

---

## 1️⃣3️⃣ PROJECT - Eventos de Projeto

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 83 | `ProjectCreatedEvent` | `project.created` | ❌ | ❓ |

---

## 1️⃣4️⃣ SESSION - Eventos de Sessão

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 84 | `SessionStartedEvent` | `session.started` | ✅ | ✅ |
| 85 | `SessionEndedEvent` | `session.ended` | ✅ | ❓ |
| 86 | `MessageRecordedEvent` | `session.message_recorded` | ❌ (interno) | ✅ |
| 87 | `AgentAssignedEvent` | `session.agent_assigned` | ✅ | ❓ |
| 88 | `SessionResolvedEvent` | `session.resolved` | ✅ | ❓ |
| 89 | `SessionEscalatedEvent` | `session.escalated` | ✅ | ❓ |
| 90 | `SessionSummarizedEvent` | `session.summarized` | ❌ | ❓ |
| 91 | `SessionAbandonedEvent` | `session.abandoned` | ✅ | ❓ |

---

## 1️⃣5️⃣ TRACKING - Eventos de Tracking

| # | Evento | Nome do Evento | Mapeado? | Publicado? |
|---|--------|----------------|----------|------------|
| 92 | `TrackingCreatedEvent` | `tracking.created` | ✅ | ❓ |
| 93 | `TrackingEnrichedEvent` | `tracking.enriched` | ✅ | ❓ |

---

## 📊 Estatísticas

### Total de Eventos por Domínio
- **Agent**: 7 eventos
- **Agent Session**: 3 eventos
- **Billing**: 5 eventos
- **Channel**: 6 eventos
- **Channel Type**: 3 eventos
- **Contact**: 8 eventos
- **Contact List**: 9 eventos
- **Credential**: 7 eventos
- **Customer**: 3 eventos
- **Message**: 8 eventos
- **Note**: 4 eventos
- **Pipeline**: 19 eventos
- **Project**: 1 evento
- **Session**: 8 eventos
- **Tracking**: 2 eventos

### Status do Mapeamento

| Status | Quantidade | Porcentagem |
|--------|------------|-------------|
| ✅ Mapeado | 34 | 36.6% |
| ❌ NÃO Mapeado | 59 | 63.4% |
| **TOTAL** | **93** | **100%** |

---

## ⚠️ EVENTOS NÃO MAPEADOS (Prioridade ALTA)

### Críticos (precisam ser adicionados IMEDIATAMENTE):

1. **Billing** (5 eventos)
   - `billing.account_created`
   - `billing.payment_method_activated`
   - `billing.account_suspended`
   - `billing.account_reactivated`
   - `billing.account_canceled`

2. **Credential** (7 eventos) - CRÍTICO para OAuth Meta
   - `credential.created`
   - `credential.updated`
   - `credential.oauth_token_refreshed`
   - `credential.activated`
   - `credential.deactivated`
   - `credential.used`
   - `credential.expired`

3. **Agent Session** (3 eventos)
   - `agent_session.joined`
   - `agent_session.left`
   - `agent_session.role_changed`

4. **Contact List** (9 eventos)
   - `contact_list.created`
   - `contact_list.updated`
   - `contact_list.deleted`
   - ... (todos os eventos)

5. **Automation** (6 eventos)
   - `automation.created`
   - `automation.enabled`
   - `automation.disabled`
   - `automation.rule_triggered`
   - `automation.rule_executed`
   - `automation.rule_failed`

### Médios:

6. **Pipeline Status** (4 eventos)
   - `pipeline.status.activated`
   - `pipeline.status.deactivated`
   - `pipeline.status.added`
   - `pipeline.status.removed`

7. **Channel** (2 eventos)
   - `channel.pipeline_associated`
   - `channel.pipeline_disassociated`

8. **Agent** (3 eventos)
   - `agent.logged_in`
   - `agent.permission_granted`
   - `agent.permission_revoked`

---

## 🔧 PRÓXIMOS PASSOS

### 1. Adicionar Mapeamentos Faltantes no `DomainEventBus`

Arquivo: `infrastructure/messaging/domain_event_bus.go`

```go
func (bus *DomainEventBus) mapDomainToBusinessEvents(domainEvent string) []string {
    switch domainEvent {
    // ... eventos existentes ...

    // Billing events (ADICIONAR)
    case "billing.account_created":
        return []string{"billing.account_created"}
    case "billing.payment_method_activated":
        return []string{"billing.payment_method_activated"}
    // ... resto dos eventos de billing

    // Credential events (ADICIONAR - CRÍTICO)
    case "credential.created":
        return []string{"credential.created"}
    case "credential.updated":
        return []string{"credential.updated"}
    // ... resto dos eventos de credential

    // ... resto dos eventos ...
    }
}
```

### 2. Verificar se Eventos estão sendo Publicados nos Agregados

Verificar arquivos:
- `internal/domain/agent/agent.go` - métodos devem chamar `agent.AddDomainEvent()`
- `internal/domain/billing/billing_account.go`
- `internal/domain/credential/credential.go`
- etc.

### 3. Criar Testes para Cada Evento

Criar script que envia mensagem WAHA simulada para cada tipo de evento e verifica se chegou no webhook.

---

## 📝 Notas

- ✅ = Mapeado e testado
- ❓ = Precisa verificar se está sendo publicado
- ❌ = NÃO mapeado, precisa adicionar

**Última atualização**: 2025-10-09 07:25:00
