package pipeline

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAutomation(t *testing.T) {
	pipelineID := uuid.New()

	tests := []struct {
		name           string
		automationType AutomationType
		tenantID       string
		ruleName       string
		trigger        AutomationTrigger
		pipelineID     *uuid.UUID
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "valid pipeline automation",
			automationType: AutomationTypePipelineBased,
			tenantID:       "tenant-123",
			ruleName:       "Follow-up Rule",
			trigger:        TriggerSessionTimeout,
			pipelineID:     &pipelineID,
			wantErr:        false,
		},
		{
			name:           "valid scheduled report (no pipeline required)",
			automationType: AutomationTypeScheduledReport,
			tenantID:       "tenant-123",
			ruleName:       "Daily Report",
			trigger:        TriggerScheduled,
			pipelineID:     nil,
			wantErr:        false,
		},
		{
			name:           "empty tenantID",
			automationType: AutomationTypePipelineBased,
			tenantID:       "",
			ruleName:       "Rule",
			trigger:        TriggerStatusChanged,
			pipelineID:     &pipelineID,
			wantErr:        true,
			errMsg:         "tenantID cannot be empty",
		},
		{
			name:           "empty name",
			automationType: AutomationTypePipelineBased,
			tenantID:       "tenant-123",
			ruleName:       "",
			trigger:        TriggerStatusChanged,
			pipelineID:     &pipelineID,
			wantErr:        true,
			errMsg:         "name cannot be empty",
		},
		{
			name:           "empty trigger",
			automationType: AutomationTypePipelineBased,
			tenantID:       "tenant-123",
			ruleName:       "Rule",
			trigger:        "",
			pipelineID:     &pipelineID,
			wantErr:        true,
			errMsg:         "trigger cannot be empty",
		},
		{
			name:           "empty automation type",
			automationType: "",
			tenantID:       "tenant-123",
			ruleName:       "Rule",
			trigger:        TriggerStatusChanged,
			pipelineID:     &pipelineID,
			wantErr:        true,
			errMsg:         "automationType cannot be empty",
		},
		{
			name:           "pipeline automation without pipelineID",
			automationType: AutomationTypePipelineBased,
			tenantID:       "tenant-123",
			ruleName:       "Rule",
			trigger:        TriggerStatusChanged,
			pipelineID:     nil,
			wantErr:        true,
			errMsg:         "pipeline automations require a valid pipelineID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewAutomation(tt.automationType, tt.tenantID, tt.ruleName, tt.trigger, tt.pipelineID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, rule)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rule)

				assert.NotEqual(t, uuid.Nil, rule.ID())
				assert.Equal(t, tt.automationType, rule.Type())
				assert.Equal(t, tt.tenantID, rule.TenantID())
				assert.Equal(t, tt.ruleName, rule.Name())
				assert.Equal(t, tt.trigger, rule.Trigger())
				assert.True(t, rule.IsEnabled())
				assert.Equal(t, 0, rule.Priority())
				assert.NotZero(t, rule.CreatedAt())
				assert.NotZero(t, rule.UpdatedAt())

				// Check domain event
				events := rule.DomainEvents()
				require.Len(t, events, 1)
				event, ok := events[0].(AutomationCreatedEvent)
				require.True(t, ok)
				assert.Equal(t, rule.ID(), event.RuleID)
			}
		})
	}
}

func TestAutomation_AddCondition(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	t.Run("add valid condition", func(t *testing.T) {
		err := rule.AddCondition("message_count", "gt", 5)
		require.NoError(t, err)

		conditions := rule.Conditions()
		require.Len(t, conditions, 1)
		assert.Equal(t, "message_count", conditions[0].Field)
		assert.Equal(t, "gt", conditions[0].Operator)
		assert.Equal(t, 5, conditions[0].Value)
	})

	t.Run("empty field", func(t *testing.T) {
		err := rule.AddCondition("", "gt", 5)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "field cannot be empty")
	})

	t.Run("empty operator", func(t *testing.T) {
		err := rule.AddCondition("status", "", "active")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operator cannot be empty")
	})
}

func TestAutomation_AddAction(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	t.Run("add valid action", func(t *testing.T) {
		params := map[string]interface{}{
			"content": "Hello!",
		}
		err := rule.AddAction(ActionSendMessage, params, 0)
		require.NoError(t, err)

		actions := rule.Actions()
		require.Len(t, actions, 1)
		assert.Equal(t, ActionSendMessage, actions[0].Type)
		assert.Equal(t, "Hello!", actions[0].Params["content"])
		assert.Equal(t, 0, actions[0].Delay)
	})

	t.Run("add action with delay", func(t *testing.T) {
		params := map[string]interface{}{
			"template_name": "welcome",
		}
		err := rule.AddAction(ActionSendTemplate, params, 30)
		require.NoError(t, err)

		actions := rule.Actions()
		require.Len(t, actions, 2)
		assert.Equal(t, 30, actions[1].Delay)
	})

	t.Run("empty action type", func(t *testing.T) {
		err := rule.AddAction("", nil, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "action type cannot be empty")
	})
}

func TestAutomation_SetConditionsAndActions(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	t.Run("set conditions", func(t *testing.T) {
		conditions := []RuleCondition{
			{Field: "status", Operator: "eq", Value: "active"},
			{Field: "message_count", Operator: "gt", Value: 10},
		}
		rule.SetConditions(conditions)

		retrieved := rule.Conditions()
		assert.Len(t, retrieved, 2)
	})

	t.Run("set actions", func(t *testing.T) {
		actions := []RuleAction{
			{Type: ActionSendMessage, Params: map[string]interface{}{"content": "Hi"}, Delay: 0},
			{Type: ActionAddTag, Params: map[string]interface{}{"tag": "vip"}, Delay: 5},
		}
		rule.SetActions(actions)

		retrieved := rule.Actions()
		assert.Len(t, retrieved, 2)
	})
}

func TestAutomation_UpdateDescription(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	rule.UpdateDescription("New description")
	assert.Equal(t, "New description", rule.Description())
}

func TestAutomation_SetPriority(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	t.Run("set valid priority", func(t *testing.T) {
		err := rule.SetPriority(5)
		require.NoError(t, err)
		assert.Equal(t, 5, rule.Priority())
	})

	t.Run("negative priority", func(t *testing.T) {
		err := rule.SetPriority(-1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "priority cannot be negative")
	})
}

func TestAutomation_EnableDisable(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)
	rule.ClearEvents()

	t.Run("disable enabled rule", func(t *testing.T) {
		assert.True(t, rule.IsEnabled())

		rule.Disable()
		assert.False(t, rule.IsEnabled())

		events := rule.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(AutomationDisabledEvent)
		require.True(t, ok)
	})

	t.Run("disable already disabled rule", func(t *testing.T) {
		rule.ClearEvents()
		rule.Disable()

		// Should not generate duplicate event
		events := rule.DomainEvents()
		assert.Len(t, events, 0)
	})

	t.Run("enable disabled rule", func(t *testing.T) {
		rule.ClearEvents()
		rule.Enable()
		assert.True(t, rule.IsEnabled())

		events := rule.DomainEvents()
		require.Len(t, events, 1)
		_, ok := events[0].(AutomationEnabledEvent)
		require.True(t, ok)
	})

	t.Run("enable already enabled rule", func(t *testing.T) {
		rule.ClearEvents()
		rule.Enable()

		// Should not generate duplicate event
		events := rule.DomainEvents()
		assert.Len(t, events, 0)
	})
}

func TestAutomation_EvaluateConditions(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	t.Run("no conditions - always true", func(t *testing.T) {
		context := map[string]interface{}{}
		assert.True(t, rule.EvaluateConditions(context))
	})

	t.Run("all conditions satisfied", func(t *testing.T) {
		rule.AddCondition("message_count", "gt", 5)
		rule.AddCondition("status", "eq", "active")

		context := map[string]interface{}{
			"message_count": 10,
			"status":        "active",
		}
		assert.True(t, rule.EvaluateConditions(context))
	})

	t.Run("one condition not satisfied", func(t *testing.T) {
		context := map[string]interface{}{
			"message_count": 3,
			"status":        "active",
		}
		assert.False(t, rule.EvaluateConditions(context))
	})

	t.Run("missing field in context", func(t *testing.T) {
		context := map[string]interface{}{
			"status": "active",
		}
		assert.False(t, rule.EvaluateConditions(context))
	})
}

func TestEvaluateCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition RuleCondition
		context   map[string]interface{}
		expected  bool
	}{
		{
			name:      "eq - equal",
			condition: RuleCondition{Field: "status", Operator: "eq", Value: "active"},
			context:   map[string]interface{}{"status": "active"},
			expected:  true,
		},
		{
			name:      "eq - not equal",
			condition: RuleCondition{Field: "status", Operator: "eq", Value: "inactive"},
			context:   map[string]interface{}{"status": "active"},
			expected:  false,
		},
		{
			name:      "ne - not equal",
			condition: RuleCondition{Field: "status", Operator: "ne", Value: "inactive"},
			context:   map[string]interface{}{"status": "active"},
			expected:  true,
		},
		{
			name:      "gt - greater than true",
			condition: RuleCondition{Field: "count", Operator: "gt", Value: 5},
			context:   map[string]interface{}{"count": 10},
			expected:  true,
		},
		{
			name:      "gt - greater than false",
			condition: RuleCondition{Field: "count", Operator: "gt", Value: 10},
			context:   map[string]interface{}{"count": 5},
			expected:  false,
		},
		{
			name:      "gte - greater than or equal true",
			condition: RuleCondition{Field: "count", Operator: "gte", Value: 10},
			context:   map[string]interface{}{"count": 10},
			expected:  true,
		},
		{
			name:      "lt - less than true",
			condition: RuleCondition{Field: "count", Operator: "lt", Value: 10},
			context:   map[string]interface{}{"count": 5},
			expected:  true,
		},
		{
			name:      "lte - less than or equal true",
			condition: RuleCondition{Field: "count", Operator: "lte", Value: 10},
			context:   map[string]interface{}{"count": 10},
			expected:  true,
		},
		{
			name:      "contains - true",
			condition: RuleCondition{Field: "message", Operator: "contains", Value: "urgent"},
			context:   map[string]interface{}{"message": "urgent request"},
			expected:  true,
		},
		{
			name:      "contains - false",
			condition: RuleCondition{Field: "message", Operator: "contains", Value: "urgent"},
			context:   map[string]interface{}{"message": "normal request"},
			expected:  false,
		},
		{
			name:      "in - value in slice true",
			condition: RuleCondition{Field: "status", Operator: "in", Value: []interface{}{"active", "pending"}},
			context:   map[string]interface{}{"status": "active"},
			expected:  true,
		},
		{
			name:      "in - value not in slice",
			condition: RuleCondition{Field: "status", Operator: "in", Value: []interface{}{"active", "pending"}},
			context:   map[string]interface{}{"status": "inactive"},
			expected:  false,
		},
		{
			name:      "field not in context",
			condition: RuleCondition{Field: "missing", Operator: "eq", Value: "test"},
			context:   map[string]interface{}{"status": "active"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateCondition(tt.condition, tt.context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluateConditionGroup(t *testing.T) {
	t.Run("empty group returns true", func(t *testing.T) {
		group := ConditionGroup{Logic: LogicAND}
		context := map[string]interface{}{}
		assert.True(t, EvaluateConditionGroup(group, context))
	})

	t.Run("AND logic - all conditions true", func(t *testing.T) {
		group := ConditionGroup{
			Logic: LogicAND,
			Conditions: []RuleCondition{
				{Field: "count", Operator: "gt", Value: 5},
				{Field: "status", Operator: "eq", Value: "active"},
			},
		}
		context := map[string]interface{}{
			"count":  10,
			"status": "active",
		}
		assert.True(t, EvaluateConditionGroup(group, context))
	})

	t.Run("AND logic - one condition false", func(t *testing.T) {
		group := ConditionGroup{
			Logic: LogicAND,
			Conditions: []RuleCondition{
				{Field: "count", Operator: "gt", Value: 5},
				{Field: "status", Operator: "eq", Value: "inactive"},
			},
		}
		context := map[string]interface{}{
			"count":  10,
			"status": "active",
		}
		assert.False(t, EvaluateConditionGroup(group, context))
	})

	t.Run("OR logic - at least one true", func(t *testing.T) {
		group := ConditionGroup{
			Logic: LogicOR,
			Conditions: []RuleCondition{
				{Field: "count", Operator: "gt", Value: 100},
				{Field: "status", Operator: "eq", Value: "active"},
			},
		}
		context := map[string]interface{}{
			"count":  10,
			"status": "active",
		}
		assert.True(t, EvaluateConditionGroup(group, context))
	})

	t.Run("OR logic - all false", func(t *testing.T) {
		group := ConditionGroup{
			Logic: LogicOR,
			Conditions: []RuleCondition{
				{Field: "count", Operator: "gt", Value: 100},
				{Field: "status", Operator: "eq", Value: "inactive"},
			},
		}
		context := map[string]interface{}{
			"count":  10,
			"status": "active",
		}
		assert.False(t, EvaluateConditionGroup(group, context))
	})

	t.Run("nested groups with AND", func(t *testing.T) {
		group := ConditionGroup{
			Logic: LogicAND,
			Conditions: []RuleCondition{
				{Field: "count", Operator: "gt", Value: 5},
			},
			Groups: []ConditionGroup{
				{
					Logic: LogicOR,
					Conditions: []RuleCondition{
						{Field: "status", Operator: "eq", Value: "active"},
						{Field: "status", Operator: "eq", Value: "pending"},
					},
				},
			},
		}
		context := map[string]interface{}{
			"count":  10,
			"status": "active",
		}
		assert.True(t, EvaluateConditionGroup(group, context))
	})
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		op       string
		expected bool
	}{
		{"int > int true", 10, 5, ">", true},
		{"int > int false", 5, 10, ">", false},
		{"float64 >= float64 true", 10.5, 10.5, ">=", true},
		{"int < int true", 5, 10, "<", true},
		{"int32 <= int64 true", int32(5), int64(10), "<=", true},
		{"float32 > int true", float32(10.5), 5, ">", true},
		{"invalid type", "string", 5, ">", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareNumeric(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"float64", float64(10.5), 10.5, true},
		{"float32", float32(10.5), 10.5, true},
		{"int", 10, 10.0, true},
		{"int64", int64(10), 10.0, true},
		{"int32", int32(10), 10.0, true},
		{"string", "invalid", 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toFloat64(tt.input)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		haystack interface{}
		needle   interface{}
		expected bool
	}{
		{"contains - true", "hello world", "hello", true},
		{"contains - false", "hello world", "goodbye", false},
		{"exact match", "hello", "hello", true},
		{"invalid haystack type", 123, "test", false},
		{"invalid needle type", "test", 123, false},
		{"empty strings", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.haystack, tt.needle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInSlice(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		slice    interface{}
		expected bool
	}{
		{"value in slice", "active", []interface{}{"active", "pending"}, true},
		{"value not in slice", "inactive", []interface{}{"active", "pending"}, false},
		{"number in slice", 5, []interface{}{1, 5, 10}, true},
		{"invalid slice type", "test", "not a slice", false},
		{"empty slice", "test", []interface{}{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inSlice(tt.value, tt.slice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReconstructAutomation(t *testing.T) {
	id := uuid.New()
	pipelineID := uuid.New()
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	conditions := []RuleCondition{
		{Field: "status", Operator: "eq", Value: "active"},
	}
	actions := []RuleAction{
		{Type: ActionSendMessage, Params: map[string]interface{}{"content": "Hi"}},
	}

	rule := ReconstructAutomation(
		id,
		AutomationTypePipelineBased,
		&pipelineID,
		"tenant-123",
		"Reconstructed Rule",
		"Test description",
		TriggerStatusChanged,
		conditions,
		actions,
		5,
		true,
		createdAt,
		updatedAt,
	)

	assert.Equal(t, id, rule.ID())
	assert.Equal(t, AutomationTypePipelineBased, rule.Type())
	assert.Equal(t, &pipelineID, rule.PipelineID())
	assert.Equal(t, "tenant-123", rule.TenantID())
	assert.Equal(t, "Reconstructed Rule", rule.Name())
	assert.Equal(t, "Test description", rule.Description())
	assert.Equal(t, TriggerStatusChanged, rule.Trigger())
	assert.Len(t, rule.Conditions(), 1)
	assert.Len(t, rule.Actions(), 1)
	assert.Equal(t, 5, rule.Priority())
	assert.True(t, rule.IsEnabled())
	assert.Equal(t, createdAt, rule.CreatedAt())
	assert.Equal(t, updatedAt, rule.UpdatedAt())
	assert.Len(t, rule.DomainEvents(), 0) // No events on reconstruction
}

func TestGetAvailableOperators(t *testing.T) {
	operators := GetAvailableOperators()

	assert.NotEmpty(t, operators)
	assert.GreaterOrEqual(t, len(operators), 8)

	// Check that all expected operators are present
	codes := make(map[string]bool)
	for _, op := range operators {
		codes[op.Code] = true
		assert.NotEmpty(t, op.Name)
		assert.NotEmpty(t, op.Description)
	}

	assert.True(t, codes["eq"])
	assert.True(t, codes["ne"])
	assert.True(t, codes["gt"])
	assert.True(t, codes["gte"])
	assert.True(t, codes["lt"])
	assert.True(t, codes["lte"])
	assert.True(t, codes["contains"])
	assert.True(t, codes["in"])
}

func TestGetAvailableActions(t *testing.T) {
	actions := GetAvailableActions()

	assert.NotEmpty(t, actions)
	assert.GreaterOrEqual(t, len(actions), 10)

	// Check that all actions have required metadata
	codes := make(map[string]bool)
	for _, action := range actions {
		codes[action.Code] = true
		assert.NotEmpty(t, action.Name)
		assert.NotEmpty(t, action.Description)
		assert.NotEmpty(t, action.Category)
		assert.NotNil(t, action.Parameters)
		assert.NotNil(t, action.Example)
	}

	// Verify some key actions exist
	assert.True(t, codes[string(ActionSendMessage)])
	assert.True(t, codes[string(ActionSendTemplate)])
	assert.True(t, codes[string(ActionChangeStatus)])
	assert.True(t, codes[string(ActionCreateNote)])
}

func TestAutomation_EventManagement(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	t.Run("clear events", func(t *testing.T) {
		assert.Len(t, rule.DomainEvents(), 1) // Creation event

		rule.ClearEvents()
		assert.Len(t, rule.DomainEvents(), 0)
	})

	t.Run("events are immutable copies", func(t *testing.T) {
		rule.Disable()

		events1 := rule.DomainEvents()
		events2 := rule.DomainEvents()

		// Should be different slices (copies)
		assert.NotSame(t, &events1, &events2)
		assert.Equal(t, len(events1), len(events2))
	})
}

func TestAutomation_AllGetters(t *testing.T) {
	pipelineID := uuid.New()
	rule, err := NewAutomation(AutomationTypePipelineBased, "tenant-123", "Test Rule", TriggerStatusChanged, &pipelineID)
	require.NoError(t, err)

	rule.UpdateDescription("Test Description")
	rule.SetPriority(7)
	rule.AddCondition("status", "eq", "active")
	rule.AddAction(ActionSendMessage, map[string]interface{}{"content": "Hi"}, 0)

	t.Run("verify all getters", func(t *testing.T) {
		assert.NotEqual(t, uuid.Nil, rule.ID())
		assert.Equal(t, AutomationTypePipelineBased, rule.Type())
		assert.Equal(t, &pipelineID, rule.PipelineID())
		assert.Equal(t, "tenant-123", rule.TenantID())
		assert.Equal(t, "Test Rule", rule.Name())
		assert.Equal(t, "Test Description", rule.Description())
		assert.Equal(t, TriggerStatusChanged, rule.Trigger())
		assert.Len(t, rule.Conditions(), 1)
		assert.Len(t, rule.Actions(), 1)
		assert.Equal(t, 7, rule.Priority())
		assert.True(t, rule.IsEnabled())
		assert.NotZero(t, rule.CreatedAt())
		assert.NotZero(t, rule.UpdatedAt())
	})
}
