package chat

import (
	"context"
	"errors"
	"fmt"

	domainchat "github.com/caloi/ventros-crm/internal/domain/crm/chat"
	"github.com/google/uuid"
)

// FindChatUseCase handles finding chats
type FindChatUseCase struct {
	chatRepo domainchat.Repository
}

// NewFindChatUseCase creates a new instance of the use case
func NewFindChatUseCase(chatRepo domainchat.Repository) *FindChatUseCase {
	return &FindChatUseCase{
		chatRepo: chatRepo,
	}
}

// FindByID finds a single chat by ID
func (uc *FindChatUseCase) FindByID(ctx context.Context, input FindChatInput) (*FindChatOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Return result
	return &FindChatOutput{
		Chat: ChatToDTO(chat),
	}, nil
}

// ListChats lists chats based on filters
func (uc *FindChatUseCase) ListChats(ctx context.Context, input ListChatsInput) (*ListChatsOutput, error) {
	var chats []*domainchat.Chat
	var err error

	// Apply filters based on input
	if input.ContactID != nil {
		// Find chats where contact is a participant
		chats, err = uc.chatRepo.FindByContact(ctx, *input.ContactID)
	} else if input.ProjectID != nil {
		// Find chats by project
		if input.Status != nil && *input.Status == "active" {
			chats, err = uc.chatRepo.FindActiveByProject(ctx, *input.ProjectID)
		} else {
			chats, err = uc.chatRepo.FindByProject(ctx, *input.ProjectID)
		}
	} else if input.TenantID != nil {
		// Find chats by tenant
		chats, err = uc.chatRepo.FindByTenant(ctx, *input.TenantID)
	} else {
		return nil, errors.New("at least one filter (contact_id, project_id, or tenant_id) is required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	// Post-filter by status and chat_type if specified
	filteredChats := chats
	if input.Status != nil || input.ChatType != nil {
		filteredChats = make([]*domainchat.Chat, 0, len(chats))
		for _, c := range chats {
			// Check status filter
			if input.Status != nil && c.Status().String() != *input.Status {
				continue
			}
			// Check chat_type filter
			if input.ChatType != nil && c.ChatType().String() != *input.ChatType {
				continue
			}
			filteredChats = append(filteredChats, c)
		}
	}

	// Convert to DTOs
	chatDTOs := ChatsToDTO(filteredChats)

	return &ListChatsOutput{
		Chats: chatDTOs,
		Total: len(chatDTOs),
	}, nil
}
