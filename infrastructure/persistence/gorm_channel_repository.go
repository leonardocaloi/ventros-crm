package persistence

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormChannelRepository struct {
	db *gorm.DB
}

func NewGormChannelRepository(db *gorm.DB) channel.Repository {
	return &GormChannelRepository{db: db}
}

func (r *GormChannelRepository) Create(ch *channel.Channel) error {
	entity := r.toEntity(ch)
	if err := r.db.Create(entity).Error; err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	return nil
}

func (r *GormChannelRepository) GetByID(id uuid.UUID) (*channel.Channel, error) {
	var entity entities.ChannelEntity
	if err := r.db.First(&entity, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return r.toDomain(&entity), nil
}

func (r *GormChannelRepository) GetByUserID(userID uuid.UUID) ([]*channel.Channel, error) {
	var entities []entities.ChannelEntity
	if err := r.db.Where("user_id = ?", userID).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels by user: %w", err)
	}

	channels := make([]*channel.Channel, len(entities))
	for i, entity := range entities {
		channels[i] = r.toDomain(&entity)
	}
	return channels, nil
}

func (r *GormChannelRepository) GetByProjectID(projectID uuid.UUID) ([]*channel.Channel, error) {
	var entities []entities.ChannelEntity
	if err := r.db.Where("project_id = ?", projectID).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get channels by project: %w", err)
	}

	channels := make([]*channel.Channel, len(entities))
	for i, entity := range entities {
		channels[i] = r.toDomain(&entity)
	}
	return channels, nil
}

func (r *GormChannelRepository) GetByExternalID(externalID string) (*channel.Channel, error) {
	var entity entities.ChannelEntity
	if err := r.db.Where("external_id = ?", externalID).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return r.toDomain(&entity), nil
}

func (r *GormChannelRepository) GetByWebhookID(webhookID string) (*channel.Channel, error) {
	var entity entities.ChannelEntity
	if err := r.db.Where("webhook_id = ?", webhookID).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("channel not found")
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}
	return r.toDomain(&entity), nil
}

func (r *GormChannelRepository) GetActiveWAHAChannels() ([]*channel.Channel, error) {
	var entities []entities.ChannelEntity
	if err := r.db.Where("type = ? AND status = ?", "waha", "active").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get active WAHA channels: %w", err)
	}

	channels := make([]*channel.Channel, len(entities))
	for i, entity := range entities {
		channels[i] = r.toDomain(&entity)
	}
	return channels, nil
}

func (r *GormChannelRepository) Update(ch *channel.Channel) error {
	entity := r.toEntity(ch)
	if err := r.db.Save(entity).Error; err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}
	return nil
}

func (r *GormChannelRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&entities.ChannelEntity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	return nil
}

func (r *GormChannelRepository) toEntity(ch *channel.Channel) *entities.ChannelEntity {
	// Convert map[string]interface{} to datatypes.JSON
	var config datatypes.JSON
	if ch.Config != nil {
		if configBytes, err := json.Marshal(ch.Config); err == nil {
			config = configBytes
		}
	}

	return &entities.ChannelEntity{
		ID:         ch.ID,
		UserID:     ch.UserID,
		ProjectID:  ch.ProjectID,
		TenantID:   ch.TenantID,
		Name:       ch.Name,
		Type:       entities.ChannelType(ch.Type),
		Status:     entities.ChannelStatus(ch.Status),
		ExternalID: ch.ExternalID,
		Config:     config,

		// Webhook fields
		WebhookID:           ch.WebhookID,
		WebhookURL:          ch.WebhookURL,
		WebhookConfiguredAt: ch.WebhookConfiguredAt,
		WebhookActive:       ch.WebhookActive,

		// AI Features
		AIEnabled:       ch.AIEnabled,
		AIAgentsEnabled: ch.AIAgentsEnabled,
		AllowGroups:     ch.AllowGroups,
		TrackingEnabled: ch.TrackingEnabled,

		// Message Debouncer
		DebounceTimeoutMs: ch.DebounceTimeoutMs,

		// Statistics
		MessagesReceived: ch.MessagesReceived,
		MessagesSent:     ch.MessagesSent,
		LastMessageAt:    ch.LastMessageAt,
		LastErrorAt:      ch.LastErrorAt,
		LastError:        ch.LastError,

		CreatedAt: ch.CreatedAt,
		UpdatedAt: ch.UpdatedAt,
	}
}

func (r *GormChannelRepository) toDomain(entity *entities.ChannelEntity) *channel.Channel {
	// Convert datatypes.JSON to map[string]interface{}
	var config map[string]interface{}
	if len(entity.Config) > 0 {
		if err := json.Unmarshal(entity.Config, &config); err != nil {
			config = make(map[string]interface{})
		}
	} else {
		config = make(map[string]interface{})
	}

	return &channel.Channel{
		ID:         entity.ID,
		UserID:     entity.UserID,
		ProjectID:  entity.ProjectID,
		TenantID:   entity.TenantID,
		Name:       entity.Name,
		Type:       channel.ChannelType(entity.Type),
		ExternalID: entity.ExternalID,
		Status:     channel.ChannelStatus(entity.Status),
		Config:     config,

		// Webhook fields
		WebhookID:           entity.WebhookID,
		WebhookURL:          entity.WebhookURL,
		WebhookConfiguredAt: entity.WebhookConfiguredAt,
		WebhookActive:       entity.WebhookActive,

		// AI Features
		AIEnabled:       entity.AIEnabled,
		AIAgentsEnabled: entity.AIAgentsEnabled,
		AllowGroups:     entity.AllowGroups,
		TrackingEnabled: entity.TrackingEnabled,

		// Message Debouncer
		DebounceTimeoutMs: entity.DebounceTimeoutMs,

		// Statistics
		MessagesReceived: entity.MessagesReceived,
		MessagesSent:     entity.MessagesSent,
		LastMessageAt:    entity.LastMessageAt,
		LastErrorAt:      entity.LastErrorAt,
		LastError:        entity.LastError,

		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
