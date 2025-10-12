package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/agent"
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

	// Update with optimistic locking
	result := r.db.WithContext(ctx).Model(&entities.AgentEntity{}).
		Where("id = ? AND version = ?", entity.ID, existing.Version).
		Updates(map[string]interface{}{
			"version":            existing.Version + 1, // Increment version
			"project_id":         entity.ProjectID,
			"user_id":            entity.UserID,
			"tenant_id":          entity.TenantID,
			"name":               entity.Name,
			"email":              entity.Email,
			"type":               entity.Type,
			"status":             entity.Status,
			"active":             entity.Active,
			"config":             entity.Config,
			"sessions_handled":   entity.SessionsHandled,
			"average_response_ms": entity.AverageResponseMs,
			"last_activity_at":   entity.LastActivityAt,
			"virtual_metadata":   entity.VirtualMetadata,
			"updated_at":         entity.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	// Check optimistic locking - if 0 rows affected, version mismatch (concurrent update)
	if result.RowsAffected == 0 {
		return shared.NewOptimisticLockError(
			"Agent",
			entity.ID.String(),
			existing.Version,
			entity.Version,
		)
	}

	return nil
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
		ID:                a.ID(),
		Version:           a.Version(),
		ProjectID:         a.ProjectID(),
		UserID:            a.UserID(),
		TenantID:          a.TenantID(),
		Name:              a.Name(),
		Email:             a.Email(),
		Type:              entities.AgentType(a.Type()),
		Status:            entities.AgentStatus(a.Status()),
		Active:            a.IsActive(),
		Config:            a.Config(),
		SessionsHandled:   a.SessionsHandled(),
		AverageResponseMs: a.AverageResponseMs(),
		LastActivityAt:    a.LastActivityAt(),
		CreatedAt:         a.CreatedAt(),
		UpdatedAt:         a.UpdatedAt(),
	}

	// Serialize VirtualMetadata to JSONB
	if a.VirtualMetadata() != nil {
		vm := a.VirtualMetadata()
		entity.VirtualMetadata = map[string]interface{}{
			"represents_person_name": vm.RepresentsPersonName,
			"period_start":           vm.PeriodStart,
			"period_end":             vm.PeriodEnd,
			"reason":                 vm.Reason,
			"source_device":          vm.SourceDevice,
			"notes":                  vm.Notes,
		}
	}

	return entity
}

// entityToDomain converte entity para domain model
func (r *GormAgentRepository) entityToDomain(entity *entities.AgentEntity) *agent.Agent {
	// Deserialize permissions and settings if needed
	permissions := make(map[string]bool)
	settings := make(map[string]interface{})
	config := entity.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	// Deserialize VirtualMetadata from JSONB
	var virtualMetadata *agent.VirtualAgentMetadata
	if entity.VirtualMetadata != nil {
		vm := entity.VirtualMetadata
		virtualMetadata = &agent.VirtualAgentMetadata{}

		if v, ok := vm["represents_person_name"].(string); ok {
			virtualMetadata.RepresentsPersonName = v
		}
		if v, ok := vm["period_start"].(string); ok {
			// Parse time from string
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				virtualMetadata.PeriodStart = t
			}
		}
		if v, ok := vm["period_end"].(string); ok && v != "" {
			// Parse time from string
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				virtualMetadata.PeriodEnd = &t
			}
		}
		if v, ok := vm["reason"].(string); ok {
			virtualMetadata.Reason = v
		}
		if v, ok := vm["source_device"].(string); ok && v != "" {
			virtualMetadata.SourceDevice = &v
		}
		if v, ok := vm["notes"].(string); ok {
			virtualMetadata.Notes = v
		}
	}

	return agent.ReconstructAgent(
		entity.ID,
		entity.Version,
		entity.ProjectID,
		entity.UserID,
		entity.TenantID,
		entity.Name,
		entity.Email,
		agent.AgentType(entity.Type),
		agent.AgentStatus(entity.Status),
		agent.RoleHumanAgent, // Default role - TODO: store role in DB
		entity.Active,
		config,
		permissions,
		settings,
		entity.SessionsHandled,
		entity.AverageResponseMs,
		entity.LastActivityAt,
		virtualMetadata,
		entity.CreatedAt,
		entity.UpdatedAt,
		nil, // LastLoginAt - TODO: add to entity
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
