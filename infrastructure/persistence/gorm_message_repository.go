package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/application/shared"
	"github.com/ventros/crm/internal/domain/core/saga"
	"github.com/ventros/crm/internal/domain/crm/message"
	"gorm.io/gorm"
)

type GormMessageRepository struct {
	db *gorm.DB
}

func NewGormMessageRepository(db *gorm.DB) message.Repository {
	return &GormMessageRepository{db: db}
}

// getDB extracts transaction from context if present, otherwise returns default connection
func (r *GormMessageRepository) getDB(ctx context.Context) *gorm.DB {
	// Try to extract transaction from context
	if tx := shared.TransactionFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	// If no transaction, use default connection
	return r.db.WithContext(ctx)
}

func (r *GormMessageRepository) Save(ctx context.Context, m *message.Message) error {
	entity := r.domainToEntity(ctx, m)
	db := r.getDB(ctx) // ✅ Use transaction context
	return db.Save(entity).Error
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

func (r *GormMessageRepository) FindByChannelAndMessageID(ctx context.Context, channelID uuid.UUID, channelMessageID string) (*message.Message, error) {
	var entity entities.MessageEntity
	err := r.db.WithContext(ctx).Where("channel_id = ? AND channel_message_id = ?", channelID, channelMessageID).First(&entity).Error
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

func (r *GormMessageRepository) FindByTenantWithFilters(ctx context.Context, filters message.MessageFilters) ([]*message.Message, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.MessageEntity{})

	// Apply tenant filter (required)
	query = query.Where("tenant_id = ?", filters.TenantID)

	// Apply optional filters
	if filters.ContactID != nil {
		query = query.Where("contact_id = ?", *filters.ContactID)
	}
	if filters.SessionID != nil {
		query = query.Where("session_id = ?", *filters.SessionID)
	}
	if filters.ChannelID != nil {
		query = query.Where("channel_id = ?", *filters.ChannelID)
	}
	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}
	if filters.ChannelTypeID != nil {
		query = query.Where("channel_type_id = ?", *filters.ChannelTypeID)
	}
	if filters.FromMe != nil {
		query = query.Where("from_me = ?", *filters.FromMe)
	}
	if filters.ContentType != nil {
		query = query.Where("content_type = ?", *filters.ContentType)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.AgentID != nil {
		query = query.Where("agent_id = ?", *filters.AgentID)
	}
	if filters.TimestampAfter != nil {
		query = query.Where("timestamp >= ?", *filters.TimestampAfter)
	}
	if filters.TimestampBefore != nil {
		query = query.Where("timestamp <= ?", *filters.TimestampBefore)
	}
	if filters.HasMedia != nil {
		if *filters.HasMedia {
			query = query.Where("media_url IS NOT NULL")
		} else {
			query = query.Where("media_url IS NULL")
		}
	}

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "timestamp"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Execute query
	var messageEntities []entities.MessageEntity
	if err := query.Find(&messageEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	messages := make([]*message.Message, len(messageEntities))
	for i, entity := range messageEntities {
		messages[i] = r.entityToDomain(&entity)
	}

	return messages, total, nil
}

func (r *GormMessageRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*message.Message, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.MessageEntity{})

	// Apply tenant filter
	query = query.Where("tenant_id = ?", tenantID)

	// Text search in message text content
	searchPattern := "%" + searchText + "%"
	query = query.Where("text ILIKE ?", searchPattern)

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	query = query.Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	var messageEntities []entities.MessageEntity
	if err := query.Find(&messageEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	messages := make([]*message.Message, len(messageEntities))
	for i, entity := range messageEntities {
		messages[i] = r.entityToDomain(&entity)
	}

	return messages, total, nil
}

// FindBySessionIDForEnrichment retorna informações simplificadas das mensagens para enrichment de eventos
func (r *GormMessageRepository) FindBySessionIDForEnrichment(ctx context.Context, sessionID uuid.UUID) ([]MessageInfoForEnrichment, error) {
	type Result struct {
		ID        uuid.UUID
		ChannelID *uuid.UUID
		FromMe    bool
		Timestamp time.Time
	}

	var results []Result
	err := r.db.WithContext(ctx).
		Model(&entities.MessageEntity{}).
		Select("id, channel_id, from_me, timestamp").
		Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	messages := make([]MessageInfoForEnrichment, len(results))
	for i, r := range results {
		direction := "inbound"
		if r.FromMe {
			direction = "outbound"
		}
		messages[i] = MessageInfoForEnrichment{
			ID:        r.ID,
			ChannelID: r.ChannelID,
			Direction: direction,
			Timestamp: r.Timestamp,
		}
	}

	return messages, nil
}

// MessageInfoForEnrichment informações simplificadas de mensagem para enrichment
type MessageInfoForEnrichment struct {
	ID        uuid.UUID
	ChannelID *uuid.UUID
	Direction string
	Timestamp time.Time
}

// UpdateSessionIDForSession updates all messages from oldSessionID to newSessionID
// Used for session consolidation during history import post-processing
// Returns the number of messages updated
func (r *GormMessageRepository) UpdateSessionIDForSession(ctx context.Context, oldSessionID, newSessionID uuid.UUID) (int64, error) {
	db := r.getDB(ctx) // Use transaction context if present

	result := db.Model(&entities.MessageEntity{}).
		Where("session_id = ?", oldSessionID).
		Update("session_id", newSessionID)

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// Mappers: Domain → Entity
func (r *GormMessageRepository) domainToEntity(ctx context.Context, m *message.Message) *entities.MessageEntity {
	// ✅ Extract tenant_id from saga context
	tenantID, _ := saga.GetTenantID(ctx)

	entity := &entities.MessageEntity{
		TenantID:         tenantID, // ✅ Extracted from saga context
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
		Source:           m.Source().String(),
		Metadata:         m.Metadata(),
		Mentions:         m.Mentions(),
		DeliveredAt:      m.DeliveredAt(),
		ReadAt:           m.ReadAt(),
		PlayedAt:         m.PlayedAt(),
		CreatedAt:        m.Timestamp(),
		UpdatedAt:        time.Now(),
	}

	return entity
}

// Mappers: Entity → Domain
func (r *GormMessageRepository) entityToDomain(entity *entities.MessageEntity) *message.Message {
	contentType, _ := message.ParseContentType(entity.ContentType)
	status, _ := message.ParseStatus(entity.Status)

	// Parse source with default fallback
	source := message.Source(entity.Source)
	if !source.IsValid() {
		source = message.SourceManual // Default to manual if invalid
	}

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
		source,
		entity.Metadata,
		entity.DeliveredAt,
		entity.ReadAt,
		entity.PlayedAt,
		entity.Mentions,
	)
}
