package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type DomainEvent = shared.DomainEvent

type SessionStartedEvent struct {
	shared.BaseEvent
	SessionID     uuid.UUID
	ContactID     uuid.UUID
	TenantID      string
	ChannelTypeID *int
	StartedAt     time.Time
}

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
	Duration      int
	MessageIDs    []uuid.UUID
	TriggerMsgID  *uuid.UUID
	EventsSummary map[string]int
	Metrics       SessionEndedMetrics
}

type SessionEndedMetrics struct {
	TotalMessages    int
	InboundMessages  int
	OutboundMessages int
	FirstMessageAt   *time.Time
	LastMessageAt    *time.Time
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

func (e SessionEndedEvent) WithEventsSummary(summary map[string]int) SessionEndedEvent {
	e.EventsSummary = summary
	return e
}

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

// AssignmentSource indica a origem/motivo do assignment
type AssignmentSource string

const (
	// Assignment manual por um agente
	AssignmentSourceManual AssignmentSource = "manual"

	// Assignment automático na criação da sessão
	AssignmentSourceAutomatic AssignmentSource = "automatic"

	// Reassignment automático por regra de inatividade
	AssignmentSourceReassignmentInactivity AssignmentSource = "reassignment_inactivity"

	// Reassignment automático por falta de resposta
	AssignmentSourceReassignmentNoResponse AssignmentSource = "reassignment_no_response"

	// Reassignment automático por balanceamento de carga
	AssignmentSourceReassignmentWorkload AssignmentSource = "reassignment_workload"

	// Reassignment manual por outro agente
	AssignmentSourceReassignmentManual AssignmentSource = "reassignment_manual"

	// Assignment inicial do sistema (quando sessão é criada sem agent)
	AssignmentSourceSystem AssignmentSource = "system"
)

// AgentAssignedEvent evento rico com informações sobre origem do assignment
type AgentAssignedEvent struct {
	shared.BaseEvent
	SessionID  uuid.UUID
	AgentID    uuid.UUID
	AssignedAt time.Time

	// Origem do assignment (manual, automatic, reassignment, etc)
	Source AssignmentSource

	// ID do agente que fez o assignment (se Source = manual ou reassignment_manual)
	AssignedByAgentID *uuid.UUID

	// ID do agente anterior (se é um reassignment)
	PreviousAgentID *uuid.UUID

	// Motivo do reassignment (nome da regra, ex: "inactivity_5min")
	ReassignmentReason *string

	// Estratégia usada (round_robin, least_sessions, etc)
	AssignmentStrategy *string

	// Número de reassignments anteriores desta sessão
	ReassignmentCount int
}

// NewAgentAssignedEvent cria evento de assignment básico (para compatibilidade)
func NewAgentAssignedEvent(sessionID, agentID uuid.UUID) AgentAssignedEvent {
	return AgentAssignedEvent{
		BaseEvent:  shared.NewBaseEvent("session.agent_assigned", time.Now()),
		SessionID:  sessionID,
		AgentID:    agentID,
		AssignedAt: time.Now(),
		Source:     AssignmentSourceSystem,
	}
}

// NewAgentAssignedEventWithSource cria evento com informações completas sobre origem
func NewAgentAssignedEventWithSource(
	sessionID uuid.UUID,
	agentID uuid.UUID,
	source AssignmentSource,
	assignedByAgentID *uuid.UUID,
	previousAgentID *uuid.UUID,
	reassignmentReason *string,
	strategy *string,
	reassignmentCount int,
) AgentAssignedEvent {
	return AgentAssignedEvent{
		BaseEvent:          shared.NewBaseEvent("session.agent_assigned", time.Now()),
		SessionID:          sessionID,
		AgentID:            agentID,
		AssignedAt:         time.Now(),
		Source:             source,
		AssignedByAgentID:  assignedByAgentID,
		PreviousAgentID:    previousAgentID,
		ReassignmentReason: reassignmentReason,
		AssignmentStrategy: strategy,
		ReassignmentCount:  reassignmentCount,
	}
}

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
