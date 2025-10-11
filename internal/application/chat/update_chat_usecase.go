package chat

import (
	"context"
	"errors"
	"fmt"

	domainchat "github.com/caloi/ventros-crm/internal/domain/chat"
	"github.com/google/uuid"
)

// UpdateChatUseCase handles updating chat properties (subject, description)
type UpdateChatUseCase struct {
	chatRepo domainchat.Repository
	eventBus EventBus
}

// NewUpdateChatUseCase creates a new instance of the use case
func NewUpdateChatUseCase(
	chatRepo domainchat.Repository,
	eventBus EventBus,
) *UpdateChatUseCase {
	return &UpdateChatUseCase{
		chatRepo: chatRepo,
		eventBus: eventBus,
	}
}

// UpdateSubject updates the subject of a group or channel chat
func (uc *UpdateChatUseCase) UpdateSubject(ctx context.Context, input UpdateChatSubjectInput) (*UpdateChatSubjectOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}
	if input.Subject == "" {
		return nil, errors.New("subject is required")
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Update subject
	if err := chat.UpdateSubject(input.Subject); err != nil {
		return nil, fmt.Errorf("failed to update subject: %w", err)
	}

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
	return &UpdateChatSubjectOutput{
		Chat: ChatToDTO(chat),
	}, nil
}
