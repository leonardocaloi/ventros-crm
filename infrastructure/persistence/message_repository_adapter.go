package persistence

import (
	"context"

	"github.com/caloi/ventros-crm/internal/workflows/session"
	"github.com/google/uuid"
)

// MessageRepositoryAdapter adapta GormMessageRepository para a interface do workflow
type MessageRepositoryAdapter struct {
	repo *GormMessageRepository
}

// NewMessageRepositoryAdapter cria um novo adapter
func NewMessageRepositoryAdapter(repo *GormMessageRepository) session.MessageRepository {
	return &MessageRepositoryAdapter{repo: repo}
}

// FindBySessionID implementa a interface do workflow
func (a *MessageRepositoryAdapter) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]session.MessageInfo, error) {
	messages, err := a.repo.FindBySessionIDForEnrichment(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Converte para o tipo esperado pelo workflow
	result := make([]session.MessageInfo, len(messages))
	for i, msg := range messages {
		result[i] = session.MessageInfo{
			ID:        msg.ID,
			ChannelID: msg.ChannelID,
			Direction: msg.Direction,
			Timestamp: msg.Timestamp,
		}
	}

	return result, nil
}
