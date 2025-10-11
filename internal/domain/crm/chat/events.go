package chat

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/google/uuid"
)

// DomainEvent represents a domain event (compatible with shared.DomainEvent)
type DomainEvent interface {
	EventType() string // Legacy method name
	shared.DomainEvent // Embed shared.DomainEvent interface
}

// ChatCreatedEvent is emitted when a new chat is created
type ChatCreatedEvent struct {
	shared.BaseEvent
	ChatID    uuid.UUID
	ChatType  ChatType
	ProjectID uuid.UUID
}

func (e ChatCreatedEvent) EventType() string {
	return "chat.created"
}

func NewChatCreatedEvent(chatID uuid.UUID, chatType ChatType, projectID uuid.UUID) ChatCreatedEvent {
	return ChatCreatedEvent{
		BaseEvent: shared.NewBaseEvent("chat.created", time.Now()),
		ChatID:    chatID,
		ChatType:  chatType,
		ProjectID: projectID,
	}
}

// ParticipantAddedEvent is emitted when a participant is added to a chat
type ParticipantAddedEvent struct {
	shared.BaseEvent
	ChatID          uuid.UUID
	ParticipantID   uuid.UUID
	ParticipantType ParticipantType
}

func (e ParticipantAddedEvent) EventType() string {
	return "chat.participant_added"
}

func NewParticipantAddedEvent(chatID, participantID uuid.UUID, participantType ParticipantType) ParticipantAddedEvent {
	return ParticipantAddedEvent{
		BaseEvent:       shared.NewBaseEvent("chat.participant_added", time.Now()),
		ChatID:          chatID,
		ParticipantID:   participantID,
		ParticipantType: participantType,
	}
}

// ParticipantRemovedEvent is emitted when a participant is removed from a chat
type ParticipantRemovedEvent struct {
	shared.BaseEvent
	ChatID        uuid.UUID
	ParticipantID uuid.UUID
}

func (e ParticipantRemovedEvent) EventType() string {
	return "chat.participant_removed"
}

func NewParticipantRemovedEvent(chatID, participantID uuid.UUID) ParticipantRemovedEvent {
	return ParticipantRemovedEvent{
		BaseEvent:     shared.NewBaseEvent("chat.participant_removed", time.Now()),
		ChatID:        chatID,
		ParticipantID: participantID,
	}
}

// ChatArchivedEvent is emitted when a chat is archived
type ChatArchivedEvent struct {
	shared.BaseEvent
	ChatID uuid.UUID
}

func (e ChatArchivedEvent) EventType() string {
	return "chat.archived"
}

func NewChatArchivedEvent(chatID uuid.UUID) ChatArchivedEvent {
	return ChatArchivedEvent{
		BaseEvent: shared.NewBaseEvent("chat.archived", time.Now()),
		ChatID:    chatID,
	}
}

// ChatUnarchivedEvent is emitted when a chat is unarchived
type ChatUnarchivedEvent struct {
	shared.BaseEvent
	ChatID uuid.UUID
}

func (e ChatUnarchivedEvent) EventType() string {
	return "chat.unarchived"
}

func NewChatUnarchivedEvent(chatID uuid.UUID) ChatUnarchivedEvent {
	return ChatUnarchivedEvent{
		BaseEvent: shared.NewBaseEvent("chat.unarchived", time.Now()),
		ChatID:    chatID,
	}
}

// ChatClosedEvent is emitted when a chat is permanently closed
type ChatClosedEvent struct {
	shared.BaseEvent
	ChatID uuid.UUID
}

func (e ChatClosedEvent) EventType() string {
	return "chat.closed"
}

func NewChatClosedEvent(chatID uuid.UUID) ChatClosedEvent {
	return ChatClosedEvent{
		BaseEvent: shared.NewBaseEvent("chat.closed", time.Now()),
		ChatID:    chatID,
	}
}

// ChatSubjectUpdatedEvent is emitted when a chat subject (group/channel name) is updated
type ChatSubjectUpdatedEvent struct {
	shared.BaseEvent
	ChatID     uuid.UUID
	NewSubject string
}

func (e ChatSubjectUpdatedEvent) EventType() string {
	return "chat.subject_updated"
}

func NewChatSubjectUpdatedEvent(chatID uuid.UUID, newSubject string) ChatSubjectUpdatedEvent {
	return ChatSubjectUpdatedEvent{
		BaseEvent:  shared.NewBaseEvent("chat.subject_updated", time.Now()),
		ChatID:     chatID,
		NewSubject: newSubject,
	}
}

// ChatDescriptionUpdatedEvent is emitted when a chat description is updated
type ChatDescriptionUpdatedEvent struct {
	shared.BaseEvent
	ChatID         uuid.UUID
	NewDescription string
}

func (e ChatDescriptionUpdatedEvent) EventType() string {
	return "chat.description_updated"
}

func NewChatDescriptionUpdatedEvent(chatID uuid.UUID, newDescription string) ChatDescriptionUpdatedEvent {
	return ChatDescriptionUpdatedEvent{
		BaseEvent:      shared.NewBaseEvent("chat.description_updated", time.Now()),
		ChatID:         chatID,
		NewDescription: newDescription,
	}
}

// ParticipantPromotedEvent is emitted when a participant is promoted to admin
type ParticipantPromotedEvent struct {
	shared.BaseEvent
	ChatID        uuid.UUID
	ParticipantID uuid.UUID
}

func (e ParticipantPromotedEvent) EventType() string {
	return "chat.participant_promoted"
}

func NewParticipantPromotedEvent(chatID, participantID uuid.UUID) ParticipantPromotedEvent {
	return ParticipantPromotedEvent{
		BaseEvent:     shared.NewBaseEvent("chat.participant_promoted", time.Now()),
		ChatID:        chatID,
		ParticipantID: participantID,
	}
}

// ParticipantDemotedEvent is emitted when a participant is demoted from admin
type ParticipantDemotedEvent struct {
	shared.BaseEvent
	ChatID        uuid.UUID
	ParticipantID uuid.UUID
}

func (e ParticipantDemotedEvent) EventType() string {
	return "chat.participant_demoted"
}

func NewParticipantDemotedEvent(chatID, participantID uuid.UUID) ParticipantDemotedEvent {
	return ParticipantDemotedEvent{
		BaseEvent:     shared.NewBaseEvent("chat.participant_demoted", time.Now()),
		ChatID:        chatID,
		ParticipantID: participantID,
	}
}

// ChatLabelAddedEvent is emitted when a label is added to a chat
type ChatLabelAddedEvent struct {
	shared.BaseEvent
	ChatID  uuid.UUID
	LabelID string
}

func (e ChatLabelAddedEvent) EventType() string {
	return "chat.label_added"
}

func NewChatLabelAddedEvent(chatID uuid.UUID, labelID string) ChatLabelAddedEvent {
	return ChatLabelAddedEvent{
		BaseEvent: shared.NewBaseEvent("chat.label_added", time.Now()),
		ChatID:    chatID,
		LabelID:   labelID,
	}
}

// ChatLabelRemovedEvent is emitted when a label is removed from a chat
type ChatLabelRemovedEvent struct {
	shared.BaseEvent
	ChatID  uuid.UUID
	LabelID string
}

func (e ChatLabelRemovedEvent) EventType() string {
	return "chat.label_removed"
}

func NewChatLabelRemovedEvent(chatID uuid.UUID, labelID string) ChatLabelRemovedEvent {
	return ChatLabelRemovedEvent{
		BaseEvent: shared.NewBaseEvent("chat.label_removed", time.Now()),
		ChatID:    chatID,
		LabelID:   labelID,
	}
}
