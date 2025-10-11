package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/chat"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormChatRepository struct {
	db *gorm.DB
}

func NewGormChatRepository(db *gorm.DB) chat.Repository {
	return &GormChatRepository{db: db}
}

// Create saves a new chat
func (r *GormChatRepository) Create(ctx context.Context, c *chat.Chat) error {
	entity := r.domainToEntity(c)
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID finds a chat by ID
func (r *GormChatRepository) FindByID(ctx context.Context, id uuid.UUID) (*chat.Chat, error) {
	var entity entities.ChatEntity
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, chat.ErrChatNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity)
}

// FindByExternalID finds a chat by external ID (WhatsApp group ID, etc)
func (r *GormChatRepository) FindByExternalID(ctx context.Context, externalID string) (*chat.Chat, error) {
	var entity entities.ChatEntity
	err := r.db.WithContext(ctx).First(&entity, "external_id = ?", externalID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, chat.ErrChatNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity)
}

// FindByProject finds all chats for a project
func (r *GormChatRepository) FindByProject(ctx context.Context, projectID uuid.UUID) ([]*chat.Chat, error) {
	var entities []entities.ChatEntity
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	chats := make([]*chat.Chat, len(entities))
	for i, entity := range entities {
		c, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		chats[i] = c
	}
	return chats, nil
}

// FindByTenant finds all chats for a tenant
func (r *GormChatRepository) FindByTenant(ctx context.Context, tenantID string) ([]*chat.Chat, error) {
	var entities []entities.ChatEntity
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	chats := make([]*chat.Chat, len(entities))
	for i, entity := range entities {
		c, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		chats[i] = c
	}
	return chats, nil
}

// FindByContact finds all chats where contact is a participant
func (r *GormChatRepository) FindByContact(ctx context.Context, contactID uuid.UUID) ([]*chat.Chat, error) {
	var entities []entities.ChatEntity
	// Use JSONB query to find chats where contact is a participant
	err := r.db.WithContext(ctx).
		Where("participants @> ?", datatypes.JSON(`[{"id":"`+contactID.String()+`"}]`)).
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	chats := make([]*chat.Chat, len(entities))
	for i, entity := range entities {
		c, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		chats[i] = c
	}
	return chats, nil
}

// FindActiveByProject finds all active chats (not archived, not closed) for a project
func (r *GormChatRepository) FindActiveByProject(ctx context.Context, projectID uuid.UUID) ([]*chat.Chat, error) {
	var entities []entities.ChatEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND status = ?", projectID, "active").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	chats := make([]*chat.Chat, len(entities))
	for i, entity := range entities {
		c, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		chats[i] = c
	}
	return chats, nil
}

// FindIndividualByContact finds an individual chat for a contact in a project
func (r *GormChatRepository) FindIndividualByContact(ctx context.Context, contactID uuid.UUID, projectID uuid.UUID) (*chat.Chat, error) {
	var entity entities.ChatEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND chat_type = ? AND participants @> ?",
			projectID,
			"individual",
			datatypes.JSON(`[{"id":"`+contactID.String()+`"}]`),
		).
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, chat.ErrChatNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity)
}

// Update updates an existing chat
func (r *GormChatRepository) Update(ctx context.Context, c *chat.Chat) error {
	entity := r.domainToEntity(c)
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete soft deletes a chat
func (r *GormChatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.ChatEntity{}, "id = ?", id).Error
}

// SearchBySubject searches chats by subject
func (r *GormChatRepository) SearchBySubject(ctx context.Context, tenantID string, subject string) ([]*chat.Chat, error) {
	var entities []entities.ChatEntity
	searchPattern := "%" + subject + "%"
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND subject ILIKE ?", tenantID, searchPattern).
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	chats := make([]*chat.Chat, len(entities))
	for i, entity := range entities {
		c, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		chats[i] = c
	}
	return chats, nil
}

// Mappers: Domain → Entity

func (r *GormChatRepository) domainToEntity(c *chat.Chat) *entities.ChatEntity {
	entity := &entities.ChatEntity{
		ID:            c.ID(),
		ProjectID:     c.ProjectID(),
		TenantID:      c.TenantID(),
		ChatType:      c.ChatType().String(),
		ExternalID:    c.ExternalID(),
		Subject:       c.Subject(),
		Description:   c.Description(),
		Participants:  participantsToJSON(c.Participants()),
		Status:        c.Status().String(),
		Metadata:      metadataToJSON(c.Metadata()),
		LastMessageAt: c.LastMessageAt(),
		CreatedAt:     c.CreatedAt(),
		UpdatedAt:     c.UpdatedAt(),
	}

	return entity
}

// Mappers: Entity → Domain

func (r *GormChatRepository) entityToDomain(entity *entities.ChatEntity) (*chat.Chat, error) {
	// Parse chat type
	chatType, err := chat.ParseChatType(entity.ChatType)
	if err != nil {
		return nil, err
	}

	// Parse status
	status, err := chat.ParseChatStatus(entity.Status)
	if err != nil {
		return nil, err
	}

	// Parse participants
	participants, err := jsonToParticipants(entity.Participants)
	if err != nil {
		return nil, err
	}

	// Parse metadata
	metadata := jsonToMetadata(entity.Metadata)

	return chat.ReconstructChat(
		entity.ID,
		entity.ProjectID,
		entity.TenantID,
		chatType,
		entity.ExternalID,
		entity.Subject,
		entity.Description,
		participants,
		status,
		metadata,
		entity.LastMessageAt,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}

// Helper functions for JSON conversion

func participantsToJSON(participants []chat.Participant) datatypes.JSON {
	if participants == nil {
		participants = []chat.Participant{}
	}

	jsonParticipants := make([]entities.ParticipantJSON, len(participants))
	for i, p := range participants {
		jsonParticipants[i] = entities.ParticipantJSON{
			ID:       p.ID,
			Type:     p.Type.String(),
			JoinedAt: p.JoinedAt,
			LeftAt:   p.LeftAt,
			IsAdmin:  p.IsAdmin,
		}
	}

	data, err := json.Marshal(jsonParticipants)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return datatypes.JSON(data)
}

func jsonToParticipants(j datatypes.JSON) ([]chat.Participant, error) {
	if len(j) == 0 {
		return []chat.Participant{}, nil
	}

	var jsonParticipants []entities.ParticipantJSON
	if err := json.Unmarshal(j, &jsonParticipants); err != nil {
		return nil, err
	}

	participants := make([]chat.Participant, len(jsonParticipants))
	for i, jp := range jsonParticipants {
		participantType, err := chat.ParseParticipantType(jp.Type)
		if err != nil {
			return nil, err
		}

		participants[i] = chat.Participant{
			ID:       jp.ID,
			Type:     participantType,
			JoinedAt: jp.JoinedAt,
			LeftAt:   jp.LeftAt,
			IsAdmin:  jp.IsAdmin,
		}
	}

	return participants, nil
}

func metadataToJSON(m map[string]interface{}) datatypes.JSON {
	if m == nil {
		return datatypes.JSON([]byte("{}"))
	}
	data, err := json.Marshal(m)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(data)
}

func jsonToMetadata(j datatypes.JSON) map[string]interface{} {
	if len(j) == 0 {
		return make(map[string]interface{})
	}
	var m map[string]interface{}
	if err := json.Unmarshal(j, &m); err != nil {
		return make(map[string]interface{})
	}
	return m
}
