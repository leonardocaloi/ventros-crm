package contact_list

import (
	"context"

	"github.com/google/uuid"
)

// Repository define as operações de persistência para ContactList
type Repository interface {
	// Create cria uma nova lista de contatos
	Create(ctx context.Context, list *ContactList) error

	// Update atualiza uma lista de contatos existente
	Update(ctx context.Context, list *ContactList) error

	// Delete deleta uma lista de contatos (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID busca uma lista de contatos por ID
	FindByID(ctx context.Context, id uuid.UUID) (*ContactList, error)

	// FindByProjectID busca todas as listas de contatos de um projeto
	FindByProjectID(ctx context.Context, projectID uuid.UUID) ([]*ContactList, error)

	// FindByTenantID busca todas as listas de contatos de um tenant
	FindByTenantID(ctx context.Context, tenantID string) ([]*ContactList, error)

	// ListByProject lista as listas de contatos de um projeto com paginação
	ListByProject(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*ContactList, int, error)

	// GetContactsInList retorna os IDs dos contatos que atendem aos filtros da lista
	GetContactsInList(ctx context.Context, listID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error)

	// RecalculateContactCount recalcula o número de contatos na lista
	RecalculateContactCount(ctx context.Context, listID uuid.UUID) (int, error)

	// AddContactToStaticList adiciona um contato a uma lista estática
	AddContactToStaticList(ctx context.Context, listID, contactID uuid.UUID) error

	// RemoveContactFromStaticList remove um contato de uma lista estática
	RemoveContactFromStaticList(ctx context.Context, listID, contactID uuid.UUID) error

	// IsContactInList verifica se um contato está em uma lista
	IsContactInList(ctx context.Context, listID, contactID uuid.UUID) (bool, error)
}
