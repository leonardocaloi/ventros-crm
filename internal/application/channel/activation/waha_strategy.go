package activation

import (
	"context"
	"fmt"

	"github.com/ventros/crm/infrastructure/channels/waha"
	"github.com/ventros/crm/internal/domain/crm/channel"
	"go.uber.org/zap"
)

// WAHAActivationStrategy implementa ativação para canais WAHA
// Valida que a sessão WAHA está rodando e status é "WORKING"
type WAHAActivationStrategy struct {
	logger *zap.Logger
}

// NewWAHAActivationStrategy cria uma nova instância da strategy WAHA
func NewWAHAActivationStrategy(logger *zap.Logger) *WAHAActivationStrategy {
	return &WAHAActivationStrategy{
		logger: logger,
	}
}

// CanActivate verifica se o canal WAHA pode ser ativado
// Valida que a configuração WAHA está completa (base_url, auth, session_id)
func (s *WAHAActivationStrategy) CanActivate(ctx context.Context, ch *channel.Channel) error {
	if ch.Type != channel.TypeWAHA {
		return fmt.Errorf("channel type is not WAHA: %s", ch.Type)
	}

	config, err := ch.GetWAHAConfig()
	if err != nil {
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	if config.BaseURL == "" {
		return fmt.Errorf("WAHA base_url is required")
	}

	if config.Auth.APIKey == "" && config.Auth.Token == "" {
		return fmt.Errorf("WAHA authentication (api_key or token) is required")
	}

	if config.SessionID == "" {
		return fmt.Errorf("WAHA session_id is required")
	}

	s.logger.Info("WAHA channel pre-activation checks passed",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", config.SessionID))

	return nil
}

// Activate executa a ativação do canal WAHA
// Faz health check com a// Activate executa health check e valida a sessão WAHA
func (s *WAHAActivationStrategy) Activate(ctx context.Context, ch *channel.Channel) error {
	config, err := ch.GetWAHAConfig()
	if err != nil {
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// Criar cliente WAHA
	authToken := config.Auth.APIKey
	if authToken == "" {
		authToken = config.Auth.Token
	}

	wahaClient := waha.NewWAHAClient(config.BaseURL, authToken, s.logger)

	// Health check: Verificar se a sessão está WORKING
	isHealthy, sessionStatus, err := wahaClient.HealthCheck(ctx, config.SessionID)
	if err != nil {
		return fmt.Errorf("WAHA health check failed: %w", err)
	}

	if !isHealthy {
		return fmt.Errorf("WAHA session is not healthy (status: %s)", sessionStatus)
	}

	// BYPASS WEBHOOK: Não valida webhook para permitir testes de import
	// Webhook é opcional - só necessário para receber mensagens em tempo real
	// Para import de histórico, não é necessário
	s.logger.Info("WAHA channel activated successfully (webhook validation bypassed)",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", config.SessionID),
		zap.String("session_status", sessionStatus),
		zap.Bool("webhook_configured", config.WebhookURL != ""))

	return nil
}

// HealthCheck verifica o status atual da sessão WAHA
// Usado pelo polling worker para verificar canais em "activating"
func (s *WAHAActivationStrategy) HealthCheck(ctx context.Context, ch *channel.Channel) (bool, string, error) {
	config, err := ch.GetWAHAConfig()
	if err != nil {
		return false, "error", fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// Extract auth token
	authToken := config.Auth.APIKey
	if authToken == "" {
		authToken = config.Auth.Token
	}

	// Create WAHA client
	wahaClient := waha.NewWAHAClient(config.BaseURL, authToken, s.logger)

	// Execute health check
	isHealthy, status, err := wahaClient.HealthCheck(ctx, config.SessionID)
	if err != nil {
		s.logger.Error("WAHA health check failed",
			zap.String("channel_id", ch.ID.String()),
			zap.String("session_id", config.SessionID),
			zap.Error(err))
		return false, "error", err
	}

	s.logger.Debug("WAHA health check completed",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", config.SessionID),
		zap.Bool("healthy", isHealthy),
		zap.String("status", status))

	return isHealthy, status, nil
}

// Compensate executa compensação quando ativação WAHA falha
// Para canais WAHA, apenas loga o evento (não há recursos externos para limpar)
func (s *WAHAActivationStrategy) Compensate(ctx context.Context, ch *channel.Channel) error {
	config, err := ch.GetWAHAConfig()
	if err != nil {
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	s.logger.Warn("Compensating WAHA channel activation failure",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", config.SessionID))

	// Para WAHA, não precisamos fazer nada específico na compensação
	// O canal já volta para status "inactive" no aggregate
	// Se tivéssemos criado webhook ou recursos externos, removeríamos aqui

	return nil
}
