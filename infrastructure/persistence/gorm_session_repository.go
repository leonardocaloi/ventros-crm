package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/application/shared"
	domainShared "github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/session"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GormSessionRepository struct {
	db *gorm.DB
}

func NewGormSessionRepository(db *gorm.DB) session.Repository {
	return &GormSessionRepository{db: db}
}

// getDB extracts transaction from context if present, otherwise returns default connection
func (r *GormSessionRepository) getDB(ctx context.Context) *gorm.DB {
	// Try to extract transaction from context
	if tx := shared.TransactionFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	// If no transaction, use default connection
	return r.db.WithContext(ctx)
}

func (r *GormSessionRepository) Save(ctx context.Context, s *session.Session) error {
	entity := r.domainToEntity(s)
	db := r.getDB(ctx) // ✅ Use transaction context

	// Check if exists
	var existing entities.SessionEntity
	err := db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update with optimistic locking
		result := db.Model(&entities.SessionEntity{}).
			Where("id = ? AND version = ?", entity.ID, existing.Version).
			Updates(map[string]interface{}{
				"version":                     existing.Version + 1, // Increment version
				"contact_id":                  entity.ContactID,
				"tenant_id":                   entity.TenantID,
				"pipeline_id":                 entity.PipelineID,
				"channel_type_id":             entity.ChannelTypeID,
				"started_at":                  entity.StartedAt,
				"ended_at":                    entity.EndedAt,
				"status":                      entity.Status,
				"end_reason":                  entity.EndReason,
				"timeout_duration":            entity.TimeoutDuration,
				"last_activity_at":            entity.LastActivityAt,
				"message_count":               entity.MessageCount,
				"messages_from_contact":       entity.MessagesFromContact,
				"messages_from_agent":         entity.MessagesFromAgent,
				"duration_seconds":            entity.DurationSeconds,
				"first_contact_message_at":    entity.FirstContactMessageAt,
				"first_agent_response_at":     entity.FirstAgentResponseAt,
				"agent_response_time_seconds": entity.AgentResponseTimeSeconds,
				"contact_wait_time_seconds":   entity.ContactWaitTimeSeconds,
				"agent_ids":                   entity.AgentIDs,
				"agent_transfers":             entity.AgentTransfers,
				"summary":                     entity.Summary,
				"sentiment":                   entity.Sentiment,
				"sentiment_score":             entity.SentimentScore,
				"topics":                      entity.Topics,
				"next_steps":                  entity.NextSteps,
				"key_entities":                entity.KeyEntities,
				"resolved":                    entity.Resolved,
				"escalated":                   entity.Escalated,
				"converted":                   entity.Converted,
				"outcome_tags":                entity.OutcomeTags,
				"updated_at":                  entity.UpdatedAt,
			})

		if result.Error != nil {
			return result.Error
		}

		// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
		if result.RowsAffected == 0 {
			return domainShared.NewOptimisticLockError(
				"Session",
				entity.ID.String(),
				existing.Version,
				entity.Version,
			)
		}

		return nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert - use transaction context
		return db.Create(entity).Error
	}

	return err
}

func (r *GormSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	var entity entities.SessionEntity
	// ✅ READ COMMITTED: Use r.db to see committed sessions from other transactions
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
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
	// ✅ READ COMMITTED: Use r.db to see committed sessions from other transactions
	// This is CRITICAL for session reuse across multiple message imports
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

// FindByChannelPaginated retrieves sessions associated with messages from a specific channel
// Ordered by contact_id and started_at for efficient consolidation processing
// Used for history import post-processing
func (r *GormSessionRepository) FindByChannelPaginated(ctx context.Context, channelID uuid.UUID, limit int, offset int) ([]*session.Session, error) {
	// Query sessions via messages table (sessions don't have direct channel_id reference)
	// Order by contact_id, started_at for consolidation efficiency
	query := `
		SELECT DISTINCT s.*
		FROM sessions s
		INNER JOIN messages m ON m.session_id = s.id
		WHERE m.channel_id = ?
		ORDER BY s.contact_id, s.started_at
		LIMIT ? OFFSET ?
	`

	var sessionEntities []entities.SessionEntity
	if err := r.db.WithContext(ctx).Raw(query, channelID, limit, offset).Scan(&sessionEntities).Error; err != nil {
		return nil, err
	}

	// Convert to domain
	sessions := make([]*session.Session, len(sessionEntities))
	for i, entity := range sessionEntities {
		sessions[i] = r.entityToDomain(&entity)
	}

	return sessions, nil
}

// CountByChannel counts total sessions associated with a channel
func (r *GormSessionRepository) CountByChannel(ctx context.Context, channelID uuid.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(DISTINCT s.id)
		FROM sessions s
		INNER JOIN messages m ON m.session_id = s.id
		WHERE m.channel_id = ?
	`

	if err := r.db.WithContext(ctx).Raw(query, channelID).Scan(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteBatch deletes multiple sessions by ID
// Used to remove orphaned sessions after consolidation
func (r *GormSessionRepository) DeleteBatch(ctx context.Context, sessionIDs []uuid.UUID) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	db := r.getDB(ctx) // Use transaction context if present
	return db.Where("id IN ?", sessionIDs).Delete(&entities.SessionEntity{}).Error
}

// 🔥 FIX Bug 1: Get unique contact IDs for a channel
func (r *GormSessionRepository) GetContactIDsByChannel(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	var contactIDs []uuid.UUID
	query := `
		SELECT DISTINCT s.contact_id
		FROM sessions s
		INNER JOIN messages m ON m.session_id = s.id
		WHERE m.channel_id = ?
		ORDER BY s.contact_id
	`

	if err := r.db.WithContext(ctx).Raw(query, channelID).Scan(&contactIDs).Error; err != nil {
		return nil, err
	}

	return contactIDs, nil
}

// 🔥 FIX Bug 1: Get ALL sessions for specific contacts in a channel
func (r *GormSessionRepository) FindByChannelAndContacts(ctx context.Context, channelID uuid.UUID, contactIDs []uuid.UUID) ([]*session.Session, error) {
	if len(contactIDs) == 0 {
		return []*session.Session{}, nil
	}

	// Query ALL sessions for these specific contacts (ordered by contact_id, started_at)
	query := `
		SELECT DISTINCT s.*
		FROM sessions s
		INNER JOIN messages m ON m.session_id = s.id
		WHERE m.channel_id = ? AND s.contact_id IN ?
		ORDER BY s.contact_id, s.started_at
	`

	var sessionEntities []entities.SessionEntity
	if err := r.db.WithContext(ctx).Raw(query, channelID, contactIDs).Scan(&sessionEntities).Error; err != nil {
		return nil, err
	}

	// Convert to domain
	sessions := make([]*session.Session, len(sessionEntities))
	for i, entity := range sessionEntities {
		sessions[i] = r.entityToDomain(&entity)
	}

	return sessions, nil
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
