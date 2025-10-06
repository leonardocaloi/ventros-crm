package message

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

type Repository interface {
	Save(ctx context.Context, message *Message) error
	FindByID(ctx context.Context, id uuid.UUID) (*Message, error)
	FindBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*Message, error)
	FindByContact(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*Message, error)
	FindByChannelMessageID(ctx context.Context, channelMessageID string) (*Message, error)
	CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error)
}
