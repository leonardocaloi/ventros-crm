package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
	"go.uber.org/zap"
)

// Client wrapper para Stripe API
type Client struct {
	api    *client.API
	logger *zap.Logger
}

// NewClient cria novo cliente Stripe
func NewClient(apiKey string, logger *zap.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("stripe API key is required")
	}

	// Configura Stripe SDK
	stripe.Key = apiKey

	// Cria API client
	stripeClient := &client.API{}
	stripeClient.Init(apiKey, nil)

	return &Client{
		api:    stripeClient,
		logger: logger,
	}, nil
}

// API retorna o cliente nativo do Stripe
func (c *Client) API() *client.API {
	return c.api
}

// SetLogLevel configura nível de log do Stripe SDK
// TODO: Update when stripe-go v81 backend configuration API is available
func (c *Client) SetLogLevel(level int) {
	// stripe.SetBackend configuration has changed in v81
	// Keeping method for API compatibility but not implementing for now
	c.logger.Debug("SetLogLevel called but not implemented for stripe-go v81")
}

// stripeLogger adapta zap.Logger para stripe.LeveledLogger
type stripeLogger struct {
	logger *zap.Logger
}

func (l *stripeLogger) Debugf(format string, v ...interface{}) {
	l.logger.Sugar().Debugf(format, v...)
}

func (l *stripeLogger) Errorf(format string, v ...interface{}) {
	l.logger.Sugar().Errorf(format, v...)
}

func (l *stripeLogger) Infof(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v...)
}

func (l *stripeLogger) Warnf(format string, v ...interface{}) {
	l.logger.Sugar().Warnf(format, v...)
}
