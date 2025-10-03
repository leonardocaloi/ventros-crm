package channel_type

import "context"

// Repository define o contrato de persistência para ChannelType.
type Repository interface {
	// Save persiste um ChannelType (insert ou update).
	Save(ctx context.Context, ct *ChannelType) error

	// FindByID busca um ChannelType por ID.
	FindByID(ctx context.Context, id int) (*ChannelType, error)

	// FindByName busca um ChannelType por nome.
	FindByName(ctx context.Context, name string) (*ChannelType, error)

	// FindActive retorna todos os ChannelTypes ativos.
	FindActive(ctx context.Context) ([]*ChannelType, error)

	// FindAll retorna todos os ChannelTypes.
	FindAll(ctx context.Context) ([]*ChannelType, error)
}
