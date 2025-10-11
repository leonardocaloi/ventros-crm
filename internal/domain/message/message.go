package message

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	id               uuid.UUID
	timestamp        time.Time
	customerID       uuid.UUID
	projectID        uuid.UUID
	channelTypeID    *int
	fromMe           bool
	channelID        uuid.UUID
	contactID        uuid.UUID
	sessionID        *uuid.UUID
	contentType      ContentType
	text             *string
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
	mentions         []string // IDs externos mencionados (formato WAHA: "phone@c.us")

	events []DomainEvent
}

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

	msg.addEvent(NewMessageCreatedEvent(msg.id, contactID, fromMe))

	return msg, nil
}

func ReconstructMessage(
	id uuid.UUID,
	timestamp time.Time,
	customerID uuid.UUID,
	projectID uuid.UUID,
	channelTypeID *int,
	fromMe bool,
	channelID uuid.UUID,
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
	mentions []string,
) *Message {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	if mentions == nil {
		mentions = []string{}
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
		mentions:         mentions,
		events:           []DomainEvent{},
	}
}

func (m *Message) SetText(text string) error {
	if !m.contentType.IsText() {
		return errors.New("cannot set text on non-text message")
	}
	m.text = &text
	return nil
}

func (m *Message) SetMediaContent(url, mimetype string) error {
	if !m.contentType.IsMedia() {
		return errors.New("cannot set media content on non-media message")
	}
	m.mediaURL = &url
	m.mediaMimetype = &mimetype
	return nil
}

func (m *Message) HasMediaURL() bool {
	return m.contentType.IsMedia() && m.mediaURL != nil
}

func (m *Message) AssignToChannel(channelID uuid.UUID, channelTypeID *int) {
	m.channelID = channelID
	m.channelTypeID = channelTypeID
}

func (m *Message) AssignToSession(sessionID uuid.UUID) {
	m.sessionID = &sessionID
}

func (m *Message) SetChannelMessageID(channelMessageID string) {
	m.channelMessageID = &channelMessageID
}

func (m *Message) MarkAsDelivered() {
	now := time.Now()
	m.status = StatusDelivered
	m.deliveredAt = &now

	m.addEvent(MessageDeliveredEvent{
		MessageID:   m.id,
		DeliveredAt: now,
	})
}

func (m *Message) MarkAsRead() {
	now := time.Now()
	m.status = StatusRead
	m.readAt = &now

	m.addEvent(MessageReadEvent{
		MessageID: m.id,
		ReadAt:    now,
	})
}

func (m *Message) MarkAsFailed() {
	m.status = StatusFailed
}

// SetMentions define as menções da mensagem (IDs externos no formato WAHA: "phone@c.us")
func (m *Message) SetMentions(mentions []string) {
	if mentions == nil {
		m.mentions = []string{}
	} else {
		m.mentions = append([]string{}, mentions...) // Copiar para evitar mutação externa
	}
}

// HasMentions verifica se a mensagem contém menções
func (m *Message) HasMentions() bool {
	return len(m.mentions) > 0
}

// IsMentioned verifica se um ID externo específico foi mencionado
// externalID deve estar no formato WAHA: "phone@c.us"
func (m *Message) IsMentioned(externalID string) bool {
	for _, mention := range m.mentions {
		if mention == externalID {
			return true
		}
	}
	return false
}

func (m *Message) IsInbound() bool {
	return !m.fromMe
}

func (m *Message) IsOutbound() bool {
	return m.fromMe
}

func (m *Message) ID() uuid.UUID             { return m.id }
func (m *Message) Timestamp() time.Time      { return m.timestamp }
func (m *Message) CustomerID() uuid.UUID     { return m.customerID }
func (m *Message) ProjectID() uuid.UUID      { return m.projectID }
func (m *Message) ChannelTypeID() *int       { return m.channelTypeID }
func (m *Message) FromMe() bool              { return m.fromMe }
func (m *Message) ChannelID() uuid.UUID      { return m.channelID }
func (m *Message) ContactID() uuid.UUID      { return m.contactID }
func (m *Message) SessionID() *uuid.UUID     { return m.sessionID }
func (m *Message) ContentType() ContentType  { return m.contentType }
func (m *Message) Text() *string             { return m.text }
func (m *Message) MediaMimetype() *string    { return m.mediaMimetype }
func (m *Message) MediaURL() *string         { return m.mediaURL }
func (m *Message) ChannelMessageID() *string { return m.channelMessageID }
func (m *Message) ReplyToID() *uuid.UUID     { return m.replyToID }
func (m *Message) Status() Status            { return m.status }
func (m *Message) Language() *string         { return m.language }
func (m *Message) AgentID() *uuid.UUID       { return m.agentID }
func (m *Message) Metadata() map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range m.metadata {
		copy[k] = v
	}
	return copy
}
func (m *Message) DeliveredAt() *time.Time { return m.deliveredAt }
func (m *Message) ReadAt() *time.Time      { return m.readAt }
func (m *Message) Mentions() []string {
	return append([]string{}, m.mentions...) // Retornar cópia para evitar mutação
}

func (m *Message) DomainEvents() []DomainEvent {
	return append([]DomainEvent{}, m.events...)
}

func (m *Message) ClearEvents() {
	m.events = []DomainEvent{}
}

func (m *Message) addEvent(event DomainEvent) {
	m.events = append(m.events, event)
}

func (m *Message) RequestAIProcessing(channelConfig AIProcessingConfig) {
	if m.mediaURL == nil || *m.mediaURL == "" {
		return
	}

	if m.fromMe {
		return
	}

	if m.sessionID == nil {
		return
	}

	mimeType := ""
	if m.mediaMimetype != nil {
		mimeType = *m.mediaMimetype
	}

	switch m.contentType {
	case ContentTypeImage:
		if channelConfig.ProcessImage {
			m.addEvent(NewAIProcessImageRequestedEvent(
				m.id,
				m.channelID,
				m.contactID,
				*m.sessionID,
				*m.mediaURL,
				mimeType,
			))
		}

	case ContentTypeVideo:
		if channelConfig.ProcessVideo {
			duration := 0
			m.addEvent(NewAIProcessVideoRequestedEvent(
				m.id,
				m.channelID,
				m.contactID,
				*m.sessionID,
				*m.mediaURL,
				mimeType,
				duration,
			))
		}

	case ContentTypeAudio:
		if channelConfig.ProcessAudio {
			duration := 0
			m.addEvent(NewAIProcessAudioRequestedEvent(
				m.id,
				m.channelID,
				m.contactID,
				*m.sessionID,
				*m.mediaURL,
				mimeType,
				duration,
			))
		}

	case ContentTypeVoice:
		if channelConfig.ProcessVoice {
			duration := 0
			m.addEvent(NewAIProcessVoiceRequestedEvent(
				m.id,
				m.channelID,
				m.contactID,
				*m.sessionID,
				*m.mediaURL,
				mimeType,
				duration,
			))
		}
	}
}

type AIProcessingConfig struct {
	ProcessImage bool
	ProcessVideo bool
	ProcessAudio bool
	ProcessVoice bool
}
