package pipeline

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent é a interface base para eventos de domínio
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// Pipeline Events

// PipelineCreatedEvent - Pipeline criado
type PipelineCreatedEvent struct {
	PipelineID uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
	CreatedAt  time.Time
}

func (e PipelineCreatedEvent) EventName() string     { return "pipeline.created" }
func (e PipelineCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// PipelineUpdatedEvent - Pipeline atualizado
type PipelineUpdatedEvent struct {
	PipelineID uuid.UUID
	Field      string
	OldValue   interface{}
	NewValue   interface{}
	UpdatedAt  time.Time
}

func (e PipelineUpdatedEvent) EventName() string     { return "pipeline.updated" }
func (e PipelineUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

// PipelineActivatedEvent - Pipeline ativado
type PipelineActivatedEvent struct {
	PipelineID  uuid.UUID
	ActivatedAt time.Time
}

func (e PipelineActivatedEvent) EventName() string     { return "pipeline.activated" }
func (e PipelineActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

// PipelineDeactivatedEvent - Pipeline desativado
type PipelineDeactivatedEvent struct {
	PipelineID    uuid.UUID
	DeactivatedAt time.Time
}

func (e PipelineDeactivatedEvent) EventName() string     { return "pipeline.deactivated" }
func (e PipelineDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

// Status Events

// StatusCreatedEvent - Status criado
type StatusCreatedEvent struct {
	StatusID   uuid.UUID
	PipelineID uuid.UUID
	Name       string
	StatusType StatusType
	CreatedAt  time.Time
}

func (e StatusCreatedEvent) EventName() string     { return "status.created" }
func (e StatusCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// StatusUpdatedEvent - Status atualizado
type StatusUpdatedEvent struct {
	StatusID  uuid.UUID
	Field     string
	OldValue  interface{}
	NewValue  interface{}
	UpdatedAt time.Time
}

func (e StatusUpdatedEvent) EventName() string     { return "status.updated" }
func (e StatusUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

// StatusActivatedEvent - Status ativado
type StatusActivatedEvent struct {
	StatusID    uuid.UUID
	ActivatedAt time.Time
}

func (e StatusActivatedEvent) EventName() string     { return "status.activated" }
func (e StatusActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

// StatusDeactivatedEvent - Status desativado
type StatusDeactivatedEvent struct {
	StatusID      uuid.UUID
	DeactivatedAt time.Time
}

func (e StatusDeactivatedEvent) EventName() string     { return "status.deactivated" }
func (e StatusDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

// Pipeline-Status Relationship Events

// StatusAddedToPipelineEvent - Status adicionado ao pipeline
type StatusAddedToPipelineEvent struct {
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	AddedAt    time.Time
}

func (e StatusAddedToPipelineEvent) EventName() string     { return "pipeline.status_added" }
func (e StatusAddedToPipelineEvent) OccurredAt() time.Time { return e.AddedAt }

// StatusRemovedFromPipelineEvent - Status removido do pipeline
type StatusRemovedFromPipelineEvent struct {
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	RemovedAt  time.Time
}

func (e StatusRemovedFromPipelineEvent) EventName() string     { return "pipeline.status_removed" }
func (e StatusRemovedFromPipelineEvent) OccurredAt() time.Time { return e.RemovedAt }

// Contact Status Change Events

// ContactStatusChangedEvent - Status do contato alterado
type ContactStatusChangedEvent struct {
	ContactID     uuid.UUID
	PipelineID    uuid.UUID
	OldStatusID   *uuid.UUID
	NewStatusID   uuid.UUID
	OldStatusName *string
	NewStatusName string
	ChangedAt     time.Time
	ChangedBy     *uuid.UUID // ID do usuário que fez a mudança
	Reason        string     // Motivo da mudança
}

func (e ContactStatusChangedEvent) EventName() string     { return "contact.status_changed" }
func (e ContactStatusChangedEvent) OccurredAt() time.Time { return e.ChangedAt }

// ContactEnteredPipelineEvent - Contato entrou no pipeline
type ContactEnteredPipelineEvent struct {
	ContactID   uuid.UUID
	PipelineID  uuid.UUID
	StatusID    uuid.UUID
	StatusName  string
	EnteredAt   time.Time
	EnteredBy   *uuid.UUID // ID do usuário que adicionou
}

func (e ContactEnteredPipelineEvent) EventName() string     { return "contact.entered_pipeline" }
func (e ContactEnteredPipelineEvent) OccurredAt() time.Time { return e.EnteredAt }

// ContactExitedPipelineEvent - Contato saiu do pipeline
type ContactExitedPipelineEvent struct {
	ContactID    uuid.UUID
	PipelineID   uuid.UUID
	LastStatusID uuid.UUID
	LastStatusName string
	ExitedAt     time.Time
	ExitedBy     *uuid.UUID // ID do usuário que removeu
	Reason       string     // Motivo da saída
}

func (e ContactExitedPipelineEvent) EventName() string     { return "contact.exited_pipeline" }
func (e ContactExitedPipelineEvent) OccurredAt() time.Time { return e.ExitedAt }
