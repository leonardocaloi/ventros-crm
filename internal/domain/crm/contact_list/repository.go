package contact_list

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, list *ContactList) error

	Update(ctx context.Context, list *ContactList) error

	Delete(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*ContactList, error)

	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*ContactList, error)

	FindByTenantID(ctx context.Context, tenantID string) ([]*ContactList, error)

	ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*ContactList, int, error)

	GetContactsInList(ctx context.Context, listID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error)

	RecalculateContactCount(ctx context.Context, listID uuid.UUID) (int, error)

	AddContactToStaticList(ctx context.Context, listID, contactID uuid.UUID) error

	RemoveContactFromStaticList(ctx context.Context, listID, contactID uuid.UUID) error

	IsContactInList(ctx context.Context, listID, contactID uuid.UUID) (bool, error)
}
