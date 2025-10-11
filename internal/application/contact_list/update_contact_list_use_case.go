package contact_list

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/crm/contact_list"
	"github.com/google/uuid"
)

type UpdateContactListRequest struct {
	ContactListID   uuid.UUID
	Name            *string
	Description     *string
	LogicalOperator *contact_list.LogicalOperator
	FilterRules     *[]FilterRuleRequest
}

type UpdateContactListUseCase struct {
	repo contact_list.Repository
}

func NewUpdateContactListUseCase(repo contact_list.Repository) *UpdateContactListUseCase {
	return &UpdateContactListUseCase{repo: repo}
}

func (uc *UpdateContactListUseCase) Execute(ctx context.Context, req UpdateContactListRequest) error {
	// Buscar lista existente
	list, err := uc.repo.FindByID(ctx, req.ContactListID)
	if err != nil {
		return err
	}

	// Atualizar campos
	if req.Name != nil {
		if err := list.UpdateName(*req.Name); err != nil {
			return err
		}
	}

	if req.Description != nil {
		list.UpdateDescription(*req.Description)
	}

	if req.LogicalOperator != nil {
		if err := list.UpdateLogicalOperator(*req.LogicalOperator); err != nil {
			return err
		}
	}

	if req.FilterRules != nil {
		// Limpar regras antigas
		list.ClearFilterRules()

		// Adicionar novas regras
		for _, ruleReq := range *req.FilterRules {
			var rule *contact_list.FilterRule
			var err error

			switch ruleReq.FilterType {
			case contact_list.FilterTypeCustomField:
				if ruleReq.FieldType == nil {
					return errors.New("field_type is required for custom_field filters")
				}
				rule, err = contact_list.NewCustomFieldFilterRule(
					ruleReq.FieldKey,
					parseFieldType(*ruleReq.FieldType),
					ruleReq.Operator,
					ruleReq.Value,
				)
			case contact_list.FilterTypePipelineStatus:
				if ruleReq.PipelineID == nil {
					return errors.New("pipeline_id is required for pipeline_status filters")
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
				return err
			}

			if err := list.AddFilterRule(rule); err != nil {
				return err
			}
		}
	}

	// Persistir
	return uc.repo.Update(ctx, list)
}
