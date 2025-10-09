package contact_list

import (
	"errors"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// FilterOperator define os operadores de filtro disponíveis
type FilterOperator string

const (
	OperatorEquals       FilterOperator = "eq"
	OperatorNotEquals    FilterOperator = "ne"
	OperatorGreaterThan  FilterOperator = "gt"
	OperatorLessThan     FilterOperator = "lt"
	OperatorGreaterEqual FilterOperator = "gte"
	OperatorLessEqual    FilterOperator = "lte"
	OperatorContains     FilterOperator = "contains"
	OperatorNotContains  FilterOperator = "not_contains"
	OperatorStartsWith   FilterOperator = "starts_with"
	OperatorEndsWith     FilterOperator = "ends_with"
	OperatorIn           FilterOperator = "in"
	OperatorNotIn        FilterOperator = "not_in"
	OperatorIsNull       FilterOperator = "is_null"
	OperatorIsNotNull    FilterOperator = "is_not_null"
)

// IsValid verifica se o operador é válido
func (fo FilterOperator) IsValid() bool {
	switch fo {
	case OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorLessThan,
		OperatorGreaterEqual, OperatorLessEqual, OperatorContains, OperatorNotContains,
		OperatorStartsWith, OperatorEndsWith, OperatorIn, OperatorNotIn,
		OperatorIsNull, OperatorIsNotNull:
		return true
	default:
		return false
	}
}

// FilterType define os tipos de filtro disponíveis
type FilterType string

const (
	FilterTypeCustomField    FilterType = "custom_field"
	FilterTypePipelineStatus FilterType = "pipeline_status"
	FilterTypeTag            FilterType = "tag"
	FilterTypeEvent          FilterType = "event"
	FilterTypeInteraction    FilterType = "interaction"
	FilterTypeAttribute      FilterType = "attribute" // name, email, phone, etc.
)

// IsValid verifica se o tipo de filtro é válido
func (ft FilterType) IsValid() bool {
	switch ft {
	case FilterTypeCustomField, FilterTypePipelineStatus, FilterTypeTag,
		FilterTypeEvent, FilterTypeInteraction, FilterTypeAttribute:
		return true
	default:
		return false
	}
}

// FilterRule representa uma regra de filtro
type FilterRule struct {
	id         uuid.UUID
	filterType FilterType
	operator   FilterOperator
	fieldKey   string            // Campo a ser filtrado (ex: "custom_field_name", "email", "tag")
	fieldType  *shared.FieldType // Tipo do campo (apenas para custom fields)
	value      interface{}       // Valor de comparação
	pipelineID *uuid.UUID        // Apenas para FilterTypePipelineStatus
	createdAt  time.Time
}

// NewFilterRule cria uma nova regra de filtro
func NewFilterRule(
	filterType FilterType,
	operator FilterOperator,
	fieldKey string,
	value interface{},
) (*FilterRule, error) {
	if !filterType.IsValid() {
		return nil, fmt.Errorf("invalid filter type: %s", filterType)
	}
	if !operator.IsValid() {
		return nil, fmt.Errorf("invalid filter operator: %s", operator)
	}
	if fieldKey == "" {
		return nil, errors.New("field key cannot be empty")
	}

	// Validações específicas
	switch operator {
	case OperatorIsNull, OperatorIsNotNull:
		if value != nil {
			return nil, errors.New("value must be nil for is_null/is_not_null operators")
		}
	case OperatorIn, OperatorNotIn:
		// Validar que value é um array
		if value == nil {
			return nil, errors.New("value cannot be nil for in/not_in operators")
		}
		// TODO: adicionar validação de tipo array
	default:
		if value == nil && operator != OperatorIsNull && operator != OperatorIsNotNull {
			return nil, errors.New("value cannot be nil for this operator")
		}
	}

	return &FilterRule{
		id:         uuid.New(),
		filterType: filterType,
		operator:   operator,
		fieldKey:   fieldKey,
		value:      value,
		createdAt:  time.Now(),
	}, nil
}

// NewCustomFieldFilterRule cria uma regra de filtro para campo customizado
func NewCustomFieldFilterRule(
	fieldKey string,
	fieldType shared.FieldType,
	operator FilterOperator,
	value interface{},
) (*FilterRule, error) {
	if !fieldType.IsValid() {
		return nil, fmt.Errorf("invalid field type: %s", fieldType)
	}

	rule, err := NewFilterRule(FilterTypeCustomField, operator, fieldKey, value)
	if err != nil {
		return nil, err
	}

	rule.fieldType = &fieldType
	return rule, nil
}

// NewPipelineStatusFilterRule cria uma regra de filtro para status de pipeline
func NewPipelineStatusFilterRule(
	pipelineID uuid.UUID,
	statusName string,
	operator FilterOperator,
) (*FilterRule, error) {
	if pipelineID == uuid.Nil {
		return nil, errors.New("pipeline ID cannot be nil")
	}

	rule, err := NewFilterRule(FilterTypePipelineStatus, operator, "status", statusName)
	if err != nil {
		return nil, err
	}

	rule.pipelineID = &pipelineID
	return rule, nil
}

// NewTagFilterRule cria uma regra de filtro para tags
func NewTagFilterRule(
	operator FilterOperator,
	tagValue interface{},
) (*FilterRule, error) {
	return NewFilterRule(FilterTypeTag, operator, "tag", tagValue)
}

// NewEventFilterRule cria uma regra de filtro para eventos
func NewEventFilterRule(
	eventType string,
	operator FilterOperator,
	value interface{},
) (*FilterRule, error) {
	return NewFilterRule(FilterTypeEvent, operator, eventType, value)
}

// NewAttributeFilterRule cria uma regra de filtro para atributos padrão (name, email, phone)
func NewAttributeFilterRule(
	attributeName string,
	operator FilterOperator,
	value interface{},
) (*FilterRule, error) {
	// Validar atributos permitidos
	allowedAttributes := map[string]bool{
		"name":                 true,
		"email":                true,
		"phone":                true,
		"language":             true,
		"timezone":             true,
		"source_channel":       true,
		"first_interaction_at": true,
		"last_interaction_at":  true,
		"created_at":           true,
		"updated_at":           true,
	}

	if !allowedAttributes[attributeName] {
		return nil, fmt.Errorf("invalid attribute name: %s", attributeName)
	}

	return NewFilterRule(FilterTypeAttribute, operator, attributeName, value)
}

// ReconstructFilterRule reconstrói uma regra de filtro a partir de dados persistidos
func ReconstructFilterRule(
	id uuid.UUID,
	filterType FilterType,
	operator FilterOperator,
	fieldKey string,
	fieldType *shared.FieldType,
	value interface{},
	pipelineID *uuid.UUID,
	createdAt time.Time,
) *FilterRule {
	return &FilterRule{
		id:         id,
		filterType: filterType,
		operator:   operator,
		fieldKey:   fieldKey,
		fieldType:  fieldType,
		value:      value,
		pipelineID: pipelineID,
		createdAt:  createdAt,
	}
}

// Getters
func (fr *FilterRule) ID() uuid.UUID                { return fr.id }
func (fr *FilterRule) FilterType() FilterType       { return fr.filterType }
func (fr *FilterRule) Operator() FilterOperator     { return fr.operator }
func (fr *FilterRule) FieldKey() string             { return fr.fieldKey }
func (fr *FilterRule) FieldType() *shared.FieldType { return fr.fieldType }
func (fr *FilterRule) Value() interface{}           { return fr.value }
func (fr *FilterRule) PipelineID() *uuid.UUID       { return fr.pipelineID }
func (fr *FilterRule) CreatedAt() time.Time         { return fr.createdAt }

// String retorna uma representação em string da regra
func (fr *FilterRule) String() string {
	return fmt.Sprintf("%s %s %s: %v", fr.filterType, fr.fieldKey, fr.operator, fr.value)
}
