package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Session é o Aggregate Root para conversas.
// Agrupa messages e garante invariantes de negócio.
type Session struct {
	id               uuid.UUID
	contactID        uuid.UUID
	tenantID         string
	channelTypeID    *int
	startedAt        time.Time
	endedAt          *time.Time
	status           Status
	endReason        *EndReason
	timeoutDuration  time.Duration
	lastActivityAt   time.Time
	
	// Métricas
	messageCount         int
	messagesFromContact  int
	messagesFromAgent    int
	durationSeconds      int
	
	// Agentes
	agentIDs     []uuid.UUID
	agentTransfers int
	
	// AI/Analytics
	summary        *string
	sentiment      *Sentiment
	sentimentScore *float64
	topics         []string
	nextSteps      []string
	keyEntities    map[string]interface{}
	
	// Flags de negócio
	resolved   bool
	escalated  bool
	converted  bool
	outcomeTags []string
	
	// Domain Events
	events []DomainEvent
}

// NewSession cria uma nova sessão (factory method).
func NewSession(
	contactID uuid.UUID,
	tenantID string,
	channelTypeID *int,
	timeoutDuration time.Duration,
) (*Session, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if timeoutDuration <= 0 {
		timeoutDuration = 30 * time.Minute // default
	}

	now := time.Now()
	session := &Session{
		id:              uuid.New(),
		contactID:       contactID,
		tenantID:        tenantID,
		channelTypeID:   channelTypeID,
		startedAt:       now,
		status:          StatusActive,
		timeoutDuration: timeoutDuration,
		lastActivityAt:  now,
		agentIDs:        []uuid.UUID{},
		topics:          []string{},
		nextSteps:       []string{},
		outcomeTags:     []string{},
		keyEntities:     make(map[string]interface{}),
		events:          []DomainEvent{},
	}

	session.addEvent(SessionStartedEvent{
		SessionID:     session.id,
		ContactID:     contactID,
		TenantID:      tenantID,
		ChannelTypeID: channelTypeID,
		StartedAt:     now,
	})

	return session, nil
}

// ReconstructSession reconstrói uma Session a partir de dados persistidos.
func ReconstructSession(
	id uuid.UUID,
	contactID uuid.UUID,
	tenantID string,
	channelTypeID *int,
	startedAt time.Time,
	endedAt *time.Time,
	status Status,
	endReason *EndReason,
	timeoutDuration time.Duration,
	lastActivityAt time.Time,
	messageCount int,
	messagesFromContact int,
	messagesFromAgent int,
	durationSeconds int,
	agentIDs []uuid.UUID,
	agentTransfers int,
	summary *string,
	sentiment *Sentiment,
	sentimentScore *float64,
	topics []string,
	nextSteps []string,
	keyEntities map[string]interface{},
	resolved bool,
	escalated bool,
	converted bool,
	outcomeTags []string,
) *Session {
	if agentIDs == nil {
		agentIDs = []uuid.UUID{}
	}
	if topics == nil {
		topics = []string{}
	}
	if nextSteps == nil {
		nextSteps = []string{}
	}
	if outcomeTags == nil {
		outcomeTags = []string{}
	}
	if keyEntities == nil {
		keyEntities = make(map[string]interface{})
	}

	return &Session{
		id:                  id,
		contactID:           contactID,
		tenantID:            tenantID,
		channelTypeID:       channelTypeID,
		startedAt:           startedAt,
		endedAt:             endedAt,
		status:              status,
		endReason:           endReason,
		timeoutDuration:     timeoutDuration,
		lastActivityAt:      lastActivityAt,
		messageCount:        messageCount,
		messagesFromContact: messagesFromContact,
		messagesFromAgent:   messagesFromAgent,
		durationSeconds:     durationSeconds,
		agentIDs:            agentIDs,
		agentTransfers:      agentTransfers,
		summary:             summary,
		sentiment:           sentiment,
		sentimentScore:      sentimentScore,
		topics:              topics,
		nextSteps:           nextSteps,
		keyEntities:         keyEntities,
		resolved:            resolved,
		escalated:           escalated,
		converted:           converted,
		outcomeTags:         outcomeTags,
		events:              []DomainEvent{},
	}
}

// RecordMessage registra uma mensagem na sessão.
func (s *Session) RecordMessage(fromContact bool) error {
	if s.status != StatusActive {
		return errors.New("cannot add message to non-active session")
	}

	s.messageCount++
	if fromContact {
		s.messagesFromContact++
	} else {
		s.messagesFromAgent++
	}
	
	s.lastActivityAt = time.Now()
	
	s.addEvent(MessageRecordedEvent{
		SessionID:   s.id,
		FromContact: fromContact,
		RecordedAt:  s.lastActivityAt,
	})

	return nil
}

// AssignAgent atribui um agente à sessão.
func (s *Session) AssignAgent(agentID uuid.UUID) error {
	if s.status != StatusActive {
		return errors.New("cannot assign agent to non-active session")
	}

	// Verifica se já existe
	for _, id := range s.agentIDs {
		if id == agentID {
			return nil // Já atribuído
		}
	}

	// Se já tinha outros agentes, é uma transferência
	if len(s.agentIDs) > 0 {
		s.agentTransfers++
	}

	s.agentIDs = append(s.agentIDs, agentID)
	
	s.addEvent(AgentAssignedEvent{
		SessionID: s.id,
		AgentID:   agentID,
		AssignedAt: time.Now(),
	})

	return nil
}

// CheckTimeout verifica se a sessão deve ser encerrada por inatividade.
func (s *Session) CheckTimeout() bool {
	if s.status != StatusActive {
		return false
	}

	if time.Since(s.lastActivityAt) > s.timeoutDuration {
		s.End(ReasonInactivityTimeout)
		return true
	}

	return false
}

// End encerra a sessão manualmente ou por timeout.
func (s *Session) End(reason EndReason) error {
	if s.status != StatusActive {
		return errors.New("session is not active")
	}

	now := time.Now()
	s.status = StatusEnded
	s.endedAt = &now
	s.endReason = &reason
	s.durationSeconds = int(now.Sub(s.startedAt).Seconds())

	s.addEvent(SessionEndedEvent{
		SessionID: s.id,
		EndedAt:   now,
		Reason:    reason,
		Duration:  s.durationSeconds,
	})

	return nil
}

// Resolve marca a sessão como resolvida.
func (s *Session) Resolve() error {
	if s.status == StatusActive {
		return errors.New("cannot resolve active session")
	}
	
	s.resolved = true
	
	s.addEvent(SessionResolvedEvent{
		SessionID:  s.id,
		ResolvedAt: time.Now(),
	})
	
	return nil
}

// Escalate marca a sessão como escalada.
func (s *Session) Escalate() error {
	s.escalated = true
	
	s.addEvent(SessionEscalatedEvent{
		SessionID:   s.id,
		EscalatedAt: time.Now(),
	})
	
	return nil
}

// SetSummary define o resumo gerado por IA.
func (s *Session) SetSummary(summary string, sentiment Sentiment, score float64, topics, nextSteps []string) {
	s.summary = &summary
	s.sentiment = &sentiment
	s.sentimentScore = &score
	s.topics = topics
	s.nextSteps = nextSteps

	s.addEvent(SessionSummarizedEvent{
		SessionID:      s.id,
		Summary:        summary,
		Sentiment:      sentiment,
		SentimentScore: score,
		GeneratedAt:    time.Now(),
	})
}

// IsActive verifica se a sessão está ativa.
func (s *Session) IsActive() bool {
	return s.status == StatusActive
}

// ShouldGenerateSummary verifica se deve gerar resumo.
func (s *Session) ShouldGenerateSummary() bool {
	return s.status == StatusEnded && s.messageCount >= 3 && s.summary == nil
}

// Getters (acesso controlado ao estado)

func (s *Session) ID() uuid.UUID                { return s.id }
func (s *Session) ContactID() uuid.UUID         { return s.contactID }
func (s *Session) TenantID() string             { return s.tenantID }
func (s *Session) ChannelTypeID() *int          { return s.channelTypeID }
func (s *Session) StartedAt() time.Time         { return s.startedAt }
func (s *Session) EndedAt() *time.Time          { return s.endedAt }
func (s *Session) Status() Status               { return s.status }
func (s *Session) EndReason() *EndReason        { return s.endReason }
func (s *Session) TimeoutDuration() time.Duration { return s.timeoutDuration }
func (s *Session) LastActivityAt() time.Time    { return s.lastActivityAt }
func (s *Session) MessageCount() int            { return s.messageCount }
func (s *Session) MessagesFromContact() int     { return s.messagesFromContact }
func (s *Session) MessagesFromAgent() int       { return s.messagesFromAgent }
func (s *Session) DurationSeconds() int         { return s.durationSeconds }
func (s *Session) AgentIDs() []uuid.UUID        { return append([]uuid.UUID{}, s.agentIDs...) } // Copy
func (s *Session) AgentTransfers() int          { return s.agentTransfers }
func (s *Session) Summary() *string             { return s.summary }
func (s *Session) Sentiment() *Sentiment        { return s.sentiment }
func (s *Session) SentimentScore() *float64     { return s.sentimentScore }
func (s *Session) Topics() []string             { return append([]string{}, s.topics...) } // Copy
func (s *Session) NextSteps() []string          { return append([]string{}, s.nextSteps...) } // Copy
func (s *Session) KeyEntities() map[string]interface{} {
	// Return copy
	copy := make(map[string]interface{})
	for k, v := range s.keyEntities {
		copy[k] = v
	}
	return copy
}
func (s *Session) IsResolved() bool   { return s.resolved }
func (s *Session) IsEscalated() bool  { return s.escalated }
func (s *Session) IsConverted() bool  { return s.converted }
func (s *Session) OutcomeTags() []string { return append([]string{}, s.outcomeTags...) } // Copy

// DomainEvents retorna os eventos de domínio.
func (s *Session) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, s.events...)
}

// ClearEvents limpa os eventos (após publicação).
func (s *Session) ClearEvents() {
	s.events = []DomainEvent{}
}

// addEvent adiciona um evento de domínio.
func (s *Session) addEvent(event DomainEvent) {
	s.events = append(s.events, event)
}
