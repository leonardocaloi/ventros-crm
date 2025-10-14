package pipeline

import (
	"time"

	"github.com/google/uuid"
	"github.com/ventros/crm/internal/domain/core/shared"
)

type PipelineCreatedEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
	ProjectID  uuid.UUID
	TenantID   string
	Name       string
}

func NewPipelineCreatedEvent(pipelineID, projectID uuid.UUID, tenantID, name string) PipelineCreatedEvent {
	return PipelineCreatedEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.created", time.Now()),
		PipelineID: pipelineID,
		ProjectID:  projectID,
		TenantID:   tenantID,
		Name:       name,
	}
}

type PipelineUpdatedEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
	Field      string
	OldValue   interface{}
	NewValue   interface{}
}

func NewPipelineUpdatedEvent(pipelineID uuid.UUID, field string, oldValue, newValue interface{}) PipelineUpdatedEvent {
	return PipelineUpdatedEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.updated", time.Now()),
		PipelineID: pipelineID,
		Field:      field,
		OldValue:   oldValue,
		NewValue:   newValue,
	}
}

type PipelineActivatedEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
}

func NewPipelineActivatedEvent(pipelineID uuid.UUID) PipelineActivatedEvent {
	return PipelineActivatedEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.activated", time.Now()),
		PipelineID: pipelineID,
	}
}

type PipelineDeactivatedEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
}

func NewPipelineDeactivatedEvent(pipelineID uuid.UUID) PipelineDeactivatedEvent {
	return PipelineDeactivatedEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.deactivated", time.Now()),
		PipelineID: pipelineID,
	}
}

type StatusCreatedEvent struct {
	shared.BaseEvent
	StatusID   uuid.UUID
	PipelineID uuid.UUID
	Name       string
	StatusType StatusType
}

func NewStatusCreatedEvent(statusID, pipelineID uuid.UUID, name string, statusType StatusType) StatusCreatedEvent {
	return StatusCreatedEvent{
		BaseEvent:  shared.NewBaseEvent("status.created", time.Now()),
		StatusID:   statusID,
		PipelineID: pipelineID,
		Name:       name,
		StatusType: statusType,
	}
}

type StatusUpdatedEvent struct {
	shared.BaseEvent
	StatusID uuid.UUID
	Field    string
	OldValue interface{}
	NewValue interface{}
}

func NewStatusUpdatedEvent(statusID uuid.UUID, field string, oldValue, newValue interface{}) StatusUpdatedEvent {
	return StatusUpdatedEvent{
		BaseEvent: shared.NewBaseEvent("status.updated", time.Now()),
		StatusID:  statusID,
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
	}
}

type StatusActivatedEvent struct {
	shared.BaseEvent
	StatusID uuid.UUID
}

func NewStatusActivatedEvent(statusID uuid.UUID) StatusActivatedEvent {
	return StatusActivatedEvent{
		BaseEvent: shared.NewBaseEvent("status.activated", time.Now()),
		StatusID:  statusID,
	}
}

type StatusDeactivatedEvent struct {
	shared.BaseEvent
	StatusID uuid.UUID
}

func NewStatusDeactivatedEvent(statusID uuid.UUID) StatusDeactivatedEvent {
	return StatusDeactivatedEvent{
		BaseEvent: shared.NewBaseEvent("status.deactivated", time.Now()),
		StatusID:  statusID,
	}
}

type StatusAddedToPipelineEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
}

func NewStatusAddedToPipelineEvent(pipelineID, statusID uuid.UUID, statusName string) StatusAddedToPipelineEvent {
	return StatusAddedToPipelineEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.status_added", time.Now()),
		PipelineID: pipelineID,
		StatusID:   statusID,
		StatusName: statusName,
	}
}

type StatusRemovedFromPipelineEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
}

func NewStatusRemovedFromPipelineEvent(pipelineID, statusID uuid.UUID, statusName string) StatusRemovedFromPipelineEvent {
	return StatusRemovedFromPipelineEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.status_removed", time.Now()),
		PipelineID: pipelineID,
		StatusID:   statusID,
		StatusName: statusName,
	}
}

type ContactStatusChangedEvent struct {
	shared.BaseEvent
	ContactID     uuid.UUID
	PipelineID    uuid.UUID
	OldStatusID   *uuid.UUID
	NewStatusID   uuid.UUID
	OldStatusName *string
	NewStatusName string
	ChangedBy     *uuid.UUID
	Reason        string
}

func NewContactStatusChangedEvent(contactID, pipelineID uuid.UUID, oldStatusID *uuid.UUID, newStatusID uuid.UUID, oldStatusName *string, newStatusName string, changedBy *uuid.UUID, reason string) ContactStatusChangedEvent {
	return ContactStatusChangedEvent{
		BaseEvent:     shared.NewBaseEvent("contact.status_changed", time.Now()),
		ContactID:     contactID,
		PipelineID:    pipelineID,
		OldStatusID:   oldStatusID,
		NewStatusID:   newStatusID,
		OldStatusName: oldStatusName,
		NewStatusName: newStatusName,
		ChangedBy:     changedBy,
		Reason:        reason,
	}
}

type ContactEnteredPipelineEvent struct {
	shared.BaseEvent
	ContactID  uuid.UUID
	PipelineID uuid.UUID
	StatusID   uuid.UUID
	StatusName string
	EnteredBy  *uuid.UUID
}

func NewContactEnteredPipelineEvent(contactID, pipelineID, statusID uuid.UUID, statusName string, enteredBy *uuid.UUID) ContactEnteredPipelineEvent {
	return ContactEnteredPipelineEvent{
		BaseEvent:  shared.NewBaseEvent("contact.entered_pipeline", time.Now()),
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		StatusName: statusName,
		EnteredBy:  enteredBy,
	}
}

type ContactExitedPipelineEvent struct {
	shared.BaseEvent
	ContactID      uuid.UUID
	PipelineID     uuid.UUID
	LastStatusID   uuid.UUID
	LastStatusName string
	ExitedBy       *uuid.UUID
	Reason         string
}

func NewContactExitedPipelineEvent(contactID, pipelineID, lastStatusID uuid.UUID, lastStatusName string, exitedBy *uuid.UUID, reason string) ContactExitedPipelineEvent {
	return ContactExitedPipelineEvent{
		BaseEvent:      shared.NewBaseEvent("contact.exited_pipeline", time.Now()),
		ContactID:      contactID,
		PipelineID:     pipelineID,
		LastStatusID:   lastStatusID,
		LastStatusName: lastStatusName,
		ExitedBy:       exitedBy,
		Reason:         reason,
	}
}

type AutomationCreatedEvent struct {
	shared.BaseEvent
	RuleID     uuid.UUID
	PipelineID uuid.UUID
	TenantID   string
	Name       string
	Trigger    AutomationTrigger
}

func NewAutomationCreatedEvent(ruleID, pipelineID uuid.UUID, tenantID, name string, trigger AutomationTrigger) AutomationCreatedEvent {
	return AutomationCreatedEvent{
		BaseEvent:  shared.NewBaseEvent("automation.created", time.Now()),
		RuleID:     ruleID,
		PipelineID: pipelineID,
		TenantID:   tenantID,
		Name:       name,
		Trigger:    trigger,
	}
}

type AutomationEnabledEvent struct {
	shared.BaseEvent
	RuleID uuid.UUID
}

func NewAutomationEnabledEvent(ruleID uuid.UUID) AutomationEnabledEvent {
	return AutomationEnabledEvent{
		BaseEvent: shared.NewBaseEvent("automation.enabled", time.Now()),
		RuleID:    ruleID,
	}
}

type AutomationDisabledEvent struct {
	shared.BaseEvent
	RuleID uuid.UUID
}

func NewAutomationDisabledEvent(ruleID uuid.UUID) AutomationDisabledEvent {
	return AutomationDisabledEvent{
		BaseEvent: shared.NewBaseEvent("automation.disabled", time.Now()),
		RuleID:    ruleID,
	}
}

type AutomationRuleTriggeredEvent struct {
	shared.BaseEvent
	RuleID      uuid.UUID
	SessionID   *uuid.UUID
	ContactID   *uuid.UUID
	TriggerType AutomationTrigger
	Context     map[string]interface{}
}

func NewAutomationRuleTriggeredEvent(ruleID uuid.UUID, sessionID, contactID *uuid.UUID, triggerType AutomationTrigger, context map[string]interface{}) AutomationRuleTriggeredEvent {
	return AutomationRuleTriggeredEvent{
		BaseEvent:   shared.NewBaseEvent("automation_rule.triggered", time.Now()),
		RuleID:      ruleID,
		SessionID:   sessionID,
		ContactID:   contactID,
		TriggerType: triggerType,
		Context:     context,
	}
}

type AutomationRuleExecutedEvent struct {
	shared.BaseEvent
	RuleID       uuid.UUID
	SessionID    *uuid.UUID
	ContactID    *uuid.UUID
	ActionsCount int
}

func NewAutomationRuleExecutedEvent(ruleID uuid.UUID, sessionID, contactID *uuid.UUID, actionsCount int) AutomationRuleExecutedEvent {
	return AutomationRuleExecutedEvent{
		BaseEvent:    shared.NewBaseEvent("automation_rule.executed", time.Now()),
		RuleID:       ruleID,
		SessionID:    sessionID,
		ContactID:    contactID,
		ActionsCount: actionsCount,
	}
}

type AutomationRuleFailedEvent struct {
	shared.BaseEvent
	RuleID    uuid.UUID
	SessionID *uuid.UUID
	ContactID *uuid.UUID
	Error     string
}

func NewAutomationRuleFailedEvent(ruleID uuid.UUID, sessionID, contactID *uuid.UUID, errorMsg string) AutomationRuleFailedEvent {
	return AutomationRuleFailedEvent{
		BaseEvent: shared.NewBaseEvent("automation_rule.failed", time.Now()),
		RuleID:    ruleID,
		SessionID: sessionID,
		ContactID: contactID,
		Error:     errorMsg,
	}
}

// Lead Qualification Events

type LeadQualificationEnabledEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
}

func NewLeadQualificationEnabledEvent(pipelineID uuid.UUID) LeadQualificationEnabledEvent {
	return LeadQualificationEnabledEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.lead_qualification_enabled", time.Now()),
		PipelineID: pipelineID,
	}
}

type LeadQualificationDisabledEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
}

func NewLeadQualificationDisabledEvent(pipelineID uuid.UUID) LeadQualificationDisabledEvent {
	return LeadQualificationDisabledEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.lead_qualification_disabled", time.Now()),
		PipelineID: pipelineID,
	}
}

type LeadQualificationConfigUpdatedEvent struct {
	shared.BaseEvent
	PipelineID uuid.UUID
}

func NewLeadQualificationConfigUpdatedEvent(pipelineID uuid.UUID) LeadQualificationConfigUpdatedEvent {
	return LeadQualificationConfigUpdatedEvent{
		BaseEvent:  shared.NewBaseEvent("pipeline.lead_qualification_config_updated", time.Now()),
		PipelineID: pipelineID,
	}
}

// ProfilePictureReceivedEvent - dispara quando contato recebe foto de perfil
// Este evento é consumido para processar qualificação
type ProfilePictureReceivedEvent struct {
	shared.BaseEvent
	ContactID         uuid.UUID
	PipelineID        *uuid.UUID // Pipeline atual do contato (se tiver)
	ProfilePictureURL string
}

func NewProfilePictureReceivedEvent(contactID uuid.UUID, pipelineID *uuid.UUID, profilePictureURL string) ProfilePictureReceivedEvent {
	return ProfilePictureReceivedEvent{
		BaseEvent:         shared.NewBaseEvent("contact.profile_picture_received", time.Now()),
		ContactID:         contactID,
		PipelineID:        pipelineID,
		ProfilePictureURL: profilePictureURL,
	}
}

// LeadQualifiedEvent - dispara após análise da foto completar
type LeadQualifiedEvent struct {
	shared.BaseEvent
	ContactID  uuid.UUID
	PipelineID uuid.UUID
	Score      int               // 0-10
	Qualified  bool              // Se passou no score mínimo
	Answers    map[string]string // Respostas da IA
	Confidence string            // high, medium, low
}

func NewLeadQualifiedEvent(contactID, pipelineID uuid.UUID, score int, qualified bool, answers map[string]string, confidence string) LeadQualifiedEvent {
	return LeadQualifiedEvent{
		BaseEvent:  shared.NewBaseEvent("contact.lead_qualified", time.Now()),
		ContactID:  contactID,
		PipelineID: pipelineID,
		Score:      score,
		Qualified:  qualified,
		Answers:    answers,
		Confidence: confidence,
	}
}
