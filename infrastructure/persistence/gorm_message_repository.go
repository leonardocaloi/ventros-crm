package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/message"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormMessageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepository(db *gorm.DB) message.Repository {
	return &GormMessageRepository{db: db}
}

func (r *GormMessageRepository) Save(ctx context.Context, m *message.Message) error {
	entity := r.domainToEntity(m)
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*message.Message, error) {
	var entity entities.MessageEntity
	err := r.db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, message.ErrMessageNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormMessageRepository) FindBySession(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*message.Message, error) {
	var entities []entities.MessageEntity
	query := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("timestamp ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&entities).Error
	if err != nil {
		return nil, err
	}

	messages := make([]*message.Message, len(entities))
	for i, entity := range entities {
		messages[i] = r.entityToDomain(&entity)
	}
	return messages, nil
}

func (r *GormMessageRepository) FindByContact(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*message.Message, error) {
	var entities []entities.MessageEntity
	query := r.db.WithContext(ctx).Where("contact_id = ?", contactID).Order("timestamp DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&entities).Error
	if err != nil {
		return nil, err
	}

	messages := make([]*message.Message, len(entities))
	for i, entity := range entities {
		messages[i] = r.entityToDomain(&entity)
	}
	return messages, nil
}

func (r *GormMessageRepository) FindByChannelMessageID(ctx context.Context, channelMessageID string) (*message.Message, error) {
	var entity entities.MessageEntity
	err := r.db.WithContext(ctx).Where("channel_message_id = ?", channelMessageID).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, message.ErrMessageNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormMessageRepository) CountBySession(ctx context.Context, sessionID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.MessageEntity{}).Where("session_id = ?", sessionID).Count(&count).Error
	return int(count), err
}

// Mappers: Domain → Entity
func (r *GormMessageRepository) domainToEntity(m *message.Message) *entities.MessageEntity {
	entity := &entities.MessageEntity{
		Timestamp:        m.Timestamp(),
		UserID:           m.CustomerID(),
		ProjectID:        m.ProjectID(),
		ChannelTypeID:    m.ChannelTypeID(),
		FromMe:           m.FromMe(),
		ChannelID:        m.ChannelID(),
		ContactID:        m.ContactID(),
		SessionID:        m.SessionID(),
		ContentType:      m.ContentType().String(),
		Text:             m.Text(),
		MediaURL:         m.MediaURL(),
		ChannelMessageID: m.ChannelMessageID(),
		ReplyToID:        m.ReplyToID(),
		Status:           m.Status().String(),
		Language:         m.Language(),
		AgentID:          m.AgentID(),
		Metadata:         m.Metadata(),
		DeliveredAt:      m.DeliveredAt(),
		ReadAt:           m.ReadAt(),
		CreatedAt:        m.Timestamp(),
		UpdatedAt:        time.Now(),
	}

	return entity
}

// Mappers: Entity → Domain
func (r *GormMessageRepository) entityToDomain(entity *entities.MessageEntity) *message.Message {
	contentType, _ := message.ParseContentType(entity.ContentType)
	status, _ := message.ParseStatus(entity.Status)

	return message.ReconstructMessage(
		entity.ID,
		entity.Timestamp,
		entity.UserID,
		entity.ProjectID,
		entity.ChannelTypeID,
		entity.FromMe,
		entity.ChannelID,
		entity.ContactID,
		entity.SessionID,
		contentType,
		entity.Text,
		entity.MediaURL,
		entity.MediaMimetype,
		entity.ChannelMessageID,
		entity.ReplyToID,
		status,
		entity.Language,
		entity.AgentID,
		entity.Metadata,
		entity.DeliveredAt,
		entity.ReadAt,
	)
}
