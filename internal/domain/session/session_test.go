package session_test

import (
	"testing"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err := sess.RecordMessage(true)

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
	sess.RecordMessage(true)  // contato
	sess.RecordMessage(false) // agente
	sess.RecordMessage(true)  // contato
	sess.RecordMessage(false) // agente

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
	err := sess.RecordMessage(true)

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
				sess.RecordMessage(true)
				sess.RecordMessage(true)
				sess.RecordMessage(true)
				return sess
			},
			expectedResult: false,
		},
		{
			name: "ended with enough messages should generate",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true)
				sess.RecordMessage(true)
				sess.RecordMessage(true)
				sess.End(session.ReasonManualClose)
				return sess
			},
			expectedResult: true,
		},
		{
			name: "ended with too few messages should not generate",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true)
				sess.End(session.ReasonManualClose)
				return sess
			},
			expectedResult: false,
		},
		{
			name: "already summarized should not generate again",
			setupSession: func() *session.Session {
				sess, _ := session.NewSession(uuid.New(), "tenant-1", nil, 30*time.Minute)
				sess.RecordMessage(true)
				sess.RecordMessage(true)
				sess.RecordMessage(true)
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
		contactID,
		tenantID,
		&channelTypeID,
		startedAt,
		&endedAt,
		status,
		&endReason,
		timeout,
		lastActivity,
		10,  // messageCount
		6,   // messagesFromContact
		4,   // messagesFromAgent
		3600, // durationSeconds
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
