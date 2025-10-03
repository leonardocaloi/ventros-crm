package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormContactRepository struct {
	db *gorm.DB
}

func NewGormContactRepository(db *gorm.DB) contact.Repository {
	return &GormContactRepository{db: db}
}

func (r *GormContactRepository) Save(ctx context.Context, c *contact.Contact) error {
	entity := r.domainToEntity(c)
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error) {
	var entity entities.ContactEntity
	err := r.db.First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.ErrContactNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormContactRepository) FindByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*contact.Contact, error) {
	var entities []entities.ContactEntity
	query := r.db.WithContext(ctx).Where("project_id = ? AND deleted_at IS NULL", projectID)
	
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

	contacts := make([]*contact.Contact, len(entities))
	for i, entity := range entities {
		contacts[i] = r.entityToDomain(&entity)
	}
	return contacts, nil
}

func (r *GormContactRepository) CountByProject(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.ContactEntity{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).Count(&count).Error
	return int(count), err
}

func (r *GormContactRepository) FindByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*contact.Contact, error) {
	var entity entities.ContactEntity
	err := r.db.WithContext(ctx).Where("project_id = ? AND external_id = ? AND deleted_at IS NULL", projectID, externalID).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.ErrContactNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormContactRepository) FindByPhone(ctx context.Context, projectID uuid.UUID, phone string) (*contact.Contact, error) {
	var entity entities.ContactEntity
	err := r.db.WithContext(ctx).Where("project_id = ? AND phone = ? AND deleted_at IS NULL", projectID, phone).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.ErrContactNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

func (r *GormContactRepository) FindByEmail(ctx context.Context, projectID uuid.UUID, email string) (*contact.Contact, error) {
	var entity entities.ContactEntity
	err := r.db.WithContext(ctx).Where("project_id = ? AND email = ? AND deleted_at IS NULL", projectID, email).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.ErrContactNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

// Mappers: Domain → Entity
func (r *GormContactRepository) domainToEntity(c *contact.Contact) *entities.ContactEntity {
	entity := &entities.ContactEntity{
		ID:                 c.ID(),
		ProjectID:          c.ProjectID(),
		TenantID:           c.TenantID(),
		Name:               c.Name(),
		Language:           c.Language(),
		Tags:               entities.StringArray(c.Tags()),
		FirstInteractionAt: c.FirstInteractionAt(),
		LastInteractionAt:  c.LastInteractionAt(),
		CreatedAt:          c.CreatedAt(),
		UpdatedAt:          c.UpdatedAt(),
	}

	// Handle Email value object
	if email := c.Email(); email != nil {
		emailStr := email.String()
		entity.Email = emailStr
	}

	// Handle Phone value object
	if phone := c.Phone(); phone != nil {
		phoneStr := phone.String()
		entity.Phone = phoneStr
	}

	// Handle optional string fields
	if externalID := c.ExternalID(); externalID != nil {
		entity.ExternalID = *externalID
	}
	if sourceChannel := c.SourceChannel(); sourceChannel != nil {
		entity.SourceChannel = *sourceChannel
	}
	if timezone := c.Timezone(); timezone != nil {
		entity.Timezone = *timezone
	}

	if c.DeletedAt() != nil {
		entity.DeletedAt = gorm.DeletedAt{Time: *c.DeletedAt(), Valid: true}
	}

	return entity
}

// Mappers: Entity → Domain
func (r *GormContactRepository) entityToDomain(entity *entities.ContactEntity) *contact.Contact {
	var deletedAt *time.Time
	if entity.DeletedAt.Valid {
		deletedAt = &entity.DeletedAt.Time
	}

	// Convert string fields to value objects
	var email *contact.Email
	if entity.Email != "" {
		if e, err := contact.NewEmail(entity.Email); err == nil {
			email = &e
		}
	}

	var phone *contact.Phone
	if entity.Phone != "" {
		if p, err := contact.NewPhone(entity.Phone); err == nil {
			phone = &p
		}
	}

	// Handle optional string fields
	var externalID *string
	if entity.ExternalID != "" {
		externalID = &entity.ExternalID
	}
	var sourceChannel *string
	if entity.SourceChannel != "" {
		sourceChannel = &entity.SourceChannel
	}

	// Handle optional string fields for domain
	var timezone *string
	if entity.Timezone != "" {
		timezone = &entity.Timezone
	}

	// Reconstruct domain object
	c := contact.ReconstructContact(
		entity.ID,
		entity.ProjectID,
		entity.TenantID,
		entity.Name,
		email,
		phone,
		externalID,
		sourceChannel,
		entity.Language,
		timezone,
		[]string(entity.Tags),
		entity.FirstInteractionAt,
		entity.LastInteractionAt,
		entity.CreatedAt,
		entity.UpdatedAt,
		deletedAt,
	)
	return c
}
