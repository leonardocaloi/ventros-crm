package webhook

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// WebhookDeliveryWorkflowInput represents the input for webhook delivery workflow
type WebhookDeliveryWorkflowInput struct {
	WebhookID   string                 `json:"webhook_id"`
	URL         string                 `json:"url"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers"`
	Payload     map[string]interface{} `json:"payload"`
	MaxRetries  int                    `json:"max_retries"`
	TimeoutSecs int                    `json:"timeout_secs"`
}

// WebhookDeliveryWorkflowResult represents the result of webhook delivery
type WebhookDeliveryWorkflowResult struct {
	Success       bool       `json:"success"`
	StatusCode    int        `json:"status_code"`
	ResponseBody  string     `json:"response_body"`
	AttemptCount  int        `json:"attempt_count"`
	LastAttemptAt time.Time  `json:"last_attempt_at"`
	ErrorMessage  string     `json:"error_message,omitempty"`
	CompensatedAt *time.Time `json:"compensated_at,omitempty"`
}

// WebhookDeliveryActivity represents the activity input/output
type WebhookDeliveryActivity struct {
	URL          string                 `json:"url"`
	Method       string                 `json:"method"`
	Headers      map[string]string      `json:"headers"`
	Payload      map[string]interface{} `json:"payload"`
	TimeoutSecs  int                    `json:"timeout_secs"`
	AttemptCount int                    `json:"attempt_count"`
}

type WebhookDeliveryActivityResult struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code"`
	ResponseBody string `json:"response_body"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// WebhookDeliveryWorkflow implements webhook delivery with exponential backoff retry
func WebhookDeliveryWorkflow(ctx workflow.Context, input WebhookDeliveryWorkflowInput) (*WebhookDeliveryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting webhook delivery workflow", "webhook_id", input.WebhookID, "url", input.URL)

	result := &WebhookDeliveryWorkflowResult{
		Success:      false,
		AttemptCount: 0,
	}

	// Configure retry policy with exponential backoff
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:        time.Second * 1,
		BackoffCoefficient:     2.0,
		MaximumInterval:        time.Minute * 5,
		MaximumAttempts:        int32(input.MaxRetries),
		NonRetryableErrorTypes: []string{"PermanentWebhookError"},
	}

	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Duration(input.TimeoutSecs) * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Attempt webhook delivery
	for attempt := 1; attempt <= input.MaxRetries; attempt++ {
		result.AttemptCount = attempt
		result.LastAttemptAt = workflow.Now(ctx)

		activityInput := WebhookDeliveryActivity{
			URL:          input.URL,
			Method:       input.Method,
			Headers:      input.Headers,
			Payload:      input.Payload,
			TimeoutSecs:  input.TimeoutSecs,
			AttemptCount: attempt,
		}

		var activityResult WebhookDeliveryActivityResult
		err := workflow.ExecuteActivity(ctx, DeliverWebhookActivity, activityInput).Get(ctx, &activityResult)

		if err != nil {
			logger.Warn("Webhook delivery attempt failed", "attempt", attempt, "error", err.Error())
			result.ErrorMessage = err.Error()

			// Check if it's a permanent error (4xx status codes)
			if temporal.IsApplicationError(err) {
				appErr := err.(*temporal.ApplicationError)
				if appErr.Type() == "PermanentWebhookError" {
					logger.Error("Permanent webhook error, stopping retries", "error", err.Error())
					break
				}
			}

			// If this is the last attempt, break
			if attempt == input.MaxRetries {
				break
			}

			// Wait before next attempt (exponential backoff handled by Temporal)
			continue
		}

		// Success!
		result.Success = true
		result.StatusCode = activityResult.StatusCode
		result.ResponseBody = activityResult.ResponseBody
		logger.Info("Webhook delivered successfully", "attempt", attempt, "status_code", activityResult.StatusCode)
		break
	}

	// If all attempts failed, trigger compensation
	if !result.Success {
		logger.Error("All webhook delivery attempts failed, triggering compensation", "webhook_id", input.WebhookID)

		compensationInput := WebhookCompensationActivity{
			WebhookID:    input.WebhookID,
			URL:          input.URL,
			AttemptCount: result.AttemptCount,
			ErrorMessage: result.ErrorMessage,
		}

		var compensationResult WebhookCompensationActivityResult
		compensationCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 30,
		})

		err := workflow.ExecuteActivity(compensationCtx, CompensateWebhookActivity, compensationInput).Get(ctx, &compensationResult)
		if err != nil {
			logger.Error("Webhook compensation failed", "error", err.Error())
		} else {
			now := workflow.Now(ctx)
			result.CompensatedAt = &now
			logger.Info("Webhook compensation completed", "webhook_id", input.WebhookID)
		}
	}

	return result, nil
}

// WebhookCompensationActivity represents compensation activity input
type WebhookCompensationActivity struct {
	WebhookID    string `json:"webhook_id"`
	URL          string `json:"url"`
	AttemptCount int    `json:"attempt_count"`
	ErrorMessage string `json:"error_message"`
}

type WebhookCompensationActivityResult struct {
	CompensatedAt time.Time `json:"compensated_at"`
	Action        string    `json:"action"`
}
