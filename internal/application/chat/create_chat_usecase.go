package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	domainchat "github.com/ventros/crm/internal/domain/crm/chat"
)

// CreateChatUseCase handles chat creation
type CreateChatUseCase struct {
	chatRepo domainchat.Repository
	eventBus EventBus
}

// NewCreateChatUseCase creates a new instance of the use case
func NewCreateChatUseCase(
	chatRepo domainchat.Repository,
	eventBus EventBus,
) *CreateChatUseCase {
	return &CreateChatUseCase{
		chatRepo: chatRepo,
		eventBus: eventBus,
	}
}

// Execute executes the create chat use case
func (uc *CreateChatUseCase) Execute(ctx context.Context, input CreateChatInput) (*CreateChatOutput, error) {
	// 1. Basic validations
	if input.ProjectID == uuid.Nil {
		return nil, errors.New("project_id is required")
	}
	if input.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if input.ChatType == "" {
		return nil, errors.New("chat_type is required")
	}

	// 2. Create chat based on type
	var chat *domainchat.Chat
	var err error

	switch input.ChatType {
	case "individual":
		// Individual chat requires contact_id
		if input.ContactID == nil || *input.ContactID == uuid.Nil {
			return nil, errors.New("contact_id is required for individual chats")
		}

		// Check if individual chat already exists for this contact in this project
		existingChat, err := uc.chatRepo.FindIndividualByContact(ctx, *input.ContactID, input.ProjectID)
		if err == nil && existingChat != nil {
			// Chat already exists, return it
			return &CreateChatOutput{
				Chat: ChatToDTO(existingChat),
			}, nil
		}

		chat, err = domainchat.NewIndividualChat(input.ProjectID, input.TenantID, *input.ContactID)
		if err != nil {
			return nil, fmt.Errorf("failed to create individual chat: %w", err)
		}

	case "group":
		// Group chat requires subject and creator_id
		if input.Subject == nil || *input.Subject == "" {
			return nil, errors.New("subject is required for group chats")
		}
		if input.CreatorID == nil || *input.CreatorID == uuid.Nil {
			return nil, errors.New("creator_id is required for group chats")
		}

		chat, err = domainchat.NewGroupChat(input.ProjectID, input.TenantID, *input.Subject, *input.CreatorID, input.ExternalID)
		if err != nil {
			return nil, fmt.Errorf("failed to create group chat: %w", err)
		}

	case "channel":
		// Channel chat requires subject
		if input.Subject == nil || *input.Subject == "" {
			return nil, errors.New("subject is required for channel chats")
		}

		chat, err = domainchat.NewChannelChat(input.ProjectID, input.TenantID, *input.Subject)
		if err != nil {
			return nil, fmt.Errorf("failed to create channel chat: %w", err)
		}

	default:
		return nil, fmt.Errorf("invalid chat_type: %s (must be 'individual', 'group', or 'channel')", input.ChatType)
	}

	// 3. Save chat to repository
	if err := uc.chatRepo.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to save chat: %w", err)
	}

	// 4. Publish domain events
	for _, event := range chat.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the operation (event is secondary)
			fmt.Printf("Failed to publish chat event: %v\n", err)
		}
	}

	// 5. Clear events after publishing
	chat.ClearEvents()

	// 6. Return result
	return &CreateChatOutput{
		Chat: ChatToDTO(chat),
	}, nil
}
