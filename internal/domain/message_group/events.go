package message_group

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// MessageGroupCreatedEvent - Grupo de mensagens criado
type MessageGroupCreatedEvent struct {
	shared.BaseEvent
	GroupID   uuid.UUID `json:"group_id"`
	ContactID uuid.UUID `json:"contact_id"`
	ChannelID uuid.UUID `json:"channel_id"`
}

func NewMessageGroupCreatedEvent(groupID, contactID, channelID uuid.UUID) MessageGroupCreatedEvent {
	return MessageGroupCreatedEvent{
		BaseEvent: shared.NewBaseEvent("message_group.created", time.Now()),
		GroupID:   groupID,
		ContactID: contactID,
		ChannelID: channelID,
	}
}

// MessageAddedToGroupEvent - Mensagem adicionada ao grupo
type MessageAddedToGroupEvent struct {
	shared.BaseEvent
	GroupID   uuid.UUID `json:"group_id"`
	MessageID uuid.UUID `json:"message_id"`
}

func NewMessageAddedToGroupEvent(groupID, messageID uuid.UUID) MessageAddedToGroupEvent {
	return MessageAddedToGroupEvent{
		BaseEvent: shared.NewBaseEvent("message_group.message_added", time.Now()),
		GroupID:   groupID,
		MessageID: messageID,
	}
}

// MessageGroupProcessingEvent - Grupo começou processamento de enriquecimentos
type MessageGroupProcessingEvent struct {
	shared.BaseEvent
	GroupID      uuid.UUID `json:"group_id"`
	MessageCount int       `json:"message_count"`
}

func NewMessageGroupProcessingEvent(groupID uuid.UUID, messageCount int) MessageGroupProcessingEvent {
	return MessageGroupProcessingEvent{
		BaseEvent:    shared.NewBaseEvent("message_group.processing", time.Now()),
		GroupID:      groupID,
		MessageCount: messageCount,
	}
}

// MessageGroupCompletedEvent - Grupo concluído e enviado para AI Agent
type MessageGroupCompletedEvent struct {
	shared.BaseEvent
	GroupID      uuid.UUID `json:"group_id"`
	MessageCount int       `json:"message_count"`
}

func NewMessageGroupCompletedEvent(groupID uuid.UUID, messageCount int) MessageGroupCompletedEvent {
	return MessageGroupCompletedEvent{
		BaseEvent:    shared.NewBaseEvent("message_group.completed", time.Now()),
		GroupID:      groupID,
		MessageCount: messageCount,
	}
}

// MessageGroupExpiredEvent - Grupo expirou sem completar
type MessageGroupExpiredEvent struct {
	shared.BaseEvent
	GroupID uuid.UUID `json:"group_id"`
}

func NewMessageGroupExpiredEvent(groupID uuid.UUID) MessageGroupExpiredEvent {
	return MessageGroupExpiredEvent{
		BaseEvent: shared.NewBaseEvent("message_group.expired", time.Now()),
		GroupID:   groupID,
	}
}
