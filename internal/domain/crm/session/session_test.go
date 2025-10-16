package session_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ventros/crm/internal/domain/crm/session"
)

func TestNewSession_Success(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	tenantID := "tenant-123"
	channelTypeID := 1
	timeout := 30 * time.Minute

	// Act
	sess, err := session.NewSession(contactID, tenantID, &channelTypeID, timeout)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, sess.ID())
	assert.Equal(t, contactID, sess.ContactID())
	assert.Equal(t, tenantID, sess.TenantID())
	assert.NotNil(t, sess.ChannelTypeID())
	assert.Equal(t, channelTypeID, *sess.ChannelTypeID())
	assert.Equal(t, session.StatusActive, sess.Status())
	assert.Equal(t, timeout, sess.TimeoutDuration())
	assert.True(t, sess.IsActive())
	assert.Equal(t, 0, sess.MessageCount())

	// Deve ter evento SessionStarted
	events := sess.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "session.started", events[0].EventName())
}

func TestNewSession_DefaultTimeout(t *testing.T) {
	// Act
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 0)

	// Assert - deve usar default de 30 minutos
	assert.Equal(t, 30*time.Minute, sess.TimeoutDuration())
}

func TestNewSession_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		contactID uuid.UUID
		tenantID  string
		wantErr   string
	}{
		{
			name:      "nil contact ID",
			contactID: uuid.Nil,
			tenantID:  "tenant-1",
			wantErr:   "contactID cannot be nil",
		},
		{
			name:      "empty tenant ID",
			contactID: uuid.New(),
			tenantID:  "",
			wantErr:   "tenantID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess, err := session.NewSession(tt.contactID, tt.tenantID, nil, 30*time.Minute)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Nil(t, sess)
		})
	}
}

func TestSession_RecordMessage_WhenActive_Success(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	sess.ClearEvents() // Limpa evento de criação

	// Act - registra mensagem de contato
	err := sess.RecordMessage(true, time.Now())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, sess.MessageCount())
	assert.Equal(t, 1, sess.MessagesFromContact())
	assert.Equal(t, 0, sess.MessagesFromAgent())

	// Deve emitir evento
	events := sess.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "session.message_recorded", events[0].EventName())
}

func TestSession_RecordMessage_MultipleMessages(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)

	// Act
	sess.RecordMessage(true, time.Now())  // contato
	sess.RecordMessage(false, time.Now()) // agente
	sess.RecordMessage(true, time.Now())  // contato
	sess.RecordMessage(false, time.Now()) // agente

	// Assert
	assert.Equal(t, 4, sess.MessageCount())
	assert.Equal(t, 2, sess.MessagesFromContact())
	assert.Equal(t, 2, sess.MessagesFromAgent())
}

func TestSession_RecordMessage_WhenEnded_Error(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	sess.End(session.ReasonManualClose)

	// Act
	err := sess.RecordMessage(true, time.Now())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add message to non-active session")
}

func TestSession_AssignAgent_FirstAssignment(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	agentID := uuid.New()
	sess.ClearEvents()

	// Act
	err := sess.AssignAgent(agentID)

	// Assert
	require.NoError(t, err)
	agents := sess.AgentIDs()
	assert.Len(t, agents, 1)
	assert.Equal(t, agentID, agents[0])
	assert.Equal(t, 0, sess.AgentTransfers()) // Primeira atribuição não é transferência

	// Deve emitir evento
	events := sess.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "session.agent_assigned", events[0].EventName())
}

func TestSession_AssignAgent_Transfer(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	agent1 := uuid.New()
	agent2 := uuid.New()
	sess.AssignAgent(agent1)

	// Act - segunda atribuição é transferência
	err := sess.AssignAgent(agent2)

	// Assert
	require.NoError(t, err)
	agents := sess.AgentIDs()
	assert.Len(t, agents, 2)
	assert.Equal(t, 1, sess.AgentTransfers())
}

func TestSession_AssignAgent_Idempotent(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	agentID := uuid.New()
	sess.AssignAgent(agentID)

	// Act - atribui o mesmo agente novamente
	err := sess.AssignAgent(agentID)

	// Assert
	require.NoError(t, err)
	agents := sess.AgentIDs()
	assert.Len(t, agents, 1) // Não duplica
}

func TestSession_CheckTimeout_WhenInactive_ReturnsFalse(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 1*time.Second)
	time.Sleep(2 * time.Second)

	// Act
	timedOut := sess.CheckTimeout()

	// Assert
	assert.True(t, timedOut)
	assert.Equal(t, session.StatusEnded, sess.Status())
	assert.NotNil(t, sess.EndedAt())
	assert.NotNil(t, sess.EndReason())
	assert.Equal(t, session.ReasonInactivityTimeout, *sess.EndReason())
}

func TestSession_CheckTimeout_WhenStillActive_ReturnsFalse(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 1*time.Hour)

	// Act
	timedOut := sess.CheckTimeout()

	// Assert
	assert.False(t, timedOut)
	assert.True(t, sess.IsActive())
}

func TestSession_End_Success(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	sess.ClearEvents()

	// Act
	err := sess.End(session.ReasonManualClose)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, session.StatusEnded, sess.Status())
	assert.NotNil(t, sess.EndedAt())
	assert.GreaterOrEqual(t, sess.DurationSeconds(), 0)
	assert.Equal(t, session.ReasonManualClose, *sess.EndReason())

	// Deve emitir evento
	events := sess.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "session.ended", events[0].EventName())
}

func TestSession_Escalate(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	sess.ClearEvents()

	// Act
	err := sess.Escalate()

	// Assert
	require.NoError(t, err)
	assert.True(t, sess.IsEscalated())

	// Deve emitir evento
	events := sess.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "session.escalated", events[0].EventName())
}

func TestSession_SetSummary(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	sess.End(session.ReasonManualClose)
	sess.ClearEvents()

	summary := "Customer asked about pricing"
	sentiment := session.SentimentPositive
	score := 0.8
	topics := []string{"pricing", "sales"}
	nextSteps := []string{"send proposal"}

	// Act
	sess.SetSummary(summary, sentiment, score, topics, nextSteps)

	// Assert
	assert.NotNil(t, sess.Summary())
	assert.Equal(t, summary, *sess.Summary())
	assert.NotNil(t, sess.Sentiment())
	assert.Equal(t, sentiment, *sess.Sentiment())
	assert.NotNil(t, sess.SentimentScore())
	assert.Equal(t, score, *sess.SentimentScore())
	assert.Equal(t, topics, sess.Topics())
	assert.Equal(t, nextSteps, sess.NextSteps())

	// Deve emitir evento
	events := sess.DomainEvents()
	assert.Len(t, events, 1)
	assert.Equal(t, "session.summarized", events[0].EventName())
}

func TestSession_ShouldGenerateSummary(t *testing.T) {
	tests := []struct {
		name           string
		setupSession   func() *session.Session
		expectedResult bool
	}{
		{
			name: "active session should not generate",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true, time.Now())
				sess.RecordMessage(true, time.Now())
				sess.RecordMessage(true, time.Now())
				return sess
			},
			expectedResult: false,
		},
		{
			name: "ended with enough messages should generate",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true, time.Now())
				sess.RecordMessage(true, time.Now())
				sess.RecordMessage(true, time.Now())
				sess.End(session.ReasonManualClose)
				return sess
			},
			expectedResult: true,
		},
		{
			name: "ended with too few messages should not generate",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true, time.Now())
				sess.End(session.ReasonManualClose)
				return sess
			},
			expectedResult: false,
		},
		{
			name: "already summarized should not generate again",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true, time.Now())
				sess.RecordMessage(true, time.Now())
				sess.RecordMessage(true, time.Now())
				sess.End(session.ReasonManualClose)
				sess.SetSummary("summary", session.SentimentNeutral, 0.0, []string{}, []string{})
				return sess
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := tt.setupSession()
			result := sess.ShouldGenerateSummary()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// ===========================
// 1.3.4 - Testes de Métricas de Resposta (complementares)
// ===========================

func TestRecordMessage_FirstContactMessage_SetsTimestamp(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	messageTimestamp := time.Now()

	// Act
	err := sess.RecordMessage(true, messageTimestamp)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, sess.FirstContactMessageAt())
	assert.Nil(t, sess.FirstAgentResponseAt(), "Agent response should still be nil")
}

func TestRecordMessage_FirstAgentResponse_SetsTimestamp(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	messageTimestamp := time.Now()

	// Act
	err := sess.RecordMessage(false, messageTimestamp)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, sess.FirstAgentResponseAt())
	assert.Nil(t, sess.FirstContactMessageAt(), "Contact message should still be nil")
}

func TestRecordMessage_AgentResponseTime_WhenContactFirst(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	contactTime := time.Now()

	// Contact sends first
	err := sess.RecordMessage(true, contactTime)
	require.NoError(t, err)

	// Wait at least 1 second for measurable difference
	time.Sleep(1100 * time.Millisecond)
	agentTime := time.Now()

	// Act
	err = sess.RecordMessage(false, agentTime)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, sess.AgentResponseTimeSeconds())
	assert.GreaterOrEqual(t, *sess.AgentResponseTimeSeconds(), 1, "Response time should be at least 1 second")
	assert.LessOrEqual(t, *sess.AgentResponseTimeSeconds(), 3, "Response time should be less than 3 seconds")
}

func TestRecordMessage_ContactWaitTime_WhenAgentFirst(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	agentTime := time.Now()

	// Agent sends first (proactive message)
	err := sess.RecordMessage(false, agentTime)
	require.NoError(t, err)

	// Wait at least 1 second for measurable difference
	time.Sleep(1100 * time.Millisecond)
	contactTime := time.Now()

	// Act
	err = sess.RecordMessage(true, contactTime)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, sess.ContactWaitTimeSeconds())
	assert.GreaterOrEqual(t, *sess.ContactWaitTimeSeconds(), 1, "Wait time should be at least 1 second")
	assert.LessOrEqual(t, *sess.ContactWaitTimeSeconds(), 3, "Wait time should be less than 3 seconds")
}

func TestRecordMessage_UpdatesLastActivityTimestamp(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	originalLastActivity := sess.LastActivityAt()
	time.Sleep(50 * time.Millisecond)

	// Act
	err := sess.RecordMessage(true, time.Now())

	// Assert
	require.NoError(t, err)
	assert.True(t, sess.LastActivityAt().After(originalLastActivity),
		"LastActivityAt should be updated after recording message")
}

// ===========================
// Additional Tests
// ===========================

func TestNewSessionWithPipeline_Success(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	tenantID := "tenant-123"
	channelTypeID := 1
	pipelineID := uuid.New()
	timeout := 45 * time.Minute

	// Act
	sess, err := session.NewSessionWithPipeline(contactID, tenantID, &channelTypeID, pipelineID, timeout)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, sess)
	require.NotNil(t, sess.PipelineID())
	assert.Equal(t, pipelineID, *sess.PipelineID())
	assert.Equal(t, timeout, sess.TimeoutDuration())
}

func TestNewSessionWithPipeline_NilPipelineID_Error(t *testing.T) {
	// Arrange
	contactID := uuid.New()
	tenantID := "tenant-123"
	channelTypeID := 1
	timeout := 30 * time.Minute

	// Act
	sess, err := session.NewSessionWithPipeline(contactID, tenantID, &channelTypeID, uuid.Nil, timeout)

	// Assert
	require.Error(t, err)
	assert.Nil(t, sess)
	assert.Contains(t, err.Error(), "pipelineID cannot be nil")
}

func TestSession_End_AlreadyEnded_Error(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	_ = sess.End(session.ReasonManualClose)

	// Act - try to end again
	err := sess.End(session.ReasonInactivityTimeout)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session is not active")
}

func TestSession_End_CalculatesDuration(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	time.Sleep(1100 * time.Millisecond)

	// Act
	err := sess.End(session.ReasonManualClose)

	// Assert
	require.NoError(t, err)
	assert.GreaterOrEqual(t, sess.DurationSeconds(), 1, "Duration should be at least 1 second")
}

func TestSession_AssignAgent_WhenEnded_Error(t *testing.T) {
	// Arrange
	sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
	_ = sess.End(session.ReasonManualClose)
	agentID := uuid.New()

	// Act
	err := sess.AssignAgent(agentID)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot assign agent to non-active session")
}

func TestReconstructSession(t *testing.T) {
	// Arrange
	id := uuid.New()
	contactID := uuid.New()
	tenantID := "tenant-123"
	channelTypeID := 1
	startedAt := time.Now().Add(-1 * time.Hour)
	endedAt := time.Now()
	status := session.StatusEnded
	endReason := session.ReasonManualClose
	timeout := 30 * time.Minute
	lastActivity := time.Now().Add(-5 * time.Minute)
	agentID1 := uuid.New()
	agentID2 := uuid.New()
	summary := "Test summary"
	sentiment := session.SentimentPositive
	score := 0.8

	// Act
	sess := session.ReconstructSession(
		id,
		1, // version
		contactID,
		tenantID,
		&channelTypeID,
		nil, // pipelineID
		startedAt,
		&endedAt,
		status,
		&endReason,
		timeout,
		lastActivity,
		10,   // messageCount
		6,    // messagesFromContact
		4,    // messagesFromAgent
		3600, // durationSeconds
		nil,  // firstContactMessageAt
		nil,  // firstAgentResponseAt
		nil,  // agentResponseTimeSeconds
		nil,  // contactWaitTimeSeconds
		[]uuid.UUID{agentID1, agentID2},
		1, // agentTransfers
		&summary,
		&sentiment,
		&score,
		[]string{"topic1"},
		[]string{"next1"},
		map[string]interface{}{"key": "value"},
		true,  // resolved
		false, // escalated
		true,  // converted
		[]string{"tag1"},
	)

	// Assert
	assert.Equal(t, id, sess.ID())
	assert.Equal(t, contactID, sess.ContactID())
	assert.Equal(t, tenantID, sess.TenantID())
	assert.Equal(t, status, sess.Status())
	assert.Equal(t, 10, sess.MessageCount())
	assert.Equal(t, 6, sess.MessagesFromContact())
	assert.Equal(t, 4, sess.MessagesFromAgent())
	assert.Len(t, sess.AgentIDs(), 2)
	assert.Equal(t, 1, sess.AgentTransfers())
	assert.True(t, sess.IsResolved())
	assert.False(t, sess.IsEscalated())
	assert.True(t, sess.IsConverted())
	assert.NotNil(t, sess.Summary())
	assert.Equal(t, summary, *sess.Summary())

	// Não deve ter eventos após reconstituir
	assert.Empty(t, sess.DomainEvents())
}
