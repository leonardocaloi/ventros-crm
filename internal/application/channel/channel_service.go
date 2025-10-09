package channel

import (
	"context"
	"fmt"
	"time"

	"github.com/caloi/ventros-crm/infrastructure/channels/waha"
	"github.com/caloi/ventros-crm/internal/domain/channel"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ChannelService gerencia canais de comunicação
type ChannelService struct {
	repo            channel.Repository
	logger          *zap.Logger
	wahaClient      *waha.WAHAClient
	historyImporter *WAHAHistoryImporter
}

// NewChannelService cria um novo serviço de canais
func NewChannelService(
	repo channel.Repository,
	logger *zap.Logger,
	wahaClient *waha.WAHAClient,
	historyImporter *WAHAHistoryImporter,
) *ChannelService {
	return &ChannelService{
		repo:            repo,
		logger:          logger,
		wahaClient:      wahaClient,
		historyImporter: historyImporter,
	}
}

// CreateChannelRequest representa os dados para criar um canal
type CreateChannelRequest struct {
	UserID                uuid.UUID          `json:"user_id"`
	ProjectID             uuid.UUID          `json:"project_id"`
	TenantID              string             `json:"tenant_id"`
	Name                  string             `json:"name" binding:"required"`
	Type                  string             `json:"type" binding:"required"`
	SessionTimeoutMinutes *int               `json:"session_timeout_minutes,omitempty"` // Timeout das sessões em minutos (sobrescreve pipeline se ambos existirem)
	WAHAConfig            *WAHAConfigRequest `json:"waha_config,omitempty"`
	AIEnabled             bool               `json:"ai_enabled"`        // Canal Inteligente
	AIAgentsEnabled       bool               `json:"ai_agents_enabled"` // Agentes IA
}

// WAHAConfigRequest represents a WAHA configuration
type WAHAConfigRequest struct {
	BaseURL        string `json:"base_url" binding:"required"`
	APIKey         string `json:"api_key"`    // Chave da API para autenticação
	Token          string `json:"token"`      // Token de acesso (alternativo à API key)
	SessionID      string `json:"session_id"` // ID da sessão WAHA (equivale ao ExternalID)
	WebhookURL     string `json:"webhook_url"`
	ImportStrategy string `json:"import_strategy"` // Estratégia de importação: none, new_only, all
}

// ChannelResponse representa a resposta de um canal
type ChannelResponse struct {
	ID         uuid.UUID              `json:"id"`
	UserID     uuid.UUID              `json:"user_id"`
	ProjectID  uuid.UUID              `json:"project_id"`
	TenantID   string                 `json:"tenant_id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Status     string                 `json:"status"`
	ExternalID string                 `json:"external_id,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`

	// Webhook info
	WebhookURL          string  `json:"webhook_url,omitempty"`
	WebhookConfiguredAt *string `json:"webhook_configured_at,omitempty"`
	WebhookActive       bool    `json:"webhook_active"`

	// AI Features
	AIEnabled       bool `json:"ai_enabled,omitempty"`        // Canal Inteligente
	AIAgentsEnabled bool `json:"ai_agents_enabled,omitempty"` // Agentes IA

	// Statistics
	MessagesReceived int     `json:"messages_received"`
	MessagesSent     int     `json:"messages_sent"`
	LastMessageAt    *string `json:"last_message_at,omitempty"`
	LastErrorAt      *string `json:"last_error_at,omitempty"`
	LastError        string  `json:"last_error,omitempty"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateChannel cria um novo canal
func (s *ChannelService) CreateChannel(ctx context.Context, req CreateChannelRequest) (*ChannelResponse, error) {
	channelType := channel.ChannelType(req.Type)

	var ch *channel.Channel
	var err error

	// Criar canal baseado no tipo
	switch channelType {
	case channel.TypeWAHA:
		if req.WAHAConfig == nil {
			return nil, fmt.Errorf("WAHA configuration is required for WAHA channels")
		}

		// Parse import strategy
		importStrategy := channel.WAHAImportStrategy(req.WAHAConfig.ImportStrategy)
		if importStrategy == "" {
			importStrategy = channel.WAHAImportNone
		}

		wahaConfig := channel.WAHAConfig{
			BaseURL: req.WAHAConfig.BaseURL,
			Auth: channel.WAHAAuth{
				APIKey: req.WAHAConfig.APIKey,
				Token:  req.WAHAConfig.Token,
			},
			SessionID:      req.WAHAConfig.SessionID,
			WebhookURL:     req.WAHAConfig.WebhookURL,
			ImportStrategy: importStrategy,
		}

		ch, err = channel.NewWAHAChannel(req.UserID, req.ProjectID, req.TenantID, req.Name, wahaConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create WAHA channel: %w", err)
		}

	default:
		ch, err = channel.NewChannel(req.UserID, req.ProjectID, req.TenantID, req.Name, channelType)
		if err != nil {
			return nil, fmt.Errorf("failed to create channel: %w", err)
		}
	}

	// Configurar AI Features
	ch.AIEnabled = req.AIEnabled
	ch.AIAgentsEnabled = req.AIAgentsEnabled

	// Validar: Agentes IA requer Canal Inteligente
	if ch.AIAgentsEnabled && !ch.AIEnabled {
		return nil, fmt.Errorf("AI agents require AI-enabled channel")
	}

	// Configurar timeout da sessão (se fornecido)
	if req.SessionTimeoutMinutes != nil {
		if err := ch.SetDefaultTimeout(*req.SessionTimeoutMinutes); err != nil {
			return nil, fmt.Errorf("failed to set session timeout: %w", err)
		}
	}

	// Persistir no banco
	if err := s.repo.Create(ch); err != nil {
		s.logger.Error("Failed to create channel",
			zap.Error(err),
			zap.String("name", req.Name),
			zap.String("type", req.Type),
			zap.String("user_id", req.UserID.String()))
		return nil, fmt.Errorf("failed to save channel: %w", err)
	}

	// Se for canal WAHA, gerar URL do webhook automaticamente
	if channelType == channel.TypeWAHA {
		baseURL := "http://localhost:8080" // Default, pode vir de config
		if req.WAHAConfig != nil && req.WAHAConfig.WebhookURL != "" {
			baseURL = req.WAHAConfig.WebhookURL
		}

		webhookURL := fmt.Sprintf("%s/api/v1/webhooks/waha/%s", baseURL, ch.ExternalID)

		// Atualizar canal com webhook URL
		now := time.Now()
		ch.WebhookURL = webhookURL
		ch.WebhookConfiguredAt = &now
		ch.WebhookActive = true

		if err := s.repo.Update(ch); err != nil {
			s.logger.Warn("Failed to update channel with webhook URL", zap.Error(err))
		}

		s.logger.Info("WAHA channel created with webhook URL",
			zap.String("id", ch.ID.String()),
			zap.String("webhook_url", webhookURL))
	}

	s.logger.Info("Channel created successfully",
		zap.String("id", ch.ID.String()),
		zap.String("name", ch.Name),
		zap.String("type", string(ch.Type)),
		zap.String("user_id", ch.UserID.String()))

	return s.toResponse(ch), nil
}

// GetChannelsByUser retorna todos os canais de um usuário
func (s *ChannelService) GetChannelsByUser(ctx context.Context, userID uuid.UUID) ([]*ChannelResponse, error) {
	channels, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	responses := make([]*ChannelResponse, len(channels))
	for i, ch := range channels {
		responses[i] = s.toResponse(ch)
	}

	return responses, nil
}

// GetChannelsByProject retorna todos os canais de um projeto
func (s *ChannelService) GetChannelsByProject(ctx context.Context, projectID uuid.UUID) ([]*ChannelResponse, error) {
	channels, err := s.repo.GetByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	responses := make([]*ChannelResponse, len(channels))
	for i, ch := range channels {
		responses[i] = s.toResponse(ch)
	}

	return responses, nil
}

// GetChannel retorna um canal específico
func (s *ChannelService) GetChannel(ctx context.Context, channelID uuid.UUID) (*ChannelResponse, error) {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return s.toResponse(ch), nil
}

// ActivateChannel ativa um canal
func (s *ChannelService) ActivateChannel(ctx context.Context, channelID uuid.UUID) error {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	ch.Activate()

	if err := s.repo.Update(ch); err != nil {
		return fmt.Errorf("failed to activate channel: %w", err)
	}

	s.logger.Info("Channel activated",
		zap.String("id", ch.ID.String()),
		zap.String("name", ch.Name))

	return nil
}

// DeactivateChannel desativa um canal
func (s *ChannelService) DeactivateChannel(ctx context.Context, channelID uuid.UUID) error {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	ch.Deactivate()

	if err := s.repo.Update(ch); err != nil {
		return fmt.Errorf("failed to deactivate channel: %w", err)
	}

	s.logger.Info("Channel deactivated",
		zap.String("id", ch.ID.String()),
		zap.String("name", ch.Name))

	return nil
}

// DeleteChannel deleta um canal
func (s *ChannelService) DeleteChannel(ctx context.Context, channelID uuid.UUID) error {
	if err := s.repo.Delete(channelID); err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	s.logger.Info("Channel deleted", zap.String("id", channelID.String()))
	return nil
}

// toResponse converte domínio para response
func (s *ChannelService) toResponse(ch *channel.Channel) *ChannelResponse {
	response := &ChannelResponse{
		ID:         ch.ID,
		UserID:     ch.UserID,
		ProjectID:  ch.ProjectID,
		TenantID:   ch.TenantID,
		Name:       ch.Name,
		Type:       string(ch.Type),
		Status:     string(ch.Status),
		ExternalID: ch.ExternalID,
		Config:     ch.Config,

		// Webhook
		WebhookURL:    ch.WebhookURL,
		WebhookActive: ch.WebhookActive,

		// AI Features
		AIEnabled:       ch.AIEnabled,
		AIAgentsEnabled: ch.AIAgentsEnabled,

		// Statistics
		MessagesReceived: ch.MessagesReceived,
		MessagesSent:     ch.MessagesSent,
		LastError:        ch.LastError,

		CreatedAt: ch.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ch.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if ch.WebhookConfiguredAt != nil {
		webhookConfigured := ch.WebhookConfiguredAt.Format("2006-01-02T15:04:05Z07:00")
		response.WebhookConfiguredAt = &webhookConfigured
	}

	if ch.LastMessageAt != nil {
		lastMsg := ch.LastMessageAt.Format("2006-01-02T15:04:05Z07:00")
		response.LastMessageAt = &lastMsg
	}

	if ch.LastErrorAt != nil {
		lastErr := ch.LastErrorAt.Format("2006-01-02T15:04:05Z07:00")
		response.LastErrorAt = &lastErr
	}

	return response
}

// GetWebhookURL retorna a URL do webhook para o canal
func (s *ChannelService) GetWebhookURL(ctx context.Context, channelID uuid.UUID, baseURL string) (string, error) {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return "", fmt.Errorf("failed to get channel: %w", err)
	}

	// URL do webhook é baseada no tipo do canal
	switch ch.Type {
	case channel.TypeWAHA:
		// Para WAHA, usa o endpoint genérico
		return fmt.Sprintf("%s/api/v1/webhooks/waha/%s", baseURL, ch.ExternalID), nil
	case channel.TypeWhatsApp:
		return fmt.Sprintf("%s/api/v1/webhooks/whatsapp/%s", baseURL, ch.ID.String()), nil
	case channel.TypeTelegram:
		return fmt.Sprintf("%s/api/v1/webhooks/telegram/%s", baseURL, ch.ID.String()), nil
	default:
		return "", fmt.Errorf("webhook URL not supported for channel type: %s", ch.Type)
	}
}

// ConfigureWebhook configura o webhook no canal externo (ex: WAHA)
func (s *ChannelService) ConfigureWebhook(ctx context.Context, channelID uuid.UUID, webhookURL string) error {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	switch ch.Type {
	case channel.TypeWAHA:
		return s.configureWAHAWebhook(ctx, ch, webhookURL)
	default:
		return fmt.Errorf("webhook configuration not supported for channel type: %s", ch.Type)
	}
}

// configureWAHAWebhook configura webhook na WAHA
func (s *ChannelService) configureWAHAWebhook(ctx context.Context, ch *channel.Channel, webhookURL string) error {
	if s.wahaClient == nil {
		return fmt.Errorf("WAHA client not configured")
	}

	wahaConfig, err := ch.GetWAHAConfig()
	if err != nil {
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// Eventos padrão do WAHA
	events := waha.GetDefaultWebhookEvents()

	// Configurar webhook na WAHA
	if err := s.wahaClient.SetWebhook(ctx, wahaConfig.SessionID, webhookURL, events); err != nil {
		s.logger.Error("Failed to configure WAHA webhook",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("session_id", wahaConfig.SessionID),
			zap.String("webhook_url", webhookURL))
		return fmt.Errorf("failed to configure WAHA webhook: %w", err)
	}

	// Atualizar webhook URL no canal (campos dedicados + config)
	now := time.Now()
	ch.WebhookURL = webhookURL
	ch.WebhookConfiguredAt = &now
	ch.WebhookActive = true
	ch.Config["webhook_url"] = webhookURL
	ch.Config["webhook_configured_at"] = now.Format(time.RFC3339)

	if err := s.repo.Update(ch); err != nil {
		s.logger.Warn("Failed to update channel with webhook URL", zap.Error(err))
	}

	s.logger.Info("WAHA webhook configured successfully",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", wahaConfig.SessionID),
		zap.String("webhook_url", webhookURL),
		zap.Int("events_count", len(events)))

	return nil
}

// GetWebhookInfo retorna informações sobre o webhook do canal
func (s *ChannelService) GetWebhookInfo(ctx context.Context, channelID uuid.UUID) (map[string]interface{}, error) {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	info := map[string]interface{}{
		"channel_id":   ch.ID,
		"channel_name": ch.Name,
		"channel_type": string(ch.Type),
		"external_id":  ch.ExternalID,
	}

	// Adicionar URL configurada se existir
	if webhookURL, ok := ch.Config["webhook_url"].(string); ok {
		info["webhook_url"] = webhookURL
	}

	if configuredAt, ok := ch.Config["webhook_configured_at"]; ok {
		info["webhook_configured_at"] = configuredAt
	}

	// Informações específicas por tipo
	switch ch.Type {
	case channel.TypeWAHA:
		info["supported_events"] = waha.GetDefaultWebhookEvents()
		info["webhook_method"] = "POST"
		info["content_type"] = "application/json"
	}

	return info, nil
}

// ActivateWAHAChannel ativa um canal WAHA verificando saúde da sessão
func (s *ChannelService) ActivateWAHAChannel(ctx context.Context, channelID uuid.UUID) error {
	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	if !ch.IsWAHA() {
		return fmt.Errorf("channel is not WAHA type")
	}

	wahaConfig, err := ch.GetWAHAConfig()
	if err != nil {
		return fmt.Errorf("failed to get WAHA config: %w", err)
	}

	// Criar cliente WAHA específico para este canal usando seu base_url
	authToken := wahaConfig.Auth.APIKey
	if authToken == "" {
		authToken = wahaConfig.Auth.Token
	}
	channelWAHAClient := waha.NewWAHAClient(wahaConfig.BaseURL, authToken, s.logger)

	isHealthy, status, err := channelWAHAClient.HealthCheck(ctx, wahaConfig.SessionID)
	if err != nil {
		s.logger.Error("WAHA health check failed",
			zap.Error(err),
			zap.String("channel_id", ch.ID.String()),
			zap.String("session_id", wahaConfig.SessionID))
		ch.SetError(fmt.Sprintf("Health check failed: %v", err))
		s.repo.Update(ch)
		return fmt.Errorf("WAHA health check failed: %w", err)
	}

	if !isHealthy {
		s.logger.Warn("WAHA session not healthy",
			zap.String("channel_id", ch.ID.String()),
			zap.String("session_id", wahaConfig.SessionID),
			zap.String("status", status))
		ch.SetError(fmt.Sprintf("Session not healthy: %s", status))
		s.repo.Update(ch)
		return fmt.Errorf("WAHA session not healthy: %s", status)
	}

	// Atualizar status da sessão
	ch.SetWAHASessionStatus(channel.WAHASessionStatusWorking)
	ch.Activate()

	if err := s.repo.Update(ch); err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	s.logger.Info("WAHA channel activated",
		zap.String("channel_id", ch.ID.String()),
		zap.String("session_id", wahaConfig.SessionID),
		zap.String("status", status))

	return nil
}

// ImportWAHAHistory importa histórico de mensagens do WAHA
func (s *ChannelService) ImportWAHAHistory(ctx context.Context, channelID uuid.UUID, limit int) (*ImportHistoryResult, error) {
	if s.historyImporter == nil {
		return nil, fmt.Errorf("history importer not configured")
	}

	ch, err := s.repo.GetByID(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	if !ch.IsWAHA() {
		return nil, fmt.Errorf("channel is not WAHA type")
	}

	strategy := ch.GetWAHAImportStrategy()
	if strategy == channel.WAHAImportNone {
		return nil, fmt.Errorf("channel has no import strategy configured")
	}

	req := ImportHistoryRequest{
		ChannelID: channelID,
		Strategy:  strategy,
		Limit:     limit,
	}

	return s.historyImporter.ImportHistory(ctx, req)
}
