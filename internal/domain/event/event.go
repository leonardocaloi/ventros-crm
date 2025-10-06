package event

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Event é o Aggregate Root para eventos do sistema (timeline).
// Representa fatos que aconteceram e são visíveis na timeline de contatos.
type Event struct {
	id             uuid.UUID
	contactID      *uuid.UUID
	sessionID      *uuid.UUID
	messageID      *uuid.UUID
	tenantID       string
	eventType      string
	payload        map[string]interface{}
	source         EventSource
	sequenceNumber *int
	timestamp      time.Time
	createdAt      time.Time
}

// NewEvent cria um novo evento de timeline.
func NewEvent(
	tenantID string,
	eventType string,
	source EventSource,
	payload map[string]interface{},
) (*Event, error) {
	if tenantID == "" {
		return nil, errors.New("tenantID cannot be empty")
	}
	if eventType == "" {
		return nil, errors.New("eventType cannot be empty")
	}
	if !source.IsValid() {
		return nil, errors.New("invalid event source")
	}

	if payload == nil {
		payload = make(map[string]interface{})
	}

	now := time.Now()
	return &Event{
		id:        uuid.New(),
		tenantID:  tenantID,
		eventType: eventType,
		source:    source,
		payload:   payload,
		timestamp: now,
		createdAt: now,
	}, nil
}

// ReconstructEvent reconstrói um Event a partir de dados persistidos.
func ReconstructEvent(
	id uuid.UUID,
	contactID *uuid.UUID,
	sessionID *uuid.UUID,
	messageID *uuid.UUID,
	tenantID string,
	eventType string,
	payload map[string]interface{},
	source EventSource,
	sequenceNumber *int,
	timestamp time.Time,
	createdAt time.Time,
) *Event {
	if payload == nil {
		payload = make(map[string]interface{})
	}

	return &Event{
		id:             id,
		contactID:      contactID,
		sessionID:      sessionID,
		messageID:      messageID,
		tenantID:       tenantID,
		eventType:      eventType,
		payload:        payload,
		source:         source,
		sequenceNumber: sequenceNumber,
		timestamp:      timestamp,
		createdAt:      createdAt,
	}
}

// AttachToContact associa o evento a um contato.
func (e *Event) AttachToContact(contactID uuid.UUID) error {
	if contactID == uuid.Nil {
		return errors.New("contactID cannot be nil")
	}
	e.contactID = &contactID
	return nil
}

// AttachToSession associa o evento a uma sessão.
func (e *Event) AttachToSession(sessionID uuid.UUID) error {
	if sessionID == uuid.Nil {
		return errors.New("sessionID cannot be nil")
	}
	e.sessionID = &sessionID
	return nil
}

// AttachToMessage associa o evento a uma mensagem.
func (e *Event) AttachToMessage(messageID uuid.UUID) error {
	if messageID == uuid.Nil {
		return errors.New("messageID cannot be nil")
	}
	e.messageID = &messageID
	return nil
}

// SetSequenceNumber define o número de sequência do evento dentro da sessão.
func (e *Event) SetSequenceNumber(seq int) error {
	if seq < 0 {
		return errors.New("sequence number cannot be negative")
	}
	e.sequenceNumber = &seq
	return nil
}

// AddPayloadField adiciona um campo ao payload.
func (e *Event) AddPayloadField(key string, value interface{}) {
	e.payload[key] = value
}

// GetPayloadField retorna um campo do payload.
func (e *Event) GetPayloadField(key string) (interface{}, bool) {
	val, ok := e.payload[key]
	return val, ok
}

// IsSystemGenerated verifica se o evento foi gerado pelo sistema.
func (e *Event) IsSystemGenerated() bool {
	return e.source == EventSourceSystem
}

// IsWebhookGenerated verifica se o evento foi gerado por webhook.
func (e *Event) IsWebhookGenerated() bool {
	return e.source == EventSourceWebhook
}

// IsManual verifica se o evento foi criado manualmente.
func (e *Event) IsManual() bool {
	return e.source == EventSourceManual
}

// HasContact verifica se o evento está associado a um contato.
func (e *Event) HasContact() bool {
	return e.contactID != nil
}

// HasSession verifica se o evento está associado a uma sessão.
func (e *Event) HasSession() bool {
	return e.sessionID != nil
}

// HasMessage verifica se o evento está associado a uma mensagem.
func (e *Event) HasMessage() bool {
	return e.messageID != nil
}

// Getters
func (e *Event) ID() uuid.UUID                       { return e.id }
func (e *Event) ContactID() *uuid.UUID               { return e.contactID }
func (e *Event) SessionID() *uuid.UUID               { return e.sessionID }
func (e *Event) MessageID() *uuid.UUID               { return e.messageID }
func (e *Event) TenantID() string                    { return e.tenantID }
func (e *Event) EventType() string                   { return e.eventType }
func (e *Event) Source() EventSource                 { return e.source }
func (e *Event) SequenceNumber() *int                { return e.sequenceNumber }
func (e *Event) Timestamp() time.Time                { return e.timestamp }
func (e *Event) CreatedAt() time.Time                { return e.createdAt }
func (e *Event) Payload() map[string]interface{} {
	// Return copy
	copy := make(map[string]interface{})
	for k, v := range e.payload {
		copy[k] = v
	}
	return copy
}
