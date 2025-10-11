package contact_list

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/crm/contact_list"
	"github.com/google/uuid"
)

type CreateContactListRequest struct {
	ProjectID       uuid.UUID
	TenantID        string
	Name            string
	Description     *string
	LogicalOperator contact_list.LogicalOperator
	IsStatic        bool
	FilterRules     []FilterRuleRequest
}

type FilterRuleRequest struct {
	FilterType contact_list.FilterType
	Operator   contact_list.FilterOperator
	FieldKey   string
	FieldType  *string // Apenas para custom fields
	Value      interface{}
	PipelineID *uuid.UUID // Apenas para pipeline_status
}

type CreateContactListResponse struct {
	ContactListID uuid.UUID
}

type CreateContactListUseCase struct {
	repo contact_list.Repository
}

func NewCreateContactListUseCase(repo contact_list.Repository) *CreateContactListUseCase {
	return &CreateContactListUseCase{repo: repo}
}

func (uc *CreateContactListUseCase) Execute(ctx context.Context, req CreateContactListRequest) (*CreateContactListResponse, error) {
	// Validações
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.ProjectID == uuid.Nil {
		return nil, errors.New("project_id is required")
	}
	if req.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}

	// Criar lista
	list, err := contact_list.NewContactList(
		req.ProjectID,
		req.TenantID,
		req.Name,
		req.LogicalOperator,
		req.IsStatic,
	)
	if err != nil {
		return nil, err
	}

	// Adicionar descrição se fornecida
	if req.Description != nil && *req.Description != "" {
		list.UpdateDescription(*req.Description)
	}

	// Adicionar regras de filtro
	for _, ruleReq := range req.FilterRules {
		var rule *contact_list.FilterRule
		var err error

		switch ruleReq.FilterType {
		case contact_list.FilterTypeCustomField:
			if ruleReq.FieldType == nil {
				return nil, errors.New("field_type is required for custom_field filters")
			}
			// Converter string para FieldType
			rule, err = contact_list.NewCustomFieldFilterRule(
				ruleReq.FieldKey,
				parseFieldType(*ruleReq.FieldType),
				ruleReq.Operator,
				ruleReq.Value,
			)
		case contact_list.FilterTypePipelineStatus:
			if ruleReq.PipelineID == nil {
				return nil, errors.New("pipeline_id is required for pipeline_status filters")
			}
			rule, err = contact_list.NewPipelineStatusFilterRule(
				*ruleReq.PipelineID,
				ruleReq.FieldKey,
				ruleReq.Operator,
			)
		case contact_list.FilterTypeTag:
			rule, err = contact_list.NewTagFilterRule(ruleReq.Operator, ruleReq.Value)
		case contact_list.FilterTypeAttribute:
			rule, err = contact_list.NewAttributeFilterRule(ruleReq.FieldKey, ruleReq.Operator, ruleReq.Value)
		default:
			rule, err = contact_list.NewFilterRule(ruleReq.FilterType, ruleReq.Operator, ruleReq.FieldKey, ruleReq.Value)
		}

		if err != nil {
			return nil, err
		}

		if err := list.AddFilterRule(rule); err != nil {
			return nil, err
		}
	}

	// Persistir
	if err := uc.repo.Create(ctx, list); err != nil {
		return nil, err
	}

	return &CreateContactListResponse{
		ContactListID: list.ID(),
	}, nil
}
