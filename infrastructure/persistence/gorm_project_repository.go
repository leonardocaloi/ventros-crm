package persistence

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/project"
	"github.com/google/uuid"
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

	// Use GORM's Save method which handles both create and update
	if err := r.db.WithContext(ctx).Save(entity).Error; err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
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

// FindByCustomerID finds projects by customer ID
func (r *GormProjectRepository) FindByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*project.Project, error) {
	var entities []entities.ProjectEntity

	query := r.db.WithContext(ctx).Where("user_id = ?", customerID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&entities).Error
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
	return project.ReconstructProject(
		entity.ID,
		entity.UserID, // UserID maps to CustomerID in domain
		entity.BillingAccountID,
		entity.TenantID,
		entity.Name,
		entity.Description,
		entity.Configuration,
		entity.Active,
		entity.SessionTimeoutMinutes,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}
