package contact_list

import (
	"context"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/contact_list"
)

type ListContactListsRequest struct {
	ProjectID uuid.UUID
	Limit     int
	Offset    int
}

type ContactListDTO struct {
	ID               uuid.UUID       `json:"id"`
	ProjectID        uuid.UUID       `json:"project_id"`
	TenantID         string          `json:"tenant_id"`
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	LogicalOperator  string          `json:"logical_operator"`
	IsStatic         bool            `json:"is_static"`
	ContactCount     int             `json:"contact_count"`
	LastCalculatedAt *string         `json:"last_calculated_at,omitempty"`
	FilterRules      []FilterRuleDTO `json:"filter_rules"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

type FilterRuleDTO struct {
	ID         uuid.UUID   `json:"id"`
	FilterType string      `json:"filter_type"`
	Operator   string      `json:"operator"`
	FieldKey   string      `json:"field_key"`
	FieldType  *string     `json:"field_type,omitempty"`
	Value      interface{} `json:"value"`
	PipelineID *uuid.UUID  `json:"pipeline_id,omitempty"`
}

type ListContactListsResponse struct {
	ContactLists []*ContactListDTO `json:"contact_lists"`
	Total        int               `json:"total"`
}

type ListContactListsUseCase struct {
	repo contact_list.Repository
}

func NewListContactListsUseCase(repo contact_list.Repository) *ListContactListsUseCase {
	return &ListContactListsUseCase{repo: repo}
}

func (uc *ListContactListsUseCase) Execute(ctx context.Context, req ListContactListsRequest) (*ListContactListsResponse, error) {
	lists, total, err := uc.repo.ListByProject(ctx, req.ProjectID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ContactListDTO, len(lists))
	for i, list := range lists {
		dtos[i] = uc.toDTO(list)
	}

	return &ListContactListsResponse{
		ContactLists: dtos,
		Total:        total,
	}, nil
}

func (uc *ListContactListsUseCase) toDTO(list *contact_list.ContactList) *ContactListDTO {
	dto := &ContactListDTO{
		ID:              list.ID(),
		ProjectID:       list.ProjectID(),
		TenantID:        list.TenantID(),
		Name:            list.Name(),
		Description:     list.Description(),
		LogicalOperator: string(list.LogicalOperator()),
		IsStatic:        list.IsStatic(),
		ContactCount:    list.ContactCount(),
		FilterRules:     make([]FilterRuleDTO, 0),
		CreatedAt:       list.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       list.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	if list.LastCalculatedAt() != nil {
		lastCalc := list.LastCalculatedAt().Format("2006-01-02T15:04:05Z07:00")
		dto.LastCalculatedAt = &lastCalc
	}

	for _, rule := range list.FilterRules() {
		ruleDTO := FilterRuleDTO{
			ID:         rule.ID(),
			FilterType: string(rule.FilterType()),
			Operator:   string(rule.Operator()),
			FieldKey:   rule.FieldKey(),
			Value:      rule.Value(),
			PipelineID: rule.PipelineID(),
		}

		if rule.FieldType() != nil {
			fieldTypeStr := string(*rule.FieldType())
			ruleDTO.FieldType = &fieldTypeStr
		}

		dto.FilterRules = append(dto.FilterRules, ruleDTO)
	}

	return dto
}
