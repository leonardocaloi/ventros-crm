package contact_event

import (
	"context"
	"fmt"

	"github.com/caloi/ventros-crm/internal/domain/crm/contact_event"
	"github.com/google/uuid"
)

// CreateContactEventCommand representa o comando para criar um evento de contato
type CreateContactEventCommand struct {
	ContactID         uuid.UUID
	SessionID         *uuid.UUID
	TenantID          string
	EventType         string
	Category          contact_event.Category
	Priority          contact_event.Priority
	Source            contact_event.Source
	Title             *string
	Description       *string
	Payload           map[string]interface{}
	Metadata          map[string]interface{}
	TriggeredBy       *uuid.UUID
	IntegrationSource *string
	IsRealtime        bool
	VisibleToClient   bool
	VisibleToAgent    bool
}

// CreateContactEventUseCase cria eventos de contato para a timeline
type CreateContactEventUseCase struct {
	repo contact_event.Repository
}

func NewCreateContactEventUseCase(repo contact_event.Repository) *CreateContactEventUseCase {
	return &CreateContactEventUseCase{
		repo: repo,
	}
}

// Execute cria e persiste um evento de contato
func (uc *CreateContactEventUseCase) Execute(ctx context.Context, cmd CreateContactEventCommand) (*contact_event.ContactEvent, error) {
	// Validações de negócio
	if cmd.ContactID == uuid.Nil {
		return nil, fmt.Errorf("contact_id is required")
	}
	if cmd.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if cmd.EventType == "" {
		return nil, fmt.Errorf("event_type is required")
	}

	// Criar aggregate root
	event, err := contact_event.NewContactEvent(
		cmd.ContactID,
		cmd.TenantID,
		cmd.EventType,
		cmd.Category,
		cmd.Priority,
		cmd.Source,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact event: %w", err)
	}

	// Configurar propriedades opcionais
	if cmd.SessionID != nil {
		if err := event.AttachToSession(*cmd.SessionID); err != nil {
			return nil, fmt.Errorf("failed to attach session: %w", err)
		}
	}

	if cmd.Title != nil {
		event.SetTitle(*cmd.Title)
	}

	if cmd.Description != nil {
		event.SetDescription(*cmd.Description)
	}

	// Adicionar payload
	if cmd.Payload != nil {
		for key, value := range cmd.Payload {
			event.AddPayloadField(key, value)
		}
	}

	// Adicionar metadata
	if cmd.Metadata != nil {
		for key, value := range cmd.Metadata {
			event.AddMetadataField(key, value)
		}
	}

	if cmd.TriggeredBy != nil {
		if err := event.SetTriggeredBy(*cmd.TriggeredBy); err != nil {
			return nil, fmt.Errorf("failed to set triggered_by: %w", err)
		}
	}

	if cmd.IntegrationSource != nil {
		event.SetIntegrationSource(*cmd.IntegrationSource)
	}

	// Configurar visibilidade e realtime
	event.SetRealtimeDelivery(cmd.IsRealtime)
	event.SetVisibility(cmd.VisibleToClient, cmd.VisibleToAgent)

	// Persistir no repositório
	if err := uc.repo.Save(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save contact event: %w", err)
	}

	return event, nil
}
