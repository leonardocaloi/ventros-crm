package contact_list

import (
	"context"

	"github.com/caloi/ventros-crm/internal/domain/contact_list"
	"github.com/google/uuid"
)

type DeleteContactListRequest struct {
	ContactListID uuid.UUID
}

type DeleteContactListUseCase struct {
	repo contact_list.Repository
}

func NewDeleteContactListUseCase(repo contact_list.Repository) *DeleteContactListUseCase {
	return &DeleteContactListUseCase{repo: repo}
}

func (uc *DeleteContactListUseCase) Execute(ctx context.Context, req DeleteContactListRequest) error {
	// Buscar lista para validar que existe
	list, err := uc.repo.FindByID(ctx, req.ContactListID)
	if err != nil {
		return err
	}

	// Marcar como deletada
	list.Delete()

	// Persistir
	return uc.repo.Update(ctx, list)
}
