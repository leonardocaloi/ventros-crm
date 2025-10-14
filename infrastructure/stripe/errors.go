package stripe

import (
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/v81"
)

// Stripe-specific domain errors
var (
	// Customer errors
	ErrCustomerNotFound       = errors.New("stripe customer not found")
	ErrCustomerAlreadyExists  = errors.New("stripe customer already exists")
	ErrCustomerCreationFailed = errors.New("failed to create stripe customer")
	ErrCustomerUpdateFailed   = errors.New("failed to update stripe customer")
	ErrCustomerDeletionFailed = errors.New("failed to delete stripe customer")

	// Payment Method errors
	ErrPaymentMethodNotFound       = errors.New("payment method not found")
	ErrPaymentMethodAttachFailed   = errors.New("failed to attach payment method")
	ErrPaymentMethodDetachFailed   = errors.New("failed to detach payment method")
	ErrInvalidPaymentMethod        = errors.New("invalid payment method")
	ErrPaymentMethodRequiresAction = errors.New("payment method requires additional action (3D Secure)")

	// SetupIntent errors
	ErrSetupIntentCreationFailed = errors.New("failed to create setup intent")
	ErrSetupIntentNotFound       = errors.New("setup intent not found")
	ErrSetupIntentFailed         = errors.New("setup intent failed")
	ErrSetupIntentCanceled       = errors.New("setup intent was canceled")

	// Subscription errors
	ErrSubscriptionNotFound       = errors.New("subscription not found")
	ErrSubscriptionCreationFailed = errors.New("failed to create subscription")
	ErrSubscriptionUpdateFailed   = errors.New("failed to update subscription")
	ErrSubscriptionCancelFailed   = errors.New("failed to cancel subscription")
	ErrInvalidSubscriptionStatus  = errors.New("invalid subscription status")

	// Invoice errors
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrInvoiceCreationFailed = errors.New("failed to create invoice")
	ErrInvoiceUpdateFailed   = errors.New("failed to update invoice")
	ErrInvoicePaymentFailed  = errors.New("invoice payment failed")
	ErrInvoiceVoidFailed     = errors.New("failed to void invoice")

	// Usage/Meter errors
	ErrMeterNotFound               = errors.New("billing meter not found")
	ErrMeterCreationFailed         = errors.New("failed to create billing meter")
	ErrMeterEventCreationFailed    = errors.New("failed to create meter event")
	ErrInvalidMeterEventQuantity   = errors.New("meter event quantity must be positive")
	ErrMeterEventTooOld            = errors.New("meter event timestamp too old (must be within 35 days)")
	ErrMeterEventRateLimitExceeded = errors.New("meter event rate limit exceeded (1000/s)")

	// Webhook errors
	ErrWebhookVerificationFailed = errors.New("webhook signature verification failed")
	ErrWebhookInvalidPayload     = errors.New("invalid webhook payload")
	ErrWebhookEventNotFound      = errors.New("webhook event not found")

	// API errors
	ErrAPIKeyInvalid        = errors.New("stripe API key is invalid")
	ErrAPIRequestFailed     = errors.New("stripe API request failed")
	ErrRateLimitExceeded    = errors.New("stripe API rate limit exceeded")
	ErrNetworkError         = errors.New("network error communicating with Stripe")
	ErrInvalidRequest       = errors.New("invalid request to Stripe API")
	ErrAuthenticationFailed = errors.New("stripe authentication failed")
)

// StripeError encapsula erros da Stripe API com contexto adicional
type StripeError struct {
	Type    string // Stripe error type (card_error, invalid_request_error, etc.)
	Code    string // Stripe error code (card_declined, invalid_number, etc.)
	Message string // Human-readable error message
	Param   string // Parameter that caused the error
	Cause   error  // Original error
}

func (e *StripeError) Error() string {
	if e.Param != "" {
		return fmt.Sprintf("stripe error [%s/%s] on param '%s': %s", e.Type, e.Code, e.Param, e.Message)
	}
	return fmt.Sprintf("stripe error [%s/%s]: %s", e.Type, e.Code, e.Message)
}

func (e *StripeError) Unwrap() error {
	return e.Cause
}

// WrapStripeError converte erro da Stripe SDK em StripeError
func WrapStripeError(err error) error {
	if err == nil {
		return nil
	}

	// Se já é StripeError, retorna
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return err
	}

	// Converte erro do Stripe SDK
	var sdkErr *stripe.Error
	if errors.As(err, &sdkErr) {
		return &StripeError{
			Type:    string(sdkErr.Type),
			Code:    string(sdkErr.Code),
			Message: sdkErr.Msg,
			Param:   sdkErr.Param,
			Cause:   err,
		}
	}

	// Erro desconhecido
	return fmt.Errorf("unexpected stripe error: %w", err)
}

// IsCardDeclined verifica se o erro é de cartão recusado
func IsCardDeclined(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Type == string(stripe.ErrorTypeCard) &&
			stripeErr.Code == string(stripe.ErrorCodeCardDeclined)
	}
	return false
}

// IsInsufficientFunds verifica se o erro é de saldo insuficiente
func IsInsufficientFunds(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Code == string(stripe.ErrorCodeInsufficientFunds)
	}
	return false
}

// IsCardExpired verifica se o erro é de cartão expirado
func IsCardExpired(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Code == string(stripe.ErrorCodeExpiredCard)
	}
	return false
}

// RequiresAuthentication verifica se o erro requer autenticação (3D Secure)
func RequiresAuthentication(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Code == string(stripe.ErrorCodeAuthenticationRequired)
	}
	return false
}

// IsRateLimitError verifica se é erro de rate limit
// TODO: Update when stripe-go v81 exposes ErrorTypeRateLimit
func IsRateLimitError(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		// Check for rate_limit error type string directly
		return stripeErr.Type == "rate_limit"
	}
	return false
}

// IsAPIConnectionError verifica se é erro de conexão com API
// TODO: Update when stripe-go v81 exposes ErrorTypeAPIConnection
func IsAPIConnectionError(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		// Check for api_connection_error type string directly
		return stripeErr.Type == "api_connection_error"
	}
	return false
}

// IsInvalidRequestError verifica se é erro de request inválido
func IsInvalidRequestError(err error) bool {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Type == string(stripe.ErrorTypeInvalidRequest)
	}
	return false
}

// GetErrorCode retorna o código de erro do Stripe
func GetErrorCode(err error) string {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Code
	}
	return ""
}

// GetErrorType retorna o tipo de erro do Stripe
func GetErrorType(err error) string {
	var stripeErr *StripeError
	if errors.As(err, &stripeErr) {
		return stripeErr.Type
	}
	return ""
}
