package chat

import (
	domainchat "github.com/caloi/ventros-crm/internal/domain/chat"
)

// ChatToDTO converts a domain Chat to ChatDTO
func ChatToDTO(c *domainchat.Chat) *ChatDTO {
	if c == nil {
		return nil
	}

	participants := make([]ParticipantDTO, len(c.Participants()))
	for i, p := range c.Participants() {
		participants[i] = ParticipantDTO{
			ID:       p.ID,
			Type:     p.Type.String(),
			JoinedAt: p.JoinedAt,
			LeftAt:   p.LeftAt,
			IsAdmin:  p.IsAdmin,
		}
	}

	return &ChatDTO{
		ID:            c.ID(),
		ProjectID:     c.ProjectID(),
		TenantID:      c.TenantID(),
		ChatType:      c.ChatType().String(),
		ExternalID:    c.ExternalID(),
		Subject:       c.Subject(),
		Description:   c.Description(),
		Participants:  participants,
		Status:        c.Status().String(),
		Metadata:      c.Metadata(),
		LastMessageAt: c.LastMessageAt(),
		CreatedAt:     c.CreatedAt(),
		UpdatedAt:     c.UpdatedAt(),
	}
}

// ChatsToDTO converts multiple domain Chats to ChatDTOs
func ChatsToDTO(chats []*domainchat.Chat) []*ChatDTO {
	if chats == nil {
		return []*ChatDTO{}
	}

	dtos := make([]*ChatDTO, len(chats))
	for i, c := range chats {
		dtos[i] = ChatToDTO(c)
	}
	return dtos
}
