package message

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/saga"
	"github.com/ventros/crm/internal/domain/core/shared"
	domainagent "github.com/ventros/crm/internal/domain/crm/agent"
	"github.com/ventros/crm/internal/domain/crm/contact"
	domaincontact "github.com/ventros/crm/internal/domain/crm/contact"
	contact_event "github.com/ventros/crm/internal/domain/crm/contact_event"
	domainmessage "github.com/ventros/crm/internal/domain/crm/message"
	domainsession "github.com/ventros/crm/internal/domain/crm/session"
	"github.com/ventros/crm/internal/domain/crm/tracking"
	sagaworkflow "github.com/ventros/crm/internal/workflows/saga"
	sessionworkflow "github.com/ventros/crm/internal/workflows/session"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

// ProcessInboundMessageCommand represents the command to process an inbound message
// IMPORTANTE: SessionID aqui refere-se à SESSÃO DO CRM (conversa com timeout gerenciado pelo Temporal),
// NÃO ao session_id do WAHA (que é o ExternalID do canal). Não confundir os dois conceitos!
type ProcessInboundMessageCommand struct {
	MessageID        string // ID interno (legado, pode ser removido)
	ChannelMessageID string // ID externo do WhatsApp (usado para deduplicação)
	FromPhone        string
	MessageText      string
	Timestamp        time.Time
	MessageType      string
	MediaURL         string
	MediaType        string
	// Channel context (OBRIGATÓRIO - toda mensagem vem de um canal)
	ChannelID uuid.UUID // UUID do canal de onde a mensagem veio
	// User context (obtido do canal)
	ProjectID  uuid.UUID
	CustomerID uuid.UUID
	TenantID   string
	// Additional fields
	ContactPhone  string
	ContactName   string
	ChannelTypeID int
	ContentType   string
	Text          string
	MediaMimetype string
	TrackingData  map[string]interface{}
	ReceivedAt    time.Time
	Metadata      map[string]interface{} // Pode conter "waha_session" (ExternalID do canal), "channel_name", etc.
	FromMe        bool                   // Se a mensagem foi enviada pelo sistema (fromMe: true)
	// Group and mentions support
	IsGroupMessage  bool       // Se a mensagem é de um grupo do WhatsApp
	GroupExternalID string     // ID externo do grupo (ex: "123456789@g.us")
	Participant     string     // Em grupos: quem ENVIOU a mensagem (participant)
	Mentions        []string   // IDs dos usuários mencionados (@marcados) no formato WAHA
	ChatID          *uuid.UUID // ID do Chat (grupo ou individual) - será preenchido durante processamento
}

// EventBus é a interface para publicar eventos de domínio.
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
	PublishBatch(ctx context.Context, events []shared.DomainEvent) error
}

// SessionTimeoutResolver interface para resolver timeout de sessão
type SessionTimeoutResolver interface {
	ResolveForChannel(ctx context.Context, channelID uuid.UUID) (time.Duration, *uuid.UUID, error)
}

// TransactionManager gerencia transações de banco de dados.
type TransactionManager interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// ProcessInboundMessageUseCase handles processing of inbound messages
type ProcessInboundMessageUseCase struct {
	contactRepo          contact.Repository
	sessionRepo          domainsession.Repository
	messageRepo          domainmessage.Repository
	agentRepo            domainagent.Repository // ✅ NOVO: Para atribuir system agents
	contactEventRepo     contact_event.Repository
	eventBus             EventBus
	sessionManager       *sessionworkflow.SessionManager
	timeoutResolver      SessionTimeoutResolver
	db                   *gorm.DB // Usado apenas para invisible tracking detection
	ternaryEncoder       *tracking.TernaryEncoder
	messageDebouncerSvc  *MessageDebouncerService // 🎯 Debouncer para agrupamento de mensagens
	txManager            TransactionManager
	temporalClient       client.Client
	useSagaOrchestration bool
}

// NewProcessInboundMessageUseCase creates a new use case instance
func NewProcessInboundMessageUseCase(
	contactRepo contact.Repository,
	sessionRepo domainsession.Repository,
	messageRepo domainmessage.Repository,
	agentRepo domainagent.Repository, // ✅ NOVO: Para atribuir system agents
	contactEventRepo contact_event.Repository,
	eventBus EventBus,
	sessionManager *sessionworkflow.SessionManager,
	timeoutResolver SessionTimeoutResolver,
	db *gorm.DB, // Manter apenas para invisible tracking detection
	messageDebouncerSvc *MessageDebouncerService, // 🎯 Debouncer service
	txManager TransactionManager,
	temporalClient client.Client,
	useSagaOrchestration bool,
) *ProcessInboundMessageUseCase {
	return &ProcessInboundMessageUseCase{
		contactRepo:          contactRepo,
		sessionRepo:          sessionRepo,
		messageRepo:          messageRepo,
		agentRepo:            agentRepo, // ✅ NOVO
		contactEventRepo:     contactEventRepo,
		eventBus:             eventBus,
		sessionManager:       sessionManager,
		timeoutResolver:      timeoutResolver,
		db:                   db,
		ternaryEncoder:       tracking.NewTernaryEncoder(),
		messageDebouncerSvc:  messageDebouncerSvc,
		txManager:            txManager,
		temporalClient:       temporalClient,
		useSagaOrchestration: useSagaOrchestration,
	}
}

// Execute processes an inbound message following DDD best practices
// ✅ Saga Pattern (Choreography - Fast Path): Process Inbound Message Saga
//
// Este use case implementa uma Saga coreografada para processar mensagens WAHA.
// Todos os eventos publicados incluem correlation_id para rastreamento distribuído.
//
// **Fluxo da Saga:**
// 1. WAHA_RECEIVED → Mensagem recebida do webhook
// 2. CONTACT_FOUND_OR_CREATED → Contato criado/atualizado
// 3. SESSION_STARTED → Sessão iniciada/retomada
// 4. MESSAGE_CREATED → Mensagem persistida
// 5. TRACKING_CREATED → Tracking associado (se aplicável)
//
// **Compensação:** Se qualquer step falhar, eventos de compensação são disparados:
// - message.created ❌ → compensate.message.created (deleta mensagem)
// - session.started ❌ → compensate.session.started (encerra sessão)
// - contact.created ❌ → compensate.contact.created (deleta contato)
func (uc *ProcessInboundMessageUseCase) Execute(ctx context.Context, cmd ProcessInboundMessageCommand) error {
	// Feature flag: Use Saga Orchestration (Temporal) or transaction-based processing
	if uc.useSagaOrchestration && uc.temporalClient != nil {
		return uc.executeViaSaga(ctx, cmd)
	}
	return uc.executeViaTransaction(ctx, cmd)
}

// executeViaSaga executa o processamento via Temporal Saga Orchestration
func (uc *ProcessInboundMessageUseCase) executeViaSaga(ctx context.Context, cmd ProcessInboundMessageCommand) error {
	// Mapeia ProcessInboundMessageCommand para ProcessInboundMessageInput (Saga)
	input := sagaworkflow.ProcessInboundMessageInput{
		MessageID:        cmd.MessageID,
		ChannelMessageID: cmd.ChannelMessageID,
		FromPhone:        cmd.FromPhone,
		MessageText:      cmd.MessageText,
		Timestamp:        cmd.Timestamp,
		MessageType:      cmd.MessageType,
		MediaURL:         cmd.MediaURL,
		MediaType:        cmd.MediaType,
		ChannelID:        cmd.ChannelID,
		ProjectID:        cmd.ProjectID,
		CustomerID:       cmd.CustomerID,
		TenantID:         cmd.TenantID,
		ContactPhone:     cmd.ContactPhone,
		ContactName:      cmd.ContactName,
		ChannelTypeID:    cmd.ChannelTypeID,
		ContentType:      cmd.ContentType,
		Text:             cmd.Text,
		MediaMimetype:    cmd.MediaMimetype,
		TrackingData:     cmd.TrackingData,
		ReceivedAt:       cmd.ReceivedAt,
		Metadata:         cmd.Metadata,
		FromMe:           cmd.FromMe,
		IsGroupMessage:   cmd.IsGroupMessage,
		GroupExternalID:  cmd.GroupExternalID,
		Participant:      cmd.Participant,
		Mentions:         cmd.Mentions,
		ChatID:           cmd.ChatID,
	}

	// Inicia Temporal Workflow (Saga Orchestration)
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("process-inbound-%s", cmd.ChannelMessageID),
		TaskQueue: "message-processing",
	}

	we, err := uc.temporalClient.ExecuteWorkflow(ctx, workflowOptions, sagaworkflow.ProcessInboundMessageSaga, input)
	if err != nil {
		return fmt.Errorf("failed to start saga: %w", err)
	}

	// Aguarda conclusão (síncrono para webhooks)
	return we.Get(ctx, nil)
}

// executeViaTransaction executa o processamento via transação (método atual)
func (uc *ProcessInboundMessageUseCase) executeViaTransaction(ctx context.Context, cmd ProcessInboundMessageCommand) error {
	// 🎬 Inicia Saga: Process Inbound Message (Choreography - Fast Path)
	ctx = saga.WithSaga(ctx, string(saga.ProcessInboundMessageSaga))
	ctx = saga.WithTenantID(ctx, cmd.TenantID)

	correlationID, _ := saga.GetCorrelationID(ctx)
	fmt.Printf("🎬 Saga started: ProcessInboundMessage (correlation_id: %s)\n", correlationID)

	// ✅ TRANSAÇÃO ATÔMICA: Todas as operações de persistência + eventos juntos
	var c *domaincontact.Contact
	var s *domainsession.Session
	var m *domainmessage.Message

	err := uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Step 1: FindOrCreate Contact
		ctx = saga.NextStep(txCtx, saga.StepContactCreated)
		var err error
		c, err = uc.findOrCreateContact(txCtx, cmd)
		if err != nil {
			return fmt.Errorf("saga step failed [contact_created]: %w", err)
		}

		// Step 2: FindOrCreate Active Session
		ctx = saga.NextStep(txCtx, saga.StepSessionStarted)
		s, err = uc.findOrCreateSession(txCtx, c, cmd)
		if err != nil {
			return fmt.Errorf("saga step failed [session_started]: %w", err)
		}

		// Step 3: Create and Save Message
		ctx = saga.NextStep(txCtx, saga.StepMessageCreated)
		m, err = uc.createMessage(txCtx, c, s, cmd)
		if err != nil {
			return fmt.Errorf("saga step failed [message_created]: %w", err)
		}

		// Step 4: Record message in session (updates metrics)
		// ✅ FIX: Use cmd.ReceivedAt (timestamp histórico correto) ao invés de cmd.Timestamp (legado, sempre zero)
		if err := s.RecordMessage(true, cmd.ReceivedAt); err != nil {
			return fmt.Errorf("saga step failed [session_updated]: %w", err)
		}

		// Step 5: Update contact interaction timestamp
		c.RecordInteraction()

		// Step 6: Persist updates (usa transação do contexto)
		if err := uc.sessionRepo.Save(txCtx, s); err != nil {
			return fmt.Errorf("saga step failed [persist_session]: %w", err)
		}
		if err := uc.contactRepo.Save(txCtx, c); err != nil {
			return fmt.Errorf("saga step failed [persist_contact]: %w", err)
		}

		// Step 7: Publish domain events (choreography) - usa mesma transação
		// ✅ Todos os eventos publicados incluirão correlation_id automaticamente via DomainEventBus
		if err := uc.publishDomainEvents(txCtx, c, s, m); err != nil {
			return fmt.Errorf("saga step failed [publish_events]: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Step 6.5: 🎯 Process message debouncer (group or pass through) - FORA da transação
	// IMPORTANTE: Só processa debouncer se não for fromMe (mensagens enviadas pelo sistema não são agrupadas)
	if !cmd.FromMe && uc.messageDebouncerSvc != nil {
		if err := uc.messageDebouncerSvc.ProcessInboundMessage(ctx, m, cmd.ChannelID, s.ID()); err != nil {
			// Log erro mas não falha - debouncer é opcional
			fmt.Printf("⚠️  Warning: failed to process message debouncer: %v\n", err)
		}
	}

	// Step 8: Track ad conversion if applicable - FORA da transação (opcional)
	ctx = saga.NextStep(ctx, saga.StepTrackingCreated)
	if err := uc.trackAdConversion(ctx, c, s, cmd); err != nil {
		// Log but don't fail - tracking is optional
		fmt.Printf("⚠️  Warning: failed to track ad conversion: %v\n", err)
	}

	// Step 9: Detect and process invisible tracking code - FORA da transação (opcional)
	if err := uc.detectInvisibleTracking(ctx, c, s, cmd); err != nil {
		// Log but don't fail - invisible tracking is optional
		fmt.Printf("⚠️  Warning: failed to detect invisible tracking: %v\n", err)
	}

	fmt.Printf("✅ Saga completed: ProcessInboundMessage (contact=%s, session=%s, message=%s, correlation_id=%s)\n",
		c.ID(), s.ID(), m.ID(), correlationID)

	return nil
}

// findOrCreateContact busca contato por telefone ou cria novo
func (uc *ProcessInboundMessageUseCase) findOrCreateContact(ctx context.Context, cmd ProcessInboundMessageCommand) (*domaincontact.Contact, error) {
	fmt.Printf("🔍 [findOrCreateContact] Starting - phone: %s, projectID: %s\n", cmd.ContactPhone, cmd.ProjectID)

	// Busca por telefone
	existing, err := uc.contactRepo.FindByPhone(ctx, cmd.ProjectID, cmd.ContactPhone)

	fmt.Printf("🔍 [findOrCreateContact] FindByPhone result - existing: %v, err: %v\n", existing != nil, err)
	if err != nil {
		fmt.Printf("🔍 [findOrCreateContact] Error details - type: %T, value: %+v\n", err, err)
		// ✅ FIX: Use errors.Is() para funcionar com wrapped errors (*shared.DomainError)
		isNotFound := errors.Is(err, domaincontact.ErrContactNotFound)
		fmt.Printf("🔍 [findOrCreateContact] Is ErrContactNotFound (using errors.Is)? %v\n", isNotFound)

		// Se é um erro que NÃO é "not found", retornar o erro
		if !isNotFound {
			fmt.Printf("❌ [findOrCreateContact] Returning error because it's not ErrContactNotFound\n")
			return nil, err
		}
		// Se é "not found", continua para criar o contato (existing será nil)
		fmt.Printf("✅ [findOrCreateContact] Contact not found, will create new one\n")
	}

	if existing != nil {
		fmt.Printf("✅ [findOrCreateContact] Found existing contact: %s\n", existing.ID())
		// Atualiza nome se necessário
		if cmd.ContactName != "" && existing.Name() != cmd.ContactName {
			existing.UpdateName(cmd.ContactName)
		}
		return existing, nil
	}

	fmt.Printf("🆕 [findOrCreateContact] Creating new contact - name: %s, phone: %s\n", cmd.ContactName, cmd.ContactPhone)

	// Cria novo contato
	name := cmd.ContactName
	if name == "" {
		name = cmd.ContactPhone // Fallback
	}

	c, err := domaincontact.NewContact(cmd.ProjectID, cmd.TenantID, name)
	if err != nil {
		fmt.Printf("❌ [findOrCreateContact] Failed to create contact: %v\n", err)
		return nil, err
	}
	fmt.Printf("✅ [findOrCreateContact] Contact created: %s\n", c.ID())

	// Define telefone
	if err := c.SetPhone(cmd.ContactPhone); err != nil {
		fmt.Printf("❌ [findOrCreateContact] Failed to set phone: %v\n", err)
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}
	fmt.Printf("✅ [findOrCreateContact] Phone set: %s\n", cmd.ContactPhone)

	// Adiciona tag do canal
	c.AddTag("whatsapp")
	fmt.Printf("✅ [findOrCreateContact] Tag added: whatsapp\n")

	// Persiste
	fmt.Printf("💾 [findOrCreateContact] Saving contact to repository...\n")
	if err := uc.contactRepo.Save(ctx, c); err != nil {
		fmt.Printf("❌ [findOrCreateContact] Failed to save contact: %v (type: %T)\n", err, err)
		return nil, err
	}
	fmt.Printf("✅ [findOrCreateContact] Contact saved successfully: %s\n", c.ID())

	return c, nil
}

// findOrCreateSession busca sessão ativa ou cria nova
func (uc *ProcessInboundMessageUseCase) findOrCreateSession(ctx context.Context, c *domaincontact.Contact, cmd ProcessInboundMessageCommand) (*domainsession.Session, error) {
	// Busca sessão ativa para o contato + canal
	channelTypeID := &cmd.ChannelTypeID
	existing, err := uc.sessionRepo.FindActiveByContact(ctx, c.ID(), channelTypeID)
	if err != nil && err != domainsession.ErrSessionNotFound {
		return nil, err
	}

	if existing != nil {
		// Verifica timeout
		if existing.CheckTimeout() {
			// Sessão expirou, salva e cria nova
			if err := uc.sessionRepo.Save(ctx, existing); err != nil {
				return nil, err
			}
		} else {
			// Sessão ainda ativa - estende timeout via Temporal
			if uc.sessionManager != nil {
				err = uc.sessionManager.ExtendSessionTimeout(
					ctx,
					existing.ID(),
					1*time.Minute, // Reset para 1 min (teste)
				)
				if err != nil {
					fmt.Printf("Warning: failed to extend session timeout: %v\n", err)
				}
			}
			return existing, nil
		}
	}

	// 🎯 TIMEOUT HIERARCHY: Project (base) → Channel (override) → Pipeline (final override)
	// Usa SessionTimeoutResolver para seguir a hierarquia de forma elegante
	timeoutDuration, pipelineID, err := uc.timeoutResolver.ResolveForChannel(ctx, cmd.ChannelID)
	if err != nil {
		fmt.Printf("Warning: failed to resolve timeout, using default 4h: %v\n", err)
		timeoutDuration = 4 * time.Hour // 240 minutos para máxima consolidação
		pipelineID = nil
	}

	timeoutMinutes := int(timeoutDuration.Minutes())
	fmt.Printf("⏱️  Resolved session timeout: %d minutes (pipelineID: %v)\n", timeoutMinutes, pipelineID)

	var s *domainsession.Session

	// 🎯 Cria Session com ou sem pipeline baseado no resultado do resolver
	if pipelineID != nil && *pipelineID != uuid.Nil {
		// ✅ Pipeline encontrado: Cria Session COM pipeline_id
		s, err = domainsession.NewSessionWithPipeline(
			c.ID(),
			cmd.TenantID,
			channelTypeID,
			*pipelineID,
			timeoutDuration,
		)
		if err != nil {
			return nil, err
		}
	} else {
		// ⚠️ SEM PIPELINE ATIVO: Cria Session SEM pipeline_id
		// Sessão será criada apenas para agrupar mensagens, sem associação a pipeline
		fmt.Printf("⚠️  No active pipeline found for project %s, creating session without pipeline association\n", c.ProjectID())

		s, err = domainsession.NewSession(
			c.ID(),
			cmd.TenantID,
			channelTypeID,
			timeoutDuration,
		)
		if err != nil {
			return nil, err
		}
	}

	// ✅ Persiste IMEDIATAMENTE (ANTES de createMessage usar s.ID())
	if err := uc.sessionRepo.Save(ctx, s); err != nil {
		return nil, err
	}

	fmt.Printf("✅ Session persisted with ID: %s, tenant_id context will be used\n", s.ID())

	// Inicia workflow Temporal para gerenciar o ciclo de vida da sessão
	if uc.sessionManager != nil {
		err = uc.sessionManager.StartSessionLifecycle(
			ctx,
			s.ID(),
			c.ID(),
			cmd.TenantID,
			channelTypeID,
			timeoutDuration,
		)
		if err != nil {
			// Log erro mas não falha o processamento da mensagem
			fmt.Printf("Warning: failed to start session lifecycle workflow: %v\n", err)
		}
	}

	return s, nil
}

// createMessage cria e persiste a mensagem
func (uc *ProcessInboundMessageUseCase) createMessage(ctx context.Context, c *domaincontact.Contact, s *domainsession.Session, cmd ProcessInboundMessageCommand) (*domainmessage.Message, error) {
	// 🎯 DEDUPLICAÇÃO: Verifica se mensagem já existe pelo channel_message_id
	if cmd.ChannelMessageID != "" {
		existingMsg, err := uc.messageRepo.FindByChannelMessageID(ctx, cmd.ChannelMessageID)
		if err == nil && existingMsg != nil {
			// Mensagem já existe, retorna a existente
			fmt.Printf("⚠️  Message already exists (channel_message_id=%s), skipping creation\n", cmd.ChannelMessageID)
			return existingMsg, nil
		}
		// Se não encontrou ou deu erro, continua criação
	}

	// Parse content type
	contentType, err := domainmessage.ParseContentType(cmd.ContentType)
	if err != nil {
		return nil, fmt.Errorf("invalid content type: %w", err)
	}

	// Cria mensagem (usa cmd.FromMe para determinar direção)
	m, err := domainmessage.NewMessage(c.ID(), cmd.ProjectID, cmd.CustomerID, contentType, cmd.FromMe)
	if err != nil {
		return nil, err
	}

	// Associa ao canal (OBRIGATÓRIO)
	m.AssignToChannel(cmd.ChannelID, &cmd.ChannelTypeID)

	// ✅ Associa à sessão (session JÁ foi persistida em findOrCreateSession linha 442)
	sessionID := s.ID()
	fmt.Printf("🔗 Assigning message to session: %s\n", sessionID)
	m.AssignToSession(sessionID)

	// Define channel_message_id (ID externo do WhatsApp)
	if cmd.ChannelMessageID != "" {
		m.SetChannelMessageID(cmd.ChannelMessageID)
	}

	// Define conteúdo
	if contentType.IsText() && cmd.Text != "" {
		if err := m.SetText(cmd.Text); err != nil {
			return nil, err
		}
	}

	// Define mídia se aplicável
	if contentType.IsMedia() && cmd.MediaURL != "" {
		if err := m.SetMediaContent(cmd.MediaURL, cmd.MediaMimetype); err != nil {
			return nil, err
		}
	}

	// ✅ Define menções se aplicável
	if len(cmd.Mentions) > 0 {
		m.SetMentions(cmd.Mentions)
		fmt.Printf("✅ Message mentions set: %d mentions\n", len(cmd.Mentions))
	}

	// ✅ OBRIGATÓRIO: Atribui system agent baseado na source da mensagem
	// Toda mensagem DEVE ter um agente atribuído (invariante de domínio)
	systemAgentID := uc.getSystemAgentForSource(m.Source())
	if err := m.AssignAgent(systemAgentID); err != nil {
		return nil, fmt.Errorf("failed to assign agent: %w", err)
	}
	fmt.Printf("✅ System agent assigned: %s (source: %s)\n", systemAgentID, m.Source())

	// ✅ Persiste (session JÁ está no banco, FK deve funcionar)
	fmt.Printf("💾 Saving message with session_id=%s to repository...\n", m.SessionID())
	if err := uc.messageRepo.Save(ctx, m); err != nil {
		fmt.Printf("❌ Failed to save message: %v (session_id=%s)\n", err, m.SessionID())
		return nil, err
	}
	fmt.Printf("✅ Message saved successfully: %s (session_id=%s)\n", m.ID(), m.SessionID())

	return m, nil
}

// createContactEvent was removed - messages should not create contact events
// Only important contact-related events (status changes, pipeline movements, assignments, etc)
// should create contact events for streaming

// publishDomainEvents publica todos os eventos de domínio
func (uc *ProcessInboundMessageUseCase) publishDomainEvents(ctx context.Context, c *domaincontact.Contact, s *domainsession.Session, m *domainmessage.Message) error {
	var events []shared.DomainEvent

	// Coleta eventos do contato
	for _, e := range c.DomainEvents() {
		events = append(events, e)
	}
	c.ClearEvents()

	// Coleta eventos da sessão
	for _, e := range s.DomainEvents() {
		events = append(events, e)
	}
	s.ClearEvents()

	// Coleta eventos da mensagem
	for _, e := range m.DomainEvents() {
		events = append(events, e)
	}
	m.ClearEvents()

	// Publica em batch
	if len(events) > 0 {
		return uc.eventBus.PublishBatch(ctx, events)
	}

	return nil
}

// trackAdConversion rastreia conversão de anúncio se aplicável
func (uc *ProcessInboundMessageUseCase) trackAdConversion(ctx context.Context, c *domaincontact.Contact, s *domainsession.Session, cmd ProcessInboundMessageCommand) error {
	// Verifica se tem dados de tracking
	isFromAd, ok := cmd.Metadata["is_from_ad"].(bool)
	if !ok || !isFromAd {
		return nil // Não é de ad
	}

	// Extrai tracking data
	trackingData := cmd.TrackingData
	if len(trackingData) == 0 {
		return nil
	}

	// Converte map[string]interface{} para map[string]string
	trackingDataStr := make(map[string]string)
	for k, v := range trackingData {
		if str, ok := v.(string); ok {
			trackingDataStr[k] = str
		}
	}

	// Cria evento de conversão
	conversionEvent := domaincontact.NewAdConversionTrackedEvent(
		c.ID(),
		s.ID(),
		cmd.TenantID,
		trackingDataStr,
	)

	// Publica evento
	return uc.eventBus.Publish(ctx, conversionEvent)
}

// truncate trunca string para max caracteres
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// detectInvisibleTracking detecta código invisível ternário na mensagem e cria tracking automaticamente
func (uc *ProcessInboundMessageUseCase) detectInvisibleTracking(ctx context.Context, c *domaincontact.Contact, s *domainsession.Session, cmd ProcessInboundMessageCommand) error {
	// Verifica se mensagem tem texto
	if cmd.Text == "" {
		return nil // Não é mensagem de texto
	}

	// Verifica se mensagem contém código invisível
	if !uc.ternaryEncoder.HasInvisibleCode(cmd.Text) {
		return nil // Não tem código invisível
	}

	// Tenta decodificar mensagem
	trackingIDPtr, cleanMessage, err := uc.ternaryEncoder.DecodeMessage(cmd.Text)
	if err != nil || trackingIDPtr == nil {
		// Não conseguiu decodificar, ignora
		return nil
	}

	trackingID := *trackingIDPtr

	fmt.Printf("🔍 Invisible tracking code detected: tracking_id=%d, contact=%s, clean_message=%s\n",
		trackingID, c.ID(), truncate(cleanMessage, 50))

	// Busca o tracking no banco pelo ID (convertido de base 3 para base 10)
	var trackingExists bool
	err = uc.db.Raw(`
		SELECT EXISTS(SELECT 1 FROM trackings WHERE id = ?)
	`, trackingID).Scan(&trackingExists).Error

	if err != nil {
		return fmt.Errorf("failed to check tracking existence: %w", err)
	}

	if !trackingExists {
		fmt.Printf("⚠️  Tracking ID %d not found in database, skipping association\n", trackingID)
		return nil
	}

	// Associa tracking ao contato e sessão
	err = uc.db.Exec(`
		UPDATE trackings
		SET
			contact_id = ?,
			session_id = ?,
			updated_at = NOW()
		WHERE id = ? AND contact_id IS NULL
	`, c.ID(), s.ID(), trackingID).Error

	if err != nil {
		return fmt.Errorf("failed to associate tracking: %w", err)
	}

	fmt.Printf("✅ Tracking %d associated with contact %s and session %s\n",
		trackingID, c.ID(), s.ID())

	return nil
}

// getSystemAgentForSource retorna o system agent apropriado baseado na source da mensagem
// Implementa a estratégia de atribuição automática de agentes:
// - Cada source tem um system agent correspondente
// - Garante que TODA mensagem tenha um agente (invariante de domínio)
// - Fallback: SystemAgentDefault para sources não mapeadas
func (uc *ProcessInboundMessageUseCase) getSystemAgentForSource(source domainmessage.Source) uuid.UUID {
	switch source {
	case domainmessage.SourceHistoryImport:
		// Mensagens importadas do histórico (WAHA history import, etc)
		return domainagent.SystemAgentDefault

	case domainmessage.SourceWebhook:
		// Respostas automáticas via webhook
		return domainagent.SystemAgentWebhook

	case domainmessage.SourceBroadcast:
		// Campanhas broadcast
		return domainagent.SystemAgentBroadcast

	case domainmessage.SourceSequence:
		// Sequências de automação
		return domainagent.SystemAgentSequence

	case domainmessage.SourceTrigger:
		// Triggers/regras de pipeline
		return domainagent.SystemAgentTrigger

	case domainmessage.SourceScheduled:
		// Mensagens agendadas
		return domainagent.SystemAgentScheduled

	case domainmessage.SourceTest:
		// Testes E2E e envios de teste
		return domainagent.SystemAgentTest

	case domainmessage.SourceManual:
		// Mensagens manuais enviadas por agentes humanos
		// NOTA: ProcessInboundMessageUseCase processa mensagens INBOUND (recebidas)
		// Mensagens manuais OUTBOUND (enviadas) são processadas por SendMessageCommand
		// e já têm agente autenticado. Se chegou aqui, é uma mensagem recebida
		// sem source definida, então usamos Default como fallback.
		return domainagent.SystemAgentDefault

	case domainmessage.SourceBot:
		// Bot/AI responses
		return domainagent.SystemAgentDefault

	case domainmessage.SourceSystem:
		// Sistema interno
		return domainagent.SystemAgentDefault

	default:
		// Fallback para sources desconhecidas ou não mapeadas
		fmt.Printf("⚠️  Unknown message source '%s', using SystemAgentDefault\n", source)
		return domainagent.SystemAgentDefault
	}
}
