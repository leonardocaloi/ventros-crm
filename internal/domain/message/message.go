package message

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Message é o Aggregate Root para mensagens individuais.
type Message struct {
	id               uuid.UUID
	timestamp        time.Time
	customerID       uuid.UUID
	projectID        uuid.UUID
	channelTypeID    *int
	fromMe           bool
	channelID        *uuid.UUID
	contactID        uuid.UUID
	sessionID        *uuid.UUID
	contentType      ContentType
	text             *string
	// Campos técnicos de mídia (preenchidos pela camada de infraestrutura)
	mediaURL         *string
	mediaMimetype    *string
	channelMessageID *string
	replyToID        *uuid.UUID
	status           Status
	language         *string
	agentID          *uuid.UUID
	metadata         map[string]interface{}
	deliveredAt      *time.Time
	readAt           *time.Time
	
	events []DomainEvent
}

// NewMessage cria uma nova mensagem.
func NewMessage(
	contactID, projectID, customerID uuid.UUID,
	contentType ContentType,
	fromMe bool,
) (*Message, error) {
	if contactID == uuid.Nil {
		return nil, errors.New("contactID cannot be nil")
	}
	if projectID == uuid.Nil {
		return nil, errors.New("projectID cannot be nil")
	}
	if customerID == uuid.Nil {
		return nil, errors.New("customerID cannot be nil")
	}
	if !contentType.IsValid() {
		return nil, errors.New("invalid content type")
	}

	now := time.Now()
	msg := &Message{
		id:          uuid.New(),
		timestamp:   now,
		customerID:  customerID,
		projectID:   projectID,
		contactID:   contactID,
		contentType: contentType,
		fromMe:      fromMe,
		status:      StatusSent,
		metadata:    make(map[string]interface{}),
		events:      []DomainEvent{},
	}

	msg.addEvent(MessageCreatedEvent{
		MessageID: msg.id,
		ContactID: contactID,
		FromMe:    fromMe,
		CreatedAt: now,
	})

	return msg, nil
}

// ReconstructMessage reconstrói uma Message a partir de dados persistidos.
func ReconstructMessage(
	id uuid.UUID,
	timestamp time.Time,
	customerID uuid.UUID,
	projectID uuid.UUID,
	channelTypeID *int,
	fromMe bool,
	channelID *uuid.UUID,
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	contentType ContentType,
	text *string,
	mediaURL *string,
	mediaMimetype *string,
	channelMessageID *string,
	replyToID *uuid.UUID,
	status Status,
	language *string,
	agentID *uuid.UUID,
	metadata map[string]interface{},
	deliveredAt *time.Time,
	readAt *time.Time,
) *Message {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &Message{
		id:               id,
		timestamp:        timestamp,
		customerID:       customerID,
		projectID:        projectID,
		channelTypeID:    channelTypeID,
		fromMe:           fromMe,
		channelID:        channelID,
		contactID:        contactID,
		sessionID:        sessionID,
		contentType:      contentType,
		text:             text,
		mediaURL:         mediaURL,
		mediaMimetype:    mediaMimetype,
		channelMessageID: channelMessageID,
		replyToID:        replyToID,
		status:           status,
		language:         language,
		agentID:          agentID,
		metadata:         metadata,
		deliveredAt:      deliveredAt,
		readAt:           readAt,
		events:           []DomainEvent{},
	}
}

// SetText define o conteúdo de texto da mensagem.
func (m *Message) SetText(text string) error {
	if !m.contentType.IsText() {
		return errors.New("cannot set text on non-text message")
	}
	m.text = &text
	return nil
}

// SetMediaContent define URL e mimetype de mídia.
// Usado pela camada de infraestrutura após fazer upload/download.
func (m *Message) SetMediaContent(url, mimetype string) error {
	if !m.contentType.IsMedia() {
		return errors.New("cannot set media content on non-media message")
	}
	m.mediaURL = &url
	m.mediaMimetype = &mimetype
	return nil
}

// HasMediaURL verifica se a mensagem tem URL de mídia.
func (m *Message) HasMediaURL() bool {
	return m.contentType.IsMedia() && m.mediaURL != nil
}

// AssignToSession atribui a mensagem a uma sessão.
func (m *Message) AssignToSession(sessionID uuid.UUID) {
	m.sessionID = &sessionID
}

// MarkAsDelivered marca a mensagem como entregue.
func (m *Message) MarkAsDelivered() {
	now := time.Now()
	m.status = StatusDelivered
	m.deliveredAt = &now
	
	m.addEvent(MessageDeliveredEvent{
		MessageID:   m.id,
		DeliveredAt: now,
	})
}

// MarkAsRead marca a mensagem como lida.
func (m *Message) MarkAsRead() {
	now := time.Now()
	m.status = StatusRead
	m.readAt = &now
	
	m.addEvent(MessageReadEvent{
		MessageID: m.id,
		ReadAt:    now,
	})
}

// MarkAsFailed marca a mensagem como falha.
func (m *Message) MarkAsFailed() {
	m.status = StatusFailed
}

// IsInbound verifica se a mensagem é inbound (do contato).
func (m *Message) IsInbound() bool {
	return !m.fromMe
}

// IsOutbound verifica se a mensagem é outbound (para o contato).
func (m *Message) IsOutbound() bool {
	return m.fromMe
}

// Getters
func (m *Message) ID() uuid.UUID              { return m.id }
func (m *Message) Timestamp() time.Time       { return m.timestamp }
func (m *Message) CustomerID() uuid.UUID      { return m.customerID }
func (m *Message) ProjectID() uuid.UUID       { return m.projectID }
func (m *Message) ChannelTypeID() *int        { return m.channelTypeID }
func (m *Message) FromMe() bool               { return m.fromMe }
func (m *Message) ChannelID() *uuid.UUID      { return m.channelID }
func (m *Message) ContactID() uuid.UUID       { return m.contactID }
func (m *Message) SessionID() *uuid.UUID      { return m.sessionID }
func (m *Message) ContentType() ContentType   { return m.contentType }
func (m *Message) Text() *string              { return m.text }
func (m *Message) MediaMimetype() *string     { return m.mediaMimetype }
func (m *Message) MediaURL() *string          { return m.mediaURL }
func (m *Message) ChannelMessageID() *string  { return m.channelMessageID }
func (m *Message) ReplyToID() *uuid.UUID      { return m.replyToID }
func (m *Message) Status() Status             { return m.status }
func (m *Message) Language() *string          { return m.language }
func (m *Message) AgentID() *uuid.UUID        { return m.agentID }
func (m *Message) Metadata() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range m.metadata {
		copy[k] = v
	}
	return copy
}
func (m *Message) DeliveredAt() *time.Time { return m.deliveredAt }
func (m *Message) ReadAt() *time.Time      { return m.readAt }

// Domain Events
func (m *Message) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, m.events...)
}

func (m *Message) ClearEvents() {
	m.events = []DomainEvent{}
}

func (m *Message) addEvent(event DomainEvent) {
	m.events = append(m.events, event)
}
