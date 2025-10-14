package persistence

import (
	"context"
	"fmt"

	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/crm/channel_type"
	"gorm.io/gorm"
)

// GormChannelTypeRepository implements channel_type.Repository using GORM
type GormChannelTypeRepository struct {
	db *gorm.DB
}

// NewGormChannelTypeRepository creates a new GORM channel type repository
func NewGormChannelTypeRepository(db *gorm.DB) *GormChannelTypeRepository {
	return &GormChannelTypeRepository{db: db}
}

// Save saves a channel type to the database
func (r *GormChannelTypeRepository) Save(ctx context.Context, ct *channel_type.ChannelType) error {
	entity := r.domainToEntity(ct)

	// Use GORM's Save method which handles both create and update
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("failed to save channel type: %w", err)
	}

	return nil
}

// FindByID finds a channel type by ID
func (r *GormChannelTypeRepository) FindByID(ctx context.Context, id int) (*channel_type.ChannelType, error) {
	var entity entities.ChannelTypeEntity

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, channel_type.ErrChannelTypeNotFound
		}
		return nil, fmt.Errorf("failed to find channel type: %w", err)
	}

	return r.entityToDomain(&entity)
}

// FindByName finds a channel type by name
func (r *GormChannelTypeRepository) FindByName(ctx context.Context, name string) (*channel_type.ChannelType, error) {
	var entity entities.ChannelTypeEntity

	err := r.db.WithContext(ctx).Where("name = ?", name).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, channel_type.ErrChannelTypeNotFound
		}
		return nil, fmt.Errorf("failed to find channel type by name: %w", err)
	}

	return r.entityToDomain(&entity)
}

// FindAll finds all channel types
func (r *GormChannelTypeRepository) FindAll(ctx context.Context) ([]*channel_type.ChannelType, error) {
	var entities []entities.ChannelTypeEntity

	err := r.db.WithContext(ctx).Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all channel types: %w", err)
	}

	channelTypes := make([]*channel_type.ChannelType, len(entities))
	for i, entity := range entities {
		ct, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity to domain: %w", err)
		}
		channelTypes[i] = ct
	}

	return channelTypes, nil
}

// FindActive finds all active channel types
func (r *GormChannelTypeRepository) FindActive(ctx context.Context) ([]*channel_type.ChannelType, error) {
	var entities []entities.ChannelTypeEntity

	err := r.db.WithContext(ctx).Where("active = ?", true).Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active channel types: %w", err)
	}

	channelTypes := make([]*channel_type.ChannelType, len(entities))
	for i, entity := range entities {
		ct, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity to domain: %w", err)
		}
		channelTypes[i] = ct
	}

	return channelTypes, nil
}

// Delete deletes a channel type (soft delete)
func (r *GormChannelTypeRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&entities.ChannelTypeEntity{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete channel type: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return channel_type.ErrChannelTypeNotFound
	}

	return nil
}

// domainToEntity converts domain channel type to entity
func (r *GormChannelTypeRepository) domainToEntity(ct *channel_type.ChannelType) *entities.ChannelTypeEntity {
	return &entities.ChannelTypeEntity{
		ID:            ct.ID(),
		Name:          ct.Name(),
		Description:   ct.Description(),
		Provider:      ct.Provider(),
		Configuration: ct.Configuration(),
		Active:        ct.IsActive(),
		CreatedAt:     ct.CreatedAt(),
		UpdatedAt:     ct.UpdatedAt(),
	}
}

// entityToDomain converts entity to domain channel type
func (r *GormChannelTypeRepository) entityToDomain(entity *entities.ChannelTypeEntity) (*channel_type.ChannelType, error) {
	return channel_type.ReconstructChannelType(
		entity.ID,
		entity.Name,
		entity.Description,
		entity.Provider,
		entity.Configuration,
		entity.Active,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}
