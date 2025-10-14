package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/application/shared"
	domaincontact "github.com/ventros/crm/internal/domain/crm/contact"
	domainmessage "github.com/ventros/crm/internal/domain/crm/message"
	domainsession "github.com/ventros/crm/internal/domain/crm/session"
)

// Activities é o container para todas as activities da Saga.
// Injeta dependências (repos, event bus, etc) no construtor.
type Activities struct {
	contactRepo     domaincontact.Repository
	sessionRepo     domainsession.Repository
	messageRepo     domainmessage.Repository
	txManager       shared.TransactionManager
	eventBus        EventBus
	timeoutResolver SessionTimeoutResolver
}

// EventBus interface para publicar eventos.
type EventBus interface {
	Publish(ctx context.Context, event interface{}) error
	PublishBatch(ctx context.Context, events []interface{}) error
}

// SessionTimeoutResolver resolve timeout de sessão.
type SessionTimeoutResolver interface {
	ResolveForChannel(ctx context.Context, channelID uuid.UUID) (time.Duration, *uuid.UUID, error)
}

// NewActivities cria uma nova instância das activities.
func NewActivities(
	contactRepo domaincontact.Repository,
	sessionRepo domainsession.Repository,
	messageRepo domainmessage.Repository,
	txManager shared.TransactionManager,
	eventBus EventBus,
	timeoutResolver SessionTimeoutResolver,
) *Activities {
	return &Activities{
		contactRepo:     contactRepo,
		sessionRepo:     sessionRepo,
		messageRepo:     messageRepo,
		txManager:       txManager,
		eventBus:        eventBus,
		timeoutResolver: timeoutResolver,
	}
}

// ============================================================================
// FORWARD ACTIVITIES (Saga Steps)
// ============================================================================

// FindOrCreateContactActivity busca ou cria um contato.
func (a *Activities) FindOrCreateContactActivity(ctx context.Context, input ProcessInboundMessageInput) (*ContactCreatedResult, error) {
	var result ContactCreatedResult

	err := a.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Busca por telefone
		existing, err := a.contactRepo.FindByPhone(txCtx, input.ProjectID, input.ContactPhone)
		if err != nil && err != domaincontact.ErrContactNotFound {
			return fmt.Errorf("failed to find contact: %w", err)
		}

		if existing != nil {
			// Contato já existe - atualiza nome se necessário
			if input.ContactName != "" && existing.Name() != input.ContactName {
				existing.UpdateName(input.ContactName)
				if err := a.contactRepo.Save(txCtx, existing); err != nil {
					return fmt.Errorf("failed to update contact: %w", err)
				}
			}
			result.ContactID = existing.ID()
			result.WasCreated = false
			return nil
		}

		// Cria novo contato
		name := input.ContactName
		if name == "" {
			name = input.ContactPhone // Fallback
		}

		newContact, err := domaincontact.NewContact(input.ProjectID, input.TenantID, name)
		if err != nil {
			return fmt.Errorf("failed to create contact: %w", err)
		}

		// Define telefone
		if err := newContact.SetPhone(input.ContactPhone); err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}

		// Adiciona tag do canal
		newContact.AddTag("whatsapp")

		// Persiste
		if err := a.contactRepo.Save(txCtx, newContact); err != nil {
			return fmt.Errorf("failed to save contact: %w", err)
		}

		result.ContactID = newContact.ID()
		result.WasCreated = true
		return nil
	})

	return &result, err
}

// FindOrCreateSessionActivity busca ou cria uma sessão ativa.
func (a *Activities) FindOrCreateSessionActivity(ctx context.Context, contactID uuid.UUID, input ProcessInboundMessageInput) (*SessionCreatedResult, error) {
	var result SessionCreatedResult

	err := a.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Busca sessão ativa
		channelTypeID := &input.ChannelTypeID
		existing, err := a.sessionRepo.FindActiveByContact(txCtx, contactID, channelTypeID)
		if err != nil && err != domainsession.ErrSessionNotFound {
			return fmt.Errorf("failed to find session: %w", err)
		}

		if existing != nil {
			// Verifica timeout
			if existing.CheckTimeout() {
				// Sessão expirou, salva e cria nova
				if err := a.sessionRepo.Save(txCtx, existing); err != nil {
					return fmt.Errorf("failed to save expired session: %w", err)
				}
			} else {
				// Sessão ainda ativa
				result.SessionID = existing.ID()
				result.WasCreated = false
				return nil
			}
		}

		// Resolve timeout usando hierarquia (Project → Channel → Pipeline)
		timeoutDuration, pipelineID, err := a.timeoutResolver.ResolveForChannel(ctx, input.ChannelID)
		if err != nil {
			// Fallback para 30 minutos
			timeoutDuration = 30 * time.Minute
			pipelineID = nil
		}

		var newSession *domainsession.Session

		// Cria Session com ou sem pipeline
		if pipelineID != nil && *pipelineID != uuid.Nil {
			newSession, err = domainsession.NewSessionWithPipeline(
				contactID,
				input.TenantID,
				channelTypeID,
				*pipelineID,
				timeoutDuration,
			)
		} else {
			newSession, err = domainsession.NewSession(
				contactID,
				input.TenantID,
				channelTypeID,
				timeoutDuration,
			)
		}

		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Persiste
		if err := a.sessionRepo.Save(txCtx, newSession); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}

		result.SessionID = newSession.ID()
		result.WasCreated = true
		return nil
	})

	return &result, err
}

// CreateMessageActivity cria uma mensagem.
func (a *Activities) CreateMessageActivity(ctx context.Context, contactID, sessionID uuid.UUID, input ProcessInboundMessageInput) (*MessageCreatedResult, error) {
	var result MessageCreatedResult

	err := a.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Verifica deduplicação
		if input.ChannelMessageID != "" {
			existingMsg, err := a.messageRepo.FindByChannelMessageID(txCtx, input.ChannelMessageID)
			if err == nil && existingMsg != nil {
				// Mensagem já existe, retorna a existente
				result.MessageID = existingMsg.ID()
				return nil
			}
		}

		// Parse content type
		contentType, err := domainmessage.ParseContentType(input.ContentType)
		if err != nil {
			return fmt.Errorf("invalid content type: %w", err)
		}

		// Cria mensagem
		msg, err := domainmessage.NewMessage(contactID, input.ProjectID, input.CustomerID, contentType, input.FromMe)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		// Associa ao canal e sessão
		msg.AssignToChannel(input.ChannelID, &input.ChannelTypeID)
		msg.AssignToSession(sessionID)

		// Define channel_message_id
		if input.ChannelMessageID != "" {
			msg.SetChannelMessageID(input.ChannelMessageID)
		}

		// Define conteúdo
		if contentType.IsText() && input.Text != "" {
			if err := msg.SetText(input.Text); err != nil {
				return fmt.Errorf("failed to set text: %w", err)
			}
		}

		// Define mídia
		if contentType.IsMedia() && input.MediaURL != "" {
			if err := msg.SetMediaContent(input.MediaURL, input.MediaMimetype); err != nil {
				return fmt.Errorf("failed to set media: %w", err)
			}
		}

		// Define menções
		if len(input.Mentions) > 0 {
			msg.SetMentions(input.Mentions)
		}

		// Persiste
		if err := a.messageRepo.Save(txCtx, msg); err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		result.MessageID = msg.ID()
		return nil
	})

	return &result, err
}

// PublishDomainEventsActivity publica eventos de domínio via Outbox.
func (a *Activities) PublishDomainEventsActivity(ctx context.Context, state SagaState) error {
	// Busca os agregados para coletar eventos
	contact, err := a.contactRepo.FindByID(ctx, state.ContactID)
	if err != nil {
		return fmt.Errorf("failed to find contact: %w", err)
	}

	session, err := a.sessionRepo.FindByID(ctx, state.SessionID)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}

	message, err := a.messageRepo.FindByID(ctx, state.MessageID)
	if err != nil {
		return fmt.Errorf("failed to find message: %w", err)
	}

	// Atualiza session metrics
	if err := session.RecordMessage(true, time.Now()); err != nil {
		return fmt.Errorf("failed to record message in session: %w", err)
	}

	// Atualiza contact interaction
	contact.RecordInteraction()

	// Persiste atualizações + publica eventos atomicamente
	return a.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Persiste atualizações
		if err := a.sessionRepo.Save(txCtx, session); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		if err := a.contactRepo.Save(txCtx, contact); err != nil {
			return fmt.Errorf("failed to save contact: %w", err)
		}

		// Publica eventos
		var allEvents []interface{}
		for _, e := range contact.DomainEvents() {
			allEvents = append(allEvents, e)
		}
		for _, e := range session.DomainEvents() {
			allEvents = append(allEvents, e)
		}
		for _, e := range message.DomainEvents() {
			allEvents = append(allEvents, e)
		}

		if len(allEvents) > 0 {
			if err := a.eventBus.PublishBatch(txCtx, allEvents); err != nil {
				return fmt.Errorf("failed to publish events: %w", err)
			}
		}

		// Limpa eventos
		contact.ClearEvents()
		session.ClearEvents()
		message.ClearEvents()

		return nil
	})
}

// ProcessMessageDebouncerActivity processa debouncer (agrupamento de mensagens).
// Optional - best effort.
func (a *Activities) ProcessMessageDebouncerActivity(ctx context.Context, messageID, channelID, sessionID uuid.UUID) error {
	// TODO: Implementar lógica de debouncer
	// Por enquanto, apenas retorna sucesso
	return nil
}

// TrackAdConversionActivity rastreia conversão de anúncio.
// Optional - best effort.
func (a *Activities) TrackAdConversionActivity(ctx context.Context, state SagaState, trackingData map[string]interface{}) error {
	// TODO: Implementar tracking de conversão
	// Por enquanto, apenas retorna sucesso
	return nil
}

// ============================================================================
// COMPENSATION ACTIVITIES (Rollback)
// ============================================================================

// DeleteContactActivity soft-delete de contato (compensação).
func (a *Activities) DeleteContactActivity(ctx context.Context, contactID uuid.UUID) error {
	return a.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		contact, err := a.contactRepo.FindByID(txCtx, contactID)
		if err != nil {
			// Se não encontrou, considera sucesso (idempotência)
			if err == domaincontact.ErrContactNotFound {
				return nil
			}
			return fmt.Errorf("failed to find contact for deletion: %w", err)
		}

		// Soft delete
		contact.Delete()

		if err := a.contactRepo.Save(txCtx, contact); err != nil {
			return fmt.Errorf("failed to delete contact: %w", err)
		}

		return nil
	})
}

// CloseSessionActivity fecha sessão forçadamente (compensação).
func (a *Activities) CloseSessionActivity(ctx context.Context, sessionID uuid.UUID, reason string) error {
	return a.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		session, err := a.sessionRepo.FindByID(txCtx, sessionID)
		if err != nil {
			// Se não encontrou, considera sucesso (idempotência)
			if err == domainsession.ErrSessionNotFound {
				return nil
			}
			return fmt.Errorf("failed to find session for closing: %w", err)
		}

		// Fecha sessão com razão específica
		if err := session.End(domainsession.EndReasonSagaRollback); err != nil {
			return fmt.Errorf("failed to close session: %w", err)
		}

		if err := a.sessionRepo.Save(txCtx, session); err != nil {
			return fmt.Errorf("failed to save closed session: %w", err)
		}

		return nil
	})
}

// DeleteMessageActivity deleta mensagem (compensação).
// NOTA: Como o repository Message não tem método Delete() ainda,
// e a saga usa transações por activity, a compensação é idempotente:
// - Se CreateMessage falhou → rollback automático, nada a compensar
// - Se PublishEvents falhou → mensagem existe mas eventos não foram publicados
// TODO: Adicionar método Delete() ao repository quando necessário
func (a *Activities) DeleteMessageActivity(ctx context.Context, messageID uuid.UUID) error {
	// Compensação simplificada: não faz nada por enquanto
	// Em produção, adicionar método Delete() ao repository
	return nil
}
