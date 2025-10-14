package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ventros/crm/infrastructure/persistence/entities"
	"github.com/ventros/crm/internal/domain/core/shared"
	"github.com/ventros/crm/internal/domain/crm/project_member"
	"gorm.io/gorm"
)

// GormProjectMemberRepository implementa project_member.Repository usando GORM
type GormProjectMemberRepository struct {
	db *gorm.DB
}

// NewGormProjectMemberRepository cria uma nova instância do repositório
func NewGormProjectMemberRepository(db *gorm.DB) *GormProjectMemberRepository {
	return &GormProjectMemberRepository{db: db}
}

// Save persiste um ProjectMember (create ou update)
func (r *GormProjectMemberRepository) Save(ctx context.Context, member *project_member.ProjectMember) error {
	entity := r.toEntity(member)

	// Check if exists
	var existing entities.ProjectMemberEntity
	err := r.db.WithContext(ctx).Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// UPDATE with optimistic locking
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			result := tx.Model(&entities.ProjectMemberEntity{}).
				Where("id = ? AND version = ?", entity.ID, existing.Version).
				Updates(map[string]interface{}{
					"version":    existing.Version + 1,
					"project_id": entity.ProjectID,
					"agent_id":   entity.AgentID,
					"role":       entity.Role,
					"invited_by": entity.InvitedBy,
					"invited_at": entity.InvitedAt,
					"updated_at": entity.UpdatedAt,
				})

			if result.Error != nil {
				return result.Error
			}

			// Check optimistic locking
			if result.RowsAffected == 0 {
				return shared.NewOptimisticLockError(
					"ProjectMember",
					entity.ID.String(),
					existing.Version,
					entity.Version,
				)
			}

			return nil
		})
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// INSERT
		return r.db.WithContext(ctx).Create(entity).Error
	}

	return err
}

// FindByID busca um ProjectMember por ID
func (r *GormProjectMemberRepository) FindByID(ctx context.Context, id uuid.UUID) (*project_member.ProjectMember, error) {
	var entity entities.ProjectMemberEntity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, project_member.ErrMemberNotFound
		}
		return nil, err
	}
	return r.toDomain(&entity), nil
}

// FindByProjectAndAgent busca um membro específico de um projeto
func (r *GormProjectMemberRepository) FindByProjectAndAgent(ctx context.Context, projectID uuid.UUID, agentID string) (*project_member.ProjectMember, error) {
	var entity entities.ProjectMemberEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND agent_id = ?", projectID, agentID).
		First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, project_member.ErrMemberNotFound
		}
		return nil, err
	}
	return r.toDomain(&entity), nil
}

// FindByProject busca todos os membros de um projeto
func (r *GormProjectMemberRepository) FindByProject(ctx context.Context, projectID uuid.UUID) ([]*project_member.ProjectMember, error) {
	var entities []entities.ProjectMemberEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("role ASC, created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	members := make([]*project_member.ProjectMember, 0, len(entities))
	for _, e := range entities {
		members = append(members, r.toDomain(&e))
	}
	return members, nil
}

// FindByAgent busca todos os projetos de um agent
func (r *GormProjectMemberRepository) FindByAgent(ctx context.Context, agentID string) ([]*project_member.ProjectMember, error) {
	var entities []entities.ProjectMemberEntity
	err := r.db.WithContext(ctx).
		Where("agent_id = ?", agentID).
		Order("created_at DESC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	members := make([]*project_member.ProjectMember, 0, len(entities))
	for _, e := range entities {
		members = append(members, r.toDomain(&e))
	}
	return members, nil
}

// FindAdminsByProject busca todos os admins de um projeto
func (r *GormProjectMemberRepository) FindAdminsByProject(ctx context.Context, projectID uuid.UUID) ([]*project_member.ProjectMember, error) {
	var entities []entities.ProjectMemberEntity
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND role = ?", projectID, string(project_member.RoleAdmin)).
		Order("created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	members := make([]*project_member.ProjectMember, 0, len(entities))
	for _, e := range entities {
		members = append(members, r.toDomain(&e))
	}
	return members, nil
}

// CountAdminsByProject conta quantos admins tem em um projeto
func (r *GormProjectMemberRepository) CountAdminsByProject(ctx context.Context, projectID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ProjectMemberEntity{}).
		Where("project_id = ? AND role = ?", projectID, string(project_member.RoleAdmin)).
		Count(&count).Error
	return int(count), err
}

// ExistsInProject verifica se um agent já é membro de um projeto
func (r *GormProjectMemberRepository) ExistsInProject(ctx context.Context, projectID uuid.UUID, agentID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ProjectMemberEntity{}).
		Where("project_id = ? AND agent_id = ?", projectID, agentID).
		Count(&count).Error
	return count > 0, err
}

// Delete remove um ProjectMember (soft delete)
func (r *GormProjectMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.ProjectMemberEntity{}, "id = ?", id).Error
}

// toEntity converte domain para entity
func (r *GormProjectMemberRepository) toEntity(member *project_member.ProjectMember) *entities.ProjectMemberEntity {
	return &entities.ProjectMemberEntity{
		ID:        member.ID(),
		Version:   member.Version(),
		ProjectID: member.ProjectID(),
		AgentID:   member.AgentID(),
		Role:      string(member.Role()),
		InvitedBy: member.InvitedBy(),
		InvitedAt: member.InvitedAt(),
		CreatedAt: member.CreatedAt(),
		UpdatedAt: member.UpdatedAt(),
	}
}

// toDomain converte entity para domain
func (r *GormProjectMemberRepository) toDomain(entity *entities.ProjectMemberEntity) *project_member.ProjectMember {
	return project_member.ReconstructProjectMember(
		entity.ID,
		entity.Version,
		entity.ProjectID,
		entity.AgentID,
		project_member.ProjectMemberRole(entity.Role),
		entity.InvitedBy,
		entity.InvitedAt,
		entity.CreatedAt,
		entity.UpdatedAt,
	)
}
