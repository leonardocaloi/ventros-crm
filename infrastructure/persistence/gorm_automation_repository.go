package persistence

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormAutomationRuleRepository implementa pipeline.AutomationRepository usando GORM
type GormAutomationRuleRepository struct {
	db *gorm.DB
}

// NewGormAutomationRuleRepository cria uma nova instância do repository
func NewGormAutomationRuleRepository(db *gorm.DB) pipeline.AutomationRepository {
	return &GormAutomationRuleRepository{db: db}
}

// Save salva ou atualiza uma regra
func (r *GormAutomationRuleRepository) Save(rule *pipeline.Automation) error {
	entity, err := r.toEntity(rule)
	if err != nil {
		return fmt.Errorf("failed to convert rule to entity: %w", err)
	}

	// Verifica se já existe
	var existing entities.AutomationEntity
	err = r.db.Where("id = ?", entity.ID).First(&existing).Error

	if err == nil {
		// Update
		return r.db.Model(&existing).Updates(entity).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return r.db.Create(entity).Error
	}

	return err
}

// FindByID busca uma regra por ID
func (r *GormAutomationRuleRepository) FindByID(id uuid.UUID) (*pipeline.Automation, error) {
	var entity entities.AutomationEntity
	err := r.db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&entity)
}

// FindByPipeline busca todas as regras de um pipeline
func (r *GormAutomationRuleRepository) FindByPipeline(pipelineID uuid.UUID) ([]*pipeline.Automation, error) {
	var entities []entities.AutomationEntity
	err := r.db.
		Where("pipeline_id = ?", pipelineID).
		Order("priority ASC, created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// FindByPipelineAndTrigger busca regras de um pipeline com trigger específico
func (r *GormAutomationRuleRepository) FindByPipelineAndTrigger(
	pipelineID uuid.UUID,
	trigger pipeline.AutomationTrigger,
) ([]*pipeline.Automation, error) {
	var entities []entities.AutomationEntity
	err := r.db.
		Where("pipeline_id = ? AND trigger = ?", pipelineID, string(trigger)).
		Order("priority ASC, created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// FindEnabledByPipeline busca apenas regras ativas de um pipeline
func (r *GormAutomationRuleRepository) FindEnabledByPipeline(pipelineID uuid.UUID) ([]*pipeline.Automation, error) {
	var entities []entities.AutomationEntity
	err := r.db.
		Where("pipeline_id = ? AND enabled = ?", pipelineID, true).
		Order("priority ASC, created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}

// Delete remove uma regra
func (r *GormAutomationRuleRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&entities.AutomationEntity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Conversores

func (r *GormAutomationRuleRepository) toEntity(rule *pipeline.Automation) (*entities.AutomationEntity, error) {
	if rule == nil {
		return nil, errors.New("rule cannot be nil")
	}

	// Serializa conditions
	conditionsJSON, err := json.Marshal(rule.Conditions())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conditions: %w", err)
	}

	// Serializa actions
	actionsJSON, err := json.Marshal(rule.Actions())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal actions: %w", err)
	}

	return &entities.AutomationEntity{
		ID:             rule.ID(),
		AutomationType: string(rule.Type()),
		PipelineID:     rule.PipelineID(),
		TenantID:       rule.TenantID(),
		Name:           rule.Name(),
		Description:    rule.Description(),
		Trigger:        string(rule.Trigger()),
		Conditions:     conditionsJSON,
		Actions:        actionsJSON,
		Priority:       rule.Priority(),
		Enabled:        rule.IsEnabled(),
		CreatedAt:      rule.CreatedAt(),
		UpdatedAt:      rule.UpdatedAt(),
	}, nil
}

func (r *GormAutomationRuleRepository) toDomain(entity *entities.AutomationEntity) (*pipeline.Automation, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	// Deserializa conditions
	var conditions []pipeline.RuleCondition
	if len(entity.Conditions) > 0 {
		if err := json.Unmarshal(entity.Conditions, &conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	}

	// Deserializa actions
	var actions []pipeline.RuleAction
	if len(entity.Actions) > 0 {
		if err := json.Unmarshal(entity.Actions, &actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}
	}

	// Reconstrói o domain model
	return pipeline.ReconstructAutomation(
		entity.ID,
		pipeline.AutomationType(entity.AutomationType),
		entity.PipelineID,
		entity.TenantID,
		entity.Name,
		entity.Description,
		pipeline.AutomationTrigger(entity.Trigger),
		conditions,
		actions,
		entity.Priority,
		entity.Enabled,
		entity.CreatedAt,
		entity.UpdatedAt,
	), nil
}

func (r *GormAutomationRuleRepository) toDomainSlice(entities []entities.AutomationEntity) ([]*pipeline.Automation, error) {
	rules := make([]*pipeline.Automation, 0, len(entities))
	for _, entity := range entities {
		rule, err := r.toDomain(&entity)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// FindEnabledByPipelineAndTrigger busca regras ativas com trigger específico
// Método auxiliar útil para otimizar queries do engine
func (r *GormAutomationRuleRepository) FindEnabledByPipelineAndTrigger(
	pipelineID uuid.UUID,
	trigger pipeline.AutomationTrigger,
) ([]*pipeline.Automation, error) {
	var entities []entities.AutomationEntity
	err := r.db.
		Where("pipeline_id = ? AND trigger = ? AND enabled = ?", pipelineID, string(trigger), true).
		Order("priority ASC, created_at ASC").
		Find(&entities).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainSlice(entities)
}
