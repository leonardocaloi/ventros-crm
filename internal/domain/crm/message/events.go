package message

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

type DomainEvent = shared.DomainEvent

type MessageCreatedEvent struct {
	shared.BaseEvent
	MessageID uuid.UUID
	ContactID uuid.UUID
	FromMe    bool
	CreatedAt time.Time
}

func NewMessageCreatedEvent(messageID, contactID uuid.UUID, fromMe bool) MessageCreatedEvent {
	return MessageCreatedEvent{
		BaseEvent: shared.NewBaseEvent("message.created", time.Now()),
		MessageID: messageID,
		ContactID: contactID,
		FromMe:    fromMe,
		CreatedAt: time.Now(),
	}
}

type MessageDeliveredEvent struct {
	shared.BaseEvent
	MessageID   uuid.UUID
	DeliveredAt time.Time
}

func NewMessageDeliveredEvent(messageID uuid.UUID) MessageDeliveredEvent {
	return MessageDeliveredEvent{
		BaseEvent:   shared.NewBaseEvent("message.delivered", time.Now()),
		MessageID:   messageID,
		DeliveredAt: time.Now(),
	}
}

type MessageReadEvent struct {
	shared.BaseEvent
	MessageID uuid.UUID
	ReadAt    time.Time
}

func NewMessageReadEvent(messageID uuid.UUID) MessageReadEvent {
	return MessageReadEvent{
		BaseEvent: shared.NewBaseEvent("message.read", time.Now()),
		MessageID: messageID,
		ReadAt:    time.Now(),
	}
}

type MessageFailedEvent struct {
	shared.BaseEvent
	MessageID     uuid.UUID
	FailureReason string
	FailedAt      time.Time
}

func NewMessageFailedEvent(messageID uuid.UUID, failureReason string) MessageFailedEvent {
	return MessageFailedEvent{
		BaseEvent:     shared.NewBaseEvent("message.failed", time.Now()),
		MessageID:     messageID,
		FailureReason: failureReason,
		FailedAt:      time.Now(),
	}
}

type AIProcessImageRequestedEvent struct {
	shared.BaseEvent
	MessageID   uuid.UUID
	ChannelID   uuid.UUID
	ContactID   uuid.UUID
	SessionID   uuid.UUID
	ImageURL    string
	MimeType    string
	RequestedAt time.Time
}

func NewAIProcessImageRequestedEvent(messageID, channelID, contactID, sessionID uuid.UUID, imageURL, mimeType string) AIProcessImageRequestedEvent {
	return AIProcessImageRequestedEvent{
		BaseEvent:   shared.NewBaseEvent("message.ai.process_image_requested", time.Now()),
		MessageID:   messageID,
		ChannelID:   channelID,
		ContactID:   contactID,
		SessionID:   sessionID,
		ImageURL:    imageURL,
		MimeType:    mimeType,
		RequestedAt: time.Now(),
	}
}

type AIProcessVideoRequestedEvent struct {
	shared.BaseEvent
	MessageID   uuid.UUID
	ChannelID   uuid.UUID
	ContactID   uuid.UUID
	SessionID   uuid.UUID
	VideoURL    string
	MimeType    string
	Duration    int
	RequestedAt time.Time
}

func NewAIProcessVideoRequestedEvent(messageID, channelID, contactID, sessionID uuid.UUID, videoURL, mimeType string, duration int) AIProcessVideoRequestedEvent {
	return AIProcessVideoRequestedEvent{
		BaseEvent:   shared.NewBaseEvent("message.ai.process_video_requested", time.Now()),
		MessageID:   messageID,
		ChannelID:   channelID,
		ContactID:   contactID,
		SessionID:   sessionID,
		VideoURL:    videoURL,
		MimeType:    mimeType,
		Duration:    duration,
		RequestedAt: time.Now(),
	}
}

type AIProcessAudioRequestedEvent struct {
	shared.BaseEvent
	MessageID   uuid.UUID
	ChannelID   uuid.UUID
	ContactID   uuid.UUID
	SessionID   uuid.UUID
	AudioURL    string
	MimeType    string
	Duration    int
	RequestedAt time.Time
}

func NewAIProcessAudioRequestedEvent(messageID, channelID, contactID, sessionID uuid.UUID, audioURL, mimeType string, duration int) AIProcessAudioRequestedEvent {
	return AIProcessAudioRequestedEvent{
		BaseEvent:   shared.NewBaseEvent("message.ai.process_audio_requested", time.Now()),
		MessageID:   messageID,
		ChannelID:   channelID,
		ContactID:   contactID,
		SessionID:   sessionID,
		AudioURL:    audioURL,
		MimeType:    mimeType,
		Duration:    duration,
		RequestedAt: time.Now(),
	}
}

type AIProcessVoiceRequestedEvent struct {
	shared.BaseEvent
	MessageID   uuid.UUID
	ChannelID   uuid.UUID
	ContactID   uuid.UUID
	SessionID   uuid.UUID
	VoiceURL    string
	MimeType    string
	Duration    int
	RequestedAt time.Time
}

func NewAIProcessVoiceRequestedEvent(messageID, channelID, contactID, sessionID uuid.UUID, voiceURL, mimeType string, duration int) AIProcessVoiceRequestedEvent {
	return AIProcessVoiceRequestedEvent{
		BaseEvent:   shared.NewBaseEvent("message.ai.process_voice_requested", time.Now()),
		MessageID:   messageID,
		ChannelID:   channelID,
		ContactID:   contactID,
		SessionID:   sessionID,
		VoiceURL:    voiceURL,
		MimeType:    mimeType,
		Duration:    duration,
		RequestedAt: time.Now(),
	}
}
