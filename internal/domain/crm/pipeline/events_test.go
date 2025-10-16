package pipeline

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPipelineCreatedEvent(t *testing.T) {
	pipelineID := uuid.New()
	projectID := uuid.New()

	event := NewPipelineCreatedEvent(pipelineID, projectID, "tenant-123", "Sales Pipeline")

	assert.Equal(t, "pipeline.created", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestPipelineUpdatedEvent(t *testing.T) {
	pipelineID := uuid.New()

	event := NewPipelineUpdatedEvent(pipelineID, "name", "Old Name", "New Name")

	assert.Equal(t, "pipeline.updated", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestPipelineActivatedEvent(t *testing.T) {
	pipelineID := uuid.New()

	event := NewPipelineActivatedEvent(pipelineID)

	assert.Equal(t, "pipeline.activated", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestPipelineDeactivatedEvent(t *testing.T) {
	pipelineID := uuid.New()

	event := NewPipelineDeactivatedEvent(pipelineID)

	assert.Equal(t, "pipeline.deactivated", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestStatusCreatedEvent(t *testing.T) {
	statusID := uuid.New()
	pipelineID := uuid.New()

	event := NewStatusCreatedEvent(statusID, pipelineID, "New Lead", StatusTypeOpen)

	assert.Equal(t, "status.created", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestStatusUpdatedEvent(t *testing.T) {
	statusID := uuid.New()

	event := NewStatusUpdatedEvent(statusID, "name", "Old Status", "New Status")

	assert.Equal(t, "status.updated", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestStatusActivatedEvent(t *testing.T) {
	statusID := uuid.New()

	event := NewStatusActivatedEvent(statusID)

	assert.Equal(t, "status.activated", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestStatusDeactivatedEvent(t *testing.T) {
	statusID := uuid.New()

	event := NewStatusDeactivatedEvent(statusID)

	assert.Equal(t, "status.deactivated", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestStatusAddedToPipelineEvent(t *testing.T) {
	pipelineID := uuid.New()
	statusID := uuid.New()

	event := NewStatusAddedToPipelineEvent(pipelineID, statusID, "New Lead")

	assert.Equal(t, "pipeline.status_added", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestStatusRemovedFromPipelineEvent(t *testing.T) {
	pipelineID := uuid.New()
	statusID := uuid.New()

	event := NewStatusRemovedFromPipelineEvent(pipelineID, statusID, "Old Status")

	assert.Equal(t, "pipeline.status_removed", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestContactStatusChangedEvent(t *testing.T) {
	contactID := uuid.New()
	pipelineID := uuid.New()
	oldStatusID := uuid.New()
	newStatusID := uuid.New()
	changedBy := uuid.New()
	oldStatusName := "Old Status"

	event := NewContactStatusChangedEvent(
		contactID,
		pipelineID,
		&oldStatusID,
		newStatusID,
		&oldStatusName,
		"New Status",
		&changedBy,
		"Customer requested",
	)

	assert.Equal(t, "contact.status_changed", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestContactEnteredPipelineEvent(t *testing.T) {
	contactID := uuid.New()
	pipelineID := uuid.New()
	statusID := uuid.New()
	enteredBy := uuid.New()

	event := NewContactEnteredPipelineEvent(contactID, pipelineID, statusID, "New Lead", &enteredBy)

	assert.Equal(t, "contact.entered_pipeline", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestContactExitedPipelineEvent(t *testing.T) {
	contactID := uuid.New()
	pipelineID := uuid.New()
	lastStatusID := uuid.New()
	exitedBy := uuid.New()

	event := NewContactExitedPipelineEvent(
		contactID,
		pipelineID,
		lastStatusID,
		"Closed Lost",
		&exitedBy,
		"Not interested",
	)

	assert.Equal(t, "contact.exited_pipeline", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestAutomationCreatedEvent(t *testing.T) {
	ruleID := uuid.New()
	pipelineID := uuid.New()

	event := NewAutomationCreatedEvent(ruleID, pipelineID, "tenant-123", "Follow-up Rule", TriggerStatusChanged)

	assert.Equal(t, "automation.created", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestAutomationEnabledEvent(t *testing.T) {
	ruleID := uuid.New()

	event := NewAutomationEnabledEvent(ruleID)

	assert.Equal(t, "automation.enabled", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestAutomationDisabledEvent(t *testing.T) {
	ruleID := uuid.New()

	event := NewAutomationDisabledEvent(ruleID)

	assert.Equal(t, "automation.disabled", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestAutomationRuleTriggeredEvent(t *testing.T) {
	ruleID := uuid.New()
	sessionID := uuid.New()
	contactID := uuid.New()

	event := NewAutomationRuleTriggeredEvent(
		ruleID,
		&sessionID,
		&contactID,
		TriggerStatusChanged,
		map[string]interface{}{
			"old_status": "new",
			"new_status": "in_progress",
		},
	)

	assert.Equal(t, "automation_rule.triggered", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestAutomationRuleExecutedEvent(t *testing.T) {
	ruleID := uuid.New()
	sessionID := uuid.New()
	contactID := uuid.New()

	event := NewAutomationRuleExecutedEvent(ruleID, &sessionID, &contactID, 3)

	assert.Equal(t, "automation_rule.executed", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

func TestAutomationRuleFailedEvent(t *testing.T) {
	ruleID := uuid.New()
	sessionID := uuid.New()
	contactID := uuid.New()

	event := NewAutomationRuleFailedEvent(ruleID, &sessionID, &contactID, "Failed to send message")

	assert.Equal(t, "automation_rule.failed", event.EventName())
	assert.False(t, event.OccurredAt().IsZero())
}

// Test that all events implement DomainEvent interface
func TestAllEventsImplementDomainEvent(t *testing.T) {
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
		NewPipelineCreatedEvent(id, id, "tenant", "name"),
		NewPipelineUpdatedEvent(id, "field", "old", "new"),
		NewPipelineActivatedEvent(id),
		NewPipelineDeactivatedEvent(id),
		NewStatusCreatedEvent(id, id, "name", StatusTypeOpen),
		NewStatusUpdatedEvent(id, "field", "old", "new"),
		NewStatusActivatedEvent(id),
		NewStatusDeactivatedEvent(id),
		NewStatusAddedToPipelineEvent(id, id, "name"),
		NewStatusRemovedFromPipelineEvent(id, id, "name"),
		NewContactStatusChangedEvent(id, id, &id, id, nil, "test", nil, "reason"),
		NewContactEnteredPipelineEvent(id, id, id, "test", nil),
		NewContactExitedPipelineEvent(id, id, id, "test", nil, "reason"),
		NewAutomationCreatedEvent(id, id, "tenant", "name", TriggerStatusChanged),
		NewAutomationEnabledEvent(id),
		NewAutomationDisabledEvent(id),
		NewAutomationRuleTriggeredEvent(id, nil, nil, TriggerStatusChanged, nil),
		NewAutomationRuleExecutedEvent(id, nil, nil, 0),
		NewAutomationRuleFailedEvent(id, nil, nil, "error"),
	}

	for _, event := range events {
		assert.NotEmpty(t, event.EventName())
		assert.False(t, event.OccurredAt().IsZero())
	}
}
