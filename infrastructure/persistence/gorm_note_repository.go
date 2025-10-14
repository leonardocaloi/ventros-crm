package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/crm/note"
	"gorm.io/gorm"
)

// GormNoteRepository implementa o repositório de notas usando GORM
type GormNoteRepository struct {
	db *gorm.DB
}

// NewGormNoteRepository cria uma nova instância do repositório
func NewGormNoteRepository(db *gorm.DB) note.Repository {
	return &GormNoteRepository{db: db}
}

// Save salva uma nota (create ou update)
func (r *GormNoteRepository) Save(ctx context.Context, n *note.Note) error {
	entity := r.domainToEntity(n)

	// Verifica se já existe
	var existing entities.NoteEntity
	err := r.db.WithContext(ctx).First(&existing, "id = ?", entity.ID).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create
		return r.db.WithContext(ctx).Create(entity).Error
	}

	// Update
	return r.db.WithContext(ctx).Save(entity).Error
}

// FindByID busca uma nota por ID
func (r *GormNoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	var entity entities.NoteEntity
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, note.ErrNoteNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity)
}

// FindByContactID busca notas de um contato
func (r *GormNoteRepository) FindByContactID(ctx context.Context, contactID uuid.UUID, limit, offset int) ([]*note.Note, error) {
	var entities []entities.NoteEntity
	query := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("pinned DESC, created_at DESC")

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

	notes := make([]*note.Note, len(entities))
	for i, entity := range entities {
		n, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		notes[i] = n
	}
	return notes, nil
}

// FindBySessionID busca notas de uma sessão
func (r *GormNoteRepository) FindBySessionID(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*note.Note, error) {
	var entities []entities.NoteEntity
	query := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at DESC")

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

	notes := make([]*note.Note, len(entities))
	for i, entity := range entities {
		n, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		notes[i] = n
	}
	return notes, nil
}

// FindPinned busca notas fixadas de um contato
func (r *GormNoteRepository) FindPinned(ctx context.Context, contactID uuid.UUID) ([]*note.Note, error) {
	var entities []entities.NoteEntity
	err := r.db.WithContext(ctx).
		Where("contact_id = ? AND pinned = ?", contactID, true).
		Order("created_at DESC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	notes := make([]*note.Note, len(entities))
	for i, entity := range entities {
		n, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, err
		}
		notes[i] = n
	}
	return notes, nil
}

// Delete deleta uma nota (soft delete)
func (r *GormNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.NoteEntity{}, "id = ?", id).Error
}

// CountByContact conta notas de um contato
func (r *GormNoteRepository) CountByContact(ctx context.Context, contactID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.NoteEntity{}).
		Where("contact_id = ?", contactID).
		Count(&count).Error
	return int(count), err
}

// domainToEntity converte domain model para entity
func (r *GormNoteRepository) domainToEntity(n *note.Note) *entities.NoteEntity {
	entity := &entities.NoteEntity{
		ID:              n.ID(),
		ContactID:       n.ContactID(),
		SessionID:       n.SessionID(),
		TenantID:        n.TenantID(),
		AuthorID:        n.AuthorID(),
		AuthorType:      string(n.AuthorType()),
		AuthorName:      n.AuthorName(),
		Content:         n.Content(),
		NoteType:        string(n.NoteType()),
		Priority:        string(n.Priority()),
		VisibleToClient: n.VisibleToClient(),
		Pinned:          n.Pinned(),
		Tags:            n.Tags(),
		Attachments:     n.Attachments(),
		CreatedAt:       n.CreatedAt(),
		UpdatedAt:       n.UpdatedAt(),
	}

	// Serializar mentions como JSON
	if mentions := n.Mentions(); len(mentions) > 0 {
		mentionsJSON, _ := json.Marshal(mentions)
		entity.Mentions = mentionsJSON
	}

	if n.DeletedAt() != nil {
		entity.DeletedAt = gorm.DeletedAt{Time: *n.DeletedAt(), Valid: true}
	}

	return entity
}

// entityToDomain converte entity para domain model
func (r *GormNoteRepository) entityToDomain(entity *entities.NoteEntity) (*note.Note, error) {
	var deletedAt *time.Time
	if entity.DeletedAt.Valid {
		deletedAt = &entity.DeletedAt.Time
	}

	// Deserializar mentions
	var mentions []uuid.UUID
	if len(entity.Mentions) > 0 {
		if err := json.Unmarshal(entity.Mentions, &mentions); err != nil {
			mentions = []uuid.UUID{}
		}
	}

	return note.ReconstructNote(
		entity.ID,
		entity.ContactID,
		entity.SessionID,
		entity.TenantID,
		entity.AuthorID,
		note.AuthorType(entity.AuthorType),
		entity.AuthorName,
		entity.Content,
		note.NoteType(entity.NoteType),
		note.Priority(entity.Priority),
		entity.VisibleToClient,
		entity.Pinned,
		[]string(entity.Tags),
		mentions,
		[]string(entity.Attachments),
		entity.CreatedAt,
		entity.UpdatedAt,
		deletedAt,
	), nil
}

func (r *GormNoteRepository) FindByTenantWithFilters(ctx context.Context, filters note.NoteFilters) ([]*note.Note, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.NoteEntity{})

	// Apply tenant filter (required)
	query = query.Where("tenant_id = ?", filters.TenantID)

	// Apply optional filters
	if filters.ContactID != nil {
		query = query.Where("contact_id = ?", *filters.ContactID)
	}
	if filters.SessionID != nil {
		query = query.Where("session_id = ?", *filters.SessionID)
	}
	if filters.AuthorID != nil {
		query = query.Where("author_id = ?", *filters.AuthorID)
	}
	if filters.AuthorType != nil {
		query = query.Where("author_type = ?", *filters.AuthorType)
	}
	if filters.NoteType != nil {
		query = query.Where("note_type = ?", *filters.NoteType)
	}
	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}
	if filters.VisibleToClient != nil {
		query = query.Where("visible_to_client = ?", *filters.VisibleToClient)
	}
	if filters.Pinned != nil {
		query = query.Where("pinned = ?", *filters.Pinned)
	}
	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
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
	var noteEntities []entities.NoteEntity
	if err := query.Find(&noteEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	notes := make([]*note.Note, len(noteEntities))
	for i, entity := range noteEntities {
		n, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, 0, err
		}
		notes[i] = n
	}

	return notes, total, nil
}

func (r *GormNoteRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*note.Note, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.NoteEntity{})

	// Apply tenant filter
	query = query.Where("tenant_id = ?", tenantID)

	// Text search in content and author_name
	searchPattern := "%" + searchText + "%"
	query = query.Where(
		r.db.Where("content ILIKE ?", searchPattern).
			Or("author_name ILIKE ?", searchPattern),
	)

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	var noteEntities []entities.NoteEntity
	if err := query.Find(&noteEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	notes := make([]*note.Note, len(noteEntities))
	for i, entity := range noteEntities {
		n, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, 0, err
		}
		notes[i] = n
	}

	return notes, total, nil
}
