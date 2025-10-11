package saga

import (
	"context"
	"testing"
)

func TestSagaContext(t *testing.T) {
	t.Run("WithSaga creates correlation_id", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithSaga(ctx, string(ProcessInboundMessageSaga))

		correlationID, ok := GetCorrelationID(ctx)
		if !ok {
			t.Error("Expected correlation_id to be set")
		}
		if correlationID == "" {
			t.Error("Expected correlation_id to be non-empty")
		}
	})

	t.Run("GetMetadata returns full metadata", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithSaga(ctx, string(ProcessInboundMessageSaga))
		ctx = WithTenantID(ctx, "tenant-123")

		metadata := GetMetadata(ctx)
		if metadata == nil {
			t.Fatal("Expected metadata to be non-nil")
		}

		if metadata.CorrelationID == "" {
			t.Error("Expected correlation_id to be set")
		}
		if metadata.SagaType != string(ProcessInboundMessageSaga) {
			t.Errorf("Expected saga_type=%s, got=%s", ProcessInboundMessageSaga, metadata.SagaType)
		}
		if metadata.TenantID != "tenant-123" {
			t.Errorf("Expected tenant_id=tenant-123, got=%s", metadata.TenantID)
		}
	})

	t.Run("NextStep increments step_number", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithSaga(ctx, string(ProcessInboundMessageSaga))

		// Initial step number should be 1
		stepNum, ok := GetStepNumber(ctx)
		if !ok || stepNum != 1 {
			t.Errorf("Expected initial step_number=1, got=%d", stepNum)
		}

		// NextStep should increment
		ctx = NextStep(ctx, StepContactCreated)
		stepNum, ok = GetStepNumber(ctx)
		if !ok || stepNum != 2 {
			t.Errorf("Expected step_number=2 after NextStep, got=%d", stepNum)
		}

		// Verify step was set
		step, ok := GetSagaStep(ctx)
		if !ok || step != string(StepContactCreated) {
			t.Errorf("Expected saga_step=%s, got=%s", StepContactCreated, step)
		}
	})

	t.Run("GetMetadata returns nil when no saga active", func(t *testing.T) {
		ctx := context.Background()

		metadata := GetMetadata(ctx)
		if metadata != nil {
			t.Error("Expected metadata to be nil when no saga is active")
		}
	})

	t.Run("SagaType IsFastPath", func(t *testing.T) {
		fastPaths := []SagaType{
			ProcessInboundMessageSaga,
			ChangePipelineStatusSaga,
			CreateContactWithSessionSaga,
		}

		for _, sagaType := range fastPaths {
			if !sagaType.IsFastPath() {
				t.Errorf("Expected %s to be fast path", sagaType)
			}
			if sagaType.IsSlowPath() {
				t.Errorf("Expected %s not to be slow path", sagaType)
			}
		}
	})

	t.Run("SagaType IsSlowPath", func(t *testing.T) {
		slowPaths := []SagaType{
			OnboardCustomerSaga,
			RenewSubscriptionSaga,
		}

		for _, sagaType := range slowPaths {
			if !sagaType.IsSlowPath() {
				t.Errorf("Expected %s to be slow path", sagaType)
			}
			if sagaType.IsFastPath() {
				t.Errorf("Expected %s not to be fast path", sagaType)
			}
		}
	})
}

func TestSagaSteps(t *testing.T) {
	t.Run("GetExpectedSteps returns correct steps", func(t *testing.T) {
		steps := GetExpectedSteps(ProcessInboundMessageSaga)
		if len(steps) == 0 {
			t.Error("Expected ProcessInboundMessageSaga to have steps")
		}

		// Verify expected steps are present
		expectedSteps := []SagaStep{
			StepWAHAReceived,
			StepContactCreated,
			StepSessionStarted,
			StepMessageSaved,
		}

		for _, expected := range expectedSteps {
			found := false
			for _, step := range steps {
				if step == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected step %s not found in GetExpectedSteps", expected)
			}
		}
	})

	t.Run("GetExpectedSteps returns empty for unknown saga", func(t *testing.T) {
		steps := GetExpectedSteps(SagaType("unknown_saga"))
		if len(steps) != 0 {
			t.Error("Expected empty steps for unknown saga type")
		}
	})
}

func TestCompensationEvents(t *testing.T) {
	t.Run("GetCompensationEvent returns correct compensation", func(t *testing.T) {
		testCases := []struct {
			domainEvent     string
			expectedCompensation CompensationEventType
		}{
			{"contact.created", CompensateContactCreated},
			{"session.started", CompensateSessionStarted},
			{"message.created", CompensateMessageCreated},
		}

		for _, tc := range testCases {
			compensation := GetCompensationEvent(tc.domainEvent)
			if compensation != tc.expectedCompensation {
				t.Errorf("Expected compensation=%s for event=%s, got=%s",
					tc.expectedCompensation, tc.domainEvent, compensation)
			}
		}
	})

	t.Run("NeedsCompensation returns true for mapped events", func(t *testing.T) {
		if !NeedsCompensation("contact.created") {
			t.Error("Expected contact.created to need compensation")
		}
		if !NeedsCompensation("session.started") {
			t.Error("Expected session.started to need compensation")
		}
	})

	t.Run("NeedsCompensation returns false for unmapped events", func(t *testing.T) {
		if NeedsCompensation("unknown.event") {
			t.Error("Expected unknown.event not to need compensation")
		}
	})
}
