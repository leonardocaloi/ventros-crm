package session

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
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
	eventBus    EventBus
}

// EventBus interface para publicar eventos
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
	PublishBatch(ctx context.Context, events []shared.DomainEvent) error
}

// NewSessionActivities cria uma nova instância das activities
func NewSessionActivities(sessionRepo session.Repository, eventBus EventBus) *SessionActivities {
	return &SessionActivities{
		sessionRepo: sessionRepo,
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
	
	// Publica eventos de domínio
	events := sess.DomainEvents()
	eventsCount := len(events)
	
	if eventsCount > 0 {
		// Converte para shared.DomainEvent
		sharedEvents := make([]shared.DomainEvent, len(events))
		for i, event := range events {
			sharedEvents[i] = event
		}
		
		if err := a.eventBus.PublishBatch(ctx, sharedEvents); err != nil {
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

// RegisterActivities registra as activities no worker Temporal
func (a *SessionActivities) RegisterActivities() []interface{} {
	return []interface{}{
		a.EndSessionActivity,
		a.CleanupSessionsActivity,
	}
}
