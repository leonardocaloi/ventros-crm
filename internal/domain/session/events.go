package session

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para eventos de domínio.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// SessionStartedEvent - Sessão iniciada.
type SessionStartedEvent struct {
	SessionID     uuid.UUID
	ContactID     uuid.UUID
	TenantID      string
	ChannelTypeID *int
	StartedAt     time.Time
}

func (e SessionStartedEvent) EventName() string      { return "session.started" }
func (e SessionStartedEvent) OccurredAt() time.Time  { return e.StartedAt }

// SessionEndedEvent - Sessão encerrada.
type SessionEndedEvent struct {
	SessionID uuid.UUID
	EndedAt   time.Time
	Reason    EndReason
	Duration  int // segundos
}

func (e SessionEndedEvent) EventName() string     { return "session.ended" }
func (e SessionEndedEvent) OccurredAt() time.Time { return e.EndedAt }

// MessageRecordedEvent - Mensagem registrada na sessão.
type MessageRecordedEvent struct {
	SessionID   uuid.UUID
	FromContact bool
	RecordedAt  time.Time
}

func (e MessageRecordedEvent) EventName() string     { return "session.message_recorded" }
func (e MessageRecordedEvent) OccurredAt() time.Time { return e.RecordedAt }

// AgentAssignedEvent - Agente atribuído à sessão.
type AgentAssignedEvent struct {
	SessionID  uuid.UUID
	AgentID    uuid.UUID
	AssignedAt time.Time
}

func (e AgentAssignedEvent) EventName() string     { return "session.agent_assigned" }
func (e AgentAssignedEvent) OccurredAt() time.Time { return e.AssignedAt }

// SessionResolvedEvent - Sessão marcada como resolvida.
type SessionResolvedEvent struct {
	SessionID  uuid.UUID
	ResolvedAt time.Time
}

func (e SessionResolvedEvent) EventName() string     { return "session.resolved" }
func (e SessionResolvedEvent) OccurredAt() time.Time { return e.ResolvedAt }

// SessionEscalatedEvent - Sessão escalada.
type SessionEscalatedEvent struct {
	SessionID   uuid.UUID
	EscalatedAt time.Time
}

func (e SessionEscalatedEvent) EventName() string     { return "session.escalated" }
func (e SessionEscalatedEvent) OccurredAt() time.Time { return e.EscalatedAt }

// SessionSummarizedEvent - Resumo gerado por IA.
type SessionSummarizedEvent struct {
	SessionID      uuid.UUID
	Summary        string
	Sentiment      Sentiment
	SentimentScore float64
	GeneratedAt    time.Time
}

func (e SessionSummarizedEvent) EventName() string     { return "session.summarized" }
func (e SessionSummarizedEvent) OccurredAt() time.Time { return e.GeneratedAt }

// SessionAbandonedEvent - Sessão abandonada (cliente parou de responder).
type SessionAbandonedEvent struct {
	SessionID                  uuid.UUID
	LastAgentMessageAt         time.Time
	MinutesSinceLastResponse   int
	MessagesBeforeAbandonment  int
	ConversationStage          string
	AbandonedAt                time.Time
}

func (e SessionAbandonedEvent) EventName() string     { return "session.abandoned" }
func (e SessionAbandonedEvent) OccurredAt() time.Time { return e.AbandonedAt }
