package contact

import (
	"context"
	"errors"
	"fmt"

	"github.com/caloi/ventros-crm/internal/application/shared"
	"github.com/caloi/ventros-crm/internal/domain/crm/contact"
	"github.com/caloi/ventros-crm/internal/domain/crm/pipeline"
	"github.com/google/uuid"
)

// ChangePipelineStatusUseCase gerencia a mudança de status de um contato em um pipeline.
type ChangePipelineStatusUseCase struct {
	contactRepo  contact.Repository
	pipelineRepo pipeline.Repository
	eventBus     EventBus
	txManager    shared.TransactionManager
}

// NewChangePipelineStatusUseCase cria uma nova instância do use case.
func NewChangePipelineStatusUseCase(
	contactRepo contact.Repository,
	pipelineRepo pipeline.Repository,
	eventBus EventBus,
	txManager shared.TransactionManager,
) *ChangePipelineStatusUseCase {
	return &ChangePipelineStatusUseCase{
		contactRepo:  contactRepo,
		pipelineRepo: pipelineRepo,
		eventBus:     eventBus,
		txManager:    txManager,
	}
}

// ChangePipelineStatusInput representa os dados de entrada para mudança de status.
type ChangePipelineStatusInput struct {
	ContactID  uuid.UUID
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	ChangedBy  *uuid.UUID // ID do agente/usuário que fez a mudança
	Reason     string     // Motivo da mudança (opcional)
	TenantID   string
	ProjectID  uuid.UUID
}

// ChangePipelineStatusOutput representa o resultado da operação.
type ChangePipelineStatusOutput struct {
	ContactID          uuid.UUID
	PipelineID         uuid.UUID
	PreviousStatusID   *uuid.UUID
	PreviousStatusName string
	NewStatusID        uuid.UUID
	NewStatusName      string
	ChangedAt          string
}

// Execute executa o caso de uso de mudança de status.
func (uc *ChangePipelineStatusUseCase) Execute(ctx context.Context, input ChangePipelineStatusInput) (*ChangePipelineStatusOutput, error) {
	// 1. Validações básicas
	if input.ContactID == uuid.Nil {
		return nil, errors.New("contact_id is required")
	}
	if input.PipelineID == uuid.Nil {
		return nil, errors.New("pipeline_id is required")
	}
	if input.StatusID == uuid.Nil {
		return nil, errors.New("status_id is required")
	}
	if input.TenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if input.ProjectID == uuid.Nil {
		return nil, errors.New("project_id is required")
	}

	// 2. Busca o contato
	existingContact, err := uc.contactRepo.FindByID(ctx, input.ContactID)
	if err != nil {
		return nil, fmt.Errorf("failed to find contact: %w", err)
	}
	if existingContact == nil {
		return nil, errors.New("contact not found")
	}

	// 3. Verifica se o contato pertence ao projeto correto
	if existingContact.ProjectID() != input.ProjectID {
		return nil, errors.New("contact does not belong to this project")
	}

	// 4. Busca o pipeline com seus statuses
	pipelineEntity, statuses, err := uc.pipelineRepo.GetPipelineWithStatuses(ctx, input.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to find pipeline: %w", err)
	}
	if pipelineEntity == nil {
		return nil, errors.New("pipeline not found")
	}

	// 5. Verifica se o pipeline pertence ao projeto correto
	if pipelineEntity.ProjectID() != input.ProjectID {
		return nil, errors.New("pipeline does not belong to this project")
	}

	// 6. Verifica se o pipeline está ativo
	if !pipelineEntity.IsActive() {
		return nil, errors.New("pipeline is not active")
	}

	// 7. Verifica se o status existe no pipeline
	var newStatus *pipeline.Status
	for _, s := range statuses {
		if s.ID() == input.StatusID {
			newStatus = s
			break
		}
	}
	if newStatus == nil {
		return nil, errors.New("status not found in pipeline")
	}

	// 8. Verifica se o status está ativo
	if !newStatus.IsActiveStatus() {
		return nil, errors.New("status is not active")
	}

	// 9. Busca o status atual do contato no pipeline (se existir)
	var previousStatusID *uuid.UUID
	var previousStatusName string

	currentStatus, err := uc.pipelineRepo.GetContactStatus(ctx, input.ContactID, input.PipelineID)
	if err == nil && currentStatus != nil {
		statusID := currentStatus.ID()
		previousStatusID = &statusID
		previousStatusName = currentStatus.Name()

		// Verifica se já está no status desejado
		if currentStatus.ID() == input.StatusID {
			return nil, errors.New("contact is already in this status")
		}
	}

	// 10-11. ✅ TRANSAÇÃO ATÔMICA: SetContactStatus + Publish juntos
	var event contact.ContactPipelineStatusChangedEvent
	err = uc.txManager.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// 10. Atualiza o status do contato no pipeline (usa transação do contexto)
		err = uc.pipelineRepo.SetContactStatus(txCtx, input.ContactID, input.PipelineID, input.StatusID)
		if err != nil {
			return fmt.Errorf("failed to set contact status: %w", err)
		}

		// 11. Cria e publica o evento de domínio (usa mesma transação)
		event = contact.NewContactPipelineStatusChangedEvent(
			input.ContactID,
			input.PipelineID,
			previousStatusID,
			input.StatusID,
			previousStatusName,
			newStatus.Name(),
			input.TenantID,
			input.ProjectID,
			input.ChangedBy,
			input.Reason,
		)

		if err := uc.eventBus.Publish(txCtx, &event); err != nil {
			return fmt.Errorf("failed to publish event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 12. Retorna o resultado
	return &ChangePipelineStatusOutput{
		ContactID:          input.ContactID,
		PipelineID:         input.PipelineID,
		PreviousStatusID:   previousStatusID,
		PreviousStatusName: previousStatusName,
		NewStatusID:        input.StatusID,
		NewStatusName:      newStatus.Name(),
		ChangedAt:          event.ChangedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
