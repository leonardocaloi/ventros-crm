package chat

import (
	"context"
	"errors"
	"fmt"

	domainchat "github.com/caloi/ventros-crm/internal/domain/crm/chat"
	"github.com/google/uuid"
)

// ManageParticipantsUseCase handles adding and removing participants from chats
type ManageParticipantsUseCase struct {
	chatRepo domainchat.Repository
	eventBus EventBus
}

// NewManageParticipantsUseCase creates a new instance of the use case
func NewManageParticipantsUseCase(
	chatRepo domainchat.Repository,
	eventBus EventBus,
) *ManageParticipantsUseCase {
	return &ManageParticipantsUseCase{
		chatRepo: chatRepo,
		eventBus: eventBus,
	}
}

// AddParticipant adds a participant to a chat
func (uc *ManageParticipantsUseCase) AddParticipant(ctx context.Context, input AddParticipantInput) (*AddParticipantOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}
	if input.ParticipantID == uuid.Nil {
		return nil, errors.New("participant_id is required")
	}
	if input.ParticipantType == "" {
		return nil, errors.New("participant_type is required")
	}

	// Validate participant type
	if input.ParticipantType != "contact" && input.ParticipantType != "agent" {
		return nil, fmt.Errorf("invalid participant_type: %s (must be 'contact' or 'agent')", input.ParticipantType)
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Parse participant type
	participantType, err := domainchat.ParseParticipantType(input.ParticipantType)
	if err != nil {
		return nil, fmt.Errorf("invalid participant_type: %w", err)
	}

	// 4. Add participant
	if err := chat.AddParticipant(input.ParticipantID, participantType); err != nil {
		return nil, fmt.Errorf("failed to add participant: %w", err)
	}

	// 5. Save chat
	if err := uc.chatRepo.Update(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	// 6. Publish domain events
	for _, event := range chat.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish chat event: %v\n", err)
		}
	}

	// 7. Clear events
	chat.ClearEvents()

	// 8. Return result
	return &AddParticipantOutput{
		Chat: ChatToDTO(chat),
	}, nil
}

// RemoveParticipant removes a participant from a chat
func (uc *ManageParticipantsUseCase) RemoveParticipant(ctx context.Context, input RemoveParticipantInput) (*RemoveParticipantOutput, error) {
	// 1. Validate input
	if input.ChatID == uuid.Nil {
		return nil, errors.New("chat_id is required")
	}
	if input.ParticipantID == uuid.Nil {
		return nil, errors.New("participant_id is required")
	}

	// 2. Find chat
	chat, err := uc.chatRepo.FindByID(ctx, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chat: %w", err)
	}

	// 3. Remove participant
	if err := chat.RemoveParticipant(input.ParticipantID); err != nil {
		return nil, fmt.Errorf("failed to remove participant: %w", err)
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
	return &RemoveParticipantOutput{
		Chat: ChatToDTO(chat),
	}, nil
}
