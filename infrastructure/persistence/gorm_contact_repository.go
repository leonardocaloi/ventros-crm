package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/application/shared"
	domainShared "github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
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
	// Usa a transação do contexto se existir, senão usa a conexão padrão
	db := r.getDB(ctx)

	// Check if exists
	var existing entities.ContactEntity
	err := db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update with optimistic locking
		result := db.Model(&entities.ContactEntity{}).
			Where("id = ? AND version = ?", entity.ID, existing.Version).
			Updates(map[string]interface{}{
				"version":                  existing.Version + 1, // Increment version
				"project_id":               entity.ProjectID,
				"tenant_id":                entity.TenantID,
				"name":                     entity.Name,
				"email":                    entity.Email,
				"phone":                    entity.Phone,
				"external_id":              entity.ExternalID,
				"source_channel":           entity.SourceChannel,
				"language":                 entity.Language,
				"timezone":                 entity.Timezone,
				"tags":                     entity.Tags,
				"profile_picture_url":      entity.ProfilePictureURL,
				"profile_picture_fetched_at": entity.ProfilePictureFetchedAt,
				"first_interaction_at":     entity.FirstInteractionAt,
				"last_interaction_at":      entity.LastInteractionAt,
				"updated_at":               entity.UpdatedAt,
			})

		if result.Error != nil {
			return result.Error
		}

		// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
		if result.RowsAffected == 0 {
			return domainShared.NewOptimisticLockError(
				"Contact",
				entity.ID.String(),
				existing.Version,
				entity.Version,
			)
		}

		return nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return db.Create(entity).Error
	}

	return err
}

// getDB retorna a transação do contexto se existir, senão retorna a conexão padrão.
// Isso permite que Save() funcione tanto dentro quanto fora de transações.
func (r *GormContactRepository) getDB(ctx context.Context) *gorm.DB {
	// Tenta extrair transação do contexto (usa shared.TransactionFromContext)
	if tx := shared.TransactionFromContext(ctx); tx != nil {
		return tx.WithContext(ctx)
	}
	// Se não houver transação, usa conexão padrão
	return r.db.WithContext(ctx)
}

func (r *GormContactRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact.Contact, error) {
	var entity entities.ContactEntity
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.NewContactNotFoundError(id.String())
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
			return nil, contact.NewContactNotFoundError("external_id:" + externalID)
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
			return nil, contact.NewContactNotFoundError("phone:" + phone)
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
			return nil, contact.NewContactNotFoundError("email:" + email)
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

// Mappers: Domain → Entity
func (r *GormContactRepository) domainToEntity(c *contact.Contact) *entities.ContactEntity {
	entity := &entities.ContactEntity{
		ID:                 c.ID(),
		Version:            c.Version(),
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

	// Handle profile picture fields
	entity.ProfilePictureURL = c.ProfilePictureURL()
	entity.ProfilePictureFetchedAt = c.ProfilePictureFetchedAt()

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
		entity.Version,
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
		entity.ProfilePictureURL,
		entity.ProfilePictureFetchedAt,
		entity.FirstInteractionAt,
		entity.LastInteractionAt,
		entity.CreatedAt,
		entity.UpdatedAt,
		deletedAt,
	)
	return c
}

// FindByTenantWithFilters finds contacts by tenant with advanced filters, pagination, and sorting
func (r *GormContactRepository) FindByTenantWithFilters(
	ctx context.Context,
	tenantID string,
	filters contact.ContactFilters,
	page, limit int,
	sortBy, sortDir string,
) ([]*contact.Contact, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.ContactEntity{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// Apply filters
	if filters.Name != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Phone != "" {
		query = query.Where("phone LIKE ?", "%"+filters.Phone+"%")
	}
	if filters.Email != "" {
		query = query.Where("LOWER(email) LIKE ?", "%"+filters.Email+"%")
	}
	if len(filters.Tags) > 0 {
		// PostgreSQL array overlap operator: tags && ARRAY['tag1','tag2']
		query = query.Where("tags && ?", filters.Tags)
	}
	if filters.CreatedAfter != "" {
		query = query.Where("created_at >= ?", filters.CreatedAfter)
	}
	if filters.CreatedBefore != "" {
		query = query.Where("created_at <= ?", filters.CreatedBefore)
	}

	// TODO: Add pipeline filters when needed
	// if filters.PipelineID != "" {
	// 	query = query.Joins("JOIN pipeline_contacts ON contacts.id = pipeline_contacts.contact_id")
	// 	query = query.Where("pipeline_contacts.pipeline_id = ?", filters.PipelineID)
	// }

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := "created_at DESC" // default
	if sortBy != "" {
		direction := "ASC"
		if sortDir == "desc" || sortDir == "DESC" {
			direction = "DESC"
		}

		// Whitelist allowed sort fields to prevent SQL injection
		allowedFields := map[string]string{
			"name":       "name",
			"created_at": "created_at",
			"updated_at": "updated_at",
		}
		if field, ok := allowedFields[sortBy]; ok {
			orderClause = field + " " + direction
		}
	}
	query = query.Order(orderClause)

	// Apply pagination
	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}
	query = query.Limit(limit).Offset(offset)

	// Execute query
	var entities []entities.ContactEntity
	if err := query.Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	contacts := make([]*contact.Contact, len(entities))
	for i, entity := range entities {
		contacts[i] = r.entityToDomain(&entity)
	}

	return contacts, total, nil
}

// SearchByText performs full-text search on contacts by name, phone, and email
func (r *GormContactRepository) SearchByText(
	ctx context.Context,
	tenantID string,
	searchText string,
	limit int,
) ([]*contact.Contact, error) {
	query := r.db.WithContext(ctx).Model(&entities.ContactEntity{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// Full-text search using ILIKE (case-insensitive)
	searchPattern := "%" + searchText + "%"
	query = query.Where(
		"LOWER(name) LIKE ? OR phone LIKE ? OR LOWER(email) LIKE ?",
		searchPattern, searchPattern, searchPattern,
	)

	// Order by relevance (name match first, then phone, then email)
	// Using CASE for simple relevance scoring
	query = query.Order(gorm.Expr(`
		CASE
			WHEN LOWER(name) LIKE ? THEN 1
			WHEN phone LIKE ? THEN 2
			WHEN LOWER(email) LIKE ? THEN 3
			ELSE 4
		END, name ASC
	`, searchPattern, searchPattern, searchPattern))

	// Apply limit
	query = query.Limit(limit)

	// Execute query
	var entities []entities.ContactEntity
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}

	// Convert to domain
	contacts := make([]*contact.Contact, len(entities))
	for i, entity := range entities {
		contacts[i] = r.entityToDomain(&entity)
	}

	return contacts, nil
}

// SaveCustomFields saves custom fields for a contact in batch
func (r *GormContactRepository) SaveCustomFields(ctx context.Context, contactID uuid.UUID, fields map[string]string) error {
	// Get contact to retrieve tenant_id
	var contactEntity entities.ContactEntity
	if err := r.db.WithContext(ctx).Select("id", "tenant_id").First(&contactEntity, "id = ?", contactID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return contact.NewContactNotFoundError(contactID.String())
		}
		return err
	}

	// Use transaction for batch operations
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range fields {
			// Upsert: insert or update if exists
			// Use ON CONFLICT to update existing records
			if err := tx.Exec(`
				INSERT INTO contact_custom_fields (id, contact_id, tenant_id, field_key, field_type, field_value, created_at, updated_at)
				VALUES (gen_random_uuid(), ?, ?, ?, ?, ?::jsonb, NOW(), NOW())
				ON CONFLICT (contact_id, field_key)
				DO UPDATE SET
					field_value = EXCLUDED.field_value,
					field_type = EXCLUDED.field_type,
					updated_at = NOW()
			`, contactID, contactEntity.TenantID, key, "text", `"`+value+`"`).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByCustomField finds a contact by a custom field key-value pair
func (r *GormContactRepository) FindByCustomField(ctx context.Context, tenantID, key, value string) (*contact.Contact, error) {
	var entity entities.ContactEntity

	// Join with custom fields and filter by key-value
	err := r.db.WithContext(ctx).
		Joins("JOIN contact_custom_fields ON contact_custom_fields.contact_id = contacts.id").
		Where("contacts.tenant_id = ? AND contacts.deleted_at IS NULL", tenantID).
		Where("contact_custom_fields.field_key = ?", key).
		Where("contact_custom_fields.field_value::text = ?", `"`+value+`"`).
		First(&entity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.NewContactNotFoundError("custom_field:" + key + "=" + value)
		}
		return nil, err
	}

	return r.entityToDomain(&entity), nil
}

// GetCustomFields retrieves all custom fields for a contact
func (r *GormContactRepository) GetCustomFields(ctx context.Context, contactID uuid.UUID) (map[string]string, error) {
	var customFields []entities.ContactCustomFieldEntity

	err := r.db.WithContext(ctx).
		Where("contact_id = ? AND deleted_at IS NULL", contactID).
		Find(&customFields).Error

	if err != nil {
		return nil, err
	}

	// Convert to map
	result := make(map[string]string)
	for _, field := range customFields {
		// Extract string value from jsonb
		if strVal, ok := field.FieldValue.(string); ok {
			result[field.FieldKey] = strVal
		}
	}

	return result, nil
}
