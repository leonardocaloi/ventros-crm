package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/core/project"
	"github.com/ventros/crm/internal/domain/core/shared"
	"gorm.io/gorm"
)

// GormProjectRepository implements project.Repository using GORM
type GormProjectRepository struct {
	db *gorm.DB
}

// NewGormProjectRepository creates a new GORM project repository
func NewGormProjectRepository(db *gorm.DB) *GormProjectRepository {
	return &GormProjectRepository{db: db}
}

// Save saves a project to the database
func (r *GormProjectRepository) Save(ctx context.Context, proj *project.Project) error {
	entity := r.domainToEntity(proj)

	// Check if exists
	var existing entities.ProjectEntity
	err := r.db.WithContext(ctx).Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update with optimistic locking
		result := r.db.WithContext(ctx).Model(&entities.ProjectEntity{}).
			Where("id = ? AND version = ?", entity.ID, existing.Version).
			Updates(map[string]interface{}{
				"version":                 existing.Version + 1, // Increment version
				"user_id":                 entity.UserID,
				"billing_account_id":      entity.BillingAccountID,
				"tenant_id":               entity.TenantID,
				"name":                    entity.Name,
				"description":             entity.Description,
				"configuration":           entity.Configuration,
				"active":                  entity.Active,
				"session_timeout_minutes": entity.SessionTimeoutMinutes,
				"updated_at":              entity.UpdatedAt,
			})

		if result.Error != nil {
			return fmt.Errorf("failed to update project: %w", result.Error)
		}

		// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
		if result.RowsAffected == 0 {
			return shared.NewOptimisticLockError(
				"Project",
				entity.ID.String(),
				existing.Version,
				entity.Version,
			)
		}

		return nil
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		return nil
	}

	return fmt.Errorf("failed to save project: %w", err)
}

// FindByID finds a project by ID
func (r *GormProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	var entity entities.ProjectEntity

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, project.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to find project: %w", err)
	}

	return r.entityToDomain(&entity)
}

// FindByTenantID finds a project by tenant ID
func (r *GormProjectRepository) FindByTenantID(ctx context.Context, tenantID string) (*project.Project, error) {
	var entity entities.ProjectEntity

	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, project.ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to find project by tenant ID: %w", err)
	}

	return r.entityToDomain(&entity)
}

// FindByCustomer finds projects by customer ID (implements interface method)
func (r *GormProjectRepository) FindByCustomer(ctx context.Context, customerID uuid.UUID) ([]*project.Project, error) {
	var entities []entities.ProjectEntity

	err := r.db.WithContext(ctx).Where("user_id = ?", customerID).Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by customer ID: %w", err)
	}

	projects := make([]*project.Project, len(entities))
	for i, entity := range entities {
		proj, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity to domain: %w", err)
		}
		projects[i] = proj
	}

	return projects, nil
}

// FindActiveProjects finds all active projects
func (r *GormProjectRepository) FindActiveProjects(ctx context.Context, limit, offset int) ([]*project.Project, error) {
	var entities []entities.ProjectEntity

	query := r.db.WithContext(ctx).Where("active = ?", true)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active projects: %w", err)
	}

	projects := make([]*project.Project, len(entities))
	for i, entity := range entities {
		proj, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, fmt.Errorf("failed to convert entity to domain: %w", err)
		}
		projects[i] = proj
	}

	return projects, nil
}

// Delete deletes a project (soft delete)
func (r *GormProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&entities.ProjectEntity{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return project.ErrProjectNotFound
	}

	return nil
}

// domainToEntity converts domain project to entity
func (r *GormProjectRepository) domainToEntity(proj *project.Project) *entities.ProjectEntity {
	return &entities.ProjectEntity{
		ID:                    proj.ID(),
		Version:               proj.Version(),
		UserID:                proj.CustomerID(), // CustomerID maps to UserID in entity
		BillingAccountID:      proj.BillingAccountID(),
		TenantID:              proj.TenantID(),
		Name:                  proj.Name(),
		Description:           proj.Description(),
		Configuration:         proj.Configuration(),
		Active:                proj.IsActive(),
		SessionTimeoutMinutes: proj.SessionTimeoutMinutes(),
		CreatedAt:             proj.CreatedAt(),
		UpdatedAt:             proj.UpdatedAt(),
	}
}

// entityToDomain converts entity to domain project
func (r *GormProjectRepository) entityToDomain(entity *entities.ProjectEntity) (*project.Project, error) {
	// TODO: Store and retrieve agent_assignment from database (JSONB column)
	// For now, initialize with default config
	return project.ReconstructProject(
		entity.ID,
		entity.Version,
		entity.UserID, // UserID maps to CustomerID in domain
		entity.BillingAccountID,
		entity.TenantID,
		entity.Name,
		entity.Description,
		entity.Configuration,
		entity.Active,
		entity.SessionTimeoutMinutes,
		nil, // agentAssignment - will be initialized with default by ReconstructProject
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}

func (r *GormProjectRepository) FindByTenantWithFilters(ctx context.Context, filters project.ProjectFilters) ([]*project.Project, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.ProjectEntity{})

	// Apply tenant filter (required)
	query = query.Where("tenant_id = ?", filters.TenantID)

	// Apply optional filters
	if filters.CustomerID != nil {
		query = query.Where("user_id = ?", *filters.CustomerID)
	}
	if filters.Active != nil {
		query = query.Where("active = ?", *filters.Active)
	}

	// Count total results
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "name"
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
	var projectEntities []entities.ProjectEntity
	if err := query.Find(&projectEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	projects := make([]*project.Project, len(projectEntities))
	for i, entity := range projectEntities {
		proj, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert entity to domain: %w", err)
		}
		projects[i] = proj
	}

	return projects, total, nil
}

func (r *GormProjectRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*project.Project, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.ProjectEntity{})

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
	query = query.Order("name ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	var projectEntities []entities.ProjectEntity
	if err := query.Find(&projectEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	projects := make([]*project.Project, len(projectEntities))
	for i, entity := range projectEntities {
		proj, err := r.entityToDomain(&entity)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert entity to domain: %w", err)
		}
		projects[i] = proj
	}

	return projects, total, nil
}
