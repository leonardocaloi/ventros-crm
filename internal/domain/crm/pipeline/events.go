package pipeline

import (
	"time"

	"github.com/google/uuid"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type PipelineCreatedEvent struct {
	PipelineID uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
	CreatedAt  time.Time
}

func (e PipelineCreatedEvent) EventName() string     { return "pipeline.created" }
func (e PipelineCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type PipelineUpdatedEvent struct {
	PipelineID uuid.UUID
	Field      string
	OldValue   interface{}
	NewValue   interface{}
	UpdatedAt  time.Time
}

func (e PipelineUpdatedEvent) EventName() string     { return "pipeline.updated" }
func (e PipelineUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

type PipelineActivatedEvent struct {
	PipelineID  uuid.UUID
	ActivatedAt time.Time
}

func (e PipelineActivatedEvent) EventName() string     { return "pipeline.activated" }
func (e PipelineActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

type PipelineDeactivatedEvent struct {
	PipelineID    uuid.UUID
	DeactivatedAt time.Time
}

func (e PipelineDeactivatedEvent) EventName() string     { return "pipeline.deactivated" }
func (e PipelineDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

type StatusCreatedEvent struct {
	StatusID   uuid.UUID
	PipelineID uuid.UUID
	Name       string
	StatusType StatusType
	CreatedAt  time.Time
}

func (e StatusCreatedEvent) EventName() string     { return "status.created" }
func (e StatusCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type StatusUpdatedEvent struct {
	StatusID  uuid.UUID
	Field     string
	OldValue  interface{}
	NewValue  interface{}
	UpdatedAt time.Time
}

func (e StatusUpdatedEvent) EventName() string     { return "status.updated" }
func (e StatusUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

type StatusActivatedEvent struct {
	StatusID    uuid.UUID
	ActivatedAt time.Time
}

func (e StatusActivatedEvent) EventName() string     { return "status.activated" }
func (e StatusActivatedEvent) OccurredAt() time.Time { return e.ActivatedAt }

type StatusDeactivatedEvent struct {
	StatusID      uuid.UUID
	DeactivatedAt time.Time
}

func (e StatusDeactivatedEvent) EventName() string     { return "status.deactivated" }
func (e StatusDeactivatedEvent) OccurredAt() time.Time { return e.DeactivatedAt }

type StatusAddedToPipelineEvent struct {
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	AddedAt    time.Time
}

func (e StatusAddedToPipelineEvent) EventName() string     { return "pipeline.status_added" }
func (e StatusAddedToPipelineEvent) OccurredAt() time.Time { return e.AddedAt }

type StatusRemovedFromPipelineEvent struct {
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	RemovedAt  time.Time
}

func (e StatusRemovedFromPipelineEvent) EventName() string     { return "pipeline.status_removed" }
func (e StatusRemovedFromPipelineEvent) OccurredAt() time.Time { return e.RemovedAt }

type ContactStatusChangedEvent struct {
	ContactID     uuid.UUID
	PipelineID    uuid.UUID
	OldStatusID   *uuid.UUID
	NewStatusID   uuid.UUID
	OldStatusName *string
	NewStatusName string
	ChangedAt     time.Time
	ChangedBy     *uuid.UUID
	Reason        string
}

func (e ContactStatusChangedEvent) EventName() string     { return "contact.status_changed" }
func (e ContactStatusChangedEvent) OccurredAt() time.Time { return e.ChangedAt }

type ContactEnteredPipelineEvent struct {
	ContactID  uuid.UUID
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	EnteredAt  time.Time
	EnteredBy  *uuid.UUID
}

func (e ContactEnteredPipelineEvent) EventName() string     { return "contact.entered_pipeline" }
func (e ContactEnteredPipelineEvent) OccurredAt() time.Time { return e.EnteredAt }

type ContactExitedPipelineEvent struct {
	ContactID      uuid.UUID
	PipelineID     uuid.UUID
	LastStatusID   uuid.UUID
	LastStatusName string
	ExitedAt       time.Time
	ExitedBy       *uuid.UUID
	Reason         string
}

func (e ContactExitedPipelineEvent) EventName() string     { return "contact.exited_pipeline" }
func (e ContactExitedPipelineEvent) OccurredAt() time.Time { return e.ExitedAt }

type AutomationCreatedEvent struct {
	RuleID     uuid.UUID
	PipelineID uuid.UUID
	TenantID   string
	Name       string
	Trigger    AutomationTrigger
	CreatedAt  time.Time
}

func (e AutomationCreatedEvent) EventName() string     { return "automation.created" }
func (e AutomationCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

type AutomationEnabledEvent struct {
	RuleID    uuid.UUID
	EnabledAt time.Time
}

func (e AutomationEnabledEvent) EventName() string     { return "automation.enabled" }
func (e AutomationEnabledEvent) OccurredAt() time.Time { return e.EnabledAt }

type AutomationDisabledEvent struct {
	RuleID     uuid.UUID
	DisabledAt time.Time
}

func (e AutomationDisabledEvent) EventName() string     { return "automation.disabled" }
func (e AutomationDisabledEvent) OccurredAt() time.Time { return e.DisabledAt }

type AutomationRuleTriggeredEvent struct {
	RuleID      uuid.UUID
	SessionID   *uuid.UUID
	ContactID   *uuid.UUID
	TriggerType AutomationTrigger
	Context     map[string]interface{}
	TriggeredAt time.Time
}

func (e AutomationRuleTriggeredEvent) EventName() string     { return "automation_rule.triggered" }
func (e AutomationRuleTriggeredEvent) OccurredAt() time.Time { return e.TriggeredAt }

type AutomationRuleExecutedEvent struct {
	RuleID       uuid.UUID
	SessionID    *uuid.UUID
	ContactID    *uuid.UUID
	ActionsCount int
	ExecutedAt   time.Time
}

func (e AutomationRuleExecutedEvent) EventName() string     { return "automation_rule.executed" }
func (e AutomationRuleExecutedEvent) OccurredAt() time.Time { return e.ExecutedAt }

type AutomationRuleFailedEvent struct {
	RuleID    uuid.UUID
	SessionID *uuid.UUID
	ContactID *uuid.UUID
	Error     string
	FailedAt  time.Time
}

func (e AutomationRuleFailedEvent) EventName() string     { return "automation_rule.failed" }
func (e AutomationRuleFailedEvent) OccurredAt() time.Time { return e.FailedAt }

// Lead Qualification Events

type LeadQualificationEnabledEvent struct {
	PipelineID uuid.UUID
	EnabledAt  time.Time
}

func (e LeadQualificationEnabledEvent) EventName() string {
	return "pipeline.lead_qualification_enabled"
}
func (e LeadQualificationEnabledEvent) OccurredAt() time.Time { return e.EnabledAt }

type LeadQualificationDisabledEvent struct {
	PipelineID uuid.UUID
	DisabledAt time.Time
}

func (e LeadQualificationDisabledEvent) EventName() string {
	return "pipeline.lead_qualification_disabled"
}
func (e LeadQualificationDisabledEvent) OccurredAt() time.Time { return e.DisabledAt }

type LeadQualificationConfigUpdatedEvent struct {
	PipelineID uuid.UUID
	UpdatedAt  time.Time
}

func (e LeadQualificationConfigUpdatedEvent) EventName() string {
	return "pipeline.lead_qualification_config_updated"
}
func (e LeadQualificationConfigUpdatedEvent) OccurredAt() time.Time { return e.UpdatedAt }

// ProfilePictureReceivedEvent - dispara quando contato recebe foto de perfil
// Este evento é consumido para processar qualificação
type ProfilePictureReceivedEvent struct {
	ContactID         uuid.UUID
	PipelineID        *uuid.UUID // Pipeline atual do contato (se tiver)
	ProfilePictureURL string
	ReceivedAt        time.Time
}

func (e ProfilePictureReceivedEvent) EventName() string     { return "contact.profile_picture_received" }
func (e ProfilePictureReceivedEvent) OccurredAt() time.Time { return e.ReceivedAt }

// LeadQualifiedEvent - dispara após análise da foto completar
type LeadQualifiedEvent struct {
	ContactID   uuid.UUID
	PipelineID  uuid.UUID
	Score       int               // 0-10
	Qualified   bool              // Se passou no score mínimo
	Answers     map[string]string // Respostas da IA
	Confidence  string            // high, medium, low
	QualifiedAt time.Time
}

func (e LeadQualifiedEvent) EventName() string     { return "contact.lead_qualified" }
func (e LeadQualifiedEvent) OccurredAt() time.Time { return e.QualifiedAt }
