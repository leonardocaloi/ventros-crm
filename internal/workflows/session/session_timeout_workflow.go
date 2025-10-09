package session

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// SessionTimeoutWorkflowInput represents the input for session timeout workflow
type SessionTimeoutWorkflowInput struct {
	SessionID       string        `json:"session_id"`
	ContactID       string        `json:"contact_id"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
	WarningDuration time.Duration `json:"warning_duration"`
}

// SessionTimeoutWorkflowResult represents the result of session timeout
type SessionTimeoutWorkflowResult struct {
	SessionID     string     `json:"session_id"`
	TimedOut      bool       `json:"timed_out"`
	WarningSent   bool       `json:"warning_sent"`
	EndedAt       *time.Time `json:"ended_at,omitempty"`
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`
	ReactivatedAt *time.Time `json:"reactivated_at,omitempty"`
}

// SessionTimeoutSignal represents signals that can reset the timeout
type SessionTimeoutSignal struct {
	SessionID   string    `json:"session_id"`
	ActivityAt  time.Time `json:"activity_at"`
	MessageType string    `json:"message_type"` // "contact_message", "agent_message", "status_change"
}

// SessionTimeoutWorkflow manages session timeout with activity-based reset
func SessionTimeoutWorkflow(ctx workflow.Context, input SessionTimeoutWorkflowInput) (*SessionTimeoutWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting session timeout workflow", "session_id", input.SessionID, "timeout", input.TimeoutDuration)

	result := &SessionTimeoutWorkflowResult{
		SessionID: input.SessionID,
	}

	// Setup signal channel for session activity
	activityChannel := workflow.GetSignalChannel(ctx, "session-activity")

	// Setup query handler for checking timeout status
	err := workflow.SetQueryHandler(ctx, "timeout-status", func() (*SessionTimeoutWorkflowResult, error) {
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	// Main timeout loop with activity reset capability
	for {
		// Create selector for handling multiple events
		selector := workflow.NewSelector(ctx)

		// Timer for warning (if configured)
		var warningTimer workflow.Future
		if input.WarningDuration > 0 && input.WarningDuration < input.TimeoutDuration {
			warningTimer = workflow.NewTimer(ctx, input.WarningDuration)
			selector.AddFuture(warningTimer, func(f workflow.Future) {
				logger.Info("Session timeout warning triggered", "session_id", input.SessionID)

				// Send warning notification
				activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
					StartToCloseTimeout: time.Second * 30,
				})

				warningInput := SessionTimeoutWarningActivity{
					SessionID:     input.SessionID,
					ContactID:     input.ContactID,
					TimeRemaining: input.TimeoutDuration - input.WarningDuration,
				}

				var warningResult SessionTimeoutWarningActivityResult
				err := workflow.ExecuteActivity(activityCtx, SendSessionTimeoutWarningActivity, warningInput).Get(ctx, &warningResult)
				if err != nil {
					logger.Error("Failed to send timeout warning", "error", err.Error())
				} else {
					result.WarningSent = true
					logger.Info("Timeout warning sent successfully", "session_id", input.SessionID)
				}
			})
		}

		// Timer for actual timeout
		timeoutTimer := workflow.NewTimer(ctx, input.TimeoutDuration)
		selector.AddFuture(timeoutTimer, func(f workflow.Future) {
			logger.Info("Session timeout triggered", "session_id", input.SessionID)

			// End session due to timeout
			activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: time.Second * 30,
			})

			timeoutInput := SessionTimeoutActivity{
				SessionID: input.SessionID,
				ContactID: input.ContactID,
				Reason:    "timeout",
			}

			var timeoutResult SessionTimeoutActivityResult
			err := workflow.ExecuteActivity(activityCtx, EndSessionDueToTimeoutActivity, timeoutInput).Get(ctx, &timeoutResult)
			if err != nil {
				logger.Error("Failed to end session due to timeout", "error", err.Error())
			} else {
				result.TimedOut = true
				now := workflow.Now(ctx)
				result.EndedAt = &now
				logger.Info("Session ended due to timeout", "session_id", input.SessionID)
			}
		})

		// Signal for session activity (resets timeout)
		selector.AddReceive(activityChannel, func(c workflow.ReceiveChannel, more bool) {
			var signal SessionTimeoutSignal
			c.Receive(ctx, &signal)

			logger.Info("Session activity detected, resetting timeout",
				"session_id", signal.SessionID,
				"message_type", signal.MessageType,
				"activity_at", signal.ActivityAt)

			// Reset timeout by restarting the loop
			// This effectively cancels current timers and starts fresh
			now := workflow.Now(ctx)
			result.ReactivatedAt = &now
		})

		// Wait for one of the events
		selector.Select(ctx)

		// If session timed out, exit the workflow
		if result.TimedOut {
			break
		}

		// If we received activity signal, continue the loop (restart timers)
		if result.ReactivatedAt != nil {
			logger.Info("Timeout reset due to activity", "session_id", input.SessionID)
			// Reset reactivation timestamp for next iteration
			result.ReactivatedAt = nil
			continue
		}
	}

	return result, nil
}

// CancelSessionTimeoutWorkflow cancels an active session timeout
func CancelSessionTimeoutWorkflow(ctx workflow.Context, sessionID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Cancelling session timeout workflow", "session_id", sessionID)

	// This workflow would be called to explicitly cancel timeout
	// when session is manually ended by agent
	return nil
}

// Session timeout activity types
type SessionTimeoutWarningActivity struct {
	SessionID     string        `json:"session_id"`
	ContactID     string        `json:"contact_id"`
	TimeRemaining time.Duration `json:"time_remaining"`
}

type SessionTimeoutWarningActivityResult struct {
	WarningSentAt time.Time `json:"warning_sent_at"`
	Method        string    `json:"method"` // "whatsapp", "email", "sms"
}

type SessionTimeoutActivity struct {
	SessionID string `json:"session_id"`
	ContactID string `json:"contact_id"`
	Reason    string `json:"reason"` // "timeout", "manual", "transfer"
}

type SessionTimeoutActivityResult struct {
	EndedAt time.Time `json:"ended_at"`
	Summary string    `json:"summary"`
}
