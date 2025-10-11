package persistence

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormPipelineRepository struct {
	db *gorm.DB
}

func NewGormPipelineRepository(db *gorm.DB) pipeline.Repository {
	return &GormPipelineRepository{db: db}
}

// Pipeline operations
func (r *GormPipelineRepository) SavePipeline(ctx context.Context, p *pipeline.Pipeline) error {
	entity := r.pipelineDomainToEntity(p)
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormPipelineRepository) FindPipelineByID(ctx context.Context, id uuid.UUID) (*pipeline.Pipeline, error) {
	var entity entities.PipelineEntity
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&entity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("pipeline not found")
		}
		return nil, err
	}

	return r.pipelineEntityToDomain(&entity), nil
}

func (r *GormPipelineRepository) FindPipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*pipeline.Pipeline, error) {
	var entities []entities.PipelineEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("position ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	pipelines := make([]*pipeline.Pipeline, len(entities))
	for i, entity := range entities {
		pipelines[i] = r.pipelineEntityToDomain(&entity)
	}

	return pipelines, nil
}

func (r *GormPipelineRepository) FindPipelinesByTenant(ctx context.Context, tenantID string) ([]*pipeline.Pipeline, error) {
	var entities []entities.PipelineEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("position ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	pipelines := make([]*pipeline.Pipeline, len(entities))
	for i, entity := range entities {
		pipelines[i] = r.pipelineEntityToDomain(&entity)
	}

	return pipelines, nil
}

func (r *GormPipelineRepository) FindActivePipelinesByProject(ctx context.Context, projectID uuid.UUID) ([]*pipeline.Pipeline, error) {
	var entities []entities.PipelineEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND active = true", projectID).
		Order("position ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	pipelines := make([]*pipeline.Pipeline, len(entities))
	for i, entity := range entities {
		pipelines[i] = r.pipelineEntityToDomain(&entity)
	}

	return pipelines, nil
}

func (r *GormPipelineRepository) DeletePipeline(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.PipelineEntity{}, "id = ?", id).Error
}

// Status operations
func (r *GormPipelineRepository) SaveStatus(ctx context.Context, status *pipeline.Status) error {
	entity := r.statusDomainToEntity(status)
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *GormPipelineRepository) FindStatusByID(ctx context.Context, id uuid.UUID) (*pipeline.Status, error) {
	var entity entities.PipelineStatusEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("status not found")
		}
		return nil, err
	}
	return r.statusEntityToDomain(&entity), nil
}

func (r *GormPipelineRepository) FindStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Status, error) {
	var entities []entities.PipelineStatusEntity
	err := r.db.WithContext(ctx).
		Where("pipeline_id = ?", pipelineID).
		Order("position ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	statuses := make([]*pipeline.Status, len(entities))
	for i, entity := range entities {
		statuses[i] = r.statusEntityToDomain(&entity)
	}

	return statuses, nil
}

func (r *GormPipelineRepository) FindActiveStatusesByPipeline(ctx context.Context, pipelineID uuid.UUID) ([]*pipeline.Status, error) {
	var entities []entities.PipelineStatusEntity
	err := r.db.WithContext(ctx).
		Where("pipeline_id = ? AND active = true", pipelineID).
		Order("position ASC").
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	statuses := make([]*pipeline.Status, len(entities))
	for i, entity := range entities {
		statuses[i] = r.statusEntityToDomain(&entity)
	}

	return statuses, nil
}

func (r *GormPipelineRepository) DeleteStatus(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.PipelineStatusEntity{}, "id = ?", id).Error
}

// Pipeline-Status relationships
func (r *GormPipelineRepository) AddStatusToPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error {
	// Status já tem pipeline_id, então não precisa de tabela intermediária
	return r.db.WithContext(ctx).
		Model(&entities.PipelineStatusEntity{}).
		Where("id = ?", statusID).
		Update("pipeline_id", pipelineID).Error
}

func (r *GormPipelineRepository) RemoveStatusFromPipeline(ctx context.Context, pipelineID, statusID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND pipeline_id = ?", statusID, pipelineID).
		Delete(&entities.PipelineStatusEntity{}).Error
}

func (r *GormPipelineRepository) GetPipelineWithStatuses(ctx context.Context, pipelineID uuid.UUID) (*pipeline.Pipeline, []*pipeline.Status, error) {
	p, err := r.FindPipelineByID(ctx, pipelineID)
	if err != nil {
		return nil, nil, err
	}

	statuses, err := r.FindStatusesByPipeline(ctx, pipelineID)
	if err != nil {
		return nil, nil, err
	}

	return p, statuses, nil
}

// Contact-Status relationships (stub implementations)
func (r *GormPipelineRepository) SetContactStatus(ctx context.Context, contactID, pipelineID, statusID uuid.UUID) error {
	// TODO: Implement using ContactPipelineStatusEntity
	return errors.New("not implemented")
}

func (r *GormPipelineRepository) GetContactStatus(ctx context.Context, contactID, pipelineID uuid.UUID) (*pipeline.Status, error) {
	// TODO: Implement using ContactPipelineStatusEntity
	return nil, errors.New("not implemented")
}

func (r *GormPipelineRepository) GetContactsByStatus(ctx context.Context, pipelineID, statusID uuid.UUID) ([]uuid.UUID, error) {
	// TODO: Implement using ContactPipelineStatusEntity
	return nil, errors.New("not implemented")
}

func (r *GormPipelineRepository) GetContactStatusHistory(ctx context.Context, contactID, pipelineID uuid.UUID) ([]*pipeline.ContactStatusHistory, error) {
	// TODO: Implement using ContactPipelineStatusEntity
	return nil, errors.New("not implemented")
}

func (r *GormPipelineRepository) FindByTenantWithFilters(ctx context.Context, filters pipeline.PipelineFilters) ([]*pipeline.Pipeline, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.PipelineEntity{})

	// Apply tenant filter (required)
	query = query.Where("tenant_id = ?", filters.TenantID)

	// Apply optional filters
	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}
	if filters.Active != nil {
		query = query.Where("active = ?", *filters.Active)
	}
	if filters.Color != nil {
		query = query.Where("color = ?", *filters.Color)
	}

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "position"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortOrder := "ASC"
	if filters.SortOrder == "desc" {
		sortOrder = "DESC"
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
	var pipelineEntities []entities.PipelineEntity
	if err := query.Find(&pipelineEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	pipelines := make([]*pipeline.Pipeline, len(pipelineEntities))
	for i, entity := range pipelineEntities {
		pipelines[i] = r.pipelineEntityToDomain(&entity)
	}

	return pipelines, total, nil
}

func (r *GormPipelineRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*pipeline.Pipeline, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.PipelineEntity{})

	// Apply tenant filter
	query = query.Where("tenant_id = ?", tenantID)

	// Text search in name and description
	searchPattern := "%" + searchText + "%"
	query = query.Where(
		r.db.Where("name ILIKE ?", searchPattern).
			Or("description ILIKE ?", searchPattern),
	)

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	query = query.Order("position ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	var pipelineEntities []entities.PipelineEntity
	if err := query.Find(&pipelineEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	pipelines := make([]*pipeline.Pipeline, len(pipelineEntities))
	for i, entity := range pipelineEntities {
		pipelines[i] = r.pipelineEntityToDomain(&entity)
	}

	return pipelines, total, nil
}

// Mappers
func (r *GormPipelineRepository) pipelineDomainToEntity(p *pipeline.Pipeline) *entities.PipelineEntity {
	return &entities.PipelineEntity{
		ID:                    p.ID(),
		ProjectID:             p.ProjectID(),
		TenantID:              p.TenantID(),
		Name:                  p.Name(),
		Description:           p.Description(),
		Color:                 p.Color(),
		Position:              p.Position(),
		Active:                p.IsActive(),
		SessionTimeoutMinutes: p.SessionTimeoutMinutes(),
		CreatedAt:             p.CreatedAt(),
		UpdatedAt:             p.UpdatedAt(),
	}
}

func (r *GormPipelineRepository) pipelineEntityToDomain(entity *entities.PipelineEntity) *pipeline.Pipeline {
	// TODO: Load leadQualificationConfig from JSONB column when implemented
	return pipeline.ReconstructPipeline(
		entity.ID,
		entity.ProjectID,
		entity.TenantID,
		entity.Name,
		entity.Description,
		entity.Color,
		entity.Position,
		entity.Active,
		entity.SessionTimeoutMinutes,
		nil, // leadQualificationConfig - TODO: parse from entity
		entity.CreatedAt,
		entity.UpdatedAt,
	)
}

func (r *GormPipelineRepository) statusDomainToEntity(s *pipeline.Status) *entities.PipelineStatusEntity {
	return &entities.PipelineStatusEntity{
		ID:          s.ID(),
		PipelineID:  s.PipelineID(),
		Name:        s.Name(),
		Description: s.Description(),
		Color:       s.Color(),
		StatusType:  string(s.StatusType()),
		Position:    s.Position(),
		Active:      s.IsActiveStatus(),
		CreatedAt:   s.CreatedAt(),
		UpdatedAt:   s.UpdatedAt(),
	}
}

func (r *GormPipelineRepository) statusEntityToDomain(entity *entities.PipelineStatusEntity) *pipeline.Status {
	return pipeline.ReconstructStatus(
		entity.ID,
		entity.PipelineID,
		entity.Name,
		entity.Description,
		entity.Color,
		pipeline.StatusType(entity.StatusType),
		entity.Position,
		entity.Active,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
}
