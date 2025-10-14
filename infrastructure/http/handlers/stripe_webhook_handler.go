package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/ventros/crm/internal/domain/core/billing"
	"go.uber.org/zap"
)

// StripeWebhookHandler handles Stripe webhook events
type StripeWebhookHandler struct {
	logger           *zap.Logger
	webhookSecret    string
	billingRepo      billing.Repository
	subscriptionRepo billing.SubscriptionRepository
	invoiceRepo      billing.InvoiceRepository
	usageMeterRepo   billing.UsageMeterRepository
}

// NewStripeWebhookHandler creates a new Stripe webhook handler
func NewStripeWebhookHandler(
	logger *zap.Logger,
	webhookSecret string,
	billingRepo billing.Repository,
	subscriptionRepo billing.SubscriptionRepository,
	invoiceRepo billing.InvoiceRepository,
	usageMeterRepo billing.UsageMeterRepository,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		logger:           logger,
		webhookSecret:    webhookSecret,
		billingRepo:      billingRepo,
		subscriptionRepo: subscriptionRepo,
		invoiceRepo:      invoiceRepo,
		usageMeterRepo:   usageMeterRepo,
	}
}

// HandleWebhook receives and processes Stripe webhook events
//
//	@Summary		Receive Stripe webhook
//	@Description	Recebe e processa eventos de webhook do Stripe
//	@Tags			webhooks
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Event processed"
//	@Failure		400	{object}	map[string]interface{}	"Invalid request"
//	@Failure		500	{object}	map[string]interface{}	"Internal server error"
//	@Router			/api/v1/webhooks/stripe [post]
func (h *StripeWebhookHandler) HandleWebhook(c *gin.Context) {
	// 1. Ler o corpo da requisição
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 2. Validar assinatura do webhook
	signatureHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, signatureHeader, h.webhookSecret)
	if err != nil {
		h.logger.Warn("Invalid webhook signature",
			zap.Error(err),
			zap.String("signature", signatureHeader))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	// 3. Log do evento recebido
	h.logger.Info("Stripe webhook received",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)))

	// 4. Processar evento baseado no tipo
	ctx := c.Request.Context()
	switch event.Type {
	case "invoice.paid":
		err = h.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		err = h.handleInvoicePaymentFailed(ctx, event)
	case "invoice.payment_action_required":
		err = h.handleInvoicePaymentActionRequired(ctx, event)
	case "customer.subscription.created":
		err = h.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		err = h.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		err = h.handleSubscriptionDeleted(ctx, event)
	case "setup_intent.succeeded":
		err = h.handleSetupIntentSucceeded(ctx, event)
	case "customer.created":
		err = h.handleCustomerCreated(ctx, event)
	case "customer.updated":
		err = h.handleCustomerUpdated(ctx, event)
	default:
		h.logger.Info("Unhandled webhook event type",
			zap.String("event_type", string(event.Type)))
	}

	if err != nil {
		h.logger.Error("Failed to process webhook event",
			zap.Error(err),
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process event"})
		return
	}

	// 5. Resposta de sucesso
	c.JSON(http.StatusOK, gin.H{
		"status":     "processed",
		"event_id":   event.ID,
		"event_type": string(event.Type),
	})
}

// handleInvoicePaid processa evento de invoice paga
func (h *StripeWebhookHandler) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return err
	}

	h.logger.Info("Processing invoice.paid",
		zap.String("invoice_id", stripeInvoice.ID),
		zap.Int64("amount_paid", stripeInvoice.AmountPaid))

	// Buscar invoice no banco
	inv, err := h.invoiceRepo.FindByStripeInvoiceID(ctx, stripeInvoice.ID)
	if err != nil {
		// Se não existe, criar nova invoice
		h.logger.Warn("Invoice not found, creating new one",
			zap.String("stripe_invoice_id", stripeInvoice.ID))

		// Buscar billing account pelo customer ID
		// TODO: Implementar busca por stripe_customer_id
		return nil
	}

	// Atualizar status da invoice
	if err := inv.MarkAsPaid(stripeInvoice.AmountPaid); err != nil {
		return err
	}

	return h.invoiceRepo.Update(ctx, inv)
}

// handleInvoicePaymentFailed processa falha no pagamento
func (h *StripeWebhookHandler) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return err
	}

	h.logger.Warn("Processing invoice.payment_failed",
		zap.String("invoice_id", stripeInvoice.ID),
		zap.Int64("amount_due", stripeInvoice.AmountDue))

	inv, err := h.invoiceRepo.FindByStripeInvoiceID(ctx, stripeInvoice.ID)
	if err != nil {
		return err
	}

	inv.MarkAsFailedPayment()
	return h.invoiceRepo.Update(ctx, inv)
}

// handleInvoicePaymentActionRequired processa ação necessária (3D Secure)
func (h *StripeWebhookHandler) handleInvoicePaymentActionRequired(ctx context.Context, event stripe.Event) error {
	var stripeInvoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &stripeInvoice); err != nil {
		return err
	}

	h.logger.Info("Processing invoice.payment_action_required (3D Secure)",
		zap.String("invoice_id", stripeInvoice.ID),
		zap.String("hosted_invoice_url", stripeInvoice.HostedInvoiceURL))

	// TODO: Enviar notificação ao usuário com URL para completar autenticação
	return nil
}

// handleSubscriptionCreated processa criação de subscription
func (h *StripeWebhookHandler) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return err
	}

	h.logger.Info("Processing customer.subscription.created",
		zap.String("subscription_id", stripeSub.ID),
		zap.String("customer_id", stripeSub.Customer.ID))

	// TODO: Buscar billing account pelo customer ID e criar subscription
	return nil
}

// handleSubscriptionUpdated processa atualização de subscription
func (h *StripeWebhookHandler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return err
	}

	h.logger.Info("Processing customer.subscription.updated",
		zap.String("subscription_id", stripeSub.ID),
		zap.String("status", string(stripeSub.Status)))

	sub, err := h.subscriptionRepo.FindByStripeSubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		h.logger.Warn("Subscription not found",
			zap.String("stripe_subscription_id", stripeSub.ID))
		return nil
	}

	// Atualizar status
	sub.UpdateStatus(billing.SubscriptionStatus(stripeSub.Status))

	// Atualizar período (converter Unix timestamp para time.Time)
	sub.UpdatePeriod(
		time.Unix(stripeSub.CurrentPeriodStart, 0),
		time.Unix(stripeSub.CurrentPeriodEnd, 0),
	)

	return h.subscriptionRepo.Update(ctx, sub)
}

// handleSubscriptionDeleted processa cancelamento de subscription
func (h *StripeWebhookHandler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		return err
	}

	h.logger.Info("Processing customer.subscription.deleted",
		zap.String("subscription_id", stripeSub.ID))

	sub, err := h.subscriptionRepo.FindByStripeSubscriptionID(ctx, stripeSub.ID)
	if err != nil {
		return err
	}

	sub.CancelImmediately()
	return h.subscriptionRepo.Update(ctx, sub)
}

// handleSetupIntentSucceeded processa sucesso na coleta de payment method
func (h *StripeWebhookHandler) handleSetupIntentSucceeded(ctx context.Context, event stripe.Event) error {
	var setupIntent stripe.SetupIntent
	if err := json.Unmarshal(event.Data.Raw, &setupIntent); err != nil {
		return err
	}

	h.logger.Info("Processing setup_intent.succeeded",
		zap.String("setup_intent_id", setupIntent.ID),
		zap.String("customer_id", setupIntent.Customer.ID),
		zap.String("payment_method", setupIntent.PaymentMethod.ID))

	// TODO: Atualizar billing account com novo payment method
	return nil
}

// handleCustomerCreated processa criação de customer
func (h *StripeWebhookHandler) handleCustomerCreated(ctx context.Context, event stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return err
	}

	h.logger.Info("Processing customer.created",
		zap.String("customer_id", customer.ID),
		zap.String("email", customer.Email))

	return nil
}

// handleCustomerUpdated processa atualização de customer
func (h *StripeWebhookHandler) handleCustomerUpdated(ctx context.Context, event stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return err
	}

	h.logger.Info("Processing customer.updated",
		zap.String("customer_id", customer.ID))

	return nil
}

// GetWebhookInfo provides information about the Stripe webhook endpoint
//
//	@Summary		Get Stripe webhook info
//	@Description	Retorna informações sobre o endpoint de webhook do Stripe
//	@Tags			webhooks
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}	"Webhook info"
//	@Router			/api/v1/webhooks/stripe/info [get]
func (h *StripeWebhookHandler) GetWebhookInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"endpoint":     "/api/v1/webhooks/stripe",
		"method":       "POST",
		"content_type": "application/json",
		"supported_events": []string{
			"invoice.paid",
			"invoice.payment_failed",
			"invoice.payment_action_required",
			"customer.subscription.created",
			"customer.subscription.updated",
			"customer.subscription.deleted",
			"setup_intent.succeeded",
			"customer.created",
			"customer.updated",
		},
		"description": "Endpoint para receber eventos de webhook do Stripe",
		"configuration": map[string]interface{}{
			"signature_verification": "Required (Stripe-Signature header)",
			"webhook_secret":         "Configure STRIPE_WEBHOOK_SECRET in environment",
			"stripe_dashboard":       "Configure this URL in Stripe Dashboard > Webhooks",
		},
		"example_events": map[string]interface{}{
			"invoice_paid": map[string]interface{}{
				"type":        "invoice.paid",
				"description": "Triggered when an invoice is successfully paid",
				"action":      "Updates invoice status to 'paid' in database",
			},
			"subscription_updated": map[string]interface{}{
				"type":        "customer.subscription.updated",
				"description": "Triggered when subscription status or billing period changes",
				"action":      "Updates subscription in database",
			},
			"setup_intent_succeeded": map[string]interface{}{
				"type":        "setup_intent.succeeded",
				"description": "Triggered when payment method is successfully collected (3D Secure complete)",
				"action":      "Payment method is now available for future charges",
			},
		},
	})
}
