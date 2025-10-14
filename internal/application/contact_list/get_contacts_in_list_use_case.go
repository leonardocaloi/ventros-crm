package contact_list

import (
	"context"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/crm/contact_list"
)

type GetContactsInListRequest struct {
	ContactListID uuid.UUID
	Limit         int
	Offset        int
}

type GetContactsInListResponse struct {
	ContactIDs []uuid.UUID `json:"contact_ids"`
	Total      int         `json:"total"`
}

type GetContactsInListUseCase struct {
	repo contact_list.Repository
}

func NewGetContactsInListUseCase(repo contact_list.Repository) *GetContactsInListUseCase {
	return &GetContactsInListUseCase{repo: repo}
}

func (uc *GetContactsInListUseCase) Execute(ctx context.Context, req GetContactsInListRequest) (*GetContactsInListResponse, error) {
	contactIDs, total, err := uc.repo.GetContactsInList(ctx, req.ContactListID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &GetContactsInListResponse{
		ContactIDs: contactIDs,
		Total:      total,
	}, nil
}
