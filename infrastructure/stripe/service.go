package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v81"
	stripeClient "github.com/stripe/stripe-go/v81/client"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentmethod"
	"github.com/stripe/stripe-go/v81/setupintent"
	"go.uber.org/zap"
)

// Service encapsula operações do Stripe
type Service struct {
	client *stripeClient.API
	logger *zap.Logger
	apiKey string
}

// NewService cria novo service do Stripe
func NewService(apiKey string, logger *zap.Logger) (*Service, error) {
	if apiKey == "" {
		return nil, ErrAPIKeyInvalid
	}

	// Inicializa cliente Stripe
	sc := &stripeClient.API{}
	sc.Init(apiKey, nil)

	// Define API key globalmente (necessário para algumas operações)
	stripe.Key = apiKey

	return &Service{
		client: sc,
		logger: logger,
		apiKey: apiKey,
	}, nil
}

// ==============================
// CUSTOMER OPERATIONS
// ==============================

// CreateCustomerParams parâmetros para criar customer
type CreateCustomerParams struct {
	Email       string
	Name        string
	Phone       string
	Description string
	Metadata    map[string]string
}

// CreateCustomer cria um novo customer no Stripe
func (s *Service) CreateCustomer(ctx context.Context, params CreateCustomerParams) (*stripe.Customer, error) {
	s.logger.Info("creating stripe customer",
		zap.String("email", params.Email),
		zap.String("name", params.Name),
	)

	customerParams := &stripe.CustomerParams{
		Email:       stripe.String(params.Email),
		Name:        stripe.String(params.Name),
		Description: stripe.String(params.Description),
	}

	if params.Phone != "" {
		customerParams.Phone = stripe.String(params.Phone)
	}

	if params.Metadata != nil {
		customerParams.Metadata = params.Metadata
	}

	// Cria customer
	cust, err := customer.New(customerParams)
	if err != nil {
		s.logger.Error("failed to create stripe customer",
			zap.Error(err),
			zap.String("email", params.Email),
		)
		return nil, WrapStripeError(err)
	}

	s.logger.Info("stripe customer created successfully",
		zap.String("customer_id", cust.ID),
		zap.String("email", params.Email),
	)

	return cust, nil
}

// GetCustomer busca customer por ID
func (s *Service) GetCustomer(ctx context.Context, customerID string) (*stripe.Customer, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, WrapStripeError(err)
	}
	return cust, nil
}

// UpdateCustomer atualiza customer
func (s *Service) UpdateCustomer(ctx context.Context, customerID string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	cust, err := customer.Update(customerID, params)
	if err != nil {
		return nil, WrapStripeError(err)
	}
	return cust, nil
}

// DeleteCustomer deleta customer
func (s *Service) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := customer.Del(customerID, nil)
	if err != nil {
		return WrapStripeError(err)
	}
	return nil
}

// ==============================
// PAYMENT METHOD OPERATIONS
// ==============================

// CreateSetupIntent cria SetupIntent para coletar payment method (com 3D Secure)
func (s *Service) CreateSetupIntent(ctx context.Context, customerID string) (*stripe.SetupIntent, error) {
	s.logger.Info("creating setup intent",
		zap.String("customer_id", customerID),
	)

	params := &stripe.SetupIntentParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
		Usage: stripe.String("off_session"), // Para cobranças futuras
	}

	si, err := setupintent.New(params)
	if err != nil {
		s.logger.Error("failed to create setup intent",
			zap.Error(err),
			zap.String("customer_id", customerID),
		)
		return nil, WrapStripeError(err)
	}

	s.logger.Info("setup intent created successfully",
		zap.String("setup_intent_id", si.ID),
		zap.String("client_secret", si.ClientSecret),
	)

	return si, nil
}

// GetSetupIntent busca SetupIntent por ID
func (s *Service) GetSetupIntent(ctx context.Context, setupIntentID string) (*stripe.SetupIntent, error) {
	si, err := setupintent.Get(setupIntentID, nil)
	if err != nil {
		return nil, WrapStripeError(err)
	}
	return si, nil
}

// AttachPaymentMethod anexa payment method ao customer
func (s *Service) AttachPaymentMethod(ctx context.Context, paymentMethodID, customerID string) (*stripe.PaymentMethod, error) {
	s.logger.Info("attaching payment method to customer",
		zap.String("payment_method_id", paymentMethodID),
		zap.String("customer_id", customerID),
	)

	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}

	pm, err := paymentmethod.Attach(paymentMethodID, params)
	if err != nil {
		s.logger.Error("failed to attach payment method",
			zap.Error(err),
			zap.String("payment_method_id", paymentMethodID),
		)
		return nil, WrapStripeError(err)
	}

	return pm, nil
}

// SetDefaultPaymentMethod define payment method como padrão
func (s *Service) SetDefaultPaymentMethod(ctx context.Context, customerID, paymentMethodID string) error {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}

	_, err := customer.Update(customerID, params)
	if err != nil {
		return WrapStripeError(err)
	}

	return nil
}

// ==============================
// BILLING METER OPERATIONS (V2 API)
// ==============================

// MeterEventParams parâmetros para reportar uso
type MeterEventParams struct {
	EventName        string                 // Nome do evento (ex: "message.sent")
	StripeCustomerID string                 // Customer ID
	Value            int64                  // Quantidade de uso
	Timestamp        *time.Time             // Timestamp (opcional, default: now)
	Payload          map[string]interface{} // Payload adicional
}

// ReportUsage reporta uso via Billing Meters V2 API
func (s *Service) ReportUsage(ctx context.Context, params MeterEventParams) error {
	s.logger.Info("reporting usage to stripe billing meter",
		zap.String("event_name", params.EventName),
		zap.String("customer_id", params.StripeCustomerID),
		zap.Int64("value", params.Value),
	)

	// Valida parâmetros
	if params.Value <= 0 {
		return ErrInvalidMeterEventQuantity
	}

	// Valida timestamp (deve estar dentro de 35 dias)
	if params.Timestamp != nil {
		age := time.Since(*params.Timestamp)
		if age > 35*24*time.Hour {
			return ErrMeterEventTooOld
		}
		// Não pode estar mais de 5 minutos no futuro (clock drift)
		if age < -5*time.Minute {
			return ErrMeterEventTooOld
		}
	}

	// Constrói payload
	payload := map[string]interface{}{
		"event_name": params.EventName,
		"payload": map[string]interface{}{
			"value":              fmt.Sprintf("%d", params.Value),
			"stripe_customer_id": params.StripeCustomerID,
		},
	}

	// Adiciona timestamp se fornecido
	if params.Timestamp != nil {
		payload["payload"].(map[string]interface{})["timestamp"] = params.Timestamp.Unix()
	}

	// Adiciona payload customizado
	if params.Payload != nil {
		for k, v := range params.Payload {
			payload["payload"].(map[string]interface{})[k] = v
		}
	}

	// Serializa payload
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal meter event payload: %w", err)
	}

	// Faz request para V2 API
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.stripe.com/v2/billing/meter_events", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Version", "2025-06-30.basil") // V2 API version

	// Executa request
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		s.logger.Error("failed to report usage to stripe",
			zap.Error(err),
			zap.String("event_name", params.EventName),
		)
		return fmt.Errorf("failed to send meter event: %w", err)
	}
	defer resp.Body.Close()

	// Verifica resposta
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		s.logger.Error("stripe meter event failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("event_name", params.EventName),
		)
		return ErrMeterEventCreationFailed
	}

	s.logger.Info("usage reported successfully to stripe",
		zap.String("event_name", params.EventName),
		zap.String("customer_id", params.StripeCustomerID),
		zap.Int64("value", params.Value),
	)

	return nil
}

// ==============================
// SUBSCRIPTION OPERATIONS
// ==============================

// CreateSubscriptionParams parâmetros para criar subscription
type CreateSubscriptionParams struct {
	CustomerID string
	PriceID    string
	TrialDays  int64
	Metadata   map[string]string
}

// CreateSubscription cria subscription (usando SDK padrão V1)
func (s *Service) CreateSubscription(ctx context.Context, params CreateSubscriptionParams) (*stripe.Subscription, error) {
	s.logger.Info("creating stripe subscription",
		zap.String("customer_id", params.CustomerID),
		zap.String("price_id", params.PriceID),
	)

	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(params.CustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(params.PriceID),
			},
		},
	}

	if params.TrialDays > 0 {
		subParams.TrialPeriodDays = stripe.Int64(params.TrialDays)
	}

	if params.Metadata != nil {
		subParams.Metadata = params.Metadata
	}

	sub, err := s.client.Subscriptions.New(subParams)
	if err != nil {
		s.logger.Error("failed to create subscription",
			zap.Error(err),
			zap.String("customer_id", params.CustomerID),
		)
		return nil, WrapStripeError(err)
	}

	s.logger.Info("subscription created successfully",
		zap.String("subscription_id", sub.ID),
		zap.String("status", string(sub.Status)),
	)

	return sub, nil
}
