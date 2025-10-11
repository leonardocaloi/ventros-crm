package message

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/contact"
	domaincontact "github.com/caloi/ventros-crm/internal/domain/contact"
	contact_event "github.com/caloi/ventros-crm/internal/domain/contact_event"
	domainmessage "github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/caloi/ventros-crm/internal/domain/saga"
	domainsession "github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/caloi/ventros-crm/internal/domain/tracking"
	sessionworkflow "github.com/caloi/ventros-crm/internal/workflows/session"
	"github.com/google/uuid"
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
	IsGroupMessage bool     // Se a mensagem é de um grupo do WhatsApp
	GroupExternalID string  // ID externo do grupo (ex: "123456789@g.us")
	Participant    string   // Em grupos: quem ENVIOU a mensagem (participant)
	Mentions       []string // IDs dos usuários mencionados (@marcados) no formato WAHA
	ChatID         *uuid.UUID // ID do Chat (grupo ou individual) - será preenchido durante processamento
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

// ProcessInboundMessageUseCase handles processing of inbound messages
type ProcessInboundMessageUseCase struct {
	contactRepo         contact.Repository
	sessionRepo         domainsession.Repository
	messageRepo         domainmessage.Repository
	contactEventRepo    contact_event.Repository
	eventBus            EventBus
	sessionManager      *sessionworkflow.SessionManager
	timeoutResolver     SessionTimeoutResolver
	db                  *gorm.DB // Usado apenas para invisible tracking detection
	ternaryEncoder      *tracking.TernaryEncoder
	messageDebouncerSvc *MessageDebouncerService // 🎯 Debouncer para agrupamento de mensagens
}

// NewProcessInboundMessageUseCase creates a new use case instance
func NewProcessInboundMessageUseCase(
	contactRepo contact.Repository,
	sessionRepo domainsession.Repository,
	messageRepo domainmessage.Repository,
	contactEventRepo contact_event.Repository,
	eventBus EventBus,
	sessionManager *sessionworkflow.SessionManager,
	timeoutResolver SessionTimeoutResolver,
	db *gorm.DB, // Manter apenas para invisible tracking detection
	messageDebouncerSvc *MessageDebouncerService, // 🎯 Debouncer service
) *ProcessInboundMessageUseCase {
	return &ProcessInboundMessageUseCase{
		contactRepo:         contactRepo,
		sessionRepo:         sessionRepo,
		messageRepo:         messageRepo,
		contactEventRepo:    contactEventRepo,
		eventBus:            eventBus,
		sessionManager:      sessionManager,
		timeoutResolver:     timeoutResolver,
		db:                  db,
		ternaryEncoder:      tracking.NewTernaryEncoder(),
		messageDebouncerSvc: messageDebouncerSvc,
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
	// 🎬 Inicia Saga: Process Inbound Message (Choreography - Fast Path)
	ctx = saga.WithSaga(ctx, string(saga.ProcessInboundMessageSaga))
	ctx = saga.WithTenantID(ctx, cmd.TenantID)

	correlationID, _ := saga.GetCorrelationID(ctx)
	fmt.Printf("🎬 Saga started: ProcessInboundMessage (correlation_id: %s)\n", correlationID)

	// Step 1: FindOrCreate Contact
	ctx = saga.NextStep(ctx, saga.StepContactCreated)
	c, err := uc.findOrCreateContact(ctx, cmd)
	if err != nil {
		return fmt.Errorf("saga step failed [contact_created]: %w", err)
	}

	// Step 2: FindOrCreate Active Session
	ctx = saga.NextStep(ctx, saga.StepSessionStarted)
	s, err := uc.findOrCreateSession(ctx, c, cmd)
	if err != nil {
		return fmt.Errorf("saga step failed [session_started]: %w", err)
	}

	// Step 3: Create and Save Message
	ctx = saga.NextStep(ctx, saga.StepMessageCreated)
	m, err := uc.createMessage(ctx, c, s, cmd)
	if err != nil {
		return fmt.Errorf("saga step failed [message_created]: %w", err)
	}

	// Step 4: Record message in session (updates metrics)
	if err := s.RecordMessage(true, cmd.Timestamp); err != nil {
		return fmt.Errorf("saga step failed [session_updated]: %w", err)
	}

	// Step 5: Update contact interaction timestamp
	c.RecordInteraction()

	// Step 6: Persist updates
	if err := uc.sessionRepo.Save(ctx, s); err != nil {
		return fmt.Errorf("saga step failed [persist_session]: %w", err)
	}
	if err := uc.contactRepo.Save(ctx, c); err != nil {
		return fmt.Errorf("saga step failed [persist_contact]: %w", err)
	}

	// Step 6.5: 🎯 Process message debouncer (group or pass through)
	// IMPORTANTE: Só processa debouncer se não for fromMe (mensagens enviadas pelo sistema não são agrupadas)
	if !cmd.FromMe && uc.messageDebouncerSvc != nil {
		if err := uc.messageDebouncerSvc.ProcessInboundMessage(ctx, m, cmd.ChannelID, s.ID()); err != nil {
			// Log erro mas não falha - debouncer é opcional
			fmt.Printf("⚠️  Warning: failed to process message debouncer: %v\n", err)
		}
	}

	// Step 7: Publish domain events (choreography)
	// ✅ Todos os eventos publicados incluirão correlation_id automaticamente via DomainEventBus
	if err := uc.publishDomainEvents(ctx, c, s, m); err != nil {
		// Log but don't fail - event publishing is async
		fmt.Printf("⚠️  Warning: failed to publish domain events: %v\n", err)
	}

	// Step 8: Track ad conversion if applicable
	ctx = saga.NextStep(ctx, saga.StepTrackingCreated)
	if err := uc.trackAdConversion(ctx, c, s, cmd); err != nil {
		// Log but don't fail - tracking is optional
		fmt.Printf("⚠️  Warning: failed to track ad conversion: %v\n", err)
	}

	// Step 9: Detect and process invisible tracking code
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
	// Busca por telefone
	existing, err := uc.contactRepo.FindByPhone(ctx, cmd.ProjectID, cmd.ContactPhone)
	if err != nil && err != domaincontact.ErrContactNotFound {
		return nil, err
	}

	if existing != nil {
		// Atualiza nome se necessário
		if cmd.ContactName != "" && existing.Name() != cmd.ContactName {
			existing.UpdateName(cmd.ContactName)
		}
		return existing, nil
	}

	// Cria novo contato
	name := cmd.ContactName
	if name == "" {
		name = cmd.ContactPhone // Fallback
	}

	c, err := domaincontact.NewContact(cmd.ProjectID, cmd.TenantID, name)
	if err != nil {
		return nil, err
	}

	// Define telefone
	if err := c.SetPhone(cmd.ContactPhone); err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	// Adiciona tag do canal
	c.AddTag("whatsapp")

	// Persiste
	if err := uc.contactRepo.Save(ctx, c); err != nil {
		return nil, err
	}

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
		fmt.Printf("Warning: failed to resolve timeout, using default 30 min: %v\n", err)
		timeoutDuration = 30 * time.Minute
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

	// Persiste
	if err := uc.sessionRepo.Save(ctx, s); err != nil {
		return nil, err
	}

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

	// Associa à sessão
	m.AssignToSession(s.ID())

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

	// Persiste
	if err := uc.messageRepo.Save(ctx, m); err != nil {
		return nil, err
	}

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
