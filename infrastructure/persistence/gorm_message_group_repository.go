package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/crm/message_group"
	"gorm.io/gorm"
)

// GormMessageGroupRepository implementa message_group.Repository usando GORM
type GormMessageGroupRepository struct {
	db *gorm.DB
}

// NewGormMessageGroupRepository cria um novo repositório de grupos de mensagens
func NewGormMessageGroupRepository(db *gorm.DB) message_group.Repository {
	return &GormMessageGroupRepository{db: db}
}

// Save persiste ou atualiza um grupo de mensagens
func (r *GormMessageGroupRepository) Save(ctx context.Context, group *message_group.MessageGroup) error {
	entity := r.domainToEntity(group)

	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// FindByID busca grupo por ID
func (r *GormMessageGroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*message_group.MessageGroup, error) {
	var entity entities.MessageGroupEntity
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, message_group.ErrMessageGroupNotFound
		}
		return nil, err
	}

	return r.entityToDomain(&entity)
}

// FindActiveByContact busca grupo ativo (pending) para um contato em um canal
func (r *GormMessageGroupRepository) FindActiveByContact(ctx context.Context, contactID, channelID uuid.UUID) (*message_group.MessageGroup, error) {
	var entity entities.MessageGroupEntity
	err := r.db.WithContext(ctx).
		Where("contact_id = ? AND channel_id = ? AND status = ?", contactID, channelID, "pending").
		Order("created_at DESC").
		First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Não é erro, apenas não existe grupo ativo
		}
		return nil, err
	}

	return r.entityToDomain(&entity)
}

// FindExpired busca grupos que expiraram e estão pending
func (r *GormMessageGroupRepository) FindExpired(ctx context.Context, limit int) ([]*message_group.MessageGroup, error) {
	var entities []entities.MessageGroupEntity
	err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at <= ?", "pending", time.Now()).
		Order("expires_at ASC").
		Limit(limit).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	groups := make([]*message_group.MessageGroup, len(entities))
	for i, entity := range entities {
		group, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		groups[i] = group
	}

	return groups, nil
}

// FindByStatus busca grupos por status
func (r *GormMessageGroupRepository) FindByStatus(ctx context.Context, status message_group.GroupStatus, limit int) ([]*message_group.MessageGroup, error) {
	var entities []entities.MessageGroupEntity
	err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Order("created_at DESC").
		Limit(limit).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	groups := make([]*message_group.MessageGroup, len(entities))
	for i, entity := range entities {
		group, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		groups[i] = group
	}

	return groups, nil
}

// Delete remove um grupo
func (r *GormMessageGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entities.MessageGroupEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return message_group.ErrMessageGroupNotFound
	}

	return nil
}

// FindBySessionID busca grupos de uma sessão
func (r *GormMessageGroupRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*message_group.MessageGroup, error) {
	var entities []entities.MessageGroupEntity
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	groups := make([]*message_group.MessageGroup, len(entities))
	for i, entity := range entities {
		group, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		groups[i] = group
	}

	return groups, nil
}

// CountByStatus conta grupos por status
func (r *GormMessageGroupRepository) CountByStatus(ctx context.Context, status message_group.GroupStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.MessageGroupEntity{}).
		Where("status = ?", string(status)).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// domainToEntity converte agregado para entidade GORM
func (r *GormMessageGroupRepository) domainToEntity(group *message_group.MessageGroup) *entities.MessageGroupEntity {
	// Converter []uuid.UUID para []string para GORM array
	messageIDs := make([]string, len(group.MessageIDs()))
	for i, id := range group.MessageIDs() {
		messageIDs[i] = id.String()
	}

	return &entities.MessageGroupEntity{
		ID:          group.ID(),
		ContactID:   group.ContactID(),
		ChannelID:   group.ChannelID(),
		SessionID:   group.SessionID(),
		TenantID:    group.TenantID(),
		MessageIDs:  messageIDs,
		Status:      string(group.Status()),
		StartedAt:   group.StartedAt(),
		CompletedAt: group.CompletedAt(),
		ExpiresAt:   group.ExpiresAt(),
		CreatedAt:   group.StartedAt(), // Usar StartedAt como CreatedAt
		UpdatedAt:   time.Now(),
	}
}

// entityToDomain converte entidade GORM para agregado
func (r *GormMessageGroupRepository) entityToDomain(entity *entities.MessageGroupEntity) (*message_group.MessageGroup, error) {
	// Converter []string para []uuid.UUID
	messageIDs := make([]uuid.UUID, len(entity.MessageIDs))
	for i, idStr := range entity.MessageIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		messageIDs[i] = id
	}

	return message_group.ReconstructMessageGroup(
		entity.ID,
		entity.ContactID,
		entity.ChannelID,
		entity.SessionID,
		entity.TenantID,
		messageIDs,
		message_group.GroupStatus(entity.Status),
		entity.StartedAt,
		entity.CompletedAt,
		entity.ExpiresAt,
	), nil
}
