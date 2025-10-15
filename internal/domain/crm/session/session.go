package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type Session struct {
	id              uuid.UUID
	version         int // Optimistic locking - prevents lost updates
	contactID       uuid.UUID
	tenantID        string
	channelTypeID   *int
	pipelineID      *uuid.UUID
	startedAt       time.Time
	endedAt         *time.Time
	status          Status
	endReason       *EndReason
	timeoutDuration time.Duration
	lastActivityAt  time.Time

	messageCount        int
	messagesFromContact int
	messagesFromAgent   int
	durationSeconds     int

	firstContactMessageAt    *time.Time
	firstAgentResponseAt     *time.Time
	agentResponseTimeSeconds *int
	contactWaitTimeSeconds   *int

	agentIDs       []uuid.UUID
	agentTransfers int

	summary        *string
	sentiment      *Sentiment
	sentimentScore *float64
	topics         []string
	nextSteps      []string
	keyEntities    map[string]interface{}

	resolved    bool
	escalated   bool
	converted   bool
	outcomeTags []string

	events []DomainEvent
}

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
		timeoutDuration = 30 * time.Minute
	}

	now := time.Now()
	session := &Session{
		id:              uuid.New(),
		version:         1, // Start with version 1 for new aggregates
		contactID:       contactID,
		tenantID:        tenantID,
		channelTypeID:   channelTypeID,
		pipelineID:      nil,
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

	session.addEvent(NewSessionStartedEvent(session.id, contactID, tenantID, channelTypeID))

	return session, nil
}

// NewSessionWithTimestamp creates a new session with custom start time (for history import)
// This is critical for history import to maintain correct chronological order
func NewSessionWithTimestamp(
	contactID uuid.UUID,
	tenantID string,
	channelTypeID *int,
	timeoutDuration time.Duration,
	startTime time.Time,
) (*Session, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be nil")
	}
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if timeoutDuration <= 0 {
		timeoutDuration = 30 * time.Minute
	}

	session := &Session{
		id:              uuid.New(),
		version:         1,
		contactID:       contactID,
		tenantID:        tenantID,
		channelTypeID:   channelTypeID,
		pipelineID:      nil,
		startedAt:       startTime, // Use provided timestamp instead of time.Now()
		status:          StatusActive,
		timeoutDuration: timeoutDuration,
		lastActivityAt:  startTime, // Initialize with start time
		agentIDs:        []uuid.UUID{},
		topics:          []string{},
		nextSteps:       []string{},
		outcomeTags:     []string{},
		keyEntities:     make(map[string]interface{}),
		events:          []DomainEvent{},
	}

	session.addEvent(NewSessionStartedEvent(session.id, contactID, tenantID, channelTypeID))

	return session, nil
}

func NewSessionWithPipeline(
	contactID uuid.UUID,
	tenantID string,
	channelTypeID *int,
	pipelineID uuid.UUID,
	timeoutDuration time.Duration,
) (*Session, error) {
	if pipelineID == uuid.Nil {
		return nil, errors.New("pipelineID cannot be nil - Session requires active Pipeline")
	}

	session, err := NewSession(contactID, tenantID, channelTypeID, timeoutDuration)
	if err != nil {
		return nil, err
	}

	session.pipelineID = &pipelineID
	return session, nil
}

func ReconstructSession(
	id uuid.UUID,
	version int, // Optimistic locking version
	contactID uuid.UUID,
	tenantID string,
	channelTypeID *int,
	pipelineID *uuid.UUID,
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
	firstContactMessageAt *time.Time,
	firstAgentResponseAt *time.Time,
	agentResponseTimeSeconds *int,
	contactWaitTimeSeconds *int,
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
	if version == 0 {
		version = 1 // Default to version 1 if not set (backwards compatibility)
	}

	return &Session{
		id:                       id,
		version:                  version,
		contactID:                contactID,
		tenantID:                 tenantID,
		channelTypeID:            channelTypeID,
		pipelineID:               pipelineID,
		startedAt:                startedAt,
		endedAt:                  endedAt,
		status:                   status,
		endReason:                endReason,
		timeoutDuration:          timeoutDuration,
		lastActivityAt:           lastActivityAt,
		messageCount:             messageCount,
		messagesFromContact:      messagesFromContact,
		messagesFromAgent:        messagesFromAgent,
		durationSeconds:          durationSeconds,
		firstContactMessageAt:    firstContactMessageAt,
		firstAgentResponseAt:     firstAgentResponseAt,
		agentResponseTimeSeconds: agentResponseTimeSeconds,
		contactWaitTimeSeconds:   contactWaitTimeSeconds,
		agentIDs:                 agentIDs,
		agentTransfers:           agentTransfers,
		summary:                  summary,
		sentiment:                sentiment,
		sentimentScore:           sentimentScore,
		topics:                   topics,
		nextSteps:                nextSteps,
		keyEntities:              keyEntities,
		resolved:                 resolved,
		escalated:                escalated,
		converted:                converted,
		outcomeTags:              outcomeTags,
		events:                   []DomainEvent{},
	}
}

func (s *Session) RecordMessage(fromContact bool, messageTimestamp time.Time) error {
	if s.status != StatusActive {
		return errors.New("cannot add message to non-active session")
	}

	s.messageCount++
	now := time.Now()

	if fromContact {
		s.messagesFromContact++

		if s.firstContactMessageAt == nil {
			s.firstContactMessageAt = &messageTimestamp

			if s.firstAgentResponseAt != nil {
				waitTime := int(messageTimestamp.Sub(*s.firstAgentResponseAt).Seconds())
				s.contactWaitTimeSeconds = &waitTime
			}
		}

		if s.firstAgentResponseAt == nil && s.firstContactMessageAt != nil {
			responseTime := int(now.Sub(*s.firstContactMessageAt).Seconds())

			s.agentResponseTimeSeconds = &responseTime
		}
	} else {
		s.messagesFromAgent++

		if s.firstAgentResponseAt == nil {
			s.firstAgentResponseAt = &messageTimestamp

			if s.firstContactMessageAt != nil {
				responseTime := int(messageTimestamp.Sub(*s.firstContactMessageAt).Seconds())
				s.agentResponseTimeSeconds = &responseTime
			}
		}

		if s.firstContactMessageAt == nil && s.firstAgentResponseAt != nil {
			waitTime := int(now.Sub(*s.firstAgentResponseAt).Seconds())

			s.contactWaitTimeSeconds = &waitTime
		}
	}

	// Use message timestamp for lastActivityAt (critical for history import!)
	// This ensures sessions are ordered correctly by actual message time, not import time
	s.lastActivityAt = messageTimestamp

	s.addEvent(NewMessageRecordedEvent(s.id, fromContact))

	return nil
}

func (s *Session) AssignAgent(agentID uuid.UUID) error {
	if s.status != StatusActive {
		return errors.New("cannot assign agent to non-active session")
	}

	for _, id := range s.agentIDs {
		if id == agentID {
			return nil
		}
	}

	if len(s.agentIDs) > 0 {
		s.agentTransfers++
	}

	s.agentIDs = append(s.agentIDs, agentID)

	s.addEvent(NewAgentAssignedEvent(s.id, agentID))

	return nil
}

// AssignAgentWithSource atribui um agente com informações completas sobre a origem do assignment
// Use este método quando tiver contexto sobre quem/como o assignment foi feito
func (s *Session) AssignAgentWithSource(
	agentID uuid.UUID,
	source AssignmentSource,
	assignedByAgentID *uuid.UUID,
	strategy *string,
	reassignmentReason *string,
) error {
	if s.status != StatusActive {
		return errors.New("cannot assign agent to non-active session")
	}

	// Verifica se já está atribuído a este agente
	for _, id := range s.agentIDs {
		if id == agentID {
			return nil
		}
	}

	// Captura agente anterior se existir
	var previousAgentID *uuid.UUID
	if len(s.agentIDs) > 0 {
		prev := s.agentIDs[len(s.agentIDs)-1]
		previousAgentID = &prev
		s.agentTransfers++
	}

	s.agentIDs = append(s.agentIDs, agentID)

	// Emite evento rico com informações completas
	s.addEvent(NewAgentAssignedEventWithSource(
		s.id,
		agentID,
		source,
		assignedByAgentID,
		previousAgentID,
		reassignmentReason,
		strategy,
		s.agentTransfers,
	))

	return nil
}

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

// ShouldConsolidateWith determines if this session should be consolidated with another session.
// Business Rule: Sessions from the same contact should be consolidated if the time gap
// between the last activity of the earlier session and the first activity of the later
// session is less than or equal to the timeout duration.
//
// This is critical for history imports where messages arrive out of chronological order
// and sessions are created in parallel, resulting in fragmented sessions that need to be
// consolidated based on actual message timestamps.
//
// Example:
//   Session A: contact=123, lastActivityAt=10:00
//   Session B: contact=123, startedAt=10:15, timeout=30min
//   Result: Should consolidate (gap=15min < 30min timeout)
func (s *Session) ShouldConsolidateWith(other *Session, timeout time.Duration) bool {
	// Must be from same contact
	if s.contactID != other.contactID {
		return false
	}

	// Must be from same tenant (multi-tenancy)
	if s.tenantID != other.tenantID {
		return false
	}

	// Determine which session is earlier and which is later
	var earlier, later *Session
	if s.lastActivityAt.Before(other.lastActivityAt) {
		earlier = s
		later = other
	} else {
		earlier = other
		later = s
	}

	// Calculate gap between sessions
	// Gap = later session's first activity - earlier session's last activity
	gap := later.startedAt.Sub(earlier.lastActivityAt)

	// Consolidate if gap is within timeout
	return gap <= timeout
}

func (s *Session) End(reason EndReason) error {
	if s.status != StatusActive {
		return errors.New("session is not active")
	}

	now := time.Now()
	s.status = StatusEnded
	s.endedAt = &now
	s.endReason = &reason
	s.durationSeconds = int(now.Sub(s.startedAt).Seconds())

	s.addEvent(NewSessionEndedEvent(
		s.id,
		s.contactID,
		s.tenantID,
		nil,
		s.channelTypeID,
		s.pipelineID,
		s.startedAt,
		reason,
		s.durationSeconds,
	))

	return nil
}

func (s *Session) Resolve() error {
	if s.status == StatusActive {
		return errors.New("cannot resolve active session")
	}

	s.resolved = true

	s.addEvent(NewSessionResolvedEvent(s.id))

	return nil
}

func (s *Session) Escalate() error {
	s.escalated = true

	s.addEvent(NewSessionEscalatedEvent(s.id))

	return nil
}

func (s *Session) SetSummary(summary string, sentiment Sentiment, score float64, topics, nextSteps []string) {
	s.summary = &summary
	s.sentiment = &sentiment
	s.sentimentScore = &score
	s.topics = topics
	s.nextSteps = nextSteps

	s.addEvent(NewSessionSummarizedEvent(s.id, summary, sentiment, score))
}

func (s *Session) IsActive() bool {
	return s.status == StatusActive
}

func (s *Session) ShouldGenerateSummary() bool {
	return s.status == StatusEnded && s.messageCount >= 3 && s.summary == nil
}

// HasAssignedAgents retorna true se a sessão tem pelo menos um agente atribuído
func (s *Session) HasAssignedAgents() bool {
	return len(s.agentIDs) > 0
}

// GetCurrentAgent retorna o agente atualmente atribuído à sessão (o último da lista)
// Retorna nil se nenhum agente está atribuído
func (s *Session) GetCurrentAgent() *uuid.UUID {
	if len(s.agentIDs) == 0 {
		return nil
	}
	currentAgent := s.agentIDs[len(s.agentIDs)-1]
	return &currentAgent
}

// GetReassignmentCount retorna quantas vezes a sessão foi reatribuída
func (s *Session) GetReassignmentCount() int {
	return s.agentTransfers
}

// ===== Helper Methods for Common Assignment Scenarios =====

// AssignAgentAutomatic atribui agente automaticamente com estratégia
func (s *Session) AssignAgentAutomatic(agentID uuid.UUID, strategy string) error {
	return s.AssignAgentWithSource(
		agentID,
		AssignmentSourceAutomatic,
		nil,
		&strategy,
		nil,
	)
}

// ReassignAgentManually reatribui sessão manualmente por outro agente
func (s *Session) ReassignAgentManually(newAgentID uuid.UUID, assignedByAgentID uuid.UUID, reason string) error {
	return s.AssignAgentWithSource(
		newAgentID,
		AssignmentSourceReassignmentManual,
		&assignedByAgentID,
		nil,
		&reason,
	)
}

// ReassignAgentByInactivity reatribui sessão automaticamente por inatividade
func (s *Session) ReassignAgentByInactivity(newAgentID uuid.UUID, strategy string, ruleName string) error {
	return s.AssignAgentWithSource(
		newAgentID,
		AssignmentSourceReassignmentInactivity,
		nil,
		&strategy,
		&ruleName,
	)
}

// ReassignAgentByNoResponse reatribui sessão por falta de resposta
func (s *Session) ReassignAgentByNoResponse(newAgentID uuid.UUID, strategy string, ruleName string) error {
	return s.AssignAgentWithSource(
		newAgentID,
		AssignmentSourceReassignmentNoResponse,
		nil,
		&strategy,
		&ruleName,
	)
}

// ReassignAgentByWorkload reatribui sessão por balanceamento de carga
func (s *Session) ReassignAgentByWorkload(newAgentID uuid.UUID, strategy string, reason string) error {
	return s.AssignAgentWithSource(
		newAgentID,
		AssignmentSourceReassignmentWorkload,
		nil,
		&strategy,
		&reason,
	)
}

func (s *Session) ID() uuid.UUID                  { return s.id }
func (s *Session) Version() int                   { return s.version }
func (s *Session) ContactID() uuid.UUID           { return s.contactID }
func (s *Session) TenantID() string               { return s.tenantID }
func (s *Session) ChannelTypeID() *int            { return s.channelTypeID }
func (s *Session) PipelineID() *uuid.UUID         { return s.pipelineID }
func (s *Session) StartedAt() time.Time           { return s.startedAt }
func (s *Session) EndedAt() *time.Time            { return s.endedAt }
func (s *Session) Status() Status                 { return s.status }
func (s *Session) EndReason() *EndReason          { return s.endReason }
func (s *Session) TimeoutDuration() time.Duration { return s.timeoutDuration }
func (s *Session) LastActivityAt() time.Time      { return s.lastActivityAt }
func (s *Session) MessageCount() int              { return s.messageCount }
func (s *Session) MessagesFromContact() int       { return s.messagesFromContact }
func (s *Session) MessagesFromAgent() int         { return s.messagesFromAgent }
func (s *Session) DurationSeconds() int           { return s.durationSeconds }

func (s *Session) FirstContactMessageAt() *time.Time { return s.firstContactMessageAt }
func (s *Session) FirstAgentResponseAt() *time.Time  { return s.firstAgentResponseAt }
func (s *Session) AgentResponseTimeSeconds() *int    { return s.agentResponseTimeSeconds }
func (s *Session) ContactWaitTimeSeconds() *int      { return s.contactWaitTimeSeconds }

func (s *Session) AgentIDs() []uuid.UUID    { return append([]uuid.UUID{}, s.agentIDs...) }
func (s *Session) AgentTransfers() int      { return s.agentTransfers }
func (s *Session) Summary() *string         { return s.summary }
func (s *Session) Sentiment() *Sentiment    { return s.sentiment }
func (s *Session) SentimentScore() *float64 { return s.sentimentScore }
func (s *Session) Topics() []string         { return append([]string{}, s.topics...) }
func (s *Session) NextSteps() []string      { return append([]string{}, s.nextSteps...) }
func (s *Session) KeyEntities() map[string]interface{} {

	copy := make(map[string]interface{})
	for k, v := range s.keyEntities {
		copy[k] = v
	}
	return copy
}
func (s *Session) IsResolved() bool      { return s.resolved }
func (s *Session) IsEscalated() bool     { return s.escalated }
func (s *Session) IsConverted() bool     { return s.converted }
func (s *Session) OutcomeTags() []string { return append([]string{}, s.outcomeTags...) }

func (s *Session) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, s.events...)
}

func (s *Session) ClearEvents() {
	s.events = []DomainEvent{}
}

func (s *Session) addEvent(event DomainEvent) {
	s.events = append(s.events, event)
}

// Compile-time check that Session implements AggregateRoot interface
var _ shared.AggregateRoot = (*Session)(nil)
