package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

func TestSessionLifecycleWorkflow_Timeout(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register mock activity
	env.RegisterActivity(func(input EndSessionActivityInput) (EndSessionActivityResult, error) {
		return EndSessionActivityResult{Success: true, EventsPublished: 2}, nil
	})

	sessionID := uuid.New()
	channelTypeID := 1

	input := SessionLifecycleWorkflowInput{
		SessionID:       sessionID,
		ContactID:       uuid.New(),
		TenantID:        "tenant-123",
		ChannelTypeID:   &channelTypeID,
		TimeoutDuration: 100 * time.Millisecond, // Short for testing
	}

	env.ExecuteWorkflow(SessionLifecycleWorkflow, input)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestSessionCleanupWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register mock activity
	env.RegisterActivity(func(input CleanupSessionsActivityInput) (CleanupSessionsActivityResult, error) {
		return CleanupSessionsActivityResult{SessionsCleaned: 5, EventsPublished: 10}, nil
	})

	env.ExecuteWorkflow(SessionCleanupWorkflow)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestSessionCleanupWorkflow_Error(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register failing mock activity
	env.RegisterActivity(func(input CleanupSessionsActivityInput) (CleanupSessionsActivityResult, error) {
		return CleanupSessionsActivityResult{}, assert.AnError
	})

	env.ExecuteWorkflow(SessionCleanupWorkflow)

	assert.True(t, env.IsWorkflowCompleted())
	assert.Error(t, env.GetWorkflowError())
}

// Benchmark para avaliar performance
func BenchmarkSessionLifecycleWorkflow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestWorkflowEnvironment()

		env.RegisterActivity(func(input EndSessionActivityInput) (EndSessionActivityResult, error) {
			return EndSessionActivityResult{Success: true, EventsPublished: 2}, nil
		})

		sessionID := uuid.New()
		channelTypeID := 1

		input := SessionLifecycleWorkflowInput{
			SessionID:       sessionID,
			ContactID:       uuid.New(),
			TenantID:        "tenant-123",
			ChannelTypeID:   &channelTypeID,
			TimeoutDuration: 1 * time.Millisecond,
		}

		env.ExecuteWorkflow(SessionLifecycleWorkflow, input)
	}
}
