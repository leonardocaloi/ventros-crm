package saga

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

// ProcessInboundMessageInput representa os dados de entrada para a Saga de processamento de mensagem.
type ProcessInboundMessageInput struct {
	MessageID        string
	ChannelMessageID string
	FromPhone        string
	MessageText      string
	Timestamp        time.Time
	MessageType      string
	MediaURL         string
	MediaType        string
	// Channel context
	ChannelID uuid.UUID
	// User context
	ProjectID  uuid.UUID
	CustomerID uuid.UUID
	TenantID   string
	// Additional fields
	ContactPhone    string
	ContactName     string
	ChannelTypeID   int
	ContentType     string
	Text            string
	MediaMimetype   string
	TrackingData    map[string]interface{}
	ReceivedAt      time.Time
	Metadata        map[string]interface{}
	FromMe          bool
	IsGroupMessage  bool
	GroupExternalID string
	Participant     string
	Mentions        []string
	ChatID          *uuid.UUID
}

// SagaState mantém o estado da Saga durante a execução.
// Usado para compensação em caso de falha.
type SagaState struct {
	// IDs dos recursos criados
	ContactID uuid.UUID
	SessionID uuid.UUID
	MessageID uuid.UUID

	// Flags indicando se recursos foram criados (vs já existiam)
	ContactCreated bool
	SessionCreated bool

	// Eventos coletados durante a execução
	Events []shared.DomainEvent

	// Metadados
	CorrelationID string
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         string
}

// ContactCreatedResult resultado da activity FindOrCreateContact.
type ContactCreatedResult struct {
	ContactID  uuid.UUID
	WasCreated bool // true se foi criado, false se já existia
}

// SessionCreatedResult resultado da activity FindOrCreateSession.
type SessionCreatedResult struct {
	SessionID  uuid.UUID
	WasCreated bool // true se foi criado, false se já existia
}

// MessageCreatedResult resultado da activity CreateMessage.
type MessageCreatedResult struct {
	MessageID uuid.UUID
}

// SendMessageInput representa os dados de entrada para a Saga de envio de mensagem.
type SendMessageInput struct {
	ContactID   uuid.UUID
	ChannelID   uuid.UUID
	ContentType string
	Text        *string
	MediaURL    *string
	ReplyToID   *uuid.UUID
	Metadata    map[string]interface{}
	// Contexto de autenticação
	TenantID   string
	ProjectID  uuid.UUID
	CustomerID uuid.UUID
	AgentID    *uuid.UUID
}

// SendMessageState mantém o estado da Saga de envio.
type SendMessageState struct {
	SessionID  uuid.UUID
	MessageID  uuid.UUID
	ExternalID *string // ID retornado pelo canal (WAHA)

	SessionCreated bool
	MessageSent    bool

	CorrelationID string
	StartedAt     time.Time
	CompletedAt   *time.Time
	Error         string
}

// SendMessageResult resultado da Saga de envio.
type SendMessageResult struct {
	MessageID  uuid.UUID
	ExternalID *string
	Status     string
	SentAt     time.Time
	Error      *string
}
