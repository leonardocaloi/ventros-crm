package persistence

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/agent"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormAgentRepository implementa o repositório de agentes usando GORM
type GormAgentRepository struct {
	db *gorm.DB
}

// NewGormAgentRepository cria uma nova instância do repositório
func NewGormAgentRepository(db *gorm.DB) agent.Repository {
	return &GormAgentRepository{db: db}
}

// Save salva um agente (create ou update)
func (r *GormAgentRepository) Save(ctx context.Context, a *agent.Agent) error {
	entity := r.domainToEntity(a)

	// Verifica se já existe
	var existing entities.AgentEntity
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

// FindByID busca um agente por ID
func (r *GormAgentRepository) FindByID(ctx context.Context, id uuid.UUID) (*agent.Agent, error) {
	var entity entities.AgentEntity
	err := r.db.WithContext(ctx).First(&entity, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, agent.ErrAgentNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

// FindByEmail busca um agente por email dentro de um tenant
func (r *GormAgentRepository) FindByEmail(ctx context.Context, tenantID, email string) (*agent.Agent, error) {
	var entity entities.AgentEntity
	err := r.db.WithContext(ctx).First(&entity, "tenant_id = ? AND email = ?", tenantID, email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, agent.ErrAgentNotFound
		}
		return nil, err
	}
	return r.entityToDomain(&entity), nil
}

// FindByTenant busca agentes por tenant
func (r *GormAgentRepository) FindByTenant(ctx context.Context, tenantID string) ([]*agent.Agent, error) {
	var entities []entities.AgentEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	agents := make([]*agent.Agent, len(entities))
	for i, entity := range entities {
		agents[i] = r.entityToDomain(&entity)
	}
	return agents, nil
}

// FindActiveByTenant busca agentes ativos por tenant
func (r *GormAgentRepository) FindActiveByTenant(ctx context.Context, tenantID string) ([]*agent.Agent, error) {
	var entities []entities.AgentEntity
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND active = ?", tenantID, true).
		Find(&entities).Error

	if err != nil {
		return nil, err
	}

	agents := make([]*agent.Agent, len(entities))
	for i, entity := range entities {
		agents[i] = r.entityToDomain(&entity)
	}
	return agents, nil
}

// Delete deleta um agente (soft delete)
func (r *GormAgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.AgentEntity{}, "id = ?", id).Error
}

// domainToEntity converte domain model para entity
func (r *GormAgentRepository) domainToEntity(a *agent.Agent) *entities.AgentEntity {
	entity := &entities.AgentEntity{
		ID:       a.ID(),
		TenantID: a.TenantID(),
		Name:     a.Name(),
		Email:    a.Email(),
		Active:   a.IsActive(),
	}

	// DeletedAt não está implementado no domain Agent ainda
	// TODO: Adicionar DeletedAt() ao Agent domain

	return entity
}

// entityToDomain converte entity para domain model
func (r *GormAgentRepository) entityToDomain(entity *entities.AgentEntity) *agent.Agent {
	// Deserializar permissions e settings se necessário
	permissions := make(map[string]bool)
	settings := make(map[string]interface{})

	return agent.ReconstructAgent(
		entity.ID,
		entity.TenantID,
		entity.Name,
		entity.Email,
		agent.RoleHumanAgent, // Default role
		entity.Active,
		permissions,
		settings,
		entity.CreatedAt,
		entity.UpdatedAt,
		nil, // LastLoginAt não existe na entity ainda
	)
}

func (r *GormAgentRepository) FindByTenantWithFilters(ctx context.Context, filters agent.AgentFilters) ([]*agent.Agent, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.AgentEntity{})

	// Apply tenant filter (required)
	query = query.Where("tenant_id = ?", filters.TenantID)

	// Apply optional filters
	if filters.ProjectID != nil {
		query = query.Where("project_id = ?", *filters.ProjectID)
	}
	if filters.Type != nil {
		query = query.Where("type = ?", string(*filters.Type))
	}
	if filters.Status != nil {
		query = query.Where("status = ?", string(*filters.Status))
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
	var agentEntities []entities.AgentEntity
	if err := query.Find(&agentEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	agents := make([]*agent.Agent, len(agentEntities))
	for i, entity := range agentEntities {
		agents[i] = r.entityToDomain(&entity)
	}

	return agents, total, nil
}

func (r *GormAgentRepository) SearchByText(ctx context.Context, tenantID string, searchText string, limit int, offset int) ([]*agent.Agent, int64, error) {
	query := r.db.WithContext(ctx).Model(&entities.AgentEntity{})

	// Apply tenant filter
	query = query.Where("tenant_id = ?", tenantID)

	// Text search in name and email
	searchPattern := "%" + searchText + "%"
	query = query.Where(
		r.db.Where("name ILIKE ?", searchPattern).
			Or("email ILIKE ?", searchPattern),
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
	var agentEntities []entities.AgentEntity
	if err := query.Find(&agentEntities).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	agents := make([]*agent.Agent, len(agentEntities))
	for i, entity := range agentEntities {
		agents[i] = r.entityToDomain(&entity)
	}

	return agents, total, nil
}
