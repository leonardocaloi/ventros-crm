package contact

import (
	"context"
	"errors"

	"github.com/caloi/ventros-crm/internal/domain/contact"
	"github.com/google/uuid"
)

// CreateContactCommand contém os dados para criar um contato.
type CreateContactCommand struct {
	ProjectID uuid.UUID
	TenantID  string
	Name      string
	Email     *string
	Phone     *string
}

// CreateContactResult retorna o resultado da criação.
type CreateContactResult struct {
	ContactID uuid.UUID
}

// CreateContactUseCase implementa o caso de uso de criação de contato.
type CreateContactUseCase struct {
	contactRepo contact.Repository
	eventBus    EventBus
}

// NewCreateContactUseCase cria uma nova instância.
func NewCreateContactUseCase(
	contactRepo contact.Repository,
	eventBus EventBus,
) *CreateContactUseCase {
	return &CreateContactUseCase{
		contactRepo: contactRepo,
		eventBus:    eventBus,
	}
}

// Execute executa o caso de uso.
func (uc *CreateContactUseCase) Execute(ctx context.Context, cmd CreateContactCommand) (*CreateContactResult, error) {
	// Validação
	if cmd.ProjectID == uuid.Nil {
		return nil, errors.New("projectID is required")
	}
	if cmd.TenantID == "" {
		return nil, errors.New("tenantID is required")
	}
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}

	// Criar entidade de domínio
	newContact, err := contact.NewContact(cmd.ProjectID, cmd.TenantID, cmd.Name)
	if err != nil {
		return nil, err
	}

	// Adicionar email se fornecido
	if cmd.Email != nil && *cmd.Email != "" {
		if err := newContact.SetEmail(*cmd.Email); err != nil {
			return nil, err
		}
	}

	// Adicionar telefone se fornecido
	if cmd.Phone != nil && *cmd.Phone != "" {
		if err := newContact.SetPhone(*cmd.Phone); err != nil {
			return nil, err
		}
	}

	// Verificar duplicatas
	if cmd.Phone != nil {
		existing, err := uc.contactRepo.FindByPhone(ctx, cmd.ProjectID, *cmd.Phone)
		if err == nil && existing != nil {
			return nil, errors.New("contact with this phone already exists")
		}
	}

	// Persistir
	if err := uc.contactRepo.Save(ctx, newContact); err != nil {
		return nil, err
	}

	// Publicar eventos
	for _, event := range newContact.DomainEvents() {
		if err := uc.eventBus.Publish(ctx, event); err != nil {
			// Log error
		}
	}

	newContact.ClearEvents()

	return &CreateContactResult{
		ContactID: newContact.ID(),
	}, nil
}
