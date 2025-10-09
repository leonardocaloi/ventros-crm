package contact

import (
	"time"

	"github.com/caloi/ventros-crm/internal/domain/shared"
	"github.com/google/uuid"
)

// ContactPipelineStatusChangedEvent é emitido quando um contato muda de status em um pipeline.
type ContactPipelineStatusChangedEvent struct {
	shared.BaseEvent
	ContactID          uuid.UUID
	PipelineID         uuid.UUID
	PreviousStatusID   *uuid.UUID // nil se for o primeiro status
	NewStatusID        uuid.UUID
	PreviousStatusName string
	NewStatusName      string
	TenantID           string
	ProjectID          uuid.UUID
	ChangedBy          *uuid.UUID // ID do usuário que fez a mudança (nil se automático)
	Reason             string     // Motivo da mudança (opcional)
	ChangedAt          time.Time
}

// NewContactPipelineStatusChangedEvent cria um novo evento de mudança de status.
func NewContactPipelineStatusChangedEvent(
	contactID uuid.UUID,
	pipelineID uuid.UUID,
	previousStatusID *uuid.UUID,
	newStatusID uuid.UUID,
	previousStatusName string,
	newStatusName string,
	tenantID string,
	projectID uuid.UUID,
	changedBy *uuid.UUID,
	reason string,
) ContactPipelineStatusChangedEvent {
	return ContactPipelineStatusChangedEvent{
		BaseEvent:          shared.NewBaseEvent("contact.pipeline_status_changed", time.Now()),
		ContactID:          contactID,
		PipelineID:         pipelineID,
		PreviousStatusID:   previousStatusID,
		NewStatusID:        newStatusID,
		PreviousStatusName: previousStatusName,
		NewStatusName:      newStatusName,
		TenantID:           tenantID,
		ProjectID:          projectID,
		ChangedBy:          changedBy,
		Reason:             reason,
		ChangedAt:          time.Now(),
	}
}

// IsFirstStatus retorna true se é a primeira vez que o contato entra no pipeline.
func (e ContactPipelineStatusChangedEvent) IsFirstStatus() bool {
	return e.PreviousStatusID == nil
}

// ToContactEventPayload converte o evento para payload de ContactEvent.
func (e ContactPipelineStatusChangedEvent) ToContactEventPayload() map[string]interface{} {
	payload := map[string]interface{}{
		"pipeline_id":     e.PipelineID.String(),
		"new_status_id":   e.NewStatusID.String(),
		"new_status_name": e.NewStatusName,
	}

	if e.PreviousStatusID != nil {
		payload["previous_status_id"] = e.PreviousStatusID.String()
		payload["previous_status_name"] = e.PreviousStatusName
	}

	if e.ChangedBy != nil {
		payload["changed_by"] = e.ChangedBy.String()
	}

	if e.Reason != "" {
		payload["reason"] = e.Reason
	}

	return payload
}

// GetTitle retorna um título legível para o evento.
func (e ContactPipelineStatusChangedEvent) GetTitle() string {
	if e.IsFirstStatus() {
		return "Entered pipeline: " + e.NewStatusName
	}
	return "Status changed: " + e.PreviousStatusName + " → " + e.NewStatusName
}

// GetDescription retorna uma descrição legível para o evento.
func (e ContactPipelineStatusChangedEvent) GetDescription() string {
	if e.IsFirstStatus() {
		return "Contact entered the pipeline with status: " + e.NewStatusName
	}
	return "Contact moved from " + e.PreviousStatusName + " to " + e.NewStatusName
}
