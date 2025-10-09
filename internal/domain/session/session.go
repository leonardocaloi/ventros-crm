package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Session é o Aggregate Root para conversas.
// Agrupa messages e garante invariantes de negócio.
//
// ⚠️ IMPORTANTE - Dependências:
// - Session PRECISA de Pipeline ativo para ser criada
// - Session PRECISA de Channel para receber mensagens
// - Pipeline e Channel são independentes entre si
// - Sem Pipeline: Mensagens são salvas, mas não agrupadas em Session
//
// O timeout vem do Pipeline.SessionTimeoutMinutes (não é arbitrário).
type Session struct {
	id              uuid.UUID
	contactID       uuid.UUID
	tenantID        string
	channelTypeID   *int
	pipelineID      *uuid.UUID // Pipeline que define o timeout e fluxo
	startedAt       time.Time
	endedAt         *time.Time
	status          Status
	endReason       *EndReason
	timeoutDuration time.Duration // Vem do Pipeline.SessionTimeoutMinutes
	lastActivityAt  time.Time

	// Métricas
	messageCount        int
	messagesFromContact int
	messagesFromAgent   int
	durationSeconds     int

	// Response Time Metrics (para lead score e feedback comercial)
	firstContactMessageAt    *time.Time
	firstAgentResponseAt     *time.Time
	agentResponseTimeSeconds *int // Tempo de espera até primeira resposta do agente
	contactWaitTimeSeconds   *int // Tempo de espera do contato (se agente iniciou)

	// Agentes
	agentIDs       []uuid.UUID
	agentTransfers int

	// AI/Analytics
	summary        *string
	sentiment      *Sentiment
	sentimentScore *float64
	topics         []string
	nextSteps      []string
	keyEntities    map[string]interface{}

	// Flags de negócio
	resolved    bool
	escalated   bool
	converted   bool
	outcomeTags []string

	// Domain Events
	events []DomainEvent
}

// NewSession cria uma nova sessão (factory method).
// IMPORTANTE: pipelineID é opcional para compatibilidade, mas RECOMENDADO.
// Se não informado, usa fallback de 30min (não ideal para produção).
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
		timeoutDuration = 30 * time.Minute // default fallback
	}

	now := time.Now()
	session := &Session{
		id:              uuid.New(),
		contactID:       contactID,
		tenantID:        tenantID,
		channelTypeID:   channelTypeID,
		pipelineID:      nil, // Será definido pela aplicação
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

// NewSessionWithPipeline cria uma nova sessão COM pipeline (recomendado).
// Este é o método preferido para criar sessões em produção.
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

// ReconstructSession reconstrói uma Session a partir de dados persistidos.
func ReconstructSession(
	id uuid.UUID,
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

	return &Session{
		id:                       id,
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

// RecordMessage registra uma mensagem na sessão e calcula métricas de tempo de resposta.
func (s *Session) RecordMessage(fromContact bool, messageTimestamp time.Time) error {
	if s.status != StatusActive {
		return errors.New("cannot add message to non-active session")
	}

	s.messageCount++
	now := time.Now()

	if fromContact {
		s.messagesFromContact++

		// Registrar primeira mensagem do contato
		if s.firstContactMessageAt == nil {
			s.firstContactMessageAt = &messageTimestamp

			// Se agente já respondeu antes, calcular tempo de espera do contato
			if s.firstAgentResponseAt != nil {
				waitTime := int(messageTimestamp.Sub(*s.firstAgentResponseAt).Seconds())
				s.contactWaitTimeSeconds = &waitTime
			}
		}

		// Se agente ainda não respondeu, calcular tempo de espera do agente
		if s.firstAgentResponseAt == nil && s.firstContactMessageAt != nil {
			responseTime := int(now.Sub(*s.firstContactMessageAt).Seconds())
			// Atualiza continuamente até o agente responder
			s.agentResponseTimeSeconds = &responseTime
		}
	} else {
		s.messagesFromAgent++

		// Registrar primeira resposta do agente
		if s.firstAgentResponseAt == nil {
			s.firstAgentResponseAt = &messageTimestamp

			// Se contato já enviou mensagem, calcular tempo de resposta do agente
			if s.firstContactMessageAt != nil {
				responseTime := int(messageTimestamp.Sub(*s.firstContactMessageAt).Seconds())
				s.agentResponseTimeSeconds = &responseTime
			}
		}

		// Se contato ainda não respondeu, calcular tempo de espera do contato
		if s.firstContactMessageAt == nil && s.firstAgentResponseAt != nil {
			waitTime := int(now.Sub(*s.firstAgentResponseAt).Seconds())
			// Atualiza continuamente até o contato responder
			s.contactWaitTimeSeconds = &waitTime
		}
	}

	s.lastActivityAt = now

	s.addEvent(NewMessageRecordedEvent(s.id, fromContact))

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

	s.addEvent(NewAgentAssignedEvent(s.id, agentID))

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
// Nota: O evento criado aqui será enriquecido pela camada de aplicação com channel_id,
// message_ids e outros dados de contexto antes de ser publicado.
func (s *Session) End(reason EndReason) error {
	if s.status != StatusActive {
		return errors.New("session is not active")
	}

	now := time.Now()
	s.status = StatusEnded
	s.endedAt = &now
	s.endReason = &reason
	s.durationSeconds = int(now.Sub(s.startedAt).Seconds())

	// Cria evento básico - será enriquecido pela aplicação com channelID, messageIDs, etc
	s.addEvent(NewSessionEndedEvent(
		s.id,
		s.contactID,
		s.tenantID,
		nil, // channelID será adicionado pela camada de aplicação
		s.channelTypeID,
		s.pipelineID,
		s.startedAt,
		reason,
		s.durationSeconds,
	))

	return nil
}

// Resolve marca a sessão como resolvida.
func (s *Session) Resolve() error {
	if s.status == StatusActive {
		return errors.New("cannot resolve active session")
	}

	s.resolved = true

	s.addEvent(NewSessionResolvedEvent(s.id))

	return nil
}

// Escalate marca a sessão como escalada.
func (s *Session) Escalate() error {
	s.escalated = true

	s.addEvent(NewSessionEscalatedEvent(s.id))

	return nil
}

// SetSummary define o resumo gerado por IA.
func (s *Session) SetSummary(summary string, sentiment Sentiment, score float64, topics, nextSteps []string) {
	s.summary = &summary
	s.sentiment = &sentiment
	s.sentimentScore = &score
	s.topics = topics
	s.nextSteps = nextSteps

	s.addEvent(NewSessionSummarizedEvent(s.id, summary, sentiment, score))
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

func (s *Session) ID() uuid.UUID                  { return s.id }
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

// Response Time Metrics Getters
func (s *Session) FirstContactMessageAt() *time.Time { return s.firstContactMessageAt }
func (s *Session) FirstAgentResponseAt() *time.Time  { return s.firstAgentResponseAt }
func (s *Session) AgentResponseTimeSeconds() *int    { return s.agentResponseTimeSeconds }
func (s *Session) ContactWaitTimeSeconds() *int      { return s.contactWaitTimeSeconds }

func (s *Session) AgentIDs() []uuid.UUID    { return append([]uuid.UUID{}, s.agentIDs...) } // Copy
func (s *Session) AgentTransfers() int      { return s.agentTransfers }
func (s *Session) Summary() *string         { return s.summary }
func (s *Session) Sentiment() *Sentiment    { return s.sentiment }
func (s *Session) SentimentScore() *float64 { return s.sentimentScore }
func (s *Session) Topics() []string         { return append([]string{}, s.topics...) }    // Copy
func (s *Session) NextSteps() []string      { return append([]string{}, s.nextSteps...) } // Copy
func (s *Session) KeyEntities() map[string]interface{} {
	// Return copy
	copy := make(map[string]interface{})
	for k, v := range s.keyEntities {
		copy[k] = v
	}
	return copy
}
func (s *Session) IsResolved() bool      { return s.resolved }
func (s *Session) IsEscalated() bool     { return s.escalated }
func (s *Session) IsConverted() bool     { return s.converted }
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
