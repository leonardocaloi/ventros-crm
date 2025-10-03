package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/caloi/ventros-crm/internal/domain/webhook"
	"go.uber.org/zap"
)

type WebhookNotifier struct {
	logger     *zap.Logger
	repo       webhook.Repository
	httpClient *http.Client
}

func NewWebhookNotifier(logger *zap.Logger, repo webhook.Repository) *WebhookNotifier {
	return &WebhookNotifier{
		logger: logger,
		repo:   repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type WebhookPayload struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NotifyWebhooks sends event to all subscribed webhooks
func (n *WebhookNotifier) NotifyWebhooks(ctx context.Context, eventType string, eventData interface{}) {
	// Find all active webhooks subscribed to this event
	webhooks, err := n.repo.FindActiveByEvent(ctx, eventType)
	if err != nil {
		n.logger.Error("Failed to query webhooks", zap.Error(err))
		return
	}

	if len(webhooks) == 0 {
		n.logger.Debug("No webhooks subscribed to event", zap.String("event", eventType))
		return
	}

	n.logger.Info("Notifying webhooks",
		zap.String("event", eventType),
		zap.Int("webhook_count", len(webhooks)),
	)

	// Prepare payload
	payload := WebhookPayload{
		Event:     eventType,
		Timestamp: time.Now().UTC(),
		Data:      eventData,
	}

	// Notify each webhook in parallel
	for _, webhook := range webhooks {
		go n.notifyWebhook(webhook, payload)
	}
}

func (n *WebhookNotifier) notifyWebhook(sub *webhook.WebhookSubscription, payload WebhookPayload) {
	ctx := context.Background()
	start := time.Now()

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		n.logger.Error("Failed to marshal webhook payload",
			zap.Error(err),
			zap.String("webhook_id", sub.ID.String()),
		)
		return
	}

	// Try up to retry_count times
	var lastErr error
	for attempt := 0; attempt < sub.RetryCount; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
			n.logger.Info("Retrying webhook",
				zap.String("webhook_id", sub.ID.String()),
				zap.Int("attempt", attempt+1),
			)
		}

		err := n.sendWebhook(sub, payloadBytes)
		if err == nil {
			// Success
			duration := time.Since(start)
			if err := n.repo.RecordTrigger(ctx, sub.ID, true); err != nil {
				n.logger.Error("Failed to record success", zap.Error(err))
			}
			n.logger.Info("Webhook sent successfully",
				zap.String("webhook_id", sub.ID.String()),
				zap.String("webhook_name", sub.Name),
				zap.String("event", payload.Event),
				zap.Duration("duration", duration),
				zap.Int("attempts", attempt+1),
			)
			return
		}

		lastErr = err
	}

	// All retries failed
	duration := time.Since(start)
	if err := n.repo.RecordTrigger(ctx, sub.ID, false); err != nil {
		n.logger.Error("Failed to record failure", zap.Error(err))
	}
	n.logger.Error("Webhook failed after retries",
		zap.String("webhook_id", sub.ID.String()),
		zap.String("webhook_name", sub.Name),
		zap.String("event", payload.Event),
		zap.Error(lastErr),
		zap.Duration("duration", duration),
		zap.Int("attempts", sub.RetryCount),
	)
}

func (n *WebhookNotifier) sendWebhook(sub *webhook.WebhookSubscription, payloadBytes []byte) error {
	// Create request
	req, err := http.NewRequest("POST", sub.URL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Ventros-CRM-Webhook/1.0")

	// Add custom headers
	for key, value := range sub.Headers {
		req.Header.Set(key, value)
	}

	// Add HMAC signature if secret is provided
	if sub.Secret != "" {
		signature := n.generateHMAC(payloadBytes, sub.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(sub.TimeoutSeconds)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Send request
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for logging
	bodyBytes, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (n *WebhookNotifier) generateHMAC(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
