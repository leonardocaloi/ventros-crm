package message_group

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMessageGroupNotFound      = errors.New("message group not found")
	ErrMessageGroupAlreadyExists = errors.New("message group already exists")
)

// Repository define operações de persistência para MessageGroup
type Repository interface {
	// Save persiste ou atualiza um grupo de mensagens
	Save(ctx context.Context, group *MessageGroup) error

	// FindByID busca grupo por ID
	FindByID(ctx context.Context, id uuid.UUID) (*MessageGroup, error)

	// FindActiveByContact busca grupo ativo (pending) para um contato em um canal
	// Retorna nil se não encontrar (sem erro)
	FindActiveByContact(ctx context.Context, contactID, channelID uuid.UUID) (*MessageGroup, error)

	// FindExpired busca grupos que expiraram e estão pending
	FindExpired(ctx context.Context, limit int) ([]*MessageGroup, error)

	// FindByStatus busca grupos por status
	FindByStatus(ctx context.Context, status GroupStatus, limit int) ([]*MessageGroup, error)

	// Delete remove um grupo (após processamento completo)
	Delete(ctx context.Context, id uuid.UUID) error

	// FindBySessionID busca grupos de uma sessão
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*MessageGroup, error)

	// CountByStatus conta grupos por status
	CountByStatus(ctx context.Context, status GroupStatus) (int64, error)
}
