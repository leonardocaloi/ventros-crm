package session

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// DomainEvent é um alias para shared.DomainEvent (compatibilidade retroativa).
type DomainEvent = shared.DomainEvent

// SessionStartedEvent - Sessão iniciada.
type SessionStartedEvent struct {
	shared.BaseEvent
	SessionID     uuid.UUID
	ContactID     uuid.UUID
	TenantID      string
	ChannelTypeID *int
	StartedAt     time.Time
}

// NewSessionStartedEvent cria um novo evento de sessão iniciada.
func NewSessionStartedEvent(sessionID, contactID uuid.UUID, tenantID string, channelTypeID *int) SessionStartedEvent {
	return SessionStartedEvent{
		BaseEvent:     shared.NewBaseEvent("session.started", time.Now()),
		SessionID:     sessionID,
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: channelTypeID,
		StartedAt:     time.Now(),
	}
}

// SessionEndedEvent - Sessão encerrada com contexto completo.
// Inclui canal, contato, mensagens e eventos da sessão para facilitar processamento downstream.
type SessionEndedEvent struct {
	shared.BaseEvent
	SessionID     uuid.UUID
	ContactID     uuid.UUID
	TenantID      string
	ChannelID     *uuid.UUID
	ChannelTypeID *int
	PipelineID    *uuid.UUID
	EndedAt       time.Time
	StartedAt     time.Time
	Reason        EndReason
	Duration      int // segundos

	// Contexto completo da sessão (adicionado para webhook enrichment)
	MessageIDs    []uuid.UUID            // Lista de IDs das mensagens (ordenado por timestamp)
	TriggerMsgID  *uuid.UUID             // ID da primeira mensagem que iniciou a sessão
	EventsSummary map[string]int         // Resumo de eventos: {"message.created": 5, "tracking.captured": 1}
	Metrics       SessionEndedMetrics    // Métricas da sessão
}

// SessionEndedMetrics contém métricas da sessão encerrada
type SessionEndedMetrics struct {
	TotalMessages     int // Total de mensagens
	InboundMessages   int // Mensagens recebidas do contato
	OutboundMessages  int // Mensagens enviadas pelo sistema/agente
	FirstMessageAt    *time.Time // Timestamp da primeira mensagem
	LastMessageAt     *time.Time // Timestamp da última mensagem
}

func NewSessionEndedEvent(
	sessionID uuid.UUID,
	contactID uuid.UUID,
	tenantID string,
	channelID *uuid.UUID,
	channelTypeID *int,
	pipelineID *uuid.UUID,
	startedAt time.Time,
	reason EndReason,
	duration int,
) SessionEndedEvent {
	return SessionEndedEvent{
		BaseEvent:     shared.NewBaseEvent("session.ended", time.Now()),
		SessionID:     sessionID,
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelID:     channelID,
		ChannelTypeID: channelTypeID,
		PipelineID:    pipelineID,
		EndedAt:       time.Now(),
		StartedAt:     startedAt,
		Reason:        reason,
		Duration:      duration,
		MessageIDs:    []uuid.UUID{},
		EventsSummary: make(map[string]int),
		Metrics:       SessionEndedMetrics{},
	}
}

// WithMessages adiciona informações de mensagens ao evento
func (e SessionEndedEvent) WithMessages(messageIDs []uuid.UUID, triggerMsgID *uuid.UUID, totalMsgs, inbound, outbound int, firstMsgAt, lastMsgAt *time.Time) SessionEndedEvent {
	e.MessageIDs = messageIDs
	e.TriggerMsgID = triggerMsgID
	e.Metrics = SessionEndedMetrics{
		TotalMessages:    totalMsgs,
		InboundMessages:  inbound,
		OutboundMessages: outbound,
		FirstMessageAt:   firstMsgAt,
		LastMessageAt:    lastMsgAt,
	}
	return e
}

// WithEventsSummary adiciona resumo de eventos ao evento
func (e SessionEndedEvent) WithEventsSummary(summary map[string]int) SessionEndedEvent {
	e.EventsSummary = summary
	return e
}

// MessageRecordedEvent - Mensagem registrada na sessão.
type MessageRecordedEvent struct {
	shared.BaseEvent
	SessionID   uuid.UUID
	FromContact bool
	RecordedAt  time.Time
}

func NewMessageRecordedEvent(sessionID uuid.UUID, fromContact bool) MessageRecordedEvent {
	return MessageRecordedEvent{
		BaseEvent:   shared.NewBaseEvent("session.message_recorded", time.Now()),
		SessionID:   sessionID,
		FromContact: fromContact,
		RecordedAt:  time.Now(),
	}
}

// AgentAssignedEvent - Agente atribuído à sessão.
type AgentAssignedEvent struct {
	shared.BaseEvent
	SessionID  uuid.UUID
	AgentID    uuid.UUID
	AssignedAt time.Time
}

func NewAgentAssignedEvent(sessionID, agentID uuid.UUID) AgentAssignedEvent {
	return AgentAssignedEvent{
		BaseEvent:  shared.NewBaseEvent("session.agent_assigned", time.Now()),
		SessionID:  sessionID,
		AgentID:    agentID,
		AssignedAt: time.Now(),
	}
}

// SessionResolvedEvent - Sessão marcada como resolvida.
type SessionResolvedEvent struct {
	shared.BaseEvent
	SessionID  uuid.UUID
	ResolvedAt time.Time
}

func NewSessionResolvedEvent(sessionID uuid.UUID) SessionResolvedEvent {
	return SessionResolvedEvent{
		BaseEvent:  shared.NewBaseEvent("session.resolved", time.Now()),
		SessionID:  sessionID,
		ResolvedAt: time.Now(),
	}
}

// SessionEscalatedEvent - Sessão escalada.
type SessionEscalatedEvent struct {
	shared.BaseEvent
	SessionID   uuid.UUID
	EscalatedAt time.Time
}

func NewSessionEscalatedEvent(sessionID uuid.UUID) SessionEscalatedEvent {
	return SessionEscalatedEvent{
		BaseEvent:   shared.NewBaseEvent("session.escalated", time.Now()),
		SessionID:   sessionID,
		EscalatedAt: time.Now(),
	}
}

// SessionSummarizedEvent - Resumo gerado por IA.
type SessionSummarizedEvent struct {
	shared.BaseEvent
	SessionID      uuid.UUID
	Summary        string
	Sentiment      Sentiment
	SentimentScore float64
	GeneratedAt    time.Time
}

func NewSessionSummarizedEvent(sessionID uuid.UUID, summary string, sentiment Sentiment, score float64) SessionSummarizedEvent {
	return SessionSummarizedEvent{
		BaseEvent:      shared.NewBaseEvent("session.summarized", time.Now()),
		SessionID:      sessionID,
		Summary:        summary,
		Sentiment:      sentiment,
		SentimentScore: score,
		GeneratedAt:    time.Now(),
	}
}

// SessionAbandonedEvent - Sessão abandonada (cliente parou de responder).
type SessionAbandonedEvent struct {
	shared.BaseEvent
	SessionID                 uuid.UUID
	LastAgentMessageAt        time.Time
	MinutesSinceLastResponse  int
	MessagesBeforeAbandonment int
	ConversationStage         string
	AbandonedAt               time.Time
}

func NewSessionAbandonedEvent(sessionID uuid.UUID, lastAgentMsgAt time.Time, minutesSinceResp, msgsBefore int, stage string) SessionAbandonedEvent {
	return SessionAbandonedEvent{
		BaseEvent:                 shared.NewBaseEvent("session.abandoned", time.Now()),
		SessionID:                 sessionID,
		LastAgentMessageAt:        lastAgentMsgAt,
		MinutesSinceLastResponse:  minutesSinceResp,
		MessagesBeforeAbandonment: msgsBefore,
		ConversationStage:         stage,
		AbandonedAt:               time.Now(),
	}
}
