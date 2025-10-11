package channel_type

import "context"

type Repository interface {
	Save(ctx context.Context, ct *ChannelType) error

	FindByID(ctx context.Context, id int) (*ChannelType, error)

	FindByName(ctx context.Context, name string) (*ChannelType, error)

	FindActive(ctx context.Context) ([]*ChannelType, error)

	FindAll(ctx context.Context) ([]*ChannelType, error)
}
