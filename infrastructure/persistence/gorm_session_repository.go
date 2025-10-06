package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/session"
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
	return r.db.WithContext(ctx).Save(entity).Error
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

// Mappers: Domain → Entity
func (r *GormSessionRepository) domainToEntity(s *session.Session) *entities.SessionEntity {
	entity := &entities.SessionEntity{
		ID:                  s.ID(),
		ContactID:           s.ContactID(),
		TenantID:            s.TenantID(),
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
		entity.ContactID,
		entity.TenantID,
		entity.ChannelTypeID,
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
