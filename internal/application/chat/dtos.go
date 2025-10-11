package chat

import (
	"time"

	"github.com/google/uuid"
)

// ChatDTO represents a chat in the application layer
type ChatDTO struct {
	ID            uuid.UUID            `json:"id"`
	ProjectID     uuid.UUID            `json:"project_id"`
	TenantID      string               `json:"tenant_id"`
	ChatType      string               `json:"chat_type"` // "individual", "group", "channel"
	ExternalID    *string              `json:"external_id,omitempty"` // External ID from channel (WhatsApp @g.us, Telegram group ID)
	Subject       *string              `json:"subject,omitempty"`
	Description   *string              `json:"description,omitempty"`
	Participants  []ParticipantDTO     `json:"participants"`
	Status        string               `json:"status"` // "active", "archived", "closed"
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	LastMessageAt *time.Time           `json:"last_message_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// ParticipantDTO represents a participant in a chat
type ParticipantDTO struct {
	ID       uuid.UUID  `json:"id"`
	Type     string     `json:"type"` // "contact" or "agent"
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
	IsAdmin  bool       `json:"is_admin"`
}

// CreateChatInput represents input for creating a chat
type CreateChatInput struct {
	ProjectID  uuid.UUID `json:"project_id"`
	TenantID   string    `json:"tenant_id"`
	ChatType   string    `json:"chat_type"`
	ContactID  *uuid.UUID `json:"contact_id,omitempty"` // For individual chats
	CreatorID  *uuid.UUID `json:"creator_id,omitempty"` // For group chats
	Subject    *string   `json:"subject,omitempty"`    // For group/channel chats
	ExternalID *string   `json:"external_id,omitempty"` // External ID from channel (WhatsApp @g.us, Telegram group ID)
}

// CreateChatOutput represents the result of creating a chat
type CreateChatOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// FindChatInput represents input for finding a chat
type FindChatInput struct {
	ChatID uuid.UUID `json:"chat_id"`
}

// FindChatOutput represents the result of finding a chat
type FindChatOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// ListChatsInput represents input for listing chats
type ListChatsInput struct {
	ProjectID  *uuid.UUID `json:"project_id,omitempty"`
	TenantID   *string    `json:"tenant_id,omitempty"`
	ContactID  *uuid.UUID `json:"contact_id,omitempty"`
	Status     *string    `json:"status,omitempty"`
	ChatType   *string    `json:"chat_type,omitempty"`
}

// ListChatsOutput represents the result of listing chats
type ListChatsOutput struct {
	Chats []*ChatDTO `json:"chats"`
	Total int        `json:"total"`
}

// AddParticipantInput represents input for adding a participant
type AddParticipantInput struct {
	ChatID          uuid.UUID `json:"chat_id"`
	ParticipantID   uuid.UUID `json:"participant_id"`
	ParticipantType string    `json:"participant_type"` // "contact" or "agent"
}

// AddParticipantOutput represents the result of adding a participant
type AddParticipantOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// RemoveParticipantInput represents input for removing a participant
type RemoveParticipantInput struct {
	ChatID        uuid.UUID `json:"chat_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
}

// RemoveParticipantOutput represents the result of removing a participant
type RemoveParticipantOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// ArchiveChatInput represents input for archiving a chat
type ArchiveChatInput struct {
	ChatID uuid.UUID `json:"chat_id"`
}

// ArchiveChatOutput represents the result of archiving a chat
type ArchiveChatOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// UnarchiveChatInput represents input for unarchiving a chat
type UnarchiveChatInput struct {
	ChatID uuid.UUID `json:"chat_id"`
}

// UnarchiveChatOutput represents the result of unarchiving a chat
type UnarchiveChatOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// CloseChatInput represents input for closing a chat
type CloseChatInput struct {
	ChatID uuid.UUID `json:"chat_id"`
}

// CloseChatOutput represents the result of closing a chat
type CloseChatOutput struct {
	Chat *ChatDTO `json:"chat"`
}

// UpdateChatSubjectInput represents input for updating chat subject
type UpdateChatSubjectInput struct {
	ChatID  uuid.UUID `json:"chat_id"`
	Subject string    `json:"subject"`
}

// UpdateChatSubjectOutput represents the result of updating chat subject
type UpdateChatSubjectOutput struct {
	Chat *ChatDTO `json:"chat"`
}
