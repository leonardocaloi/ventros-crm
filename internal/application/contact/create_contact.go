package contact

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/application/shared"
	"github.com/ventros/crm/internal/domain/crm/contact"
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
	txManager   shared.TransactionManager
}

// NewCreateContactUseCase cria uma nova instância.
func NewCreateContactUseCase(
	contactRepo contact.Repository,
	eventBus EventBus,
	txManager shared.TransactionManager,
) *CreateContactUseCase {
	return &CreateContactUseCase{
		contactRepo: contactRepo,
		eventBus:    eventBus,
		txManager:   txManager,
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

	// Verificar duplicatas (FORA da transação - query read-only)
	if cmd.Phone != nil {
		existing, err := uc.contactRepo.FindByPhone(ctx, cmd.ProjectID, *cmd.Phone)
		if err == nil && existing != nil {
			return nil, errors.New("contact with this phone already exists")
		}
	}

	// ✅ TRANSAÇÃO ATÔMICA: Save + Publish juntos
	// Se qualquer operação falhar, tudo é revertido (rollback)
	var contactID uuid.UUID
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 1. Persistir contato (usa transação do contexto)
		if err := uc.contactRepo.Save(txCtx, newContact); err != nil {
			return fmt.Errorf("failed to save contact: %w", err)
		}

		// 2. Publicar eventos no outbox (usa mesma transação)
		for _, event := range newContact.DomainEvents() {
			if err := uc.eventBus.Publish(txCtx, event); err != nil {
				return fmt.Errorf("failed to publish event %s: %w", event.EventName(), err)
			}
		}

		contactID = newContact.ID()
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Limpar eventos após sucesso
	newContact.ClearEvents()

	return &CreateContactResult{
		ContactID: contactID,
	}, nil
}
