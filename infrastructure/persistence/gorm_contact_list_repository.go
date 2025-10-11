package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/persistence/entities"
	"github.com/caloi/ventros-crm/internal/domain/core/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact_list"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormContactListRepository struct {
	db *gorm.DB
}

func NewGormContactListRepository(db *gorm.DB) contact_list.Repository {
	return &GormContactListRepository{db: db}
}

func (r *GormContactListRepository) Create(ctx context.Context, list *contact_list.ContactList) error {
	entity := r.domainToEntity(list)

	// Criar dentro de uma transação para incluir as regras de filtro
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}

		// Criar regras de filtro
		for _, rule := range list.FilterRules() {
			ruleEntity := r.filterRuleToEntity(rule, list.ID())
			if err := tx.Create(ruleEntity).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormContactListRepository) Update(ctx context.Context, list *contact_list.ContactList) error {
	entity := r.domainToEntity(list)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atualizar lista
		if err := tx.Updates(entity).Error; err != nil {
			return err
		}

		// Deletar regras antigas e criar novas
		if err := tx.Where("contact_list_id = ?", list.ID()).Delete(&entities.ContactListFilterRuleEntity{}).Error; err != nil {
			return err
		}

		// Criar novas regras
		for _, rule := range list.FilterRules() {
			ruleEntity := r.filterRuleToEntity(rule, list.ID())
			if err := tx.Create(ruleEntity).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormContactListRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.ContactListEntity{}, "id = ?", id).Error
}

func (r *GormContactListRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact_list.ContactList, error) {
	var entity entities.ContactListEntity
	err := r.db.WithContext(ctx).
		Preload("FilterRules").
		First(&entity, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("contact list not found")
		}
		return nil, err
	}

	return r.entityToDomain(&entity)
}

func (r *GormContactListRepository) FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*contact_list.ContactList, error) {
	var listEntities []entities.ContactListEntity
	err := r.db.WithContext(ctx).
		Preload("FilterRules").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Find(&listEntities).Error

	if err != nil {
		return nil, err
	}

	lists := make([]*contact_list.ContactList, 0, len(listEntities))
	for _, entity := range listEntities {
		list, err := r.entityToDomain(&entity)
		if err != nil {
			continue
		}
		lists = append(lists, list)
	}

	return lists, nil
}

func (r *GormContactListRepository) FindByTenantID(ctx context.Context, tenantID string) ([]*contact_list.ContactList, error) {
	var listEntities []entities.ContactListEntity
	err := r.db.WithContext(ctx).
		Preload("FilterRules").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Find(&listEntities).Error

	if err != nil {
		return nil, err
	}

	lists := make([]*contact_list.ContactList, 0, len(listEntities))
	for _, entity := range listEntities {
		list, err := r.entityToDomain(&entity)
		if err != nil {
			continue
		}
		lists = append(lists, list)
	}

	return lists, nil
}

func (r *GormContactListRepository) ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*contact_list.ContactList, int, error) {
	var listEntities []entities.ContactListEntity
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&entities.ContactListEntity{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := r.db.WithContext(ctx).
		Preload("FilterRules").
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&listEntities).Error; err != nil {
		return nil, 0, err
	}

	lists := make([]*contact_list.ContactList, 0, len(listEntities))
	for _, entity := range listEntities {
		list, err := r.entityToDomain(&entity)
		if err != nil {
			continue
		}
		lists = append(lists, list)
	}

	return lists, int(total), nil
}

func (r *GormContactListRepository) GetContactsInList(ctx context.Context, listID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	// Buscar a lista
	list, err := r.FindByID(ctx, listID)
	if err != nil {
		return nil, 0, err
	}

	// Se for lista estática, buscar membros diretos
	if list.IsStatic() {
		return r.getStaticListContacts(ctx, listID, limit, offset)
	}

	// Se for lista dinâmica, aplicar filtros
	return r.getDynamicListContacts(ctx, list, limit, offset)
}

func (r *GormContactListRepository) getStaticListContacts(ctx context.Context, listID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	var total int64
	var members []entities.ContactListMemberEntity

	// Count
	if err := r.db.WithContext(ctx).Model(&entities.ContactListMemberEntity{}).
		Where("contact_list_id = ?", listID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get members
	query := r.db.WithContext(ctx).
		Where("contact_list_id = ?", listID).
		Order("added_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&members).Error; err != nil {
		return nil, 0, err
	}

	contactIDs := make([]uuid.UUID, len(members))
	for i, member := range members {
		contactIDs[i] = member.ContactID
	}

	return contactIDs, int(total), nil
}

func (r *GormContactListRepository) getDynamicListContacts(ctx context.Context, list *contact_list.ContactList, limit, offset int) ([]uuid.UUID, int, error) {
	// Construir query com base nas regras de filtro
	query := r.db.WithContext(ctx).Model(&entities.ContactEntity{}).
		Where("deleted_at IS NULL AND project_id = ?", list.ProjectID())

	// Aplicar filtros
	for _, rule := range list.FilterRules() {
		query = r.applyFilterRule(query, rule, list.LogicalOperator())
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get IDs
	var contacts []struct {
		ID uuid.UUID
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Select("id").Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	contactIDs := make([]uuid.UUID, len(contacts))
	for i, c := range contacts {
		contactIDs[i] = c.ID
	}

	return contactIDs, int(total), nil
}

func (r *GormContactListRepository) applyFilterRule(query *gorm.DB, rule *contact_list.FilterRule, logicalOp contact_list.LogicalOperator) *gorm.DB {
	// Implementar lógica de filtro baseada no tipo e operador
	switch rule.FilterType() {
	case contact_list.FilterTypeAttribute:
		return r.applyAttributeFilter(query, rule, logicalOp)
	case contact_list.FilterTypeTag:
		return r.applyTagFilter(query, rule, logicalOp)
	case contact_list.FilterTypeCustomField:
		return r.applyCustomFieldFilter(query, rule, logicalOp)
	case contact_list.FilterTypePipelineStatus:
		return r.applyPipelineStatusFilter(query, rule, logicalOp)
	default:
		return query
	}
}

func (r *GormContactListRepository) applyAttributeFilter(query *gorm.DB, rule *contact_list.FilterRule, logicalOp contact_list.LogicalOperator) *gorm.DB {
	fieldKey := rule.FieldKey()
	operator := rule.Operator()
	value := rule.Value()

	condition := ""
	args := []interface{}{}

	switch operator {
	case contact_list.OperatorEquals:
		condition = fmt.Sprintf("%s = ?", fieldKey)
		args = append(args, value)
	case contact_list.OperatorNotEquals:
		condition = fmt.Sprintf("%s != ?", fieldKey)
		args = append(args, value)
	case contact_list.OperatorGreaterThan:
		condition = fmt.Sprintf("%s > ?", fieldKey)
		args = append(args, value)
	case contact_list.OperatorLessThan:
		condition = fmt.Sprintf("%s < ?", fieldKey)
		args = append(args, value)
	case contact_list.OperatorContains:
		condition = fmt.Sprintf("%s ILIKE ?", fieldKey)
		args = append(args, fmt.Sprintf("%%%v%%", value))
	case contact_list.OperatorStartsWith:
		condition = fmt.Sprintf("%s ILIKE ?", fieldKey)
		args = append(args, fmt.Sprintf("%v%%", value))
	case contact_list.OperatorIsNull:
		condition = fmt.Sprintf("%s IS NULL", fieldKey)
	case contact_list.OperatorIsNotNull:
		condition = fmt.Sprintf("%s IS NOT NULL", fieldKey)
	}

	if logicalOp == contact_list.LogicalOperatorAND {
		return query.Where(condition, args...)
	}
	return query.Or(condition, args...)
}

func (r *GormContactListRepository) applyTagFilter(query *gorm.DB, rule *contact_list.FilterRule, logicalOp contact_list.LogicalOperator) *gorm.DB {
	operator := rule.Operator()
	value := rule.Value()

	condition := ""
	args := []interface{}{}

	switch operator {
	case contact_list.OperatorContains:
		// tags é jsonb, usar operador @>
		condition = "tags @> ?::jsonb"
		tagJSON, _ := json.Marshal([]string{fmt.Sprintf("%v", value)})
		args = append(args, string(tagJSON))
	case contact_list.OperatorIn:
		// value deve ser um array
		if tags, ok := value.([]string); ok {
			tagJSON, _ := json.Marshal(tags)
			condition = "tags @> ?::jsonb"
			args = append(args, string(tagJSON))
		}
	}

	if condition != "" {
		if logicalOp == contact_list.LogicalOperatorAND {
			return query.Where(condition, args...)
		}
		return query.Or(condition, args...)
	}

	return query
}

func (r *GormContactListRepository) applyCustomFieldFilter(query *gorm.DB, rule *contact_list.FilterRule, logicalOp contact_list.LogicalOperator) *gorm.DB {
	// Filtrar por custom fields usando subquery
	fieldKey := rule.FieldKey()
	operator := rule.Operator()
	value := rule.Value()

	subQuery := r.db.Table("contact_custom_fields").
		Select("contact_id").
		Where("field_key = ?", fieldKey)

	switch operator {
	case contact_list.OperatorEquals:
		subQuery = subQuery.Where("field_value = ?::jsonb", fmt.Sprintf(`"%v"`, value))
	case contact_list.OperatorContains:
		subQuery = subQuery.Where("field_value::text ILIKE ?", fmt.Sprintf("%%%v%%", value))
	}

	if logicalOp == contact_list.LogicalOperatorAND {
		return query.Where("id IN (?)", subQuery)
	}
	return query.Or("id IN (?)", subQuery)
}

func (r *GormContactListRepository) applyPipelineStatusFilter(query *gorm.DB, rule *contact_list.FilterRule, logicalOp contact_list.LogicalOperator) *gorm.DB {
	if rule.PipelineID() == nil {
		return query
	}

	statusName := rule.Value()
	pipelineID := *rule.PipelineID()

	subQuery := r.db.Table("contact_pipeline_statuses").
		Select("contact_id").
		Where("pipeline_id = ? AND current_status = ?", pipelineID, statusName)

	if logicalOp == contact_list.LogicalOperatorAND {
		return query.Where("id IN (?)", subQuery)
	}
	return query.Or("id IN (?)", subQuery)
}

func (r *GormContactListRepository) RecalculateContactCount(ctx context.Context, listID uuid.UUID) (int, error) {
	contactIDs, total, err := r.GetContactsInList(ctx, listID, 0, 0)
	if err != nil {
		return 0, err
	}

	_ = contactIDs // contactIDs não usado, mas mantido para manter interface consistente

	// Atualizar contador na lista
	now := time.Now()
	err = r.db.WithContext(ctx).Model(&entities.ContactListEntity{}).
		Where("id = ?", listID).
		Updates(map[string]interface{}{
			"contact_count":      total,
			"last_calculated_at": now,
		}).Error

	return total, err
}

func (r *GormContactListRepository) AddContactToStaticList(ctx context.Context, listID, contactID uuid.UUID) error {
	member := &entities.ContactListMemberEntity{
		ID:            uuid.New(),
		ContactListID: listID,
		ContactID:     contactID,
		AddedAt:       time.Now(),
	}

	return r.db.WithContext(ctx).Create(member).Error
}

func (r *GormContactListRepository) RemoveContactFromStaticList(ctx context.Context, listID, contactID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("contact_list_id = ? AND contact_id = ?", listID, contactID).
		Delete(&entities.ContactListMemberEntity{}).Error
}

func (r *GormContactListRepository) IsContactInList(ctx context.Context, listID, contactID uuid.UUID) (bool, error) {
	list, err := r.FindByID(ctx, listID)
	if err != nil {
		return false, err
	}

	if list.IsStatic() {
		var count int64
		err := r.db.WithContext(ctx).Model(&entities.ContactListMemberEntity{}).
			Where("contact_list_id = ? AND contact_id = ?", listID, contactID).
			Count(&count).Error
		return count > 0, err
	}

	// Para listas dinâmicas, verificar se o contato atende aos filtros
	contactIDs, _, err := r.GetContactsInList(ctx, listID, 0, 0)
	if err != nil {
		return false, err
	}

	for _, id := range contactIDs {
		if id == contactID {
			return true, nil
		}
	}

	return false, nil
}

// Mappers
func (r *GormContactListRepository) domainToEntity(list *contact_list.ContactList) *entities.ContactListEntity {
	entity := &entities.ContactListEntity{
		ID:               list.ID(),
		ProjectID:        list.ProjectID(),
		TenantID:         list.TenantID(),
		Name:             list.Name(),
		LogicalOperator:  string(list.LogicalOperator()),
		IsStatic:         list.IsStatic(),
		ContactCount:     list.ContactCount(),
		LastCalculatedAt: list.LastCalculatedAt(),
		CreatedAt:        list.CreatedAt(),
		UpdatedAt:        list.UpdatedAt(),
	}

	if desc := list.Description(); desc != nil {
		entity.Description = *desc
	}

	if deletedAt := list.DeletedAt(); deletedAt != nil {
		entity.DeletedAt = gorm.DeletedAt{Time: *deletedAt, Valid: true}
	}

	return entity
}

func (r *GormContactListRepository) filterRuleToEntity(rule *contact_list.FilterRule, listID uuid.UUID) *entities.ContactListFilterRuleEntity {
	valueJSON, _ := json.Marshal(rule.Value())

	entity := &entities.ContactListFilterRuleEntity{
		ID:            rule.ID(),
		ContactListID: listID,
		FilterType:    string(rule.FilterType()),
		Operator:      string(rule.Operator()),
		FieldKey:      rule.FieldKey(),
		Value:         string(valueJSON),
		PipelineID:    rule.PipelineID(),
		CreatedAt:     rule.CreatedAt(),
	}

	if fieldType := rule.FieldType(); fieldType != nil {
		entity.FieldType = string(*fieldType)
	}

	return entity
}

func (r *GormContactListRepository) entityToDomain(entity *entities.ContactListEntity) (*contact_list.ContactList, error) {
	var description *string
	if entity.Description != "" {
		description = &entity.Description
	}

	var deletedAt *time.Time
	if entity.DeletedAt.Valid {
		deletedAt = &entity.DeletedAt.Time
	}

	// Converter regras de filtro
	filterRules := make([]*contact_list.FilterRule, 0, len(entity.FilterRules))
	for _, ruleEntity := range entity.FilterRules {
		rule, err := r.filterRuleEntityToDomain(&ruleEntity)
		if err != nil {
			continue // Skip regras inválidas
		}
		filterRules = append(filterRules, rule)
	}

	logicalOp := contact_list.LogicalOperator(entity.LogicalOperator)

	return contact_list.ReconstructContactList(
		entity.ID,
		entity.ProjectID,
		entity.TenantID,
		entity.Name,
		description,
		filterRules,
		logicalOp,
		entity.IsStatic,
		entity.ContactCount,
		entity.LastCalculatedAt,
		entity.CreatedAt,
		entity.UpdatedAt,
		deletedAt,
	), nil
}

func (r *GormContactListRepository) filterRuleEntityToDomain(entity *entities.ContactListFilterRuleEntity) (*contact_list.FilterRule, error) {
	var value interface{}
	if err := json.Unmarshal([]byte(entity.Value), &value); err != nil {
		value = entity.Value // Fallback para string
	}

	var fieldType *shared.FieldType
	if entity.FieldType != "" {
		ft := shared.FieldType(entity.FieldType)
		fieldType = &ft
	}

	return contact_list.ReconstructFilterRule(
		entity.ID,
		contact_list.FilterType(entity.FilterType),
		contact_list.FilterOperator(entity.Operator),
		entity.FieldKey,
		fieldType,
		value,
		entity.PipelineID,
		entity.CreatedAt,
	), nil
}
