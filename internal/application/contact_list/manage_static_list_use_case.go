package contact_list

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/contact_list"
	"github.com/google/uuid"
)

type AddContactToListRequest struct {
	ContactListID uuid.UUID
	ContactID     uuid.UUID
}

type RemoveContactFromListRequest struct {
	ContactListID uuid.UUID
	ContactID     uuid.UUID
}

type ManageStaticListUseCase struct {
	repo contact_list.Repository
}

func NewManageStaticListUseCase(repo contact_list.Repository) *ManageStaticListUseCase {
	return &ManageStaticListUseCase{repo: repo}
}

func (uc *ManageStaticListUseCase) AddContact(ctx context.Context, req AddContactToListRequest) error {
	// Buscar lista
	list, err := uc.repo.FindByID(ctx, req.ContactListID)
	if err != nil {
		return err
	}

	// Validar que é uma lista estática
	if !list.IsStatic() {
		return errors.New("cannot manually add contacts to dynamic list")
	}

	// Verificar se já está na lista
	exists, err := uc.repo.IsContactInList(ctx, req.ContactListID, req.ContactID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("contact already in list")
	}

	// Adicionar contato
	return uc.repo.AddContactToStaticList(ctx, req.ContactListID, req.ContactID)
}

func (uc *ManageStaticListUseCase) RemoveContact(ctx context.Context, req RemoveContactFromListRequest) error {
	// Buscar lista
	list, err := uc.repo.FindByID(ctx, req.ContactListID)
	if err != nil {
		return err
	}

	// Validar que é uma lista estática
	if !list.IsStatic() {
		return errors.New("cannot manually remove contacts from dynamic list")
	}

	// Verificar se está na lista
	exists, err := uc.repo.IsContactInList(ctx, req.ContactListID, req.ContactID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("contact not in list")
	}

	// Remover contato
	return uc.repo.RemoveContactFromStaticList(ctx, req.ContactListID, req.ContactID)
}
