package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

// DeliverWebhookActivity performs the actual HTTP request to deliver webhook
func DeliverWebhookActivity(ctx context.Context, input WebhookDeliveryActivity) (*WebhookDeliveryActivityResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing webhook delivery activity", "url", input.URL, "attempt", input.AttemptCount)

	// Prepare request body
	var requestBody []byte
	var err error
	if input.Payload != nil {
		requestBody, err = json.Marshal(input.Payload)
		if err != nil {
			return nil, temporal.NewApplicationError("failed to marshal payload", "PayloadMarshalError", err)
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, input.Method, input.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, temporal.NewApplicationError("failed to create request", "RequestCreationError", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Ventros-CRM-Webhook/1.0")
	req.Header.Set("X-Webhook-Attempt", fmt.Sprintf("%d", input.AttemptCount))
	req.Header.Set("X-Webhook-Timestamp", time.Now().UTC().Format(time.RFC3339))

	for key, value := range input.Headers {
		req.Header.Set(key, value)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(input.TimeoutSecs) * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("HTTP request failed", "error", err.Error())
		return nil, temporal.NewApplicationError("HTTP request failed", "HTTPRequestError", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("Failed to read response body", "error", err.Error())
		responseBody = []byte("failed to read response")
	}

	result := &WebhookDeliveryActivityResult{
		StatusCode:   resp.StatusCode,
		ResponseBody: string(responseBody),
	}

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		result.Success = true
		logger.Info("Webhook delivered successfully", "status_code", resp.StatusCode)
		return result, nil
	}

	// Handle different error types
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// Client error - permanent error, don't retry
		errorMsg := fmt.Sprintf("permanent webhook error: status %d, body: %s", resp.StatusCode, string(responseBody))
		result.ErrorMessage = errorMsg
		logger.Error("Permanent webhook error", "status_code", resp.StatusCode, "response", string(responseBody))
		return result, temporal.NewApplicationError(errorMsg, "PermanentWebhookError")
	}

	// Server error (5xx) or other - temporary error, retry
	errorMsg := fmt.Sprintf("temporary webhook error: status %d, body: %s", resp.StatusCode, string(responseBody))
	result.ErrorMessage = errorMsg
	logger.Warn("Temporary webhook error", "status_code", resp.StatusCode, "response", string(responseBody))
	return result, temporal.NewApplicationError(errorMsg, "TemporaryWebhookError")
}

// CompensateWebhookActivity handles webhook delivery failure compensation
func CompensateWebhookActivity(ctx context.Context, input WebhookCompensationActivity) (*WebhookCompensationActivityResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Executing webhook compensation activity", "webhook_id", input.WebhookID)

	// Here you would typically:
	// 1. Mark webhook as failed in database
	// 2. Send notification to admin
	// 3. Create alert/incident
	// 4. Update metrics
	// 5. Trigger alternative notification method

	result := &WebhookCompensationActivityResult{
		CompensatedAt: time.Now().UTC(),
		Action:        "webhook_marked_as_failed",
	}

	logger.Info("Webhook compensation completed", "webhook_id", input.WebhookID, "action", result.Action)
	return result, nil
}

// WebhookStatusUpdateActivity updates webhook status in database
func WebhookStatusUpdateActivity(ctx context.Context, input WebhookStatusUpdateInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Updating webhook status", "webhook_id", input.WebhookID, "status", input.Status)

	// TODO: Implement database update
	// This would update the webhook record with:
	// - Status (pending, delivered, failed)
	// - Attempt count
	// - Last attempt timestamp
	// - Response details
	// - Error message if failed

	return nil
}

type WebhookStatusUpdateInput struct {
	WebhookID     string    `json:"webhook_id"`
	Status        string    `json:"status"` // pending, delivered, failed
	AttemptCount  int       `json:"attempt_count"`
	LastAttemptAt time.Time `json:"last_attempt_at"`
	StatusCode    *int      `json:"status_code,omitempty"`
	ResponseBody  *string   `json:"response_body,omitempty"`
	ErrorMessage  *string   `json:"error_message,omitempty"`
}
