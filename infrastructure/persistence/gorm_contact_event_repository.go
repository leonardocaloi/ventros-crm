package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	contact_event "github.com/ventros/crm/internal/domain/crm/contact_event"
	"gorm.io/gorm"
)

type GormContactEventRepository struct {
	db *gorm.DB
}

func NewGormContactEventRepository(db *gorm.DB) contact_event.Repository {
	return &GormContactEventRepository{db: db}
}

func (r *GormContactEventRepository) Save(ctx context.Context, event *contact_event.ContactEvent) error {
	entity := r.domainToEntity(event)
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *GormContactEventRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact_event.ContactEvent, error) {
	var entity entities.ContactEventEntity
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact_event.ErrContactEventNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormContactEventRepository) FindByContactID(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*contact_event.ContactEvent, error) {
	var entities []entities.ContactEventEntity
	query := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("occurred_at DESC")

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

	events := make([]*contact_event.ContactEvent, len(entities))
	for i, entity := range entities {
		events[i] = r.entityToDomain(&entity)
	}
	return events, nil
}

func (r *GormContactEventRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*contact_event.ContactEvent, error) {
	var entities []entities.ContactEventEntity
	query := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("occurred_at DESC")

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

	events := make([]*contact_event.ContactEvent, len(entities))
	for i, entity := range entities {
		events[i] = r.entityToDomain(&entity)
	}
	return events, nil
}

func (r *GormContactEventRepository) FindUndeliveredRealtime(ctx context.Context, limit int) ([]*contact_event.ContactEvent, error) {
	var entities []entities.ContactEventEntity
	query := r.db.WithContext(ctx).
		Where("is_realtime = ? AND delivered = ? AND (expires_at IS NULL OR expires_at > NOW())", true, false).
		Order("occurred_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&entities).Error
	if err != nil {
		return nil, err
	}

	events := make([]*contact_event.ContactEvent, len(entities))
	for i, entity := range entities {
		events[i] = r.entityToDomain(&entity)
	}
	return events, nil
}

func (r *GormContactEventRepository) Update(ctx context.Context, event *contact_event.ContactEvent) error {
	entity := r.domainToEntity(event)
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormContactEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.ContactEventEntity{}, "id = ?", id).Error
}

// Métodos ainda não implementados (interface completa)

func (r *GormContactEventRepository) FindByContactIDVisible(ctx context.Context, contactID uuid.UUID, visibleToClient bool, limit int, offset int) ([]*contact_event.ContactEvent, error) {
	var entities []entities.ContactEventEntity
	query := r.db.WithContext(ctx).
		Where("contact_id = ? AND visible_to_client = ?", contactID, visibleToClient).
		Order("occurred_at DESC")

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

	events := make([]*contact_event.ContactEvent, len(entities))
	for i, entity := range entities {
		events[i] = r.entityToDomain(&entity)
	}
	return events, nil
}

func (r *GormContactEventRepository) FindUndeliveredForContact(ctx context.Context, contactID uuid.UUID) ([]*contact_event.ContactEvent, error) {
	var entities []entities.ContactEventEntity
	err := r.db.WithContext(ctx).
		Where("contact_id = ? AND delivered = ? AND (expires_at IS NULL OR expires_at > NOW())", contactID, false).
		Order("occurred_at ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	events := make([]*contact_event.ContactEvent, len(entities))
	for i, entity := range entities {
		events[i] = r.entityToDomain(&entity)
	}
	return events, nil
}

func (r *GormContactEventRepository) FindByTenantAndType(ctx context.Context, tenantID string, eventType string, since time.Time, limit int) ([]*contact_event.ContactEvent, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

func (r *GormContactEventRepository) FindByCategory(ctx context.Context, tenantID string, category contact_event.Category, since time.Time, limit int) ([]*contact_event.ContactEvent, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

func (r *GormContactEventRepository) FindExpired(ctx context.Context, before time.Time, limit int) ([]*contact_event.ContactEvent, error) {
	// TODO: Implement
	return nil, errors.New("not implemented")
}

func (r *GormContactEventRepository) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	// TODO: Implement
	return 0, errors.New("not implemented")
}

func (r *GormContactEventRepository) CountByContact(ctx context.Context, contactID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ContactEventEntity{}).
		Where("contact_id = ?", contactID).
		Count(&count).Error
	return int(count), err
}

// FindByContactWithFilters busca eventos com filtros avançados
func (r *GormContactEventRepository) FindByContactWithFilters(
	ctx context.Context,
	contactID uuid.UUID,
	sessionID *uuid.UUID,
	eventTypes []string,
	categories []string,
	priority string,
	startDate *time.Time,
	endDate *time.Time,
	limit int,
	offset int,
) ([]*entities.ContactEventEntity, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.ContactEventEntity{})

	// Filtro obrigatório: contact_id
	query = query.Where("contact_id = ?", contactID)

	// Filtro opcional: session_id
	if sessionID != nil {
		query = query.Where("session_id = ?", *sessionID)
	}

	// Filtro opcional: event_types
	if len(eventTypes) > 0 {
		query = query.Where("event_type IN ?", eventTypes)
	}

	// Filtro opcional: categories
	if len(categories) > 0 {
		query = query.Where("category IN ?", categories)
	}

	// Filtro opcional: priority
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Filtro opcional: data range
	if startDate != nil {
		query = query.Where("occurred_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("occurred_at <= ?", *endDate)
	}

	// Contar total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Buscar com paginação
	var entities []*entities.ContactEventEntity
	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entities).Error

	return entities, total, err
}

// MarkAsRead marca um evento como lido
func (r *GormContactEventRepository) MarkAsRead(ctx context.Context, eventID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.ContactEventEntity{}).
		Where("id = ?", eventID).
		Updates(map[string]interface{}{
			"read":    true,
			"read_at": now,
		}).Error
}

// domainToEntity converte domain model para entity
func (r *GormContactEventRepository) domainToEntity(event *contact_event.ContactEvent) *entities.ContactEventEntity {
	return &entities.ContactEventEntity{
		ID:                event.ID(),
		ContactID:         event.ContactID(),
		SessionID:         event.SessionID(),
		TenantID:          event.TenantID(),
		EventType:         event.EventType(),
		Category:          string(event.Category()),
		Priority:          string(event.Priority()),
		Title:             event.Title(),
		Description:       event.Description(),
		Payload:           event.Payload(),
		Metadata:          event.Metadata(),
		Source:            string(event.Source()),
		TriggeredBy:       event.TriggeredBy(),
		IntegrationSource: event.IntegrationSource(),
		IsRealtime:        event.ShouldBeDeliveredInRealtime(),
		Delivered:         event.IsDelivered(),
		DeliveredAt:       event.DeliveredAt(),
		Read:              event.IsRead(),
		ReadAt:            event.ReadAt(),
		VisibleToClient:   event.IsVisibleToClient(),
		VisibleToAgent:    event.IsVisibleToAgent(),
		ExpiresAt:         event.ExpiresAt(),
		OccurredAt:        event.OccurredAt(),
		CreatedAt:         event.CreatedAt(),
	}
}

// entityToDomain converte entity para domain model
func (r *GormContactEventRepository) entityToDomain(entity *entities.ContactEventEntity) *contact_event.ContactEvent {
	return contact_event.ReconstructContactEvent(
		entity.ID,
		entity.ContactID,
		entity.SessionID,
		entity.TenantID,
		entity.EventType,
		contact_event.Category(entity.Category),
		contact_event.Priority(entity.Priority),
		entity.Title,
		entity.Description,
		entity.Payload,
		entity.Metadata,
		contact_event.Source(entity.Source),
		entity.TriggeredBy,
		entity.IntegrationSource,
		entity.IsRealtime,
		entity.Delivered,
		entity.DeliveredAt,
		entity.Read,
		entity.ReadAt,
		entity.VisibleToClient,
		entity.VisibleToAgent,
		entity.ExpiresAt,
		entity.OccurredAt,
		entity.CreatedAt,
	)
}
