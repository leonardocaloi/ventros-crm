package pipeline

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPipelineCreatedEvent(t *testing.T) {
	pipelineID := uuid.New()
	projectID := uuid.New()
	now := time.Now()

	event := PipelineCreatedEvent{
		PipelineID: pipelineID,
		ProjectID:  projectID,
		TenantID:   "tenant-123",
		Name:       "Sales Pipeline",
		CreatedAt:  now,
	}

	assert.Equal(t, "pipeline.created", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestPipelineUpdatedEvent(t *testing.T) {
	pipelineID := uuid.New()
	now := time.Now()

	event := PipelineUpdatedEvent{
		PipelineID: pipelineID,
		Field:      "name",
		OldValue:   "Old Name",
		NewValue:   "New Name",
		UpdatedAt:  now,
	}

	assert.Equal(t, "pipeline.updated", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestPipelineActivatedEvent(t *testing.T) {
	pipelineID := uuid.New()
	now := time.Now()

	event := PipelineActivatedEvent{
		PipelineID:  pipelineID,
		ActivatedAt: now,
	}

	assert.Equal(t, "pipeline.activated", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestPipelineDeactivatedEvent(t *testing.T) {
	pipelineID := uuid.New()
	now := time.Now()

	event := PipelineDeactivatedEvent{
		PipelineID:    pipelineID,
		DeactivatedAt: now,
	}

	assert.Equal(t, "pipeline.deactivated", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestStatusCreatedEvent(t *testing.T) {
	statusID := uuid.New()
	pipelineID := uuid.New()
	now := time.Now()

	event := StatusCreatedEvent{
		StatusID:   statusID,
		PipelineID: pipelineID,
		Name:       "New Lead",
		StatusType: StatusTypeOpen,
		CreatedAt:  now,
	}

	assert.Equal(t, "status.created", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestStatusUpdatedEvent(t *testing.T) {
	statusID := uuid.New()
	now := time.Now()

	event := StatusUpdatedEvent{
		StatusID:  statusID,
		Field:     "name",
		OldValue:  "Old Status",
		NewValue:  "New Status",
		UpdatedAt: now,
	}

	assert.Equal(t, "status.updated", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestStatusActivatedEvent(t *testing.T) {
	statusID := uuid.New()
	now := time.Now()

	event := StatusActivatedEvent{
		StatusID:    statusID,
		ActivatedAt: now,
	}

	assert.Equal(t, "status.activated", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestStatusDeactivatedEvent(t *testing.T) {
	statusID := uuid.New()
	now := time.Now()

	event := StatusDeactivatedEvent{
		StatusID:      statusID,
		DeactivatedAt: now,
	}

	assert.Equal(t, "status.deactivated", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestStatusAddedToPipelineEvent(t *testing.T) {
	pipelineID := uuid.New()
	statusID := uuid.New()
	now := time.Now()

	event := StatusAddedToPipelineEvent{
		PipelineID: pipelineID,
		StatusID:   statusID,
		StatusName: "New Lead",
		AddedAt:    now,
	}

	assert.Equal(t, "pipeline.status_added", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestStatusRemovedFromPipelineEvent(t *testing.T) {
	pipelineID := uuid.New()
	statusID := uuid.New()
	now := time.Now()

	event := StatusRemovedFromPipelineEvent{
		PipelineID: pipelineID,
		StatusID:   statusID,
		StatusName: "Old Status",
		RemovedAt:  now,
	}

	assert.Equal(t, "pipeline.status_removed", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestContactStatusChangedEvent(t *testing.T) {
	contactID := uuid.New()
	pipelineID := uuid.New()
	oldStatusID := uuid.New()
	newStatusID := uuid.New()
	changedBy := uuid.New()
	now := time.Now()
	oldStatusName := "Old Status"

	event := ContactStatusChangedEvent{
		ContactID:     contactID,
		PipelineID:    pipelineID,
		OldStatusID:   &oldStatusID,
		NewStatusID:   newStatusID,
		OldStatusName: &oldStatusName,
		NewStatusName: "New Status",
		ChangedAt:     now,
		ChangedBy:     &changedBy,
		Reason:        "Customer requested",
	}

	assert.Equal(t, "contact.status_changed", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestContactEnteredPipelineEvent(t *testing.T) {
	contactID := uuid.New()
	pipelineID := uuid.New()
	statusID := uuid.New()
	enteredBy := uuid.New()
	now := time.Now()

	event := ContactEnteredPipelineEvent{
		ContactID:  contactID,
		PipelineID: pipelineID,
		StatusID:   statusID,
		StatusName: "New Lead",
		EnteredAt:  now,
		EnteredBy:  &enteredBy,
	}

	assert.Equal(t, "contact.entered_pipeline", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestContactExitedPipelineEvent(t *testing.T) {
	contactID := uuid.New()
	pipelineID := uuid.New()
	lastStatusID := uuid.New()
	exitedBy := uuid.New()
	now := time.Now()

	event := ContactExitedPipelineEvent{
		ContactID:      contactID,
		PipelineID:     pipelineID,
		LastStatusID:   lastStatusID,
		LastStatusName: "Closed Lost",
		ExitedAt:       now,
		ExitedBy:       &exitedBy,
		Reason:         "Not interested",
	}

	assert.Equal(t, "contact.exited_pipeline", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestAutomationCreatedEvent(t *testing.T) {
	ruleID := uuid.New()
	pipelineID := uuid.New()
	now := time.Now()

	event := AutomationCreatedEvent{
		RuleID:     ruleID,
		PipelineID: pipelineID,
		TenantID:   "tenant-123",
		Name:       "Follow-up Rule",
		Trigger:    TriggerStatusChanged,
		CreatedAt:  now,
	}

	assert.Equal(t, "automation.created", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestAutomationEnabledEvent(t *testing.T) {
	ruleID := uuid.New()
	now := time.Now()

	event := AutomationEnabledEvent{
		RuleID:    ruleID,
		EnabledAt: now,
	}

	assert.Equal(t, "automation.enabled", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestAutomationDisabledEvent(t *testing.T) {
	ruleID := uuid.New()
	now := time.Now()

	event := AutomationDisabledEvent{
		RuleID:     ruleID,
		DisabledAt: now,
	}

	assert.Equal(t, "automation.disabled", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestAutomationRuleTriggeredEvent(t *testing.T) {
	ruleID := uuid.New()
	sessionID := uuid.New()
	contactID := uuid.New()
	now := time.Now()

	event := AutomationRuleTriggeredEvent{
		RuleID:      ruleID,
		SessionID:   &sessionID,
		ContactID:   &contactID,
		TriggerType: TriggerStatusChanged,
		Context: map[string]interface{}{
			"old_status": "new",
			"new_status": "in_progress",
		},
		TriggeredAt: now,
	}

	assert.Equal(t, "automation_rule.triggered", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestAutomationRuleExecutedEvent(t *testing.T) {
	ruleID := uuid.New()
	sessionID := uuid.New()
	contactID := uuid.New()
	now := time.Now()

	event := AutomationRuleExecutedEvent{
		RuleID:       ruleID,
		SessionID:    &sessionID,
		ContactID:    &contactID,
		ActionsCount: 3,
		ExecutedAt:   now,
	}

	assert.Equal(t, "automation_rule.executed", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

func TestAutomationRuleFailedEvent(t *testing.T) {
	ruleID := uuid.New()
	sessionID := uuid.New()
	contactID := uuid.New()
	now := time.Now()

	event := AutomationRuleFailedEvent{
		RuleID:    ruleID,
		SessionID: &sessionID,
		ContactID: &contactID,
		Error:     "Failed to send message",
		FailedAt:  now,
	}

	assert.Equal(t, "automation_rule.failed", event.EventName())
	assert.Equal(t, now, event.OccurredAt())
}

// Test that all events implement DomainEvent interface
func TestAllEventsImplementDomainEvent(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	var _ DomainEvent = PipelineCreatedEvent{}
	var _ DomainEvent = PipelineUpdatedEvent{}
	var _ DomainEvent = PipelineActivatedEvent{}
	var _ DomainEvent = PipelineDeactivatedEvent{}
	var _ DomainEvent = StatusCreatedEvent{}
	var _ DomainEvent = StatusUpdatedEvent{}
	var _ DomainEvent = StatusActivatedEvent{}
	var _ DomainEvent = StatusDeactivatedEvent{}
	var _ DomainEvent = StatusAddedToPipelineEvent{}
	var _ DomainEvent = StatusRemovedFromPipelineEvent{}
	var _ DomainEvent = ContactStatusChangedEvent{}
	var _ DomainEvent = ContactEnteredPipelineEvent{}
	var _ DomainEvent = ContactExitedPipelineEvent{}
	var _ DomainEvent = AutomationCreatedEvent{}
	var _ DomainEvent = AutomationEnabledEvent{}
	var _ DomainEvent = AutomationDisabledEvent{}
	var _ DomainEvent = AutomationRuleTriggeredEvent{}
	var _ DomainEvent = AutomationRuleExecutedEvent{}
	var _ DomainEvent = AutomationRuleFailedEvent{}

	// Verify all events can be used as DomainEvent
	events := []DomainEvent{
		PipelineCreatedEvent{CreatedAt: now},
		PipelineUpdatedEvent{UpdatedAt: now},
		PipelineActivatedEvent{ActivatedAt: now},
		PipelineDeactivatedEvent{DeactivatedAt: now},
		StatusCreatedEvent{CreatedAt: now},
		StatusUpdatedEvent{UpdatedAt: now},
		StatusActivatedEvent{ActivatedAt: now},
		StatusDeactivatedEvent{DeactivatedAt: now},
		StatusAddedToPipelineEvent{AddedAt: now},
		StatusRemovedFromPipelineEvent{RemovedAt: now},
		ContactStatusChangedEvent{ContactID: id, PipelineID: id, NewStatusID: id, NewStatusName: "test", ChangedAt: now},
		ContactEnteredPipelineEvent{ContactID: id, PipelineID: id, StatusID: id, StatusName: "test", EnteredAt: now},
		ContactExitedPipelineEvent{ContactID: id, PipelineID: id, LastStatusID: id, LastStatusName: "test", ExitedAt: now},
		AutomationCreatedEvent{CreatedAt: now},
		AutomationEnabledEvent{EnabledAt: now},
		AutomationDisabledEvent{DisabledAt: now},
		AutomationRuleTriggeredEvent{RuleID: id, TriggerType: TriggerStatusChanged, TriggeredAt: now},
		AutomationRuleExecutedEvent{RuleID: id, ExecutedAt: now},
		AutomationRuleFailedEvent{RuleID: id, FailedAt: now},
	}

	for _, event := range events {
		assert.NotEmpty(t, event.EventName())
		assert.False(t, event.OccurredAt().IsZero())
	}
}
