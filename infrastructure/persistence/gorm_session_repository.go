package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/session"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormSessionRepository struct {
	db *gorm.DB
}

func NewGormSessionRepository(db *gorm.DB) session.Repository {
	return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Save(ctx context.Context, s *session.Session) error {
	entity := r.domainToEntity(s)

	// Check if exists
	var existing entities.SessionEntity
	err := r.db.WithContext(ctx).Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update with optimistic locking
		result := r.db.WithContext(ctx).Model(&entities.SessionEntity{}).
			Where("id = ? AND version = ?", entity.ID, existing.Version).
			Updates(map[string]interface{}{
				"version":                    existing.Version + 1, // Increment version
				"contact_id":                 entity.ContactID,
				"tenant_id":                  entity.TenantID,
				"pipeline_id":                entity.PipelineID,
				"channel_type_id":            entity.ChannelTypeID,
				"started_at":                 entity.StartedAt,
				"ended_at":                   entity.EndedAt,
				"status":                     entity.Status,
				"end_reason":                 entity.EndReason,
				"timeout_duration":           entity.TimeoutDuration,
				"last_activity_at":           entity.LastActivityAt,
				"message_count":              entity.MessageCount,
				"messages_from_contact":      entity.MessagesFromContact,
				"messages_from_agent":        entity.MessagesFromAgent,
				"duration_seconds":           entity.DurationSeconds,
				"first_contact_message_at":   entity.FirstContactMessageAt,
				"first_agent_response_at":    entity.FirstAgentResponseAt,
				"agent_response_time_seconds": entity.AgentResponseTimeSeconds,
				"contact_wait_time_seconds":  entity.ContactWaitTimeSeconds,
				"agent_ids":                  entity.AgentIDs,
				"agent_transfers":            entity.AgentTransfers,
				"summary":                    entity.Summary,
				"sentiment":                  entity.Sentiment,
				"sentiment_score":            entity.SentimentScore,
				"topics":                     entity.Topics,
				"next_steps":                 entity.NextSteps,
				"key_entities":               entity.KeyEntities,
				"resolved":                   entity.Resolved,
				"escalated":                  entity.Escalated,
				"converted":                  entity.Converted,
				"outcome_tags":               entity.OutcomeTags,
				"updated_at":                 entity.UpdatedAt,
			})

		if result.Error != nil {
			return result.Error
		}

		// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
		if result.RowsAffected == 0 {
			return shared.NewOptimisticLockError(
				"Session",
				entity.ID.String(),
				existing.Version,
				entity.Version,
			)
		}

		return nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return r.db.WithContext(ctx).Create(entity).Error
	}

	return err
}

func (r *GormSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	var entity entities.SessionEntity
	err := r.db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, session.ErrSessionNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormSessionRepository) FindActiveByContact(ctx context.Context, contactID uuid.UUID, channelTypeID *int) (*session.Session, error) {
	var entity entities.SessionEntity
	query := r.db.WithContext(ctx).Where("contact_id = ? AND status = 'active'", contactID)

	if channelTypeID != nil {
		query = query.Where("channel_type_id = ?", *channelTypeID)
	}

	err := query.First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, session.ErrSessionNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormSessionRepository) FindInactiveSessions(ctx context.Context, tenantID string) ([]*session.Session, error) {
	var entities []entities.SessionEntity
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND status = 'active' AND last_activity_at < NOW() - INTERVAL '30 minutes'", tenantID).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	sessions := make([]*session.Session, len(entities))
	for i, entity := range entities {
		sessions[i] = r.entityToDomain(&entity)
	}
	return sessions, nil
}

func (r *GormSessionRepository) FindSessionsRequiringSummary(ctx context.Context, tenantID string, limit int) ([]*session.Session, error) {
	var entities []entities.SessionEntity
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND status = 'ended' AND summary IS NULL AND message_count >= 3", tenantID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&entities).Error
	if err != nil {
		return nil, err
	}

	sessions := make([]*session.Session, len(entities))
	for i, entity := range entities {
		sessions[i] = r.entityToDomain(&entity)
	}
	return sessions, nil
}

func (r *GormSessionRepository) CountActiveByTenant(ctx context.Context, tenantID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.SessionEntity{}).
		Where("tenant_id = ? AND status = 'active'", tenantID).Count(&count).Error
	return int(count), err
}

func (r *GormSessionRepository) FindActiveBeforeTime(ctx context.Context, cutoffTime time.Time) ([]*session.Session, error) {
	var entities []entities.SessionEntity
	err := r.db.WithContext(ctx).Where("status = 'active' AND last_activity_at < ?", cutoffTime).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	sessions := make([]*session.Session, len(entities))
	for i, entity := range entities {
		sessions[i] = r.entityToDomain(&entity)
	}
	return sessions, nil
}

func (r *GormSessionRepository) FindByTenantWithFilters(ctx context.Context, filters session.SessionFilters) ([]*session.Session, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.SessionEntity{})

	// Apply tenant filter (required)
	query = query.Where("tenant_id = ?", filters.TenantID)

	// Apply optional filters
	if filters.ContactID != nil {
		query = query.Where("contact_id = ?", *filters.ContactID)
	}
	if filters.PipelineID != nil {
		query = query.Where("pipeline_id = ?", *filters.PipelineID)
	}
	if filters.ChannelTypeID != nil {
		query = query.Where("channel_type_id = ?", *filters.ChannelTypeID)
	}
	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}
	if filters.Resolved != nil {
		query = query.Where("resolved = ?", *filters.Resolved)
	}
	if filters.Escalated != nil {
		query = query.Where("escalated = ?", *filters.Escalated)
	}
	if filters.Converted != nil {
		query = query.Where("converted = ?", *filters.Converted)
	}
	if filters.Sentiment != nil {
		query = query.Where("sentiment = ?", *filters.Sentiment)
	}
	if filters.StartedAfter != nil {
		query = query.Where("started_at >= ?", *filters.StartedAfter)
	}
	if filters.StartedBefore != nil {
		query = query.Where("started_at <= ?", *filters.StartedBefore)
	}
	if filters.EndedAfter != nil {
		query = query.Where("ended_at >= ?", *filters.EndedAfter)
	}
	if filters.EndedBefore != nil {
		query = query.Where("ended_at <= ?", *filters.EndedBefore)
	}
	if filters.MinMessages != nil {
		query = query.Where("message_count >= ?", *filters.MinMessages)
	}
	if filters.MaxMessages != nil {
		query = query.Where("message_count <= ?", *filters.MaxMessages)
	}

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "started_at"
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
	var sessionEntities []entities.SessionEntity
	if err := query.Find(&sessionEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	sessions := make([]*session.Session, len(sessionEntities))
	for i, entity := range sessionEntities {
		sessions[i] = r.entityToDomain(&entity)
	}

	return sessions, total, nil
}

func (r *GormSessionRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*session.Session, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.SessionEntity{})

	// Apply tenant filter
	query = query.Where("tenant_id = ?", tenantID)

	// Text search across summary, topics, next_steps, outcome_tags
	searchPattern := "%" + searchText + "%"
	query = query.Where(
		r.db.Where("summary ILIKE ?", searchPattern).
			Or("topics::text ILIKE ?", searchPattern).
			Or("next_steps::text ILIKE ?", searchPattern).
			Or("outcome_tags::text ILIKE ?", searchPattern).
			Or("end_reason ILIKE ?", searchPattern),
	)

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	query = query.Order("started_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	var sessionEntities []entities.SessionEntity
	if err := query.Find(&sessionEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	sessions := make([]*session.Session, len(sessionEntities))
	for i, entity := range sessionEntities {
		sessions[i] = r.entityToDomain(&entity)
	}

	return sessions, total, nil
}

// Mappers: Domain → Entity
func (r *GormSessionRepository) domainToEntity(s *session.Session) *entities.SessionEntity {
	entity := &entities.SessionEntity{
		ID:                  s.ID(),
		Version:             s.Version(),
		ContactID:           s.ContactID(),
		TenantID:            s.TenantID(),
		PipelineID:          s.PipelineID(), // ✅ Pipeline ID
		ChannelTypeID:       s.ChannelTypeID(),
		StartedAt:           s.StartedAt(),
		EndedAt:             s.EndedAt(),
		Status:              s.Status().String(),
		TimeoutDuration:     int64(s.TimeoutDuration()),
		LastActivityAt:      s.LastActivityAt(),
		MessageCount:        s.MessageCount(),
		MessagesFromContact: s.MessagesFromContact(),
		MessagesFromAgent:   s.MessagesFromAgent(),
		DurationSeconds:     s.DurationSeconds(),
		AgentIDs:            s.AgentIDs(),
		AgentTransfers:      s.AgentTransfers(),
		Summary:             s.Summary(),
		Topics:              s.Topics(),
		NextSteps:           s.NextSteps(),
		KeyEntities:         keyEntitiesToJSON(s.KeyEntities()),
		Resolved:            s.IsResolved(),
		Escalated:           s.IsEscalated(),
		Converted:           s.IsConverted(),
		OutcomeTags:         s.OutcomeTags(),
		CreatedAt:           s.StartedAt(), // Usar StartedAt como CreatedAt
		UpdatedAt:           time.Now(),
	}

	if s.EndReason() != nil {
		reason := s.EndReason().String()
		entity.EndReason = &reason
	}

	if s.Sentiment() != nil {
		sentiment := s.Sentiment().String()
		entity.Sentiment = &sentiment
	}

	if s.SentimentScore() != nil {
		entity.SentimentScore = s.SentimentScore()
	}

	return entity
}

// Mappers: Entity → Domain
func (r *GormSessionRepository) entityToDomain(entity *entities.SessionEntity) *session.Session {
	var endReason *session.EndReason
	if entity.EndReason != nil {
		if parsed, err := session.ParseEndReason(*entity.EndReason); err == nil {
			endReason = &parsed
		}
	}

	var sentiment *session.Sentiment
	if entity.Sentiment != nil {
		if parsed, err := session.ParseSentiment(*entity.Sentiment); err == nil {
			sentiment = &parsed
		}
	}

	status, _ := session.ParseStatus(entity.Status)

	return session.ReconstructSession(
		entity.ID,
		entity.Version,
		entity.ContactID,
		entity.TenantID,
		entity.ChannelTypeID,
		entity.PipelineID, // ✅ Adicionado pipeline_id
		entity.StartedAt,
		entity.EndedAt,
		status,
		endReason,
		time.Duration(entity.TimeoutDuration),
		entity.LastActivityAt,
		entity.MessageCount,
		entity.MessagesFromContact,
		entity.MessagesFromAgent,
		entity.DurationSeconds,
		entity.FirstContactMessageAt,
		entity.FirstAgentResponseAt,
		entity.AgentResponseTimeSeconds,
		entity.ContactWaitTimeSeconds,
		entity.AgentIDs,
		entity.AgentTransfers,
		entity.Summary,
		sentiment,
		entity.SentimentScore,
		entity.Topics,
		entity.NextSteps,
		jsonToKeyEntities(entity.KeyEntities),
		entity.Resolved,
		entity.Escalated,
		entity.Converted,
		entity.OutcomeTags,
	)
}

// Helper functions for JSON conversion
func keyEntitiesToJSON(m map[string]interface{}) datatypes.JSON {
	if m == nil {
		return datatypes.JSON([]byte("{}"))
	}
	data, err := json.Marshal(m)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(data)
}

func jsonToKeyEntities(j datatypes.JSON) map[string]interface{} {
	if len(j) == 0 {
		return make(map[string]interface{})
	}
	var m map[string]interface{}
	if err := json.Unmarshal(j, &m); err != nil {
		return make(map[string]interface{})
	}
	return m
}
