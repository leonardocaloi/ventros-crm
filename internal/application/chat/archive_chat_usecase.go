package chat

import (
	"context"
	"errors"
	"fmt"

	domainchat "github.com/caloi/ventros-crm/internal/domain/chat"
	"github.com/google/uuid"
)

// ArchiveChatUseCase handles chat status transitions (archive, unarchive, close)
type ArchiveChatUseCase struct {
	chatRepo domainchat.Repository
	eventBus EventBus
}

// NewArchiveChatUseCase creates a new instance of the use case
func NewArchiveChatUseCase(
	chatRepo domainchat.Repository,
	eventBus EventBus,
) *ArchiveChatUseCase {
	return &ArchiveChatUseCase{
		chatRepo: chatRepo,
		eventBus: eventBus,
	}
}

// Archive archives a chat
func (uc *ArchiveChatUseCase) Archive(ctx context.Context, input ArchiveChatInput) (*ArchiveChatOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Archive chat
	chat.Archive()

	// 4. Save chat
	if err := uc.chatRepo.Update(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	// 5. Publish domain events
	for _, event := range chat.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish chat event: %v\n", err)
		}
	}

	// 6. Clear events
	chat.ClearEvents()

	// 7. Return result
	return &ArchiveChatOutput{
		Chat: ChatToDTO(chat),
	}, nil
}

// Unarchive unarchives a chat
func (uc *ArchiveChatUseCase) Unarchive(ctx context.Context, input UnarchiveChatInput) (*UnarchiveChatOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Unarchive chat
	chat.Unarchive()

	// 4. Save chat
	if err := uc.chatRepo.Update(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	// 5. Publish domain events
	for _, event := range chat.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish chat event: %v\n", err)
		}
	}

	// 6. Clear events
	chat.ClearEvents()

	// 7. Return result
	return &UnarchiveChatOutput{
		Chat: ChatToDTO(chat),
	}, nil
}

// Close permanently closes a chat
func (uc *ArchiveChatUseCase) Close(ctx context.Context, input CloseChatInput) (*CloseChatOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Close chat
	chat.Close()

	// 4. Save chat
	if err := uc.chatRepo.Update(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	// 5. Publish domain events
	for _, event := range chat.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish chat event: %v\n", err)
		}
	}

	// 6. Clear events
	chat.ClearEvents()

	// 7. Return result
	return &CloseChatOutput{
		Chat: ChatToDTO(chat),
	}, nil
}
