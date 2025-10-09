# Documentação da Camada de Aplicação

## Visão Geral

A camada de aplicação orquestra a lógica de negócio do domínio e coordena operações entre diferentes agregados. Esta camada implementa **Use Cases** e **Application Services** que encapsulam fluxos de trabalho complexos.

**Responsabilidades:**
- Orquestrar operações entre múltiplos agregados
- Implementar transações de aplicação
- Coordenar eventos de domínio
- Gerenciar conversões entre DTOs e entidades de domínio
- Implementar políticas de retry e idempotência

---

## 1. Agent Service

**Localização:** `internal/application/agent/`

### Use Cases

#### CreateAgentUseCase
**Arquivo:** `create_agent.go`

**Responsabilidade:** Criar novo agente no sistema

**Fluxo:**
1. Validar dados de entrada
2. Criar agregado Agent via domain factory
3. Persistir via repository
4. Publicar evento `agent.created`

**Regras:**
- Email deve ser único por tenant
- Nome não pode estar vazio
- Role deve ser válida (agent, supervisor, admin)

#### UpdateAgentUseCase
**Arquivo:** `update_agent.go`

**Responsabilidade:** Atualizar informações do agente

**Fluxo:**
1. Buscar agente existente
2. Aplicar mudanças via métodos do agregado
3. Persistir alterações
4. Publicar evento `agent.updated`

#### AssignAgentToSessionUseCase
**Arquivo:** `assign_agent_to_session.go`

**Responsabilidade:** Atribuir agente a uma sessão

**Fluxo:**
1. Verificar disponibilidade do agente
2. Validar carga de trabalho atual
3. Atribuir agente à sessão
4. Publicar evento `agent.assigned`

**Regras:**
- Agente deve estar ativo
- Respeitar limite de sessões simultâneas
- Considerar habilidades do agente

---

## 2. Channel Service

**Localização:** `internal/application/channel/channel_service.go`

### Responsabilidades

#### CreateChannel
**Responsabilidade:** Criar novo canal de comunicação

**Fluxo:**
1. Validar tipo de canal (WhatsApp, Facebook, etc)
2. Validar configurações específicas do provider
3. Criar agregado Channel
4. Inicializar integração com provider (WAHA, etc)
5. Publicar evento `channel.created`

**Regras:**
- Canal deve ter nome único por tenant
- Configurações devem corresponder ao tipo de canal
- API credentials devem ser válidas

#### ActivateChannel
**Responsabilidade:** Ativar canal para receber mensagens

**Fluxo:**
1. Validar estado do canal
2. Testar conectividade com provider
3. Ativar canal
4. Publicar evento `channel.activated`

**Regras:**
- Canal não pode estar já ativo
- Provider deve estar acessível
- Webhook deve estar configurado

#### ImportHistoryFromChannel
**Arquivo:** `waha_history_import.go`

**Responsabilidade:** Importar histórico de mensagens de canal WAHA

**Fluxo:**
1. Buscar mensagens do provider
2. Processar mensagens em lote
3. Criar contatos se necessário
4. Criar sessões se necessário
5. Salvar mensagens no banco

**Regras:**
- Não duplicar mensagens existentes
- Respeitar limite de taxa do provider
- Processar em ordem cronológica

---

## 3. Channel Type Service

**Localização:** `internal/application/channel_type/`

### Use Cases

#### ListChannelTypesUseCase
**Responsabilidade:** Listar tipos de canais disponíveis

**Retorna:**
- WhatsApp (WAHA)
- Facebook Messenger
- Instagram
- Telegram
- Email
- SMS

#### GetChannelTypeCapabilitiesUseCase
**Responsabilidade:** Retornar capacidades de um tipo de canal

**Exemplo (WhatsApp):**
- Suporta texto, imagem, áudio, vídeo, documento
- Suporta localização
- Suporta contatos
- Limite de 4096 caracteres por mensagem

---

## 4. Contact Service

**Localização:** `internal/application/contact/`

### Use Cases

#### CreateContactUseCase
**Arquivo:** `create_contact.go`

**Responsabilidade:** Criar novo contato

**Fluxo:**
1. Validar phone/email
2. Verificar se contato já existe (deduplicação)
3. Criar agregado Contact
4. Persistir
5. Publicar evento `contact.created`

**Regras de Deduplicação:**
- Phone number é chave única por tenant
- Email secundário para deduplicação
- Normalizar telefone antes de comparar (+55, 0, etc)

#### UpdateContactUseCase
**Arquivo:** `update_contact.go`

**Responsabilidade:** Atualizar informações do contato

**Fluxo:**
1. Buscar contato
2. Aplicar mudanças
3. Validar custom fields
4. Persistir
5. Publicar evento `contact.updated`

#### ChangePipelineStatusUseCase
**Arquivo:** `change_pipeline_status_usecase.go`

**Responsabilidade:** Mover contato no pipeline (funil)

**Fluxo:**
1. Buscar contato e pipeline
2. Validar se status de destino existe
3. Validar transição (regras de negócio do pipeline)
4. Atualizar status do contato
5. Publicar evento `contact.pipeline_status_changed`

**Regras:**
- Não pode pular estágios obrigatórios
- Registrar motivo da mudança
- Calcular tempo em cada estágio

#### FetchProfilePictureUseCase
**Arquivo:** `fetch_profile_picture_usecase.go`

**Responsabilidade:** Buscar foto de perfil do contato no WhatsApp

**Fluxo:**
1. Chamar API WAHA para buscar foto
2. Fazer download da imagem
3. Upload para storage (S3/MinIO)
4. Atualizar URL no contato
5. Publicar evento `contact.profile_picture_updated`

**Regras:**
- Executar de forma assíncrona
- Cache de 7 dias
- Fallback para avatar padrão

#### MergeContactsUseCase
**Arquivo:** `merge_contacts.go` (TODO)

**Responsabilidade:** Mesclar contatos duplicados

**Fluxo:**
1. Identificar contato principal e duplicado
2. Mesclar custom fields
3. Transferir sessões
4. Transferir notas
5. Transferir trackings
6. Marcar duplicado como merged
7. Publicar evento `contact.merged`

---

## 5. Contact Event Service

**Localização:** `internal/application/contact_event/`

### Use Cases

#### CreateContactEventUseCase
**Arquivo:** `create_contact_event.go`

**Responsabilidade:** Registrar evento de contato (conversão de ad, clique, etc)

**Fluxo:**
1. Validar tipo de evento
2. Criar agregado ContactEvent
3. Persistir
4. Publicar evento no outbox

**Tipos de Eventos:**
- `ad_conversion` - Conversão de anúncio
- `ad_click` - Clique em anúncio
- `form_submission` - Envio de formulário
- `page_view` - Visualização de página
- `custom` - Evento customizado

**Regras:**
- Eventos são imutáveis após criação
- Timestamp preciso (timezone UTC)
- Metadata em JSON flexível

#### StreamContactEventsUseCase
**Arquivo:** `stream_contact_events.go`

**Responsabilidade:** Stream de eventos do contato em tempo real (SSE)

**Fluxo:**
1. Estabelecer conexão SSE
2. Filtrar eventos por tenant e contato
3. Enviar eventos conforme chegam
4. Manter heartbeat

**Regras:**
- Timeout de 30 minutos
- Heartbeat a cada 15 segundos
- Suporta múltiplas conexões por contato

---

## 6. Contact List Service

**Localização:** `internal/application/contact_list/`

### Use Cases

#### CreateContactListUseCase
**Responsabilidade:** Criar lista de contatos para campanhas

**Fluxo:**
1. Criar lista com nome e descrição
2. Definir critérios de filtro (tags, pipeline status, custom fields)
3. Publicar evento `contact_list.created`

#### AddContactToListUseCase
**Responsabilidade:** Adicionar contato a lista

**Regras:**
- Contato não pode estar duplicado na lista
- Verificar se contato atende critérios

#### RemoveContactFromListUseCase
**Responsabilidade:** Remover contato da lista

#### GenerateDynamicListUseCase
**Responsabilidade:** Gerar lista dinâmica baseada em critérios

**Exemplo de Critérios:**
```json
{
  "tags": ["lead", "qualified"],
  "pipeline_status": "negociacao",
  "custom_fields": {
    "interesse": "premium"
  },
  "last_interaction_days": 7
}
```

---

## 7. Message Service

**Localização:** `internal/application/message/`

### WAHAMessageService
**Arquivo:** `waha_message_service.go`

#### ProcessWAHAMessage
**Responsabilidade:** Processar mensagem recebida do webhook WAHA

**Fluxo:**
1. Validar payload do webhook
2. Converter evento WAHA para evento de domínio
3. Identificar/criar contato
4. Identificar/criar sessão
5. Criar mensagem inbound
6. Persistir tudo em transação
7. Publicar evento `message.received`
8. Disparar processamento de IA (se configurado)

**Regras:**
- Idempotência via message ID do WhatsApp
- Criar contato automaticamente se não existir
- Criar nova sessão se última sessão expirou (>30min)
- Extrair mídia e fazer upload assíncrono

**Tratamento de Tipos de Mensagem:**
- **Texto:** Extrair conteúdo direto
- **Imagem:** Download + OCR + descrição via IA
- **Áudio:** Download + transcrição via Whisper
- **Documento:** Download + extração de texto
- **Localização:** Extrair lat/long e endereço
- **Contato:** Extrair VCard
- **Sticker:** Tratar como imagem

#### SendMessage
**Responsabilidade:** Enviar mensagem para contato

**Fluxo:**
1. Validar permissões (RLS)
2. Validar janela de 24h do WhatsApp
3. Criar mensagem outbound
4. Persistir no banco
5. Enviar para fila RabbitMQ (async)
6. Publicar evento `message.sent`

**Regras:**
- Validar janela de 24h (WhatsApp Business Policy)
- Se fora da janela, exigir template aprovado
- Rate limiting por canal
- Retry com backoff exponencial

### ProcessInboundMessageUseCase
**Arquivo:** `process_inbound_message.go`

**Responsabilidade:** Processar mensagem recebida e executar automações

**Fluxo:**
1. Identificar intenção (NLP)
2. Executar regras de automação
3. Atribuir a agente se necessário
4. Atualizar pipeline
5. Disparar workflows Temporal

**Automações:**
- Auto-resposta baseada em palavra-chave
- Distribuição automática de leads
- Qualificação automática via IA
- Extração de entidades (nome, email, etc)

---

## 8. Messaging Service

**Localização:** `internal/application/messaging/`

### SendMessageUseCase
**Arquivo:** `send_message.go`

**Responsabilidade:** Enviar mensagem via canal específico

**Fluxo:**
1. Selecionar sender factory (WAHA, etc)
2. Validar mensagem
3. Aplicar rate limiting
4. Enviar para fila
5. Registrar na tabela outbound_messages

**Factories:**
- `WAHAMessageSender` - WhatsApp via WAHA
- `EmailMessageSender` - Email (TODO)
- `SMSMessageSender` - SMS (TODO)

### ScheduleMessageUseCase
**Arquivo:** `schedule_message.go`

**Responsabilidade:** Agendar mensagem para envio futuro

**Fluxo:**
1. Criar mensagem com status `scheduled`
2. Persistir com timestamp de envio
3. Scheduler do Temporal processa no horário

**Regras:**
- Respeitar timezone do contato
- Respeitar horário comercial
- Cancelável até 1 minuto antes

---

## 9. Note Service

**Localização:** `internal/application/note/`

### Use Cases

#### CreateNoteUseCase
**Responsabilidade:** Criar nota sobre contato

**Fluxo:**
1. Validar conteúdo
2. Criar agregado Note
3. Associar a sessão (opcional)
4. Persistir
5. Publicar evento `note.added`

**Tipos de Nota:**
- `general` - Nota geral
- `automation` - Follow-up necessário
- `complaint` - Reclamação
- `resolution` - Resolução
- `escalation` - Escalação
- `internal` - Interna (não visível para cliente)
- `session_summary` - Resumo de sessão (gerado por IA)
- `session_handoff` - Handoff entre agentes
- `ad_conversion` - Conversão de anúncio

#### UpdateNoteUseCase
**Responsabilidade:** Atualizar nota existente

#### PinNoteUseCase
**Responsabilidade:** Fixar nota importante

**Regras:**
- Máximo 5 notas fixadas por contato
- Notas fixadas aparecem no topo

#### DeleteNoteUseCase
**Responsabilidade:** Soft delete de nota

---

## 10. Pipeline Service

**Localização:** `internal/application/pipeline/`

### Use Cases

#### CreatePipelineUseCase
**Responsabilidade:** Criar funil de vendas/atendimento

**Fluxo:**
1. Criar pipeline com nome
2. Definir estágios (statuses)
3. Definir regras de transição
4. Publicar evento `pipeline.created`

**Exemplo de Pipeline:**
```
Lead → Qualificado → Negociação → Proposta → Fechado/Perdido
```

#### UpdatePipelineUseCase
**Responsabilidade:** Atualizar configuração do pipeline

#### AddStageUseCase
**Responsabilidade:** Adicionar novo estágio ao pipeline

**Regras:**
- Estágio deve ter nome único
- Definir ordem (position)
- Definir se é estágio final (won/lost)

#### MoveContactInPipelineUseCase
**Responsabilidade:** Mover contato entre estágios

**Fluxo:**
1. Validar transição permitida
2. Registrar motivo da mudança
3. Atualizar contato
4. Calcular métricas (tempo no estágio)
5. Publicar evento `pipeline.contact_moved`

---

## 11. Project Service

**Localização:** `internal/application/project/`

### Use Cases

#### CreateProjectUseCase
**Responsabilidade:** Criar novo projeto/workspace (tenant)

**Fluxo:**
1. Validar customer e billing account
2. Gerar tenant_id único
3. Criar agregado Project
4. Inicializar configurações padrão
5. Publicar evento `project.created`
6. Criar estruturas iniciais (pipelines, agentes, etc)

**Configurações Padrão:**
- `session_timeout_minutes`: 30
- `ai_enabled`: true
- `auto_assignment`: true
- `business_hours`: 09:00-18:00

#### UpdateProjectConfigUseCase
**Responsabilidade:** Atualizar configurações do projeto

#### DeactivateProjectUseCase
**Responsabilidade:** Desativar projeto (suspensão)

**Fluxo:**
1. Validar billing status
2. Desativar todos os canais
3. Pausar workflows Temporal
4. Marcar projeto como inativo
5. Publicar evento `project.deactivated`

---

## 12. Session Service

**Localização:** `internal/application/session/`

### RecordMessageUseCase
**Arquivo:** `record_message.go`

**Responsabilidade:** Registrar mensagem em sessão ativa

**Fluxo:**
1. Buscar ou criar sessão ativa
2. Adicionar mensagem à sessão
3. Atualizar métricas (total_messages, etc)
4. Resetar timeout timer
5. Publicar evento `session.message_added`

**Regras:**
- Criar nova sessão se não houver ativa
- Sessão expira após 30 minutos de inatividade
- Atualizar last_interaction_at

### CloseSessionUseCase
**Responsabilidade:** Encerrar sessão

**Fluxo:**
1. Buscar sessão
2. Calcular métricas finais
3. Gerar resumo via IA (opcional)
4. Marcar como closed
5. Publicar evento `session.closed`

**Métricas Calculadas:**
- Duração total
- Tempo de primeira resposta
- Tempo médio de resposta
- Total de mensagens
- Satisfação (se CSAT enviado)

### AssignAgentToSessionUseCase
**Responsabilidade:** Atribuir agente humano à sessão

**Fluxo:**
1. Validar disponibilidade do agente
2. Transferir de bot para humano
3. Notificar agente
4. Publicar evento `session.agent_assigned`

### TransferSessionUseCase
**Responsabilidade:** Transferir sessão entre agentes

**Fluxo:**
1. Validar agente destino
2. Criar nota de handoff
3. Atualizar agente na sessão
4. Notificar ambos os agentes
5. Publicar evento `session.transferred`

---

## 13. Tracking Service

**Localização:** `internal/application/tracking/`

### Use Cases

#### CreateTrackingUseCase
**Responsabilidade:** Criar tracking de origem de contato

**Fluxo:**
1. Extrair UTM parameters
2. Identificar source, medium, campaign
3. Criar agregado Tracking
4. Associar a contato
5. Persistir
6. Publicar evento `tracking.created`

**Fontes Suportadas:**
- Facebook Ads
- Google Ads
- Instagram Ads
- TikTok Ads
- Orgânico
- Direto
- Referral

#### EnrichTrackingUseCase
**Responsabilidade:** Enriquecer tracking com dados de campanha

**Fluxo:**
1. Chamar API do Facebook/Google Ads
2. Buscar dados da campanha
3. Calcular custo por lead
4. Criar TrackingEnrichment
5. Publicar evento `tracking.enriched`

**Dados Enriquecidos:**
- Nome da campanha
- Ad set
- Ad criativo
- Custo
- Impressões
- Cliques
- CPL (custo por lead)

#### AttributeConversionUseCase
**Responsabilidade:** Atribuir conversão a origem

**Fluxo:**
1. Identificar tracking original
2. Calcular ROI
3. Enviar evento de conversão para ad platform
4. Publicar evento `tracking.conversion_attributed`

**Modelos de Atribuição:**
- First-touch (primeira interação)
- Last-touch (última interação)
- Linear (distribuído igualmente)

---

## 14. User Service

**Localização:** `internal/application/user/user_service.go`

### Use Cases

#### AuthenticateUser
**Responsabilidade:** Autenticar usuário

**Fluxo:**
1. Buscar usuário por email
2. Verificar password hash (bcrypt)
3. Gerar JWT token
4. Registrar login
5. Retornar token + user info

**JWT Claims:**
```json
{
  "user_id": "uuid",
  "tenant_id": "string",
  "email": "string",
  "roles": ["agent", "admin"],
  "permissions": ["contacts.read", "messages.send"],
  "exp": 1234567890
}
```

#### CreateUser
**Responsabilidade:** Criar novo usuário

**Fluxo:**
1. Validar email único
2. Hash da senha (bcrypt, custo 12)
3. Criar usuário
4. Atribuir roles
5. Publicar evento `user.created`

#### AssignRoleToUser
**Responsabilidade:** Atribuir role ao usuário

**Roles Disponíveis:**
- `admin` - Acesso total
- `supervisor` - Gerencia agentes e relatórios
- `agent` - Atendimento
- `viewer` - Somente leitura

---

## 15. Webhook Service

**Localização:** `internal/application/webhook/`

### ManageSubscriptionUseCase
**Arquivo:** `manage_subscription.go`

**Responsabilidade:** Gerenciar subscrições de webhooks

**Fluxo:**
1. Criar subscrição com URL e eventos
2. Validar URL (fazer ping)
3. Gerar secret para assinatura
4. Persistir
5. Publicar evento `webhook.subscribed`

**Tipos de Evento:**
- `message.received`
- `message.sent`
- `session.closed`
- `contact.created`
- `contact.updated`
- `agent.assigned`

### DeliverWebhookUseCase
**Responsabilidade:** Entregar webhook para subscritor

**Fluxo:**
1. Buscar subscrições ativas para evento
2. Construir payload
3. Calcular HMAC signature
4. Enviar para fila RabbitMQ (async)
5. Registrar tentativa

**Retry Policy:**
- 3 tentativas
- Backoff: 1min, 5min, 15min
- Desativar subscrição após 10 falhas consecutivas

---

## Padrões de Implementação

### 1. Transactional Outbox

Todos os use cases que publicam eventos de domínio seguem o padrão:

```go
func (uc *SomeUseCase) Execute(ctx context.Context, input Input) error {
    // 1. Iniciar transação
    tx := uc.db.Begin()
    defer tx.Rollback()

    // 2. Executar operações de domínio
    aggregate := domain.NewAggregate(...)
    if err := uc.repo.Save(ctx, aggregate); err != nil {
        return err
    }

    // 3. Persistir eventos no outbox (mesma transação)
    events := aggregate.DomainEvents()
    for _, event := range events {
        if err := uc.outboxRepo.Save(ctx, event); err != nil {
            return err
        }
    }

    // 4. Commit
    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}
```

### 2. Idempotência

Todos os consumers implementam idempotência:

```go
func (c *Consumer) ProcessMessage(ctx context.Context, msg Message) error {
    eventID := msg.EventID

    // 1. Verificar se já processado
    processed, err := c.idempotencyChecker.IsProcessed(ctx, eventID, c.name)
    if err != nil {
        return err
    }
    if processed {
        return nil // Skip
    }

    // 2. Processar
    if err := c.useCase.Execute(ctx, msg.Data); err != nil {
        return err
    }

    // 3. Marcar como processado
    return c.idempotencyChecker.MarkAsProcessed(ctx, eventID, c.name, nil)
}
```

### 3. CQRS Leve

Separação entre comandos (write) e queries (read):

**Command:**
```go
type CreateContactCommand struct {
    Name  string
    Phone string
}

func (uc *CreateContactUseCase) Execute(ctx context.Context, cmd CreateContactCommand) error {
    contact := domain.NewContact(cmd.Name, cmd.Phone)
    return uc.repo.Save(ctx, contact)
}
```

**Query:**
```go
type FindContactQuery struct {
    TenantID string
    Phone    string
}

func (qs *ContactQueryService) FindByPhone(ctx context.Context, q FindContactQuery) (*ContactDTO, error) {
    return qs.repo.FindByPhone(ctx, q.TenantID, q.Phone)
}
```

### 4. DTOs (Data Transfer Objects)

**Localização:** `internal/application/dtos/`

Todas as operações usam DTOs para entrada/saída:

```go
// Contact DTOs
type CreateContactRequest struct {
    Name         string                 `json:"name"`
    Phone        string                 `json:"phone"`
    Email        string                 `json:"email"`
    CustomFields map[string]interface{} `json:"custom_fields"`
}

type ContactResponse struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Phone        string                 `json:"phone"`
    Email        string                 `json:"email"`
    CustomFields map[string]interface{} `json:"custom_fields"`
    CreatedAt    time.Time              `json:"created_at"`
}
```

---

## Fluxos de Negócio Importantes

### Fluxo de Mensagem Inbound

```
1. WhatsApp → WAHA Webhook → RabbitMQ (waha.messages)
2. WAHAMessageConsumer processa evento
3. WAHAMessageService.ProcessWAHAMessage:
   - Identifica/cria contato
   - Identifica/cria sessão
   - Cria mensagem inbound
   - Salva tudo + outbox events (transação atômica)
4. OutboxProcessor publica eventos:
   - message.received
   - contact.created (se novo)
   - session.started (se nova)
5. Consumers reagem:
   - AI Processor: analisa sentimento, extrai entidades
   - Auto-assign: atribui agente se regra matchear
   - Webhook Delivery: envia para webhooks subscritos
```

### Fluxo de Mensagem Outbound

```
1. API Handler recebe POST /messages/send
2. SendMessageUseCase:
   - Valida permissões (RLS)
   - Valida janela 24h WhatsApp
   - Cria mensagem outbound (status: pending)
   - Persiste + outbox event
3. OutboxProcessor publica message.sending
4. WAHAMessageSender processa:
   - Aplica rate limiting
   - Envia para WAHA API
   - Atualiza status (sent/failed)
   - Persiste + outbox event
5. OutboxProcessor publica message.sent
6. Webhooks são disparados
```

### Fluxo de Conversão de Ad

```
1. Facebook/Instagram: usuário clica em ad e envia mensagem
2. WAHA recebe mensagem com metadata do ad:
   - fb_ad_id
   - fb_ad_name
   - fb_campaign_id
   - source: "facebook_ads"
3. ProcessWAHAMessage detecta origem de ad
4. Cria ContactEvent (type: ad_conversion)
5. Cria Tracking com UTM parameters
6. EnrichTrackingUseCase:
   - Busca dados da campanha (FB Graph API)
   - Calcula CPL
   - Salva TrackingEnrichment
7. Publicar evento tracking.conversion
8. Enviar conversion event de volta para Facebook
```

### Fluxo de Sessão com IA + Handoff

```
1. Contato envia mensagem → Sessão criada (bot)
2. AI Processor analisa:
   - Sentimento: neutro
   - Intenção: dúvida_produto
   - Complexidade: baixa
3. Bot responde automaticamente
4. Contato: "quero falar com humano"
5. AI detecta intent: solicitar_atendente
6. Trigger: AssignAgentToSessionUseCase
7. Algoritmo de distribuição:
   - Busca agentes disponíveis
   - Considera habilidades
   - Considera carga atual
   - Seleciona melhor agente
8. Agente é notificado (WebSocket)
9. Sessão transferida (bot → humano)
10. Nota automática criada com contexto da conversa
```

---

## Métricas e Observabilidade

Cada use case registra métricas:

```go
// Duração de execução
startTime := time.Now()
defer func() {
    duration := time.Since(startTime)
    metrics.RecordUseCaseDuration("create_contact", duration)
}()

// Contadores
metrics.IncrementUseCaseExecutions("create_contact")
if err != nil {
    metrics.IncrementUseCaseErrors("create_contact")
}
```

**Métricas Importantes:**
- `usecase_duration_seconds` - Duração de cada use case
- `usecase_executions_total` - Total de execuções
- `usecase_errors_total` - Total de erros
- `message_processing_duration_seconds` - Tempo de processamento de mensagens
- `event_processing_lag_seconds` - Latência entre evento criado e processado

---

## Resumo

A camada de aplicação do Ventros CRM implementa **50+ use cases** organizados em 14 serviços principais. Cada use case segue padrões consistentes de:

- ✅ Transactional Outbox para consistência eventual
- ✅ Idempotência para processamento seguro
- ✅ DTOs para desacoplamento
- ✅ Event-driven para extensibilidade
- ✅ Observabilidade com métricas

A arquitetura permite:
- **Escalabilidade:** Processamento assíncrono via RabbitMQ
- **Confiabilidade:** Retry automático e DLQ
- **Auditoria:** Todos eventos registrados
- **Extensibilidade:** Novos consumers sem alterar código existente
