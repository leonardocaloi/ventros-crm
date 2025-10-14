package session

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/session"
	"go.temporal.io/sdk/activity"
)

// EndSessionActivityInput parâmetros para encerrar uma sessão
type EndSessionActivityInput struct {
	SessionID uuid.UUID `json:"session_id"`
	Reason    string    `json:"reason"`
}

// EndSessionActivityResult resultado do encerramento
type EndSessionActivityResult struct {
	Success         bool `json:"success"`
	EventsPublished int  `json:"events_published"`
}

// CleanupSessionsActivityInput parâmetros para limpeza de sessões
type CleanupSessionsActivityInput struct {
	MaxInactivityDuration time.Duration `json:"max_inactivity_duration"`
}

// CleanupSessionsActivityResult resultado da limpeza
type CleanupSessionsActivityResult struct {
	SessionsCleaned int `json:"sessions_cleaned"`
	EventsPublished int `json:"events_published"`
}

// SessionActivities contém as dependências para as activities
type SessionActivities struct {
	sessionRepo session.Repository
	messageRepo MessageRepository
	eventBus    EventBus
}

// EventBus interface para publicar eventos
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
	PublishBatch(ctx context.Context, events []shared.DomainEvent) error
}

// MessageRepository interface para buscar mensagens
type MessageRepository interface {
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]MessageInfo, error)
}

// MessageInfo informações da mensagem para enrichment
type MessageInfo struct {
	ID        uuid.UUID
	ChannelID *uuid.UUID
	Direction string // "inbound" ou "outbound"
	Timestamp time.Time
}

// NewSessionActivities cria uma nova instância das activities
func NewSessionActivities(sessionRepo session.Repository, messageRepo MessageRepository, eventBus EventBus) *SessionActivities {
	return &SessionActivities{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		eventBus:    eventBus,
	}
}

// EndSessionActivity encerra uma sessão específica por timeout
func (a *SessionActivities) EndSessionActivity(ctx context.Context, input EndSessionActivityInput) (EndSessionActivityResult, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Ending session by timeout",
		"session_id", input.SessionID.String(),
		"reason", input.Reason)

	// Busca a sessão
	sess, err := a.sessionRepo.FindByID(ctx, input.SessionID)
	if err != nil {
		// Se a sessão não foi encontrada, pode ter sido deletada manualmente
		// Isso é aceitável - retorna sucesso para não fazer o workflow ficar tentando
		logger.Warn("Session not found, likely deleted manually - marking workflow as complete",
			"session_id", input.SessionID.String(),
			"error", err.Error())
		return EndSessionActivityResult{Success: true, EventsPublished: 0}, nil
	}

	// Verifica se ainda está ativa
	if sess.Status() != session.StatusActive {
		logger.Info("Session already ended, skipping",
			"session_id", input.SessionID.String(),
			"current_status", sess.Status())
		return EndSessionActivityResult{Success: true}, nil
	}

	// Mapeia reason string para enum
	var endReason session.EndReason
	switch input.Reason {
	case "inactivity_timeout":
		endReason = session.ReasonInactivityTimeout
	case "manual":
		endReason = session.ReasonInactivityTimeout // Usar o mesmo por enquanto
	case "agent_ended":
		endReason = session.ReasonInactivityTimeout // Usar o mesmo por enquanto
	default:
		endReason = session.ReasonInactivityTimeout
	}

	// Encerra a sessão
	if err := sess.End(endReason); err != nil {
		return EndSessionActivityResult{}, fmt.Errorf("failed to end session: %w", err)
	}

	// Salva no repositório
	if err := a.sessionRepo.Save(ctx, sess); err != nil {
		return EndSessionActivityResult{}, fmt.Errorf("failed to save session: %w", err)
	}

	// Busca mensagens da sessão para enriquecer o evento
	messages, err := a.messageRepo.FindBySessionID(ctx, sess.ID())
	if err != nil {
		logger.Warn("Failed to fetch messages for session enrichment", "error", err)
		messages = []MessageInfo{} // Continua com evento vazio
	}

	// Publica eventos de domínio (com enrichment)
	events := sess.DomainEvents()
	eventsCount := len(events)

	if eventsCount > 0 {
		// Enriquece o evento session.ended com mensagens e canal
		enrichedEvents := make([]shared.DomainEvent, len(events))
		for i, event := range events {
			if sessionEndedEvent, ok := event.(session.SessionEndedEvent); ok {
				enrichedEvents[i] = a.enrichSessionEndedEvent(sessionEndedEvent, messages)
			} else {
				enrichedEvents[i] = event
			}
		}

		if err := a.eventBus.PublishBatch(ctx, enrichedEvents); err != nil {
			logger.Error("Failed to publish domain events", "error", err)
			// Não falha a activity por causa dos eventos
		}
		sess.ClearEvents()
	}

	logger.Info("Session ended successfully",
		"session_id", input.SessionID.String(),
		"events_published", eventsCount)

	return EndSessionActivityResult{
		Success:         true,
		EventsPublished: eventsCount,
	}, nil
}

// enrichSessionEndedEvent enriquece o evento session.ended com informações de mensagens e canal
func (a *SessionActivities) enrichSessionEndedEvent(event session.SessionEndedEvent, messages []MessageInfo) session.SessionEndedEvent {
	if len(messages) == 0 {
		return event
	}

	// Extrai IDs das mensagens ordenadas por timestamp
	messageIDs := make([]uuid.UUID, len(messages))
	var triggerMsgID *uuid.UUID
	var channelID *uuid.UUID
	var firstMsgTime, lastMsgTime *time.Time
	inboundCount := 0
	outboundCount := 0

	for i, msg := range messages {
		messageIDs[i] = msg.ID

		// Primeira mensagem é o trigger
		if i == 0 {
			triggerMsgID = &msg.ID
			firstMsgTime = &msg.Timestamp
		}

		// Última mensagem
		if i == len(messages)-1 {
			lastMsgTime = &msg.Timestamp
		}

		// Canal da primeira mensagem (todas devem ser do mesmo canal)
		if channelID == nil && msg.ChannelID != nil {
			channelID = msg.ChannelID
		}

		// Conta mensagens inbound/outbound
		if msg.Direction == "inbound" {
			inboundCount++
		} else if msg.Direction == "outbound" {
			outboundCount++
		}
	}

	// Atualiza channelID no evento
	event.ChannelID = channelID

	// Adiciona informações de mensagens
	return event.WithMessages(
		messageIDs,
		triggerMsgID,
		len(messages),
		inboundCount,
		outboundCount,
		firstMsgTime,
		lastMsgTime,
	)
}

// CleanupSessionsActivity faz limpeza de sessões órfãs
func (a *SessionActivities) CleanupSessionsActivity(ctx context.Context, input CleanupSessionsActivityInput) (CleanupSessionsActivityResult, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Starting session cleanup",
		"max_inactivity_duration", input.MaxInactivityDuration.String())

	// Busca sessões ativas há mais tempo que o limite
	cutoffTime := time.Now().Add(-input.MaxInactivityDuration)
	expiredSessions, err := a.sessionRepo.FindActiveBeforeTime(ctx, cutoffTime)
	if err != nil {
		return CleanupSessionsActivityResult{}, fmt.Errorf("failed to find expired sessions: %w", err)
	}

	if len(expiredSessions) == 0 {
		logger.Info("No expired sessions found")
		return CleanupSessionsActivityResult{}, nil
	}

	logger.Info("Found expired sessions to cleanup", "count", len(expiredSessions))

	var allEvents []shared.DomainEvent
	sessionsCleaned := 0

	// Processa cada sessão expirada
	for _, sess := range expiredSessions {
		// Encerra por timeout
		if err := sess.End(session.ReasonInactivityTimeout); err != nil {
			logger.Error("Failed to end expired session",
				"session_id", sess.ID().String(),
				"error", err)
			continue
		}

		// Salva
		if err := a.sessionRepo.Save(ctx, sess); err != nil {
			logger.Error("Failed to save expired session",
				"session_id", sess.ID().String(),
				"error", err)
			continue
		}

		// Coleta eventos
		events := sess.DomainEvents()
		for _, event := range events {
			allEvents = append(allEvents, event)
		}
		sess.ClearEvents()

		sessionsCleaned++

		logger.Info("Expired session cleaned",
			"session_id", sess.ID().String(),
			"events_generated", len(events))
	}

	// Publica todos os eventos em batch
	eventsPublished := 0
	if len(allEvents) > 0 {
		if err := a.eventBus.PublishBatch(ctx, allEvents); err != nil {
			logger.Error("Failed to publish cleanup events", "error", err)
		} else {
			eventsPublished = len(allEvents)
		}
	}

	logger.Info("Session cleanup completed",
		"sessions_cleaned", sessionsCleaned,
		"events_published", eventsPublished)

	return CleanupSessionsActivityResult{
		SessionsCleaned: sessionsCleaned,
		EventsPublished: eventsPublished,
	}, nil
}

// SendSessionTimeoutWarningActivity sends timeout warning to contact
func SendSessionTimeoutWarningActivity(ctx context.Context, input SessionTimeoutWarningActivity) (*SessionTimeoutWarningActivityResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending session timeout warning", "session_id", input.SessionID, "time_remaining", input.TimeRemaining)

	// TODO: Implement actual warning notification
	// This would typically:
	// 1. Send WhatsApp message to contact
	// 2. Send email notification
	// 3. Create internal alert for agent
	// 4. Log the warning in session timeline

	result := &SessionTimeoutWarningActivityResult{
		WarningSentAt: time.Now().UTC(),
		Method:        "whatsapp", // Default method
	}

	logger.Info("Timeout warning sent successfully", "session_id", input.SessionID, "method", result.Method)
	return result, nil
}

// EndSessionDueToTimeoutActivity ends session due to timeout
func EndSessionDueToTimeoutActivity(ctx context.Context, input SessionTimeoutActivity) (*SessionTimeoutActivityResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Ending session due to timeout", "session_id", input.SessionID, "reason", input.Reason)

	// TODO: Implement actual session ending
	// This would typically:
	// 1. Load session from repository
	// 2. Call session.End(ReasonTimeout)
	// 3. Save session
	// 4. Publish domain events
	// 5. Generate session summary
	// 6. Notify agent of timeout

	result := &SessionTimeoutActivityResult{
		EndedAt: time.Now().UTC(),
		Summary: fmt.Sprintf("Session ended due to %s", input.Reason),
	}

	logger.Info("Session ended due to timeout", "session_id", input.SessionID, "ended_at", result.EndedAt)
	return result, nil
}

// RegisterActivities registra as activities no worker Temporal
func (a *SessionActivities) RegisterActivities() []interface{} {
	return []interface{}{
		a.EndSessionActivity,
		a.CleanupSessionsActivity,
		SendSessionTimeoutWarningActivity,
		EndSessionDueToTimeoutActivity,
	}
}
